<template>
  <div class="p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-6 rounded-3xl shadow-sm border border-slate-100"
    >
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-2">
          <ClipboardList class="w-7 h-7 text-primary" />
          Purchase Request Bahan Baku
        </h1>
        <p class="text-sm text-slate-500 mt-0.5">
          Formulir permohonan pengadaan bahan baku pabrik oleh Admin Gudang untuk direview dan diproses oleh bagian Purchasing.
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <select
          v-model="selectedFactoryId"
          @change="fetchPRs"
          class="text-sm bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">{{ f.name }}</option>
        </select>
        <Button variant="outline" class="rounded-xl border-slate-200" @click="fetchPRs">
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': factoryStore.loading }" />
          Segarkan
        </Button>
        <Button
          @click="openCreatePR"
          :disabled="!selectedFactoryId"
          class="rounded-xl font-bold shadow-md shadow-primary/20 bg-primary hover:bg-primary/90 text-white"
        >
          <Plus class="mr-2 h-4 w-4" /> Buat Purchase Request
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
            v-model="prStatusFilter"
            @change="fetchPRs"
            class="text-xs bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-bold focus:outline-none focus:ring-2 focus:ring-primary/20"
          >
            <option value="">Semua Status</option>
            <option value="PENDING">Menunggu Review Purchasing</option>
            <option value="PROCESSED">Sudah Dibuat PO</option>
            <option value="REJECTED">Ditolak</option>
          </select>
        </div>
        <span class="text-xs text-slate-400 font-medium">
          Total {{ factoryStore.rawMaterialPRs.length }} request pengadaan
        </span>
      </div>

      <DataTable
        :columns="prColumns"
        :data="factoryStore.rawMaterialPRs"
        :isLoading="factoryStore.loading"
      >
        <template #cell-pr_number="{ row }">
          <div class="flex flex-col">
            <span class="font-black text-slate-900 text-sm font-mono">{{ (row as any).pr_number }}</span>
            <span class="text-xs text-slate-400 font-medium">Oleh: {{ (row as any).created_by_name || 'Admin Gudang' }}</span>
          </div>
        </template>

        <template #cell-request_date="{ row }">
          <span class="text-xs font-medium text-slate-600">
            {{ formatDateOnly((row as any).request_date) }}
          </span>
        </template>

        <template #cell-required_date="{ row }">
          <span class="text-xs font-medium text-slate-600">
            {{ (row as any).required_date ? formatDateOnly((row as any).required_date) : '-' }}
          </span>
        </template>

        <template #cell-items_summary="{ row }">
          <div class="text-xs text-slate-600">
            <span class="font-bold text-slate-800">{{ (row as any).items?.length || 0 }} Jenis Bahan</span>
            <div class="text-[11px] text-slate-400 truncate max-w-[240px]">
              {{ ((row as any).items || []).map((it: any) => `${it.product?.name} (${it.requested_qty} ${it.unit})`).join(', ') }}
            </div>
          </div>
        </template>

        <template #cell-status="{ row }">
          <div class="flex flex-col gap-0.5 items-start">
            <Badge
              :variant="getStatusBadgeVariant((row as any).status)"
              class="text-[10px] py-1 px-2.5 font-bold uppercase rounded-lg"
            >
              {{ formatStatus((row as any).status) }}
            </Badge>
            <div v-if="(row as any).purchase_order?.po_number" class="text-[10px] text-emerald-700 font-mono font-semibold mt-0.5">
              PO: {{ (row as any).purchase_order.po_number }}
            </div>
            <div v-if="(row as any).status === 'REJECTED' && (row as any).rejection_reason" class="text-[10px] text-red-500 truncate max-w-[180px]">
              "{{ (row as any).rejection_reason }}"
            </div>
          </div>
        </template>

        <template #cell-actions="{ row }">
          <div class="flex items-center gap-1.5 justify-end">
            <Button
              size="sm"
              variant="outline"
              class="rounded-xl text-xs font-bold border-slate-200"
              @click="openPRDetail(row)"
            >
              <Eye class="w-3.5 h-3.5 mr-1" /> Detail
            </Button>
            <Button
              v-if="(row as any).status === 'PENDING' || (row as any).status === 'REJECTED'"
              size="sm" variant="outline"
              class="rounded-xl text-xs font-bold border-blue-200 text-blue-700"
              title="Edit PR"
              @click="openEditPR(row)"
            >Edit</Button>
            <Button
              v-if="(row as any).status !== 'PROCESSED'"
              size="sm" variant="outline"
              class="rounded-xl text-xs font-bold border-red-200 text-red-600"
              title="Hapus PR"
              @click="deletePR(row)"
            >Hapus</Button>
            <Button
              v-if="canReview && (row as any).status === 'PENDING'"
              size="sm"
              class="rounded-xl text-xs font-bold bg-primary hover:bg-primary/90 text-white shadow-xs"
              @click="openConvertPO(row)"
            >
              <ShoppingCart class="w-3.5 h-3.5 mr-1" /> Jadi PO
            </Button>
          </div>
        </template>
      </DataTable>
    </Card>

    <!-- Dialogs -->
    <CreateRawMaterialPRDialog
      :open="showCreatePRDialog"
      @update:open="showCreatePRDialog = $event"
      :factory-id="selectedFactoryId"
      :pr="editingPR"
      @success="fetchPRs"
    />

    <RawMaterialPRDetailDialog
      :open="showDetailDialog"
      @update:open="showDetailDialog = $event"
      :pr="selectedPR"
      @convert-to-po="openConvertPO"
      @success="fetchPRs"
    />

    <CreateRawMaterialPODialog
      :open="showCreatePODialog"
      @update:open="showCreatePODialog = $event"
      :factory-id="selectedFactoryId"
      :prefilled-pr="prefilledPRForPO"
      @success="handlePOSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import DataTable from '@/components/DataTable.vue'
