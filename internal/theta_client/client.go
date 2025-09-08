// Package theta_client provides a shared client for interacting with Theta EdgeCloud APIs
package theta_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// ThetaClient represents the Theta EdgeCloud API client
type ThetaClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewThetaClient creates a new Theta EdgeCloud client
func NewThetaClient(baseURL, apiKey string) *ThetaClient {
	return &ThetaClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// LLMRequest represents a request to an LLM model
type LLMRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Stop        []string               `json:"stop,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LLMResponse represents a response from an LLM model
type LLMResponse struct {
	ID          string    `json:"id"`
	Object      string    `json:"object"`
	Created     int64     `json:"created"`
	Model       string    `json:"model"`
	Choices     []Choice  `json:"choices"`
	Usage       Usage     `json:"usage"`
	Error       *APIError `json:"error,omitempty"`
	ProcessTime float64   `json:"process_time,omitempty"`
}

// Choice represents a single completion choice
type Choice struct {
	Index        int    `json:"index"`
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ImageGenerationRequest represents a request for image generation
type ImageGenerationRequest struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	Width          int     `json:"width,omitempty"`
	Height         int     `json:"height,omitempty"`
	Steps          int     `json:"steps,omitempty"`
	GuidanceScale  float64 `json:"guidance_scale,omitempty"`
	Seed           int64   `json:"seed,omitempty"`
	Format         string  `json:"format,omitempty"`
}

// ImageGenerationResponse represents an image generation response
type ImageGenerationResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Images      []Image   `json:"images"`
	Error       *APIError `json:"error,omitempty"`
	ProcessTime float64   `json:"process_time,omitempty"`
}

// Image represents a generated image
type Image struct {
	URL    string `json:"url,omitempty"`
	Base64 string `json:"base64,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
}

// TTSRequest represents a text-to-speech request
type TTSRequest struct {
	Text     string                 `json:"text"`
	Voice    string                 `json:"voice,omitempty"`
	Speed    float64                `json:"speed,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TTSResponse represents a text-to-speech response
type TTSResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	AudioURL    string    `json:"audio_url,omitempty"`
	AudioData   []byte    `json:"audio_data,omitempty"`
	Error       *APIError `json:"error,omitempty"`
	ProcessTime float64   `json:"process_time,omitempty"`
}

// VisionRequest represents a vision/object detection request
type VisionRequest struct {
	Image       []byte                 `json:"image"`
	Query       string                 `json:"query,omitempty"`
	Classes     []string               `json:"classes,omitempty"`
	Threshold   float64                `json:"threshold,omitempty"`
	MaxResults  int                    `json:"max_results,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// VisionResponse represents a vision analysis response
type VisionResponse struct {
	ID          string           `json:"id"`
	Status      string           `json:"status"`
	Detections  []Detection      `json:"detections"`
	Description string           `json:"description,omitempty"`
	Error       *APIError        `json:"error,omitempty"`
	ProcessTime float64          `json:"process_time,omitempty"`
}

// Detection represents an object detection result
type Detection struct {
	Label      string    `json:"label"`
	Confidence float64   `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box"`
}

// BoundingBox represents object coordinates
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Model3DRequest represents a 3D model generation request
type Model3DRequest struct {
	Prompt          string                 `json:"prompt"`
	ReferenceImages [][]byte               `json:"reference_images,omitempty"`
	ModelType       string                 `json:"model_type,omitempty"`
	Resolution      string                 `json:"resolution,omitempty"`
	Format          string                 `json:"format,omitempty"`
	IncludeTextures bool                   `json:"include_textures,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Model3DResponse represents a 3D model generation response
type Model3DResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	ModelURL    string    `json:"model_url,omitempty"`
	ModelData   []byte    `json:"model_data,omitempty"`
	TextureURLs []string  `json:"texture_urls,omitempty"`
	Error       *APIError `json:"error,omitempty"`
	ProcessTime float64   `json:"process_time,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("Theta API Error [%d]: %s", e.Code, e.Message)
}

// GenerateWithLLM sends a request to an LLM model
func (c *ThetaClient) GenerateWithLLM(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/inference/llm", c.baseURL)
	var resp LLMResponse
	err := c.sendRequest(ctx, "POST", endpoint, req, &resp)
	return &resp, err
}

// GenerateImage generates an image using FLUX.1-schnell or similar
func (c *ThetaClient) GenerateImage(ctx context.Context, req *ImageGenerationRequest) (*ImageGenerationResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/inference/flux-schnell", c.baseURL)
	var resp ImageGenerationResponse
	err := c.sendRequest(ctx, "POST", endpoint, req, &resp)
	return &resp, err
}

// GenerateVoice generates speech using Kokoro 82M
func (c *ThetaClient) GenerateVoice(ctx context.Context, req *TTSRequest) (*TTSResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/inference/kokoro", c.baseURL)
	var resp TTSResponse
	err := c.sendRequest(ctx, "POST", endpoint, req, &resp)
	return &resp, err
}

// AnalyzeVision performs vision analysis using Grounding Dino
func (c *ThetaClient) AnalyzeVision(ctx context.Context, req *VisionRequest) (*VisionResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/inference/grounding-dino", c.baseURL)
	
	// Create multipart form for image upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Add image data
	imageWriter, err := writer.CreateFormField("image")
	if err != nil {
		return nil, fmt.Errorf("failed to create image field: %w", err)
	}
	if _, err := imageWriter.Write(req.Image); err != nil {
		return nil, fmt.Errorf("failed to write image data: %w", err)
	}
	
	// Add other fields
	if req.Query != "" {
		if err := writer.WriteField("query", req.Query); err != nil {
			return nil, fmt.Errorf("failed to write query field: %w", err)
		}
	}
	
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	var result VisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if result.Error != nil {
		return &result, result.Error
	}
	
	return &result, nil
}

// Generate3DModel generates a 3D model
func (c *ThetaClient) Generate3DModel(ctx context.Context, req *Model3DRequest) (*Model3DResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/inference/3d-generation", c.baseURL)
	var resp Model3DResponse
	err := c.sendRequest(ctx, "POST", endpoint, req, &resp)
	return &resp, err
}

// sendRequest is a helper method for sending HTTP requests
func (c *ThetaClient) sendRequest(ctx context.Context, method, endpoint string, reqBody, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(jsonBody)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Emergent-World-Engine/1.0")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
		}
		return &apiErr
	}
	
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	
	return nil
}

// GetJobStatus checks the status of an async job
func (c *ThetaClient) GetJobStatus(ctx context.Context, jobID string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/v1/jobs/%s", c.baseURL, jobID)
	var result map[string]interface{}
	err := c.sendRequest(ctx, "GET", endpoint, nil, &result)
	return result, err
}

// SetTimeout sets the HTTP client timeout
func (c *ThetaClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetBaseURL updates the base URL for the client
func (c *ThetaClient) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}