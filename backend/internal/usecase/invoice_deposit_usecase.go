package usecase

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InvoiceDepositAvailable struct {
	Invoice         domain.SalesTransaction `json:"invoice"`
	AvailableAmount float64                 `json:"available_amount"`
}

type InvoiceDepositAllocation struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
	Amount    float64   `json:"amount"`
}

type SubmitInvoiceDepositRequest struct {
	BranchID      uuid.UUID                  `json:"branch_id"`
	BankAccountID uuid.UUID                  `json:"bank_account_id"`
	DepositDate   time.Time                  `json:"deposit_date"`
	ProofURL      string                     `json:"proof_url"`
	Notes         string                     `json:"notes"`
	Allocations   []InvoiceDepositAllocation `json:"allocations"`
}

func (u *financeUsecase) GetInvoiceDepositBranch(userID uuid.UUID, role string) (*uuid.UUID, error) {
	if role != string(domain.RoleCashier) && role != string(domain.RoleAdminKasir) {
		return nil, nil
	}
	var cashier domain.Employee
	if err := u.db.Where("user_id = ?", userID).First(&cashier).Error; err != nil {
		return nil, errors.New("data karyawan Admin Kasir tidak ditemukan")
	}
	branchID := cashier.BranchID
	if cashier.OperatingBranchID != nil {
		branchID = cashier.OperatingBranchID
	}
	if branchID == nil {
		return nil, errors.New("cabang Admin Kasir belum ditentukan")
	}
	return branchID, nil
}

func invoiceDepositEligible(tx *gorm.DB, invoice *domain.SalesTransaction) bool {
	if invoice.Status == "VERIFIED" || invoice.Status == "APPROVED" {
		return true
	}
	var count int64
	tx.Model(&domain.SalesOrder{}).Where("invoice_id = ? AND admin_nota_confirmed_at IS NOT NULL", invoice.ID).Count(&count)
	return count > 0
}

func invoiceUnpaidAmount(tx *gorm.DB, invoice *domain.SalesTransaction) (float64, error) {
	var logged struct {
		Success float64
		Pending float64
	}
	if err := tx.Model(&domain.SalesPayment{}).
		Where("sales_transaction_id = ? AND payment_status IN ?", invoice.ID, []string{"PENDING", "SUCCESS"}).
		Select("COALESCE(SUM(CASE WHEN payment_status = 'SUCCESS' THEN amount ELSE 0 END), 0) AS success, COALESCE(SUM(CASE WHEN payment_status = 'PENDING' THEN amount ELSE 0 END), 0) AS pending").Scan(&logged).Error; err != nil {
		return 0, err
	}
	return math.Max(0, invoice.TotalAmount-math.Max(invoice.PaidAmount, logged.Success)-logged.Pending), nil
}

func invoiceAvailableForDeposit(tx *gorm.DB, invoice *domain.SalesTransaction) (float64, error) {
	unpaid, err := invoiceUnpaidAmount(tx, invoice)
	if err != nil {
		return 0, err
	}
	var pending float64
	if err := tx.Table("invoice_deposit_items AS item").
		Joins("JOIN invoice_deposits AS deposit ON deposit.id = item.deposit_id").
		Where("item.sales_transaction_id = ? AND deposit.status = 'PENDING'", invoice.ID).
		Select("COALESCE(SUM(item.amount), 0)").Scan(&pending).Error; err != nil {
		return 0, err
	}
	return math.Max(0, unpaid-pending), nil
}

func (u *financeUsecase) GetAvailableDepositInvoices(branchID, companyID *uuid.UUID, search string) ([]InvoiceDepositAvailable, error) {
	query := u.db.Model(&domain.SalesTransaction{}).
		Joins("LEFT JOIN employees AS invoice_employee ON invoice_employee.id = sales_transactions.employee_id").
		Preload("Store").Preload("Branch").Preload("Company").
		Where("sales_transactions.total_amount > COALESCE(sales_transactions.paid_amount, 0) AND sales_transactions.status NOT IN ?", []string{"REJECTED", "DELETED"}).
		Where("(sales_transactions.status IN ? OR EXISTS (SELECT 1 FROM sales_orders so WHERE so.invoice_id = sales_transactions.id AND so.admin_nota_confirmed_at IS NOT NULL))", []string{"VERIFIED", "APPROVED"})
	if branchID != nil {
		query = query.Where("COALESCE(sales_transactions.branch_id, invoice_employee.operating_branch_id, invoice_employee.branch_id) = ?", *branchID)
	}
	if companyID != nil {
		query = query.Where("sales_transactions.company_id = ?", *companyID)
	}
	if search = strings.TrimSpace(search); search != "" {
		like := "%" + search + "%"
		query = query.Joins("LEFT JOIN stores AS invoice_store ON invoice_store.id = sales_transactions.store_id").
			Where("sales_transactions.receipt_no ILIKE ? OR sales_transactions.buyer_name ILIKE ? OR invoice_store.name ILIKE ?", like, like, like)
	}
	var invoices []domain.SalesTransaction
	if err := query.Order("sales_transactions.transaction_date DESC").Limit(200).Find(&invoices).Error; err != nil {
		return nil, err
	}
	available := make([]InvoiceDepositAvailable, 0, len(invoices))
	for i := range invoices {
		if !invoiceDepositEligible(u.db, &invoices[i]) {
			continue
		}
		remaining, err := invoiceAvailableForDeposit(u.db, &invoices[i])
		if err != nil {
			return nil, err
		}
		if remaining > 0.009 {
			available = append(available, InvoiceDepositAvailable{Invoice: invoices[i], AvailableAmount: remaining})
		}
	}
	return available, nil
}

