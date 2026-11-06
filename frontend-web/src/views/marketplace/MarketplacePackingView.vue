<template>
  <div class="p-6 animate-in fade-in duration-500">
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900">{{ isMultiOrderMode ? 'Packing Multi-Item Order' : 'Order Packing & Stock Cut' }}</h1>
        <p class="text-sm text-slate-500 font-medium mt-1">{{ isMultiOrderMode ? 'Packing khusus order dengan pembelian lebih dari satu dalam satu Order ID.' : 'Verifikasi packing order dan potong stok Gudang otomatis.' }}</p>
      </div>
      <button 
        @click="loadOrders" 
        :disabled="marketplaceStore.isOrdersLoading"
        class="bg-gradient-to-r from-blue-600 to-cyan-500 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-blue-200 hover:shadow-xl transition-all duration-200 disabled:opacity-50 flex items-center gap-2 font-bold text-sm"
      >
        <span v-if="marketplaceStore.isOrdersLoading" class="animate-spin inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full"></span>
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-refresh-cw"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
        {{ marketplaceStore.isOrdersLoading ? 'Memuat...' : 'Refresh Daftar' }}
      </button>
    </div>

    <!-- Main Grid Layout -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Left Section: Data Orderan (col-span-2) -->
      <div class="lg:col-span-2 space-y-4">
        <div class="bg-white shadow-xl shadow-slate-200/50 rounded-2xl p-5 border border-slate-100 overflow-hidden">
          <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-4 border-b border-slate-100 pb-4">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-indigo-50 text-indigo-600 rounded-lg">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-package-check"><path d="m16 16 2 2 4-4"/><path d="M21 10V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l2-1.14"/><path d="m7.5 4.27 9 5.15"/><polyline points="3.29 7 12 12 20.71 7"/><line x1="12" x2="12" y1="22" y2="12"/></svg>
              </div>
              <h2 class="text-lg font-bold text-slate-800">{{ isMultiOrderMode ? 'Daftar Order Multi-Item Siap Packing' : 'Daftar Order Siap Packing' }}</h2>
            </div>
            
            <div class="w-full sm:w-auto flex flex-wrap gap-2">
              <select v-model="selectedMarketplace" class="w-full sm:w-auto px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-sm transition-all bg-white text-slate-700">
                <option value="">Semua Marketplace</option>
                <option value="tiktok">TikTok Shop</option>
                <option value="shopee">Shopee</option>
              </select>
              <select v-model="selectedWarehouse" class="w-full sm:w-auto px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-sm transition-all bg-white text-slate-700">
                <option value="">Pilih Warehouse untuk mulai scan</option>
                <option v-for="wh in uniqueWarehouses" :key="wh" :value="wh">{{ wh }}</option>
              </select>
              <select v-model="selectedSku" :disabled="!selectedWarehouse" class="w-full sm:w-auto px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-sm transition-all bg-white text-slate-700 max-w-[200px] disabled:opacity-50 disabled:bg-slate-50 disabled:cursor-not-allowed">
                <option value="">{{ !selectedWarehouse ? 'Pilih Warehouse terlebih dahulu' : (isMultiOrderMode ? 'Semua SKU (Multi-Order)' : 'Pilih SKU untuk mulai scan') }}</option>
                <option v-for="sku in uniqueSkus" :key="sku" :value="sku" :title="sku">{{ sku.length > 30 ? sku.substring(0, 30) + '...' : sku }}</option>
              </select>
              <select v-model="selectedStatus" class="w-full sm:w-auto px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-sm transition-all bg-white text-slate-700">
                <option value="">Menunggu Packing</option>
                <option value="PACKED">Sudah Dipacking (Sesi ini)</option>
                <option value="ALL">Semua Status (Sesi ini)</option>
              </select>
            </div>
          </div>
          
          <div class="overflow-x-auto w-full rounded-xl border border-slate-200">
            <table class="w-full text-left border-collapse whitespace-nowrap text-sm">
              <thead>
                <tr class="bg-slate-50 border-b border-slate-200">
                  <th class="p-3 font-semibold text-slate-600 sticky left-0 bg-slate-50 z-10 shadow-[1px_0_0_0_#e2e8f0]">Status Scan</th>
                  <th class="p-3 font-semibold text-slate-600">Order ID</th>
                  <th class="p-3 font-semibold text-slate-600">Dipacking Oleh</th>
                  <th class="p-3 font-semibold text-slate-600">Warehouse</th>
                  <th class="p-3 font-semibold text-slate-600">Seller SKU</th>
                  <th class="p-3 font-semibold text-slate-600">Product Name</th>
                  <th class="p-3 font-semibold text-slate-600">Tracking ID</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="marketplaceStore.isOrdersLoading">
                  <td colspan="7" class="text-center p-8 text-slate-500 font-medium">Memuat data order...</td>
                </tr>
                <tr v-else-if="filteredPendingOrders.length === 0">
                  <td colspan="7" class="text-center p-8 text-slate-500 font-medium">
                    {{ pendingOrders.length === 0 ? 'Semua pesanan sudah di-packing.' : 'Tidak ada pesanan yang sesuai dengan filter.' }}
                  </td>
                </tr>
                <tr 
                  v-for="order in filteredPendingOrders" 
                  :key="order.order_id" 
                  :id="'row-' + order.order_id"
                  class="border-b border-slate-100 transition-colors"
                  :class="[
                    order.order_status === 'PACKED' ? 'bg-emerald-50' : 'hover:bg-slate-50'
                  ]"
                >
                  <td class="p-3 sticky left-0 z-10 shadow-[1px_0_0_0_#f1f5f9]" :class="order.order_status === 'PACKED' ? 'bg-emerald-50' : 'bg-white'">
                    <span v-if="order.order_status === 'PACKED'" class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md bg-emerald-100 text-emerald-700 text-xs font-bold border border-emerald-200">
                      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-check"><path d="M20 6 9 17l-5-5"/></svg>
                      Packed
                    </span>
                    <span v-else class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md bg-amber-100 text-amber-700 text-xs font-semibold border border-amber-200">
                      Menunggu Packing
                    </span>
                  </td>
                  <td class="p-3 font-bold text-slate-700">
                    <div class="flex items-center gap-1.5">
                      <span 
                        class="text-[10px] font-extrabold px-1.5 py-0.5 rounded font-mono"
                        :class="order.marketplace === 'shopee' ? 'bg-orange-100 text-orange-700 border border-orange-200' : 'bg-slate-900 text-white'"
                      >
                        {{ order.marketplace === 'shopee' ? 'SP' : 'TT' }}
                      </span>
                      <span>{{ order.order_id }}</span>
                    </div>
                  </td>
                  <td class="p-3 font-semibold text-slate-700">{{ order.packed_by || '-' }}</td>
                  <td class="p-3">
                    <span class="inline-flex items-center gap-1.5 px-2 py-1 rounded-md bg-amber-50 text-amber-700 text-xs font-semibold border border-amber-200">
                      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-warehouse"><path d="M22 8.35V20a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V8.35A2 2 0 0 1 3.26 6.5l8-3.2a2 2 0 0 1 1.48 0l8 3.2A2 2 0 0 1 22 8.35Z"/><path d="M6 18h12"/><path d="M6 14h12"/><rect width="12" height="12" x="6" y="10"/></svg>
                      {{ order.warehouse_name || '-' }}
                    </span>
                  </td>
                  <td class="p-3">
                    <div v-for="item in order.items" :key="item.id" class="font-bold text-blue-600 whitespace-nowrap">
                      {{ item.seller_sku }} <span class="text-xs text-blue-400 font-normal ml-1">x{{ item.item_count }}</span>
                    </div>
                  </td>
                  <td class="p-3 text-slate-600">
                    <div v-for="item in order.items" :key="item.id" class="truncate max-w-[200px]" :title="item.product_name">
                      {{ item.product_name }}
                    </div>
                  </td>
                  <td class="p-3 font-mono text-slate-500 text-xs">{{ order.tracking_id || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Right Section: Scanner (col-span-1) -->
      <div class="lg:col-span-1">
        <div class="bg-gradient-to-b from-slate-900 to-slate-800 shadow-2xl rounded-2xl p-6 border border-slate-700 sticky top-6">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <!-- Dynamic Icon based on scanStatus -->
              <div v-if="scanStatus === 'idle'" class="p-2 bg-slate-800 text-slate-400 rounded-xl transition-all duration-300">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-box"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>
              </div>
              <div v-else-if="scanStatus === 'success'" class="p-2 bg-emerald-500/20 text-emerald-400 rounded-xl transition-all duration-300 scale-110">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-check-circle-2"><circle cx="12" cy="12" r="10"/><path d="m9 12 2 2 4-4"/></svg>
              </div>
              <div v-else-if="scanStatus === 'warning'" class="p-2 bg-amber-500/20 text-amber-400 rounded-xl transition-all duration-300 scale-110">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-alert-triangle"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
              </div>
              <div v-else-if="scanStatus === 'error'" class="p-2 bg-red-500/20 text-red-400 rounded-xl transition-all duration-300 scale-110">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-x-circle"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>
              </div>
              <div>
                <h2 class="text-xl font-bold text-white transition-colors duration-300"
                  :class="{
                    'text-emerald-400': scanStatus === 'success',
                    'text-amber-400': scanStatus === 'warning',
                    'text-red-400': scanStatus === 'error'
                  }">Scanner Packing</h2>
                <p class="text-slate-400 text-xs mt-0.5">Fokus kursor disini untuk memindai</p>
              </div>
            </div>
            
            <!-- Toggle Camera Button -->
            <button
              @click="toggleCamera"
              class="p-2 rounded-xl border transition-all"
              :class="isCameraActive ? 'bg-indigo-500/20 border-indigo-500/50 text-indigo-400' : 'bg-slate-800 border-slate-700 text-slate-400 hover:bg-slate-700'"
              title="Gunakan Kamera Scanner"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-camera"><path d="M14.5 4h-5L7 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-3l-2.5-3z"/><circle cx="12" cy="13" r="3"/></svg>
            </button>
          </div>

          <div class="space-y-4">
            <div class="rounded-xl border border-indigo-400/30 bg-indigo-500/10 px-3.5 py-3 text-xs text-indigo-100">
              <p class="font-bold text-indigo-200">Urutan kerja packing</p>
              <ol class="mt-1 list-decimal list-inside space-y-0.5 text-indigo-100/90">
                <li>Pilih <strong>Warehouse</strong> di filter (Wajib)</li>
                <li v-if="isMultiOrderMode">Pilih <strong>SKU</strong> <span class="text-indigo-200/70">(Opsional / default Semua SKU)</span></li>
                <li v-else>Pilih <strong>SKU</strong> yang akan dipack (Wajib)</li>
                <li>Scan barcode order / resi untuk Warehouse tersebut</li>
              </ol>
              <div class="mt-2.5 pt-2 border-t border-indigo-400/20 flex flex-wrap gap-2 text-[11px]">
                <span class="px-2 py-0.5 rounded-md font-semibold flex items-center gap-1"
                  :class="selectedWarehouse ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' : 'bg-amber-500/20 text-amber-300 border border-amber-500/30'">
                  <span class="w-1.5 h-1.5 rounded-full" :class="selectedWarehouse ? 'bg-emerald-400' : 'bg-amber-400'"></span>
                  Gudang: {{ selectedWarehouse || 'Belum dipilih' }}
                </span>
                <span class="px-2 py-0.5 rounded-md font-semibold flex items-center gap-1"
                  :class="(selectedSku || isMultiOrderMode) && selectedWarehouse ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' : 'bg-amber-500/20 text-amber-300 border border-amber-500/30'">
                  <span class="w-1.5 h-1.5 rounded-full" :class="(selectedSku || isMultiOrderMode) && selectedWarehouse ? 'bg-emerald-400' : 'bg-amber-400'"></span>
                  SKU: {{ selectedSku ? (selectedSku.length > 20 ? selectedSku.substring(0, 20) + '...' : selectedSku) : (isMultiOrderMode ? 'Semua SKU (Multi)' : 'Belum dipilih') }}
                </span>
              </div>
              <p v-if="!selectedWarehouse" class="mt-2 font-bold text-amber-300">Pilih Warehouse terlebih dahulu. Scanner belum aktif.</p>
              <p v-else-if="!isMultiOrderMode && !selectedSku" class="mt-2 font-bold text-amber-300">Pilih SKU terlebih dahulu. Scanner belum aktif.</p>
              <p v-else class="mt-2 font-bold text-emerald-300">Scanner aktif & siap memindai order.</p>
            </div>
            <!-- Camera Scanner Container -->
            <div v-if="isCameraActive" class="w-full bg-slate-900 rounded-xl overflow-hidden border-2 border-slate-700 mb-4">
              <div id="reader" class="w-full min-h-[300px]"></div>
            </div>

            <div class="relative group">
              <input
                ref="scannerInputRef"
                v-model="scanInput"
                @keyup.enter="handleScan"
                :disabled="isScannerDisabled"
                type="text"
                :placeholder="!selectedWarehouse ? 'Pilih Warehouse terlebih dahulu...' : (!isMultiOrderMode && !selectedSku ? 'Pilih SKU terlebih dahulu...' : 'Scan Order ID / Tracking ID...')"
                class="w-full rounded-xl px-4 py-4 focus:outline-none transition-all duration-300 font-mono text-lg shadow-inner disabled:cursor-not-allowed disabled:opacity-50"
                :class="{
                  'bg-emerald-950/40 border-2 border-emerald-500 ring-4 ring-emerald-500/30 text-emerald-300 scale-[1.02]': scanStatus === 'success',
                  'bg-amber-950/40 border-2 border-amber-500 ring-4 ring-amber-500/30 text-amber-300': scanStatus === 'warning',
                  'bg-red-950/40 border-2 border-red-500 ring-4 ring-red-500/30 text-red-300': scanStatus === 'error',
                  'bg-slate-950/50 border-2 border-slate-700 text-slate-200 placeholder:text-slate-600 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20': scanStatus === 'idle'
                }"
              />
              <div class="absolute inset-y-0 right-4 flex items-center pointer-events-none">
                <div v-if="isProcessingScan" class="w-4 h-4 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin"></div>
                <div v-else class="w-2 h-2 rounded-full animate-pulse bg-emerald-500"></div>
              </div>
            </div>
            <p class="text-xs text-slate-400 text-center">Scan order yang memuat Warehouse & SKU aktif. Jika barcode bukan order Warehouse/SKU aktif, scan ditolak dan order tidak berubah.</p>
          </div>

          <!-- Quick Stats -->
          <div class="mt-8 pt-6 border-t border-slate-700/50 grid grid-cols-2 gap-4">
            <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
              <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider">Packed (Hari ini)</p>
              <p class="text-3xl font-black text-emerald-400 mt-1">{{ todayPackedCount }}</p>
            </div>
            <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
              <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider">Sisa</p>
              <p class="text-3xl font-black text-slate-300 mt-1">{{ pendingOrders.length }}</p>
            </div>
          </div>
        </div>
      </div>

    </div>
  </div>

  <!-- Multi-SKU Pack Modal -->
  <MultiSkuPackModal
    v-if="showMultiSkuModal && multiSkuModalOrder"
    :order="multiSkuModalOrder!"
    :stocks="marketplaceStore.stocks"
    @confirm="onModalConfirm"
    @cancel="onModalCancel"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import type { MarketplaceOrderGroup, PackOrderPayload } from '@/api/marketplace'
