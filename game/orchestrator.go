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
	gemini "presidential-simulator/internal/gemini_client"
	llama "presidential-simulator/internal/llama_client"
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
	// Sanitize for both API and terminal flows
	event.Title = sanitizeEventText(event.Title)
	event.Description = sanitizeEventText(event.Description)

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
	backticksRE   = regexp.MustCompile("`+")
	mdBoldRE      = regexp.MustCompile(`\*\*`)
	hashLineRE    = regexp.MustCompile(`(?m)^\s*#Breaking\b.*$`)
	advisorsHdrRE = regexp.MustCompile(`(?m)^\s*💼\s*Your advisors weigh in:\s*$`)
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

func sanitizeEventText(s string) string {
	if s == "" { return s }
	s = mdBoldRE.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "*", "") // remove any remaining asterisks
	s = backticksRE.ReplaceAllString(s, "")
	s = advisorsHdrRE.ReplaceAllString(s, "")
	s = hashLineRE.ReplaceAllString(s, "")
	// collapse extra blank lines
	lines := []string{}
	prevBlank := false
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimRight(ln, " \t")
		if strings.TrimSpace(ln) == "" {
			if prevBlank { continue }
			prevBlank = true
			lines = append(lines, "")
			continue
		}
		prevBlank = false
		lines = append(lines, ln)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// Streaming advisor dialogue
func (g *GameOrchestrator) getAdvisorAdviceStream(ctx context.Context, advisor Advisor, event GameEvent) (AdvisorResponse, error) {
	// Build prompt for advisor
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

	prompt := buildPrompt()
	usedTheta := false

	// Call Llama chat completions endpoint (non-streaming)
	cctx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	out, err := llama.New().Complete(cctx, prompt)
	if err != nil {
		log.Printf("[ADVISOR] %s llama endpoint error: %v; trying Gemini fallback", advisor.Name, err)
		if adv, gerr := g.advisorOpinionViaGemini(ctx, advisor, event); gerr == nil && adv != "" {
			g.sim.state.Stats.AdvisorGemini++
			return AdvisorResponse{AdvisorID: advisor.ID, AdvisorName: advisor.Name, Title: advisor.Title, Advice: adv, Recommendation: 0}, nil
		}
		fb := synthFallbackAdvice(advisor)
		return AdvisorResponse{AdvisorID: advisor.ID, AdvisorName: advisor.Name, Title: advisor.Title, Advice: fb, Recommendation: 0}, nil
	}
	raw := strings.TrimSpace(out)
	usedTheta = true

	final := extractAdvisorOpinion(raw)
	if looksMetaLike(final) {
		log.Printf("[ADVISOR] %s meta-like advisory rejected: %q", advisor.Name, snippet(final, 120))
		final = ""
	}
	if final == "" {
		if adv, gerr := g.advisorOpinionViaGemini(ctx, advisor, event); gerr == nil && adv != "" {
			g.sim.state.Stats.AdvisorGemini++
			log.Printf("[ADVISOR] %s using Gemini fallback", advisor.Name)
			final = adv
			usedTheta = false
		} else if gerr != nil {
			log.Printf("[ADVISOR] %s Gemini fallback failed detail: %v", advisor.Name, gerr)
		}
	}
	if final == "" {
		final = synthFallbackAdvice(advisor)
		log.Printf("[ADVISOR] %s using hardcoded fallback advisory", advisor.Name)
		if final == "" { return AdvisorResponse{}, errors.New("unable to derive advisor opinion") }
	}
	if usedTheta && final != "" { g.sim.state.Stats.AdvisorTheta++ }
	return AdvisorResponse{AdvisorID: advisor.ID, AdvisorName: advisor.Name, Title: advisor.Title, Advice: final, Recommendation: 0}, nil
}

var badCharsRE = regexp.MustCompile(`[{}\[\]<>()]`)
func looksMetaLike(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" { return false }
	if badCharsRE.MatchString(s) { return true }
	return false
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
	// Avoid parentheses/brackets in fallback to not resemble meta/formatting
	return "Provide a stabilizing course"
}

