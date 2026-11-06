package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/sse"
)

type ReimbursementUseCase interface {
	Submit(userID uuid.UUID, req SubmitReimbursementRequest) error
	GetHistory(userID uuid.UUID, page, limit int) ([]domain.Reimbursement, error)
	GetAll(status, category string, page, limit int) ([]domain.Reimbursement, error)
	GetAllByBranch(status, category string, branchIDs []uuid.UUID, page, limit int) ([]domain.Reimbursement, error)
	Approve(approverID, id uuid.UUID, req ApproveReimbursementRequest) error
	AdjustBannerRate(approverUserID, id uuid.UUID, req AdjustBannerRateRequest) error
	Disburse(disburserUserID, id uuid.UUID, req DisburseReimbursementRequest) error
	UpdateMyReimbursement(userID, id uuid.UUID, req SubmitReimbursementRequest) error
	DeleteMyReimbursement(userID, id uuid.UUID) error
	AdminCreate(req AdminReimbursementRequest) error
	AdminUpdate(id uuid.UUID, req AdminReimbursementRequest) error
	AdminDelete(id uuid.UUID) error
}

// GLSettable adalah interface opsional yang memungkinkan injection GLPostingUsecase.
type GLSettable interface {
	SetGLPosting(gl GLPostingUsecase)
}

type reimbursementUseCase struct {
	repo         repository.ReimbursementRepository
	employeeRepo repository.EmployeeRepository
	fcmUseCase   FcmUsecase
	sseHub       *sse.Hub
	glPosting    GLPostingUsecase // optional: nil jika modul akuntansi belum aktif
}

func NewReimbursementUseCase(repo repository.ReimbursementRepository, employeeRepo repository.EmployeeRepository, fcmUseCase FcmUsecase, sseHub *sse.Hub) ReimbursementUseCase {
	return &reimbursementUseCase{repo: repo, employeeRepo: employeeRepo, fcmUseCase: fcmUseCase, sseHub: sseHub}
}

// SetGLPosting menyuntikkan GL posting usecase (dipanggil dari main.go setelah inisialisasi).
func (u *reimbursementUseCase) SetGLPosting(gl GLPostingUsecase) {
	u.glPosting = gl
}

type SubmitReimbursementRequest struct {
	Title              string   `json:"title" form:"title" binding:"required"`
	Description        string   `json:"description" form:"description"`
	Amount             float64  `json:"amount" form:"amount" binding:"required"`
	AttachmentURL      string   `json:"attachment_url" form:"attachment_url"`
	Category           string   `json:"category" form:"category"`
	StartKM            *float64 `json:"start_km" form:"start_km"`
	EndKM              *float64 `json:"end_km" form:"end_km"`
	TotalKM            *float64 `json:"total_km" form:"total_km"`
	StartKMPhotoURL    string   `json:"start_km_photo_url" form:"start_km_photo_url"`
	EndKMPhotoURL      string   `json:"end_km_photo_url" form:"end_km_photo_url"`
	VehiclePlateNumber string   `json:"vehicle_plate_number" form:"vehicle_plate_number"`
	FuelLiters         *float64 `json:"fuel_liters" form:"fuel_liters"`
}

type DisburseReimbursementRequest struct {
	BranchID             string `json:"branch_id" form:"branch_id"`
	CompanyID            string `json:"company_id" form:"company_id"`
	DisbursedVia         string `json:"disbursed_via" form:"disbursed_via" binding:"required,oneof=PETTY_CASH CASH BANK_TRANSFER"`
	BankAccountID        string `json:"bank_account_id" form:"bank_account_id"`
	DisbursementProofURL string `json:"disbursement_proof_url" form:"disbursement_proof_url"`
	Notes                string `json:"notes" form:"notes"`
}

