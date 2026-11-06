package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type ReimbursementRepository interface {
	Create(reimbursement *domain.Reimbursement) error
	CreateWithBannerOrders(reimbursement *domain.Reimbursement, bannerIDs []uuid.UUID) error
	GetByID(id uuid.UUID) (*domain.Reimbursement, error)
	GetByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]domain.Reimbursement, error)
	GetAll(status, category string, limit, offset int) ([]domain.Reimbursement, error)
	GetAllByBranch(status, category string, branchIDs []uuid.UUID, limit, offset int) ([]domain.Reimbursement, error)
	Update(reimbursement *domain.Reimbursement) error
	Delete(id uuid.UUID) error
	Disburse(reimbursement *domain.Reimbursement, branchID *uuid.UUID, mutation *domain.CashBankMutation) error
	ApproveWithDeduction(reimbursement *domain.Reimbursement, branchID *uuid.UUID, mutation *domain.CashBankMutation, kasTx *domain.KasTransaction, kasPabrikID *uuid.UUID) error
}

type reimbursementRepository struct {
	db *gorm.DB
}

func NewReimbursementRepository(db *gorm.DB) ReimbursementRepository {
	return &reimbursementRepository{db}
}

func (r *reimbursementRepository) Create(reimbursement *domain.Reimbursement) error {
	return r.db.Create(reimbursement).Error
}