// Use Director for evaluation (restored)
func (g *GameOrchestrator) evaluateChoice(ctx context.Context, turnResult *TurnResult) (string, WorldMetrics, error) {
	start := time.Now()
	log.Printf("[DIRECTOR] evaluating choice turn=%d option=%q category=%s severity=%d", turnResult.Turn, turnResult.Choice.Option, turnResult.Event.Category, turnResult.Event.Severity)
	de := &fw.GameEvent{Type:"player_choice", PlayerID:"president", Timestamp: time.Now(), Location:"white_house", Action:"decision", Parameters: map[string]interface{}{
		"option": turnResult.Choice.Option,
		"category": turnResult.Event.Category,
		"severity": turnResult.Event.Severity,
		"event_title": turnResult.Event.Title,
		"event_description": turnResult.Event.Description,
		"reasoning": turnResult.Choice.Reasoning,
	}}
	decision, err := g.sim.director.ProcessEvent(ctx, de)
	if err != nil {
		log.Printf("[DIRECTOR] error: %v (trying Gemini fallback)", err)
		analysis, impact, gerr := g.directorMetricsViaGemini(ctx, turnResult)
		if gerr == nil {
			g.sim.state.Stats.DirectorGemini++
			log.Printf("[DIRECTOR] Gemini fallback success latency=%s impact={eco:%.1f sec:%.1f dip:%.1f env:%.1f appr:%.1f stab:%.1f}", time.Since(start), impact.Economy, impact.Security, impact.Diplomacy, impact.Environment, impact.Approval, impact.Stability)
			if strings.TrimSpace(analysis) == "" { analysis = formatDirectorNarrative(turnResult, impact) }
			return analysis, impact, nil
		}
		log.Printf("[DIRECTOR] Gemini fallback failed detail: %v (using random)", gerr)
		return g.randomEval(turnResult), g.randomImpact(), nil
	}

	// Try to parse strict metrics JSON from Director. If absent, use Gemini path (Option 1).
	if impact, ok := parseDirectorMetricsFromReasoning(decision.Reasoning); ok {
		g.sim.state.Stats.DirectorTheta++
		log.Printf("[DIRECTOR] success (with JSON) latency=%s impact={eco:%.1f sec:%.1f dip:%.1f env:%.1f appr:%.1f stab:%.1f}", time.Since(start), impact.Economy, impact.Security, impact.Diplomacy, impact.Environment, impact.Approval, impact.Stability)
		analysis := extractAnalysisText(decision.Reasoning)
		if strings.TrimSpace(analysis) == "" { analysis = formatDirectorNarrative(turnResult, impact) }
		return analysis, impact, nil
	}

	log.Printf("[DIRECTOR] no final metrics JSON found; using Gemini evaluation path")
	analysis2, impact2, gerr2 := g.directorMetricsViaGemini(ctx, turnResult)
	if gerr2 == nil {
		g.sim.state.Stats.DirectorGemini++
		log.Printf("[DIRECTOR] Gemini fallback success latency=%s impact={eco:%.1f sec:%.1f dip:%.1f env:%.1f appr:%.1f stab:%.1f}", time.Since(start), impact2.Economy, impact2.Security, impact2.Diplomacy, impact2.Environment, impact2.Approval, impact2.Stability)
		if strings.TrimSpace(analysis2) == "" { analysis2 = formatDirectorNarrative(turnResult, impact2) }
		return analysis2, impact2, nil
	}
	log.Printf("[DIRECTOR] Gemini evaluation failed detail: %v (using random)", gerr2)
	return g.randomEval(turnResult), g.randomImpact(), nil
}

// Extracts text before the last JSON object in a string.
func extractAnalysisText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" { return "" }
	start := strings.LastIndex(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(s[:start])
	}
	return s
}

