package cache

import (
    "context"
    "os"
    "testing"
    "time"
)

// TestConnect prueba la conexión a Redis
func TestConnect(t *testing.T) {
    if testing.Short() {
        t.Skip("Saltando test de integración en modo -short")
    }

    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        redisURL = "localhost:6379"
    }

    client := Connect(redisURL)

    if client == nil {
        t.Fatal("El cliente de Redis no debería ser nil")
    }

    // Verificar que la conexión funciona
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx); err != nil {
        t.Fatalf("Error verificando conexión a Redis: %v", err)
    }
}

// TestSetAndGet prueba guardar y obtener un valor
func TestSetAndGet(t *testing.T) {
    if testing.Short() {
        t.Skip("Saltando test de integración en modo -short")
    }

    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        redisURL = "localhost:6379"
    }

    client := Connect(redisURL)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Clave y valor únicos para este test
    key := "test:setget:" + time.Now().Format("20060102150405.000000000")
    // ↑ Timestamp con nanosegundos evita colisiones
    value := "test-value-123"

    // Cleanup: eliminar la clave después del test
    defer func() {
        cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cleanupCancel()
        _ = client.Delete(cleanupCtx, key)
    }()

    // 1. Guardar valor
    err := client.Set(ctx, key, value, 1*time.Minute)
    if err != nil {
        t.Fatalf("Error guardando en Redis: %v", err)
    }

    // 2. Obtener valor
    result, err := client.Get(ctx, key)
    if err != nil {
        t.Fatalf("Error obteniendo de Redis: %v", err)
    }

    // 3. Verificar
    if result != value {
        t.Errorf("Valor esperado: %s, obtenido: %s", value, result)
    }
}

// TestDelete prueba eliminar un valor
func TestDelete(t *testing.T) {
    if testing.Short() {
        t.Skip("Saltando test de integración en modo -short")
    }

    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        redisURL = "localhost:6379"
    }

    client := Connect(redisURL)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    key := "test:delete:" + time.Now().Format("20060102150405.000000000")
    value := "to-be-deleted"

    // Guardar primero
    if err := client.Set(ctx, key, value, 1*time.Minute); err != nil {
        t.Fatalf("Error guardando: %v", err)
    }

    // Eliminar
    if err := client.Delete(ctx, key); err != nil {
        t.Fatalf("Error eliminando: %v", err)
    }

    // Verificar que ya no existe
    _, err := client.Get(ctx, key)
    if err == nil {
        t.Error("El valor debería haber sido eliminado")
    }
    // ↑ Si no hay error, significa que el valor sigue existiendo
    // ↑ Eso sería un fallo del Delete
}

// TestExpiration prueba que las claves expiran
func TestExpiration(t *testing.T) {
    if testing.Short() {
        t.Skip("Saltando test de integración en modo -short")
    }

    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        redisURL = "localhost:6379"
    }

    client := Connect(redisURL)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    key := "test:expire:" + time.Now().Format("20060102150405.000000000")
    value := "will-expire"

    // Guardar con expiración de 1 segundo
    if err := client.Set(ctx, key, value, 1*time.Second); err != nil {
        t.Fatalf("Error guardando: %v", err)
    }

    // Verificar que existe inmediatamente
    result, err := client.Get(ctx, key)
    if err != nil {
        t.Fatalf("Error obteniendo valor recién guardado: %v", err)
    }
    if result != value {
        t.Errorf("Valor esperado: %s, obtenido: %s", value, result)
    }

    // Esperar a que expire
    time.Sleep(2 * time.Second)

    // Verificar que ya no existe
    _, err = client.Get(ctx, key)
    if err == nil {
        t.Error("El valor debería haber expirado después de 1 segundo")
    }
}

// TestSetAndGet_TableDriven prueba varios escenarios
func TestSetAndGet_TableDriven(t *testing.T) {
    if testing.Short() {
        t.Skip("Saltando test de integración en modo -short")
    }

    redisURL := os.Getenv("REDIS_URL")
    if redisURL == "" {
        redisURL = "localhost:6379"
    }

    client := Connect(redisURL)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Tabla de casos
    tests := []struct {
        name       string
        key        string
        value      string
        expiration time.Duration
    }{
        {"valor simple", "test:table:1", "hello", 1 * time.Minute},
        {"valor vacío", "test:table:2", "", 1 * time.Minute},
        {"valor con JSON", "test:table:3", `{"id":"123","name":"test"}`, 1 * time.Minute},
        {"valor largo", "test:table:4", string(make([]byte, 1000)), 1 * time.Minute},
        {"expiración corta", "test:table:5", "expira", 10 * time.Second},
    }

    for _, tt := range tests {
        // ↑ tt = test actual
        t.Run(tt.name, func(t *testing.T) {
            // Añadir prefijo único
            key := tt.key + ":" + time.Now().Format("20060102150405.000000000")
            
            // Cleanup
            defer func() {
                cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
                defer cleanupCancel()
                _ = client.Delete(cleanupCtx, key)
            }()

            // Guardar
            if err := client.Set(ctx, key, tt.value, tt.expiration); err != nil {
                t.Fatalf("Error guardando: %v", err)
            }

            // Obtener
            result, err := client.Get(ctx, key)
            if err != nil {
                t.Fatalf("Error obteniendo: %v", err)
            }

            // Verificar
            if result != tt.value {
                t.Errorf("Valor esperado: '%s', obtenido: '%s'", tt.value, result)
            }
        })
    }
}