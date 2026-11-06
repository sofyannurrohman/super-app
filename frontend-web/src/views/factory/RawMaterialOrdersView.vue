<template>
  <div class="p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-6 rounded-3xl shadow-sm border border-slate-100"
    >
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900">
          PO & Pengadaan Bahan
        </h1>
        <p class="text-sm text-slate-500 mt-0.5">
          Kelola pesanan pembelian (PO) ke supplier, persetujuan SPV & Purchasing, dan progress produksi vendor.
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <select
          v-model="selectedFactoryId"
          class="text-sm bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">{{ f.name }}</option>
        </select>
        <Button variant="outline" class="rounded-xl border-slate-200" @click="fetchPOs">
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': factoryStore.loading }" />
          Segarkan
        </Button>
        <Button
          @click="showCreatePODialog = true"
          :disabled="!selectedFactoryId"
          class="rounded-xl font-bold shadow-md shadow-primary/20 bg-primary hover:bg-primary/90 text-white"
        >
          <ShoppingCart class="mr-2 h-4 w-4" /> Buat PO Bahan Baku
        </Button>
      </div>
    </div>

    <!-- Table Card -->
    <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden p-2">
      <!-- Filter Status -->
      <div class="p-4 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 border-b border-slate-100">
        <div class="flex items-center gap-2">
          <span class="text-xs font-bold text-slate-500">Filter Status:</span>
          <select
            v-model="poStatusFilter"
            @change="fetchPOs"
            class="text-xs bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-bold focus:outline-none focus:ring-2 focus:ring-primary/20"
          >
            <option value="">Semua Status</option>
            <option value="PENDING_SPV">Menunggu SPV Gudang</option>
            <option value="PENDING_PURCHASING">Menunggu Purchasing</option>
            <option value="APPROVED">Disetujui</option>
            <option value="PARTIAL_RECEIVED">Diterima Sebagian</option>
            <option value="RECEIVED">Selesai Diterima</option>
            <option value="REJECTED">Ditolak</option>
          </select>
        </div>
        <span class="text-xs text-slate-400 font-medium">
          Total {{ factoryStore.rawMaterialPOs.length }} pesanan PO
        </span>
      </div>

      <DataTable
        :columns="poColumns"
        :data="factoryStore.rawMaterialPOs"
        :isLoading="factoryStore.loading"
      >
        <template #cell-po_number="{ row }">
          <div class="flex flex-col">
            <span class="font-black text-slate-900 text-sm font-mono">{{ (row as any).po_number }}</span>
            <span class="text-xs text-slate-400 font-medium">Oleh: {{ (row as any).created_by_name || 'Admin' }}</span>
          </div>
        </template>
        <template #cell-supplier="{ row }">
          <span class="font-bold text-slate-700 text-sm">
            {{ (row as any).supplier?.name || '-' }}
          </span>
        </template>
        <template #cell-order_date="{ row }">
          <span class="text-xs font-medium text-slate-600">
            {{ formatDateOnly((row as any).order_date) }}
          </span>
        </template>
        <template #cell-items_summary="{ row }">
          <div class="text-xs text-slate-600">
            <span class="font-bold text-slate-800">{{ (row as any).items?.length || 0 }} Jenis Bahan</span>
            <div class="text-[11px] text-slate-400 truncate max-w-[200px]">
              {{ ((row as any).items || []).map((it: any) => `${it.product?.name} (${it.requested_qty} ${it.unit})`).join(', ') }}
            </div>
          </div>
        </template>
        <template #cell-total_estimated_amount="{ row }">
          <span class="font-mono font-bold text-sm text-slate-900">
            Rp {{ ((row as any).total_estimated_amount || 0).toLocaleString('id-ID') }}
          </span>
        </template>
        <template #cell-status="{ row }">
          <div class="flex flex-col gap-0.5 items-start">
            <Badge
              :variant="getStatusBadgeVariant((row as any).status)"
              class="text-[10px] py-1 px-2.5 font-bold uppercase rounded-lg"
            >
              {{ formatStatus((row as any).status) }}
            </Badge>
            <div v-if="(row as any).spv_approved_by_name" class="text-[9px] text-slate-500 leading-tight mt-0.5">
              SPV: <span class="font-semibold text-slate-700">{{ (row as any).spv_approved_by_name }}</span>
              <span v-if="(row as any).spv_approved_by_position" class="text-slate-400"> ({{ (row as any).spv_approved_by_position }})</span>
            </div>
            <div v-if="(row as any).purchasing_approved_by_name" class="text-[9px] text-slate-500 leading-tight">
              Purch: <span class="font-semibold text-slate-700">{{ (row as any).purchasing_approved_by_name }}</span>
              <span v-if="(row as any).purchasing_approved_by_position" class="text-slate-400"> ({{ (row as any).purchasing_approved_by_position }})</span>
            </div>
          </div>
        </template>
        <template #cell-progress_status="{ row }">
          <div class="flex flex-col gap-0.5">
            <Badge variant="outline" class="text-[10px] py-0 px-2 font-bold bg-indigo-50/80 text-indigo-700 border-indigo-200">
              {{ formatProgressStatus((row as any).progress_status) }}
            </Badge>
            <span v-if="(row as any).estimated_arrival" class="text-[10px] text-slate-500 font-medium">
              ETA: {{ formatDateOnly((row as any).estimated_arrival) }}
            </span>
          </div>
        </template>
        <template #cell-actions="{ row }">
          <div class="flex items-center gap-1.5 justify-end">
            <Button
              size="sm"
              variant="outline"
              class="rounded-xl text-xs font-bold border-slate-200"
              @click="openPODetail(row)"
            >
              <Eye class="w-3.5 h-3.5 mr-1" /> Detail
            </Button>
            <Button
              size="sm" variant="outline"
              class="rounded-xl text-xs font-bold border-blue-200 text-blue-700"
              title="Edit PO"
              @click="openEditPO(row)"
            >Edit</Button>
            <Button
              size="sm" variant="outline"
              class="rounded-xl text-xs font-bold border-red-200 text-red-600"
              title="Hapus PO"
              @click="deletePO(row)"
            >Hapus</Button>
            <Button
              v-if="(row as any).status === 'APPROVED' || (row as any).status === 'PARTIAL_RECEIVED'"
              size="sm"
              class="rounded-xl text-xs font-bold bg-emerald-600 hover:bg-emerald-700 text-white"
              @click="openReceiptDialogWithPO(row)"
            >
              <PackageCheck class="w-3.5 h-3.5 mr-1" /> Terima
            </Button>
          </div>
        </template>
      </DataTable>
    </Card>

    <!-- Dialogs -->
    <CreateRawMaterialPODialog
      :open="showCreatePODialog"
      @update:open="onPODialogChange"
      :factoryId="selectedFactoryId"
      :po="editingPO"
      @success="fetchPOs"
    />

    <RawMaterialPODetailDialog
      v-model:open="showPODetailDialog"
      :po="selectedPO"
      @success="fetchPOs"
      @open-receipt="openReceiptDialogFromDetail"
    />

    <RawMaterialReceiptDialog
      v-model:open="showReceiptDialog"
      :factoryId="selectedFactoryId"
      :initialPO="initialPOForReceipt"
      @success="fetchPOs"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { toast } from 'vue-sonner'