func (r *reimbursementRepository) CreateWithBannerOrders(reimbursement *domain.Reimbursement, bannerIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(reimbursement).Error; err != nil {
			return err
		}
		if len(bannerIDs) > 0 {
			now := time.Now()
			if err := tx.Model(&domain.BannerOrder{}).
				Where("id IN (?)", bannerIDs).
				Updates(map[string]interface{}{
					"reimbursement_id": reimbursement.ID,
					"updated_at":       now,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *reimbursementRepository) GetByID(id uuid.UUID) (*domain.Reimbursement, error) {
	var reimbursement domain.Reimbursement
	err := r.db.Preload("Employee").
		Preload("Employee.Branch").
		Preload("Employee.JobPosition").
		Preload("Approver").
		Preload("Approver.JobPosition").
		Preload("Disburser").
		Preload("Disburser.Employee").
		Preload("Disburser.Employee.JobPosition").
		Preload("BannerOrders").
		Preload("KasPabrik").
		Preload("ExpenseAccount").
		Preload("BankAccount").
		First(&reimbursement, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	r.enrichSingleBBM(&reimbursement)
	return &reimbursement, nil
}

func (r *reimbursementRepository) GetByEmployeeID(employeeID uuid.UUID, limit, offset int) ([]domain.Reimbursement, error) {
	var reimbursements []domain.Reimbursement
	err := r.db.Preload("Approver").
		Preload("Approver.JobPosition").
		Preload("Disburser").
		Preload("Disburser.Employee").
		Preload("Disburser.Employee.JobPosition").
		Preload("BannerOrders").
		Preload("KasPabrik").
		Preload("ExpenseAccount").
		Preload("BankAccount").
		Where("employee_id = ?", employeeID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&reimbursements).Error
	if err == nil {
		reimbursements = r.enrichList(reimbursements)
	}
	return reimbursements, err
}

func (r *reimbursementRepository) GetAll(status, category string, limit, offset int) ([]domain.Reimbursement, error) {
	var reimbursements []domain.Reimbursement
	query := r.db.Preload("Employee").
		Preload("Employee.Branch").
		Preload("Employee.JobPosition").
		Preload("Approver").
		Preload("Approver.JobPosition").
		Preload("Disburser").
		Preload("Disburser.Employee").
		Preload("Disburser.Employee.JobPosition").
		Preload("BannerOrders").
		Preload("KasPabrik").
		Preload("ExpenseAccount").
		Preload("BankAccount")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" && category != "ALL" {
		query = query.Where("category = ?", category)
	}
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&reimbursements).Error
	if err == nil {
		reimbursements = r.enrichList(reimbursements)
	}
	return reimbursements, err
}

func (r *reimbursementRepository) GetAllByBranch(status, category string, branchIDs []uuid.UUID, limit, offset int) ([]domain.Reimbursement, error) {
	var reimbursements []domain.Reimbursement
	query := r.db.Preload("Employee").
		Preload("Employee.Branch").
		Preload("Employee.JobPosition").
		Preload("Approver").
		Preload("Approver.JobPosition").
		Preload("Disburser").
		Preload("Disburser.Employee").
		Preload("Disburser.Employee.JobPosition").
		Preload("BannerOrders").
		Preload("KasPabrik").
		Preload("ExpenseAccount").
		Preload("BankAccount").
		Joins("JOIN employees ON employees.id = reimbursements.employee_id")
	if status != "" {
		query = query.Where("reimbursements.status = ?", status)
	}
	if category != "" && category != "ALL" {
		query = query.Where("reimbursements.category = ?", category)
	}
	if len(branchIDs) > 0 {
		query = query.Where("employees.branch_id IN (?)", branchIDs)
	}
	err := query.Order("reimbursements.created_at DESC").Limit(limit).Offset(offset).Find(&reimbursements).Error
	if err == nil {
		reimbursements = r.enrichList(reimbursements)
	}
	return reimbursements, err
}

func (r *reimbursementRepository) enrichSingleBBM(reimb *domain.Reimbursement) {
	if reimb == nil {
		return
	}
	isBBM := strings.EqualFold(reimb.Category, "BBM") ||
		strings.Contains(strings.ToLower(reimb.Title), "bbm") ||
		strings.Contains(strings.ToLower(reimb.Title), "bensin") ||
		strings.Contains(strings.ToLower(reimb.Title), "solar") ||
		strings.Contains(strings.ToLower(reimb.Title), "pertamax") ||
		strings.Contains(strings.ToLower(reimb.Title), "pertalite") ||
		strings.Contains(strings.ToLower(reimb.Title), "bahan bakar") ||
		strings.Contains(strings.ToLower(reimb.Description), "bbm") ||
		strings.Contains(strings.ToLower(reimb.Description), "bensin") ||
		strings.Contains(strings.ToLower(reimb.Description), "solar")

	if !isBBM && reimb.StartKM == nil && reimb.EndKM == nil {
		return
	}

	// If start_km or photos are missing, attempt lookup from SalesDailyJourney
	if reimb.StartKM == nil || reimb.StartKMPhotoURL == "" || reimb.EndKM == nil || reimb.EndKMPhotoURL == "" {
		var userID uuid.UUID
		if reimb.Employee != nil && reimb.Employee.UserID != uuid.Nil {
			userID = reimb.Employee.UserID
		} else if reimb.EmployeeID != uuid.Nil {
			var emp domain.Employee
			if err := r.db.Select("id, user_id").First(&emp, "id = ?", reimb.EmployeeID).Error; err == nil {
				userID = emp.UserID
			}
		}

		if userID != uuid.Nil {
			loc := time.FixedZone("WIB", 7*3600)
			createdWIB := reimb.CreatedAt.In(loc)
			dateStr := createdWIB.Format("2006-01-02")
			prevDateStr := createdWIB.AddDate(0, 0, -1).Format("2006-01-02")

			var journey domain.SalesDailyJourney
			// Match journey for this user on the same date or previous day
			if err := r.db.Where("user_id = ? AND (date = ? OR date = ?)", userID, dateStr, prevDateStr).
				Order("date DESC, created_at DESC").
				First(&journey).Error; err == nil {

				updated := false
				if reimb.StartKM == nil && journey.StartKM > 0 {
					startKM := journey.StartKM
					reimb.StartKM = &startKM
					updated = true
				}
				if reimb.EndKM == nil && journey.EndKM > 0 {
					endKM := journey.EndKM
					reimb.EndKM = &endKM
					updated = true
				}
				if reimb.TotalKM == nil {
					if journey.TotalKM > 0 {
						totalKM := journey.TotalKM
						reimb.TotalKM = &totalKM
						updated = true
					} else if reimb.StartKM != nil && reimb.EndKM != nil && *reimb.EndKM >= *reimb.StartKM {
						diff := *reimb.EndKM - *reimb.StartKM
						reimb.TotalKM = &diff
						updated = true
					}
				}
				if reimb.StartKMPhotoURL == "" && journey.StartPhotoURL != "" {
					reimb.StartKMPhotoURL = journey.StartPhotoURL
					updated = true
				}
				if reimb.EndKMPhotoURL == "" && journey.EndPhotoURL != "" {
					reimb.EndKMPhotoURL = journey.EndPhotoURL
					updated = true
				}
				if isBBM && (reimb.Category == "" || reimb.Category == "GENERAL") {
					reimb.Category = "BBM"
					updated = true
				}

				if updated {
					// Persist to DB in background
					go func(id uuid.UUID, startKM, endKM, totalKM *float64, startPhoto, endPhoto, cat string) {
						updates := map[string]interface{}{}
						if startKM != nil {
							updates["start_km"] = *startKM
						}
						if endKM != nil {
							updates["end_km"] = *endKM
						}
						if totalKM != nil {
							updates["total_km"] = *totalKM
						}
						if startPhoto != "" {
							updates["start_km_photo_url"] = startPhoto
						}
						if endPhoto != "" {
							updates["end_km_photo_url"] = endPhoto
						}
						if cat != "" {
							updates["category"] = cat
						}
						if len(updates) > 0 {
							_ = r.db.Model(&domain.Reimbursement{}).Where("id = ?", id).Updates(updates).Error
						}
					}(reimb.ID, reimb.StartKM, reimb.EndKM, reimb.TotalKM, reimb.StartKMPhotoURL, reimb.EndKMPhotoURL, reimb.Category)
				}
			}
		}
	}
}

func (r *reimbursementRepository) enrichList(reimbursements []domain.Reimbursement) []domain.Reimbursement {
	for i := range reimbursements {
		r.enrichSingleBBM(&reimbursements[i])
	}
	return reimbursements
}

func (r *reimbursementRepository) Update(reimbursement *domain.Reimbursement) error {
	return r.db.Save(reimbursement).Error
}

func (r *reimbursementRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Reimbursement{}, "id = ?", id).Error
}

func (r *reimbursementRepository) Disburse(reimbursement *domain.Reimbursement, branchID *uuid.UUID, mutation *domain.CashBankMutation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(reimbursement).Error; err != nil {
			return err
		}

		if mutation != nil {
			if err := tx.Create(mutation).Error; err != nil {
				return err
			}

			// 1. Petty Cash
			if mutation.FundType == "PETTY_CASH" {
				if branchID != nil {
					var branch domain.Branch
					if err := tx.First(&branch, "id = ?", *branchID).Error; err == nil {
						branch.PettyCashBalance -= reimbursement.Amount
						_ = tx.Save(&branch)
					}
				}
				if mutation.CompanyID != nil && branchID != nil {
					var bcb domain.BranchCompanyBalance
					if err := tx.Where("branch_id = ? AND company_id = ?", *branchID, *mutation.CompanyID).First(&bcb).Error; err == nil {
						bcb.PettyCashBalance -= reimbursement.Amount
						_ = tx.Save(&bcb)
					}
				}
			} else if mutation.FundType == "CASH" {
				// 2. Cash
				if branchID != nil {
					var branch domain.Branch
					if err := tx.First(&branch, "id = ?", *branchID).Error; err == nil {
						branch.ManualCashBalance -= reimbursement.Amount
						_ = tx.Save(&branch)
					}
				}
				if mutation.CompanyID != nil && branchID != nil {
					var bcb domain.BranchCompanyBalance
					if err := tx.Where("branch_id = ? AND company_id = ?", *branchID, *mutation.CompanyID).First(&bcb).Error; err == nil {
						bcb.ManualCashBalance -= reimbursement.Amount
						_ = tx.Save(&bcb)
					}
				}
			} else if mutation.FundType == "BANK" && mutation.BankAccountID != nil {
				// 3. Bank
				var bank domain.CompanyBankAccount
				if err := tx.First(&bank, "id = ?", *mutation.BankAccountID).Error; err == nil {
					bank.ManualBalance -= reimbursement.Amount
					_ = tx.Save(&bank)
				}
			}
		}

		// Update linked BannerOrders if category is BANNER_INSTALLATION
		if reimbursement.Category == "BANNER_INSTALLATION" {
			now := time.Now()
			payRef := fmt.Sprintf("Reimburse: %s", reimbursement.Title)
			_ = tx.Model(&domain.BannerOrder{}).
				Where("reimbursement_id = ?", reimbursement.ID).
				Updates(map[string]interface{}{
					"is_paid":           true,
					"payment_status":    "paid",
					"paid_at":           &now,
					"payment_reference": &payRef,
					"updated_at":        now,
				}).Error
		}

		return nil
	})
}

