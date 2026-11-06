package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"example.com/hris/backend/internal/usecase"
)

type AnalyticsHandler struct {
	useCase usecase.AnalyticsUseCase
}

func NewAnalyticsHandler(useCase usecase.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{useCase: useCase}
}

type QueryRequest struct {
	Question string `json:"question" binding:"required"`
}

func (h *AnalyticsHandler) RegisterRoutes(r *gin.RouterGroup) {
	analytics := r.Group("/analytics")
	{
		analytics.POST("/query", h.Query)
	}
}

func (h *AnalyticsHandler) Query(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question parameter is required"})
		return
	}

	result, err := h.useCase.Query(req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
