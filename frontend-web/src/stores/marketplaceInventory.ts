import { defineStore } from 'pinia'
import { ref } from 'vue'
import { marketplaceInventoryApi } from '@/api/marketplaceInventory'

export const useMarketplaceInventoryStore = defineStore('marketplaceInventory', () => {
  const inventories = ref<any[]>([])
  const logs = ref<any[]>([])
  const opnames = ref<any[]>([])
  const isLoading = ref(false)
  const isActionLoading = ref(false)
  const lastUpdated = ref<Date | null>(null)

  const fetchInventories = async () => {
    isLoading.value = true
    try {
      const response = await marketplaceInventoryApi.getAll()
      inventories.value = response.data
      lastUpdated.value = new Date()
    } catch (error) {
      console.error('Failed to fetch inventories:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const createInventory = async (data: any) => {
    try {
      const response = await marketplaceInventoryApi.create(data)
      inventories.value.unshift(response.data)
      return response.data
    } catch (error) {
      throw error
    }
  }

  const updateInventory = async (id: string, data: any) => {
    try {
      const response = await marketplaceInventoryApi.update(id, data)
      await fetchInventories() // Refresh
      return response.data
    } catch (error) {
      throw error
    }
  }

  const deleteInventory = async (id: string) => {
    try {
      await marketplaceInventoryApi.delete(id)
      inventories.value = inventories.value.filter((i) => i.id !== id)
    } catch (error) {
      throw error
    }
  }

  const stockIn = async (id: string, data: any) => {
    isActionLoading.value = true
    try {
      await marketplaceInventoryApi.stockIn(id, data)
      await fetchInventories()
    } catch (error) {
      throw error
    } finally {
      isActionLoading.value = false
    }
  }

  const stockOut = async (id: string, data: any) => {
    isActionLoading.value = true
    try {
      await marketplaceInventoryApi.stockOut(id, data)
      await fetchInventories()
    } catch (error) {
      throw error
    } finally {
      isActionLoading.value = false
    }
  }

  const doOpname = async (id: string, data: any) => {
    isActionLoading.value = true
    try {
      await marketplaceInventoryApi.doOpname(id, data)
      await fetchInventories()
    } catch (error) {
      throw error
    } finally {
      isActionLoading.value = false
    }
  }

  return {
    inventories,
    logs,
    opnames,
    isLoading,
    isActionLoading,
    lastUpdated,
    fetchInventories,
    createInventory,
    updateInventory,
    deleteInventory,
    stockIn,
    stockOut,
    doOpname,
  }
})
