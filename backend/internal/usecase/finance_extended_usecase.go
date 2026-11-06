package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

// ---- AR/AP Aging Types ----

type AgingBucket struct {
	EntityID   string  `json:"entity_id"`
	EntityName string  `json:"entity_name"` // Nama toko / supplier
	Current    float64 `json:"current"`     // 0-30 hari
	Days31to60 float64 `json:"days_31_60"`
	Days61to90 float64 `json:"days_61_90"`
	Over90     float64 `json:"over_90"`
	Total      float64 `json:"total"`
}

type AgingReport struct {
	AsOf       string        `json:"as_of"`
	ReportType string        `json:"report_type"` // AR atau AP
	Buckets    []AgingBucket `json:"buckets"`
	Summary    AgingBucket   `json:"summary"` // Grand total row
}

// ---- AR/AP Aging Usecase ----

type ARAPUsecase interface {
	GetARAgingReport(companyID uuid.UUID, asOf time.Time) (*AgingReport, error)
	GetAPAgingReport(companyID uuid.UUID, asOf time.Time) (*AgingReport, error)
}

type arAPUsecase struct {
	db *gorm.DB
}

func NewARAPUsecase(db *gorm.DB) ARAPUsecase {
	return &arAPUsecase{db: db}
}

// GetARAgingReport — Aging piutang dari sales_transactions yang belum lunas
func (u *arAPUsecase) GetARAgingReport(companyID uuid.UUID, asOf time.Time) (*AgingReport, error) {
	type rawRow struct {
		StoreID         string
		StoreName       string
		TransactionDate time.Time
		Outstanding     float64
	}
	var rows []rawRow

	q := u.db.Raw(`
		SELECT
			st.store_id::text AS store_id,
			s.name AS store_name,
			st.transaction_date,
			COALESCE(st.total_amount, 0) - COALESCE(st.paid_amount, 0) AS outstanding
		FROM sales_transactions st
		JOIN stores s ON st.store_id = s.id
		WHERE st.status = 'VERIFIED'
		  AND st.payment_status IN ('UNPAID', 'PARTIAL')
		  AND st.transaction_date <= ?
		  AND (COALESCE(st.total_amount, 0) - COALESCE(st.paid_amount, 0)) > 0
		  `+arCompanyFilter(companyID)+`
		ORDER BY s.name, st.transaction_date
	`, asOf)

	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("AR aging query failed: %w", err)
	}

	return buildAgingReport(rows, asOf, "AR", func(r rawRow) (string, string, time.Time, float64) {
		return r.StoreID, r.StoreName, r.TransactionDate, r.Outstanding
	}), nil
}

// GetAPAgingReport — Aging hutang ke supplier dari supplier_receipts + supplier_receipt_items
func (u *arAPUsecase) GetAPAgingReport(companyID uuid.UUID, asOf time.Time) (*AgingReport, error) {
	type rawRow struct {
		SupplierID   string
		SupplierName string
		ReceiptDate  time.Time
		Outstanding  float64
	}
	var rows []rawRow

	// Hutang dihitung dari faktur supplier (Gudang Cabang & Pabrik Bahan Baku) dikurangi pembayaran
	q := u.db.Raw(`
		SELECT
			sr.supplier_id::text AS supplier_id,
			COALESCE(sp.name, 'Supplier Tidak Dikenal') AS supplier_name,
			COALESCE(sr.payment_due_date, sr.invoice_date, sr.received_date, sr.created_at)::date AS receipt_date,
			GREATEST(COALESCE(NULLIF(sr.invoice_total, 0), (SELECT SUM(sri.unit_price * sri.received_qty) FROM supplier_receipt_items sri WHERE sri.supplier_receipt_id = sr.id AND sri.received_qty > 0), 0) - COALESCE(sr.paid_amount, 0), 0) AS outstanding
		FROM supplier_receipts sr
		LEFT JOIN suppliers sp ON sr.supplier_id = sp.id
		WHERE sr.status IN ('PENDING', 'RECEIVED', 'PARTIAL')
		  AND GREATEST(COALESCE(NULLIF(sr.invoice_total, 0), (SELECT SUM(sri.unit_price * sri.received_qty) FROM supplier_receipt_items sri WHERE sri.supplier_receipt_id = sr.id AND sri.received_qty > 0), 0) - COALESCE(sr.paid_amount, 0), 0) > 0
		  AND COALESCE(sr.received_date, sr.created_at)::date <= ?

		UNION ALL

		SELECT
			COALESCE(rmr.supplier_id::text, rmr.id::text) AS supplier_id,
			COALESCE(sp.name, NULLIF(rmr.supplier_name, ''), 'Supplier Bahan Baku') AS supplier_name,
			COALESCE(rmr.payment_due_date, rmr.invoice_date, rmr.received_date, rmr.created_at)::date AS receipt_date,
			GREATEST(COALESCE(NULLIF(rmr.invoice_total, 0), (SELECT SUM(ri.unit_price * ri.received_qty) FROM raw_material_receipt_items ri WHERE ri.receipt_id = rmr.id AND ri.received_qty > 0), 0) - COALESCE(rmr.paid_amount, 0), 0) AS outstanding
		FROM raw_material_receipts rmr
		LEFT JOIN suppliers sp ON rmr.supplier_id = sp.id
		WHERE rmr.payment_status != 'PAID'
		  AND GREATEST(COALESCE(NULLIF(rmr.invoice_total, 0), (SELECT SUM(ri.unit_price * ri.received_qty) FROM raw_material_receipt_items ri WHERE ri.receipt_id = rmr.id AND ri.received_qty > 0), 0) - COALESCE(rmr.paid_amount, 0), 0) > 0
		  AND COALESCE(rmr.received_date, rmr.created_at)::date <= ?
	`, asOf, asOf)

	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("AP aging query failed: %w", err)
	}

	return buildAgingReport(rows, asOf, "AP", func(r rawRow) (string, string, time.Time, float64) {
		return r.SupplierID, r.SupplierName, r.ReceiptDate, r.Outstanding
	}), nil
}

