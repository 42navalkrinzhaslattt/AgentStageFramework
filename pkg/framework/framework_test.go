package framework

import (
	"testing"
	"time"
)

// TestEngineInitialization tests basic framework setup
func TestEngineInitialization(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}

	if engine == nil {
		t.Fatal("Engine is nil")
	}

	if engine.config.ThetaAPIKey != "test_key" {
		t.Errorf("Expected API key 'test_key', got '%s'", engine.config.ThetaAPIKey)
	}

	engine.Close()
}

// TestNPCCreation tests NPC creation and configuration
func TestNPCCreation(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	npc := engine.NewNPC("test_npc",
		WithPersonality("Friendly merchant"),
		WithBackground("Former adventurer"),
		WithVoice(true),
	)

	if npc == nil {
		t.Fatal("NPC is nil")
	}

	if npc.id != "test_npc" {
		t.Errorf("Expected NPC ID 'test_npc', got '%s'", npc.id)
	}

	if npc.config == nil {
		t.Fatal("NPC config is nil")
	}

	if npc.config.Personality != "Friendly merchant" {
		t.Errorf("Expected personality 'Friendly merchant', got '%s'", npc.config.Personality)
	}

	if !npc.config.EnableVoice {
		t.Error("Expected voice to be enabled")
	}
}

// TestDirectorCreation tests Director creation and configuration
func TestDirectorCreation(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	director := engine.NewDirector(
		WithStrategicFocus("narrative"),
		WithPlayerAnalysis(true),
		WithEventGeneration(true),
	)

	if director == nil {
		t.Fatal("Director is nil")
	}

	if director.config == nil {
		t.Fatal("Director config is nil")
	}

	if director.config.StrategicFocus != "narrative" {
		t.Errorf("Expected strategic focus 'narrative', got '%s'", director.config.StrategicFocus)
	}

	if !director.config.PlayerAnalysis {
		t.Error("Expected player analysis to be enabled")
	}

	if !director.config.EventGeneration {
		t.Error("Expected event generation to be enabled")
	}
}

// TestNarrativeCreation tests Narrative system creation
func TestNarrativeCreation(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	narrative := engine.NewNarrative(
		WithGenre("fantasy"),
		WithTone("heroic"),
		WithPlayerChoice(true),
	)

	if narrative == nil {
		t.Fatal("Narrative is nil")
	}

	if narrative.config == nil {
		t.Fatal("Narrative config is nil")
	}

	if narrative.config.Genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got '%s'", narrative.config.Genre)
	}

	if narrative.config.Tone != "heroic" {
		t.Errorf("Expected tone 'heroic', got '%s'", narrative.config.Tone)
	}

	if !narrative.config.PlayerChoice {
		t.Error("Expected player choice to be enabled")
	}
}

// TestAssetGeneratorCreation tests Asset generator creation
func TestAssetGeneratorCreation(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	assetGen := engine.NewAssetGenerator(
		WithImageModel("flux.1-schnell"),
		WithQuality("high"),
		WithCache(true, 24*time.Hour),
	)

	if assetGen == nil {
		t.Fatal("AssetGenerator is nil")
	}

	if assetGen.config == nil {
		t.Fatal("AssetGenerator config is nil")
	}

	if assetGen.config.ImageModel != "flux.1-schnell" {
		t.Errorf("Expected image model 'flux.1-schnell', got '%s'", assetGen.config.ImageModel)
	}

	if assetGen.config.Quality != "high" {
		t.Errorf("Expected quality 'high', got '%s'", assetGen.config.Quality)
	}

	if !assetGen.config.CacheEnabled {
		t.Error("Expected caching to be enabled")
	}
}

