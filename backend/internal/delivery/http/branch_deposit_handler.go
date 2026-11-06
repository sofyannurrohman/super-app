package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type BranchDepositHandler struct {
	usecase usecase.BranchDepositUsecase
}

func NewBranchDepositHandler(u usecase.BranchDepositUsecase) *BranchDepositHandler {
	return &BranchDepositHandler{usecase: u}
}

func (h *BranchDepositHandler) RegisterRoutes(rg *gin.RouterGroup) {
	api := rg.Group("/branch-deposits")
	{
		api.POST("", h.SubmitDeposit)
		api.GET("", h.GetDeposits)
		api.GET("/daily-balance", h.GetDailyBalance)
		api.POST("/:id/approve", h.ApproveDeposit)
		api.POST("/:id/reject", h.RejectDeposit)
	}
}

func (h *BranchDepositHandler) SubmitDeposit(c *gin.Context) {
	var req usecase.SubmitDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "user id tidak ditemukan")
		return
	}
	submittedBy, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid user id")
		return
	}

	deposit, err := h.usecase.SubmitDeposit(c.Request.Context(), req, submittedBy)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Setoran berhasil dibuat", deposit)
}

func (h *BranchDepositHandler) GetDeposits(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	status := c.Query("status")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var branchID *uuid.UUID
	if branchIDStr != "" {
		id, err := uuid.Parse(branchIDStr)
		if err == nil {
			branchID = &id
		}
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		t, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		t, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			// adjust end date to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &t
		}
	}

	deposits, err := h.usecase.GetDeposits(c.Request.Context(), branchID, statusPtr, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data setoran", deposits)
}

func (h *BranchDepositHandler) GetDailyBalance(c *gin.Context) {
	branchIDStr := c.Query("branch_id")
	if branchIDStr == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "branch_id is required")
		return
	}
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid branch_id")
		return
	}

	dateStr := c.Query("date")
	date := time.Now()
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			date = parsedDate
		}
	}

	totalAdminNota, err := h.usecase.GetDailyBalance(c.Request.Context(), branchID, date)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil daily balance", gin.H{
		"total_admin_nota": totalAdminNota,
	})
}

func (h *BranchDepositHandler) ApproveDeposit(c *gin.Context) {
	depositIDStr := c.Param("id")
	depositID, err := uuid.Parse(depositIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID setoran tidak valid")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "user id tidak ditemukan")
		return
	}
	approvedBy, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid user id")
		return
	}

	err = h.usecase.ApproveDeposit(c.Request.Context(), depositID, approvedBy)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Setoran berhasil disetujui", nil)
}

func (h *BranchDepositHandler) RejectDeposit(c *gin.Context) {
	depositIDStr := c.Param("id")
	depositID, err := uuid.Parse(depositIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID setoran tidak valid")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "user id tidak ditemukan")
		return
	}
	approvedBy, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "invalid user id")
		return
	}

	err = h.usecase.RejectDeposit(c.Request.Context(), depositID, approvedBy)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Setoran berhasil ditolak", nil)
}