// buildAgingReport — Generic helper untuk build aging report dari rows apapun
func buildAgingReport[T any](
	rows []T,
	asOf time.Time,
	reportType string,
	extract func(T) (id string, name string, date time.Time, amount float64),
) *AgingReport {
	bucketMap := make(map[string]*AgingBucket)
	var totalSummary AgingBucket

	for _, row := range rows {
		id, name, date, amount := extract(row)
		if amount <= 0 {
			continue
		}

		if _, ok := bucketMap[id]; !ok {
			bucketMap[id] = &AgingBucket{EntityID: id, EntityName: name}
		}

		days := int(asOf.Sub(date).Hours() / 24)
		b := bucketMap[id]

		switch {
		case days <= 30:
			b.Current += amount
		case days <= 60:
			b.Days31to60 += amount
		case days <= 90:
			b.Days61to90 += amount
		default:
			b.Over90 += amount
		}
		b.Total += amount
	}

	buckets := make([]AgingBucket, 0, len(bucketMap))
	for _, b := range bucketMap {
		buckets = append(buckets, *b)
		totalSummary.Current += b.Current
		totalSummary.Days31to60 += b.Days31to60
		totalSummary.Days61to90 += b.Days61to90
		totalSummary.Over90 += b.Over90
		totalSummary.Total += b.Total
	}
	totalSummary.EntityName = "TOTAL"

	return &AgingReport{
		AsOf:       asOf.Format("2006-01-02"),
		ReportType: reportType,
		Buckets:    buckets,
		Summary:    totalSummary,
	}
}

func arCompanyFilter(companyID uuid.UUID) string {
	if companyID == uuid.Nil {
		return ""
	}
	return fmt.Sprintf("AND s.company_id = '%s'", companyID.String())
}

// ---- Fixed Asset Usecase ----

type FixedAssetUsecase interface {
	ListFixedAssets(companyID uuid.UUID) ([]domain.FixedAsset, *domain.FixedAssetSummary, error)
	GetFixedAsset(id uuid.UUID) (*domain.FixedAsset, error)
	CreateFixedAsset(asset *domain.FixedAsset) error
	UpdateFixedAsset(asset *domain.FixedAsset) error
	DeleteFixedAsset(id uuid.UUID) error
	GetDepreciationSchedule(id uuid.UUID) ([]domain.DepreciationScheduleLine, error)
	RunMonthlyDepreciation(id uuid.UUID, period time.Time) error
}

type fixedAssetUsecase struct {
	db *gorm.DB
}

func NewFixedAssetUsecase(db *gorm.DB) FixedAssetUsecase {
	return &fixedAssetUsecase{db: db}
}

func (u *fixedAssetUsecase) ListFixedAssets(companyID uuid.UUID) ([]domain.FixedAsset, *domain.FixedAssetSummary, error) {
	var assets []domain.FixedAsset
	q := u.db.Where("deleted_at IS NULL")
	if companyID != uuid.Nil {
		q = q.Where("company_id = ?", companyID)
	}
	if err := q.Find(&assets).Error; err != nil {
		return nil, nil, err
	}

	summary := &domain.FixedAssetSummary{}
	for i := range assets {
		assets[i].ComputeDerived()
		summary.TotalAssets++
		summary.TotalAcquisitionCost += assets[i].AcquisitionCost
		summary.TotalAccumulated += assets[i].AccumulatedDepreciation
		summary.TotalBookValue += assets[i].BookValue
	}
	return assets, summary, nil
}

