<template>
  <div class="relative" ref="dropdownRef">
    <!-- Megaphone Button in Header -->
    <button
      class="relative flex items-center justify-center w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-slate-50 border border-slate-200 text-slate-600 hover:text-amber-600 hover:bg-amber-50/60 hover:border-amber-200/80 transition-all duration-200 active:scale-95 shadow-sm"
      :class="{ 'border-amber-300 bg-amber-50/80 text-amber-600': isOpen || pendingCount > 0 }"
      @click="toggleDropdown"
      title="Pengumuman Perusahaan"
      type="button"
    >
      <Megaphone class="w-4 h-4 sm:w-4.5 sm:h-4.5 transition-transform duration-200" :class="{ 'scale-110 text-amber-600': isOpen }" />

      <!-- Red Pulse Badge for Pending Confirmation Popups -->
      <span
        v-if="pendingCount > 0"
        class="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-600 px-1 text-[9px] font-black text-white shadow-sm ring-2 ring-white animate-pulse"
      >
        {{ pendingCount > 9 ? '9+' : pendingCount }}
      </span>

      <!-- Normal Unread / Active Announcement Dot (if no pending popups) -->
      <span
        v-else-if="announcements.length > 0 && hasNewAnnouncement"
        class="absolute top-1 right-1 w-2 h-2 rounded-full bg-amber-500 ring-2 ring-white"
      />
    </button>

    <!-- Dropdown Panel -->
    <Transition name="dropdown-fade">
      <div
        v-if="isOpen"
        class="absolute right-0 mt-2 w-[340px] sm:w-[420px] max-w-[calc(100vw-24px)] rounded-3xl bg-white shadow-2xl border border-slate-100 overflow-hidden z-50 animate-in fade-in-50 zoom-in-95 duration-150 flex flex-col max-h-[85vh]"
      >
        <!-- Header -->
        <div class="px-5 pt-4 pb-3 border-b border-slate-100 bg-slate-50/60 flex items-center justify-between">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-xl bg-amber-100 text-amber-700 flex items-center justify-center shadow-xs">
              <Megaphone class="w-4 h-4" />
            </div>
            <div>
              <h3 class="text-xs font-black text-slate-900 tracking-wider uppercase">
                Pengumuman Perusahaan
              </h3>
              <p class="text-[10px] text-slate-500 font-semibold">
                {{ announcements.length }} pengumuman aktif
              </p>
            </div>
          </div>
          <button
            class="p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-200/60 transition-colors"
            @click="isOpen = false"
            title="Tutup"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Filter Chips -->
        <div class="px-4 py-2 bg-white border-b border-slate-100 flex items-center gap-1.5 overflow-x-auto scrollbar-none text-[11px] font-bold">
          <button
            class="px-2.5 py-1 rounded-lg transition-colors shrink-0"
            :class="selectedFilter === 'all' ? 'bg-amber-500 text-white shadow-xs' : 'bg-slate-100 text-slate-600 hover:bg-slate-200/80'"
            @click="selectedFilter = 'all'"
          >
            Semua ({{ announcements.length }})
          </button>
          <button
            v-if="pendingCount > 0"
            class="px-2.5 py-1 rounded-lg transition-colors shrink-0 flex items-center gap-1"
            :class="selectedFilter === 'pending' ? 'bg-red-600 text-white shadow-xs' : 'bg-red-50 text-red-700 border border-red-200 hover:bg-red-100'"
            @click="selectedFilter = 'pending'"
          >
            <AlertTriangle class="w-3 h-3" />
            <span>Perlu Konfirmasi ({{ pendingCount }})</span>
          </button>
          <button
            class="px-2.5 py-1 rounded-lg transition-colors shrink-0"
            :class="selectedFilter === 'urgent' ? 'bg-amber-500 text-white shadow-xs' : 'bg-slate-100 text-slate-600 hover:bg-slate-200/80'"
            @click="selectedFilter = 'urgent'"
          >
            Penting
          </button>
          <button
            class="px-2.5 py-1 rounded-lg transition-colors shrink-0"
            :class="selectedFilter === 'event' ? 'bg-amber-500 text-white shadow-xs' : 'bg-slate-100 text-slate-600 hover:bg-slate-200/80'"
            @click="selectedFilter = 'event'"
          >
            Acara
          </button>
        </div>

        <!-- Announcement List -->
        <div class="flex-1 overflow-y-auto p-4 space-y-3 divide-y-0 custom-scrollbar">
          <div v-if="isLoading" class="py-12 flex flex-col items-center justify-center text-slate-400 gap-2">
            <Loader2 class="w-6 h-6 animate-spin text-amber-500" />
            <span class="text-xs font-semibold">Memuat pengumuman...</span>
          </div>

          <template v-else-if="filteredAnnouncements.length > 0">
            <div
              v-for="item in filteredAnnouncements"
              :key="item.id"
              class="group relative rounded-2xl border transition-all duration-200 p-4 bg-slate-50/50 hover:bg-white hover:shadow-md"
              :class="isItemPending(item) ? 'border-red-200 bg-red-50/20' : 'border-slate-100 hover:border-slate-200'"
            >
              <!-- Top Row: Category + Status Badge -->
              <div class="flex items-center justify-between gap-2 mb-2 flex-wrap">
                <span
                  class="px-2 py-0.5 rounded-full text-[9px] font-black tracking-wider uppercase"
                  :class="getCategoryBadgeClass(item.category)"
                >
                  {{ getCategoryText(item.category) }}
                </span>

                <div class="flex items-center gap-1.5">
                  <span
                    v-if="isItemPending(item)"
                    class="px-2 py-0.5 rounded-full text-[9px] font-black bg-red-100 text-red-700 border border-red-200 flex items-center gap-1"
                  >
                    <AlertTriangle class="w-2.5 h-2.5" />
                    PERLU KONFIRMASI
                  </span>
                  <span
                    v-else
                    class="px-2 py-0.5 rounded-full text-[9px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-100 flex items-center gap-1"
                  >
                    <CheckCircle2 class="w-2.5 h-2.5" />
                    SUDAH DIBACA
                  </span>
                </div>
              </div>

              <!-- Title -->
              <h4 class="text-sm font-black text-slate-900 leading-snug group-hover:text-amber-600 transition-colors">
                {{ item.title }}
              </h4>

              <!-- Content Snippet -->
              <p class="text-xs text-slate-600 font-medium leading-relaxed line-clamp-2 mt-1.5 whitespace-pre-line">
                {{ item.content }}
              </p>

              <!-- Footer: Date + Action Button -->
              <div class="mt-3 pt-2.5 border-t border-slate-200/60 flex items-center justify-between gap-2">
                <span class="text-[10px] text-slate-400 font-semibold flex items-center gap-1">
                  <Calendar class="w-3 h-3" />
                  {{ formatDate(item.created_at || item.createdAt) }}
                </span>

                <button
                  type="button"
                  class="px-3 py-1.5 rounded-xl text-[11px] font-black transition-all flex items-center gap-1.5 shadow-xs active:scale-95"
                  :class="isItemPending(item)
                    ? 'bg-gradient-to-r from-red-600 to-rose-600 hover:from-red-700 hover:to-rose-700 text-white shadow-red-500/20'
                    : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-100'"
                  @click="openAnnouncementDetail(item)"
                >
                  <template v-if="isItemPending(item)">
                    <span>BACA & KONFIRMASI</span>
                    <ArrowRight class="w-3 h-3" />
                  </template>
                  <template v-else>
                    <BookOpen class="w-3 h-3 text-slate-500" />
                    <span>BACA ULANG</span>
                  </template>
                </button>
              </div>
            </div>
          </template>

          <div v-else class="py-12 flex flex-col items-center justify-center text-center px-4">
            <div class="w-12 h-12 rounded-2xl bg-slate-100 text-slate-400 flex items-center justify-center mb-3">
              <Megaphone class="w-6 h-6" />
            </div>
            <p class="text-xs font-bold text-slate-700">Belum Ada Pengumuman</p>
            <p class="text-[11px] text-slate-400 mt-0.5">Pengumuman dan informasi resmi perusahaan akan muncul di sini.</p>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import {
  Megaphone,
  X,
  Calendar,
  AlertTriangle,
  CheckCircle2,
  BookOpen,
  ArrowRight,
  Loader2,
} from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/axios'

