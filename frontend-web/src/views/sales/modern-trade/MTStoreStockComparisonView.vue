<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Boxes,
  Search,
  Plus,
  RefreshCw,
  AlertTriangle,
  CheckCircle2,
  TrendingDown,
  Calendar,
  Filter,
  ArrowUpDown,
  History,
  Store,
  UserCheck
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useRouter } from 'vue-router'

const router = useRouter()

// State
const loading = ref(false)
const storesSummary = ref<any[]>([])
const storeStocks = ref<any[]>([])
const stockLogs = ref<any[]>([])
const activeTab = ref<'comparison' | 'stocks' | 'logs'>('comparison')

// Search & Filter
const searchQuery = ref('')
const selectedStatusFilter = ref<string>('ALL')
const selectedStoreId = ref<string>('')

// Summary Stats
const totalOverallPcs = ref(0)
const totalOosStores = ref(0)
const totalLowStockStores = ref(0)
const totalSpgRecommendedStores = ref(0)

// Adjust Modal State
const showAdjustModal = ref(false)
const adjustForm = ref({
  store_id: '',
  product_id: '',
  quantity: 0,
  min_stock_threshold: 20,
  notes: '',
  adjustment_type: 'ADJUSTMENT',
  reference_no: ''
})

// Master data
const storesList = ref<any[]>([])
const productsList = ref<any[]>([])

const fetchComparisonData = async () => {
  loading.value = true
  try {
    const res = await apiClient.get('/modern-trade/store-stock-comparison')
    if (res.data?.data) {
      storesSummary.value = res.data.data.stores || []
      totalOverallPcs.value = res.data.data.total_overall_pcs || 0
      totalOosStores.value = res.data.data.total_oos_stores || 0
      totalLowStockStores.value = res.data.data.total_low_stock_stores || 0
      totalSpgRecommendedStores.value = res.data.data.total_spg_recommended_stores || 0
    }
  } catch (err: any) {
    toast.error('Gagal memuat ringkasan stok toko MT: ' + (err.response?.data?.error || err.message))
  } finally {
    loading.value = false
  }
}

const fetchStoreStocks = async () => {
  try {
    const params = selectedStoreId.value ? { store_id: selectedStoreId.value } : {}
    const res = await apiClient.get('/modern-trade/store-stocks', { params })
    storeStocks.value = res.data?.data || []
  } catch (err: any) {
    console.error(err)
  }
}

const fetchStockLogs = async () => {
  try {
    const params = selectedStoreId.value ? { store_id: selectedStoreId.value } : {}
    const res = await apiClient.get('/modern-trade/store-stock-logs', { params })
    stockLogs.value = res.data?.data || []
  } catch (err: any) {
    console.error(err)
  }
}

const fetchMasters = async () => {
  try {
    const [storesRes, productsRes] = await Promise.all([
      apiClient.get('/stores?sales_type=sales_mt&limit=200'),
      apiClient.get('/factory/products?limit=100')
    ])
    storesList.value = storesRes.data?.data || []
    productsList.value = productsRes.data?.data || []
  } catch (err) {
    console.error(err)
  }
}

const refreshAll = () => {
  fetchComparisonData()
  fetchStoreStocks()
  fetchStockLogs()
}

onMounted(() => {
  refreshAll()
  fetchMasters()
})

const filteredStores = computed(() => {
  let list = storesSummary.value
  if (selectedStatusFilter.value !== 'ALL') {
    list = list.filter(s => s.stock_status === selectedStatusFilter.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(s =>
      (s.store_name || '').toLowerCase().includes(q) ||
      (s.store_address || '').toLowerCase().includes(q) ||
      (s.owner_name || '').toLowerCase().includes(q)
    )
  }
  return list
})

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'OUT_OF_STOCK':
      return { label: 'Habis (OOS)', class: 'bg-rose-50 text-rose-600 border border-rose-200' }
    case 'LOW_STOCK':
      return { label: 'Stok Tipis', class: 'bg-amber-50 text-amber-600 border border-amber-200' }
    case 'OVERSTOCKED':
      return { label: 'Stok Melimpah', class: 'bg-emerald-50 text-emerald-600 border border-emerald-200' }
    default:
      return { label: 'Aman', class: 'bg-blue-50 text-blue-600 border border-blue-200' }
  }
}

