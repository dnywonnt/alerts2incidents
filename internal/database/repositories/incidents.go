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

// SQL queries as constants for code cleanliness and maintainability.
const (
	// insertIncidentQuery represents an SQL query for inserting a new incident into the database.
	insertIncidentQuery = `
		INSERT INTO a2i_incidents (
		    id, type, status, summary, description, from_at, to_at, is_confirmed, confirmation_time,
		    quarter, departament, client_affect, is_manageable, sale_channels, trouble_services,
		    fin_losses, failure_type, is_deploy, deploy_link, labels, is_downtime,
		    postmortem_link, creator, rule_id, matching_count, last_matching_time, alerts_data,
		    created_at, updated_at
		) VALUES (
		    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
		    $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
		)
	`

	// selectIncidentQuery represents an SQL query for selecting an incident by ID from the database.
	selectIncidentQuery = `
		SELECT id, type, status, summary, description, from_at, to_at, is_confirmed, confirmation_time,
		    quarter, departament, client_affect, is_manageable, sale_channels, trouble_services,
		    fin_losses, failure_type, is_deploy, deploy_link, labels, is_downtime,
		    postmortem_link, creator, rule_id, matching_count, last_matching_time, alerts_data,
		    created_at, updated_at
		FROM a2i_incidents
		WHERE id = $1
	`

	// updateIncidentQuery represents an SQL query for updating an existing incident in the database.
	updateIncidentQuery = `
		UPDATE a2i_incidents
		SET status = $1, summary = $2, description = $3, from_at = $4, to_at = $5, is_confirmed = $6,
		    confirmation_time = $7, departament = $8, client_affect = $9, is_manageable = $10,
		    sale_channels = $11, trouble_services = $12, fin_losses = $13, failure_type = $14, is_deploy = $15,
		    deploy_link = $16, labels = $17, is_downtime = $18, postmortem_link = $19, 
		    matching_count = $20, last_matching_time = $21, updated_at = $22
		WHERE id = $23
	`

	// deleteIncidentQuery represents an SQL query for deleting an incident from the database by ID.
	deleteIncidentQuery = `
		DELETE FROM a2i_incidents
		WHERE id = $1
	`
)

// IncidentsRepository defines a repository for managing incidents with database operations.
type IncidentsRepository struct {
	dbPool *pgxpool.Pool // dbPool is a pool of database connections handled by pgxpool.
}

// NewIncidentsRepository creates a new instance of IncidentsRepository.
// This constructor function initializes the repository with a connection pool.
func NewIncidentsRepository(dbPool *pgxpool.Pool) *IncidentsRepository {
	log.Debug("Initializing the incidents repository")
	return &IncidentsRepository{dbPool: dbPool}
}

// CreateIncident handles the creation of a new incident in the database.
// This method logs the creation process and executes an SQL query to insert the incident data.
func (ir *IncidentsRepository) CreateIncident(ctx context.Context, incident *models.Incident) error {
	log.WithFields(log.Fields{
		"id": incident.ID,
	}).Debug("Creating a new incident in the database")

	if _, err := ir.dbPool.Exec(
		ctx,
		insertIncidentQuery,
		incident.ID, incident.Type, incident.Status, incident.Summary, incident.Description, incident.FromAt, incident.ToAt,
		incident.IsConfirmed, incident.ConfirmationTime, incident.Quarter, incident.Departament, incident.ClientAffect,
		incident.IsManageable, incident.SaleChannels, incident.TroubleServices, incident.FinLosses, incident.FailureType,
		incident.IsDeploy, incident.DeployLink, incident.Labels, incident.IsDowntime, incident.PostmortemLink,
		incident.Creator, incident.RuleID, incident.MatchingCount, incident.LastMatchingTime, incident.AlertsData,
		incident.CreatedAt, incident.UpdatedAt,
	); err != nil {
		return fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"id": incident.ID,
	}).Debug("The incident has been created in the database")

	return nil
}

