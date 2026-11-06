import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  shopeeBoostApi,
  type ShopeeBoostOverview,
  type ShopeeAutoBoostItem,
  type ShopeeAutoBoostLog,
} from '@/api/shopeeBoost'

export const useShopeeBoostStore = defineStore('shopeeBoost', () => {
  const overview = ref<ShopeeBoostOverview | null>(null)
  const queue = ref<ShopeeAutoBoostItem[]>([])
  const queueTotal = ref(0)
  const logs = ref<ShopeeAutoBoostLog[]>([])
  const isLoading = ref(false)
  const isActionLoading = ref(false)
  const error = ref<string | null>(null)
  const lastUpdated = ref<Date | null>(null)

  const fetchOverview = async (branchId?: string) => {
    isLoading.value = true
    error.value = null
    try {
      const res = await shopeeBoostApi.getOverview(branchId)
      if (res.data.success) {
        overview.value = res.data.data
        if (res.data.data.recent_logs) {
          logs.value = res.data.data.recent_logs
        }
      } else {
        error.value = res.data.error || 'Gagal memuat status boost'
      }
      lastUpdated.value = new Date()
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message || 'Terjadi kesalahan sistem'
    } finally {
      isLoading.value = false
    }
  }

  const toggleAutoBoost = async (isEnabled: boolean, branchId?: string) => {
    isActionLoading.value = true
    try {
      await shopeeBoostApi.toggleAutoBoost({ branch_id: branchId, is_enabled: isEnabled })
      if (overview.value) {
        overview.value.is_enabled = isEnabled
        if (overview.value.settings) {
          overview.value.settings.is_enabled = isEnabled
        }
      }
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Gagal memperbarui sakelar auto boost')
    } finally {
      isActionLoading.value = false
    }
  }

  const fetchQueue = async (branchId?: string, search?: string) => {
    try {
      const res = await shopeeBoostApi.getQueue({ branch_id: branchId, search })
      if (res.data.success) {
        queue.value = res.data.data
        queueTotal.value = res.data.total
      }
    } catch (err: any) {
      console.error('Failed to fetch queue:', err)
    }
  }

  const addToQueue = async (itemIds: number[], branchId?: string) => {
    isActionLoading.value = true
    try {
      const res = await shopeeBoostApi.addToQueue({ branch_id: branchId, item_ids: itemIds })
      await fetchOverview(branchId)
      await fetchQueue(branchId)
      return res.data
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Gagal menambahkan produk ke antrean')
    } finally {
      isActionLoading.value = false
    }
  }

  const deleteFromQueue = async (id: number, branchId?: string) => {
    try {
      await shopeeBoostApi.deleteFromQueue(id)
      await fetchOverview(branchId)
      await fetchQueue(branchId)
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Gagal menghapus produk dari antrean')
    }
  }

  const toggleItemActive = async (id: number, isActive: boolean, branchId?: string) => {
    try {
      await shopeeBoostApi.toggleItemActive(id, isActive)
      await fetchOverview(branchId)
      await fetchQueue(branchId)
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Gagal mengubah status item')
    }
  }

  const reorderQueue = async (itemIds: number[]) => {
    try {
      await shopeeBoostApi.reorderQueue(itemIds)
    } catch (err: any) {
      throw new Error(err.response?.data?.error || 'Gagal menyimpan urutan antrean')
    }
  }

  const triggerBoostNow = async (branchId?: string) => {
    isActionLoading.value = true
    try {
      const res = await shopeeBoostApi.triggerBoostNow(branchId)
      await fetchOverview(branchId)
      await fetchQueue(branchId)
      return res.data
    } catch (err: any) {
      throw new Error(err.response?.data?.message || err.response?.data?.error || 'Gagal mengeksekusi boost manual')
    } finally {
      isActionLoading.value = false
    }
  }

  const fetchLogs = async (branchId?: string) => {
    try {
      const res = await shopeeBoostApi.getLogs({ branch_id: branchId, limit: 30 })
      if (res.data.success) {
        logs.value = res.data.data
      }
    } catch (err: any) {
      console.error('Failed to fetch logs:', err)
    }
  }

  return {
    overview,
    queue,
    queueTotal,
    logs,
    isLoading,
    isActionLoading,
    error,
    lastUpdated,
    fetchOverview,
    toggleAutoBoost,
    fetchQueue,
    addToQueue,
    deleteFromQueue,
    toggleItemActive,
    reorderQueue,
    triggerBoostNow,
    fetchLogs,
  }
})
