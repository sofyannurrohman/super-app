package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
	"gorm.io/gorm"
)

// LPSHandler menyediakan endpoint laporan LPS (Lembar Penjualan Sales) harian.
// Mendukung mekanisme:
// 1. LPS Sebelum Pengiriman (Pesanan SO terkonfirmasi Admin Nota) untuk Sales TO
// 2. LPS Pengiriman (Realisasi Surat Jalan & Pengiriman) untuk Sales TO / Driver
// 3. LPS Faktur Penjualan untuk Sales Motoris
type LPSHandler struct {
	db           *gorm.DB
	salesUsecase usecase.SalesUsecase
}

func NewLPSHandler(db *gorm.DB, salesUsecase usecase.SalesUsecase) *LPSHandler {
	return &LPSHandler{
		db:           db,
		salesUsecase: salesUsecase,
	}
}

// ─── Response types ────────────────────────────────────────────────────────────

// LPSResponse adalah root response Lembar Penjualan Sales.
type LPSResponse struct {
	Date     string      `json:"date"`
	LPSType  string      `json:"lps_type"` // PRE_DELIVERY, DELIVERY, MOTORIS
	Salesman LPSSalesman `json:"salesman"`

	// Section 1: Barang keluar & kembali dari gudang / rekap pesanan muat
	DeliveryBatches []LPSDeliveryBatch `json:"delivery_batches"`

	// Section 2: Toko & transaksi
	StoreVisits []LPSStoreVisit `json:"store_visits"`

	IsSaved bool       `json:"is_saved"`
	SavedAt *time.Time `json:"saved_at,omitempty"`
	Status  string     `json:"status,omitempty"`

	// Section 3: Grand totals
	TotalGrandAmount   float64 `json:"total_grand_amount"`
	TotalCash          float64 `json:"total_cash"`
	TotalTransfer      float64 `json:"total_transfer"`
	TotalTempo         float64 `json:"total_tempo"`
	TotalStoresVisited int     `json:"total_stores_visited"`
	TotalTransactions  int     `json:"total_transactions"`
}

// LPSSalesman berisi data ringkas salesman.
type LPSSalesman struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EmployeeCode string `json:"employee_code"`
	BranchName   string `json:"branch_name"`
	JobPosition  string `json:"job_position"`
	SalesType    string `json:"sales_type"` // to, motoris, other
}

// LPSDeliveryBatch merepresentasikan satu surat jalan pengiriman atau batch pesanan muat.
type LPSDeliveryBatch struct {
	DeliveryOrderNo  string            `json:"delivery_order_no"`
	VehicleNo        string            `json:"vehicle_no"`
	DeliveryDate     string            `json:"delivery_date"`
	Status           string            `json:"status"`
	Items            []LPSDeliveryItem `json:"items"`
	TotalItemsOut    int               `json:"total_items_out"`
	TotalItemsReturn int               `json:"total_items_return"`
}

// LPSDeliveryItem merepresentasikan satu baris item dalam surat jalan atau rekap muat.
type LPSDeliveryItem struct {
	ProductName    string  `json:"product_name"`
	QuantityOut    int     `json:"quantity_out"`
	QuantityReturn int     `json:"quantity_return"`
	StoreName      string  `json:"store_name"`
	SONumber       string  `json:"so_number"`
	Status         string  `json:"status"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentAmount  float64 `json:"payment_amount"`
}

// LPSStoreVisit merepresentasikan rincian per toko beserta seluruh transaksinya.
type LPSStoreVisit struct {
	StoreID         string           `json:"store_id,omitempty"`
	StoreName       string           `json:"store_name"`
	StoreAddress    string           `json:"store_address"`
	StorePhone      string           `json:"store_phone,omitempty"`
	CheckInTime     string           `json:"check_in_time,omitempty"`
	CheckOutTime    string           `json:"check_out_time,omitempty"`
	Transactions    []LPSTransaction `json:"transactions"`
	TotalAmount     float64          `json:"total_amount"`     // Total nilai tagihan / pesanan
	TotalPembayaran float64          `json:"total_pembayaran"` // Total uang yang diterima/dibayar
	PaymentMethod   string           `json:"payment_method"`
}

// LPSTransaction merepresentasikan satu transaksi (langsung maupun berbasis SO).
type LPSTransaction struct {
	ID            string               `json:"id,omitempty"`
	ReceiptNo     string               `json:"receipt_no"`
	Items         []LPSTransactionItem `json:"items"`
	TotalAmount   float64              `json:"total_amount"`
	PaidAmount    float64              `json:"paid_amount"`
	PaymentMethod string               `json:"payment_method"`
	PaymentStatus string               `json:"payment_status"`
	SONumber      string               `json:"so_number"`
	PaymentAmount float64              `json:"payment_amount"`
	Status        string               `json:"status"`
	IsConfirmed   bool                 `json:"is_confirmed"`
	StatusLabel   string               `json:"status_label"`
}

// LPSTransactionItem merepresentasikan satu baris produk dalam transaksi.
type LPSTransactionItem struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Unit        string  `json:"unit"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
}

// ─── Raw query result structs ───────────────────────────────────────────────

type rawSalesmanRow struct {
	ID           string `gorm:"column:id"`
	FirstName    string `gorm:"column:first_name"`
	LastName     string `gorm:"column:last_name"`
	EmployeeCode string `gorm:"column:employee_code"`
	BranchName   string `gorm:"column:branch_name"`
	JobPosition  string `gorm:"column:job_position"`
}

// ─── Routes ────────────────────────────────────────────────────────────────────

// RegisterRoutes mendaftarkan route LPS ke dalam group yang sudah terproteksi JWT.
func (h *LPSHandler) RegisterRoutes(r *gin.RouterGroup) {
	lps := r.Group("/admin/lps")

	middlewares := RoleMiddleware(
		string(domain.RoleSuperAdmin),
		string(domain.RoleHRAdmin),
		string(domain.RoleManager),
		string(domain.RoleAdminNota),
		string(domain.RoleSupervisorSales),
		string(domain.RoleAdminGudang),
		string(domain.RoleCashier),
		string(domain.RoleAdminKasir),
	)

	lps.GET("/daily", middlewares, h.GetDailyLPS)
	lps.POST("/daily/save", middlewares, h.SaveDailyLPS)
	lps.POST("/daily/approve", middlewares, h.ApproveDailyLPS)
}

// ─── Handler ───────────────────────────────────────────────────────────────────

