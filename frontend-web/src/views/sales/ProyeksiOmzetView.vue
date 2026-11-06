<script setup lang="ts">
import { ref, onMounted } from 'vue'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { Calendar, Search, TrendingUp, Truck, User, DollarSign } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const masterStore = useMasterDataStore()

const loading = ref(false)
const filterMode = ref<'daily' | 'monthly'>('daily')
const selectedDate = ref(
  new Date(new Date().getTime() + 24 * 60 * 60 * 1000).toISOString().split('T')[0]
) // Default to tomorrow
const selectedMonth = ref(
  new Date().toISOString().substring(0, 7)
)

const data = ref<any>(null)

async function fetchProyeksi() {
  const dateParam = filterMode.value === 'daily' ? selectedDate.value : selectedMonth.value
  if (!dateParam) return
  loading.value = true
  data.value = null
  try {
    const params: any = { date: dateParam }
    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterStore.selectedBranchId
    }
    const res = await apiClient.get('/admin/sales/proyeksi-omzet', { params })
    data.value = res.data
  } catch (err: any) {
    toast.error('Gagal memuat proyeksi omzet: ' + (err.response?.data?.error || err.message))
  } finally {
    loading.value = false
  }
}

const formatCurrency = (val: number) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val || 0)

const formatDate = (d: string) => {
  if (!d) return '-'
  if (d.length === 7) {
    return new Date(d + '-01').toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
  }
  return new Date(d).toLocaleDateString('id-ID', { weekday: 'long', day: '2-digit', month: 'long', year: 'numeric' })
}

