package dtos // dnywonnt.me/alerts2incidents/internal/api/v1/dtos

import "time"

// CreateRuleDTO is used to capture incoming data from API requests to create a new rule.
type CreateRuleDTO struct {
	IsMuted                          bool            `json:"is_muted"`                            // Indicates if the rule should be muted.
	Description                      string          `json:"description"`                         // Description of the rule.
	AlertsSummaryConditions          []string        `json:"alerts_summary_conditions"`           // Conditions that summarize alerts.
	AlertsActivityIntervalConditions []time.Duration `json:"alerts_activity_interval_conditions"` // List of time durations for alert activity intervals.
	IncidentLifeTime                 time.Duration   `json:"incident_life_time"`                  // Duration for which an incident is considered active.
	IncidentFinishingInterval        time.Duration   `json:"incident_finishing_interval"`         // Time interval after which the incident is considered finished.
	SetIncidentSummary               string          `json:"set_incident_summary"`                // Summary to be set for an incident.
	SetIncidentDescription           string          `json:"set_incident_description"`            // Description to be set for an incident.
	SetIncidentDepartament           string          `json:"set_incident_departament"`            // Department responsible for the incident.
	SetIncidentClientAffect          string          `json:"set_incident_client_affect"`          // Description of how the incident affects clients.
	SetIncidentIsManageable          string          `json:"set_incident_is_manageable"`          // Indicates if the incident is manageable.
	SetIncidentSaleChannels          []string        `json:"set_incident_sale_channels"`          // Sale channels affected by the incident.
	SetIncidentTroubleServices       []string        `json:"set_incident_trouble_services"`       // Services troubled by the incident.
	SetIncidentFailureType           string          `json:"set_incident_failure_type"`           // Type of failure associated with the incident.
	SetIncidentLabels                []string        `json:"set_incident_labels"`                 // Labels associated with the incident.
	SetIncidentIsDowntime            bool            `json:"set_incident_is_downtime"`            // Indicates if the incident causes downtime.
}

// UpdateRuleDTO is used to capture incoming data from API requests to update an existing rule.
type UpdateRuleDTO struct {
	IsMuted                          *bool            `json:"is_muted,omitempty"`                            // Optional update to the mute status.
	Description                      *string          `json:"description,omitempty"`                         // Optional update to the rule's description.
	AlertsSummaryConditions          *[]string        `json:"alerts_summary_conditions,omitempty"`           // Optional update to the conditions that summarize alerts.
	AlertsActivityIntervalConditions *[]time.Duration `json:"alerts_activity_interval_conditions,omitempty"` // Optional update to the list of time durations for alert activity intervals.
	IncidentLifeTime                 *time.Duration   `json:"incident_life_time,omitempty"`                  // Optional update to the duration for which an incident is considered active.
	IncidentFinishingInterval        *time.Duration   `json:"incident_finishing_interval,omitempty"`         // Optional update to the time interval after which the incident is considered finished.
	SetIncidentSummary               *string          `json:"set_incident_summary,omitempty"`                // Optional update to the summary set for an incident.
	SetIncidentDescription           *string          `json:"set_incident_description,omitempty"`            // Optional update to the description set for an incident.
	SetIncidentDepartament           *string          `json:"set_incident_departament,omitempty"`            // Optional update to the department responsible for the incident.
	SetIncidentClientAffect          *string          `json:"set_incident_client_affect,omitempty"`          // Optional update to how the incident affects clients.
	SetIncidentIsManageable          *string          `json:"set_incident_is_manageable,omitempty"`          // Optional update to whether the incident is manageable.
	SetIncidentSaleChannels          *[]string        `json:"set_incident_sale_channels,omitempty"`          // Optional update to the sale channels affected by the incident.
	SetIncidentTroubleServices       *[]string        `json:"set_incident_trouble_services,omitempty"`       // Optional update to the services troubled by the incident.
	SetIncidentFailureType           *string          `json:"set_incident_failure_type,omitempty"`           // Optional update to the type of failure associated with the incident.
	SetIncidentLabels                *[]string        `json:"set_incident_labels,omitempty"`                 // Optional update to the labels associated with the incident.
	SetIncidentIsDowntime            *bool            `json:"set_incident_is_downtime,omitempty"`            // Optional update to whether the incident causes downtime.
}
