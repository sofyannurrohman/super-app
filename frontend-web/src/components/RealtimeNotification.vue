<template>
  <!-- ─── Toast Popup (muncul di kanan bawah) ─────────────────────────── -->
  <Teleport to="body">
    <div class="nota-toast-container" aria-live="polite">
      <TransitionGroup name="toast" tag="div">
        <div
          v-for="toast in activeToasts"
          :key="toast.id"
          class="nota-toast"
          @click="openTransaction(toast)"
        >
          <!-- Glow strip -->
          <div class="toast-strip" :class="toast.type || 'nota_fisik'" />

          <div class="toast-body">
            <!-- Icon + badge -->
            <div class="toast-icon-wrap">
              <div class="toast-icon" :class="getNotifConfig(toast.type).bgClass">
                <component :is="getNotifConfig(toast.type).icon" class="w-5 h-5" :class="getNotifConfig(toast.type).colorClass" />
              </div>
              <span class="toast-badge">BARU</span>
            </div>

            <!-- Content -->
            <div class="toast-content">
              <p class="toast-title">{{ toast.title || getNotifConfig(toast.type).title }}</p>
              <p class="toast-message text-xs text-slate-600 font-semibold leading-relaxed mt-1" v-if="toast.message">
                {{ toast.message }}
              </p>
              <template v-if="toast.type === 'nota_fisik' || !toast.type">
                <p class="toast-salesman">
                  <User class="w-3 h-3 inline mr-1" />{{ toast.salesman }}
                </p>
                <p class="toast-store">
                  <Store class="w-3 h-3 inline mr-1" />{{ toast.storeName }}
                </p>
              </template>
              <p class="toast-time">{{ toast.createdAt }}</p>
            </div>

            <!-- Preview thumb -->
            <div v-if="toast.imageUrl" class="toast-thumb">
              <img :src="resolveImageUrl(toast.imageUrl)" :alt="'Nota ' + toast.receiptNo" />
              <div class="toast-thumb-overlay">
                <Eye class="w-4 h-4 text-white" />
              </div>
            </div>
          </div>

          <!-- Progress bar auto-dismiss -->
          <div class="toast-progress" :style="{ animationDuration: `${TOAST_DURATION}ms` }" />

          <!-- Close button -->
          <button class="toast-close" @click.stop="dismissToast(toast.id)">
            <X class="w-3 h-3" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>

  <!-- ─── Bell Button ────────────────────────────────────────────────── -->
  <div class="notif-bell-wrap" ref="bellRef">
    <button
      class="notif-bell-btn"
      :class="{ 'has-notif': unreadCount > 0 }"
      @click.stop="togglePanel"
      title="Notifikasi Realtime"
    >
      <Bell class="w-5 h-5" />
      <span v-if="unreadCount > 0" class="notif-badge">
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
      <!-- Pulse ring saat ada notif baru -->
      <span v-if="hasNewNotif" class="notif-pulse-ring" />
    </button>

    <!-- ─── Notification Panel ────────────────────────────────────────── -->
    <Transition name="panel">
      <div v-if="isPanelOpen" class="notif-panel" @click.stop>
        <!-- Header -->
        <div class="notif-panel-header">
          <div class="flex items-center gap-2">
            <div class="w-2 h-2 rounded-full bg-amber-400 animate-pulse" />
            <span class="text-sm font-black text-slate-900">Pemberitahuan</span>
            <span v-if="unreadCount > 0" class="notif-count-badge">
              {{ unreadCount }} baru
            </span>
          </div>
          <div class="flex gap-1">
            <button
              v-if="notifications.length > 0"
              class="notif-action-btn"
              @click="markAllAsRead"
              title="Tandai semua sudah dibaca"
            >
              <CheckCheck class="w-3.5 h-3.5" />
            </button>
            <button class="notif-action-btn" @click="clearAll" title="Hapus semua">
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- SSE status -->
        <div class="notif-sse-status" :class="isConnected ? 'connected' : 'disconnected'">
          <div class="w-1.5 h-1.5 rounded-full" :class="isConnected ? 'bg-green-400' : 'bg-red-400'" />
          <span>{{ isConnected ? 'Terhubung realtime' : 'Tidak terhubung' }}</span>
        </div>

        <!-- List -->
        <div class="notif-list" v-if="notifications.length > 0">
          <div
            v-for="notif in notifications"
            :key="notif.id"
            class="notif-item"
            :class="{ unread: !notif.isRead }"
            @click="handleNotifClick(notif)"
          >
            <!-- Unread dot -->
            <div class="notif-unread-dot" v-if="!notif.isRead" />

            <!-- Thumb / Placeholder Icon -->
            <div v-if="notif.imageUrl" class="notif-item-thumb">
              <img :src="resolveImageUrl(notif.imageUrl)" :alt="'Nota'" />
            </div>
            <div v-else class="notif-item-thumb-placeholder" :class="getNotifConfig(notif.type).bgClass">
              <component :is="getNotifConfig(notif.type).icon" class="w-4 h-4" :class="getNotifConfig(notif.type).colorClass" />
            </div>

            <!-- Info -->
            <div class="notif-item-info">
              <p class="notif-item-title">{{ notif.title || getNotifConfig(notif.type).title }}</p>
              <p class="notif-item-message text-[11px] text-slate-500 line-clamp-2 mt-0.5" v-if="notif.message">
                {{ notif.message }}
              </p>
              <p class="notif-item-sub" v-if="notif.type === 'nota_fisik' || !notif.type">
                <User class="w-2.5 h-2.5 inline" /> {{ notif.salesman }} — {{ notif.storeName }}
              </p>
              <p class="notif-item-time">{{ notif.createdAt }}</p>
            </div>

            <!-- Arrow -->
            <ChevronRight class="w-3.5 h-3.5 text-slate-300 shrink-0" />
          </div>
        </div>

        <!-- Empty state -->
        <div v-else class="notif-empty">
          <Bell class="w-8 h-8 text-slate-200 mb-2" />
          <p class="text-xs text-slate-400 font-semibold">Belum ada notifikasi baru</p>
          <p class="text-[10px] text-slate-300 mt-0.5">Notifikasi baru akan muncul otomatis</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import {
  Bell,
  FileImage,
  User,
  Store,
  Eye,
  X,
  CheckCheck,
  Trash2,
  ChevronRight,
  Calendar,
  Clock,
  CreditCard,
  ShoppingBag,
  AlertTriangle,
  MapPin,
  RefreshCw,
  Truck,
  Factory,
  DollarSign,
  Megaphone,
} from 'lucide-vue-next'
import { useNotificationStore, type NotaFisikNotification } from '@/stores/notification'
import { useNotaFisikSSE } from '@/composables/useNotaFisikSSE'

