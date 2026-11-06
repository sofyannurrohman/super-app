package usecase

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func createDepositTestTable(db *gorm.DB, model interface{}) error {
	parsed, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return err
	}
	columns := make([]string, 0, len(parsed.Fields))
	for _, field := range parsed.Fields {
		if field.DBName == "" {
			continue
		}
		kind := field.IndirectFieldType.Kind()
		dataType := "TEXT"
		if field.IndirectFieldType == reflect.TypeOf(time.Time{}) {
			dataType = "DATETIME"
		}
		switch kind {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dataType = "INTEGER"
		case reflect.Float32, reflect.Float64:
			dataType = "REAL"
		}
		primary := ""
		if field.PrimaryKey {
			primary = " PRIMARY KEY"
		}
		columns = append(columns, fmt.Sprintf("\"%s\" %s%s", field.DBName, dataType, primary))
	}
	return db.Exec(fmt.Sprintf("CREATE TABLE \"%s\" (%s)", parsed.Table, strings.Join(columns, ", "))).Error
}

func TestInvoiceDepositApprovalAllocatesInvoicesOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	for _, model := range []interface{}{
		&domain.Company{}, &domain.Branch{}, &domain.User{}, &domain.Employee{},
		&domain.Store{}, &domain.SalesTransaction{}, &domain.SalesOrder{},
		&domain.SalesPayment{}, &domain.CompanyBankAccount{},
		&domain.InvoiceDeposit{}, &domain.InvoiceDepositItem{},
	} {
		if err := createDepositTestTable(db, model); err != nil {
			t.Fatal(err)
		}
	}
	company := domain.Company{ID: uuid.New(), Name: "PT Sanke"}
	branch := domain.Branch{ID: uuid.New(), CompanyID: &company.ID, Name: "Cabang Sanke"}
	user := domain.User{ID: uuid.New(), Email: "kasir@example.com", Role: domain.RoleAdminKasir, IsActive: true}
	employee := domain.Employee{ID: uuid.New(), UserID: user.ID, BranchID: &branch.ID, FirstName: "Kasir"}
	bank := domain.CompanyBankAccount{ID: uuid.New(), CompanyID: &company.ID, BankName: "BCA", AccountNumber: "123", AccountHolder: "Pabrik", IsActive: true}
	for _, item := range []interface{}{&company, &branch, &user, &employee, &bank} {
		if err := db.Create(item).Error; err != nil {
			t.Fatal(err)
		}
	}
	date := time.Now()
	first := domain.SalesTransaction{ID: uuid.New(), CompanyID: company.ID, BranchID: &branch.ID, EmployeeID: employee.ID, ReceiptNo: "FP-001", TotalAmount: 100000, TransactionDate: date, PeriodMonth: int(date.Month()), PeriodYear: date.Year(), Status: "VERIFIED", PaymentStatus: "UNPAID", PaymentMethod: "TEMPO"}
	second := first
	second.ID = uuid.New()
	second.ReceiptNo = "FP-002"
	second.TotalAmount = 50000
	for _, invoice := range []*domain.SalesTransaction{&first, &second} {
		if err := db.Create(invoice).Error; err != nil {
			t.Fatal(err)
		}
	}
	u := &financeUsecase{db: db}
	req := SubmitInvoiceDepositRequest{
		BranchID: branch.ID, BankAccountID: bank.ID, DepositDate: date,
		ProofURL:    "/uploads/finance/proof.png",
		Allocations: []InvoiceDepositAllocation{{InvoiceID: first.ID, Amount: 60000}, {InvoiceID: second.ID, Amount: 50000}},
	}
	deposit, err := u.SubmitInvoiceDeposit(req, user.ID, string(domain.RoleAdminKasir))
	if err != nil {
		t.Fatal(err)
	}
	if deposit.Amount != 110000 {
		t.Fatalf("total setoran = %v, expected 110000", deposit.Amount)
	}
	if _, err := u.SubmitInvoiceDeposit(req, user.ID, string(domain.RoleAdminKasir)); err == nil {
		t.Fatal("duplicate pending allocation should be rejected")
	}
	if err := u.ReviewInvoiceDeposit(deposit.ID, uuid.New(), true); err != nil {
		t.Fatal(err)
	}
	if err := u.ReviewInvoiceDeposit(deposit.ID, uuid.New(), true); err == nil {
		t.Fatal("approved deposit must not credit the bank twice")
	}
	var updatedBank domain.CompanyBankAccount
	if err := db.First(&updatedBank, "id = ?", bank.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updatedBank.ManualBalance != 110000 {
		t.Fatalf("bank balance = %v, expected 110000", updatedBank.ManualBalance)
	}
	var payments []domain.SalesPayment
	if err := db.Where("sales_transaction_id IN ?", []uuid.UUID{first.ID, second.ID}).Find(&payments).Error; err != nil {
		t.Fatal(err)
	}
	if len(payments) != 2 {
		t.Fatalf("payment count = %d, expected 2", len(payments))
	}
	var updatedFirst, updatedSecond domain.SalesTransaction
	_ = db.First(&updatedFirst, "id = ?", first.ID).Error
	_ = db.First(&updatedSecond, "id = ?", second.ID).Error
	if updatedFirst.PaidAmount != 60000 || updatedFirst.PaymentStatus != "PARTIAL" || updatedSecond.PaidAmount != 50000 || updatedSecond.PaymentStatus != "PAID" {
		t.Fatalf("invoice states: first %v/%s, second %v/%s", updatedFirst.PaidAmount, updatedFirst.PaymentStatus, updatedSecond.PaidAmount, updatedSecond.PaymentStatus)
	}
}
