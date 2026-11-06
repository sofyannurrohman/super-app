package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/usecase"
)

type AccountingHandler struct {
	glPosting usecase.GLPostingUsecase
}

func NewAccountingHandler(glPosting usecase.GLPostingUsecase) *AccountingHandler {
	return &AccountingHandler{glPosting: glPosting}
}

func (h *AccountingHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/accounting")

	// Chart of Accounts
	g.GET("/accounts", h.GetChartOfAccounts)
	g.POST("/accounts", h.CreateAccount)
	g.PUT("/accounts/:id", h.UpdateAccount)
	g.POST("/accounts/seed", h.SeedDefaultAccounts)

	// Journal Entries
	g.GET("/journal-entries", h.GetJournalEntries)
	g.POST("/journal-entries", h.CreateManualJournalEntry)
	g.GET("/journal-entries/:id", h.GetJournalEntryByID)

	// Financial Reports
	g.GET("/reports/balance-sheet", h.GetBalanceSheet)
	g.GET("/reports/profit-loss", h.GetProfitLossReport)
	g.GET("/reports/cash-flow", h.GetCashFlowReport)
	g.GET("/reports/trial-balance", h.GetTrialBalance)

	// Journal Status Management
	g.POST("/journal-entries/:id/post", h.PostJournalEntry)
	g.POST("/journal-entries/:id/cancel", h.CancelJournalEntry)

	// Period Management
	g.POST("/periods/close", h.ClosePeriod)
}

// ─── Chart of Accounts ───────────────────────────────────────────────────────

// GetChartOfAccounts godoc
// @Summary Ambil daftar bagan akun (Chart of Accounts)
// @Tags Accounting
// @Produce json
// @Success 200 {array} domain.Account
// @Router /api/v1/accounting/accounts [get]
func (h *AccountingHandler) GetChartOfAccounts(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		if cidStr := c.Query("company_id"); cidStr != "" {
			if parsed, err := uuid.Parse(cidStr); err == nil {
				companyID = parsed
			}
		}
	}
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	accounts, err := h.glPosting.GetChartOfAccounts(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": accounts, "total": len(accounts)})
}

// CreateAccount godoc
// @Summary Buat akun baru dalam Chart of Accounts
// @Tags Accounting
// @Accept json
// @Produce json
// @Success 201 {object} domain.Account
// @Router /api/v1/accounting/accounts [post]
func (h *AccountingHandler) CreateAccount(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	var req usecase.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.glPosting.CreateAccount(companyID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": account, "message": "Akun berhasil dibuat"})
}

// UpdateAccount godoc
// @Summary Update akun dalam Chart of Accounts
// @Tags Accounting
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} domain.Account
// @Router /api/v1/accounting/accounts/{id} [put]
func (h *AccountingHandler) UpdateAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req usecase.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.glPosting.UpdateAccount(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": account, "message": "Akun berhasil diperbarui"})
}

