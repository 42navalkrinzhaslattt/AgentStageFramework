package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"sync"
	fw "github.com/emergent-world-engine/backend/pkg/framework"
)

// GameOrchestrator manages the 5-turn chat game flow
type GameOrchestrator struct {
	sim *PresidentSim
}

func NewGameOrchestrator(sim *PresidentSim) *GameOrchestrator { return &GameOrchestrator{sim: sim} }

// StartNewTurn begins a new turn in the game
func (g *GameOrchestrator) StartNewTurn(ctx context.Context) (*TurnResult, error) {
	if g.sim.state.Turn > g.sim.state.MaxTurns {
		return nil, fmt.Errorf("game completed after %d turns", g.sim.state.MaxTurns)
	}

	// Generate random event (later could integrate Narrative quests or Director generated events)
	event, err := g.sim.GenerateTurnEvent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate event: %w", err)
	}

	// Select 3 random advisors
	selectedAdvisors := g.selectRandomAdvisors(3)

	// Get advice from each selected advisor in parallel
	advisorResponses := make([]AdvisorResponse, 0, len(selectedAdvisors))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, advisor := range selectedAdvisors {
		wg.Add(1)
		ad := advisor
		go func() {
			defer wg.Done()
			// per-advisor timeout (extended)
			cctx, cancel := context.WithTimeout(ctx, 35*time.Second)
			defer cancel()
			resp, err := g.getAdvisorAdviceStream(cctx, ad, *event)
			if err != nil {
				log.Printf("[ADVISOR] %s error: %v (using fallback)", ad.Name, err)
				fb := synthFallbackAdvice(ad)
				resp = AdvisorResponse{AdvisorID: ad.ID, AdvisorName: ad.Name, Title: ad.Title, Advice: fb, Recommendation: 0}
			}
			mu.Lock(); advisorResponses = append(advisorResponses, resp); mu.Unlock()
		}()
	}
	wg.Wait()

	turnResult := &TurnResult{
		Turn:     g.sim.state.Turn,
		Event:    *event,
		Advisors: advisorResponses,
	}
	g.sim.state.CurrentTurn = turnResult
	return turnResult, nil
}

// ProcessPlayerChoice handles the player's decision and evaluates the outcome
func (g *GameOrchestrator) ProcessPlayerChoice(ctx context.Context, turnResult *TurnResult, choiceIndex int, reasoning string) error {
	// Ignore numeric choice; treat reasoning as the action narrative
	turnResult.Choice = PlayerChoice{EventID: turnResult.Event.ID, OptionIndex: -1, Option: "policy_response", Reasoning: reasoning}

	// Evaluate via Director using reasoning text
	turnResult.Event.Options = nil // remove options for downstream display
	evaluation, impact, err := g.evaluateChoice(ctx, turnResult)
	if err != nil {
		return fmt.Errorf("failed to evaluate reasoning: %w", err)
	}

	turnResult.Evaluation = evaluation
	turnResult.Impact = impact

	// Update world metrics
	g.updateWorldMetrics(impact)

	// Add to history and advance turn
	g.sim.state.History = append(g.sim.state.History, *turnResult)
	g.sim.state.Turn++
	g.sim.state.LastUpdated = time.Now()
	g.sim.state.CurrentTurn = nil
	return nil
}

// selectRandomAdvisors picks 3 random advisors from the 8 available
func (g *GameOrchestrator) selectRandomAdvisors(count int) []Advisor {
	advisors := make([]Advisor, len(g.sim.state.Advisors))
	copy(advisors, g.sim.state.Advisors)

	// Shuffle advisors
	rand.Shuffle(len(advisors), func(i, j int) {
		advisors[i], advisors[j] = advisors[j], advisors[i]
	})

	// Return first 'count' advisors
	if count > len(advisors) {
		count = len(advisors)
	}
	return advisors[:count]
}

var (
	jsonCandidateRE = regexp.MustCompile(`\{[\s\S]*?\}`)
)

