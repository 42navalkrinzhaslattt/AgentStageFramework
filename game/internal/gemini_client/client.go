package gemini_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Minimal client for Google AI Studio Generative Language API (Gemini 1.5/Flash)
// We keep it tiny and dependency-free.

type Client struct {
	APIKey string
	HTTP   *http.Client
	Model  string
}

func New() *Client {
	key := os.Getenv("GOOGLE_AI_API_KEY")
	return &Client{
		APIKey: key,
		HTTP:  &http.Client{Timeout: 20 * time.Second},
		Model: "gemini-1.5-flash-latest",
	}
}

type contentPart struct {
	Text string `json:"text"`
}

type contentMessage struct {
	Role  string        `json:"role"`
	Parts []contentPart `json:"parts"`
}

type generateRequest struct {
	Contents []contentMessage `json:"contents"`
}

type candidateContent struct {
	Content struct {
		Parts []struct{ Text string `json:"text"` } `json:"parts"`
	} `json:"content"`
}

type generateResponse struct {
	Candidates     []candidateContent `json:"candidates"`
	PromptFeedback any               `json:"promptFeedback"`
}

func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("missing GOOGLE_AI_API_KEY")
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)
	payload := generateRequest{
		Contents: []contentMessage{{
			Role:  "user",
			Parts: []contentPart{{Text: prompt}},
		}},
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini http %d", resp.StatusCode)
	}
	var gr generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", err
	}
	for _, cand := range gr.Candidates {
		for _, p := range cand.Content.Parts {
			if p.Text != "" {
				return p.Text, nil
			}
		}
	}
	return "", errors.New("empty gemini response")
}
