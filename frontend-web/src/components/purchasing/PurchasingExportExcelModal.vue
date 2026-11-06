<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[680px] p-0 overflow-hidden rounded-3xl border-none shadow-2xl">
      <!-- Modal Header -->
      <div class="bg-gradient-to-r from-emerald-600 via-teal-600 to-slate-900 p-6 text-white relative">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-white/10 backdrop-blur-md rounded-2xl border border-white/20 shadow-inner">
            <FileSpreadsheet class="h-6 w-6 text-white" />
          </div>
          <div>
            <h2 class="text-lg font-black tracking-tight">Export Laporan Purchasing & Pengadaan</h2>
            <p class="text-xs text-emerald-100 font-medium mt-0.5">
              Format spreadsheet resmi Karsa HR dengan kop korporat, ringkasan KPI, dan formula akuntansi.
            </p>
          </div>
        </div>
      </div>

      <!-- Modal Body -->
      <div class="p-6 space-y-5 max-h-[75vh] overflow-y-auto">
        <!-- 1. Pilihan Cakupan Laporan (Scope) -->
        <div>
          <label class="block text-xs font-black uppercase tracking-wider text-slate-500 mb-2.5">
            Pilih Cakupan Laporan (Tab Excel) *
          </label>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
            <!-- Full Report -->
            <div
              @click="exportScope = 'FULL'"
              class="relative flex flex-col p-3.5 rounded-2xl border-2 cursor-pointer transition-all duration-200"
              :class="
                exportScope === 'FULL'
                  ? 'border-emerald-600 bg-emerald-50/40 shadow-xs ring-2 ring-emerald-600/10'
                  : 'border-slate-200 bg-white hover:border-slate-300'
              "
            >
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-xs font-black text-slate-900">Lengkap (3 Tab)</span>
                <span
                  class="w-4 h-4 rounded-full flex items-center justify-center text-[10px]"
                  :class="exportScope === 'FULL' ? 'bg-emerald-600 text-white' : 'border border-slate-300'"
                >
                  ✓
                </span>
              </div>
              <p class="text-[11px] text-slate-500 leading-tight">
                Ringkasan Eksekutif, Rekap PR & Tindak Lanjut, serta Histori PO.
              </p>
              <span class="mt-2 text-[9px] font-bold text-emerald-700 bg-emerald-100/70 px-1.5 py-0.5 rounded w-max">
                Rekomendasi
              </span>
            </div>

            <!-- PR Only -->
            <div
              @click="exportScope = 'PR_ONLY'"
              class="relative flex flex-col p-3.5 rounded-2xl border-2 cursor-pointer transition-all duration-200"
              :class="
                exportScope === 'PR_ONLY'
                  ? 'border-emerald-600 bg-emerald-50/40 shadow-xs ring-2 ring-emerald-600/10'
                  : 'border-slate-200 bg-white hover:border-slate-300'
              "
            >
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-xs font-black text-slate-900">Hanya PR</span>
                <span
                  class="w-4 h-4 rounded-full flex items-center justify-center text-[10px]"
                  :class="exportScope === 'PR_ONLY' ? 'bg-emerald-600 text-white' : 'border border-slate-300'"
                >
                  ✓
                </span>
              </div>
              <p class="text-[11px] text-slate-500 leading-tight">
                Daftar permohonan PR, status review, dan nomor PO tindak lanjut.
              </p>
            </div>

            <!-- PO Only -->
            <div
              @click="exportScope = 'PO_ONLY'"
              class="relative flex flex-col p-3.5 rounded-2xl border-2 cursor-pointer transition-all duration-200"
              :class="
                exportScope === 'PO_ONLY'
                  ? 'border-emerald-600 bg-emerald-50/40 shadow-xs ring-2 ring-emerald-600/10'
                  : 'border-slate-200 bg-white hover:border-slate-300'
              "
            >
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-xs font-black text-slate-900">Hanya Histori PO</span>
                <span
                  class="w-4 h-4 rounded-full flex items-center justify-center text-[10px]"
                  :class="exportScope === 'PO_ONLY' ? 'bg-emerald-600 text-white' : 'border border-slate-300'"
                >
                  ✓
                </span>
              </div>
              <p class="text-[11px] text-slate-500 leading-tight">
                Transaksi PO, rincian barang, supplier, progres vendor, dan formula total.
              </p>
            </div>
          </div>
        </div>

        <!-- 2. Filter Pabrik & Status -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Filter Unit / Pabrik</label>
            <select
              v-model="filterFactoryId"
              class="w-full text-xs bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 font-bold focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
            >
              <option value="">Semua Unit Pabrik</option>
              <option v-for="f in factories" :key="f.id" :value="f.id">{{ f.name }}</option>
            </select>
          </div>

          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Filter Status</label>
            <select
              v-model="filterStatus"
              class="w-full text-xs bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 font-bold focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
            >
              <option value="">Semua Status Data</option>
              <option value="ACTIVE">Sedang Berjalan (Pending / On-Delivery)</option>
              <option value="COMPLETED">Selesai (PO Terbit / Barang Diterima)</option>
            </select>
          </div>
        </div>

        <!-- 3. Periode Waktu & Presets -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <label class="text-xs font-black uppercase tracking-wider text-slate-500">
              Periode Waktu
            </label>
            <div class="flex items-center gap-1">
              <button
                v-for="p in datePresets"
                :key="p.id"
                type="button"
                @click="applyPreset(p.id)"
                class="px-2 py-0.5 text-[10px] font-bold rounded-lg transition-all"
                :class="
                  selectedPreset === p.id
                    ? 'bg-emerald-600 text-white shadow-2xs'
                    : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                "
              >
                {{ p.label }}
              </button>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <span class="text-[11px] font-medium text-slate-500">Dari Tanggal</span>
              <input
                type="date"
                v-model="dateFrom"
                class="w-full text-xs bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
              />
            </div>
            <div class="space-y-1">
              <span class="text-[11px] font-medium text-slate-500">Sampai Tanggal</span>
              <input
                type="date"
                v-model="dateTo"
                class="w-full text-xs bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
              />
            </div>
          </div>
        </div>

        <!-- 4. Live Data Preview Scorecards -->
        <div class="bg-slate-50 p-4 rounded-2xl border border-slate-100 space-y-2">
          <div class="flex items-center justify-between text-xs font-bold text-slate-700">
            <span class="flex items-center gap-1.5">
              <Sparkles class="w-4 h-4 text-emerald-600" />
              Estimasi Data yang Akan Diekspor
            </span>
            <span class="text-[11px] text-slate-400 font-medium">Berdasarkan filter saat ini</span>
          </div>

          <div class="grid grid-cols-3 gap-2 pt-1 text-center">
            <div class="bg-white p-2.5 rounded-xl border border-slate-200/80 shadow-2xs">
              <span class="text-[10px] text-slate-400 font-bold block uppercase">Purchase Request</span>
              <span class="text-base font-black text-slate-900">{{ previewStats.prCount }}</span>
              <span class="text-[10px] text-emerald-600 block font-medium">Permohonan</span>
            </div>

            <div class="bg-white p-2.5 rounded-xl border border-slate-200/80 shadow-2xs">
              <span class="text-[10px] text-slate-400 font-bold block uppercase">Purchase Order</span>
              <span class="text-base font-black text-slate-900">{{ previewStats.poCount }}</span>
              <span class="text-[10px] text-blue-600 block font-medium">PO Terbit</span>
            </div>

            <div class="bg-white p-2.5 rounded-xl border border-slate-200/80 shadow-2xs">
              <span class="text-[10px] text-slate-400 font-bold block uppercase">Total Nilai PO</span>
              <span class="text-sm font-black text-emerald-800 font-mono block truncate">
                Rp {{ previewStats.totalAmount.toLocaleString('id-ID') }}
              </span>
              <span class="text-[10px] text-slate-400 block">Akumulasi Belanja</span>
            </div>
          </div>
        </div>

        <!-- 5. Visual Template Preview -->
        <div class="p-3.5 bg-gradient-to-br from-slate-900 to-slate-800 rounded-2xl text-white text-xs space-y-2 shadow-inner">
          <div class="flex items-center justify-between text-[11px] font-bold text-slate-300">
            <span class="flex items-center gap-1.5">
              <CheckCircle class="w-3.5 h-3.5 text-emerald-400" />
              Struktur Sheet Spreadsheet yang Dihasilkan
            </span>
            <span class="text-[10px] text-emerald-400 font-mono">Format .xlsx (Excel)</span>
          </div>

          <div class="flex flex-wrap items-center gap-1.5 pt-0.5">
            <span
              v-if="exportScope === 'FULL'"
              class="px-2.5 py-1 rounded-lg bg-white/10 text-white font-mono text-[10px] font-bold border border-white/10 flex items-center gap-1"
            >
              📄 Tab 1: Ringkasan Eksekutif
            </span>
            <span
              v-if="exportScope === 'FULL' || exportScope === 'PR_ONLY'"
              class="px-2.5 py-1 rounded-lg bg-emerald-500/20 text-emerald-300 font-mono text-[10px] font-bold border border-emerald-500/30 flex items-center gap-1"
            >
              📋 Tab {{ exportScope === 'FULL' ? '2' : '1' }}: Rekap PR & Tindak Lanjut
            </span>
            <span
              v-if="exportScope === 'FULL' || exportScope === 'PO_ONLY'"
              class="px-2.5 py-1 rounded-lg bg-blue-500/20 text-blue-300 font-mono text-[10px] font-bold border border-blue-500/30 flex items-center gap-1"
            >
              🛒 Tab {{ exportScope === 'FULL' ? '3' : '1' }}: Histori PO & Pengadaan
            </span>
          </div>
          <p class="text-[10px] text-slate-400">
            Dilengkapi Freeze Panes (header terkunci saat scroll), AutoFilter aktif, soft pill status badges, dan formula perhitungan otomatis.
          </p>
        </div>

        <!-- Progress Bar (Visible saat exporting) -->
        <div v-if="isExporting" class="p-3.5 bg-emerald-50 rounded-2xl border border-emerald-200 space-y-2 animate-in fade-in duration-200">
          <div class="flex items-center justify-between text-xs font-bold text-emerald-900">
            <span class="flex items-center gap-1.5">
              <Loader2 class="w-3.5 h-3.5 animate-spin text-emerald-600" />
              {{ progressMessage }}
            </span>
            <span class="font-mono text-emerald-700">{{ progressPercent }}%</span>
          </div>
          <div class="w-full bg-emerald-200/60 rounded-full h-2 overflow-hidden">
            <div
              class="bg-emerald-600 h-full rounded-full transition-all duration-300 ease-out"
              :style="{ width: `${progressPercent}%` }"
            ></div>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="p-4 bg-slate-50 border-t border-slate-100 flex items-center justify-between gap-3">
        <Button
          variant="outline"
          class="rounded-xl font-bold border-slate-200"
          :disabled="isExporting"
          @click="$emit('update:open', false)"
        >
          Batal
        </Button>

        <Button
          class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-700 text-white shadow-md shadow-emerald-600/20 px-5"
          :disabled="isExporting || (previewStats.prCount === 0 && previewStats.poCount === 0)"
          @click="startExport"
        >
          <Loader2 v-if="isExporting" class="mr-2 h-4 w-4 animate-spin" />
          <Download v-else class="mr-2 h-4 w-4" />
          {{ isExporting ? 'Mengekspor Laporan...' : 'Unduh File Excel (.xlsx)' }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  Dialog,
  DialogContent,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import {
  FileSpreadsheet,
  Download,
  Loader2,
  Sparkles,
  CheckCircle,
} from 'lucide-vue-next'
import type { RawMaterialPR, RawMaterialPO, Factory } from '@/stores/factory'
import { exportPurchasingReportToExcel } from '@/utils/purchasingExcelReport'
import { toast } from 'vue-sonner'

const props = withDefaults(
  defineProps<{
    open: boolean
    prs: RawMaterialPR[]
    pos: RawMaterialPO[]
    factories: Factory[]
    defaultFactoryId?: string
    initialScope?: 'FULL' | 'PR_ONLY' | 'PO_ONLY'
  }>(),
  {
    open: false,
    defaultFactoryId: '',
    initialScope: 'FULL',
  }
)

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
}>()

