package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	fw "github.com/emergent-world-engine/backend/pkg/framework"
	imgc "presidential-simulator/internal/ondemand_image_client"
)

// PresidentSim encapsulates the turn-based chat game simulation
type PresidentSim struct {
	engine    *fw.Engine
	director  *fw.Director
	narrative *fw.Narrative
	state     *GameState
	advisors  map[string]*fw.NPC
}

func NewPresidentSim(apiKey string) (*PresidentSim, error) {
	// loadDotEnv handled by caller (main)
	if apiKey == "" {
		apiKey = getenvFirst([]string{"THETA_API_KEY", "THETA_KEY"})
	}
	eng, err := fw.NewEngine(&fw.Config{ThetaAPIKey: apiKey, EnableLogging: true, ThetaEndpoint: getenv("THETA_BASE_URL")})
	if err != nil {
		return nil, err
	}
	cfg := loadGameConfig()
	// randomize initial metrics within configured range
	minV, maxV := cfg.MetricMin, cfg.MetricMax
	randVal := func() float64 { if maxV > minV { return float64(minV + rand.Intn(maxV-minV+1)) }; return float64(minV) }
	gameState := &GameState{
		Turn:     1,
		MaxTurns: cfg.MaxTurns,
		Metrics: WorldMetrics{
			Economy:     randVal(),
			Security:    randVal(),
			Diplomacy:   randVal(),
			Environment: randVal(),
			Approval:    randVal(),
			Stability:   randVal(),
		},
		History:     []TurnResult{},
		LastUpdated: time.Now(),
		Stats:       AIUsageStats{},
	}

	// Initialize the 8 advisors
	advisorDefinitions := []Advisor{
		{ID: "sec_state", Name: "Sarah Mitchell", Title: "Secretary of State", Personality: "Diplomatic, measured, internationally focused", Specialty: "diplomacy"},
		{ID: "sec_defense", Name: "General Marcus Torres", Title: "Secretary of Defense", Personality: "Decisive, security-focused, strategic", Specialty: "security"},
		{ID: "sec_treasury", Name: "Dr. Rachel Chen", Title: "Secretary of Treasury", Personality: "Analytical, data-driven, economically minded", Specialty: "economy"},
		{ID: "chief_staff", Name: "David Rodriguez", Title: "Chief of Staff", Personality: "Pragmatic, political, big-picture thinker", Specialty: "domestic"},
		{ID: "epa_admin", Name: "Dr. Amanda Green", Title: "EPA Administrator", Personality: "Passionate, science-based, future-oriented", Specialty: "environment"},
		{ID: "nsc_advisor", Name: "Colonel James Wright", Title: "National Security Advisor", Personality: "Intelligence-focused, cautious, thorough", Specialty: "military"},
		{ID: "domestic_policy", Name: "Maria Santos", Title: "Domestic Policy Advisor", Personality: "People-focused, empathetic, reform-minded", Specialty: "social"},
		{ID: "tech_advisor", Name: "Dr. Alex Kim", Title: "Technology Advisor", Personality: "Innovation-focused, forward-thinking, disruptive", Specialty: "tech"},
	}

	gameState.Advisors = advisorDefinitions

	// Create NPC instances for all advisors
	advisorNPCs := make(map[string]*fw.NPC)
	for _, advisor := range advisorDefinitions {
		npc := eng.NewNPC(advisor.ID,
			fw.WithPersonality(advisor.Personality),
			fw.WithBackground(fmt.Sprintf("%s with expertise in %s", advisor.Title, advisor.Specialty)),
		)
		// Removed explicit llama model assignment; external Llama endpoint is used in orchestrator
		advisorNPCs[advisor.ID] = npc
	}

	ps := &PresidentSim{
		engine:   eng,
		state:    gameState,
		advisors: advisorNPCs,
	}

	ps.director = eng.NewDirector(fw.WithStrategicFocus("balance"))
	ps.narrative = eng.NewNarrative(fw.WithGenre("political"), fw.WithTone("tense"), fw.WithPlayerChoice(true))

	return ps, nil
}

func (p *PresidentSim) Close() {
	p.engine.Close()
}

func getenvFirst(keys []string) string {
	for _, k := range keys {
		if v := getenv(k); v != "" {
			return v
		}
	}
	return ""
}
func getenv(k string) string {
	return strings.TrimSpace(os.Getenv(k))
}

