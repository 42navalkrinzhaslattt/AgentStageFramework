package ondemand_image_client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/chai2010/webp"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
	APIKey  string
}

func New() *Client {
	base := os.Getenv("ON_DEMAND_FLUX_URL")
	if base == "" {
		base = "https://ondemand.thetaedgecloud.com/infer_request/flux"
	}
	key := os.Getenv("ON_DEMAND_API_ACCESS_TOKEN")
	if key == "" {
		key = os.Getenv("THETA_API_KEY")
	}
	return &Client{BaseURL: base, HTTP: &http.Client{Timeout: 40 * time.Second}, APIKey: key}
}

// debug helpers
func imgDebug() bool { return true }
func snip(b []byte, n int) string {
	if len(b) <= n { return string(b) }
	return string(b[:n]) + "..."
}

type fluxInput struct {
	Guidance            float64 `json:"guidance,omitempty"`
	Height              int     `json:"height,omitempty"`
	Image2ImageStrength float64 `json:"image2image_strength,omitempty"`
	NumSteps            int     `json:"num_steps,omitempty"`
	Prompt              string  `json:"prompt"`
	Seed                string  `json:"seed,omitempty"`
	Width               int     `json:"width,omitempty"`
}

type fluxReq struct {
	Input   fluxInput `json:"input"`
	Wait    int       `json:"wait,omitempty"`
	Webhook string    `json:"webhook,omitempty"`
}

type inferResp struct {
	Status string `json:"status"`
	Body   struct {
		InferRequests []struct {
			Output struct {
				ImageURL string `json:"image_url"`
			} `json:"output"`
		} `json:"infer_requests"`
	} `json:"body"`
}

// minimal flexible response parsing
func extractImageURL(b []byte) string {
	// Preferred: body.infer_requests[0].output.image_url
	var ir inferResp
	if json.Unmarshal(b, &ir) == nil {
		if len(ir.Body.InferRequests) > 0 {
			if url := ir.Body.InferRequests[0].Output.ImageURL; url != "" {
				return url
			}
		}
	}
	// Fallbacks for other shapes
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	if v, ok := m["image_url"].(string); ok && v != "" {
		return v
	}
	if res, ok := m["result"].(map[string]any); ok {
		if v, ok2 := res["image_url"].(string); ok2 && v != "" {
			return v
		}
		if imgs, ok2 := res["images"].([]any); ok2 && len(imgs) > 0 {
			if obj, ok3 := imgs[0].(map[string]any); ok3 {
				if u, ok4 := obj["url"].(string); ok4 && u != "" {
					return u
				}
			}
		}
	}
	if imgs, ok := m["images"].([]any); ok && len(imgs) > 0 {
		if obj, ok3 := imgs[0].(map[string]any); ok3 {
			if u, ok4 := obj["url"].(string); ok4 && u != "" {
				return u
			}
		}
	}
	if u, ok := m["url"].(string); ok && u != "" {
		return u
	}
	return ""
}

// --- Google Gemini 2.0 Flash image generation (fallback) ---

type geminiGenReq struct {
	Contents []struct {
		Role  string `json:"role,omitempty"`
		Parts []struct {
			Text string `json:"text,omitempty"`
		} `json:"parts"`
	} `json:"contents"`
}

