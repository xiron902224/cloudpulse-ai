package cache

import (
    "context"
    "log"
    "strings"
    "time"

    "github.com/redis/go-redis/v9"
)

// RedisClient es el cliente de Redis
type RedisClient struct {
    Client *redis.Client
}

// Connect crea una conexión a Redis
func Connect(redisURL string) *RedisClient {
    // Quitar el prefijo "redis://" si existe
    addr := strings.TrimPrefix(redisURL, "redis://")
    // ↑ strings.TrimPrefix quita el prefijo si está presente
    // ↑ "redis://localhost:6379" → "localhost:6379"
    // ↑ "localhost:6379" → "localhost:6379" (sin cambios)

    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: "",
        DB:       0,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        log.Printf("⚠️  No se pudo conectar a Redis: %v", err)
        log.Println("   Continuando sin caché...")
        return &RedisClient{Client: client}
    }

    log.Println("✅ Conectado a Redis")
    return &RedisClient{Client: client}
}

// Set guarda un valor en Redis
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    // Verificar si el cliente está disponible
    if r.Client == nil {
        return nil
        // ↑ Si no hay Redis, no fallamos
        // ↑ Simplemente no cacheamos
    }
    return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get obtiene un valor de Redis
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
    if r.Client == nil {
        return "", redis.Nil
        // ↑ Si no hay Redis, simulamos que no existe
        // ↑ El handler consultará PostgreSQL
    }
    return r.Client.Get(ctx, key).Result()
}

// Delete elimina un valor de Redis
func (r *RedisClient) Delete(ctx context.Context, key string) error {
    if r.Client == nil {
        return nil
    }
    return r.Client.Del(ctx, key).Err()
}

// Ping verifica la conexión
func (r *RedisClient) Ping(ctx context.Context) error {
    // ↑ Método agregado para tests
    if r.Client == nil {
        return redis.Nil
    }
    return r.Client.Ping(ctx).Err()
}