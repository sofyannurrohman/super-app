package usecase

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
)

func TestAttendanceSummaryCalculation(t *testing.T) {
	// Setup test employee
	empID := uuid.New()

	// Setup dates: 26 working days (Mon-Sat for ~4.5 weeks)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, loc)
	endDate := time.Date(2026, 1, 31, 23, 59, 59, 0, loc)

	var dates []time.Time
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}

	type LogKey struct {
		EmployeeID uuid.UUID
		DateStr    string
	}
	type LeaveKey struct {
		EmployeeID uuid.UUID
		DateStr    string
	}

	logMap := make(map[LogKey]domain.AttendanceLog)
	leaveMap := make(map[LeaveKey]string)

	// Skenario:
	// Total hari kerja bukan hari minggu: 26 hari (misal tanggal-tanggal selain Minggu)
	// - 24 hari hadir normal (ada clock-in, tidak ada leave)
	// - 2 hari cuti setengah hari (ada clock-in, ada cuti disetujui "Cuti Tahunan")
	workdayCount := 0
	for _, d := range dates {
		if d.Weekday() == time.Sunday {
			continue
		}
		workdayCount++
		dKey := d.Format("2006-01-02")
		clockIn := d.Add(8 * time.Hour)
		clockOut := d.Add(17 * time.Hour)

		if workdayCount <= 2 {
			// 2 hari cuti setengah hari (hadir clock in + cuti disetujui)
			logMap[LogKey{EmployeeID: empID, DateStr: dKey}] = domain.AttendanceLog{
				EmployeeID:   empID,
				ClockInTime:  &clockIn,
				ClockOutTime: &clockOut,
				Status:       "ON_TIME",
			}
			leaveMap[LeaveKey{EmployeeID: empID, DateStr: dKey}] = "Cuti Tahunan"
		} else {
			// Normal present untuk sisa semua hari kerja (misal 25 hari normal present)
			logMap[LogKey{EmployeeID: empID, DateStr: dKey}] = domain.AttendanceLog{
				EmployeeID:   empID,
				ClockInTime:  &clockIn,
				ClockOutTime: &clockOut,
				Status:       "ON_TIME",
			}
		}
	}

	// Hitung summary
	type EmployeeSummaryCounters struct {
		Hadir float64
		Dinas int
		Telat int
		Sakit float64
		Cuti  float64
		Izin  float64
		Alfa  int
	}

	summary := &EmployeeSummaryCounters{}
	for _, d := range dates {
		dKey := d.Format("2006-01-02")
		lKey := LogKey{EmployeeID: empID, DateStr: dKey}
		lvKey := LeaveKey{EmployeeID: empID, DateStr: dKey}
		isWeekend := d.Weekday() == time.Sunday

		log, hasLog := logMap[lKey]
		lvType, hasLeave := leaveMap[lvKey]

		isHadirOrTelat := hasLog && (log.Status == "ON_TIME" || log.Status == "PRESENT" || log.Status == "LATE")

		if hasLog && (log.IsOutOfTown || log.Status == "BUSINESS_TRIP") {
			summary.Dinas++
			summary.Hadir += 1.0
			if log.Status == "LATE" {
				summary.Telat++
			}
		} else if isHadirOrTelat {
			if hasLeave {
				summary.Hadir += 0.5
				lvTypeLower := strings.ToLower(lvType)
				if strings.Contains(lvTypeLower, "sakit") {
					summary.Sakit += 0.5
				} else if strings.Contains(lvTypeLower, "cuti") {
					summary.Cuti += 0.5
				} else {
					summary.Izin += 0.5
				}
			} else {
				summary.Hadir += 1.0
			}
			if log.Status == "LATE" {
				summary.Telat++
			}
		} else if hasLeave {
			lvTypeLower := strings.ToLower(lvType)
			if strings.Contains(lvTypeLower, "sakit") {
				summary.Sakit += 1.0
			} else if strings.Contains(lvTypeLower, "cuti") {
				summary.Cuti += 1.0
			} else {
				summary.Izin += 1.0
			}
		} else if hasLog && (log.Status == "ABSENT" || log.Status == "ALFA") {
			summary.Alfa++
		} else if !isWeekend {
			summary.Alfa++
		}
	}

	// Verifikasi hasil
	expectedHadir := float64(workdayCount-2) + 1.0 // (workdays - 2) + 0.5 + 0.5
	expectedCuti := 1.0                            // 0.5 + 0.5
	expectedAlfa := 0

	if summary.Hadir != expectedHadir {
		t.Errorf("Expected Hadir = %v, got %v", expectedHadir, summary.Hadir)
	}
	if summary.Cuti != expectedCuti {
		t.Errorf("Expected Cuti = %v, got %v", expectedCuti, summary.Cuti)
	}
	if summary.Alfa != expectedAlfa {
		t.Errorf("Expected Alfa = %v, got %v", expectedAlfa, summary.Alfa)
	}

	expectedTotalPaid := expectedHadir + expectedCuti + summary.Sakit
	totalPaid := summary.Hadir + summary.Cuti + summary.Sakit
	if totalPaid != expectedTotalPaid {
		t.Errorf("Expected Total Paid Days = %v, got %v", expectedTotalPaid, totalPaid)
	}

	// Helper formatting verification
	formatCount := func(v float64) interface{} {
		if v == math.Floor(v) {
			return int(v)
		}
		return v
	}

	if formatCount(summary.Hadir) != int(expectedHadir) {
		t.Errorf("Expected formatted Hadir = %v (int), got %v", int(expectedHadir), formatCount(summary.Hadir))
	}
	if formatCount(25.5) != 25.5 {
		t.Errorf("Expected formatted 25.5 = 25.5 (float), got %v", formatCount(25.5))
	}
}

