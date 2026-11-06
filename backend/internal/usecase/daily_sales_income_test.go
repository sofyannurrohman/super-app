package usecase_test

import (
	"os"
	"testing"
	"time"

	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetDailySalesIncomeReport(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN tidak disetel")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping DB test, database not reachable: %v", err)
		return
	}

	repo := repository.NewFinanceRepository(db)
	u := usecase.NewFinanceUsecase(repo, nil, db, nil)

	// Test today query
	todayStr := time.Now().Format("2006-01-02")
	report, err := u.GetDailySalesIncomeReport(todayStr, "", "", "ALL", "ALL_BRANCHES")
	if err != nil {
		t.Fatalf("GetDailySalesIncomeReport failed: %v", err)
	}

	if report == nil {
		t.Fatal("Expected non-nil report")
	}

	t.Logf("Summary: Date=%s, Cash=%.0f, Transfer=%.0f, Total=%.0f, TrxCount=%d",
		report.Summary.Date,
		report.Summary.TotalCash,
		report.Summary.TotalTransfer,
		report.Summary.GrandTotal,
		report.Summary.TotalTransactions,
	)

	t.Logf("Companies count: %d", len(report.Companies))
	for _, comp := range report.Companies {
		t.Logf("- Company: %s (Total: %.0f, Branches: %d)", comp.CompanyName, comp.TotalAmount, len(comp.Branches))
	}

	t.Logf("Payments count: %d", len(report.Payments))
}
