package main // dnywonnt.me/alerts2incidents/cmd/handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"dnywonnt.me/alerts2incidents/internal/cache"
	"dnywonnt.me/alerts2incidents/internal/config"
	"dnywonnt.me/alerts2incidents/internal/database"
	"dnywonnt.me/alerts2incidents/internal/database/repositories"
	"dnywonnt.me/alerts2incidents/internal/models"
	"dnywonnt.me/alerts2incidents/internal/service"
	"dnywonnt.me/alerts2incidents/internal/service/impl"
	"dnywonnt.me/alerts2incidents/internal/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	log "github.com/sirupsen/logrus"
)

// Handler struct to hold necessary components for handling incidents
type Handler struct {
	dbPool         *pgxpool.Pool
	rulesCache     *cache.Cache
	incidentsCache *cache.Cache
	rulesRepo      *repositories.RulesRepository
	incidentsRepo  *repositories.IncidentsRepository
	dataCh         chan map[service.CollectorType][]byte
	alertsCh       chan []models.Alert
	collectors     []service.Collector
	alertsParser   *service.AlertsParser
}

// InitializeHandler initializes the Handler with necessary configurations and returns it
func InitializeHandler() *Handler {
	log.Info("Initializing the incidents handler")

	// Load database configuration
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to load database config")
	}

	// Load service configuration
	serviceConfig, err := config.LoadServiceConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to load service config")
	}

	// Create a connection string for the database
	encodedPassword := url.QueryEscape(dbConfig.Password)
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?pool_max_conns=%d",
		dbConfig.User, encodedPassword, dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.MaxConnections)
	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to create database pool")
	}

	// Migrate the database
	if err := database.Migrate(dbPool, "./migrations", database.MigrateUp); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to migrate database")
	}

	// Initialize and return the handler
	return &Handler{
		dbPool:         dbPool,
		rulesRepo:      repositories.NewRulesRepository(dbPool),
		incidentsRepo:  repositories.NewIncidentsRepository(dbPool),
		rulesCache:     cache.NewCache(serviceConfig.RulesCacheMaxSize, "rules"),
		incidentsCache: cache.NewCache(serviceConfig.IncidentsCacheMaxSize, "incidents"),
		dataCh:         make(chan map[service.CollectorType][]byte, serviceConfig.DataChanMaxSize),
		alertsCh:       make(chan []models.Alert, serviceConfig.AlertsChanMaxSize),
		collectors: []service.Collector{
			impl.NewGrafanaCollector(serviceConfig.GrafanaCollector),
			impl.NewZabbixCollector(serviceConfig.ZabbixCollector),
		},
		alertsParser: service.NewAlertsParser(serviceConfig.AlertsParser),
	}
}

// Run starts the handler, initializes caches, and listens for signals to stop
func (h *Handler) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Starting the incidents handler")

	// Initialize caches
	h.initializeCaches(ctx)

	wg := &sync.WaitGroup{}

	// Update rules cache periodically
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.updateRulesCache(ctx)
	}()

	// Update incidents cache periodically
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.updateIncidentsCache(ctx)
	}()

	// Start data collectors
	for _, col := range h.collectors {
		wg.Add(1)
		go func(c service.Collector) {
			defer wg.Done()
			c.CollectData(ctx, h.dataCh)
		}(col)
	}

	// Start alerts parser
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.alertsParser.ParseAndAggregateAlerts(ctx, h.dataCh, h.alertsCh)
	}()

	// Process alerts
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case alerts := <-h.alertsCh:
				h.processAlerts(ctx, alerts)
			}
		}
	}()

	log.Info("The incidents handler successfully started; waiting for alerts")

	// Listen for termination signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
	log.Info("Stopping the incidents handler")

	cancel()
	wg.Wait()

	close(h.dataCh)
	close(h.alertsCh)
	h.dbPool.Close()
	h.rulesCache.Clear()
	h.incidentsCache.Clear()

	log.Info("The incidents handler has been stopped")
}

// initializeCaches populates the caches with initial data from the database
func (h *Handler) initializeCaches(ctx context.Context) {
	pageSize := 100
	zeroTime := time.Time{}

	// Initialize rules cache
	totalRules, err := h.rulesRepo.GetTotalRules(ctx, nil, zeroTime, zeroTime)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to get total number of rules")
	} else if totalRules > 0 {
		pages := utils.CalculatePages(totalRules, pageSize)
		for i := 1; i <= pages; i++ {
			rules, err := h.rulesRepo.GetRules(ctx, nil, "created_at", "desc", i, pageSize, zeroTime, zeroTime)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("Failed to get rules to initialize cache")
				break
			} else if len(rules) > 0 {
				for _, rule := range rules {
					h.rulesCache.SetItem(rule.ID, rule)
				}
			}
		}
	}

	// Initialize incidents cache
	cacheMaxSize := h.incidentsCache.GetMaxSize()
	incidents, err := h.incidentsRepo.GetIncidents(ctx, nil, "created_at", "desc", 1, cacheMaxSize, zeroTime, zeroTime)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to get incidents to initialize cache")
	} else if len(incidents) > 0 {
		for _, incident := range incidents {
			h.incidentsCache.SetItem(incident.ID, incident)
		}
	}
}

