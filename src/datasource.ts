import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { MqttOptions, MqttQuery } from './types';

export class DataSource extends DataSourceWithBackend<MqttQuery, MqttOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MqttOptions>) {
    super(instanceSettings);
  }
}
