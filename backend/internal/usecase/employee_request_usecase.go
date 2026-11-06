package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
)

type CreateEmployeeRequestPayload struct {
	DepartmentID        *uuid.UUID `json:"department_id"`
	BranchID            *uuid.UUID `json:"branch_id"`
	CustomBranchName    string     `json:"custom_branch_name"`
	JobPositionID       *uuid.UUID `json:"job_position_id"`
	CustomPositionTitle string     `json:"custom_position_title"`
	EmploymentType      string     `json:"employment_type"` // FULL_TIME, CONTRACT, DAILY, INTERNSHIP
	RequestReason       string     `json:"request_reason"`  // NEW_HEADCOUNT, REPLACEMENT, PROJECT_BASED, WORKLOAD_INCREASE, OTHER
	ReplacementForName  string     `json:"replacement_for_name"`
	HeadcountNeeded     int        `json:"headcount_needed"`
	MinEducation        string     `json:"min_education"`
	GenderPreference    string     `json:"gender_preference"`
	MinAge              int        `json:"min_age"`
	MaxAge              int        `json:"max_age"`
	MinExperienceYears  int        `json:"min_experience_years"`
	RequiredSkills      string     `json:"required_skills"`
	JobDescription      string     `json:"job_description"`
	Notes               string     `json:"notes"`
	TargetDate          string     `json:"target_date"` // YYYY-MM-DD
}

type UpdateEmployeeRequestPayload struct {
	DepartmentID        *uuid.UUID `json:"department_id"`
	BranchID            *uuid.UUID `json:"branch_id"`
	CustomBranchName    string     `json:"custom_branch_name"`
	JobPositionID       *uuid.UUID `json:"job_position_id"`
	CustomPositionTitle string     `json:"custom_position_title"`
	EmploymentType      string     `json:"employment_type"`
	RequestReason       string     `json:"request_reason"`
	ReplacementForName  string     `json:"replacement_for_name"`
	HeadcountNeeded     int        `json:"headcount_needed"`
	MinEducation        string     `json:"min_education"`
	GenderPreference    string     `json:"gender_preference"`
	MinAge              int        `json:"min_age"`
	MaxAge              int        `json:"max_age"`
	MinExperienceYears  int        `json:"min_experience_years"`
	RequiredSkills      string     `json:"required_skills"`
	JobDescription      string     `json:"job_description"`
	Notes               string     `json:"notes"`
	TargetDate          string     `json:"target_date"`
}

type CandidateInputPayload struct {
	CandidateName string `json:"candidate_name"`
	CandidatePhone string `json:"candidate_phone"`
	CandidateEmail string `json:"candidate_email"`
	JoinDate      string `json:"join_date"` // YYYY-MM-DD
	DocumentURL   string `json:"document_url"`
	DocumentName  string `json:"document_name"`
	DocumentType  string `json:"document_type"`
	Notes         string `json:"notes"`
}

type AddCandidatesPayload struct {
	Candidates       []CandidateInputPayload `json:"candidates"`
	RecruitmentNotes string                  `json:"recruitment_notes"`
}

type UpdateProgressPayload struct {
	Status             string                  `json:"status"` // IN_PROGRESS, FULFILLED, CANCELLED
	HeadcountFulfilled int                     `json:"headcount_fulfilled"`
	RecruitmentNotes   string                  `json:"recruitment_notes"`
	Candidates         []CandidateInputPayload `json:"candidates,omitempty"`
}

type EmployeeRequestUseCase interface {
	Create(payload *CreateEmployeeRequestPayload, requesterID *uuid.UUID, requesterName string, requesterJobTitle string) (*domain.EmployeeRequest, error)
	Update(id uuid.UUID, payload *UpdateEmployeeRequestPayload) (*domain.EmployeeRequest, error)
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*domain.EmployeeRequest, error)
	GetAll(filter repository.EmployeeRequestFilter, limit, offset int) ([]domain.EmployeeRequest, int64, error)
	GetSummaryStats(branchID *uuid.UUID, departmentID *uuid.UUID) (*repository.EmployeeRequestSummaryStats, error)
	ApproveHR(id uuid.UUID, approverID uuid.UUID, notes string) (*domain.EmployeeRequest, error)
	ApproveDirector(id uuid.UUID, approverID uuid.UUID, notes string) (*domain.EmployeeRequest, error)
	Approve(id uuid.UUID, approverID uuid.UUID, notes string) (*domain.EmployeeRequest, error)
	Reject(id uuid.UUID, approverID uuid.UUID, reason string, role string) (*domain.EmployeeRequest, error)
	UpdateProgress(id uuid.UUID, payload *UpdateProgressPayload) (*domain.EmployeeRequest, error)
	AddCandidates(id uuid.UUID, payload *AddCandidatesPayload) (*domain.EmployeeRequest, error)
	DeleteCandidate(requestID uuid.UUID, candidateID uuid.UUID) (*domain.EmployeeRequest, error)
}

