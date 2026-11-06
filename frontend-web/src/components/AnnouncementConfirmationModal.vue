<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div
        v-if="isOpen && currentAnnouncement"
        class="fixed inset-0 z-[9999] flex items-center justify-center p-4 sm:p-6 bg-slate-950/80 backdrop-blur-md select-none"
        @click.self="handleBackdropClick"
      >
        <div
          class="relative w-full max-w-xl bg-white rounded-3xl shadow-2xl border border-slate-100 overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-200"
        >
          <!-- Top Accent Stripe -->
          <div
            class="h-2 w-full"
            :class="categoryConfig.stripeClass"
          />

          <!-- Header Section -->
          <div class="px-6 pt-5 pb-4 border-b border-slate-100 flex items-start justify-between gap-4 bg-slate-50/50">
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 shadow-sm"
                :class="categoryConfig.iconWrapClass"
              >
                <component :is="categoryConfig.icon" class="w-6 h-6" :class="categoryConfig.iconClass" />
              </div>
              <div class="min-w-0">
                <div class="flex items-center gap-2 flex-wrap">
                  <span
                    class="px-2.5 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider"
                    :class="categoryConfig.badgeClass"
                  >
                    {{ categoryConfig.badgeText }}
                  </span>
                  <span
                    v-if="queue.length > 1"
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-200 text-slate-700"
                  >
                    {{ currentIndex + 1 }} dari {{ queue.length }}
                  </span>
                </div>
                <p class="text-xs text-slate-500 font-semibold mt-1">
                  {{ formatDate(currentAnnouncement.created_at || currentAnnouncement.createdAt) }}
                </p>
              </div>
            </div>

            <!-- Close button only enabled if already confirmed / re-reading -->
            <button
              v-if="isReReadMode"
              class="w-8 h-8 rounded-full flex items-center justify-center text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
              @click="closeModal"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Title & Scrollable Content -->
          <div class="px-6 pt-4 pb-2 flex-1 flex flex-col min-h-0">
            <h2 class="text-lg sm:text-xl font-black text-slate-900 leading-tight mb-3">
              {{ currentAnnouncement.title }}
            </h2>

            <!-- Scrollable Content Box -->
            <div
              ref="contentBoxRef"
              class="flex-1 overflow-y-auto pr-2 text-sm text-slate-700 leading-relaxed space-y-3 bg-slate-50/60 p-4 rounded-2xl border border-slate-100 font-medium whitespace-pre-line custom-scrollbar"
              @scroll="handleScroll"
            >
              {{ currentAnnouncement.content }}
            </div>
          </div>

          <!-- Bottom Floating Status / Progress Bar -->
          <div class="px-6 py-2">
            <div
              v-if="!isReReadMode"
              class="flex items-center justify-center gap-2 py-1.5 px-3 rounded-xl text-xs font-bold transition-all duration-300"
              :class="hasScrolledToBottom ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-amber-50 text-amber-800 border border-amber-200 animate-pulse'"
            >
              <template v-if="!hasScrolledToBottom">
                <ArrowDownCircle class="w-4 h-4 shrink-0 text-amber-600 animate-bounce" />
                <span>Gulir ke bawah untuk membaca seluruh isi & mengaktifkan tombol konfirmasi</span>
              </template>
              <template v-else>
                <CheckCircle2 class="w-4 h-4 shrink-0 text-emerald-600" />
                <span>Anda telah membaca seluruh isi pengumuman</span>
              </template>
            </div>
          </div>

          <!-- Footer Action Buttons -->
          <div class="p-6 pt-2 border-t border-slate-100 bg-white flex flex-col sm:flex-row items-center justify-end gap-3">
            <button
              v-if="isReReadMode"
              type="button"
              class="w-full sm:w-auto px-6 py-3 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-100 transition-colors"
              @click="closeModal"
            >
              Tutup
            </button>
            <button
              type="button"
              class="w-full sm:flex-1 py-3.5 px-6 rounded-2xl text-sm font-black tracking-wide flex items-center justify-center gap-2 transition-all duration-200 shadow-md"
              :class="buttonClass"
              :disabled="!hasScrolledToBottom || isSubmitting"
              @click="confirmAnnouncement"
            >
              <template v-if="isSubmitting">
                <Loader2 class="w-4 h-4 animate-spin" />
                <span>Menyimpan Konfirmasi...</span>
              </template>
              <template v-else-if="!hasScrolledToBottom">
                <Lock class="w-4 h-4" />
                <span>BACA HINGGA AKHIR UNTUK KONFIRMASI</span>
              </template>
              <template v-else>
                <Check class="w-4 h-4" />
                <span>SAYA MENGERTI & KONFIRMASI</span>
              </template>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import {
  AlertTriangle,
  Calendar,
  Megaphone,
  CheckCircle2,
  ArrowDownCircle,
  Check,
  Lock,
  Loader2,
  X,
} from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'