const TOAST_DURATION = 6000 // ms

const router = useRouter()
const notifStore = useNotificationStore()
const { isConnected, connect, disconnect } = useNotaFisikSSE()

const notifications = computed(() => notifStore.notifications)
const unreadCount = computed(() => notifStore.unreadCount)

// Panel state
const isPanelOpen = ref(false)
const bellRef = ref<HTMLElement>()
const hasNewNotif = ref(false)

// Toast queue
const activeToasts = ref<NotaFisikNotification[]>([])
const toastTimers = new Map<string, ReturnType<typeof setTimeout>>()

// Helper config untuk visual tiap tipe notifikasi
function getNotifConfig(type: string) {
  switch (type) {
    case 'leave':
      return {
        icon: Calendar,
        colorClass: 'text-blue-600',
        bgClass: 'bg-blue-100',
        title: 'Pengajuan Cuti',
        route: '/leave-requests',
      }
    case 'overtime':
      return {
        icon: Clock,
        colorClass: 'text-purple-600',
        bgClass: 'bg-purple-100',
        title: 'Pengajuan Lembur',
        route: '/overtimes',
      }
    case 'reimbursement':
      return {
        icon: CreditCard,
        colorClass: 'text-green-600',
        bgClass: 'bg-green-100',
        title: 'Klaim Reimbursement',
        route: '/reimbursements',
      }
    case 'sales_order':
      return {
        icon: ShoppingBag,
        colorClass: 'text-indigo-600',
        bgClass: 'bg-indigo-100',
        title: 'Sales Order Baru',
        route: '/sales/pesanan',
      }
    case 'low_stock':
      return {
        icon: AlertTriangle,
        colorClass: 'text-red-600',
        bgClass: 'bg-red-100',
        title: 'Peringatan Stok',
        route: '/warehouse/stock',
      }
    case 'sales_visit':
      return {
        icon: MapPin,
        colorClass: 'text-teal-600',
        bgClass: 'bg-teal-100',
        title: 'Kunjungan Sales',
        route: '/sales/kunjungan',
      }
    case 'sales_return':
      return {
        icon: RefreshCw,
        colorClass: 'text-rose-600',
        bgClass: 'bg-rose-100',
        title: 'Retur Penjualan Baru',
        route: '/warehouse/retur',
      }
    case 'pod':
      return {
        icon: Truck,
        colorClass: 'text-emerald-600',
        bgClass: 'bg-emerald-100',
        title: 'Pengiriman Selesai (POD)',
        route: '/sales/delivery',
      }
    case 'factory_return':
      return {
        icon: Factory,
        colorClass: 'text-orange-600',
        bgClass: 'bg-orange-100',
        title: 'Retur ke Pabrik Baru',
        route: '/factory/retur',
      }
    case 'payment':
      return {
        icon: DollarSign,
        colorClass: 'text-amber-600',
        bgClass: 'bg-amber-100',
        title: 'Pembayaran Diterima',
        route: '/finance/approval',
      }
    case 'announcement':
      return {
        icon: Megaphone,
        colorClass: 'text-amber-600',
        bgClass: 'bg-amber-100',
        title: 'Pengumuman Baru',
        route: '/announcements',
      }
    default:
      return {
        icon: FileImage,
        colorClass: 'text-amber-600',
        bgClass: 'bg-amber-100',
        title: 'Nota Fisik',
        route: '/sales/penjualan',
      }
  }
}

