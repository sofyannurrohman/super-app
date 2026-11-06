<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="sm:max-w-[850px] max-h-[92vh] flex flex-col p-0 overflow-hidden bg-white border-none shadow-2xl rounded-3xl">
      <!-- Modal Header -->
      <div class="p-6 border-b border-slate-100 bg-slate-50/60 relative shrink-0">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div class="flex items-start sm:items-center gap-4">
            <div
              class="w-12 h-12 rounded-2xl bg-amber-50 border border-amber-200/60 flex items-center justify-center text-amber-600 shadow-sm shrink-0"
            >
              <FileText class="h-6 w-6" />
            </div>
            <div>
              <div class="flex flex-wrap items-center gap-2 mb-1">
                <Badge
                  variant="outline"
                  class="bg-amber-100/80 text-amber-900 border-amber-300 text-[10px] font-black uppercase tracking-wider px-2 py-0.5"
                >
                  Kartu Stok Bahan Baku
                </Badge>
                <Badge
                  v-if="stockItem?.factory_warehouse?.name"
                  variant="outline"
                  class="bg-slate-100 text-slate-700 border-slate-200 text-[10px] font-bold px-2 py-0.5"
                >
                  <Warehouse class="w-3 h-3 mr-1" />
                  {{ stockItem.factory_warehouse.name }}
                </Badge>
                <Badge
                  v-if="stockItem?.product?.category"
                  variant="secondary"
                  class="bg-slate-200 text-slate-800 text-[10px] font-bold px-2 py-0.5"
                >
                  {{ stockItem.product.category }}
                </Badge>
              </div>

              <h2 class="text-xl font-black tracking-tight text-slate-900 flex items-center gap-2">
                {{ stockItem?.product?.name || 'Bahan Baku' }}
                <span
                  v-if="stockItem?.product?.sku"
                  class="text-xs font-mono font-normal text-slate-500"
                >
                  ({{ stockItem.product.sku }})
                </span>
              </h2>

              <div class="flex flex-wrap items-center gap-3 mt-1 text-xs text-slate-500 font-medium">
                <span v-if="stockItem?.batch_no" class="flex items-center gap-1 font-mono font-bold text-slate-700 bg-white px-2 py-0.5 rounded-lg border border-slate-200/60 shadow-2xs">
                  <Tag class="w-3 h-3 text-amber-600" />
                  Batch: {{ stockItem.batch_no }}
                </span>
                <span v-if="stockItem?.product?.unit" class="text-slate-500">
                  Satuan: <strong class="text-slate-800 uppercase">{{ stockItem.product.unit }}</strong>
                </span>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-2 self-end sm:self-center">
            <Button
              variant="outline"
              size="sm"
              @click="loadLogs"
              class="h-9 rounded-xl border-slate-200 hover:bg-slate-100 text-xs font-bold"
              :disabled="loading"
            >
              <RefreshCw class="h-3.5 w-3.5 mr-1.5" :class="{ 'animate-spin': loading }" />
              Segarkan
            </Button>
          </div>
        </div>

        <!-- Quick Summary Metrics -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-5">
          <div class="p-3 bg-white rounded-2xl border border-slate-100 shadow-2xs flex flex-col">
            <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Sisa Stok Fisik</span>
            <div class="flex items-baseline gap-1 mt-1">
              <span class="text-lg font-black text-slate-900 font-mono">
                {{ formatMultiUnit(stockItem?.quantity || 0, stockItem?.product?.unit, stockItem?.product?.pcs_per_unit).formattedHighest }}
              </span>
            </div>
            <span class="text-[10px] text-slate-400 font-mono">
              ({{ (stockItem?.quantity || 0).toLocaleString('id-ID') }} {{ stockItem?.product?.unit || 'PCS' }})
            </span>
          </div>

          <div class="p-3 bg-emerald-50/60 rounded-2xl border border-emerald-100 shadow-2xs flex flex-col">
            <span class="text-[10px] font-bold text-emerald-700 uppercase tracking-wider flex items-center gap-1">
              <ArrowDownLeft class="w-3 h-3 text-emerald-600" /> Total Masuk
            </span>
            <div class="flex items-baseline gap-1 mt-1">
              <span class="text-lg font-black text-emerald-700 font-mono">
                +{{ formatMultiUnit(summaryMetrics.totalIn, stockItem?.product?.unit, stockItem?.product?.pcs_per_unit).formattedHighest }}
              </span>
            </div>
            <span class="text-[10px] text-emerald-600/80 font-mono">
              ({{ summaryMetrics.countIn }}x Transaksi)
            </span>
          </div>

          <div class="p-3 bg-red-50/60 rounded-2xl border border-red-100 shadow-2xs flex flex-col">
            <span class="text-[10px] font-bold text-red-700 uppercase tracking-wider flex items-center gap-1">
              <ArrowUpRight class="w-3 h-3 text-red-600" /> Total Keluar
            </span>
            <div class="flex items-baseline gap-1 mt-1">
              <span class="text-lg font-black text-red-700 font-mono">
                -{{ formatMultiUnit(summaryMetrics.totalOut, stockItem?.product?.unit, stockItem?.product?.pcs_per_unit).formattedHighest }}
              </span>
            </div>
            <span class="text-[10px] text-red-600/80 font-mono">
              ({{ summaryMetrics.countOut }}x Transaksi)
            </span>
          </div>

          <div class="p-3 bg-blue-50/60 rounded-2xl border border-blue-100 shadow-2xs flex flex-col">
            <span class="text-[10px] font-bold text-blue-700 uppercase tracking-wider flex items-center gap-1">
              <History class="w-3 h-3 text-blue-600" /> Total Mutasi
            </span>
            <div class="flex items-baseline gap-1 mt-1">
              <span class="text-lg font-black text-blue-700 font-mono">{{ logs.length }}</span>
              <span class="text-[10px] font-bold text-blue-600 uppercase ml-1">Catatan</span>
            </div>
            <span class="text-[10px] text-blue-600/80">Histori Pergerakan</span>
          </div>
        </div>
      </div>

      <!-- Filter Controls -->
      <div class="p-4 border-b border-slate-100 flex flex-col sm:flex-row items-center justify-between gap-3 bg-white">
        <div class="flex items-center gap-2 w-full sm:w-auto">
          <!-- Type Filter -->
          <select
            v-model="filterType"
            class="h-9 text-xs font-semibold bg-slate-50 border border-slate-200 rounded-xl px-3 py-1 text-slate-700 focus:outline-none focus:ring-2 focus:ring-amber-500/20"
          >
            <option value="ALL">Semua Tipe Mutasi</option>
            <option value="IN">Hanya Masuk (Penerimaan PO / Tambah)</option>
            <option value="OUT">Hanya Keluar (Pemakaian Produksi / Retur)</option>
            <option value="ADJUSTMENT">Hanya Penyesuaian (Stock Opname)</option>
          </select>

          <!-- Search Input -->
          <div class="relative flex-1 sm:w-60">
            <Search class="absolute left-3 top-2.5 h-3.5 w-3.5 text-slate-400" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari referensi / no batch..."
              class="w-full pl-8 pr-3 py-1.5 text-xs bg-slate-50 border border-slate-200 rounded-xl font-medium focus:outline-none focus:ring-2 focus:ring-amber-500/20"
            />
          </div>
        </div>

        <!-- Format Unit Selector -->
        <div class="flex items-center gap-1 text-xs">
          <span class="text-[10px] font-bold uppercase text-slate-400 mr-1">Unit:</span>
          <button
            type="button"
            @click="unitDisplayMode = 'highest'"
            :class="[
              'px-2.5 py-1 rounded-lg font-bold text-[11px] transition-all',
              unitDisplayMode === 'highest'
                ? 'bg-slate-900 text-white'
                : 'text-slate-600 hover:bg-slate-100'
            ]"
          >
            Tertinggi
          </button>
          <button
            type="button"
            @click="unitDisplayMode = 'base'"
            :class="[
              'px-2.5 py-1 rounded-lg font-bold text-[11px] transition-all',
              unitDisplayMode === 'base'
                ? 'bg-slate-900 text-white'
                : 'text-slate-600 hover:bg-slate-100'
            ]"
          >
            Asli
          </button>
        </div>
      </div>

      <!-- Logs Table Content -->
      <div class="flex-1 overflow-y-auto max-h-[420px] p-4 bg-slate-50/30">
        <div class="border border-slate-200/80 rounded-2xl overflow-hidden bg-white shadow-2xs">
          <table class="w-full text-left border-collapse text-xs">
            <thead class="sticky top-0 bg-slate-50/90 backdrop-blur-xs border-b border-slate-200 text-[10px] font-black uppercase text-slate-500 tracking-wider">
              <tr>
                <th class="py-3 px-3.5">Waktu Mutasi</th>
                <th class="py-3 px-3">Tipe</th>
                <th class="py-3 px-3 text-right">Jumlah Perubahan</th>
                <th class="py-3 px-3">Batch No</th>
                <th class="py-3 px-3.5">Sumber & Keterangan</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="loading" class="text-center">
                <td colspan="5" class="py-12 text-slate-400">
                  <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2 text-amber-600" />
                  Memuat riwayat mutasi bahan baku...
                </td>
              </tr>

              <tr v-else-if="filteredLogs.length === 0" class="text-center">
                <td colspan="5" class="py-12 text-slate-400">
                  <div class="flex flex-col items-center justify-center">
                    <FileText class="w-8 h-8 text-slate-300 mb-2" />
                    <p class="font-bold text-slate-600">Belum ada riwayat mutasi</p>
                    <p class="text-[11px] text-slate-400 mt-0.5">Tidak ditemukan transaksi log untuk bahan baku ini.</p>
                  </div>
                </td>
              </tr>

              <tr
                v-for="log in filteredLogs"
                :key="log.id"
                class="hover:bg-slate-50/80 transition-colors"
              >
                <!-- Waktu -->
                <td class="py-2.5 px-3.5 font-medium text-slate-600 whitespace-nowrap">
                  <div class="flex flex-col">
                    <span class="font-bold text-slate-800">{{ formatDateOnly(log.created_at) }}</span>
                    <span class="text-[10px] text-slate-400 font-mono">{{ formatTimeOnly(log.created_at) }}</span>
                  </div>
                </td>

                <!-- Tipe -->
                <td class="py-2.5 px-3">
                  <Badge
                    :variant="
                      log.type === 'IN' ? 'default' :
                      log.type === 'OUT' ? 'destructive' : 'secondary'
                    "
                    class="rounded-lg text-[10px] py-0.5 px-2 font-black uppercase shadow-none"
                  >
                    {{ log.type === 'IN' ? 'MASUK' : log.type === 'OUT' ? 'KELUAR' : log.type }}
                  </Badge>
                </td>

                <!-- Kuantitas Perubahan -->
                <td class="py-2.5 px-3 text-right">
                  <div class="flex flex-col items-end">
                    <span
                      class="font-mono font-black text-xs"
                      :class="[
                        log.type === 'IN' ? 'text-emerald-600' :
                        log.type === 'OUT' ? 'text-red-600' : 'text-slate-800'
                      ]"
                    >
                      {{ log.type === 'IN' ? '+' : log.type === 'OUT' ? '-' : '' }}
                      <template v-if="unitDisplayMode === 'highest'">
                        {{ formatMultiUnit(Math.abs(log.quantity), stockItem?.product?.unit, stockItem?.product?.pcs_per_unit).formattedHighest }}
                      </template>
                      <template v-else>
                        {{ Math.abs(log.quantity).toLocaleString('id-ID') }} {{ stockItem?.product?.unit || 'PCS' }}
                      </template>
                    </span>
                    <span
                      v-if="unitDisplayMode === 'highest' && formatMultiUnit(Math.abs(log.quantity), stockItem?.product?.unit, stockItem?.product?.pcs_per_unit).category !== 'other'"
                      class="text-[10px] text-slate-400 font-mono"
                    >
                      ({{ Math.abs(log.quantity).toLocaleString('id-ID') }} {{ stockItem?.product?.unit || 'PCS' }})
                    </span>
                  </div>
                </td>

                <!-- Batch No -->
                <td class="py-2.5 px-3">
                  <span class="font-mono font-bold text-[11px] text-slate-700 bg-slate-100 px-2 py-0.5 rounded-md">
                    {{ log.batch_no || 'DEFAULT' }}
                  </span>
                </td>

                <!-- Sumber & Ref -->
                <td class="py-2.5 px-3.5">
                  <div class="flex flex-col">
                    <span class="font-bold text-slate-900 text-xs">
                      {{ formatSourceTitle(log.source) }}
                    </span>
                    <span v-if="log.source" class="text-[10px] text-slate-500 font-mono">
                      {{ log.source }}
                    </span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 border-t border-slate-100 bg-slate-50 flex items-center justify-between shrink-0">
        <div class="text-xs text-slate-500">
          Menampilkan <strong class="text-slate-800">{{ filteredLogs.length }}</strong> dari <strong class="text-slate-800">{{ logs.length }}</strong> catatan mutasi
        </div>
        <Button variant="outline" class="rounded-xl font-bold" @click="emit('update:open', false)">
          Tutup
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Dialog, DialogContent } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  FileText,
  RefreshCw,
  Tag,
  Warehouse,
  ArrowDownLeft,
  ArrowUpRight,
  History,
  Search,
  Loader2,
} from 'lucide-vue-next'
import { useFactoryStore } from '@/stores/factory'
import { formatMultiUnit } from '@/utils/productUnits'

