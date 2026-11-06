<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  CheckCircle2,
  XCircle,
  Clock,
  AlertTriangle,
  X,
  RefreshCw,
  ExternalLink,
  Search,
  ChevronLeft,
  ChevronRight,
  Image as ImageIcon,
  Palette,
  Store,
  Sparkles,
  ShieldCheck,
  Send,
  Eye,
  Check,
  MessageCircle,
  Phone,
  Paperclip,
  FileDown,
  User,
  Building2,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import StoreAddressHoverCard from '@/components/StoreAddressHoverCard.vue'
import { Button } from '@/components/ui/button'
import { calculateBannerDesignSize, formatBannerDimension } from '@/utils/bannerSize'

const getSizeSuggestion = (task: any) => {
  if (!task) return null
  return calculateBannerDesignSize(task.size_length, task.size_width)
}

const masterStore = useMasterDataStore()
const authStore = useAuthStore()

const isDesignValidator = computed(() => {
  return authStore.user?.role === 'validator_desain'
})

const isStoreValidator = computed(() => {
  return authStore.user?.role === 'validator_toko'
})

const isSuperAdmin = computed(() => {
  return authStore.user && ['superadmin', 'admin', 'hr_admin', 'HR', 'manager'].includes(authStore.user.role)
})

const canValidateDesign = computed(() => {
  return isDesignValidator.value || isSuperAdmin.value
})

const canValidateStore = computed(() => {
  return isStoreValidator.value || isSuperAdmin.value
})

const canValidateBoth = computed(() => {
  return isSuperAdmin.value
})

const canValidate = computed(() => {
  return canValidateDesign.value || canValidateStore.value
})

const isLoading = ref(false)
const isOnline = ref(true)

// Tab state
const activeTab = ref<'WAITING' | 'HISTORY'>('WAITING')

// Global search, branch & editor filter
const search = ref('')
const selectedBranchFilter = computed<string>({
  get: () => masterStore.selectedBrandingBranchId || 'ALL_BRANCHES',
  set: (val: string) => {
    masterStore.setSelectedBrandingBranchId(val)
  },
})
const selectedEditorFilter = ref<string>('all')
const editors = ref<{ id: number; name: string }[]>([])
let searchTimeout: any = null

const selectedBranchName = computed(() => {
  return masterStore.selectedBrandingBranchName
})

const fetchEditors = async () => {
  try {
    const res = await apiClient.get('/branding/editors')
    editors.value = res.data?.data || []
  } catch (e) {
    console.error('Failed to fetch editors in validator view', e)
  }
}

// Waiting state
const validationTasks = ref<any[]>([])
const page = ref(1)
const limit = ref(10)
const total = ref(0)

// History state
const historyTasks = ref<any[]>([])
const historyPage = ref(1)
const historyTotal = ref(0)

// Review & Validation Modal state
const showReviewModal = ref(false)
const selectedTask = ref<any>(null)
const previewUrl = ref('')
const isSubmitting = ref(false)

// Revision mode inside review modal
const isRevisionMode = ref(false)
const revisionType = ref<'design' | 'store'>('design')
const revisionNote = ref('')

const openReviewModal = (task: any) => {
  selectedTask.value = task
  isRevisionMode.value = false
  revisionNote.value = ''
  
  // Default revision type based on role
  if (isStoreValidator.value) {
    revisionType.value = 'store'
  } else {
    revisionType.value = 'design'
  }

  const rawUrl = task.design_url || task.design_link || ''
  if (rawUrl) {
    let finalUrl = rawUrl
    if (finalUrl.includes('drive.google.com')) {
      const dMatch = finalUrl.match(/\/d\/([a-zA-Z0-9_-]+)/)
      const idMatch = finalUrl.match(/[?&]id=([a-zA-Z0-9_-]+)/)
      const fileId = (dMatch && dMatch[1]) || (idMatch && idMatch[1])
      if (fileId) {
        finalUrl = `https://drive.google.com/file/d/${fileId}/preview`
      }
    }
    previewUrl.value = finalUrl
  } else {
    previewUrl.value = ''
  }

  showReviewModal.value = true
}

const closeReviewModal = () => {
  if (isSubmitting.value) return
  showReviewModal.value = false
  selectedTask.value = null
  previewUrl.value = ''
  isRevisionMode.value = false
}

const fetchValidationTasks = async (silent = false) => {
  if (!silent) isLoading.value = true
  try {
    const branchId = selectedBranchFilter.value
    const filter = branchId && branchId !== 'ALL_BRANCHES' ? `&branch_id=${branchId}` : ''
    const editorFilter = selectedEditorFilter.value && selectedEditorFilter.value !== 'all' ? `&editor_id=${selectedEditorFilter.value}` : ''
    const res = await apiClient.get(
      `/branding/banners?status=validation&page=${page.value}&limit=${limit.value}&search=${search.value}${filter}${editorFilter}`,
    )
    validationTasks.value = (res.data?.data || []).map((b: any) => ({
      id: b.id,
      shop_name: b.shop_name || 'Toko Unknown',
      shop_address: b.shop_address || '',
      shop_phone: b.shop_phone || '',
      province: b.province || '',
      regency: b.regency || '',
      district: b.district || '',
      village: b.village || '',
      latitude: b.latitude || 0,
      longitude: b.longitude || 0,
      sales_name: b.sales_name || '-',
      sales_phone: b.sales_phone || '',
      sales_branch: b.sales_branch || '',
      editor_name: b.editor_name || '-',
      type: b.type || 'Custom',
      is_bihun: Boolean(b.is_bihun),
      qty: Number(b.qty) || 1,
      size_length: b.size_length || 0,
      size_width: b.size_width || 0,
      design_url: b.design_link || b.design_url || '',
      documentation_image_url: b.documentation_image_url || '',
      notes: b.notes || '-',
      validator_name: b.validator_name || '',
      validator_position: b.validator_position || '',
      design_validator_name: b.design_validator_name || (b.is_design_approved ? b.validator_name : '') || '',
      design_validator_position: b.design_validator_position || (b.is_design_approved ? b.validator_position : '') || '',
      design_approved_at: b.design_approved_at ? formatDate(b.design_approved_at) : '',
      store_validator_name: b.store_validator_name || (b.is_store_approved ? b.validator_name : '') || '',
      store_validator_position: b.store_validator_position || (b.is_store_approved ? b.validator_position : '') || '',
      store_approved_at: b.store_approved_at ? formatDate(b.store_approved_at) : '',
      is_design_approved: Boolean(b.is_design_approved),
      is_store_approved: Boolean(b.is_store_approved),
      status: b.status,
      created_at: formatDate(b.editing_at || b.updated_at),
    }))
    total.value = res.data?.total || 0
    isOnline.value = true
  } catch (e: any) {
    isOnline.value = false
    toast.error('Gagal memuat data dari server.')
  } finally {
    if (!silent) isLoading.value = false
  }
}

