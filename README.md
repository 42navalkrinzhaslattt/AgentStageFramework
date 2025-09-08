# Emergent World Engine

A Go framework for building AI-driven games with intelligent NPCs, dynamic narratives, procedural assets, and strategic game direction. Simply import the package and start creating immersive game experiences powered by cutting-edge AI.

## 🎯 Overview

The Emergent World E## 🔍 Framework Integration Patterns

### Game Engine Integration

```go
// Unity C# example (via CGO or HTTP bridge)
public class NPCController : MonoBehaviour {
    private EmergentEngine engine;
    
    void Start() {
        engine = new EmergentEngine("your_api_key");
        npc = engine.CreateNPC("village_guard");
    }
}
```

```go
// Pure Go game integration
type GameWorld struct {
    engine *framework.Engine
    npcs   map[string]*framework.NPC
}

func (gw *GameWorld) HandlePlayerInteraction(npcID, message string) {
    npc := gw.npcs[npcID]
    response, _ := npc.GenerateDialogue(ctx, &framework.DialogueRequest{
        PlayerMessage: message,
        Context:       gw.GetGameContext(),
    })
    
    gw.DisplayDialogue(npc.ID, response.Message)
    if len(response.AudioData) > 0 {
        gw.PlayAudio(response.AudioData)
    }
}
```a **Go framework** that game developers import directly into their projects. No microservices, no containers, no network complexity - just a clean API that brings AI superpowers to your game.

### Core Components

