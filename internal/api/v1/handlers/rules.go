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

// RegisterRulesRoutes sets up the routing for rule-related API endpoints.
func RegisterRulesRoutes(router *gin.Engine, repo *repositories.RulesRepository, apiCfg *config.ApiConfig) {
	routerGroup := router.Group("/api/v1/rules")

	routerGroup.Use(v1.JWTMiddleware(apiCfg))

	routerGroup.GET("/:id", getRule(repo))
	routerGroup.GET("/", getRules(repo))
	routerGroup.POST("/", createRule(repo))
	routerGroup.PUT("/:id", updateRule(repo))
	routerGroup.DELETE("/:id", deleteRule(repo))
}

// getRule returns a handler for retrieving a single rule by its ID.
func getRule(repo *repositories.RulesRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		rule, err := repo.GetRule(c, c.Param("id"))
		if err != nil {
			log.WithFields(log.Fields{
				"id":    c.Param("id"),
				"error": err.Error(),
			}).Error("Failed to retrieve rule")
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, rule)
	}
}

// getRules returns a handler for retrieving a list of rules based on filter and pagination parameters.
func getRules(repo *repositories.RulesRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		filterBy, err := buildFilterForRules(c)
		if err != nil {
			log.WithFields(log.Fields{
				"filterBy": filterBy,
				"error":    err.Error(),
			}).Error("Failed to build filter for rules")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		sortBy := c.DefaultQuery("sortBy", "created_at")
		sortOrder := c.DefaultQuery("sortOrder", "desc")

		rules, err := repo.GetRules(c, filterBy, sortBy, sortOrder, page, pageSize, startTime.UTC(), endTime.UTC())
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to retrieve rules")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		totalRules, err := repo.GetTotalRules(c, filterBy, startTime.UTC(), endTime.UTC())
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to get total rules count")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		totalPages := utils.CalculatePages(totalRules, pageSize)

		c.JSON(http.StatusOK, gin.H{
			"rules":        rules,
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
		})
	}
}

// createRule returns a handler for creating a new rule based on provided data.
func createRule(repo *repositories.RulesRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		dto := &dtos.CreateRuleDTO{}
		if err := c.ShouldBindJSON(dto); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to unmarshal request JSON data")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		rule, err := v1.MapCreateRuleDTOToModel(dto)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to map DTO to rule model")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		if err := repo.CreateRule(c, rule); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to create rule")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, rule)
	}
}

// updateRule returns a handler for updating an existing rule.
func updateRule(repo *repositories.RulesRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		dto := &dtos.UpdateRuleDTO{}
		if err := c.ShouldBindJSON(dto); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to unmarshal request JSON data")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		rule, err := repo.GetRule(c, c.Param("id"))
		if err != nil {
			log.WithFields(log.Fields{
				"id":    c.Param("id"),
				"error": err.Error(),
			}).Error("Failed to retrieve rule")
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}

		if err := v1.MapUpdateRuleDTOToModel(dto, rule); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to map DTO to rule model")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		if err := repo.UpdateRule(c, rule); err != nil {
			log.WithFields(log.Fields{
				"id":    c.Param("id"),
				"error": err.Error(),
			}).Error("Failed to update rule")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, rule)
	}
}

// deleteRule returns a handler for deleting a rule by its ID.
func deleteRule(repo *repositories.RulesRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := repo.DeleteRule(c, c.Param("id")); err != nil {
			log.WithFields(log.Fields{
				"id":    c.Param("id"),
				"error": err.Error(),
			}).Error("Failed to delete rule")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	}
}

// buildFilterForRules constructs a map of filters for querying rules based on request query parameters.
func buildFilterForRules(c *gin.Context) (map[string]interface{}, error) {
	filter := make(map[string]interface{})

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

	parseBool := func(str string) (interface{}, error) {
		return strconv.ParseBool(str)
	}

	if err := parseAndAddToFilter("is_muted", parseBool); err != nil {
		return nil, err
	}
	if err := parseAndAddToFilter("set_incident_is_downtime", parseBool); err != nil {
		return nil, err
	}

	addToFilterIfArrayNotEmpty("set_incident_sale_channels", c.QueryArray("set_incident_sale_channels"))
	addToFilterIfArrayNotEmpty("set_incident_trouble_services", c.QueryArray("set_incident_trouble_services"))
	addToFilterIfArrayNotEmpty("set_incident_labels", c.QueryArray("set_incident_labels"))

	addToFilterIfNotEmpty("set_incident_departament", c.Query("set_incident_departament"))
	addToFilterIfNotEmpty("set_incident_is_manageable", c.Query("set_incident_is_manageable"))
	addToFilterIfNotEmpty("set_incident_failure_type", c.Query("set_incident_failure_type"))

	return filter, nil
}
