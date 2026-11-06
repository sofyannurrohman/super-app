import { ref } from 'vue'
import { defineStore } from 'pinia'

export interface RealtimeNotification {
  id: string
  type: 'nota_fisik' | 'leave' | 'overtime' | 'reimbursement' | 'sales_order' | 'low_stock' | 'sales_visit' | 'sales_return' | 'pod' | 'factory_return' | 'payment' | 'announcement' | 'branding_new_order' | 'branding_revision' | 'warehouse_new_sj' | 'warehouse_new_return' | 'factory_shipment_request' | 'warehouse_transfer_approved' | 'warehouse_transfer_adjusted' | 'warehouse_transfer_shipped' | 'warehouse_transfer_rejected'
  title: string
  message: string
  companyId: string
  createdAt: string
  isRead: boolean
  timestamp: number
  
  // Backward compatibility fields
  transactionId: string
  receiptNo: string
  salesman: string
  storeName: string
  imageUrl: string
  data?: any
}

// Alias for backward compatibility in components
export type NotaFisikNotification = RealtimeNotification

export const useNotificationStore = defineStore('notification', () => {
  const notifications = ref<RealtimeNotification[]>([])
  const unreadCount = ref(0)

  function addNotification(data: Omit<RealtimeNotification, 'id' | 'isRead' | 'timestamp'>) {
    const now = Date.now()
    
    // Deduplikasi: Cek apakah notifikasi yang sama persis baru saja masuk dalam window 15 detik terakhir
    const isDuplicate = notifications.value.some((n) => {
      // Jika memiliki transactionId / visit_id yang sama dan tipe sama
      if (data.transactionId && n.transactionId === data.transactionId && n.type === data.type) {
        return now - n.timestamp < 15000 || (!n.isRead && n.message === data.message)
      }
      // Fallback matching berdasarkan isi pesan & title dalam window 10 detik
      if (n.type === data.type && n.title === data.title && n.message === data.message) {
        return now - n.timestamp < 10000
      }
      return false
    })

    if (isDuplicate) {
      console.log('[NotificationStore] Ignored duplicate realtime notification:', data.type, data.transactionId)
      return
    }

    const notif: RealtimeNotification = {
      ...data,
      id: `${data.transactionId || 'notif'}-${now}-${Math.random().toString(36).substr(2, 9)}`,
      isRead: false,
      timestamp: now,
    }
    // Tambah ke depan list
    notifications.value.unshift(notif)
    unreadCount.value++

    // Batas simpan 50 notifikasi di memori
    if (notifications.value.length > 50) {
      notifications.value.pop()
    }
  }

  function markAsRead(id: string) {
    const notif = notifications.value.find((n) => n.id === id)
    if (notif && !notif.isRead) {
      notif.isRead = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
  }

  function markAllAsRead() {
    notifications.value.forEach((n) => (n.isRead = true))
    unreadCount.value = 0
  }

  function clearAll() {
    notifications.value = []
    unreadCount.value = 0
  }

  return {
    notifications,
    unreadCount,
    addNotification,
    markAsRead,
    markAllAsRead,
    clearAll,
  }
})
