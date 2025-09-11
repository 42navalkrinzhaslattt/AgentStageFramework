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
	"sync"
	"time"

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

	// Check defeat condition: any metric at or below zero -> game over
	if metricTriggersGameOver(g.sim.state.Metrics) {
		// Mark game complete by advancing beyond MaxTurns
		g.sim.state.Turn = g.sim.state.MaxTurns + 1
	}

	// Add to history and advance turn if not already completed
	g.sim.state.History = append(g.sim.state.History, *turnResult)
	if !g.IsGameComplete() {
		g.sim.state.Turn++
	}
	g.sim.state.LastUpdated = time.Now()
	g.sim.state.CurrentTurn = nil
	return nil
}

func metricTriggersGameOver(m WorldMetrics) bool {
	return m.Economy <= 0 || m.Security <= 0 || m.Diplomacy <= 0 || m.Environment <= 0 || m.Approval <= 0 || m.Stability <= 0
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
	backticksRE     = regexp.MustCompile("`+")
	mdBoldRE        = regexp.MustCompile(`\*\*`)
	hashLineRE      = regexp.MustCompile(`(?m)^\s*#Breaking\b.*$`)
	advisorsHdrRE   = regexp.MustCompile(`(?m)^\s*💼\s*Your advisors weigh in:\s*$`)
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
Style and Voice: Address the President directly using second-person ("you", "your"); do not refer to the President in third person. No self-reference (avoid "I", "we").
Constraints: 2-4 sentences. No internal reasoning, no preamble.
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
	return "You should take a stabilizing course."
}

// ImpactDecision represents the model's impact decision for an evaluation
type ImpactDecision struct {
	Level         string `json:"level"`
	Direction     string `json:"direction"` // "+", "-", or "0"
	Justification string `json:"justification,omitempty"`
}

// parseImpactLevelsFromText scans the text for a JSON object containing an `impacts` map
// with keys for metrics and values of {level, direction, justification?}.
func parseImpactLevelsFromText(text string) (map[string]ImpactDecision, bool) {
	text = strings.TrimSpace(text)
	// Trim simple code fences/backticks if present
	text = strings.Trim(text, "`")

	// Prefer extracting the balanced JSON that contains the "impacts" key
	if frag, ok := extractImpactsJSON(text); ok {
		var obj map[string]any
		if json.Unmarshal([]byte(frag), &obj) == nil {
			var impactsRaw any
			if v, ok := obj["impacts"]; ok {
				impactsRaw = v
			} else if v, ok := obj["impact"]; ok {
				impactsRaw = v
			}
			if impactsRaw != nil {
				if m, ok := impactsRaw.(map[string]any); ok {
					res := map[string]ImpactDecision{}
					for k, v := range m {
						key := normalizeMetricKey(k)
						mv, ok2 := v.(map[string]any)
						if !ok2 { continue }
						lvl, _ := mv["level"].(string)
						dir, _ := mv["direction"].(string)
						just, _ := mv["justification"].(string)
						if lvl == "" || dir == "" { continue }
						res[key] = ImpactDecision{Level: strings.ToLower(lvl), Direction: dir, Justification: just}
					}
					if len(res) > 0 { return res, true }
				}
			}
		}
	}

	// NEW: robust fallback – scan for any balanced JSON object containing the key "impacts" (or "impact")
	if frag, ok := findAnyBalancedJSONWithImpact(text); ok {
		var obj map[string]any
		if json.Unmarshal([]byte(frag), &obj) == nil {
			var impactsRaw any
			if v, ok := obj["impacts"]; ok { impactsRaw = v } else if v, ok := obj["impact"]; ok { impactsRaw = v }
			if impactsRaw != nil {
				if m, ok := impactsRaw.(map[string]any); ok {
					res := map[string]ImpactDecision{}
					for k, v := range m {
						key := normalizeMetricKey(k)
						mv, ok2 := v.(map[string]any)
						if !ok2 { continue }
						lvl, _ := mv["level"].(string)
						dir, _ := mv["direction"].(string)
						just, _ := mv["justification"].(string)
						if lvl == "" || dir == "" { continue }
						res[key] = ImpactDecision{Level: strings.ToLower(lvl), Direction: dir, Justification: just}
					}
					if len(res) > 0 { return res, true }
				}
			}
		}
	}

	// Fallback: regex-based scan of JSON candidates (may fail on nested objects but useful as backup)
	locs := jsonCandidateRE.FindAllStringIndex(text, -1)
	for i := len(locs) - 1; i >= 0; i-- {
		frag := text[locs[i][0]:locs[i][1]]
		var obj map[string]any
		if json.Unmarshal([]byte(frag), &obj) != nil { continue }
		var impactsRaw any
		if v, ok := obj["impacts"]; ok { impactsRaw = v } else if v, ok := obj["impact"]; ok { impactsRaw = v }
		if impactsRaw == nil { continue }
		m, ok := impactsRaw.(map[string]any)
		if !ok { continue }
		res := map[string]ImpactDecision{}
		for k, v := range m {
			key := normalizeMetricKey(k)
			mv, ok2 := v.(map[string]any)
			if !ok2 { continue }
			lvl, _ := mv["level"].(string)
			dir, _ := mv["direction"].(string)
			just, _ := mv["justification"].(string)
			if lvl == "" || dir == "" { continue }
			res[key] = ImpactDecision{Level: strings.ToLower(lvl), Direction: dir, Justification: just}
		}
		if len(res) > 0 { return res, true }
	}
	return nil, false
}

