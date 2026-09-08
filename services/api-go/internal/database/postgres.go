package database

import (
    "context"
    "log"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

// Connect crea una conexión a PostgreSQL
func Connect(databaseURL string) *pgxpool.Pool {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Crear pool de conexiones
    pool, err := pgxpool.New(ctx, databaseURL)
    if err != nil {
        log.Fatalf("Error conectando a PostgreSQL: %v", err)
    }

    // Verificar conexión
    if err := pool.Ping(ctx); err != nil {
        log.Fatalf("Error verificando conexión a PostgreSQL: %v", err)
    }

    log.Println("✅ Conectado a PostgreSQL")
    return pool
}

// CreateTables crea las tablas necesarias
func CreateTables(pool *pgxpool.Pool) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // SQL para crear tabla metrics
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

    // Ejecutar creación de tabla metrics
    if _, err := pool.Exec(ctx, metricsTable); err != nil {
        return err
    }

    // SQL para crear tabla alerts
    alertsTable := `
    CREATE TABLE IF NOT EXISTS alerts (
        id UUID PRIMARY KEY,
        severity VARCHAR(20) NOT NULL DEFAULT 'info',
        message TEXT NOT NULL,
        status VARCHAR(20) NOT NULL DEFAULT 'active',
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP
    );`

    // Ejecutar creación de tabla alerts
    if _, err := pool.Exec(ctx, alertsTable); err != nil {
        return err
    }

    log.Println("✅ Tablas creadas correctamente")
    return nil
}