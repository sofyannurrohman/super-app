package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/repository"
)

type NotificationHandler struct {
	repo repository.NotificationRepository
}

func NewNotificationHandler(repo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repo}
}

func (h *NotificationHandler) RegisterRoutes(r *gin.RouterGroup) {
	notifications := r.Group("/notifications")
	{
		notifications.GET("", h.GetNotifications)
		notifications.POST("/:id/read", h.MarkAsRead)
	}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	val, _ := c.Get("companyID")
	if val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: company ID missing"})
		return
	}
	companyID := val.(uuid.UUID)

	branchID := uuid.Nil
	branchIDStr := c.Query("branch_id")
	if branchIDStr != "" {
		if bid, err := uuid.Parse(branchIDStr); err == nil {
			branchID = bid
		}
	}

	notifications, err := h.repo.GetByBranch(companyID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := uuid.MustParse(c.Param("id"))
	if err := h.repo.MarkAsRead(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}
