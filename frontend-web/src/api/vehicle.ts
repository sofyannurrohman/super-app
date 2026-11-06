import apiClient from './axios'

export const vehicleApi = {
  getVehicles: (params?: any) => apiClient.get('/vehicles', { params }),
  getVehicle: (id: string) => apiClient.get(`/vehicles/${id}`),
  createVehicle: (data: any) => apiClient.post('/vehicles', data),
  updateVehicle: (id: string, data: any) => apiClient.put(`/vehicles/${id}`, data),
  deleteVehicle: (id: string) => apiClient.delete(`/vehicles/${id}`),
  getVehicleLogs: (id: string) => apiClient.get(`/vehicles/${id}/logs`),
  getActiveTask: (id: string) => apiClient.get(`/vehicles/${id}/active-task`),
}
