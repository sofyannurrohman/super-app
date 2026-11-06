<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useSalesOrderStore, type SalesOrder, type SOItem } from '@/stores/salesOrder'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import {
  X,
  AlertTriangle,
  CheckCircle2,
  Clock,
  RefreshCw,
  SlidersHorizontal,
  Package,
  Layers,
  Building2,
  User,
  Store,
  Calendar,
  AlertCircle,
  ArrowUpRight,
  ShieldAlert,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const props = defineProps<{
  show: boolean
  order: SalesOrder | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'refresh'): void
  (e: 'open-adjust', order: SalesOrder): void
}>()

const soStore = useSalesOrderStore()
const masterStore = useMasterDataStore()

const loadingInventory = ref(false)
const branchStockMap = ref<Record<string, { quantity: number }>>({})
const isRechecking = ref(false)
const isOverriding = ref(false)

// Fetch real-time branch stock for comparison
async function fetchBranchStock() {
  if (!props.order?.branch_id) return
  loadingInventory.value = true
  try {
    const res = await apiClient.get('/warehouse/stock', {
      params: { branch_id: props.order.branch_id },
    })
    const map: Record<string, { quantity: number }> = {}
    if (Array.isArray(res.data)) {
      res.data.forEach((item: any) => {
        map[item.product_id] = {
          quantity: Number(item.quantity) || 0,
        }
      })
    }
    branchStockMap.value = map
  } catch (e) {
    console.error('Failed to fetch branch stock', e)
  } finally {
    loadingInventory.value = false
  }
}

watch(
  () => props.show,
  (val) => {
    if (val && props.order) {
      fetchBranchStock()
    }
  },
  { immediate: true },
)

