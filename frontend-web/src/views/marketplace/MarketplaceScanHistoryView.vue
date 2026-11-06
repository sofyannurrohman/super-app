<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-800 tracking-tight">History Scan Picking &amp; Packing</h1>
        <p class="text-slate-500 mt-1">Lacak dan lihat data historis scan pesanan</p>
      </div>
      
      <!-- Filters -->
      <div class="flex flex-wrap items-center gap-3 w-full sm:w-auto">
        <!-- Action Filter -->
        <select 
          v-model="actionFilter"
          @change="fetchHistory"
          class="w-full sm:w-auto px-4 py-2 bg-white border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 text-sm font-medium text-slate-700 shadow-sm"
        >
          <option value="ALL">Semua Aktivitas</option>
          <option value="PICK">Hanya Picking</option>
          <option value="PACK">Hanya Packing</option>
        </select>

        <!-- Warehouse Filter -->
        <select
          v-model="warehouseFilter"
          @change="fetchHistory"
          class="w-full sm:w-auto px-4 py-2 bg-white border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 text-sm font-medium text-slate-700 shadow-sm"
        >
          <option value="ALL">Semua Gudang</option>
          <option v-for="wh in availableWarehouses" :key="wh" :value="wh">{{ wh }}</option>
        </select>
        
        <!-- Filter Mode -->
        <select 
          v-model="filterMode"
          @change="applyFilters"
          class="w-full sm:w-auto px-4 py-2 bg-white border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 text-sm font-medium text-slate-700 shadow-sm"
        >
          <option value="day">Harian</option>
          <option value="week">Mingguan</option>
          <option value="month">Bulanan</option>
        </select>

        <!-- Dynamic Inputs -->
        <input 
          v-if="filterMode === 'day'"
          type="date"
          v-model="selectedDate"
          @change="applyFilters"
          class="w-full sm:w-auto px-4 py-2 bg-white border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 text-sm font-medium text-slate-700 shadow-sm"
        />
        <input 
          v-if="filterMode === 'week'"
          type="week"
          v-model="selectedWeek"
          @change="applyFilters"
          class="w-full sm:w-auto px-4 py-2 bg-white border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 text-sm font-medium text-slate-700 shadow-sm"
        />
        <input 
          v-if="filterMode === 'month'"
          type="month"
          v-model="selectedMonth"
          @change="applyFilters"
          class="w-full sm:w-auto px-4 py-2 bg-white border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 text-sm font-medium text-slate-700 shadow-sm"
        />
      </div>
    </div>

    <!-- Summary Cards -->
    <div v-if="!loading && (Object.keys(warehouseSummary).length > 0 || grandTotal > 0)" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <!-- Card per Warehouse (dinamis) -->
      <div
        v-for="(stats, warehouseName) in warehouseSummary"
        :key="warehouseName"
        class="bg-white rounded-2xl border border-slate-200 shadow-sm p-5 transition-all hover:shadow-md hover:-translate-y-0.5"
      >
        <div class="flex items-start justify-between mb-3">
          <div class="flex items-center gap-2 min-w-0">
            <div class="w-8 h-8 rounded-lg bg-indigo-50 flex items-center justify-center flex-shrink-0">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-indigo-500">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
              </svg>
            </div>
            <h3 class="font-semibold text-slate-700 text-sm leading-tight truncate">{{ warehouseName }}</h3>
          </div>
          <span class="text-2xl font-bold text-indigo-600 flex-shrink-0 ml-2">{{ stats.total }}</span>
        </div>
        <div class="flex flex-wrap gap-2 mt-1">
          <span class="inline-flex items-center gap-1 text-xs px-2.5 py-1 bg-blue-50 text-blue-700 rounded-full font-semibold">
            <span class="w-1.5 h-1.5 rounded-full bg-blue-400 inline-block"></span>
            Picking: {{ stats.pick }}
          </span>
          <span class="inline-flex items-center gap-1 text-xs px-2.5 py-1 bg-emerald-50 text-emerald-700 rounded-full font-semibold">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 inline-block"></span>
            Packing: {{ stats.pack }}
          </span>
        </div>
      </div>

      <!-- Card Total Keseluruhan -->
      <div class="bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl shadow-md p-5 text-white transition-all hover:shadow-lg hover:-translate-y-0.5">
        <div class="flex items-start justify-between mb-3">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-white/20 flex items-center justify-center flex-shrink-0">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-white">
                <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
              </svg>
            </div>
            <h3 class="font-semibold text-sm text-white/90">Total Keseluruhan</h3>
          </div>
          <span class="text-2xl font-bold text-white">{{ grandTotal }}</span>
        </div>
        <div class="flex flex-wrap gap-2 mt-1">
          <span class="inline-flex items-center gap-1 text-xs px-2.5 py-1 bg-white/20 rounded-full font-semibold text-white">
            <span class="w-1.5 h-1.5 rounded-full bg-white/60 inline-block"></span>
            Picking: {{ grandTotalPick }}
          </span>
          <span class="inline-flex items-center gap-1 text-xs px-2.5 py-1 bg-white/20 rounded-full font-semibold text-white">
            <span class="w-1.5 h-1.5 rounded-full bg-white/60 inline-block"></span>
            Packing: {{ grandTotalPack }}
          </span>
        </div>
      </div>
    </div>

    <!-- Skeleton cards saat loading -->
    <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="i in 3" :key="i" class="bg-white rounded-2xl border border-slate-200 shadow-sm p-5 animate-pulse">
        <div class="flex items-center justify-between mb-3">
          <div class="h-4 bg-slate-100 rounded w-28"></div>
          <div class="h-7 bg-slate-100 rounded w-10"></div>
        </div>
        <div class="flex gap-2">
          <div class="h-6 bg-slate-100 rounded-full w-20"></div>
          <div class="h-6 bg-slate-100 rounded-full w-20"></div>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex border-b border-slate-200 mb-4">
      <button 
        @click="activeTab = 'log'" 
        :class="activeTab === 'log' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'"
        class="px-4 py-2 border-b-2 font-medium text-sm transition-colors"
      >
        History Log
      </button>
      <button 
        @click="activeTab = 'recap'" 
        :class="activeTab === 'recap' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'"
        class="px-4 py-2 border-b-2 font-medium text-sm transition-colors"
      >
        Rekap Pengeluaran Stok
      </button>
    </div>

    <!-- Data Table -->
    <div v-if="activeTab === 'log'" class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-slate-500 uppercase bg-slate-50/50 border-b border-slate-100">
            <tr>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">Waktu Scan</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">Order ID</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">Aksi</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">Warehouse</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">Produk</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">SKU</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">Discan Oleh</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading" class="animate-pulse">
              <td colspan="7" class="px-6 py-8 text-center text-slate-400">
                Memuat data history...
              </td>
            </tr>
            <tr v-else-if="histories.length === 0">
              <td colspan="7" class="px-6 py-12 text-center">
                <div class="flex flex-col items-center justify-center text-slate-400">
                  <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-history mb-3 opacity-50"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M12 7v5l4 2"/></svg>
                  <p class="font-medium">Belum ada history scan untuk periode ini.</p>
                </div>
              </td>
            </tr>
            <tr 
              v-else 
              v-for="item in paginatedHistories" 
              :key="item.id"
              class="hover:bg-slate-50/50 transition-colors"
            >
              <td class="px-6 py-4 whitespace-nowrap text-slate-600">
                {{ formatDateTime(item.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap font-medium text-slate-900">
                {{ item.order_id }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span 
                  class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold"
                  :class="item.action === 'PICK' ? 'bg-blue-50 text-blue-700' : 'bg-emerald-50 text-emerald-700'"
                >
                  {{ item.action === 'PICK' ? 'PICKING' : 'PACKING' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span v-if="item.warehouse_name" class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-50 text-slate-700 text-xs font-medium border border-slate-100">
                  <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="text-slate-400">
                    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
                  </svg>
                  {{ item.warehouse_name }}
                </span>
                <span v-else class="text-slate-400 text-xs">-</span>
              </td>
              <td class="px-6 py-4">
                <p class="text-slate-800 line-clamp-2 min-w-[200px]">{{ item.product_name }}</p>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-slate-600">
                {{ item.seller_sku }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center gap-2">
                  <div class="w-6 h-6 rounded-full bg-slate-100 flex items-center justify-center text-xs font-bold text-slate-500">
                    {{ item.actor ? item.actor.substring(0, 2).toUpperCase() : 'AD' }}
                  </div>
                  <span class="font-medium text-slate-700">{{ item.actor }}</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="px-6 py-4 border-t border-slate-100 flex items-center justify-between">
        <span class="text-sm text-slate-500">
          Menampilkan <span class="font-medium text-slate-900">{{ ((currentPage - 1) * itemsPerPage) + 1 }}</span>
          sampai <span class="font-medium text-slate-900">{{ Math.min(currentPage * itemsPerPage, histories.length) }}</span>
          dari <span class="font-medium text-slate-900">{{ histories.length }}</span> data
        </span>
        <div class="flex gap-2">
          <button 
            @click="prevPage" 
            :disabled="currentPage === 1"
            class="px-3 py-1 text-sm font-medium border border-slate-200 rounded-lg hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          >
            Sblmnya
          </button>
          <div class="flex items-center px-3 text-sm font-medium text-slate-700">
            Hal {{ currentPage }} dari {{ totalPages }}
          </div>
          <button 
            @click="nextPage" 
            :disabled="currentPage === totalPages"
            class="px-3 py-1 text-sm font-medium border border-slate-200 rounded-lg hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          >
            Lanjut
          </button>
        </div>
      </div>
    </div>

    <!-- Recap Table -->
    <div v-else-if="activeTab === 'recap'" class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-slate-500 uppercase bg-slate-50/50 border-b border-slate-100">
            <tr>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">No</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">SKU</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">Produk</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap">Warehouse</th>
              <th class="px-6 py-4 font-semibold whitespace-nowrap text-right">Jumlah Berkurang (PACK)</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading" class="animate-pulse">
              <td colspan="5" class="px-6 py-8 text-center text-slate-400">
                Memuat data rekap...
              </td>
            </tr>
            <tr v-else-if="stockReductionRecap.length === 0">
              <td colspan="5" class="px-6 py-12 text-center">
                <div class="flex flex-col items-center justify-center text-slate-400">
                  <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-history mb-3 opacity-50"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M12 7v5l4 2"/></svg>
                  <p class="font-medium">Belum ada data rekap pengeluaran stok untuk periode ini.</p>
                </div>
              </td>
            </tr>
            <tr 
              v-else 
              v-for="(item, index) in stockReductionRecap" 
              :key="item.sku + '_' + item.warehouse"
              class="hover:bg-slate-50/50 transition-colors"
            >
              <td class="px-6 py-4 whitespace-nowrap text-slate-600">
                {{ index + 1 }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap font-medium text-slate-900">
                {{ item.sku }}
              </td>
              <td class="px-6 py-4">
                <p class="text-slate-800 line-clamp-2 min-w-[200px]">{{ item.name }}</p>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span v-if="item.warehouse !== '-'" class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-50 text-slate-700 text-xs font-medium border border-slate-100">
                  <svg xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="text-slate-400">
                    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
                  </svg>
                  {{ item.warehouse }}
                </span>
                <span v-else class="text-slate-400 text-xs">-</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right">
                <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold bg-rose-50 text-rose-700 border border-rose-100">
                  -{{ item.count }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

interface ScanHistory {
  id: number
  order_id: string
  action: string
  actor: string
  seller_sku: string
  product_name: string
  warehouse_name: string
  created_at: string
}

const histories = ref<ScanHistory[]>([])
const loading = ref(true)
const activeTab = ref<'log' | 'recap'>('log')

// Filters
const actionFilter = ref('ALL')
const warehouseFilter = ref('ALL')
const filterMode = ref('day')
const availableWarehouses = ref<string[]>([])

const getTodayDate = () => new Date().toISOString().split('T')[0]
const getThisMonth = () => {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}
const getThisWeek = () => {
  const d = new Date()
  // Adjust to Thursday to find the correct week year
  d.setDate(d.getDate() + 4 - (d.getDay() || 7))
  const yearStart = new Date(d.getFullYear(), 0, 1)
  const weekNo = Math.ceil((((d.getTime() - yearStart.getTime()) / 86400000) + 1) / 7)
  return `${d.getFullYear()}-W${String(weekNo).padStart(2, '0')}`
}

const selectedDate = ref(getTodayDate())
const selectedWeek = ref(getThisWeek())
const selectedMonth = ref(getThisMonth())

const dateFrom = ref('')
const dateTo = ref('')

// Pagination
const currentPage = ref(1)
const itemsPerPage = 20

const totalPages = computed(() => Math.ceil(histories.value.length / itemsPerPage))

const paginatedHistories = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage
  return histories.value.slice(start, end)
})

const prevPage = () => {
  if (currentPage.value > 1) currentPage.value--
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) currentPage.value++
}

// Summary cards — reactive terhadap semua filter aktif
const warehouseSummary = computed(() => {
  const summary: Record<string, { pick: number; pack: number; total: number }> = {}

  histories.value.forEach(item => {
    const wh = item.warehouse_name?.trim() || 'Tidak Diketahui'
    if (!summary[wh]) summary[wh] = { pick: 0, pack: 0, total: 0 }
    if (item.action === 'PICK') summary[wh].pick++
    else if (item.action === 'PACK') summary[wh].pack++
    summary[wh].total++
  })

  return summary
})

const grandTotal = computed(() => histories.value.length)

const grandTotalPick = computed(() =>
  histories.value.filter(h => h.action === 'PICK').length
)

const grandTotalPack = computed(() =>
  histories.value.filter(h => h.action === 'PACK').length
)

const formatDateTime = (dateString: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

const stockReductionRecap = computed(() => {
  const recapMap = new Map<string, { sku: string, name: string, warehouse: string, count: number }>()

  histories.value.forEach(item => {
    if (item.action === 'PACK') {
      const key = `${item.seller_sku}_${item.warehouse_name}`
      if (recapMap.has(key)) {
        recapMap.get(key)!.count++
      } else {
        recapMap.set(key, {
          sku: item.seller_sku,
          name: item.product_name,
          warehouse: item.warehouse_name || '-',
          count: 1
        })
      }
    }
  })

  // Convert to array and sort by count descending
  return Array.from(recapMap.values()).sort((a, b) => b.count - a.count)
})

const applyFilters = () => {
  let from = ''
  let to = ''

  if (filterMode.value === 'day' && selectedDate.value) {
    from = selectedDate.value
    to = selectedDate.value
  } else if (filterMode.value === 'month' && selectedMonth.value) {
    const [year, month] = selectedMonth.value.split('-')
    from = `${year}-${month}-01`
    const lastDay = new Date(Number(year), Number(month), 0).getDate()
    to = `${year}-${month}-${String(lastDay).padStart(2, '0')}`
  } else if (filterMode.value === 'week' && selectedWeek.value) {
    const [yearStr, weekStr] = selectedWeek.value.split('-W')
    const year = Number(yearStr)
    const week = Number(weekStr)
    const jan4 = new Date(year, 0, 4)
    const day = jan4.getDay() || 7
    const week1Start = new Date(year, 0, 4 - day + 1)
    const start = new Date(week1Start.getTime() + (week - 1) * 7 * 24 * 60 * 60 * 1000)
    const end = new Date(start.getTime() + 6 * 24 * 60 * 60 * 1000)
    
    from = start.toISOString().split('T')[0] || ''
    to = end.toISOString().split('T')[0] || ''
  }

  dateFrom.value = from
  dateTo.value = to
  
  fetchHistory()
}

const fetchWarehouses = async () => {
  try {
    const bid = masterDataStore.selectedBranchId
    const branchFilter = bid && bid !== 'ALL_BRANCHES' ? `?branch_id=${bid}` : ''
    const response = await apiClient.get(`/marketplace/scan-history/warehouses${branchFilter}`)
    availableWarehouses.value = response.data.data || []
  } catch (error) {
    console.error('Error fetching warehouses:', error)
  }
}

const fetchHistory = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    const bid = masterDataStore.selectedBranchId
    if (bid && bid !== 'ALL_BRANCHES') params.append('branch_id', bid)
    if (dateFrom.value) params.append('from', dateFrom.value)
    if (dateTo.value) params.append('to', dateTo.value)
    if (actionFilter.value !== 'ALL') params.append('action', actionFilter.value)
    if (warehouseFilter.value !== 'ALL') params.append('warehouse', warehouseFilter.value)
    
    const response = await apiClient.get(`/marketplace/scan-history?${params.toString()}`)
    histories.value = response.data.data || []
    currentPage.value = 1
  } catch (error: any) {
    console.error('Error fetching scan history:', error)
    toast.error('Gagal memuat history scan')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchWarehouses()
  applyFilters()
})

watch(() => masterDataStore.selectedBranchId, () => {
  warehouseFilter.value = 'ALL'
  fetchWarehouses()
  applyFilters()
})
</script>
