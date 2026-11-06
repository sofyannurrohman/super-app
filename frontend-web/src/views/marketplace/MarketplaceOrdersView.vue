<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import {
  ShoppingCart,
  Store,
  RefreshCw,
  Search,
  Filter,
  Calendar,
  Package,
  Truck,
  CheckCircle,
  XCircle,
  X,
  Clock,
  AlertCircle,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  TrendingUp,
  Wallet,
  SlidersHorizontal,
  Building2,
  RotateCcw,
} from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()

// Filters
const searchKeyword = ref('')
const selectedMarketplace = ref('')
const selectedStatus = ref('')
const filterProduct = ref('')
const filterWarehouse = ref('')
const filterShippingProvider = ref('')
const syncingMarketplace = ref<string | null>(null)
const getTodayDate = () => {
  const d = new Date()
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
  return d.toISOString().split('T')[0]
}

const dateFrom = ref(getTodayDate())
const dateTo = ref(getTodayDate())
const currentPage = ref(1)
const pageSize = ref(20)

const marketplaceOptions = [
  { value: '', label: 'Semua Platform' },
  { value: 'tiktok', label: 'TikTok Shop' },
  { value: 'shopee', label: 'Shopee' },
]

const statusOptions = [
  { value: '', label: 'Semua Status' },
  { value: 'AWAITING_SHIPMENT', label: 'Menunggu Pengiriman' },
  { value: 'AWAITING_COLLECTION', label: 'Menunggu Penjemputan' },
  { value: 'IN_TRANSIT', label: 'Dalam Pengiriman' },
  { value: 'DELIVERED', label: 'Terkirim' },
  { value: 'PACKED', label: 'Selesai' },
  { value: 'CANCELED', label: 'Dibatalkan' },
  { value: 'PENDING', label: 'Pending' },
  { value: 'RETURN_SUCCESS', label: 'Retur Sukses' },
]

function getStatusBadge(status: string) {
  const map: Record<string, { label: string; class: string }> = {
    AWAITING_SHIPMENT:   { label: 'Menunggu Pengiriman', class: 'bg-blue-100 text-blue-700 border border-blue-200 font-semibold' },
    AWAITING_COLLECTION: { label: 'Menunggu Penjemputan', class: 'bg-indigo-100 text-indigo-700 border border-indigo-200 font-semibold' },
    IN_TRANSIT:          { label: 'Dalam Pengiriman',    class: 'bg-purple-100 text-purple-700 border border-purple-200 font-semibold' },
    DELIVERED:           { label: 'Terkirim',            class: 'bg-teal-100 text-teal-700 border border-teal-200 font-semibold' },
    COMPLETED:           { label: 'Selesai',             class: 'bg-emerald-100 text-emerald-800 border border-emerald-300 font-bold' },
    PACKED:              { label: 'Selesai',             class: 'bg-emerald-100 text-emerald-700 border border-emerald-200 font-semibold' },
    VERIFIED:            { label: 'Selesai',             class: 'bg-emerald-100 text-emerald-700 border border-emerald-200 font-semibold' },
    CANCELED:            { label: 'Dibatalkan',          class: 'bg-rose-100 text-rose-700 border border-rose-200 font-bold' },
    CANCELLED:           { label: 'Dibatalkan',          class: 'bg-rose-100 text-rose-700 border border-rose-200 font-bold' },
    PENDING:             { label: 'Pending',             class: 'bg-amber-100 text-amber-700 border border-amber-200 font-semibold' },
    RETURN_SUCCESS:      { label: 'Retur Sukses',        class: 'bg-orange-100 text-orange-700 border border-orange-200 font-semibold' },
  }
  return map[status] ?? { label: status, class: 'bg-slate-100 text-slate-600 border border-slate-200' }
}

function getStatusIcon(status: string) {
  const icons: Record<string, typeof CheckCircle> = {
    COMPLETED: CheckCircle,
    PACKED: CheckCircle,
    VERIFIED: CheckCircle,
    DELIVERED: CheckCircle,
    CANCELED: XCircle,
    CANCELLED: XCircle,
    AWAITING_SHIPMENT: Clock,
    AWAITING_COLLECTION: Clock,
    IN_TRANSIT: Truck,
    RETURN_SUCCESS: RefreshCw,
  }
  return icons[status] ?? AlertCircle
}

function getShippingBadge(provider: string) {
  if (!provider || provider === '-' || provider === 'Belum Ditentukan') {
    return { label: '-', class: 'text-slate-400 font-mono text-xs', isCargo: false }
  }
  const provLower = provider.toLowerCase()
  if (provLower.includes('cargo') || provLower.includes('kargo')) {
    return {
      label: 'J&T Cargo',
      class: 'bg-amber-50 text-amber-800 border border-amber-300 font-bold',
      isCargo: true,
    }
  }
  if (
    provLower.includes('express') ||
    provLower.includes('j&t') ||
    provLower.includes('jnt')
  ) {
    return {
      label: 'J&T Express',
      class: 'bg-blue-50 text-blue-700 border border-blue-200 font-semibold',
      isCargo: false,
    }
  }
  if (provLower.includes('spx') || provLower.includes('shopee express')) {
    return {
      label: 'SPX Express',
      class: 'bg-orange-50 text-orange-700 border border-orange-200 font-semibold',
      isCargo: false,
    }
  }
  if (provLower.includes('ninja')) {
    return {
      label: 'Ninja Van',
      class: 'bg-rose-50 text-rose-700 border border-rose-200 font-semibold',
      isCargo: false,
    }
  }
  if (provLower.includes('sicepat')) {
    return {
      label: 'SiCepat',
      class: 'bg-red-50 text-red-700 border border-red-200 font-semibold',
      isCargo: false,
    }
  }
  if (
    provLower.includes('standard') ||
    provLower.includes('economical') ||
    provLower.includes('nextday') ||
    provLower.includes('hemat')
  ) {
    return { label: '-', class: 'text-slate-400 font-mono text-xs', isCargo: false }
  }
  return {
    label: provider,
    class: 'bg-slate-100 text-slate-700 border border-slate-200 font-medium',
    isCargo: false,
  }
}