const formatCurrency = (v: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(v)

const formatDate = (d?: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : '-'

// Item breakdown detailed analysis
const analyzedItems = computed(() => {
  if (!props.order || !props.order.items) return []

  return props.order.items.map((item: SOItem) => {
    const totalPieces = (Number(item.ordered_quantity) || 1) * (Number(item.pieces_per_unit) || 1)
    const batches = item.batches || []

    const backorderBatch = batches.find((b) => b.batch_no === 'BACKORDER')
    const backorderPieces = backorderBatch ? Number(backorderBatch.quantity || 0) : 0

    const realBatches = batches.filter((b) => b.batch_no !== 'BACKORDER' && Number(b.quantity || 0) > 0)
    const readyPieces = realBatches.reduce((sum, b) => sum + Number(b.quantity || 0), 0)

    const piecesPerUnit = Number(item.pieces_per_unit) || 1
    const backorderUnits = Math.ceil(backorderPieces / piecesPerUnit)
    const readyUnits = Math.floor(readyPieces / piecesPerUnit)

    const readyPercentage = totalPieces > 0 ? Math.min(100, Math.round((readyPieces / totalPieces) * 100)) : 100

    const stockInWarehouse = branchStockMap.value[item.product_id] || { quantity: 0 }
    const warehouseStock = stockInWarehouse.quantity

    let statusType: 'READY' | 'PARTIAL' | 'EMPTY' = 'READY'
    if (readyPieces === 0) {
      statusType = 'EMPTY'
    } else if (backorderPieces > 0) {
      statusType = 'PARTIAL'
    }

    return {
      ...item,
      productName: item.product?.name || 'Produk',
      sku: item.product?.sku || '-',
      totalPieces,
      readyPieces,
      readyUnits,
      backorderPieces,
      backorderUnits,
      readyPercentage,
      statusType,
      realBatches,
      hasBackorder: backorderPieces > 0,
      totalWarehouseStock: warehouseStock,
    }
  })
})

// Global SO readiness metrics
const readinessSummary = computed(() => {
  const items = analyzedItems.value
  if (items.length === 0) {
    return {
      totalItems: 0,
      readyItemsCount: 0,
      backorderItemsCount: 0,
      emptyItemsCount: 0,
      totalOrderedPieces: 0,
      totalReadyPieces: 0,
      totalBackorderPieces: 0,
      overallPercentage: 0,
    }
  }

  let totalOrderedPieces = 0
  let totalReadyPieces = 0
  let totalBackorderPieces = 0
  let readyItemsCount = 0
  let backorderItemsCount = 0
  let emptyItemsCount = 0

  items.forEach((item) => {
    totalOrderedPieces += item.totalPieces
    totalReadyPieces += item.readyPieces
    totalBackorderPieces += item.backorderPieces

    if (item.statusType === 'READY') readyItemsCount++
    else if (item.statusType === 'PARTIAL') backorderItemsCount++
    else if (item.statusType === 'EMPTY') emptyItemsCount++
  })

  const overallPercentage =
    totalOrderedPieces > 0 ? Math.min(100, Math.round((totalReadyPieces / totalOrderedPieces) * 100)) : 100

  return {
    totalItems: items.length,
    readyItemsCount,
    backorderItemsCount,
    emptyItemsCount,
    totalOrderedPieces,
    totalReadyPieces,
    totalBackorderPieces,
    overallPercentage,
  }
})

async function handleRecheck() {
  if (!props.order) return
  isRechecking.value = true
  try {
    const res = await soStore.recheckStock(props.order.id)
    toast.success(res.message || 'Pengecekan stok berhasil')
    await fetchBranchStock()
    emit('refresh')
    if (res.data?.status === 'CONFIRMED') {
      emit('close')
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal memeriksa ketersediaan stok')
  } finally {
    isRechecking.value = false
  }
}

async function handleOverride() {
  if (!props.order) return
  if (!confirm(`Konfirmasi override pesanan ${props.order.so_number}? Status pesanan akan dipaksa menjadi CONFIRMED.`)) {
    return
  }

  isOverriding.value = true
  try {
    await soStore.overrideBackorder(props.order.id)
    toast.success('Pesanan berhasil di-override ke status CONFIRMED')
    emit('refresh')
    emit('close')
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal melakukan override backorder')
  } finally {
    isOverriding.value = false
  }
}

function handleOpenAdjust() {
  if (!props.order) return
  emit('open-adjust', props.order)
  emit('close')
}
</script>

<template>
  <div v-if="show && order" class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
    <!-- Backdrop -->
    <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="emit('close')" />

    <!-- Modal Content -->
    <div
      class="bg-white rounded-[32px] w-full max-w-4xl relative z-10 shadow-2xl flex flex-col max-h-[92vh] overflow-hidden animate-in fade-in zoom-in-95 duration-200"
    >
      <!-- Modal Header -->
      <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-white shrink-0">
        <div class="flex items-center gap-3">
          <div class="w-11 h-11 rounded-2xl bg-orange-500 text-white flex items-center justify-center shadow-lg shadow-orange-500/20 shrink-0">
            <AlertTriangle class="w-6 h-6" />
          </div>
          <div>
            <div class="flex items-center gap-2 flex-wrap">
              <h3 class="text-lg sm:text-xl font-black text-slate-900 tracking-tight">
                Analisis Ketersediaan Stok & Backorder
              </h3>
              <Badge class="bg-orange-600 text-white border-0 text-[10px] font-black uppercase tracking-wider">
                WAITING STOCK
              </Badge>
            </div>
            <p class="text-xs text-slate-500 font-mono mt-0.5">
              No. SO: <span class="font-bold text-slate-800">{{ order.so_number }}</span> • Toko: <span class="font-bold text-slate-800">{{ order.store?.name || 'Toko' }}</span>
            </p>
          </div>
        </div>

        <button
          @click="emit('close')"
          class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900 transition-colors"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Modal Body (Scrollable) -->
      <div class="p-6 overflow-y-auto space-y-6">
        <!-- 1. Order Info Strip -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-slate-50 p-4 rounded-2xl border border-slate-100 text-xs">
          <div class="flex items-center gap-2.5">
            <Store class="w-4 h-4 text-slate-400 shrink-0" />
            <div class="truncate">
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Toko</p>
              <p class="font-bold text-slate-800 truncate" :title="order.store?.name">{{ order.store?.name || '-' }}</p>
            </div>
          </div>
          <div class="flex items-center gap-2.5">
            <User class="w-4 h-4 text-slate-400 shrink-0" />
            <div class="truncate">
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Salesman</p>
              <p class="font-bold text-slate-800 truncate">{{ order.employee?.first_name }} {{ order.employee?.last_name }}</p>
            </div>
          </div>
          <div class="flex items-center gap-2.5">
            <Calendar class="w-4 h-4 text-slate-400 shrink-0" />
            <div>
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Tanggal Order</p>
              <p class="font-bold text-slate-800">{{ formatDate(order.order_date) }}</p>
            </div>
          </div>
          <div class="flex items-center gap-2.5">
            <Building2 class="w-4 h-4 text-slate-400 shrink-0" />
            <div>
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Total Tagihan</p>
              <p class="font-black text-primary">{{ formatCurrency(order.total_amount) }}</p>
            </div>
          </div>
        </div>

        <!-- 2. Global Readiness Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-4 gap-3.5">
          <!-- Readiness Percentage -->
          <div class="bg-gradient-to-br from-orange-500 to-amber-600 text-white rounded-2xl p-4 shadow-md shadow-orange-500/10 flex flex-col justify-between">
            <div class="flex items-center justify-between text-[10px] font-black uppercase tracking-wider text-orange-100">
              <span>Kesiapan Stok</span>
              <Layers class="w-4 h-4" />
            </div>
            <div class="mt-2">
              <p class="text-2xl font-black">{{ readinessSummary.overallPercentage }}%</p>
              <div class="w-full bg-black/20 rounded-full h-1.5 mt-2 overflow-hidden">
                <div class="bg-white h-full transition-all duration-500" :style="{ width: `${readinessSummary.overallPercentage}%` }"></div>
              </div>
            </div>
          </div>

          <!-- Ready Items -->
          <div class="bg-emerald-50/80 border border-emerald-200/80 rounded-2xl p-4 flex flex-col justify-between">
            <div class="flex items-center justify-between text-[10px] font-black uppercase tracking-wider text-emerald-700">
              <span>Item Siap Lengkap</span>
              <CheckCircle2 class="w-4 h-4 text-emerald-600" />
            </div>
            <div class="mt-2">
              <p class="text-2xl font-black text-emerald-950">{{ readinessSummary.readyItemsCount }} <span class="text-xs font-bold text-emerald-600">/ {{ readinessSummary.totalItems }} Produk</span></p>
              <p class="text-[11px] font-bold text-emerald-700 mt-0.5">{{ readinessSummary.totalReadyPieces }} PCS dialokasikan</p>
            </div>
          </div>

          <!-- Shortage Items -->
          <div class="bg-rose-50/80 border border-rose-200/80 rounded-2xl p-4 flex flex-col justify-between">
            <div class="flex items-center justify-between text-[10px] font-black uppercase tracking-wider text-rose-700">
              <span>Item Menunggu Stok</span>
              <Clock class="w-4 h-4 text-rose-600" />
            </div>
            <div class="mt-2">
              <p class="text-2xl font-black text-rose-950">{{ readinessSummary.backorderItemsCount + readinessSummary.emptyItemsCount }} <span class="text-xs font-bold text-rose-600">Produk</span></p>
              <p class="text-[11px] font-bold text-rose-700 mt-0.5">Defisit {{ readinessSummary.totalBackorderPieces }} PCS</p>
            </div>
          </div>

          <!-- Total Deficit Unit -->
          <div class="bg-slate-900 text-white rounded-2xl p-4 flex flex-col justify-between shadow-sm">
            <div class="flex items-center justify-between text-[10px] font-black uppercase tracking-wider text-slate-400">
              <span>Status Alur</span>
              <Package class="w-4 h-4 text-slate-300" />
            </div>
            <div class="mt-2">
              <p class="text-sm font-black text-orange-400 flex items-center gap-1.5">
                <span class="inline-block w-2 h-2 rounded-full bg-orange-400 animate-ping"></span>
                Tertahan Pasokan
              </p>
              <p class="text-[11px] text-slate-300 mt-1">Otomatis confirm saat restock</p>
            </div>
          </div>
        </div>

        <!-- 3. Informational Alert & Hint -->
        <div class="p-4 rounded-2xl bg-amber-50/80 border border-amber-200/80 flex items-start gap-3 text-xs text-amber-900">
          <AlertCircle class="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
          <div class="space-y-1">
            <p class="font-bold">Mekanisme Pemenuhan Stok (Auto-Fulfillment FIFO):</p>
            <p class="text-[11px] text-amber-800 leading-relaxed">
              Kekurangan stok di atas telah dicatat sebagai antrean <strong>BACKORDER</strong>. Setiap kali gudang menerima pasokan baru (dari Pabrik atau Surat Jalan Mutasi), sistem secara otomatis mengalokasikan stok ke pesanan ini. Anda juga dapat memilih <strong>Sesuaikan Pesanan</strong> jika toko bersedia menerima pengiriman sebagian.
            </p>
          </div>
        </div>

        <!-- 4. Detailed Product Comparison Table -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h4 class="text-xs font-black text-slate-900 uppercase tracking-wider flex items-center gap-2">
              <Package class="w-4 h-4 text-primary" />
              Rincian Ketersediaan & Alokasi Batch per Produk
            </h4>
            <span class="text-[11px] text-slate-400 font-medium" v-if="loadingInventory">
              Memeriksa stok gudang terkini...
            </span>
          </div>

          <div class="border border-slate-200 rounded-2xl overflow-hidden shadow-sm bg-white">
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs">
                <thead class="bg-slate-50 border-b border-slate-200 text-[10px] font-black text-slate-500 uppercase tracking-wider">
                  <tr>
                    <th class="py-3 px-4">Produk & SKU</th>
                    <th class="py-3 px-3 text-center">Dipesan</th>
                    <th class="py-3 px-3">Alokasi Siap (Batch Fisik)</th>
                    <th class="py-3 px-3">Kekurangan (Backorder)</th>
                    <th class="py-3 px-3 text-center">Stok Fisik Gudang</th>
                    <th class="py-3 px-4 text-center">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr
                    v-for="item in analyzedItems"
                    :key="item.id"
                    class="hover:bg-slate-50/70 transition-colors"
                  >
                    <!-- 1. Produk -->
                    <td class="py-3.5 px-4 align-top">
                      <p class="font-black text-slate-900 text-sm leading-tight">{{ item.productName }}</p>
                      <p class="text-[10px] text-slate-400 font-mono mt-0.5">{{ item.sku }}</p>
                      <p class="text-[10px] text-slate-500 font-semibold mt-1">
                        {{ formatCurrency(item.price) }} / unit • Total {{ formatCurrency(item.subtotal) }}
                      </p>
                    </td>

                    <!-- 2. Qty Dipesan -->
                    <td class="py-3.5 px-3 text-center align-top">
                      <p class="font-black text-slate-800 text-sm">{{ item.ordered_quantity }} {{ item.unit }}</p>
                      <p class="text-[10px] text-slate-400 font-bold" v-if="item.pieces_per_unit > 1">
                        ({{ item.totalPieces }} PCS)
                      </p>
                    </td>

                    <!-- 3. Alokasi Siap (Batch Fisik) -->
                    <td class="py-3.5 px-3 align-top">
                      <div v-if="item.readyPieces > 0" class="space-y-1.5">
                        <div class="flex items-center gap-1 text-emerald-700 font-black text-xs">
                          <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" />
                          <span>{{ item.readyUnits }} {{ item.unit }} ({{ item.readyPieces }} PCS)</span>
                        </div>
                        <div class="flex flex-wrap gap-1">
                          <span
                            v-for="b in item.realBatches"
                            :key="b.id || b.batch_no"
                            class="text-[9px] font-bold bg-emerald-50 text-emerald-800 border border-emerald-200 px-2 py-0.5 rounded-md"
                          >
                            Batch {{ b.batch_no }}: {{ b.quantity }} PCS
                          </span>
                        </div>
                      </div>
                      <div v-else class="text-[11px] font-bold text-slate-400 italic">
                        Belum ada stok fisik teralokasi
                      </div>
                    </td>

                    <!-- 4. Kekurangan / Backorder -->
                    <td class="py-3.5 px-3 align-top">
                      <div v-if="item.hasBackorder" class="space-y-1.5">
                        <div class="flex items-center gap-1 text-orange-700 font-black text-xs">
                          <Clock class="w-3.5 h-3.5 text-orange-600" />
                          <span>Kurang {{ item.backorderUnits }} {{ item.unit }} ({{ item.backorderPieces }} PCS)</span>
                        </div>
                        <span class="text-[9px] font-bold bg-orange-100 text-orange-800 border border-orange-200 px-2 py-0.5 rounded-md inline-block">
                          ⏳ Antrean Backorder Gudang
                        </span>
                      </div>
                      <div v-else class="text-[11px] font-bold text-emerald-600 flex items-center gap-1">
                        <CheckCircle2 class="w-3.5 h-3.5" />
                        <span>0 (Lengkap 100%)</span>
                      </div>
                    </td>

                    <!-- 5. Stok Fisik Gudang Saat Ini -->
                    <td class="py-3.5 px-3 text-center align-top">
                      <div class="flex flex-col items-center">
                        <span
                          class="font-black text-xs px-2.5 py-0.5 rounded-full"
                          :class="item.totalWarehouseStock > 0 ? 'bg-blue-50 text-blue-700 border border-blue-200' : 'bg-slate-100 text-slate-500'"
                        >
                          {{ item.totalWarehouseStock }} PCS
                        </span>
                      </div>
                    </td>

                    <!-- 6. Status Kesiapan Badge -->
                    <td class="py-3.5 px-4 text-center align-top">
                      <Badge
                        v-if="item.statusType === 'READY'"
                        class="bg-emerald-100 text-emerald-700 border-emerald-200 text-[10px] font-bold shadow-none"
                      >
                        ✅ Siap Lengkap
                      </Badge>
                      <Badge
                        v-else-if="item.statusType === 'PARTIAL'"
                        class="bg-amber-100 text-amber-700 border-amber-200 text-[10px] font-bold shadow-none"
                      >
                        ⚠️ Parsial ({{ item.readyPercentage }}%)
                      </Badge>
                      <Badge
                        v-else
                        class="bg-rose-100 text-rose-700 border-rose-200 text-[10px] font-bold shadow-none"
                      >
                        ❌ Kosong (0%)
                      </Badge>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- Modal Footer (Action Hub) -->
      <div class="bg-slate-50 px-6 py-4 border-t border-slate-100 flex flex-col sm:flex-row items-center justify-between gap-3 shrink-0">
        <div class="flex items-center gap-2 w-full sm:w-auto">
          <!-- Adjust Button -->
          <Button
            variant="outline"
            size="sm"
            @click="handleOpenAdjust"
            class="rounded-xl font-black text-xs gap-1.5 text-violet-700 bg-violet-50 hover:bg-violet-100 border-violet-200 w-full sm:w-auto h-9"
          >
            <SlidersHorizontal class="w-4 h-4" />
            Sesuaikan Pesanan (Adjust SO)
          </Button>

          <!-- Override Backorder Option (Admin emergency) -->
          <Button
            variant="ghost"
            size="sm"
            @click="handleOverride"
            :disabled="isOverriding"
            class="rounded-xl font-bold text-xs gap-1 text-slate-600 hover:text-amber-700 hover:bg-amber-50 h-9 hidden lg:inline-flex"
            title="Paksa status pesanan menjadi CONFIRMED tanpa menunggu stok lengkap"
          >
            <ShieldAlert class="w-4 h-4 text-amber-600" />
            Override Kirim
          </Button>
        </div>

        <div class="flex items-center gap-2.5 w-full sm:w-auto justify-end">
          <Button
            variant="ghost"
            size="sm"
            @click="emit('close')"
            class="rounded-xl font-bold text-xs h-9 px-4"
          >
            Tutup
          </Button>

          <!-- Recheck Stock Button -->
          <Button
            @click="handleRecheck"
            :disabled="isRechecking"
            size="sm"
            class="rounded-xl font-black text-xs bg-orange-600 hover:bg-orange-700 text-white shadow-lg shadow-orange-600/20 gap-2 h-9 px-5 w-full sm:w-auto"
          >
            <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isRechecking }" />
            <span>{{ isRechecking ? 'Memeriksa...' : 'Cek Ulang Ketersediaan Stok' }}</span>
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
