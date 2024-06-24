package models // dnywonnt.me/alerts2incidents/internal/models

import (
	"fmt"
	"time"

	"dnywonnt.me/alerts2incidents/internal/utils"
)

// Rule defines the structure for rules that dictate how alerts translate into incidents.
type Rule struct {
	ID                               string          `json:"id" validate:"required"`                                                                                                                                                          // Unique identifier for the rule, mandatory.
	IsMuted                          bool            `json:"is_muted" validate:"-"`                                                                                                                                                           // Indicates whether the rule is currently muted.
	Description                      string          `json:"description" validate:"omitempty"`                                                                                                                                                // Optional description of the rule.
	AlertsSummaryConditions          []string        `json:"alerts_summary_conditions" validate:"required,min=1"`                                                                                                                             // Conditions under which alerts are summarized, at least one condition is required.
	AlertsActivityIntervalConditions []time.Duration `json:"alerts_activity_interval_conditions" validate:"required,min=1"`                                                                                                                   // Time intervals for monitoring alert activity, at least one interval is required.
	IncidentLifeTime                 time.Duration   `json:"incident_life_time" validate:"omitempty"`                                                                                                                                         // Optional duration for which an incident is considered active.
	IncidentFinishingInterval        time.Duration   `json:"incident_finishing_interval" validate:"required,min=1m"`                                                                                                                          // Duration after which an incident is considered finished.
	SetIncidentSummary               string          `json:"set_incident_summary" validate:"required"`                                                                                                                                        // Mandatory summary description for the incident.
	SetIncidentDescription           string          `json:"set_incident_description" validate:"omitempty"`                                                                                                                                   // Optional detailed description of the incident.
	SetIncidentDepartament           string          `json:"set_incident_departament" validate:"required,oneof=internal_digital internal_it external_service"`                                                                                // Department responsible for handling the incident, required.
	SetIncidentClientAffect          string          `json:"set_incident_client_affect" validate:"omitempty"`                                                                                                                                 // Optional description of how the incident affects clients.
	SetIncidentIsManageable          string          `json:"set_incident_is_manageable" validate:"required,oneof=yes no indirectly"`                                                                                                          // Required field indicating if the incident is manageable, valid values are 'yes', 'no', 'indirectly'.
	SetIncidentSaleChannels          []string        `json:"set_incident_sale_channels" validate:"required,min=1"`                                                                                                                            // Sales channels affected by the incident, requires at least one channel.
	SetIncidentTroubleServices       []string        `json:"set_incident_trouble_services" validate:"required,min=1"`                                                                                                                         // Services troubled by the incident, requires at least one service.
	SetIncidentFailureType           string          `json:"set_incident_failure_type" validate:"required,oneof=err_network err_acquiring err_development err_security err_infrastructure err_configuration err_menu err_external err_other"` // Specifies the type of failure that caused the incident. This field is required and must be one of the predefined error types: 'err_network', 'err_acquiring', 'err_development', 'err_security', 'err_infrastructure', 'err_configuration', 'err_menu', 'err_external', or 'err_other'.
	SetIncidentLabels                []string        `json:"set_incident_labels" validate:"required"`                                                                                                                                         // Labels associated with the incident, requires at least one label.
	SetIncidentIsDowntime            bool            `json:"set_incident_is_downtime" validate:"-"`                                                                                                                                           // Indicates if the incident causes downtime.
	CreatedAt                        time.Time       `json:"created_at" validate:"required"`                                                                                                                                                  // Timestamp of when the rule was created, required.
	UpdatedAt                        time.Time       `json:"updated_at" validate:"required"`                                                                                                                                                  // Timestamp of the last update to the rule, required.
}

// Validate performs custom validation on the Rule struct.
func (r *Rule) Validate() error {
	// Ensure the number of summary conditions matches the number of activity intervals.
	if len(r.AlertsSummaryConditions) != len(r.AlertsActivityIntervalConditions) {
		return fmt.Errorf("mismatch in number of summary conditions (%d) and activity intervals (%d)",
			len(r.AlertsSummaryConditions), len(r.AlertsActivityIntervalConditions))
	}
	// Use the validator library to validate the struct according to tags.
	return utils.ValidateStruct(r)
}
