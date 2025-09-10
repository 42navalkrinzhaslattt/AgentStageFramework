package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	loadDotEnv() // ensure .env loaded
	apiKey := getenvFirst([]string{"THETA_API_KEY","THETA_KEY"})
	if apiKey == "" { fmt.Println("Warning: THETA_API_KEY (or THETA_KEY) not set; limited AI features.") }
	// Initialize the game
	sim, err := NewPresidentSim(apiKey)
	if err != nil { fmt.Println("Failed to initialize game:", err); return }
	defer sim.Close()
	orchestrator := NewGameOrchestrator(sim)
	if len(os.Args) > 1 && os.Args[1] == "web" {
		port := "8080"; if len(os.Args) > 2 { port = os.Args[2] }
		server := NewWebServer(orchestrator, port); log.Fatal(server.Start()); return }
	runTerminalMode(orchestrator)
}

func runTerminalMode(orchestrator *GameOrchestrator) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🏛️  PRESIDENTIAL SIMULATOR - CHAT MODE")
	fmt.Println("=====================================")
	fmt.Println("You are the President of the United States.")
	fmt.Println("Over 5 critical turns, you'll face major decisions.")
	fmt.Println("Each turn: Event → 3 Advisor Opinions → Your Choice → Consequences")
	fmt.Println()

	// Game loop for 5 turns
	for !orchestrator.IsGameComplete() {
		ctx := context.Background()
		
		fmt.Printf("\n" + strings.Repeat("=", 60))
		fmt.Printf("\n🏛️  TURN %d of %d", orchestrator.sim.state.Turn, orchestrator.sim.state.MaxTurns)
		fmt.Printf("\n%s\n", strings.Repeat("=", 60)) // ensure a newline after the separator
		
		// Display current metrics
		displayMetrics(orchestrator.sim.state.Metrics)
		
		// Start new turn - generate event and get advisor advice
		fmt.Println("\n📧 [BREAKING NEWS] New situation requires your immediate attention...")
		
		turnResult, err := orchestrator.StartNewTurn(ctx)
		if err != nil {
			fmt.Printf("Error starting turn: %v\n", err)
			continue
		}

		// Display the event
		fmt.Printf("\n🚨 %s\n", turnResult.Event.Title)
		fmt.Printf("📝 %s\n", turnResult.Event.Description)
		
		// Display advisor responses (no banner line)
		for i, advisor := range turnResult.Advisors {
			fmt.Printf("\n%d. 👤 %s:\n", i+1, advisor.AdvisorName)
			fmt.Printf("   💭 %s\n", advisor.Advice)
		}
		fmt.Print("\nYour strategic response (free-form): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" { fmt.Println("Response required."); continue }
		// Process using reasoning only
		if err := orchestrator.ProcessPlayerChoice(ctx, turnResult, -1, input); err != nil { fmt.Printf("Error processing: %v\n", err); continue }

		// Display results
		fmt.Printf("\n🎯 You chose: %s\n", turnResult.Choice.Option)
		fmt.Printf("📊 Evaluation: %s\n", turnResult.Evaluation)
		
		// Display impact
		fmt.Println("\n📈 Impact on metrics:")
		displayImpact(turnResult.Impact)
	}

	// Game complete
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🏁 PRESIDENCY COMPLETED!")
	fmt.Println(strings.Repeat("=", 60))
	
	finalScore := calculateFinalScore(orchestrator.sim.state.Metrics)
	fmt.Printf("📊 Final Score: %.1f/100\n", finalScore)
	
	if finalScore > 70 {
		fmt.Println("🎉 Excellent presidency! You've led the nation with distinction.")
	} else if finalScore > 50 {
		fmt.Println("👍 Good presidency. You handled most challenges well.")
	} else if finalScore > 30 {
		fmt.Println("😐 Mixed results. Your presidency had both successes and failures.")
	} else {
		fmt.Println("😬 Challenging presidency. The nation faced significant difficulties.")
	}
}

func displayMetrics(metrics WorldMetrics) {
	fmt.Printf("📊 Current Metrics:\n")
	fmt.Printf("   🏢 Economy: %.1f/100\n", metrics.Economy)
	fmt.Printf("   🛡️  Security: %.1f/100\n", metrics.Security)
	fmt.Printf("   🤝 Diplomacy: %.1f/100\n", metrics.Diplomacy)
	fmt.Printf("   🌱 Environment: %.1f/100\n", metrics.Environment)
	fmt.Printf("   📊 Approval: %.1f/100\n", metrics.Approval)
	fmt.Printf("   ⚖️  Stability: %.1f/100\n", metrics.Stability)
}

func displayImpact(impact WorldMetrics) {
	fmt.Printf("   🏢 Economy: %+.1f\n", impact.Economy)
	fmt.Printf("   🛡️  Security: %+.1f\n", impact.Security)
	fmt.Printf("   🤝 Diplomacy: %+.1f\n", impact.Diplomacy)
	fmt.Printf("   🌱 Environment: %+.1f\n", impact.Environment)
	fmt.Printf("   📊 Approval: %+.1f\n", impact.Approval)
	fmt.Printf("   ⚖️  Stability: %+.1f\n", impact.Stability)
}

func calculateFinalScore(metrics WorldMetrics) float64 {
	return (metrics.Economy + metrics.Security + metrics.Diplomacy + 
		   metrics.Environment + metrics.Approval + metrics.Stability) / 6.0
}
