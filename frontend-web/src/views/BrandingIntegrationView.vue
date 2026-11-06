<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { resolvePhotoUrl } from '@/utils/imageUrl'
import {
  Store,
  Users,
  MapPin,
  Image,
  TrendingUp,
  CheckCircle2,
  Clock,
  Search,
  RefreshCw,
  ExternalLink,
  Phone,
  Wifi,
  WifiOff,
  Camera,
  Download,
  CheckCircle,
  PenTool,
  Calendar,
  ArrowUpDown,
  Sparkles,
  Wallet,
} from 'lucide-vue-next'

// ── State ──────────────────────────────────────────────────────────────────
const isLoading = ref(false)
const isSyncing = ref(false)
const activeTab = ref<'shops' | 'banners' | 'agents' | 'operators' | 'commissions'>('shops')
const searchQuery = ref('')
const sourceStatus = ref<'live' | 'cache' | 'stale_cache' | ''>('')
const isOnline = ref(true)

const masterStore = useMasterDataStore()
const route = useRoute()
const isCommissionPage = computed(() => route.path === '/sales/komisi-operator')

// ── Modals & Forms State ───────────────────────────────────────────────────
const showDesignModal = ref(false)
const showValidateModal = ref(false)
const showVendorModal = ref(false)
const showProofModal = ref(false)
const selectedBanner = ref<any>(null)
const proofData = ref<any>(null)

const designFile = ref<File | null>(null)
const validateNote = ref('')
const vendorName = ref('')
const vendorPrice = ref(0)
const isSubmitting = ref(false)

// ── Data ───────────────────────────────────────────────────────────────────
const shops = ref<any[]>([])
const banners = ref<any[]>([])
const agents = ref<any[]>([])
const operators = ref<any[]>([])
const commissions = ref<any[]>([])
const commissionMonth = ref(new Date().getMonth() + 1)
const commissionYear = ref(new Date().getFullYear())
const commissionDate = ref(new Date().toISOString().slice(0, 10))
const stats = ref<any>(null)

// Operator Target Modal
const showTargetModal = ref(false)
const selectedOperator = ref<any>(null)
const operatorTargetValue = ref(0)

const statusFilter = ref('all')
const editorFilter = ref('all')
const designDateFilter = ref<'all' | 'today' | 'yesterday' | 'custom'>('all')
const customDesignDate = ref('')
const sortBy = ref<'created_at' | 'editing_finished_at' | 'installed_at'>('editing_finished_at')
const sortDir = ref<'asc' | 'desc'>('desc')

// ── Fetch ──────────────────────────────────────────────────────────────────
const fetchStats = async () => {
  try {
    const branchId = masterStore.selectedBranchId
    const query = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/branding/stats${query}`)
    if (res.data?.data) stats.value = res.data.data
  } catch {}
}

const fetchShops = async () => {
  isLoading.value = true
  try {
    const branchId = masterStore.selectedBranchId
    const query = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/branding/shops${query}`)
    shops.value = res.data?.data || []
    sourceStatus.value = res.data?.source || 'live'
    isOnline.value = res.data?.source !== 'stale_cache'
  } catch (e: any) {
    isOnline.value = false
    toast.error(
      'Gagal memuat data toko dari Branding App: ' + (e.response?.data?.error || e.message),
    )
  } finally {
    isLoading.value = false
  }
}

const fetchBanners = async () => {
  isLoading.value = true
  try {
    const branchId = masterStore.selectedBranchId
    let filter =
      statusFilter.value !== 'all' ? `?status=${statusFilter.value}&limit=99999` : '?limit=99999'
    if (branchId && branchId !== 'ALL_BRANCHES') {
      filter += `&branch_id=${branchId}`
    }
    if (editorFilter.value && editorFilter.value !== 'all') {
      filter += `&editor_id=${encodeURIComponent(editorFilter.value)}`
    }
    if (designDateFilter.value === 'today') {
      const today = new Date().toISOString().split('T')[0]
      filter += `&design_date=${today}`
    } else if (designDateFilter.value === 'yesterday') {
      const y = new Date()
      y.setDate(y.getDate() - 1)
      filter += `&design_date=${y.toISOString().split('T')[0]}`
    } else if (designDateFilter.value === 'custom' && customDesignDate.value) {
      filter += `&design_date=${customDesignDate.value}`
    }
    if (sortBy.value) {
      filter += `&sort_by=${sortBy.value}&sort_dir=${sortDir.value}`
    }
    const res = await apiClient.get(`/branding/banners${filter}`)
    banners.value = res.data?.data || []
    sourceStatus.value = res.data?.source || 'live'
    isOnline.value = res.data?.source !== 'stale_cache'
  } catch (e: any) {
    isOnline.value = false
    toast.error('Gagal memuat data banner: ' + (e.response?.data?.error || e.message))
  } finally {
    isLoading.value = false
  }
}