// GetIncident retrieves an incident from the database based on the incident ID.
// This method performs a database query to select the incident and maps the result to an Incident model.
func (ir *IncidentsRepository) GetIncident(ctx context.Context, id string) (*models.Incident, error) {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("Retrieving an incident from the database")

	incident := &models.Incident{}
	if err := ir.dbPool.QueryRow(ctx, selectIncidentQuery, id).Scan(
		&incident.ID, &incident.Type, &incident.Status, &incident.Summary, &incident.Description, &incident.FromAt, &incident.ToAt,
		&incident.IsConfirmed, &incident.ConfirmationTime, &incident.Quarter, &incident.Departament, &incident.ClientAffect,
		&incident.IsManageable, &incident.SaleChannels, &incident.TroubleServices, &incident.FinLosses, &incident.FailureType,
		&incident.IsDeploy, &incident.DeployLink, &incident.Labels, &incident.IsDowntime, &incident.PostmortemLink,
		&incident.Creator, &incident.RuleID, &incident.MatchingCount, &incident.LastMatchingTime,
		&incident.AlertsData, &incident.CreatedAt, &incident.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Debug("Incident successfully retrieved from the database")

	return incident, nil
}

// UpdateIncident updates an existing incident in the database.
// This method logs the update process and executes an SQL query to update the incident data based on provided changes.
func (ir *IncidentsRepository) UpdateIncident(ctx context.Context, incident *models.Incident) error {
	log.WithFields(log.Fields{
		"id": incident.ID,
	}).Debug("Updating an incident in the database")

	if _, err := ir.dbPool.Exec(
		ctx,
		updateIncidentQuery,
		incident.Status, incident.Summary, incident.Description, incident.FromAt, incident.ToAt, incident.IsConfirmed,
		incident.ConfirmationTime, incident.Departament, incident.ClientAffect, incident.IsManageable,
		incident.SaleChannels, incident.TroubleServices, incident.FinLosses, incident.FailureType, incident.IsDeploy,
		incident.DeployLink, incident.Labels, incident.IsDowntime, incident.PostmortemLink,
		incident.MatchingCount, incident.LastMatchingTime,
		incident.UpdatedAt, incident.ID,
	); err != nil {
		return fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"id": incident.ID,
	}).Debug("The incident has been updated in the database")

	return nil
}

// DeleteIncident removes an incident from the database by its ID.
// This method logs the deletion process and executes an SQL query to delete the incident.
func (ir *IncidentsRepository) DeleteIncident(ctx context.Context, id string) error {
	log.WithFields(log.Fields{
		"id": id,
	}).Debug("Deleting an incident from the database")

	if _, err := ir.dbPool.Exec(ctx, deleteIncidentQuery, id); err != nil {
		return fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"id": id,
	}).Debug("The incident has been deleted from the database")

	return nil
}

// GetIncidents retrieves a list of incidents from the database based on provided filters, sorting, and pagination settings.
// This method constructs a dynamic SQL query based on input parameters and fetches the results accordingly.
func (ir *IncidentsRepository) GetIncidents(ctx context.Context, filterBy map[string]interface{}, sortBy string, sortOrder string, pageNum int, pageSize int, startTime time.Time, endTime time.Time) ([]*models.Incident, error) {
	log.WithFields(log.Fields{
		"filterBy":  filterBy,
		"sortBy":    sortBy,
		"sortOrder": sortOrder,
		"pageNum":   pageNum,
		"pageSize":  pageSize,
		"startTime": startTime,
		"endTime":   endTime,
	}).Debug("Retrieving incidents with filters, sorting, and pagination from the database")

	// Build the SQL query dynamically based on filters and pagination settings.
	query, args := buildGetQueryForIncidents(filterBy, sortBy, sortOrder, pageNum, pageSize, startTime, endTime)

	// Execute the query and collect results.
	rows, err := ir.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing the query: %w", err)
	}
	defer rows.Close()

	// Iterate through the result set and populate the incidents slice.
	incidents := []*models.Incident{}
	for rows.Next() {
		incident := &models.Incident{}
		if err := rows.Scan(
			&incident.ID, &incident.Type, &incident.Status, &incident.Summary, &incident.Description, &incident.FromAt, &incident.ToAt,
			&incident.IsConfirmed, &incident.ConfirmationTime, &incident.Quarter, &incident.Departament, &incident.ClientAffect,
			&incident.IsManageable, &incident.SaleChannels, &incident.TroubleServices, &incident.FinLosses, &incident.FailureType,
			&incident.IsDeploy, &incident.DeployLink, &incident.Labels, &incident.IsDowntime, &incident.PostmortemLink,
			&incident.Creator, &incident.RuleID, &incident.MatchingCount, &incident.LastMatchingTime,
			&incident.AlertsData, &incident.CreatedAt, &incident.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning the row: %w", err)
		}
		incidents = append(incidents, incident)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	log.WithFields(log.Fields{
		"incidentsCount": len(incidents),
	}).Debug("Incidents successfully retrieved from the database")

	return incidents, nil
}

