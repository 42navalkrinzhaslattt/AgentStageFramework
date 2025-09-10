package ondemand_image_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
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

type fluxInput struct {
	Guidance           float64 `json:"guidance,omitempty"`
	Height             int     `json:"height,omitempty"`
	Image2ImageStrength float64 `json:"image2image_strength,omitempty"`
	NumSteps           int     `json:"num_steps,omitempty"`
	Prompt             string  `json:"prompt"`
	Seed               string  `json:"seed,omitempty"`
	Width              int     `json:"width,omitempty"`
}

type fluxReq struct {
	Input   fluxInput `json:"input"`
	Wait    int       `json:"wait,omitempty"`
	Webhook string    `json:"webhook,omitempty"`
}

type inferResp struct {
	Status string `json:"status"`
	Body struct {
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
			if url := ir.Body.InferRequests[0].Output.ImageURL; url != "" { return url }
		}
	}
	// Fallbacks for other shapes
	var m map[string]any
	if json.Unmarshal(b, &m) != nil { return "" }
	if v, ok := m["image_url"].(string); ok && v != "" { return v }
	if res, ok := m["result"].(map[string]any); ok {
		if v, ok2 := res["image_url"].(string); ok2 && v != "" { return v }
		if imgs, ok2 := res["images"].([]any); ok2 && len(imgs) > 0 {
			if obj, ok3 := imgs[0].(map[string]any); ok3 {
				if u, ok4 := obj["url"].(string); ok4 && u != "" { return u }
			}
		}
	}
	if imgs, ok := m["images"].([]any); ok && len(imgs) > 0 {
		if obj, ok3 := imgs[0].(map[string]any); ok3 {
			if u, ok4 := obj["url"].(string); ok4 && u != "" { return u }
		}
	}
	if u, ok := m["url"].(string); ok && u != "" { return u }
	return ""
}

// --- Google Imagen fallback types and helpers ---

type googlePredictReq struct {
	Instances  []struct{ Prompt string `json:"prompt"` } `json:"instances"`
	Parameters struct{ SampleCount int `json:"sampleCount,omitempty"` } `json:"parameters,omitempty"`
}

type googlePredictResp struct {
	Predictions []struct {
		Bytes string `json:"bytesBase64Encoded"`
		Mime  string `json:"mimeType"`
	} `json:"predictions"`
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

func (c *Client) googleImagenGenerate(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" { apiKey = os.Getenv("GEMINI_API_KEY") }
	if apiKey == "" { return "", errors.New("missing GOOGLE_AI_API_KEY/GEMINI_API_KEY for Imagen fallback") }
	endpoint := os.Getenv("GOOGLE_IMAGEN_PREDICT_URL")
	if endpoint == "" { endpoint = "https://generativelanguage.googleapis.com/v1beta/models/imagen-4.0-generate-001:predict" }

	var reqBody googlePredictReq
	reqBody.Instances = []struct{ Prompt string `json:"prompt"` }{{Prompt: prompt}}
	reqBody.Parameters.SampleCount = 1
	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil { return "", err }
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)
	resp, err := c.HTTP.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // up to ~10MB
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("imagen http %d: %s", resp.StatusCode, string(data))
	}
	// Try predictions shape
	var pr googlePredictResp
	if json.Unmarshal(data, &pr) == nil && len(pr.Predictions) > 0 {
		mime := pr.Predictions[0].Mime
		if mime == "" { mime = "image/png" }
		if pr.Predictions[0].Bytes != "" {
			return fmt.Sprintf("data:%s;base64,%s", mime, pr.Predictions[0].Bytes), nil
		}
	}
	// Try generateContent shape
	var gr genContentResp
	if json.Unmarshal(data, &gr) == nil {
		for _, c := range gr.Candidates {
			for _, p := range c.Content.Parts {
				if p.InlineData.Data != "" {
					mime := p.InlineData.Mime
					if mime == "" { mime = "image/png" }
					return fmt.Sprintf("data:%s;base64,%s", mime, p.InlineData.Data), nil
				}
			}
		}
	}
	return "", fmt.Errorf("imagen response unrecognized: %s", string(data))
}

func (c *Client) Generate(ctx context.Context, prompt string, width, height int) (string, error) {
	// If no Theta token, try Google Imagen directly
	if c.APIKey == "" {
		if url, err := c.googleImagenGenerate(ctx, prompt); err == nil {
			return url, nil
		}
		return "", errors.New("missing ON_DEMAND_API_ACCESS_TOKEN and Google Imagen fallback failed")
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
		// Flux error -> try Google Imagen fallback
		if url, err2 := c.googleImagenGenerate(ctx, prompt); err2 == nil {
			return url, nil
		}
		return "", fmt.Errorf("flux http %d: %s", resp.StatusCode, string(data))
	}
	if url := extractImageURL(data); url != "" { return url, nil }
	// If Flux returned but no URL parsed, try Google Imagen
	if url, err2 := c.googleImagenGenerate(ctx, prompt); err2 == nil {
		return url, nil
	}
	return "", fmt.Errorf("flux response has no image url: %s", string(data))
}
