import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface JsonpathOption {
  alias: string;
  jsonpath: string;
  dataType: string;
}

export interface MqttQuery extends DataQuery {
  topic?: string;
  jsonpathOptions: JsonpathOption[];
  type: string;
}

export const defaultQuery: Partial<MqttQuery> = {
  topic: 'topic',
  jsonpathOptions: [{ jsonpath: '$', alias: 'mqtt_message',  dataType: 'string' }],
};

/**
 * These are options configured for each DataSource instance.
 */
export interface MqttOptions extends DataSourceJsonData {
  endpoint?: string;
  username?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MqttSecureJsonData {
  password?: string;
}