const exportScope = ref<'FULL' | 'PR_ONLY' | 'PO_ONLY'>('FULL')
const filterFactoryId = ref('')
const filterStatus = ref('')
const selectedPreset = ref('all')
const dateFrom = ref('')
const dateTo = ref('')

const isExporting = ref(false)
const progressPercent = ref(0)
const progressMessage = ref('Menyiapkan export...')

const datePresets = [
  { id: 'today', label: 'Hari Ini' },
  { id: '7d', label: '7 Hari' },
  { id: 'this_month', label: 'Bulan Ini' },
  { id: 'quarter', label: 'Kuartal Ini' },
  { id: 'all', label: 'Semua' },
]

watch(
  () => props.open,
  (newVal) => {
    if (newVal) {
      exportScope.value = props.initialScope
      filterFactoryId.value = props.defaultFactoryId || ''
      filterStatus.value = ''
      applyPreset('all')
      isExporting.value = false
      progressPercent.value = 0
    }
  }
)

function applyPreset(presetId: string) {
  selectedPreset.value = presetId
  const now = new Date()

  if (presetId === 'today') {
    const todayStr = now.toISOString().slice(0, 10)
    dateFrom.value = todayStr
    dateTo.value = todayStr
  } else if (presetId === '7d') {
    const d = new Date()
    d.setDate(d.getDate() - 7)
    dateFrom.value = d.toISOString().slice(0, 10)
    dateTo.value = now.toISOString().slice(0, 10)
  } else if (presetId === 'this_month') {
    const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1)
    dateFrom.value = startOfMonth.toISOString().slice(0, 10)
    dateTo.value = now.toISOString().slice(0, 10)
  } else if (presetId === 'quarter') {
    const currentQuarter = Math.floor(now.getMonth() / 3)
    const startOfQuarter = new Date(now.getFullYear(), currentQuarter * 3, 1)
    dateFrom.value = startOfQuarter.toISOString().slice(0, 10)
    dateTo.value = now.toISOString().slice(0, 10)
  } else if (presetId === 'all') {
    dateFrom.value = ''
    dateTo.value = ''
  }
}