export interface AnnouncementData {
  id?: string
  announcement_id?: string
  title: string
  content: string
  category: string
  is_popup?: boolean
  target_type?: string
  target_roles?: string[]
  target_user_ids?: string[]
  created_at?: string
  createdAt?: string
  author?: string
}

const authStore = useAuthStore()
const notifStore = useNotificationStore()

const isOpen = ref(false)
const queue = ref<AnnouncementData[]>([])
const currentIndex = ref(0)
const isReReadMode = ref(false)
const hasScrolledToBottom = ref(false)
const isSubmitting = ref(false)
const contentBoxRef = ref<HTMLElement | null>(null)

const currentAnnouncement = computed<AnnouncementData | null>(() => {
  if (queue.value.length === 0) return null
  return queue.value[currentIndex.value] || null
})

// Visual Config per Category
const categoryConfig = computed(() => {
  const cat = (currentAnnouncement.value?.category || '').toLowerCase()
  if (cat === 'urgent_call' || cat === 'alert') {
    return {
      badgeText: '⚠️ PANGGILAN MENGHADAP SEGERA',
      badgeClass: 'bg-red-100 text-red-700 border border-red-200',
      stripeClass: 'bg-gradient-to-r from-red-500 to-rose-600',
      iconWrapClass: 'bg-red-100 text-red-600',
      iconClass: 'text-red-600',
      icon: AlertTriangle,
      btnColor: 'bg-gradient-to-r from-red-600 to-rose-600 hover:from-red-700 hover:to-rose-700 text-white shadow-red-500/25',
    }
  }
  if (cat === 'event') {
    return {
      badgeText: '📅 UNDANGAN MEETING / ACARA',
      badgeClass: 'bg-blue-100 text-blue-700 border border-blue-200',
      stripeClass: 'bg-gradient-to-r from-blue-500 to-indigo-600',
      iconWrapClass: 'bg-blue-100 text-blue-600',
      iconClass: 'text-blue-600',
      icon: Calendar,
      btnColor: 'bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 text-white shadow-blue-500/25',
    }
  }
  return {
    badgeText: '📢 PENGUMUMAN PENTING',
    badgeClass: 'bg-emerald-100 text-emerald-800 border border-emerald-200',
    stripeClass: 'bg-gradient-to-r from-emerald-500 to-teal-600',
    iconWrapClass: 'bg-emerald-100 text-emerald-700',
    iconClass: 'text-emerald-700',
    icon: Megaphone,
    btnColor: 'bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 text-white shadow-emerald-500/25',
  }
})

const buttonClass = computed(() => {
  if (!hasScrolledToBottom.value) {
    return 'bg-slate-200 text-slate-400 cursor-not-allowed shadow-none'
  }
  return categoryConfig.value.btnColor
})

