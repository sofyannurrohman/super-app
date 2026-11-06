<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useAuthStore } from '@/stores/auth'
import { useMasterDataStore } from '@/stores/masterData'
import { resolvePhotoUrl } from '@/utils/imageUrl'
import { calculateBannerDesignSize, formatBannerDimension } from '@/utils/bannerSize'
import StoreAddressHoverCard from '@/components/StoreAddressHoverCard.vue'
import {
  Brush,
  Clock,
  CheckCircle2,
  Upload,
  X,
  AlertTriangle,
  RefreshCw,
  WifiOff,
  Search,
  ChevronLeft,
  ChevronRight,
  Inbox,
  Hourglass,
  BadgeCheck,
  FileImage,
  Ruler,
  User,
  Store,
  Eye,
  ExternalLink,
  Copy,
  Check,
  MessageCircle,
  Info,
  ChevronDown,
  ChevronUp,
  FileText,
  Sparkles,
  Paperclip,
  FileDown,
  MapPin,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  SlidersHorizontal,
  Building2,
} from 'lucide-vue-next'

// ── Types ──────────────────────────────────────────────────────────────────
interface Banner {
  id: number
  shop_name: string
  shop_address?: string
  shop_photo?: string | null
  shop_phone?: string | null
  sales_name: string
  type: string
  qty?: number
  size_length: number | string
  size_width: number | string
  status: string
  notes: string | null
  revision_note: string | null
  created_at: string
  design_file_url?: string | null
  design_link?: string | null
  design_url?: string | null
  design_file_name?: string | null
  documentation_image_url?: string | null
  editor_name?: string | null
  editor_id?: number | string | null
  editor_user_id?: string | null
  sales_branch?: string | null
  sales_phone?: string | null
  province?: string | null
  regency?: string | null
  district?: string | null
  village?: string | null
  latitude?: number | string | null
  longitude?: number | string | null
}

interface AIDesignJob {
  id: string
  status: 'queued_hermes' | 'hermes_processing' | 'awaiting_corel' | 'corel_processing' | 'completed' | 'failed'
  reference_image_url?: string | null
  svg_url?: string | null
  preview_url?: string | null
  cdr_url?: string | null
  cdr_file_name?: string | null
  error_message?: string | null
}

interface RequestError {
  response?: { data?: { error?: string } }
  message?: string
  code?: string
  isAxiosError?: boolean
}

// ── State ──────────────────────────────────────────────────────────────────
const activeTab = ref<'request' | 'editing' | 'validation' | 'done'>('request')

const isLoading = ref(false)
const isOffline = ref(false)

const banners = ref<Banner[]>([])
const page = ref(1)
const limit = ref(10)
const search = ref('')
const total = ref(0)

// Filter & Sort State
const selectedBranchFilter = computed<string>({
  get: () => masterStore.selectedBrandingBranchId || 'ALL_BRANCHES',
  set: (val: string) => {
    masterStore.setSelectedBrandingBranchId(val)
  },
})
const selectedEditorFilter = ref<string>('all') // 'all' | 'my_tasks' | 'unassigned' | editor_id
const onlyRevision = ref<boolean>(false)
const sortBy = ref<'created_at' | 'editor_name' | 'shop_name' | 'revision'>('created_at')
const sortDir = ref<'ASC' | 'DESC'>('DESC')

const selectedBranchName = computed(() => {
  return masterStore.selectedBrandingBranchName
})

// Expanded notes & copy state
const expandedNotes = ref<Record<number, boolean>>({})
const copiedNoteId = ref<number | null>(null)

// Detail modal state
const showDetailModal = ref(false)
const selectedBannerForDetail = ref<Banner | null>(null)

// Revision detail modal state
const showRevisionModal = ref(false)
const selectedBannerForRevision = ref<Banner | null>(null)

// Preview Photo state
const previewPhoto = ref<string | null>(null)
const previewTitle = ref('Pratinjau Foto Toko')

// Upload modal
const showUploadModal = ref(false)
const selectedBanner = ref<Banner | null>(null)
const uploadFile = ref<File | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const isUploading = ref(false)
const uploadDragOver = ref(false)
const uploadProgress = ref(0)
const uploadStage = ref<'idle' | 'uploading' | 'processing_gdrive' | 'success'>('idle')
const uploadedBytes = ref(0)
const totalBytes = ref(0)

// AI draft jobs are intentionally separate from the final upload/validation status.
const aiDrafts = ref<Record<string, AIDesignJob | null>>({})
const aiGenerating = ref<Record<string, boolean>>({})
const aiInstructions = ref<Record<string, string>>({})

// Assign modal & editors state
const showAssignModal = ref(false)
const selectedBannerForAssign = ref<Banner | null>(null)
const selectedEditorId = ref<number | string | null>(null)
const isAssigning = ref(false)
const editors = ref<{ id: number; name: string; email?: string }[]>([])

const authStore = useAuthStore()
const masterStore = useMasterDataStore()

const currentEditor = computed(() => {
  if (!authStore.user || !authStore.user.email) return null
  return (
    editors.value.find((e) => e.email?.toLowerCase() === authStore.user.email.toLowerCase()) ||
    null
  )
})

// Hanya admin murni (bukan admin_designer) yang bisa assign tugas ke orang lain
const isSuperAdmin = computed(() => {
  return authStore.user && ['superadmin', 'admin', 'hr_admin', 'HR'].includes(authStore.user.role)
})

// admin_designer: bisa ambil tugas sendiri & upload jika sudah di-assign ke dirinya
const isDesignerAdmin = computed(() => {
  return authStore.user?.role === 'admin_designer'
})

// ── Helpers ────────────────────────────────────────────────────────────────
const formatDate = (dateStr: string | null | undefined): string => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    return d.toLocaleDateString('id-ID', {
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

const formatSize = (banner: Banner): string => {
  if (!banner) return '-'
  return formatBannerDimension(banner.size_length, banner.size_width, 'dual')
}

const getSizeSuggestion = (banner: Banner) => {
  return calculateBannerDesignSize(banner.size_length, banner.size_width)
}

const photoUrl = (path: string | null | undefined) => {
  return resolvePhotoUrl(path)
}

const openPhotoPreview = (url: string | null | undefined, title = 'Pratinjau Foto Toko') => {
  if (!url) return
  previewPhoto.value = photoUrl(url)
  previewTitle.value = title
}

const toggleExpandNote = (id: number) => {
  expandedNotes.value[id] = !expandedNotes.value[id]
}

const copyToClipboard = async (text: string | null | undefined, id?: number) => {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    if (id) {
      copiedNoteId.value = id
      setTimeout(() => {
        if (copiedNoteId.value === id) copiedNoteId.value = null
      }, 2000)
    }
    toast.success('Materi catatan berhasil disalin ke clipboard!')
  } catch {
    toast.error('Gagal menyalin catatan')
  }
}

const getBannerFullAddress = (banner: Banner | null | undefined): string => {
  if (!banner) return 'Alamat tidak tersedia'
  const parts: string[] = []
  if (banner.shop_address && banner.shop_address.trim()) {
    parts.push(banner.shop_address.trim())
  }
  const adminParts = [banner.village, banner.district, banner.regency, banner.province].filter(
    (p) => !!p && p.trim() !== '',
  )
  if (adminParts.length > 0) {
    const adminStr = adminParts.join(', ')
    if (!banner.shop_address || !banner.shop_address.toLowerCase().includes(adminParts[0]?.toLowerCase() || '')) {
      parts.push(adminStr)
    }
  }
  if (parts.length > 0) {
    return parts.join(', ')
  }
  const lat = Number(banner.latitude)
  const lng = Number(banner.longitude)
  if (!isNaN(lat) && !isNaN(lng) && (lat !== 0 || lng !== 0)) {
    return `Koordinat (${banner.latitude}, ${banner.longitude})`
  }
  return 'Alamat tidak tersedia'
}

const openDetailModal = (banner: Banner) => {
  selectedBannerForDetail.value = banner
  showDetailModal.value = true
}

const closeDetailModal = () => {
  showDetailModal.value = false
  selectedBannerForDetail.value = null
}

const openRevisionModal = (banner: Banner) => {
  selectedBannerForRevision.value = banner
  showRevisionModal.value = true
}

const closeRevisionModal = () => {
  showRevisionModal.value = false
  selectedBannerForRevision.value = null
}

const getRevisionLines = (note: string | null | undefined): string[] => {
  if (!note) return []
  return note
    .split(/\n+/)
    .map((s) => s.trim())
    .filter((s) => s.length > 0)
}

const getRevisionTypeInfo = (note: string | null | undefined) => {
  if (!note) return { label: 'Revisi', isDesign: false, isShop: false }
  const lower = note.toLowerCase()
  if (lower.includes('[revisi desain]')) {
    return { label: 'Revisi Desain', isDesign: true, isShop: false }
  }
  if (lower.includes('[revisi toko]')) {
    return { label: 'Revisi Toko', isDesign: false, isShop: true }
  }
  return { label: 'Perlu Revisi', isDesign: false, isShop: false }
}

const formatWhatsAppUrl = (phone: string | null | undefined, banner: Banner) => {
  if (!phone) return '#'
  let clean = phone.replace(/[^0-9]/g, '')
  if (clean.startsWith('0')) {
    clean = '62' + clean.slice(1)
  }
  const msg = encodeURIComponent(
    `Halo Mas/Mbak ${banner.sales_name || 'Sales'}, saya mau konfirmasi desain spanduk untuk ${banner.shop_name || 'Toko'} (Ukuran: ${formatSize(banner)}).`,
  )
  return `https://wa.me/${clean}?text=${msg}`
}

const aiJobKey = (banner: Banner) => String(banner.id)
const requestError = (error: unknown): RequestError => error as RequestError

const aiStatusLabel = (status: AIDesignJob['status']): string => {
  const labels: Record<AIDesignJob['status'], string> = {
    queued_hermes: 'Antrean Hermes',
    hermes_processing: 'Hermes menyusun desain',
    awaiting_corel: 'Menunggu CorelDRAW',
    corel_processing: 'Membuat CDR',
    completed: 'CDR siap',
    failed: 'Gagal diproses',
  }
  return labels[status] || status
}

const fetchAIDraft = async (banner: Banner) => {
  try {
    const res = await apiClient.get(`/branding/banners/${banner.id}/ai-draft`)
    aiDrafts.value[aiJobKey(banner)] = res.data?.data || null
  } catch {
    // Status draft is supplementary; a failure must not hide the editor's task list.
  }
}

const fetchVisibleAIDrafts = async () => {
  if (activeTab.value !== 'editing') return
  await Promise.all(banners.value.map((banner) => fetchAIDraft(banner)))
}

const generateAIDraft = async (banner: Banner) => {
  const key = aiJobKey(banner)
  aiGenerating.value[key] = true
  try {
    const res = await apiClient.post(`/branding/banners/${banner.id}/ai-draft`, {
      instructions: (aiInstructions.value[key] || '').trim(),
    })
    aiDrafts.value[key] = res.data?.data || null
    toast.success('Draft masuk antrean Hermes di Mac mini.')
  } catch (error: unknown) {
    const e = requestError(error)
    toast.error('Gagal membuat draft AI: ' + (e.response?.data?.error || e.message))
    await fetchAIDraft(banner)
  } finally {
    aiGenerating.value[key] = false
  }
}

// ── API Calls ──────────────────────────────────────────────────────────────
const fetchBanners = async () => {
  isLoading.value = true
  isOffline.value = false
  try {
    const branchId = selectedBranchFilter.value
    let editorIdParam: string | undefined = undefined
    if (selectedEditorFilter.value === 'my_tasks') {
      if (currentEditor.value) {
        editorIdParam = String(currentEditor.value.id)
      } else if (authStore.user?.full_name) {
        editorIdParam = authStore.user.full_name
      }
    } else if (selectedEditorFilter.value === 'unassigned') {
      editorIdParam = 'unassigned'
    } else if (selectedEditorFilter.value !== 'all') {
      editorIdParam = selectedEditorFilter.value
    }

    const res = await apiClient.get('/branding/banners', {
      params: {
        status: activeTab.value,
        page: page.value,
        limit: limit.value,
        search: search.value,
        branch_id: branchId && branchId !== 'ALL_BRANCHES' ? branchId : undefined,
        editor_id: editorIdParam,
        only_revision: onlyRevision.value ? 'true' : undefined,
        sort_by: sortBy.value,
        sort_dir: sortDir.value,
      },
    })

    banners.value = res.data?.data || []
    total.value = res.data?.total || 0
    await fetchVisibleAIDrafts()
  } catch (error: unknown) {
    const e = requestError(error)
    isOffline.value = true
    toast.error('Gagal memuat data banner: ' + (e.response?.data?.error || e.message))
  } finally {
    isLoading.value = false
  }
}

const handleBranchFilterChange = () => {
  page.value = 1
  fetchBanners()
}

watch(() => masterStore.selectedBrandingBranchId, () => {
  handleBranchFilterChange()
})

const toggleSort = (column: 'created_at' | 'editor_name' | 'shop_name' | 'revision') => {
  if (sortBy.value === column) {
    sortDir.value = sortDir.value === 'ASC' ? 'DESC' : 'ASC'
  } else {
    sortBy.value = column
    sortDir.value = column === 'editor_name' || column === 'shop_name' ? 'ASC' : 'DESC'
  }
  page.value = 1
  fetchBanners()
}

const handleEditorFilterChange = () => {
  page.value = 1
  fetchBanners()
}

const toggleOnlyRevision = () => {
  onlyRevision.value = !onlyRevision.value
  page.value = 1
  fetchBanners()
}

const handleSortChange = (e: Event) => {
  const val = (e.target as HTMLSelectElement).value
  if (val === 'revision') {
    sortBy.value = 'revision'
    sortDir.value = 'DESC'
  } else if (val === 'editor_asc') {
    sortBy.value = 'editor_name'
    sortDir.value = 'ASC'
  } else if (val === 'editor_desc') {
    sortBy.value = 'editor_name'
    sortDir.value = 'DESC'
  } else if (val === 'shop_asc') {
    sortBy.value = 'shop_name'
    sortDir.value = 'ASC'
  } else if (val === 'shop_desc') {
    sortBy.value = 'shop_name'
    sortDir.value = 'DESC'
  } else if (val === 'date_asc') {
    sortBy.value = 'created_at'
    sortDir.value = 'ASC'
  } else {
    sortBy.value = 'created_at'
    sortDir.value = 'DESC'
  }
  page.value = 1
  fetchBanners()
}

const currentSortValue = computed(() => {
  if (sortBy.value === 'revision') return 'revision'
  if (sortBy.value === 'editor_name') return sortDir.value === 'ASC' ? 'editor_asc' : 'editor_desc'
  if (sortBy.value === 'shop_name') return sortDir.value === 'ASC' ? 'shop_asc' : 'shop_desc'
  return sortDir.value === 'ASC' ? 'date_asc' : 'date_desc'
})

const revisionCount = computed(() => {
  return banners.value.filter((b) => !!b.revision_note).length
})

let searchTimeout: ReturnType<typeof setTimeout> | undefined
const onSearchInput = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    page.value = 1
    fetchBanners()
  }, 350)
}

