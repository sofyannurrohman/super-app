package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
)

type KasPabrikHandler struct {
	usecase usecase.KasPabrikUsecase
}

func NewKasPabrikHandler(u usecase.KasPabrikUsecase) *KasPabrikHandler {
	return &KasPabrikHandler{usecase: u}
}

func (h *KasPabrikHandler) RegisterRoutes(r *gin.RouterGroup) {
	kas := r.Group("/kas-pabrik")
	kas.Use(RoleMiddleware(string(domain.RoleSuperAdmin), string(domain.RoleHRAdmin), string(domain.RoleManager), string(domain.RoleCashierPusat), string(domain.RoleCashier), string(domain.RoleAdminKasir), string(domain.RoleAdminGudang)))
	{
		kas.GET("", h.ListKas)
		kas.POST("", h.CreateKas)
		kas.GET("/transactions", h.ListTransactions)
		kas.POST("/transfer", h.Transfer)
		kas.GET("/pending-deposits", h.GetPendingDeposits)
		kas.POST("/pending-deposits/:mutationId/approve", h.ApproveDeposit)
		kas.POST("/pending-deposits/:mutationId/reject", h.RejectDeposit)

		// Master Categories Routes
		kas.GET("/master/income-sources", h.ListIncomeSources)
		kas.GET("/master/income-sources/:id", h.GetIncomeSourceDetail)
		kas.POST("/master/income-sources", h.CreateIncomeSource)
		kas.PUT("/master/income-sources/:id", h.UpdateIncomeSource)
		kas.DELETE("/master/income-sources/:id", h.DeleteIncomeSource)

		kas.GET("/master/expense-types", h.ListExpenseTypes)
		kas.GET("/master/expense-types/:id", h.GetExpenseTypeDetail)
		kas.POST("/master/expense-types", h.CreateExpenseType)
		kas.PUT("/master/expense-types/:id", h.UpdateExpenseType)
		kas.DELETE("/master/expense-types/:id", h.DeleteExpenseType)

		kas.GET("/master/expense-accounts", h.ListExpenseAccounts)
		kas.GET("/master/expense-accounts/:id", h.GetExpenseAccountDetail)
		kas.POST("/master/expense-accounts", h.CreateExpenseAccount)
		kas.PUT("/master/expense-accounts/:id", h.UpdateExpenseAccount)
		kas.DELETE("/master/expense-accounts/:id", h.DeleteExpenseAccount)

		kas.GET("/:id", h.GetKasDetail)
		kas.PUT("/:id", h.UpdateKas)
		kas.DELETE("/:id", h.DeleteKas)
		kas.POST("/:id/transactions", h.RecordTransaction)
		kas.POST("/:id/recalculate", h.RecalculateLedger)
		kas.POST("/transactions/:txId/settle-advance", h.SettleAdvance)
		kas.PUT("/transactions/:txId", h.UpdateTransaction)
		kas.DELETE("/transactions/:txId", h.DeleteTransaction)
	}
}

func (h *KasPabrikHandler) ListKas(c *gin.Context) {
	branchID := c.Query("branch_id")
	kas, err := h.usecase.GetAllByBranch(branchID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", kas)
}

func (h *KasPabrikHandler) CreateKas(c *gin.Context) {
	var req domain.CreateKasPabrikRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	kas, err := h.usecase.Create(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Kas created successfully", kas)
}

func (h *KasPabrikHandler) GetKasDetail(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID KAS tidak valid")
		return
	}
	kas, err := h.usecase.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Kas not found")
		return
	}

	filter := domain.ListKasTransactionsFilter{
		KasPabrikID: id,
		Page:        1,
		Limit:       20,
	}
	transactions, _, _ := h.usecase.ListTransactions(filter)
	kas.Transactions = transactions

	utils.SuccessResponse(c, http.StatusOK, "Success", kas)
}