// extractAdvisorOpinion attempts layered extraction strategies
func extractAdvisorOpinion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	// 1) Prefer valid JSON with advisor_opinion (scan all JSON candidates, prefer the last)
	if strings.Contains(raw, "advisor_opinion") {
		locs := jsonCandidateRE.FindAllStringIndex(raw, -1)
		for i := len(locs) - 1; i >= 0; i-- { // prefer the last JSON block
			frag := raw[locs[i][0]:locs[i][1]]
			var m map[string]any
			if json.Unmarshal([]byte(frag), &m) == nil {
				if v, ok := m["advisor_opinion"].(string); ok {
					return sanitizeOpinion(v)
				}
			}
		}
		// If advisor_opinion key is present but JSON is malformed, try regex extraction then fall through
		if mm := regexp.MustCompile(`(?i)\"advisor_opinion\"\s*:\s*\"([^\"]+)\"`).FindStringSubmatch(raw); len(mm) > 1 {
			return sanitizeOpinion(mm[1])
		}
		// else continue to non-JSON strategies below
	}
	// 2) Look for a line like Advisor Opinion: or similar
	lines := strings.Split(raw, "\n")
	for _, ln := range lines {
		low := strings.ToLower(ln)
		if strings.Contains(low, "advisor opinion") || strings.Contains(low, "final advisory") {
			return sanitizeOpinion(afterColon(ln))
		}
	}
	// 3) Take first non-meta paragraph
	paras := splitParas(raw)
	for _, p := range paras {
		p2 := strings.TrimSpace(p)
		if p2 == "" {
			continue
		}
		if looksMeta(p2) {
			continue
		}
		return sanitizeOpinion(p2)
	}
	return ""
}

func afterColon(s string) string {
	if idx := strings.IndexRune(s, ':'); idx >= 0 {
		return strings.TrimSpace(s[idx+1:])
	}
	return strings.TrimSpace(s)
}

func looksMeta(s string) bool {
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "i am ") || strings.HasPrefix(low, "i'm ") {
		return true
	}
	if strings.Contains(low, "thinking") || strings.Contains(low, "reasoning") {
		return true
	}
	if strings.Contains(low, "let me") {
		return true
	}
	return false
}

func splitParas(s string) []string { return strings.Split(s, "\n\n") }

func sanitizeOpinion(s string) string {
	// Remove leading quotes/backticks
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "`")
	s = strings.Trim(s, "\"")
	// Collapse whitespace
	fs := strings.Fields(s)
	if len(fs) == 0 {
		return ""
	}
	out := strings.Join(fs, " ")
	// Limit to ~3 sentences
	sents := regexp.MustCompile(`([^.?!]*[.?!])`).FindAllString(out, -1)
	if len(sents) > 0 {
		if len(sents) > 3 {
			sents = sents[:3]
		}
		return strings.TrimSpace(strings.Join(sents, " "))
	}
	return out
}

