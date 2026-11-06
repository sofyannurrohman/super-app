package http

import "testing"

func TestCampaignVideoLinks(t *testing.T) {
	for _, tc := range []struct {
		url  string
		want bool
	}{
		{"https://drive.google.com/file/d/abc_123/view?usp=sharing", true},
		{"https://drive.google.com/open?id=abc_123", true},
		{"https://evil.example/file/d/abc_123/view", false},
		{"http://drive.google.com/file/d/abc_123/view", false},
		{"https://drive.google.com/drive/folders/abc_123", false},
	} {
		if got := validDriveVideoURL(tc.url); got != tc.want {
			t.Errorf("validDriveVideoURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
	for _, tc := range []struct {
		url  string
		want bool
	}{
		{"https://shopee.co.id/product/123", true},
		{"javascript:alert(1)", false},
		{"http://example.com/product", false},
		{"https://user:pass@example.com/product", false},
	} {
		if got := validPurchaseURL(tc.url); got != tc.want {
			t.Errorf("validPurchaseURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}
