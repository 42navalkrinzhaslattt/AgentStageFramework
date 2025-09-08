package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/emergent-world-engine/backend/internal/theta_client"
)

// Director represents the AI Game Director for strategic decisions
type Director struct {
	engine    *Engine
	gameState map[string]interface{}
	config    *DirectorConfig
}

// DirectorConfig holds director-specific configuration
type DirectorConfig struct {
	ReasoningModel    string
	StrategicFocus    string // e.g., "balance", "narrative", "challenge"
	PlayerAnalysis    bool
	EventGeneration   bool
	DifficultyScaling bool
}

// DirectorOption allows configuring Director behavior
type DirectorOption func(*Director)

// WithStrategicFocus sets the director's primary focus
func WithStrategicFocus(focus string) DirectorOption {
	return func(d *Director) {
		if d.config == nil {
			d.config = &DirectorConfig{}
		}
		d.config.StrategicFocus = focus
	}
}

// WithPlayerAnalysis enables player behavior analysis
func WithPlayerAnalysis(enabled bool) DirectorOption {
	return func(d *Director) {
		if d.config == nil {
			d.config = &DirectorConfig{}
		}
		d.config.PlayerAnalysis = enabled
	}
}

// WithEventGeneration enables dynamic event creation
func WithEventGeneration(enabled bool) DirectorOption {
	return func(d *Director) {
		if d.config == nil {
			d.config = &DirectorConfig{}
		}
		d.config.EventGeneration = enabled
	}
}

// WithDifficultyScaling enables automatic difficulty adjustment
func WithDifficultyScaling(enabled bool) DirectorOption {
	return func(d *Director) {
		if d.config == nil {
			d.config = &DirectorConfig{}
		}
		d.config.DifficultyScaling = enabled
	}
}

// GameEvent represents an event that occurred in the game
type GameEvent struct {
	Type        string                 `json:"type"`
	PlayerID    string                 `json:"player_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Location    string                 `json:"location"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Context     *GameContext           `json:"context"`
}

