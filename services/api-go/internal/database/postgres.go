package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a PostgreSQL connection.
func Connect(databaseURL string) *pgxpool.Pool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create connection pool
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Error connecting to PostgreSQL: %v", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Error verifying PostgreSQL connection: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return pool
}

// CreateTables creates the required tables.
func CreateTables(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// SQL to create the metrics table
	metricsTable := `
    CREATE TABLE IF NOT EXISTS metrics (
        id UUID PRIMARY KEY,
        cpu_usage FLOAT NOT NULL DEFAULT 0,
        memory_usage FLOAT NOT NULL DEFAULT 0,
        disk_usage FLOAT NOT NULL DEFAULT 0,
        network_in FLOAT NOT NULL DEFAULT 0,
        network_out FLOAT NOT NULL DEFAULT 0,
        timestamp BIGINT NOT NULL DEFAULT 0,
        created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );`

	// Create the metrics table
	if _, err := pool.Exec(ctx, metricsTable); err != nil {
		return err
	}

	// SQL to create the alerts table
	alertsTable := `
    CREATE TABLE IF NOT EXISTS alerts (
        id UUID PRIMARY KEY,
        severity VARCHAR(20) NOT NULL DEFAULT 'info',
        message TEXT NOT NULL,
        status VARCHAR(20) NOT NULL DEFAULT 'active',
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP
    );`

	// Create the alerts table
	if _, err := pool.Exec(ctx, alertsTable); err != nil {
		return err
	}

	log.Println("✅ Tables created successfully")
	return nil
}