// Filtered lists for live stat preview & export
const filteredData = computed(() => {
  let prList = [...props.prs]
  let poList = [...props.pos]

  // Filter Factory
  if (filterFactoryId.value) {
    prList = prList.filter((p) => p.factory_id === filterFactoryId.value)
    poList = poList.filter((p) => p.factory_id === filterFactoryId.value)
  }

  // Filter Status
  if (filterStatus.value === 'ACTIVE') {
    prList = prList.filter((p) => p.status === 'PENDING')
    poList = poList.filter((p) => p.status === 'APPROVED' || p.status === 'PARTIAL_RECEIVED' || p.status === 'PENDING_SPV')
  } else if (filterStatus.value === 'COMPLETED') {
    prList = prList.filter((p) => p.status === 'PROCESSED')
    poList = poList.filter((p) => p.status === 'RECEIVED')
  }

  // Filter Date Range
  if (dateFrom.value && dateTo.value) {
    const fromTime = new Date(dateFrom.value).getTime()
    const toTime = new Date(dateTo.value).getTime() + 86400000

    prList = prList.filter((p) => {
      const t = new Date(p.request_date).getTime()
      return t >= fromTime && t <= toTime
    })

    poList = poList.filter((p) => {
      const t = new Date(p.order_date).getTime()
      return t >= fromTime && t <= toTime
    })
  }

  return { prList, poList }
})

