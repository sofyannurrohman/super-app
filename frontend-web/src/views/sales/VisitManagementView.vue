<script setup lang="ts">
import { ref, shallowRef, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useSalesVisitStore, type SalesVisit } from '@/stores/salesVisit'
import { useMasterDataStore } from '@/stores/masterData'

const props = withDefaults(
  defineProps<{
    isModernTrade?: boolean
  }>(),
  {
    isModernTrade: false,
  }
)

const route = useRoute()
const isMT = computed(() => props.isModernTrade || route.path.includes('modern-trade') || masterStore.selectedSalesType === 'sales_mt')
import { toast } from 'vue-sonner'
import { onClickOutside } from '@vueuse/core'
import {
  MapPin,
  Calendar,
  User,
  Search,
  Filter,
  RefreshCw,
  Camera,
  ExternalLink,
  Clock,
  CheckCircle2,
  AlertCircle,
  FileText,
  Maximize2,
  Plus,
  Trash2,
  Edit,
  X,
  XCircle,
  Map as MapIcon,
  List as ListIcon,
  Wallet as WalletIcon,
  CheckCircle,
  Clock as ClockIcon,
  Download,
  Store,
  TrendingUp,
  Users,
  ShoppingBag,
  Gauge,
  Fuel,
  Navigation2,
  Layers,
  Image as ImageIcon,
  Sparkles,
  Bike,
  ClipboardCheck,
  Zap,
  Check,
  ChevronsUpDown,
  ChevronUp,
  ChevronDown,
  ArrowRight,
  Compass,
  Route as RouteIcon,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import apiClient from '@/api/axios'
import * as XLSX from 'xlsx'
import ExcelJS from 'exceljs'
import { saveAs } from 'file-saver'

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

const formatTime = (d: Date | string) => {
  const date = typeof d === 'string' ? new Date(d) : d
  return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}
const formatFullTime = (d: Date | string) => {
  const date = typeof d === 'string' ? new Date(d) : d
  return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
const formatDate = (d: Date | string) => {
  const date = typeof d === 'string' ? new Date(d) : d
  return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}
const formatDateISO = (d: Date) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatRupiah = (val: number | string | null | undefined): string => {
  const num = typeof val === 'string' ? parseFloat(val) : Number(val)
  if (isNaN(num) || num === null || num === undefined) return '0'
  return new Intl.NumberFormat('id-ID').format(num)
}

const getNominalFontSizeClass = (val: number | string | null | undefined): string => {
  const str = formatRupiah(val)
  if (str.length > 14) return 'text-xs sm:text-sm md:text-base' // Ratusan juta ke atas
  if (str.length > 9) return 'text-sm sm:text-base md:text-lg'  // Puluhan juta
  return 'text-base sm:text-lg md:text-xl'
}

const visitStore = useSalesVisitStore()
const masterStore = useMasterDataStore()
const toRolePattern = /(^|[^a-z0-9])(to|take order|task order)(?=[^a-z0-9]|$)/i
let visitsAbortController: AbortController | null = null
let transactionsAbortController: AbortController | null = null
let rekapAbortController: AbortController | null = null
let plansAbortController: AbortController | null = null

const salesmen = computed(() => {
  return masterStore.employees.filter((emp: any) => {
    // 0. Filter by active status
    const isActive = emp.user?.is_active !== false && emp.employment_status !== 'RESIGNED' && emp.employment_status !== 'TERMINATED'

    // 1. Filter by branch
    const branchId = masterStore.selectedBranchId
    const empBranchId = emp.operating_branch_id || emp.branch_id
    const belongsToBranch = !branchId || branchId === 'ALL_BRANCHES' || empBranchId === branchId
    
    // 2. Filter by job title (sales / motoris / booster / marketing)
    const jobTitle = (emp.job_position?.title || '').toLowerCase()
    const isSales = /\b(sales|motoris|booster|marketing|to|mt|kam|take order|modern trade)\b/i.test(jobTitle)

    // 3. Filter by Sales Type
    let matchSalesType = true
    const type = masterStore.selectedSalesType
    if (type === 'sales_mt') {
      matchSalesType = /\b(mt|modern trade|kam|key account)\b/i.test(jobTitle)
    } else if (type === 'sales_to') {
      matchSalesType = toRolePattern.test(jobTitle)
    } else if (type === 'sales_motoris_booster') {
      matchSalesType = /\b(motoris|booster)\b/i.test(jobTitle)
    }
    
    return isActive && belongsToBranch && isSales && matchSalesType
  })
})

const tzOffset = (new Date()).getTimezoneOffset() * 60000
const todayLocal = (new Date(Date.now() - tzOffset)).toISOString().split('T')[0]

const activeTab = ref(route.query.tab ? String(route.query.tab) : 'table')
const searchQuery = ref('')
const selectedEmployee = ref('ALL')
const selectedDate = ref(todayLocal) // Default to today
const selectedMonth = ref('') // Default empty

const isLoading = ref(false)

const filterMode = ref<'date' | 'month'>('date')

const currentPage = ref(1)
const itemsPerPage = ref(50)

const effectiveVisitTotal = computed(() => {
  if ((activeTab.value === 'map' || itemsPerPage.value === -1) && selectedEmployee.value !== 'ALL') {
    return filteredVisits.value.length
  }
  return visitStore.total
})

const totalPages = computed(() => {
  if (itemsPerPage.value === -1) return 1
  return Math.ceil(effectiveVisitTotal.value / itemsPerPage.value) || 1
})

const visiblePages = computed(() => {
  const pages = []
  const maxVisible = 5
  let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2))
  let end = Math.min(totalPages.value, start + maxVisible - 1)

  if (end - start + 1 < maxVisible) {
    start = Math.max(1, end - maxVisible + 1)
  }

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  return pages
})

const paginationInfo = computed(() => {
  const total = effectiveVisitTotal.value
  if (total === 0) return 'Menampilkan 0 sampai 0 dari 0 entri'
  if (itemsPerPage.value === -1) {
    return `Menampilkan seluruh ${total} entri`
  }
  const start = (currentPage.value - 1) * itemsPerPage.value + 1
  const end = Math.min(currentPage.value * itemsPerPage.value, total)
  return `Menampilkan ${start} sampai ${end} dari ${total} entri`
})

const onItemsPerPageChange = () => {
  currentPage.value = 1
  fetchVisits()
}

const goToPage = (page: number) => {
  currentPage.value = page
  fetchVisits()
}

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchVisits()
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    fetchVisits()
  }
}

const getPeriodContextKey = () => {
  const mode = filterMode.value
  const periodVal = mode === 'date' ? (selectedDate.value || todayLocal) : (selectedMonth.value || '')
  const branchVal = masterStore.selectedBranchId || 'ALL_BRANCHES'
  const salesTypeVal = masterStore.selectedSalesType || 'ALL_SALES'
  return `${mode}|${periodVal}|${branchVal}|${salesTypeVal}`
}

let lastTransactionsContextKey = ''
let lastRekapVisitsContextKey = ''
let lastRekapDataContextKey = ''
const visitsCache = new Map<string, { visits: SalesVisit[]; total: number }>()
const visitPlansCache = new Map<string, any[]>()

const invalidateVisitCaches = () => {
  lastTransactionsContextKey = ''
  lastRekapVisitsContextKey = ''
  lastRekapDataContextKey = ''
  visitsCache.clear()
  visitPlansCache.clear()
}

const fetchActiveTabData = (force: any = false) => {
  const isForce = force === true
  if (isForce) {
    invalidateVisitCaches()
  }
  if (activeTab.value === 'table' || activeTab.value === 'map') {
    fetchVisits(isForce)
    fetchTransactions(isForce)
    fetchRekapVisits(isForce)
  } else if (activeTab.value === 'omzet') {
    fetchTransactions(isForce)
  } else if (activeTab.value === 'rekap') {
    fetchRekapData(isForce)
    fetchTransactions(isForce)
  } else if (activeTab.value === 'schedule') {
    fetchVisitPlans(isForce)
  } else if (activeTab.value === 'journey') {
    fetchJourneys()
  } else if (activeTab.value === 'merchandising') {
    fetchMerchandising()
  }
}

const resetAndFetch = (force: any = false) => {
  currentPage.value = 1
  omzetCurrentPage.value = 1
  fetchActiveTabData(force === true)
}

const onEmployeeFilterChange = () => {
  currentPage.value = 1
  omzetCurrentPage.value = 1

  if (activeTab.value === 'table' || activeTab.value === 'map') {
    fetchVisits(false)
    // Ensure summary & transaction data are loaded for current period (no-op if already cached)
    fetchTransactions(false)
    fetchRekapVisits(false)
  } else if (activeTab.value === 'omzet') {
    fetchTransactions(false)
  } else if (activeTab.value === 'rekap') {
    fetchRekapData(false)
    fetchTransactions(false)
  } else if (activeTab.value === 'schedule') {
    fetchVisitPlans(false)
  } else if (activeTab.value === 'journey') {
    fetchJourneys()
  }
}

const handleResetDateFilter = () => {
  const offset = (new Date()).getTimezoneOffset() * 60000
  selectedDate.value = (new Date(Date.now() - offset)).toISOString().split('T')[0]
  selectedMonth.value = ''
  filterMode.value = 'date'
  resetAndFetch(false)
}

const handleRefreshAll = () => {
  resetAndFetch(true)
}

const fetchVisits = async (force: any = false) => {
  const isForce = force === true
  let start: string | undefined
  let end: string | undefined
  if (filterMode.value === 'date' && selectedDate.value) {
    start = selectedDate.value
    end = selectedDate.value
  } else if (filterMode.value === 'month') {
    const monthVal = selectedMonth.value || (new Date().getFullYear() + '-' + String(new Date().getMonth() + 1).padStart(2, '0'))
    start = `${monthVal}-01`
    const [year, month] = monthVal.split('-')
    const lastDay = new Date(parseInt(year || '2026'), parseInt(month || '1'), 0).getDate()
    end = `${monthVal}-${lastDay}`
  }

  const isUnpaginated = activeTab.value === 'map' || itemsPerPage.value === -1
  // On Map view or 'Semua' itemsPerPage, fetch all visits for the period without employee_id
  // so switching salesman filters in 0ms client-side.
  const effectiveEmployeeId = isUnpaginated
    ? undefined
    : (selectedEmployee.value === 'ALL' ? undefined : selectedEmployee.value)

  const params: any = {
    start_date: start,
    end_date: end,
    employee_id: effectiveEmployeeId,
  }

  if (searchQuery.value.trim()) {
    params.search = searchQuery.value.trim()
  }

  if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
    params.branch_id = masterStore.selectedBranchId
  }

  if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
    params.sales_type = masterStore.selectedSalesType
  }

  if (isUnpaginated) {
    params.all = 'true'
  } else {
    params.limit = itemsPerPage.value
    params.offset = (currentPage.value - 1) * itemsPerPage.value
  }

  const baseAllCacheKey = `${getPeriodContextKey()}|search:${params.search || ''}|ALL`
  const requestCacheKey = isUnpaginated
    ? baseAllCacheKey
    : `${getPeriodContextKey()}|search:${params.search || ''}|emp:${effectiveEmployeeId || 'ALL'}|limit:${params.limit}|offset:${params.offset}`

  if (!isForce) {
    // 1. Exact cache hit
    const cached = visitsCache.get(requestCacheKey)
    if (cached) {
      visitStore.visits = cached.visits
      visitStore.total = cached.total
      return
    }

    // 2. If we already have the complete unpaginated dataset (or page 1 ALL where total <= limit) in cache,
    // we can derive any salesman's filtered page in 0ms without a network request!
    const allCached = visitsCache.get(baseAllCacheKey)
    if (allCached && effectiveEmployeeId) {
      const empVisits = allCached.visits.filter((v) => v.employee_id === effectiveEmployeeId)
      const total = empVisits.length
      const sliceStart = params.offset || 0
      const sliceEnd = params.limit ? sliceStart + params.limit : empVisits.length
      const pageVisits = isUnpaginated ? empVisits : empVisits.slice(sliceStart, sliceEnd)
      visitStore.visits = pageVisits
      visitStore.total = total
      visitsCache.set(requestCacheKey, { visits: pageVisits, total })
      return
    }
  }

  isLoading.value = true
  try {
    if (visitsAbortController) visitsAbortController.abort()
    visitsAbortController = new AbortController()
    const config = { signal: visitsAbortController.signal }

    await visitStore.fetchVisits(params, config)
    visitsCache.set(requestCacheKey, {
      visits: [...visitStore.visits],
      total: visitStore.total,
    })
    // If this was an 'ALL' query and all rows fit in this single response, also populate baseAllCacheKey
    if (!effectiveEmployeeId && (isUnpaginated || (params.offset === 0 && visitStore.total <= (params.limit || 50)))) {
      visitsCache.set(baseAllCacheKey, {
        visits: [...visitStore.visits],
        total: visitStore.total,
      })
    }
  } finally {
    isLoading.value = false
  }
}

// Transaction data for Omzet Tab and Summary Cards
const transactions = shallowRef<any[]>([])
const salesOrders = shallowRef<any[]>([])
const rekapVisits = shallowRef<any[]>([])
const isLoadingTransactions = ref(false)
const showOtherDetails = ref(false)

let rekapVisitsAbortController: AbortController | null = null
const fetchRekapVisits = async (force: any = false) => {
  const isForce = force === true
  const ctxKey = getPeriodContextKey()
  if (!isForce && lastRekapVisitsContextKey === ctxKey && rekapVisits.value.length > 0) {
    return
  }

  if (rekapVisitsAbortController) rekapVisitsAbortController.abort()
  rekapVisitsAbortController = new AbortController()
  const config = { signal: rekapVisitsAbortController.signal }

  try {
    let start: string | undefined
    let end: string | undefined
    if (filterMode.value === 'date' && selectedDate.value) {
      start = selectedDate.value
      end = selectedDate.value
    } else if (filterMode.value === 'month') {
      const monthVal = selectedMonth.value || (new Date().getFullYear() + '-' + String(new Date().getMonth() + 1).padStart(2, '0'))
      start = `${monthVal}-01`
      const [year, month] = monthVal.split('-')
      const lastDay = new Date(parseInt(year || '2026'), parseInt(month || '1'), 0).getDate()
      end = `${monthVal}-${String(lastDay).padStart(2, '0')}`
    } else {
      const today = new Date().toISOString().split('T')[0] || ''
      start = today
      end = today
    }

    const params: any = {
      all: 'true',
      summary: 'true',
      start_date: start,
      end_date: end,
    }

    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterStore.selectedBranchId
    }
    if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
      params.sales_type = masterStore.selectedSalesType
    }

    const res = await apiClient.get('/admin/sales/visits', { params, ...config })
    rekapVisits.value = res.data?.data || []
    lastRekapVisitsContextKey = ctxKey
  } catch (error: any) {
    if (error.name !== 'CanceledError') {
      console.error('Failed to fetch summary visits:', error)
    }
  }
}

const fetchTransactions = async (force: any = false) => {
  const isForce = force === true
  const ctxKey = getPeriodContextKey()
  if (!isForce && lastTransactionsContextKey === ctxKey) {
    return
  }

  if (transactionsAbortController) transactionsAbortController.abort()
  transactionsAbortController = new AbortController()
  const config = { signal: transactionsAbortController.signal }

  isLoadingTransactions.value = true
  try {
    const params = new URLSearchParams()
    params.append('limit', '-1')

    const soParams = new URLSearchParams()
    soParams.append('limit', '0') // Summary must include every order in the selected period.

    if (filterMode.value === 'date' && selectedDate.value) {
      params.append('date', selectedDate.value)
      soParams.append('date', selectedDate.value)
    } else if (filterMode.value === 'month') {
      const monthVal = selectedMonth.value || (() => {
        const today = new Date()
        return `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}`
      })()
      params.append('month', monthVal)
      soParams.append('month', monthVal)
    } else {
      const todayStr = new Date().toISOString().split('T')[0] || ''
      params.append('date', todayStr)
      soParams.append('date', todayStr)
    }

    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.append('branch_id', masterStore.selectedBranchId)
      soParams.append('branch_id', masterStore.selectedBranchId)
    }

    if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
      params.append('sales_type', masterStore.selectedSalesType)
      soParams.append('sales_type', masterStore.selectedSalesType)
    }

    const [trxRes, soRes] = await Promise.all([
      apiClient.get(`/admin/sales/transactions/all?${params.toString()}`, config),
      apiClient.get(`/admin/sales-orders?${soParams.toString()}`, config).catch((e: any) => {
        if (e.name !== 'CanceledError') {
          console.error('Failed to fetch sales orders:', e)
        }
        return { data: { data: [] } }
      }),
    ])

    transactions.value = trxRes.data?.data || []
    salesOrders.value = soRes.data?.data || []
    lastTransactionsContextKey = ctxKey
  } catch (error: any) {
    if (error.name !== 'CanceledError') {
      console.error('Failed to fetch transactions:', error)
    }
  } finally {
    isLoadingTransactions.value = false
  }
}

onMounted(async () => {
  if (props.isModernTrade || route.path.includes('modern-trade')) {
    masterStore.setSelectedSalesType('sales_mt')
  }
  await masterStore.fetchBranches()
  await masterStore.fetchEmployees()
  fetchActiveTabData()
})

watch(
  () => masterStore.selectedBranchId,
  () => {
    if (selectedEmployee.value !== 'ALL') {
      const exists = salesmen.value.some((s) => s.id === selectedEmployee.value)
      if (!exists) {
        selectedEmployee.value = 'ALL'
      }
    }
    resetAndFetch()
  },
)

watch(
  () => masterStore.selectedSalesType,
  () => {
    if (selectedEmployee.value !== 'ALL') {
      const exists = salesmen.value.some((s) => s.id === selectedEmployee.value)
      if (!exists) {
        selectedEmployee.value = 'ALL'
      }
    }
    resetAndFetch()
  },
)

let searchDebounceTimeout: any = null
watch(searchQuery, () => {
  if (searchDebounceTimeout) clearTimeout(searchDebounceTimeout)
  searchDebounceTimeout = setTimeout(() => {
    currentPage.value = 1
    omzetCurrentPage.value = 1
    if (activeTab.value === 'table' || activeTab.value === 'map') {
      fetchVisits()
    }
  }, 350)
})

const filteredVisits = computed(() => {
  const list = visitStore.visits || []
  if (selectedEmployee.value !== 'ALL') {
    return list.filter((v) => v.employee_id === selectedEmployee.value)
  }
  return list
})

const getSalesCategory = (employeeId?: string, employeeObj?: any): 'TASK_ORDER' | 'MOTORIS' | 'BOOSTER' | 'OTHER' => {
  let title = (employeeObj?.job_position?.title || employeeObj?.job_title || '').toLowerCase()
  if (!title && employeeId) {
    const found = masterStore.employees.find((e: any) => e.id === employeeId)
    if (found) {
      title = (found.job_position?.title || found.job_title || '').toLowerCase()
    }
  }
  if (title.includes('booster')) return 'BOOSTER'
  if (title.includes('motoris')) return 'MOTORIS'
  if (toRolePattern.test(title)) return 'TASK_ORDER'
  return 'OTHER'
}

// Keep the TO card in sync with the sales_to filter in sales_order_repository.go.
const isTaskOrderSalesOrder = (so: any): boolean => {
  if (so.is_mt) return false
  const title = String(so.employee?.job_position?.title || '').toLowerCase()
  return toRolePattern.test(title)
}

const visitStats = computed(() => {
  let totalKunjungan = 0
  let tokoBaru = 0
  let tokoLama = 0
  let totalTransaksi = 0
  let totalNominal = 0

  let nominalTO = 0
  let transaksiTO = 0
  let nominalMotoris = 0
  let transaksiMotoris = 0
  let nominalBooster = 0
  let transaksiBooster = 0
  let nominalOther = 0
  let transaksiOther = 0
  const otherDetails: Array<{ id: string; number: string; source: string; employee: string; store: string; amount: number }> = []

  const addOther = (record: any, amount: number, source: string, number: string) => {
    nominalOther += amount
    transaksiOther++
    otherDetails.push({
      id: String(record.id || number),
      number,
      source,
      employee: [record.employee?.first_name, record.employee?.last_name].filter(Boolean).join(' ') || '-',
      store: record.store?.name || '-',
      amount,
    })
  }

  rekapVisits.value.forEach(visit => {
    // Only count if it matches the selected employee
    if (selectedEmployee.value !== 'ALL' && visit.employee_id !== selectedEmployee.value) return;
    if (visit.type === 'SKIP' || visit.status === 'SKIPPED') return;

    totalKunjungan++
    if (visit.store?.category === 'TOKO_BARU') {
      tokoBaru++
    } else {
      tokoLama++
    }
  })

  // Set of SO and Invoice IDs to avoid double-counting
  const countedOrderIds = new Set<string>()

  // 1. Process Sales Orders (Task Order / Pre-Order)
  salesOrders.value.forEach(so => {
    if (selectedEmployee.value !== 'ALL' && so.employee_id !== selectedEmployee.value) return;
    if (so.status === 'CANCELLED' || so.status === 'REJECTED' || so.status === 'DELETED') return;

    // Apply searchQuery filter if user is searching
    const q = searchQuery.value.trim().toLowerCase()
    if (q) {
      const matchStore = so.store?.name?.toLowerCase().includes(q)
      const matchSO = so.so_number?.toLowerCase().includes(q)
      const matchSales = `${so.employee?.first_name || ''} ${so.employee?.last_name || ''}`.toLowerCase().includes(q)
      if (!matchStore && !matchSO && !matchSales) return;
    }

    const soId = String(so.id || '')
    const soNum = String(so.so_number || '')
    if (soId) countedOrderIds.add(soId)
    if (soNum) countedOrderIds.add(soNum)
    if (so.invoice_id) countedOrderIds.add(String(so.invoice_id))
    if (so.invoice?.id) countedOrderIds.add(String(so.invoice.id))
    if (so.invoice?.receipt_no) countedOrderIds.add(String(so.invoice.receipt_no))

    const amount = Number(so.total_amount) || 0
    totalTransaksi++
    totalNominal += amount

    const cat = getSalesCategory(so.employee_id, so.employee)
    if (isTaskOrderSalesOrder(so)) {
      nominalTO += amount
      transaksiTO++
    } else if (cat === 'BOOSTER' && !so.is_mt) {
      nominalBooster += amount
      transaksiBooster++
    } else if (cat === 'MOTORIS' && !so.is_mt) {
      nominalMotoris += amount
      transaksiMotoris++
    } else {
      addOther(so, amount, so.is_mt ? 'SO Modern Trade' : 'SO jabatan lain', so.so_number || '-')
    }
  })

  // 2. Process Direct Sales Transactions (Motoris / Canvas / Booster Direct Sale)
  // Exclude transactions that originated from Sales Orders or non-salesman channels
  filteredTransactions.value.forEach(t => {
    if (t.status === 'CANCELLED' || t.status === 'REJECTED' || t.status === 'DELETED') return;
    if (t.origin_sales_order_id) return;
    if (t.payment_method === 'SALES_ORDER') return;

    const tId = String(t.id || '')
    const tReceipt = String(t.receipt_no || '')
    if ((tId && countedOrderIds.has(tId)) || (tReceipt && countedOrderIds.has(tReceipt))) {
      return
    }

    // Exclude non-salesman channels
    const channel = String(t.sales_channel || '').toUpperCase()
    if (channel === 'DIRECT_BRANCH' || channel === 'DIRECT_FACTORY' || channel === 'EMPLOYEE_SALE') {
      return
    }

    if (tId) countedOrderIds.add(tId)
    if (tReceipt) countedOrderIds.add(tReceipt)

    const amount = Number(t.total_amount) || 0
    totalTransaksi++
    totalNominal += amount

    const cat = getSalesCategory(t.employee_id, t.employee)
    if (cat === 'BOOSTER') {
      nominalBooster += amount
      transaksiBooster++
    } else if (cat === 'MOTORIS' || channel === 'MOTORIS') {
      nominalMotoris += amount
      transaksiMotoris++
    } else {
      addOther(t, amount, 'Transaksi langsung lainnya', t.receipt_no || '-')
    }
  })


  return {
    totalKunjungan,
    tokoBaru,
    tokoLama,
    totalTransaksi,
    totalNominal,
    nominalTO,
    transaksiTO,
    nominalMotoris,
    transaksiMotoris,
    nominalBooster,
    transaksiBooster,
    nominalOther,
    transaksiOther,
    otherDetails,
  }
})

const filteredTransactions = computed(() => {
  let filtered = transactions.value

  // Exclude delivery invoices originating from SOs and SO placeholders
  filtered = filtered.filter(t => !t.origin_sales_order_id && t.payment_method !== 'SALES_ORDER')

  // Exclude non-salesman channels
  filtered = filtered.filter(t => {
    const channel = String(t.sales_channel || '').toUpperCase()
    return channel !== 'DIRECT_BRANCH' && channel !== 'DIRECT_FACTORY' && channel !== 'EMPLOYEE_SALE'
  })

  if (selectedEmployee.value !== 'ALL') {
    filtered = filtered.filter((t) => t.employee_id === selectedEmployee.value)
  }

  if (filterMode.value === 'date' && selectedDate.value) {
    filtered = filtered.filter((t) => {
      if (!t.transaction_date) return true
      return t.transaction_date.startsWith(selectedDate.value)
    })
  } else if (filterMode.value === 'month' && selectedMonth.value) {
    filtered = filtered.filter((t) => {
      if (!t.transaction_date) return true
      return t.transaction_date.startsWith(selectedMonth.value)
    })
  }

  const q = searchQuery.value.toLowerCase()
  if (q) {
    filtered = filtered.filter(
      (t) => t.receipt_no?.toLowerCase().includes(q) || t.store?.name?.toLowerCase().includes(q),
    )
  }

  return filtered
})

const omzetCurrentPage = ref(1)
const omzetItemsPerPage = ref(50)

const omzetTotalPages = computed(() => {
  return Math.ceil(filteredTransactions.value.length / omzetItemsPerPage.value) || 1
})

const paginatedTransactions = computed(() => {
  const start = (omzetCurrentPage.value - 1) * omzetItemsPerPage.value
  const end = start + omzetItemsPerPage.value
  return filteredTransactions.value.slice(start, end)
})

const pendingTransactions = computed(() =>
  filteredTransactions.value.filter((t) => t.status === 'PENDING' && t.payment_status !== 'PAID'),
)
const verifiedTransactions = computed(() =>
  filteredTransactions.value.filter((t) => t.status === 'VERIFIED' || t.payment_status === 'PAID'),
)

const totalPendingAmount = computed(() =>
  pendingTransactions.value.reduce((sum, t) => sum + (t.total_amount || 0), 0),
)
const totalVerifiedAmount = computed(() =>
  verifiedTransactions.value.reduce((sum, t) => sum + (t.total_amount || 0), 0),
)

// ─── Journey (KM & BBM) State ───
const journeys = shallowRef<any[]>([])
const isLoadingJourneys = ref(false)
let journeysAbortController: AbortController | null = null

const fetchJourneys = async () => {
  if (journeysAbortController) journeysAbortController.abort()
  journeysAbortController = new AbortController()
  const config = { signal: journeysAbortController.signal }

  isLoadingJourneys.value = true
  try {
    const params = new URLSearchParams()
    if (filterMode.value === 'date' && selectedDate.value) {
      params.append('start_date', selectedDate.value)
      params.append('end_date', selectedDate.value)
    } else if (filterMode.value === 'month') {
      const monthVal = selectedMonth.value || (new Date().getFullYear() + '-' + String(new Date().getMonth() + 1).padStart(2, '0'))
      const [year, month] = monthVal.split('-')
      const lastDay = new Date(parseInt(year || '2026'), parseInt(month || '1'), 0).getDate()
      params.append('start_date', `${monthVal}-01`)
      params.append('end_date', `${monthVal}-${lastDay}`)
    }

    if (selectedEmployee.value !== 'ALL') {
      params.append('employee_id', selectedEmployee.value)
    }

    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.append('branch_id', masterStore.selectedBranchId)
    }

    const res = await apiClient.get('/admin/sales/journeys?' + params.toString(), config)
    journeys.value = res.data?.data || []
  } catch (e: any) {
    if (e.name !== 'CanceledError' && e.name !== 'AbortError') {
      console.error('Failed to fetch sales journeys:', e)
      toast.error('Gagal mengambil data perjalanan sales')
    }
  } finally {
    isLoadingJourneys.value = false
  }
}

// Image modal for speedometer & fuel receipts
const showImageModal = ref(false)
const modalImageUrl = ref('')
const modalImageTitle = ref('')
const modalImageSub = ref('')
const openImageModal = (url: string, title: string, subtitle = '') => {
  modalImageUrl.value = url
  modalImageTitle.value = title
  modalImageSub.value = subtitle
  showImageModal.value = true
}

// ─── Admin Journey CRUD & BBM Management ───
const showJourneyModal = ref(false)
const isEditingJourney = ref(false)
const isSavingJourney = ref(false)
const journeyForm = ref({
  id: '',
  employee_id: '',
  date: todayLocal,
  start_km: 0,
  start_time: '08:00',
  start_location: '',
  start_photo_url: '',
  status: 'STARTED',
  end_km: null as number | null,
  end_time: '' as string,
  end_location: '',
  end_photo_url: '',
  bbm_amount: null as number | null,
  bbm_proof_url: '',
})

const openCreateJourneyModal = () => {
  isEditingJourney.value = false
  journeyForm.value = {
    id: '',
    employee_id: selectedEmployee.value !== 'ALL' ? selectedEmployee.value : (salesmen.value[0]?.id || ''),
    date: filterMode.value === 'date' && selectedDate.value ? selectedDate.value : todayLocal,
    start_km: 0,
    start_time: '08:00',
    start_location: '',
    start_photo_url: '',
    status: 'STARTED',
    end_km: null,
    end_time: '',
    end_location: '',
    end_photo_url: '',
    bbm_amount: null,
    bbm_proof_url: '',
  }
  showJourneyModal.value = true
}

const openEditJourneyModal = (j: any) => {
  isEditingJourney.value = true
  const dateStr = typeof j.date === 'string' ? j.date.split('T')[0] : formatDateISO(new Date(j.date))
  const formatTimeHM = (t: any) => {
    if (!t) return ''
    const d = new Date(t)
    return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }).replace('.', ':')
  }

  journeyForm.value = {
    id: j.id,
    employee_id: j.employee?.id || '',
    date: dateStr,
    start_km: j.start_km || 0,
    start_time: formatTimeHM(j.start_time) || '08:00',
    start_location: j.start_location || '',
    start_photo_url: j.start_photo_url || '',
    status: j.status || 'STARTED',
    end_km: j.end_km !== undefined && j.end_km !== null ? j.end_km : null,
    end_time: j.end_time ? formatTimeHM(j.end_time) : '',
    end_location: j.end_location || '',
    end_photo_url: j.end_photo_url || '',
    bbm_amount: null,
    bbm_proof_url: '',
  }
  showJourneyModal.value = true
}

const submitJourneyForm = async () => {
  if (!isEditingJourney.value && !journeyForm.value.employee_id) {
    toast.error('Pilih Salesman terlebih dahulu')
    return
  }
  if (!journeyForm.value.date) {
    toast.error('Pilih Tanggal perjalanan')
    return
  }
  if (journeyForm.value.start_km < 0) {
    toast.error('KM Berangkat tidak boleh negatif')
    return
  }
  if (journeyForm.value.status === 'ENDED' && (journeyForm.value.end_km === null || journeyForm.value.end_km === undefined)) {
    toast.error('Jika status Selesai, KM Pulang wajib diisi')
    return
  }
  if (journeyForm.value.status === 'ENDED' && journeyForm.value.end_km !== null && journeyForm.value.end_km < journeyForm.value.start_km) {
    toast.error('KM Pulang tidak boleh lebih kecil dari KM Berangkat')
    return
  }

  isSavingJourney.value = true
  try {
    if (isEditingJourney.value) {
      const payload: any = {
        date: journeyForm.value.date,
        start_km: Number(journeyForm.value.start_km),
        start_time: journeyForm.value.start_time,
        start_location: journeyForm.value.start_location,
        start_photo_url: journeyForm.value.start_photo_url,
        status: journeyForm.value.status,
      }
      if (journeyForm.value.end_km !== null) {
        payload.end_km = Number(journeyForm.value.end_km)
      }
      if (journeyForm.value.end_time) {
        payload.end_time = journeyForm.value.end_time
      }
      if (journeyForm.value.end_location) {
        payload.end_location = journeyForm.value.end_location
      }
      if (journeyForm.value.end_photo_url) {
        payload.end_photo_url = journeyForm.value.end_photo_url
      }

      await apiClient.put(`/admin/sales/journeys/${journeyForm.value.id}`, payload)
      toast.success('Perjalanan berhasil diperbarui')
    } else {
      const payload: any = {
        employee_id: journeyForm.value.employee_id,
        date: journeyForm.value.date,
        start_km: Number(journeyForm.value.start_km),
        start_time: journeyForm.value.start_time,
        start_location: journeyForm.value.start_location,
        start_photo_url: journeyForm.value.start_photo_url,
        status: journeyForm.value.status,
      }
      if (journeyForm.value.status === 'ENDED' && journeyForm.value.end_km !== null) {
        payload.end_km = Number(journeyForm.value.end_km)
        payload.end_time = journeyForm.value.end_time || '17:00'
        payload.end_location = journeyForm.value.end_location
        payload.end_photo_url = journeyForm.value.end_photo_url
      }
      if (journeyForm.value.bbm_amount && journeyForm.value.bbm_amount > 0) {
        payload.bbm_amount = Number(journeyForm.value.bbm_amount)
        payload.bbm_proof_url = journeyForm.value.bbm_proof_url
      }

      await apiClient.post('/admin/sales/journeys', payload)
      toast.success('Perjalanan sales berhasil dicatat (Bypass / Manual)')
    }
    showJourneyModal.value = false
    fetchJourneys()
  } catch (err: any) {
    console.error('Failed to save journey:', err)
    toast.error(err.response?.data?.error || 'Gagal menyimpan data perjalanan')
  } finally {
    isSavingJourney.value = false
  }
}

const deleteJourney = async (j: any) => {
  const salesName = j.employee?.first_name ? `${j.employee.first_name} ${j.employee.last_name || ''}` : 'Sales'
  if (!confirm(`Hapus catatan perjalanan ${salesName} pada tanggal ${formatDate(j.date)}?`)) return

  try {
    await apiClient.delete(`/admin/sales/journeys/${j.id}`)
    toast.success('Catatan perjalanan berhasil dihapus')
    fetchJourneys()
  } catch (err: any) {
    console.error('Failed to delete journey:', err)
    toast.error(err.response?.data?.error || 'Gagal menghapus perjalanan')
  }
}

