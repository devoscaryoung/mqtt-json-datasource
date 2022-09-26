import { defaults } from 'lodash';

import React, { ChangeEvent, PureComponent } from 'react';
import { Button, LegacyForms, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, MqttOptions, MqttQuery } from './types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, MqttQuery, MqttOptions>;

export class QueryEditor extends PureComponent<Props> {
  onTopicChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, topic: event.target.value });
  };

  onJsonPathChange = (index: number) => (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    const { jsonpathOptions: jsonpathOptions } = query;
    jsonpathOptions[index].jsonpath = event.target.value;

    onChange({ ...query, jsonpathOptions: jsonpathOptions });
  };

  onAliasChange = (index: number) => (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    const { jsonpathOptions: jsonpathOptions } = query;
    jsonpathOptions[index].alias = event.target.value;

    onChange({ ...query, jsonpathOptions: jsonpathOptions });
  };

  onAppendField = () => {
    const { onChange, query } = this.props;
    const { jsonpathOptions: jsonpathOptions } = query;
    jsonpathOptions.push({ jsonpath: '$', alias: 'mqtt_message', dataType: 'string' });

    onChange({ ...query, jsonpathOptions: jsonpathOptions });
  };
  onRemoveField = (index: number) => () => {
    const { onChange, query } = this.props;
    const { jsonpathOptions: jsonpathOptions } = query;
    jsonpathOptions.splice(index, 1);

    onChange({ ...query, jsonpathOptions: jsonpathOptions });
  };

  onDataTypeChange = (index: number) => (event: SelectableValue) => {
    const { onChange, query } = this.props;
    const { jsonpathOptions: jsonpathOptions } = query;

    jsonpathOptions[index].dataType = event.value;

    onChange({ ...query, jsonpathOptions: jsonpathOptions });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { topic, jsonpathOptions: jsonpathOptions } = query;

    return (
      <>
        <div className="gf-form">
          <FormField
            labelWidth={16}
            value={topic || 'topic'}
            onChange={this.onTopicChange}
            label="Topic"
            tooltip="The topic to subscribe to"
          />
        </div>

        {jsonpathOptions.map((jsonpathOption, index) => {
          return (
            <div className="gf-form" key={index}>
              <FormField
                labelWidth={8}
                value={jsonpathOption.jsonpath || '$'}
                onChange={this.onJsonPathChange(index)}
                label="Jsonpath"
                tooltip="jsonpath"
              />
              <FormField
                labelWidth={8}
                value={jsonpathOption.alias || 'mqtt_message'}
                onChange={this.onAliasChange(index)}
                label="Alias"
                tooltip="alias"
              />
              <Select
                width={12}
                onChange={this.onDataTypeChange(index)}
                value={jsonpathOption.dataType || "string"}
                options={[
                  { label: 'String', value: 'string' },
                  { label: 'Number', value: 'number' },
                ]}
              />
              <Button variant="secondary" icon="minus" onClick={this.onRemoveField(index)}>
                Remove
              </Button>
            </div>
          );
        })}

        <Button variant="primary" icon="plus" onClick={this.onAppendField}>
          Add
        </Button>
      </>
    );
  }
}
