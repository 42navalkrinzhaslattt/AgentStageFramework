package main

import (
	"context"
	"fmt"
	"log"

	"github.com/emergent-world-engine/backend/pkg/framework"
)

func main() {
	fmt.Println("🎬 Video Generation Example - AgentStage Framework")
	fmt.Println("=================================================")

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

	// Create asset generator with video support
	fmt.Println("\n🎨 Creating Asset Generator...")
	assetGen := engine.NewAssetGenerator(
		framework.WithImageModel("flux.1-schnell"),
		framework.WithVideoModel("stable-diffusion-video"),
		framework.WithQuality("high"),
		framework.WithDefaultStyle("cinematic"),
	)

	// Example 1: Generate a fantasy battle scene video
	fmt.Println("\n⚔️  Example 1: Generating Fantasy Battle Video")
	fmt.Println("=============================================")

	videoReq1 := &framework.VideoRequest{
		Prompt:         "Epic dragon battle in a mystical forest, fire breathing, knights in armor",
		Width:          1920,
		Height:         1080,
		Duration:       15.0, // 15 seconds
		FPS:            30,
		MotionStrength: 0.8,  // High motion for action scene
		Quality:        "high",
		Format:         "mp4",
	}

	video1, err := assetGen.GenerateVideo(ctx, videoReq1)
	if err != nil {
		fmt.Printf("⚠️  Video generation failed (expected with placeholder API key): %v\n", err)
	} else {
		fmt.Printf("✅ Generated fantasy battle video:\n")
		fmt.Printf("   - ID: %s\n", video1.ID)
		fmt.Printf("   - Dimensions: %dx%d\n", video1.Dimensions.Width, video1.Dimensions.Height)
		fmt.Printf("   - Duration: %.1fs at %d FPS\n", video1.Dimensions.Duration, video1.Dimensions.FPS)
		fmt.Printf("   - URL: %s\n", video1.URL)
	}

	// Example 2: Generate a peaceful landscape timelapse
	fmt.Println("\n🌅 Example 2: Generating Peaceful Landscape Timelapse")
	fmt.Println("====================================================")

	videoReq2 := &framework.VideoRequest{
		Prompt:         "Serene mountain landscape at sunrise, clouds moving slowly, golden hour lighting",
		Width:          1280,
		Height:         720,
		Duration:       10.0, // 10 seconds
		FPS:            24,
		MotionStrength: 0.3,  // Subtle motion for peaceful scene
		Quality:        "standard",
		Format:         "mp4",
		Style:          "photorealistic",
	}

	video2, err := assetGen.GenerateVideo(ctx, videoReq2)
	if err != nil {
		fmt.Printf("⚠️  Video generation failed (expected with placeholder API key): %v\n", err)
	} else {
		fmt.Printf("✅ Generated landscape timelapse:\n")
		fmt.Printf("   - ID: %s\n", video2.ID)
		fmt.Printf("   - Dimensions: %dx%d\n", video2.Dimensions.Width, video2.Dimensions.Height)
		fmt.Printf("   - Duration: %.1fs at %d FPS\n", video2.Dimensions.Duration, video2.Dimensions.FPS)
		fmt.Printf("   - URL: %s\n", video2.URL)
	}

	// Example 3: Generate a magical spell effect
	fmt.Println("\n✨ Example 3: Generating Magical Spell Effect")
	fmt.Println("===========================================")

	videoReq3 := &framework.VideoRequest{
		Prompt:         "Magical lightning spell casting, purple energy swirling, mystical runes glowing",
		Width:          512,
		Height:         512,
		Duration:       5.0,  // Short 5-second spell
		FPS:            60,   // High FPS for smooth effects
		MotionStrength: 0.9,  // Very high motion for magical effects
		Quality:        "high",
		Format:         "gif", // GIF format for web use
		Style:          "fantasy",
	}

	video3, err := assetGen.GenerateVideo(ctx, videoReq3)
	if err != nil {
		fmt.Printf("⚠️  Video generation failed (expected with placeholder API key): %v\n", err)
	} else {
		fmt.Printf("✅ Generated magical spell effect:\n")
		fmt.Printf("   - ID: %s\n", video3.ID)
		fmt.Printf("   - Format: %s\n", video3.Format)
		fmt.Printf("   - Dimensions: %dx%d\n", video3.Dimensions.Width, video3.Dimensions.Height)
		fmt.Printf("   - Duration: %.1fs at %d FPS\n", video3.Dimensions.Duration, video3.Dimensions.FPS)
		fmt.Printf("   - URL: %s\n", video3.URL)
	}

	fmt.Println("\n🎉 Video generation examples completed!")
	fmt.Println("💡 To use with real API key, replace 'your_theta_api_key_here' with actual Theta EdgeCloud API key")
}
