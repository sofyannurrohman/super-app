package http

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CashBankMutationHandler struct {
	db *gorm.DB
}

func NewCashBankMutationHandler(db *gorm.DB) *CashBankMutationHandler {
	return &CashBankMutationHandler{db}
}

func (h *CashBankMutationHandler) SetupRoutes(r *gin.RouterGroup) {
	cbm := r.Group("/cashbank-mutations")
	{
		cbm.GET("", h.ListMutations)
		cbm.GET("/available-invoice-payments", h.GetAvailableInvoicePayments)
		cbm.POST("/:id/approve", h.ApproveFactoryInvoiceDeposit)
		cbm.POST("/:id/reject", h.RejectFactoryInvoiceDeposit)
		cbm.POST("", h.CreateMutation)
		cbm.PUT("/:id", h.UpdateMutation)
		cbm.DELETE("/:id", h.DeleteMutation)
		cbm.GET("/cash-balance", h.GetCashBalance)
		cbm.PUT("/branches/:id/petty-cash-balance", h.SetPettyCashBalance)
		cbm.PUT("/branches/:id/cash-balance", h.SetCashBalance)
	}
}

type availableInvoicePayment struct {
	PaymentID       uuid.UUID `json:"payment_id"`
	InvoiceID       uuid.UUID `json:"invoice_id"`
	ReceiptNo       string    `json:"receipt_no"`
	BuyerName       string    `json:"buyer_name"`
	SalesmanName    string    `json:"salesman_name"`
	PaymentDate     time.Time `json:"payment_date"`
	TransactionDate time.Time `json:"transaction_date"`
	PaymentAmount   float64   `json:"payment_amount"`
	Available       float64   `json:"available_amount"`
	CompanyID       uuid.UUID `json:"company_id"`
}

func isApprovedInvoicePaymentStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SUCCESS", "PAID", "APPROVED", "CASHIER_APPROVED":
		return true
	default:
		return false
	}
}

func cashierDepositBranch(employee domain.Employee, requested string) (uuid.UUID, error) {
	if requested != "" {
		branchID, err := uuid.Parse(requested)
		if err != nil {
			return uuid.Nil, errors.New("ID cabang setoran tidak valid")
		}
		if branchID != uuid.Nil && ((employee.BranchID != nil && *employee.BranchID == branchID) || (employee.OperatingBranchID != nil && *employee.OperatingBranchID == branchID)) {
			return branchID, nil
		}
		return uuid.Nil, errors.New("setoran hanya dapat dicatat untuk cabang utama atau cabang operasional Anda")
	}
	if employee.BranchID != nil && *employee.BranchID != uuid.Nil {
		return *employee.BranchID, nil
	}
	if employee.OperatingBranchID != nil && *employee.OperatingBranchID != uuid.Nil {
		return *employee.OperatingBranchID, nil
	}
	return uuid.Nil, errors.New("cabang kasir belum ditentukan")
}

func (h *CashBankMutationHandler) GetAvailableInvoicePayments(c *gin.Context) {
	branchScope := c.Query("branch_id")
	roleValue, _ := c.Get("role")
	role, _ := roleValue.(string)
	if role == string(domain.RoleCashier) || role == string(domain.RoleAdminKasir) {
		userValue, _ := c.Get("userID")
		userID, _ := userValue.(uuid.UUID)
		var employee domain.Employee
		if err := h.db.Where("user_id = ?", userID).First(&employee).Error; err != nil {
			utils.ErrorResponse(c, http.StatusForbidden, "Data karyawan kasir tidak ditemukan")
			return
		}
		branchID, err := cashierDepositBranch(employee, branchScope)
		if err != nil {
			utils.ErrorResponse(c, http.StatusForbidden, err.Error())
			return
		}
		branchScope = branchID.String()
	}
	query := h.db.Table("sales_payments AS sp").
		Select(`sp.id AS payment_id, st.id AS invoice_id, st.receipt_no, COALESCE(NULLIF(st.buyer_name, ''), store.name, '') AS buyer_name,
			TRIM(CONCAT(COALESCE(emp.first_name, ''), ' ', COALESCE(emp.last_name, ''))) AS salesman_name,
			sp.payment_date, st.transaction_date, sp.amount AS payment_amount, st.company_id AS company_id,
			sp.amount - COALESCE(allocated.allocated_amount, 0) AS available`).
		Joins("LEFT JOIN sales_orders AS so ON so.id = sp.sales_order_id").
		Joins("JOIN sales_transactions AS st ON st.id = COALESCE(sp.sales_transaction_id, so.invoice_id)").
		Joins("LEFT JOIN employees AS emp ON emp.id = st.employee_id").
		Joins("LEFT JOIN stores AS store ON store.id = st.store_id").
		Joins(`LEFT JOIN (
			SELECT allocation.sales_payment_id, SUM(allocation.amount) AS allocated_amount
			FROM cash_bank_mutation_invoices AS allocation
			JOIN cash_bank_mutations AS mutation ON mutation.id = allocation.mutation_id
			WHERE mutation.category = 'FACTORY_INVOICE_DEPOSIT' AND mutation.status_setoran <> 'REJECTED'
			GROUP BY allocation.sales_payment_id
		) AS allocated ON allocated.sales_payment_id = sp.id`).
		Where("UPPER(TRIM(sp.payment_status)) IN ? AND UPPER(TRIM(sp.payment_method)) IN ('CASH', 'TUNAI') AND sp.deleted_at IS NULL AND st.deleted_at IS NULL", []string{"SUCCESS", "PAID", "APPROVED", "CASHIER_APPROVED"}).
		Where("sp.amount - COALESCE(allocated.allocated_amount, 0) > 0.009")
	if branchScope != "" {
		query = query.Where("COALESCE(st.branch_id, emp.operating_branch_id, emp.branch_id) = ?", branchScope)
	}
	if companyID := c.Query("company_id"); companyID != "" && companyID != "ALL" {
		query = query.Where("st.company_id = ?", companyID)
	}
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		like := "%" + search + "%"
		query = query.Where("st.receipt_no ILIKE ? OR st.buyer_name ILIKE ? OR store.name ILIKE ? OR emp.first_name ILIKE ? OR emp.last_name ILIKE ?", like, like, like, like, like)
	}
	var receipts []availableInvoicePayment
	if err := query.Order("sp.payment_date DESC, st.transaction_date DESC, st.receipt_no ASC").Scan(&receipts).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", receipts)
}