// ─── BBM Modal State ───
const showBBMModal = ref(false)
const isSavingBBM = ref(false)
const selectedJourneyForBBM = ref<any>(null)
const bbmForm = ref({
  amount: 0,
  attachment_url: '',
  description: '',
})

const openAddBBMModal = (j: any) => {
  selectedJourneyForBBM.value = j
  bbmForm.value = {
    amount: 0,
    attachment_url: '',
    description: `BBM Sales - ${j.employee?.first_name || ''}`,
  }
  showBBMModal.value = true
}

const submitBBMForm = async () => {
  if (!bbmForm.value.amount || bbmForm.value.amount <= 0) {
    toast.error('Masukkan nominal klaim BBM yang valid')
    return
  }

  isSavingBBM.value = true
  try {
    await apiClient.post(`/admin/sales/journeys/${selectedJourneyForBBM.value.id}/bbm`, {
      amount: Number(bbmForm.value.amount),
      attachment_url: bbmForm.value.attachment_url,
      description: bbmForm.value.description,
    })
    toast.success('Klaim BBM berhasil ditambahkan')
    showBBMModal.value = false
    fetchJourneys()
  } catch (err: any) {
    console.error('Failed to add BBM:', err)
    toast.error(err.response?.data?.error || 'Gagal menambahkan klaim BBM')
  } finally {
    isSavingBBM.value = false
  }
}

const deleteBBM = async (bbm: any, j: any) => {
  if (!confirm(`Hapus klaim BBM senilai Rp ${new Intl.NumberFormat('id-ID').format(bbm.amount)}?`)) return

  try {
    await apiClient.delete(`/admin/sales/journeys/bbm/${bbm.id}`)
    toast.success('Klaim BBM berhasil dihapus')
    fetchJourneys()
  } catch (err: any) {
    console.error('Failed to delete BBM:', err)
    toast.error(err.response?.data?.error || 'Gagal menghapus klaim BBM')
  }
}

const journeyStats = computed(() => {
  const totalJourneys = journeys.value.length
  const totalKm = journeys.value.reduce((s, j) => s + (j.total_km || 0), 0)
  const totalBbm = journeys.value.reduce((s, j) => s + (j.total_bbm_amount || 0), 0)
  const avgKm = totalJourneys > 0 ? (totalKm / totalJourneys) : 0
  const costPerKm = totalKm > 0 ? (totalBbm / totalKm) : 0
  return {
    totalJourneys,
    totalKm,
    totalBbm,
    avgKm,
    costPerKm,
  }
})


const mapContainer = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let markerGroup: L.FeatureGroup | null = null
let routeGroup: L.FeatureGroup | null = null

const visitPlans = ref<any[]>([])
const isLoadingPlans = ref(false)

const rekapPerformances = shallowRef<any[]>([])
const prevRekapPerformances = shallowRef<any[]>([])
const rekapBanners = shallowRef<any[]>([])
const isLoadingRekap = ref(false)

const fetchRekapData = async (force: any = false) => {
  const isForce = force === true
  const ctxKey = getPeriodContextKey()
  if (!isForce && lastRekapDataContextKey === ctxKey && rekapPerformances.value.length > 0) {
    return
  }

  isLoadingRekap.value = true
  try {
    const perfParams: any = {}
    let currentMonth = new Date().getMonth() + 1
    let currentYear = new Date().getFullYear()

    if (filterMode.value === 'date' && selectedDate.value) {
      perfParams.date = selectedDate.value
      const d = new Date(selectedDate.value)
      currentMonth = d.getMonth() + 1
      currentYear = d.getFullYear()
    } else if (filterMode.value === 'month' && selectedMonth.value) {
      const parts = selectedMonth.value.split('-')
      perfParams.year = Number(parts[0] || '')
      perfParams.month = Number(parts[1] || '')
      currentYear = perfParams.year || currentYear
      currentMonth = perfParams.month || currentMonth
    } else {
      const now = new Date()
      perfParams.month = now.getMonth() + 1
      perfParams.year = now.getFullYear()
      currentMonth = perfParams.month
      currentYear = perfParams.year
    }

    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      perfParams.branch_id = masterStore.selectedBranchId
    }
    if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
      perfParams.sales_type = masterStore.selectedSalesType
    }

    const prevMonth = currentMonth === 1 ? 12 : currentMonth - 1
    const prevYear = currentMonth === 1 ? currentYear - 1 : currentYear
    
    const prevPerfParams: any = {
      month: prevMonth,
      year: prevYear,
      branch_id: perfParams.branch_id,
      sales_type: perfParams.sales_type
    }

    if (rekapAbortController) rekapAbortController.abort()
    rekapAbortController = new AbortController()
    const config = { signal: rekapAbortController.signal }

    const needsRekapVisits = isForce || lastRekapVisitsContextKey !== ctxKey || rekapVisits.value.length === 0

    const [perfRes, prevPerfRes, bannerRes, visitsRes] = await Promise.all([
      apiClient.get('/admin/sales/reports/performance', { params: perfParams, ...config }).catch((e) => {
        if (e.name !== 'CanceledError') console.error(e);
        return { data: { data: [] } }
      }),
      apiClient.get('/admin/sales/reports/performance', { params: prevPerfParams, ...config }).catch((e) => {
        if (e.name !== 'CanceledError') console.error(e);
        return { data: { data: [] } }
      }),
      apiClient.get('/banner-orders', config).catch((e) => {
        if (e.name !== 'CanceledError') console.error(e);
        return { data: { data: [] } }
      }),
      needsRekapVisits
        ? apiClient.get('/admin/sales/visits', { 
            params: { 
              all: 'true',
              limit: 10000,
              summary: 'true',
              start_date: perfParams.date || (perfParams.year ? `${perfParams.year}-${String(perfParams.month).padStart(2, '0')}-01` : undefined),
              end_date: perfParams.date || (perfParams.year ? `${perfParams.year}-${String(perfParams.month).padStart(2, '0')}-${new Date(perfParams.year, perfParams.month, 0).getDate()}` : undefined),
              branch_id: perfParams.branch_id,
              sales_type: perfParams.sales_type
            },
            ...config
          }).catch((e) => {
            if (e.name !== 'CanceledError') console.error(e);
            return { data: { data: [] } }
          })
        : Promise.resolve(null)
    ])
    
    rekapPerformances.value = perfRes.data?.data || []
    prevRekapPerformances.value = prevPerfRes.data?.data || []
    rekapBanners.value = bannerRes.data?.data || []
    if (visitsRes) {
      rekapVisits.value = visitsRes.data?.data || []
      lastRekapVisitsContextKey = ctxKey
    }
    lastRekapDataContextKey = ctxKey
  } catch (error: any) {
    if (error.name !== 'CanceledError') {
      console.error('Failed to fetch rekap data:', error)
    }
  } finally {
    isLoadingRekap.value = false
  }
}

const computedRekap = computed(() => {
  const result: any[] = []

  let filterDateStr: string = ''
  if (filterMode.value === 'date') {
    // If no date selected, default to today in local time (YYYY-MM-DD)
    if (selectedDate.value) {
      filterDateStr = selectedDate.value
    } else {
      const tzOffset = (new Date()).getTimezoneOffset() * 60000
      filterDateStr = (new Date(Date.now() - tzOffset)).toISOString().split('T')[0] || ''
    }
  } else {
    filterDateStr = selectedMonth.value || (new Date().getFullYear() + '-' + String(new Date().getMonth() + 1).padStart(2, '0'))
  }

  salesmen.value.forEach((emp: any) => {
    if (selectedEmployee.value !== 'ALL' && emp.id !== selectedEmployee.value) return

    let actualTotal = 0
    let actualBaru = 0

    rekapVisits.value.forEach(visit => {
      if (visit.employee_id === emp.id) {
        const visitDate = visit.check_in_time ? formatDateISO(new Date(visit.check_in_time)) : ''
        if (filterMode.value === 'date') {
          if (visitDate === filterDateStr && visit.type !== 'SKIP') {
            actualTotal++
            if (visit.store?.category === 'TOKO_BARU') actualBaru++
          }
        } else {
          if (visitDate?.startsWith(filterDateStr) && visit.type !== 'SKIP') {
            actualTotal++
            if (visit.store?.category === 'TOKO_BARU') actualBaru++
          }
        }
      }
    })

    const perf = rekapPerformances.value.find((p: any) => p.EmployeeID === emp.id || p.employee_id === emp.id) || {}
    const prevPerf = prevRekapPerformances.value.find((p: any) => p.EmployeeID === emp.id || p.employee_id === emp.id) || {}
    
    const targetTotal = perf.target_daily_visits || perf.TargetDailyVisits || prevPerf.target_daily_visits || prevPerf.TargetDailyVisits || 0
    const targetBaru = perf.target_daily_new_stores || perf.TargetDailyNewStores || prevPerf.target_daily_new_stores || prevPerf.TargetDailyNewStores || 0

    let bannersCreated = 0
    rekapBanners.value.forEach(b => {
      if (b.employee_id === emp.id) {
        const createdAt = b.created_at ? formatDateISO(new Date(b.created_at)) : ''
        if (filterMode.value === 'date') {
          if (createdAt === filterDateStr) bannersCreated++
        } else {
          if (createdAt?.startsWith(filterDateStr)) bannersCreated++
        }
      }
    })

    result.push({
      employee: emp,
      actualTotal,
      targetTotal,
      actualBaru,
      targetBaru,
      bannersCreated
    })
  })

  return result
})

// Adapt computedRekap to the shape expected by the template
const rekapSummary = computed(() => {
  return computedRekap.value.map(r => ({
    employeeId: r.employee?.id || '',
    name: `${r.employee?.first_name || ''} ${r.employee?.last_name || ''}`.trim(),
    branch: r.employee?.branch?.name || '-',
    totalKunjungan: r.actualTotal,
    tokoBaru: r.actualBaru,
    tokoLama: r.actualTotal - r.actualBaru,
    targetDailyVisits: r.targetTotal,
    targetDailyNewStores: r.targetBaru,
    spandukDibuat: r.bannersCreated,
  })).sort((a, b) => b.totalKunjungan - a.totalKunjungan)
})

const fetchVisitPlans = async (force: any = false) => {
  const isForce = force === true
  if (selectedEmployee.value === 'ALL') {
    visitPlans.value = []
    return
  }

  const dateParam = selectedDate.value || new Date().toISOString().split('T')[0] || ''
  const planCacheKey = `${dateParam}|${selectedEmployee.value}|${masterStore.selectedBranchId || 'ALL'}|${masterStore.selectedSalesType || 'ALL'}`
  if (!isForce && visitPlansCache.has(planCacheKey)) {
    visitPlans.value = visitPlansCache.get(planCacheKey) || []
    return
  }

  if (plansAbortController) plansAbortController.abort()
  plansAbortController = new AbortController()
  const config = { signal: plansAbortController.signal }

  isLoadingPlans.value = true
  try {
    const params: any = {
      employee_id: selectedEmployee.value,
      date: dateParam,
    }

    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterStore.selectedBranchId
    }

    if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
      params.sales_type = masterStore.selectedSalesType
    }

    const res = await apiClient.get('/admin/sales/visit-plans', {
      params,
      ...config
    })
    const data = res.data.data || []
    visitPlans.value = data
    visitPlansCache.set(planCacheKey, data)
  } catch (error: any) {
    if (error.name !== 'CanceledError') {
      console.error('Failed to fetch visit plans:', error)
    }
  } finally {
    isLoadingPlans.value = false
  }
}

// ─── Map Dynamic Scaling, Route & Numbering State ───
const SALESMAN_COLORS = [
  '#2563eb', // Blue
  '#059669', // Emerald
  '#d97706', // Amber
  '#dc2626', // Red
  '#7c3aed', // Violet
  '#0891b2', // Cyan
  '#db2777', // Pink
  '#ea580c', // Orange
  '#4f46e5', // Indigo
  '#0d9488', // Teal
]

const getSalesmanColor = (id?: string | null) => {
  if (!id) return '#2563eb'
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash)
  }
  const index = Math.abs(hash) % SALESMAN_COLORS.length
  return SALESMAN_COLORS[index]
}

const isMobileScreen = () => {
  return typeof window !== 'undefined' && window.innerWidth < 640
}

const getMarkerConfig = (zoom: number, isMobile: boolean) => {
  if (zoom <= 9) {
    const size = isMobile ? 14 : 16
    return { size, fontSize: isMobile ? 8 : 9, borderWidth: 1.5 }
  }
  if (zoom <= 12) {
    const size = isMobile ? 20 : 24
    return { size, fontSize: isMobile ? 9 : 11, borderWidth: 2 }
  }
  if (zoom <= 15) {
    const size = isMobile ? 26 : 30
    return { size, fontSize: isMobile ? 11 : 12, borderWidth: 2 }
  }
  const size = isMobile ? 32 : 36
  return { size, fontSize: isMobile ? 12 : 13, borderWidth: 2.5 }
}

const fitMapToBounds = () => {
  if (!map) return
  const bounds = L.latLngBounds([])
  filteredVisits.value.forEach((visit) => {
    if (visit.latitude && visit.longitude && visit.latitude !== 0 && visit.longitude !== 0) {
      bounds.extend(L.latLng(visit.latitude, visit.longitude))
    }
  })
  if (bounds.isValid()) {
    map.fitBounds(bounds, { padding: isMobileScreen() ? [30, 30] : [50, 50], maxZoom: 16 })
  }
}

const renderMapElements = (shouldFitBounds = false) => {
  if (!map || !markerGroup || !routeGroup) return

  markerGroup.clearLayers()
  routeGroup.clearLayers()

  const currentZoom = map.getZoom()
  const isMobile = isMobileScreen()
  const { size, fontSize, borderWidth } = getMarkerConfig(currentZoom, isMobile)
  const bounds = L.latLngBounds([])

  // Filter visits with valid coordinates
  const validVisits = filteredVisits.value.filter(
    (v) => v.latitude && v.longitude && v.latitude !== 0 && v.longitude !== 0
  )

  // Group visits by salesman
  const visitsBySalesman: Record<string, any[]> = {}
  validVisits.forEach((v) => {
    const empId = v.employee_id || 'UNKNOWN'
    if (!visitsBySalesman[empId]) {
      visitsBySalesman[empId] = []
    }
    visitsBySalesman[empId].push(v)
  })

  // Sort each salesman's visits chronologically & render markers + polyline
  Object.keys(visitsBySalesman).forEach((empId) => {
    const group = visitsBySalesman[empId]
    if (!group || group.length === 0) return
    group.sort((a, b) => {
      const timeA = new Date(a.check_in_time || a.created_at || 0).getTime()
      const timeB = new Date(b.check_in_time || b.created_at || 0).getTime()
      return timeA - timeB
    })

    const salesmanColor = getSalesmanColor(empId)
    const routeCoords: L.LatLngExpression[] = []
    const firstEmp = group[0]?.employee
    const salesmanName = firstEmp
      ? `${firstEmp.first_name || ''} ${firstEmp.last_name || ''}`.trim()
      : (group[0]?.employee_name || 'Sales')

    group.forEach((visit, index) => {
      const seqNumber = index + 1
      const latlng = L.latLng(visit.latitude, visit.longitude)
      routeCoords.push(latlng)
      bounds.extend(latlng)

      const isOngoing = !visit.check_out_time

      const iconHtml = `
        <div style="position: relative; width: ${size}px; height: ${size}px; display: flex; align-items: center; justify-content: center; cursor: pointer;">
          ${isOngoing ? `<div style="position: absolute; inset: -4px; border-radius: 9999px; background-color: ${salesmanColor}; opacity: 0.35; animation: ping 1.5s cubic-bezier(0, 0, 0.2, 1) infinite;"></div>` : ''}
          <div style="
            width: ${size}px;
            height: ${size}px;
            border-radius: 9999px;
            background-color: ${salesmanColor};
            border: ${borderWidth}px solid #ffffff;
            box-shadow: 0 2px 6px rgba(0,0,0,0.35);
            display: flex;
            align-items: center;
            justify-content: center;
            color: #ffffff;
            font-weight: 900;
            font-size: ${fontSize}px;
            line-height: 1;
            user-select: none;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
          ">
            ${seqNumber}
          </div>
        </div>
      `

      const icon = L.divIcon({
        className: 'sales-seq-marker',
        html: iconHtml,
        iconSize: [size, size],
        iconAnchor: [size / 2, size / 2],
        popupAnchor: [0, -size / 2],
      })

      const popupContent = `
        <div style="min-width: 220px; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;">
          <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 6px;">
            <span style="background: ${salesmanColor}; color: #ffffff; font-size: 10px; font-weight: 800; padding: 2px 8px; border-radius: 999px;">
              Kunjungan #${seqNumber}
            </span>
            <span style="font-size: 10px; font-weight: 700; color: #64748b;">
              ${visit.store?.category === 'TOKO_BARU' || visit.is_noo ? '⭐ Toko Baru' : 'Toko Lama'}
            </span>
          </div>
          <h3 style="font-weight: 900; margin: 0 0 2px 0; font-size: 14px; color: #0f172a; line-height: 1.2;">
            ${visit.store?.name || 'Toko'}
          </h3>
          <p style="margin: 0 0 8px 0; font-size: 10px; font-weight: 500; color: #94a3b8; line-height: 1.3;">
            ${visit.store?.address || 'Alamat tidak tersedia'}
          </p>
          
          <div style="background: #f8fafc; padding: 6px 8px; border-radius: 6px; margin-bottom: 8px; border-left: 3px solid ${salesmanColor};">
            <p style="margin: 0; font-size: 11px; font-weight: 700; color: #334155;">
              Sales: ${salesmanName}
            </p>
          </div>

          <div style="margin-top: 8px; border-top: 1px solid #f1f5f9; padding-top: 8px; font-size: 10px; font-weight: 800; color: #475569;">
            <p style="margin: 0 0 2px 0;">🗓️ ${formatDate(visit.check_in_time)}</p>
            <p style="margin: 0 0 2px 0;">🟢 Masuk: ${formatTime(visit.check_in_time)}</p>
            ${visit.check_out_time ? `<p style="margin: 0;">🔴 Keluar: ${formatTime(visit.check_out_time)}</p>` : `<p style="margin: 0; color: #f59e0b;">⚡ Sedang Berkunjung</p>`}
          </div>

          <div style="margin-top: 8px; display: flex; gap: 4px;">
            <a href="https://www.google.com/maps/@?api=1&map_action=pano&viewpoint=${visit.latitude},${visit.longitude}" target="_blank" style="flex: 1; text-align: center; background: #e0e7ff; color: #4338ca; font-size: 9px; font-weight: 800; padding: 4px 6px; border-radius: 6px; text-decoration: none; border: 1px solid #c7d2fe;">Street View 360°</a>
            <a href="https://www.google.com/maps/search/?api=1&query=${visit.latitude},${visit.longitude}" target="_blank" style="flex: 1; text-align: center; background: #f1f5f9; color: #334155; font-size: 9px; font-weight: 800; padding: 4px 6px; border-radius: 6px; text-decoration: none; border: 1px solid #e2e8f0;">Google Maps</a>
          </div>
        </div>
      `

      const marker = L.marker(latlng, { icon }).bindPopup(popupContent)
      marker.addTo(markerGroup!)
    })

    // Draw route line if this salesman has at least 2 visit stops
    if (routeCoords.length >= 2) {
      const polylineWeight = isMobile ? (currentZoom <= 12 ? 2.5 : 3.5) : (currentZoom <= 12 ? 3 : 4)
      const polyline = L.polyline(routeCoords, {
        color: salesmanColor,
        weight: polylineWeight,
        opacity: 0.8,
        dashArray: '6, 6',
        lineJoin: 'round',
      })
      polyline.bindTooltip(`Rute ${salesmanName}: ${routeCoords.length} Kunjungan`, {
        sticky: true,
      })
      polyline.addTo(routeGroup!)
    }
  })

  if (shouldFitBounds && bounds.isValid() && map) {
    map.fitBounds(bounds, { padding: isMobile ? [30, 30] : [50, 50], maxZoom: 16 })
  }
}

const handleMapResize = () => {
  if (activeTab.value === 'map' && map) {
    map.invalidateSize()
    renderMapElements(false)
  }
}

const initMap = () => {
  if (!mapContainer.value) return

  if (!map) {
    map = L.map(mapContainer.value, {
      zoomControl: true,
    }).setView([-6.2, 106.816666], 10) // default Jakarta

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap contributors',
    }).addTo(map)

    routeGroup = L.featureGroup().addTo(map)
    markerGroup = L.featureGroup().addTo(map)

    // Listen to zoom changes to dynamically scale markers up & down
    map.on('zoomend', () => {
      renderMapElements(false)
    })

    window.addEventListener('resize', handleMapResize)
  }

  renderMapElements(true)
}

onUnmounted(() => {
  if (visitsAbortController) visitsAbortController.abort()
  if (transactionsAbortController) transactionsAbortController.abort()
  if (rekapAbortController) rekapAbortController.abort()
  if (plansAbortController) plansAbortController.abort()
  window.removeEventListener('resize', handleMapResize)
  if (map) {
    map.remove()
    map = null
    markerGroup = null
    routeGroup = null
  }
})

watch(activeTab, async (newVal) => {
  fetchActiveTabData()
  if (newVal === 'map') {
    await nextTick()

    // Initialize map first if not yet created
    initMap()

    // Give browser a moment to apply CSS layout before invalidating size
    setTimeout(() => {
      if (map) {
        map.invalidateSize()
        fitMapToBounds()
      }
    }, 100)
  }
})

watch(filteredVisits, () => {
  if (activeTab.value === 'map') {
    if (!map) {
      initMap()
    } else {
      renderMapElements(true)
    }
  }
})

// ─── Merchandising MT State & Handlers ───
const mdSubmissions = ref<any[]>([])
const keyAccounts = ref<any[]>([])
const isLoadingMD = ref(false)
const mdSearchQuery = ref('')
const selectedMDStatus = ref<'ALL' | 'COMPLIANT' | 'NON_COMPLIANT' | 'PENDING'>('ALL')
const previewImageUrl = ref('')
const zoomLevel = ref(1)

const openPreview = (url: string) => {
  previewImageUrl.value = url
  zoomLevel.value = 1
}
const closePreview = () => {
  previewImageUrl.value = ''
}
const zoomIn = () => {
  if (zoomLevel.value < 4) zoomLevel.value += 0.25
}
const zoomOut = () => {
  if (zoomLevel.value > 0.5) zoomLevel.value -= 0.25
}

const fetchMerchandising = async () => {
  isLoadingMD.value = true
  try {
    const res = await apiClient.get('/modern-trade/merchandising?all_branches=true')
    mdSubmissions.value = (res.data.data || []).map((sub: any) => {
      let oosList: string[] = []
      if (sub.out_of_stock_skus) {
        try {
          oosList = JSON.parse(sub.out_of_stock_skus)
        } catch {
          oosList = sub.out_of_stock_skus.split(',').map((s: string) => s.trim()).filter(Boolean)
        }
      }
      const cleanUrl = (url: string) => {
        if (!url) return ''
        let cleaned = url.replace(/([^:]\/)\/+/g, '$1')
        cleaned = cleaned.replace('/v1/uploads', '/api/v1/uploads')
        return cleaned
      }

      return {
        ...sub,
        store_name: sub.store_name || 'Ritel Modern Store',
        visit_time: sub.visit_time ? sub.visit_time.replace('T', ' ').substring(0, 16) : '',
        out_of_stock_skus: oosList,
        photo_url: cleanUrl(sub.photo_url),
        invoice_photo_url: cleanUrl(sub.invoice_photo_url),
        receipt_photo_url: cleanUrl(sub.receipt_photo_url),
      }
    })
  } catch (err: any) {
    console.error('Failed to fetch Merchandising:', err)
  } finally {
    isLoadingMD.value = false
  }
}

const fetchKeyAccounts = async () => {
  try {
    const res = await apiClient.get('/modern-trade/key-accounts')
    keyAccounts.value = res.data.data || []
  } catch (err: any) {
    console.error('Failed to fetch Key Accounts:', err)
  }
}

const filteredMDSubmissions = computed(() => {
  return mdSubmissions.value.filter((sub) => {
    const matchesSearch =
      !mdSearchQuery.value ||
      sub.merchandiser_name?.toLowerCase().includes(mdSearchQuery.value.toLowerCase()) ||
      sub.store_name?.toLowerCase().includes(mdSearchQuery.value.toLowerCase()) ||
      sub.supermarket_name?.toLowerCase().includes(mdSearchQuery.value.toLowerCase())
    const matchesStatus =
      selectedMDStatus.value === 'ALL' || sub.status === selectedMDStatus.value
    return matchesSearch && matchesStatus
  })
})

const totalMDLaporan = computed(() => mdSubmissions.value.length)
const totalMDOOS = computed(() => mdSubmissions.value.filter(s => s.out_of_stock_skus && s.out_of_stock_skus.length > 0).length)
const totalMDCompliant = computed(() => mdSubmissions.value.filter(s => s.planogram_compliant).length)

// Repeat Order Action from Merchandising
const createRepeatOrder = async (sub: any) => {
  if (!confirm(`Buat Repeat Order (Purchase Order) untuk toko ${sub.store_name}?`)) return
  try {
    const poDate = new Date()
    poDate.setDate(poDate.getDate() + 7)
    
    const payload = {
      key_account_id: sub.key_account_id,
      po_number: `PO-REP-${Date.now()}`,
      expected_delivery: poDate.toISOString().substring(0, 10),
      total_amount: 0,
      status: 'DRAFT',
      invoice_photo_url: sub.invoice_photo_url,
      receipt_photo_url: sub.receipt_photo_url
    }
    
    await apiClient.post('/modern-trade/purchase-orders', payload)
    toast.success('Repeat Order berhasil dibuat (Draft Purchase Order)!')
  } catch (err: any) {
    toast.error('Gagal membuat Repeat Order: ' + (err.response?.data?.error || err.message))
  }
}

// Modal Merchandising Form
const showMDModal = ref(false)
const isEditingMD = ref(false)
const isSavingMD = ref(false)
const isUploadingMDPhoto = ref(false)
const mdForm = ref({
  id: '',
  key_account_id: '',
  merchandiser_name: '',
  store_name: '',
  visit_time: '',
  shelf_share_percent: 0,
  planogram_compliant: true,
  out_of_stock_skus: '',
  photo_url: '',
  supermarket_name: '',
  md_name: '',
  stock_update_log: '',
  promo_update_log: '',
  invoice_photo_url: '',
  receipt_photo_url: '',
})

const openNewMDModal = () => {
  isEditingMD.value = false
  const now = new Date().toISOString().substring(0, 16)
  mdForm.value = {
    id: '',
    key_account_id: keyAccounts.value[0]?.id || '',
    merchandiser_name: '',
    store_name: '',
    visit_time: now,
    shelf_share_percent: 0,
    planogram_compliant: true,
    out_of_stock_skus: '',
    photo_url: 'https://images.unsplash.com/photo-1534482421-64566f976cfa?auto=format&fit=crop&w=600&q=80',
    supermarket_name: '',
    md_name: '',
    stock_update_log: '',
    promo_update_log: '',
    invoice_photo_url: '',
    receipt_photo_url: '',
  }
  showMDModal.value = true
}

const openEditMDModal = (sub: any) => {
  isEditingMD.value = true
  mdForm.value = {
    ...sub,
    out_of_stock_skus: Array.isArray(sub.out_of_stock_skus) ? sub.out_of_stock_skus.join(', ') : sub.out_of_stock_skus,
    visit_time: sub.visit_time ? sub.visit_time.replace(' ', 'T').substring(0, 16) : '',
  }
  showMDModal.value = true
}

const saveMD = async () => {
  if (!mdForm.value.merchandiser_name || !mdForm.value.store_name) {
    toast.error('Nama Merchandiser dan Nama Toko wajib diisi!')
    return
  }

  isSavingMD.value = true
  try {
    const payload = {
      ...mdForm.value,
      shelf_share_percent: Number(mdForm.value.shelf_share_percent),
    }

    if (isEditingMD.value) {
      await apiClient.put(`/modern-trade/merchandising/${mdForm.value.id}`, payload)
      toast.success('Laporan Merchandising berhasil diperbarui!')
    } else {
      await apiClient.post('/modern-trade/merchandising', payload)
      toast.success('Laporan Merchandising baru berhasil disimpan!')
    }
    showMDModal.value = false
    fetchMerchandising()
  } catch (err: any) {
    toast.error('Gagal menyimpan laporan merchandising: ' + (err.response?.data?.error || err.message))
  } finally {
    isSavingMD.value = false
  }
}

const deleteMD = async (id: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus data merchandising ini?')) return
  try {
    await apiClient.delete(`/modern-trade/merchandising/${id}?all_branches=true`)
    toast.success('Data merchandising berhasil dihapus')
    fetchMerchandising()
  } catch (err: any) {
    toast.error('Gagal menghapus data: ' + (err.response?.data?.error || err.message))
  }
}

const uploadMDPhoto = async (event: Event, field: 'photo_url' | 'invoice_photo_url' | 'receipt_photo_url') => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  isUploadingMDPhoto.value = true
  try {
    const res = await apiClient.post('/modern-trade/merchandising/upload-photo', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
    const rawUrl = baseUrl + res.data.url
    mdForm.value[field] = rawUrl.replace(/([^:]\/)\/+/g, '$1')
    toast.success('Foto berhasil diunggah')
  } catch (err: any) {
    toast.error('Gagal mengunggah foto: ' + (err.response?.data?.error || err.message))
  } finally {
    isUploadingMDPhoto.value = false
    target.value = ''
  }
}

watch(activeTab, (newTab) => {
  if (newTab === 'merchandising') {
    fetchMerchandising()
    fetchKeyAccounts()
  }
})

const selectedVisit = ref<SalesVisit | null>(null)
const showDetail = ref(false)
const showFormModal = ref(false)
const isEditing = ref(false)

const actualCheckinAddress = ref('Memuat lokasi GPS aktual...')
const isGeocodingCheckin = ref(false)

const fetchActualCheckinAddress = async (lat: number, lng: number) => {
  if (!lat || !lng || (lat === 0 && lng === 0)) {
    actualCheckinAddress.value = 'Titik koordinat tidak tersedia (0.0, 0.0)'
    return
  }
  
  actualCheckinAddress.value = `GPS: ${lat.toFixed(5)}, ${lng.toFixed(5)} (Memuat alamat...)`
  isGeocodingCheckin.value = true
  
  try {
    const res = await fetch(`https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}`)
    const data = await res.json()
    if (data.display_name) {
      actualCheckinAddress.value = data.display_name
    }
  } catch (error) {
    actualCheckinAddress.value = `Gagal memuat alamat (GPS: ${lat}, ${lng})`
  } finally {
    isGeocodingCheckin.value = false
  }
}

const openGoogleStreetView = (lat?: number, lng?: number) => {
  if (!lat || !lng || (lat === 0 && lng === 0)) {
    toast.error('Titik koordinat GPS tidak valid atau belum tercatat')
    return
  }
  window.open(`https://www.google.com/maps/@?api=1&map_action=pano&viewpoint=${lat},${lng}`, '_blank')
}

const openGoogleMaps = (lat?: number, lng?: number, queryName?: string) => {
  if (!lat || !lng || (lat === 0 && lng === 0)) {
    toast.error('Titik koordinat GPS tidak valid atau belum tercatat')
    return
  }
  const query = queryName ? encodeURIComponent(`${lat},${lng} (${queryName})`) : `${lat},${lng}`
  window.open(`https://www.google.com/maps/search/?api=1&query=${query}`, '_blank')
}

watch(selectedVisit, (newVisit) => {
  if (newVisit) {
    fetchActualCheckinAddress(newVisit.latitude, newVisit.longitude)
  }
})

const form = ref({
  id: '',
  employee_id: '',
  store_id: '',
  check_time: formatDateISO(new Date()) + 'T08:00',
  type: 'CHECKIN',
  notes: '',
  latitude: 0,
  longitude: 0,
  selfie_url: '',
})

const manualEntryMode = ref<'single' | 'batch'>('single')
const isSavingSingle = ref(false)
const isSavingBatch = ref(false)
const isRealizingPlan = ref<string | null>(null)

// Single Mode state
const singleStoreSearch = ref('')
const isSingleStoreDropdownOpen = ref(false)
const singleStoreDropdownRef = ref<HTMLElement | null>(null)

onClickOutside(singleStoreDropdownRef, () => {
  isSingleStoreDropdownOpen.value = false
})

// Batch Mode state
const batchEmployeeId = ref('')
const batchDate = ref(todayLocal)
const batchStartTime = ref('08:30')
const batchIntervalMinutes = ref(30)
const batchStores = ref<Array<{
  id: string
  name: string
  address: string
  latitude: number
  longitude: number
  notes: string
}>>([])
const batchStoreSearch = ref('')
const isBatchStoreDropdownOpen = ref(false)
const batchStoreDropdownRef = ref<HTMLElement | null>(null)

onClickOutside(batchStoreDropdownRef, () => {
  isBatchStoreDropdownOpen.value = false
})

const getAvailableStoresForEmployee = (employeeId?: string) => {
  let list = masterStore.stores || []
  if (employeeId) {
    const emp = masterStore.employees.find((e: any) => e.id === employeeId)
    const empEffectiveBranch = emp?.operating_branch_id || emp?.branch_id
    const assignedStores = list.filter((s: any) => s.assigned_employee_id === employeeId)
    const branchStores = empEffectiveBranch ? list.filter((s: any) => s.branch_id === empEffectiveBranch && s.assigned_employee_id !== employeeId) : []
    const otherStores = list.filter((s: any) => s.assigned_employee_id !== employeeId && (!empEffectiveBranch || s.branch_id !== empEffectiveBranch))
    return [...assignedStores, ...branchStores, ...otherStores]
  }
  return list
}

const filteredStoresSingle = computed(() => {
  const query = singleStoreSearch.value.trim().toLowerCase()
  const base = getAvailableStoresForEmployee(form.value.employee_id)
  if (!query) {
    return base.slice(0, 60)
  }
  return base
    .filter((s: any) => {
      return (
        s.name?.toLowerCase().includes(query) ||
        s.address?.toLowerCase().includes(query) ||
        s.category?.toLowerCase().includes(query) ||
        s.code?.toLowerCase().includes(query) ||
        s.owner_name?.toLowerCase().includes(query) ||
        s.phone?.includes(query)
      )
    })
    .slice(0, 60)
})

