package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type AuditLogHandler struct {
	usecase usecase.AuditLogUsecase
}

func NewAuditLogHandler(u usecase.AuditLogUsecase) *AuditLogHandler {
	return &AuditLogHandler{u}
}

func (h *AuditLogHandler) SetupRoutes(r *gin.RouterGroup) {
	logs := r.Group("/sales-marketing/audit-logs")
	{
		logs.GET("", h.GetAuditLogs)
	}

	financeLogs := r.Group("/finance/audit-logs")
	{
		financeLogs.GET("", h.GetAuditLogs)
	}
}

func (h *AuditLogHandler) GetAuditLogs(c *gin.Context) {
	val, exists := c.Get("companyID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	companyID := val.(uuid.UUID)
	entity := c.DefaultQuery("entity", "ALL")
	action := c.DefaultQuery("action", "ALL")
	search := c.Query("search")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, total, err := h.usecase.GetAuditLogs(companyID, entity, action, search, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Audit logs fetched successfully",
		"data":    logs,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
