// Package redis_client provides a shared Redis client for caching and pub/sub operations
package redis_client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client   *redis.Client
	pubsub   *redis.PubSub
	ctx      context.Context
	cancel   context.CancelFunc
}

// Config holds Redis client configuration
type Config struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config *Config) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	ctx, cancel := context.WithCancel(context.Background())
	
	return &RedisClient{
		client: rdb,
		ctx:    ctx,
		cancel: cancel,
	}
}

// NewRedisClientFromEnv creates a Redis client from environment variables
func NewRedisClientFromEnv() *RedisClient {
	config := DefaultConfig()
	return NewRedisClient(config)
}

// DefaultConfig returns default Redis configuration
func DefaultConfig() *Config {
	return &Config{
		Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		Password: getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:       getEnvIntOrDefault("REDIS_DB", 0),
		PoolSize: getEnvIntOrDefault("REDIS_POOL_SIZE", 10),
	}
}

// Ping checks Redis connectivity
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	if r.pubsub != nil {
		r.pubsub.Close()
	}
	r.cancel()
	return r.client.Close()
}

// === Key-Value Operations ===

// Set stores a value with optional expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves and unmarshals a value
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// GetString retrieves a string value
func (r *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// SetString stores a string value
func (r *RedisClient) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Delete removes a key
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// === Hash Operations ===

// HSet sets a field in a hash
func (r *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.HSet(ctx, key, field, data).Err()
}

// HGet gets a field from a hash
func (r *RedisClient) HGet(ctx context.Context, key, field string, dest interface{}) error {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// HGetAll gets all fields from a hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDelete deletes fields from a hash
func (r *RedisClient) HDelete(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// === List Operations ===

// LPush pushes values to the left of a list
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush pushes values to the right of a list
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop pops a value from the left of a list
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop pops a value from the right of a list
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LRange gets a range of elements from a list
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

// === Set Operations ===

// SAdd adds members to a set
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers gets all members of a set
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SIsMember checks if a value is a member of a set
func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// SRemove removes members from a set
func (r *RedisClient) SRemove(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// === Pub/Sub Operations ===

// Publish publishes a message to a channel
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return r.client.Publish(ctx, channel, data).Err()
}

// Subscribe subscribes to channels and returns a message channel
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) (<-chan *redis.Message, error) {
	pubsub := r.client.Subscribe(ctx, channels...)
	r.pubsub = pubsub
	
	// Wait for subscription confirmation
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}
	
	return pubsub.Channel(), nil
}

// Unsubscribe unsubscribes from channels
func (r *RedisClient) Unsubscribe(ctx context.Context, channels ...string) error {
	if r.pubsub == nil {
		return fmt.Errorf("no active subscription")
	}
	return r.pubsub.Unsubscribe(ctx, channels...)
}

// === Game-Specific Operations ===

// StoreNPCState stores NPC state data
func (r *RedisClient) StoreNPCState(ctx context.Context, npcID string, state interface{}) error {
	key := fmt.Sprintf("npc:state:%s", npcID)
	return r.Set(ctx, key, state, 0) // No expiration for NPC state
}

// GetNPCState retrieves NPC state data
func (r *RedisClient) GetNPCState(ctx context.Context, npcID string, dest interface{}) error {
	key := fmt.Sprintf("npc:state:%s", npcID)
	return r.Get(ctx, key, dest)
}

// StoreNPCMemory stores an NPC memory
func (r *RedisClient) StoreNPCMemory(ctx context.Context, npcID, memoryID string, memory interface{}) error {
	key := fmt.Sprintf("npc:memory:%s", npcID)
	return r.HSet(ctx, key, memoryID, memory)
}

// GetNPCMemories retrieves all memories for an NPC
func (r *RedisClient) GetNPCMemories(ctx context.Context, npcID string) (map[string]string, error) {
	key := fmt.Sprintf("npc:memory:%s", npcID)
	return r.HGetAll(ctx, key)
}

// StoreWorldState stores global world state
func (r *RedisClient) StoreWorldState(ctx context.Context, state interface{}) error {
	return r.Set(ctx, "world:state", state, 5*time.Minute) // 5 minute expiration
}

// GetWorldState retrieves global world state
func (r *RedisClient) GetWorldState(ctx context.Context, dest interface{}) error {
	return r.Get(ctx, "world:state", dest)
}

// AddActiveQuest adds a quest to the active quests set
func (r *RedisClient) AddActiveQuest(ctx context.Context, questID string) error {
	return r.SAdd(ctx, "quests:active", questID)
}

// RemoveActiveQuest removes a quest from the active quests set
func (r *RedisClient) RemoveActiveQuest(ctx context.Context, questID string) error {
	return r.SRemove(ctx, "quests:active", questID)
}

// GetActiveQuests retrieves all active quest IDs
func (r *RedisClient) GetActiveQuests(ctx context.Context) ([]string, error) {
	return r.SMembers(ctx, "quests:active")
}

// StoreQuestData stores quest data
func (r *RedisClient) StoreQuestData(ctx context.Context, questID string, quest interface{}) error {
	key := fmt.Sprintf("quest:data:%s", questID)
	return r.Set(ctx, key, quest, 24*time.Hour) // 24 hour expiration
}

// GetQuestData retrieves quest data
func (r *RedisClient) GetQuestData(ctx context.Context, questID string, dest interface{}) error {
	key := fmt.Sprintf("quest:data:%s", questID)
	return r.Get(ctx, key, dest)
}

// CacheAsset caches generated asset metadata
func (r *RedisClient) CacheAsset(ctx context.Context, assetID string, metadata interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("asset:metadata:%s", assetID)
	return r.Set(ctx, key, metadata, expiration)
}

// GetCachedAsset retrieves cached asset metadata
func (r *RedisClient) GetCachedAsset(ctx context.Context, assetID string, dest interface{}) error {
	key := fmt.Sprintf("asset:metadata:%s", assetID)
	return r.Get(ctx, key, dest)
}

// === Event System ===

// PublishGameEvent publishes a game event to the events channel
func (r *RedisClient) PublishGameEvent(ctx context.Context, event interface{}) error {
	return r.Publish(ctx, "events:game", event)
}

// PublishNPCEvent publishes an NPC-specific event
func (r *RedisClient) PublishNPCEvent(ctx context.Context, npcID string, event interface{}) error {
	channel := fmt.Sprintf("events:npc:%s", npcID)
	return r.Publish(ctx, channel, event)
}

// PublishPlayerEvent publishes a player-specific event
func (r *RedisClient) PublishPlayerEvent(ctx context.Context, playerID string, event interface{}) error {
	channel := fmt.Sprintf("events:player:%s", playerID)
	return r.Publish(ctx, channel, event)
}

// === Utility Functions ===

// Increment atomically increments a counter
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// SetExpire sets expiration for a key
func (r *RedisClient) SetExpire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// GetTTL gets the remaining time to live for a key
func (r *RedisClient) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}