interface AnnouncementItem {
  id: string
  title: string
  content: string
  category: string
  created_at?: string
  createdAt?: string
  is_popup?: boolean
  target_type?: string
  author?: string
}

const authStore = useAuthStore()
const isOpen = ref(false)
const isLoading = ref(false)
const announcements = ref<AnnouncementItem[]>([])
const pendingPopups = ref<AnnouncementItem[]>([])
const selectedFilter = ref<'all' | 'pending' | 'urgent' | 'event'>('all')
const hasNewAnnouncement = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const pendingCount = computed(() => pendingPopups.value.length)

const filteredAnnouncements = computed(() => {
  if (selectedFilter.value === 'pending') {
    return announcements.value.filter((a) => isItemPending(a))
  }
  if (selectedFilter.value === 'urgent') {
    return announcements.value.filter((a) => {
      const cat = (a.category || '').toLowerCase()
      return cat === 'urgent_call' || cat === 'alert'
    })
  }
  if (selectedFilter.value === 'event') {
    return announcements.value.filter((a) => (a.category || '').toLowerCase() === 'event')
  }
  return announcements.value
})

function isItemPending(item: AnnouncementItem): boolean {
  return pendingPopups.value.some((p) => p.id === item.id)
}

function getCategoryText(category?: string) {
  const cat = (category || '').toLowerCase()
  if (cat === 'urgent_call' || cat === 'alert') return 'Panggilan Urgent'
  if (cat === 'event') return 'Acara / Meeting'
  if (cat === 'policy') return 'Kebijakan'
  if (cat === 'birthday') return 'Ulang Tahun'
  return 'Pengumuman'
}