type employeeRequestUseCase struct {
	repo         repository.EmployeeRequestRepository
	employeeRepo repository.EmployeeRepository
}

func NewEmployeeRequestUseCase(
	repo repository.EmployeeRequestRepository,
	employeeRepo repository.EmployeeRepository,
) EmployeeRequestUseCase {
	return &employeeRequestUseCase{
		repo:         repo,
		employeeRepo: employeeRepo,
	}
}

func (u *employeeRequestUseCase) Create(payload *CreateEmployeeRequestPayload, requesterID *uuid.UUID, requesterName string, requesterJobTitle string) (*domain.EmployeeRequest, error) {
	if payload.HeadcountNeeded <= 0 {
		payload.HeadcountNeeded = 1
	}

	targetDate, err := time.Parse("2006-01-02", payload.TargetDate)
	if err != nil {
		return nil, errors.New("format tanggal target (deadline) tidak valid (gunakan YYYY-MM-DD)")
	}

	reqNumber, err := u.repo.GenerateRequestNumber()
	if err != nil {
		return nil, err
	}

	req := &domain.EmployeeRequest{
		RequestNumber:       reqNumber,
		RequesterID:         requesterID,
		RequesterName:       requesterName,
		RequesterJobTitle:   requesterJobTitle,
		DepartmentID:        payload.DepartmentID,
		BranchID:            payload.BranchID,
		CustomBranchName:    payload.CustomBranchName,
		JobPositionID:       payload.JobPositionID,
		CustomPositionTitle: payload.CustomPositionTitle,
		EmploymentType:      payload.EmploymentType,
		RequestReason:       payload.RequestReason,
		ReplacementForName:  payload.ReplacementForName,
		HeadcountNeeded:     payload.HeadcountNeeded,
		HeadcountFulfilled:  0,
		MinEducation:        payload.MinEducation,
		GenderPreference:    payload.GenderPreference,
		MinAge:              payload.MinAge,
		MaxAge:              payload.MaxAge,
		MinExperienceYears:  payload.MinExperienceYears,
		RequiredSkills:      payload.RequiredSkills,
		JobDescription:      payload.JobDescription,
		Notes:               payload.Notes,
		TargetDate:          targetDate,
		Status:              "PENDING",
		HRStatus:            "PENDING",
		DirectorStatus:      "PENDING",
	}

	if req.EmploymentType == "" {
		req.EmploymentType = "FULL_TIME"
	}
	if req.RequestReason == "" {
		req.RequestReason = "NEW_HEADCOUNT"
	}
	if req.GenderPreference == "" {
		req.GenderPreference = "ANY"
	}
	if req.MinAge <= 0 {
		req.MinAge = 18
	}
	if req.MaxAge <= 0 {
		req.MaxAge = 45
	}

	if err := u.repo.Create(req); err != nil {
		return nil, err
	}

	return u.repo.GetByID(req.ID)
}

func (u *employeeRequestUseCase) Update(id uuid.UUID, payload *UpdateEmployeeRequestPayload) (*domain.EmployeeRequest, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("data permintaan karyawan tidak ditemukan")
	}

	if existing.Status != "PENDING" && existing.Status != "WAITING_DIRECTOR_APPROVAL" && existing.Status != "WAITING_HR_APPROVAL" {
		return nil, errors.New("hanya permintaan yang belum disetujui penuh yang dapat diubah")
	}

	if payload.TargetDate != "" {
		targetDate, err := time.Parse("2006-01-02", payload.TargetDate)
		if err == nil {
			existing.TargetDate = targetDate
		}
	}

	if payload.DepartmentID != nil {
		existing.DepartmentID = payload.DepartmentID
	}
	if payload.BranchID != nil {
		existing.BranchID = payload.BranchID
		existing.CustomBranchName = payload.CustomBranchName
	} else if payload.CustomBranchName != "" {
		existing.BranchID = nil
		existing.CustomBranchName = payload.CustomBranchName
	}
	if payload.JobPositionID != nil {
		existing.JobPositionID = payload.JobPositionID
	}
	if payload.CustomPositionTitle != "" {
		existing.CustomPositionTitle = payload.CustomPositionTitle
	}
	if payload.EmploymentType != "" {
		existing.EmploymentType = payload.EmploymentType
	}
	if payload.RequestReason != "" {
		existing.RequestReason = payload.RequestReason
	}
	existing.ReplacementForName = payload.ReplacementForName
	if payload.HeadcountNeeded > 0 {
		existing.HeadcountNeeded = payload.HeadcountNeeded
	}
	existing.MinEducation = payload.MinEducation
	if payload.GenderPreference != "" {
		existing.GenderPreference = payload.GenderPreference
	}
	if payload.MinAge > 0 {
		existing.MinAge = payload.MinAge
	}
	if payload.MaxAge > 0 {
		existing.MaxAge = payload.MaxAge
	}
	existing.MinExperienceYears = payload.MinExperienceYears
	existing.RequiredSkills = payload.RequiredSkills
	existing.JobDescription = payload.JobDescription
	existing.Notes = payload.Notes

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id)
}

