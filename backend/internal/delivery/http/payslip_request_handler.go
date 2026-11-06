package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type PayslipRequestHandler struct {
	usecase usecase.PayslipRequestUsecase
}

func NewPayslipRequestHandler(usecase usecase.PayslipRequestUsecase) *PayslipRequestHandler {
	return &PayslipRequestHandler{usecase: usecase}
}

func (h *PayslipRequestHandler) SetupRoutes(router *gin.RouterGroup) {
	reqs := router.Group("/payslip-requests")
	{
		reqs.POST("", h.CreateRequest)
		reqs.GET("", h.GetMyRequests)
	}

	adminReqs := router.Group("/admin/payslip-requests")
	adminReqs.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin)))
	{
		adminReqs.GET("", h.GetAllRequests)
		adminReqs.GET("/:id/lookup", h.LookupPayrollData)
		adminReqs.PUT("/:id/approve", h.ApproveRequest)
		adminReqs.PUT("/:id/reject", h.RejectRequest)
	}
}

func (h *PayslipRequestHandler) CreateRequest(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req struct {
		PeriodMonth int     `json:"period_month"`
		PeriodYear  int     `json:"period_year"`
		Reason      *string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.CreateRequest(userID, req.PeriodMonth, req.PeriodYear, req.Reason); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Permohonan slip gaji berhasil dikirim", nil)
}

func (h *PayslipRequestHandler) GetMyRequests(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	reqs, total, err := h.usecase.GetMyRequests(userID, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reqs,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *PayslipRequestHandler) GetAllRequests(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit
	status := c.Query("status")

	reqs, total, err := h.usecase.GetAllRequests(status, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reqs,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *PayslipRequestHandler) LookupPayrollData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	result, err := h.usecase.LookupPayrollData(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data lookup payroll berhasil dihitung", result)
}

func (h *PayslipRequestHandler) ApproveRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var approvedByPtr *uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		uid := uidVal.(uuid.UUID)
		approvedByPtr = &uid
	}

	nominalStr := c.PostForm("nominal")
	var nominalPtr *float64
	if nominalStr != "" {
		nominal, err := strconv.ParseFloat(nominalStr, 64)
		if err == nil {
			nominalPtr = &nominal
		}
	}

	payslipDataStr := c.PostForm("payslip_data")
	var payslipDataPtr *string
	if payslipDataStr != "" {
		payslipDataPtr = &payslipDataStr
	}

	var fileURL *string
	file, err := c.FormFile("file")
	if err == nil {
		filename := uuid.New().String() + "-" + filepath.Base(file.Filename)
		uploadDir := "uploads/payslips"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat direktori unggahan")
			return
		}
		savePath := uploadDir + "/" + filename
		if err := utils.SaveUploadedFile(file, savePath); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan berkas")
			return
		}
		url := "/" + savePath
		fileURL = &url
	} else if err != http.ErrMissingFile {
		utils.ErrorResponse(c, http.StatusBadRequest, "Gagal membaca berkas unggahan")
		return
	}

	if nominalPtr == nil && fileURL == nil && payslipDataPtr == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Harap masukkan nominal, berkas, atau rincian slip gaji")
		return
	}

	if err := h.usecase.ApproveRequest(id, nominalPtr, fileURL, payslipDataPtr, approvedByPtr); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Permohonan slip gaji berhasil disetujui & diterbitkan", nil)
}

func (h *PayslipRequestHandler) RejectRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.RejectRequest(id, req.Reason); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Permohonan slip gaji berhasil ditolak", nil)
}