const openAdjustModal = (store?: any) => {
  adjustForm.value = {
    store_id: store ? store.store_id : (storesList.value[0]?.id || ''),
    product_id: productsList.value[0]?.id || '',
    quantity: 0,
    min_stock_threshold: 20,
    notes: '',
    adjustment_type: 'ADJUSTMENT',
    reference_no: 'OPNAME-' + new Date().toISOString().slice(0, 10)
  }
  showAdjustModal.value = true
}

const submitAdjustStock = async () => {
  if (!adjustForm.value.store_id || !adjustForm.value.product_id) {
    toast.error('Pilih Toko dan Produk terlebih dahulu')
    return
  }

  try {
    await apiClient.post('/modern-trade/store-stocks/adjust', adjustForm.value)
    toast.success('Stok toko berhasil disesuaikan!')
    showAdjustModal.value = false
    refreshAll()
  } catch (err: any) {
    toast.error('Gagal menyesuaikan stok: ' + (err.response?.data?.error || err.message))
  }
}

const navigateToScheduleSpg = (storeName: string) => {
  router.push({
    path: '/sales/modern-trade/spg',
    query: { tab: 'schedules', store_name: storeName }
  })
}

const formatNumber = (num: number) => {
  return new Intl.NumberFormat('id-ID').format(num || 0)
}
</script>

