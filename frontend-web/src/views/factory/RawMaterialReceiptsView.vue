<template>
  <div class="p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-6 rounded-3xl shadow-sm border border-slate-100"
    >
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900">
          Penerimaan Supplier & Bongkar Muat
        </h1>
        <p class="text-sm text-slate-500 mt-0.5">
          Pencatatan kedatangan bahan baku, verifikasi Surat Jalan supplier, dan ongkos bongkar kuli (Kas Pabrik).
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <select
          v-model="selectedFactoryId"
          class="text-sm bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 font-medium focus:outline-none focus:ring-2 focus:ring-primary/20"
        >
          <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">{{ f.name }}</option>
        </select>
        <Button variant="outline" class="rounded-xl border-slate-200" @click="fetchReceipts">
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': factoryStore.loading }" />
          Segarkan
        </Button>
        <Button
          @click="showReceiptDialog = true"
          :disabled="!selectedFactoryId"
          class="rounded-xl font-bold shadow-md shadow-emerald-500/10 bg-emerald-600 hover:bg-emerald-700 text-white"
        >
          <PackageCheck class="mr-2 h-4 w-4" /> Catat Penerimaan Supplier
        </Button>
      </div>
    </div>

    <!-- Table Card -->
    <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden p-2">
      <DataTable
        :columns="receiptColumns"
        :data="factoryStore.rawMaterialReceipts"
        :isLoading="factoryStore.loading"
      >
        <template #cell-receipt_no="{ row }">
          <div class="flex flex-col">
            <span class="font-black text-slate-900 text-sm font-mono">{{ (row as any).receipt_no }}</span>
            <span class="text-xs text-slate-400 font-medium">{{ formatDateOnly((row as any).received_date) }}</span>
          </div>
        </template>
        <template #cell-supplier="{ row }">
          <span class="font-bold text-slate-700 text-sm">
            {{ (row as any).supplier?.name || '-' }}
          </span>
        </template>
        <template #cell-supplier_delivery_note_no="{ row }">
          <Badge variant="outline" class="font-mono text-xs font-semibold bg-slate-50">
            {{ (row as any).supplier_delivery_note_no || '-' }}
          </Badge>
        </template>
        <template #cell-supplier_invoice_no="{ row }">
          <span class="text-xs font-mono font-semibold text-slate-700">
            {{ (row as any).supplier_invoice_no || 'Belum ada' }}
          </span>
        </template>
        <template #cell-po_ref="{ row }">
          <span v-if="(row as any).purchase_order" class="font-mono text-xs text-primary font-bold">
            {{ (row as any).purchase_order?.po_number }}
          </span>
          <span v-else class="text-slate-400 text-xs">-</span>
        </template>
        <template #cell-items_received="{ row }">
          <div class="text-xs text-slate-700 space-y-0.5">
            <div v-for="it in ((row as any).items || [])" :key="it.id" class="flex items-center gap-1.5">
              <span class="font-medium">{{ it.product?.name }}:</span>
              <span class="font-bold text-emerald-700 font-mono">+{{ it.received_qty }} {{ it.unit }}</span>
              <span v-if="it.rejected_qty > 0" class="text-red-500 font-mono text-[10px]">(Rijek: {{ it.rejected_qty }})</span>
            </div>
          </div>
        </template>
        <template #cell-unloading_cost="{ row }">
          <div v-if="(row as any).unloading_cost > 0" class="flex flex-col text-xs">
            <span class="font-mono font-bold text-amber-800">
              Rp {{ ((row as any).unloading_cost || 0).toLocaleString('id-ID') }}
            </span>
            <span v-if="(row as any).kas_pabrik" class="text-[10px] text-slate-400">
              Kas: {{ (row as any).kas_pabrik?.name }}
            </span>
          </div>
          <span v-else class="text-xs text-slate-400">-</span>
        </template>
        <template #cell-received_by_name="{ row }">
          <span class="text-xs font-semibold text-slate-600">
            {{ (row as any).received_by_name || 'Admin' }}
          </span>
        </template>
        <template #cell-actions="{ row }">
          <Button v-if="((row as any).invoice_total || 0) > ((row as any).paid_amount || 0)" size="sm" class="bg-emerald-600 text-white" @click="openPayment(row)">Bayar</Button>
        </template>
      </DataTable>
    </Card>

    <!-- Dialog -->
    <RawMaterialReceiptDialog
      v-model:open="showReceiptDialog"
      :factoryId="selectedFactoryId"
      @success="fetchReceipts"
    />
    <div v-if="paymentReceipt" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <form class="bg-white rounded-2xl p-6 w-full max-w-sm space-y-4" @submit.prevent="payReceipt">
        <h2 class="font-black text-lg">Bayar Faktur Bahan</h2>
        <p class="text-xs text-slate-500">Sisa: Rp {{ remaining.toLocaleString('id-ID') }}</p>
        <input v-model.number="paymentAmount" type="number" min="1" :max="remaining" required class="w-full border rounded-xl p-3" />
        <div class="flex justify-end gap-2"><Button type="button" variant="outline" @click="paymentReceipt = null">Batal</Button><Button class="bg-emerald-600 text-white">Simpan</Button></div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { computed } from 'vue'