- **🤖 Intelligent NPCs**: Context-aware dialogue, memory, personality, and voice synthesis
- **🎭 AI Game Director**: Strategic decision making, difficulty scaling, and player behavior analysis  
- **📚 Dynamic Narratives**: Procedural quest generation, branching storylines, and lore management
- **🎨 Asset Generation**: On-demand creation of images, textures, and concept art using FLUX.1-schnell

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Game Process                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │    NPCs     │ │  Director   │ │  Narrative  │ │ Assets │ │
│  │             │ │             │ │             │ │        │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
│                           │                                 │
│                    ┌─────────────┐                         │
│                    │  Framework  │                         │
│                    │   Engine    │                         │
│                    └─────────────┘                         │
└─────────────────────────────────────────────────────────────┘
                            │
                   ┌─────────────────┐
                   │  Theta          │
                   │  EdgeCloud      │
                   │  (AI Models)    │
                   └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- Theta EdgeCloud API key (sign up at [Theta EdgeCloud](https://www.thetaedgecloud.com/))

### Installation

```bash
go get github.com/emergent-world-engine/backend
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/emergent-world-engine/backend/pkg/framework"
)

func main() {
    // Initialize the framework
    config := &framework.Config{
        ThetaAPIKey: "your_theta_api_key_here",
    }
    
    engine, err := framework.NewEngine(config)
    if err != nil {
        panic(err)
    }
    defer engine.Close()
    
    // Create an intelligent NPC
    guard := engine.NewNPC("village_guard",
        framework.WithPersonality("Friendly but cautious"),
        framework.WithVoice(true),
    )
    
    // Generate contextual dialogue
    ctx := context.Background()
    response, _ := guard.GenerateDialogue(ctx, &framework.DialogueRequest{
        PlayerMessage: "Hello, what's happening in town?",
        Context: &framework.GameContext{
            Location: "village_square",
            TimeOfDay: "morning",
        },
    })
    
    fmt.Println("Guard:", response.Message)
}
```

### Advanced Configuration

```go
// Full configuration example
config := &framework.Config{
    ThetaAPIKey:   "your_api_key",
    ThetaEndpoint: "https://api.thetaedgecloud.com", // optional
    RedisURL:      "redis://localhost:6379",          // optional for persistence
    RedisPassword: "",                               // optional
    EnableRedis:   true,                              // enables advanced features
    EnableLogging: true,
}

engine, err := framework.NewEngine(config)
```

### Optional Redis Integration

For advanced features like persistent NPC memory and cross-session state:

```bash
# Install Redis (optional)
brew install redis
redis-server
```

Then enable in config:
```go
config.EnableRedis = true
config.RedisURL = "redis://localhost:6379"
```

## 🎮 Framework Components

### Intelligent NPCs
Create AI-driven characters with personality, memory, and contextual awareness.

**Key Features:**
- Contextual dialogue generation using GPT-OSS-20B
- Persistent memory and relationship tracking
- Environmental perception with Grounding Dino
- Voice synthesis with Kokoro 82M
- Personality-driven responses

**Usage:**
```go
npc := engine.NewNPC("shopkeeper",
    framework.WithPersonality("Greedy but helpful merchant"),
    framework.WithBackground("Former adventurer turned trader"),
    framework.WithVoice(true),
)
```

### AI Game Director
Intelligent game orchestration and player experience optimization.

**Key Features:**
- Strategic decision making using GPT-OSS-120B
- Player behavior analysis and adaptation
- Dynamic difficulty scaling
- Event generation and orchestration

**Usage:**
```go
director := engine.NewDirector(
    framework.WithStrategicFocus("narrative"),
    framework.WithPlayerAnalysis(true),
)
```

### Dynamic Narrative System
Procedural storytelling and quest generation.

**Key Features:**
- AI-powered quest generation
- Branching storylines with player choice
- Lore consistency validation
- Dynamic story events

**Usage:**
```go
narrative := engine.NewNarrative(
    framework.WithGenre("fantasy"),
    framework.WithPlayerChoice(true),
)
```

### Asset Generation
On-demand creation of game assets using AI.

**Key Features:**
- Image generation with FLUX.1-schnell
- Texture creation for 3D models
- Concept art generation
- Asset caching and management

**Usage:**
```go
assets := engine.NewAssetGenerator(
    framework.WithQuality("high"),
    framework.WithCache(true, 24*time.Hour),
)
```

## 🔧 Framework Usage Examples

### Creating Intelligent NPCs

```go
// Create an NPC with personality and voice
innkeeper := engine.NewNPC("tavern_keeper",
    framework.WithPersonality("Warm, gossipy, knows everyone's business"),
    framework.WithBackground("Long-time tavern owner, former bard"),
    framework.WithVoice(true),
    framework.WithRelationship("mayor", "old_friend"),
)

// Generate contextual dialogue
response, err := innkeeper.GenerateDialogue(ctx, &framework.DialogueRequest{
    PlayerMessage: "Any interesting news around town?",
    Context: &framework.GameContext{
        Location:    "cozy_tavern",
        TimeOfDay:   "evening",
        NearbyNPCs:  []string{"drunk_patron", "traveling_merchant"},
        Environment: "warm firelight, smell of ale and stew",
    },
})

fmt.Println("Innkeeper:", response.Message)
if len(response.AudioData) > 0 {
    // Play voice audio
    playAudio(response.AudioData)
}
```

### Generating Game Assets

```go
// Generate weapon artwork
weaponArt, err := assetGen.GenerateImage(ctx, &framework.ImageRequest{
    Prompt: "Legendary flaming sword with intricate engravings",
    Style:  "fantasy game art",
    Width:  512,
    Height: 512,
})

// Generate tileable stone texture  
stoneTexture, err := assetGen.GenerateTexture(ctx, &framework.TextureRequest{
    BasePrompt:  "Ancient castle stone wall",
    TextureType: "diffuse",
    Material:    "stone",
    Resolution:  1024,
    Tileable:    true,
})

fmt.Printf("Generated assets: %s, %s\n", weaponArt.ID, stoneTexture.ID)
```

### Dynamic Quest Generation

```go
// Generate contextual quests
quest, err := narrative.GenerateQuest(ctx, &framework.GameContext{
    Location:     "haunted_forest",
    TimeOfDay:    "midnight",
    NearbyItems:  []string{"mysterious_artifact"},
    PlayerStats:  map[string]interface{}{"level": 10, "class": "ranger"},
})

fmt.Printf("New Quest: %s\n", quest.Title)
fmt.Printf("Description: %s\n", quest.Description)
for _, obj := range quest.Objectives {
    fmt.Printf("- %s\n", obj.Description)
}
```

## 🔌 AI Model Integration

The framework seamlessly integrates with multiple Theta EdgeCloud AI models:

| Model | Purpose | Framework Component |
|-------|---------|---------|
| GPT-OSS-120B | Strategic reasoning & planning | Director, Narrative |
| GPT-OSS-20B | NPC dialogue & interactions | NPCs |
| Kokoro 82M | Text-to-speech synthesis | NPCs (with voice) |
| Grounding Dino | Visual perception & analysis | NPCs (with vision) |
| FLUX.1-schnell | Image & texture generation | AssetGenerator |

### Model Selection

```go
// Use specific models for different components
npc := engine.NewNPC("wizard",
    framework.WithDialogueModel("gpt-oss-120b"),  // Use larger model for complex NPCs
    framework.WithVoiceModel("kokoro-82m"),
)

// Configure asset generation models
assets := engine.NewAssetGenerator(
    framework.WithImageModel("flux.1-schnell"),
    framework.WithQuality("high"),
)
```

## 📊 Monitoring & Observability

The engine includes comprehensive monitoring:

- **Prometheus**: Metrics collection (http://localhost:9090)
- **Grafana**: Visualization dashboards (http://localhost:3000)
- **Redis Commander**: Redis data inspection (http://localhost:8081)

Default Grafana credentials: `admin/admin`

## 🛠️ Framework Development

### Project Structure

```
emergent-world-engine/
├── pkg/framework/          # Core framework package
│   ├── framework.go        # Main engine
│   ├── npc.go              # NPC system
│   ├── director.go         # AI game director
│   ├── narrative.go        # Storytelling system
│   └── assets.go           # Asset generation
├── internal/               # Internal packages
│   ├── theta_client/       # Theta EdgeCloud client
│   └── redis_client/       # Optional Redis client
├── examples/               # Usage examples
│   └── simple/             # Basic framework demo
└── go.mod                  # Go module definition
```

### Building and Testing

```bash
# Run the example
go run examples/simple/main.go

# Run tests
go test ./pkg/framework/

# Build your game with the framework
go build -o mygame ./cmd/mygame/
```

### Extending the Framework

```go
// Add custom NPC behaviors
type CustomNPC struct {
    *framework.NPC
    customBehavior string
}

func (c *CustomNPC) SpecialAction(ctx context.Context) error {
    // Your custom NPC logic
    return nil
}

// Extend with your own components
type GameSpecificComponent struct {
    engine *framework.Engine
}
```

## 🔒 Security & Best Practices

### API Key Security
```go
// ✅ Good: Use environment variables
config := &framework.Config{
    ThetaAPIKey: os.Getenv("THETA_API_KEY"),
}

// ❌ Bad: Hard-coded API keys
config := &framework.Config{
    ThetaAPIKey: "your_key_here", // Never do this!
}
```

### Performance Optimization

```go
// Configure timeouts and limits
config := &framework.Config{
    ThetaAPIKey: os.Getenv("THETA_API_KEY"),
    // Add timeout configuration
}

// Use caching for assets
assets := engine.NewAssetGenerator(
    framework.WithCache(true, 24*time.Hour),
    framework.WithQuality("standard"), // Balance quality vs speed
)

// Batch operations when possible
for _, npc := range npcs {
    go func(n *framework.NPC) {
        // Process NPCs concurrently
        n.GenerateDialogue(ctx, req)
    }(npc)
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🔗 Links

- [Theta EdgeCloud Documentation](https://docs.thetaedgecloud.com/)
- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [Redis Documentation](https://redis.io/documentation)

## 📧 Support

For questions and support, please open an issue on the repository or contact the development team.

---

---

## 🎮 Ready to Build the Future?

```bash
# Get started in 30 seconds
go get github.com/emergent-world-engine/backend
```

Transform your game with AI superpowers:
- **🤖 NPCs that remember and respond intelligently**
- **📚 Stories that adapt to player choices**  
- **🎨 Assets generated on-demand**
- **🎭 Game balance that evolves with players**

**Built with ❤️ for the next generation of AI-driven gaming**