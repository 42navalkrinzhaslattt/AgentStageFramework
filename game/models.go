package main

import "time"

// GameEvent represents a random event that occurs each turn
type GameEvent struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"` // economy, security, diplomacy, environment, domestic
	Severity    int      `json:"severity"` // 1-10
	Options     []string `json:"options"`  // Available choices for the player
	ImageURL    string   `json:"imageUrl,omitempty"`
}

// Advisor represents one of the 8 possible advisors
type Advisor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Personality string `json:"personality"`
	Specialty   string `json:"specialty"` // economy, security, diplomacy, environment, domestic, military, social, tech
}

// AdvisorResponse represents advice from an advisor about an event
type AdvisorResponse struct {
	AdvisorID      string `json:"advisorId"`
	AdvisorName    string `json:"advisorName"`
	Name           string `json:"name,omitempty"`
	Title          string `json:"title,omitempty"`
	Advice         string `json:"advice"`
	Recommendation int    `json:"recommendation"` // Which option they recommend (0-based index)
}

// PlayerChoice represents the player's decision
type PlayerChoice struct {
	EventID     string `json:"eventId"`
	OptionIndex int    `json:"optionIndex"`
	Option      string `json:"option"`
	Reasoning   string `json:"reasoning"`
}

// WorldMetrics holds the global state metrics
type WorldMetrics struct {
	Economy     float64 `json:"economy"`    // -100 to 100
	Security    float64 `json:"security"`   // -100 to 100
	Diplomacy   float64 `json:"diplomacy"`  // -100 to 100
	Environment float64 `json:"environment"`// -100 to 100
	Approval    float64 `json:"approval"`   // -100 to 100
	Stability   float64 `json:"stability"`  // -100 to 100
}

// TurnResult represents the outcome of a turn
type TurnResult struct {
	Turn       int           `json:"turn"`
	Event      GameEvent     `json:"event"`
	Advisors   []AdvisorResponse `json:"advisors"`
	Choice     PlayerChoice  `json:"choice"`
	Evaluation string        `json:"evaluation"`
	Impact     WorldMetrics  `json:"impact"`
}

// AIUsageStats holds the statistics for AI usage
type AIUsageStats struct {
	AdvisorTheta   int `json:"advisorTheta"`
	AdvisorGemini  int `json:"advisorGemini"`
	DirectorTheta  int `json:"directorTheta"`
	DirectorGemini int `json:"directorGemini"`
	RewriteGemini  int `json:"rewriteGemini"`
}

// GameState holds the current game state
type GameState struct {
	Turn        int          `json:"turn"`
	MaxTurns    int          `json:"maxTurns"`
	Metrics     WorldMetrics `json:"metrics"`
	History     []TurnResult `json:"history"`
	Advisors    []Advisor    `json:"advisors"`
	CurrentTurn *TurnResult  `json:"currentTurn,omitempty"`
	LastUpdated time.Time    `json:"lastUpdated"`
	Stats       AIUsageStats `json:"stats"`
}
