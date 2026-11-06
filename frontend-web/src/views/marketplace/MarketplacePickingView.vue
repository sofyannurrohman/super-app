<template>
  <div class="p-6 animate-in fade-in duration-500">
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900">Order Picking & Verification</h1>
        <p class="text-sm text-slate-500 font-medium mt-1">Verifikasi order dengan memindai resi menggunakan scanner.</p>
      </div>
      <button 
        @click="syncTikTokOrders" 
        :disabled="syncing"
        class="bg-gradient-to-r from-blue-600 to-cyan-500 text-white px-5 py-2.5 rounded-xl shadow-lg shadow-blue-200 hover:shadow-xl transition-all duration-200 disabled:opacity-50 flex items-center gap-2 font-bold text-sm"
      >
        <span v-if="syncing" class="animate-spin inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full"></span>
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-refresh-cw"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
        {{ syncing ? 'Menarik...' : 'Tarik Pesanan TikTok' }}
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
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-list-ordered"><line x1="10" x2="21" y1="6" y2="6"/><line x1="10" x2="21" y1="12" y2="12"/><line x1="10" x2="21" y1="18" y2="18"/><path d="M4 6h1v4"/><path d="M4 10h2"/><path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1"/></svg>
              </div>
              <h2 class="text-lg font-bold text-slate-800">Daftar Order</h2>
            </div>
            
            <div class="w-full sm:w-auto flex flex-wrap gap-2">
              <select v-model="selectedProduct" class="w-full sm:w-auto px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-sm transition-all bg-white text-slate-700 max-w-[200px]">
                <option value="">Semua Product</option>
                <option v-for="product in uniqueProducts" :key="product" :value="product" :title="product">{{ product.length > 30 ? product.substring(0, 30) + '...' : product }}</option>
              </select>
              <select v-model="selectedWarehouse" class="w-full sm:w-auto px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-sm transition-all bg-white text-slate-700">
                <option value="">Semua Warehouse</option>
                <option v-for="wh in uniqueWarehouses" :key="wh" :value="wh">{{ wh }}</option>
              </select>
              <select v-model="selectedStatus" class="w-full sm:w-auto px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-sm transition-all bg-white text-slate-700">
                <option value="">Menunggu Picking</option>
                <option value="VERIFIED">Sudah Dipicking (Sesi ini)</option>
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
                  <th class="p-3 font-semibold text-slate-600">Order Status</th>
                  <th class="p-3 font-semibold text-slate-600">Dipicking Oleh</th>
                  <th class="p-3 font-semibold text-slate-600">Warehouse</th>
                  <th class="p-3 font-semibold text-slate-600">Seller SKU</th>
                  <th class="p-3 font-semibold text-slate-600">Product Name</th>
                  <th class="p-3 font-semibold text-slate-600">Tracking ID</th>
                </tr>
              </thead>
              <TransitionGroup name="list" tag="tbody" class="divide-y divide-slate-100">
                <tr v-if="loading">
                  <td colspan="8" class="text-center p-8 text-slate-500 font-medium">Memuat data order...</td>
                </tr>
                <tr v-else-if="orders.length === 0">
                  <td colspan="8" class="text-center p-8 text-slate-500 font-medium">Tidak ada pesanan yang perlu di-picking.</td>
                </tr>
                <tr 
                  v-for="order in paginatedOrders" 
                  :key="order.order_id" 
                  :id="'row-' + order.order_id"
                  class="border-b border-slate-100 transition-all duration-300"
                  :class="[
                    (order.order_status === 'VERIFIED' || recentlyVerified.includes(order.order_id)) ? 'bg-emerald-50' : 'hover:bg-slate-50'
                  ]"
                >
                  <td class="p-3 sticky left-0 z-10 shadow-[1px_0_0_0_#f1f5f9]" :class="(order.order_status === 'VERIFIED' || recentlyVerified.includes(order.order_id)) ? 'bg-emerald-50' : 'bg-white'">
                    <span v-if="order.order_status === 'VERIFIED' || recentlyVerified.includes(order.order_id)" class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md bg-emerald-100 text-emerald-700 text-xs font-bold border border-emerald-200">
                      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-check"><path d="M20 6 9 17l-5-5"/></svg>
                      Verified
                    </span>
                    <span v-else class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md bg-slate-100 text-slate-600 text-xs font-semibold border border-slate-200">
                      Pending
                    </span>
                  </td>
                  <td class="p-3 font-bold text-slate-700">{{ order.order_id }}</td>
                  <td class="p-3">
                    <span class="px-2 py-1 bg-indigo-50 text-indigo-700 text-xs font-bold rounded border border-indigo-100">{{ order.order_status }}</span>
                  </td>
                  <td class="p-3 font-semibold text-slate-700">{{ order.picked_by || '-' }}</td>
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
              </TransitionGroup>
            </table>
          </div>
          
          <div class="flex justify-between items-center mt-6">
            <button 
              @click="prevPage" 
              :disabled="currentPage === 1" 
              class="px-4 py-2 border border-slate-200 rounded-lg bg-white hover:bg-slate-50 disabled:opacity-50 transition text-sm font-semibold text-slate-700 shadow-sm"
            >
              Sebelumnya
            </button>
            <span class="text-sm font-medium text-slate-500 bg-slate-100 px-3 py-1 rounded-full">
              Halaman {{ currentPage }} dari {{ totalPages || 1 }}
            </span>
            <button 
              @click="nextPage" 
              :disabled="currentPage === totalPages || totalPages === 0" 
              class="px-4 py-2 border border-slate-200 rounded-lg bg-white hover:bg-slate-50 disabled:opacity-50 transition text-sm font-semibold text-slate-700 shadow-sm"
            >
              Selanjutnya
            </button>
          </div>
        </div>
      </div>

      <!-- Right Section: Scanner (col-span-1) -->
      <div class="lg:col-span-1">
        <div class="bg-gradient-to-b from-slate-900 to-slate-800 shadow-2xl rounded-2xl p-6 border border-slate-700 sticky top-6">
          <div class="flex items-center gap-3 mb-6">
            <!-- Dynamic Icon based on scanStatus -->
            <div v-if="scanStatus === 'idle'" class="p-2 bg-slate-800 text-slate-400 rounded-xl transition-all duration-300">
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-scan-barcode"><path d="M3 7V5a2 2 0 0 1 2-2h2"/><path d="M17 3h2a2 2 0 0 1 2 2v2"/><path d="M21 17v2a2 2 0 0 1-2 2h-2"/><path d="M7 21H5a2 2 0 0 1-2-2v-2"/><path d="M8 7v10"/><path d="M12 7v10"/><path d="M17 7v10"/></svg>
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
                }">Scanner U2B</h2>
              <p class="text-slate-400 text-xs mt-0.5">Fokus kursor disini untuk memindai</p>
            </div>
          </div>

          <div class="space-y-4">
            <div class="relative group">
              <input
                ref="scannerInputRef"
                v-model="scanInput"
                @keyup.enter="handleScan"
                type="text"
                placeholder="Scan Order ID / Tracking ID..."
                class="w-full rounded-xl px-4 py-4 focus:outline-none transition-all duration-300 font-mono text-lg shadow-inner"
                :class="{
                  'bg-emerald-950/40 border-2 border-emerald-500 ring-4 ring-emerald-500/30 text-emerald-300 scale-[1.02]': scanStatus === 'success',
                  'bg-amber-950/40 border-2 border-amber-500 ring-4 ring-amber-500/30 text-amber-300': scanStatus === 'warning',
                  'bg-red-950/40 border-2 border-red-500 ring-4 ring-red-500/30 text-red-300': scanStatus === 'error',
                  'bg-slate-950/50 border-2 border-slate-700 text-slate-200 placeholder:text-slate-600 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20': scanStatus === 'idle'
                }"
              />
              <div class="absolute inset-y-0 right-4 flex items-center pointer-events-none">
                <div class="w-2 h-2 rounded-full animate-pulse" :class="isScanning ? 'bg-emerald-500' : 'bg-slate-600'"></div>
              </div>
            </div>
            <p class="text-xs text-slate-400 text-center">Pastikan alat scanner Anda tersambung dan dapat mengirimkan input otomatis.</p>
          </div>

          <!-- Quick Stats -->
          <div class="mt-8 pt-6 border-t border-slate-700/50 grid grid-cols-2 gap-4">
            <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
              <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider">Terverifikasi (Hari ini)</p>
              <p class="text-3xl font-black text-emerald-400 mt-1">{{ todayVerifiedCount }}</p>
            </div>
            <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
              <p class="text-3xl font-black text-slate-300 mt-1">{{ pendingCount }}</p>
            </div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()

