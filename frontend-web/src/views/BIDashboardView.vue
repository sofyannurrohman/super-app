<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent, MarkLineComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import {
  Users,
  TrendingUp,
  Wallet,
  Clock,
  RefreshCw,
  BarChart3,
  Trophy,
  Activity
} from 'lucide-vue-next'

use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent, MarkLineComponent])

// TypeScript interfaces
interface HRSummary {
  total_employees: number;
  attendance_rate: number;
  pending_leave: number;
  pending_overtime: number;
  pending_reimburse: number;
}
interface SalesSummary {
  total_revenue: number;
  total_transactions: number;
  avg_order_value: number;
}
interface PayrollSummary {
  total_payout: number;
  total_employees_paid: number;
  last_run_date: string;
}
interface MonthlyTrend {
  label: string;
  value: number;
}
interface SalespersonRank {
  rank: number;
  name: string;
  amount: number;
}
interface RevenueExpense {
  label: string;
  revenue: number;
  expense: number;
  profit: number;
}

const currentYear = new Date().getFullYear()
const currentMonth = new Date().getMonth() + 1

const selectedMonth = ref(currentMonth)
const selectedYear = ref(currentYear)
const trendRange = ref(6)
const isLoading = ref(false)

const hrData = ref<HRSummary | null>(null)
const salesData = ref<SalesSummary | null>(null)
const payrollData = ref<PayrollSummary | null>(null)
const salesTrendData = ref<MonthlyTrend[]>([])
const attendanceTrendData = ref<MonthlyTrend[]>([])
const topSalespersonData = ref<SalespersonRank[]>([])
const revenueExpenseData = ref<RevenueExpense[]>([])

const formatMoney = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(val)
}

const fetchData = async () => {
  isLoading.value = true
  try {
    const [hrRes, salesRes, payrollRes, salesTrendRes, attTrendRes, topSalesRes, revExpRes] = await Promise.allSettled([
      apiClient.get('/api/v1/bi/summary/hr'),
      apiClient.get(`/api/v1/bi/summary/sales?month=${selectedMonth.value}&year=${selectedYear.value}`),
      apiClient.get(`/api/v1/bi/summary/payroll?month=${selectedMonth.value}&year=${selectedYear.value}`),
      apiClient.get(`/api/v1/bi/trend/sales?months=${trendRange.value}`),
      apiClient.get(`/api/v1/bi/trend/attendance?months=${trendRange.value}`),
      apiClient.get(`/api/v1/bi/top-salesperson?month=${selectedMonth.value}&year=${selectedYear.value}&limit=5`),
      apiClient.get(`/api/v1/bi/revenue-expense?months=${trendRange.value}`),
    ])

    if (hrRes.status === 'fulfilled') hrData.value = hrRes.value.data?.data || null
    if (salesRes.status === 'fulfilled') salesData.value = salesRes.value.data?.data || null
    if (payrollRes.status === 'fulfilled') payrollData.value = payrollRes.value.data?.data || null
    if (salesTrendRes.status === 'fulfilled') salesTrendData.value = salesTrendRes.value.data?.data || []
    if (attTrendRes.status === 'fulfilled') attendanceTrendData.value = attTrendRes.value.data?.data || []
    if (topSalesRes.status === 'fulfilled') topSalespersonData.value = topSalesRes.value.data?.data || []
    if (revExpRes.status === 'fulfilled') revenueExpenseData.value = revExpRes.value.data?.data || []

  } catch (error) {
    toast.error('Gagal memuat data BI Dashboard')
  } finally {
    isLoading.value = false
  }
}

watch([selectedMonth, selectedYear, trendRange], fetchData)

onMounted(() => {
  fetchData()
})

