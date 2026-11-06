package http

import (
	"log"
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

type LeaveHandler struct {
	leaveUseCase usecase.LeaveUseCase
}

func NewLeaveHandler(leaveUseCase usecase.LeaveUseCase) *LeaveHandler {
	return &LeaveHandler{leaveUseCase}
}

func (h *LeaveHandler) SetupRoutes(router *gin.RouterGroup) {
	leaves := router.Group("/time-off")
	{
		leaves.POST("/request", h.SubmitLeave)
		leaves.GET("/balances", h.GetMyBalances)
		leaves.GET("/history", h.GetMyLeaves)
		leaves.PUT("/:id", h.UpdateMyLeave)
		leaves.DELETE("/:id", h.DeleteMyLeave)

		// Admin/Manager only
		managedRead := leaves.Group("/manage")
		managedRead.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager), string(domain.RoleAdminNota), string(domain.RoleSupervisorSales), string(domain.RoleSupervisorSalesMT)))
		{
			managedRead.GET("", h.GetAllLeaves)
		}

		managed := leaves.Group("/manage")
		managed.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager)))
		{
			managed.POST("", h.AdminCreateLeave)
			managed.PUT("/:id", h.AdminUpdateLeave)
			managed.DELETE("/:id", h.AdminDeleteLeave)
			managed.PUT("/:id/approve", h.ApproveLeave)
			managed.PUT("/balances", h.AdminUpdateBalance)
			managed.DELETE("/balances", h.AdminDeleteBalance)
		}
	}

	// Admin-only leave balances endpoint
	leaveBalances := router.Group("/leave-balances")
	leaveBalances.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager)))
	{
		leaveBalances.GET("", h.GetAllLeaveBalances)
	}
}

func (h *LeaveHandler) SubmitLeave(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	var req usecase.SubmitLeaveRequest
	// Try to bind from form first (for multipart)
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Handle file upload
	attachmentURL, err := h.handleFileUpload(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save attachment")
		return
	}
	if attachmentURL != "" {
		req.AttachmentURL = attachmentURL
	}

	err = h.leaveUseCase.SubmitLeave(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Leave request submitted successfully", nil)
}

func (h *LeaveHandler) GetMyLeaves(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	leaves, err := h.leaveUseCase.GetMyLeaves(userID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get leaves")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Leaves fetched", leaves)
}

func (h *LeaveHandler) GetMyBalances(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	balances, err := h.leaveUseCase.GetMyBalances(userID)
	if err != nil {
		log.Printf("Error in GetMyBalances: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get leave balances")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Leave balances fetched", balances)
}

func (h *LeaveHandler) GetAllLeaves(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))
	leaves, total, err := h.leaveUseCase.GetAllLeaves(status, search, page, limit, branchIDs)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get leaves")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Leaves fetched",
		"data":    leaves,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *LeaveHandler) ApproveLeave(c *gin.Context) {
	approverIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	approverID := approverIDStr.(uuid.UUID)

	leaveIDStr := c.Param("id")
	leaveID, err := uuid.Parse(leaveIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid leave ID format")
		return
	}

	var req usecase.ApproveLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.leaveUseCase.ApproveReturnLeave(leaveID, approverID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	message := "Leave status updated successfully"
	if req.Status == "REJECTED" {
		message = "Leave successfully rejected and balance refunded"
	}

	utils.SuccessResponse(c, http.StatusOK, message, nil)
}

func (h *LeaveHandler) GetAllLeaveBalances(c *gin.Context) {
	balances, err := h.leaveUseCase.GetAllLeaveBalances()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get all leave balances")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Leave balances fetched", balances)
}

func (h *LeaveHandler) AdminCreateLeave(c *gin.Context) {
	var req usecase.AdminLeaveRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Handle file upload
	attachmentURL, err := h.handleFileUpload(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save attachment")
		return
	}
	if attachmentURL != "" {
		req.AttachmentURL = attachmentURL
	}

	if err := h.leaveUseCase.AdminCreateLeave(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Leave request created successfully", nil)
}

func (h *LeaveHandler) AdminUpdateLeave(c *gin.Context) {
	idStr := c.Param("id")
	leaveID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid leave ID")
		return
	}

	var req usecase.AdminLeaveRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Handle file upload
	attachmentURL, err := h.handleFileUpload(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save attachment")
		return
	}
	if attachmentURL != "" {
		req.AttachmentURL = attachmentURL
	}

	if err := h.leaveUseCase.AdminUpdateLeave(leaveID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Leave request updated successfully", nil)
}

func (h *LeaveHandler) AdminDeleteLeave(c *gin.Context) {
	idStr := c.Param("id")
	leaveID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid leave ID")
		return
	}
	if err := h.leaveUseCase.AdminDeleteLeave(leaveID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete: "+err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Leave request deleted", nil)
}
func (h *LeaveHandler) UpdateMyLeave(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	leaveID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid leave ID")
		return
	}

	var req usecase.SubmitLeaveRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Handle file upload if any
	attachmentURL, err := h.handleFileUpload(c)
	if err == nil && attachmentURL != "" {
		req.AttachmentURL = attachmentURL
	}

	err = h.leaveUseCase.UpdateMyLeave(leaveID, userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Leave request updated successfully", nil)
}

func (h *LeaveHandler) DeleteMyLeave(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	leaveID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid leave ID")
		return
	}

	err = h.leaveUseCase.DeleteMyLeave(leaveID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Leave request cancelled/deleted successfully", nil)
}

func (h *LeaveHandler) AdminUpdateBalance(c *gin.Context) {
	var req usecase.AdminUpdateBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.leaveUseCase.AdminUpdateBalance(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Leave balance updated successfully", nil)
}

func (h *LeaveHandler) AdminDeleteBalance(c *gin.Context) {
	var req usecase.AdminUpdateBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := h.leaveUseCase.AdminDeleteBalance(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Leave balance deleted successfully", nil)
}

func (h *LeaveHandler) handleFileUpload(c *gin.Context) (string, error) {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return "", nil
	}

	file, err := c.FormFile("attachment")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		log.Println("Error in FormFile:", err)
		return "", err
	}

	filename := uuid.New().String() + "-" + filepath.Base(file.Filename)
	uploadDir := "uploads/attachments"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Println("Error in MkdirAll:", err)
		return "", err
	}
	savePath := uploadDir + "/" + filename
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		log.Println("Error in SaveUploadedFile:", err)
		return "", err
	}

	return "/uploads/attachments/" + filename, nil
}
