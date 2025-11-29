package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-backed translation cache.
type RedisCache struct {
	client    *redis.Client
	ttl       time.Duration
	keyPrefix string
}

// RedisConfig holds configuration for the Redis cache.
type RedisConfig struct {
	URL       string // Redis connection URL (e.g., "redis://localhost:6379")
	TTL       int    // TTL in seconds (0 = no expiration)
	KeyPrefix string // Prefix for all keys (default: "gotlai:")
}

// NewRedisCache creates a new Redis cache with the given configuration.
func NewRedisCache(cfg RedisConfig) (*RedisCache, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "gotlai:"
	}

	ttl := time.Duration(cfg.TTL) * time.Second
	if cfg.TTL <= 0 {
		ttl = 0
	}

	return &RedisCache{
		client:    client,
		ttl:       ttl,
		keyPrefix: prefix,
	}, nil
}

// NewRedisCacheFromClient creates a RedisCache from an existing Redis client.
func NewRedisCacheFromClient(client *redis.Client, ttlSeconds int, keyPrefix string) *RedisCache {
	if keyPrefix == "" {
		keyPrefix = "gotlai:"
	}

	ttl := time.Duration(ttlSeconds) * time.Second
	if ttlSeconds <= 0 {
		ttl = 0
	}

	return &RedisCache{
		client:    client,
		ttl:       ttl,
		keyPrefix: keyPrefix,
	}
}

// Get retrieves a value from Redis.
func (c *RedisCache) Get(key string) (string, bool) {
	ctx := context.Background()
	val, err := c.client.Get(ctx, c.keyPrefix+key).Result()
	if err == redis.Nil {
		return "", false
	}
	if err != nil {
		// Log error but return as cache miss
		return "", false
	}
	return val, true
}

// Set stores a value in Redis.
func (c *RedisCache) Set(key string, value string) error {
	ctx := context.Background()
	fullKey := c.keyPrefix + key

	if c.ttl > 0 {
		return c.client.Set(ctx, fullKey, value, c.ttl).Err()
	}
	return c.client.Set(ctx, fullKey, value, 0).Err()
}

// Close closes the Redis connection.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Ping tests the Redis connection.
func (c *RedisCache) Ping() error {
	ctx := context.Background()
	return c.client.Ping(ctx).Err()
}

// Verify RedisCache implements TranslationCache
var _ TranslationCache = (*RedisCache)(nil)
