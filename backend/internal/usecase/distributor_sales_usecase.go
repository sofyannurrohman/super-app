package usecase

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type DistributorSalesUsecase interface {
	GetDistributorStores(distributorStoreID uuid.UUID, search string, page, limit int) ([]domain.Store, int64, error)
	CreateDistributorStore(distributorStoreID uuid.UUID, createdByUserID *uuid.UUID, req CreateDistributorStoreRequest) (*domain.Store, error)
	RecordDistributorVisit(distributorStoreID, storeID uuid.UUID, userID *uuid.UUID, employeeID *uuid.UUID, lat, lng float64, selfieURL, notes string) (*domain.SalesVisit, error)
	GetDistributorVisits(distributorStoreID uuid.UUID, startDate, endDate *time.Time, page, limit int) ([]domain.SalesVisit, int64, error)
	CreateDistributorSale(distributorStoreID uuid.UUID, userID *uuid.UUID, employeeID *uuid.UUID, req CreateDistributorSaleRequest) (*domain.SalesTransaction, error)
	GetDistributorSales(distributorStoreID uuid.UUID, startDate, endDate *time.Time, page, limit int) ([]domain.SalesTransaction, int64, error)
	ImportSellOutExcel(distributorStoreID uuid.UUID, userID *uuid.UUID, r io.Reader) (*DistributorImportResult, error)
	GenerateExcelTemplate() ([]byte, error)
	GetDistributorSummary(distributorStoreID uuid.UUID, month, year int) (*DistributorSummaryResponse, error)
}

type distributorSalesUsecase struct {
	db *gorm.DB
}

func NewDistributorSalesUsecase(db *gorm.DB) DistributorSalesUsecase {
	return &distributorSalesUsecase{db: db}
}