const orders = ref<any[]>([])
const loading = ref(true)
const syncing = ref(false)

const recentlyVerified = ref<string[]>([])
const scanStatus = ref<'idle' | 'success' | 'warning' | 'error'>('idle')
let scanStatusTimeout: number | null = null

const setScanStatus = (status: 'success' | 'warning' | 'error') => {
  scanStatus.value = status
  if (scanStatusTimeout) clearTimeout(scanStatusTimeout)
  scanStatusTimeout = window.setTimeout(() => {
    scanStatus.value = 'idle'
  }, 1000)
}


const selectedProduct = ref('')
const selectedWarehouse = ref('')
const selectedStatus = ref('')

const uniqueProducts = computed(() => {
  const products = orders.value.flatMap(o => o.items ? o.items.map((i: any) => i.product_name) : [])
  return [...new Set(products.filter(Boolean))].sort()
})

const uniqueWarehouses = computed(() => {
  return [...new Set(orders.value.map(o => o.warehouse_name).filter(Boolean))].sort()
})

const pendingCount = computed(() => {
  return orders.value.filter(o => o.order_status !== 'VERIFIED' && o.order_status !== 'PACKED').length
})

const filteredOrders = computed(() => {
  return orders.value.filter(order => {
    const matchProduct = selectedProduct.value ? order.items && order.items.some((i: any) => i.product_name === selectedProduct.value) : true
    const matchWarehouse = selectedWarehouse.value ? order.warehouse_name === selectedWarehouse.value : true
    let matchStatus = true;
    const isPicked = order.order_status === 'VERIFIED';
    const isRecentlyVerified = recentlyVerified.value.includes(order.order_id);
    
    if (selectedStatus.value === '') {
      matchStatus = !isPicked || isRecentlyVerified; // Menunggu Picking, tp keep jika recently verified (animasi)
    } else if (selectedStatus.value === 'VERIFIED') {
      matchStatus = isPicked; // Sudah dipicking
    } else if (selectedStatus.value === 'ALL') {
      matchStatus = true; // Semua
    }
    
    return matchProduct && matchWarehouse && matchStatus
  })
})

