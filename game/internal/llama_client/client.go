package llama_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	BaseURL string
	HTTP   *http.Client
	APIKey string
}

func New() *Client {
	base := os.Getenv("LLAMA_CHAT_URL")
	if base == "" {
		base = "https://llama3170b2oczc2osyg-07554694ea35fad5.tec-s20.onthetaedgecloud.com/v1/chat/completions"
	}
	key := os.Getenv("ON_DEMAND_API_ACCESS_TOKEN")
	if key == "" {
		key = os.Getenv("THETA_API_KEY")
	}
	return &Client{BaseURL: base, HTTP: &http.Client{Timeout: 35 * time.Second}, APIKey: key}
}

type LlamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LlamaInput struct {
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Messages    []LlamaMessage `json:"messages"`
	Stream      bool           `json:"stream,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
}

type CompleteReq struct {
	Input LlamaInput `json:"input"`
}

type CompleteResp struct {
	Text    string `json:"text"`
	Output  string `json:"output"`
	Choices []struct{ Text string `json:"text"` } `json:"choices"`
}

func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	payload := CompleteReq{Input: LlamaInput{
		MaxTokens:   500,
		Messages:    []LlamaMessage{{Role: "system", Content: "You are a helpful assistant"}, {Role: "user", Content: prompt}},
		Stream:      false,
		Temperature: 0.5,
		TopP:        0.7,
	}}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(b))
	if err != nil { return "", err }
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.APIKey != "" { req.Header.Set("Authorization", "Bearer "+c.APIKey) }
	resp, err := c.HTTP.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("llama http %d: %s", resp.StatusCode, string(body))
	}
	var cr CompleteResp
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil { return "", err }
	if cr.Text != "" { return cr.Text, nil }
	if cr.Output != "" { return cr.Output, nil }
	if len(cr.Choices) > 0 && cr.Choices[0].Text != "" { return cr.Choices[0].Text, nil }
	return "", errors.New("llama empty response")
}
