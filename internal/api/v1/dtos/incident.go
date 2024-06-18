package dtos // dnywonnt.me/alerts2incidents/internal/api/v1/dtos

import "time"

// CreateIncidentDTO is used to receive data from API requests to create a new incident.
type CreateIncidentDTO struct {
	Status          string    `json:"status"`           // Current status of the incident.
	Summary         string    `json:"summary"`          // Summary of the incident.
	Description     string    `json:"description"`      // Detailed description of the incident.
	FromAt          time.Time `json:"from_at"`          // Start time of the incident.
	ToAt            time.Time `json:"to_at"`            // End time of the incident.
	Departament     string    `json:"departament"`      // Department responsible for handling the incident.
	ClientAffect    string    `json:"client_affect"`    // Description of how clients are affected by the incident.
	IsManageable    string    `json:"is_manageable"`    // Indicates if the incident is manageable.
	SaleChannels    []string  `json:"sale_channels"`    // List of sales channels affected by the incident.
	TroubleServices []string  `json:"trouble_services"` // List of services troubled by the incident.
	FinLosses       int       `json:"fin_losses"`       // Estimated financial losses due to the incident.
	FailureType     string    `json:"failure_type"`     // Type of failure associated with the incident.
	IsDeploy        bool      `json:"is_deploy"`        // Indicates if there is a deployment associated with the incident.
	DeployLink      string    `json:"deploy_link"`      // Link to deployment details or documentation.
	Labels          []string  `json:"labels"`           // Labels associated with the incident for categorization.
	IsDowntime      bool      `json:"is_downtime"`      // Indicates if the incident causes downtime.
	PostmortemLink  string    `json:"postmortem_link"`  // Link to the postmortem report if available.
	Creator         string    `json:"creator"`          // Identifier of the user creating the incident.
}

// UpdateIncidentDTO is used to receive data from API requests to update an existing incident.
type UpdateIncidentDTO struct {
	Status           *string    `json:"status,omitempty"`            // Optional updated status of the incident.
	Summary          *string    `json:"summary,omitempty"`           // Optional updated summary of the incident.
	Description      *string    `json:"description,omitempty"`       // Optional updated detailed description.
	FromAt           *time.Time `json:"from_at,omitempty"`           // Optional updated start time of the incident.
	ToAt             *time.Time `json:"to_at,omitempty"`             // Optional updated end time of the incident.
	IsConfirmed      *bool      `json:"is_confirmed,omitempty"`      // Optional updated confirmation status.
	ConfirmationTime *time.Time `json:"confirmation_time,omitempty"` // Optional updated confirmation time.
	Departament      *string    `json:"departament,omitempty"`       // Optional updated department responsible.
	ClientAffect     *string    `json:"client_affect,omitempty"`     // Optional updated description of client affect.
	IsManageable     *string    `json:"is_manageable,omitempty"`     // Optional updated manageability status.
	SaleChannels     *[]string  `json:"sale_channels,omitempty"`     // Optional updated list of affected sale channels.
	TroubleServices  *[]string  `json:"trouble_services,omitempty"`  // Optional updated list of troubled services.
	FinLosses        *int       `json:"fin_losses,omitempty"`        // Optional updated estimate of financial losses.
	FailureType      *string    `json:"failure_type,omitempty"`      // Optional updated type of failure.
	IsDeploy         *bool      `json:"is_deploy,omitempty"`         // Optional updated deployment status.
	DeployLink       *string    `json:"deploy_link,omitempty"`       // Optional updated link to deployment details.
	Labels           *[]string  `json:"labels,omitempty"`            // Optional updated list of labels.
	IsDowntime       *bool      `json:"is_downtime,omitempty"`       // Optional updated downtime status.
	PostmortemLink   *string    `json:"postmortem_link,omitempty"`   // Optional updated link to the postmortem report.
}
