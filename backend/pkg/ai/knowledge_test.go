package ai

import (
	"strings"
	"testing"
)

func TestPublicHelpScope(t *testing.T) {
	tests := []struct {
		question string
		allowed  bool
	}{
		{"Apa saja kategori produk perusahaan contoh?", true},
		{"Bagaimana cara input data di Super App?", true},
		{"Cara input sales order di aplikasi", true},
		{"Jumlah karyawan perusahaan contoh berapa?", false},
		{"Produk perusahaan contoh dan jumlah karyawan berapa?", false},
		{"Cara melihat data karyawan perusahaan contoh", false},
		{"Siapa pemilik perusahaan contoh?", false},
		{"Cek stok produk perusahaan contoh", false},
		{"Cara akses database perusahaan contoh?", false},
		{"Apa kabar?", false},
	}
	for _, tt := range tests {
		if got := IsPublicHelpQuestion(tt.question); got != tt.allowed {
			t.Errorf("IsPublicHelpQuestion(%q) = %v, want %v", tt.question, got, tt.allowed)
		}
	}
}

func TestBuildSystemPromptIgnoresPageContext(t *testing.T) {
	prompt := BuildSystemPrompt("Gaji Rahasia: 123")
	if strings.Contains(prompt, "123") || strings.Contains(prompt, "Gaji Rahasia") {
		t.Fatal("page context must not reach the AI provider")
	}
	if !strings.Contains(prompt, "fitur HRIS") || !strings.Contains(prompt, "Karsa HR") {
		t.Fatal("public help guidance missing")
	}
}