// GetDailyLPS mengumpulkan data LPS harian untuk satu salesman.
//
// Query params:
//   - date        (required)  : tanggal laporan, format YYYY-MM-DD
//   - employee_id (optional)  : UUID salesman; jika tidak diberikan akan menggunakan employee_id dari token
//   - branch_id   (optional)  : UUID cabang untuk filter tambahan
//   - lps_type    (optional)  : pre_delivery, delivery, motoris (otomatis ditentukan jika kosong)
func (h *LPSHandler) GetDailyLPS(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query param 'date' is required (format: YYYY-MM-DD)"})
		return
	}

	employeeID := c.Query("employee_id")
	if employeeID == "" {
		if val, ok := c.Get("employee_id"); ok {
			employeeID = val.(string)
		}
	}
	if employeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required (query param or login context)"})
		return
	}

	branchID := c.Query("branch_id")
	lpsTypeParam := strings.ToLower(c.Query("lps_type"))

	// 1. Data salesman & Posisi ─────────────────────────────────────────────────
	var salesmanRow rawSalesmanRow
	salesmanQuery := h.db.Raw(`
		SELECT
			e.id::text             AS id,
			e.first_name           AS first_name,
			e.last_name            AS last_name,
			e.employee_id_number   AS employee_code,
			COALESCE(b.name, '')   AS branch_name,
			COALESCE(jp.title, '') AS job_position
		FROM employees e
		LEFT JOIN branches b ON b.id = e.branch_id
		LEFT JOIN job_positions jp ON jp.id = e.job_position_id
		WHERE e.id = ?
		LIMIT 1
	`, employeeID)

	if err := salesmanQuery.Scan(&salesmanRow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch salesman data: " + err.Error()})
		return
	}

	titleLower := strings.ToLower(salesmanRow.JobPosition)
	isMotoris := strings.Contains(titleLower, "motoris") || strings.Contains(titleLower, "booster")
	salesType := "to"
	if isMotoris {
		salesType = "motoris"
	}

	salesman := LPSSalesman{
		ID:           salesmanRow.ID,
		Name:         strings.TrimSpace(salesmanRow.FirstName + " " + salesmanRow.LastName),
		EmployeeCode: salesmanRow.EmployeeCode,
		BranchName:   salesmanRow.BranchName,
		JobPosition:  salesmanRow.JobPosition,
		SalesType:    salesType,
	}

	// Tentukan target LPS Type
	var effectiveLPSType string
	switch lpsTypeParam {
	case "pre_delivery", "pre-delivery", "sebelum_pengiriman", "so":
		effectiveLPSType = "PRE_DELIVERY"
	case "delivery", "pengiriman", "sj", "surat_jalan":
		effectiveLPSType = "DELIVERY"
	case "motoris", "faktur", "canvas":
		effectiveLPSType = "MOTORIS"
	default:
		if isMotoris {
			effectiveLPSType = "MOTORIS"
		} else {
			effectiveLPSType = "PRE_DELIVERY"
		}
	}

	// 2. Cek apakah sudah ada snapshot tersimpan di database untuk tipe ini ─────
	forceLive := c.Query("force_live") == "true" || c.Query("refresh") == "true"
	var savedReport domain.LPSReport
	checkSavedQuery := h.db.Where("employee_id::text = ? AND report_date = ?", employeeID, dateStr)
	if effectiveLPSType != "" {
		checkSavedQuery = checkSavedQuery.Where("lps_type = ?", effectiveLPSType)
	}
	if !forceLive && checkSavedQuery.Order("created_at DESC").First(&savedReport).Error == nil && savedReport.SnapshotData != "" {
		var resp LPSResponse
		if err := json.Unmarshal([]byte(savedReport.SnapshotData), &resp); err == nil {
			if len(resp.StoreVisits) > 0 && savedReport.Status == "APPROVED" {
				resp.IsSaved = true
				resp.SavedAt = &savedReport.CreatedAt
				resp.Status = savedReport.Status
				if resp.LPSType == "" {
					resp.LPSType = effectiveLPSType
				}
				c.JSON(http.StatusOK, resp)
				return
			}
		}
	}

	// 3. Bangun data on the fly sesuai effectiveLPSType (TANPA sales_visits) ─────
	var resp LPSResponse
	switch effectiveLPSType {
	case "PRE_DELIVERY":
		resp = h.buildPreDeliveryLPS(dateStr, employeeID, branchID, salesman)
	case "DELIVERY":
		resp = h.buildDeliveryLPS(dateStr, employeeID, branchID, salesman)
	case "MOTORIS":
		resp = h.buildMotorisLPS(dateStr, employeeID, branchID, salesman)
	default:
		resp = h.buildPreDeliveryLPS(dateStr, employeeID, branchID, salesman)
	}

	c.JSON(http.StatusOK, resp)
}

