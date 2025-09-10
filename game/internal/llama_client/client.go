package llama_client

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

type Client struct {
	BaseURL string
	HTTP   *http.Client
}

func New() *Client {
	base := os.Getenv("LLAMA_CHAT_URL")
	if base == "" {
		base = "https://llama3170bqobu7bby0a-9c97d6aa8a18a92f.tec-s20.onthetaedgecloud.com/v1/chat/completions"
	}
	return &Client{BaseURL: base, HTTP: &http.Client{Timeout: 35 * time.Second}}
}

type ChatReq struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	body := ChatReq{
		Model:    "llama-3.1-70b-instruct",
		Messages: []ChatMessage{{Role: "user", Content: prompt}},
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(b))
	if err != nil { return "", err }
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llama http %d", resp.StatusCode)
	}
	var cr ChatResp
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil { return "", err }
	if len(cr.Choices) == 0 { return "", errors.New("llama empty choices") }
	return cr.Choices[0].Message.Content, nil
}