type AdminReimbursementRequest struct {
	EmployeeID         string   `json:"employee_id" form:"employee_id" binding:"required"`
	Title              string   `json:"title" form:"title" binding:"required"`
	Description        string   `json:"description" form:"description"`
	Amount             float64  `json:"amount" form:"amount" binding:"required"`
	AttachmentURL      string   `json:"attachment_url" form:"attachment_url"`
	Status             string   `json:"status" form:"status"`
	Category           string   `json:"category" form:"category"`
	BannerCount        int      `json:"banner_count" form:"banner_count"`
	UnitRate           float64  `json:"unit_rate" form:"unit_rate"`
	BannerIDs          []string `json:"banner_ids" form:"banner_ids"`
	StartKM            *float64 `json:"start_km" form:"start_km"`
	EndKM              *float64 `json:"end_km" form:"end_km"`
	TotalKM            *float64 `json:"total_km" form:"total_km"`
	StartKMPhotoURL    string   `json:"start_km_photo_url" form:"start_km_photo_url"`
	EndKMPhotoURL      string   `json:"end_km_photo_url" form:"end_km_photo_url"`
	VehiclePlateNumber string   `json:"vehicle_plate_number" form:"vehicle_plate_number"`
	FuelLiters         *float64 `json:"fuel_liters" form:"fuel_liters"`
}

type ApproveReimbursementRequest struct {
	Status           string     `json:"status" binding:"required,oneof=APPROVED REJECTED"`
	RejectedReason   string     `json:"rejected_reason"`
	FundSource       string     `json:"fund_source"` // KAS_PABRIK, BRANCH_PETTY_CASH, BRANCH_CASH, BANK_TRANSFER
	KasPabrikID      *uuid.UUID `json:"kas_pabrik_id"`
	BankAccountID    *uuid.UUID `json:"bank_account_id"`
	ExpenseAccountID *uuid.UUID `json:"expense_account_id"`
	Notes            string     `json:"notes"`
}

type AdjustBannerRateRequest struct {
	UnitRate         float64    `json:"unit_rate" binding:"required,min=0"`
	Amount           float64    `json:"amount"`  // optional, if 0 defaults to UnitRate * BannerCount
	Approve          bool       `json:"approve"` // if true, also approve directly
	FundSource       string     `json:"fund_source"`
	KasPabrikID      *uuid.UUID `json:"kas_pabrik_id"`
	BankAccountID    *uuid.UUID `json:"bank_account_id"`
	ExpenseAccountID *uuid.UUID `json:"expense_account_id"`
	Notes            string     `json:"notes"`
}

func (u *reimbursementUseCase) Submit(userID uuid.UUID, req SubmitReimbursementRequest) error {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee record not found")
	}

	if req.AttachmentURL == "" {
		return errors.New("Bukti pembayaran (nota/struk) wajib dilampirkan")
	}

	category := req.Category
	if category == "" {
		category = "GENERAL"
	}

	var totalKM *float64
	if req.StartKM != nil && req.EndKM != nil {
		if *req.EndKM < *req.StartKM {
			return errors.New("KM Pulang tidak boleh lebih kecil dari KM Berangkat")
		}
		diff := *req.EndKM - *req.StartKM
		totalKM = &diff
	} else if req.TotalKM != nil {
		totalKM = req.TotalKM
	}

	reimbursement := &domain.Reimbursement{
		EmployeeID:         employee.ID,
		Title:              req.Title,
		Description:        req.Description,
		Amount:             req.Amount,
		AttachmentURL:      req.AttachmentURL,
		Status:             domain.ReimbursementPending,
		Category:           category,
		StartKM:            req.StartKM,
		EndKM:              req.EndKM,
		TotalKM:            totalKM,
		StartKMPhotoURL:    req.StartKMPhotoURL,
		EndKMPhotoURL:      req.EndKMPhotoURL,
		VehiclePlateNumber: req.VehiclePlateNumber,
		FuelLiters:         req.FuelLiters,
	}

	err = u.repo.Create(reimbursement)
	if err == nil && u.sseHub != nil {
		go func() {
			branchIDStr := ""
			if employee.BranchID != nil {
				branchIDStr = employee.BranchID.String()
			}
			event := sse.NotificationEvent{
				ID:        reimbursement.ID.String(),
				Type:      "reimbursement",
				Title:     "Klaim Reimbursement Baru",
				Message:   fmt.Sprintf("%s %s mengajukan reimbursement %s sebesar Rp %.2f", employee.FirstName, employee.LastName, reimbursement.Title, reimbursement.Amount),
				CompanyID: employee.CompanyID.String(),
				CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				Data: map[string]interface{}{
					"reimbursement_id": reimbursement.ID.String(),
					"employee":         fmt.Sprintf("%s %s", employee.FirstName, employee.LastName),
					"amount":           reimbursement.Amount,
					"branch_id":        branchIDStr,
				},
				TargetRoles: []string{
					string(domain.RoleSuperAdmin),
					string(domain.RoleCashier),
					string(domain.RoleCashierPusat),
					string(domain.RoleAdminKasir),
					string(domain.RoleHRAdmin),
					string(domain.RoleManager),
				},
			}
			if employee.BranchID != nil {
				event.TargetBranchIDs = []string{employee.BranchID.String()}
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}()
	}
	return err
}

