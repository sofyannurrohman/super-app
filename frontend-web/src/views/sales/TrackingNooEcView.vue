<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import {
  TrendingUp,
  ShoppingBag,
  Store,
  Users,
  Calendar,
  Clock,
  MapPin,
  Sparkles,
  Phone,
  Search,
  FileDown,
  RefreshCw,
  Loader2,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  MessageSquare,
  Building2,
  Filter,
  BarChart3,
  Layers
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const masterDataStore = useMasterDataStore()

// Filter States
const periodPreset = ref<'TODAY' | 'YESTERDAY' | 'LAST_7_DAYS' | 'THIS_MONTH' | 'CUSTOM'>('TODAY')
const startDate = ref('')
const endDate = ref('')
const selectedSalesman = ref('ALL')
const activeTab = ref<'SALESMAN_MATRIX' | 'STORE_LOGS'>('SALESMAN_MATRIX')
const storeFilterType = ref<'ALL' | 'NOO_ONLY' | 'EC_ONLY' | 'NON_EC_ONLY'>('ALL')
const searchQuery = ref('')

const isLoading = ref(true)
const isExporting = ref(false)
const trackingData = ref<any>({
  start_date: '',
  end_date: '',
  total_calls: 0,
  total_ec: 0,
  ec_ratio: 0,
  total_noo: 0,
  total_order_amount: 0,
  salesman_matrix: [],
  store_logs: []
})
const salesmenList = ref<any[]>([])

// Date Presets Calculation
const setPeriodPreset = (preset: 'TODAY' | 'YESTERDAY' | 'LAST_7_DAYS' | 'THIS_MONTH' | 'CUSTOM') => {
  periodPreset.value = preset
  const today = new Date()
  const formatDateStr = (d: Date): string => d.toISOString().split('T')[0] || ''

  if (preset === 'TODAY') {
    startDate.value = formatDateStr(today)
    endDate.value = formatDateStr(today)
  } else if (preset === 'YESTERDAY') {
    const yest = new Date(today)
    yest.setDate(yest.getDate() - 1)
    startDate.value = formatDateStr(yest)
    endDate.value = formatDateStr(yest)
  } else if (preset === 'LAST_7_DAYS') {
    const last7 = new Date(today)
    last7.setDate(last7.getDate() - 6)
    startDate.value = formatDateStr(last7)
    endDate.value = formatDateStr(today)
  } else if (preset === 'THIS_MONTH') {
    const firstDay = new Date(today.getFullYear(), today.getMonth(), 1)
    startDate.value = formatDateStr(firstDay)
    endDate.value = formatDateStr(today)
  }
  fetchTrackingData()
}

// Formatters
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount || 0)
}

const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

const formatTime = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
  }) + ' WIB'
}

// Fetch Data
const fetchTrackingData = async () => {
  isLoading.value = true
  try {
    const params = new URLSearchParams()
    if (startDate.value) params.append('start_date', startDate.value)
    if (endDate.value) params.append('end_date', endDate.value)

    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.append('branch_id', masterDataStore.selectedBranchId)
    }
    if (selectedSalesman.value && selectedSalesman.value !== 'ALL') {
      params.append('employee_id', selectedSalesman.value)
    }
    if (masterDataStore.selectedSalesType && masterDataStore.selectedSalesType !== 'ALL_SALES') {
      params.append('sales_type', masterDataStore.selectedSalesType)
    }

    const res = await apiClient.get(`/admin/sales/tracking/noo-ec?${params.toString()}`)
    trackingData.value = res.data.data || {
      start_date: startDate.value,
      end_date: endDate.value,
      total_calls: 0,
      total_ec: 0,
      ec_ratio: 0,
      total_noo: 0,
      total_order_amount: 0,
      salesman_matrix: [],
      store_logs: []
    }
  } catch (error: any) {
    console.error('Error fetching NOO & EC tracking:', error)
    toast.error(error.response?.data?.error || 'Gagal memuat data tracking NOO & EC')
  } finally {
    isLoading.value = false
  }
}

const fetchSalesmen = async () => {
  try {
    const branchFilter = masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES'
      ? `&branch_id=${masterDataStore.selectedBranchId}`
      : ''
    const res = await apiClient.get(`/employees?limit=all&is_sales=true${branchFilter}&sales_type=${masterDataStore.selectedSalesType}`)
    salesmenList.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load salesmen list:', e)
  }
}

