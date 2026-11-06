<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import apiClient from '@/api/axios'
import { filterSellableProducts } from '@/utils/productUnits'
import { getAccurateStoreRegion, toTitleCase, formatCleanRegencyName } from '@/utils/indonesiaRegions'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  DataZoomComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import {
  Users,
  TrendingUp,
  Banknote,
  Megaphone,
  ArrowRight,
  Target,
  Award,
  CheckCircle2,
  MapPin,
  Package,
  Maximize2,
  RefreshCw,
  Trophy,
  Medal,
  ChevronRight,
  Sparkles,
  Layers,
  Download,
  Search,
  Building2,
  CheckCircle,
  XCircle,
  Percent,
  FileSpreadsheet,
  Filter,
} from 'lucide-vue-next'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet-defaulticon-compatibility'
import 'leaflet-defaulticon-compatibility/dist/leaflet-defaulticon-compatibility.css'

import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'

delete (L.Icon.Default.prototype as any)._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow
})
import { useMasterDataStore } from '@/stores/masterData'
import ProductMultiSelectDropdown from '@/components/sales/ProductMultiSelectDropdown.vue'

const masterDataStore = useMasterDataStore()

use([
  CanvasRenderer,
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  DataZoomComponent,
])

const isLoading = ref(true)
const mapContainer = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let abortController: AbortController | null = null
const stores = ref<any[]>([])
const products = ref<any[]>([])
const mapMode = ref<'STORES' | 'PRODUCTS'>('STORES')
const tileLayerMode = ref<'STREET' | 'SATELLITE'>('STREET')
let currentTileLayer: L.TileLayer | null = null
const selectedProductIds = ref<string[]>([])
const productSalesData = ref<any[]>([])
const isMapLoading = ref(false)

const isFullDistributionLoaded = ref(false)
const isLoadingDistribution = ref(false)
const totalDistributionCount = ref(0)
const mappedDistributionCount = ref(0)

const activeRegionTab = ref<'KABUPATEN' | 'KECAMATAN' | 'DESA'>('KABUPATEN')
const selectedRegion = ref<string | null>(null)
const regionSearchQuery = ref('')

const getNormalizedRegion = (store: any, type: 'PROVINCE' | 'KABUPATEN' | 'KECAMATAN' | 'DESA'): string | null => {
  return getAccurateStoreRegion(store, type)
}

const regionStats = computed(() => {
  const stats = new Map<string, number>()
  stores.value.forEach(s => {
    let key = 'Tidak Terdeteksi'
    const region = getNormalizedRegion(s, activeRegionTab.value)
    if (region) key = region
    
    stats.set(key, (stats.get(key) || 0) + 1)
  })
  
  let list = Array.from(stats.entries())
    .map(([name, count]) => ({ 
      name, 
      count,
      percentage: stores.value.length > 0 ? Math.round((count / stores.value.length) * 100) : 0 
    }))
    .sort((a, b) => {
      // Pastikan 'Tidak Terdeteksi' selalu di paling akhir jika ada
      if (a.name === 'Tidak Terdeteksi') return 1
      if (b.name === 'Tidak Terdeteksi') return -1
      return b.count - a.count
    })

  if (regionSearchQuery.value.trim()) {
    const q = regionSearchQuery.value.toLowerCase().trim()
    list = list.filter(item => item.name.toLowerCase().includes(q))
  }

  return list
})

const handleRegionClick = (regionName: string) => {
  if (selectedRegion.value === regionName) {
    selectedRegion.value = null // toggle off
  } else {
    selectedRegion.value = regionName
  }
  updateMarkers(true)

  if (selectedRegion.value && mapContainer.value) {
    mapContainer.value.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
}

// --- Helper Ringkas Rupiah ---
const formatCompactRupiah = (val: number): string => {
  if (Math.abs(val) >= 1_000_000_000) {
    return `${(val / 1_000_000_000).toFixed(val % 1_000_000_000 === 0 ? 0 : 1)} M`
  }
  if (Math.abs(val) >= 1_000_000) {
    return `${(val / 1_000_000).toFixed(val % 1_000_000 === 0 ? 0 : 1)} Jt`
  }
  if (Math.abs(val) >= 1_000) {
    return `${(val / 1_000).toFixed(0)} Rb`
  }
  return val.toString()
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

// Helper untuk mendapatkan area operasional / wilayah penjualan salesman
const getSalesmanArea = (salesOrKpi: any): string => {
  if (!salesOrKpi) return '-'
  
  // 1. Explicit working_territory
  const explicit = salesOrKpi.workingTerritory || salesOrKpi.working_territory || salesOrKpi.WorkingTerritory || salesOrKpi.Employee?.working_territory
  if (explicit && typeof explicit === 'string' && explicit.trim() && explicit.trim() !== '-') {
    return explicit.trim()
  }

  // 2. Derive from assigned stores
  const empId = salesOrKpi.employeeId || salesOrKpi.EmployeeID || salesOrKpi.employee_id || salesOrKpi.id
  if (empId && stores.value && stores.value.length > 0) {
    const assignedStores = stores.value.filter((s: any) => s.assigned_employee_id === empId)
    if (assignedStores.length > 0) {
      const distinctRegencies = Array.from(new Set(
        assignedStores
          .map((s: any) => getNormalizedRegion(s, 'KABUPATEN'))
          .filter((r): r is string => !!r && r !== 'Tidak Terdeteksi')
      ))
      if (distinctRegencies.length > 0) {
        if (distinctRegencies.length <= 2) {
          return distinctRegencies.join(', ')
        }
        return `${distinctRegencies.slice(0, 2).join(', ')} (+${distinctRegencies.length - 2} area)`
      }
    }
  }

  // 3. Fallback to branch name
  const branchName = salesOrKpi.branchName || salesOrKpi.branch_name || salesOrKpi.Employee?.branch?.name || salesOrKpi.branch
  if (branchName && branchName !== '-') {
    return `Cabang ${branchName}`
  }

  return 'Area belum diset'
}

// --- Chart Options ---
const omzetChartOption = ref<any>({
  tooltip: {
    trigger: 'axis',
    backgroundColor: '#ffffff',
    borderColor: '#e2e8f0',
    borderWidth: 1,
    padding: [12, 16],
    shadowBlur: 16,
    shadowColor: 'rgba(15, 23, 42, 0.08)',
    textStyle: { color: '#0f172a', fontFamily: 'inherit' },
    axisPointer: {
      type: 'shadow',
      shadowStyle: { color: 'rgba(241, 245, 249, 0.6)' }
    },
    formatter: (params: any[]) => {
      if (!params || params.length === 0) return ''
      const name = params[0].name
      let target = 0
      let actual = 0
      
      params.forEach(p => {
        if (p.seriesName === 'Target') target = Number(p.value) || 0
        if (p.seriesName === 'Aktual') actual = Number(p.value) || 0
      })

      const hasTarget = target > 0
      const pct = hasTarget ? Math.round((actual / target) * 100) : (actual > 0 ? 100 : 0)
      const diff = actual - target

      // Match salesman area
      const matchedSales = allSalesmanRankings.value.find((s: any) => s.name === name)
      const areaInfo = matchedSales ? getSalesmanArea(matchedSales) : ''

      let statusBadge = ''
      if (!hasTarget) {
        statusBadge = '<span style="color:#64748b;background:#f1f5f9;padding:2px 8px;border-radius:9999px;font-weight:800;font-size:10px;">Target Belum Diset</span>'
      } else if (pct >= 100) {
        statusBadge = '<span style="color:#059669;background:#ecfdf5;padding:2px 8px;border-radius:9999px;font-weight:800;font-size:10px;">★ Tercapai</span>'
      } else if (pct >= 75) {
        statusBadge = '<span style="color:#2563eb;background:#eff6ff;padding:2px 8px;border-radius:9999px;font-weight:800;font-size:10px;">On Track</span>'
      } else {
        statusBadge = '<span style="color:#d97706;background:#fffbeb;padding:2px 8px;border-radius:9999px;font-weight:800;font-size:10px;">Perlu Kejar</span>'
      }

      return `
        <div style="font-family: inherit; min-width: 200px;">
          <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:4px; border-bottom:1px solid #f1f5f9; padding-bottom:6px;">
            <strong style="font-size:13px; color:#0f172a;">${name}</strong>
            ${statusBadge}
          </div>
          ${areaInfo && areaInfo !== '-' ? `
          <div style="font-size:10px; color:#64748b; margin-bottom:6px; display:flex; align-items:center; gap:4px;">
            <span>📍</span>
            <span style="font-weight:600; color:#475569;">${areaInfo}</span>
          </div>` : ''}
          <div style="font-size:11px; display:flex; justify-content:space-between; margin-bottom:4px;">
            <span style="color:#64748b;">🎯 Target:</span>
            <strong style="color:#334155;">${hasTarget ? formatCurrency(target) : '<span style="color:#94a3b8;font-weight:normal;">Belum diset</span>'}</strong>
          </div>
          <div style="font-size:11px; display:flex; justify-content:space-between; margin-bottom:6px;">
            <span style="color:#64748b;">💰 Aktual:</span>
            <strong style="color:#990000; font-size:12px;">${formatCurrency(actual)}</strong>
          </div>
          ${hasTarget ? `
          <div style="font-size:11px; display:flex; justify-content:space-between; margin-top:6px; border-top:1px dashed #e2e8f0; padding-top:6px;">
            <span style="color:#64748b;">Pencapaian:</span>
            <strong style="color:${pct >= 100 ? '#059669' : (pct >= 75 ? '#2563eb' : '#d97706')}; font-size:12px;">${pct}%</strong>
          </div>
          <div style="font-size:10px; display:flex; justify-content:space-between; margin-top:2px;">
            <span style="color:#94a3b8;">Gap / Selisih:</span>
            <span style="color:${diff >= 0 ? '#059669' : '#dc2626'}; font-weight:700;">${diff >= 0 ? '+' : ''}${formatCurrency(diff)}</span>
          </div>
          ` : `
          <div style="font-size:10px; color:#94a3b8; margin-top:6px; border-top:1px dashed #e2e8f0; padding-top:6px; text-align:center;">
            Progress omzet berjalan saat ini
          </div>
          `}
        </div>
      `
    }
  },
  legend: { 
    bottom: 4, 
    icon: 'roundRect', 
    itemWidth: 14,
    itemHeight: 10,
    itemGap: 24,
    textStyle: { fontWeight: '700', fontSize: 11, color: '#475569' } 
  },
  grid: { left: '2%', right: '3%', bottom: '16%', top: '6%', containLabel: true },
  xAxis: {
    type: 'category',
    data: [],
    axisLine: { lineStyle: { color: '#e2e8f0' } },
    axisTick: { show: false },
    axisLabel: { 
      fontSize: 10, 
      fontWeight: '700', 
      color: '#64748b',
      interval: 0,
      rotate: 0,
      formatter: (val: string) => {
        return val.length > 12 ? val.slice(0, 10) + '…' : val
      }
    },
  },
  yAxis: { 
    type: 'value', 
    splitLine: { lineStyle: { color: '#f8fafc', type: 'dashed' } },
    axisLabel: {
      fontSize: 10,
      fontWeight: '700',
      color: '#94a3b8',
      formatter: (val: number) => formatCompactRupiah(val)
    }
  },
  dataZoom: [
    {
      type: 'inside',
      start: 0,
      end: 100,
      zoomLock: false
    }
  ],
  series: [
    {
      name: 'Target',
      type: 'bar',
      data: [],
      itemStyle: { 
        color: '#e2e8f0', 
        borderRadius: [6, 6, 0, 0] 
      },
      barMaxWidth: 22,
    },
    {
      name: 'Aktual',
      type: 'bar',
      data: [],
      itemStyle: { 
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: '#b91c1c' }, // dark red
            { offset: 1, color: '#dc2626' }  // bright red
          ]
        },
        borderRadius: [6, 6, 0, 0],
        shadowColor: 'rgba(185, 28, 28, 0.25)',
        shadowBlur: 6,
        shadowOffsetY: 3
      },
      barMaxWidth: 22,
    },
  ],
})

const spandukChartOption = ref({
  tooltip: { trigger: 'item' },
  series: [
    {
      name: 'Progress Pemasangan',
      type: 'pie',
      radius: ['60%', '85%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      data: [
        { value: 45, name: 'Selesai', itemStyle: { color: '#10b981' } },
        { value: 20, name: 'Proses', itemStyle: { color: '#3b82f6' } },
        { value: 15, name: 'Pending', itemStyle: { color: '#f59e0b' } },
      ],
    },
  ],
})

const kpiSummary = ref({
  totalOmzetLama: 0,
  totalOmzetBaru: 0,
  totalOmzetAll: 0,
  totalTokoBaru: 0,
})

const spandukStats = ref({ completed: 0, inProgress: 0, pending: 0, total: 1 })
const totalSalesmen = ref(0)
const pendingNotesCount = ref(0)
const pendingTaxDocsCount = ref(0)
const allSalesmanRankings = ref<any[]>([])
const allBranchRankings = ref<any[]>([])
const isRankingsExpanded = ref(false)

const topSalesman = computed(() => {
  if (isRankingsExpanded.value) return allSalesmanRankings.value
  return allSalesmanRankings.value.slice(0, 5)
})