<template>
  <div class="p-4 sm:p-8 max-w-7xl mx-auto space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-white p-6 rounded-3xl border border-slate-100 shadow-sm">
      <div class="space-y-1">
        <div class="flex items-center gap-2">
          <Boxes class="w-6 h-6 text-red-600" />
          <h1 class="text-xl sm:text-2xl font-black text-slate-900 tracking-tight">
            Perbandingan Stok & Konsinyasi Toko MT
          </h1>
        </div>
        <p class="text-xs text-slate-500 font-medium">
          Monitoring perbandingan stok antar gerai MT (satuan Pcs) untuk prioritas penjadwalan SPG & mitigasi Out of Stock
        </p>
      </div>

      <div class="flex items-center gap-3 w-full sm:w-auto">
        <Button
          @click="refreshAll"
          variant="outline"
          class="rounded-xl border-slate-200 text-xs font-bold gap-2 text-slate-600 hover:bg-slate-50"
        >
          <RefreshCw class="w-3.5 h-3.5" :class="loading ? 'animate-spin' : ''" />
          Refresh
        </Button>
        <Button
          @click="openAdjustModal()"
          class="rounded-xl bg-red-600 hover:bg-red-700 text-xs font-bold gap-2 text-white shadow-lg shadow-red-500/20"
        >
          <Plus class="w-3.5 h-3.5" />
          Opname / Adjust Stok
        </Button>
      </div>
    </div>

    <!-- Metric KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Total Stock In Field -->
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-red-50 text-red-600 rounded-xl">
          <Boxes class="w-6 h-6" />
        </div>
        <div>
          <p class="text-[10px] uppercase font-bold text-slate-400">Total Stok di Toko MT</p>
          <h3 class="text-xl font-black text-slate-900">{{ formatNumber(totalOverallPcs) }} <span class="text-xs font-normal text-slate-500">Pcs</span></h3>
        </div>
      </div>

      <!-- SPG Booster Recommended -->
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
          <UserCheck class="w-6 h-6" />
        </div>
        <div>
          <p class="text-[10px] uppercase font-bold text-slate-400">Rekomendasi Jadwal SPG</p>
          <h3 class="text-xl font-black text-emerald-600">{{ totalSpgRecommendedStores }} <span class="text-xs font-normal text-slate-500">Toko (Stok > 100)</span></h3>
        </div>
      </div>

      <!-- Low Stock Alert -->
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-amber-50 text-amber-600 rounded-xl">
          <TrendingDown class="w-6 h-6" />
        </div>
        <div>
          <p class="text-[10px] uppercase font-bold text-slate-400">Stok Menipis (Reorder)</p>
          <h3 class="text-xl font-black text-amber-600">{{ totalLowStockStores }} <span class="text-xs font-normal text-slate-500">Toko</span></h3>
        </div>
      </div>

      <!-- Out of Stock Alert -->
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-rose-50 text-rose-600 rounded-xl">
          <AlertTriangle class="w-6 h-6" />
        </div>
        <div>
          <p class="text-[10px] uppercase font-bold text-slate-400">Habis Total (OOS Urgent)</p>
          <h3 class="text-xl font-black text-rose-600">{{ totalOosStores }} <span class="text-xs font-normal text-slate-500">Toko</span></h3>
        </div>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex items-center gap-2 border-b border-slate-200">
      <button
        @click="activeTab = 'comparison'"
        class="pb-3 px-4 text-xs font-extrabold transition-all border-b-2"
        :class="activeTab === 'comparison' ? 'border-red-600 text-red-600' : 'border-transparent text-slate-400 hover:text-slate-700'"
      >
        Perbandingan Antar Gerai
      </button>
      <button
        @click="activeTab = 'stocks'; fetchStoreStocks()"
        class="pb-3 px-4 text-xs font-extrabold transition-all border-b-2"
        :class="activeTab === 'stocks' ? 'border-red-600 text-red-600' : 'border-transparent text-slate-400 hover:text-slate-700'"
      >
        Rincian Stok SKU
      </button>
      <button
        @click="activeTab = 'logs'; fetchStockLogs()"
        class="pb-3 px-4 text-xs font-extrabold transition-all border-b-2"
        :class="activeTab === 'logs' ? 'border-red-600 text-red-600' : 'border-transparent text-slate-400 hover:text-slate-700'"
      >
        Log Mutasi Stok
      </button>
    </div>

    <!-- TAB 1: Comparison Hub -->
    <div v-if="activeTab === 'comparison'" class="space-y-4">
      <!-- Search & Filters -->
      <div class="flex flex-col sm:flex-row justify-between gap-4">
        <div class="relative w-full sm:w-[320px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari Toko / Supermarket MT..."
            class="w-full pl-9 pr-4 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-red-500 bg-white"
          />
        </div>

        <div class="flex items-center gap-2">
          <select
            v-model="selectedStatusFilter"
            class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-red-500 bg-white font-medium text-slate-700"
          >
            <option value="ALL">Semua Level Stok</option>
            <option value="OVERSTOCKED">Stok Melimpah (Push SPG)</option>
            <option value="SUFFICIENT">Stok Aman</option>
            <option value="LOW_STOCK">Stok Tipis (Perlu PO)</option>
            <option value="OUT_OF_STOCK">Habis Total (OOS)</option>
          </select>
        </div>
      </div>

      <!-- Comparison Table -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full border-collapse text-left">
            <thead>
              <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
                <th class="py-4 px-6">Nama Gerai MT</th>
                <th class="py-4 px-6">Alamat & Pemilik</th>
                <th class="py-4 px-6 text-center">Variasi SKU</th>
                <th class="py-4 px-6 text-right">Saldo Stok (Pcs)</th>
                <th class="py-4 px-6 text-center">Status Stok</th>
                <th class="py-4 px-6 text-center">Rekomendasi Aksi</th>
                <th class="py-4 px-6 text-center">Aksi Cepat</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50 text-slate-700 text-xs">
              <tr v-for="s in filteredStores" :key="s.store_id" class="hover:bg-slate-50/40 transition-all">
                <td class="py-4 px-6">
                  <div class="flex items-center gap-2">
                    <Store class="w-4 h-4 text-red-500 shrink-0" />
                    <div>
                      <span class="font-extrabold text-slate-900 block text-sm">{{ s.store_name }}</span>
                      <span class="text-[10px] text-slate-400 font-medium" v-if="s.phone_number">{{ s.phone_number }}</span>
                    </div>
                  </div>
                </td>
                <td class="py-4 px-6 text-slate-500">
                  <span class="block font-medium truncate max-w-[200px]" :title="s.store_address">{{ s.store_address || '-' }}</span>
                  <span class="text-[10px] text-slate-400" v-if="s.owner_name">PIC: {{ s.owner_name }}</span>
                </td>
                <td class="py-4 px-6 text-center font-bold text-slate-700">
                  {{ s.total_skus }} SKU
                </td>
                <td class="py-4 px-6 text-right">
                  <span class="text-sm font-black" :class="s.total_stock_pcs === 0 ? 'text-rose-600' : s.total_stock_pcs <= s.min_stock_threshold ? 'text-amber-600' : 'text-slate-900'">
                    {{ formatNumber(s.total_stock_pcs) }}
                  </span>
                  <span class="text-[10px] text-slate-400 ml-1">Pcs</span>
                </td>
                <td class="py-4 px-6 text-center">
                  <span
                    class="px-2.5 py-1 rounded-lg font-bold text-[10px] tracking-wide inline-block"
                    :class="getStatusBadge(s.stock_status).class"
                  >
                    {{ getStatusBadge(s.stock_status).label }}
                  </span>
                </td>
                <td class="py-4 px-6 text-center">
                  <span
                    v-if="s.recommendation === 'SPG_BOOSTER'"
                    class="bg-emerald-50 text-emerald-700 font-bold px-2 py-1 rounded-md text-[10px] border border-emerald-200 inline-flex items-center gap-1"
                  >
                    <UserCheck class="w-3 h-3 text-emerald-600" />
                    Jadwalkan SPG (Weekend)
                  </span>
                  <span
                    v-else-if="s.recommendation === 'REPEAT_ORDER_URGENT' || s.recommendation === 'REPEAT_ORDER'"
                    class="bg-rose-50 text-rose-700 font-bold px-2 py-1 rounded-md text-[10px] border border-rose-200 inline-flex items-center gap-1"
                  >
                    <AlertTriangle class="w-3 h-3 text-rose-600" />
                    Buat Repeat Order PO
                  </span>
                  <span
                    v-else
                    class="text-slate-400 font-medium text-[10px]"
                  >
                    Stok Stabil
                  </span>
                </td>
                <td class="py-4 px-6 text-center">
                  <div class="flex items-center justify-center gap-1.5">
                    <Button
                      v-if="s.stock_status === 'OVERSTOCKED' || s.total_stock_pcs > 80"
                      @click="navigateToScheduleSpg(s.store_name)"
                      size="sm"
                      variant="outline"
                      class="h-7 text-[10px] font-bold text-emerald-600 border-emerald-200 hover:bg-emerald-50 rounded-lg"
                    >
                      Assign SPG
                    </Button>
                    <Button
                      @click="openAdjustModal(s)"
                      size="sm"
                      variant="outline"
                      class="h-7 text-[10px] font-bold text-slate-600 border-slate-200 hover:bg-slate-50 rounded-lg"
                    >
                      Adjust Stok
                    </Button>
                  </div>
                </td>
              </tr>
              <tr v-if="filteredStores.length === 0">
                <td colspan="7" class="py-8 text-center text-slate-400">
                  Tidak ada gerai MT ditemukan.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: SKU Stocks List -->
    <div v-else-if="activeTab === 'stocks'" class="space-y-4">
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full border-collapse text-left">
            <thead>
              <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
                <th class="py-4 px-6">Gerai Toko</th>
                <th class="py-4 px-6">Produk / SKU</th>
                <th class="py-4 px-6 text-right">Saldo Stok (Pcs)</th>
                <th class="py-4 px-6 text-right">Min. Threshold</th>
                <th class="py-4 px-6 text-center">Terakhir Diupdate</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50 text-slate-700 text-xs">
              <tr v-for="stk in storeStocks" :key="stk.id" class="hover:bg-slate-50/40">
                <td class="py-4 px-6 font-bold text-slate-900">{{ stk.store?.name || '-' }}</td>
                <td class="py-4 px-6 font-semibold text-slate-750">{{ stk.product?.name || '-' }} ({{ stk.product?.sku || '-' }})</td>
                <td class="py-4 px-6 text-right font-black text-slate-900">{{ formatNumber(stk.quantity) }} Pcs</td>
                <td class="py-4 px-6 text-right text-slate-500">{{ stk.min_stock_threshold }} Pcs</td>
                <td class="py-4 px-6 text-center text-slate-400 text-[11px]">{{ stk.last_adjusted_at ? stk.last_adjusted_at.replace('T', ' ').slice(0, 16) : '-' }}</td>
              </tr>
              <tr v-if="storeStocks.length === 0">
                <td colspan="5" class="py-8 text-center text-slate-400">Belum ada data saldo stok produk di toko MT.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 3: Mutation Logs -->
    <div v-else-if="activeTab === 'logs'" class="space-y-4">
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full border-collapse text-left">
            <thead>
              <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
                <th class="py-4 px-6">Waktu</th>
                <th class="py-4 px-6">Toko</th>
                <th class="py-4 px-6">Produk</th>
                <th class="py-4 px-6">Tipe Mutasi</th>
                <th class="py-4 px-6 text-right">Perubahan Qty</th>
                <th class="py-4 px-6 text-right">Saldo Akhir</th>
                <th class="py-4 px-6">Keterangan / Ref</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50 text-slate-700 text-xs">
              <tr v-for="log in stockLogs" :key="log.id" class="hover:bg-slate-50/40">
                <td class="py-4 px-6 text-slate-400 text-[11px]">{{ log.created_at ? log.created_at.replace('T', ' ').slice(0, 16) : '-' }}</td>
                <td class="py-4 px-6 font-bold text-slate-900">{{ log.store?.name || '-' }}</td>
                <td class="py-4 px-6 font-semibold text-slate-750">{{ log.product?.name || '-' }}</td>
                <td class="py-4 px-6">
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-bold"
                    :class="log.type === 'DELIVERY_IN' ? 'bg-emerald-50 text-emerald-600' : log.type === 'SPG_SALE_OUT' ? 'bg-rose-50 text-rose-600' : 'bg-blue-50 text-blue-600'"
                  >
                    {{ log.type }}
                  </span>
                </td>
                <td class="py-4 px-6 text-right font-black" :class="log.quantity_after >= log.quantity_before ? 'text-emerald-600' : 'text-rose-600'">
                  {{ log.quantity_after >= log.quantity_before ? '+' : '' }}{{ log.quantity_after - log.quantity_before }} Pcs
                </td>
                <td class="py-4 px-6 text-right font-bold text-slate-900">{{ log.quantity_after }} Pcs</td>
                <td class="py-4 px-6 text-slate-500 text-[11px]">
                  <span>{{ log.reference_no || '-' }}</span>
                  <span class="block text-[10px] text-slate-400" v-if="log.notes">{{ log.notes }}</span>
                </td>
              </tr>
              <tr v-if="stockLogs.length === 0">
                <td colspan="7" class="py-8 text-center text-slate-400">Belum ada riwayat mutasi stok toko.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Adjust Stock Modal -->
    <div
      v-if="showAdjustModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
    >
      <div class="bg-white rounded-[24px] w-full max-w-md shadow-2xl p-6 relative">
        <h3 class="text-base font-black text-slate-900 mb-4">Stock Opname / Penyesuaian Sisa Stok</h3>

        <div class="space-y-4 text-xs font-semibold text-slate-650">
          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Pilih Toko MT</label>
            <select
              v-model="adjustForm.store_id"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 bg-white"
            >
              <option v-for="st in storesList" :key="st.id" :value="st.id">{{ st.name }}</option>
            </select>
          </div>

          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Pilih Produk</label>
            <select
              v-model="adjustForm.product_id"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 bg-white"
            >
              <option v-for="p in productsList" :key="p.id" :value="p.id">{{ p.name }} ({{ p.sku }})</option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Sisa Stok Fisik (Pcs)</label>
              <input
                v-model.number="adjustForm.quantity"
                type="number"
                min="0"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Min Threshold Alert (Pcs)</label>
              <input
                v-model.number="adjustForm.min_stock_threshold"
                type="number"
                min="1"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500"
              />
            </div>
          </div>

          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Catatan / Alasan Adjustment</label>
            <textarea
              v-model="adjustForm.notes"
              rows="2"
              placeholder="Misal: Hasil Stock Opname Fisik Kunjungan Sales MT"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500"
            ></textarea>
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 mt-6 pt-4 border-t border-slate-100">
          <Button
            @click="showAdjustModal = false"
            variant="outline"
            class="rounded-xl border-slate-200 text-xs font-bold text-slate-600"
          >
            Batal
          </Button>
          <Button
            @click="submitAdjustStock"
            class="rounded-xl bg-red-600 hover:bg-red-700 text-xs font-bold text-white"
          >
            Simpan Penyesuaian
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
