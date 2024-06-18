package service // dnywonnt.me/alerts2incidents/internal/service

import (
	"context"
)

// CollectorType is a custom type defined as a string.
// It's used to represent the type of collector.
type CollectorType string

// Below are constants of type CollectorType, each representing a different type of collector.
const (
	// GrafanaAMCollector represents a Grafana Alert Manager collector.
	GrafanaAMCollector CollectorType = "grafana_alertmanager"
	// GrafanaPrometheusCollector represents a Grafana Prometheus collector.
	GrafanaPrometheusCollector CollectorType = "grafana_prometheus"
	// ZabbixCollector represents a Zabbix collector.
	ZabbixCollector CollectorType = "zabbix"
)

// Collector represents an interface for data collectors.
type Collector interface {
	// CollectData is a method that implementations of Collector should define.
	// It should asynchronously collect data and send it to the provided channel.
	// The method accepts a context for handling cancellations and timeouts,
	// and a channel for sending collected data.
	//
	// ctx: a context.Context used for cancellation signals and deadlines.
	// dataCh: a channel for sending collected data, where each data item is
	// a map from CollectorType to a byte slice ([]byte).
	CollectData(ctx context.Context, dataCh chan<- map[CollectorType][]byte)
}