import { ClipboardList, Plus, RefreshCw, Eye, ShoppingCart } from 'lucide-vue-next'
import { useFactoryStore, type RawMaterialPR } from '@/stores/factory'
import { useAuthStore } from '@/stores/auth'
import CreateRawMaterialPRDialog from '@/components/factory/CreateRawMaterialPRDialog.vue'
import RawMaterialPRDetailDialog from '@/components/factory/RawMaterialPRDetailDialog.vue'
import CreateRawMaterialPODialog from '@/components/factory/CreateRawMaterialPODialog.vue'
import { toast } from 'vue-sonner'

const factoryStore = useFactoryStore()
const authStore = useAuthStore()

const selectedFactoryId = ref('')
const prStatusFilter = ref('')
const showCreatePRDialog = ref(false)
const showDetailDialog = ref(false)
const showCreatePODialog = ref(false)
const selectedPR = ref<RawMaterialPR | null>(null)
const editingPR = ref<RawMaterialPR | null>(null)
const prefilledPRForPO = ref<RawMaterialPR | null>(null)

const userRole = computed(() => authStore.user?.role || '')
const canReview = computed(() => {
  return ['superadmin', 'admin_purchasing', 'manager'].includes(userRole.value)
})

const prColumns = [
  { header: 'No. Request', accessorKey: 'pr_number' },
  { header: 'Tgl Pengajuan', accessorKey: 'request_date' },
  { header: 'Tgl Dibutuhkan', accessorKey: 'required_date' },
  { header: 'Ringkasan Bahan', accessorKey: 'items_summary' },
  { header: 'Status', accessorKey: 'status' },
  { header: 'Aksi', accessorKey: 'actions', align: 'right' as const },
]

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
      return 'Menunggu Review Purchasing'
    case 'PROCESSED':
      return 'Sudah Dibuat PO'
    case 'REJECTED':
      return 'Ditolak'
    case 'CANCELLED':
      return 'Dibatalkan'
    default:
      return status || 'PENDING'
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

const fetchPRs = () => {
  if (!selectedFactoryId.value) return
  factoryStore.fetchRawMaterialPRs({
    factory_id: selectedFactoryId.value,
    status: prStatusFilter.value || undefined,
  })
}

const openCreatePR = () => {
  editingPR.value = null
  showCreatePRDialog.value = true
}

const openPRDetail = (pr: RawMaterialPR) => {
  selectedPR.value = pr
  showDetailDialog.value = true
}

const openEditPR = (pr: RawMaterialPR) => {
  editingPR.value = pr
  showCreatePRDialog.value = true
}

const openConvertPO = (pr: RawMaterialPR) => {
  showDetailDialog.value = false
  prefilledPRForPO.value = pr
  showCreatePODialog.value = true
}

const handlePOSuccess = () => {
  fetchPRs()
}

const deletePR = async (pr: RawMaterialPR) => {
  if (confirm(`Apakah Anda yakin ingin menghapus Purchase Request ${pr.pr_number}?`)) {
    try {
      await factoryStore.deleteRawMaterialPR(pr.id, selectedFactoryId.value)
      toast.success('Purchase Request berhasil dihapus')
    } catch (err: any) {
      toast.error(typeof err === 'string' ? err : err.message || 'Gagal menghapus PR')
    }
  }
}

watch(
  () => factoryStore.factories,
  (factories) => {
    if (factories && factories.length > 0 && !selectedFactoryId.value && factories[0]) {
      selectedFactoryId.value = factories[0].id
      fetchPRs()
    }
  },
  { immediate: true }
)

onMounted(async () => {
  if (factoryStore.factories.length === 0) {
    await factoryStore.fetchFactories()
    if (factoryStore.factories.length > 0 && factoryStore.factories[0]) {
      selectedFactoryId.value = factoryStore.factories[0].id
    }
  } else if (!selectedFactoryId.value && factoryStore.factories[0]) {
    selectedFactoryId.value = factoryStore.factories[0].id
  }

  if (selectedFactoryId.value) {
    fetchPRs()
  }

  // Pre-fetch raw material products and warehouses
  if (factoryStore.products.length === 0) {
    factoryStore.fetchProducts()
  }
  if (selectedFactoryId.value) {
    factoryStore.fetchFactoryWarehouses(selectedFactoryId.value)
  }
})
</script>