// ─── BUILDER 1: PRE_DELIVERY (Sales TO - Sebelum Pengiriman / Pesanan SO) ───────
func (h *LPSHandler) buildPreDeliveryLPS(dateStr, employeeID, branchID string, salesman LPSSalesman) LPSResponse {
	type rawSORow struct {
		SOID          string  `gorm:"column:so_id"`
		SONumber      string  `gorm:"column:so_number"`
		StoreID       string  `gorm:"column:store_id"`
		StoreName     string  `gorm:"column:store_name"`
		StoreAddress  string  `gorm:"column:store_address"`
		StorePhone    string  `gorm:"column:store_phone"`
		TotalAmount   float64 `gorm:"column:total_amount"`
		PaymentMethod string  `gorm:"column:payment_method"`
		PaymentStatus string  `gorm:"column:payment_status"`
		Status        string  `gorm:"column:status"`
		ConfirmedAt   *string `gorm:"column:confirmed_at"`
	}

	soArgs := []interface{}{
		employeeID, employeeID,
		dateStr, dateStr, dateStr, dateStr, dateStr, dateStr, dateStr, dateStr, dateStr, dateStr,
	}

	var rawSoRows []rawSORow
	soSQL := `
		SELECT
			so.id::text                                           AS so_id,
			so.so_number                                          AS so_number,
			so.store_id::text                                     AS store_id,
			COALESCE(st.name, 'Toko Non-Terdaftar')               AS store_name,
			COALESCE(st.address, '')                              AS store_address,
			COALESCE(st.phone_number, '')                         AS store_phone,
			COALESCE(so.total_amount, 0)                          AS total_amount,
			COALESCE(so.payment_method, 'TEMPO')                  AS payment_method,
			COALESCE(so.payment_status, 'UNPAID')                 AS payment_status,
			COALESCE(so.status, '')                               AS status,
			so.confirmed_at::text                                 AS confirmed_at
		FROM sales_orders so
		LEFT JOIN stores st ON st.id = so.store_id
		WHERE (so.employee_id::text = ? OR st.assigned_employee_id::text = ?)
		  AND (
		    DATE(so.confirmed_at AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(so.confirmed_at) = ?::date
		    OR DATE(so.created_at AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(so.created_at) = ?::date
		    OR DATE(so.order_date AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(so.order_date) = ?::date
		    OR TO_CHAR(so.order_date AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') = ?
		    OR TO_CHAR(so.order_date, 'YYYY-MM-DD') = ?
		    OR TO_CHAR(so.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD') = ?
		    OR TO_CHAR(so.created_at, 'YYYY-MM-DD') = ?
		  )
		  AND UPPER(so.status) NOT IN ('CANCELLED', 'REJECTED')
		  AND so.deleted_at IS NULL
		ORDER BY so.created_at ASC, so.so_number ASC
	`

	if err := h.db.Raw(soSQL, soArgs...).Scan(&rawSoRows).Error; err != nil {
		log.Printf("[LPS-PRE_DELIVERY] Query error: %v", err)
	}

	// Deduplikasi soRows berdasarkan SOID
	seenSO := make(map[string]bool)
	var soRows []rawSORow
	var soIDs []string
	for _, so := range rawSoRows {
		if so.SOID != "" && !seenSO[so.SOID] {
			seenSO[so.SOID] = true
			soRows = append(soRows, so)
			soIDs = append(soIDs, so.SOID)
		}
	}

	// Ambil item per SO
	type rawSOItemRow struct {
		SOID        string  `gorm:"column:so_id"`
		ProductName string  `gorm:"column:product_name"`
		OrderedQty  int     `gorm:"column:ordered_qty"`
		Unit        string  `gorm:"column:unit"`
		Price       float64 `gorm:"column:price"`
		Subtotal    float64 `gorm:"column:subtotal"`
	}

	soItemsBySOID := map[string][]LPSTransactionItem{}
	productSummaryMap := map[string]int{}

	if len(soIDs) > 0 {
		var itemRows []rawSOItemRow
		itemSQL := `
			SELECT
				soi.sales_order_id::text  AS so_id,
				COALESCE(p.name, '')       AS product_name,
				COALESCE(NULLIF(soi.ordered_quantity, 0), soi.quantity, 0) AS ordered_qty,
				COALESCE(soi.unit, p.unit, 'PCS')     AS unit,
				COALESCE(soi.price, 0)    AS price,
				COALESCE(soi.subtotal, 0) AS subtotal
			FROM sales_order_items soi
			LEFT JOIN products p ON p.id = soi.product_id
			WHERE soi.sales_order_id::text IN (?)
			ORDER BY soi.sales_order_id, soi.id
		`
		if err := h.db.Raw(itemSQL, soIDs).Scan(&itemRows).Error; err == nil {
			for _, item := range itemRows {
				if item.ProductName == "" {
					continue
				}
				soItemsBySOID[item.SOID] = append(soItemsBySOID[item.SOID], LPSTransactionItem{
					ProductName: item.ProductName,
					Quantity:    item.OrderedQty,
					Unit:        item.Unit,
					Price:       item.Price,
					Subtotal:    item.Subtotal,
				})
				productSummaryMap[item.ProductName] += item.OrderedQty
			}
		}
	}

	// Bangun Delivery Batches (Rekap Kebutuhan Muat Gudang)
	var deliveryBatches []LPSDeliveryBatch
	if len(productSummaryMap) > 0 {
		batchItems := make([]LPSDeliveryItem, 0, len(productSummaryMap))
		totalOut := 0
		for pName, qty := range productSummaryMap {
			batchItems = append(batchItems, LPSDeliveryItem{
				ProductName:    pName,
				QuantityOut:    qty,
				QuantityReturn: 0,
				StoreName:      "Gudang (Rencana Muat)",
				SONumber:       "RENCANA_MUAT",
				Status:         "CONFIRMED",
			})
			totalOut += qty
		}
		deliveryBatches = append(deliveryBatches, LPSDeliveryBatch{
			DeliveryOrderNo:  "RENCANA MUAT (PESANAN SO)",
			VehicleNo:        "-",
			DeliveryDate:     dateStr,
			Status:           "CONFIRMED",
			Items:            batchItems,
			TotalItemsOut:    totalOut,
			TotalItemsReturn: 0,
		})
	}

	// Bangun StoreVisits
	storeVisits := make([]LPSStoreVisit, 0, len(soRows))
	var totalGrand, totalCash, totalTransfer, totalTempo float64

	for _, so := range soRows {
		items := soItemsBySOID[so.SOID]
		if items == nil {
			items = []LPSTransactionItem{}
		}

		stUpper := strings.ToUpper(strings.TrimSpace(so.Status))
		isConfirmed := (so.ConfirmedAt != nil && *so.ConfirmedAt != "") ||
			stUpper == "CONFIRMED" || stUpper == "APPROVED" || stUpper == "IN_DELIVERY" ||
			stUpper == "SHIPPED" || stUpper == "DELIVERED" || stUpper == "CONVERTED" ||
			stUpper == "PAID"

		statusLabel := "Belum Dikonfirmasi Admin"
		if stUpper == "CONVERTED" {
			statusLabel = "Terkonversi ke Faktur"
		} else if isConfirmed {
			statusLabel = "Terkonfirmasi Admin"
		}

		tx := LPSTransaction{
			ID:            so.SOID,
			ReceiptNo:     so.SONumber,
			Items:         items,
			TotalAmount:   so.TotalAmount,
			PaidAmount:    0, // Belum ditagih/dikirim
			PaymentMethod: so.PaymentMethod,
			PaymentStatus: so.PaymentStatus,
			SONumber:      so.SONumber,
			PaymentAmount: 0,
			Status:        so.Status,
			IsConfirmed:   isConfirmed,
			StatusLabel:   statusLabel,
		}

		sv := LPSStoreVisit{
			StoreID:         so.StoreID,
			StoreName:       so.StoreName,
			StoreAddress:    so.StoreAddress,
			StorePhone:      so.StorePhone,
			Transactions:    []LPSTransaction{tx},
			TotalAmount:     so.TotalAmount,
			TotalPembayaran: 0,
			PaymentMethod:   so.PaymentMethod,
		}

		storeVisits = append(storeVisits, sv)
		totalGrand += so.TotalAmount

		switch normalizePaymentMethod(so.PaymentMethod) {
		case "cash":
			totalCash += so.TotalAmount
		case "transfer":
			totalTransfer += so.TotalAmount
		case "tempo":
			totalTempo += so.TotalAmount
		default:
			totalTempo += so.TotalAmount
		}
	}

	return LPSResponse{
		Date:               dateStr,
		LPSType:            "PRE_DELIVERY",
		Salesman:           salesman,
		DeliveryBatches:    deliveryBatches,
		StoreVisits:        storeVisits,
		TotalGrandAmount:   totalGrand,
		TotalCash:          totalCash,
		TotalTransfer:      totalTransfer,
		TotalTempo:         totalTempo,
		TotalStoresVisited: len(storeVisits),
		TotalTransactions:  len(soRows),
	}
}