const selectedStoreSingleObj = computed(() => {
  if (!form.value.store_id) return null
  return masterStore.stores.find((s: any) => s.id === form.value.store_id) || null
})

const selectStoreSingle = (store: any) => {
  form.value.store_id = store.id
  form.value.latitude = store.latitude || 0
  form.value.longitude = store.longitude || 0
  singleStoreSearch.value = store.name
  isSingleStoreDropdownOpen.value = false
}

const clearSelectedStoreSingle = () => {
  form.value.store_id = ''
  form.value.latitude = 0
  form.value.longitude = 0
  singleStoreSearch.value = ''
}

const applySingleTimeOffset = (minutes: number) => {
  const current = new Date(form.value.check_time)
  if (!isNaN(current.getTime())) {
    current.setMinutes(current.getMinutes() + minutes)
    const y = current.getFullYear()
    const m = String(current.getMonth() + 1).padStart(2, '0')
    const d = String(current.getDate()).padStart(2, '0')
    const hh = String(current.getHours()).padStart(2, '0')
    const mm = String(current.getMinutes()).padStart(2, '0')
    form.value.check_time = `${y}-${m}-${d}T${hh}:${mm}`
  }
}

const setSingleTimeNow = () => {
  const now = new Date()
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  const d = String(now.getDate()).padStart(2, '0')
  const hh = String(now.getHours()).padStart(2, '0')
  const mm = String(now.getMinutes()).padStart(2, '0')
  form.value.check_time = `${y}-${m}-${d}T${hh}:${mm}`
}

const filteredStoresBatch = computed(() => {
  const query = batchStoreSearch.value.trim().toLowerCase()
  const base = getAvailableStoresForEmployee(batchEmployeeId.value)
  const existingIds = new Set(batchStores.value.map(s => s.id))
  
  const unselected = base.filter((s: any) => !existingIds.has(s.id))
  if (!query) {
    return unselected.slice(0, 60)
  }
  return unselected
    .filter((s: any) => {
      return (
        s.name?.toLowerCase().includes(query) ||
        s.address?.toLowerCase().includes(query) ||
        s.category?.toLowerCase().includes(query) ||
        s.code?.toLowerCase().includes(query) ||
        s.owner_name?.toLowerCase().includes(query)
      )
    })
    .slice(0, 60)
})

const addStoreToBatch = (store: any) => {
  if (batchStores.value.some(s => s.id === store.id)) {
    toast.error('Toko sudah ada di dalam daftar batch')
    return
  }
  batchStores.value.push({
    id: store.id,
    name: store.name,
    address: store.address || '',
    latitude: store.latitude || 0,
    longitude: store.longitude || 0,
    notes: '',
  })
  batchStoreSearch.value = ''
  isBatchStoreDropdownOpen.value = false
}

const removeStoreFromBatch = (index: number) => {
  batchStores.value.splice(index, 1)
}

const moveBatchStore = (index: number, direction: 'up' | 'down') => {
  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= batchStores.value.length) return
  const item = batchStores.value[index]
  if (!item) return
  batchStores.value.splice(index, 1)
  batchStores.value.splice(targetIndex, 0, item)
}

const calculatedBatchTimeline = computed(() => {
  const parts = (batchStartTime.value || '08:30').split(':')
  const startH = Number(parts[0] ?? 8)
  const startM = Number(parts[1] ?? 30)
  const interval = Number(batchIntervalMinutes.value) || 30
  
  return batchStores.value.map((store, idx) => {
    const validH = isNaN(startH) ? 8 : startH
    const validM = isNaN(startM) ? 30 : startM
    const totalMinutes = (validH * 60 + validM) + (idx * interval)
    const hours = Math.floor(totalMinutes / 60) % 24
    const mins = totalMinutes % 60
    const timeStr = `${String(hours).padStart(2, '0')}:${String(mins).padStart(2, '0')}`
    return {
      ...store,
      timeStr,
    }
  })
})

const openDetail = (visit: SalesVisit) => {
  selectedVisit.value = visit
  showDetail.value = true
}

const openCreate = () => {
  isEditing.value = false
  manualEntryMode.value = 'single'
  
  // Auto-fill salesman from active filter if filtered
  const defaultEmp = selectedEmployee.value !== 'ALL' ? selectedEmployee.value : (salesmen.value[0]?.id || '')
  
  // Default check_time: today + current hour rounded or active filter date
  const now = new Date()
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(Math.floor(now.getMinutes() / 5) * 5).padStart(2, '0')
  const datePart = filterMode.value === 'date' && selectedDate.value ? selectedDate.value : formatDateISO(now)
  
  form.value = {
    id: '',
    employee_id: defaultEmp,
    store_id: '',
    check_time: `${datePart}T${hours}:${minutes}`,
    type: 'CHECKIN',
    notes: '',
    latitude: 0,
    longitude: 0,
    selfie_url: '',
  }
  singleStoreSearch.value = ''
  isSingleStoreDropdownOpen.value = false

  // Initialize batch mode form
  batchEmployeeId.value = defaultEmp
  batchDate.value = datePart
  batchStartTime.value = `${hours}:${minutes}`
  batchIntervalMinutes.value = 30
  batchStores.value = []
  batchStoreSearch.value = ''
  isBatchStoreDropdownOpen.value = false

  masterStore.fetchStores(masterStore.selectedBranchCompanyId)
  showFormModal.value = true
}

const openEdit = (visit: SalesVisit) => {
  isEditing.value = true
  manualEntryMode.value = 'single'
  form.value = {
    id: visit.id,
    employee_id: visit.employee_id,
    store_id: visit.store_id,
    check_time: visit.check_in_time.substring(0, 16),
    type: visit.type || 'CHECKIN',
    notes: visit.notes || '',
    latitude: visit.latitude || 0,
    longitude: visit.longitude || 0,
    selfie_url: visit.selfie_url || '',
  }
  const store = masterStore.stores.find((s: any) => s.id === visit.store_id)
  singleStoreSearch.value = store ? store.name : ''
  isSingleStoreDropdownOpen.value = false
  masterStore.fetchStores(masterStore.selectedBranchCompanyId)
  showFormModal.value = true
}

const handleSaveSingle = async (keepOpenAndAddAnother = false) => {
  if (!form.value.employee_id) {
    toast.error('Silakan pilih salesman')
    return
  }
  if (!form.value.store_id) {
    toast.error('Silakan pilih toko')
    return
  }
  if (!form.value.check_time) {
    toast.error('Silakan tentukan waktu kunjungan')
    return
  }

  isSavingSingle.value = true
  try {
    const payload = {
      ...form.value,
      check_time: new Date(form.value.check_time).toISOString(),
    }

    const selectedStore = masterStore.stores.find((s: any) => s.id === form.value.store_id)
    const storeName = selectedStore?.name || 'Toko'

    if (isEditing.value) {
      await visitStore.updateVisit(form.value.id, payload)
      toast.success(`Kunjungan ${storeName} berhasil diperbarui`)
      showFormModal.value = false
    } else {
      await visitStore.createVisit(payload)
      if (keepOpenAndAddAnother) {
        toast.success(`Kunjungan ke ${storeName} tersimpan. Silakan pilih toko berikutnya!`)
        applySingleTimeOffset(30)
        form.value.store_id = ''
        form.value.notes = ''
        form.value.latitude = 0
        form.value.longitude = 0
        singleStoreSearch.value = ''
        isSingleStoreDropdownOpen.value = true
      } else {
        toast.success(`Kunjungan ke ${storeName} berhasil dibuat`)
        showFormModal.value = false
      }
    }

    invalidateVisitCaches()
    fetchVisits(true)
    fetchRekapVisits(true)
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menyimpan data')
  } finally {
    isSavingSingle.value = false
  }
}

const handleSubmit = async () => {
  await handleSaveSingle(false)
}

const handleSaveBatch = async () => {
  if (!batchEmployeeId.value) {
    toast.error('Silakan pilih salesman')
    return
  }
  if (!batchDate.value) {
    toast.error('Silakan pilih tanggal kunjungan')
    return
  }
  if (batchStores.value.length === 0) {
    toast.error('Pilih minimal 1 toko untuk disimpan')
    return
  }

  isSavingBatch.value = true
  try {
    const timeline = calculatedBatchTimeline.value
    const payloads = timeline.map(item => {
      const [h, m] = item.timeStr.split(':')
      const visitDate = new Date(`${batchDate.value}T${h}:${m}:00`)
      return {
        employee_id: batchEmployeeId.value,
        store_id: item.id,
        check_time: visitDate.toISOString(),
        type: 'CHECKIN',
        latitude: item.latitude || 0,
        longitude: item.longitude || 0,
        notes: item.notes || 'Input Batch Manual Admin',
      }
    })

    await visitStore.createBatchVisits(payloads)
    toast.success(`${payloads.length} kunjungan berhasil dicatat sekaligus`)
    showFormModal.value = false
    invalidateVisitCaches()
    fetchVisits(true)
    fetchRekapVisits(true)
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menyimpan batch kunjungan')
  } finally {
    isSavingBatch.value = false
  }
}

const realizePlan = async (plan: any) => {
  if (!selectedEmployee.value || selectedEmployee.value === 'ALL') {
    toast.error('Salesman tidak terdeteksi')
    return
  }

  isRealizingPlan.value = plan.store_id
  try {
    const store = masterStore.stores.find((s: any) => s.id === plan.store_id)
    const now = new Date()
    const payload = {
      employee_id: selectedEmployee.value,
      store_id: plan.store_id,
      check_time: now.toISOString(),
      type: 'CHECKIN',
      latitude: store?.latitude || 0,
      longitude: store?.longitude || 0,
      notes: 'Realisasi dari Jadwal Rencana Kunjungan',
    }

    await visitStore.createVisit(payload)
    toast.success(`Kunjungan ke ${plan.store_name} berhasil direalisasikan!`)
    invalidateVisitCaches()
    fetchVisitPlans(true)
    fetchVisits(true)
    fetchRekapVisits(true)
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal merealisasikan kunjungan')
  } finally {
    isRealizingPlan.value = null
  }
}

const handleDelete = async (visit: any) => {
  let warningMessage = 'Apakah Anda yakin ingin menghapus data kunjungan ini?'
  
  const trxCount = visit.transactions?.length || 0
  const soCount = visit.sales_orders?.length || 0
  
  if (trxCount > 0 || soCount > 0) {
    warningMessage += '\n\nPERHATIAN: Menghapus kunjungan ini juga akan ikut menghapus secara permanen:'
    if (trxCount > 0) warningMessage += `\n- ${trxCount} Data Transaksi`
    if (soCount > 0) warningMessage += `\n- ${soCount} Data Pesanan (Sales Order)`
    warningMessage += '\n\nLanjutkan penghapusan?'
  }

  if (!confirm(warningMessage)) return

  try {
    await visitStore.deleteVisit(visit.id)
    toast.success('Kunjungan berhasil dihapus')
    invalidateVisitCaches()
    fetchVisits(true)
    fetchRekapVisits(true)
    fetchTransactions(true)
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menghapus data')
  }
}

const getDuration = (visit: SalesVisit) => {
  if (!visit.check_out_time) return 'Active'
  const start = new Date(visit.check_in_time)
  const end = new Date(visit.check_out_time)
  const diff = Math.floor((end.getTime() - start.getTime()) / 1000 / 60)
  return `${diff} menit`
}

const getStatusColor = (visit: SalesVisit) => {
  if (visit.check_out_time) return 'text-emerald-600 bg-emerald-50 border-emerald-100'
  return 'text-amber-600 bg-amber-50 border-amber-100'
}

