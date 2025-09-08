package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/emergent-world-engine/backend/internal/theta_client"
)

// AssetGenerator handles AI-powered asset generation
type AssetGenerator struct {
	engine *Engine
	cache  map[string]*Asset
	config *AssetConfig
}

// AssetConfig holds asset generation configuration
type AssetConfig struct {
	ImageModel     string
	Quality        string // "draft", "standard", "high"
	CacheEnabled   bool
	CacheDuration  time.Duration
	OutputFormat   string
	DefaultStyle   string
	MaxConcurrent  int
}

// AssetOption allows configuring asset generation behavior
type AssetOption func(*AssetGenerator)

// WithImageModel sets the AI model for image generation
func WithImageModel(model string) AssetOption {
	return func(ag *AssetGenerator) {
		if ag.config == nil {
			ag.config = &AssetConfig{}
		}
		ag.config.ImageModel = model
	}
}

// WithQuality sets the generation quality level
func WithQuality(quality string) AssetOption {
	return func(ag *AssetGenerator) {
		if ag.config == nil {
			ag.config = &AssetConfig{}
		}
		ag.config.Quality = quality
	}
}

// WithCache enables asset caching
func WithCache(enabled bool, duration time.Duration) AssetOption {
	return func(ag *AssetGenerator) {
		if ag.config == nil {
			ag.config = &AssetConfig{}
		}
		ag.config.CacheEnabled = enabled
		ag.config.CacheDuration = duration
	}
}

// WithDefaultStyle sets a default art style
func WithDefaultStyle(style string) AssetOption {
	return func(ag *AssetGenerator) {
		if ag.config == nil {
			ag.config = &AssetConfig{}
		}
		ag.config.DefaultStyle = style
	}
}

// Asset represents a generated game asset
type Asset struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "image", "texture", "model", "audio"
	Format      string                 `json:"format"`
	Data        []byte                 `json:"data,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Prompt      string                 `json:"prompt"`
	Style       string                 `json:"style"`
	Dimensions  *Dimensions            `json:"dimensions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	GeneratedAt time.Time              `json:"generated_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// Dimensions represents asset dimensions
type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Depth  int `json:"depth,omitempty"`
}