// Streaming advisor dialogue
func (g *GameOrchestrator) getAdvisorAdviceStream(ctx context.Context, advisor Advisor, event GameEvent) (AdvisorResponse, error) {
	npc := g.sim.advisors[advisor.ID]
	if npc == nil {
		log.Printf("[ADVISOR] NPC not found for %s (%s); using fallback", advisor.Name, advisor.ID)
		fb := synthFallbackAdvice(advisor)
		return AdvisorResponse{AdvisorID: advisor.ID, AdvisorName: advisor.Name, Title: advisor.Title, Advice: fb, Recommendation: 0}, nil
	}

	buildPrompt := func() string {
		persona := fmt.Sprintf("%s (%s) specialty=%s traits=%s", advisor.Name, advisor.Title, advisor.Specialty, advisor.Personality)
		return fmt.Sprintf(`ROLE: Senior presidential advisor.
Persona: %s
Event: %s
Category: %s Severity: %d/10
Description: %s
Task: Provide one concise, actionable advisory opinion (policy recommendation or strategic action).
Constraints: 2-4 sentences. No internal reasoning, no preamble, no self-reference.
Output ONLY valid JSON: {"advisor_opinion":"<your concise advisory>"}
If unsure, still give best judgment.`,
			persona, event.Title, event.Category, event.Severity, event.Description)
	}

	var buf strings.Builder
	onChunk := func(tok string) { buf.WriteString(tok) }
	prompt := buildPrompt()
	resp, err := npc.GenerateDialogueStream(ctx, &fw.DialogueRequest{PlayerMessage: prompt}, onChunk)
	if err != nil {
		log.Printf("[ADVISOR] %s stream error: %v", advisor.Name, err)
		fb := synthFallbackAdvice(advisor)
		return AdvisorResponse{AdvisorID: advisor.ID, AdvisorName: advisor.Name, Title: advisor.Title, Advice: fb, Recommendation: 0}, nil
	}
	raw := strings.TrimSpace(resp.Message)
	if raw == "" { raw = strings.TrimSpace(buf.String()) }
	final := extractAdvisorOpinion(raw)
	if final == "" {
		final = synthFallbackAdvice(advisor)
		if final == "" { return AdvisorResponse{}, errors.New("unable to derive advisor opinion") }
	}
	return AdvisorResponse{AdvisorID: advisor.ID, AdvisorName: advisor.Name, Title: advisor.Title, Advice: final, Recommendation: 0}, nil
}

// fallback evaluation
func (g *GameOrchestrator) randomEval(t *TurnResult) string {
	return fmt.Sprintf("Fallback evaluation for '%s' in %s context.", t.Choice.Option, t.Event.Category)
}
func (g *GameOrchestrator) randomImpact() WorldMetrics {
	return WorldMetrics{
		Economy:     (rand.Float64() - 0.5) * 20,
		Security:    (rand.Float64() - 0.5) * 20,
		Diplomacy:   (rand.Float64() - 0.5) * 20,
		Environment: (rand.Float64() - 0.5) * 20,
		Approval:    (rand.Float64() - 0.5) * 10,
		Stability:   (rand.Float64() - 0.5) * 10,
	}
}

func mapDirectorDecisionToMetrics(decision *fw.DirectorDecision, category string) WorldMetrics {
	text := decision.Reasoning
	// Find last JSON object (metrics expected at end)
	start := strings.LastIndex(text, "{")
	end := strings.LastIndex(text, "}")
	if start >=0 && end>start {
		fragment := text[start:end+1]
		var outer struct { Metrics map[string]int `json:"metrics"` }
		if json.Unmarshal([]byte(fragment), &outer) == nil && len(outer.Metrics) > 0 {
			m := outer.Metrics
			return WorldMetrics{
				Economy: float64(m["economy"]),
				Security: float64(m["security"]),
				Diplomacy: float64(m["diplomacy"]),
				Environment: float64(m["environment"]),
				Approval: float64(m["approval"]),
				Stability: float64(m["stability"]),
			}
		}
	}
	// Fallback heuristic with cross-domain spillovers (≥3 non-zero deltas)
	switch category {
	case "economy":
		return WorldMetrics{Economy: 2, Approval: 1, Stability: 1}
	case "security", "military":
		return WorldMetrics{Security: 3, Stability: 2, Diplomacy: -1}
	case "diplomacy", "geopolitics":
		return WorldMetrics{Diplomacy: 2, Stability: 1, Approval: 1}
	case "environment", "climate":
		return WorldMetrics{Environment: 3, Approval: 1, Economy: -1}
	case "public_health":
		return WorldMetrics{Approval: 2, Stability: 1, Economy: -1}
	case "technology":
		return WorldMetrics{Economy: 1, Stability: 1, Security: 1}
	// New categories
	case "civil_rights":
		return WorldMetrics{Approval: 2, Stability: 1, Diplomacy: 1}
	case "military_intervention":
		return WorldMetrics{Security: 3, Stability: 1, Diplomacy: -2, Economy: -1}
	case "immigration":
		return WorldMetrics{Economy: 1, Security: 1, Diplomacy: 1}
	case "social_safety_net":
		return WorldMetrics{Approval: 2, Stability: 1, Economy: -1}
	case "gun_policy":
		return WorldMetrics{Security: 2, Stability: 1, Approval: 1}
	case "judicial_appointments":
		return WorldMetrics{Stability: 2, Approval: 1, Diplomacy: 1}
	}
	return WorldMetrics{}
}

