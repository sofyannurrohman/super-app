import { defineStore } from 'pinia'
import { ref } from 'vue'
import { rajaTransportApi } from '@/api/rajaTransport'

export const useRajaTransportStore = defineStore('rajaTransport', () => {
  const loading = ref(false)
  const error = ref<string | null>(null)

  const stats = ref<any>({
    active_trips_count: 0,
    completed_trips_count: 0,
    pending_backhauls_count: 0,
    approved_backhaul_fee_total: 0,
    pending_claims_count: 0,
    approved_claims_amount_total: 0,
    fleets_in_maintenance_count: 0,
  })

  const trips = ref<any[]>([])
  const currentTrip = ref<any | null>(null)

  const backhauls = ref<any[]>([])
  const currentBackhaul = ref<any | null>(null)

  const expenseClaims = ref<any[]>([])
  const currentExpenseClaim = ref<any | null>(null)

  const inspections = ref<any[]>([])
  const currentInspection = ref<any | null>(null)

  const attendances = ref<any[]>([])

  // ──────────────── STATS ────────────────
  async function fetchDashboardStats(params?: any) {
    try {
      const res = await rajaTransportApi.getDashboardStats(params)
      stats.value = res.data
    } catch (err: any) {
      console.error('Error fetching transport dashboard stats:', err)
    }
  }

  // ──────────────── TRIPS ────────────────
  async function fetchTrips(params?: any) {
    loading.value = true
    error.value = null
    try {
      const res = await rajaTransportApi.getTrips(params)
      trips.value = res.data || []
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message
    } finally {
      loading.value = false
    }
  }

  async function createTrip(data: any) {
    loading.value = true
    error.value = null
    try {
      const res = await rajaTransportApi.createTrip(data)
      await fetchTrips()
      await fetchDashboardStats()
      return res.data
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateTrip(id: string, data: any) {
    loading.value = true
    try {
      const res = await rajaTransportApi.updateTrip(id, data)
      await fetchTrips()
      return res.data
    } catch (err: any) {
      throw err
    } finally {
      loading.value = false
    }
  }

  async function completeTrip(id: string, data: { actual_return_date: string; end_odometer: number; notes?: string }) {
    loading.value = true
    try {
      const res = await rajaTransportApi.completeTrip(id, data)
      await fetchTrips()
      await fetchDashboardStats()
      return res.data
    } catch (err: any) {
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteTrip(id: string) {
    try {
      await rajaTransportApi.deleteTrip(id)
      await fetchTrips()
      await fetchDashboardStats()
    } catch (err: any) {
      throw err
    }
  }

  // ──────────────── BACKHAULS ────────────────
  async function fetchBackhauls(params?: any) {
    loading.value = true
    error.value = null
    try {
      const res = await rajaTransportApi.getBackhauls(params)
      backhauls.value = res.data || []
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message
    } finally {
      loading.value = false
    }
  }

  async function createBackhaul(data: FormData | any) {
    loading.value = true
    try {
      const res = await rajaTransportApi.createBackhaul(data)
      await fetchBackhauls()
      await fetchDashboardStats()
      return res.data
    } catch (err: any) {
      throw err
    } finally {
      loading.value = false
    }
  }

  async function approveBackhaul(id: string, data: { action: 'APPROVED' | 'REJECTED'; notes?: string; reason?: string }) {
    try {
      const res = await rajaTransportApi.approvalBackhaul(id, data)
      await fetchBackhauls()
      await fetchDashboardStats()
      return res.data
    } catch (err: any) {
      throw err
    }
  }

  async function deleteBackhaul(id: string) {
    try {
      await rajaTransportApi.deleteBackhaul(id)
      await fetchBackhauls()
      await fetchDashboardStats()
    } catch (err: any) {
      throw err
    }
  }

  // ──────────────── EXPENSE CLAIMS ────────────────
  async function fetchExpenseClaims(params?: any) {
    loading.value = true
    error.value = null
    try {
      const res = await rajaTransportApi.getExpenseClaims(params)
      expenseClaims.value = res.data || []
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message
    } finally {
      loading.value = false
    }
  }

  async function createExpenseClaim(data: FormData | any) {
    loading.value = true
    try {
      const res = await rajaTransportApi.createExpenseClaim(data)
      await fetchExpenseClaims()
      await fetchDashboardStats()
      return res.data
    } catch (err: any) {
      throw err
    } finally {
      loading.value = false
    }
  }

  async function approveExpenseClaim(id: string, data: { action: 'APPROVED' | 'REJECTED'; reason?: string }) {
    try {
      const res = await rajaTransportApi.approvalExpenseClaim(id, data)
      await fetchExpenseClaims()
      await fetchDashboardStats()
      return res.data
    } catch (err: any) {
      throw err
    }
  }

  async function disburseExpenseClaim(id: string) {
    try {
      const res = await rajaTransportApi.disburseExpenseClaim(id)
      await fetchExpenseClaims()
      return res.data
    } catch (err: any) {
      throw err
    }
  }

  async function deleteExpenseClaim(id: string) {
    try {
      await rajaTransportApi.deleteExpenseClaim(id)
      await fetchExpenseClaims()
      await fetchDashboardStats()
    } catch (err: any) {
      throw err
    }
  }

  // ──────────────── FLEET INSPECTIONS ────────────────
  async function fetchFleetInspections(params?: any) {
    loading.value = true
    error.value = null
    try {
      const res = await rajaTransportApi.getFleetInspections(params)
      inspections.value = res.data || []
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message
    } finally {
      loading.value = false
    }
  }

  async function createFleetInspection(data: FormData | any) {
    loading.value = true
    try {
      const res = await rajaTransportApi.createFleetInspection(data)
      await fetchFleetInspections()
      await fetchDashboardStats()
      return res.data
    } catch (err: any) {
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateInspectionFollowUp(id: string, data: any) {
    try {
      const res = await rajaTransportApi.updateInspectionFollowUp(id, data)
      await fetchFleetInspections()
      await fetchDashboardStats()
      return res.data
    } catch (err: any) {
      throw err
    }
  }

  async function deleteFleetInspection(id: string) {
    try {
      await rajaTransportApi.deleteFleetInspection(id)
      await fetchFleetInspections()
      await fetchDashboardStats()
    } catch (err: any) {
      throw err
    }
  }

  // ──────────────── DRIVER ATTENDANCES ────────────────
  async function fetchDriverAttendances(params?: any) {
    loading.value = true
    try {
      const res = await rajaTransportApi.getDriverAttendances(params)
      attendances.value = res.data || []
    } catch (err: any) {
      console.error(err)
    } finally {
      loading.value = false
    }
  }

  async function recordDriverAttendance(data: any) {
    try {
      const res = await rajaTransportApi.recordDriverAttendance(data)
      await fetchDriverAttendances()
      return res.data
    } catch (err: any) {
      throw err
    }
  }

  return {
    loading,
    error,
    stats,
    trips,
    currentTrip,
    backhauls,
    currentBackhaul,
    expenseClaims,
    currentExpenseClaim,
    inspections,
    currentInspection,
    attendances,
    // Actions
    fetchDashboardStats,
    fetchTrips,
    createTrip,
    updateTrip,
    completeTrip,
    deleteTrip,
    fetchBackhauls,
    createBackhaul,
    approveBackhaul,
    deleteBackhaul,
    fetchExpenseClaims,
    createExpenseClaim,
    approveExpenseClaim,
    disburseExpenseClaim,
    deleteExpenseClaim,
    fetchFleetInspections,
    createFleetInspection,
    updateInspectionFollowUp,
    deleteFleetInspection,
    fetchDriverAttendances,
    recordDriverAttendance,
  }
})