// Real-world inspired topic seeds
var historicalTopicSeeds = []struct{Topic, Title, Desc string; Options []string}{
	{"geopolitics","Border Standoff Escalation","Two neighboring countries are posturing militarily near a disputed border; satellite imagery shows armor repositioning.",[]string{"Broker emergency ceasefire talks","Issue strong deterrent statement","Quietly reinforce regional allies","Stay neutral publicly"}},
	{"economy","Energy Market Shock","A sudden supply disruption in a major energy corridor spikes global prices and stresses domestic logistics.",[]string{"Release strategic reserves","Implement temporary price controls","Accelerate renewable deployment package","Let market self-correct"}},
	{"environment","Rapid Glacier Collapse","Unexpected rapid glacial collapse is accelerating sea-level projections and flooding coastal defenses.",[]string{"Announce national resilience initiative","Immediate coastal emergency funding","Commission independent climate audit","Downplay until data verified"}},
	{"technology","Critical AI System Failure","A widely used AI infrastructure provider suffered cascading outages causing transport and finance delays.",[]string{"Mandate emergency patch coordination","Launch federal resilience task force","Advise voluntary rollback to legacy systems","Allow private sector self-recovery"}},
	{"public_health","Novel Pathogen Cluster","Localized cluster of respiratory illness with atypical markers raises early containment concerns.",[]string{"Activate rapid response team","Issue precautionary public guidance","Expand silent surveillance only","Defer—insufficient data"}},
	{"security","Coordinated Cyber Recon","Distributed probing of federal networks suggests staging for larger intrusion attempt.",[]string{"Deploy active defense hunt teams","Increase network segmentation directives","Public-private intelligence brief","Maintain passive monitoring"}},
	{"immigration","Sudden Border Surge","A sharp increase in asylum seekers at the southern border strains facilities and stokes political tensions.",[]string{"Expand humanitarian processing capacity and surge judges","Tighten entry controls and accelerate removals","Negotiate regional burden-sharing with partners","Announce parole program with strict vetting"}},
	{"civil_rights","Voting Access Showdown","Multiple states enact new voting laws challenged in court; civil rights groups mobilize nationwide.",[]string{"Back federal voting rights legislation push","File DOJ challenges and expand consent decrees","Convene bipartisan commission for standards","Defer to states; issue narrow guidance"}},
	{"military_intervention","Crisis Request for Intervention","An allied government requests limited U.S. air support amid an escalating insurgency.",[]string{"Authorize limited airstrikes with strict ROE","Deploy advisors only; no direct combat","Lead diplomatic ceasefire initiative","Decline military involvement"}},
	{"social_safety_net","Benefits Cliff Debate","Rising unemployment triggers debate over extending enhanced benefits amid funding shortfalls.",[]string{"Extend benefits temporarily with oversight","Target aid to high-need sectors only","Pair extensions with workforce retraining","Let enhancements expire on schedule"}},
	{"gun_policy","High-Profile Mass Shooting","A mass shooting reignites national debate over gun safety and constitutional rights.",[]string{"Push universal background checks","Advance assault weapons ban framework","Promote red-flag laws with due-process safeguards","Focus on mental health and school security investments"}},
	{"judicial_appointments","Supreme Court Vacancy","An unexpected Supreme Court vacancy opens with razor-thin Senate margins and intense public scrutiny.",[]string{"Nominate a consensus moderate jurist","Nominate an ideologically aligned jurist to energize base","Delay nomination; pursue procedural deal","Consider recess appointment pathway"}},
}

// --- Named-entity injection to ensure exactly one proper name per event ---
var usStates = []string{"Texas","California","Florida","Ohio","Arizona","Michigan","Georgia","Virginia","Pennsylvania","New York"}
var countries = []string{"Poland","Turkey","Japan","Germany","France","Canada","Mexico","Brazil","India","South Korea"}
var companies = []string{"NorthStar Energy","Orion Analytics","Pioneer Biotech","Apex Dynamics","BlueRidge Systems","Summit Aerospace","Cobalt Rail"}

func pickOne(list []string) string { return list[rand.Intn(len(list))] }

func textContainsAny(text string, names []string) bool {
	lower := strings.ToLower(text)
	for _, n := range names {
		if strings.Contains(lower, strings.ToLower(n)) { return true }
	}
	return false
}