const handleExportExcel = async () => {
  isExporting.value = true
  try {
    const params = new URLSearchParams()
    if (startDate.value) params.append('start_date', startDate.value)
    if (endDate.value) params.append('end_date', endDate.value)
    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.append('branch_id', masterDataStore.selectedBranchId)
    }
    if (selectedSalesman.value && selectedSalesman.value !== 'ALL') {
      params.append('employee_id', selectedSalesman.value)
    }
    if (masterDataStore.selectedSalesType && masterDataStore.selectedSalesType !== 'ALL_SALES') {
      params.append('sales_type', masterDataStore.selectedSalesType)
    }

    const res = await apiClient.get(`/admin/sales/tracking/noo-ec/export?${params.toString()}`, {
      responseType: 'blob',
    })

    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `tracking_noo_ec_${startDate.value}_sd_${endDate.value}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)

    toast.success('Laporan Excel NOO & EC berhasil diunduh!')
  } catch (error) {
    toast.error('Gagal mengekspor data laporan')
  } finally {
    isExporting.value = false
  }
}

const filterStoresForSalesman = (empId: string) => {
  selectedSalesman.value = empId
  activeTab.value = 'STORE_LOGS'
  fetchTrackingData()
}

const resetSalesmanFilter = () => {
  selectedSalesman.value = 'ALL'
  fetchTrackingData()
}

// Filtered Lists
const filteredSalesmanMatrix = computed(() => {
  const list = trackingData.value.salesman_matrix || []
  if (!searchQuery.value) return list
  const q = searchQuery.value.toLowerCase()
  return list.filter((s: any) =>
    (s.employee_name && s.employee_name.toLowerCase().includes(q)) ||
    (s.employee_id_number && s.employee_id_number.toLowerCase().includes(q)) ||
    (s.sales_area && s.sales_area.toLowerCase().includes(q)) ||
    (s.job_title && s.job_title.toLowerCase().includes(q)) ||
    (s.branch_name && s.branch_name.toLowerCase().includes(q))
  )
})

const filteredStoreLogs = computed(() => {
  let list = trackingData.value.store_logs || []

  // Filter type
  if (storeFilterType.value === 'NOO_ONLY') {
    list = list.filter((s: any) => s.is_noo_in_period)
  } else if (storeFilterType.value === 'EC_ONLY') {
    list = list.filter((s: any) => s.is_effective_call)
  } else if (storeFilterType.value === 'NON_EC_ONLY') {
    list = list.filter((s: any) => !s.is_effective_call)
  }

  // Search query
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter((s: any) =>
      (s.store_name && s.store_name.toLowerCase().includes(q)) ||
      (s.store_id && s.store_id.toLowerCase().includes(q)) ||
      (s.owner_name && s.owner_name.toLowerCase().includes(q)) ||
      (s.salesman_name && s.salesman_name.toLowerCase().includes(q)) ||
      (s.salesman_id_number && s.salesman_id_number.toLowerCase().includes(q)) ||
      (s.salesman_area && s.salesman_area.toLowerCase().includes(q)) ||
      (s.address && s.address.toLowerCase().includes(q)) ||
      (s.city && s.city.toLowerCase().includes(q)) ||
      (s.latest_order_number && s.latest_order_number.toLowerCase().includes(q))
    )
  }

  return list
})

const openWhatsApp = (phone: string, storeName: string, salesmanName: string) => {
  if (!phone || phone === '-') {
    toast.info('Nomor WhatsApp toko tidak tersedia')
    return
  }
  let cleanPhone = phone.replace(/[^0-9]/g, '')
  if (cleanPhone.startsWith('0')) {
    cleanPhone = '62' + cleanPhone.substring(1)
  } else if (!cleanPhone.startsWith('62')) {
    cleanPhone = '62' + cleanPhone
  }
  const text = `Halo Kak/Owner Toko *${storeName}*,\n\nTerima kasih atas kerja samanya bersama PT Karsa HR Purnomo Jaya melalui Sales kami (${salesmanName}). Ada yang bisa kami bantu kembali? 🙏`
  window.open(`https://wa.me/${cleanPhone}?text=${encodeURIComponent(text)}`, '_blank')
}

// Watchers
watch(
  () => [masterDataStore.selectedBranchId, masterDataStore.selectedSalesType],
  () => {
    fetchSalesmen()
    fetchTrackingData()
  }
)

