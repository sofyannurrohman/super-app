package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateReturnNumber(t *testing.T) {
	date := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	branchID := uuid.New()

	no1 := GenerateReturnNumber(branchID, date, 0)
	if no1 != "RET-20260930-0001" {
		t.Errorf("expected RET-20260930-0001, got %s", no1)
	}

	no2 := GenerateReturnNumber(branchID, date, 1)
	if no2 != "RET-20260930-0002" {
		t.Errorf("expected RET-20260930-0002, got %s", no2)
	}
}

func TestGenerateExchangeNumber(t *testing.T) {
	date := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	branchID := uuid.New()

	no1 := GenerateExchangeNumber(branchID, date, 0)
	if no1 != "EXC-20260930-0001" {
		t.Errorf("expected EXC-20260930-0001, got %s", no1)
	}

	no2 := GenerateExchangeNumber(branchID, date, 1)
	if no2 != "EXC-20260930-0002" {
		t.Errorf("expected EXC-20260930-0002, got %s", no2)
	}
}

func TestSequenceParsingLogic(t *testing.T) {
	date := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	prefix := fmt.Sprintf("RET-%s-", date.Format("20060102"))

	tests := []struct {
		lastNo   string
		expected string
	}{
		{"", "RET-20260930-0001"},
		{"RET-20260930-0001", "RET-20260930-0002"},
		{"RET-20260930-0009", "RET-20260930-0010"},
		{"RET-20260930-0099", "RET-20260930-0100"},
		{"RET-20260930-0999", "RET-20260930-1000"},
	}

	for _, tt := range tests {
		seq := 1
		if tt.lastNo != "" {
			var prevSeq int
			if _, err := fmt.Sscanf(tt.lastNo, prefix+"%d", &prevSeq); err == nil {
				seq = prevSeq + 1
			}
		}
		result := fmt.Sprintf("%s%04d", prefix, seq)
		if result != tt.expected {
			t.Errorf("for lastNo %s, expected %s, got %s", tt.lastNo, tt.expected, result)
		}
	}
}