// Pagination logic
const currentPage = ref(1)
const itemsPerPage = 20

const totalPages = computed(() => Math.ceil(filteredOrders.value.length / itemsPerPage))

const paginatedOrders = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage
  return filteredOrders.value.slice(start, end)
})

watch([selectedProduct, selectedWarehouse, selectedStatus], () => {
  currentPage.value = 1
})

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

const getActiveBranchId = () => {
  const bid = masterDataStore.selectedBranchId
  return bid && bid !== 'ALL_BRANCHES' ? bid : ''
}

// Data fetching
const fetchOrders = async () => {
  loading.value = true
  try {
    const branchId = getActiveBranchId()
    const branchFilter = branchId ? `&branch_id=${branchId}` : ''
    const res = await apiClient.get(`/marketplace/orders?status=AWAITING_COLLECTION,VERIFIED&page_size=5000${branchFilter}`)
    orders.value = res.data.data || []
  } catch (error) {
    toast.error('Gagal memuat daftar pesanan marketplace')
  } finally {
    loading.value = false
  }
}

const syncTikTokOrders = async () => {
  syncing.value = true
  try {
    const branchId = getActiveBranchId()
    const branchFilter = branchId ? `?branch_id=${branchId}` : ''
    const res = await apiClient.post(`/marketplace/sync-orders${branchFilter}`)
    toast.success(`Berhasil! ${res.data.synced_count} pesanan baru ditarik.`)
    fetchOrders() // Refresh data
  } catch (error: any) {
    toast.error('Gagal menarik data dari TikTok: ' + (error.response?.data?.error || error.message))
  } finally {
    syncing.value = false
  }
}

// Scanner logic
const scanInput = ref('')
const scannerInputRef = ref<HTMLInputElement | null>(null)

const todayVerifiedCount = ref(0)

