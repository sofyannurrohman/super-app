<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import { useMasterDataStore } from '@/stores/masterData'
import {
  Building2,
  TrendingUp,
  TrendingDown,
  DollarSign,
  ArrowRight,
  Printer,
  Sparkles,
  PieChart,
  MapPin,
  Layers,
  Filter,
  CheckCircle2,
  AlertCircle,
  ChevronRight,
  Search,
  Wallet,
  Coins,
  Receipt,
  Users,
  Building,
  Truck,
  FileSpreadsheet,
  X,
  BadgeAlert,
  ArrowDownRight,
  ArrowUpRight,
  Eye,
  ExternalLink,
  FileText,
  Loader2,
  Calendar,
  CalendarDays,
  RotateCcw
} from 'lucide-vue-next'

const authStore = useAuthStore()
const companyName = computed(() => authStore.user?.company_name || 'Perusahaan')

interface IncomeDetailItem {
  id: string
  transaction_type: string
  reference_no: string
  transaction_date: string
  entity_name: string
  payment_method?: string
  description: string
  amount: number
  status: string
  created_by_name?: string
  proof_url?: string
}

interface IncomeDetailResponse {
  income_type: string
  income_title: string
  period_month: number
  period_year: number
  branch_id?: string
  total_amount: number
  total_count: number
  items: IncomeDetailItem[]
}

interface ExpenseDetailItem {
  id: string
  reference_no: string
  transaction_date: string
  kas_name: string
  category_key: string
  category_name: string
  recipient_name?: string
  description: string
  amount: number
  status: string
  proof_url?: string
}

interface ExpenseDetailResponse {
  category_key: string
  category_name: string
  group: string
  period_month: number
  period_year: number
  start_date?: string
  end_date?: string
  branch_id?: string
  total_amount: number
  total_count: number
  items: ExpenseDetailItem[]
}

interface ReceivableDetailItem {
  id: string
  transaction_type: 'DEBET' | 'KREDIT'
  reference_no: string
  transaction_date: string
  kas_name: string
  description: string
  amount: number
  status: string
  proof_url?: string
}

interface ReceivableDetailResponse {
  entity_code: string
  entity_name: string
  category: string
  period_month: number
  period_year: number
  start_date?: string
  end_date?: string
  branch_id?: string
  initial_balance: number
  period_addition: number
  period_payment: number
  ending_balance: number
  status: string
  total_count: number
  items: ReceivableDetailItem[]
}

interface ProfitLossItem {
  company_id: string
  company_name: string
  period_month: number
  period_year: number
  total_revenue: number
  total_cogs: number
  total_payroll: number
  total_expenses: number
  gross_profit: number
  net_profit: number
}

interface BranchProfitLossItem {
  branch_id: string
  branch_name: string
  total_revenue: number
  total_cogs: number
  total_payroll: number
  total_expenses: number
  gross_profit: number
  net_profit: number
}

interface ProfitLossDetailReport {
  company_id: string
  company_name: string
  period_month: number
  period_year: number
  total_revenue: number
  total_cogs: number
  total_payroll: number
  total_expenses: number
  gross_profit: number
  net_profit: number
  branches: BranchProfitLossItem[]
}

interface IncomeBreakdown {
  product_sales: number
  interest_income: number
  ampas_income: number
  rosok_income: number
  claim_penalty: number
  other_income: number
  total_other_income: number
  total_revenue: number
}

interface ExpenseAccountItem {
  category_key: string
  name: string
  group: 'LABOR' | 'UTILITIES' | 'FLEET' | 'TAX' | 'GENERAL'
  amount: number
}

interface ReceivableEntityItem {
  code: string
  name: string
  category: 'CABANG' | 'BOOSTER' | 'MARKETPLACE' | 'KARYAWAN' | 'PROYEK'
  initial_balance: number
  period_addition: number
  period_payment: number
  ending_balance: number
  status: string
}

interface CashierProfitLossReport {
  period_month: number
  period_year: number
  branch_id?: string
  branch_name?: string
  income: IncomeBreakdown
  total_cogs: number
  gross_profit: number
  expenses: ExpenseAccountItem[]
  total_expenses: number
  total_payroll_direct: number
  net_profit: number
  profit_margin: number
  receivables: ReceivableEntityItem[]
  total_receivables_sum: number
}

const masterDataStore = useMasterDataStore()

// Main Tab State: 'cashier' (Laba Rugi Kasir Pabrik) vs 'executive' (Multi-PT Komparatif)
const activeMainTab = ref<'cashier' | 'executive'>('cashier')

// Filter State
const isLoading = ref(false)
const filterMode = ref<'month' | 'range'>('month') // 'month' (default) | 'range' (custom date range)
const selectedMonth = ref(new Date().getMonth() + 1)
const selectedYear = ref(new Date().getFullYear())

// Helper to format Date to YYYY-MM-DD
const formatDateToISO = (d: Date): string => {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

// Default start and end dates (awal bulan s/d akhir bulan berjalan)
const now = new Date()
const firstDayOfMonth = new Date(now.getFullYear(), now.getMonth(), 1)
const lastDayOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0)
const startDate = ref<string>(formatDateToISO(firstDayOfMonth))
const endDate = ref<string>(formatDateToISO(lastDayOfMonth))

const selectedBranchId = ref<string>('ALL')
const selectedCompanyId = ref<string>('ALL')

// Preset Range Helpers
const setPresetRange = (preset: 'today' | 'this_month' | 'last_30_days' | 'this_year') => {
  const today = new Date()
  if (preset === 'today') {
    startDate.value = formatDateToISO(today)
    endDate.value = formatDateToISO(today)
  } else if (preset === 'this_month') {
    startDate.value = formatDateToISO(new Date(today.getFullYear(), today.getMonth(), 1))
    endDate.value = formatDateToISO(new Date(today.getFullYear(), today.getMonth() + 1, 0))
  } else if (preset === 'last_30_days') {
    const past30 = new Date()
    past30.setDate(today.getDate() - 29)
    startDate.value = formatDateToISO(past30)
    endDate.value = formatDateToISO(today)
  } else if (preset === 'this_year') {
    startDate.value = `${today.getFullYear()}-01-01`
    endDate.value = `${today.getFullYear()}-12-31`
  }
  fetchProfitLoss()
}

// Data Containers for Executive Multi-PT View
const profitLossList = ref<ProfitLossItem[]>([])
const profitLossDetail = ref<ProfitLossDetailReport | null>(null)
const singleBranchReport = ref<ProfitLossItem | null>(null)

// Data Container for Cashier Profit & Loss View
const cashierReport = ref<CashierProfitLossReport | null>(null)

// Income Detail Drilldown Modal State
const showIncomeDetailModal = ref(false)
const isLoadingIncomeDetail = ref(false)
const incomeDetailData = ref<IncomeDetailResponse | null>(null)
const incomeDetailSearch = ref('')

const openIncomeDetail = async (incomeType: string, customTitle?: string) => {
  showIncomeDetailModal.value = true
  isLoadingIncomeDetail.value = true
  incomeDetailSearch.value = ''
  incomeDetailData.value = null
  try {
    const branchParam = selectedBranchId.value !== 'ALL' ? `&branch_id=${selectedBranchId.value}` : ''
    let queryParam = ''
    if (filterMode.value === 'range' && startDate.value && endDate.value) {
      queryParam = `&start_date=${startDate.value}&end_date=${endDate.value}`
    } else {
      queryParam = `&month=${selectedMonth.value}&year=${selectedYear.value}`
    }

    const res = await apiClient.get(
      `/finance/profit-loss/income-detail?income_type=${incomeType}${queryParam}${branchParam}`
    )
    incomeDetailData.value = res.data?.data || null
    if (customTitle && incomeDetailData.value) {
      incomeDetailData.value.income_title = customTitle
    }
  } catch (err: any) {
    toast.error('Gagal memuat rincian transaksi pendapatan')
  } finally {
    isLoadingIncomeDetail.value = false
  }
}

const filteredIncomeDetailItems = computed(() => {
  if (!incomeDetailData.value?.items) return []
  if (!incomeDetailSearch.value.trim()) return incomeDetailData.value.items
  const q = incomeDetailSearch.value.toLowerCase()
  return incomeDetailData.value.items.filter((item) => {
    return (
      item.reference_no?.toLowerCase().includes(q) ||
      item.entity_name?.toLowerCase().includes(q) ||
      item.description?.toLowerCase().includes(q) ||
      item.payment_method?.toLowerCase().includes(q) ||
      item.status?.toLowerCase().includes(q)
    )
  })
})

const incomeDetailTotalFiltered = computed(() => {
  return filteredIncomeDetailItems.value.reduce((sum, it) => sum + it.amount, 0)
})

// Expense Detail Drilldown Modal State
const showExpenseDetailModal = ref(false)
const isLoadingExpenseDetail = ref(false)
const expenseDetailData = ref<ExpenseDetailResponse | null>(null)
const expenseDetailSearch = ref('')

const openExpenseDetail = async (categoryKey: string, customTitle?: string) => {
  showExpenseDetailModal.value = true
  isLoadingExpenseDetail.value = true
  expenseDetailSearch.value = ''
  expenseDetailData.value = null
  try {
    const branchParam = selectedBranchId.value !== 'ALL' ? `&branch_id=${selectedBranchId.value}` : ''
    let queryParam = ''
    if (filterMode.value === 'range' && startDate.value && endDate.value) {
      queryParam = `&start_date=${startDate.value}&end_date=${endDate.value}`
    } else {
      queryParam = `&month=${selectedMonth.value}&year=${selectedYear.value}`
    }

    const res = await apiClient.get(
      `/finance/profit-loss/expense-detail?category_key=${categoryKey}${queryParam}${branchParam}`
    )
    expenseDetailData.value = res.data?.data || null
    if (customTitle && expenseDetailData.value) {
      expenseDetailData.value.category_name = customTitle
    }
  } catch (err: any) {
    toast.error('Gagal memuat rincian transaksi pengeluaran kas')
  } finally {
    isLoadingExpenseDetail.value = false
  }
}

const filteredExpenseDetailItems = computed(() => {
  if (!expenseDetailData.value?.items) return []
  if (!expenseDetailSearch.value.trim()) return expenseDetailData.value.items
  const q = expenseDetailSearch.value.toLowerCase()
  return expenseDetailData.value.items.filter((item) => {
    return (
      item.reference_no?.toLowerCase().includes(q) ||
      item.kas_name?.toLowerCase().includes(q) ||
      item.recipient_name?.toLowerCase().includes(q) ||
      item.description?.toLowerCase().includes(q) ||
      item.category_name?.toLowerCase().includes(q) ||
      item.status?.toLowerCase().includes(q)
    )
  })
})

const expenseDetailTotalFiltered = computed(() => {
  return filteredExpenseDetailItems.value.reduce((sum, it) => sum + it.amount, 0)
})

