<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useShopeeBoostStore } from '@/stores/shopeeBoost'
import { useMasterDataStore } from '@/stores/masterData'
import { useMarketplaceStore } from '@/stores/marketplace'
import { marketplaceApi, type MarketplaceProduct } from '@/api/marketplace'
import {
  Rocket,
  RefreshCw,
  Search,
  Plus,
  Trash2,
  ArrowUp,
  ArrowDown,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Sparkles,
  Zap,
  RotateCcw,
  Store,
  Layers,
  Check,
  ChevronRight,
  Info,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const boostStore = useShopeeBoostStore()
const masterDataStore = useMasterDataStore()
const marketplaceStore = useMarketplaceStore()

const effectiveBranchId = computed<string | undefined>(() => {
  if (
    masterDataStore.isAllBranchesSelected ||
    !masterDataStore.selectedBranchId ||
    masterDataStore.selectedBranchId === 'ALL_BRANCHES'
  ) {
    return undefined
  }
  const firstValid = masterDataStore.selectedBranchIds.find(id => id && id !== 'ALL_BRANCHES')
  return firstValid || undefined
})

const currentBranchName = computed(() => {
  if (masterDataStore.isAllBranchesSelected || masterDataStore.selectedBranchId === 'ALL_BRANCHES') {
    return 'Semua Cabang (Global)'
  }
  return masterDataStore.selectedBranchName || 'Semua Cabang'
})

const searchQueue = ref('')
const isAddModalOpen = ref(false)
const modalSearch = ref('')
const selectedProductIds = ref<number[]>([])
const activeTab = ref<'queue' | 'logs'>('queue')

const modalProducts = ref<MarketplaceProduct[]>([])
const isModalProductsLoading = ref(false)
const isSyncing = ref(false)

// Realtime countdown clock ticker
const currentTimestamp = ref(Date.now())
let timerInterval: any = null

onMounted(async () => {
  if (!masterDataStore.branches.length) {
    await masterDataStore.fetchBranches()
  }
  await loadData()
  timerInterval = setInterval(() => {
    currentTimestamp.value = Date.now()
  }, 1000)
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})

watch(() => masterDataStore.selectedBranchId, async () => {
  await loadData()
})

const loadData = async () => {
  const branchId = effectiveBranchId.value
  await Promise.all([
    boostStore.fetchOverview(branchId),
    boostStore.fetchQueue(branchId),
  ])
}

const isEnabled = computed({
  get: () => boostStore.overview?.is_enabled ?? false,
  set: async (val: boolean) => {
    try {
      await boostStore.toggleAutoBoost(val, effectiveBranchId.value)
      toast.success(val ? 'Auto Boost Shopee diaktifkan!' : 'Auto Boost Shopee dinonaktifkan!')
    } catch (err: any) {
      toast.error(err.message || 'Gagal mengubah status auto boost')
    }
  },
})

// Format Countdown Timer
function getCooldownDisplay(cooldownUntilStr?: string): { text: string; isEnded: boolean; percent: number } {
  if (!cooldownUntilStr) return { text: '00:00:00', isEnded: true, percent: 100 }
  const target = new Date(cooldownUntilStr).getTime()
  const diff = target - currentTimestamp.value
  if (diff <= 0) return { text: 'Selesai', isEnded: true, percent: 100 }

  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  const seconds = Math.floor((diff % (1000 * 60)) / 1000)

  const pad = (n: number) => (n < 10 ? '0' + n : n)
  // Shopee boost is 4 hours = 14400 seconds
  const totalDuration = 4 * 60 * 60 * 1000
  const elapsed = totalDuration - diff
  const percent = Math.min(100, Math.max(0, (elapsed / totalDuration) * 100))

  return {
    text: `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`,
    isEnded: false,
    percent,
  }
}

// Format Currency
function formatCurrency(val?: number) {
  if (!val) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(val)
}

// Format Date
function formatDate(dateStr?: string) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Filtered Queue
const filteredQueue = computed(() => {
  if (!searchQueue.value.trim()) return boostStore.queue
  const q = searchQueue.value.toLowerCase()
  return boostStore.queue.filter(
    item => item.item_name.toLowerCase().includes(q) || (item.item_sku && item.item_sku.toLowerCase().includes(q))
  )
})

// Trigger Boost Now
const handleTriggerNow = async () => {
  try {
    const res = await boostStore.triggerBoostNow(effectiveBranchId.value)
    toast.success(res.message)
  } catch (err: any) {
    toast.error(err.message)
  }
}

// Reorder Up / Down
const moveItem = async (index: number, direction: 'up' | 'down') => {
  const currentList = [...boostStore.queue]
  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= currentList.length) return

  const temp = currentList[index]
  const target = currentList[targetIndex]
  if (!temp || !target) return

  currentList[index] = target
  currentList[targetIndex] = temp

  boostStore.queue = currentList
  try {
    await boostStore.reorderQueue(currentList.map(item => item.id))
    toast.success('Urutan diperbarui')
  } catch (err: any) {
    toast.error('Gagal memperbarui urutan')
    await boostStore.fetchQueue(effectiveBranchId.value)
  }
}