function getCategoryBadgeClass(category?: string) {
  const cat = (category || '').toLowerCase()
  if (cat === 'urgent_call' || cat === 'alert') {
    return 'bg-red-100 text-red-700 border border-red-200'
  }
  if (cat === 'event') {
    return 'bg-blue-100 text-blue-700 border border-blue-200'
  }
  if (cat === 'birthday') {
    return 'bg-amber-100 text-amber-800 border border-amber-200'
  }
  return 'bg-emerald-100 text-emerald-800 border border-emerald-200'
}

function formatDate(dateStr?: string) {
  if (!dateStr) return 'Hari ini'
  try {
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    return d.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

async function fetchAnnouncements() {
  if (!authStore.token) return
  isLoading.value = true
  try {
    const [resAll, resPending] = await Promise.all([
      apiClient.get('/announcements'),
      apiClient.get('/announcements/pending-popup'),
    ])

    const allData = resAll.data
    announcements.value = Array.isArray(allData?.data)
      ? allData.data
      : Array.isArray(allData)
        ? allData
        : []

    const pendingData = resPending.data
    pendingPopups.value = Array.isArray(pendingData?.data)
      ? pendingData.data
      : Array.isArray(pendingData)
        ? pendingData
        : []
  } catch (err) {
    console.debug('[AnnouncementHub] Failed to fetch announcements:', err)
  } finally {
    isLoading.value = false
  }
}

function toggleDropdown() {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    fetchAnnouncements()
    hasNewAnnouncement.value = false
  }
}

function openAnnouncementDetail(item: AnnouncementItem) {
  isOpen.value = false
  // Dispatch custom event to open AnnouncementConfirmationModal
  window.dispatchEvent(
    new CustomEvent('hris:open-announcement-modal', {
      detail: {
        id: item.id,
        announcement_id: item.id,
        title: item.title,
        content: item.content,
        category: item.category,
        created_at: item.created_at || item.createdAt,
        is_popup: item.is_popup,
      },
    }),
  )
}

function onClickOutside(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

function onRealtimeAnnouncement() {
  hasNewAnnouncement.value = true
  fetchAnnouncements()
}

watch(
  () => authStore.token,
  (token) => {
    if (token) fetchAnnouncements()
  },
)

onMounted(() => {
  fetchAnnouncements()
  document.addEventListener('click', onClickOutside)
  window.addEventListener('hris:announcement-event', onRealtimeAnnouncement)
})

onUnmounted(() => {
  document.removeEventListener('click', onClickOutside)
  window.removeEventListener('hris:announcement-event', onRealtimeAnnouncement)
})
</script>

<style scoped>
.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}
.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.96);
}

.custom-scrollbar::-webkit-scrollbar {
  width: 5px;
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
