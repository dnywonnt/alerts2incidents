package repositories // dnywonnt.me/alerts2incidents/internal/database/repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dnywonnt.me/alerts2incidents/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"

	log "github.com/sirupsen/logrus"
)

// SQL queries as constants for CRUD operations on rules
const (
	// Query for inserting a new rule into the database
	insertRuleQuery = `
		INSERT INTO a2i_rules (
			id, is_muted, description, alerts_summary_conditions, alerts_activity_interval_conditions, 
			incident_life_time, set_incident_summary, set_incident_description, set_incident_departament, 
			set_incident_client_affect, set_incident_is_manageable, set_incident_sale_channels, 
			set_incident_trouble_services, set_incident_failure_type, set_incident_labels, 
			set_incident_is_downtime, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	// Query for selecting a rule by ID
	selectRuleQuery = `
		SELECT id, is_muted, description, alerts_summary_conditions, alerts_activity_interval_conditions, 
			incident_life_time, set_incident_summary, set_incident_description, set_incident_departament, 
			set_incident_client_affect, set_incident_is_manageable, set_incident_sale_channels, 
			set_incident_trouble_services, set_incident_failure_type, set_incident_labels, 
			set_incident_is_downtime, created_at, updated_at
		FROM a2i_rules
		WHERE id = $1
	`

	// Query for updating an existing rule
	updateRuleQuery = `
		UPDATE a2i_rules
		SET is_muted = $1, description = $2, alerts_summary_conditions = $3, alerts_activity_interval_conditions = $4, 
			incident_life_time = $5, set_incident_summary = $6, set_incident_description = $7, 
			set_incident_departament = $8, set_incident_client_affect = $9, set_incident_is_manageable = $10, 
			set_incident_sale_channels = $11, set_incident_trouble_services = $12, set_incident_failure_type = $13, 
			set_incident_labels = $14, set_incident_is_downtime = $15, 
			updated_at = $16
		WHERE id = $17
	`

	// Query for deleting a rule by ID
	deleteRuleQuery = `
		DELETE FROM a2i_rules
		WHERE id = $1
	`
)

// RulesRepository struct defines the structure for the repository
type RulesRepository struct {
	dbPool *pgxpool.Pool // Database pool for PostgreSQL
}

// Constructor for RulesRepository
func NewRulesRepository(dbPool *pgxpool.Pool) *RulesRepository {
	log.Debug("Initializing the rules repository")
	return &RulesRepository{dbPool: dbPool}
}

// CreateRule inserts a new rule into the database
func (rr *RulesRepository) CreateRule(ctx context.Context, rule *models.Rule) error {
	log.WithFields(log.Fields{
		"id": rule.ID,
	}).Debug("Creating a new rule in the database")

	if _, err := rr.dbPool.Exec(
		ctx,
		insertRuleQuery,
		rule.ID, rule.IsMuted, rule.Description, rule.AlertsSummaryConditions, rule.AlertsActivityIntervalConditions,
		rule.IncidentLifeTime, rule.SetIncidentSummary, rule.SetIncidentDescription, rule.SetIncidentDepartament,
		rule.SetIncidentClientAffect, rule.SetIncidentIsManageable, rule.SetIncidentSaleChannels, rule.SetIncidentTroubleServices,
		rule.SetIncidentFailureType, rule.SetIncidentLabels, rule.SetIncidentIsDowntime, rule.CreatedAt, rule.UpdatedAt,
	); err != nil {
		return fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"id": rule.ID,
	}).Debug("The rule has been created in the database")

	return nil
}

// GetRule retrieves a rule by ID from the database
func (rr *RulesRepository) GetRule(ctx context.Context, id string) (*models.Rule, error) {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("Retrieving a rule from the database")

	rule := &models.Rule{}
	if err := rr.dbPool.QueryRow(ctx, selectRuleQuery, id).Scan(
		&rule.ID, &rule.IsMuted, &rule.Description, &rule.AlertsSummaryConditions, &rule.AlertsActivityIntervalConditions,
		&rule.IncidentLifeTime, &rule.SetIncidentSummary, &rule.SetIncidentDescription, &rule.SetIncidentDepartament,
		&rule.SetIncidentClientAffect, &rule.SetIncidentIsManageable, &rule.SetIncidentSaleChannels, &rule.SetIncidentTroubleServices,
		&rule.SetIncidentFailureType, &rule.SetIncidentLabels, &rule.SetIncidentIsDowntime, &rule.CreatedAt, &rule.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Debug("Rule successfully retrieved from the database")

	return rule, nil
}

// UpdateRule updates an existing rule in the database
func (rr *RulesRepository) UpdateRule(ctx context.Context, rule *models.Rule) error {
	log.WithFields(log.Fields{
		"id": rule.ID,
	}).Debug("Updating a rule in the database")

	if _, err := rr.dbPool.Exec(
		ctx,
		updateRuleQuery,
		rule.IsMuted, rule.Description, rule.AlertsSummaryConditions, rule.AlertsActivityIntervalConditions,
		rule.IncidentLifeTime, rule.SetIncidentSummary, rule.SetIncidentDescription, rule.SetIncidentDepartament,
		rule.SetIncidentClientAffect, rule.SetIncidentIsManageable, rule.SetIncidentSaleChannels, rule.SetIncidentTroubleServices,
		rule.SetIncidentFailureType, rule.SetIncidentLabels, rule.SetIncidentIsDowntime, rule.UpdatedAt, rule.ID,
	); err != nil {
		return fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"id": rule.ID,
	}).Debug("The rule has been updated in the database")

	return nil
}

// DeleteRule deletes a rule by ID from the database
func (rr *RulesRepository) DeleteRule(ctx context.Context, id string) error {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("Deleting a rule from the database")

	if _, err := rr.dbPool.Exec(ctx, deleteRuleQuery, id); err != nil {
		return fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Debug("The rule has been deleted from the database")

	return nil
}

// GetRules retrieves rules with filters, sorting, and pagination
func (rr *RulesRepository) GetRules(ctx context.Context, filterBy map[string]interface{}, sortBy string, sortOrder string, pageNum int, pageSize int, startTime time.Time, endTime time.Time) ([]*models.Rule, error) {
	log.WithFields(log.Fields{
		"filterBy":  filterBy,
		"sortBy":    sortBy,
		"sortOrder": sortOrder,
		"pageNum":   pageNum,
		"pageSize":  pageSize,
		"startTime": startTime,
		"endTime":   endTime,
	}).Debug("Retrieving rules with filters, sorting, and pagination from the database")

	query, args := buildGetQueryForRules(filterBy, sortBy, sortOrder, pageNum, pageSize, startTime, endTime)

	rows, err := rr.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing the query: %w", err)
	}
	defer rows.Close()

	rules := []*models.Rule{}
	for rows.Next() {
		rule := &models.Rule{}
		if err := rows.Scan(
			&rule.ID, &rule.IsMuted, &rule.Description, &rule.AlertsSummaryConditions, &rule.AlertsActivityIntervalConditions,
			&rule.IncidentLifeTime, &rule.SetIncidentSummary, &rule.SetIncidentDescription, &rule.SetIncidentDepartament,
			&rule.SetIncidentClientAffect, &rule.SetIncidentIsManageable, &rule.SetIncidentSaleChannels, &rule.SetIncidentTroubleServices,
			&rule.SetIncidentFailureType, &rule.SetIncidentLabels, &rule.SetIncidentIsDowntime, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning the row: %w", err)
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	log.WithFields(log.Fields{
		"rulesCount": len(rules),
	}).Debug("Rules successfully retrieved from the database")

	return rules, nil
}

// GetTotalRules counts the total number of rules with filters
func (rr *RulesRepository) GetTotalRules(ctx context.Context, filterBy map[string]interface{}, startTime, endTime time.Time) (int, error) {
	log.WithFields(log.Fields{
		"filterBy":  filterBy,
		"startTime": startTime,
		"endTime":   endTime,
	}).Debug("Counting the total number of rules with filters")

	query, args := buildCountQueryForRules(filterBy, startTime, endTime)

	totalRules := 0
	if err := rr.dbPool.QueryRow(ctx, query, args...).Scan(&totalRules); err != nil {
		return 0, fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"totalRulesCount": totalRules,
	}).Debug("Total number of rules with filters counted successfully")

	return totalRules, nil
}

// buildGetQueryForRules builds a dynamic query for retrieving rules based on filters and pagination
func buildGetQueryForRules(filterBy map[string]interface{}, sortBy string, sortOrder string, pageNum int, pageSize int, startTime time.Time, endTime time.Time) (string, []interface{}) {
	baseQuery := `SELECT id, is_muted, description, alerts_summary_conditions, alerts_activity_interval_conditions,
		incident_life_time, set_incident_summary, set_incident_description, set_incident_departament,
		set_incident_client_affect, set_incident_is_manageable, set_incident_sale_channels, set_incident_trouble_services,
		set_incident_failure_type, set_incident_labels, set_incident_is_downtime,
		created_at, updated_at FROM a2i_rules WHERE 1 = 1`
	args := make([]interface{}, 0)
	argId := 1

	for field, value := range filterBy {
		if field == "created_at" || field == "updated_at" {
			continue
		}
		switch v := value.(type) {
		case []interface{}:
			placeholders := make([]string, len(v))
			for i := range v {
				placeholders[i] = fmt.Sprintf("$%d", argId)
				args = append(args, v[i])
				argId++
			}
			baseQuery += fmt.Sprintf(" AND %s @> ARRAY[%s]", field, strings.Join(placeholders, ","))
		default:
			baseQuery += fmt.Sprintf(" AND %s = $%d", field, argId)
			args = append(args, value)
			argId++
		}
	}

	if !startTime.IsZero() && !endTime.IsZero() {
		baseQuery += fmt.Sprintf(" AND created_at BETWEEN $%d AND $%d", argId, argId+1)
		args = append(args, startTime, endTime)
		argId += 2
	}

	if sortBy != "" && sortOrder != "" {
		baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	}

	offset := (pageNum - 1) * pageSize
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argId, argId+1)
	args = append(args, pageSize, offset)

	return baseQuery, args
}

// buildCountQueryForRules builds a dynamic query for counting rules based on filters
func buildCountQueryForRules(filterBy map[string]interface{}, startTime, endTime time.Time) (string, []interface{}) {
	baseQuery := "SELECT COUNT(id) FROM a2i_rules WHERE 1 = 1"
	args := make([]interface{}, 0)
	argId := 1

	for field, value := range filterBy {
		if field == "created_at" || field == "updated_at" {
			continue
		}
		switch v := value.(type) {
		case []interface{}:
			placeholders := make([]string, len(v))
			for i := range v {
				placeholders[i] = fmt.Sprintf("$%d", argId)
				args = append(args, v[i])
				argId++
			}
			baseQuery += fmt.Sprintf(" AND %s @> ARRAY[%s]", field, strings.Join(placeholders, ","))
		default:
			baseQuery += fmt.Sprintf(" AND %s = $%d", field, argId)
			args = append(args, value)
			argId++
		}
	}

	if !startTime.IsZero() && !endTime.IsZero() {
		baseQuery += fmt.Sprintf(" AND created_at BETWEEN $%d AND $%d", argId, argId+1)
		args = append(args, startTime, endTime)
		argId += 2
	}

	return baseQuery, args
}