// Toggle Item Active
const handleToggleItem = async (item: any) => {
  const nextStatus = !item.is_active
  try {
    await boostStore.toggleItemActive(item.id, nextStatus, effectiveBranchId.value)
    toast.success(nextStatus ? 'Produk diaktifkan dalam antrean' : 'Produk dijeda dari antrean')
  } catch (err: any) {
    toast.error(err.message)
  }
}

// Delete Item from Queue
const handleDeleteItem = async (id: number) => {
  if (!confirm('Yakin ingin menghapus produk ini dari antrean auto-boost?')) return
  try {
    await boostStore.deleteFromQueue(id, effectiveBranchId.value)
    toast.success('Produk dihapus dari antrean')
  } catch (err: any) {
    toast.error(err.message)
  }
}

// Open Add Modal & Fetch Shopee Products
const openAddModal = async () => {
  isAddModalOpen.value = true
  selectedProductIds.value = []
  modalSearch.value = ''
  await fetchModalProducts()
}

const fetchModalProducts = async () => {
  isModalProductsLoading.value = true
  try {
    const res = await marketplaceApi.getMarketplaceProducts('', effectiveBranchId.value, 'shopee')
    modalProducts.value = res.data
  } catch (err: any) {
    toast.error('Gagal memuat produk Shopee untuk cabang terpilih')
  } finally {
    isModalProductsLoading.value = false
  }
}

const handleSyncShopee = async () => {
  isSyncing.value = true
  try {
    const res = await marketplaceApi.syncMarketplaceProducts(effectiveBranchId.value, 'shopee')
    toast.success(res.data?.message || 'Produk Shopee berhasil disinkronkan!')
    await fetchModalProducts()
  } catch (err: any) {
    toast.error(err.message || 'Gagal menyinkronkan produk dari Shopee')
  } finally {
    isSyncing.value = false
  }
}

// Shopee products available to add (not yet in queue)
const availableShopeeProducts = computed(() => {
  const existingItemIDs = new Set(boostStore.queue.map(q => String(q.item_id)))
  
  // Agregasi stok dan produk unik berdasarkan product_id (Shopee item_id)
  const productMap = new Map<string, {
    product_id: string
    product_name: string
    seller_sku: string
    price_amount: number
    total_stock: number
    status: string
  }>()

  for (const p of modalProducts.value) {
    if (p.marketplace !== 'shopee') continue
    if (existingItemIDs.has(p.product_id)) continue

    const existing = productMap.get(p.product_id)
    if (!existing) {
      productMap.set(p.product_id, {
        product_id: p.product_id,
        product_name: p.product_name,
        seller_sku: p.seller_sku,
        price_amount: p.price_amount,
        total_stock: p.stock || 0,
        status: p.status,
      })
    } else {
      existing.total_stock += (p.stock || 0)
    }
  }

  const list = Array.from(productMap.values())
  if (!modalSearch.value.trim()) return list
  const q = modalSearch.value.toLowerCase()
  return list.filter(
    p => p.product_name.toLowerCase().includes(q) || (p.seller_sku && p.seller_sku.toLowerCase().includes(q))
  )
})

const toggleSelectAllModal = () => {
  const selectable = availableShopeeProducts.value.filter(p => p.total_stock > 0)
  if (selectedProductIds.value.length === selectable.length) {
    selectedProductIds.value = []
  } else {
    selectedProductIds.value = selectable.map(p => Number(p.product_id)).filter(id => !isNaN(id))
  }
}

const toggleSelectProduct = (prod: any) => {
  if (prod.total_stock <= 0) {
    toast.error('Produk ini tidak dapat dipilih karena stok 0 di Shopee.')
    return
  }
  const numId = Number(prod.product_id)
  if (isNaN(numId)) return
  const idx = selectedProductIds.value.indexOf(numId)
  if (idx > -1) {
    selectedProductIds.value.splice(idx, 1)
  } else {
    selectedProductIds.value.push(numId)
  }
}

