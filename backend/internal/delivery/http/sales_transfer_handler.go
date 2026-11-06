package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type SalesTransferHandler struct {
	usecase      usecase.SalesTransferUsecase
	salesUsecase usecase.SalesUsecase
}

func NewSalesTransferHandler(u usecase.SalesTransferUsecase, su usecase.SalesUsecase) *SalesTransferHandler {
	return &SalesTransferHandler{u, su}
}

func (h *SalesTransferHandler) CreateTransfer(c *gin.Context) {
	branchIDStr := c.Param("branchId")
	var branchID uuid.UUID
	if branchIDStr != "ALL_BRANCHES" && branchIDStr != "all" {
		var err error
		branchID, err = uuid.Parse(branchIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid branch id")
			return
		}
	}

	var req usecase.CreateSalesTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.CreateTransfer(branchID, req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "transfer created successfully", nil)
}

func (h *SalesTransferHandler) CreateBatchTransfer(c *gin.Context) {
	branchIDStr := c.Param("branchId")
	var branchID uuid.UUID
	if branchIDStr != "" && branchIDStr != "ALL_BRANCHES" && branchIDStr != "all" {
		var err error
		branchID, err = uuid.Parse(branchIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid branch id")
			return
		}
	}

	var req usecase.CreateBatchSalesTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.BranchID != uuid.Nil && branchID == uuid.Nil {
		branchID = req.BranchID
	}

	if err := h.usecase.CreateBatchTransfer(branchID, req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "batch transfer created successfully", nil)
}

func (h *SalesTransferHandler) GenerateDO(c *gin.Context) {
	var req usecase.GenerateDORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.GenerateDOForTransfers(req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Surat Jalan & DO berhasil dibuat", nil)
}

func (h *SalesTransferHandler) CompleteTransfer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.usecase.CompleteTransfer(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "transfer completed successfully", nil)
}

func (h *SalesTransferHandler) CancelTransfer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.usecase.CancelTransfer(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "transfer cancelled successfully", nil)
}

func (h *SalesTransferHandler) GetTransfers(c *gin.Context) {
	branchIDStr := c.Param("branchId")

	var transfers []domain.SalesTransfer
	var err error

	if branchIDStr == "ALL_BRANCHES" || branchIDStr == "all" {
		companyIDStr, ok := c.Get("company_id")
		if !ok {
			utils.ErrorResponse(c, http.StatusBadRequest, "company id not found in context")
			return
		}
		companyID := uuid.MustParse(companyIDStr.(string))
		transfers, err = h.usecase.GetTransfersByCompany(companyID)
	} else {
		branchID, errParse := uuid.Parse(branchIDStr)
		if errParse != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid branch id")
			return
		}
		transfers, err = h.usecase.GetTransfersByBranch(branchID)
	}

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "transfers fetched successfully", transfers)
}

func (h *SalesTransferHandler) GetSalesStock(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid employee id")
		return
	}

	stocks, err := h.usecase.GetSalesStock(employeeID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "stocks fetched successfully", stocks)
}

func (h *SalesTransferHandler) DeleteTransfer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.usecase.DeleteTransfer(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "transfer deleted successfully", nil)
}

func (h *SalesTransferHandler) CreateRequest(c *gin.Context) {
	var req usecase.CreateSalesTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Resolve EmployeeID from token if not provided
	if req.EmployeeID == uuid.Nil {
		if employeeIDStr, ok := c.Get("employee_id"); ok && employeeIDStr != "" {
			req.EmployeeID = uuid.MustParse(employeeIDStr.(string))
		}
	}

	if req.EmployeeID == uuid.Nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "employee_id is required")
		return
	}

	// Parse branch ID from context if available
	var branchID uuid.UUID
	if branchIDStr, ok := c.Get("branch_id"); ok && branchIDStr != "" {
		branchID = uuid.MustParse(branchIDStr.(string))
	}

	// For mobile requests, if it's outbound (TRANSFER / Warehouse -> Salesman), set initial status to WAITING_SUPERVISOR
	if req.Type == domain.SalesTransferOut {
		req.Status = domain.SalesTransferWaitingSupervisor
	}

	if err := h.usecase.CreateTransfer(branchID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Request stok berhasil dikirim", nil)
}

