// Package theta_client provides a shared client for interacting with Theta EdgeCloud APIs
package theta_client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	fallbackDialogueMaxTokens = 150
	fallbackReasoningMaxTokens = 300
)

// ThetaClient represents the Theta EdgeCloud API client
type ThetaClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	retryAttempts int
	retryBackoff  time.Duration
	rateLimitRPS  int
	tokens        chan struct{}
	onceInit      sync.Once
	metrics       *clientMetrics
}

type clientMetrics struct {
	llmRequests    atomic.Int64
	llmFailures    atomic.Int64
	llmStreamReqs  atomic.Int64
	llmStreamTokens atomic.Int64
}

// NewThetaClient creates a new Theta EdgeCloud client
func NewThetaClient(baseURL, apiKey string) *ThetaClient {
	c := &ThetaClient{
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		baseURL:       baseURL,
		apiKey:        apiKey,
		retryAttempts: 3,
		retryBackoff:  200 * time.Millisecond,
		rateLimitRPS:  8,
		metrics:       &clientMetrics{},
	}
	c.initRateLimiter()
	return c
}

func (c *ThetaClient) initRateLimiter() {
	c.onceInit.Do(func() {
		c.tokens = make(chan struct{}, c.rateLimitRPS)
		for i := 0; i < c.rateLimitRPS; i++ { c.tokens <- struct{}{} }
		go func() {
			ticker := time.NewTicker(time.Second)
			for range ticker.C {
				// refill up to capacity
				for i := len(c.tokens); i < c.rateLimitRPS; i++ {
					select { case c.tokens <- struct{}{}: default: }
				}
			}
		}()
	})
}

// reinitRateLimiter safely replaces limiter after RPS change
func (c *ThetaClient) reinitRateLimiter() {
	capCh := c.rateLimitRPS
	newCh := make(chan struct{}, capCh)
	for i := 0; i < capCh; i++ { newCh <- struct{}{} }
	c.tokens = newCh
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			for i := len(c.tokens); i < cap(c.tokens); i++ { select { case c.tokens <- struct{}{}: default: } }
		}
	}()
}

func (c *ThetaClient) acquire() { <-c.tokens }

// SetRetry configures retry behaviour
func (c *ThetaClient) SetRetry(attempts int, backoff time.Duration) { if attempts>0 { c.retryAttempts = attempts }; if backoff>0 { c.retryBackoff = backoff } }
// SetRateLimit sets requests per second
func (c *ThetaClient) SetRateLimit(rps int) { if rps<=0 { return }; c.rateLimitRPS = rps; c.reinitRateLimiter() }

