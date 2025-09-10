package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
	"strings"
	"os"
	imgc "presidential-simulator/internal/ondemand_image_client"
)

// WebServer handles HTTP requests for the Presidential Simulator
type WebServer struct {
	orchestrator *GameOrchestrator
	port         string
}

// ChatMessage is a UI-friendly message item for the client feed
type ChatMessage struct {
	Name           string `json:"name"`
	Text           string `json:"text"`
	Title          string `json:"title"`
	TitleColor     string `json:"titleColor"`
	Time           string `json:"time"`
	ProfilePicture string `json:"profilePicture"`
}

// GameStateResponse represents the current game state for the frontend
type GameStateResponse struct {
	Turn        int             `json:"turn"`
	MaxTurns    int             `json:"maxTurns"`
	Metrics     WorldMetrics    `json:"metrics"`
	IsComplete  bool            `json:"isComplete"`
	CurrentTurn *TurnResult     `json:"currentTurn,omitempty"`
	History     []TurnResult    `json:"history"`
	Stats       AIUsageStats    `json:"stats"`
}

// NewRoundResponse is returned by the NewRound endpoint
type NewRoundResponse struct {
	GameOver   bool          `json:"gameOver"`
	Turn       int           `json:"turn"`
	MaxTurns   int           `json:"maxTurns"`
	TurnResult *TurnResult   `json:"turnResult,omitempty"`
	Metrics    *WorldMetrics `json:"metrics,omitempty"`
	Newspaper  string        `json:"newspaper,omitempty"`
	Stats      AIUsageStats  `json:"stats"`
	Messages   []ChatMessage `json:"messages,omitempty"`
}

// EvaluateResponse is returned by the EvaluatePlayersChoice endpoint
type EvaluateResponse struct {
	Evaluation string        `json:"evaluation"`
	Impact     WorldMetrics  `json:"impact"`
	Metrics    WorldMetrics  `json:"metrics"`
	IsComplete bool          `json:"isComplete"`
	Turn       int           `json:"turn"`
	MaxTurns   int           `json:"maxTurns"`
	Stats      AIUsageStats  `json:"stats"`
	Messages   []ChatMessage `json:"messages,omitempty"`
}

// NewWebServer creates a new web server instance
func NewWebServer(orchestrator *GameOrchestrator, port string) *WebServer {
	return &WebServer{
		orchestrator: orchestrator,
		port:         port,
	}
}

func (ws *WebServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// Start starts the web server
func (ws *WebServer) Start() error {
	// Serve static files
	http.HandleFunc("/", ws.serveStaticFile)

	// Existing API endpoints (kept for compatibility)
	http.HandleFunc("/api/start", ws.corsMiddleware(ws.handleStart))
	http.HandleFunc("/api/state", ws.corsMiddleware(ws.handleGetState))
	http.HandleFunc("/api/new-turn", ws.corsMiddleware(ws.handleNewTurn))
	http.HandleFunc("/api/choice", ws.corsMiddleware(ws.handlePlayerChoice))

	// New requested endpoints
	http.HandleFunc("/api/new-round", ws.corsMiddleware(ws.handleNewRound))
	http.HandleFunc("/api/evaluate-choice", ws.corsMiddleware(ws.handleEvaluateChoice))
	// Stats-only endpoint
	http.HandleFunc("/api/stats", ws.corsMiddleware(ws.handleStats))
	// New: on-demand image generation for current event
	http.HandleFunc("/api/generate-image", ws.corsMiddleware(ws.handleGenerateImage))

	log.Printf("🌐 Presidential Simulator server starting on http://localhost:%s", ws.port)
	return http.ListenAndServe(":"+ws.port, nil)
}

// serveStaticFile serves the HTML interface
func (ws *WebServer) serveStaticFile(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Try game/chat.html relative to executable and project root
		candidates := []string{
			"game/chat.html",
			"./game/chat.html",
			"chat.html",
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				http.ServeFile(w, r, p)
				return
			}
		}
		// fallback 404
		http.Error(w, "chat.html not found", http.StatusNotFound)
		return
	}
	// allow direct access to assets under game/
	if strings.HasPrefix(r.URL.Path, "/game/") {
		fp := "." + r.URL.Path
		if _, err := os.Stat(fp); err == nil {
			http.ServeFile(w, r, fp)
			return
		}
	}
	http.NotFound(w, r)
}