const topBranches = computed(() => {
  if (isRankingsExpanded.value) return allBranchRankings.value
  return allBranchRankings.value.slice(0, 5)
})
const topProducts = ref<any[]>([])
const recentNotas = ref<any[]>([])
const activeInsightTab = ref<'WILAYAH' | 'PRODUK'>('WILAYAH')
const activeRankingTab = ref<'SALESMAN' | 'CABANG'>('SALESMAN')

const regionCounts = computed(() => {
  const provMap = new Map<string, number>()
  const palette = [
    { color: 'bg-emerald-500', text: 'text-emerald-600', bg: 'bg-emerald-50' },
    { color: 'bg-blue-500', text: 'text-blue-600', bg: 'bg-blue-50' },
    { color: 'bg-indigo-500', text: 'text-indigo-600', bg: 'bg-indigo-50' },
    { color: 'bg-amber-500', text: 'text-amber-600', bg: 'bg-amber-50' },
    { color: 'bg-purple-500', text: 'text-purple-600', bg: 'bg-purple-50' },
    { color: 'bg-rose-500', text: 'text-rose-600', bg: 'bg-rose-50' },
    { color: 'bg-teal-500', text: 'text-teal-600', bg: 'bg-teal-50' },
    { color: 'bg-cyan-500', text: 'text-cyan-600', bg: 'bg-cyan-50' },
  ]

  stores.value.forEach((s) => {
    const prov = getNormalizedRegion(s, 'PROVINCE') || 'Lainnya'
    provMap.set(prov, (provMap.get(prov) || 0) + 1)
  })

  return Array.from(provMap.entries())
    .map(([name, value], idx) => {
      const theme = palette[idx % palette.length]
      return {
        name,
        value,
        color: theme?.color || 'bg-slate-500',
        text: theme?.text || 'text-slate-600',
        bg: theme?.bg || 'bg-slate-50',
      }
    })
    .sort((a, b) => {
      // Pastikan 'Lainnya' selalu di paling akhir jika ada
      if (a.name === 'Lainnya') return 1
      if (b.name === 'Lainnya') return -1
      return b.value - a.value
    })
    .filter((x) => x.value > 0)
})

// --- Filter Periode State ---
const periodType = ref<'daily' | 'monthly' | 'yearly' | 'custom'>('daily')
const selectedDate = ref<string>(new Date().toISOString().split('T')[0] as string)
const selectedStartDate = ref<string>(new Date().toISOString().split('T')[0] as string)
const selectedEndDate = ref<string>(new Date().toISOString().split('T')[0] as string)
const selectedMonth = ref(new Date().getMonth() + 1)
const selectedYear = ref(new Date().getFullYear())
const availableYears = computed(() => {
  const current = new Date().getFullYear()
  return [current - 2, current - 1, current, current + 1]
})

// --- State Rekapitulasi Omzet Per Kota / Penetrasi Wilayah ---
const cityRevenueData = ref<any>(null)
const isCityRevenueLoading = ref(false)
const isExportingCityExcel = ref(false)
const cityRevenueSearch = ref('')
const cityPenetrationFilter = ref<'ALL' | 'PENETRATED' | 'UNPENETRATED'>('ALL')
const selectedCityProducts = ref<string[]>([])

const fetchCityRevenueReport = async () => {
  isCityRevenueLoading.value = true
  try {
    const branchId = masterDataStore.selectedBranchId === 'ALL_BRANCHES' ? '' : masterDataStore.selectedBranchId || ''
    const companyId = masterDataStore.selectedBranchCompanyId || ''
    const salesType = masterDataStore.selectedSalesType || ''
    const prodIdsStr = selectedCityProducts.value.length > 0 ? selectedCityProducts.value.join(',') : ''

    const startDate = periodType.value === 'custom' ? selectedStartDate.value : selectedDate.value
    const endDate = periodType.value === 'custom' ? selectedEndDate.value : selectedDate.value

    const res = await apiClient.get('/admin/sales/reports/city-revenue', {
      params: {
        period_type: periodType.value,
        date: selectedDate.value,
        start_date: startDate,
        end_date: endDate,
        month: selectedMonth.value,
        year: selectedYear.value,
        branch_id: branchId,
        company_id: companyId,
        sales_type: salesType,
        product_ids: prodIdsStr,
      }
    })
    if (res.data?.data) {
      cityRevenueData.value = res.data.data
    }
  } catch (err: any) {
    console.error('Gagal mengambil rekapitulasi omzet per kota:', err)
  } finally {
    isCityRevenueLoading.value = false
  }
}

