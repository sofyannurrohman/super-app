package domain

import (
	"database/sql"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMarketplaceOrderCreateUsesValidEmptyFinanceJSON(t *testing.T) {
	conn, err := sql.Open("pgx", "host=localhost user=test dbname=test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	order := &MarketplaceOrder{OrderID: "test-order", Marketplace: "tiktok"}
	result := db.Create(order)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if order.FeeTaxBreakdown != "{}" {
		t.Fatalf("new order finance JSON = %q, want valid empty object", order.FeeTaxBreakdown)
	}
	var boundJSON bool
	for _, value := range result.Statement.Vars {
		if value == "{}" {
			boundJSON = true
			break
		}
	}
	if !boundJSON {
		t.Fatal("order INSERT did not bind valid JSON for fee_tax_breakdown")
	}
}