const fetchAgents = async () => {
  isLoading.value = true
  try {
    const branchId = masterStore.selectedBranchId
    const query = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/branding/agents${query}`)
    agents.value = res.data?.data || []
    sourceStatus.value = res.data?.source || 'live'
  } catch (e: any) {
    toast.error('Gagal memuat data sales agent: ' + (e.response?.data?.error || e.message))
  } finally {
    isLoading.value = false
  }
}

const fetchOperators = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get('/branding/operators')
    operators.value = res.data?.data || []
  } catch (e: any) {
    toast.error('Gagal memuat data operator: ' + (e.response?.data?.error || e.message))
  } finally {
    isLoading.value = false
  }
}

const fetchCommissions = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get('/branding/operator/performance', {
      params: { date: commissionDate.value },
    })
    commissions.value = res.data?.data || []
  } catch (e: any) {
    toast.error('Gagal memuat komisi operator: ' + (e.response?.data?.error || e.message))
  } finally {
    isLoading.value = false
  }
}

const updateCommissionPayment = async (operator: any) => {
  const isUnpaid = Number(operator.komisi_belum_dibayar || 0) > 0
  const reference = isUnpaid ? window.prompt('Nomor referensi pembayaran:') : null
  if (isUnpaid && reference === null) return
  try {
    await apiClient.post('/branding/operator/commission/payment', {
      operator_id: operator.id,
      month: commissionMonth.value,
      year: commissionYear.value,
      action: isUnpaid ? 'mark_paid' : 'mark_unpaid',
      payment_reference: reference || '',
    })
    toast.success(isUnpaid ? 'Komisi operator ditandai sudah dibayar' : 'Pembayaran komisi dibatalkan')
    fetchCommissions()
  } catch (e: any) {
    toast.error('Gagal memperbarui pembayaran: ' + (e.response?.data?.error || e.message))
  }
}

const commissionPeriodLabel = computed(() =>
  new Date(commissionYear.value, commissionMonth.value - 1, 1).toLocaleDateString('id-ID', {
    month: 'long',
    year: 'numeric',
  }),
)

const saveOperatorTarget = async () => {
  if (!selectedOperator.value) return
  isSubmitting.value = true
  try {
    await apiClient.put(`/branding/operators/${selectedOperator.value.id}/target`, {
      target_daily_installations: operatorTargetValue.value
    })
    toast.success('Target harian operator berhasil disimpan')
    showTargetModal.value = false
    fetchOperators()
  } catch (e: any) {
    toast.error('Gagal menyimpan target: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const syncFromBrandingApp = async () => {
  isSyncing.value = true
  try {
    const res = await apiClient.post('/branding/sync')
    const updated = res.data?.updated_count || 0
    const created = res.data?.created_count || 0
    const completed = res.data?.completed_count || 0
    toast.success(
      `Sinkronisasi berhasil! ${updated} diperbarui, ${created} baru, ${completed} selesai/terpasang.`
    )
    refreshAll()
  } catch (e: any) {
    toast.error(
      'Gagal menyinkronkan dari Branding App: ' + (e.response?.data?.error || e.message)
    )
  } finally {
    isSyncing.value = false
  }
}

const handleViewProof = (b: any) => {
  proofData.value = b
  showProofModal.value = true
}

const refreshAll = () => {
  fetchStats()
  if (activeTab.value === 'shops') fetchShops()
  else if (activeTab.value === 'banners') fetchBanners()
  else if (activeTab.value === 'agents') fetchAgents()
  else if (activeTab.value === 'operators') fetchOperators()
  else fetchCommissions()
}

const switchTab = (tab: 'shops' | 'banners' | 'agents' | 'operators' | 'commissions') => {
  activeTab.value = tab
  searchQuery.value = ''
  if (tab === 'shops' && shops.value.length === 0) fetchShops()
  else if (tab === 'banners' && banners.value.length === 0) fetchBanners()
  else if (tab === 'agents' && agents.value.length === 0) fetchAgents()
  else if (tab === 'operators' && operators.value.length === 0) fetchOperators()
  else if (tab === 'commissions' && commissions.value.length === 0) fetchCommissions()
}

watch(
  () => masterStore.selectedBranchId,
  () => {
    refreshAll()
  },
)

onMounted(() => {
  fetchStats()
  if (route.path === '/sales/komisi-operator') {
    activeTab.value = 'commissions'
    fetchCommissions()
  } else {
    fetchShops()
  }
})

// ── Filtered Lists ─────────────────────────────────────────────────────────
const filteredShops = computed(() => {
  const q = searchQuery.value.toLowerCase()
  return shops.value.filter(
    (s) =>
      !q ||
      s.name?.toLowerCase().includes(q) ||
      s.owner_name?.toLowerCase().includes(q) ||
      s.address_detail?.toLowerCase().includes(q) ||
      s.sales_name?.toLowerCase().includes(q),
  )
})

const filteredBanners = computed(() => {
  const q = searchQuery.value.toLowerCase()
  return banners.value.filter(
    (b) =>
      !q ||
      b.shop_name?.toLowerCase().includes(q) ||
      b.sales_name?.toLowerCase().includes(q) ||
      b.status?.toLowerCase().includes(q),
  )
})

const filteredAgents = computed(() => {
  const q = searchQuery.value.toLowerCase()
  return agents.value.filter(
    (a) =>
      !q ||
      a.name?.toLowerCase().includes(q) ||
      a.email?.toLowerCase().includes(q) ||
      a.area?.toLowerCase().includes(q),
  )
})

const filteredOperators = computed(() => {
  if (!searchQuery.value) return operators.value
  const q = searchQuery.value.toLowerCase()
  return operators.value.filter(
    (o) =>
      o.name.toLowerCase().includes(q) ||
      o.email.toLowerCase().includes(q) ||
      (o.area && o.area.toLowerCase().includes(q)),
  )
})

const filteredCommissions = computed(() => {
  const q = searchQuery.value.toLowerCase()
  return commissions.value.filter((operator) =>
    !q || `${operator.name} ${operator.area} ${operator.email}`.toLowerCase().includes(q),
  )
})

const totalCommission = computed(() =>
  commissions.value.reduce((total, operator) => total + Number(operator.total_komisi || 0), 0),
)
const paidCommission = computed(() =>
  commissions.value.reduce((total, operator) => total + Number(operator.komisi_dibayar || 0), 0),
)
const unpaidCommission = computed(() => totalCommission.value - paidCommission.value)

const editorsList = computed(() => {
  const set = new Set<string>()
  banners.value.forEach((b) => {
    if (b.editor_name && b.editor_name.trim()) {
      set.add(b.editor_name.trim())
    }
  })
  return Array.from(set)
})

const dailyDesignStats = computed(() => {
  const statsMap: Record<string, number> = {}
  let totalFinished = 0

  banners.value.forEach((b) => {
    if (b.editing_finished_at) {
      totalFinished++
      const editor = b.editor_name || 'Tanpa Desainer'
      statsMap[editor] = (statsMap[editor] || 0) + 1
    }
  })

  return {
    total: totalFinished,
    byEditor: statsMap,
  }
})

const formatDateTime = (dateStr: string | null | undefined): string => {
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

const toggleSort = (col: 'created_at' | 'editing_finished_at' | 'installed_at') => {
  if (sortBy.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = col
    sortDir.value = 'desc'
  }
  fetchBanners()
}

// ── File Handlers ────────────────────────────────────────────────────────────────
const handleFilterBanners = (s: string) => {
  statusFilter.value = s
  fetchBanners()
}

const handleUploadDesignClick = (b: any) => {
  selectedBanner.value = b
  showDesignModal.value = true
}

const handleValidateClick = (b: any) => {
  selectedBanner.value = b
  showValidateModal.value = true
}

const handleProcessVendorClick = (b: any) => {
  selectedBanner.value = b
  showVendorModal.value = true
}

const handleCancelDesign = () => {
  showDesignModal.value = false
  designFile.value = null
}

const formatCurrency = (v: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(v || 0)

const shopStatusColor = (status: string) => {
  if (status === 'hot') return 'bg-red-500/10 text-red-700 border-red-200/50'
  if (status === 'warm') return 'bg-amber-500/10 text-amber-700 border-amber-200/50'
  return 'bg-blue-500/10 text-blue-700 border-blue-200/50'
}

const bannerStatusColor = (status: string) => {
  if (status === 'installed') return 'bg-emerald-500/10 text-emerald-700 border-emerald-200/50'
  if (status === 'request') return 'bg-amber-500/10 text-amber-700 border-amber-200/50'
  if (status === 'ready_for_operator') return 'bg-blue-500/10 text-blue-700 border-blue-200/50'
  if (status === 'validation') return 'bg-purple-500/10 text-purple-700 border-purple-200/50'
  if (status === 'purchasing_queue') return 'bg-indigo-500/10 text-indigo-700 border-indigo-200/50'
  return 'bg-slate-100 text-slate-500 border-slate-200'
}

const initials = (name: string) =>
  (name || 'U')
    .split(' ')
    .slice(0, 2)
    .map((w) => w[0])
    .join('')
    .toUpperCase()

// ── Form Submit Handlers ───────────────────────────────────────────────────
const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    designFile.value = target.files[0] as File
  }
}

const submitDesign = async () => {
  if (!designFile.value || !selectedBanner.value) return
  isSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('design_file', designFile.value)

    await apiClient.put(`/branding/banners/${selectedBanner.value.id}/design`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    toast.success('Desain berhasil diunggah')
    showDesignModal.value = false
    fetchBanners()
  } catch (e: any) {
    toast.error('Gagal mengunggah desain: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
    designFile.value = null
  }
}

const submitValidation = async (isApproved: boolean) => {
  if (!selectedBanner.value) return
  if (!isApproved && !validateNote.value) {
    toast.error('Catatan revisi wajib diisi jika menolak desain')
    return
  }
  isSubmitting.value = true
  try {
    await apiClient.put(`/branding/banners/${selectedBanner.value.id}/validate`, {
      status: isApproved ? 'approved' : 'rejected',
      is_approved: isApproved,
      type: 'both',
      revision_note: isApproved ? '' : validateNote.value,
    })
    toast.success(isApproved ? 'Desain disetujui' : 'Catatan revisi berhasil dikirim ke Editor')
    showValidateModal.value = false
    fetchBanners()
  } catch (e: any) {
    toast.error('Gagal memvalidasi desain: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
    validateNote.value = ''
  }
}

const submitVendor = async () => {
  if (!selectedBanner.value || !vendorName.value) return
  isSubmitting.value = true
  try {
    await apiClient.put(`/branding/banners/${selectedBanner.value.id}/vendor-process`, {
      vendor_name: vendorName.value,
      total_price: Number(vendorPrice.value),
    })
    toast.success('Spanduk berhasil diproses ke vendor')
    showVendorModal.value = false
    fetchBanners()
  } catch (e: any) {
    toast.error('Gagal memproses vendor: ' + (e.response?.data?.error || e.message))
  } finally {
    isSubmitting.value = false
    vendorName.value = ''
    vendorPrice.value = 0
  }
}
</script>

<template>
  <div class="space-y-8 pb-12 animate-in fade-in duration-700">
    <!-- Header -->
    <div class="flex flex-col xl:flex-row xl:items-center justify-between gap-6">
      <div>
        <div class="flex items-center gap-3 mb-2">
          <h1 class="text-4xl font-black text-slate-900 tracking-tight">Komisi Operator Spanduk</h1>
          <!-- Connection status pill -->
          <span
            :class="[
              'flex items-center gap-1.5 px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest border',
              isOnline
                ? 'bg-emerald-500/10 text-emerald-700 border-emerald-200/50'
                : 'bg-amber-500/10 text-amber-700 border-amber-200/50',
            ]"
          >
            <Wifi v-if="isOnline" class="w-3 h-3" />
            <WifiOff v-else class="w-3 h-3" />
            {{ isOnline ? 'Live' : 'Cache' }}
          </span>
        </div>
        <p class="text-[15px] font-medium text-slate-500 max-w-2xl leading-relaxed">
          Komisi operator dan reimbursement sales untuk pemasangan banner.
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-3 shrink-0">
        <button
          @click="syncFromBrandingApp"
          :disabled="isSyncing || isLoading"
          class="flex items-center gap-2 px-5 py-3 bg-gradient-to-r from-emerald-600 to-teal-600 text-white rounded-2xl font-bold text-sm shadow-lg shadow-emerald-600/20 hover:from-emerald-500 hover:to-teal-500 active:scale-95 transition-all disabled:opacity-50"
        >
          <RefreshCw class="w-4 h-4" :class="isSyncing ? 'animate-spin' : ''" />
          {{ isSyncing ? 'Menyinkronkan...' : 'Tarik Data Branding' }}
        </button>

        <button
          @click="refreshAll"
          :disabled="isLoading || isSyncing"
          class="flex items-center gap-2 px-5 py-3 bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 rounded-2xl font-bold text-sm shadow-sm active:scale-95 transition-all disabled:opacity-50"
        >
          <RefreshCw class="w-4 h-4" :class="isLoading ? 'animate-spin' : ''" />
          Refresh
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div v-if="!isCommissionPage && stats" class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="bg-white p-6 rounded-[2rem] border border-slate-200 shadow-sm">
        <Store class="w-7 h-7 text-primary mb-3 opacity-80" />
        <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Total Toko</p>
        <h3 class="text-3xl font-black text-slate-900 mt-1">{{ shops.length }}</h3>
      </div>
      <div class="bg-white p-6 rounded-[2rem] border border-slate-200 shadow-sm">
        <Image class="w-7 h-7 text-emerald-500 mb-3 opacity-80" />
        <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
          Banner Terpasang
        </p>
        <h3 class="text-3xl font-black text-emerald-600 mt-1">{{ stats.installed_banners }}</h3>
      </div>
      <div class="bg-white p-6 rounded-[2rem] border border-slate-200 shadow-sm">
        <Clock class="w-7 h-7 text-amber-500 mb-3 opacity-80" />
        <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
          Menunggu Pasang
        </p>
        <h3 class="text-3xl font-black text-amber-600 mt-1">{{ stats.pending_banners }}</h3>
      </div>
      <div
        class="bg-gradient-to-br from-primary to-primary/80 p-6 rounded-[2rem] shadow-lg shadow-primary/20 relative overflow-hidden"
      >
        <TrendingUp class="absolute -right-3 -bottom-3 w-20 h-20 opacity-10" />
        <p class="text-[10px] font-black text-white/80 uppercase tracking-widest">Revenue Banner</p>
        <h3 class="text-2xl font-black text-white mt-1 leading-tight">
          {{ formatCurrency(stats.total_revenue_banner) }}
        </h3>
      </div>
    </div>
    <div v-else-if="isLoading && !isCommissionPage" class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="i in 4" :key="i" class="bg-slate-100 h-28 rounded-[2rem] animate-pulse"></div>
    </div>

    <!-- Tab Switcher + Search -->
    <div class="flex flex-col md:flex-row items-start md:items-center gap-4">
      <div
        class="bg-white/60 backdrop-blur-md border border-slate-200 p-1.5 rounded-2xl flex shadow-sm"
      >
        <button v-if="!isCommissionPage"
          @click="switchTab('shops')"
          :class="[
            'px-5 py-2.5 rounded-xl text-sm font-bold transition-all duration-300 flex items-center gap-2',
            activeTab === 'shops'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          <Store class="w-4 h-4" /> Toko ({{ shops.length }})
        </button>
        <button v-if="!isCommissionPage"
          @click="switchTab('banners')"
          :class="[
            'px-5 py-2.5 rounded-xl text-sm font-bold transition-all duration-300 flex items-center gap-2',
            activeTab === 'banners'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          <Image class="w-4 h-4" /> Banner ({{ banners.length }})
        </button>
        <button v-if="!isCommissionPage"
          @click="switchTab('agents')"
          :class="[
            'px-5 py-2.5 rounded-xl text-sm font-bold transition-all duration-300 flex items-center gap-2',
            activeTab === 'agents'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          <Users class="w-4 h-4" /> Sales Agent ({{ agents.length }})
        </button>
        <button v-if="!isCommissionPage"
          @click="switchTab('operators')"
          :class="[
            'px-5 py-2.5 rounded-xl text-sm font-bold transition-all duration-300 flex items-center gap-2',
            activeTab === 'operators'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          <Users class="w-4 h-4" /> Operator ({{ operators.length }})
        </button>
        <button v-if="isCommissionPage"
          @click="switchTab('commissions')"
          :class="[
            'px-5 py-2.5 rounded-xl text-sm font-bold transition-all duration-300 flex items-center gap-2',
            activeTab === 'commissions'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800',
          ]"
        >
          <Wallet class="w-4 h-4" /> Komisi ({{ commissions.length }})
        </button>
      </div>

      <div class="relative flex-1 max-w-md ml-auto">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="
            activeTab === 'shops'
              ? 'Cari nama toko, pemilik, atau sales...'
              : activeTab === 'banners'
                ? 'Cari toko atau status...'
                : 'Cari nama, email, atau area...'
          "
          class="w-full bg-white border border-slate-200 rounded-2xl py-3 pl-11 pr-4 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all placeholder:text-slate-400"
        />
      </div>
    </div>

    <!-- Loading Spinner -->
    <div v-if="isLoading" class="py-20 flex items-center justify-center">
      <div class="text-center space-y-4">
        <span
          class="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin inline-block"
        ></span>
        <p class="text-sm font-bold text-slate-400">Mengambil data dari Branding App...</p>
      </div>
    </div>

    <template v-else>
      <!-- ── TAB: SHOPS ──────────────────────────────────────────────── -->
      <div
        v-if="activeTab === 'shops'"
        class="space-y-4 animate-in slide-in-from-left-4 duration-500"
      >
        <div
          v-if="filteredShops.length === 0"
          class="bg-white rounded-[2.5rem] border border-slate-200 p-20 text-center shadow-sm"
        >
          <Store class="w-16 h-16 text-slate-200 mx-auto mb-4" />
          <h3 class="text-xl font-black text-slate-900">Tidak Ada Data Toko</h3>
          <p class="text-slate-400 text-sm mt-2">
            Pastikan Branding API URL sudah dikonfigurasi dengan benar.
          </p>
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
          <div
            v-for="shop in filteredShops"
            :key="shop.id"
            class="bg-white p-6 rounded-[2rem] border border-slate-200 shadow-sm hover:shadow-xl hover:-translate-y-0.5 hover:border-primary/20 transition-all duration-300 group relative overflow-hidden"
          >
            <div
              class="absolute top-0 right-0 w-32 h-32 bg-primary/5 blur-3xl rounded-full -mr-10 -mt-10 group-hover:bg-primary/10 transition-colors"
            ></div>

            <div class="flex items-start justify-between gap-4 relative z-10">
              <div class="flex items-center gap-3">
                <div
                  class="w-11 h-11 rounded-2xl bg-primary/10 flex items-center justify-center font-black text-primary text-sm shrink-0"
                >
                  {{ initials(shop.name) }}
                </div>
                <div>
                  <h3
                    class="font-black text-slate-900 text-sm leading-tight group-hover:text-primary transition-colors"
                  >
                    {{ shop.name }}
                  </h3>
                  <p class="text-[10px] font-bold text-slate-400 mt-0.5">{{ shop.owner_name }}</p>
                </div>
              </div>
              <span
                :class="[
                  'px-2.5 py-1 rounded-full text-[9px] font-black uppercase tracking-widest border shrink-0',
                  shopStatusColor(shop.status),
                ]"
              >
                {{ shop.status }}
              </span>
            </div>

            <div class="mt-4 space-y-2 relative z-10">
              <div v-if="shop.address_detail" class="flex items-start gap-2 text-xs text-slate-500">
                <MapPin class="w-3.5 h-3.5 mt-0.5 shrink-0 text-slate-400" />
                <span class="line-clamp-2 font-medium">{{ shop.address_detail }}</span>
              </div>
              <div v-if="shop.phone" class="flex items-center gap-2 text-xs text-slate-500">
                <Phone class="w-3.5 h-3.5 shrink-0 text-slate-400" />
                <span class="font-medium">{{ shop.phone }}</span>
              </div>
              <div v-if="shop.sales_name" class="flex items-center gap-2">
                <div
                  class="w-5 h-5 rounded-lg bg-emerald-500/10 flex items-center justify-center shrink-0"
                >
                  <Users class="w-3 h-3 text-emerald-600" />
                </div>
                <span class="text-[11px] font-bold text-emerald-700">{{ shop.sales_name }}</span>
                <span
                  v-if="shop.sales_area"
                  class="text-[9px] font-black text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full ml-auto"
                >
                  {{ shop.sales_area }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── TAB: BANNERS ───────────────────────────────────────────── -->
      <div
        v-if="activeTab === 'banners'"
        class="space-y-4 animate-in slide-in-from-right-4 duration-500"
      >
        <!-- Widget Rekap Output Tim Desain Harian -->
        <div
          class="bg-gradient-to-br from-indigo-950 via-slate-900 to-slate-950 text-white rounded-[2rem] p-6 shadow-xl relative overflow-hidden border border-indigo-900/40"
        >
          <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 relative z-10">
            <div>
              <div
                class="flex items-center gap-2 text-[11px] font-black text-indigo-300 uppercase tracking-widest"
              >
                <PenTool class="w-4 h-4 text-indigo-400" />
                Rekap Output Tim Desain ({{
                  designDateFilter === 'all'
                    ? 'Semua Data'
                    : designDateFilter === 'today'
                      ? 'Hari Ini'
                      : designDateFilter === 'yesterday'
                        ? 'Kemarin'
                        : customDesignDate || 'Tanggal Khusus'
                }})
              </div>
              <div class="flex items-baseline gap-3 mt-1.5">
                <h3 class="text-3xl font-black text-white">{{ dailyDesignStats.total }}</h3>
                <span class="text-xs text-slate-300 font-semibold">spanduk telah selesai didesain</span>
              </div>
            </div>

            <!-- Breakdown per Editor / Designer -->
            <div class="flex flex-wrap items-center gap-2">
              <div
                v-for="(count, edName) in dailyDesignStats.byEditor"
                :key="edName"
                class="bg-white/10 backdrop-blur-md border border-white/15 px-3 py-1.5 rounded-xl flex items-center gap-2 text-xs shadow-sm"
              >
                <div class="w-2 h-2 rounded-full bg-emerald-400"></div>
                <span class="font-bold text-slate-200">{{ edName }}:</span>
                <span class="font-black text-white bg-indigo-500/40 px-1.5 py-0.5 rounded-md text-[11px]">{{
                  count
                }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Filter & Sort Toolbar -->
        <div
          class="bg-white p-4 rounded-[2rem] border border-slate-200 shadow-sm flex flex-wrap items-center justify-between gap-4"
        >
          <!-- Status Filter Pills -->
          <div class="flex flex-wrap items-center gap-2">
            <button
              v-for="s in ['all', 'installed', 'request', 'ready_for_operator']"
              :key="s"
              @click="handleFilterBanners(s)"
              :class="[
                'px-3.5 py-1.5 rounded-xl text-xs font-black uppercase tracking-wide border transition-all',
                statusFilter === s
                  ? 'bg-primary text-white border-primary shadow-sm shadow-primary/20'
                  : 'bg-white text-slate-500 border-slate-200 hover:border-primary/30',
              ]"
            >
              {{
                s === 'all'
                  ? 'Semua'
                  : s === 'installed'
                    ? '✅ Terpasang'
                    : s === 'request'
                      ? '🕐 Request'
                      : '🔧 Siap Pasang'
              }}
            </button>
          </div>

          <!-- Secondary Filters: Designer & Design Date & Sorting -->
          <div class="flex flex-wrap items-center gap-3">
            <!-- Filter Desainer -->
            <div
              class="flex items-center gap-1.5 bg-slate-50 border border-slate-200 rounded-xl px-3 py-1.5 text-xs font-bold text-slate-700"
            >
              <PenTool class="w-3.5 h-3.5 text-slate-400" />
              <select
                v-model="editorFilter"
                @change="fetchBanners"
                class="bg-transparent focus:outline-none cursor-pointer"
              >
                <option value="all">Semua Desainer</option>
                <option v-for="ed in editorsList" :key="ed" :value="ed">{{ ed }}</option>
              </select>
            </div>

            <!-- Filter Tanggal Selesai Desain -->
            <div
              class="flex items-center gap-1.5 bg-slate-50 border border-slate-200 rounded-xl px-3 py-1.5 text-xs font-bold text-slate-700"
            >
              <Calendar class="w-3.5 h-3.5 text-slate-400" />
              <select
                v-model="designDateFilter"
                @change="fetchBanners"
                class="bg-transparent focus:outline-none cursor-pointer"
              >
                <option value="all">Semua Tgl Desain</option>
                <option value="today">Desain Hari Ini</option>
                <option value="yesterday">Desain Kemarin</option>
                <option value="custom">Pilih Tanggal...</option>
              </select>
              <input
                v-if="designDateFilter === 'custom'"
                v-model="customDesignDate"
                @change="fetchBanners"
                type="date"
                class="bg-white border border-slate-200 rounded-lg px-2 py-0.5 text-xs font-bold text-slate-700 focus:outline-none ml-1"
              />
            </div>

            <!-- Sorting Button -->
            <button
              @click="toggleSort('editing_finished_at')"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold transition-all border"
              :class="
                sortBy === 'editing_finished_at'
                  ? 'bg-primary/10 text-primary border-primary/30'
                  : 'bg-slate-50 hover:bg-slate-100 text-slate-700 border-slate-200'
              "
              title="Urutkan berdasarkan tanggal desain selesai"
            >
              <ArrowUpDown class="w-3.5 h-3.5" />
              Sort Tgl Selesai
              <span v-if="sortBy === 'editing_finished_at'">{{ sortDir === 'desc' ? '⬇️' : '⬆️' }}</span>
            </button>
          </div>
        </div>

        <div
          v-if="filteredBanners.length === 0"
          class="bg-white rounded-[2.5rem] border border-slate-200 p-20 text-center shadow-sm"
        >
          <Image class="w-16 h-16 text-slate-200 mx-auto mb-4" />
          <h3 class="text-xl font-black text-slate-900">Tidak Ada Data Banner</h3>
        </div>

        <div
          v-else
          class="bg-white rounded-[2.5rem] border border-slate-200 shadow-sm overflow-hidden"
        >
          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr
                  class="bg-slate-50 text-[10px] font-black uppercase tracking-[0.12em] text-slate-400 border-b border-slate-100"
                >
                  <th class="px-6 py-4">Toko</th>
                  <th class="px-6 py-4">Ukuran & Qty</th>
                  <th class="px-6 py-4">Desainer</th>
                  <th
                    @click="toggleSort('editing_finished_at')"
                    class="px-6 py-4 cursor-pointer hover:text-primary transition-colors select-none"
                  >
                    <div class="flex items-center gap-1">
                      Tgl Selesai Desain
                      <ArrowUpDown class="w-3 h-3 opacity-60" />
                    </div>
                  </th>
                  <th class="px-6 py-4">Total Harga</th>
                  <th class="px-6 py-4">Status</th>
                  <th class="px-6 py-4">Desain</th>
                  <th class="px-6 py-4">Bukti Pasang</th>
                  <th class="px-6 py-4">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="b in filteredBanners"
                  :key="b.id"
                  class="hover:bg-slate-50/50 transition-colors"
                >
                  <td class="px-6 py-4">
                    <span class="font-bold text-slate-800 text-sm">{{
                      b.shop_name || `Shop #${b.shop_id}`
                    }}</span>
                    <p class="text-[10px] text-slate-500 mt-0.5">Oleh: {{ b.sales_name || '-' }}</p>
                  </td>
                  <td class="px-6 py-4">
                    <div class="text-xs font-bold text-slate-700">
                      {{ b.size_length }}cm × {{ b.size_width }}cm
                    </div>
                    <span class="text-[10px] font-black text-slate-400">{{ b.qty }} pcs</span>
                  </td>
                  <td class="px-6 py-4">
                    <div v-if="b.editor_name" class="flex items-center gap-1.5">
                      <div class="w-6 h-6 rounded-full bg-primary/10 text-primary flex items-center justify-center text-[10px] font-black">
                        {{ initials(b.editor_name) }}
                      </div>
                      <span class="text-xs font-bold text-slate-700">{{ b.editor_name }}</span>
                    </div>
                    <span v-else class="text-xs text-slate-400 italic">Belum di-assign</span>
                  </td>
                  <td class="px-6 py-4">
                    <span
                      v-if="b.editing_finished_at"
                      class="text-xs font-semibold text-slate-700 bg-slate-100 px-2.5 py-1 rounded-lg border border-slate-200/60 inline-block"
                    >
                      {{ formatDateTime(b.editing_finished_at) }}
                    </span>
                    <span v-else class="text-xs text-slate-400 italic">Belum selesai</span>
                  </td>
                  <td class="px-6 py-4 text-sm font-black text-slate-800">
                    {{ formatCurrency(b.total_price) }}
                  </td>
                  <td class="px-6 py-4">
                    <span
                      :class="[
                        'px-2.5 py-1 rounded-full text-[9px] font-black uppercase tracking-widest border',
                        bannerStatusColor(b.status),
                      ]"
                    >
                      {{ b.status }}
                    </span>
                  </td>
                  <td class="px-6 py-4">
                    <div v-if="b.design_link" class="flex flex-col gap-1">
                      <a
                        :href="b.design_link"
                        target="_blank"
                        class="flex items-center gap-1 text-[11px] font-bold text-primary hover:underline"
                      >
                        <ExternalLink class="w-3 h-3" /> Lihat
                      </a>
                      <span
                        v-if="b.design_file_name"
                        class="text-[10px] font-mono text-slate-500 truncate max-w-[160px]"
                        :title="b.design_file_name"
                      >
                        📎 {{ b.design_file_name }}
                      </span>
                    </div>
                    <span v-else class="text-[11px] text-slate-400">—</span>
                  </td>
                  <td class="px-6 py-4">
                    <button
                      v-if="b.photo_installed"
                      @click="handleViewProof(b)"
                      class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 text-emerald-700 hover:bg-emerald-100 border border-emerald-200/60 rounded-xl text-xs font-bold transition-all shadow-sm"
                    >
                      <Camera class="w-3.5 h-3.5" />
                      Lihat Foto
                    </button>
                    <span
                      v-else-if="b.status === 'installed' || b.status === 'terpasang'"
                      class="text-[11px] font-bold text-amber-600 italic"
                    >
                      Belum ada foto
                    </span>
                    <span v-else class="text-[11px] text-slate-400">—</span>
                  </td>
                  <td class="px-6 py-4">
                    <button
                      v-if="b.status === 'request'"
                      @click="handleUploadDesignClick(b)"
                      class="px-3 py-1.5 bg-primary/10 text-primary hover:bg-primary/20 rounded-lg text-xs font-bold transition-colors"
                    >
                      Upload Desain
                    </button>
                    <button
                      v-else-if="b.status === 'validation'"
                      @click="handleValidateClick(b)"
                      class="px-3 py-1.5 bg-amber-500/10 text-amber-700 hover:bg-amber-500/20 rounded-lg text-xs font-bold transition-colors"
                    >
                      Validasi
                    </button>
                    <button
                      v-else-if="b.status === 'purchasing_queue'"
                      @click="handleProcessVendorClick(b)"
                      class="px-3 py-1.5 bg-indigo-500/10 text-indigo-700 hover:bg-indigo-500/20 rounded-lg text-xs font-bold transition-colors"
                    >
                      Proses Vendor
                    </button>
                    <span v-else class="text-[11px] text-slate-400">—</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- ── TAB: AGENTS ────────────────────────────────────────────── -->
      <div
        v-if="activeTab === 'agents'"
        class="space-y-4 animate-in slide-in-from-right-4 duration-500"
      >
        <div
          v-if="filteredAgents.length === 0"
          class="bg-white rounded-[2.5rem] border border-slate-200 p-20 text-center shadow-sm"
        >
          <Users class="w-16 h-16 text-slate-200 mx-auto mb-4" />
          <h3 class="text-xl font-black text-slate-900">Tidak Ada Sales Agent</h3>
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
          <div
            v-for="agent in filteredAgents"
            :key="agent.id"
            class="bg-white p-6 rounded-[2rem] border border-slate-200 shadow-sm hover:shadow-lg hover:-translate-y-0.5 hover:border-primary/20 transition-all duration-300"
          >
            <div class="flex items-center gap-4 mb-4">
              <div
                class="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center font-black text-primary text-sm"
              >
                {{ initials(agent.name) }}
              </div>
              <div class="flex-1 min-w-0">
                <h3 class="font-black text-slate-900 truncate">{{ agent.name }}</h3>
                <p class="text-[11px] text-slate-400 font-bold">{{ agent.email }}</p>
              </div>
              <span
                :class="[
                  'px-2.5 py-1 rounded-full text-[9px] font-black uppercase tracking-widest border shrink-0',
                  agent.is_active
                    ? 'bg-emerald-500/10 text-emerald-700 border-emerald-200/50'
                    : 'bg-red-500/10 text-red-700 border-red-200/50',
                ]"
              >
                {{ agent.is_active ? 'Aktif' : 'Non-aktif' }}
              </span>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div class="bg-slate-50 p-3 rounded-2xl">
                <p class="text-[9px] font-black text-slate-400 uppercase tracking-wider">Area</p>
                <p class="text-xs font-black text-slate-700 mt-0.5">{{ agent.area || '-' }}</p>
              </div>
              <div class="bg-slate-50 p-3 rounded-2xl">
                <p class="text-[9px] font-black text-slate-400 uppercase tracking-wider">Telepon</p>
                <p class="text-xs font-black text-slate-700 mt-0.5">{{ agent.phone || '-' }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── TAB: OPERATORS ────────────────────────────────────────────── -->
      <div
        v-if="activeTab === 'operators'"
        class="space-y-4 animate-in slide-in-from-right-4 duration-500"
      >
        <div
          v-if="filteredOperators.length === 0"
          class="bg-white rounded-[2.5rem] border border-slate-200 p-20 text-center shadow-sm"
        >
          <Users class="w-16 h-16 text-slate-200 mx-auto mb-4" />
          <h3 class="text-xl font-black text-slate-900">Tidak Ada Operator Spanduk</h3>
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
          <div
            v-for="operator in filteredOperators"
            :key="operator.id"
            class="bg-white p-6 rounded-[2rem] border border-slate-200 shadow-sm hover:shadow-lg hover:-translate-y-0.5 hover:border-primary/20 transition-all duration-300"
          >
            <div class="flex items-center gap-4 mb-4">
              <div
                class="w-12 h-12 rounded-2xl bg-indigo-500/10 flex items-center justify-center font-black text-indigo-500 text-sm"
              >
                {{ initials(operator.name) }}
              </div>
              <div class="flex-1 min-w-0">
                <h3 class="font-black text-slate-900 truncate">{{ operator.name }}</h3>
                <p class="text-[11px] text-slate-400 font-bold">{{ operator.email }}</p>
              </div>
              <span
                :class="[
                  'px-2.5 py-1 rounded-full text-[9px] font-black uppercase tracking-widest border shrink-0',
                  operator.is_active
                    ? 'bg-emerald-500/10 text-emerald-700 border-emerald-200/50'
                    : 'bg-red-500/10 text-red-700 border-red-200/50',
                ]"
              >
                {{ operator.is_active ? 'Aktif' : 'Non-aktif' }}
              </span>
            </div>
            <div class="grid grid-cols-2 gap-3 mb-4">
              <div class="bg-slate-50 p-3 rounded-2xl">
                <p class="text-[9px] font-black text-slate-400 uppercase tracking-wider">Target Harian</p>
                <p class="text-xs font-black text-indigo-600 mt-0.5">{{ operator.target_daily_installations || 0 }} Toko</p>
              </div>
              <div class="bg-slate-50 p-3 rounded-2xl">
                <p class="text-[9px] font-black text-slate-400 uppercase tracking-wider">Telepon</p>
                <p class="text-xs font-black text-slate-700 mt-0.5">{{ operator.phone || '-' }}</p>
              </div>
            </div>
            <button
              @click="selectedOperator = operator; operatorTargetValue = operator.target_daily_installations; showTargetModal = true"
              class="w-full bg-slate-100 hover:bg-slate-200 text-slate-700 py-2.5 rounded-xl text-xs font-bold transition-colors"
            >
              Set Target Minimal Pemasangan
            </button>
          </div>
        </div>
      </div>

      <!-- ── TAB: OPERATOR COMMISSIONS ──────────────────────────────── -->
      <div
        v-if="activeTab === 'commissions'"
        class="space-y-4 animate-in slide-in-from-right-4 duration-500"
      >
        <div class="bg-white rounded-[2rem] border border-slate-200 p-4 flex flex-wrap items-end gap-3">
          <label class="text-xs font-bold text-slate-500">
            Tanggal pemasangan
            <input v-model="commissionDate" @change="fetchCommissions" type="date" class="block mt-1 rounded-xl border-slate-200 text-sm font-bold" />
          </label>
          <a href="/reimbursements" class="ml-auto text-xs font-black text-primary hover:underline pb-2">Lihat Reimbursement Sales →</a>
          <!-- Filter bulan/tahun dipertahankan untuk kompatibilitas laporan lama. -->
          <span class="hidden">{{ commissionMonth }}{{ commissionYear }}</span>
          <label class="hidden text-xs font-bold text-slate-500">
            Bulan
            <select v-model.number="commissionMonth" @change="fetchCommissions" class="block mt-1 rounded-xl border-slate-200 text-sm font-bold">
              <option v-for="month in 12" :key="month" :value="month">{{ month }}</option>
            </select>
          </label>
          <label class="hidden text-xs font-bold text-slate-500">
            Tahun
            <input v-model.number="commissionYear" @change="fetchCommissions" type="number" min="2000" max="2100" class="block mt-1 w-28 rounded-xl border-slate-200 text-sm font-bold" />
          </label>
          <span class="text-xs text-slate-400 pb-2">Tanggal: <strong class="text-slate-700">{{ commissionDate }}</strong></span>
        </div>
        <div class="bg-gradient-to-r from-emerald-600 to-teal-600 text-white p-6 rounded-[2rem] shadow-lg">
          <p class="text-[10px] font-black uppercase tracking-widest text-white/75">Total Komisi Operator</p>
          <p class="text-3xl font-black mt-1">{{ formatCurrency(totalCommission) }}</p>
          <p class="text-xs text-white/75 mt-1">Berdasarkan spanduk terpasang pada periode terpilih</p>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="bg-white rounded-[2rem] border border-emerald-200 p-5">
            <p class="text-[10px] font-black uppercase tracking-widest text-emerald-600">Sudah Dibayar</p>
            <p class="text-xl font-black text-slate-800 mt-1">{{ formatCurrency(paidCommission) }}</p>
          </div>
          <div class="bg-white rounded-[2rem] border border-amber-200 p-5">
            <p class="text-[10px] font-black uppercase tracking-widest text-amber-600">Belum Dibayar</p>
            <p class="text-xl font-black text-slate-800 mt-1">{{ formatCurrency(unpaidCommission) }}</p>
          </div>
        </div>
        <div class="bg-white rounded-[2rem] border border-slate-200 overflow-hidden shadow-sm">
          <div v-if="filteredCommissions.length === 0" class="p-16 text-center text-slate-400">
            Belum ada data komisi operator.
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-left">
              <thead class="bg-slate-50 text-[10px] uppercase tracking-wider text-slate-500">
                <tr>
                  <th class="px-6 py-4 font-black">Operator</th>
                  <th class="px-6 py-4 font-black">Area</th>
                  <th class="px-6 py-4 font-black text-center">Spanduk Terpasang</th>
                  <th class="px-6 py-4 font-black text-right">Total Komisi</th>
                  <th class="px-6 py-4 font-black text-right">Status Pembayaran</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="operator in filteredCommissions" :key="operator.id" class="hover:bg-slate-50">
                  <td class="px-6 py-4">
                    <p class="font-black text-slate-800">{{ operator.name }}</p>
                    <p class="text-xs text-slate-400">{{ operator.email || '-' }}</p>
                  </td>
                  <td class="px-6 py-4 text-sm text-slate-600">{{ operator.area || 'Semua Area' }}</td>
                  <td class="px-6 py-4 text-center font-bold text-slate-700">{{ operator.banner_terpasang || 0 }}</td>
                  <td class="px-6 py-4 text-right font-black text-emerald-600">{{ formatCurrency(operator.total_komisi) }}</td>
                  <td class="px-6 py-4 text-right">
                    <span :class="Number(operator.komisi_belum_dibayar || 0) > 0 ? 'text-amber-600' : 'text-emerald-600'" class="text-xs font-black">
                      {{ Number(operator.komisi_belum_dibayar || 0) > 0 ? 'Belum Lunas' : 'Lunas' }}
                    </span>
                    <button @click="updateCommissionPayment(operator)" class="block ml-auto mt-1 text-[10px] font-bold text-primary hover:underline">
                      {{ Number(operator.komisi_belum_dibayar || 0) > 0 ? 'Tandai Dibayar' : 'Batalkan Pembayaran' }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>

    <!-- Source Badge -->
    <div v-if="sourceStatus" class="flex items-center justify-end">
      <span class="text-[10px] font-bold text-slate-400 flex items-center gap-1.5">
        <CheckCircle2 class="w-3 h-3 text-emerald-500" />
        Sumber data: <strong class="uppercase">{{ sourceStatus }}</strong> dari Karsa HR Branding App
      </span>
    </div>
  </div>

  <!-- MODALS -->
  <!-- 1. Modal Upload Desain (Editor) -->
  <div
    v-if="showDesignModal"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-sm p-4 animate-in fade-in"
  >
    <div class="bg-white rounded-3xl p-6 w-full max-w-md shadow-2xl">
      <h3 class="text-lg font-black text-slate-900 mb-2">Upload Desain Spanduk</h3>
      <p class="text-sm text-slate-500 mb-6">
        Untuk toko: <strong class="text-slate-800">{{ selectedBanner?.shop_name }}</strong>
      </p>

      <div class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-slate-700 mb-1"
            >Pilih File Desain (PNG, JPG, JPEG, PDF)</label
          >
          <input
            type="file"
            @change="handleFileChange"
            accept=".png,.jpg,.jpeg,.pdf"
            class="w-full text-sm text-slate-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-semibold file:bg-primary/10 file:text-primary hover:file:bg-primary/20 transition-colors"
          />
          <!-- Preview nama file yang dipilih — penting untuk validasi ukuran banner -->
          <div
            v-if="designFile"
            class="mt-2 bg-amber-50 border border-amber-200 rounded-xl px-3 py-2"
          >
            <p class="text-[10px] font-bold text-amber-700 uppercase tracking-wide mb-0.5">
              📎 Nama file yang akan dikirim:
            </p>
            <p class="text-xs font-mono text-slate-800 break-all">{{ designFile.name }}</p>
            <p class="text-[10px] text-amber-600 mt-1">
              Nama file ini akan tersimpan di sistem untuk keperluan cross-check ukuran oleh
              validator.
            </p>
          </div>
        </div>
      </div>

      <div class="flex justify-end gap-3 mt-8">
        <button
          @click="handleCancelDesign"
          class="px-5 py-2.5 text-sm font-bold text-slate-500 hover:bg-slate-100 rounded-xl transition-colors"
        >
          Batal
        </button>
        <button
          @click="submitDesign"
          :disabled="isSubmitting || !designFile"
          class="px-5 py-2.5 bg-primary text-white text-sm font-bold rounded-xl shadow-lg shadow-primary/30 hover:bg-primary/90 transition-all disabled:opacity-50"
        >
          Upload & Lanjut
        </button>
      </div>
    </div>
  </div>

  <!-- 2. Modal Validasi Desain (Validator) -->
  <div
    v-if="showValidateModal"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-sm p-4 animate-in fade-in"
  >
    <div class="bg-white rounded-3xl p-6 w-full max-w-lg shadow-2xl max-h-[90vh] overflow-y-auto">
      <h3 class="text-lg font-black text-slate-900 mb-2">Validasi Desain Spanduk</h3>
      <p class="text-sm text-slate-500 mb-6">
        Untuk toko: <strong class="text-slate-800">{{ selectedBanner?.shop_name }}</strong>
      </p>

      <div class="space-y-6">
        <div class="bg-slate-50 rounded-2xl p-4 border border-slate-100">
          <p class="text-xs font-bold text-slate-700 mb-2">Desain Saat Ini:</p>
          <!-- Nama file asli — untuk cross-check ukuran banner oleh validator -->
          <div
            v-if="selectedBanner?.design_file_name"
            class="flex items-start gap-2 mb-3 bg-amber-50 border border-amber-200 rounded-xl px-3 py-2"
          >
            <span class="text-amber-500 text-sm mt-0.5">📎</span>
            <div>
              <p class="text-[10px] font-bold text-amber-700 uppercase tracking-wide">
                Nama File Desain:
              </p>
              <p class="text-xs font-mono text-slate-800 break-all">
                {{ selectedBanner.design_file_name }}
              </p>
            </div>
          </div>
          <img
            v-if="selectedBanner?.design_link"
            :src="selectedBanner.design_link"
            alt="Desain"
            class="w-full rounded-xl object-contain max-h-60"
          />
          <a
            v-if="selectedBanner?.design_link"
            :href="selectedBanner.design_link"
            target="_blank"
            class="mt-2 flex items-center gap-1 text-xs font-bold text-primary hover:underline"
          >
            <ExternalLink class="w-3 h-3" /> Buka di GDrive
          </a>
          <p v-if="!selectedBanner?.design_link" class="text-sm text-red-500 italic">Desain tidak ditemukan.</p>
        </div>

        <div>
          <label class="block text-xs font-bold text-slate-700 mb-1"
            >Catatan Revisi (Jika Ditolak)</label
          >
          <textarea
            v-model="validateNote"
            rows="3"
            placeholder="Contoh: Ukuran font terlalu kecil..."
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
          ></textarea>
        </div>
      </div>

      <div class="flex justify-end gap-3 mt-8">
        <button
          @click="showValidateModal = false"
          class="px-5 py-2.5 text-sm font-bold text-slate-500 hover:bg-slate-100 rounded-xl transition-colors"
        >
          Batal
        </button>
        <button
          @click="submitValidation(false)"
          :disabled="isSubmitting || !validateNote"
          class="px-5 py-2.5 bg-red-100 text-red-700 hover:bg-red-200 text-sm font-bold rounded-xl transition-colors disabled:opacity-50"
        >
          Tolak (Revisi)
        </button>
        <button
          @click="submitValidation(true)"
          :disabled="isSubmitting"
          class="px-5 py-2.5 bg-emerald-500 text-white text-sm font-bold rounded-xl shadow-lg shadow-emerald-500/30 hover:bg-emerald-600 transition-all disabled:opacity-50"
        >
          Setujui Desain
        </button>
      </div>
    </div>
  </div>

  <!-- 3. Modal Proses Vendor (Purchasing) -->
  <div
    v-if="showVendorModal"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-sm p-4 animate-in fade-in"
  >
    <div class="bg-white rounded-3xl p-6 w-full max-w-md shadow-2xl">
      <h3 class="text-lg font-black text-slate-900 mb-2">Proses Vendor Spanduk</h3>
      <p class="text-sm text-slate-500 mb-6">
        Untuk toko: <strong class="text-slate-800">{{ selectedBanner?.shop_name }}</strong>
      </p>

      <div class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-slate-700 mb-1">Nama Vendor Pencetak</label>
          <input
            v-model="vendorName"
            type="text"
            placeholder="Masukkan nama vendor..."
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
          />
        </div>
        <div>
          <label class="block text-xs font-bold text-slate-700 mb-1">Total Biaya (Rp)</label>
          <input
            v-model="vendorPrice"
            type="number"
            placeholder="Contoh: 150000"
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
          />
        </div>
      </div>

      <div class="flex justify-end gap-3 mt-8">
        <button
          @click="showVendorModal = false"
          class="px-5 py-2.5 text-sm font-bold text-slate-500 hover:bg-slate-100 rounded-xl transition-colors"
        >
          Batal
        </button>
        <button
          @click="submitVendor"
          :disabled="isSubmitting || !vendorName"
          class="px-5 py-2.5 bg-primary text-white text-sm font-bold rounded-xl shadow-lg shadow-primary/30 hover:bg-primary/90 transition-all disabled:opacity-50"
        >
          Proses ke Vendor
        </button>
      </div>
    </div>
  </div>

  <!-- 4. Modal Set Target Operator -->
  <div
    v-if="showTargetModal"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 backdrop-blur-sm p-4 animate-in fade-in"
  >
    <div class="bg-white rounded-3xl p-6 w-full max-w-md shadow-2xl">
      <h3 class="text-lg font-black text-slate-900 mb-2">Set Target Harian</h3>
      <p class="text-sm text-slate-500 mb-6">
        Untuk Operator: <strong class="text-slate-800">{{ selectedOperator?.name }}</strong>
      </p>

      <div class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-slate-700 mb-1">Target Minimal Pasang Harian (Jumlah Toko)</label>
          <input
            v-model="operatorTargetValue"
            type="number"
            min="0"
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm font-bold focus:outline-none focus:ring-2 focus:ring-primary/20"
          />
        </div>
      </div>

      <div class="flex justify-end gap-3 mt-8">
        <button
          @click="showTargetModal = false"
          class="px-5 py-2.5 text-sm font-bold text-slate-500 hover:bg-slate-100 rounded-xl transition-colors"
        >
          Batal
        </button>
        <button
          @click="saveOperatorTarget"
          :disabled="isSubmitting || operatorTargetValue < 0"
          class="px-5 py-2.5 bg-primary text-white text-sm font-bold rounded-xl shadow-lg shadow-primary/30 hover:bg-primary/90 transition-all disabled:opacity-50"
        >
          Simpan Target
        </button>
      </div>
    </div>
  </div>

  <!-- 5. Modal Bukti Pasang (Photo Installed) -->
  <div
    v-if="showProofModal"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4 animate-in fade-in"
  >
    <div class="bg-white rounded-3xl p-6 w-full max-w-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
      <div class="flex items-start justify-between gap-4 mb-4 pb-3 border-b border-slate-100">
        <div>
          <div class="flex items-center gap-2">
            <span class="px-2.5 py-0.5 rounded-full text-[10px] font-black uppercase tracking-wider bg-emerald-500/10 text-emerald-700 border border-emerald-200">
              Bukti Pemasangan
            </span>
            <span v-if="proofData?.installed_at" class="text-xs font-semibold text-slate-400">
              {{ new Date(proofData.installed_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' }) }}
            </span>
          </div>
          <h3 class="text-lg font-black text-slate-900 mt-1">
            {{ proofData?.shop_name || 'Toko' }}
          </h3>
          <p class="text-xs text-slate-500">
            Sales: <strong class="text-slate-700">{{ proofData?.sales_name || '-' }}</strong> 
            <span v-if="proofData?.operator_name" class="ml-2">| Operator: <strong class="text-slate-700">{{ proofData?.operator_name }}</strong></span>
          </p>
        </div>

        <button
          @click="showProofModal = false"
          class="w-9 h-9 rounded-xl bg-slate-100 text-slate-500 hover:bg-slate-200 flex items-center justify-center font-bold text-sm transition-colors shrink-0"
        >
          ✕
        </button>
      </div>

      <!-- Image Area -->
      <div class="flex-1 overflow-y-auto rounded-2xl bg-slate-950 flex items-center justify-center p-2 min-h-[250px]">
        <img
          v-if="proofData?.photo_installed"
          :src="resolvePhotoUrl(proofData.photo_installed)"
          alt="Bukti Pasang"
          class="max-w-full max-h-[55vh] object-contain rounded-xl"
        />
        <div v-else class="text-center p-8 text-slate-400 text-sm">
          <Camera class="w-12 h-12 mx-auto mb-2 opacity-40 text-slate-400" />
          Foto bukti pasang tidak tersedia
        </div>
      </div>

      <!-- Footer / Actions -->
      <div class="flex items-center justify-between gap-3 mt-4 pt-3 border-t border-slate-100">
        <div class="text-xs text-slate-500">
          Ukuran: <span class="font-bold text-slate-700">{{ proofData?.size_length }}cm × {{ proofData?.size_width }}cm</span> ({{ proofData?.qty }} pcs)
        </div>

        <div class="flex items-center gap-2">
          <a
            v-if="proofData?.photo_installed"
            :href="resolvePhotoUrl(proofData.photo_installed)"
            target="_blank"
            class="flex items-center gap-1.5 px-4 py-2 bg-slate-100 text-slate-700 hover:bg-slate-200 rounded-xl text-xs font-bold transition-colors"
          >
            <ExternalLink class="w-3.5 h-3.5" />
            Buka Resolusi Penuh
          </a>
          <button
            @click="showProofModal = false"
            class="px-5 py-2 bg-primary text-white rounded-xl text-xs font-bold hover:bg-primary/90 transition-all shadow-sm"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