// Request and Response Structs
type CreateDistributorStoreRequest struct {
	Name         string  `json:"name" binding:"required"`
	OwnerName    string  `json:"owner_name"`
	PhoneNumber  string  `json:"phone_number"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	CustomerType string  `json:"customer_type"` // Defaults to TOKO
	Province     string  `json:"province"`
	Regency      string  `json:"regency"`
	District     string  `json:"district"`
	Village      string  `json:"village"`
	PhotoURL     string  `json:"photo_url"`
}

type DistributorSaleItemRequest struct {
	ProductID       uuid.UUID `json:"product_id" binding:"required"`
	Quantity        int       `json:"quantity" binding:"required,gt=0"`
	Price           float64   `json:"price"` // Optional, if 0 will auto-use RBP or SellingPrice
	DiscountPercent float64   `json:"discount_percent"`
}

type CreateDistributorSaleRequest struct {
	StoreID         uuid.UUID                    `json:"store_id" binding:"required"`
	TransactionDate string                       `json:"transaction_date"` // YYYY-MM-DD
	ReceiptNo       string                       `json:"receipt_no"`        // Optional
	Notes           string                       `json:"notes"`
	VisitID         *uuid.UUID                   `json:"visit_id"`
	Latitude        *float64                     `json:"latitude"`
	Longitude       *float64                     `json:"longitude"`
	PaymentMethod   string                       `json:"payment_method"`
	Items           []DistributorSaleItemRequest `json:"items" binding:"required,min=1"`
}

type DistributorSummaryResponse struct {
	Month               int                    `json:"month"`
	Year                int                    `json:"year"`
	TotalSellOutAmount  float64                `json:"total_sell_out_amount"`
	TotalTransactions   int64                  `json:"total_transactions"`
	TotalAssignedStores int64                  `json:"total_assigned_stores"`
	TotalVisits         int64                  `json:"total_visits"`
	TopProducts         []domain.ProductRecap `json:"top_products"`
}

type DistributorImportResult struct {
	TotalRows    int      `json:"total_rows"`
	SuccessCount int      `json:"success_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors"`
}

// 1. Get Stores assigned to this distributor
func (u *distributorSalesUsecase) GetDistributorStores(distributorStoreID uuid.UUID, search string, page, limit int) ([]domain.Store, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var stores []domain.Store
	var total int64

	query := u.db.Model(&domain.Store{}).Where("distributor_id = ? AND is_active = ?", distributorStoreID, true)
	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		query = query.Where("(LOWER(name) LIKE ? OR LOWER(address) LIKE ? OR LOWER(owner_name) LIKE ? OR phone_number LIKE ?)", s, s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("name ASC").Offset(offset).Limit(limit).Find(&stores).Error; err != nil {
		return nil, 0, err
	}

	return stores, total, nil
}

// 2. Create a new store under this distributor
func (u *distributorSalesUsecase) CreateDistributorStore(distributorStoreID uuid.UUID, createdByUserID *uuid.UUID, req CreateDistributorStoreRequest) (*domain.Store, error) {
	// Verify distributor store exists
	var distStore domain.Store
	if err := u.db.First(&distStore, "id = ?", distributorStoreID).Error; err != nil {
		return nil, errors.New("distributor tidak ditemukan")
	}

	customerType := req.CustomerType
	if customerType == "" {
		customerType = "TOKO"
	}

	store := &domain.Store{
		ID:               uuid.New(),
		CompanyID:        distStore.CompanyID,
		DistributorID:    &distributorStoreID,
		Name:             strings.TrimSpace(req.Name),
		OwnerName:        strings.TrimSpace(req.OwnerName),
		PhoneNumber:      strings.TrimSpace(req.PhoneNumber),
		Address:          strings.TrimSpace(req.Address),
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		CustomerType:     customerType,
		Province:         strings.TrimSpace(req.Province),
		Regency:          strings.TrimSpace(req.Regency),
		District:         strings.TrimSpace(req.District),
		Village:          strings.TrimSpace(req.Village),
		PhotoURL:         req.PhotoURL,
		IsActive:         true,
		Category:         "TOKO_BARU",
		DistributionCity: distStore.DistributionCity,
	}

	if err := u.db.Create(store).Error; err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan toko retail: %w", err)
	}

	return store, nil
}

// 3. Record Visit by Distributor Salesman
func (u *distributorSalesUsecase) RecordDistributorVisit(distributorStoreID, storeID uuid.UUID, userID *uuid.UUID, employeeID *uuid.UUID, lat, lng float64, selfieURL, notes string) (*domain.SalesVisit, error) {
	var store domain.Store
	if err := u.db.First(&store, "id = ?", storeID).Error; err != nil {
		return nil, errors.New("toko tidak ditemukan")
	}

	now := time.Now()
	var empID uuid.UUID
	if employeeID != nil {
		empID = *employeeID
	}

	visit := &domain.SalesVisit{
		ID:                 uuid.New(),
		StoreID:            storeID,
		EmployeeID:         empID,
		UserID:             userID,
		DistributorStoreID: &distributorStoreID,
		IsDistributorVisit: true,
		CheckInTime:        now,
		Latitude:           lat,
		Longitude:          lng,
		SelfieURL:          selfieURL,
		Notes:              notes,
		Type:               "CHECKIN",
		Status:             "COMPLETED",
	}

	if err := u.db.Create(visit).Error; err != nil {
		return nil, fmt.Errorf("gagal mencatat kunjungan distributor: %w", err)
	}

	return visit, nil
}

// 4. Get visits recorded by distributor
func (u *distributorSalesUsecase) GetDistributorVisits(distributorStoreID uuid.UUID, startDate, endDate *time.Time, page, limit int) ([]domain.SalesVisit, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var visits []domain.SalesVisit
	var total int64

	query := u.db.Model(&domain.SalesVisit{}).
		Where("distributor_store_id = ? AND is_distributor_visit = ?", distributorStoreID, true)

	if startDate != nil {
		query = query.Where("check_in_time >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("check_in_time <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Store").Preload("User").Preload("Employee").
		Order("check_in_time DESC").Offset(offset).Limit(limit).Find(&visits).Error; err != nil {
		return nil, 0, err
	}

	return visits, total, nil
}

// 5. Create Sell-out Transaction (No physical warehouse stock deduction - reporting only)
func (u *distributorSalesUsecase) CreateDistributorSale(distributorStoreID uuid.UUID, userID *uuid.UUID, employeeID *uuid.UUID, req CreateDistributorSaleRequest) (*domain.SalesTransaction, error) {
	var distStore domain.Store
	if err := u.db.First(&distStore, "id = ?", distributorStoreID).Error; err != nil {
		return nil, errors.New("distributor tidak ditemukan")
	}

	var retailStore domain.Store
	if err := u.db.First(&retailStore, "id = ?", req.StoreID).Error; err != nil {
		return nil, errors.New("toko retail tidak ditemukan")
	}

	txDate := time.Now()
	if req.TransactionDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.TransactionDate); err == nil {
			txDate = parsed
		}
	}

	receiptNo := req.ReceiptNo
	if receiptNo == "" {
		receiptNo = fmt.Sprintf("SO-DIST-%s-%d", distStore.ID.String()[:6], time.Now().Unix())
	}

	var empID uuid.UUID
	if employeeID != nil {
		empID = *employeeID
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "CASH"
	}

	var totalAmount float64
	var discountAmount float64
	var salesItems []domain.SalesItem

	for _, itemReq := range req.Items {
		var product domain.Product
		if err := u.db.First(&product, "id = ?", itemReq.ProductID).Error; err != nil {
			return nil, fmt.Errorf("produk ID %s tidak ditemukan", itemReq.ProductID)
		}

		// Use RBP (Retailer Buying Price - the price distributor sells to retailer)
		// Fallback to SellingPrice if RBP is 0
		price := itemReq.Price
		if price <= 0 {
			if product.RBP > 0 {
				price = product.RBP
			} else if product.SellingPrice > 0 {
				price = product.SellingPrice
			} else {
				price = product.CostPrice
			}
		}

		itemTotal := float64(itemReq.Quantity) * price
		itemDiscount := itemTotal * (itemReq.DiscountPercent / 100.0)
		subtotal := itemTotal - itemDiscount

		totalAmount += subtotal
		discountAmount += itemDiscount

		salesItems = append(salesItems, domain.SalesItem{
			ID:                 uuid.New(),
			ProductID:          product.ID,
			Quantity:           itemReq.Quantity,
			PriceAtTransaction: price,
			DiscountPercent:    itemReq.DiscountPercent,
			Subtotal:           subtotal,
		})
	}

	notes := req.Notes
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	tx := &domain.SalesTransaction{
		ID:                 uuid.New(),
		CompanyID:          distStore.CompanyID,
		StoreID:            &retailStore.ID,
		EmployeeID:         empID,
		UserID:             userID,
		DistributorStoreID: &distributorStoreID,
		IsDistributorOrder: true,
		SalesChannel:       domain.SalesChannelDistributorSellOut,
		BuyerType:          domain.BuyerTypeRegisteredStore,
		ReceiptNo:          receiptNo,
		TotalAmount:        totalAmount,
		DiscountAmount:     discountAmount,
		PaidAmount:         totalAmount, // Reporting sell-out defaults to full settled or cash
		PaymentStatus:      string(domain.PaymentStatusPaid),
		PaymentMethod:      paymentMethod,
		StoreCategory:      retailStore.Category,
		PeriodMonth:        int(txDate.Month()),
		PeriodYear:         txDate.Year(),
		TransactionDate:    txDate,
		Status:             "VERIFIED",
		Notes:              notesPtr,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
		VisitID:            req.VisitID,
	}

	err := u.db.Transaction(func(dbTx *gorm.DB) error {
		if err := dbTx.Create(tx).Error; err != nil {
			return err
		}

		for i := range salesItems {
			salesItems[i].SalesTransactionID = tx.ID
			if err := dbTx.Create(&salesItems[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("gagal mencatat transaksi penjualan distributor: %w", err)
	}

	tx.Items = salesItems
	return tx, nil
}

// 6. Get distributor sales transactions
func (u *distributorSalesUsecase) GetDistributorSales(distributorStoreID uuid.UUID, startDate, endDate *time.Time, page, limit int) ([]domain.SalesTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var sales []domain.SalesTransaction
	var total int64

	query := u.db.Model(&domain.SalesTransaction{}).
		Where("distributor_store_id = ? AND is_distributor_order = ?", distributorStoreID, true)

	if startDate != nil {
		query = query.Where("transaction_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("transaction_date <= ?", *endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Store").Preload("User").Preload("Employee").Preload("Items.Product").
		Order("transaction_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&sales).Error; err != nil {
		return nil, 0, err
	}

	return sales, total, nil
}

// 7. Bulk Import Sell-out from Excel
func (u *distributorSalesUsecase) ImportSellOutExcel(distributorStoreID uuid.UUID, userID *uuid.UUID, r io.Reader) (*DistributorImportResult, error) {
	excelFile, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("format file excel tidak valid: %w", err)
	}
	defer excelFile.Close()

	sheets := excelFile.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("file excel kosong atau tidak memiliki lembar kerja")
	}

	rows, err := excelFile.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("gagal membaca baris excel: %w", err)
	}

	if len(rows) <= 1 {
		return nil, errors.New("tidak ada data transaksi pada file excel (hanya header atau kosong)")
	}

	// Verify distributor exists
	var distStore domain.Store
	if err := u.db.First(&distStore, "id = ?", distributorStoreID).Error; err != nil {
		return nil, errors.New("distributor tidak ditemukan")
	}

	// Pre-cache products by SKU and Name for fast lookup
	var products []domain.Product
	_ = u.db.Find(&products).Error
	skuMap := make(map[string]domain.Product)
	nameMap := make(map[string]domain.Product)
	for _, p := range products {
		if p.SKU != "" {
			skuMap[strings.ToUpper(strings.TrimSpace(p.SKU))] = p
		}
		nameMap[strings.ToUpper(strings.TrimSpace(p.Name))] = p
	}

	// Pre-cache stores under this distributor
	var stores []domain.Store
	_ = u.db.Where("distributor_id = ?", distributorStoreID).Find(&stores).Error
	storeMap := make(map[string]domain.Store)
	for _, s := range stores {
		storeMap[strings.ToUpper(strings.TrimSpace(s.Name))] = s
	}

	type ParsedItem struct {
		Date      time.Time
		StoreName string
		Product   domain.Product
		Quantity  int
		Price     float64
		Notes     string
	}

	var parsedItems []ParsedItem
	var errMsgs []string

	// Expected Header Row 1, Data starts Row 2 (index 1)
	// Col A (0): Tanggal (YYYY-MM-DD atau DD/MM/YYYY)
	// Col B (1): Nama Toko
	// Col C (2): SKU atau Nama Produk
	// Col D (3): Jumlah / Qty (Pcs)
	// Col E (4): Harga Satuan (Opsional, angka)
	// Col F (5): Catatan
	for idx, row := range rows[1:] {
		rowNum := idx + 2
		if len(row) < 4 {
			continue // Skip incomplete empty rows
		}

		dateStr := strings.TrimSpace(row[0])
		storeName := strings.TrimSpace(row[1])
		productRef := strings.TrimSpace(row[2])
		qtyStr := strings.TrimSpace(row[3])

		if dateStr == "" && storeName == "" && productRef == "" {
			continue // Blank row
		}

		if dateStr == "" || storeName == "" || productRef == "" || qtyStr == "" {
			errMsgs = append(errMsgs, fmt.Sprintf("Baris %d: Tanggal, Nama Toko, Produk, dan Qty wajib diisi", rowNum))
			continue
		}

		// Parse date
		var txDate time.Time
		var parseDateErr error
		if strings.Contains(dateStr, "-") {
			txDate, parseDateErr = time.Parse("2006-01-02", dateStr)
		} else if strings.Contains(dateStr, "/") {
			txDate, parseDateErr = time.Parse("02/01/2006", dateStr)
		} else {
			txDate, parseDateErr = time.Parse("2006-01-02", dateStr)
		}

		if parseDateErr != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("Baris %d: Format tanggal '%s' tidak valid (gunakan YYYY-MM-DD)", rowNum, dateStr))
			continue
		}

		qty, err := strconv.Atoi(qtyStr)
		if err != nil || qty <= 0 {
			errMsgs = append(errMsgs, fmt.Sprintf("Baris %d: Jumlah Qty '%s' harus berupa angka bulat positif", rowNum, qtyStr))
			continue
		}

		// Find product
		prod, found := skuMap[strings.ToUpper(productRef)]
		if !found {
			prod, found = nameMap[strings.ToUpper(productRef)]
		}
		if !found {
			errMsgs = append(errMsgs, fmt.Sprintf("Baris %d: Produk '%s' tidak ditemukan di katalog", rowNum, productRef))
			continue
		}

		// Determine price
		var price float64
		if len(row) > 4 && strings.TrimSpace(row[4]) != "" {
			cleanPrice := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(row[4]), ",", ""), "Rp", "")
			if pVal, err := strconv.ParseFloat(cleanPrice, 64); err == nil && pVal > 0 {
				price = pVal
			}
		}
		if price <= 0 {
			if prod.RBP > 0 {
				price = prod.RBP
			} else if prod.SellingPrice > 0 {
				price = prod.SellingPrice
			} else {
				price = prod.CostPrice
			}
		}

		notes := ""
		if len(row) > 5 {
			notes = strings.TrimSpace(row[5])
		}

		parsedItems = append(parsedItems, ParsedItem{
			Date:      txDate,
			StoreName: storeName,
			Product:   prod,
			Quantity:  qty,
			Price:     price,
			Notes:     notes,
		})
	}

	if len(parsedItems) == 0 {
		return &DistributorImportResult{
			TotalRows:    len(rows) - 1,
			SuccessCount: 0,
			ErrorCount:   len(errMsgs),
			Errors:       errMsgs,
		}, nil
	}

	// Group items by (StoreName, Date)
	type GroupKey struct {
		StoreName string
		DateStr   string
	}
	grouped := make(map[GroupKey][]ParsedItem)
	for _, item := range parsedItems {
		k := GroupKey{
			StoreName: strings.ToUpper(item.StoreName),
			DateStr:   item.Date.Format("2006-01-02"),
		}
		grouped[k] = append(grouped[k], item)
	}

	successCount := 0

	err = u.db.Transaction(func(dbTx *gorm.DB) error {
		for key, items := range grouped {
			// Find or auto-create store under this distributor
			st, exists := storeMap[key.StoreName]
			if !exists {
				// Auto-create basic store record so import is frictionless
				st = domain.Store{
					ID:               uuid.New(),
					CompanyID:        distStore.CompanyID,
					DistributorID:    &distributorStoreID,
					Name:             items[0].StoreName,
					CustomerType:     "TOKO",
					Category:         "TOKO_BARU",
					IsActive:         true,
					DistributionCity: distStore.DistributionCity,
				}
				if err := dbTx.Create(&st).Error; err != nil {
					return err
				}
				storeMap[key.StoreName] = st
			}

			// Build sales transaction
			txDate := items[0].Date
			var totalAmount float64
			var salesItems []domain.SalesItem

			for _, it := range items {
				sub := float64(it.Quantity) * it.Price
				totalAmount += sub
				salesItems = append(salesItems, domain.SalesItem{
					ID:                 uuid.New(),
					ProductID:          it.Product.ID,
					Quantity:           it.Quantity,
					PriceAtTransaction: it.Price,
					Subtotal:           sub,
				})
			}

			receiptNo := fmt.Sprintf("SO-IMP-%s-%d-%d", distStore.ID.String()[:4], txDate.Unix(), time.Now().Nanosecond()%1000)
			notesStr := fmt.Sprintf("Import Excel oleh Distributor: %s", distStore.Name)

			tx := &domain.SalesTransaction{
				ID:                 uuid.New(),
				CompanyID:          distStore.CompanyID,
				StoreID:            &st.ID,
				UserID:             userID,
				DistributorStoreID: &distributorStoreID,
				IsDistributorOrder: true,
				SalesChannel:       domain.SalesChannelDistributorSellOut,
				BuyerType:          domain.BuyerTypeRegisteredStore,
				ReceiptNo:          receiptNo,
				TotalAmount:        totalAmount,
				PaidAmount:         totalAmount,
				PaymentStatus:      string(domain.PaymentStatusPaid),
				PaymentMethod:      "CASH",
				StoreCategory:      st.Category,
				PeriodMonth:        int(txDate.Month()),
				PeriodYear:         txDate.Year(),
				TransactionDate:    txDate,
				Status:             "VERIFIED",
				Notes:              &notesStr,
			}

			if err := dbTx.Create(tx).Error; err != nil {
				return err
			}

			for i := range salesItems {
				salesItems[i].SalesTransactionID = tx.ID
				if err := dbTx.Create(&salesItems[i]).Error; err != nil {
					return err
				}
			}

			successCount += len(items)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaksi database gagal: %w", err)
	}

	return &DistributorImportResult{
		TotalRows:    len(rows) - 1,
		SuccessCount: successCount,
		ErrorCount:   len(errMsgs),
		Errors:       errMsgs,
	}, nil
}

// 8. Generate Excel Template for Distributor
func (u *distributorSalesUsecase) GenerateExcelTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Template Penjualan"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Tanggal (YYYY-MM-DD)",
		"Nama Toko / Warung",
		"SKU atau Nama Produk",
		"Jumlah (Qty)",
		"Harga Satuan (Kosongkan utk harga standar)",
		"Catatan",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1F3864"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	_ = f.SetRowStyle(sheet, 1, 1, headerStyle)
	_ = f.SetRowHeight(sheet, 1, 28)

	// Sample Rows
	today := time.Now().Format("2006-01-02")
	samples := [][]interface{}{
		{today, "Toko Berkah Abadi", "SKU-SAMPLE-01", 10, "", "Penjualan tunai"},
		{today, "Toko Sumber Rejeki", "SKU-SAMPLE-02", 5, 25000, "Harga promo toko"},
	}

	for rowIdx, sample := range samples {
		for colIdx, val := range sample {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			_ = f.SetCellValue(sheet, cell, val)
		}
	}

	_ = f.SetColWidth(sheet, "A", "A", 22)
	_ = f.SetColWidth(sheet, "B", "B", 30)
	_ = f.SetColWidth(sheet, "C", "C", 25)
	_ = f.SetColWidth(sheet, "D", "D", 15)
	_ = f.SetColWidth(sheet, "E", "E", 35)
	_ = f.SetColWidth(sheet, "F", "F", 25)

	buf := new(bytes.Buffer)
	if err := f.Write(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 9. Dashboard Summary for Distributor
func (u *distributorSalesUsecase) GetDistributorSummary(distributorStoreID uuid.UUID, month, year int) (*DistributorSummaryResponse, error) {
	now := time.Now()
	if month < 1 || month > 12 {
		month = int(now.Month())
	}
	if year < 2000 {
		year = now.Year()
	}

	var totalAmount float64
	var totalTx int64
	_ = u.db.Model(&domain.SalesTransaction{}).
		Where("distributor_store_id = ? AND is_distributor_order = ? AND period_month = ? AND period_year = ?",
			distributorStoreID, true, month, year).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalAmount).Error

	_ = u.db.Model(&domain.SalesTransaction{}).
		Where("distributor_store_id = ? AND is_distributor_order = ? AND period_month = ? AND period_year = ?",
			distributorStoreID, true, month, year).
		Count(&totalTx).Error

	var totalStores int64
	_ = u.db.Model(&domain.Store{}).
		Where("distributor_id = ? AND is_active = ?", distributorStoreID, true).
		Count(&totalStores).Error

	var totalVisits int64
	startMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endMonth := startMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	_ = u.db.Model(&domain.SalesVisit{}).
		Where("distributor_store_id = ? AND is_distributor_visit = ? AND check_in_time >= ? AND check_in_time <= ?",
			distributorStoreID, true, startMonth, endMonth).
		Count(&totalVisits).Error

	// Top 5 Products
	var topProducts []domain.ProductRecap
	_ = u.db.Table("sales_items").
		Select("sales_items.product_id, products.sku, products.name as product_name, SUM(sales_items.quantity) as total_qty, SUM(sales_items.subtotal) as total_amount").
		Joins("JOIN sales_transactions ON sales_transactions.id = sales_items.sales_transaction_id").
		Joins("JOIN products ON products.id = sales_items.product_id").
		Where("sales_transactions.distributor_store_id = ? AND sales_transactions.is_distributor_order = ? AND sales_transactions.period_month = ? AND sales_transactions.period_year = ?",
			distributorStoreID, true, month, year).
		Group("sales_items.product_id, products.sku, products.name").
		Order("total_qty DESC").
		Limit(5).
		Scan(&topProducts).Error

	return &DistributorSummaryResponse{
		Month:               month,
		Year:                year,
		TotalSellOutAmount:  totalAmount,
		TotalTransactions:   totalTx,
		TotalAssignedStores: totalStores,
		TotalVisits:         totalVisits,
		TopProducts:         topProducts,
	}, nil
}
