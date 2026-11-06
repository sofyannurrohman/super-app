<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import {
  Landmark,
  FileText,
  Plus,
  RefreshCw,
  Save,
  CheckCircle,
  Building,
  Upload,
  Calendar,
  Wallet,
  ArrowUpRight,
  ArrowDownRight,
  Eye,
  File,
  X,
  Receipt,
  CheckCircle2,
  AlertTriangle,
  Coins,
  DollarSign,
  Edit2,
  Trash2,
  Settings,
  CreditCard,
  Layers,
  ChevronRight,
  ZoomIn,
  ZoomOut,
  RotateCw,
  RotateCcw,
  ExternalLink
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import { useMasterDataStore } from '@/stores/masterData'

const authStore = useAuthStore()
const masterDataStore = useMasterDataStore()

// Top Navigation Tabs
const currentTab = ref<'reports' | 'reimbursements' | 'mutations'>('reports')

// Global PT / Company Filter
const selectedCompanyId = ref<string>('')

// State for Banks & Cash
const banks = ref<any[]>([])
const cashBalance = ref(0)
const pettyCashBalance = ref(0)
const companyBalances = ref<any[]>([])
const isBanksLoading = ref(false)
const showBalanceModal = ref(false)
const selectedBank = ref<any>(null)
const newBalance = ref(0)
const systemBankAccounts = ref<any[]>([])

// State for Bank Management (Add / Edit Bank)
const showAddEditBankModal = ref(false)
const isEditingBank = ref(false)
const isSavingBank = ref(false)
const bankForm = ref({
  id: '',
  company_id: '',
  branch_id: '',
  bank_name: '',
  account_number: '',
  account_holder: '',
  manual_balance: 0,
  notes: '',
  is_active: true
})

// State for Petty Cash Direct Balance Update (Per PT)
const showPettyCashBalanceModal = ref(false)
const pettyCashBalanceForm = ref({
  company_id: '',
  petty_cash_balance: 0,
  notes: ''
})
const isSavingPettyCash = ref(false)

// State for Manual Cash (Kas Besar) Direct Balance Update (Per PT)
const showCashBalanceModal = ref(false)
const cashBalanceForm = ref({
  company_id: '',
  manual_cash_balance: 0,
  notes: ''
})
const isSavingCashBalance = ref(false)

// State for Mutation Editing & Deleting
const editingMutationId = ref<string | null>(null)
const isEditingMutation = ref(false)
const deletingMutationId = ref<string | null>(null)

// State for Reports
const reports = ref<any[]>([])
const isReportsLoading = ref(false)
const showReportModal = ref(false)
const isEditingReport = ref(false)
const editingReportId = ref<string | null>(null)
const autoApprovedCashSales = ref(0)
const autoApprovedTransferSales = ref(0)
const isFetchingSummary = ref(false)
const reportForm = ref({
  branch_id: authStore.user?.branch_id || '',
  company_id: '',
  report_date: new Date().toISOString().split('T')[0],
  total_cash_sales: 0,
  total_transfer_sales: 0,
  allocated_to_petty_cash: 0,
  deposited_amount: 0,
  deposit_destination_type: 'BRANCH_BANK', // 'BRANCH_BANK' | 'FACTORY_KAS'
  target_bank_account_id: '',
  target_kas_pabrik_id: '',
  target_bank_name: '',
  target_account_number: '',
  target_account_name: '',
  notes: '',
  deposit_proof_url: ''
})
const selectedReportFile = ref<File | null>(null)
const isUploadingReport = ref(false)

const handleReportFileChange = (e: any) => {
  if (e.target.files && e.target.files[0]) {
    selectedReportFile.value = e.target.files[0]
  }
}

// State for Reimbursements
const reimbursements = ref<any[]>([])
const isReimbursementsLoading = ref(false)
const reimbursementFilterStatus = ref('APPROVED') // 'ALL' | 'APPROVED' | 'DISBURSED'
const showDisburseModal = ref(false)
const selectedReimbursement = ref<any>(null)
const disburseForm = ref({
  disbursed_via: 'PETTY_CASH', // 'PETTY_CASH' | 'CASH' | 'BANK_TRANSFER'
  company_id: '',
  bank_account_id: '',
  proof_url: '',
  notes: '',
  branch_id: authStore.user?.branch_id || ''
})
const selectedDisburseFile = ref<File | null>(null)
const isDisbursing = ref(false)

const handleDisburseFileChange = (e: any) => {
  if (e.target.files && e.target.files[0]) {
    selectedDisburseFile.value = e.target.files[0]
  }
}

// State for Mutations
const mutations = ref<any[]>([])
const isMutationsLoading = ref(false)
const showMutationModal = ref(false)
const mutationFundTypeFilter = ref('')
const mutationCategoryFilter = ref('')

const currentBranchId = computed(() => {
  const role = authStore.user?.role || ''
  const isBranchBound = role === 'cashier' || role === 'admin_kasir'
  if (isBranchBound && authStore.user?.branch_id && authStore.user.branch_id !== '00000000-0000-0000-0000-000000000000') {
    return authStore.user.branch_id
  }
  const selectedBranches = (masterDataStore.selectedBranchIds || []).filter((id: string) => id !== 'ALL_BRANCHES' && id !== '00000000-0000-0000-0000-000000000000')
  if (selectedBranches.length === 1) {
    return selectedBranches[0]
  }
  if (selectedBranches.length > 1) {
    return ''
  }
  if (masterDataStore.branches && masterDataStore.branches.length > 0) {
    return masterDataStore.branches[0].id
  }
  return ''
})

const currentBranchName = computed(() => {
  const isBranchBound = ['cashier', 'admin_kasir'].includes(authStore.user?.role || '')
  const userBranch = isBranchBound ? (authStore.user?.branch_name || authStore.user?.branch?.name) : ''
  if (userBranch) return userBranch
  return masterDataStore.branches?.find((b: any) => b.id === currentBranchId.value)?.name || 'Cabang aktif'
})

const selectedCompanyName = computed(() => {
  if (!selectedCompanyId.value) return `Semua PT di ${currentBranchName.value}`
  return masterDataStore.companies?.find((c: any) => c.id === selectedCompanyId.value)?.name || 'PT terpilih'
})

// Daftar PT pada halaman laporan mengikuti saldo/aktivitas cabang aktif,
// bukan seluruh master perusahaan global.
const branchCompanies = computed(() => {
  const ids = new Set(
    companyBalances.value
      .map((item: any) => item.company_id)
      .filter(Boolean)
  )
  if (ids.size === 0) return masterDataStore.companies || []
  return (masterDataStore.companies || []).filter((company: any) => ids.has(company.id))
})

const mutationForm = ref({
  branch_id: '',
  company_id: '',
  destination_company_id: '',
  mutation_type: 'IN',
  fund_type: 'PETTY_CASH', // 'PETTY_CASH' | 'CASH' | 'BANK'
  category: 'OPERATIONAL_EXPENSE',
  bank_account_id: '',
  is_setor_pusat: false,
  target_kas_pabrik_id: '',
  amount: 0,
  notes: '',
  proof_url: '',
  fund_origin: '',
  mutation_date: new Date().toISOString().split('T')[0]
})
const availableInvoicePayments = ref<any[]>([])
const availableInvoicePaymentsError = ref('')
const invoicePaymentSearch = ref('')
const invoiceDateGroupMode = ref<'payment_date' | 'transaction_date'>('payment_date')
const activeInvoiceDateFilter = ref('')
const collapsedInvoiceDates = ref<Record<string, boolean>>({})
const isLoadingInvoicePayments = ref(false)
const selectedInvoicePaymentCompanies = ref<Record<string, string>>({})
const selectedInvoicePaymentReceipts = ref<Record<string, string>>({})
let availableInvoiceRequestId = 0
const selectedInvoicePayments = ref<Record<string, number>>({})
const isFactoryInvoiceDeposit = computed(() => mutationForm.value.category === 'FACTORY_INVOICE_DEPOSIT')
const isFactoryDeposit = computed(() => ['FACTORY_INVOICE_DEPOSIT', 'FACTORY_NON_INVOICE_DEPOSIT'].includes(mutationForm.value.category))
const isFactoryDepositItem = (item: any) => ['FACTORY_INVOICE_DEPOSIT', 'FACTORY_NON_INVOICE_DEPOSIT'].includes(item.category)
const extraDepositAmount = ref(0)
const isBranchDepositReviewer = computed(() => ['admin_kasir', 'manager', 'superadmin'].includes(authStore.user?.role || ''))
const selectedInvoicePaymentTotal = computed(() => Object.values(selectedInvoicePayments.value).reduce((sum, amount) => sum + Number(amount || 0), 0))
const selectedInvoiceCount = computed(() => Object.keys(selectedInvoicePayments.value).length)
const mutationInvoiceLabels = (item: any) => (item.invoice_items || [])
  .map((row: any) => row.sales_payment?.sales_transaction?.receipt_no || row.sales_transaction?.receipt_no)
  .filter(Boolean)
  .join(', ')
const mutationInvoiceTotal = (item: any) => (item.invoice_items || []).reduce((total: number, row: any) => total + Number(row.amount || 0), 0)
const mutationNonInvoiceAmount = (item: any) => Math.max(0, Number(item.amount || 0) - mutationInvoiceTotal(item))
const selectedFile = ref<File | null>(null)
const isUploading = ref(false)

const toDateKey = (dateStr?: string) => {
  if (!dateStr) return 'tanpa-tanggal'
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return String(dateStr).split('T')[0] || 'tanpa-tanggal'
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatShortDate = (dateStr?: string) => {
  if (!dateStr || dateStr === 'tanpa-tanggal') return '-'
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return dateStr
  return d.toLocaleDateString('id-ID', {
    weekday: 'short', day: 'numeric', month: 'short', year: 'numeric'
  })
}

const getInvoiceGroupDateRaw = (payment: any) => {
  if (invoiceDateGroupMode.value === 'transaction_date') {
    return payment.transaction_date || payment.payment_date
  }
  return payment.payment_date || payment.transaction_date
}

const getCompanyNameById = (companyId?: string) => {
  if (!companyId) return 'PT tidak diketahui'
  return masterDataStore.companies.find((company: any) => company.id === companyId)?.name || 'PT tidak diketahui'
}

const companyAndSearchFilteredInvoices = computed(() => {
  const compId = mutationForm.value.company_id
  const q = invoicePaymentSearch.value.trim().toLowerCase()
  return availableInvoicePayments.value.filter((payment: any) => {
    if (compId && payment.company_id !== compId) return false
    if (q) {
      const receipt = (payment.receipt_no || '').toLowerCase()
      const buyer = (payment.buyer_name || '').toLowerCase()
      const salesman = (payment.salesman_name || '').toLowerCase()
      const compName = getCompanyNameById(payment.company_id).toLowerCase()
      if (!receipt.includes(q) && !buyer.includes(q) && !salesman.includes(q) && !compName.includes(q)) {
        return false
      }
    }
    return true
  })
})

const buildDateGroups = (list: any[]) => {
  const map = new Map<string, {
    dateKey: string
    rawDate: string
    shortLabel: string
    fullLabel: string
    items: any[]
    count: number
    totalAvailable: number
    selectedCount: number
    selectedAmount: number
    isAllSelected: boolean
    isPartiallySelected: boolean
  }>()

  for (const payment of list) {
    const rawDate = getInvoiceGroupDateRaw(payment)
    const dateKey = toDateKey(rawDate)
    let group = map.get(dateKey)
    if (!group) {
      group = {
        dateKey,
        rawDate,
        shortLabel: formatShortDate(rawDate),
        fullLabel: formatDate(rawDate),
        items: [],
        count: 0,
        totalAvailable: 0,
        selectedCount: 0,
        selectedAmount: 0,
        isAllSelected: false,
        isPartiallySelected: false
      }
      map.set(dateKey, group)
    }
    group.items.push(payment)
    group.count += 1
    group.totalAvailable += Number(payment.available_amount || 0)
    const selVal = selectedInvoicePayments.value[payment.payment_id]
    if (selVal) {
      group.selectedCount += 1
      group.selectedAmount += Number(selVal || 0)
    }
  }

  const groups = Array.from(map.values()).sort((a, b) => b.dateKey.localeCompare(a.dateKey))
  for (const g of groups) {
    g.isAllSelected = g.count > 0 && g.selectedCount === g.count
    g.isPartiallySelected = g.selectedCount > 0 && g.selectedCount < g.count
  }
  return groups
}

const invoiceDateChips = computed(() => buildDateGroups(companyAndSearchFilteredInvoices.value))

const filteredInvoicePayments = computed(() => {
  if (!activeInvoiceDateFilter.value) return companyAndSearchFilteredInvoices.value
  return companyAndSearchFilteredInvoices.value.filter(
    (payment: any) => toDateKey(getInvoiceGroupDateRaw(payment)) === activeInvoiceDateFilter.value
  )
})

const groupedInvoicePaymentsByDate = computed(() => buildDateGroups(filteredInvoicePayments.value))

const selectedInvoiceDaysCount = computed(() => {
  const days = new Set<string>()
  for (const payment of availableInvoicePayments.value) {
    if (selectedInvoicePayments.value[payment.payment_id]) {
      days.add(toDateKey(getInvoiceGroupDateRaw(payment)))
    }
  }
  return days.size
})

const fetchAvailableInvoicePayments = async () => {
  const requestId = ++availableInvoiceRequestId
  if (!isFactoryInvoiceDeposit.value || !mutationForm.value.branch_id) {
    availableInvoicePayments.value = []
    isLoadingInvoicePayments.value = false
    return
  }
  try {
    isLoadingInvoicePayments.value = true
    availableInvoicePaymentsError.value = ''
    const params: any = { branch_id: mutationForm.value.branch_id }
    const response = await apiClient.get('/cashbank-mutations/available-invoice-payments', { params })
    if (requestId !== availableInvoiceRequestId) return
    availableInvoicePayments.value = response.data?.data || []
    for (const id of Object.keys(selectedInvoicePayments.value)) {
      const row = availableInvoicePayments.value.find(payment => payment.payment_id === id)
      if (row && Number(selectedInvoicePayments.value[id]) > Number(row.available_amount)) {
        selectedInvoicePayments.value[id] = Number(row.available_amount)
      }
    }
  } catch (error: any) {
    if (requestId !== availableInvoiceRequestId) return
    availableInvoicePayments.value = []
    availableInvoicePaymentsError.value = error.response?.data?.message || error.response?.data?.error || 'Gagal memuat penerimaan faktur yang belum disetor'
    toast.error(availableInvoicePaymentsError.value)
  } finally {
    if (requestId === availableInvoiceRequestId) isLoadingInvoicePayments.value = false
  }
}

const toggleInvoicePayment = (row: any) => {
  if (selectedInvoicePayments.value[row.payment_id]) {
    delete selectedInvoicePayments.value[row.payment_id]
    delete selectedInvoicePaymentCompanies.value[row.payment_id]
    delete selectedInvoicePaymentReceipts.value[row.payment_id]
  } else {
    if (!mutationForm.value.company_id) {
      mutationForm.value.company_id = row.company_id
    } else if (mutationForm.value.company_id !== row.company_id) {
      toast.error('Satu setoran hanya boleh berisi faktur dari PT Sumber yang sama')
      return
    }
    selectedInvoicePayments.value[row.payment_id] = Number(row.available_amount)
    selectedInvoicePaymentCompanies.value[row.payment_id] = row.company_id
    selectedInvoicePaymentReceipts.value[row.payment_id] = row.receipt_no
  }
}

const toggleDateGroupSelection = (group: { dateKey: string; items: any[]; isAllSelected: boolean }) => {
  if (group.isAllSelected) {
    for (const item of group.items) {
      delete selectedInvoicePayments.value[item.payment_id]
      delete selectedInvoicePaymentCompanies.value[item.payment_id]
      delete selectedInvoicePaymentReceipts.value[item.payment_id]
    }
  } else {
    let targetCompanyId = mutationForm.value.company_id
    if (!targetCompanyId) {
      const existingComp = Object.values(selectedInvoicePaymentCompanies.value)[0]
      targetCompanyId = existingComp || group.items[0]?.company_id || ''
      if (targetCompanyId) mutationForm.value.company_id = targetCompanyId
    }
    let skippedDifferentCompany = false
    for (const item of group.items) {
      if (targetCompanyId && item.company_id !== targetCompanyId) {
        skippedDifferentCompany = true
        continue
      }
      if (!selectedInvoicePayments.value[item.payment_id]) {
        selectedInvoicePayments.value[item.payment_id] = Number(item.available_amount)
        selectedInvoicePaymentCompanies.value[item.payment_id] = item.company_id
        selectedInvoicePaymentReceipts.value[item.payment_id] = item.receipt_no
      }
    }
    if (skippedDifferentCompany) {
      toast.info('Hanya memilih faktur yang sesuai dengan PT Sumber aktif')
    }
  }
}

const selectAllVisibleInvoices = () => {
  const list = filteredInvoicePayments.value
  if (!list.length) return
  let targetCompanyId = mutationForm.value.company_id
  if (!targetCompanyId) {
    const existingComp = Object.values(selectedInvoicePaymentCompanies.value)[0]
    targetCompanyId = existingComp || list[0]?.company_id || ''
    if (targetCompanyId) mutationForm.value.company_id = targetCompanyId
  }
  for (const item of list) {
    if (targetCompanyId && item.company_id !== targetCompanyId) continue
    if (!selectedInvoicePayments.value[item.payment_id]) {
      selectedInvoicePayments.value[item.payment_id] = Number(item.available_amount)
      selectedInvoicePaymentCompanies.value[item.payment_id] = item.company_id
      selectedInvoicePaymentReceipts.value[item.payment_id] = item.receipt_no
    }
  }
}

const clearSelectedInvoices = () => {
  selectedInvoicePayments.value = {}
  selectedInvoicePaymentCompanies.value = {}
  selectedInvoicePaymentReceipts.value = {}
}

const toggleCollapseInvoiceDate = (dateKey: string) => {
  collapsedInvoiceDates.value[dateKey] = !collapsedInvoiceDates.value[dateKey]
}

const removeSelectedInvoicePayment = (paymentId: string) => {
  delete selectedInvoicePayments.value[paymentId]
  delete selectedInvoicePaymentCompanies.value[paymentId]
  delete selectedInvoicePaymentReceipts.value[paymentId]
}

watch(() => [mutationForm.value.category, mutationForm.value.branch_id], ([newCat, newBranch], [oldCat, oldBranch]) => {
  if (isFactoryDeposit.value) {
    mutationForm.value.mutation_type = 'OUT'
    mutationForm.value.fund_type = 'CASH'
  }
  if (isFactoryInvoiceDeposit.value) {
    selectedInvoicePayments.value = {}
    selectedInvoicePaymentCompanies.value = {}
    selectedInvoicePaymentReceipts.value = {}
    activeInvoiceDateFilter.value = ''
    collapsedInvoiceDates.value = {}
    extraDepositAmount.value = 0
    mutationForm.value.amount = 0
    if (newCat !== oldCat || newBranch !== oldBranch || !availableInvoicePayments.value.length) {
      fetchAvailableInvoicePayments()
    }
  } else {
    selectedInvoicePayments.value = {}
    selectedInvoicePaymentCompanies.value = {}
    selectedInvoicePaymentReceipts.value = {}
    activeInvoiceDateFilter.value = ''
    extraDepositAmount.value = 0
    if (isFactoryDeposit.value) mutationForm.value.amount = 0
  }
})
watch(() => mutationForm.value.company_id, (newCompanyId) => {
  if (!isFactoryInvoiceDeposit.value) return
  if (newCompanyId) {
    for (const id of Object.keys(selectedInvoicePayments.value)) {
      if (selectedInvoicePaymentCompanies.value[id] !== newCompanyId) {
        delete selectedInvoicePayments.value[id]
        delete selectedInvoicePaymentCompanies.value[id]
        delete selectedInvoicePaymentReceipts.value[id]
      }
    }
  }
  if (activeInvoiceDateFilter.value && !invoiceDateChips.value.some(c => c.dateKey === activeInvoiceDateFilter.value)) {
    activeInvoiceDateFilter.value = ''
  }
})
watch(invoiceDateGroupMode, () => {
  activeInvoiceDateFilter.value = ''
  collapsedInvoiceDates.value = {}
})
watch([selectedInvoicePaymentTotal, extraDepositAmount], ([total, extra]) => {
  if (isFactoryInvoiceDeposit.value) mutationForm.value.amount = total + Number(extra || 0)
})

const formatRupiah = (value: number) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value || 0)
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    weekday: 'long', year: 'numeric', month: 'long', day: 'numeric'
  })
}

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('id-ID')
}