func (h *CashBankMutationHandler) GetCashBalance(c *gin.Context) {
	branchID := c.Query("branch_id")
	if branchID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "branch_id is required")
		return
	}
	var branch domain.Branch
	if err := h.db.First(&branch, "id = ?", branchID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	companyIDStr := c.Query("company_id")
	var manualCash, pettyCash float64
	manualCash = branch.ManualCashBalance
	pettyCash = branch.PettyCashBalance

	// Query per-company balances for this branch
	var branchCompanyBalances []domain.BranchCompanyBalance
	_ = h.db.Preload("Company").Where("branch_id = ?", branchID).Find(&branchCompanyBalances).Error

	var companies []domain.Company
	_ = h.db.Order("name ASC").Find(&companies).Error

	type CompanyBalanceDTO struct {
		CompanyID         uuid.UUID `json:"company_id"`
		CompanyName       string    `json:"company_name"`
		CompanyCode       string    `json:"company_code"`
		ManualCashBalance float64   `json:"manual_cash_balance"`
		PettyCashBalance  float64   `json:"petty_cash_balance"`
	}

	companyBalancesMap := make(map[uuid.UUID]*CompanyBalanceDTO)
	for _, comp := range companies {
		companyBalancesMap[comp.ID] = &CompanyBalanceDTO{
			CompanyID:         comp.ID,
			CompanyName:       comp.Name,
			CompanyCode:       comp.Code,
			ManualCashBalance: 0,
			PettyCashBalance:  0,
		}
	}

	for _, bcb := range branchCompanyBalances {
		if dto, exists := companyBalancesMap[bcb.CompanyID]; exists {
			dto.ManualCashBalance = bcb.ManualCashBalance
			dto.PettyCashBalance = bcb.PettyCashBalance
		}
	}

	companyBalanceList := make([]CompanyBalanceDTO, 0, len(companies))
	for _, comp := range companies {
		if dto, exists := companyBalancesMap[comp.ID]; exists {
			companyBalanceList = append(companyBalanceList, *dto)
		}
	}

	if companyIDStr != "" && companyIDStr != "ALL" {
		if compUUID, err := uuid.Parse(companyIDStr); err == nil {
			if dto, exists := companyBalancesMap[compUUID]; exists {
				manualCash = dto.ManualCashBalance
				pettyCash = dto.PettyCashBalance
			} else {
				manualCash = 0
				pettyCash = 0
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"manual_cash_balance": manualCash,
			"petty_cash_balance":  pettyCash,
			"company_balances":    companyBalanceList,
		},
	})
}

func (h *CashBankMutationHandler) SetPettyCashBalance(c *gin.Context) {
	branchID := c.Param("id")
	var req struct {
		CompanyID        string  `json:"company_id"`
		PettyCashBalance float64 `json:"petty_cash_balance"`
		Notes            string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var branch domain.Branch
	if err := h.db.First(&branch, "id = ?", branchID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Branch tidak ditemukan: "+err.Error())
		return
	}

	companyID := parseOptionalUUID(req.CompanyID)
	if companyID == nil && branch.CompanyID != nil {
		companyID = branch.CompanyID
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var oldBalance float64
		if companyID != nil && *companyID != uuid.Nil {
			var bcb domain.BranchCompanyBalance
			err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *companyID).First(&bcb).Error
			if err != nil {
				bcb = domain.BranchCompanyBalance{
					ID:                uuid.New(),
					BranchID:          branch.ID,
					CompanyID:         *companyID,
					PettyCashBalance:  req.PettyCashBalance,
					ManualCashBalance: 0,
				}
				if err := tx.Create(&bcb).Error; err != nil {
					return err
				}
			} else {
				oldBalance = bcb.PettyCashBalance
				bcb.PettyCashBalance = req.PettyCashBalance
				if err := tx.Save(&bcb).Error; err != nil {
					return err
				}
			}
		} else {
			oldBalance = branch.PettyCashBalance
		}

		branch.PettyCashBalance = req.PettyCashBalance
		if err := tx.Save(&branch).Error; err != nil {
			return err
		}

		diff := req.PettyCashBalance - oldBalance
		if diff != 0 {
			mutationType := "IN"
			amount := diff
			if diff < 0 {
				mutationType = "OUT"
				amount = -diff
			}

			noteText := req.Notes
			if noteText == "" {
				noteText = "Penyesuaian Saldo Awal / Manual Kas Kecil"
			}

			mutation := domain.CashBankMutation{
				ID:           uuid.New(),
				BranchID:     branch.ID,
				CompanyID:    companyID,
				MutationType: mutationType,
				FundType:     "PETTY_CASH",
				Category:     "INITIAL_BALANCE",
				Amount:       amount,
				Notes:        noteText,
				ProofURL:     "",
				MutationDate: time.Now(),
			}
			if err := tx.Create(&mutation).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Saldo kas kecil berhasil diperbarui", gin.H{
		"petty_cash_balance": req.PettyCashBalance,
	})
}