func (u *fixedAssetUsecase) GetFixedAsset(id uuid.UUID) (*domain.FixedAsset, error) {
	var asset domain.FixedAsset
	if err := u.db.Where("id = ?", id).First(&asset).Error; err != nil {
		return nil, err
	}
	asset.ComputeDerived()
	return &asset, nil
}

func (u *fixedAssetUsecase) CreateFixedAsset(asset *domain.FixedAsset) error {
	if asset.AssetCode == "" {
		return errors.New("asset_code wajib diisi")
	}
	if asset.AcquisitionCost <= 0 {
		return errors.New("acquisition_cost harus lebih dari 0")
	}
	if asset.UsefulLifeYears <= 0 {
		return errors.New("useful_life_years harus lebih dari 0")
	}
	return u.db.Create(asset).Error
}

func (u *fixedAssetUsecase) UpdateFixedAsset(asset *domain.FixedAsset) error {
	return u.db.Save(asset).Error
}

func (u *fixedAssetUsecase) DeleteFixedAsset(id uuid.UUID) error {
	return u.db.Where("id = ?", id).Delete(&domain.FixedAsset{}).Error
}

// GetDepreciationSchedule — Menghasilkan jadwal depresiasi per bulan untuk seluruh umur aset
func (u *fixedAssetUsecase) GetDepreciationSchedule(id uuid.UUID) ([]domain.DepreciationScheduleLine, error) {
	asset, err := u.GetFixedAsset(id)
	if err != nil {
		return nil, err
	}

	totalMonths := asset.UsefulLifeYears * 12
	depreciableAmount := asset.AcquisitionCost - asset.ResidualValue
	if depreciableAmount < 0 {
		depreciableAmount = 0
	}

	schedule := make([]domain.DepreciationScheduleLine, 0, totalMonths)
	currentBook := asset.AcquisitionCost
	accumulated := 0.0
	period := asset.AcquisitionDate

	for i := 0; i < totalMonths; i++ {
		var monthly float64
		switch asset.DepreciationMethod {
		case "DECLINING_BALANCE":
			rate := 2.0 / float64(asset.UsefulLifeYears)
			monthly = (currentBook - asset.ResidualValue) * rate / 12
		default: // STRAIGHT_LINE
			monthly = depreciableAmount / float64(totalMonths)
		}

		if currentBook-monthly < asset.ResidualValue {
			monthly = currentBook - asset.ResidualValue
		}
		if monthly < 0 {
			monthly = 0
		}

		accumulated += monthly
		closing := currentBook - monthly

		schedule = append(schedule, domain.DepreciationScheduleLine{
			Period:             period.Format("Jan 2006"),
			OpeningBookValue:   currentBook,
			DepreciationAmount: monthly,
			ClosingBookValue:   closing,
			AccumulatedToDate:  accumulated,
		})

		currentBook = closing
		period = period.AddDate(0, 1, 0)

		if closing <= asset.ResidualValue {
			break
		}
	}

	return schedule, nil
}

// RunMonthlyDepreciation — Jalankan depresiasi satu bulan untuk aset tertentu
func (u *fixedAssetUsecase) RunMonthlyDepreciation(id uuid.UUID, period time.Time) error {
	asset, err := u.GetFixedAsset(id)
	if err != nil {
		return err
	}

	if asset.Status != "ACTIVE" {
		return errors.New("hanya aset berstatus ACTIVE yang bisa didepresiasi")
	}
	if asset.BookValue <= asset.ResidualValue {
		return errors.New("nilai buku aset sudah mencapai nilai sisa — tidak perlu depresiasi lagi")
	}

	// Hitung depresiasi bulan ini
	var monthly float64
	depreciableAmount := asset.AcquisitionCost - asset.ResidualValue
	switch asset.DepreciationMethod {
	case "DECLINING_BALANCE":
		rate := 2.0 / float64(asset.UsefulLifeYears)
		monthly = (asset.BookValue - asset.ResidualValue) * rate / 12
	default: // STRAIGHT_LINE
		monthly = depreciableAmount / float64(asset.UsefulLifeYears*12)
	}

	if asset.BookValue-monthly < asset.ResidualValue {
		monthly = asset.BookValue - asset.ResidualValue
	}

	now := period
	return u.db.Model(asset).Updates(map[string]interface{}{
		"accumulated_depreciation": asset.AccumulatedDepreciation + monthly,
		"last_depreciated_month":   &now,
	}).Error
}
