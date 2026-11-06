import apiClient from './axios'

export const factoryReturnApi = {
  createReturn: (data: any) => apiClient.post('/factory-returns', data),
  receiveReturn: (id: string, data?: any) => apiClient.post(`/factory-returns/${id}/receive`, data || {}),
  getByBranch: (branchId: string) => apiClient.get(`/factory-returns/branch/${branchId}`),
  getByFactory: (factoryId: string) => apiClient.get(`/factory-returns/factory/${factoryId}`),
  getByCompany: () => apiClient.get('/factory-returns/company'),
}
