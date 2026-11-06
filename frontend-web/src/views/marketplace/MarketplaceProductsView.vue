<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import {
  Tag,
  RefreshCw,
  Search,
  Filter,
  Package,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Clock,
  BarChart3,
  TrendingDown,
  Layers,
} from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()

const searchKeyword = ref('')
const selectedStatus = ref('')
const selectedMarketplace = ref('')
const lastSyncInfo = computed(() => {
  if (!marketplaceStore.marketplaceProducts.length) return null
  const dates = marketplaceStore.marketplaceProducts
    .map(p => new Date(p.last_synced_at).getTime())
    .filter(t => !isNaN(t))
  if (!dates.length) return null
  return new Date(Math.max(...dates))
})

const statusOptions = [
  { value: '', label: 'Semua Status' },
  { value: 'ACTIVATE', label: 'Aktif' },
  { value: 'DRAFT', label: 'Draft' },
  { value: 'SELLER_DEACTIVATED', label: 'Dinonaktifkan' },
  { value: 'PLATFORM_DEACTIVATED', label: 'Dinonaktifkan Platform' },
  { value: 'FREEZE', label: 'Dibekukan' },
  { value: 'DELETED', label: 'Dihapus' },
]

const marketplaceOptions = [
  { value: '', label: 'Semua Marketplace' },
  { value: 'tiktok', label: 'TikTok Shop' },
  { value: 'shopee', label: 'Shopee' },
]

function getStatusBadge(status: string) {
  const map: Record<string, { label: string; class: string; dotClass: string }> = {
    ACTIVATE:            { label: 'Aktif',              class: 'bg-emerald-50 text-emerald-700 border border-emerald-200', dotClass: 'bg-emerald-500' },
    DRAFT:               { label: 'Draft',              class: 'bg-slate-100 text-slate-600 border border-slate-200', dotClass: 'bg-slate-400' },
    SELLER_DEACTIVATED:  { label: 'Dinonaktifkan',      class: 'bg-amber-50 text-amber-700 border border-amber-200', dotClass: 'bg-amber-500' },
    PLATFORM_DEACTIVATED:{ label: 'Nonaktif Platform',  class: 'bg-orange-50 text-orange-700 border border-orange-200', dotClass: 'bg-orange-500' },
    FREEZE:              { label: 'Dibekukan',          class: 'bg-blue-50 text-blue-700 border border-blue-200', dotClass: 'bg-blue-400' },
    DELETED:             { label: 'Dihapus',            class: 'bg-red-50 text-red-700 border border-red-200', dotClass: 'bg-red-500' },
    PENDING:             { label: 'Menunggu Review',    class: 'bg-violet-50 text-violet-700 border border-violet-200', dotClass: 'bg-violet-400' },
  }
  return map[status] ?? { label: status, class: 'bg-slate-100 text-slate-600', dotClass: 'bg-slate-400' }
}

function getStatusIcon(status: string) {
  const icons: Record<string, typeof CheckCircle2> = {
    ACTIVATE: CheckCircle2,
    DRAFT: Clock,
    SELLER_DEACTIVATED: XCircle,
    PLATFORM_DEACTIVATED: XCircle,
    FREEZE: AlertCircle,
    DELETED: XCircle,
  }
  return icons[status] ?? AlertCircle
}

const filteredProducts = computed(() => {
  let products = marketplaceStore.marketplaceProducts || []
  if (selectedMarketplace.value) {
    const mp = selectedMarketplace.value.toLowerCase()
    products = products.filter(p => (p.marketplace || 'tiktok').toLowerCase() === mp)
  }
  if (selectedStatus.value) {
    products = products.filter(p => p.status === selectedStatus.value)
  }
  if (searchKeyword.value.trim()) {
    const kw = searchKeyword.value.trim().toLowerCase()
    products = products.filter(p => {
      const name = String(p.product_name || '').toLowerCase()
      const sellerSku = String(p.seller_sku || '').toLowerCase()
      const skuId = String(p.sku_id || '').toLowerCase()
      const productId = String(p.product_id || '').toLowerCase()
      const category = String(p.category || '').toLowerCase()
      return (
        name.includes(kw) ||
        sellerSku.includes(kw) ||
        skuId.includes(kw) ||
        productId.includes(kw) ||
        category.includes(kw)
      )
    })
  }
  return products
})