func (u *employeeRequestUseCase) Delete(id uuid.UUID) error {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("data tidak ditemukan")
	}
	return u.repo.Delete(id)
}

func (u *employeeRequestUseCase) GetByID(id uuid.UUID) (*domain.EmployeeRequest, error) {
	return u.repo.GetByID(id)
}

func (u *employeeRequestUseCase) GetAll(filter repository.EmployeeRequestFilter, limit, offset int) ([]domain.EmployeeRequest, int64, error) {
	return u.repo.GetAll(filter, limit, offset)
}

func (u *employeeRequestUseCase) GetSummaryStats(branchID *uuid.UUID, departmentID *uuid.UUID) (*repository.EmployeeRequestSummaryStats, error) {
	return u.repo.GetSummaryStats(branchID, departmentID)
}

func (u *employeeRequestUseCase) ApproveHR(id uuid.UUID, approverID uuid.UUID, notes string) (*domain.EmployeeRequest, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("permintaan karyawan tidak ditemukan")
	}

	if existing.Status == "REJECTED" || existing.Status == "CANCELLED" {
		return nil, errors.New("pengajuan yang sudah ditolak atau dibatalkan tidak dapat disetujui")
	}

	now := time.Now()
	existing.HRApproverID = &approverID
	existing.HRApprovalDate = &now
	existing.HRApprovalNotes = notes
	existing.HRStatus = "APPROVED"

	if existing.DirectorStatus == "APPROVED" {
		existing.Status = "APPROVED"
		existing.ApproverID = &approverID
		existing.ApprovalDate = &now
		existing.ApprovalNotes = notes
	} else {
		existing.Status = "WAITING_DIRECTOR_APPROVAL"
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id)
}

func (u *employeeRequestUseCase) ApproveDirector(id uuid.UUID, approverID uuid.UUID, notes string) (*domain.EmployeeRequest, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("permintaan karyawan tidak ditemukan")
	}

	if existing.Status == "REJECTED" || existing.Status == "CANCELLED" {
		return nil, errors.New("pengajuan yang sudah ditolak atau dibatalkan tidak dapat disetujui")
	}

	now := time.Now()
	existing.DirectorApproverID = &approverID
	existing.DirectorApprovalDate = &now
	existing.DirectorApprovalNotes = notes
	existing.DirectorStatus = "APPROVED"

	if existing.HRStatus == "APPROVED" {
		existing.Status = "APPROVED"
		existing.ApproverID = &approverID
		existing.ApprovalDate = &now
		existing.ApprovalNotes = notes
	} else {
		existing.Status = "WAITING_HR_APPROVAL"
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id)
}

func (u *employeeRequestUseCase) Approve(id uuid.UUID, approverID uuid.UUID, notes string) (*domain.EmployeeRequest, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("permintaan karyawan tidak ditemukan")
	}

	now := time.Now()
	existing.HRApproverID = &approverID
	existing.HRApprovalDate = &now
	existing.HRApprovalNotes = notes
	existing.HRStatus = "APPROVED"

	existing.DirectorApproverID = &approverID
	existing.DirectorApprovalDate = &now
	existing.DirectorApprovalNotes = notes
	existing.DirectorStatus = "APPROVED"

	existing.Status = "APPROVED"
	existing.ApproverID = &approverID
	existing.ApprovalDate = &now
	existing.ApprovalNotes = notes
	existing.RejectReason = ""

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id)
}

func (u *employeeRequestUseCase) Reject(id uuid.UUID, approverID uuid.UUID, reason string, role string) (*domain.EmployeeRequest, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("permintaan karyawan tidak ditemukan")
	}

	if reason == "" {
		return nil, errors.New("alasan penolakan wajib diisi")
	}

	now := time.Now()
	existing.Status = "REJECTED"
	existing.ApproverID = &approverID
	existing.ApprovalDate = &now
	existing.RejectReason = reason

	if role == "DIRECTOR" {
		existing.DirectorApproverID = &approverID
		existing.DirectorApprovalDate = &now
		existing.DirectorStatus = "REJECTED"
		existing.DirectorApprovalNotes = reason
		existing.RejectedByRole = "DIRECTOR"
	} else {
		existing.HRApproverID = &approverID
		existing.HRApprovalDate = &now
		existing.HRStatus = "REJECTED"
		existing.HRApprovalNotes = reason
		existing.RejectedByRole = "HR_MANAGER"
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id)
}

