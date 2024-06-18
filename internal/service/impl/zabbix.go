package impl // dnywonnt.me/alerts2incidents/internal/service/impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dnywonnt.me/alerts2incidents/internal/config"
	"dnywonnt.me/alerts2incidents/internal/service"

	log "github.com/sirupsen/logrus"
)

// ZabbixCollector is a struct that holds the configuration and client for interacting with a Zabbix server.
type ZabbixCollector struct {
	cfg    *config.ZabbixCollectorConfig // Configuration specific to the Zabbix collector.
	client http.Client                   // HTTP client used to make API requests to Zabbix.
}

// NewZabbixCollector initializes a new instance of ZabbixCollector with the given configuration.
// It logs an event indicating that the Zabbix collector is being initialized.
func NewZabbixCollector(cfg *config.ZabbixCollectorConfig) *ZabbixCollector {
	log.Debug("Initializing the Zabbix collector")
	return &ZabbixCollector{
		cfg:    cfg,
		client: http.Client{},
	}
}

// CollectData is the method that ZabbixCollector implements from the Collector interface.
// It periodically fetches data from the Zabbix based on a specified timeout and sends this data to a channel.
func (zc *ZabbixCollector) CollectData(ctx context.Context, dataCh chan<- map[service.CollectorType][]byte) {
	// First, check if the Zabbix collector is active. If not, log a warning and exit the method.
	if !zc.cfg.IsActive {
		log.WithFields(log.Fields{
			"isActive": zc.cfg.IsActive,
		}).Warn("Zabbix collector is inactive; exiting data collection process")
		return
	}

	// Log the beginning of the data collection process and ensure the collect timeout is valid.
	log.WithFields(log.Fields{
		"collectTimeout": zc.cfg.CollectTimeout.String(),
	}).Debug("Starting the Zabbix data collection")

	// Create a ticker that will trigger data collection at intervals specified by the collect timeout.
	ticker := time.NewTicker(zc.cfg.CollectTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// If the context signals done, stop the data collection and log this event.
			log.Debug("Stopping the Zabbix data collection")
			return
		case <-ticker.C:
			// When the ticker ticks, fetch data from the Zabbix and attempt to send it through the data channel.
			data, err := zc.fetchData(ctx)
			if err != nil {
				// If there's an error fetching data, log it and continue to the next iteration.
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("Failed to fetch data from the Zabbix")
				continue
			}

			// If data is successfully fetched, wrap it in a map and send it on the data channel.
			dataMap := map[service.CollectorType][]byte{service.ZabbixCollector: data}
			select {
			case dataCh <- dataMap:
				// Log that the data was successfully sent.
				log.WithFields(log.Fields{
					"dataChannelSize": len(dataCh),
				}).Debug("Data sent to the channel")
			default:
				// If the channel is blocked, log a warning.
				log.WithFields(log.Fields{
					"dataChannelSize": len(dataCh),
				}).Warn("The data channel is blocked; skipping data send")
			}
		}
	}
}

// fetchData is a ZabbixCollector method that makes a POST request to the Zabbix API to retrieve data.
// It constructs a JSON-RPC request based on the collector's configuration and returns the fetched data.
func (zc *ZabbixCollector) fetchData(ctx context.Context) ([]byte, error) {
	// Log the data fetching event.
	log.WithFields(log.Fields{
		"APIUrl":          zc.cfg.APIUrl,
		"TriggerMinLevel": zc.cfg.TriggerMinLevel,
	}).Debug("Fetching data from the Zabbix")

	// Construct the request payload for the Zabbix JSON-RPC API.
	requestPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "trigger.get",
		"params": map[string]interface{}{
			"only_true":                   1,
			"active":                      1,
			"withLastEventUnacknowledged": 1,
			"min_severity":                zc.cfg.TriggerMinLevel,
			"expandDescription":           1,
			"selectHosts":                 []string{"host"},
			"monitored":                   1,
			"filter": map[string]interface{}{
				"value": 1,
			},
		},
		"auth": zc.cfg.Token,
		"id":   1,
	}

	// Serialize the request payload to JSON.
	data, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request payload: %w", err)
	}

	// Create a new HTTP POST request with the serialized data.
	req, err := http.NewRequestWithContext(ctx, "POST", zc.cfg.APIUrl+"/api_jsonrpc.php", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set the necessary headers for the request.
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request.
	resp, err := zc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read and return the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Log that data was successfully fetched.
	log.WithFields(log.Fields{
		"dataLength": len(body),
	}).Debug("Data successfully fetched from the Zabbix")

	return body, nil
}
