<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, provide } from 'vue'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import { RouterLink } from 'vue-router'
import {
  ShoppingBag,
  Package,
  AlertTriangle,
  TrendingUp,
  TrendingDown,
  BarChart3,
  RefreshCw,
  Clock,
  ArrowRight,
  Boxes,
  Calendar,
} from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import 'echarts'
import VChart, { THEME_KEY } from 'vue-echarts'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()

provide(THEME_KEY, 'light')

const currentDate = computed(() => {
  return new Date().toLocaleDateString('id-ID', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
})

const currentBranchName = computed(() => {
  if (masterDataStore.selectedBranchId === 'ALL_BRANCHES') return 'Semua Cabang'
  const branch = masterDataStore.branches.find(b => b.id === masterDataStore.selectedBranchId)
  return branch ? branch.name : 'Semua Cabang'
})
const isRefreshing = ref(false)
function getTodayString() {
  const d = new Date()
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const period = ref('daily')
const selectedDate = ref(getTodayString())

let pollingInterval: ReturnType<typeof setInterval> | null = null

async function loadAll() {
  await Promise.all([marketplaceStore.fetchDashboard(period.value, selectedDate.value), marketplaceStore.fetchStocks()])
}

async function manualRefresh() {
  isRefreshing.value = true
  await loadAll()
  setTimeout(() => (isRefreshing.value = false), 600)
}

onMounted(async () => {
  await loadAll()
  // Realtime polling every 15 seconds
  pollingInterval = setInterval(() => {
    loadAll()
  }, 15000)
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
})

watch(() => masterDataStore.selectedBranchId, () => {
  loadAll()
})


const topMovingStocks = computed(() =>
  [...marketplaceStore.stocks]
    .sort((a, b) => b.quantity - a.quantity)
    .slice(0, 5),
)

const chartOptions = computed(() => {
  const data = marketplaceStore.dashboardStats.chart_data || []
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { 
      type: 'category', 
      boundaryGap: false, 
      data: data.map((d: any) => d.date) 
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: 'Omzet',
        type: 'line',
        data: data.map((d: any) => d.omzet),
        smooth: true,
        areaStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(59, 130, 246, 0.5)' },
              { offset: 1, color: 'rgba(59, 130, 246, 0.05)' }
            ]
          }
        },
        itemStyle: { color: '#3b82f6' }
      }
    ]
  }
})
</script>

