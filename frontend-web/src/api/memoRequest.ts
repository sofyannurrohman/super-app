import apiClient from './axios'

export default {
  getMemoRequests(params?: any) {
    return apiClient.get('/memo-requests', { params })
  },
  getMemoRequestById(id: string) {
    return apiClient.get(`/memo-requests/${id}`)
  },
  createMemoRequest(data: any) {
    return apiClient.post('/memo-requests', data)
  },
  approveMemoRequest(id: string) {
    return apiClient.put(`/memo-requests/${id}/approve`)
  }
}
