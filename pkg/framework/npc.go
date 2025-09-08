package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/emergent-world-engine/backend/internal/theta_client"
)

// NPC represents an AI-driven non-player character
type NPC struct {
	id          string
	engine      *Engine
	memory      map[string]interface{}
	personality map[string]interface{}
	state       map[string]interface{}
	config      *NPCConfig
}

// NPCConfig holds NPC-specific configuration
type NPCConfig struct {
	DialogueModel  string
	VoiceModel     string
	VisionModel    string
	Personality    string
	Background     string
	Relationships  map[string]string
	MemoryLimit    int
	EnableVoice    bool
	EnableVision   bool
}

// NPCOption allows configuring NPC behavior
type NPCOption func(*NPC)

// WithPersonality sets the NPC's personality description
func WithPersonality(personality string) NPCOption {
	return func(npc *NPC) {
		if npc.config == nil {
			npc.config = &NPCConfig{}
		}
		npc.config.Personality = personality
	}
}

// WithBackground sets the NPC's background story
func WithBackground(background string) NPCOption {
	return func(npc *NPC) {
		if npc.config == nil {
			npc.config = &NPCConfig{}
		}
		npc.config.Background = background
	}
}

// WithVoice enables text-to-speech for this NPC
func WithVoice(enabled bool) NPCOption {
	return func(npc *NPC) {
		if npc.config == nil {
			npc.config = &NPCConfig{}
		}
		npc.config.EnableVoice = enabled
		if enabled && npc.config.VoiceModel == "" {
			npc.config.VoiceModel = "kokoro-82m"
		}
	}
}

// WithVision enables environmental perception for this NPC
func WithVision(enabled bool) NPCOption {
	return func(npc *NPC) {
		if npc.config == nil {
			npc.config = &NPCConfig{}
		}
		npc.config.EnableVision = enabled
		if enabled && npc.config.VisionModel == "" {
			npc.config.VisionModel = "grounding-dino"
		}
	}
}

// WithRelationship defines a relationship with another entity
func WithRelationship(entityID, relationship string) NPCOption {
	return func(npc *NPC) {
		if npc.config == nil {
			npc.config = &NPCConfig{}
		}
		if npc.config.Relationships == nil {
			npc.config.Relationships = make(map[string]string)
		}
		npc.config.Relationships[entityID] = relationship
	}
}

// DialogueRequest contains context for generating dialogue
type DialogueRequest struct {
	PlayerMessage string
	Context       *GameContext
	History       []DialogueEntry
}

// DialogueResponse contains the generated dialogue
type DialogueResponse struct {
	Message     string
	AudioData   []byte // If voice is enabled
	Emotion     string
	Actions     []string // Suggested actions the NPC might take
	Memory      []string // New memories formed
}

// DialogueEntry represents a single dialogue exchange
type DialogueEntry struct {
	Speaker   string
	Message   string
	Timestamp time.Time
}

// GameContext provides situational awareness for NPCs
type GameContext struct {
	Location     string
	TimeOfDay    string
	Weather      string
	NearbyNPCs   []string
	NearbyItems  []string
	PlayerStats  map[string]interface{}
	GameState    map[string]interface{}
	Environment  string
}

// GenerateDialogue creates contextual dialogue for the NPC
func (npc *NPC) GenerateDialogue(ctx context.Context, req *DialogueRequest) (*DialogueResponse, error) {
	// Build context-aware prompt
	prompt := npc.buildDialoguePrompt(req)
	
	// Get dialogue model (default to GPT-OSS-20B for dialogue)
	model := "gpt-oss-20b"
	if npc.config != nil && npc.config.DialogueModel != "" {
		model = npc.config.DialogueModel
	}
	
	// Generate dialogue using LLM
	llmReq := &theta_client.LLMRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   150,
		Temperature: 0.8,
	}
	
	llmResp, err := npc.engine.thetaClient.GenerateWithLLM(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dialogue: %w", err)
	}
	
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("no dialogue generated")
	}
	
	dialogue := llmResp.Choices[0].Text
	
	// Create response
	response := &DialogueResponse{
		Message: dialogue,
		Emotion: "neutral", // Could be enhanced with emotion detection
	}
	
	// Generate voice if enabled
	if npc.config != nil && npc.config.EnableVoice {
		audioData, err := npc.generateVoice(ctx, dialogue)
		if err == nil {
			response.AudioData = audioData
		}
	}
	
	// Store in memory if Redis is available
	if npc.engine.IsRedisEnabled() {
		npc.addToMemory(req.PlayerMessage, dialogue)
	}
	
	return response, nil
}

