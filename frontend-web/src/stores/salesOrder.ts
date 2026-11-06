import { defineStore } from 'pinia'
import apiClient from '@/api/axios'
import { useMasterDataStore } from './masterData'

export type SOStatus =
  | 'DRAFT'
  | 'WAITING_TAX_DOCS'
  | 'CONFIRMED'
  | 'APPROVED'
  | 'IN_DELIVERY'
  | 'DELIVERED'
  | 'PAID'
  | 'WAITING_WAREHOUSE'
  | 'WAITING_STOCK'
  | 'PROCESSING'
  | 'SHIPPED'
  | 'CONVERTED'
  | 'CANCELLED'
  | 'REJECTED'

export interface SOItem {
  id: string
  sales_order_id: string
  product_id: string
  quantity: number
  ordered_quantity: number
  unit: string
  pieces_per_unit: number
  reserved_quantity: number
  actual_quantity: number
  price: number
  discount_percent?: number
  subtotal: number
  product?: {
    id: string
    name: string
    sku: string
    unit: string
  }
  batches?: {
    id: string
    sales_order_item_id: string
    batch_no: string
    quantity: number
  }[]
}

export interface SalesOrder {
  id: string
  so_number: string
  delivery_order_no?: string
  shipped_at?: string
  pod_image_url?: string
  received_at?: string
  received_by?: string
  payment_collected_amount?: number
  payment_method?: string
  branch_id: string
  company_id: string
  employee_id: string
  store_id: string
  store_category: string
  status: SOStatus
  total_amount: number
  discount_amount?: number
  notes: string
  invoice_id?: string
  delivery_batch_id?: string
  confirmed_at?: string
  converted_at?: string
  rejected_at?: string
  reject_notes?: string
  order_date: string
  created_at: string
  items: SOItem[]
  bonus_items?: {
    id: string
    bonus_inventory_id: string
    quantity: number
    notes: string
    bonus_inventory?: { id: string; name: string; unit: string }
  }[]
  is_mt?: boolean
  factory_company_id?: string
  employee?: { first_name: string; last_name: string; job_position?: { title: string } }
  store?: { name: string; address: string }
  confirmed_by?: { first_name: string; last_name: string; job_position?: { title: string } }
  admin_nota_confirmed_at?: string
  admin_nota_confirmed_by_id?: string
  admin_nota_confirmed_by?: { first_name: string; last_name: string; job_position?: { title: string } }
  // Modern Trade tax approval files
  billing_note_url?: string
  tax_invoice_url?: string
  tax_invoice_number?: string
  tax_invoice_date?: string
  delivery_note_url?: string
  destination_branch_id?: string
  is_physical_nota?: boolean
  physical_nota_image_url?: string
  deleted_at?: string
  deleted_by_id?: string
  deletion_reason?: string
  deleted_by?: { first_name: string; last_name: string; job_position?: { title: string } }
}

