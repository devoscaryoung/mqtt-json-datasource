package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
	"github.com/oliveagle/jsonpath"
)

var (
	_ backend.QueryDataHandler      = (*MqttDatasource)(nil)
	_ backend.CheckHealthHandler    = (*MqttDatasource)(nil)
	_ backend.StreamHandler         = (*MqttDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*MqttDatasource)(nil)
)

func NewMqttDatasource(s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {

	// Reading MQTT connection options from datasource settings.
	mqttOptions := &mqttOptions{}
	err := json.Unmarshal(s.JSONData, &mqttOptions)
	if err != nil {
		return nil, err
	}

	mqttOptions.Password = s.DecryptedSecureJSONData["password"]

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", mqttOptions.Endpoint))
	if mqttOptions.Username != "" {
		opts.SetUsername(mqttOptions.Username)
	}
	if mqttOptions.Password != "" {
		opts.SetPassword(mqttOptions.Password)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}

	// Connect to MQTT broker.
	if err := token.Error(); err != nil {
		log.DefaultLogger.Error("Error connecting to MQTT broker", "error", err)
	}

	return &MqttDatasource{
		mqttOptions:  mqttOptions,
		queryOptions: &queryOptionsModel{},
		client:       client,
		msgChan:      make(chan mqtt.Message),
	}, nil
}

type mqttOptions struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type MqttDatasource struct {
	mqttOptions  *mqttOptions
	queryOptions *queryOptionsModel
	client       mqtt.Client
	msgChan      chan mqtt.Message
}

func (d *MqttDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *MqttDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	data, _ := json.Marshal(req)
	log.DefaultLogger.Info("QueryData called", "request", string(data))
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type jsonpathOptionModel struct {
	Jsonpath string `json:"jsonpath"`
	Alias    string `json:"alias"`
	DataType string `json:"dataType"`
}

type queryOptionsModel struct {
	Topic           string                `json:"topic"`
	JsonpathOptions []jsonpathOptionModel `json:"jsonpathOptions"`
}

func (d *MqttDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// When refresh query is called, we need to subscribe to the new topic.
	d.client.Subscribe(d.queryOptions.Topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		// Send received message to the channel.
		d.msgChan <- msg
	})

	var option *queryOptionsModel

	response.Error = json.Unmarshal(query.JSON, &option)
	if response.Error != nil {
		log.DefaultLogger.Error("Error unmarshalling query", "error", response.Error)
		return response
	}

	d.queryOptions.Topic = option.Topic
	d.queryOptions.JsonpathOptions = option.JsonpathOptions

	// create data frame response.
	frame := data.NewFrame("response")

	channel := live.Channel{
		Scope:     live.ScopeDatasource,
		Namespace: pCtx.DataSourceInstanceSettings.UID,
		Path:      "stream",
	}
	frame.SetMeta(&data.FrameMeta{Channel: channel.String()})

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *MqttDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	status := backend.HealthStatusOk
	message := "Data source is working"

	if !d.client.IsConnected() {
		status = backend.HealthStatusError
		message = "Not connected to MQTT broker"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *MqttDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusPermissionDenied

	if d.client.IsConnected() {
		status = backend.SubscribeStreamStatusOK
	}

	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *MqttDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			return nil
		case msg := <-d.msgChan:

			var json_data interface{}
			json.Unmarshal([]byte(msg.Payload()), &json_data)

			// Create the same data frame as for query data.
			frame := data.NewFrame("response")

			// Parsing data with jsonpath

			haveTimeAlias := false

			for index, jsonField := range d.queryOptions.JsonpathOptions {

				if jsonField.Alias == "time" {
					haveTimeAlias = true
					continue
				}

				if jsonField.DataType == "string" {
					frame.Fields = append(frame.Fields, data.NewField(jsonField.Alias, nil, make([]string, 1)))
				} else if jsonField.DataType == "number" {
					frame.Fields = append(frame.Fields, data.NewField(jsonField.Alias, nil, make([]float64, 1)))
				}

				// Parseing json with jsonpath
				res, err := jsonpath.JsonPathLookup(json_data, jsonField.Jsonpath)
				if err != nil {
					continue
				}

				jsonBytes, _ := json.Marshal(res)

				if jsonField.DataType == "string" {
					frame.Fields[index].Set(0, string(jsonBytes))
				} else if jsonField.DataType == "number" {
					num, _ := strconv.ParseFloat(string(jsonBytes), 64)
					frame.Fields[index].Set(0, num)
				}
			}

			if !haveTimeAlias {
				frame.Fields = append(frame.Fields, data.NewField("time", nil, make([]time.Time, 1)))
				frame.Fields[len(frame.Fields)-1].Set(0, time.Now())
			}

			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
			}
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (d *MqttDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
