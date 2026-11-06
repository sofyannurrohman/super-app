package usecase

import "testing"

func TestExchangeInvoiceBalance(t *testing.T) {
	tests := []struct {
		name       string
		total      float64
		paid       float64
		difference float64
		wantTotal  float64
		wantStatus string
		wantError  bool
	}{
		{name: "unpaid exchange", total: 100000, difference: 15000, wantTotal: 115000, wantStatus: "UNPAID"},
		{name: "paid exchange costs more", total: 100000, paid: 100000, difference: 15000, wantTotal: 115000, wantStatus: "PARTIAL"},
		{name: "paid exchange costs less", total: 100000, paid: 100000, difference: -15000, wantTotal: 85000, wantStatus: "PAID"},
		{name: "partial payment covers new total", total: 100000, paid: 60000, difference: -50000, wantTotal: 50000, wantStatus: "PAID"},
		{name: "invalid negative invoice", total: 10000, paid: 10000, difference: -15000, wantError: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, status, err := exchangeInvoiceBalance(tc.total, tc.paid, tc.difference)
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tc.wantError)
			}
			if err == nil && (total != tc.wantTotal || status != tc.wantStatus) {
				t.Fatalf("total/status = %.2f/%s, want %.2f/%s", total, status, tc.wantTotal, tc.wantStatus)
			}
		})
	}
}