func (h *CashBankMutationHandler) SetCashBalance(c *gin.Context) {
	branchID := c.Param("id")
	var req struct {
		CompanyID         string  `json:"company_id"`
		ManualCashBalance float64 `json:"manual_cash_balance"`
		Notes             string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var branch domain.Branch
	if err := h.db.First(&branch, "id = ?", branchID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Branch tidak ditemukan: "+err.Error())
		return
	}

	companyID := parseOptionalUUID(req.CompanyID)
	if companyID == nil && branch.CompanyID != nil {
		companyID = branch.CompanyID
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var oldBalance float64
		if companyID != nil && *companyID != uuid.Nil {
			var bcb domain.BranchCompanyBalance
			err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *companyID).First(&bcb).Error
			if err != nil {
				bcb = domain.BranchCompanyBalance{
					ID:                uuid.New(),
					BranchID:          branch.ID,
					CompanyID:         *companyID,
					ManualCashBalance: req.ManualCashBalance,
					PettyCashBalance:  0,
				}
				if err := tx.Create(&bcb).Error; err != nil {
					return err
				}
			} else {
				oldBalance = bcb.ManualCashBalance
				bcb.ManualCashBalance = req.ManualCashBalance
				if err := tx.Save(&bcb).Error; err != nil {
					return err
				}
			}
		} else {
			oldBalance = branch.ManualCashBalance
		}

		branch.ManualCashBalance = req.ManualCashBalance
		if err := tx.Save(&branch).Error; err != nil {
			return err
		}

		diff := req.ManualCashBalance - oldBalance
		if diff != 0 {
			mutationType := "IN"
			amount := diff
			if diff < 0 {
				mutationType = "OUT"
				amount = -diff
			}

			noteText := req.Notes
			if noteText == "" {
				noteText = "Penyesuaian Saldo Awal / Manual Kas Besar"
			}

			mutation := domain.CashBankMutation{
				ID:           uuid.New(),
				BranchID:     branch.ID,
				CompanyID:    companyID,
				MutationType: mutationType,
				FundType:     "CASH",
				Category:     "INITIAL_BALANCE",
				Amount:       amount,
				Notes:        noteText,
				ProofURL:     "",
				MutationDate: time.Now(),
			}
			if err := tx.Create(&mutation).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Saldo kas besar berhasil diperbarui", gin.H{
		"manual_cash_balance": req.ManualCashBalance,
	})
}

func (h *CashBankMutationHandler) ListMutations(c *gin.Context) {
	var mutations []domain.CashBankMutation
	query := h.db.Preload("Branch").Preload("BankAccount.Branch").Preload("TargetKas").Preload("Reimbursement.Employee").Preload("Company").Preload("DestinationCompany").
		Preload("InvoiceItems.SalesPayment.SalesTransaction").Preload("InvoiceItems.SalesTransaction").Order("mutation_date DESC, created_at DESC")
	roleValue, _ := c.Get("role")
	role, _ := roleValue.(string)
	if role == string(domain.RoleCashier) || role == string(domain.RoleAdminKasir) || role == string(domain.RoleManager) {
		userValue, _ := c.Get("userID")
		userID, _ := userValue.(uuid.UUID)
		var employee domain.Employee
		if err := h.db.Where("user_id = ?", userID).First(&employee).Error; err != nil {
			utils.ErrorResponse(c, http.StatusForbidden, "Data karyawan tidak ditemukan")
			return
		}
		branchIDs := assignedDepositBranches(employee)
		if len(branchIDs) == 0 {
			utils.ErrorResponse(c, http.StatusForbidden, "Cabang karyawan belum ditentukan")
			return
		}
		query = query.Where("branch_id IN ?", branchIDs)
	}

	if branchID := c.Query("branch_id"); branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	if companyID := c.Query("company_id"); companyID != "" && companyID != "ALL" {
		query = query.Where("company_id = ?", companyID)
	}

	if fundType := c.Query("fund_type"); fundType != "" {
		query = query.Where("fund_type = ?", fundType)
	}

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if date := c.Query("date"); date != "" {
		query = query.Where("mutation_date = ?", date)
	} else {
		if startDate := c.Query("start_date"); startDate != "" {
			query = query.Where("mutation_date >= ?", startDate)
		}
		if endDate := c.Query("end_date"); endDate != "" {
			query = query.Where("mutation_date <= ?", endDate)
		}
	}

	if err := query.Find(&mutations).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", mutations)
}

func canReviewFactoryDeposits(role string) bool {
	return role == string(domain.RoleAdminKasir) || role == string(domain.RoleManager) || role == string(domain.RoleSuperAdmin)
}

func assignedDepositBranches(employee domain.Employee) []uuid.UUID {
	branchIDs := make([]uuid.UUID, 0, 2)
	if employee.BranchID != nil && *employee.BranchID != uuid.Nil {
		branchIDs = append(branchIDs, *employee.BranchID)
	}
	if employee.OperatingBranchID != nil && *employee.OperatingBranchID != uuid.Nil && (len(branchIDs) == 0 || branchIDs[0] != *employee.OperatingBranchID) {
		branchIDs = append(branchIDs, *employee.OperatingBranchID)
	}
	return branchIDs
}

func (h *CashBankMutationHandler) authorizeDepositReviewer(tx *gorm.DB, role string, userID uuid.UUID, branchID uuid.UUID) error {
	if !canReviewFactoryDeposits(role) {
		return errors.New("hanya Admin Kasir cabang yang dapat memverifikasi setoran")
	}
	if role == string(domain.RoleSuperAdmin) {
		return nil
	}
	var employee domain.Employee
	if err := tx.Where("user_id = ?", userID).First(&employee).Error; err != nil {
		return errors.New("data karyawan admin cabang tidak ditemukan")
	}
	_, err := cashierDepositBranch(employee, branchID.String())
	return err
}

func factoryDepositBankMatchesBranch(bank domain.CompanyBankAccount, branchID uuid.UUID) bool {
	return bank.BranchID != nil && *bank.BranchID == branchID
}

func isFactoryDepositCategory(category string) bool {
	return category == "FACTORY_INVOICE_DEPOSIT" || category == "FACTORY_NON_INVOICE_DEPOSIT"
}

func isNonInvoiceFundOrigin(origin string) bool {
	switch origin {
	case "OPENING_BALANCE", "CASH_ADJUSTMENT", "OTHER_NON_SALES":
		return true
	default:
		return false
	}
}

