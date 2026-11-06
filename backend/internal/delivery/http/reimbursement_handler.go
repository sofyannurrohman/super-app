package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type ReimbursementHandler struct {
	useCase usecase.ReimbursementUseCase
}

func NewReimbursementHandler(useCase usecase.ReimbursementUseCase) *ReimbursementHandler {
	return &ReimbursementHandler{useCase}
}

func (h *ReimbursementHandler) SetupRoutes(router *gin.RouterGroup) {
	reimbursements := router.Group("/reimbursements")
	{
		reimbursements.POST("", h.Submit)
		reimbursements.GET("/my", h.GetHistory)
		reimbursements.PUT("/:id", h.Update)
		reimbursements.DELETE("/:id", h.Delete)

		// Admin only
		manage := reimbursements.Group("/manage")
		manage.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager), string(domain.RoleCashier), string(domain.RoleCashierPusat), string(domain.RoleAdminKasir)))
		{
			manage.GET("", h.GetAll)
			manage.POST("", h.AdminCreate)
			manage.PUT("/:id", h.AdminUpdate)
			manage.DELETE("/:id", h.AdminDelete)
			manage.PUT("/:id/approve", h.Approve)
			manage.PUT("/:id/adjust-banner-rate", h.AdjustBannerRate)
			manage.POST("/:id/disburse", h.Disburse)
		}
	}
}

func (h *ReimbursementHandler) Submit(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req usecase.SubmitReimbursementRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Handle file upload
	attachmentURL, err := h.handleFileUpload(c)
	if err != nil {
		if err.Error() == "file size too large" {
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "File size too large (max 100MB)")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save attachment: "+err.Error())
		return
	}
	if attachmentURL != "" {
		req.AttachmentURL = attachmentURL
	}

	if startKMPhoto, _ := h.handleNamedFileUpload(c, "start_km_photo"); startKMPhoto != "" {
		req.StartKMPhotoURL = startKMPhoto
	}
	if endKMPhoto, _ := h.handleNamedFileUpload(c, "end_km_photo"); endKMPhoto != "" {
		req.EndKMPhotoURL = endKMPhoto
	}

	err = h.useCase.Submit(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Reimbursement submitted successfully", nil)
}

func (h *ReimbursementHandler) GetHistory(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reimbursements, err := h.useCase.GetHistory(userID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reimbursements fetched", reimbursements)
}

func (h *ReimbursementHandler) GetAll(c *gin.Context) {
	status := c.Query("status")
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Optional branch filter
	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))

	var reimbursements []domain.Reimbursement
	var err error
	if len(branchIDs) > 0 {
		reimbursements, err = h.useCase.GetAllByBranch(status, category, branchIDs, page, limit)
	} else {
		reimbursements, err = h.useCase.GetAll(status, category, page, limit)
	}
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "All reimbursements fetched", reimbursements)
}

func (h *ReimbursementHandler) Approve(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id format")
		return
	}

	var req usecase.ApproveReimbursementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.useCase.Approve(userID, id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reimbursement status updated", nil)
}

func (h *ReimbursementHandler) AdjustBannerRate(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id format")
		return
	}

	var req usecase.AdjustBannerRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.useCase.AdjustBannerRate(userID, id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tarif reimbursement spanduk berhasil diperbarui", nil)
}

func (h *ReimbursementHandler) Disburse(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id format")
		return
	}

	var req usecase.DisburseReimbursementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.useCase.Disburse(userID, id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reimbursement berhasil dicairkan", nil)
}

func (h *ReimbursementHandler) Update(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id format")
		return
	}

	var req usecase.SubmitReimbursementRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Handle file upload if any
	attachmentURL, err := h.handleFileUpload(c)
	if err != nil && err != http.ErrMissingFile {
		if err.Error() == "file size too large" {
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "File size too large (max 100MB)")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save attachment: "+err.Error())
		return
	}
	if attachmentURL != "" {
		req.AttachmentURL = attachmentURL
	}

	if startKMPhoto, _ := h.handleNamedFileUpload(c, "start_km_photo"); startKMPhoto != "" {
		req.StartKMPhotoURL = startKMPhoto
	}
	if endKMPhoto, _ := h.handleNamedFileUpload(c, "end_km_photo"); endKMPhoto != "" {
		req.EndKMPhotoURL = endKMPhoto
	}

	err = h.useCase.UpdateMyReimbursement(userID, id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reimbursement updated successfully", nil)
}

