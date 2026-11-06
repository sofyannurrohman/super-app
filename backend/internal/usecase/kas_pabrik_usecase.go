package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"gorm.io/gorm"
)

type KasPabrikUsecase interface {
	GetAllByBranch(branchID string) ([]domain.KasPabrik, error)
	GetByID(id string) (*domain.KasPabrik, error)
	Create(req domain.CreateKasPabrikRequest) (*domain.KasPabrik, error)
	RecordTransaction(kasID string, req domain.CreateKasTransactionRequest, createdBy *uuid.UUID) (*domain.KasTransaction, error)
	SettleAdvance(txID string, req domain.SettleKasAdvanceRequest, settledBy *uuid.UUID) (*domain.KasAdvanceSettlement, error)
	Transfer(req domain.TransferKasRequest, createdBy *uuid.UUID) error
	ListTransactions(filter domain.ListKasTransactionsFilter) ([]domain.KasTransaction, int64, error)
	DeleteTransaction(txID string) error
	Update(id string, req domain.UpdateKasPabrikRequest) (*domain.KasPabrik, error)
	Delete(id string) error
	UpdateTransaction(id string, req domain.UpdateKasTransactionRequest) (*domain.KasTransaction, error)
	RecalculateKasLedger(kasID string) error
	GetPendingDeposits() ([]domain.CashBankMutation, error)
	ApproveDeposit(mutationID string, createdBy *uuid.UUID) error
	RejectDeposit(mutationID string, createdBy *uuid.UUID) error

	// Master Income Sources
	ListIncomeSources(onlyActive bool) ([]domain.KasIncomeSource, error)
	GetIncomeSourceByID(id string) (*domain.KasIncomeSource, error)
	CreateIncomeSource(req domain.CreateKasIncomeSourceRequest) (*domain.KasIncomeSource, error)
	UpdateIncomeSource(id string, req domain.UpdateKasIncomeSourceRequest) (*domain.KasIncomeSource, error)
	DeleteIncomeSource(id string) error

	// Master Expense Types
	ListExpenseTypes(onlyActive bool) ([]domain.KasExpenseType, error)
	GetExpenseTypeByID(id string) (*domain.KasExpenseType, error)
	CreateExpenseType(req domain.CreateKasExpenseTypeRequest) (*domain.KasExpenseType, error)
	UpdateExpenseType(id string, req domain.UpdateKasExpenseTypeRequest) (*domain.KasExpenseType, error)
	DeleteExpenseType(id string) error

	// Master Expense Accounts
	ListExpenseAccounts(typeID string, onlyActive bool) ([]domain.KasExpenseAccount, error)
	GetExpenseAccountByID(id string) (*domain.KasExpenseAccount, error)
	CreateExpenseAccount(req domain.CreateKasExpenseAccountRequest) (*domain.KasExpenseAccount, error)
	UpdateExpenseAccount(id string, req domain.UpdateKasExpenseAccountRequest) (*domain.KasExpenseAccount, error)
	DeleteExpenseAccount(id string) error
}

type kasPabrikUsecase struct {
	repo repository.KasPabrikRepository
	db   *gorm.DB
}

func NewKasPabrikUsecase(repo repository.KasPabrikRepository, db *gorm.DB) KasPabrikUsecase {
	return &kasPabrikUsecase{repo: repo, db: db}
}

func (u *kasPabrikUsecase) GetAllByBranch(branchID string) ([]domain.KasPabrik, error) {
	if branchID == "" {
		return u.repo.GetAll()
	}
	parsedBranchID, err := uuid.Parse(branchID)
	if err != nil {
		return nil, err
	}
	return u.repo.GetAllByBranch(parsedBranchID)
}

func (u *kasPabrikUsecase) GetByID(id string) (*domain.KasPabrik, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByID(parsedID)
}

func (u *kasPabrikUsecase) Create(req domain.CreateKasPabrikRequest) (*domain.KasPabrik, error) {
	parsedBranchID, err := uuid.Parse(req.BranchID)
	if err != nil {
		return nil, err
	}
	kas := &domain.KasPabrik{
		BranchID:      parsedBranchID,
		KasCode:       req.KasCode,
		KasName:       req.KasName,
		Description:   req.Description,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		SortOrder:     req.SortOrder,
		IsActive:      true,
	}

	if err := u.repo.Create(kas); err != nil {
		return nil, err
	}
	return kas, nil
}