func validateFactoryInvoiceDepositDifference(depositAmount, allocatedTotal float64, origin, notes string) error {
	extraAmount := math.Round((depositAmount-allocatedTotal)*100) / 100
	if extraAmount < -0.009 {
		return errors.New("nominal setoran tidak boleh kurang dari alokasi faktur; kurangi nominal per faktur agar sisa tetap belum disetor")
	}
	if extraAmount > 0.009 && (!isNonInvoiceFundOrigin(origin) || strings.TrimSpace(notes) == "") {
		return errors.New("kelebihan setoran wajib mencantumkan asal dana nonfaktur dan alasan")
	}
	return nil
}

func (h *CashBankMutationHandler) ApproveFactoryInvoiceDeposit(c *gin.Context) {
	roleValue, _ := c.Get("role")
	role, _ := roleValue.(string)
	if !canReviewFactoryDeposits(role) {
		utils.ErrorResponse(c, http.StatusForbidden, "Hanya Admin Kasir cabang yang dapat memverifikasi setoran")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID mutasi tidak valid")
		return
	}
	var reviewerID *uuid.UUID
	if value, ok := c.Get("userID"); ok {
		if parsed, ok := value.(uuid.UUID); ok {
			reviewerID = &parsed
		}
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var mutation domain.CashBankMutation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&mutation, "id = ?", id).Error; err != nil {
			return err
		}
		if !isFactoryDepositCategory(mutation.Category) || mutation.StatusSetoran != "PENDING" || mutation.BankAccountID == nil {
			return errors.New("setoran pabrik ini tidak berstatus menunggu verifikasi")
		}
		if reviewerID == nil {
			return errors.New("identitas admin cabang tidak ditemukan")
		}
		if err := h.authorizeDepositReviewer(tx, role, *reviewerID, mutation.BranchID); err != nil {
			return err
		}
		var bank domain.CompanyBankAccount
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&bank, "id = ?", *mutation.BankAccountID).Error; err != nil {
			return err
		}
		legacyCompanyDestination := mutation.DestinationCompanyID != nil && bank.CompanyID != nil && *bank.CompanyID == *mutation.DestinationCompanyID
		if !factoryDepositBankMatchesBranch(bank, mutation.BranchID) && !legacyCompanyDestination {
			return errors.New("rekening tujuan tidak terdaftar pada cabang setoran")
		}
		bank.ManualBalance += mutation.Amount
		now := time.Now()
		bank.LastBalanceUpdate = &now
		if err := tx.Save(&bank).Error; err != nil {
			return err
		}
		mutation.StatusSetoran = "APPROVED"
		mutation.StatusSetoranAt = &now
		mutation.StatusSetoranBy = reviewerID
		mutation.StatusSetoranNote = "Diverifikasi dan diterima Admin Kasir cabang"
		return tx.Save(&mutation).Error
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Setoran faktur diterima", nil)
}

func (h *CashBankMutationHandler) RejectFactoryInvoiceDeposit(c *gin.Context) {
	roleValue, _ := c.Get("role")
	role, _ := roleValue.(string)
	if !canReviewFactoryDeposits(role) {
		utils.ErrorResponse(c, http.StatusForbidden, "Hanya Admin Kasir cabang yang dapat memverifikasi setoran")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Alasan penolakan wajib diisi")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID mutasi tidak valid")
		return
	}
	var reviewerID *uuid.UUID
	if value, ok := c.Get("userID"); ok {
		if parsed, ok := value.(uuid.UUID); ok {
			reviewerID = &parsed
		}
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var mutation domain.CashBankMutation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&mutation, "id = ?", id).Error; err != nil {
			return err
		}
		if !isFactoryDepositCategory(mutation.Category) || mutation.StatusSetoran != "PENDING" {
			return errors.New("setoran pabrik ini tidak berstatus menunggu verifikasi")
		}
		if reviewerID == nil {
			return errors.New("identitas admin cabang tidak ditemukan")
		}
		if err := h.authorizeDepositReviewer(tx, role, *reviewerID, mutation.BranchID); err != nil {
			return err
		}
		var branch domain.Branch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&branch, "id = ?", mutation.BranchID).Error; err != nil {
			return err
		}
		branch.ManualCashBalance += mutation.Amount
		if err := tx.Save(&branch).Error; err != nil {
			return err
		}
		if mutation.CompanyID != nil {
			var balance domain.BranchCompanyBalance
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("branch_id = ? AND company_id = ?", mutation.BranchID, *mutation.CompanyID).First(&balance).Error; err == nil {
				balance.ManualCashBalance += mutation.Amount
				if err := tx.Save(&balance).Error; err != nil {
					return err
				}
			}
		}
		now := time.Now()
		mutation.StatusSetoran = "REJECTED"
		mutation.StatusSetoranAt = &now
		mutation.StatusSetoranBy = reviewerID
		mutation.StatusSetoranNote = strings.TrimSpace(req.Reason)
		return tx.Save(&mutation).Error
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Setoran faktur ditolak dan saldo kas cabang dipulihkan", nil)
}

type CashBankMutationRequest struct {
	BranchID             string                        `json:"branch_id"`
	CompanyID            string                        `json:"company_id,omitempty"`
	DestinationCompanyID string                        `json:"destination_company_id,omitempty"`
	MutationType         string                        `json:"mutation_type"`
	FundType             string                        `json:"fund_type"`
	BankAccountID        string                        `json:"bank_account_id,omitempty"`
	Amount               float64                       `json:"amount"`
	Notes                string                        `json:"notes"`
	ProofURL             string                        `json:"proof_url"`
	FundOrigin           string                        `json:"fund_origin,omitempty"`
	MutationDate         string                        `json:"mutation_date"`
	TargetKasPabrikID    string                        `json:"target_kas_pabrik_id,omitempty"`
	StatusSetoran        string                        `json:"status_setoran,omitempty"`
	Category             string                        `json:"category,omitempty"`
	ReimbursementID      string                        `json:"reimbursement_id,omitempty"`
	InvoicePayments      []MutationInvoicePaymentInput `json:"invoice_payments,omitempty"`
}

type MutationInvoicePaymentInput struct {
	PaymentID uuid.UUID `json:"payment_id"`
	Amount    float64   `json:"amount"`
}