// processAlerts processes incoming alerts and creates or updates incidents based on matching rules
func (h *Handler) processAlerts(ctx context.Context, alerts []models.Alert) {
	pageSize := 100

	totalRules := h.rulesCache.GetTotalItems()
	rulesTotalPages := utils.CalculatePages(totalRules, pageSize)

	for i := 1; i <= rulesTotalPages; i++ {
		cachedRules := h.rulesCache.GetItems(i, pageSize)

		for _, item := range cachedRules {
			rule, ok := item.Value.(*models.Rule)
			if !ok || (ok && rule.IsMuted) {
				continue
			}

			// Find matching alerts for the rule
			matchingAlerts, err := service.FindMatchingAlerts(alerts, rule)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("Failed to find matching alerts")
				continue
			}

			if matchingAlerts != nil {
				currentTimeUTC := time.Now().UTC()
				incident, exists := h.getIncidentFromCacheForRule(rule.ID)

				// Update existing incident if it matches the rule and is within its lifetime
				if exists && time.Since(incident.CreatedAt) <= rule.IncidentLifeTime {
					log.WithFields(log.Fields{
						"id":               incident.ID,
						"ruleID":           rule.ID,
						"matchingCount":    incident.MatchingCount,
						"incidentLifeTime": rule.IncidentLifeTime.String(),
					}).Info("The incident already exists for the rule; updating info")

					incident.MatchingCount += 1
					incident.LastMatchingTime = currentTimeUTC
					incident.UpdatedAt = currentTimeUTC

					if err := h.incidentsRepo.UpdateIncident(ctx, incident); err != nil {
						log.WithFields(log.Fields{
							"error": err.Error(),
						}).Error("Failed to update incident in the database")
					}
					continue
				}

				// Create a new incident if no existing incident is found
				alertsData, err := json.Marshal(matchingAlerts)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("Failed to marshal alerts for a new incident")
					continue
				}

				zeroTime := time.Time{}
				newIncident := &models.Incident{
					ID:               uuid.NewString(),
					Type:             "auto",
					Status:           "actual",
					Summary:          rule.SetIncidentSummary,
					Description:      rule.SetIncidentDescription,
					FromAt:           currentTimeUTC,
					ToAt:             zeroTime,
					IsConfirmed:      false,
					ConfirmationTime: zeroTime,
					Quarter:          utils.GetCurrentQuarter(),
					Departament:      rule.SetIncidentDepartament,
					ClientAffect:     rule.SetIncidentClientAffect,
					IsManageable:     rule.SetIncidentIsManageable,
					SaleChannels:     rule.SetIncidentSaleChannels,
					TroubleServices:  rule.SetIncidentTroubleServices,
					FinLosses:        0,
					FailureType:      rule.SetIncidentFailureType,
					DeployLink:       "",
					Labels:           rule.SetIncidentLabels,
					IsDowntime:       rule.SetIncidentIsDowntime,
					PostmortemLink:   "",
					Creator:          "handler",
					RuleID:           &rule.ID,
					MatchingCount:    1,
					LastMatchingTime: currentTimeUTC,
					AlertsData:       string(alertsData),
					CreatedAt:        currentTimeUTC,
					UpdatedAt:        currentTimeUTC,
				}

				if err := newIncident.Validate(); err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("Failed to validate the incident model")
					continue
				}

				if err := h.incidentsRepo.CreateIncident(ctx, newIncident); err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
						"id":    newIncident.ID,
					}).Error("Failed to create a new incident in the database")
					continue
				}

				log.WithFields(log.Fields{
					"id":     newIncident.ID,
					"ruleID": rule.ID,
				}).Info("A new incident has been detected")
			}
		}
	}
}

// updateIncidentsCache listens for notifications and updates the incidents cache accordingly
func (h *Handler) updateIncidentsCache(ctx context.Context) {
	database.ListenToNotifications(ctx, h.dbPool, database.IncidentsChannel, func(notification *pgconn.Notification) {
		if err := updateCacheFromNotification(ctx, h.incidentsCache, notification, func(ctx context.Context, id string) (interface{}, error) {
			return h.incidentsRepo.GetIncident(ctx, id)
		}); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to update incidents cache from notification")
		}
	})
}

// updateRulesCache listens for notifications and updates the rules cache accordingly
func (h *Handler) updateRulesCache(ctx context.Context) {
	database.ListenToNotifications(ctx, h.dbPool, database.RulesChannel, func(notification *pgconn.Notification) {
		if err := updateCacheFromNotification(ctx, h.rulesCache, notification, func(ctx context.Context, id string) (interface{}, error) {
			return h.rulesRepo.GetRule(ctx, id)
		}); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to update rules cache from notification")
		}
	})
}

// updateCacheFromNotification updates the cache based on the notification received
func updateCacheFromNotification(ctx context.Context, cache *cache.Cache, notification *pgconn.Notification, fetchItem func(context.Context, string) (interface{}, error)) error {
	parts := strings.SplitN(notification.Payload, ":", 2)
	if len(parts) < 2 {
		return fmt.Errorf("invalid payload in notification: %s", notification.Payload)
	}

	action, id := parts[0], parts[1]

	switch action {
	case "INSERT", "UPDATE":
		item, err := fetchItem(ctx, id)
		if err != nil {
			return err
		}
		cache.SetItem(id, item)

	case "DELETE":
		cache.DeleteItem(id)

	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}

// getIncidentFromCacheForRule retrieves an incident from the cache based on the rule ID
func (h *Handler) getIncidentFromCacheForRule(ruleID string) (*models.Incident, bool) {
	items := h.incidentsCache.GetAllItems()

	for _, item := range items {
		if incident, ok := item.Value.(*models.Incident); ok && (incident.RuleID != nil && *incident.RuleID == ruleID) {
			return incident, true
		}
	}

	return nil, false
}
