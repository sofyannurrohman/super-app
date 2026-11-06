<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMarketplaceInventoryStore } from '@/stores/marketplaceInventory'
import { useMasterDataStore } from '@/stores/masterData'
import {
  Network,
  RefreshCw,
  Search,
  Package,
  Plus,
  Trash2,
  Save,
  X,
} from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'

const marketplaceStore = useMarketplaceStore()
const marketplaceInventoryStore = useMarketplaceInventoryStore()
const masterDataStore = useMasterDataStore()

const searchKeyword = ref('')
const selectedProduct = ref<any>(null)
const mappings = ref<any[]>([])
const isMappingLoading = ref(false)
const isSaving = ref(false)

const filteredProducts = computed(() => {
  let products = marketplaceStore.marketplaceProducts || []
  if (searchKeyword.value.trim()) {
    const kw = searchKeyword.value.trim().toLowerCase()
    products = products.filter(p => {
      const name = String(p.product_name || '').toLowerCase()
      const sellerSku = String(p.seller_sku || '').toLowerCase()
      const skuId = String(p.sku_id || '').toLowerCase()
      const productId = String(p.product_id || '').toLowerCase()
      return (
        name.includes(kw) ||
        sellerSku.includes(kw) ||
        skuId.includes(kw) ||
        productId.includes(kw)
      )
    })
  }
  return products
})

async function handleSync() {
  try {
    const res = await marketplaceStore.syncMarketplaceProducts()
    toast.success(res?.message || 'Sync produk marketplace berhasil!')
  } catch (err: any) {
    const errorMsg = err.response?.data?.error || err.message || 'Gagal melakukan sinkronisasi produk'
    toast.error('Sync produk gagal: ' + errorMsg)
  }
}

async function selectProduct(product: any) {
  selectedProduct.value = product
  isMappingLoading.value = true
  try {
    const res = await marketplaceStore.fetchProductMappings(product.sku_id)
    mappings.value = res.map((m: any) => ({
      type: m.marketplace_inventory_id ? 'ATK' : 'STOCK',
      marketplace_stock_id: m.marketplace_stock_id || '',
      marketplace_inventory_id: m.marketplace_inventory_id || '',
      quantity: m.quantity
    }))
  } catch (e: any) {
    toast.error('Gagal memuat mapping produk')
    mappings.value = []
  } finally {
    isMappingLoading.value = false
  }
}

function addMapping() {
  mappings.value.push({
    type: 'STOCK',
    marketplace_stock_id: '',
    marketplace_inventory_id: '',
    quantity: 1
  })
}

function removeMapping(index: number) {
  mappings.value.splice(index, 1)
}

async function saveMapping() {
  if (!selectedProduct.value) return
  
  // Validate
  const invalid = mappings.value.some(m => 
    (m.type === 'STOCK' && !m.marketplace_stock_id) || 
    (m.type === 'ATK' && !m.marketplace_inventory_id) || 
    m.quantity < 1
  )
  if (invalid) {
    toast.error('Pastikan semua form mapping terisi dan kuantitas minimal 1')
    return
  }

  // Format mapping data
  const dataToSave = mappings.value.map(m => ({
    marketplace_stock_id: m.type === 'STOCK' ? m.marketplace_stock_id : null,
    marketplace_inventory_id: m.type === 'ATK' ? m.marketplace_inventory_id : null,
    quantity: m.quantity
  }))

  isSaving.value = true
  try {
    await marketplaceStore.saveProductMappings(selectedProduct.value.sku_id, dataToSave)
    toast.success('Mapping berhasil disimpan')
    selectedProduct.value = null
  } catch (e: any) {
    toast.error(e.response?.data?.error || 'Gagal menyimpan mapping')
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  marketplaceStore.fetchMarketplaceProducts()
  marketplaceStore.fetchStocks()
  marketplaceInventoryStore.fetchInventories()
})