func availableFactoryDepositPayment(tx *gorm.DB, paymentID uuid.UUID) (float64, error) {
	var payment domain.SalesPayment
	if err := tx.First(&payment, "id = ?", paymentID).Error; err != nil {
		return 0, err
	}
	var allocated float64
	if err := tx.Table("cash_bank_mutation_invoices AS allocation").
		Joins("JOIN cash_bank_mutations AS mutation ON mutation.id = allocation.mutation_id").
		Where("allocation.sales_payment_id = ? AND mutation.category = ? AND mutation.status_setoran <> ?", paymentID, "FACTORY_INVOICE_DEPOSIT", "REJECTED").
		Select("COALESCE(SUM(allocation.amount), 0)").Scan(&allocated).Error; err != nil {
		return 0, err
	}
	return math.Max(0, payment.Amount-allocated), nil
}

func parseFlexibleDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
}

func (h *CashBankMutationHandler) CreateMutation(c *gin.Context) {
	var req CashBankMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	branchID, err := parseRequiredUUID(req.BranchID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Cabang (branch_id) wajib dipilih: "+err.Error())
		return
	}

	companyID := parseOptionalUUID(req.CompanyID)

	mutationDate, _ := parseFlexibleDate(req.MutationDate)
	if mutationDate.IsZero() {
		mutationDate = time.Now()
	}

	mutation := domain.CashBankMutation{
		ID:                   uuid.New(),
		BranchID:             branchID,
		CompanyID:            companyID,
		DestinationCompanyID: parseOptionalUUID(req.DestinationCompanyID),
		MutationType:         req.MutationType,
		FundType:             req.FundType,
		BankAccountID:        parseOptionalUUID(req.BankAccountID),
		Amount:               req.Amount,
		Notes:                req.Notes,
		ProofURL:             req.ProofURL,
		FundOrigin:           req.FundOrigin,
		MutationDate:         mutationDate,
		TargetKasPabrikID:    parseOptionalUUID(req.TargetKasPabrikID),
		StatusSetoran:        req.StatusSetoran,
		Category:             req.Category,
		ReimbursementID:      parseOptionalUUID(req.ReimbursementID),
	}

	if mutation.BranchID == uuid.Nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Cabang (branch_id) wajib dipilih")
		return
	}
	if mutation.TargetKasPabrikID != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Setoran ke pabrik harus memakai alur Setor Faktur atau Setor Dana Nonfaktur")
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		var branch domain.Branch
		if err := tx.First(&branch, "id = ?", mutation.BranchID).Error; err != nil {
			return fmt.Errorf("cabang tidak ditemukan: %w", err)
		}

		if mutation.CompanyID == nil && branch.CompanyID != nil {
			mutation.CompanyID = branch.CompanyID
		}

		var invoiceItems []domain.CashBankMutationInvoice
		if isFactoryDepositCategory(mutation.Category) {
			roleValue, _ := c.Get("role")
			role, _ := roleValue.(string)
			if role != string(domain.RoleCashier) && role != string(domain.RoleAdminKasir) && role != string(domain.RoleCashierPusat) && role != string(domain.RoleManager) && role != string(domain.RoleSuperAdmin) {
				return errors.New("peran Anda tidak diizinkan mencatat setoran pabrik")
			}
			if mutation.MutationType != "OUT" || mutation.FundType != "CASH" || mutation.BankAccountID == nil || mutation.Amount <= 0 {
				return errors.New("setoran pabrik harus memilih rekening bank cabang tujuan dan nominal positif")
			}
			if mutation.Category == "FACTORY_INVOICE_DEPOSIT" && len(req.InvoicePayments) == 0 {
				return errors.New("setoran hasil penjualan wajib memilih penerimaan faktur yang disetujui kasir")
			}
			if mutation.Category == "FACTORY_NON_INVOICE_DEPOSIT" && (len(req.InvoicePayments) != 0 || parseOptionalUUID(req.CompanyID) == nil || !isNonInvoiceFundOrigin(mutation.FundOrigin) || strings.TrimSpace(mutation.Notes) == "") {
				return errors.New("setoran nonfaktur wajib memilih PT sumber, asal dana nonpenjualan, dan alasan")
			}
			mutation.StatusSetoran = "PENDING"
			if strings.TrimSpace(mutation.ProofURL) == "" {
				return errors.New("bukti setoran wajib dilampirkan")
			}
			if role == string(domain.RoleCashier) || role == string(domain.RoleAdminKasir) {
				userValue, _ := c.Get("userID")
				userID, _ := userValue.(uuid.UUID)
				var cashier domain.Employee
				if err := tx.Where("user_id = ?", userID).First(&cashier).Error; err != nil {
					return errors.New("data karyawan kasir tidak ditemukan")
				}
				if _, err := cashierDepositBranch(cashier, mutation.BranchID.String()); err != nil {
					return err
				}
			}
			var bank domain.CompanyBankAccount
			if err := tx.First(&bank, "id = ? AND is_active = ?", *mutation.BankAccountID, true).Error; err != nil {
				return errors.New("rekening cabang tujuan tidak ditemukan atau tidak aktif")
			}
			if !factoryDepositBankMatchesBranch(bank, mutation.BranchID) {
				return errors.New("rekening bank tujuan harus terdaftar pada cabang setoran")
			}
			mutation.DestinationCompanyID = nil
			seen := make(map[uuid.UUID]bool, len(req.InvoicePayments))
			var allocatedTotal float64
			var selectedCompanyID uuid.UUID
			for _, allocation := range req.InvoicePayments {
				if allocation.PaymentID == uuid.Nil || seen[allocation.PaymentID] || allocation.Amount <= 0 || math.IsNaN(allocation.Amount) || math.IsInf(allocation.Amount, 0) {
					return errors.New("rincian penerimaan faktur tidak valid")
				}
				seen[allocation.PaymentID] = true
				var payment domain.SalesPayment
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payment, "id = ?", allocation.PaymentID).Error; err != nil {
					return errors.New("penerimaan kasir tidak ditemukan")
				}
				if !isApprovedInvoicePaymentStatus(payment.PaymentStatus) || (strings.ToUpper(strings.TrimSpace(payment.PaymentMethod)) != "CASH" && strings.ToUpper(strings.TrimSpace(payment.PaymentMethod)) != "TUNAI") {
					return errors.New("setoran hanya dapat memakai penerimaan tunai yang sudah disetujui kasir")
				}
				invoiceID := payment.SalesTransactionID
				if invoiceID == nil && payment.SalesOrderID != nil {
					var order domain.SalesOrder
					if err := tx.Select("invoice_id").First(&order, "id = ?", *payment.SalesOrderID).Error; err != nil {
						return errors.New("pesanan dari penerimaan kasir tidak ditemukan")
					}
					invoiceID = order.InvoiceID
				}
				if invoiceID == nil {
					return errors.New("penerimaan kasir belum terhubung ke faktur penjualan")
				}
				var invoice domain.SalesTransaction
				if err := tx.First(&invoice, "id = ?", *invoiceID).Error; err != nil {
					return errors.New("faktur dari penerimaan kasir tidak ditemukan")
				}
				invoiceBranchID := invoice.BranchID
				if invoiceBranchID == nil {
					var employee domain.Employee
					if err := tx.First(&employee, "id = ?", invoice.EmployeeID).Error; err != nil {
						return err
					}
					invoiceBranchID = employee.OperatingBranchID
					if invoiceBranchID == nil {
						invoiceBranchID = employee.BranchID
					}
				}
				if invoiceBranchID == nil || *invoiceBranchID != mutation.BranchID || (mutation.CompanyID != nil && invoice.CompanyID != *mutation.CompanyID) {
					return errors.New("penerimaan faktur harus berasal dari cabang dan PT yang sama dengan mutasi")
				}
				if selectedCompanyID == uuid.Nil {
					selectedCompanyID = invoice.CompanyID
				} else if selectedCompanyID != invoice.CompanyID {
					return errors.New("satu mutasi setoran hanya boleh berisi faktur dari PT yang sama")
				}
				available, err := availableFactoryDepositPayment(tx, payment.ID)
				if err != nil {
					return err
				}
				if allocation.Amount > available+0.009 {
					return fmt.Errorf("nominal setoran untuk faktur %s melebihi penerimaan tunai yang belum disetor", invoice.ReceiptNo)
				}
				allocatedTotal += allocation.Amount
				invoiceItems = append(invoiceItems, domain.CashBankMutationInvoice{
					ID: uuid.New(), MutationID: mutation.ID, SalesPaymentID: payment.ID,
					SalesTransactionID: invoice.ID, Amount: allocation.Amount,
				})
			}
			if mutation.CompanyID == nil && selectedCompanyID != uuid.Nil {
				mutation.CompanyID = &selectedCompanyID
			}
			if mutation.Category == "FACTORY_INVOICE_DEPOSIT" {
				if err := validateFactoryInvoiceDepositDifference(mutation.Amount, allocatedTotal, mutation.FundOrigin, mutation.Notes); err != nil {
					return err
				}
				if math.Round((mutation.Amount-allocatedTotal)*100)/100 <= 0.009 {
					mutation.FundOrigin = ""
				}
			}
		}

		if err := tx.Create(&mutation).Error; err != nil {
			return err
		}
		if len(invoiceItems) > 0 {
			if err := tx.Create(&invoiceItems).Error; err != nil {
				return err
			}
		}

		// Update balance logic
		if mutation.FundType == "PETTY_CASH" {
			if mutation.MutationType == "IN" {
				branch.PettyCashBalance += mutation.Amount
			} else {
				branch.PettyCashBalance -= mutation.Amount
			}
			if err := tx.Save(&branch).Error; err != nil {
				return err
			}

			// Update BranchCompanyBalance if companyID is present
			if mutation.CompanyID != nil && *mutation.CompanyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *mutation.CompanyID).First(&bcb).Error; err != nil {
					newBCB := domain.BranchCompanyBalance{
						ID:                uuid.New(),
						BranchID:          branch.ID,
						CompanyID:         *mutation.CompanyID,
						ManualCashBalance: 0,
					}
					if mutation.MutationType == "IN" {
						newBCB.PettyCashBalance = mutation.Amount
					} else {
						newBCB.PettyCashBalance = -mutation.Amount
					}
					if err := tx.Create(&newBCB).Error; err != nil {
						return err
					}
				} else {
					if mutation.MutationType == "IN" {
						bcb.PettyCashBalance += mutation.Amount
					} else {
						bcb.PettyCashBalance -= mutation.Amount
					}
					if err := tx.Save(&bcb).Error; err != nil {
						return err
					}
				}
			}
		} else if mutation.FundType == "CASH" {
			if mutation.MutationType == "IN" {
				branch.ManualCashBalance += mutation.Amount
			} else {
				branch.ManualCashBalance -= mutation.Amount
			}
			if err := tx.Save(&branch).Error; err != nil {
				return err
			}

			// Update BranchCompanyBalance if companyID is present
			if mutation.CompanyID != nil && *mutation.CompanyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *mutation.CompanyID).First(&bcb).Error; err != nil {
					newBCB := domain.BranchCompanyBalance{
						ID:               uuid.New(),
						BranchID:         branch.ID,
						CompanyID:        *mutation.CompanyID,
						PettyCashBalance: 0,
					}
					if mutation.MutationType == "IN" {
						newBCB.ManualCashBalance = mutation.Amount
					} else {
						newBCB.ManualCashBalance = -mutation.Amount
					}
					if err := tx.Create(&newBCB).Error; err != nil {
						return err
					}
				} else {
					if mutation.MutationType == "IN" {
						bcb.ManualCashBalance += mutation.Amount
					} else {
						bcb.ManualCashBalance -= mutation.Amount
					}
					if err := tx.Save(&bcb).Error; err != nil {
						return err
					}
				}
			}
		} else if mutation.FundType == "BANK" && mutation.BankAccountID != nil {
			var bank domain.CompanyBankAccount
			if err := tx.First(&bank, "id = ?", *mutation.BankAccountID).Error; err != nil {
				return err
			}
			now := time.Now()
			if mutation.MutationType == "IN" {
				bank.ManualBalance += mutation.Amount
			} else {
				bank.ManualBalance -= mutation.Amount
			}
			bank.LastBalanceUpdate = &now
			if err := tx.Save(&bank).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Mutasi berhasil dicatat", mutation)
}

