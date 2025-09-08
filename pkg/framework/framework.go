// Package framework provides the core Emergent World Engine framework for AI-driven games
package framework

import (
	"context"
	"fmt"

	"github.com/emergent-world-engine/backend/internal/theta_client"
	"github.com/emergent-world-engine/backend/internal/redis_client"
)

// Engine represents the main framework instance
type Engine struct {
	thetaClient *theta_client.ThetaClient
	redisClient *redis_client.RedisClient
	config      *Config
}

// Config holds framework configuration
type Config struct {
	ThetaAPIKey    string
	ThetaEndpoint  string
	RedisURL       string
	RedisPassword  string
	EnableRedis    bool // Optional Redis for advanced features
	EnableLogging  bool
}

// NewEngine creates a new Emergent World Engine instance
func NewEngine(config *Config) (*Engine, error) {
	if config.ThetaAPIKey == "" {
		return nil, fmt.Errorf("theta API key is required")
	}

	// Initialize Theta client (required for AI features)
	thetaEndpoint := config.ThetaEndpoint
	if thetaEndpoint == "" {
		thetaEndpoint = "https://api.thetaedgecloud.com"
	}
	
	thetaClient := theta_client.NewThetaClient(thetaEndpoint, config.ThetaAPIKey)

	// Initialize Redis client (optional for advanced features)
	var redisClient *redis_client.RedisClient
	if config.EnableRedis && config.RedisURL != "" {
		redisConfig := &redis_client.Config{
			Addr:     config.RedisURL,
			Password: config.RedisPassword,
			DB:       0,
			PoolSize: 10,
		}
		redisClient = redis_client.NewRedisClient(redisConfig)
	}

	return &Engine{
		thetaClient: thetaClient,
		redisClient: redisClient,
		config:      config,
	}, nil
}

// NewNPC creates a new NPC instance
func (e *Engine) NewNPC(id string, opts ...NPCOption) *NPC {
	npc := &NPC{
		id:          id,
		engine:      e,
		memory:      make(map[string]interface{}),
		personality: make(map[string]interface{}),
		state:       make(map[string]interface{}),
	}
	
	// Apply options
	for _, opt := range opts {
		opt(npc)
	}
	
	return npc
}

// NewDirector creates a new Game Director instance
func (e *Engine) NewDirector(opts ...DirectorOption) *Director {
	director := &Director{
		engine:    e,
		gameState: make(map[string]interface{}),
		config:    &DirectorConfig{},
	}
	
	// Apply options
	for _, opt := range opts {
		opt(director)
	}
	
	return director
}

// NewNarrative creates a new narrative system instance
func (e *Engine) NewNarrative(opts ...NarrativeOption) *Narrative {
	narrative := &Narrative{
		engine:      e,
		storyState:  make(map[string]interface{}),
		lore:        make(map[string]interface{}),
		activeQuests: make(map[string]*Quest),
	}
	
	// Apply options
	for _, opt := range opts {
		opt(narrative)
	}
	
	return narrative
}

// NewAssetGenerator creates a new asset generation system
func (e *Engine) NewAssetGenerator(opts ...AssetOption) *AssetGenerator {
	generator := &AssetGenerator{
		engine: e,
		cache:  make(map[string]*Asset),
	}
	
	// Apply options
	for _, opt := range opts {
		opt(generator)
	}
	
	return generator
}

// Close gracefully shuts down the engine
func (e *Engine) Close() error {
	if e.redisClient != nil {
		return e.redisClient.Close()
	}
	return nil
}

// IsRedisEnabled returns whether Redis features are available
func (e *Engine) IsRedisEnabled() bool {
	return e.redisClient != nil
}

// Health checks the health of connected services
func (e *Engine) Health(ctx context.Context) error {
	// Test Theta connection with a simple request
	_, err := e.thetaClient.GenerateWithLLM(ctx, &theta_client.LLMRequest{
		Model:     "gpt-oss-20b",
		Prompt:    "Hello",
		MaxTokens: 1,
	})
	if err != nil {
		return fmt.Errorf("theta client health check failed: %w", err)
	}
	
	// Test Redis if enabled
	if e.redisClient != nil {
		if err := e.redisClient.Ping(ctx); err != nil {
			return fmt.Errorf("redis health check failed: %w", err)
		}
	}
	
	return nil
}