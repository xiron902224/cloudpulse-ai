// Package testutils proporciona utilidades para tests.
package testutils

import (
    "context"
    "fmt"
    "log"
    "os"
    "testing"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

// ============================================
// Configuración
// ============================================

// TestDatabaseURL devuelve la URL de la base de datos para tests.
//
// Prioriza la variable de entorno TEST_DATABASE_URL (útil en CI),
// con fallback a una DB local de desarrollo.
func TestDatabaseURL() string {
    if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
        return url
    }
    return "postgres://cloudpulse:cloudpulse_dev@localhost:5432/cloudpulse_test?sslmode=disable"
}

// ============================================
// Setup / Teardown
// ============================================

// SetupTestDB crea una conexión de test, verifica que funciona y limpia las tablas.
//
// Si la DB no está disponible, salta el test (útil cuando no quieres correr tests de integración).
// Registra un cleanup automático para cerrar el pool al finalizar el test.
//
// Uso típico:
//
//	func TestSomething(t *testing.T) {
//	    pool := testutils.SetupTestDB(t)
//	    // ... usar pool ...
//	}
func SetupTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Crear pool de conexiones
    pool, err := pgxpool.New(ctx, TestDatabaseURL())
    if err != nil {
        t.Skipf("No se pudo crear el pool de test (¿está PostgreSQL corriendo?): %v", err)
    }

    // Verificar conexión
    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        t.Skipf("No se pudo hacer ping a la DB de test: %v", err)
    }

    // Crear tablas si no existen
    if err := createTestTables(ctx, pool); err != nil {
        pool.Close()
        t.Fatalf("Error creando tablas de test: %v", err)
    }

    // Limpiar datos previos
    if err := CleanupTables(ctx, pool); err != nil {
        pool.Close()
        t.Fatalf("Error limpiando tablas: %v", err)
    }

    // Cleanup automático al terminar el test
    t.Cleanup(func() {
        pool.Close()
    })

    return pool
}

// createTestTables crea las tablas necesarias para tests.
//
// Usa IF NOT EXISTS para ser idempotente (los tests se pueden correr múltiples veces).
func createTestTables(ctx context.Context, pool *pgxpool.Pool) error {
    queries := []string{
        `
        CREATE TABLE IF NOT EXISTS metrics (
            id UUID PRIMARY KEY,
            cpu_usage FLOAT NOT NULL DEFAULT 0,
            memory_usage FLOAT NOT NULL DEFAULT 0,
            disk_usage FLOAT NOT NULL DEFAULT 0,
            network_in FLOAT NOT NULL DEFAULT 0,
            network_out FLOAT NOT NULL DEFAULT 0,
            timestamp BIGINT NOT NULL DEFAULT 0,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        )`,
        `
        CREATE TABLE IF NOT EXISTS alerts (
            id UUID PRIMARY KEY,
            severity VARCHAR(20) NOT NULL DEFAULT 'info',
            message TEXT NOT NULL,
            status VARCHAR(20) NOT NULL DEFAULT 'active',
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP
        )`,
    }

    for i, q := range queries {
        if _, err := pool.Exec(ctx, q); err != nil {
            return fmt.Errorf("query %d: %w", i, err)
        }
    }

    return nil
}

// CleanupTables limpia todas las tablas de test.
//
// Usa TRUNCATE (más rápido que DELETE) y CASCADE para limpiar relaciones.
func CleanupTables(ctx context.Context, pool *pgxpool.Pool) error {
    tables := []string{"metrics", "alerts"}
    for _, table := range tables {
        if _, err := pool.Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE"); err != nil {
            return fmt.Errorf("error limpiando tabla %s: %w", table, err)
        }
    }
    return nil
}

// ============================================
// Helpers adicionales
// ============================================

// MustSetupTestDB es como SetupTestDB pero falla (no salta) si la DB no está disponible.
//
// Útil cuando QUIERES que los tests fallen si la DB no está corriendo (por ejemplo en CI).
func MustSetupTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    pool, err := pgxpool.New(ctx, TestDatabaseURL())
    if err != nil {
        t.Fatalf("No se pudo crear el pool de test: %v", err)
    }

    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        t.Fatalf("No se pudo hacer ping a la DB de test: %v", err)
    }

    if err := createTestTables(ctx, pool); err != nil {
        pool.Close()
        t.Fatalf("Error creando tablas: %v", err)
    }

    if err := CleanupTables(ctx, pool); err != nil {
        pool.Close()
        t.Fatalf("Error limpiando tablas: %v", err)
    }

    t.Cleanup(func() {
        pool.Close()
    })

    return pool
}

// RunTestMain ejecuta el flujo estándar de TestMain.
//
// Uso:
//
//	func TestMain(m *testing.M) {
//	    testutils.RunTestMain(m)
//	}
func RunTestMain(m *testing.M) {
    log.Println("🧪 Iniciando tests...")

    code := m.Run()

    log.Println("🧪 Tests finalizados")
    os.Exit(code)
}