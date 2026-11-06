import { defineStore } from 'pinia'
import apiClient from '@/api/axios'
import { useMasterDataStore } from './masterData'

export interface SalesStock {
  id: string
  employee_id: string
  product_id?: string
  bonus_inventory_id?: string
  quantity: number
  updated_at: string
  product?: {
    name: string
    unit: string
    sku: string
  }
  bonus_inventory?: {
    name: string
    unit: string
  }
}

export interface SalesTransferCorrection {
  id: string
  sales_transfer_id: string
  corrected_by: string
  old_quantity: number
  new_quantity: number
  old_damaged_qty: number
  new_damaged_qty: number
  correction_reason: string
  created_at: string
}

export interface SalesTransfer {
  id: string
  branch_id: string
  employee_id: string
  product_id?: string
  bonus_inventory_id?: string
  batch_no?: string
  quantity: number
  actual_quantity: number
  damaged_quantity: number
  type: 'TRANSFER' | 'RETURN'
  status: 'WAITING_SUPERVISOR' | 'PENDING' | 'COMPLETED' | 'CANCELLED'
  reference_no: string
  notes: string
  transfer_date: string
  employee?: {
    first_name: string
    last_name: string
    employee_id_number?: string
    job_position?: {
      title: string
    }
  }
  product?: {
    name: string
    unit: string
    sku?: string
  }
  bonus_inventory?: {
    name: string
    unit: string
  }
  corrections?: SalesTransferCorrection[]
  distributor_store_id?: string
  request_source?: 'SALES' | 'DISTRIBUTOR'
}

export const useSalesTransferStore = defineStore('salesTransfer', {
  state: () => ({
    transfers: [] as SalesTransfer[],
    salesStock: [] as SalesStock[],
    bonusRequests: [] as any[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchTransfers(branchId?: string) {
      const masterDataStore = useMasterDataStore()
      this.loading = true
      try {
        const id = branchId || masterDataStore.selectedWarehouseBranchId
        const response = await apiClient.get(
          `/sales-transfers/branch/${id}`,
        )
        this.transfers = response.data.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch transfers'
      } finally {
        this.loading = false
      }
    },

    async fetchSupervisorPending(branchId?: string, requestSource?: 'SALES' | 'DISTRIBUTOR') {
      this.loading = true
      try {
        const response = await apiClient.get('/sales-transfers/supervisor/pending', {
          params: {
            branch_id: branchId,
            request_source: requestSource,
          },
        })
        return response.data.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch supervisor pending transfers'
        throw err
      } finally {
        this.loading = false
      }
    },

    async fetchMyDistributorRequests() {
      this.loading = true
      try {
        const response = await apiClient.get('/sales-transfers/distributor/my-requests')
        return response.data.data as SalesTransfer[]
      } finally {
        this.loading = false
      }
    },

    async createDistributorRequest(data: any) {
      this.loading = true
      try {
        const response = await apiClient.post('/sales-transfers/distributor/request', {
          ...data,
          request_source: 'DISTRIBUTOR',
        })
        return response.data
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to create distributor request'
      } finally {
        this.loading = false
      }
    },

    async fetchSalesStock(employeeId: string) {
      try {
        const response = await apiClient.get(`/sales-transfers/stock/${employeeId}`)
        this.salesStock = response.data.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch sales stock'
      }
    },

    async fetchBonusRequests(branchId?: string) {
      const masterDataStore = useMasterDataStore()
      try {
        const id = branchId || masterDataStore.selectedWarehouseBranchId
        const params = id ? { branch_id: id } : {}
        const response = await apiClient.get('/sales-bonus-requests/pending', { params })
        this.bonusRequests = response.data.data || []
      } catch (err: any) {
        console.error('Failed to fetch bonus requests:', err)
      }
    },

    async approveBonusRequest(id: string) {
      try {
        await apiClient.patch(`/sales-bonus-requests/${id}/approve`)
        await this.fetchBonusRequests()
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to approve bonus request'
      }
    },

    async rejectBonusRequest(id: string) {
      try {
        await apiClient.patch(`/sales-bonus-requests/${id}/reject`)
        await this.fetchBonusRequests()
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to reject bonus request'
      }
    },

    async createTransfer(data: any, branchId?: string) {
      const masterDataStore = useMasterDataStore()
      try {
        const id = branchId || masterDataStore.selectedWarehouseBranchId
        await apiClient.post(`/sales-transfers/branch/${id}`, data)
        await this.fetchTransfers(id)
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to create transfer'
      }
    },

    async createBatchTransfer(data: any, branchId?: string) {
      const masterDataStore = useMasterDataStore()
      try {
        const id = branchId || masterDataStore.selectedWarehouseBranchId || masterDataStore.selectedBranchId
        let url = '/sales-transfers/batch'
        if (id && id !== 'ALL_BRANCHES') {
          url = `/sales-transfers/branch/${id}/batch`
        }
        const res = await apiClient.post(url, data)
        return res.data
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to create batch transfer'
      }
    },

    async completeTransfer(id: string, branchId?: string) {
      try {
        await apiClient.patch(`/sales-transfers/${id}/complete`)
        await this.fetchTransfers(branchId)
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to complete transfer'
      }
    },

    async approveSupervisorTransfer(id: string, quantity?: number) {
      try {
        await apiClient.patch(`/sales-transfers/${id}/approve-supervisor`, quantity ? { quantity } : {})
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to approve transfer'
      }
    },

    async rejectSupervisorTransfer(id: string) {
      try {
        await apiClient.patch(`/sales-transfers/${id}/reject-supervisor`)
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to reject transfer'
      }
    },

    async cancelTransfer(id: string, branchId?: string) {
      try {
        await apiClient.patch(`/sales-transfers/${id}/cancel`)
        await this.fetchTransfers(branchId)
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to cancel transfer'
      }
    },

    async deleteTransfer(id: string, branchId?: string) {
      try {
        await apiClient.delete(`/sales-transfers/${id}`)
        await this.fetchTransfers(branchId)
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to delete transfer'
      }
    },

    async updateTransfer(id: string, data: any, branchId?: string) {
      try {
        await apiClient.put(`/sales-transfers/${id}`, data)
        await this.fetchTransfers(branchId)
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to update transfer'
      }
    },
    async generateDO(salesTransferIds: string[]) {
      try {
        // Endpoint baru: membuat DO Motoris standalone tanpa SJ/DeliveryBatch
        const res = await apiClient.post(`/admin/warehouse-delivery-orders/motoris`, {
          sales_transfer_ids: salesTransferIds
        })
        return res.data?.data
      } catch (err: any) {
        throw err.response?.data?.message || err.response?.data?.error || err.message || 'Failed to generate DO Motoris'
      }
    },
  },
})
