<script setup lang="ts">
import { computed } from 'vue'
import {
  Boxes,
  Package,
  Layers,
  Calendar,
  AlertTriangle,
  X,
  Warehouse,
  CheckCircle2,
  TrendingUp,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface Props {
  items: any[]
  searchQuery?: string
  allStockCount?: number
}

const props = withDefaults(defineProps<Props>(), {
  items: () => [],
  searchQuery: '',
  allStockCount: 0,
})

const emit = defineEmits<{
  (e: 'select-batch', batchNo: string): void
  (e: 'clear-search'): void
}>()

// Helper: Check expired status
const checkExpiryStatus = (dateStr?: string) => {
  if (!dateStr) return { isExpired: false, isExpiringSoon: false, text: '-' }
  const expDate = new Date(dateStr)
  if (isNaN(expDate.getTime())) return { isExpired: false, isExpiringSoon: false, text: '-' }

  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const diffDays = Math.ceil((expDate.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))

  const formattedDate = expDate.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })

  if (diffDays < 0) {
    return { isExpired: true, isExpiringSoon: false, text: `${formattedDate} (Expired)` }
  }
  if (diffDays <= 30) {
    return { isExpired: false, isExpiringSoon: true, text: `${formattedDate} (${diffDays} hari lagi)` }
  }
  return { isExpired: false, isExpiringSoon: false, text: formattedDate }
}

// Group items by product
const productSummaries = computed(() => {
  const map = new Map<string, any>()

  for (const item of props.items) {
    const prodId = item.product_id || item.product?.id || item.product?.name || 'unknown'
    const prodName = item.product?.name || 'Produk Tanpa Nama'
    const sku = item.product?.sku || '-'
    const unit = item.product?.unit || 'PCS'
    const category = item.product?.category || ''
    const pcsPerUnit = item.product?.pcs_per_unit || 1
    const qty = Number(item.quantity) || 0

    const expStatus = checkExpiryStatus(item.expiry_date)

    if (!map.has(prodId)) {
      map.set(prodId, {
        productId: prodId,
        name: prodName,
        sku,
        unit,
        category,
        pcsPerUnit,
        totalQuantity: 0,
        batches: [] as any[],
      })
    }

    const prodEntry = map.get(prodId)!
    prodEntry.totalQuantity += qty
    prodEntry.batches.push({
      id: item.id,
      batchNo: item.batch_no || 'DEFAULT',
      quantity: qty,
      expiryDate: item.expiry_date,
      expiryText: expStatus.text,
      isExpired: expStatus.isExpired,
      isExpiringSoon: expStatus.isExpiringSoon,
      factoryName: item.factory?.name || 'Pabrik',
      warehouseName: item.warehouse?.name || 'Gudang Umum',
    })
  }

  const summaries = Array.from(map.values()).map((p) => {
    // Sort batches: earliest expiry or largest qty
    p.batches.sort((a: any, b: any) => {
      if (a.isExpired !== b.isExpired) return a.isExpired ? -1 : 1
      if (a.isExpiringSoon !== b.isExpiringSoon) return a.isExpiringSoon ? -1 : 1
      return b.quantity - a.quantity
    })

    // Calculate carton info if pcs_per_unit > 1
    let cartonInfo = ''
    if (p.pcsPerUnit > 1 && p.totalQuantity > 0) {
      const cartons = Math.floor(p.totalQuantity / p.pcsPerUnit)
      const remainder = p.totalQuantity % p.pcsPerUnit
      cartonInfo = `~ ${cartons.toLocaleString('id-ID')} Karton/Dus${
        remainder > 0 ? ` + ${remainder} ${p.unit}` : ''
      } (@ ${p.pcsPerUnit} ${p.unit})`
    }

    return {
      ...p,
      cartonInfo,
    }
  })

  // Sort products by total quantity descending
  return summaries.sort((a, b) => b.totalQuantity - a.totalQuantity)
})

const grandTotalQuantity = computed(() => {
  return props.items.reduce((sum, item) => sum + (Number(item.quantity) || 0), 0)
})

const totalBatches = computed(() => {
  return props.items.length
})

const expiredBatchesCount = computed(() => {
  return props.items.filter((item) => checkExpiryStatus(item.expiry_date).isExpired).length
})

const expiringSoonBatchesCount = computed(() => {
  return props.items.filter((item) => checkExpiryStatus(item.expiry_date).isExpiringSoon).length
})
</script>

