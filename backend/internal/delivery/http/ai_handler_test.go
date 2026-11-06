package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"example.com/hris/backend/pkg/ai"
)

type helpProvider struct {
	calls  int
	prompt string
}

func (p *helpProvider) Generate(_ context.Context, prompt string) (string, error) {
	p.calls++
	p.prompt = prompt
	return "Produk perusahaan contoh meliputi kecap, saus, dan cuka makan.", nil
}

func TestAIChatNeverUsesPrivateContextOrOperationalRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &helpProvider{}
	r := gin.New()
	NewAIHandler(ai.NewRouter(provider, nil)).RegisterRoutes(r.Group("/api/v1"))

	for _, path := range []string{"/api/v1/ai/plan", "/api/v1/ai/answer", "/api/v1/ai/tools/execute"} {
		response := httptest.NewRecorder()
		r.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, strings.NewReader("{}")))
		if response.Code != http.StatusNotFound {
			t.Errorf("%s should be unavailable, got %d", path, response.Code)
		}
	}

	post := func(body string) map[string]any {
		t.Helper()
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Role", "super_admin")
		r.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("chat status = %d: %s", response.Code, response.Body.String())
		}
		var data map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
			t.Fatal(err)
		}
		return data
	}

	denied := post(`{"message":"Berapa jumlah karyawan perusahaan contoh?","page":"gaji rahasia 123","history":[{"role":"assistant","content":"gaji rahasia 123"}]}`)
	if denied["answer"] != ai.OutOfScopeAnswer || provider.calls != 0 {
		t.Fatalf("private question reached AI provider: %#v", denied)
	}

	post(`{"message":"Apa saja kategori produk perusahaan contoh?","page":"gaji rahasia 123","history":[{"role":"assistant","content":"gaji rahasia 123"}]}`)
	if provider.calls != 1 || strings.Contains(provider.prompt, "gaji rahasia") || strings.Contains(provider.prompt, "123") {
		t.Fatalf("private client context reached AI provider: %q", provider.prompt)
	}
}