watch(() => masterDataStore.selectedBranchId, () => {
  selectedProduct.value = null
  mappings.value = []
  marketplaceStore.fetchMarketplaceProducts()
  marketplaceStore.fetchStocks()
  marketplaceInventoryStore.fetchInventories()
})
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-500">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-indigo-500 to-blue-600 shadow-lg shadow-blue-200">
            <Network class="h-6 w-6 text-white" />
          </div>
          Product Mapping
        </h1>
        <p class="text-slate-500 text-sm font-medium mt-1">
          Petakan produk Marketplace (TikTok / Shopee) ke Stok Gudang
        </p>
      </div>
      <button
        @click="handleSync"
        :disabled="marketplaceStore.isSyncing"
        class="flex items-center gap-2 bg-gradient-to-r from-indigo-500 to-blue-600 text-white px-5 py-2.5 rounded-xl font-bold text-sm shadow-lg shadow-blue-200 hover:shadow-xl hover:shadow-blue-300 transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
      >
        <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': marketplaceStore.isSyncing }" />
        {{ marketplaceStore.isSyncing ? 'Sync Produk...' : 'Sync Produk Marketplace' }}
      </button>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Left: Product List -->
      <Card class="lg:col-span-1 shadow-sm border-slate-200">
        <CardHeader class="pb-3 border-b border-slate-100">
          <CardTitle class="text-lg">Produk Marketplace</CardTitle>
          <div class="relative mt-2">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              v-model="searchKeyword"
              type="text"
              placeholder="Cari SKU atau nama..."
              class="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition-all"
            />
          </div>
        </CardHeader>
        <CardContent class="p-0 max-h-[600px] overflow-y-auto">
          <div v-if="filteredProducts.length === 0" class="p-8 text-center text-slate-500">
            <Package class="h-12 w-12 mx-auto text-slate-300 mb-3" />
            <p>Tidak ada produk ditemukan</p>
          </div>
          <div class="divide-y divide-slate-100">
            <button
              v-for="product in filteredProducts"
              :key="product.sku_id"
              @click="selectProduct(product)"
              class="w-full text-left p-4 hover:bg-slate-50 transition-colors flex flex-col gap-1.5"
              :class="{ 'bg-indigo-50 border-l-4 border-indigo-500': selectedProduct?.sku_id === product.sku_id }"
            >
              <div class="flex items-center gap-1.5">
                <span
                  class="text-[10px] font-bold px-1.5 py-0.5 rounded border uppercase font-mono"
                  :class="product.marketplace === 'shopee'
                    ? 'bg-orange-50 text-orange-700 border-orange-200'
                    : 'bg-cyan-50 text-cyan-700 border-cyan-200'"
                >
                  {{ product.marketplace === 'shopee' ? 'SP' : 'TK' }}
                </span>
                <span class="font-semibold text-slate-900 text-sm line-clamp-1">
                  {{ product.product_name }}
                </span>
              </div>
              <div class="text-xs text-slate-500 flex justify-between w-full">
                <span>SKU: {{ product.seller_sku || '-' }}</span>
              </div>
            </button>
          </div>
        </CardContent>
      </Card>

      <!-- Right: Mapping Editor -->
      <Card class="lg:col-span-2 shadow-sm border-slate-200">
        <CardHeader class="pb-3 border-b border-slate-100">
          <CardTitle class="text-lg">Mapping Stok Gudang</CardTitle>
        </CardHeader>
        <CardContent class="p-6">
          <div v-if="!selectedProduct" class="h-full flex flex-col items-center justify-center text-slate-400 py-12">
            <Network class="h-16 w-16 mb-4 text-slate-200" />
            <p>Pilih produk di sebelah kiri untuk mengatur mapping.</p>
          </div>
          <div v-else-if="isMappingLoading" class="py-12 flex justify-center">
            <RefreshCw class="h-8 w-8 text-indigo-500 animate-spin" />
          </div>
          <div v-else class="space-y-6">
            <div class="bg-slate-50 rounded-xl p-4 border border-slate-100">
              <h3 class="font-semibold text-slate-900">{{ selectedProduct.product_name }}</h3>
              <p class="text-sm text-slate-500 mt-1">SKU: <span class="font-medium text-slate-700">{{ selectedProduct.seller_sku }}</span></p>
            </div>

            <div>
              <div class="flex justify-between items-center mb-3">
                <h4 class="font-medium text-slate-700">Daftar Potongan Stok</h4>
                <button
                  @click="addMapping"
                  class="flex items-center gap-1.5 text-sm font-medium text-indigo-600 hover:text-indigo-700 bg-indigo-50 hover:bg-indigo-100 px-3 py-1.5 rounded-lg transition-colors"
                >
                  <Plus class="h-4 w-4" /> Tambah
                </button>
              </div>

              <div v-if="mappings.length === 0" class="text-center py-8 bg-slate-50 rounded-xl border border-dashed border-slate-200">
                <p class="text-sm text-slate-500">Belum ada mapping stok. Tambahkan untuk memotong stok gudang otomatis saat packing.</p>
              </div>

              <div v-else class="space-y-3">
                <div v-for="(mapping, index) in mappings" :key="index" class="flex flex-col sm:flex-row items-start sm:items-end gap-3 p-3 bg-white border border-slate-200 rounded-xl shadow-sm">
                  <div class="w-full sm:w-40">
                    <label class="block text-xs font-medium text-slate-500 mb-1">Jenis Mapping</label>
                    <select
                      v-model="mapping.type"
                      @change="mapping.type === 'STOCK' ? mapping.marketplace_inventory_id = '' : mapping.marketplace_stock_id = ''"
                      class="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    >
                      <option value="STOCK">Produk Gudang</option>
                      <option value="ATK">Inventaris ATK</option>
                    </select>
                  </div>
                  <div class="flex-1 w-full">
                    <label class="block text-xs font-medium text-slate-500 mb-1">
                      {{ mapping.type === 'STOCK' ? 'Stok Gudang Marketplace' : 'Barang Inventaris ATK' }}
                    </label>
                    <!-- Stock Dropdown -->
                    <select
                      v-if="mapping.type === 'STOCK'"
                      v-model="mapping.marketplace_stock_id"
                      class="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    >
                      <option value="" disabled>Pilih Stok...</option>
                      <option v-for="stock in marketplaceStore.stocks" :key="stock.id" :value="stock.id">
                        {{ stock.product_name }} (SKU: {{ stock.sku }})
                      </option>
                    </select>
                    <!-- ATK Dropdown -->
                    <select
                      v-if="mapping.type === 'ATK'"
                      v-model="mapping.marketplace_inventory_id"
                      class="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    >
                      <option value="" disabled>Pilih Barang ATK...</option>
                      <option v-for="inv in marketplaceInventoryStore.inventories" :key="inv.id" :value="inv.id">
                        {{ inv.name }} (SKU: {{ inv.sku || '-' }})
                      </option>
                    </select>
                  </div>
                  <div class="w-24">
                    <label class="block text-xs font-medium text-slate-500 mb-1">Kuantitas</label>
                    <input
                      v-model.number="mapping.quantity"
                      type="number"
                      min="1"
                      class="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    />
                  </div>
                  <button
                    @click="removeMapping(index)"
                    class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors"
                    title="Hapus"
                  >
                    <Trash2 class="h-5 w-5" />
                  </button>
                </div>
              </div>
            </div>

            <div class="pt-4 flex justify-end gap-3 border-t border-slate-100">
              <button
                @click="selectedProduct = null"
                class="px-4 py-2 text-sm font-medium text-slate-600 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
              >
                Batal
              </button>
              <button
                @click="saveMapping"
                :disabled="isSaving"
                class="flex items-center gap-2 px-6 py-2 text-sm font-bold text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 shadow-md shadow-indigo-200 transition-all disabled:opacity-60"
              >
                <Save class="h-4 w-4" />
                {{ isSaving ? 'Menyimpan...' : 'Simpan Mapping' }}
              </button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