// Receivable Detail Drilldown Modal State
const showReceivableDetailModal = ref(false)
const isLoadingReceivableDetail = ref(false)
const receivableDetailData = ref<ReceivableDetailResponse | null>(null)
const receivableDetailSearch = ref('')

const openReceivableDetail = async (entityCode: string, customTitle?: string) => {
  showReceivableDetailModal.value = true
  isLoadingReceivableDetail.value = true
  receivableDetailSearch.value = ''
  receivableDetailData.value = null
  try {
    const branchParam = selectedBranchId.value !== 'ALL' ? `&branch_id=${selectedBranchId.value}` : ''
    let queryParam = ''
    if (filterMode.value === 'range' && startDate.value && endDate.value) {
      queryParam = `&start_date=${startDate.value}&end_date=${endDate.value}`
    } else {
      queryParam = `&month=${selectedMonth.value}&year=${selectedYear.value}`
    }

    const res = await apiClient.get(
      `/finance/profit-loss/receivable-detail?entity_code=${entityCode}${queryParam}${branchParam}`
    )
    receivableDetailData.value = res.data?.data || null
    if (customTitle && receivableDetailData.value) {
      receivableDetailData.value.entity_name = customTitle
    }
  } catch (err: any) {
    toast.error('Gagal memuat rincian mutasi piutang entitas')
  } finally {
    isLoadingReceivableDetail.value = false
  }
}

const filteredReceivableDetailItems = computed(() => {
  if (!receivableDetailData.value?.items) return []
  if (!receivableDetailSearch.value.trim()) return receivableDetailData.value.items
  const q = receivableDetailSearch.value.toLowerCase()
  return receivableDetailData.value.items.filter((item) => {
    return (
      item.reference_no?.toLowerCase().includes(q) ||
      item.kas_name?.toLowerCase().includes(q) ||
      item.description?.toLowerCase().includes(q) ||
      item.transaction_type?.toLowerCase().includes(q) ||
      item.status?.toLowerCase().includes(q)
    )
  })
})

const receivableDetailTotalDebetFiltered = computed(() => {
  return filteredReceivableDetailItems.value
    .filter((it) => it.transaction_type === 'DEBET')
    .reduce((sum, it) => sum + it.amount, 0)
})

const receivableDetailTotalKreditFiltered = computed(() => {
  return filteredReceivableDetailItems.value
    .filter((it) => it.transaction_type === 'KREDIT')
    .reduce((sum, it) => sum + it.amount, 0)
})

// Filter & Search inside Cashier Tab
const expenseSearch = ref('')
const selectedExpenseGroup = ref<string>('ALL')
const receivableSearch = ref('')
const selectedReceivableCategory = ref<string>('ALL')

// View mode for Executive tab
const executiveViewMode = computed<'all' | 'detail' | 'branch'>(() => {
  if (selectedBranchId.value !== 'ALL') return 'branch'
  if (selectedCompanyId.value !== 'ALL') return 'detail'
  return 'all'
})

const months = [
  { value: 1, name: 'Januari' },
  { value: 2, name: 'Februari' },
  { value: 3, name: 'Maret' },
  { value: 4, name: 'April' },
  { value: 5, name: 'Mei' },
  { value: 6, name: 'Juni' },
  { value: 7, name: 'Juli' },
  { value: 8, name: 'Agustus' },
  { value: 9, name: 'September' },
  { value: 10, name: 'Oktober' },
  { value: 11, name: 'November' },
  { value: 12, name: 'Desember' },
]

const years = [2025, 2026, 2027, 2028]

const selectedMonthName = computed(() => {
  return months.find((m) => m.value === selectedMonth.value)?.name || ''
})

const formatDisplayDate = (dStr: string) => {
  if (!dStr) return ''
  const parts = dStr.split('-')
  if (parts.length === 3) {
    return `${parts[2]}/${parts[1]}/${parts[0]}`
  }
  return dStr
}

const activePeriodLabel = computed(() => {
  if (filterMode.value === 'range') {
    if (startDate.value === endDate.value) {
      return formatDisplayDate(startDate.value)
    }
    return `${formatDisplayDate(startDate.value)} s/d ${formatDisplayDate(endDate.value)}`
  }
  return `${selectedMonthName.value} ${selectedYear.value}`
})

const currentBranchName = computed(() => {
  if (selectedBranchId.value === 'ALL') return 'Seluruh Pabrik & Unit'
  const br = masterDataStore.branches.find((b: any) => b.id === selectedBranchId.value)
  return br?.name || 'Cabang Terpilih'
})

const currentCompanyName = computed(() => {
  if (selectedCompanyId.value === 'ALL') return 'Seluruh PT'
  const comp = masterDataStore.companies.find((c: any) => c.id === selectedCompanyId.value)
  return comp?.name || profitLossDetail.value?.company_name || 'PT Terpilih'
})

// Filtered Expense Accounts
const filteredExpenses = computed(() => {
  if (!cashierReport.value?.expenses) return []
  return cashierReport.value.expenses.filter((item) => {
    const matchesSearch =
      item.name.toLowerCase().includes(expenseSearch.value.toLowerCase()) ||
      item.category_key.toLowerCase().includes(expenseSearch.value.toLowerCase())
    const matchesGroup =
      selectedExpenseGroup.value === 'ALL' || item.group === selectedExpenseGroup.value
    return matchesSearch && matchesGroup
  })
})

// Filtered 28 Receivables
const filteredReceivables = computed(() => {
  if (!cashierReport.value?.receivables) return []
  return cashierReport.value.receivables.filter((item) => {
    const matchesSearch =
      item.name.toLowerCase().includes(receivableSearch.value.toLowerCase()) ||
      item.code.toLowerCase().includes(receivableSearch.value.toLowerCase())
    const matchesCategory =
      selectedReceivableCategory.value === 'ALL' || item.category === selectedReceivableCategory.value
    return matchesSearch && matchesCategory
  })
})

// Receivables Statistics
const totalActiveReceivables = computed(() => {
  if (!cashierReport.value?.receivables) return 0
  return cashierReport.value.receivables.reduce((sum, r) => sum + r.ending_balance, 0)
})

const totalSettledReceivablesCount = computed(() => {
  if (!cashierReport.value?.receivables) return 0
  return cashierReport.value.receivables.filter((r) => r.ending_balance <= 0).length
})

const totalActiveReceivablesCount = computed(() => {
  if (!cashierReport.value?.receivables) return 0
  return cashierReport.value.receivables.filter((r) => r.ending_balance > 0).length
})

// Expense Groups Subtotals
const getGroupLabel = (group: string) => {
  switch (group) {
    case 'LABOR':
      return 'Beban Tenaga Kerja & Borongan'
    case 'UTILITIES':
      return 'Beban Utilitas & Operasional Produksi'
    case 'FLEET':
      return 'Beban Pemeliharaan, Armada & Ekspedisi'
    case 'TAX':
      return 'Beban Pajak, Legal & Perizinan'
    case 'GENERAL':
      return 'Beban Umum & Administrasi'
    default:
      return 'Beban Operasional'
  }
}

const getGroupTotal = (group: string) => {
  if (!cashierReport.value?.expenses) return 0
  return cashierReport.value.expenses
    .filter((e) => e.group === group)
    .reduce((sum, e) => sum + e.amount, 0)
}

// Fetch Logic
const fetchProfitLoss = async () => {
  isLoading.value = true
  try {
    let dateParams = ''
    if (filterMode.value === 'range' && startDate.value && endDate.value) {
      dateParams = `start_date=${startDate.value}&end_date=${endDate.value}`
    } else {
      dateParams = `month=${selectedMonth.value}&year=${selectedYear.value}`
    }

    if (activeMainTab.value === 'cashier') {
      const branchParam = selectedBranchId.value !== 'ALL' ? `&branch_id=${selectedBranchId.value}` : ''
      const res = await apiClient.get(
        `/finance/profit-loss/cashier?${dateParams}${branchParam}`
      )
      cashierReport.value = res.data?.data || null
    } else {
      if (selectedBranchId.value !== 'ALL') {
        const res = await apiClient.get(
          `/finance/profit-loss/branch?branch_id=${selectedBranchId.value}&${dateParams}`
        )
        singleBranchReport.value = res.data?.data || null
        profitLossDetail.value = null
        profitLossList.value = []
      } else if (selectedCompanyId.value !== 'ALL') {
        const res = await apiClient.get(
          `/finance/profit-loss/detail?company_id=${selectedCompanyId.value}&${dateParams}`
        )
        profitLossDetail.value = res.data?.data || null
        singleBranchReport.value = null
        profitLossList.value = []
      } else {
        const res = await apiClient.get(
          `/finance/profit-loss/all?${dateParams}`
        )
        profitLossList.value = res.data?.data || []
        profitLossDetail.value = null
        singleBranchReport.value = null
      }
    }
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal memuat data Laporan Laba Rugi')
  } finally {
    isLoading.value = false
  }
}

watch(activeMainTab, () => {
  fetchProfitLoss()
})

const handleBranchChange = () => {
  fetchProfitLoss()
}

const handleCompanyChange = () => {
  fetchProfitLoss()
}

onMounted(async () => {
  await Promise.all([
    masterDataStore.fetchCompanies(),
    masterDataStore.fetchBranches(),
  ])
  if (selectedBranchId.value !== 'ALL') {
    const exists = masterDataStore.branches.some((b: any) => b.id === selectedBranchId.value)
    if (!exists) {
      selectedBranchId.value = 'ALL'
    }
  }
  fetchProfitLoss()
})

