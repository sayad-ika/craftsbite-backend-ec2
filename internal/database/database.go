package database

import (
	"context"
	"fmt"
	"time"

	"craftsbite-backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect establishes a connection to PostgreSQL using GORM
func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Build DSN (Data Source Name)
	dsn := cfg.GetDSN()

	// Open connection with GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// You can add GORM configuration here if needed
		// For example: Logger, NamingStrategy, etc.
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying *sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)       // Maximum number of open connections
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)       // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime) // Maximum lifetime of a connection

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// HealthCheck performs a health check on the database connection
func HealthCheck(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database: %w", err)
	}

	// Ping the database with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Close gracefully closes the database connection
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