import MultiSkuPackModal from './MultiSkuPackModal.vue'
import { Html5QrcodeScanner } from 'html5-qrcode'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()
const authStore = useAuthStore()
const route = useRoute()
const isMultiOrderMode = computed(() => route.name === 'marketplacePackingMulti')
const isMultiItemOrder = (order: MarketplaceOrderGroup) =>
  order.items.length > 1 || order.items.some(item => Number(item.item_count || 0) > 1)
let eventSource: EventSource | null = null

const isCameraActive = ref(false)
let html5QrcodeScanner: Html5QrcodeScanner | null = null

const toggleCamera = () => {
  isCameraActive.value = !isCameraActive.value
  
  if (isCameraActive.value) {
    nextTick(() => {
      try {
        html5QrcodeScanner = new Html5QrcodeScanner(
          "reader",
          { fps: 10, qrbox: {width: 250, height: 250} },
          false
        )
        html5QrcodeScanner.render(onScanSuccess, (err) => {
          // ignore continuous scan failures (no qr found)
        })
      } catch (err: any) {
        toast.error(`Gagal memulai scanner: ${err.message || err}`)
        isCameraActive.value = false
      }
    })
  } else {
    if (html5QrcodeScanner) {
      html5QrcodeScanner.clear().catch(e => console.error(e))
      html5QrcodeScanner = null
    }
    nextTick(() => {
      focusScannerIfManual()
    })
  }
}

