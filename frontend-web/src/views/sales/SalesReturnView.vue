<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useSalesReturnStore, type SalesReturn } from '@/stores/salesReturn'
import { useSalesTransferStore } from '@/stores/salesTransfer'
import { useMasterDataStore } from '@/stores/masterData'
import { useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  RotateCcw,
  CheckCircle2,
  AlertCircle,
  Search,
  RefreshCw,
  X,
  Box,
  Info,
  Pencil,
  Trash2,
  Plus,
  Minus,
  FileText,
  Truck,
  Building2,
  Package,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import DataTable from '@/components/DataTable.vue'
import { Card } from '@/components/ui/card/index'
import BranchFactoryReturnModal from '@/components/sales/BranchFactoryReturnModal.vue'
import DeliveryReturnModal from '@/components/sales/DeliveryReturnModal.vue'
import SalesStockReturnModal from '@/components/sales/SalesStockReturnModal.vue'
import RefreshDataButton from '@/components/ui/RefreshDataButton.vue'
import { useDataRefresh } from '@/composables/useDataRefresh'
import apiClient from '@/api/axios'

const returnStore = useSalesReturnStore()
const transferStore = useSalesTransferStore()
const masterStore = useMasterDataStore()
const route = useRoute()

const showDeliveryReturnModal = ref(false)
const showSalesStockReturnModal = ref(false)

const isWarehouse = computed(() => route.path.includes('/warehouse'))

const activeBranchId = computed(() => {
  return isWarehouse.value 
    ? masterStore.selectedWarehouseBranchId 
    : masterStore.selectedBranchId
})

const searchQuery = ref('')
const selectedStatus = ref('ALL')
const loading = ref(false)
const activeTab = ref<'nota' | 'rusak' | 'muatan' | 'cabang'>('nota')

// Factory Returns (Cabang ke Pabrik)
const factoryReturns = ref<any[]>([])
const showBranchReturnModal = ref(false)
const isLoadingFactoryReturns = ref(false)

const fetchFactoryReturns = async () => {
  isLoadingFactoryReturns.value = true
  try {
    const res = await apiClient.get('/factory-returns/company')
    factoryReturns.value = res.data || []
  } catch (error) {
    console.error('Failed to load factory returns:', error)
  } finally {
    isLoadingFactoryReturns.value = false
  }
}

const filteredFactoryReturns = computed(() => {
  return factoryReturns.value.filter((r) => {
    const matchStatus = selectedStatus.value === 'ALL' || r.status === selectedStatus.value
    const q = searchQuery.value.toLowerCase()
    const matchSearch =
      !q ||
      r.return_no?.toLowerCase().includes(q) ||
      r.from_branch?.name?.toLowerCase().includes(q) ||
      r.to_factory?.name?.toLowerCase().includes(q) ||
      r.items?.some((it: any) => it.product?.name?.toLowerCase().includes(q))
    return matchStatus && matchSearch
  })
})

// Factory Return Inspection / Receive State
const showFactoryReceiveModal = ref(false)
const isSubmittingFactoryReceive = ref(false)
const selectedFactoryReturn = ref<any>(null)
const factoryReceiveForm = ref({
  receiver_name: '',
  inspected_by_name: '',
  inspection_notes: '',
  items: [] as Array<{
    item_id: string
    product_name: string
    batch_no: string
    unit: string
    sent_qty: number
    return_type: string
    damage_type: string
    damage_reason: string
    received_good_qty: number
    received_bottle_damaged_qty: number
    received_content_damaged_qty: number
    inspection_notes: string
  }>,
})

const openFactoryReceiveModal = (rtn: any) => {
  selectedFactoryReturn.value = rtn
  factoryReceiveForm.value.receiver_name = 'Checker Gudang Pabrik'
  factoryReceiveForm.value.inspected_by_name = 'QC Pabrik'
  factoryReceiveForm.value.inspection_notes = rtn.notes || ''

  factoryReceiveForm.value.items = (rtn.items || []).map((item: any) => {
    const isGood = item.return_type === 'GOOD'
    const isContent =
      item.damage_type === 'PRODUCT_CONTENT' ||
      item.damage_type === 'EXPIRED' ||
      item.damage_type === 'TOTAL_DAMAGE' ||
      (item.damage_reason && item.damage_reason.toLowerCase().includes('bocor'))

    return {
      item_id: item.id,
      product_name: item.product?.name || 'Produk',
      batch_no: item.batch_no || 'DEFAULT',
      unit: item.unit || item.product?.unit || 'PCS',
      sent_qty: item.quantity,
      return_type: item.return_type || 'DAMAGED',
      damage_type: item.damage_type || 'OTHER',
      damage_reason: item.damage_reason || '',
      received_good_qty: isGood ? item.quantity : 0,
      received_bottle_damaged_qty: !isGood && !isContent ? item.quantity : 0,
      received_content_damaged_qty: isContent ? item.quantity : 0,
      inspection_notes: item.damage_reason || '',
    }
  })

  showFactoryReceiveModal.value = true
}

const submitFactoryReceive = async () => {
  if (!selectedFactoryReturn.value) return

  const invalidQty = factoryReceiveForm.value.items.some((i) => {
    const total =
      (i.received_good_qty || 0) +
      (i.received_bottle_damaged_qty || 0) +
      (i.received_content_damaged_qty || 0)
    return total < 0 || total > i.sent_qty
  })

  if (invalidQty) {
    toast.error('Kuantitas fisik diterima tidak boleh negatif atau melebihi kuantitas yang dikirim cabang')
    return
  }

  isSubmittingFactoryReceive.value = true
  try {
    const payload = {
      receiver_name: factoryReceiveForm.value.receiver_name,
      inspected_by_name: factoryReceiveForm.value.inspected_by_name,
      inspection_notes: factoryReceiveForm.value.inspection_notes,
      items: factoryReceiveForm.value.items.map((i) => ({
        item_id: i.item_id,
        received_good_qty: Number(i.received_good_qty) || 0,
        received_bottle_damaged_qty: Number(i.received_bottle_damaged_qty) || 0,
        received_content_damaged_qty: Number(i.received_content_damaged_qty) || 0,
        inspection_notes: i.inspection_notes || '',
      })),
    }

    await apiClient.post(`/factory-returns/${selectedFactoryReturn.value.id}/receive`, payload)
    toast.success('Penerimaan retur cabang berhasil diproses! Stok pabrik dan piutang telah disesuaikan.')
    showFactoryReceiveModal.value = false
    await fetchFactoryReturns()
  } catch (error: any) {
    toast.error(error.response?.data?.error || error.message || 'Gagal memproses penerimaan retur')
  } finally {
    isSubmittingFactoryReceive.value = false
  }
}

const isDirectSaleReturn = (r: any) => {
  return !r.sales_order_id || r.notes?.includes('[DIRECT_SALE_RETURN]') || (r.transaction_id && !r.sales_order?.so_number)
}

// 1. Direct Sale Returns (Admin Nota)
const filteredNotaReturns = computed(() => {
  return returnStore.returns.filter((r) => {
    if (!isDirectSaleReturn(r)) return false
    const matchStatus = selectedStatus.value === 'ALL' || r.status === selectedStatus.value
    const q = searchQuery.value.toLowerCase()
    const matchSearch =
      !q ||
      r.return_no.toLowerCase().includes(q) ||
      r.transaction?.receipt_no?.toLowerCase().includes(q) ||
      r.transaction?.buyer_name?.toLowerCase().includes(q) ||
      r.transaction?.store?.name?.toLowerCase().includes(q) ||
      r.employee?.first_name?.toLowerCase().includes(q)
    return matchStatus && matchSearch
  })
})

// 2. Delivery / SO Returns (Barang Rusak)
const filteredDeliveryReturns = computed(() => {
  return returnStore.returns.filter((r) => {
    if (isDirectSaleReturn(r)) return false
    const matchStatus = selectedStatus.value === 'ALL' || r.status === selectedStatus.value
    const q = searchQuery.value.toLowerCase()
    const doNo = (r.sales_order?.delivery_batch?.delivery_order_no || r.notes || '').toLowerCase()
    const plate = (r.sales_order?.delivery_batch?.vehicle?.plate_number || '').toLowerCase()
    const matchSearch =
      !q ||
      r.return_no.toLowerCase().includes(q) ||
      r.sales_order?.so_number?.toLowerCase().includes(q) ||
      r.sales_order?.store?.name?.toLowerCase().includes(q) ||
      r.employee?.first_name?.toLowerCase().includes(q) ||
      doNo.includes(q) ||
      plate.includes(q)
    return matchStatus && matchSearch
  })
})

// 3. Failed Delivery / Muatan Returns
const filteredTransfers = computed(() => {
  return transferStore.transfers.filter((t) => {
    if (t.type !== 'RETURN') return false
    const matchStatus = selectedStatus.value === 'ALL' || t.status === selectedStatus.value
    const q = searchQuery.value.toLowerCase()
    const matchSearch =
      !q ||
      t.reference_no?.toLowerCase().includes(q) ||
      t.employee?.first_name?.toLowerCase().includes(q) ||
      t.product?.name?.toLowerCase().includes(q)
    return matchStatus && matchSearch
  })
})

const notaColumns = [
  { key: 'return_no', label: 'No. Retur' },
  { key: 'transaction', label: 'Faktur / Pelanggan' },
  { key: 'created_at', label: 'Tgl Pengajuan' },
  { key: 'employee', label: 'Pelapor (Admin Nota)' },
  { key: 'total_amount', label: 'Total Nilai' },
  { key: 'status', label: 'Status' },
  { key: 'actions', label: 'Aksi' },
]

const deliveryColumns = [
  { key: 'return_no', label: 'No. Retur' },
  { key: 'sales_order', label: 'No. SO / Toko' },
  { key: 'created_at', label: 'Tgl Pengajuan' },
  { key: 'employee', label: 'Pelapor / Armada' },
  { key: 'total_amount', label: 'Total Kredit' },
  { key: 'status', label: 'Status' },
  { key: 'actions', label: 'Aksi' },
]

const transferColumns = [
  { key: 'reference_no', label: 'Referensi' },
  { key: 'transfer_date', label: 'Tgl Tarik' },
  { key: 'employee', label: 'Armada/Driver' },
  { key: 'product', label: 'Produk Muatan' },
  { key: 'quantity', label: 'Kuantitas Sisa' },
  { key: 'status', label: 'Status' },
  { key: 'actions', label: 'Aksi' },
]

const cabangColumns = [
  { key: 'return_no', label: 'No. Retur' },
  { key: 'from_branch', label: 'Cabang & Pabrik' },
  { key: 'created_at', label: 'Tgl Retur' },
  { key: 'items', label: 'Rincian Produk Retur' },
  { key: 'status', label: 'Status' },
  { key: 'actions', label: 'Aksi' },
]

async function fetchReturns(force = false) {
  const tasks: Promise<unknown>[] = [fetchFactoryReturns()]
  if (force || masterStore.branches.length === 0) {
    tasks.push(masterStore.fetchBranches(force))
  }
  if (activeBranchId.value) {
    tasks.push(
      returnStore.fetchByBranch(activeBranchId.value, selectedStatus.value),
      transferStore.fetchTransfers(activeBranchId.value),
    )
  }
  await Promise.allSettled(tasks)
}

async function handleApprove(row: SalesReturn) {
  const { id, status } = row
  if (status === 'PENDING') {
    if (!confirm('Setujui retur ini secara administratif? Tagihan / faktur akan dipotong senilai retur. Setelah ini, fisik barang harus diterima oleh Gudang.')) return
  } else if (status === 'PENDING_WAREHOUSE') {
    const message = row.return_type === 'TUKAR_PRODUK'
      ? 'Terima barang retur dan keluarkan produk pengganti? Faktur sudah disesuaikan saat pengajuan; persetujuan ini akan memindahkan stok.'
      : 'Terima fisik barang retur ini? Stok fisik di Gudang Anda akan bertambah sesuai kondisi.'
    if (!confirm(message)) return
  }

  loading.value = true
  try {
    if (status === 'PENDING') {
      await returnStore.approveReturnNota(id)
      toast.success('Retur secara administratif disetujui. Menunggu penerimaan fisik Gudang.')
    } else {
      await returnStore.approveReturnWarehouse(id)
      toast.success(row.return_type === 'TUKAR_PRODUK' ? 'Pertukaran barang disetujui; stok asal dan pengganti sudah disesuaikan.' : 'Barang fisik retur berhasil diterima ke dalam stok Gudang.')
    }
    await fetchReturns()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyetujui retur')
  } finally {
    loading.value = false
  }
}

async function handleApproveTransfer(id: string) {
  if (!confirm('Setujui pengembalian muatan ini? Stok Gudang UTAMA akan bertambah dan stok truk akan menjadi 0.')) return

  loading.value = true
  try {
    await transferStore.completeTransfer(id, activeBranchId.value)
    toast.success('Pengembalian muatan disetujui')
    await fetchReturns()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyetujui muatan')
  } finally {
    loading.value = false
  }
}

const formatDate = (d: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
    : '-'
const formatCurrency = (v: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(v)

const currentBranchName = computed(() => {
  const branch = masterStore.branches.find((b) => b.id === activeBranchId.value)
  return branch ? branch.name : 'Semua Cabang'
})

// Dialog and Edit State
const showEditModal = ref(false)
const editLoading = ref(false)
const selectedReturn = ref<any>(null)
const editForm = ref({
  id: '',
  notes: '',
  items: [] as Array<{
    id: string
    product_id: string
    product_name: string
    quantity: number
    reason: string
    condition: string
  }>
})

function openEditModal(row: any) {
  selectedReturn.value = row
  editForm.value = {
    id: row.id,
    notes: row.notes || '',
    items: (row.items || []).map((item: any) => ({
      id: item.id,
      product_id: item.product_id,
      product_name: item.product?.name || 'Produk',
      quantity: item.quantity,
      reason: item.reason || 'RUSAK',
      condition: item.condition || 'GOOD'
    }))
  }
  showEditModal.value = true
}

async function handleUpdateReturn() {
  if (editForm.value.items.some(item => item.quantity <= 0)) {
    toast.error('Jumlah retur produk harus lebih dari 0')
    return
  }

  editLoading.value = true
  try {
    const payload = {
      notes: editForm.value.notes,
      items: editForm.value.items.map(item => ({
        product_id: item.product_id,
        quantity: item.quantity,
        reason: item.reason,
        condition: item.condition
      }))
    }
    await returnStore.updateReturn(editForm.value.id, payload)
    toast.success('Retur berhasil diperbarui')
    showEditModal.value = false
    await fetchReturns()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal memperbarui retur')
  } finally {
    editLoading.value = false
  }
}

async function handleDeleteReturn(id: string) {
  if (!confirm('Apakah Anda yakin ingin menghapus pengajuan retur ini?')) return

  loading.value = true
  try {
    await returnStore.deleteReturn(id)
    toast.success('Retur berhasil dihapus')
    await fetchReturns()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menghapus retur')
  } finally {
    loading.value = false
  }
}

const isAnyLoading = computed(
  () =>
    isRefreshing.value ||
    loading.value ||
    returnStore.loading ||
    transferStore.loading ||
    isLoadingFactoryReturns.value,
)

const {
  isRefreshing,
  lastRefreshedTimeStr,
  markRefreshed,
  refresh: triggerRefresh,
} = useDataRefresh(
  async ({ isManual }) => {
    await fetchReturns(isManual)
  },
  {
    successMessage: 'Data Retur berhasil diperbarui',
    canAutoRefresh: () =>
      !showEditModal.value &&
      !showDeliveryReturnModal.value &&
      !showSalesStockReturnModal.value &&
      !showBranchReturnModal.value &&
      !showFactoryReceiveModal.value &&
      !loading.value &&
      !editLoading.value &&
      !isSubmittingFactoryReceive.value,
  },
)

watch(
  () => activeBranchId.value,
  (newId) => {
    if (newId) fetchReturns().then(() => markRefreshed())
  },
)

watch(
  () => masterStore.selectedSalesType,
  () => {
    fetchReturns().then(() => markRefreshed())
  },
)

watch(selectedStatus, () => {
  fetchReturns().then(() => markRefreshed())
})

onMounted(() => {
  fetchReturns().then(() => markRefreshed())
})
</script>

<template>
  <div class="space-y-8 p-6">
    <!-- Header -->
    <div
      class="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4 bg-white text-slate-900 p-8 rounded-2xl shadow-lg border border-slate-100"
    >
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-3">
          <RotateCcw class="w-8 h-8 text-red-600" />
          <span>Manajemen Retur</span>
          <span v-if="currentBranchName" class="text-primary text-2xl font-bold">
            — {{ currentBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Kelola pengembalian barang dari penjualan langsung (Admin Nota), pengiriman armada, dan sisa muatan gudang.
        </p>
      </div>
      <RefreshDataButton
        :loading="isAnyLoading"
        :last-updated="lastRefreshedTimeStr"
        @click="triggerRefresh()"
      />
    </div>

    <!-- Search & Filter Controls -->
    <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
      <!-- Tabs -->
      <div class="flex items-center gap-2 bg-slate-100 p-1.5 rounded-2xl w-full sm:w-fit overflow-x-auto">
        <button
          @click="activeTab = 'nota'"
          class="px-5 py-2.5 rounded-xl font-bold text-sm transition-all whitespace-nowrap flex items-center gap-2"
          :class="activeTab === 'nota' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
        >
          <FileText class="w-4 h-4" :class="activeTab === 'nota' ? 'text-red-600' : 'text-slate-400'" />
          Retur Admin Nota (Penjualan Langsung)
          <span
            v-if="filteredNotaReturns.length > 0"
            class="px-2 py-0.5 rounded-full text-xs font-black"
            :class="activeTab === 'nota' ? 'bg-red-50 text-red-600' : 'bg-slate-200 text-slate-600'"
          >
            {{ filteredNotaReturns.length }}
          </span>
        </button>
        <button
          @click="activeTab = 'rusak'"
          class="px-5 py-2.5 rounded-xl font-bold text-sm transition-all whitespace-nowrap flex items-center gap-2"
          :class="activeTab === 'rusak' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
        >
          <Building2 class="w-4 h-4" :class="activeTab === 'rusak' ? 'text-amber-600' : 'text-slate-400'" />
          Retur Pengiriman Toko (Barang Rusak)
          <span
            v-if="filteredDeliveryReturns.length > 0"
            class="px-2 py-0.5 rounded-full text-xs font-black"
            :class="activeTab === 'rusak' ? 'bg-amber-50 text-amber-600' : 'bg-slate-200 text-slate-600'"
          >
            {{ filteredDeliveryReturns.length }}
          </span>
        </button>
        <button
          @click="activeTab = 'muatan'"
          class="px-5 py-2.5 rounded-xl font-bold text-sm transition-all whitespace-nowrap flex items-center gap-2"
          :class="activeTab === 'muatan' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
        >
          <Truck class="w-4 h-4" :class="activeTab === 'muatan' ? 'text-blue-600' : 'text-slate-400'" />
          Pengembalian Muatan & Bawaan Sales
          <span
            v-if="filteredTransfers.length > 0"
            class="px-2 py-0.5 rounded-full text-xs font-black"
            :class="activeTab === 'muatan' ? 'bg-blue-50 text-blue-600' : 'bg-slate-200 text-slate-600'"
          >
            {{ filteredTransfers.length }}
          </span>
        </button>
        <button
          @click="activeTab = 'cabang'"
          class="px-5 py-2.5 rounded-xl font-bold text-sm transition-all whitespace-nowrap flex items-center gap-2"
          :class="activeTab === 'cabang' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
        >
          <RotateCcw class="w-4 h-4" :class="activeTab === 'cabang' ? 'text-rose-600' : 'text-slate-400'" />
          Retur Cabang ke Pabrik
          <span
            v-if="filteredFactoryReturns.length > 0"
            class="px-2 py-0.5 rounded-full text-xs font-black"
            :class="activeTab === 'cabang' ? 'bg-rose-50 text-rose-600' : 'bg-slate-200 text-slate-600'"
          >
            {{ filteredFactoryReturns.length }}
          </span>
        </button>
      </div>

      <!-- Actions & Search Box -->
      <div class="flex items-center gap-2.5 flex-wrap">
        <Button
          @click="showSalesStockReturnModal = true"
          class="bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl px-4 py-2 font-bold text-xs shadow-md shadow-emerald-600/20 flex items-center gap-1.5"
        >
          <RotateCcw class="w-4 h-4 text-emerald-200" /> + Retur Bawaan Sales
        </Button>
        <Button
          @click="showDeliveryReturnModal = true"
          class="bg-slate-900 hover:bg-slate-800 text-white rounded-xl px-4 py-2 font-bold text-xs shadow-md flex items-center gap-1.5"
        >
          <Truck class="w-4 h-4 text-rose-400" /> + Retur Pengiriman (SJ)
        </Button>
        <Button
          @click="showBranchReturnModal = true"
          class="bg-rose-600 hover:bg-rose-700 text-white rounded-xl px-4 py-2 font-bold text-xs shadow-md shadow-rose-600/20 flex items-center gap-1.5"
        >
          <Plus class="w-4 h-4" /> + Retur Cabang
        </Button>
        <div class="relative w-full sm:w-64">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari no. retur / faktur / SJ..."
            class="w-full pl-10 pr-4 py-2.5 rounded-xl border border-slate-200 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500 transition-all shadow-sm"
          />
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-if="(activeTab === 'nota' && filteredNotaReturns.length === 0 && !returnStore.loading) || (activeTab === 'rusak' && filteredDeliveryReturns.length === 0 && !returnStore.loading) || (activeTab === 'muatan' && filteredTransfers.length === 0 && !transferStore.loading) || (activeTab === 'cabang' && filteredFactoryReturns.length === 0 && !isLoadingFactoryReturns)"
      class="text-center py-24 bg-white border border-slate-200 rounded-[40px] shadow-sm"
    >
      <Box class="w-16 h-16 text-slate-200 mx-auto mb-4" />
      <h3 class="text-lg font-black text-slate-900">Data Kosong</h3>
      <p class="text-slate-500 text-sm mt-1">Belum ada pengajuan retur pada tab ini.</p>
    </div>

    <!-- Tables -->
    <div v-else class="animate-in fade-in slide-in-from-bottom-4 duration-500">
      <Card
        class="border-none shadow-xl shadow-slate-200/50 bg-white rounded-[2.5rem] overflow-hidden border border-slate-100 p-0"
      >
        <!-- Tab 1: Retur Admin Nota (Penjualan Langsung) -->
        <DataTable v-if="activeTab === 'nota'" :columns="notaColumns" :data="filteredNotaReturns">
          <template #cell-return_no="{ row }">
            <div class="flex flex-col">
              <span class="font-black text-slate-900">{{ row.return_no }}</span>
              <span v-if="row.return_type" class="text-[10px] font-bold text-red-600 uppercase">
                {{ row.return_type }}
              </span>
            </div>
          </template>
          <template #cell-transaction="{ row }">
            <div class="flex flex-col">
              <span class="font-bold text-slate-800">{{ row.transaction?.receipt_no || 'Faktur Langsung' }}</span>
              <span class="text-xs text-slate-500 font-medium">
                {{ row.transaction?.buyer_name || row.transaction?.store?.name || 'Pelanggan Langsung' }}
              </span>
            </div>
          </template>
          <template #cell-created_at="{ row }">
            <span class="text-xs font-bold text-slate-500 flex items-center gap-2">
              {{ formatDate(row.created_at) }}
            </span>
          </template>
          <template #cell-employee="{ row }">
            <span class="text-xs font-bold text-slate-700">
              {{ row.employee?.first_name }} {{ row.employee?.last_name || '' }}
            </span>
          </template>
          <template #cell-total_amount="{ row }">
            <div class="flex flex-col">
              <span class="font-black text-red-600">{{ formatCurrency(row.total_amount) }}</span>
              <span v-if="row.return_type === 'TUKAR_PRODUK'" class="text-[10px] font-bold text-amber-700">
                Selisih faktur: {{ row.price_difference > 0 ? '+' : '' }}{{ formatCurrency(row.price_difference || 0) }}
                {{ row.invoice_adjusted_at ? '(faktur diperbarui)' : '(menunggu penyesuaian)' }}
              </span>
              <span v-if="row.return_type === 'TUKAR_PRODUK' && row.invoice_adjusted_at && Number(row.transaction?.paid_amount || 0) > Number(row.transaction?.total_amount || 0)" class="text-[10px] font-bold text-rose-700">
                Kelebihan bayar untuk diselesaikan kasir: {{ formatCurrency(Number(row.transaction?.paid_amount || 0) - Number(row.transaction?.total_amount || 0)) }}
              </span>
              <div v-if="row.items" class="mt-1 space-y-0.5">
                <div
                  v-for="item in row.items"
                  :key="item.id"
                  class="text-[9px] text-slate-500 font-bold uppercase"
                >
                  • {{ item.product?.name || 'Item' }}
                  <span class="text-slate-400">({{ item.quantity }} PCS - {{ item.condition === 'GOOD' ? 'BAGUS' : 'RUSAK' }})</span>
                </div>
              </div>
            </div>
          </template>
          <template #cell-status="{ row }">
            <div
              class="px-3 py-1.5 rounded-xl text-[10px] font-black tracking-widest uppercase w-fit border"
              :class="{
                'bg-emerald-50 text-emerald-600 border-emerald-100': row.status === 'APPROVED',
                'bg-amber-50 text-amber-600 border-amber-100': row.status === 'PENDING',
                'bg-blue-50 text-blue-600 border-blue-100': row.status === 'PENDING_WAREHOUSE',
                'bg-rose-50 text-rose-600 border-rose-100': row.status === 'REJECTED'
              }"
            >
              {{ row.status === 'PENDING_WAREHOUSE' ? 'MENUNGGU GUDANG (ACC)' : row.status }}
            </div>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <!-- Admin Gudang ACC Terima Fisik (Hanya tampil di tab Gudang) -->
              <template v-if="isWarehouse">
                <Button
                  v-if="row.status === 'PENDING_WAREHOUSE' || row.status === 'PENDING'"
                  @click="handleApprove(row)"
                  :disabled="loading"
                  size="sm"
                  class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-700 text-white text-xs px-4 shadow-sm shadow-emerald-600/20"
                >
                  <CheckCircle2 class="w-3.5 h-3.5 mr-1" />
                  Terima Barang / ACC
                </Button>
                <div
                  v-else-if="row.status === 'APPROVED'"
                  class="flex flex-col items-start gap-0.5 text-emerald-600 font-bold text-[10px] bg-emerald-50 px-2.5 py-1.5 rounded-xl border border-emerald-100"
                >
                  <div class="flex items-center gap-1">
                    <CheckCircle2 class="w-3.5 h-3.5" /> <span>SELESAI (ACC)</span>
                  </div>
                  <span v-if="row.approved_by?.first_name" class="text-[9px] text-slate-700 font-semibold mt-0.5">
                    ✓ {{ row.approved_by.first_name }} {{ row.approved_by.last_name || '' }}
                  </span>
                </div>
              </template>

              <!-- Di tab Sales (Monitoring/Read-only untuk aksi gudang) -->
              <template v-else>
                <div
                  v-if="row.status === 'APPROVED'"
                  class="flex flex-col items-start gap-0.5 text-emerald-600 font-bold text-[10px] bg-emerald-50 px-2.5 py-1.5 rounded-xl border border-emerald-100"
                >
                  <div class="flex items-center gap-1">
                    <CheckCircle2 class="w-3.5 h-3.5" /> <span>DISETUJUI</span>
                  </div>
                  <span v-if="row.approved_by?.first_name" class="text-[9px] text-slate-700 font-semibold mt-0.5">
                    ✓ {{ row.approved_by.first_name }} {{ row.approved_by.last_name || '' }}
                  </span>
                </div>
                <div
                  v-else
                  class="px-2.5 py-1.5 rounded-xl text-[10px] font-bold text-amber-600 bg-amber-50 border border-amber-100"
                >
                  Menunggu ACC Gudang
                </div>
              </template>
              
              <!-- Edit & Delete Buttons (bisa diakses selagi belum APPROVED) -->
              <button
                v-if="row.status !== 'APPROVED' && row.return_type !== 'TUKAR_PRODUK'"
                @click="openEditModal(row)"
                class="p-1.5 rounded-lg text-slate-400 hover:text-blue-600 hover:bg-blue-50 transition-all"
                title="Edit"
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>
              <button
                v-if="row.status !== 'APPROVED'"
                @click="handleDeleteReturn(row.id)"
                :disabled="loading"
                class="p-1.5 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 transition-all"
                title="Hapus"
              >
                <Trash2 class="w-3.5 h-3.5" />
              </button>
            </div>
          </template>
        </DataTable>

        <!-- Tab 2: Retur Pengiriman Toko (Barang Rusak & Tukar Produk) -->
        <DataTable v-if="activeTab === 'rusak'" :columns="deliveryColumns" :data="filteredDeliveryReturns">
          <template #cell-return_no="{ row }">
            <div class="flex flex-col">
              <span class="font-black text-slate-900">{{ row.return_no }}</span>
              <span v-if="row.return_type" class="text-[10px] font-bold text-red-600 uppercase">
                {{ row.return_type.replace('_', ' ') }}
              </span>
            </div>
          </template>
          <template #cell-sales_order="{ row }">
            <div class="flex flex-col">
              <div class="flex items-center gap-1.5 flex-wrap">
                <span class="font-bold text-slate-800">{{ row.sales_order?.so_number || '-' }}</span>
                <span
                  v-if="row.sales_order?.delivery_batch?.delivery_order_no"
                  class="text-[9px] font-black uppercase text-blue-700 bg-blue-50 px-1.5 py-0.5 rounded border border-blue-200"
                >
                  {{ row.sales_order.delivery_batch.delivery_order_no }}
                </span>
                <span
                  v-else-if="row.notes && row.notes.includes('[DELIVERY_RETURN:')"
                  class="text-[9px] font-black uppercase text-blue-700 bg-blue-50 px-1.5 py-0.5 rounded border border-blue-200"
                >
                  {{ row.notes.match(/\[DELIVERY_RETURN:\s*([^\]]+)\]/)?.[1] || 'SJ' }}
                </span>
              </div>
              <span class="text-xs text-slate-500 font-medium">{{ row.sales_order?.store?.name || '-' }}</span>
            </div>
          </template>
          <template #cell-created_at="{ row }">
            <span class="text-xs font-bold text-slate-500 flex items-center gap-2">
              {{ formatDate(row.created_at) }}
            </span>
          </template>
          <template #cell-employee="{ row }">
            <span class="text-xs font-bold text-slate-700">{{ row.employee?.first_name }} {{ row.employee?.last_name || '' }}</span>
          </template>
          <template #cell-total_amount="{ row }">
            <div class="flex flex-col">
              <span class="font-black text-red-600">{{ formatCurrency(row.total_amount) }}</span>
              <span v-if="row.return_type === 'TUKAR_PRODUK'" class="text-[10px] font-bold text-amber-700">
                Selisih faktur: {{ row.price_difference > 0 ? '+' : '' }}{{ formatCurrency(row.price_difference || 0) }}
                {{ row.invoice_adjusted_at ? '(faktur diperbarui)' : '(menunggu penyesuaian)' }}
              </span>
              <span v-if="row.return_type === 'TUKAR_PRODUK' && row.invoice_adjusted_at && Number(row.transaction?.paid_amount || 0) > Number(row.transaction?.total_amount || 0)" class="text-[10px] font-bold text-rose-700">
                Kelebihan bayar untuk diselesaikan kasir: {{ formatCurrency(Number(row.transaction?.paid_amount || 0) - Number(row.transaction?.total_amount || 0)) }}
              </span>
              <div v-if="row.items" class="mt-1 space-y-1">
                <div
                  v-for="item in row.items"
                  :key="item.id"
                  class="text-[9px] text-slate-700 font-bold uppercase flex flex-col bg-slate-50 p-1.5 rounded-lg border border-slate-200/60"
                >
                  <div class="flex items-center justify-between">
                    <span>• {{ item.product?.name || 'Item' }} ({{ item.quantity }} PCS)</span>
                    <span :class="item.condition === 'GOOD' ? 'text-emerald-700 bg-emerald-100/70 border-emerald-200' : 'text-red-700 bg-red-100/70 border-red-200'" class="px-1.5 py-0.5 rounded text-[8px] font-black border">
                      {{ item.condition === 'GOOD' ? '🟢 BAGUS' : '🔴 RUSAK' }}
                    </span>
                  </div>
                  <div v-if="item.replacement_product" class="text-[8px] text-violet-700 font-black mt-0.5">
                    ↳ Diganti: {{ item.replacement_product.name }} ({{ item.replacement_quantity }} PCS)
                  </div>
                </div>
              </div>
            </div>
          </template>
          <template #cell-status="{ row }">
            <div
              class="px-3 py-1.5 rounded-xl text-[10px] font-black tracking-widest uppercase w-fit border"
              :class="{
                'bg-emerald-50 text-emerald-600 border-emerald-100': row.status === 'APPROVED',
                'bg-amber-50 text-amber-600 border-amber-100': row.status === 'PENDING',
                'bg-blue-50 text-blue-600 border-blue-100': row.status === 'PENDING_WAREHOUSE',
                'bg-rose-50 text-rose-600 border-rose-100': row.status === 'REJECTED'
              }"
            >
              {{ row.status === 'PENDING_WAREHOUSE' ? 'MENUNGGU GUDANG (ACC)' : row.status }}
            </div>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <template v-if="isWarehouse">
                <Button
                  v-if="row.status === 'PENDING_WAREHOUSE' || row.status === 'PENDING'"
                  @click="handleApprove(row)"
                  :disabled="loading"
                  size="sm"
                  class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-700 text-white text-xs px-4 shadow-sm shadow-emerald-600/20"
                >
                  <CheckCircle2 class="w-3.5 h-3.5 mr-1" />
                  Terima Barang / ACC
                </Button>
                <div
                  v-else-if="row.status === 'APPROVED'"
                  class="flex flex-col items-start gap-0.5 text-emerald-600 font-bold text-[10px] bg-emerald-50 px-2.5 py-1.5 rounded-xl border border-emerald-100"
                >
                  <div class="flex items-center gap-1">
                    <CheckCircle2 class="w-3.5 h-3.5" /> <span>DITERIMA (ACC)</span>
                  </div>
                  <span v-if="row.approved_by?.first_name" class="text-[9px] text-slate-700 font-semibold mt-0.5">
                    ✓ {{ row.approved_by.first_name }} {{ row.approved_by.last_name || '' }}
                  </span>
                </div>
              </template>
              <template v-else>
                <Button
                  v-if="row.status === 'PENDING'"
                  @click="handleApprove(row)"
                  :disabled="loading"
                  size="sm"
                  class="rounded-xl font-bold bg-slate-900 hover:bg-amber-600 text-white text-xs px-4"
                >
                  Potong Nota
                </Button>
                <div
                  v-else-if="row.status === 'APPROVED'"
                  class="flex flex-col items-start gap-0.5 text-emerald-600 font-bold text-[10px] bg-emerald-50 px-2.5 py-1.5 rounded-xl border border-emerald-100"
                >
                  <div class="flex items-center gap-1">
                    <CheckCircle2 class="w-3.5 h-3.5" /> <span>SELESAI</span>
                  </div>
                  <span v-if="row.approved_by?.first_name" class="text-[9px] text-slate-700 font-semibold mt-0.5">
                    ✓ {{ row.approved_by.first_name }} {{ row.approved_by.last_name || '' }}
                  </span>
                </div>
                <div v-else-if="row.status === 'PENDING_WAREHOUSE'" class="text-[10px] font-bold text-blue-600 bg-blue-50 px-2.5 py-1 rounded-lg border border-blue-200">
                  Menunggu ACC Gudang
                </div>
              </template>

              <!-- Edit & Delete Buttons untuk Retur Pengiriman Toko (selagi belum APPROVED) -->
              <button
                v-if="row.status !== 'APPROVED' && row.return_type !== 'TUKAR_PRODUK'"
                @click="openEditModal(row)"
                class="p-1.5 rounded-lg text-slate-400 hover:text-blue-600 hover:bg-blue-50 transition-all"
                title="Edit Retur"
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>
              <button
                v-if="row.status !== 'APPROVED'"
                @click="handleDeleteReturn(row.id)"
                :disabled="loading"
                class="p-1.5 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 transition-all"
                title="Hapus Retur"
              >
                <Trash2 class="w-3.5 h-3.5" />
              </button>
            </div>
          </template>
        </DataTable>

        <!-- Tab 3: Pengembalian Muatan -->
        <DataTable v-if="activeTab === 'muatan'" :columns="transferColumns" :data="filteredTransfers">
          <template #cell-reference_no="{ row }">
            <span class="font-black text-slate-900">{{ row.reference_no }}</span>
          </template>
          <template #cell-transfer_date="{ row }">
            <span class="text-xs font-bold text-slate-500 flex items-center gap-2">
              {{ formatDate(row.transfer_date) }}
            </span>
          </template>
          <template #cell-employee="{ row }">
            <span class="text-xs font-bold text-slate-500">{{ row.employee?.first_name }} {{ row.employee?.last_name }}</span>
          </template>
          <template #cell-product="{ row }">
            <span class="text-xs font-bold text-slate-700">{{ row.product?.name }}</span>
          </template>
          <template #cell-quantity="{ row }">
            <div class="flex flex-col">
              <span class="font-black text-slate-900">{{ row.quantity }} {{ row.product?.unit || 'pcs' }}</span>
              <div v-if="row.actual_quantity > 0 || row.damaged_quantity > 0" class="flex items-center gap-2 mt-0.5 text-[10px]">
                <span v-if="row.actual_quantity > 0" class="text-emerald-700 font-bold bg-emerald-50 px-1.5 py-0.2 rounded border border-emerald-200">
                  🟢 {{ row.actual_quantity }} Bagus
                </span>
                <span v-if="row.damaged_quantity > 0" class="text-rose-700 font-bold bg-rose-50 px-1.5 py-0.2 rounded border border-rose-200">
                  🔴 {{ row.damaged_quantity }} Rusak
                </span>
              </div>
            </div>
          </template>
          <template #cell-status="{ row }">
            <div
              class="px-3 py-1.5 rounded-xl text-[10px] font-black tracking-widest uppercase w-fit border"
              :class="
                row.status === 'COMPLETED'
                  ? 'bg-emerald-50 text-emerald-600 border-emerald-100'
                  : 'bg-amber-50 text-amber-600 border-amber-100'
              "
            >
              {{ row.status === 'WAITING_SUPERVISOR' || row.status === 'PENDING' ? 'MENUNGGU' : row.status }}
            </div>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <Button
                v-if="(row.status === 'PENDING' || row.status === 'WAITING_SUPERVISOR') && isWarehouse"
                @click="handleApproveTransfer(row.id)"
                :disabled="loading"
                size="sm"
                class="rounded-xl font-bold bg-slate-900 hover:bg-emerald-600 text-white text-xs px-4"
              >
                Terima Muatan
              </Button>
              <div
                v-else-if="row.status === 'COMPLETED'"
                class="flex items-center gap-1.5 text-emerald-600 font-black text-[10px] bg-emerald-50 px-3 py-2 rounded-xl border border-emerald-100"
              >
                <CheckCircle2 class="w-3.5 h-3.5" /> DITERIMA
              </div>
              <div
                v-else
                class="text-[10px] font-bold text-slate-400 bg-slate-50 px-2.5 py-1.5 rounded-xl border border-slate-100"
              >
                Menunggu Gudang
              </div>
            </div>
          </template>
        </DataTable>

        <!-- Tab 4: Retur Cabang ke Pabrik -->
        <DataTable v-if="activeTab === 'cabang'" :columns="cabangColumns" :data="filteredFactoryReturns">
          <template #cell-return_no="{ row }">
            <div class="font-extrabold text-slate-900 text-sm flex items-center gap-1.5">
              <RotateCcw class="w-3.5 h-3.5 text-rose-600" />
              {{ row.return_no }}
            </div>
            <div class="text-[10px] text-slate-400 mt-0.5">
              {{ row.notes || 'Retur Stok Cabang' }}
            </div>
          </template>

          <template #cell-from_branch="{ row }">
            <div class="font-bold text-slate-900">{{ row.from_branch?.name || 'Cabang' }}</div>
            <div class="text-[11px] text-slate-500">Tujuan: {{ row.to_factory?.name || 'Pabrik' }}</div>
          </template>

          <template #cell-created_at="{ row }">
            <span class="text-xs font-bold text-slate-600">
              {{ formatDate(row.shipped_at || row.created_at) }}
            </span>
          </template>

          <template #cell-items="{ row }">
            <div class="space-y-1 max-w-xs">
              <div
                v-for="(item, idx) in row.items"
                :key="idx"
                class="text-[11px] font-bold text-slate-800 flex items-center justify-between gap-2"
              >
                <span>• {{ item.product?.name || 'Produk' }}</span>
                <span class="text-rose-600 font-mono text-[10px]">
                  {{ item.quantity }} {{ item.unit || 'PCS' }} ({{ item.return_type === 'GOOD' ? 'BAIK' : 'RUSAK' }})
                </span>
              </div>
            </div>
          </template>

          <template #cell-status="{ row }">
            <div
              class="px-3 py-1.5 rounded-xl text-[10px] font-black tracking-widest uppercase w-fit border"
              :class="{
                'bg-emerald-50 text-emerald-600 border-emerald-100': row.status === 'RECEIVED' || row.is_received,
                'bg-amber-50 text-amber-600 border-amber-100': row.status === 'SHIPPED' && !row.is_received,
                'bg-blue-50 text-blue-600 border-blue-100': row.status === 'REQUESTED',
              }"
            >
              {{ row.is_received || row.status === 'RECEIVED' ? 'DITERIMA (POTONG PIUTANG)' : row.status }}
            </div>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <Button
                v-if="!row.is_received && row.status !== 'RECEIVED'"
                @click="openFactoryReceiveModal(row)"
                size="sm"
                class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-700 text-white text-xs px-3 shadow-sm shadow-emerald-600/20"
              >
                <CheckCircle2 class="w-3.5 h-3.5 mr-1" />
                Terima & Potong Piutang
              </Button>
              <div
                v-else
                class="flex items-center gap-1 text-emerald-600 font-bold text-[10px] bg-emerald-50 px-2.5 py-1.5 rounded-xl border border-emerald-100"
              >
                <CheckCircle2 class="w-3.5 h-3.5" /> <span>SELESAI (ACC)</span>
              </div>
            </div>
          </template>
        </DataTable>
      </Card>
    </div>

    <!-- Edit Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="showEditModal"
          class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
          @click.self="showEditModal = false"
        >
          <div class="bg-white rounded-3xl shadow-2xl w-full max-w-xl max-h-[90vh] flex flex-col overflow-hidden">
            <!-- Modal Header -->
            <div class="flex items-center justify-between px-8 py-6 border-b border-slate-100">
              <div>
                <h2 class="text-xl font-extrabold text-slate-900 flex items-center gap-2">
                  <Pencil class="w-5 h-5 text-blue-600" />
                  Edit Retur
                </h2>
                <div class="flex items-center gap-2 mt-1">
                  <span class="text-xs font-bold text-slate-700 bg-slate-100 px-2 py-0.5 rounded-md">{{ selectedReturn?.return_no }}</span>
                  <span v-if="selectedReturn?.transaction?.receipt_no" class="text-xs text-slate-500 font-medium">
                    Faktur: <span class="font-bold text-slate-700">{{ selectedReturn.transaction.receipt_no }}</span>
                  </span>
                  <span v-else-if="selectedReturn?.sales_order?.so_number" class="text-xs text-slate-500 font-medium">
                    SO: <span class="font-bold text-slate-700">{{ selectedReturn.sales_order.so_number }}</span>
                  </span>
                  <span v-if="selectedReturn?.return_type" class="text-[10px] font-black uppercase text-red-600 bg-red-50 px-2 py-0.5 rounded-full border border-red-100">
                    {{ selectedReturn.return_type.replace('_', ' ') }}
                  </span>
                </div>
              </div>
              <button
                @click="showEditModal = false"
                class="p-2 rounded-xl hover:bg-slate-100 text-slate-400 hover:text-slate-600 transition-all"
              >
                <X class="w-5 h-5" />
              </button>
            </div>

            <!-- Modal Body -->
            <div class="overflow-y-auto flex-1 px-8 py-6 space-y-5">
              <!-- Catatan -->
              <div class="space-y-1.5">
                <label class="text-sm font-bold text-slate-700">Catatan Retur</label>
                <textarea
                  v-model="editForm.notes"
                  rows="2"
                  placeholder="Tambahkan catatan penyesuaian oleh Gudang..."
                  class="w-full px-4 py-2.5 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                />
              </div>

              <!-- Items -->
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <label class="text-sm font-bold text-slate-700">Daftar Produk Retur</label>
                  <span class="text-xs text-slate-400 font-medium">{{ editForm.items.length }} produk terdaftar</span>
                </div>
                
                <div
                  v-for="(item, idx) in editForm.items"
                  :key="item.product_id"
                  class="bg-slate-50 border border-slate-200 rounded-2xl p-4 space-y-3 relative group"
                >
                  <div class="flex items-center justify-between">
                    <div>
                      <p class="text-sm font-extrabold text-slate-800">{{ item.product_name }}</p>
                      <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">Item #{{ idx + 1 }}</span>
                    </div>
                    <button
                      v-if="editForm.items.length > 1"
                      @click="editForm.items.splice(idx, 1)"
                      type="button"
                      class="text-slate-400 hover:text-red-600 p-1 rounded-lg hover:bg-red-50 transition-all text-xs flex items-center gap-1 font-semibold"
                      title="Hapus item ini dari retur"
                    >
                      <Trash2 class="w-3.5 h-3.5" /> Hapus
                    </button>
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <!-- Jumlah -->
                    <div class="space-y-1">
                      <label class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Jumlah (PCS)</label>
                      <div class="flex items-center gap-2">
                        <button
                          type="button"
                          @click="item.quantity = Math.max(1, item.quantity - 1)"
                          class="w-7 h-7 flex items-center justify-center rounded-lg bg-slate-200 hover:bg-slate-300 text-slate-600 transition"
                        >
                          <Minus class="w-3.5 h-3.5" />
                        </button>
                        <input
                          type="number"
                          v-model.number="item.quantity"
                          min="1"
                          class="w-16 text-center border border-slate-200 rounded-lg px-2 py-1 text-sm font-bold focus:outline-none focus:ring-2 focus:ring-blue-400"
                        />
                        <button
                          type="button"
                          @click="item.quantity++"
                          class="w-7 h-7 flex items-center justify-center rounded-lg bg-slate-200 hover:bg-slate-300 text-slate-600 transition"
                        >
                          <Plus class="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>
                    <!-- Kondisi -->
                    <div class="space-y-1">
                      <label class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Kondisi</label>
                      <select
                        v-model="item.condition"
                        class="w-full border border-slate-200 rounded-lg px-2.5 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400 bg-white"
                      >
                        <option value="GOOD">Baik (Masuk Stok Baik)</option>
                        <option value="DAMAGED">Rusak (Masuk Stok Rusak)</option>
                      </select>
                    </div>
                    <!-- Alasan (full width) -->
                    <div class="col-span-2 space-y-1">
                      <label class="text-[11px] font-bold text-slate-500 uppercase tracking-wider">Alasan Retur</label>
                      <select
                        v-model="item.reason"
                        class="w-full border border-slate-200 rounded-lg px-2.5 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400 bg-white"
                      >
                        <option value="RUSAK">Rusak</option>
                        <option value="KADALUARSA">Kadaluarsa</option>
                        <option value="TOLAK_CUSTOMER">Ditolak Customer</option>
                        <option value="SALAH_KIRIM">Salah Kirim</option>
                        <option value="TOKO_TUTUP">Toko Tutup</option>
                        <option value="LAINNYA">Lainnya</option>
                      </select>
                    </div>
                  </div>
                </div>

                <div v-if="editForm.items.length === 0" class="text-center py-6 text-slate-400 text-sm">
                  Tidak ada item untuk diedit.
                </div>
              </div>
            </div>

            <!-- Modal Footer -->
            <div class="px-8 py-5 border-t border-slate-100 flex items-center justify-end gap-3 bg-slate-50">
              <Button
                variant="outline"
                class="rounded-xl"
                @click="showEditModal = false"
              >Batal</Button>
              <Button
                @click="handleUpdateReturn"
                :disabled="editLoading"
                class="rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-bold px-6"
              >
                <RefreshCw v-if="editLoading" class="w-4 h-4 mr-2 animate-spin" />
                Simpan Perubahan
              </Button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <BranchFactoryReturnModal
      :show="showBranchReturnModal"
      @close="showBranchReturnModal = false"
      @created="fetchFactoryReturns"
    />

    <DeliveryReturnModal
      :show="showDeliveryReturnModal"
      :branch-id="activeBranchId"
      @close="showDeliveryReturnModal = false"
      @created="fetchReturns"
    />

    <SalesStockReturnModal
      v-model:open="showSalesStockReturnModal"
      :branch-id="activeBranchId"
      @success="fetchReturns"
    />

    <!-- MODAL PENYESUAIAN & PENERIMAAN RETUR CABANG KE PABRIK -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="showFactoryReceiveModal"
          class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
          @click.self="showFactoryReceiveModal = false"
        >
          <div class="bg-white rounded-3xl shadow-2xl w-full max-w-2xl max-h-[92vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
            <!-- Modal Header -->
            <div class="flex items-center justify-between px-8 py-6 border-b border-slate-100 bg-slate-50/50">
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <RotateCcw class="w-5 h-5 text-rose-600" />
                  <h2 class="text-xl font-extrabold text-slate-900">
                    Penerimaan & Pemeriksaan Fisik Retur
                  </h2>
                </div>
                <div class="flex flex-wrap items-center gap-2 text-xs">
                  <span class="font-black text-rose-600 bg-rose-50 px-2.5 py-0.5 rounded-lg border border-rose-100 font-mono">
                    {{ selectedFactoryReturn?.return_no }}
                  </span>
                  <span class="text-slate-400">•</span>
                  <span class="text-slate-600 font-bold">
                    {{ selectedFactoryReturn?.from_branch?.name || 'Cabang' }}
                  </span>
                  <span class="text-slate-400">➔</span>
                  <span class="text-slate-600 font-bold">
                    {{ selectedFactoryReturn?.to_factory?.name || 'Pabrik' }}
                  </span>
                </div>
              </div>
              <button
                @click="showFactoryReceiveModal = false"
                class="p-2 rounded-xl hover:bg-slate-200/60 text-slate-400 hover:text-slate-600 transition-all"
              >
                <X class="w-5 h-5" />
              </button>
            </div>

            <!-- Modal Body -->
            <div class="overflow-y-auto flex-1 px-8 py-6 space-y-6">
              <!-- Info Box -->
              <div class="bg-amber-50 border border-amber-200/70 p-4 rounded-2xl flex items-start gap-3">
                <Info class="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
                <div class="text-xs text-amber-900 leading-relaxed">
                  <span class="font-bold">Verifikasi Fisik Retur:</span> Sesuaikan kuantitas barang aktual yang sampai di pabrik. Stok pabrik akan bertambah dan potongan piutang cabang akan dihitung otomatis sesuai dengan rincian fisik yang diterima.
                </div>
              </div>

              <!-- Item List -->
              <div class="space-y-4">
                <div class="flex justify-between items-center">
                  <label class="text-xs font-black uppercase tracking-wider text-slate-500">
                    Daftar Produk Retur ({{ factoryReceiveForm.items.length }} Item)
                  </label>
                  <span class="text-[11px] font-bold text-slate-400">Kuantitas dalam satuan unit / PCS</span>
                </div>

                <div
                  v-for="(item, idx) in factoryReceiveForm.items"
                  :key="item.item_id"
                  class="bg-slate-50 border border-slate-200 rounded-2xl p-5 space-y-4"
                >
                  <!-- Header Item -->
                  <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-2 pb-3 border-b border-slate-200">
                    <div>
                      <h4 class="text-sm font-extrabold text-slate-900">{{ item.product_name }}</h4>
                      <div class="flex items-center gap-2 mt-0.5 text-[10px]">
                        <span class="font-bold text-slate-500">Batch: {{ item.batch_no }}</span>
                        <span class="text-slate-300">•</span>
                        <span
                          class="font-black px-2 py-0.5 rounded uppercase"
                          :class="item.return_type === 'GOOD' ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'"
                        >
                          Klaim Cabang: {{ item.return_type === 'GOOD' ? 'Bagus' : 'Rusak' }}
                        </span>
                        <span v-if="item.damage_reason" class="text-slate-500 italic bg-white px-2 py-0.5 rounded border border-slate-200">
                          "{{ item.damage_reason }}"
                        </span>
                      </div>
                    </div>
                    <div class="bg-white px-3.5 py-1.5 rounded-xl border border-slate-200 text-right">
                      <span class="text-[9px] font-bold uppercase text-slate-400 block">Kirim Cabang</span>
                      <span class="text-xs font-black text-slate-900">{{ item.sent_qty }} {{ item.unit }}</span>
                    </div>
                  </div>

                  <!-- Allocation Inputs Grid -->
                  <div>
                    <p class="text-[10px] font-black uppercase tracking-wider text-slate-500 mb-2">
                      Alokasi Fisik yang Diterima Pabrik:
                    </p>
                    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
                      <!-- 1. Good Qty -->
                      <div class="bg-white p-3 rounded-xl border border-emerald-200 space-y-1">
                        <label class="text-[10px] font-black uppercase tracking-wider text-emerald-800 flex items-center gap-1">
                          <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" /> Qty Bagus (Restock)
                        </label>
                        <input
                          type="number"
                          min="0"
                          :max="item.sent_qty"
                          v-model.number="item.received_good_qty"
                          class="w-full px-3 py-1.5 border border-emerald-300 rounded-lg text-sm font-black text-emerald-900 focus:outline-none focus:ring-2 focus:ring-emerald-400"
                        />
                        <p class="text-[8px] text-emerald-600 font-semibold">Masuk stok barang jadi</p>
                      </div>

                      <!-- 2. Repack Qty -->
                      <div class="bg-white p-3 rounded-xl border border-amber-200 space-y-1">
                        <label class="text-[10px] font-black uppercase tracking-wider text-amber-800 flex items-center gap-1">
                          <Package class="w-3.5 h-3.5 text-amber-600" /> Rusak Kemasan (Repack)
                        </label>
                        <input
                          type="number"
                          min="0"
                          :max="item.sent_qty"
                          v-model.number="item.received_bottle_damaged_qty"
                          class="w-full px-3 py-1.5 border border-amber-300 rounded-lg text-sm font-black text-amber-900 focus:outline-none focus:ring-2 focus:ring-amber-400"
                        />
                        <p class="text-[8px] text-amber-600 font-semibold">Masuk area rework pabrik</p>
                      </div>

                      <!-- 3. Scrap Qty -->
                      <div class="bg-white p-3 rounded-xl border border-rose-200 space-y-1">
                        <label class="text-[10px] font-black uppercase tracking-wider text-rose-800 flex items-center gap-1">
                          <Trash2 class="w-3.5 h-3.5 text-rose-600" /> Rusak Isi (Bocor/Scrap)
                        </label>
                        <input
                          type="number"
                          min="0"
                          :max="item.sent_qty"
                          v-model.number="item.received_content_damaged_qty"
                          class="w-full px-3 py-1.5 border border-rose-300 rounded-lg text-sm font-black text-rose-900 focus:outline-none focus:ring-2 focus:ring-rose-400"
                        />
                        <p class="text-[8px] text-rose-600 font-semibold">Pemusnahan limbah</p>
                      </div>
                    </div>
                  </div>

                  <!-- Catatan Item & Subtotal -->
                  <div class="space-y-1.5">
                    <label class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">
                      Catatan Fisik / Temuan QC Item
                    </label>
                    <input
                      type="text"
                      v-model="item.inspection_notes"
                      placeholder="Contoh: Kemasan bocor 7 PCS, saus merembes di kardus..."
                      class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-xs font-medium focus:outline-none focus:ring-2 focus:ring-blue-400"
                    />
                  </div>

                  <!-- Summary Item -->
                  <div class="flex justify-between items-center bg-white px-4 py-2 rounded-xl border border-slate-200 text-xs font-bold">
                    <span class="text-slate-600">Total Fisik Diterima Produk Ini:</span>
                    <div class="flex items-center gap-2">
                      <span class="text-slate-900 font-black">
                        {{ (item.received_good_qty || 0) + (item.received_bottle_damaged_qty || 0) + (item.received_content_damaged_qty || 0) }} {{ item.unit }}
                      </span>
                      <span
                        v-if="((item.received_good_qty || 0) + (item.received_bottle_damaged_qty || 0) + (item.received_content_damaged_qty || 0)) !== item.sent_qty"
                        class="text-[10px] font-black text-rose-700 bg-rose-100 px-2 py-0.5 rounded-md"
                      >
                        Selisih: {{ ((item.received_good_qty || 0) + (item.received_bottle_damaged_qty || 0) + (item.received_content_damaged_qty || 0)) - item.sent_qty }} {{ item.unit }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- General Checker Notes -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 bg-slate-50 p-4 rounded-2xl border border-slate-200">
                <div class="space-y-1">
                  <label class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Checker / Penerima</label>
                  <input
                    type="text"
                    v-model="factoryReceiveForm.receiver_name"
                    placeholder="Nama staf penerimaan gudang pabrik..."
                    class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-blue-400"
                  />
                </div>
                <div class="space-y-1">
                  <label class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Pemeriksa QC / Supervisor</label>
                  <input
                    type="text"
                    v-model="factoryReceiveForm.inspected_by_name"
                    placeholder="Nama petugas QC / SPV..."
                    class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-blue-400"
                  />
                </div>
                <div class="sm:col-span-2 space-y-1">
                  <label class="text-[10px] font-bold uppercase tracking-wider text-slate-500">Catatan Berita Acara Penerimaan</label>
                  <textarea
                    v-model="factoryReceiveForm.inspection_notes"
                    rows="2"
                    placeholder="Catatan penyesuaian fisik retur umum..."
                    class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-xs focus:outline-none focus:ring-2 focus:ring-blue-400 resize-none"
                  />
                </div>
              </div>
            </div>

            <!-- Modal Footer -->
            <div class="px-8 py-5 border-t border-slate-100 flex items-center justify-end gap-3 bg-slate-50">
              <Button
                variant="outline"
                class="rounded-xl font-bold"
                @click="showFactoryReceiveModal = false"
              >
                Batal
              </Button>
              <Button
                @click="submitFactoryReceive"
                :disabled="isSubmittingFactoryReceive"
                class="rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-extrabold px-6 shadow-md shadow-emerald-600/20"
              >
                <RefreshCw v-if="isSubmittingFactoryReceive" class="w-4 h-4 mr-2 animate-spin" />
                <CheckCircle2 v-else class="w-4 h-4 mr-2" />
                Konfirmasi & Potong Piutang
              </Button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