const fetchHistoryTasks = async (silent = false) => {
  if (!silent) isLoading.value = true
  try {
    const branchId = selectedBranchFilter.value
    const filter = branchId && branchId !== 'ALL_BRANCHES' ? `&branch_id=${branchId}` : ''
    const editorFilter = selectedEditorFilter.value && selectedEditorFilter.value !== 'all' ? `&editor_id=${selectedEditorFilter.value}` : ''
    const res = await apiClient.get(
      `/branding/banners?status=purchasing_queue&page=${historyPage.value}&limit=${limit.value}&search=${search.value}${filter}${editorFilter}`,
    )
    historyTasks.value = (res.data?.data || []).map((b: any) => ({
      id: b.id,
      shop_name: b.shop_name || 'Toko Unknown',
      shop_address: b.shop_address || '',
      shop_phone: b.shop_phone || '',
      province: b.province || '',
      regency: b.regency || '',
      district: b.district || '',
      village: b.village || '',
      latitude: b.latitude || 0,
      longitude: b.longitude || 0,
      sales_name: b.sales_name || '-',
      sales_phone: b.sales_phone || '',
      sales_branch: b.sales_branch || '',
      editor_name: b.editor_name || '-',
      validator_name: b.validator_name || '',
      validator_position: b.validator_position || '',
      design_validator_name: b.design_validator_name || b.validator_name || '',
      design_validator_position: b.design_validator_position || b.validator_position || 'VALIDATOR DESAIN',
      design_approved_at: b.design_approved_at ? formatDate(b.design_approved_at) : '',
      store_validator_name: b.store_validator_name || b.validator_name || '',
      store_validator_position: b.store_validator_position || b.validator_position || 'VALIDATOR TOKO',
      store_approved_at: b.store_approved_at ? formatDate(b.store_approved_at) : '',
      type: b.type || 'Custom',
      is_bihun: Boolean(b.is_bihun),
      qty: Number(b.qty) || 1,
      size_length: b.size_length || 0,
      size_width: b.size_width || 0,
      status: b.status,
      design_url: b.design_link || b.design_url || '',
      documentation_image_url: b.documentation_image_url || '',
      notes: b.notes || '-',
      is_design_approved: true,
      is_store_approved: true,
      validated_at: b.validated_at ? formatDate(b.validated_at) : '',
      created_at: formatDate(b.validated_at || b.updated_at),
    }))
    historyTotal.value = res.data?.total || 0
    isOnline.value = true
  } catch (e: any) {
    isOnline.value = false
    toast.error('Gagal memuat riwayat dari server.')
  } finally {
    if (!silent) isLoading.value = false
  }
}

const fetchTasks = (silent = false) => {
  if (activeTab.value === 'WAITING') {
    fetchValidationTasks(silent)
  } else {
    fetchHistoryTasks(silent)
  }
}

watch(() => masterStore.selectedBrandingBranchId, () => {
  page.value = 1
  historyPage.value = 1
  fetchTasks()
})

const onSearchInput = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    if (activeTab.value === 'WAITING') {
      page.value = 1
    } else {
      historyPage.value = 1
    }
    fetchTasks()
  }, 500)
}

watch(activeTab, () => {
  search.value = ''
  if (activeTab.value === 'WAITING') {
    page.value = 1
  } else {
    historyPage.value = 1
  }
  fetchTasks()
})

onMounted(async () => {
  if (!masterStore.branches || masterStore.branches.length === 0) {
    await masterStore.fetchBranches()
  }
  fetchTasks()
  fetchEditors()
})

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

const formatWhatsAppUrl = (
  phone: string | null | undefined,
  type: 'store' | 'sales',
  task: any,
) => {
  if (!phone) return '#'
  let clean = phone.replace(/[^0-9]/g, '')
  if (clean.startsWith('0')) {
    clean = '62' + clean.slice(1)
  }
  const sizeInfo = formatBannerDimension(task?.size_length, task?.size_width, 'dual')
  const qtyInfo = `Jumlah: ${task?.qty || 1} Pcs`
  const msg =
    type === 'store'
      ? encodeURIComponent(
          `Halo Bapak/Ibu pemilik ${task?.shop_name || 'Toko'}, kami dari tim validasi spanduk Karsa HR ingin konfirmasi terkait spanduk toko Anda (Ukuran: ${sizeInfo}, ${qtyInfo}).`,
        )
      : encodeURIComponent(
          `Halo Mas/Mbak ${task?.sales_name || 'Sales'}, kami dari tim validasi spanduk ingin konfirmasi terkait spanduk untuk toko ${task?.shop_name || 'Toko'} (Ukuran: ${sizeInfo}, ${qtyInfo}).`,
        )
  return `https://wa.me/${clean}?text=${msg}`
}

// ── Validation Actions ───────────────────────────────────────────────────────