func (u *kasPabrikUsecase) Update(id string, req domain.UpdateKasPabrikRequest) (*domain.KasPabrik, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	kas, err := u.repo.GetByID(parsedID)
	if err != nil {
		return nil, err
	}

	if req.KasCode != "" {
		kas.KasCode = req.KasCode
	}
	if req.KasName != "" {
		kas.KasName = req.KasName
	}
	if req.Description != "" {
		kas.Description = req.Description
	}
	if req.BankName != "" {
		kas.BankName = req.BankName
	}
	if req.AccountNumber != "" {
		kas.AccountNumber = req.AccountNumber
	}
	if req.SortOrder != 0 {
		kas.SortOrder = req.SortOrder
	}
	if req.IsActive != nil {
		kas.IsActive = *req.IsActive
	}

	if err := u.repo.Update(kas); err != nil {
		return nil, err
	}
	return kas, nil
}

func (u *kasPabrikUsecase) Delete(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return u.repo.Delete(parsedID)
}

func (u *kasPabrikUsecase) RecordTransaction(kasID string, req domain.CreateKasTransactionRequest, createdBy *uuid.UUID) (*domain.KasTransaction, error) {
	parsedKasID, err := uuid.Parse(kasID)
	if err != nil {
		return nil, err
	}

	txDate := time.Now()
	if req.TransactionDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.TransactionDate); err == nil {
			txDate = parsedDate
		}
	}

	if float64(req.Amount) <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	var transaction *domain.KasTransaction

	err = u.db.Transaction(func(tx *gorm.DB) error {
		var kas domain.KasPabrik
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", parsedKasID).Error; err != nil {
			return err
		}

		balanceBefore := kas.Balance
		balanceAfter := kas.Balance

		if req.TransactionType == "IN" {
			balanceAfter += float64(req.Amount)
		} else if req.TransactionType == "OUT" {
			if balanceBefore < float64(req.Amount) {
				return errors.New("Saldo tidak mencukupi")
			}
			balanceAfter -= float64(req.Amount)
		} else {
			return errors.New("invalid transaction type")
		}

		kas.Balance = balanceAfter
		if err := tx.Save(&kas).Error; err != nil {
			return err
		}

		advanceStatus := "NONE"
		if req.IsAdvance {
			advanceStatus = "OUTSTANDING"
		}

		var targetBranchIDPtr, vehicleIDPtr, employeeIDPtr, salesPaymentIDPtr, salesOrderIDPtr *uuid.UUID
		var incomeSourceIDPtr, expenseTypeIDPtr, expenseAccountIDPtr *uuid.UUID

		if req.TargetBranchID != "" {
			if bid, err := uuid.Parse(req.TargetBranchID); err == nil {
				targetBranchIDPtr = &bid
			}
		}
		if req.VehicleID != "" {
			if vid, err := uuid.Parse(req.VehicleID); err == nil {
				vehicleIDPtr = &vid
			}
		}
		if req.EmployeeID != "" {
			if eid, err := uuid.Parse(req.EmployeeID); err == nil {
				employeeIDPtr = &eid
			}
		}
		if req.SalesPaymentID != "" {
			if spid, err := uuid.Parse(req.SalesPaymentID); err == nil {
				salesPaymentIDPtr = &spid
			}
		}
		if req.SalesOrderID != "" {
			if soid, err := uuid.Parse(req.SalesOrderID); err == nil {
				salesOrderIDPtr = &soid
			}
		}

		incomeSourceStr := req.IncomeSource
		if req.IncomeSourceID != "" {
			if isid, err := uuid.Parse(req.IncomeSourceID); err == nil {
				incomeSourceIDPtr = &isid
				var isObj domain.KasIncomeSource
				if err := tx.First(&isObj, "id = ?", isid).Error; err == nil && incomeSourceStr == "" {
					incomeSourceStr = isObj.Code
				}
			}
		}

		expenseCategoryStr := req.ExpenseCategory
		if req.ExpenseTypeID != "" {
			if etid, err := uuid.Parse(req.ExpenseTypeID); err == nil {
				expenseTypeIDPtr = &etid
			}
		}
		if req.ExpenseAccountID != "" {
			if eaid, err := uuid.Parse(req.ExpenseAccountID); err == nil {
				expenseAccountIDPtr = &eaid
				var eaObj domain.KasExpenseAccount
				if err := tx.First(&eaObj, "id = ?", eaid).Error; err == nil && expenseCategoryStr == "" {
					expenseCategoryStr = eaObj.AccountName
				}
			}
		}

		transaction = &domain.KasTransaction{
			KasPabrikID:        parsedKasID,
			TransactionType:    req.TransactionType,
			Amount:             float64(req.Amount),
			BalanceBefore:      balanceBefore,
			BalanceAfter:       balanceAfter,
			IncomeSourceID:     incomeSourceIDPtr,
			ExpenseTypeID:      expenseTypeIDPtr,
			ExpenseAccountID:   expenseAccountIDPtr,
			IncomeSource:       incomeSourceStr,
			ExpenseCategory:    expenseCategoryStr,
			Description:        req.Description,
			ProofURL:           req.ProofURL,
			TransactionDate:    txDate,
			CreatedBy:          createdBy,
			IsBranchReceivable: req.IsBranchReceivable,
			TargetBranchID:     targetBranchIDPtr,
			VehicleID:          vehicleIDPtr,
			EmployeeID:         employeeIDPtr,
			AdvanceStatus:      advanceStatus,
			SalesPaymentID:     salesPaymentIDPtr,
			SalesOrderID:       salesOrderIDPtr,
		}

		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		if err := u.recalculateKasLedgerTx(tx, parsedKasID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (u *kasPabrikUsecase) SettleAdvance(txID string, req domain.SettleKasAdvanceRequest, settledBy *uuid.UUID) (*domain.KasAdvanceSettlement, error) {
	parsedID, err := uuid.Parse(txID)
	if err != nil {
		return nil, err
	}

	actualAmount := float64(req.ActualExpenseAmount)
	var settlement *domain.KasAdvanceSettlement

	err = u.db.Transaction(func(tx *gorm.DB) error {
		var transaction domain.KasTransaction
		if err := tx.First(&transaction, "id = ?", parsedID).Error; err != nil {
			return err
		}

		if transaction.AdvanceStatus != "OUTSTANDING" {
			return errors.New("transaksi ini bukan kasbon yang sedang outstanding")
		}

		var refundAmount, shortageAmount float64
		if actualAmount < transaction.Amount {
			refundAmount = transaction.Amount - actualAmount
		} else if actualAmount > transaction.Amount {
			shortageAmount = actualAmount - transaction.Amount
		}

		settlement = &domain.KasAdvanceSettlement{
			KasTransactionID:    transaction.ID,
			ActualExpenseAmount: actualAmount,
			RefundAmount:        refundAmount,
			ShortageAmount:      shortageAmount,
			Notes:               req.Notes,
			ProofURL:            req.ProofURL,
			SettledBy:           settledBy,
			SettledAt:           time.Now(),
		}

		if err := tx.Create(settlement).Error; err != nil {
			return err
		}

		transaction.AdvanceStatus = "SETTLED"
		if err := tx.Save(&transaction).Error; err != nil {
			return err
		}

		if refundAmount > 0 {
			var kas domain.KasPabrik
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", transaction.KasPabrikID).Error; err != nil {
				return err
			}

			balanceBefore := kas.Balance
			balanceAfter := kas.Balance + refundAmount
			kas.Balance = balanceAfter
			if err := tx.Save(&kas).Error; err != nil {
				return err
			}

			refundIncomeSource := "ADVANCE_REFUND"
			if transaction.IsBranchReceivable {
				refundIncomeSource = "BRANCH_RECEIVABLE_PAYMENT"
			}

			refundTx := domain.KasTransaction{
				KasPabrikID:          transaction.KasPabrikID,
				TransactionType:      "IN",
				Amount:               refundAmount,
				BalanceBefore:        balanceBefore,
				BalanceAfter:         balanceAfter,
				IncomeSource:         refundIncomeSource,
				RelatedTransactionID: &transaction.ID,
				TargetBranchID:       transaction.TargetBranchID,
				IsBranchReceivable:   transaction.IsBranchReceivable,
				Description:          "Pengembalian sisa kasbon/talangan: " + transaction.Description,
				ProofURL:             req.ProofURL,
				TransactionDate:      time.Now(),
				CreatedBy:            settledBy,
				AdvanceStatus:        "NONE",
			}
			if err := tx.Create(&refundTx).Error; err != nil {
				return err
			}
		}

		if err := u.recalculateKasLedgerTx(tx, transaction.KasPabrikID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return settlement, nil
}

func (u *kasPabrikUsecase) Transfer(req domain.TransferKasRequest, createdBy *uuid.UUID) error {
	fromKasID, err := uuid.Parse(req.FromKasID)
	if err != nil {
		return err
	}
	toKasID, err := uuid.Parse(req.ToKasID)
	if err != nil {
		return err
	}

	if fromKasID == toKasID {
		return errors.New("kas asal dan tujuan tidak boleh sama")
	}

	if float64(req.Amount) <= 0 {
		return errors.New("amount must be greater than 0")
	}

	txDate := time.Now()
	if req.TransactionDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.TransactionDate); err == nil {
			txDate = parsedDate
		}
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		var kasList []domain.KasPabrik
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id IN ?", []uuid.UUID{fromKasID, toKasID}).Find(&kasList).Error; err != nil {
			return err
		}

		if len(kasList) != 2 {
			return errors.New("kas not found")
		}

		var fromKas, toKas *domain.KasPabrik
		for i := range kasList {
			if kasList[i].ID == fromKasID {
				fromKas = &kasList[i]
			} else if kasList[i].ID == toKasID {
				toKas = &kasList[i]
			}
		}

		if fromKas.Balance < float64(req.Amount) {
			return errors.New("Saldo tidak mencukupi")
		}

		fromBalanceBefore := fromKas.Balance
		fromBalanceAfter := fromKas.Balance - float64(req.Amount)
		fromKas.Balance = fromBalanceAfter
		if err := tx.Save(fromKas).Error; err != nil {
			return err
		}

		toBalanceBefore := toKas.Balance
		toBalanceAfter := toKas.Balance + float64(req.Amount)
		toKas.Balance = toBalanceAfter
		if err := tx.Save(toKas).Error; err != nil {
			return err
		}

		txOutID := uuid.New()
		txInID := uuid.New()

		txOut := domain.KasTransaction{
			ID:                 txOutID,
			KasPabrikID:        fromKasID,
			TransactionType:    "TRANSFER_OUT",
			Amount:             float64(req.Amount),
			BalanceBefore:      fromBalanceBefore,
			BalanceAfter:       fromBalanceAfter,
			RelatedKasPabrikID: &toKasID,
			Description:        req.Description,
			TransactionDate:    txDate,
			CreatedBy:          createdBy,
		}
		if err := tx.Create(&txOut).Error; err != nil {
			return err
		}

		txIn := domain.KasTransaction{
			ID:                   txInID,
			KasPabrikID:          toKasID,
			TransactionType:      "TRANSFER_IN",
			Amount:               float64(req.Amount),
			BalanceBefore:        toBalanceBefore,
			BalanceAfter:         toBalanceAfter,
			RelatedKasPabrikID:   &fromKasID,
			RelatedTransactionID: &txOutID,
			Description:          req.Description,
			TransactionDate:      txDate,
			CreatedBy:            createdBy,
		}
		if err := tx.Create(&txIn).Error; err != nil {
			return err
		}

		if err := tx.Model(&txOut).Update("related_transaction_id", txInID).Error; err != nil {
			return err
		}

		if err := u.recalculateKasLedgerTx(tx, fromKasID); err != nil {
			return err
		}
		if err := u.recalculateKasLedgerTx(tx, toKasID); err != nil {
			return err
		}

		return nil
	})
}

