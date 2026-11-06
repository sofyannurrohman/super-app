import apiClient from './axios'

export interface ShopeeBoostSlot {
  slot_index: number
  item_id: number
  item_name: string
  image_url: string
  cooldown_seconds: number
  cooldown_until: string
  is_occupied: boolean
}

export interface ShopeeAutoBoostSetting {
  id: number
  branch_id?: string
  shop_id: string
  is_enabled: boolean
  rotation_mode: string
  check_interval_minutes: number
  last_run_at?: string
  last_run_status?: string
  last_run_message?: string
}

export interface ShopeeAutoBoostItem {
  id: number
  branch_id?: string
  shop_id: string
  item_id: number
  item_name: string
  item_sku?: string
  image_url?: string
  price_amount: number
  sequence: number
  is_active: boolean
  last_boosted_at?: string
  cooldown_until?: string
  status: 'IDLE' | 'BOOSTED' | 'FAILED'
  last_error?: string
  current_stock?: number
  product_status?: string
  created_at: string
  updated_at: string
}

export interface ShopeeAutoBoostLog {
  id: number
  shop_id: string
  item_id: number
  item_name: string
  triggered_by: 'WORKER' | 'MANUAL'
  status: 'SUCCESS' | 'FAILED'
  message: string
  boosted_at: string
  created_at: string
}

export interface ShopeeBoostOverview {
  shop_id: string
  is_enabled: boolean
  max_slots: number
  occupied_slots: number
  available_slots: number
  total_in_queue: number
  active_in_queue: number
  slots: ShopeeBoostSlot[]
  settings: ShopeeAutoBoostSetting
  recent_logs: ShopeeAutoBoostLog[]
}

export const shopeeBoostApi = {
  getOverview(branchId?: string) {
    const params: Record<string, string> = {}
    if (branchId && branchId !== 'all') params.branch_id = branchId
    return apiClient.get<{ success: boolean; data: ShopeeBoostOverview; error?: string }>('/marketplaces/shopee/boost/overview', { params })
  },

  toggleAutoBoost(data: { branch_id?: string; is_enabled: boolean }) {
    return apiClient.post('/marketplaces/shopee/boost/toggle', data)
  },

  getQueue(params?: { branch_id?: string; search?: string }) {
    return apiClient.get<{ success: boolean; data: ShopeeAutoBoostItem[]; total: number }>('/marketplaces/shopee/boost/queue', { params })
  },

  addToQueue(data: { branch_id?: string; item_ids: number[] }) {
    return apiClient.post('/marketplaces/shopee/boost/queue', data)
  },

  deleteFromQueue(id: number) {
    return apiClient.delete(`/marketplaces/shopee/boost/queue/${id}`)
  },

  toggleItemActive(id: number, isActive: boolean) {
    return apiClient.patch(`/marketplaces/shopee/boost/queue/${id}/toggle`, { is_active: isActive })
  },

  reorderQueue(itemIds: number[]) {
    return apiClient.patch('/marketplaces/shopee/boost/queue/reorder', { item_ids: itemIds })
  },

  triggerBoostNow(branchId?: string) {
    return apiClient.post<{ success: boolean; boosted_count: number; message: string }>('/marketplaces/shopee/boost/trigger-now', {
      branch_id: branchId && branchId !== 'all' ? branchId : undefined,
    })
  },

  getLogs(params?: { branch_id?: string; limit?: number }) {
    return apiClient.get<{ success: boolean; data: ShopeeAutoBoostLog[] }>('/marketplaces/shopee/boost/logs', { params })
  },
}
