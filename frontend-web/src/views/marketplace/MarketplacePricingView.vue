<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import { Tag, Search, TrendingUp, HelpCircle, Save, Package } from 'lucide-vue-next'
import { Card, CardContent } from '@/components/ui/card'
import type { MarketplaceStock } from '@/api/marketplace'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()
const searchQuery = ref('')
const platformFilter = ref('ALL')

const isSaving = ref(false)
// Storing temporary edit state: { stock_id: new_price }
const editingPrices = ref<Record<string, number>>({})

let pollingInterval: ReturnType<typeof setInterval> | null = null

async function loadData() {
  await marketplaceStore.fetchStocks()
}

onMounted(async () => {
  await loadData()
  pollingInterval = setInterval(() => {
    // Only poll if not editing anything to prevent overwriting user input
    if (Object.keys(editingPrices.value).length === 0) {
      loadData()
    }
  }, 15000)
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
})

watch(() => masterDataStore.selectedBranchId, () => {
  loadData()
})

const filteredStocks = computed(() => {
  let result = marketplaceStore.stocks
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(
      (s) =>
        s.product_name.toLowerCase().includes(q) || s.sku.toLowerCase().includes(q),
    )
  }
  return result
})

function formatRupiah(val: number) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val)
}

function handlePriceInput(stock: MarketplaceStock, ev: any) {
  const val = Number(ev.target.value)
  if (val !== stock.price) {
    editingPrices.value[stock.id] = val
  } else {
    delete editingPrices.value[stock.id]
  }
}