func (h *CashBankMutationHandler) UpdateMutation(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	mutationID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID mutasi tidak valid: "+err.Error())
		return
	}

	var existing domain.CashBankMutation
	if err := h.db.First(&existing, "id = ?", mutationID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Mutasi tidak ditemukan: "+err.Error())
		return
	}
	if isFactoryDepositCategory(existing.Category) {
		utils.ErrorResponse(c, http.StatusBadRequest, "Mutasi setoran pabrik tidak dapat diedit. Hapus lalu catat ulang agar alokasi dan saldo tetap sesuai.")
		return
	}

	var req CashBankMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if isFactoryDepositCategory(req.Category) {
		utils.ErrorResponse(c, http.StatusBadRequest, "Setoran pabrik harus dibuat melalui alur setoran baru dan diverifikasi Admin Kasir cabang")
		return
	}
	if parseOptionalUUID(req.TargetKasPabrikID) != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Setoran ke pabrik harus memakai alur Setor Faktur atau Setor Dana Nonfaktur")
		return
	}

	mutationDate, _ := parseFlexibleDate(req.MutationDate)

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 1. Revert previous balance effect
		if existing.FundType == "PETTY_CASH" {
			var branch domain.Branch
			if err := tx.First(&branch, "id = ?", existing.BranchID).Error; err == nil {
				if existing.MutationType == "IN" {
					branch.PettyCashBalance -= existing.Amount
				} else {
					branch.PettyCashBalance += existing.Amount
				}
				_ = tx.Save(&branch)
			}
			if existing.CompanyID != nil && *existing.CompanyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", existing.BranchID, *existing.CompanyID).First(&bcb).Error; err == nil {
					if existing.MutationType == "IN" {
						bcb.PettyCashBalance -= existing.Amount
					} else {
						bcb.PettyCashBalance += existing.Amount
					}
					_ = tx.Save(&bcb)
				}
			}
		} else if existing.FundType == "CASH" {
			if !isFactoryDepositCategory(existing.Category) || existing.StatusSetoran != "REJECTED" {
				var branch domain.Branch
				if err := tx.First(&branch, "id = ?", existing.BranchID).Error; err == nil {
					if existing.MutationType == "IN" {
						branch.ManualCashBalance -= existing.Amount
					} else {
						branch.ManualCashBalance += existing.Amount
					}
					_ = tx.Save(&branch)
				}
				if existing.CompanyID != nil && *existing.CompanyID != uuid.Nil {
					var bcb domain.BranchCompanyBalance
					if err := tx.Where("branch_id = ? AND company_id = ?", existing.BranchID, *existing.CompanyID).First(&bcb).Error; err == nil {
						if existing.MutationType == "IN" {
							bcb.ManualCashBalance -= existing.Amount
						} else {
							bcb.ManualCashBalance += existing.Amount
						}
						_ = tx.Save(&bcb)
					}
				}
			}
		} else if existing.FundType == "BANK" && existing.BankAccountID != nil {
			var bank domain.CompanyBankAccount
			if err := tx.First(&bank, "id = ?", *existing.BankAccountID).Error; err == nil {
				if existing.MutationType == "IN" {
					bank.ManualBalance -= existing.Amount
				} else {
					bank.ManualBalance += existing.Amount
				}
				_ = tx.Save(&bank)
			}
		}
		// 2. Apply new balance effect
		fundType := req.FundType
		if fundType == "" {
			fundType = existing.FundType
		}
		mutationType := req.MutationType
		if mutationType == "" {
			mutationType = existing.MutationType
		}
		branchID := parseOptionalUUID(req.BranchID)
		effectiveBranchID := existing.BranchID
		if branchID != nil && *branchID != uuid.Nil {
			effectiveBranchID = *branchID
		}

		newCompanyID := parseOptionalUUID(req.CompanyID)
		if newCompanyID == nil {
			newCompanyID = existing.CompanyID
		}

		newBankAccountID := parseOptionalUUID(req.BankAccountID)

		if fundType == "PETTY_CASH" {
			var branch domain.Branch
			if err := tx.First(&branch, "id = ?", effectiveBranchID).Error; err != nil {
				return err
			}
			if mutationType == "IN" {
				branch.PettyCashBalance += req.Amount
			} else {
				branch.PettyCashBalance -= req.Amount
			}
			if err := tx.Save(&branch).Error; err != nil {
				return err
			}
			if newCompanyID != nil && *newCompanyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", effectiveBranchID, *newCompanyID).First(&bcb).Error; err != nil {
					newBCB := domain.BranchCompanyBalance{
						ID:                uuid.New(),
						BranchID:          effectiveBranchID,
						CompanyID:         *newCompanyID,
						ManualCashBalance: 0,
					}
					if mutationType == "IN" {
						newBCB.PettyCashBalance = req.Amount
					} else {
						newBCB.PettyCashBalance = -req.Amount
					}
					_ = tx.Create(&newBCB)
				} else {
					if mutationType == "IN" {
						bcb.PettyCashBalance += req.Amount
					} else {
						bcb.PettyCashBalance -= req.Amount
					}
					_ = tx.Save(&bcb)
				}
			}
		} else if fundType == "CASH" {
			var branch domain.Branch
			if err := tx.First(&branch, "id = ?", effectiveBranchID).Error; err != nil {
				return err
			}
			if mutationType == "IN" {
				branch.ManualCashBalance += req.Amount
			} else {
				branch.ManualCashBalance -= req.Amount
			}
			if err := tx.Save(&branch).Error; err != nil {
				return err
			}
			if newCompanyID != nil && *newCompanyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", effectiveBranchID, *newCompanyID).First(&bcb).Error; err != nil {
					newBCB := domain.BranchCompanyBalance{
						ID:               uuid.New(),
						BranchID:         effectiveBranchID,
						CompanyID:        *newCompanyID,
						PettyCashBalance: 0,
					}
					if mutationType == "IN" {
						newBCB.ManualCashBalance = req.Amount
					} else {
						newBCB.ManualCashBalance = -req.Amount
					}
					_ = tx.Create(&newBCB)
				} else {
					if mutationType == "IN" {
						bcb.ManualCashBalance += req.Amount
					} else {
						bcb.ManualCashBalance -= req.Amount
					}
					_ = tx.Save(&bcb)
				}
			}
		} else if fundType == "BANK" && newBankAccountID != nil {
			var bank domain.CompanyBankAccount
			if err := tx.First(&bank, "id = ?", *newBankAccountID).Error; err != nil {
				return err
			}
			now := time.Now()
			if mutationType == "IN" {
				bank.ManualBalance += req.Amount
			} else {
				bank.ManualBalance -= req.Amount
			}
			bank.LastBalanceUpdate = &now
			if err := tx.Save(&bank).Error; err != nil {
				return err
			}
		}

		// 3. Update mutation fields
		existing.MutationType = req.MutationType
		existing.FundType = req.FundType
		existing.CompanyID = newCompanyID
		existing.BankAccountID = newBankAccountID
		existing.Amount = req.Amount
		existing.Notes = req.Notes
		if req.ProofURL != "" {
			existing.ProofURL = req.ProofURL
		}
		if !mutationDate.IsZero() {
			existing.MutationDate = mutationDate
		}
		existing.Category = req.Category
		if newTargetKasPabrik := parseOptionalUUID(req.TargetKasPabrikID); newTargetKasPabrik != nil {
			existing.TargetKasPabrikID = newTargetKasPabrik
		}
		if req.StatusSetoran != "" {
			existing.StatusSetoran = req.StatusSetoran
		}

		return tx.Save(&existing).Error
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Mutasi berhasil diperbarui", existing)
}

