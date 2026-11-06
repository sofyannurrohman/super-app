<template>
  <div class="p-4 md:p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 bg-white p-6 rounded-3xl shadow-xs border border-slate-100"
    >
      <div>
        <div class="flex items-center gap-3">
          <div class="p-2.5 rounded-2xl bg-primary/10 text-primary">
            <ClipboardList class="w-7 h-7" />
          </div>
          <div>
            <h1 class="text-2xl font-black tracking-tight text-slate-900">
              Purchase Request & Tindak Lanjut
            </h1>
            <p class="text-xs md:text-sm text-slate-500 mt-0.5">
              Review permintaan pembelian karyawan dan kebutuhan pabrik, lalu catat tindak lanjut Purchasing.
            </p>
          </div>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2.5 w-full lg:w-auto">
        <!-- Selector Pabrik -->
        <select
          v-model="selectedFactoryId"
          @change="fetchData"
          class="text-xs md:text-sm bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option value="">Semua Permintaan</option>
          <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">
            {{ f.name }}
          </option>
        </select>

        <!-- Refresh -->
        <Button variant="outline" size="sm" class="rounded-xl border-slate-200 h-9 font-bold" @click="fetchData">
          <RefreshCw class="mr-1.5 h-3.5 w-3.5" :class="{ 'animate-spin': factoryStore.loading }" />
          Segarkan
        </Button>

        <!-- Export Excel -->
        <Button
          variant="outline"
          size="sm"
          class="rounded-xl border-emerald-200 bg-emerald-50/50 hover:bg-emerald-100/70 text-emerald-700 h-9 font-bold shadow-2xs"
          @click="showExportModal = true"
        >
          <FileSpreadsheet class="mr-1.5 h-4 w-4 text-emerald-600" />
          Export Excel
        </Button>

        <!-- Buat PR Baru -->
        <Button
          size="sm"
          class="rounded-xl font-bold shadow-sm shadow-primary/20 bg-primary hover:bg-primary/90 text-white h-9"
          @click="openCreatePR"
        >
          <Plus class="mr-1.5 h-4 w-4" /> Buat Request
        </Button>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
      <Card class="bg-white p-4 rounded-2xl border-slate-100 shadow-2xs">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Total Request</span>
          <span class="p-2 rounded-xl bg-slate-100 text-slate-700">
            <ClipboardList class="w-4 h-4" />
          </span>
        </div>
        <div class="mt-2 text-2xl font-black text-slate-900">{{ factoryStore.rawMaterialPRs.length }}</div>
        <div class="text-[11px] text-slate-400 mt-0.5">Semua permohonan pengadaan</div>
      </Card>

      <Card class="bg-white p-4 rounded-2xl border-amber-100 bg-amber-50/20 shadow-2xs">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-amber-700 uppercase tracking-wider">Perlu Tindak Lanjut</span>
          <span class="p-2 rounded-xl bg-amber-100 text-amber-700">
            <Clock class="w-4 h-4" />
          </span>
        </div>
        <div class="mt-2 text-2xl font-black text-amber-700">{{ pendingCount }}</div>
        <div class="text-[11px] text-amber-600 font-medium mt-0.5">Menunggu respon Purchasing</div>
      </Card>

      <Card class="bg-white p-4 rounded-2xl border-emerald-100 bg-emerald-50/20 shadow-2xs">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-emerald-700 uppercase tracking-wider">Sudah Ditindaklanjuti</span>
          <span class="p-2 rounded-xl bg-emerald-100 text-emerald-700">
            <CheckCircle2 class="w-4 h-4" />
          </span>
        </div>
        <div class="mt-2 text-2xl font-black text-emerald-700">{{ processedCount }}</div>
        <div class="text-[11px] text-emerald-600 font-medium mt-0.5">Permintaan telah diproses</div>
      </Card>

      <Card class="bg-white p-4 rounded-2xl border-red-100 bg-red-50/20 shadow-2xs">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-red-700 uppercase tracking-wider">Ditolak</span>
          <span class="p-2 rounded-xl bg-red-100 text-red-700">
            <XCircle class="w-4 h-4" />
          </span>
        </div>
        <div class="mt-2 text-2xl font-black text-red-700">{{ rejectedCount }}</div>
        <div class="text-[11px] text-red-600 font-medium mt-0.5">Permohonan tidak disetujui</div>
      </Card>
    </div>

    <!-- Table Card -->
    <Card class="border-none shadow-xs bg-white rounded-3xl overflow-hidden p-2">
      <!-- Filter Bar -->
      <div class="p-4 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 border-b border-slate-100">
        <div class="flex flex-wrap items-center gap-2.5 w-full sm:w-auto">
          <!-- Search input -->
          <div class="relative w-full sm:w-64">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari No. PR, bahan, pemohon..."
              class="w-full text-xs pl-8 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>

          <!-- Status Filter -->
          <div class="flex items-center gap-2">
            <select
              v-model="prStatusFilter"
              class="text-xs bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-bold focus:outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option value="">Semua Status PR</option>
              <option value="PENDING">Perlu Tindak Lanjut (Pending)</option>
              <option value="PROCESSED">Sudah Ditindaklanjuti</option>
              <option value="REJECTED">Ditolak (Rejected)</option>
            </select>
          </div>
        </div>

        <span class="text-xs text-slate-400 font-medium">
          Menampilkan {{ filteredPRs.length }} dari {{ factoryStore.rawMaterialPRs.length }} permohonan
        </span>
      </div>

      <DataTable
        :columns="prColumns"
        :data="filteredPRs"
        :isLoading="factoryStore.loading"
      >
        <!-- No PR & Pemohon -->
        <template #cell-pr_number="{ row }">
          <div class="flex flex-col py-1">
            <span class="font-black text-slate-900 text-sm font-mono tracking-tight">{{ (row as any).pr_number }}</span>
            <div class="flex items-center gap-1.5 text-xs text-slate-500 font-medium mt-0.5">
              <span>{{ (row as any).created_by_name || 'Karyawan' }}</span>
              <span class="text-slate-300">•</span>
              <span class="text-[11px] text-slate-400">{{ formatDateOnly((row as any).request_date) }}</span>
            </div>
          </div>
        </template>

        <!-- Asal permintaan -->
        <template #cell-factory="{ row }">
          <div class="flex flex-col text-xs font-semibold text-slate-700">
            <span>{{ (row as any).factory?.name || 'Karyawan' }}</span>
          </div>
        </template>

        <!-- Ringkasan Bahan -->
        <template #cell-items_summary="{ row }">
          <div class="text-xs py-1">
            <div class="font-bold text-slate-800 flex items-center gap-1.5">
              <span>{{ (row as any).items?.length || 0 }} Jenis Item</span>
            </div>
            <div class="text-[11px] text-slate-500 mt-0.5 max-w-[280px] line-clamp-2">
              {{
                ((row as any).items || [])
                  .map((it: any) => `${it.item_name || it.product?.name || 'Barang'} (${it.requested_qty} ${it.unit})`)
                  .join(', ') || '-'
              }}
            </div>
          </div>
        </template>

        <!-- Tgl Dibutuhkan -->
        <template #cell-required_date="{ row }">
          <span class="text-xs font-medium text-slate-600">
            {{ (row as any).required_date ? formatDateOnly((row as any).required_date) : '-' }}
          </span>
        </template>

        <!-- Status PR -->
        <template #cell-status="{ row }">
          <Badge
            :variant="getStatusBadgeVariant((row as any).status)"
            class="text-[10px] py-0.5 px-2.5 font-bold uppercase rounded-lg"
          >
            {{ formatStatus((row as any).status) }}
          </Badge>
        </template>

        <!-- Tindak Lanjut Purchasing -->
        <template #cell-tindak_lanjut="{ row }">
          <div class="flex flex-col gap-1 py-1">
            <template v-if="(row as any).status === 'PROCESSED' && (row as any).purchase_order_id">
              <div class="flex items-center gap-1.5">
                <span class="text-[11px] font-mono font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-md border border-emerald-200">
                  {{ (row as any).purchase_order?.po_number || 'PO Dibuat' }}
                </span>
                <span v-if="(row as any).purchase_order?.supplier?.name" class="text-[11px] text-slate-600 font-medium truncate max-w-[140px]">
                  ({{ (row as any).purchase_order?.supplier?.name }})
                </span>
              </div>
              <div v-if="(row as any).purchase_order?.progress_status" class="text-[10px] text-slate-500">
                Status: <span class="font-semibold text-slate-700">{{ formatProgressStatus((row as any).purchase_order.progress_status) }}</span>
              </div>
            </template>

            <template v-else-if="(row as any).status === 'PROCESSED'">
              <span class="text-[11px] font-semibold text-emerald-700">Sudah ditindaklanjuti</span>
              <span v-if="(row as any).review_notes" class="text-[10px] text-slate-500">{{ (row as any).review_notes }}</span>
            </template>

            <template v-else-if="(row as any).status === 'REJECTED'">
              <span class="text-[11px] font-semibold text-red-600">
                Ditolak
              </span>
              <span v-if="(row as any).rejection_reason" class="text-[10px] text-red-500 truncate max-w-[200px]" :title="(row as any).rejection_reason">
                "{{ (row as any).rejection_reason }}"
              </span>
            </template>

            <template v-else>
              <div class="flex items-center gap-1.5 text-amber-600 text-xs font-bold">
                <Clock class="w-3.5 h-3.5" />
                <span>Menunggu Tindak Lanjut</span>
              </div>
            </template>
          </div>
        </template>

        <!-- Aksi -->
        <template #cell-actions="{ row }">
          <div class="flex items-center gap-1.5 justify-end py-1">
            <Button
              size="sm"
              variant="outline"
              class="rounded-xl text-xs font-bold border-slate-200 h-8"
              @click="openPRDetail(row)"
            >
              <Eye class="w-3.5 h-3.5 mr-1 text-slate-500" /> Detail
            </Button>

            <!-- Tombol Tindak Lanjut Langsung: Konversi jadi PO -->
            <Button
              v-if="(row as any).status === 'PENDING' && (row as any).factory_id && (row as any).items?.every((it: any) => it.product_id)"
              size="sm"
              class="rounded-xl text-xs font-bold bg-primary hover:bg-primary/90 text-white shadow-2xs h-8"
              @click="openConvertPO(row)"
            >
              <ShoppingCart class="w-3.5 h-3.5 mr-1" /> Tindak Lanjut / Buat PO
            </Button>

            <Button v-if="(row as any).status === 'PENDING' && !(row as any).factory_id"
              size="sm" class="rounded-xl text-xs h-8" @click="openCompleteModal(row)">
              <CheckCircle2 class="w-3.5 h-3.5 mr-1" /> Catat Tindak Lanjut
            </Button>

            <!-- Tombol Tolak PR -->
            <Button
              v-if="(row as any).status === 'PENDING'"
              size="sm"
              variant="outline"
              class="rounded-xl text-xs font-bold border-red-200 text-red-600 hover:bg-red-50 h-8"
              @click="openRejectModal(row)"
            >
              <X class="w-3.5 h-3.5 mr-1" /> Tolak
            </Button>
          </div>
        </template>
      </DataTable>
    </Card>

    <!-- Dialog Tolak PR -->
    <Dialog :open="showRejectDialog" @update:open="showRejectDialog = $event">
      <DialogContent class="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle class="text-lg font-bold text-slate-900 flex items-center gap-2">
            <AlertTriangle class="w-5 h-5 text-red-600" />
            Tolak Purchase Request
          </DialogTitle>
          <DialogDescription>
            Beri alasan penolakan agar pemohon di gudang/pabrik dapat mengetahui tindak lanjut permohonan ini.
          </DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-2">
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">No. PR</label>
            <input
              type="text"
              disabled
              :value="targetPRForReject?.pr_number"
              class="w-full text-xs font-mono font-bold bg-slate-100 border border-slate-200 rounded-xl px-3 py-2 text-slate-600"
            />
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Alasan Penolakan <span class="text-red-500">*</span></label>
            <textarea
              v-model="rejectionReason"
              rows="3"
              placeholder="Contoh: Stok bahan di gudang pusat masih mencukupi, atau revisi jumlah kuantitas..."
              class="w-full text-xs bg-slate-50 border border-slate-200 rounded-xl p-3 focus:outline-none focus:ring-2 focus:ring-primary/20"
            ></textarea>
          </div>
        </div>
        <DialogFooter class="flex gap-2">
          <Button variant="outline" class="rounded-xl font-bold" @click="showRejectDialog = false">
            Batal
          </Button>
          <Button
            variant="destructive"
            class="rounded-xl font-bold"
            :disabled="!rejectionReason.trim() || isSubmittingReject"
            @click="submitReject"
          >
            {{ isSubmittingReject ? 'Menolak...' : 'Konfirmasi Tolak PR' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog :open="showCompleteDialog" @update:open="showCompleteDialog = $event">
      <DialogContent class="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle>Catat Tindak Lanjut PR</DialogTitle>
          <DialogDescription>Jelaskan tindakan Purchasing untuk {{ targetPRForComplete?.pr_number }}.</DialogDescription>
        </DialogHeader>
        <textarea v-model="completeNotes" rows="4" placeholder="Contoh: Barang telah dipesan ke vendor..."
          class="w-full rounded-xl border border-slate-200 p-3 text-sm"></textarea>
        <DialogFooter>
          <Button variant="outline" @click="showCompleteDialog = false">Batal</Button>
          <Button :disabled="!completeNotes.trim() || isSubmittingComplete" @click="submitComplete">Simpan Tindak Lanjut</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Dialogs Terintegrasi -->
    <CreateRawMaterialPRDialog
      :open="showCreatePRDialog"
      @update:open="showCreatePRDialog = $event"
      :factory-id="selectedFactoryId || (factoryStore.factories[0]?.id || '')"
      :pr="editingPR"
      @success="fetchData"
    />

    <RawMaterialPRDetailDialog
      :open="showDetailDialog"
      @update:open="showDetailDialog = $event"
      :pr="selectedPR"
      @convert-to-po="openConvertPO"
      @complete="openCompleteModal"
      @success="fetchData"
    />

    <CreateRawMaterialPODialog
      :open="showCreatePODialog"
      @update:open="showCreatePODialog = $event"
      :factory-id="targetFactoryIdForPO"
      :prefilled-pr="prefilledPRForPO"
      @success="handlePOSuccess"
    />

    <!-- Modal Export Excel Lengkap & Interaktif -->
    <PurchasingExportExcelModal
      :open="showExportModal"
      @update:open="showExportModal = $event"
      :prs="factoryStore.rawMaterialPRs"
      :pos="factoryStore.rawMaterialPOs"
      :factories="factoryStore.factories"
      :default-factory-id="selectedFactoryId"
      initial-scope="FULL"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import DataTable from '@/components/DataTable.vue'
import {
  ClipboardList,
  Plus,
  RefreshCw,
  Eye,
  ShoppingCart,
  Clock,
  CheckCircle2,
  XCircle,
  X,
  Search,
  FileSpreadsheet,
  AlertTriangle,
} from 'lucide-vue-next'
import { useFactoryStore, type RawMaterialPR } from '@/stores/factory'
import CreateRawMaterialPRDialog from '@/components/factory/CreateRawMaterialPRDialog.vue'
import RawMaterialPRDetailDialog from '@/components/factory/RawMaterialPRDetailDialog.vue'
import CreateRawMaterialPODialog from '@/components/factory/CreateRawMaterialPODialog.vue'
import PurchasingExportExcelModal from '@/components/purchasing/PurchasingExportExcelModal.vue'
import { toast } from 'vue-sonner'

const factoryStore = useFactoryStore()

const selectedFactoryId = ref('')
const prStatusFilter = ref('')
const searchQuery = ref('')

const showCreatePRDialog = ref(false)
const showDetailDialog = ref(false)
const showCreatePODialog = ref(false)
const showRejectDialog = ref(false)
const showCompleteDialog = ref(false)
const showExportModal = ref(false)

const selectedPR = ref<RawMaterialPR | null>(null)
const editingPR = ref<RawMaterialPR | null>(null)
const prefilledPRForPO = ref<RawMaterialPR | null>(null)
const targetPRForReject = ref<RawMaterialPR | null>(null)
const targetPRForComplete = ref<RawMaterialPR | null>(null)
const completeNotes = ref('')
const isSubmittingComplete = ref(false)
const rejectionReason = ref('')
const isSubmittingReject = ref(false)

const targetFactoryIdForPO = computed(() => {
  if (prefilledPRForPO.value?.factory_id) return prefilledPRForPO.value.factory_id
  if (selectedFactoryId.value) return selectedFactoryId.value
  return factoryStore.factories[0]?.id || ''
})

const prColumns = [
  { header: 'No. PR & Pemohon', accessorKey: 'pr_number' },
  { header: 'Asal Permintaan', accessorKey: 'factory' },
  { header: 'Ringkasan Barang', accessorKey: 'items_summary' },
  { header: 'Tgl Dibutuhkan', accessorKey: 'required_date' },
  { header: 'Status PR', accessorKey: 'status' },
  { header: 'Tindak Lanjut Purchasing', accessorKey: 'tindak_lanjut' },
  { header: 'Aksi', accessorKey: 'actions', align: 'right' as const },
]

const pendingCount = computed(() => factoryStore.rawMaterialPRs.filter((p) => p.status === 'PENDING').length)
const processedCount = computed(() => factoryStore.rawMaterialPRs.filter((p) => p.status === 'PROCESSED').length)
const rejectedCount = computed(() => factoryStore.rawMaterialPRs.filter((p) => p.status === 'REJECTED').length)

const filteredPRs = computed(() => {
  return factoryStore.rawMaterialPRs.filter((pr) => {
    // 1. Status Filter
    if (prStatusFilter.value && pr.status !== prStatusFilter.value) {
      return false
    }

    // 2. Search Query
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const matchNumber = (pr.pr_number || '').toLowerCase().includes(q)
      const matchCreator = (pr.created_by_name || '').toLowerCase().includes(q)
      const matchFactory = (pr.factory?.name || '').toLowerCase().includes(q)
      const matchItems = (pr.items || []).some((it: any) =>
        (it.item_name || it.product?.name || '').toLowerCase().includes(q)
      )
      const matchPO = (pr.purchase_order?.po_number || '').toLowerCase().includes(q)
      if (!matchNumber && !matchCreator && !matchFactory && !matchItems && !matchPO) {
        return false
      }
    }

    return true
  })
})

const fetchData = async () => {
  try {
    const params: any = {}
    if (selectedFactoryId.value) {
      params.factory_id = selectedFactoryId.value
    }
    await factoryStore.fetchRawMaterialPRs(params)
    // Juga muat data PO untuk referensi tindak lanjut
    await factoryStore.fetchRawMaterialPOs(params)
  } catch (err: any) {
    console.error('Error fetching purchasing PRs:', err)
  }
}

onMounted(async () => {
  await factoryStore.fetchFactories()
  await fetchData()
})

const openCreatePR = () => {
  editingPR.value = null
  showCreatePRDialog.value = true
}

const openPRDetail = (pr: RawMaterialPR) => {
  selectedPR.value = pr
  showDetailDialog.value = true
}

const openConvertPO = (pr: RawMaterialPR) => {
  prefilledPRForPO.value = pr
  showDetailDialog.value = false
  showCreatePODialog.value = true
}

const openCompleteModal = (pr: RawMaterialPR) => {
  targetPRForComplete.value = pr
  completeNotes.value = ''
  showDetailDialog.value = false
  showCompleteDialog.value = true
}

const submitComplete = async () => {
  if (!targetPRForComplete.value || !completeNotes.value.trim()) return
  isSubmittingComplete.value = true
  try {
    await factoryStore.completeRawMaterialPR(targetPRForComplete.value.id, completeNotes.value.trim())
    showCompleteDialog.value = false
    toast.success('Tindak lanjut Purchase Request tersimpan')
    await fetchData()
  } catch (err: any) {
    toast.error(typeof err === 'string' ? err : err.message || 'Gagal menyimpan tindak lanjut')
  } finally {
    isSubmittingComplete.value = false
  }
}

const handlePOSuccess = async () => {
  showCreatePODialog.value = false
  prefilledPRForPO.value = null
  toast.success('Tindak lanjut berhasil: Purchase Order (PO) telah diterbitkan!')
  await fetchData()
}

const openRejectModal = (pr: RawMaterialPR) => {
  targetPRForReject.value = pr
  rejectionReason.value = ''
  showRejectDialog.value = true
}

const submitReject = async () => {
  if (!targetPRForReject.value) return
  isSubmittingReject.value = true
  try {
    await factoryStore.rejectRawMaterialPR(
      targetPRForReject.value.id,
      rejectionReason.value,
      selectedFactoryId.value || undefined
    )
    toast.success(`Purchase Request ${targetPRForReject.value.pr_number} berhasil ditolak.`)
    showRejectDialog.value = false
    targetPRForReject.value = null
    await fetchData()
  } catch (err: any) {
    toast.error('Gagal menolak PR: ' + (err.message || String(err)))
  } finally {
    isSubmittingReject.value = false
  }
}


const formatDateOnly = (dateStr: any) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const formatStatus = (status?: string) => {
  switch (status) {
    case 'PENDING':
      return 'Perlu Tindak Lanjut'
    case 'PROCESSED':
      return 'Sudah Ditindaklanjuti'
    case 'REJECTED':
      return 'Ditolak'
    case 'CANCELLED':
      return 'Dibatalkan'
    default:
      return status || 'PENDING'
  }
}

const formatProgressStatus = (status?: string) => {
  switch (status) {
    case 'PENDING_APPROVAL':
      return 'Menunggu Persetujuan'
    case 'ORDERED':
      return 'Dipesan ke Vendor'
    case 'PRODUCTION':
      return 'Produksi Vendor'
    case 'SHIPPING':
      return 'Dalam Pengiriman'
    case 'DELIVERED':
      return 'Tiba di Pabrik'
    case 'COMPLETED':
      return 'Selesai'
    default:
      return status || '-'
  }
}

const getStatusBadgeVariant = (status?: string): any => {
  switch (status) {
    case 'PENDING':
      return 'secondary'
    case 'PROCESSED':
      return 'default'
    case 'REJECTED':
      return 'destructive'
    default:
      return 'outline'
  }
}
</script>
