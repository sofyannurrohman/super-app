import { defineStore } from 'pinia'
import apiClient from '@/api/axios'

export interface SalesVisit {
  id: string
  employee_id: string
  store_id: string
  check_in_time: string
  check_out_time: string | null
  latitude: number
  longitude: number
  selfie_url: string
  notes: string
  type: string
  created_at: string
  employee?: {
    first_name: string
    last_name: string
    user?: {
      email: string
    }
    branch?: {
      name: string
    }
  }
  store?: {
    name: string
    address: string
    category?: string
    latitude?: number
    longitude?: number
  }
  transactions?: any[]
  sales_orders?: any[]
}

export const useSalesVisitStore = defineStore('salesVisit', {
  state: () => ({
    visits: [] as SalesVisit[],
    total: 0,
    loading: false,
    stats: {
      total_today: 0,
      unique_stores: 0,
    },
  }),
  actions: {
    async fetchVisits(params?: any, config?: any) {
      this.loading = true
      try {
        const reqConfig = config ? { params, ...config } : { params }
        const res = await apiClient.get('/admin/sales/visits', reqConfig)
        this.visits = res.data.data || []
        this.total = res.data.total || this.visits.length
      } catch (error: any) {
        if (error.name !== 'CanceledError') {
          console.error('Failed to fetch visits:', error)
        }
      } finally {
        this.loading = false
      }
    },
    async createVisit(data: any) {
      try {
        await apiClient.post('/admin/sales/visits', data)
        return true
      } catch (error) {
        console.error('Failed to create visit:', error)
        throw error
      }
    },
    async createBatchVisits(visits: any[]) {
      try {
        const res = await apiClient.post('/admin/sales/visits/batch', { visits })
        return res.data
      } catch (error) {
        console.error('Failed to create batch visits:', error)
        throw error
      }
    },
    async updateVisit(id: string, data: any) {
      try {
        await apiClient.put(`/admin/sales/visits/${id}`, data)
        return true
      } catch (error) {
        console.error('Failed to update visit:', error)
        throw error
      }
    },
    async deleteVisit(id: string) {
      try {
        await apiClient.delete(`/admin/sales/visits/${id}`)
        return true
      } catch (error) {
        console.error('Failed to delete visit:', error)
        throw error
      }
    },
  },
})
