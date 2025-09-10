# Presidential Simulator API

Base URL: http://localhost:8080

All responses are JSON. Set header: Content-Type: application/json for POST bodies.

Status codes:
- 200: OK
- 400: Bad request or game state invalid (e.g., no active turn)
- 405: Method not allowed
- 502: Upstream AI/image generation error

Environment prerequisites (server side):
- Text models: ON_DEMAND_API_ACCESS_TOKEN (or THETA_API_KEY), GOOGLE_AI_API_KEY (fallback)
- Image models: ON_DEMAND_API_ACCESS_TOKEN (Flux). Fallback to Google Imagen uses GOOGLE_AI_API_KEY

---

## Types (selected)

ChatMessage
- name: string
- text: string
- title: string
- titleColor: string (hex color)
- time: RFC3339 string (UTC)
- profilePicture: string (URL)

GameStateResponse
- turn: number
- maxTurns: number
- metrics: WorldMetrics
- isComplete: boolean
- currentTurn: TurnResult | null
- history: TurnResult[]
- stats: AIUsageStats

NewRoundResponse
- gameOver: boolean
- turn: number
- maxTurns: number
- turnResult?: TurnResult (present when gameOver=false)
- metrics?: WorldMetrics (present when gameOver=true)
- newspaper?: string (present when gameOver=true)
- stats: AIUsageStats
- messages?: ChatMessage[]

EvaluateResponse
- evaluation: string
- impact: WorldMetrics
- metrics: WorldMetrics
- isComplete: boolean
- turn: number
- maxTurns: number
- stats: AIUsageStats
- messages?: ChatMessage[]

---

## POST /api/start
Initialize a new game run.

Request body: {}

Response: GameStateResponse

Example:
```
curl -sS -X POST http://localhost:8080/api/start -H 'Content-Type: application/json' -d '{}'
```

---

## GET /api/state
Return current game state.

Response: GameStateResponse

Example:
```
curl -sS http://localhost:8080/api/state
```

---

## POST /api/new-round
Start a new round (create event and advisor opinions). If the game already ended, returns a newspaper summary instead.

Request body: {}

Response: NewRoundResponse
- messages[] contains:
  - News Desk message for the event (title/description), with category-based titleColor
  - One message per advisor (advisor name/advice), with specialty-based titleColor

Example:
```
curl -sS -X POST http://localhost:8080/api/new-round -H 'Content-Type: application/json' -d '{}'
```

---

## POST /api/evaluate-choice
Submit the player’s reasoning for the current event and receive evaluation + impact.

Request body (one of):
- { "reasoning": string }  // preferred
- { "choiceIndex": number, "reasoning": string }
- { "choice": string, "reasoning": string }

Response: EvaluateResponse
- messages[] includes a Director message with the evaluation

Example:
```
curl -sS -X POST http://localhost:8080/api/evaluate-choice \
  -H 'Content-Type: application/json' \
  -d '{"reasoning":"Activate rapid response + targeted relief"}'
```

---

## GET /api/stats
Return AI usage counters.

Response: AIUsageStats

Example:
```
curl -sS http://localhost:8080/api/stats
```

---

## POST /api/generate-image
Generate an illustrative image for the current event. Returns a hosted URL (Flux) or a data URL (Imagen fallback).

Request body (optional):
- { "width": number, "height": number }

Response:
- { "eventId": string, "imageUrl": string }

Example:
```
curl -sS -X POST http://localhost:8080/api/generate-image -H 'Content-Type: application/json' -d '{}'
```

Notes:
- Requires an active turn (start + new-round first)
- If Flux fails or no Theta token, falls back to Google Imagen and returns a data:image/...;base64,... URL

---

## Legacy endpoints

### POST /api/new-turn
Legacy new turn with simpler TurnResult response.

### POST /api/choice
Legacy choice submission with simpler GameStateResponse return.

---

## Colors and profile pictures
- Event titleColor is derived from event category (e.g., security=#B22222, economy=#DAA520)
- Advisor titleColor is derived from advisor specialty
- profilePicture uses RoboHash: https://robohash.org/<slug>.png?size=64x64&set=set3