func (u *reimbursementUseCase) GetHistory(userID uuid.UUID, page, limit int) ([]domain.Reimbursement, error) {
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("employee record not found")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	return u.repo.GetByEmployeeID(employee.ID, limit, offset)
}

func (u *reimbursementUseCase) GetAll(status, category string, page, limit int) ([]domain.Reimbursement, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	return u.repo.GetAll(status, category, limit, offset)
}

func (u *reimbursementUseCase) GetAllByBranch(status, category string, branchIDs []uuid.UUID, page, limit int) ([]domain.Reimbursement, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	return u.repo.GetAllByBranch(status, category, branchIDs, limit, offset)
}

func (u *reimbursementUseCase) Approve(approverID, id uuid.UUID, req ApproveReimbursementRequest) error {
	reimbursement, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if reimbursement == nil {
		return errors.New("reimbursement request not found")
	}

	if reimbursement.Status != domain.ReimbursementPending {
		return errors.New("reimbursement is already processed")
	}

	// Resolve approver employee ID from User ID
	approver, err := u.employeeRepo.FindByUserID(approverID)
	if err != nil {
		return err
	}
	if approver == nil {
		return errors.New("approver employee record not found")
	}

	emp, err := u.employeeRepo.FindByID(reimbursement.EmployeeID)
	if err != nil || emp == nil {
		return errors.New("karyawan pemohon reimbursement tidak ditemukan")
	}

	reimbursement.Status = domain.ReimbursementStatus(req.Status)
	reimbursement.ApprovedBy = &approver.ID

	if req.Status == string(domain.ReimbursementRejected) {
		reimbursement.RejectedReason = &req.RejectedReason
		if err := u.repo.Update(reimbursement); err != nil {
			return err
		}
	} else if req.Status == string(domain.ReimbursementApproved) {
		now := time.Now()
		var mutation *domain.CashBankMutation
		var kasTx *domain.KasTransaction
		var kasPabrikID *uuid.UUID

		branchID := emp.BranchID
		companyID := emp.CompanyID

		isPabrik := false
		if emp.Branch != nil && strings.Contains(strings.ToLower(emp.Branch.Name), "pabrik") {
			isPabrik = true
		}

		fundSource := req.FundSource
		if fundSource == "" {
			if isPabrik {
				fundSource = "KAS_PABRIK"
			} else {
				fundSource = "BRANCH_PETTY_CASH"
			}
		}

		// Aturan Bisnis: KAS_PABRIK / Petty Cash Pabrik hanya untuk Pabrik Utama
		if isPabrik && fundSource != "KAS_PABRIK" {
			return errors.New("Karyawan Pabrik Utama hanya dapat dicairkan menggunakan Petty Cash / KAS Pabrik")
		}
		if !isPabrik && fundSource == "KAS_PABRIK" {
			return errors.New("Petty Cash / KAS Pabrik hanya diperuntukkan untuk cabang Pabrik Utama")
		}

		descNotes := fmt.Sprintf("Reimbursement (%s): %s", emp.FirstName, reimbursement.Title)
		if req.Notes != "" {
			descNotes += " - " + req.Notes
		}

		// 1. Eksekusi KAS_PABRIK (Khusus Pabrik Utama)
		if fundSource == "KAS_PABRIK" {
			if req.KasPabrikID == nil {
				return errors.New("Pilih akun KAS Pabrik untuk pencairan reimbursement pabrik")
			}
			kasPabrikID = req.KasPabrikID

			kasTx = &domain.KasTransaction{
				ID:               uuid.New(),
				TransactionType:  "OUT",
				Amount:           reimbursement.Amount,
				ExpenseCategory:  "REIMBURSEMENT",
				ExpenseAccountID: nil, // Foreign key kas_transactions.expense_account_id mengacu ke tabel kas_expense_accounts, bukan accounts PSAK
				Description:      descNotes,
				ProofURL:         reimbursement.AttachmentURL,
				TransactionDate:  now,
				EmployeeID:       &emp.ID,
				CreatedBy:        &approverID,
				ReimbursementID:  &reimbursement.ID,
			}

			reimbursement.KasPabrikID = kasPabrikID
			reimbursement.DisbursedVia = "PETTY_CASH"
		} else if fundSource == "BRANCH_PETTY_CASH" {
			// 2. Eksekusi Kas Kecil Cabang
			if branchID == nil {
				return errors.New("Cabang karyawan tidak terdeteksi untuk pemotongan kas kecil")
			}
			mutation = &domain.CashBankMutation{
				ID:              uuid.New(),
				BranchID:        *branchID,
				CompanyID:       companyID,
				MutationType:    "OUT",
				FundType:        "PETTY_CASH",
				Category:        "REIMBURSEMENT",
				ReimbursementID: &reimbursement.ID,
				Amount:          reimbursement.Amount,
				Notes:           descNotes,
				ProofURL:        reimbursement.AttachmentURL,
				MutationDate:    now,
			}
			reimbursement.DisbursedVia = "PETTY_CASH"
		} else if fundSource == "BRANCH_CASH" {
			// 3. Eksekusi Kas Tunai Cabang
			if branchID == nil {
				return errors.New("Cabang karyawan tidak terdeteksi untuk pemotongan kas tunai")
			}
			mutation = &domain.CashBankMutation{
				ID:              uuid.New(),
				BranchID:        *branchID,
				CompanyID:       companyID,
				MutationType:    "OUT",
				FundType:        "CASH",
				Category:        "REIMBURSEMENT",
				ReimbursementID: &reimbursement.ID,
				Amount:          reimbursement.Amount,
				Notes:           descNotes,
				ProofURL:        reimbursement.AttachmentURL,
				MutationDate:    now,
			}
			reimbursement.DisbursedVia = "CASH"
		} else if fundSource == "BANK_TRANSFER" {
			// 4. Eksekusi Transfer Bank
			if req.BankAccountID == nil {
				return errors.New("Rekening bank sumber pencairan wajib dipilih")
			}
			if branchID == nil {
				return errors.New("Cabang karyawan tidak terdeteksi untuk transaksi bank")
			}
			mutation = &domain.CashBankMutation{
				ID:              uuid.New(),
				BranchID:        *branchID,
				CompanyID:       companyID,
				MutationType:    "OUT",
				FundType:        "BANK",
				BankAccountID:   req.BankAccountID,
				Category:        "REIMBURSEMENT",
				ReimbursementID: &reimbursement.ID,
				Amount:          reimbursement.Amount,
				Notes:           descNotes,
				ProofURL:        reimbursement.AttachmentURL,
				MutationDate:    now,
			}
			reimbursement.DisbursedVia = "BANK_TRANSFER"
			reimbursement.BankAccountID = req.BankAccountID
		}

		reimbursement.DisbursementStatus = "DISBURSED"
		reimbursement.DisbursedAt = &now
		reimbursement.DisbursedBy = &approverID
		reimbursement.DisbursementSourceType = fundSource
		reimbursement.ExpenseAccountID = req.ExpenseAccountID

		if err := u.repo.ApproveWithDeduction(reimbursement, branchID, mutation, kasTx, kasPabrikID); err != nil {
			return err
		}
	}

	// Send push notification to the employee in the background
	go func() {
		if emp != nil && emp.UserID != uuid.Nil {
			title := "Persetujuan Reimbursement"
			statusText := "disetujui dan dicairkan"
			if req.Status == string(domain.ReimbursementRejected) {
				statusText = "ditolak"
			}
			body := "Pengajuan klaim \"" + reimbursement.Title + "\" Anda telah " + statusText + "."
			if req.Status == string(domain.ReimbursementRejected) && req.RejectedReason != "" {
				body += " Alasan: " + req.RejectedReason
			}
			_ = u.fcmUseCase.SendPushNotification(emp.UserID, title, body)
		}
	}()

	// Auto-post ke General Ledger jika reimbursement di-approve dan GL aktif
	if req.Status == string(domain.ReimbursementApproved) && u.glPosting != nil {
		go func() {
			if emp != nil && emp.CompanyID != nil {
				_ = u.glPosting.PostReimbursement(*emp.CompanyID, reimbursement.ID, reimbursement.Amount, time.Now())
			}
		}()
	}

	return nil
}