func (h *ReimbursementHandler) Delete(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id format")
		return
	}

	err = h.useCase.DeleteMyReimbursement(userID, id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reimbursement deleted successfully", nil)
}

func (h *ReimbursementHandler) AdminCreate(c *gin.Context) {
	var req usecase.AdminReimbursementRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Parse banner_ids if provided as form parameter
	if len(req.BannerIDs) == 0 {
		if rawIDs := c.PostForm("banner_ids"); rawIDs != "" {
			var parsed []string
			if err := json.Unmarshal([]byte(rawIDs), &parsed); err == nil {
				req.BannerIDs = parsed
			} else {
				for _, part := range strings.Split(rawIDs, ",") {
					if trimmed := strings.TrimSpace(part); trimmed != "" {
						req.BannerIDs = append(req.BannerIDs, trimmed)
					}
				}
			}
		}
	}

	// Handle file upload
	attachmentURL, err := h.handleFileUpload(c)
	if err != nil {
		if err.Error() == "file size too large" {
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "File size too large (max 100MB)")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save attachment: "+err.Error())
		return
	}
	if attachmentURL != "" {
		req.AttachmentURL = attachmentURL
	}

	if startKMPhoto, _ := h.handleNamedFileUpload(c, "start_km_photo"); startKMPhoto != "" {
		req.StartKMPhotoURL = startKMPhoto
	}
	if endKMPhoto, _ := h.handleNamedFileUpload(c, "end_km_photo"); endKMPhoto != "" {
		req.EndKMPhotoURL = endKMPhoto
	}

	err = h.useCase.AdminCreate(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Reimbursement created successfully", nil)
}

func (h *ReimbursementHandler) AdminUpdate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id format")
		return
	}

	var req usecase.AdminReimbursementRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Handle file upload
	attachmentURL, err := h.handleFileUpload(c)
	if err != nil {
		if err.Error() == "file size too large" {
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "File size too large (max 100MB)")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save attachment: "+err.Error())
		return
	}
	if attachmentURL != "" {
		req.AttachmentURL = attachmentURL
	}

	if startKMPhoto, _ := h.handleNamedFileUpload(c, "start_km_photo"); startKMPhoto != "" {
		req.StartKMPhotoURL = startKMPhoto
	}
	if endKMPhoto, _ := h.handleNamedFileUpload(c, "end_km_photo"); endKMPhoto != "" {
		req.EndKMPhotoURL = endKMPhoto
	}

	err = h.useCase.AdminUpdate(id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reimbursement updated successfully", nil)
}

func (h *ReimbursementHandler) AdminDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid id format")
		return
	}

	err = h.useCase.AdminDelete(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reimbursement deleted successfully", nil)
}

func (h *ReimbursementHandler) handleNamedFileUpload(c *gin.Context, fieldName string) (string, error) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return "", nil
	}

	file, err := c.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}

	// Check file size (100MB limit)
	if file.Size > 100*1024*1024 {
		return "", fmt.Errorf("file size too large")
	}

	filename := uuid.New().String() + "-" + filepath.Base(file.Filename)
	uploadDir := "uploads/attachments"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}
	savePath := uploadDir + "/" + filename
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		return "", err
	}

	return "/uploads/attachments/" + filename, nil
}

func (h *ReimbursementHandler) handleFileUpload(c *gin.Context) (string, error) {
	res, err := h.handleNamedFileUpload(c, "attachment")
	if err == nil && res == "" {
		// Fallback check "file" field name
		res, err = h.handleNamedFileUpload(c, "file")
	}
	return res, err
}