onMounted(() => {
  const today = new Date().toISOString().split('T')[0] || ''
  startDate.value = today
  endDate.value = today
  fetchSalesmen()
  fetchTrackingData()
})
</script>

<template>
  <div class="space-y-6 pb-12">
    <!-- Header Title & Action Buttons -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5">
          <div class="w-10 h-10 rounded-2xl bg-indigo-600 text-white flex items-center justify-center shadow-md shadow-indigo-600/20">
            <TrendingUp class="w-5 h-5" />
          </div>
          <div>
            <h1 class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
              <span>Tracking NOO & Effective Call (EC)</span>
              <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl sm:text-2xl font-bold">
                — {{ masterDataStore.selectedBranchName }}
              </span>
            </h1>
            <p class="text-xs sm:text-sm font-medium text-slate-500 mt-0.5">
              Monitoring real-time akuisisi outlet baru (NOO), efektivitas kunjungan (EC), dan performa omzet per salesman.
            </p>
          </div>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2 sm:gap-3">
        <Button
          @click="fetchTrackingData"
          variant="outline"
          :disabled="isLoading"
          class="rounded-xl sm:rounded-2xl px-4 py-2.5 font-bold text-xs sm:text-sm border-slate-200 hover:bg-slate-50 flex items-center gap-2 cursor-pointer h-auto"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isLoading }" />
          <span>Segarkan Data</span>
        </Button>
        <Button
          @click="handleExportExcel"
          :disabled="isExporting"
          class="bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl sm:rounded-2xl px-4 sm:px-5 py-2.5 font-bold text-xs sm:text-sm shadow-md shadow-emerald-600/20 flex items-center gap-2 cursor-pointer h-auto disabled:opacity-60"
        >
          <Loader2 v-if="isExporting" class="w-4 h-4 animate-spin" />
          <FileDown v-else class="w-4 h-4" />
          <span>{{ isExporting ? 'Mengekspor...' : 'Export Laporan Excel' }}</span>
        </Button>
      </div>
    </div>

    <!-- Filter Control Bar -->
    <div class="bg-white p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-slate-200 shadow-sm space-y-4">
      <!-- Period Presets -->
      <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 pb-3 border-b border-slate-100">
        <div class="flex items-center gap-1.5 flex-wrap">
          <span class="text-xs font-bold text-slate-400 mr-1 flex items-center gap-1">
            <Calendar class="w-3.5 h-3.5 text-slate-400" /> Periode:
          </span>
          <button
            @click="setPeriodPreset('TODAY')"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="periodPreset === 'TODAY' ? 'bg-slate-900 text-white shadow-sm' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
          >
            Hari Ini
          </button>
          <button
            @click="setPeriodPreset('YESTERDAY')"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="periodPreset === 'YESTERDAY' ? 'bg-slate-900 text-white shadow-sm' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
          >
            Kemarin
          </button>
          <button
            @click="setPeriodPreset('LAST_7_DAYS')"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="periodPreset === 'LAST_7_DAYS' ? 'bg-slate-900 text-white shadow-sm' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
          >
            7 Hari Terakhir
          </button>
          <button
            @click="setPeriodPreset('THIS_MONTH')"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="periodPreset === 'THIS_MONTH' ? 'bg-slate-900 text-white shadow-sm' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
          >
            Bulan Ini
          </button>
          <button
            @click="periodPreset = 'CUSTOM'"
            class="px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
            :class="periodPreset === 'CUSTOM' ? 'bg-slate-900 text-white shadow-sm' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
          >
            Rentang Khusus
          </button>
        </div>

        <!-- Custom Date Pickers -->
        <div v-if="periodPreset === 'CUSTOM'" class="flex items-center gap-2 flex-wrap">
          <input
            v-model="startDate"
            type="date"
            class="px-3 py-1.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
          />
          <span class="text-xs text-slate-400 font-bold">s/d</span>
          <input
            v-model="endDate"
            type="date"
            class="px-3 py-1.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
          />
          <Button
            @click="fetchTrackingData"
            size="sm"
            class="bg-primary hover:bg-primary/90 text-white rounded-xl text-xs font-bold px-3 py-1.5 h-auto cursor-pointer"
          >
            Terapkan
          </Button>
        </div>
      </div>

      <!-- Second Row: Search, Sales Type & Salesman Filter -->
      <div class="flex flex-col sm:flex-row gap-3 items-center flex-wrap">
        <div class="relative flex-1 w-full min-w-[220px]">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="activeTab === 'SALESMAN_MATRIX' ? 'Cari nama salesman, NIK, area kerja, atau cabang...' : 'Cari nama toko, ID toko, pemilik, salesman, atau area...'"
            class="w-full pl-11 pr-4 py-2.5 sm:py-3 bg-slate-50 border border-slate-200/80 rounded-xl sm:rounded-2xl focus:ring-2 focus:ring-primary/20 transition-all font-medium text-xs sm:text-sm text-slate-900"
          />
        </div>

        <select
          :value="masterDataStore.selectedSalesType"
          @change="(e) => { selectedSalesman = 'ALL'; masterDataStore.setSelectedSalesType((e.target as HTMLSelectElement).value) }"
          class="px-4 py-2.5 sm:py-3 bg-slate-50 border border-slate-200/80 rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 w-full sm:w-auto cursor-pointer"
        >
          <option value="ALL_SALES">Semua Tipe Sales</option>
          <option value="sales_motoris_booster">Sales Motoris & Booster</option>
          <option value="sales_to">Sales TO</option>
          <option value="sales_mt">Sales MT</option>
        </select>

        <select
          v-model="selectedSalesman"
          @change="fetchTrackingData"
          class="px-4 py-2.5 sm:py-3 bg-slate-50 border border-slate-200/80 rounded-xl sm:rounded-2xl font-bold text-xs sm:text-sm text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 w-full sm:w-auto cursor-pointer"
        >
          <option value="ALL">Semua Salesman PIC</option>
          <option v-for="s in salesmenList" :key="s.id" :value="s.id">
            {{ s.first_name }} {{ s.last_name }} ({{ s.job_position?.title || 'Sales' }})
          </option>
        </select>

        <div v-if="selectedSalesman !== 'ALL'" class="w-full sm:w-auto">
          <Button
            @click="resetSalesmanFilter"
            variant="ghost"
            size="sm"
            class="text-xs font-bold text-rose-600 hover:text-rose-700 hover:bg-rose-50 rounded-xl cursor-pointer"
          >
            Hapus Filter Sales
          </Button>
        </div>
      </div>
    </div>

    <!-- Executive KPI Summary Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-5 gap-3 sm:gap-4">
      <!-- Total Call -->
      <div class="bg-white p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-slate-200 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-[10px] sm:text-xs font-bold text-slate-400 uppercase tracking-wider">Total Kunjungan</span>
          <div class="w-8 h-8 rounded-xl bg-slate-100 flex items-center justify-center text-slate-600">
            <MapPin class="w-4 h-4" />
          </div>
        </div>
        <p class="text-xl sm:text-2xl font-black text-slate-900 mt-2">{{ trackingData.total_calls }}</p>
        <p class="text-[10px] text-slate-400 font-medium mt-0.5">Check-in Call Lapangan</p>
      </div>

      <!-- Total EC -->
      <div class="bg-gradient-to-br from-emerald-50 to-teal-50/40 p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-emerald-200/80 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-[10px] sm:text-xs font-bold text-emerald-700 uppercase tracking-wider">Effective Call (EC)</span>
          <div class="w-8 h-8 rounded-xl bg-emerald-100 flex items-center justify-center text-emerald-700">
            <ShoppingBag class="w-4 h-4" />
          </div>
        </div>
        <p class="text-xl sm:text-2xl font-black text-emerald-900 mt-2">{{ trackingData.total_ec }}</p>
        <p class="text-[10px] text-emerald-600 font-bold mt-0.5">Toko Menghasilkan Order</p>
      </div>

      <!-- Rasio EC % -->
      <div class="bg-gradient-to-br from-indigo-50 to-purple-50/40 p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-indigo-200/80 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-[10px] sm:text-xs font-bold text-indigo-700 uppercase tracking-wider">Rasio EC %</span>
          <div class="w-8 h-8 rounded-xl bg-indigo-100 flex items-center justify-center text-indigo-700">
            <BarChart3 class="w-4 h-4" />
          </div>
        </div>
        <p class="text-xl sm:text-2xl font-black text-indigo-900 mt-2">{{ trackingData.ec_ratio }}%</p>
        <div class="w-full bg-indigo-200/60 rounded-full h-1.5 mt-2 overflow-hidden">
          <div
            class="bg-indigo-600 h-1.5 rounded-full transition-all duration-500"
            :style="{ width: `${Math.min(100, trackingData.ec_ratio)}%` }"
          ></div>
        </div>
      </div>

      <!-- Total NOO -->
      <div class="bg-gradient-to-br from-blue-50 to-cyan-50/40 p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-blue-200/80 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-[10px] sm:text-xs font-bold text-blue-700 uppercase tracking-wider">Toko Baru (NOO)</span>
          <div class="w-8 h-8 rounded-xl bg-blue-100 flex items-center justify-center text-blue-700">
            <Sparkles class="w-4 h-4" />
          </div>
        </div>
        <p class="text-xl sm:text-2xl font-black text-blue-900 mt-2">{{ trackingData.total_noo }}</p>
        <p class="text-[10px] text-blue-600 font-bold mt-0.5">Outlet Baru Diakuisisi</p>
      </div>

      <!-- Total Omzet Pesanan -->
      <div class="col-span-2 lg:col-span-1 bg-white p-4 sm:p-5 rounded-2xl sm:rounded-3xl border border-slate-200 shadow-sm">
        <div class="flex items-center justify-between">
          <span class="text-[10px] sm:text-xs font-bold text-slate-400 uppercase tracking-wider">Total Nilai Order</span>
          <div class="w-8 h-8 rounded-xl bg-emerald-50 flex items-center justify-center text-emerald-600">
            <TrendingUp class="w-4 h-4" />
          </div>
        </div>
        <p class="text-base sm:text-lg font-black text-emerald-700 mt-2 truncate">{{ formatCurrency(trackingData.total_order_amount) }}</p>
        <p class="text-[10px] text-slate-400 font-medium mt-0.5">Omzet Pesanan Penjualan</p>
      </div>
    </div>

    <!-- Dual Tab Navigation & Main Tables -->
    <div class="bg-white rounded-2xl sm:rounded-[32px] border border-slate-200 shadow-sm overflow-hidden flex flex-col">
      <!-- Tab Header -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between px-4 sm:px-6 pt-4 sm:pt-6 pb-4 border-b border-slate-100 gap-3">
        <div class="flex items-center gap-2 p-1 bg-slate-100 rounded-2xl w-full sm:w-auto">
          <button
            @click="activeTab = 'SALESMAN_MATRIX'"
            class="flex-1 sm:flex-none px-4 sm:px-5 py-2.5 rounded-xl font-black text-xs transition-all flex items-center justify-center gap-2 cursor-pointer"
            :class="activeTab === 'SALESMAN_MATRIX' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'"
          >
            <Users class="w-4 h-4" />
            <span>Matriks Performa Salesman ({{ filteredSalesmanMatrix.length }})</span>
          </button>
          <button
            @click="activeTab = 'STORE_LOGS'"
            class="flex-1 sm:flex-none px-4 sm:px-5 py-2.5 rounded-xl font-black text-xs transition-all flex items-center justify-center gap-2 cursor-pointer"
            :class="activeTab === 'STORE_LOGS' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'"
          >
            <Store class="w-4 h-4" />
            <span>Log Outlet & Pesanan ({{ filteredStoreLogs.length }})</span>
          </button>
        </div>

        <!-- Sub Filter for Store Logs -->
        <div v-if="activeTab === 'STORE_LOGS'" class="flex items-center gap-1.5 flex-wrap">
          <span class="text-xs font-bold text-slate-400 mr-1">Filter Toko:</span>
          <button
            @click="storeFilterType = 'ALL'"
            class="px-3 py-1.5 rounded-xl text-[11px] font-bold transition-all cursor-pointer"
            :class="storeFilterType === 'ALL' ? 'bg-slate-900 text-white' : 'bg-slate-50 text-slate-600 hover:bg-slate-100'"
          >
            Semua
          </button>
          <button
            @click="storeFilterType = 'NOO_ONLY'"
            class="px-3 py-1.5 rounded-xl text-[11px] font-bold transition-all cursor-pointer flex items-center gap-1"
            :class="storeFilterType === 'NOO_ONLY' ? 'bg-blue-600 text-white' : 'bg-blue-50 text-blue-700 hover:bg-blue-100'"
          >
            <Sparkles class="w-3 h-3" /> Hanya NOO
          </button>
          <button
            @click="storeFilterType = 'EC_ONLY'"
            class="px-3 py-1.5 rounded-xl text-[11px] font-bold transition-all cursor-pointer flex items-center gap-1"
            :class="storeFilterType === 'EC_ONLY' ? 'bg-emerald-600 text-white' : 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100'"
          >
            <ShoppingBag class="w-3 h-3" /> Hanya EC
          </button>
          <button
            @click="storeFilterType = 'NON_EC_ONLY'"
            class="px-3 py-1.5 rounded-xl text-[11px] font-bold transition-all cursor-pointer"
            :class="storeFilterType === 'NON_EC_ONLY' ? 'bg-slate-700 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
          >
            Non-EC
          </button>
        </div>
      </div>

      <!-- TAB 1: SALESMAN MATRIX TABLE -->
      <div v-if="activeTab === 'SALESMAN_MATRIX'" class="overflow-x-auto no-scrollbar">
        <table class="w-full border-collapse text-left">
          <thead class="bg-slate-50/60 border-b border-slate-100">
            <tr>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center w-12">No.</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">ID / NIK</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">Salesman</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">Area Sales Bekerja</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">Cabang</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">Toko Terkelola</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">Kunjungan (Call)</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">Effective Call (EC)</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">Rasio EC</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">Toko Baru (NOO)</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right">Nilai Omzet Order</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr v-if="isLoading" v-for="i in 4" :key="i" class="animate-pulse">
              <td v-for="j in 12" :key="j" class="p-5">
                <div class="h-4 bg-slate-100 rounded w-full"></div>
              </td>
            </tr>

            <tr v-else-if="filteredSalesmanMatrix.length === 0">
              <td colspan="12" class="py-16 text-center text-slate-400">
                <Users class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                <p class="font-bold text-sm">Tidak ada data performa salesman pada periode ini.</p>
              </td>
            </tr>

            <tr
              v-else
              v-for="(sm, idx) in filteredSalesmanMatrix"
              :key="sm.employee_id"
              class="hover:bg-slate-50/60 transition-colors group"
            >
              <td class="p-5 text-center text-xs font-bold text-slate-400">
                {{ Number(idx) + 1 }}
              </td>

              <td class="p-5 whitespace-nowrap">
                <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-mono font-bold bg-slate-100 text-slate-700 border border-slate-200">
                  {{ sm.employee_id_number || '-' }}
                </span>
              </td>

              <td class="p-5">
                <div class="flex items-center gap-3">
                  <div class="w-9 h-9 rounded-xl bg-primary/10 text-primary flex items-center justify-center font-bold text-xs shrink-0">
                    {{ sm.employee_name.substring(0, 2).toUpperCase() }}
                  </div>
                  <div>
                    <p class="text-sm font-black text-slate-900">{{ sm.employee_name }}</p>
                    <p class="text-[11px] font-semibold text-slate-500">{{ sm.job_title }}</p>
                  </div>
                </div>
              </td>

              <td class="p-5 whitespace-nowrap">
                <div class="flex items-center gap-1.5">
                  <MapPin class="w-3.5 h-3.5 text-blue-500 shrink-0" />
                  <span class="text-xs font-bold text-blue-700">{{ sm.sales_area || '-' }}</span>
                </div>
              </td>

              <td class="p-5">
                <span class="text-xs font-bold text-slate-700">{{ sm.branch_name || 'Pusat' }}</span>
              </td>

              <td class="p-5 text-center">
                <span class="text-xs font-bold text-slate-800">{{ sm.total_assigned_stores }} Outlet</span>
              </td>

              <td class="p-5 text-center">
                <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-black bg-slate-100 text-slate-800">
                  {{ sm.total_visits }}
                </span>
              </td>

              <td class="p-5 text-center">
                <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-black bg-emerald-100 text-emerald-800">
                  {{ sm.effective_calls }}
                </span>
              </td>

              <td class="p-5 text-center">
                <div class="flex flex-col items-center gap-1">
                  <span
                    :class="[
                      'text-xs font-black',
                      sm.ec_ratio >= 60 ? 'text-emerald-600' : sm.ec_ratio >= 30 ? 'text-amber-600' : 'text-slate-600'
                    ]"
                  >
                    {{ sm.ec_ratio }}%
                  </span>
                  <div class="w-16 bg-slate-100 rounded-full h-1.5 overflow-hidden">
                    <div
                      class="h-1.5 rounded-full"
                      :class="sm.ec_ratio >= 60 ? 'bg-emerald-500' : sm.ec_ratio >= 30 ? 'bg-amber-500' : 'bg-slate-400'"
                      :style="{ width: `${Math.min(100, sm.ec_ratio)}%` }"
                    ></div>
                  </div>
                </div>
              </td>

              <td class="p-5 text-center">
                <span
                  :class="[
                    'inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-black',
                    sm.noo_count > 0 ? 'bg-blue-100 text-blue-800' : 'bg-slate-100 text-slate-500'
                  ]"
                >
                  <Sparkles v-if="sm.noo_count > 0" class="w-3 h-3" />
                  {{ sm.noo_count }} NOO
                </span>
              </td>

              <td class="p-5 text-right">
                <p class="text-sm font-black text-slate-900">{{ formatCurrency(sm.total_order_amount) }}</p>
                <p class="text-[10px] font-bold text-slate-400">{{ sm.total_order_count }} Dokumen Pesanan</p>
              </td>

              <td class="p-5 text-right">
                <Button
                  @click="filterStoresForSalesman(sm.employee_id)"
                  size="sm"
                  variant="outline"
                  class="text-xs font-bold gap-1.5 rounded-xl border-slate-200 hover:bg-slate-100 cursor-pointer"
                >
                  <span>Lihat Toko</span>
                  <ExternalLink class="w-3 h-3" />
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- TAB 2: STORE & ORDER LOGS TABLE -->
      <div v-else class="overflow-x-auto no-scrollbar">
        <table class="w-full border-collapse text-left">
          <thead class="bg-slate-50/60 border-b border-slate-100">
            <tr>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center w-12">No.</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">ID Toko</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">Toko / Outlet</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest whitespace-nowrap">Waktu Dibuat</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">Status NOO</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">Salesman PIC</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">Area Sales Bekerja</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">Waktu Kunjungan</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center">Status EC</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right">Pesanan Penjualan</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest">Wilayah / Alamat</th>
              <th class="p-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr v-if="isLoading" v-for="i in 5" :key="i" class="animate-pulse">
              <td v-for="j in 12" :key="j" class="p-5">
                <div class="h-4 bg-slate-100 rounded w-full"></div>
              </td>
            </tr>

            <tr v-else-if="filteredStoreLogs.length === 0">
              <td colspan="12" class="py-16 text-center text-slate-400">
                <Store class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                <p class="font-bold text-sm">Tidak ada log toko / aktivitas pesanan pada filter ini.</p>
              </td>
            </tr>

            <tr
              v-else
              v-for="(st, idx) in filteredStoreLogs"
              :key="st.store_id"
              class="hover:bg-slate-50/60 transition-colors group"
            >
              <td class="p-5 text-center text-xs font-bold text-slate-400">
                {{ Number(idx) + 1 }}
              </td>

              <td class="p-5 whitespace-nowrap">
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-mono font-bold bg-slate-100 text-slate-600 border border-slate-200"
                  :title="st.store_id"
                >
                  {{ st.store_id ? st.store_id.substring(0, 8) + '...' : '-' }}
                </span>
              </td>

              <td class="p-5">
                <div class="flex items-center gap-3">
                  <div class="w-9 h-9 rounded-xl bg-slate-100 flex items-center justify-center text-slate-500 shrink-0">
                    <Store class="w-4 h-4" />
                  </div>
                  <div>
                    <p class="text-sm font-black text-slate-900 leading-tight">{{ st.store_name }}</p>
                    <div class="flex items-center gap-2 mt-0.5">
                      <span class="text-[10px] font-bold text-slate-400">Owner: {{ st.owner_name }}</span>
                      <span class="text-slate-300">•</span>
                      <span class="text-[10px] font-semibold text-slate-500">{{ st.phone_number }}</span>
                    </div>
                  </div>
                </div>
              </td>

              <td class="p-5 whitespace-nowrap">
                <div class="flex flex-col">
                  <span class="text-xs font-bold text-slate-800 flex items-center gap-1">
                    <Calendar class="w-3 h-3 text-slate-400" /> {{ formatDate(st.created_at) }}
                  </span>
                  <span class="text-[11px] font-medium text-slate-500 mt-0.5 flex items-center gap-1">
                    <Clock class="w-3 h-3 text-slate-400" /> {{ formatTime(st.created_at) }}
                  </span>
                </div>
              </td>

              <td class="p-5 text-center">
                <span
                  v-if="st.is_noo_in_period"
                  class="inline-flex items-center gap-1 px-2.5 py-1 text-[10px] font-black bg-blue-600 text-white rounded-md uppercase tracking-wider shadow-sm shadow-blue-500/20"
                >
                  <Sparkles class="w-2.5 h-2.5" /> NOO PERIODE INI
                </span>
                <span
                  v-else
                  class="inline-flex items-center px-2 py-0.5 text-[9px] font-bold bg-slate-100 text-slate-500 rounded uppercase"
                >
                  Existing
                </span>
              </td>

              <td class="p-5">
                <div>
                  <p class="text-xs font-bold text-slate-800">{{ st.salesman_name }}</p>
                  <p v-if="st.salesman_id_number && st.salesman_id_number !== '-'" class="text-[10px] font-mono text-slate-400 mt-0.5">
                    NIK: {{ st.salesman_id_number }}
                  </p>
                </div>
              </td>

              <td class="p-5 whitespace-nowrap">
                <div class="flex items-center gap-1.5">
                  <MapPin class="w-3.5 h-3.5 text-blue-500 shrink-0" />
                  <span class="text-xs font-bold text-blue-700">{{ st.salesman_area || '-' }}</span>
                </div>
              </td>

              <td class="p-5 text-center whitespace-nowrap">
                <div v-if="st.is_visited_in_period && st.last_visit_time">
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 text-[10px] font-bold bg-indigo-50 text-indigo-700 rounded-md">
                    <CheckCircle2 class="w-3 h-3 text-indigo-600" /> Dikunjungi
                  </span>
                  <p class="text-[10px] font-medium text-slate-500 mt-0.5">{{ formatTime(st.last_visit_time) }}</p>
                </div>
                <span v-else class="text-[10px] font-semibold text-slate-400">Belum Dikunjungi</span>
              </td>

              <td class="p-5 text-center whitespace-nowrap">
                <div v-if="st.is_effective_call">
                  <span class="inline-flex items-center gap-1 px-2.5 py-1 text-[10px] font-black bg-emerald-100 text-emerald-800 rounded-full uppercase tracking-wider border border-emerald-200">
                    <ShoppingBag class="w-3 h-3 text-emerald-700" /> EC: ADA ORDER
                  </span>
                </div>
                <span v-else class="inline-flex items-center px-2.5 py-1 text-[10px] font-bold bg-slate-100 text-slate-500 rounded-full uppercase">
                  Non-EC
                </span>
              </td>

              <td class="p-5 text-right whitespace-nowrap">
                <div v-if="st.order_amount > 0">
                  <p class="text-sm font-black text-emerald-700">{{ formatCurrency(st.order_amount) }}</p>
                  <p class="text-[10px] font-bold text-slate-400">{{ st.order_count }} Order • {{ st.latest_order_number }}</p>
                </div>
                <span v-else class="text-xs font-bold text-slate-400">-</span>
              </td>

              <td class="p-5 max-w-xs">
                <p class="text-xs font-medium text-slate-700 line-clamp-1">{{ st.address }}</p>
                <p v-if="st.city" class="text-[10px] font-bold text-primary mt-0.5">{{ st.city }}</p>
              </td>

              <td class="p-5 text-right whitespace-nowrap">
                <div class="flex items-center justify-end gap-1">
                  <button
                    @click="openWhatsApp(st.phone_number, st.store_name, st.salesman_name)"
                    class="p-2 hover:bg-emerald-50 text-emerald-600 rounded-xl transition-all cursor-pointer"
                    title="Kirim WhatsApp"
                  >
                    <MessageSquare class="w-4 h-4" />
                  </button>
                  <a
                    v-if="st.latitude && st.longitude"
                    :href="`https://www.google.com/maps/search/?api=1&query=${st.latitude},${st.longitude}`"
                    target="_blank"
                    class="p-2 hover:bg-blue-50 text-blue-600 rounded-xl transition-all inline-flex"
                    title="Buka Peta"
                  >
                    <MapPin class="w-4 h-4" />
                  </a>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