<template>
  <div>
    <!-- STATE 1: PENCARIAN AKTIF (Search Query Filled) -->
    <div
      v-if="searchQuery.trim()"
      class="bg-gradient-to-br from-indigo-50/80 via-blue-50/50 to-slate-50 border border-blue-200/80 rounded-3xl p-5 shadow-sm space-y-4 transition-all animate-in fade-in slide-in-from-top-2 duration-200"
    >
      <!-- Summary Header -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-3 border-b border-blue-100/80">
        <div class="flex items-start sm:items-center gap-3">
          <div
            class="w-10 h-10 rounded-2xl bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center shadow-md shadow-blue-500/20 shrink-0"
          >
            <Boxes class="w-5 h-5" />
          </div>
          <div>
            <div class="flex items-center gap-2 flex-wrap">
              <h3 class="text-sm font-black text-slate-900 tracking-tight">
                Ringkasan Stok Hasil Pencarian:
              </h3>
              <Badge
                variant="outline"
                class="bg-blue-600 text-white border-none font-bold text-xs px-2.5 py-0.5 rounded-lg shadow-sm"
              >
                "{{ searchQuery }}"
              </Badge>
            </div>
            <p class="text-xs text-slate-500 mt-0.5">
              Ditemukan
              <span class="font-bold text-slate-800">{{ productSummaries.length }}</span> jenis produk
              tersebar dalam
              <span class="font-bold text-slate-800">{{ totalBatches }}</span> data batch.
            </p>
          </div>
        </div>

        <div class="flex items-center gap-3 self-end md:self-center">
          <!-- Total Metrics Pill -->
          <div
            class="flex items-center gap-4 bg-white px-4 py-2 rounded-2xl border border-blue-100 shadow-sm"
          >
            <div>
              <span class="text-[9px] font-black uppercase tracking-wider text-slate-400 block">
                Total Akumulasi
              </span>
              <div class="flex items-baseline gap-1">
                <span class="text-xl font-black text-blue-700 font-mono tracking-tight">
                  {{ grandTotalQuantity.toLocaleString('id-ID') }}
                </span>
                <span class="text-[10px] font-bold text-slate-400 uppercase">Unit</span>
              </div>
            </div>
            <div class="h-7 w-px bg-slate-100"></div>
            <div>
              <span class="text-[9px] font-black uppercase tracking-wider text-slate-400 block">
                Total Batch
              </span>
              <span class="text-base font-black text-slate-800 font-mono">
                {{ totalBatches }}
              </span>
            </div>
          </div>

          <!-- Clear Search Button -->
          <Button
            variant="outline"
            size="sm"
            @click="$emit('clear-search')"
            class="h-9 px-3 rounded-xl border-slate-200 text-slate-600 hover:text-rose-600 hover:bg-rose-50 hover:border-rose-200 text-xs font-bold transition-colors"
            title="Bersihkan filter pencarian"
          >
            <X class="w-3.5 h-3.5 mr-1 text-slate-400 group-hover:text-rose-600" />
            Reset
          </Button>
        </div>
      </div>

      <!-- Warning Badges if Any Expired / Expiring Soon -->
      <div v-if="expiredBatchesCount > 0 || expiringSoonBatchesCount > 0" class="flex flex-wrap gap-2 pt-0.5">
        <Badge
          v-if="expiredBatchesCount > 0"
          variant="outline"
          class="bg-rose-50 text-rose-700 border-rose-200 font-bold text-[11px] px-2.5 py-1 rounded-xl flex items-center gap-1.5"
        >
          <AlertTriangle class="w-3.5 h-3.5" />
          {{ expiredBatchesCount }} batch telah melewati masa kedaluwarsa
        </Badge>
        <Badge
          v-if="expiringSoonBatchesCount > 0"
          variant="outline"
          class="bg-amber-50 text-amber-700 border-amber-200 font-bold text-[11px] px-2.5 py-1 rounded-xl flex items-center gap-1.5"
        >
          <Calendar class="w-3.5 h-3.5" />
          {{ expiringSoonBatchesCount }} batch akan kedaluwarsa dalam &le; 30 hari
        </Badge>
      </div>

      <!-- Empty State if no product matches -->
      <div
        v-if="productSummaries.length === 0"
        class="bg-white rounded-2xl p-6 text-center border border-dashed border-slate-200"
      >
        <Package class="w-8 h-8 text-slate-300 mx-auto mb-2" />
        <p class="text-xs font-bold text-slate-600">
          Tidak ditemukan produk atau batch dengan kata kunci "{{ searchQuery }}"
        </p>
        <p class="text-[11px] text-slate-400 mt-1">
          Coba periksa kembali ejaan nama produk, SKU, atau nomor batch.
        </p>
      </div>

      <!-- Product Summary Cards Grid -->
      <div
        v-else
        class="grid grid-cols-1"
        :class="productSummaries.length === 1 ? 'grid-cols-1' : 'md:grid-cols-2 xl:grid-cols-3 gap-4'"
      >
        <div
          v-for="prod in productSummaries"
          :key="prod.productId"
          class="bg-white rounded-2xl p-4 border border-slate-100 shadow-sm flex flex-col justify-between hover:border-blue-200 transition-all hover:shadow-md"
        >
          <div>
            <!-- Top Card: Title & Aggregated Total -->
            <div class="flex items-start justify-between gap-3">
              <div class="space-y-1">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <Badge
                    variant="outline"
                    class="font-mono text-[10px] font-bold text-blue-600 border-blue-100 bg-blue-50/60 py-0 px-1.5 rounded"
                  >
                    {{ prod.sku }}
                  </Badge>
                  <Badge
                    v-if="prod.category"
                    variant="outline"
                    class="text-[9px] font-semibold text-slate-400 border-slate-100 bg-slate-50 py-0 px-1.5 rounded uppercase"
                  >
                    {{ prod.category }}
                  </Badge>
                </div>
                <h4 class="font-bold text-slate-900 text-sm leading-snug">
                  {{ prod.name }}
                </h4>
              </div>

              <!-- Big Aggregated Quantity -->
              <div class="text-right shrink-0 bg-emerald-50/70 border border-emerald-100 px-3 py-1.5 rounded-xl">
                <span class="text-[9px] font-black uppercase tracking-wider text-emerald-600 block">
                  Total Fisik
                </span>
                <div class="flex items-baseline justify-end gap-1">
                  <span class="text-lg font-black text-emerald-700 font-mono tracking-tight">
                    {{ prod.totalQuantity.toLocaleString('id-ID') }}
                  </span>
                  <span class="text-[10px] font-black text-emerald-600 uppercase">
                    {{ prod.unit }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Carton Conversion info if available -->
            <div
              v-if="prod.cartonInfo"
              class="mt-2 text-[11px] font-medium text-slate-600 bg-slate-50 border border-slate-100 rounded-lg px-2.5 py-1 flex items-center gap-1.5"
            >
              <Package class="w-3.5 h-3.5 text-blue-500 shrink-0" />
              <span>Konversi: <strong class="text-slate-800">{{ prod.cartonInfo }}</strong></span>
            </div>

            <!-- Batches Breakdown -->
            <div class="mt-3 pt-3 border-t border-slate-100 space-y-2">
              <div class="flex items-center justify-between text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                <span class="flex items-center gap-1">
                  <Layers class="w-3 h-3 text-slate-400" />
                  {{ prod.batches.length }} Batch Terdaftar
                </span>
                <span>Klik batch untuk filter</span>
              </div>

              <!-- Batches Pill List -->
              <div class="flex flex-wrap gap-1.5 max-h-40 overflow-y-auto pr-1">
                <button
                  type="button"
                  v-for="b in prod.batches"
                  :key="b.id || b.batchNo"
                  @click="$emit('select-batch', b.batchNo)"
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-xl text-xs font-semibold border transition-all hover:scale-[1.02] cursor-pointer text-left"
                  :class="[
                    b.isExpired
                      ? 'bg-rose-50/90 text-rose-800 border-rose-200 hover:bg-rose-100'
                      : b.isExpiringSoon
                      ? 'bg-amber-50/90 text-amber-800 border-amber-200 hover:bg-amber-100'
                      : 'bg-slate-50 text-slate-700 border-slate-200/70 hover:bg-blue-50 hover:border-blue-200 hover:text-blue-700'
                  ]"
                  :title="`Klik untuk memfilter Batch ${b.batchNo}. Lokasi: ${b.factoryName} - ${b.warehouseName}. Exp: ${b.expiryText}`"
                >
                  <span class="font-mono font-bold">{{ b.batchNo }}</span>
                  <span class="opacity-40">|</span>
                  <span class="font-black text-slate-900 font-mono">{{ b.quantity.toLocaleString('id-ID') }}</span>
                  <span class="text-[9px] opacity-60 uppercase">{{ prod.unit }}</span>
                  <span
                    v-if="b.isExpired"
                    class="w-1.5 h-1.5 rounded-full bg-rose-500"
                    title="Sudah Kedaluwarsa"
                  ></span>
                  <span
                    v-else-if="b.isExpiringSoon"
                    class="w-1.5 h-1.5 rounded-full bg-amber-500"
                    title="Mendekati Kedaluwarsa"
                  ></span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- STATE 2: DEFAULT / NO SEARCH (Compact KPI Stats Overview) -->
    <div
      v-else-if="items.length > 0"
      class="bg-white border border-slate-100 rounded-3xl p-4 shadow-sm"
    >
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-2xl bg-blue-50 text-blue-600 flex items-center justify-center">
            <TrendingUp class="w-4 h-4" />
          </div>
          <div>
            <h4 class="text-xs font-bold text-slate-800">Ringkasan Cepat Inventaris Global</h4>
            <p class="text-[11px] text-slate-400">
              Gunakan kolom pencarian di atas untuk menghitung akumulasi total batch per produk secara spesifik.
            </p>
          </div>
        </div>

        <div class="flex items-center gap-4 text-xs">
          <div class="flex items-center gap-1.5">
            <span class="text-slate-400 font-medium">Total Batch:</span>
            <span class="font-mono font-black text-slate-800">{{ totalBatches }}</span>
          </div>
          <div class="h-3 w-px bg-slate-200"></div>
          <div class="flex items-center gap-1.5">
            <span class="text-slate-400 font-medium">Total Akumulasi Fisik:</span>
            <span class="font-mono font-black text-emerald-600">
              {{ grandTotalQuantity.toLocaleString('id-ID') }} Unit
            </span>
          </div>
          <div v-if="expiredBatchesCount > 0" class="h-3 w-px bg-slate-200"></div>
          <div v-if="expiredBatchesCount > 0" class="flex items-center gap-1 text-rose-600 font-bold">
            <AlertTriangle class="w-3.5 h-3.5" />
            <span>{{ expiredBatchesCount }} Expired</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