func (u *reimbursementUseCase) AdjustBannerRate(approverUserID, id uuid.UUID, req AdjustBannerRateRequest) error {
	reimbursement, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if reimbursement == nil {
		return errors.New("pengajuan reimbursement tidak ditemukan")
	}

	count := reimbursement.BannerCount
	if count <= 0 {
		count = len(reimbursement.BannerOrders)
	}
	if count <= 0 {
		count = 1
	}

	amount := req.Amount
	if amount <= 0 && req.UnitRate > 0 {
		amount = req.UnitRate * float64(count)
	}

	reimbursement.UnitRate = req.UnitRate
	reimbursement.Amount = amount

	if req.Approve {
		// Panggil Approve dengan logic deduction
		approveReq := ApproveReimbursementRequest{
			Status:           string(domain.ReimbursementApproved),
			FundSource:       req.FundSource,
			KasPabrikID:      req.KasPabrikID,
			BankAccountID:    req.BankAccountID,
			ExpenseAccountID: req.ExpenseAccountID,
			Notes:            req.Notes,
		}
		if err := u.repo.Update(reimbursement); err != nil {
			return err
		}
		return u.Approve(approverUserID, id, approveReq)
	}

	return u.repo.Update(reimbursement)
}

func (u *reimbursementUseCase) Disburse(disburserUserID, id uuid.UUID, req DisburseReimbursementRequest) error {
	reimbursement, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if reimbursement == nil {
		return errors.New("pengajuan reimbursement tidak ditemukan")
	}

	if reimbursement.Status != domain.ReimbursementApproved {
		return errors.New("hanya reimbursement yang sudah disetujui (APPROVED) yang dapat dicairkan")
	}

	if reimbursement.DisbursementStatus == "DISBURSED" {
		return errors.New("reimbursement ini sudah pernah dicairkan")
	}

	// Resolve employee
	employee, err := u.employeeRepo.FindByID(reimbursement.EmployeeID)
	if err != nil || employee == nil {
		return errors.New("karyawan pemilik reimbursement tidak ditemukan")
	}

	var branchID *uuid.UUID
	if req.BranchID != "" {
		if bID, err := uuid.Parse(req.BranchID); err == nil {
			branchID = &bID
		}
	}
	if branchID == nil && employee.BranchID != nil {
		branchID = employee.BranchID
	}

	var companyID *uuid.UUID
	if req.CompanyID != "" {
		if cID, err := uuid.Parse(req.CompanyID); err == nil {
			companyID = &cID
		}
	}
	if companyID == nil && employee.CompanyID != nil {
		companyID = employee.CompanyID
	}

	now := time.Now()
	var mutation *domain.CashBankMutation

	notes := fmt.Sprintf("Pencairan Reimburse (%s): %s", employee.FirstName, reimbursement.Title)
	if req.Notes != "" {
		notes += " - " + req.Notes
	}

	if branchID == nil {
		return errors.New("cabang tidak terdeteksi untuk pencairan kas")
	}

	if req.DisbursedVia == "PETTY_CASH" {
		mutation = &domain.CashBankMutation{
			ID:              uuid.New(),
			BranchID:        *branchID,
			CompanyID:       companyID,
			MutationType:    "OUT",
			FundType:        "PETTY_CASH",
			Category:        "REIMBURSEMENT",
			ReimbursementID: &reimbursement.ID,
			Amount:          reimbursement.Amount,
			Notes:           notes,
			ProofURL:        req.DisbursementProofURL,
			MutationDate:    now,
		}
	} else if req.DisbursedVia == "CASH" {
		mutation = &domain.CashBankMutation{
			ID:              uuid.New(),
			BranchID:        *branchID,
			CompanyID:       companyID,
			MutationType:    "OUT",
			FundType:        "CASH",
			Category:        "REIMBURSEMENT",
			ReimbursementID: &reimbursement.ID,
			Amount:          reimbursement.Amount,
			Notes:           notes,
			ProofURL:        req.DisbursementProofURL,
			MutationDate:    now,
		}
	} else if req.DisbursedVia == "BANK_TRANSFER" {
		var bankAccountID *uuid.UUID
		if req.BankAccountID != "" {
			if bID, err := uuid.Parse(req.BankAccountID); err == nil {
				bankAccountID = &bID
			}
		}
		if bankAccountID == nil {
			return errors.New("rekening bank sumber pencairan wajib dipilih")
		}

		mutation = &domain.CashBankMutation{
			ID:              uuid.New(),
			BranchID:        *branchID,
			CompanyID:       companyID,
			MutationType:    "OUT",
			FundType:        "BANK",
			BankAccountID:   bankAccountID,
			Category:        "REIMBURSEMENT",
			ReimbursementID: &reimbursement.ID,
			Amount:          reimbursement.Amount,
			Notes:           notes,
			ProofURL:        req.DisbursementProofURL,
			MutationDate:    now,
		}
	}

	reimbursement.DisbursementStatus = "DISBURSED"
	reimbursement.DisbursedVia = req.DisbursedVia
	reimbursement.DisbursedAt = &now
	reimbursement.DisbursedBy = &disburserUserID
	reimbursement.DisbursementProofURL = req.DisbursementProofURL

	if err := u.repo.Disburse(reimbursement, branchID, mutation); err != nil {
		return err
	}

	// Send Push Notification & SSE
	go func() {
		if employee.UserID != uuid.Nil {
			title := "Pencairan Reimbursement Berhasil"
			body := fmt.Sprintf("Dana klaim \"%s\" sebesar Rp %.2f telah dicairkan via %s.", reimbursement.Title, reimbursement.Amount, req.DisbursedVia)
			_ = u.fcmUseCase.SendPushNotification(employee.UserID, title, body)
		}
		if u.sseHub != nil && employee.CompanyID != nil {
			event := sse.NotificationEvent{
				ID:        reimbursement.ID.String(),
				Type:      "reimbursement_disbursed",
				Title:     "Pencairan Reimbursement",
				Message:   fmt.Sprintf("Reimbursement %s telah dicairkan oleh kasir", reimbursement.Title),
				CompanyID: employee.CompanyID.String(),
				CreatedAt: now.Format("2006-01-02 15:04:05"),
			}
			_ = u.sseHub.PublishNotification(context.Background(), event)
		}
	}()

	return nil
}

