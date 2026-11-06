import { defineStore } from 'pinia'
import apiClient from '../api/axios'
import { useMasterDataStore } from './masterData'
import { useAuthStore } from './auth'
import { useLayoutStore } from './layout'

export const useDeliveryStore = defineStore('delivery', {
  state: () => ({
    batches: [] as any[],
    // Antrean nota yang sudah diverifikasi Admin Nota, siap dibatch oleh Supervisor
    confirmedOrders: [] as any[],
    // Legacy: pending transactions (dipertahankan agar tidak break kode lama)
    pendingOrders: [] as any[],
    // Nota yang sudah terkirim tapi masih menunggak pembayaran (Piutang)
    unpaidOrders: [] as any[],
    // Faktur Tempo (SalesTransaction) yang belum lunas – untuk assign ke Surat Jalan
    outstandingInvoices: [] as any[],
    // Permintaan Sampel yang sudah APPROVED
    approvedSamples: [] as any[],
    loading: false,
  }),
  actions: {
    async fetchConfirmedOrders(branchId: string = '') {
      const masterDataStore = useMasterDataStore()
      const authStore = useAuthStore()
      const layoutStore = useLayoutStore()
      this.loading = true
      try {
        let salesType = masterDataStore.selectedSalesType
        if (authStore.user?.role === 'admin_gudang' || layoutStore.currentModule === 'warehouse') {
          salesType = 'ALL_SALES'
        }
        const url = branchId
          ? `/admin/sales-orders?branch_id=${branchId}&status=CONFIRMED&sales_type=${salesType}`
          : `/admin/sales-orders?status=CONFIRMED&sales_type=${salesType}`
        const response = await apiClient.get(url)
        this.confirmedOrders = response.data?.data || []
        // Sync ke pendingOrders agar kode lama yang masih pakai pendingOrders tidak break
        this.pendingOrders = this.confirmedOrders
      } finally {
        this.loading = false
      }
    },
    /** Legacy: dipertahankan agar kode lama tidak break */
    async fetchPendingOrders(companyId: string = '') {
      return this.fetchConfirmedOrders()
    },
    async fetchUnpaidOrders(branchId: string = '') {
      const masterDataStore = useMasterDataStore()
      const authStore = useAuthStore()
      const layoutStore = useLayoutStore()
      this.loading = true
      try {
        let salesType = masterDataStore.selectedSalesType
        if (authStore.user?.role === 'admin_gudang' || layoutStore.currentModule === 'warehouse') {
          salesType = 'ALL_SALES'
        }
        // Ambil SO dengan status DELIVERED yang payment_status nya belum lunas
        const url = branchId
          ? `/admin/sales-orders?branch_id=${branchId}&status=DELIVERED&payment_status=UNPAID,PARTIAL&sales_type=${salesType}`
          : `/admin/sales-orders?status=DELIVERED&payment_status=UNPAID,PARTIAL&sales_type=${salesType}`
        const response = await apiClient.get(url)
        this.unpaidOrders = response.data?.data || []
      } finally {
        this.loading = false
      }
    },
    async fetchApprovedSamples(branchId: string = '') {
      this.loading = true
      try {
        const url = branchId && branchId !== 'ALL_BRANCHES'
          ? `/admin/sales-tester-requests?branch_id=${branchId}&status=APPROVED`
          : `/admin/sales-tester-requests?status=APPROVED`
        const response = await apiClient.get(url)
        this.approvedSamples = response.data?.data || []
      } finally {
        this.loading = false
      }
    },
    async createBatch(data: any) {
      const response = await apiClient.post('/delivery/batch', data)
      return response.data
    },
    async fetchOutstandingInvoices(branchId: string = '') {
      const masterDataStore = useMasterDataStore()
      const authStore = useAuthStore()
      const layoutStore = useLayoutStore()
      this.loading = true
      try {
        let salesType = masterDataStore.selectedSalesType
        if (authStore.user?.role === 'admin_gudang' || layoutStore.currentModule === 'warehouse') {
          salesType = 'ALL_SALES'
        }
        const params = new URLSearchParams()
        if (branchId && branchId !== 'ALL_BRANCHES') params.set('branch_id', branchId)
        if (salesType && salesType !== 'ALL_SALES') params.set('sales_type', salesType)
        const response = await apiClient.get(`/admin/sales/transactions/outstanding?${params.toString()}`)
        this.outstandingInvoices = response.data?.data || []
      } finally {
        this.loading = false
      }
    },
    async fetchBatches(branchId: string = '') {
      const masterDataStore = useMasterDataStore()
      const authStore = useAuthStore()
      const layoutStore = useLayoutStore()
      
      let salesType = masterDataStore.selectedSalesType
      if (authStore.user?.role === 'admin_gudang' || layoutStore.currentModule === 'warehouse') {
        salesType = 'ALL_SALES'
      }

      const url =
        branchId && branchId !== 'ALL_BRANCHES'
          ? `/delivery/batches?branch_id=${branchId}&sales_type=${salesType}`
          : `/delivery/batches?sales_type=${salesType}`
      // Clear dulu agar Vue trigger reactivity re-render
      this.batches = []
      const response = await apiClient.get(url)
      this.batches = response.data?.data || []
    },
    async fetchEligibleOrdersForBatch(batchId: string, search: string = '') {
      const params = new URLSearchParams()
      if (search) params.set('search', search)
      const response = await apiClient.get(`/delivery/batch/${batchId}/eligible-orders?${params.toString()}`)
      return response.data?.data || []
    },
    async appendOrdersToBatch(batchId: string, payload: { sales_order_ids: string[]; notes?: string }) {
      const response = await apiClient.post(`/delivery/batch/${batchId}/append-orders`, payload)
      return response.data
    },
  },
})

