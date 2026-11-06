import { defineStore } from 'pinia'
import apiClient from '@/api/axios'

export type ReturnStatus = 'PENDING' | 'PENDING_WAREHOUSE' | 'APPROVED' | 'REJECTED'

export interface SalesReturn {
  id: string
  return_no: string
  sales_order_id?: string
  transaction_id?: string
  branch_id: string
  employee_id: string
  return_type?: string
  status: ReturnStatus
  total_amount: number
  price_difference?: number
  invoice_adjusted_at?: string
  notes: string
  approved_at?: string
  created_at: string
  items: any[]
  employee?: { first_name: string; last_name?: string }
  approved_by?: { first_name: string; last_name?: string; job_position?: { title: string } }
  sales_order?: {
    so_number: string
    store?: { name: string }
    delivery_batch?: {
      delivery_order_no?: string
      driver?: { first_name?: string; last_name?: string }
      vehicle?: { plate_number?: string }
    }
  }
  transaction?: { receipt_no: string; buyer_name?: string; total_amount?: number; paid_amount?: number; store?: { name: string } }
}

export const useSalesReturnStore = defineStore('salesReturn', {
  state: () => ({
    returns: [] as SalesReturn[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchByBranch(branchID: string, status = 'ALL') {
      this.loading = true
      try {
        const res = await apiClient.get(
          `/admin/sales-returns?branch_id=${branchID}&status=${status}`,
        )
        this.returns = res.data.data || []
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Gagal memuat data retur'
      } finally {
        this.loading = false
      }
    },

    async createReturn(data: any) {
      const res = await apiClient.post('/admin/sales-returns', data)
      return res.data.data
    },

    async approveReturnNota(id: string) {
      await apiClient.patch(`/admin/sales-returns/${id}/approve-nota`)
      const branchID = this.returns.find((r) => r.id === id)?.branch_id
      if (branchID) await this.fetchByBranch(branchID)
    },

    async approveReturnWarehouse(id: string) {
      await apiClient.patch(`/admin/sales-returns/${id}/approve-warehouse`)
      const branchID = this.returns.find((r) => r.id === id)?.branch_id
      if (branchID) await this.fetchByBranch(branchID)
    },

    // Kept for backward compatibility if needed, but redirects to nota by default
    async approveReturn(id: string) {
      return this.approveReturnNota(id)
    },

    async updateReturn(id: string, data: any) {
      const res = await apiClient.put(`/admin/sales-returns/${id}`, data)
      const branchID = this.returns.find((r) => r.id === id)?.branch_id
      if (branchID) await this.fetchByBranch(branchID)
      return res.data.data
    },

    async deleteReturn(id: string) {
      const branchID = this.returns.find((r) => r.id === id)?.branch_id
      await apiClient.delete(`/admin/sales-returns/${id}`)
      if (branchID) await this.fetchByBranch(branchID)
    },
  },
})
