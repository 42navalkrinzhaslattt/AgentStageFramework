package framework

// Model and system constants to avoid hard-coded literals
const (
	ModelDialogueDefault   = "deepseek_r1"
	ModelReasoningDefault  = "deepseek_r1"
	ModelStoryDefault      = "deepseek_r1"
	ModelVoiceDefault      = "kokoro-82m"
	ModelVisionDefault     = "grounding-dino"
	ModelImageDefault      = "flux.1-schnell"
	ModelVideoDefault      = "stable-diffusion-video"
	ModelLlama70B         = "llama_3_1_70b"
)

// Other defaults
const (
	DefaultDialogueMaxTokens = 220
	DefaultReasoningMaxTokens = 300
	DefaultStoryMaxTokens = 400
	DefaultRetryAttempts      = 3
	DefaultRetryBackoffMs     = 200
	DefaultMaxNPCMemory       = 200
	DefaultAssetCacheMax      = 500
)
