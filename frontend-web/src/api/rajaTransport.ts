import apiClient from './axios'

export const rajaTransportApi = {
  // Stats
  getDashboardStats: (params?: any) => apiClient.get('/transport/dashboard/stats', { params }),

  // Trips
  getTrips: (params?: any) => apiClient.get('/transport/trips', { params }),
  getTrip: (id: string) => apiClient.get(`/transport/trips/${id}`),
  createTrip: (data: any) => apiClient.post('/transport/trips', data),
  updateTrip: (id: string, data: any) => apiClient.put(`/transport/trips/${id}`, data),
  completeTrip: (id: string, data: { actual_return_date: string; end_odometer: number; notes?: string }) =>
    apiClient.post(`/transport/trips/${id}/complete`, data),
  deleteTrip: (id: string) => apiClient.delete(`/transport/trips/${id}`),

  // Backhauls (Pengiriman Kembali)
  getBackhauls: (params?: any) => apiClient.get('/transport/backhauls', { params }),
  getBackhaul: (id: string) => apiClient.get(`/transport/backhauls/${id}`),
  createBackhaul: (data: FormData | any) => {
    const isFormData = data instanceof FormData
    return apiClient.post('/transport/backhauls', data, {
      headers: isFormData ? { 'Content-Type': 'multipart/form-data' } : undefined,
    })
  },
  approvalBackhaul: (id: string, data: { action: 'APPROVED' | 'REJECTED'; notes?: string; reason?: string }) =>
    apiClient.put(`/transport/backhauls/${id}/approval`, data),
  deleteBackhaul: (id: string) => apiClient.delete(`/transport/backhauls/${id}`),

  // Expense Claims (Struk BBM & Part)
  getExpenseClaims: (params?: any) => apiClient.get('/transport/claims', { params }),
  getExpenseClaim: (id: string) => apiClient.get(`/transport/claims/${id}`),
  createExpenseClaim: (data: FormData | any) => {
    const isFormData = data instanceof FormData
    return apiClient.post('/transport/claims', data, {
      headers: isFormData ? { 'Content-Type': 'multipart/form-data' } : undefined,
    })
  },
  approvalExpenseClaim: (id: string, data: { action: 'APPROVED' | 'REJECTED'; reason?: string }) =>
    apiClient.put(`/transport/claims/${id}/approval`, data),
  disburseExpenseClaim: (id: string) => apiClient.put(`/transport/claims/${id}/disburse`),
  deleteExpenseClaim: (id: string) => apiClient.delete(`/transport/claims/${id}`),

  // Fleet Inspections (Kondisi & Perawatan Armada)
  getFleetInspections: (params?: any) => apiClient.get('/transport/inspections', { params }),
  getFleetInspection: (id: string) => apiClient.get(`/transport/inspections/${id}`),
  createFleetInspection: (data: FormData | any) => {
    const isFormData = data instanceof FormData
    return apiClient.post('/transport/inspections', data, {
      headers: isFormData ? { 'Content-Type': 'multipart/form-data' } : undefined,
    })
  },
  updateInspectionFollowUp: (id: string, data: {
    workshop_name?: string
    parts_replaced?: string
    actual_cost?: number
    follow_up_status: 'PENDING_REPAIR' | 'IN_PROGRESS' | 'COMPLETED'
    repair_notes?: string
  }) => apiClient.put(`/transport/inspections/${id}/follow-up`, data),
  deleteFleetInspection: (id: string) => apiClient.delete(`/transport/inspections/${id}`),

  // Driver Attendances
  getDriverAttendances: (params?: any) => apiClient.get('/transport/attendances', { params }),
  recordDriverAttendance: (data: any) => apiClient.post('/transport/attendances', data),
}