// GetTotalIncidents counts the total number of incidents in the database that match specified filters and date range.
// This method constructs a count query dynamically and executes it to obtain the total count of matching incidents.
func (ir *IncidentsRepository) GetTotalIncidents(ctx context.Context, filterBy map[string]interface{}, startTime, endTime time.Time) (int, error) {
	log.WithFields(log.Fields{
		"filterBy":  filterBy,
		"startTime": startTime,
		"endTime":   endTime,
	}).Debug("Counting the total number of incidents with filters")

	// Build the count query based on the filter and date range.
	query, args := buildCountQueryForIncidents(filterBy, startTime, endTime)

	// Execute the count query.
	totalIncidents := 0
	if err := ir.dbPool.QueryRow(ctx, query, args...).Scan(&totalIncidents); err != nil {
		return 0, fmt.Errorf("error executing the query: %w", err)
	}

	log.WithFields(log.Fields{
		"totalIncidentsCount": totalIncidents,
	}).Debug("Total number of incidents with filters counted successfully")

	return totalIncidents, nil
}

// buildGetQueryForIncidents constructs a dynamic SQL query for retrieving incidents based on various filters and pagination settings.
// This internal function assembles the SQL query string and corresponding arguments based on the specified criteria.
func buildGetQueryForIncidents(filterBy map[string]interface{}, sortBy, sortOrder string, pageNum, pageSize int, startTime, endTime time.Time) (string, []interface{}) {
	baseQuery := `SELECT id, type, status, summary, description, from_at, to_at, is_confirmed, confirmation_time,
        quarter, departament, client_affect, is_manageable, sale_channels, trouble_services, fin_losses, failure_type,
        is_deploy, deploy_link, labels, is_downtime, postmortem_link, creator, rule_id, matching_count, last_matching_time, alerts_data,
        created_at, updated_at FROM a2i_incidents WHERE 1 = 1`
	args := []interface{}{}
	argId := 1

	// Append conditions for each filter field.
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

	// Add time range filter if specified.
	if !startTime.IsZero() && !endTime.IsZero() {
		baseQuery += fmt.Sprintf(" AND created_at BETWEEN $%d AND $%d", argId, argId+1)
		args = append(args, startTime, endTime)
		argId += 2
	}

	// Append sorting and pagination parameters.
	if sortBy != "" && sortOrder != "" {
		baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	}

	offset := (pageNum - 1) * pageSize
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argId, argId+1)
	args = append(args, pageSize, offset)

	return baseQuery, args
}

// buildCountQueryForIncidents constructs a dynamic SQL query to count incidents based on filters and a date range.
// This internal function creates a SQL string and a list of arguments for the count query.
func buildCountQueryForIncidents(filterBy map[string]interface{}, startTime, endTime time.Time) (string, []interface{}) {
	baseQuery := "SELECT COUNT(id) FROM a2i_incidents WHERE 1 = 1"
	args := []interface{}{}
	argId := 1

	// Add conditions based on the filters provided.
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

	// Filter by date range if specified.
	if !startTime.IsZero() && !endTime.IsZero() {
		baseQuery += fmt.Sprintf(" AND created_at BETWEEN $%d AND $%d", argId, argId+1)
		args = append(args, startTime, endTime)
		argId += 2
	}

	return baseQuery, args
}