// ─── BUILDER 2: DELIVERY (Sales TO - Realisasi Pengiriman / Surat Jalan) ─────────
func (h *LPSHandler) buildDeliveryLPS(dateStr, employeeID, branchID string, salesman LPSSalesman) LPSResponse {
	type rawDeliveryBatchItemRow struct {
		BatchID          string  `gorm:"column:batch_id"`
		DeliveryOrderNo  string  `gorm:"column:delivery_order_no"`
		VehiclePlateNo   string  `gorm:"column:vehicle_plate_no"`
		DeliveryDate     string  `gorm:"column:delivery_date"`
		BatchStatus      string  `gorm:"column:batch_status"`
		DeliveryItemID   string  `gorm:"column:delivery_item_id"`
		ItemStatus       string  `gorm:"column:item_status"`
		ReturnQuantity   int     `gorm:"column:return_quantity"`
		HasReturn        bool    `gorm:"column:has_return"`
		PaymentCollected bool    `gorm:"column:payment_collected"`
		PaymentAmount    float64 `gorm:"column:payment_amount"`
		PaymentMethod    string  `gorm:"column:payment_method"`
		SOID             string  `gorm:"column:so_id"`
		TxID             string  `gorm:"column:tx_id"`
		SONumber         string  `gorm:"column:so_number"`
		SOTotalAmount    float64 `gorm:"column:so_total_amount"`
		SOPaymentMethod  string  `gorm:"column:so_payment_method"`
		StoreID          string  `gorm:"column:store_id"`
		StoreName        string  `gorm:"column:store_name"`
		StoreAddress     string  `gorm:"column:store_address"`
		StorePhone       string  `gorm:"column:store_phone"`
	}

	deliveryArgs := []interface{}{
		dateStr, dateStr, dateStr, dateStr, dateStr, dateStr, dateStr, dateStr, dateStr, dateStr,
		employeeID, employeeID, employeeID, employeeID,
	}

	var deliveryRows []rawDeliveryBatchItemRow
	deliverySQL := `
		SELECT
			db.id::text                                           AS batch_id,
			COALESCE(db.delivery_order_no, '')                   AS delivery_order_no,
			COALESCE(v.license_plate, v.name, '')                 AS vehicle_plate_no,
			TO_CHAR(COALESCE(db.delivery_date, db.created_at), 'YYYY-MM-DD') AS delivery_date,
			db.status                                             AS batch_status,
			di.id::text                                           AS delivery_item_id,
			di.status                                             AS item_status,
			COALESCE(di.return_quantity, 0)                       AS return_quantity,
			COALESCE(di.has_return, FALSE)                        AS has_return,
			COALESCE(di.payment_collected, FALSE)                 AS payment_collected,
			COALESCE(NULLIF(di.payment_amount, 0), so.total_amount, stx.total_amount, 0) AS payment_amount,
			COALESCE(NULLIF(di.payment_method, ''), so.payment_method, stx.payment_method, 'TEMPO') AS payment_method,
			COALESCE(so.id::text, '')                             AS so_id,
			COALESCE(stx.id::text, '')                            AS tx_id,
			COALESCE(so.so_number, stx.receipt_no, '')            AS so_number,
			COALESCE(so.total_amount, stx.total_amount, 0)        AS so_total_amount,
			COALESCE(so.payment_method, stx.payment_method, 'TEMPO') AS so_payment_method,
			COALESCE(st.id::text, s_stx.id::text, '')             AS store_id,
			COALESCE(st.name, s_stx.name, 'Toko Non-Terdaftar')   AS store_name,
			COALESCE(st.address, s_stx.address, '')               AS store_address,
			COALESCE(st.phone_number, s_stx.phone_number, '')    AS store_phone
		FROM delivery_batches db
		JOIN delivery_items di ON di.delivery_batch_id = db.id
		LEFT JOIN vehicles v ON v.id = db.vehicle_id
		LEFT JOIN sales_orders so ON so.id = di.sales_order_id
		LEFT JOIN stores st ON st.id = so.store_id
		LEFT JOIN sales_transactions stx ON stx.id = di.sales_transaction_id
		LEFT JOIN stores s_stx ON s_stx.id = stx.store_id
		WHERE (
		    DATE(db.delivery_date) = ?::date
		    OR DATE(db.delivery_date AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(db.created_at AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(db.created_at) = ?::date
		    OR DATE(db.approved_at AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(so.created_at AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(so.order_date AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(so.order_date) = ?::date
		    OR TO_CHAR(db.delivery_date, 'YYYY-MM-DD') = ?
		    OR TO_CHAR(so.order_date, 'YYYY-MM-DD') = ?
		  )
		  AND (db.driver_id::text = ? OR so.employee_id::text = ? OR stx.employee_id::text = ? OR st.assigned_employee_id::text = ?)
		  AND UPPER(db.status) NOT IN ('CANCELLED', 'REJECTED')
		ORDER BY db.delivery_order_no, di.sequence, di.id
	`

	if err := h.db.Raw(deliverySQL, deliveryArgs...).Scan(&deliveryRows).Error; err != nil {
		log.Printf("[LPS-DELIVERY] Query error: %v", err)
	}

	var soIDs []string
	var txIDs []string
	soIDSet := map[string]bool{}
	txIDSet := map[string]bool{}
	for _, r := range deliveryRows {
		if r.SOID != "" && !soIDSet[r.SOID] {
			soIDs = append(soIDs, r.SOID)
			soIDSet[r.SOID] = true
		}
		if r.TxID != "" && !txIDSet[r.TxID] {
			txIDs = append(txIDs, r.TxID)
			txIDSet[r.TxID] = true
		}
	}

	// Fetch SO Items
	type rawSOItemRow struct {
		SOID        string  `gorm:"column:so_id"`
		ProductName string  `gorm:"column:product_name"`
		OrderedQty  int     `gorm:"column:ordered_qty"`
		Unit        string  `gorm:"column:unit"`
		Price       float64 `gorm:"column:price"`
		Subtotal    float64 `gorm:"column:subtotal"`
	}

	soItemsByDocID := map[string][]LPSTransactionItem{}
	if len(soIDs) > 0 {
		var itemRows []rawSOItemRow
		itemSQL := `
			SELECT
				soi.sales_order_id::text  AS so_id,
				COALESCE(p.name, '')       AS product_name,
				COALESCE(NULLIF(soi.ordered_quantity, 0), soi.quantity, 0) AS ordered_qty,
				COALESCE(soi.unit, p.unit, 'PCS') AS unit,
				COALESCE(soi.price, 0)    AS price,
				COALESCE(soi.subtotal, 0) AS subtotal
			FROM sales_order_items soi
			LEFT JOIN products p ON p.id = soi.product_id
			WHERE soi.sales_order_id::text IN (?)
			ORDER BY soi.sales_order_id, soi.id
		`
		if err := h.db.Raw(itemSQL, soIDs).Scan(&itemRows).Error; err == nil {
			for _, item := range itemRows {
				if item.ProductName == "" {
					continue
				}
				soItemsByDocID[item.SOID] = append(soItemsByDocID[item.SOID], LPSTransactionItem{
					ProductName: item.ProductName,
					Quantity:    item.OrderedQty,
					Unit:        item.Unit,
					Price:       item.Price,
					Subtotal:    item.Subtotal,
				})
			}
		}
	}

	// Also fetch from sales_items for transactions linked in delivery items
	if len(txIDs) > 0 {
		var txItemRows []struct {
			TxID        string  `gorm:"column:tx_id"`
			ProductName string  `gorm:"column:product_name"`
			Quantity    int     `gorm:"column:quantity"`
			Unit        string  `gorm:"column:unit"`
			Price       float64 `gorm:"column:price"`
			Subtotal    float64 `gorm:"column:subtotal"`
		}
		txItemSQL := `
			SELECT
				si.sales_transaction_id::text AS tx_id,
				COALESCE(p.name, '')           AS product_name,
				COALESCE(si.quantity, 0)       AS quantity,
				COALESCE(p.unit, 'PCS')        AS unit,
				COALESCE(si.price_at_transaction, si.price, 0) AS price,
				COALESCE(si.subtotal, 0)       AS subtotal
			FROM sales_items si
			LEFT JOIN products p ON p.id = si.product_id
			WHERE si.sales_transaction_id::text IN (?)
			ORDER BY si.sales_transaction_id, si.id
		`
		if err := h.db.Raw(txItemSQL, txIDs).Scan(&txItemRows).Error; err == nil {
			for _, item := range txItemRows {
				if item.ProductName == "" {
					continue
				}
				soItemsByDocID[item.TxID] = append(soItemsByDocID[item.TxID], LPSTransactionItem{
					ProductName: item.ProductName,
					Quantity:    item.Quantity,
					Unit:        item.Unit,
					Price:       item.Price,
					Subtotal:    item.Subtotal,
				})
			}
		}
	}

	// Group delivery items by Delivery Batch
	type batchData struct {
		Batch            LPSDeliveryBatch
		TotalItemsOut    int
		TotalItemsReturn int
	}
	batchMap := map[string]*LPSDeliveryBatch{}
	batchOrder := []string{}

	for _, row := range deliveryRows {
		if _, exists := batchMap[row.BatchID]; !exists {
			batchMap[row.BatchID] = &LPSDeliveryBatch{
				DeliveryOrderNo:  row.DeliveryOrderNo,
				VehicleNo:        row.VehiclePlateNo,
				DeliveryDate:     row.DeliveryDate,
				Status:           row.BatchStatus,
				Items:            []LPSDeliveryItem{},
				TotalItemsOut:    0,
				TotalItemsReturn: 0,
			}
			batchOrder = append(batchOrder, row.BatchID)
		}

		b := batchMap[row.BatchID]
		docKey := row.SOID
		if docKey == "" {
			docKey = row.TxID
		}
		items := soItemsByDocID[docKey]
		if len(items) > 0 {
			for idx, itm := range items {
				retQty := 0
				if idx == 0 {
					retQty = row.ReturnQuantity
				}
				b.Items = append(b.Items, LPSDeliveryItem{
					ProductName:    itm.ProductName,
					QuantityOut:    itm.Quantity,
					QuantityReturn: retQty,
					StoreName:      row.StoreName,
					SONumber:       row.SONumber,
					Status:         row.ItemStatus,
					PaymentMethod:  row.PaymentMethod,
					PaymentAmount:  row.PaymentAmount,
				})
				b.TotalItemsOut += itm.Quantity
				b.TotalItemsReturn += retQty
			}
		} else {
			b.Items = append(b.Items, LPSDeliveryItem{
				ProductName:    "Barang Kiriman",
				QuantityOut:    1,
				QuantityReturn: row.ReturnQuantity,
				StoreName:      row.StoreName,
				SONumber:       row.SONumber,
				Status:         row.ItemStatus,
				PaymentMethod:  row.PaymentMethod,
				PaymentAmount:  row.PaymentAmount,
			})
			b.TotalItemsOut += 1
			b.TotalItemsReturn += row.ReturnQuantity
		}
	}

	deliveryBatches := make([]LPSDeliveryBatch, 0, len(batchOrder))
	for _, id := range batchOrder {
		deliveryBatches = append(deliveryBatches, *batchMap[id])
	}

	// Build StoreVisits with deduplication per document (SO or Tx)
	seenDoc := make(map[string]bool)
	storeVisits := make([]LPSStoreVisit, 0, len(deliveryRows))
	var totalGrand, totalCash, totalTransfer, totalTempo float64
	totalTxCount := 0

	for _, row := range deliveryRows {
		docKey := row.SOID
		if docKey == "" {
			docKey = row.TxID
		}
		if docKey == "" {
			docKey = row.SONumber
		}

		// Prevent duplicate visit entry for the same document in delivery batch
		if docKey != "" && seenDoc[docKey] {
			continue
		}
		if docKey != "" {
			seenDoc[docKey] = true
		}
		totalTxCount++

		items := soItemsByDocID[docKey]
		if items == nil {
			items = []LPSTransactionItem{}
		}

		effectivePaymentMethod := row.PaymentMethod
		if effectivePaymentMethod == "" {
			effectivePaymentMethod = row.SOPaymentMethod
		}
		if effectivePaymentMethod == "" {
			effectivePaymentMethod = "TEMPO"
		}

		tx := LPSTransaction{
			ID:            docKey,
			ReceiptNo:     row.SONumber,
			Items:         items,
			TotalAmount:   row.SOTotalAmount,
			PaidAmount:    row.PaymentAmount,
			PaymentMethod: effectivePaymentMethod,
			PaymentStatus: row.ItemStatus,
			SONumber:      row.SONumber,
			PaymentAmount: row.PaymentAmount,
			Status:        row.ItemStatus,
			IsConfirmed:   true,
			StatusLabel:   "Terkonfirmasi Admin",
		}

		sv := LPSStoreVisit{
			StoreID:         row.StoreID,
			StoreName:       row.StoreName,
			StoreAddress:    row.StoreAddress,
			StorePhone:      row.StorePhone,
			Transactions:    []LPSTransaction{tx},
			TotalAmount:     row.SOTotalAmount,
			TotalPembayaran: row.PaymentAmount,
			PaymentMethod:   effectivePaymentMethod,
		}

		storeVisits = append(storeVisits, sv)
		totalGrand += row.PaymentAmount

		switch normalizePaymentMethod(effectivePaymentMethod) {
		case "cash":
			totalCash += row.PaymentAmount
		case "transfer":
			totalTransfer += row.PaymentAmount
		case "tempo":
			totalTempo += row.PaymentAmount
		default:
			totalCash += row.PaymentAmount
		}
	}

	return LPSResponse{
		Date:               dateStr,
		LPSType:            "DELIVERY",
		Salesman:           salesman,
		DeliveryBatches:    deliveryBatches,
		StoreVisits:        storeVisits,
		TotalGrandAmount:   totalGrand,
		TotalCash:          totalCash,
		TotalTransfer:      totalTransfer,
		TotalTempo:         totalTempo,
		TotalStoresVisited: len(storeVisits),
		TotalTransactions:  totalTxCount,
	}
}

