import { defineStore } from 'pinia'
import { warehouseApi } from '@/api/warehouse'
import { factoryReturnApi } from '@/api/factoryReturn'
import { factoryApi } from '@/api/factory'
import { useMasterDataStore } from './masterData'
import apiClient from '@/api/axios'

export interface WarehouseStock {
  id: string
  branch_id: string
  product_id: string
  batch_no: string
  expiry_date?: string
  quantity: number
  reserved_quantity: number
  min_limit: number
  stock_type?: string
  updated_at: string
  product?: {
    id: string
    name: string
    sku: string
    unit: string
    weight: number
  }
}

export interface ProductTransfer {
  id: string
  from_factory_id: string
  to_branch_id: string
  fulfillment_branch_id?: string | null
  product_id: string
  quantity: number
  total_weight: number
  total_volume?: number
  status: 'REQUESTED' | 'APPROVED' | 'LOADING' | 'SHIPPED' | 'ARRIVED' | 'RECEIVED' | 'REJECTED'
  batch_no: string
  expiry_date?: string
  notes: string
  target_shipment_date?: string
  estimated_arrival?: string
  shipped_at?: string
  rejection_reason?: string
  initiated_by?: string
  created_at: string
  product?: {
    name: string
    unit: string
  }
  from_factory?: {
    name: string
  }
  fulfillment_branch?: {
    name: string
  }
  to_branch?: {
    name: string
  }
  delivery_order_no?: string
  received_at?: string
  vehicle?: {
    name: string
    plate_number: string
  }
  driver?: {
    name: string
  }
}

export const useWarehouseStore = defineStore('warehouse', {
  state: () => ({
    inventory: [] as WarehouseStock[],
    bonusInventory: [] as any[],
    pendingShipments: [] as ProductTransfer[],
    allTransfers: [] as ProductTransfer[],
    logs: [] as any[],
    factoryReturns: [] as any[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchAllTransfers() {
      const masterDataStore = useMasterDataStore()
      try {
        const response = await warehouseApi.getAllTransfers(masterDataStore.selectedWarehouseBranchId)
        this.allTransfers = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch all transfers'
      }
    },
    async fetchInventory() {
      const masterDataStore = useMasterDataStore()
      this.loading = true
      try {
        const response = await warehouseApi.getInventory(masterDataStore.selectedWarehouseBranchId)
        this.inventory = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch inventory'
      } finally {
        this.loading = false
      }
    },
    async fetchBonusInventory() {
      this.loading = true
      try {
        const response = await apiClient.get('/bonus-inventory')
        this.bonusInventory = response.data.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch bonus inventory'
      } finally {
        this.loading = false
      }
    },
    async fetchGlobalBranchInventory() {
      this.loading = true
      try {
        const response = await warehouseApi.getInventory('ALL_BRANCHES')
        return response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch global branch inventory'
        return []
      } finally {
        this.loading = false
      }
    },

    async fetchPendingShipments() {
      const masterDataStore = useMasterDataStore()
      try {
        const response = await warehouseApi.getPendingShipments(masterDataStore.selectedWarehouseBranchId)
        this.pendingShipments = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch shipments'
      }
    },

    async fetchLogs() {
      const masterDataStore = useMasterDataStore()
      try {
        const response = await warehouseApi.getLogs(masterDataStore.selectedWarehouseBranchId)
        this.logs = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch logs'
      }
    },

    async receiveShipment(id: string, data: any) {
      try {
        await warehouseApi.receiveShipment(id, data)
        await this.fetchInventory()
        await this.fetchPendingShipments()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to receive shipment'
      }
    },

    async correctShipmentReceipt(id: string, data: any) {
      try {
        await warehouseApi.correctShipmentReceipt(id, data)
        await this.fetchInventory()
        await this.fetchAllTransfers()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to correct shipment receipt'
      }
    },

    async arriveShipment(id: string) {
      try {
        await warehouseApi.arriveShipment(id)
        await this.fetchPendingShipments()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to confirm arrival'
      }
    },

    async approveShipment(id: string) {
      try {
        await warehouseApi.approveShipment(id)
        await this.fetchPendingShipments()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to approve shipment'
      }
    },

    async rejectShipment(id: string, reason: string) {
      try {
        await warehouseApi.rejectShipment(id, reason)
        await this.fetchPendingShipments()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to reject shipment'
      }
    },
    async deleteTransfer(id: string) {
      try {
        await factoryApi.deleteTransfer(id)
        await this.fetchPendingShipments()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to delete transfer'
      }
    },
    async confirmShipment(id: string) {
      try {
        await warehouseApi.confirmShipment(id)
        await this.fetchPendingShipments()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to confirm shipment'
      }
    },

    async setStockLimit(data: {
      product_id: string
      batch_no?: string
      limit: number
      branch_id?: string | null
    }) {
      try {
        await warehouseApi.setStockLimit(data)
        await this.fetchInventory()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to set stock limit'
      }
    },

    async adjustStock(data: any) {
      try {
        await warehouseApi.adjustStock(data)
        await this.fetchInventory()
        await this.fetchLogs()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to adjust stock'
      }
    },

    async deleteStock(stockId: string, reason: string) {
      const masterDataStore = useMasterDataStore()
      try {
        await warehouseApi.deleteStock(stockId, {
          reason,
          branch_id: masterDataStore.selectedWarehouseBranchId || undefined,
        })
        await this.fetchInventory()
        await this.fetchLogs()
      } catch (err: any) {
        throw err.response?.data?.error || 'Gagal menghapus data stok'
      }
    },

    async receiveByDO(doNo: string, data: any) {
      try {
        await warehouseApi.receiveByDO(doNo, data)
        await this.fetchInventory()
        await this.fetchPendingShipments()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to receive by DO'
      }
    },
    async requestGoods(factoryId: string, data: any) {
      try {
        await warehouseApi.requestGoods(factoryId, data)
        await this.fetchPendingShipments()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to request goods'
      }
    },
    async fetchFactoryReturns() {
      const masterDataStore = useMasterDataStore()
      if (!masterDataStore.selectedWarehouseBranchId) return
      try {
        const response = await factoryReturnApi.getByBranch(masterDataStore.selectedWarehouseBranchId)
        this.factoryReturns = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch factory returns'
      }
    },
    async createFactoryReturn(data: any) {
      try {
        await factoryReturnApi.createReturn(data)
        await this.fetchFactoryReturns()
        await this.fetchInventory() // Refresh inventory because stock is deducted
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to create factory return'
      }
    },
  },
})
