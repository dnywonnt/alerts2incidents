package main // dnywonnt.me/alerts2incidents/cmd/server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "dnywonnt.me/alerts2incidents/internal/api/v1"
	"dnywonnt.me/alerts2incidents/internal/api/v1/handlers"
	"dnywonnt.me/alerts2incidents/internal/config"
	"dnywonnt.me/alerts2incidents/internal/database"
	"dnywonnt.me/alerts2incidents/internal/database/repositories"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	log "github.com/sirupsen/logrus"
)

// Server structure holds the server's runtime configuration and state.
type Server struct {
	srv           *http.Server                      // HTTP server
	dbPool        *pgxpool.Pool                     // Connection pool to the PostgreSQL database
	incidentsRepo *repositories.IncidentsRepository // Repository for incidents data
	rulesRepo     *repositories.RulesRepository     // Repository for rules data
}

// InitializeServer initializes a new server with configuration and database connection.
func InitializeServer() *Server {
	log.Info("Initializing the server")

	// Load API configuration settings
	apiCfg, err := config.LoadApiConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to load API config")
	}

	// Load database configuration settings
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to load database config")
	}

	// Form a connection string and create a database connection pool
	encodedPassword := url.QueryEscape(dbConfig.Password)
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?pool_max_conns=%d",
		dbConfig.User, encodedPassword, dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.MaxConnections)
	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to create database pool")
	}

	// Apply database migrations
	if err := database.Migrate(dbPool, "./migrations", database.MigrateUp); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to migrate database")
	}

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(v1.LoggerMiddleware())

	// Initialize repositories
	incidentsRepo := repositories.NewIncidentsRepository(dbPool)
	rulesRepo := repositories.NewRulesRepository(dbPool)

	// Register HTTP routes
	handlers.RegisterAuthRoutes(router, apiCfg)
	handlers.RegisterIncidentsRoutes(router, incidentsRepo, apiCfg)
	handlers.RegisterRulesRoutes(router, rulesRepo, apiCfg)

	// Returning Server instance with initialized components
	return &Server{
		srv: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", apiCfg.Host, apiCfg.Port),
			Handler: router,
		},
		dbPool:        dbPool,
		incidentsRepo: incidentsRepo,
		rulesRepo:     rulesRepo,
	}
}

// Run starts the HTTP server and handles graceful shutdown on system signals.
func (s *Server) Run() {
	log.WithFields(log.Fields{
		"addr": s.srv.Addr,
	}).Info("Starting the server")

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Failed to listen and serve")
		}
	}()

	log.WithFields(log.Fields{
		"addr": s.srv.Addr,
	}).Info("The server successfully started")

	// Setup signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	log.Info("Stopping the server; handling the last requests")

	// Attempt a graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to shut down the server")
	}
	s.dbPool.Close()

	log.Info("The server successfully stopped")
}
