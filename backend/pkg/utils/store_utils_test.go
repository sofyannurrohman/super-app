package utils

import (
	"testing"
)

func TestSanitizePhoneNumber(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"0812-3456-7890", "6281234567890"},
		{"+62 812 3456 7890", "6281234567890"},
		{"6281234567890", "6281234567890"},
		{"81234567890", "6281234567890"},
		{"0812 999", ""}, // too short (< 8)
		{"", ""},
	}

	for _, tc := range cases {
		got := SanitizePhoneNumber(tc.input)
		if got != tc.expected {
			t.Errorf("SanitizePhoneNumber(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestCalculateSimilarity(t *testing.T) {
	// Normalization removes prefixes like TB. / Toko / UD
	sim1 := CalculateSimilarity("TB. Berkah Jaya", "Toko Berkah Jaya")
	if sim1 != 1.0 {
		t.Errorf("Expected similarity 1.0 for 'TB. Berkah Jaya' and 'Toko Berkah Jaya', got %f", sim1)
	}

	// Minor typo / spelling variations
	sim2 := CalculateSimilarity("Sumber Rejeki", "Sumber Rezeki")
	if sim2 < 0.80 {
		t.Errorf("Expected similarity >= 0.80 for 'Sumber Rejeki' and 'Sumber Rezeki', got %f", sim2)
	}

	// Completely different names
	sim3 := CalculateSimilarity("Toko Maju Bersama", "Depo Bangunan Sentosa")
	if sim3 > 0.40 {
		t.Errorf("Expected low similarity for different store names, got %f", sim3)
	}
}