// CreateDistributorRequest keeps distributor requests distinct from regular
// salesman cargo requests while reusing the existing stock reservation flow.
func (h *SalesTransferHandler) CreateDistributorRequest(c *gin.Context) {
	var req usecase.CreateSalesTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.DistributorStoreID == nil || *req.DistributorStoreID == uuid.Nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "distributor_store_id is required")
		return
	}
	if req.EmployeeID == uuid.Nil {
		if employeeID, ok := c.Get("employee_id"); ok && employeeID != "" {
			req.EmployeeID = uuid.MustParse(employeeID.(string))
		}
	}
	if req.EmployeeID == uuid.Nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "employee_id is required")
		return
	}
	req.RequestSource = "DISTRIBUTOR"
	req.Status = domain.SalesTransferWaitingSupervisor
	var branchID uuid.UUID
	if branch, ok := c.Get("branch_id"); ok && branch != "" {
		branchID = uuid.MustParse(branch.(string))
	}
	if err := h.usecase.CreateTransfer(branchID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Request distributor berhasil dikirim", nil)
}

func (h *SalesTransferHandler) GetPendingRequests(c *gin.Context) {
	branchIDStr, okBranch := c.Get("branch_id")
	companyIDStr, okCompany := c.Get("company_id")

	var transfers []domain.SalesTransfer
	var err error

	if okBranch {
		branchID := uuid.MustParse(branchIDStr.(string))
		transfers, err = h.usecase.GetTransfersByBranch(branchID)
	} else if okCompany {
		companyID := uuid.MustParse(companyIDStr.(string))
		transfers, err = h.usecase.GetTransfersByCompany(companyID)
	} else {
		utils.ErrorResponse(c, http.StatusBadRequest, "No branch or company context found")
		return
	}

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Filter only pending
	var pending []domain.SalesTransfer
	for _, t := range transfers {
		if t.Status == domain.SalesTransferPending {
			pending = append(pending, t)
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Pending transfers fetched", pending)
}

func (h *SalesTransferHandler) GetMyDistributorRequests(c *gin.Context) {
	employeeIDValue, ok := c.Get("employee_id")
	if !ok || employeeIDValue == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "employee context not found")
		return
	}
	employeeID, err := uuid.Parse(employeeIDValue.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid employee context")
		return
	}
	transfers, err := h.usecase.GetTransfersByEmployee(employeeID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]domain.SalesTransfer, 0)
	for _, transfer := range transfers {
		if transfer.RequestSource == "DISTRIBUTOR" {
			result = append(result, transfer)
		}
	}
	utils.SuccessResponse(c, http.StatusOK, "Distributor requests fetched", result)
}

func (h *SalesTransferHandler) RegisterRoutes(r *gin.RouterGroup) {
	transfers := r.Group("/sales-transfers")
	{
		transfers.POST("/branch/:branchId", h.CreateTransfer)
		transfers.POST("/branch/:branchId/batch", h.CreateBatchTransfer)
		transfers.POST("/batch", h.CreateBatchTransfer)
		transfers.GET("/branch/:branchId", h.GetTransfers)
		transfers.PUT("/:id", h.UpdateTransfer)
		transfers.PATCH("/:id/complete", h.CompleteTransfer)
		transfers.PATCH("/:id/cancel", h.CancelTransfer)
		transfers.DELETE("/:id", h.DeleteTransfer)
		transfers.POST("/generate-do", h.GenerateDO)
		transfers.GET("/stock/:employeeId", h.GetSalesStock)
		transfers.GET("/history/:employeeId", h.GetEmployeeHistory)
		transfers.POST("/reconcile/branch/:branchId", h.ReconcileStock)
		transfers.POST("/adjust", h.AdjustSalesStock)

		// Mobile Optimized
		transfers.POST("/request", h.CreateRequest)
		transfers.POST("/distributor/request", h.CreateDistributorRequest)
		transfers.GET("/pending", h.GetPendingRequests)
		transfers.GET("/distributor/my-requests", h.GetMyDistributorRequests)

		// Supervisor Approval Routes
		approvalRoles := RoleMiddleware(
			string(domain.RoleSuperAdmin),
			string(domain.RoleSupervisorGudang),
			string(domain.RoleSupervisorSales),
			string(domain.RoleSupervisorSalesMT),
			string(domain.RoleManager),
		)
		transfers.GET("/supervisor/pending", approvalRoles, h.GetSupervisorPendingRequests)
		transfers.PATCH("/:id/approve-supervisor", approvalRoles, h.ApproveSupervisorTransfer)
		transfers.PATCH("/:id/reject-supervisor", approvalRoles, h.RejectSupervisorTransfer)

		// Attendance / Visit
		transfers.POST("/attendance", h.RecordAttendance)
	}
}

