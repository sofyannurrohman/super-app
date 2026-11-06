<template>
  <div class="p-4 md:p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 bg-white p-6 rounded-3xl shadow-xs border border-slate-100"
    >
      <div>
        <div class="flex items-center gap-3">
          <div class="p-2.5 rounded-2xl bg-emerald-50 text-emerald-700">
            <History class="w-7 h-7" />
          </div>
          <div>
            <h1 class="text-2xl font-black tracking-tight text-slate-900">
              History Purchasing & Pengadaan
            </h1>
            <p class="text-xs md:text-sm text-slate-500 mt-0.5">
              Histori transaksi pesanan pembelian (PO), pelacakan progres vendor, status kedatangan barang, dan arsip pengadaan.
            </p>
          </div>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2.5 w-full lg:w-auto">
        <!-- Selector Pabrik -->
        <select
          v-model="selectedFactoryId"
          @change="fetchPOs"
          class="text-xs md:text-sm bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option value="">Semua Unit Pabrik</option>
          <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">
            {{ f.name }}
          </option>
        </select>

        <!-- Refresh -->
        <Button variant="outline" size="sm" class="rounded-xl border-slate-200 h-9 font-bold" @click="fetchPOs">
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

        <!-- Buat PO Baru -->
        <Button
          size="sm"
          class="rounded-xl font-bold shadow-sm shadow-primary/20 bg-primary hover:bg-primary/90 text-white h-9"
          @click="showCreatePODialog = true"
        >
          <ShoppingCart class="mr-1.5 h-4 w-4" /> Buat PO Baru
        </Button>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
      <Card class="bg-white p-4 rounded-2xl border-slate-100 shadow-2xs">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Total PO Diterbitkan</span>
          <span class="p-2 rounded-xl bg-slate-100 text-slate-700">
            <ShoppingCart class="w-4 h-4" />
          </span>
        </div>
        <div class="mt-2 text-2xl font-black text-slate-900">{{ factoryStore.rawMaterialPOs.length }}</div>
        <div class="text-[11px] text-slate-400 mt-0.5">Semua riwayat purchase order</div>
      </Card>

      <Card class="bg-white p-4 rounded-2xl border-blue-100 bg-blue-50/20 shadow-2xs">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-blue-700 uppercase tracking-wider">Total Nilai Pembelian</span>
          <span class="p-2 rounded-xl bg-blue-100 text-blue-700">
            <Coins class="w-4 h-4" />
          </span>
        </div>
        <div class="mt-2 text-xl font-black text-blue-800 font-mono">
          Rp {{ totalPOAmount.toLocaleString('id-ID') }}
        </div>
        <div class="text-[11px] text-blue-600 font-medium mt-0.5">Akumulasi estimasi nilai PO</div>
      </Card>

      <Card class="bg-white p-4 rounded-2xl border-emerald-100 bg-emerald-50/20 shadow-2xs">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-emerald-700 uppercase tracking-wider">Selesai Diterima</span>
          <span class="p-2 rounded-xl bg-emerald-100 text-emerald-700">
            <CheckCircle2 class="w-4 h-4" />
          </span>
        </div>
        <div class="mt-2 text-2xl font-black text-emerald-700">{{ receivedCount }}</div>
        <div class="text-[11px] text-emerald-600 font-medium mt-0.5">Barang telah tiba di gudang</div>
      </Card>

      <Card class="bg-white p-4 rounded-2xl border-amber-100 bg-amber-50/20 shadow-2xs">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold text-amber-700 uppercase tracking-wider">Sedang Berjalan</span>
          <span class="p-2 rounded-xl bg-amber-100 text-amber-700">
            <Truck class="w-4 h-4" />
          </span>
        </div>
        <div class="mt-2 text-2xl font-black text-amber-700">{{ inProgressCount }}</div>
        <div class="text-[11px] text-amber-600 font-medium mt-0.5">Dalam proses vendor/kirim</div>
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
              placeholder="Cari No. PO, Supplier, PR, bahan..."
              class="w-full text-xs pl-8 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>

          <!-- Status Filter -->
          <div class="flex items-center gap-2">
            <select
              v-model="poStatusFilter"
              class="text-xs bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-bold focus:outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option value="">Semua Status PO</option>
              <option value="PENDING_SPV">Menunggu SPV Gudang</option>
              <option value="PENDING_PURCHASING">Menunggu Purchasing</option>
              <option value="APPROVED">Disetujui</option>
              <option value="PARTIAL_RECEIVED">Diterima Sebagian</option>
              <option value="RECEIVED">Selesai Diterima</option>
              <option value="REJECTED">Ditolak</option>
            </select>
          </div>
        </div>

        <span class="text-xs text-slate-400 font-medium">
          Menampilkan {{ filteredPOs.length }} dari {{ factoryStore.rawMaterialPOs.length }} purchase order
        </span>
      </div>

      <DataTable
        :columns="poColumns"
        :data="filteredPOs"
        :isLoading="factoryStore.loading"
      >
        <!-- No PO & Tanggal -->
        <template #cell-po_number="{ row }">
          <div class="flex flex-col py-1">
            <span class="font-black text-slate-900 text-sm font-mono">{{ (row as any).po_number }}</span>
            <div class="flex items-center gap-1.5 text-xs text-slate-500 font-medium mt-0.5">
              <span>{{ (row as any).supplier?.name || '-' }}</span>
            </div>
            <div v-if="(row as any).purchase_request?.pr_number" class="mt-0.5">
              <span class="text-[10px] font-mono text-primary font-semibold bg-primary/10 px-1.5 py-0.5 rounded">
                Ref PR: {{ (row as any).purchase_request.pr_number }}
              </span>
            </div>
          </div>
        </template>

        <!-- Pabrik / Unit -->
        <template #cell-factory="{ row }">
          <div class="flex flex-col text-xs font-semibold text-slate-700">
            <span>{{ (row as any).factory?.name || '-' }}</span>
            <span class="text-[11px] text-slate-400 font-medium">{{ formatDateOnly((row as any).order_date) }}</span>
          </div>
        </template>

        <!-- Ringkasan Bahan & Kuantitas -->
        <template #cell-items_summary="{ row }">
          <div class="text-xs py-1">
            <div class="font-bold text-slate-800">
              {{ (row as any).items?.length || 0 }} Jenis Bahan
            </div>
            <div class="text-[11px] text-slate-500 mt-0.5 max-w-[260px] line-clamp-2">
              {{
                ((row as any).items || [])
                  .map(
                    (it: any) =>
                      `${it.product?.name || 'Bahan'} (${it.received_qty || 0}/${it.requested_qty} ${it.unit})`
                  )
                  .join(', ') || '-'
              }}
            </div>
          </div>
        </template>

        <!-- Total Nilai Estimasi -->
        <template #cell-total_estimated_amount="{ row }">
          <div class="flex flex-col py-1">
            <span class="font-mono font-bold text-sm text-slate-900">
              Rp {{ ((row as any).total_estimated_amount || 0).toLocaleString('id-ID') }}
            </span>
            <span class="text-[10px] text-slate-400">
              Status Bayar: <strong class="text-slate-600">{{ (row as any).payment_status || 'UNPAID' }}</strong>
            </span>
          </div>
        </template>

        <!-- Status PO -->
        <template #cell-status="{ row }">
          <div class="flex flex-col gap-0.5 items-start py-1">
            <Badge
              :variant="getStatusBadgeVariant((row as any).status)"
              class="text-[10px] py-0.5 px-2.5 font-bold uppercase rounded-lg"
            >
              {{ formatStatus((row as any).status) }}
            </Badge>
            <div v-if="(row as any).purchasing_approved_by_name" class="text-[10px] text-slate-500">
              Purch: <span class="font-semibold text-slate-700">{{ (row as any).purchasing_approved_by_name }}</span>
            </div>
          </div>
        </template>

        <!-- Progress Status Vendor -->
        <template #cell-progress_status="{ row }">
          <div class="flex flex-col gap-0.5 py-1">
            <Badge variant="outline" class="text-[10px] py-0.5 px-2 font-bold bg-indigo-50/80 text-indigo-700 border-indigo-200">
              {{ formatProgressStatus((row as any).progress_status) }}
            </Badge>
            <span v-if="(row as any).estimated_arrival" class="text-[10px] text-slate-500 font-medium">
              ETA: {{ formatDateOnly((row as any).estimated_arrival) }}
            </span>
          </div>
        </template>

        <!-- Aksi -->
        <template #cell-actions="{ row }">
          <div class="flex items-center gap-1.5 justify-end py-1">
            <Button
              size="sm"
              variant="outline"
              class="rounded-xl text-xs font-bold border-slate-200 h-8"
              @click="openPODetail(row)"
            >
              <Eye class="w-3.5 h-3.5 mr-1 text-slate-500" /> Detail PO
            </Button>
            <Button
              size="sm"
              variant="outline"
              class="rounded-xl text-xs font-bold border-indigo-200 text-indigo-700 hover:bg-indigo-50 h-8"
              @click="openUpdateProgress(row)"
            >
              <Truck class="w-3.5 h-3.5 mr-1" /> Progres
            </Button>
          </div>
        </template>
      </DataTable>
    </Card>

    <!-- Dialog Update Progress & Milestone -->
    <Dialog :open="showProgressDialog" @update:open="showProgressDialog = $event">
      <DialogContent class="sm:max-w-[460px]">
        <DialogHeader>
          <DialogTitle class="text-lg font-bold text-slate-900 flex items-center gap-2">
            <Truck class="w-5 h-5 text-indigo-600" />
            Update Progres Pengadaan PO
          </DialogTitle>
          <DialogDescription>
            Perbarui tahapan produksi atau pengiriman dari vendor ke pabrik.
          </DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-2 text-xs">
          <div class="space-y-1">
            <label class="font-bold text-slate-700">No. PO</label>
            <input
              type="text"
              disabled
              :value="targetPOForProgress?.po_number"
              class="w-full font-mono font-bold bg-slate-100 border border-slate-200 rounded-xl px-3 py-2 text-slate-700"
            />
          </div>
          <div class="space-y-1">
            <label class="font-bold text-slate-700">Tahapan Progres</label>
            <select
              v-model="progressForm.status"
              class="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-bold focus:outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option value="PENDING_APPROVAL">Menunggu Persetujuan</option>
              <option value="ORDERED">Dipesan ke Supplier</option>
              <option value="PRODUCTION">Sedang Diproduksi Vendor</option>
              <option value="SHIPPING">Dalam Pengiriman ke Pabrik</option>
              <option value="DELIVERED">Tiba di Pabrik (Siap Bongkar)</option>
              <option value="COMPLETED">Selesai</option>
            </select>
          </div>
          <div class="space-y-1">
            <label class="font-bold text-slate-700">Estimasi Tiba (ETA)</label>
            <input
              type="date"
              v-model="progressForm.eta"
              class="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>
          <div class="space-y-1">
            <label class="font-bold text-slate-700">Catatan Progres</label>
            <textarea
              v-model="progressForm.notes"
              rows="3"
              placeholder="Contoh: Barang sudah dikirim via ekspedisi armada vendor..."
              class="w-full bg-slate-50 border border-slate-200 rounded-xl p-3 focus:outline-none focus:ring-2 focus:ring-primary/20"
            ></textarea>
          </div>
        </div>
        <DialogFooter class="flex gap-2">
          <Button variant="outline" class="rounded-xl font-bold" @click="showProgressDialog = false">
            Batal
          </Button>
          <Button
            class="rounded-xl font-bold bg-indigo-600 hover:bg-indigo-700 text-white"
            :disabled="isUpdatingProgress"
            @click="submitUpdateProgress"
          >
            {{ isUpdatingProgress ? 'Menyimpan...' : 'Simpan Progres' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Dialogs Terintegrasi -->
    <RawMaterialPODetailDialog
      :open="showPODetailDialog"
      @update:open="showPODetailDialog = $event"
      :po="selectedPO"
      @success="fetchPOs"
    />

    <CreateRawMaterialPODialog
      :open="showCreatePODialog"
      @update:open="showCreatePODialog = $event"
      :factory-id="selectedFactoryId || (factoryStore.factories[0]?.id || '')"
      @success="handlePOCreated"
    />

    <!-- Modal Export Excel Lengkap & Interaktif -->
    <PurchasingExportExcelModal
      :open="showExportModal"
      @update:open="showExportModal = $event"
      :prs="factoryStore.rawMaterialPRs"
      :pos="factoryStore.rawMaterialPOs"
      :factories="factoryStore.factories"
      :default-factory-id="selectedFactoryId"
      initial-scope="PO_ONLY"
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
  History,
  ShoppingCart,
  RefreshCw,
  Eye,
  Truck,
  Coins,
  CheckCircle2,
  Search,
  FileSpreadsheet,
} from 'lucide-vue-next'
import { useFactoryStore, type RawMaterialPO } from '@/stores/factory'
import RawMaterialPODetailDialog from '@/components/factory/RawMaterialPODetailDialog.vue'
import CreateRawMaterialPODialog from '@/components/factory/CreateRawMaterialPODialog.vue'
import PurchasingExportExcelModal from '@/components/purchasing/PurchasingExportExcelModal.vue'
import { toast } from 'vue-sonner'

const factoryStore = useFactoryStore()

const selectedFactoryId = ref('')
const poStatusFilter = ref('')
const searchQuery = ref('')

const showPODetailDialog = ref(false)
const showCreatePODialog = ref(false)
const showProgressDialog = ref(false)
const showExportModal = ref(false)

const selectedPO = ref<RawMaterialPO | null>(null)
const targetPOForProgress = ref<RawMaterialPO | null>(null)
const isUpdatingProgress = ref(false)

const progressForm = ref({
  status: 'PRODUCTION',
  eta: '',
  notes: '',
})

const poColumns = [
  { header: 'No. PO & Supplier', accessorKey: 'po_number' },
  { header: 'Unit Pabrik & Tgl', accessorKey: 'factory' },
  { header: 'Ringkasan Bahan & Qty', accessorKey: 'items_summary' },
  { header: 'Total Nilai PO', accessorKey: 'total_estimated_amount' },
  { header: 'Status PO', accessorKey: 'status' },
  { header: 'Progres Pengiriman', accessorKey: 'progress_status' },
  { header: 'Aksi', accessorKey: 'actions', align: 'right' as const },
]

const totalPOAmount = computed(() => {
  return factoryStore.rawMaterialPOs.reduce((acc, p) => acc + (p.total_estimated_amount || 0), 0)
})

const receivedCount = computed(() => {
  return factoryStore.rawMaterialPOs.filter((p) => p.status === 'RECEIVED').length
})

const inProgressCount = computed(() => {
  return factoryStore.rawMaterialPOs.filter(
    (p) => p.status === 'APPROVED' || p.status === 'PARTIAL_RECEIVED'
  ).length
})

const filteredPOs = computed(() => {
  return factoryStore.rawMaterialPOs.filter((po) => {
    // 1. Status Filter
    if (poStatusFilter.value && po.status !== poStatusFilter.value) {
      return false
    }

    // 2. Search Query
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const matchPONumber = (po.po_number || '').toLowerCase().includes(q)
      const matchPRNumber = (po.purchase_request?.pr_number || '').toLowerCase().includes(q)
      const matchSupplier = (po.supplier?.name || '').toLowerCase().includes(q)
      const matchFactory = (po.factory?.name || '').toLowerCase().includes(q)
      const matchItems = (po.items || []).some((it: any) =>
        (it.product?.name || '').toLowerCase().includes(q)
      )
      if (!matchPONumber && !matchPRNumber && !matchSupplier && !matchFactory && !matchItems) {
        return false
      }
    }

    return true
  })
})