func (u *reimbursementUseCase) UpdateMyReimbursement(userID, id uuid.UUID, req SubmitReimbursementRequest) error {
	reimbursement, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if reimbursement == nil {
		return errors.New("reimbursement request not found")
	}

	// Resolve employee from userID
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee record not found")
	}

	if reimbursement.EmployeeID != employee.ID {
		return errors.New("you can only update your own reimbursements")
	}

	if reimbursement.Status != domain.ReimbursementPending {
		return errors.New("cannot update reimbursement that is already processed")
	}

	reimbursement.Title = req.Title
	reimbursement.Description = req.Description
	reimbursement.Amount = req.Amount
	if req.AttachmentURL != "" {
		reimbursement.AttachmentURL = req.AttachmentURL
	}
	if req.Category != "" {
		reimbursement.Category = req.Category
	}
	if req.StartKM != nil {
		reimbursement.StartKM = req.StartKM
	}
	if req.EndKM != nil {
		reimbursement.EndKM = req.EndKM
	}
	if req.StartKM != nil && req.EndKM != nil {
		if *req.EndKM < *req.StartKM {
			return errors.New("KM Pulang tidak boleh lebih kecil dari KM Berangkat")
		}
		diff := *req.EndKM - *req.StartKM
		reimbursement.TotalKM = &diff
	} else if req.TotalKM != nil {
		reimbursement.TotalKM = req.TotalKM
	}
	if req.StartKMPhotoURL != "" {
		reimbursement.StartKMPhotoURL = req.StartKMPhotoURL
	}
	if req.EndKMPhotoURL != "" {
		reimbursement.EndKMPhotoURL = req.EndKMPhotoURL
	}
	if req.VehiclePlateNumber != "" {
		reimbursement.VehiclePlateNumber = req.VehiclePlateNumber
	}
	if req.FuelLiters != nil {
		reimbursement.FuelLiters = req.FuelLiters
	}

	return u.repo.Update(reimbursement)
}