async function saveAllPrices() {
  const keys = Object.keys(editingPrices.value)
  if (keys.length === 0) return

  isSaving.value = true
  try {
    for (const id of keys) {
      await marketplaceStore.updateStock(id, { price: editingPrices.value[id] })
    }
    editingPrices.value = {}
    await marketplaceStore.fetchStocks()
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal menyimpan perubahan harga')
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-700">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-primary to-primary/80 shadow-md shadow-primary/20">
            <Tag class="h-6 w-6 text-white" />
          </div>
          Manajemen Harga
        </h1>
        <p class="text-slate-500 text-sm font-medium mt-1">
          Kelola harga jual produk marketplace untuk perhitungan omzet
        </p>
      </div>
      <div class="flex items-center gap-3 flex-wrap">
        <button
          @click="saveAllPrices"
          :disabled="Object.keys(editingPrices).length === 0 || isSaving"
          class="flex items-center gap-2 px-5 py-2.5 rounded-xl font-bold shadow-md transition-all"
          :class="Object.keys(editingPrices).length > 0 && !isSaving ? 'bg-gradient-to-r from-emerald-500 to-teal-600 text-white hover:from-emerald-600 hover:to-teal-700 shadow-emerald-200' : 'bg-slate-100 text-slate-400 cursor-not-allowed'"
        >
          <Save class="h-4 w-4" :class="{ 'animate-spin': isSaving }" />
          {{ isSaving ? 'Menyimpan...' : Object.keys(editingPrices).length > 0 ? `Simpan (${Object.keys(editingPrices).length}) Perubahan` : 'Tidak Ada Perubahan' }}
        </button>
      </div>
    </div>

    <!-- Stats & Filters -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <Card class="border-none shadow-sm lg:col-span-2">
        <CardContent class="p-4 flex flex-col sm:flex-row gap-3 items-center h-full">
          <div class="relative flex-1 w-full">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari produk / SKU..."
              class="w-full pl-9 pr-4 py-3 text-sm font-medium bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/50 transition-all"
            />
          </div>
          <div class="flex items-center gap-2 flex-wrap">
          </div>
        </CardContent>
      </Card>
      
      <Card class="border-none shadow-sm bg-gradient-to-br from-amber-500 to-orange-600 text-white">
        <CardContent class="p-6 relative overflow-hidden h-full flex flex-col justify-center">
          <div class="relative z-10">
            <div class="flex items-center gap-2 text-amber-100 mb-1">
              <TrendingUp class="h-4 w-4" />
              <h3 class="text-sm font-bold uppercase tracking-wider">Estimasi Harga Rata-Rata</h3>
            </div>
            <p class="text-3xl font-black">
              {{ formatRupiah(filteredStocks.length ? (filteredStocks.reduce((sum, s) => sum + s.price, 0) / filteredStocks.length) : 0) }}
            </p>
          </div>
          <Tag class="absolute -right-6 -bottom-6 h-32 w-32 text-white opacity-10 -rotate-12" />
        </CardContent>
      </Card>
    </div>

    <!-- Price Table -->
    <Card class="border-none shadow-sm overflow-hidden">
      <div v-if="marketplaceStore.isLoading && filteredStocks.length === 0" class="p-16 text-center">
        <div class="animate-spin h-8 w-8 border-2 border-primary border-t-transparent rounded-full mx-auto mb-3"></div>
        <p class="text-slate-400 text-sm font-medium">Memuat data harga...</p>
      </div>
      <div v-else-if="filteredStocks.length === 0" class="p-16 text-center">
        <Tag class="h-12 w-12 text-slate-200 mx-auto mb-3" />
        <p class="text-slate-400 font-bold">Produk tidak ditemukan</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-100">
              <th class="text-left px-5 py-4 text-xs font-black uppercase tracking-widest text-slate-400">Produk & SKU</th>
              <th class="text-left px-5 py-4 text-xs font-black uppercase tracking-widest text-slate-400">Harga Pabrik</th>
              <th class="text-left px-5 py-4 text-xs font-black uppercase tracking-widest text-slate-400 w-1/4">Harga Marketplace (Rp)</th>
              <th class="text-right px-5 py-4 text-xs font-black uppercase tracking-widest text-slate-400">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr
              v-for="stock in filteredStocks"
              :key="stock.id"
              class="hover:bg-primary/5/30 transition-colors group"
            >
              <td class="px-5 py-4">
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-lg overflow-hidden bg-slate-100 flex-shrink-0 border border-slate-200 flex items-center justify-center">
                    <img v-if="stock.product?.image_url" :src="stock.product.image_url" class="h-full w-full object-cover" :alt="stock.product_name" />
                    <Package v-else class="h-5 w-5 text-slate-400" />
                  </div>
                  <div>
                    <p class="font-bold text-slate-900 text-sm">{{ stock.product_name }}</p>
                    <p class="text-xs text-slate-400 font-medium mt-0.5">SKU: {{ stock.sku }}</p>
                    <span v-if="stock.product_id" class="mt-1 inline-flex items-center gap-1 text-[10px] font-black text-emerald-600 bg-emerald-50 px-1.5 py-0.5 rounded-md border border-emerald-100">
                      ✓ Terhubung ke Pabrik
                    </span>
                  </div>
                </div>
              </td>
              <!-- Factory price reference -->
              <td class="px-5 py-4">
                <div v-if="stock.product" class="space-y-0.5">
                  <p class="text-xs font-black text-slate-700">{{ formatRupiah(stock.product.marketplace_factory_price) }}</p>
                  <p class="text-[10px] text-slate-400 font-bold">Harga Beli dari Pabrik</p>
                  <p v-if="stock.price > stock.product.marketplace_factory_price" class="text-[10px] font-black text-emerald-600">
                    +{{ formatRupiah(stock.price - stock.product.marketplace_factory_price) }} estimasi margin
                  </p>
                  <p v-else-if="stock.price < stock.product.marketplace_factory_price" class="text-[10px] font-black text-red-500">
                    -{{ formatRupiah(stock.product.marketplace_factory_price - stock.price) }} rugi (jual di bawah modal)
                  </p>
                </div>
                <span v-else class="text-xs text-slate-300 font-bold">—</span>
              </td>
              <td class="px-5 py-4">
                <div class="relative max-w-[180px]">
                  <span class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 font-bold text-sm">Rp</span>
                  <input
                    type="number"
                    :value="editingPrices[stock.id] !== undefined ? editingPrices[stock.id] : stock.price"
                    @input="(e) => handlePriceInput(stock, e)"
                    class="w-full pl-9 pr-4 py-2.5 text-sm font-black border-2 rounded-xl focus:outline-none transition-all"
                    :class="editingPrices[stock.id] !== undefined ? 'bg-primary/5 border-primary/50 text-primary shadow-sm' : 'bg-slate-50 border-slate-200 text-slate-700 focus:border-primary/50 focus:bg-white'"
                  />
                </div>
              </td>
              <td class="px-5 py-4 text-right">
                <span
                  v-if="editingPrices[stock.id] !== undefined"
                  class="text-[10px] font-black uppercase tracking-wider px-2 py-1 bg-amber-100 text-primary rounded-md"
                >
                  Belum Disimpan
                </span>
                <span v-else class="text-[10px] font-black uppercase tracking-wider text-slate-300">
                  Tersimpan
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>
  </div>
</template>