// LLMRequest represents a request to an LLM model
type LLMRequest struct {
	Model         string                 `json:"model"`
	Prompt        string                 `json:"prompt"`
	MaxTokens     int                    `json:"max_tokens,omitempty"`
	Temperature   float64                `json:"temperature,omitempty"`
	TopP          float64                `json:"top_p,omitempty"`
	Stop          []string               `json:"stop,omitempty"`
	Stream        bool                   `json:"stream,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ResponseFormat interface{}           `json:"response_format,omitempty"`
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

// VideoGenerationRequest represents a request for video generation
type VideoGenerationRequest struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	Width          int     `json:"width,omitempty"`
	Height         int     `json:"height,omitempty"`
	Duration       float64 `json:"duration,omitempty"`        // Video duration in seconds
	FPS            int     `json:"fps,omitempty"`             // Frames per second
	Steps          int     `json:"steps,omitempty"`
	GuidanceScale  float64 `json:"guidance_scale,omitempty"`
	Seed           int64   `json:"seed,omitempty"`
	Format         string  `json:"format,omitempty"`          // "mp4", "webm", "gif"
	Quality        string  `json:"quality,omitempty"`         // "low", "medium", "high"
	MotionStrength float64 `json:"motion_strength,omitempty"` // Controls animation intensity
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// VideoGenerationResponse represents a video generation response
type VideoGenerationResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Videos      []Video   `json:"videos"`
	Error       *APIError `json:"error,omitempty"`
	ProcessTime float64   `json:"process_time,omitempty"`
}

// Video represents a generated video
type Video struct {
	URL      string  `json:"url,omitempty"`
	Base64   string  `json:"base64,omitempty"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Duration float64 `json:"duration"`
	FPS      int     `json:"fps"`
	Format   string  `json:"format"`
	FileSize int64   `json:"file_size,omitempty"`
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
	var endpoint string
	if req.Model == "deepseek_r1" { endpoint = "https://ondemand.thetaedgecloud.com/infer_request/deepseek_r1/completions" } else if req.Model == "llama_3_1_70b" { endpoint = "https://llama3170b2oczc2osyg-07554694ea35fad5.tec-s20.onthetaedgecloud.com/v1/chat/completions" } else { endpoint = fmt.Sprintf("%s/v1/inference/llm", c.baseURL) }
	// DeepSeek custom handling
	if req.Model == "deepseek_r1" || req.Model == "llama_3_1_70b" {
		messages := []map[string]string{{"role":"system","content":"You are an adaptive strategic assistant."},{"role":"user","content":req.Prompt}}
		if req.MaxTokens == 0 { if req.Model == "deepseek_r1" { req.MaxTokens = fallbackReasoningMaxTokens } else { req.MaxTokens = fallbackDialogueMaxTokens } }
		payload := map[string]interface{}{"input": map[string]interface{}{"messages":messages, "max_tokens":req.MaxTokens, "temperature":req.Temperature}}
		if req.ResponseFormat != nil { payload["response_format"] = req.ResponseFormat }
		rawBody, err := json.Marshal(payload); if err != nil { return nil, fmt.Errorf("marshal payload: %w", err) }
		reqHTTP, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(rawBody)); if err != nil { return nil, fmt.Errorf("create request: %w", err) }
		reqHTTP.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		reqHTTP.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(reqHTTP); if err != nil { return nil, fmt.Errorf("request failed: %w", err) }
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body); if err != nil { return nil, fmt.Errorf("read body: %w", err) }
		if resp.StatusCode >= 400 { return nil, fmt.Errorf("%s http %d: %s", req.Model, resp.StatusCode, snippet(string(data),180)) }
		// Parse SSE style lines if they are streamed, else treat as direct JSON
		text := parseSSEorJSONCompletion(data)
		if text == "" { return nil, fmt.Errorf("%s produced no content", req.Model) }
		c.metrics.llmRequests.Add(1)
		return &LLMResponse{Model: req.Model, Choices: []Choice{{Index:0, Text: text}}}, nil
	}
	var resp LLMResponse
	err := c.sendRequest(ctx, "POST", endpoint, req, &resp)
	return &resp, err
}

// helper to parse either SSE style or plain JSON for llama/deepseek endpoints
func parseSSEorJSONCompletion(data []byte) string {
	str := string(data)
	if strings.Contains(str, "\ndata:") { // SSE style
		lines := strings.Split(str, "\n")
		var agg strings.Builder
		for _, line := range lines { line = strings.TrimSpace(line); if line=="" || line=="[DONE]" { continue }; if strings.HasPrefix(line, "data:") { line = strings.TrimSpace(strings.TrimPrefix(line, "data:")) } else { continue }; if line=="" { continue }; var obj map[string]interface{}; if json.Unmarshal([]byte(line), &obj)!=nil { continue }; if choices, ok := obj["choices"].([]interface{}); ok { for _, ch := range choices { if m,ok2 := ch.(map[string]interface{}); ok2 { if delta,ok3 := m["delta"].(map[string]interface{}); ok3 { if content,ok4 := delta["content"].(string); ok4 { agg.WriteString(content) } } ; if txt,ok5 := m["text"].(string); ok5 { agg.WriteString(txt) } } } }
		}
		return agg.String()
	}
	// Try plain JSON
	var obj map[string]interface{}
	if json.Unmarshal(data, &obj)==nil { if choices, ok := obj["choices"].([]interface{}); ok { var agg strings.Builder; for _, ch := range choices { if m,ok2 := ch.(map[string]interface{}); ok2 { if txt,ok3 := m["text"].(string); ok3 { agg.WriteString(txt) }; if delta,ok4 := m["delta"].(map[string]interface{}); ok4 { if content,ok5 := delta["content"].(string); ok5 { agg.WriteString(content) } } } }; if agg.Len()>0 { return agg.String() } } }
	return strings.TrimSpace(str)
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

// GenerateVideo generates video using Stable Diffusion Video
func (c *ThetaClient) GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/inference/stable-video-diffusion", c.baseURL)
	var resp VideoGenerationResponse
	err := c.sendRequest(ctx, "POST", endpoint, req, &resp)
	return &resp, err
}