func (u *kasPabrikUsecase) ListTransactions(filter domain.ListKasTransactionsFilter) ([]domain.KasTransaction, int64, error) {
	return u.repo.ListTransactions(filter)
}

func (u *kasPabrikUsecase) DeleteTransaction(txID string) error {
	parsedID, err := uuid.Parse(txID)
	if err != nil {
		return err
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		var transaction domain.KasTransaction
		if err := tx.First(&transaction, "id = ?", parsedID).Error; err != nil {
			return err
		}

		var transactionsToDelete []domain.KasTransaction
		transactionsToDelete = append(transactionsToDelete, transaction)

		if transaction.RelatedTransactionID != nil {
			var relatedTx domain.KasTransaction
			if err := tx.First(&relatedTx, "id = ?", *transaction.RelatedTransactionID).Error; err == nil {
				transactionsToDelete = append(transactionsToDelete, relatedTx)
			}
		}

		var txIDs []uuid.UUID
		for _, t := range transactionsToDelete {
			txIDs = append(txIDs, t.ID)
		}
		if len(txIDs) > 0 {
			// Unlink self/cross references in kas_transactions
			if err := tx.Model(&domain.KasTransaction{}).Where("id IN ?", txIDs).Update("related_transaction_id", nil).Error; err != nil {
				return err
			}
			if err := tx.Model(&domain.KasTransaction{}).Where("related_transaction_id IN ?", txIDs).Update("related_transaction_id", nil).Error; err != nil {
				return err
			}

			// Clean up any settlements referencing these transactions
			if err := tx.Where("kas_transaction_id IN ?", txIDs).Delete(&domain.KasAdvanceSettlement{}).Error; err != nil {
				return err
			}

			// Unlink references in other tables if any
			_ = tx.Model(&domain.DailyWorkerSalary{}).Where("kas_transaction_id IN ?", txIDs).Update("kas_transaction_id", nil).Error
			_ = tx.Model(&domain.RawMaterialReceipt{}).Where("kas_transaction_id IN ?", txIDs).Update("kas_transaction_id", nil).Error
		}

		for _, t := range transactionsToDelete {
			var k domain.KasPabrik
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&k, "id = ?", t.KasPabrikID).Error; err != nil {
				return err
			}

			if t.TransactionType == "IN" || t.TransactionType == "TRANSFER_IN" {
				k.Balance -= t.Amount
			} else if t.TransactionType == "OUT" || t.TransactionType == "TRANSFER_OUT" {
				k.Balance += t.Amount
			}

			if err := tx.Save(&k).Error; err != nil {
				return err
			}

			if err := tx.Delete(&t).Error; err != nil {
				return err
			}
		}

		for _, t := range transactionsToDelete {
			if err := u.recalculateKasLedgerTx(tx, t.KasPabrikID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (u *kasPabrikUsecase) UpdateTransaction(id string, req domain.UpdateKasTransactionRequest) (*domain.KasTransaction, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var updatedTx *domain.KasTransaction

	err = u.db.Transaction(func(tx *gorm.DB) error {
		var transaction domain.KasTransaction
		if err := tx.First(&transaction, "id = ?", parsedID).Error; err != nil {
			return err
		}

		var transactionsToUpdate []domain.KasTransaction
		transactionsToUpdate = append(transactionsToUpdate, transaction)

		if transaction.RelatedTransactionID != nil {
			var relatedTx domain.KasTransaction
			if err := tx.First(&relatedTx, "id = ?", *transaction.RelatedTransactionID).Error; err == nil {
				transactionsToUpdate = append(transactionsToUpdate, relatedTx)
			}
		}

		amountDiff := float64(req.Amount) - transaction.Amount
		if amountDiff == 0 && req.TransactionDate == "" && req.Description == "" && req.IncomeSource == "" && req.ExpenseCategory == "" && req.ProofURL == "" && req.IncomeSourceID == "" && req.ExpenseTypeID == "" && req.ExpenseAccountID == "" {
			updatedTx = &transaction
			return nil
		}

		for i, t := range transactionsToUpdate {
			var k domain.KasPabrik
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&k, "id = ?", t.KasPabrikID).Error; err != nil {
				return err
			}

			if amountDiff != 0 {
				if t.TransactionType == "IN" || t.TransactionType == "TRANSFER_IN" {
					k.Balance += amountDiff
				} else if t.TransactionType == "OUT" || t.TransactionType == "TRANSFER_OUT" {
					k.Balance -= amountDiff
				}

				if err := tx.Save(&k).Error; err != nil {
					return err
				}

				t.Amount = float64(req.Amount)
				if t.TransactionType == "IN" || t.TransactionType == "TRANSFER_IN" {
					t.BalanceAfter = t.BalanceBefore + float64(req.Amount)
				} else {
					t.BalanceAfter = t.BalanceBefore - float64(req.Amount)
				}
			}

			if req.Description != "" {
				t.Description = req.Description
			}
			if req.TransactionDate != "" {
				if parsedDate, err := time.Parse("2006-01-02", req.TransactionDate); err == nil {
					t.TransactionDate = parsedDate
				}
			}

			if t.ID == parsedID {
				if req.IncomeSourceID != "" {
					if isid, err := uuid.Parse(req.IncomeSourceID); err == nil {
						t.IncomeSourceID = &isid
						var isObj domain.KasIncomeSource
						if err := tx.First(&isObj, "id = ?", isid).Error; err == nil && req.IncomeSource == "" {
							t.IncomeSource = isObj.Code
						}
					}
				}
				if req.ExpenseTypeID != "" {
					if etid, err := uuid.Parse(req.ExpenseTypeID); err == nil {
						t.ExpenseTypeID = &etid
					}
				}
				if req.ExpenseAccountID != "" {
					if eaid, err := uuid.Parse(req.ExpenseAccountID); err == nil {
						t.ExpenseAccountID = &eaid
						var eaObj domain.KasExpenseAccount
						if err := tx.First(&eaObj, "id = ?", eaid).Error; err == nil && req.ExpenseCategory == "" {
							t.ExpenseCategory = eaObj.AccountName
						}
					}
				}
				if req.IncomeSource != "" {
					t.IncomeSource = req.IncomeSource
				}
				if req.ExpenseCategory != "" {
					t.ExpenseCategory = req.ExpenseCategory
				}
				if req.ProofURL != "" {
					t.ProofURL = req.ProofURL
				}
			}

			if err := tx.Save(&t).Error; err != nil {
				return err
			}

			if i == 0 {
				updatedTx = &t
			}
		}

		for _, t := range transactionsToUpdate {
			if err := u.recalculateKasLedgerTx(tx, t.KasPabrikID); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedTx, nil
}

func (u *kasPabrikUsecase) RecalculateKasLedger(kasID string) error {
	parsedID, err := uuid.Parse(kasID)
	if err != nil {
		return err
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		return u.recalculateKasLedgerTx(tx, parsedID)
	})
}

func (u *kasPabrikUsecase) recalculateKasLedgerTx(tx *gorm.DB, kasID uuid.UUID) error {
	var kas domain.KasPabrik
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", kasID).Error; err != nil {
		return err
	}

	var transactions []domain.KasTransaction
	if err := tx.Where("kas_pabrik_id = ?", kasID).
		Order("transaction_date asc, created_at asc").
		Find(&transactions).Error; err != nil {
		return err
	}

	if len(transactions) == 0 {
		kas.Balance = 0
		return tx.Model(&kas).Update("balance", 0).Error
	}

	runningBalance := transactions[0].BalanceBefore

	for i := range transactions {
		t := &transactions[i]
		t.BalanceBefore = runningBalance
		if t.TransactionType == "IN" || t.TransactionType == "TRANSFER_IN" {
			runningBalance += t.Amount
		} else if t.TransactionType == "OUT" || t.TransactionType == "TRANSFER_OUT" {
			runningBalance -= t.Amount
		}
		t.BalanceAfter = runningBalance

		if err := tx.Model(t).Updates(map[string]interface{}{
			"balance_before": t.BalanceBefore,
			"balance_after":  t.BalanceAfter,
		}).Error; err != nil {
			return err
		}
	}

	kas.Balance = runningBalance
	return tx.Model(&kas).Update("balance", runningBalance).Error
}

func (u *kasPabrikUsecase) GetPendingDeposits() ([]domain.CashBankMutation, error) {
	var mutations []domain.CashBankMutation
	err := u.db.Preload("Branch").
		Preload("TargetKas").
		Where("target_kas_pabrik_id IS NOT NULL AND status_setoran = ?", "PENDING").
		Order("created_at DESC").
		Find(&mutations).Error
	if err != nil {
		return nil, err
	}
	return mutations, nil
}

func (u *kasPabrikUsecase) ApproveDeposit(mutationID string, createdBy *uuid.UUID) error {
	parsedMutationID, err := uuid.Parse(mutationID)
	if err != nil {
		return err
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		var mutation domain.CashBankMutation
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&mutation, "id = ?", parsedMutationID).Error; err != nil {
			return err
		}

		if mutation.TargetKasPabrikID == nil || mutation.StatusSetoran != "PENDING" {
			return errors.New("mutation is not a pending deposit")
		}

		mutation.StatusSetoran = "APPROVED"
		if err := tx.Save(&mutation).Error; err != nil {
			return err
		}

		var kas domain.KasPabrik
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&kas, "id = ?", *mutation.TargetKasPabrikID).Error; err != nil {
			return err
		}

		balanceBefore := kas.Balance
		balanceAfter := kas.Balance + mutation.Amount
		kas.Balance = balanceAfter
		if err := tx.Save(&kas).Error; err != nil {
			return err
		}

		txIn := domain.KasTransaction{
			ID:              uuid.New(),
			KasPabrikID:     kas.ID,
			TransactionType: "IN",
			Amount:          mutation.Amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceAfter,
			IncomeSource:    "Setoran Cabang",
			Description:     "Setoran dari " + mutation.Notes,
			ProofURL:        mutation.ProofURL,
			TransactionDate: time.Now(),
			CreatedBy:       createdBy,
		}
		if err := tx.Create(&txIn).Error; err != nil {
			return err
		}

		if err := u.recalculateKasLedgerTx(tx, kas.ID); err != nil {
			return err
		}

		return nil
	})
}