func (r *reimbursementRepository) ApproveWithDeduction(reimbursement *domain.Reimbursement, branchID *uuid.UUID, mutation *domain.CashBankMutation, kasTx *domain.KasTransaction, kasPabrikID *uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 0. Lookup nama akun beban dari tabel accounts jika ExpenseAccountID terisi dan ExpenseAccountName kosong
		if reimbursement.ExpenseAccountID != nil && reimbursement.ExpenseAccountName == "" {
			var acc domain.Account
			if err := tx.First(&acc, "id = ?", *reimbursement.ExpenseAccountID).Error; err == nil {
				reimbursement.ExpenseAccountName = fmt.Sprintf("%s - %s", acc.Code, acc.Name)
			}
		}

		// 1. Simpan reimbursement update
		if err := tx.Save(reimbursement).Error; err != nil {
			return err
		}

		// 2. Jika mutasi cabang ada (BRANCH_PETTY_CASH, BRANCH_CASH, BANK_TRANSFER)
		if mutation != nil {
			if err := tx.Create(mutation).Error; err != nil {
				return err
			}

			// Petty Cash Cabang (Bukan Pabrik)
			if mutation.FundType == "PETTY_CASH" {
				if branchID != nil {
					var branch domain.Branch
					if err := tx.First(&branch, "id = ?", *branchID).Error; err == nil {
						branch.PettyCashBalance -= reimbursement.Amount
						_ = tx.Save(&branch)
					}
				}
				if mutation.CompanyID != nil && branchID != nil {
					var bcb domain.BranchCompanyBalance
					if err := tx.Where("branch_id = ? AND company_id = ?", *branchID, *mutation.CompanyID).First(&bcb).Error; err == nil {
						bcb.PettyCashBalance -= reimbursement.Amount
						_ = tx.Save(&bcb)
					}
				}
			} else if mutation.FundType == "CASH" {
				if branchID != nil {
					var branch domain.Branch
					if err := tx.First(&branch, "id = ?", *branchID).Error; err == nil {
						branch.ManualCashBalance -= reimbursement.Amount
						_ = tx.Save(&branch)
					}
				}
				if mutation.CompanyID != nil && branchID != nil {
					var bcb domain.BranchCompanyBalance
					if err := tx.Where("branch_id = ? AND company_id = ?", *branchID, *mutation.CompanyID).First(&bcb).Error; err == nil {
						bcb.ManualCashBalance -= reimbursement.Amount
						_ = tx.Save(&bcb)
					}
				}
			} else if mutation.FundType == "BANK" && mutation.BankAccountID != nil {
				var bank domain.CompanyBankAccount
				if err := tx.First(&bank, "id = ?", *mutation.BankAccountID).Error; err == nil {
					bank.ManualBalance -= reimbursement.Amount
					_ = tx.Save(&bank)
				}
			}
		}

		// 3. Jika Kas Pabrik ada (khusus cabang pabrik)
		if kasTx != nil && kasPabrikID != nil {
			var kas domain.KasPabrik
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", *kasPabrikID).Error; err != nil {
				return fmt.Errorf("KAS Pabrik tidak ditemukan: %w", err)
			}

			balBefore := kas.Balance
			balAfter := balBefore - reimbursement.Amount
			kas.Balance = balAfter
			if err := tx.Save(&kas).Error; err != nil {
				return err
			}

			kasTx.BalanceBefore = balBefore
			kasTx.BalanceAfter = balAfter
			kasTx.KasPabrikID = kas.ID

			// Validasi foreign key: kas_transactions.expense_account_id mengacu ke kas_expense_accounts
			if kasTx.ExpenseAccountID != nil {
				var count int64
				tx.Table("kas_expense_accounts").Where("id = ?", *kasTx.ExpenseAccountID).Count(&count)
				if count == 0 {
					kasTx.ExpenseAccountID = nil
				}
			}

			// Tambahkan info akun beban PSAK ke deskripsi jika belum tercantum
			if reimbursement.ExpenseAccountName != "" && !strings.Contains(kasTx.Description, reimbursement.ExpenseAccountName) {
				kasTx.Description = fmt.Sprintf("%s [Pos: %s]", kasTx.Description, reimbursement.ExpenseAccountName)
			}

			if err := tx.Create(kasTx).Error; err != nil {
				return err
			}
		}

		// 4. Update linked BannerOrders if category is BANNER_INSTALLATION
		if reimbursement.Category == "BANNER_INSTALLATION" {
			now := time.Now()
			payRef := fmt.Sprintf("Reimburse: %s", reimbursement.Title)
			_ = tx.Model(&domain.BannerOrder{}).
				Where("reimbursement_id = ?", reimbursement.ID).
				Updates(map[string]interface{}{
					"is_paid":           true,
					"payment_status":    "paid",
					"paid_at":           &now,
					"payment_reference": &payRef,
					"updated_at":        now,
				}).Error
		}

		return nil
	})
}