func TestIzinSummaryAndComment(t *testing.T) {
	type EmployeeSummaryCounters struct {
		Hadir             float64
		Dinas             int
		Telat             int
		Sakit             float64
		Cuti              float64
		Izin              float64
		Alfa              int
		IzinHalfBreakdown map[string]int
		IzinFullBreakdown map[string]int
	}

	formatCount := func(v float64) interface{} {
		if v == math.Floor(v) {
			return int(v)
		}
		return v
	}

	buildIzinComment := func(s *EmployeeSummaryCounters) string {
		if s == nil || s.Izin <= 0 {
			return ""
		}
		var lines []string
		lines = append(lines, "Rincian Izin:")
		for name, count := range s.IzinFullBreakdown {
			lines = append(lines, fmt.Sprintf("• %s: %d hari", name, count))
		}
		for name, count := range s.IzinHalfBreakdown {
			halfDays := float64(count) * 0.5
			lines = append(lines, fmt.Sprintf("• %s: %d kali (%v hari)", name, count, formatCount(halfDays)))
		}
		lines = append(lines, fmt.Sprintf("Total: %v hari", formatCount(s.Izin)))
		return strings.Join(lines, "\n")
	}

	// Case 1: 1x "Izin 1/2 Hari" (0.5) + 1x "Izin Lainnya" (1.0) -> Total = 1.5 hari
	summary1 := &EmployeeSummaryCounters{
		Izin:              1.5,
		IzinHalfBreakdown: map[string]int{"Izin 1/2 Hari": 1},
		IzinFullBreakdown: map[string]int{"Izin Lainnya": 1},
	}

	comment1 := buildIzinComment(summary1)
	expectedTotal1 := formatCount(summary1.Izin)
	if expectedTotal1 != 1.5 {
		t.Errorf("Expected total 1.5, got %v", expectedTotal1)
	}

	if !strings.Contains(comment1, "Izin Lainnya: 1 hari") {
		t.Errorf("Comment missing Izin Lainnya: %s", comment1)
	}
	if !strings.Contains(comment1, "Izin 1/2 Hari: 1 kali (0.5 hari)") {
		t.Errorf("Comment missing Izin 1/2 Hari: %s", comment1)
	}
	if !strings.Contains(comment1, "Total: 1.5 hari") {
		t.Errorf("Comment missing Total: %s", comment1)
	}

	// Case 2: 2x "Izin 1/2 Hari" (1.0) + 1x "Izin Lainnya" (1.0) -> Total = 2 hari
	summary2 := &EmployeeSummaryCounters{
		Izin:              2.0,
		IzinHalfBreakdown: map[string]int{"Izin 1/2 Hari": 2},
		IzinFullBreakdown: map[string]int{"Izin Lainnya": 1},
	}

	comment2 := buildIzinComment(summary2)
	expectedTotal2 := formatCount(summary2.Izin)
	if expectedTotal2 != 2 {
		t.Errorf("Expected total 2, got %v", expectedTotal2)
	}

	if !strings.Contains(comment2, "Izin Lainnya: 1 hari") {
		t.Errorf("Comment missing Izin Lainnya: %s", comment2)
	}
	if !strings.Contains(comment2, "Izin 1/2 Hari: 2 kali (1 hari)") {
		t.Errorf("Comment missing Izin 1/2 Hari: %s", comment2)
	}
	if !strings.Contains(comment2, "Total: 2 hari") {
		t.Errorf("Comment missing Total: %s", comment2)
	}
}