func (u *kasPabrikUsecase) RejectDeposit(mutationID string, createdBy *uuid.UUID) error {
	parsedMutationID, err := uuid.Parse(mutationID)
	if err != nil {
		return err
	}

	return u.db.Transaction(func(tx *gorm.DB) error {
		var mutation domain.CashBankMutation
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&mutation, "id = ?", parsedMutationID).Error; err != nil {
			return err
		}

		if mutation.TargetKasPabrikID == nil || mutation.StatusSetoran != "PENDING" {
			return errors.New("mutation is not a pending deposit")
		}

		mutation.StatusSetoran = "REJECTED"
		if err := tx.Save(&mutation).Error; err != nil {
			return err
		}

		return nil
	})
}

// Master Income Sources Usecase
func (u *kasPabrikUsecase) ListIncomeSources(onlyActive bool) ([]domain.KasIncomeSource, error) {
	return u.repo.ListIncomeSources(onlyActive)
}

func (u *kasPabrikUsecase) GetIncomeSourceByID(id string) (*domain.KasIncomeSource, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return u.repo.GetIncomeSourceByID(parsedID)
}

func (u *kasPabrikUsecase) CreateIncomeSource(req domain.CreateKasIncomeSourceRequest) (*domain.KasIncomeSource, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	src := &domain.KasIncomeSource{
		Code:        req.Code,
		Name:        req.Name,
		AccountCode: req.AccountCode,
		AccountName: req.AccountName,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    isActive,
	}
	if err := u.repo.CreateIncomeSource(src); err != nil {
		return nil, err
	}
	return src, nil
}

