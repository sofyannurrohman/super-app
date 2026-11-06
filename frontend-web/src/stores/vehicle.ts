import { defineStore } from 'pinia'
import { vehicleApi } from '@/api/vehicle'

export interface ActiveDeliveryAssignment {
  delivery_type: 'SALES_DELIVERY' | 'FACTORY_TRANSFER' | 'MT_DELIVERY' | 'FACTORY_RETURN'
  type_label: string
  reference_id: string
  do_number: string
  status: string
  status_label: string
  driver_name: string
  driver_phone?: string
  helper_name?: string
  origin_name: string
  destination_name: string
  total_items?: number
  started_at?: string
  created_at: string
  redirect_path: string
}

export interface Vehicle {
  id: string
  company_id: string
  branch_id?: string
  name: string
  license_plate: string
  model: string
  type: string
  status: 'AVAILABLE' | 'IN_USE' | 'MAINTENANCE' | 'BROKEN'
  year: number
  mileage: number
  capacity: number
  bed_length?: number
  bed_width?: number
  bed_height?: number
  bed_volume?: number
  image_url?: string
  photo_front?: string
  photo_left?: string
  photo_right?: string
  photo_back?: string
  tax_expiry_date?: string
  branch?: {
    id: string
    name: string
  }
  active_assignment?: ActiveDeliveryAssignment
}

export const useVehicleStore = defineStore('vehicle', {
  state: () => ({
    vehicles: [] as Vehicle[],
    logs: [] as any[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchVehicles() {
      this.loading = true
      try {
        const response = await vehicleApi.getVehicles({ bypass_branch: true })
        this.vehicles = response.data
      } catch (err: any) {
        this.error = err.response?.data?.error || 'Failed to fetch vehicles'
      } finally {
        this.loading = false
      }
    },

    async addVehicle(data: FormData) {
      try {
        await vehicleApi.createVehicle(data)
        await this.fetchVehicles()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to add vehicle'
      }
    },

    async updateVehicle(id: string, data: FormData) {
      try {
        await vehicleApi.updateVehicle(id, data)
        await this.fetchVehicles()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to update vehicle'
      }
    },

    async removeVehicle(id: string) {
      try {
        await vehicleApi.deleteVehicle(id)
        await this.fetchVehicles()
      } catch (err: any) {
        throw err.response?.data?.error || 'Failed to delete vehicle'
      }
    },

    async fetchVehicleLogs(id: string) {
      try {
        const response = await vehicleApi.getVehicleLogs(id)
        this.logs = response.data
      } catch (err: any) {
        console.error('Failed to fetch vehicle logs:', err)
      }
    },

    async fetchVehicleActiveTask(id: string): Promise<ActiveDeliveryAssignment | null> {
      try {
        const response = await vehicleApi.getActiveTask(id)
        if (response.data?.active && response.data?.task) {
          return response.data.task
        }
        return null
      } catch (err: any) {
        console.error('Failed to fetch vehicle active task:', err)
        return null
      }
    },
  },
})