const onScanSuccess = async (decodedText: string) => {
  if (isProcessingScan.value) return

  scanInput.value = decodedText
  
  if (html5QrcodeScanner) {
    html5QrcodeScanner.pause(true)
  }

  await handleScan()
  
  if (html5QrcodeScanner && isCameraActive.value) {
    setTimeout(() => {
      try {
        html5QrcodeScanner?.resume()
      } catch (e) {
        console.error(e)
      }
    }, 1500)
  }
}

const loadOrders = async () => {
  await marketplaceStore.fetchMarketplaceOrders({
    status: 'VERIFIED',
    page: 1,
    pageSize: 5000
  })
}

const selectedMarketplace = ref('')
const selectedWarehouse = ref('')
const selectedSku = ref('')
const selectedStatus = ref('')

const uniqueWarehouses = computed(() => {
  const orders = marketplaceStore.marketplaceOrders.filter(o =>
    isMultiOrderMode.value ? isMultiItemOrder(o) : !isMultiItemOrder(o)
  )
  return [...new Set(orders.map(o => o.warehouse_name).filter(Boolean))].sort()
})

const uniqueSkus = computed(() => {
  let orders = marketplaceStore.marketplaceOrders.filter(o =>
    isMultiOrderMode.value ? isMultiItemOrder(o) : !isMultiItemOrder(o)
  )
  if (selectedWarehouse.value) {
    orders = orders.filter(o => o.warehouse_name === selectedWarehouse.value)
  }
  const skus = orders.flatMap(o => o.items.map(i => i.seller_sku || i.product_name))
  return [...new Set(skus.filter(Boolean))].sort()
})

