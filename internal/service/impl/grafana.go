package impl // dnywonnt.me/alerts2incidents/internal/service/impl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"dnywonnt.me/alerts2incidents/internal/config"
	"dnywonnt.me/alerts2incidents/internal/service"

	log "github.com/sirupsen/logrus"
)

// GrafanaCollector is a struct that defines the configuration and client for a Grafana data collector.
type GrafanaCollector struct {
	cfg    *config.GrafanaCollectorConfig // Configuration for the Grafana collector.
	client http.Client                    // HTTP client to make requests to Grafana.
}

// NewGrafanaCollector creates a new instance of GrafanaCollector with the provided configuration.
// It logs an initializing event and returns the initialized collector.
func NewGrafanaCollector(cfg *config.GrafanaCollectorConfig) *GrafanaCollector {
	log.Debug("Initializing the Grafana collector")
	return &GrafanaCollector{
		cfg:    cfg,
		client: http.Client{},
	}
}

// CollectData implements the Collector interface for GrafanaCollector.
// It periodically fetches data based on the configured timeout and sends it to the provided data channel.
func (gc *GrafanaCollector) CollectData(ctx context.Context, dataCh chan<- map[service.CollectorType][]byte) {
	// Check if the collector is active; if not, log a warning and return.
	if !gc.cfg.IsActive {
		log.WithFields(log.Fields{
			"isActive": gc.cfg.IsActive,
		}).Warn("Grafana collector is inactive; exiting data collection process")
		return
	}

	// Log the start of data collection and verify the collect timeout is valid.
	log.WithFields(log.Fields{
		"collectTimeout": gc.cfg.CollectTimeout.String(),
	}).Debug("Starting the Grafana data collection")

	// Create a ticker that triggers data collection at intervals defined by the collect timeout.
	ticker := time.NewTicker(gc.cfg.CollectTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// If the context is done, stop data collection and log the event.
			log.Debug("Stopping the Grafana data collection")
			return
		case <-ticker.C:
			// On each tick, fetch data from the Grafana Alertmanager and Prometheus endpoints.
			const alertManagerEndpoint = "/api/alertmanager/grafana/api/v2/alerts?active=true&silenced=false&inhibited=true"
			amData, err := gc.fetchData(ctx, alertManagerEndpoint)
			if err != nil {
				// If fetching data fails, log the error and continue to the next iteration.
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("Failed to fetch data from the Grafana Alertmanager")
				continue
			}
			// Send the fetched data to the data channel.
			sendData(dataCh, service.GrafanaAMCollector, amData)

			// If configured, fetch data from the Prometheus as well.
			if gc.cfg.IncludePrometheusAlerts {
				for _, uid := range gc.cfg.PrometheusUIDs {
					prometheusEndpoint := fmt.Sprintf("/api/prometheus/%s/api/v1/alerts", uid)
					promData, err := gc.fetchData(ctx, prometheusEndpoint)
					if err != nil {
						// Log any errors encountered while fetching Prometheus data.
						log.WithFields(log.Fields{
							"error": err.Error(),
						}).Error("Failed to fetch data from the Grafana Prometheus")
						continue
					}
					// Send the Prometheus data to the data channel.
					sendData(dataCh, service.GrafanaPrometheusCollector, promData)
				}
			}
		}
	}
}

// fetchData makes a GET request to the specified Grafana API endpoint and returns the response data.
func (gc *GrafanaCollector) fetchData(ctx context.Context, endpoint string) ([]byte, error) {
	// Log the data fetching event.
	log.WithFields(log.Fields{
		"endpoint": endpoint,
	}).Debug("Fetching data from the Grafana")

	// Create a new HTTP request with the provided context and endpoint.
	req, err := http.NewRequestWithContext(ctx, "GET", gc.cfg.APIUrl+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set the Authorization header with the configured bearer token.
	req.Header.Set("Authorization", "Bearer "+gc.cfg.Token)

	// Execute the HTTP request.
	resp, err := gc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read and return the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Log the successful data fetch.
	log.WithFields(log.Fields{
		"endpoint":   endpoint,
		"dataLength": len(body),
	}).Debug("Data successfully fetched from the Grafana")

	return body, nil
}

// sendData sends the collected data to the specified data channel, logging the event.
func sendData(dataCh chan<- map[service.CollectorType][]byte, collectorType service.CollectorType, data []byte) {
	dataMap := map[service.CollectorType][]byte{collectorType: data}
	select {
	case dataCh <- dataMap:
		// Log successful data transmission.
		log.WithFields(log.Fields{
			"collectorType":   collectorType,
			"dataChannelSize": len(dataCh),
		}).Debug("Data sent to the channel")
	default:
		// Log a warning if the data channel is blocked.
		log.WithFields(log.Fields{
			"collectorType":   collectorType,
			"dataChannelSize": len(dataCh),
		}).Warn("The data channel is blocked; skipping data send")
	}
}
