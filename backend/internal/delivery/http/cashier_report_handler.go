package http

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
)

type CashierReportHandler struct {
	db        *gorm.DB
	glPosting usecase.GLPostingUsecase
}

func NewCashierReportHandler(db *gorm.DB, glPosting usecase.GLPostingUsecase) *CashierReportHandler {
	return &CashierReportHandler{db: db, glPosting: glPosting}
}

func (h *CashierReportHandler) SetupRoutes(r *gin.RouterGroup) {
	cr := r.Group("/cashier-reports")
	{
		cr.GET("", h.ListReports)
		cr.GET("/:id", h.GetReportByID)
		cr.POST("", h.CreateOrUpdateReport)
		cr.PUT("/:id", h.UpdateReport)
		cr.DELETE("/:id", h.DeleteReport)
		cr.GET("/daily-summary", h.GetDailySummary)
		cr.POST("/upload", h.UploadProof)
	}

	ba := r.Group("/bank-accounts")
	{
		ba.PUT("/:id/manual-balance", h.UpdateManualBalance)
	}
}

func (h *CashierReportHandler) ListReports(c *gin.Context) {
	var reports []domain.CashierFinancialReport
	query := h.db.Preload("Branch").Preload("Company").Preload("TargetBankAccount").Preload("TargetKasPabrik").Order("report_date DESC, created_at DESC")

	if branchID := c.Query("branch_id"); branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	if companyID := c.Query("company_id"); companyID != "" && companyID != "ALL" {
		query = query.Where("company_id = ?", companyID)
	}

	if date := c.Query("date"); date != "" {
		query = query.Where("report_date = ?", date)
	} else {
		if startDate := c.Query("start_date"); startDate != "" {
			query = query.Where("report_date >= ?", startDate)
		}
		if endDate := c.Query("end_date"); endDate != "" {
			query = query.Where("report_date <= ?", endDate)
		}
	}

	if err := query.Find(&reports).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", reports)
}

func (h *CashierReportHandler) GetReportByID(c *gin.Context) {
	id := c.Param("id")
	if id == "banks" || id == "cash" {
		utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{})
		return
	}
	reportID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID laporan tidak valid")
		return
	}

	var report domain.CashierFinancialReport
	if err := h.db.Preload("Branch").Preload("Company").Preload("TargetBankAccount").Preload("TargetKasPabrik").First(&report, "id = ?", reportID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Laporan tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", report)
}