// Auto-reset SKU jika warehouse berubah dan SKU sebelumnya tidak ada di gudang baru
watch(selectedWarehouse, () => {
  if (selectedSku.value && !uniqueSkus.value.includes(selectedSku.value)) {
    selectedSku.value = ''
  }
  nextTick(() => {
    focusScannerIfManual()
  })
})

const pendingOrders = computed(() => {
  return marketplaceStore.marketplaceOrders.filter(o => o.order_status !== 'PACKED' && (isMultiOrderMode.value ? isMultiItemOrder(o) : !isMultiItemOrder(o)))
})

const filteredPendingOrders = computed(() => {
  return marketplaceStore.marketplaceOrders.filter(order => {
    if (isMultiOrderMode.value && !isMultiItemOrder(order)) return false
    if (!isMultiOrderMode.value && isMultiItemOrder(order)) return false
    let matchStatus = true;
    
    // Status packed
    const isPacked = order.order_status === 'PACKED';

    if (selectedStatus.value === '') {
      matchStatus = !isPacked; // Menunggu Packing
    } else if (selectedStatus.value === 'PACKED') {
      matchStatus = isPacked; // Sudah Dipacking
    } else if (selectedStatus.value === 'ALL') {
      matchStatus = true; // Semua
    }

    const matchMarketplace = selectedMarketplace.value ? order.marketplace === selectedMarketplace.value : true;
    const matchWarehouse = selectedWarehouse.value ? order.warehouse_name === selectedWarehouse.value : true;
    const matchSku = selectedSku.value ? order.items.some(i => (i.seller_sku || i.product_name) === selectedSku.value) : true;
    
    return matchStatus && matchMarketplace && matchWarehouse && matchSku;
  })
})