// extractImpactsJSON finds a balanced JSON object around the last occurrence of the "impacts" (or "impact") key.
func extractImpactsJSON(s string) (string, bool) {
	low := strings.ToLower(s)
	idx := strings.LastIndex(low, "\"impacts\"")
	if idx == -1 { idx = strings.LastIndex(low, "\"impact\"") }
	if idx == -1 { return "", false }
	// Find opening '{' before the key
	open := -1
	for i := idx; i >= 0; i-- {
		if s[i] == '{' { open = i; break }
	}
	if open == -1 { return "", false }
	// Scan forward to find the matching closing '}' with brace counting while respecting strings and escapes
	end, ok := matchBalancedClosingBrace(s, open)
	if !ok { return "", false }
	return s[open : end+1], true
}

// findAnyBalancedJSONWithImpact scans the entire string, and for each JSON object found (balanced braces),
// returns the last one that contains the key "impacts" (or "impact").
func findAnyBalancedJSONWithImpact(s string) (string, bool) {
	lastFrag := ""
	for i := 0; i < len(s); i++ {
		if s[i] != '{' { continue }
		end, ok := matchBalancedClosingBrace(s, i)
		if !ok { continue }
		frag := s[i : end+1]
		lowFrag := strings.ToLower(frag)
		if strings.Contains(lowFrag, "\"impacts\"") || strings.Contains(lowFrag, "\"impact\"") {
			lastFrag = frag
		}
		// continue scanning to prefer the last occurrence
		i = end
	}
	if lastFrag != "" { return lastFrag, true }
	return "", false
}

func normalizeMetricKey(k string) string {
	k = strings.ToLower(strings.TrimSpace(k))
	switch k {
	case "public_opinion", "public opinion", "opinion":
		return "approval"
	case "national_security", "national security", "security":
		return "security"
	case "geopolitical_standing", "geopolitics", "geopolitical", "diplomacy":
		return "diplomacy"
	case "tech_sector_confidence", "tech", "technology", "stability":
		return "stability"
	case "civil_liberties":
		return "approval"
	case "environment", "climate":
		return "environment"
	case "economy":
		return "economy"
	case "approval":
		return "approval"
	default:
		return k
	}
}

