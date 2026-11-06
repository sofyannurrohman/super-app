package http

import "testing"

func TestClassifyUnsettledKeepsUnknownReasonVisible(t *testing.T) {
	tests := []struct {
		reason    string
		delivered int64
		want      string
	}{
		{"AFTER_SALES_REFUND", 0, "AFTER_SALES"},
		{"IN_TRANSIT", 0, "AWAITING_DELIVERY"},
		{"SETTLEMENT_PROCESSING", 123, "PROCESSING"},
		{"", 123, "PROCESSING"},
		{"UNKNOWN_TIKTOK_REASON", 0, "UNCLASSIFIED"},
	}
	for _, tt := range tests {
		if got := classifyUnsettled(tt.reason, tt.delivered); got != tt.want {
			t.Errorf("classifyUnsettled(%q, %d) = %q; want %q", tt.reason, tt.delivered, got, tt.want)
		}
	}
}

func TestParseIDRAmountRejectsMissingValue(t *testing.T) {
	if _, err := parseIDRAmount(""); err == nil {
		t.Fatal("missing estimated amount must fail rather than silently produce zero")
	}
	if got, err := parseIDRAmount("14935044"); err != nil || got != 14935044 {
		t.Fatalf("got %d, %v; want 14935044", got, err)
	}
}