func (u *kasPabrikUsecase) UpdateIncomeSource(id string, req domain.UpdateKasIncomeSourceRequest) (*domain.KasIncomeSource, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	src, err := u.repo.GetIncomeSourceByID(parsedID)
	if err != nil {
		return nil, err
	}
	if req.Code != "" {
		src.Code = req.Code
	}
	if req.Name != "" {
		src.Name = req.Name
	}
	if req.AccountCode != "" {
		src.AccountCode = req.AccountCode
	}
	if req.AccountName != "" {
		src.AccountName = req.AccountName
	}
	if req.Description != "" {
		src.Description = req.Description
	}
	if req.SortOrder != 0 {
		src.SortOrder = req.SortOrder
	}
	if req.IsActive != nil {
		src.IsActive = *req.IsActive
	}
	if err := u.repo.UpdateIncomeSource(src); err != nil {
		return nil, err
	}
	return src, nil
}

func (u *kasPabrikUsecase) DeleteIncomeSource(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return u.repo.DeleteIncomeSource(parsedID)
}

// Master Expense Types Usecase
func (u *kasPabrikUsecase) ListExpenseTypes(onlyActive bool) ([]domain.KasExpenseType, error) {
	return u.repo.ListExpenseTypes(onlyActive)
}

func (u *kasPabrikUsecase) GetExpenseTypeByID(id string) (*domain.KasExpenseType, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return u.repo.GetExpenseTypeByID(parsedID)
}