import { useFactoryStore, type RawMaterialPO } from '@/stores/factory'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import DataTable from '@/components/DataTable.vue'
import { RefreshCw, ShoppingCart, PackageCheck, Eye } from 'lucide-vue-next'
import CreateRawMaterialPODialog from '@/components/factory/CreateRawMaterialPODialog.vue'
import RawMaterialPODetailDialog from '@/components/factory/RawMaterialPODetailDialog.vue'
import RawMaterialReceiptDialog from '@/components/factory/RawMaterialReceiptDialog.vue'

const factoryStore = useFactoryStore()
const selectedFactoryId = ref('')
const poStatusFilter = ref('')

const showCreatePODialog = ref(false)
const showPODetailDialog = ref(false)
const showReceiptDialog = ref(false)
const selectedPO = ref<RawMaterialPO | null>(null)
const editingPO = ref<RawMaterialPO | null>(null)
const initialPOForReceipt = ref<RawMaterialPO | null>(null)

const poColumns = [
  { key: 'po_number', label: 'No. PO & Pembuat', sortable: true },
  {
    id: 'supplier',
    header: 'Supplier',
    accessorFn: (row: any) => row.supplier?.name || '',
    sortable: true,
  },
  { key: 'order_date', label: 'Tgl Pengajuan', sortable: true },
  {
    id: 'items_summary',
    header: 'Bahan Baku',
    accessorFn: (row: any) => (row.items || []).map((it: any) => it.product?.name || '').join(' '),
    sortable: false,
  },
  { key: 'total_estimated_amount', label: 'Estimasi Nilai', sortable: true },
  { key: 'status', label: 'Status Approval', sortable: true },
  { key: 'progress_status', label: 'Progress Vendor', sortable: true },
  { key: 'actions', label: 'Aksi', sortable: false },
]