func (u *employeeRequestUseCase) AddCandidates(id uuid.UUID, payload *AddCandidatesPayload) (*domain.EmployeeRequest, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("permintaan karyawan tidak ditemukan")
	}

	for _, cand := range payload.Candidates {
		if cand.CandidateName == "" {
			continue
		}
		joinDate := time.Now()
		if cand.JoinDate != "" {
			if parsedDate, err := time.Parse("2006-01-02", cand.JoinDate); err == nil {
				joinDate = parsedDate
			}
		}

		c := &domain.EmployeeRequestCandidate{
			EmployeeRequestID: id,
			CandidateName:     cand.CandidateName,
			CandidatePhone:    cand.CandidatePhone,
			CandidateEmail:    cand.CandidateEmail,
			JoinDate:          joinDate,
			DocumentURL:       cand.DocumentURL,
			DocumentName:      cand.DocumentName,
			DocumentType:      cand.DocumentType,
			Notes:             cand.Notes,
		}
		if err := u.repo.AddCandidate(c); err != nil {
			return nil, err
		}
	}

	// Update headcount fulfilled from database candidate count
	candidates, err := u.repo.GetCandidatesByRequestID(id)
	if err == nil {
		existing.HeadcountFulfilled = len(candidates)
	}

	if payload.RecruitmentNotes != "" {
		existing.RecruitmentNotes = payload.RecruitmentNotes
	}

	if existing.HeadcountFulfilled >= existing.HeadcountNeeded {
		existing.Status = "FULFILLED"
	} else if existing.HeadcountFulfilled > 0 {
		existing.Status = "IN_PROGRESS"
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id)
}

func (u *employeeRequestUseCase) DeleteCandidate(requestID uuid.UUID, candidateID uuid.UUID) (*domain.EmployeeRequest, error) {
	if err := u.repo.DeleteCandidate(candidateID); err != nil {
		return nil, err
	}

	existing, err := u.repo.GetByID(requestID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("permintaan karyawan tidak ditemukan")
	}

	candidates, err := u.repo.GetCandidatesByRequestID(requestID)
	if err == nil {
		existing.HeadcountFulfilled = len(candidates)
	}

	if existing.HeadcountFulfilled >= existing.HeadcountNeeded {
		existing.Status = "FULFILLED"
	} else if existing.HeadcountFulfilled > 0 {
		existing.Status = "IN_PROGRESS"
	} else if existing.Status == "FULFILLED" {
		existing.Status = "IN_PROGRESS"
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return u.repo.GetByID(requestID)
}

func (u *employeeRequestUseCase) UpdateProgress(id uuid.UUID, payload *UpdateProgressPayload) (*domain.EmployeeRequest, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("permintaan karyawan tidak ditemukan")
	}

	if len(payload.Candidates) > 0 {
		for _, cand := range payload.Candidates {
			if cand.CandidateName == "" {
				continue
			}
			joinDate := time.Now()
			if cand.JoinDate != "" {
				if parsedDate, err := time.Parse("2006-01-02", cand.JoinDate); err == nil {
					joinDate = parsedDate
				}
			}

			c := &domain.EmployeeRequestCandidate{
				EmployeeRequestID: id,
				CandidateName:     cand.CandidateName,
				CandidatePhone:    cand.CandidatePhone,
				CandidateEmail:    cand.CandidateEmail,
				JoinDate:          joinDate,
				DocumentURL:       cand.DocumentURL,
				DocumentName:      cand.DocumentName,
				DocumentType:      cand.DocumentType,
				Notes:             cand.Notes,
			}
			_ = u.repo.AddCandidate(c)
		}

		candidates, err := u.repo.GetCandidatesByRequestID(id)
		if err == nil {
			existing.HeadcountFulfilled = len(candidates)
		}
	} else if payload.HeadcountFulfilled >= 0 {
		existing.HeadcountFulfilled = payload.HeadcountFulfilled
	}

	if payload.RecruitmentNotes != "" {
		existing.RecruitmentNotes = payload.RecruitmentNotes
	}

	// Auto update status if fulfilled
	if payload.Status != "" {
		existing.Status = payload.Status
	} else if existing.HeadcountFulfilled >= existing.HeadcountNeeded {
		existing.Status = "FULFILLED"
	} else if existing.HeadcountFulfilled > 0 {
		existing.Status = "IN_PROGRESS"
	}

	if err := u.repo.Update(existing); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id)
}