// injectSingleNamedEntity adds exactly one named entity (state, company, or foreign country)
// into the event, preferring to annotate the title with "in/at NAME".
// If the text already contains any of our known names, it will not add another.
func injectSingleNamedEntity(topic, title, desc string) (string, string) {
	combined := title + " " + desc
	allNames := append(append([]string{}, usStates...), append(countries, companies...)...)
	if textContainsAny(combined, allNames) {
		return title, desc // already has a proper name; don't add more
	}

	var name, prep string
	switch topic {
	case "geopolitics", "military_intervention", "security":
		name, prep = pickOne(countries), "in"
	case "economy", "technology":
		name, prep = pickOne(companies), "at"
	case "environment", "public_health":
		name, prep = pickOne(usStates), "in"
	default:
		name, prep = pickOne(usStates), "in"
	}

	// Only add to title to keep a single mention overall
	if !strings.Contains(strings.ToLower(title), strings.ToLower(name)) {
		title = fmt.Sprintf("%s %s %s", title, prep, name)
	}
	return title, desc
}

// GenerateTurnEvent now selects a random topic seed each turn.
func (p *PresidentSim) GenerateTurnEvent(ctx context.Context) (*GameEvent, error) {
	// Build a set of topics already used this game (full history + current)
	used := map[string]bool{}
	for _, t := range p.state.History {
		used[strings.ToLower(t.Event.Category)] = true
	}
	if p.state.CurrentTurn != nil {
		used[strings.ToLower(p.state.CurrentTurn.Event.Category)] = true
	}

	// Collect candidates that have not been used yet
	remaining := make([]struct{ Topic, Title, Desc string; Options []string }, 0, len(historicalTopicSeeds))
	for _, s := range historicalTopicSeeds {
		if !used[strings.ToLower(s.Topic)] {
			remaining = append(remaining, s)
		}
	}

	var seed struct{ Topic, Title, Desc string; Options []string }
	if len(remaining) > 0 {
		seed = remaining[rand.Intn(len(remaining))]
	} else {
		// All topics exhausted (e.g., MaxTurns > unique topics); fallback to any
		seed = historicalTopicSeeds[rand.Intn(len(historicalTopicSeeds))]
	}

	id := fmt.Sprintf("evt_%s_%d", seed.Topic, time.Now().UnixNano())
	sev := 5 + rand.Intn(5)
	// Slight variation injection
	variant := []string{"People are unsure what happens next.", "News reports disagree on what's going on.", "An internal note says we should move quickly but carefully.", "Advisors say we should act soon, but not rush."}[rand.Intn(4)]
	desc := fmt.Sprintf("%s %s", seed.Desc, variant)

	// Add exactly one named entity suited to the topic
	title := seed.Title
	title, desc = injectSingleNamedEntity(seed.Topic, title, desc)

	// Use free-form seed title and description (no templated BREAKING format)
	evt := &GameEvent{ID: id, Title: title, Description: desc, Category: seed.Topic, Severity: sev, Options: seed.Options}

	// Kick off async image generation
	go p.enqueueEventImage(context.Background(), evt)

	return evt, nil
}

func severityLabel(s int) string { switch { case s>=8: return "high"; case s>=6: return "moderate"; default: return "low" } }

// enqueueEventImage builds a news-photo style prompt and requests an image; stores URL on the event when available
func (p *PresidentSim) enqueueEventImage(ctx context.Context, evt *GameEvent) {
	defer func(){ recover() }()
	prompt := buildBBCPhotoPrompt(evt)
	client := imgc.New()
	url, err := client.Generate(ctx, prompt, 800, 450)
	if err != nil { fmt.Println("[IMAGE] generation error:", err); return }
	evt.ImageURL = url
	if p.state != nil && p.state.CurrentTurn != nil && p.state.CurrentTurn.Event.ID == evt.ID {
		p.state.CurrentTurn.Event.ImageURL = url
	}
	fmt.Println("[IMAGE] generated URL:", url)
}

// buildBBCPhotoPrompt creates the requested BBC/AP style prompt with the event details
func buildBBCPhotoPrompt(evt *GameEvent) string {
	return fmt.Sprintf("Create a realistic news photo of this event. Keep it neutral and grounded.\n\nTitle: %s\nCategory: %s (Severity %d/10)\nDetails: %s\n\nStyle:\n- Photojournalism look (BBC/AP).\n- Realistic lighting.\n- Show the place and context (signs, buildings, equipment).\n- Medium-wide shot. Avoid close-ups of faces.\n- Professional camera look (35–50mm).", evt.Title, evt.Category, evt.Severity, evt.Description)
}
