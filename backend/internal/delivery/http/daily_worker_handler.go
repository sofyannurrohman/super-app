package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/pkg/utils"
	"gorm.io/gorm"
)

type DailyWorkerHandler struct {
	db *gorm.DB
}

func NewDailyWorkerHandler(db *gorm.DB) *DailyWorkerHandler {
	return &DailyWorkerHandler{db}
}

func (h *DailyWorkerHandler) SetupRoutes(r *gin.RouterGroup) {
	dw := r.Group("/daily-workers")
	{
		dw.GET("", h.ListWorkers)
		dw.POST("", h.CreateWorker)
		dw.PUT("/:id", h.UpdateWorker)
		dw.DELETE("/:id", h.DeleteWorker)
	}
	salary := r.Group("/daily-worker-salaries")
	{
		salary.GET("", h.ListSalaries)
		salary.POST("", h.CreateSalary)
		salary.POST("/batch-pay", h.BatchPaySalaries)
		salary.PUT("/:id", h.UpdateSalary)
		salary.DELETE("/:id", h.DeleteSalary)
		salary.PUT("/:id/pay", h.PaySalary)
	}
}

func (h *DailyWorkerHandler) ListWorkers(c *gin.Context) {
	var workers []domain.DailyWorker
	query := h.db.Preload("Branch").Order("name ASC")

	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))
	if len(branchIDs) > 0 {
		query = query.Where(`
			daily_workers.branch_id IN (?) 
			OR daily_workers.branch_id IS NULL
			OR daily_workers.name IN (
				SELECT TRIM(CONCAT(e.first_name, ' ', COALESCE(e.last_name, ''))) FROM employees e WHERE e.branch_id IN (?)
				UNION
				SELECT TRIM(e.first_name) FROM employees e WHERE e.branch_id IN (?)
			)
		`, branchIDs, branchIDs, branchIDs)
	}

	if err := query.Find(&workers).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", workers)
}

