package utils

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func StringToInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return val
}

func StringToFloat(s string, fallback float64) float64 {
	if s == "" {
		return fallback
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fallback
	}
	return val
}

// ParseUUIDs memecah string comma-separated UUID dan mengembalikan slice uuid.UUID yang valid.
func ParseUUIDs(s string) []uuid.UUID {
	if s == "" || s == "ALL" || s == "ALL_BRANCHES" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []uuid.UUID
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" || trimmed == "ALL" || trimmed == "ALL_BRANCHES" {
			continue
		}
		if id, err := uuid.Parse(trimmed); err == nil && id != uuid.Nil {
			result = append(result, id)
		}
	}
	return result
}