// convertImpactLevelsToDeltas maps level+direction to numeric deltas using ranges.
// low: 5-10, medium: 15-30, high: 30-50, extreme: to boundary.
func convertImpactLevelsToDeltas(levels map[string]ImpactDecision, curr WorldMetrics) WorldMetrics {
	pick := func(min, max int) float64 {
		if max < min { max = min }
		if max == min { return float64(min) }
		return float64(min + rand.Intn(max-min+1))
	}
	magFor := func(level string, dir string, current float64) float64 {
		l := strings.ToLower(strings.TrimSpace(level))
		d := strings.TrimSpace(dir)
		if d == "0" || d == "none" { return 0 }
		neg := d == "-"
		switch l {
		case "low":
			m := pick(5, 10)
			if neg { return -m }
			return m
		case "medium":
			m := pick(15, 30)
			if neg { return -m }
			return m
		case "high":
			m := pick(30, 50)
			if neg { return -m }
			return m
		case "extreme":
			// Push to boundary [-100, 100]
			if neg { return -(current + 100) }
			return 100 - current
		default:
			// default to low
			m := pick(5, 10)
			if neg { return -m }
			return m
		}
	}
	out := WorldMetrics{}
	out.Economy = magFor(levels["economy"].Level, levels["economy"].Direction, curr.Economy)
	out.Security = magFor(levels["security"].Level, levels["security"].Direction, curr.Security)
	out.Diplomacy = magFor(levels["diplomacy"].Level, levels["diplomacy"].Direction, curr.Diplomacy)
	out.Environment = magFor(levels["environment"].Level, levels["environment"].Direction, curr.Environment)
	out.Approval = magFor(levels["approval"].Level, levels["approval"].Direction, curr.Approval)
	out.Stability = magFor(levels["stability"].Level, levels["stability"].Direction, curr.Stability)
	return out
}

// Use Director for evaluation with impact-level parsing
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
	if err == nil {
		// Try new impact-levels parser first
		if levels, ok := parseImpactLevelsFromText(decision.Reasoning); ok {
			imp := convertImpactLevelsToDeltas(levels, g.sim.state.Metrics)
			g.sim.state.Stats.DirectorTheta++
			analysis := extractActionAnalysisText(decision.Reasoning)
			if strings.TrimSpace(analysis) == "" { analysis = formatDirectorNarrative(turnResult, imp) }
			log.Printf("[DIRECTOR] levels parsed latency=%s", time.Since(start))
			return analysis, imp, nil
		}
		// Backward compatibility: try legacy metrics JSON
		if impact, ok := parseDirectorMetricsFromReasoning(decision.Reasoning); ok {
			g.sim.state.Stats.DirectorTheta++
			analysis := extractActionAnalysisText(decision.Reasoning)
			if strings.TrimSpace(analysis) == "" { analysis = formatDirectorNarrative(turnResult, impact) }
			log.Printf("[DIRECTOR] legacy metrics parsed latency=%s", time.Since(start))
			return analysis, impact, nil
		}
	}
	if err != nil { log.Printf("[DIRECTOR] error: %v (trying Gemini fallback)", err) } else { log.Printf("[DIRECTOR] no parsable output; using Gemini path") }

	analysis2, impact2, gerr2 := g.directorMetricsViaGemini(ctx, turnResult)
	if gerr2 == nil {
		g.sim.state.Stats.DirectorGemini++
		log.Printf("[DIRECTOR] Gemini success latency=%s", time.Since(start))
		if strings.TrimSpace(analysis2) == "" { analysis2 = formatDirectorNarrative(turnResult, impact2) }
		return analysis2, impact2, nil
	}
	log.Printf("[DIRECTOR] Gemini evaluation failed detail: %v (using random)", gerr2)
	return g.randomEval(turnResult), g.randomImpact(), nil
}