func (h *DailyWorkerHandler) CreateWorker(c *gin.Context) {
	var req struct {
		Name      string  `json:"name" binding:"required"`
		Type      string  `json:"type" binding:"required"` // Kuli or Driver
		BranchID  *string `json:"branch_id"`
		CompanyID *string `json:"company_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var branchUUID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" && *req.BranchID != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(*req.BranchID); err == nil {
			branchUUID = &parsed
		}
	}

	var companyUUID *uuid.UUID
	if req.CompanyID != nil && *req.CompanyID != "" && *req.CompanyID != "ALL_COMPANIES" {
		if parsed, err := uuid.Parse(*req.CompanyID); err == nil {
			companyUUID = &parsed
		}
	}

	worker := domain.DailyWorker{
		Name:      req.Name,
		Type:      req.Type,
		BranchID:  branchUUID,
		CompanyID: companyUUID,
	}
	if err := h.db.Create(&worker).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Worker created", worker)
}

func (h *DailyWorkerHandler) ListSalaries(c *gin.Context) {
	var salaries []domain.DailyWorkerSalary
	query := h.db.Preload("DailyWorker").
		Preload("Branch").
		Preload("DeliveryBatch").
		Preload("KasPabrik").
		Preload("KasTransaction").
		Preload("Payer").
		Order("created_at DESC")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	branchIDs := utils.ParseUUIDs(c.Query("branch_id"))
	if len(branchIDs) > 0 {
		query = query.Where(`
			daily_worker_salaries.branch_id IN (?) 
			OR daily_worker_salaries.kas_pabrik_id IN (SELECT id FROM kas_pabrik WHERE branch_id IN (?))
			OR daily_worker_salaries.delivery_batch_id IN (SELECT id FROM delivery_batches WHERE branch_id IN (?))
			OR daily_worker_salaries.daily_worker_id IN (SELECT id FROM daily_workers WHERE branch_id IN (?))
			OR daily_worker_salaries.daily_worker_id IN (
				SELECT dw.id FROM daily_workers dw 
				JOIN employees e ON (
					LOWER(TRIM(CONCAT(e.first_name, ' ', COALESCE(e.last_name, '')))) = LOWER(TRIM(dw.name)) 
					OR LOWER(TRIM(e.first_name)) = LOWER(TRIM(dw.name))
				) 
				WHERE e.branch_id IN (?)
			)
		`, branchIDs, branchIDs, branchIDs, branchIDs, branchIDs)
	}

	if err := query.Find(&salaries).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", salaries)
}

func mapWorkerTypeToExpenseCategory(workerType string) string {
	upper := strings.ToUpper(strings.TrimSpace(workerType))
	switch {
	case strings.Contains(upper, "KULI"):
		return "KULI"
	case strings.Contains(upper, "PACKING KECAP"):
		return "GAJI_BORONGAN_PACKING_KECAP"
	case strings.Contains(upper, "PACKING SAOS"):
		return "GAJI_BORONGAN_PACKING_SAOS"
	case strings.Contains(upper, "OVEN"):
		return "GAJI_BORONGAN_OVEN"
	case strings.Contains(upper, "PRODUKSI KECAP"):
		return "GAJI_PRODUKSI_KECAP"
	case strings.Contains(upper, "PRODUKSI SAOS"):
		return "GAJI_PRODUKSI_SAOS"
	case strings.Contains(upper, "CLEANING"):
		return "PEMBERSIHAN_PRODUKSI"
	case strings.Contains(upper, "TUKANG") || strings.Contains(upper, "TEKNISI"):
		return "GAJI_TUKANG_PABRIK"
	default:
		return "DAILY_WORKER"
	}
}

func (h *DailyWorkerHandler) CreateSalary(c *gin.Context) {
	var req struct {
		DailyWorkerID   string  `form:"daily_worker_id" json:"daily_worker_id" binding:"required"`
		BranchID        string  `form:"branch_id" json:"branch_id"`
		TotalPaid       float64 `form:"total_paid" json:"total_paid" binding:"required"`
		Notes           string  `form:"notes" json:"notes"`
		KasPabrikID     string  `form:"kas_pabrik_id" json:"kas_pabrik_id"`
		ExpenseCategory string  `form:"expense_category" json:"expense_category"`
		PaymentDate     string  `form:"payment_date" json:"payment_date"`
		AutoPay         bool    `form:"auto_pay" json:"auto_pay"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workerID, err := uuid.Parse(req.DailyWorkerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID pekerja tidak valid"})
		return
	}

	var worker domain.DailyWorker
	if err := h.db.First(&worker, "id = ?", workerID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pekerja tidak ditemukan"})
		return
	}

	var branchUUID *uuid.UUID
	if req.BranchID != "" && req.BranchID != "ALL_BRANCHES" {
		if parsed, err := uuid.Parse(req.BranchID); err == nil {
			branchUUID = &parsed
		}
	}

	// Auto-detect branch_id if not explicitly provided
	if branchUUID == nil && worker.BranchID != nil {
		branchUUID = worker.BranchID
	}
	if branchUUID == nil {
		var emp domain.Employee
		if err := h.db.Where("LOWER(TRIM(CONCAT(first_name, ' ', COALESCE(last_name, '')))) = LOWER(TRIM(?)) OR LOWER(TRIM(first_name)) = LOWER(TRIM(?))", worker.Name, worker.Name).First(&emp).Error; err == nil {
			branchUUID = emp.BranchID
		}
	}

	var proofURL string
	file, err := c.FormFile("proof_image")
	if err == nil && file != nil {
		if file.Size > 20*1024*1024 {
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "File size too large")
			return
		}
		filename := uuid.New().String() + "-" + filepath.Base(file.Filename)
		uploadDir := "uploads/daily_workers"
		os.MkdirAll(uploadDir, 0755)
		savePath := uploadDir + "/" + filename
		if err := utils.SaveUploadedFile(file, savePath); err == nil {
			proofURL = "/" + savePath
		}
	}

	var currentUserID *uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			currentUserID = &uid
		}
	}

	salary := domain.DailyWorkerSalary{
		DailyWorkerID: workerID,
		BranchID:      branchUUID,
		Date:          time.Now(),
		TotalPaid:     req.TotalPaid,
		Notes:         req.Notes,
		ProofURL:      proofURL,
		Status:        "UNPAID",
	}

	// If Kas is provided and AutoPay is true, perform deduction and mark PAID in 1 transaction
	if req.KasPabrikID != "" && req.AutoPay {
		kasUUID, err := uuid.Parse(req.KasPabrikID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Format ID Kas Pabrik tidak valid")
			return
		}

		txErr := h.db.Transaction(func(tx *gorm.DB) error {
			var kas domain.KasPabrik
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", kasUUID).Error; err != nil {
				return fmt.Errorf("KAS tidak ditemukan: %w", err)
			}

			if salary.BranchID == nil && kas.BranchID != uuid.Nil {
				salary.BranchID = &kas.BranchID
			}

			balBefore := kas.Balance
			balAfter := balBefore - req.TotalPaid
			kas.Balance = balAfter
			if err := tx.Save(&kas).Error; err != nil {
				return err
			}

			cat := req.ExpenseCategory
			if cat == "" {
				cat = mapWorkerTypeToExpenseCategory(worker.Type)
			}

			txDate := time.Now()
			if req.PaymentDate != "" {
				if parsed, err := time.Parse("2006-01-02", req.PaymentDate); err == nil {
					txDate = parsed
				}
			}

			desc := fmt.Sprintf("Gaji Pekerja Harian: %s (%s)", worker.Name, worker.Type)
			if req.Notes != "" {
				desc += " - " + req.Notes
			}

			kasTx := domain.KasTransaction{
				KasPabrikID:     kas.ID,
				TransactionType: "OUT",
				Amount:          req.TotalPaid,
				BalanceBefore:   balBefore,
				BalanceAfter:    balAfter,
				ExpenseCategory: cat,
				Description:     desc,
				ProofURL:        proofURL,
				TransactionDate: txDate,
				CreatedBy:       currentUserID,
			}
			if err := tx.Create(&kasTx).Error; err != nil {
				return err
			}

			now := time.Now()
			salary.Status = "PAID"
			salary.KasPabrikID = &kas.ID
			salary.KasTransactionID = &kasTx.ID
			salary.PaidAt = &now
			salary.PaidBy = currentUserID

			return tx.Create(&salary).Error
		})

		if txErr != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, txErr.Error())
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Salary recorded and paid from Kas", salary)
		return
	}

	if err := h.db.Create(&salary).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Salary recorded", salary)
}