func formatDirectorNarrative(t *TurnResult, impact WorldMetrics) string {
	reason := strings.TrimSpace(t.Choice.Reasoning)
	if len(reason) > 220 { reason = reason[:220] + "..." }
	header := fmt.Sprintf("Action Analysis: In response to \"%s\" (%s, severity %d), the administration chose: %s.", t.Event.Title, t.Event.Category, t.Event.Severity, reason)
	parts := make([]string, 0, 6)
	add := func(name string, v float64) {
		if v == 0 { return }
		sign := "+"
		val := v
		if v < 0 { sign = "-"; val = -v }
		parts = append(parts, fmt.Sprintf("%s %s%.0f", name, sign, val))
	}
	add("Economy", impact.Economy)
	add("National Security", impact.Security)
	add("Geopolitical Standing", impact.Diplomacy)
	add("Environment", impact.Environment)
	add("Public Opinion", impact.Approval)
	add("Stability", impact.Stability)
	impactLine := "Metric Impact: " + strings.Join(parts, ", ")
	return strings.TrimSpace(header + " " + impactLine)
}

// parseDirectorMetricsFromReasoning extracts only if a final JSON with metrics exists; no heuristics.
func parseDirectorMetricsFromReasoning(text string) (WorldMetrics, bool) {
	text = strings.TrimSpace(text)
	start := strings.LastIndex(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		fragment := text[start : end+1]
		var outer struct{ Metrics map[string]int `json:"metrics"` }
		if json.Unmarshal([]byte(fragment), &outer) == nil && len(outer.Metrics) > 0 {
			m := outer.Metrics
			return WorldMetrics{
				Economy:     float64(m["economy"]),
				Security:    float64(m["security"]),
				Diplomacy:   float64(m["diplomacy"]),
				Environment: float64(m["environment"]),
				Approval:    float64(m["approval"]),
				Stability:   float64(m["stability"]),
			}, true
		}
	}
	return WorldMetrics{}, false
}

// --- Gemini fallbacks ---

func (g *GameOrchestrator) advisorOpinionViaGemini(ctx context.Context, advisor Advisor, event GameEvent) (string, error) {
	c := gemini.New()
	if c.APIKey == "" {
		return "", errors.New("GOOGLE_AI_API_KEY not set")
	}
	pp := fmt.Sprintf(`You are %s (%s), a senior presidential advisor.
Event: %s
Category: %s (severity %d/10)
Description: %s
Task: Provide one concise, actionable advisory opinion.
Constraints: 2-4 sentences. No internal reasoning, no preamble, no self-reference.
Output ONLY valid JSON exactly like: {"advisor_opinion":"<your concise advisory>"}
No markdown.`, advisor.Name, advisor.Title, event.Title, event.Category, event.Severity, event.Description)
	ctx2, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	out, err := c.GenerateText(ctx2, pp)
	if err != nil { return "", err }
	op := extractAdvisorOpinion(strings.TrimSpace(out))
	if op == "" || looksMetaLike(op) { return "", errors.New("gemini returned invalid advisor_opinion") }
	return op, nil
}

