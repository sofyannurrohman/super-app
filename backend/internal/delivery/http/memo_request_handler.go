package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

type MemoRequestHandler struct {
	useCase usecase.MemoRequestUseCase
}

func NewMemoRequestHandler(useCase usecase.MemoRequestUseCase) *MemoRequestHandler {
	return &MemoRequestHandler{useCase: useCase}
}

func (h *MemoRequestHandler) SetupRoutes(r *gin.RouterGroup) {
	memoRequests := r.Group("/memo-requests")
	{
		memoRequests.POST("", h.Create)
		memoRequests.GET("", h.GetAll)
		memoRequests.GET("/:id", h.GetByID)
		memoRequests.PUT("/:id/approve", h.Approve)
	}
}

func (h *MemoRequestHandler) Create(c *gin.Context) {
	var req usecase.CreateMemoRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.CreateMemoRequest(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Memo request created successfully"})
}

func (h *MemoRequestHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	branchIDStr := c.Query("branch_id")

	var branchID *uuid.UUID
	if branchIDStr != "" {
		parsed, err := uuid.Parse(branchIDStr)
		if err == nil {
			branchID = &parsed
		}
	}

	limit := 100
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil {
			limit = parsedLimit
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil {
			offset = parsedOffset
		}
	}

	memos, total, err := h.useCase.GetMemoRequests(status, branchID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": memos,
		"meta": gin.H{
			"total": total,
		},
	})
}

func (h *MemoRequestHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	memo, err := h.useCase.GetMemoRequestByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if memo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Memo request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": memo})
}

func (h *MemoRequestHandler) Approve(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Assuming approver ID is the user ID from context/token
	var approverID uuid.UUID
	if val, exists := c.Get("userID"); exists {
		if uid, ok := val.(uuid.UUID); ok {
			approverID = uid
		}
	}

	if err := h.useCase.ApproveMemoRequest(id, approverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Memo request approved successfully"})
}