func (h *DailyWorkerHandler) PaySalary(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		KasPabrikID     string   `json:"kas_pabrik_id" binding:"required"`
		TotalPaid       *float64 `json:"total_paid"`
		PaymentDate     string   `json:"payment_date"`
		Notes           string   `json:"notes"`
		ExpenseCategory string   `json:"expense_category"`
		ProofURL        string   `json:"proof_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Kas wajib dipilih: "+err.Error())
		return
	}

	kasUUID, err := uuid.Parse(req.KasPabrikID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID Kas tidak valid")
		return
	}

	var currentUserID *uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			currentUserID = &uid
		}
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		var salary domain.DailyWorkerSalary
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Preload("DailyWorker").First(&salary, "id = ?", id).Error; err != nil {
			return fmt.Errorf("tagihan gaji tidak ditemukan: %w", err)
		}

		if salary.Status == "PAID" {
			return fmt.Errorf("tagihan gaji ini sudah berstatus LUNAS")
		}

		finalAmount := salary.TotalPaid
		if req.TotalPaid != nil && *req.TotalPaid > 0 {
			finalAmount = *req.TotalPaid
		}

		var kas domain.KasPabrik
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", kasUUID).Error; err != nil {
			return fmt.Errorf("data KAS tidak ditemukan: %w", err)
		}

		balBefore := kas.Balance
		balAfter := balBefore - finalAmount
		kas.Balance = balAfter
		if err := tx.Save(&kas).Error; err != nil {
			return fmt.Errorf("gagal memperbarui saldo kas: %w", err)
		}

		workerName := "Pekerja Harian"
		workerType := "Harian"
		if salary.DailyWorker != nil {
			workerName = salary.DailyWorker.Name
			workerType = salary.DailyWorker.Type
		}

		cat := req.ExpenseCategory
		if cat == "" {
			cat = mapWorkerTypeToExpenseCategory(workerType)
		}

		txDate := time.Now()
		if req.PaymentDate != "" {
			if parsed, err := time.Parse("2006-01-02", req.PaymentDate); err == nil {
				txDate = parsed
			}
		}

		desc := fmt.Sprintf("Gaji Pekerja Harian: %s (%s)", workerName, workerType)
		if req.Notes != "" {
			desc += " - " + req.Notes
		} else if salary.Notes != "" {
			desc += " - " + salary.Notes
		}

		proofURL := salary.ProofURL
		if req.ProofURL != "" {
			proofURL = req.ProofURL
		}

		kasTx := domain.KasTransaction{
			KasPabrikID:     kas.ID,
			TransactionType: "OUT",
			Amount:          finalAmount,
			BalanceBefore:   balBefore,
			BalanceAfter:    balAfter,
			ExpenseCategory: cat,
			Description:     desc,
			ProofURL:        proofURL,
			TransactionDate: txDate,
			CreatedBy:       currentUserID,
		}
		if err := tx.Create(&kasTx).Error; err != nil {
			return fmt.Errorf("gagal mencatat transaksi pengeluaran kas: %w", err)
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":             "PAID",
			"total_paid":         finalAmount,
			"kas_pabrik_id":      kas.ID,
			"kas_transaction_id": kasTx.ID,
			"paid_at":            &now,
			"paid_by":            currentUserID,
		}
		if salary.BranchID == nil && kas.BranchID != uuid.Nil {
			updates["branch_id"] = kas.BranchID
		}
		if req.Notes != "" {
			updates["notes"] = req.Notes
		}
		if req.ProofURL != "" {
			updates["proof_url"] = req.ProofURL
		}

		if err := tx.Model(&domain.DailyWorkerSalary{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return fmt.Errorf("gagal memperbarui status gaji: %w", err)
		}

		return nil
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Gaji berhasil dibayar dari Kas", nil)
}

func (h *DailyWorkerHandler) BatchPaySalaries(c *gin.Context) {
	var req struct {
		SalaryIDs       []string `json:"salary_ids" binding:"required"`
		KasPabrikID     string   `json:"kas_pabrik_id" binding:"required"`
		PaymentDate     string   `json:"payment_date"`
		ExpenseCategory string   `json:"expense_category"`
		Notes           string   `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data batch payment tidak lengkap: "+err.Error())
		return
	}

	if len(req.SalaryIDs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Pilih setidaknya 1 tagihan gaji untuk dibayar")
		return
	}

	kasUUID, err := uuid.Parse(req.KasPabrikID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format ID Kas tidak valid")
		return
	}

	var currentUserID *uuid.UUID
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uuid.UUID); ok {
			currentUserID = &uid
		}
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		var salaries []domain.DailyWorkerSalary
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Preload("DailyWorker").Where("id IN ? AND status = 'UNPAID'", req.SalaryIDs).Find(&salaries).Error; err != nil {
			return fmt.Errorf("gagal mengambil tagihan gaji: %w", err)
		}

		if len(salaries) == 0 {
			return fmt.Errorf("tidak ada tagihan berstatus UNPAID yang valid untuk dibayar")
		}

		var kas domain.KasPabrik
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", kasUUID).Error; err != nil {
			return fmt.Errorf("KAS tidak ditemukan: %w", err)
		}

		txDate := time.Now()
		if req.PaymentDate != "" {
			if parsed, err := time.Parse("2006-01-02", req.PaymentDate); err == nil {
				txDate = parsed
			}
		}

		now := time.Now()
		for _, salary := range salaries {
			balBefore := kas.Balance
			balAfter := balBefore - salary.TotalPaid
			kas.Balance = balAfter

			workerName := "Pekerja Harian"
			workerType := "Harian"
			if salary.DailyWorker != nil {
				workerName = salary.DailyWorker.Name
				workerType = salary.DailyWorker.Type
			}

			cat := req.ExpenseCategory
			if cat == "" {
				cat = mapWorkerTypeToExpenseCategory(workerType)
			}

			desc := fmt.Sprintf("Gaji Pekerja Harian: %s (%s)", workerName, workerType)
			if req.Notes != "" {
				desc += " - " + req.Notes
			} else if salary.Notes != "" {
				desc += " - " + salary.Notes
			}

			kasTx := domain.KasTransaction{
				KasPabrikID:     kas.ID,
				TransactionType: "OUT",
				Amount:          salary.TotalPaid,
				BalanceBefore:   balBefore,
				BalanceAfter:    balAfter,
				ExpenseCategory: cat,
				Description:     desc,
				ProofURL:        salary.ProofURL,
				TransactionDate: txDate,
				CreatedBy:       currentUserID,
			}
			if err := tx.Create(&kasTx).Error; err != nil {
				return fmt.Errorf("gagal mencatat transaksi pengeluaran kas: %w", err)
			}

			updateData := map[string]interface{}{
				"status":             "PAID",
				"kas_pabrik_id":      kas.ID,
				"kas_transaction_id": kasTx.ID,
				"paid_at":            &now,
				"paid_by":            currentUserID,
			}
			if salary.BranchID == nil && kas.BranchID != uuid.Nil {
				updateData["branch_id"] = kas.BranchID
			}

			if err := tx.Model(&domain.DailyWorkerSalary{}).Where("id = ?", salary.ID).Updates(updateData).Error; err != nil {
				return fmt.Errorf("gagal mengupdate status gaji: %w", err)
			}
		}

		if err := tx.Save(&kas).Error; err != nil {
			return fmt.Errorf("gagal mengupdate saldo kas: %w", err)
		}

		return nil
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, fmt.Sprintf("%d tagihan gaji berhasil dibayar sekaligus dari Kas", len(req.SalaryIDs)), nil)
}

