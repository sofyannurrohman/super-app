<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useSalesOrderStore, type SalesOrder } from '@/stores/salesOrder'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import {
  Truck,
  Package,
  Search,
  RefreshCw,
  ChevronRight,
  Calendar,
  User,
  MapPin,
  ClipboardList,
  CheckCircle2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const soStore = useSalesOrderStore()
const masterStore = useMasterDataStore()

const searchQuery = ref('')
const loading = ref(false)

const filteredOrders = computed(() => {
  return soStore.orders.filter((o) => {
    const isApproved = o.status === 'WAITING_WAREHOUSE' || o.status === 'PROCESSING'
    const q = searchQuery.value.toLowerCase()
    return (
      isApproved &&
      (!q || o.so_number.toLowerCase().includes(q) || o.store?.name?.toLowerCase().includes(q))
    )
  })
})

async function fetchOrders() {
  if (!masterStore.selectedBranchId) return
  await soStore.fetchByBranch(masterStore.selectedBranchId)
}

const formatDate = (d: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
    : '-'

onMounted(() => {
  if (masterStore.selectedBranchId) fetchOrders()
})
</script>

<template>
  <div class="space-y-8 p-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center gap-3">
          <Truck class="w-8 h-8 text-indigo-600" />
          Manajemen Pengiriman
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Kelola daftar pesanan yang siap diproses oleh gudang dan dikirim ke pelanggan.
        </p>
      </div>
      <Button @click="fetchOrders" variant="outline" class="rounded-2xl shadow-sm">
        <RefreshCw class="w-4 h-4 mr-2" :class="{ 'animate-spin': soStore.loading }" /> Segarkan
        Antrian
      </Button>
    </div>

    <!-- Search & Stats -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="relative md:col-span-2">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari No. Pesanan atau Nama Toko..."
          class="w-full pl-11 pr-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-medium outline-none focus:ring-2 focus:ring-indigo-500/20 shadow-sm"
        />
      </div>
      <div
        class="bg-indigo-50 rounded-2xl p-3 px-6 flex items-center justify-between border border-indigo-100"
      >
        <span class="text-[10px] font-black text-indigo-400 uppercase tracking-widest"
          >Siap Proses</span
        >
        <span class="text-2xl font-black text-indigo-700">{{ filteredOrders.length }}</span>
      </div>
      <div
        class="bg-slate-50 rounded-2xl p-3 px-6 flex items-center justify-between border border-slate-100"
      >
        <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
          >Total Antrian</span
        >
        <span class="text-2xl font-black text-slate-700">{{ soStore.orders.length }}</span>
      </div>
    </div>

    <!-- List -->
    <div v-if="soStore.loading" class="flex justify-center py-24">
      <RefreshCw class="w-8 h-8 animate-spin text-indigo-600" />
    </div>

    <div
      v-else-if="filteredOrders.length === 0"
      class="text-center py-24 bg-white border border-slate-200 rounded-[40px] shadow-sm"
    >
      <Package class="w-16 h-16 text-slate-200 mx-auto mb-4" />
      <h3 class="text-lg font-black text-slate-900">Belum Ada Antrian Pengiriman</h3>
      <p class="text-slate-500 text-sm mt-1">
        Pastikan Salesman telah mengonfirmasi Pesanan Order (SO) di aplikasi mobile.
      </p>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div
        v-for="order in filteredOrders"
        :key="order.id"
        class="bg-white border border-slate-200 rounded-[40px] p-8 shadow-sm hover:shadow-xl transition-all relative overflow-hidden group"
      >
        <div class="flex flex-col sm:flex-row gap-6">
          <!-- Info Column -->
          <div class="flex-1">
            <div class="flex items-center gap-2 mb-4">
              <span
                class="text-[10px] font-black bg-indigo-100 text-indigo-700 px-3 py-1 rounded-full uppercase tracking-widest"
              >
                {{ order.status === 'PROCESSING' ? 'SEDANG DIPROSES' : 'MENUNGGU GUDANG' }}
              </span>
              <span class="text-[10px] font-bold text-slate-400">{{
                formatDate(order.created_at)
              }}</span>
            </div>

            <h4 class="text-2xl font-black text-slate-900 mb-1 tracking-tight">
              {{ order.store?.name }}
            </h4>
            <p class="text-xs text-slate-500 font-medium mb-6 flex items-center gap-1.5 italic">
              {{ order.so_number }}
            </p>

            <div class="grid grid-cols-2 gap-4 mb-6">
              <div class="bg-slate-50 p-4 rounded-2xl border border-slate-100">
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">
                  Salesman
                </p>
                <p class="text-xs font-black text-slate-700">{{ order.employee?.first_name }}</p>
              </div>
              <div class="bg-slate-50 p-4 rounded-2xl border border-slate-100">
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">
                  Estimasi Berat
                </p>
                <p class="text-xs font-black text-slate-700">-- kg</p>
              </div>
            </div>

            <div
              class="flex items-start gap-2 text-xs text-slate-500 bg-slate-50/50 p-4 rounded-2xl italic"
            >
              <MapPin class="w-4 h-4 shrink-0 text-slate-300" />
              {{ order.store?.address }}
            </div>
          </div>

          <!-- Action Column -->
          <div class="sm:w-48 flex flex-col justify-center gap-3">
            <Button
              class="w-full bg-slate-900 hover:bg-indigo-600 text-white rounded-2xl py-6 font-black text-xs transition-all flex items-center justify-center gap-2"
            >
              Detail Pesanan <ChevronRight class="w-4 h-4" />
            </Button>
            <p class="text-[9px] text-center text-slate-400 font-bold uppercase tracking-tighter">
              Klik untuk cetak Packing List
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
