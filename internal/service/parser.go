package service // dnywonnt.me/alerts2incidents/internal/service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"dnywonnt.me/alerts2incidents/internal/config"
	"dnywonnt.me/alerts2incidents/internal/models"

	log "github.com/sirupsen/logrus"
)

// AlertsParser is a struct that holds configuration details for parsing alerts.
type AlertsParser struct {
	cfg *config.AlertsParserConfig
}

// NewAlertsParser initializes a new instance of AlertsParser with the provided configuration.
func NewAlertsParser(cfg *config.AlertsParserConfig) *AlertsParser {
	log.Debug("Initializing the alerts parser")
	return &AlertsParser{cfg: cfg}
}

// ParseAndAggregateAlerts listens for incoming alert data and aggregates them over a specified interval.
func (ap *AlertsParser) ParseAndAggregateAlerts(ctx context.Context, dataCh <-chan map[CollectorType][]byte, alertsCh chan<- []models.Alert) {
	log.WithFields(log.Fields{
		"aggregationInterval": ap.cfg.AggregationInterval.String(),
	}).Debug("Starting the alerts parsing and aggregation process")

	// Create a ticker to trigger aggregation at the specified interval.
	aggregationTicker := time.NewTicker(ap.cfg.AggregationInterval)
	defer aggregationTicker.Stop()

	// Initialize a slice to store aggregated alerts.
	aggregatedAlerts := []models.Alert{}

	for {
		select {
		case <-ctx.Done():
			// If the context is cancelled, exit the aggregation loop.
			log.Debug("Stopping alerts parsing and aggregation process")
			return

		case data := <-dataCh:
			// Process incoming data for each collector.
			for collector, jsonData := range data {
				parsedAlerts := []models.Alert{}

				// Parse alerts based on the collector type.
				switch collector {
				case GrafanaAMCollector:
					// Parse Grafana Alertmanager alerts.
					grafanaAMAlerts, err := ap.parseGrafanaAMAlerts(jsonData)
					if err != nil {
						log.WithFields(log.Fields{
							"error": err.Error(),
						}).Error("Failed to parse Grafana Alertmanager alerts")
						continue
					}
					parsedAlerts = append(parsedAlerts, grafanaAMAlerts...)

				case GrafanaPrometheusCollector:
					// Parse Grafana Prometheus alerts.
					grafanaPrometheusAlerts, err := ap.parseGrafanaPrometheusAlerts(jsonData)
					if err != nil {
						log.WithFields(log.Fields{
							"error": err.Error(),
						}).Error("Failed to parse Grafana Prometheus alerts")
						continue
					}
					parsedAlerts = append(parsedAlerts, grafanaPrometheusAlerts...)

				case ZabbixCollector:
					// Parse Zabbix alerts.
					zabbixAlerts, err := parseZabbixAlerts(jsonData)
					if err != nil {
						log.WithFields(log.Fields{
							"error": err.Error(),
						}).Error("Failed to parse Zabbix alerts")
						continue
					}
					parsedAlerts = append(parsedAlerts, zabbixAlerts...)

				default:
					// Log a warning if the collector type is unrecognized.
					log.WithFields(log.Fields{
						"collector": collector,
					}).Warn("Received data from an unknown collector")
					continue
				}

				// Aggregate parsed alerts.
				aggregatedAlerts = append(aggregatedAlerts, parsedAlerts...)
				log.WithFields(log.Fields{
					"currentTotal":  len(aggregatedAlerts),
					"recentlyAdded": len(parsedAlerts),
				}).Debug("Alerts have been aggregated")
			}

		case <-aggregationTicker.C:
			// Periodically send aggregated alerts through the channel.
			if len(aggregatedAlerts) > 0 {
				select {
				case alertsCh <- aggregatedAlerts:
					log.WithFields(log.Fields{
						"alertsCount":       len(aggregatedAlerts),
						"alertsChannelSize": len(alertsCh),
					}).Debug("Aggregated alerts have been sent to the channel")
					aggregatedAlerts = []models.Alert{}
				default:
					// If the channel is blocked, log a warning and skip sending.
					log.WithFields(log.Fields{
						"alertsCount":       len(aggregatedAlerts),
						"alertsChannelSize": len(alertsCh),
					}).Warn("The alerts channel is blocked; skipping sending of alerts")
				}
			}
		}
	}
}

