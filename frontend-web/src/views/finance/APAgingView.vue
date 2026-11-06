<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import {
  TrendingDown,
  RefreshCw,
  Download,
  Building2
} from 'lucide-vue-next'
import { useMasterDataStore } from '@/stores/masterData'

// State
const isLoading = ref(false)
const asOfDate = ref(new Date().toISOString().split('T')[0])
const apData = ref<any[]>([])
const summaryData = ref<any>(null)
const masterDataStore = useMasterDataStore()

const fetchAPAging = async () => {
  isLoading.value = true
  try {
    const companyId = masterDataStore.selectedBranchCompanyId || ''
    const res = await apiClient.get(
      `/finance/ap-aging?as_of=${asOfDate.value}&company_id=${companyId}`,
    )
    if (res.data?.data) {
      apData.value = res.data.data.buckets || []
      summaryData.value = res.data.data.summary || null
    }
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal memuat data AP Aging')
  } finally {
    isLoading.value = false
  }
}

import { watch } from 'vue'
watch(
  () => masterDataStore.selectedBranchId,
  () => {
    fetchAPAging()
  },
)

onMounted(() => {
  fetchAPAging()
})

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const exportCSV = () => {
  if (apData.value.length === 0) {
    toast.error('Tidak ada data untuk diekspor')
    return
  }

  const headers = ['Nama Supplier', '0-30 Hari', '31-60 Hari', '61-90 Hari', '>90 Hari', 'Total']
  const rows = apData.value.map(row => [
    row.entity_name,
    row.current || 0,
    row.days_31_60 || 0,
    row.days_61_90 || 0,
    row.over_90 || 0,
    row.total || 0
  ])
  
  if (summaryData.value) {
    rows.push([
      'TOTAL',
      summaryData.value.current || 0,
      summaryData.value.days_31_60 || 0,
      summaryData.value.days_61_90 || 0,
      summaryData.value.over_90 || 0,
      summaryData.value.total || 0
    ])
  }

  const csvContent = [
    headers.join(','),
    ...rows.map(r => r.map(cell => `"${cell}"`).join(','))
  ].join('\n')

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.setAttribute('href', url)
  link.setAttribute('download', `AP_Aging_${asOfDate.value}.csv`)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center gap-2">
          Laporan Hutang (AP Aging)
        </h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <TrendingDown class="w-4 h-4 text-rose-500" />
          Menampilkan berapa lama faktur supplier belum dibayar.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <input
          type="date"
          v-model="asOfDate"
          @change="fetchAPAging"
          class="bg-white border border-slate-200 px-4 py-2.5 rounded-2xl text-sm font-bold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-rose-500/20"
        />

        <button
          @click="fetchAPAging"
          class="p-2.5 bg-white border border-slate-200 text-slate-600 hover:text-slate-900 rounded-2xl transition-all shadow-sm active:scale-95"
        >
          <RefreshCw class="w-5 h-5" :class="{ 'animate-spin': isLoading }" />
        </button>

        <button
          @click="exportCSV"
          class="flex items-center gap-2 bg-slate-900 text-white hover:bg-slate-800 px-5 py-2.5 rounded-2xl text-sm font-bold transition-all shadow-md active:scale-95"
        >
          <Download class="w-4 h-4" />
          <span>Export CSV</span>
        </button>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-5 gap-6">
      <div class="bg-white p-6 rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-slate-500 mb-2 uppercase">0-30 Hari</p>
        <p class="text-xl font-black text-green-600">{{ formatCurrency(summaryData?.current || 0) }}</p>
      </div>
      <div class="bg-white p-6 rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-slate-500 mb-2 uppercase">31-60 Hari</p>
        <p class="text-xl font-black text-amber-600">{{ formatCurrency(summaryData?.days_31_60 || 0) }}</p>
      </div>
      <div class="bg-white p-6 rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-slate-500 mb-2 uppercase">61-90 Hari</p>
        <p class="text-xl font-black text-orange-600">{{ formatCurrency(summaryData?.days_61_90 || 0) }}</p>
      </div>
      <div class="bg-white p-6 rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-slate-500 mb-2 uppercase">>90 Hari</p>
        <p class="text-xl font-black text-red-600">{{ formatCurrency(summaryData?.over_90 || 0) }}</p>
      </div>
      <div class="bg-rose-50/50 p-6 rounded-3xl border border-rose-100 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-rose-600 mb-2 uppercase">Total Hutang</p>
        <p class="text-xl font-black text-rose-700">{{ formatCurrency(summaryData?.total || 0) }}</p>
      </div>
    </div>

    <!-- Data Table -->
    <div v-if="isLoading" class="space-y-4">
      <div v-for="i in 3" :key="i" class="h-16 bg-slate-100 animate-pulse rounded-2xl w-full"></div>
    </div>

    <div v-else-if="apData.length === 0" class="bg-white rounded-3xl border border-slate-200/60 p-16 text-center text-slate-400 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
      <Building2 class="w-16 h-16 mx-auto mb-4 text-slate-300 opacity-50" />
      <h3 class="text-lg font-bold text-slate-700">Belum ada hutang outstanding</h3>
      <p class="text-sm text-slate-500 mt-1">
        Tidak ada data hutang untuk periode ini.
      </p>
    </div>

    <div v-else class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_25px_rgba(0,0,0,0.03)] overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse min-w-[800px]">
          <thead>
            <tr class="bg-slate-900 text-white uppercase text-[10px] font-black tracking-[0.15em]">
              <th class="px-6 py-5.5 border-r border-slate-800 w-16 text-center">No</th>
              <th class="px-6 py-5.5 border-r border-slate-800">Nama Supplier</th>
              <th class="px-6 py-5.5 text-right border-r border-slate-800">0-30hr</th>
              <th class="px-6 py-5.5 text-right border-r border-slate-800">31-60hr</th>
              <th class="px-6 py-5.5 text-right border-r border-slate-800">61-90hr</th>
              <th class="px-6 py-5.5 text-right border-r border-slate-800">>90hr</th>
              <th class="px-6 py-5.5 text-right">Total</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700 font-medium">
            <tr v-for="(row, idx) in apData" :key="row.entity_id" class="hover:bg-slate-50">
              <td class="px-6 py-4.5 text-center text-slate-400 border-r border-slate-100">{{ idx + 1 }}</td>
              <td class="px-6 py-4.5 font-bold text-slate-800 border-r border-slate-100">{{ row.entity_name }}</td>
              <td class="px-6 py-4.5 text-right text-green-600 border-r border-slate-100">{{ formatCurrency(row.current) }}</td>
              <td class="px-6 py-4.5 text-right text-amber-600 border-r border-slate-100">{{ formatCurrency(row.days_31_60) }}</td>
              <td class="px-6 py-4.5 text-right text-orange-600 border-r border-slate-100">{{ formatCurrency(row.days_61_90) }}</td>
              <td class="px-6 py-4.5 text-right text-red-600 font-bold border-r border-slate-100">{{ formatCurrency(row.over_90) }}</td>
              <td class="px-6 py-4.5 text-right font-black text-slate-900">{{ formatCurrency(row.total) }}</td>
            </tr>
          </tbody>
          <tfoot v-if="summaryData" class="bg-slate-50 border-t-2 border-slate-200">
            <tr>
              <td colspan="2" class="px-6 py-5 font-black text-slate-900 border-r border-slate-200">GRAND TOTAL</td>
              <td class="px-6 py-5 text-right font-black text-green-700 border-r border-slate-200">{{ formatCurrency(summaryData.current) }}</td>
              <td class="px-6 py-5 text-right font-black text-amber-700 border-r border-slate-200">{{ formatCurrency(summaryData.days_31_60) }}</td>
              <td class="px-6 py-5 text-right font-black text-orange-700 border-r border-slate-200">{{ formatCurrency(summaryData.days_61_90) }}</td>
              <td class="px-6 py-5 text-right font-black text-red-700 border-r border-slate-200">{{ formatCurrency(summaryData.over_90) }}</td>
              <td class="px-6 py-5 text-right font-black text-rose-700">{{ formatCurrency(summaryData.total) }}</td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  </div>
</template>