func (h *CashBankMutationHandler) DeleteMutation(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	mutationID, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID mutasi tidak valid: "+err.Error())
		return
	}

	var existing domain.CashBankMutation
	if err := h.db.First(&existing, "id = ?", mutationID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Mutasi tidak ditemukan: "+err.Error())
		return
	}
	if isFactoryDepositCategory(existing.Category) && existing.StatusSetoran == "APPROVED" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Setoran yang sudah diterima Admin Kasir cabang tidak dapat dihapus; catat mutasi koreksi untuk menjaga audit")
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 1. Revert balance effect
		if existing.FundType == "PETTY_CASH" {
			var branch domain.Branch
			if err := tx.First(&branch, "id = ?", existing.BranchID).Error; err == nil {
				if existing.MutationType == "IN" {
					branch.PettyCashBalance -= existing.Amount
				} else {
					branch.PettyCashBalance += existing.Amount
				}
				_ = tx.Save(&branch)
			}
			if existing.CompanyID != nil && *existing.CompanyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", existing.BranchID, *existing.CompanyID).First(&bcb).Error; err == nil {
					if existing.MutationType == "IN" {
						bcb.PettyCashBalance -= existing.Amount
					} else {
						bcb.PettyCashBalance += existing.Amount
					}
					_ = tx.Save(&bcb)
				}
			}
		} else if existing.FundType == "CASH" {
			if !isFactoryDepositCategory(existing.Category) || existing.StatusSetoran != "REJECTED" {
				var branch domain.Branch
				if err := tx.First(&branch, "id = ?", existing.BranchID).Error; err == nil {
					if existing.MutationType == "IN" {
						branch.ManualCashBalance -= existing.Amount
					} else {
						branch.ManualCashBalance += existing.Amount
					}
					_ = tx.Save(&branch)
				}
				if existing.CompanyID != nil && *existing.CompanyID != uuid.Nil {
					var bcb domain.BranchCompanyBalance
					if err := tx.Where("branch_id = ? AND company_id = ?", existing.BranchID, *existing.CompanyID).First(&bcb).Error; err == nil {
						if existing.MutationType == "IN" {
							bcb.ManualCashBalance -= existing.Amount
						} else {
							bcb.ManualCashBalance += existing.Amount
						}
						_ = tx.Save(&bcb)
					}
				}
			}
		} else if existing.FundType == "BANK" && existing.BankAccountID != nil {
			var bank domain.CompanyBankAccount
			if err := tx.First(&bank, "id = ?", *existing.BankAccountID).Error; err == nil {
				if existing.MutationType == "IN" {
					bank.ManualBalance -= existing.Amount
				} else {
					bank.ManualBalance += existing.Amount
				}
				_ = tx.Save(&bank)
			}
		}
		if isFactoryDepositCategory(existing.Category) && existing.StatusSetoran == "APPROVED" && existing.BankAccountID != nil {
			var bank domain.CompanyBankAccount
			if err := tx.First(&bank, "id = ?", *existing.BankAccountID).Error; err == nil {
				bank.ManualBalance -= existing.Amount
				now := time.Now()
				bank.LastBalanceUpdate = &now
				if err := tx.Save(&bank).Error; err != nil {
					return err
				}
			}
		}

		// 2. Revert Reimbursement status if this mutation is linked to a reimbursement
		if existing.ReimbursementID != nil && *existing.ReimbursementID != uuid.Nil {
			var reimb domain.Reimbursement
			if err := tx.First(&reimb, "id = ?", *existing.ReimbursementID).Error; err == nil {
				reimb.DisbursementStatus = "UNPAID"
				reimb.DisbursedVia = ""
				reimb.DisbursedAt = nil
				reimb.DisbursedBy = nil
				reimb.DisbursementProofURL = ""
				_ = tx.Save(&reimb)
			}
		}

		// 3. Delete mutation record
		return tx.Delete(&existing).Error
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Mutasi berhasil dihapus dan saldo telah dikoreksi", nil)
}