// parseGrafanaAMAlerts parses alerts from Grafana Alertmanager's JSON data.
func (ap AlertsParser) parseGrafanaAMAlerts(jsonData []byte) ([]models.Alert, error) {
	log.WithFields(log.Fields{
		"parseField": ap.cfg.GrafanaAMParseField,
	}).Debug("Parsing Grafana Alertmanager alerts")

	alerts := []models.Alert{}

	// Define the structure to which the JSON data will be unmarshaled.
	var responseData []struct {
		Annotations struct {
			Summary     *string `json:"summary,omitempty"`
			Description *string `json:"description,omitempty"`
		} `json:"annotations"`
		StartsAt string `json:"startsAt"`
	}

	// Unmarshal the JSON data into the predefined structure.
	if err := json.Unmarshal(jsonData, &responseData); err != nil {
		return nil, fmt.Errorf("error unmarshaling response data: %w", err)
	}

	// Convert each alert from the JSON structure to the Alert model.
	for _, grafanaAMAlert := range responseData {
		startsAtTime, err := time.Parse(time.RFC3339, grafanaAMAlert.StartsAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing time for alert: %w", err)
		}

		// Determine which field to use based on the configuration.
		var summary string
		switch ap.cfg.GrafanaAMParseField {
		case "summary":
			if grafanaAMAlert.Annotations.Summary != nil {
				summary = *grafanaAMAlert.Annotations.Summary
			}

		case "description":
			if grafanaAMAlert.Annotations.Description != nil {
				summary = *grafanaAMAlert.Annotations.Description
			}
		}

		alerts = append(alerts, models.Alert{
			Summary:   summary,
			CreatedAt: startsAtTime.UTC(),
		})
	}

	log.WithFields(log.Fields{
		"parseField":  ap.cfg.GrafanaAMParseField,
		"alertsCount": len(alerts),
	}).Debug("Grafana Alertmanager alerts have been successfully parsed")

	return alerts, nil
}

// parseGrafanaPrometheusAlerts parses alerts from Grafana Prometheus's JSON data.
func (ap *AlertsParser) parseGrafanaPrometheusAlerts(jsonData []byte) ([]models.Alert, error) {
	log.WithFields(log.Fields{
		"parseField": ap.cfg.GrafanaPrometheusParseField,
	}).Debug("Parsing Grafana Prometheus alerts")

	alerts := []models.Alert{}

	// Define the structure to which the JSON data will be unmarshaled.
	var responseData struct {
		Data struct {
			Alerts []struct {
				Annotations struct {
					Summary     *string `json:"summary,omitempty"`
					Description *string `json:"description,omitempty"`
				} `json:"annotations"`
				State    string `json:"state"`
				ActiveAt string `json:"activeAt"`
			} `json:"alerts"`
		} `json:"data"`
	}

	// Unmarshal the JSON data into the predefined structure.
	if err := json.Unmarshal(jsonData, &responseData); err != nil {
		return nil, fmt.Errorf("error unmarshaling response data: %w", err)
	}

	// Convert each firing alert from the JSON structure to the Alert model.
	for _, grafanaPrometheusAlert := range responseData.Data.Alerts {
		if grafanaPrometheusAlert.State != "firing" {
			continue
		}

		activeAtTime, err := time.Parse(time.RFC3339, grafanaPrometheusAlert.ActiveAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing time for alert: %w", err)
		}

		// Determine which field to use based on the configuration.
		var summary string
		switch ap.cfg.GrafanaPrometheusParseField {
		case "summary":
			if grafanaPrometheusAlert.Annotations.Summary != nil {
				summary = *grafanaPrometheusAlert.Annotations.Summary
			}

		case "description":
			if grafanaPrometheusAlert.Annotations.Description != nil {
				summary = *grafanaPrometheusAlert.Annotations.Description
			}
		}

		alerts = append(alerts, models.Alert{
			Summary:   summary,
			CreatedAt: activeAtTime.UTC(),
		})
	}

	log.WithFields(log.Fields{
		"parseField":  ap.cfg.GrafanaPrometheusParseField,
		"alertsCount": len(alerts),
	}).Debug("Grafana Prometheus alerts have been successfully parsed")

	return alerts, nil
}

// parseZabbixAlerts parses alerts from Zabbix's JSON data.
func parseZabbixAlerts(jsonData []byte) ([]models.Alert, error) {
	log.WithFields(log.Fields{}).Debug("Parsing Zabbix alerts")

	alerts := []models.Alert{}

	// Define the structure to which the JSON data will be unmarshaled.
	var responseData struct {
		Result []struct {
			Description string `json:"description"`
			LastChange  string `json:"lastchange"`
			Hosts       []struct {
				Host string `json:"host"`
			} `json:"hosts"`
		} `json:"result"`
	}

	// Unmarshal the JSON data into the predefined structure.
	if err := json.Unmarshal(jsonData, &responseData); err != nil {
		return nil, fmt.Errorf("error unmarshaling response data: %w", err)
	}

	// Convert each alert from the JSON structure to the Alert model.
	for _, zabbixTrigger := range responseData.Result {
		lastChangeTime, err := strconv.ParseInt(zabbixTrigger.LastChange, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing time for alert: %w", err)
		}
		lastChange := time.Unix(lastChangeTime, 0)

		alerts = append(alerts, models.Alert{
			Summary:   fmt.Sprintf("[%s] %s", zabbixTrigger.Hosts[0].Host, zabbixTrigger.Description),
			CreatedAt: lastChange.UTC(),
		})
	}

	log.WithFields(log.Fields{
		"alertsCount": len(alerts),
	}).Debug("Zabbix alerts have been successfully parsed")

	return alerts, nil
}
