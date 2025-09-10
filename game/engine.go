package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	fw "github.com/emergent-world-engine/backend/pkg/framework"
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
	eng, err := fw.NewEngine(&fw.Config{ThetaAPIKey: apiKey, EnableLogging: true})
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
		// switch dialogue model to llama 70b
		if npc != nil { npc.Config().DialogueModel = fw.ModelLlama70B }
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

// GenerateTurnEvent now selects a random topic seed each turn.
func (p *PresidentSim) GenerateTurnEvent(ctx context.Context) (*GameEvent, error) {
	seed := historicalTopicSeeds[rand.Intn(len(historicalTopicSeeds))]
	id := fmt.Sprintf("evt_%s_%d", seed.Topic, time.Now().UnixNano())
	sev := 5 + rand.Intn(5)
	// Slight variation injection
	variant := []string{"Analysts warn second-order impacts may cascade.","Media narratives are diverging, amplifying uncertainty.","Internal brief suggests window for decisive framing.","Advisory councils urge calibrated but swift response."}[rand.Intn(4)]
	desc := fmt.Sprintf("%s %s", seed.Desc, variant)

	// NEW: Pre-format into BREAKING NEWS style at generation time
	title, formatted := formatBreakingNewsPost(seed.Title, seed.Topic, sev, desc)
	return &GameEvent{ID:id, Title:title, Description:formatted, Category:seed.Topic, Severity:sev, Options: seed.Options}, nil
}

func severityLabel(s int) string { switch { case s>=8: return "high"; case s>=6: return "moderate"; default: return "low" } }

// formatBreakingNewsPost produces a viral BREAKING NEWS style post without model calls
func formatBreakingNewsPost(title, category string, severity int, baseDesc string) (string, string) {
	upper := strings.ToUpper(title)
	headline := "🚨 BREAKING: " + upper

	// Pick an entity name deterministically by category
	entity := map[string]string{
		"technology":           "AegisNet",
		"economy":              "Capstone Energy",
		"environment":          "Coastal Resilience Authority",
		"public_health":        "Sterling Biolabs",
		"security":             "Homeland Cyber Command",
		"geopolitics":          "Atlantic Council Desk",
		"diplomacy":            "Atlantic Council Desk",
		"immigration":          "Border Operations Directorate",
		"civil_rights":         "Liberty Watch Coalition",
		"military_intervention": "Joint Operations Center",
		"social_safety_net":    "National Benefits Agency",
		"gun_policy":           "Federal Safety Commission",
		"judicial_appointments": "Senate Judiciary Staff",
	}[strings.ToLower(category)]
	if entity == "" { entity = "National Response Center" }

	// On-the-ground consequences by category
	onGround := map[string]string{
		"technology":            "Payment rails stall, transit apps fail, call centers flooded.",
		"economy":               "Prices spike, freight queues grow, small businesses pause orders.",
		"environment":           "Coastal defenses overwhelmed, evacuations and sandbag lines form.",
		"public_health":         "Clinics crowd, pharmacies ration, local schools reassess schedules.",
		"security":              "Agencies shift to high alert; federal networks segment and harden.",
		"geopolitics":           "Airspace notices change, insurers reprice risk, embassies restrict movement.",
		"diplomacy":             "Airspace notices change, insurers reprice risk, embassies restrict movement.",
		"immigration":           "Shelters overflow; buses rerouted; city resources stretched thin.",
		"civil_rights":          "Courthouses packed; rallies surge; legal hotlines spike calls.",
		"military_intervention": "Runways light up; tankers launch; allied press briefings multiply.",
		"social_safety_net":     "Benefits portals jam; food banks surge; city councils call emergencies.",
		"gun_policy":            "Vigils form; police presence expands; cable panels go wall-to-wall.",
		"judicial_appointments": "Advocacy groups mobilize; Senate corridors teem; fundraising explodes.",
	}[strings.ToLower(category)]
	if onGround == "" { onGround = "Systems strain and local responders scramble as uncertainty grows." }

	// Official vs rumor framing
	official := "officials call it a 'technical issue'"
	rumor := "early intel hints at a coordinated operation"
	switch strings.ToLower(category) {
	case "technology", "security":
		official = "providers cite a 'software glitch'"
		rumor = "sources fear a foreign cyber operation"
	case "environment":
		official = "agencies cite 'anomalous weather'"
		rumor = "leaked memos point to neglected defenses"
	case "public_health":
		official = "health brief calls it 'contained'"
		rumor = "labs whisper novel mutations and undercounts"
	case "economy":
		official = "Treasury frames a 'temporary dislocation'"
		rumor = "traders cite deeper structural fractures"
	case "civil_rights":
		official = "courts urge patience"
		rumor = "advocates allege coordinated suppression"
	case "immigration":
		official = "DHS says 'surge protocols active'"
		rumor = "border agents cite policy whiplash and gaps"
	}

	// Hashtags
	tag := func(s string) string { return "#" + strings.ReplaceAll(strings.Title(strings.ReplaceAll(s, "_", " ")), " ", "") }
	cats := []string{"Breaking", strings.ToLower(category), "Crisis", "WhiteHouse"}
	for i, c := range cats { cats[i] = tag(c) }
	hashtags := strings.Join(cats, " ")

	var b strings.Builder
	fmt.Fprintf(&b, "🧭 **What’s Happening:** %s (severity %d/10).\n\n", baseDesc, severity)
	fmt.Fprintf(&b, "🌐 **On The Ground:** %s\n\n", onGround)
	fmt.Fprintf(&b, "🏢 **Named Entity:** %s is central to unfolding operations.\n\n", entity)
	fmt.Fprintf(&b, "🗣️ **Official vs Rumors:** %s, but %s.\n\n", official, rumor)
	b.WriteString("🏛️ **Politics:** The White House is in crisis mode; opponents blame the administration for unforced errors. All eyes are on the President to act next.\n\n")
	b.WriteString("``\n\n") // single-line image placeholder
	fmt.Fprintf(&b, "%s", hashtags)

	return headline, strings.TrimSpace(b.String())
}
