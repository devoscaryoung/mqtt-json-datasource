import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MqttOptions, MqttSecureJsonData } from './types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MqttOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onEndpointChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      endpoint: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };
  onUsernameChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      username: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        password: event.target.value,
      },
    });
  };

  onResetPassword = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as MqttSecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="Endpoint"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onEndpointChange}
            value={jsonData.endpoint || ''}
            placeholder="127.0.0.1:1883"
          />
        </div>
        <div className="gf-form">
          <FormField
            label="Username"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onUsernameChange}
            value={jsonData.username || ''}
            placeholder="Username"
          />
        </div>
        <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
              value={secureJsonData.password || ''}
              label="Password"
              placeholder="Password"
              labelWidth={6}
              inputWidth={20}
              onReset={this.onResetPassword}
              onChange={this.onPasswordChange}
            />
          </div>
        </div>
      </div>
    );
  }
}
