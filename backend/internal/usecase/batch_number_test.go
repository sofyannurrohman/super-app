package usecase

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateInventoryBatch(t *testing.T) {
	at := time.Date(2026, 9, 21, 14, 5, 6, 0, time.UTC)
	got := generateInventoryBatch("rm", uuid.MustParse("12345678-1234-1234-1234-123456789abc"), at)
	want := "RM-20260921-12345678-140506"
	if got != want {
		t.Fatalf("generateInventoryBatch() = %q, want %q", got, want)
	}
	if strings.Contains(got, "DEFAULT") {
		t.Fatalf("generated batch must not use DEFAULT: %q", got)
	}
}