func (u *reimbursementUseCase) DeleteMyReimbursement(userID, id uuid.UUID) error {
	reimbursement, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if reimbursement == nil {
		return errors.New("reimbursement request not found")
	}

	// Resolve employee from userID
	employee, err := u.employeeRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if employee == nil {
		return errors.New("employee record not found")
	}

	if reimbursement.EmployeeID != employee.ID {
		return errors.New("you can only delete your own reimbursements")
	}

	if reimbursement.Status != domain.ReimbursementPending {
		return errors.New("cannot delete reimbursement that is already processed")
	}

	return u.repo.Delete(id)
}

func (u *reimbursementUseCase) AdminCreate(req AdminReimbursementRequest) error {
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return errors.New("invalid employee_id")
	}

	category := req.Category
	if category == "" {
		if len(req.BannerIDs) > 0 {
			category = "BANNER_INSTALLATION"
		} else {
			category = "GENERAL"
		}
	}

	bannerCount := req.BannerCount
	if bannerCount <= 0 && len(req.BannerIDs) > 0 {
		bannerCount = len(req.BannerIDs)
	}

	var totalKM *float64
	if req.StartKM != nil && req.EndKM != nil {
		if *req.EndKM < *req.StartKM {
			return errors.New("KM Pulang tidak boleh lebih kecil dari KM Berangkat")
		}
		diff := *req.EndKM - *req.StartKM
		totalKM = &diff
	} else if req.TotalKM != nil {
		totalKM = req.TotalKM
	}

	reimbursement := &domain.Reimbursement{
		EmployeeID:         empID,
		Title:              req.Title,
		Description:        req.Description,
		Amount:             req.Amount,
		AttachmentURL:      req.AttachmentURL,
		Status:             domain.ReimbursementStatus(req.Status),
		Category:           category,
		BannerCount:        bannerCount,
		UnitRate:           req.UnitRate,
		StartKM:            req.StartKM,
		EndKM:              req.EndKM,
		TotalKM:            totalKM,
		StartKMPhotoURL:    req.StartKMPhotoURL,
		EndKMPhotoURL:      req.EndKMPhotoURL,
		VehiclePlateNumber: req.VehiclePlateNumber,
		FuelLiters:         req.FuelLiters,
	}
	if reimbursement.Status == "" {
		reimbursement.Status = domain.ReimbursementPending
	}

	var bannerUUIDs []uuid.UUID
	for _, bIDStr := range req.BannerIDs {
		if bUUID, err := uuid.Parse(bIDStr); err == nil {
			bannerUUIDs = append(bannerUUIDs, bUUID)
		}
	}

	if len(bannerUUIDs) > 0 {
		return u.repo.CreateWithBannerOrders(reimbursement, bannerUUIDs)
	}

	return u.repo.Create(reimbursement)
}

