package handlers // dnywonnt.me/alerts2incidents/internal/api/v1/handlers

import (
	"net/http"
	"strconv"
	"time"

	v1 "dnywonnt.me/alerts2incidents/internal/api/v1"
	"dnywonnt.me/alerts2incidents/internal/api/v1/dtos"
	"dnywonnt.me/alerts2incidents/internal/config"
	"dnywonnt.me/alerts2incidents/internal/database/repositories"
	"dnywonnt.me/alerts2incidents/internal/utils"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

// RegisterIncidentsRoutes sets up the routing of incident endpoints.
func RegisterIncidentsRoutes(router *gin.Engine, repo *repositories.IncidentsRepository, apiCfg *config.ApiConfig) {
	routerGroup := router.Group("/api/v1/incidents")

	routerGroup.Use(v1.JWTMiddleware(apiCfg))

	// Define route handlers for each operation.
	routerGroup.GET("/:id", getIncident(repo))
	routerGroup.GET("/", getIncidents(repo))
	routerGroup.POST("/", createIncident(repo))
	routerGroup.PUT("/:id", updateIncident(repo))
	routerGroup.DELETE("/:id", deleteIncident(repo))
}

// getIncident returns a handler for retrieving a single incident by ID.
func getIncident(repo *repositories.IncidentsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve incident by ID from repository.
		incident, err := repo.GetIncident(c, c.Param("id"))
		if err != nil {
			log.WithFields(log.Fields{
				"id":    c.Param("id"),
				"error": err.Error(),
			}).Error("Failed to retrieve incident")
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}

		// Respond with the found incident.
		c.JSON(http.StatusOK, incident)
	}
}

// getIncidents returns a handler for retrieving a list of incidents with optional filters.
func getIncidents(repo *repositories.IncidentsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve and validate pagination parameters.
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("pageSize", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			log.WithFields(log.Fields{
				"page":  pageStr,
				"error": err.Error(),
			}).Error("Failed to parse pagination parameter")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			log.WithFields(log.Fields{
				"pageSize": pageSizeStr,
				"error":    err.Error(),
			}).Error("Failed to parse pagination parameter")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Parse and validate date filters.
		startTimeStr := c.Query("startTime")
		startTime := time.Time{}
		if startTimeStr != "" {
			startTime, err = time.Parse(time.RFC3339, startTimeStr)
			if err != nil {
				log.WithFields(log.Fields{
					"startTime": startTimeStr,
					"error":     err.Error(),
				}).Error("Failed to parse start time")
				c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
				return
			}
		}

		endTimeStr := c.Query("endTime")
		endTime := time.Time{}
		if endTimeStr != "" {
			endTime, err = time.Parse(time.RFC3339, endTimeStr)
			if err != nil {
				log.WithFields(log.Fields{
					"endTime": endTimeStr,
					"error":   err.Error(),
				}).Error("Failed to parse end time")
				c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
				return
			}
		}

		// Build filters from query parameters.
		filterBy, err := buildFilterForIncidents(c)
		if err != nil {
			log.WithFields(log.Fields{
				"filterBy": filterBy,
				"error":    err.Error(),
			}).Error("Failed to build filter for incidents")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Retrieve and validate sorting parameters.
		sortBy := c.DefaultQuery("sortBy", "created_at")
		sortOrder := c.DefaultQuery("sortOrder", "desc")

		// Fetch filtered and sorted incidents from the repository.
		incidents, err := repo.GetIncidents(c, filterBy, sortBy, sortOrder, page, pageSize, startTime.UTC(), endTime.UTC())
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to retrieve incidents")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// Calculate total number of pages.
		totalIncidents, err := repo.GetTotalIncidents(c, filterBy, startTime.UTC(), endTime.UTC())
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to get total incidents count")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		totalPages := utils.CalculatePages(totalIncidents, pageSize)

		// Respond with the list of incidents and pagination details.
		c.JSON(http.StatusOK, gin.H{
			"incidents":    incidents,
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
		})
	}
}

// createIncident returns a handler for creating a new incident.
func createIncident(repo *repositories.IncidentsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		dto := &dtos.CreateIncidentDTO{}
		if err := c.ShouldBindJSON(dto); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to unmarshal request JSON data")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Map DTO to the incident model.
		incident, err := v1.MapCreateIncidentDTOToModel(dto)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to map DTO to incident model")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Create a new incident in the repository.
		if err := repo.CreateIncident(c, incident); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to create incident")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// Respond with the newly created incident.
		c.JSON(http.StatusOK, incident)
	}
}