const handleToggleBihun = async (task: any) => {
  if (!task) return
  isSubmitting.value = true
  try {
    await apiClient.put(`/branding/banners/${task.id}`, {
      is_bihun: Boolean(task.is_bihun),
      shop_name: task.shop_name,
      size_length: task.size_length,
      size_width: task.size_width,
      type: task.type,
      notes: task.notes,
    })
    toast.success(`Kategori produk diperbarui: ${task.is_bihun ? '🍜 Desain Bihun' : 'Bukan Bihun'}`)
    // Sinkronkan data di list
    const found = validationTasks.value.find((t) => t.id === task.id)
    if (found) found.is_bihun = Boolean(task.is_bihun)
    const histFound = historyTasks.value.find((t) => t.id === task.id)
    if (histFound) histFound.is_bihun = Boolean(task.is_bihun)
  } catch (e: any) {
    task.is_bihun = !task.is_bihun // Rollback on error
    toast.error('Gagal memperbarui kategori bihun: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const handleApproveDesign = async (task: any) => {
  isSubmitting.value = true
  try {
    await apiClient.put(`/branding/banners/${task.id}/validate`, {
      status: 'approved',
      type: 'design',
      is_bihun: Boolean(task.is_bihun),
    })
    toast.success(`ACC Desain untuk ${task.shop_name} berhasil!`)
    const approverName = authStore.user?.full_name || authStore.user?.employee?.first_name || 'Validator'
    task.is_design_approved = true
    task.design_validator_name = approverName
    if (task.is_store_approved) {
      closeReviewModal()
    }
    fetchValidationTasks(true)
  } catch (e: any) {
    toast.error('Gagal ACC desain: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const handleApproveStore = async (task: any) => {
  isSubmitting.value = true
  try {
    await apiClient.put(`/branding/banners/${task.id}/validate`, {
      status: 'approved',
      type: 'store',
      is_bihun: Boolean(task.is_bihun),
    })
    toast.success(`ACC Persetujuan Toko untuk ${task.shop_name} berhasil!`)
    const approverName = authStore.user?.full_name || authStore.user?.employee?.first_name || 'Validator'
    task.is_store_approved = true
    task.store_validator_name = approverName
    if (task.is_design_approved) {
      closeReviewModal()
    }
    fetchValidationTasks(true)
  } catch (e: any) {
    toast.error('Gagal ACC toko: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const handleApproveBoth = async (task: any) => {
  if (!confirm(`ACC Penuh (Desain & Toko) untuk "${task.shop_name}" dan lanjutkan ke antrean cetak?`)) return

  isSubmitting.value = true
  try {
    await apiClient.put(`/branding/banners/${task.id}/validate`, {
      status: 'approved',
      type: 'both',
      is_bihun: Boolean(task.is_bihun),
    })
    toast.success(`Validasi lengkap! Banner ${task.shop_name} siap dicetak.`)
    const approverName = authStore.user?.full_name || authStore.user?.employee?.first_name || 'Validator'
    task.is_design_approved = true
    task.is_store_approved = true
    task.design_validator_name = approverName
    task.store_validator_name = approverName
    closeReviewModal()
    fetchValidationTasks(true)
  } catch (e: any) {
    toast.error('Gagal ACC: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const submitRevision = async () => {
  if (!selectedTask.value || !revisionNote.value.trim()) {
    toast.error('Tuliskan catatan revisi terlebih dahulu')
    return
  }

  isSubmitting.value = true
  try {
    const prefix = revisionType.value === 'design' ? '[Revisi Desain] ' : '[Revisi Toko] '
    await apiClient.put(`/branding/banners/${selectedTask.value.id}/validate`, {
      status: 'rejected',
      type: revisionType.value,
      revision_note: prefix + revisionNote.value.trim(),
    })

    toast.success('Catatan revisi berhasil dikirim kembali ke tim Editor')
    closeReviewModal()
    fetchValidationTasks(true)
  } catch (e: any) {
    toast.error('Gagal mengirim revisi: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
  }
}

// ── Pagination Helpers ───────────────────────────────────────────────────────

const nextPage = () => {
  if (activeTab.value === 'WAITING') {
    if (page.value * limit.value < total.value) {
      page.value++
      fetchValidationTasks()
    }
  } else {
    if (historyPage.value * limit.value < historyTotal.value) {
      historyPage.value++
      fetchHistoryTasks()
    }
  }
}

const prevPage = () => {
  if (activeTab.value === 'WAITING') {
    if (page.value > 1) {
      page.value--
      fetchValidationTasks()
    }
  } else {
    if (historyPage.value > 1) {
      historyPage.value--
      fetchHistoryTasks()
    }
  }
}

const totalPages = () => {
  if (activeTab.value === 'WAITING') {
    return Math.ceil(total.value / limit.value) || 1
  }
  return Math.ceil(historyTotal.value / limit.value) || 1
}

const currentPage = () => {
  return activeTab.value === 'WAITING' ? page.value : historyPage.value
}

const canNextPage = () => {
  return currentPage() < totalPages()
}

const canPrevPage = () => {
  return currentPage() > 1
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <div class="flex items-center flex-wrap gap-2 sm:gap-3">
          <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <ShieldCheck class="w-8 h-8 text-primary" />
            <span>Validator Spanduk</span>
            <span class="text-primary text-2xl font-bold">
              — {{ selectedBranchName }}
            </span>
          </h1>
          <span
            v-if="!canValidate"
            class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-black bg-slate-100 text-slate-600 border border-slate-200 shadow-sm"
          >
            <Eye class="w-3.5 h-3.5 text-slate-400" /> Mode Pantau
          </span>
        </div>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          Review desain dan konfirmasi persetujuan toko sebelum spanduk dialokasikan ke percetakan.
        </p>
      </div>
      <button
        @click="fetchTasks()"
        class="flex items-center gap-2 bg-white px-4 py-2.5 rounded-xl border border-slate-200 shadow-sm font-bold text-slate-600 hover:bg-slate-50 transition"
      >
        <RefreshCw :class="['w-4 h-4', { 'animate-spin': isLoading }]" /> Refresh
      </button>
    </div>

    <!-- Offline Alert -->
    <div
      v-if="!isOnline"
      class="bg-red-50 text-red-600 p-4 rounded-2xl border border-red-200 flex items-center gap-3"
    >
      <AlertTriangle class="w-5 h-5 shrink-0" />
      <span class="font-medium text-sm">Sistem sedang offline atau tidak dapat menghubungi server Branding.</span>
    </div>

    <!-- Tab Switcher -->
    <div class="flex gap-4 border-b border-slate-200 overflow-x-auto pb-1">
      <button
        @click="activeTab = 'WAITING'"
        :class="[
          'whitespace-nowrap pb-3 text-sm font-bold transition-all border-b-2 flex items-center gap-2',
          activeTab === 'WAITING'
            ? 'border-primary text-primary'
            : 'border-transparent text-slate-400 hover:text-slate-600',
        ]"
      >
        <span>Perlu Divalidasi</span>
        <span
          v-if="total > 0 && activeTab === 'WAITING'"
          class="px-2 py-0.5 rounded-full text-xs font-black bg-primary/10 text-primary"
        >
          {{ total }}
        </span>
      </button>
      <button
        @click="activeTab = 'HISTORY'"
        :class="[
          'whitespace-nowrap pb-3 text-sm font-bold transition-all border-b-2 flex items-center gap-2',
          activeTab === 'HISTORY'
            ? 'border-emerald-500 text-emerald-600'
            : 'border-transparent text-slate-400 hover:text-slate-600',
        ]"
      >
        <span>Riwayat Validasi (Siap Cetak)</span>
        <span
          v-if="historyTotal > 0 && activeTab === 'HISTORY'"
          class="px-2 py-0.5 rounded-full text-xs font-black bg-emerald-50 text-emerald-600"
        >
          {{ historyTotal }}
        </span>
      </button>
    </div>

    <!-- Main Card & Table -->
    <div class="bg-white border border-slate-200 rounded-3xl shadow-sm overflow-hidden flex flex-col">
      <!-- Toolbar -->
      <div class="p-5 border-b border-slate-200 flex flex-col sm:flex-row gap-4 justify-between items-center bg-slate-50/50">
        <div class="flex flex-wrap items-center gap-3 w-full sm:w-auto">
          <div class="relative w-full sm:w-72">
            <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              type="text"
              :placeholder="activeTab === 'WAITING' ? 'Cari toko atau sales...' : 'Cari riwayat toko...'"
              v-model="search"
              @input="onSearchInput"
              class="w-full pl-10 pr-4 py-2.5 bg-white border border-slate-200 rounded-xl text-sm font-medium focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all placeholder:text-slate-400 shadow-sm"
            />
          </div>

          <!-- Filter Cabang -->
          <div class="flex items-center gap-1.5 bg-white border border-slate-200 px-3 py-2 rounded-xl shadow-xs">
            <Building2 class="w-4 h-4 text-slate-400 shrink-0" />
            <select
              v-model="selectedBranchFilter"
              @change="() => { page = 1; historyPage = 1; fetchTasks() }"
              class="bg-transparent text-xs sm:text-sm font-bold text-slate-700 outline-none cursor-pointer pr-2"
            >
              <option value="ALL_BRANCHES">Semua Cabang</option>
              <option disabled>──────────</option>
              <option v-for="branch in masterStore.branches" :key="branch.id" :value="String(branch.id)">
                {{ branch.name }}
              </option>
            </select>
          </div>

          <!-- Filter Desainer -->
          <div class="flex items-center gap-1.5 bg-white border border-slate-200 px-3 py-2 rounded-xl shadow-xs">
            <User class="w-4 h-4 text-slate-400 shrink-0" />
            <select
              v-model="selectedEditorFilter"
              @change="() => { page = 1; historyPage = 1; fetchTasks() }"
              class="bg-transparent text-xs sm:text-sm font-bold text-slate-700 outline-none cursor-pointer pr-2"
            >
              <option value="all">Semua Desainer</option>
              <option v-for="e in editors" :key="e.id" :value="String(e.id)">{{ e.name }}</option>
            </select>
          </div>
        </div>

        <div class="text-xs font-bold text-slate-500 self-end sm:self-center">
          Total Spanduk:
          <span class="text-slate-800 text-sm font-extrabold">
            {{ activeTab === 'WAITING' ? total : historyTotal }}
          </span>
        </div>
      </div>

      <!-- Table -->
      <div class="overflow-x-auto w-full">
        <table class="w-full text-left border-collapse min-w-[850px]">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-[10px] font-black text-slate-500 uppercase tracking-wider">
              <th class="py-4 px-6 whitespace-nowrap">Tanggal & Toko</th>
              <th class="py-4 px-6 whitespace-nowrap">Detail Spanduk</th>
              <th class="py-4 px-6 whitespace-nowrap text-center">Jumlah (Qty)</th>
              <th class="py-4 px-6 whitespace-nowrap">Editor & Sales</th>
              <th class="py-4 px-6 whitespace-nowrap">Status Persetujuan</th>
              <th class="py-4 px-6 whitespace-nowrap text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <!-- Loading -->
            <tr v-if="isLoading && (activeTab === 'WAITING' ? validationTasks.length === 0 : historyTasks.length === 0)">
              <td colspan="6" class="py-16 text-center">
                <RefreshCw class="w-8 h-8 text-primary animate-spin mx-auto mb-3" />
                <p class="text-sm font-medium text-slate-500">Memuat data validator...</p>
              </td>
            </tr>

            <!-- Empty Waiting -->
            <tr v-else-if="activeTab === 'WAITING' && validationTasks.length === 0">
              <td colspan="6" class="py-16 text-center">
                <CheckCircle2 class="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                <p class="text-base font-bold text-slate-800">Semua Desain Telah Divalidasi!</p>
                <p class="text-xs text-slate-400 mt-1">Tidak ada desain spanduk yang menunggu review saat ini.</p>
              </td>
            </tr>

            <!-- Empty History -->
            <tr v-else-if="activeTab === 'HISTORY' && historyTasks.length === 0">
              <td colspan="6" class="py-16 text-center">
                <ShieldCheck class="w-12 h-12 text-slate-300 mx-auto mb-3" />
                <p class="text-base font-bold text-slate-700">Belum ada riwayat validasi.</p>
              </td>
            </tr>

            <!-- Data Rows -->
            <template v-else>
              <tr
                v-for="task in activeTab === 'WAITING' ? validationTasks : historyTasks"
                :key="task.id"
                class="hover:bg-slate-50/60 transition-colors group"
              >
                <!-- Tanggal & Toko -->
                <td class="py-4 px-6 align-middle max-w-[260px]">
                  <div class="font-extrabold text-slate-900 text-sm truncate">{{ task.shop_name }}</div>
                  <!-- Hover Card Alamat Toko -->
                  <StoreAddressHoverCard
                    :address="task.shop_address"
                    :shop-name="task.shop_name"
                    :province="task.province"
                    :regency="task.regency"
                    :district="task.district"
                    :village="task.village"
                    :phone="task.shop_phone"
                    :latitude="task.latitude"
                    :longitude="task.longitude"
                    :sales-name="task.sales_name"
                    :sales-branch="task.sales_branch"
                    class="mt-0.5 text-[11px]"
                  />
                  <div class="mt-1.5 flex flex-col gap-1">
                    <a
                      v-if="task.shop_phone"
                      :href="formatWhatsAppUrl(task.shop_phone, 'store', task)"
                      target="_blank"
                      class="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-600 hover:text-emerald-700 hover:underline w-fit"
                      title="Klik untuk chat WhatsApp Toko/Warung"
                    >
                      <MessageCircle class="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                      <span>{{ task.shop_phone }} (WA Toko)</span>
                    </a>
                    <span v-else class="text-[10px] text-slate-400 italic">
                      No. Toko: Tidak ada
                    </span>
                    <div class="flex items-center gap-1.5 text-[11px] font-semibold text-slate-500">
                      <Clock class="w-3.5 h-3.5 text-slate-400" />
                      <span>{{ task.created_at }}</span>
                    </div>
                  </div>
                </td>

                <!-- Detail Spanduk -->
                <td class="py-4 px-6 align-middle">
                  <div class="flex flex-col gap-1.5">
                    <div class="flex flex-wrap items-center gap-1.5">
                      <span
                        class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-black uppercase tracking-wider border"
                        :class="(task.type || '').toLowerCase() === 'universal' ? 'bg-sky-50 text-sky-700 border-sky-200' : 'bg-indigo-50 text-indigo-700 border-indigo-200'"
                      >
                        {{ task.type || 'Custom' }}
                      </span>
                      <span
                        v-if="task.is_bihun"
                        class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-black uppercase tracking-wider bg-amber-50 text-amber-700 border border-amber-200"
                      >
                        🍜 Bihun
                      </span>
                    </div>

                    <!-- Ukuran Request & Desain -->
                    <div class="space-y-1">
                      <div class="flex items-center gap-1 text-[11px] font-bold text-slate-600">
                        <span class="text-slate-400 font-medium">Req:</span>
                        <span class="bg-slate-100 text-slate-800 px-2 py-0.5 rounded-md border border-slate-200/60 font-extrabold">
                          {{ formatBannerDimension(task.size_length, task.size_width, 'dual') }}
                        </span>
                      </div>

                      <div v-if="getSizeSuggestion(task)" class="flex flex-col gap-0.5">
                        <div class="flex items-center gap-1 text-[11px] font-extrabold text-blue-700">
                          <span class="text-blue-500 font-semibold">Desain:</span>
                          <span class="bg-blue-50 text-blue-800 px-2 py-0.5 rounded-md border border-blue-200/80 inline-flex items-center gap-1 shadow-xs">
                            <Sparkles class="w-3 h-3 text-blue-500 shrink-0" />
                            {{ getSizeSuggestion(task)?.formattedDesignDual || `${getSizeSuggestion(task)?.designLength} × ${getSizeSuggestion(task)?.designWidth} cm` }}
                          </span>
                        </div>
                        <span class="text-[9px] font-bold text-blue-600/80 pl-0.5">
                          ⓘ {{ getSizeSuggestion(task)?.ruleText }}
                        </span>
                      </div>
                    </div>
                  </div>
                  <p v-if="task.notes && task.notes !== '-'" class="text-[11px] text-slate-500 mt-1.5 line-clamp-1 max-w-[200px]" :title="task.notes">
                    Catatan: {{ task.notes }}
                  </p>
                </td>

                <!-- Jumlah (Qty) -->
                <td class="py-4 px-6 align-middle text-center">
                  <div class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-100 text-slate-800 border border-slate-200/80 shadow-2xs">
                    <span class="text-xs font-black">{{ task.qty || 1 }}</span>
                    <span class="text-[10px] font-bold text-slate-500 uppercase">Pcs</span>
                  </div>
                </td>

                <!-- Editor & Sales -->
                <td class="py-4 px-6 align-middle">
                  <div class="space-y-1 text-xs">
                    <div class="font-medium text-slate-500">
                      Editor: <span class="font-bold text-slate-800">{{ task.editor_name || '-' }}</span>
                    </div>
                    <div>
                      <div class="font-medium text-slate-500">
                        Sales: <span class="font-bold text-slate-800">{{ task.sales_name || '-' }}</span>
                      </div>
                      <a
                        v-if="task.sales_phone"
                        :href="formatWhatsAppUrl(task.sales_phone, 'sales', task)"
                        target="_blank"
                        class="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-600 hover:text-emerald-700 hover:underline mt-0.5"
                        title="Klik untuk chat WhatsApp Sales"
                      >
                        <MessageCircle class="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                        <span>{{ task.sales_phone }} (WA Sales)</span>
                      </a>
                    </div>
                  </div>
                </td>

                <!-- Status Persetujuan (Dual Checklist Badge) -->
                <td class="py-4 px-6 align-middle">
                  <div class="flex flex-col gap-2">
                    <!-- ACC Desain -->
                    <div class="flex flex-col">
                      <div
                        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-bold w-fit border"
                        :class="task.is_design_approved ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-amber-50 text-amber-700 border-amber-200/60'"
                      >
                        <Palette class="w-3.5 h-3.5 shrink-0" />
                        <span>Desain: {{ task.is_design_approved ? 'Disetujui' : 'Menunggu' }}</span>
                        <Check v-if="task.is_design_approved" class="w-3.5 h-3.5 text-emerald-600 ml-0.5" />
                      </div>
                      <span
                        v-if="task.is_design_approved && (task.design_validator_name || task.validator_name)"
                        class="text-[10px] font-semibold text-slate-500 pl-1 mt-0.5"
                      >
                        ✓ {{ task.design_validator_name || task.validator_name }}
                      </span>
                    </div>

                    <!-- ACC Toko -->
                    <div class="flex flex-col">
                      <div
                        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-bold w-fit border"
                        :class="task.is_store_approved ? 'bg-blue-50 text-blue-700 border-blue-200' : 'bg-slate-100 text-slate-500 border-slate-200'"
                      >
                        <Store class="w-3.5 h-3.5 shrink-0" />
                        <span>Toko: {{ task.is_store_approved ? 'Disetujui' : 'Menunggu' }}</span>
                        <Check v-if="task.is_store_approved" class="w-3.5 h-3.5 text-blue-600 ml-0.5" />
                      </div>
                      <span
                        v-if="task.is_store_approved && (task.store_validator_name || task.validator_name)"
                        class="text-[10px] font-semibold text-slate-500 pl-1 mt-0.5"
                      >
                        ✓ {{ task.store_validator_name || task.validator_name }}
                      </span>
                    </div>
                  </div>
                </td>

                <!-- Aksi (Simpel & Elegan) -->
                <td class="py-4 px-6 align-middle text-center">
                  <div v-if="activeTab === 'WAITING'" class="flex items-center justify-center gap-2">
                    <Button
                      @click="openReviewModal(task)"
                      class="bg-primary hover:bg-primary/90 text-white rounded-xl font-extrabold text-xs px-3.5 py-2 shadow-sm shadow-primary/20 flex items-center gap-1.5"
                    >
                      <Eye class="w-3.5 h-3.5" />
                      {{ canValidate ? 'Review & Validasi' : 'Lihat Desain & Status' }}
                    </Button>
                  </div>
                  <div v-else class="flex flex-col items-center gap-2">
                    <div class="flex items-center justify-center gap-2">
                      <button
                        v-if="task.design_url"
                        @click="openReviewModal(task)"
                        class="p-2 text-indigo-600 hover:bg-indigo-50 rounded-xl transition"
                        title="Lihat Desain"
                      >
                        <Eye class="w-4 h-4" />
                      </button>
                      <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-emerald-50 text-emerald-600 text-xs font-black border border-emerald-100">
                        <CheckCircle2 class="w-3.5 h-3.5" /> Siap Cetak
                      </span>
                    </div>

                    <!-- Dual Validator Card in History -->
                    <div
                      v-if="task.design_validator_name || task.store_validator_name || task.validator_name"
                      class="bg-slate-50 border border-slate-200/80 rounded-xl p-2 text-left w-full max-w-[210px] space-y-1.5 shadow-2xs"
                    >
                      <!-- Desain Validator -->
                      <div class="flex items-start gap-1.5 text-[10px] leading-tight">
                        <Palette class="w-3 h-3 text-indigo-600 shrink-0 mt-0.5" />
                        <div class="min-w-0 flex-1">
                          <span class="text-slate-400 font-bold block text-[9px] uppercase tracking-wider">Validator Desain</span>
                          <span class="font-bold text-slate-800 truncate block">
                            {{ task.design_validator_name || task.validator_name }}
                          </span>
                        </div>
                      </div>
                      <!-- Toko Validator -->
                      <div class="flex items-start gap-1.5 text-[10px] leading-tight pt-1 border-t border-slate-200/60">
                        <Store class="w-3 h-3 text-blue-600 shrink-0 mt-0.5" />
                        <div class="min-w-0 flex-1">
                          <span class="text-slate-400 font-bold block text-[9px] uppercase tracking-wider">Validator Toko</span>
                          <span class="font-bold text-slate-800 truncate block">
                            {{ task.store_validator_name || task.validator_name }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="p-4 border-t border-slate-200 flex items-center justify-between bg-slate-50/50">
        <div class="text-xs font-bold text-slate-500">
          Halaman <span class="text-slate-800">{{ currentPage() }}</span> dari <span class="text-slate-800">{{ totalPages() }}</span>
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="prevPage"
            :disabled="!canPrevPage()"
            class="flex items-center justify-center w-8 h-8 rounded-xl border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            <ChevronLeft class="w-4 h-4" />
          </button>
          <button
            @click="nextPage"
            :disabled="!canNextPage()"
            class="flex items-center justify-center w-8 h-8 rounded-xl border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- ── Modal Terpadu: Review Desain & Keputusan Validasi ─────────────────── -->
    <div v-if="showReviewModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/70 backdrop-blur-sm" @click="closeReviewModal"></div>
      <div
        class="bg-white rounded-3xl w-full max-w-5xl max-h-[92vh] relative z-10 shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in duration-200"
      >
        <!-- Modal Header -->
        <div class="flex justify-between items-center px-6 py-4 border-b border-slate-200 bg-slate-50/60">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center">
              <ShieldCheck class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-lg font-black text-slate-900 leading-tight">
                Review & Validasi Desain Spanduk
              </h3>
              <p class="text-xs font-bold text-slate-500 flex flex-wrap items-center gap-1.5 mt-0.5">
                <span>Toko: <strong class="text-primary">{{ selectedTask?.shop_name }}</strong></span>
                <span class="text-slate-300">•</span>
                <span class="bg-slate-100 text-slate-700 px-2 py-0.5 rounded-md border border-slate-200/60 font-extrabold">
                  {{ formatBannerDimension(selectedTask?.size_length, selectedTask?.size_width, 'dual') }}
                </span>
                <span class="text-slate-300">•</span>
                <span class="bg-primary/10 text-primary px-2 py-0.5 rounded-md font-black">
                  {{ selectedTask?.qty || 1 }} Pcs
                </span>
                <span class="text-slate-300">•</span>
                <span class="text-slate-600 font-semibold">{{ selectedTask?.type }}</span>
                <span class="text-slate-300">•</span>
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-wider"
                  :class="selectedTask?.is_bihun ? 'bg-amber-100 text-amber-800' : 'bg-slate-100 text-slate-600'"
                >
                  {{ selectedTask?.is_bihun ? '🍜 Desain Bihun' : 'Bukan Bihun' }}
                </span>
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <a
              v-if="selectedTask?.design_url"
              :href="selectedTask.design_url"
              target="_blank"
              class="hidden sm:inline-flex items-center gap-1.5 text-xs font-bold text-primary bg-primary/10 hover:bg-primary/20 px-3 py-1.5 rounded-xl transition"
            >
              Buka di Tab Baru <ExternalLink class="w-3.5 h-3.5" />
            </a>
            <button
              @click="closeReviewModal"
              class="p-2 hover:bg-slate-200 rounded-xl text-slate-400 hover:text-slate-700 transition"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>

        <!-- Modal Body (Split: Preview & Control Panel) -->
        <div class="flex-1 flex flex-col lg:flex-row overflow-hidden bg-slate-100">
          <!-- Left/Top: Design Preview Area -->
          <div class="flex-1 min-h-[300px] lg:min-h-0 bg-slate-900 relative flex items-center justify-center p-3 sm:p-4">
            <iframe
              v-if="previewUrl"
              :src="previewUrl"
              class="w-full h-full rounded-2xl border-0 bg-white shadow-lg"
              allow="autoplay"
              allowfullscreen
            ></iframe>
            <div v-else class="text-center text-slate-400">
              <ImageIcon class="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p class="text-sm font-bold">Link desain belum diupload oleh editor.</p>
            </div>
          </div>

          <!-- Right/Bottom: Decision Action Panel -->
          <div class="w-full lg:w-96 bg-white border-t lg:border-t-0 lg:border-l border-slate-200 p-6 flex flex-col justify-between overflow-y-auto">
            <!-- Normal Decision State -->
            <div v-if="!isRevisionMode" class="space-y-6">
              <!-- Info Ukuran Request vs Sugesti Desain -->
              <div v-if="selectedTask && getSizeSuggestion(selectedTask)" class="bg-blue-50/80 border border-blue-200/80 rounded-2xl p-4 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[10px] font-black uppercase tracking-wider text-blue-700 flex items-center gap-1.5">
                    <Sparkles class="w-3.5 h-3.5 text-blue-600" /> Ukuran Kanvas Desain
                  </span>
                  <span class="text-[10px] font-extrabold text-blue-700 bg-white px-2 py-0.5 rounded-full border border-blue-200 shadow-xs">
                    {{ getSizeSuggestion(selectedTask)?.ruleText }}
                  </span>
                </div>
                <div class="grid grid-cols-2 gap-2 text-xs pt-1">
                  <div class="bg-white/90 p-2.5 rounded-xl border border-blue-100">
                    <p class="text-[10px] text-slate-400 font-bold uppercase">Req Asli Sales</p>
                    <p class="text-xs font-black text-slate-700 mt-0.5">
                      {{ formatBannerDimension(selectedTask?.size_length, selectedTask?.size_width, 'dual') }}
                      <span class="block text-[11px] font-extrabold text-primary mt-0.5">Qty: {{ selectedTask?.qty || 1 }} Pcs</span>
                    </p>
                  </div>
                  <div class="bg-blue-600 text-white p-2.5 rounded-xl shadow-xs">
                    <p class="text-[10px] text-blue-100 font-bold uppercase">Sugesti Desain</p>
                    <p class="text-xs font-black mt-0.5">{{ getSizeSuggestion(selectedTask)?.formattedDesignDual || `${getSizeSuggestion(selectedTask)?.designLength} × ${getSizeSuggestion(selectedTask)?.designWidth} cm` }}</p>
                  </div>
                </div>
              </div>

              <!-- Verifikasi Desain Bihun Interaktif -->
              <div class="p-3.5 rounded-2xl border transition-all" :class="selectedTask?.is_bihun ? 'bg-amber-50 border-amber-200/90 shadow-xs' : 'bg-slate-50/80 border-slate-200'">
                <div class="flex items-center justify-between gap-3">
                  <div class="flex items-center gap-2.5 min-w-0">
                    <span class="text-xl shrink-0">🍜</span>
                    <div class="min-w-0">
                      <div class="flex items-center gap-1.5 flex-wrap">
                        <span class="text-xs font-black text-slate-800">Kategori Desain Bihun</span>
                        <span
                          class="px-2 py-0.2 text-[9px] font-black uppercase tracking-wider rounded-md"
                          :class="selectedTask?.is_bihun ? 'bg-amber-200/80 text-amber-900' : 'bg-slate-200 text-slate-600'"
                        >
                          {{ selectedTask?.is_bihun ? 'Ada Produk Bihun' : 'Bukan Bihun' }}
                        </span>
                      </div>
                      <p class="text-[10px] text-slate-500 font-medium leading-tight mt-0.5">
                        Centang jika di dalam gambar desain terdapat produk Bihun
                      </p>
                    </div>
                  </div>
                  <!-- Switch Toggle Button -->
                  <label class="relative inline-flex items-center cursor-pointer shrink-0">
                    <input
                      type="checkbox"
                      v-model="selectedTask.is_bihun"
                      @change="handleToggleBihun(selectedTask)"
                      :disabled="!canValidate || isSubmitting"
                      class="sr-only peer"
                    />
                    <div class="w-11 h-6 bg-slate-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-amber-500"></div>
                  </label>
                </div>
              </div>

              <div>
                <h4 class="text-xs font-black text-slate-400 uppercase tracking-widest mb-3">
                  Status Kelengkapan Validasi
                </h4>
                
                <!-- Status List -->
                <div class="space-y-3">
                  <!-- 1. Desain Visual -->
                  <div class="p-3.5 rounded-2xl border transition-all" :class="selectedTask?.is_design_approved ? 'bg-emerald-50/70 border-emerald-200' : 'bg-slate-50 border-slate-200'">
                    <div class="flex items-center justify-between mb-2">
                      <div class="flex items-center gap-2 text-xs font-black text-slate-800">
                        <Palette class="w-4 h-4 text-indigo-600" /> 1. Validasi Desain Visual
                      </div>
                      <span
                        class="text-[10px] font-black uppercase px-2 py-0.5 rounded-full"
                        :class="selectedTask?.is_design_approved ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800'"
                      >
                        {{ selectedTask?.is_design_approved ? 'Disetujui' : 'Menunggu' }}
                      </span>
                    </div>
                    <!-- Validator info if approved -->
                    <div
                      v-if="selectedTask?.is_design_approved && (selectedTask?.design_validator_name || selectedTask?.validator_name)"
                      class="text-[11px] bg-white/80 p-2 rounded-xl border border-emerald-200/80 flex items-center justify-between mt-1 mb-1.5"
                    >
                      <span class="text-slate-600 font-semibold">
                        Di-ACC oleh: <strong class="text-emerald-800">{{ selectedTask.design_validator_name || selectedTask.validator_name }}</strong>
                      </span>
                      <span v-if="selectedTask.design_approved_at" class="text-[10px] text-slate-400 font-medium">{{ selectedTask.design_approved_at }}</span>
                    </div>
                    <div v-if="canValidateDesign && activeTab === 'WAITING'" class="flex gap-2 mt-2">
                      <button
                        v-if="!selectedTask?.is_design_approved"
                        @click="handleApproveDesign(selectedTask)"
                        :disabled="isSubmitting"
                        class="flex-1 bg-emerald-500 hover:bg-emerald-600 disabled:opacity-50 text-white py-2 rounded-xl text-xs font-black shadow-sm flex items-center justify-center gap-1.5 transition"
                      >
                        <Check class="w-3.5 h-3.5" /> ACC Desain
                      </button>
                      <button
                        @click="isRevisionMode = true; revisionType = 'design'"
                        :disabled="isSubmitting"
                        class="px-3 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl text-xs font-bold border border-rose-200/60 transition"
                      >
                        Revisi Desain
                      </button>
                    </div>
                  </div>

                  <!-- 2. Persetujuan Toko -->
                  <div class="p-3.5 rounded-2xl border transition-all" :class="selectedTask?.is_store_approved ? 'bg-blue-50/70 border-blue-200' : 'bg-slate-50 border-slate-200'">
                    <div class="flex items-center justify-between mb-2">
                      <div class="flex items-center gap-2 text-xs font-black text-slate-800">
                        <Store class="w-4 h-4 text-blue-600" /> 2. Persetujuan Toko
                      </div>
                      <span
                        class="text-[10px] font-black uppercase px-2 py-0.5 rounded-full"
                        :class="selectedTask?.is_store_approved ? 'bg-blue-100 text-blue-800' : 'bg-slate-200 text-slate-600'"
                      >
                        {{ selectedTask?.is_store_approved ? 'Disetujui' : 'Menunggu' }}
                      </span>
                    </div>
                    <!-- Validator info if approved -->
                    <div
                      v-if="selectedTask?.is_store_approved && (selectedTask?.store_validator_name || selectedTask?.validator_name)"
                      class="text-[11px] bg-white/80 p-2 rounded-xl border border-blue-200/80 flex items-center justify-between mt-1 mb-1.5"
                    >
                      <span class="text-slate-600 font-semibold">
                        Di-ACC oleh: <strong class="text-blue-800">{{ selectedTask.store_validator_name || selectedTask.validator_name }}</strong>
                      </span>
                      <span v-if="selectedTask.store_approved_at" class="text-[10px] text-slate-400 font-medium">{{ selectedTask.store_approved_at }}</span>
                    </div>
                    <div v-if="canValidateStore && activeTab === 'WAITING'" class="flex gap-2 mt-2">
                      <button
                        v-if="!selectedTask?.is_store_approved"
                        @click="handleApproveStore(selectedTask)"
                        :disabled="isSubmitting"
                        class="flex-1 bg-blue-500 hover:bg-blue-600 disabled:opacity-50 text-white py-2 rounded-xl text-xs font-black shadow-sm flex items-center justify-center gap-1.5 transition"
                      >
                        <Check class="w-3.5 h-3.5" /> ACC Toko
                      </button>
                      <button
                        @click="isRevisionMode = true; revisionType = 'store'"
                        :disabled="isSubmitting"
                        class="px-3 bg-orange-50 hover:bg-orange-100 text-orange-600 rounded-xl text-xs font-bold border border-orange-200/60 transition"
                      >
                        Revisi Toko
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Information Notes -->
              <div class="bg-slate-50 rounded-2xl p-3.5 border border-slate-200 text-xs space-y-1.5">
                <div class="flex items-center justify-between">
                  <div class="text-[10px] font-black uppercase text-slate-400 tracking-wider">Catatan Tambahan</div>
                  <span
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-extrabold border"
                    :class="selectedTask?.is_bihun ? 'bg-amber-50 text-amber-800 border-amber-200' : 'bg-slate-100 text-slate-600 border-slate-200'"
                  >
                    {{ selectedTask?.is_bihun ? '🍜 Desain Bihun' : 'Bukan Bihun' }}
                  </span>
                </div>
                <p class="font-medium text-slate-700">{{ selectedTask?.notes || '-' }}</p>
                <div class="text-[10px] text-slate-400 pt-1">
                  Req by: {{ selectedTask?.sales_name }} • Editor: {{ selectedTask?.editor_name }}
                </div>
              </div>

              <!-- Lampiran File / Logo / Dokumen Pendukung Sales -->
              <div v-if="selectedTask?.documentation_image_url" class="bg-indigo-50/70 border border-indigo-200/80 rounded-2xl p-3.5 text-xs space-y-2">
                <div class="flex items-center justify-between">
                  <div class="text-[10px] font-black uppercase text-indigo-900 tracking-wider flex items-center gap-1">
                    <Paperclip class="w-3.5 h-3.5 text-indigo-600" /> Lampiran Request Sales
                  </div>
                  <template v-if="selectedTask.documentation_image_url.toLowerCase().endsWith('.pdf')">
                    <a
                      :href="selectedTask.documentation_image_url.startsWith('http') ? selectedTask.documentation_image_url : '/api/v1' + selectedTask.documentation_image_url"
                      target="_blank"
                      class="inline-flex items-center gap-1 px-2.5 py-1 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-[10px] font-black transition shadow-xs"
                    >
                      <FileDown class="w-3 h-3" /> Buka PDF
                    </a>
                  </template>
                  <template v-else>
                    <a
                      :href="selectedTask.documentation_image_url.startsWith('http') ? selectedTask.documentation_image_url : '/api/v1' + selectedTask.documentation_image_url"
                      target="_blank"
                      class="inline-flex items-center gap-1 px-2.5 py-1 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-[10px] font-black transition shadow-xs"
                    >
                      <ExternalLink class="w-3 h-3" /> Buka Foto/Logo
                    </a>
                  </template>
                </div>
              </div>

              <!-- Contact & Chat WhatsApp Card -->
              <div class="bg-slate-50 rounded-2xl p-3.5 border border-slate-200 text-xs space-y-2.5">
                <div class="text-[10px] font-black uppercase text-slate-400 tracking-wider">Kontak & Chat WhatsApp</div>
                
                <!-- Chat Toko -->
                <div class="flex items-center justify-between gap-2 p-2.5 rounded-xl bg-white border border-slate-200/80 shadow-xs">
                  <div class="min-w-0 flex-1">
                    <div class="text-[11px] font-extrabold text-slate-800 truncate">{{ selectedTask?.shop_name }}</div>
                    <div class="text-[10px] text-slate-500 font-medium">{{ selectedTask?.shop_phone || 'No. telp tidak tersedia' }}</div>
                  </div>
                  <a
                    v-if="selectedTask?.shop_phone"
                    :href="formatWhatsAppUrl(selectedTask.shop_phone, 'store', selectedTask)"
                    target="_blank"
                    class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-xl bg-emerald-50 hover:bg-emerald-100 text-emerald-700 text-xs font-black transition border border-emerald-200 shrink-0 shadow-xs"
                    title="Chat WhatsApp Toko"
                  >
                    <MessageCircle class="w-3.5 h-3.5 text-emerald-600" /> WA Toko
                  </a>
                </div>

                <!-- Chat Sales -->
                <div class="flex items-center justify-between gap-2 p-2.5 rounded-xl bg-white border border-slate-200/80 shadow-xs">
                  <div class="min-w-0 flex-1">
                    <div class="text-[11px] font-extrabold text-slate-800 truncate">{{ selectedTask?.sales_name }} <span class="text-[10px] text-slate-400 font-normal">(Sales)</span></div>
                    <div class="text-[10px] text-slate-500 font-medium">{{ selectedTask?.sales_phone || 'No. telp tidak tersedia' }}</div>
                  </div>
                  <a
                    v-if="selectedTask?.sales_phone"
                    :href="formatWhatsAppUrl(selectedTask.sales_phone, 'sales', selectedTask)"
                    target="_blank"
                    class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-xl bg-emerald-50 hover:bg-emerald-100 text-emerald-700 text-xs font-black transition border border-emerald-200 shrink-0 shadow-xs"
                    title="Chat WhatsApp Sales"
                  >
                    <MessageCircle class="w-3.5 h-3.5 text-emerald-600" /> WA Sales
                  </a>
                </div>
              </div>

              <!-- Notice for Read-Only (Non-Validator) -->
              <div v-if="!canValidate" class="p-3.5 bg-amber-50 border border-amber-200/80 rounded-2xl text-xs text-amber-800 flex items-start gap-2.5">
                <ShieldCheck class="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
                <span class="leading-relaxed">
                  <strong>Mode Pantau:</strong> Anda dapat melihat pratinjau desain. Keputusan ACC Desain dilakukan oleh <strong>Validator Desain</strong> dan Persetujuan Toko dilakukan oleh <strong>Validator Toko</strong>.
                </span>
              </div>

              <!-- 1-Click Approve Both (Hanya SuperAdmin / Wewenang Penuh) -->
              <div v-if="canValidateBoth && activeTab === 'WAITING' && (!selectedTask?.is_design_approved || !selectedTask?.is_store_approved)" class="pt-2">
                <button
                  @click="handleApproveBoth(selectedTask)"
                  :disabled="isSubmitting"
                  class="w-full py-3.5 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 text-white rounded-2xl font-black text-xs shadow-md shadow-emerald-500/20 flex items-center justify-center gap-2 transition active:scale-[0.98]"
                >
                  <Sparkles class="w-4 h-4 text-amber-200" />
                  ACC Keduanya & Lanjut Cetak
                </button>
              </div>
            </div>

            <!-- Revision Mode State -->
            <div v-else class="space-y-4 animate-in fade-in duration-150">
              <div class="flex items-center justify-between pb-2 border-b border-slate-100">
                <div class="flex items-center gap-2 text-rose-600 font-extrabold text-sm">
                  <AlertTriangle class="w-4 h-4" /> Form Permintaan Revisi
                </div>
                <button
                  @click="isRevisionMode = false"
                  class="text-xs font-bold text-slate-400 hover:text-slate-600"
                >
                  Batal
                </button>
              </div>

              <div>
                <label class="block text-xs font-bold text-slate-600 uppercase tracking-wider mb-2">
                  Kategori Revisi
                </label>
                <!-- Jika superadmin, bisa memilih antara Revisi Desain / Toko. Jika validator spesifik, terkunci sesuai role -->
                <div v-if="canValidateBoth" class="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    @click="revisionType = 'design'"
                    :class="[
                      'py-2 px-3 rounded-xl text-xs font-black border transition-all text-center',
                      revisionType === 'design'
                        ? 'bg-rose-50 border-rose-500 text-rose-700'
                        : 'bg-white border-slate-200 text-slate-500 hover:bg-slate-50'
                    ]"
                  >
                    Revisi Desain
                  </button>
                  <button
                    type="button"
                    @click="revisionType = 'store'"
                    :class="[
                      'py-2 px-3 rounded-xl text-xs font-black border transition-all text-center',
                      revisionType === 'store'
                        ? 'bg-orange-50 border-orange-500 text-orange-700'
                        : 'bg-white border-slate-200 text-slate-500 hover:bg-slate-50'
                    ]"
                  >
                    Revisi Toko
                  </button>
                </div>
                <div v-else class="p-2.5 rounded-xl border font-black text-xs text-center" :class="revisionType === 'design' ? 'bg-rose-50 border-rose-200 text-rose-700' : 'bg-orange-50 border-orange-200 text-orange-700'">
                  {{ revisionType === 'design' ? '🎨 Kategori: Revisi Desain Visual' : '🏪 Kategori: Revisi Persetujuan Toko' }}
                </div>
              </div>

              <div>
                <label class="block text-xs font-bold text-slate-600 uppercase tracking-wider mb-2">
                  Catatan untuk Editor <span class="text-rose-500">*</span>
                </label>
                <textarea
                  v-model="revisionNote"
                  rows="4"
                  placeholder="Tuliskan detail perbaikan yang diperlukan..."
                  class="w-full p-3.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium text-slate-800 focus:bg-white focus:outline-none focus:border-rose-400 transition resize-none"
                ></textarea>
              </div>

              <div class="pt-2 flex gap-2">
                <button
                  @click="isRevisionMode = false"
                  type="button"
                  class="w-1/3 py-3 border border-slate-200 rounded-xl text-xs font-bold text-slate-600 hover:bg-slate-50 transition"
                >
                  Kembali
                </button>
                <button
                  @click="submitRevision"
                  :disabled="isSubmitting || !revisionNote.trim()"
                  type="button"
                  class="w-2/3 py-3 bg-rose-600 hover:bg-rose-700 disabled:bg-slate-200 disabled:text-slate-400 text-white rounded-xl text-xs font-black shadow-md shadow-rose-600/20 flex items-center justify-center gap-1.5 transition"
                >
                  <RefreshCw v-if="isSubmitting" class="w-3.5 h-3.5 animate-spin" />
                  <Send v-else class="w-3.5 h-3.5" />
                  Kirim ke Editor
                </button>
              </div>
            </div>

            <!-- Footer Close Button -->
            <div class="pt-4 mt-auto border-t border-slate-100">
              <Button
                variant="outline"
                @click="closeReviewModal"
                class="w-full rounded-xl text-xs font-bold text-slate-500 hover:bg-slate-100"
              >
                Tutup Modal
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
