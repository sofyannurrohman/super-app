<script setup lang="ts">
import { ref, computed } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import {
  Sparkles,
  Search,
  Bot,
  Database,
  BarChart3,
  LineChartIcon,
  PieChartIcon,
  TableProperties,
  ArrowRight,
  Terminal,
  ChevronDown,
  ChevronUp,
  RotateCcw,
  AlertCircle
} from 'lucide-vue-next'

// Register ECharts modules
use([
  CanvasRenderer,
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
])

interface QueryResponse {
  sql: string
  chart_type: 'bar' | 'line' | 'pie' | 'table'
  summary: string
  columns: string[]
  data: Record<string, any>[]
}

const queryInput = ref('')
const isLoading = ref(false)
const showSql = ref(false)
const result = ref<QueryResponse | null>(null)

// Preconfigured suggestions
const suggestions = [
  {
    title: 'Keterlambatan Absensi',
    desc: 'Daftar karyawan yang paling sering terlambat bulan ini',
    prompt: 'Tampilkan nama karyawan beserta jumlah keterlambatannya di bulan ini, urutkan dari yang terbanyak',
  },
  {
    title: 'Kunjungan & Revenue Sales',
    desc: 'Perbandingan total kunjungan toko dengan total transaksi sales',
    prompt: 'Bandingkan total kunjungan toko sales_visits dengan total_amount dari sales_transactions untuk setiap rayon toko',
  },
  {
    title: 'Statistik Cuti Karyawan',
    desc: 'Persentase status persetujuan pengajuan cuti',
    prompt: 'Hitung persentase jumlah pengajuan cuti berdasarkan statusnya (APPROVED, REJECTED, PENDING)',
  },
  {
    title: 'Manajemen Kendaraan',
    desc: 'Daftar kendaraan kantor dan statusnya',
    prompt: 'Tampilkan daftar kendaraan (plat nomor, merk, jenis) beserta status terbarunya',
  },
]

const selectSuggestion = (prompt: string) => {
  queryInput.value = prompt
  handleSearch()
}

const handleSearch = async () => {
  if (!queryInput.value.trim()) {
    toast.error('Masukkan pertanyaan Anda terlebih dahulu')
    return
  }

  isLoading.value = true
  result.value = null
  showSql.value = false

  try {
    const res = await apiClient.post('/analytics/query', {
      question: queryInput.value,
    })
    result.value = res.data
    toast.success('Analisis selesai!')
  } catch (error: any) {
    console.error('Failed to run analytics query:', error)
    const errorMsg = error.response?.data?.error || error.message || 'Terjadi kesalahan sistem'
    toast.error(`Gagal menganalisis: ${errorMsg}`)
  } finally {
    isLoading.value = false
  }
}

const handleReset = () => {
  queryInput.value = ''
  result.value = null
  showSql.value = false
}

// Compute ECharts option based on API response
const chartOption = computed(() => {
  if (!result.value || !result.value.data || result.value.data.length === 0) return {}

  const data = result.value.data
  const cols = result.value.columns

  // We assume the first column is the label/category, and other numerical columns are series
  const labelKey = cols[0] || ''
  const categories = data.map((row) => String(row[labelKey] || 'N/A'))

  if (result.value.chart_type === 'pie') {
    // For pie charts, find a numeric key for values
    const valueKey = cols.find((c) => c !== labelKey) || cols[1] || ''
    const pieData = data.map((row) => ({
      name: String(row[labelKey] || 'N/A'),
      value: Number(row[valueKey] || 0),
    }))

    return {
      title: {
        text: result.value.summary,
        left: 'center',
        textStyle: { fontFamily: 'Outfit, sans-serif', color: '#1e293b', fontSize: 16 },
      },
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b} : {c} ({d}%)',
      },
      legend: {
        orient: 'horizontal',
        bottom: 'bottom',
      },
      series: [
        {
          name: labelKey,
          type: 'pie',
          radius: '55%',
          center: ['50%', '50%'],
          data: pieData,
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
          },
        },
      ],
    }
  }

  // Bar or Line chart configuration
  const numericKeys = cols.filter((c) => c !== labelKey)
  const seriesList = numericKeys.map((key) => ({
    name: key.replace(/_/g, ' ').toUpperCase(),
    type: result.value?.chart_type === 'line' ? 'line' : 'bar',
    data: data.map((row) => Number(row[key] || 0)),
    smooth: true,
    itemStyle: {
      borderRadius: result.value?.chart_type === 'line' ? 0 : [6, 6, 0, 0],
    },
  }))

  return {
    title: {
      text: result.value.summary,
      left: 'center',
      textStyle: { fontFamily: 'Outfit, sans-serif', color: '#1e293b', fontSize: 16 },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
    },
    legend: {
      bottom: 'bottom',
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: categories,
      axisLabel: { interval: 0, rotate: categories.length > 5 ? 30 : 0 },
    },
    yAxis: {
      type: 'value',
    },
    series: seriesList,
  }
})
</script>