// DirectorDecision represents a strategic decision made by the director
type DirectorDecision struct {
	Decision    string                 `json:"decision"`
	Reasoning   string                 `json:"reasoning"`
	Actions     []DirectorAction       `json:"actions"`
	Confidence  float64                `json:"confidence"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DirectorAction represents an action the director wants to execute
type DirectorAction struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
	Delay      time.Duration          `json:"delay"`
}

// ProcessEvent analyzes a game event and makes strategic decisions
func (d *Director) ProcessEvent(ctx context.Context, event *GameEvent) (*DirectorDecision, error) {
	// Build context-aware prompt for strategic reasoning
	prompt := d.buildEventAnalysisPrompt(event)
	
	// Get reasoning model (default to DeepSeek R1 for strategic decisions)
	model := "deepseek_r1"
	if d.config != nil && d.config.ReasoningModel != "" {
		model = d.config.ReasoningModel
	}
	
	// Generate strategic response using LLM
	llmReq := &theta_client.LLMRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   300,
		Temperature: 0.6, // Lower temperature for more consistent strategic decisions
	}
	
	llmResp, err := d.engine.thetaClient.GenerateWithLLM(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to process event: %w", err)
	}
	
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("no decision generated")
	}
	
	reasoning := llmResp.Choices[0].Text
	
	// Create decision
	decision := &DirectorDecision{
		Decision:   "analyze_and_respond",
		Reasoning:  reasoning,
		Actions:    d.generateActions(event),
		Confidence: 0.8,
		Priority:   d.calculatePriority(event),
		Metadata:   map[string]interface{}{
			"event_type": event.Type,
			"player_id":  event.PlayerID,
			"timestamp":  event.Timestamp,
		},
	}
	
	// Store decision for future reference
	d.storeDecision(event, decision)
	
	return decision, nil
}

// AnalyzePlayerBehavior analyzes player patterns and suggests adaptations
func (d *Director) AnalyzePlayerBehavior(ctx context.Context, playerID string, events []GameEvent) (*PlayerAnalysis, error) {
	if d.config == nil || !d.config.PlayerAnalysis {
		return nil, fmt.Errorf("player analysis not enabled")
	}
	
	// Build analysis prompt
	prompt := d.buildPlayerAnalysisPrompt(playerID, events)
	
	model := "deepseek_r1"
	if d.config.ReasoningModel != "" {
		model = d.config.ReasoningModel
	}
	
	llmReq := &theta_client.LLMRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   400,
		Temperature: 0.7,
	}
	
	llmResp, err := d.engine.thetaClient.GenerateWithLLM(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze player behavior: %w", err)
	}
	
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("no analysis generated")
	}
	
	return &PlayerAnalysis{
		PlayerID:      playerID,
		Analysis:      llmResp.Choices[0].Text,
		PlayStyle:     d.extractPlayStyle(events),
		Recommendations: d.generateRecommendations(events),
		Confidence:    0.75,
		Timestamp:     time.Now(),
	}, nil
}

// GenerateEvent creates dynamic events based on current game state
func (d *Director) GenerateEvent(ctx context.Context, context *GameContext) (*GeneratedEvent, error) {
	if d.config == nil || !d.config.EventGeneration {
		return nil, fmt.Errorf("event generation not enabled")
	}
	
	// Build event generation prompt
	prompt := d.buildEventGenerationPrompt(context)
	
	model := "deepseek_r1"
	if d.config.ReasoningModel != "" {
		model = d.config.ReasoningModel
	}
	
	llmReq := &theta_client.LLMRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   250,
		Temperature: 0.9, // Higher temperature for creative event generation
	}
	
	llmResp, err := d.engine.thetaClient.GenerateWithLLM(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate event: %w", err)
	}
	
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("no event generated")
	}
	
	return &GeneratedEvent{
		Type:        "dynamic",
		Title:       "AI Generated Event",
		Description: llmResp.Choices[0].Text,
		Location:    context.Location,
		Triggers:    []string{"time_based", "location_based"},
		Rewards:     map[string]interface{}{"xp": 100},
		Timestamp:   time.Now(),
	}, nil
}

// AdjustDifficulty automatically adjusts game difficulty based on player performance
func (d *Director) AdjustDifficulty(ctx context.Context, playerStats *PlayerStats) (*DifficultyAdjustment, error) {
	if d.config == nil || !d.config.DifficultyScaling {
		return nil, fmt.Errorf("difficulty scaling not enabled")
	}
	
	// Calculate current difficulty metrics
	successRate := float64(playerStats.SuccessfulActions) / float64(playerStats.TotalActions)
	avgCompletionTime := playerStats.AvgCompletionTime
	deathRate := float64(playerStats.Deaths) / float64(playerStats.Sessions)
	
	// Determine adjustment
	var adjustment float64
	var reasoning string
	
	switch {
	case successRate > 0.85 && deathRate < 0.1:
		adjustment = 0.15 // Increase difficulty
		reasoning = "Player performing very well, increasing challenge"
	case successRate < 0.4 || deathRate > 0.5:
		adjustment = -0.20 // Decrease difficulty
		reasoning = "Player struggling, reducing difficulty for better experience"
	case successRate > 0.7 && avgCompletionTime < playerStats.TargetCompletionTime*0.7:
		adjustment = 0.1
		reasoning = "Player completing tasks too quickly, slight difficulty increase"
	default:
		adjustment = 0.0
		reasoning = "Player performance within target range, no adjustment needed"
	}
	
	return &DifficultyAdjustment{
		PlayerID:      playerStats.PlayerID,
		Adjustment:    adjustment,
		NewDifficulty: playerStats.CurrentDifficulty + adjustment,
		Reasoning:     reasoning,
		Metrics: map[string]float64{
			"success_rate":      successRate,
			"death_rate":        deathRate,
			"completion_time":   avgCompletionTime,
		},
		Timestamp: time.Now(),
	}, nil
}

// UpdateGameState updates the director's understanding of the game world
func (d *Director) UpdateGameState(key string, value interface{}) {
	d.gameState[key] = value
}

// GetGameState retrieves current game state information
func (d *Director) GetGameState(key string) (interface{}, bool) {
	value, exists := d.gameState[key]
	return value, exists
}

// PlayerAnalysis contains insights about player behavior
type PlayerAnalysis struct {
	PlayerID        string                 `json:"player_id"`
	Analysis        string                 `json:"analysis"`
	PlayStyle       string                 `json:"play_style"`
	Recommendations []string               `json:"recommendations"`
	Confidence      float64                `json:"confidence"`
	Timestamp       time.Time              `json:"timestamp"`
}

// GeneratedEvent represents a dynamically created game event
type GeneratedEvent struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Location    string                 `json:"location"`
	Triggers    []string               `json:"triggers"`
	Rewards     map[string]interface{} `json:"rewards"`
	Timestamp   time.Time              `json:"timestamp"`
}

// PlayerStats contains player performance metrics
type PlayerStats struct {
	PlayerID             string    `json:"player_id"`
	TotalActions         int       `json:"total_actions"`
	SuccessfulActions    int       `json:"successful_actions"`
	Deaths               int       `json:"deaths"`
	Sessions             int       `json:"sessions"`
	AvgCompletionTime    float64   `json:"avg_completion_time"`
	TargetCompletionTime float64   `json:"target_completion_time"`
	CurrentDifficulty    float64   `json:"current_difficulty"`
	LastUpdated          time.Time `json:"last_updated"`
}

// DifficultyAdjustment represents a change in game difficulty
type DifficultyAdjustment struct {
	PlayerID      string             `json:"player_id"`
	Adjustment    float64            `json:"adjustment"`
	NewDifficulty float64            `json:"new_difficulty"`
	Reasoning     string             `json:"reasoning"`
	Metrics       map[string]float64 `json:"metrics"`
	Timestamp     time.Time          `json:"timestamp"`
}

// Helper methods

func (d *Director) buildEventAnalysisPrompt(event *GameEvent) string {
	prompt := fmt.Sprintf("As an AI Game Director, analyze this game event and provide strategic guidance.\n")
	prompt += fmt.Sprintf("Event Type: %s\n", event.Type)
	prompt += fmt.Sprintf("Player Action: %s\n", event.Action)
	prompt += fmt.Sprintf("Location: %s\n", event.Location)
	
	if d.config != nil && d.config.StrategicFocus != "" {
		prompt += fmt.Sprintf("Strategic Focus: %s\n", d.config.StrategicFocus)
	}
	
	prompt += "Provide strategic analysis and recommend actions to enhance the player experience:"
	
	return prompt
}

func (d *Director) buildPlayerAnalysisPrompt(playerID string, events []GameEvent) string {
	prompt := fmt.Sprintf("Analyze player %s's behavior based on recent actions:\n", playerID)
	
	for i, event := range events {
		if i >= 10 { // Limit to last 10 events
			break
		}
		prompt += fmt.Sprintf("- %s: %s at %s\n", event.Type, event.Action, event.Location)
	}
	
	prompt += "Identify the player's style, preferences, and suggest how to improve their experience:"
	
	return prompt
}

func (d *Director) buildEventGenerationPrompt(context *GameContext) string {
	prompt := "Generate an interesting game event based on current context:\n"
	prompt += fmt.Sprintf("Location: %s\n", context.Location)
	prompt += fmt.Sprintf("Time: %s\n", context.TimeOfDay)
	
	if context.Weather != "" {
		prompt += fmt.Sprintf("Weather: %s\n", context.Weather)
	}
	
	prompt += "Create an engaging event that would enhance the player's experience:"
	
	return prompt
}

func (d *Director) generateActions(event *GameEvent) []DirectorAction {
	actions := []DirectorAction{
		{
			Type:   "log_event",
			Target: "analytics",
			Parameters: map[string]interface{}{
				"event_type": event.Type,
				"player_id":  event.PlayerID,
			},
		},
	}
	
	// Add context-specific actions based on event type
	switch event.Type {
	case "player_death":
		actions = append(actions, DirectorAction{
			Type:   "adjust_difficulty",
			Target: "game_balance",
			Parameters: map[string]interface{}{
				"adjustment": -0.1,
				"reason":     "player_death",
			},
		})
	case "quest_completion":
		actions = append(actions, DirectorAction{
			Type:   "generate_reward",
			Target: "reward_system",
			Parameters: map[string]interface{}{
				"type":      "experience",
				"amount":    100,
				"player_id": event.PlayerID,
			},
		})
	}
	
	return actions
}

func (d *Director) calculatePriority(event *GameEvent) int {
	switch event.Type {
	case "player_death", "game_crash":
		return 10 // High priority
	case "quest_completion", "level_up":
		return 7 // Medium-high priority
	case "item_pickup", "npc_interaction":
		return 5 // Medium priority
	default:
		return 3 // Low priority
	}
}

func (d *Director) storeDecision(event *GameEvent, decision *DirectorDecision) {
	if d.engine.IsRedisEnabled() {
		key := fmt.Sprintf("director:decisions:%s:%d", event.PlayerID, event.Timestamp.Unix())
		d.engine.redisClient.Set(context.Background(), key, decision, 24*time.Hour)
	}
}

func (d *Director) extractPlayStyle(events []GameEvent) string {
	// Simple heuristic - could be enhanced with more sophisticated analysis
	actionCounts := make(map[string]int)
	for _, event := range events {
		actionCounts[event.Action]++
	}
	
	// Determine dominant play style
	if actionCounts["combat"] > len(events)/2 {
		return "aggressive"
	} else if actionCounts["explore"] > len(events)/2 {
		return "explorer"
	} else if actionCounts["craft"] > len(events)/3 {
		return "builder"
	}
	
	return "balanced"
}

func (d *Director) generateRecommendations(events []GameEvent) []string {
	recommendations := []string{}
	
	// Basic recommendations based on event patterns
	actionCounts := make(map[string]int)
	for _, event := range events {
		actionCounts[event.Action]++
	}
	
	if actionCounts["combat"] == 0 {
		recommendations = append(recommendations, "Introduce combat tutorial or easier enemies")
	}
	
	if actionCounts["explore"] > int(float64(len(events))*0.8) {
		recommendations = append(recommendations, "Add hidden areas or exploration rewards")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Continue current game balance")
	}
	
	return recommendations
}