// TestNPCMemory tests NPC memory functionality
func TestNPCMemory(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	npc := engine.NewNPC("memory_test_npc")

	// Test setting and getting memory
	npc.UpdateMemory("player_name", "TestPlayer")
	npc.UpdateMemory("last_interaction", "friendly_greeting")

	if value, exists := npc.GetMemory("player_name"); !exists || value != "TestPlayer" {
		t.Errorf("Expected memory 'player_name' to be 'TestPlayer', got %v (exists: %t)", value, exists)
	}

	if value, exists := npc.GetMemory("last_interaction"); !exists || value != "friendly_greeting" {
		t.Errorf("Expected memory 'last_interaction' to be 'friendly_greeting', got %v (exists: %t)", value, exists)
	}

	// Test non-existent memory
	if _, exists := npc.GetMemory("non_existent"); exists {
		t.Error("Expected non-existent memory to not exist")
	}
}

// TestNPCState tests NPC state management
func TestNPCState(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	npc := engine.NewNPC("state_test_npc")

	// Test setting and getting state
	npc.SetState("mood", "happy")
	npc.SetState("energy", 85)

	state := npc.GetState()

	if mood, exists := state["mood"]; !exists || mood != "happy" {
		t.Errorf("Expected state 'mood' to be 'happy', got %v (exists: %t)", mood, exists)
	}

	if energy, exists := state["energy"]; !exists || energy != 85 {
		t.Errorf("Expected state 'energy' to be 85, got %v (exists: %t)", energy, exists)
	}
}

// TestDirectorGameState tests Director game state management
func TestDirectorGameState(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	director := engine.NewDirector()

	// Test setting and getting game state
	director.UpdateGameState("player_count", 5)
	director.UpdateGameState("difficulty", "medium")

	if value, exists := director.GetGameState("player_count"); !exists || value != 5 {
		t.Errorf("Expected game state 'player_count' to be 5, got %v (exists: %t)", value, exists)
	}

	if value, exists := director.GetGameState("difficulty"); !exists || value != "medium" {
		t.Errorf("Expected game state 'difficulty' to be 'medium', got %v (exists: %t)", value, exists)
	}
}

// TestAssetCaching tests asset caching functionality
func TestAssetCaching(t *testing.T) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	assetGen := engine.NewAssetGenerator(
		WithCache(true, 1*time.Hour),
	)

	// Create a mock asset
	asset := &Asset{
		ID:          "test_asset_1",
		Type:        "image",
		Prompt:      "test prompt",
		GeneratedAt: time.Now(),
	}

	// Manually add to cache for testing
	key := assetGen.getCacheKey("test prompt", "image")
	expiration := time.Now().Add(1 * time.Hour)
	asset.ExpiresAt = &expiration
	assetGen.cache[key] = asset

	// Test retrieval
	retrieved := assetGen.getCachedAsset("test prompt", "image")
	if retrieved == nil {
		t.Error("Expected to retrieve cached asset, got nil")
	} else if retrieved.ID != "test_asset_1" {
		t.Errorf("Expected cached asset ID 'test_asset_1', got '%s'", retrieved.ID)
	}

	// Test cache listing
	assets := assetGen.ListAssets()
	if len(assets) != 1 {
		t.Errorf("Expected 1 cached asset, got %d", len(assets))
	}

	// Test cache clearing
	assetGen.ClearCache()
	assets = assetGen.ListAssets()
	if len(assets) != 0 {
		t.Errorf("Expected 0 cached assets after clearing, got %d", len(assets))
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	// Test with missing API key
	config := &Config{
		EnableRedis: false,
	}

	_, err := NewEngine(config)
	if err == nil {
		t.Error("Expected error for missing API key, got nil")
	}

	// Test with valid config
	config = &Config{
		ThetaAPIKey: "valid_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
	if engine != nil {
		engine.Close()
	}
}

// BenchmarkNPCCreation benchmarks NPC creation performance
func BenchmarkNPCCreation(b *testing.B) {
	config := &Config{
		ThetaAPIKey: "test_key",
		EnableRedis: false,
	}

	engine, err := NewEngine(config)
	if err != nil {
		b.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		npc := engine.NewNPC("benchmark_npc",
			WithPersonality("Test personality"),
			WithVoice(true),
		)
		_ = npc
	}
}