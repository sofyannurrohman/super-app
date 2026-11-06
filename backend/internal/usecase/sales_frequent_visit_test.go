package usecase

import (
	"testing"
	"time"
)

func TestFrequentVisitCalendarDayCalculation(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")

	calculateDaysDiff := func(currentTime, lastVisitTime time.Time) int {
		currYear, currMonth, currDay := currentTime.Date()
		currDate := time.Date(currYear, currMonth, currDay, 0, 0, 0, 0, loc)

		lastYear, lastMonth, lastDay := lastVisitTime.Date()
		lastDate := time.Date(lastYear, lastMonth, lastDay, 0, 0, 0, 0, loc)

		return int(currDate.Sub(lastDate) / (24 * time.Hour))
	}

	tests := []struct {
		name              string
		currentTime       time.Time
		lastVisitTime     time.Time
		expectedDaysDiff  int
		expectedIsFrequent bool // < 2 hari kalender
	}{
		{
			name:              "Same day visit (morning to afternoon)",
			currentTime:       time.Date(2026, 10, 15, 14, 0, 0, 0, loc),
			lastVisitTime:     time.Date(2026, 10, 15, 9, 0, 0, 0, loc),
			expectedDaysDiff:  0,
			expectedIsFrequent: true,
		},
		{
			name:              "Consecutive day visit (H-1 to today)",
			currentTime:       time.Date(2026, 10, 15, 10, 0, 0, 0, loc),
			lastVisitTime:     time.Date(2026, 10, 14, 15, 0, 0, 0, loc),
			expectedDaysDiff:  1,
			expectedIsFrequent: true,
		},
		{
			name:              "Two days apart (Tuesday morning to Thursday morning)",
			currentTime:       time.Date(2026, 10, 15, 10, 39, 0, 0, loc),
			lastVisitTime:     time.Date(2026, 10, 13, 10, 29, 0, 0, loc),
			expectedDaysDiff:  2,
			expectedIsFrequent: false,
		},
		{
			name:              "Two days apart (Tuesday afternoon to Thursday morning, 43 hours apart)",
			currentTime:       time.Date(2026, 10, 15, 10, 0, 0, 0, loc),
			lastVisitTime:     time.Date(2026, 10, 13, 15, 0, 0, 0, loc),
			expectedDaysDiff:  2,
			expectedIsFrequent: false,
		},
		{
			name:              "Three days apart (Monday to Thursday)",
			currentTime:       time.Date(2026, 10, 15, 10, 0, 0, 0, loc),
			lastVisitTime:     time.Date(2026, 10, 12, 10, 0, 0, 0, loc),
			expectedDaysDiff:  3,
			expectedIsFrequent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daysDiff := calculateDaysDiff(tt.currentTime, tt.lastVisitTime)
			if daysDiff != tt.expectedDaysDiff {
				t.Errorf("daysDiff = %v, expected %v", daysDiff, tt.expectedDaysDiff)
			}
			isFrequent := daysDiff < 2
			if isFrequent != tt.expectedIsFrequent {
				t.Errorf("isFrequent = %v, expected %v", isFrequent, tt.expectedIsFrequent)
			}
		})
	}
}
