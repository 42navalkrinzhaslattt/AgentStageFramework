package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/emergent-world-engine/backend/internal/theta_client"
)

// Narrative represents the dynamic storytelling system
type Narrative struct {
	engine       *Engine
	storyState   map[string]interface{}
	lore         map[string]interface{}
	activeQuests map[string]*Quest
	config       *NarrativeConfig
}

// NarrativeConfig holds narrative system configuration
type NarrativeConfig struct {
	StoryModel        string
	Genre             string
	Tone              string
	ComplexityLevel   string
	BranchingFactor   int
	ConsistencyCheck  bool
	PlayerChoice      bool
}

// NarrativeOption allows configuring narrative behavior
type NarrativeOption func(*Narrative)

// WithGenre sets the narrative genre (fantasy, sci-fi, horror, etc.)
func WithGenre(genre string) NarrativeOption {
	return func(n *Narrative) {
		if n.config == nil {
			n.config = &NarrativeConfig{}
		}
		n.config.Genre = genre
	}
}

// WithTone sets the narrative tone (dark, heroic, humorous, etc.)
func WithTone(tone string) NarrativeOption {
	return func(n *Narrative) {
		if n.config == nil {
			n.config = &NarrativeConfig{}
		}
		n.config.Tone = tone
	}
}

// WithPlayerChoice enables player choice tracking for branching narratives
func WithPlayerChoice(enabled bool) NarrativeOption {
	return func(n *Narrative) {
		if n.config == nil {
			n.config = &NarrativeConfig{}
		}
		n.config.PlayerChoice = enabled
	}
}

// WithConsistencyCheck enables lore consistency validation
func WithConsistencyCheck(enabled bool) NarrativeOption {
	return func(n *Narrative) {
		if n.config == nil {
			n.config = &NarrativeConfig{}
		}
		n.config.ConsistencyCheck = enabled
	}
}

// Quest represents a game quest with dynamic elements
type Quest struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Objectives   []Objective            `json:"objectives"`
	Rewards      map[string]interface{} `json:"rewards"`
	Prerequisites []string              `json:"prerequisites"`
	Status       string                 `json:"status"` // "available", "active", "completed", "failed"
	Type         string                 `json:"type"`   // "main", "side", "fetch", "kill", "escort"
	Difficulty   int                    `json:"difficulty"`
	EstimatedTime time.Duration         `json:"estimated_time"`
	Location     string                 `json:"location"`
	NPCGiver     string                 `json:"npc_giver"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

// Objective represents a single quest objective
type Objective struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // "kill", "collect", "talk", "reach"
	Target      string                 `json:"target"`
	Current     int                    `json:"current"`
	Required    int                    `json:"required"`
	Completed   bool                   `json:"completed"`
	Optional    bool                   `json:"optional"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StoryEvent represents a narrative event that affects the story
type StoryEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`
	Choices     []Choice               `json:"choices"`
	Consequences map[string]interface{} `json:"consequences"`
	Timestamp   time.Time              `json:"timestamp"`
	Location    string                 `json:"location"`
	Characters  []string               `json:"characters"`
}

// Choice represents a player choice in the narrative
type Choice struct {
	ID          string                 `json:"id"`
	Text        string                 `json:"text"`
	Consequences map[string]interface{} `json:"consequences"`
	Requirements []string              `json:"requirements"`
	Impact      string                 `json:"impact"`
}

// GenerateQuest creates a new quest based on current game state and player context
func (n *Narrative) GenerateQuest(ctx context.Context, playerContext *GameContext) (*Quest, error) {
	// Build quest generation prompt
	prompt := n.buildQuestGenerationPrompt(playerContext)
	
	// Get story model (default to DeepSeek R1 for complex narrative generation)
	model := "deepseek_r1"
	if n.config != nil && n.config.StoryModel != "" {
		model = n.config.StoryModel
	}
	
	// Generate quest using LLM
	llmReq := &theta_client.LLMRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   400,
		Temperature: 0.8,
	}
	
	llmResp, err := n.engine.thetaClient.GenerateWithLLM(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quest: %w", err)
	}
	
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("no quest generated")
	}
	
	questContent := llmResp.Choices[0].Text
	
	// Parse and structure the quest
	quest := n.parseGeneratedQuest(questContent, playerContext)
	
	// Add to active quests
	n.activeQuests[quest.ID] = quest
	
	// Store in Redis if available
	if n.engine.IsRedisEnabled() {
		questKey := fmt.Sprintf("narrative:quest:%s", quest.ID)
		n.engine.redisClient.Set(ctx, questKey, quest, 7*24*time.Hour)
	}
	
	return quest, nil
}

// GenerateStoryEvent creates dynamic story events that affect the narrative
func (n *Narrative) GenerateStoryEvent(ctx context.Context, eventContext *EventContext) (*StoryEvent, error) {
	// Build story event prompt
	prompt := n.buildStoryEventPrompt(eventContext)
	
	model := "deepseek_r1"
	if n.config != nil && n.config.StoryModel != "" {
		model = n.config.StoryModel
	}
	
	llmReq := &theta_client.LLMRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   350,
		Temperature: 0.9, // Higher creativity for story events
	}
	
	llmResp, err := n.engine.thetaClient.GenerateWithLLM(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate story event: %w", err)
	}
	
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("no story event generated")
	}
	
	eventContent := llmResp.Choices[0].Text
	
	// Create structured story event
	event := &StoryEvent{
		ID:          fmt.Sprintf("event_%d", time.Now().Unix()),
		Type:        eventContext.Type,
		Title:       "Dynamic Story Event",
		Description: eventContent,
		Impact:      "medium",
		Timestamp:   time.Now(),
		Location:    eventContext.Location,
		Characters:  eventContext.Characters,
	}
	
	// Generate choices if player choice is enabled
	if n.config != nil && n.config.PlayerChoice {
		choices, err := n.generateChoices(ctx, event)
		if err == nil {
			event.Choices = choices
		}
	}
	
	return event, nil
}

// UpdateLore adds or updates lore information with consistency checking
func (n *Narrative) UpdateLore(key string, loreEntry *LoreEntry) error {
	// Consistency check if enabled
	if n.config != nil && n.config.ConsistencyCheck {
		if err := n.validateLoreConsistency(loreEntry); err != nil {
			return fmt.Errorf("lore consistency check failed: %w", err)
		}
	}
	
	n.lore[key] = loreEntry
	
	// Store in Redis if available
	if n.engine.IsRedisEnabled() {
		loreKey := fmt.Sprintf("narrative:lore:%s", key)
		n.engine.redisClient.Set(context.Background(), loreKey, loreEntry, 30*24*time.Hour)
	}
	
	return nil
}

// GetLore retrieves lore information
func (n *Narrative) GetLore(key string) (*LoreEntry, bool) {
	// Check local cache first
	if entry, exists := n.lore[key]; exists {
		if loreEntry, ok := entry.(*LoreEntry); ok {
			return loreEntry, true
		}
	}
	
	// Check Redis if available
	if n.engine.IsRedisEnabled() {
		loreKey := fmt.Sprintf("narrative:lore:%s", key)
		var loreEntry LoreEntry
		if err := n.engine.redisClient.Get(context.Background(), loreKey, &loreEntry); err == nil {
			n.lore[key] = &loreEntry // Cache locally
			return &loreEntry, true
		}
	}
	
	return nil, false
}

// TrackPlayerChoice records a player choice and its consequences
func (n *Narrative) TrackPlayerChoice(playerID string, choice *Choice) error {
	if n.config == nil || !n.config.PlayerChoice {
		return fmt.Errorf("player choice tracking not enabled")
	}
	
	// Store choice in narrative state
	choiceKey := fmt.Sprintf("choice_%s_%d", playerID, time.Now().Unix())
	n.storyState[choiceKey] = choice
	
	// Apply consequences
	for key, consequence := range choice.Consequences {
		n.storyState[key] = consequence
	}
	
	return nil
}

// GetActiveQuests returns all active quests
func (n *Narrative) GetActiveQuests() map[string]*Quest {
	return n.activeQuests
}

// UpdateQuestProgress updates the progress of a quest objective
func (n *Narrative) UpdateQuestProgress(questID, objectiveID string, progress int) error {
	quest, exists := n.activeQuests[questID]
	if !exists {
		return fmt.Errorf("quest %s not found", questID)
	}
	
	// Find and update the objective
	for i, objective := range quest.Objectives {
		if objective.ID == objectiveID {
			quest.Objectives[i].Current += progress
			if quest.Objectives[i].Current >= quest.Objectives[i].Required {
				quest.Objectives[i].Completed = true
			}
			
			// Check if quest is completed
			if n.isQuestCompleted(quest) {
				quest.Status = "completed"
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("objective %s not found in quest %s", objectiveID, questID)
}

// EventContext provides context for story event generation
type EventContext struct {
	Type       string
	Location   string
	Characters []string
	Trigger    string
	PlayerActions []string
}

// LoreEntry represents a piece of world lore
type LoreEntry struct {
	ID          string                 `json:"id"`
	Category    string                 `json:"category"` // "character", "location", "event", "item"
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	References  []string               `json:"references"`
	Importance  int                    `json:"importance"` // 1-10 scale
	Verified    bool                   `json:"verified"`
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Helper methods

func (n *Narrative) buildQuestGenerationPrompt(playerContext *GameContext) string {
	prompt := "Generate an engaging quest for the player based on current context:\n"
	
	if n.config != nil {
		if n.config.Genre != "" {
			prompt += fmt.Sprintf("Genre: %s\n", n.config.Genre)
		}
		if n.config.Tone != "" {
			prompt += fmt.Sprintf("Tone: %s\n", n.config.Tone)
		}
	}
	
	if playerContext.Location != "" {
		prompt += fmt.Sprintf("Current Location: %s\n", playerContext.Location)
	}
	
	if len(playerContext.NearbyNPCs) > 0 {
		prompt += fmt.Sprintf("Nearby NPCs: %v\n", playerContext.NearbyNPCs)
	}
	
	prompt += "Create a quest with clear objectives, appropriate difficulty, and engaging narrative elements:"
	
	return prompt
}

func (n *Narrative) buildStoryEventPrompt(eventContext *EventContext) string {
	prompt := fmt.Sprintf("Generate a %s story event for the current situation:\n", eventContext.Type)
	prompt += fmt.Sprintf("Location: %s\n", eventContext.Location)
	
	if len(eventContext.Characters) > 0 {
		prompt += fmt.Sprintf("Characters involved: %v\n", eventContext.Characters)
	}
	
	if eventContext.Trigger != "" {
		prompt += fmt.Sprintf("Triggered by: %s\n", eventContext.Trigger)
	}
	
	if n.config != nil {
		if n.config.Genre != "" {
			prompt += fmt.Sprintf("Genre: %s\n", n.config.Genre)
		}
		if n.config.Tone != "" {
			prompt += fmt.Sprintf("Tone: %s\n", n.config.Tone)
		}
	}
	
	prompt += "Create an impactful story event that advances the narrative:"
	
	return prompt
}

func (n *Narrative) parseGeneratedQuest(questContent string, playerContext *GameContext) *Quest {
	// Simple parsing - could be enhanced with more sophisticated NLP
	questID := fmt.Sprintf("quest_%d", time.Now().Unix())
	
	// Create basic quest structure
	quest := &Quest{
		ID:          questID,
		Title:       "Generated Quest",
		Description: questContent,
		Status:      "available",
		Type:        "side",
		Difficulty:  5, // Medium difficulty
		EstimatedTime: 30 * time.Minute,
		Location:    playerContext.Location,
		CreatedAt:   time.Now(),
		Objectives: []Objective{
			{
				ID:          fmt.Sprintf("%s_obj_1", questID),
				Description: "Complete the quest objective",
				Type:        "general",
				Current:     0,
				Required:    1,
				Completed:   false,
			},
		},
		Rewards: map[string]interface{}{
			"experience": 100,
			"gold":       50,
		},
		Metadata: make(map[string]interface{}),
	}
	
	return quest
}

func (n *Narrative) generateChoices(ctx context.Context, event *StoryEvent) ([]Choice, error) {
	prompt := fmt.Sprintf("Generate 2-3 meaningful player choices for this story event:\n%s\n", event.Description)
	prompt += "Each choice should have different consequences and impact on the story:"
	
	model := "deepseek_r1" // Use DeepSeek R1 for choice generation
	if n.config != nil && n.config.StoryModel != "" {
		model = n.config.StoryModel
	}
	
	llmReq := &theta_client.LLMRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   200,
		Temperature: 0.7,
	}
	
	llmResp, err := n.engine.thetaClient.GenerateWithLLM(ctx, llmReq)
	if err != nil {
		return nil, err
	}
	
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices generated")
	}
	
	// Parse choices from response - simplified implementation
	choices := []Choice{
		{
			ID:   "choice_1",
			Text: "Accept the challenge",
			Consequences: map[string]interface{}{
				"reputation": +10,
				"difficulty": +1,
			},
			Impact: "positive",
		},
		{
			ID:   "choice_2",
			Text: "Decline and seek alternatives",
			Consequences: map[string]interface{}{
				"reputation": -5,
				"safety":     +1,
			},
			Impact: "neutral",
		},
	}
	
	return choices, nil
}

func (n *Narrative) validateLoreConsistency(newEntry *LoreEntry) error {
	// Simple consistency check - could be enhanced with AI-powered validation
	for _, existingEntry := range n.lore {
		if lore, ok := existingEntry.(*LoreEntry); ok {
			if lore.Category == newEntry.Category && lore.Title == newEntry.Title {
				if lore.ID != newEntry.ID {
					return fmt.Errorf("conflicting lore entry: %s already exists", newEntry.Title)
				}
			}
		}
	}
	return nil
}

func (n *Narrative) isQuestCompleted(quest *Quest) bool {
	for _, objective := range quest.Objectives {
		if !objective.Optional && !objective.Completed {
			return false
		}
	}
	return true
}