const handleAddSelected = async () => {
  if (!selectedProductIds.value.length) return
  try {
    const res = await boostStore.addToQueue(selectedProductIds.value, effectiveBranchId.value)
    toast.success(res.message || 'Produk berhasil ditambahkan ke antrean')
    isAddModalOpen.value = false
  } catch (err: any) {
    toast.error(err.message)
  }
}
</script>

<template>
  <div class="space-y-6 pb-12">
    <!-- Header Section -->
    <div class="bg-gradient-to-r from-orange-600 via-orange-500 to-amber-500 rounded-2xl p-6 text-white shadow-lg relative overflow-hidden">
      <!-- Decorative background icon -->
      <div class="absolute -right-6 -bottom-6 opacity-15 pointer-events-none">
        <Rocket class="w-64 h-64 text-white" />
      </div>

      <div class="relative z-10 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
        <div>
          <div class="flex items-center gap-2.5 mb-2">
            <span class="p-2 bg-white/20 backdrop-blur-md rounded-xl text-white">
              <Rocket class="w-6 h-6" />
            </span>
            <span class="text-xs font-bold tracking-wider uppercase px-2.5 py-1 bg-white/20 rounded-full">
              Shopee Open API
            </span>
          </div>
          <h1 class="text-2xl lg:text-3xl font-black tracking-tight">Shopee Auto Boost Produk</h1>
          <p class="text-orange-100 text-sm mt-1 max-w-2xl leading-relaxed">
            Otomatis menaikkan hingga 5 produk pilihan ke hasil pencarian teratas Shopee setiap 4 jam secara bergilir (Smart Round-Robin).
          </p>
          <div class="flex flex-wrap items-center gap-2 mt-2">
            <span class="text-[11px] px-2.5 py-1 bg-white/20 rounded-lg text-white font-semibold flex items-center gap-1.5 backdrop-blur-md">
              <Store class="w-3.5 h-3.5 text-orange-200" />
              <span>Cabang: {{ currentBranchName }}</span>
            </span>
            <span v-if="boostStore.overview?.shop_id" class="text-[11px] px-2.5 py-1 bg-white/20 rounded-lg text-white font-mono">
              Shop ID: {{ boostStore.overview.shop_id }}
            </span>
            <span v-if="boostStore.overview.available_slots === 0" class="text-[11px] px-2.5 py-1 bg-amber-400/30 text-amber-100 rounded-lg font-bold flex items-center gap-1">
              <Clock class="w-3.5 h-3.5 text-amber-300" />
              5/5 Slot Sedang Aktif — Antrean berputar otomatis saat countdown selesai
            </span>
          </div>
        </div>

        <!-- Controls: Master Toggle & Actions -->
        <div class="flex flex-wrap items-center gap-3 bg-white/10 backdrop-blur-md p-3 rounded-2xl border border-white/20">
          <!-- Master Auto-Boost Switch -->
          <div class="flex items-center gap-3 px-4 py-2 bg-white text-slate-900 rounded-xl shadow-md">
            <div class="flex flex-col">
              <span class="text-[10px] uppercase font-bold text-slate-400 leading-none">Auto Boost</span>
              <span class="text-xs font-black" :class="isEnabled ? 'text-emerald-600' : 'text-slate-500'">
                {{ isEnabled ? 'AKTIF' : 'NONAKTIF' }}
              </span>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                v-model="isEnabled"
                class="sr-only peer"
                :disabled="boostStore.isActionLoading"
              />
              <div class="w-11 h-6 bg-slate-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-orange-600"></div>
            </label>
          </div>

          <!-- Quick Action Buttons -->
          <button
            @click="handleTriggerNow"
            :disabled="boostStore.isActionLoading || (boostStore.overview?.available_slots === 0)"
            class="flex items-center gap-1.5 px-3.5 py-2.5 font-bold text-xs rounded-xl shadow transition active:scale-95 disabled:opacity-75"
            :class="boostStore.overview?.available_slots === 0
              ? 'bg-amber-100 text-amber-900 cursor-not-allowed border border-amber-300'
              : 'bg-amber-400 hover:bg-amber-300 text-amber-950'"
            :title="boostStore.overview?.available_slots === 0
              ? 'Semua 5 slot sedang aktif. Produk berikutnya akan otomatis dinaikkan saat countdown slot selesai.'
              : 'Langsung naikkan produk berikutnya jika ada slot kosong tanpa menunggu jadwal'"
          >
            <Zap class="w-4 h-4" :class="boostStore.overview?.available_slots === 0 ? 'text-amber-800' : 'fill-amber-950 text-amber-950'" />
            <span>{{ boostStore.overview?.available_slots === 0 ? '5 Slot Penuh (Aktif)' : 'Boost Sekarang' }}</span>
          </button>

          <button
            @click="loadData"
            :disabled="boostStore.isLoading"
            class="p-2.5 bg-white/20 hover:bg-white/30 text-white rounded-xl transition"
            title="Segarkan data slot"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': boostStore.isLoading }" />
          </button>
        </div>
      </div>
    </div>

    <!-- 5 SLOTS VISUALIZATION -->
    <div>
      <div class="flex items-center justify-between mb-3 px-1">
        <div class="flex items-center gap-2">
          <Sparkles class="w-5 h-5 text-orange-500" />
          <h2 class="text-base font-black text-slate-800">Status 5 Slot Resmi Shopee</h2>
          <span class="text-xs px-2.5 py-0.5 rounded-full font-bold bg-orange-100 text-orange-700">
            {{ boostStore.overview?.occupied_slots ?? 0 }} / 5 Terisi
          </span>
        </div>
        <span class="text-xs text-slate-400 flex items-center gap-1">
          <Clock class="w-3.5 h-3.5" />
          Durasi dorongan: 4 jam / produk
        </span>
      </div>

      <!-- Grid of 5 Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3.5">
        <template v-if="boostStore.overview?.slots">
          <div
            v-for="slot in boostStore.overview.slots"
            :key="slot.slot_index"
            class="relative rounded-2xl border transition-all duration-200 overflow-hidden flex flex-col justify-between"
            :class="slot.is_occupied
              ? 'bg-gradient-to-b from-orange-50/70 to-white border-orange-200 shadow-sm'
              : 'bg-slate-50/70 border-dashed border-slate-300 text-slate-400'"
          >
            <!-- Slot Header Badge -->
            <div class="p-3 pb-2 flex items-center justify-between">
              <span
                class="text-[11px] font-black px-2 py-0.5 rounded-lg"
                :class="slot.is_occupied ? 'bg-orange-500 text-white shadow-xs' : 'bg-slate-200 text-slate-600'"
              >
                Slot #{{ slot.slot_index }}
              </span>
              <span
                v-if="slot.is_occupied"
                class="flex items-center gap-1 text-[11px] font-bold"
                :class="getCooldownDisplay(slot.cooldown_until).isEnded ? 'text-amber-600' : 'text-orange-600'"
              >
                <span
                  class="w-2 h-2 rounded-full"
                  :class="getCooldownDisplay(slot.cooldown_until).isEnded ? 'bg-amber-500' : 'bg-orange-500 animate-ping'"
                ></span>
                {{ getCooldownDisplay(slot.cooldown_until).isEnded ? 'Selesai (Siap Rotasi)' : 'Aktif' }}
              </span>
              <span v-else class="text-[11px] font-medium text-slate-400">Tersedia</span>
            </div>

            <!-- Slot Content Body -->
            <div class="p-3 pt-0 flex-1 flex flex-col justify-center">
              <template v-if="slot.is_occupied">
                <div class="flex items-start gap-2.5">
                  <div class="w-12 h-12 rounded-xl bg-orange-100 border border-orange-200 flex-shrink-0 overflow-hidden flex items-center justify-center">
                    <img
                      v-if="slot.image_url"
                      :src="slot.image_url"
                      :alt="slot.item_name"
                      class="w-full h-full object-cover"
                    />
                    <Rocket v-else class="w-6 h-6 text-orange-500" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="text-xs font-bold text-slate-800 line-clamp-2 leading-tight" :title="slot.item_name">
                      {{ slot.item_name }}
                    </p>
                    <span class="text-[10px] text-slate-400 font-mono">ID: {{ slot.item_id }}</span>
                  </div>
                </div>

                <!-- Live Countdown Timer -->
                <div class="mt-3 bg-white p-2.5 rounded-xl border border-orange-100 shadow-2xs">
                  <div class="flex items-center justify-between text-[11px] mb-1">
                    <span class="text-slate-400 font-medium flex items-center gap-1">
                      <Clock class="w-3 h-3 text-orange-500" />
                      Sisa Waktu
                    </span>
                    <span class="font-black font-mono text-orange-700 text-xs">
                      {{ getCooldownDisplay(slot.cooldown_until).text }}
                    </span>
                  </div>
                  <!-- Progress Bar -->
                  <div class="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                    <div
                      class="bg-orange-500 h-full rounded-full transition-all duration-1000"
                      :style="{ width: `${getCooldownDisplay(slot.cooldown_until).percent}%` }"
                    ></div>
                  </div>
                </div>
              </template>

              <!-- Slot is empty -->
              <template v-else>
                <div class="py-6 flex flex-col items-center justify-center text-center">
                  <div class="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center text-slate-400 mb-2">
                    <Plus class="w-5 h-5" />
                  </div>
                  <p class="text-xs font-bold text-slate-600">Slot Kosong</p>
                  <p class="text-[10px] text-slate-400 mt-0.5">Siap diisi produk antrean</p>
                </div>
              </template>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- MAIN CONTENT: QUEUE TABLE & LOGS TABS -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <!-- Tabs Navigation -->
      <div class="flex items-center justify-between border-b border-slate-200 px-6 pt-4 pb-0 bg-slate-50/50">
        <div class="flex items-center gap-6">
          <button
            @click="activeTab = 'queue'"
            class="pb-3 text-sm font-bold border-b-2 flex items-center gap-2 transition"
            :class="activeTab === 'queue'
              ? 'border-orange-600 text-orange-600'
              : 'border-transparent text-slate-500 hover:text-slate-800'"
          >
            <Layers class="w-4 h-4" />
            <span>Antrean Rotasi Produk</span>
            <span class="text-xs px-2 py-0.5 rounded-full font-bold"
              :class="activeTab === 'queue' ? 'bg-orange-100 text-orange-700' : 'bg-slate-200 text-slate-600'">
              {{ boostStore.queue.length }}
            </span>
          </button>

          <button
            @click="activeTab = 'logs'"
            class="pb-3 text-sm font-bold border-b-2 flex items-center gap-2 transition"
            :class="activeTab === 'logs'
              ? 'border-orange-600 text-orange-600'
              : 'border-transparent text-slate-500 hover:text-slate-800'"
          >
            <RotateCcw class="w-4 h-4" />
            <span>Riwayat Log Eksekusi</span>
          </button>
        </div>

        <!-- Add product button (visible when on queue tab) -->
        <div v-if="activeTab === 'queue'" class="pb-3">
          <button
            @click="openAddModal"
            class="flex items-center gap-1.5 px-4 py-2 bg-orange-600 hover:bg-orange-700 text-white text-xs font-bold rounded-xl shadow-xs transition active:scale-95"
          >
            <Plus class="w-4 h-4" />
            <span>Tambah Produk ke Antrean</span>
          </button>
        </div>
      </div>

      <!-- TAB 1: ANTREN ROTASI PRODUK -->
      <div v-if="activeTab === 'queue'" class="p-6 space-y-4">
        <!-- Search & Info Bar -->
        <div class="flex flex-col sm:flex-row items-center justify-between gap-3">
          <div class="relative w-full sm:w-80">
            <Search class="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              v-model="searchQueue"
              placeholder="Cari produk di antrean..."
              class="w-full pl-10 pr-4 py-2 text-xs rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 bg-slate-50/50"
            />
          </div>

          <div class="flex items-center gap-2 text-xs text-slate-500 bg-amber-50 border border-amber-200/60 px-3 py-1.5 rounded-xl">
            <Info class="w-4 h-4 text-amber-600 flex-shrink-0" />
            <span>
              <strong>Metode Round-Robin:</strong> Produk akan dinaikkan bergiliran 5 per 4 jam. Atur prioritas urutan dengan tombol panah naik/turun.
            </span>
          </div>
        </div>

        <!-- Table -->
        <div class="overflow-x-auto rounded-xl border border-slate-100">
          <table class="w-full text-left text-xs text-slate-600">
            <thead class="bg-slate-50 text-[11px] uppercase tracking-wider text-slate-400 font-bold border-b border-slate-100">
              <tr>
                <th class="py-3 px-4 w-12 text-center">Urutan</th>
                <th class="py-3 px-4">Produk</th>
                <th class="py-3 px-4 w-28">Harga</th>
                <th class="py-3 px-4 w-24 text-center">Stok</th>
                <th class="py-3 px-4 w-36">Terakhir Dinaikkan</th>
                <th class="py-3 px-4 w-36 text-center">Status</th>
                <th class="py-3 px-4 w-20 text-center">Aktifkan</th>
                <th class="py-3 px-4 w-16 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <template v-if="filteredQueue.length > 0">
                <tr
                  v-for="(item, index) in filteredQueue"
                  :key="item.id"
                  class="hover:bg-slate-50/70 transition"
                  :class="{ 'opacity-60 bg-slate-50/40': !item.is_active }"
                >
                  <!-- Sequence & Reorder Controls -->
                  <td class="py-3 px-4 text-center">
                    <div class="flex items-center justify-center gap-1">
                      <span class="font-bold font-mono text-slate-700 w-5">#{{ index + 1 }}</span>
                      <div class="flex flex-col">
                        <button
                          @click="moveItem(index, 'up')"
                          :disabled="index === 0"
                          class="p-0.5 text-slate-400 hover:text-slate-800 disabled:opacity-20"
                        >
                          <ArrowUp class="w-3 h-3" />
                        </button>
                        <button
                          @click="moveItem(index, 'down')"
                          :disabled="index === filteredQueue.length - 1"
                          class="p-0.5 text-slate-400 hover:text-slate-800 disabled:opacity-20"
                        >
                          <ArrowDown class="w-3 h-3" />
                        </button>
                      </div>
                    </div>
                  </td>

                  <!-- Product Info -->
                  <td class="py-3 px-4">
                    <div class="flex items-center gap-3">
                      <div class="w-10 h-10 rounded-lg bg-slate-100 border border-slate-200 flex-shrink-0 overflow-hidden flex items-center justify-center">
                        <img
                          v-if="item.image_url"
                          :src="item.image_url"
                          :alt="item.item_name"
                          class="w-full h-full object-cover"
                        />
                        <Tag class="w-5 h-5 text-slate-400" />
                      </div>
                      <div>
                        <p class="font-bold text-slate-800 line-clamp-1" :title="item.item_name">
                          {{ item.item_name }}
                        </p>
                        <div class="flex items-center gap-2 mt-0.5">
                          <span class="text-[10px] text-slate-400 font-mono">ID: {{ item.item_id }}</span>
                          <span v-if="item.item_sku" class="text-[10px] bg-slate-100 text-slate-600 px-1.5 py-0.2 rounded font-mono">
                            SKU: {{ item.item_sku }}
                          </span>
                        </div>
                      </div>
                    </div>
                  </td>

                  <!-- Price -->
                  <td class="py-3 px-4 font-semibold text-slate-800">
                    {{ formatCurrency(item.price_amount) }}
                  </td>

                  <!-- Stock -->
                  <td class="py-3 px-4 text-center">
                    <span
                      v-if="item.current_stock !== undefined && item.current_stock > 0"
                      class="font-black text-slate-700 font-mono"
                    >
                      {{ item.current_stock }}
                    </span>
                    <span
                      v-else-if="item.current_stock !== undefined && item.current_stock === 0"
                      class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-black bg-rose-100 text-rose-700"
                    >
                      Habis (0)
                    </span>
                    <span v-else class="text-slate-400">-</span>
                  </td>

                  <!-- Last Boosted -->
                  <td class="py-3 px-4">
                    <div class="flex flex-col">
                      <span class="font-medium text-slate-700">
                        {{ item.last_boosted_at ? formatDate(item.last_boosted_at) : 'Belum pernah' }}
                      </span>
                      <span v-if="item.status === 'BOOSTED'" class="text-[10px] text-orange-600 font-bold">
                        Sedang dinaikkan
                      </span>
                    </div>
                  </td>

                  <!-- Status Badge -->
                  <td class="py-3 px-4 text-center">
                    <div class="flex flex-col items-center gap-0.5">
                      <span
                        v-if="item.status === 'BOOSTED'"
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[10px] font-bold bg-orange-100 text-orange-700"
                      >
                        <span class="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse"></span>
                        Boosted
                      </span>
                      <span
                        v-else-if="item.status === 'FAILED'"
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[10px] font-bold bg-red-100 text-red-700"
                        :title="item.last_error"
                      >
                        <XCircle class="w-3 h-3 text-red-600" />
                        Gagal
                      </span>
                      <span
                        v-else
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[10px] font-bold bg-slate-100 text-slate-600"
                      >
                        Menunggu
                      </span>

                      <!-- Error explanation if failed -->
                      <span
                        v-if="item.status === 'FAILED' && item.last_error"
                        class="text-[9px] text-red-600 font-medium max-w-[130px] truncate cursor-help"
                        :title="item.last_error"
                      >
                        {{ item.last_error }}
                      </span>
                    </div>
                  </td>

                  <!-- Toggle Active Switch -->
                  <td class="py-3 px-4 text-center">
                    <button
                      @click="handleToggleItem(item)"
                      class="relative inline-flex items-center h-5 rounded-full w-9 transition-colors focus:outline-none"
                      :class="item.is_active ? 'bg-emerald-500' : 'bg-slate-300'"
                    >
                      <span
                        class="inline-block w-3.5 h-3.5 transform transition bg-white rounded-full"
                        :class="item.is_active ? 'translate-x-4' : 'translate-x-1'"
                      />
                    </button>
                  </td>

                  <!-- Actions -->
                  <td class="py-3 px-4 text-center">
                    <button
                      @click="handleDeleteItem(item.id)"
                      class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition"
                      title="Hapus dari antrean"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              </template>

              <!-- Empty state -->
              <tr v-else>
                <td colspan="7" class="py-12 text-center text-slate-400">
                  <div class="flex flex-col items-center justify-center">
                    <div class="w-12 h-12 rounded-full bg-orange-50 flex items-center justify-center text-orange-500 mb-3">
                      <Rocket class="w-6 h-6" />
                    </div>
                    <p class="font-bold text-slate-700 text-sm">Belum Ada Produk di Antrean</p>
                    <p class="text-xs text-slate-400 mt-1 max-w-sm">
                      Tambahkan produk Shopee ke dalam antrean agar sistem dapat menaikkan produk secara otomatis setiap 4 jam.
                    </p>
                    <button
                      @click="openAddModal"
                      class="mt-4 flex items-center gap-1.5 px-4 py-2 bg-orange-600 hover:bg-orange-700 text-white text-xs font-bold rounded-xl shadow transition"
                    >
                      <Plus class="w-4 h-4" />
                      <span>Tambah Produk Sekarang</span>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- TAB 2: RIWAYAT LOG EKSEKUSI -->
      <div v-else class="p-6">
        <div class="overflow-x-auto rounded-xl border border-slate-100">
          <table class="w-full text-left text-xs text-slate-600">
            <thead class="bg-slate-50 text-[11px] uppercase tracking-wider text-slate-400 font-bold border-b border-slate-100">
              <tr>
                <th class="py-3 px-4 w-40">Waktu Boost</th>
                <th class="py-3 px-4">Nama Produk</th>
                <th class="py-3 px-4 w-32">Pemicu</th>
                <th class="py-3 px-4 w-28 text-center">Status</th>
                <th class="py-3 px-4">Keterangan</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <template v-if="boostStore.logs.length > 0">
                <tr v-for="log in boostStore.logs" :key="log.id" class="hover:bg-slate-50/60">
                  <td class="py-3 px-4 font-mono text-slate-700">
                    {{ formatDate(log.boosted_at) }}
                  </td>
                  <td class="py-3 px-4">
                    <span class="font-bold text-slate-800">{{ log.item_name }}</span>
                    <span class="text-[10px] text-slate-400 font-mono block">ID: {{ log.item_id }}</span>
                  </td>
                  <td class="py-3 px-4">
                    <span
                      class="text-[10px] font-bold px-2 py-0.5 rounded-full"
                      :class="log.triggered_by === 'MANUAL' ? 'bg-blue-100 text-blue-700' : 'bg-slate-100 text-slate-700'"
                    >
                      {{ log.triggered_by === 'MANUAL' ? 'Manual Klik' : 'Auto Worker' }}
                    </span>
                  </td>
                  <td class="py-3 px-4 text-center">
                    <span
                      v-if="log.status === 'SUCCESS'"
                      class="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-700"
                    >
                      <CheckCircle2 class="w-3 h-3 text-emerald-600" />
                      Sukses
                    </span>
                    <span
                      v-else
                      class="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-red-100 text-red-700"
                    >
                      <XCircle class="w-3 h-3 text-red-600" />
                      Gagal
                    </span>
                  </td>
                  <td class="py-3 px-4 text-slate-500">
                    {{ log.message }}
                  </td>
                </tr>
              </template>
              <tr v-else>
                <td colspan="5" class="py-8 text-center text-slate-400">
                  Belum ada catatan aktivitas boost.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- MODAL: PILIH PRODUK DARI KATALOG SHOPEE -->
    <div
      v-if="isAddModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-xs p-4"
    >
      <div class="bg-white w-full max-w-2xl rounded-2xl shadow-2xl flex flex-col max-h-[85vh] overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <!-- Modal Header -->
        <div class="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50/70">
          <div>
            <h3 class="text-base font-black text-slate-800">Pilih Produk Shopee untuk Antrean</h3>
            <p class="text-xs text-slate-400 mt-0.5">
              Pilih produk yang ingin diikutkan dalam rotasi auto-boost — <span class="font-semibold text-slate-600">{{ currentBranchName }}</span>.
            </p>
          </div>
          <button
            @click="isAddModalOpen = false"
            class="text-slate-400 hover:text-slate-600 p-1 rounded-lg"
          >
            ✕
          </button>
        </div>

        <!-- Search Bar & Actions in Modal -->
        <div class="p-4 border-b border-slate-100 flex flex-wrap items-center justify-between gap-3 bg-white">
          <div class="relative flex-1 min-w-[200px]">
            <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              v-model="modalSearch"
              placeholder="Cari nama produk atau SKU..."
              class="w-full pl-9 pr-4 py-2 text-xs rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
            />
          </div>

          <div class="flex items-center gap-2">
            <button
              @click="handleSyncShopee"
              :disabled="isSyncing || isModalProductsLoading"
              class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 rounded-xl transition disabled:opacity-50"
              title="Tarik produk terbaru langsung dari Shopee Seller Centre"
            >
              <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': isSyncing }" />
              <span>{{ isSyncing ? 'Menyinkronkan...' : 'Sinkronkan Shopee' }}</span>
            </button>

            <button
              @click="toggleSelectAllModal"
              class="text-xs text-orange-600 hover:text-orange-700 font-bold px-2 py-1"
            >
              {{ selectedProductIds.length === availableShopeeProducts.filter(p => p.total_stock > 0).length && availableShopeeProducts.filter(p => p.total_stock > 0).length > 0 ? 'Batal Pilih Semua' : 'Pilih Semua' }}
            </button>
          </div>
        </div>

        <!-- Product List -->
        <div class="flex-1 overflow-y-auto p-4 divide-y divide-slate-100">
          <div v-if="isModalProductsLoading" class="py-12 flex flex-col items-center justify-center text-slate-400">
            <RefreshCw class="w-8 h-8 animate-spin text-orange-500 mb-2" />
            <p class="text-xs font-medium">Memuat katalog produk Shopee...</p>
          </div>
          <template v-else-if="availableShopeeProducts.length > 0">
            <div
              v-for="prod in availableShopeeProducts"
              :key="prod.product_id"
              @click="toggleSelectProduct(prod)"
              class="py-3 px-3 rounded-xl flex items-center justify-between transition"
              :class="[
                prod.total_stock <= 0
                  ? 'opacity-40 bg-slate-50 cursor-not-allowed'
                  : 'hover:bg-orange-50/60 cursor-pointer',
                selectedProductIds.includes(Number(prod.product_id))
                  ? 'bg-orange-50/80 border border-orange-200'
                  : ''
              ]"
            >
              <div class="flex items-center gap-3">
                <input
                  type="checkbox"
                  :disabled="prod.total_stock <= 0"
                  :checked="selectedProductIds.includes(Number(prod.product_id))"
                  class="rounded text-orange-600 focus:ring-orange-500 w-4 h-4 cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                />
                <div>
                  <p class="text-xs font-bold text-slate-800 line-clamp-1">
                    {{ prod.product_name }}
                  </p>
                  <div class="flex items-center gap-2 mt-0.5">
                    <span class="text-[10px] text-slate-400 font-mono">ID: {{ prod.product_id }}</span>
                    <span v-if="prod.seller_sku" class="text-[10px] text-slate-500 font-mono">
                      SKU: {{ prod.seller_sku }}
                    </span>
                    <span
                      v-if="prod.total_stock > 0"
                      class="text-[10px] px-2 py-0.2 rounded-full font-bold bg-emerald-100 text-emerald-700 font-mono"
                    >
                      Stok: {{ prod.total_stock }}
                    </span>
                    <span
                      v-else
                      class="text-[10px] px-2 py-0.2 rounded-full font-bold bg-rose-100 text-rose-700"
                    >
                      Stok 0 (Habis)
                    </span>
                  </div>
                </div>
              </div>

              <div class="flex flex-col items-end">
                <span class="text-xs font-bold text-slate-700">
                  {{ formatCurrency(prod.price_amount) }}
                </span>
                <span v-if="prod.total_stock <= 0" class="text-[10px] text-rose-500 font-semibold">
                  Tidak bisa diboost
                </span>
              </div>
            </div>
          </template>

          <div v-else class="py-12 text-center text-slate-400">
            <p class="text-sm font-semibold">Tidak ada produk Shopee yang tersedia</p>
            <p class="text-xs mt-1">Semua produk sudah berada di antrean atau belum disinkronkan.</p>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="px-6 py-3 border-t border-slate-200 bg-slate-50 flex items-center justify-between">
          <span class="text-xs text-slate-500 font-medium">
            <strong>{{ selectedProductIds.length }}</strong> produk dipilih
          </span>

          <div class="flex items-center gap-2">
            <button
              @click="isAddModalOpen = false"
              class="px-4 py-2 text-xs font-semibold text-slate-600 hover:text-slate-800 rounded-xl"
            >
              Batal
            </button>
            <button
              @click="handleAddSelected"
              :disabled="!selectedProductIds.length || boostStore.isActionLoading"
              class="px-5 py-2 bg-orange-600 hover:bg-orange-700 text-white text-xs font-bold rounded-xl shadow transition disabled:opacity-50"
            >
              Tambahkan ke Antrean
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