onMounted(() => {
  fetchProyeksi()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header & Filter -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-5 bg-white/70 backdrop-blur-xl p-7 rounded-[2rem] shadow-sm border border-slate-100/50">
      <div class="flex items-center gap-4">
        <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-lg shadow-indigo-200">
          <TrendingUp class="h-7 w-7 text-white" />
        </div>
        <div>
          <h1 class="text-2xl font-black text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <span>Proyeksi Omzet</span>
            <span v-if="masterStore.selectedBranchName" class="text-primary text-xl font-bold">
              — {{ masterStore.selectedBranchName }}
            </span>
          </h1>
          <p class="text-sm text-slate-500 font-medium mt-0.5">Pantau perkiraan omzet dari Pengiriman TO dan Sales Motoris</p>
        </div>
      </div>

      <div class="flex flex-wrap gap-3 items-center">
        <div class="relative min-w-[150px]">
          <select
            v-model="filterMode"
            @change="fetchProyeksi"
            class="w-full px-4 py-2.5 rounded-xl border border-slate-200 focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100 outline-none text-sm font-medium transition-all bg-white"
          >
            <option value="daily">Harian</option>
            <option value="monthly">Bulanan</option>
          </select>
        </div>
        <div class="relative min-w-[200px]">
          <Calendar class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-if="filterMode === 'daily'"
            v-model="selectedDate"
            type="date"
            class="w-full pl-9 pr-4 py-2.5 rounded-xl border border-slate-200 focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100 outline-none text-sm font-medium transition-all bg-white"
          />
          <input
            v-else
            v-model="selectedMonth"
            type="month"
            class="w-full pl-9 pr-4 py-2.5 rounded-xl border border-slate-200 focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100 outline-none text-sm font-medium transition-all bg-white"
          />
        </div>
        <button
          @click="fetchProyeksi"
          :disabled="loading"
          class="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-slate-900 text-white text-sm font-bold transition-all hover:bg-slate-700 disabled:opacity-50 shadow-md"
        >
          <Search class="h-4 w-4" />
          Tampilkan
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex flex-col items-center justify-center py-24 gap-4">
      <div class="w-14 h-14 rounded-full border-4 border-indigo-100 border-t-indigo-600 animate-spin" />
      <p class="text-indigo-600 font-medium">Memuat data proyeksi...</p>
    </div>

    <div v-else-if="data" class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <!-- Summary Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <!-- TO Card -->
        <div class="bg-white/70 backdrop-blur-xl p-6 rounded-[2rem] border border-slate-100/50 shadow-sm relative overflow-hidden group">
          <div class="absolute -right-4 -top-4 w-24 h-24 bg-blue-50 rounded-full transition-transform group-hover:scale-150 duration-500 z-0"></div>
          <div class="relative z-10">
            <div class="flex items-center gap-3 mb-4">
              <div class="p-2.5 bg-blue-100 rounded-xl text-blue-600">
                <Truck class="h-5 w-5" />
              </div>
              <h3 class="font-bold text-slate-500 text-sm tracking-wider uppercase">Pengiriman TO</h3>
            </div>
            <div class="text-3xl font-black text-slate-800">
              {{ formatCurrency(data.total_to_omzet) }}
            </div>
          </div>
        </div>

        <!-- Motoris Card -->
        <div class="bg-white/70 backdrop-blur-xl p-6 rounded-[2rem] border border-slate-100/50 shadow-sm relative overflow-hidden group">
          <div class="absolute -right-4 -top-4 w-24 h-24 bg-purple-50 rounded-full transition-transform group-hover:scale-150 duration-500 z-0"></div>
          <div class="relative z-10">
            <div class="flex items-center gap-3 mb-4">
              <div class="p-2.5 bg-purple-100 rounded-xl text-purple-600">
                <User class="h-5 w-5" />
              </div>
              <h3 class="font-bold text-slate-500 text-sm tracking-wider uppercase">Sales Motoris</h3>
            </div>
            <div class="text-3xl font-black text-slate-800">
              {{ formatCurrency(data.total_motoris_omzet) }}
            </div>
          </div>
        </div>

        <!-- Pending TO Card -->
        <div class="bg-white/70 backdrop-blur-xl p-6 rounded-[2rem] border border-slate-100/50 shadow-sm relative overflow-hidden group">
          <div class="absolute -right-4 -top-4 w-24 h-24 bg-amber-50 rounded-full transition-transform group-hover:scale-150 duration-500 z-0"></div>
          <div class="relative z-10">
            <div class="flex items-center gap-3 mb-4">
              <div class="p-2.5 bg-amber-100 rounded-xl text-amber-600">
                <Truck class="h-5 w-5" />
              </div>
              <h3 class="font-bold text-slate-500 text-sm tracking-wider uppercase">SO Belum Dikirim</h3>
            </div>
            <div class="text-3xl font-black text-slate-800">
              {{ formatCurrency(data.total_pending_to_omzet) }}
            </div>
          </div>
        </div>

        <!-- Grand Total Card -->
        <div class="bg-gradient-to-br from-indigo-600 to-indigo-800 p-6 rounded-[2rem] shadow-lg shadow-indigo-200 relative overflow-hidden group">
          <div class="absolute -right-4 -top-4 w-32 h-32 bg-white/10 rounded-full transition-transform group-hover:scale-150 duration-500 z-0"></div>
          <div class="relative z-10 text-white">
            <div class="flex items-center gap-3 mb-4">
              <div class="p-2.5 bg-white/20 rounded-xl text-white backdrop-blur-md">
                <DollarSign class="h-5 w-5" />
              </div>
              <h3 class="font-bold text-indigo-100 text-sm tracking-wider uppercase">Total Keseluruhan</h3>
            </div>
            <div class="text-4xl font-black tracking-tight">
              {{ formatCurrency(data.grand_total_omzet) }}
            </div>
            <p class="text-indigo-200 text-xs mt-2 font-medium">Proyeksi tanggal {{ formatDate(data.date) }}</p>
          </div>
        </div>
      </div>

      <!-- Detail Tables -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        <!-- Table Pending TO -->
        <div class="bg-white/70 backdrop-blur-xl rounded-[2rem] border border-slate-100/50 shadow-sm overflow-hidden flex flex-col h-full">
          <div class="p-6 border-b border-slate-100 flex items-center gap-3 bg-white/50">
            <Truck class="h-5 w-5 text-amber-500" />
            <h2 class="text-lg font-black text-slate-800">Rincian SO Belum Dikirim</h2>
          </div>
          <div class="p-6 flex-1 overflow-x-auto">
            <table v-if="data.pending_to_sales && data.pending_to_sales.length > 0" class="w-full text-left text-sm whitespace-nowrap">
              <thead>
                <tr class="text-slate-400 font-bold uppercase tracking-wider text-xs border-b border-slate-100">
                  <th class="pb-3 pr-4">No. SO</th>
                  <th class="pb-3 px-4">Nama Sales</th>
                  <th class="pb-3 pl-4 text-right">Omzet</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50">
                <tr v-for="(item, i) in data.pending_to_sales" :key="i" class="hover:bg-slate-50/50 transition-colors">
                  <td class="py-3 pr-4 font-bold text-slate-700">{{ item.so_number }}</td>
                  <td class="py-3 px-4 text-slate-500">{{ item.salesman_name || '-' }}</td>
                  <td class="py-3 pl-4 text-right font-black text-amber-600">{{ formatCurrency(item.total_omzet) }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="flex flex-col items-center justify-center py-10 text-slate-400">
              <Truck class="h-8 w-8 mb-2 opacity-50" />
              <p class="text-sm font-medium">Semua SO sudah dikirim.</p>
            </div>
          </div>
        </div>

        <!-- Table TO -->
        <div class="bg-white/70 backdrop-blur-xl rounded-[2rem] border border-slate-100/50 shadow-sm overflow-hidden flex flex-col h-full">
          <div class="p-6 border-b border-slate-100 flex items-center gap-3 bg-white/50">
            <Truck class="h-5 w-5 text-blue-500" />
            <h2 class="text-lg font-black text-slate-800">Rincian Pengiriman TO</h2>
          </div>
          <div class="p-6 flex-1 overflow-x-auto">
            <table v-if="data.to_shipments && data.to_shipments.length > 0" class="w-full text-left text-sm whitespace-nowrap">
              <thead>
                <tr class="text-slate-400 font-bold uppercase tracking-wider text-xs border-b border-slate-100">
                  <th class="pb-3 pr-4">No. DO</th>
                  <th class="pb-3 px-4">Supir / Kendaraan</th>
                  <th class="pb-3 pl-4 text-right">Omzet</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50">
                <tr v-for="(item, i) in data.to_shipments" :key="i" class="hover:bg-slate-50/50 transition-colors">
                  <td class="py-3 pr-4 font-bold text-slate-700">{{ item.delivery_order_no }}</td>
                  <td class="py-3 px-4 text-slate-500">{{ item.vehicle_plate_no || '-' }}</td>
                  <td class="py-3 pl-4 text-right font-black text-blue-600">{{ formatCurrency(item.total_omzet) }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="flex flex-col items-center justify-center py-10 text-slate-400">
              <Truck class="h-8 w-8 mb-2 opacity-50" />
              <p class="text-sm font-medium">Belum ada jadwal pengiriman TO.</p>
            </div>
          </div>
        </div>

        <!-- Table Motoris -->
        <div class="bg-white/70 backdrop-blur-xl rounded-[2rem] border border-slate-100/50 shadow-sm overflow-hidden flex flex-col h-full">
          <div class="p-6 border-b border-slate-100 flex items-center gap-3 bg-white/50">
            <User class="h-5 w-5 text-purple-500" />
            <h2 class="text-lg font-black text-slate-800">Rincian Sales Motoris</h2>
          </div>
          <div class="p-6 flex-1 overflow-x-auto">
            <table v-if="data.motoris_sales && data.motoris_sales.length > 0" class="w-full text-left text-sm whitespace-nowrap">
              <thead>
                <tr class="text-slate-400 font-bold uppercase tracking-wider text-xs border-b border-slate-100">
                  <th class="pb-3 pr-4">Nama Sales</th>
                  <th class="pb-3 pl-4 text-right">Omzet</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50">
                <tr v-for="(item, i) in data.motoris_sales" :key="i" class="hover:bg-slate-50/50 transition-colors">
                  <td class="py-3 pr-4 font-bold text-slate-700">{{ item.salesman_name }}</td>
                  <td class="py-3 pl-4 text-right font-black text-purple-600">{{ formatCurrency(item.total_omzet) }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="flex flex-col items-center justify-center py-10 text-slate-400">
              <User class="h-8 w-8 mb-2 opacity-50" />
              <p class="text-sm font-medium">Belum ada proyeksi dari Sales Motoris.</p>
            </div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>