func (g *GameOrchestrator) updateWorldMetrics(impact WorldMetrics) {
	g.sim.state.Metrics.Economy = clamp(g.sim.state.Metrics.Economy+impact.Economy, -100, 100)
	g.sim.state.Metrics.Security = clamp(g.sim.state.Metrics.Security+impact.Security, -100, 100)
	g.sim.state.Metrics.Diplomacy = clamp(g.sim.state.Metrics.Diplomacy+impact.Diplomacy, -100, 100)
	g.sim.state.Metrics.Environment = clamp(g.sim.state.Metrics.Environment+impact.Environment, -100, 100)
	g.sim.state.Metrics.Approval = clamp(g.sim.state.Metrics.Approval+impact.Approval, -100, 100)
	g.sim.state.Metrics.Stability = clamp(g.sim.state.Metrics.Stability+impact.Stability, -100, 100)
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (g *GameOrchestrator) GetCurrentState() *GameState { return g.sim.state }
func (g *GameOrchestrator) IsGameComplete() bool { return g.sim.state.Turn > g.sim.state.MaxTurns }

// snippet utility
func snippet(s string, n int) string { if len(s) <= n { return s }; return s[:n] + "..." }

// synthFallbackAdvice (restored)
func synthFallbackAdvice(advisor Advisor) string {
	return fmt.Sprintf("%s (%s): Provide a stabilizing course.", advisor.Name, advisor.Title)
}

// Use Director for evaluation (restored)
func (g *GameOrchestrator) evaluateChoice(ctx context.Context, turnResult *TurnResult) (string, WorldMetrics, error) {
	start := time.Now()
	log.Printf("[DIRECTOR] evaluating choice turn=%d option=%q category=%s severity=%d", turnResult.Turn, turnResult.Choice.Option, turnResult.Event.Category, turnResult.Event.Severity)
	de := &fw.GameEvent{Type:"player_choice", PlayerID:"president", Timestamp: time.Now(), Location:"white_house", Action:"decision", Parameters: map[string]interface{}{ "option": turnResult.Choice.Option, "category": turnResult.Event.Category, "severity": turnResult.Event.Severity }}
	decision, err := g.sim.director.ProcessEvent(ctx, de)
	if err != nil { log.Printf("[DIRECTOR] error: %v (fallback)", err); return g.randomEval(turnResult), g.randomImpact(), nil }
	impact := mapDirectorDecisionToMetrics(decision, turnResult.Event.Category)
	log.Printf("[DIRECTOR] success latency=%s impact={eco:%.1f sec:%.1f dip:%.1f env:%.1f appr:%.1f stab:%.1f}", time.Since(start), impact.Economy, impact.Security, impact.Diplomacy, impact.Environment, impact.Approval, impact.Stability)
	// Narrative consequence description for UI
	narr := formatDirectorNarrative(turnResult, impact)
	return narr, impact, nil
}

func formatDirectorNarrative(t *TurnResult, _ WorldMetrics) string {
	// Pure narrative without explicit metric changes; the numbers will be shown elsewhere
	var b strings.Builder
	fmt.Fprintf(&b, "In response to %s (%s, severity %d/10), you moved decisively. ", t.Event.Title, t.Event.Category, t.Event.Severity)
	// Category-informed ripple description (no deltas)
	s := strings.ToLower(t.Event.Category)
	switch s {
	case "economy":
		b.WriteString("Markets brace and messaging shifts toward coordination and relief. ")
	case "security", "military", "military_intervention":
		b.WriteString("Readiness tightens, allies take note, and diplomatic friction simmers. ")
	case "civil_rights":
		b.WriteString("Rule-of-law framing and rights protections shape public trust and coalitions. ")
	case "environment", "climate":
		b.WriteString("Resilience and long-horizon investment narratives dominate the airwaves. ")
	case "technology":
		b.WriteString("Systems hardening and incident command take center stage as confidence rebuilds. ")
	case "public_health":
		b.WriteString("Response cadence steadies nerves and restores a sense of predictability. ")
	case "diplomacy", "geopolitics":
		b.WriteString("Signal discipline and backchannels recalibrate the regional balance. ")
	default:
		b.WriteString("Stakeholders reposition quickly as your stance becomes the new baseline. ")
	}
	b.WriteString("All eyes are on the White House as the next moves are queued up.")
	return strings.TrimSpace(b.String())
}

// rewriteEventDescription converts the raw event description into a viral BREAKING NEWS style post
func (g *GameOrchestrator) rewriteEventDescription(ctx context.Context, event GameEvent) (string, error) {
	id := fmt.Sprintf("news_rewriter_%d", time.Now().UnixNano())
	npc := g.sim.engine.NewNPC(id,
		fw.WithPersonality("Urgent, sensational social media editor who writes concise breaking news threads"),
		fw.WithBackground("Veteran crisis reporter and newsroom copy chief"),
	)
	if npc == nil {
		return "", fmt.Errorf("failed to init news rewriter")
	}
	npc.Config().DialogueModel = fw.ModelLlama70B

	prompt := fmt.Sprintf(
		"Rewrite the following event into a viral BREAKING NEWS social media post with these rules:\n\n"+
		"1) Formatting and Style (CRUCIAL):\n"+
		"- Headline: short, ALL-CAPS, clickbait; include emojis like ⚡️ or 🚨.\n"+
		"- Tone: urgent, sensational, short direct sentences.\n"+
		"- Structure: small thematic paragraphs with bolded emoji-led titles (e.g., 🛑 **Economic Standstill:**).\n"+
		"- Visuals: use emojis; include a single line placeholder for a representative image like ``.\n"+
		"- Hashtags: conclude with relevant, trending-style hashtags.\n\n"+
		"2) Narrative Content (include within the post):\n"+
		"- Name the entity: if an AI provider/tech firm is implicated, call it \"AegisNet\"; otherwise invent a specific plausible entity for this event.\n"+
		"- Show the chaos: vivid real-world consequences (e.g., NYSE halts trading, ATMs fail, trucks stranded).\n"+
		"- Inject the politics: explicitly mention the White House or The President in crisis; add opponents blaming the administration.\n"+
		"- Create a mystery: contrast an official statement (e.g., \"technical glitch\") vs a more alarming rumor/intelligence leak (e.g., \"foreign cyberattack\").\n"+
		"- Focus on the player: end by stating that all eyes are on the President to act next.\n\n"+
		"Event Context:\n"+
		"- Title: %s\n- Category: %s\n- Severity: %d/10\n- Description: %s\n\n"+
		"Output only the finished post. No preambles.",
		event.Title, event.Category, event.Severity, event.Description,
	)

	// Use streaming with timeout for robustness
	cctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var buf strings.Builder
	onChunk := func(tok string) { buf.WriteString(tok) }
	resp, err := npc.GenerateDialogueStream(cctx, &fw.DialogueRequest{PlayerMessage: prompt}, onChunk)
	if err != nil {
		return "", err
	}
	out := strings.TrimSpace(resp.Message)
	if out == "" { out = strings.TrimSpace(buf.String()) }
	if out == "" { return "", fmt.Errorf("empty rewrite") }
	return out, nil
}
