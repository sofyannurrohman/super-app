package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type EmployeeRequestFilter struct {
	Status        string
	DepartmentID  *uuid.UUID
	BranchID      *uuid.UUID
	RequesterID   *uuid.UUID
	Search        string
	StartDate     *time.Time
	EndDate       *time.Time
}

type EmployeeRequestSummaryStats struct {
	TotalRequests           int64 `json:"total_requests"`
	PendingRequests         int64 `json:"pending_requests"`
	WaitingHRApproval       int64 `json:"waiting_hr_approval"`
	WaitingDirectorApproval int64 `json:"waiting_director_approval"`
	ApprovedRequests        int64 `json:"approved_requests"`
	InProgressRequests      int64 `json:"in_progress_requests"`
	FulfilledRequests       int64 `json:"fulfilled_requests"`
	RejectedRequests        int64 `json:"rejected_requests"`
	TotalHeadcountNeeded    int64 `json:"total_headcount_needed"`
	TotalHeadcountFulfilled int64 `json:"total_headcount_fulfilled"`
}

type EmployeeRequestRepository interface {
	Create(req *domain.EmployeeRequest) error
	Update(req *domain.EmployeeRequest) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*domain.EmployeeRequest, error)
	GetAll(filter EmployeeRequestFilter, limit, offset int) ([]domain.EmployeeRequest, int64, error)
	GenerateRequestNumber() (string, error)
	GetSummaryStats(branchID *uuid.UUID, departmentID *uuid.UUID) (*EmployeeRequestSummaryStats, error)
	AddCandidate(candidate *domain.EmployeeRequestCandidate) error
	DeleteCandidate(id uuid.UUID) error
	GetCandidatesByRequestID(requestID uuid.UUID) ([]domain.EmployeeRequestCandidate, error)
}

type employeeRequestRepository struct {
	db *gorm.DB
}

func NewEmployeeRequestRepository(db *gorm.DB) EmployeeRequestRepository {
	return &employeeRequestRepository{db: db}
}

func (r *employeeRequestRepository) Create(req *domain.EmployeeRequest) error {
	return r.db.Create(req).Error
}

func (r *employeeRequestRepository) Update(req *domain.EmployeeRequest) error {
	return r.db.Save(req).Error
}

func (r *employeeRequestRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.EmployeeRequest{}, "id = ?", id).Error
}