// Filtered Bank Accounts for Display
const filteredBanks = computed(() => {
  return banks.value.filter(b => {
    const branchMatches = !!currentBranchId.value && b.branch_id === currentBranchId.value
    return branchMatches && b.is_active !== false
  })
})

// Bank accounts filtered for current disburse or mutation form
const disburseBankAccounts = computed(() => {
  const compId = disburseForm.value.company_id
  const targetBranchId = disburseForm.value.branch_id || currentBranchId.value
  return banks.value.filter(b => {
    const companyMatches = !compId || !b.company_id || b.company_id === compId
    const branchMatches = !targetBranchId || !b.branch_id || b.branch_id === targetBranchId
    return companyMatches && branchMatches && b.is_active !== false
  })
})

const mutationBankAccounts = computed(() => {
  const compId = mutationForm.value.company_id
  const targetBranchId = mutationForm.value.branch_id || currentBranchId.value
  return banks.value.filter(b => {
    const companyMatches = isFactoryDeposit.value || !compId || !b.company_id || b.company_id === compId
    const branchMatches = isFactoryDeposit.value ? !!targetBranchId && b.branch_id === targetBranchId : (!targetBranchId || !b.branch_id || b.branch_id === targetBranchId)
    return companyMatches && branchMatches && b.is_active !== false
  })
})

// Fetch Data
const fetchBanksAndCash = async () => {
  try {
    isBanksLoading.value = true
    const branchId = currentBranchId.value
    // Muat rekening cabang untuk pilihan tujuan setoran.
    let bankUrl = '/finance/bank-accounts'
    let cashUrl = `/cashbank-mutations/cash-balance?branch_id=${branchId || ''}`
    if (selectedCompanyId.value) {
      cashUrl += `&company_id=${selectedCompanyId.value}`
    }

    const [banksRes, cashRes] = await Promise.all([
      apiClient.get(bankUrl),
      branchId ? apiClient.get(cashUrl) : Promise.resolve({ data: { data: { manual_cash_balance: 0, petty_cash_balance: 0, company_balances: [] } } })
    ])
    banks.value = banksRes.data?.data || banksRes.data || []
    cashBalance.value = cashRes.data?.data?.manual_cash_balance || 0
    pettyCashBalance.value = cashRes.data?.data?.petty_cash_balance || 0
    companyBalances.value = cashRes.data?.data?.company_balances || []
  } catch (error) {
    console.error('Failed to load banks/cash:', error)
  } finally {
    isBanksLoading.value = false
  }
}

// Date Filters
const reportFilterDate = ref('')
const mutationFilterDate = ref('')

const fetchReports = async () => {
  try {
    isReportsLoading.value = true
    const branchId = currentBranchId.value
    let queryParams = ''
    if (branchId) queryParams += `branch_id=${branchId}`
    if (selectedCompanyId.value) queryParams += `${queryParams ? '&' : ''}company_id=${selectedCompanyId.value}`
    if (reportFilterDate.value) queryParams += `${queryParams ? '&' : ''}date=${reportFilterDate.value}`
    const res = await apiClient.get(`/cashier-reports?${queryParams}`)
    reports.value = res.data?.data || res.data || []
  } catch (error) {
    console.error('Failed to load reports:', error)
  } finally {
    isReportsLoading.value = false
  }
}

const fetchMutations = async () => {
  try {
    isMutationsLoading.value = true
    const branchId = currentBranchId.value
    let queryParams = isBranchDepositReviewer.value ? '?' : (branchId ? `?branch_id=${branchId}` : '?')
    if (selectedCompanyId.value && !isBranchDepositReviewer.value) queryParams += `&company_id=${selectedCompanyId.value}`
    if (mutationFilterDate.value) queryParams += `&date=${mutationFilterDate.value}`
    if (mutationFundTypeFilter.value) queryParams += `&fund_type=${mutationFundTypeFilter.value}`
    if (mutationCategoryFilter.value) queryParams += `&category=${mutationCategoryFilter.value}`

    const res = await apiClient.get(`/cashbank-mutations${queryParams}`)
    mutations.value = res.data?.data || res.data || []
  } catch (error) {
    console.error('Failed to load mutations:', error)
  } finally {
    isMutationsLoading.value = false
  }
}

const fetchReimbursements = async () => {
  try {
    isReimbursementsLoading.value = true
    const branchId = currentBranchId.value
    let queryParams = `?limit=100`
    if (branchId) queryParams += `&branch_id=${branchId}`
    if (reimbursementFilterStatus.value === 'APPROVED') {
      queryParams += `&status=APPROVED`
    }
    const res = await apiClient.get(`/reimbursements/manage${queryParams}`)
    let allData = res.data?.data || []
    
    if (selectedCompanyId.value) {
      allData = allData.filter((r: any) => r.employee?.company_id === selectedCompanyId.value)
    }

    if (reimbursementFilterStatus.value === 'APPROVED') {
      // Yang siap dicairkan (APPROVED dan belum DISBURSED)
      reimbursements.value = allData.filter((r: any) => r.status === 'APPROVED' && r.disbursement_status !== 'DISBURSED')
    } else if (reimbursementFilterStatus.value === 'DISBURSED') {
      reimbursements.value = allData.filter((r: any) => r.disbursement_status === 'DISBURSED')
    } else {
      reimbursements.value = allData
    }
  } catch (error) {
    console.error('Failed to load reimbursements:', error)
  } finally {
    isReimbursementsLoading.value = false
  }
}

watch([selectedCompanyId, reportFilterDate], () => {
  fetchReports()
})

watch([selectedCompanyId, mutationFilterDate, mutationFundTypeFilter, mutationCategoryFilter], () => {
  fetchMutations()
})

watch([selectedCompanyId, reimbursementFilterStatus], () => {
  fetchReimbursements()
})

watch(selectedCompanyId, () => {
  fetchBanksAndCash()
})

watch(currentBranchId, (newId, oldId) => {
  if (newId !== oldId) {
    selectedCompanyId.value = ''
    if (!editingMutationId.value) mutationForm.value.branch_id = newId
  }
  if (newId) {
    if (!reportForm.value.branch_id) reportForm.value.branch_id = newId
    if (!mutationForm.value.branch_id) mutationForm.value.branch_id = newId
    if (!disburseForm.value.branch_id) disburseForm.value.branch_id = newId
    fetchBanksAndCash()
    fetchReports()
    fetchMutations()
    fetchReimbursements()
  }
})

// Auto sync bank accounts when branch or PT changes in modal
watch(() => mutationForm.value.branch_id, async (newBranchId) => {
  if (newBranchId && newBranchId !== currentBranchId.value) {
    try {
      const res = await apiClient.get(`/finance/bank-accounts?branch_id=${newBranchId}`)
      const branchBanks = res.data?.data || res.data || []
      const existingIds = new Set(banks.value.map((b: any) => b.id))
      for (const b of branchBanks) {
        if (!existingIds.has(b.id)) {
          banks.value.push(b)
        }
      }
    } catch (e) {
      console.error('Failed to load banks for selected branch:', e)
    }
  }
})

watch(() => [mutationForm.value.fund_type, mutationForm.value.company_id, mutationForm.value.destination_company_id, mutationForm.value.branch_id, mutationForm.value.category], () => {
  if (mutationForm.value.fund_type === 'BANK' || isFactoryDeposit.value) {
    const available = mutationBankAccounts.value
    if (available.length > 0) {
      const exists = available.some(b => b.id === mutationForm.value.bank_account_id)
      if (!exists) {
        mutationForm.value.bank_account_id = available[0].id
      }
    } else {
      mutationForm.value.bank_account_id = ''
    }
  }
})

const kasPabriks = ref<any[]>([])

const fetchKasPabriks = async () => {
  try {
    const res = await apiClient.get('/kas-pabrik?branch_id=')
    kasPabriks.value = res.data?.data || res.data || []
  } catch (error) {
    console.error('Failed to load Kas Pabrik:', error)
  }
}

onMounted(async () => {
  await masterDataStore.fetchBranches()
  await masterDataStore.fetchCompanies()
  if (currentBranchId.value) {
    reportForm.value.branch_id = currentBranchId.value
    mutationForm.value.branch_id = currentBranchId.value
    disburseForm.value.branch_id = currentBranchId.value
  }
  fetchBanksAndCash()
  fetchReports()
  fetchMutations()
  fetchKasPabriks()
  fetchReimbursements()
})

// Bank Actions
const openBalanceModal = (bank: any) => {
  selectedBank.value = bank
  newBalance.value = bank.manual_balance || 0
  showBalanceModal.value = true
}

const saveBalance = async () => {
  try {
    await apiClient.put(`/bank-accounts/${selectedBank.value.id}/manual-balance`, {
      manual_balance: newBalance.value
    })
    toast.success('Saldo bank berhasil diupdate!')
    showBalanceModal.value = false
    fetchBanksAndCash()
  } catch (error) {
    toast.error('Gagal mengupdate saldo bank')
  }
}

// Add / Edit Bank Account Profile
const openAddBankModal = () => {
  isEditingBank.value = false
  bankForm.value = {
    id: '',
    company_id: '',
    branch_id: currentBranchId.value,
    bank_name: 'BCA',
    account_number: '',
    account_holder: '',
    manual_balance: 0,
    notes: '',
    is_active: true
  }
  showAddEditBankModal.value = true
}

const openEditBankModal = (bank: any) => {
  isEditingBank.value = true
  bankForm.value = {
    id: bank.id,
    company_id: bank.company_id || '',
    branch_id: bank.branch_id || currentBranchId.value,
    bank_name: bank.bank_name || '',
    account_number: bank.account_number || '',
    account_holder: bank.account_holder || '',
    manual_balance: bank.manual_balance || 0,
    notes: bank.notes || '',
    is_active: bank.is_active ?? true
  }
  showAddEditBankModal.value = true
}

const saveBankAccount = async () => {
  if (!bankForm.value.bank_name || !bankForm.value.account_number || !bankForm.value.account_holder) {
    toast.error('Nama Bank, Nomor Rekening, dan Nama Pemilik wajib diisi')
    return
  }

  isSavingBank.value = true
  try {
    const payload: any = {
      bank_name: bankForm.value.bank_name,
      account_number: bankForm.value.account_number,
      account_holder: bankForm.value.account_holder,
      company_id: bankForm.value.company_id || null,
      branch_id: bankForm.value.branch_id || currentBranchId.value || null,
      manual_balance: Number(bankForm.value.manual_balance) || 0,
      notes: bankForm.value.notes,
      is_active: bankForm.value.is_active
    }

    if (isEditingBank.value && bankForm.value.id) {
      await apiClient.put(`/finance/bank-accounts/${bankForm.value.id}`, payload)
      toast.success('Rekening bank berhasil diperbarui!')
    } else {
      await apiClient.post('/finance/bank-accounts', payload)
      toast.success('Rekening bank baru berhasil ditambahkan!')
    }
    showAddEditBankModal.value = false
    fetchBanksAndCash()
  } catch (error: any) {
    toast.error(error.response?.data?.message || error.response?.data?.error || 'Gagal menyimpan rekening bank')
  } finally {
    isSavingBank.value = false
  }
}

const deleteBankAccount = async (bank: any) => {
  if (!confirm(`Hapus profil rekening "${bank.bank_name} - ${bank.account_number}"? Rekening yang sudah digunakan dalam transaksi harus dinonaktifkan agar riwayat tetap tersimpan.`)) return
  try {
    await apiClient.delete(`/finance/bank-accounts/${bank.id}`)
    toast.success('Rekening bank berhasil dihapus')
    fetchBanksAndCash()
  } catch (error: any) {
    toast.error(error.response?.data?.error || error.response?.data?.message || 'Gagal menghapus rekening bank')
  }
}

// Petty Cash Actions (Per PT)
const openPettyCashBalanceModal = (companyId?: string) => {
  const targetCompId = companyId || selectedCompanyId.value || (companyBalances.value[0]?.company_id || '')
  let currentVal = pettyCashBalance.value
  if (targetCompId) {
    const found = companyBalances.value.find(c => c.company_id === targetCompId)
    if (found) currentVal = found.petty_cash_balance
  }
  pettyCashBalanceForm.value = {
    company_id: targetCompId,
    petty_cash_balance: currentVal || 0,
    notes: ''
  }
  showPettyCashBalanceModal.value = true
}

const savePettyCashBalance = async () => {
  const branchId = currentBranchId.value
  if (!branchId) {
    toast.error('Cabang tidak terdeteksi')
    return
  }

  isSavingPettyCash.value = true
  try {
    await apiClient.put(`/cashbank-mutations/branches/${branchId}/petty-cash-balance`, {
      company_id: pettyCashBalanceForm.value.company_id || null,
      petty_cash_balance: pettyCashBalanceForm.value.petty_cash_balance,
      notes: pettyCashBalanceForm.value.notes
    })
    toast.success('Saldo Kas Kecil berhasil diperbarui!')
    showPettyCashBalanceModal.value = false
    fetchBanksAndCash()
    fetchMutations()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal mengupdate saldo kas kecil')
  } finally {
    isSavingPettyCash.value = false
  }
}

// Kas Besar (Cash Balance) Actions (Per PT)
const openCashBalanceModal = (companyId?: string) => {
  const targetCompId = companyId || selectedCompanyId.value || (companyBalances.value[0]?.company_id || '')
  let currentVal = cashBalance.value
  if (targetCompId) {
    const found = companyBalances.value.find(c => c.company_id === targetCompId)
    if (found) currentVal = found.manual_cash_balance
  }
  cashBalanceForm.value = {
    company_id: targetCompId,
    manual_cash_balance: currentVal || 0,
    notes: ''
  }
  showCashBalanceModal.value = true
}

const saveCashBalance = async () => {
  const branchId = currentBranchId.value
  if (!branchId) {
    toast.error('Cabang tidak terdeteksi')
    return
  }

  isSavingCashBalance.value = true
  try {
    await apiClient.put(`/cashbank-mutations/branches/${branchId}/cash-balance`, {
      company_id: cashBalanceForm.value.company_id || null,
      manual_cash_balance: cashBalanceForm.value.manual_cash_balance,
      notes: cashBalanceForm.value.notes
    })
    toast.success('Saldo Kas Besar berhasil diperbarui!')
    showCashBalanceModal.value = false
    fetchBanksAndCash()
    fetchMutations()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal mengupdate saldo kas besar')
  } finally {
    isSavingCashBalance.value = false
  }
}

// Fetch System Banks for Dropdown in Report Modal
const fetchSystemBanks = async (companyId: string) => {
  try {
    const res = await apiClient.get('/finance/bank-accounts')
    let allBanks = res.data?.data || []
    if (companyId) {
      allBanks = allBanks.filter((b: any) => b.company_id === companyId)
    }
    if (currentBranchId.value) {
      allBanks = allBanks.filter((b: any) => b.branch_id === currentBranchId.value)
    }
    systemBankAccounts.value = allBanks
  } catch (error) {
    console.error('Failed to load system bank accounts:', error)
  }
}

const fetchDailySummaryForCompany = async (companyId: string) => {
  const selectedDate = reportForm.value.report_date || new Date().toISOString().split('T')[0]
  const branchId = authStore.user?.branch_id || ''
  try {
    isFetchingSummary.value = true
    let url = `/cashier-reports/daily-summary?date=${selectedDate}`
    if (companyId) {
      url += `&company_id=${companyId}`
    } else {
      url += `&company_id=ALL`
    }
    if (branchId) {
      url += `&branch_id=${branchId}`
    }
    const res = await apiClient.get(url)
    const summary = res.data?.data || { total_cash_sales: 0, total_transfer_sales: 0 }
    autoApprovedCashSales.value = summary.total_cash_sales || 0
    autoApprovedTransferSales.value = summary.total_transfer_sales || 0
    reportForm.value.total_cash_sales = autoApprovedCashSales.value
    reportForm.value.total_transfer_sales = autoApprovedTransferSales.value
    calculateDepositAmount()
  } catch (error) {
    toast.error('Gagal mengambil ringkasan persetujuan kasir untuk tanggal & PT terpilih')
  } finally {
    isFetchingSummary.value = false
  }
}

const resetToAutoValues = () => {
  reportForm.value.total_cash_sales = autoApprovedCashSales.value
  reportForm.value.total_transfer_sales = autoApprovedTransferSales.value
  calculateDepositAmount()
  toast.info('Nominal dikembalikan sesuai hasil persetujuan kasir')
}

const calculateDepositAmount = () => {
  const cashSales = Number(reportForm.value.total_cash_sales) || 0
  const allocated = Number(reportForm.value.allocated_to_petty_cash) || 0
  reportForm.value.deposited_amount = Math.max(0, cashSales - allocated)
}

watch(() => reportForm.value.allocated_to_petty_cash, () => {
  calculateDepositAmount()
})

watch(() => reportForm.value.total_cash_sales, () => {
  calculateDepositAmount()
})

watch(() => reportForm.value.report_date, (newDate) => {
  if (showReportModal.value && newDate) {
    fetchDailySummaryForCompany(reportForm.value.company_id)
  }
})

watch(() => reportForm.value.company_id, (newCompanyId) => {
  if (showReportModal.value) {
    fetchDailySummaryForCompany(newCompanyId)
    fetchSystemBanks(newCompanyId)
  }
})

watch(() => reportForm.value.target_bank_account_id, (newBankId) => {
  if (newBankId) {
    const selected = systemBankAccounts.value.find(b => b.id === newBankId)
    if (selected) {
      reportForm.value.target_bank_name = selected.bank_name
      reportForm.value.target_account_number = selected.account_number
      reportForm.value.target_account_name = selected.account_holder
    }
  }
})

watch(() => reportForm.value.target_kas_pabrik_id, (newKasId) => {
  if (newKasId) {
    const selected = kasPabriks.value.find(k => k.id === newKasId)
    if (selected) {
      reportForm.value.target_bank_name = selected.kas_name
      reportForm.value.target_account_number = selected.bank_account_number || '-'
      reportForm.value.target_account_name = selected.bank_account_name || 'Kas Pusat'
    }
  }
})

// Report Actions
const openReportModal = async () => {
  isEditingReport.value = false
  editingReportId.value = null
  const today = new Date().toISOString().split('T')[0]
  reportForm.value = {
    branch_id: currentBranchId.value,
    company_id: selectedCompanyId.value || '',
    report_date: today,
    total_cash_sales: 0,
    total_transfer_sales: 0,
    allocated_to_petty_cash: 0,
    deposited_amount: 0,
    deposit_destination_type: 'BRANCH_BANK',
    target_bank_account_id: '',
    target_kas_pabrik_id: '',
    target_bank_name: '',
    target_account_number: '',
    target_account_name: '',
    notes: '',
    deposit_proof_url: ''
  }
  
  await fetchDailySummaryForCompany(selectedCompanyId.value)
  await fetchSystemBanks(selectedCompanyId.value)
  
  selectedReportFile.value = null
  showReportModal.value = true
}

const openEditReportModal = async (item: any) => {
  isEditingReport.value = true
  editingReportId.value = item.id
  
  const reportDateStr = item.report_date ? item.report_date.split('T')[0] : new Date().toISOString().split('T')[0]
  
  reportForm.value = {
    branch_id: item.branch_id || currentBranchId.value,
    company_id: item.company_id || '',
    report_date: reportDateStr,
    total_cash_sales: item.total_cash_sales || 0,
    total_transfer_sales: item.total_transfer_sales || 0,
    allocated_to_petty_cash: item.allocated_to_petty_cash || 0,
    deposited_amount: item.deposited_amount || 0,
    deposit_destination_type: item.deposit_destination_type || 'BRANCH_BANK',
    target_bank_account_id: item.target_bank_account_id || '',
    target_kas_pabrik_id: item.target_kas_pabrik_id || '',
    target_bank_name: item.target_bank_name || '',
    target_account_number: item.target_account_number || '',
    target_account_name: item.target_account_name || '',
    notes: item.notes || '',
    deposit_proof_url: item.deposit_proof_url || ''
  }

  await fetchSystemBanks(item.company_id || '')
  selectedReportFile.value = null
  showReportModal.value = true
}