const props = defineProps<{
  open: boolean
  factoryId: string
  stockItem: any | null
}>()

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
}>()

const factoryStore = useFactoryStore()
const loading = ref(false)
const logs = ref<any[]>([])
const filterType = ref('ALL')
const searchQuery = ref('')
const unitDisplayMode = ref<'highest' | 'base'>('highest')

const summaryMetrics = computed(() => {
  let totalIn = 0
  let totalOut = 0
  let countIn = 0
  let countOut = 0

  for (const log of logs.value) {
    const qty = Math.abs(log.quantity || 0)
    if (log.type === 'IN') {
      totalIn += qty
      countIn++
    } else if (log.type === 'OUT') {
      totalOut += qty
      countOut++
    }
  }

  return {
    totalIn,
    totalOut,
    countIn,
    countOut,
  }
})

const filteredLogs = computed(() => {
  return logs.value.filter((log) => {
    if (filterType.value !== 'ALL' && log.type !== filterType.value) {
      return false
    }
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const matchBatch = log.batch_no?.toLowerCase().includes(q)
      const matchSource = log.source?.toLowerCase().includes(q)
      if (!matchBatch && !matchSource) {
        return false
      }
    }
    return true
  })
})

function formatDateOnly(dt: string | Date) {
  if (!dt) return '-'
  return new Date(dt).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function formatTimeOnly(dt: string | Date) {
  if (!dt) return ''
  return new Date(dt).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatSourceTitle(source?: string) {
  if (!source) return 'Penyesuaian Manual'
  const src = source.toUpperCase()
  if (src.startsWith('PO_RECEIPT') || src.includes('RECEIPT')) {
    return 'Penerimaan PO Supplier'
  }
  if (src.startsWith('PRODUCTION') || src.includes('MASAK') || src.includes('PRODUKSI')) {
    return 'Pemakaian Produksi'
  }
  if (src.includes('OPNAME') || src.includes('ADJUSTMENT')) {
    return 'Stock Opname / Saldo Awal'
  }
  if (src.includes('DELETE') || src.includes('REMOVE')) {
    return 'Penghapusan Saldo Stok'
  }
  if (src.includes('MUTATION')) {
    return 'Mutasi Manual Gudang'
  }
  return source
}

async function loadLogs() {
  if (!props.factoryId || !props.stockItem?.product_id) return
  loading.value = true
  try {
    const res = await factoryStore.fetchRawMaterialLogs(props.factoryId, {
      product_id: props.stockItem.product_id,
    })
    logs.value = res || []
  } catch (err) {
    logs.value = []
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.open, props.stockItem],
  ([newOpen, newItem]) => {
    if (newOpen && newItem) {
      filterType.value = 'ALL'
      searchQuery.value = ''
      loadLogs()
    } else {
      logs.value = []
    }
  },
  { immediate: true }
)
</script>