func (h *CashierReportHandler) GetDailySummary(c *gin.Context) {
	dateStr := c.Query("date")
	loc := time.FixedZone("WIB", 7*3600)
	if dateStr == "" {
		dateStr = time.Now().In(loc).Format("2006-01-02")
	}

	startOfDay, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		startOfDay = time.Now().In(loc).Truncate(24 * time.Hour)
	}
	endOfDay := startOfDay.Add(24 * time.Hour)

	var totalCash, totalTransfer float64

	type SummaryResult struct {
		PaymentMethod string
		Total         float64
	}
	var results []SummaryResult

	query := `
		SELECT sp.payment_method, SUM(sp.amount) as total
		FROM sales_payments sp
		LEFT JOIN sales_orders so ON so.id = sp.sales_order_id
		LEFT JOIN sales_transactions st ON st.id = sp.sales_transaction_id
		LEFT JOIN employees e ON e.id = sp.employee_id
		WHERE sp.payment_status IN ('PAID', 'APPROVED', 'CASHIER_APPROVED', 'SUCCESS')
		AND COALESCE(sp.approved_at, sp.updated_at) >= ? AND COALESCE(sp.approved_at, sp.updated_at) < ?
	`
	args := []interface{}{startOfDay, endOfDay}

	if companyIDStr := c.Query("company_id"); companyIDStr != "" && companyIDStr != "ALL" {
		query += ` AND (so.company_id = ? OR st.company_id = ?)`
		args = append(args, companyIDStr, companyIDStr)
	}

	var branchCashBalance, branchPettyCashBalance float64
	if branchIDStr := c.Query("branch_id"); branchIDStr != "" && branchIDStr != "ALL_BRANCHES" {
		query += ` AND COALESCE(e.operating_branch_id, e.branch_id, so.branch_id, st.branch_id) = ?`
		args = append(args, branchIDStr)

		var branch domain.Branch
		if err := h.db.First(&branch, "id = ?", branchIDStr).Error; err == nil {
			branchCashBalance = branch.ManualCashBalance
			branchPettyCashBalance = branch.PettyCashBalance

			// Check per-company balance if company_id is provided
			if companyIDStr := c.Query("company_id"); companyIDStr != "" && companyIDStr != "ALL" {
				if compUUID, err := uuid.Parse(companyIDStr); err == nil {
					var bcb domain.BranchCompanyBalance
					if err := h.db.Where("branch_id = ? AND company_id = ?", branch.ID, compUUID).First(&bcb).Error; err == nil {
						branchCashBalance = bcb.ManualCashBalance
						branchPettyCashBalance = bcb.PettyCashBalance
					} else {
						branchCashBalance = 0
						branchPettyCashBalance = 0
					}
				}
			}
		}
	}

	query += ` GROUP BY sp.payment_method`

	h.db.Raw(query, args...).Scan(&results)

	for _, res := range results {
		if res.PaymentMethod == "CASH" || res.PaymentMethod == "TUNAI" {
			totalCash += res.Total
		} else {
			totalTransfer += res.Total
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"total_cash_sales":     totalCash,
			"total_transfer_sales": totalTransfer,
			"manual_cash_balance":  branchCashBalance,
			"petty_cash_balance":   branchPettyCashBalance,
		},
	})
}

type CashierReportRequest struct {
	ID                     string  `json:"id,omitempty"`
	BranchID               string  `json:"branch_id"`
	CompanyID              string  `json:"company_id,omitempty"`
	ReportDate             string  `json:"report_date"`
	TotalCashSales         float64 `json:"total_cash_sales"`
	TotalTransferSales     float64 `json:"total_transfer_sales"`
	AllocatedToPettyCash   float64 `json:"allocated_to_petty_cash"`
	DepositedAmount        float64 `json:"deposited_amount"`
	DepositDestinationType string  `json:"deposit_destination_type"`
	TargetBankAccountID    string  `json:"target_bank_account_id,omitempty"`
	TargetKasPabrikID      string  `json:"target_kas_pabrik_id,omitempty"`
	TargetBankName         string  `json:"target_bank_name"`
	TargetAccountNumber    string  `json:"target_account_number"`
	TargetAccountName      string  `json:"target_account_name"`
	DepositProofURL        string  `json:"deposit_proof_url"`
	Status                 string  `json:"status"`
	Notes                  string  `json:"notes"`
}

// parseOptionalUUID returns nil for empty strings, otherwise parses UUID.
func parseOptionalUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

// parseRequiredUUID returns error for empty strings.
func parseRequiredUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, fmt.Errorf("UUID tidak boleh kosong")
	}
	return uuid.Parse(s)
}

