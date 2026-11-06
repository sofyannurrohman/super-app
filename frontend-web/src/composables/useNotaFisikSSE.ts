import { ref, onUnmounted } from 'vue'
import { useNotificationStore } from '@/stores/notification'
import { useAuthStore } from '@/stores/auth'

/**
 * useNotaFisikSSE — composable untuk terkoneksi ke SSE endpoint backend.
 * Otomatis publish notifikasi ke notificationStore ketika ada nota fisik baru.
 */
export function useNotaFisikSSE() {
  const notifStore = useNotificationStore()
  const authStore = useAuthStore()

  const isConnected = ref(false)
  let eventSource: EventSource | null = null
  let retryTimeout: ReturnType<typeof setTimeout> | null = null
  let retryCount = 0
  const MAX_RETRIES = 10

  function connect() {
    if (!authStore.token) return

    // SSE tidak support custom header, jadi token via query param
    // employee_id & branch_id dikirim agar backend bisa melakukan personalized routing (mis: notif revisi/surat jalan per cabang)
    const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api'
    const employeeId =
      authStore.user?.employee?.id || authStore.user?.employee_id || ''
    const branchId =
      authStore.user?.employee?.branch_id || authStore.user?.branch_id || ''
    const queryParams = new URLSearchParams({
      token: authStore.token,
      ...(employeeId ? { employee_id: employeeId } : {}),
      ...(branchId ? { branch_id: branchId } : {}),
    })
    const url = `${baseUrl}/v1/sse/notifications?${queryParams.toString()}`

    eventSource = new EventSource(url)

    eventSource.addEventListener('connected', () => {
      isConnected.value = true
      retryCount = 0
      console.log('[SSE] Connected to nota fisik notification stream')
    })

// Matriks hak akses role untuk penerimaan jenis notifikasi realtime
const NOTIFICATION_ROLE_PERMISSIONS: Record<string, string[]> = {
  // Sales Check-in / Check-out & Kunjungan
  sales_visit: ['superadmin', 'admin_nota', 'supervisor_sales', 'supervisor_sales_mt', 'manager'],
  // Transaksi Penjualan & Nota Fisik
  nota_fisik: ['superadmin', 'admin_nota', 'supervisor_sales', 'supervisor_sales_mt', 'manager', 'cashier', 'cashier_pusat', 'admin_kasir'],
  sales_order: ['superadmin', 'admin_nota', 'supervisor_sales', 'supervisor_sales_mt', 'manager', 'cashier', 'cashier_pusat', 'admin_kasir'],
  sales_return: ['superadmin', 'admin_nota', 'supervisor_sales', 'supervisor_sales_mt', 'manager', 'admin_gudang', 'supervisor_gudang'],
  payment: ['superadmin', 'admin_nota', 'supervisor_sales', 'supervisor_sales_mt', 'cashier', 'cashier_pusat', 'admin_kasir', 'manager'],
  // Gudang & Logistik
  low_stock: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_pabrik', 'manager'],
  pod: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_nota', 'supervisor_sales', 'manager'],
  factory_return: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_pabrik', 'manager'],
  warehouse_new_sj: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_pabrik', 'manager'],
  warehouse_new_return: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_pabrik', 'admin_nota', 'supervisor_sales', 'manager'],
  warehouse_transfer_approved: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_pabrik', 'manager'],
  warehouse_transfer_adjusted: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_pabrik', 'manager'],
  warehouse_transfer_shipped: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_pabrik', 'manager'],
  warehouse_transfer_rejected: ['superadmin', 'admin_gudang', 'supervisor_gudang', 'admin_pabrik', 'manager'],
  // Pabrik
  factory_shipment_request: ['superadmin', 'admin_pabrik', 'manager'],
  // HR & Kepegawaian & Keuangan
  leave: ['superadmin', 'hr_admin', 'manager'],
  overtime: ['superadmin', 'hr_admin', 'manager'],
  reimbursement: ['superadmin', 'hr_admin', 'manager', 'admin_nota', 'cashier', 'cashier_pusat', 'admin_kasir'],
  // Branding — routing utama sudah personalized di backend, ini sebagai safety net frontend
  branding_new_order: ['superadmin', 'validator_desain', 'validator_toko', 'admin_purchasing', 'admin_designer'],
  branding_revision: ['superadmin', 'admin_designer'], // Hanya editor yang di-target backend yang menerima ini
}

function isNotificationAllowedForRole(type: string, user: any): boolean {
  if (!user) return false
  const userRole = String(user.role || '').toLowerCase()
  if (userRole === 'superadmin' || userRole === 'admin') {
    return true
  }

  const allowedRoles = NOTIFICATION_ROLE_PERMISSIONS[type]
  if (!allowedRoles) {
    return true
  }

  return allowedRoles.includes(userRole)
}

function isAnnouncementTargeted(data: any, user: any): boolean {
  if (!user) return false
  const targetType = String(data.target_type || data.data?.target_type || 'broadcast').toLowerCase()
  if (targetType === 'broadcast' || targetType === '' || targetType === 'null' || targetType === 'undefined') {
    return true
  }

  const userRole = String(user.role || '').toLowerCase()
  const userId = String(user.id || '').toLowerCase()

  const rawRoles = data.target_roles || data.data?.target_roles || []
  const targetRoles: string[] = Array.isArray(rawRoles)
    ? rawRoles.map((r: any) => String(r || '').toLowerCase())
    : []
  if (targetRoles.length > 0 && userRole && targetRoles.includes(userRole)) {
    return true
  }

  const rawUserIds = data.target_user_ids || data.data?.target_user_ids || []
  const targetUserIds: string[] = Array.isArray(rawUserIds)
    ? rawUserIds.map((id: any) => String(id || '').toLowerCase())
    : []
  if (targetUserIds.length > 0 && userId && targetUserIds.includes(userId)) {
    return true
  }

  return false
}

    const handleEvent = (type: any, titleDefault: string) => (event: MessageEvent) => {
      try {
        // Cek hak akses role untuk tipe notifikasi ini
        if (!isNotificationAllowedForRole(type, authStore.user)) {
          return
        }

        const data = JSON.parse(event.data)

        // For announcements, check audience targeting before adding to store & showing notification toast
        if (type === 'announcement') {
          if (!isAnnouncementTargeted(data, authStore.user)) {
            return
          }
        }

        notifStore.addNotification({
          type: type,
          title: data.title || titleDefault,
          message: data.message || '',
          companyId: data.company_id || '',
          createdAt: data.created_at || new Date().toLocaleString('id-ID'),
          transactionId: data.transaction_id || data.data?.batch_id || data.data?.announcement_id || data.data?.leave_id || data.data?.overtime_id || data.data?.reimbursement_id || data.data?.so_id || data.data?.product_id || data.data?.visit_id || data.data?.return_id || data.data?.item_id || data.data?.payment_id || '',
          receiptNo: data.receipt_no || data.data?.do_number || data.data?.so_number || data.data?.return_no || 'N/A',
          salesman: data.salesman || data.data?.employee || data.data?.salesman || data.data?.driver_name || 'System',
          storeName: data.store_name || data.data?.store_name || data.data?.branch_name || data.data?.factory_name || 'N/A',
          imageUrl: data.image_url || '',
          data: data.data || null,
        })

        // Dispatch custom browser event for realtime components (e.g. auto popup announcement)
        if (type === 'announcement') {
          window.dispatchEvent(new CustomEvent('hris:announcement-event', { detail: data }))
        }
      } catch (e) {
        console.error(`[SSE] Failed to parse ${type} event:`, e)
      }
    }

    eventSource.addEventListener('announcement', handleEvent('announcement', 'Pengumuman Baru'))
    eventSource.addEventListener('nota_fisik', handleEvent('nota_fisik', 'Nota Fisik Masuk!'))
    eventSource.addEventListener('leave', handleEvent('leave', 'Pengajuan Cuti'))
    eventSource.addEventListener('overtime', handleEvent('overtime', 'Pengajuan Lembur'))
    eventSource.addEventListener('reimbursement', handleEvent('reimbursement', 'Klaim Reimbursement'))
    eventSource.addEventListener('sales_order', handleEvent('sales_order', 'Sales Order Baru'))
    eventSource.addEventListener('low_stock', handleEvent('low_stock', 'Peringatan Stok'))
    eventSource.addEventListener('sales_visit', handleEvent('sales_visit', 'Kunjungan Sales'))
    eventSource.addEventListener('sales_return', handleEvent('sales_return', 'Retur Penjualan Baru'))
    eventSource.addEventListener('pod', handleEvent('pod', 'Pengiriman Selesai (POD)'))
    eventSource.addEventListener('factory_return', handleEvent('factory_return', 'Retur ke Pabrik Baru'))
    eventSource.addEventListener('payment', handleEvent('payment', 'Pembayaran Diterima'))
    // Branding workflow notifications
    eventSource.addEventListener('branding_new_order', handleEvent('branding_new_order', '🎨 Order Branding Baru'))
    eventSource.addEventListener('branding_revision', handleEvent('branding_revision', '✏️ Revisi Desain Diperlukan'))
    // Warehouse & Logistics notifications
    eventSource.addEventListener('warehouse_new_sj', handleEvent('warehouse_new_sj', '🚚 Surat Jalan Baru'))
    eventSource.addEventListener('warehouse_new_return', handleEvent('warehouse_new_return', '📦 Retur Barang Dilaporkan'))
    eventSource.addEventListener('warehouse_transfer_approved', handleEvent('warehouse_transfer_approved', '✅ Permintaan Pengiriman Disetujui'))
    eventSource.addEventListener('warehouse_transfer_adjusted', handleEvent('warehouse_transfer_adjusted', '✏️ Permintaan Pengiriman Disesuaikan'))
    eventSource.addEventListener('warehouse_transfer_shipped', handleEvent('warehouse_transfer_shipped', '🚚 Barang Sedang Dikirim dari Pabrik'))
    eventSource.addEventListener('warehouse_transfer_rejected', handleEvent('warehouse_transfer_rejected', '❌ Permintaan Pengiriman Ditolak'))
    // Factory notifications
    eventSource.addEventListener('factory_shipment_request', handleEvent('factory_shipment_request', '🏭 Permintaan Kirim Barang Baru'))

    eventSource.onerror = () => {
      isConnected.value = false
      eventSource?.close()
      eventSource = null

      if (retryCount < MAX_RETRIES) {
        const delay = Math.min(1000 * 2 ** retryCount, 30000) // exponential backoff max 30s
        console.log(`[SSE] Reconnecting in ${delay}ms (attempt ${retryCount + 1})`)
        retryTimeout = setTimeout(() => {
          retryCount++
          connect()
        }, delay)
      } else {
        console.warn('[SSE] Max retries reached. SSE disconnected.')
      }
    }
  }

  function disconnect() {
    if (retryTimeout) {
      clearTimeout(retryTimeout)
      retryTimeout = null
    }
    eventSource?.close()
    eventSource = null
    isConnected.value = false
    retryCount = 0
  }

  onUnmounted(() => {
    disconnect()
  })

  return { isConnected, connect, disconnect }
}