export const useSalesOrderStore = defineStore('salesOrder', {
  state: () => ({
    orders: [] as SalesOrder[],
    currentOrder: null as SalesOrder | null,
    loading: false,
    error: null as string | null,
    totalOrders: 0,
    currentPage: 1,
    limit: 10,
    summary: {
      total_approved_amount: 0,
      total_approved_count: 0,
      total_pending_amount: 0,
      total_pending_count: 0,
      total_rejected_amount: 0,
      total_rejected_count: 0,
      total_waiting_stock_amount: 0,
      total_waiting_stock_count: 0,
      total_filtered_amount: 0,
      total_filtered_count: 0,
    },
    pendingStats: {
      total_pending: 0,
      confirmed_count: 0,
      waiting_stock_count: 0,
      draft_count: 0,
      other_pending_count: 0,
      total_pending_amount: 0,
    },
  }),

  getters: {
    draftOrders: (state) => state.orders.filter((o) => o.status === 'DRAFT'),
    waitingWarehouseOrders: (state) => state.orders.filter((o) => o.status === 'WAITING_WAREHOUSE'),
    shippedOrders: (state) => state.orders.filter((o) => o.status === 'SHIPPED'),
    convertedOrders: (state) => state.orders.filter((o) => o.status === 'CONVERTED'),
  },

  actions: {
    async fetchByBranch(
      branchID: string,
      status = 'ALL',
      page = 1,
      limit = 0,
      extraParams?: {
        employee_id?: string
        company_id?: string
        date?: string
        month?: string
        search?: string
      },
    ) {
      const masterDataStore = useMasterDataStore()
      this.loading = true
      try {
        const params = new URLSearchParams()
        if (branchID) params.append('branch_id', branchID)
        if (status) params.append('status', status)
        if (masterDataStore.selectedSalesType) {
          params.append('sales_type', masterDataStore.selectedSalesType)
        }
        if (page) params.append('page', page.toString())
        if (limit !== undefined) params.append('limit', limit.toString())

        if (extraParams?.employee_id && extraParams.employee_id !== 'ALL') {
          params.append('employee_id', extraParams.employee_id)
        }
        if (extraParams?.company_id && extraParams.company_id !== 'ALL') {
          params.append('company_id', extraParams.company_id)
        }
        if (extraParams?.date) {
          params.append('date', extraParams.date)
        } else if (extraParams?.month && extraParams.month !== 'ALL') {
          params.append('month', extraParams.month)
        }
        if (extraParams?.search?.trim()) {
          params.append('search', extraParams.search.trim())
        }

        const res = await apiClient.get(`/admin/sales-orders?${params.toString()}`)
        this.orders = res.data.data || []
        this.totalOrders = res.data.meta?.total || 0
        this.currentPage = res.data.meta?.page || 1
        this.limit = res.data.meta?.limit || 10
        if (res.data.meta?.summary) {
          this.summary = res.data.meta.summary
        }
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Gagal memuat pesanan'
      } finally {
        this.loading = false
      }
    },

    async fetchByID(id: string) {
      const res = await apiClient.get(`/admin/sales-orders/${id}`)
      this.currentOrder = res.data.data
      return this.currentOrder
    },

    async createSO(data: any): Promise<SalesOrder> {
      const res = await apiClient.post('/admin/sales-orders/manual', data)
      await this.fetchByBranch(data.branch_id)
      return res.data.data
    },

    async confirmSO(id: string): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/confirm`)
      await this._refreshOrder(id)
    },

    /** Admin Nota: Verifikasi nota dari salesman → SO status CONFIRMED */
    async adminConfirmSO(id: string): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/admin-confirm`)
      await this._refreshOrder(id)
    },

    /** Admin Pajak / Tax approval: Approve PO MT & upload dokumen perpajakan */
    async approveTaxDocuments(
      id: string,
      data: {
        billing_note_url: string
        tax_invoice_url: string
        delivery_note_url: string
        destination_branch_id?: string
      },
    ): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/tax-approve`, data)
      await this._refreshOrder(id)
    },

    /** Admin Nota: Tolak nota dari salesman */
    async adminRejectSO(id: string, notes: string): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/admin-reject`, { notes })
      await this._refreshOrder(id)
    },

    async processByWarehouse(id: string, items: any[]): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/process-warehouse`, {
        so_id: id,
        items,
      })
      await this._refreshOrder(id)
    },

    async confirmPOD(
      id: string,
      data: { received_by: string; pod_image_url?: string },
    ): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/confirm-pod`, {
        so_id: id,
        ...data,
      })
      await this._refreshOrder(id)
    },

    async rejectSO(id: string, notes: string): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/reject`, { notes })
      await this._refreshOrder(id)
    },

    async cancelSO(id: string): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/cancel`)
      await this._refreshOrder(id)
    },

    async convertToInvoice(id: string, companyID: string): Promise<any> {
      const res = await apiClient.post(`/admin/sales-orders/${id}/convert`, {
        company_id: companyID,
      })
      await this._refreshOrder(id)
      return res.data.data
    },

    async convertDirectToInvoice(id: string): Promise<any> {
      const res = await apiClient.post(`/admin/sales-orders/${id}/convert-direct`)
      await this._refreshOrder(id)
      return res.data.data
    },

    async deleteSO(id: string, branchID: string, reason = ''): Promise<void> {
      await apiClient.delete(`/admin/sales-orders/${id}`, { data: { reason } })
      await this.fetchByBranch(branchID)
    },

    async restoreSO(id: string, branchID: string): Promise<void> {
      await apiClient.post(`/admin/sales-orders/${id}/restore`)
      await this.fetchByBranch(branchID)
    },

    async adjustSO(
      id: string,
      data: {
        discount_amount?: number
        adjust_notes?: string
        items?: any[]
        bonus_items?: any[]
      },
    ): Promise<SalesOrder> {
      const res = await apiClient.patch(`/admin/sales-orders/${id}/adjust`, data)
      await this._refreshOrder(id)
      return res.data.data
    },

    async fetchStoreDebtHistory(storeId: string): Promise<any[]> {
      const res = await apiClient.get(`/admin/sales-orders/store/${storeId}/debt`)
      return res.data.data || []
    },

    async overrideBackorder(id: string): Promise<void> {
      await apiClient.patch(`/admin/sales-orders/${id}/override-backorder`)
      await this._refreshOrder(id)
    },

    async recheckStock(id: string): Promise<{ message: string; data: SalesOrder }> {
      const res = await apiClient.post(`/admin/sales-orders/${id}/recheck-stock`)
      await this._refreshOrder(id)
      return res.data
    },

    /** Super Admin: Selesaikan pesanan SO secara instan (1-klik fast-track complete) */
    async superAdminCompleteSO(id: string): Promise<SalesOrder> {
      const res = await apiClient.post(`/admin/sales-orders/${id}/superadmin-complete`)
      await this._refreshOrder(id)
      return res.data.data
    },

    /** Super Admin: Selesaikan SEMUA pesanan SO yang tertahan secara massal (per cabang / seluruh cabang) */
    async superAdminBulkCompleteSO(params: {
      branch_id?: string
      company_id?: string
      so_ids?: string[]
    }): Promise<any> {
      const payload: any = {}
      if (params.branch_id && params.branch_id !== 'ALL_BRANCHES') {
        payload.branch_id = params.branch_id
      }
      if (params.company_id && params.company_id !== 'ALL') {
        payload.company_id = params.company_id
      }
      if (params.so_ids && params.so_ids.length > 0) {
        payload.so_ids = params.so_ids
      }

      const res = await apiClient.post('/admin/sales-orders/bulk-superadmin-complete', payload)
      return res.data
    },

    /** Ambil statistik riil pesanan tertahan langsung dari database (semua halaman & riwayat bulan lalu) */
    async fetchPendingStats(branchId?: string, companyId?: string): Promise<any> {
      try {
        const params = new URLSearchParams()
        if (branchId && branchId !== 'ALL_BRANCHES') params.append('branch_id', branchId)
        if (companyId && companyId !== 'ALL') params.append('company_id', companyId)
        const res = await apiClient.get(`/admin/sales-orders/pending-stats?${params.toString()}`)
        if (res.data?.data) {
          this.pendingStats = res.data.data
        }
        return this.pendingStats
      } catch (e) {
        console.warn('Failed to fetch pending SO stats:', e)
        return null
      }
    },

    async _refreshOrder(id: string) {
      try {
        const res = await apiClient.get(`/admin/sales-orders/${id}`)
        const updated = res.data.data as SalesOrder
        const idx = this.orders.findIndex((o) => o.id === id)
        if (idx !== -1) this.orders[idx] = updated
        this.currentOrder = updated

        // Sinkronkan data ke deliveryStore jika ada di cache waiting list pengiriman
        try {
          const { useDeliveryStore } = await import('./delivery')
          const deliveryStore = useDeliveryStore()
          const cIdx = deliveryStore.confirmedOrders.findIndex((o: any) => o.id === id)
          if (cIdx !== -1) {
            deliveryStore.confirmedOrders[cIdx] = updated
          }
        } catch {}
      } catch {}
    },
  },
})
