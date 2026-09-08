package cache

import (
    "context"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
)

// RedisClient es el cliente de Redis
type RedisClient struct {
    Client *redis.Client
}

// Connect crea una conexión a Redis
func Connect(redisURL string) *RedisClient {
    // Crear cliente de Redis
    client := redis.NewClient(&redis.Options{
        Addr:     redisURL,
        Password: "",
        DB:       0,
    })

    // Verificar conexión
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
    return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get obtiene un valor de Redis
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
    return r.Client.Get(ctx, key).Result()
}

// Delete elimina un valor de Redis
func (r *RedisClient) Delete(ctx context.Context, key string) error {
    return r.Client.Del(ctx, key).Err()
}