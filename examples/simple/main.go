package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/emergent-world-engine/backend/pkg/framework"
)

func main() {
	fmt.Println("🎮 Emergent World Engine - Framework Example")
	fmt.Println("============================================")

	// Initialize the framework
	config := &framework.Config{
		ThetaAPIKey:   "your_theta_api_key_here", // Replace with actual key
		EnableRedis:   false,                     // Optional Redis for advanced features
		EnableLogging: true,
	}

	engine, err := framework.NewEngine(config)
	if err != nil {
		log.Fatalf("Failed to initialize engine: %v", err)
	}
	defer engine.Close()

	ctx := context.Background()

	// Test framework health
	fmt.Println("\n🔍 Testing framework health...")
	if err := engine.Health(ctx); err != nil {
		fmt.Printf("⚠️  Health check failed (expected with placeholder API key): %v\n", err)
	} else {
		fmt.Println("✅ Framework is healthy!")
	}

	// Example 1: Create an intelligent NPC
	fmt.Println("\n🤖 Example 1: Creating an Intelligent NPC")
	fmt.Println("==========================================")

	guard := engine.NewNPC("village_guard",
		framework.WithPersonality("Stern but fair, protective of villagers"),
		framework.WithBackground("Former soldier turned village protector"),
		framework.WithVoice(true),
		framework.WithRelationship("mayor", "reports_to"),
	)

	// Simulate NPC dialogue
	dialogueReq := &framework.DialogueRequest{
		PlayerMessage: "Hello, I'm new to this village. Can you tell me about any quests?",
		Context: &framework.GameContext{
			Location:    "village_entrance",
			TimeOfDay:   "morning",
			NearbyNPCs:  []string{"merchant", "villager"},
			Environment: "peaceful village with cobblestone paths",
		},
	}

	fmt.Printf("Player: %s\n", dialogueReq.PlayerMessage)
	
	// Note: This would work with a real Theta API key
	if config.ThetaAPIKey != "your_theta_api_key_here" {
		response, err := guard.GenerateDialogue(ctx, dialogueReq)
		if err != nil {
			fmt.Printf("Error generating dialogue: %v\n", err)
		} else {
			fmt.Printf("Guard: %s\n", response.Message)
			if len(response.AudioData) > 0 {
				fmt.Printf("🔊 Voice audio generated (%d bytes)\n", len(response.AudioData))
			}
		}
	} else {
		fmt.Println("Guard: [Would respond with AI-generated dialogue - requires Theta API key]")
	}

	// Example 2: AI Game Director
	fmt.Println("\n🎯 Example 2: AI Game Director")
	fmt.Println("==============================")

	director := engine.NewDirector(
		framework.WithStrategicFocus("narrative"),
		framework.WithPlayerAnalysis(true),
		framework.WithEventGeneration(true),
	)

	gameEvent := &framework.GameEvent{
		Type:      "player_level_up",
		PlayerID:  "player_001",
		Timestamp: time.Now(),
		Location:  "village_tavern",
		Action:    "reached_level_5",
		Parameters: map[string]interface{}{
			"new_level": 5,
			"class":     "warrior",
		},
	}

	fmt.Printf("Event: Player leveled up to %v\n", gameEvent.Parameters["new_level"])
	
	if config.ThetaAPIKey != "your_theta_api_key_here" {
		decision, err := director.ProcessEvent(ctx, gameEvent)
		if err != nil {
			fmt.Printf("Error processing event: %v\n", err)
		} else {
			fmt.Printf("Director Decision: %s\n", decision.Decision)
			fmt.Printf("Reasoning: %s\n", decision.Reasoning)
		}
	} else {
		fmt.Println("Director: [Would analyze event and make strategic decisions - requires Theta API key]")
	}

	// Example 3: Dynamic Narrative System
	fmt.Println("\n📚 Example 3: Dynamic Narrative System")
	fmt.Println("======================================")

	narrative := engine.NewNarrative(
		framework.WithGenre("fantasy"),
		framework.WithTone("heroic"),
		framework.WithPlayerChoice(true),
	)

	playerContext := &framework.GameContext{
		Location:    "ancient_ruins",
		TimeOfDay:   "dusk",
		NearbyNPCs:  []string{"mysterious_sage"},
		NearbyItems: []string{"glowing_artifact", "ancient_tome"},
	}

	fmt.Printf("Generating quest for location: %s\n", playerContext.Location)
	
	if config.ThetaAPIKey != "your_theta_api_key_here" {
		quest, err := narrative.GenerateQuest(ctx, playerContext)
		if err != nil {
			fmt.Printf("Error generating quest: %v\n", err)
		} else {
			fmt.Printf("Quest: %s\n", quest.Title)
			fmt.Printf("Description: %s\n", quest.Description)
			fmt.Printf("Difficulty: %d/10\n", quest.Difficulty)
		}
	} else {
		fmt.Println("Quest: [Would generate dynamic quest based on context - requires Theta API key]")
	}

	// Example 4: AI Asset Generation
	fmt.Println("\n🎨 Example 4: AI Asset Generation")
	fmt.Println("=================================")

	assetGen := engine.NewAssetGenerator(
		framework.WithImageModel("flux.1-schnell"),
		framework.WithQuality("high"),
		framework.WithCache(true, 24*time.Hour),
		framework.WithDefaultStyle("fantasy art"),
	)

	imageReq := &framework.ImageRequest{
		Prompt: "Ancient mystical sword with glowing runes",
		Style:  "fantasy game art",
		Width:  512,
		Height: 512,
	}

	fmt.Printf("Generating image: %s\n", imageReq.Prompt)
	
	if config.ThetaAPIKey != "your_theta_api_key_here" {
		asset, err := assetGen.GenerateImage(ctx, imageReq)
		if err != nil {
			fmt.Printf("Error generating image: %v\n", err)
		} else {
			fmt.Printf("Generated Asset ID: %s\n", asset.ID)
			fmt.Printf("Dimensions: %dx%d\n", asset.Dimensions.Width, asset.Dimensions.Height)
			fmt.Printf("Format: %s\n", asset.Format)
			if asset.URL != "" {
				fmt.Printf("URL: %s\n", asset.URL)
			}
		}
	} else {
		fmt.Println("Asset: [Would generate AI-powered game assets - requires Theta API key]")
	}

	// Example 5: Framework Integration
	fmt.Println("\n🔧 Example 5: Framework Integration")
	fmt.Println("===================================")

	fmt.Println("Framework Features:")
	fmt.Printf("├── Redis Enabled: %t\n", engine.IsRedisEnabled())
	fmt.Printf("├── Active NPCs: 1 (village_guard)\n")
	fmt.Printf("├── Active Quests: %d\n", len(narrative.GetActiveQuests()))
	fmt.Printf("└── Cached Assets: %d\n", len(assetGen.ListAssets()))

	fmt.Println("\nFramework Usage Patterns:")
	fmt.Println("1. Import the framework package")
	fmt.Println("2. Configure with your Theta EdgeCloud API key")
	fmt.Println("3. Create AI components (NPCs, Director, Narrative, Assets)")
	fmt.Println("4. Use simple method calls - no network setup required!")
	fmt.Println("5. Framework handles all AI integration complexity")

	fmt.Println("\n✨ Framework is ready for integration into your game!")
	fmt.Println("See README.md for complete documentation and examples.")
}