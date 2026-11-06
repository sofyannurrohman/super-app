<template>
  <div>
    <div class="p-6 lg:p-8 space-y-8 bg-[#f8fafc] min-h-screen print:hidden">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-6 lg:p-8 rounded-[2rem] shadow-sm border border-slate-100">
      <div class="flex items-center gap-5">
        <div class="w-14 h-14 rounded-[1.5rem] bg-gradient-to-br from-indigo-500 to-blue-600 flex items-center justify-center text-white shadow-xl shadow-indigo-200">
          <ClipboardList class="h-7 w-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900">Kelola Left Over</h1>
          <p class="text-sm text-slate-500 font-medium mt-0.5">Manajemen barang susulan, gagal kirim, retur rusak & klaim pelanggan</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 rounded-full bg-blue-500 animate-pulse" v-if="pendingCount > 0"></div>
        <span v-if="pendingCount > 0" class="text-sm font-black text-blue-700 bg-blue-50 border border-blue-200 px-4 py-2 rounded-2xl">
          {{ pendingCount }} Item menunggu jadwal
        </span>
      </div>
    </div>

    <!-- Status Filter and Actions -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div class="flex flex-wrap gap-2">
        <button
          v-for="f in statusFilters"
          :key="f.value"
          @click="activeFilter = f.value"
          :class="activeFilter === f.value
            ? 'bg-indigo-500 text-white shadow-lg shadow-indigo-200'
            : 'bg-white text-slate-600 border border-slate-200 hover:border-indigo-300'"
          class="px-4 py-2 rounded-xl font-bold text-xs transition-all"
        >
          {{ f.label }}
        </button>
      </div>
      <button
        @click="printDO"
        class="flex items-center gap-2 bg-slate-900 text-white px-5 py-2.5 rounded-xl text-sm font-bold shadow-md hover:bg-slate-800 transition-all active:scale-95"
      >
        <Printer class="w-4 h-4" /> Cetak DO Spare / Susulan
      </button>
    </div>

    <!-- Content -->
    <div v-if="isLoading" class="flex items-center justify-center py-20">
      <div class="w-8 h-8 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
    </div>

    <div v-else-if="filteredItems.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400 bg-white rounded-[2rem] border border-slate-100">
      <ClipboardList class="h-16 w-16 mb-4 opacity-20" />
      <p class="font-bold text-lg">Tidak ada item left over</p>
      <p class="text-sm">Semua barang susulan telah ditangani</p>
    </div>

    <div v-else class="grid grid-cols-1 xl:grid-cols-2 gap-5">
      <div
        v-for="item in filteredItems"
        :key="item.id"
        class="bg-white rounded-[2rem] border shadow-sm overflow-hidden transition-all hover:shadow-md"
        :class="{
          'border-amber-200': item.status === 'PENDING',
          'border-blue-200': item.status === 'IN_DELIVERY',
          'border-purple-200': item.status === 'SHIPPED',
          'border-emerald-200': item.status === 'DELIVERED',
          'border-slate-100': !['PENDING','IN_DELIVERY','SHIPPED','DELIVERED'].includes(item.status),
        }"
      >
        <!-- Card Header -->
        <div class="px-6 py-5 border-b border-slate-50 flex items-start justify-between gap-4">
          <div>
            <div class="flex items-center gap-2 mb-1">
              <span :class="statusBadge(item.status)" class="inline-flex px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest">
                {{ statusLabel(item.status) }}
              </span>
              <span class="text-xs text-slate-400 font-medium">{{ formatDate(item.created_at) }}</span>
            </div>
            <p class="font-black text-lg text-slate-900">{{ item.product?.name || 'Produk Tidak Diketahui' }}</p>
            <p class="text-sm text-slate-500 font-medium mt-0.5">
              Ref: {{ getReferenceInfo(item) }}
            </p>
          </div>
          <div class="text-right shrink-0">
            <p class="text-xs text-slate-400 font-bold">Qty Susulan</p>
            <p class="text-xl font-black text-slate-900 leading-tight">{{ item.quantity || 0 }}</p>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="px-6 py-5 border-t border-slate-50 flex flex-wrap gap-2">
          <template v-if="item.status === 'PENDING'">
            <button
              class="flex-1 py-2.5 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl text-sm font-bold shadow-sm hover:shadow-lg hover:-translate-y-0.5 transition-all flex items-center justify-center gap-2"
              @click="handleAjukan(item)"
            >
              <Send class="h-4 w-4" /> Ajukan ke Admin Nota
            </button>
          </template>
          
          <template v-else-if="item.status === 'IN_DELIVERY'">
            <div class="flex-1 flex items-center justify-center gap-2 text-sm text-blue-600 font-bold py-2.5 bg-blue-50 rounded-xl">
              <ClipboardCheck class="h-4 w-4" /> Masuk antrean Surat Jalan
            </div>
          </template>
        </div>
      </div>
    </div>
    </div>

    <!-- Print Layout -->
    <div class="hidden print:block bg-white text-black font-sans w-full p-4">
      <div class="border-b-2 border-black pb-4 mb-4">
        <h2 class="text-2xl font-black uppercase text-center mb-2 tracking-wide">LEMBAR BUKTI MUAT KENDARAAN (DO) SUSULAN</h2>
        <div class="flex justify-between text-sm mt-4">
           <div>
              <p><strong>Status Filter:</strong> {{ activeFilter ? statusLabel(activeFilter) : 'Semua' }}</p>
           </div>
           <div class="text-right">
              <p><strong>Tanggal Cetak:</strong> {{ new Date().toLocaleDateString('id-ID') }}</p>
           </div>
        </div>
      </div>
      
      <table class="w-full text-left border-collapse border border-black text-xs">
        <thead>
          <tr>
             <th class="border border-black p-2 w-10 text-center">No</th>
             <th class="border border-black p-2">Barang (Nama)</th>
             <th class="border border-black p-2 w-24 text-center">Qty Susulan</th>
             <th class="border border-black p-2">Referensi</th>
             <th class="border border-black p-2">Status</th>
             <th class="border border-black p-2 w-20 text-center">Check</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in filteredItems" :key="item.id">
             <td class="border border-black p-1.5 text-center">{{ index + 1 }}</td>
             <td class="border border-black p-1.5">
                <span class="font-bold">{{ item.product?.name || 'Produk Tidak Diketahui' }}</span>
             </td>
             <td class="border border-black p-1.5 text-center font-bold text-lg">
                {{ item.quantity || 0 }}
             </td>
             <td class="border border-black p-1.5">
                {{ getReferenceInfo(item) }}
             </td>
             <td class="border border-black p-1.5">
                {{ statusLabel(item.status) }}
             </td>
             <td class="border border-black p-1.5"></td>
          </tr>
        </tbody>
      </table>
      
      <div class="mt-8 flex justify-between px-10 text-sm">
         <div class="text-center">
            <p>Disiapkan Oleh (Gudang)</p>
            <br><br><br>
            <p>(.........................)</p>
         </div>
         <div class="text-center">
            <p>Diperiksa Oleh (Checker)</p>
            <br><br><br>
            <p>(.........................)</p>
         </div>
         <div class="text-center">
            <p>Diterima Oleh (Supir)</p>
            <br><br><br>
            <p>(.........................)</p>
         </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  ClipboardList, Package, Truck, Boxes, Send, ClipboardCheck, Printer
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterStore = useMasterDataStore()