func (h *KasPabrikHandler) RecordTransaction(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID KAS tidak valid")
		return
	}
	var req domain.CreateKasTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	employeeID, _ := c.Get("employee_id")
	var empIDPtr *uuid.UUID
	if eid, ok := employeeID.(uuid.UUID); ok {
		empIDPtr = &eid
	}

	tx, err := h.usecase.RecordTransaction(id, req, empIDPtr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Transaction recorded successfully", tx)
}

func (h *KasPabrikHandler) Transfer(c *gin.Context) {
	var req domain.TransferKasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	employeeID, _ := c.Get("employee_id")
	var empIDPtr *uuid.UUID
	if eid, ok := employeeID.(uuid.UUID); ok {
		empIDPtr = &eid
	}

	if err := h.usecase.Transfer(req, empIDPtr); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transfer successful", nil)
}

func (h *KasPabrikHandler) ListTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limitStr := c.DefaultQuery("limit", "20")
	var limit int
	if limitStr == "all" || limitStr == "0" || limitStr == "-1" {
		limit = 0
	} else {
		limit, _ = strconv.Atoi(limitStr)
		if limit < 0 {
			limit = 20
		}
	}

	filter := domain.ListKasTransactionsFilter{
		BranchID:         c.Query("branch_id"),
		KasPabrikID:      c.Query("kas_pabrik_id"),
		Type:             c.Query("type"),
		DateFrom:         c.Query("date_from"),
		DateTo:           c.Query("date_to"),
		AdvanceStatus:    c.Query("advance_status"),
		TargetBranchID:   c.Query("target_branch_id"),
		EmployeeID:       c.Query("employee_id"),
		ExpenseTypeID:    c.Query("expense_type_id"),
		ExpenseAccountID: c.Query("expense_account_id"),
		IncomeSourceID:   c.Query("income_source_id"),
		Page:             page,
		Limit:            limit,
	}

	transactions, total, err := h.usecase.ListTransactions(filter)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"data":  transactions,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *KasPabrikHandler) SettleAdvance(c *gin.Context) {
	txID := c.Param("txId")
	var req domain.SettleKasAdvanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userIDVal, _ := c.Get("user_id")
	var userIDPtr *uuid.UUID
	if uid, ok := userIDVal.(uuid.UUID); ok {
		userIDPtr = &uid
	}

	settlement, err := h.usecase.SettleAdvance(txID, req, userIDPtr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Kasbon berhasil diselesaikan", settlement)
}

func (h *KasPabrikHandler) DeleteTransaction(c *gin.Context) {
	txID := c.Param("txId")
	if err := h.usecase.DeleteTransaction(txID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transaction deleted successfully", nil)
}

func (h *KasPabrikHandler) UpdateKas(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateKasPabrikRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	kas, err := h.usecase.Update(id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Kas updated successfully", kas)
}

func (h *KasPabrikHandler) DeleteKas(c *gin.Context) {
	id := c.Param("id")
	if err := h.usecase.Delete(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Kas deleted successfully", nil)
}

func (h *KasPabrikHandler) UpdateTransaction(c *gin.Context) {
	txID := c.Param("txId")
	var req domain.UpdateKasTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tx, err := h.usecase.UpdateTransaction(txID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Transaction updated successfully", tx)
}

func (h *KasPabrikHandler) GetPendingDeposits(c *gin.Context) {
	mutations, err := h.usecase.GetPendingDeposits()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", mutations)
}

func (h *KasPabrikHandler) ApproveDeposit(c *gin.Context) {
	mutationID := c.Param("mutationId")

	employeeID, _ := c.Get("employee_id")
	var empIDPtr *uuid.UUID
	if eid, ok := employeeID.(uuid.UUID); ok {
		empIDPtr = &eid
	}

	if err := h.usecase.ApproveDeposit(mutationID, empIDPtr); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Deposit approved successfully", nil)
}

func (h *KasPabrikHandler) RejectDeposit(c *gin.Context) {
	mutationID := c.Param("mutationId")

	employeeID, _ := c.Get("employee_id")
	var empIDPtr *uuid.UUID
	if eid, ok := employeeID.(uuid.UUID); ok {
		empIDPtr = &eid
	}

	if err := h.usecase.RejectDeposit(mutationID, empIDPtr); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Deposit rejected successfully", nil)
}

// Master Income Sources Handler
func (h *KasPabrikHandler) ListIncomeSources(c *gin.Context) {
	onlyActive := c.Query("active_only") == "true"
	list, err := h.usecase.ListIncomeSources(onlyActive)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", list)
}

func (h *KasPabrikHandler) GetIncomeSourceDetail(c *gin.Context) {
	id := c.Param("id")
	src, err := h.usecase.GetIncomeSourceByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Income source not found")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", src)
}

func (h *KasPabrikHandler) CreateIncomeSource(c *gin.Context) {
	var req domain.CreateKasIncomeSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	src, err := h.usecase.CreateIncomeSource(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Income source created successfully", src)
}

func (h *KasPabrikHandler) UpdateIncomeSource(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateKasIncomeSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	src, err := h.usecase.UpdateIncomeSource(id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Income source updated successfully", src)
}

func (h *KasPabrikHandler) DeleteIncomeSource(c *gin.Context) {
	id := c.Param("id")
	if err := h.usecase.DeleteIncomeSource(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Income source deleted successfully", nil)
}

// Master Expense Types Handler
func (h *KasPabrikHandler) ListExpenseTypes(c *gin.Context) {
	onlyActive := c.Query("active_only") == "true"
	list, err := h.usecase.ListExpenseTypes(onlyActive)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", list)
}

func (h *KasPabrikHandler) GetExpenseTypeDetail(c *gin.Context) {
	id := c.Param("id")
	et, err := h.usecase.GetExpenseTypeByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Expense type not found")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", et)
}

func (h *KasPabrikHandler) CreateExpenseType(c *gin.Context) {
	var req domain.CreateKasExpenseTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	et, err := h.usecase.CreateExpenseType(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Expense type created successfully", et)
}

func (h *KasPabrikHandler) UpdateExpenseType(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateKasExpenseTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	et, err := h.usecase.UpdateExpenseType(id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Expense type updated successfully", et)
}

func (h *KasPabrikHandler) DeleteExpenseType(c *gin.Context) {
	id := c.Param("id")
	if err := h.usecase.DeleteExpenseType(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Expense type deleted successfully", nil)
}

// Master Expense Accounts Handler
func (h *KasPabrikHandler) ListExpenseAccounts(c *gin.Context) {
	typeID := c.Query("expense_type_id")
	onlyActive := c.Query("active_only") == "true"
	list, err := h.usecase.ListExpenseAccounts(typeID, onlyActive)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", list)
}

func (h *KasPabrikHandler) GetExpenseAccountDetail(c *gin.Context) {
	id := c.Param("id")
	acc, err := h.usecase.GetExpenseAccountByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Expense account not found")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", acc)
}

func (h *KasPabrikHandler) CreateExpenseAccount(c *gin.Context) {
	var req domain.CreateKasExpenseAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	acc, err := h.usecase.CreateExpenseAccount(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Expense account created successfully", acc)
}

func (h *KasPabrikHandler) UpdateExpenseAccount(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateKasExpenseAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	acc, err := h.usecase.UpdateExpenseAccount(id, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Expense account updated successfully", acc)
}

func (h *KasPabrikHandler) DeleteExpenseAccount(c *gin.Context) {
	id := c.Param("id")
	if err := h.usecase.DeleteExpenseAccount(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Expense account deleted successfully", nil)
}

func (h *KasPabrikHandler) RecalculateLedger(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID KAS tidak valid")
		return
	}
	if err := h.usecase.RecalculateKasLedger(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghitung ulang buku kas: "+err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Buku kas dan saldo berhasil dihitung ulang dan disinkronkan", nil)
}