// ECharts Options Computed Properties
const salesTrendOption = computed(() => {
  if (!salesTrendData.value || salesTrendData.value.length === 0) return {}
  return {
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => `${params[0].name}<br/>${formatMoney(params[0].value)}`
    },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: salesTrendData.value.map(d => d.label) },
    yAxis: { type: 'value', axisLabel: { formatter: (value: number) => `Rp ${value / 1000000}M` } },
    series: [
      {
        name: 'Omzet',
        type: 'line',
        smooth: true,
        data: salesTrendData.value.map(d => d.value),
        areaStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [{ offset: 0, color: 'rgba(59, 130, 246, 0.5)' }, { offset: 1, color: 'rgba(59, 130, 246, 0.05)' }]
          }
        },
        itemStyle: { color: '#3B82F6' }
      }
    ]
  }
})

const topSalesOption = computed(() => {
  if (!topSalespersonData.value || topSalespersonData.value.length === 0) return {}
  const sorted = [...topSalespersonData.value].reverse() // Echarts bottom-up
  return {
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => `${params[0].name}<br/>${formatMoney(params[0].value)}`
    },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'value', axisLabel: { formatter: (value: number) => `${value / 1000000}M` } },
    yAxis: { type: 'category', data: sorted.map(d => d.name) },
    series: [
      {
        type: 'bar',
        data: sorted.map((d, index) => {
          const colors = ['#10B981', '#3B82F6', '#F59E0B', '#8B5CF6', '#EF4444']
          return { value: d.amount, itemStyle: { color: colors[index % colors.length] } }
        })
      }
    ]
  }
})

const attendanceTrendOption = computed(() => {
  if (!attendanceTrendData.value || attendanceTrendData.value.length === 0) return {}
  return {
    tooltip: { trigger: 'axis', formatter: '{b}<br/>{c}%' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: attendanceTrendData.value.map(d => d.label) },
    yAxis: { type: 'value', max: 100 },
    series: [
      {
        type: 'bar',
        data: attendanceTrendData.value.map(d => {
          let color = '#EF4444' // red
          if (d.value >= 95) color = '#10B981' // green
          else if (d.value >= 85) color = '#F59E0B' // amber
          return { value: d.value, itemStyle: { color } }
        }),
        markLine: {
          data: [{ yAxis: 95, name: 'Target (95%)' }],
          lineStyle: { color: '#8B5CF6', type: 'dashed' }
        }
      }
    ]
  }
})

const revExpOption = computed(() => {
  if (!revenueExpenseData.value || revenueExpenseData.value.length === 0) return {}
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['Revenue', 'Expense', 'Profit'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: revenueExpenseData.value.map(d => d.label) },
    yAxis: { type: 'value', axisLabel: { formatter: (value: number) => `Rp ${value / 1000000}M` } },
    series: [
      { name: 'Revenue', type: 'bar', data: revenueExpenseData.value.map(d => d.revenue), itemStyle: { color: '#3B82F6' } },
      { name: 'Expense', type: 'bar', data: revenueExpenseData.value.map(d => d.expense), itemStyle: { color: '#EF4444' } },
      { name: 'Profit', type: 'bar', data: revenueExpenseData.value.map(d => d.profit), itemStyle: { color: '#10B981' } }
    ]
  }
})
</script>

