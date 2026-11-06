<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useSalesOrderStore, type SalesOrder } from '@/stores/salesOrder'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import {
  Truck,
  CheckCircle2,
  MapPin,
  Calendar,
  Search,
  RefreshCw,
  Camera,
  X,
  User,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const soStore = useSalesOrderStore()
const masterStore = useMasterDataStore()

const searchQuery = ref('')
const showPODModal = ref(false)
const selectedOrder = ref<SalesOrder | null>(null)
const loading = ref(false)

const podData = ref({
  received_by: '',
  pod_image_url: '',
})

const filteredOrders = computed(() => {
  return soStore.orders.filter((o) => {
    const isShipped = o.status === 'SHIPPED'
    const q = searchQuery.value.toLowerCase()
    return (
      isShipped &&
      (!q || o.so_number.toLowerCase().includes(q) || o.store?.name?.toLowerCase().includes(q))
    )
  })
})

async function fetchOrders() {
  if (!masterStore.selectedWarehouseBranchId) return
  await soStore.fetchByBranch(masterStore.selectedWarehouseBranchId, 'SHIPPED')
}

function openPODModal(order: SalesOrder) {
  selectedOrder.value = order
  podData.value = { received_by: '', pod_image_url: '' }
  showPODModal.value = true
}

async function handleConfirmPOD() {
  if (!selectedOrder.value || !podData.value.received_by) {
    toast.error('Nama penerima harus diisi')
    return
  }

  loading.value = true
  try {
    await soStore.confirmPOD(selectedOrder.value.id, podData.value)
    toast.success('Penerimaan barang berhasil dikonfirmasi')
    showPODModal.value = false
    await fetchOrders()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal konfirmasi POD')
  } finally {
    loading.value = false
  }
}

const formatDate = (d: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
    : '-'

watch(
  () => masterStore.selectedWarehouseBranchId,
  () => {
    if (masterStore.selectedWarehouseBranchId) fetchOrders()
  }
)

onMounted(() => {
  masterStore.fetchBranches()
  if (masterStore.selectedWarehouseBranchId) fetchOrders()
})
</script>

<template>
  <div class="space-y-8 p-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center gap-3">
          <Truck class="w-8 h-8 text-indigo-600" />
          Kontrol Pengiriman (POD)
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Konfirmasi penerimaan barang oleh pelanggan untuk mengaktifkan penagihan.
        </p>
      </div>
      <Button
        @click="fetchOrders"
        variant="outline"
        class="rounded-2xl flex items-center gap-2 shadow-sm"
      >
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': soStore.loading }" /> Segarkan Data
      </Button>
    </div>

    <!-- Search -->
    <div class="relative max-w-md">
      <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Cari No. SO atau Nama Toko..."
        class="w-full pl-11 pr-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-medium outline-none focus:ring-2 focus:ring-indigo-500/20"
      />
    </div>

    <!-- Empty State -->
    <div
      v-if="filteredOrders.length === 0 && !soStore.loading"
      class="text-center py-24 bg-slate-50 rounded-[40px] border-2 border-dashed border-slate-200"
    >
      <div
        class="bg-white w-20 h-20 rounded-3xl shadow-sm flex items-center justify-center mx-auto mb-6"
      >
        <Truck class="w-10 h-10 text-slate-300" />
      </div>
      <h3 class="text-xl font-black text-slate-900">Tidak Ada Pengiriman Aktif</h3>
      <p class="text-slate-500 mt-2 max-w-xs mx-auto text-sm font-medium">
        Semua barang yang dikirim telah dikonfirmasi atau belum ada barang keluar dari gudang.
      </p>
    </div>

    <!-- List -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="order in filteredOrders"
        :key="order.id"
        class="bg-white border border-slate-200 rounded-[32px] p-6 shadow-sm hover:shadow-xl transition-all group border-b-4 border-b-indigo-500"
      >
        <div class="flex items-center justify-between mb-4">
          <span
            class="text-[10px] font-black bg-indigo-100 text-indigo-700 px-3 py-1 rounded-full uppercase tracking-widest"
            >IN TRANSIT</span
          >
          <span class="text-[10px] font-mono text-slate-400"
            >DO: {{ order.delivery_order_no }}</span
          >
        </div>

        <h4 class="text-xl font-black text-slate-900 mb-1">{{ order.store?.name }}</h4>
        <p class="text-xs text-slate-500 font-medium mb-6 flex items-center gap-1.5">
          <MapPin class="w-3.5 h-3.5" /> {{ order.store?.address }}
        </p>

        <div class="space-y-3 mb-6">
          <div
            class="flex items-center gap-3 text-xs text-slate-600 font-bold bg-slate-50 p-3 rounded-2xl"
          >
            <Calendar class="w-4 h-4 text-slate-400" /> Dikirim: {{ formatDate(order.shipped_at!) }}
          </div>
          <div
            class="flex items-center gap-3 text-xs text-slate-600 font-bold bg-slate-50 p-3 rounded-2xl"
          >
            <User class="w-4 h-4 text-slate-400" /> Sales: {{ order.employee?.first_name }}
          </div>
        </div>

        <Button
          @click="openPODModal(order)"
          class="w-full bg-slate-900 hover:bg-indigo-600 text-white rounded-2xl py-6 font-black text-xs flex items-center justify-center gap-2 transition-all"
        >
          <CheckCircle2 class="w-4 h-4" /> Konfirmasi Barang Sampai
        </Button>
      </div>
    </div>

    <!-- POD Modal -->
    <div
      v-if="showPODModal && selectedOrder"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
    >
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showPODModal = false"
      />
      <div
        class="bg-white rounded-[40px] w-full max-w-md relative z-10 shadow-2xl overflow-hidden flex flex-col"
      >
        <div class="p-8 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-2xl font-black text-slate-900">Bukti Terima (POD)</h3>
          <button
            @click="showPODModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400"
          >
            <X class="w-6 h-6" />
          </button>
        </div>

        <div class="p-8 space-y-6">
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Nama Penerima</label
            >
            <input
              v-model="podData.received_by"
              type="text"
              placeholder="Siapa yang menerima barang?"
              class="w-full bg-slate-50 border-2 border-transparent focus:border-indigo-500 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
            />
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Foto Bukti (Opsional)</label
            >
            <div
              class="border-2 border-dashed border-slate-200 rounded-3xl p-8 text-center hover:border-indigo-400 transition-all cursor-pointer bg-slate-50"
            >
              <Camera class="w-10 h-10 text-slate-300 mx-auto mb-3" />
              <p class="text-xs font-black text-slate-400">Klik untuk ambil foto / upload</p>
            </div>
          </div>
        </div>

        <div class="p-8 border-t border-slate-100 bg-slate-50/50 flex gap-4">
          <Button
            @click="handleConfirmPOD"
            :disabled="loading"
            class="w-full bg-indigo-600 hover:bg-indigo-700 text-white rounded-2xl py-7 font-black shadow-xl shadow-indigo-500/20"
          >
            {{ loading ? 'Menyimpan...' : 'Konfirmasi Penerimaan' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