const formatCurrency = (val: number | undefined | null) => {
  if (val === undefined || val === null || isNaN(val)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

const printReport = () => {
  window.print()
}
</script>

<template>
  <div class="space-y-8 pb-16 print:bg-white print:p-0">
    <!-- Header & Filter Bar -->
    <div class="bg-white rounded-[2rem] border border-slate-200/70 p-6 shadow-sm print:hidden space-y-6">
      <div class="flex flex-col xl:flex-row xl:items-center justify-between gap-6">
        <div>
          <div class="flex items-center gap-3">
            <span class="p-2.5 bg-emerald-50 text-emerald-600 rounded-2xl shadow-inner">
              <PieChart class="w-6 h-6" />
            </span>
            <div>
              <h1 class="text-2xl font-black text-slate-900 tracking-tight">
                Laporan Laba Rugi
              </h1>
              <p class="text-xs font-semibold text-slate-500 mt-0.5">
                Pembukuan Kasir & Pabrik terintegrasi otomatis realtime dengan Manajemen Kas.
              </p>
            </div>
          </div>
        </div>

        <!-- Filter Controls -->
        <div class="flex flex-wrap items-center gap-3">
          <!-- Branch Selector -->
          <div class="flex items-center gap-1.5 bg-slate-50 border border-slate-200/80 rounded-2xl px-3.5 py-2 shadow-sm">
            <MapPin class="w-4 h-4 text-slate-400" />
            <select
              v-model="selectedBranchId"
              @change="handleBranchChange"
              class="bg-transparent text-xs font-bold text-slate-700 focus:outline-none cursor-pointer pr-2"
            >
              <option value="ALL">📍 Seluruh Pabrik & Unit</option>
              <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">
                {{ b.name }}
              </option>
            </select>
          </div>

          <!-- Filter Mode Toggle (Bulanan vs Rentang Tanggal) -->
          <div class="inline-flex bg-slate-100 p-1 rounded-2xl border border-slate-200/80 shadow-inner">
            <button
              @click="filterMode = 'month'; fetchProfitLoss()"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
              :class="filterMode === 'month' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
              title="Filter per periode bulan penuh"
            >
              <Calendar class="w-3.5 h-3.5" />
              <span>Bulanan</span>
            </button>
            <button
              @click="filterMode = 'range'; fetchProfitLoss()"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer"
              :class="filterMode === 'range' ? 'bg-white text-emerald-700 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
              title="Filter rentang tanggal kustom"
            >
              <CalendarDays class="w-3.5 h-3.5" />
              <span>Rentang Tanggal</span>
            </button>
          </div>

          <!-- Controls for Month Mode -->
          <template v-if="filterMode === 'month'">
            <!-- Month Picker -->
            <select
              v-model="selectedMonth"
              @change="fetchProfitLoss"
              class="bg-white border border-slate-200 px-3.5 py-2.5 rounded-2xl text-xs font-bold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
            >
              <option v-for="m in months" :key="m.value" :value="m.value">{{ m.name }}</option>
            </select>

            <!-- Year Picker -->
            <select
              v-model="selectedYear"
              @change="fetchProfitLoss"
              class="bg-white border border-slate-200 px-3.5 py-2.5 rounded-2xl text-xs font-bold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer"
            >
              <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
            </select>
          </template>

          <!-- Controls for Date Range Mode -->
          <template v-else>
            <div class="flex items-center gap-2 bg-slate-50 border border-slate-200/90 rounded-2xl px-3 py-1.5 shadow-sm">
              <div class="flex items-center gap-1.5">
                <span class="text-[10px] font-black uppercase text-slate-400">Dari:</span>
                <input
                  type="date"
                  v-model="startDate"
                  @change="fetchProfitLoss"
                  class="bg-white border border-slate-200 rounded-xl px-2.5 py-1 text-xs font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
                />
              </div>
              <span class="text-xs text-slate-400 font-bold">-</span>
              <div class="flex items-center gap-1.5">
                <span class="text-[10px] font-black uppercase text-slate-400">S/d:</span>
                <input
                  type="date"
                  v-model="endDate"
                  @change="fetchProfitLoss"
                  class="bg-white border border-slate-200 rounded-xl px-2.5 py-1 text-xs font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
                />
              </div>
            </div>

            <!-- Quick Presets -->
            <div class="hidden sm:flex items-center gap-1 bg-slate-100 p-1 rounded-2xl border border-slate-200/80">
              <button
                @click="setPresetRange('today')"
                class="px-2.5 py-1 text-[11px] font-bold rounded-xl text-slate-600 hover:bg-white hover:text-slate-900 transition cursor-pointer"
              >
                Hari Ini
              </button>
              <button
                @click="setPresetRange('this_month')"
                class="px-2.5 py-1 text-[11px] font-bold rounded-xl text-slate-600 hover:bg-white hover:text-slate-900 transition cursor-pointer"
              >
                Bulan Ini
              </button>
              <button
                @click="setPresetRange('last_30_days')"
                class="px-2.5 py-1 text-[11px] font-bold rounded-xl text-slate-600 hover:bg-white hover:text-slate-900 transition cursor-pointer"
              >
                30 Hari
              </button>
            </div>
          </template>

          <!-- Print Button -->
          <button
            @click="printReport"
            class="flex items-center gap-2 bg-slate-900 hover:bg-slate-800 text-white px-5 py-2.5 rounded-2xl text-xs font-bold transition-all shadow-md active:scale-95 cursor-pointer"
          >
            <Printer class="w-4 h-4 text-emerald-400" />
            <span>Cetak PDF Laba Rugi</span>
          </button>
        </div>
      </div>

      <!-- Tab Switcher Navigation -->
      <div class="flex items-center gap-2 pt-4 border-t border-slate-100">
        <button
          @click="activeMainTab = 'cashier'"
          class="flex items-center gap-2 px-5 py-2.5 rounded-2xl text-xs font-black transition-all cursor-pointer"
          :class="
            activeMainTab === 'cashier'
              ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-600/20'
              : 'bg-slate-100 hover:bg-slate-200 text-slate-600'
          "
        >
          <Receipt class="w-4 h-4" />
          <span>Laporan Laba Rugi Kasir & Pabrik</span>
        </button>

        <button
          @click="activeMainTab = 'executive'"
          class="flex items-center gap-2 px-5 py-2.5 rounded-2xl text-xs font-black transition-all cursor-pointer"
          :class="
            activeMainTab === 'executive'
              ? 'bg-slate-900 text-white shadow-lg shadow-slate-900/20'
              : 'bg-slate-100 hover:bg-slate-200 text-slate-600'
          "
        >
          <Layers class="w-4 h-4" />
          <span>Laba Rugi Eksekutif (Multi-PT)</span>
        </button>
      </div>
    </div>

    <!-- Printable Header (Visible Only on Print) -->
    <div class="hidden print:block border-b-2 border-slate-900 pb-5 mb-8 text-center">
      <h1 class="text-2xl font-black text-slate-900 uppercase tracking-tight">{{ companyName }}</h1>
      <h2 class="text-base font-bold text-slate-800 mt-1 uppercase">
        {{ activeMainTab === 'cashier' ? 'LAPORAN LABA RUGI & MONITORING PIUTANG KASIR' : 'LAPORAN LABA RUGI EKSEKUTIF MULTI-PT' }}
      </h2>
      <p class="text-xs text-slate-600 font-bold mt-1">
        Unit: {{ currentBranchName }} | Periode: {{ activePeriodLabel }}
      </p>
    </div>

    <!-- Loading Spinner -->
    <div v-if="isLoading" class="py-24 flex flex-col items-center justify-center gap-4 print:hidden">
      <span class="w-12 h-12 border-4 border-emerald-500 border-t-transparent rounded-full animate-spin"></span>
      <p class="text-sm font-bold text-slate-500">Mengkalkulasi Laporan Laba Rugi Realtime...</p>
    </div>

    <div v-else class="space-y-8">
      <!-- ================================================================= -->
      <!-- TAB 1: CASHIER & FACTORY PROFIT & LOSS (MAIN TAB)                 -->
      <!-- ================================================================= -->
      <template v-if="activeMainTab === 'cashier' && cashierReport">
        <!-- KPI Summary Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 print:hidden">
          <div class="bg-white p-5 rounded-3xl border border-slate-200/70 shadow-sm flex flex-col justify-between">
            <div class="flex items-center justify-between text-slate-400">
              <span class="text-[11px] font-black uppercase tracking-wider">Total Pendapatan</span>
              <TrendingUp class="w-4 h-4 text-emerald-500" />
            </div>
            <p class="text-2xl font-black text-slate-900 mt-3">
              {{ formatCurrency(cashierReport.income.total_revenue) }}
            </p>
            <p class="text-[11px] font-bold text-slate-400 mt-1">
              Termasuk Pendapatan Lain-lain: {{ formatCurrency(cashierReport.income.total_other_income) }}
            </p>
          </div>

          <div class="bg-white p-5 rounded-3xl border border-slate-200/70 shadow-sm flex flex-col justify-between">
            <div class="flex items-center justify-between text-slate-400">
              <span class="text-[11px] font-black uppercase tracking-wider">Laba Kotor (Gross Profit)</span>
              <DollarSign class="w-4 h-4 text-blue-500" />
            </div>
            <p class="text-2xl font-black text-blue-600 mt-3">
              {{ formatCurrency(cashierReport.gross_profit) }}
            </p>
            <p class="text-[11px] font-bold text-slate-400 mt-1">Setelah estimasi HPP Bahan Baku</p>
          </div>

          <div class="bg-white p-5 rounded-3xl border border-slate-200/70 shadow-sm flex flex-col justify-between">
            <div class="flex items-center justify-between text-slate-400">
              <span class="text-[11px] font-black uppercase tracking-wider">Total Beban Terinput</span>
              <TrendingDown class="w-4 h-4 text-rose-500" />
            </div>
            <p class="text-2xl font-black text-rose-600 mt-3">
              {{ formatCurrency(cashierReport.total_expenses) }}
            </p>
            <p class="text-[11px] font-bold text-slate-400 mt-1">Akumulasi realtime dari Kas Pabrik</p>
          </div>

          <div
            class="p-5 rounded-3xl border shadow-sm flex flex-col justify-between"
            :class="cashierReport.net_profit >= 0 ? 'bg-emerald-500/10 border-emerald-300' : 'bg-rose-500/10 border-rose-300'"
          >
            <div class="flex items-center justify-between">
              <span class="text-[11px] font-black uppercase tracking-wider text-slate-700">Laba Bersih (Net Profit)</span>
              <Sparkles class="w-4 h-4 text-emerald-600" />
            </div>
            <p
              class="text-2xl font-black mt-3"
              :class="cashierReport.net_profit >= 0 ? 'text-emerald-700' : 'text-rose-700'"
            >
              {{ formatCurrency(cashierReport.net_profit) }}
            </p>
            <p class="text-[11px] font-black text-slate-600 mt-1">
              Margin Bersih: {{ cashierReport.profit_margin.toFixed(1) }}%
            </p>
          </div>
        </div>

        <!-- SECTION 1: RINCIAN PENDAPATAN LAIN-LAIN -->
        <div class="bg-white rounded-3xl border border-slate-200/70 shadow-sm overflow-hidden">
          <div class="bg-slate-900 text-white p-5 flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <div class="p-2 bg-emerald-500/20 text-emerald-400 rounded-xl">
                <TrendingUp class="w-5 h-5" />
              </div>
              <div>
                <h3 class="text-base font-black tracking-tight">1. PENDAPATAN OPERASIONAL & LAIN-LAIN</h3>
                <p class="text-xs text-slate-400">Rincian pendapatan usaha, penjualan ampas, rosok, dan bunga bank</p>
              </div>
            </div>
            <div class="text-right">
              <span class="text-[10px] uppercase font-bold text-slate-400">Total Pendapatan</span>
              <p class="text-lg font-black text-emerald-400">{{ formatCurrency(cashierReport.income.total_revenue) }}</p>
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse text-sm">
              <tbody class="divide-y divide-slate-100 font-medium">
                <!-- Penjualan Utama -->
                <tr
                  class="bg-slate-50/70 hover:bg-emerald-50/60 cursor-pointer transition-colors group"
                  @click="openIncomeDetail('PRODUCT_SALE', '1.1 Pendapatan Penjualan Produk Pabrik (Faktur / Telemarketing)')"
                  title="Klik untuk melihat rincian faktur & pesanan"
                >
                  <td class="px-6 py-4 font-black text-slate-800 flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <span>1.1 Pendapatan Penjualan Produk Pabrik (Faktur / Telemarketing)</span>
                    </div>
                    <span class="text-[11px] font-bold text-emerald-700 bg-emerald-100/80 px-2 py-0.5 rounded-md flex items-center gap-1 opacity-80 group-hover:opacity-100 transition shadow-sm">
                      <Eye class="w-3.5 h-3.5" />
                      Rincian Faktur
                    </span>
                  </td>
                  <td class="px-6 py-4 text-right font-black text-slate-900 group-hover:text-emerald-700 transition">{{ formatCurrency(cashierReport.income.product_sales) }}</td>
                </tr>

                <!-- Pendapatan Bunga Bank -->
                <tr
                  class="hover:bg-emerald-50/50 cursor-pointer transition-colors group"
                  @click="openIncomeDetail('INTEREST_INCOME', '1.2 Pendapatan Bunga Bank')"
                  title="Klik untuk melihat rincian bunga bank"
                >
                  <td class="px-6 py-3.5 pl-10 text-slate-700 flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
                      <span>1.2 Pendapatan Bunga Bank</span>
                      <span class="text-[10px] bg-emerald-50 text-emerald-700 font-bold px-2 py-0.5 rounded-full border border-emerald-200">Realtime Kas Masuk</span>
                    </div>
                    <span class="text-[10px] font-bold text-slate-500 group-hover:text-emerald-700 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition mr-4">
                      <Eye class="w-3 h-3" />
                      Lihat Rincian
                    </span>
                  </td>
                  <td class="px-6 py-3.5 text-right font-bold text-slate-800 group-hover:text-emerald-700 transition">{{ formatCurrency(cashierReport.income.interest_income) }}</td>
                </tr>

                <!-- Pendapatan Ampas -->
                <tr
                  class="hover:bg-amber-50/50 cursor-pointer transition-colors group"
                  @click="openIncomeDetail('AMPAS_SALE', '1.3 Pendapatan Ampas (Penjualan Ampas Pabrik)')"
                  title="Klik untuk melihat rincian kas masuk ampas"
                >
                  <td class="px-6 py-3.5 pl-10 text-slate-700 flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <span class="w-2 h-2 rounded-full bg-amber-500"></span>
                      <span class="font-semibold text-slate-900">1.3 Pendapatan Ampas (Penjualan Ampas Pabrik)</span>
                      <span class="text-[10px] bg-amber-50 text-amber-700 font-bold px-2 py-0.5 rounded-full border border-amber-200">Realtime Kas Masuk</span>
                    </div>
                    <span class="text-[10px] font-bold text-slate-500 group-hover:text-amber-700 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition mr-4">
                      <Eye class="w-3 h-3" />
                      Lihat Rincian
                    </span>
                  </td>
                  <td class="px-6 py-3.5 text-right font-bold text-slate-800 group-hover:text-amber-700 transition">{{ formatCurrency(cashierReport.income.ampas_income) }}</td>
                </tr>

                <!-- Penjualan Rosok -->
                <tr
                  class="hover:bg-blue-50/50 cursor-pointer transition-colors group"
                  @click="openIncomeDetail('ROSOK_SALE', '1.4 Penjualan Rosok / Afval (Jurigen, Drum, dll)')"
                  title="Klik untuk melihat rincian kas masuk rosok"
                >
                  <td class="px-6 py-3.5 pl-10 text-slate-700 flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <span class="w-2 h-2 rounded-full bg-blue-500"></span>
                      <span class="font-semibold text-slate-900">1.4 Penjualan Rosok / Afval (Jurigen, Drum, dll)</span>
                    </div>
                    <span class="text-[10px] font-bold text-slate-500 group-hover:text-blue-700 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition mr-4">
                      <Eye class="w-3 h-3" />
                      Lihat Rincian
                    </span>
                  </td>
                  <td class="px-6 py-3.5 text-right font-bold text-slate-800 group-hover:text-blue-700 transition">{{ formatCurrency(cashierReport.income.rosok_income) }}</td>
                </tr>

                <!-- Klaim / Penalty -->
                <tr
                  class="hover:bg-purple-50/50 cursor-pointer transition-colors group"
                  @click="openIncomeDetail('CLAIM_PENALTY', '1.5 Penerimaan Klaim / Denda Penalty')"
                  title="Klik untuk melihat rincian kas masuk klaim/denda"
                >
                  <td class="px-6 py-3.5 pl-10 text-slate-700 flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <span class="w-2 h-2 rounded-full bg-purple-500"></span>
                      <span>1.5 Penerimaan Klaim / Denda Penalty</span>
                    </div>
                    <span class="text-[10px] font-bold text-slate-500 group-hover:text-purple-700 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition mr-4">
                      <Eye class="w-3 h-3" />
                      Lihat Rincian
                    </span>
                  </td>
                  <td class="px-6 py-3.5 text-right font-bold text-slate-800 group-hover:text-purple-700 transition">{{ formatCurrency(cashierReport.income.claim_penalty) }}</td>
                </tr>

                <!-- Pendapatan Lain-lain -->
                <tr
                  class="hover:bg-slate-100/80 cursor-pointer transition-colors group"
                  @click="openIncomeDetail('OTHER', '1.6 Pendapatan Lain-Lain Umum')"
                  title="Klik untuk melihat rincian pendapatan lain"
                >
                  <td class="px-6 py-3.5 pl-10 text-slate-700 flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <span class="w-2 h-2 rounded-full bg-slate-400"></span>
                      <span>1.6 Pendapatan Lain-Lain Umum</span>
                    </div>
                    <span class="text-[10px] font-bold text-slate-500 group-hover:text-slate-800 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition mr-4">
                      <Eye class="w-3 h-3" />
                      Lihat Rincian
                    </span>
                  </td>
                  <td class="px-6 py-3.5 text-right font-bold text-slate-800 group-hover:text-slate-900 transition">{{ formatCurrency(cashierReport.income.other_income) }}</td>
                </tr>

                <!-- Subtotal Pendapatan Lain-lain -->
                <tr class="bg-emerald-50/50 font-black text-slate-900 border-t border-emerald-200">
                  <td class="px-6 py-3.5 pl-8 text-emerald-900">SUBTOTAL PENDAPATAN LAIN-LAIN</td>
                  <td class="px-6 py-3.5 text-right text-emerald-800 font-black">{{ formatCurrency(cashierReport.income.total_other_income) }}</td>
                </tr>

                <!-- Total Pendapatan Keseluruhan -->
                <tr class="bg-slate-900 text-white font-black text-sm">
                  <td class="px-6 py-4">TOTAL AKUMULASI PENDAPATAN KASIR</td>
                  <td class="px-6 py-4 text-right text-emerald-400 font-black text-base">{{ formatCurrency(cashierReport.income.total_revenue) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- SECTION 2: DAFTAR AKUN BEBAN BIAYA TERINTEGRASI REALTIME -->
        <div class="bg-white rounded-3xl border border-slate-200/70 shadow-sm overflow-hidden space-y-4 p-6">
          <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <div class="flex items-center gap-2.5">
                <span class="p-2 bg-rose-50 text-rose-600 rounded-xl">
                  <TrendingDown class="w-5 h-5" />
                </span>
                <div>
                  <h3 class="text-base font-black text-slate-900">2. RINCIAN AKUN BEBAN BIAYA PABRIK (REALTIME)</h3>
                  <p class="text-xs text-slate-500">Saldo otomatis bertambah setiap ada pengeluaran kas pabrik terinput</p>
                </div>
              </div>
            </div>

            <div class="flex flex-wrap items-center gap-2">
              <span class="text-xs font-black text-slate-700 bg-slate-100 px-3 py-1.5 rounded-xl border border-slate-200">
                Total Beban: {{ formatCurrency(cashierReport.total_expenses) }}
              </span>
            </div>
          </div>

          <!-- Controls for Expense Table -->
          <div class="grid grid-cols-1 md:grid-cols-12 gap-3 print:hidden">
            <div class="md:col-span-6 relative">
              <input
                v-model="expenseSearch"
                type="text"
                placeholder="Cari akun biaya (contoh: Solar, Muat Kecap, PLN, dll)..."
                class="w-full pl-10 pr-8 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-emerald-500 outline-none transition-all"
              />
              <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
              <button
                v-if="expenseSearch"
                @click="expenseSearch = ''"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
              >
                <X class="w-3.5 h-3.5" />
              </button>
            </div>

            <div class="md:col-span-6 flex flex-wrap gap-1.5">
              <button
                @click="selectedExpenseGroup = 'ALL'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedExpenseGroup === 'ALL' ? 'bg-slate-900 text-white shadow-sm' : 'bg-slate-100 hover:bg-slate-200 text-slate-600'"
              >
                Semua Akun
              </button>
              <button
                @click="selectedExpenseGroup = 'LABOR'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedExpenseGroup === 'LABOR' ? 'bg-indigo-600 text-white shadow-sm' : 'bg-indigo-50 hover:bg-indigo-100 text-indigo-700'"
              >
                Tenaga Kerja / Borongan
              </button>
              <button
                @click="selectedExpenseGroup = 'UTILITIES'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedExpenseGroup === 'UTILITIES' ? 'bg-amber-600 text-white shadow-sm' : 'bg-amber-50 hover:bg-amber-100 text-amber-700'"
              >
                Utilitas & Produksi
              </button>
              <button
                @click="selectedExpenseGroup = 'FLEET'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedExpenseGroup === 'FLEET' ? 'bg-blue-600 text-white shadow-sm' : 'bg-blue-50 hover:bg-blue-100 text-blue-700'"
              >
                Armada & Ekspedisi
              </button>
              <button
                @click="selectedExpenseGroup = 'TAX'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedExpenseGroup === 'TAX' ? 'bg-rose-600 text-white shadow-sm' : 'bg-rose-50 hover:bg-rose-100 text-rose-700'"
              >
                Pajak & Legal
              </button>
              <button
                @click="selectedExpenseGroup = 'GENERAL'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedExpenseGroup === 'GENERAL' ? 'bg-slate-600 text-white shadow-sm' : 'bg-slate-100 hover:bg-slate-200 text-slate-700'"
              >
                Umum & Operasional
              </button>
            </div>
          </div>

          <!-- Expense Accounts Table -->
          <div class="border border-slate-200/80 rounded-2xl overflow-hidden">
            <div class="overflow-x-auto max-h-[600px] custom-scrollbar">
              <table class="w-full text-left border-collapse text-xs">
                <thead class="sticky top-0 bg-slate-900 text-white uppercase text-[10px] font-black tracking-wider z-10">
                  <tr>
                    <th class="px-4 py-3.5 w-12 text-center">No</th>
                    <th class="px-4 py-3.5">Nama Akun Biaya</th>
                    <th class="px-4 py-3.5">Grup Klasifikasi</th>
                    <th class="px-4 py-3.5 text-right">Nominal Saldo (Rp)</th>
                    <th class="px-4 py-3.5 text-right w-24">Proporsi (%)</th>
                    <th class="px-4 py-3.5 text-center w-24">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 font-medium">
                  <tr v-if="filteredExpenses.length === 0">
                    <td colspan="6" class="px-6 py-8 text-center text-slate-400 font-bold">
                      Tidak ada akun biaya yang sesuai dengan pencarian atau filter.
                    </td>
                  </tr>
                  <tr
                    v-for="(item, idx) in filteredExpenses"
                    :key="item.category_key"
                    @click="openExpenseDetail(item.category_key, item.name)"
                    class="hover:bg-rose-50/40 transition-colors cursor-pointer group"
                    :class="item.amount > 0 ? 'bg-white' : 'opacity-60'"
                    title="Klik untuk melihat rincian riwayat pengeluaran kas"
                  >
                    <td class="px-4 py-3 text-center text-slate-400 font-bold">{{ idx + 1 }}</td>
                    <td class="px-4 py-3">
                      <div class="font-bold text-slate-800 group-hover:text-rose-700 transition">{{ item.name }}</div>
                      <div class="text-[10px] text-slate-400 font-mono">{{ item.category_key }}</div>
                    </td>
                    <td class="px-4 py-3">
                      <span
                        class="px-2 py-0.5 rounded-md text-[10px] font-bold inline-block"
                        :class="{
                          'bg-indigo-50 text-indigo-700 border border-indigo-200': item.group === 'LABOR',
                          'bg-amber-50 text-amber-700 border border-amber-200': item.group === 'UTILITIES',
                          'bg-blue-50 text-blue-700 border border-blue-200': item.group === 'FLEET',
                          'bg-rose-50 text-rose-700 border border-rose-200': item.group === 'TAX',
                          'bg-slate-100 text-slate-700 border border-slate-200': item.group === 'GENERAL',
                        }"
                      >
                        {{ getGroupLabel(item.group) }}
                      </span>
                    </td>
                    <td class="px-4 py-3 text-right font-black" :class="item.amount > 0 ? 'text-rose-600' : 'text-slate-400'">
                      {{ formatCurrency(item.amount) }}
                    </td>
                    <td class="px-4 py-3 text-right font-bold text-slate-500">
                      {{ cashierReport.income.total_revenue > 0 ? ((item.amount / cashierReport.income.total_revenue) * 100).toFixed(2) : '0.00' }}%
                    </td>
                    <td class="px-4 py-3 text-center">
                      <button
                        @click.stop="openExpenseDetail(item.category_key, item.name)"
                        class="inline-flex items-center gap-1 text-[11px] font-bold text-rose-700 hover:text-rose-900 bg-rose-50 hover:bg-rose-100 px-2.5 py-1 rounded-lg border border-rose-200 transition shadow-sm"
                        title="Lihat riwayat pengeluaran kas"
                      >
                        <Eye class="w-3.5 h-3.5" />
                        <span>Rincian</span>
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- SECTION 3: TABEL 28 ENTITAS PIUTANG CABANG, BOOSTER, MARKETPLACE, PROYEK -->
        <div class="bg-white rounded-3xl border border-slate-200/70 shadow-sm overflow-hidden space-y-4 p-6">
          <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div class="flex items-center gap-2.5">
              <span class="p-2 bg-amber-50 text-amber-600 rounded-xl">
                <Building class="w-5 h-5" />
              </span>
              <div>
                <h3 class="text-base font-black text-slate-900">3. MONITORING 28 ENTITAS PIUTANG CABANG & PROYEK</h3>
                <p class="text-xs text-slate-500">Rincian piutang cabang PT, booster, marketplace, karyawan, distributor & pos proyek</p>
              </div>
            </div>

            <!-- Mini Summary Badges -->
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-xs font-black text-amber-900 bg-amber-50 px-3 py-1.5 rounded-xl border border-amber-200">
                Total Piutang Berjalan: {{ formatCurrency(totalActiveReceivables) }}
              </span>
            </div>
          </div>

          <!-- Search & Category Filters for Receivables -->
          <div class="grid grid-cols-1 md:grid-cols-12 gap-3 print:hidden">
            <div class="md:col-span-6 relative">
              <input
                v-model="receivableSearch"
                type="text"
                placeholder="Cari entitas (contoh: Cabang, Distributor, Marketplace)..."
                class="w-full pl-10 pr-8 py-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-amber-500 outline-none transition-all"
              />
              <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
              <button
                v-if="receivableSearch"
                @click="receivableSearch = ''"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
              >
                <X class="w-3.5 h-3.5" />
              </button>
            </div>

            <div class="md:col-span-6 flex flex-wrap gap-1.5">
              <button
                @click="selectedReceivableCategory = 'ALL'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedReceivableCategory === 'ALL' ? 'bg-slate-900 text-white shadow-sm' : 'bg-slate-100 hover:bg-slate-200 text-slate-600'"
              >
                Semua (28)
              </button>
              <button
                @click="selectedReceivableCategory = 'CABANG'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedReceivableCategory === 'CABANG' ? 'bg-amber-600 text-white shadow-sm' : 'bg-amber-50 hover:bg-amber-100 text-amber-700'"
              >
                Cabang PT (13)
              </button>
              <button
                @click="selectedReceivableCategory = 'BOOSTER'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedReceivableCategory === 'BOOSTER' ? 'bg-indigo-600 text-white shadow-sm' : 'bg-indigo-50 hover:bg-indigo-100 text-indigo-700'"
              >
                Booster & Distributor (5)
              </button>
              <button
                @click="selectedReceivableCategory = 'MARKETPLACE'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedReceivableCategory === 'MARKETPLACE' ? 'bg-emerald-600 text-white shadow-sm' : 'bg-emerald-50 hover:bg-emerald-100 text-emerald-700'"
              >
                Marketplace (3)
              </button>
              <button
                @click="selectedReceivableCategory = 'KARYAWAN'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedReceivableCategory === 'KARYAWAN' ? 'bg-purple-600 text-white shadow-sm' : 'bg-purple-50 hover:bg-purple-100 text-purple-700'"
              >
                Kasbon Karyawan (2)
              </button>
              <button
                @click="selectedReceivableCategory = 'PROYEK'"
                class="px-3 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer"
                :class="selectedReceivableCategory === 'PROYEK' ? 'bg-rose-600 text-white shadow-sm' : 'bg-rose-50 hover:bg-rose-100 text-rose-700'"
              >
                Pos Proyek (5)
              </button>
            </div>
          </div>

          <!-- Receivables 28 Matrix Table -->
          <div class="border border-slate-200/80 rounded-2xl overflow-hidden">
            <div class="overflow-x-auto max-h-[600px] custom-scrollbar">
              <table class="w-full text-left border-collapse text-xs">
                <thead class="sticky top-0 bg-slate-900 text-white uppercase text-[10px] font-black tracking-wider z-10">
                  <tr>
                    <th class="px-4 py-3.5 w-12 text-center">No</th>
                    <th class="px-4 py-3.5">Kode</th>
                    <th class="px-4 py-3.5">Nama Entitas Piutang</th>
                    <th class="px-4 py-3.5">Kategori</th>
                    <th class="px-4 py-3.5 text-right">Saldo Awal</th>
                    <th class="px-4 py-3.5 text-right">Penambahan (Debet)</th>
                    <th class="px-4 py-3.5 text-right">Pelunasan (Kredit)</th>
                    <th class="px-4 py-3.5 text-right">Saldo Akhir</th>
                    <th class="px-4 py-3.5 text-center">Status</th>
                    <th class="px-4 py-3.5 text-center w-24">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 font-medium">
                  <tr v-if="filteredReceivables.length === 0">
                    <td colspan="10" class="px-6 py-8 text-center text-slate-400 font-bold">
                      Tidak ada entitas piutang yang sesuai dengan kriteria pencarian.
                    </td>
                  </tr>
                  <tr
                    v-for="(item, idx) in filteredReceivables"
                    :key="item.code"
                    @click="openReceivableDetail(item.code, item.name)"
                    class="hover:bg-amber-50/40 transition-colors cursor-pointer group"
                    title="Klik untuk melihat rincian mutasi piutang entitas"
                  >
                    <td class="px-4 py-3 text-center text-slate-400 font-bold">{{ idx + 1 }}</td>
                    <td class="px-4 py-3 font-mono font-bold text-slate-600">{{ item.code }}</td>
                    <td class="px-4 py-3 font-black text-slate-800 group-hover:text-amber-800 transition">{{ item.name }}</td>
                    <td class="px-4 py-3">
                      <span
                        class="px-2 py-0.5 rounded-md text-[10px] font-bold inline-block"
                        :class="{
                          'bg-amber-50 text-amber-700 border border-amber-200': item.category === 'CABANG',
                          'bg-indigo-50 text-indigo-700 border border-indigo-200': item.category === 'BOOSTER',
                          'bg-emerald-50 text-emerald-700 border border-emerald-200': item.category === 'MARKETPLACE',
                          'bg-purple-50 text-purple-700 border border-purple-200': item.category === 'KARYAWAN',
                          'bg-rose-50 text-rose-700 border border-rose-200': item.category === 'PROYEK',
                        }"
                      >
                        {{ item.category }}
                      </span>
                    </td>
                    <td class="px-4 py-3 text-right text-slate-500 font-bold">{{ formatCurrency(item.initial_balance) }}</td>
                    <td class="px-4 py-3 text-right text-rose-600 font-bold">{{ formatCurrency(item.period_addition) }}</td>
                    <td class="px-4 py-3 text-right text-emerald-600 font-bold">{{ formatCurrency(item.period_payment) }}</td>
                    <td class="px-4 py-3 text-right font-black" :class="item.ending_balance > 0 ? 'text-amber-700' : 'text-slate-800'">
                      {{ formatCurrency(item.ending_balance) }}
                    </td>
                    <td class="px-4 py-3 text-center">
                      <span
                        v-if="item.ending_balance <= 0"
                        class="inline-flex items-center gap-1 text-[10px] font-black bg-emerald-50 text-emerald-700 px-2.5 py-0.5 rounded-full border border-emerald-200"
                      >
                        <CheckCircle2 class="w-3 h-3 text-emerald-600" />
                        Lunas
                      </span>
                      <span
                        v-else
                        class="inline-flex items-center gap-1 text-[10px] font-black bg-amber-50 text-amber-700 px-2.5 py-0.5 rounded-full border border-amber-200"
                      >
                        <BadgeAlert class="w-3 h-3 text-amber-600" />
                        Berjalan
                      </span>
                    </td>
                    <td class="px-4 py-3 text-center">
                      <button
                        @click.stop="openReceivableDetail(item.code, item.name)"
                        class="inline-flex items-center gap-1 text-[11px] font-bold text-amber-800 hover:text-amber-950 bg-amber-50 hover:bg-amber-100 px-2.5 py-1 rounded-lg border border-amber-200 transition shadow-sm"
                        title="Lihat riwayat mutasi piutang"
                      >
                        <Eye class="w-3.5 h-3.5" />
                        <span>Mutasi</span>
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Signature Block (Visible for Print) -->
        <div class="hidden print:grid grid-cols-3 gap-8 pt-12 text-center text-xs font-bold text-slate-800">
          <div>
            <p>Dibuat Oleh,</p>
            <div class="h-20"></div>
            <p class="border-t border-slate-900 pt-1 font-black">Kasir Pabrik (Linda)</p>
          </div>
          <div>
            <p>Diperiksa Oleh,</p>
            <div class="h-20"></div>
            <p class="border-t border-slate-900 pt-1 font-black">Finance & Accounting</p>
          </div>
          <div>
            <p>Disetujui Oleh,</p>
            <div class="h-20"></div>
            <p class="border-t border-slate-900 pt-1 font-black">Pimpinan / Direktur Utama</p>
          </div>
        </div>
      </template>

      <!-- ================================================================= -->
      <!-- TAB 2: EXECUTIVE MULTI-PT COMPARATIVE VIEW                       -->
      <!-- ================================================================= -->
      <template v-else-if="activeMainTab === 'executive'">
        <!-- Existing Executive multi-PT view -->
        <div
          v-if="profitLossList.length === 0"
          class="bg-white rounded-3xl border border-slate-200/60 p-16 text-center text-slate-400 shadow-sm"
        >
          <Building2 class="w-16 h-16 mx-auto mb-4 text-slate-300 opacity-50" />
          <h3 class="text-lg font-bold text-slate-700">Data Laba Rugi Eksekutif Kosong</h3>
          <p class="text-sm text-slate-500 mt-1">
            Tidak ada transaksi keuangan atau payroll tercatat di periode {{ selectedMonthName }} {{ selectedYear }}.
          </p>
        </div>

        <div v-else class="space-y-8">
          <div class="bg-white rounded-3xl border border-slate-200/70 shadow-sm overflow-hidden">
            <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Layers class="w-5 h-5 text-primary" />
                <h3 class="text-base font-black text-slate-800">Perbandingan Komparatif Antar PT</h3>
              </div>
              <span class="text-xs font-bold text-slate-400">Periode: {{ activePeriodLabel }}</span>
            </div>

            <div class="overflow-x-auto">
              <table class="w-full text-left border-collapse min-w-[850px]">
                <thead>
                  <tr class="bg-slate-900 text-white uppercase text-[11px] font-black tracking-wider">
                    <th class="px-6 py-5 border-r border-slate-800">Komponen Laba Rugi (IDR)</th>
                    <th
                      v-for="pt in profitLossList"
                      :key="pt.company_id"
                      class="px-6 py-5 text-right border-r border-slate-800 last:border-r-0"
                    >
                      {{ pt.company_name }}
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 text-slate-700 font-medium text-sm">
                  <tr class="bg-slate-50/70 hover:bg-slate-50">
                    <td class="px-6 py-4 font-black text-slate-900 border-r border-slate-100 flex items-center gap-2">
                      <TrendingUp class="w-4 h-4 text-emerald-500" />
                      PENDAPATAN USAHA (REVENUE)
                    </td>
                    <td
                      v-for="pt in profitLossList"
                      :key="pt.company_id"
                      class="px-6 py-4 text-right font-black text-slate-800 border-r border-slate-100 last:border-r-0"
                    >
                      {{ formatCurrency(pt.total_revenue) }}
                    </td>
                  </tr>

                  <tr class="hover:bg-slate-50">
                    <td class="px-6 py-3.5 text-slate-500 border-r border-slate-100 pl-10">
                      Harga Pokok Penjualan (HPP / COGS)
                    </td>
                    <td
                      v-for="pt in profitLossList"
                      :key="pt.company_id"
                      class="px-6 py-3.5 text-right text-slate-600 border-r border-slate-100 last:border-r-0"
                    >
                      {{ formatCurrency(pt.total_cogs || pt.total_revenue * 0.7) }}
                    </td>
                  </tr>

                  <tr class="bg-slate-100/40 font-bold">
                    <td class="px-6 py-3.5 text-slate-900 border-r border-slate-100 pl-6">
                      LABA KOTOR (GROSS PROFIT)
                    </td>
                    <td
                      v-for="pt in profitLossList"
                      :key="pt.company_id"
                      class="px-6 py-3.5 text-right text-slate-900 font-black border-r border-slate-100 last:border-r-0"
                    >
                      {{ formatCurrency(pt.gross_profit || pt.total_revenue * 0.3) }}
                    </td>
                  </tr>

                  <tr class="hover:bg-slate-50">
                    <td class="px-6 py-3.5 text-slate-500 border-r border-slate-100 pl-10 flex items-center gap-2">
                      <TrendingDown class="w-3.5 h-3.5 text-primary" />
                      Beban Gaji Karyawan (Payroll)
                    </td>
                    <td
                      v-for="pt in profitLossList"
                      :key="pt.company_id"
                      class="px-6 py-3.5 text-right text-slate-600 border-r border-slate-100 last:border-r-0"
                    >
                      {{ formatCurrency(pt.total_payroll) }}
                    </td>
                  </tr>

                  <tr class="hover:bg-slate-50">
                    <td class="px-6 py-3.5 text-slate-500 border-r border-slate-100 pl-10 flex items-center gap-2">
                      <TrendingDown class="w-3.5 h-3.5 text-primary" />
                      Beban Operasional / Kas Pabrik
                    </td>
                    <td
                      v-for="pt in profitLossList"
                      :key="pt.company_id"
                      class="px-6 py-3.5 text-right text-slate-600 border-r border-slate-100 last:border-r-0"
                    >
                      {{ formatCurrency(pt.total_expenses) }}
                    </td>
                  </tr>

                  <tr class="bg-rose-50/20 text-slate-900 font-bold">
                    <td class="px-6 py-3.5 border-r border-slate-100 pl-8">
                      TOTAL BEBAN OPERASIONAL
                    </td>
                    <td
                      v-for="pt in profitLossList"
                      :key="pt.company_id"
                      class="px-6 py-3.5 text-right text-rose-600 font-black border-r border-slate-100 last:border-r-0"
                    >
                      {{ formatCurrency(pt.total_payroll + pt.total_expenses) }}
                    </td>
                  </tr>

                  <tr class="bg-primary/5 font-black text-primary text-base">
                    <td class="px-6 py-4.5 border-r border-slate-200/50 flex items-center gap-2">
                      <Sparkles class="w-5 h-5 text-primary animate-pulse" />
                      LABA BERSIH (NET PROFIT)
                    </td>
                    <td
                      v-for="pt in profitLossList"
                      :key="pt.company_id"
                      class="px-6 py-4.5 text-right border-r border-slate-200/50 last:border-r-0"
                      :class="pt.net_profit >= 0 ? 'text-primary' : 'text-rose-600'"
                    >
                      {{ formatCurrency(pt.net_profit) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- MODAL DRILLDOWN RINCIAN TRANSAKSI PENDAPATAN -->
    <div
      v-if="showIncomeDetailModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
    >
      <div
        class="bg-white w-full max-w-5xl rounded-3xl shadow-2xl border border-slate-200 overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-200"
      >
        <!-- Modal Header -->
        <div class="bg-slate-900 text-white p-6 flex items-start justify-between">
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <span class="p-1.5 bg-emerald-500/20 text-emerald-400 rounded-lg">
                <Receipt class="w-5 h-5" />
              </span>
              <h3 class="text-lg font-black tracking-tight">
                {{ incomeDetailData?.income_title || 'Rincian Transaksi Pendapatan' }}
              </h3>
            </div>
            <div class="flex flex-wrap items-center gap-2 pt-1 text-xs text-slate-300">
              <span class="bg-slate-800 border border-slate-700 px-2.5 py-0.5 rounded-full font-semibold flex items-center gap-1.5">
                <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
                Periode: {{ activePeriodLabel }}
              </span>
              <span class="bg-slate-800 border border-slate-700 px-2.5 py-0.5 rounded-full font-semibold flex items-center gap-1.5">
                <MapPin class="w-3 h-3 text-slate-400" />
                {{ currentBranchName }}
              </span>
            </div>
          </div>

          <button
            @click="showIncomeDetailModal = false"
            class="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Filter & Summary Bar -->
        <div class="p-5 bg-slate-50 border-b border-slate-200/80 flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div class="relative flex-1 max-w-md">
            <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              v-model="incomeDetailSearch"
              type="text"
              placeholder="Cari no. faktur, toko, kas, atau keterangan..."
              class="w-full pl-9 pr-8 py-2 text-xs bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 transition shadow-sm font-medium"
            />
            <button
              v-if="incomeDetailSearch"
              @click="incomeDetailSearch = ''"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>

          <div class="flex items-center gap-4 text-xs font-bold">
            <div class="bg-white border border-slate-200 px-3.5 py-2 rounded-xl shadow-sm text-right">
              <span class="text-[10px] text-slate-500 uppercase block font-semibold">Total Transaksi</span>
              <span class="text-sm font-black text-slate-800">{{ filteredIncomeDetailItems.length }} Item</span>
            </div>
            <div class="bg-emerald-50 border border-emerald-200 px-3.5 py-2 rounded-xl shadow-sm text-right">
              <span class="text-[10px] text-emerald-700 uppercase block font-semibold">Total Akumulasi</span>
              <span class="text-sm font-black text-emerald-800">{{ formatCurrency(incomeDetailTotalFiltered) }}</span>
            </div>
          </div>
        </div>

        <!-- Modal Body Table -->
        <div class="overflow-y-auto flex-1 p-5">
          <!-- Loading State -->
          <div v-if="isLoadingIncomeDetail" class="py-16 text-center space-y-3">
            <Loader2 class="w-8 h-8 animate-spin text-emerald-600 mx-auto" />
            <p class="text-xs font-bold text-slate-500">Memuat rincian transaksi...</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="filteredIncomeDetailItems.length === 0" class="py-16 text-center space-y-2">
            <div class="w-12 h-12 rounded-2xl bg-slate-100 flex items-center justify-center mx-auto text-slate-400">
              <Receipt class="w-6 h-6" />
            </div>
            <h4 class="text-sm font-bold text-slate-800">Tidak ada transaksi ditemukan</h4>
            <p class="text-xs text-slate-500 max-w-sm mx-auto">
              {{ incomeDetailSearch ? 'Tidak ada transaksi yang cocok dengan kata kunci pencarian.' : 'Belum ada transaksi pendapatan pada periode dan filter cabang yang dipilih.' }}
            </p>
          </div>

          <!-- Data Table -->
          <div v-else class="overflow-x-auto rounded-2xl border border-slate-200 shadow-sm">
            <table class="w-full text-left border-collapse text-xs">
              <thead class="bg-slate-900 text-white font-bold text-[11px] uppercase tracking-wider">
                <tr>
                  <th class="px-4 py-3">Tanggal</th>
                  <th class="px-4 py-3">No. Referensi</th>
                  <th class="px-4 py-3">Pihak / Toko / Kas</th>
                  <th class="px-4 py-3">Keterangan / Metode</th>
                  <th class="px-4 py-3 text-center">Status</th>
                  <th class="px-4 py-3 text-right">Nominal</th>
                  <th class="px-4 py-3 text-center">Bukti</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 font-medium">
                <tr
                  v-for="item in filteredIncomeDetailItems"
                  :key="item.id"
                  class="hover:bg-slate-50 transition"
                >
                  <td class="px-4 py-3 whitespace-nowrap text-slate-600 font-semibold">
                    {{ item.transaction_date }}
                  </td>
                  <td class="px-4 py-3 whitespace-nowrap">
                    <div class="flex items-center gap-1.5">
                      <span class="font-bold text-slate-900">{{ item.reference_no }}</span>
                      <span
                        v-if="item.transaction_type === 'SALES_INVOICE'"
                        class="text-[9px] bg-emerald-100 text-emerald-800 font-bold px-1.5 py-0.5 rounded"
                      >
                        INVOICE
                      </span>
                      <span
                        v-else-if="item.transaction_type === 'SALES_ORDER'"
                        class="text-[9px] bg-blue-100 text-blue-800 font-bold px-1.5 py-0.5 rounded"
                      >
                        SO
                      </span>
                      <span
                        v-else
                        class="text-[9px] bg-amber-100 text-amber-800 font-bold px-1.5 py-0.5 rounded"
                      >
                        KAS MASUK
                      </span>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-slate-800 font-bold">
                    {{ item.entity_name }}
                  </td>
                  <td class="px-4 py-3 text-slate-600 max-w-xs truncate">
                    <div>{{ item.description || '-' }}</div>
                    <div v-if="item.payment_method" class="text-[10px] text-slate-400 font-semibold mt-0.5">
                      Metode: {{ item.payment_method }}
                    </div>
                  </td>
                  <td class="px-4 py-3 text-center whitespace-nowrap">
                    <span
                      class="text-[10px] font-black px-2 py-0.5 rounded-full uppercase"
                      :class="
                        item.status === 'VERIFIED' || item.status === 'CONFIRMED' || item.status === 'PAID'
                          ? 'bg-emerald-100 text-emerald-800'
                          : 'bg-amber-100 text-amber-800'
                      "
                    >
                      {{ item.status }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-right font-black text-slate-900 whitespace-nowrap">
                    {{ formatCurrency(item.amount) }}
                  </td>
                  <td class="px-4 py-3 text-center whitespace-nowrap">
                    <a
                      v-if="item.proof_url"
                      :href="item.proof_url"
                      target="_blank"
                      class="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-700 hover:text-emerald-900 bg-emerald-50 hover:bg-emerald-100 px-2 py-1 rounded-lg transition"
                      title="Buka lampiran bukti transaksi"
                    >
                      <ExternalLink class="w-3 h-3" />
                      Bukti
                    </a>
                    <span v-else class="text-slate-300">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 bg-slate-100/70 border-t border-slate-200 flex items-center justify-between text-xs font-semibold text-slate-600">
          <div>
            Menampilkan {{ filteredIncomeDetailItems.length }} dari {{ incomeDetailData?.total_count || 0 }} data transaksi
          </div>
          <button
            @click="showIncomeDetailModal = false"
            class="px-5 py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-xl font-bold transition shadow-sm"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL DRILLDOWN RINCIAN PENGELUARAN BEBAN BIAYA PABRIK -->
    <div
      v-if="showExpenseDetailModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
    >
      <div
        class="bg-white w-full max-w-5xl rounded-3xl shadow-2xl border border-slate-200 overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-200"
      >
        <!-- Modal Header -->
        <div class="bg-slate-900 text-white p-6 flex items-start justify-between">
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <span class="p-1.5 bg-rose-500/20 text-rose-400 rounded-lg">
                <TrendingDown class="w-5 h-5" />
              </span>
              <h3 class="text-lg font-black tracking-tight">
                {{ expenseDetailData?.category_name || 'Rincian Pengeluaran Beban Biaya' }}
              </h3>
              <span
                v-if="expenseDetailData?.group"
                class="ml-2 px-2 py-0.5 rounded-md text-[10px] font-black uppercase bg-slate-800 text-rose-300 border border-rose-500/30"
              >
                {{ getGroupLabel(expenseDetailData.group) }}
              </span>
            </div>
            <div class="flex flex-wrap items-center gap-2 pt-1 text-xs text-slate-300">
              <span class="bg-slate-800 border border-slate-700 px-2.5 py-0.5 rounded-full font-semibold flex items-center gap-1.5">
                <span class="w-2 h-2 rounded-full bg-rose-400"></span>
                Periode: {{ activePeriodLabel }}
              </span>
              <span class="bg-slate-800 border border-slate-700 px-2.5 py-0.5 rounded-full font-semibold flex items-center gap-1.5">
                <MapPin class="w-3 h-3 text-slate-400" />
                {{ currentBranchName }}
              </span>
              <span
                v-if="expenseDetailData?.category_key"
                class="bg-slate-800 border border-slate-700 px-2.5 py-0.5 rounded-full font-mono text-[11px] text-slate-400"
              >
                Kode: {{ expenseDetailData.category_key }}
              </span>
            </div>
          </div>

          <button
            @click="showExpenseDetailModal = false"
            class="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Filter & Summary Bar -->
        <div class="p-5 bg-slate-50 border-b border-slate-200/80 flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div class="relative flex-1 max-w-md">
            <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              v-model="expenseDetailSearch"
              type="text"
              placeholder="Cari no. transaksi, kas, penerima, armada, atau keterangan..."
              class="w-full pl-9 pr-8 py-2 text-xs bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-rose-500 focus:border-rose-500 transition shadow-sm font-medium"
            />
            <button
              v-if="expenseDetailSearch"
              @click="expenseDetailSearch = ''"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>

          <div class="flex items-center gap-4 text-xs font-bold">
            <div class="bg-white border border-slate-200 px-3.5 py-2 rounded-xl shadow-sm text-right">
              <span class="text-[10px] text-slate-500 uppercase block font-semibold">Total Pengeluaran</span>
              <span class="text-sm font-black text-slate-800">{{ filteredExpenseDetailItems.length }} Transaksi</span>
            </div>
            <div class="bg-rose-50 border border-rose-200 px-3.5 py-2 rounded-xl shadow-sm text-right">
              <span class="text-[10px] text-rose-700 uppercase block font-semibold">Total Akumulasi Biaya</span>
              <span class="text-sm font-black text-rose-700">{{ formatCurrency(expenseDetailTotalFiltered) }}</span>
            </div>
          </div>
        </div>

        <!-- Modal Body Table -->
        <div class="overflow-y-auto flex-1 p-5">
          <!-- Loading State -->
          <div v-if="isLoadingExpenseDetail" class="py-16 text-center space-y-3">
            <Loader2 class="w-8 h-8 animate-spin text-rose-600 mx-auto" />
            <p class="text-xs font-bold text-slate-500">Memuat rincian transaksi pengeluaran kas...</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="filteredExpenseDetailItems.length === 0" class="py-16 text-center space-y-2">
            <div class="w-12 h-12 rounded-2xl bg-slate-100 flex items-center justify-center mx-auto text-slate-400">
              <TrendingDown class="w-6 h-6" />
            </div>
            <h4 class="text-sm font-bold text-slate-800">Tidak ada transaksi pengeluaran ditemukan</h4>
            <p class="text-xs text-slate-500 max-w-sm mx-auto">
              {{ expenseDetailSearch ? 'Tidak ada transaksi yang cocok dengan kata kunci pencarian.' : 'Belum ada pengeluaran kas tercatat untuk akun biaya ini pada periode yang dipilih.' }}
            </p>
          </div>

          <!-- Data Table -->
          <div v-else class="overflow-x-auto rounded-2xl border border-slate-200 shadow-sm">
            <table class="w-full text-left border-collapse text-xs">
              <thead class="bg-slate-900 text-white font-bold text-[11px] uppercase tracking-wider">
                <tr>
                  <th class="px-4 py-3">Tanggal</th>
                  <th class="px-4 py-3">No. Referensi</th>
                  <th class="px-4 py-3">Sumber Kas</th>
                  <th class="px-4 py-3">Penerima / Cabang / Armada</th>
                  <th class="px-4 py-3">Keterangan</th>
                  <th class="px-4 py-3 text-center">Status</th>
                  <th class="px-4 py-3 text-right">Nominal</th>
                  <th class="px-4 py-3 text-center">Bukti</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 font-medium">
                <tr
                  v-for="item in filteredExpenseDetailItems"
                  :key="item.id"
                  class="hover:bg-slate-50 transition"
                >
                  <td class="px-4 py-3 whitespace-nowrap text-slate-600 font-semibold">
                    {{ item.transaction_date }}
                  </td>
                  <td class="px-4 py-3 whitespace-nowrap">
                    <span class="font-bold text-slate-900 font-mono">{{ item.reference_no }}</span>
                  </td>
                  <td class="px-4 py-3 text-slate-800 font-bold whitespace-nowrap">
                    {{ item.kas_name }}
                  </td>
                  <td class="px-4 py-3 text-slate-700">
                    {{ item.recipient_name || '-' }}
                  </td>
                  <td class="px-4 py-3 text-slate-600 max-w-xs truncate" :title="item.description">
                    {{ item.description || '-' }}
                  </td>
                  <td class="px-4 py-3 text-center whitespace-nowrap">
                    <span
                      class="text-[10px] font-black px-2 py-0.5 rounded-full uppercase"
                      :class="
                        item.status === 'CONFIRMED' || item.status === 'SETTLED'
                          ? 'bg-emerald-100 text-emerald-800'
                          : 'bg-amber-100 text-amber-800'
                      "
                    >
                      {{ item.status }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-right font-black text-rose-600 whitespace-nowrap">
                    {{ formatCurrency(item.amount) }}
                  </td>
                  <td class="px-4 py-3 text-center whitespace-nowrap">
                    <a
                      v-if="item.proof_url"
                      :href="item.proof_url"
                      target="_blank"
                      class="inline-flex items-center gap-1 text-[11px] font-bold text-rose-700 hover:text-rose-900 bg-rose-50 hover:bg-rose-100 px-2 py-1 rounded-lg transition"
                      title="Buka lampiran bukti transaksi"
                    >
                      <ExternalLink class="w-3 h-3" />
                      Bukti
                    </a>
                    <span v-else class="text-slate-300">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 bg-slate-100/70 border-t border-slate-200 flex items-center justify-between text-xs font-semibold text-slate-600">
          <div>
            Menampilkan {{ filteredExpenseDetailItems.length }} dari {{ expenseDetailData?.total_count || 0 }} data pengeluaran
          </div>
          <button
            @click="showExpenseDetailModal = false"
            class="px-5 py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-xl font-bold transition shadow-sm"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL DRILLDOWN RINCIAN MUTASI ENTITAS PIUTANG -->
    <div
      v-if="showReceivableDetailModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
    >
      <div
        class="bg-white w-full max-w-5xl rounded-3xl shadow-2xl border border-slate-200 overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-200"
      >
        <!-- Modal Header -->
        <div class="bg-slate-900 text-white p-6 flex items-start justify-between">
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <span class="p-1.5 bg-amber-500/20 text-amber-400 rounded-lg">
                <Building class="w-5 h-5" />
              </span>
              <h3 class="text-lg font-black tracking-tight">
                {{ receivableDetailData?.entity_name || 'Rincian Mutasi Entitas Piutang' }}
              </h3>
              <span
                v-if="receivableDetailData?.entity_code"
                class="ml-2 px-2 py-0.5 rounded-md text-[10px] font-black font-mono bg-slate-800 text-amber-300 border border-amber-500/30"
              >
                {{ receivableDetailData.entity_code }}
              </span>
              <span
                v-if="receivableDetailData?.category"
                class="px-2 py-0.5 rounded-md text-[10px] font-black uppercase bg-slate-800 text-slate-300 border border-slate-700"
              >
                {{ receivableDetailData.category }}
              </span>
            </div>
            <div class="flex flex-wrap items-center gap-2 pt-1 text-xs text-slate-300">
              <span class="bg-slate-800 border border-slate-700 px-2.5 py-0.5 rounded-full font-semibold flex items-center gap-1.5">
                <span class="w-2 h-2 rounded-full bg-amber-400"></span>
                Periode: {{ activePeriodLabel }}
              </span>
              <span class="bg-slate-800 border border-slate-700 px-2.5 py-0.5 rounded-full font-semibold flex items-center gap-1.5">
                <MapPin class="w-3 h-3 text-slate-400" />
                {{ currentBranchName }}
              </span>
            </div>
          </div>

          <button
            @click="showReceivableDetailModal = false"
            class="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Receivables Balance Summary Cards Header -->
        <div class="p-5 bg-slate-50 border-b border-slate-200/80 space-y-4">
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
            <div class="bg-white p-3.5 rounded-2xl border border-slate-200 shadow-sm">
              <span class="text-[10px] uppercase font-bold text-slate-400 block">Saldo Awal</span>
              <span class="text-base font-black text-slate-700 mt-1 block">
                {{ formatCurrency(receivableDetailData?.initial_balance) }}
              </span>
            </div>

            <div class="bg-white p-3.5 rounded-2xl border border-rose-200 bg-rose-50/20 shadow-sm">
              <span class="text-[10px] uppercase font-bold text-rose-600 block">Penambahan (Talangan Kas)</span>
              <span class="text-base font-black text-rose-600 mt-1 block">
                {{ formatCurrency(receivableDetailData?.period_addition) }}
              </span>
            </div>

            <div class="bg-white p-3.5 rounded-2xl border border-emerald-200 bg-emerald-50/20 shadow-sm">
              <span class="text-[10px] uppercase font-bold text-emerald-600 block">Pelunasan (Setoran Masuk)</span>
              <span class="text-base font-black text-emerald-600 mt-1 block">
                {{ formatCurrency(receivableDetailData?.period_payment) }}
              </span>
            </div>

            <div
              class="p-3.5 rounded-2xl border shadow-sm flex flex-col justify-between"
              :class="receivableDetailData && receivableDetailData.ending_balance > 0 ? 'bg-amber-50 border-amber-300' : 'bg-emerald-50 border-emerald-300'"
            >
              <div class="flex items-center justify-between">
                <span class="text-[10px] uppercase font-bold text-slate-600">Saldo Akhir</span>
                <span
                  class="text-[9px] font-black uppercase px-2 py-0.5 rounded-full"
                  :class="receivableDetailData && receivableDetailData.ending_balance > 0 ? 'bg-amber-200 text-amber-900' : 'bg-emerald-200 text-emerald-900'"
                >
                  {{ receivableDetailData?.status || 'LUNAS' }}
                </span>
              </div>
              <span
                class="text-base font-black mt-1 block"
                :class="receivableDetailData && receivableDetailData.ending_balance > 0 ? 'text-amber-800' : 'text-emerald-800'"
              >
                {{ formatCurrency(receivableDetailData?.ending_balance) }}
              </span>
            </div>
          </div>

          <!-- Live Search Bar inside Modal -->
          <div class="relative">
            <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              v-model="receivableDetailSearch"
              type="text"
              placeholder="Cari no. bukti, sumber kas, atau keterangan mutasi..."
              class="w-full pl-9 pr-8 py-2 text-xs bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-amber-500 focus:border-amber-500 transition shadow-sm font-medium"
            />
            <button
              v-if="receivableDetailSearch"
              @click="receivableDetailSearch = ''"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- Modal Body Table -->
        <div class="overflow-y-auto flex-1 p-5">
          <!-- Loading State -->
          <div v-if="isLoadingReceivableDetail" class="py-16 text-center space-y-3">
            <Loader2 class="w-8 h-8 animate-spin text-amber-600 mx-auto" />
            <p class="text-xs font-bold text-slate-500">Memuat rincian mutasi piutang...</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="filteredReceivableDetailItems.length === 0" class="py-16 text-center space-y-2">
            <div class="w-12 h-12 rounded-2xl bg-slate-100 flex items-center justify-center mx-auto text-slate-400">
              <Building class="w-6 h-6" />
            </div>
            <h4 class="text-sm font-bold text-slate-800">Tidak ada mutasi transaksi</h4>
            <p class="text-xs text-slate-500 max-w-sm mx-auto">
              {{ receivableDetailSearch ? 'Tidak ada data yang cocok dengan kata kunci pencarian.' : 'Belum ada mutasi talangan atau pelunasan kas untuk entitas ini pada periode yang dipilih.' }}
            </p>
          </div>

          <!-- Data Table -->
          <div v-else class="overflow-x-auto rounded-2xl border border-slate-200 shadow-sm">
            <table class="w-full text-left border-collapse text-xs">
              <thead class="bg-slate-900 text-white font-bold text-[11px] uppercase tracking-wider">
                <tr>
                  <th class="px-4 py-3">Tanggal</th>
                  <th class="px-4 py-3">No. Referensi</th>
                  <th class="px-4 py-3 text-center">Jenis Mutasi</th>
                  <th class="px-4 py-3">Sumber Kas</th>
                  <th class="px-4 py-3">Keterangan</th>
                  <th class="px-4 py-3 text-center">Status</th>
                  <th class="px-4 py-3 text-right">Nominal</th>
                  <th class="px-4 py-3 text-center">Bukti</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 font-medium">
                <tr
                  v-for="item in filteredReceivableDetailItems"
                  :key="item.id"
                  class="hover:bg-slate-50 transition"
                >
                  <td class="px-4 py-3 whitespace-nowrap text-slate-600 font-semibold">
                    {{ item.transaction_date }}
                  </td>
                  <td class="px-4 py-3 whitespace-nowrap font-mono font-bold text-slate-900">
                    {{ item.reference_no }}
                  </td>
                  <td class="px-4 py-3 text-center whitespace-nowrap">
                    <span
                      v-if="item.transaction_type === 'DEBET'"
                      class="inline-flex items-center gap-1 text-[10px] font-black px-2.5 py-0.5 rounded-full bg-rose-100 text-rose-800 border border-rose-200"
                    >
                      <ArrowUpRight class="w-3 h-3" />
                      TALANGAN (DEBET)
                    </span>
                    <span
                      v-else
                      class="inline-flex items-center gap-1 text-[10px] font-black px-2.5 py-0.5 rounded-full bg-emerald-100 text-emerald-800 border border-emerald-200"
                    >
                      <ArrowDownRight class="w-3 h-3" />
                      PELUNASAN (KREDIT)
                    </span>
                  </td>
                  <td class="px-4 py-3 text-slate-800 font-bold whitespace-nowrap">
                    {{ item.kas_name }}
                  </td>
                  <td class="px-4 py-3 text-slate-600 max-w-xs truncate" :title="item.description">
                    {{ item.description || '-' }}
                  </td>
                  <td class="px-4 py-3 text-center whitespace-nowrap">
                    <span
                      class="text-[10px] font-black px-2 py-0.5 rounded-full uppercase"
                      :class="
                        item.status === 'CONFIRMED' || item.status === 'SETTLED'
                          ? 'bg-emerald-100 text-emerald-800'
                          : 'bg-amber-100 text-amber-800'
                      "
                    >
                      {{ item.status }}
                    </span>
                  </td>
                  <td
                    class="px-4 py-3 text-right font-black whitespace-nowrap"
                    :class="item.transaction_type === 'DEBET' ? 'text-rose-600' : 'text-emerald-600'"
                  >
                    {{ item.transaction_type === 'DEBET' ? '+' : '-' }} {{ formatCurrency(item.amount) }}
                  </td>
                  <td class="px-4 py-3 text-center whitespace-nowrap">
                    <a
                      v-if="item.proof_url"
                      :href="item.proof_url"
                      target="_blank"
                      class="inline-flex items-center gap-1 text-[11px] font-bold text-amber-800 hover:text-amber-950 bg-amber-50 hover:bg-amber-100 px-2 py-1 rounded-lg transition"
                      title="Buka lampiran bukti transaksi"
                    >
                      <ExternalLink class="w-3 h-3" />
                      Bukti
                    </a>
                    <span v-else class="text-slate-300">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 bg-slate-100/70 border-t border-slate-200 flex items-center justify-between text-xs font-semibold text-slate-600">
          <div>
            Menampilkan {{ filteredReceivableDetailItems.length }} dari {{ receivableDetailData?.total_count || 0 }} mutasi transaksi
          </div>
          <button
            @click="showReceivableDetailModal = false"
            class="px-5 py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-xl font-bold transition shadow-sm"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@media print {
  aside,
  header,
  nav,
  select,
  button,
  .print\:hidden {
    display: none !important;
  }
  .print\:bg-white {
    background-color: #fff !important;
  }
  .print\:p-0 {
    padding: 0 !important;
  }
  body {
    font-size: 10pt;
  }
}
</style>