func (u *financeUsecase) ListInvoiceDeposits(branchID *uuid.UUID) ([]domain.InvoiceDeposit, error) {
	var deposits []domain.InvoiceDeposit
	query := u.db.Preload("Branch").Preload("Company").Preload("BankAccount").
		Preload("Items").Preload("Items.SalesTransaction").Preload("Items.SalesTransaction.Store")
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	err := query.Order("created_at DESC").Limit(200).Find(&deposits).Error
	return deposits, err
}

func (u *financeUsecase) SubmitInvoiceDeposit(req SubmitInvoiceDepositRequest, submittedBy uuid.UUID, role string) (*domain.InvoiceDeposit, error) {
	if req.BranchID == uuid.Nil || req.BankAccountID == uuid.Nil || len(req.Allocations) == 0 || submittedBy == uuid.Nil {
		return nil, errors.New("cabang, rekening bank, dan minimal satu faktur wajib dipilih")
	}
	if strings.TrimSpace(req.ProofURL) == "" {
		return nil, errors.New("bukti setoran wajib dilampirkan")
	}
	if req.DepositDate.IsZero() {
		req.DepositDate = time.Now()
	}
	var result domain.InvoiceDeposit
	err := u.db.Transaction(func(tx *gorm.DB) error {
		var bank domain.CompanyBankAccount
		if err := tx.First(&bank, "id = ? AND is_active = ?", req.BankAccountID, true).Error; err != nil {
			return errors.New("rekening bank tujuan tidak ditemukan atau tidak aktif")
		}
		var branch domain.Branch
		if err := tx.First(&branch, "id = ?", req.BranchID).Error; err != nil {
			return errors.New("cabang tidak ditemukan")
		}
		var cashier domain.Employee
		if err := tx.Where("user_id = ?", submittedBy).First(&cashier).Error; err != nil {
			return errors.New("data karyawan Admin Kasir tidak ditemukan")
		}
		if role == string(domain.RoleCashier) || role == string(domain.RoleAdminKasir) {
			cashierBranch := cashier.BranchID
			if cashier.OperatingBranchID != nil {
				cashierBranch = cashier.OperatingBranchID
			}
			if cashierBranch == nil || *cashierBranch != req.BranchID {
				return errors.New("setoran hanya dapat dibuat untuk cabang Admin Kasir")
			}
		}
		seen := make(map[uuid.UUID]bool, len(req.Allocations))
		items := make([]domain.InvoiceDepositItem, 0, len(req.Allocations))
		var companyID uuid.UUID
		var total float64
		for _, allocation := range req.Allocations {
			if allocation.InvoiceID == uuid.Nil || seen[allocation.InvoiceID] || allocation.Amount <= 0 || math.IsNaN(allocation.Amount) || math.IsInf(allocation.Amount, 0) || math.Abs(allocation.Amount*100-math.Round(allocation.Amount*100)) > 0.000001 {
				return errors.New("rincian faktur atau nominal setoran tidak valid")
			}
			seen[allocation.InvoiceID] = true
			var invoice domain.SalesTransaction
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&invoice, "id = ?", allocation.InvoiceID).Error; err != nil {
				return fmt.Errorf("faktur %s tidak ditemukan", allocation.InvoiceID)
			}
			if !invoiceDepositEligible(tx, &invoice) {
				return fmt.Errorf("faktur %s belum disahkan Admin Nota", invoice.ReceiptNo)
			}
			invoiceBranch := invoice.BranchID
			if invoiceBranch == nil {
				var salesman domain.Employee
				if err := tx.First(&salesman, "id = ?", invoice.EmployeeID).Error; err != nil {
					return err
				}
				invoiceBranch = salesman.BranchID
				if salesman.OperatingBranchID != nil {
					invoiceBranch = salesman.OperatingBranchID
				}
			}
			if invoiceBranch == nil || *invoiceBranch != req.BranchID {
				return fmt.Errorf("faktur %s berasal dari cabang lain", invoice.ReceiptNo)
			}
			if companyID == uuid.Nil {
				companyID = invoice.CompanyID
			} else if invoice.CompanyID != companyID {
				return errors.New("satu setoran hanya boleh berisi faktur dari PT yang sama")
			}
			available, err := invoiceAvailableForDeposit(tx, &invoice)
			if err != nil {
				return err
			}
			if allocation.Amount > available+0.009 {
				return fmt.Errorf("setoran faktur %s melebihi sisa tagihan", invoice.ReceiptNo)
			}
			total += allocation.Amount
			items = append(items, domain.InvoiceDepositItem{ID: uuid.New(), SalesTransactionID: invoice.ID, Amount: allocation.Amount})
		}
		if bank.CompanyID != nil && *bank.CompanyID != companyID {
			return errors.New("rekening tujuan tidak terdaftar untuk PT pada faktur")
		}
		if bank.BranchID != nil {
			return errors.New("rekening tujuan harus merupakan rekening pabrik, bukan rekening cabang")
		}
		result = domain.InvoiceDeposit{
			ID: uuid.New(), BranchID: req.BranchID, CompanyID: companyID,
			BankAccountID: bank.ID, DepositDate: req.DepositDate, Amount: total,
			ProofURL: req.ProofURL, Notes: req.Notes, Status: "PENDING", SubmittedByID: submittedBy,
		}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].DepositID = result.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		result.Items = items
		return nil
	})
	return &result, err
}

