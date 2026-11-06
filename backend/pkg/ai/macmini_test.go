package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://192.168.1.50:11434",
			expected: "http://192.168.1.50:11434/v1/chat/completions",
		},
		{
			input:    "http://192.168.1.50:11434/",
			expected: "http://192.168.1.50:11434/v1/chat/completions",
		},
		{
			input:    "http://192.168.1.50:11434/v1",
			expected: "http://192.168.1.50:11434/v1/chat/completions",
		},
		{
			input:    "http://192.168.1.50:8000/v1/chat/completions",
			expected: "http://192.168.1.50:8000/v1/chat/completions",
		},
		{
			input:    "http://localhost:3000",
			expected: "http://localhost:3000/v1/chat/completions",
		},
		{
			input:    "http://localhost:1234/v1/chat/completions",
			expected: "http://localhost:1234/v1/chat/completions",
		},
	}

	for _, tc := range tests {
		got := NormalizeEndpoint(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeEndpoint(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestMacMiniProviderGenerateSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Errorf("expected Authorization header, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Title") != "Karsa HR Assistant" {
			t.Errorf("expected X-Title header, got %s", r.Header.Get("X-Title"))
		}

		var req chatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "hermes-combo" {
			t.Errorf("expected model hermes-combo, got %s", req.Model)
		}

		resp := chatResponse{
			Choices: []struct {
				Text    string      `json:"text,omitempty"`
				Message chatMessage `json:"message,omitempty"`
			}{
				{
					Message: chatMessage{
						Role:    "assistant",
						Content: `{"intent":"attendance_check","needs_clarification":false,"tool":"get_employee_attendance","arguments":{}}`,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	provider := MacMiniProvider{
		URL:     ts.URL,
		Model:   "hermes-combo",
		APIKey:  "secret-token",
		Timeout: 5 * time.Second,
	}

	answer, err := provider.Generate(context.Background(), "Siapa saja yang hadir hari ini?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if answer != `{"intent":"attendance_check","needs_clarification":false,"tool":"get_employee_attendance","arguments":{}}` {
		t.Fatalf("unexpected answer: %s", answer)
	}
}