const exportCityRevenueExcel = async () => {
  isExportingCityExcel.value = true
  try {
    const branchId = masterDataStore.selectedBranchId === 'ALL_BRANCHES' ? '' : masterDataStore.selectedBranchId || ''
    const companyId = masterDataStore.selectedBranchCompanyId || ''
    const salesType = masterDataStore.selectedSalesType || ''
    const prodIdsStr = selectedCityProducts.value.length > 0 ? selectedCityProducts.value.join(',') : ''

    const startDate = periodType.value === 'custom' ? selectedStartDate.value : selectedDate.value
    const endDate = periodType.value === 'custom' ? selectedEndDate.value : selectedDate.value

    const res = await apiClient.get('/admin/sales/reports/city-revenue/export', {
      params: {
        period_type: periodType.value,
        date: selectedDate.value,
        start_date: startDate,
        end_date: endDate,
        month: selectedMonth.value,
        year: selectedYear.value,
        branch_id: branchId,
        company_id: companyId,
        sales_type: salesType,
        product_ids: prodIdsStr,
      },
      responseType: 'blob'
    })

    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `rekap_omzet_per_kota_${new Date().toISOString().split('T')[0]}.xlsx`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch (err: any) {
    console.error('Gagal mengunduh excel rekap omzet kota:', err)
    alert('Gagal mengunduh data excel omzet per kota.')
  } finally {
    isExportingCityExcel.value = false
  }
}

const filteredCityItems = computed(() => {
  if (!cityRevenueData.value?.cities) return []
  let items = cityRevenueData.value.cities

  if (cityPenetrationFilter.value !== 'ALL') {
    items = items.filter((c: any) => c.status === cityPenetrationFilter.value)
  }

  if (cityRevenueSearch.value.trim()) {
    const q = cityRevenueSearch.value.toLowerCase().trim()
    items = items.filter((c: any) => 
      (c.city && c.city.toLowerCase().includes(q)) ||
      (c.province && c.province.toLowerCase().includes(q)) ||
      (c.branch_name && c.branch_name.toLowerCase().includes(q)) ||
      (c.top_salesman_name && c.top_salesman_name.toLowerCase().includes(q))
    )
  }

  return items
})

const handleCityRowClick = async (cityName: string) => {
  if (!cityName || cityName === 'Tidak Terdeteksi') return
  activeRegionTab.value = 'KABUPATEN'
  selectedRegion.value = cityName
  if (!isFullDistributionLoaded.value && !isLoadingDistribution.value) {
    await fetchStoreDistribution(true)
  } else {
    updateMarkers(true)
  }
  if (mapContainer.value) {
    mapContainer.value.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
}

const handleSalesmanRowClick = (sales: any) => {
  if (!sales) return
  cityRevenueSearch.value = sales.name
  const citySection = document.getElementById('city-penetration-section')
  if (citySection) {
    citySection.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

const fetchDashboardData = async () => {
  if (!masterDataStore.selectedBranchId) return

  if (abortController) {
    abortController.abort()
  }
  abortController = new AbortController()
  const config = { signal: abortController.signal }

  isLoading.value = true
  const month = selectedMonth.value
  const year = selectedYear.value
  const date = selectedDate.value
  const companyId = masterDataStore.selectedBranchCompanyId || ''

  try {
    const branchId =
      masterDataStore.selectedBranchId === 'ALL_BRANCHES'
        ? ''
        : masterDataStore.selectedBranchId || ''

    const startDate = periodType.value === 'custom' ? selectedStartDate.value : selectedDate.value
    const endDate = periodType.value === 'custom' ? selectedEndDate.value : selectedDate.value

    const pSummary = apiClient.get(
      `/admin/sales/reports/summary?period_type=${periodType.value}&date=${date}&start_date=${startDate}&end_date=${endDate}&month=${month}&year=${year}&branch_id=${branchId}&sales_type=${masterDataStore.selectedSalesType}`,
      config
    )
    const pPerf = apiClient.get(
      `/admin/sales/reports/performance?period_type=${periodType.value}&date=${date}&start_date=${startDate}&end_date=${endDate}&month=${month}&year=${year}&branch_id=${branchId}&sales_type=${masterDataStore.selectedSalesType}`,
      config
    )
    
    // We no longer fetch 1,000 employees. We use masterDataStore.employees which is globally cached.
    const pEmp = Promise.resolve({ data: { data: masterDataStore.employees } })
    
    let pPending;
    if (masterDataStore.selectedSalesType === 'sales_mt') {
      pPending = apiClient.get(`/admin/sales-orders?limit=3&branch_id=${branchId}&status=WAITING_TAX_DOCS&sales_type=sales_mt`, config)
    } else {
      pPending = apiClient.get(`/admin/sales/transactions/all?limit=3&status=PENDING&company_id=${companyId}&sales_type=${masterDataStore.selectedSalesType}&branch_id=${branchId}`, config)
    }

    const pBanner = apiClient.get(`/branding/stats?sales_type=${masterDataStore.selectedSalesType}&branch_id=${branchId}`, config)
    const pProd = apiClient.get(`/factory/products?company_id=${companyId}`, config)
    const pTopProd = apiClient.get(`/admin/sales/reports/top-products?company_id=${companyId}`, config)

    const [
      summaryRes,
      perfRes,
      empRes,
      pendingRes,
      bannerRes,
      prodRes,
      topProdRes,
    ] = await Promise.allSettled([
      pSummary, pPerf, pEmp, pPending, pBanner, pProd, pTopProd
    ])

    // 1. Process Summary
    if (summaryRes.status === 'fulfilled' && summaryRes.value.data?.data) {
      const data = summaryRes.value.data.data
      kpiSummary.value = {
        totalOmzetLama: data.total_omzet_lama || 0,
        totalOmzetBaru: data.total_omzet_baru || 0,
        totalOmzetAll: data.total_omzet_all || 0,
        totalTokoBaru: data.total_toko_baru || 0,
      }
    }

    // 2. Process Performance
    let perfData: any = null
    if (perfRes.status === 'fulfilled' && perfRes.value.data?.data) {
      perfData = perfRes.value.data.data
      
      const mappedSales = perfData
        .map((kpi: any) => {
          const emp = kpi.Employee
          const fullName = emp ? `${emp.first_name || ''} ${emp.last_name || ''}`.trim() : 'Salesman'
          const target = kpi.TargetOmzet || 0
          const actual = kpi.AchievedOmzet || 0
          const hasTarget = target > 0
          const achievement = hasTarget ? Math.round((actual / target) * 100) : 0
          
          return {
            id: kpi.ID || kpi.EmployeeID,
            employeeId: kpi.EmployeeID,
            name: fullName || 'Salesman',
            target: target,
            omzet: actual,
            hasTarget: hasTarget,
            achievement: achievement,
            branchName: emp?.branch?.name || '',
            workingTerritory: kpi.working_territory || kpi.WorkingTerritory || emp?.working_territory || '',
            rank: 0,
            relativeProgress: 0,
          }
        })
        .sort((a: any, b: any) => b.omzet - a.omzet)
        .map((item: any, index: number) => ({ ...item, rank: index + 1 }))

      const maxSalesOmzet = mappedSales.length > 0 ? Math.max(...mappedSales.map((s: any) => s.omzet), 1) : 1
      mappedSales.forEach((s: any) => {
        s.relativeProgress = s.hasTarget ? Math.min(s.achievement, 100) : Math.round((s.omzet / maxSalesOmzet) * 100)
      })

      allSalesmanRankings.value = mappedSales

      if (
        omzetChartOption.value.xAxis &&
        !Array.isArray(omzetChartOption.value.xAxis) &&
        omzetChartOption.value.series
      ) {
        // Tampilkan nama lengkap / proporsional di chart
        omzetChartOption.value.xAxis.data = perfData.map((k: any) => {
          const emp = k.Employee
          if (!emp) return 'Sales'
          return `${emp.first_name || ''} ${emp.last_name || ''}`.trim() || emp.first_name || 'Sales'
        })

        if (omzetChartOption.value.series[0])
          omzetChartOption.value.series[0].data = perfData.map((k: any) => k.TargetOmzet || 0)
        if (omzetChartOption.value.series[1])
          omzetChartOption.value.series[1].data = perfData.map((k: any) => k.AchievedOmzet || 0)

        // Aktifkan dataZoom jika jumlah salesman > 6 agar chart tetap lega
        if (perfData.length > 6) {
          omzetChartOption.value.dataZoom = [
            {
              type: 'inside',
              start: 0,
              end: Math.min(100, Math.round((6 / perfData.length) * 100)),
              zoomLock: false
            }
          ]
        } else {
          omzetChartOption.value.dataZoom = [{ type: 'inside', start: 0, end: 100 }]
        }
      }
    }

    // 2.1 Process Employees
    if (empRes.status === 'fulfilled' && empRes.value.data?.data) {
      let allEmps = empRes.value.data.data
      if (branchId) {
        const branchIds = branchId.split(',').map((id: string) => id.trim()).filter(Boolean)
        allEmps = allEmps.filter((emp: any) => 
          branchIds.includes(emp.branch_id) || 
          (emp.branch && branchIds.includes(emp.branch.id)) ||
          branchIds.includes(emp.operating_branch_id)
        )
      }
      const salesmenEmps = allEmps.filter((emp: any) => {
        const jobTitle = (emp.job_position?.title || '').toLowerCase()
        return jobTitle.includes('sales') || jobTitle.includes('marketing')
      })
      totalSalesmen.value = salesmenEmps.length

      if (perfData) {
        const branchMap = new Map()
        perfData.forEach((kpi: any) => {
          const empMatch = allEmps.find((e: any) => e.id === kpi.EmployeeID)
          const branchName = empMatch?.branch?.name || 'Cabang Lainnya'

          if (!branchMap.has(branchName)) {
            branchMap.set(branchName, {
              id: branchName,
              name: branchName,
              omzet: 0,
              target: 0,
            })
          }
          const b = branchMap.get(branchName)
          b.omzet += kpi.AchievedOmzet || 0
          b.target += kpi.TargetOmzet || 0
        })

        const mappedBranches = Array.from(branchMap.values())
          .map((b) => {
            const hasTarget = b.target > 0
            const achievement = hasTarget ? Math.round((b.omzet / b.target) * 100) : 0
            return {
              ...b,
              hasTarget: hasTarget,
              achievement: achievement,
              relativeProgress: 0,
            }
          })
          .sort((a: any, b: any) => b.omzet - a.omzet)
          .map((item: any, index: number) => ({ ...item, rank: index + 1 }))

        const maxBranchOmzet = mappedBranches.length > 0 ? Math.max(...mappedBranches.map((b: any) => b.omzet), 1) : 1
        mappedBranches.forEach((b: any) => {
          b.relativeProgress = b.hasTarget ? Math.min(b.achievement, 100) : Math.round((b.omzet / maxBranchOmzet) * 100)
        })

        allBranchRankings.value = mappedBranches
      }
    }

    // 3. Process Pending Notas / Tax Docs
    if (pendingRes.status === 'fulfilled' && pendingRes.value.data?.data) {
      if (masterDataStore.selectedSalesType === 'sales_mt') {
        pendingTaxDocsCount.value = pendingRes.value.data.meta?.total || pendingRes.value.data.data.length
        recentNotas.value = pendingRes.value.data.data.slice(0, 3).map((n: any) => ({
          id: n.so_number || n.id.slice(0, 8),
          store: n.store?.name || 'Outlet MT',
          amount: n.total_amount,
          status: n.status,
          time: new Date(n.created_at).toLocaleTimeString(),
        }))
      } else {
        let pendingList = pendingRes.value.data.data
        // Filtering by branch is no longer strictly necessary because the API already filters it, 
        // but we'll leave the array processing as is for the recent 3 items.
        pendingNotesCount.value = pendingRes.value.data.meta?.total || pendingList.length
        recentNotas.value = pendingList.slice(0, 3).map((n: any) => ({
          id: n.receipt_no || n.id.slice(0, 8),
          store: n.store?.name || 'Toko',
          amount: n.total_amount,
          status: n.status === 'VERIFIED' ? 'APPROVED' : n.status,
          time: new Date(n.created_at).toLocaleTimeString(),
        }))
      }
    }

    // 4. Process Banner Stats
    if (bannerRes.status === 'fulfilled' && bannerRes.value.data?.data) {
      const stats = bannerRes.value.data.data
      const completed = stats.installed_banners || 0
      const pending = stats.pending_banners || 0
      const total = stats.total_banners || 1
      const inProgress = Math.max(0, (stats.total_banners || 0) - completed - pending)

      spandukStats.value = { completed, inProgress, pending, total }
      if (spandukChartOption.value.series && spandukChartOption.value.series[0]) {
        spandukChartOption.value.series[0].data = [
          { value: completed, name: 'Selesai', itemStyle: { color: '#10b981' } },
          { value: inProgress, name: 'Proses', itemStyle: { color: '#3b82f6' } },
          { value: pending, name: 'Pending', itemStyle: { color: '#f59e0b' } },
        ]
      }
    }

    // 5. Process Products
    if (prodRes.status === 'fulfilled' && prodRes.value.data) {
      const rawList = Array.isArray(prodRes.value.data) ? prodRes.value.data : prodRes.value.data.data || []
      products.value = filterSellableProducts(rawList)
    }

    // 7. Process Top Products
    if (topProdRes.status === 'fulfilled' && topProdRes.value.data?.data) {
      topProducts.value = topProdRes.value.data.data
    }

    // 8. Process City Revenue Report
    fetchCityRevenueReport()

  } catch (error) {
    console.error('Failed to fetch dashboard data:', error)
  } finally {
    isLoading.value = false
  }
}

const fetchStoreDistribution = async (fitBounds = true) => {
  if (abortController) {
    abortController.abort()
  }
  abortController = new AbortController()
  const config = { signal: abortController.signal }

  isLoadingDistribution.value = true
  isMapLoading.value = true
  try {
    const companyId = masterDataStore.selectedBranchCompanyId || ''
    const branchId = masterDataStore.selectedBranchId === 'ALL_BRANCHES' ? '' : masterDataStore.selectedBranchId
    const salesType = masterDataStore.selectedSalesType || ''

    const res = await apiClient.get(
      `/stores/distribution-map?company_id=${companyId}&branch_id=${branchId}&sales_type=${salesType}&status=ACTIVE`,
      config
    )
    if (res.data && res.data.data) {
      const rawStores = Array.isArray(res.data.data) ? res.data.data : []
      stores.value = rawStores.map((s: any) => ({
        id: s.id,
        name: s.name,
        owner_name: s.owner_name,
        phone_number: s.phone_number,
        latitude: Number(s.latitude) || 0,
        longitude: Number(s.longitude) || 0,
        address: s.address,
        province: s.province,
        regency: s.regency,
        district: s.district,
        village: s.village,
        category: s.category,
        customer_type: s.customer_type,
        is_active: s.is_active,
        assigned_employee_id: s.assigned_employee_id,
        assigned_employee: s.assigned_employee_id ? {
          id: s.assigned_employee_id,
          first_name: s.employee_first_name || '',
          last_name: s.employee_last_name || '',
          branch: s.branch_id ? {
            id: s.branch_id,
            name: s.branch_name || '',
          } : null,
        } : null,
      }))
      totalDistributionCount.value = res.data.total || stores.value.length
      mappedDistributionCount.value = res.data.mapped_total || stores.value.filter((s: any) => s.latitude && s.longitude).length
      isFullDistributionLoaded.value = true

      if (mapMode.value === 'STORES') {
        updateMarkers(fitBounds)
      }
    }
  } catch (error: any) {
    if (error.name !== 'CanceledError') {
      console.error('Failed to fetch store distribution map:', error)
    }
  } finally {
    isLoadingDistribution.value = false
    isMapLoading.value = false
  }
}

const fetchProductDistribution = async () => {
  if (selectedProductIds.value.length === 0) {
    productSalesData.value = []
    if (markerGroup) markerGroup.clearLayers()
    return
  }
  if (abortController) {
    abortController.abort()
  }
  abortController = new AbortController()
  const config = { signal: abortController.signal }

  isMapLoading.value = true
  try {
    const companyId = masterDataStore.selectedBranchCompanyId || ''
    const prodIdsStr = selectedProductIds.value.join(',')
    const res = await apiClient.get(
      `/admin/sales/reports/product-distribution?product_ids=${prodIdsStr}&company_id=${companyId}`,
      config
    )
    if (res.data) {
      productSalesData.value = res.data.data || []
      
      // Preload product image of first selected item into browser cache
      const selectedProduct = products.value.find((p) => p.id === selectedProductIds.value[0])
      const baseUrl = import.meta.env.VITE_API_BASE_URL?.replace('/api', '') || ''
      const rawImage = selectedProduct?.image_url
      if (rawImage) {
        const imgUrl = rawImage.startsWith('http') ? rawImage : baseUrl + rawImage
        const img = new Image()
        img.src = imgUrl
      }

      updateMarkers(true)
    }
  } catch (error: any) {
    if (error.name !== 'CanceledError') {
      console.error('Failed to fetch product distribution:', error)
    }
  } finally {
    isMapLoading.value = false
  }
}

const getEmployeeColor = (id: string | null | undefined) => {
  if (!id) return '#94a3b8' // Default slate-400
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash)
  }
  const c = (hash & 0x00ffffff).toString(16).toUpperCase()
  return '#' + '000000'.substring(0, 6 - c.length) + c
}

let markerGroup: L.LayerGroup | null = null

const renderMarkers = () => {
  if (!map || !markerGroup) return

  markerGroup.clearLayers()
  const zoom = map.getZoom()

  if (mapMode.value === 'STORES') {
    let validStores = stores.value.filter((s) => s.latitude && s.longitude && s.latitude !== 0 && s.longitude !== 0)
    
    // Filter by selected region if any
    if (selectedRegion.value) {
      validStores = validStores.filter(s => {
        const region = getNormalizedRegion(s, activeRegionTab.value)
        if (selectedRegion.value === 'Tidak Terdeteksi') {
          return !region
        }
        return region === selectedRegion.value
      })
    }

    // Limit DOM markers if zoom is low to avoid excessive DOM nodes
    const maxMarkers = selectedRegion.value ? 2000 : (zoom <= 8 ? 800 : (zoom <= 11 ? 1500 : 2500))
    const storesToRender = validStores.slice(0, maxMarkers)

    storesToRender.forEach((store) => {
      const salesmanColor = getEmployeeColor(store.assigned_employee_id)

      let iconHtml = ''
      let iconSize: [number, number] = [16, 16]
      let iconAnchor: [number, number] = [8, 8]

      if (zoom <= 9) {
        // Skala Sangat Jauh (Provinsi / Nasional) -> Titik Dot Kecil Minimalis
        iconSize = [8, 8]
        iconAnchor = [4, 4]
        iconHtml = `
          <div class="group relative flex items-center justify-center cursor-pointer transition-transform duration-200 hover:scale-150">
            <div class="w-2 h-2 rounded-full ring-1 ring-white shadow-sm" 
                 style="background-color: ${salesmanColor}; opacity: 0.9;"></div>
            <div class="absolute hidden group-hover:block left-3 top-[-8px] bg-slate-900/90 backdrop-blur-sm text-white border border-slate-700 rounded-md px-2 py-1 shadow-xl z-[1000] whitespace-nowrap pointer-events-none">
              <p class="text-[9px] font-bold">${store.name}</p>
            </div>
          </div>
        `
      } else if (zoom <= 12) {
        // Skala Menengah (Kabupaten / Kota) -> Dot Sedang dengan Border Putih
        iconSize = [12, 12]
        iconAnchor = [6, 6]
        iconHtml = `
          <div class="group relative flex items-center justify-center cursor-pointer transition-transform duration-200 hover:scale-125">
            <div class="w-3 h-3 rounded-full border-1.5 border-white shadow-md" 
                 style="background-color: ${salesmanColor};"></div>
            <div class="absolute hidden group-hover:block left-4 top-[-10px] bg-white border border-slate-200 rounded-lg px-2.5 py-1.5 shadow-xl z-[1000] whitespace-nowrap">
              <p class="text-[9px] font-black text-slate-900">${store.name}</p>
              <p class="text-[8px] font-bold text-slate-400 mt-0.5 uppercase">${store.assigned_employee?.first_name || 'Unassigned'}</p>
            </div>
          </div>
        `
      } else {
        // Skala Dekat (Kecamatan / Jalan) -> Pin Interaktif Penuh
        iconSize = [18, 18]
        iconAnchor = [9, 9]
        iconHtml = `
          <div class="relative group cursor-pointer">
            <div class="w-4.5 h-4.5 rounded-full border-2 border-white shadow-md transition-all duration-200 group-hover:scale-125 flex items-center justify-center" 
                 style="background-color: ${salesmanColor};">
              <div class="w-1.5 h-1.5 bg-white rounded-full"></div>
            </div>
            <div class="absolute hidden group-hover:block left-6 top-[-6px] bg-white border border-slate-200 rounded-lg p-2 shadow-xl z-[1000] whitespace-nowrap">
              <p class="text-[10px] font-black text-slate-900">${store.name}</p>
              <p class="text-[8px] font-bold text-slate-400 mt-0.5 uppercase">${store.assigned_employee?.first_name || 'Unassigned'}</p>
            </div>
          </div>
        `
      }

      const icon = L.divIcon({
        html: iconHtml,
        className: 'custom-store-marker',
        iconSize: iconSize,
        iconAnchor: iconAnchor,
      })

      const marker = L.marker([store.latitude, store.longitude], { icon })
      marker.bindPopup(`
        <div class="p-3 min-w-[200px] font-sans">
          <div class="flex items-center gap-2 mb-2 pb-2 border-b border-slate-100">
             <div class="w-3 h-3 rounded-full" style="background-color: ${salesmanColor};"></div>
             <h4 class="font-black text-slate-900 leading-tight">${store.name}</h4>
          </div>
          <div class="space-y-1.5">
            <div class="flex justify-between items-center text-[10px]">
               <span class="font-bold text-slate-400 uppercase">Salesman</span>
               <span class="font-black text-primary">${store.assigned_employee?.first_name || 'Belum Ditugaskan'}</span>
            </div>
            <div class="flex justify-between items-center text-[10px]">
               <span class="font-bold text-slate-400 uppercase">Status</span>
               <span class="font-black ${store.is_active ? 'text-emerald-500' : 'text-rose-500'}">${store.is_active ? 'AKTIF' : 'NON-AKTIF'}</span>
            </div>
            <p class="text-[9px] text-slate-500 mt-2 italic border-t border-slate-50 pt-2">${store.address || 'Tanpa Alamat'}</p>
          </div>
        </div>
      `)
      marker.addTo(markerGroup!)
    })
  } else {
    // Ultra-Lightweight Product Image Distribution Mode
    const validData = productSalesData.value.filter((d) => d.latitude && d.longitude && d.latitude !== 0 && d.longitude !== 0)
    const maxQty = Math.max(...validData.map((d) => d.total_quantity || d.totalQuantity || 0), 1)

    // Find primary selected product & compute image once
    const firstSelectedProduct = products.value.find((p) => selectedProductIds.value.includes(p.id))
    const baseUrl = import.meta.env.VITE_API_BASE_URL?.replace('/api', '') || ''
    const rawImage = firstSelectedProduct?.image_url
    const productImage = rawImage 
      ? (rawImage.startsWith('http') ? rawImage : baseUrl + rawImage)
      : 'https://cdn-icons-png.flaticon.com/512/679/679821.png'

    const selectedProductCount = selectedProductIds.value.length
    const productLabel = selectedProductCount === 1 
      ? (firstSelectedProduct?.name || 'Produk') 
      : `${selectedProductCount} Varian Produk Terpilih`

    validData.forEach((data) => {
      const qty = Number(data.total_quantity || data.totalQuantity || 0)
      const formattedQty = qty >= 1000 ? (qty / 1000).toFixed(1) + 'k' : qty.toString()
      
      let iconHtml = ''
      let iconSize: [number, number] = [28, 28]
      let iconAnchor: [number, number] = [14, 14]

      if (zoom <= 9) {
        // Zoom Jauh (Provinsi / Nasional): Mini Product Avatar (22px)
        iconSize = [24, 24]
        iconAnchor = [12, 12]
        iconHtml = `
          <div class="relative group cursor-pointer flex items-center justify-center transition-transform duration-150 hover:scale-135" style="transform: translate3d(0,0,0);">
            <div class="w-5.5 h-5.5 rounded-full bg-white border-1.5 border-emerald-500 shadow-md p-0.5 overflow-hidden flex items-center justify-center">
              <img src="${productImage}" decoding="async" class="w-full h-full object-contain rounded-full" />
            </div>
            <div class="absolute -top-1 -right-1 bg-emerald-600 text-white text-[7px] font-black px-1 rounded-full border border-white shadow-xs leading-tight">
              ${formattedQty}
            </div>
            <div class="absolute hidden group-hover:block left-6 top-[-8px] bg-slate-900/90 text-white rounded-md px-2 py-1 text-[9px] font-bold shadow-xl z-[1000] whitespace-nowrap pointer-events-none">
              ${data.store_name} (${qty} PCS)
            </div>
          </div>
        `
      } else if (zoom <= 12) {
        // Zoom Menengah (Kabupaten / Kota): Medium Product Avatar Card (32px)
        iconSize = [34, 34]
        iconAnchor = [17, 17]
        iconHtml = `
          <div class="relative group cursor-pointer flex items-center justify-center transition-transform duration-150 hover:scale-120" style="transform: translate3d(0,0,0);">
            <div class="w-8 h-8 rounded-xl bg-white border-2 border-emerald-500 shadow-md p-0.5 overflow-hidden flex items-center justify-center">
              <img src="${productImage}" decoding="async" class="w-full h-full object-contain rounded-lg" />
            </div>
            <div class="absolute -top-1.5 -right-1.5 bg-emerald-600 text-white text-[8px] font-black px-1.5 py-0.2 rounded-full border border-white shadow-sm leading-tight">
              ${formattedQty}
            </div>
            <div class="absolute hidden group-hover:block left-8 top-[-10px] bg-white border border-slate-200 rounded-lg p-2 shadow-xl z-[1000] whitespace-nowrap pointer-events-none">
              <p class="text-[10px] font-black text-slate-900">${data.store_name}</p>
              <p class="text-[9px] font-bold text-emerald-600">${qty} PCS Terjual</p>
            </div>
          </div>
        `
      } else {
        // Zoom Dekat (Kecamatan / Jalan): Full Interactive Product Pin (42px) with Store Label
        iconSize = [44, 58]
        iconAnchor = [22, 29]
        iconHtml = `
          <div class="relative group cursor-pointer flex flex-col items-center transition-transform duration-150 hover:scale-115" style="transform: translate3d(0,0,0);">
            <div class="relative w-10 h-10 rounded-xl bg-white border-2 border-emerald-500 shadow-lg p-0.5 overflow-hidden flex items-center justify-center">
              <img src="${productImage}" decoding="async" class="w-full h-full object-contain rounded-lg" />
              <div class="absolute -top-1.5 -right-1.5 bg-emerald-600 text-white text-[8px] font-black px-1.5 py-0.2 rounded-full border border-white shadow-sm">
                ${formattedQty}
              </div>
            </div>
            <div class="bg-slate-900/90 backdrop-blur-xs text-white text-[8px] font-black px-2 py-0.5 rounded-full shadow-md mt-1 max-w-[120px] truncate border border-white/20">
              ${data.store_name}
            </div>
          </div>
        `
      }

      const icon = L.divIcon({
        html: iconHtml,
        className: 'custom-product-marker',
        iconSize: iconSize,
        iconAnchor: iconAnchor,
      })

      const marker = L.marker([data.latitude, data.longitude], { icon })

      marker.bindPopup(`
        <div class="p-3 min-w-[220px] font-sans">
          <div class="flex items-center gap-3 mb-2.5 pb-2.5 border-b border-slate-100">
             <img src="${productImage}" class="w-10 h-10 rounded-lg object-contain bg-slate-50 border border-slate-100 p-0.5" />
             <div class="min-w-0">
                <h4 class="font-black text-slate-900 leading-tight truncate">${data.store_name}</h4>
                <p class="text-[9px] text-slate-400 font-bold uppercase tracking-wider mt-0.5">${productLabel}</p>
             </div>
          </div>
          <div class="space-y-2">
            <div class="flex justify-between items-center text-[10px]">
               <span class="font-bold text-slate-400 uppercase">Total Serapan</span>
               <span class="text-sm font-black text-emerald-600">${qty} PCS</span>
            </div>
            <div class="h-1.5 w-full bg-slate-100 rounded-full overflow-hidden">
               <div class="h-full bg-emerald-500 rounded-full" style="width: ${(qty / maxQty) * 100}%"></div>
            </div>
            <div class="flex justify-between text-[8px] text-slate-400 font-medium">
               <span>Kontribusi Outlet</span>
               <span>${Math.round((qty / maxQty) * 100)}% dari top outlet</span>
            </div>
          </div>
        </div>
      `)

      marker.addTo(markerGroup!)
    })
  }
}

const switchTileLayer = (mode: 'STREET' | 'SATELLITE') => {
  if (!map) return
  if (currentTileLayer) {
    map.removeLayer(currentTileLayer)
  }
  if (mode === 'SATELLITE') {
    currentTileLayer = L.tileLayer(
      'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
      {
        attribution:
          'Tiles &copy; Esri &mdash; Source: Esri, Maxar, GeoEye, Earthstar Geographics, CNES/Airbus DS, USDA, USGS, AeroGRID, IGN, and the GIS User Community',
        maxZoom: 19,
      }
    ).addTo(map)
  } else {
    currentTileLayer = L.tileLayer(
      'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
      {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
        maxZoom: 19,
      }
    ).addTo(map)
  }
  // Pastikan marker group tetap di atas tile layer
  if (markerGroup) {
    (markerGroup as any)?.eachLayer?.((layer: any) => layer.bringToFront?.())
  }
}

const updateMarkers = (fitBounds: boolean | any = true) => {
  const shouldFit = typeof fitBounds === 'boolean' ? fitBounds : true
  if (!mapContainer.value) return

  if (!map) {
    const center: L.LatLngExpression =
      stores.value.length > 0
        ? [stores.value[0].latitude, stores.value[0].longitude]
        : [-6.2088, 106.8456]
    map = L.map(mapContainer.value).setView(center, 12)

    // Inisialisasi tile layer sesuai mode yang sedang aktif
    if (tileLayerMode.value === 'SATELLITE') {
      currentTileLayer = L.tileLayer(
        'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
        {
          attribution:
            'Tiles &copy; Esri &mdash; Source: Esri, Maxar, GeoEye, Earthstar Geographics, CNES/Airbus DS, USDA, USGS, AeroGRID, IGN, and the GIS User Community',
          maxZoom: 19,
        }
      ).addTo(map)
    } else {
      currentTileLayer = L.tileLayer(
        'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
        {
          attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
          maxZoom: 19,
        }
      ).addTo(map)
    }

    markerGroup = L.layerGroup().addTo(map)

    // Pasang listener event zoomend agar marker otomatis berubah saat user zoom in/out
    map.on('zoomend', () => {
      renderMarkers()
    })
  }

  renderMarkers()

  if (shouldFit && map) {
    if (mapMode.value === 'STORES') {
      let validStores = stores.value.filter((s) => s.latitude && s.longitude && s.latitude !== 0 && s.longitude !== 0)
      if (selectedRegion.value) {
        validStores = validStores.filter(s => {
          const region = getNormalizedRegion(s, activeRegionTab.value)
          if (selectedRegion.value === 'Tidak Terdeteksi') return !region
          return region === selectedRegion.value
        })
      }
      if (validStores.length > 0) {
        const bounds = L.latLngBounds(validStores.map(s => [s.latitude, s.longitude]))
        map.fitBounds(bounds, { padding: [40, 40], maxZoom: selectedRegion.value ? 14 : 12 })
      }
    } else {
      const validData = productSalesData.value.filter((d) => d.latitude && d.longitude && d.latitude !== 0 && d.longitude !== 0)
      if (validData.length > 0) {
        const bounds = L.latLngBounds(validData.map((d) => [d.latitude, d.longitude]))
        map.fitBounds(bounds, { padding: [30, 30] })
      }
    }
  }
}

const isSyncing = ref(false)

const syncUnknownRegions = async () => {
  if (!confirm('Apakah Anda yakin ingin menyinkronkan wilayah yang Tidak Terdeteksi secara otomatis? Proses ini mungkin memerlukan waktu beberapa menit.')) return
  
  isSyncing.value = true
  try {
    const res = await apiClient.post('/stores/sync-regions')
    alert(res.data.message)
  } catch (error: any) {
    alert(error.response?.data?.error || 'Gagal memulai sinkronisasi.')
  } finally {
    isSyncing.value = false
  }
}

onMounted(async () => {
  await masterDataStore.fetchEmployees()
  fetchDashboardData()
})

watch(
  () => masterDataStore.selectedBranchId,
  () => {
    stores.value = []
    isFullDistributionLoaded.value = false
    totalDistributionCount.value = 0
    mappedDistributionCount.value = 0
    if (markerGroup) markerGroup.clearLayers()
    fetchDashboardData()
  },
)

watch(
  () => masterDataStore.selectedSalesType,
  () => {
    stores.value = []
    isFullDistributionLoaded.value = false
    totalDistributionCount.value = 0
    mappedDistributionCount.value = 0
    if (markerGroup) markerGroup.clearLayers()
    fetchDashboardData()
  },
)

watch(mapMode, (newMode) => {
  if (newMode === 'STORES') {
    if (isFullDistributionLoaded.value) {
      updateMarkers()
    }
  } else if (selectedProductIds.value.length > 0) {
    fetchProductDistribution()
  }
})

watch(selectedProductIds, () => {
  if (mapMode.value === 'PRODUCTS') {
    fetchProductDistribution()
  }
}, { deep: true })

watch(selectedCityProducts, () => {
  fetchCityRevenueReport()
}, { deep: true })

watch(tileLayerMode, (newMode) => {
  switchTileLayer(newMode)
})
const globalFilterText = computed(() => {
  return masterDataStore.selectedBranchName
})

// --- Dynamic Texts based on Sales Type ---
const dashboardTitle = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'Modern Trade (MT) Dashboard'
  if (type === 'sales_to') return 'Taking Order (TO) Dashboard'
  if (type === 'sales_motoris_booster') return 'Motoris & Booster Dashboard'
  return 'Sales & Marketing Dashboard'
})