async function loadOrders() {
  await marketplaceStore.fetchMarketplaceOrders({
    marketplace: selectedMarketplace.value || undefined,
    status: selectedStatus.value,
    keyword: searchKeyword.value,
    product_name: filterProduct.value,
    warehouse_name: filterWarehouse.value,
    shipping_provider: filterShippingProvider.value,
    from: dateFrom.value,
    to: dateTo.value,
    page: currentPage.value,
    pageSize: pageSize.value,
  })
}

async function handleSync(marketplace?: string) {
  syncingMarketplace.value = marketplace || 'all'
  try {
    const res = await marketplaceStore.syncMarketplaceOrders(marketplace)
    const targetLabel = marketplace === 'shopee' ? 'Shopee' : marketplace === 'tiktok' ? 'TikTok Shop' : 'Semua Marketplace'
    toast.success(`Berhasil! ${res?.synced_count || 0} pesanan ${targetLabel} baru ditarik.`)
    await loadOrders()
  } catch (error: any) {
    const targetLabel = marketplace === 'shopee' ? 'Shopee' : marketplace === 'tiktok' ? 'TikTok' : 'Order'
    toast.error(`Gagal menarik data ${targetLabel}: ` + (error.response?.data?.error || error.message || 'Error tidak diketahui'))
  } finally {
    syncingMarketplace.value = null
  }
}

// Single order fetch modal state
const isSingleModalOpen = ref(false)
const singleOrderId = ref('')
const singleMarketplace = ref('auto')
const singleAsPendingPack = ref(true)
const isFetchingSingle = ref(false)

// Direct pack state
const packingOrderId = ref<string | null>(null)

function openSingleFetchModal() {
  singleOrderId.value = ''
  singleMarketplace.value = 'auto'
  singleAsPendingPack.value = true
  isSingleModalOpen.value = true
}

async function handleFetchSingle() {
  const oid = singleOrderId.value.trim()
  if (!oid) {
    toast.error('Masukkan Order ID yang valid')
    return
  }

  isFetchingSingle.value = true
  try {
    const res = await marketplaceStore.fetchSingleMarketplaceOrder(
      oid,
      singleMarketplace.value === 'auto' ? undefined : singleMarketplace.value,
      singleAsPendingPack.value
    )
    toast.success(res?.message || `Order ${oid} berhasil ditarik!`)
    isSingleModalOpen.value = false
    searchKeyword.value = oid
    await loadOrders()
  } catch (error: any) {
    toast.error(`Gagal menarik order: ` + (error.response?.data?.error || error.message || 'Error tidak diketahui'))
  } finally {
    isFetchingSingle.value = false
  }
}

async function handleDirectPack(order: any) {
  if (!confirm(`Konfirmasi Pack & Potong Stok untuk Order ${order.order_id}?\nStok gudang akan dipotong otomatis di sistem.`)) {
    return
  }

  packingOrderId.value = order.order_id
  try {
    await marketplaceStore.packMarketplaceOrder(order.order_id)
    toast.success(`Order ${order.order_id} berhasil dipacking dan stok terpotong!`)
    await loadOrders()
  } catch (error: any) {
    toast.error(`Gagal memproses packing: ` + (error.response?.data?.error || error.message || 'Terjadi kesalahan'))
  } finally {
    packingOrderId.value = null
  }
}

function handleSearch() {
  currentPage.value = 1
  loadOrders()
}

const isAdvancedFilterOpen = ref(false)
const activePreset = ref<'today' | '7days' | '30days' | 'this_month' | 'custom'>('today')

function applyDatePreset(preset: 'today' | '7days' | '30days' | 'this_month') {
  activePreset.value = preset
  const today = new Date()
  const toStr = getTodayDate()

  if (preset === 'today') {
    dateFrom.value = toStr
    dateTo.value = toStr
  } else if (preset === '7days') {
    const d = new Date()
    d.setDate(d.getDate() - 6)
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
    dateFrom.value = d.toISOString().split('T')[0]
    dateTo.value = toStr
  } else if (preset === '30days') {
    const d = new Date()
    d.setDate(d.getDate() - 29)
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
    dateFrom.value = d.toISOString().split('T')[0]
    dateTo.value = toStr
  } else if (preset === 'this_month') {
    const d = new Date(today.getFullYear(), today.getMonth(), 1)
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
    dateFrom.value = d.toISOString().split('T')[0]
    dateTo.value = toStr
  }
  currentPage.value = 1
  loadOrders()
}

function formatDateOnly(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
}