func parseFlexibleDateReport(dateStr string) (time.Time, error) {
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

func (h *CashierReportHandler) CreateOrUpdateReport(c *gin.Context) {
	var req CashierReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.DepositedAmount > 0 && req.DepositDestinationType == "FACTORY_KAS" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Setoran ke pabrik dicatat melalui Setor Faktur ke Pabrik atau Setor Dana Nonfaktur agar dapat diverifikasi Kasir Pusat")
		return
	}

	branchID, err := parseRequiredUUID(req.BranchID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "branch_id tidak valid: "+err.Error())
		return
	}

	reportDate, _ := parseFlexibleDateReport(req.ReportDate)
	if reportDate.IsZero() {
		reportDate = time.Now()
	}

	var branch domain.Branch
	if err := h.db.First(&branch, "id = ?", branchID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Cabang tidak ditemukan: "+err.Error())
		return
	}
	companyID := parseOptionalUUID(req.CompanyID)
	if companyID == nil && branch.CompanyID != nil {
		companyID = branch.CompanyID
	}

	report := domain.CashierFinancialReport{
		BranchID:               branchID,
		CompanyID:              companyID,
		ReportDate:             reportDate,
		TotalCashSales:         req.TotalCashSales,
		TotalTransferSales:     req.TotalTransferSales,
		AllocatedToPettyCash:   req.AllocatedToPettyCash,
		DepositedAmount:        req.DepositedAmount,
		DepositDestinationType: req.DepositDestinationType,
		TargetBankAccountID:    parseOptionalUUID(req.TargetBankAccountID),
		TargetKasPabrikID:      parseOptionalUUID(req.TargetKasPabrikID),
		TargetBankName:         req.TargetBankName,
		TargetAccountNumber:    req.TargetAccountNumber,
		TargetAccountName:      req.TargetAccountName,
		DepositProofURL:        req.DepositProofURL,
		Status:                 req.Status,
		Notes:                  req.Notes,
	}

	if reqID := parseOptionalUUID(req.ID); reqID != nil && *reqID != uuid.Nil {
		report.ID = *reqID
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if report.ID != uuid.Nil {
			if err := tx.Save(&report).Error; err != nil {
				return err
			}
		} else {
			report.ID = uuid.New()
			report.Status = "DEPOSITED"
			if err := tx.Create(&report).Error; err != nil {
				return err
			}
		}

		// 1. Alokasi ke Kas Kecil Cabang (Petty Cash)
		if report.AllocatedToPettyCash > 0 {
			branch.PettyCashBalance += report.AllocatedToPettyCash
			pettyCashMutation := domain.CashBankMutation{
				ID:              uuid.New(),
				BranchID:        branch.ID,
				CompanyID:       companyID,
				CashierReportID: &report.ID,
				MutationType:    "IN",
				FundType:        "PETTY_CASH",
				Category:        "PETTY_CASH_TOPUP",
				Amount:          report.AllocatedToPettyCash,
				Notes:           "Penyisihan Kas Besar ke Kas Kecil (Lap. Kasir " + report.ReportDate.Format("02/01/2006") + ")",
				ProofURL:        report.DepositProofURL,
				MutationDate:    report.ReportDate,
			}
			if err := tx.Create(&pettyCashMutation).Error; err != nil {
				return err
			}

			// Update BranchCompanyBalance
			if companyID != nil && *companyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *companyID).First(&bcb).Error; err != nil {
					newBCB := domain.BranchCompanyBalance{
						ID:                uuid.New(),
						BranchID:          branch.ID,
						CompanyID:         *companyID,
						PettyCashBalance:  report.AllocatedToPettyCash,
						ManualCashBalance: 0,
					}
					_ = tx.Create(&newBCB)
				} else {
					bcb.PettyCashBalance += report.AllocatedToPettyCash
					_ = tx.Save(&bcb)
				}
			}
		}

		// 2. Setoran ke Bank Cabang atau Kas Pabrik Pusat
		if report.DepositedAmount > 0 {
			if report.DepositDestinationType == "FACTORY_KAS" && report.TargetKasPabrikID != nil {
				factoryMutation := domain.CashBankMutation{
					ID:                uuid.New(),
					BranchID:          branch.ID,
					CompanyID:         companyID,
					CashierReportID:   &report.ID,
					MutationType:      "OUT",
					FundType:          "CASH",
					Category:          "SALES_DEPOSIT",
					TargetKasPabrikID: report.TargetKasPabrikID,
					StatusSetoran:     "PENDING",
					Amount:            report.DepositedAmount,
					Notes:             "Setor Kas Besar ke Kas Pabrik Pusat: " + report.TargetBankName + " - " + report.Notes,
					ProofURL:          report.DepositProofURL,
					MutationDate:      report.ReportDate,
				}
				if err := tx.Create(&factoryMutation).Error; err != nil {
					return err
				}
			} else if report.DepositDestinationType == "BRANCH_BANK" && report.TargetBankAccountID != nil {
				bankMutation := domain.CashBankMutation{
					ID:              uuid.New(),
					BranchID:        branch.ID,
					CompanyID:       companyID,
					CashierReportID: &report.ID,
					MutationType:    "OUT",
					FundType:        "CASH",
					Category:        "SALES_DEPOSIT",
					BankAccountID:   report.TargetBankAccountID,
					Amount:          report.DepositedAmount,
					Notes:           "Setor Kas Besar ke Bank Cabang: " + report.TargetBankName + " (" + report.TargetAccountNumber + ")",
					ProofURL:        report.DepositProofURL,
					MutationDate:    report.ReportDate,
				}
				if err := tx.Create(&bankMutation).Error; err != nil {
					return err
				}

				// Update saldo rekening bank cabang
				var bank domain.CompanyBankAccount
				if err := tx.First(&bank, "id = ?", *report.TargetBankAccountID).Error; err == nil {
					now := time.Now()
					bank.ManualBalance += report.DepositedAmount
					bank.LastBalanceUpdate = &now
					_ = tx.Save(&bank)
				}
			}
		}

		// 3. Update saldo Kas Besar cabang (berkurang sesuai yang disisihkan dan disetor)
		totalDeducted := report.AllocatedToPettyCash + report.DepositedAmount
		if branch.ManualCashBalance >= totalDeducted {
			branch.ManualCashBalance -= totalDeducted
		} else {
			branch.ManualCashBalance = 0
		}

		// Update BranchCompanyBalance cash balance
		if companyID != nil && *companyID != uuid.Nil {
			var bcb domain.BranchCompanyBalance
			if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *companyID).First(&bcb).Error; err == nil {
				if bcb.ManualCashBalance >= totalDeducted {
					bcb.ManualCashBalance -= totalDeducted
				} else {
					bcb.ManualCashBalance = 0
				}
				_ = tx.Save(&bcb)
			}
		}

		return tx.Save(&branch).Error
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Trigger GL Posting
	effectiveCompID := uuid.Nil
	if companyID != nil {
		effectiveCompID = *companyID
	}
	go func(cID uuid.UUID, rID uuid.UUID, amount float64, dt time.Time) {
		_ = h.glPosting.PostCashierReport(cID, rID, amount, dt)
	}(effectiveCompID, report.ID, report.DepositedAmount, report.ReportDate)

	utils.SuccessResponse(c, http.StatusOK, "Laporan berhasil disimpan dan dana telah dialokasikan", report)
}

