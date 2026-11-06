package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestParseTikTokFeeAmount(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "negative platform deduction", raw: "-12500.50", want: 12500.50},
		{name: "positive platform fee", raw: "12500.50", want: 12500.50},
		{name: "surrounding whitespace", raw: " -2500 ", want: 2500},
		{name: "missing amount", raw: "", want: 0},
		{name: "invalid amount", raw: "NaN", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTikTokFeeAmount(tt.raw); got != tt.want {
				t.Fatalf("parseTikTokFeeAmount(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseOptionalTikTokFeeAmountDistinguishesMissingFromZero(t *testing.T) {
	for _, raw := range []string{"", " ", "invalid", "NaN"} {
		if amount, provided := parseOptionalTikTokFeeAmount(raw); provided || amount != 0 {
			t.Fatalf("fee %q should be unavailable, got amount=%v provided=%v", raw, amount, provided)
		}
	}
	if amount, provided := parseOptionalTikTokFeeAmount("0"); !provided || amount != 0 {
		t.Fatalf("explicit zero fee should be present, got amount=%v provided=%v", amount, provided)
	}
	if amount, provided := parseOptionalTikTokFeeAmount("-5163"); !provided || amount != 5163 {
		t.Fatalf("negative deduction should be present, got amount=%v provided=%v", amount, provided)
	}
}

func TestParseOptionalTikTokSettlementDistinguishesMissingFromZero(t *testing.T) {
	if _, provided := parseOptionalTikTokAmount(""); provided {
		t.Fatal("missing payout must not be treated as zero payout")
	}
	if amount, provided := parseOptionalTikTokAmount("0"); !provided || amount != 0 {
		t.Fatalf("explicit zero payout not preserved: amount=%v provided=%v", amount, provided)
	}
	if amount, provided := parseOptionalTikTokAmount("20805"); !provided || amount != 20805 {
		t.Fatalf("official payout not parsed: amount=%v provided=%v", amount, provided)
	}
}

func TestTikTokSettlementRequiresPaidStatementAndUsesPaymentTime(t *testing.T) {
	paidAt := time.Unix(1792713600, 0)
	statementAt := paidAt.Add(-24 * time.Hour)
	status, date, reason := classifyTikTokStatementPayment("PAID", &paidAt, &statementAt)
	if status != "SETTLED" || date == nil || !date.Equal(paidAt) || reason != "" {
		t.Fatalf("paid statement classified incorrectly: %s %v %s", status, date, reason)
	}
	status, date, reason = classifyTikTokStatementPayment("SETTLED", nil, &statementAt)
	if status != "SETTLED" || date == nil || !date.Equal(statementAt) || reason != "" {
		t.Fatalf("seller balance settlement classified incorrectly: %s %v %s", status, date, reason)
	}
	for _, input := range []string{"", "PROCESSING", "FAILED"} {
		status, date, _ = classifyTikTokStatementPayment(input, &paidAt, &statementAt)
		if status != "HOLD" || date != nil {
			t.Fatalf("unpaid statement %q classified as %s at %v", input, status, date)
		}
	}
}

func TestTikTokFinanceRetriesRateLimitedPage(t *testing.T) {
	requests, paced := 0, 0
	var delays []time.Duration
	result, _, err := retryTikTokFinanceRequest(func() { paced++ }, func(delay time.Duration) { delays = append(delays, delay) }, nil, func() (int, *http.Response, error) {
		requests++
		if requests == 1 {
			return 36009002, &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"7"}}}, nil
		}
		return 0, &http.Response{StatusCode: http.StatusOK}, nil
	}, func(code int) int { return code })
	if err != nil || result != 0 || requests != 2 || paced != 2 || len(delays) != 1 || delays[0] != 7*time.Second {
		t.Fatalf("rate limit retry failed: result=%d err=%v requests=%d paced=%d delays=%v", result, err, requests, paced, delays)
	}
}

func TestTikTokFinanceDoesNotRetryOtherErrors(t *testing.T) {
	requests := 0
	_, _, err := retryTikTokFinanceRequest(func() {}, func(time.Duration) { t.Fatal("unexpected retry") }, nil, func() (int, *http.Response, error) {
		requests++
		return 0, &http.Response{StatusCode: http.StatusUnauthorized}, fmt.Errorf("unauthorized")
	}, func(code int) int { return code })
	if requests != 1 || err == nil || err.Error() != "unauthorized" {
		t.Fatalf("unexpected non-rate-limit handling: requests=%d err=%v", requests, err)
	}
}

func TestTikTokFinanceStopsAfterBoundedRateLimitRetries(t *testing.T) {
	requests, waits := 0, 0
	_, _, err := retryTikTokFinanceRequest(func() {}, func(time.Duration) { waits++ }, nil, func() (int, *http.Response, error) {
		requests++
		return 36009002, &http.Response{StatusCode: http.StatusTooManyRequests}, nil
	}, func(code int) int { return code })
	if err == nil || requests != 7 || waits != 6 {
		t.Fatalf("rate limit retries were not bounded: requests=%d waits=%d err=%v", requests, waits, err)
	}
}

func TestTikTokFinanceReusesRunningJob(t *testing.T) {
	h := &MarketplaceHandler{financeSyncJobs: make(map[string]financeSyncStatus)}
	first, started := h.beginTikTokFinanceSync()
	if !started || first == "" {
		t.Fatal("first sync did not start")
	}
	second, started := h.beginTikTokFinanceSync()
	if started || second != first {
		t.Fatalf("duplicate sync started: first=%s second=%s started=%v", first, second, started)
	}
	h.setFinanceSyncStatus(first, financeSyncStatus{Status: "completed"})
	third, started := h.beginTikTokFinanceSync()
	if !started || third == first {
		t.Fatalf("new sync blocked after completion: third=%s started=%v", third, started)
	}
}

func TestFinanceSyncStatusReportsCompletion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &MarketplaceHandler{financeSyncJobs: make(map[string]financeSyncStatus)}
	h.setFinanceSyncStatus("sync-1", financeSyncStatus{Status: "completed", SyncedCount: 2, StatementsCount: 1})

	router := gin.New()
	router.GET("/status/:sync_id", h.GetTikTokFinanceSyncStatus)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/status/sync-1", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status endpoint returned %d: %s", response.Code, response.Body.String())
	}
	var result financeSyncStatus
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || result.SyncedCount != 2 || result.StatementsCount != 1 {
		t.Fatalf("unexpected sync status: %+v", result)
	}
}

func TestFinanceSyncStatusReportsProgress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &MarketplaceHandler{financeSyncJobs: make(map[string]financeSyncStatus)}
	h.setFinanceSyncStatus("sync-progress", financeSyncStatus{Status: "processing", Platform: "shopee", Phase: "Membaca escrow", ProcessedCount: 20, SyncedCount: 3})
	router := gin.New()
	router.GET("/status/:sync_id", h.GetTikTokFinanceSyncStatus)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/status/sync-progress", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status endpoint returned %d: %s", response.Code, response.Body.String())
	}
	var result financeSyncStatus
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "processing" || result.ProcessedCount != 20 || result.SyncedCount != 3 || result.Phase != "Membaca escrow" || result.UpdatedAt.IsZero() {
		t.Fatalf("progress missing from status: %+v", result)
	}
}