// ─── BUILDER 3: MOTORIS (Sales Motoris - Faktur Penjualan Langsung) ─────────────
func (h *LPSHandler) buildMotorisLPS(dateStr, employeeID, branchID string, salesman LPSSalesman) LPSResponse {
	type rawTxRow struct {
		TxID          string  `gorm:"column:tx_id"`
		ReceiptNo     string  `gorm:"column:receipt_no"`
		TotalAmount   float64 `gorm:"column:total_amount"`
		PaidAmount    float64 `gorm:"column:paid_amount"`
		PaymentMethod string  `gorm:"column:payment_method"`
		PaymentStatus string  `gorm:"column:payment_status"`
		Status        string  `gorm:"column:status"`
		StoreID       string  `gorm:"column:store_id"`
		StoreName     string  `gorm:"column:store_name"`
		StoreAddress  string  `gorm:"column:store_address"`
		StorePhone    string  `gorm:"column:store_phone"`
		VerifiedAt    *string `gorm:"column:verified_at"`
	}

	txArgs := []interface{}{
		employeeID, employeeID,
		dateStr,
	}

	var rawTxRows []rawTxRow
	txSQL := `
		SELECT
			stx.id::text                                          AS tx_id,
			COALESCE(stx.receipt_no, '')                          AS receipt_no,
			COALESCE(stx.total_amount, 0)                         AS total_amount,
			COALESCE(NULLIF(stx.paid_amount, 0), stx.total_amount, 0) AS paid_amount,
			COALESCE(stx.payment_method, 'CASH')                  AS payment_method,
			COALESCE(stx.payment_status, 'PAID')                  AS payment_status,
			COALESCE(stx.status, '')                              AS status,
			COALESCE(stx.store_id::text, '')                      AS store_id,
			COALESCE(s.name, stx.buyer_name, 'Penjualan Langsung') AS store_name,
			COALESCE(s.address, '')                               AS store_address,
			COALESCE(s.phone_number, stx.buyer_phone, '')         AS store_phone,
			stx.verified_at::text                                 AS verified_at
		FROM sales_transactions stx
		LEFT JOIN stores s ON s.id = stx.store_id
		WHERE (stx.employee_id::text = ? OR s.assigned_employee_id::text = ?)
		  -- LPS MOTORIS is a report of sales made on the selected date.
		  -- Do not fall back to created_at/verified_at: approval can happen on a
		  -- different day and must not move an older sale into today's LPS.
		  AND DATE(stx.transaction_date AT TIME ZONE 'Asia/Jakarta') = ?::date
		  AND UPPER(stx.status) NOT IN ('CANCELLED', 'REJECTED', 'VOID', 'DELETED')
		  AND stx.deleted_at IS NULL
		ORDER BY stx.created_at ASC, stx.receipt_no ASC
	`

	if err := h.db.Raw(txSQL, txArgs...).Scan(&rawTxRows).Error; err != nil {
		log.Printf("[LPS-MOTORIS] Query error: %v", err)
	}

	seenTx := make(map[string]bool)
	var txRows []rawTxRow
	var txIDs []string
	for _, tx := range rawTxRows {
		if tx.TxID != "" && !seenTx[tx.TxID] {
			seenTx[tx.TxID] = true
			txRows = append(txRows, tx)
			txIDs = append(txIDs, tx.TxID)
		}
	}

	// Fetch transaction items
	type rawItemRow struct {
		TxID        string  `gorm:"column:tx_id"`
		ProductName string  `gorm:"column:product_name"`
		Quantity    int     `gorm:"column:quantity"`
		Unit        string  `gorm:"column:unit"`
		Price       float64 `gorm:"column:price"`
		Subtotal    float64 `gorm:"column:subtotal"`
	}

	itemsByTxID := map[string][]LPSTransactionItem{}
	if len(txIDs) > 0 {
		var itemRows []rawItemRow
		itemSQL := `
			SELECT
				si.sales_transaction_id::text AS tx_id,
				COALESCE(p.name, '')           AS product_name,
				COALESCE(si.quantity, 0)       AS quantity,
				COALESCE(p.unit, 'PCS')        AS unit,
				COALESCE(si.price_at_transaction, si.price, 0) AS price,
				COALESCE(si.subtotal, 0)       AS subtotal
			FROM sales_items si
			LEFT JOIN products p ON p.id = si.product_id
			WHERE si.sales_transaction_id::text IN (?)
			ORDER BY si.sales_transaction_id, si.id
		`
		if err := h.db.Raw(itemSQL, txIDs).Scan(&itemRows).Error; err == nil {
			for _, item := range itemRows {
				if item.ProductName == "" {
					continue
				}
				itemsByTxID[item.TxID] = append(itemsByTxID[item.TxID], LPSTransactionItem{
					ProductName: item.ProductName,
					Quantity:    item.Quantity,
					Unit:        item.Unit,
					Price:       item.Price,
					Subtotal:    item.Subtotal,
				})
			}
		}
	}

	// Fallback untuk transaksi fisik/langsung yang belum memiliki sales_items (e.g. Butuh Input Foto Nota)
	for _, tx := range txRows {
		if len(itemsByTxID[tx.TxID]) == 0 {
			itemsByTxID[tx.TxID] = []LPSTransactionItem{
				{
					ProductName: "Penjualan Langsung / Nota Fisik",
					Quantity:    1,
					Unit:        "NOTA",
					Price:       tx.TotalAmount,
					Subtotal:    tx.TotalAmount,
				},
			}
		}
	}

	// Fetch Motoris stock movements (sales_transfers & bonus_requests)
	var transferRows []struct {
		ProductName string `gorm:"column:product_name"`
		Quantity    int    `gorm:"column:quantity"`
		Type        string `gorm:"column:type"`
	}

	transferSQL := `
		SELECT
			COALESCE(p.name, '') AS product_name,
			COALESCE(NULLIF(st.actual_quantity, 0), st.quantity, 0) AS quantity,
			st.type              AS type
		FROM sales_transfers st
		LEFT JOIN products p ON p.id = st.product_id
		WHERE (
		    st.transfer_date::date = ?::date
		    OR DATE(st.transfer_date AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(st.created_at AT TIME ZONE 'Asia/Jakarta') = ?::date
		  )
		  AND st.employee_id = ?
		  AND st.status = 'COMPLETED'
	`
	_ = h.db.Raw(transferSQL, dateStr, dateStr, dateStr, employeeID).Scan(&transferRows)

	var bonusRows []struct {
		ProductName string `gorm:"column:product_name"`
		Quantity    int    `gorm:"column:quantity"`
		Type        string `gorm:"column:type"`
	}
	bonusSQL := `
		SELECT
			COALESCE(bi.name, '') AS product_name,
			sb.quantity           AS quantity,
			'BONUS_REQUEST'       AS type
		FROM sales_bonus_requests sb
		LEFT JOIN bonus_inventories bi ON bi.id = sb.bonus_inventory_id
		WHERE (
		    DATE(sb.updated_at AT TIME ZONE 'Asia/Jakarta') = ?::date
		    OR DATE(sb.created_at AT TIME ZONE 'Asia/Jakarta') = ?::date
		  )
		  AND sb.employee_id = ?
		  AND sb.status = 'APPROVED'
	`
	_ = h.db.Raw(bonusSQL, dateStr, dateStr, employeeID).Scan(&bonusRows)
	transferRows = append(transferRows, bonusRows...)

	var deliveryBatches []LPSDeliveryBatch
	if len(transferRows) > 0 {
		directBatch := LPSDeliveryBatch{
			DeliveryOrderNo: "REKAP STOK BAWAAN MOTORIS",
			VehicleNo:       "-",
			DeliveryDate:    dateStr,
			Status:          "COMPLETED",
			Items:           []LPSDeliveryItem{},
		}

		type prodMut struct {
			out int
			ret int
		}
		mutations := map[string]*prodMut{}

		for _, r := range transferRows {
			if r.ProductName == "" {
				continue
			}
			if _, ok := mutations[r.ProductName]; !ok {
				mutations[r.ProductName] = &prodMut{}
			}
			if r.Type == "TRANSFER" || r.Type == "BONUS_REQUEST" {
				mutations[r.ProductName].out += r.Quantity
			} else if r.Type == "RETURN" {
				mutations[r.ProductName].ret += r.Quantity
			}
		}

		for pName, mut := range mutations {
			if mut.out == 0 && mut.ret == 0 {
				continue
			}
			item := LPSDeliveryItem{
				ProductName:    pName,
				QuantityOut:    mut.out,
				QuantityReturn: mut.ret,
				StoreName:      "Gudang",
				SONumber:       "DIRECT",
				Status:         "COMPLETED",
			}
			directBatch.Items = append(directBatch.Items, item)
			directBatch.TotalItemsOut += mut.out
			directBatch.TotalItemsReturn += mut.ret
		}

		if len(directBatch.Items) > 0 {
			deliveryBatches = append(deliveryBatches, directBatch)
		}
	}

	// Assemble StoreVisits
	storeVisits := make([]LPSStoreVisit, 0, len(txRows))
	var totalGrand, totalCash, totalTransfer, totalTempo float64

	for _, tx := range txRows {
		items := itemsByTxID[tx.TxID]
		if items == nil {
			items = []LPSTransactionItem{}
		}

		stUpper := strings.ToUpper(strings.TrimSpace(tx.Status))
		payUpper := strings.ToUpper(strings.TrimSpace(tx.PaymentStatus))
		isConfirmed := (tx.VerifiedAt != nil && *tx.VerifiedAt != "") ||
			stUpper == "VERIFIED" || stUpper == "APPROVED" || stUpper == "COMPLETED" ||
			payUpper == "PAID"

		statusLabel := "Belum Dikonfirmasi Admin"
		if isConfirmed {
			statusLabel = "Terkonfirmasi Admin"
		}

		lpsTx := LPSTransaction{
			ID:            tx.TxID,
			ReceiptNo:     tx.ReceiptNo,
			Items:         items,
			TotalAmount:   tx.TotalAmount,
			PaidAmount:    tx.PaidAmount,
			PaymentMethod: tx.PaymentMethod,
			PaymentStatus: tx.PaymentStatus,
			SONumber:      "",
			PaymentAmount: tx.PaidAmount,
			Status:        tx.Status,
			IsConfirmed:   isConfirmed,
			StatusLabel:   statusLabel,
		}

		sv := LPSStoreVisit{
			StoreID:         tx.StoreID,
			StoreName:       tx.StoreName,
			StoreAddress:    tx.StoreAddress,
			StorePhone:      tx.StorePhone,
			Transactions:    []LPSTransaction{lpsTx},
			TotalAmount:     tx.TotalAmount,
			TotalPembayaran: tx.PaidAmount,
			PaymentMethod:   tx.PaymentMethod,
		}

		storeVisits = append(storeVisits, sv)
		totalGrand += tx.PaidAmount

		switch normalizePaymentMethod(tx.PaymentMethod) {
		case "cash":
			totalCash += tx.PaidAmount
		case "transfer":
			totalTransfer += tx.PaidAmount
		case "tempo":
			totalTempo += tx.PaidAmount
		default:
			totalCash += tx.PaidAmount
		}
	}

	return LPSResponse{
		Date:               dateStr,
		LPSType:            "MOTORIS",
		Salesman:           salesman,
		DeliveryBatches:    deliveryBatches,
		StoreVisits:        storeVisits,
		TotalGrandAmount:   totalGrand,
		TotalCash:          totalCash,
		TotalTransfer:      totalTransfer,
		TotalTempo:         totalTempo,
		TotalStoresVisited: len(storeVisits),
		TotalTransactions:  len(txRows),
	}
}