type genContentResp struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				InlineData struct {
					Mime string `json:"mime_type"`
					Data string `json:"data"`
				} `json:"inline_data"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Gemini Go client fallback that converts output to WebP data URL
func (c *Client) googleGeminiImageGenerateClient(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" { apiKey = os.Getenv("GEMINI_API_KEY") }
	if apiKey == "" { return "", errors.New("missing GOOGLE_AI_API_KEY/GEMINI_API_KEY for Gemini image generation") }

	modelName := os.Getenv("GOOGLE_GEMINI_IMAGE_MODEL")
	if modelName == "" { modelName = "gemini-2.5-flash-image-preview" }

	cli, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil { return "", err }
	defer cli.Close()

	model := cli.GenerativeModel(modelName)
	// Do not set ResponseMIMEType here; image models return Blob parts directly

	if imgDebug() { fmt.Printf("[GEMINI-IMG] go-client model=%s promptLen=%d\n", modelName, len(prompt)) }
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		if imgDebug() { fmt.Printf("[GEMINI-IMG] error: %v\n", err) }
		return "", err
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0] == nil || resp.Candidates[0].Content == nil {
		if imgDebug() { fmt.Println("[GEMINI-IMG] empty candidates/content") }
		return "", errors.New("gemini image empty response")
	}
	if imgDebug() { fmt.Printf("[GEMINI-IMG] candidates=%d parts=%d\n", len(resp.Candidates), len(resp.Candidates[0].Content.Parts)) }
	for _, part := range resp.Candidates[0].Content.Parts {
		if imgDebug() { fmt.Printf("[GEMINI-IMG] part type: %T\n", part) }
		switch p := part.(type) {
		case *genai.Blob:
			if imgDebug() { fmt.Printf("[GEMINI-IMG] blob mime=%s bytes=%d\n", p.MIMEType, len(p.Data)) }
			// Try to convert to WebP
			img, _, decErr := image.Decode(bytes.NewReader(p.Data))
			if decErr == nil {
				buf := new(bytes.Buffer)
				if err := webp.Encode(buf, img, &webp.Options{Quality: 80}); err == nil {
					if imgDebug() { fmt.Println("[GEMINI-IMG] converted to webp") }
					return "data:image/webp;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
				}
			}
			// Fallback to original bytes as data URL
			mime := p.MIMEType
			if mime == "" { mime = "image/png" }
			return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(p.Data)), nil
		default:
			// ignore non-blob parts
		}
	}
	if imgDebug() { fmt.Println("[GEMINI-IMG] no blob image part found") }
	return "", errors.New("gemini image parts contained no blob data")
}

func (c *Client) googleGeminiImageGenerate(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return "", errors.New("missing GOOGLE_AI_API_KEY/GEMINI_API_KEY for Gemini image generation")
	}
	model := os.Getenv("GOOGLE_GEMINI_IMAGE_MODEL")
	if model == "" {
		model = "gemini-2.0-flash-preview-image-generation"
	}
	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)

	if imgDebug() {
		fmt.Printf("[GEMINI-IMG] endpoint=%s promptLen=%d\n", endpoint, len(prompt))
	}

	var reqBody geminiGenReq
	reqBody.Contents = []struct {
		Role  string `json:"role,omitempty"`
		Parts []struct {
			Text string `json:"text,omitempty"`
		} `json:"parts"`
	}{
		{Role: "user", Parts: []struct{ Text string `json:"text,omitempty"` }{{Text: prompt}}},
	}
	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // up to ~10MB
	if imgDebug() {
		fmt.Printf("[GEMINI-IMG] http %d, body: %s\n", resp.StatusCode, snip(data, 1200))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini image http %d: %s", resp.StatusCode, string(data))
	}
	// Parse generateContent inline_data shape
	var gr genContentResp
	if json.Unmarshal(data, &gr) == nil {
		for _, c := range gr.Candidates {
			for _, p := range c.Content.Parts {
				if p.InlineData.Data != "" {
					mime := p.InlineData.Mime
					if mime == "" {
						mime = "image/png"
					}
					if imgDebug() {
						fmt.Println("[GEMINI-IMG] parsed inline_data bytes")
					}
					return fmt.Sprintf("data:%s;base64,%s", mime, p.InlineData.Data), nil
				}
			}
		}
	}
	return "", fmt.Errorf("gemini image response unrecognized: %s", string(data))
}

func (c *Client) Generate(ctx context.Context, prompt string, width, height int) (string, error) {
	// If no Theta token, try Google Gemini image generation directly via Go client
	if c.APIKey == "" {
		if imgDebug() { fmt.Println("[IMAGE] No Flux token; using Gemini (Go client) directly") }
		if url, err := c.googleGeminiImageGenerateClient(ctx, prompt); err == nil {
			if imgDebug() { fmt.Println("[IMAGE] Gemini direct success") }
			return url, nil
		} else {
			if imgDebug() { fmt.Printf("[IMAGE] Gemini direct failed: %v\n", err) }
		}
		return "", errors.New("missing ON_DEMAND_API_ACCESS_TOKEN and Gemini image generation failed")
	}

	seed := fmt.Sprintf("%d", rand.Int63())
	payload := fluxReq{Input: fluxInput{Prompt: prompt, Width: width, Height: height, Guidance: 3.5, NumSteps: 4, Seed: seed}, Wait: 6}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(b))
	if err != nil { return "", err }
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.APIKey != "" { req.Header.Set("Authorization", "Bearer "+c.APIKey) }
	resp, err := c.HTTP.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if imgDebug() { fmt.Printf("[FLUX] http %d: %s\n", resp.StatusCode, snip(data, 600)) }
		// Flux error -> try Google Gemini fallback (Go client)
		if url, err2 := c.googleGeminiImageGenerateClient(ctx, prompt); err2 == nil {
			if imgDebug() { fmt.Println("[IMAGE] Fallback to Gemini succeeded") }
			return url, nil
		} else {
			if imgDebug() { fmt.Printf("[IMAGE] Gemini fallback failed: %v\n", err2) }
		}
		return "", fmt.Errorf("flux http %d: %s", resp.StatusCode, string(data))
	}
	if url := extractImageURL(data); url != "" {
		if imgDebug() { fmt.Println("[FLUX] parsed image url from response") }
		return url, nil
	}
	// If Flux returned but no URL parsed, try Google Gemini
	if url, err2 := c.googleGeminiImageGenerateClient(ctx, prompt); err2 == nil {
		if imgDebug() { fmt.Println("[IMAGE] Flux no URL; Gemini succeeded") }
		return url, nil
	} else {
		if imgDebug() { fmt.Printf("[IMAGE] Flux no URL; Gemini failed: %v\n", err2) }
	}
	return "", fmt.Errorf("flux response has no image url: %s", string(data))
}