// SeedDefaultAccounts godoc
// @Summary Inisialisasi bagan akun PSAK default untuk perusahaan
// @Tags Accounting
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/accounting/accounts/seed [post]
func (h *AccountingHandler) SeedDefaultAccounts(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	if err := h.glPosting.SeedDefaultAccounts(companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bagan Akun PSAK default berhasil diinisialisasi"})
}

// ─── Journal Entries ──────────────────────────────────────────────────────────

// GetJournalEntries godoc
// @Summary Ambil daftar jurnal umum
// @Tags Accounting
// @Produce json
// @Param month query int false "Bulan (1-12)"
// @Param year query int false "Tahun"
// @Param page query int false "Halaman" default(1)
// @Param limit query int false "Jumlah per halaman" default(50)
// @Success 200 {array} domain.JournalEntry
// @Router /api/v1/accounting/journal-entries [get]
func (h *AccountingHandler) GetJournalEntries(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	entries, total, err := h.glPosting.GetJournalEntries(companyID, month, year, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// CreateManualJournalEntry godoc
// @Summary Buat jurnal manual (double-entry)
// @Tags Accounting
// @Accept json
// @Produce json
// @Success 201 {object} domain.JournalEntry
// @Router /api/v1/accounting/journal-entries [post]
func (h *AccountingHandler) CreateManualJournalEntry(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	// Ambil employee/user ID dari context (untuk created_by)
	userID := getUserIDFromContext(c)

	var req domain.CreateJournalEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.EntryDate.IsZero() {
		req.EntryDate = time.Now()
	}

	entry, err := h.glPosting.CreateManualJournalEntry(companyID, userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": entry, "message": "Jurnal manual berhasil dibuat"})
}

// GetJournalEntryByID godoc
// @Summary Ambil detail jurnal berdasarkan ID
// @Tags Accounting
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Success 200 {object} domain.JournalEntry
// @Router /api/v1/accounting/journal-entries/{id} [get]
func (h *AccountingHandler) GetJournalEntryByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	entry, err := h.glPosting.GetJournalEntryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jurnal tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entry})
}

// ─── Financial Reports ────────────────────────────────────────────────────────

// GetBalanceSheet godoc
// @Summary Laporan Neraca (Balance Sheet) PSAK
// @Tags Accounting
// @Produce json
// @Param year query int true "Tahun"
// @Param month query int true "Bulan (1-12)"
// @Success 200 {object} domain.BalanceSheetReport
// @Router /api/v1/accounting/reports/balance-sheet [get]
func (h *AccountingHandler) GetBalanceSheet(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year == 0 {
		year = time.Now().Year()
	}
	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month == 0 {
		month = int(time.Now().Month())
	}

	report, err := h.glPosting.GetBalanceSheet(companyID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

// GetCashFlowReport godoc
// @Summary Laporan Arus Kas (Cash Flow Statement) PSAK 2
// @Tags Accounting
// @Produce json
// @Param month query int true "Bulan (1-12)"
// @Param year query int true "Tahun"
// @Success 200 {object} domain.CashFlowReport
// @Router /api/v1/accounting/reports/cash-flow [get]
func (h *AccountingHandler) GetCashFlowReport(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month == 0 {
		month = int(time.Now().Month())
	}
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year == 0 {
		year = time.Now().Year()
	}

	report, err := h.glPosting.GetCashFlowReport(companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

// GetTrialBalance godoc
// @Summary Neraca Saldo (Trial Balance)
// @Tags Accounting
// @Produce json
// @Param month query int true "Bulan (1-12)"
// @Param year query int true "Tahun"
// @Success 200 {array} domain.TrialBalanceLine
// @Router /api/v1/accounting/reports/trial-balance [get]
func (h *AccountingHandler) GetTrialBalance(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month == 0 {
		month = int(time.Now().Month())
	}
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year == 0 {
		year = time.Now().Year()
	}

	lines, err := h.glPosting.GetTrialBalance(companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lines, "total": len(lines)})
}

// ─── Period Management ────────────────────────────────────────────────────────

// ClosePeriod godoc
// @Summary Tutup periode akuntansi
// @Tags Accounting
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/accounting/periods/close [post]
func (h *AccountingHandler) ClosePeriod(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}

	var body struct {
		Year  int `json:"year"`
		Month int `json:"month"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Year == 0 || body.Month == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year dan month wajib diisi"})
		return
	}

	closedBy := getUserIDFromContext(c)

	if err := h.glPosting.ClosePeriod(companyID, body.Year, body.Month, closedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Periode akuntansi berhasil ditutup",
		"year":    body.Year,
		"month":   body.Month,
	})
}

// ─── Profit/Loss Report ───────────────────────────────────────────────────────

// GetProfitLossReport godoc
// @Summary Laporan Laba Rugi
// @Tags Accounting
// @Produce json
// @Param month query int false "Bulan (1-12)"
// @Param year query int false "Tahun"
// @Success 200 {object} domain.ProfitLossReport
// @Router /api/v1/accounting/reports/profit-loss [get]
func (h *AccountingHandler) GetProfitLossReport(c *gin.Context) {
	companyID := getCompanyIDFromContext(c)
	if companyID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company_id tidak ditemukan"})
		return
	}
	now := time.Now()
	month := parseIntQuery(c, "month", int(now.Month()))
	year := parseIntQuery(c, "year", now.Year())

	report, err := h.glPosting.GetProfitLossReport(companyID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}

// ─── Journal Status Management ────────────────────────────────────────────────

// PostJournalEntry godoc
// @Summary Posting jurnal DRAFT menjadi POSTED
// @Tags Accounting
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Success 200 {object} gin.H
// @Router /api/v1/accounting/journal-entries/{id}/post [post]
func (h *AccountingHandler) PostJournalEntry(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID jurnal tidak valid"})
		return
	}
	if err := h.glPosting.PostJournalEntry(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Jurnal berhasil diposting"})
}

// CancelJournalEntry godoc
// @Summary Batalkan jurnal DRAFT
// @Tags Accounting
// @Produce json
// @Param id path string true "Journal Entry ID"
// @Success 200 {object} gin.H
// @Router /api/v1/accounting/journal-entries/{id}/cancel [post]
func (h *AccountingHandler) CancelJournalEntry(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID jurnal tidak valid"})
		return
	}
	if err := h.glPosting.CancelJournalEntry(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Jurnal berhasil dibatalkan"})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// getCompanyIDFromContext mengambil company_id dari context middleware.
func getCompanyIDFromContext(c *gin.Context) uuid.UUID {
	if v, exists := c.Get("company_id"); exists {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
		if s, ok := v.(string); ok {
			id, _ := uuid.Parse(s)
			return id
		}
	}
	return uuid.Nil
}

// getUserIDFromContext mengambil user_id dari JWT context.
func getUserIDFromContext(c *gin.Context) uuid.UUID {
	if v, exists := c.Get("user_id"); exists {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
		if s, ok := v.(string); ok {
			id, _ := uuid.Parse(s)
			return id
		}
	}
	return uuid.Nil
}

// parseIntQuery mengambil query param sebagai int dengan default fallback.
func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	s := c.Query(key)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}