func (u *kasPabrikUsecase) CreateExpenseType(req domain.CreateKasExpenseTypeRequest) (*domain.KasExpenseType, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	et := &domain.KasExpenseType{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    isActive,
	}
	if err := u.repo.CreateExpenseType(et); err != nil {
		return nil, err
	}
	return et, nil
}

func (u *kasPabrikUsecase) UpdateExpenseType(id string, req domain.UpdateKasExpenseTypeRequest) (*domain.KasExpenseType, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	et, err := u.repo.GetExpenseTypeByID(parsedID)
	if err != nil {
		return nil, err
	}
	if req.Code != "" {
		et.Code = req.Code
	}
	if req.Name != "" {
		et.Name = req.Name
	}
	if req.Description != "" {
		et.Description = req.Description
	}
	if req.SortOrder != 0 {
		et.SortOrder = req.SortOrder
	}
	if req.IsActive != nil {
		et.IsActive = *req.IsActive
	}
	if err := u.repo.UpdateExpenseType(et); err != nil {
		return nil, err
	}
	return et, nil
}

func (u *kasPabrikUsecase) DeleteExpenseType(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return u.repo.DeleteExpenseType(parsedID)
}

// Master Expense Accounts Usecase
func (u *kasPabrikUsecase) ListExpenseAccounts(typeID string, onlyActive bool) ([]domain.KasExpenseAccount, error) {
	var typeIDPtr *uuid.UUID
	if typeID != "" {
		if pid, err := uuid.Parse(typeID); err == nil {
			typeIDPtr = &pid
		}
	}
	return u.repo.ListExpenseAccounts(typeIDPtr, onlyActive)
}

