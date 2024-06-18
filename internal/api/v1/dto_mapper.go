package v1 // dnywonnt.me/alerts2incidents/internal/api/v1

import (
	"errors"
	"time"

	"dnywonnt.me/alerts2incidents/internal/api/v1/dtos"
	"dnywonnt.me/alerts2incidents/internal/models"
	"dnywonnt.me/alerts2incidents/internal/utils"
	"github.com/google/uuid"
)

// updateField updates a destination field with the source value if the source is not nil.
func updateField[T any](src, dst *T, updated *bool) {
	if src != nil {
		*dst = *src
		*updated = true
	}
}

// updateTimeField updates a destination time field with the source value in UTC if the source is not nil.
func updateTimeField(src, dst *time.Time, updated *bool) {
	if src != nil {
		*dst = src.UTC()
		*updated = true
	}
}

// MapCreateIncidentDTOToModel converts a DTO from the API input into an Incident model for further processing.
func MapCreateIncidentDTOToModel(dto *dtos.CreateIncidentDTO) (*models.Incident, error) {
	currentTimeUTC := time.Now().UTC() // Get the current time in UTC.
	zeroTime := time.Time{}

	incident := &models.Incident{
		ID:               uuid.NewString(),          // Generate a new unique ID.
		Type:             "manual",                  // Set the type to manual as default.
		Status:           dto.Status,                // Map status from DTO.
		Summary:          dto.Summary,               // Map summary from DTO.
		Description:      dto.Description,           // Map description from DTO.
		FromAt:           dto.FromAt.UTC(),          // Ensure date is in UTC.
		ToAt:             dto.ToAt.UTC(),            // Ensure date is in UTC.
		IsConfirmed:      false,                     // Map confirmation status.
		ConfirmationTime: zeroTime,                  // Ensure date is in UTC.
		Quarter:          utils.GetCurrentQuarter(), // Calculate the current quarter.
		Departament:      dto.Departament,           // Map department from DTO.
		ClientAffect:     dto.ClientAffect,          // Map client affect from DTO.
		IsManageable:     dto.IsManageable,          // Map manageability status.
		SaleChannels:     dto.SaleChannels,          // Map sale channels.
		TroubleServices:  dto.TroubleServices,       // Map troubled services.
		FinLosses:        dto.FinLosses,             // Map financial losses.
		FailureType:      dto.FailureType,           // Map failure type.
		IsDeploy:         dto.IsDeploy,              // Map deploy status.
		DeployLink:       dto.DeployLink,            // Map deploy link.
		Labels:           dto.Labels,                // Map labels.
		IsDowntime:       dto.IsDowntime,            // Map downtime status.
		PostmortemLink:   dto.PostmortemLink,        // Map postmortem link.
		Creator:          dto.Creator,               // Map creator.
		RuleID:           nil,                       // Initialize RuleID as nil.
		MatchingCount:    0,                         // Initialize matching count as 0.
		LastMatchingTime: zeroTime,                  // Initialize last matching time.
		AlertsData:       "",                        // Initialize alerts data.
		CreatedAt:        currentTimeUTC,            // Set creation time.
		UpdatedAt:        currentTimeUTC,            // Set update time.
	}

	// Validate the newly created incident model.
	if err := incident.Validate(); err != nil {
		return nil, err
	}

	return incident, nil
}

// MapUpdateIncidentDTOToModel updates an existing incident model with data from an UpdateIncidentDTO.
func MapUpdateIncidentDTOToModel(dto *dtos.UpdateIncidentDTO, incident *models.Incident) error {
	anyFieldUpdated := false // Track if any field has been updated.

	// Update each field from the DTO if provided, and mark the record as updated.
	updateField(dto.Status, &incident.Status, &anyFieldUpdated)
	updateField(dto.Summary, &incident.Summary, &anyFieldUpdated)
	updateField(dto.Description, &incident.Description, &anyFieldUpdated)
	updateTimeField(dto.FromAt, &incident.FromAt, &anyFieldUpdated)
	updateTimeField(dto.ToAt, &incident.ToAt, &anyFieldUpdated)
	updateField(dto.IsConfirmed, &incident.IsConfirmed, &anyFieldUpdated)
	updateTimeField(dto.ConfirmationTime, &incident.ConfirmationTime, &anyFieldUpdated)
	updateField(dto.Departament, &incident.Departament, &anyFieldUpdated)
	updateField(dto.ClientAffect, &incident.ClientAffect, &anyFieldUpdated)
	updateField(dto.IsManageable, &incident.IsManageable, &anyFieldUpdated)
	updateField(dto.SaleChannels, &incident.SaleChannels, &anyFieldUpdated)
	updateField(dto.TroubleServices, &incident.TroubleServices, &anyFieldUpdated)
	updateField(dto.FinLosses, &incident.FinLosses, &anyFieldUpdated)
	updateField(dto.FailureType, &incident.FailureType, &anyFieldUpdated)
	updateField(dto.IsDeploy, &incident.IsDeploy, &anyFieldUpdated)
	updateField(dto.DeployLink, &incident.DeployLink, &anyFieldUpdated)
	updateField(dto.Labels, &incident.Labels, &anyFieldUpdated)
	updateField(dto.IsDowntime, &incident.IsDowntime, &anyFieldUpdated)
	updateField(dto.PostmortemLink, &incident.PostmortemLink, &anyFieldUpdated)

	// If no fields were updated, return an error.
	if !anyFieldUpdated {
		return errors.New("no fields provided for update")
	} else {
		// Validate the updated incident and update the timestamp.
		if err := incident.Validate(); err != nil {
			return err
		}
		incident.UpdatedAt = time.Now().UTC()
	}

	return nil
}