// sendRequest is a helper method for sending HTTP requests
func (c *ThetaClient) sendRequest(ctx context.Context, method, endpoint string, reqBody, respBody interface{}) error {
	var rawBody []byte
	var err error
	if reqBody != nil {
		rawBody, err = json.Marshal(reqBody)
		if err != nil { return fmt.Errorf("failed to marshal request: %w", err) }
	}
	attempts := c.retryAttempts
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		c.acquire()
		var body io.Reader
		if rawBody != nil { body = bytes.NewReader(rawBody) }
		req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
		if err != nil { return fmt.Errorf("failed to create request: %w", err) }
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Emergent-World-Engine/1.0")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < attempts-1 { time.Sleep(time.Duration(attempt+1)*c.retryBackoff); continue }
			c.metrics.llmFailures.Add(1)
			return fmt.Errorf("request failed: %w", err)
		}
		func() {
			defer resp.Body.Close()
			// Read full body for logging then decode
			data, readErr := io.ReadAll(resp.Body)
			if readErr != nil { lastErr = readErr; err = fmt.Errorf("read body: %w", readErr); return }
			if resp.StatusCode >= 400 {
				var apiErr APIError
				_ = json.Unmarshal(data, &apiErr)
				if apiErr.Message == "" { apiErr.Message = strings.TrimSpace(string(data)) }
				apiErr.Code = resp.StatusCode
				log.Printf("[THETA][HTTP %d] endpoint=%s body_snip=%q", resp.StatusCode, endpoint, snippet(string(data), 240))
				lastErr = &apiErr
				// Retry on 5xx or 429
				if resp.StatusCode >=500 || resp.StatusCode==429 {
					if attempt < attempts-1 { time.Sleep(time.Duration(attempt+1)*c.retryBackoff); err = &apiErr; return }
				}
				err = &apiErr
				return
			}
			if respBody != nil {
				if decErr := json.Unmarshal(data, respBody); decErr != nil {
					log.Printf("[THETA][DECODE ERR] endpoint=%s err=%v raw_snip=%q", endpoint, decErr, snippet(string(data), 240))
					err = fmt.Errorf("failed to decode response: %w", decErr)
					lastErr = err
					return
				}
				c.metrics.llmRequests.Add(1)
			}
		}()
		if err == nil { return nil }
		lastErr = err
	}
	if lastErr == nil { return fmt.Errorf("exhausted retries: unknown error") }
	return fmt.Errorf("exhausted retries: last error: %v", lastErr)
}