func (u *financeUsecase) ReviewInvoiceDeposit(id, reviewedBy uuid.UUID, approve bool) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		var deposit domain.InvoiceDeposit
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&deposit, "id = ?", id).Error; err != nil {
			return err
		}
		if deposit.Status != "PENDING" {
			return errors.New("setoran ini sudah diproses")
		}
		now := time.Now()
		deposit.ReviewedAt = &now
		deposit.ReviewedByID = &reviewedBy
		if !approve {
			deposit.Status = "REJECTED"
			return tx.Save(&deposit).Error
		}
		var bank domain.CompanyBankAccount
		if err := tx.First(&bank, "id = ? AND is_active = ?", deposit.BankAccountID, true).Error; err != nil {
			return errors.New("rekening bank tujuan tidak ditemukan atau tidak aktif")
		}
		if (bank.CompanyID != nil && *bank.CompanyID != deposit.CompanyID) || bank.BranchID != nil {
			return errors.New("rekening tujuan tidak sesuai dengan PT pabrik")
		}
		var cashier domain.Employee
		if err := tx.Where("user_id = ?", deposit.SubmittedByID).First(&cashier).Error; err != nil {
			return errors.New("data karyawan Admin Kasir tidak ditemukan")
		}
		for _, item := range deposit.Items {
			var invoice domain.SalesTransaction
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&invoice, "id = ?", item.SalesTransactionID).Error; err != nil {
				return err
			}
			// The deposit itself reserves this amount while it is pending.
			available, err := invoiceUnpaidAmount(tx, &invoice)
			if err != nil {
				return err
			}
			if item.Amount > available+0.009 {
				return fmt.Errorf("sisa faktur %s berubah; tinjau ulang setoran", invoice.ReceiptNo)
			}
			payment := domain.SalesPayment{
				SalesTransactionID: &invoice.ID, EmployeeID: cashier.ID, Amount: item.Amount,
				PaymentMethod: "TRANSFER", PaymentStatus: "SUCCESS", IsVerified: true,
				PaymentDate: deposit.DepositDate, CompanyBankAccountID: &deposit.BankAccountID,
				TransferProofURL: &deposit.ProofURL,
				Notes:            fmt.Sprintf("Setoran faktur ke pabrik %s", deposit.ID), ApprovedAt: &now,
			}
			if err := tx.Create(&payment).Error; err != nil {
				return err
			}
			invoice.PaidAmount += item.Amount
			if invoice.PaidAmount+0.009 >= invoice.TotalAmount {
				invoice.PaymentStatus = "PAID"
				invoice.PaymentDueDate = nil
			} else {
				invoice.PaymentStatus = "PARTIAL"
			}
			if err := tx.Model(&invoice).Updates(map[string]interface{}{
				"paid_amount": invoice.PaidAmount, "payment_status": invoice.PaymentStatus,
				"payment_due_date": invoice.PaymentDueDate,
			}).Error; err != nil {
				return err
			}
			var so domain.SalesOrder
			if err := tx.Where("invoice_id = ?", invoice.ID).First(&so).Error; err == nil {
				newPaid := so.PaymentCollectedAmount + item.Amount
				status := "PARTIAL"
				if newPaid+0.009 >= so.TotalAmount {
					status = "PAID"
				}
				if err := tx.Model(&so).Updates(map[string]interface{}{
					"payment_collected_amount": newPaid, "payment_status": status,
				}).Error; err != nil {
					return err
				}
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if err := tx.Model(&domain.CompanyBankAccount{}).Where("id = ?", bank.ID).
			Updates(map[string]interface{}{
				"manual_balance":      gorm.Expr("COALESCE(manual_balance, 0) + ?", deposit.Amount),
				"last_balance_update": now,
			}).Error; err != nil {
			return err
		}
		deposit.Status = "APPROVED"
		return tx.Save(&deposit).Error
	})
}
