// Package theta_client provides configuration for Theta EdgeCloud client
package theta_client

import (
	"os"
	"strconv"
	"time"
)

// Config holds configuration for the Theta client
type Config struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    getEnvOrDefault("THETA_BASE_URL", "https://api.theta.ai"),
		APIKey:     getEnvOrDefault("THETA_API_KEY", ""),
		Timeout:    getEnvDurationOrDefault("THETA_TIMEOUT", 30*time.Second),
		RetryCount: getEnvIntOrDefault("THETA_RETRY_COUNT", 3),
		RetryDelay: getEnvDurationOrDefault("THETA_RETRY_DELAY", 1*time.Second),
	}
}

// NewClientFromConfig creates a new Theta client from configuration
func NewClientFromConfig(config *Config) *ThetaClient {
	client := NewThetaClient(config.BaseURL, config.APIKey)
	client.SetTimeout(config.Timeout)
	return client
}

// NewClientFromEnv creates a new Theta client from environment variables
func NewClientFromEnv() *ThetaClient {
	config := DefaultConfig()
	return NewClientFromConfig(config)
}

// Model constants for easy reference
const (
	// LLM Models
	ModelDeepSeekR1     = "deepseek_r1"
	ModelGPTOSS120B     = "gpt-oss-120b"
	ModelGPTOSS20B      = "gpt-oss-20b"
	ModelLlama3170B     = "llama-3.1-70b-instruct"
	ModelLlama3180B     = "llama-3.1-80b-instruct"
	
	// Image Generation Models
	ModelFluxSchnell    = "flux.1-schnell"
	ModelFluxDev        = "flux.1-dev"
	ModelSDXL           = "stable-diffusion-xl"
	
	// Video Generation Models
	ModelStableDiffusionVideo = "stable-diffusion-video"
	
	// Voice Models
	ModelKokoro82M      = "kokoro-82m"
	
	// Vision Models
	ModelGroundingDino  = "grounding-dino"
	
	// 3D Models
	Model3DGeneration   = "3d-generation"
)

// Predefined voice styles for Kokoro
const (
	VoiceStyleNeutral    = "neutral"
	VoiceStyleFriendly   = "friendly"
	VoiceStyleSerious    = "serious"
	VoiceStyleExcited    = "excited"
	VoiceStyleMysterious = "mysterious"
)

// Image formats
const (
	FormatPNG  = "png"
	FormatJPG  = "jpg"
	FormatWEBP = "webp"
)

// Audio formats
const (
	FormatWAV = "wav"
	FormatMP3 = "mp3"
	FormatOGG = "ogg"
)

// Video formats
const (
	FormatMP4  = "mp4"
	FormatWEBM = "webm"
	FormatGIF  = "gif"
	FormatMOV  = "mov"
)

// 3D model formats
const (
	FormatOBJ  = "obj"
	FormatFBX  = "fbx"
	FormatGLTF = "gltf"
	FormatPLY  = "ply"
)

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}