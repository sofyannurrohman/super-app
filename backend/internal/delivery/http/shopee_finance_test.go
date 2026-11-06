package http

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"example.com/hris/backend/internal/domain"
)

type shopeeFinanceRoundTripFunc func(*http.Request) (*http.Response, error)

func (f shopeeFinanceRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func shopeeFinanceTestClient(handler func(*http.Request) string) *ShopeeClient {
	return &ShopeeClient{PartnerID: "123", PartnerKey: "secret", BaseURL: "https://example.invalid", HTTPClient: &http.Client{
		Transport: shopeeFinanceRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(handler(r)))}, nil
		}),
	}}
}

func TestShopeePaymentDetailUsesSignedShopRequestAndFeeBreakdown(t *testing.T) {
	client := shopeeFinanceTestClient(func(r *http.Request) string {
		if r.URL.Path != "/api/v2/payment/get_escrow_detail" || r.URL.Query().Get("order_sn") != "261018VMX02NBM" {
			t.Errorf("unexpected request: %s", r.URL.String())
		}
		if r.URL.Query().Get("partner_id") != "123" || r.URL.Query().Get("shop_id") != "456" || r.URL.Query().Get("access_token") != "access" {
			t.Error("shop credentials missing")
		}
		ts := r.URL.Query().Get("timestamp")
		if ts == "" {
			t.Error("timestamp missing")
		}
		return `{"error":"","response":{"order_sn":"261018VMX02NBM","order_income":{"escrow_amount":10000,"escrow_amount_after_adjustment":9500,"commission_fee":1000,"service_fee":500,"seller_transaction_fee":200,"campaign_fee":100,"order_ams_commission_fee":300,"actual_shipping_fee":800}}}`
	})
	token := &domain.MarketplaceToken{ShopID: "456", AccessToken: "access"}
	detail, err := client.fetchEscrowDetail(token, "261018VMX02NBM")
	if err != nil {
		t.Fatal(err)
	}
	if got := shopeeFeeAmount(detail); got != 2100 {
		t.Fatalf("fee = %v, want 2100", got)
	}
	if got := shopeeNetAmount(detail); got != 9500 {
		t.Fatalf("net = %v, want 9500", got)
	}
}

func TestShopeePaymentAPIErrorIsNotSilent(t *testing.T) {
	client := shopeeFinanceTestClient(func(*http.Request) string {
		return `{"error":"error_auth","message":"No permission to current api"}`
	})
	_, err := client.fetchEscrowDetail(&domain.MarketplaceToken{ShopID: "456", AccessToken: "access"}, "261018VMX02NBM")
	if err == nil || !strings.Contains(err.Error(), "error_auth") {
		t.Fatalf("expected Shopee permission error, got %v", err)
	}
}

func TestShopeeEscrowListRequiresReleaseData(t *testing.T) {
	client := shopeeFinanceTestClient(func(r *http.Request) string {
		if r.URL.Query().Get("release_time_from") != "100" || r.URL.Query().Get("page_no") != "2" {
			t.Errorf("unexpected list query: %s", r.URL.RawQuery)
		}
		return `{"error":"","response":{"escrow_list":[{"order_sn":"A","escrow_release_time":123,"escrow_amount":9500}],"more":false}}`
	})
	items, more, err := client.fetchEscrowListPage(&domain.MarketplaceToken{ShopID: "456", AccessToken: "access"}, 100, 200, 2)
	if err != nil || more || len(items) != 1 || items[0].EscrowReleaseTime != 123 {
		t.Fatalf("unexpected list result: items=%+v more=%v err=%v", items, more, err)
	}
}
