package models // dnywonnt.me/alerts2incidents/internal/models

import (
	"time"

	"dnywonnt.me/alerts2incidents/internal/utils"
)

// Incident represents an incident data structure.
type Incident struct {
	ID               string    `json:"id" validate:"required"`                                                                                                                                             // Unique identifier for the incident
	Type             string    `json:"type" validate:"required,oneof=manual auto"`                                                                                                                         // Type of incident, either 'manual' or 'auto'
	Status           string    `json:"status" validate:"required,oneof=actual finished closed"`                                                                                                            // Current status of the incident
	Summary          string    `json:"summary" validate:"required"`                                                                                                                                        // Brief summary of the incident
	Description      string    `json:"description" validate:"omitempty"`                                                                                                                                   // Detailed description of the incident
	FromAt           time.Time `json:"from_at" validate:"required"`                                                                                                                                        // Start time of the incident
	ToAt             time.Time `json:"to_at" validate:"required_if_m=Status finished closed|gtefield=FromAt"`                                                                                              // End time of the incident, required if status is 'finished' or 'closed'
	IsConfirmed      bool      `json:"is_confirmed" validate:"-"`                                                                                                                                          // Flag indicating if the incident is confirmed
	ConfirmationTime time.Time `json:"confirmation_time" validate:"required_with=IsConfirmed|gtefield=FromAt,ltefield=ToAt"`                                                                               // Time of confirmation, required if IsConfirmed is true
	Quarter          int       `json:"quarter" validate:"gte=1,lte=4"`                                                                                                                                     // Quarter in which the incident occurred, value between 1 and 4
	Departament      string    `json:"departament" validate:"required,oneof=internal_digital internal_it external_service"`                                                                                // Department affected by the incident
	ClientAffect     string    `json:"client_affect" validate:"omitempty"`                                                                                                                                 // Details on how clients are affected
	IsManageable     string    `json:"is_manageable" validate:"required,oneof=yes no indirectly"`                                                                                                          // Flag indicating if the incident is manageable
	SaleChannels     []string  `json:"sale_channels" validate:"required,min=1"`                                                                                                                            // Sales channels affected by the incident
	TroubleServices  []string  `json:"trouble_services" validate:"required,min=1"`                                                                                                                         // Services troubled by the incident
	FinLosses        int       `json:"fin_losses" validate:"gte=0"`                                                                                                                                        // Financial losses incurred due to the incident
	FailureType      string    `json:"failure_type" validate:"required,oneof=err_network err_acquiring err_development err_security err_infrastructure err_configuration err_menu err_external err_other"` // Type of failure that caused the incident
	IsDeploy         bool      `json:"is_deploy" validate:"-"`                                                                                                                                             // Flag indicating if the incident involves deployment
	DeployLink       string    `json:"deploy_link" validate:"required_with=IsDeploy|url"`                                                                                                                  // Link to deployment details, required if IsDeploy is true
	Labels           []string  `json:"labels" validate:"required"`                                                                                                                                         // Labels associated with the incident
	IsDowntime       bool      `json:"is_downtime" validate:"-"`                                                                                                                                           // Flag indicating if there was downtime
	PostmortemLink   string    `json:"postmortem_link" validate:"omitempty"`                                                                                                                               // Link to postmortem report
	Creator          string    `json:"creator" validate:"required"`                                                                                                                                        // Creator of the incident record
	RuleID           *string   `json:"rule_id" validate:"required_if=Type auto"`                                                                                                                           // Rule ID, required if the type is 'auto'
	MatchingCount    int       `json:"matching_count" validate:"required_if=Type auto"`                                                                                                                    // Count of matches, required if the type is 'auto'
	LastMatchingTime time.Time `json:"last_matching_time" validate:"required_if=Type auto|gtefield=FromAt"`                                                                                                // Time of the last match, required if the type is 'auto'
	AlertsData       string    `json:"alerts_data" validate:"required_if=Type auto|json"`                                                                                                                  // Alerts data in JSON format, required if the type is 'auto'
	CreatedAt        time.Time `json:"created_at" validate:"required"`                                                                                                                                     // Timestamp when the incident was created
	UpdatedAt        time.Time `json:"updated_at" validate:"required"`                                                                                                                                     // Timestamp when the incident was last updated
}

// Validate runs validation rules on an Incident instance.
func (i *Incident) Validate() error {
	return utils.ValidateStruct(i)
}