// GenerateWithLLMStream streams an LLM completion (best-effort generic SSE/line JSON parser)
func (c *ThetaClient) GenerateWithLLMStream(ctx context.Context, req *LLMRequest) (<-chan string, <-chan error) {
	out := make(chan string, 32); errCh := make(chan error, 1)
	go func(){
		defer close(out); defer close(errCh); c.acquire()
		endpoint := ""; var body io.Reader
		if req.Model == "deepseek_r1" {
			endpoint = "https://ondemand.thetaedgecloud.com/infer_request/deepseek_r1/completions?stream=true"
			messages := []map[string]string{{"role":"system","content":"You are an adaptive strategic assistant."},{"role":"user","content":req.Prompt}}
			if req.MaxTokens == 0 { req.MaxTokens = fallbackDialogueMaxTokens }
			payload := map[string]interface{}{"input": map[string]interface{}{"messages":messages,"max_tokens":req.MaxTokens,"temperature":req.Temperature,"stream":true}}
			if req.TopP > 0 { payload["input"].(map[string]interface{})["top_p"] = req.TopP }
			jsonBody, e := json.Marshal(payload); if e != nil { errCh <- e; return }; body = bytes.NewReader(jsonBody)
		} else if req.Model == "llama_3_1_70b" {
			endpoint = "https://llama3170b2oczc2osyg-07554694ea35fad5.tec-s20.onthetaedgecloud.com/v1/chat/completions?stream=true"
			messages := []map[string]string{{"role":"system","content":"You are an adaptive strategic assistant."},{"role":"user","content":req.Prompt}}
			if req.MaxTokens == 0 { req.MaxTokens = fallbackDialogueMaxTokens }
			payload := map[string]interface{}{"input": map[string]interface{}{"messages":messages,"max_tokens":req.MaxTokens,"temperature":req.Temperature,"stream":true}}
			if req.TopP > 0 { payload["input"].(map[string]interface{})["top_p"] = req.TopP }
			jsonBody, e := json.Marshal(payload); if e != nil { errCh <- e; return }; body = bytes.NewReader(jsonBody)
		} else {
			endpoint = fmt.Sprintf("%s/v1/inference/llm?stream=true", c.baseURL); req.Stream = true; jsonBody, e := json.Marshal(req); if e != nil { errCh <- e; return }; body = bytes.NewReader(jsonBody)
		}
		httpReq, e := http.NewRequestWithContext(ctx, "POST", endpoint, body); if e != nil { errCh <- e; return }
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey)); httpReq.Header.Set("Content-Type","application/json")
		resp, e := c.httpClient.Do(httpReq); if e != nil { errCh <- e; return }
		if resp.StatusCode >=400 { b,_ := io.ReadAll(resp.Body); errCh <- fmt.Errorf("stream http %d: %s", resp.StatusCode, snippet(string(b),180)); resp.Body.Close(); return }
		c.metrics.llmStreamReqs.Add(1); reader := bufio.NewReader(resp.Body)
		for { line, e := reader.ReadString('\n'); if len(line)>0 { line = strings.TrimSpace(line); if strings.HasPrefix(line,"data:") { line = strings.TrimSpace(strings.TrimPrefix(line,"data:")) }; if line=="[DONE]" { break }; if line=="" { continue }; var obj map[string]interface{}; if json.Unmarshal([]byte(line), &obj)==nil { if delta,ok:=obj["delta"].(string); ok { out<-delta; c.metrics.llmStreamTokens.Add(1); continue }; if text,ok:=obj["text"].(string); ok { out<-text; c.metrics.llmStreamTokens.Add(1); continue }; if choices,ok:=obj["choices"].([]interface{}); ok { for _,ch := range choices { if m,ok2:=ch.(map[string]interface{}); ok2 { // handle direct text
				if t,ok3:=m["text"].(string); ok3 { out<-t; c.metrics.llmStreamTokens.Add(1) }
				// handle nested delta.content (DeepSeek/Llama style)
				if deltaMap,ok4:=m["delta"].(map[string]interface{}); ok4 { if content,ok5:=deltaMap["content"].(string); ok5 { out<-content; c.metrics.llmStreamTokens.Add(1) } }
			} }; continue } }; out<-line }; if e!=nil { if errors.Is(e, io.EOF) { break }; errCh<-e; break } }
		resp.Body.Close()
	}(); return out, errCh
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

// Metrics public snapshot
type ClientMetrics struct {
	LLMRequests int64
	LLMFailures int64
	LLMStreamRequests int64
	LLMStreamTokens int64
}

func (c *ThetaClient) Metrics() ClientMetrics {
	return ClientMetrics{ LLMRequests: c.metrics.llmRequests.Load(), LLMFailures: c.metrics.llmFailures.Load(), LLMStreamRequests: c.metrics.llmStreamReqs.Load(), LLMStreamTokens: c.metrics.llmStreamTokens.Load() }
}

// AnalyzeVision performs vision analysis using Grounding Dino (improved multipart with file field)
func (c *ThetaClient) AnalyzeVision(ctx context.Context, req *VisionRequest) (*VisionResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/inference/grounding-dino", c.baseURL)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// file part
	fileWriter, err := writer.CreateFormFile("image", "upload.png")
	if err != nil { return nil, fmt.Errorf("failed form file: %w", err) }
	if _, err := fileWriter.Write(req.Image); err != nil { return nil, fmt.Errorf("failed write image: %w", err) }
	if req.Query != "" { _ = writer.WriteField("query", req.Query) }
	if err := writer.Close(); err != nil { return nil, fmt.Errorf("close multipart: %w", err) }
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, body)
	if err != nil { return nil, fmt.Errorf("failed to create request: %w", err) }
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := c.httpClient.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("request failed: %w", err) }
	defer resp.Body.Close()
	var result VisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return nil, fmt.Errorf("decode response: %w", err) }
	if result.Error != nil { return &result, result.Error }
	return &result, nil
}

func snippet(s string, n int) string { if len(s) <= n { return s }; return s[:n] + "..." }