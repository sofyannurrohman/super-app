import apiClient from './axios'

export interface EmployeeRequestCandidate {
  id: string
  employee_request_id: string
  candidate_name: string
  candidate_phone?: string
  candidate_email?: string
  join_date: string
  document_url?: string
  document_name?: string
  document_type?: 'PDF' | 'IMAGE' | string
  notes?: string
  created_at: string
  updated_at: string
}

export interface EmployeeRequest {
  id: string
  request_number: string
  requester_id?: string
  requester_name: string
  requester_job_title?: string
  department_id?: string
  branch_id?: string
  custom_branch_name?: string
  job_position_id?: string
  custom_position_title?: string
  employment_type: 'FULL_TIME' | 'CONTRACT' | 'DAILY' | 'INTERNSHIP'
  request_reason: 'NEW_HEADCOUNT' | 'REPLACEMENT' | 'PROJECT_BASED' | 'WORKLOAD_INCREASE' | 'OTHER'
  replacement_for_name?: string
  headcount_needed: number
  headcount_fulfilled: number
  min_education?: string
  gender_preference?: 'ANY' | 'MALE' | 'FEMALE'
  min_age: number
  max_age: number
  min_experience_years: number
  required_skills?: string
  job_description?: string
  notes?: string
  target_date: string
  status:
    | 'PENDING'
    | 'WAITING_DIRECTOR_APPROVAL'
    | 'WAITING_HR_APPROVAL'
    | 'APPROVED'
    | 'REJECTED'
    | 'IN_PROGRESS'
    | 'FULFILLED'
    | 'CANCELLED'

  approver_id?: string
  approval_date?: string
  approval_notes?: string

  hr_approver_id?: string
  hr_approval_date?: string
  hr_approval_notes?: string
  hr_status?: 'PENDING' | 'APPROVED' | 'REJECTED'

  director_approver_id?: string
  director_approval_date?: string
  director_approval_notes?: string
  director_status?: 'PENDING' | 'APPROVED' | 'REJECTED'

  rejected_by_role?: 'HR_MANAGER' | 'DIRECTOR'
  reject_reason?: string
  recruitment_notes?: string
  created_at: string
  updated_at: string
  requester?: any
  approver?: any
  hr_approver?: any
  director_approver?: any
  department?: any
  branch?: any
  job_position?: any
  candidates?: EmployeeRequestCandidate[]
}

export interface EmployeeRequestSummary {
  total_requests: number
  pending_requests: number
  waiting_hr_approval: number
  waiting_director_approval: number
  approved_requests: number
  in_progress_requests: number
  fulfilled_requests: number
  rejected_requests: number
  total_headcount_needed: number
  total_headcount_fulfilled: number
}

export interface EmployeeRequestParams {
  status?: string
  department_id?: string
  branch_id?: string
  search?: string
  start_date?: string
  end_date?: string
  limit?: number
  offset?: number
}

export interface CandidatePayload {
  candidate_name: string
  candidate_phone?: string
  candidate_email?: string
  join_date: string
  document_url?: string
  document_name?: string
  document_type?: string
  notes?: string
}

const employeeRequestApi = {
  getEmployeeRequests(params?: EmployeeRequestParams) {
    return apiClient.get<{ data: EmployeeRequest[]; meta: { total: number; limit: number; offset: number } }>(
      '/employee-requests',
      { params }
    )
  },

  getEmployeeRequestSummary(params?: { branch_id?: string; department_id?: string }) {
    return apiClient.get<{ data: EmployeeRequestSummary }>('/employee-requests/summary', { params })
  },

  getEmployeeRequestById(id: string) {
    return apiClient.get<{ data: EmployeeRequest }>(`/employee-requests/${id}`)
  },

  createEmployeeRequest(data: Partial<EmployeeRequest>) {
    return apiClient.post<{ message: string; data: EmployeeRequest }>('/employee-requests', data)
  },

  updateEmployeeRequest(id: string, data: Partial<EmployeeRequest>) {
    return apiClient.put<{ message: string; data: EmployeeRequest }>(`/employee-requests/${id}`, data)
  },

  deleteEmployeeRequest(id: string) {
    return apiClient.delete<{ message: string }>(`/employee-requests/${id}`)
  },

  // Persetujuan HR Manager
  approveByHR(id: string, notes?: string) {
    return apiClient.put<{ message: string; data: EmployeeRequest }>(`/employee-requests/${id}/approve-hr`, {
      notes: notes || '',
    })
  },

  // Persetujuan Direktur
  approveByDirector(id: string, notes?: string) {
    return apiClient.put<{ message: string; data: EmployeeRequest }>(`/employee-requests/${id}/approve-director`, {
      notes: notes || '',
    })
  },

  // Full Approval (Superadmin)
  approveEmployeeRequest(id: string, notes?: string) {
    return apiClient.put<{ message: string; data: EmployeeRequest }>(`/employee-requests/${id}/approve`, {
      notes: notes || '',
    })
  },

  rejectEmployeeRequest(id: string, reason: string, role: 'HR_MANAGER' | 'DIRECTOR' = 'HR_MANAGER') {
    return apiClient.put<{ message: string; data: EmployeeRequest }>(`/employee-requests/${id}/reject`, {
      reason,
      role,
    })
  },

  uploadCandidateDocument(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return apiClient.post<{
      message: string
      document_url: string
      document_name: string
      document_type: string
    }>('/employee-requests/upload-cv', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  addCandidates(
    requestId: string,
    payload: { candidates: CandidatePayload[]; recruitment_notes?: string }
  ) {
    return apiClient.post<{ message: string; data: EmployeeRequest }>(
      `/employee-requests/${requestId}/candidates`,
      payload
    )
  },

  deleteCandidate(requestId: string, candidateId: string) {
    return apiClient.delete<{ message: string; data: EmployeeRequest }>(
      `/employee-requests/${requestId}/candidates/${candidateId}`
    )
  },

  updateProgress(
    id: string,
    payload: {
      status?: string
      headcount_fulfilled?: number
      recruitment_notes?: string
      candidates?: CandidatePayload[]
    }
  ) {
    return apiClient.put<{ message: string; data: EmployeeRequest }>(`/employee-requests/${id}/progress`, payload)
  },
}

export default employeeRequestApi
