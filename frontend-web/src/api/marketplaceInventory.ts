import apiClient from './axios'

export const marketplaceInventoryApi = {
  getAll: () => apiClient.get('/marketplace/inventories'),
  getById: (id: string) => apiClient.get(`/marketplace/inventories/${id}`),
  create: (data: any) => apiClient.post('/marketplace/inventories', data),
  update: (id: string, data: any) => apiClient.put(`/marketplace/inventories/${id}`, data),
  delete: (id: string) => apiClient.delete(`/marketplace/inventories/${id}`),
  stockIn: (id: string, data: any) => apiClient.post(`/marketplace/inventories/${id}/in`, data),
  stockOut: (id: string, data: any) => apiClient.post(`/marketplace/inventories/${id}/out`, data),
  doOpname: (id: string, data: any) => apiClient.post(`/marketplace/inventories/${id}/opname`, data),
  getLogs: (id: string, params?: any) => apiClient.get(`/marketplace/inventories/${id}/logs`, { params }),
  getOpnames: (id: string, params?: any) => apiClient.get(`/marketplace/inventories/${id}/opnames`, { params }),
}