// Resolve image URL yang mungkin relative path
function resolveImageUrl(url: string): string {
  if (!url) return ''
  if (url.startsWith('http')) return url
  const base = (import.meta.env.VITE_API_BASE_URL || '/api').replace('/api', '')
  return `${base}${url}`
}

// Watch notifikasi baru → tampilkan toast
let prevLength = 0
watch(
  () => notifStore.notifications.length,
  (newLen) => {
    if (newLen > prevLength) {
      const newest = notifStore.notifications[0]
      if (newest) {
        showToast(newest)
        hasNewNotif.value = true
        setTimeout(() => (hasNewNotif.value = false), 3000)
      }
    }
    prevLength = newLen
  },
)

function showToast(notif: NotaFisikNotification) {
  activeToasts.value.unshift(notif)
  // Batasi 3 toast sekaligus
  if (activeToasts.value.length > 3) {
    const removed = activeToasts.value.pop()
    if (removed) clearToastTimer(removed.id)
  }
  const timer = setTimeout(() => dismissToast(notif.id), TOAST_DURATION)
  toastTimers.set(notif.id, timer)
}

function dismissToast(id: string) {
  clearToastTimer(id)
  activeToasts.value = activeToasts.value.filter((t) => t.id !== id)
}

function clearToastTimer(id: string) {
  const t = toastTimers.get(id)
  if (t) {
    clearTimeout(t)
    toastTimers.delete(id)
  }
}

function openTransaction(notif: NotaFisikNotification) {
  notifStore.markAsRead(notif.id)
  dismissToast(notif.id)

  if (notif.type === 'announcement') {
    window.dispatchEvent(
      new CustomEvent('hris:open-announcement-modal', {
        detail: {
          id: notif.transactionId || notif.data?.announcement_id,
          title: notif.title,
          content: notif.data?.content || notif.message,
          category: notif.data?.category || 'general',
          created_at: notif.createdAt,
        },
      }),
    )
    return
  }

  const config = getNotifConfig(notif.type)
  const query: Record<string, string> = {}
  
  if (notif.transactionId) {
    query.highlight = notif.transactionId
  }
  
  if (notif.data && typeof notif.data === 'object') {
    Object.keys(notif.data).forEach(key => {
      query[key] = String(notif.data[key])
    })
  }

  router.push({ path: config.route, query })
}

function handleNotifClick(notif: NotaFisikNotification) {
  notifStore.markAsRead(notif.id)
  isPanelOpen.value = false

  if (notif.type === 'announcement') {
    window.dispatchEvent(
      new CustomEvent('hris:open-announcement-modal', {
        detail: {
          id: notif.transactionId || notif.data?.announcement_id,
          title: notif.title,
          content: notif.data?.content || notif.message,
          category: notif.data?.category || 'general',
          created_at: notif.createdAt,
        },
      }),
    )
    return
  }

  const config = getNotifConfig(notif.type)
  const query: Record<string, string> = {}
  
  if (notif.transactionId) {
    query.highlight = notif.transactionId
  }
  
  if (notif.data && typeof notif.data === 'object') {
    Object.keys(notif.data).forEach(key => {
      query[key] = String(notif.data[key])
    })
  }

  router.push({ path: config.route, query })
}