// Perceive analyzes the visual environment using AI vision
func (npc *NPC) Perceive(ctx context.Context, imageData []byte, query string) (*PerceptionResult, error) {
	if npc.config == nil || !npc.config.EnableVision {
		return nil, fmt.Errorf("vision not enabled for this NPC")
	}
	
	visionReq := &theta_client.VisionRequest{
		Image: imageData,
		Query: query,
	}
	
	visionResp, err := npc.engine.thetaClient.AnalyzeVision(ctx, visionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze vision: %w", err)
	}
	
	// Convert detections to our format
	perceptions := make([]Perception, len(visionResp.Detections))
	for i, detection := range visionResp.Detections {
		perceptions[i] = Perception{
			Object:     detection.Label,
			Confidence: detection.Confidence,
			Location:   fmt.Sprintf("x:%f y:%f", detection.BoundingBox.X, detection.BoundingBox.Y),
		}
	}
	
	return &PerceptionResult{
		Description: visionResp.Description,
		Objects:     perceptions,
	}, nil
}

// UpdateMemory adds new information to the NPC's memory
func (npc *NPC) UpdateMemory(key string, value interface{}) {
	npc.memory[key] = value
	
	// Also store in Redis if available for persistence
	if npc.engine.IsRedisEnabled() {
		memoryKey := fmt.Sprintf("npc:%s:memory:%s", npc.id, key)
		npc.engine.redisClient.Set(context.Background(), memoryKey, value, 24*time.Hour)
	}
}

// GetMemory retrieves information from the NPC's memory
func (npc *NPC) GetMemory(key string) (interface{}, bool) {
	// Check local memory first
	if value, exists := npc.memory[key]; exists {
		return value, true
	}
	
	// Check Redis if available
	if npc.engine.IsRedisEnabled() {
		memoryKey := fmt.Sprintf("npc:%s:memory:%s", npc.id, key)
		var value interface{}
		if err := npc.engine.redisClient.Get(context.Background(), memoryKey, &value); err == nil {
			npc.memory[key] = value // Cache locally
			return value, true
		}
	}
	
	return nil, false
}

// GetState returns the current state of the NPC
func (npc *NPC) GetState() map[string]interface{} {
	return npc.state
}

// SetState updates the NPC's state
func (npc *NPC) SetState(key string, value interface{}) {
	npc.state[key] = value
}

// PerceptionResult contains the results of environmental perception
type PerceptionResult struct {
	Description string
	Objects     []Perception
}

// Perception represents a single perceived object or entity
type Perception struct {
	Object     string
	Confidence float64
	Location   string
}

// buildDialoguePrompt creates a context-aware prompt for dialogue generation
func (npc *NPC) buildDialoguePrompt(req *DialogueRequest) string {
	prompt := fmt.Sprintf("You are %s.", npc.id)
	
	if npc.config != nil {
		if npc.config.Personality != "" {
			prompt += fmt.Sprintf(" Your personality: %s.", npc.config.Personality)
		}
		if npc.config.Background != "" {
			prompt += fmt.Sprintf(" Your background: %s.", npc.config.Background)
		}
	}
	
	if req.Context != nil {
		if req.Context.Location != "" {
			prompt += fmt.Sprintf(" You are currently in %s.", req.Context.Location)
		}
		if req.Context.TimeOfDay != "" {
			prompt += fmt.Sprintf(" It is %s.", req.Context.TimeOfDay)
		}
		if req.Context.Environment != "" {
			prompt += fmt.Sprintf(" The environment: %s.", req.Context.Environment)
		}
	}
	
	// Add recent dialogue history
	if len(req.History) > 0 {
		prompt += " Recent conversation:"
		for _, entry := range req.History {
			prompt += fmt.Sprintf(" %s: %s", entry.Speaker, entry.Message)
		}
	}
	
	prompt += fmt.Sprintf(" Player says: \"%s\" Respond naturally as the character:", req.PlayerMessage)
	
	return prompt
}

// generateVoice creates speech audio for the given text
func (npc *NPC) generateVoice(ctx context.Context, text string) ([]byte, error) {
	if npc.config == nil || npc.config.VoiceModel == "" {
		return nil, fmt.Errorf("voice model not configured")
	}
	
	ttsReq := &theta_client.TTSRequest{
		Text:  text,
		Voice: npc.config.VoiceModel,
	}
	
	ttsResp, err := npc.engine.thetaClient.GenerateVoice(ctx, ttsReq)
	if err != nil {
		return nil, err
	}
	
	return ttsResp.AudioData, nil
}

// addToMemory stores dialogue in memory (local and Redis if available)
func (npc *NPC) addToMemory(playerMessage, npcResponse string) {
	timestamp := time.Now()
	
	// Store in local memory
	memoryKey := fmt.Sprintf("dialogue_%d", timestamp.Unix())
	npc.memory[memoryKey] = DialogueEntry{
		Speaker:   "player",
		Message:   playerMessage,
		Timestamp: timestamp,
	}
	
	responseKey := fmt.Sprintf("response_%d", timestamp.Unix())
	npc.memory[responseKey] = DialogueEntry{
		Speaker:   npc.id,
		Message:   npcResponse,
		Timestamp: timestamp,
	}
}