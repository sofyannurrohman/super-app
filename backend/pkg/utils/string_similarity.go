package utils

import (
	"math"
	"regexp"
	"strings"
)

var (
	storePrefixes = []string{
		"tb.", "tb ", "toko ", "ud.", "ud ", "pd.", "pd ", "cv.", "cv ",
		"pt.", "pt ", "warung ", "wr.", "wr ", "kios ", "bengkel ", "depo ",
	}
	punctRegex = regexp.MustCompile(`[^\w\s]`)
	multiSpaceRegex = regexp.MustCompile(`\s+`)
)

// NormalizeStoreName removes prefixes like TB, Toko, UD, punctuations, and trims extra spaces.
func NormalizeStoreName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))

	// Remove common prefixes
	for _, prefix := range storePrefixes {
		if strings.HasPrefix(s, prefix) {
			s = strings.TrimPrefix(s, prefix)
			break
		}
	}

	// Remove punctuation
	s = punctRegex.ReplaceAllString(s, " ")

	// Replace multiple spaces with single space
	s = multiSpaceRegex.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}

// CalculateLevenshtein calculates the Levenshtein distance between two strings.
func CalculateLevenshtein(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	d := make([][]int, n+1)
	for i := range d {
		d[i] = make([]int, m+1)
		d[i][0] = i
	}
	for j := 0; j <= m; j++ {
		d[0][j] = j
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			d[i][j] = int(math.Min(
				float64(d[i-1][j]+1),
				math.Min(
					float64(d[i][j-1]+1),
					float64(d[i-1][j-1]+cost),
				),
			))
		}
	}

	return d[n][m]
}

// CalculateSimilarity returns a similarity score between 0.0 and 1.0 using normalized Levenshtein.
func CalculateSimilarity(s1, s2 string) float64 {
	n1 := NormalizeStoreName(s1)
	n2 := NormalizeStoreName(s2)

	if n1 == n2 && n1 != "" {
		return 1.0
	}

	maxLen := math.Max(float64(len([]rune(n1))), float64(len([]rune(n2))))
	if maxLen == 0 {
		return 0.0
	}

	dist := float64(CalculateLevenshtein(n1, n2))
	similarity := 1.0 - (dist / maxLen)
	if similarity < 0 {
		return 0
	}
	return similarity
}