const isExporting = ref(false)
const exportToExcel = async () => {
  if (isExporting.value) return
  isExporting.value = true

  try {
    let start = '', end = ''
    if (filterMode.value === 'date') {
      const dateVal = selectedDate.value || todayLocal || ''
      start = dateVal
      end = dateVal
    } else if (filterMode.value === 'month') {
      const monthVal = selectedMonth.value || (new Date().getFullYear() + '-' + String(new Date().getMonth() + 1).padStart(2, '0'))
      start = `${monthVal}-01`
      const [year, month] = monthVal.split('-')
      const lastDay = new Date(parseInt(year || '2026'), parseInt(month || '1'), 0).getDate()
      end = `${monthVal}-${String(lastDay).padStart(2, '0')}`
    }

    const params: any = {
      start_date: start,
      end_date: end,
      employee_id: selectedEmployee.value === 'ALL' ? undefined : selectedEmployee.value,
      all: 'true'
    }

    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterStore.selectedBranchId
    }
    if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
      params.sales_type = masterStore.selectedSalesType
    }

    const res = await apiClient.get('/admin/sales/visits', { params })
    const rawVisits = res.data.data || []

    const isDateMatch = (dateStr: string | undefined | null) => {
      if (!dateStr) return false
      const itemDate = typeof dateStr === 'string' && dateStr.length >= 10 ? dateStr.substring(0, 10) : formatDateISO(new Date(dateStr))
      if (filterMode.value === 'date') {
        return itemDate === start
      } else if (filterMode.value === 'month') {
        const m = selectedMonth.value || (start ? start.substring(0, 7) : '')
        return itemDate.startsWith(m)
      }
      return true
    }

    const visitSeen = new Set<string>()
    const visits = rawVisits.filter((v: any) => {
      if (!isDateMatch(v.check_in_time)) return false
      const timeMin = v.check_in_time ? v.check_in_time.substring(0, 16) : ''
      const key = `${v.employee_id}_${v.store_id}_${timeMin}`
      if (visitSeen.has(key)) {
        return false
      }
      visitSeen.add(key)
      return true
    })

    // Sort visits by Salesman name alphabetically
    visits.sort((a: any, b: any) => {
      const nameA = `${a.employee?.first_name || ''} ${a.employee?.last_name || ''}`.trim().toLowerCase()
      const nameB = `${b.employee?.first_name || ''} ${b.employee?.last_name || ''}`.trim().toLowerCase()
      if (nameA < nameB) return -1
      if (nameA > nameB) return 1
      
      // Secondary sort by check-in time descending
      const timeA = new Date(a.check_in_time || 0).getTime()
      const timeB = new Date(b.check_in_time || 0).getTime()
      return timeB - timeA
    })

    if (visits.length === 0) {
      toast.error('Tidak ada data kunjungan untuk di-export')
      return
    }

    let exportTransactions: any[] = []
    try {
      const trxParams = new URLSearchParams()
      trxParams.append('limit', '-1')
      if (filterMode.value === 'date') {
        trxParams.append('date', start)
      } else if (filterMode.value === 'month' && selectedMonth.value) {
        trxParams.append('month', selectedMonth.value)
      } else if (start) {
        trxParams.append('month', start.substring(0, 7))
      }

      if (selectedEmployee.value !== 'ALL') {
        trxParams.append('salesman', selectedEmployee.value)
      }
      if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
        trxParams.append('branch_id', masterStore.selectedBranchId)
      }
      if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
        trxParams.append('sales_type', masterStore.selectedSalesType)
      }
      
      const trxRes = await apiClient.get(`/admin/sales/transactions/all?${trxParams.toString()}`)
      exportTransactions = trxRes.data.data || []
    } catch (err) {
      console.error('Failed to fetch transactions for export:', err)
    }

    let salesOrders: any[] = []
    try {
      const soParams = new URLSearchParams()
      soParams.append('limit', '5000')
      
      if (filterMode.value === 'date') {
        soParams.append('date', start)
      } else if (filterMode.value === 'month' && selectedMonth.value) {
        soParams.append('month', selectedMonth.value)
      } else if (start) {
        soParams.append('month', start.substring(0, 7))
      }

      if (selectedEmployee.value !== 'ALL') {
        soParams.append('employee_id', selectedEmployee.value)
      }
      if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
        soParams.append('branch_id', masterStore.selectedBranchId)
      }
      if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
        soParams.append('sales_type', masterStore.selectedSalesType)
      }
      const soRes = await apiClient.get(`/admin/sales-orders?${soParams.toString()}`)
      salesOrders = soRes.data.data || []
    } catch (soError) {
      console.error('Failed to fetch sales orders:', soError)
    }

    // Fetch fresh journeys specifically for this export date range
    let exportJourneys: any[] = []
    try {
      const jParams = new URLSearchParams()
      if (start && end) {
        jParams.append('start_date', start)
        jParams.append('end_date', end)
      }
      if (selectedEmployee.value !== 'ALL') {
        jParams.append('employee_id', selectedEmployee.value)
      }
      if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
        jParams.append('branch_id', masterStore.selectedBranchId)
      }
      const jRes = await apiClient.get('/admin/sales/journeys?' + jParams.toString())
      exportJourneys = jRes.data?.data || []
    } catch (jErr) {
      console.error('Failed to fetch journeys for export:', jErr)
      exportJourneys = journeys.value || []
    }

    // Fetch cashier payments (cash & bank transfer) specifically for this export date range (approved only)
    let exportCashPayments: any[] = []
    let exportTransferPayments: any[] = []
    try {
      const pParams: any = {
        limit: -1,
        status: 'success',
        start_date: start,
        end_date: end,
      }
      const hasSelectedBranch = masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES'
      if (hasSelectedBranch) {
        pParams.branch_id = masterStore.selectedBranchId
      } else if (masterStore.selectedBranchCompanyId && masterStore.selectedBranchCompanyId !== 'ALL') {
        pParams.company_id = masterStore.selectedBranchCompanyId
      }

      const [cashRes, tfRes] = await Promise.all([
        apiClient.get('/finance/payments/pending', { params: pParams }),
        apiClient.get('/finance/payments/midtrans', { params: pParams })
      ])

      exportCashPayments = cashRes.data?.data || []
      exportTransferPayments = tfRes.data?.data || []
    } catch (payErr) {
      console.error('Failed to fetch cashier payments for export:', payErr)
      throw new Error('Gagal mengambil seluruh data persetujuan kasir')
    }

    const summaryMap = new Map<string, any>()
    salesmen.value.forEach((emp: any) => {
      if (selectedEmployee.value !== 'ALL' && emp.id !== selectedEmployee.value) return;
      const key = String(emp.id).toLowerCase()
      summaryMap.set(key, {
        name: `${emp.first_name || ''} ${emp.last_name || ''}`.trim().toUpperCase(),
        totalKunjungan: 0,
        tokoBaru: 0,
        tokoLama: 0,
        trxCount: 0,
        trxNominal: 0,
        tunaiNominal: 0,
        transferNominal: 0,
        tempoNominal: 0,
        tempoPaidNominal: 0
      })
    })

    const getOrCreateSummary = (empId: any, empObj: any) => {
      const key = empId ? String(empId).toLowerCase() : 'other_sales'
      if (!summaryMap.has(key)) {
        let name = ''
        if (empObj) {
          name = `${empObj.first_name || ''} ${empObj.last_name || ''}`.trim()
        }
        if (!name) name = 'SALES LAINNYA'
        summaryMap.set(key, {
          name: name.toUpperCase(),
          totalKunjungan: 0,
          tokoBaru: 0,
          tokoLama: 0,
          trxCount: 0,
          trxNominal: 0,
          tunaiNominal: 0,
          transferNominal: 0,
          tempoNominal: 0,
          tempoPaidNominal: 0
        })
      }
      return summaryMap.get(key)
    }

    const workbook = new ExcelJS.Workbook()
    const worksheet = workbook.addWorksheet('Data Kunjungan')

    worksheet.mergeCells('A1:J1')
    const titleCell = worksheet.getCell('A1')
    titleCell.value = 'REKAPITULASI KUNJUNGAN & TRANSAKSI SALES'
    titleCell.font = { name: 'Calibri', size: 14, bold: true }
    titleCell.alignment = { vertical: 'middle', horizontal: 'center' }
    worksheet.getRow(1).height = 25

    const headers = [
      'NO', 'TANGGAL & WAKTU', 'NAMA SALES', 'CABANG', 'KUNJUNGAN TOKO', 
      'STATUS / KETERANGAN', 'NAMA PRODUK', 'QTY', 'NOMINAL (RP)', 'TOTAL PENJUALAN'
    ]
    const headerRow = worksheet.addRow(headers)
    
    headerRow.eachCell((cell) => {
      cell.fill = {
        type: 'pattern',
        pattern: 'solid',
        fgColor: { argb: 'FF1E293B' } 
      }
      cell.font = { bold: true, name: 'Calibri', size: 11, color: { argb: 'FFFFFFFF' } }
      cell.border = {
        top: { style: 'thin' },
        left: { style: 'thin' },
        bottom: { style: 'thin' },
        right: { style: 'thin' }
      }
      cell.alignment = { vertical: 'middle', horizontal: 'center', wrapText: true }
    })
    worksheet.getRow(2).height = 20

    let rowNum = 1
    let grandTotalSales = 0
    let visitIndex = 0
    const processedTrxIds = new Set<string>()
    const processedSoIds = new Set<string>()

    const classifyAndAccumulatePayment = (s: any, method: string, totalAmount: number, paidAmount: number, paymentStatus?: string) => {
      if (!s) return
      const m = (method || 'CASH').toUpperCase()
      const total = Number(totalAmount) || 0
      const paid = Number(paidAmount) || 0

      if (m === 'CASH') {
        s.tunaiNominal += total
      } else if (m === 'TRANSFER' || m === 'VA' || m === 'QRIS' || m === 'MIDTRANS_QRIS') {
        s.transferNominal += total
      } else if (m === 'TEMPO') {
        s.tempoNominal += total
        // If payment status is PAID, the entire tempo is considered paid
        if (paymentStatus === 'PAID') {
          s.tempoPaidNominal += total
        } else {
          s.tempoPaidNominal += Math.min(total, Math.max(0, paid))
        }
      } else {
        // Fallback for other methods like POTONG_GAJI or unknown
        s.tunaiNominal += total
      }
    }

    visits.forEach((visit: any) => {
      if (selectedEmployee.value !== 'ALL' && visit.employee_id !== selectedEmployee.value) return;

      visitIndex++
      const visitDate = visit.check_in_time ? new Date(visit.check_in_time) : new Date()
      const formattedDate = visitDate.toLocaleString('id-ID', {
        day: '2-digit', month: '2-digit', year: 'numeric',
        hour: '2-digit', minute: '2-digit', second: '2-digit'
      })

      // Combine transactions from visit preloaded data and fetched exportTransactions
      const visitTrxMap = new Map<string, any>()
      if (Array.isArray(visit.transactions)) {
        visit.transactions.forEach((t: any) => {
          if (t && (t.id || t.receipt_no)) {
            const key = String(t.id || t.receipt_no)
            visitTrxMap.set(key, t)
            processedTrxIds.add(key)
            if (t.id) processedTrxIds.add(String(t.id))
            if (t.receipt_no) processedTrxIds.add(String(t.receipt_no))
          }
        })
      }
      exportTransactions.filter(t => t.visit_id === visit.id).forEach((t: any) => {
        const key = String(t.id || t.receipt_no)
        if (key && !visitTrxMap.has(key)) {
          visitTrxMap.set(key, t)
          processedTrxIds.add(key)
          if (t.id) processedTrxIds.add(String(t.id))
          if (t.receipt_no) processedTrxIds.add(String(t.receipt_no))
        }
      })
      const visitTransactions = Array.from(visitTrxMap.values())

      // Combine sales orders from visit preloaded data and fetched salesOrders
      const visitSOMap = new Map<string, any>()
      if (Array.isArray(visit.sales_orders)) {
        visit.sales_orders.forEach((so: any) => {
          if (so && (so.id || so.so_number)) {
            if (so.status === 'CANCELLED' || so.status === 'REJECTED') return
            // If this SO has already been converted to an invoice that is in processedTrxIds / visitTrxMap, skip it
            if (so.invoice_id && (processedTrxIds.has(String(so.invoice_id)) || visitTrxMap.has(String(so.invoice_id)))) return
            if (so.invoice?.id && (processedTrxIds.has(String(so.invoice.id)) || visitTrxMap.has(String(so.invoice.id)))) return
            if (so.invoice?.receipt_no && (processedTrxIds.has(String(so.invoice.receipt_no)) || visitTrxMap.has(String(so.invoice.receipt_no)))) return

            const key = String(so.id || so.so_number)
            visitSOMap.set(key, so)
            processedSoIds.add(key)
            if (so.id) processedSoIds.add(String(so.id))
            if (so.so_number) processedSoIds.add(String(so.so_number))
            if (so.invoice_id) processedTrxIds.add(String(so.invoice_id))
            if (so.invoice?.id) processedTrxIds.add(String(so.invoice.id))
            if (so.invoice?.receipt_no) processedTrxIds.add(String(so.invoice.receipt_no))
          }
        })
      }
      salesOrders.filter(so => so.visit_id === visit.id).forEach((so: any) => {
        if (so.status === 'CANCELLED' || so.status === 'REJECTED') return
        if (so.invoice_id && (processedTrxIds.has(String(so.invoice_id)) || visitTrxMap.has(String(so.invoice_id)))) return
        if (so.invoice?.id && (processedTrxIds.has(String(so.invoice.id)) || visitTrxMap.has(String(so.invoice.id)))) return
        if (so.invoice?.receipt_no && (processedTrxIds.has(String(so.invoice.receipt_no)) || visitTrxMap.has(String(so.invoice.receipt_no)))) return

        const key = String(so.id || so.so_number)
        if (key && !visitSOMap.has(key)) {
          visitSOMap.set(key, so)
          processedSoIds.add(key)
          if (so.id) processedSoIds.add(String(so.id))
          if (so.so_number) processedSoIds.add(String(so.so_number))
          if (so.invoice_id) processedTrxIds.add(String(so.invoice_id))
          if (so.invoice?.id) processedTrxIds.add(String(so.invoice.id))
          if (so.invoice?.receipt_no) processedTrxIds.add(String(so.invoice.receipt_no))
        }
      })
      const visitSalesOrders = Array.from(visitSOMap.values())
      
      const allItems: any[] = []

      visitTransactions.forEach((trx: any) => {
        let trxItemNominalSum = 0
        if (!trx.items || trx.items.length === 0) {
          if (trx.total_amount && Number(trx.total_amount) > 0) {
            const nom = Number(trx.total_amount)
            trxItemNominalSum += nom
            allItems.push({
              name: trx.receipt_image_url ? 'Upload Bukti Transaksi' : 'Transaksi Tanpa Rincian Item',
              qty: 1,
              nominal: nom
            })
          }
        } else {
          trx.items.forEach((item: any) => {
            const prodName = item.product?.name || item.product_name || item.name || 'Item'
            const qty = Number(item.quantity || item.ordered_quantity || 0)
            const price = Number(item.price_at_transaction || item.price || 0)
            const nominal = Number(item.subtotal) || (price * qty) || 0
            trxItemNominalSum += nominal
            allItems.push({
              name: prodName,
              qty: qty,
              nominal: nominal
            })
          })
        }
        const trxTotal = trxItemNominalSum > 0 ? trxItemNominalSum : Number(trx.total_amount || 0)
        const s = getOrCreateSummary(visit.employee_id, visit.employee)
        classifyAndAccumulatePayment(s, trx.payment_method, trxTotal, trx.paid_amount, trx.payment_status)
      })

      visitSalesOrders.forEach((so: any) => {
        let soItemNominalSum = 0
        if (!so.items || so.items.length === 0) {
          if (so.total_amount && Number(so.total_amount) > 0) {
            const nom = Number(so.total_amount)
            soItemNominalSum += nom
            allItems.push({
              name: so.physical_nota_image_url || so.is_physical_nota ? 'Upload Nota Fisik (SO)' : 'Pesanan Tanpa Rincian Item',
              qty: 1,
              nominal: nom
            })
          }
        } else {
          so.items.forEach((item: any) => {
            const prodName = item.product?.name || item.product_name || item.name || 'Item'
            const qty = Number(item.ordered_quantity || item.quantity || 0)
            const price = Number(item.price || 0)
            const nominal = Number(item.subtotal) || (price * qty) || 0
            soItemNominalSum += nominal
            allItems.push({
              name: prodName,
              qty: qty,
              nominal: nominal
            })
          })
        }
        const soTotal = soItemNominalSum > 0 ? soItemNominalSum : Number(so.total_amount || 0)
        const soPaid = Number(so.payment_collected_amount || so.paid_amount || 0)
        const s = getOrCreateSummary(visit.employee_id, visit.employee)
        classifyAndAccumulatePayment(s, so.payment_method, soTotal, soPaid, so.payment_status)
      })

      let totalSales = allItems.reduce((sum, item) => sum + (Number(item.nominal) || 0), 0)
      grandTotalSales += totalSales
      
      const s = getOrCreateSummary(visit.employee_id, visit.employee)
      if (s) {
        s.totalKunjungan += 1
        if (visit.store?.category === 'TOKO_BARU') s.tokoBaru += 1
        else s.tokoLama += 1
        s.trxCount += (visitTransactions.length + visitSalesOrders.length)
        s.trxNominal += totalSales
      }

      let statusKeterangan = visit.store?.category === 'TOKO_BARU' ? 'TOKO BARU' : 'KUNJUNGAN / FOLLOW UP'
      if (visit.notes) {
        statusKeterangan += ` - ${visit.notes}`
      }

      if (allItems.length === 0) {
        const row = worksheet.addRow([
          rowNum++,
          formattedDate,
          `${visit.employee?.first_name || ''} ${visit.employee?.last_name || ''}`.trim().toUpperCase(),
          (visit.employee?.branch?.name || '').toUpperCase(),
          (visit.store?.name || '').toUpperCase(),
          statusKeterangan.toUpperCase(),
          '-',
          0,
          0,
          0
        ])
        applyRowStyle(row, visitIndex, statusKeterangan.toUpperCase(), true)
      } else {
        const startRowIndex = worksheet.rowCount + 1;
        allItems.forEach((item, index) => {
          const row = worksheet.addRow([
            index === 0 ? rowNum++ : '',
            index === 0 ? formattedDate : '',
            index === 0 ? `${visit.employee?.first_name || ''} ${visit.employee?.last_name || ''}`.trim().toUpperCase() : '',
            index === 0 ? (visit.employee?.branch?.name || '').toUpperCase() : '',
            index === 0 ? (visit.store?.name || '').toUpperCase() : '',
            index === 0 ? statusKeterangan.toUpperCase() : '',
            item.name.toUpperCase(),
            item.qty,
            item.nominal,
            index === 0 ? totalSales : ''
          ])
          applyRowStyle(row, visitIndex, statusKeterangan.toUpperCase(), index === 0)
        })
        
        const endRowIndex = worksheet.rowCount;
        if (allItems.length > 1) {
          worksheet.mergeCells(`A${startRowIndex}:A${endRowIndex}`);
          worksheet.mergeCells(`B${startRowIndex}:B${endRowIndex}`);
          worksheet.mergeCells(`C${startRowIndex}:C${endRowIndex}`);
          worksheet.mergeCells(`D${startRowIndex}:D${endRowIndex}`);
          worksheet.mergeCells(`E${startRowIndex}:E${endRowIndex}`);
          worksheet.mergeCells(`F${startRowIndex}:F${endRowIndex}`);
          worksheet.mergeCells(`J${startRowIndex}:J${endRowIndex}`);
        }
      }
    })

    // Process unhandled transactions (transactions without visit checkin)
    const unhandledTransactions = exportTransactions.filter((t: any) => {
      if (selectedEmployee.value !== 'ALL' && t.employee_id !== selectedEmployee.value) return false

      // Exclude cancelled, rejected, or deleted transactions
      if (t.status === 'CANCELLED' || t.status === 'REJECTED' || t.status === 'DELETED') return false

      // Exclude transactions that originated from Sales Orders (delivery invoices for SO)
      if (t.origin_sales_order_id) return false
      if (t.payment_method === 'SALES_ORDER') return false

      // Exclude delivery invoices (items fulfilled via DeliveryBatch)
      if (Array.isArray(t.delivery_items) && t.delivery_items.length > 0) return false

      // Exclude non-salesman channels (office, factory, branch direct, employee sale)
      const channel = String(t.sales_channel || '').toUpperCase()
      if (channel === 'DIRECT_BRANCH' || channel === 'DIRECT_FACTORY' || channel === 'EMPLOYEE_SALE') return false

      const tDate = t.transaction_date || t.created_at
      if (!isDateMatch(tDate)) return false
      const isProcessed = (t.id && processedTrxIds.has(String(t.id))) || (t.receipt_no && processedTrxIds.has(String(t.receipt_no)))
      return !isProcessed
    })

    unhandledTransactions.forEach((trx: any) => {
      visitIndex++
      if (trx.id) processedTrxIds.add(String(trx.id))
      if (trx.receipt_no) processedTrxIds.add(String(trx.receipt_no))

      // Prioritize created_at to preserve actual clock time instead of 07.00.00 from UTC midnight
      const trxDateRaw = trx.created_at || trx.transaction_date
      const trxDate = trxDateRaw ? new Date(trxDateRaw) : new Date()
      const formattedDate = trxDate.toLocaleString('id-ID', {
        day: '2-digit', month: '2-digit', year: 'numeric',
        hour: '2-digit', minute: '2-digit', second: '2-digit'
      })

      const empName = `${trx.employee?.first_name || ''} ${trx.employee?.last_name || ''}`.trim() || 'SALES'
      const branchName = trx.employee?.branch?.name || trx.store?.assigned_employee?.branch?.name || '-'
      const storeName = trx.store?.name || trx.store_name || '-'
      const statusKeterangan = 'ORDER DIRECT / TANPA CHECK-IN'

      const allItems: any[] = []
      if (!trx.items || trx.items.length === 0) {
        if (trx.total_amount && Number(trx.total_amount) > 0) {
          allItems.push({
            name: trx.receipt_image_url ? 'Upload Bukti Transaksi' : 'Transaksi Tanpa Rincian Item',
            qty: 1,
            nominal: Number(trx.total_amount)
          })
        }
      } else {
        trx.items.forEach((item: any) => {
          const prodName = item.product?.name || item.product_name || item.name || 'Item'
          const qty = Number(item.quantity || item.ordered_quantity || 0)
          const price = Number(item.price_at_transaction || item.price || 0)
          const nominal = Number(item.subtotal) || (price * qty) || 0
          allItems.push({
            name: prodName,
            qty: qty,
            nominal: nominal
          })
        })
      }

      let totalSales = allItems.reduce((sum, item) => sum + (Number(item.nominal) || 0), 0)
      if (totalSales === 0 && trx.total_amount) totalSales = Number(trx.total_amount)
      grandTotalSales += totalSales

      const s = getOrCreateSummary(trx.employee_id, trx.employee)
      if (s) {
        s.trxCount += 1
        s.trxNominal += totalSales
        classifyAndAccumulatePayment(s, trx.payment_method, totalSales, trx.paid_amount, trx.payment_status)
      }

      if (allItems.length === 0) {
        const row = worksheet.addRow([
          rowNum++,
          formattedDate,
          empName.toUpperCase(),
          branchName.toUpperCase(),
          storeName.toUpperCase(),
          statusKeterangan,
          '-',
          0,
          totalSales,
          totalSales
        ])
        applyRowStyle(row, visitIndex, statusKeterangan, true)
      } else {
        const startRowIndex = worksheet.rowCount + 1
        allItems.forEach((item, index) => {
          const row = worksheet.addRow([
            index === 0 ? rowNum++ : '',
            index === 0 ? formattedDate : '',
            index === 0 ? empName.toUpperCase() : '',
            index === 0 ? branchName.toUpperCase() : '',
            index === 0 ? storeName.toUpperCase() : '',
            index === 0 ? statusKeterangan : '',
            item.name.toUpperCase(),
            item.qty,
            item.nominal,
            index === 0 ? totalSales : ''
          ])
          applyRowStyle(row, visitIndex, statusKeterangan, index === 0)
        })
        const endRowIndex = worksheet.rowCount
        if (allItems.length > 1) {
          worksheet.mergeCells(`A${startRowIndex}:A${endRowIndex}`)
          worksheet.mergeCells(`B${startRowIndex}:B${endRowIndex}`)
          worksheet.mergeCells(`C${startRowIndex}:C${endRowIndex}`)
          worksheet.mergeCells(`D${startRowIndex}:D${endRowIndex}`)
          worksheet.mergeCells(`E${startRowIndex}:E${endRowIndex}`)
          worksheet.mergeCells(`F${startRowIndex}:F${endRowIndex}`)
          worksheet.mergeCells(`J${startRowIndex}:J${endRowIndex}`)
        }
      }
    })

    // Process unhandled sales orders (sales orders without visit checkin)
    const unhandledSalesOrders = salesOrders.filter((so: any) => {
      if (selectedEmployee.value !== 'ALL' && so.employee_id !== selectedEmployee.value) return false
      if (so.status === 'CANCELLED' || so.status === 'REJECTED') return false
      const sDate = so.order_date || so.created_at
      if (!isDateMatch(sDate)) return false
      const isProcessed = (so.id && processedSoIds.has(String(so.id))) ||
                          (so.so_number && processedSoIds.has(String(so.so_number))) ||
                          (so.invoice_id && processedTrxIds.has(String(so.invoice_id))) ||
                          (so.invoice?.id && processedTrxIds.has(String(so.invoice.id))) ||
                          (so.invoice?.receipt_no && processedTrxIds.has(String(so.invoice.receipt_no)))
      return !isProcessed
    })

    unhandledSalesOrders.forEach((so: any) => {
      visitIndex++
      if (so.id) processedSoIds.add(String(so.id))
      if (so.so_number) processedSoIds.add(String(so.so_number))
      if (so.invoice_id) processedTrxIds.add(String(so.invoice_id))
      if (so.invoice?.id) processedTrxIds.add(String(so.invoice.id))
      if (so.invoice?.receipt_no) processedTrxIds.add(String(so.invoice.receipt_no))

      const soDateRaw = so.created_at || so.order_date
      const soDate = soDateRaw ? new Date(soDateRaw) : new Date()
      const formattedDate = soDate.toLocaleString('id-ID', {
        day: '2-digit', month: '2-digit', year: 'numeric',
        hour: '2-digit', minute: '2-digit', second: '2-digit'
      })

      const empName = `${so.employee?.first_name || ''} ${so.employee?.last_name || ''}`.trim() || 'SALES'
      const branchName = so.branch?.name || so.employee?.branch?.name || '-'
      const storeName = so.store?.name || '-'
      const statusKeterangan = 'SALES ORDER DIRECT / TANPA CHECK-IN'

      const allItems: any[] = []
      if (!so.items || so.items.length === 0) {
        if (so.total_amount && Number(so.total_amount) > 0) {
          allItems.push({
            name: so.physical_nota_image_url || so.is_physical_nota ? 'Upload Nota Fisik (SO)' : 'Pesanan Tanpa Rincian Item',
            qty: 1,
            nominal: Number(so.total_amount)
          })
        }
      } else {
        so.items.forEach((item: any) => {
          const prodName = item.product?.name || item.product_name || item.name || 'Item'
          const qty = Number(item.ordered_quantity || item.quantity || 0)
          const price = Number(item.price || 0)
          const nominal = Number(item.subtotal) || (price * qty) || 0
          allItems.push({
            name: prodName,
            qty: qty,
            nominal: nominal
          })
        })
      }

      let totalSales = allItems.reduce((sum, item) => sum + (Number(item.nominal) || 0), 0)
      if (totalSales === 0 && so.total_amount) totalSales = Number(so.total_amount)
      grandTotalSales += totalSales

      const s = getOrCreateSummary(so.employee_id, so.employee)
      if (s) {
        s.trxCount += 1
        s.trxNominal += totalSales
        const soPaid = Number(so.payment_collected_amount || so.paid_amount || 0)
        classifyAndAccumulatePayment(s, so.payment_method, totalSales, soPaid, so.payment_status)
      }

      if (allItems.length === 0) {
        const row = worksheet.addRow([
          rowNum++,
          formattedDate,
          empName.toUpperCase(),
          branchName.toUpperCase(),
          storeName.toUpperCase(),
          statusKeterangan,
          '-',
          0,
          totalSales,
          totalSales
        ])
        applyRowStyle(row, visitIndex, statusKeterangan, true)
      } else {
        const startRowIndex = worksheet.rowCount + 1
        allItems.forEach((item, index) => {
          const row = worksheet.addRow([
            index === 0 ? rowNum++ : '',
            index === 0 ? formattedDate : '',
            index === 0 ? empName.toUpperCase() : '',
            index === 0 ? branchName.toUpperCase() : '',
            index === 0 ? storeName.toUpperCase() : '',
            index === 0 ? statusKeterangan : '',
            item.name.toUpperCase(),
            item.qty,
            item.nominal,
            index === 0 ? totalSales : ''
          ])
          applyRowStyle(row, visitIndex, statusKeterangan, index === 0)
        })
        const endRowIndex = worksheet.rowCount
        if (allItems.length > 1) {
          worksheet.mergeCells(`A${startRowIndex}:A${endRowIndex}`)
          worksheet.mergeCells(`B${startRowIndex}:B${endRowIndex}`)
          worksheet.mergeCells(`C${startRowIndex}:C${endRowIndex}`)
          worksheet.mergeCells(`D${startRowIndex}:D${endRowIndex}`)
          worksheet.mergeCells(`E${startRowIndex}:E${endRowIndex}`)
          worksheet.mergeCells(`F${startRowIndex}:F${endRowIndex}`)
          worksheet.mergeCells(`J${startRowIndex}:J${endRowIndex}`)
        }
      }
    })

    const summaryRow = worksheet.addRow([
      'TOTAL KESELURUHAN OMSET:', '', '', '', '', '', '', '', '', grandTotalSales
    ])
    
    worksheet.mergeCells(`A${summaryRow.number}:I${summaryRow.number}`)
    
    const labelCell = summaryRow.getCell(1)
    labelCell.fill = {
      type: 'pattern',
      pattern: 'solid',
      fgColor: { argb: 'FFD9D9D9' }
    }
    labelCell.font = { name: 'Calibri', size: 12, bold: true }
    labelCell.alignment = { vertical: 'middle', horizontal: 'right' }
    labelCell.border = {
        top: { style: 'thin' }, left: { style: 'thin' },
        bottom: { style: 'thin' }, right: { style: 'thin' }
    }
    
    // Add border to all merged cells so it looks continuous in some excel versions
    for (let i = 2; i <= 9; i++) {
        summaryRow.getCell(i).border = {
            top: { style: 'thin' }, bottom: { style: 'thin' }
        }
    }

    const valueCell = summaryRow.getCell(10)
    valueCell.fill = {
      type: 'pattern',
      pattern: 'solid',
      fgColor: { argb: 'FF92D050' }
    }
    valueCell.font = { name: 'Calibri', size: 12, bold: true }
    valueCell.alignment = { vertical: 'middle', horizontal: 'right' }
    valueCell.border = {
        top: { style: 'thin' }, left: { style: 'thin' },
        bottom: { style: 'thin' }, right: { style: 'thin' }
    }
    valueCell.numFmt = '"Rp"#,##0'
    summaryRow.height = 25

    worksheet.columns = [
      { width: 5 },
      { width: 22 },
      { width: 20 },
      { width: 25 },
      { width: 30 },
      { width: 30 },
      { width: 35 },
      { width: 10 },
      { width: 15 },
      { width: 15 }
    ]

    function applyRowStyle(row: ExcelJS.Row, visitIndex: number, status: string, isFirstRowOfVisit: boolean) {
      const bgColor = visitIndex % 2 === 0 ? 'FFFFFFFF' : 'FFF8FAFC'
      row.eachCell({ includeEmpty: true }, (cell) => {
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: bgColor } }
        cell.border = {
          top: { style: 'thin' },
          left: { style: 'thin' },
          bottom: { style: 'thin' },
          right: { style: 'thin' }
        }
        cell.font = { name: 'Calibri', size: 11 }
        cell.alignment = { vertical: 'middle', horizontal: 'center' }
        if (Number(cell.col) >= 7) {
            cell.alignment = { vertical: 'middle', horizontal: 'right' }
        }
      })
      row.getCell(5).alignment = { vertical: 'middle', horizontal: 'left' } 
      row.getCell(6).alignment = { vertical: 'middle', horizontal: 'left' } 
      row.getCell(7).alignment = { vertical: 'middle', horizontal: 'left' }
      
      if (isFirstRowOfVisit || row.getCell(6).value !== '') {
        if (status.includes('TOKO BARU')) {
          row.getCell(6).font = { name: 'Calibri', size: 11, bold: true, color: { argb: 'FF059669' } }
        } else {
          row.getCell(6).font = { name: 'Calibri', size: 11, color: { argb: 'FF334155' } }
        }
      }
      
      row.getCell(9).numFmt = '"Rp"#,##0'
      row.getCell(10).numFmt = '"Rp"#,##0'
    }

    const wsSummary = workbook.addWorksheet('Summary Kunjungan')

    wsSummary.mergeCells('A1:M1')
    const summaryTitle = wsSummary.getCell('A1')
    summaryTitle.value = 'REKAPITULASI KUNJUNGAN & TRANSAKSI SALES (SUMMARY)'
    summaryTitle.font = { name: 'Calibri', size: 14, bold: true }
    summaryTitle.alignment = { vertical: 'middle', horizontal: 'center' }
    wsSummary.getRow(1).height = 25

    const summaryHeaders = [
      'NO',
      'NAMA SALES',
      'TOTAL KUNJUNGAN',
      'TOKO BARU',
      'TOKO LAMA',
      'TOTAL JUMLAH TRANSAKSI',
      'TOTAL PENJUALAN (RP)',
      'TUNAI (RP)',
      'TRANSFER (RP)',
      'TEMPO (RP)',
      'TEMPO TERBAYAR (RP)',
      'SISA TEMPO (RP)',
      'PROGRESS PEMBAYARAN'
    ]
    const headerSummaryRow = wsSummary.addRow(summaryHeaders)
    headerSummaryRow.eachCell((cell) => {
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E293B' } }
      cell.font = { bold: true, name: 'Calibri', size: 11, color: { argb: 'FFFFFFFF' } }
      cell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
      cell.alignment = { vertical: 'middle', horizontal: 'center', wrapText: true }
    })
    wsSummary.getRow(2).height = 25

    let sumIndex = 1
    let grandTotalTrx = 0
    let grandTotalKunjungan = 0
    let grandTotalTokoBaru = 0
    let grandTotalTokoLama = 0
    let grandTotalTrxCount = 0
    let grandTotalTunai = 0
    let grandTotalTransfer = 0
    let grandTotalTempo = 0
    let grandTotalTempoPaid = 0

    for (const s of summaryMap.values()) {
      const tempoRemaining = Math.max(0, (s.tempoNominal || 0) - (s.tempoPaidNominal || 0))
      let progressText = '-'
      if (s.tempoNominal > 0) {
        const pct = Math.min(100, Math.max(0, (s.tempoPaidNominal / s.tempoNominal) * 100))
        progressText = `${pct.toFixed(1)}%`
      } else if (s.trxNominal > 0) {
        progressText = '100% (Lunas)'
      }

      const row = wsSummary.addRow([
        sumIndex++,
        s.name,
        s.totalKunjungan,
        s.tokoBaru,
        s.tokoLama,
        s.trxCount,
        s.trxNominal,
        s.tunaiNominal || 0,
        s.transferNominal || 0,
        s.tempoNominal || 0,
        s.tempoPaidNominal || 0,
        tempoRemaining,
        progressText
      ])

      grandTotalTrx += s.trxNominal
      grandTotalKunjungan += s.totalKunjungan
      grandTotalTokoBaru += s.tokoBaru
      grandTotalTokoLama += s.tokoLama
      grandTotalTrxCount += s.trxCount
      grandTotalTunai += (s.tunaiNominal || 0)
      grandTotalTransfer += (s.transferNominal || 0)
      grandTotalTempo += (s.tempoNominal || 0)
      grandTotalTempoPaid += (s.tempoPaidNominal || 0)

      const bgColor = sumIndex % 2 === 0 ? 'FFFFFFFF' : 'FFF8FAFC'
      row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: bgColor } }
        cell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
        cell.font = { name: 'Calibri', size: 11 }
        
        const col = Number(colNumber)
        if (col === 1) {
          cell.alignment = { vertical: 'middle', horizontal: 'center' }
        } else if (col === 2) {
          cell.alignment = { vertical: 'middle', horizontal: 'left' }
        } else if (col >= 3 && col <= 6) {
          cell.alignment = { vertical: 'middle', horizontal: 'center' }
        } else if (col >= 7 && col <= 12) {
          cell.alignment = { vertical: 'middle', horizontal: 'right' }
        } else if (col === 13) {
          cell.alignment = { vertical: 'middle', horizontal: 'center' }
        }
      })

      // Format currency for col 7 to 12
      row.getCell(7).numFmt = '"Rp"#,##0'
      row.getCell(8).numFmt = '"Rp"#,##0'
      row.getCell(9).numFmt = '"Rp"#,##0'
      row.getCell(10).numFmt = '"Rp"#,##0'
      row.getCell(11).numFmt = '"Rp"#,##0'
      row.getCell(12).numFmt = '"Rp"#,##0'
    }

    const grandTotalTempoRemaining = Math.max(0, grandTotalTempo - grandTotalTempoPaid)
    let grandProgressText = '-'
    if (grandTotalTempo > 0) {
      const grandPct = Math.min(100, Math.max(0, (grandTotalTempoPaid / grandTotalTempo) * 100))
      grandProgressText = `${grandPct.toFixed(1)}%`
    } else if (grandTotalTrx > 0) {
      grandProgressText = '100% (Lunas)'
    }

    const sumTotalRow = wsSummary.addRow([
      'TOTAL KESELURUHAN:',
      '',
      grandTotalKunjungan,
      grandTotalTokoBaru,
      grandTotalTokoLama,
      grandTotalTrxCount,
      grandTotalTrx,
      grandTotalTunai,
      grandTotalTransfer,
      grandTotalTempo,
      grandTotalTempoPaid,
      grandTotalTempoRemaining,
      grandProgressText
    ])
    wsSummary.mergeCells(`A${sumTotalRow.number}:B${sumTotalRow.number}`)

    const labelSumCell = sumTotalRow.getCell(1)
    labelSumCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD9D9D9' } }
    labelSumCell.font = { name: 'Calibri', size: 12, bold: true }
    labelSumCell.alignment = { vertical: 'middle', horizontal: 'right' }
    labelSumCell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
    sumTotalRow.getCell(2).border = { top: { style: 'thin' }, bottom: { style: 'thin' } }
    
    for (let c = 3; c <= 6; c++) {
      const cell = sumTotalRow.getCell(c)
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD9D9D9' } }
      cell.font = { name: 'Calibri', size: 12, bold: true }
      cell.alignment = { vertical: 'middle', horizontal: 'center' }
      cell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
    }

    for (let c = 7; c <= 12; c++) {
      const cell = sumTotalRow.getCell(c)
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: c === 7 ? 'FF92D050' : 'FFD9D9D9' } }
      cell.font = { name: 'Calibri', size: 12, bold: true }
      cell.alignment = { vertical: 'middle', horizontal: 'right' }
      cell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
      cell.numFmt = '"Rp"#,##0'
    }

    const progressCell = sumTotalRow.getCell(13)
    progressCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD9D9D9' } }
    progressCell.font = { name: 'Calibri', size: 12, bold: true }
    progressCell.alignment = { vertical: 'middle', horizontal: 'center' }
    progressCell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
    
    sumTotalRow.height = 25

    wsSummary.columns = [
      { width: 5 },   // NO
      { width: 32 },  // NAMA SALES
      { width: 18 },  // TOTAL KUNJUNGAN
      { width: 14 },  // TOKO BARU
      { width: 14 },  // TOKO LAMA
      { width: 22 },  // TOTAL JUMLAH TRANSAKSI
      { width: 24 },  // TOTAL PENJUALAN (RP)
      { width: 20 },  // TUNAI (RP)
      { width: 20 },  // TRANSFER (RP)
      { width: 20 },  // TEMPO (RP)
      { width: 22 },  // TEMPO TERBAYAR (RP)
      { width: 20 },  // SISA TEMPO (RP)
      { width: 22 }   // PROGRESS PEMBAYARAN
    ]

    // ─── Sheet 3: Laporan Setoran Kasir ───
    const wsCashier = workbook.addWorksheet('Laporan Setoran Kasir')

    // Filter and unify payments (Approved payments only)
    const isApprovedPayment = (p: any) => {
      const st = (p.payment_status || '').toUpperCase()
      return ['SUCCESS', 'PAID', 'APPROVED', 'CASHIER_APPROVED'].includes(st)
    }

    const seenPayIds = new Set<string>()
    const cashierPayments: any[] = []
    const rawAllPayments = [...exportCashPayments, ...exportTransferPayments]

    for (const p of rawAllPayments) {
      if (!p) continue
      const pid = String(p.id)
      if (pid && seenPayIds.has(pid)) continue
      if (pid) seenPayIds.add(pid)

      // Only include transactions that have been approved by cashier
      if (!isApprovedPayment(p)) {
        continue
      }

      // Cashier approval sheet represents all approved cash/transfer deposits for the branch/period,
      // so it includes all salesmen approved by the cashier (not filtered by single salesman selection)
      cashierPayments.push(p)
    }

    // Sort payments by approved_at / payment_date descending
    cashierPayments.sort((a, b) => {
      const timeA = new Date(a.approved_at || a.payment_date || a.created_at || 0).getTime()
      const timeB = new Date(b.approved_at || b.payment_date || b.created_at || 0).getTime()
      return timeB - timeA
    })

    // Accumulate KPI for approved cashier deposits
    let totalCashSetoran = 0
    let totalTransferSetoran = 0

    cashierPayments.forEach(p => {
      const amt = Number(p.amount) || 0
      const isCash = (p.payment_method || '').toUpperCase() === 'CASH' || (p.payment_method || '').toUpperCase() === 'TUNAI'
      if (isCash) {
        totalCashSetoran += amt
      } else {
        totalTransferSetoran += amt
      }
    })

    const grandTotalCashier = totalCashSetoran + totalTransferSetoran

    // 1. Header Title
    wsCashier.mergeCells('A1:L1')
    const cTitle = wsCashier.getCell('A1')
    cTitle.value = 'REKAPITULASI LAPORAN SETORAN & PERSETUJUAN KASIR'
    cTitle.font = { name: 'Calibri', size: 14, bold: true, color: { argb: 'FFFFFFFF' } }
    cTitle.alignment = { vertical: 'middle', horizontal: 'center' }
    cTitle.fill = {
      type: 'pattern',
      pattern: 'solid',
      fgColor: { argb: 'FF0F766E' } // Deep Teal/Emerald
    }
    wsCashier.getRow(1).height = 28

    // 2. Subtitle
    wsCashier.mergeCells('A2:L2')
    const cSub = wsCashier.getCell('A2')
    const periodText = filterMode.value === 'date'
      ? (selectedDate.value || todayLocal || 'Hari Ini')
      : (selectedMonth.value || 'Semua Bulan')
    const branchText = (masterStore.branches.find(b => b.id === masterStore.selectedBranchId)?.name || 'Semua Cabang').toUpperCase()
    cSub.value = `Periode: ${periodText}   |   Cabang: ${branchText}   |   Status: Disetujui Kasir   |   Waktu Unduh: ${new Date().toLocaleString('id-ID')} WIB`
    cSub.font = { name: 'Calibri', size: 10, italic: true, color: { argb: 'FF475569' } }
    cSub.alignment = { vertical: 'middle', horizontal: 'center' }
    wsCashier.getRow(2).height = 20

    // 3. Executive KPI Box (Baris 4 - 7)
    wsCashier.mergeCells('B4:D4')
    wsCashier.getCell('B4').value = 'RINGKASAN SETORAN KASIR (TERVALIDASI)'
    wsCashier.getCell('B4').font = { name: 'Calibri', size: 11, bold: true, color: { argb: 'FF1E293B' } }
    wsCashier.mergeCells('E4:G4')
    wsCashier.getCell('E4').value = `Total: ${cashierPayments.length} Transaksi Disetujui`
    wsCashier.getCell('E4').font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF64748B' } }
    wsCashier.getCell('E4').alignment = { horizontal: 'right', vertical: 'middle' }

    // KPI Row 5: Tunai vs Status
    const kpi5 = wsCashier.getRow(5)
    kpi5.getCell(2).value = 'Total Setoran Tunai (Cash):'
    kpi5.getCell(3).value = totalCashSetoran
    kpi5.getCell(3).numFmt = '"Rp"#,##0'
    kpi5.getCell(2).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF047857' } }
    kpi5.getCell(3).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF065F46' } }

    kpi5.getCell(5).value = 'Status Validasi Kasir:'
    kpi5.getCell(6).value = 'LUNAS / DISETUJUI'
    kpi5.getCell(5).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF047857' } }
    kpi5.getCell(6).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF059669' } }

    // KPI Row 6: Transfer vs Approval
    const kpi6 = wsCashier.getRow(6)
    kpi6.getCell(2).value = 'Total Setoran Transfer Bank:'
    kpi6.getCell(3).value = totalTransferSetoran
    kpi6.getCell(3).numFmt = '"Rp"#,##0'
    kpi6.getCell(2).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF4338CA' } }
    kpi6.getCell(3).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF3730A3' } }

    kpi6.getCell(5).value = 'Verifikasi Penerimaan:'
    kpi6.getCell(6).value = 'SUDAH DI-ACC KASIR'
    kpi6.getCell(5).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF4338CA' } }
    kpi6.getCell(6).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF3730A3' } }

    // KPI Row 7: Grand Total vs Periode
    const kpi7 = wsCashier.getRow(7)
    kpi7.getCell(2).value = 'GRAND TOTAL SETORAN KAS:'
    kpi7.getCell(3).value = grandTotalCashier
    kpi7.getCell(3).numFmt = '"Rp"#,##0'
    kpi7.getCell(2).font = { name: 'Calibri', size: 11, bold: true, color: { argb: 'FF0F172A' } }
    kpi7.getCell(3).font = { name: 'Calibri', size: 11, bold: true, color: { argb: 'FF047857' } }
    kpi7.getCell(2).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD1FAE5' } }
    kpi7.getCell(3).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD1FAE5' } }

    kpi7.getCell(5).value = 'Cabang / Entitas:'
    kpi7.getCell(6).value = branchText
    kpi7.getCell(5).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF475569' } }
    kpi7.getCell(6).font = { name: 'Calibri', size: 10, bold: true, color: { argb: 'FF1E293B' } }

    // KPI Borders
    for (let r = 5; r <= 7; r++) {
      const row = wsCashier.getRow(r)
      ;[2, 3, 5, 6].forEach(c => {
        row.getCell(c).border = {
          top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          right: { style: 'thin', color: { argb: 'FFE2E8F0' } }
        }
      })
    }

    // 4. Table Headers
    const cashierHeaders = [
      'NO',
      'TANGGAL & WAKTU SETOR',
      'NO. NOTA / FAKTUR',
      'NAMA SALESMAN / PENYETOR',
      'CABANG',
      'NAMA TOKO / PELANGGAN',
      'CHANNEL / TIPE',
      'METODE BAYAR',
      'NOMINAL SETORAN (RP)',
      'STATUS KASIR',
      'KASIR VERIFIKATOR',
      'CATATAN'
    ]

    const cHeaderRow = wsCashier.getRow(9)
    cHeaderRow.values = cashierHeaders
    cHeaderRow.eachCell(cell => {
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E293B' } }
      cell.font = { name: 'Calibri', size: 10.5, bold: true, color: { argb: 'FFFFFFFF' } }
      cell.alignment = { vertical: 'middle', horizontal: 'center', wrapText: true }
      cell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
    })
    cHeaderRow.height = 25

    // 5. Data Rows
    let cRowIdx = 1
    if (cashierPayments.length === 0) {
      const emptyRow = wsCashier.addRow([
        1, '-', '-', '-', '-', 'Tidak ada data setoran kasir yang disetujui pada periode ini', '-', '-', 0, '-', '-', '-'
      ])
      wsCashier.mergeCells(`A${emptyRow.number}:L${emptyRow.number}`)
      const emptyCell = emptyRow.getCell(1)
      emptyCell.alignment = { vertical: 'middle', horizontal: 'center' }
      emptyCell.font = { name: 'Calibri', size: 11, italic: true, color: { argb: 'FF94A3B8' } }
    } else {
      for (const p of cashierPayments) {
        const payDate = p.approved_at || p.payment_date || p.created_at ? new Date(p.approved_at || p.payment_date || p.created_at) : new Date()
        const formattedPayDate = payDate.toLocaleString('id-ID', {
          day: '2-digit', month: '2-digit', year: 'numeric',
          hour: '2-digit', minute: '2-digit', second: '2-digit'
        })

        const refNo = p.sales_transaction?.receipt_no || p.sales_order?.so_number || p.midtrans_transaction_id || '-'

        const salesmanName = (
          `${p.sales_transaction?.employee?.first_name || ''} ${p.sales_transaction?.employee?.last_name || ''}`.trim() ||
          `${p.sales_order?.employee?.first_name || ''} ${p.sales_order?.employee?.last_name || ''}`.trim() ||
          `${p.employee?.first_name || ''} ${p.employee?.last_name || ''}`.trim() ||
          'Sales'
        ).toUpperCase()

        const branchName = (
          p.sales_transaction?.branch?.name ||
          p.sales_order?.branch?.name ||
          p.employee?.operating_branch?.name ||
          p.employee?.branch?.name ||
          '-'
        ).toUpperCase()

        let customerName = '-'
        if (p.sales_transaction?.buyer_type === 'EMPLOYEE' && p.sales_transaction?.buyer_employee) {
          customerName = `${p.sales_transaction.buyer_employee.first_name} ${p.sales_transaction.buyer_employee.last_name || ''} (Karyawan)`
        } else if (p.sales_transaction?.buyer_name) {
          customerName = p.sales_transaction.buyer_name
        } else if (p.sales_order?.store?.name) {
          customerName = p.sales_order.store.name
        } else if (p.sales_transaction?.store?.name) {
          customerName = p.sales_transaction.store.name
        }

        let channelName = 'Sales (Motoris)'
        if (p.sales_transaction?.sales_channel === 'DIRECT_FACTORY') {
          channelName = 'Penjualan Pabrik'
        } else if (p.sales_transaction?.sales_channel === 'DIRECT_BRANCH') {
          channelName = 'Penjualan Cabang'
        } else if (p.sales_transaction?.sales_channel === 'EMPLOYEE_SALE') {
          channelName = 'Penjualan Karyawan'
        } else if (p.sales_order) {
          channelName = 'Driver (Delivery)'
        }

        let methodDisplay = (p.payment_method || 'CASH').toUpperCase()
        if (methodDisplay === 'CASH' || methodDisplay === 'TUNAI') {
          methodDisplay = 'TUNAI'
        } else if (['TRANSFER', 'VA', 'QRIS', 'MIDTRANS_QRIS'].includes(methodDisplay)) {
          methodDisplay = 'TRANSFER'
        }
        if (p.notes && (p.notes.includes('Uang Muka') || p.notes.includes('(DP)'))) {
          methodDisplay += ' (DP)'
        } else if (p.notes && (p.notes.includes('Cicilan') || p.notes.includes('Penagihan'))) {
          methodDisplay += ' (CICILAN)'
        }

        const amt = Number(p.amount) || 0
        const statusDisplay = 'DISETUJUI KASIR'
        const statusColor = 'FF059669' // emerald

        let verifier = '-'
        if (p.verified_by) {
          verifier = `${p.verified_by.first_name || ''} ${p.verified_by.last_name || ''}`.trim()
        }
        if (p.approved_at) {
          const appDate = new Date(p.approved_at).toLocaleString('id-ID', {
            day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit'
          })
          verifier = verifier !== '-' ? `${verifier} (${appDate})` : `ACC (${appDate})`
        }

        const row = wsCashier.addRow([
          cRowIdx++,
          formattedPayDate,
          refNo,
          salesmanName,
          branchName,
          customerName.toUpperCase(),
          channelName,
          methodDisplay,
          amt,
          statusDisplay,
          verifier,
          p.notes || '-'
        ])

        const bgCol = cRowIdx % 2 === 0 ? 'FFFFFFFF' : 'FFF8FAFC'
        row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: bgCol } }
          cell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
          cell.font = { name: 'Calibri', size: 10.5 }
          
          const c = Number(colNumber)
          if (c === 1 || c === 2 || c === 5 || c === 7 || c === 8 || c === 10) {
            cell.alignment = { vertical: 'middle', horizontal: 'center' }
          } else if (c === 9) {
            cell.alignment = { vertical: 'middle', horizontal: 'right' }
          } else {
            cell.alignment = { vertical: 'middle', horizontal: 'left' }
          }
        })

        // Format currency for col 9
        row.getCell(9).numFmt = '"Rp"#,##0'
        // Format status font color & bold for col 10
        row.getCell(10).font = { name: 'Calibri', size: 10.5, bold: true, color: { argb: statusColor } }
      }
    }

    // 6. Summary Total Row
    const cTotalRow = wsCashier.addRow([
      'TOTAL KESELURUHAN SETORAN KASIR:', '', '', '', '', '', '', '', grandTotalCashier, '', '', ''
    ])
    wsCashier.mergeCells(`A${cTotalRow.number}:H${cTotalRow.number}`)

    const cTotalLabel = cTotalRow.getCell(1)
    cTotalLabel.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD9D9D9' } }
    cTotalLabel.font = { name: 'Calibri', size: 12, bold: true }
    cTotalLabel.alignment = { vertical: 'middle', horizontal: 'right' }
    cTotalLabel.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
    for (let c = 2; c <= 8; c++) {
      cTotalRow.getCell(c).border = { top: { style: 'thin' }, bottom: { style: 'thin' } }
    }

    const cTotalVal = cTotalRow.getCell(9)
    cTotalVal.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF92D050' } }
    cTotalVal.font = { name: 'Calibri', size: 12, bold: true }
    cTotalVal.alignment = { vertical: 'middle', horizontal: 'right' }
    cTotalVal.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
    cTotalVal.numFmt = '"Rp"#,##0'

    for (let c = 10; c <= 12; c++) {
      const cell = cTotalRow.getCell(c)
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD9D9D9' } }
      cell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
    }
    cTotalRow.height = 25

    // 7. Column Widths
    wsCashier.columns = [
      { width: 6 },   // NO
      { width: 22 },  // TANGGAL & WAKTU SETOR
      { width: 24 },  // NO. NOTA / FAKTUR
      { width: 28 },  // NAMA SALESMAN / PENYETOR
      { width: 18 },  // CABANG
      { width: 30 },  // NAMA TOKO / PELANGGAN
      { width: 20 },  // CHANNEL / TIPE
      { width: 18 },  // METODE BAYAR
      { width: 22 },  // NOMINAL SETORAN (RP)
      { width: 22 },  // STATUS KASIR
      { width: 26 },  // KASIR VERIFIKATOR
      { width: 32 }   // CATATAN
    ]

    // ─── Sheet 4: Rekap Odometer & BBM ───
    if (exportJourneys.length > 0) {
      const wsJourney = workbook.addWorksheet('Perjalanan & BBM')
      wsJourney.mergeCells('A1:I1')
      const jTitle = wsJourney.getCell('A1')
      jTitle.value = 'REKAPITULASI ODOMETER (KM) & KLAIM BBM SALES'
      jTitle.font = { name: 'Calibri', size: 14, bold: true }
      jTitle.alignment = { vertical: 'middle', horizontal: 'center' }
      wsJourney.getRow(1).height = 25

      const jHeaders = ['NO', 'SALESMAN', 'CABANG', 'TANGGAL', 'KM BERANGKAT', 'KM PULANG', 'TOTAL JARAK (KM)', 'TOKO DIKUNJUNGI', 'KLAIM BBM (RP)']
      const jHeaderRow = wsJourney.addRow(jHeaders)
      jHeaderRow.eachCell((cell) => {
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E293B' } }
        cell.font = { name: 'Calibri', size: 11, bold: true, color: { argb: 'FFFFFFFF' } }
        cell.alignment = { vertical: 'middle', horizontal: 'center' }
      })
      wsJourney.getRow(2).height = 20

      let jIdx = 1
      let totalJarakAll = 0
      let totalBbmAll = 0

      for (const j of exportJourneys) {
        const row = wsJourney.addRow([
          jIdx++,
          `${j.employee?.first_name || ''} ${j.employee?.last_name || ''}`.trim() || 'Sales',
          j.employee?.branch?.name || 'Pusat',
          j.date,
          j.start_km || 0,
          j.end_km || 0,
          j.total_km || 0,
          j.total_visits || 0,
          j.total_bbm_amount || 0,
        ])
        totalJarakAll += (j.total_km || 0)
        totalBbmAll += (j.total_bbm_amount || 0)

        const bgCol = jIdx % 2 === 0 ? 'FFFFFFFF' : 'FFF8FAFC'
        row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: bgCol } }
          cell.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }
          cell.font = { name: 'Calibri', size: 11 }
          cell.alignment = { vertical: 'middle', horizontal: Number(colNumber) === 2 || Number(colNumber) === 3 ? 'left' : 'center' }
        })
        row.getCell(9).numFmt = '"Rp"#,##0'
      }

      const jSummaryRow = wsJourney.addRow([
        'TOTAL KESELURUHAN:', '', '', '', '', '', totalJarakAll, '', totalBbmAll
      ])
      wsJourney.mergeCells(`A${jSummaryRow.number}:F${jSummaryRow.number}`)
      const jSumLabel = jSummaryRow.getCell(1)
      jSumLabel.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD9D9D9' } }
      jSumLabel.font = { name: 'Calibri', size: 12, bold: true }
      jSumLabel.alignment = { vertical: 'middle', horizontal: 'right' }
      jSumLabel.border = { top: { style: 'thin' }, left: { style: 'thin' }, bottom: { style: 'thin' }, right: { style: 'thin' } }

      const jSumKm = jSummaryRow.getCell(7)
      jSumKm.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF92D050' } }
      jSumKm.font = { name: 'Calibri', size: 12, bold: true }
      jSumKm.alignment = { vertical: 'middle', horizontal: 'center' }

      const jSumBbm = jSummaryRow.getCell(9)
      jSumBbm.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF92D050' } }
      jSumBbm.font = { name: 'Calibri', size: 12, bold: true }
      jSumBbm.alignment = { vertical: 'middle', horizontal: 'right' }
      jSumBbm.numFmt = '"Rp"#,##0'

      wsJourney.columns = [
        { width: 5 }, { width: 30 }, { width: 20 }, { width: 15 }, { width: 18 }, { width: 18 }, { width: 20 }, { width: 18 }, { width: 22 }
      ]
    }

    const filterName = selectedDate.value || selectedMonth.value || 'Semua'
    const prefix = isMT.value ? 'Rekap_Kunjungan_MT_' : 'Rekap_Kunjungan_Dan_Transaksi_'
    const buffer = await workbook.xlsx.writeBuffer()
    saveAs(new Blob([buffer]), `${prefix}${filterName}.xlsx`)
    toast.success('Data berhasil di-export ke Excel')
  } catch (error) {
    console.error('Export failed:', error)
    toast.error(error instanceof Error ? error.message : 'Gagal meng-export data')
  } finally {
    isExporting.value = false
  }
}
</script>

