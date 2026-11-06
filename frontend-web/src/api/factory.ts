import apiClient from './axios'

export const factoryApi = {
  // Factories
  getFactories: () => apiClient.get('/factory'),
  createFactory: (data: any) => apiClient.post('/factory', data),
  getFactory: (id: string) => apiClient.get(`/factory/${id}`),
  updateFactory: (id: string, data: any) => apiClient.put(`/factory/${id}`, data),
  deleteFactory: (id: string) => apiClient.delete(`/factory/${id}`),

  // Factory Warehouses
  getFactoryWarehouses: (factoryId: string) => apiClient.get(`/factory/${factoryId}/warehouses`),
  createFactoryWarehouse: (factoryId: string, data: any) => apiClient.post(`/factory/${factoryId}/warehouses`, data),
  updateFactoryWarehouse: (warehouseId: string, data: any) => apiClient.put(`/factory/warehouses/${warehouseId}`, data),
  deleteFactoryWarehouse: (warehouseId: string) => apiClient.delete(`/factory/warehouses/${warehouseId}`),

  // Products
  getProducts: () => apiClient.get('/factory/products'),
  createProduct: (data: any) => apiClient.post('/factory/products', data),
  updateProduct: (id: string, data: any) => apiClient.put(`/factory/products/${id}`, data),
  deleteProduct: (id: string) => apiClient.delete(`/factory/products/${id}`),

  // Inventory
  getInventory: (factoryId: string) => apiClient.get(`/factory/${factoryId}/inventory`),
  getAllInventory: () => apiClient.get('/factory/stock/inventory'),
  getInventoryLogs: (factoryId: string) => apiClient.get(`/factory/${factoryId}/inventory/logs`),
  getAllInventoryLogs: () => apiClient.get('/factory/stock/inventory/logs'),
  adjustStock: (factoryId: string, data: any) =>
    apiClient.post(`/factory/${factoryId}/inventory/adjust`, data),
  syncWarehouseStock: (factoryId: string) =>
    apiClient.post(`/factory/${factoryId}/inventory/sync-warehouse`),
  syncAllWarehouseStock: () =>
    apiClient.post('/factory/sync-all-warehouses'),

  // Production
  logProduction: (factoryId: string, data: any) =>
    apiClient.post(`/factory/${factoryId}/production`, data),
  logBatchProduction: (factoryId: string, data: any) =>
    apiClient.post(`/factory/${factoryId}/production/batch`, data),
  getProductionHistory: (factoryId: string) => apiClient.get(`/factory/${factoryId}/production`),
  getAllProductionHistory: () => apiClient.get('/factory/stock/production'),
  updateProductionLog: (id: string, data: any) => apiClient.put(`/factory/production/${id}`, data),
  deleteProductionLog: (id: string) => apiClient.delete(`/factory/production/${id}`),

  // Transfers
  sendToBranch: (factoryId: string, data: any) =>
    apiClient.post(`/factory/${factoryId}/transfer/request`, data),
  getTransferHistory: (factoryId: string) => apiClient.get(`/factory/${factoryId}/transfer`),
  getAllTransfers: (params?: any) => apiClient.get('/factory/stock/transfers', { params }),
  getDelegatedShipments: () => apiClient.get('/factory/delegated-shipments'),
  shipTransfer: (id: string) => apiClient.post(`/factory/transfer/${id}/ship`),
  approveTransfer: (id: string) => apiClient.put(`/factory/transfer/${id}/approve`),
  updateTransfer: (id: string, data: any) => apiClient.put(`/factory/transfer/${id}`, data),
  deleteTransfer: (id: string) => apiClient.delete(`/factory/transfer/${id}`),
  createTransfer: (data: any) => apiClient.post('/factory/transfer', data),
  getBackorderDemand: (companyId: string) =>
    apiClient.get(`/factory/demand?company_id=${companyId}`),
  getDashboardStats: () => apiClient.get('/factory/dashboard/stats'),
  getLowStockAlerts: () => apiClient.get('/factory/stock/low-alerts'),
  getBranchReceivables: () => apiClient.get('/factory/receivables'),
  getRawMaterialStock: (id: string) => apiClient.get(`/factory/${id}/raw-materials`),
  getRawMaterialLogs: (id: string, params?: { product_id?: string }) =>
    apiClient.get(`/factory/${id}/raw-materials/logs`, { params }),
  mutateRawMaterial: (id: string, data: any) => apiClient.post(`/factory/${id}/raw-materials/mutate`, data),
  deleteStock: (id: string) => apiClient.delete(`/factory/stocks/${id}`),
  deleteRawMaterialLog: (id: string) => apiClient.delete(`/factory/raw-materials/logs/${id}`),

  // Raw Material Purchase Requests (PR)
  getRawMaterialPRs: (params?: { factory_id?: string; status?: string }) =>
    apiClient.get('/factory/raw-materials/requests', { params }),
  getRawMaterialPRDetail: (id: string) =>
    apiClient.get(`/factory/raw-materials/requests/${id}`),
  createRawMaterialPR: (data: any) =>
    apiClient.post('/factory/raw-materials/requests', data),
  updateRawMaterialPR: (id: string, data: any) =>
    apiClient.put(`/factory/raw-materials/requests/${id}`, data),
  rejectRawMaterialPR: (id: string, data: { reason: string }) =>
    apiClient.post(`/factory/raw-materials/requests/${id}/reject`, data),
  completeRawMaterialPR: (id: string, data: { notes: string }) =>
    apiClient.post(`/factory/raw-materials/requests/${id}/complete`, data),
  deleteRawMaterialPR: (id: string) =>
    apiClient.delete(`/factory/raw-materials/requests/${id}`),

  // Raw Material Purchase Orders
  getRawMaterialPOs: (params?: { factory_id?: string; status?: string }) =>
    apiClient.get('/factory/raw-materials/orders', { params }),
  getRawMaterialPODetail: (id: string) =>
    apiClient.get(`/factory/raw-materials/orders/${id}`),
  createRawMaterialPO: (data: any) =>
    apiClient.post('/factory/raw-materials/orders', data),
  updateRawMaterialPO: (id: string, data: any) =>
    apiClient.put(`/factory/raw-materials/orders/${id}`, data),
  deleteRawMaterialPO: (id: string) =>
    apiClient.delete(`/factory/raw-materials/orders/${id}`),
  approvePOBySPV: (id: string, data?: { notes?: string }) =>
    apiClient.post(`/factory/raw-materials/orders/${id}/approve-spv`, data),
  approvePOByPurchasing: (id: string, data?: { notes?: string }) =>
    apiClient.post(`/factory/raw-materials/orders/${id}/approve-purchasing`, data),
  updatePOProgress: (id: string, data: { progress_status: string; progress_notes?: string; estimated_arrival?: string }) =>
    apiClient.put(`/factory/raw-materials/orders/${id}/progress`, data),
  updatePOPayment: (id: string, data: { payment_status: string; payment_notes?: string }) =>
    apiClient.put(`/factory/raw-materials/orders/${id}/payment`, data),
  rejectRawMaterialPO: (id: string, data: { reason: string }) =>
    apiClient.post(`/factory/raw-materials/orders/${id}/reject`, data),

  // Raw Material Supplier Receipts
  getRawMaterialReceipts: (params?: { factory_id?: string }) =>
    apiClient.get('/factory/raw-materials/receipts', { params }),
  getRawMaterialReceiptDetail: (id: string) =>
    apiClient.get(`/factory/raw-materials/receipts/${id}`),
  createRawMaterialReceipt: (data: any) =>
    apiClient.post('/factory/raw-materials/receipts', data),

  // Technician / Maintenance Items & Spareparts
  getTechnicianItems: (factoryId: string, params?: { category?: string; low_stock?: boolean; search?: string }) =>
    apiClient.get(`/factory/${factoryId}/technician-items`, { params }),
  getTechnicianItem: (factoryId: string, itemId: string) =>
    apiClient.get(`/factory/${factoryId}/technician-items/${itemId}`),
  createTechnicianItem: (factoryId: string, data: any) =>
    apiClient.post(`/factory/${factoryId}/technician-items`, data),
  updateTechnicianItem: (factoryId: string, itemId: string, data: any) =>
    apiClient.put(`/factory/${factoryId}/technician-items/${itemId}`, data),
  deleteTechnicianItem: (factoryId: string, itemId: string) =>
    apiClient.delete(`/factory/${factoryId}/technician-items/${itemId}`),
  opnameTechnicianItem: (factoryId: string, itemId: string, data: { physical_quantity: number; reason: string; notes?: string }) =>
    apiClient.post(`/factory/${factoryId}/technician-items/${itemId}/opname`, data),
  batchOpnameTechnicianItems: (factoryId: string, data: { items: Array<{ technician_item_id: string; physical_quantity: number; reason?: string; notes?: string }> }) =>
    apiClient.post(`/factory/${factoryId}/technician-items/batch-opname`, data),
  recordTechnicianUsage: (factoryId: string, data: { technician_name: string; machine_or_purpose: string; usage_date?: string; items: Array<{ technician_item_id: string; quantity: number; notes?: string }>; notes?: string }) =>
    apiClient.post(`/factory/${factoryId}/technician-items/usage`, data),
  getTechnicianStockLogs: (factoryId: string, params?: { item_id?: string; type?: string }) =>
    apiClient.get(`/factory/${factoryId}/technician-items/logs`, { params }),

  // QC Samples
  getQcSamples: (factoryId: string, params?: { status?: string }) =>
    apiClient.get(`/factory/${factoryId}/qc-samples`, { params }),
  getAllQcSamples: (params?: { company_id?: string; status?: string }) =>
    apiClient.get('/factory/qc-samples', { params }),
  getQcSampleDetail: (id: string) => apiClient.get(`/factory/qc-samples/${id}`),
  requestQcSample: (factoryId: string, data: any) =>
    apiClient.post(`/factory/${factoryId}/qc-samples`, data),
  approveQcSample: (id: string) => apiClient.post(`/factory/qc-samples/${id}/approve`),
  rejectQcSample: (id: string, data: { reason: string }) =>
    apiClient.post(`/factory/qc-samples/${id}/reject`, data),
}