function formatDate(dateStr?: string) {
  if (!dateStr) return 'Hari ini'
  try {
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    return d.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

// Scroll checking
function handleScroll() {
  if (hasScrolledToBottom.value || !contentBoxRef.value) return
  const el = contentBoxRef.value
  // Scrolled within 15px of bottom
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 15) {
    hasScrolledToBottom.value = true
  }
}

function checkScrollability() {
  nextTick(() => {
    if (!contentBoxRef.value) return
    const el = contentBoxRef.value
    // If content doesn't overflow, immediately unlock
    if (el.scrollHeight <= el.clientHeight + 10) {
      hasScrolledToBottom.value = true
    } else {
      hasScrolledToBottom.value = false
      el.scrollTop = 0
    }
  })
}

function handleBackdropClick() {
  if (isReReadMode.value) {
    closeModal()
  }
}

function closeModal() {
  isOpen.value = false
  queue.value = []
  currentIndex.value = 0
  isReReadMode.value = false
}

// Confirm Announcement (Acknowledge API)
async function confirmAnnouncement() {
  if (!currentAnnouncement.value || !hasScrolledToBottom.value || isSubmitting.value) return

  const annId = currentAnnouncement.value.id || currentAnnouncement.value.announcement_id
  if (!annId) {
    closeModal()
    return
  }

  isSubmitting.value = true
  try {
    await apiClient.post(`/announcements/${annId}/acknowledge`)
    toast.success('Pengumuman telah berhasil dikonfirmasi', {
      description: currentAnnouncement.value.title,
    })

    // Mark notification in store as read if matched
    const notifItem = notifStore.notifications.find(
      (n) => n.transactionId === annId || (n.data && n.data.announcement_id === annId)
    )
    if (notifItem) {
      notifStore.markAsRead(notifItem.id)
    }

    // Move to next announcement if queue has more
    if (currentIndex.value < queue.value.length - 1) {
      currentIndex.value++
      hasScrolledToBottom.value = false
      checkScrollability()
    } else {
      closeModal()
    }
  } catch (error: any) {
    console.error('Failed to acknowledge announcement:', error)
    toast.error('Gagal mengonfirmasi pengumuman', {
      description: error.response?.data?.error || 'Silakan coba lagi.',
    })
  } finally {
    isSubmitting.value = false
  }
}

// Method to open single announcement in re-read mode
function openForReRead(ann: AnnouncementData) {
  queue.value = [ann]
  currentIndex.value = 0
  isReReadMode.value = true
  hasScrolledToBottom.value = true
  isOpen.value = true
  checkScrollability()
}

// Fetch pending unconfirmed popup announcements from backend
async function fetchPendingPopups() {
  if (!authStore.token || !authStore.user) return
  try {
    const res = await apiClient.get('/announcements/pending-popup')
    const rawData = res.data
    const list: AnnouncementData[] = Array.isArray(rawData?.data)
      ? rawData.data
      : Array.isArray(rawData)
        ? rawData
        : []
    if (list.length > 0) {
      queue.value = list
      currentIndex.value = 0
      isReReadMode.value = false
      hasScrolledToBottom.value = false
      isOpen.value = true
      checkScrollability()
    }
  } catch (err) {
    console.debug('[Pending Announcement Check]', err)
  }
}

// Check audience targeting match
function isUserTargeted(data: any): boolean {
  if (!authStore.user) return false
  const targetType = String(data.target_type || data.data?.target_type || 'broadcast').toLowerCase()
  if (targetType === 'broadcast' || targetType === '' || targetType === 'null' || targetType === 'undefined') {
    return true
  }

  const userRole = String(authStore.user.role || '').toLowerCase()
  const userId = String(authStore.user.id || '').toLowerCase()

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

// Handle Realtime SSE Event
function onAnnouncementEvent(event: any) {
  const detail = event.detail
  if (!detail) return

  const isPopup = detail.is_popup ?? detail.data?.is_popup ?? true
  if (isPopup && isUserTargeted(detail)) {
    const annData: AnnouncementData = {
      id: detail.announcement_id || detail.data?.announcement_id || detail.id,
      title: detail.title || detail.data?.title || 'Pengumuman Baru',
      content: detail.content || detail.data?.content || detail.message || '',
      category: detail.category || detail.data?.category || 'general',
      created_at: detail.created_at || detail.data?.created_at || new Date().toISOString(),
      is_popup: true,
    }

    // Add to queue and open immediately
    if (!isOpen.value) {
      queue.value = [annData]
      currentIndex.value = 0
      isReReadMode.value = false
      hasScrolledToBottom.value = false
      isOpen.value = true
      checkScrollability()
    } else {
      // If already open, append to queue
      queue.value.push(annData)
    }
  }
}

watch(
  () => authStore.token,
  (newToken) => {
    if (newToken) {
      fetchPendingPopups()
    }
  }
)

onMounted(() => {
  fetchPendingPopups()
  window.addEventListener('hris:announcement-event', onAnnouncementEvent)
  window.addEventListener('hris:open-announcement-modal', (e: any) => {
    if (e.detail) openForReRead(e.detail)
  })
})

onUnmounted(() => {
  window.removeEventListener('hris:announcement-event', onAnnouncementEvent)
  window.removeEventListener('hris:open-announcement-modal', () => {})
})

defineExpose({
  openForReRead,
  fetchPendingPopups,
})
</script>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.25s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 9999px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>