func (h *DailyWorkerHandler) UpdateWorker(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name string `json:"name" binding:"required"`
		Type string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.db.Model(&domain.DailyWorker{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": req.Name,
		"type": req.Type,
	}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Worker updated", nil)
}

func (h *DailyWorkerHandler) DeleteWorker(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Where("id = ?", id).Delete(&domain.DailyWorker{}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Worker deleted", nil)
}

func (h *DailyWorkerHandler) UpdateSalary(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		TotalPaid float64 `json:"total_paid" binding:"required"`
		Notes     string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.db.Model(&domain.DailyWorkerSalary{}).Where("id = ?", id).Updates(map[string]interface{}{
		"total_paid": req.TotalPaid,
		"notes":      req.Notes,
	}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Salary updated", nil)
}

func (h *DailyWorkerHandler) DeleteSalary(c *gin.Context) {
	id := c.Param("id")

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var salary domain.DailyWorkerSalary
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&salary, "id = ?", id).Error; err != nil {
			return err
		}

		// If this salary was PAID and has a KasTransactionID, roll back the Kas balance
		if salary.Status == "PAID" && salary.KasTransactionID != nil && salary.KasPabrikID != nil {
			var kas domain.KasPabrik
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", *salary.KasPabrikID).Error; err == nil {
				kas.Balance += salary.TotalPaid
				_ = tx.Save(&kas)
			}
			_ = tx.Where("id = ?", *salary.KasTransactionID).Delete(&domain.KasTransaction{})
		}

		return tx.Where("id = ?", id).Delete(&domain.DailyWorkerSalary{}).Error
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Salary deleted", nil)
}