func (u *reimbursementUseCase) AdminUpdate(id uuid.UUID, req AdminReimbursementRequest) error {
	reimbursement, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if reimbursement == nil {
		return errors.New("reimbursement not found")
	}

	if req.EmployeeID != "" {
		empID, err := uuid.Parse(req.EmployeeID)
		if err == nil {
			reimbursement.EmployeeID = empID
		}
	}

	reimbursement.Title = req.Title
	reimbursement.Description = req.Description
	reimbursement.Amount = req.Amount
	if req.AttachmentURL != "" {
		reimbursement.AttachmentURL = req.AttachmentURL
	}
	if req.Status != "" {
		reimbursement.Status = domain.ReimbursementStatus(req.Status)
	}
	if req.Category != "" {
		reimbursement.Category = req.Category
	}
	if req.StartKM != nil {
		reimbursement.StartKM = req.StartKM
	}
	if req.EndKM != nil {
		reimbursement.EndKM = req.EndKM
	}
	if req.StartKM != nil && req.EndKM != nil {
		if *req.EndKM < *req.StartKM {
			return errors.New("KM Pulang tidak boleh lebih kecil dari KM Berangkat")
		}
		diff := *req.EndKM - *req.StartKM
		reimbursement.TotalKM = &diff
	} else if req.TotalKM != nil {
		reimbursement.TotalKM = req.TotalKM
	}
	if req.StartKMPhotoURL != "" {
		reimbursement.StartKMPhotoURL = req.StartKMPhotoURL
	}
	if req.EndKMPhotoURL != "" {
		reimbursement.EndKMPhotoURL = req.EndKMPhotoURL
	}
	if req.VehiclePlateNumber != "" {
		reimbursement.VehiclePlateNumber = req.VehiclePlateNumber
	}
	if req.FuelLiters != nil {
		reimbursement.FuelLiters = req.FuelLiters
	}

	return u.repo.Update(reimbursement)
}

func (u *reimbursementUseCase) AdminDelete(id uuid.UUID) error {
	return u.repo.Delete(id)
}
