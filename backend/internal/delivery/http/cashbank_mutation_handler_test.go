package http

import (
	"testing"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
)

func TestCashierDepositBranch(t *testing.T) {
	home := uuid.New()
	operating := uuid.New()
	other := uuid.New()
	employee := domain.Employee{BranchID: &home, OperatingBranchID: &operating}

	for _, tt := range []struct {
		name      string
		requested string
		want      uuid.UUID
		wantError bool
	}{
		{name: "default uses home branch shown in cashier UI", want: home},
		{name: "home branch is allowed", requested: home.String(), want: home},
		{name: "operating branch is allowed", requested: operating.String(), want: operating},
		{name: "unassigned branch is denied", requested: other.String(), wantError: true},
		{name: "malformed branch is denied", requested: "wrong", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cashierDepositBranch(employee, tt.requested)
			if (err != nil) != tt.wantError || (!tt.wantError && got != tt.want) {
				t.Fatalf("cashierDepositBranch() = %s, %v; want %s, error %t", got, err, tt.want, tt.wantError)
			}
		})
	}
}

func TestFactoryDepositUsesBranchBankAndBranchReviewer(t *testing.T) {
	branchID := uuid.New()
	otherBranchID := uuid.New()
	companyID := uuid.New()
	bank := domain.CompanyBankAccount{BranchID: &branchID, CompanyID: &companyID}
	if !factoryDepositBankMatchesBranch(bank, branchID) {
		t.Fatal("rekening cabang harus dapat dipilih tanpa memfilter PT")
	}
	if factoryDepositBankMatchesBranch(bank, otherBranchID) || factoryDepositBankMatchesBranch(domain.CompanyBankAccount{CompanyID: &companyID}, branchID) {
		t.Fatal("rekening cabang lain atau rekening tanpa cabang tidak boleh menjadi tujuan")
	}
	if !canReviewFactoryDeposits(string(domain.RoleAdminKasir)) || !canReviewFactoryDeposits(string(domain.RoleSuperAdmin)) {
		t.Fatal("Admin Kasir cabang dan Super Admin harus dapat memverifikasi")
	}
	if canReviewFactoryDeposits(string(domain.RoleCashier)) || canReviewFactoryDeposits(string(domain.RoleCashierPusat)) {
		t.Fatal("Kasir penyetor dan Kasir Pusat tidak boleh memverifikasi setoran cabang")
	}
	employee := domain.Employee{BranchID: &branchID, OperatingBranchID: &otherBranchID}
	branches := assignedDepositBranches(employee)
	if len(branches) != 2 || branches[0] != branchID || branches[1] != otherBranchID {
		t.Fatalf("cabang yang dapat ditinjau = %v, ingin [%s %s]", branches, branchID, otherBranchID)
	}
}

func TestFactoryInvoiceDepositDifference(t *testing.T) {
	tests := []struct {
		name      string
		deposited float64
		allocated float64
		origin    string
		notes     string
		wantError bool
	}{
		{name: "full invoice allocation", deposited: 100000, allocated: 100000},
		{name: "partial invoice allocation", deposited: 70000, allocated: 70000},
		{name: "cannot allocate more than transferred", deposited: 70000, allocated: 100000, wantError: true},
		{name: "surplus needs origin", deposited: 120000, allocated: 100000, notes: "Dana koreksi", wantError: true},
		{name: "surplus needs explanation", deposited: 120000, allocated: 100000, origin: "CASH_ADJUSTMENT", wantError: true},
		{name: "surplus with explanation", deposited: 120000, allocated: 100000, origin: "CASH_ADJUSTMENT", notes: "Koreksi kas cabang", wantError: false},
		{name: "sales cannot be a surplus origin", deposited: 120000, allocated: 100000, origin: "SALES", notes: "Hasil penjualan", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFactoryInvoiceDepositDifference(tt.deposited, tt.allocated, tt.origin, tt.notes)
			if (err != nil) != tt.wantError {
				t.Fatalf("validateFactoryInvoiceDepositDifference() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
