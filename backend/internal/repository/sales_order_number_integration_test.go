package repository

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Run with SO_NUMBER_TEST_DSN pointed at a disposable PostgreSQL database.
func TestCreateWithNextSONumberPostgres(t *testing.T) {
	dsn := os.Getenv("SO_NUMBER_TEST_DSN")
	if dsn == "" {
		t.Skip("SO_NUMBER_TEST_DSN is not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domain.SalesOrder{}); err != nil {
		t.Fatal(err)
	}
	// This legacy column is deliberately excluded from GORM migrations.
	if err := db.Exec("ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS delivery_order_no varchar(100)").Error; err != nil {
		t.Fatal(err)
	}

	repo := NewSalesOrderRepository(db)
	date := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	prefix := "TO" + uuid.NewString()[:8]
	number := func(n int) string { return fmt.Sprintf("SO-%s-20261026-%04d", prefix, n) }
	newOrder := func() *domain.SalesOrder {
		return &domain.SalesOrder{
			BranchID: uuid.New(), CompanyID: uuid.New(), EmployeeID: uuid.New(),
			StoreID: uuid.New(), OrderDate: date,
		}
	}

	first := newOrder()
	if err := repo.CreateWithNextSONumber(first, prefix); err != nil {
		t.Fatal(err)
	}
	if first.SONumber != number(1) {
		t.Fatalf("first number = %s", first.SONumber)
	}
	if err := db.Delete(first).Error; err != nil {
		t.Fatal(err)
	}
	second := newOrder()
	if err := repo.CreateWithNextSONumber(second, prefix); err != nil {
		t.Fatal(err)
	}
	if second.SONumber != number(2) {
		t.Fatalf("number after soft delete = %s", second.SONumber)
	}

	const workers = 12
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			so := newOrder()
			if err := repo.CreateWithNextSONumber(so, prefix); err != nil {
				errs <- err
				return
			}
			errs <- nil
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for n := 3; n <= workers+2; n++ {
		var count int64
		if err := db.Unscoped().Model(&domain.SalesOrder{}).
			Where("so_number = ?", number(n)).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("number %d: count=%d err=%v", n, count, err)
		}
	}
}