// updateIncident returns a handler for updating an existing incident.
func updateIncident(repo *repositories.IncidentsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		dto := &dtos.UpdateIncidentDTO{}
		if err := c.ShouldBindJSON(dto); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to unmarshal request JSON data")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Retrieve the existing incident to be updated.
		incident, err := repo.GetIncident(c, c.Param("id"))
		if err != nil {
			log.WithFields(log.Fields{
				"id":    c.Param("id"),
				"error": err.Error(),
			}).Error("Failed to retrieve incident")
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}

		// Map the updated fields from DTO to the incident model.
		if err := v1.MapUpdateIncidentDTOToModel(dto, incident); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to map DTO to incident model")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Save the updated incident in the repository.
		if err := repo.UpdateIncident(c, incident); err != nil {
			log.WithFields(log.Fields{
				"id":    c.Param("id"),
				"error": err.Error(),
			}).Error("Failed to update incident")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// Respond with the updated incident.
		c.JSON(http.StatusOK, incident)
	}
}

// deleteIncident returns a handler for deleting an incident.
func deleteIncident(repo *repositories.IncidentsRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attempt to delete the incident by ID.
		if err := repo.DeleteIncident(c, c.Param("id")); err != nil {
			log.WithFields(log.Fields{
				"id":    c.Param("id"),
				"error": err.Error(),
			}).Error("Failed to delete incident")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// Confirm deletion.
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// buildFilterForIncidents constructs a filter map based on the query parameters.
func buildFilterForIncidents(c *gin.Context) (map[string]interface{}, error) {
	filter := make(map[string]interface{})

	// Helper functions to add filters if they are not empty.
	addToFilterIfNotEmpty := func(key, value string) {
		if value != "" {
			filter[key] = value
		}
	}

	addToFilterIfArrayNotEmpty := func(key string, values []string) {
		if len(values) > 0 {
			filter[key] = values
		}
	}

	// Helper function to parse and add values to the filter.
	parseAndAddToFilter := func(key string, parseFunc func(string) (interface{}, error)) error {
		valueStr := c.Query(key)
		if valueStr != "" {
			value, err := parseFunc(valueStr)
			if err != nil {
				return err
			}
			filter[key] = value
		}
		return nil
	}

	// Parsing specific query parameters.
	parseBool := func(str string) (interface{}, error) {
		return strconv.ParseBool(str)
	}

	parseInt := func(str string) (interface{}, error) {
		return strconv.Atoi(str)
	}

	// Add specific filters based on query parameters.
	if err := parseAndAddToFilter("is_confirmed", parseBool); err != nil {
		return nil, err
	}
	if err := parseAndAddToFilter("quarter", parseInt); err != nil {
		return nil, err
	}
	if err := parseAndAddToFilter("is_deploy", parseBool); err != nil {
		return nil, err
	}
	if err := parseAndAddToFilter("is_downtime", parseBool); err != nil {
		return nil, err
	}

	// Add array and simple filters.
	addToFilterIfArrayNotEmpty("sale_channels", c.QueryArray("sale_channels"))
	addToFilterIfArrayNotEmpty("trouble_services", c.QueryArray("trouble_services"))
	addToFilterIfArrayNotEmpty("labels", c.QueryArray("labels"))

	// Add remaining simple filters.
	addToFilterIfNotEmpty("type", c.Query("type"))
	addToFilterIfNotEmpty("creator", c.Query("creator"))
	addToFilterIfNotEmpty("status", c.Query("status"))
	addToFilterIfNotEmpty("departament", c.Query("departament"))
	addToFilterIfNotEmpty("rule_id", c.Query("rule_id"))
	addToFilterIfNotEmpty("failure_type", c.Query("failure_type"))
	addToFilterIfNotEmpty("is_manageable", c.Query("is_manageable"))

	return filter, nil
}
