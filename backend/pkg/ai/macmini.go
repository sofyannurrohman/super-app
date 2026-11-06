package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type MacMiniProvider struct {
	URL, Model, APIKey string
	Timeout            time.Duration
	Client             *http.Client
}

type chatRequest struct {
	Model       string        `json:"model,omitempty"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	Stream      bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Text    string      `json:"text,omitempty"`
		Message chatMessage `json:"message,omitempty"`
	} `json:"choices"`
}

func (p MacMiniProvider) Generate(ctx context.Context, prompt string) (string, error) {
	trimmedURL := strings.TrimSpace(p.URL)
	if trimmedURL == "" {
		return "", fmt.Errorf("MAC_MINI_LLM_URL kosong")
	}

	endpoint := NormalizeEndpoint(trimmedURL)
	model := strings.TrimSpace(p.Model)
	if model == "" {
		model = "hermes-combo"
	}

	body, _ := json.Marshal(chatRequest{
		Model:       model,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
		Temperature: 0.2,
		Stream:      false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "http://localhost")
	req.Header.Set("X-Title", "Karsa HR Assistant")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	client := p.Client
	if client == nil {
		timeout := p.Timeout
		if timeout <= 0 {
			timeout = 25 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Mac Mini LLM HTTP %d: %s", resp.StatusCode, string(data))
	}

	var decoded chatResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return "", fmt.Errorf("gagal unmarshal respons Mac Mini (%w): %s", err, string(data))
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("respons Mac Mini kosong")
	}

	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if content == "" {
		content = strings.TrimSpace(decoded.Choices[0].Text)
	}
	if content == "" {
		return "", fmt.Errorf("respons Mac Mini kosong")
	}
	return content, nil
}

// NormalizeEndpoint menormalisasi endpoint URL sehingga kompatibel dengan Ollama, 9router, LM Studio, dan vLLM
func NormalizeEndpoint(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	rawURL = strings.TrimRight(rawURL, "/")
	if strings.HasSuffix(rawURL, "/chat/completions") {
		return rawURL
	}
	if strings.HasSuffix(rawURL, "/v1") {
		return rawURL + "/chat/completions"
	}
	return rawURL + "/v1/chat/completions"
}