const totalPages = computed(() => Math.ceil(total.value / limit.value) || 1)

const prevPage = () => {
  if (page.value > 1) {
    page.value--
    fetchBanners()
  }
}

const nextPage = () => {
  if (page.value < totalPages.value) {
    page.value++
    fetchBanners()
  }
}

const handleTabChange = (tab: 'request' | 'editing' | 'validation' | 'done') => {
  activeTab.value = tab
  page.value = 1
  fetchBanners()
}

// ── Upload ─────────────────────────────────────────────────────────────────
const openUploadModal = (banner: Banner) => {
  selectedBanner.value = banner
  uploadFile.value = null
  uploadProgress.value = 0
  uploadStage.value = 'idle'
  uploadedBytes.value = 0
  totalBytes.value = 0
  showUploadModal.value = true
}

const closeUploadModal = () => {
  if (isUploading.value) {
    toast.warning('Proses upload sedang berjalan, mohon tunggu hingga selesai.')
    return
  }
  showUploadModal.value = false
  selectedBanner.value = null
  uploadFile.value = null
  uploadProgress.value = 0
  uploadStage.value = 'idle'
}

const onFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files[0]) {
    const file = target.files[0]
    if (file.size > 100 * 1024 * 1024) {
      toast.warning('Ukuran file maksimal 100 MB. Mohon kompres file terlebih dahulu.')
      return
    }
    uploadFile.value = file
  }
}

const onDrop = (e: DragEvent) => {
  uploadDragOver.value = false
  const file = e.dataTransfer?.files[0]
  if (file) {
    if (file.size > 100 * 1024 * 1024) {
      toast.warning('Ukuran file maksimal 100 MB. Mohon kompres file terlebih dahulu.')
      return
    }
    uploadFile.value = file
  }
}

const fileSizeLabel = (file: File): string => {
  if (file.size < 1024) return `${file.size} B`
  if (file.size < 1024 * 1024) return `${(file.size / 1024).toFixed(1)} KB`
  return `${(file.size / 1024 / 1024).toFixed(2)} MB`
}

const submitDesign = async () => {
  if (!selectedBanner.value || !uploadFile.value) {
    toast.warning('Pilih file desain terlebih dahulu.')
    return
  }

  const bannerId = selectedBanner.value.id
  const formData = new FormData()
  formData.append('design_file', uploadFile.value)

  isUploading.value = true
  uploadProgress.value = 0
  uploadStage.value = 'uploading'
  uploadedBytes.value = 0
  totalBytes.value = uploadFile.value.size

  try {
    await apiClient.put(`/branding/banners/${bannerId}/design`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 600000, // 10 menit khusus upload file besar ke GDrive
      onUploadProgress: (progressEvent) => {
        if (progressEvent.total) {
          totalBytes.value = progressEvent.total
          uploadedBytes.value = progressEvent.loaded
          const percentCompleted = Math.min(99, Math.round((progressEvent.loaded * 100) / progressEvent.total))
          uploadProgress.value = percentCompleted
          if (progressEvent.loaded >= progressEvent.total) {
            uploadProgress.value = 100
            uploadStage.value = 'processing_gdrive'
          }
        }
      },
    })
    uploadStage.value = 'success'
    toast.success('Desain berhasil diunggah ke Google Drive dan dikirim ke Validator!')
    setTimeout(async () => {
      closeUploadModal()
      activeTab.value = 'validation'
      page.value = 1
      await fetchBanners()
    }, 600)
  } catch (error: unknown) {
    const e = requestError(error)
    uploadStage.value = 'idle'
    uploadProgress.value = 0
    let errMsg = e.response?.data?.error || e.message
    if (e.code === 'ECONNABORTED' || e.message?.includes('timeout')) {
      errMsg = 'Proses upload melebihi batas waktu (timeout). Pastikan koneksi internet stabil dan coba kembali.'
    } else if (e.message === 'Network Error' || (!e.response && e.isAxiosError)) {
      errMsg = 'Koneksi terputus saat mengunggah (Network Error). Pastikan ukuran file tidak melebihi 100MB dan koneksi internet stabil, lalu coba kembali.'
    }
    toast.error('Gagal mengunggah desain: ' + errMsg)
  } finally {
    isUploading.value = false
  }
}

// ── Assign Editor ──────────────────────────────────────────────────────────
const fetchEditors = async () => {
  try {
    const res = await apiClient.get('/branding/editors')
    editors.value = res.data?.data || []
  } catch (error: unknown) {
    console.error('Gagal mengambil daftar editor:', error)
  }
}

const openAssignModal = (banner: Banner) => {
  selectedBannerForAssign.value = banner
  selectedEditorId.value = banner.editor_id || null
  showAssignModal.value = true
}

const closeAssignModal = () => {
  showAssignModal.value = false
  selectedBannerForAssign.value = null
  selectedEditorId.value = null
}

const submitAssign = async () => {
  if (!selectedBannerForAssign.value || !selectedEditorId.value) {
    toast.warning('Silakan pilih desainer terlebih dahulu.')
    return
  }

  const bannerId = selectedBannerForAssign.value.id
  isAssigning.value = true
  try {
    const res = await apiClient.put(`/branding/banners/${bannerId}/assign`, {
      editor_id: selectedEditorId.value,
    })
    toast.success(res.data?.message || 'Tugas desainer berhasil diperbarui!')
    closeAssignModal()
    await fetchBanners()
  } catch (error: unknown) {
    const e = requestError(error)
    toast.error('Gagal menugaskan desainer: ' + (e.response?.data?.error || e.message))
  } finally {
    isAssigning.value = false
  }
}

const claimTask = async (banner: Banner) => {
  try {
    const res = await apiClient.put(`/branding/banners/${banner.id}/claim`)
    toast.success(res.data?.message || 'Tugas berhasil diambil!')
    await fetchBanners()
  } catch (error: unknown) {
    const e = requestError(error)
    toast.error('Gagal mengambil tugas: ' + (e.response?.data?.error || e.message))
  }
}

// ── Lifecycle ─────────────────────────────────────────────────────────────
onMounted(async () => {
  if (!masterStore.branches || masterStore.branches.length === 0) {
    await masterStore.fetchBranches()
  }
  fetchBanners()
  fetchEditors()
})

const aiStatusTimer = window.setInterval(() => {
  const hasRunningJob = Object.values(aiDrafts.value).some(
    (job) => job && !['completed', 'failed'].includes(job.status),
  )
  if (hasRunningJob) void fetchVisibleAIDrafts()
}, 5000)

onUnmounted(() => window.clearInterval(aiStatusTimer))
</script>