<template>
  <div class="p-6 max-w-7xl mx-auto space-y-8 min-h-screen">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-black text-slate-900 tracking-tight flex items-center gap-2.5">
          <Sparkles class="w-8 h-8 text-primary animate-pulse" />
          AI Analytics Hub
        </h1>
        <p class="text-sm font-bold text-slate-500 mt-1 uppercase tracking-wider">
          Natural Language to SQL Database Assistant
        </p>
      </div>
    </div>

    <!-- Main Prompt Input -->
    <div class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_10px_30px_rgba(0,0,0,0.02)] p-6 relative overflow-hidden">
      <!-- Decorative Gradient Corner -->
      <div class="absolute -top-20 -right-20 w-48 h-48 bg-primary/5 rounded-full blur-3xl pointer-events-none"></div>

      <div class="flex items-center gap-3 mb-4">
        <Bot class="w-6 h-6 text-primary" />
        <span class="font-bold text-slate-700">Tanyakan database Karsa HR menggunakan bahasa natural:</span>
      </div>

      <form @submit.prevent="handleSearch" class="flex flex-col md:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-4.5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
          <input
            v-model="queryInput"
            type="text"
            placeholder="Contoh: Berapa jumlah salesman di cabang Surabaya yang aktif?"
            class="w-full pl-12 pr-4 py-4 bg-slate-50 border border-slate-200 rounded-2xl font-semibold text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all text-sm sm:text-base"
            :disabled="isLoading"
          />
        </div>
        <div class="flex gap-2">
          <button
            v-if="queryInput || result"
            type="button"
            @click="handleReset"
            class="px-4 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-2xl font-bold transition-all flex items-center justify-center gap-2 active:scale-95"
          >
            <RotateCcw class="w-5 h-5" />
          </button>
          <button
            type="submit"
            :disabled="isLoading || !queryInput.trim()"
            class="px-6 py-4 bg-slate-900 hover:bg-primary text-white rounded-2xl font-black text-sm uppercase tracking-wider transition-all duration-300 shadow-lg hover:shadow-primary/20 flex items-center justify-center gap-2 disabled:opacity-50 disabled:pointer-events-none active:scale-[0.98]"
          >
            <span v-if="isLoading">Menganalisis...</span>
            <span v-else class="flex items-center gap-2">
              Kirim
              <ArrowRight class="w-4 h-4" />
            </span>
          </button>
        </div>
      </form>
    </div>

    <!-- Suggested Queries (Only shown when no results yet) -->
    <div v-if="!result && !isLoading" class="space-y-4">
      <h2 class="text-xs font-black text-slate-400 uppercase tracking-widest">
        Pertanyaan Populer / Ide Eksplorasi
      </h2>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="item in suggestions"
          :key="item.title"
          @click="selectSuggestion(item.prompt)"
          class="group p-5 bg-white border border-slate-200/60 rounded-2xl cursor-pointer hover:border-primary/50 hover:shadow-lg hover:shadow-primary/2 transition-all duration-300 relative flex flex-col justify-between"
        >
          <div>
            <h3 class="font-bold text-slate-800 group-hover:text-primary transition-colors">
              {{ item.title }}
            </h3>
            <p class="text-xs font-semibold text-slate-400 mt-1">
              {{ item.desc }}
            </p>
          </div>
          <div class="flex items-center justify-between mt-4">
            <span class="text-xs font-mono bg-slate-50 text-slate-600 group-hover:bg-primary/5 group-hover:text-primary px-3 py-1.5 rounded-lg truncate max-w-[85%] font-medium">
              "{{ item.prompt }}"
            </span>
            <ArrowRight class="w-4 h-4 text-slate-300 group-hover:text-primary transition-colors group-hover:translate-x-1 duration-300" />
          </div>
        </div>
      </div>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="isLoading" class="flex flex-col items-center justify-center py-20 space-y-4 bg-white border border-slate-200/60 rounded-3xl shadow-sm">
      <div class="relative flex items-center justify-center">
        <div class="w-16 h-16 border-4 border-slate-100 border-t-primary rounded-full animate-spin"></div>
        <Bot class="absolute w-6 h-6 text-primary animate-bounce" />
      </div>
      <div class="text-center">
        <p class="font-black text-slate-800">Menghubungi AI Model...</p>
        <p class="text-xs font-bold text-slate-400 mt-1 uppercase tracking-widest animate-pulse">
          Menerjemahkan ke SQL & Menghitung Data
        </p>
      </div>
    </div>

    <!-- Results Panel -->
    <div v-if="result" class="space-y-6 animate-in fade-in slide-in-from-bottom-5 duration-500">
      <!-- AI Interpretation / Summary -->
      <div class="p-6 bg-primary/5 border border-primary/10 rounded-3xl flex items-start gap-4">
        <div class="p-2.5 bg-primary/10 rounded-2xl text-primary shrink-0">
          <Bot class="w-6 h-6" />
        </div>
        <div>
          <h3 class="font-black text-slate-900 leading-none">Analisis AI</h3>
          <p class="text-slate-700 font-semibold mt-2.5 leading-relaxed text-sm md:text-base">
            {{ result.summary }}
          </p>
        </div>
      </div>

      <!-- Chart View -->
      <div v-if="result.chart_type !== 'table' && result.data && result.data.length > 0" class="p-6 bg-white border border-slate-200/60 rounded-3xl shadow-sm">
        <div class="flex items-center justify-between mb-6 pb-4 border-b border-slate-100">
          <div class="flex items-center gap-2">
            <BarChart3 v-if="result.chart_type === 'bar'" class="w-5 h-5 text-primary" />
            <LineChartIcon v-else-if="result.chart_type === 'line'" class="w-5 h-5 text-primary" />
            <PieChartIcon v-else class="w-5 h-5 text-primary" />
            <span class="font-bold text-slate-800 uppercase text-xs tracking-wider">Visualisasi Grafik</span>
          </div>
        </div>
        <div class="h-[400px] w-full">
          <VChart :option="chartOption" class="w-full h-full" autoresize />
        </div>
      </div>

      <!-- Raw Data / Table View -->
      <div class="bg-white border border-slate-200/60 rounded-3xl shadow-sm overflow-hidden">
        <div class="p-6 flex items-center justify-between border-b border-slate-100">
          <div class="flex items-center gap-2">
            <TableProperties class="w-5 h-5 text-slate-800" />
            <span class="font-bold text-slate-800 uppercase text-xs tracking-wider">Rincian Data ({{ result.data ? result.data.length : 0 }} baris)</span>
          </div>
        </div>

        <div v-if="!result.data || result.data.length === 0" class="p-12 text-center flex flex-col items-center">
          <AlertCircle class="w-12 h-12 text-slate-300" />
          <p class="font-black text-slate-500 mt-2">Tidak Ada Data Ditemukan</p>
          <p class="text-xs font-semibold text-slate-400 mt-1">Kueri berhasil dijalankan namun tidak mengembalikan baris.</p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-100 text-sm">
            <thead class="bg-slate-50">
              <tr>
                <th
                  v-for="col in result.columns"
                  :key="col"
                  class="px-6 py-4 text-left text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap"
                >
                  {{ col.replace(/_/g, ' ') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 font-semibold text-slate-700">
              <tr
                v-for="(row, idx) in result.data"
                :key="idx"
                class="hover:bg-slate-50/50 transition-colors"
              >
                <td
                  v-for="col in result.columns"
                  :key="col"
                  class="px-6 py-4 whitespace-nowrap"
                >
                  {{ row[col] !== null && row[col] !== undefined ? row[col] : '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- SQL Developer Details -->
      <div class="bg-white border border-slate-200/60 rounded-3xl overflow-hidden shadow-sm">
        <button
          @click="showSql = !showSql"
          class="w-full p-6 flex items-center justify-between text-left hover:bg-slate-50 transition-colors focus:outline-none"
        >
          <div class="flex items-center gap-2">
            <Terminal class="w-5 h-5 text-slate-500" />
            <span class="font-bold text-slate-700 text-xs uppercase tracking-wider">Developer Log & SQL Query</span>
          </div>
          <ChevronDown v-if="!showSql" class="w-5 h-5 text-slate-400" />
          <ChevronUp v-else class="w-5 h-5 text-slate-400" />
        </button>

        <div v-if="showSql" class="p-6 border-t border-slate-100 bg-slate-900 text-slate-200 font-mono text-xs overflow-x-auto leading-relaxed">
          <div class="flex items-center justify-between text-[10px] text-slate-400 font-bold uppercase tracking-widest pb-3 mb-3 border-b border-slate-800">
            <span>Query Dialek: PostgreSQL</span>
            <span class="text-green-500 font-black">Success Log</span>
          </div>
          <pre class="whitespace-pre-wrap">{{ result.sql }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.analytics-card {
  font-family: 'Outfit', 'Inter', sans-serif;
}
</style>