// Gemini path now requests impact levels + directions and converts to numeric deltas
func (g *GameOrchestrator) directorMetricsViaGemini(ctx context.Context, t *TurnResult) (string, WorldMetrics, error) {
	c := gemini.New()
	if c.APIKey == "" {
		return "", WorldMetrics{}, errors.New("GOOGLE_AI_API_KEY not set")
	}
	pp := fmt.Sprintf(`Event Evaluation Prompt
You are an expert political and economic analyst AI. Evaluate the player's action for its impact on game metrics.

Provide:
1) Action Analysis: 2–4 concise sentences.
2) A final single-line JSON object with categorical impact levels and directions per metric. Use exactly this schema keys and ranges:
{"impacts":{
  "economy":{"level":"low|medium|high|extreme","direction":"+|-|0","justification":"<why>"},
  "security":{"level":"low|medium|high|extreme","direction":"+|-|0","justification":"<why>"},
  "diplomacy":{"level":"low|medium|high|extreme","direction":"+|-|0","justification":"<why>"},
  "environment":{"level":"low|medium|high|extreme","direction":"+|-|0","justification":"<why>"},
  "approval":{"level":"low|medium|high|extreme","direction":"+|-|0","justification":"<why>"},
  "stability":{"level":"low|medium|high|extreme","direction":"+|-|0","justification":"<why>"}
}}
Rules:
- Choose a LEVEL per metric: low (5–10), medium (15–30), high (30–50), extreme (maximal effect).
- Direction: "+" increases the metric, "-" decreases it, "0" means no change.
- Output ONLY the JSON object on the final line. No markdown after it.

Event Description:
%s

Player's Chosen Action:
%s`, t.Event.Description, t.Choice.Reasoning)
	ctx2, cancel := context.WithTimeout(ctx, 22*time.Second)
	defer cancel()
	out, err := c.GenerateText(ctx2, pp)
	if err != nil { return "", WorldMetrics{}, err }
	raw := strings.TrimSpace(out)
	analysis := extractActionAnalysisText(raw)
	levels, ok := parseImpactLevelsFromText(raw)
	if !ok {
		// Log raw Gemini response to help debug formatting/prompting issues
		log.Printf("[GEMINI RAW OUTPUT] %s", raw)
		return analysis, WorldMetrics{}, errors.New("gemini did not return impact levels")
	}
	imp := convertImpactLevelsToDeltas(levels, g.sim.state.Metrics)
	return strings.TrimSpace(analysis), imp, nil
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

// extractActionAnalysisText returns only the "Action Analysis" narrative, without any "Metric Impact" section or trailing JSON.
func extractActionAnalysisText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" { return "" }
	// 0) If an impacts/metrics JSON starts but may be truncated, cut from its opening brace
	lowAll := strings.ToLower(s)
	cutIdx := -1
	if idx := strings.LastIndex(lowAll, "\"impacts\""); idx >= 0 { cutIdx = idx }
	if cutIdx == -1 {
		if idx := strings.LastIndex(lowAll, "\"impact\""); idx >= 0 { cutIdx = idx }
	}
	if cutIdx == -1 {
		if idx := strings.LastIndex(lowAll, "\"metrics\""); idx >= 0 { cutIdx = idx }
	}
	if cutIdx >= 0 {
		open := -1
		for i := cutIdx; i >= 0; i-- { if s[i] == '{' { open = i; break } }
		if open >= 0 { s = strings.TrimSpace(s[:open]) } else { s = strings.TrimSpace(s[:cutIdx]) }
	}
	// 1) Remove trailing JSON block if present
	if start, end := strings.LastIndex(s, "{"), strings.LastIndex(s, "}"); start >= 0 && end > start {
		s = strings.TrimSpace(s[:start])
	}
	// 2) Cut off any Metric Impact section (case-insensitive)
	low := strings.ToLower(s)
	if idx := strings.Index(low, "metric impact"); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}
	// 3) If there is an explicit "Action Analysis" header, keep from there
	low = strings.ToLower(s)
	if idx := strings.Index(low, "action analysis"); idx >= 0 {
		s = strings.TrimSpace(s[idx:])
	}
	return s
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

// formatDirectorNarrative builds a concise analysis header when model analysis text is empty.
func formatDirectorNarrative(t *TurnResult, impact WorldMetrics) string {
	reason := strings.TrimSpace(t.Choice.Reasoning)
	if len(reason) > 220 { reason = reason[:220] + "..." }
	return fmt.Sprintf("Action Analysis: In response to \"%s\" (%s, severity %d), the administration chose: %s.", t.Event.Title, t.Event.Category, t.Event.Severity, reason)
}

// matchBalancedClosingBrace returns the index of the matching '}' starting from an opening '{',
// handling nested braces and JSON string literals with escapes.
func matchBalancedClosingBrace(s string, start int) (int, bool) {
	if start < 0 || start >= len(s) || s[start] != '{' { return -1, false }
	depth := 0
	inStr := false
	esc := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inStr {
			if esc { esc = false; continue }
			if ch == '\\' { esc = true; continue }
			if ch == '"' { inStr = false; continue }
			continue
		}
		if ch == '"' { inStr = true; continue }
		if ch == '{' { depth++ }
		if ch == '}' {
			depth--
			if depth == 0 { return i, true }
		}
	}
	return -1, false
}
