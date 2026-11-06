import apiClient from './axios'

export const warehouseApi = {
  getInventory: (branchId?: string | null) =>
    apiClient.get('/warehouse/stock', { params: { branch_id: branchId } }),
  getLogs: (branchId?: string | null) =>
    apiClient.get('/warehouse/logs', { params: { branch_id: branchId } }),
  getPendingShipments: (branchId?: string | null) =>
    apiClient.get('/warehouse/transfers/pending', { params: { branch_id: branchId } }),
  getAllTransfers: (branchId?: string | null) =>
    apiClient.get('/warehouse/transfers/all', { params: { branch_id: branchId } }),
  receiveShipment: (id: string, data: any) =>
    apiClient.post(`/warehouse/transfers/${id}/receive`, data),
  correctShipmentReceipt: (id: string, data: any) =>
    apiClient.post(`/warehouse/transfers/${id}/correct`, data),
  arriveShipment: (id: string) => apiClient.post(`/warehouse/transfers/${id}/arrive`),
  approveShipment: (id: string) => apiClient.post(`/warehouse/transfers/${id}/approve`),
  rejectShipment: (id: string, reason: string) =>
    apiClient.post(`/warehouse/transfers/${id}/reject`, { rejection_reason: reason }),
  confirmShipment: (id: string) => apiClient.post(`/warehouse/transfers/${id}/confirm-shipment`),
  setStockLimit: (data: any) =>
    apiClient.post('/warehouse/stock-limit', data, {
      params: { branch_id: data.branch_id },
    }),
  adjustStock: (data: any) =>
    apiClient.post('/warehouse/adjust', data, {
      params: { branch_id: data.branch_id },
    }),
  deleteStock: (id: string, data: { reason?: string; branch_id?: string }) =>
    apiClient.delete(`/warehouse/stock/${id}`, { data }),
  getByDO: (doNo: string) => apiClient.get(`/warehouse/transfers/do/${doNo}`),
  receiveByDO: (doNo: string, data: any) =>
    apiClient.post(`/warehouse/transfers/do/${doNo}/receive`, data),
  requestGoods: (factoryId: string, data: any) =>
    apiClient.post(`/factory/${factoryId}/transfer/request`, data),
  getWarehouseDO: (batchId: string) =>
    apiClient.get(`/admin/warehouse-delivery-orders/batch/${batchId}`),
  getWarehouseDOById: (id: string) =>
    apiClient.get(`/admin/warehouse-delivery-orders/id/${id}`),
  getMotorisDOs: (branchId?: string | null) =>
    apiClient.get('/admin/warehouse-delivery-orders/motoris/list', { params: { branch_id: branchId } }),
  createOrGetWarehouseDO: (batchId: string) =>
    apiClient.post('/admin/warehouse-delivery-orders/get-or-create', { delivery_batch_id: batchId }),
  updateWarehouseDO: (data: any) =>
    apiClient.patch('/admin/warehouse-delivery-orders', data),
  dispatchWarehouseDO: (id: string) =>
    apiClient.post(`/admin/warehouse-delivery-orders/${id}/dispatch`),
  finalizeWarehouseDO: (id: string) =>
    apiClient.post(`/admin/warehouse-delivery-orders/${id}/finalize`),
}