<template>
  <div class="flex flex-col h-full bg-gray-50 min-h-screen">
    <!-- Header & Filters -->
    <div class="bg-gradient-to-r from-slate-900 to-blue-900 text-white p-6 pb-10">
      <div class="flex flex-col md:flex-row md:items-center justify-between mb-4">
        <div>
          <h1 class="text-3xl font-bold flex items-center gap-2">
            <BarChart3 class="w-8 h-8" /> BI Dashboard
          </h1>
          <p class="text-blue-200 mt-1">Data operasional real-time</p>
        </div>
        <div class="flex flex-wrap gap-3 mt-4 md:mt-0">
          <select v-model="selectedMonth" class="bg-white/10 border border-white/20 rounded px-3 py-2 text-white">
            <option v-for="m in 12" :key="m" :value="m" class="text-slate-900">Bulan {{ m }}</option>
          </select>
          <select v-model="selectedYear" class="bg-white/10 border border-white/20 rounded px-3 py-2 text-white">
            <option v-for="y in [2023, 2024, 2025, 2026]" :key="y" :value="y" class="text-slate-900">{{ y }}</option>
          </select>
          <select v-model="trendRange" class="bg-white/10 border border-white/20 rounded px-3 py-2 text-white">
            <option :value="3" class="text-slate-900">Tren 3 Bulan</option>
            <option :value="6" class="text-slate-900">Tren 6 Bulan</option>
            <option :value="12" class="text-slate-900">Tren 12 Bulan</option>
          </select>
          <button @click="fetchData" class="bg-blue-600 hover:bg-blue-500 rounded px-4 py-2 flex items-center gap-2 transition-colors">
            <RefreshCw class="w-4 h-4" :class="{'animate-spin': isLoading}" /> Refresh
          </button>
        </div>
      </div>
    </div>

    <!-- Main Content (overlap header) -->
    <div class="px-6 -mt-6">
      
      <!-- KPIs -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <!-- KPI 1 -->
        <div class="bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl p-5 text-white shadow-lg relative overflow-hidden">
          <Users class="absolute right-[-10%] top-[-10%] w-32 h-32 text-white/10" />
          <p class="text-blue-100 text-sm font-medium">Total Karyawan</p>
          <div v-if="isLoading" class="h-8 w-24 bg-white/20 animate-pulse rounded mt-2 mb-1"></div>
          <h3 v-else class="text-3xl font-bold mt-1">{{ hrData?.total_employees || 0 }}</h3>
          <p class="text-xs text-blue-100 mt-2">
            <span v-if="!isLoading">✅ Kehadiran hari ini: {{ hrData?.attendance_rate?.toFixed(1) || 0 }}%</span>
            <span v-else class="h-4 w-32 bg-white/20 animate-pulse rounded inline-block"></span>
          </p>
        </div>
        
        <!-- KPI 2 -->
        <div class="bg-gradient-to-br from-green-500 to-green-600 rounded-xl p-5 text-white shadow-lg relative overflow-hidden">
          <Activity class="absolute right-[-10%] top-[-10%] w-32 h-32 text-white/10" />
          <p class="text-green-100 text-sm font-medium">Omzet Bulan Ini</p>
          <div v-if="isLoading" class="h-8 w-32 bg-white/20 animate-pulse rounded mt-2 mb-1"></div>
          <h3 v-else class="text-2xl lg:text-3xl font-bold mt-1">{{ formatMoney(salesData?.total_revenue || 0) }}</h3>
          <p class="text-xs text-green-100 mt-2">
            <span v-if="!isLoading">{{ salesData?.total_transactions || 0 }} transaksi</span>
            <span v-else class="h-4 w-24 bg-white/20 animate-pulse rounded inline-block"></span>
          </p>
        </div>

        <!-- KPI 3 -->
        <div class="bg-gradient-to-br from-amber-500 to-amber-600 rounded-xl p-5 text-white shadow-lg relative overflow-hidden">
          <Wallet class="absolute right-[-10%] top-[-10%] w-32 h-32 text-white/10" />
          <p class="text-amber-100 text-sm font-medium">Total Gaji Bulan Ini</p>
          <div v-if="isLoading" class="h-8 w-32 bg-white/20 animate-pulse rounded mt-2 mb-1"></div>
          <h3 v-else class="text-2xl lg:text-3xl font-bold mt-1">{{ formatMoney(payrollData?.total_payout || 0) }}</h3>
          <p class="text-xs text-amber-100 mt-2">
            <span v-if="!isLoading">{{ payrollData?.total_employees_paid || 0 }} karyawan</span>
            <span v-else class="h-4 w-24 bg-white/20 animate-pulse rounded inline-block"></span>
          </p>
        </div>

        <!-- KPI 4 -->
        <div class="bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl p-5 text-white shadow-lg relative overflow-hidden">
          <Clock class="absolute right-[-10%] top-[-10%] w-32 h-32 text-white/10" />
          <p class="text-purple-100 text-sm font-medium">Pending Approval</p>
          <div v-if="isLoading" class="h-8 w-16 bg-white/20 animate-pulse rounded mt-2 mb-1"></div>
          <h3 v-else class="text-3xl font-bold mt-1">
            {{ (hrData?.pending_leave || 0) + (hrData?.pending_overtime || 0) + (hrData?.pending_reimburse || 0) }}
          </h3>
          <p class="text-xs text-purple-100 mt-2 truncate" :title="`Cuti: ${hrData?.pending_leave || 0} | Lembur: ${hrData?.pending_overtime || 0} | Reimb: ${hrData?.pending_reimburse || 0}`">
            <span v-if="!isLoading">Cuti: {{ hrData?.pending_leave || 0 }} | Lembur: {{ hrData?.pending_overtime || 0 }} | Reimb: {{ hrData?.pending_reimburse || 0 }}</span>
            <span v-else class="h-4 w-40 bg-white/20 animate-pulse rounded inline-block"></span>
          </p>
        </div>
      </div>

      <!-- Charts Row 1 -->
      <div class="grid grid-cols-1 lg:grid-cols-5 gap-6 mb-6">
        <div class="bg-white rounded-xl shadow-sm p-5 lg:col-span-3 border border-slate-100">
          <h4 class="font-semibold text-slate-700 flex items-center gap-2 mb-4">
            <TrendingUp class="w-5 h-5 text-blue-500" /> Tren Penjualan
          </h4>
          <div v-if="isLoading" class="h-[300px] flex items-center justify-center bg-slate-50 animate-pulse rounded">
            <span class="text-slate-400">Loading chart...</span>
          </div>
          <div v-else-if="!salesTrendData.length" class="h-[300px] flex items-center justify-center text-slate-400">
            Belum ada data
          </div>
          <v-chart v-else class="h-[300px] w-full" :option="salesTrendOption" autoresize />
        </div>

        <div class="bg-white rounded-xl shadow-sm p-5 lg:col-span-2 border border-slate-100">
          <h4 class="font-semibold text-slate-700 flex items-center gap-2 mb-4">
            <Trophy class="w-5 h-5 text-amber-500" /> Top 5 Salesperson
          </h4>
          <div v-if="isLoading" class="h-[300px] flex items-center justify-center bg-slate-50 animate-pulse rounded">
            <span class="text-slate-400">Loading chart...</span>
          </div>
          <div v-else-if="!topSalespersonData.length" class="h-[300px] flex items-center justify-center text-slate-400">
            Belum ada data
          </div>
          <v-chart v-else class="h-[300px] w-full" :option="topSalesOption" autoresize />
        </div>
      </div>

      <!-- Charts Row 2 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 pb-10">
        <div class="bg-white rounded-xl shadow-sm p-5 border border-slate-100">
          <h4 class="font-semibold text-slate-700 flex items-center gap-2 mb-4">
            <BarChart3 class="w-5 h-5 text-purple-500" /> Tren Kehadiran
          </h4>
          <div v-if="isLoading" class="h-[300px] flex items-center justify-center bg-slate-50 animate-pulse rounded">
            <span class="text-slate-400">Loading chart...</span>
          </div>
          <div v-else-if="!attendanceTrendData.length" class="h-[300px] flex items-center justify-center text-slate-400">
            Belum ada data
          </div>
          <v-chart v-else class="h-[300px] w-full" :option="attendanceTrendOption" autoresize />
        </div>

        <div class="bg-white rounded-xl shadow-sm p-5 border border-slate-100">
          <h4 class="font-semibold text-slate-700 flex items-center gap-2 mb-4">
            <Wallet class="w-5 h-5 text-green-500" /> Pendapatan vs Pengeluaran
          </h4>
          <div v-if="isLoading" class="h-[300px] flex items-center justify-center bg-slate-50 animate-pulse rounded">
            <span class="text-slate-400">Loading chart...</span>
          </div>
          <div v-else-if="!revenueExpenseData.length" class="h-[300px] flex items-center justify-center text-slate-400">
            Belum ada data
          </div>
          <v-chart v-else class="h-[300px] w-full" :option="revExpOption" autoresize />
        </div>
      </div>
      
    </div>
  </div>
</template>