const previewStats = computed(() => {
  const { prList, poList } = filteredData.value
  const totalAmount = poList.reduce((acc, p) => acc + (p.total_estimated_amount || 0), 0)
  return {
    prCount: prList.length,
    poCount: poList.length,
    totalAmount,
  }
})

const startExport = async () => {
  isExporting.value = true
  progressPercent.value = 5
  progressMessage.value = 'Memulai proses export...'

  try {
    const { prList, poList } = filteredData.value
    const currentFactory = props.factories.find((f) => f.id === filterFactoryId.value)
    const factoryName = currentFactory ? currentFactory.name : 'Semua Unit Pabrik'

    const fileName = await exportPurchasingReportToExcel({
      prs: prList,
      pos: poList,
      factoryName,
      scope: exportScope.value,
      dateRange: dateFrom.value && dateTo.value ? { from: dateFrom.value, to: dateTo.value } : undefined,
      onProgress: (percent, msg) => {
        progressPercent.value = percent
        progressMessage.value = msg
      },
    })

    toast.success(`Laporan Excel berhasil diunduh: ${fileName}`)
    emit('update:open', false)
  } catch (err: any) {
    console.error('Export failed:', err)
    toast.error('Gagal mengekspor berkas Excel: ' + (err.message || String(err)))
  } finally {
    isExporting.value = false
  }
}
</script>