<template>
  <div class="space-y-6 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <div class="flex items-center flex-wrap gap-2 sm:gap-3">
          <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <span>Editor Desain</span>
            <span class="text-primary text-2xl font-bold">
              — {{ selectedBranchName }}
            </span>
          </h1>
          <span
            v-if="!isDesignerAdmin && !isSuperAdmin"
            class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-black bg-slate-100 text-slate-600 border border-slate-200 shadow-sm"
          >
            <Eye class="w-3.5 h-3.5 text-slate-400" /> Mode Pantau
          </span>
        </div>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <Brush class="w-4 h-4 text-primary" />
          Workspace Editor — proses dan unggah desain spanduk dari request Sales.
        </p>
      </div>
      <button
        @click="fetchBanners"
        :disabled="isLoading"
        class="flex items-center gap-2 px-5 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-2xl text-sm font-bold transition-all disabled:opacity-60"
      >
        <RefreshCw :class="['w-4 h-4', isLoading ? 'animate-spin' : '']" />
        Refresh
      </button>
    </div>

    <!-- Offline Banner -->
    <div
      v-if="isOffline"
      class="flex items-center gap-3 bg-rose-50 border border-rose-200 text-rose-700 px-5 py-4 rounded-2xl text-sm font-bold"
    >
      <WifiOff class="w-5 h-5 shrink-0" />
      <span>Koneksi ke server gagal. Data mungkin tidak up-to-date. Coba refresh kembali.</span>
    </div>

    <!-- Controls: Tabs, Filters, Sort & Search -->
    <div class="space-y-4">
      <!-- Top Row: Tabs & Search -->
      <div class="flex flex-col lg:flex-row gap-4 items-stretch lg:items-center justify-between">
        <!-- Tabs -->
        <div class="flex gap-1 bg-slate-100 p-1 rounded-2xl overflow-x-auto">
          <button
            @click="handleTabChange('request')"
            :class="[
              'flex items-center gap-2 px-4 sm:px-5 py-2.5 rounded-xl text-sm font-bold transition-all whitespace-nowrap',
              activeTab === 'request'
                ? 'bg-white text-primary shadow-sm'
                : 'text-slate-500 hover:text-slate-700',
            ]"
          >
            <Brush class="w-4 h-4" />
            Antrean Request
          </button>

          <button
            @click="handleTabChange('editing')"
            :class="[
              'flex items-center gap-2 px-4 sm:px-5 py-2.5 rounded-xl text-sm font-bold transition-all whitespace-nowrap',
              activeTab === 'editing'
                ? 'bg-white text-purple-600 shadow-sm'
                : 'text-slate-500 hover:text-slate-700',
            ]"
          >
            <Clock class="w-4 h-4" />
            Sedang Dikerjakan
            <span
              v-if="activeTab === 'editing' && revisionCount > 0"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-black bg-rose-100 text-rose-700 border border-rose-200"
              title="Jumlah banner revisi pada halaman ini"
            >
              <AlertTriangle class="w-3 h-3" />
              {{ revisionCount }} Revisi
            </span>
          </button>

          <button
            @click="handleTabChange('validation')"
            :class="[
              'flex items-center gap-2 px-4 sm:px-5 py-2.5 rounded-xl text-sm font-bold transition-all whitespace-nowrap',
              activeTab === 'validation'
                ? 'bg-white text-orange-600 shadow-sm'
                : 'text-slate-500 hover:text-slate-700',
            ]"
          >
            <Hourglass class="w-4 h-4" />
            Menunggu Validasi
          </button>

          <button
            @click="handleTabChange('done')"
            :class="[
              'flex items-center gap-2 px-4 sm:px-5 py-2.5 rounded-xl text-sm font-bold transition-all whitespace-nowrap',
              activeTab === 'done'
                ? 'bg-white text-emerald-600 shadow-sm'
                : 'text-slate-500 hover:text-slate-700',
            ]"
          >
            <BadgeCheck class="w-4 h-4" />
            Selesai
          </button>
        </div>

        <!-- Search -->
        <div class="relative w-full lg:w-72 shrink-0">
          <Search class="w-5 h-5 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            v-model="search"
            @input="onSearchInput"
            type="text"
            placeholder="Cari toko atau sales..."
            class="w-full pl-10 pr-4 py-2.5 bg-white border border-slate-200 rounded-2xl text-sm font-medium focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all placeholder:text-slate-400 shadow-sm"
          />
        </div>
      </div>

      <!-- Secondary Row: Filter Cabang, Filter Desainer, Toggle Hanya Revisi, Dropdown Urutkan -->
      <div class="flex flex-wrap items-center justify-between gap-3 bg-slate-50/80 p-3 rounded-2xl border border-slate-200/80">
        <div class="flex flex-wrap items-center gap-2.5 w-full sm:w-auto">
          <!-- Filter Cabang -->
          <div class="flex items-center gap-1.5 bg-white border border-slate-200 px-3 py-1.5 rounded-xl shadow-2xs">
            <Building2 class="w-4 h-4 text-slate-400 shrink-0" />
            <select
              v-model="selectedBranchFilter"
              @change="handleBranchFilterChange"
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
          <div class="flex items-center gap-1.5 bg-white border border-slate-200 px-3 py-1.5 rounded-xl shadow-2xs">
            <User class="w-4 h-4 text-slate-400 shrink-0" />
            <select
              v-model="selectedEditorFilter"
              @change="handleEditorFilterChange"
              class="bg-transparent text-xs sm:text-sm font-bold text-slate-700 outline-none cursor-pointer pr-2"
            >
              <option value="all">Semua Desainer</option>
              <option v-if="currentEditor || authStore.user?.full_name" value="my_tasks">
                👤 Tugas Saya ({{ currentEditor?.name || authStore.user?.full_name }})
              </option>
              <option value="unassigned">⏳ Belum Diambil (Tanpa Desainer)</option>
              <option disabled>──────────</option>
              <option v-for="editor in editors" :key="editor.id" :value="String(editor.id)">
                {{ editor.name }}
              </option>
            </select>
          </div>

          <!-- Quick Toggle Hanya Revisi -->
          <button
            @click="toggleOnlyRevision"
            type="button"
            :class="[
              'inline-flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-bold transition-all border shadow-2xs cursor-pointer',
              onlyRevision
                ? 'bg-rose-600 text-white border-rose-600 ring-2 ring-rose-600/30'
                : 'bg-white text-slate-600 border-slate-200 hover:border-rose-300 hover:text-rose-600',
            ]"
            title="Saring banner yang memiliki catatan revisi dari validator"
          >
            <AlertTriangle :class="['w-3.5 h-3.5', onlyRevision ? 'text-white' : 'text-rose-500']" />
            <span>Hanya Revisi</span>
            <span
              v-if="onlyRevision"
              class="ml-0.5 px-1.5 py-0.2 rounded-full bg-white/20 text-[10px] font-black"
            >
              Aktif
            </span>
          </button>
        </div>

        <div class="flex items-center gap-2.5 w-full sm:w-auto justify-end">
          <!-- Sort Dropdown -->
          <div class="flex items-center gap-1.5 bg-white border border-slate-200 px-3 py-1.5 rounded-xl shadow-2xs">
            <SlidersHorizontal class="w-4 h-4 text-slate-400 shrink-0" />
            <span class="text-xs text-slate-400 font-semibold hidden md:inline">Urutkan:</span>
            <select
              :value="currentSortValue"
              @change="handleSortChange"
              class="bg-transparent text-xs sm:text-sm font-bold text-slate-700 outline-none cursor-pointer pr-2"
            >
              <option value="revision">⚡ Prioritas Revisi</option>
              <option value="editor_asc">👤 Nama Desainer (A - Z)</option>
              <option value="editor_desc">👤 Nama Desainer (Z - A)</option>
              <option value="shop_asc">🏪 Nama Toko (A - Z)</option>
              <option value="shop_desc">🏪 Nama Toko (Z - A)</option>
              <option value="date_desc">🕒 Tanggal Terbaru</option>
              <option value="date_asc">🕒 Tanggal Terlama</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <div class="bg-white border border-slate-200 rounded-3xl overflow-hidden shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm text-slate-600">
          <thead class="bg-slate-50 text-slate-500 font-bold border-b border-slate-200 select-none">
            <tr>
              <!-- Tanggal -->
              <th
                @click="toggleSort('created_at')"
                class="px-5 py-4 whitespace-nowrap cursor-pointer hover:bg-slate-100/80 transition-colors group"
                title="Klik untuk mengubah urutan tanggal"
              >
                <div class="flex items-center gap-1.5">
                  <span>Tanggal</span>
                  <ArrowUp v-if="sortBy === 'created_at' && sortDir === 'ASC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowDown v-else-if="sortBy === 'created_at' && sortDir === 'DESC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowUpDown v-else class="w-3.5 h-3.5 text-slate-300 opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
              </th>

              <!-- Nama Toko -->
              <th
                @click="toggleSort('shop_name')"
                class="px-5 py-4 min-w-[220px] cursor-pointer hover:bg-slate-100/80 transition-colors group"
                title="Klik untuk mengubah urutan nama toko"
              >
                <div class="flex items-center gap-1.5">
                  <span>Nama Toko & Kontak</span>
                  <ArrowUp v-if="sortBy === 'shop_name' && sortDir === 'ASC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowDown v-else-if="sortBy === 'shop_name' && sortDir === 'DESC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowUpDown v-else class="w-3.5 h-3.5 text-slate-300 opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
              </th>

              <!-- Tipe & Ukuran -->
              <th class="px-5 py-4 whitespace-nowrap">Tipe & Ukuran</th>

              <!-- Qty -->
              <th class="px-4 py-4 text-center whitespace-nowrap">Qty</th>

              <!-- Desainer -->
              <th
                @click="toggleSort('editor_name')"
                class="px-5 py-4 whitespace-nowrap cursor-pointer hover:bg-slate-100/80 transition-colors group"
                title="Klik untuk mengubah urutan desainer"
              >
                <div class="flex items-center gap-1.5">
                  <span>Desainer</span>
                  <ArrowUp v-if="sortBy === 'editor_name' && sortDir === 'ASC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowDown v-else-if="sortBy === 'editor_name' && sortDir === 'DESC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowUpDown v-else class="w-3.5 h-3.5 text-slate-300 opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
              </th>

              <!-- Catatan -->
              <th class="px-5 py-4 min-w-[320px] max-w-[420px]">Catatan / Brief Materi</th>

              <!-- Status / Revisi -->
              <th
                @click="toggleSort('revision')"
                class="px-5 py-4 min-w-[140px] cursor-pointer hover:bg-slate-100/80 transition-colors group"
                title="Klik untuk mengurutkan prioritas revisi"
              >
                <div class="flex items-center gap-1.5">
                  <span>Status / Revisi</span>
                  <ArrowUp v-if="sortBy === 'revision' && sortDir === 'ASC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowDown v-else-if="sortBy === 'revision' && sortDir === 'DESC'" class="w-3.5 h-3.5 text-primary" />
                  <ArrowUpDown v-else class="w-3.5 h-3.5 text-slate-300 opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
              </th>

              <!-- Aksi -->
              <th class="px-5 py-4 text-center min-w-[150px]">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <!-- Skeleton -->
            <template v-if="isLoading">
              <tr
                v-for="i in 5"
                :key="i"
                class="animate-pulse border-b border-slate-100 last:border-none"
              >
                <td class="px-5 py-4"><div class="h-4 bg-slate-100 rounded w-24"></div></td>
                <td class="px-5 py-4">
                  <div class="h-4 bg-slate-100 rounded w-32 mb-2"></div>
                  <div class="h-3 bg-slate-100 rounded w-20"></div>
                </td>
                <td class="px-5 py-4">
                  <div class="h-6 bg-slate-100 rounded w-20 mb-2"></div>
                  <div class="h-6 bg-slate-100 rounded w-24"></div>
                </td>
                <td class="px-4 py-4"><div class="h-6 bg-slate-100 rounded-lg w-12 mx-auto"></div></td>
                <td class="px-5 py-4"><div class="h-6 bg-slate-100 rounded w-28"></div></td>
                <td class="px-5 py-4">
                  <div class="h-14 bg-slate-100 rounded-2xl w-full"></div>
                </td>
                <td class="px-5 py-4"><div class="h-6 bg-slate-100 rounded-full w-24"></div></td>
                <td class="px-5 py-4">
                  <div class="h-8 bg-slate-100 rounded-xl w-full max-w-[120px] mx-auto"></div>
                </td>
              </tr>
            </template>

            <!-- Empty State -->
            <template v-else-if="banners.length === 0">
              <tr>
                <td colspan="8" class="px-6 py-16 text-center">
                  <div class="flex flex-col items-center justify-center gap-3">
                    <Inbox class="w-12 h-12 text-slate-300" />
                    <p class="text-sm font-bold text-slate-500">
                      <template v-if="activeTab === 'request'"
                        >Tidak ada task yang perlu dikerjakan.</template
                      >
                      <template v-else-if="activeTab === 'editing'"
                        >Tidak ada banner yang sedang dikerjakan.</template
                      >
                      <template v-else-if="activeTab === 'validation'"
                        >Tidak ada desain yang menunggu validasi.</template
                      >
                      <template v-else>Belum ada banner yang selesai.</template>
                    </p>
                  </div>
                </td>
              </tr>
            </template>

            <!-- Data Rows -->
            <template v-else>
              <tr
                v-for="banner in banners"
                :key="banner.id"
                class="border-b border-slate-100 hover:bg-slate-50/50 transition-colors last:border-none group"
              >
                <!-- Tanggal -->
                <td class="px-5 py-4 align-top whitespace-nowrap">
                  <div class="flex items-center gap-1.5 text-slate-500 font-medium">
                    <Clock class="w-4 h-4 shrink-0 text-slate-400" />
                    <span class="text-xs">{{ formatDate(banner.created_at) }}</span>
                  </div>
                </td>
                <!-- Nama Toko & Kontak Sales -->
                <td class="px-5 py-4 align-top max-w-[280px]">
                  <div class="flex items-start gap-3">
                    <!-- Foto Toko / Fallback Avatar -->
                    <div
                      v-if="banner.shop_photo"
                      class="relative w-10 h-10 rounded-xl overflow-hidden bg-slate-100 border border-slate-200/80 shrink-0 cursor-pointer group/shop hover:ring-2 hover:ring-primary/40 transition-all shadow-sm mt-0.5"
                      @click="openPhotoPreview(banner.shop_photo, `Foto Toko - ${banner.shop_name || 'Outlet'}`)"
                      title="Klik untuk pratinjau foto toko"
                    >
                      <img
                        :src="photoUrl(banner.shop_photo)!"
                        :alt="banner.shop_name"
                        class="w-full h-full object-cover group-hover/shop:scale-110 transition-transform duration-300"
                      />
                      <div class="absolute inset-0 bg-black/20 opacity-0 group-hover/shop:opacity-100 transition-opacity flex items-center justify-center">
                        <Eye class="w-3.5 h-3.5 text-white drop-shadow" />
                      </div>
                    </div>
                    <div
                      v-else
                      class="w-10 h-10 rounded-xl bg-slate-100 border border-slate-200/60 flex items-center justify-center shrink-0 text-slate-400 mt-0.5"
                      title="Foto toko belum tersedia"
                    >
                      <Store class="w-4 h-4 text-slate-400" />
                    </div>

                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-1.5">
                        <p
                          class="font-bold text-slate-900 group-hover:text-primary transition-colors truncate text-sm"
                        >
                          {{ banner.shop_name || '-' }}
                        </p>
                        <button
                          @click="openDetailModal(banner)"
                          class="p-1 text-slate-400 hover:text-primary hover:bg-slate-100 rounded-lg transition-colors shrink-0"
                          title="Lihat Detail Brief Lengkap"
                        >
                          <Info class="w-3.5 h-3.5" />
                        </button>
                      </div>
                      <!-- Hover Card Alamat Toko -->
                      <StoreAddressHoverCard
                        :address="banner.shop_address"
                        :shop-name="banner.shop_name"
                        :province="banner.province"
                        :regency="banner.regency"
                        :district="banner.district"
                        :village="banner.village"
                        :phone="banner.shop_phone"
                        :latitude="banner.latitude"
                        :longitude="banner.longitude"
                        :sales-name="banner.sales_name"
                        :sales-branch="banner.sales_branch"
                        class="mt-0.5 text-xs"
                      />
                      <div class="text-[11px] font-bold text-slate-400 mt-1.5 flex flex-col gap-0.5">
                        <p>Req by <span class="text-slate-700 font-semibold">{{ banner.sales_name || '-' }}</span> <span v-if="banner.sales_branch" class="text-indigo-500 font-semibold">({{ banner.sales_branch }})</span></p>
                        <a
                          v-if="banner.sales_phone"
                          :href="formatWhatsAppUrl(banner.sales_phone, banner)"
                          target="_blank"
                          class="inline-flex items-center gap-1 text-emerald-600 hover:text-emerald-700 hover:underline mt-0.5 font-bold"
                          title="Klik untuk chat WhatsApp Sales"
                        >
                          <MessageCircle class="w-3 h-3" />
                          {{ banner.sales_phone }}
                        </a>
                      </div>
                    </div>
                  </div>
                </td>
                <!-- Tipe & Ukuran -->
                <td class="px-5 py-4 align-top whitespace-nowrap">
                  <div class="flex flex-col gap-1.5">
                    <span
                      class="inline-flex w-fit items-center gap-1.5 text-xs font-bold px-2.5 py-1 rounded-lg border"
                      :class="(banner.type || '').toLowerCase() === 'universal' ? 'bg-sky-50 text-sky-700 border-sky-200' : 'bg-indigo-50 text-indigo-700 border-indigo-200'"
                    >
                      <FileImage class="w-3.5 h-3.5" />
                      {{ banner.type || 'Custom' }}
                    </span>
                    
                    <!-- Ukuran Request Asli -->
                    <div class="flex flex-col gap-1">
                      <div class="flex items-center gap-1 text-[11px] font-bold text-slate-500">
                        <span class="text-slate-400 font-medium">Req:</span>
                        <span class="bg-slate-100 text-slate-700 px-2 py-0.5 rounded-md border border-slate-200/60 inline-flex items-center gap-1">
                          <Ruler class="w-3 h-3 text-slate-400" />
                          {{ formatSize(banner) }}
                        </span>
                      </div>

                      <!-- Sugesti Ukuran Desain by Sistem -->
                      <div v-if="getSizeSuggestion(banner)" class="flex flex-col gap-0.5">
                        <div class="flex items-center gap-1 text-[11px] font-extrabold text-blue-700">
                          <span class="text-blue-500 font-semibold">Desain:</span>
                          <span class="bg-blue-50 text-blue-800 px-2 py-0.5 rounded-md border border-blue-200/80 inline-flex items-center gap-1 shadow-xs">
                            <Sparkles class="w-3 h-3 text-blue-500 shrink-0" />
                            {{ getSizeSuggestion(banner)?.designLength }} × {{ getSizeSuggestion(banner)?.designWidth }} cm
                          </span>
                        </div>
                        <span class="text-[9px] font-bold text-blue-600/80 pl-0.5 flex items-center gap-0.5">
                          ⓘ {{ getSizeSuggestion(banner)?.ruleText }}
                        </span>
                      </div>
                    </div>
                  </div>
                </td>
                <!-- Qty -->
                <td class="px-4 py-4 align-top text-center whitespace-nowrap">
                  <span
                    class="inline-flex items-center justify-center px-2.5 py-1 bg-slate-100 text-slate-800 text-xs font-black rounded-lg border border-slate-200/80 shadow-2xs"
                    :title="`Jumlah: ${banner.qty || 1} pcs`"
                  >
                    {{ banner.qty || 1 }} Pcs
                  </span>
                </td>
                <!-- Desainer -->
                <td class="px-5 py-4 align-top whitespace-nowrap">
                  <div v-if="banner.editor_name" class="flex items-center gap-1.5">
                    <span
                      class="inline-flex items-center gap-1.5 text-xs font-bold text-blue-700 bg-blue-50 px-2.5 py-1.5 rounded-lg border border-blue-100"
                    >
                      <User class="w-3.5 h-3.5 shrink-0" />
                      {{ banner.editor_name }}
                    </span>
                  </div>
                  <div v-else class="flex items-center gap-1.5">
                    <span
                      class="inline-flex items-center gap-1.5 text-xs font-bold text-amber-700 bg-amber-50 px-2.5 py-1.5 rounded-lg border border-amber-100"
                    >
                      <User class="w-3.5 h-3.5 shrink-0" />
                      Belum diambil
                    </span>
                  </div>
                </td>
                <!-- Notes / Brief Materi -->
                <td class="px-5 py-4 align-top min-w-[320px] max-w-[420px]">
                  <div
                    v-if="banner.notes"
                    class="bg-slate-50/80 hover:bg-slate-50 p-3 rounded-2xl border border-slate-200/80 relative transition-all group/note shadow-xs"
                  >
                    <!-- Header with Copy button -->
                    <div class="flex items-center justify-between gap-2 mb-1.5 pb-1 border-b border-slate-200/60 text-[10px] font-black uppercase tracking-wider text-slate-400">
                      <span class="flex items-center gap-1">
                        <FileText class="w-3 h-3 text-primary" /> Materi Brief
                      </span>
                      <button
                        @click="copyToClipboard(banner.notes, banner.id)"
                        class="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-lg border transition-all cursor-pointer"
                        :class="copiedNoteId === banner.id ? 'bg-emerald-50 text-emerald-700 border-emerald-300' : 'bg-white text-slate-600 hover:text-primary border-slate-200 shadow-xs'"
                        title="Salin seluruh teks materi"
                      >
                        <Check v-if="copiedNoteId === banner.id" class="w-3 h-3 text-emerald-600" />
                        <Copy v-else class="w-3 h-3" />
                        {{ copiedNoteId === banner.id ? 'Tersalin' : 'Salin' }}
                      </button>
                    </div>

                    <!-- Note Text -->
                    <p
                      :class="[
                        'text-xs text-slate-700 font-medium leading-relaxed whitespace-pre-line',
                        !expandedNotes[banner.id] ? 'line-clamp-3' : ''
                      ]"
                    >
                      {{ banner.notes }}
                    </p>

                    <!-- Expand / Collapse Button -->
                    <button
                      v-if="banner.notes.length > 80"
                      @click="toggleExpandNote(banner.id)"
                      class="inline-flex items-center gap-1 text-[11px] font-bold text-primary hover:text-primary/80 mt-2 transition-colors cursor-pointer"
                    >
                      <template v-if="!expandedNotes[banner.id]">
                        <span>Lihat Selengkapnya</span>
                        <ChevronDown class="w-3.5 h-3.5" />
                      </template>
                      <template v-else>
                        <span>Sembunyikan</span>
                        <ChevronUp class="w-3.5 h-3.5" />
                      </template>
                    </button>

                    <!-- Lampiran File / Foto / PDF Pendukung -->
                    <div v-if="banner.documentation_image_url" class="mt-2.5 pt-2 border-t border-slate-200/60 flex items-center justify-between gap-2">
                      <span class="text-[10px] font-black uppercase tracking-wider text-slate-400 flex items-center gap-1">
                        <Paperclip class="w-3 h-3 text-indigo-500" /> File Pendukung
                      </span>
                      <template v-if="banner.documentation_image_url.toLowerCase().endsWith('.pdf')">
                        <a
                          :href="photoUrl(banner.documentation_image_url)!"
                          target="_blank"
                          class="inline-flex items-center gap-1 text-[11px] font-extrabold text-rose-600 bg-rose-50 hover:bg-rose-100 px-2 py-0.5 rounded border border-rose-200 transition-colors"
                          title="Buka / Download File Dokumen PDF"
                        >
                          <FileDown class="w-3 h-3" /> Buka PDF
                        </a>
                      </template>
                      <template v-else>
                        <button
                          @click="openPhotoPreview(banner.documentation_image_url, `File/Logo Pendukung - ${banner.shop_name}`)"
                          class="inline-flex items-center gap-1 text-[11px] font-extrabold text-indigo-600 bg-indigo-50 hover:bg-indigo-100 px-2 py-0.5 rounded border border-indigo-200 transition-colors cursor-pointer"
                          title="Lihat Pratinjau Foto/Logo"
                        >
                          <Eye class="w-3 h-3" /> Lihat Foto/Logo
                        </button>
                      </template>
                    </div>
                  </div>
                  <span v-else class="text-slate-400 text-xs font-medium italic">- Tidak ada catatan -</span>
                </td>
                <!-- Status / Revisi -->
                <td class="px-5 py-4 align-top min-w-[170px] max-w-[230px]">
                  <div
                    v-if="banner.revision_note"
                    class="bg-rose-50/90 hover:bg-rose-50 border border-rose-200/90 rounded-2xl p-2.5 transition-all shadow-2xs group/rev"
                  >
                    <!-- Header Badge & Quick Copy -->
                    <div class="flex items-center justify-between gap-1.5 pb-1 mb-1.5 border-b border-rose-200/60">
                      <span class="inline-flex items-center gap-1 text-[10px] font-black uppercase tracking-wider text-rose-700">
                        <AlertTriangle class="w-3 h-3 text-rose-600 shrink-0" />
                        {{ getRevisionTypeInfo(banner.revision_note).label }}
                      </span>
                      <button
                        @click="copyToClipboard(banner.revision_note, banner.id)"
                        class="p-1 rounded-md text-rose-500 hover:text-rose-700 hover:bg-rose-100/80 transition-colors cursor-pointer"
                        title="Salin catatan revisi"
                      >
                        <Check v-if="copiedNoteId === banner.id" class="w-3 h-3 text-emerald-600" />
                        <Copy v-else class="w-3 h-3" />
                      </button>
                    </div>

                    <!-- Cuplikan Teks (Clamped 2 baris agar row tabel tetap seragam dan tidak memanjang) -->
                    <p class="text-xs text-rose-900 font-medium line-clamp-2 leading-relaxed">
                      {{ banner.revision_note.replace(/^\[.*?\]\s*/, '') }}
                    </p>

                    <!-- Tombol Lihat Detail Catatan Lengkap -->
                    <button
                      @click="openRevisionModal(banner)"
                      class="mt-2 w-full inline-flex items-center justify-center gap-1.5 py-1.5 px-2.5 bg-white hover:bg-rose-100/70 text-rose-700 rounded-xl text-[11px] font-extrabold border border-rose-200 shadow-2xs transition-all cursor-pointer"
                      title="Buka detail instruksi revisi"
                    >
                      <Eye class="w-3.5 h-3.5 text-rose-600 shrink-0" />
                      <span>Detail Revisi ({{ getRevisionLines(banner.revision_note).length }} Poin)</span>
                    </button>
                  </div>
                  <div v-else>
                    <span
                      class="inline-flex items-center justify-center px-3 py-1.5 bg-slate-100 text-slate-600 rounded-lg text-[11px] font-bold uppercase tracking-wider border border-slate-200"
                    >
                      {{ (banner.status ?? '').replace(/_/g, ' ') }}
                    </span>
                  </div>
                </td>
                <!-- Aksi -->
                <td class="px-5 py-4 align-middle text-center min-w-[150px]">
                  <!-- TAB 1: Antrean Request -->
                  <div
                    v-if="activeTab === 'request'"
                    class="flex flex-col sm:flex-row items-center justify-center gap-2"
                  >
                    <!-- If not assigned yet -->
                    <template v-if="!banner.editor_id">
                      <!-- Editor can claim -->
                      <button
                        v-if="currentEditor || isDesignerAdmin"
                        @click="claimTask(banner)"
                        class="inline-flex items-center justify-center gap-2 px-3 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold transition-all shadow-sm w-full sm:w-auto"
                      >
                        <Brush class="w-3.5 h-3.5 shrink-0" />
                        Ambil Tugas
                      </button>
                      <!-- Admin can assign -->
                      <button
                        v-if="isSuperAdmin"
                        @click="openAssignModal(banner)"
                        class="inline-flex items-center justify-center gap-2 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl text-xs font-bold transition-all border border-slate-200 w-full sm:w-auto"
                        title="Tugaskan Desainer"
                      >
                        <User class="w-3.5 h-3.5 shrink-0" />
                        Tugaskan
                      </button>
                      <!-- Read-only monitoring -->
                      <span
                        v-if="!isDesignerAdmin && !isSuperAdmin && !currentEditor"
                        class="text-xs text-amber-700 font-bold bg-amber-50 px-2.5 py-1 rounded-lg border border-amber-200/60"
                      >
                        Menunggu Desainer
                      </span>
                    </template>

                    <!-- If already assigned -->
                    <template v-else>
                      <!-- Assigned Editor or Admin can upload design/revision -->
                      <button
                        v-if="currentEditor?.id === banner.editor_id || authStore.user?.id === banner.editor_user_id || isSuperAdmin || isDesignerAdmin"
                        @click="openUploadModal(banner)"
                        class="inline-flex items-center justify-center gap-2 px-3 py-2 bg-primary hover:bg-primary/90 text-white rounded-xl text-xs font-bold transition-all shadow-sm w-full sm:w-auto"
                      >
                        <Upload class="w-3.5 h-3.5 shrink-0" />
                        {{ banner.revision_note ? 'Revisi' : 'Upload' }}
                      </button>
                      <!-- Admin can re-assign -->
                      <button
                        v-if="isSuperAdmin"
                        @click="openAssignModal(banner)"
                        class="inline-flex items-center justify-center gap-2 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl text-xs font-bold transition-all border border-slate-200 w-full sm:w-auto"
                        title="Tugaskan ulang Desainer"
                      >
                        <User class="w-3.5 h-3.5 shrink-0" />
                        Re-Tugaskan
                      </button>
                      <!-- If another editor is assigned and current user is not admin -->
                      <span
                        v-if="!isSuperAdmin && !isDesignerAdmin && !(currentEditor?.id === banner.editor_id || authStore.user?.id === banner.editor_user_id)"
                        class="text-xs text-slate-400 font-medium"
                      >
                        Dikerjakan oleh {{ banner.editor_name }}
                      </span>
                    </template>
                  </div>

                  <!-- TAB 2: Sedang Dikerjakan -->
                  <div
                    v-else-if="activeTab === 'editing'"
                    class="flex flex-col items-center justify-center gap-2"
                  >
                    <div
                      v-if="currentEditor?.id === banner.editor_id || authStore.user?.id === banner.editor_user_id || isSuperAdmin || isDesignerAdmin"
                      class="flex flex-col items-center gap-1.5 w-full"
                    >
                      <label
                        v-if="!aiDrafts[aiJobKey(banner)] || ['completed', 'failed'].includes(aiDrafts[aiJobKey(banner)]?.status || '')"
                        class="w-full text-left text-[10px] text-slate-600"
                      >
                        Arahan AI (opsional)
                        <textarea
                          v-model="aiInstructions[aiJobKey(banner)]"
                          :disabled="aiGenerating[aiJobKey(banner)]"
                          maxlength="3000"
                          rows="3"
                          placeholder="Contoh: warna merah, nama toko besar, tambahkan promo."
                          class="mt-1 w-full rounded-lg border border-slate-200 p-2 text-xs font-normal focus:outline-none focus:ring-2 focus:ring-violet-400 disabled:opacity-60"
                        />
                        <span>Ukuran, materi sales, dan catatan revisi disertakan otomatis.</span>
                      </label>
                      <button
                        v-if="!aiDrafts[aiJobKey(banner)] || ['completed', 'failed'].includes(aiDrafts[aiJobKey(banner)]?.status || '')"
                        @click="generateAIDraft(banner)"
                        :disabled="aiGenerating[aiJobKey(banner)]"
                        class="inline-flex items-center justify-center gap-2 px-3.5 py-2 bg-violet-600 hover:bg-violet-700 text-white rounded-xl text-xs font-bold transition-all shadow-sm disabled:opacity-60 w-full sm:w-auto"
                      >
                        <RefreshCw v-if="aiGenerating[aiJobKey(banner)]" class="w-3.5 h-3.5 animate-spin" />
                        <Sparkles v-else class="w-3.5 h-3.5" />
                        {{ aiDrafts[aiJobKey(banner)]?.status === 'completed' ? 'Buat Ulang AI' : aiDrafts[aiJobKey(banner)]?.status === 'failed' ? 'Coba Lagi AI' : 'Generate Draft AI' }}
                      </button>

                      <span
                        v-if="aiDrafts[aiJobKey(banner)] && !['completed', 'failed'].includes(aiDrafts[aiJobKey(banner)]?.status || '')"
                        class="inline-flex items-center gap-1.5 px-2.5 py-1.5 bg-violet-50 text-violet-700 border border-violet-200 rounded-lg text-[10px] font-bold"
                      >
                        <RefreshCw class="w-3 h-3 animate-spin" />
                        {{ aiStatusLabel(aiDrafts[aiJobKey(banner)]?.status || 'queued_hermes') }}
                      </span>

                      <a
                        v-if="aiDrafts[aiJobKey(banner)]?.svg_url"
                        :href="aiDrafts[aiJobKey(banner)]?.svg_url || undefined"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="inline-flex items-center justify-center gap-1.5 px-3 py-2 bg-violet-50 hover:bg-violet-100 text-violet-700 border border-violet-200 rounded-xl text-xs font-bold w-full sm:w-auto"
                      >
                        <FileDown class="w-3.5 h-3.5" /> Buka SVG Editable
                      </a>

                      <a
                        v-if="aiDrafts[aiJobKey(banner)]?.status === 'completed' && aiDrafts[aiJobKey(banner)]?.cdr_url"
                        :href="aiDrafts[aiJobKey(banner)]?.cdr_url || undefined"
                        target="_blank"
                        class="inline-flex items-center justify-center gap-1.5 px-3 py-2 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 border border-emerald-200 rounded-xl text-xs font-bold w-full sm:w-auto"
                      >
                        <FileDown class="w-3.5 h-3.5" /> Download CDR AI
                      </a>

                      <div
                        v-if="aiDrafts[aiJobKey(banner)]?.preview_url || aiDrafts[aiJobKey(banner)]?.reference_image_url"
                        class="flex items-center justify-center gap-2 text-[10px] font-bold"
                      >
                        <a
                          v-if="aiDrafts[aiJobKey(banner)]?.preview_url"
                          :href="aiDrafts[aiJobKey(banner)]?.preview_url || undefined"
                          target="_blank"
                          class="text-violet-600 hover:underline"
                        >Preview AI</a>
                        <a
                          v-if="aiDrafts[aiJobKey(banner)]?.reference_image_url"
                          :href="aiDrafts[aiJobKey(banner)]?.reference_image_url || undefined"
                          target="_blank"
                          class="text-slate-500 hover:underline"
                        >Referensi</a>
                      </div>

                      <span
                        v-if="aiDrafts[aiJobKey(banner)]?.status === 'failed'"
                        :title="aiDrafts[aiJobKey(banner)]?.error_message || 'Proses AI gagal'"
                        class="max-w-[190px] text-[10px] leading-snug text-rose-700 bg-rose-50 border border-rose-200 rounded-lg px-2 py-1.5"
                      >
                        AI gagal: {{ aiDrafts[aiJobKey(banner)]?.error_message || 'lihat log worker' }}
                      </span>
                    </div>
                    <!-- Assigned Editor or Admin can upload design/revision -->
                    <button
                      v-if="currentEditor?.id === banner.editor_id || authStore.user?.id === banner.editor_user_id || isSuperAdmin || isDesignerAdmin"
                      @click="openUploadModal(banner)"
                      class="inline-flex items-center justify-center gap-2 px-3.5 py-2 bg-primary hover:bg-primary/90 text-white rounded-xl text-xs font-bold transition-all shadow-sm w-full sm:w-auto"
                    >
                      <Upload class="w-3.5 h-3.5 shrink-0" />
                      {{ banner.revision_note ? 'Upload Revisi' : 'Upload Desain' }}
                    </button>
                    <!-- Admin can re-assign -->
                    <button
                      v-if="isSuperAdmin"
                      @click="openAssignModal(banner)"
                      class="inline-flex items-center justify-center gap-2 px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl text-xs font-bold transition-all border border-slate-200 w-full sm:w-auto"
                      title="Tugaskan ulang Desainer"
                    >
                      <User class="w-3.5 h-3.5 shrink-0" />
                      Re-Tugaskan
                    </button>
                    <!-- If another editor is assigned and current user cannot edit -->
                    <span
                      v-if="!isSuperAdmin && !isDesignerAdmin && !(currentEditor?.id === banner.editor_id || authStore.user?.id === banner.editor_user_id)"
                      class="text-xs text-slate-400 font-medium"
                    >
                      Dikerjakan oleh {{ banner.editor_name || 'Desainer' }}
                    </span>
                  </div>

                  <!-- TAB 3: Menunggu Validasi -->
                  <div
                    v-else-if="activeTab === 'validation'"
                    class="flex flex-col items-center justify-center gap-1.5"
                  >
                    <span
                      class="inline-flex items-center justify-center gap-1.5 px-3 py-1.5 bg-amber-50 text-amber-700 rounded-xl text-xs font-bold border border-amber-200 w-full md:w-auto"
                    >
                      <Hourglass class="w-3.5 h-3.5 shrink-0" />
                      Menunggu
                    </span>
                    <a
                      v-if="banner.design_link || banner.design_url || banner.design_file_url"
                      :href="(banner.design_link || banner.design_url || banner.design_file_url) || undefined"
                      target="_blank"
                      class="text-[11px] font-bold text-violet-600 hover:underline flex items-center gap-1 mt-0.5"
                    >
                      <ExternalLink class="w-3 h-3" /> Lihat Desain
                    </a>
                    <button
                      v-if="currentEditor?.id === banner.editor_id || authStore.user?.id === banner.editor_user_id || isSuperAdmin || isDesignerAdmin"
                      @click="openUploadModal(banner)"
                      class="inline-flex items-center justify-center gap-1.5 px-3 py-1.5 bg-amber-500 hover:bg-amber-600 text-white rounded-xl text-xs font-bold transition-all shadow-xs w-full sm:w-auto cursor-pointer mt-1"
                      title="Ganti / Upload Ulang File Desain"
                    >
                      <Upload class="w-3.5 h-3.5 shrink-0" />
                      Ganti Desain
                    </button>
                  </div>

                  <!-- TAB 4: Selesai -->
                  <div
                    v-else
                    class="flex flex-col items-center justify-center gap-1.5"
                  >
                    <span
                      class="inline-flex items-center justify-center gap-1.5 px-3 py-1.5 bg-emerald-50 text-emerald-700 rounded-xl text-xs font-bold border border-emerald-200 w-full md:w-auto"
                    >
                      <CheckCircle2 class="w-3.5 h-3.5 shrink-0" />
                      Selesai
                    </span>
                    <a
                      v-if="banner.design_link || banner.design_url || banner.design_file_url"
                      :href="(banner.design_link || banner.design_url || banner.design_file_url) || undefined"
                      target="_blank"
                      class="text-[11px] font-bold text-violet-600 hover:underline flex items-center gap-1 mt-0.5"
                    >
                      <ExternalLink class="w-3 h-3" /> Lihat Desain
                    </a>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <!-- Pagination Controls -->
      <div
        v-if="banners.length > 0 || page > 1"
        class="px-6 py-4 border-t border-slate-200 flex flex-col sm:flex-row items-center justify-between gap-4 bg-slate-50"
      >
        <div class="text-sm font-medium text-slate-500">
          Menampilkan
          <span class="font-bold text-slate-700">{{ total > 0 ? (page - 1) * limit + 1 : 0 }}</span>
          - <span class="font-bold text-slate-700">{{ Math.min(page * limit, total) }}</span> dari
          <span class="font-bold text-slate-700">{{ total }}</span> data
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="prevPage"
            :disabled="page === 1"
            class="p-2 bg-white border border-slate-200 rounded-xl text-slate-600 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-sm"
          >
            <ChevronLeft class="w-4 h-4" />
          </button>
          <div
            class="px-4 py-2 bg-white border border-slate-200 rounded-xl text-sm font-bold text-slate-700 shadow-sm min-w-[4rem] text-center"
          >
            {{ page }} / {{ totalPages }}
          </div>
          <button
            @click="nextPage"
            :disabled="page >= totalPages"
            class="p-2 bg-white border border-slate-200 rounded-xl text-slate-600 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-sm"
          >
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- ── Upload Modal ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showUploadModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="closeUploadModal" />

        <!-- Modal Card -->
        <div
          class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl p-8 animate-in fade-in zoom-in duration-200"
        >
          <!-- Modal Header -->
          <div class="flex justify-between items-center mb-6">
            <div>
              <h3 class="text-xl font-black text-slate-900">Upload Hasil Desain</h3>
              <p class="text-sm font-bold text-slate-500 mt-1">
                Toko: <span class="text-primary">{{ selectedBanner?.shop_name }}</span>
              </p>
            </div>
            <button
              @click="closeUploadModal"
              class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Banner meta summary -->
          <div class="grid grid-cols-3 gap-3 mb-6 bg-slate-50 p-4 rounded-2xl text-sm">
            <div>
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">
                Ukuran Request
              </p>
              <p class="font-bold text-slate-700">
                {{ selectedBanner ? formatSize(selectedBanner) : '-' }}
              </p>
            </div>
            <div>
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">
                Tipe
              </p>
              <p class="font-bold text-slate-700">{{ selectedBanner?.type || '-' }}</p>
            </div>
            <div>
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">
                Jumlah (Qty)
              </p>
              <p class="font-bold text-slate-700">{{ selectedBanner?.qty || 1 }} Pcs</p>
            </div>

            <!-- Sugesti Desain -->
            <div v-if="selectedBanner && getSizeSuggestion(selectedBanner)" class="col-span-3 bg-blue-50/80 border border-blue-200/80 rounded-xl p-3">
              <div class="flex items-center justify-between">
                <span class="text-[10px] font-black uppercase text-blue-700 tracking-wider flex items-center gap-1">
                  <Sparkles class="w-3.5 h-3.5 text-blue-600" /> Sugesti Ukuran Desain (Sistem)
                </span>
                <span class="text-[10px] font-bold text-blue-600 bg-white px-2 py-0.5 rounded border border-blue-200">
                  {{ getSizeSuggestion(selectedBanner)?.ruleText }}
                </span>
              </div>
              <p class="text-sm font-black text-blue-900 mt-1">
                {{ getSizeSuggestion(selectedBanner)?.designLength }} × {{ getSizeSuggestion(selectedBanner)?.designWidth }} cm
              </p>
              <p class="text-[11px] text-blue-700/80 mt-0.5 font-medium">
                Gunakan ukuran kanvas ini untuk desain agar proporsi dan margin pemasangan pas.
              </p>
            </div>
          </div>

          <!-- Brief Notes reminder inside modal -->
          <div
            v-if="selectedBanner?.notes"
            class="bg-slate-50 border border-slate-200 p-4 rounded-2xl mb-4 text-xs"
          >
            <div class="flex items-center justify-between gap-2 mb-1">
              <span class="font-black text-slate-500 uppercase tracking-wider flex items-center gap-1">
                <FileText class="w-3.5 h-3.5 text-primary" /> Materi Brief Desain:
              </span>
              <button
                @click="copyToClipboard(selectedBanner.notes)"
                class="text-[10px] font-bold text-slate-600 hover:text-primary flex items-center gap-1 cursor-pointer"
              >
                <Copy class="w-3 h-3" /> Salin
              </button>
            </div>
            <p class="text-slate-700 font-medium whitespace-pre-line leading-relaxed max-h-32 overflow-y-auto">
              {{ selectedBanner.notes }}
            </p>
          </div>

          <!-- Revision note reminder inside modal -->
          <div
            v-if="selectedBanner?.revision_note"
            class="bg-rose-50 border border-rose-200 p-4 rounded-2xl mb-4"
          >
            <div class="flex items-center justify-between gap-1.5 mb-2 pb-1 border-b border-rose-200/60">
              <span class="text-xs font-black text-rose-700 flex items-center gap-1.5">
                <AlertTriangle class="w-4 h-4 text-rose-600 shrink-0" />
                {{ getRevisionTypeInfo(selectedBanner.revision_note).label }} dari Validator:
              </span>
              <button
                @click="copyToClipboard(selectedBanner.revision_note)"
                class="text-[10px] font-bold text-rose-700 hover:text-rose-900 flex items-center gap-1 cursor-pointer"
              >
                <Copy class="w-3 h-3" /> Salin
              </button>
            </div>
            <div class="space-y-1.5 mt-2">
              <div
                v-for="(line, idx) in getRevisionLines(selectedBanner.revision_note)"
                :key="idx"
                class="flex items-start gap-2 text-xs text-rose-900 font-medium leading-relaxed"
              >
                <span class="inline-flex items-center justify-center w-4 h-4 rounded-full bg-rose-200/80 text-rose-800 text-[10px] font-black shrink-0 mt-0.5">
                  {{ idx + 1 }}
                </span>
                <span class="flex-1 whitespace-pre-line">{{ line.replace(/^\[.*?\]\s*/, '').replace(/^[-*•]\s*/, '') }}</span>
              </div>
            </div>
          </div>

          <!-- Drop Zone -->
          <div
            class="border-2 border-dashed rounded-3xl p-7 flex flex-col items-center justify-center gap-3 transition-all cursor-pointer mb-5"
            :class="[
              isUploading
                ? 'pointer-events-none opacity-80 border-indigo-300 bg-indigo-50/40'
                : uploadDragOver
                  ? 'border-primary bg-primary/5'
                  : uploadFile
                    ? 'border-emerald-400 bg-emerald-50/60'
                    : 'border-slate-200 bg-slate-50 hover:border-primary/40',
            ]"
            @dragover.prevent="!isUploading && (uploadDragOver = true)"
            @dragleave.prevent="!isUploading && (uploadDragOver = false)"
            @drop.prevent="!isUploading && onDrop($event)"
            @click="!isUploading && fileInputRef?.click()"
          >
            <input
              ref="fileInputRef"
              type="file"
              accept="image/*,application/pdf"
              class="hidden"
              :disabled="isUploading"
              @change="onFileChange"
            />

            <template v-if="uploadFile">
              <CheckCircle2 class="w-10 h-10 text-emerald-500" />
              <div class="text-center">
                <p class="text-sm font-black text-slate-800">{{ uploadFile.name }}</p>
                <p class="text-xs font-bold text-slate-400 mt-1">{{ fileSizeLabel(uploadFile) }}</p>
              </div>
              <p v-if="!isUploading" class="text-xs text-slate-400 font-medium">Klik untuk ganti file</p>
            </template>
            <template v-else>
              <Upload class="w-10 h-10 text-slate-400" />
              <div class="text-center">
                <p class="text-sm font-bold text-slate-700">
                  Drag & drop atau klik untuk memilih file
                </p>
                <p class="text-[11px] font-bold text-slate-400 mt-1">JPG, PNG, PDF (maks. 150 MB)</p>
              </div>
            </template>
          </div>

          <!-- Progress Bar & Multi-stage status (when uploading) -->
          <div v-if="isUploading" class="mb-5 p-4 rounded-2xl bg-indigo-50/80 border border-indigo-100 space-y-3 animate-in fade-in duration-200">
            <div class="flex items-center justify-between text-xs font-bold">
              <span class="text-indigo-950 flex items-center gap-2">
                <RefreshCw class="w-4 h-4 animate-spin text-primary shrink-0" />
                <template v-if="uploadStage === 'processing_gdrive'">
                  Menyimpan ke Google Drive...
                </template>
                <template v-else>
                  Mengunggah File ke Server...
                </template>
              </span>
              <span class="font-black text-primary text-sm tracking-tight">
                {{ uploadProgress }}%
              </span>
            </div>

            <!-- Progress Bar Track -->
            <div class="w-full bg-slate-200/90 rounded-full h-3 overflow-hidden p-0.5 shadow-inner">
              <div
                class="bg-gradient-to-r from-primary via-indigo-600 to-emerald-500 h-full rounded-full transition-all duration-300 relative overflow-hidden"
                :style="{ width: `${uploadProgress}%` }"
              >
                <div class="absolute inset-0 bg-white/25 animate-pulse"></div>
              </div>
            </div>

            <!-- Stage Details -->
            <div class="flex items-center justify-between text-[11px] font-bold text-slate-500">
              <span v-if="totalBytes > 0">
                {{ (uploadedBytes / (1024 * 1024)).toFixed(1) }} MB / {{ (totalBytes / (1024 * 1024)).toFixed(1) }} MB
              </span>
              <span v-else>Memulai upload...</span>

              <span v-if="uploadStage === 'processing_gdrive'" class="text-indigo-600 font-extrabold animate-pulse flex items-center gap-1">
                <Upload class="w-3 h-3" /> Memproses link Google Drive...
              </span>
              <span v-else class="text-slate-400">
                Mohon jangan tutup jendela ini
              </span>
            </div>
          </div>

          <!-- Submit Button -->
          <button
            @click="submitDesign"
            :disabled="!uploadFile || isUploading"
            class="w-full bg-primary hover:bg-primary/90 disabled:bg-slate-200 disabled:text-slate-400 text-white py-4 rounded-2xl font-black transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2 cursor-pointer"
          >
            <template v-if="isUploading">
              <RefreshCw class="w-5 h-5 animate-spin" />
              <span v-if="uploadStage === 'processing_gdrive'">Menyimpan ke GDrive... ({{ uploadProgress }}%)</span>
              <span v-else>Mengunggah... ({{ uploadProgress }}%)</span>
            </template>
            <template v-else>
              <Upload class="w-5 h-5" />
              Kirim ke Validator
            </template>
          </button>
        </div>
      </div>

      <!-- ── Detail Brief Modal ──────────────────────────────────────────── -->
      <div v-if="showDetailModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="closeDetailModal" />
        <div class="bg-white rounded-[32px] w-full max-w-2xl relative z-10 shadow-2xl p-8 animate-in fade-in zoom-in duration-200 max-h-[90vh] overflow-y-auto">
          <!-- Header -->
          <div class="flex justify-between items-start mb-6">
            <div>
              <div class="flex items-center gap-2">
                <span
                  class="px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-wider border"
                  :class="(selectedBannerForDetail?.type || '').toLowerCase() === 'universal' ? 'bg-sky-50 text-sky-700 border-sky-200' : 'bg-indigo-50 text-indigo-700 border-indigo-200'"
                >
                  {{ selectedBannerForDetail?.type || 'Custom' }}
                </span>
                <span class="text-xs font-bold text-slate-400">Order #{{ selectedBannerForDetail?.id }}</span>
              </div>
              <h3 class="text-2xl font-black text-slate-900 mt-1">
                {{ selectedBannerForDetail?.shop_name }}
              </h3>
            </div>
            <button
              @click="closeDetailModal"
              class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <div class="space-y-5">
            <!-- Grid Specs -->
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-slate-50 p-4 rounded-2xl">
              <div>
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Ukuran Request</p>
                <p class="font-black text-slate-800 text-sm mt-0.5">{{ selectedBannerForDetail ? formatSize(selectedBannerForDetail) : '-' }}</p>
              </div>
              <div>
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Jumlah (Qty)</p>
                <p class="font-black text-slate-800 text-sm mt-0.5">{{ selectedBannerForDetail?.qty || 1 }} Pcs</p>
              </div>
              <div>
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Tanggal Order</p>
                <p class="font-bold text-slate-700 text-xs mt-0.5">{{ formatDate(selectedBannerForDetail?.created_at) }}</p>
              </div>
              <div>
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Status</p>
                <p class="font-bold text-slate-700 text-xs mt-0.5 uppercase">{{ selectedBannerForDetail?.status }}</p>
              </div>

              <!-- Sugesti Desain Detail -->
              <div v-if="selectedBannerForDetail && getSizeSuggestion(selectedBannerForDetail)" class="col-span-2 sm:col-span-4 bg-blue-50/80 border border-blue-200/80 rounded-xl p-3.5 mt-1">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-black uppercase text-blue-700 tracking-wider flex items-center gap-1.5">
                    <Sparkles class="w-4 h-4 text-blue-600" /> Sugesti Ukuran Desain (Sistem):
                  </span>
                  <span class="text-[10px] font-bold text-blue-700 bg-white px-2.5 py-0.5 rounded-full border border-blue-200 shadow-xs">
                    {{ getSizeSuggestion(selectedBannerForDetail)?.ruleText }}
                  </span>
                </div>
                <div class="text-base font-black text-blue-900 mt-1 flex items-baseline gap-2">
                  <span>{{ getSizeSuggestion(selectedBannerForDetail)?.formattedDesignDual || `${getSizeSuggestion(selectedBannerForDetail)?.designLength} × ${getSizeSuggestion(selectedBannerForDetail)?.designWidth} cm` }}</span>
                  <span class="text-xs font-semibold text-blue-600">(Panjang -{{ getSizeSuggestion(selectedBannerForDetail)?.reduction }}cm, lebar proporsional)</span>
                </div>
              </div>
            </div>

            <!-- Alamat & Lokasi Toko -->
            <div class="bg-slate-50 border border-slate-200/80 p-4 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div class="space-y-1 min-w-0 flex-1">
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest flex items-center gap-1.5">
                  <MapPin class="w-3.5 h-3.5 text-rose-500" /> Alamat & Lokasi Toko
                </p>
                <p class="font-bold text-slate-800 text-sm leading-relaxed">
                  {{ getBannerFullAddress(selectedBannerForDetail) }}
                </p>
                <div v-if="selectedBannerForDetail?.village || selectedBannerForDetail?.district || selectedBannerForDetail?.regency || selectedBannerForDetail?.province" class="flex flex-wrap gap-1 pt-1 text-[10px]">
                  <span v-if="selectedBannerForDetail.village" class="bg-white px-2 py-0.5 rounded-md border border-slate-200 text-slate-600 font-semibold">Ds. {{ selectedBannerForDetail.village }}</span>
                  <span v-if="selectedBannerForDetail.district" class="bg-white px-2 py-0.5 rounded-md border border-slate-200 text-slate-600 font-semibold">Kec. {{ selectedBannerForDetail.district }}</span>
                  <span v-if="selectedBannerForDetail.regency" class="bg-white px-2 py-0.5 rounded-md border border-slate-200 text-slate-600 font-semibold">{{ selectedBannerForDetail.regency }}</span>
                  <span v-if="selectedBannerForDetail.province" class="bg-white px-2 py-0.5 rounded-md border border-slate-200 text-slate-600 font-semibold">Prov. {{ selectedBannerForDetail.province }}</span>
                </div>
              </div>
              <div class="flex items-center gap-2 shrink-0 self-start sm:self-center">
                <button
                  v-if="getBannerFullAddress(selectedBannerForDetail) !== 'Alamat tidak tersedia'"
                  type="button"
                  @click="copyToClipboard(getBannerFullAddress(selectedBannerForDetail))"
                  class="inline-flex items-center gap-1.5 text-xs font-bold px-3 py-2 bg-white hover:bg-slate-50 text-slate-700 rounded-xl border border-slate-200 shadow-xs transition cursor-pointer"
                >
                  <Copy class="w-3.5 h-3.5" /> Salin Alamat
                </button>
              </div>
            </div>

            <!-- Kontak Sales -->
            <div class="bg-indigo-50/60 border border-indigo-100 p-4 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div>
                <p class="text-[10px] font-black text-indigo-500 uppercase tracking-widest">Sales Pemohon</p>
                <p class="font-bold text-slate-800 text-sm mt-0.5">
                  {{ selectedBannerForDetail?.sales_name }} <span v-if="selectedBannerForDetail?.sales_branch" class="text-indigo-600 font-semibold">({{ selectedBannerForDetail?.sales_branch }})</span>
                </p>
              </div>
              <a
                v-if="selectedBannerForDetail?.sales_phone"
                :href="formatWhatsAppUrl(selectedBannerForDetail.sales_phone, selectedBannerForDetail)"
                target="_blank"
                class="inline-flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition-all shadow-sm self-start sm:self-auto"
              >
                <MessageCircle class="w-4 h-4" />
                Chat WhatsApp Sales
              </a>
            </div>

            <!-- Notes Section with Copy -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <h4 class="text-xs font-black text-slate-500 uppercase tracking-wider flex items-center gap-1.5">
                  <FileText class="w-4 h-4 text-primary" /> Catatan & Materi Spanduk
                </h4>
                <button
                  v-if="selectedBannerForDetail?.notes"
                  @click="copyToClipboard(selectedBannerForDetail.notes)"
                  class="inline-flex items-center gap-1.5 text-xs font-bold px-3 py-1 bg-white hover:bg-slate-50 text-slate-700 rounded-xl border border-slate-200 shadow-xs transition-all cursor-pointer"
                >
                  <Copy class="w-3.5 h-3.5" /> Salin Materi
                </button>
              </div>
              <div class="bg-slate-50 border border-slate-200 rounded-2xl p-4 text-sm font-medium text-slate-800 whitespace-pre-line leading-relaxed">
                {{ selectedBannerForDetail?.notes || 'Tidak ada catatan khusus.' }}
              </div>
            </div>

            <!-- Lampiran File / Logo / Dokumen Pendukung -->
            <div v-if="selectedBannerForDetail?.documentation_image_url" class="bg-indigo-50/70 border border-indigo-200/80 rounded-2xl p-4 space-y-2">
              <div class="flex items-center justify-between">
                <h4 class="text-xs font-black text-indigo-900 uppercase tracking-wider flex items-center gap-1.5">
                  <Paperclip class="w-4 h-4 text-indigo-600" /> Lampiran File / Logo / Draft Sales
                </h4>
                <template v-if="selectedBannerForDetail.documentation_image_url.toLowerCase().endsWith('.pdf')">
                  <a
                    :href="photoUrl(selectedBannerForDetail.documentation_image_url)!"
                    target="_blank"
                    class="inline-flex items-center gap-1.5 px-3 py-1 bg-rose-600 hover:bg-rose-700 text-white rounded-xl text-xs font-bold transition shadow-xs"
                  >
                    <FileDown class="w-3.5 h-3.5" /> Buka Dokumen PDF
                  </a>
                </template>
                <template v-else>
                  <a
                    :href="photoUrl(selectedBannerForDetail.documentation_image_url)!"
                    target="_blank"
                    class="inline-flex items-center gap-1.5 px-3 py-1 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold transition shadow-xs"
                  >
                    <ExternalLink class="w-3.5 h-3.5" /> Buka Tab Baru
                  </a>
                </template>
              </div>

              <!-- Thumbnail if image -->
              <div
                v-if="!selectedBannerForDetail.documentation_image_url.toLowerCase().endsWith('.pdf')"
                class="relative h-44 rounded-xl overflow-hidden bg-slate-100 border border-indigo-200 cursor-pointer group mt-2"
                @click="openPhotoPreview(selectedBannerForDetail.documentation_image_url, `Lampiran Logo/File - ${selectedBannerForDetail.shop_name}`)"
              >
                <img
                  :src="photoUrl(selectedBannerForDetail.documentation_image_url)!"
                  class="w-full h-full object-contain bg-white group-hover:scale-105 transition-transform duration-500"
                />
                <div class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white font-bold text-xs gap-1.5">
                  <Eye class="w-4 h-4" /> Klik untuk perbesar
                </div>
              </div>
            </div>

            <!-- Revision Note if any -->
            <div v-if="selectedBannerForDetail?.revision_note" class="bg-rose-50 border border-rose-200 rounded-2xl p-4">
              <div class="flex items-center justify-between gap-1.5 mb-2 pb-1.5 border-b border-rose-200/60">
                <span class="text-xs font-black text-rose-700 flex items-center gap-1.5">
                  <AlertTriangle class="w-4 h-4 text-rose-600 shrink-0" />
                  {{ getRevisionTypeInfo(selectedBannerForDetail.revision_note).label }} dari Validator:
                </span>
                <button
                  @click="copyToClipboard(selectedBannerForDetail.revision_note)"
                  class="text-[10px] font-bold text-rose-700 hover:text-rose-900 bg-white hover:bg-rose-100 px-2 py-0.5 rounded border border-rose-200 transition-colors cursor-pointer"
                >
                  <Copy class="w-3 h-3 inline mr-1" /> Salin
                </button>
              </div>
              <div class="space-y-1.5 mt-2">
                <div
                  v-for="(line, idx) in getRevisionLines(selectedBannerForDetail.revision_note)"
                  :key="idx"
                  class="flex items-start gap-2 text-xs text-rose-900 font-medium leading-relaxed"
                >
                  <span class="inline-flex items-center justify-center w-4 h-4 rounded-full bg-rose-200 text-rose-800 text-[10px] font-black shrink-0 mt-0.5">
                    {{ idx + 1 }}
                  </span>
                  <span class="flex-1 whitespace-pre-line">{{ line.replace(/^\[.*?\]\s*/, '').replace(/^[-*•]\s*/, '') }}</span>
                </div>
              </div>
            </div>

            <!-- Desain Saat Ini & Opsi Re-Upload Master Kualitas Penuh -->
            <div v-if="selectedBannerForDetail?.design_url || selectedBannerForDetail?.design_link" class="bg-slate-50 border border-slate-200 rounded-2xl p-4 space-y-3">
              <div class="flex items-center justify-between">
                <h4 class="text-xs font-black text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
                  <Palette class="w-4 h-4 text-purple-600" /> File Desain Banner
                </h4>
                <a
                  :href="(selectedBannerForDetail.design_link || selectedBannerForDetail.design_url)!"
                  target="_blank"
                  class="inline-flex items-center gap-1 text-xs font-bold text-primary hover:underline"
                >
                  <ExternalLink class="w-3.5 h-3.5" /> Buka Desain
                </a>
              </div>

              <div v-if="selectedBannerForDetail.design_file_name" class="text-xs font-mono bg-white p-2.5 rounded-xl border border-slate-200/80 text-slate-700">
                📎 {{ selectedBannerForDetail.design_file_name }}
              </div>

              <div class="pt-1 flex items-center justify-between gap-3">
                <p class="text-[11px] text-slate-500 font-medium">
                  Perlu mengganti atau mengunggah ulang file desain master kualitas penuh (100% Raw)?
                </p>
                <button
                  type="button"
                  @click="closeDetailModal(); openUploadModal(selectedBannerForDetail!)"
                  class="px-3.5 py-2 bg-primary hover:bg-primary/90 text-white rounded-xl text-xs font-bold transition-all shadow-xs whitespace-nowrap shrink-0 flex items-center gap-1.5"
                >
                  <Upload class="w-3.5 h-3.5" /> Upload Ulang Master
                </button>
              </div>
            </div>

            <!-- Foto Toko Preview -->
            <div v-if="selectedBannerForDetail?.shop_photo">
              <h4 class="text-xs font-black text-slate-500 uppercase tracking-wider mb-2 flex items-center gap-1.5">
                <Store class="w-4 h-4 text-primary" /> Foto Toko / Lokasi
              </h4>
              <div
                class="relative h-48 rounded-2xl overflow-hidden bg-slate-100 border border-slate-200 cursor-pointer group"
                @click="openPhotoPreview(selectedBannerForDetail.shop_photo, `Foto Toko - ${selectedBannerForDetail.shop_name}`)"
              >
                <img
                  :src="photoUrl(selectedBannerForDetail.shop_photo)!"
                  class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                />
                <div class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white font-bold text-xs gap-1.5">
                  <Eye class="w-4 h-4" /> Klik untuk perbesar
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Assign Modal -->
      <div v-if="showAssignModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="closeAssignModal" />

        <!-- Modal Card -->
        <div
          class="bg-white rounded-[32px] w-full max-w-md relative z-10 shadow-2xl p-8 animate-in fade-in zoom-in duration-200"
        >
          <!-- Modal Header -->
          <div class="flex justify-between items-center mb-6">
            <div>
              <h3 class="text-xl font-black text-slate-900">Tugaskan Desainer</h3>
              <p class="text-sm font-bold text-slate-500 mt-1">
                Toko: <span class="text-primary">{{ selectedBannerForAssign?.shop_name }}</span>
              </p>
            </div>
            <button
              @click="closeAssignModal"
              class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Select Editor Dropdown -->
          <div class="space-y-4 mb-6">
            <label class="block text-xs font-black text-slate-400 uppercase tracking-widest"
              >Pilih Editor / Desainer</label
            >
            <div class="relative">
              <select
                v-model="selectedEditorId"
                class="w-full px-4 py-3.5 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-medium focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all appearance-none cursor-pointer text-slate-700"
              >
                <option :value="null" disabled>-- Pilih Desainer --</option>
                <option v-for="editor in editors" :key="editor.id" :value="editor.id">
                  {{ editor.name }}
                </option>
              </select>
              <!-- Custom dropdown arrow -->
              <div
                class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-slate-500"
              >
                <svg
                  class="fill-current h-4 w-4"
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 20 20"
                >
                  <path
                    d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z"
                  />
                </svg>
              </div>
            </div>
          </div>

          <!-- Submit Button -->
          <button
            @click="submitAssign"
            :disabled="!selectedEditorId || isAssigning"
            class="w-full bg-primary hover:bg-primary/90 disabled:bg-slate-200 disabled:text-slate-400 text-white py-4 rounded-2xl font-black transition-all shadow-lg shadow-primary/20 flex items-center justify-center gap-2"
          >
            <RefreshCw v-if="isAssigning" class="w-5 h-5 animate-spin" />
            <User v-else class="w-5 h-5" />
            {{ isAssigning ? 'Memproses...' : 'Tugaskan Sekarang' }}
          </button>
        </div>
      </div>
    </Teleport>

    <!-- Revision Detail Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showRevisionModal && selectedBannerForRevision"
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
        >
          <!-- Backdrop -->
          <div
            class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
            @click="closeRevisionModal"
          />

          <!-- Modal Card -->
          <div
            class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl p-6 sm:p-7 animate-in fade-in zoom-in duration-200 max-h-[90vh] flex flex-col"
          >
            <!-- Modal Header -->
            <div class="flex justify-between items-start mb-4 pb-3 border-b border-slate-100">
              <div>
                <div class="flex items-center gap-2 mb-1">
                  <span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-black bg-rose-100 text-rose-700 border border-rose-200">
                    <AlertTriangle class="w-3.5 h-3.5 text-rose-600" />
                    {{ getRevisionTypeInfo(selectedBannerForRevision.revision_note).label }}
                  </span>
                  <span class="text-xs font-bold text-slate-400">
                    {{ formatDate(selectedBannerForRevision.created_at) }}
                  </span>
                </div>
                <h3 class="text-lg sm:text-xl font-black text-slate-900 leading-snug">
                  {{ selectedBannerForRevision.shop_name }}
                </h3>
                <p class="text-xs font-semibold text-slate-500 mt-0.5">
                  Sales: <span class="text-slate-800">{{ selectedBannerForRevision.sales_name }}</span>
                  <span v-if="selectedBannerForRevision.sales_branch" class="text-indigo-600"> ({{ selectedBannerForRevision.sales_branch }})</span>
                </p>
              </div>
              <button
                @click="closeRevisionModal"
                class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-600 transition-colors cursor-pointer"
              >
                <X class="w-5 h-5" />
              </button>
            </div>

            <!-- Modal Scrollable Content -->
            <div class="space-y-4 overflow-y-auto pr-1 flex-1">
              <!-- Alert Container for Revision Points -->
              <div class="bg-rose-50/80 border border-rose-200 rounded-2xl p-4">
                <div class="flex items-center justify-between mb-2 pb-1.5 border-b border-rose-200/60">
                  <span class="text-xs font-black uppercase tracking-wider text-rose-700 flex items-center gap-1.5">
                    <AlertTriangle class="w-4 h-4 text-rose-600" /> Instruksi Revisi dari Validator
                  </span>
                  <button
                    @click="copyToClipboard(selectedBannerForRevision.revision_note)"
                    class="inline-flex items-center gap-1 text-[11px] font-bold text-rose-700 hover:text-rose-900 bg-white hover:bg-rose-100/80 px-2.5 py-1 rounded-lg border border-rose-200 shadow-2xs transition-all cursor-pointer"
                    title="Salin seluruh instruksi revisi"
                  >
                    <Copy class="w-3 h-3" /> Salin Teks
                  </button>
                </div>

                <!-- List of Points -->
                <div class="space-y-2 mt-2.5">
                  <div
                    v-for="(line, idx) in getRevisionLines(selectedBannerForRevision.revision_note)"
                    :key="idx"
                    class="flex items-start gap-2.5 bg-white p-3 rounded-xl border border-rose-100 text-xs text-rose-950 font-medium leading-relaxed shadow-2xs"
                  >
                    <span class="flex items-center justify-center w-5 h-5 rounded-full bg-rose-100 text-rose-700 text-[10px] font-black shrink-0 mt-0.5">
                      {{ idx + 1 }}
                    </span>
                    <div class="flex-1 whitespace-pre-line">
                      {{ line.replace(/^\[.*?\]\s*/, '').replace(/^[-*•]\s*/, '') }}
                    </div>
                  </div>
                </div>
              </div>

              <!-- Brief Materi Asli (Sebagai Acuan Desainer) -->
              <div v-if="selectedBannerForRevision.notes" class="bg-slate-50 border border-slate-200 rounded-2xl p-4">
                <div class="flex items-center justify-between mb-1.5 pb-1 border-b border-slate-200/60 text-xs font-bold text-slate-500">
                  <span class="flex items-center gap-1.5">
                    <FileText class="w-3.5 h-3.5 text-primary" /> Materi Brief Awal:
                  </span>
                  <button
                    @click="copyToClipboard(selectedBannerForRevision.notes)"
                    class="text-[10px] font-bold text-slate-600 hover:text-primary flex items-center gap-1 cursor-pointer"
                  >
                    <Copy class="w-3 h-3" /> Salin
                  </button>
                </div>
                <p class="text-xs text-slate-700 font-medium whitespace-pre-line leading-relaxed max-h-32 overflow-y-auto">
                  {{ selectedBannerForRevision.notes }}
                </p>
              </div>
            </div>

            <!-- Modal Footer Actions -->
            <div class="mt-5 pt-3 border-t border-slate-100 flex flex-col sm:flex-row items-center justify-end gap-2.5">
              <button
                type="button"
                @click="closeRevisionModal"
                class="w-full sm:w-auto px-5 py-2.5 rounded-xl border border-slate-200 text-slate-600 hover:bg-slate-50 text-xs font-bold transition-all cursor-pointer"
              >
                Tutup
              </button>
              <button
                v-if="currentEditor?.id === selectedBannerForRevision.editor_id || authStore.user?.id === selectedBannerForRevision.editor_user_id || isSuperAdmin || isDesignerAdmin"
                type="button"
                @click="const b = selectedBannerForRevision; closeRevisionModal(); openUploadModal(b)"
                class="w-full sm:w-auto px-5 py-2.5 bg-primary hover:bg-primary/90 text-white rounded-xl text-xs font-bold transition-all shadow-md shadow-primary/20 flex items-center justify-center gap-1.5 cursor-pointer"
              >
                <Upload class="w-3.5 h-3.5" />
                Upload Hasil Revisi
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Photo Preview Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="previewPhoto"
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
          @click.self="previewPhoto = null"
        >
          <div
            class="absolute inset-0 bg-black/80 backdrop-blur-sm"
            @click="previewPhoto = null"
          ></div>
          <div class="relative z-10 max-w-3xl w-full bg-slate-900 rounded-[2rem] overflow-hidden shadow-2xl flex flex-col border border-white/10">
            <!-- Modal Header -->
            <div class="px-6 py-4 bg-slate-900/90 border-b border-white/10 flex items-center justify-between text-white">
              <div class="flex items-center gap-2.5">
                <Store class="w-4 h-4 text-primary" />
                <h4 class="font-extrabold text-sm text-slate-100 truncate">{{ previewTitle }}</h4>
              </div>
              <div class="flex items-center gap-2">
                <a
                  :href="previewPhoto"
                  target="_blank"
                  class="p-2 hover:bg-white/10 rounded-xl text-slate-300 hover:text-white transition-colors text-xs font-bold flex items-center gap-1.5"
                  title="Buka foto asli"
                >
                  <ExternalLink class="w-4 h-4" />
                </a>
                <button
                  @click="previewPhoto = null"
                  class="p-2 hover:bg-white/10 rounded-xl text-slate-300 hover:text-white transition-colors"
                >
                  <X class="w-5 h-5" />
                </button>
              </div>
            </div>

            <!-- Modal Content -->
            <div class="p-4 flex items-center justify-center bg-black/40 min-h-[300px]">
              <iframe
                v-if="previewPhoto.includes('drive.google.com')"
                :src="previewPhoto"
                class="w-full h-[65vh] rounded-xl border-0 bg-white"
                allow="autoplay"
                allowfullscreen
              ></iframe>
              <img
                v-else
                :src="previewPhoto"
                class="w-full rounded-xl shadow-lg object-contain max-h-[75vh]"
              />
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