// Stats
const statsActive = computed(() => marketplaceStore.marketplaceProducts.filter(p => p.status === 'ACTIVATE').length)
const statsDraft = computed(() => marketplaceStore.marketplaceProducts.filter(p => p.status === 'DRAFT').length)
const statsInactive = computed(() => marketplaceStore.marketplaceProducts.filter(p => ['SELLER_DEACTIVATED', 'PLATFORM_DEACTIVATED'].includes(p.status)).length)
const statsTotal = computed(() => marketplaceStore.marketplaceProducts.length)

const syncingMarketplace = ref<string | null>(null)

async function handleSync(marketplace?: string) {
  syncingMarketplace.value = marketplace || 'all'
  try {
    const res = await marketplaceStore.syncMarketplaceProducts(marketplace)
    const targetLabel = marketplace === 'shopee' ? 'Shopee' : marketplace === 'tiktok' ? 'TikTok Shop' : 'Marketplace'
    toast.success(res?.message || `Sync produk ${targetLabel} berhasil!`)
  } catch (err: any) {
    const targetLabel = marketplace === 'shopee' ? 'Shopee' : marketplace === 'tiktok' ? 'TikTok' : 'Produk'
    const errorMsg = err.response?.data?.error || err.message || 'Gagal melakukan sinkronisasi'
    toast.error(`Sync produk ${targetLabel} gagal: ` + errorMsg)
  } finally {
    syncingMarketplace.value = null
  }
}

