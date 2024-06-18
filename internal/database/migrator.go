package database // dnywonnt.me/alerts2incidents/internal/database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	log "github.com/sirupsen/logrus"
)

// MigrateDirection is an enumeration type defining the direction of migration.
type MigrateDirection string

// Define constants to represent the direction of the migration.
const (
	MigrateUp   MigrateDirection = "up"   // MigrateUp constant for upgrading the database schema.
	MigrateDown MigrateDirection = "down" // MigrateDown constant for downgrading the database schema.
)

// Migrate function handles the database migrations, either upgrading or downgrading, based on the provided direction.
func Migrate(pool *pgxpool.Pool, migrationsPath string, direction MigrateDirection) error {
	log.WithFields(log.Fields{
		"migrationsPath": migrationsPath,
		"direction":      direction,
	}).Debug("Starting the database migration process")

	// Retrieve the database connection configuration from the pool.
	connConfig := pool.Config().ConnConfig
	// Open a new database connection using the connection configuration.
	db := stdlib.OpenDB(*connConfig)
	defer db.Close() // Ensure the database connection is closed after the migration.

	// Initialize a new Postgres driver instance for the migration.
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("error initializing postgres instance for migration: %w", err)
	}

	// Create a new migration instance with the specified migrations path and database driver.
	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationsPath), "postgres", driver)
	if err != nil {
		return fmt.Errorf("error creating migration instance: %w", err)
	}

	// Execute the migration in the specified direction (up or down).
	switch direction {
	case MigrateUp:
		log.WithFields(log.Fields{}).Debug("Applying migration in the 'up' direction")
		// Apply all up migrations.
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error applying migration up: %w", err)
		}
	case MigrateDown:
		log.WithFields(log.Fields{}).Debug("Applying migration in the 'down' direction")
		// Revert the last migration step.
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error applying migration down: %w", err)
		}
	default:
		// Return an error if the migration direction is neither 'up' nor 'down'.
		return fmt.Errorf("invalid migration direction specified: %s", direction)
	}

	log.WithFields(log.Fields{
		"direction": direction,
	}).Debug("Database migration successfully completed")

	return nil
}