const scanInput = ref('')
const scannerInputRef = ref<HTMLInputElement | null>(null)
const isProcessingScan = ref(false)
const scanStatus = ref<'idle' | 'success' | 'warning' | 'error'>('idle')
let scanStatusTimeout: number | null = null

const isScannerDisabled = computed(() => {
  if (isProcessingScan.value) return true
  if (!selectedWarehouse.value) return true
  if (!isMultiOrderMode.value && !selectedSku.value) return true
  return false
})

// Multi-SKU Modal state
const showMultiSkuModal = ref(false)
const multiSkuModalOrder = ref<MarketplaceOrderGroup | null>(null)

const setScanStatus = (status: 'success' | 'warning' | 'error') => {
  scanStatus.value = status
  if (scanStatusTimeout) clearTimeout(scanStatusTimeout)
  scanStatusTimeout = window.setTimeout(() => {
    scanStatus.value = 'idle'
  }, 1000)
}

const todayPackedCount = ref(0)

const fetchTodayProgress = async () => {
  try {
    const today = new Date().toISOString().split('T')[0]
    const bid = masterDataStore.selectedBranchId
    const branchFilter = bid && bid !== 'ALL_BRANCHES' ? `&branch_id=${bid}` : ''
    const res = await apiClient.get(`/marketplace/scan-history?action=PACK&from=${today}&to=${today}${branchFilter}`)
    todayPackedCount.value = res.data.data ? res.data.data.length : 0
  } catch (error) {
    console.error('Failed to fetch today progress:', error)
  }
}