const deleteReport = async (item: any) => {
  const dateFormatted = formatDate(item.report_date)
  if (!confirm(`Hapus laporan harian kas besar tanggal ${dateFormatted}? Saldo dan mutasi terkait akan dikembalikan secara otomatis.`)) {
    return
  }

  try {
    await apiClient.delete(`/cashier-reports/${item.id}`)
    toast.success('Laporan harian berhasil dihapus dan saldo telah dipulihkan')
    fetchReports()
    fetchBanksAndCash()
    fetchMutations()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menghapus laporan harian')
  }
}

const saveReport = async () => {
  if (reportForm.value.deposited_amount > 0 && reportForm.value.deposit_destination_type === 'FACTORY_KAS') {
    toast.error('Setoran ke pabrik dicatat melalui Setor Faktur ke Pabrik atau Setor Dana Nonfaktur')
    return
  }
  if (!reportForm.value.branch_id) {
    reportForm.value.branch_id = currentBranchId.value
  }
  if (!reportForm.value.branch_id) {
    toast.error('Cabang tidak terdeteksi atau belum dipilih')
    return
  }

  if (reportForm.value.deposit_destination_type === 'BRANCH_BANK') {
    if (!reportForm.value.target_bank_name || !reportForm.value.target_account_number || !reportForm.value.target_account_name) {
      toast.error('Lengkapi detail bank tujuan cabang')
      return
    }
  } else if (reportForm.value.deposit_destination_type === 'FACTORY_KAS') {
    if (!reportForm.value.target_kas_pabrik_id) {
      toast.error('Pilih Kas Pabrik / Pusat tujuan setoran')
      return
    }
  }

  if (reportForm.value.deposited_amount > 0 && !reportForm.value.deposit_proof_url && !selectedReportFile.value) {
    toast.error('Upload bukti transfer / setoran diwajibkan jika ada setoran')
    return
  }

  isUploadingReport.value = true
  try {
    let proofUrl = reportForm.value.deposit_proof_url || ''
    if (selectedReportFile.value) {
      const formData = new FormData()
      formData.append('file', selectedReportFile.value)
      const uploadRes = await apiClient.post('/cashier-reports/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      proofUrl = uploadRes.data?.data?.url || uploadRes.data?.url
    }
    
    reportForm.value.deposit_proof_url = proofUrl

    const reportPayload: any = { ...reportForm.value }
    if (!reportPayload.target_bank_account_id) reportPayload.target_bank_account_id = null
    if (!reportPayload.target_kas_pabrik_id) reportPayload.target_kas_pabrik_id = null
    if (!reportPayload.company_id) reportPayload.company_id = null

    if (isEditingReport.value && editingReportId.value) {
      await apiClient.put(`/cashier-reports/${editingReportId.value}`, reportPayload)
      toast.success('Laporan harian berhasil diperbarui!')
    } else {
      await apiClient.post('/cashier-reports', reportPayload)
      toast.success('Laporan harian berhasil disimpan!')
    }

    showReportModal.value = false
    fetchReports()
    fetchBanksAndCash()
    fetchMutations()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyimpan laporan')
  } finally {
    isUploadingReport.value = false
  }
}

// Reimbursement Actions
const openDisburseModal = (item: any) => {
  selectedReimbursement.value = item
  disburseForm.value.disbursed_via = 'PETTY_CASH'
  disburseForm.value.company_id = item.employee?.company_id || selectedCompanyId.value || ''
  disburseForm.value.bank_account_id = ''
  disburseForm.value.proof_url = ''
  disburseForm.value.notes = ''
  disburseForm.value.branch_id = currentBranchId.value
  selectedDisburseFile.value = null
  showDisburseModal.value = true
}

// Check balance for selected reimbursement source
const selectedFundBalance = computed(() => {
  const compId = disburseForm.value.company_id
  if (disburseForm.value.disbursed_via === 'PETTY_CASH') {
    if (compId) {
      const bcb = companyBalances.value.find(b => b.company_id === compId)
      return bcb ? bcb.petty_cash_balance : 0
    }
    return pettyCashBalance.value
  } else if (disburseForm.value.disbursed_via === 'CASH') {
    if (compId) {
      const bcb = companyBalances.value.find(b => b.company_id === compId)
      return bcb ? bcb.manual_cash_balance : 0
    }
    return cashBalance.value
  } else if (disburseForm.value.disbursed_via === 'BANK_TRANSFER') {
    const bank = banks.value.find(b => b.id === disburseForm.value.bank_account_id)
    return bank ? bank.manual_balance : 0
  }
  return 0
})

const saveDisburse = async () => {
  if (!selectedReimbursement.value) return

  if (disburseForm.value.disbursed_via === 'BANK_TRANSFER' && !disburseForm.value.bank_account_id) {
    toast.error('Pilih rekening bank sumber pencairan')
    return
  }

  isDisbursing.value = true
  try {
    let proofUrl = ''
    if (selectedDisburseFile.value) {
      const formData = new FormData()
      formData.append('file', selectedDisburseFile.value)
      const uploadRes = await apiClient.post('/cashier-reports/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      proofUrl = uploadRes.data?.data?.url || uploadRes.data?.url
    }

    const branchId = disburseForm.value.branch_id || currentBranchId.value

    await apiClient.post(`/reimbursements/manage/${selectedReimbursement.value.id}/disburse`, {
      branch_id: branchId,
      company_id: disburseForm.value.company_id || null,
      disbursed_via: disburseForm.value.disbursed_via,
      bank_account_id: disburseForm.value.bank_account_id || null,
      disbursement_proof_url: proofUrl,
      notes: disburseForm.value.notes
    })

    toast.success('Dana reimbursement berhasil dicairkan!')
    showDisburseModal.value = false
    fetchReimbursements()
    fetchBanksAndCash()
    fetchMutations()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal mencairkan reimbursement')
  } finally {
    isDisbursing.value = false
  }
}

// Mutation Actions
const openCreateMutationModal = (defaultFundType = 'PETTY_CASH', defaultMutationType = 'IN') => {
  editingMutationId.value = null
  selectedFile.value = null
  selectedInvoicePayments.value = {}
  selectedInvoicePaymentCompanies.value = {}
  selectedInvoicePaymentReceipts.value = {}
  invoicePaymentSearch.value = ''
  activeInvoiceDateFilter.value = ''
  collapsedInvoiceDates.value = {}
  extraDepositAmount.value = 0
  const defaultCompId = selectedCompanyId.value || (branchCompanies.value[0]?.id || masterDataStore.companies[0]?.id || '')
  mutationForm.value = {
    branch_id: currentBranchId.value,
    company_id: defaultCompId,
    destination_company_id: '',
    mutation_type: defaultMutationType,
    fund_type: defaultFundType,
    category: defaultFundType === 'PETTY_CASH' ? (defaultMutationType === 'IN' ? 'PETTY_CASH_TOPUP' : 'OPERATIONAL_EXPENSE') : 'OPERATIONAL_EXPENSE',
    bank_account_id: '',
    is_setor_pusat: false,
    target_kas_pabrik_id: '',
    amount: 0,
    notes: '',
    proof_url: '',
    fund_origin: '',
    mutation_date: new Date().toISOString().split('T')[0]
  }
  if (defaultFundType === 'BANK' && mutationBankAccounts.value.length > 0) {
    mutationForm.value.bank_account_id = mutationBankAccounts.value[0].id
  }
  showMutationModal.value = true
}

const openFactoryInvoiceDepositModal = () => {
  openCreateMutationModal('CASH', 'OUT')
  invoicePaymentSearch.value = ''
  activeInvoiceDateFilter.value = ''
  collapsedInvoiceDates.value = {}
  mutationForm.value.company_id = ''
  mutationForm.value.category = 'FACTORY_INVOICE_DEPOSIT'
  fetchAvailableInvoicePayments()
}

const openFactoryNonInvoiceDepositModal = () => {
  openCreateMutationModal('CASH', 'OUT')
  mutationForm.value.company_id = selectedCompanyId.value || ''
  mutationForm.value.category = 'FACTORY_NON_INVOICE_DEPOSIT'
}

const editMutation = (item: any) => {
  editingMutationId.value = item.id
  selectedFile.value = null
  selectedInvoicePayments.value = {}
  selectedInvoicePaymentCompanies.value = {}
  selectedInvoicePaymentReceipts.value = {}
  mutationForm.value = {
    branch_id: item.branch_id || currentBranchId.value,
    company_id: item.company_id || '',
    destination_company_id: item.destination_company_id || '',
    mutation_type: item.mutation_type || 'IN',
    fund_type: item.fund_type || 'PETTY_CASH',
    category: item.category || 'OPERATIONAL_EXPENSE',
    bank_account_id: item.bank_account_id || '',
    is_setor_pusat: !!item.target_kas_pabrik_id,
    target_kas_pabrik_id: item.target_kas_pabrik_id || '',
    amount: item.amount || 0,
    notes: item.notes || '',
    proof_url: item.proof_url || '',
    fund_origin: item.fund_origin || '',
    mutation_date: item.mutation_date ? item.mutation_date.split('T')[0] : new Date().toISOString().split('T')[0]
  }
  showMutationModal.value = true
}

const deleteMutation = async (item: any) => {
  if (deletingMutationId.value) return
  if (!confirm(`Apakah Anda yakin ingin menghapus mutasi "${item.notes || item.category}" sebesar ${formatRupiah(item.amount)}? Saldo akan dikoreksi otomatis.`)) {
    return
  }

  deletingMutationId.value = item.id
  try {
    await apiClient.delete(`/cashbank-mutations/${item.id}`)
    toast.success('Mutasi berhasil dihapus dan saldo telah disesuaikan')
    await Promise.all([
      fetchMutations(),
      fetchBanksAndCash()
    ])
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menghapus mutasi')
    // Sinkronisasi ulang jika data stale di browser
    fetchMutations()
    fetchBanksAndCash()
  } finally {
    deletingMutationId.value = null
  }
}

const reviewFactoryDeposit = async (item: any, decision: 'approve' | 'reject') => {
  let reason = ''
  if (decision === 'approve') {
    const breakdown = item.category === 'FACTORY_INVOICE_DEPOSIT'
      ? `\nAlokasi faktur: ${formatRupiah(mutationInvoiceTotal(item))}\nDana nonfaktur: ${formatRupiah(mutationNonInvoiceAmount(item))}`
      : `\nDana nonfaktur: ${formatRupiah(item.amount)}`
    if (!confirm(`Terima setoran ${formatRupiah(item.amount)} untuk rekening ${item.bank_account?.bank_name || '-'} ${item.bank_account?.account_number || ''} milik cabang ${item.branch?.name || item.bank_account?.branch?.name || 'tujuan'}?${breakdown}`)) return
  } else {
    reason = window.prompt('Tuliskan alasan penolakan setoran:')?.trim() || ''
    if (!reason) return
  }
  try {
    await apiClient.post(`/cashbank-mutations/${item.id}/${decision}`, decision === 'reject' ? { reason } : {})
    toast.success(decision === 'approve' ? 'Setoran diterima dan saldo bank tujuan diperbarui' : 'Setoran ditolak dan saldo kas cabang dipulihkan')
    await Promise.all([fetchMutations(), fetchBanksAndCash()])
  } catch (error: any) {
    toast.error(error.response?.data?.message || `Gagal ${decision === 'approve' ? 'menerima' : 'menolak'} setoran`)
  }
}

const handleFileChange = (e: any) => {
  if (e.target.files && e.target.files[0]) {
    selectedFile.value = e.target.files[0]
  }
}

const saveMutation = async () => {
  if (isFactoryInvoiceDeposit.value) {
    mutationForm.value.amount = selectedInvoicePaymentTotal.value + Number(extraDepositAmount.value || 0)
  }
  if (!mutationForm.value.branch_id || mutationForm.value.branch_id === '00000000-0000-0000-0000-000000000000') {
    mutationForm.value.branch_id = currentBranchId.value
  }
  if (!mutationForm.value.branch_id) {
    toast.error('Cabang wajib dipilih')
    return
  }

  if (mutationForm.value.amount <= 0) {
    toast.error('Nominal harus lebih dari 0')
    return
  }
  if ((mutationForm.value.fund_type === 'BANK' || isFactoryDeposit.value) && !mutationForm.value.bank_account_id) {
    toast.error('Pilih rekening bank')
    return
  }
  if (isFactoryInvoiceDeposit.value && (!Object.keys(selectedInvoicePayments.value).length || selectedInvoicePaymentTotal.value <= 0)) {
    toast.error('Pilih penerimaan faktur yang sudah disetujui kasir')
    return
  }
  if (isFactoryInvoiceDeposit.value && (!Number.isFinite(extraDepositAmount.value) || extraDepositAmount.value < 0)) {
    toast.error('Tambahan dana nonfaktur tidak valid')
    return
  }
  if (isFactoryDeposit.value && !mutationForm.value.company_id) {
    toast.error('Pilih PT sumber setoran')
    return
  }
  if ((mutationForm.value.category === 'FACTORY_NON_INVOICE_DEPOSIT' || extraDepositAmount.value > 0) && (!mutationForm.value.fund_origin || !mutationForm.value.notes.trim())) {
    toast.error('Isi asal dana nonfaktur dan alasan setoran')
    return
  }
  if (isFactoryDeposit.value && !selectedFile.value && !mutationForm.value.proof_url) {
    toast.error('Unggah bukti setoran ke rekening cabang tujuan')
    return
  }
  if (mutationForm.value.mutation_type === 'OUT' && mutationForm.value.is_setor_pusat && !mutationForm.value.target_kas_pabrik_id) {
    toast.error('Pilih Kas Pusat tujuan')
    return
  }

  isUploading.value = true
  try {
    let proofUrl = mutationForm.value.proof_url || ''
    if (selectedFile.value) {
      const formData = new FormData()
      formData.append('file', selectedFile.value)
      const uploadRes = await apiClient.post('/cashier-reports/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      proofUrl = uploadRes.data?.data?.url || uploadRes.data?.url
    }
    
    const payload: any = { ...mutationForm.value, proof_url: proofUrl }
    if (isFactoryDeposit.value) {
      payload.mutation_type = 'OUT'
      payload.fund_type = 'CASH'
      if (isFactoryInvoiceDeposit.value) {
        payload.invoice_payments = Object.entries(selectedInvoicePayments.value).map(([payment_id, amount]) => ({ payment_id, amount: Number(amount) }))
        payload.amount = selectedInvoicePaymentTotal.value + Number(extraDepositAmount.value || 0)
        if (!payload.notes && !extraDepositAmount.value) payload.notes = 'Setoran tunai faktur ke rekening cabang'
      }
    }
    if (!payload.company_id) {
      payload.company_id = null
    }
    if (payload.fund_type !== 'BANK' && !isFactoryDeposit.value) {
      payload.bank_account_id = null
    }
    if (payload.mutation_type !== 'OUT' || !payload.is_setor_pusat) {
      delete payload.target_kas_pabrik_id
    }
    delete payload.is_setor_pusat

    if (editingMutationId.value) {
      await apiClient.put(`/cashbank-mutations/${editingMutationId.value}`, payload)
      toast.success('Mutasi berhasil diperbarui!')
    } else {
      await apiClient.post('/cashbank-mutations', payload)
      toast.success('Mutasi berhasil dicatat!')
    }

    showMutationModal.value = false
    fetchMutations()
    fetchBanksAndCash()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyimpan mutasi')
  } finally {
    isUploading.value = false
  }
}

// Modal Preview State with Zoom & Rotate
const showPreviewModal = ref(false)
const previewImageUrl = ref('')
const previewModalTitle = ref('Bukti Transaksi')
const previewModalSubtitle = ref('')
const zoomScale = ref(1.0)
const rotationDegrees = ref(0)

const getFullImageUrl = (path?: string | null): string => {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:')) return path
  const apiUrl =
    import.meta.env.VITE_API_URL ||
    import.meta.env.VITE_API_BASE_URL ||
    window.location.origin
  try {
    const origin = new URL(apiUrl).origin
    return origin + (path.startsWith('/') ? path : `/${path}`)
  } catch {
    const host = apiUrl.replace(/\/api\/v1\/?$/, '').replace(/\/api\/?$/, '')
    return host + (path.startsWith('/') ? path : `/${path}`)
  }
}

const openPreviewModal = (url: string, title = 'Bukti Transaksi', subtitle = '') => {
  if (!url) return
  previewImageUrl.value = getFullImageUrl(url)
  previewModalTitle.value = title
  previewModalSubtitle.value = subtitle
  zoomScale.value = 1.0
  rotationDegrees.value = 0
  showPreviewModal.value = true
}

const closePreviewModal = () => {
  showPreviewModal.value = false
  previewImageUrl.value = ''
  zoomScale.value = 1.0
  rotationDegrees.value = 0
}

const zoomIn = () => {
  if (zoomScale.value < 4.0) {
    zoomScale.value = Math.min(4.0, Number((zoomScale.value + 0.25).toFixed(2)))
  }
}

const zoomOut = () => {
  if (zoomScale.value > 0.25) {
    zoomScale.value = Math.max(0.25, Number((zoomScale.value - 0.25).toFixed(2)))
  }
}

const rotateImage = (direction: 'cw' | 'ccw' = 'cw') => {
  if (direction === 'cw') {
    rotationDegrees.value = (rotationDegrees.value + 90) % 360
  } else {
    rotationDegrees.value = (rotationDegrees.value - 90 + 360) % 360
  }
}

const resetZoom = () => {
  zoomScale.value = 1.0
  rotationDegrees.value = 0
}

const openProof = (url: string, title = 'Bukti Transaksi', subtitle = '') => {
  openPreviewModal(url, title, subtitle)
}
</script>

<template>
  <div class="p-6 max-w-7xl mx-auto space-y-8">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 flex items-center gap-2">
          <Landmark class="w-6 h-6 text-primary" />
          Laporan Kasir, Kas Besar & Kas Kecil
        </h1>
        <p class="text-slate-500 mt-1">Kelola kas besar, kas kecil, rekening bank, setoran, dan reimbursement untuk cabang yang sedang aktif.</p>
      </div>
      
      <!-- Filter Controls: Company & Refresh -->
      <div class="flex flex-wrap items-center gap-3">
        <!-- PT / Company Selector Dropdown -->
        <div class="flex items-center gap-2 bg-white border border-slate-200 rounded-xl px-3 py-1.5 shadow-sm">
          <Building class="w-4 h-4 text-primary shrink-0" />
          <select 
            v-model="selectedCompanyId" 
            class="text-xs font-bold text-slate-700 outline-none bg-transparent cursor-pointer"
          >
            <option value="">Semua PT di {{ currentBranchName }}</option>
            <option v-for="c in branchCompanies" :key="c.id" :value="c.id">
              {{ c.name }}
            </option>
          </select>
        </div>

        <button @click="fetchBanksAndCash" class="p-2 text-slate-500 hover:text-primary rounded-xl border border-slate-200 bg-white hover:bg-slate-50 transition-colors shadow-sm" title="Refresh Saldo">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isBanksLoading }" />
        </button>
      </div>
    </div>

    <div class="flex items-center justify-between gap-3 rounded-2xl border border-primary/20 bg-primary/5 px-4 py-3">
      <div class="flex items-center gap-3">
        <div class="rounded-xl bg-primary/10 p-2 text-primary"><Building class="w-5 h-5" /></div>
        <div>
          <p class="text-[10px] font-bold uppercase tracking-wider text-primary">Konteks laporan</p>
          <p class="font-black text-slate-900">Cabang {{ currentBranchName }}</p>
        </div>
      </div>
      <span class="text-xs font-semibold text-slate-500">PT: {{ selectedCompanyName }}</span>
    </div>

    <!-- Saldo Overview Section: Kas Besar, Kas Kecil, Rekening Bank Kelolaan -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
      <!-- 1. Kas Besar (Penerimaan Penjualan) -->
      <div class="bg-gradient-to-br from-amber-500 to-amber-600 rounded-2xl p-5 text-white shadow-md relative overflow-hidden flex flex-col justify-between">
        <div class="absolute -right-4 -bottom-4 w-28 h-28 bg-white/10 rounded-full blur-xl pointer-events-none"></div>
        <div>
          <div class="flex justify-between items-start">
            <div>
              <span class="inline-block px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-white/20 uppercase tracking-wider mb-2">
                {{ selectedCompanyName }}
              </span>
              <h3 class="text-lg font-bold">Kas Besar (Sales)</h3>
            </div>
            <div class="p-2.5 bg-white/15 rounded-xl backdrop-blur-sm">
              <Coins class="w-6 h-6" />
            </div>
          </div>
          <p class="text-xs text-amber-100 mt-1">Uang fisik penerimaan penjualan tunai di cabang.</p>

          <!-- Detail PT ditampilkan setelah PT dipilih agar dashboard tetap ringkas. -->
          <div v-if="selectedCompanyId && companyBalances.length > 0" class="mt-3 space-y-1 bg-black/10 p-2.5 rounded-xl border border-white/10">
            <p class="text-[10px] font-bold text-amber-200 uppercase tracking-wider">Saldo PT terpilih:</p>
            <div v-for="cb in companyBalances" :key="cb.company_id" class="flex justify-between items-center text-xs text-amber-50">
              <span class="truncate max-w-[150px]">{{ cb.company_name }}</span>
              <span class="font-bold">{{ formatRupiah(cb.manual_cash_balance) }}</span>
            </div>
          </div>
        </div>

        <div class="mt-4 pt-4 border-t border-white/20">
          <p class="text-xs text-amber-100 font-medium mb-1">Saldo Kas Besar Belum Disetor</p>
          <div class="flex items-baseline justify-between mb-3">
            <span class="text-2xl font-black tracking-tight">{{ formatRupiah(cashBalance) }}</span>
            <button @click="openCashBalanceModal()" class="px-2 py-1 bg-white/20 hover:bg-white/30 text-white text-xs font-bold rounded-lg transition-colors flex items-center gap-1 backdrop-blur-sm" title="Atur / Koreksi Saldo Kas Besar PT">
              <Settings class="w-3.5 h-3.5" /> Saldo Awal
            </button>
          </div>
          <div class="grid grid-cols-2 gap-2">
            <button @click="openCreateMutationModal('CASH', 'IN')" class="px-2 py-1.5 bg-white/90 text-amber-900 text-xs font-bold rounded-lg shadow-sm hover:bg-white transition-colors flex items-center justify-center gap-1">
              <ArrowDownRight class="w-3.5 h-3.5 text-emerald-600" /> + Kas Masuk
            </button>
            <button @click="openReportModal" class="px-3 py-1.5 bg-white text-amber-800 text-xs font-bold rounded-lg shadow-sm hover:bg-amber-50 transition-colors flex items-center justify-center gap-1">
              Setor Harian
            </button>
          </div>
        </div>
      </div>

      <!-- 2. Kas Kecil (Petty Cash Operasional) -->
      <div class="bg-gradient-to-br from-emerald-600 to-teal-700 rounded-2xl p-5 text-white shadow-md relative overflow-hidden flex flex-col justify-between">
        <div class="absolute -right-4 -bottom-4 w-28 h-28 bg-white/10 rounded-full blur-xl pointer-events-none"></div>
        <div>
          <div class="flex justify-between items-start">
            <div>
              <span class="inline-block px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-white/20 uppercase tracking-wider mb-2">
                {{ selectedCompanyName }}
              </span>
              <h3 class="text-lg font-bold">Kas Kecil (Petty Cash)</h3>
            </div>
            <div class="p-2.5 bg-white/15 rounded-xl backdrop-blur-sm">
              <Wallet class="w-6 h-6" />
            </div>
          </div>
          <p class="text-xs text-emerald-100 mt-1">Disisihkan untuk operasional & klaim reimburse.</p>

          <div v-if="selectedCompanyId && companyBalances.length > 0" class="mt-3 space-y-1 bg-black/10 p-2.5 rounded-xl border border-white/10">
            <p class="text-[10px] font-bold text-emerald-200 uppercase tracking-wider">Kas kecil PT terpilih:</p>
            <div v-for="cb in companyBalances" :key="cb.company_id" class="flex justify-between items-center text-xs text-emerald-50">
              <span class="truncate max-w-[150px]">{{ cb.company_name }}</span>
              <span class="font-bold">{{ formatRupiah(cb.petty_cash_balance) }}</span>
            </div>
          </div>
        </div>

        <div class="mt-4 pt-4 border-t border-white/20">
          <div class="flex items-center justify-between mb-1">
            <p class="text-xs text-emerald-100 font-medium">Saldo Kas Kecil Brankas</p>
            <span v-if="pettyCashBalance < 500000" class="text-[10px] bg-red-500/80 text-white font-bold px-2 py-0.5 rounded-full">Perlu Top-up</span>
            <span v-else class="text-[10px] bg-emerald-500/80 text-white font-bold px-2 py-0.5 rounded-full">Tersedia</span>
          </div>
          <div class="flex items-baseline justify-between mb-3">
            <span class="text-2xl font-black tracking-tight">{{ formatRupiah(pettyCashBalance) }}</span>
            <button @click="openPettyCashBalanceModal()" class="px-2 py-1 bg-white/20 hover:bg-white/30 text-white text-xs font-bold rounded-lg transition-colors flex items-center gap-1 backdrop-blur-sm" title="Atur Saldo Awal / Koreksi Kas Kecil PT">
              <Settings class="w-3.5 h-3.5" /> Saldo Awal
            </button>
          </div>
          <div class="grid grid-cols-3 gap-1.5">
            <button @click="openCreateMutationModal('PETTY_CASH', 'IN')" class="px-2 py-1.5 bg-white text-emerald-800 text-[11px] font-bold rounded-lg shadow-sm hover:bg-emerald-50 transition-colors flex items-center justify-center gap-1">
              <ArrowDownRight class="w-3 h-3 text-emerald-600" /> + Masuk
            </button>
            <button @click="openCreateMutationModal('PETTY_CASH', 'OUT')" class="px-2 py-1.5 bg-white/90 text-emerald-900 text-[11px] font-bold rounded-lg shadow-sm hover:bg-white transition-colors flex items-center justify-center gap-1">
              <ArrowUpRight class="w-3 h-3 text-red-600" /> - Keluar
            </button>
            <button @click="currentTab = 'reimbursements'" class="px-2 py-1.5 bg-emerald-800 text-white text-[11px] font-bold rounded-lg shadow-sm hover:bg-emerald-900 transition-colors flex items-center justify-center gap-1">
              <Receipt class="w-3 h-3" /> Reimburse
            </button>
          </div>
        </div>
      </div>

      <!-- 3. Rekening Bank Kelolaan Cabang (Multi-Bank Flexible) -->
      <div class="bg-white rounded-2xl border border-slate-200 p-5 shadow-sm relative overflow-hidden flex flex-col justify-between">
        <div>
          <div class="flex justify-between items-start">
            <div>
              <span class="inline-block px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-100 uppercase tracking-wider mb-2">
                {{ filteredBanks.length }} Rekening Kelolaan
              </span>
              <h3 class="text-lg font-bold text-slate-900">Rekening Bank Cabang {{ currentBranchName }}</h3>
            </div>
            <button 
              @click="openAddBankModal"
              class="px-2.5 py-1.5 bg-primary/10 hover:bg-primary/20 text-primary text-xs font-bold rounded-xl transition-colors flex items-center gap-1 shadow-sm"
              title="Tambah Akun Rekening Bank yang Dikelola Kasir"
            >
              <Plus class="w-3.5 h-3.5" /> + Rekening
            </button>
          </div>
          <p class="text-xs text-slate-500 mt-1">Rekening yang didaftarkan cabang ini untuk setoran dan pengeluaran operasional.</p>
        </div>

        <div class="mt-4 pt-4 border-t border-slate-100 space-y-2.5 max-h-[190px] overflow-y-auto pr-1">
          <div v-if="filteredBanks.length > 0" class="space-y-2">
            <div 
              v-for="b in filteredBanks" 
              :key="b.id" 
              class="p-2.5 rounded-xl border border-slate-100 bg-slate-50/70 hover:bg-slate-50 transition-all space-y-1.5"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-1.5">
                  <span v-if="b.company" class="px-1.5 py-0.5 bg-blue-100 text-blue-800 text-[10px] font-bold rounded">
                    {{ b.company.name || b.company.code }}
                  </span>
                  <span class="text-xs font-black text-slate-900">{{ b.bank_name }}</span>
                  <span class="text-xs font-mono text-slate-700">{{ b.account_number }}</span>
                </div>
                <div class="flex items-center gap-1">
                  <button @click="openBalanceModal(b)" class="text-[11px] font-bold text-primary hover:underline px-1 py-0.5" title="Update Saldo">
                    Update Saldo
                  </button>
                  <button @click="openEditBankModal(b)" class="text-slate-400 hover:text-slate-600 p-1" title="Edit Detail Rekening">
                    <Edit2 class="w-3 h-3" />
                  </button>
                </div>
              </div>
              <div class="flex items-center justify-between text-xs">
                <span class="text-slate-500 text-[11px] truncate max-w-[130px]">a.n. {{ b.account_holder }}</span>
                <span class="font-extrabold text-blue-700">{{ formatRupiah(b.manual_balance || 0) }}</span>
              </div>
            </div>
          </div>
          <div v-else class="text-xs text-slate-600 bg-slate-50 p-3 rounded-xl border border-slate-200 space-y-2">
            <div class="flex items-center gap-2">
              <AlertTriangle class="w-4 h-4 text-amber-500 shrink-0" />
              <span>Belum ada rekening bank terdaftar di cabang ini.</span>
            </div>
            <button @click="openAddBankModal" class="w-full py-1.5 bg-primary text-white text-xs font-bold rounded-lg shadow-sm hover:bg-primary/90 transition-colors flex items-center justify-center gap-1">
              <Plus class="w-3.5 h-3.5" /> Tambah Rekening Pertama
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Navigation Tabs -->
    <div class="flex border-b border-slate-200 gap-2">
      <button 
        @click="currentTab = 'reports'"
        class="px-5 py-3 font-bold text-sm transition-colors border-b-2 flex items-center gap-2"
        :class="currentTab === 'reports' ? 'border-primary text-primary' : 'border-transparent text-slate-500 hover:text-slate-800'"
      >
        <FileText class="w-4 h-4" />
        Buku Kas & Setoran
      </button>

      <button 
        @click="currentTab = 'reimbursements'"
        class="px-5 py-3 font-bold text-sm transition-colors border-b-2 flex items-center gap-2"
        :class="currentTab === 'reimbursements' ? 'border-primary text-primary' : 'border-transparent text-slate-500 hover:text-slate-800'"
      >
        <Receipt class="w-4 h-4" />
        Reimbursement
      </button>

      <button 
        @click="currentTab = 'mutations'"
        class="px-5 py-3 font-bold text-sm transition-colors border-b-2 flex items-center gap-2"
        :class="currentTab === 'mutations' ? 'border-primary text-primary' : 'border-transparent text-slate-500 hover:text-slate-800'"
      >
        <RefreshCw class="w-4 h-4" />
        Mutasi Kas & Bank
      </button>
    </div>

    <!-- TAB 1: LAPORAN HARIAN KAS BESAR -->
    <div v-if="currentTab === 'reports'" class="space-y-4 animate-in fade-in duration-300">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-lg font-bold text-slate-800 flex items-center gap-2">
            <FileText class="w-5 h-5 text-slate-500" /> Riwayat Laporan Harian Kas Besar
          </h2>
          <p class="text-xs text-slate-500 mt-0.5">Mencatat penerimaan tunai, penyisihan kas kecil operasional, dan bukti setoran bank/pabrik.</p>
        </div>
        <div class="flex items-center gap-3">
          <div class="flex items-center gap-2 bg-white border border-slate-200 rounded-xl px-3 py-1.5 shadow-sm">
            <Calendar class="w-4 h-4 text-slate-400" />
            <input
              type="date"
              v-model="reportFilterDate"
              title="Filter Tanggal Laporan"
              class="text-xs font-bold text-slate-700 outline-none bg-transparent cursor-pointer"
            />
            <button
              v-if="reportFilterDate"
              @click="reportFilterDate = ''"
              title="Reset Tanggal"
              class="text-slate-400 hover:text-slate-600 p-0.5"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>
          <button
            v-if="['cashier', 'admin_kasir', 'cashier_pusat', 'manager', 'superadmin'].includes(authStore.user?.role || '')"
            @click="openFactoryInvoiceDepositModal"
            class="px-4 py-2 bg-indigo-700 text-white rounded-xl hover:bg-indigo-800 transition-colors flex items-center gap-2 font-bold shadow-sm text-sm"
          >
            <Receipt class="w-4 h-4" /> Setor Faktur ke Pabrik
          </button>
          <button
            v-if="['cashier', 'admin_kasir', 'cashier_pusat', 'manager', 'superadmin'].includes(authStore.user?.role || '')"
            @click="openFactoryNonInvoiceDepositModal"
            class="px-4 py-2 bg-amber-700 text-white rounded-xl hover:bg-amber-800 transition-colors flex items-center gap-2 font-bold shadow-sm text-sm"
          >
            <Receipt class="w-4 h-4" /> Setor Dana Nonfaktur
          </button>
          <button
            @click="openReportModal"
            class="px-4 py-2 bg-primary text-white rounded-xl hover:bg-primary/90 transition-colors flex items-center gap-2 font-bold shadow-sm text-sm"
          >
            <Plus class="w-4 h-4" /> Buat Laporan Harian
          </button>
        </div>
      </div>

      <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200">
            <thead class="bg-slate-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Tanggal & PT</th>
                <th scope="col" class="px-6 py-3 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Penjualan Tunai</th>
                <th scope="col" class="px-6 py-3 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Ke Kas Kecil</th>
                <th scope="col" class="px-6 py-3 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Disetorkan</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Tujuan Setoran</th>
                <th scope="col" class="px-6 py-3 text-center text-xs font-semibold text-slate-500 uppercase tracking-wider">Bukti</th>
                <th scope="col" class="px-6 py-3 text-center text-xs font-semibold text-slate-500 uppercase tracking-wider">Aksi</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-slate-200">
              <tr v-for="item in reports" :key="item.id" class="hover:bg-slate-50 transition-colors">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="font-bold text-slate-900">{{ formatDate(item.report_date) }}</div>
                  <div class="flex items-center gap-1.5 mt-1">
                    <span v-if="item.company" class="px-2 py-0.5 bg-blue-50 text-blue-700 text-[10px] font-bold rounded border border-blue-200">
                      {{ item.company.name }}
                    </span>
                    <span class="text-xs text-slate-500">{{ item.notes || '-' }}</span>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right">
                  <div class="font-bold text-slate-800">{{ formatRupiah(item.total_cash_sales) }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right">
                  <span v-if="item.allocated_to_petty_cash > 0" class="inline-flex items-center gap-1 font-bold text-emerald-700 bg-emerald-50 px-2.5 py-1 rounded-full text-xs border border-emerald-200">
                    +{{ formatRupiah(item.allocated_to_petty_cash) }}
                  </span>
                  <span v-else class="text-xs text-slate-400">-</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right">
                  <div class="font-extrabold text-blue-600">{{ formatRupiah(item.deposited_amount) }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm">
                    <span v-if="item.deposit_destination_type === 'FACTORY_KAS'" class="px-2 py-0.5 text-[10px] font-extrabold rounded-full bg-amber-100 text-amber-800 border border-amber-200 mb-1 inline-block">
                      Kas Pabrik / Pusat
                    </span>
                    <span v-else class="px-2 py-0.5 text-[10px] font-extrabold rounded-full bg-blue-100 text-blue-800 border border-blue-200 mb-1 inline-block">
                      Bank PT Cabang
                    </span>
                    <div class="font-semibold text-slate-900">{{ item.target_bank_name || 'Bank/Kas' }}</div>
                    <div class="text-xs text-slate-500">{{ item.target_account_number }} ({{ item.target_account_name }})</div>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  <button 
                    v-if="item.deposit_proof_url" 
                    @click="openProof(item.deposit_proof_url, `Bukti Setoran Kas Besar (${formatDate(item.report_date)})`, item.company?.name || item.target_bank_name)" 
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-bold transition-colors"
                  >
                    <Eye class="w-3.5 h-3.5" /> Lihat
                  </button>
                  <span v-else class="text-slate-400 text-xs">-</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  <div class="flex items-center justify-center gap-1.5">
                    <button
                      @click="openEditReportModal(item)"
                      class="p-1.5 text-slate-500 hover:text-amber-600 hover:bg-amber-50 rounded-lg transition-colors"
                      title="Edit Laporan Harian"
                    >
                      <Edit2 class="w-4 h-4" />
                    </button>
                    <button
                      @click="deleteReport(item)"
                      class="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                      title="Hapus Laporan Harian"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="reports.length === 0">
                <td colspan="7" class="px-6 py-12 text-center text-slate-500">
                  <FileText class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                  <p class="font-bold text-slate-800">Belum ada laporan kasir harian</p>
                  <p class="text-xs text-slate-400 mt-1">Klik tombol "Buat Laporan Harian" untuk memulai.</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: PENCAIRAN REIMBURSEMENT (KAS & BANK PT) -->
    <div v-if="currentTab === 'reimbursements'" class="space-y-4 animate-in fade-in duration-300">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-lg font-bold text-slate-800 flex items-center gap-2">
            <Receipt class="w-5 h-5 text-emerald-600" /> Pencairan Klaim Reimbursement Karyawan / Sales
          </h2>
          <p class="text-xs text-slate-500 mt-0.5">Pencairan klaim yang telah disetujui dapat dipotongkan ke Kas Kecil, Kas Besar, atau Rekening Bank milik PT salesman.</p>
        </div>

        <div class="flex items-center gap-3">
          <div class="flex rounded-xl bg-slate-100 p-1 text-xs font-bold">
            <button 
              @click="reimbursementFilterStatus = 'APPROVED'" 
              class="px-3 py-1.5 rounded-lg transition-colors"
              :class="reimbursementFilterStatus === 'APPROVED' ? 'bg-white text-emerald-700 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
            >
              Siap Dicairkan
            </button>
            <button 
              @click="reimbursementFilterStatus = 'DISBURSED'" 
              class="px-3 py-1.5 rounded-lg transition-colors"
              :class="reimbursementFilterStatus === 'DISBURSED' ? 'bg-white text-blue-700 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
            >
              Riwayat Dicairkan
            </button>
            <button 
              @click="reimbursementFilterStatus = 'ALL'" 
              class="px-3 py-1.5 rounded-lg transition-colors"
              :class="reimbursementFilterStatus === 'ALL' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
            >
              Semua
            </button>
          </div>
        </div>
      </div>

      <!-- Informative banner -->
      <div class="bg-emerald-50 border border-emerald-200 rounded-2xl p-4 flex items-center justify-between text-emerald-900 text-sm">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-emerald-100 rounded-xl text-emerald-700">
            <Wallet class="w-5 h-5" />
          </div>
          <div>
            <p class="font-bold">Saldo Kas Kecil Siap Digunakan: {{ formatRupiah(pettyCashBalance) }}</p>
            <p class="text-xs text-emerald-700">Kasir dapat fleksibel memilih sumber dana: Kas Kecil PT, Kas Besar PT, atau Transfer Bank PT.</p>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200">
            <thead class="bg-slate-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Karyawan & PT</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Judul & Keterangan</th>
                <th scope="col" class="px-6 py-3 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Nominal</th>
                <th scope="col" class="px-6 py-3 text-center text-xs font-semibold text-slate-500 uppercase tracking-wider">Nota Bukti</th>
                <th scope="col" class="px-6 py-3 text-center text-xs font-semibold text-slate-500 uppercase tracking-wider">Status Pencairan</th>
                <th scope="col" class="px-6 py-3 text-center text-xs font-semibold text-slate-500 uppercase tracking-wider">Aksi</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-slate-200">
              <tr v-for="item in reimbursements" :key="item.id" class="hover:bg-slate-50 transition-colors">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="font-bold text-slate-900">{{ item.employee ? `${item.employee.first_name} ${item.employee.last_name || ''}` : '-' }}</div>
                  <div class="flex items-center gap-1.5 mt-0.5">
                    <span v-if="item.employee?.company" class="px-1.5 py-0.5 bg-blue-50 text-blue-700 text-[10px] font-bold rounded border border-blue-200">
                      {{ item.employee.company.name }}
                    </span>
                    <span class="text-xs text-slate-500">{{ item.employee?.job_position?.title || 'Karyawan' }}</span>
                  </div>
                </td>
                <td class="px-6 py-4">
                  <div class="font-semibold text-slate-900 text-sm flex items-center gap-2 flex-wrap">
                    <span>{{ item.title }}</span>
                    <span
                      v-if="item.category === 'BANNER_INSTALLATION'"
                      class="px-2 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 text-[10px] font-black rounded-md inline-flex items-center gap-1"
                    >
                      🏷️ Pasang Spanduk ({{ item.banner_count || item.banner_orders?.length || 0 }} Unit)
                    </span>
                    <span
                      v-else-if="item.category === 'BBM' || item.start_km != null"
                      class="px-2 py-0.5 bg-amber-50 text-amber-800 border border-amber-200 text-[10px] font-black rounded-md inline-flex items-center gap-1"
                    >
                      ⛽ Klaim BBM
                    </span>
                  </div>
                  <div class="text-xs text-slate-500 line-clamp-2 mt-0.5">{{ item.description || '-' }}</div>
                  <div class="text-[10px] text-slate-400 mt-1 flex items-center gap-2 flex-wrap">
                    <span>Diajukan: {{ formatDate(item.created_at) }}</span>
                    <span v-if="item.unit_rate > 0" class="text-emerald-700 font-bold">
                      • Tarif: {{ formatRupiah(item.unit_rate) }}/spanduk
                    </span>
                    <span v-if="item.start_km != null && item.end_km != null" class="text-amber-800 font-bold bg-amber-50 border border-amber-200 px-1.5 py-0.5 rounded">
                      KM: {{ Number(item.start_km).toLocaleString('id-ID') }} → {{ Number(item.end_km).toLocaleString('id-ID') }} ({{ item.total_km || (item.end_km - item.start_km) }} KM)
                    </span>
                    <span v-if="item.vehicle_plate_number" class="text-slate-600 font-bold bg-slate-100 px-1.5 py-0.5 rounded">
                      🚗 {{ item.vehicle_plate_number }}
                    </span>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right">
                  <div class="font-bold text-slate-900 text-base">{{ formatRupiah(item.amount) }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  <button v-if="item.attachment_url" @click="openProof(item.attachment_url, 'Nota Klaim Reimbursement', item.title + (item.employee ? ' (' + item.employee.first_name + ')' : ''))" class="inline-flex items-center gap-1 px-2.5 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-semibold transition-colors">
                    <Eye class="w-3.5 h-3.5" /> Struk
                  </button>
                  <span v-else class="text-xs text-slate-400">-</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  <span v-if="item.disbursement_status === 'DISBURSED'" class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold bg-emerald-100 text-emerald-800 border border-emerald-200">
                    <CheckCircle2 class="w-3.5 h-3.5" /> Dicairkan ({{ item.disbursed_via === 'PETTY_CASH' ? 'Kas Kecil' : (item.disbursed_via === 'CASH' ? 'Kas Tunai' : 'Transfer') }})
                  </span>
                  <span v-else-if="item.status === 'APPROVED'" class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-bold bg-amber-100 text-amber-800 border border-amber-200">
                    Siap Dicairkan
                  </span>
                  <span v-else class="px-2.5 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-600">
                    {{ item.status }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  <button 
                    v-if="item.status === 'APPROVED' && item.disbursement_status !== 'DISBURSED'"
                    @click="openDisburseModal(item)"
                    class="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition-colors shadow-sm flex items-center gap-1.5 mx-auto"
                  >
                    <Wallet class="w-3.5 h-3.5" /> Cairkan Uang
                  </button>
                  <button 
                    v-else-if="item.disbursement_proof_url"
                    @click="openProof(item.disbursement_proof_url, 'Bukti Serah Terima Pencairan', item.title)"
                    class="px-2.5 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-medium transition-colors"
                  >
                    Bukti Serah Terima
                  </button>
                  <span v-else class="text-xs text-slate-400">-</span>
                </td>
              </tr>
              <tr v-if="reimbursements.length === 0">
                <td colspan="6" class="px-6 py-12 text-center text-slate-500">
                  <Receipt class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                  <p class="font-bold text-slate-800">Tidak ada pengajuan reimbursement</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 3: RIWAYAT MUTASI & BUKU KAS -->
    <div v-if="currentTab === 'mutations'" class="space-y-4 animate-in fade-in duration-300">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-lg font-bold text-slate-800 flex items-center gap-2">
            <RefreshCw class="w-5 h-5 text-slate-500" /> Buku Mutasi Dana Per PT (Kas Kecil, Kas Besar, Bank)
          </h2>
          <p class="text-xs text-slate-500 mt-0.5">Audit log seluruh dana masuk & keluar per PT secara realtime.</p>
          <p class="text-xs text-indigo-700 mt-1">Untuk setoran faktur pabrik, pilih faktur yang sudah disetujui dan rekening cabang tujuan, lalu kirim bukti. Admin Kasir cabang memverifikasi sebelum saldo rekening diperbarui.</p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <!-- Filter Sumber Dana -->
          <select v-model="mutationFundTypeFilter" class="text-xs font-bold px-3 py-1.5 border border-slate-200 rounded-xl bg-white focus:outline-none">
            <option value="">Semua Sumber Dana</option>
            <option value="PETTY_CASH">Kas Kecil (Petty Cash)</option>
            <option value="CASH">Kas Tunai / Kas Besar</option>
            <option value="BANK">Rekening Bank</option>
          </select>

          <!-- Filter Kategori -->
          <select v-model="mutationCategoryFilter" class="text-xs font-bold px-3 py-1.5 border border-slate-200 rounded-xl bg-white focus:outline-none">
            <option value="">Semua Kategori</option>
            <option value="INITIAL_BALANCE">Saldo Awal / Koreksi</option>
            <option value="PETTY_CASH_TOPUP">Top-up Kas Kecil</option>
            <option value="REIMBURSEMENT">Pencairan Reimbursement</option>
            <option value="SALES_DEPOSIT">Setoran Kas Besar</option>
            <option value="FACTORY_INVOICE_DEPOSIT">Setoran Faktur ke Pabrik</option>
            <option value="FACTORY_NON_INVOICE_DEPOSIT">Setoran Dana Nonfaktur ke Pabrik</option>
            <option value="OPERATIONAL_EXPENSE">Operasional Umum</option>
          </select>

          <div class="flex items-center gap-2 bg-white border border-slate-200 rounded-xl px-3 py-1.5 shadow-sm">
            <Calendar class="w-4 h-4 text-slate-400" />
            <input
              type="date"
              v-model="mutationFilterDate"
              title="Filter Tanggal Mutasi"
              class="text-xs font-bold text-slate-700 outline-none bg-transparent cursor-pointer"
            />
            <button
              v-if="mutationFilterDate"
              @click="mutationFilterDate = ''"
              title="Reset Tanggal"
              class="text-slate-400 hover:text-slate-600 p-0.5"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>

          <button 
            @click="openCreateMutationModal('PETTY_CASH', 'IN')"
            class="px-4 py-2 bg-slate-900 text-white rounded-xl hover:bg-slate-800 transition-colors flex items-center gap-2 font-bold shadow-sm text-sm"
          >
            <Plus class="w-4 h-4" /> Catat Mutasi
          </button>
        </div>
      </div>

      <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200">
            <thead class="bg-slate-50">
              <tr>
                <th scope="col" class="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Tanggal, PT & Sumber</th>
                <th scope="col" class="px-6 py-3 text-center text-xs font-semibold text-slate-500 uppercase tracking-wider">Kategori & Jenis</th>
                <th scope="col" class="px-6 py-3 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Nominal</th>
                <th scope="col" class="px-6 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wider">Keterangan / Terkait</th>
                <th scope="col" class="px-6 py-3 text-center text-xs font-semibold text-slate-500 uppercase tracking-wider">Bukti</th>
                <th scope="col" class="px-6 py-3 text-center text-xs font-semibold text-slate-500 uppercase tracking-wider">Aksi</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-slate-200">
              <tr v-for="item in mutations" :key="item.id" class="hover:bg-slate-50 transition-colors">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="font-bold text-slate-900">{{ formatDate(item.mutation_date) }}</div>
                  <div class="flex flex-wrap items-center gap-1.5 mt-1">
                    <span v-if="item.company" class="px-2 py-0.5 rounded bg-blue-50 text-blue-700 font-bold text-[10px] border border-blue-200">
                      {{ item.company.name }}
                    </span>
                    <span v-if="item.fund_type === 'PETTY_CASH'" class="px-2 py-0.5 rounded-full bg-emerald-100 text-emerald-800 font-bold border border-emerald-200 text-[10px]">
                      Kas Kecil
                    </span>
                    <span v-else-if="item.fund_type === 'CASH'" class="px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 font-bold border border-amber-200 text-[10px]">
                      Kas Besar
                    </span>
                    <span v-else class="px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 font-medium border border-blue-100 text-[10px]">
                      Bank: {{ item.bank_account?.bank_name }} ({{ item.bank_account?.account_number }})
                    </span>
                  </div>
                  <p v-if="isFactoryDepositItem(item)" class="mt-1 text-[10px] font-semibold text-blue-700">Tujuan: {{ item.bank_account?.branch?.name || item.branch?.name || 'Cabang tujuan' }} · {{ item.bank_account?.bank_name || '-' }} {{ item.bank_account?.account_number || '' }}</p>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  <div class="mb-1">
                    <span v-if="item.category === 'INITIAL_BALANCE'" class="px-2 py-0.5 text-[10px] font-extrabold rounded-full bg-teal-100 text-teal-800 border border-teal-200">
                      Saldo Awal / Koreksi
                    </span>
                    <span v-else-if="item.category === 'REIMBURSEMENT'" class="px-2 py-0.5 text-[10px] font-extrabold rounded-full bg-purple-100 text-purple-800 border border-purple-200">
                      Reimbursement
                    </span>
                    <span v-else-if="item.category === 'PETTY_CASH_TOPUP'" class="px-2 py-0.5 text-[10px] font-extrabold rounded-full bg-emerald-100 text-emerald-800 border border-emerald-200">
                      Top-up Kas Kecil
                    </span>
                    <span v-else-if="item.category === 'SALES_DEPOSIT'" class="px-2 py-0.5 text-[10px] font-extrabold rounded-full bg-blue-100 text-blue-800 border border-blue-200">
                      Setoran Penjualan
                    </span>
                    <span v-else-if="item.category === 'FACTORY_INVOICE_DEPOSIT'" class="px-2 py-0.5 text-[10px] font-extrabold rounded-full bg-indigo-100 text-indigo-800 border border-indigo-200">
                      Setoran Faktur ke Pabrik
                    </span>
                    <span v-else-if="item.category === 'FACTORY_NON_INVOICE_DEPOSIT'" class="px-2 py-0.5 text-[10px] font-extrabold rounded-full bg-amber-100 text-amber-800 border border-amber-200">
                      Setoran Dana Nonfaktur
                    </span>
                    <span v-else class="px-2 py-0.5 text-[10px] font-medium rounded-full bg-slate-100 text-slate-600">
                      Operasional
                    </span>
                  </div>
                  <span v-if="item.mutation_type === 'IN'" class="inline-flex items-center gap-1 text-green-700 bg-green-50 px-2 py-0.5 rounded-full text-xs font-bold border border-green-200">
                    <ArrowDownRight class="w-3 h-3" /> Masuk
                  </span>
                  <span v-else class="inline-flex items-center gap-1 text-red-700 bg-red-50 px-2 py-0.5 rounded-full text-xs font-bold border border-red-200">
                    <ArrowUpRight class="w-3 h-3" /> Keluar
                  </span>
                  <div v-if="isFactoryDepositItem(item)" class="mt-1">
                    <span :class="item.status_setoran === 'APPROVED' ? 'bg-emerald-100 text-emerald-800' : (item.status_setoran === 'REJECTED' ? 'bg-red-100 text-red-800' : 'bg-amber-100 text-amber-800')" class="rounded-full px-2 py-0.5 text-[10px] font-bold">
                      {{ item.status_setoran === 'APPROVED' ? 'DITERIMA ADMIN CABANG' : (item.status_setoran === 'REJECTED' ? 'DITOLAK' : 'MENUNGGU ADMIN CABANG') }}
                    </span>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right">
                  <div class="font-black text-slate-900 text-base">{{ formatRupiah(item.amount) }}</div>
                </td>
                <td class="px-6 py-4">
                  <p class="text-sm font-medium text-slate-800">{{ item.notes || '-' }}</p>
                  <p v-if="item.invoice_items?.length" class="mt-1 text-[11px] text-slate-600">Faktur: {{ mutationInvoiceLabels(item) }}</p>
                  <p v-if="isFactoryDepositItem(item)" class="mt-1 text-[11px] text-slate-600">Alokasi faktur: {{ formatRupiah(mutationInvoiceTotal(item)) }} · Dana nonfaktur: {{ formatRupiah(mutationNonInvoiceAmount(item)) }}</p>
                  <p v-if="item.fund_origin" class="mt-1 text-[11px] text-amber-700">Asal dana nonfaktur: {{ item.fund_origin === 'OPENING_BALANCE' ? 'Saldo awal' : item.fund_origin === 'CASH_ADJUSTMENT' ? 'Koreksi kas' : 'Dana nonpenjualan lainnya' }}</p>
                  <p v-if="isFactoryDepositItem(item) && item.status_setoran_note" class="mt-1 text-[11px] text-slate-500">Verifikasi: {{ item.status_setoran_note }}</p>
                  <p v-if="item.reimbursement?.employee" class="text-xs text-primary font-semibold mt-0.5">
                    Karyawan: {{ item.reimbursement.employee.first_name }} {{ item.reimbursement.employee.last_name || '' }}
                  </p>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  <button v-if="item.proof_url" @click="openProof(item.proof_url, 'Bukti Mutasi Kas & Bank', item.notes || formatDate(item.mutation_date))" class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-bold transition-colors">
                    <Eye class="w-3.5 h-3.5" /> Lihat
                  </button>
                  <span v-else class="text-slate-400 text-xs">-</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  <div class="flex items-center justify-center gap-1.5">
                    <button v-if="isFactoryDepositItem(item) && item.status_setoran === 'PENDING' && isBranchDepositReviewer" @click="reviewFactoryDeposit(item, 'approve')" class="rounded-lg bg-emerald-50 p-1.5 text-emerald-700 hover:bg-emerald-100" title="Terima setoran">
                      <CheckCircle class="h-4 w-4" />
                    </button>
                    <button v-if="isFactoryDepositItem(item) && item.status_setoran === 'PENDING' && isBranchDepositReviewer" @click="reviewFactoryDeposit(item, 'reject')" class="rounded-lg bg-red-50 p-1.5 text-red-700 hover:bg-red-100" title="Tolak setoran">
                      <X class="h-4 w-4" />
                    </button>
                    <button 
                      v-if="!isFactoryDepositItem(item)"
                      @click="editMutation(item)" 
                      class="p-1.5 text-slate-500 hover:text-primary hover:bg-primary/10 rounded-lg transition-colors" 
                      title="Edit Mutasi"
                    >
                      <Edit2 class="w-4 h-4" />
                    </button>
                    <button 
                      v-if="!isFactoryDepositItem(item) || item.status_setoran !== 'APPROVED'"
                      @click="deleteMutation(item)" 
                      :disabled="deletingMutationId === item.id"
                      class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed" 
                      title="Hapus Mutasi"
                    >
                      <RefreshCw v-if="deletingMutationId === item.id" class="w-4 h-4 animate-spin text-red-500" />
                      <Trash2 v-else class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="mutations.length === 0">
                <td colspan="6" class="px-6 py-12 text-center text-slate-500">
                  <RefreshCw class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                  <p class="font-bold text-slate-800">Belum ada mutasi</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- MODAL: TAMBAH / EDIT PROFIL REKENING BANK -->
    <div v-if="showAddEditBankModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showAddEditBankModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between bg-blue-50/50">
          <div>
            <h3 class="text-lg font-bold text-slate-900">{{ isEditingBank ? 'Edit Profil Rekening Bank' : 'Tambah Rekening Bank Kelolaan' }}</h3>
            <p class="text-xs text-slate-500">Kasir dapat mengelola lebih dari 1 rekening bank di cabangnya.</p>
          </div>
          <button @click="showAddEditBankModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">✕</button>
        </div>

        <div class="p-6 space-y-4">
          <!-- PT pemilik rekening hanya informasi tambahan; pilihan setoran berdasarkan cabang. -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">Perusahaan (PT) Pemilik Rekening (Opsional)</label>
            <select v-model="bankForm.company_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 bg-white font-semibold text-sm">
              <option value="">Tanpa PT khusus</option>
              <option v-for="c in branchCompanies" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>

          <!-- Nama Bank & Nomor Rekening -->
          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Nama Bank</label>
              <input v-model="bankForm.bank_name" type="text" placeholder="Contoh: BCA, Mandiri, BRI" class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm font-bold" />
            </div>
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Nomor Rekening</label>
              <input v-model="bankForm.account_number" type="text" placeholder="1234567890" class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm font-mono font-bold" />
            </div>
          </div>

          <!-- Nama Pemilik Rekening (Atas Nama) -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Atas Nama (Account Holder)</label>
            <input v-model="bankForm.account_holder" type="text" placeholder="Perusahaan Demo" class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm font-semibold" />
          </div>

          <!-- Saldo Awal Rekening -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Saldo Awal (Rp)</label>
            <input v-model.number="bankForm.manual_balance" type="number" min="0" placeholder="0" class="w-full px-3 py-2.5 border border-blue-200 rounded-xl text-lg font-black text-blue-900 bg-blue-50/30" />
          </div>

          <!-- Catatan / Keterangan -->
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-slate-700">Keterangan Tambahan</label>
            <input v-model="bankForm.notes" type="text" placeholder="Rekening penampung sales cabang..." class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm" />
          </div>

          <label v-if="isEditingBank" class="flex items-center gap-2 text-sm text-slate-700">
            <input v-model="bankForm.is_active" type="checkbox" />
            Rekening aktif (hapus centang untuk menonaktifkan rekening yang sudah dipakai)
          </label>
        </div>

        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-between items-center">
          <div>
            <button v-if="isEditingBank" @click="deleteBankAccount(bankForm)" type="button" class="px-3 py-2 text-red-600 hover:bg-red-50 text-xs font-bold rounded-xl transition-colors">
              Hapus Rekening
            </button>
          </div>
          <div class="flex gap-2">
            <button @click="showAddEditBankModal = false" :disabled="isSavingBank" class="px-4 py-2 font-bold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">Batal</button>
            <button @click="saveBankAccount" :disabled="isSavingBank" class="px-5 py-2 bg-primary text-white font-bold rounded-xl hover:bg-primary/90 transition-colors flex items-center gap-2 shadow-sm text-sm disabled:opacity-50">
              <span v-if="isSavingBank" class="animate-spin w-4 h-4 border-2 border-white/20 border-t-white rounded-full"></span>
              <Save v-else class="w-4 h-4" /> 
              {{ isSavingBank ? 'Menyimpan...' : (isEditingBank ? 'Perbarui' : 'Simpan Rekening') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- MODAL: BUAT / EDIT LAPORAN HARIAN KASIR (KAS BESAR & KAS KECIL) -->
    <div v-if="showReportModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showReportModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-xl w-full max-w-xl overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div>
            <h3 class="text-lg font-bold text-slate-900">{{ isEditingReport ? 'Edit Laporan Kasir Harian' : 'Buat Laporan Kasir Harian' }}</h3>
            <p class="text-xs text-slate-500">{{ isEditingReport ? 'Perbarui rincian penjualan, alokasi kas kecil, atau bukti setoran.' : 'Penerimaan penjualan, alokasi kas kecil, dan setoran bank/pabrik.' }}</p>
          </div>
          <button @click="showReportModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">✕</button>
        </div>
        
        <div class="p-6 space-y-5 max-h-[80vh] overflow-y-auto">
          <!-- Tanggal & PT Penjualan -->
          <div class="space-y-4">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <!-- Tanggal Laporan / Penjualan -->
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
                  <Calendar class="w-3.5 h-3.5 text-primary" />
                  <span>Tanggal Penjualan / Setor</span>
                </label>
                <input
                  type="date"
                  v-model="reportForm.report_date"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white font-bold text-sm text-slate-800"
                />
              </div>

              <!-- Perusahaan (PT) -->
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
                  <Building class="w-3.5 h-3.5 text-primary" />
                  <span>Perusahaan (PT)</span>
                </label>
                <select v-model="reportForm.company_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white font-medium text-sm">
                  <option value="">Semua Perusahaan (Gabungan)</option>
                  <option v-for="c in branchCompanies" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
            </div>
            
            <!-- Ringkasan Penjualan & Input Nilai -->
            <div class="bg-amber-50/70 border border-amber-200/80 rounded-2xl p-4 space-y-3">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 text-amber-900 font-bold text-sm">
                  <Coins class="w-4 h-4 text-amber-600" />
                  <span>Penerimaan Penjualan (Persetujuan Kasir)</span>
                  <RefreshCw v-if="isFetchingSummary" class="w-3.5 h-3.5 text-amber-600 animate-spin" />
                </div>
                <button
                  v-if="reportForm.total_cash_sales !== autoApprovedCashSales || reportForm.total_transfer_sales !== autoApprovedTransferSales"
                  type="button"
                  @click="resetToAutoValues"
                  class="text-[11px] font-bold text-amber-800 hover:text-amber-900 bg-amber-200/70 hover:bg-amber-200 px-2 py-0.5 rounded-lg transition-colors flex items-center gap-1"
                  title="Kembalikan ke angka asli dari persetujuan kasir"
                >
                  <RefreshCw class="w-3 h-3" /> Reset ke Approval
                </button>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <!-- Kas Fisik (Tunai) -->
                <div class="bg-white/90 p-3 rounded-xl border border-amber-200 shadow-sm space-y-1.5">
                  <div class="flex items-center justify-between">
                    <label class="text-[11px] font-bold text-slate-600 uppercase tracking-wider">Total Kas Fisik (Tunai)</label>
                    <span class="text-[10px] font-semibold text-amber-800 bg-amber-100/80 px-1.5 py-0.5 rounded">
                      Approve: {{ formatRupiah(autoApprovedCashSales) }}
                    </span>
                  </div>
                  <input
                    v-model.number="reportForm.total_cash_sales"
                    type="number"
                    min="0"
                    placeholder="0"
                    class="w-full px-3 py-2 border border-amber-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-amber-400 bg-amber-50/30 text-lg font-black text-amber-900"
                  />
                  <p class="text-[10px] text-slate-400">Dapat diedit jika fisik kas berbeda</p>
                </div>

                <!-- Transfer Sales -->
                <div class="bg-white/90 p-3 rounded-xl border border-slate-200 shadow-sm space-y-1.5">
                  <div class="flex items-center justify-between">
                    <label class="text-[11px] font-bold text-slate-600 uppercase tracking-wider">Total Transfer Sales</label>
                    <span class="text-[10px] font-semibold text-slate-700 bg-slate-100 px-1.5 py-0.5 rounded">
                      Approve: {{ formatRupiah(autoApprovedTransferSales) }}
                    </span>
                  </div>
                  <input
                    v-model.number="reportForm.total_transfer_sales"
                    type="number"
                    min="0"
                    placeholder="0"
                    class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/20 bg-slate-50/50 text-lg font-black text-slate-800"
                  />
                  <p class="text-[10px] text-slate-400">Dapat diedit jika mutasi transfer berbeda</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 1: Sisihkan ke Kas Kecil (Petty Cash) - Opsional -->
          <div class="bg-emerald-50/70 border border-emerald-200/80 rounded-2xl p-4 space-y-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2 text-emerald-900 font-bold text-sm">
                <Wallet class="w-4 h-4 text-emerald-600" />
                <span>1. Sisihkan untuk Kas Kecil (Opsional)</span>
              </div>
              <span class="text-xs text-emerald-700 font-semibold">Saldo Brankas: {{ formatRupiah(pettyCashBalance) }}</span>
            </div>
            <p class="text-xs text-emerald-800/80">Jika ada penerimaan tunai kas besar yang ingin disisihkan ke kas kecil (isi 0 jika tidak ada).</p>
            
            <div class="space-y-1">
              <label class="text-xs font-bold text-slate-700">Nominal Disisihkan ke Kas Kecil (Rp)</label>
              <input 
                v-model.number="reportForm.allocated_to_petty_cash" 
                type="number" 
                min="0" 
                placeholder="0" 
                class="w-full px-3 py-2.5 border border-emerald-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-400 bg-white text-base font-bold text-emerald-900"
              />
            </div>
          </div>

          <!-- Step 2: Setoran Sisa Kas Besar -->
          <div class="space-y-4 pt-2 border-t border-slate-100">
            <div class="flex items-center justify-between">
              <h4 class="font-bold text-slate-900 text-sm">2. Setoran Sisa Kas Besar</h4>
              <span class="text-xs font-extrabold text-blue-700 bg-blue-50 px-2.5 py-1 rounded-full border border-blue-200">
                Sisa Disetor: {{ formatRupiah(reportForm.deposited_amount) }}
              </span>
            </div>

            <!-- Pilih Tujuan Setoran: Bank Cabang atau Kas Pabrik -->
            <div class="space-y-2">
              <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">Pilih Tujuan Setoran</label>
              <div class="grid grid-cols-2 gap-3">
                <label 
                  class="flex items-center gap-2.5 p-3 rounded-xl border cursor-pointer transition-all"
                  :class="reportForm.deposit_destination_type === 'BRANCH_BANK' ? 'bg-primary/5 border-primary text-primary font-bold shadow-sm' : 'bg-white border-slate-200 text-slate-600'"
                >
                  <input type="radio" value="BRANCH_BANK" v-model="reportForm.deposit_destination_type" class="hidden" />
                  <Building class="w-4 h-4 shrink-0" />
                  <span class="text-xs">Rekening Bank PT Cabang</span>
                </label>

                <div class="flex items-center gap-2.5 rounded-xl border border-amber-200 bg-amber-50 p-3 text-amber-900">
                  <Coins class="w-4 h-4 shrink-0" />
                  <span class="text-xs">Ke pabrik: gunakan tombol Setor Faktur atau Setor Dana Nonfaktur di Buku Kas & Setoran.</span>
                </div>
              </div>
            </div>

            <!-- Detail Bank jika Bank Cabang -->
            <div v-if="reportForm.deposit_destination_type === 'BRANCH_BANK'" class="space-y-3 bg-slate-50 p-4 rounded-2xl border border-slate-200/80">
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-700">Pilih Rekening Bank Cabang</label>
                <select v-model="reportForm.target_bank_account_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 bg-white font-medium text-sm">
                  <option value="">-- Pilih Rekening Cabang Tersimpan --</option>
                  <option v-for="b in systemBankAccounts" :key="b.id" :value="b.id">
                    {{ b.bank_name }} - {{ b.account_number }} ({{ b.account_holder }})
                  </option>
                </select>
              </div>

              <div class="grid grid-cols-2 gap-3">
                <div class="space-y-1">
                  <label class="text-xs font-medium text-slate-600">Nomor Rekening</label>
                  <input v-model="reportForm.target_account_number" type="text" placeholder="123456789" class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm" />
                </div>
                <div class="space-y-1">
                  <label class="text-xs font-medium text-slate-600">Atas Nama</label>
                  <input v-model="reportForm.target_account_name" type="text" placeholder="Nama Pemilik Rekening" class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm" />
                </div>
              </div>
            </div>

            <!-- Detail Kas Pabrik jika Setor ke Pabrik/Pusat -->
            <div v-if="reportForm.deposit_destination_type === 'FACTORY_KAS'" class="space-y-3 bg-amber-50/50 p-4 rounded-2xl border border-amber-200/70">
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-800 flex items-center gap-1.5">
                  <Coins class="w-4 h-4 text-amber-600" />
                  <span>Pilih Rekening Kas Pabrik / Kas Pusat Tujuan</span>
                </label>
                <select v-model="reportForm.target_kas_pabrik_id" class="w-full px-3 py-2.5 border border-amber-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-amber-300 bg-white font-bold text-sm">
                  <option value="">-- Pilih Kas Pabrik / Bank Pusat --</option>
                  <option v-for="k in kasPabriks" :key="k.id" :value="k.id">
                    {{ k.kas_name }} ({{ k.bank_name ? `${k.bank_name} - ${k.bank_account_number}` : 'Kas Tunai Pabrik' }})
                  </option>
                </select>
                <p class="text-[11px] text-amber-700 mt-1">Gunakan opsi ini jika cabang Anda belum memiliki rekening bank PT sendiri. Setoran ini akan diverifikasi oleh Kasir Pusat.</p>
              </div>
            </div>

            <!-- Upload Bukti Setoran -->
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700 flex items-center gap-1">
                Upload Bukti Setoran / Transfer 
                <span v-if="!isEditingReport || !reportForm.deposit_proof_url" class="text-xs text-red-500 font-bold">*</span>
                <span v-else class="text-[10px] text-slate-400 font-normal">(Opsional jika tidak diganti)</span>
              </label>
              <label class="flex flex-col items-center justify-center w-full h-28 border-2 border-slate-200 border-dashed rounded-2xl cursor-pointer hover:bg-slate-50 transition-colors bg-slate-50/50">
                <div class="flex flex-col items-center justify-center p-3 text-center">
                  <Upload v-if="!selectedReportFile && !reportForm.deposit_proof_url" class="w-6 h-6 mb-1.5 text-slate-400" />
                  <File v-else class="w-6 h-6 mb-1.5 text-primary" />
                  <p class="text-xs text-slate-600 font-bold">
                    {{ selectedReportFile ? selectedReportFile.name : (reportForm.deposit_proof_url ? 'Bukti terlampir (klik untuk ganti berkas)' : 'Klik untuk upload struk setoran bank') }}
                  </p>
                  <p class="text-[10px] text-slate-400 mt-0.5">JPG, PNG, atau PDF (Maks. 10MB)</p>
                </div>
                <input type="file" class="hidden" accept="image/*,application/pdf" @change="handleReportFileChange" />
              </label>
            </div>

            <div class="space-y-1">
              <label class="text-xs font-medium text-slate-600">Catatan Tambahan (Opsional)</label>
              <textarea v-model="reportForm.notes" rows="2" placeholder="Catatan shift kasir..." class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"></textarea>
            </div>
          </div>
        </div>

        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button @click="showReportModal = false" :disabled="isUploadingReport" class="px-4 py-2 font-bold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">
            Batal
          </button>
          <button @click="saveReport" :disabled="isUploadingReport" class="px-5 py-2 bg-primary text-white font-bold rounded-xl hover:bg-primary/90 transition-colors flex items-center gap-2 shadow-sm text-sm disabled:opacity-50">
            <span v-if="isUploadingReport" class="animate-spin w-4 h-4 border-2 border-white/20 border-t-white rounded-full"></span>
            <Save v-else class="w-4 h-4" /> 
            {{ isUploadingReport ? 'Menyimpan...' : (isEditingReport ? 'Perbarui Laporan' : 'Simpan Laporan') }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: CAIRKAN REIMBURSEMENT (FLEKSIBEL: KAS KECIL, KAS BESAR, TRANSFER BANK PT) -->
    <div v-if="showDisburseModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showDisburseModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between bg-emerald-50/50">
          <div>
            <h3 class="text-lg font-bold text-slate-900">Pencairan Dana Reimbursement</h3>
            <p class="text-xs text-slate-500">Pilih sumber dana dari kas/bank PT salesman bekerja.</p>
          </div>
          <button @click="showDisburseModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">✕</button>
        </div>

        <div class="p-6 space-y-4 max-h-[80vh] overflow-y-auto">
          <!-- Detail Klaim -->
          <div class="bg-slate-50 p-4 rounded-2xl border border-slate-200 space-y-2">
            <div class="flex justify-between items-start">
              <div>
                <p class="text-xs text-slate-500">Karyawan / Salesman</p>
                <p class="text-sm font-bold text-slate-900">{{ selectedReimbursement?.employee?.first_name }} {{ selectedReimbursement?.employee?.last_name || '' }}</p>
                <span v-if="selectedReimbursement?.employee?.company" class="inline-block px-2 py-0.5 bg-blue-100 text-blue-800 text-[10px] font-bold rounded mt-1">
                  PT: {{ selectedReimbursement.employee.company.name }}
                </span>
              </div>
              <div class="text-right">
                <p class="text-xs text-slate-500">Nominal Klaim</p>
                <p class="text-lg font-black text-emerald-700">{{ formatRupiah(selectedReimbursement?.amount) }}</p>
              </div>
            </div>
            <p class="text-xs text-slate-600 font-medium pt-2 border-t border-slate-200">
              Judul: <strong>{{ selectedReimbursement?.title }}</strong>
            </p>
            <div v-if="selectedReimbursement?.start_km != null && selectedReimbursement?.end_km != null" class="flex items-center gap-2 text-xs text-amber-900 bg-amber-50 p-2 rounded-xl border border-amber-200">
              <span>⛽ Perjalanan: <strong>{{ Number(selectedReimbursement.start_km).toLocaleString('id-ID') }} → {{ Number(selectedReimbursement.end_km).toLocaleString('id-ID') }} ({{ selectedReimbursement.total_km || (selectedReimbursement.end_km - selectedReimbursement.start_km) }} KM)</strong></span>
              <span v-if="selectedReimbursement.vehicle_plate_number" class="text-slate-600 font-bold">• Plat: {{ selectedReimbursement.vehicle_plate_number }}</span>
            </div>
          </div>

          <!-- Pilih PT Pembebanan Dana -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">Perusahaan (PT) Sumber Dana</label>
            <select v-model="disburseForm.company_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-400 bg-white font-semibold text-sm">
              <option v-for="c in branchCompanies" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>

          <!-- Pilihan Metode Pencairan -->
          <div class="space-y-2">
            <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">Sumber Dana Pencairan</label>
            <div class="grid grid-cols-3 gap-2">
              <label 
                class="flex flex-col items-center text-center p-2.5 rounded-xl border cursor-pointer transition-all"
                :class="disburseForm.disbursed_via === 'PETTY_CASH' ? 'bg-emerald-50 border-emerald-500 text-emerald-800 font-bold shadow-sm' : 'bg-white border-slate-200 text-slate-600'"
              >
                <input type="radio" value="PETTY_CASH" v-model="disburseForm.disbursed_via" class="hidden" />
                <Wallet class="w-4 h-4 mb-1 text-emerald-600" />
                <span class="text-[11px]">Kas Kecil PT</span>
              </label>

              <label 
                class="flex flex-col items-center text-center p-2.5 rounded-xl border cursor-pointer transition-all"
                :class="disburseForm.disbursed_via === 'CASH' ? 'bg-amber-50 border-amber-500 text-amber-800 font-bold shadow-sm' : 'bg-white border-slate-200 text-slate-600'"
              >
                <input type="radio" value="CASH" v-model="disburseForm.disbursed_via" class="hidden" />
                <Coins class="w-4 h-4 mb-1 text-amber-600" />
                <span class="text-[11px]">Kas Besar PT</span>
              </label>

              <label 
                class="flex flex-col items-center text-center p-2.5 rounded-xl border cursor-pointer transition-all"
                :class="disburseForm.disbursed_via === 'BANK_TRANSFER' ? 'bg-blue-50 border-blue-500 text-blue-800 font-bold shadow-sm' : 'bg-white border-slate-200 text-slate-600'"
              >
                <input type="radio" value="BANK_TRANSFER" v-model="disburseForm.disbursed_via" class="hidden" />
                <Building class="w-4 h-4 mb-1 text-blue-600" />
                <span class="text-[11px]">Transfer Bank PT</span>
              </label>
            </div>
          </div>

          <!-- Jika Transfer Bank: Pilih Rekening Bank Milik PT Tersebut -->
          <div v-if="disburseForm.disbursed_via === 'BANK_TRANSFER'" class="space-y-1.5 bg-blue-50/50 p-3.5 rounded-2xl border border-blue-200">
            <label class="text-xs font-bold text-slate-800">Pilih Rekening Bank Sumber</label>
            <select v-model="disburseForm.bank_account_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold bg-white">
              <option value="">-- Pilih Rekening Bank --</option>
              <option v-for="b in disburseBankAccounts" :key="b.id" :value="b.id">
                {{ b.company?.name ? `[${b.company.name}] ` : '' }}{{ b.bank_name }} - {{ b.account_number }} ({{ b.account_holder }}) - Saldo: {{ formatRupiah(b.manual_balance) }}
              </option>
            </select>
          </div>

          <!-- Real-time Balance Safeguard -->
          <div class="p-3 rounded-xl text-xs font-medium" :class="selectedFundBalance >= (selectedReimbursement?.amount || 0) ? 'bg-emerald-50 text-emerald-800 border border-emerald-200' : 'bg-red-50 text-red-800 border border-red-200'">
            <div class="flex justify-between items-center">
              <span>Saldo Sumber Terpilih: <strong>{{ formatRupiah(selectedFundBalance) }}</strong></span>
              <span v-if="selectedFundBalance >= (selectedReimbursement?.amount || 0)" class="text-emerald-700 font-bold">Saldo Cukup</span>
              <span v-else class="text-red-700 font-bold">Saldo Kurang!</span>
            </div>
          </div>

          <!-- Upload Bukti Serah Terima -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Bukti Tanda Terima / Kwitansi (Opsional)</label>
            <label class="flex flex-col items-center justify-center w-full h-24 border-2 border-slate-200 border-dashed rounded-2xl cursor-pointer hover:bg-slate-50 transition-colors bg-slate-50/50">
              <div class="flex flex-col items-center justify-center p-2 text-center">
                <Upload v-if="!selectedDisburseFile" class="w-5 h-5 mb-1 text-slate-400" />
                <File v-else class="w-5 h-5 mb-1 text-emerald-600" />
                <p class="text-xs text-slate-600 font-bold">{{ selectedDisburseFile ? selectedDisburseFile.name : 'Foto nota/tanda tangan penerimaan' }}</p>
              </div>
              <input type="file" class="hidden" accept="image/*,application/pdf" @change="handleReportFileChange" />
            </label>
          </div>

          <div class="space-y-1">
            <label class="text-xs font-medium text-slate-600">Catatan Kasir</label>
            <textarea v-model="disburseForm.notes" rows="2" placeholder="Diserahkan langsung tunai..." class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none"></textarea>
          </div>
        </div>

        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button @click="showDisburseModal = false" :disabled="isDisbursing" class="px-4 py-2 font-bold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">
            Batal
          </button>
          <button 
            @click="saveDisburse" 
            :disabled="isDisbursing" 
            class="px-5 py-2 bg-emerald-600 text-white font-bold rounded-xl hover:bg-emerald-700 transition-colors flex items-center gap-2 shadow-sm text-sm disabled:opacity-50"
          >
            <span v-if="isDisbursing" class="animate-spin w-4 h-4 border-2 border-white/20 border-t-white rounded-full"></span>
            <CheckCircle2 v-else class="w-4 h-4" />
            Konfirmasi Pencairan
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: CATAT / EDIT MUTASI (KAS KECIL, KAS BESAR, ATAU BANK PT) -->
    <div v-if="showMutationModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-3 sm:p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showMutationModal = false"></div>
      <div
        class="relative bg-white rounded-3xl shadow-xl w-full overflow-hidden animate-in fade-in zoom-in-95 duration-200 transition-all"
        :class="isFactoryInvoiceDeposit ? 'max-w-5xl' : 'max-w-xl'"
      >
        <!-- Header Modal -->
        <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/70">
          <div>
            <h3 class="text-lg font-bold text-slate-900 flex items-center gap-2">
              <Receipt v-if="isFactoryDeposit" class="w-5 h-5 text-indigo-600" />
              <Wallet v-else class="w-5 h-5 text-primary" />
              {{
                editingMutationId
                  ? 'Edit Mutasi Dana'
                  : isFactoryInvoiceDeposit
                    ? 'Setoran Faktur Penjualan ke Pabrik'
                    : mutationForm.category === 'FACTORY_NON_INVOICE_DEPOSIT'
                      ? 'Setoran Dana Nonfaktur ke Pabrik'
                      : 'Catat Mutasi Kas / Bank'
              }}
            </h3>
            <p class="text-xs text-slate-500 mt-0.5">
              {{
                isFactoryInvoiceDeposit
                  ? 'Pilih faktur per hari atau per transaksi, tentukan rekening cabang tujuan, lalu unggah bukti transfer.'
                  : isFactoryDeposit
                    ? 'Catat setoran dana nonfaktur ke rekening bank cabang untuk diverifikasi Admin Kasir cabang.'
                    : 'Catat transaksi dana masuk atau keluar pada Kas Kecil, Kas Besar, atau Rekening Bank cabang.'
              }}
            </p>
          </div>
          <button @click="showMutationModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-200/60 transition-colors">✕</button>
        </div>

        <!-- Body Modal -->
        <div class="p-5 sm:p-6 max-h-[82vh] overflow-y-auto space-y-5">
          <!-- Baris Konteks Utama: Kategori Mutasi & Tanggal Mutasi -->
          <div class="grid grid-cols-1 sm:grid-cols-12 gap-3 bg-slate-50 p-3.5 rounded-2xl border border-slate-200/80">
            <div class="sm:col-span-8 space-y-1">
              <label class="text-[11px] font-bold text-slate-700 uppercase tracking-wider">Kategori Mutasi</label>
              <select v-model="mutationForm.category" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 text-sm font-bold bg-white text-slate-900">
                <option value="OPERATIONAL_EXPENSE">Operasional Cabang (Listrik/Air/Bensin/ATK/dll)</option>
                <option value="PETTY_CASH_TOPUP">Top-up / Pemasukan Kas Kecil</option>
                <option value="INITIAL_BALANCE">Saldo Awal / Koreksi Saldo</option>
                <option value="SALES_DEPOSIT">Setoran Kas Penjualan</option>
                <option value="FACTORY_INVOICE_DEPOSIT">Setoran Faktur ke Pabrik</option>
                <option value="FACTORY_NON_INVOICE_DEPOSIT">Setoran Dana Nonfaktur ke Pabrik</option>
                <option value="OTHER">Lainnya</option>
              </select>
            </div>
            <div class="sm:col-span-4 space-y-1">
              <label class="text-[11px] font-bold text-slate-700 uppercase tracking-wider">Tanggal Mutasi</label>
              <input v-model="mutationForm.mutation_date" type="date" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 text-sm font-semibold bg-white text-slate-900" />
            </div>
          </div>

          <!-- ========================================================================= -->
          <!-- MODE 1: SETORAN FAKTUR KE PABRIK (LAYOUT 2 KOLOM + PILIH FAKTUR PER HARI) -->
          <!-- ========================================================================= -->
          <div v-if="isFactoryInvoiceDeposit" class="grid grid-cols-1 lg:grid-cols-12 gap-5 items-start">
            <!-- KOLOM KIRI (7/12): Pemilihan Faktur Berdasarkan Hari -->
            <div class="lg:col-span-7 space-y-3 rounded-2xl border border-indigo-200/80 bg-indigo-50/30 p-4">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <h4 class="text-sm font-extrabold text-indigo-950 flex items-center gap-1.5">
                    <Calendar class="w-4 h-4 text-indigo-600" />
                    1. Pilih Penerimaan Faktur per Hari
                  </h4>
                  <p class="text-[11px] text-indigo-800/80">
                    Centang per tanggal untuk memilih seluruh faktur di hari tersebut sekaligus.
                  </p>
                </div>
                <!-- Toggle Acuan Hari -->
                <div class="inline-flex rounded-xl bg-white p-0.5 border border-indigo-200 shadow-sm text-[11px] font-bold">
                  <button
                    type="button"
                    @click="invoiceDateGroupMode = 'payment_date'"
                    class="px-2.5 py-1 rounded-lg transition-colors"
                    :class="invoiceDateGroupMode === 'payment_date' ? 'bg-indigo-600 text-white' : 'text-slate-600 hover:text-slate-900'"
                    title="Kelompokkan berdasarkan tanggal uang diterima kasir"
                  >
                    Tgl Terima Kasir
                  </button>
                  <button
                    type="button"
                    @click="invoiceDateGroupMode = 'transaction_date'"
                    class="px-2.5 py-1 rounded-lg transition-colors"
                    :class="invoiceDateGroupMode === 'transaction_date' ? 'bg-indigo-600 text-white' : 'text-slate-600 hover:text-slate-900'"
                    title="Kelompokkan berdasarkan tanggal terbit faktur penjualan"
                  >
                    Tgl Faktur
                  </button>
                </div>
              </div>

              <!-- Filter Cabang & PT Sumber -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                <div class="space-y-1">
                  <label class="text-[11px] font-bold text-slate-700 uppercase">Cabang Faktur & Rekening Tujuan</label>
                  <select v-model="mutationForm.branch_id" class="w-full px-3 py-1.5 border border-slate-200 rounded-xl focus:outline-none text-xs font-semibold bg-white">
                    <option value="">Pilih Cabang...</option>
                    <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">{{ b.name }}</option>
                  </select>
                </div>
                <div class="space-y-1">
                  <label class="text-[11px] font-bold text-slate-700 uppercase">PT Sumber (Sesuai Faktur)</label>
                  <select v-model="mutationForm.company_id" class="w-full px-3 py-1.5 border border-slate-200 rounded-xl focus:outline-none text-xs font-semibold bg-white">
                    <option value="">Semua PT (Otomatis mengikuti pilihan)</option>
                    <option v-for="c in branchCompanies" :key="c.id" :value="c.id">{{ c.name }}</option>
                  </select>
                </div>
              </div>
              <p v-if="['cashier', 'admin_kasir'].includes(authStore.user?.role || '')" class="text-[11px] text-slate-500">Kasir dapat memilih faktur dari cabang utama atau cabang operasional yang terdaftar pada akun.</p>

              <!-- Pencarian & Tombol Aksi Massal -->
              <div class="flex flex-wrap items-center gap-2">
                <div class="relative flex-1 min-w-[200px]">
                  <input
                    v-model="invoicePaymentSearch"
                    type="search"
                    placeholder="Cari nomor faktur, toko/pelanggan, atau sales..."
                    aria-label="Cari penerimaan faktur"
                    class="w-full rounded-xl border border-indigo-200 bg-white px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-indigo-300"
                  />
                </div>
                <button
                  v-if="filteredInvoicePayments.length > 0"
                  type="button"
                  @click="selectAllVisibleInvoices"
                  class="px-2.5 py-2 rounded-xl bg-indigo-100 hover:bg-indigo-200 text-indigo-900 text-xs font-bold transition-colors whitespace-nowrap"
                >
                  Pilih Semua Tampil ({{ filteredInvoicePayments.length }})
                </button>
                <button
                  v-if="selectedInvoiceCount > 0"
                  type="button"
                  @click="clearSelectedInvoices"
                  class="px-2.5 py-2 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 text-xs font-bold transition-colors whitespace-nowrap"
                >
                  Reset ({{ selectedInvoiceCount }})
                </button>
              </div>

              <!-- Quick Date Chips (Filter & Ringkasan Hari) -->
              <div v-if="invoiceDateChips.length > 0" class="space-y-1.5">
                <div class="flex items-center justify-between text-[11px]">
                  <span class="font-bold text-indigo-900">Filter Cepat per Hari ({{ invoiceDateChips.length }} hari tersedia):</span>
                  <button
                    v-if="activeInvoiceDateFilter"
                    type="button"
                    @click="activeInvoiceDateFilter = ''"
                    class="text-indigo-700 font-bold hover:underline"
                  >
                    Tampilkan Semua Hari
                  </button>
                </div>
                <div class="flex items-center gap-1.5 overflow-x-auto pb-1">
                  <button
                    type="button"
                    @click="activeInvoiceDateFilter = ''"
                    class="shrink-0 px-2.5 py-1.5 rounded-xl text-xs font-bold border transition-all"
                    :class="!activeInvoiceDateFilter ? 'bg-slate-900 text-white border-slate-900 shadow-sm' : 'bg-white text-slate-700 border-slate-200 hover:bg-slate-50'"
                  >
                    Semua ({{ companyAndSearchFilteredInvoices.length }})
                  </button>
                  <button
                    v-for="chip in invoiceDateChips"
                    :key="chip.dateKey"
                    type="button"
                    @click="activeInvoiceDateFilter = activeInvoiceDateFilter === chip.dateKey ? '' : chip.dateKey"
                    class="shrink-0 px-2.5 py-1.5 rounded-xl text-left border transition-all flex items-center gap-2"
                    :class="[
                      activeInvoiceDateFilter === chip.dateKey
                        ? 'bg-indigo-600 text-white border-indigo-600 shadow-sm'
                        : chip.isAllSelected
                          ? 'bg-emerald-50 text-emerald-900 border-emerald-300'
                          : chip.isPartiallySelected
                            ? 'bg-indigo-50 text-indigo-900 border-indigo-300'
                            : 'bg-white text-slate-700 border-slate-200 hover:bg-slate-50'
                    ]"
                  >
                    <div>
                      <div class="text-[11px] font-extrabold leading-tight flex items-center gap-1">
                        <span>{{ chip.shortLabel }}</span>
                        <span
                          class="px-1.5 py-0.2 rounded-full text-[10px]"
                          :class="activeInvoiceDateFilter === chip.dateKey ? 'bg-white/20 text-white' : 'bg-slate-100 text-slate-700'"
                        >
                          {{ chip.selectedCount > 0 ? `${chip.selectedCount}/${chip.count}` : chip.count }}
                        </span>
                      </div>
                      <div
                        class="text-[10px] font-semibold"
                        :class="activeInvoiceDateFilter === chip.dateKey ? 'text-indigo-100' : 'text-emerald-700'"
                      >
                        {{ formatRupiah(chip.totalAvailable) }}
                      </div>
                    </div>
                  </button>
                </div>
              </div>

              <!-- Loading State -->
              <div v-if="isLoadingInvoicePayments" class="rounded-2xl bg-white p-8 text-center border border-indigo-100">
                <RefreshCw class="w-6 h-6 mx-auto mb-2 animate-spin text-indigo-600" />
                <p class="text-xs font-semibold text-slate-600">Memuat daftar penerimaan faktur...</p>
              </div>

              <!-- Daftar Faktur Terkelompok per Hari -->
              <div v-else-if="groupedInvoicePaymentsByDate.length > 0" class="max-h-[430px] overflow-y-auto space-y-2.5 pr-1">
                <div
                  v-for="group in groupedInvoicePaymentsByDate"
                  :key="group.dateKey"
                  class="rounded-2xl border overflow-hidden transition-colors bg-white"
                  :class="group.isAllSelected ? 'border-emerald-300 ring-1 ring-emerald-200' : group.isPartiallySelected ? 'border-indigo-300' : 'border-slate-200'"
                >
                  <!-- Header Grup Hari -->
                  <div
                    class="px-3.5 py-2.5 flex items-center justify-between gap-2 border-b transition-colors"
                    :class="group.isAllSelected ? 'bg-emerald-50/80 border-emerald-200' : group.isPartiallySelected ? 'bg-indigo-50/70 border-indigo-100' : 'bg-slate-50 border-slate-100'"
                  >
                    <label class="flex items-center gap-2.5 cursor-pointer select-none min-w-0 flex-1">
                      <input
                        type="checkbox"
                        class="w-4 h-4 rounded text-indigo-600 focus:ring-indigo-500 cursor-pointer"
                        :checked="group.isAllSelected"
                        :indeterminate.prop="group.isPartiallySelected"
                        @change="toggleDateGroupSelection(group)"
                      />
                      <div class="min-w-0">
                        <div class="flex items-center gap-2 flex-wrap">
                          <span class="text-xs font-extrabold text-slate-900">{{ group.fullLabel }}</span>
                          <span
                            class="px-2 py-0.5 rounded-full text-[10px] font-bold"
                            :class="group.isAllSelected ? 'bg-emerald-200 text-emerald-900' : group.selectedCount > 0 ? 'bg-indigo-200 text-indigo-900' : 'bg-slate-200 text-slate-700'"
                          >
                            {{ group.selectedCount }}/{{ group.count }} faktur
                          </span>
                        </div>
                        <p class="text-[11px] text-slate-600 mt-0.5">
                          <template v-if="group.selectedCount > 0">
                            Terpilih: <strong class="text-emerald-700">{{ formatRupiah(group.selectedAmount) }}</strong> dari {{ formatRupiah(group.totalAvailable) }}
                          </template>
                          <template v-else>
                            Total hari ini: <strong class="text-slate-800">{{ formatRupiah(group.totalAvailable) }}</strong>
                          </template>
                        </p>
                      </div>
                    </label>

                    <div class="flex items-center gap-1.5 shrink-0">
                      <button
                        type="button"
                        @click="toggleDateGroupSelection(group)"
                        class="px-2.5 py-1 rounded-lg text-[11px] font-bold transition-colors"
                        :class="group.isAllSelected ? 'bg-white text-slate-700 border border-slate-200 hover:bg-slate-100' : 'bg-indigo-600 text-white hover:bg-indigo-700'"
                      >
                        {{ group.isAllSelected ? 'Batal Hari Ini' : 'Pilih Hari Ini' }}
                      </button>
                      <button
                        type="button"
                        @click="toggleCollapseInvoiceDate(group.dateKey)"
                        class="p-1.5 rounded-lg text-slate-500 hover:bg-slate-200/60 transition-transform"
                        :title="collapsedInvoiceDates[group.dateKey] ? 'Buka rincian faktur' : 'Tutup rincian faktur'"
                      >
                        <ChevronRight class="w-4 h-4 transition-transform duration-200" :class="!collapsedInvoiceDates[group.dateKey] ? 'rotate-90' : ''" />
                      </button>
                    </div>
                  </div>

                  <!-- Daftar Item Faktur di Hari Ini -->
                  <div v-show="!collapsedInvoiceDates[group.dateKey]" class="divide-y divide-slate-100">
                    <div
                      v-for="payment in group.items"
                      :key="payment.payment_id"
                      class="p-3 flex items-start gap-2.5 transition-colors hover:bg-slate-50/80"
                      :class="selectedInvoicePayments[payment.payment_id] ? 'bg-indigo-50/25' : ''"
                    >
                      <input
                        type="checkbox"
                        class="mt-1 w-4 h-4 rounded text-indigo-600 focus:ring-indigo-500 cursor-pointer"
                        :checked="!!selectedInvoicePayments[payment.payment_id]"
                        @change="toggleInvoicePayment(payment)"
                      />
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center justify-between gap-2">
                          <div class="flex items-center gap-1.5 flex-wrap">
                            <strong class="text-xs font-extrabold text-slate-900 font-mono">{{ payment.receipt_no }}</strong>
                            <span class="px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-bold">
                              {{ getCompanyNameById(payment.company_id) }}
                            </span>
                          </div>
                          <span class="text-xs font-extrabold text-emerald-700 shrink-0">
                            {{ formatRupiah(payment.available_amount) }}
                          </span>
                        </div>

                        <p class="truncate text-[11px] font-medium text-slate-700 mt-0.5">
                          {{ payment.buyer_name || 'Pelanggan' }} · <span class="text-slate-500">Sales: {{ payment.salesman_name || '-' }}</span>
                        </p>

                        <div class="flex flex-wrap items-center gap-x-3 gap-y-0.5 mt-1 text-[10px] text-slate-500">
                          <span>Terima Kasir: <strong class="text-slate-700">{{ formatShortDate(payment.payment_date) }}</strong></span>
                          <span v-if="payment.transaction_date">Tgl Faktur: <strong class="text-slate-700">{{ formatShortDate(payment.transaction_date) }}</strong></span>
                        </div>

                        <!-- Input Nominal Setoran per Faktur (Bisa Parsial) -->
                        <div v-if="selectedInvoicePayments[payment.payment_id]" class="mt-2 flex items-center gap-2 bg-white p-2 rounded-xl border border-indigo-200">
                          <label class="text-[11px] font-bold text-slate-600 whitespace-nowrap">Disetor (Rp):</label>
                          <input
                            v-model.number="selectedInvoicePayments[payment.payment_id]"
                            type="number"
                            min="0.01"
                            :max="payment.available_amount"
                            step="0.01"
                            class="w-full rounded-lg border border-slate-200 px-2.5 py-1 text-xs font-bold text-slate-900 focus:outline-none focus:ring-2 focus:ring-indigo-300"
                            aria-label="Nominal penerimaan faktur yang disetor"
                          />
                          <button
                            v-if="Number(selectedInvoicePayments[payment.payment_id]) !== Number(payment.available_amount)"
                            type="button"
                            @click="selectedInvoicePayments[payment.payment_id] = Number(payment.available_amount)"
                            class="px-2 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 text-[10px] font-bold text-slate-700 whitespace-nowrap"
                          >
                            Penuh
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Empty State -->
              <div v-else class="rounded-2xl bg-white p-6 text-center border border-indigo-100">
                <p class="text-xs font-semibold text-slate-600">
                  {{
                    availableInvoicePaymentsError ||
                    (invoicePaymentSearch || activeInvoiceDateFilter
                      ? 'Tidak ada penerimaan faktur yang cocok dengan filter tanggal/pencarian ini.'
                      : 'Belum ada penerimaan tunai faktur yang disetujui kasir dan belum disetor untuk cabang/PT ini.')
                  }}
                </p>
              </div>
            </div>

            <!-- KOLOM KANAN (5/12): Rekening Cabang Tujuan, Ringkasan Alokasi & Bukti -->
            <div class="lg:col-span-5 space-y-4">
              <!-- Card 1: Tujuan Setoran Cabang -->
              <div class="rounded-2xl border border-slate-200 bg-white p-4 space-y-3 shadow-sm">
                <h4 class="text-sm font-extrabold text-slate-900 flex items-center gap-1.5">
                  <Landmark class="w-4 h-4 text-primary" />
                  2. Tujuan Rekening Cabang
                </h4>

                <div class="space-y-1.5">
                  <label class="text-xs font-bold text-slate-700">Rekening Bank Cabang Tujuan <span class="text-rose-500">*</span></label>
                  <select v-model="mutationForm.bank_account_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold bg-white">
                    <option value="">{{ mutationBankAccounts.length === 0 ? '-- Belum ada rekening bank aktif di cabang ini --' : 'Pilih rekening tujuan...' }}</option>
                    <option v-for="bank in mutationBankAccounts" :key="bank.id" :value="bank.id">
                      {{ bank.bank_name }} - {{ bank.account_number }} ({{ bank.account_holder || '-' }})
                    </option>
                  </select>
                  <p v-if="mutationForm.branch_id && mutationBankAccounts.length === 0" class="text-xs text-amber-700 bg-amber-50 p-2 rounded-lg border border-amber-200">
                    Belum ada rekening bank aktif yang didaftarkan cabang ini di Laporan Kasir Bank.
                  </p>
                </div>
              </div>

              <!-- Card 2: Ringkasan Alokasi & Total Setoran -->
              <div class="rounded-2xl border border-indigo-200 bg-indigo-50/50 p-4 space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold text-indigo-950">Ringkasan Faktur Terpilih</span>
                  <span class="px-2 py-0.5 rounded-full bg-indigo-100 text-indigo-900 text-[11px] font-extrabold">
                    {{ selectedInvoiceCount }} faktur · {{ selectedInvoiceDaysCount }} hari
                  </span>
                </div>

                <!-- Chip Faktur yang Dipilih -->
                <div v-if="selectedInvoiceCount > 0" class="flex flex-wrap gap-1 max-h-24 overflow-y-auto p-1.5 rounded-xl bg-white/80 border border-indigo-100">
                  <button
                    v-for="id in Object.keys(selectedInvoicePayments)"
                    :key="id"
                    type="button"
                    @click="removeSelectedInvoicePayment(id)"
                    class="rounded-lg bg-indigo-100 hover:bg-rose-100 hover:text-rose-900 px-2 py-0.5 text-[11px] font-bold text-indigo-950 transition-colors"
                    :title="`Hapus ${selectedInvoicePaymentReceipts[id]} dari pilihan`"
                  >
                    {{ selectedInvoicePaymentReceipts[id] }} ×
                  </button>
                </div>

                <div class="space-y-1.5 text-xs border-t border-indigo-200/70 pt-2.5">
                  <div class="flex justify-between text-slate-700">
                    <span>Total Alokasi Faktur:</span>
                    <strong class="font-bold text-slate-900">{{ formatRupiah(selectedInvoicePaymentTotal) }}</strong>
                  </div>
                  <div class="space-y-1 pt-1">
                    <label class="text-[11px] font-bold text-indigo-950 block">Tambahan Dana Nonfaktur (Opsional, Rp)</label>
                    <input
                      v-model.number="extraDepositAmount"
                      type="number"
                      min="0"
                      step="0.01"
                      placeholder="0"
                      class="w-full rounded-xl border border-indigo-200 bg-white px-3 py-1.5 text-xs font-bold"
                    />
                    <p class="text-[10px] text-indigo-800">Isi hanya jika nominal bukti transfer melebihi total faktur yang dipilih.</p>
                  </div>
                </div>

                <div v-if="extraDepositAmount > 0" class="space-y-1.5 rounded-xl border border-amber-200 bg-amber-50 p-2.5">
                  <label class="text-xs font-bold text-amber-900">Asal Dana Nonfaktur <span class="text-rose-500">*</span></label>
                  <select v-model="mutationForm.fund_origin" class="w-full rounded-lg border border-amber-200 bg-white px-2.5 py-1.5 text-xs font-semibold">
                    <option value="">Pilih asal dana...</option>
                    <option value="OPENING_BALANCE">Saldo awal kas cabang</option>
                    <option value="CASH_ADJUSTMENT">Koreksi kas cabang</option>
                    <option value="OTHER_NON_SALES">Dana nonpenjualan lainnya</option>
                  </select>
                </div>

                <div class="pt-2 border-t border-indigo-200 flex items-baseline justify-between">
                  <span class="text-xs font-extrabold uppercase tracking-wider text-indigo-950">Total Setoran</span>
                  <span class="text-xl font-black text-indigo-950">{{ formatRupiah(mutationForm.amount) }}</span>
                </div>
              </div>

              <!-- Card 3: Catatan & Bukti Setoran -->
              <div class="space-y-3">
                <div class="space-y-1">
                  <label class="text-xs font-bold text-slate-700">
                    {{ extraDepositAmount > 0 ? 'Alasan & Rincian Asal Dana Nonfaktur (Wajib)' : 'Catatan Setoran (Opsional)' }}
                  </label>
                  <textarea
                    v-model="mutationForm.notes"
                    rows="2"
                    placeholder="Contoh: Setoran tunai faktur penjualan..."
                    class="w-full px-3 py-2 border border-slate-200 rounded-xl text-xs focus:outline-none"
                  ></textarea>
                </div>

                <div class="space-y-1">
                  <label class="text-xs font-bold text-slate-700 flex items-center gap-1">
                    Upload Bukti Setoran / Transfer <span class="text-rose-500">*</span>
                  </label>
                  <label class="flex flex-col items-center justify-center w-full h-24 border-2 border-indigo-200 border-dashed rounded-2xl cursor-pointer hover:bg-indigo-50/40 transition-colors bg-slate-50/50">
                    <div class="flex flex-col items-center justify-center p-2 text-center">
                      <Upload v-if="!selectedFile && !mutationForm.proof_url" class="w-5 h-5 mb-1 text-indigo-500" />
                      <File v-else class="w-5 h-5 mb-1 text-emerald-600" />
                      <p class="text-xs text-slate-700 font-bold">
                        {{ selectedFile ? selectedFile.name : (mutationForm.proof_url ? 'Bukti terlampir (klik untuk ganti)' : 'Klik untuk unggah foto / PDF bukti transfer') }}
                      </p>
                    </div>
                    <input type="file" class="hidden" accept="image/*,application/pdf" @change="handleFileChange" />
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- ========================================================================= -->
          <!-- MODE 2: MUTASI KAS/BANK BIASA ATAU SETORAN DANA NONFAKTUR (1 KOLOM RAPI)   -->
          <!-- ========================================================================= -->
          <div v-else class="space-y-4">
            <!-- Pilihan Cabang & PT -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">{{ isFactoryDeposit ? 'Cabang Setoran & Rekening Tujuan' : 'Cabang' }}</label>
                <select v-model="mutationForm.branch_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold bg-white">
                  <option value="">Pilih Cabang...</option>
                  <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">{{ b.name }}</option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">{{ isFactoryDeposit ? 'PT Sumber Dana' : 'Perusahaan (PT)' }}</label>
                <select v-model="mutationForm.company_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold bg-white">
                  <option value="">Pilih PT...</option>
                  <option v-for="c in branchCompanies" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">Jenis Mutasi</label>
                <select v-model="mutationForm.mutation_type" :disabled="isFactoryDeposit" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold disabled:bg-slate-100">
                  <option value="IN">Dana Masuk (Pemasukan)</option>
                  <option value="OUT">Dana Keluar (Pengeluaran)</option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-700 uppercase tracking-wider">Sumber Dana</label>
                <select v-model="mutationForm.fund_type" :disabled="isFactoryDeposit" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold disabled:bg-slate-100">
                  <option value="PETTY_CASH">Kas Kecil (Petty Cash PT)</option>
                  <option value="CASH">Kas Besar / Tunai PT</option>
                  <option value="BANK">Rekening Bank PT</option>
                </select>
              </div>
            </div>

            <div v-if="mutationForm.fund_type === 'BANK' || isFactoryDeposit" class="space-y-1.5">
              <div class="flex items-center justify-between">
                <label class="text-xs font-bold text-slate-700">{{ isFactoryDeposit ? 'Rekening Bank Cabang Tujuan' : 'Pilih Rekening Bank' }}</label>
                <button
                  v-if="!isFactoryDeposit"
                  type="button"
                  @click="openAddBankModal"
                  class="text-[11px] text-primary font-bold hover:underline flex items-center gap-1"
                  title="Daftarkan Rekening Bank Baru untuk Cabang Ini"
                >
                  <Plus class="w-3 h-3" /> Tambah Rekening
                </button>
              </div>
              <select v-model="mutationForm.bank_account_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold bg-white">
                <option value="">{{ mutationBankAccounts.length === 0 ? '-- Belum ada rekening bank aktif di cabang ini --' : 'Pilih rekening tujuan...' }}</option>
                <option v-for="bank in mutationBankAccounts" :key="bank.id" :value="bank.id">
                  {{ isFactoryDeposit ? '' : (bank.company?.name ? `[${bank.company.name}] ` : '') }}{{ bank.bank_name }} - {{ bank.account_number }} ({{ bank.account_holder || '-' }})
                </option>
              </select>
              <p v-if="mutationBankAccounts.length === 0" class="text-xs text-amber-600 bg-amber-50 p-2 rounded-lg border border-amber-200">
                {{ isFactoryDeposit ? 'Belum ada rekening bank aktif yang didaftarkan cabang ini di Laporan Kasir Bank.' : 'Belum ada rekening bank terdaftar untuk pilihan ini.' }}
              </p>
            </div>

            <div v-if="mutationForm.category === 'FACTORY_NON_INVOICE_DEPOSIT'" class="space-y-1.5 rounded-xl border border-amber-200 bg-amber-50 p-3">
              <label class="text-xs font-bold text-amber-900">Asal dana nonfaktur</label>
              <select v-model="mutationForm.fund_origin" class="w-full rounded-lg border border-amber-200 bg-white px-3 py-2 text-sm">
                <option value="">Pilih asal dana...</option>
                <option value="OPENING_BALANCE">Saldo awal kas cabang</option>
                <option value="CASH_ADJUSTMENT">Koreksi kas cabang</option>
                <option value="OTHER_NON_SALES">Dana nonpenjualan lainnya</option>
              </select>
              <p class="text-[11px] text-amber-800">Dana hasil penjualan tetap harus dialokasikan ke penerimaan faktur yang disetujui kasir.</p>
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">{{ isFactoryDeposit ? 'Total Setoran ke Pabrik (Rp)' : 'Nominal Mutasi (Rp)' }}</label>
              <input v-model.number="mutationForm.amount" type="number" min="0" placeholder="0" class="w-full px-3 py-2.5 border border-slate-200 rounded-xl focus:outline-none text-lg font-black text-slate-900" />
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-medium text-slate-700">{{ mutationForm.category === 'FACTORY_NON_INVOICE_DEPOSIT' ? 'Alasan dan rincian asal dana (wajib)' : 'Catatan / Keterangan' }}</label>
              <textarea v-model="mutationForm.notes" rows="2" :placeholder="isFactoryDeposit ? 'Jelaskan asal dana dan tujuan setoran...' : 'Beli token listrik cabang...'" class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none"></textarea>
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700 flex items-center gap-1">Upload Bukti <span class="text-xs text-slate-400 font-normal">{{ isFactoryDeposit ? '(Wajib)' : '(Opsional)' }}</span></label>
              <label class="flex flex-col items-center justify-center w-full h-24 border-2 border-slate-200 border-dashed rounded-2xl cursor-pointer hover:bg-slate-50 transition-colors bg-slate-50/50">
                <div class="flex flex-col items-center justify-center p-2 text-center">
                  <Upload v-if="!selectedFile && !mutationForm.proof_url" class="w-6 h-6 mb-1 text-slate-400" />
                  <File v-else class="w-6 h-6 mb-1 text-primary" />
                  <p class="text-xs text-slate-600 font-semibold">{{ selectedFile ? selectedFile.name : (mutationForm.proof_url ? 'Bukti terlampir (klik untuk ganti)' : (isFactoryDeposit ? 'Upload bukti setoran (wajib)' : 'Upload struk / kwitansi (opsional)')) }}</p>
                </div>
                <input type="file" class="hidden" accept="image/*,application/pdf" @change="handleFileChange" />
              </label>
            </div>
          </div>
        </div>

        <!-- Footer Modal -->
        <div class="p-4 border-t border-slate-100 bg-slate-50 flex flex-wrap items-center justify-between gap-3">
          <div v-if="isFactoryInvoiceDeposit" class="text-xs text-slate-600">
            Terpilih: <strong class="text-slate-900">{{ selectedInvoiceCount }} faktur</strong> ({{ selectedInvoiceDaysCount }} hari) · Total: <strong class="text-indigo-700">{{ formatRupiah(mutationForm.amount) }}</strong>
          </div>
          <div v-else></div>
          <div class="flex items-center gap-3 ml-auto">
            <button @click="showMutationModal = false" :disabled="isUploading" class="px-4 py-2 font-bold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">
              Batal
            </button>
            <button @click="saveMutation" :disabled="isUploading" class="px-5 py-2 bg-slate-900 text-white font-bold rounded-xl hover:bg-slate-800 transition-colors flex items-center gap-2 shadow-sm text-sm disabled:opacity-50">
              <span v-if="isUploading" class="animate-spin w-4 h-4 border-2 border-white/20 border-t-white rounded-full"></span>
              <Save v-else class="w-4 h-4" />
              {{ isUploading ? 'Menyimpan...' : (editingMutationId ? 'Perbarui Mutasi' : (isFactoryDeposit ? 'Kirim Setoran ke Pabrik' : 'Simpan Mutasi')) }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- MODAL: UPDATE / SET SALDO AWAL KAS KECIL PER PT -->
    <div v-if="showPettyCashBalanceModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showPettyCashBalanceModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-xl w-full max-w-sm overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between bg-emerald-50/50">
          <div>
            <h3 class="text-lg font-bold text-slate-900">Saldo Kas Kecil Per PT</h3>
            <p class="text-xs text-slate-500">Atur saldo awal fisik atau koreksi brankas kas kecil.</p>
          </div>
          <button @click="showPettyCashBalanceModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">✕</button>
        </div>
        <div class="p-6 space-y-4">
          <!-- Pilih PT -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Pilih Perusahaan (PT)</label>
            <select v-model="pettyCashBalanceForm.company_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold bg-white">
              <option v-for="c in branchCompanies" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>

          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Nominal Saldo Kas Kecil (Rp)</label>
            <input v-model.number="pettyCashBalanceForm.petty_cash_balance" type="number" min="0" placeholder="0" class="w-full px-3 py-2.5 border border-emerald-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-400 text-xl font-black text-emerald-900" />
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-slate-700">Keterangan / Alasan Perubahan</label>
            <input v-model="pettyCashBalanceForm.notes" type="text" placeholder="Contoh: Saldo awal kas kecil cabang..." class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none" />
          </div>
        </div>
        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button @click="showPettyCashBalanceModal = false" :disabled="isSavingPettyCash" class="px-4 py-2 font-bold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">Batal</button>
          <button @click="savePettyCashBalance" :disabled="isSavingPettyCash" class="px-5 py-2 bg-emerald-600 text-white font-bold rounded-xl hover:bg-emerald-700 transition-colors flex items-center gap-2 shadow-sm text-sm disabled:opacity-50">
            <span v-if="isSavingPettyCash" class="animate-spin w-4 h-4 border-2 border-white/20 border-t-white rounded-full"></span>
            <Save v-else class="w-4 h-4" /> 
            {{ isSavingPettyCash ? 'Menyimpan...' : 'Simpan Saldo' }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: UPDATE / SET SALDO AWAL KAS BESAR PER PT -->
    <div v-if="showCashBalanceModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showCashBalanceModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-xl w-full max-w-sm overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between bg-amber-50/50">
          <div>
            <h3 class="text-lg font-bold text-slate-900">Saldo Kas Besar Per PT</h3>
            <p class="text-xs text-slate-500">Atur saldo awal fisik kas besar penerimaan tunai.</p>
          </div>
          <button @click="showCashBalanceModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">✕</button>
        </div>
        <div class="p-6 space-y-4">
          <!-- Pilih PT -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Pilih Perusahaan (PT)</label>
            <select v-model="cashBalanceForm.company_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none text-sm font-semibold bg-white">
              <option v-for="c in branchCompanies" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>

          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Nominal Saldo Kas Besar (Rp)</label>
            <input v-model.number="cashBalanceForm.manual_cash_balance" type="number" min="0" placeholder="0" class="w-full px-3 py-2.5 border border-amber-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-amber-400 text-xl font-black text-amber-900" />
          </div>
          <div class="space-y-1.5">
            <label class="text-xs font-medium text-slate-700">Keterangan / Alasan Perubahan</label>
            <input v-model="cashBalanceForm.notes" type="text" placeholder="Contoh: Saldo awal kas besar cabang..." class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none" />
          </div>
        </div>
        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button @click="showCashBalanceModal = false" :disabled="isSavingCashBalance" class="px-4 py-2 font-bold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">Batal</button>
          <button @click="saveCashBalance" :disabled="isSavingCashBalance" class="px-5 py-2 bg-amber-600 text-white font-bold rounded-xl hover:bg-amber-700 transition-colors flex items-center gap-2 shadow-sm text-sm disabled:opacity-50">
            <span v-if="isSavingCashBalance" class="animate-spin w-4 h-4 border-2 border-white/20 border-t-white rounded-full"></span>
            <Save v-else class="w-4 h-4" /> 
            {{ isSavingCashBalance ? 'Menyimpan...' : 'Simpan Saldo' }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: UPDATE SALDO MANUAL BANK -->
    <div v-if="showBalanceModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showBalanceModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-xl w-full max-w-sm overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-lg font-bold text-slate-900">Update Saldo Bank</h3>
          <button @click="showBalanceModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">✕</button>
        </div>
        <div class="p-6 space-y-3">
          <p class="text-xs text-slate-500">Saldo terbaru untuk rekening <strong>{{ selectedBank?.bank_name }}</strong> ({{ selectedBank?.account_number }}).</p>
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Nominal Saldo (Rp)</label>
            <input v-model.number="newBalance" type="number" min="0" class="w-full px-3 py-2.5 border border-slate-200 rounded-xl focus:outline-none text-lg font-black text-slate-900" />
          </div>
        </div>
        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button @click="showBalanceModal = false" class="px-4 py-2 font-bold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">Batal</button>
          <button @click="saveBalance" class="px-5 py-2 bg-primary text-white font-bold rounded-xl hover:bg-primary/90 transition-colors flex items-center gap-2 shadow-sm text-sm">
            <Save class="w-4 h-4" /> Simpan
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: PREVIEW BUKTI INTERAKTIF (ZOOM IN/OUT & ROTATE) -->
    <div v-if="showPreviewModal" class="fixed inset-0 z-[100] flex items-center justify-center p-3 sm:p-6 bg-slate-950/80 backdrop-blur-md transition-all duration-300">
      <!-- Backdrop click to close -->
      <div class="fixed inset-0" @click="closePreviewModal"></div>

      <div class="relative bg-slate-900 border border-slate-800 rounded-3xl shadow-2xl w-full max-w-4xl h-[88vh] flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-200 z-10">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-950/60 shrink-0">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-primary/20 text-primary rounded-xl">
              <Eye class="w-5 h-5" />
            </div>
            <div>
              <h3 class="font-bold text-white text-base leading-tight">{{ previewModalTitle }}</h3>
              <p v-if="previewModalSubtitle" class="text-xs text-slate-400 mt-0.5">{{ previewModalSubtitle }}</p>
            </div>
          </div>

          <!-- Quick Actions & Close -->
          <div class="flex items-center gap-2">
            <a 
              :href="previewImageUrl" 
              target="_blank" 
              class="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white rounded-xl text-xs font-bold transition-colors flex items-center gap-1.5"
              title="Buka gambar di tab baru"
            >
              <ExternalLink class="w-3.5 h-3.5" /> Tab Baru
            </a>
            <button 
              @click="closePreviewModal" 
              class="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors"
              title="Tutup Preview"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>

        <!-- Viewport Container -->
        <div class="flex-1 relative overflow-hidden bg-slate-950/50 flex items-center justify-center p-4 select-none">
          <!-- Scaled & Rotated Container -->
          <div 
            class="transition-transform duration-200 ease-out flex items-center justify-center max-w-full max-h-full"
            :style="{
              transform: `scale(${zoomScale}) rotate(${rotationDegrees}deg)`
            }"
          >
            <!-- Check if PDF -->
            <div v-if="previewImageUrl.toLowerCase().endsWith('.pdf')" class="text-center p-8 bg-slate-900 rounded-2xl border border-slate-800 text-slate-200 space-y-3">
              <FileText class="w-16 h-16 mx-auto text-primary" />
              <p class="font-bold text-sm">Dokumen Format PDF</p>
              <a :href="previewImageUrl" target="_blank" class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-white text-xs font-bold rounded-xl hover:bg-primary/90 transition-colors">
                <ExternalLink class="w-4 h-4" /> Buka / Unduh PDF
              </a>
            </div>
            <!-- Image -->
            <img 
              v-else 
              :src="previewImageUrl" 
              alt="Bukti Preview" 
              class="max-w-full max-h-[64vh] object-contain rounded-xl shadow-2xl border border-slate-800/80 pointer-events-none"
            />
          </div>
        </div>

        <!-- Bottom Controls Toolbar -->
        <div class="px-6 py-3.5 bg-slate-950 border-t border-slate-800 flex items-center justify-between shrink-0">
          <div class="flex items-center gap-2">
            <!-- Zoom Out -->
            <button 
              @click="zoomOut" 
              :disabled="zoomScale <= 0.25"
              class="p-2 bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white rounded-xl transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              title="Perkecil (Zoom Out)"
            >
              <ZoomOut class="w-4 h-4" />
            </button>

            <!-- Zoom Indicator -->
            <span class="px-3 py-1 bg-slate-900 border border-slate-800 text-slate-300 font-mono text-xs font-bold rounded-lg min-w-[56px] text-center">
              {{ Math.round(zoomScale * 100) }}%
            </span>

            <!-- Zoom In -->
            <button 
              @click="zoomIn" 
              :disabled="zoomScale >= 4.0"
              class="p-2 bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white rounded-xl transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              title="Perbesar (Zoom In)"
            >
              <ZoomIn class="w-4 h-4" />
            </button>
          </div>

          <!-- Rotation & Reset Controls -->
          <div class="flex items-center gap-2">
            <button 
              @click="rotateImage('ccw')" 
              class="p-2 bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white rounded-xl transition-colors flex items-center gap-1.5 text-xs font-bold"
              title="Putar Berlawanan Jarum Jam (-90°)"
            >
              <RotateCcw class="w-4 h-4" />
            </button>
            <button 
              @click="rotateImage('cw')" 
              class="p-2 bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white rounded-xl transition-colors flex items-center gap-1.5 text-xs font-bold"
              title="Putar Searah Jarum Jam (90°)"
            >
              <RotateCw class="w-4 h-4" />
            </button>
            <button 
              @click="resetZoom" 
              class="px-3 py-2 bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white rounded-xl transition-colors flex items-center gap-1.5 text-xs font-bold"
              title="Reset Tampilan"
            >
              <RefreshCw class="w-3.5 h-3.5" /> Reset
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
