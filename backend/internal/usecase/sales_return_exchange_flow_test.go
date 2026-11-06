package usecase

import (
	"testing"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestProductExchangeAdjustsInvoiceBeforeWarehouseAndCanBeReversed(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	for _, model := range []interface{}{
		&domain.SalesTransaction{}, &domain.SalesItem{}, &domain.SalesPayment{}, &domain.SalesReturnItem{},
	} {
		if err := createDepositTestTable(db, model); err != nil {
			t.Fatal(err)
		}
	}

	invoiceID, originalID, returnedID := uuid.New(), uuid.New(), uuid.New()
	originalProductID, replacementProductID := uuid.New(), uuid.New()
	invoice := domain.SalesTransaction{
		ID: invoiceID, CompanyID: uuid.New(), EmployeeID: uuid.New(),
		TotalAmount: 100000, PaidAmount: 100000, PaymentStatus: "PAID", ReceiptNo: "INV-TEST",
	}
	original := domain.SalesItem{
		ID: originalID, SalesTransactionID: invoiceID, ProductID: originalProductID,
		Quantity: 2, PriceAtTransaction: 30000, Subtotal: 60000,
	}
	returnItem := domain.SalesReturnItem{
		ID: returnedID, SalesReturnID: uuid.New(), ProductID: originalProductID,
		SalesItemID: &originalID, Quantity: 1, Price: 30000,
		ReplacementProductID: &replacementProductID, ReplacementQuantity: 1, ReplacementPrice: 45000,
	}
	for _, model := range []interface{}{&invoice, &original, &returnItem} {
		if err := db.Create(model).Error; err != nil {
			t.Fatal(err)
		}
	}
	ret := domain.SalesReturn{
		ID: returnItem.SalesReturnID, TransactionID: &invoiceID,
		ReturnNo: "RET-TEST", PriceDifference: 15000, Items: []domain.SalesReturnItem{returnItem},
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return applyProductExchangeToInvoice(tx, &ret) }); err != nil {
		t.Fatal(err)
	}
	var adjusted domain.SalesTransaction
	if err := db.First(&adjusted, "id = ?", invoiceID).Error; err != nil {
		t.Fatal(err)
	}
	if adjusted.TotalAmount != 115000 || adjusted.PaidAmount != 100000 || adjusted.PaymentStatus != "PARTIAL" || ret.InvoiceAdjustedAt == nil {
		t.Fatalf("invoice after submission = total %.0f, paid %.0f, status %s, adjusted %v", adjusted.TotalAmount, adjusted.PaidAmount, adjusted.PaymentStatus, ret.InvoiceAdjustedAt)
	}
	var remaining domain.SalesItem
	if err := db.First(&remaining, "id = ?", originalID).Error; err != nil || remaining.Quantity != 1 {
		t.Fatalf("original invoice item = %+v, error %v", remaining, err)
	}
	var savedReturnItem domain.SalesReturnItem
	if err := db.First(&savedReturnItem, "id = ?", returnedID).Error; err != nil || savedReturnItem.ReplacementSalesItemID == nil {
		t.Fatalf("replacement invoice item reference = %+v, error %v", savedReturnItem, err)
	}
	ret.Items[0] = savedReturnItem
	if err := db.Transaction(func(tx *gorm.DB) error { return reverseProductExchangeInvoice(tx, &ret) }); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&adjusted, "id = ?", invoiceID).Error; err != nil {
		t.Fatal(err)
	}
	if adjusted.TotalAmount != 100000 || adjusted.PaidAmount != 100000 || adjusted.PaymentStatus != "PAID" {
		t.Fatalf("invoice after cancellation = total %.0f, paid %.0f, status %s", adjusted.TotalAmount, adjusted.PaidAmount, adjusted.PaymentStatus)
	}
	if err := db.First(&remaining, "id = ?", originalID).Error; err != nil || remaining.Quantity != 2 {
		t.Fatalf("original invoice item after cancellation = %+v, error %v", remaining, err)
	}
	var replacementCount int64
	if err := db.Model(&domain.SalesItem{}).Where("id = ?", *savedReturnItem.ReplacementSalesItemID).Count(&replacementCount).Error; err != nil || replacementCount != 0 {
		t.Fatalf("replacement invoice item still exists: count %d, error %v", replacementCount, err)
	}
}