const dashboardSubtitle = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'Pantau performa Key Account (KAM), sewa display, dan realisasi omzet MT.'
  return 'Pantau performa penjualan dan kampanye pemasaran hari ini.'
})

const salesmanLabel = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'Total Sales MT & KAM'
  if (type === 'sales_to') return 'Total Sales TO'
  if (type === 'sales_motoris_booster') return 'Total Motoris & Booster'
  return 'Total Salesman'
})

const omzetLabel = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'Omzet Modern Trade'
  if (type === 'sales_to') return 'Omzet Sales TO'
  if (type === 'sales_motoris_booster') return 'Omzet Motoris'
  return 'Total Omzet'
})

const pendingLabel = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'PO Menunggu Dok Pajak'
  return 'Nota Menunggu'
})

const campaignLabel = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'Display Sewa / Promosi'
  return 'Spanduk Terpasang'
})

const mapTitle = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'Persebaran Outlet Modern Trade'
  return 'Persebaran Toko & Wilayah'
})

const mapSubtitle = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'outlet MT aktif'
  return 'titik distribusi aktif'
})

const chartOmzetTitle = computed(() => {
  const type = masterDataStore.selectedSalesType
  if (type === 'sales_mt') return 'Tren Realisasi Omzet MT'
  return 'Tren Realisasi Omzet'
})