import apiClient from '@/api/axios'
import { useFactoryStore } from '@/stores/factory'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import DataTable from '@/components/DataTable.vue'
import { RefreshCw, PackageCheck } from 'lucide-vue-next'
import RawMaterialReceiptDialog from '@/components/factory/RawMaterialReceiptDialog.vue'
import { toast } from 'vue-sonner'

const factoryStore = useFactoryStore()
const selectedFactoryId = ref('')
const showReceiptDialog = ref(false)
const paymentReceipt = ref<any>(null)
const paymentAmount = ref(0)
const remaining = computed(() => Math.max(0, (paymentReceipt.value?.invoice_total || 0) - (paymentReceipt.value?.paid_amount || 0)))

const receiptColumns = [
  { key: 'receipt_no', label: 'No. Penerimaan', sortable: true },
  {
    id: 'supplier',
    header: 'Supplier',
    accessorFn: (row: any) => row.supplier?.name || '',
    sortable: true,
  },
  { key: 'supplier_delivery_note_no', label: 'Surat Jalan Supplier', sortable: true },
  { key: 'supplier_invoice_no', label: 'Faktur Supplier', sortable: true },
  {
    id: 'po_ref',
    header: 'Ref. PO',
    accessorFn: (row: any) => row.purchase_order?.po_number || '',
    sortable: true,
  },
  {
    id: 'items_received',
    header: 'Bahan Masuk',
    accessorFn: (row: any) => (row.items || []).map((it: any) => it.product?.name || '').join(' '),
    sortable: false,
  },
  { key: 'unloading_cost', label: 'Ongkos Bongkar', sortable: true },
  { key: 'received_by_name', label: 'Diterima Oleh', sortable: true },
  { key: 'actions', label: 'Aksi', sortable: false },
]

const openPayment = (row: any) => { paymentReceipt.value = row; paymentAmount.value = remaining.value }
const payReceipt = async () => {
  try { await apiClient.post(`/factory/raw-materials/receipts/${paymentReceipt.value.id}/pay`, { amount: paymentAmount.value }); toast.success('Pembayaran berhasil dicatat'); paymentReceipt.value = null; fetchReceipts() }
  catch (err: any) { toast.error(err.response?.data?.error || 'Gagal mencatat pembayaran') }
}

const formatDateOnly = (val?: string) => {
  if (!val) return '-'
  return new Date(val).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const fetchReceipts = () => {
  if (!selectedFactoryId.value) return
  factoryStore.fetchRawMaterialReceipts({ factory_id: selectedFactoryId.value })
}

onMounted(async () => {
  await factoryStore.fetchFactories()
  if (factoryStore.factories.length > 0 && factoryStore.factories[0]) {
    selectedFactoryId.value = factoryStore.factories[0].id
  }
})

watch(selectedFactoryId, (newId) => {
  if (newId) {
    fetchReceipts()
  }
})
</script>