func (h *CashierReportHandler) UpdateReport(c *gin.Context) {
	id := c.Param("id")
	reportID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID laporan tidak valid")
		return
	}

	var req CashierReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.DepositedAmount > 0 && req.DepositDestinationType == "FACTORY_KAS" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Setoran ke pabrik dicatat melalui Setor Faktur ke Pabrik atau Setor Dana Nonfaktur agar dapat diverifikasi Kasir Pusat")
		return
	}

	var oldReport domain.CashierFinancialReport
	if err := h.db.First(&oldReport, "id = ?", reportID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Laporan tidak ditemukan: "+err.Error())
		return
	}

	reportDate, _ := parseFlexibleDateReport(req.ReportDate)
	if reportDate.IsZero() {
		reportDate = oldReport.ReportDate
	}

	branchID := oldReport.BranchID
	if req.BranchID != "" {
		if bID, err := uuid.Parse(req.BranchID); err == nil {
			branchID = bID
		}
	}

	var branch domain.Branch
	if err := h.db.First(&branch, "id = ?", branchID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Cabang tidak ditemukan: "+err.Error())
		return
	}

	companyID := parseOptionalUUID(req.CompanyID)
	if companyID == nil && branch.CompanyID != nil {
		companyID = branch.CompanyID
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// --- 1. REVERT OLD REPORT BALANCES ---
		// Revert old petty cash allocation
		if oldReport.AllocatedToPettyCash > 0 {
			if branch.PettyCashBalance >= oldReport.AllocatedToPettyCash {
				branch.PettyCashBalance -= oldReport.AllocatedToPettyCash
			} else {
				branch.PettyCashBalance = 0
			}

			if oldReport.CompanyID != nil && *oldReport.CompanyID != uuid.Nil {
				var oldBcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *oldReport.CompanyID).First(&oldBcb).Error; err == nil {
					if oldBcb.PettyCashBalance >= oldReport.AllocatedToPettyCash {
						oldBcb.PettyCashBalance -= oldReport.AllocatedToPettyCash
					} else {
						oldBcb.PettyCashBalance = 0
					}
					_ = tx.Save(&oldBcb)
				}
			}
		}

		// Revert old bank deposit balance
		if oldReport.DepositedAmount > 0 && oldReport.DepositDestinationType == "BRANCH_BANK" && oldReport.TargetBankAccountID != nil {
			var oldBank domain.CompanyBankAccount
			if err := tx.First(&oldBank, "id = ?", *oldReport.TargetBankAccountID).Error; err == nil {
				if oldBank.ManualBalance >= oldReport.DepositedAmount {
					oldBank.ManualBalance -= oldReport.DepositedAmount
				} else {
					oldBank.ManualBalance = 0
				}
				now := time.Now()
				oldBank.LastBalanceUpdate = &now
				_ = tx.Save(&oldBank)
			}
		}

		// Restore old manual cash balance to branch & old company balance
		oldTotalDeducted := oldReport.AllocatedToPettyCash + oldReport.DepositedAmount
		branch.ManualCashBalance += oldTotalDeducted
		if oldReport.CompanyID != nil && *oldReport.CompanyID != uuid.Nil {
			var oldBcb domain.BranchCompanyBalance
			if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *oldReport.CompanyID).First(&oldBcb).Error; err == nil {
				oldBcb.ManualCashBalance += oldTotalDeducted
				_ = tx.Save(&oldBcb)
			}
		}

		// Delete old mutations linked to this cashier report
		if err := tx.Where("cashier_report_id = ?", oldReport.ID).Delete(&domain.CashBankMutation{}).Error; err != nil {
			return err
		}

		// --- 2. UPDATE REPORT DATA ---
		oldReport.BranchID = branchID
		oldReport.CompanyID = companyID
		oldReport.ReportDate = reportDate
		oldReport.TotalCashSales = req.TotalCashSales
		oldReport.TotalTransferSales = req.TotalTransferSales
		oldReport.AllocatedToPettyCash = req.AllocatedToPettyCash
		oldReport.DepositedAmount = req.DepositedAmount
		oldReport.DepositDestinationType = req.DepositDestinationType
		oldReport.TargetBankAccountID = parseOptionalUUID(req.TargetBankAccountID)
		oldReport.TargetKasPabrikID = parseOptionalUUID(req.TargetKasPabrikID)
		oldReport.TargetBankName = req.TargetBankName
		oldReport.TargetAccountNumber = req.TargetAccountNumber
		oldReport.TargetAccountName = req.TargetAccountName
		if req.DepositProofURL != "" {
			oldReport.DepositProofURL = req.DepositProofURL
		}
		if req.Status != "" {
			oldReport.Status = req.Status
		}
		oldReport.Notes = req.Notes

		if err := tx.Save(&oldReport).Error; err != nil {
			return err
		}

		// --- 3. APPLY NEW BALANCES & CREATE NEW MUTATIONS ---
		// A. New Petty Cash allocation
		if oldReport.AllocatedToPettyCash > 0 {
			branch.PettyCashBalance += oldReport.AllocatedToPettyCash
			pettyCashMutation := domain.CashBankMutation{
				ID:              uuid.New(),
				BranchID:        branch.ID,
				CompanyID:       companyID,
				CashierReportID: &oldReport.ID,
				MutationType:    "IN",
				FundType:        "PETTY_CASH",
				Category:        "PETTY_CASH_TOPUP",
				Amount:          oldReport.AllocatedToPettyCash,
				Notes:           "Penyisihan Kas Besar ke Kas Kecil (Lap. Kasir " + oldReport.ReportDate.Format("02/01/2006") + ")",
				ProofURL:        oldReport.DepositProofURL,
				MutationDate:    oldReport.ReportDate,
			}
			if err := tx.Create(&pettyCashMutation).Error; err != nil {
				return err
			}

			if companyID != nil && *companyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *companyID).First(&bcb).Error; err != nil {
					newBCB := domain.BranchCompanyBalance{
						ID:                uuid.New(),
						BranchID:          branch.ID,
						CompanyID:         *companyID,
						PettyCashBalance:  oldReport.AllocatedToPettyCash,
						ManualCashBalance: 0,
					}
					_ = tx.Create(&newBCB)
				} else {
					bcb.PettyCashBalance += oldReport.AllocatedToPettyCash
					_ = tx.Save(&bcb)
				}
			}
		}

		// B. New Deposit
		if oldReport.DepositedAmount > 0 {
			if oldReport.DepositDestinationType == "FACTORY_KAS" && oldReport.TargetKasPabrikID != nil {
				factoryMutation := domain.CashBankMutation{
					ID:                uuid.New(),
					BranchID:          branch.ID,
					CompanyID:         companyID,
					CashierReportID:   &oldReport.ID,
					MutationType:      "OUT",
					FundType:          "CASH",
					Category:          "SALES_DEPOSIT",
					TargetKasPabrikID: oldReport.TargetKasPabrikID,
					StatusSetoran:     "PENDING",
					Amount:            oldReport.DepositedAmount,
					Notes:             "Setor Kas Besar ke Kas Pabrik Pusat: " + oldReport.TargetBankName + " - " + oldReport.Notes,
					ProofURL:          oldReport.DepositProofURL,
					MutationDate:      oldReport.ReportDate,
				}
				if err := tx.Create(&factoryMutation).Error; err != nil {
					return err
				}
			} else if oldReport.DepositDestinationType == "BRANCH_BANK" && oldReport.TargetBankAccountID != nil {
				bankMutation := domain.CashBankMutation{
					ID:              uuid.New(),
					BranchID:        branch.ID,
					CompanyID:       companyID,
					CashierReportID: &oldReport.ID,
					MutationType:    "OUT",
					FundType:        "CASH",
					Category:        "SALES_DEPOSIT",
					BankAccountID:   oldReport.TargetBankAccountID,
					Amount:          oldReport.DepositedAmount,
					Notes:           "Setor Kas Besar ke Bank Cabang: " + oldReport.TargetBankName + " (" + oldReport.TargetAccountNumber + ")",
					ProofURL:        oldReport.DepositProofURL,
					MutationDate:    oldReport.ReportDate,
				}
				if err := tx.Create(&bankMutation).Error; err != nil {
					return err
				}

				var bank domain.CompanyBankAccount
				if err := tx.First(&bank, "id = ?", *oldReport.TargetBankAccountID).Error; err == nil {
					now := time.Now()
					bank.ManualBalance += oldReport.DepositedAmount
					bank.LastBalanceUpdate = &now
					_ = tx.Save(&bank)
				}
			}
		}

		// C. Deduct new total from manual cash
		newTotalDeducted := oldReport.AllocatedToPettyCash + oldReport.DepositedAmount
		if branch.ManualCashBalance >= newTotalDeducted {
			branch.ManualCashBalance -= newTotalDeducted
		} else {
			branch.ManualCashBalance = 0
		}

		if companyID != nil && *companyID != uuid.Nil {
			var bcb domain.BranchCompanyBalance
			if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *companyID).First(&bcb).Error; err == nil {
				if bcb.ManualCashBalance >= newTotalDeducted {
					bcb.ManualCashBalance -= newTotalDeducted
				} else {
					bcb.ManualCashBalance = 0
				}
				_ = tx.Save(&bcb)
			}
		}

		return tx.Save(&branch).Error
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan berhasil diperbarui", oldReport)
}