func (u *kasPabrikUsecase) GetExpenseAccountByID(id string) (*domain.KasExpenseAccount, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return u.repo.GetExpenseAccountByID(parsedID)
}

func (u *kasPabrikUsecase) CreateExpenseAccount(req domain.CreateKasExpenseAccountRequest) (*domain.KasExpenseAccount, error) {
	parsedTypeID, err := uuid.Parse(req.ExpenseTypeID)
	if err != nil {
		return nil, errors.New("invalid expense_type_id")
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	acc := &domain.KasExpenseAccount{
		ExpenseTypeID:     parsedTypeID,
		AccountCode:       req.AccountCode,
		AccountName:       req.AccountName,
		AccurateAccountID: req.AccurateAccountID,
		Description:       req.Description,
		SortOrder:         req.SortOrder,
		IsActive:          isActive,
	}
	if err := u.repo.CreateExpenseAccount(acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (u *kasPabrikUsecase) UpdateExpenseAccount(id string, req domain.UpdateKasExpenseAccountRequest) (*domain.KasExpenseAccount, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	acc, err := u.repo.GetExpenseAccountByID(parsedID)
	if err != nil {
		return nil, err
	}
	if req.ExpenseTypeID != "" {
		if parsedTypeID, err := uuid.Parse(req.ExpenseTypeID); err == nil {
			acc.ExpenseTypeID = parsedTypeID
			acc.ExpenseType = nil
		}
	}
	if req.AccountCode != "" {
		acc.AccountCode = req.AccountCode
	}
	if req.AccountName != "" {
		acc.AccountName = req.AccountName
	}
	if req.AccurateAccountID != "" {
		acc.AccurateAccountID = req.AccurateAccountID
	}
	if req.Description != "" {
		acc.Description = req.Description
	}
	if req.SortOrder != 0 {
		acc.SortOrder = req.SortOrder
	}
	if req.IsActive != nil {
		acc.IsActive = *req.IsActive
	}
	if err := u.repo.UpdateExpenseAccount(acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (u *kasPabrikUsecase) DeleteExpenseAccount(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return u.repo.DeleteExpenseAccount(parsedID)
}