// handleStart initializes a new game
func (ws *WebServer) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := loadGameConfig()
	// Reset game state
	ws.orchestrator.sim.state.Turn = 1
	ws.orchestrator.sim.state.History = []TurnResult{}
	ws.orchestrator.sim.state.CurrentTurn = nil
	ws.orchestrator.sim.state.Stats = AIUsageStats{}
	minV, maxV := cfg.MetricMin, cfg.MetricMax
	randVal := func() float64 { return float64(minV + rand.Intn(maxV-minV+1)) }
	ws.orchestrator.sim.state.Metrics = WorldMetrics{
		Economy:     randVal(), // Random within configured range
		Security:    randVal(),
		Diplomacy:   randVal(),
		Environment: randVal(),
		Approval:    randVal(),
		Stability:   randVal(),
	}
	ws.orchestrator.sim.state.MaxTurns = cfg.MaxTurns

	response := GameStateResponse{
		Turn:       ws.orchestrator.sim.state.Turn,
		MaxTurns:   ws.orchestrator.sim.state.MaxTurns,
		Metrics:    ws.orchestrator.sim.state.Metrics,
		IsComplete: false,
		History:    ws.orchestrator.sim.state.History,
		Stats:      ws.orchestrator.sim.state.Stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetState returns current game state
func (ws *WebServer) handleGetState(w http.ResponseWriter, r *http.Request) {
	response := GameStateResponse{
		Turn:       ws.orchestrator.sim.state.Turn,
		MaxTurns:   ws.orchestrator.sim.state.MaxTurns,
		Metrics:    ws.orchestrator.sim.state.Metrics,
		IsComplete: ws.orchestrator.IsGameComplete(),
		CurrentTurn: ws.orchestrator.sim.state.CurrentTurn,
		History:    ws.orchestrator.sim.state.History,
		Stats:      ws.orchestrator.sim.state.Stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleNewTurn generates a new crisis and advisor responses (legacy)
func (ws *WebServer) handleNewTurn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if ws.orchestrator.IsGameComplete() {
		http.Error(w, "game complete", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate new turn using the orchestrator
	turnResult, err := ws.orchestrator.StartNewTurn(ctx)
	if err != nil {
		log.Printf("Error starting new turn: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate new turn: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(turnResult)
}

// handlePlayerChoice processes the player's decision (legacy)
func (ws *WebServer) handlePlayerChoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ChoiceIndex *int    `json:"choiceIndex"`
		Reasoning   string `json:"reasoning"`
		Choice      string `json:"choice"`
		Turn        string `json:"turn"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Determine which turn to apply to (use current active turn)
	turnResult := ws.orchestrator.sim.state.CurrentTurn
	if turnResult == nil {
		http.Error(w, "no active turn", http.StatusBadRequest)
		return
	}

	// Derive choice index: if explicit, use; else match against options; else 0
	choiceIndex := 0
	if request.ChoiceIndex != nil {
		choiceIndex = *request.ChoiceIndex
	} else if request.Choice != "" {
		for i, opt := range turnResult.Event.Options {
			if opt == request.Choice {
				choiceIndex = i
				break
			}
		}
	} else if request.Choice != "" {
		// attempt fuzzy contains match
		lc := strings.ToLower(request.Choice)
		for i, opt := range turnResult.Event.Options {
			if strings.Contains(lc, strings.ToLower(opt)) {
				choiceIndex = i
				break
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ws.orchestrator.ProcessPlayerChoice(ctx, turnResult, choiceIndex, request.Reasoning); err != nil {
		log.Printf("Error processing player choice: %v", err)
		http.Error(w, fmt.Sprintf("Failed to process choice: %v", err), http.StatusInternalServerError)
		return
	}

	response := GameStateResponse{
		Turn:        ws.orchestrator.sim.state.Turn,
		MaxTurns:    ws.orchestrator.sim.state.MaxTurns,
		Metrics:     ws.orchestrator.sim.state.Metrics,
		IsComplete:  ws.orchestrator.IsGameComplete(),
		CurrentTurn: nil,
		History:     ws.orchestrator.sim.state.History,
		Stats:       ws.orchestrator.sim.state.Stats,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func colorForCategory(cat string) string {
	switch strings.ToLower(cat) {
	case "environment", "climate":
		return "#2E8B57"
	case "security", "military":
		return "#B22222"
	case "economy":
		return "#DAA520"
	case "diplomacy", "geopolitics":
		return "#4682B4"
	case "technology":
		return "#7B68EE"
	case "public_health":
		return "#008080"
	case "civil_rights":
		return "#6A5ACD"
	case "immigration":
		return "#A0522D"
	case "social_safety_net":
		return "#FF8C00"
	case "gun_policy":
		return "#8B0000"
	case "judicial_appointments":
		return "#4B0082"
	default:
		return "#333333"
	}
}

func colorForSpecialty(spec string) string {
	switch strings.ToLower(spec) {
	case "environment":
		return "#2E8B57"
	case "security", "military":
		return "#B22222"
	case "economy":
		return "#DAA520"
	case "diplomacy":
		return "#4682B4"
	case "tech", "technology":
		return "#7B68EE"
	case "social":
		return "#FF8C00"
	case "domestic":
		return "#5F9EA0"
	default:
		return "#1E90FF"
	}
}

func avatarURL(slug string) string {
	return fmt.Sprintf("https://robohash.org/%s.png?size=64x64&set=set3", slug)
}

// handleNewRound returns event + advisors or newspaper if game is over
func (ws *WebServer) handleNewRound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// If already complete, return newspaper now
	if ws.orchestrator.IsGameComplete() {
		paper := buildEndgameNewspaper(ws.orchestrator.sim.state)
		resp := NewRoundResponse{
			GameOver:  true,
			Turn:      ws.orchestrator.sim.state.Turn,
			MaxTurns:  ws.orchestrator.sim.state.MaxTurns,
			Metrics:   &ws.orchestrator.sim.state.Metrics,
			Newspaper: paper,
			Stats:     ws.orchestrator.sim.state.Stats,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	turnResult, err := ws.orchestrator.StartNewTurn(ctx)
	if err != nil {
		// If exceeded rounds, return newspaper instead of error
		paper := buildEndgameNewspaper(ws.orchestrator.sim.state)
		resp := NewRoundResponse{
			GameOver:  true,
			Turn:      ws.orchestrator.sim.state.Turn,
			MaxTurns:  ws.orchestrator.sim.state.MaxTurns,
			Metrics:   &ws.orchestrator.sim.state.Metrics,
			Newspaper: paper,
			Stats:     ws.orchestrator.sim.state.Stats,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Build message list: 1) Event message 2) Advisor messages
	msgs := make([]ChatMessage, 0, 1+len(turnResult.Advisors))
	now := time.Now().UTC().Format(time.RFC3339)
	evt := turnResult.Event
	msgs = append(msgs, ChatMessage{
		Name:           "News Desk",
		Text:           evt.Description,
		Title:          evt.Title,
		TitleColor:     colorForCategory(evt.Category),
		Time:           now,
		ProfilePicture: avatarURL("news"),
	})
	for _, a := range turnResult.Advisors {
		msgs = append(msgs, ChatMessage{
			Name:           a.AdvisorName,
			Text:           a.Advice,
			Title:          a.Title,
			TitleColor:     colorForSpecialty(strings.ToLower(findAdvisorSpecialty(ws.orchestrator.sim.state.Advisors, a.AdvisorID))),
			Time:           now,
			ProfilePicture: avatarURL(a.AdvisorID),
		})
	}

	resp := NewRoundResponse{
		GameOver:   false,
		Turn:       ws.orchestrator.sim.state.Turn,
		MaxTurns:   ws.orchestrator.sim.state.MaxTurns,
		TurnResult: turnResult,
		Stats:      ws.orchestrator.sim.state.Stats,
		Messages:   msgs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func findAdvisorSpecialty(list []Advisor, id string) string {
	for _, a := range list {
		if a.ID == id { return a.Specialty }
	}
	return ""
}

// handleEvaluateChoice evaluates player's decision and returns evaluation + impact
func (ws *WebServer) handleEvaluateChoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Reasoning   string `json:"reasoning"`
		ChoiceIndex *int   `json:"choiceIndex"`
		Choice      string `json:"choice"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	turnResult := ws.orchestrator.sim.state.CurrentTurn
	if turnResult == nil {
		http.Error(w, "no active turn", http.StatusBadRequest)
		return
	}
	choiceIndex := 0
	if request.ChoiceIndex != nil {
		choiceIndex = *request.ChoiceIndex
	} else if request.Choice != "" {
		for i, opt := range turnResult.Event.Options {
			if strings.EqualFold(opt, request.Choice) { choiceIndex = i; break }
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	if err := ws.orchestrator.ProcessPlayerChoice(ctx, turnResult, choiceIndex, request.Reasoning); err != nil {
		log.Printf("Error processing player choice: %v", err)
		http.Error(w, fmt.Sprintf("Failed to process choice: %v", err), http.StatusInternalServerError)
		return
	}

	// Last history item has evaluation and impact
	hist := ws.orchestrator.sim.state.History
	last := hist[len(hist)-1]

	now := time.Now().UTC().Format(time.RFC3339)
	msgs := []ChatMessage{
		{
			Name:           "Director",
			Text:           last.Evaluation,
			Title:          "Evaluation",
			TitleColor:     "#333333",
			Time:           now,
			ProfilePicture: avatarURL("director"),
		},
	}

	resp := EvaluateResponse{
		Evaluation: last.Evaluation,
		Impact:     last.Impact,
		Metrics:    ws.orchestrator.sim.state.Metrics,
		IsComplete: ws.orchestrator.IsGameComplete(),
		Turn:       ws.orchestrator.sim.state.Turn,
		MaxTurns:   ws.orchestrator.sim.state.MaxTurns,
		Stats:      ws.orchestrator.sim.state.Stats,
		Messages:   msgs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleStats returns only the AI usage stats
func (ws *WebServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ws.orchestrator.sim.state.Stats)
}

// handleGenerateImage generates a BBC/AP style image for the current event and returns the URL
func (ws *WebServer) handleGenerateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	turn := ws.orchestrator.sim.state.CurrentTurn
	if turn == nil {
		http.Error(w, "no active turn", http.StatusBadRequest)
		return
	}
	// Optional body: { width?: number, height?: number }
	var req struct{ Width, Height int }
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Width <= 0 { req.Width = 800 }
	if req.Height <= 0 { req.Height = 450 }

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	prompt := buildBBCPhotoPrompt(&turn.Event)
	client := imgc.New()
	url, err := client.Generate(ctx, prompt, req.Width, req.Height)
	if err != nil {
		http.Error(w, fmt.Sprintf("image generation failed: %v", err), http.StatusBadGateway)
		return
	}
	// Attach to current event
	turn.Event.ImageURL = url

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"eventId": turn.Event.ID,
		"imageUrl": url,
	})
}

// buildEndgameNewspaper creates a simple newspaper-style summary of the run
func buildEndgameNewspaper(state *GameState) string {
	var b strings.Builder
	fmt.Fprintf(&b, "🗞️ NATIONAL LEDGER — PRESIDENCY COMPLETE\n")
	fmt.Fprintf(&b, "=======================================\n\n")
	fmt.Fprintf(&b, "Final Metrics — Economy %.1f | Security %.1f | Diplomacy %.1f | Environment %.1f | Approval %.1f | Stability %.1f\n\n",
		state.Metrics.Economy, state.Metrics.Security, state.Metrics.Diplomacy, state.Metrics.Environment, state.Metrics.Approval, state.Metrics.Stability)
	for _, t := range state.History {
		fmt.Fprintf(&b, "TURN %d — %s (%s, sev %d/10)\n", t.Turn, t.Event.Title, t.Event.Category, t.Event.Severity)
		// Print first line of evaluation
		line := t.Evaluation
		if idx := strings.IndexRune(line, '\n'); idx >= 0 { line = line[:idx] }
		if len(line) > 180 { line = line[:180] + "..." }
		fmt.Fprintf(&b, "Outcome: %s\n\n", line)
	}
	b.WriteString("— End of Term —\n")
	return b.String()
}