watch([periodType, selectedDate, selectedStartDate, selectedEndDate, selectedMonth, selectedYear], () => {
  fetchDashboardData()
})

const periodDisplayText = computed(() => {
  if (periodType.value === 'daily') {
    return new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
  } else if (periodType.value === 'monthly') {
    const d = new Date(selectedYear.value, selectedMonth.value - 1, 1)
    return d.toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
  } else if (periodType.value === 'yearly') {
    return `Tahun ${selectedYear.value}`
  } else if (periodType.value === 'custom') {
    const startStr: string = (selectedStartDate.value || selectedDate.value || new Date().toISOString().split('T')[0]) as string
    const endStr: string = (selectedEndDate.value || selectedDate.value || new Date().toISOString().split('T')[0]) as string

    if (startStr === endStr) {
      const d = new Date(startStr)
      return d.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
    }
    const dStart = new Date(startStr)
    const dEnd = new Date(endStr)
    return `${dStart.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })} s/d ${dEnd.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}`
  }
  return 'Periode Aktif'
})

const periodSubtitleBadge = computed(() => {
  if (periodType.value === 'daily') return 'Berjalan hari ini'
  if (periodType.value === 'monthly') return `Berjalan ${new Date(selectedYear.value, selectedMonth.value - 1, 1).toLocaleDateString('id-ID', { month: 'short', year: 'numeric' })}`
  if (periodType.value === 'yearly') return `Berjalan tahun ${selectedYear.value}`
  if (periodType.value === 'custom') {
    if (selectedStartDate.value === selectedEndDate.value) {
      return `Tanggal ${selectedStartDate.value}`
    }
    return `${selectedStartDate.value} s/d ${selectedEndDate.value}`
  }
  return 'Berjalan periode ini'
})