// MapCreateRuleDTOToModel converts a CreateRuleDTO into a Rule model, setting fields and performing initial validation.
func MapCreateRuleDTOToModel(dto *dtos.CreateRuleDTO) (*models.Rule, error) {
	currentTime := time.Now().UTC() // Capture the current time in UTC for timestamps.

	rule := &models.Rule{
		ID:                               uuid.NewString(),                     // Generate a new unique ID for the rule.
		IsMuted:                          dto.IsMuted,                          // Map the mute status from DTO.
		Description:                      dto.Description,                      // Map the description from DTO.
		AlertsSummaryConditions:          dto.AlertsSummaryConditions,          // Map summary conditions from DTO.
		AlertsActivityIntervalConditions: dto.AlertsActivityIntervalConditions, // Map activity interval conditions from DTO.
		IncidentLifeTime:                 dto.IncidentLifeTime,                 // Map the incident lifetime from DTO.
		SetIncidentSummary:               dto.SetIncidentSummary,               // Map the incident summary to be set from DTO.
		SetIncidentDescription:           dto.SetIncidentDescription,           // Map the incident description to be set from DTO.
		SetIncidentDepartament:           dto.SetIncidentDepartament,           // Map the department from DTO.
		SetIncidentClientAffect:          dto.SetIncidentClientAffect,          // Map how the incident affects clients from DTO.
		SetIncidentIsManageable:          dto.SetIncidentIsManageable,          // Map the manageability status from DTO.
		SetIncidentSaleChannels:          dto.SetIncidentSaleChannels,          // Map the affected sale channels from DTO.
		SetIncidentTroubleServices:       dto.SetIncidentTroubleServices,       // Map the troubled services from DTO.
		SetIncidentFailureType:           dto.SetIncidentFailureType,           // Map the type of failure from DTO.
		SetIncidentLabels:                dto.SetIncidentLabels,                // Map the incident labels from DTO.
		SetIncidentIsDowntime:            dto.SetIncidentIsDowntime,            // Map the downtime status from DTO.
		CreatedAt:                        currentTime,                          // Set the creation time.
		UpdatedAt:                        currentTime,                          // Set the update time.
	}

	// Validate the newly created rule model.
	if err := rule.Validate(); err != nil {
		return nil, err // Return error if validation fails.
	}

	return rule, nil // Return the new rule if validation passes.
}

// MapUpdateRuleDTOToModel updates an existing rule model with data from an UpdateRuleDTO.
func MapUpdateRuleDTOToModel(dto *dtos.UpdateRuleDTO, rule *models.Rule) error {
	anyFieldUpdated := false // Track if any field has been updated.

	// Update each field from the DTO if provided, and mark the record as updated.
	updateField(dto.IsMuted, &rule.IsMuted, &anyFieldUpdated)
	updateField(dto.Description, &rule.Description, &anyFieldUpdated)
	updateField(dto.AlertsSummaryConditions, &rule.AlertsSummaryConditions, &anyFieldUpdated)
	updateField(dto.AlertsActivityIntervalConditions, &rule.AlertsActivityIntervalConditions, &anyFieldUpdated)
	updateField(dto.IncidentLifeTime, &rule.IncidentLifeTime, &anyFieldUpdated)
	updateField(dto.SetIncidentSummary, &rule.SetIncidentSummary, &anyFieldUpdated)
	updateField(dto.SetIncidentDescription, &rule.SetIncidentDescription, &anyFieldUpdated)
	updateField(dto.SetIncidentDepartament, &rule.SetIncidentDepartament, &anyFieldUpdated)
	updateField(dto.SetIncidentClientAffect, &rule.SetIncidentClientAffect, &anyFieldUpdated)
	updateField(dto.SetIncidentIsManageable, &rule.SetIncidentIsManageable, &anyFieldUpdated)
	updateField(dto.SetIncidentSaleChannels, &rule.SetIncidentSaleChannels, &anyFieldUpdated)
	updateField(dto.SetIncidentTroubleServices, &rule.SetIncidentTroubleServices, &anyFieldUpdated)
	updateField(dto.SetIncidentFailureType, &rule.SetIncidentFailureType, &anyFieldUpdated)
	updateField(dto.SetIncidentLabels, &rule.SetIncidentLabels, &anyFieldUpdated)
	updateField(dto.SetIncidentIsDowntime, &rule.SetIncidentIsDowntime, &anyFieldUpdated)

	// If no fields were updated, return an error indicating no update was made.
	if !anyFieldUpdated {
		return errors.New("no fields provided for update")
	} else {
		// Validate the updated rule and update the timestamp.
		if err := rule.Validate(); err != nil {
			return err
		}
		rule.UpdatedAt = time.Now().UTC()
	}

	return nil
}