function playSuccessSound() {
  try {
    const AudioContext = window.AudioContext || (window as any).webkitAudioContext
    if (!AudioContext) return
    const ctx = new AudioContext()
    
    // Short high beep 1
    const osc1 = ctx.createOscillator()
    const gain1 = ctx.createGain()
    osc1.type = 'sine'
    osc1.frequency.setValueAtTime(600, ctx.currentTime)
    gain1.gain.setValueAtTime(0.1, ctx.currentTime)
    osc1.connect(gain1)
    gain1.connect(ctx.destination)
    osc1.start(ctx.currentTime)
    osc1.stop(ctx.currentTime + 0.1)

    // Short high beep 2 (higher)
    const osc2 = ctx.createOscillator()
    const gain2 = ctx.createGain()
    osc2.type = 'sine'
    osc2.frequency.setValueAtTime(800, ctx.currentTime + 0.1)
    gain2.gain.setValueAtTime(0.1, ctx.currentTime + 0.1)
    osc2.connect(gain2)
    gain2.connect(ctx.destination)
    osc2.start(ctx.currentTime + 0.1)
    osc2.stop(ctx.currentTime + 0.25)
  } catch (e) {
    console.error('Audio playback failed', e)
  }
}

function playAlreadyScannedSound() {
  try {
    const AudioContext = window.AudioContext || (window as any).webkitAudioContext
    if (!AudioContext) return
    const ctx = new AudioContext()
    
    // First bloop
    const osc1 = ctx.createOscillator()
    const gain1 = ctx.createGain()
    osc1.type = 'square'
    osc1.frequency.setValueAtTime(400, ctx.currentTime)
    gain1.gain.setValueAtTime(0.05, ctx.currentTime)
    osc1.connect(gain1)
    gain1.connect(ctx.destination)
    osc1.start(ctx.currentTime)
    osc1.stop(ctx.currentTime + 0.1)

    // Second bloop (lower)
    const osc2 = ctx.createOscillator()
    const gain2 = ctx.createGain()
    osc2.type = 'square'
    osc2.frequency.setValueAtTime(300, ctx.currentTime + 0.15)
    gain2.gain.setValueAtTime(0.05, ctx.currentTime + 0.15)
    osc2.connect(gain2)
    gain2.connect(ctx.destination)
    osc2.start(ctx.currentTime + 0.15)
    osc2.stop(ctx.currentTime + 0.3)
  } catch (e) {
    console.error('Audio playback failed', e)
  }
}

function playErrorSound() {
  try {
    const AudioContext = window.AudioContext || (window as any).webkitAudioContext
    if (!AudioContext) return
    const ctx = new AudioContext()
    const osc = ctx.createOscillator()
    const gainNode = ctx.createGain()
    osc.type = 'sawtooth'
    osc.frequency.setValueAtTime(200, ctx.currentTime)
    gainNode.gain.setValueAtTime(0.1, ctx.currentTime) 
    osc.connect(gainNode)
    gainNode.connect(ctx.destination)
    osc.start()
    osc.stop(ctx.currentTime + 0.3)
  } catch (e) {
    console.error('Audio playback failed', e)
  }
}

function focusScannerIfManual() {
  if (scannerInputRef.value && !isCameraActive.value) {
    scannerInputRef.value.focus()
  }
}

const clearScanInput = () => {
  scanInput.value = ''
  nextTick(() => {
    focusScannerIfManual()
  })
}

