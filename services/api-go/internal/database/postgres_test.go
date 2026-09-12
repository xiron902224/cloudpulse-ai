package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestConnect prueba la conexión a PostgreSQL
func TestConnect(t *testing.T) {
	// Saltar si estamos en modo -short
	if testing.Short() {
		t.Skip("Saltando test de integración en modo -short")
	}

	// Obtener URL de la variable de entorno
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://cloudpulse:cloudpulse_dev@localhost:5432/cloudpulse?sslmode=disable"
	}

	// Conectar
	pool := Connect(databaseURL)

	// Verificar que no sea nil
	if pool == nil {
		t.Fatal("El pool de conexiones no debería ser nil")
	}
	// ↑ t.Fatal detiene el test si falla
	// ↑ t.Error continúa

	// Cerrar al final
	defer pool.Close()
	// ↑ Defer: se ejecuta al final del test
	// ↑ Garantiza que la conexión se cierre

	// Verificar que la conexión funciona
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// ↑ Contexto con timeout de 5 segundos
	// ↑ Si algo tarda más, se cancela

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Error verificando conexión: %v", err)
	}
	// ↑ Ping verifica que la conexión está viva
}

// TestCreateTables prueba la creación de tablas
func TestCreateTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo -short")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://cloudpulse:cloudpulse_dev@localhost:5432/cloudpulse?sslmode=disable"
	}

	pool := Connect(databaseURL)
	defer pool.Close()

	// Crear tablas
	if err := CreateTables(pool); err != nil {
		t.Fatalf("Error creando tablas: %v", err)
	}

	// Verificar que las tablas existen
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Consultar si existe la tabla metrics
	var exists bool
	query := `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'metrics'
        )
    `
	err := pool.QueryRow(ctx, query).Scan(&exists)
	if err != nil {
		t.Fatalf("Error consultando tablas: %v", err)
	}

	if !exists {
		t.Error("La tabla 'metrics' no existe después de CreateTables")
	}

	// Verificar tabla alerts
	query = `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'alerts'
        )
    `
	err = pool.QueryRow(ctx, query).Scan(&exists)
	if err != nil {
		t.Fatalf("Error consultando tablas: %v", err)
	}

	if !exists {
		t.Error("La tabla 'alerts' no existe después de CreateTables")
	}
}

// TestInsertAndQueryMetric prueba insertar y consultar una métrica
func TestInsertAndQueryMetric(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo -short")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://cloudpulse:cloudpulse_dev@localhost:5432/cloudpulse?sslmode=disable"
	}

	pool := Connect(databaseURL)
	defer pool.Close()

	// Asegurar que las tablas existen
	if err := CreateTables(pool); err != nil {
		t.Fatalf("Error creando tablas: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insertar métrica de prueba
	testID := uuid.New().String()

	insertQuery := `
        INSERT INTO metrics (id, cpu_usage, memory_usage, disk_usage, network_in, network_out, timestamp, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := pool.Exec(ctx, insertQuery,
		testID,
		45.5,
		62.3,
		71.0,
		250.5,
		120.8,
		time.Now().Unix(),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Error insertando métrica: %v", err)
	}

	// Cleanup: eliminar la métrica después del test
	defer func() {
		// ↑ Defer con función anónima
		// ↑ Así podemos usar el mismo contexto
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()

		_, _ = pool.Exec(cleanupCtx, "DELETE FROM metrics WHERE id = $1", testID)
		// ↑ Ignoramos el error (cleanup no debe fallar el test)
	}()

	// Consultar la métrica
	var cpu, memory, disk, networkIn, networkOut float64
	var timestamp int64

	selectQuery := `
        SELECT cpu_usage, memory_usage, disk_usage, network_in, network_out, timestamp
        FROM metrics
        WHERE id = $1
    `

	err = pool.QueryRow(ctx, selectQuery, testID).Scan(
		&cpu,
		&memory,
		&disk,
		&networkIn,
		&networkOut,
		&timestamp,
	)
	// ↑ Scan pasa punteros para que la función pueda escribir en las variables
	// ↑ Si no usas &, Scan no puede modificar el valor

	if err != nil {
		t.Fatalf("Error consultando métrica: %v", err)
	}

	// Verificar valores
	if cpu != 45.5 {
		t.Errorf("cpu_usage esperado: 45.5, obtenido: %f", cpu)
	}

	if memory != 62.3 {
		t.Errorf("memory_usage esperado: 62.3, obtenido: %f", memory)
	}
}