onUnmounted(() => {
  if (abortController) {
    abortController.abort()
  }
  if (map) {
    map.remove()
    map = null
  }
})
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight">
          {{ dashboardTitle }}
          <span v-if="globalFilterText" class="text-primary text-2xl font-bold ml-2"
            >— {{ globalFilterText }}</span
          >
        </h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <TrendingUp class="w-4 h-4 text-primary" />
          {{ dashboardSubtitle }}
        </p>
      </div>
      
      <!-- Filter Bar Periode -->
      <div class="flex flex-wrap items-center gap-3 bg-white p-2 border border-slate-200 rounded-2xl shadow-sm">
        <!-- Switch Tabs -->
        <div class="flex items-center bg-slate-100/80 p-1 rounded-xl">
          <button
            @click="periodType = 'daily'"
            :class="[
              'px-3 py-1.5 rounded-lg text-xs font-black transition-all',
              periodType === 'daily'
                ? 'bg-white text-primary shadow-sm'
                : 'text-slate-500 hover:text-slate-800'
            ]"
          >
            Hari Ini
          </button>
          <button
            @click="periodType = 'monthly'"
            :class="[
              'px-3 py-1.5 rounded-lg text-xs font-black transition-all',
              periodType === 'monthly'
                ? 'bg-white text-primary shadow-sm'
                : 'text-slate-500 hover:text-slate-800'
            ]"
          >
            Bulanan
          </button>
          <button
            @click="periodType = 'yearly'"
            :class="[
              'px-3 py-1.5 rounded-lg text-xs font-black transition-all',
              periodType === 'yearly'
                ? 'bg-white text-primary shadow-sm'
                : 'text-slate-500 hover:text-slate-800'
            ]"
          >
            Tahunan
          </button>
          <button
            @click="periodType = 'custom'"
            :class="[
              'px-3 py-1.5 rounded-lg text-xs font-black transition-all',
              periodType === 'custom'
                ? 'bg-white text-primary shadow-sm'
                : 'text-slate-500 hover:text-slate-800'
            ]"
          >
            Kustom
          </button>
        </div>

        <!-- Dynamic Controls based on Period Type -->
        <div v-if="periodType === 'monthly'" class="flex items-center gap-2">
          <select
            v-model="selectedMonth"
            class="bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1.5 text-xs font-black text-slate-700 outline-none focus:ring-2 focus:ring-primary/20"
          >
            <option v-for="m in 12" :key="m" :value="m">
              {{ new Date(2024, m - 1, 1).toLocaleString('id-ID', { month: 'long' }) }}
            </option>
          </select>
          <select
            v-model="selectedYear"
            class="bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1.5 text-xs font-black text-slate-700 outline-none focus:ring-2 focus:ring-primary/20"
          >
            <option v-for="y in availableYears" :key="y" :value="y">{{ y }}</option>
          </select>
        </div>

        <div v-else-if="periodType === 'yearly'" class="flex items-center gap-2">
          <select
            v-model="selectedYear"
            class="bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1.5 text-xs font-black text-slate-700 outline-none focus:ring-2 focus:ring-primary/20"
          >
            <option v-for="y in availableYears" :key="y" :value="y">{{ y }}</option>
          </select>
        </div>

        <div v-else-if="periodType === 'custom'" class="flex items-center gap-2">
          <input
            type="date"
            v-model="selectedStartDate"
            class="bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1 text-xs font-black text-slate-700 outline-none focus:ring-2 focus:ring-primary/20"
          />
          <span class="text-xs font-bold text-slate-400">s/d</span>
          <input
            type="date"
            v-model="selectedEndDate"
            class="bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1 text-xs font-black text-slate-700 outline-none focus:ring-2 focus:ring-primary/20"
          />
        </div>

        <div v-else class="text-xs font-bold text-slate-500 px-2 uppercase tracking-wider hidden sm:block">
          {{ periodDisplayText }}
        </div>
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div
        class="bg-white p-6 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all"
      >
        <div class="flex items-start justify-between">
          <div>
            <p class="text-[13px] font-bold text-slate-500 uppercase tracking-wider">
              {{ salesmanLabel }}
            </p>
            <h3 class="text-3xl font-extrabold text-slate-900 mt-2">{{ totalSalesmen }}</h3>
          </div>
          <div class="p-3 bg-blue-50 rounded-2xl text-blue-600">
            <Users class="w-6 h-6" />
          </div>
        </div>
        <p class="text-xs font-bold text-slate-400 mt-4">Aktif di cabang pilihan</p>
      </div>

      <div
        class="bg-white p-6 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all"
      >
        <div class="flex items-start justify-between">
          <div>
            <p class="text-[13px] font-bold text-slate-500 uppercase tracking-wider">
              {{ omzetLabel }}
            </p>
            <h3
              class="text-2xl font-black text-slate-900 mt-2 truncate max-w-[180px]"
              :title="formatCurrency(kpiSummary.totalOmzetAll)"
            >
              {{ formatCurrency(kpiSummary.totalOmzetAll) }}
            </h3>
          </div>
          <div class="p-3 bg-emerald-50 rounded-2xl text-emerald-600">
            <Banknote class="w-6 h-6" />
          </div>
        </div>
        <div class="mt-4 flex items-center gap-2 text-emerald-600 font-bold text-xs">
          <TrendingUp class="w-3 h-3" /> {{ periodSubtitleBadge }}
        </div>
      </div>

      <div
        class="bg-white p-6 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all"
      >
        <div class="flex items-start justify-between">
          <div>
            <p class="text-[13px] font-bold text-slate-500 uppercase tracking-wider">
              {{ pendingLabel }}
            </p>
            <h3 class="text-3xl font-extrabold text-slate-900 mt-2">
              {{ masterDataStore.selectedSalesType === 'sales_mt' ? pendingTaxDocsCount : pendingNotesCount }}
            </h3>
          </div>
          <div class="p-3 bg-amber-50 rounded-2xl text-amber-600">
            <Target class="w-6 h-6" />
          </div>
        </div>
        <p class="text-xs font-bold text-amber-600 mt-4">Perlu approval dokumen</p>
      </div>

      <div
        class="bg-white p-6 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all"
      >
        <div class="flex items-start justify-between">
          <div>
            <p class="text-[13px] font-bold text-slate-500 uppercase tracking-wider">
              {{ campaignLabel }}
            </p>
            <h3 class="text-3xl font-extrabold text-slate-900 mt-2">
              {{ spandukStats.completed }}
            </h3>
          </div>
          <div class="p-3 bg-primary/5 rounded-2xl text-primary">
            <Megaphone class="w-6 h-6" />
          </div>
        </div>
      </div>
    </div>

    <!-- Store Distribution Map -->
    <div
      class="bg-white rounded-[32px] border border-slate-200 shadow-sm overflow-hidden animate-in fade-in slide-in-from-bottom-4 duration-700"
    >
      <div class="p-6 border-b border-slate-100 flex flex-col md:flex-row md:items-center justify-between gap-4 bg-slate-50/30">
        <div>
          <div class="flex items-center gap-2.5 flex-wrap">
            <h3 class="text-lg font-extrabold text-slate-900 tracking-tight flex items-center gap-2">
              <MapPin class="w-5 h-5 text-primary" />
              {{ mapTitle }}
            </h3>
            <span
              v-if="isFullDistributionLoaded"
              class="inline-flex items-center gap-1.5 bg-emerald-50 text-emerald-700 border border-emerald-200 text-[10px] font-extrabold px-2.5 py-0.5 rounded-full"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
              Semua Data Terpetakan
            </span>
            <span
              v-else-if="mapMode === 'STORES'"
              class="inline-flex items-center gap-1.5 bg-slate-100 text-slate-600 border border-slate-200 text-[10px] font-bold px-2.5 py-0.5 rounded-full"
            >
              Mode Cepat (On-Demand)
            </span>
          </div>
          <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-1">
            <span v-if="isFullDistributionLoaded">
              Total {{ stores.length }} {{ mapSubtitle }} <span v-if="mappedDistributionCount">({{ mappedDistributionCount }} titik koordinat valid)</span>
            </span>
            <span v-else>
              Klik tombol muat untuk memetakan sebaran titik toko
            </span>
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <!-- Trigger / Refresh All Stores Button -->
          <button
            v-if="mapMode === 'STORES'"
            @click="fetchStoreDistribution(true)"
            :disabled="isLoadingDistribution"
            class="px-3.5 py-1.5 bg-gradient-to-r from-primary to-primary-600 hover:opacity-95 text-white text-xs font-black rounded-xl shadow-xs hover:shadow-md transition-all flex items-center gap-1.5 cursor-pointer disabled:opacity-50"
            title="Muat atau perbarui seluruh data toko & koordinat wilayah"
          >
            <RefreshCw v-if="isLoadingDistribution" class="w-3.5 h-3.5 animate-spin" />
            <Sparkles v-else class="w-3.5 h-3.5 text-amber-300" />
            <span>{{ isFullDistributionLoaded ? 'Segarkan Peta Sebaran' : '⚡ Muat Peta Sebaran' }}</span>
          </button>

          <div class="flex items-center bg-white border border-slate-200 rounded-xl p-1 shadow-sm">
            <button
              @click="mapMode = 'STORES'"
              :class="[
                'px-4 py-1.5 rounded-lg text-xs font-black transition-all cursor-pointer',
                mapMode === 'STORES'
                  ? 'bg-primary text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-600',
              ]"
            >
              TOKO
            </button>
            <button
              @click="mapMode = 'PRODUCTS'"
              :class="[
                'px-4 py-1.5 rounded-lg text-xs font-black transition-all cursor-pointer',
                mapMode === 'PRODUCTS'
                  ? 'bg-emerald-500 text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-600',
              ]"
            >
              PRODUK
            </button>
          </div>

          <ProductMultiSelectDropdown
            v-if="mapMode === 'PRODUCTS'"
            v-model="selectedProductIds"
            :products="products"
            placeholder="Pilih Produk / Induk Tracking..."
            compact
          />

          <button
            v-if="isFullDistributionLoaded || mapMode === 'PRODUCTS'"
            @click="updateMarkers(true)"
            title="Reset Posisi Peta"
            class="p-2 bg-white border border-slate-200 text-slate-400 hover:text-primary rounded-xl transition-all shadow-sm cursor-pointer"
          >
            <Maximize2 class="w-4 h-4" />
          </button>

          <!-- Toggle Tile Layer: Peta vs Satelit -->
          <div class="flex items-center bg-white border border-slate-200 rounded-xl p-1 shadow-sm" title="Pilih tampilan peta">
            <button
              @click="tileLayerMode = 'STREET'"
              :class="[
                'px-3 py-1.5 rounded-lg text-xs font-black transition-all cursor-pointer',
                tileLayerMode === 'STREET'
                  ? 'bg-slate-700 text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-600',
              ]"
              title="Mode Peta Jalan (OpenStreetMap)"
            >
              🗺️ Peta
            </button>
            <button
              @click="tileLayerMode = 'SATELLITE'"
              :class="[
                'px-3 py-1.5 rounded-lg text-xs font-black transition-all cursor-pointer',
                tileLayerMode === 'SATELLITE'
                  ? 'bg-sky-600 text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-600',
              ]"
              title="Mode Satelit (Esri World Imagery)"
            >
              🛰️ Satelit
            </button>
          </div>
        </div>
      </div>
      <div class="relative">
        <div
          v-if="isMapLoading || isLoadingDistribution"
          class="absolute inset-0 bg-white/70 backdrop-blur-[2px] z-20 flex items-center justify-center"
        >
          <div class="flex flex-col items-center gap-3">
            <RefreshCw class="w-8 h-8 text-primary animate-spin" />
            <p class="text-xs font-black text-slate-900 uppercase tracking-widest">
              Memetakan Data Persebaran...
            </p>
          </div>
        </div>

        <!-- Placeholder On-Demand when not loaded yet in STORES mode -->
        <div
          v-if="!isFullDistributionLoaded && mapMode === 'STORES' && !isLoadingDistribution"
          class="absolute inset-0 z-10 flex flex-col items-center justify-center p-6 bg-slate-50/90 backdrop-blur-xs text-center"
        >
          <div class="w-16 h-16 rounded-3xl bg-primary/10 border border-primary/20 text-primary flex items-center justify-center mb-4 shadow-sm">
            <MapPin class="w-8 h-8 text-primary" />
          </div>
          <h4 class="text-base font-extrabold text-slate-800 tracking-tight">
            Peta Persebaran Toko Siap Dimuat
          </h4>
          <p class="text-xs text-slate-500 max-w-md mt-1.5 mb-5 font-medium leading-relaxed">
            Data titik toko tidak dimuat otomatis saat membuka halaman agar proses load dashboard instan dan ringan. Klik tombol di bawah untuk menampilkan seluruh sebaran outlet.
          </p>
          <button
            @click="fetchStoreDistribution(true)"
            class="px-5 py-2.5 bg-gradient-to-r from-primary to-primary-600 hover:opacity-95 text-white font-extrabold text-xs rounded-xl shadow-md hover:shadow-lg transition-all flex items-center gap-2 cursor-pointer"
          >
            <Sparkles class="w-4 h-4 text-amber-300" />
            <span>Tampilkan & Muat Peta Sebaran Toko</span>
          </button>
        </div>

        <div ref="mapContainer" class="h-[450px] w-full z-0"></div>
      </div>
    </div>

    <!-- Serapan Penjualan Wilayah -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden flex flex-col">
      <div class="p-6 border-b border-slate-100 bg-slate-50/50">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div>
            <div class="flex items-center gap-2.5 flex-wrap">
              <h3 class="text-lg font-extrabold text-slate-900 tracking-tight flex items-center gap-2">
                <MapPin class="w-5 h-5 text-primary" />
                Serapan Penjualan Wilayah
              </h3>
              <span
                v-if="isFullDistributionLoaded"
                class="inline-flex items-center gap-1.5 bg-emerald-50 text-emerald-700 border border-emerald-200 text-[10px] font-extrabold px-2.5 py-0.5 rounded-full"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                Dihitung dari 100% Toko ({{ stores.length }} Toko)
              </span>
            </div>
            <p class="text-sm text-slate-500 mt-1">Total sebaran toko per area administratif dari database resmi.</p>
          </div>

          <div v-if="selectedRegion" class="flex items-center gap-2 bg-primary/10 border border-primary/20 px-3 py-1.5 rounded-xl">
            <span class="text-xs font-bold text-primary">Filter Aktif: <strong>{{ selectedRegion }}</strong></span>
            <button 
              @click="selectedRegion = null; updateMarkers(true)" 
              class="text-xs font-black text-primary hover:text-red-500 transition-colors ml-1 cursor-pointer"
            >
              ✕ Reset
            </button>
          </div>
        </div>
        
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mt-5">
          <div class="flex items-center gap-3 flex-wrap">
            <div class="flex gap-1 bg-white p-1 rounded-xl border border-slate-200 shadow-xs">
              <button
                v-for="tab in ['KABUPATEN', 'KECAMATAN', 'DESA']"
                :key="tab"
                @click="activeRegionTab = tab as any; selectedRegion = null; updateMarkers(true)"
                :class="[
                  'px-4 py-2 rounded-lg text-xs font-black transition-all capitalize cursor-pointer',
                  activeRegionTab === tab ? 'bg-primary text-white shadow-sm' : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                {{ tab.toLowerCase() }}
              </button>
            </div>

            <!-- Search Wilayah -->
            <div class="relative">
              <input
                type="text"
                v-model="regionSearchQuery"
                :placeholder="`Cari ${activeRegionTab.toLowerCase()}...`"
                class="bg-white border border-slate-200 rounded-xl pl-3 pr-8 py-1.5 text-xs font-medium text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 transition-all w-[180px] sm:w-[220px]"
              />
              <button
                v-if="regionSearchQuery"
                @click="regionSearchQuery = ''"
                class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 text-xs cursor-pointer"
              >
                ✕
              </button>
            </div>
          </div>
          
          <div class="flex items-center gap-2 flex-wrap">
            <button
              @click="fetchStoreDistribution(true)"
              :disabled="isLoadingDistribution"
              class="px-3.5 py-2 bg-primary/10 hover:bg-primary/20 text-primary border border-primary/20 text-xs font-bold rounded-xl transition-all flex items-center gap-1.5 cursor-pointer disabled:opacity-50 shadow-xs"
              title="Perbarui pemetaan serapan wilayah dari seluruh toko"
            >
              <RefreshCw v-if="isLoadingDistribution" class="w-3.5 h-3.5 animate-spin" />
              <Sparkles v-else class="w-3.5 h-3.5" />
              <span>Perbarui Data Sebaran</span>
            </button>
            
            <button
              @click="syncUnknownRegions"
              :disabled="isSyncing"
              class="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-bold rounded-xl transition-all flex items-center gap-2 border border-slate-200 disabled:opacity-50 shadow-xs cursor-pointer"
            >
              <RefreshCw v-if="isSyncing" class="w-3.5 h-3.5 animate-spin text-primary" />
              {{ isSyncing ? 'Memulai Sync...' : 'Sync Wilayah Tidak Terdeteksi' }}
            </button>
          </div>
        </div>
      </div>
      
      <div class="p-6">
        <!-- Placeholder On-Demand saat sebaran toko belum dimuat -->
        <div
          v-if="!isFullDistributionLoaded && !isLoadingDistribution"
          class="py-12 flex flex-col items-center justify-center text-center p-6 border border-dashed border-slate-200 rounded-2xl bg-slate-50/50"
        >
          <div class="w-12 h-12 rounded-2xl bg-primary/10 text-primary flex items-center justify-center mb-3">
            <Sparkles class="w-6 h-6 text-primary" />
          </div>
          <h4 class="text-sm font-extrabold text-slate-800">Detail Serapan Toko Per Wilayah Belum Dimuat</h4>
          <p class="text-xs text-slate-500 max-w-md mt-1 mb-4 font-medium">
            Muat sebaran toko untuk menganalisis data distribusi berdasarkan tingkat kabupaten, kecamatan, hingga desa secara menyeluruh.
          </p>
          <button
            @click="fetchStoreDistribution(true)"
            class="px-4 py-2 bg-gradient-to-r from-primary to-primary-600 hover:opacity-95 text-white text-xs font-bold rounded-xl shadow-sm transition-all flex items-center gap-2 cursor-pointer"
          >
            <Sparkles class="w-3.5 h-3.5 text-amber-300" />
            <span>Muat Data Sebaran Wilayah</span>
          </button>
        </div>

        <div v-else-if="regionStats.length === 0" class="py-12 text-center text-slate-400 font-bold text-sm">
          Tidak ada data wilayah yang cocok dengan pencarian.
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-h-[420px] overflow-y-auto pr-1">
          <div 
            v-for="(stat, idx) in regionStats" 
            :key="idx"
            @click="handleRegionClick(stat.name)"
            :class="[
              'p-4 rounded-2xl border transition-all cursor-pointer flex items-center justify-between group',
              stat.name === 'Tidak Terdeteksi' 
                ? (selectedRegion === stat.name ? 'border-amber-500 bg-amber-500/10 ring-1 ring-amber-500' : 'border-dashed border-amber-300 bg-amber-50/40 hover:bg-amber-50')
                : (selectedRegion === stat.name ? 'border-primary bg-primary/5 ring-1 ring-primary shadow-sm' : 'border-slate-200 hover:border-primary/50 hover:bg-slate-50')
            ]"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div 
                :class="[
                  'w-8 h-8 rounded-xl flex items-center justify-center font-bold text-xs shrink-0',
                  stat.name === 'Tidak Terdeteksi'
                    ? 'bg-amber-100 text-amber-700'
                    : (selectedRegion === stat.name ? 'bg-primary text-white' : 'bg-slate-100 text-slate-600 group-hover:bg-primary/10 group-hover:text-primary')
                ]"
              >
                {{ idx + 1 }}
              </div>
              <div class="truncate">
                <p 
                  :class="[
                    'text-sm font-bold truncate',
                    stat.name === 'Tidak Terdeteksi' ? 'text-amber-800 italic' : 'text-slate-800'
                  ]"
                >
                  {{ stat.name }}
                </p>
                <p class="text-[11px] text-slate-400 font-medium">
                  {{ stat.percentage }}% dari total toko
                </p>
              </div>
            </div>
            <div class="flex flex-col items-end shrink-0 pl-3">
              <span 
                :class="[
                  'text-lg font-black',
                  stat.name === 'Tidak Terdeteksi' ? 'text-amber-600' : 'text-primary'
                ]"
              >
                {{ stat.count }}
              </span>
              <span class="text-[9px] font-black text-slate-400 uppercase tracking-tight">Toko</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Rekapitulasi Omzet & Penetrasi Wilayah (Kota / Kabupaten) -->
    <div id="city-penetration-section" class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden flex flex-col">
      <!-- Header -->
      <div class="p-6 border-b border-slate-100 flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-slate-50/50">
        <div>
          <div class="flex items-center gap-2.5">
            <div class="p-2 bg-primary/10 text-primary rounded-xl">
              <Building2 class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-lg font-extrabold text-slate-900 tracking-tight">
                Rekapitulasi Omzet & Penetrasi Wilayah (Kota / Kabupaten)
              </h3>
              <p class="text-xs font-semibold text-slate-500 mt-0.5">
                Monitoring kota yang sudah masuk vs belum masuk produk, serta penarikan data omzet per wilayah.
              </p>
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex flex-wrap items-center gap-2.5">
          <button
            @click="fetchCityRevenueReport"
            :disabled="isCityRevenueLoading"
            class="px-3.5 py-2 bg-white hover:bg-slate-50 text-slate-700 border border-slate-200 rounded-xl text-xs font-bold transition-all shadow-xs flex items-center gap-2 cursor-pointer disabled:opacity-50"
            title="Muat Ulang Data"
          >
            <RefreshCw class="w-3.5 h-3.5 text-slate-500" :class="{ 'animate-spin': isCityRevenueLoading }" />
            <span>Segarkan</span>
          </button>

          <button
            @click="exportCityRevenueExcel"
            :disabled="isExportingCityExcel || isCityRevenueLoading"
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 active:bg-emerald-800 text-white rounded-xl text-xs font-black transition-all shadow-sm flex items-center gap-2 cursor-pointer disabled:opacity-50"
          >
            <FileSpreadsheet v-if="!isExportingCityExcel" class="w-4 h-4 text-white" />
            <RefreshCw v-else class="w-4 h-4 animate-spin text-white" />
            <span>{{ isExportingCityExcel ? 'Mengunduh...' : 'Tarik Data Excel' }}</span>
          </button>
        </div>
      </div>

      <!-- Quick KPI Cards for City Penetration -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 p-6 border-b border-slate-100 bg-white">
        <div class="p-4 rounded-2xl bg-emerald-50/70 border border-emerald-100/80 flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-[11px] font-black text-emerald-800 uppercase tracking-wider">Kota Terpenetrasi</span>
            <span class="p-1.5 bg-emerald-100/80 rounded-lg text-emerald-700">
              <CheckCircle class="w-4 h-4" />
            </span>
          </div>
          <div class="mt-3 flex items-baseline gap-2">
            <span class="text-2xl font-black text-emerald-950">
              {{ cityRevenueData?.penetrated_count || 0 }}
            </span>
            <span class="text-[11px] font-bold text-emerald-700">
              ({{ cityRevenueData?.penetration_rate || 0 }}%)
            </span>
          </div>
          <p class="text-[10px] font-semibold text-emerald-600 mt-1">Sudah ada transaksi aktif</p>
        </div>

        <div class="p-4 rounded-2xl bg-slate-50 border border-slate-200/80 flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-[11px] font-black text-slate-600 uppercase tracking-wider">Kota Belum Masuk</span>
            <span class="p-1.5 bg-slate-200/70 rounded-lg text-slate-600">
              <XCircle class="w-4 h-4" />
            </span>
          </div>
          <div class="mt-3 flex items-baseline gap-2">
            <span class="text-2xl font-black text-slate-800">
              {{ cityRevenueData?.unpenetrated_count || 0 }}
            </span>
            <span class="text-[11px] font-bold text-slate-500">Kota</span>
          </div>
          <p class="text-[10px] font-semibold text-slate-500 mt-1">Peluang ekspansi baru</p>
        </div>

        <div class="p-4 rounded-2xl bg-blue-50/70 border border-blue-100/80 flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-[11px] font-black text-blue-800 uppercase tracking-wider">Total Omzet Wilayah</span>
            <span class="p-1.5 bg-blue-100/80 rounded-lg text-blue-700">
              <Banknote class="w-4 h-4" />
            </span>
          </div>
          <div class="mt-3">
            <span class="text-xl font-black text-blue-950 truncate block" :title="formatCurrency(cityRevenueData?.total_revenue || 0)">
              {{ formatCurrency(cityRevenueData?.total_revenue || 0) }}
            </span>
          </div>
          <p class="text-[10px] font-semibold text-blue-600 mt-1">{{ cityRevenueData?.total_quantity || 0 }} Pcs Terjual</p>
        </div>

        <div class="p-4 rounded-2xl bg-amber-50/70 border border-amber-100/80 flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-[11px] font-black text-amber-800 uppercase tracking-wider">Top Kota Kontributor</span>
            <span class="p-1.5 bg-amber-100/80 rounded-lg text-amber-700">
              <Trophy class="w-4 h-4" />
            </span>
          </div>
          <div class="mt-3">
            <span class="text-lg font-black text-amber-950 truncate block" :title="cityRevenueData?.top_city_name || '-'">
              {{ cityRevenueData?.top_city_name || '-' }}
            </span>
          </div>
          <p class="text-[10px] font-semibold text-amber-700 mt-1">
            {{ formatCurrency(cityRevenueData?.top_city_revenue || 0) }}
          </p>
        </div>
      </div>

      <!-- Filters & Search Bar -->
      <div class="p-4 bg-slate-50/70 border-b border-slate-100 flex flex-wrap items-center justify-between gap-3">
        <!-- Status Penetrasi Tabs -->
        <div class="flex items-center bg-white p-1 rounded-xl border border-slate-200 shadow-2xs">
          <button
            @click="cityPenetrationFilter = 'ALL'"
            :class="[
              'px-3 py-1.5 rounded-lg text-xs font-black transition-all cursor-pointer',
              cityPenetrationFilter === 'ALL'
                ? 'bg-primary text-white shadow-xs'
                : 'text-slate-600 hover:text-slate-900'
            ]"
          >
            Semua ({{ cityRevenueData?.total_monitored_cities || 0 }})
          </button>
          <button
            @click="cityPenetrationFilter = 'PENETRATED'"
            :class="[
              'px-3 py-1.5 rounded-lg text-xs font-black transition-all cursor-pointer flex items-center gap-1.5',
              cityPenetrationFilter === 'PENETRATED'
                ? 'bg-emerald-600 text-white shadow-xs'
                : 'text-emerald-700 hover:text-emerald-900'
            ]"
          >
            <CheckCircle class="w-3.5 h-3.5" />
            Sudah Masuk ({{ cityRevenueData?.penetrated_count || 0 }})
          </button>
          <button
            @click="cityPenetrationFilter = 'UNPENETRATED'"
            :class="[
              'px-3 py-1.5 rounded-lg text-xs font-black transition-all cursor-pointer flex items-center gap-1.5',
              cityPenetrationFilter === 'UNPENETRATED'
                ? 'bg-slate-700 text-white shadow-xs'
                : 'text-slate-600 hover:text-slate-900'
            ]"
          >
            <XCircle class="w-3.5 h-3.5" />
            Belum Masuk ({{ cityRevenueData?.unpenetrated_count || 0 }})
          </button>
        </div>

        <!-- Right Side: Filter Produk & Search -->
        <div class="flex flex-wrap items-center gap-2.5 flex-1 max-w-2xl justify-end">
          <!-- Filter Produk Multi-Select & Induk -->
          <ProductMultiSelectDropdown
            v-model="selectedCityProducts"
            :products="products"
            placeholder="Semua Produk & Induk"
            compact
          />

          <!-- Search Input -->
          <div class="relative min-w-[200px] flex-1">
            <Search class="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              v-model="cityRevenueSearch"
              type="text"
              placeholder="Cari kota / provinsi / salesman..."
              class="w-full bg-white pl-8 pr-3 py-1.5 border border-slate-200 rounded-xl text-xs font-semibold text-slate-700 placeholder-slate-400 outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>
        </div>
      </div>

      <!-- Table Content -->
      <div class="overflow-x-auto min-h-[280px] relative">
        <div v-if="isCityRevenueLoading" class="absolute inset-0 bg-white/70 backdrop-blur-2xs flex items-center justify-center z-10">
          <div class="flex items-center gap-2.5 text-xs font-bold text-slate-600 bg-white px-4 py-2.5 rounded-2xl shadow-md border border-slate-100">
            <RefreshCw class="w-4 h-4 text-primary animate-spin" />
            Memuat data persebaran omzet kota...
          </div>
        </div>

        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50/80 text-slate-500 font-extrabold uppercase text-[10px] tracking-wider border-b border-slate-200/80">
            <tr>
              <th class="py-3 px-4 w-12 text-center">#</th>
              <th class="py-3 px-4">Kota / Kabupaten</th>
              <th class="py-3 px-4 text-center">Status Penetrasi</th>
              <th class="py-3 px-4 text-center">Total Toko</th>
              <th class="py-3 px-4 text-center">Toko Aktif (EC)</th>
              <th class="py-3 px-4 text-right">Volume</th>
              <th class="py-3 px-4 text-right">Total Omzet</th>
              <th class="py-3 px-4 min-w-[140px]">Kontribusi (%)</th>
              <th class="py-3 px-4">Top Salesman</th>
              <th class="py-3 px-4">Top Produk</th>
              <th class="py-3 px-4 text-center w-24">Peta</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700 font-medium">
            <tr v-if="filteredCityItems.length === 0">
              <td colspan="11" class="py-12 text-center text-slate-400 font-bold">
                <div class="flex flex-col items-center justify-center gap-2">
                  <MapPin class="w-8 h-8 text-slate-300" />
                  <p>Tidak ada data kota yang sesuai dengan filter atau pencarian.</p>
                </div>
              </td>
            </tr>

            <tr
              v-for="(item, idx) in filteredCityItems"
              :key="item.city"
              @click="handleCityRowClick(item.city)"
              class="hover:bg-slate-50/80 transition-colors group cursor-pointer"
            >
              <td class="py-3 px-4 text-center font-bold text-slate-400 group-hover:text-slate-700">
                {{ Number(idx) + 1 }}
              </td>

              <!-- Kota & Provinsi -->
              <td class="py-3 px-4">
                <div class="font-extrabold text-slate-900 group-hover:text-primary transition-colors flex items-center gap-1.5">
                  {{ item.city }}
                </div>
                <div class="text-[10px] font-semibold text-slate-400 mt-0.5 flex items-center gap-1">
                  <span>{{ item.province || '-' }}</span>
                  <span v-if="item.branch_name" class="text-slate-300">•</span>
                  <span v-if="item.branch_name" class="text-slate-500">{{ item.branch_name }}</span>
                </div>
              </td>

              <!-- Status Penetrasi -->
              <td class="py-3 px-4 text-center">
                <span
                  v-if="item.status === 'PENETRATED'"
                  class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[10px] font-black bg-emerald-50 text-emerald-700 border border-emerald-200"
                >
                  <CheckCircle class="w-3 h-3 text-emerald-600" />
                  Sudah Masuk
                </span>
                <span
                  v-else
                  class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[10px] font-black bg-slate-100 text-slate-500 border border-slate-200"
                >
                  <XCircle class="w-3 h-3 text-slate-400" />
                  Belum Masuk
                </span>
              </td>

              <!-- Total Toko -->
              <td class="py-3 px-4 text-center font-bold text-slate-700">
                {{ item.total_stores }}
              </td>

              <!-- Toko Aktif (EC) -->
              <td class="py-3 px-4 text-center">
                <span 
                  :class="[
                    'font-black px-2 py-0.5 rounded-md text-xs',
                    item.active_stores_count > 0 ? 'bg-blue-50 text-blue-700' : 'text-slate-400'
                  ]"
                >
                  {{ item.active_stores_count }}
                </span>
              </td>

              <!-- Volume -->
              <td class="py-3 px-4 text-right font-bold text-slate-700">
                {{ item.total_quantity.toLocaleString('id-ID') }} <span class="text-[10px] text-slate-400">pcs</span>
              </td>

              <!-- Total Omzet -->
              <td class="py-3 px-4 text-right">
                <span :class="['font-black', item.total_revenue > 0 ? 'text-slate-900' : 'text-slate-400']">
                  {{ formatCurrency(item.total_revenue) }}
                </span>
              </td>

              <!-- Kontribusi Omzet % -->
              <td class="py-3 px-4">
                <div class="flex items-center gap-2">
                  <div class="flex-1 h-2 bg-slate-100 rounded-full overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all duration-300"
                      :class="item.total_revenue > 0 ? 'bg-primary' : 'bg-transparent'"
                      :style="`width: ${Math.min(item.revenue_share_pct, 100)}%`"
                    ></div>
                  </div>
                  <span class="text-[11px] font-extrabold text-slate-700 w-10 text-right">
                    {{ item.revenue_share_pct }}%
                  </span>
                </div>
              </td>

              <!-- Top Salesman -->
              <td class="py-3 px-4 text-slate-700 truncate max-w-[130px]" :title="item.top_salesman_name">
                <span :class="item.top_salesman_name !== '-' ? 'font-bold text-slate-800' : 'text-slate-400'">
                  {{ item.top_salesman_name }}
                </span>
              </td>

              <!-- Top Produk -->
              <td class="py-3 px-4 text-slate-700 truncate max-w-[140px]" :title="item.top_product_name">
                <span :class="item.top_product_name !== '-' ? 'font-bold text-slate-800' : 'text-slate-400'">
                  {{ item.top_product_name }}
                </span>
              </td>

              <!-- Action button: Zoom to map -->
              <td class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click.stop="handleCityRowClick(item.city)"
                  class="p-1.5 bg-slate-100 hover:bg-primary hover:text-white text-slate-600 rounded-lg transition-colors cursor-pointer"
                  title="Fokuskan peta ke kota ini"
                >
                  <MapPin class="w-3.5 h-3.5" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Footer Info -->
      <div class="p-4 bg-slate-50 border-t border-slate-100 text-xs font-semibold text-slate-500 flex flex-col sm:flex-row items-center justify-between gap-2">
        <p>
          Menampilkan <strong class="text-slate-800">{{ filteredCityItems.length }}</strong> dari <strong class="text-slate-800">{{ cityRevenueData?.total_monitored_cities || 0 }}</strong> kota/kabupaten terpantau.
        </p>
        <p class="text-[11px] text-slate-400">
          💡 Klik pada baris kota untuk otomatis memfilter dan memfokuskan peta di atas ke wilayah tersebut.
        </p>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <div
        class="lg:col-span-2 bg-white rounded-3xl border border-slate-200 shadow-sm p-8 flex flex-col h-[480px]"
      >
        <div class="flex items-center justify-between mb-8 shrink-0">
          <div>
            <h3 class="text-lg font-extrabold text-slate-900 tracking-tight">
              {{ chartOmzetTitle }}
            </h3>
            <p class="text-sm text-slate-500">Target vs Aktual ({{ periodDisplayText }})</p>
          </div>
        </div>
        <div class="flex-1 min-h-0 relative">
          <v-chart class="absolute inset-0" :option="omzetChartOption" autoresize />
        </div>
      </div>

      <div
        class="bg-white rounded-3xl border border-slate-200 shadow-sm p-8 flex flex-col h-[480px]"
      >
        <h3 class="text-lg font-extrabold text-slate-900 tracking-tight mb-2 shrink-0">
          {{ campaignLabel }}
        </h3>
        <p class="text-sm text-slate-500 mb-8 shrink-0">
          {{ masterDataStore.selectedSalesType === 'sales_mt' ? 'Progress sewa display & promosi' : 'Progress pemasangan materi promosi' }}
        </p>
        <div class="relative flex-1 min-h-0 flex items-center justify-center">
          <v-chart class="h-full w-full" :option="spandukChartOption" autoresize />
          <div
            class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none"
          >
            <span class="text-3xl font-black text-slate-900">{{ spandukStats.completed }}</span>
            <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-1"
              >Titik Selesai</span
            >
          </div>
        </div>
        <div class="grid grid-cols-3 gap-4 mt-8 shrink-0 border-t border-slate-50 pt-6">
          <div class="text-center">
            <p class="text-[10px] font-bold text-slate-400 uppercase">Selesai</p>
            <p class="text-lg font-black text-emerald-600">{{ spandukStats.completed }}</p>
          </div>
          <div class="text-center">
            <p class="text-[10px] font-bold text-slate-400 uppercase">Proses</p>
            <p class="text-lg font-black text-blue-600">{{ spandukStats.inProgress }}</p>
          </div>
          <div class="text-center">
            <p class="text-[10px] font-bold text-slate-400 uppercase">Pending</p>
            <p class="text-lg font-black text-amber-600">{{ spandukStats.pending }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Leaderboard & Recent Activity -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <div
        class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden flex flex-col"
      >
        <div class="p-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div class="flex items-center gap-2 bg-slate-100/80 p-1 rounded-xl">
            <button
              @click="activeRankingTab = 'SALESMAN'"
              :class="[
                'px-4 py-1.5 rounded-lg text-sm font-black transition-all flex items-center gap-2',
                activeRankingTab === 'SALESMAN'
                  ? 'bg-white text-primary shadow-sm'
                  : 'text-slate-500 hover:text-slate-700',
              ]"
            >
              <Users class="w-4 h-4" /> {{ masterDataStore.selectedSalesType === 'sales_mt' ? 'Sales MT & KAM' : 'Salesman' }}
            </button>
            <button
              @click="activeRankingTab = 'CABANG'"
              :class="[
                'px-4 py-1.5 rounded-lg text-sm font-black transition-all flex items-center gap-2',
                activeRankingTab === 'CABANG'
                  ? 'bg-white text-primary shadow-sm'
                  : 'text-slate-500 hover:text-slate-700',
              ]"
            >
              <MapPin class="w-4 h-4" /> Cabang
            </button>
          </div>
          <router-link
            to="/sales/salesman"
            class="text-xs font-bold text-primary flex items-center gap-1.5 hover:underline"
          >
            Manajemen Salesman <ArrowRight class="w-3.5 h-3.5" />
          </router-link>
        </div>

        <!-- Salesman Rankings -->
        <div v-if="activeRankingTab === 'SALESMAN'" class="flex-1 flex flex-col">
          <div v-if="allSalesmanRankings.length === 0" class="p-12 text-center text-slate-400 font-bold text-sm">
            Belum ada data pencapaian omzet salesman pada periode ini.
          </div>
          <div v-else class="divide-y divide-slate-100 flex-1">
            <div
              v-for="sales in topSalesman"
              :key="sales.id"
              @click="handleSalesmanRowClick(sales)"
              class="p-4 flex items-center gap-4 hover:bg-slate-50/80 transition-all group cursor-pointer"
              :title="`Klik untuk memfilter dan memfokuskan data area ${sales.name}`"
            >
              <!-- Rank Medal / Badge -->
              <div
                :class="[
                  'w-10 h-10 rounded-2xl flex items-center justify-center font-black text-sm shrink-0 shadow-xs transition-transform group-hover:scale-105',
                  sales.rank === 1
                    ? 'bg-amber-100 text-amber-700 border border-amber-300 ring-2 ring-amber-400/20'
                    : sales.rank === 2
                    ? 'bg-slate-200 text-slate-700 border border-slate-300'
                    : sales.rank === 3
                    ? 'bg-amber-700/10 text-amber-800 border border-amber-600/30'
                    : 'bg-slate-100 text-slate-500'
                ]"
              >
                <Trophy v-if="sales.rank === 1" class="w-5 h-5 text-amber-600" />
                <Medal v-else-if="sales.rank <= 3" class="w-5 h-5" />
                <span v-else>#{{ sales.rank }}</span>
              </div>

              <!-- Name & Progress -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between gap-2">
                  <p class="font-bold text-slate-900 truncate text-sm group-hover:text-primary transition-colors">
                    {{ sales.name }}
                  </p>
                  <span 
                    v-if="sales.hasTarget"
                    :class="[
                      'text-[10px] font-black px-2 py-0.5 rounded-md shrink-0',
                      sales.achievement >= 100 
                        ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' 
                        : sales.achievement >= 75
                        ? 'bg-blue-50 text-blue-700 border border-blue-200'
                        : 'bg-amber-50 text-amber-700 border border-amber-200'
                    ]"
                  >
                    {{ sales.achievement }}%
                  </span>
                  <span 
                    v-else
                    class="text-[9px] font-black px-2 py-0.5 rounded-md shrink-0 bg-slate-100 text-slate-600 border border-slate-200"
                  >
                    Omzet Berjalan
                  </span>
                </div>

                <!-- Area / Wilayah Penjualan Badge -->
                <div class="flex items-center gap-1.5 mt-0.5 text-[11px] text-slate-500">
                  <MapPin class="w-3 h-3 text-slate-400 shrink-0" />
                  <span class="truncate font-semibold text-slate-600" :title="getSalesmanArea(sales)">
                    {{ getSalesmanArea(sales) }}
                  </span>
                  <span v-if="sales.branchName && !getSalesmanArea(sales).includes(sales.branchName)" class="text-[10px] text-slate-400 shrink-0 font-medium">
                    ({{ sales.branchName }})
                  </span>
                </div>

                <!-- Target vs Aktual Subtitle -->
                <div class="flex items-center justify-between text-[11px] text-slate-400 mt-1">
                  <span class="truncate">
                    Target: {{ sales.hasTarget ? formatCompactRupiah(sales.target) : 'Belum diset' }}
                  </span>
                  <span class="font-bold text-slate-700">{{ formatCurrency(sales.omzet) }}</span>
                </div>

                <!-- Progress Bar -->
                <div class="h-1.5 w-full bg-slate-100 rounded-full overflow-hidden mt-1.5">
                  <div
                    class="h-full rounded-full transition-all duration-500"
                    :class="[
                      !sales.hasTarget
                        ? 'bg-blue-500'
                        : (sales.achievement >= 100 
                            ? 'bg-emerald-500' 
                            : sales.achievement >= 75 
                            ? 'bg-blue-500' 
                            : 'bg-primary')
                    ]"
                    :style="`width: ${sales.relativeProgress}%`"
                  ></div>
                </div>
              </div>
            </div>
          </div>

          <!-- Toggle Expand Button -->
          <div 
            v-if="allSalesmanRankings.length > 5" 
            class="p-3 bg-slate-50 border-t border-slate-100 text-center"
          >
            <button
              @click="isRankingsExpanded = !isRankingsExpanded"
              class="text-xs font-black text-primary hover:text-primary/80 transition-colors inline-flex items-center gap-1 cursor-pointer"
            >
              {{ isRankingsExpanded ? 'Ciutkan Tampilan (Top 5)' : `Lihat Semua (${allSalesmanRankings.length} Salesman)` }}
              <ChevronRight 
                class="w-3.5 h-3.5 transition-transform" 
                :class="{ 'rotate-90': isRankingsExpanded }" 
              />
            </button>
          </div>
        </div>

        <!-- Branch Rankings -->
        <div v-if="activeRankingTab === 'CABANG'" class="flex-1 flex flex-col">
          <div v-if="allBranchRankings.length === 0" class="p-12 text-center text-slate-400 font-bold text-sm">
            Belum ada data pencapaian omzet cabang pada periode ini.
          </div>
          <div v-else class="divide-y divide-slate-100 flex-1">
            <div
              v-for="branch in topBranches"
              :key="branch.id"
              class="p-4 flex items-center gap-4 hover:bg-slate-50/60 transition-all group"
            >
              <div
                :class="[
                  'w-10 h-10 rounded-2xl flex items-center justify-center font-black text-sm shrink-0 shadow-xs transition-transform group-hover:scale-105',
                  branch.rank === 1
                    ? 'bg-emerald-100 text-emerald-700 border border-emerald-300 ring-2 ring-emerald-400/20'
                    : branch.rank === 2
                    ? 'bg-slate-200 text-slate-700 border border-slate-300'
                    : branch.rank === 3
                    ? 'bg-amber-700/10 text-amber-800 border border-amber-600/30'
                    : 'bg-slate-100 text-slate-500',
                ]"
              >
                <Trophy v-if="branch.rank === 1" class="w-5 h-5 text-emerald-600" />
                <Medal v-else-if="branch.rank <= 3" class="w-5 h-5" />
                <span v-else>#{{ branch.rank }}</span>
              </div>

              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between gap-2">
                  <p class="font-bold text-slate-900 truncate text-sm">{{ branch.name }}</p>
                  <span 
                    v-if="branch.hasTarget"
                    :class="[
                      'text-[10px] font-black px-2 py-0.5 rounded-md shrink-0',
                      branch.achievement >= 100 
                        ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' 
                        : branch.achievement >= 75
                        ? 'bg-blue-50 text-blue-700 border border-blue-200'
                        : 'bg-amber-50 text-amber-700 border border-amber-200'
                    ]"
                  >
                    {{ branch.achievement }}%
                  </span>
                  <span 
                    v-else
                    class="text-[9px] font-black px-2 py-0.5 rounded-md shrink-0 bg-slate-100 text-slate-600 border border-slate-200"
                  >
                    Omzet Berjalan
                  </span>
                </div>

                <div class="flex items-center justify-between text-[11px] text-slate-400 mt-0.5">
                  <span class="truncate">
                    Target: {{ branch.hasTarget ? formatCompactRupiah(branch.target) : 'Belum diset' }}
                  </span>
                  <span class="font-bold text-slate-700">{{ formatCurrency(branch.omzet) }}</span>
                </div>

                <div class="h-1.5 w-full bg-slate-100 rounded-full overflow-hidden mt-1.5">
                  <div
                    class="h-full rounded-full transition-all duration-500"
                    :class="[
                      !branch.hasTarget
                        ? 'bg-blue-500'
                        : (branch.achievement >= 100 
                            ? 'bg-emerald-500' 
                            : branch.achievement >= 75 
                            ? 'bg-blue-500' 
                            : 'bg-primary')
                    ]"
                    :style="`width: ${branch.relativeProgress}%`"
                  ></div>
                </div>
              </div>
            </div>
          </div>

          <!-- Toggle Expand Button for Branches -->
          <div 
            v-if="allBranchRankings.length > 5" 
            class="p-3 bg-slate-50 border-t border-slate-100 text-center"
          >
            <button
              @click="isRankingsExpanded = !isRankingsExpanded"
              class="text-xs font-black text-primary hover:text-primary/80 transition-colors inline-flex items-center gap-1 cursor-pointer"
            >
              {{ isRankingsExpanded ? 'Ciutkan Tampilan (Top 5)' : `Lihat Semua (${allBranchRankings.length} Cabang)` }}
              <ChevronRight 
                class="w-3.5 h-3.5 transition-transform" 
                :class="{ 'rotate-90': isRankingsExpanded }" 
              />
            </button>
          </div>
        </div>
      </div>

      <div
        class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden flex flex-col"
      >
        <div class="p-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div class="flex items-center gap-2 bg-slate-100/80 p-1 rounded-xl">
            <button
              @click="activeInsightTab = 'WILAYAH'"
              :class="[
                'px-4 py-1.5 rounded-lg text-sm font-black transition-all flex items-center gap-2',
                activeInsightTab === 'WILAYAH'
                  ? 'bg-white text-primary shadow-sm'
                  : 'text-slate-500 hover:text-slate-700',
              ]"
            >
              <MapPin class="w-4 h-4" /> Wilayah
            </button>
            <button
              @click="activeInsightTab = 'PRODUK'"
              :class="[
                'px-4 py-1.5 rounded-lg text-sm font-black transition-all flex items-center gap-2',
                activeInsightTab === 'PRODUK'
                  ? 'bg-white text-primary shadow-sm'
                  : 'text-slate-500 hover:text-slate-700',
              ]"
            >
              <Package class="w-4 h-4" /> Produk
            </button>
          </div>
        </div>

        <!-- Persebaran Wilayah -->
        <div
          v-if="activeInsightTab === 'WILAYAH'"
          class="divide-y divide-slate-100 flex-1 overflow-auto max-h-[360px]"
        >
          <div v-if="!isFullDistributionLoaded && !isLoadingDistribution" class="p-8 text-center text-slate-500 font-medium text-xs flex flex-col items-center gap-3">
            <p>Data persebaran toko belum dimuat.</p>
            <button
              @click="fetchStoreDistribution(true)"
              class="px-3.5 py-1.5 bg-primary/10 hover:bg-primary/20 text-primary font-bold rounded-xl text-xs transition-colors cursor-pointer flex items-center gap-1.5"
            >
              <Sparkles class="w-3.5 h-3.5" />
              <span>⚡ Muat Sebaran Toko</span>
            </button>
          </div>
          <div v-else-if="regionCounts.length === 0" class="p-8 text-center text-slate-400 font-bold">
            Belum ada data persebaran toko
          </div>
          <div
            v-else
            v-for="(region, index) in regionCounts"
            :key="region.name"
            class="p-4 flex items-center gap-4 hover:bg-slate-50/50 transition-all"
          >
            <div
              :class="[
                'w-10 h-10 rounded-2xl flex items-center justify-center font-black text-sm',
                region.bg,
                region.text,
              ]"
            >
              {{ index + 1 }}
            </div>
            <div class="flex-1">
              <p class="font-bold text-slate-900">{{ region.name }}</p>
              <div class="flex items-center gap-2 mt-0.5">
                <div class="h-1.5 flex-1 bg-slate-100 rounded-full overflow-hidden">
                  <div
                    class="h-full rounded-full"
                    :class="region.color"
                    :style="`width: ${Math.min((region.value / (stores.length || 1)) * 100, 100)}%`"
                  ></div>
                </div>
              </div>
            </div>
            <div class="text-right">
              <p class="text-sm font-black text-slate-900">{{ region.value }}</p>
              <p class="text-[10px] font-bold text-slate-400 uppercase">Toko</p>
            </div>
          </div>
        </div>

        <!-- Produk Terlaris -->
        <div
          v-if="activeInsightTab === 'PRODUK'"
          class="divide-y divide-slate-100 flex-1 overflow-auto max-h-[360px]"
        >
          <div v-if="topProducts.length === 0" class="p-8 text-center text-slate-400 font-bold">
            Belum ada data penjualan produk
          </div>
          <div
            v-for="(prod, index) in topProducts"
            :key="prod.id"
            class="p-4 flex items-center gap-4 hover:bg-slate-50/50 transition-all"
          >
            <div
              class="w-10 h-10 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center font-black text-sm border border-amber-100 shadow-sm"
            >
              #{{ index + 1 }}
            </div>
            <div class="flex-1">
              <p class="font-bold text-slate-900 line-clamp-1">{{ prod.name }}</p>
              <p class="text-[11px] font-black text-slate-400 uppercase mt-0.5 tracking-wider">
                Top Product
              </p>
            </div>
            <div class="text-right">
              <p class="text-sm font-black text-slate-900">{{ prod.total_sold }}</p>
              <p class="text-[10px] font-bold text-slate-400 uppercase">Terjual (Pcs)</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.v-chart {
  font-family: 'Plus Jakarta Sans', sans-serif !important;
}
</style>