const execPack = async (orderId: string, payload?: PackOrderPayload) => {
  isProcessingScan.value = true
  try {
    await marketplaceStore.packMarketplaceOrder(orderId, payload)
    setScanStatus('success')
    playSuccessSound()
    toast.success(`Order ${orderId} berhasil dipacking dan stok dipotong!`)

    // Update local state di grouped orders
    const order = marketplaceStore.marketplaceOrders.find(o => o.order_id === orderId)
    if (order) {
      order.order_status = 'PACKED'
      order.packed_by = authStore.user?.username || authStore.user?.name || 'Anda'
    }
    todayPackedCount.value++

    nextTick(() => {
      const row = document.getElementById(`row-${orderId}`)
      if (row) {
        row.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    })
  } catch (e: any) {
    setScanStatus('error')
    playErrorSound()
    const errMsg = e.response?.data?.error || e.message || 'Gagal memproses packing'
    toast.error(`Gagal: ${errMsg}`)
  } finally {
    isProcessingScan.value = false
    nextTick(() => {
      focusScannerIfManual()
    })
  }
}

const handleScan = async () => {
  const code = scanInput.value.trim()
  if (!code) return

  if (!selectedWarehouse.value) {
    setScanStatus('warning')
    playAlreadyScannedSound()
    toast.info('Pilih Warehouse pada filter sebelum melakukan scan.')
    clearScanInput()
    return
  }

  if (!isMultiOrderMode.value && !selectedSku.value) {
    setScanStatus('warning')
    playAlreadyScannedSound()
    toast.info('Pilih SKU pada filter sebelum melakukan scan.')
    clearScanInput()
    return
  }

  // Cari di grouped orders berdasarkan order_id atau tracking_id
  let matchedOrder = marketplaceStore.marketplaceOrders.find(
    o => o.order_id === code || o.tracking_id === code
  )

  if (!matchedOrder) {
    // Coba cari langsung ke backend (misal statusnya IN_TRANSIT atau belum termuat di filter VERIFIED)
    try {
      const searchRes = await apiClient.get('/marketplace/orders', {
        params: { keyword: code, pageSize: 5 }
      })
      const foundList = searchRes.data?.data || []
      const found = foundList.find((o: any) => o.order_id === code || o.tracking_id === code)
      if (found && found.order_status !== 'PACKED') {
        matchedOrder = found
        marketplaceStore.marketplaceOrders.unshift(found)
        toast.info(`Order ${found.order_id} ditemukan di sistem (Status: ${found.order_status}). Memproses packing...`)
      }
    } catch (e) {
      // ignore
    }
  }

  if (!matchedOrder) {
    setScanStatus('error')
    playErrorSound()
    toast.error(`Data order dengan resi/ID ${code} tidak ditemukan atau sudah di-packing!`)
    clearScanInput()
    return
  }

  if (matchedOrder.order_status === 'PACKED') {
    setScanStatus('warning')
    playAlreadyScannedSound()
    toast.info(`Semua pesanan untuk resi/order ${code} sudah di-pack sebelumnya.`)
    clearScanInput()
    return
  }

  if (isMultiOrderMode.value && !isMultiItemOrder(matchedOrder)) {
    setScanStatus('warning')
    playAlreadyScannedSound()
    toast.info('Menu ini khusus untuk order dengan pembelian lebih dari satu.')
    clearScanInput()
    return
  }

  if (!isMultiOrderMode.value && isMultiItemOrder(matchedOrder)) {
    setScanStatus('warning')
    playAlreadyScannedSound()
    toast.warning('Order ini adalah Multi-Item Order. Silakan berpindah ke menu Packing Multi-Order.')
    clearScanInput()
    return
  }

  // Validasi Warehouse
  if (matchedOrder.warehouse_name && matchedOrder.warehouse_name !== selectedWarehouse.value) {
    setScanStatus('warning')
    playAlreadyScannedSound()
    toast.warning(`Order ${matchedOrder.order_id} terdaftar di Warehouse "${matchedOrder.warehouse_name}", bukan "${selectedWarehouse.value}".`)
    clearScanInput()
    return
  }

  // Validasi SKU (jika selectedSku diisi)
  if (selectedSku.value) {
    const matchesSelectedSku = matchedOrder.items.some(
      item => (item.seller_sku || item.product_name) === selectedSku.value
    )
    if (!matchesSelectedSku) {
      setScanStatus('warning')
      playAlreadyScannedSound()
      toast.warning(`Order ${matchedOrder.order_id} bukan SKU "${selectedSku.value}". Pilih order SKU yang sesuai.`)
      clearScanInput()
      return
    }
  }

  // Deteksi: multi-line item ATAU ada item dengan qty > 1
  const isMultiItem = matchedOrder.items.length > 1 ||
                      matchedOrder.items.some(i => i.item_count > 1)

  if (isMultiItem) {
    // Buka modal untuk konfirmasi + pilih kardus
    multiSkuModalOrder.value = matchedOrder
    showMultiSkuModal.value = true
    clearScanInput()
    return
  }

  // Single item, qty 1 — pack langsung
  clearScanInput()
  await execPack(matchedOrder.order_id)
}

// Callback dari modal saat admin klik PACK SEKARANG
const onModalConfirm = async (payload: { box_stock_id?: string; box_qty?: number; additional_items?: { stock_id: string; qty: number }[] }) => {
  if (!multiSkuModalOrder.value) return
  const orderId = multiSkuModalOrder.value.order_id
  showMultiSkuModal.value = false
  multiSkuModalOrder.value = null
  await execPack(orderId, {
    ...payload,
    skip_default_box: true,
  })
  nextTick(() => focusScannerIfManual())
}

const onModalCancel = () => {
  showMultiSkuModal.value = false
  multiSkuModalOrder.value = null
  nextTick(() => focusScannerIfManual())
}

onMounted(() => {
  loadOrders()
  fetchTodayProgress()
  // Load stok marketplace untuk dropdown kardus di modal
  marketplaceStore.fetchStocks()

  focusScannerIfManual()

  // Set up Server-Sent Events (SSE) for real-time updates
  const apiUrl = apiClient.defaults.baseURL || import.meta.env.VITE_API_URL || ''
  eventSource = new EventSource(`${apiUrl}/marketplace/stream`)
  
  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.status === 'PACKED') {
        const order = marketplaceStore.marketplaceOrders.find(o => o.order_id === data.order_id)
        if (order) {
          if (order.order_status !== 'PACKED') {
            order.order_status = 'PACKED'
            order.packed_by = data.actor
            todayPackedCount.value++
            
            // Show toast if it wasn't scanned by this device (this device is currently not processing it)
            if (!isProcessingScan.value) {
              toast.info(`Pesanan ${data.order_id} telah dipacking oleh ${data.actor}`)
            }
          }
        }
      } else if (data.status === 'VERIFIED') {
        // Automatically fetch new order when it's verified in Picking menu
        loadOrders()
        toast.info(`Pesanan ${data.order_id} baru saja diverifikasi dan siap dipacking.`)
      }
    } catch (e) {
      console.error('SSE Error processing message', e)
    }
  }
})

onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
  }
  if (html5QrcodeScanner) {
    html5QrcodeScanner.clear().catch(e => console.error(e))
  }
})

watch(() => masterDataStore.selectedBranchId, () => {
  loadOrders()
  fetchTodayProgress()
  marketplaceStore.fetchStocks()
})
</script>

<style scoped>
/* Target the internal elements generated by html5-qrcode */
#reader :deep(span),
#reader :deep(select),
#reader :deep(button) {
  color: white !important;
}

#reader :deep(select) {
  background-color: #1e293b !important; /* slate-800 */
  border: 1px solid #334155 !important; /* slate-700 */
  padding: 4px 8px;
  border-radius: 6px;
  margin: 4px 0;
}

#reader :deep(button) {
  background-color: #3b82f6 !important; /* blue-500 */
  border: none !important;
  padding: 6px 12px !important;
  border-radius: 6px !important;
  margin: 4px !important;
  font-weight: 500 !important;
  cursor: pointer;
  transition: all 0.2s;
}

#reader :deep(button:hover) {
  background-color: #2563eb !important; /* blue-600 */
}

#reader :deep(a) {
  color: #60a5fa !important; /* blue-400 */
  text-decoration: none !important;
}

#reader :deep(a:hover) {
  text-decoration: underline !important;
}
</style>