const items = ref<any[]>([])
const isLoading = ref(false)
const activeFilter = ref('')

const statusFilters = [
  { label: 'Semua', value: '' },
  { label: '🕒 Menunggu', value: 'PENDING' },
  { label: '📦 Masuk DO', value: 'IN_DELIVERY' },
  { label: '🚀 Dikirim', value: 'SHIPPED' },
  { label: '✅ Selesai', value: 'DELIVERED' },
]

const pendingCount = computed(() => items.value.filter(i => i.status === 'PENDING').length)

const filteredItems = computed(() =>
  items.value.filter(i => !activeFilter.value || i.status === activeFilter.value)
)

const statusBadge = (s: string) => {
  const m: Record<string, string> = {
    PENDING: 'bg-amber-50 text-amber-700 border border-amber-200',
    IN_DELIVERY: 'bg-blue-50 text-blue-700 border border-blue-200',
    SHIPPED: 'bg-purple-50 text-purple-700 border border-purple-200',
    DELIVERED: 'bg-emerald-50 text-emerald-700 border border-emerald-200',
  }
  return m[s] || 'bg-slate-100 text-slate-600'
}

const statusLabel = (s: string) => {
  const m: Record<string, string> = {
    PENDING: 'Menunggu', IN_DELIVERY: 'Masuk Antrean SJ',
    SHIPPED: 'Sedang Dikirim', DELIVERED: 'Selesai',
  }
  return m[s] || s
}

const formatDate = (d: string) => d ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '-'

const getReferenceInfo = (item: any) => {
  if (item.sales_order_id && item.sales_order) return `SO: ${item.sales_order.order_number}`
  if (item.product_exchange_id && item.product_exchange) return `Tukar: ${item.product_exchange.exchange_number}`
  return 'Gagal Kirim / Retur / Klaim'
}

const fetchLeftovers = async () => {
  isLoading.value = true
  try {
    const branchId = masterStore.selectedWarehouseBranchId
    const url = branchId && branchId !== 'ALL_BRANCHES'
      ? `/delivery/subsequent/pending?branch_id=${branchId}`
      : '/delivery/subsequent/pending'
    const res = await apiClient.get(url)
    items.value = res.data?.data || []
  } catch (err: any) {
    if (err.response?.status !== 404) {
      toast.error('Gagal memuat data left over')
    }
    items.value = []
  } finally {
    isLoading.value = false
  }
}

const handleAjukan = (item: any) => {
  // Dalam sistem ini, PENDING akan muncul di menu pembuatan DO/Surat Jalan oleh Admin Nota
  toast.success(`Left over untuk ${item.product?.name || 'produk'} telah diajukan ke antrean pembuatan Surat Jalan berikutnya.`)
}

const printDO = () => {
  window.print()
}

watch(
  () => masterStore.selectedWarehouseBranchId,
  () => {
    fetchLeftovers()
  }
)

onMounted(async () => {
  masterStore.fetchBranches()
  await fetchLeftovers()
})
</script>

<style scoped>
@media print {
  @page { margin: 0.5cm; }
}
</style>
