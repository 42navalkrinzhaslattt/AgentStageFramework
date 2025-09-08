// Package redis_client provides Redis configuration and utilities
package redis_client

import (
	"os"
	"strconv"
)

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Event types for the game event system
const (
	EventTypePlayerAction    = "player_action"
	EventTypePlayerConnect   = "player_connect"
	EventTypePlayerDisconnect = "player_disconnect"
	EventTypeNPCUpdate       = "npc_update"
	EventTypeNPCDialogue     = "npc_dialogue"
	EventTypeNPCStateChange  = "npc_state_change"
	EventTypeQuestStart      = "quest_start"
	EventTypeQuestComplete   = "quest_complete"
	EventTypeQuestUpdate     = "quest_update"
	EventTypeWorldEvent      = "world_event"
	EventTypeAssetGenerated  = "asset_generated"
)

// Redis key patterns
const (
	KeyPatternNPCState     = "npc:state:%s"
	KeyPatternNPCMemory    = "npc:memory:%s"
	KeyPatternNPCGoals     = "npc:goals:%s"
	KeyPatternPlayerState  = "player:state:%s"
	KeyPatternQuestData    = "quest:data:%s"
	KeyPatternAssetMeta    = "asset:metadata:%s"
	KeyPatternWorldState   = "world:state"
	KeyPatternActiveQuests = "quests:active"
	KeyPatternActiveNPCs   = "npcs:active"
)

// Channel patterns
const (
	ChannelGameEvents   = "events:game"
	ChannelNPCEvents    = "events:npc:%s"
	ChannelPlayerEvents = "events:player:%s"
	ChannelDirector     = "events:director"
)

// GameEvent represents a structured game event
type GameEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// NPCMemoryEntry represents a memory entry for NPCs
type NPCMemoryEntry struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Type        string                 `json:"type"`
	Importance  float64                `json:"importance"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   int64                  `json:"created_at"`
}

// WorldStateData represents the global world state
type WorldStateData struct {
	ActivePlayers int                    `json:"active_players"`
	ActiveNPCs    int                    `json:"active_npcs"`
	ActiveQuests  int                    `json:"active_quests"`
	GlobalFlags   map[string]interface{} `json:"global_flags"`
	LastUpdated   int64                  `json:"last_updated"`
}