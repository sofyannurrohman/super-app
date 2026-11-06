package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"example.com/hris/backend/pkg/gemini"
)

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type Router struct {
	primary  Provider
	fallback Provider
	timeout  time.Duration
}

func NewRouter(primary, fallback Provider) *Router {
	return &Router{primary: primary, fallback: fallback, timeout: 60 * time.Second}
}

func (r *Router) Generate(ctx context.Context, prompt string) (string, string, error) {
	if r.primary != nil {
		primaryCtx, cancel := context.WithTimeout(ctx, r.timeout)
		answer, err := r.primary.Generate(primaryCtx, prompt)
		cancel()
		if err == nil && strings.TrimSpace(answer) != "" {
			return answer, "mac-mini", nil
		}
		if err != nil {
			log.Printf("[AI] provider mac-mini gagal, fallback ke Gemini: %v", err)
		} else {
			log.Printf("[AI] provider mac-mini mengembalikan jawaban kosong, fallback ke Gemini")
		}
	} else {
		log.Printf("[AI] provider mac-mini tidak dikonfigurasi (MAC_MINI_LLM_URL kosong), memakai Gemini")
	}
	if r.fallback != nil {
		answer, err := r.fallback.Generate(ctx, prompt)
		if err == nil && strings.TrimSpace(answer) != "" {
			return answer, "gemini", nil
		}
		return "", "", err
	}
	return "", "", fmt.Errorf("tidak ada provider AI yang dikonfigurasi")
}

type GeminiProvider struct{ Client *gemini.Client }

func (p GeminiProvider) Generate(ctx context.Context, prompt string) (string, error) {
	// Client Gemini lama belum menerima context; jalankan tetap melalui API yang ada.
	return p.Client.GenerateText(prompt)
}