// ─── Helper functions ──────────────────────────────────────────────────────────

// normalizePaymentMethod menormalkan string metode pembayaran
func normalizePaymentMethod(method string) string {
	m := strings.ToLower(strings.TrimSpace(method))
	switch {
	case m == "":
		return "other"
	case strings.Contains(m, "cash") || strings.Contains(m, "tunai"):
		return "cash"
	case strings.Contains(m, "transfer") || strings.Contains(m, "qris") || strings.Contains(m, "va"):
		return "transfer"
	case strings.Contains(m, "tempo") || strings.Contains(m, "credit") || strings.Contains(m, "piutang"):
		return "tempo"
	default:
		return "other"
	}
}

// SaveDailyLPS menyimpan snapshot LPS untuk dibekukan secara permanen (Tutup Buku).
func (h *LPSHandler) SaveDailyLPS(c *gin.Context) {
	var req struct {
		Date         string      `json:"date"`
		EmployeeID   string      `json:"employee_id"`
		LPSType      string      `json:"lps_type"`
		SnapshotData LPSResponse `json:"snapshot_data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	dateParsed, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	lpsType := req.LPSType
	if lpsType == "" {
		lpsType = req.SnapshotData.LPSType
	}
	if lpsType == "" {
		lpsType = "DELIVERY"
	}

	snapshotBytes, err := json.Marshal(req.SnapshotData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize snapshot"})
		return
	}

	empUUID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee_id format"})
		return
	}

	report := domain.LPSReport{
		EmployeeID:    empUUID,
		ReportDate:    dateParsed,
		LPSType:       lpsType,
		TotalCash:     req.SnapshotData.TotalCash,
		TotalTransfer: req.SnapshotData.TotalTransfer,
		TotalTempo:    req.SnapshotData.TotalTempo,
		Status:        "DRAFT", // Default status until validated by Kasir
		SnapshotData:  string(snapshotBytes),
	}

	// Cek jika sudah ada (Upsert per employee, report_date, dan lps_type)
	var existing domain.LPSReport
	if err := h.db.Where("employee_id = ? AND report_date = ? AND lps_type = ?", report.EmployeeID, report.ReportDate, report.LPSType).First(&existing).Error; err == nil {
		existing.SnapshotData = report.SnapshotData
		existing.TotalCash = report.TotalCash
		existing.TotalTransfer = report.TotalTransfer
		existing.TotalTempo = report.TotalTempo
		if err := h.db.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update saved report"})
			return
		}
	} else {
		if err := h.db.Create(&report).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save report: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "LPS successfully saved and frozen"})
}

// ApproveDailyLPS performs bulk approval for all transactions of a salesman on a given date.
func (h *LPSHandler) ApproveDailyLPS(c *gin.Context) {
	var req struct {
		Date       string `json:"date"`
		EmployeeID string `json:"employee_id"`
		LPSType    string `json:"lps_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	dateParsed, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	empUUID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee_id format"})
		return
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction: " + tx.Error.Error()})
		return
	}

	// 1. Update LPS Report Status to APPROVED
	var report domain.LPSReport
	checkQuery := tx.Where("employee_id = ? AND report_date = ?", empUUID, dateParsed)
	if req.LPSType != "" {
		checkQuery = checkQuery.Where("lps_type = ?", req.LPSType)
	}
	if err := checkQuery.First(&report).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "LPS Report not found. Please save (Tutup Buku) first."})
		return
	}

	report.Status = "APPROVED"
	// Also update the JSON snapshot to reflect new status
	var snapshot LPSResponse
	if err := json.Unmarshal([]byte(report.SnapshotData), &snapshot); err == nil {
		snapshot.Status = "APPROVED"
		if updatedBytes, err := json.Marshal(snapshot); err == nil {
			report.SnapshotData = string(updatedBytes)
		}
	}
	if err := tx.Save(&report).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve LPS report: " + err.Error()})
		return
	}

	// 2. Bulk Approve Direct Sales Transactions (sales_transactions)
	if err := tx.Exec(`
		UPDATE sales_transactions 
		SET status = 'VERIFIED',
		    verified_at = COALESCE(verified_at, NOW()),
		    updated_at = NOW()
		WHERE employee_id = ? 
		  AND DATE(transaction_date) = ?
	`, empUUID, req.Date).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk approve sales transactions: " + err.Error()})
		return
	}

	// 3. Bulk Approve Sales Payments (sales_payments)
	if err := tx.Exec(`
		UPDATE sales_payments
		SET payment_status = 'SUCCESS',
		    updated_at = NOW()
		WHERE employee_id = ?
		  AND DATE(payment_date AT TIME ZONE 'Asia/Jakarta') = ?
	`, empUUID, req.Date).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk approve sales payments: " + err.Error()})
		return
	}

	// 4. Bulk Approve SO Transactions (sales_orders)
	if err := tx.Exec(`
		UPDATE sales_orders 
		SET status = 'PAID',
		    payment_status = 'PAID',
		    payment_collected_amount = total_amount,
		    updated_at = NOW()
		WHERE employee_id = ? 
		  AND DATE(created_at AT TIME ZONE 'Asia/Jakarta') = ?
	`, empUUID, req.Date).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk approve sales orders: " + err.Error()})
		return
	}

	tx.Commit()

	// 5. Recalculate Salesman KPI Omzet
	if h.salesUsecase != nil {
		_ = h.salesUsecase.CalculateKPI(empUUID, int(dateParsed.Month()), dateParsed.Year())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bulk approval successful! All related transactions are now VERIFIED and PAID."})
}
