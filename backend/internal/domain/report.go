package domain

import "time"

// DailyReportData berisi semua data agregat untuk laporan harian managerial.
type DailyReportData struct {
	Date         time.Time
	ReportPeriod string // contoh: "Selasa, 08 Juli 2026"

	// ── HRIS Metrics ────────────────────────────────────────────────────────────
	TotalEmployees       int
	PresentToday         int
	LateToday            int
	OnLeaveToday         int
	AbsentToday          int
	PendingLeaveReqs     int
	PendingOvertimeReqs  int
	PendingReimbursement int

	// ── Sales Metrics ───────────────────────────────────────────────────────────
	TotalSalesToday     float64
	TotalTransactions   int
	TopSalesmanName     string
	TopSalesmanAmount   float64
	PendingOrders       int
	CompletedDeliveries int
	TotalDeliveries     int
}

// CityRevenueReportItem merepresentasikan baris analitik omzet & penetrasi per Kota/Kabupaten
type CityRevenueReportItem struct {
	City              string  `json:"city"`
	Province          string  `json:"province"`
	Status            string  `json:"status"` // "PENETRATED", "UNPENETRATED"
	TotalStores       int     `json:"total_stores"`
	ActiveStoresCount int     `json:"active_stores_count"`
	TotalQuantity     float64 `json:"total_quantity"`
	TotalRevenue      float64 `json:"total_revenue"`
	RevenueSharePct   float64 `json:"revenue_share_pct"`
	BranchName        string  `json:"branch_name"`
	TopSalesmanName   string  `json:"top_salesman_name"`
	TopProductName    string  `json:"top_product_name"`
}

// CityRevenueReportResponse adalah payload respon laporan omzet per kota
type CityRevenueReportResponse struct {
	StartDate               string                  `json:"start_date"`
	EndDate                 string                  `json:"end_date"`
	PeriodType              string                  `json:"period_type"`
	TotalCitiesMonitored    int                     `json:"total_monitored_cities"`
	PenetratedCitiesCount   int                     `json:"penetrated_count"`
	UnpenetratedCitiesCount int                     `json:"unpenetrated_count"`
	PenetrationRate         float64                 `json:"penetration_rate"`
	TotalRevenue            float64                 `json:"total_revenue"`
	TotalQuantity           float64                 `json:"total_quantity"`
	TopCityName             string                  `json:"top_city_name"`
	TopCityRevenue          float64                 `json:"top_city_revenue"`
	Cities                  []CityRevenueReportItem `json:"cities"`
}


