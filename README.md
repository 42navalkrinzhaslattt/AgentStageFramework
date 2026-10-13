# AgentStage Framework (Emergent World Engine)

A Go library for LLM-driven game characters and game direction, plus a turn-based Presidential Simulator built
on it: a Go game server that runs AI advisors and AI evaluation on Theta EdgeCloud models with Gemini fallbacks,
and a React chat frontend.

**Stack:** Go · Theta EdgeCloud (DeepSeek R1, Llama 3.1 70B, FLUX) · Google Gemini · Redis (optional) · React    **Context:** Team project with [mechitall](https://github.com/mechitall) and [madvil2](https://github.com/madvil2), September 2025

## Overview

The repository has two parts:

- **`pkg/framework`**, the "Emergent World Engine": an in-process Go package (no separate services) that
  wraps model calls in game-oriented types: `NPC`, `Director`, `Narrative` and `AssetGenerator`, all created from
  one `Engine`.
- **`game/`**, the Presidential Simulator: a separate Go module that imports the framework through a `replace`
  directive. Each turn you get a crisis, three of eight advisors give one-line recommendations, you answer in free
  text, and an LLM "director" scores the consequences on six metrics. It runs in the terminal or as an HTTP API
  for the React app in `frontend/`.

## How the simulator works

**Turn loop** (`game/orchestrator.go`):

1. **Event.** An event is picked from 12 hand-written topic seeds (border standoff, energy shock, cyber recon,
   Supreme Court vacancy, ...), never repeating a topic within a game. One named entity (a US state, country or
   fictional company, depending on the topic) is added to the title, plus a random severity of 5 to 9.
2. **Event image.** Image generation starts in a goroutine with a photojournalism-style prompt. The client tries
   Theta on-demand FLUX first (returns an image URL) and falls back to Gemini image generation, whose output is
   converted to a WebP data URL.
3. **Advisors.** Three of eight advisors are shuffled out and queried in parallel (one goroutine each, 35 s
   timeout). Each must return `{"advisor_opinion": "..."}` from a Llama 3.1 70B endpoint on Theta EdgeCloud. The
   output goes through layered extraction (last valid JSON block, then an "Advisor opinion:" line, then the first
   paragraph that does not look like reasoning) and is rejected if it still contains brackets or braces. On
   failure it falls back to Gemini 2.5 Flash Lite, and finally to a canned line.
4. **Evaluation.** The player's free-text answer goes to the framework `Director` (DeepSeek R1 by default), which
   is asked for a short analysis followed by a one-line JSON of integer deltas (-20 to 20) per metric. If the
   director fails or returns nothing parsable, Gemini evaluates instead and returns an impact level
   (`low` / `medium` / `high` / `extreme`) and direction per metric, which is converted to numeric deltas by range
   (low 5-10, medium 15-30, high 30-50, extreme pushes the metric to the boundary). The JSON is pulled out of free-form
   model output; the impact-level parser uses a balanced-brace scanner to find the nested object. If both providers fail, a small random impact
   (at most ±10 per metric) is used.
5. **State.** Metrics (economy, security, diplomacy, environment, approval, stability) start at random values
   between 40 and 70 and are clamped to [-100, 100]. The game ends after 5 turns, or immediately when any metric
   reaches 0. The web API then returns a plain-text "newspaper" summary of the term.

The server counts which provider answered each call (`AIUsageStats`: Theta or Gemini for advisors and director)
and exposes the counts at `/api/stats`.

## The framework

```go
engine, _ := framework.NewEngine(&framework.Config{ThetaAPIKey: key, EnableLogging: true})
defer engine.Close()

guard := engine.NewNPC("village_guard",
    framework.WithPersonality("Stern but fair"),
    framework.WithBackground("Former soldier"),
    framework.WithRelationship("mayor", "reports_to"),
)
resp, err := guard.GenerateDialogue(ctx, &framework.DialogueRequest{
    PlayerMessage: "Any work around here?",
    Context:       &framework.GameContext{Location: "village_entrance", TimeOfDay: "morning"},
})
```

| Component | What it does |
|---|---|
| `Engine` | holds the Theta client and an optional Redis client; `Health()` and `Metrics()` (request/failure counters) |
| `NPC` | dialogue from personality, background, relationships and memory; streaming variant; optional voice and vision calls; memory in-process with optional Redis persistence (24 h TTL) |
| `Director` | `ProcessEvent` (strategic reasoning over a game event), `AnalyzePlayerBehavior`, `GenerateEvent`, `AdjustDifficulty` |
| `Narrative` | quest and story-event generation, generated choices, lore entries, quest progress tracking |
| `AssetGenerator` | image, video, texture and concept-art requests with a SHA-1-keyed cache, TTL and oldest-first eviction |

All components use functional options (`WithPersonality`, `WithStrategicFocus`, `WithGenre`, `WithCache`, ...)
and guard their state with `sync.RWMutex`.

`internal/theta_client` is the HTTP layer. Generic requests go through a channel-based token-bucket rate limiter
(8 requests/s by default) and retry with linear backoff on 5xx and 429. DeepSeek R1 and Llama 3.1 70B are routed
to their specific EdgeCloud endpoints, and their responses are parsed whether they come back as an SSE stream or
plain JSON.

**Scope note.** The game only uses `Engine` and `Director` (advisors call the Llama client in `game/internal`
directly). The image, video, voice and vision wrappers in the framework post to generic `/v1/inference/...` paths
under the configured base URL and were not exercised by the game. Treat them as an API sketch.

## Build & run

Requires Go 1.25+ and Node.js for the frontend.

```bash
cp game/.env.example game/.env    # THETA_API_KEY / ON_DEMAND_API_ACCESS_TOKEN, GOOGLE_AI_API_KEY

# Game server (web mode, port 8080)
make game-dev                     # = cd game && go run . web 8080
# or build a binary to bin/presidential-sim
make game-build && ./bin/presidential-sim web 8080

# Terminal mode
cd game && go run .

# React frontend (talks to http://localhost:8080/api)
cd frontend && npm install && npm start    # http://localhost:3000
```

`THETA_API_KEY` is required to start the game. When a provider call fails (or its key is missing), the game
falls through to the next fallback: canned advisor lines, random impacts, no image.

Framework targets:

```bash
make test           # go test ./... (framework unit tests, no network)
make run-example    # examples/simple: NPC, director, narrative, asset calls
go run examples/video_generation/main.go
```

To use the framework from another module, add a `replace github.com/emergent-world-engine/backend => <path>`
directive, as `game/go.mod` does (the module path is not a published repository).

## API

The game server exposes `POST /api/start`, `GET /api/state`, `POST /api/new-round`, `POST /api/evaluate-choice`,
`GET /api/stats` and `POST /api/generate-image`, plus two legacy endpoints. Request and response shapes are in
[game/API.md](game/API.md).

## Project structure

```
pkg/framework/          Engine, NPC, Director, Narrative, AssetGenerator, tests
internal/theta_client/  Theta EdgeCloud HTTP client (rate limit, retries, SSE parsing)
internal/redis_client/  go-redis wrapper with NPC/quest/world-state helpers
examples/               framework usage examples
game/                   Presidential Simulator (separate Go module)
  engine.go             advisors, topic seeds, event generation, image prompt
  orchestrator.go       turn loop, advisor extraction, director evaluation, fallbacks
  server.go             HTTP API and end-of-game summary
  internal/             Llama, Gemini and FLUX/Gemini image clients
frontend/               React app (Create React App, styled-components)
```

`game/chat.html` (an AgentStage logo concept, served at `/` by the Go server) and `game/index.html` (a static
newspaper mock-up) are design drafts. The game UI is the React app.

## Team & credits

Forked from [mechitall/AgentStage](https://github.com/mechitall/AgentStage).

- **mechitall** (the upstream's initial commit, git author `neo`): initial Emergent World Engine framework (`pkg/framework`, `internal/`,
  `examples/simple`).
- **Nam Lap Vu** ([42navalkrinzhaslattt](https://github.com/42navalkrinzhaslattt)): framework hardening (DeepSeek R1
  and Llama endpoint routing, rate limiter and retries in the Theta client, SSE parsing and streaming dialogue,
  mutex-guarded state, asset cache eviction, logger, video-generation types); the whole Presidential Simulator
  in `game/` (turn engine, orchestrator, REST API, Llama/Gemini/FLUX clients and fallbacks, impact-level
  parsing, API docs); frontend changes for inline event images and advisor messages.
- **madvil2**: React frontend (`frontend/`).

License: MIT (see [LICENSE](LICENSE)).