const fetchPOs = async () => {
  try {
    const params: any = {}
    if (selectedFactoryId.value) {
      params.factory_id = selectedFactoryId.value
    }
    await factoryStore.fetchRawMaterialPOs(params)
    // Ambil data PR untuk cross-reference excel
    await factoryStore.fetchRawMaterialPRs(params)
  } catch (err: any) {
    console.error('Error fetching purchasing POs:', err)
  }
}

onMounted(async () => {
  await factoryStore.fetchFactories()
  await fetchPOs()
})

const openPODetail = (po: RawMaterialPO) => {
  selectedPO.value = po
  showPODetailDialog.value = true
}

const openUpdateProgress = (po: RawMaterialPO) => {
  targetPOForProgress.value = po
  progressForm.value = {
    status: po.progress_status || 'PRODUCTION',
    eta: po.estimated_arrival ? po.estimated_arrival.slice(0, 10) : '',
    notes: po.progress_notes || '',
  }
  showProgressDialog.value = true
}

const submitUpdateProgress = async () => {
  if (!targetPOForProgress.value) return
  isUpdatingProgress.value = true
  try {
    await factoryStore.updatePOProgress(
      targetPOForProgress.value.id,
      {
        progress_status: progressForm.value.status,
        estimated_arrival: progressForm.value.eta || undefined,
        progress_notes: progressForm.value.notes || undefined,
      },
      selectedFactoryId.value || undefined
    )
    toast.success('Progres pengadaan PO berhasil diperbarui!')
    showProgressDialog.value = false
    targetPOForProgress.value = null
    await fetchPOs()
  } catch (err: any) {
    toast.error('Gagal memperbarui progres PO: ' + (err.message || String(err)))
  } finally {
    isUpdatingProgress.value = false
  }
}

const handlePOCreated = async () => {
  showCreatePODialog.value = false
  toast.success('Purchase Order berhasil diterbitkan!')
  await fetchPOs()
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
    case 'PENDING_SPV':
      return 'Menunggu SPV'
    case 'PENDING_PURCHASING':
      return 'Menunggu Purchasing'
    case 'APPROVED':
      return 'Disetujui'
    case 'PARTIAL_RECEIVED':
      return 'Diterima Sebagian'
    case 'RECEIVED':
      return 'Selesai Diterima'
    case 'REJECTED':
      return 'Ditolak'
    default:
      return status || 'DRAFT'
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
      return status || 'Menunggu Persetujuan'
  }
}

const getStatusBadgeVariant = (status?: string): any => {
  switch (status) {
    case 'APPROVED':
      return 'default'
    case 'RECEIVED':
      return 'default'
    case 'PARTIAL_RECEIVED':
      return 'secondary'
    case 'REJECTED':
      return 'destructive'
    default:
      return 'outline'
  }
}
</script>