function formatCurrency(amount: number, currency = 'IDR') {
  if (!amount || amount === 0) return '-'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatDate(dateStr: string) {
  if (!dateStr || dateStr === '0001-01-01T00:00:00Z') return '-'
  return new Date(dateStr).toLocaleString('id-ID', {
    day: '2-digit', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

onMounted(() => {
  marketplaceStore.fetchMarketplaceProducts()
})

watch(() => masterDataStore.selectedBranchId, () => {
  marketplaceStore.fetchMarketplaceProducts()
})
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-500">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 shadow-lg shadow-purple-200">
            <Tag class="h-6 w-6 text-white" />
          </div>
          Produk & SKU Marketplace
        </h1>
        <p class="text-slate-500 text-sm font-medium mt-1">
          Data produk yang dijual di TikTok Shop &amp; Shopee
          <span v-if="lastSyncInfo" class="text-slate-400">
            — Terakhir sync: {{ formatDate(lastSyncInfo.toISOString()) }}
          </span>
        </p>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-2.5 flex-wrap">
        <!-- Sync TikTok -->
        <button
          @click="handleSync('tiktok')"
          :disabled="syncingMarketplace !== null || marketplaceStore.isSyncing"
          class="flex items-center gap-2 bg-slate-900 hover:bg-slate-800 text-white px-4 py-2.5 rounded-xl font-bold text-sm shadow-md shadow-slate-300 transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': syncingMarketplace === 'tiktok' }" />
          <span class="w-2 h-2 rounded-full bg-cyan-400"></span>
          {{ syncingMarketplace === 'tiktok' ? 'Menarik TikTok...' : 'Sync Produk TikTok' }}
        </button>

        <!-- Sync Shopee -->
        <button
          @click="handleSync('shopee')"
          :disabled="syncingMarketplace !== null || marketplaceStore.isSyncing"
          class="flex items-center gap-2 bg-gradient-to-r from-orange-500 to-amber-600 hover:from-orange-600 hover:to-amber-700 text-white px-4 py-2.5 rounded-xl font-bold text-sm shadow-md shadow-orange-200 transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': syncingMarketplace === 'shopee' }" />
          <span class="w-2 h-2 rounded-full bg-white"></span>
          {{ syncingMarketplace === 'shopee' ? 'Menarik Shopee...' : 'Sync Produk Shopee' }}
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <Card class="border-none shadow-md bg-gradient-to-br from-violet-50 to-purple-100">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2 bg-violet-500 rounded-lg">
            <Layers class="h-5 w-5 text-white" />
          </div>
          <div>
            <p class="text-xs text-violet-600 font-bold uppercase tracking-wide">Total SKU</p>
            <p class="text-2xl font-black text-violet-900">{{ statsTotal.toLocaleString('id-ID') }}</p>
          </div>
        </CardContent>
      </Card>
      <Card class="border-none shadow-md bg-gradient-to-br from-emerald-50 to-green-100">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2 bg-emerald-500 rounded-lg">
            <CheckCircle2 class="h-5 w-5 text-white" />
          </div>
          <div>
            <p class="text-xs text-emerald-700 font-bold uppercase tracking-wide">Aktif</p>
            <p class="text-2xl font-black text-emerald-900">{{ statsActive.toLocaleString('id-ID') }}</p>
          </div>
        </CardContent>
      </Card>
      <Card class="border-none shadow-md bg-gradient-to-br from-amber-50 to-yellow-100">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2 bg-amber-500 rounded-lg">
            <Clock class="h-5 w-5 text-white" />
          </div>
          <div>
            <p class="text-xs text-amber-700 font-bold uppercase tracking-wide">Draft</p>
            <p class="text-2xl font-black text-amber-900">{{ statsDraft.toLocaleString('id-ID') }}</p>
          </div>
        </CardContent>
      </Card>
      <Card class="border-none shadow-md bg-gradient-to-br from-red-50 to-rose-100">
        <CardContent class="p-4 flex items-center gap-3">
          <div class="p-2 bg-rose-500 rounded-lg">
            <TrendingDown class="h-5 w-5 text-white" />
          </div>
          <div>
            <p class="text-xs text-rose-700 font-bold uppercase tracking-wide">Nonaktif</p>
            <p class="text-2xl font-black text-rose-900">{{ statsInactive.toLocaleString('id-ID') }}</p>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Filters -->
    <Card class="border-none shadow-sm">
      <CardContent class="p-4">
        <div class="flex flex-col md:flex-row gap-3">
          <div class="relative flex-1">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              v-model="searchKeyword"
              type="text"
              placeholder="Cari nama produk, SKU, ID produk..."
              class="w-full pl-9 pr-4 py-2.5 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-purple-500/20 focus:border-purple-400 bg-slate-50"
            />
          </div>
          <div class="relative">
            <Filter class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <select
              v-model="selectedMarketplace"
              class="pl-9 pr-8 py-2.5 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-purple-500/20 focus:border-purple-400 bg-slate-50 min-w-[170px] appearance-none"
            >
              <option v-for="opt in marketplaceOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
          <div class="relative">
            <Filter class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <select
              v-model="selectedStatus"
              class="pl-9 pr-8 py-2.5 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-purple-500/20 focus:border-purple-400 bg-slate-50 min-w-[170px] appearance-none"
            >
              <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Table -->
    <Card class="border-none shadow-sm">
      <CardHeader class="pb-3 border-b border-slate-100">
        <div class="flex items-center justify-between">
          <CardTitle class="text-base font-black text-slate-800 flex items-center gap-2">
            <BarChart3 class="h-4 w-4 text-violet-500" />
            Daftar Produk & SKU
          </CardTitle>
          <span class="text-xs text-slate-500 font-medium">
            Menampilkan {{ filteredProducts.length }} dari {{ statsTotal }} produk
          </span>
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <!-- Loading -->
        <div v-if="marketplaceStore.isProductsLoading" class="flex items-center justify-center py-16">
          <div class="flex flex-col items-center gap-3">
            <RefreshCw class="h-8 w-8 text-purple-500 animate-spin" />
            <p class="text-sm text-slate-500 font-medium">Memuat data produk...</p>
          </div>
        </div>

        <!-- Empty -->
        <div v-else-if="filteredProducts.length === 0" class="flex flex-col items-center justify-center py-16 gap-3">
          <div class="p-4 bg-slate-100 rounded-2xl">
            <Tag class="h-10 w-10 text-slate-300" />
          </div>
          <p class="text-slate-500 font-semibold">Belum ada produk</p>
          <p class="text-slate-400 text-sm">Klik "Sync Produk Marketplace" untuk mengambil data produk</p>
        </div>

        <!-- Table -->
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead class="bg-slate-50 border-b border-slate-100">
              <tr>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Produk</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Marketplace</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">SKU ID</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Seller SKU</th>
                <th class="text-right py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Harga</th>
                <th class="text-right py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Stok</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Status</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-500 uppercase tracking-wider">Terakhir Sync</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50">
              <tr
                v-for="product in filteredProducts"
                :key="product.id"
                class="hover:bg-slate-50/80 transition-colors"
              >
                <td class="py-3 px-4">
                  <div class="flex items-center gap-3">
                    <div class="p-1.5 bg-violet-100 rounded-lg flex-shrink-0">
                      <Package class="h-4 w-4 text-violet-600" />
                    </div>
                    <div>
                      <p class="font-bold text-slate-800 text-sm line-clamp-1 max-w-[200px]">{{ product.product_name || '-' }}</p>
                      <p class="text-xs text-slate-400 font-mono mt-0.5">ID: {{ product.product_id }}</p>
                    </div>
                  </div>
                </td>
                <td class="py-3 px-4">
                  <span
                    class="text-[11px] font-bold px-2 py-0.5 rounded-lg border uppercase font-mono"
                    :class="(product.marketplace || '').toLowerCase() === 'shopee'
                      ? 'bg-orange-50 text-orange-700 border-orange-200'
                      : 'bg-cyan-50 text-cyan-700 border-cyan-200'"
                  >
                    {{ (product.marketplace || '').toLowerCase() === 'shopee' ? 'Shopee' : 'TikTok' }}
                  </span>
                </td>
                <td class="py-3 px-4">
                  <span class="font-mono text-xs text-slate-600 bg-slate-100 px-2 py-0.5 rounded">{{ product.sku_id || '-' }}</span>
                </td>
                <td class="py-3 px-4">
                  <span class="font-mono text-xs text-slate-700 font-semibold">{{ product.seller_sku || '-' }}</span>
                </td>
                <td class="py-3 px-4 text-right">
                  <span class="font-bold text-slate-900 text-sm">
                    {{ formatCurrency(product.price_amount, product.price_currency) }}
                  </span>
                </td>
                <td class="py-3 px-4 text-right">
                  <span
                    class="font-black text-base"
                    :class="product.stock === 0 ? 'text-red-500' : product.stock <= 5 ? 'text-amber-500' : 'text-slate-800'"
                  >
                    {{ product.stock.toLocaleString('id-ID') }}
                  </span>
                </td>
                <td class="py-3 px-4">
                  <div class="flex items-center gap-1.5">
                    <span class="w-2 h-2 rounded-full flex-shrink-0" :class="getStatusBadge(product.status).dotClass"></span>
                    <span
                      class="text-[11px] font-bold px-2 py-0.5 rounded-lg"
                      :class="getStatusBadge(product.status).class"
                    >
                      {{ getStatusBadge(product.status).label }}
                    </span>
                  </div>
                </td>
                <td class="py-3 px-4">
                  <span class="text-xs text-slate-400">{{ formatDate(product.last_synced_at) }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
