package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/usecase"
)

func optionalInvoiceDepositUUID(raw string) (*uuid.UUID, error) {
	if raw == "" || raw == "ALL_BRANCHES" || raw == "ALL" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (h *FinanceHandler) invoiceDepositBranchScope(c *gin.Context, requested *uuid.UUID) (*uuid.UUID, error) {
	userID, _ := c.Get("userID")
	id, _ := userID.(uuid.UUID)
	role, _ := c.Get("role")
	roleName, _ := role.(string)
	branchID, err := h.usecase.GetInvoiceDepositBranch(id, roleName)
	if err != nil {
		return nil, err
	}
	if branchID != nil {
		return branchID, nil
	}
	return requested, nil
}

func (h *FinanceHandler) GetAvailableDepositInvoices(c *gin.Context) {
	branchID, err := optionalInvoiceDepositUUID(c.Query("branch_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cabang tidak valid"})
		return
	}
	branchID, err = h.invoiceDepositBranchScope(c, branchID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	companyID, err := optionalInvoiceDepositUUID(c.Query("company_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID PT tidak valid"})
		return
	}
	items, err := h.usecase.GetAvailableDepositInvoices(branchID, companyID, c.Query("search"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *FinanceHandler) ListInvoiceDeposits(c *gin.Context) {
	branchID, err := optionalInvoiceDepositUUID(c.Query("branch_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cabang tidak valid"})
		return
	}
	branchID, err = h.invoiceDepositBranchScope(c, branchID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	items, err := h.usecase.ListInvoiceDeposits(branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *FinanceHandler) SubmitInvoiceDeposit(c *gin.Context) {
	var body struct {
		BranchID      string                             `json:"branch_id"`
		BankAccountID string                             `json:"bank_account_id"`
		DepositDate   string                             `json:"deposit_date"`
		ProofURL      string                             `json:"proof_url"`
		Notes         string                             `json:"notes"`
		Allocations   []usecase.InvoiceDepositAllocation `json:"allocations"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	branchID, branchErr := uuid.Parse(body.BranchID)
	bankID, bankErr := uuid.Parse(body.BankAccountID)
	depositDate, dateErr := time.Parse("2006-01-02", body.DepositDate)
	userID, userOK := c.Get("userID")
	submittedBy, idOK := userID.(uuid.UUID)
	if branchErr != nil || bankErr != nil || dateErr != nil || !userOK || !idOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cabang, rekening, tanggal, atau pengguna tidak valid"})
		return
	}
	role, _ := c.Get("role")
	roleName, _ := role.(string)
	deposit, err := h.usecase.SubmitInvoiceDeposit(usecase.SubmitInvoiceDepositRequest{
		BranchID: branchID, BankAccountID: bankID, DepositDate: depositDate,
		ProofURL: body.ProofURL, Notes: body.Notes, Allocations: body.Allocations,
	}, submittedBy, roleName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Setoran faktur menunggu verifikasi Kasir Pusat", "data": deposit})
}

func (h *FinanceHandler) reviewInvoiceDeposit(c *gin.Context, approve bool) {
	id, err := uuid.Parse(c.Param("id"))
	userID, userOK := c.Get("userID")
	reviewedBy, idOK := userID.(uuid.UUID)
	if err != nil || !userOK || !idOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID setoran atau pengguna tidak valid"})
		return
	}
	if err := h.usecase.ReviewInvoiceDeposit(id, reviewedBy, approve); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Setoran berhasil diproses"})
}

func (h *FinanceHandler) ApproveInvoiceDeposit(c *gin.Context) { h.reviewInvoiceDeposit(c, true) }
func (h *FinanceHandler) RejectInvoiceDeposit(c *gin.Context)  { h.reviewInvoiceDeposit(c, false) }