func (r *employeeRequestRepository) GetByID(id uuid.UUID) (*domain.EmployeeRequest, error) {
	var req domain.EmployeeRequest
	err := r.db.Preload("Requester.JobPosition").
		Preload("Requester.Department").
		Preload("Approver.JobPosition").
		Preload("HRApprover.JobPosition").
		Preload("DirectorApprover.JobPosition").
		Preload("Department").
		Preload("Branch").
		Preload("JobPosition").
		Preload("Candidates", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ?", id).
		First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *employeeRequestRepository) GetAll(filter EmployeeRequestFilter, limit, offset int) ([]domain.EmployeeRequest, int64, error) {
	var list []domain.EmployeeRequest
	var total int64

	query := r.db.Model(&domain.EmployeeRequest{})

	if filter.Status != "" && filter.Status != "ALL" {
		if filter.Status == "WAITING_HR" {
			query = query.Where("status = 'PENDING' AND hr_status = 'PENDING'")
		} else if filter.Status == "WAITING_DIRECTOR" {
			query = query.Where("status = 'WAITING_DIRECTOR_APPROVAL' OR (hr_status = 'APPROVED' AND director_status = 'PENDING')")
		} else {
			query = query.Where("status = ?", filter.Status)
		}
	}
	if filter.DepartmentID != nil {
		query = query.Where("department_id = ?", filter.DepartmentID)
	}
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", filter.BranchID)
	}
	if filter.RequesterID != nil {
		query = query.Where("requester_id = ?", filter.RequesterID)
	}
	if filter.StartDate != nil {
		query = query.Where("target_date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("target_date <= ?", filter.EndDate)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("request_number ILIKE ? OR requester_name ILIKE ? OR requester_job_title ILIKE ? OR custom_position_title ILIKE ? OR custom_branch_name ILIKE ?", searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}

	err := query.Preload("Requester.JobPosition").
		Preload("Requester.Department").
		Preload("Approver.JobPosition").
		Preload("HRApprover.JobPosition").
		Preload("DirectorApprover.JobPosition").
		Preload("Department").
		Preload("Branch").
		Preload("JobPosition").
		Preload("Candidates").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&list).Error

	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *employeeRequestRepository) GenerateRequestNumber() (string, error) {
	var seq int64
	err := r.db.Raw("SELECT nextval('employee_requests_seq')").Scan(&seq).Error
	if err != nil {
		var count int64
		r.db.Model(&domain.EmployeeRequest{}).Count(&count)
		seq = count + 100
	}

	now := time.Now()
	// Format: REQ-YYYYMM-XXXX
	reqNum := fmt.Sprintf("REQ-%s-%04d", now.Format("200601"), seq)
	return reqNum, nil
}

func (r *employeeRequestRepository) GetSummaryStats(branchID *uuid.UUID, departmentID *uuid.UUID) (*EmployeeRequestSummaryStats, error) {
	var stats EmployeeRequestSummaryStats

	query := r.db.Model(&domain.EmployeeRequest{})
	if branchID != nil {
		query = query.Where("branch_id = ?", branchID)
	}
	if departmentID != nil {
		query = query.Where("department_id = ?", departmentID)
	}

	query.Count(&stats.TotalRequests)

	scopeFilter := func(db *gorm.DB) *gorm.DB {
		if branchID != nil {
			db = db.Where("branch_id = ?", branchID)
		}
		if departmentID != nil {
			db = db.Where("department_id = ?", departmentID)
		}
		return db
	}

	r.db.Model(&domain.EmployeeRequest{}).Scopes(scopeFilter).Where("status = 'PENDING'").Count(&stats.PendingRequests)
	r.db.Model(&domain.EmployeeRequest{}).Scopes(scopeFilter).Where("hr_status = 'PENDING' AND status != 'REJECTED' AND status != 'CANCELLED'").Count(&stats.WaitingHRApproval)
	r.db.Model(&domain.EmployeeRequest{}).Scopes(scopeFilter).Where("hr_status = 'APPROVED' AND director_status = 'PENDING' AND status != 'REJECTED' AND status != 'CANCELLED'").Count(&stats.WaitingDirectorApproval)
	r.db.Model(&domain.EmployeeRequest{}).Scopes(scopeFilter).Where("status = 'APPROVED'").Count(&stats.ApprovedRequests)
	r.db.Model(&domain.EmployeeRequest{}).Scopes(scopeFilter).Where("status = 'IN_PROGRESS'").Count(&stats.InProgressRequests)
	r.db.Model(&domain.EmployeeRequest{}).Scopes(scopeFilter).Where("status = 'FULFILLED'").Count(&stats.FulfilledRequests)
	r.db.Model(&domain.EmployeeRequest{}).Scopes(scopeFilter).Where("status = 'REJECTED'").Count(&stats.RejectedRequests)

	type SumResult struct {
		Needed    int64 `gorm:"column:needed"`
		Fulfilled int64 `gorm:"column:fulfilled"`
	}
	var sumRes SumResult
	r.db.Model(&domain.EmployeeRequest{}).Scopes(scopeFilter).Where("status != 'CANCELLED' AND status != 'REJECTED'").
		Select("COALESCE(SUM(headcount_needed), 0) as needed, COALESCE(SUM(headcount_fulfilled), 0) as fulfilled").
		Scan(&sumRes)

	stats.TotalHeadcountNeeded = sumRes.Needed
	stats.TotalHeadcountFulfilled = sumRes.Fulfilled

	return &stats, nil
}

func (r *employeeRequestRepository) AddCandidate(candidate *domain.EmployeeRequestCandidate) error {
	return r.db.Create(candidate).Error
}

func (r *employeeRequestRepository) DeleteCandidate(id uuid.UUID) error {
	return r.db.Delete(&domain.EmployeeRequestCandidate{}, "id = ?", id).Error
}

func (r *employeeRequestRepository) GetCandidatesByRequestID(requestID uuid.UUID) ([]domain.EmployeeRequestCandidate, error) {
	var list []domain.EmployeeRequestCandidate
	err := r.db.Where("employee_request_id = ?", requestID).Order("created_at ASC").Find(&list).Error
	return list, err
}