func (h *SalesTransferHandler) RecordAttendance(c *gin.Context) {
	var req usecase.RecordVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.salesUsecase.RecordVisit(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visit recorded successfully"})
}

func (h *SalesTransferHandler) GetEmployeeHistory(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("employeeId"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid employee id")
		return
	}
	transfers, err := h.usecase.GetTransfersByEmployee(employeeID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "transfers fetched successfully", transfers)
}

func (h *SalesTransferHandler) GetSupervisorPendingRequests(c *gin.Context) {
	branchIDQuery := c.Query("branch_id")

	var transfers []domain.SalesTransfer
	var err error

	if branchIDQuery != "" && branchIDQuery != "ALL_BRANCHES" {
		branchID, errParse := uuid.Parse(branchIDQuery)
		if errParse == nil {
			transfers, err = h.usecase.GetTransfersByBranch(branchID)
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid branch id in query")
			return
		}
	} else if branchIDQuery == "ALL_BRANCHES" {
		companyIDStr, okCompany := c.Get("company_id")
		if okCompany {
			companyID := uuid.MustParse(companyIDStr.(string))
			transfers, err = h.usecase.GetTransfersByCompany(companyID)
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "No company context found for ALL_BRANCHES")
			return
		}
	} else {
		branchIDStr, okBranch := c.Get("branch_id")
		companyIDStr, okCompany := c.Get("company_id")

		if okBranch && branchIDStr != "" {
			branchID := uuid.MustParse(branchIDStr.(string))
			transfers, err = h.usecase.GetTransfersByBranch(branchID)
		} else if okCompany {
			companyID := uuid.MustParse(companyIDStr.(string))
			transfers, err = h.usecase.GetTransfersByCompany(companyID)
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "No branch or company context found")
			return
		}
	}

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Filter only WAITING_SUPERVISOR
	var pending []domain.SalesTransfer
	sourceFilter := c.Query("request_source")
	for _, t := range transfers {
		if t.Status == domain.SalesTransferWaitingSupervisor && (sourceFilter == "" || t.RequestSource == sourceFilter) {
			pending = append(pending, t)
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "Supervisor pending transfers fetched", pending)
}

func (h *SalesTransferHandler) ApproveSupervisorTransfer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}
	_ = c.ShouldBindJSON(&req) // Ignore binding error if body is empty or not provided

	if err := h.usecase.ApproveSupervisorTransfer(id, req.Quantity); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "transfer approved by supervisor successfully", nil)
}

func (h *SalesTransferHandler) RejectSupervisorTransfer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.usecase.RejectSupervisorTransfer(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "transfer rejected by supervisor successfully", nil)
}

func (h *SalesTransferHandler) ReconcileStock(c *gin.Context) {
	branchIDStr := c.Param("branchId")
	var branchID uuid.UUID
	var err error
	if branchIDStr != "ALL_BRANCHES" && branchIDStr != "all" {
		branchID, err = uuid.Parse(branchIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "invalid branch id")
			return
		}
	}

	var req struct {
		EmployeeID uuid.UUID                    `json:"employee_id" binding:"required"`
		Items      []usecase.ReconciliationItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var userID uuid.UUID
	if userIDStr, exists := c.Get("userID"); exists {
		userID = userIDStr.(uuid.UUID)
	}

	reconcileReq := usecase.SalesStockReconciliationRequest{
		Items: req.Items,
	}

	if err := h.usecase.ReconcileSalesStock(req.EmployeeID, branchID, reconcileReq, userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sales stock reconciled successfully", nil)
}

func (h *SalesTransferHandler) UpdateTransfer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req usecase.UpdateSalesTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	correctedBy := "Admin Gudang"
	if val, ok := c.Get("username"); ok && val != "" {
		correctedBy = val.(string)
	} else if val, ok := c.Get("name"); ok && val != "" {
		correctedBy = val.(string)
	}
	req.CorrectedBy = correctedBy

	if err := h.usecase.UpdateTransfer(id, req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "transfer updated/corrected successfully", nil)
}

func (h *SalesTransferHandler) AdjustSalesStock(c *gin.Context) {
	var req struct {
		EmployeeID uuid.UUID                      `json:"employee_id" binding:"required"`
		Items      []usecase.AdjustSalesStockItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.usecase.AdjustSalesStock(req.EmployeeID, req.Items); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sales stock adjusted successfully", nil)
}