// ImageRequest contains parameters for image generation
type ImageRequest struct {
	Prompt      string            `json:"prompt"`
	Style       string            `json:"style,omitempty"`
	Width       int               `json:"width"`
	Height      int               `json:"height"`
	Quality     string            `json:"quality,omitempty"`
	Seed        int64             `json:"seed,omitempty"`
	Variations  int               `json:"variations"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TextureRequest contains parameters for texture generation
type TextureRequest struct {
	BasePrompt   string                 `json:"base_prompt"`
	TextureType  string                 `json:"texture_type"` // "diffuse", "normal", "roughness", "metallic"
	Material     string                 `json:"material"`     // "stone", "wood", "metal", "fabric"
	Resolution   int                    `json:"resolution"`
	Tileable     bool                   `json:"tileable"`
	Style        string                 `json:"style,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ConceptArtRequest contains parameters for concept art generation
type ConceptArtRequest struct {
	Description   string                 `json:"description"`
	ArtStyle      string                 `json:"art_style"` // "sketch", "painted", "technical"
	Perspective   string                 `json:"perspective"` // "front", "side", "isometric", "perspective"
	Details       string                 `json:"details"`
	ColorPalette  string                 `json:"color_palette,omitempty"`
	References    []string               `json:"references,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateImage creates images using FLUX.1-schnell or other AI models
func (ag *AssetGenerator) GenerateImage(ctx context.Context, req *ImageRequest) (*Asset, error) {
	// Check cache first
	if ag.config != nil && ag.config.CacheEnabled {
		if cached := ag.getCachedAsset(req.Prompt, "image"); cached != nil {
			return cached, nil
		}
	}
	
	// Set defaults
	if req.Width == 0 {
		req.Width = 512
	}
	if req.Height == 0 {
		req.Height = 512
	}
	if req.Variations == 0 {
		req.Variations = 1
	}
	
	// Apply default style if not specified
	style := req.Style
	if style == "" && ag.config != nil && ag.config.DefaultStyle != "" {
		style = ag.config.DefaultStyle
	}
	
	// Enhance prompt with style
	enhancedPrompt := req.Prompt
	if style != "" {
		enhancedPrompt = fmt.Sprintf("%s, %s style", req.Prompt, style)
	}
	
	// Generate image using Theta client
	imgReq := &theta_client.ImageGenerationRequest{
		Prompt: enhancedPrompt,
		Width:  req.Width,
		Height: req.Height,
		Format: "png",
	}
	
	imgResp, err := ag.engine.thetaClient.GenerateImage(ctx, imgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}
	
	// Create asset from response
	var imageURL string
	var imageData []byte
	if len(imgResp.Images) > 0 {
		imageURL = imgResp.Images[0].URL
		// Convert base64 to bytes if available
		if imgResp.Images[0].Base64 != "" {
			// In a real implementation, decode base64 here
			// imageData, _ = base64.StdEncoding.DecodeString(imgResp.Images[0].Base64)
		}
	}

	asset := &Asset{
		ID:     fmt.Sprintf("img_%d", time.Now().UnixNano()),
		Type:   "image",
		Format: "png",
		URL:    imageURL,
		Data:   imageData,
		Prompt: req.Prompt,
		Style:  style,
		Dimensions: &Dimensions{
			Width:  req.Width,
			Height: req.Height,
		},
		Metadata:    req.Metadata,
		GeneratedAt: time.Now(),
	}
	
	// Set expiration if caching
	if ag.config != nil && ag.config.CacheEnabled {
		expiration := time.Now().Add(ag.config.CacheDuration)
		asset.ExpiresAt = &expiration
		ag.cache[ag.getCacheKey(req.Prompt, "image")] = asset
	}
	
	return asset, nil
}

// GenerateTexture creates game textures with specific properties
func (ag *AssetGenerator) GenerateTexture(ctx context.Context, req *TextureRequest) (*Asset, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s_%s_%s", req.BasePrompt, req.TextureType, req.Material)
	if ag.config != nil && ag.config.CacheEnabled {
		if cached := ag.getCachedAsset(cacheKey, "texture"); cached != nil {
			return cached, nil
		}
	}
	
	// Build texture-specific prompt
	prompt := ag.buildTexturePrompt(req)
	
	// Set defaults
	if req.Resolution == 0 {
		req.Resolution = 512
	}
	
	// Generate texture
	imgReq := &theta_client.ImageGenerationRequest{
		Prompt: prompt,
		Width:  req.Resolution,
		Height: req.Resolution,
		Format: "png",
	}
	
	imgResp, err := ag.engine.thetaClient.GenerateImage(ctx, imgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate texture: %w", err)
	}
	
	// Create texture asset from response
	var textureURL string
	var textureData []byte
	if len(imgResp.Images) > 0 {
		textureURL = imgResp.Images[0].URL
		// Convert base64 to bytes if available
		if imgResp.Images[0].Base64 != "" {
			// In a real implementation, decode base64 here
			// textureData, _ = base64.StdEncoding.DecodeString(imgResp.Images[0].Base64)
		}
	}

	asset := &Asset{
		ID:     fmt.Sprintf("tex_%d", time.Now().UnixNano()),
		Type:   "texture",
		Format: "png",
		URL:    textureURL,
		Data:   textureData,
		Prompt: req.BasePrompt,
		Style:  req.Style,
		Dimensions: &Dimensions{
			Width:  req.Resolution,
			Height: req.Resolution,
		},
		Metadata: map[string]interface{}{
			"texture_type": req.TextureType,
			"material":     req.Material,
			"tileable":     req.Tileable,
		},
		GeneratedAt: time.Now(),
	}
	
	// Add to cache
	if ag.config != nil && ag.config.CacheEnabled {
		expiration := time.Now().Add(ag.config.CacheDuration)
		asset.ExpiresAt = &expiration
		ag.cache[ag.getCacheKey(cacheKey, "texture")] = asset
	}
	
	return asset, nil
}

// GenerateConceptArt creates concept art for game design
func (ag *AssetGenerator) GenerateConceptArt(ctx context.Context, req *ConceptArtRequest) (*Asset, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s_%s_%s", req.Description, req.ArtStyle, req.Perspective)
	if ag.config != nil && ag.config.CacheEnabled {
		if cached := ag.getCachedAsset(cacheKey, "concept"); cached != nil {
			return cached, nil
		}
	}
	
	// Build concept art prompt
	prompt := ag.buildConceptArtPrompt(req)
	
	// Generate concept art
	imgReq := &theta_client.ImageGenerationRequest{
		Prompt: prompt,
		Width:  1024,
		Height: 768,
		Format: "png",
	}
	
	imgResp, err := ag.engine.thetaClient.GenerateImage(ctx, imgReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate concept art: %w", err)
	}
	
	// Create concept art asset from response
	var conceptURL string
	var conceptData []byte
	if len(imgResp.Images) > 0 {
		conceptURL = imgResp.Images[0].URL
		// Convert base64 to bytes if available
		if imgResp.Images[0].Base64 != "" {
			// In a real implementation, decode base64 here
			// conceptData, _ = base64.StdEncoding.DecodeString(imgResp.Images[0].Base64)
		}
	}

	asset := &Asset{
		ID:     fmt.Sprintf("concept_%d", time.Now().UnixNano()),
		Type:   "concept_art",
		Format: "png",
		URL:    conceptURL,
		Data:   conceptData,
		Prompt: req.Description,
		Style:  req.ArtStyle,
		Dimensions: &Dimensions{
			Width:  1024,
			Height: 768,
		},
		Metadata: map[string]interface{}{
			"art_style":     req.ArtStyle,
			"perspective":   req.Perspective,
			"color_palette": req.ColorPalette,
		},
		GeneratedAt: time.Now(),
	}
	
	// Add to cache
	if ag.config != nil && ag.config.CacheEnabled {
		expiration := time.Now().Add(ag.config.CacheDuration)
		asset.ExpiresAt = &expiration
		ag.cache[ag.getCacheKey(cacheKey, "concept")] = asset
	}
	
	return asset, nil
}

// GetAsset retrieves a generated asset by ID
func (ag *AssetGenerator) GetAsset(assetID string) (*Asset, bool) {
	// Check cache first
	for _, asset := range ag.cache {
		if asset.ID == assetID {
			// Check if expired
			if asset.ExpiresAt != nil && time.Now().After(*asset.ExpiresAt) {
				delete(ag.cache, ag.getCacheKey(asset.Prompt, asset.Type))
				return nil, false
			}
			return asset, true
		}
	}
	
	// Check Redis if available
	if ag.engine.IsRedisEnabled() {
		assetKey := fmt.Sprintf("asset:%s", assetID)
		var asset Asset
		if err := ag.engine.redisClient.Get(context.Background(), assetKey, &asset); err == nil {
			return &asset, true
		}
	}
	
	return nil, false
}

// ListAssets returns all cached assets
func (ag *AssetGenerator) ListAssets() []*Asset {
	assets := make([]*Asset, 0, len(ag.cache))
	now := time.Now()
	
	for key, asset := range ag.cache {
		// Remove expired assets
		if asset.ExpiresAt != nil && now.After(*asset.ExpiresAt) {
			delete(ag.cache, key)
			continue
		}
		assets = append(assets, asset)
	}
	
	return assets
}

// ClearCache removes all cached assets
func (ag *AssetGenerator) ClearCache() {
	ag.cache = make(map[string]*Asset)
}

// Helper methods

func (ag *AssetGenerator) buildTexturePrompt(req *TextureRequest) string {
	prompt := fmt.Sprintf("High-resolution %s %s texture", req.Material, req.TextureType)
	
	if req.BasePrompt != "" {
		prompt = fmt.Sprintf("%s, %s", req.BasePrompt, prompt)
	}
	
	if req.Tileable {
		prompt += ", seamless tileable pattern"
	}
	
	if req.Style != "" {
		prompt += fmt.Sprintf(", %s style", req.Style)
	}
	
	// Add technical specifications
	switch req.TextureType {
	case "normal":
		prompt += ", normal map, blue and purple tones, surface detail"
	case "roughness":
		prompt += ", roughness map, grayscale, surface roughness variations"
	case "metallic":
		prompt += ", metallic map, black and white, metallic surface properties"
	case "diffuse":
		prompt += ", diffuse color map, realistic surface colors and details"
	}
	
	return prompt
}

func (ag *AssetGenerator) buildConceptArtPrompt(req *ConceptArtRequest) string {
	prompt := req.Description
	
	if req.ArtStyle != "" {
		switch req.ArtStyle {
		case "sketch":
			prompt += ", pencil sketch, line art, conceptual drawing"
		case "painted":
			prompt += ", digital painting, concept art, detailed illustration"
		case "technical":
			prompt += ", technical drawing, blueprint style, precise details"
		default:
			prompt += fmt.Sprintf(", %s art style", req.ArtStyle)
		}
	}
	
	if req.Perspective != "" {
		prompt += fmt.Sprintf(", %s view", req.Perspective)
	}
	
	if req.ColorPalette != "" {
		prompt += fmt.Sprintf(", %s color palette", req.ColorPalette)
	}
	
	if req.Details != "" {
		prompt += fmt.Sprintf(", %s", req.Details)
	}
	
	return prompt
}

func (ag *AssetGenerator) getQualityLevel() string {
	if ag.config != nil && ag.config.Quality != "" {
		return ag.config.Quality
	}
	return "standard"
}

func (ag *AssetGenerator) getCacheKey(prompt, assetType string) string {
	return fmt.Sprintf("%s_%s", assetType, prompt)
}

func (ag *AssetGenerator) getCachedAsset(prompt, assetType string) *Asset {
	key := ag.getCacheKey(prompt, assetType)
	if asset, exists := ag.cache[key]; exists {
		// Check expiration
		if asset.ExpiresAt != nil && time.Now().After(*asset.ExpiresAt) {
			delete(ag.cache, key)
			return nil
		}
		return asset
	}
	return nil
}