const activeFilterChips = computed(() => {
  const chips: { key: string; label: string; value: string }[] = []

  if (searchKeyword.value.trim()) {
    chips.push({ key: 'keyword', label: 'Cari', value: `"${searchKeyword.value.trim()}"` })
  }
  if (selectedMarketplace.value) {
    const opt = marketplaceOptions.find((o) => o.value === selectedMarketplace.value)
    chips.push({ key: 'marketplace', label: 'Platform', value: opt ? opt.label : selectedMarketplace.value })
  }
  if (selectedStatus.value) {
    const opt = statusOptions.find((o) => o.value === selectedStatus.value)
    chips.push({ key: 'status', label: 'Status', value: opt ? opt.label : selectedStatus.value })
  }
  if (filterProduct.value) {
    chips.push({ key: 'product', label: 'Produk', value: filterProduct.value })
  }
  if (filterWarehouse.value) {
    chips.push({ key: 'warehouse', label: 'Gudang', value: filterWarehouse.value })
  }
  if (filterShippingProvider.value) {
    chips.push({ key: 'shipping', label: 'Kurir', value: filterShippingProvider.value })
  }
  const todayStr = getTodayDate()
  if (dateFrom.value !== todayStr || dateTo.value !== todayStr) {
    chips.push({ key: 'date', label: 'Tanggal', value: `${formatDateOnly(dateFrom.value)} - ${formatDateOnly(dateTo.value)}` })
  }
  return chips
})

const advancedFiltersCount = computed(() => {
  let count = 0
  if (filterProduct.value) count++
  if (filterWarehouse.value) count++
  if (filterShippingProvider.value) count++
  const todayStr = getTodayDate()
  if (dateFrom.value !== todayStr || dateTo.value !== todayStr) count++
  return count
})

function removeFilter(key: string) {
  if (key === 'keyword') searchKeyword.value = ''
  else if (key === 'marketplace') selectedMarketplace.value = ''
  else if (key === 'status') selectedStatus.value = ''
  else if (key === 'product') filterProduct.value = ''
  else if (key === 'warehouse') filterWarehouse.value = ''
  else if (key === 'shipping') filterShippingProvider.value = ''
  else if (key === 'date') {
    dateFrom.value = getTodayDate()
    dateTo.value = getTodayDate()
    activePreset.value = 'today'
  }
  handleSearch()
}

function resetFilters() {
  searchKeyword.value = ''
  selectedMarketplace.value = ''
  selectedStatus.value = ''
  filterProduct.value = ''
  filterWarehouse.value = ''
  filterShippingProvider.value = ''
  dateFrom.value = getTodayDate()
  dateTo.value = getTodayDate()
  activePreset.value = 'today'
  currentPage.value = 1
  loadOrders()
}

const totalPages = computed(() =>
  Math.max(1, Math.ceil(marketplaceStore.marketplaceOrdersTotal / pageSize.value))
)

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  loadOrders()
}