const formatDateOnly = (val?: string) => {
  if (!val) return '-'
  return new Date(val).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const formatProgressStatus = (status?: string) => {
  switch (status) {
    case 'PENDING_APPROVAL':
      return 'Menunggu Approval'
    case 'PO_SENT_TO_SUPPLIER':
      return 'PO ke Vendor'
    case 'IN_PRODUCTION':
      return 'Produksi Vendor'
    case 'IN_PRINTING':
      return 'Proses Cetak'
    case 'SHIPPED':
      return 'Pengiriman'
    case 'PARTIAL_DELIVERED':
      return 'Tiba Sebagian'
    case 'DELIVERED':
      return 'Telah Tiba'
    default:
      return status || '-'
  }
}

const getStatusBadgeVariant = (status?: string) => {
  switch (status) {
    case 'APPROVED':
    case 'RECEIVED':
      return 'default'
    case 'PARTIAL_RECEIVED':
    case 'PENDING_SPV':
    case 'PENDING_PURCHASING':
      return 'secondary'
    case 'REJECTED':
      return 'destructive'
    default:
      return 'outline'
  }
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
    case 'CANCELLED':
      return 'Dibatalkan'
    default:
      return status || '-'
  }
}

const fetchPOs = () => {
  if (!selectedFactoryId.value) return
  factoryStore.fetchRawMaterialPOs({
    factory_id: selectedFactoryId.value,
    status: poStatusFilter.value || undefined,
  })
}

const openPODetail = (po: RawMaterialPO) => {
  selectedPO.value = po
  showPODetailDialog.value = true
}

const openEditPO = (po: RawMaterialPO) => {
  editingPO.value = po
  showCreatePODialog.value = true
}

const onPODialogChange = (open: boolean) => {
  showCreatePODialog.value = open
  if (!open) editingPO.value = null
}

const deletePO = async (po: RawMaterialPO) => {
  if (!confirm(`Hapus PO ${po.po_number}? Tindakan ini akan menghapus data PO secara permanen.`)) return
  try {
    await factoryStore.deleteRawMaterialPO(po.id, selectedFactoryId.value)
    toast.success('Purchase Order berhasil dihapus')
  } catch (err: any) {
    toast.error(err?.response?.data?.error || err?.message || 'Gagal menghapus Purchase Order')
  }
}

const openReceiptDialogWithPO = (po: RawMaterialPO) => {
  initialPOForReceipt.value = po
  showReceiptDialog.value = true
}

const openReceiptDialogFromDetail = (po: RawMaterialPO) => {
  showPODetailDialog.value = false
  openReceiptDialogWithPO(po)
}

onMounted(async () => {
  await factoryStore.fetchFactories()
  if (factoryStore.factories.length > 0 && factoryStore.factories[0]) {
    selectedFactoryId.value = factoryStore.factories[0].id
  }
})

watch(selectedFactoryId, (newId) => {
  if (newId) {
    fetchPOs()
  }
})
</script>