const fetchTodayProgress = async () => {
  try {
    const today = new Date().toISOString().split('T')[0]
    const branchId = getActiveBranchId()
    const branchFilter = branchId ? `&branch_id=${branchId}` : ''
    const res = await apiClient.get(`/marketplace/scan-history?action=PICK&from=${today}&to=${today}${branchFilter}`)
    todayVerifiedCount.value = res.data.data ? res.data.data.length : 0
  } catch (error) {
    console.error('Failed to fetch today progress:', error)
  }
}

const isScanning = ref(true) // Visual indicator

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
    osc.frequency.setValueAtTime(200, ctx.currentTime) // Low pitch buzz
    gainNode.gain.setValueAtTime(0.1, ctx.currentTime) 
    osc.connect(gainNode)
    gainNode.connect(ctx.destination)
    osc.start()
    osc.stop(ctx.currentTime + 0.3) // Play for 300ms
  } catch (e) {
    console.error('Audio playback failed', e)
  }
}

const handleScan = async () => {
  const code = scanInput.value.trim()
  if (!code) return

  // Find ALL orders matching order_id or tracking_id
  const matchedOrders = orders.value.filter(
    o => o.order_id === code || o.tracking_id === code
  )

  if (matchedOrders.length > 0) {
    const unverifiedOrders = matchedOrders.filter(o => o.order_status !== 'VERIFIED')
    
    if (unverifiedOrders.length === 0) {
      setScanStatus('warning')
      toast.info(`Semua pesanan untuk resi/order ${code} sudah diverifikasi sebelumnya.`)
      playAlreadyScannedSound()
    } else {
      isScanning.value = true
      try {
        const orderIdToVerify = unverifiedOrders[0]!.order_id
        await marketplaceStore.verifyMarketplaceOrder(orderIdToVerify)
        
        unverifiedOrders.forEach(matchedOrder => {
          todayVerifiedCount.value++
          matchedOrder.order_status = 'VERIFIED'
          // Add to recentlyVerified for visual feedback before popping out
          recentlyVerified.value.push(matchedOrder.order_id)
        })

        setScanStatus('success')
        playSuccessSound()
        toast.success(`Semua item untuk pesanan ${orderIdToVerify} terverifikasi!`)
        
        setTimeout(() => {
          unverifiedOrders.forEach(matchedOrder => {
            recentlyVerified.value = recentlyVerified.value.filter(id => id !== matchedOrder.order_id)
          })
        }, 1500)
      } catch (e: any) {
        setScanStatus('error')
        playErrorSound()
        toast.error(`Gagal verifikasi: ${e.response?.data?.error || e.message}`)
      } finally {
        isScanning.value = false
      }
    }
  } else {
    setScanStatus('error')
    playErrorSound()
    toast.error(`Data order dengan resi/ID ${code} tidak ditemukan!`)
  }

  // Clear input and maintain focus
  scanInput.value = ''
  if (scannerInputRef.value) {
    scannerInputRef.value.focus()
  }
}

let eventSource: EventSource | null = null

onMounted(() => {
  fetchOrders()
  fetchTodayProgress()
  // Auto focus scanner input on mount
  if (scannerInputRef.value) {
    scannerInputRef.value.focus()
  }

  // Set up Server-Sent Events (SSE) for real-time updates
  const apiUrl = apiClient.defaults.baseURL || import.meta.env.VITE_API_URL || ''
  eventSource = new EventSource(`${apiUrl}/marketplace/stream`)
  
  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.status === 'VERIFIED') {
        const order = orders.value.find(o => o.order_id === data.order_id)
        if (order) {
          if (order.order_status !== 'VERIFIED') {
            order.order_status = 'VERIFIED'
            order.picked_by = data.actor
            todayVerifiedCount.value++
            
            // Show toast if it wasn't scanned by this device (this device is currently not scanning it)
            if (!isScanning.value) {
              toast.info(`Pesanan ${data.order_id} telah diverifikasi oleh ${data.actor}`)
            }
            // Add to recentlyVerified for pop out animation
            recentlyVerified.value.push(data.order_id)
            setTimeout(() => {
              recentlyVerified.value = recentlyVerified.value.filter(id => id !== data.order_id)
            }, 1500)
          }
        }
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
})

watch(() => masterDataStore.selectedBranchId, () => {
  currentPage.value = 1
  fetchOrders()
  fetchTodayProgress()
})
</script>

<style scoped>
.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.5s ease;
}
.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
.list-leave-active {
  position: absolute;
}
</style>