func (g *GameOrchestrator) directorMetricsViaGemini(ctx context.Context, t *TurnResult) (string, WorldMetrics, error) {
	c := gemini.New()
	if c.APIKey == "" {
		return "", WorldMetrics{}, errors.New("GOOGLE_AI_API_KEY not set")
	}
	// Build evaluation prompt per spec, but require the final line JSON to use internal keys
	metricsList := `Public Opinion:\n\nEconomy:\n\nNational Security:\n\nGeopolitical Standing:\n\nTech Sector Confidence:\n\nCivil Liberties:`
	pp := fmt.Sprintf(`Event Evaluation Prompt\nYou are an expert political and economic analyst AI. Your task is to evaluate a player's action in response to a specific event within a presidential simulator game.\n\nAnalyze the provided Event Description and the player's Chosen Action. Based on this analysis, determine the numerical impact on the given Game Metrics. For each metric change, you must provide a brief, clear justification.\n\n1. Event Description\n%s\n\n2. Player's Chosen Action\n%s\n\n3. Game Metrics\n%s\n\n4. Evaluation Task\nInstructions:\n- Step 1: Analyze the Action's Logic and Consequences. Briefly summarize immediate and long-term consequences.\n- Step 2: Determine Metric Changes and Provide Justification. For each game metric, provide a numerical change (e.g., +15, -20, 0) and a one-sentence justification.\n\nExample Output Structure:\nAction Analysis: <2-4 sentences>\n\nMetric Impact:\nPublic Opinion: +10. Justification: <why>.\nEconomy: -5. Justification: <why>.\nNational Security: +20. Justification: <why>.\nGeopolitical Standing: +5. Justification: <why>.\nTech Sector Confidence: -15. Justification: <why>.\nCivil Liberties: -10. Justification: <why>.\n\nCRUCIAL: After your analysis and metric impact lines, output exactly ONE final line containing ONLY a JSON object with integer deltas for: {"metrics":{"economy":E,"security":S,"diplomacy":D,"environment":Env,"approval":A,"stability":St}}. Map as follows: Public Opinion->approval, Economy->economy, National Security->security, Geopolitical Standing->diplomacy, Tech Sector Confidence->stability, Civil Liberties->approval (also subtract half into stability if negative). Use range -20..20. Do NOT include any text or markdown after the JSON.`,
		t.Event.Description,
		t.Choice.Reasoning,
		metricsList,
	)
	ctx2, cancel := context.WithTimeout(ctx, 22*time.Second)
	defer cancel()
	out, err := c.GenerateText(ctx2, pp)
	if err != nil { return "", WorldMetrics{}, err }
	analysis := extractAnalysisText(strings.TrimSpace(out))
	// Reuse existing JSON extraction by fabricating a DirectorDecision
	dd := &fw.DirectorDecision{Reasoning: strings.TrimSpace(out)}
	imp := mapDirectorDecisionToMetrics(dd, t.Event.Category)
	// If still zero, apply category fallback
	if imp == (WorldMetrics{}) {
		imp = mapDirectorDecisionToMetrics(&fw.DirectorDecision{Reasoning: "{}"}, t.Event.Category)
	}
	return strings.TrimSpace(analysis), imp, nil
}

// rewriteEventDescription converts the raw event description into a viral BREAKING NEWS style post
func (g *GameOrchestrator) rewriteEventDescription(ctx context.Context, event GameEvent) (string, error) {
	// Use Llama chat completions endpoint directly; Gemini as fallback
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
	cctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	out, err := llama.New().Complete(cctx, prompt)
	if err == nil && strings.TrimSpace(out) != "" {
		return strings.TrimSpace(out), nil
	}
	if err != nil { log.Printf("[REWRITE] llama endpoint error: %v; trying Gemini fallback", err) }
	// Gemini fallback
	if oo, gerr := g.rewriteEventViaGemini(ctx, event); gerr == nil { 
		g.sim.state.Stats.RewriteGemini++
		return oo, nil 
	}
	return "", fmt.Errorf("empty rewrite")
}

func (g *GameOrchestrator) rewriteEventViaGemini(ctx context.Context, event GameEvent) (string, error) {
	c := gemini.New()
	if c.APIKey == "" { return "", errors.New("GOOGLE_AI_API_KEY not set") }
	pp := fmt.Sprintf(
		"Rewrite the following event into a viral BREAKING NEWS social media post with these rules. Output only the finished post, no preamble.\n\n"+
		"- Headline ALL-CAPS with emojis.\n- Urgent short sentences.\n- Emoji-led bolded section titles.\n- Include a single line placeholder for an image like ``.\n- End with relevant hashtags.\n\n"+
		"Name an entity (use 'AegisNet' if AI/tech implicated), show chaos, mention the White House/The President, contrast official statement vs rumor/intel, end noting all eyes are on the President.\n\n"+
		"Title: %s\nCategory: %s\nSeverity: %d/10\nDescription: %s",
		event.Title, event.Category, event.Severity, event.Description,
	)
	ctx2, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return c.GenerateText(ctx2, pp)
}