function formatCurrency(amount: number, currency = 'IDR') {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('id-ID', {
    day: '2-digit', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

onMounted(() => {
  marketplaceStore.fetchOrderFilters()
  loadOrders()
})

watch(() => masterDataStore.selectedBranchId, () => {
  currentPage.value = 1
  loadOrders()
})
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-500">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-cyan-500 to-blue-600 shadow-lg shadow-blue-200">
            <ShoppingCart class="h-6 w-6 text-white" />
          </div>
          Order Marketplace
        </h1>
        <p class="text-slate-500 text-sm font-medium mt-1">
          Lihat dan tarik data order dari TikTok Shop &amp; Shopee
        </p>
      </div>
      <div class="flex items-center gap-2.5 flex-wrap">
        <RouterLink
          to="/marketplace/settlements"
          class="flex items-center gap-2 bg-white border border-slate-200 text-slate-700 px-3.5 py-2.5 rounded-xl font-bold text-sm shadow-sm hover:bg-slate-50 hover:border-slate-300 transition-all"
        >
          <Wallet class="h-4 w-4 text-emerald-600" />
          Pencairan &amp; Keuangan
        </RouterLink>

        <!-- Tarik Order Shopee -->
        <button
          @click="handleSync('shopee')"
          :disabled="syncingMarketplace !== null || marketplaceStore.isSyncing"
          class="flex items-center gap-2 bg-gradient-to-r from-orange-500 to-amber-600 hover:from-orange-600 hover:to-amber-700 text-white px-4 py-2.5 rounded-xl font-bold text-sm shadow-md shadow-orange-200 transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
          title="Tarik pesanan khusus toko Shopee"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': syncingMarketplace === 'shopee' }" />
          <span class="w-2 h-2 rounded-full bg-white"></span>
          {{ syncingMarketplace === 'shopee' ? 'Menarik Shopee...' : 'Tarik Order Shopee' }}
        </button>

        <!-- Tarik Order TikTok -->
        <button
          @click="handleSync('tiktok')"
          :disabled="syncingMarketplace !== null || marketplaceStore.isSyncing"
          class="flex items-center gap-2 bg-slate-900 hover:bg-slate-800 text-white px-4 py-2.5 rounded-xl font-bold text-sm shadow-md shadow-slate-300 transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
          title="Tarik pesanan khusus TikTok Shop"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': syncingMarketplace === 'tiktok' }" />
          <span class="w-2 h-2 rounded-full bg-cyan-400"></span>
          {{ syncingMarketplace === 'tiktok' ? 'Menarik TikTok...' : 'Tarik Order TikTok' }}
        </button>

        <!-- Tarik Semua Marketplace -->
        <button
          @click="handleSync()"
          :disabled="syncingMarketplace !== null || marketplaceStore.isSyncing"
          class="flex items-center gap-2 bg-gradient-to-r from-cyan-500 to-blue-600 text-white px-4 py-2.5 rounded-xl font-bold text-sm shadow-md shadow-blue-200 hover:shadow-lg hover:shadow-blue-300 transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
          title="Tarik seluruh pesanan dari semua marketplace (TikTok & Shopee)"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': syncingMarketplace === 'all' }" />
          {{ syncingMarketplace === 'all' ? 'Menarik Semua...' : 'Tarik Semua' }}
        </button>

        <!-- Tarik Order Spesifik (By ID) -->
        <button
          @click="openSingleFetchModal"
          :disabled="syncingMarketplace !== null || marketplaceStore.isSyncing"
          class="flex items-center gap-2 bg-white hover:bg-indigo-50 text-indigo-700 border-2 border-indigo-200 hover:border-indigo-400 px-4 py-2.5 rounded-xl font-bold text-sm shadow-sm transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
          title="Tarik 1 pesanan tertentu langsung dari Shopee/TikTok menggunakan Order ID"
        >
          <Search class="h-4 w-4 text-indigo-600" />
          Tarik by Order ID
        </button>
      </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
      <!-- Total Omzet -->
      <Card class="border-none shadow-md bg-gradient-to-br from-indigo-50 via-blue-50 to-indigo-100/70 border border-indigo-100/60">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2.5 bg-indigo-600 rounded-xl flex-shrink-0 shadow-sm shadow-indigo-300">
            <Wallet class="h-5 w-5 text-white" />
          </div>
          <div class="min-w-0">
            <p class="text-[11px] text-indigo-700 font-bold uppercase tracking-wide truncate">Total Omzet</p>
            <p class="text-xl lg:text-2xl font-black text-indigo-950 truncate" :title="formatCurrency(marketplaceStore.marketplaceOrdersTotalOmzet)">
              {{ formatCurrency(marketplaceStore.marketplaceOrdersTotalOmzet) }}
            </p>
          </div>
        </CardContent>
      </Card>

      <!-- Total Order -->
      <Card class="border-none shadow-md bg-gradient-to-br from-blue-50 to-blue-100">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2.5 bg-blue-500 rounded-xl flex-shrink-0">
            <ShoppingCart class="h-5 w-5 text-white" />
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-[11px] text-blue-600 font-bold uppercase tracking-wide truncate">Total Order</p>
            <p class="text-xl lg:text-2xl font-black text-blue-900 truncate">{{ marketplaceStore.marketplaceOrdersTotal.toLocaleString('id-ID') }}</p>
            <div class="flex items-center gap-1.5 mt-1 flex-wrap">
              <span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-bold bg-slate-900 text-white shadow-xs" title="Total Order TikTok Shop">
                <span class="w-1.5 h-1.5 rounded-full bg-[#00f2fe]"></span>
                TT: {{ marketplaceStore.marketplaceOrdersTotalTiktok.toLocaleString('id-ID') }}
              </span>
              <span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-bold bg-[#ee4d2d] text-white shadow-xs" title="Total Order Shopee">
                <span class="w-1.5 h-1.5 rounded-full bg-amber-200"></span>
                SP: {{ marketplaceStore.marketplaceOrdersTotalShopee.toLocaleString('id-ID') }}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Menunggu Penjemputan -->
      <Card class="border-none shadow-md bg-gradient-to-br from-yellow-50 to-amber-100">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2.5 bg-amber-500 rounded-xl flex-shrink-0">
            <Clock class="h-5 w-5 text-white" />
          </div>
          <div class="min-w-0">
            <p class="text-[11px] text-amber-700 font-bold uppercase tracking-wide truncate">Menunggu Jemput</p>
            <p class="text-xl lg:text-2xl font-black text-amber-900 truncate">
              {{ marketplaceStore.statsAwaitingCollection.toLocaleString('id-ID') }}
            </p>
          </div>
        </CardContent>
      </Card>

      <!-- Selesai -->
      <Card class="border-none shadow-md bg-gradient-to-br from-emerald-50 to-green-100">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2.5 bg-emerald-500 rounded-xl flex-shrink-0">
            <CheckCircle class="h-5 w-5 text-white" />
          </div>
          <div class="min-w-0">
            <p class="text-[11px] text-emerald-700 font-bold uppercase tracking-wide truncate">Selesai</p>
            <p class="text-xl lg:text-2xl font-black text-emerald-900 truncate">
              {{ marketplaceStore.statsPacked.toLocaleString('id-ID') }}
            </p>
          </div>
        </CardContent>
      </Card>

      <!-- Dibatalkan -->
      <Card class="border-none shadow-md bg-gradient-to-br from-red-50 to-rose-100">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2.5 bg-rose-500 rounded-xl flex-shrink-0">
            <XCircle class="h-5 w-5 text-white" />
          </div>
          <div class="min-w-0">
            <p class="text-[11px] text-rose-700 font-bold uppercase tracking-wide truncate">Dibatalkan</p>
            <p class="text-xl lg:text-2xl font-black text-rose-900 truncate">
              {{ marketplaceStore.statsCanceled.toLocaleString('id-ID') }}
            </p>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Filters -->
    <Card class="border-none shadow-sm rounded-3xl bg-white overflow-hidden">
      <CardContent class="p-5 space-y-4">
        <!-- Tier 1: Baris Utama (Primary Bar) -->
        <div class="flex flex-col md:flex-row items-stretch md:items-center gap-3">
          <!-- Search Input (Flex-1) -->
          <div class="relative flex-1 min-w-[220px]">
            <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              v-model="searchKeyword"
              @keyup.enter="handleSearch"
              type="text"
              placeholder="Cari Order ID, SKU produk, nama..."
              class="w-full pl-10 pr-9 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 outline-none transition-all"
            />
            <button
              v-if="searchKeyword"
              @click="searchKeyword = ''; handleSearch()"
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5 rounded-full"
            >
              <X class="h-3.5 w-3.5" />
            </button>
          </div>

          <!-- Platform Selector -->
          <div class="relative w-full sm:w-44 shrink-0">
            <Store class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" />
            <select
              v-model="selectedMarketplace"
              @change="handleSearch"
              class="w-full pl-10 pr-8 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all appearance-none cursor-pointer"
            >
              <option v-for="opt in marketplaceOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
            <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
          </div>

          <!-- Status Selector -->
          <div class="relative w-full sm:w-52 shrink-0">
            <Filter class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" />
            <select
              v-model="selectedStatus"
              @change="handleSearch"
              class="w-full pl-10 pr-8 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all appearance-none cursor-pointer"
            >
              <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
            <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
          </div>

          <!-- Toggle Filter Lanjutan Button -->
          <button
            type="button"
            @click="isAdvancedFilterOpen = !isAdvancedFilterOpen"
            class="flex items-center justify-center gap-2 px-4 py-2.5 rounded-2xl text-xs font-bold border transition-all shrink-0 cursor-pointer"
            :class="
              isAdvancedFilterOpen || advancedFiltersCount > 0
                ? 'bg-blue-50 border-blue-200 text-blue-700'
                : 'bg-slate-50 border-slate-200 text-slate-600 hover:bg-slate-100 hover:text-slate-900'
            "
          >
            <SlidersHorizontal class="h-3.5 w-3.5" />
            <span>Filter Lanjutan</span>
            <span
              v-if="advancedFiltersCount > 0"
              class="px-1.5 py-0.5 rounded-full text-[10px] font-black bg-blue-600 text-white leading-none"
            >
              {{ advancedFiltersCount }}
            </span>
            <ChevronDown
              class="h-3.5 w-3.5 transition-transform duration-200"
              :class="{ 'rotate-180': isAdvancedFilterOpen }"
            />
          </button>

          <!-- Action Buttons (Cari & Reset) -->
          <div class="flex items-center gap-2 shrink-0">
            <button
              @click="handleSearch"
              class="flex-1 sm:flex-none flex items-center justify-center gap-1.5 bg-blue-600 hover:bg-blue-700 text-white px-5 py-2.5 rounded-2xl text-xs font-black transition-all shadow-sm shadow-blue-200 cursor-pointer"
            >
              <Search class="h-3.5 w-3.5" />
              <span>Cari</span>
            </button>
            <button
              @click="resetFilters"
              class="flex-1 sm:flex-none flex items-center justify-center gap-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 px-4 py-2.5 rounded-2xl text-xs font-bold transition-all cursor-pointer"
              title="Reset semua filter"
            >
              <RotateCcw class="h-3.5 w-3.5" />
              <span>Reset</span>
            </button>
          </div>
        </div>

        <!-- Tier 2: Grid Filter Sekunder / Lanjutan (Collapsible) -->
        <div
          v-show="isAdvancedFilterOpen"
          class="pt-4 border-t border-slate-100 space-y-3 animate-in fade-in duration-200"
        >
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            <!-- 1. Produk -->
            <div class="space-y-1.5">
              <label class="text-[10px] font-black uppercase tracking-wider text-slate-400 block">
                Filter Produk
              </label>
              <div class="relative">
                <Package class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" />
                <select
                  v-model="filterProduct"
                  @change="handleSearch"
                  class="w-full pl-10 pr-8 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all truncate appearance-none cursor-pointer"
                >
                  <option value="">Semua Produk</option>
                  <option v-for="prod in marketplaceStore.orderFilters.products" :key="prod" :value="prod">
                    {{ prod }}
                  </option>
                </select>
                <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
              </div>
            </div>

            <!-- 2. Gudang Cabang -->
            <div class="space-y-1.5">
              <label class="text-[10px] font-black uppercase tracking-wider text-slate-400 block">
                Gudang Cabang
              </label>
              <div class="relative">
                <Building2 class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" />
                <select
                  v-model="filterWarehouse"
                  @change="handleSearch"
                  class="w-full pl-10 pr-8 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all truncate appearance-none cursor-pointer"
                >
                  <option value="">Semua Gudang</option>
                  <option v-for="wh in marketplaceStore.orderFilters.warehouses" :key="wh" :value="wh">
                    {{ wh }}
                  </option>
                </select>
                <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
              </div>
            </div>

            <!-- 3. Ekspedisi / Kurir -->
            <div class="space-y-1.5">
              <label class="text-[10px] font-black uppercase tracking-wider text-slate-400 block">
                Jasa Pengiriman
              </label>
              <div class="relative">
                <Truck class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" />
                <select
                  v-model="filterShippingProvider"
                  @change="handleSearch"
                  class="w-full pl-10 pr-8 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all truncate appearance-none cursor-pointer"
                >
                  <option value="">Semua Pengiriman</option>
                  <option value="Cargo">J&T Cargo / Kargo</option>
                  <option value="Express">J&T Express / Reguler</option>
                  <option
                    v-for="provider in marketplaceStore.orderFilters.shipping_providers.filter(
                      (p) => !p.toLowerCase().includes('cargo') && !p.toLowerCase().includes('express')
                    )"
                    :key="provider"
                    :value="provider"
                  >
                    {{ provider }}
                  </option>
                </select>
                <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
              </div>
            </div>

            <!-- 4. Rentang Tanggal & Preset Cepat -->
            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <label class="text-[10px] font-black uppercase tracking-wider text-slate-400 block">
                  Rentang Tanggal
                </label>
                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    @click="applyDatePreset('today')"
                    class="text-[10px] font-bold px-1.5 py-0.5 rounded transition-all cursor-pointer"
                    :class="activePreset === 'today' ? 'bg-blue-100 text-blue-700' : 'text-slate-400 hover:text-slate-700'"
                  >
                    Hari Ini
                  </button>
                  <span class="text-slate-300 text-[10px]">•</span>
                  <button
                    type="button"
                    @click="applyDatePreset('7days')"
                    class="text-[10px] font-bold px-1.5 py-0.5 rounded transition-all cursor-pointer"
                    :class="activePreset === '7days' ? 'bg-blue-100 text-blue-700' : 'text-slate-400 hover:text-slate-700'"
                  >
                    7 Hari
                  </button>
                  <span class="text-slate-300 text-[10px]">•</span>
                  <button
                    type="button"
                    @click="applyDatePreset('30days')"
                    class="text-[10px] font-bold px-1.5 py-0.5 rounded transition-all cursor-pointer"
                    :class="activePreset === '30days' ? 'bg-blue-100 text-blue-700' : 'text-slate-400 hover:text-slate-700'"
                  >
                    30 Hari
                  </button>
                </div>
              </div>
              <div class="grid grid-cols-2 gap-2">
                <div class="relative">
                  <Calendar class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
                  <input
                    v-model="dateFrom"
                    @change="activePreset = 'custom'; handleSearch()"
                    type="date"
                    class="w-full pl-8 pr-2 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all"
                  />
                </div>
                <div class="relative">
                  <Calendar class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
                  <input
                    v-model="dateTo"
                    @change="activePreset = 'custom'; handleSearch()"
                    type="date"
                    class="w-full pl-8 pr-2 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Tier 3: Active Filter Chips -->
        <div
          v-if="activeFilterChips.length > 0"
          class="flex flex-wrap items-center gap-2 pt-3 border-t border-slate-100"
        >
          <span class="text-xs font-bold text-slate-400 flex items-center gap-1">
            <Filter class="h-3 w-3" /> Filter Aktif:
          </span>
          <span
            v-for="chip in activeFilterChips"
            :key="chip.key"
            class="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-200/80 group"
          >
            <span class="text-slate-400 font-normal">{{ chip.label }}:</span>
            <span class="font-bold text-slate-900">{{ chip.value }}</span>
            <button
              @click="removeFilter(chip.key)"
              type="button"
              class="hover:text-rose-600 transition-colors ml-0.5 cursor-pointer"
              :title="`Hapus filter ${chip.label}`"
            >
              <X class="h-3 w-3" />
            </button>
          </span>
          <button
            @click="resetFilters"
            type="button"
            class="text-xs font-bold text-rose-600 hover:text-rose-700 hover:underline px-2 py-1 ml-1 cursor-pointer"
          >
            Hapus Semua
          </button>
        </div>
      </CardContent>
    </Card>

    <!-- Table -->
    <Card class="border-none shadow-sm">
      <CardHeader class="pb-3 border-b border-slate-100">
        <div class="flex items-center justify-between">
          <CardTitle class="text-base font-black text-slate-800 flex items-center gap-2">
            <TrendingUp class="h-4 w-4 text-blue-500" />
            Daftar Order
          </CardTitle>
          <span class="text-xs text-slate-500 font-medium">
            Menampilkan {{ marketplaceStore.marketplaceOrders.length }} dari {{ marketplaceStore.marketplaceOrdersTotal }} order
          </span>
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <!-- Loading -->
        <div v-if="marketplaceStore.isOrdersLoading" class="flex items-center justify-center py-16">
          <div class="flex flex-col items-center gap-3">
            <RefreshCw class="h-8 w-8 text-blue-500 animate-spin" />
            <p class="text-sm text-slate-500 font-medium">Memuat data order...</p>
          </div>
        </div>

        <!-- Empty -->
        <div v-else-if="marketplaceStore.marketplaceOrders.length === 0" class="flex flex-col items-center justify-center py-16 gap-3">
          <div class="p-4 bg-slate-100 rounded-2xl">
            <ShoppingCart class="h-10 w-10 text-slate-300" />
          </div>
          <p class="text-slate-500 font-semibold">Belum ada order</p>
          <p class="text-slate-400 text-sm">Klik "Tarik Order TikTok" untuk sinkronisasi data</p>
        </div>

        <!-- Table -->
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead class="bg-slate-50 border-b border-slate-100">
              <tr>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Order ID</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Produk</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">SKU</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Warehouse</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Pengiriman</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Total</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Status</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Tanggal</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Tracking</th>
                <th class="text-center py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50">
              <template v-for="order in marketplaceStore.marketplaceOrders" :key="order.order_id">
                <tr
                  v-for="(item, idx) in order.items"
                  :key="item.id"
                  class="hover:bg-slate-50/80 transition-colors"
                >
                  <td v-if="idx === 0" :rowspan="order.items.length" class="py-3 px-4 border-b border-r border-slate-100 align-top">
                    <div class="flex items-center gap-1.5">
                      <span
                        class="text-xs px-2 py-0.5 rounded-lg font-mono font-bold border"
                        :class="order.marketplace === 'shopee'
                          ? 'bg-orange-50 text-orange-700 border-orange-200'
                          : 'bg-cyan-50 text-cyan-700 border-cyan-200'"
                      >
                        {{ order.marketplace === 'shopee' ? 'SP' : 'TK' }}
                      </span>
                      <span class="font-bold text-slate-800 text-xs">{{ order.order_id }}</span>
                    </div>
                  </td>
                  <td class="py-3 px-4 align-top">
                    <div class="flex items-center gap-2">
                      <div class="p-1.5 bg-slate-100 rounded-lg flex-shrink-0">
                        <Package class="h-3.5 w-3.5 text-slate-500" />
                      </div>
                      <p class="font-semibold text-slate-800 text-xs line-clamp-2 max-w-[180px]">{{ item.product_name || '-' }}</p>
                    </div>
                  </td>
                  <td class="py-3 px-4 align-top">
                    <div class="flex flex-col gap-1">
                      <span class="font-mono text-xs text-slate-600 bg-slate-100 px-2 py-0.5 rounded w-fit">
                        {{ item.seller_sku || '-' }}
                      </span>
                      <span class="text-[11px] text-cyan-700 font-bold bg-cyan-50 border border-cyan-200 px-2 py-0.5 rounded w-fit flex items-center gap-1">
                        📦 Qty: {{ item.item_count }} pcs
                      </span>
                    </div>
                  </td>
                  <td v-if="idx === 0" :rowspan="order.items.length" class="py-3 px-4 border-b border-l border-slate-100 align-top">
                    <span class="text-slate-700 text-xs font-medium">{{ order.warehouse_name || '-' }}</span>
                  </td>
                  <td v-if="idx === 0" :rowspan="order.items.length" class="py-3 px-4 border-b border-slate-100 align-top">
                    <div v-if="order.shipping_provider" class="flex items-center gap-1.5">
                      <span
                        class="text-[11px] px-2.5 py-1 rounded-lg flex items-center gap-1.5 shadow-sm whitespace-nowrap"
                        :class="getShippingBadge(order.shipping_provider).class"
                      >
                        <Truck class="h-3 w-3 flex-shrink-0" :class="getShippingBadge(order.shipping_provider).isCargo ? 'text-amber-600' : 'text-blue-500'" />
                        {{ order.shipping_provider }}
                      </span>
                    </div>
                    <span v-else class="text-xs text-slate-400 font-mono">-</span>
                  </td>
                  <td v-if="idx === 0" :rowspan="order.items.length" class="py-3 px-4 border-b border-slate-100 align-top">
                    <span class="text-slate-900 font-bold text-xs">
                      {{ order.total_amount > 0 ? formatCurrency(order.total_amount, order.currency) : '-' }}
                    </span>
                  </td>
                  <td v-if="idx === 0" :rowspan="order.items.length" class="py-3 px-4 border-b border-slate-100 align-top">
                    <div class="flex items-center gap-1.5">
                      <component :is="getStatusIcon(order.order_status)" class="h-3.5 w-3.5 flex-shrink-0" :class="getStatusBadge(order.order_status).class.includes('emerald') ? 'text-emerald-600' : getStatusBadge(order.order_status).class.includes('red') ? 'text-red-600' : getStatusBadge(order.order_status).class.includes('blue') ? 'text-blue-600' : 'text-slate-600'" />
                      <span
                        class="text-[11px] font-bold px-2 py-0.5 rounded-lg"
                        :class="getStatusBadge(order.order_status).class"
                      >
                        {{ getStatusBadge(order.order_status).label }}
                      </span>
                    </div>
                  </td>
                  <td v-if="idx === 0" :rowspan="order.items.length" class="py-3 px-4 border-b border-slate-100 align-top">
                    <span class="text-xs text-slate-500">{{ formatDate(order.created_at) }}</span>
                  </td>
                  <td v-if="idx === 0" :rowspan="order.items.length" class="py-3 px-4 border-b border-slate-100 align-top">
                    <span class="text-xs font-mono text-slate-500 truncate max-w-[120px] block">{{ order.tracking_id || '-' }}</span>
                  </td>
                  <td v-if="idx === 0" :rowspan="order.items.length" class="py-3 px-4 border-b border-l border-slate-100 align-top text-center whitespace-nowrap">
                    <button
                      v-if="order.order_status !== 'PACKED'"
                      @click="handleDirectPack(order)"
                      :disabled="packingOrderId === order.order_id"
                      class="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-bold bg-emerald-50 text-emerald-700 hover:bg-emerald-100 border border-emerald-200 transition-colors shadow-sm disabled:opacity-50"
                      title="Potong stok gudang dan tandai order sebagai PACKED"
                    >
                      <CheckCircle class="h-3.5 w-3.5 text-emerald-600" />
                      {{ packingOrderId === order.order_id ? 'Memproses...' : 'Pack & Potong Stok' }}
                    </button>
                    <span v-else class="text-xs text-slate-400 font-semibold inline-flex items-center gap-1">
                      <CheckCircle class="h-3.5 w-3.5 text-emerald-500" />
                      Packed
                    </span>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="!marketplaceStore.isOrdersLoading && marketplaceStore.marketplaceOrders.length > 0" class="flex items-center justify-between px-4 py-3 border-t border-slate-100">
          <p class="text-xs text-slate-500 font-medium">
            Halaman {{ currentPage }} dari {{ totalPages }}
          </p>
          <div class="flex items-center gap-1">
            <button
              @click="goToPage(currentPage - 1)"
              :disabled="currentPage <= 1"
              class="p-2 rounded-lg text-slate-500 hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronLeft class="h-4 w-4" />
            </button>
            <template v-for="page in totalPages" :key="page">
              <button
                v-if="Math.abs(page - currentPage) <= 2"
                @click="goToPage(page)"
                class="w-8 h-8 rounded-lg text-xs font-bold transition-colors"
                :class="page === currentPage ? 'bg-blue-600 text-white' : 'text-slate-600 hover:bg-slate-100'"
              >
                {{ page }}
              </button>
            </template>
            <button
              @click="goToPage(currentPage + 1)"
              :disabled="currentPage >= totalPages"
              class="p-2 rounded-lg text-slate-500 hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronRight class="h-4 w-4" />
            </button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Modal Tarik Pesanan Manual By Order ID -->
    <div
      v-if="isSingleModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
    >
      <div class="bg-white rounded-2xl shadow-2xl border border-slate-100 w-full max-w-md overflow-hidden animate-in zoom-in-95 duration-200">
        <div class="flex items-center justify-between p-5 border-b border-slate-100 bg-slate-50/50">
          <div class="flex items-center gap-2.5">
            <div class="p-2 bg-indigo-100 text-indigo-700 rounded-xl">
              <Search class="h-5 w-5" />
            </div>
            <div>
              <h3 class="font-black text-slate-900 text-base">Tarik Pesanan Manual</h3>
              <p class="text-xs text-slate-500">Tarik data spesifik langsung dari Shopee / TikTok</p>
            </div>
          </div>
          <button
            @click="isSingleModalOpen = false"
            class="text-slate-400 hover:text-slate-600 p-1.5 rounded-lg hover:bg-slate-100 transition-colors"
          >
            <X class="h-5 w-5" />
          </button>
        </div>

        <div class="p-5 space-y-4">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Order ID / Nomor Pesanan <span class="text-rose-500">*</span>
            </label>
            <input
              v-model="singleOrderId"
              type="text"
              placeholder="Contoh: 261012CRR7WVKF"
              class="w-full px-3.5 py-2.5 border border-slate-200 rounded-xl text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-slate-800"
              @keyup.enter="handleFetchSingle"
            />
            <p class="text-[11px] text-slate-400 mt-1">Masukkan Nomor Pesanan yang tertera di Seller Center.</p>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Platform Marketplace
            </label>
            <select
              v-model="singleMarketplace"
              class="w-full px-3.5 py-2.5 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-slate-700 bg-white"
            >
              <option value="auto">Deteksi Otomatis (Shopee / TikTok)</option>
              <option value="shopee">Shopee</option>
              <option value="tiktok">TikTok Shop</option>
            </select>
          </div>

          <div class="bg-indigo-50/70 border border-indigo-100 rounded-xl p-3.5">
            <label class="flex items-start gap-2.5 cursor-pointer">
              <input
                v-model="singleAsPendingPack"
                type="checkbox"
                class="mt-0.5 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
              />
              <div>
                <span class="text-xs font-bold text-slate-800 block">Masukkan ke Antrean Packing (Status: Menunggu Packing)</span>
                <span class="text-[11px] text-slate-500 block leading-relaxed mt-0.5">
                  Centang opsi ini jika paket sudah dikirim di marketplace tapi stok gudang di SuperApp belum dipotong, agar order dapat di-pack dan stok terpotong sesuai alur operasional.
                </span>
              </div>
            </label>
          </div>
        </div>

        <div class="p-4 border-t border-slate-100 bg-slate-50 flex items-center justify-end gap-2.5">
          <button
            @click="isSingleModalOpen = false"
            class="px-4 py-2 text-sm font-semibold text-slate-600 hover:text-slate-800 hover:bg-slate-200 rounded-xl transition-colors"
          >
            Batal
          </button>
          <button
            @click="handleFetchSingle"
            :disabled="isFetchingSingle || !singleOrderId.trim()"
            class="flex items-center gap-2 px-5 py-2 text-sm font-bold text-white bg-indigo-600 hover:bg-indigo-700 rounded-xl shadow-md shadow-indigo-200 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <RefreshCw v-if="isFetchingSingle" class="h-4 w-4 animate-spin" />
            <Search v-else class="h-4 w-4" />
            {{ isFetchingSingle ? 'Menarik Data...' : 'Tarik Sekarang' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