<template>
  <div class="h-full flex flex-col gap-4 sm:gap-6 p-3 sm:p-6 overflow-y-auto md:overflow-hidden bg-[#F8FAFC]">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl sm:text-3xl font-black text-slate-900 tracking-tight flex items-center flex-wrap gap-2.5 sm:gap-3">
          <MapPin class="w-6 h-6 sm:w-8 sm:h-8 text-primary shrink-0" />
          <span>{{ isMT ? 'Manajemen Kunjungan MT' : 'Manajemen Kunjungan' }}</span>
          <span v-if="masterStore.selectedBranchName" class="text-primary text-xl sm:text-2xl font-bold">
            — {{ masterStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-xs sm:text-sm text-slate-500 font-medium mt-1">
          {{ isMT ? 'Pantau aktivitas dan hasil kunjungan tim sales Modern Trade di lapangan.' : 'Pantau aktivitas dan hasil kunjungan tim sales di lapangan.' }}
        </p>
      </div>

      <div class="flex flex-wrap items-stretch sm:items-center gap-2 sm:gap-3">
        <Button @click="exportToExcel" :disabled="isExporting" variant="outline" class="flex-1 sm:flex-initial rounded-2xl px-3.5 sm:px-6 py-2 sm:py-2.5 text-xs sm:text-sm font-black gap-2 bg-white text-slate-700 border-slate-200 hover:bg-slate-50 shadow-sm">
          <Download class="w-4 h-4 shrink-0" :class="{ 'animate-bounce': isExporting }" /> 
          <span>{{ isExporting ? 'Mengekspor...' : 'Export Excel' }}</span>
        </Button>
        <Button @click="openCreate" class="flex-1 sm:flex-initial rounded-2xl px-3.5 sm:px-6 py-2 sm:py-2.5 text-xs sm:text-sm font-black gap-2 shadow-sm">
          <Plus class="w-4 h-4 shrink-0" /> <span>Tambah Kunjungan</span>
        </Button>
        <div
          class="bg-white p-1 sm:p-1.5 rounded-2xl border border-slate-200 shadow-sm flex items-center gap-1.5 sm:gap-2 w-full sm:w-auto justify-between sm:justify-start"
        >
          <select
            v-model="filterMode"
            @change="resetAndFetch"
            class="bg-transparent border-none focus:ring-0 text-xs font-bold text-slate-500 px-1.5 sm:px-2 outline-none cursor-pointer"
          >
            <option value="date">Harian</option>
            <option value="month">Bulanan</option>
          </select>
          <div class="w-px h-4 bg-slate-200"></div>

          <input
            v-if="filterMode === 'date'"
            type="date"
            v-model="selectedDate"
            @change="resetAndFetch"
            class="bg-transparent border-none focus:ring-0 text-xs sm:text-sm font-bold text-slate-700 px-2 sm:px-3 min-w-[120px] sm:min-w-[140px] flex-1 sm:flex-initial"
            placeholder="Semua Tanggal"
          />
          <input
            v-else
            type="month"
            v-model="selectedMonth"
            @change="resetAndFetch"
            class="bg-transparent border-none focus:ring-0 text-xs sm:text-sm font-bold text-slate-700 px-2 sm:px-3 min-w-[120px] sm:min-w-[140px] flex-1 sm:flex-initial"
          />

          <button
            v-if="selectedDate || selectedMonth"
            @click="handleResetDateFilter"
            class="p-1 hover:bg-slate-100 rounded-md transition-all text-slate-400"
            title="Reset Filter"
          >
            <XCircle class="w-4 h-4" />
          </button>
          <div class="w-px h-6 bg-slate-200"></div>
          <button
            @click="handleRefreshAll"
            class="p-1.5 sm:p-2 hover:bg-slate-50 rounded-xl transition-all text-slate-400"
            title="Refresh Data"
          >
            <RefreshCw
              class="w-4 h-4"
              :class="{ 'animate-spin': isLoading || isLoadingTransactions }"
            />
          </button>
        </div>
      </div>
    </div>

    <!-- Filters & View Toggle Row -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 sm:gap-4 shrink-0">
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 sm:gap-4 flex-1">
        <!-- Search -->
        <div class="relative flex-1 min-w-[180px] w-full sm:max-w-md">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="isMT ? 'Cari toko atau sales MT...' : 'Cari toko atau salesman...'"
            class="w-full pl-11 pr-10 py-2.5 sm:py-3 bg-white border border-slate-200 rounded-2xl text-xs sm:text-sm font-medium focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all shadow-sm"
          />
          <button
            v-if="searchQuery"
            type="button"
            @click="searchQuery = ''"
            class="absolute right-3.5 top-1/2 -translate-y-1/2 p-1 hover:bg-slate-100 rounded-lg text-slate-400 hover:text-slate-600 transition-colors"
            title="Hapus pencarian"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>

        <!-- Employee Filter -->
        <select
          v-model="selectedEmployee"
          @change="onEmployeeFilterChange"
          class="w-full sm:w-auto px-4 sm:px-6 py-2.5 sm:py-3 bg-white border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 transition-all shadow-sm outline-none max-w-full sm:max-w-[220px]"
        >
          <option value="ALL">{{ isMT ? 'Semua Sales MT' : 'Semua Salesman' }}</option>
          <option v-for="emp in salesmen" :key="emp.id" :value="emp.id">
            {{ emp.first_name }} {{ emp.last_name }} - {{ emp.job_position?.title || 'No Position' }}
          </option>
        </select>
      </div>

      <!-- View Toggle -->
      <div class="bg-white border border-slate-200 p-1 sm:p-1.5 rounded-2xl flex shadow-sm overflow-x-auto no-scrollbar max-w-full shrink-0">
        <button
          @click="activeTab = 'table'"
          :class="[
            'px-3 sm:px-5 py-2 sm:py-2.5 rounded-xl text-xs font-bold transition-all duration-300 flex items-center gap-1.5 sm:gap-2 whitespace-nowrap shrink-0',
            activeTab === 'table'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
          ]"
        >
          <ListIcon class="w-3.5 h-3.5 sm:w-4 sm:h-4" /> Tabel Data
        </button>
        <button
          @click="activeTab = 'map'"
          :class="[
            'px-3 sm:px-5 py-2 sm:py-2.5 rounded-xl text-xs font-bold transition-all duration-300 flex items-center gap-1.5 sm:gap-2 whitespace-nowrap shrink-0',
            activeTab === 'map'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
          ]"
        >
          <MapIcon class="w-3.5 h-3.5 sm:w-4 sm:h-4" /> Peta Kunjungan
        </button>
        <button
          @click="activeTab = 'schedule'"
          :class="[
            'px-3 sm:px-5 py-2 sm:py-2.5 rounded-xl text-xs font-bold transition-all duration-300 flex items-center gap-1.5 sm:gap-2 whitespace-nowrap shrink-0',
            activeTab === 'schedule'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
          ]"
        >
          <Calendar class="w-3.5 h-3.5 sm:w-4 sm:h-4" /> Rencana Kunjungan
        </button>
        <button
          @click="activeTab = 'omzet'"
          :class="[
            'px-3 sm:px-5 py-2 sm:py-2.5 rounded-xl text-xs font-bold transition-all duration-300 flex items-center gap-1.5 sm:gap-2 whitespace-nowrap shrink-0',
            activeTab === 'omzet'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
          ]"
        >
          <WalletIcon class="w-3.5 h-3.5 sm:w-4 sm:h-4" /> Status Setoran
        </button>
        <button
          @click="activeTab = 'rekap'"
          :class="[
            'px-3 sm:px-5 py-2 sm:py-2.5 rounded-xl text-xs font-bold transition-all duration-300 flex items-center gap-1.5 sm:gap-2 whitespace-nowrap shrink-0',
            activeTab === 'rekap'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
          ]"
        >
          <Users class="w-3.5 h-3.5 sm:w-4 sm:h-4" /> Rekap Sales
        </button>
        <button
          v-if="isMT"
          @click="activeTab = 'merchandising'"
          :class="[
            'px-3 sm:px-5 py-2 sm:py-2.5 rounded-xl text-xs font-bold transition-all duration-300 flex items-center gap-1.5 sm:gap-2 whitespace-nowrap shrink-0',
            activeTab === 'merchandising'
              ? 'bg-rose-700 text-white shadow-lg shadow-rose-200'
              : 'text-rose-700 hover:text-rose-900 hover:bg-rose-50 font-black',
          ]"
        >
          <Layers class="w-3.5 h-3.5 sm:w-4 sm:h-4" /> Audit Merchandising
        </button>
        <button
          @click="activeTab = 'journey'"
          :class="[
            'px-3 sm:px-5 py-2 sm:py-2.5 rounded-xl text-xs font-bold transition-all duration-300 flex items-center gap-1.5 sm:gap-2 whitespace-nowrap shrink-0',
            activeTab === 'journey'
              ? 'bg-primary text-white shadow-lg shadow-primary/20'
              : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
          ]"
        >
          <Gauge class="w-3.5 h-3.5 sm:w-4 sm:h-4" /> Perjalanan & BBM
        </button>
      </div>
    </div>

    <!-- Summary Cards (Visible only on Table View) -->
    <div v-if="activeTab === 'table'" class="grid grid-cols-2 lg:grid-cols-4 gap-2.5 sm:gap-4 shrink-0">
      <!-- Total Kunjungan -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-5 border border-slate-200 shadow-sm flex items-center justify-between gap-2.5 sm:gap-4">
        <div class="min-w-0 flex-1">
          <p class="text-[10px] sm:text-xs font-bold text-slate-400">Total Kunjungan</p>
          <p class="text-base sm:text-xl font-black text-slate-800 leading-tight mt-0.5">
            {{ visitStats.totalKunjungan }} <span class="text-[10px] sm:text-xs font-medium text-slate-500">toko</span>
          </p>
        </div>
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl bg-blue-50 flex items-center justify-center shrink-0">
          <MapPin class="w-5 h-5 text-blue-500" />
        </div>
      </div>

      <!-- Jenis Toko -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-5 border border-slate-200 shadow-sm flex items-center justify-between gap-2.5 sm:gap-4">
        <div class="min-w-0 flex-1">
          <p class="text-[10px] sm:text-xs font-bold text-slate-400">Toko Baru / Lama</p>
          <p class="text-base sm:text-xl font-black text-slate-800 leading-tight mt-0.5">
            {{ visitStats.tokoBaru }} <span class="text-xs sm:text-sm text-slate-300 font-normal">/</span> {{ visitStats.tokoLama }}
          </p>
        </div>
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl bg-purple-50 flex items-center justify-center shrink-0">
          <Store class="w-5 h-5 text-purple-500" />
        </div>
      </div>

      <!-- Total Order -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-5 border border-slate-200 shadow-sm flex items-center justify-between gap-2.5 sm:gap-4">
        <div class="min-w-0 flex-1">
          <p class="text-[10px] sm:text-xs font-bold text-slate-400">Total Order</p>
          <p class="text-base sm:text-xl font-black text-slate-800 leading-tight mt-0.5">
            {{ visitStats.totalTransaksi }} <span class="text-[10px] sm:text-xs font-medium text-slate-500">trx</span>
          </p>
        </div>
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl bg-amber-50 flex items-center justify-center shrink-0">
          <ShoppingBag class="w-5 h-5 text-amber-500" />
        </div>
      </div>

      <!-- Nominal Transaksi (Grand Total) -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-5 border border-slate-200 shadow-sm flex items-center justify-between gap-2.5 sm:gap-4">
        <div class="min-w-0 flex-1">
          <p class="text-[10px] sm:text-xs font-bold text-slate-400">Total Nominal (Semua)</p>
          <p 
            :class="[getNominalFontSizeClass(visitStats.totalNominal), 'font-black text-slate-800 tracking-tight leading-tight mt-0.5 whitespace-nowrap']"
            :title="`Rp ${formatRupiah(visitStats.totalNominal)}`"
          >
            <span class="text-[10px] sm:text-xs font-bold text-emerald-600">Rp</span> {{ formatRupiah(visitStats.totalNominal) }}
          </p>
        </div>
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl bg-emerald-50 flex items-center justify-center shrink-0">
          <TrendingUp class="w-5 h-5 text-emerald-500" />
        </div>
      </div>
    </div>

    <!-- Breakdown Omset per Divisi Sales -->
    <div v-if="activeTab === 'table'" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5 sm:gap-4 shrink-0">
      <!-- Omset Sales TO -->
      <div class="bg-gradient-to-br from-amber-500/5 via-amber-50/40 to-white rounded-2xl p-3.5 sm:p-4 border border-amber-200/70 shadow-sm flex items-center justify-between gap-3 hover:border-amber-300 transition-all">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-1.5 mb-1">
            <span class="w-2 h-2 rounded-full bg-amber-500"></span>
            <p class="text-[10px] sm:text-xs font-black text-amber-800 uppercase tracking-wider">Omset TO</p>
          </div>
          <p 
            :class="[getNominalFontSizeClass(visitStats.nominalTO), 'font-black text-slate-900 tracking-tight leading-tight whitespace-nowrap']"
            :title="`Rp ${formatRupiah(visitStats.nominalTO)}`"
          >
            <span class="text-[10px] sm:text-xs font-bold text-amber-600">Rp</span> {{ formatRupiah(visitStats.nominalTO) }}
          </p>
          <p class="text-[10px] sm:text-[11px] font-semibold text-slate-500 mt-0.5">
            {{ visitStats.transaksiTO }} transaksi SO / Pre-Order
          </p>
        </div>
        <div class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-amber-100/80 flex items-center justify-center shrink-0 border border-amber-200/60">
          <ClipboardCheck class="w-4 h-4 sm:w-5 sm:h-5 text-amber-600" />
        </div>
      </div>

      <!-- Omset Sales Motoris -->
      <div class="bg-gradient-to-br from-blue-500/5 via-blue-50/40 to-white rounded-2xl p-3.5 sm:p-4 border border-blue-200/70 shadow-sm flex items-center justify-between gap-3 hover:border-blue-300 transition-all">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-1.5 mb-1">
            <span class="w-2 h-2 rounded-full bg-blue-500"></span>
            <p class="text-[10px] sm:text-xs font-black text-blue-800 uppercase tracking-wider">Omset Sales Motoris</p>
          </div>
          <p 
            :class="[getNominalFontSizeClass(visitStats.nominalMotoris), 'font-black text-slate-900 tracking-tight leading-tight whitespace-nowrap']"
            :title="`Rp ${formatRupiah(visitStats.nominalMotoris)}`"
          >
            <span class="text-[10px] sm:text-xs font-bold text-blue-600">Rp</span> {{ formatRupiah(visitStats.nominalMotoris) }}
          </p>
          <p class="text-[10px] sm:text-[11px] font-semibold text-slate-500 mt-0.5">
            {{ visitStats.transaksiMotoris }} transaksi langsung / nota
          </p>
        </div>
        <div class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-blue-100/80 flex items-center justify-center shrink-0 border border-blue-200/60">
          <Bike class="w-4 h-4 sm:w-5 sm:h-5 text-blue-600" />
        </div>
      </div>

      <!-- Omset Sales Booster -->
      <div class="bg-gradient-to-br from-purple-500/5 via-purple-50/40 to-white rounded-2xl p-3.5 sm:p-4 border border-purple-200/70 shadow-sm flex items-center justify-between gap-3 hover:border-purple-300 transition-all">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-1.5 mb-1">
            <span class="w-2 h-2 rounded-full bg-purple-500"></span>
            <p class="text-[10px] sm:text-xs font-black text-purple-800 uppercase tracking-wider">Omset Sales Booster</p>
          </div>
          <p 
            :class="[getNominalFontSizeClass(visitStats.nominalBooster), 'font-black text-slate-900 tracking-tight leading-tight whitespace-nowrap']"
            :title="`Rp ${formatRupiah(visitStats.nominalBooster)}`"
          >
            <span class="text-[10px] sm:text-xs font-bold text-purple-600">Rp</span> {{ formatRupiah(visitStats.nominalBooster) }}
          </p>
          <p class="text-[10px] sm:text-[11px] font-semibold text-slate-500 mt-0.5">
            {{ visitStats.transaksiBooster }} transaksi booster
          </p>
        </div>
        <div class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-purple-100/80 flex items-center justify-center shrink-0 border border-purple-200/60">
          <Zap class="w-4 h-4 sm:w-5 sm:h-5 text-purple-600" />
        </div>
      </div>

      <!-- Orders from MT and transactions outside the three sales categories -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-4 border border-slate-200 shadow-sm flex items-center justify-between gap-3">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-1.5 mb-1">
            <span class="w-2 h-2 rounded-full bg-slate-500"></span>
            <p class="text-[10px] sm:text-xs font-black text-slate-700 uppercase tracking-wider">Omset MT / Lainnya</p>
          </div>
          <p :class="[getNominalFontSizeClass(visitStats.nominalOther), 'font-black text-slate-900 tracking-tight leading-tight whitespace-nowrap']"
            :title="`Rp ${formatRupiah(visitStats.nominalOther)}`">
            <span class="text-[10px] sm:text-xs font-bold text-slate-500">Rp</span> {{ formatRupiah(visitStats.nominalOther) }}
          </p>
          <p class="text-[10px] sm:text-[11px] font-semibold text-slate-500 mt-0.5">
            {{ visitStats.transaksiOther }} transaksi MT / lainnya
          </p>
          <button type="button" class="text-[10px] sm:text-[11px] font-bold text-blue-700 hover:underline mt-1"
            :aria-expanded="showOtherDetails" @click="showOtherDetails = !showOtherDetails">
            {{ showOtherDetails ? 'Tutup rincian' : 'Lihat rincian transaksi' }}
          </button>
        </div>
        <div class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-slate-100 flex items-center justify-center shrink-0 border border-slate-200">
          <ShoppingBag class="w-4 h-4 sm:w-5 sm:h-5 text-slate-600" />
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'table' && showOtherDetails" class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="flex items-center justify-between gap-3 px-4 py-3 border-b border-slate-100">
        <div>
          <h3 class="text-sm font-bold text-slate-900">Rincian Omset MT / Lainnya</h3>
          <p class="text-xs text-slate-500">{{ visitStats.transaksiOther }} transaksi · Rp {{ formatRupiah(visitStats.nominalOther) }}</p>
        </div>
        <button type="button" class="text-xs font-semibold text-slate-500 hover:text-slate-800" @click="showOtherDetails = false">Tutup</button>
      </div>
      <div class="max-h-80 overflow-auto">
        <table class="min-w-full text-xs text-left">
          <thead class="sticky top-0 bg-slate-50 text-slate-600">
            <tr>
              <th class="px-4 py-2 font-bold">Nomor</th>
              <th class="px-4 py-2 font-bold">Kategori</th>
              <th class="px-4 py-2 font-bold">Sales</th>
              <th class="px-4 py-2 font-bold">Toko</th>
              <th class="px-4 py-2 font-bold text-right">Nominal</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="detail in visitStats.otherDetails" :key="`${detail.source}-${detail.id}`" class="border-t border-slate-100">
              <td class="px-4 py-2 font-semibold text-slate-800 whitespace-nowrap">{{ detail.number }}</td>
              <td class="px-4 py-2 text-slate-600 whitespace-nowrap">{{ detail.source }}</td>
              <td class="px-4 py-2 text-slate-600">{{ detail.employee }}</td>
              <td class="px-4 py-2 text-slate-600">{{ detail.store }}</td>
              <td class="px-4 py-2 font-semibold text-slate-800 text-right whitespace-nowrap">Rp {{ formatRupiah(detail.amount) }}</td>
            </tr>
            <tr v-if="visitStats.otherDetails.length === 0">
              <td colspan="5" class="px-4 py-6 text-center text-slate-500">Tidak ada transaksi dalam kategori ini.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Summary Cards (Visible only on Journey / Perjalanan & BBM Tab) -->
    <div v-if="activeTab === 'journey'" class="grid grid-cols-2 lg:grid-cols-4 gap-2.5 sm:gap-4 shrink-0">
      <!-- Total Jarak Tempuh -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-5 border border-slate-200 shadow-sm flex items-center justify-between gap-2.5 sm:gap-4">
        <div class="min-w-0 flex-1">
          <p class="text-[10px] sm:text-xs font-bold text-slate-400">Total Jarak Tempuh</p>
          <p class="text-base sm:text-xl font-black text-slate-800 leading-tight mt-0.5">
            {{ journeyStats.totalKm.toLocaleString('id-ID', { maximumFractionDigits: 1 }) }} <span class="text-[10px] sm:text-xs font-medium text-slate-500">KM</span>
          </p>
        </div>
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl bg-blue-50 flex items-center justify-center shrink-0">
          <Gauge class="w-5 h-5 text-blue-600" />
        </div>
      </div>

      <!-- Rata-rata KM / Sales -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-5 border border-slate-200 shadow-sm flex items-center justify-between gap-2.5 sm:gap-4">
        <div class="min-w-0 flex-1">
          <p class="text-[10px] sm:text-xs font-bold text-slate-400">Rata-rata Jarak / Hari</p>
          <p class="text-base sm:text-xl font-black text-slate-800 leading-tight mt-0.5">
            {{ journeyStats.avgKm.toLocaleString('id-ID', { maximumFractionDigits: 1 }) }} <span class="text-[10px] sm:text-xs font-medium text-slate-500">KM/sales</span>
          </p>
        </div>
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl bg-indigo-50 flex items-center justify-center shrink-0">
          <Navigation2 class="w-5 h-5 text-indigo-600" />
        </div>
      </div>

      <!-- Total Klaim BBM -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-5 border border-slate-200 shadow-sm flex items-center justify-between gap-2.5 sm:gap-4">
        <div class="min-w-0 flex-1">
          <p class="text-[10px] sm:text-xs font-bold text-slate-400">Total Klaim BBM</p>
          <p 
            :class="[getNominalFontSizeClass(journeyStats.totalBbm), 'font-black text-slate-800 tracking-tight leading-tight mt-0.5 whitespace-nowrap']"
            :title="`Rp ${formatRupiah(journeyStats.totalBbm)}`"
          >
            <span class="text-[10px] sm:text-xs font-bold text-amber-600">Rp</span> {{ formatRupiah(journeyStats.totalBbm) }}
          </p>
        </div>
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl bg-amber-50 flex items-center justify-center shrink-0">
          <Fuel class="w-5 h-5 text-amber-600" />
        </div>
      </div>

      <!-- Biaya per KM -->
      <div class="bg-white rounded-2xl p-3.5 sm:p-5 border border-slate-200 shadow-sm flex items-center justify-between gap-2.5 sm:gap-4">
        <div class="min-w-0 flex-1">
          <p class="text-[10px] sm:text-xs font-bold text-slate-400">Estimasi Biaya / KM</p>
          <p 
            :class="[getNominalFontSizeClass(journeyStats.costPerKm), 'font-black text-slate-800 tracking-tight leading-tight mt-0.5 whitespace-nowrap']"
            :title="`Rp ${formatRupiah(journeyStats.costPerKm)} /KM`"
          >
            <span class="text-[10px] sm:text-xs font-bold text-emerald-600">Rp</span> {{ journeyStats.costPerKm.toLocaleString('id-ID', { maximumFractionDigits: 0 }) }} <span class="text-[10px] sm:text-xs font-medium text-slate-500">/KM</span>
          </p>
        </div>
        <div class="w-9 h-9 sm:w-11 sm:h-11 rounded-xl bg-emerald-50 flex items-center justify-center shrink-0">
          <TrendingUp class="w-5 h-5 text-emerald-600" />
        </div>
      </div>
    </div>

    <!-- Main Content Area -->
    <div
      class="flex-1 min-h-0 bg-white rounded-2xl sm:rounded-[32px] border border-slate-200 shadow-sm overflow-hidden flex flex-col"
    >
      <div v-if="activeTab === 'table'" class="flex-1 flex flex-col min-h-0">
        <div class="flex-1 overflow-auto no-scrollbar">
          <table class="w-full min-w-[750px] border-collapse text-left">
            <thead class="sticky top-0 bg-white/80 backdrop-blur-md z-10">
              <tr class="border-b border-slate-100">
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Salesman
                </th>
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Toko / Customer
                </th>
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Tanggal
                </th>
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Waktu
                </th>
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Hasil
                </th>
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Catatan
                </th>
                <th
                  class="p-4 sm:p-6 text-right text-[11px] font-black text-slate-400 uppercase tracking-widest"
                >
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50">
              <tr v-if="isLoading" v-for="i in 5" :key="i" class="animate-pulse">
                <td v-for="j in 7" :key="j" class="p-4 sm:p-6">
                  <div class="h-4 bg-slate-100 rounded-lg w-full"></div>
                </td>
              </tr>

              <tr v-else-if="filteredVisits.length === 0">
                <td colspan="7" class="py-24 text-center">
                  <div class="flex flex-col items-center gap-3">
                    <AlertCircle class="w-12 h-12 text-slate-200" />
                    <p class="font-bold text-slate-400">Tidak ada data kunjungan ditemukan.</p>
                  </div>
                </td>
              </tr>

              <tr
                v-for="visit in filteredVisits"
                :key="visit.id"
                class="group hover:bg-slate-50/50 transition-all"
              >
                <td class="p-4 sm:p-6">
                  <div>
                    <p class="font-bold text-slate-900">
                      {{ visit.employee?.first_name }} {{ visit.employee?.last_name }}
                    </p>
                    <p class="text-[10px] font-bold text-slate-400">
                      {{ visit.employee?.branch?.name || 'Cabang' }}
                    </p>
                  </div>
                </td>
                <td class="p-4 sm:p-6">
                  <div class="max-w-[200px] relative group/addr">
                    <p class="font-bold text-slate-900 truncate">
                      {{ visit.store?.name || 'New Store' }}
                    </p>
                    <p class="text-[10px] font-medium text-slate-400 truncate cursor-pointer">
                      {{ visit.store?.address }}
                    </p>
                    <!-- Tooltip preview -->
                    <div
                      class="absolute left-0 bottom-full mb-1 hidden group-hover/addr:block bg-slate-900 text-white text-[11px] p-3 rounded-xl shadow-xl z-50 max-w-[250px] break-words font-medium pointer-events-none transition-all"
                    >
                      {{ visit.store?.address || 'Alamat tidak tersedia' }}
                    </div>
                  </div>
                </td>
                <td class="p-4 sm:p-6">
                  <div>
                    <p class="text-xs font-black text-slate-700">
                      {{ formatDate(visit.check_in_time) }}
                    </p>
                  </div>
                </td>
                <td class="p-4 sm:p-6">
                  <div>
                    <p class="text-xs font-black text-slate-700">
                      IN: {{ formatTime(visit.check_in_time) }}
                    </p>
                    <p class="text-[10px] font-bold text-slate-400" v-if="visit.check_out_time">
                      OUT: {{ formatTime(visit.check_out_time) }}
                    </p>
                  </div>
                </td>

                <td class="p-4 sm:p-6">
                  <div class="flex flex-col gap-1">
                    <div
                      v-if="(visit.transactions?.length || 0) + (visit.sales_orders?.length || 0) > 0"
                      class="flex items-center gap-1.5"
                    >
                      <CheckCircle2 class="w-3.5 h-3.5 text-emerald-500" />
                      <span
                        class="text-[10px] font-black text-emerald-600 uppercase tracking-tighter"
                      >
                        Order ({{ (visit.transactions?.length || 0) + (visit.sales_orders?.length || 0) }})
                      </span>
                    </div>

                    <div
                      v-if="visit.transactions?.length || visit.sales_orders?.length"
                      class="ml-4 space-y-1.5 text-xs text-slate-700 mt-1"
                    >
                      <!-- Transaksi -->
                      <div v-for="trx in visit.transactions" :key="'trx-'+trx.id" class="leading-snug">
                        <span class="font-medium text-slate-500">{{ trx.invoice_no || 'Transaksi' }}: </span>
                        <span v-for="(item, i) in trx.items" :key="'trx-item-'+i" class="font-bold text-slate-900">
                          {{ item.product?.name || 'Item' }} ({{ item.quantity }} {{ item.unit }})<template v-if="Number(i) < (trx.items?.length || 0) - 1">, </template>
                        </span>
                      </div>
                      <!-- Sales Order -->
                      <div v-for="so in visit.sales_orders" :key="'so-'+so.id" class="leading-snug">
                        <span class="font-medium text-slate-500">{{ so.so_number || 'SO' }}: </span>
                        <span v-for="(item, i) in so.items" :key="'so-item-'+i" class="font-bold text-slate-900">
                          {{ item.product?.name || 'Item' }} ({{ item.quantity }} {{ item.unit }})<template v-if="Number(i) < (so.items?.length || 0) - 1">, </template>
                        </span>
                      </div>
                    </div>

                    <div v-if="!visit.transactions?.length && !visit.sales_orders?.length" class="flex items-center gap-1.5 text-slate-400">
                      <XCircle class="w-3.5 h-3.5" />
                      <span class="text-[10px] font-black uppercase tracking-tighter">No Order</span>
                    </div>
                  </div>
                </td>
                <td class="p-4 sm:p-6">
                  <div class="max-w-[200px]">
                    <div v-if="visit.notes && visit.notes.includes('[SISTEM: Kunjungan berulang')" class="mb-1">
                      <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-800 border border-amber-300">
                        <span>⚠️</span> Kunjungan &lt; 2 Hari
                      </span>
                    </div>
                    <p class="text-[11px] font-medium text-slate-600 truncate" :title="visit.notes">
                      {{ visit.notes || '-' }}
                    </p>
                  </div>
                </td>
                <td class="p-4 sm:p-6 text-right">
                  <div class="flex justify-end gap-1">
                    <button
                      @click="openDetail(visit)"
                      class="p-2 hover:bg-white hover:shadow-md rounded-xl transition-all text-slate-400 hover:text-primary border border-transparent hover:border-slate-100"
                      title="Detail"
                    >
                      <Maximize2 class="w-4 h-4" />
                    </button>
                    <button
                      @click="openEdit(visit)"
                      class="p-2 hover:bg-white hover:shadow-md rounded-xl transition-all text-slate-400 hover:text-amber-500 border border-transparent hover:border-slate-100"
                      title="Edit"
                    >
                      <Edit class="w-4 h-4" />
                    </button>
                    <button
                      @click="handleDelete(visit)"
                      class="p-2 hover:bg-white hover:shadow-md rounded-xl transition-all text-slate-400 hover:text-red-500 border border-transparent hover:border-slate-100"
                      title="Hapus"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination UI -->
        <div
          class="flex flex-col sm:flex-row items-center justify-between p-3.5 sm:p-6 border-t border-slate-100 bg-white gap-3 sm:gap-4 shrink-0"
        >
          <div class="flex items-center gap-2">
            <span class="text-xs font-bold text-slate-500">Tampilkan</span>
            <select
              v-model="itemsPerPage"
              @change="onItemsPerPageChange"
              class="px-3 py-1.5 bg-white border border-slate-200 rounded-xl text-xs font-black text-slate-700 outline-none shadow-sm focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer"
            >
              <option :value="10">10</option>
              <option :value="20">20</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
              <option :value="-1">Semua</option>
            </select>
            <span class="text-xs font-bold text-slate-500">entri</span>
          </div>

          <div class="text-xs font-semibold text-slate-500 text-center sm:text-left">
            {{ paginationInfo }}
          </div>

          <div class="flex items-center gap-1.5 sm:gap-2 flex-wrap justify-center">
            <Button
              variant="outline"
              size="sm"
              class="rounded-xl font-bold h-8 sm:h-9 text-xs border-slate-200 hover:bg-slate-50"
              :disabled="currentPage === 1 || itemsPerPage === -1"
              @click="prevPage"
            >
              Sebelumnya
            </Button>
            <div class="flex items-center gap-1">
              <Button
                v-for="page in visiblePages"
                :key="page"
                :variant="currentPage === page ? 'default' : 'outline'"
                size="sm"
                class="rounded-xl font-black h-8 sm:h-9 w-8 sm:w-9 p-0 text-xs border-slate-200"
                @click="goToPage(page)"
              >
                {{ page }}
              </Button>
            </div>
            <Button
              variant="outline"
              size="sm"
              class="rounded-xl font-bold h-8 sm:h-9 text-xs border-slate-200 hover:bg-slate-50"
              :disabled="currentPage === totalPages || itemsPerPage === -1"
              @click="nextPage"
            >
              Selanjutnya
            </Button>
          </div>
        </div>
      </div>

      <!-- Rencana Kunjungan View -->
      <div v-if="activeTab === 'schedule'" class="flex-1 flex flex-col min-h-0">
        <div class="flex-1 overflow-auto no-scrollbar">
          <!-- State 1: No salesman selected -->
          <div v-if="selectedEmployee === 'ALL'" class="py-24 text-center">
            <div class="flex flex-col items-center gap-3">
              <User class="w-12 h-12 text-slate-300" />
              <p class="font-bold text-slate-500">
                Silakan pilih salesman di filter atas untuk melihat rencana kunjungan.
              </p>
            </div>
          </div>

          <!-- State 2: Loading plans -->
          <div v-else-if="isLoadingPlans" class="p-6 divide-y divide-slate-100 animate-pulse">
            <div v-for="i in 4" :key="i" class="py-6 flex flex-col gap-2">
              <div class="h-5 bg-slate-100 rounded-lg w-1/3"></div>
              <div class="h-4 bg-slate-100 rounded-lg w-1/2"></div>
              <div class="h-4 bg-slate-100 rounded-lg w-1/4"></div>
            </div>
          </div>

          <!-- State 3: Empty plans -->
          <div v-else-if="visitPlans.length === 0" class="py-24 text-center">
            <div class="flex flex-col items-center gap-3">
              <AlertCircle class="w-12 h-12 text-slate-200" />
              <p class="font-bold text-slate-400">
                Tidak ada rencana kunjungan hari ini untuk salesman terpilih.
              </p>
            </div>
          </div>

          <!-- State 4: List of Plans -->
          <table v-else class="w-full min-w-[650px] border-collapse text-left">
            <thead class="sticky top-0 bg-white/80 backdrop-blur-md z-10">
              <tr class="border-b border-slate-100">
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Toko / Customer
                </th>
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Status Rencana
                </th>
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Kunjungan Terakhir
                </th>
                <th class="p-4 sm:p-6 text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Transaksi Terakhir
                </th>
                <th class="p-4 sm:p-6 text-right text-[11px] font-black text-slate-400 uppercase tracking-widest">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50">
              <tr
                v-for="plan in visitPlans"
                :key="plan.store_id"
                class="group hover:bg-slate-50/50 transition-all"
              >
                <td class="p-4 sm:p-6">
                  <div class="max-w-[300px] relative group/addr">
                    <p class="font-bold text-slate-900 truncate">{{ plan.store_name }}</p>
                    <p class="text-[10px] font-medium text-slate-400 truncate cursor-pointer">
                      {{ plan.address }}
                    </p>
                    <div
                      class="absolute left-0 bottom-full mb-1 hidden group-hover/addr:block bg-slate-900 text-white text-[11px] p-3 rounded-xl shadow-xl z-50 max-w-[250px] break-words font-medium pointer-events-none transition-all"
                    >
                      {{ plan.address || 'Alamat tidak tersedia' }}
                    </div>
                  </div>
                </td>
                <td class="p-4 sm:p-6">
                  <span
                    v-if="plan.status === 'COMPLETED'"
                    class="px-3 py-1 rounded-full text-[10px] font-black border uppercase text-emerald-600 bg-emerald-50 border-emerald-100"
                  >
                    Selesai (Completed)
                  </span>
                  <span
                    v-else-if="plan.status === 'CHECKED_IN'"
                    class="px-3 py-1 rounded-full text-[10px] font-black border uppercase text-blue-600 bg-blue-50 border-blue-100"
                  >
                    Sedang Berkunjung (Checked In)
                  </span>
                  <span
                    v-else-if="plan.status === 'EXTRA'"
                    class="px-3 py-1 rounded-full text-[10px] font-black border uppercase text-purple-600 bg-purple-50 border-purple-100"
                  >
                    Di Luar Rencana (Extra)
                  </span>
                  <span
                    v-else
                    class="px-3 py-1 rounded-full text-[10px] font-black border uppercase text-slate-500 bg-slate-50 border-slate-200"
                  >
                    Rencana (Planned)
                  </span>
                </td>
                <td class="p-4 sm:p-6">
                  <div v-if="plan.last_visit_date">
                    <p class="text-xs font-black text-slate-700">
                      {{ formatDate(plan.last_visit_date) }}
                    </p>
                    <p class="text-[10px] font-bold text-slate-400">
                      {{ formatTime(plan.last_visit_date) }}
                    </p>
                  </div>
                  <div v-else>
                    <span class="text-xs text-slate-400 font-medium italic"
                      >Belum pernah dikunjungi</span
                    >
                  </div>
                </td>
                <td class="p-4 sm:p-6">
                  <div v-if="plan.last_transaction_date">
                    <p class="text-xs font-black text-slate-700">
                      {{ formatDate(plan.last_transaction_date) }}
                    </p>
                    <p class="text-xs font-black text-emerald-600 mt-0.5">
                      Rp {{ new Intl.NumberFormat('id-ID').format(plan.last_transaction_amount) }}
                    </p>
                    <p
                      class="text-[10px] font-medium text-slate-400 mt-0.5 truncate max-w-[200px]"
                      :title="plan.last_transaction_items"
                    >
                      {{ plan.last_transaction_items || 'Tanpa detail produk' }}
                    </p>
                  </div>
                  <div v-else>
                    <span class="text-xs text-slate-400 font-medium italic"
                      >Belum ada transaksi</span
                    >
                  </div>
                </td>
                <td class="p-4 sm:p-6 text-right">
                  <button
                    v-if="plan.status !== 'COMPLETED' && plan.status !== 'CHECKED_IN'"
                    @click="realizePlan(plan)"
                    :disabled="isRealizingPlan === plan.store_id"
                    class="px-3 py-1.5 bg-primary/10 hover:bg-primary text-primary hover:text-white rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 ml-auto shadow-sm cursor-pointer disabled:opacity-50"
                  >
                    <CheckCircle2 class="w-3.5 h-3.5" />
                    <span>{{ isRealizingPlan === plan.store_id ? 'Menyimpan...' : 'Realisasikan' }}</span>
                  </button>
                  <span v-else class="text-xs text-emerald-600 font-bold flex items-center justify-end gap-1">
                    <CheckCircle2 class="w-4 h-4" /> Terealisasi
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Leaflet Map View -->
      <div v-show="activeTab === 'map'" class="flex-1 w-full relative min-h-[400px] sm:min-h-[550px]">
        <div ref="mapContainer" class="absolute inset-0 z-0"></div>

        <!-- Floating Map Controls (Top Right) -->
        <div
          v-if="filteredVisits.length > 0"
          class="absolute top-3 right-3 z-[400] flex flex-col items-end gap-2 pointer-events-auto"
        >
          <button
            @click="fitMapToBounds"
            type="button"
            class="bg-white/95 hover:bg-white text-slate-700 hover:text-primary px-3 py-1.5 rounded-xl text-xs font-bold shadow-lg border border-slate-200/80 backdrop-blur-sm flex items-center gap-1.5 transition-all active:scale-95 cursor-pointer"
            title="Pusatkan Peta ke Seluruh Kunjungan"
          >
            <Compass class="w-3.5 h-3.5 text-primary" />
            <span class="hidden sm:inline">Pusatkan Peta</span>
          </button>
        </div>

        <!-- Floating Legend / Route Summary (Bottom Left) -->
        <div
          v-if="filteredVisits.length > 0"
          class="absolute bottom-4 left-3 z-[400] bg-white/95 backdrop-blur-sm border border-slate-200/80 rounded-xl px-3 py-2 shadow-lg max-w-[240px] sm:max-w-xs text-[11px] pointer-events-auto"
        >
          <div class="flex items-center gap-1.5 font-bold text-slate-800">
            <RouteIcon class="w-3.5 h-3.5 text-primary" />
            <span>Urutan Rute Kunjungan</span>
          </div>
          <p class="text-[10px] text-slate-500 mt-0.5 leading-tight">
            Nomor 1, 2, 3... berurutan sesuai jam check-in kunjungan.
          </p>
        </div>

        <div
          v-if="filteredVisits.length === 0"
          class="absolute inset-0 z-10 bg-white/80 backdrop-blur-sm flex flex-col items-center justify-center"
        >
          <MapIcon class="w-12 h-12 text-slate-300 mb-3" />
          <p class="font-bold text-slate-500 text-center px-4">Tidak ada kunjungan untuk ditampilkan di peta.</p>
        </div>
      </div>

      <!-- Omzet Validasi View -->
      <div
        v-if="activeTab === 'omzet'"
        class="flex-1 overflow-y-auto no-scrollbar p-3 sm:p-6 bg-slate-50/50"
      >
        <!-- Dashboard Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-6 mb-4 sm:mb-8">
          <!-- Pending Card -->
          <div
            class="bg-white rounded-2xl sm:rounded-3xl p-4 sm:p-6 border border-orange-100 shadow-xl shadow-orange-500/5 relative overflow-hidden"
          >
            <div class="absolute -right-6 -top-6">
              <ClockIcon class="w-24 h-24 sm:w-32 sm:h-32 text-orange-50" />
            </div>
            <div class="relative z-10">
              <div class="flex items-center gap-3 mb-4 sm:mb-6">
                <div class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-orange-100 flex items-center justify-center shrink-0">
                  <ClockIcon class="w-4 h-4 sm:w-5 sm:h-5 text-orange-600" />
                </div>
                <div>
                  <h3 class="font-black text-slate-900 text-sm sm:text-base">Belum Divalidasi Kasir</h3>
                  <p class="text-[10px] sm:text-xs font-bold text-slate-400">Menunggu Penyetoran</p>
                </div>
              </div>
              <p 
                :class="[getNominalFontSizeClass(totalPendingAmount), 'font-black text-orange-600 mb-2 whitespace-nowrap tracking-tight']"
                :title="`Rp ${formatRupiah(totalPendingAmount)}`"
              >
                <span class="text-sm sm:text-lg font-bold">Rp</span> {{ formatRupiah(totalPendingAmount) }}
              </p>
              <div class="inline-flex items-center gap-1.5 px-2.5 py-1 sm:px-3 sm:py-1.5 bg-orange-50 rounded-lg">
                <AlertCircle class="w-3.5 h-3.5 text-orange-500" />
                <span class="text-[11px] sm:text-xs font-black text-orange-600"
                  >{{ pendingTransactions.length }} Nota Tertahan</span
                >
              </div>
            </div>
          </div>

          <!-- Verified Card -->
          <div
            class="bg-white rounded-2xl sm:rounded-3xl p-4 sm:p-6 border border-emerald-100 shadow-xl shadow-emerald-500/5 relative overflow-hidden"
          >
            <div class="absolute -right-6 -top-6">
              <CheckCircle class="w-24 h-24 sm:w-32 sm:h-32 text-emerald-50" />
            </div>
            <div class="relative z-10">
              <div class="flex items-center gap-3 mb-4 sm:mb-6">
                <div class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-emerald-100 flex items-center justify-center shrink-0">
                  <CheckCircle class="w-4 h-4 sm:w-5 sm:h-5 text-emerald-600" />
                </div>
                <div>
                  <h3 class="font-black text-slate-900 text-sm sm:text-base">Sudah Divalidasi</h3>
                  <p class="text-[10px] sm:text-xs font-bold text-slate-400">Telah Diterima Kasir</p>
                </div>
              </div>
              <p 
                :class="[getNominalFontSizeClass(totalVerifiedAmount), 'font-black text-emerald-600 mb-2 whitespace-nowrap tracking-tight']"
                :title="`Rp ${formatRupiah(totalVerifiedAmount)}`"
              >
                <span class="text-sm sm:text-lg font-bold">Rp</span> {{ formatRupiah(totalVerifiedAmount) }}
              </p>
              <div class="inline-flex items-center gap-1.5 px-2.5 py-1 sm:px-3 sm:py-1.5 bg-emerald-50 rounded-lg">
                <CheckCircle2 class="w-3.5 h-3.5 text-emerald-500" />
                <span class="text-[11px] sm:text-xs font-black text-emerald-600"
                  >{{ verifiedTransactions.length }} Nota Selesai</span
                >
              </div>
            </div>
          </div>
        </div>

        <!-- Transactions Table -->
        <div class="bg-white rounded-2xl sm:rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
          <div class="p-4 sm:p-6 border-b border-slate-100 flex items-center justify-between">
            <h3 class="font-black text-slate-900 text-sm sm:text-base truncate">
              Rincian Transaksi {{ selectedDate ? formatDate(selectedDate) : 'Hari Ini' }}
            </h3>
            <span class="text-[10px] sm:text-xs font-bold text-slate-400 shrink-0 ml-2"
              >{{ filteredTransactions.length }} Transaksi</span
            >
          </div>
          <div class="overflow-x-auto">
            <table class="w-full min-w-[700px] text-left">
              <thead class="bg-slate-50/50">
                <tr>
                  <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                    No. Faktur
                  </th>
                  <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                    Waktu
                  </th>
                  <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                    Salesman
                  </th>
                  <th class="p-4 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                    Toko
                  </th>
                  <th
                    class="p-4 text-right text-[10px] font-black text-slate-400 uppercase tracking-widest"
                  >
                    Nominal
                  </th>
                  <th
                    class="p-4 text-center text-[10px] font-black text-slate-400 uppercase tracking-widest"
                  >
                    Metode Bayar
                  </th>
                  <th
                    class="p-4 text-center text-[10px] font-black text-slate-400 uppercase tracking-widest"
                  >
                    Status Setoran
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-50">
                <tr v-if="isLoadingTransactions" v-for="i in 3" :key="i" class="animate-pulse">
                  <td colspan="7" class="p-4">
                    <div class="h-4 bg-slate-100 rounded w-full"></div>
                  </td>
                </tr>
                <tr v-else-if="paginatedTransactions.length === 0">
                  <td colspan="7" class="p-12 text-center text-slate-400 font-bold">
                    Tidak ada transaksi ditemukan.
                  </td>
                </tr>
                <tr
                  v-else
                  v-for="t in paginatedTransactions"
                  :key="t.id"
                  class="hover:bg-slate-50/50 transition-colors"
                >
                  <td class="p-4">
                    <span class="font-black text-slate-700 text-sm">{{ t.receipt_no }}</span>
                  </td>
                  <td class="p-4">
                    <span class="font-bold text-slate-500 text-xs">{{
                      formatTime(t.visit?.check_in_time || t.created_at || t.transaction_date)
                    }}</span>
                  </td>
                  <td class="p-4">
                    <span class="font-bold text-slate-900 text-sm"
                      >{{ t.employee?.first_name }} {{ t.employee?.last_name }}</span
                    >
                  </td>
                  <td class="p-4">
                    <p class="font-bold text-slate-900 text-sm max-w-[200px] truncate">
                      {{ t.store?.name }}
                    </p>
                    <p class="font-medium text-slate-400 text-[10px] uppercase">
                      {{ t.store_category?.replace('_', ' ') }}
                    </p>
                  </td>
                  <td class="p-4 text-right">
                    <span class="font-black text-slate-900 text-sm"
                      >Rp {{ new Intl.NumberFormat('id-ID').format(t.total_amount) }}</span
                    >
                    <!-- DP info for TEMPO -->
                    <p
                      v-if="t.payment_method === 'TEMPO' && t.paid_amount > 0"
                      class="text-[10px] font-bold text-orange-500 mt-0.5"
                    >
                      DP: Rp {{ new Intl.NumberFormat('id-ID').format(t.paid_amount) }}
                    </p>
                    <p
                      v-if="t.payment_method === 'TEMPO' && t.paid_amount > 0"
                      class="text-[10px] font-bold text-slate-400"
                    >
                      Sisa: Rp
                      {{ new Intl.NumberFormat('id-ID').format(t.total_amount - t.paid_amount) }}
                    </p>
                  </td>
                  <!-- Payment Method -->
                  <td class="p-4 text-center">
                    <div class="flex flex-col items-center gap-1">
                      <!-- CASH -->
                      <span
                        v-if="t.payment_method === 'CASH'"
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-green-100 text-green-700 text-[10px] font-black uppercase border border-green-200"
                      >
                        💵 CASH
                      </span>
                      <!-- TEMPO with DP Cash -->
                      <span
                        v-else-if="
                          t.payment_method === 'TEMPO' &&
                          t.dp_payment_method === 'CASH' &&
                          t.paid_amount > 0
                        "
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-orange-100 text-orange-700 text-[10px] font-black uppercase border border-orange-200"
                      >
                        ⏳ TEMPO + DP CASH
                      </span>
                      <!-- TEMPO with DP QRIS -->
                      <span
                        v-else-if="
                          t.payment_method === 'TEMPO' &&
                          t.dp_payment_method === 'QRIS' &&
                          t.paid_amount > 0
                        "
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-purple-100 text-purple-700 text-[10px] font-black uppercase border border-purple-200"
                      >
                        ⏳ TEMPO + DP QRIS
                      </span>
                      <!-- TEMPO with DP VA -->
                      <span
                        v-else-if="
                          t.payment_method === 'TEMPO' &&
                          t.dp_payment_method === 'VA' &&
                          t.paid_amount > 0
                        "
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-blue-100 text-blue-700 text-[10px] font-black uppercase border border-blue-200"
                      >
                        ⏳ TEMPO + DP BANK
                      </span>
                      <!-- TEMPO pure (no DP) -->
                      <span
                        v-else-if="t.payment_method === 'TEMPO'"
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-orange-100 text-orange-700 text-[10px] font-black uppercase border border-orange-200"
                      >
                        ⏳ TEMPO
                      </span>
                      <!-- QRIS -->
                      <span
                        v-else-if="t.payment_method === 'QRIS'"
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-purple-100 text-purple-700 text-[10px] font-black uppercase border border-purple-200"
                      >
                        📲 QRIS
                      </span>
                      <!-- VA -->
                      <span
                        v-else-if="t.payment_method === 'VA'"
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-blue-100 text-blue-700 text-[10px] font-black uppercase border border-blue-200"
                      >
                        🏦 BANK
                      </span>
                      <!-- Unknown -->
                      <span
                        v-else
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-slate-100 text-slate-500 text-[10px] font-black uppercase"
                      >
                        {{ t.payment_method || '-' }}
                      </span>
                    </div>
                  </td>
                  <!-- Status Setoran -->
                  <td class="p-4 text-center">
                    <!-- TEMPO with DP Cash (PENDING = waiting cashier approval) -->
                    <div v-if="t.payment_method === 'TEMPO' && t.paid_amount > 0">
                      <span
                        v-if="t.payment_status === 'PARTIAL'"
                        class="inline-flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-full bg-amber-100 text-amber-700 text-[10px] font-black uppercase border border-amber-200"
                      >
                        <span>DP TERVALIDASI</span>
                        <span class="text-[9px] font-bold text-amber-500">Sisa belum lunas</span>
                      </span>
                      <span
                        v-else-if="t.payment_status === 'PAID'"
                        class="inline-flex px-3 py-1 rounded-full bg-emerald-100 text-emerald-700 text-[10px] font-black uppercase border border-emerald-200"
                      >
                        LUNAS ✓
                      </span>
                      <span
                        v-else
                        class="inline-flex flex-col items-center gap-0.5 px-3 py-1.5 rounded-full bg-orange-100 text-orange-700 text-[10px] font-black uppercase border border-orange-200"
                      >
                        <span>DP BELUM SETOR</span>
                        <span class="text-[9px] font-bold text-orange-500"
                          >Menunggu validasi kasir</span
                        >
                      </span>
                    </div>
                    <!-- Regular status -->
                    <template v-else>
                      <span
                        v-if="t.status === 'PENDING' && t.payment_status !== 'PAID'"
                        class="inline-flex px-3 py-1 rounded-full bg-orange-100 text-orange-700 text-[10px] font-black uppercase border border-orange-200"
                      >
                        BELUM SETOR
                      </span>
                      <span
                        v-else-if="t.status === 'VERIFIED' || t.payment_status === 'PAID'"
                        class="inline-flex px-3 py-1 rounded-full bg-emerald-100 text-emerald-700 text-[10px] font-black uppercase border border-emerald-200"
                      >
                        TERVALIDASI
                      </span>
                      <span
                        v-else
                        class="inline-flex px-3 py-1 rounded-full bg-slate-100 text-slate-700 text-[10px] font-black uppercase"
                      >
                        {{ t.status }}
                      </span>
                    </template>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination UI for Omzet Tab -->
          <div v-if="omzetTotalPages > 1" class="flex flex-col sm:flex-row items-center justify-between p-4 sm:p-6 border-t border-slate-100 bg-white gap-3">
            <div class="text-xs font-bold text-slate-500 text-center sm:text-left">
              Menampilkan Halaman {{ omzetCurrentPage }} dari {{ omzetTotalPages }}
            </div>
            <div class="flex items-center gap-2">
              <button
                @click="omzetCurrentPage--"
                :disabled="omzetCurrentPage === 1"
                class="px-3 py-1.5 rounded-lg border border-slate-200 text-xs font-bold text-slate-600 disabled:opacity-50 hover:bg-slate-50 transition-colors"
              >
                Sebelumnya
              </button>
              <button
                @click="omzetCurrentPage++"
                :disabled="omzetCurrentPage === omzetTotalPages"
                class="px-3 py-1.5 rounded-lg border border-slate-200 text-xs font-bold text-slate-600 disabled:opacity-50 hover:bg-slate-50 transition-colors"
              >
                Berikutnya
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- ─── Rekap Sales Tab ─────────────────────────────────────────── -->
      <div v-if="activeTab === 'rekap'" class="flex-1 flex flex-col min-h-0">
        <!-- Loading State -->
        <div v-if="isLoadingRekap" class="flex-1 flex flex-col items-center justify-center gap-3 py-20">
          <div class="w-10 h-10 border-4 border-slate-100 border-t-primary rounded-full animate-spin"></div>
          <p class="text-sm font-bold text-slate-400">Memuat rekap kunjungan...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="rekapSummary.length === 0" class="flex-1 flex flex-col items-center justify-center gap-3 py-20">
          <Users class="w-12 h-12 text-slate-200" />
          <p class="font-bold text-slate-400">Belum ada data kunjungan untuk rekap.</p>
          <p class="text-xs text-slate-300">Pilih tanggal atau bulan yang memiliki data kunjungan.</p>
        </div>

        <!-- Rekap Grid -->
        <div v-else class="flex-1 overflow-y-auto no-scrollbar">
          <!-- Header Info -->
          <div class="px-3 sm:px-6 pt-3.5 sm:pt-5 pb-3 border-b border-slate-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
            <div>
              <p class="text-[10px] sm:text-xs font-black text-slate-400 uppercase tracking-widest">Rekap Kunjungan Sales</p>
              <p class="text-xs sm:text-sm font-bold text-slate-700 mt-0.5">
                {{ selectedDate ? formatDate(selectedDate) : selectedMonth || 'Semua Tanggal' }}
                — <span class="text-primary">{{ rekapSummary.length }} Salesman Aktif</span>
              </p>
            </div>
            <div class="flex items-center gap-2 text-xs font-bold text-slate-400">
              <span class="w-2.5 h-2.5 sm:w-3 sm:h-3 rounded-full bg-emerald-400 inline-block"></span> Toko Baru
              <span class="w-2.5 h-2.5 sm:w-3 sm:h-3 rounded-full bg-slate-300 inline-block ml-2"></span> Toko Lama
            </div>
          </div>

          <!-- Cards Grid -->
          <div class="p-3 sm:p-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 sm:gap-4">
            <div
              v-for="(entry, idx) in rekapSummary"
              :key="entry.employeeId"
              class="bg-white rounded-2xl border border-slate-200 shadow-sm hover:shadow-md hover:border-primary/20 transition-all duration-200 overflow-hidden"
            >
              <!-- Card Header -->
              <div class="p-3.5 sm:p-4 bg-gradient-to-r from-slate-50 to-white border-b border-slate-100 flex items-center gap-3">
                <div class="w-8 h-8 sm:w-9 sm:h-9 rounded-xl bg-primary/10 flex items-center justify-center font-black text-primary text-xs sm:text-sm shrink-0">
                  {{ (entry.name.charAt(0) || '#').toUpperCase() }}
                </div>
                <div class="min-w-0">
                  <p class="font-black text-slate-900 text-xs sm:text-sm truncate leading-tight">{{ entry.name }}</p>
                  <p class="text-[9px] sm:text-[10px] font-bold text-slate-400 uppercase tracking-wide truncate">{{ entry.branch }}</p>
                </div>
                <div class="ml-auto shrink-0">
                  <span class="text-[10px] font-black text-slate-400">#{{ idx + 1 }}</span>
                </div>
              </div>

              <!-- Stats -->
              <div class="p-3.5 sm:p-4 space-y-2.5 sm:space-y-3">
                <!-- Total Kunjungan -->
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <MapPin class="w-3.5 h-3.5 text-blue-400 shrink-0" />
                    <span class="text-[10px] sm:text-[11px] font-bold text-slate-500">Total Kunjungan</span>
                  </div>
                  <div class="flex items-center gap-1">
                    <span
                      class="text-xs sm:text-sm font-black"
                      :class="entry.targetDailyVisits > 0 && entry.totalKunjungan >= entry.targetDailyVisits ? 'text-emerald-600' : 'text-slate-900'"
                    >
                      {{ entry.totalKunjungan }}
                    </span>
                    <span class="text-xs font-bold text-slate-300">/</span>
                    <span class="text-xs font-bold text-slate-400">{{ entry.targetDailyVisits || '-' }}</span>
                  </div>
                </div>
                <!-- Progress bar total kunjungan -->
                <div class="h-1.5 bg-slate-100 rounded-full overflow-hidden" v-if="entry.targetDailyVisits > 0">
                  <div
                    class="h-full rounded-full transition-all duration-500"
                    :class="entry.totalKunjungan >= entry.targetDailyVisits ? 'bg-emerald-500' : 'bg-blue-400'"
                    :style="`width: ${Math.min((entry.totalKunjungan / entry.targetDailyVisits) * 100, 100)}%`"
                  ></div>
                </div>

                <!-- Toko Baru vs Lama -->
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <Store class="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                    <span class="text-[10px] sm:text-[11px] font-bold text-slate-500">Toko Baru</span>
                  </div>
                  <div class="flex items-center gap-1">
                    <span
                      class="text-xs sm:text-sm font-black"
                      :class="entry.targetDailyNewStores > 0 && entry.tokoBaru >= entry.targetDailyNewStores ? 'text-emerald-600' : 'text-slate-900'"
                    >
                      {{ entry.tokoBaru }}
                    </span>
                    <span class="text-xs font-bold text-slate-300">/</span>
                    <span class="text-xs font-bold text-slate-400">{{ entry.targetDailyNewStores || '-' }}</span>
                  </div>
                </div>
                <!-- Progress bar toko baru -->
                <div class="h-1.5 bg-slate-100 rounded-full overflow-hidden" v-if="entry.targetDailyNewStores > 0">
                  <div
                    class="h-full rounded-full bg-emerald-500 transition-all duration-500"
                    :style="`width: ${Math.min((entry.tokoBaru / entry.targetDailyNewStores) * 100, 100)}%`"
                  ></div>
                </div>

                <!-- Toko Lama -->
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <Store class="w-3.5 h-3.5 text-slate-300 shrink-0" />
                    <span class="text-[10px] sm:text-[11px] font-bold text-slate-400">Toko Lama</span>
                  </div>
                  <span class="text-xs sm:text-sm font-bold text-slate-500">{{ entry.tokoLama }}</span>
                </div>

                <!-- Banner / Spanduk -->
                <div class="flex items-center justify-between pt-1 border-t border-dashed border-slate-100">
                  <div class="flex items-center gap-2">
                    <TrendingUp class="w-3.5 h-3.5 text-violet-400 shrink-0" />
                    <span class="text-[10px] sm:text-[11px] font-bold text-slate-500">Banner / Spanduk</span>
                  </div>
                  <div class="flex items-center gap-1">
                    <span class="text-xs sm:text-sm font-black text-slate-900">{{ entry.spandukDibuat }}</span>
                    <span class="text-[10px] font-bold text-slate-400">dibuat</span>
                  </div>
                </div>
              </div>

              <!-- Achievement Badge -->
              <div
                v-if="entry.targetDailyVisits > 0 && entry.totalKunjungan >= entry.targetDailyVisits"
                class="px-3.5 sm:px-4 pb-3"
              >
                <div class="flex items-center gap-1.5 bg-emerald-50 text-emerald-700 rounded-xl px-2.5 sm:px-3 py-1 sm:py-1.5 border border-emerald-100">
                  <CheckCircle2 class="w-3 h-3 shrink-0" />
                  <span class="text-[9px] sm:text-[10px] font-black uppercase tracking-wide">Target Tercapai</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Rekap Footer Summary -->
          <div class="mx-3 sm:mx-6 mb-3 sm:mb-6 bg-slate-900 rounded-2xl p-4 sm:p-5 grid grid-cols-2 sm:grid-cols-3 md:flex md:flex-wrap gap-4 sm:gap-6">
            <div>
              <p class="text-[9px] sm:text-[10px] font-black text-slate-400 uppercase tracking-widest">Total Salesman</p>
              <p class="text-lg sm:text-2xl font-black text-white mt-0.5 sm:mt-1">{{ rekapSummary.length }}</p>
            </div>
            <div>
              <p class="text-[9px] sm:text-[10px] font-black text-slate-400 uppercase tracking-widest">Total Kunjungan</p>
              <p class="text-lg sm:text-2xl font-black text-blue-400 mt-0.5 sm:mt-1">{{ rekapSummary.reduce((s, e) => s + e.totalKunjungan, 0) }}</p>
            </div>
            <div>
              <p class="text-[9px] sm:text-[10px] font-black text-slate-400 uppercase tracking-widest">Total Toko Baru</p>
              <p class="text-lg sm:text-2xl font-black text-emerald-400 mt-0.5 sm:mt-1">{{ rekapSummary.reduce((s, e) => s + e.tokoBaru, 0) }}</p>
            </div>
            <div>
              <p class="text-[9px] sm:text-[10px] font-black text-slate-400 uppercase tracking-widest">Total Toko Lama</p>
              <p class="text-lg sm:text-2xl font-black text-slate-300 mt-0.5 sm:mt-1">{{ rekapSummary.reduce((s, e) => s + e.tokoLama, 0) }}</p>
            </div>
            <div>
              <p class="text-[9px] sm:text-[10px] font-black text-slate-400 uppercase tracking-widest">Total Banner/Spanduk</p>
              <p class="text-lg sm:text-2xl font-black text-violet-400 mt-0.5 sm:mt-1">{{ rekapSummary.reduce((s, e) => s + e.spandukDibuat, 0) }}</p>
            </div>
            <div class="border-t sm:border-t-0 sm:border-l border-slate-800 pt-3 sm:pt-0 sm:pl-4">
              <p class="text-[9px] sm:text-[10px] font-black text-amber-400 uppercase tracking-widest">Omset TO</p>
              <p class="text-base sm:text-xl font-black text-amber-300 mt-0.5 sm:mt-1">Rp {{ formatRupiah(visitStats.nominalTO) }}</p>
            </div>
            <div class="border-t sm:border-t-0 sm:border-l border-slate-800 pt-3 sm:pt-0 sm:pl-4">
              <p class="text-[9px] sm:text-[10px] font-black text-blue-400 uppercase tracking-widest">Omset Motoris</p>
              <p class="text-base sm:text-xl font-black text-blue-300 mt-0.5 sm:mt-1">Rp {{ formatRupiah(visitStats.nominalMotoris) }}</p>
            </div>
            <div class="border-t sm:border-t-0 sm:border-l border-slate-800 pt-3 sm:pt-0 sm:pl-4">
              <p class="text-[9px] sm:text-[10px] font-black text-purple-400 uppercase tracking-widest">Omset Booster</p>
              <p class="text-base sm:text-xl font-black text-purple-300 mt-0.5 sm:mt-1">Rp {{ formatRupiah(visitStats.nominalBooster) }}</p>
            </div>
            <div class="border-t sm:border-t-0 sm:border-l border-slate-800 pt-3 sm:pt-0 sm:pl-4">
              <p class="text-[9px] sm:text-[10px] font-black text-slate-400 uppercase tracking-widest">Omset MT / Lainnya</p>
              <p class="text-base sm:text-xl font-black text-slate-300 mt-0.5 sm:mt-1">Rp {{ formatRupiah(visitStats.nominalOther) }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- ─── Perjalanan & BBM (Journey) Tab ────────────────────────── -->
      <div v-if="activeTab === 'journey'" class="flex-1 flex flex-col min-h-0">
        <!-- Top Toolbar for Journey Tab -->
        <div class="p-4 sm:p-5 border-b border-slate-100 flex flex-wrap items-center justify-between gap-3 bg-slate-50/50">
          <div>
            <h3 class="text-sm sm:text-base font-black text-slate-900">Data Perjalanan & Odometer Sales</h3>
            <p class="text-xs text-slate-400 font-medium">Kelola pencatatan KM dan klaim BBM salesman harian</p>
          </div>
          <button
            @click="openCreateJourneyModal"
            class="px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground font-black text-xs rounded-xl shadow-lg shadow-primary/20 flex items-center gap-2 transition-all cursor-pointer"
          >
            <Plus class="w-4 h-4" /> Tambah Perjalanan Manual
          </button>
        </div>

        <!-- Loading State -->
        <div v-if="isLoadingJourneys" class="flex-1 flex flex-col items-center justify-center gap-3 py-20">
          <div class="w-10 h-10 border-4 border-slate-100 border-t-primary rounded-full animate-spin"></div>
          <p class="text-sm font-bold text-slate-400">Memuat data odometer dan BBM sales...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="journeys.length === 0" class="flex-1 flex flex-col items-center justify-center gap-3 py-20">
          <Gauge class="w-12 h-12 text-slate-200" />
          <p class="font-bold text-slate-400">Belum ada data pencatatan KM perjalanan.</p>
          <p class="text-xs text-slate-300">Sales akan mencatat KM berangkat dan pulang melalui aplikasi mobile atau klik tombol di atas untuk bypass manual.</p>
          <button
            @click="openCreateJourneyModal"
            class="mt-2 px-4 py-2 bg-primary text-white font-bold text-xs rounded-xl flex items-center gap-1.5"
          >
            <Plus class="w-4 h-4" /> Catat Perjalanan Manual
          </button>
        </div>

        <!-- Journeys Table -->
        <div v-else class="flex-1 overflow-auto no-scrollbar">
          <table class="w-full min-w-[1050px] border-collapse text-left">
            <thead class="sticky top-0 bg-white/90 backdrop-blur-md z-10">
              <tr class="border-b border-slate-100">
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest">No</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest">Salesman & Cabang</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest">Tanggal</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest">KM Berangkat</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest">KM Pulang</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest text-center">Total Jarak</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest text-center">Toko Dikunjungi</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest text-right">Klaim BBM</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest text-center">Status</th>
                <th class="p-4 sm:p-5 text-[11px] font-black text-slate-400 uppercase tracking-widest text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="(j, idx) in journeys"
                :key="j.id"
                class="hover:bg-slate-50/50 transition-colors"
              >
                <!-- No -->
                <td class="p-4 sm:p-5 text-xs font-bold text-slate-400">{{ idx + 1 }}</td>

                <!-- Salesman -->
                <td class="p-4 sm:p-5">
                  <div class="flex items-center gap-3">
                    <div class="w-9 h-9 rounded-xl bg-blue-50 flex items-center justify-center text-blue-600 font-black text-xs shrink-0">
                      {{ (j.employee?.first_name?.[0] || 'S').toUpperCase() }}
                    </div>
                    <div>
                      <p class="font-bold text-slate-900 text-sm">
                        {{ j.employee?.first_name }} {{ j.employee?.last_name }}
                      </p>
                      <p class="text-[11px] font-medium text-slate-400">
                        {{ j.employee?.branch?.name || 'Pusat' }} • {{ j.employee?.job_position?.title || 'Sales' }}
                      </p>
                    </div>
                  </div>
                </td>

                <!-- Tanggal -->
                <td class="p-4 sm:p-5">
                  <span class="font-bold text-slate-700 text-xs">{{ formatDate(j.date) }}</span>
                </td>

                <!-- KM Berangkat -->
                <td class="p-4 sm:p-5">
                  <div class="flex items-start gap-2.5">
                    <div
                      v-if="j.start_photo_url"
                      @click="openImageModal(j.start_photo_url, `KM Berangkat - ${j.employee?.first_name}`, `Odometer: ${j.start_km} KM • ${formatFullTime(j.start_time)}`)"
                      class="w-12 h-12 rounded-xl bg-slate-100 overflow-hidden shrink-0 cursor-pointer border border-slate-200 hover:opacity-80 transition-opacity relative group"
                    >
                      <img :src="j.start_photo_url" class="w-full h-full object-cover" alt="KM Berangkat" />
                      <div class="absolute inset-0 bg-black/30 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                        <Camera class="w-4 h-4 text-white" />
                      </div>
                    </div>
                    <div class="min-w-0">
                      <p class="font-black text-slate-900 text-sm tracking-tight">{{ j.start_km.toLocaleString('id-ID') }} <span class="text-[10px] font-bold text-slate-400">KM</span></p>
                      <p class="text-[10px] font-bold text-blue-600 flex items-center gap-1 mt-0.5">
                        <Clock class="w-3 h-3" /> {{ formatTime(j.start_time) }}
                      </p>
                      <p v-if="j.start_location" class="text-[9px] text-slate-400 truncate max-w-[140px]" :title="j.start_location">
                        {{ j.start_location }}
                      </p>
                    </div>
                  </div>
                </td>

                <!-- KM Pulang -->
                <td class="p-4 sm:p-5">
                  <div v-if="j.status === 'ENDED'" class="flex items-start gap-2.5">
                    <div
                      v-if="j.end_photo_url"
                      @click="openImageModal(j.end_photo_url, `KM Pulang - ${j.employee?.first_name}`, `Odometer: ${j.end_km} KM • ${j.end_time ? formatFullTime(j.end_time) : ''}`)"
                      class="w-12 h-12 rounded-xl bg-slate-100 overflow-hidden shrink-0 cursor-pointer border border-slate-200 hover:opacity-80 transition-opacity relative group"
                    >
                      <img :src="j.end_photo_url" class="w-full h-full object-cover" alt="KM Pulang" />
                      <div class="absolute inset-0 bg-black/30 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                        <Camera class="w-4 h-4 text-white" />
                      </div>
                    </div>
                    <div class="min-w-0">
                      <p class="font-black text-slate-900 text-sm tracking-tight">{{ j.end_km.toLocaleString('id-ID') }} <span class="text-[10px] font-bold text-slate-400">KM</span></p>
                      <p v-if="j.end_time" class="text-[10px] font-bold text-emerald-600 flex items-center gap-1 mt-0.5">
                        <Clock class="w-3 h-3" /> {{ formatTime(j.end_time) }}
                      </p>
                      <p v-if="j.end_location" class="text-[9px] text-slate-400 truncate max-w-[140px]" :title="j.end_location">
                        {{ j.end_location }}
                      </p>
                    </div>
                  </div>
                  <span v-else class="inline-flex px-2.5 py-1 rounded-full bg-amber-50 text-amber-600 text-[10px] font-bold border border-amber-200">
                    Belum Pulang
                  </span>
                </td>

                <!-- Total Jarak -->
                <td class="p-4 sm:p-5 text-center">
                  <div v-if="j.status === 'ENDED'">
                    <span class="inline-flex items-center gap-1 px-3 py-1.5 rounded-xl bg-blue-50 text-blue-700 font-black text-xs border border-blue-100">
                      <Gauge class="w-3.5 h-3.5 text-blue-600" />
                      {{ j.total_km.toLocaleString('id-ID', { maximumFractionDigits: 1 }) }} KM
                    </span>
                  </div>
                  <span v-else class="text-xs font-bold text-slate-300">-</span>
                </td>

                <!-- Toko Dikunjungi -->
                <td class="p-4 sm:p-5 text-center">
                  <span class="font-black text-slate-800 text-sm">{{ j.total_visits || 0 }}</span>
                  <span class="text-[10px] font-bold text-slate-400 ml-1">toko</span>
                </td>

                <!-- Klaim BBM -->
                <td class="p-4 sm:p-5 text-right">
                  <div v-if="j.total_bbm_amount > 0">
                    <p class="font-black text-emerald-600 text-sm">
                      Rp {{ new Intl.NumberFormat('id-ID').format(j.total_bbm_amount) }}
                    </p>
                    <div class="flex items-center justify-end gap-1.5 mt-1">
                      <template v-for="bbm in j.bbm_reimbursements" :key="bbm.id">
                        <div class="flex items-center gap-0.5">
                          <button
                            v-if="bbm.attachment_url"
                            @click="openImageModal(bbm.attachment_url, `Struk BBM - ${j.employee?.first_name}`, `Rp ${new Intl.NumberFormat('id-ID').format(bbm.amount)} (${bbm.status})`)"
                            class="p-1 rounded-md bg-amber-50 text-amber-700 hover:bg-amber-100 text-[9px] font-bold flex items-center gap-0.5 border border-amber-200 cursor-pointer"
                            title="Lihat Struk BBM"
                          >
                            <Fuel class="w-3 h-3" /> Struk
                          </button>
                          <button
                            @click="deleteBBM(bbm, j)"
                            class="p-1 rounded-md bg-rose-50 text-rose-600 hover:bg-rose-100 text-[9px] font-bold border border-rose-200 cursor-pointer"
                            title="Hapus Klaim BBM"
                          >
                            <Trash2 class="w-3 h-3" />
                          </button>
                        </div>
                      </template>
                    </div>
                  </div>
                  <div v-else class="flex items-center justify-end">
                    <span class="text-xs font-bold text-slate-300">Rp 0</span>
                  </div>
                </td>

                <!-- Status -->
                <td class="p-4 sm:p-5 text-center">
                  <span
                    v-if="j.status === 'ENDED'"
                    class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-emerald-100 text-emerald-700 text-[10px] font-black uppercase border border-emerald-200"
                  >
                    ✓ SELESAI
                  </span>
                  <span
                    v-else
                    class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-blue-100 text-blue-700 text-[10px] font-black uppercase border border-blue-200 animate-pulse"
                  >
                    • DI LAPANGAN
                  </span>
                </td>

                <!-- Aksi -->
                <td class="p-4 sm:p-5 text-center">
                  <div class="flex items-center justify-center gap-1.5">
                    <button
                      @click="openAddBBMModal(j)"
                      class="p-1.5 rounded-lg bg-amber-50 hover:bg-amber-100 text-amber-700 transition-colors border border-amber-200 cursor-pointer"
                      title="Tambah Klaim BBM"
                    >
                      <Fuel class="w-4 h-4" />
                    </button>
                    <button
                      @click="openEditJourneyModal(j)"
                      class="p-1.5 rounded-lg bg-blue-50 hover:bg-blue-100 text-blue-700 transition-colors border border-blue-200 cursor-pointer"
                      title="Edit Perjalanan"
                    >
                      <Edit class="w-4 h-4" />
                    </button>
                    <button
                      @click="deleteJourney(j)"
                      class="p-1.5 rounded-lg bg-rose-50 hover:bg-rose-100 text-rose-600 transition-colors border border-rose-200 cursor-pointer"
                      title="Hapus Catatan Perjalanan"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Modal Catat / Edit Perjalanan Manual -->
    <div v-if="showJourneyModal" class="fixed inset-0 z-[100] flex items-center justify-center p-3 sm:p-6 animate-in fade-in duration-200">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showJourneyModal = false"></div>
      <div class="relative w-full max-w-lg bg-white rounded-3xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh] z-10">
        <!-- Modal Header -->
        <div class="p-5 sm:p-6 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-blue-50 flex items-center justify-center text-blue-600">
              <Gauge class="w-5 h-5" />
            </div>
            <div>
              <h3 class="font-black text-slate-900 text-base">
                {{ isEditingJourney ? 'Edit Perjalanan Sales' : 'Tambah Perjalanan Manual (Bypass)' }}
              </h3>
              <p class="text-xs text-slate-400 font-medium">
                {{ isEditingJourney ? 'Perbarui data KM & status perjalanan' : 'Input KM awal agar salesman bisa langsung absen toko' }}
              </p>
            </div>
          </div>
          <button @click="showJourneyModal = false" class="p-2 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-all cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-5 sm:p-6 overflow-y-auto space-y-4 no-scrollbar">
          <!-- Salesman Selector (Only on Create) -->
          <div v-if="!isEditingJourney">
            <label class="block text-xs font-black text-slate-600 uppercase tracking-wider mb-1.5">Salesman *</label>
            <select
              v-model="journeyForm.employee_id"
              class="w-full h-11 px-3 rounded-xl border border-slate-200 bg-slate-50 focus:bg-white text-sm font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all"
            >
              <option value="" disabled>Pilih Salesman</option>
              <option v-for="s in salesmen" :key="s.id" :value="s.id">
                {{ s.first_name }} {{ s.last_name || '' }} ({{ s.branch?.name || 'Pusat' }})
              </option>
            </select>
          </div>

          <!-- Tanggal & Status -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-black text-slate-600 uppercase tracking-wider mb-1.5">Tanggal *</label>
              <input
                type="date"
                v-model="journeyForm.date"
                class="w-full h-11 px-3 rounded-xl border border-slate-200 bg-slate-50 focus:bg-white text-sm font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all"
              />
            </div>
            <div>
              <label class="block text-xs font-black text-slate-600 uppercase tracking-wider mb-1.5">Status Perjalanan *</label>
              <select
                v-model="journeyForm.status"
                class="w-full h-11 px-3 rounded-xl border border-slate-200 bg-slate-50 focus:bg-white text-sm font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all"
              >
                <option value="STARTED">STARTED (Di Lapangan)</option>
                <option value="ENDED">ENDED (Selesai Pulang)</option>
              </select>
            </div>
          </div>

          <!-- Section: KM Berangkat -->
          <div class="p-4 rounded-2xl bg-blue-50/50 border border-blue-100 space-y-3">
            <div class="flex items-center gap-2 text-blue-700">
              <Gauge class="w-4 h-4" />
              <span class="text-xs font-black uppercase tracking-wider">KM Berangkat (Awal)</span>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-[11px] font-bold text-slate-500 mb-1">Angka Odometer (KM) *</label>
                <input
                  type="number"
                  step="0.1"
                  v-model.number="journeyForm.start_km"
                  placeholder="Contoh: 12500"
                  class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-sm font-black text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
                />
              </div>
              <div>
                <label class="block text-[11px] font-bold text-slate-500 mb-1">Jam Berangkat</label>
                <input
                  type="text"
                  v-model="journeyForm.start_time"
                  placeholder="08:00"
                  class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-sm font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
                />
              </div>
            </div>
            <div>
              <label class="block text-[11px] font-bold text-slate-500 mb-1">Lokasi Berangkat (Opsional)</label>
              <input
                type="text"
                v-model="journeyForm.start_location"
                placeholder="Contoh: Kantor Cabang / Rumah"
                class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-sm font-medium text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
              />
            </div>
            <div>
              <label class="block text-[11px] font-bold text-slate-500 mb-1">URL Foto Speedometer Berangkat (Opsional)</label>
              <input
                type="text"
                v-model="journeyForm.start_photo_url"
                placeholder="https://..."
                class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-xs font-mono text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500"
              />
            </div>
          </div>

          <!-- Section: KM Pulang (Conditional / Editable) -->
          <div v-if="journeyForm.status === 'ENDED'" class="p-4 rounded-2xl bg-emerald-50/50 border border-emerald-100 space-y-3">
            <div class="flex items-center gap-2 text-emerald-700">
              <CheckCircle class="w-4 h-4" />
              <span class="text-xs font-black uppercase tracking-wider">KM Pulang (Akhir)</span>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-[11px] font-bold text-slate-500 mb-1">Angka Odometer Pulang (KM) *</label>
                <input
                  type="number"
                  step="0.1"
                  v-model.number="journeyForm.end_km"
                  placeholder="Contoh: 12560"
                  class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-sm font-black text-slate-800 focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
                />
              </div>
              <div>
                <label class="block text-[11px] font-bold text-slate-500 mb-1">Jam Pulang</label>
                <input
                  type="text"
                  v-model="journeyForm.end_time"
                  placeholder="17:00"
                  class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-sm font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
                />
              </div>
            </div>
            <div>
              <label class="block text-[11px] font-bold text-slate-500 mb-1">Lokasi Pulang (Opsional)</label>
              <input
                type="text"
                v-model="journeyForm.end_location"
                placeholder="Contoh: Kantor Cabang / Rumah"
                class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-sm font-medium text-slate-800 focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
              />
            </div>
            <div>
              <label class="block text-[11px] font-bold text-slate-500 mb-1">URL Foto Speedometer Pulang (Opsional)</label>
              <input
                type="text"
                v-model="journeyForm.end_photo_url"
                placeholder="https://..."
                class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-xs font-mono text-slate-800 focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
              />
            </div>
          </div>

          <!-- Section: Klaim BBM Awal (Only on Create) -->
          <div v-if="!isEditingJourney" class="p-4 rounded-2xl bg-amber-50/50 border border-amber-100 space-y-3">
            <div class="flex items-center gap-2 text-amber-700">
              <Fuel class="w-4 h-4" />
              <span class="text-xs font-black uppercase tracking-wider">Klaim BBM (Opsional)</span>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-[11px] font-bold text-slate-500 mb-1">Nominal Beli BBM (Rp)</label>
                <input
                  type="number"
                  v-model.number="journeyForm.bbm_amount"
                  placeholder="Contoh: 30000"
                  class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-sm font-black text-slate-800 focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500"
                />
              </div>
              <div>
                <label class="block text-[11px] font-bold text-slate-500 mb-1">URL Struk BBM (Opsional)</label>
                <input
                  type="text"
                  v-model="journeyForm.bbm_proof_url"
                  placeholder="https://..."
                  class="w-full h-10 px-3 rounded-xl border border-slate-200 bg-white text-xs font-mono text-slate-800 focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 sm:p-5 border-t border-slate-100 flex items-center justify-end gap-2.5 bg-slate-50/50">
          <button
            type="button"
            @click="showJourneyModal = false"
            class="px-4 py-2 rounded-xl text-slate-600 hover:bg-slate-200 font-bold text-xs transition-colors cursor-pointer"
          >
            Batal
          </button>
          <button
            type="button"
            @click="submitJourneyForm"
            :disabled="isSavingJourney"
            class="px-5 py-2.5 rounded-xl bg-primary text-white font-black text-xs hover:bg-primary/90 transition-all flex items-center gap-2 shadow-lg shadow-primary/20 disabled:opacity-50 cursor-pointer"
          >
            <div v-if="isSavingJourney" class="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            {{ isEditingJourney ? 'Simpan Perubahan' : 'Catat Perjalanan' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Tambah Klaim BBM -->
    <div v-if="showBBMModal" class="fixed inset-0 z-[100] flex items-center justify-center p-3 sm:p-6 animate-in fade-in duration-200">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showBBMModal = false"></div>
      <div class="relative w-full max-w-md bg-white rounded-3xl shadow-2xl overflow-hidden flex flex-col z-10">
        <!-- Header -->
        <div class="p-5 sm:p-6 border-b border-slate-100 flex items-center justify-between bg-amber-50/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-amber-100 flex items-center justify-center text-amber-700">
              <Fuel class="w-5 h-5" />
            </div>
            <div>
              <h3 class="font-black text-slate-900 text-base">Tambah Klaim BBM</h3>
              <p class="text-xs text-slate-500 font-medium">
                {{ selectedJourneyForBBM?.employee?.first_name }} • {{ formatDate(selectedJourneyForBBM?.date) }}
              </p>
            </div>
          </div>
          <button @click="showBBMModal = false" class="p-2 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-all cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-5 sm:p-6 space-y-4">
          <div>
            <label class="block text-xs font-black text-slate-600 uppercase tracking-wider mb-1.5">Nominal Klaim (Rp) *</label>
            <input
              type="number"
              v-model.number="bbmForm.amount"
              placeholder="Contoh: 35000"
              class="w-full h-11 px-3 rounded-xl border border-slate-200 bg-slate-50 focus:bg-white text-base font-black text-slate-800 focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 transition-all"
            />
          </div>
          <div>
            <label class="block text-xs font-black text-slate-600 uppercase tracking-wider mb-1.5">Keterangan / SPBU</label>
            <input
              type="text"
              v-model="bbmForm.description"
              placeholder="Contoh: Isi BBM Pertalite SPBU Jogja"
              class="w-full h-11 px-3 rounded-xl border border-slate-200 bg-slate-50 focus:bg-white text-sm font-medium text-slate-800 focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 transition-all"
            />
          </div>
          <div>
            <label class="block text-xs font-black text-slate-600 uppercase tracking-wider mb-1.5">URL Foto Struk BBM (Opsional)</label>
            <input
              type="text"
              v-model="bbmForm.attachment_url"
              placeholder="https://..."
              class="w-full h-11 px-3 rounded-xl border border-slate-200 bg-slate-50 focus:bg-white text-xs font-mono text-slate-800 focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 transition-all"
            />
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 sm:p-5 border-t border-slate-100 flex items-center justify-end gap-2.5 bg-slate-50/50">
          <button
            type="button"
            @click="showBBMModal = false"
            class="px-4 py-2 rounded-xl text-slate-600 hover:bg-slate-200 font-bold text-xs transition-colors cursor-pointer"
          >
            Batal
          </button>
          <button
            type="button"
            @click="submitBBMForm"
            :disabled="isSavingBBM"
            class="px-5 py-2.5 rounded-xl bg-amber-600 text-white font-black text-xs hover:bg-amber-700 transition-all flex items-center gap-2 shadow-lg shadow-amber-600/20 disabled:opacity-50 cursor-pointer"
          >
            <div v-if="isSavingBBM" class="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            Simpan Klaim BBM
          </button>
        </div>
      </div>
    </div>

    <!-- MERCHANDISING MT TAB CONTENT -->
    <div v-if="activeTab === 'merchandising'" class="space-y-6">
      <!-- Merchandising Metrics -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-rose-50 text-rose-700 flex items-center justify-center font-bold">
            <Layers class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Total Kunjungan Audit</p>
            <p class="text-2xl font-black text-slate-900">{{ totalMDLaporan }} <span class="text-xs font-medium text-slate-400">Laporan</span></p>
          </div>
        </div>

        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold">
            <CheckCircle2 class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Patuh Planogram Rak</p>
            <p class="text-2xl font-black text-emerald-600">{{ totalMDCompliant }} <span class="text-xs font-medium text-slate-400">Audit Sesuai</span></p>
          </div>
        </div>

        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
            <AlertCircle class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-slate-400">Ada Out of Stock (OOS)</p>
            <p class="text-2xl font-black text-amber-600">{{ totalMDOOS }} <span class="text-xs font-medium text-slate-400">Outlet Perlu Restock</span></p>
          </div>
        </div>
      </div>

      <!-- Controls Row -->
      <div class="flex flex-col md:flex-row justify-between items-stretch md:items-center gap-4">
        <div class="flex flex-1 items-center gap-3">
          <div class="relative flex-1 max-w-md">
            <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              v-model="mdSearchQuery"
              type="text"
              placeholder="Cari toko, supermarket, atau nama MD..."
              class="w-full pl-11 pr-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs sm:text-sm font-medium focus:ring-2 focus:ring-rose-700/20 focus:border-rose-700 outline-none"
            />
          </div>

          <select
            v-model="selectedMDStatus"
            class="px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold text-slate-700 focus:ring-2 focus:ring-rose-700/20 outline-none"
          >
            <option value="ALL">Semua Kepatuhan</option>
            <option value="COMPLIANT">Patuh Planogram</option>
            <option value="NON_COMPLIANT">Tidak Patuh Planogram</option>
          </select>
        </div>

        <button
          @click="openNewMDModal"
          class="flex items-center justify-center gap-2 bg-rose-700 hover:bg-rose-800 text-white px-5 py-2.5 rounded-xl font-bold text-xs shadow-md shadow-rose-200 transition-all cursor-pointer"
        >
          <Plus class="w-4 h-4" />
          Tambah Audit Merchandising
        </button>
      </div>

      <!-- Table Merchandising -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div v-if="isLoadingMD" class="flex items-center justify-center py-20">
          <div class="w-8 h-8 border-4 border-rose-700 border-t-transparent rounded-full animate-spin"></div>
        </div>

        <div v-else-if="filteredMDSubmissions.length === 0" class="flex flex-col items-center justify-center py-16 text-slate-400">
          <Layers class="w-16 h-16 mb-3 opacity-20 text-rose-700" />
          <p class="font-bold text-base text-slate-700">Belum ada data audit merchandising</p>
          <p class="text-xs text-slate-400">Klik Tambah Audit Merchandising untuk mencatat audit planogram rak & stok kosong</p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-left">
            <thead>
              <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
                <th class="py-4 px-6">Toko & MD</th>
                <th class="py-4 px-6">Waktu Kunjungan</th>
                <th class="py-4 px-6 text-center">Performa & Planogram</th>
                <th class="py-4 px-6">Status Stok Rak (OOS)</th>
                <th class="py-4 px-6 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50 text-slate-700 text-sm">
              <tr v-for="sub in filteredMDSubmissions" :key="sub.id" class="hover:bg-slate-50/40 transition-colors">
                <td class="py-4 px-6">
                  <p class="font-bold text-sm text-slate-900">{{ sub.supermarket_name || sub.store_name }}</p>
                  <p class="text-xs text-rose-700 font-bold">MD: {{ sub.md_name || sub.merchandiser_name }}</p>
                </td>

                <td class="py-4 px-6 text-xs font-medium text-slate-500">
                  <div class="flex items-center gap-1.5">
                    <Calendar class="w-3.5 h-3.5 text-slate-400" />
                    <span>{{ sub.visit_time }}</span>
                  </div>
                </td>

                <td class="py-4 px-6 text-center">
                  <span
                    class="px-2.5 py-1 text-[10px] font-black uppercase rounded-full border"
                    :class="sub.planogram_compliant ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-rose-50 text-rose-700 border-rose-200'"
                  >
                    {{ sub.planogram_compliant ? 'Planogram OK' : 'Mismatch' }}
                  </span>
                  <p v-if="sub.shelf_share_percent" class="text-[10px] text-slate-400 font-bold mt-1">
                    Shelf Share: {{ sub.shelf_share_percent }}%
                  </p>
                </td>

                <td class="py-4 px-6">
                  <div v-if="sub.out_of_stock_skus && sub.out_of_stock_skus.length > 0" class="flex flex-wrap gap-1">
                    <span
                      v-for="sku in sub.out_of_stock_skus"
                      :key="sku"
                      class="px-2 py-0.5 text-[10px] font-bold text-rose-700 bg-rose-50 border border-rose-100 rounded-md"
                    >
                      Kosong: {{ sku }}
                    </span>
                  </div>
                  <span v-else class="text-xs text-emerald-600 font-bold flex items-center gap-1">
                    <CheckCircle2 class="w-3.5 h-3.5" /> Semua SKU Ready
                  </span>

                  <div class="flex gap-2 mt-1.5" v-if="sub.photo_url || sub.invoice_photo_url || sub.receipt_photo_url">
                    <button v-if="sub.photo_url" @click="openPreview(sub.photo_url)" class="text-[10px] text-rose-700 hover:underline flex items-center gap-0.5 font-bold cursor-pointer">
                      <Camera class="w-3 h-3"/> Foto Display
                    </button>
                    <button v-if="sub.invoice_photo_url" @click="openPreview(sub.invoice_photo_url)" class="text-[10px] text-rose-700 hover:underline flex items-center gap-0.5 font-bold cursor-pointer">
                      <ImageIcon class="w-3 h-3"/> Nota
                    </button>
                    <button v-if="sub.receipt_photo_url" @click="openPreview(sub.receipt_photo_url)" class="text-[10px] text-rose-700 hover:underline flex items-center gap-0.5 font-bold cursor-pointer">
                      <ImageIcon class="w-3 h-3"/> Tanda Terima
                    </button>
                  </div>
                </td>

                <td class="py-4 px-6 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <button
                      @click="createRepeatOrder(sub)"
                      class="px-2 py-1 text-[11px] font-bold text-emerald-700 bg-emerald-50 hover:bg-emerald-100 border border-emerald-200 rounded-lg transition-all flex items-center gap-1 cursor-pointer"
                      title="Buat Repeat Order PO"
                    >
                      <Plus class="w-3 h-3" /> PO
                    </button>
                    <button
                      @click="openEditMDModal(sub)"
                      class="p-1.5 text-slate-400 hover:text-rose-700 hover:bg-rose-50 rounded-lg transition-all cursor-pointer"
                      title="Edit Laporan"
                    >
                      <Edit class="w-3.5 h-3.5" />
                    </button>
                    <button
                      @click="deleteMD(sub.id)"
                      class="p-1.5 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-lg transition-all cursor-pointer"
                      title="Hapus Laporan"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- MODAL FORM AUDIT MERCHANDISING -->
    <div v-if="showMDModal" class="fixed inset-0 z-[100] flex items-center justify-center p-3 sm:p-6">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showMDModal = false"></div>
      <div class="relative w-full max-w-xl max-h-[90vh] bg-white rounded-3xl shadow-2xl overflow-hidden flex flex-col">
        <div class="p-5 sm:p-6 border-b border-slate-100 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-rose-50 text-rose-700 flex items-center justify-center font-bold">
              <Layers class="w-5 h-5" />
            </div>
            <div>
              <h2 class="text-lg font-black text-slate-900">
                {{ isEditingMD ? 'Edit Audit Merchandising' : 'Input Audit Merchandising MT' }}
              </h2>
              <p class="text-xs text-slate-500 font-medium">Audit planogram rak, OOS SKU, dan foto display supermarket</p>
            </div>
          </div>
          <button @click="showMDModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
            <X class="w-5 h-5" />
          </button>
        </div>

        <form @submit.prevent="saveMD" class="p-6 space-y-4 overflow-y-auto flex-1">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1">Key Account MT *</label>
              <select
                v-model="mdForm.key_account_id"
                required
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 focus:ring-2 focus:ring-rose-700/20 outline-none"
              >
                <option v-for="ka in keyAccounts" :key="ka.id" :value="ka.id">{{ ka.name }}</option>
              </select>
            </div>

            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1">Nama Petugas MD / Sales *</label>
              <input
                v-model="mdForm.merchandiser_name"
                type="text"
                required
                placeholder="Nama MD..."
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium focus:ring-2 focus:ring-rose-700/20 outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1">Nama Toko / Outlet MT *</label>
              <input
                v-model="mdForm.store_name"
                type="text"
                required
                placeholder="Contoh: Superindo Dago..."
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium focus:ring-2 focus:ring-rose-700/20 outline-none"
              />
            </div>

            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1">Waktu Kunjungan</label>
              <input
                v-model="mdForm.visit_time"
                type="datetime-local"
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium focus:ring-2 focus:ring-rose-700/20 outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1">Kepatuhan Planogram</label>
              <select
                v-model="mdForm.planogram_compliant"
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 focus:ring-2 focus:ring-rose-700/20 outline-none"
              >
                <option :value="true">Patuh Planogram (OK)</option>
                <option :value="false">Tidak Patuh (Mismatch)</option>
              </select>
            </div>

            <div>
              <label class="block text-xs font-black uppercase text-slate-500 mb-1">Shelf Share (%)</label>
              <input
                v-model.number="mdForm.shelf_share_percent"
                type="number"
                min="0"
                max="100"
                placeholder="Contoh: 35"
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold focus:ring-2 focus:ring-rose-700/20 outline-none"
              />
            </div>
          </div>

          <div>
            <label class="block text-xs font-black uppercase text-slate-500 mb-1">SKU yang Kosong (OOS) - Pisahkan koma</label>
            <input
              v-model="mdForm.out_of_stock_skus"
              type="text"
              placeholder="Contoh: Karsa HR Coklat 250ml, Karsa HR Strawberi 250ml"
              class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium focus:ring-2 focus:ring-rose-700/20 outline-none"
            />
          </div>

          <!-- Upload Foto Display Rak -->
          <div>
            <label class="block text-xs font-black uppercase text-slate-500 mb-1">Foto Display Rak Supermarket</label>
            <div class="flex items-center gap-3">
              <div v-if="mdForm.photo_url" class="w-12 h-12 rounded-xl border border-slate-200 overflow-hidden shrink-0">
                <img :src="mdForm.photo_url" class="w-full h-full object-cover" />
              </div>
              <label class="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 border-2 border-dashed border-slate-200 hover:border-rose-700 rounded-xl cursor-pointer bg-slate-50 text-xs font-bold text-slate-600">
                <Camera class="w-4 h-4 text-rose-700" />
                <span>{{ isUploadingMDPhoto ? 'Mengunggah...' : 'Unggah Foto Rak Supermarket' }}</span>
                <input type="file" accept="image/*" class="hidden" @change="uploadMDPhoto($event, 'photo_url')" :disabled="isUploadingMDPhoto" />
              </label>
            </div>
          </div>

          <div class="pt-3 flex gap-3">
            <button
              type="button"
              @click="showMDModal = false"
              class="flex-1 px-4 py-2.5 text-xs font-bold text-slate-700 bg-slate-100 rounded-xl hover:bg-slate-200 transition-colors"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isSavingMD"
              class="flex-1 px-4 py-2.5 text-xs font-bold text-white bg-rose-700 rounded-xl hover:bg-rose-800 transition-colors disabled:opacity-50"
            >
              {{ isSavingMD ? 'Menyimpan...' : 'Simpan Audit' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- PREVIEW IMAGE MODAL -->
    <div v-if="previewImageUrl" class="fixed inset-0 z-[120] flex items-center justify-center p-4 bg-black/80" @click="closePreview">
      <div class="relative max-w-xl max-h-[85vh] bg-white p-2 rounded-2xl overflow-hidden" @click.stop>
        <img :src="previewImageUrl" class="w-full h-auto max-h-[75vh] object-contain rounded-xl" />
        <div class="pt-3 text-right">
          <button @click="closePreview" class="px-4 py-2 bg-slate-100 text-slate-700 font-bold text-xs rounded-xl hover:bg-slate-200 cursor-pointer">
            Tutup
          </button>
        </div>
      </div>
    </div>

    <!-- Form Modal (Single & Batch Dual-Mode) -->
    <div v-if="showFormModal" class="fixed inset-0 z-[100] flex items-center justify-center p-3 sm:p-6">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showFormModal = false"
      ></div>
      <div
        class="relative w-full max-w-2xl max-h-[92vh] bg-white rounded-3xl sm:rounded-[36px] shadow-2xl overflow-hidden flex flex-col animate-in fade-in zoom-in duration-300 z-10"
      >
        <!-- Modal Header -->
        <div class="p-5 sm:p-7 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div>
            <div class="flex items-center gap-2">
              <div class="w-8 h-8 rounded-xl bg-primary/10 text-primary flex items-center justify-center font-bold">
                <Store class="w-4 h-4" />
              </div>
              <h2 class="text-base sm:text-xl font-black text-slate-900 tracking-tight">
                {{ isEditing ? 'Edit Kunjungan Sales' : 'Input Kunjungan Sales Manual' }}
              </h2>
            </div>
            <p class="text-slate-500 font-medium text-xs mt-1">
              {{ isEditing ? 'Perbarui data rincian kunjungan sales' : 'Pilih mode input satuan cepat atau sekaligus banyak toko (batch)' }}
            </p>
          </div>
          <button
            @click="showFormModal = false"
            class="p-2 sm:p-2.5 bg-slate-100 hover:bg-slate-200 rounded-xl sm:rounded-2xl transition-all text-slate-500 cursor-pointer"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Mode Tabs (Single vs Batch) -->
        <div v-if="!isEditing" class="px-5 sm:px-7 pt-4 pb-2 bg-slate-50/30 flex items-center gap-2 border-b border-slate-100">
          <button
            type="button"
            @click="manualEntryMode = 'single'"
            :class="[
              'px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 cursor-pointer',
              manualEntryMode === 'single'
                ? 'bg-primary text-white shadow-md shadow-primary/20'
                : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'
            ]"
          >
            <Zap class="w-3.5 h-3.5" />
            <span>Input Satuan Cepat</span>
          </button>
          <button
            type="button"
            @click="manualEntryMode = 'batch'"
            :class="[
              'px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 cursor-pointer',
              manualEntryMode === 'batch'
                ? 'bg-primary text-white shadow-md shadow-primary/20'
                : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'
            ]"
          >
            <Layers class="w-3.5 h-3.5" />
            <span>Input Sekaligus (Batch Toko)</span>
            <span v-if="batchStores.length > 0" class="ml-1 px-1.5 py-0.5 rounded-full text-[10px] bg-white text-primary font-black">
              {{ batchStores.length }}
            </span>
          </button>
        </div>

        <!-- Single Mode Form -->
        <form v-if="manualEntryMode === 'single'" @submit.prevent="handleSaveSingle(false)" class="p-5 sm:p-7 space-y-4 sm:space-y-5 overflow-y-auto no-scrollbar flex-1">
          <!-- Salesman Selector -->
          <div class="space-y-1.5">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
              Salesman <span class="text-rose-500">*</span>
            </label>
            <select
              v-model="form.employee_id"
              required
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
            >
              <option value="" disabled>Pilih Salesman</option>
              <option v-for="emp in salesmen" :key="emp.id" :value="emp.id">
                {{ emp.first_name }} {{ emp.last_name }} — {{ emp.job_position?.title || 'Sales' }} ({{ emp.branch?.name || 'Pusat' }})
              </option>
            </select>
          </div>

          <!-- Store Searchable Combobox -->
          <div class="space-y-1.5 relative" ref="singleStoreDropdownRef">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1 flex items-center justify-between">
              <span>Toko / Customer <span class="text-rose-500">*</span></span>
              <span v-if="selectedStoreSingleObj?.latitude && selectedStoreSingleObj?.longitude" class="text-emerald-600 font-bold flex items-center gap-1 normal-case text-[10px]">
                <MapPin class="w-3 h-3" /> GPS Terhubung ({{ selectedStoreSingleObj.latitude.toFixed(4) }}, {{ selectedStoreSingleObj.longitude.toFixed(4) }})
              </span>
            </label>

            <!-- Selected Store Summary Banner if selected -->
            <div v-if="selectedStoreSingleObj" class="p-3 bg-emerald-50/70 border border-emerald-200 rounded-2xl flex items-center justify-between gap-3">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <p class="font-black text-slate-800 text-xs sm:text-sm truncate">{{ selectedStoreSingleObj.name }}</p>
                  <span v-if="selectedStoreSingleObj.category" class="px-2 py-0.5 rounded-md text-[9px] font-bold bg-emerald-100 text-emerald-800 shrink-0">
                    {{ selectedStoreSingleObj.category }}
                  </span>
                </div>
                <p class="text-[11px] text-slate-500 truncate mt-0.5">{{ selectedStoreSingleObj.address || 'Alamat tidak terdata' }}</p>
              </div>
              <button
                type="button"
                @click="clearSelectedStoreSingle"
                class="px-2.5 py-1 text-[11px] font-bold text-emerald-700 hover:bg-emerald-100 rounded-xl transition-all shrink-0 cursor-pointer"
              >
                Ganti Toko
              </button>
            </div>

            <!-- Search Input Box if not selected or changing -->
            <div v-else class="relative">
              <div class="relative flex items-center">
                <Search class="w-4 h-4 text-slate-400 absolute left-4 pointer-events-none" />
                <input
                  v-model="singleStoreSearch"
                  type="text"
                  placeholder="Ketik nama toko, alamat, atau kode..."
                  @focus="isSingleStoreDropdownOpen = true"
                  class="w-full pl-10 pr-10 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
                />
                <button
                  v-if="singleStoreSearch"
                  type="button"
                  @click="singleStoreSearch = ''"
                  class="absolute right-3 p-1 text-slate-400 hover:text-slate-600 rounded-lg"
                >
                  <X class="w-4 h-4" />
                </button>
              </div>

              <!-- Search Results Dropdown List -->
              <div
                v-if="isSingleStoreDropdownOpen"
                class="absolute left-0 right-0 top-full mt-1.5 bg-white rounded-2xl border border-slate-200 shadow-2xl max-h-56 overflow-y-auto no-scrollbar z-50 divide-y divide-slate-50 animate-in fade-in zoom-in-95 duration-150"
              >
                <div v-if="filteredStoresSingle.length === 0" class="p-4 text-center text-xs text-slate-400 font-medium">
                  Tidak ada toko yang cocok dengan "{{ singleStoreSearch }}"
                </div>
                <button
                  v-for="s in filteredStoresSingle"
                  :key="s.id"
                  type="button"
                  @click="selectStoreSingle(s)"
                  class="w-full px-4 py-2.5 text-left hover:bg-primary/5 flex items-center justify-between gap-2 transition-all cursor-pointer group"
                >
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-1.5">
                      <p class="font-bold text-xs text-slate-800 group-hover:text-primary transition-colors truncate">{{ s.name }}</p>
                      <span v-if="s.category" class="px-1.5 py-0.5 rounded text-[9px] font-bold bg-slate-100 text-slate-600">
                        {{ s.category }}
                      </span>
                    </div>
                    <p class="text-[10px] text-slate-400 truncate">{{ s.address || 'Tanpa alamat' }}</p>
                  </div>
                  <span v-if="s.latitude && s.longitude" class="text-[10px] font-bold text-emerald-600 shrink-0 flex items-center gap-0.5">
                    <MapPin class="w-3 h-3" /> Ada GPS
                  </span>
                </button>
              </div>
            </div>
          </div>

          <!-- Waktu Kunjungan & Tipe -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1 flex items-center justify-between">
                <span>Waktu Kunjungan <span class="text-rose-500">*</span></span>
              </label>
              <input
                v-model="form.check_time"
                type="datetime-local"
                required
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
              />
              <!-- Quick Time Preset Chips -->
              <div class="flex flex-wrap items-center gap-1 pt-1">
                <button
                  type="button"
                  @click="setSingleTimeNow"
                  class="px-2 py-0.5 bg-slate-100 hover:bg-primary hover:text-white text-[10px] font-bold text-slate-600 rounded-lg transition-all cursor-pointer"
                >
                  Sekarang
                </button>
                <button
                  type="button"
                  @click="applySingleTimeOffset(15)"
                  class="px-2 py-0.5 bg-slate-100 hover:bg-primary hover:text-white text-[10px] font-bold text-slate-600 rounded-lg transition-all cursor-pointer"
                >
                  +15 mnt
                </button>
                <button
                  type="button"
                  @click="applySingleTimeOffset(30)"
                  class="px-2 py-0.5 bg-slate-100 hover:bg-primary hover:text-white text-[10px] font-bold text-slate-600 rounded-lg transition-all cursor-pointer"
                >
                  +30 mnt
                </button>
                <button
                  type="button"
                  @click="applySingleTimeOffset(60)"
                  class="px-2 py-0.5 bg-slate-100 hover:bg-primary hover:text-white text-[10px] font-bold text-slate-600 rounded-lg transition-all cursor-pointer"
                >
                  +1 jam
                </button>
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
                Tipe Kunjungan
              </label>
              <select
                v-model="form.type"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
              >
                <option value="CHECKIN">CHECKIN (Masuk)</option>
                <option value="CHECKOUT">CHECKOUT (Keluar)</option>
              </select>
            </div>
          </div>

          <!-- Catatan -->
          <div class="space-y-1.5">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
              Catatan / Feedback Kunjungan
            </label>
            <textarea
              v-model="form.notes"
              rows="2"
              class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none resize-none"
              placeholder="Contoh: Toko sedang ramai, titip display poster, prospek order minggu depan..."
            ></textarea>
          </div>

          <!-- Action Buttons -->
          <div class="pt-2 flex flex-col sm:flex-row items-center gap-2 sm:gap-3">
            <button
              v-if="!isEditing"
              type="button"
              @click="handleSaveSingle(true)"
              :disabled="isSavingSingle"
              class="w-full sm:flex-1 py-3 px-4 rounded-2xl font-bold text-xs sm:text-sm bg-slate-100 hover:bg-slate-200 text-slate-800 transition-all flex items-center justify-center gap-1.5 cursor-pointer disabled:opacity-50"
            >
              <Plus class="w-4 h-4 text-primary" />
              <span>{{ isSavingSingle ? 'Menyimpan...' : 'Simpan & Tambah Toko Lain' }}</span>
            </button>
            <Button
              type="submit"
              :disabled="isSavingSingle"
              class="w-full sm:flex-1 py-3 sm:py-3.5 rounded-2xl font-black text-xs sm:text-sm shadow-lg shadow-primary/20 flex items-center justify-center gap-1.5 cursor-pointer"
            >
              <Check class="w-4 h-4" />
              <span>{{ isSavingSingle ? 'Menyimpan...' : (isEditing ? 'Simpan Perubahan' : 'Simpan Kunjungan') }}</span>
            </Button>
          </div>
        </form>

        <!-- Batch Mode Form -->
        <div v-else class="p-5 sm:p-7 space-y-4 sm:space-y-5 overflow-y-auto no-scrollbar flex-1">
          <!-- Salesman & Tanggal -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
                Salesman <span class="text-rose-500">*</span>
              </label>
              <select
                v-model="batchEmployeeId"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
              >
                <option value="" disabled>Pilih Salesman</option>
                <option v-for="emp in salesmen" :key="emp.id" :value="emp.id">
                  {{ emp.first_name }} {{ emp.last_name }} — ({{ emp.branch?.name || 'Pusat' }})
                </option>
              </select>
            </div>
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
                Tanggal Kunjungan <span class="text-rose-500">*</span>
              </label>
              <input
                v-model="batchDate"
                type="date"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
              />
            </div>
          </div>

          <!-- Jam Mulai & Interval -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 bg-slate-50/60 p-3.5 rounded-2xl border border-slate-100">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">
                Jam Mulai Kunjungan
              </label>
              <input
                v-model="batchStartTime"
                type="time"
                class="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none"
              />
            </div>
            <div class="space-y-1.5">
              <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">
                Estimasi Durasi / Interval Toko
              </label>
              <select
                v-model="batchIntervalMinutes"
                class="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none"
              >
                <option :value="15">15 Menit / Toko</option>
                <option :value="20">20 Menit / Toko</option>
                <option :value="30">30 Menit / Toko (Standar)</option>
                <option :value="45">45 Menit / Toko</option>
                <option :value="60">60 Menit / Toko</option>
              </select>
            </div>
          </div>

          <!-- Add Stores Combobox for Batch -->
          <div class="space-y-1.5 relative" ref="batchStoreDropdownRef">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
              Cari & Tambah Toko ke Daftar Batch
            </label>
            <div class="relative flex items-center">
              <Search class="w-4 h-4 text-slate-400 absolute left-4 pointer-events-none" />
              <input
                v-model="batchStoreSearch"
                type="text"
                placeholder="Ketik nama toko untuk menambahkannya ke batch..."
                @focus="isBatchStoreDropdownOpen = true"
                class="w-full pl-10 pr-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs sm:text-sm font-bold text-slate-700 focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all outline-none"
              />
            </div>

            <!-- Batch Store Results Dropdown -->
            <div
              v-if="isBatchStoreDropdownOpen && batchStoreSearch"
              class="absolute left-0 right-0 top-full mt-1.5 bg-white rounded-2xl border border-slate-200 shadow-2xl max-h-52 overflow-y-auto no-scrollbar z-50 divide-y divide-slate-50"
            >
              <div v-if="filteredStoresBatch.length === 0" class="p-4 text-center text-xs text-slate-400 font-medium">
                Tidak ada toko ditemukan
              </div>
              <button
                v-for="s in filteredStoresBatch"
                :key="s.id"
                type="button"
                @click="addStoreToBatch(s)"
                class="w-full px-4 py-2.5 text-left hover:bg-primary/5 flex items-center justify-between gap-2 transition-all cursor-pointer group"
              >
                <div class="min-w-0 flex-1">
                  <p class="font-bold text-xs text-slate-800 group-hover:text-primary truncate">{{ s.name }}</p>
                  <p class="text-[10px] text-slate-400 truncate">{{ s.address || 'Tanpa alamat' }}</p>
                </div>
                <span class="px-2 py-1 rounded-lg text-[10px] font-black bg-primary/10 text-primary flex items-center gap-1 shrink-0">
                  <Plus class="w-3 h-3" /> Tambah
                </span>
              </button>
            </div>
          </div>

          <!-- Batch Stores List & Timeline Preview -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
                Urutan Rute & Waktu Kunjungan ({{ batchStores.length }} Toko)
              </label>
              <button
                v-if="batchStores.length > 0"
                type="button"
                @click="batchStores = []"
                class="text-[10px] font-bold text-rose-500 hover:text-rose-700 cursor-pointer"
              >
                Hapus Semua
              </button>
            </div>

            <div v-if="batchStores.length === 0" class="py-8 px-4 text-center border-2 border-dashed border-slate-200 rounded-2xl">
              <Store class="w-8 h-8 text-slate-300 mx-auto mb-1.5" />
              <p class="text-xs font-bold text-slate-500">Belum ada toko yang ditambahkan ke batch</p>
              <p class="text-[11px] text-slate-400 mt-0.5">Ketik nama toko di kotak pencarian di atas untuk memasukkan ke timeline.</p>
            </div>

            <div v-else class="space-y-2 max-h-56 overflow-y-auto no-scrollbar pr-1">
              <div
                v-for="(item, idx) in calculatedBatchTimeline"
                :key="item.id"
                class="p-3 bg-slate-50 hover:bg-slate-100/80 rounded-2xl border border-slate-200 flex items-center justify-between gap-2.5 transition-all"
              >
                <!-- Index & Time Badge -->
                <div class="flex items-center gap-2 shrink-0">
                  <span class="w-5 h-5 rounded-full bg-slate-200 text-slate-700 text-[10px] font-black flex items-center justify-center">
                    {{ idx + 1 }}
                  </span>
                  <span class="px-2 py-1 rounded-xl bg-primary/10 text-primary font-black text-xs flex items-center gap-1">
                    <Clock class="w-3 h-3" /> {{ item.timeStr }}
                  </span>
                </div>

                <!-- Store Info -->
                <div class="min-w-0 flex-1">
                  <p class="font-black text-slate-800 text-xs truncate">{{ item.name }}</p>
                  <p class="text-[10px] text-slate-400 truncate">{{ item.address || 'Alamat tidak tersedia' }}</p>
                </div>

                <!-- Reorder & Remove Actions -->
                <div class="flex items-center gap-1 shrink-0">
                  <button
                    type="button"
                    @click="moveBatchStore(idx, 'up')"
                    :disabled="idx === 0"
                    class="p-1 text-slate-400 hover:text-slate-700 disabled:opacity-20 cursor-pointer"
                  >
                    <ChevronUp class="w-4 h-4" />
                  </button>
                  <button
                    type="button"
                    @click="moveBatchStore(idx, 'down')"
                    :disabled="idx === calculatedBatchTimeline.length - 1"
                    class="p-1 text-slate-400 hover:text-slate-700 disabled:opacity-20 cursor-pointer"
                  >
                    <ChevronDown class="w-4 h-4" />
                  </button>
                  <button
                    type="button"
                    @click="removeStoreFromBatch(idx)"
                    class="p-1 text-rose-400 hover:text-rose-600 cursor-pointer ml-1"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Batch Action Button -->
          <div class="pt-2">
            <Button
              type="button"
              @click="handleSaveBatch"
              :disabled="isSavingBatch || batchStores.length === 0"
              class="w-full py-4 rounded-2xl font-black text-xs sm:text-sm shadow-xl shadow-primary/20 flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50"
            >
              <CheckCircle2 class="w-4 h-4" />
              <span>{{ isSavingBatch ? 'Menyimpan Batch...' : `Simpan Semua (${batchStores.length} Kunjungan Sekaligus)` }}</span>
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <div v-if="showDetail" class="fixed inset-0 z-[100] flex items-center justify-center p-3 sm:p-6">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showDetail = false"
      ></div>

      <div
        class="relative w-full max-w-4xl bg-white rounded-3xl sm:rounded-[40px] shadow-2xl overflow-hidden flex flex-col md:flex-row max-h-[90vh] md:h-[80vh] animate-in fade-in zoom-in duration-300"
      >
        <!-- Left: Image & Info -->
        <div class="w-full md:w-[380px] lg:w-[400px] min-h-[260px] md:min-h-0 bg-slate-950 relative flex items-center justify-center shrink-0">
          <img
            v-if="selectedVisit?.selfie_url"
            :src="selectedVisit?.selfie_url"
            class="w-full h-full object-contain max-h-[280px] md:max-h-none"
            alt="Visit Selfie"
          />
          <a
            v-if="selectedVisit?.selfie_url"
            :href="selectedVisit?.selfie_url"
            target="_blank"
            class="absolute top-4 right-4 bg-slate-900/60 backdrop-blur-md hover:bg-slate-900/80 text-white rounded-full p-2.5 transition-all shadow-lg border border-white/10 z-20"
            title="Buka Gambar Penuh"
          >
            <ExternalLink class="w-4 h-4" />
          </a>
          <div
            v-else
            class="w-full h-full min-h-[160px] flex flex-col items-center justify-center text-slate-500 gap-3"
          >
            <Camera class="w-12 h-12 md:w-16 md:h-16 opacity-20" />
            <p class="text-[10px] md:text-xs font-bold uppercase tracking-widest">Tidak ada foto</p>
          </div>
          <div
            class="absolute inset-0 bg-gradient-to-t from-slate-900 via-transparent to-transparent pointer-events-none"
          ></div>

          <div class="absolute bottom-4 left-4 right-4 md:bottom-8 md:left-8 md:right-8 text-white">
            <div class="flex items-center gap-3 mb-3 md:mb-4">
              <div
                class="w-9 h-9 md:w-12 md:h-12 rounded-xl md:rounded-2xl bg-white/20 backdrop-blur-md flex items-center justify-center shrink-0"
              >
                <Camera class="w-4 h-4 md:w-6 md:h-6 text-white" />
              </div>
              <div>
                <p class="text-[9px] md:text-[10px] font-black text-white/60 uppercase tracking-widest">
                  Selfie Kunjungan
                </p>
                <p class="text-xs md:text-sm font-bold truncate">{{ selectedVisit?.employee?.first_name }}</p>
              </div>
            </div>

            <div class="bg-white/10 backdrop-blur-md rounded-xl md:rounded-2xl p-3 md:p-4 border border-white/10 flex flex-col gap-2 md:gap-3">
              <!-- Lokasi Check-in Aktual -->
              <div class="flex items-start gap-2 sm:gap-3">
                <MapPin class="w-3.5 h-3.5 md:w-4 md:h-4 text-emerald-400 mt-0.5 shrink-0" />
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between gap-1.5 mb-1">
                    <p class="text-[8px] md:text-[9px] font-black text-emerald-400 uppercase tracking-widest">
                      📍 LOKASI CHECK-IN (GPS AKTUAL)
                    </p>
                    <div class="flex items-center gap-1 shrink-0">
                      <button
                        type="button"
                        v-if="selectedVisit?.latitude && selectedVisit?.longitude"
                        @click.stop="openGoogleStreetView(selectedVisit.latitude, selectedVisit.longitude)"
                        class="px-2 py-0.5 rounded-lg bg-emerald-500/30 hover:bg-emerald-500/50 text-emerald-200 hover:text-white text-[8.5px] font-black border border-emerald-400/40 flex items-center gap-1 transition-all shadow-sm cursor-pointer"
                        title="Buka Google Street View 360°"
                      >
                        <Navigation2 class="w-2.5 h-2.5" /> Street View 360°
                      </button>
                      <button
                        type="button"
                        v-if="selectedVisit?.latitude && selectedVisit?.longitude"
                        @click.stop="openGoogleMaps(selectedVisit.latitude, selectedVisit.longitude, selectedVisit.store?.name)"
                        class="px-2 py-0.5 rounded-lg bg-white/15 hover:bg-white/30 text-white text-[8.5px] font-black border border-white/20 flex items-center gap-1 transition-all shadow-sm cursor-pointer"
                        title="Buka di Google Maps"
                      >
                        <ExternalLink class="w-2.5 h-2.5" /> Maps
                      </button>
                    </div>
                  </div>
                  <p class="text-[10px] md:text-[11px] font-medium leading-relaxed line-clamp-2 md:line-clamp-none">
                    {{ actualCheckinAddress }}
                  </p>
                </div>
              </div>
              
              <!-- Alamat Toko Terdaftar -->
              <div class="flex items-start gap-2 sm:gap-3 opacity-85">
                <Store class="w-3.5 h-3.5 md:w-4 md:h-4 text-white mt-0.5 shrink-0" />
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between gap-1.5 mb-0.5">
                    <p class="text-[8px] md:text-[9px] font-black text-white/80 uppercase tracking-widest">
                      🏪 ALAMAT TOKO (TERDAFTAR)
                    </p>
                    <button
                      type="button"
                      v-if="selectedVisit?.store?.latitude && selectedVisit?.store?.longitude"
                      @click.stop="openGoogleMaps(selectedVisit.store.latitude, selectedVisit.store.longitude, selectedVisit.store?.name)"
                      class="px-1.5 py-0.5 rounded bg-white/10 hover:bg-white/20 text-white/90 text-[8px] font-bold border border-white/10 flex items-center gap-1 cursor-pointer"
                      title="Lihat Alamat Toko di Maps"
                    >
                      <ExternalLink class="w-2 h-2" /> Toko Maps
                    </button>
                  </div>
                  <p class="text-[10px] md:text-[11px] font-medium leading-relaxed line-clamp-2 md:line-clamp-none">
                    {{ selectedVisit?.store?.address || 'Tidak ada alamat terdaftar' }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Right: Activity Data -->
        <div class="flex-1 p-5 sm:p-8 md:p-10 flex flex-col gap-5 sm:gap-8 overflow-y-auto no-scrollbar">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-lg sm:text-2xl font-black text-slate-900 tracking-tight">
                {{ selectedVisit?.store?.name }}
              </h2>
              <p class="text-slate-400 font-bold text-xs sm:text-sm uppercase tracking-wider mt-0.5">
                {{ selectedVisit?.type }} KUNJUNGAN
              </p>
            </div>
            <button
              @click="showDetail = false"
              class="p-2 sm:p-3 bg-slate-100 hover:bg-slate-200 rounded-xl sm:rounded-2xl transition-all text-slate-500"
            >
              <X class="w-5 h-5 sm:w-6 sm:h-6" />
            </button>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 sm:gap-4">
            <div class="bg-slate-50 p-3.5 sm:p-4 rounded-2xl sm:rounded-3xl border border-slate-100">
              <div class="flex items-center gap-2 text-slate-400 mb-1">
                <Clock class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                <span class="text-[9px] sm:text-[10px] font-black uppercase tracking-wider">Waktu Masuk</span>
              </div>
              <p class="text-xs sm:text-sm font-black text-slate-900">
                {{ selectedVisit ? formatFullTime(selectedVisit.check_in_time) : '-' }}
              </p>
            </div>
            <div class="bg-slate-50 p-3.5 sm:p-4 rounded-2xl sm:rounded-3xl border border-slate-100">
              <div class="flex items-center gap-2 text-slate-400 mb-1">
                <Clock class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
                <span class="text-[9px] sm:text-[10px] font-black uppercase tracking-wider">Waktu Keluar</span>
              </div>
              <p class="text-xs sm:text-sm font-black text-slate-900">
                {{
                  selectedVisit?.check_out_time
                    ? formatFullTime(selectedVisit.check_out_time)
                    : 'Masih Aktif'
                }}
              </p>
            </div>
            <div class="bg-slate-50 p-3.5 sm:p-4 rounded-2xl sm:rounded-3xl border border-slate-100 flex flex-col justify-between">
              <div>
                <div class="flex items-center gap-2 text-slate-400 mb-1">
                  <MapPin class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-emerald-600" />
                  <span class="text-[9px] sm:text-[10px] font-black uppercase tracking-wider">Koordinat GPS</span>
                </div>
                <p
                  class="text-[11px] sm:text-xs font-black text-slate-900 leading-tight truncate"
                  :title="`${selectedVisit?.latitude}, ${selectedVisit?.longitude}`"
                >
                  {{ selectedVisit?.latitude || '0.0' }}, {{ selectedVisit?.longitude || '0.0' }}
                </p>
              </div>
              <div class="flex items-center gap-1.5 mt-2">
                <button
                  type="button"
                  v-if="selectedVisit?.latitude && selectedVisit?.longitude"
                  @click="openGoogleStreetView(selectedVisit.latitude, selectedVisit.longitude)"
                  class="flex-1 py-1 px-2 rounded-xl bg-indigo-50 hover:bg-indigo-100 text-indigo-700 text-[9.5px] font-black flex items-center justify-center gap-1 transition-all border border-indigo-200 cursor-pointer shadow-xs"
                >
                  <Navigation2 class="w-3 h-3" /> Street View 360°
                </button>
                <button
                  type="button"
                  v-if="selectedVisit?.latitude && selectedVisit?.longitude"
                  @click="openGoogleMaps(selectedVisit.latitude, selectedVisit.longitude, selectedVisit.store?.name)"
                  class="p-1 px-2 rounded-xl bg-slate-200 hover:bg-slate-300 text-slate-800 text-[9.5px] font-black flex items-center justify-center gap-1 transition-all cursor-pointer shadow-xs"
                  title="Google Maps"
                >
                  <ExternalLink class="w-3 h-3" /> Maps
                </button>
              </div>
            </div>
          </div>

          <div>
            <h4 class="text-xs font-black text-slate-400 uppercase tracking-[0.2em] mb-3 sm:mb-4">
              Hasil / Transaksi
            </h4>
            <div
              v-if="
                (selectedVisit?.transactions?.length || 0) +
                  (selectedVisit?.sales_orders?.length || 0) >
                0
              "
              class="space-y-3"
            >
              <!-- Regular Transactions -->
              <div
                v-for="t in selectedVisit?.transactions"
                :key="t.id"
                :class="[
                  'p-3.5 sm:p-4 rounded-2xl sm:rounded-3xl border flex flex-col gap-3',
                  t.payment_method === 'TEMPO'
                    ? 'bg-orange-50 border-orange-100'
                    : 'bg-emerald-50 border-emerald-100',
                ]"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-3">
                    <div
                      :class="[
                        'w-9 h-9 sm:w-10 sm:h-10 rounded-xl sm:rounded-2xl flex items-center justify-center shadow-lg shrink-0',
                        t.payment_method === 'TEMPO'
                          ? 'bg-orange-500 shadow-orange-200'
                          : 'bg-emerald-500 shadow-emerald-200',
                      ]"
                    >
                      <FileText class="w-4 h-4 sm:w-5 sm:h-5 text-white" />
                    </div>
                    <div>
                      <p class="font-bold text-slate-900 text-xs sm:text-sm">{{ t.receipt_no }}</p>
                      <p
                        :class="[
                          'text-xs sm:text-sm font-black',
                          t.payment_method === 'TEMPO' ? 'text-orange-600' : 'text-emerald-600',
                        ]"
                      >
                        Rp {{ new Intl.NumberFormat('id-ID').format(t.total_amount) }}
                      </p>
                    </div>
                  </div>
                  <div class="flex flex-col items-end gap-1">
                    <!-- Payment Method Badge -->
                    <span
                      v-if="t.payment_method === 'CASH'"
                      class="px-2 py-0.5 bg-green-100 text-green-700 text-[9px] sm:text-[10px] font-black rounded-full uppercase border border-green-200"
                      >💵 CASH</span
                    >
                    <span
                      v-else-if="t.payment_method === 'TEMPO'"
                      class="px-2 py-0.5 bg-orange-100 text-orange-700 text-[9px] sm:text-[10px] font-black rounded-full uppercase border border-orange-200"
                      >⏳ TEMPO</span
                    >
                    <span
                      v-else-if="t.payment_method === 'QRIS'"
                      class="px-2 py-0.5 bg-purple-100 text-purple-700 text-[9px] sm:text-[10px] font-black rounded-full uppercase border border-purple-200"
                      >📲 QRIS</span
                    >
                    <span
                      v-else-if="t.payment_method === 'VA'"
                      class="px-2 py-0.5 bg-blue-100 text-blue-700 text-[9px] sm:text-[10px] font-black rounded-full uppercase border border-blue-200"
                      >🏦 VA Bank</span
                    >
                    <span
                      v-else
                      class="px-2 py-0.5 bg-slate-100 text-slate-500 text-[9px] sm:text-[10px] font-black rounded-full uppercase"
                      >{{ t.payment_method || '-' }}</span
                    >
                    <!-- Status Badge -->
                    <span
                      :class="[
                        'px-2 py-0.5 text-[9px] sm:text-[10px] font-black rounded-full uppercase',
                        t.payment_method === 'TEMPO'
                          ? 'bg-orange-500 text-white'
                          : 'bg-emerald-500 text-white',
                      ]"
                      >{{ t.payment_status || t.status }}</span
                    >
                  </div>
                </div>
                <!-- TEMPO DP Info block -->
                <div
                  v-if="t.payment_method === 'TEMPO'"
                  class="bg-white/70 rounded-2xl p-3 border border-orange-200 space-y-1.5"
                >
                  <p class="text-[10px] font-black text-orange-700 uppercase tracking-wider">
                    📋 Info Pembayaran Tempo
                  </p>
                  <div class="grid grid-cols-2 gap-x-4 gap-y-1 mt-1">
                    <div>
                      <p class="text-[9px] font-black text-slate-400 uppercase">Total Tagihan</p>
                      <p class="text-xs font-black text-slate-800">
                        Rp {{ new Intl.NumberFormat('id-ID').format(t.total_amount) }}
                      </p>
                    </div>
                    <div v-if="t.paid_amount > 0">
                      <p class="text-[9px] font-black text-slate-400 uppercase">DP Dibayar</p>
                      <p class="text-xs font-black text-orange-600">
                        Rp {{ new Intl.NumberFormat('id-ID').format(t.paid_amount) }}
                        <span
                          v-if="t.dp_payment_method"
                          class="text-[9px] font-bold text-orange-400"
                          >({{ t.dp_payment_method }})</span
                        >
                      </p>
                    </div>
                    <div v-if="t.paid_amount > 0">
                      <p class="text-[9px] font-black text-slate-400 uppercase">Sisa Tagihan</p>
                      <p class="text-xs font-black text-red-600">
                        Rp
                        {{ new Intl.NumberFormat('id-ID').format(t.total_amount - t.paid_amount) }}
                      </p>
                    </div>
                    <div v-if="t.payment_due_date">
                      <p class="text-[9px] font-black text-slate-400 uppercase">Jatuh Tempo</p>
                      <p class="text-xs font-black text-slate-700">
                        {{
                          new Date(t.payment_due_date).toLocaleDateString('id-ID', {
                            day: '2-digit',
                            month: 'short',
                            year: 'numeric',
                          })
                        }}
                      </p>
                    </div>
                  </div>
                  <!-- DP Status indicator -->
                  <div v-if="t.paid_amount > 0" class="mt-2 pt-2 border-t border-orange-100">
                    <span
                      v-if="t.payment_status === 'PARTIAL'"
                      class="inline-flex items-center gap-1 px-2.5 py-1 bg-amber-100 text-amber-700 text-[9px] font-black rounded-lg uppercase border border-amber-200"
                    >
                      ✅ DP Tervalidasi kasir — sisa belum lunas
                    </span>
                    <span
                      v-else
                      class="inline-flex items-center gap-1 px-2.5 py-1 bg-orange-100 text-orange-700 text-[9px] font-black rounded-lg uppercase border border-orange-200"
                    >
                      ⏳ DP belum divalidasi kasir
                    </span>
                  </div>
                </div>
              </div>

              <!-- Sales Orders (TO) -->
              <div
                v-for="so in selectedVisit?.sales_orders"
                :key="so.id"
                class="p-3.5 sm:p-4 bg-blue-50 rounded-2xl sm:rounded-3xl border border-blue-100 flex items-center justify-between"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl sm:rounded-2xl bg-blue-500 flex items-center justify-center shadow-lg shadow-blue-200 shrink-0"
                  >
                    <FileText class="w-4 h-4 sm:w-5 sm:h-5 text-white" />
                  </div>
                  <div>
                    <p class="font-bold text-slate-900 text-xs sm:text-sm">{{ so.so_number }}</p>
                    <p class="text-xs font-black text-blue-600">
                      Rp {{ new Intl.NumberFormat('id-ID').format(so.total_amount) }}
                    </p>
                  </div>
                </div>
                <div
                  class="px-2.5 py-1 bg-blue-500 text-white text-[9px] sm:text-[10px] font-black rounded-full uppercase"
                >
                  {{ so.status }}
                </div>
              </div>
            </div>
            <div
              v-else
              class="py-8 sm:py-10 text-center bg-slate-50 rounded-2xl sm:rounded-3xl border border-dashed border-slate-200 px-4"
            >
              <p class="text-xs sm:text-sm font-bold text-slate-400 italic">
                Tidak ada transaksi atau order tercatat dalam kunjungan ini.
              </p>
            </div>
          </div>

          <div v-if="selectedVisit?.notes">
            <h4 class="text-xs font-black text-slate-400 uppercase tracking-[0.2em] mb-2 sm:mb-3">
              Catatan / Feedback
            </h4>
            <div
              class="bg-amber-50 p-4 sm:p-6 rounded-2xl sm:rounded-3xl border border-amber-100 italic text-slate-700 text-xs sm:text-sm leading-relaxed"
            >
              "{{ selectedVisit.notes }}"
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Image Inspection Modal (Speedometer / BBM Receipt) -->
    <div v-if="showImageModal" class="fixed inset-0 z-[110] flex items-center justify-center p-3 sm:p-6 animate-in fade-in duration-200">
      <div class="absolute inset-0 bg-black/80 backdrop-blur-md" @click="showImageModal = false"></div>
      <div class="relative max-w-2xl w-full bg-slate-900 rounded-3xl overflow-hidden shadow-2xl border border-white/10 z-10 flex flex-col">
        <!-- Header -->
        <div class="p-4 sm:p-5 flex items-center justify-between border-b border-white/10 bg-slate-950/80">
          <div>
            <h3 class="text-white font-black text-sm sm:text-base">{{ modalImageTitle }}</h3>
            <p v-if="modalImageSub" class="text-white/60 text-xs mt-0.5">{{ modalImageSub }}</p>
          </div>
          <div class="flex items-center gap-2">
            <a
              :href="modalImageUrl"
              target="_blank"
              class="p-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-all text-xs flex items-center gap-1.5"
              title="Buka Tab Baru"
            >
              <ExternalLink class="w-4 h-4" />
            </a>
            <button
              @click="showImageModal = false"
              class="p-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-all"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>
        <!-- Photo Container -->
        <div class="p-4 flex items-center justify-center bg-black/90 max-h-[70vh] overflow-auto">
          <img :src="modalImageUrl" class="max-h-[65vh] w-auto object-contain rounded-xl shadow-lg" alt="Preview" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
:deep(.sales-seq-marker) {
  background: transparent !important;
  border: none !important;
}
</style>
