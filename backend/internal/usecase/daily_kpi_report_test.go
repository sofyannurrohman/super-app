package usecase

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDailyKpiReportRepository
type MockDailyKpiReportRepository struct {
	mock.Mock
}

func (m *MockDailyKpiReportRepository) CreateReport(report *domain.DailyKpiReport) error {
	args := m.Called(report)
	return args.Error(0)
}

func (m *MockDailyKpiReportRepository) UpdateReport(report *domain.DailyKpiReport) error {
	args := m.Called(report)
	return args.Error(0)
}

func (m *MockDailyKpiReportRepository) DeleteReport(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDailyKpiReportRepository) GetReportByID(id uuid.UUID) (*domain.DailyKpiReport, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DailyKpiReport), args.Error(1)
}

func (m *MockDailyKpiReportRepository) GetReports(filter repository.DailyKpiFilter) ([]domain.DailyKpiReport, int64, error) {
	args := m.Called(filter)
	return args.Get(0).([]domain.DailyKpiReport), args.Get(1).(int64), args.Error(2)
}

func (m *MockDailyKpiReportRepository) GetSubordinateActualData(reporterID uuid.UUID, branchID *uuid.UUID, date time.Time) ([]repository.SubordinateActualKPI, error) {
	args := m.Called(reporterID, branchID, date)
	return args.Get(0).([]repository.SubordinateActualKPI), args.Error(1)
}

func TestDailyKpiReport_ExportExcel(t *testing.T) {
	mockRepo := new(MockDailyKpiReportRepository)
	reportID := uuid.New()
	repEmpID := uuid.New()
	revEmpID := uuid.New()

	testReport := &domain.DailyKpiReport{
		ID:                 reportID,
		ReportDate:         time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC),
		ReporterEmployeeID: repEmpID,
		ReviewerEmployeeID: &revEmpID,
		ReporterRole:       "SPV",
		AreaName:           "Denpasar Timur",
		Status:             "SUBMITTED",
		ReporterEmployee: &domain.Employee{
			FirstName: "Dodik",
			LastName:  "Contoh",
		},
		ReviewerEmployee: &domain.Employee{
			FirstName: "Budi",
			LastName:  "ASM",
		},
		SummaryTargetCall:              30,
		SummaryRealizationCall:         28,
		SummaryAchievementCallPercent:  93.3,
		SummaryTargetEC:                10,
		SummaryRealizationEC:           9,
		SummaryAchievementECPercent:   90.0,
		SummaryTargetSales:             5000000,
		SummaryRealizationSales:        4500000,
		SummaryAchievementSalesPercent: 90.0,
		TomorrowPriority:               "Fokus penetrasi outlet rute Sesetan & Sanur",
		Notes:                          "Cuaca hujan lebat di siang hari menyebabkan beberapa toko tutup awal",
		Items: []domain.DailyKpiReportItem{
			{
				EmployeeName:           "Salesman A",
				AreaRoute:              "Rute Sanur",
				TargetCall:             15,
				RealizationCall:        14,
				AchievementCallPercent: 93.3,
				TargetEC:               5,
				RealizationEC:          5,
				AchievementECPercent:   100.0,
				TargetSales:            2500000,
				RealizationSales:       2600000,
				AchievementSalesPercent: 104.0,
				DiffSales:              100000,
			},
			{
				EmployeeName:           "Salesman B",
				AreaRoute:              "Rute Sesetan",
				TargetCall:             15,
				RealizationCall:        14,
				AchievementCallPercent: 93.3,
				TargetEC:               5,
				RealizationEC:          4,
				AchievementECPercent:   80.0,
				TargetSales:            2500000,
				RealizationSales:       1900000,
				AchievementSalesPercent: 76.0,
				DiffSales:              -600000,
			},
		},
		ActionPlans: []domain.DailyKpiReportActionPlan{
			{
				EmployeeName:     "Salesman B",
				ReasonUnachieved: "4 outlet tidak order karena stok masih penuh",
				FieldObstacle:    "Stok produk fast moving kosong di depo",
				ActionFollowUp:   "SPV koordinasi percepatan restock gudang",
				PicName:          "Dodik Contoh",
				Status:           "OPEN",
				SupportNeeded:    "Bantuan alokasi darurat dari cabang utama",
			},
		},
	}

	mockRepo.On("GetReportByID", reportID).Return(testReport, nil)

	u := &dailyKpiReportUsecase{
		repo: mockRepo,
	}

	fileBytes, filename, err := u.ExportExcel(reportID)
	assert.NoError(t, err)
	assert.NotEmpty(t, fileBytes)
	assert.Contains(t, filename, "2026-10-23")
	assert.Contains(t, filename, ".xlsx")
	assert.Greater(t, len(fileBytes), 1000)
}
