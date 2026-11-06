package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client adalah wrapper ringan untuk Gemini REST API.
// Menggunakan pure HTTP (sama seperti analytics_usecase.go) tanpa SDK tambahan.
type Client struct {
	apiKey    string
	modelName string
	baseURL   string
	client    *http.Client
}

// NewClient membuat Gemini client baru dari env vars yang sudah ada.
func NewClient(apiKey, modelName string) *Client {
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}
	modelName = strings.TrimPrefix(modelName, "models/")
	return &Client{
		apiKey:    apiKey,
		modelName: modelName,
		baseURL:   "https://generativelanguage.googleapis.com/v1beta",
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// GenerateText mengirim prompt ke Gemini dan mengembalikan teks respons.
func (c *Client) GenerateText(prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY tidak dikonfigurasi")
	}

	apiURL := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		c.baseURL, c.modelName, c.apiKey)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 2048,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpClient := c.client
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("HTTP request ke Gemini gagal: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("baca respons Gemini: %w", err)
	}

	var gResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &gResp); err != nil {
		return "", fmt.Errorf("unmarshal respons Gemini: %w", err)
	}

	if gResp.Error != nil {
		return "", fmt.Errorf("Gemini API error %d: %s", gResp.Error.Code, gResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini HTTP status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("Gemini mengembalikan respons kosong")
	}

	return strings.TrimSpace(gResp.Candidates[0].Content.Parts[0].Text), nil
}