<template>
  <div class="space-y-8 animate-in fade-in duration-700">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-3xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2.5 rounded-2xl bg-gradient-to-br from-primary to-primary/80 shadow-lg shadow-primary/20">
            <ShoppingBag class="h-7 w-7 text-white" />
          </div>
          Dashboard Marketplace
        </h1>
        <p class="text-slate-500 font-medium mt-1">
          Pantau stok marketplace secara realtime — {{ currentBranchName }}
        </p>
      </div>
      <div class="flex items-center gap-3">
        <!-- Live badge -->
        <div class="flex items-center gap-2 bg-emerald-50 border border-emerald-200 px-3 py-2 rounded-xl">
          <span class="relative flex h-2.5 w-2.5">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
          </span>
          <span class="text-xs font-black text-emerald-700 uppercase tracking-widest">LIVE</span>
        </div>
        <!-- Date Picker (Only if daily is selected) -->
        <div v-if="period === 'daily'" class="flex items-center gap-2 bg-white border border-slate-200 px-3 py-1.5 rounded-xl shadow-sm focus-within:border-primary transition-colors">
          <Calendar class="h-4 w-4 text-primary" />
          <input type="date" v-model="selectedDate" @change="loadAll" class="text-xs font-bold text-slate-600 outline-none bg-transparent appearance-none cursor-pointer" />
        </div>
        <!-- Default Date Display for Monthly -->
        <div v-else class="flex items-center gap-2 bg-white border border-slate-100 px-3 py-2 rounded-xl shadow-sm">
          <Calendar class="h-4 w-4 text-primary" />
          <span class="text-xs font-bold text-slate-600">{{ currentDate }}</span>
        </div>
        <!-- Filter Period -->
        <select v-model="period" @change="loadAll" class="bg-white border border-slate-200 px-3 py-2 rounded-xl text-xs font-bold text-slate-600 shadow-sm focus:outline-none focus:border-primary">
          <option value="daily">Harian</option>
          <option value="monthly">Bulanan</option>
        </select>
        <!-- Refresh button -->
        <button
          @click="manualRefresh"
          class="flex items-center gap-2 bg-white border border-slate-200 px-3 py-2 rounded-xl text-slate-500 hover:text-primary hover:border-primary/30 transition-all text-xs font-bold shadow-sm"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': isRefreshing }" />
          Refresh
        </button>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-5">
      <!-- Total SKU -->
      <Card class="relative overflow-hidden group hover:shadow-md transition-all duration-300 border border-slate-200 bg-white shadow-sm">
        <div class="absolute -right-4 -top-4 opacity-5 group-hover:opacity-10 transition-opacity">
          <Boxes class="h-32 w-32 text-violet-600" />
        </div>
        <CardHeader class="pb-1 flex flex-row items-center justify-between">
          <div>
            <CardDescription class="text-[10px] font-bold uppercase tracking-widest text-slate-500">Total SKU Aktif</CardDescription>
            <CardTitle class="text-3xl font-black text-slate-900 mt-1">{{ marketplaceStore.dashboardStats.total_sku }}</CardTitle>
          </div>
          <div class="p-2.5 bg-violet-50 rounded-xl">
             <Boxes class="h-5 w-5 text-violet-600" />
          </div>
        </CardHeader>
        <CardContent>
          <p class="text-xs font-medium text-slate-500">Produk terdaftar di marketplace</p>
        </CardContent>
      </Card>

      <!-- Critical Stock -->
      <Card class="relative overflow-hidden group hover:shadow-md transition-all duration-300 border border-slate-200 bg-white shadow-sm">
        <div class="absolute -right-4 -top-4 opacity-5 group-hover:opacity-10 transition-opacity">
          <AlertTriangle class="h-32 w-32" :class="marketplaceStore.dashboardStats.critical_count > 0 ? 'text-red-600' : 'text-slate-400'" />
        </div>
        <CardHeader class="pb-1 flex flex-row items-center justify-between">
          <div>
            <CardDescription class="text-[10px] font-bold uppercase tracking-widest text-slate-500">Stok Kritis</CardDescription>
            <CardTitle class="text-3xl font-black mt-1" :class="marketplaceStore.dashboardStats.critical_count > 0 ? 'text-red-600' : 'text-slate-900'">
              {{ marketplaceStore.dashboardStats.critical_count }}
            </CardTitle>
          </div>
          <div class="p-2.5 rounded-xl" :class="marketplaceStore.dashboardStats.critical_count > 0 ? 'bg-red-50' : 'bg-slate-50'">
             <AlertTriangle class="h-5 w-5" :class="marketplaceStore.dashboardStats.critical_count > 0 ? 'text-red-600' : 'text-slate-400'" />
          </div>
        </CardHeader>
        <CardContent>
          <p class="text-xs font-medium" :class="marketplaceStore.dashboardStats.critical_count > 0 ? 'text-red-500' : 'text-slate-500'">
            {{ marketplaceStore.dashboardStats.critical_count > 0 ? 'Segera lakukan opname!' : 'Semua stok aman' }}
          </p>
        </CardContent>
      </Card>

      <!-- Today In -->
      <Card class="relative overflow-hidden group hover:shadow-md transition-all duration-300 border border-slate-200 bg-white shadow-sm">
        <div class="absolute -right-4 -top-4 opacity-5 group-hover:opacity-10 transition-opacity">
          <TrendingUp class="h-32 w-32 text-emerald-600" />
        </div>
        <CardHeader class="pb-1 flex flex-row items-center justify-between">
          <div>
            <CardDescription class="text-[10px] font-bold uppercase tracking-widest text-slate-500">{{ period === 'monthly' ? 'Barang Masuk Bulan Ini' : 'Barang Masuk Hari Ini' }}</CardDescription>
            <CardTitle class="text-3xl font-black text-slate-900 mt-1">{{ marketplaceStore.dashboardStats.today_in }}</CardTitle>
          </div>
          <div class="p-2.5 bg-emerald-50 rounded-xl">
             <TrendingUp class="h-5 w-5 text-emerald-600" />
          </div>
        </CardHeader>
        <CardContent>
          <p class="text-xs font-medium text-slate-500">Total transaksi masuk</p>
        </CardContent>
      </Card>

      <!-- Today Out -->
      <Card class="relative overflow-hidden group hover:shadow-md transition-all duration-300 border border-slate-200 bg-white shadow-sm">
        <div class="absolute -right-4 -top-4 opacity-5 group-hover:opacity-10 transition-opacity">
          <TrendingDown class="h-32 w-32 text-amber-600" />
        </div>
        <CardHeader class="pb-1 flex flex-row items-center justify-between">
          <div>
            <CardDescription class="text-[10px] font-bold uppercase tracking-widest text-slate-500">{{ period === 'monthly' ? 'Barang Keluar Bulan Ini' : 'Barang Keluar Hari Ini' }}</CardDescription>
            <CardTitle class="text-3xl font-black text-slate-900 mt-1">{{ marketplaceStore.dashboardStats.today_out }}</CardTitle>
          </div>
          <div class="p-2.5 bg-amber-50 rounded-xl">
             <TrendingDown class="h-5 w-5 text-amber-600" />
          </div>
        </CardHeader>
        <CardContent>
          <p class="text-xs font-medium text-slate-500">Total transaksi keluar</p>
        </CardContent>
      </Card>

      <!-- Today Omzet -->
      <Card class="relative overflow-hidden group hover:shadow-md transition-all duration-300 border border-slate-200 bg-white shadow-sm">
        <div class="absolute -right-4 -top-4 opacity-5 group-hover:opacity-10 transition-opacity">
          <ShoppingBag class="h-32 w-32 text-blue-600" />
        </div>
        <CardHeader class="pb-1 flex flex-row items-center justify-between">
          <div>
            <CardDescription class="text-[10px] font-bold uppercase tracking-widest text-slate-500">{{ period === 'monthly' ? 'Omzet Keluar Bulan Ini' : 'Omzet Keluar Hari Ini' }}</CardDescription>
            <CardTitle class="text-xl lg:text-2xl font-black text-slate-900 mt-1 truncate max-w-[120px] lg:max-w-full" :title="'Rp ' + (marketplaceStore.dashboardStats.today_omzet || 0).toLocaleString('id-ID')">
              Rp {{ (marketplaceStore.dashboardStats.today_omzet || 0).toLocaleString('id-ID') }}
            </CardTitle>
          </div>
          <div class="p-2.5 bg-blue-50 rounded-xl flex-shrink-0">
             <ShoppingBag class="h-5 w-5 text-blue-600" />
          </div>
        </CardHeader>
        <CardContent>
          <p class="text-xs font-medium text-slate-500">Berdasarkan data stok keluar</p>
        </CardContent>
      </Card>
    </div>

    <!-- Chart -->
    <div v-if="period === 'monthly'" class="grid grid-cols-1 gap-6">
      <Card class="border-none shadow-sm bg-white">
        <CardHeader>
          <CardTitle class="text-base font-black text-slate-800">Proyeksi Omzet Harian (Bulan Ini)</CardTitle>
        </CardHeader>
        <CardContent>
          <v-chart class="h-[300px] w-full" :option="chartOptions" autoresize />
        </CardContent>
      </Card>
    </div>

    <!-- Top Stocks -->
    <div class="grid grid-cols-1 gap-6">
      <div class="w-full">
        <Card class="border-none shadow-sm h-full">
          <CardHeader class="pb-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Package class="h-5 w-5 text-primary" />
                <CardTitle class="text-base font-black text-slate-800">Top 5 Produk Stok Tertinggi</CardTitle>
              </div>
              <RouterLink
                to="/marketplace/stock"
                class="text-xs font-bold text-primary hover:underline flex items-center gap-1"
              >
                Lihat Semua <ArrowRight class="h-3 w-3" />
              </RouterLink>
            </div>
          </CardHeader>
          <CardContent class="space-y-3">
            <div v-if="topMovingStocks.length === 0" class="text-center py-6 text-slate-400 text-sm">
              Belum ada produk
            </div>
            <div
              v-for="(stock, idx) in topMovingStocks"
              :key="stock.id"
              class="flex items-center gap-3 p-3 rounded-xl bg-slate-50 hover:bg-slate-100 transition-colors"
            >
              <div class="flex-shrink-0 w-8 h-8 rounded-xl bg-gradient-to-br from-primary to-primary/80 flex items-center justify-center text-white text-xs font-black shadow-sm">
                {{ idx + 1 }}
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-bold text-slate-900 truncate">{{ stock.product_name }}</p>
                <p class="text-xs text-slate-500 font-medium">{{ stock.sku }}</p>
              </div>
              <div class="text-right">
                <p class="text-base font-black" :class="marketplaceStore.getStockStatus(stock) === 'safe' ? 'text-slate-900' : marketplaceStore.getStockStatus(stock) === 'critical' ? 'text-amber-600' : 'text-red-600'">
                  {{ stock.quantity.toLocaleString('id-ID') }}
                </p>
                <p class="text-[10px] text-slate-400 font-medium">unit</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>

    <!-- Quick Links -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
      <RouterLink to="/marketplace/stock" class="group">
        <Card class="border-none shadow-sm hover:shadow-lg transition-all duration-300 cursor-pointer group-hover:border-primary/30 border">
          <CardContent class="p-5 flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-primary/10 group-hover:bg-primary transition-colors">
              <Package class="h-6 w-6 text-primary group-hover:text-white transition-colors" />
            </div>
            <div class="flex-1">
              <h3 class="font-black text-slate-900 text-base">Manajemen Stok</h3>
              <p class="text-xs text-slate-500 font-medium mt-0.5">CRUD produk & opname barang masuk/keluar</p>
            </div>
            <ArrowRight class="h-5 w-5 text-slate-300 group-hover:text-primary group-hover:translate-x-1 transition-all" />
          </CardContent>
        </Card>
      </RouterLink>
      <RouterLink to="/marketplace/audit-trail" class="group">
        <Card class="border-none shadow-sm hover:shadow-lg transition-all duration-300 cursor-pointer group-hover:border-primary/30 border">
          <CardContent class="p-5 flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-primary/10 group-hover:bg-primary transition-colors">
              <Clock class="h-6 w-6 text-primary group-hover:text-white transition-colors" />
            </div>
            <div class="flex-1">
              <h3 class="font-black text-slate-900 text-base">Riwayat Pergerakan Stok</h3>
              <p class="text-xs text-slate-500 font-medium mt-0.5">History barang masuk, keluar, dan penyesuaian stok</p>
            </div>
            <ArrowRight class="h-5 w-5 text-slate-300 group-hover:text-primary group-hover:translate-x-1 transition-all" />
          </CardContent>
        </Card>
      </RouterLink>
    </div>
  </div>
</template>