function togglePanel() {
  isPanelOpen.value = !isPanelOpen.value
}

function markAllAsRead() {
  notifStore.markAllAsRead()
}

function clearAll() {
  notifStore.clearAll()
  activeToasts.value = []
  toastTimers.forEach((t) => clearTimeout(t))
  toastTimers.clear()
}

// Close panel on outside click
function onClickOutside(e: MouseEvent) {
  if (bellRef.value && !bellRef.value.contains(e.target as Node)) {
    isPanelOpen.value = false
  }
}

onMounted(() => {
  connect()
  document.addEventListener('click', onClickOutside)
})

onBeforeUnmount(() => {
  disconnect()
  document.removeEventListener('click', onClickOutside)
  toastTimers.forEach((t) => clearTimeout(t))
})
</script>

<style scoped>
/* ── Toast Container ─────────────────────────── */
.nota-toast-container {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 10px;
  pointer-events: none;
}

.nota-toast {
  pointer-events: all;
  position: relative;
  width: 320px;
  background: white;
  border-radius: 16px;
  box-shadow:
    0 20px 60px rgba(0, 0, 0, 0.15),
    0 0 0 1px rgba(245, 158, 11, 0.15);
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.nota-toast:hover {
  transform: translateY(-2px);
  box-shadow:
    0 24px 70px rgba(0, 0, 0, 0.18),
    0 0 0 2px rgba(245, 158, 11, 0.3);
}

.toast-strip {
  height: 3px;
  background: linear-gradient(90deg, #f59e0b, #f97316, #ef4444);
}

.toast-strip.leave {
  background: linear-gradient(90deg, #3b82f6, #60a5fa, #93c5fd);
}

.toast-strip.overtime {
  background: linear-gradient(90deg, #8b5cf6, #a78bfa, #c4b5fd);
}

.toast-strip.reimbursement {
  background: linear-gradient(90deg, #10b981, #34d399, #6ee7b7);
}

.toast-strip.sales_order {
  background: linear-gradient(90deg, #6366f1, #818cf8, #a5b4fc);
}

.toast-strip.low_stock {
  background: linear-gradient(90deg, #ef4444, #f87171, #fca5a5);
}

.toast-strip.sales_visit {
  background: linear-gradient(90deg, #0d9488, #2dd4bf, #99f6e4);
}

.toast-strip.sales_return {
  background: linear-gradient(90deg, #e11d48, #fb7185, #fecdd3);
}

.toast-strip.pod {
  background: linear-gradient(90deg, #059669, #34d399, #a7f3d0);
}

.toast-strip.factory_return {
  background: linear-gradient(90deg, #ea580c, #f97316, #ffedd5);
}

.toast-strip.payment {
  background: linear-gradient(90deg, #fbbf24, #f59e0b, #d97706);
}

.toast-body {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px 12px;
}

.toast-icon-wrap {
  position: relative;
  shrink: 0;
}

.toast-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: linear-gradient(135deg, #fef3c7, #fde68a);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.toast-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  background: #ef4444;
  color: white;
  font-size: 8px;
  font-weight: 900;
  padding: 1px 4px;
  border-radius: 6px;
  letter-spacing: 0.5px;
}

.toast-content {
  flex: 1;
  min-width: 0;
}

.toast-title {
  font-size: 13px;
  font-weight: 900;
  color: #0f172a;
  margin-bottom: 2px;
}

.toast-salesman {
  font-size: 11px;
  font-weight: 700;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 4px;
}

.toast-store {
  font-size: 11px;
  color: #94a3b8;
  display: flex;
  align-items: center;
  gap: 4px;
}

.toast-time {
  font-size: 10px;
  color: #cbd5e1;
  margin-top: 4px;
}

.toast-thumb {
  position: relative;
  width: 56px;
  height: 56px;
  border-radius: 10px;
  overflow: hidden;
  flex-shrink: 0;
}

.toast-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.toast-thumb-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}

.nota-toast:hover .toast-thumb-overlay {
  opacity: 1;
}

/* Progress bar */
.toast-progress {
  height: 3px;
  background: linear-gradient(90deg, #f59e0b, #fde68a);
  animation: toast-progress-shrink linear forwards;
  transform-origin: left;
}

@keyframes toast-progress-shrink {
  from { transform: scaleX(1); }
  to { transform: scaleX(0); }
}

/* Close button */
.toast-close {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #f1f5f9;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  opacity: 0;
  transition: opacity 0.2s;
}

.nota-toast:hover .toast-close {
  opacity: 1;
}

/* Toast transitions */
.toast-enter-active {
  animation: toast-in 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.toast-leave-active {
  animation: toast-out 0.3s ease forwards;
}

@keyframes toast-in {
  from {
    opacity: 0;
    transform: translateX(100%) scale(0.9);
  }
  to {
    opacity: 1;
    transform: translateX(0) scale(1);
  }
}

@keyframes toast-out {
  from {
    opacity: 1;
    transform: translateX(0) scale(1);
  }
  to {
    opacity: 0;
    transform: translateX(100%) scale(0.9);
  }
}

/* ── Bell Button ──────────────────────────────── */
.notif-bell-wrap {
  position: relative;
}

.notif-bell-btn {
  position: relative;
  width: 38px;
  height: 38px;
  border-radius: 10px;
  background: #f8fafc;
  border: 1.5px solid #e2e8f0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  transition: all 0.2s;
}

.notif-bell-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #334155;
}

.notif-bell-btn.has-notif {
  background: linear-gradient(135deg, #fffbeb, #fef3c7);
  border-color: #fbbf24;
  color: #d97706;
}

.notif-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
  font-size: 10px;
  font-weight: 900;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid white;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.4);
}

.notif-pulse-ring {
  position: absolute;
  inset: -4px;
  border-radius: 14px;
  border: 2px solid #fbbf24;
  animation: bell-pulse 1.5s ease-out infinite;
  pointer-events: none;
}

@keyframes bell-pulse {
  0% { transform: scale(1); opacity: 0.8; }
  100% { transform: scale(1.3); opacity: 0; }
}

/* ── Notification Panel ───────────────────────── */
.notif-panel {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 340px;
  background: white;
  border-radius: 20px;
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.12),
    0 0 0 1px rgba(0, 0, 0, 0.06);
  z-index: 1000;
  overflow: hidden;
}

.notif-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 16px 10px;
  border-bottom: 1px solid #f1f5f9;
}

.notif-count-badge {
  background: #fef3c7;
  color: #d97706;
  font-size: 10px;
  font-weight: 900;
  padding: 2px 7px;
  border-radius: 99px;
}

.notif-action-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  transition: all 0.15s;
}

.notif-action-btn:hover {
  background: #f1f5f9;
  color: #64748b;
}

.notif-sse-status {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  font-size: 10px;
  font-weight: 700;
}

.notif-sse-status.connected {
  background: #f0fdf4;
  color: #16a34a;
}

.notif-sse-status.disconnected {
  background: #fef2f2;
  color: #dc2626;
}

.notif-list {
  max-height: 380px;
  overflow-y: auto;
}

.notif-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f8fafc;
  transition: background 0.15s;
  position: relative;
}

.notif-item:hover {
  background: #fafafa;
}

.notif-item.unread {
  background: linear-gradient(90deg, #fffbeb 0%, white 100%);
}

.notif-unread-dot {
  position: absolute;
  left: 6px;
  top: 50%;
  transform: translateY(-50%);
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #f59e0b;
}

.notif-item-thumb {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  overflow: hidden;
  flex-shrink: 0;
  background: #f1f5f9;
}

.notif-item-thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.notif-item-thumb-placeholder {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.notif-item-info {
  flex: 1;
  min-width: 0;
}

.notif-item-title {
  font-size: 12px;
  font-weight: 800;
  color: #0f172a;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.notif-item-sub {
  font-size: 11px;
  color: #64748b;
  margin-top: 1px;
}

.notif-item-time {
  font-size: 10px;
  color: #cbd5e1;
  margin-top: 1px;
}

.notif-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

/* Panel transition */
.panel-enter-active {
  animation: panel-in 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.panel-leave-active {
  animation: panel-out 0.2s ease forwards;
}

@keyframes panel-in {
  from {
    opacity: 0;
    transform: translateY(-8px) scale(0.97);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes panel-out {
  from {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
  to {
    opacity: 0;
    transform: translateY(-8px) scale(0.97);
  }
}
</style>
