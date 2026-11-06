package http

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"example.com/hris/backend/pkg/ai"
)

// AIHandler hanya menjawab pengetahuan produk dan panduan penggunaan aplikasi.
// Tidak ada tool, database, atau data halaman yang diberikan kepada model.
type AIHandler struct {
	router *ai.Router
}

func NewAIHandler(router *ai.Router) *AIHandler {
	return &AIHandler{router: router}
}

type aiChatRequest struct {
	Message string `json:"message"`
}

func (h *AIHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/ai/chat", h.Chat)
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req aiChatRequest
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message wajib diisi"})
		return
	}

	message := strings.TrimSpace(req.Message)
	if !ai.IsPublicHelpQuestion(message) {
		c.JSON(http.StatusOK, gin.H{
			"answer":   ai.OutOfScopeAnswer,
			"provider": "scope-policy",
		})
		return
	}

	// Abaikan page dan history dari klien: keduanya tidak tepercaya dan dapat berisi data privat.
	prompt := ai.BuildSystemPrompt("") + "\n\nPertanyaan pengguna: " + message
	answer, provider, err := h.router.Generate(c.Request.Context(), prompt)
	if err != nil {
		log.Printf("[AIHandler] gagal menjawab bantuan publik: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Asisten AI sedang tidak tersedia"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"answer": answer, "provider": provider})
}