func (h *CashierReportHandler) DeleteReport(c *gin.Context) {
	id := c.Param("id")
	reportID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID laporan tidak valid")
		return
	}

	var report domain.CashierFinancialReport
	if err := h.db.First(&report, "id = ?", reportID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Laporan tidak ditemukan: "+err.Error())
		return
	}

	var branch domain.Branch
	if err := h.db.First(&branch, "id = ?", report.BranchID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Cabang tidak ditemukan: "+err.Error())
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 1. Revert petty cash allocation
		if report.AllocatedToPettyCash > 0 {
			if branch.PettyCashBalance >= report.AllocatedToPettyCash {
				branch.PettyCashBalance -= report.AllocatedToPettyCash
			} else {
				branch.PettyCashBalance = 0
			}

			if report.CompanyID != nil && *report.CompanyID != uuid.Nil {
				var bcb domain.BranchCompanyBalance
				if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *report.CompanyID).First(&bcb).Error; err == nil {
					if bcb.PettyCashBalance >= report.AllocatedToPettyCash {
						bcb.PettyCashBalance -= report.AllocatedToPettyCash
					} else {
						bcb.PettyCashBalance = 0
					}
					_ = tx.Save(&bcb)
				}
			}
		}

		// 2. Revert bank balance
		if report.DepositedAmount > 0 && report.DepositDestinationType == "BRANCH_BANK" && report.TargetBankAccountID != nil {
			var bank domain.CompanyBankAccount
			if err := tx.First(&bank, "id = ?", *report.TargetBankAccountID).Error; err == nil {
				if bank.ManualBalance >= report.DepositedAmount {
					bank.ManualBalance -= report.DepositedAmount
				} else {
					bank.ManualBalance = 0
				}
				now := time.Now()
				bank.LastBalanceUpdate = &now
				_ = tx.Save(&bank)
			}
		}

		// 3. Restore manual cash balance
		totalRestored := report.AllocatedToPettyCash + report.DepositedAmount
		branch.ManualCashBalance += totalRestored

		if report.CompanyID != nil && *report.CompanyID != uuid.Nil {
			var bcb domain.BranchCompanyBalance
			if err := tx.Where("branch_id = ? AND company_id = ?", branch.ID, *report.CompanyID).First(&bcb).Error; err == nil {
				bcb.ManualCashBalance += totalRestored
				_ = tx.Save(&bcb)
			}
		}

		// 4. Delete linked mutations
		if err := tx.Where("cashier_report_id = ?", report.ID).Delete(&domain.CashBankMutation{}).Error; err != nil {
			return err
		}

		// 5. Delete report
		if err := tx.Delete(&report).Error; err != nil {
			return err
		}

		return tx.Save(&branch).Error
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus laporan: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Laporan berhasil dihapus dan saldo telah dipulihkan", nil)
}

func (h *CashierReportHandler) UpdateManualBalance(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ManualBalance float64 `json:"manual_balance"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now()
	if err := h.db.Model(&domain.CompanyBankAccount{}).Where("id = ?", id).Updates(map[string]interface{}{
		"manual_balance":      req.ManualBalance,
		"last_balance_update": &now,
	}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Saldo berhasil diupdate", nil)
}

func (h *CashierReportHandler) UploadProof(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File tidak ditemukan")
		return
	}

	filename := uuid.New().String() + "_" + file.Filename
	uploadDir := "uploads/finance"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat folder upload")
		return
	}

	savePath := uploadDir + "/" + filename
	if err := utils.SaveUploadedFile(file, savePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan file")
		return
	}

	url := "/uploads/finance/" + filename
	utils.SuccessResponse(c, http.StatusOK, "File berhasil diupload", gin.H{"url": url})
}
