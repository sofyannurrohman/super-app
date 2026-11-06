<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { onClickOutside } from '@vueuse/core'
import {
  Landmark,
  Plus,
  RefreshCw,
  ArrowDownToLine,
  ArrowUpFromLine,
  ArrowRightLeft,
  Search,
  Trash2,
  AlertCircle,
  X,
  Calendar,
  Wallet,
  Pencil,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  Check,
  Truck,
  User,
  Building,
  CheckCircle2,
  CreditCard,
  FileText,
  Upload,
  Coins,
  Calculator,
  Copy,
  Printer,
  Layers,
  Tag,
  FolderTree,
  Edit,
  SlidersHorizontal,
  ChevronUp
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'

interface Branch {
  id: string
  name: string
}

interface Vehicle {
  id: string
  name: string
  license_plate: string
  vehicle_type?: string
  branch?: { id: string, name: string }
}

interface Employee {
  id: string
  first_name: string
  last_name?: string
  job_position?: { title: string }
}

interface KasIncomeSource {
  id: string
  code: string
  name: string
  account_code?: string
  account_name?: string
  description?: string
  is_active: boolean
  sort_order: number
}

interface KasExpenseType {
  id: string
  code: string
  name: string
  description?: string
  is_active: boolean
  sort_order: number
  accounts?: KasExpenseAccount[]
}

interface KasExpenseAccount {
  id: string
  expense_type_id: string
  account_code: string
  account_name: string
  accurate_account_id?: string
  description?: string
  is_active: boolean
  sort_order: number
  expense_type?: KasExpenseType
}

interface KasPabrik {
  id: string
  branch_id: string
  kas_code: string
  kas_name: string
  description: string
  bank_name?: string
  account_number?: string
  balance: number
  is_active: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

interface KasTransaction {
  id: string
  kas_pabrik_id: string
  transaction_type: 'IN' | 'OUT' | 'TRANSFER_IN' | 'TRANSFER_OUT'
  amount: number
  balance_before: number
  balance_after: number
  income_source_id?: string
  expense_type_id?: string
  expense_account_id?: string
  income_source?: string
  expense_category?: string
  income_source_obj?: KasIncomeSource
  expense_type_obj?: KasExpenseType
  expense_account_obj?: KasExpenseAccount
  related_kas_pabrik_id?: string
  description?: string
  proof_url?: string
  transaction_date: string
  created_at: string
  kas_pabrik?: KasPabrik
  related_kas_pabrik?: KasPabrik
  is_branch_receivable?: boolean
  target_branch_id?: string
  target_branch?: { id: string, name: string }
  vehicle_id?: string
  vehicle?: { id: string, name: string, license_plate: string }
  employee_id?: string
  employee?: { id: string, first_name: string, last_name?: string }
  advance_status?: string // NONE, OUTSTANDING, SETTLED
  settlement?: {
    id: string
    actual_expense_amount: number
    refund_amount: number
    shortage_amount: number
    notes?: string
    proof_url?: string
    settled_at: string
    settler?: { name: string }
  }
}

// State
const mainTab = ref<'BOOK' | 'INCOME_SOURCE' | 'EXPENSE_TYPE'>('BOOK')
const branches = ref<Branch[]>([])
const selectedBranch = ref('')
const kasList = ref<KasPabrik[]>([])
const isLoadingKas = ref(false)

const currentBranchName = computed(() => {
  const found = branches.value.find(b => b.id === selectedBranch.value)
  return found?.name || 'Pabrik Utama'
})

// Master Data States
const incomeSourceList = ref<KasIncomeSource[]>([])
const isLoadingIncomeSources = ref(false)
const expenseTypeList = ref<KasExpenseType[]>([])
const isLoadingExpenseTypes = ref(false)
const expenseAccountList = ref<KasExpenseAccount[]>([])
const isLoadingExpenseAccounts = ref(false)

// Modal Master States
const showIncomeSourceModal = ref(false)
const isEditingIncomeSource = ref(false)
const incomeSourceForm = ref({
  id: '',
  code: '',
  name: '',
  account_code: '',
  account_name: '',
  description: '',
  sort_order: 0,
  is_active: true
})

const showExpenseTypeModal = ref(false)
const isEditingExpenseType = ref(false)
const expenseTypeForm = ref({
  id: '',
  code: '',
  name: '',
  description: '',
  sort_order: 0,
  is_active: true
})

const showExpenseAccountModal = ref(false)
const isEditingExpenseAccount = ref(false)
const expenseAccountForm = ref({
  id: '',
  expense_type_id: '',
  account_code: '',
  account_name: '',
  accurate_account_id: '',
  description: '',
  sort_order: 0,
  is_active: true
})

// Tab aktif KAS: 'all' = semua KAS, atau kas.id = tab khusus kas itu
const activeKasTab = ref<string>('all')

const activeKas = computed<KasPabrik | null>(() =>
  activeKasTab.value === 'all'
    ? null
    : kasList.value.find(k => k.id === activeKasTab.value) ?? null
)

const displayedKasList = computed<KasPabrik[]>(() =>
  activeKasTab.value === 'all'
    ? kasList.value
    : kasList.value.filter(k => k.id === activeKasTab.value)
)

const transactions = ref<KasTransaction[]>([])
const isLoadingTx = ref(false)
const txPage = ref(1)
const txLimit = ref<number>(25) // 10, 25, 50, 100, 0 (Semua data)
const totalTx = ref(0)

const totalPages = computed(() => {
  if (txLimit.value <= 0) return 1
  return Math.ceil(totalTx.value / txLimit.value) || 1
})

const startEntry = computed(() => {
  if (totalTx.value === 0) return 0
  if (txLimit.value <= 0) return 1
  return (txPage.value - 1) * txLimit.value + 1
})

const endEntry = computed(() => {
  if (totalTx.value === 0) return 0
  if (txLimit.value <= 0) return totalTx.value
  return Math.min(txPage.value * txLimit.value, totalTx.value)
})

const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  txPage.value = page
  fetchTransactions()
}

const handleLimitChange = () => {
  txPage.value = 1
  fetchTransactions()
}

// Filters
const filterKas = ref('')
const filterType = ref('')
const filterDateFrom = ref('')
const filterDateTo = ref('')
const filterAdvanceStatus = ref('')
const filterTargetBranchId = ref('')
const filterEmployeeId = ref('')
const activeSubTab = ref<'ALL' | 'ADVANCE_OUTSTANDING' | 'BRANCH_RECEIVABLE' | 'VEHICLE'>('ALL')

const vehicles = ref<Vehicle[]>([])
const employees = ref<Employee[]>([])

const fetchVehiclesAndEmployees = async () => {
  try {
    const empBranchParam = selectedBranch.value ? `&branch_id=${selectedBranch.value}&branch_match=primary` : ''
    const [vehRes, empRes] = await Promise.all([
      apiClient.get('/vehicles?bypass_branch=true&bypass_company=true&all=true'),
      apiClient.get(`/employees?limit=1000${empBranchParam}`)
    ])
    vehicles.value = vehRes.data?.data || vehRes.data || []
    employees.value = empRes.data?.data || empRes.data || []
  } catch (error) {
    console.error('Failed to fetch vehicles / employees:', error)
  }
}

const handleSearch = () => {
  txPage.value = 1
  fetchTransactions()
}

// Modals state
const showAddKasModal = ref(false)
const showInModal = ref(false)
const showOutModal = ref(false)
const showTransferModal = ref(false)

// Forms
const kasForm = ref({
  branch_id: '',
  kas_code: '',
  kas_name: '',
  description: '',
  bank_name: '',
  account_number: '',
  sort_order: 0
})

const inForm = ref({
  kas_pabrik_id: '',
  amount: 0,
  income_source_id: '',
  income_source: '',
  target_branch_id: '',
  transaction_date: new Date().toISOString().split('T')[0],
  description: '',
  proof_url: ''
})

const outForm = ref({
  kas_pabrik_id: '',
  amount: 0,
  expense_type_id: '',
  expense_account_id: '',
  expense_category: '',
  expense_type: 'OPERATIONAL' as 'OPERATIONAL' | 'VEHICLE' | 'BRANCH_RECEIVABLE' | 'ADVANCE_LOAN',
  vehicle_id: '',
  target_branch_id: '',
  employee_id: '',
  transaction_date: new Date().toISOString().split('T')[0],
  description: '',
  proof_url: ''
})

// Settlement Modal State
const showSettleModal = ref(false)
const selectedAdvanceTx = ref<KasTransaction | null>(null)
const settleForm = ref({
  actual_expense_amount: 0,
  notes: '',
  proof_url: ''
})
const selectedSettleFile = ref<File | null>(null)
const isSettling = ref(false)

const handleSettleFileChange = (e: any) => {
  if (e.target.files && e.target.files[0]) {
    selectedSettleFile.value = e.target.files[0]
  }
}

const openSettleModal = (tx: KasTransaction) => {
  selectedAdvanceTx.value = tx
  settleForm.value = {
    actual_expense_amount: tx.amount,
    notes: '',
    proof_url: ''
  }
  selectedSettleFile.value = null
  showSettleModal.value = true
}

const handleSettleAdvance = async () => {
  if (!selectedAdvanceTx.value) return
  isSettling.value = true
  try {
    let proofUrl = ''
    if (selectedSettleFile.value) {
      const formData = new FormData()
      formData.append('file', selectedSettleFile.value)
      const uploadRes = await apiClient.post('/cashier-reports/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      proofUrl = uploadRes.data?.data?.url || uploadRes.data?.url
    }

    const actualExpense = Number(settleForm.value.actual_expense_amount) || 0
    await apiClient.post(`/kas-pabrik/transactions/${selectedAdvanceTx.value.id}/settle-advance`, {
      actual_expense_amount: actualExpense,
      notes: settleForm.value.notes,
      proof_url: proofUrl
    })

    toast.success('Penyelesaian kasbon dinas berhasil dicatat!')
    showSettleModal.value = false
    fetchKasList()
    fetchTransactions()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyelesaikan kasbon')
  } finally {
    isSettling.value = false
  }
}

const transferForm = ref({
  from_kas_id: '',
  to_kas_id: '',
  amount: 0,
  transaction_date: new Date().toISOString().split('T')[0],
  description: ''
})

const showEditKasModal = ref(false)
const showEditTxModal = ref(false)

const editKasForm = ref({
  id: '',
  kas_code: '',
  kas_name: '',
  description: '',
  bank_name: '',
  account_number: '',
  sort_order: 0,
  is_active: true
})

const editTxForm = ref({
  id: '',
  transaction_type: '',
  amount: 0,
  income_source_id: '',
  income_source: '',
  expense_type_id: '',
  expense_account_id: '',
  expense_category: '',
  transaction_date: '',
  description: '',
  proof_url: ''
})

// Searchable Expense Account State (Modal OUT)
const expenseAccountSearch = ref('')
const showExpenseAccountDropdown = ref(false)
const expenseAccountDropdownRef = ref<HTMLElement | null>(null)

const availableExpenseAccounts = computed(() => {
  if (outForm.value.expense_type_id) {
    return expenseAccountList.value.filter(a => a.expense_type_id === outForm.value.expense_type_id)
  }
  return expenseAccountList.value
})

const filteredExpenseAccounts = computed(() => {
  const list = availableExpenseAccounts.value
  if (!expenseAccountSearch.value.trim()) return list
  const q = expenseAccountSearch.value.toLowerCase()
  return list.filter(acc => acc.account_name.toLowerCase().includes(q) || acc.account_code.toLowerCase().includes(q))
})

const selectExpenseAccount = (acc: KasExpenseAccount) => {
  outForm.value.expense_account_id = acc.id
  outForm.value.expense_type_id = acc.expense_type_id
  outForm.value.expense_category = acc.account_name
  showExpenseAccountDropdown.value = false
  expenseAccountSearch.value = ''
}

// Penautan Khusus Toggles & Smart Detection
const showVehicleField = ref(false)
const showBranchField = ref(false)
const showAdvanceField = ref(false)

const selectedExpenseType = computed(() => {
  return expenseTypeList.value.find(t => t.id === outForm.value.expense_type_id)
})

const selectedExpenseAccount = computed(() => {
  return expenseAccountList.value.find(a => a.id === outForm.value.expense_account_id)
})

const isVehicleSuggested = computed(() => {
  const text = ((selectedExpenseType.value?.name || '') + ' ' + (selectedExpenseType.value?.code || '') + ' ' + (selectedExpenseAccount.value?.account_name || '') + ' ' + (selectedExpenseAccount.value?.account_code || '')).toUpperCase()
  return text.includes('ARMADA') || text.includes('KENDARAAN') || text.includes('BBM') || text.includes('LOGISTIK') || text.includes('SERVIS')
})

const isBranchSuggested = computed(() => {
  const text = ((selectedExpenseType.value?.name || '') + ' ' + (selectedExpenseType.value?.code || '') + ' ' + (selectedExpenseAccount.value?.account_name || '') + ' ' + (selectedExpenseAccount.value?.account_code || '')).toUpperCase()
  return text.includes('CABANG') || text.includes('TALANGAN') || text.includes('PIUTANG') || text.includes('DROP')
})

const isAdvanceSuggested = computed(() => {
  const text = ((selectedExpenseType.value?.name || '') + ' ' + (selectedExpenseType.value?.code || '') + ' ' + (selectedExpenseAccount.value?.account_name || '') + ' ' + (selectedExpenseAccount.value?.account_code || '')).toUpperCase()
  return text.includes('KASBON') || text.includes('UANG MUKA') || text.includes('DINAS') || text.includes('AUDIT') || text.includes('ADVANCE')
})

const onExpenseTypeChange = (typeId: string) => {
  outForm.value.expense_type_id = typeId
  if (outForm.value.expense_account_id) {
    const acc = expenseAccountList.value.find(a => a.id === outForm.value.expense_account_id)
    if (acc && acc.expense_type_id !== typeId) {
      outForm.value.expense_account_id = ''
      outForm.value.expense_category = ''
    }
  }
}

const onEditExpenseTypeChange = (typeId: string) => {
  editTxForm.value.expense_type_id = typeId
  if (editTxForm.value.expense_account_id) {
    const acc = expenseAccountList.value.find(a => a.id === editTxForm.value.expense_account_id)
    if (acc && acc.expense_type_id !== typeId) {
      editTxForm.value.expense_account_id = ''
      editTxForm.value.expense_category = ''
    }
  }
}

const getSelectedExpenseAccountLabel = computed(() => {
  const acc = expenseAccountList.value.find(c => c.id === outForm.value.expense_account_id)
  return acc ? `${acc.account_name} (${acc.account_code})` : ''
})

onClickOutside(expenseAccountDropdownRef, () => {
  showExpenseAccountDropdown.value = false
})

// Searchable Expense Account State (Modal Edit TX)
const editExpenseAccountSearch = ref('')
const showEditExpenseAccountDropdown = ref(false)
const editExpenseAccountDropdownRef = ref<HTMLElement | null>(null)

const editAvailableExpenseAccounts = computed(() => {
  if (editTxForm.value.expense_type_id) {
    return expenseAccountList.value.filter(a => a.expense_type_id === editTxForm.value.expense_type_id)
  }
  return expenseAccountList.value
})

const filteredEditExpenseAccounts = computed(() => {
  const list = editAvailableExpenseAccounts.value
  if (!editExpenseAccountSearch.value.trim()) return list
  const q = editExpenseAccountSearch.value.toLowerCase()
  return list.filter(acc => acc.account_name.toLowerCase().includes(q) || acc.account_code.toLowerCase().includes(q))
})

const selectEditExpenseAccount = (acc: KasExpenseAccount) => {
  editTxForm.value.expense_account_id = acc.id
  editTxForm.value.expense_type_id = acc.expense_type_id
  editTxForm.value.expense_category = acc.account_name
  showEditExpenseAccountDropdown.value = false
  editExpenseAccountSearch.value = ''
}

const getSelectedEditExpenseAccountLabel = computed(() => {
  const acc = expenseAccountList.value.find(c => c.id === editTxForm.value.expense_account_id)
  if (acc) return `${acc.account_name} (${acc.account_code})`
  return editTxForm.value.expense_category || ''
})

onClickOutside(editExpenseAccountDropdownRef, () => {
  showEditExpenseAccountDropdown.value = false
})

// Formatting
const formatRupiah = (value: number) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(value || 0)


const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

const getCategoryOrSourceLabel = (tx: KasTransaction) => {
  if (tx.transaction_type === 'IN') {
    if (tx.income_source_obj) return tx.income_source_obj.name
    const found = incomeSourceList.value.find(s => s.id === tx.income_source_id || s.code === tx.income_source)
    if (found) return found.name
    return tx.income_source || '-'
  }
  if (tx.transaction_type === 'OUT') {
    if (tx.expense_account_obj) {
      const typeName = tx.expense_type_obj ? `[${tx.expense_type_obj.name}] ` : ''
      return `${typeName}${tx.expense_account_obj.account_name} (${tx.expense_account_obj.account_code})`
    }
    const foundAcc = expenseAccountList.value.find(a => a.id === tx.expense_account_id || a.account_code === tx.expense_category)
    if (foundAcc) {
      const typeObj = expenseTypeList.value.find(t => t.id === foundAcc.expense_type_id)
      const typeName = typeObj ? `[${typeObj.name}] ` : ''
      return `${typeName}${foundAcc.account_name} (${foundAcc.account_code})`
    }
    return tx.expense_category || '-'
  }
  if (['TRANSFER_IN', 'TRANSFER_OUT'].includes(tx.transaction_type)) {
    return 'Transfer Antar KAS'
  }
  return tx.income_source || tx.expense_category || '-'
}

// Otomatisasi Kalkulasi Total Nominal Kas Terfilter
const totalFilteredIn = computed(() => {
  return transactions.value
    .filter(t => ['IN', 'TRANSFER_IN'].includes(t.transaction_type))
    .reduce((sum, t) => sum + (Number(t.amount) || 0), 0)
})

const totalFilteredOut = computed(() => {
  return transactions.value
    .filter(t => ['OUT', 'TRANSFER_OUT'].includes(t.transaction_type))
    .reduce((sum, t) => sum + (Number(t.amount) || 0), 0)
})

const netFilteredMutation = computed(() => {
  return totalFilteredIn.value - totalFilteredOut.value
})

const copyFilterSummary = () => {
  const kasLabel = activeKas.value ? activeKas.value.kas_name : 'Semua KAS Pabrik'
  const filterTypeLabel = filterType.value === 'IN' ? 'Penerimaan Saja' : filterType.value === 'OUT' ? 'Pengeluaran Saja' : 'Semua Mutasi'
  const text = `📊 LAPORAN RINGKASAN KAS
Buku Kas: ${kasLabel}
Filter Jenis: ${filterTypeLabel}
Periode: ${filterDateFrom.value || 'Awal'} s/d ${filterDateTo.value || 'Hari ini'}
Jumlah Transaksi: ${transactions.value.length} data
───────────────────────
Total Masuk (+) : ${formatRupiah(totalFilteredIn.value)}
Total Keluar (-) : ${formatRupiah(totalFilteredOut.value)}
Net Mutasi Kas  : ${formatRupiah(netFilteredMutation.value)}`

  navigator.clipboard.writeText(text)
  toast.success('Ringkasan perhitungan berhasil disalin ke clipboard!')
}

const getKasKey = (name: string) => {
  const upper = name.toUpperCase()
  if (upper.includes('CSR')) return 'csr'
  if (upper.includes('ROSOK')) return 'rosok'
  if (upper.includes('BESAR')) return 'besar'
  if (upper.includes('BCA')) return 'bca'
  if (upper.includes('PETTY')) return 'petty'
  return 'default'
}

const getKasColorClass = (name: string) => {
  const upper = (name || '').toUpperCase()
  if (upper.includes('CSR')) return 'kas-csr'
  if (upper.includes('ROSOK')) return 'kas-rosok'
  if (upper.includes('BESAR')) return 'kas-besar'
  if (upper.includes('BCA_LINDA') || upper.includes('BCA LINDA')) return 'kas-bca-linda'
  if (upper.includes('PETTY_CASH') || upper.includes('PETTY CASH')) return 'kas-petty-cash'
  return 'kas-default'
}

const getKasColorBadgeClass = (name?: string) => {
  const upper = (name || '').toUpperCase()
  if (upper.includes('PETTY')) return 'bg-emerald-100 text-emerald-800 border border-emerald-200'
  if (upper.includes('BESAR')) return 'bg-blue-100 text-blue-800 border border-blue-200'
  if (upper.includes('CSR')) return 'bg-purple-100 text-purple-800 border border-purple-200'
  if (upper.includes('ROSOK')) return 'bg-amber-100 text-amber-800 border border-amber-200'
  if (upper.includes('BCA')) return 'bg-cyan-100 text-cyan-800 border border-cyan-200'
  return 'bg-slate-100 text-slate-800 border border-slate-200'
}

const getTxTypeBadgeClass = (type: string) => {
  switch (type) {
    case 'IN': return 'badge-in'
    case 'OUT': return 'badge-out'
    case 'TRANSFER_IN': return 'badge-tf-in'
    case 'TRANSFER_OUT': return 'badge-tf-out'
    default: return 'badge-default'
  }
}

// API Calls - Master Data
const fetchIncomeSources = async () => {
  isLoadingIncomeSources.value = true
  try {
    const res = await apiClient.get('/kas-pabrik/master/income-sources')
    incomeSourceList.value = res.data?.data || []
  } catch (error) {
    console.error('Failed to fetch income sources:', error)
  } finally {
    isLoadingIncomeSources.value = false
  }
}

const fetchExpenseTypes = async () => {
  isLoadingExpenseTypes.value = true
  try {
    const res = await apiClient.get('/kas-pabrik/master/expense-types')
    expenseTypeList.value = res.data?.data || []
  } catch (error) {
    console.error('Failed to fetch expense types:', error)
  } finally {
    isLoadingExpenseTypes.value = false
  }
}

const fetchExpenseAccounts = async () => {
  isLoadingExpenseAccounts.value = true
  try {
    const res = await apiClient.get('/kas-pabrik/master/expense-accounts')
    expenseAccountList.value = res.data?.data || []
  } catch (error) {
    console.error('Failed to fetch expense accounts:', error)
  } finally {
    isLoadingExpenseAccounts.value = false
  }
}

// CRUD Master Sumber Penerimaan
const openAddIncomeSourceModal = () => {
  isEditingIncomeSource.value = false
  incomeSourceForm.value = {
    id: '',
    code: '',
    name: '',
    account_code: '',
    account_name: '',
    description: '',
    sort_order: (incomeSourceList.value.length + 1),
    is_active: true
  }
  showIncomeSourceModal.value = true
}

const openEditIncomeSourceModal = (src: KasIncomeSource) => {
  isEditingIncomeSource.value = true
  incomeSourceForm.value = {
    id: src.id,
    code: src.code,
    name: src.name,
    account_code: src.account_code || '',
    account_name: src.account_name || '',
    description: src.description || '',
    sort_order: src.sort_order,
    is_active: src.is_active
  }
  showIncomeSourceModal.value = true
}

const handleSaveIncomeSource = async () => {
  try {
    if (isEditingIncomeSource.value) {
      await apiClient.put(`/kas-pabrik/master/income-sources/${incomeSourceForm.value.id}`, incomeSourceForm.value)
      toast.success('Sumber penerimaan berhasil diperbarui')
    } else {
      await apiClient.post('/kas-pabrik/master/income-sources', incomeSourceForm.value)
      toast.success('Sumber penerimaan berhasil ditambahkan')
    }
    showIncomeSourceModal.value = false
    fetchIncomeSources()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyimpan sumber penerimaan')
  }
}

const handleDeleteIncomeSource = async (id: string) => {
  if (!confirm('Apakah anda yakin ingin menghapus sumber penerimaan ini?')) return
  try {
    await apiClient.delete(`/kas-pabrik/master/income-sources/${id}`)
    toast.success('Sumber penerimaan berhasil dihapus')
    fetchIncomeSources()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menghapus sumber penerimaan')
  }
}

// CRUD Master Tipe Pengeluaran
const openAddExpenseTypeModal = () => {
  isEditingExpenseType.value = false
  expenseTypeForm.value = {
    id: '',
    code: '',
    name: '',
    description: '',
    sort_order: (expenseTypeList.value.length + 1),
    is_active: true
  }
  showExpenseTypeModal.value = true
}

const openEditExpenseTypeModal = (et: KasExpenseType) => {
  isEditingExpenseType.value = true
  expenseTypeForm.value = {
    id: et.id,
    code: et.code,
    name: et.name,
    description: et.description || '',
    sort_order: et.sort_order,
    is_active: et.is_active
  }
  showExpenseTypeModal.value = true
}

const isSavingExpenseType = ref(false)
const isSavingExpenseAccount = ref(false)

const handleSaveExpenseType = async () => {
  isSavingExpenseType.value = true
  try {
    if (isEditingExpenseType.value) {
      await apiClient.put(`/kas-pabrik/master/expense-types/${expenseTypeForm.value.id}`, expenseTypeForm.value)
      toast.success('Tipe pengeluaran berhasil diperbarui')
    } else {
      await apiClient.post('/kas-pabrik/master/expense-types', expenseTypeForm.value)
      toast.success('Tipe pengeluaran berhasil ditambahkan')
    }
    showExpenseTypeModal.value = false
    await Promise.all([fetchExpenseTypes(), fetchExpenseAccounts()])
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyimpan tipe pengeluaran')
  } finally {
    isSavingExpenseType.value = false
  }
}

const handleDeleteExpenseType = async (id: string) => {
  if (!confirm('Menghapus tipe pengeluaran juga akan menghapus akun biaya di dalamnya. Lanjutkan?')) return
  try {
    await apiClient.delete(`/kas-pabrik/master/expense-types/${id}`)
    toast.success('Tipe pengeluaran berhasil dihapus')
    await Promise.all([fetchExpenseTypes(), fetchExpenseAccounts()])
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menghapus tipe pengeluaran')
  }
}

// CRUD Master Akun Pengeluaran
const openAddExpenseAccountModal = (typeId?: string) => {
  isEditingExpenseAccount.value = false
  expenseAccountForm.value = {
    id: '',
    expense_type_id: typeId || (expenseTypeList.value[0]?.id || ''),
    account_code: '',
    account_name: '',
    accurate_account_id: '',
    description: '',
    sort_order: 0,
    is_active: true
  }
  showExpenseAccountModal.value = true
}

const openEditExpenseAccountModal = (acc: KasExpenseAccount) => {
  isEditingExpenseAccount.value = true
  expenseAccountForm.value = {
    id: acc.id,
    expense_type_id: acc.expense_type_id || acc.expense_type?.id || '',
    account_code: acc.account_code,
    account_name: acc.account_name,
    accurate_account_id: acc.accurate_account_id || '',
    description: acc.description || '',
    sort_order: acc.sort_order,
    is_active: acc.is_active
  }
  showExpenseAccountModal.value = true
}

const handleSaveExpenseAccount = async () => {
  isSavingExpenseAccount.value = true
  try {
    if (isEditingExpenseAccount.value) {
      await apiClient.put(`/kas-pabrik/master/expense-accounts/${expenseAccountForm.value.id}`, expenseAccountForm.value)
      toast.success('Akun pengeluaran berhasil diperbarui')
    } else {
      await apiClient.post('/kas-pabrik/master/expense-accounts', expenseAccountForm.value)
      toast.success('Akun pengeluaran berhasil ditambahkan')
    }
    showExpenseAccountModal.value = false
    await Promise.all([fetchExpenseTypes(), fetchExpenseAccounts()])
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menyimpan akun pengeluaran')
  } finally {
    isSavingExpenseAccount.value = false
  }
}

const handleDeleteExpenseAccount = async (id: string) => {
  if (!confirm('Apakah anda yakin ingin menghapus akun pengeluaran ini?')) return
  try {
    await apiClient.delete(`/kas-pabrik/master/expense-accounts/${id}`)
    toast.success('Akun pengeluaran berhasil dihapus')
    await Promise.all([fetchExpenseTypes(), fetchExpenseAccounts()])
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menghapus akun pengeluaran')
  }
}

// API Calls Kas Pabrik
const fetchBranches = async () => {
  try {
    const res = await apiClient.get('/branches')
    branches.value = res.data?.data || res.data || []
    if (branches.value.length > 0) {
      // Manajemen Kas Pabrik harus selalu memakai cabang Pabrik Utama.
      // Jangan fallback ke Cabang Timur karena saldo kasnya berbeda.
      const pabrikBranch = branches.value.find(b =>
        (b.name || '').toLowerCase().includes('pabrik utama')
      ) || branches.value.find(b =>
        (b.name || '').toLowerCase().includes('pabrik')
      )
      selectedBranch.value = pabrikBranch?.id || ''
      kasForm.value.branch_id = selectedBranch.value
    }
  } catch (error) {
    console.error('Error fetching branches', error)
  }
}

const fetchKasList = async () => {
  isLoadingKas.value = true
  try {
    const res = await apiClient.get(`/kas-pabrik?branch_id=${selectedBranch.value}`)
    kasList.value = res.data?.data || res.data || []
  } catch (error) {
    console.error('Error fetching kas', error)
    toast.error('Gagal mengambil data KAS')
  } finally {
    isLoadingKas.value = false
  }
}

const fetchTransactions = async () => {
  isLoadingTx.value = true
  try {
    const limitParam = txLimit.value <= 0 ? 'all' : txLimit.value
    let url = `/kas-pabrik/transactions?page=${txPage.value}&limit=${limitParam}`
    if (filterKas.value && filterKas.value !== 'all' && filterKas.value !== 'pending') {
      url += `&kas_pabrik_id=${filterKas.value}`
    }
    if (filterType.value) url += `&type=${filterType.value}`
    if (filterDateFrom.value) url += `&date_from=${filterDateFrom.value}`
    if (filterDateTo.value) url += `&date_to=${filterDateTo.value}`
    if (selectedBranch.value) url += `&branch_id=${selectedBranch.value}`
    if (filterAdvanceStatus.value) url += `&advance_status=${filterAdvanceStatus.value}`
    if (filterTargetBranchId.value) url += `&target_branch_id=${filterTargetBranchId.value}`
    if (filterEmployeeId.value) url += `&employee_id=${filterEmployeeId.value}`

    const res = await apiClient.get(url)
    const resData = res.data?.data
    if (resData && typeof resData === 'object' && ('data' in resData || 'items' in resData)) {
      transactions.value = resData.data || resData.items || []
      totalTx.value = resData.total || 0
    } else if (Array.isArray(resData)) {
      transactions.value = resData
      totalTx.value = resData.length
    } else {
      transactions.value = res.data?.items || res.data || []
      totalTx.value = transactions.value.length
    }
  } catch (error) {
    console.error('Error fetching tx', error)
    toast.error('Gagal mengambil riwayat transaksi')
  } finally {
    isLoadingTx.value = false
  }
}

const pendingDeposits = ref<any[]>([])
const isLoadingPending = ref(false)

const fetchPendingDeposits = async () => {
  isLoadingPending.value = true
  try {
    const res = await apiClient.get('/kas-pabrik/pending-deposits')
    pendingDeposits.value = res.data?.data || res.data || []
  } catch (error) {
    console.error('Error fetching pending deposits', error)
  } finally {
    isLoadingPending.value = false
  }
}

const handleApproveDeposit = async (id: string) => {
  if (!confirm('Apakah anda yakin menerima setoran ini?')) return
  try {
    await apiClient.post(`/kas-pabrik/pending-deposits/${id}/approve`)
    toast.success('Setoran berhasil diterima')
    fetchPendingDeposits()
    fetchKasList()
    fetchTransactions()
  } catch (error) {
    toast.error('Gagal menerima setoran')
  }
}

const handleRejectDeposit = async (id: string) => {
  if (!confirm('Apakah anda yakin menolak setoran ini?')) return
  try {
    await apiClient.post(`/kas-pabrik/pending-deposits/${id}/reject`)
    toast.success('Setoran berhasil ditolak')
    fetchPendingDeposits()
  } catch (error) {
    toast.error('Gagal menolak setoran')
  }
}

const handleAddKas = async () => {
  try {
    await apiClient.post('/kas-pabrik', kasForm.value)
    toast.success('KAS berhasil ditambahkan')
    showAddKasModal.value = false
    fetchKasList()
  } catch (error) {
    toast.error('Gagal menambahkan KAS')
  }
}

const handleTransaction = async (type: 'IN' | 'OUT') => {
  try {
    const form = type === 'IN' ? inForm.value : outForm.value
    const payload: any = {
      transaction_type: type,
      amount: form.amount,
      transaction_date: form.transaction_date,
      description: form.description,
      proof_url: form.proof_url,
    }

    if (type === 'IN') {
      const inVal = inForm.value
      payload.income_source_id = inVal.income_source_id
      payload.income_source = inVal.income_source
      if (inVal.target_branch_id) {
        payload.target_branch_id = inVal.target_branch_id
      }
    } else {
      const out = outForm.value
      payload.expense_type_id = out.expense_type_id
      payload.expense_account_id = out.expense_account_id
      payload.expense_category = out.expense_category
      if ((showVehicleField.value || isVehicleSuggested.value) && out.vehicle_id) {
        payload.vehicle_id = out.vehicle_id
      }
      if ((showBranchField.value || isBranchSuggested.value) && out.target_branch_id) {
        payload.is_branch_receivable = true
        payload.target_branch_id = out.target_branch_id
      }
      if ((showAdvanceField.value || isAdvanceSuggested.value) && out.employee_id) {
        payload.is_advance = true
        payload.employee_id = out.employee_id
      }
    }

    await apiClient.post(`/kas-pabrik/${form.kas_pabrik_id}/transactions`, payload)
    toast.success(`Transaksi ${type === 'IN' ? 'Masuk' : 'Keluar'} berhasil dicatat`)
    if (type === 'IN') showInModal.value = false
    else showOutModal.value = false
    fetchKasList()
    fetchTransactions()
  } catch (error: any) {
    toast.error(error.response?.data?.message || `Gagal mencatat transaksi ${type === 'IN' ? 'Masuk' : 'Keluar'}`)
  }
}

const handleTransfer = async () => {
  try {
    await apiClient.post('/kas-pabrik/transfer', transferForm.value)
    toast.success('Transfer berhasil')
    showTransferModal.value = false
    fetchKasList()
    fetchTransactions()
  } catch (error) {
    toast.error('Gagal melakukan transfer')
  }
}

const handleDeleteTx = async (txId: string) => {
  if (!confirm('Apakah anda yakin ingin menghapus transaksi ini? Saldo KAS akan disesuaikan kembali.')) return
  try {
    await apiClient.delete(`/kas-pabrik/transactions/${txId}`)
    toast.success('Transaksi berhasil dihapus')
    fetchKasList()
    fetchTransactions()
  } catch (error) {
    toast.error('Gagal menghapus transaksi')
  }
}

const handleEditKas = async () => {
  try {
    await apiClient.put(`/kas-pabrik/${editKasForm.value.id}`, editKasForm.value)
    toast.success('KAS berhasil diubah')
    showEditKasModal.value = false
    fetchKasList()
  } catch (error) {
    toast.error('Gagal mengubah KAS')
  }
}

const handleDeleteKas = async (kasId: string) => {
  if (!confirm('Apakah anda yakin ingin menghapus KAS ini? (Hanya bisa dihapus jika saldo 0)')) return
  try {
    await apiClient.delete(`/kas-pabrik/${kasId}`)
    toast.success('KAS berhasil dihapus')
    if (activeKasTab.value === kasId) switchTab('all')
    fetchKasList()
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menghapus KAS')
  }
}

const handleEditTx = async () => {
  try {
    await apiClient.put(`/kas-pabrik/transactions/${editTxForm.value.id}`, editTxForm.value)
    toast.success('Transaksi berhasil diubah')
    showEditTxModal.value = false
    fetchKasList()
    fetchTransactions()
  } catch (error) {
    toast.error('Gagal mengubah transaksi')
  }
}

const isRecalculating = ref(false)

const handleRecalculateKas = async (kasId?: string) => {
  let targetId = kasId || (activeKasTab.value !== 'all' && activeKasTab.value !== 'pending' ? activeKasTab.value : '')
  if (!targetId || targetId === 'all') {
    if (kasList.value.length === 0) return
    if (!confirm('Apakah anda ingin menghitung ulang dan menyinkronkan seluruh running balance buku kas pabrik secara kronologis?')) return
    isRecalculating.value = true
    try {
      for (const k of kasList.value) {
        await apiClient.post(`/kas-pabrik/${k.id}/recalculate`)
      }
      toast.success('Seluruh buku kas berhasil dihitung ulang dan disinkronkan!')
      await Promise.all([fetchKasList(), fetchTransactions()])
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Gagal menghitung ulang buku kas')
    } finally {
      isRecalculating.value = false
    }
    return
  }

  const foundKas = kasList.value.find(k => k.id === targetId)
  const kasName = foundKas ? foundKas.kas_name : 'KAS'
  if (!confirm(`Apakah anda ingin menghitung ulang dan menyinkronkan saldo buku kas ${kasName}?`)) return
  isRecalculating.value = true
  try {
    await apiClient.post(`/kas-pabrik/${targetId}/recalculate`)
    toast.success(`Buku kas ${kasName} berhasil dihitung ulang dan disinkronkan!`)
    await Promise.all([fetchKasList(), fetchTransactions()])
  } catch (error: any) {
    toast.error(error.response?.data?.message || 'Gagal menghitung ulang buku kas')
  } finally {
    isRecalculating.value = false
  }
}

// Watchers and lifecycle
watch(selectedBranch, () => {
  fetchKasList()
  kasForm.value.branch_id = selectedBranch.value
  if (activeKasTab.value === 'pending') {
    fetchPendingDeposits()
  } else {
    fetchTransactions()
  }
  fetchVehiclesAndEmployees()
})

const onIncomeSourceChange = (srcId: string) => {
  const found = incomeSourceList.value.find(s => s.id === srcId)
  if (found) {
    inForm.value.income_source = found.code
  }
}

const onEditIncomeSourceChange = (srcId: string) => {
  const found = incomeSourceList.value.find(s => s.id === srcId)
  if (found) {
    editTxForm.value.income_source = found.code
  }
}

const openInModal = (kasId?: string) => {
  const firstSrc = incomeSourceList.value[0]
  inForm.value = {
    kas_pabrik_id: kasId || '',
    amount: 0,
    income_source_id: firstSrc?.id || '',
    income_source: firstSrc?.code || '',
    target_branch_id: '',
    transaction_date: new Date().toISOString().split('T')[0],
    description: '',
    proof_url: ''
  }
  showInModal.value = true
}

const openOutModal = (kasId?: string) => {
  outForm.value = {
    kas_pabrik_id: kasId || (kasList.value[0]?.id || ''),
    amount: 0,
    expense_type_id: expenseTypeList.value[0]?.id || '',
    expense_account_id: '',
    expense_category: '',
    expense_type: 'OPERATIONAL',
    vehicle_id: '',
    target_branch_id: '',
    employee_id: '',
    transaction_date: new Date().toISOString().split('T')[0],
    description: '',
    proof_url: ''
  }
  showVehicleField.value = false
  showBranchField.value = false
  showAdvanceField.value = false
  showExpenseAccountDropdown.value = false
  expenseAccountSearch.value = ''
  if (vehicles.value.length === 0) {
    fetchVehiclesAndEmployees()
  }
  showOutModal.value = true
}

const openTransferModal = () => {
  transferForm.value = {
    from_kas_id: '',
    to_kas_id: '',
    amount: 0,
    transaction_date: new Date().toISOString().split('T')[0],
    description: ''
  }
  showTransferModal.value = true
}

const switchTab = (kasId: string) => {
  activeKasTab.value = kasId
  if (kasId === 'pending') {
    filterKas.value = ''
    fetchPendingDeposits()
    return
  }
  filterKas.value = kasId === 'all' ? '' : kasId
  txPage.value = 1
  fetchTransactions()
}

const setSubTab = (tab: 'ALL' | 'ADVANCE_OUTSTANDING' | 'BRANCH_RECEIVABLE' | 'VEHICLE') => {
  activeSubTab.value = tab
  if (tab === 'ALL') {
    filterAdvanceStatus.value = ''
    filterTargetBranchId.value = ''
  } else if (tab === 'ADVANCE_OUTSTANDING') {
    filterAdvanceStatus.value = 'OUTSTANDING'
    filterTargetBranchId.value = ''
  }
  txPage.value = 1
  fetchTransactions()
}

const openEditKasModal = (kas: KasPabrik) => {
  editKasForm.value = {
    id: kas.id,
    kas_code: kas.kas_code,
    kas_name: kas.kas_name,
    description: kas.description,
    bank_name: kas.bank_name || '',
    account_number: kas.account_number || '',
    sort_order: kas.sort_order,
    is_active: kas.is_active
  }
  showEditKasModal.value = true
}

const openEditTxModal = (tx: KasTransaction) => {
  let matchedSrcId = tx.income_source_id || ''
  if (!matchedSrcId && tx.income_source) {
    const src = incomeSourceList.value.find(s => s.code === tx.income_source)
    if (src) matchedSrcId = src.id
  }

  let matchedTypeId = tx.expense_type_id || ''
  let matchedAccId = tx.expense_account_id || ''
  if (!matchedAccId && tx.expense_category) {
    const acc = expenseAccountList.value.find(a => a.account_code === tx.expense_category || a.account_name === tx.expense_category)
    if (acc) {
      matchedAccId = acc.id
      if (!matchedTypeId) matchedTypeId = acc.expense_type_id
    }
  }

  editTxForm.value = {
    id: tx.id,
    transaction_type: tx.transaction_type,
    amount: tx.amount,
    income_source_id: matchedSrcId,
    income_source: tx.income_source || '',
    expense_type_id: matchedTypeId,
    expense_account_id: matchedAccId,
    expense_category: tx.expense_category || '',
    transaction_date: tx.transaction_date ? (tx.transaction_date.split('T')[0] || '') : '',
    description: tx.description || '',
    proof_url: tx.proof_url || ''
  }
  showEditExpenseAccountDropdown.value = false
  editExpenseAccountSearch.value = ''
  showEditTxModal.value = true
}

onMounted(async () => {
  await Promise.all([
    fetchBranches(),
    fetchIncomeSources(),
    fetchExpenseTypes(),
    fetchExpenseAccounts()
  ])
  if (selectedBranch.value) {
    fetchKasList()
    fetchTransactions()
  }
  fetchPendingDeposits()
  fetchVehiclesAndEmployees()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon-box">
          <Landmark class="header-icon" />
        </div>
        <div>
          <h1 class="page-title">Manajemen KAS Pabrik</h1>
          <p class="page-subtitle">Kelola KAS Kasir Pabrik — pantau saldo & riwayat transaksi</p>
        </div>
      </div>
      <div class="header-actions">
        <div class="inline-flex items-center gap-1.5 px-3 py-2 bg-indigo-50/80 text-indigo-700 font-bold text-xs rounded-xl border border-indigo-200/60 shadow-xs">
          <Building class="w-4 h-4 text-indigo-600" />
          <span>{{ currentBranchName }}</span>
        </div>
        <button class="btn-secondary" @click="fetchKasList(); fetchTransactions()">
          <RefreshCw class="w-4 h-4" :class="{ 'spin': isLoadingKas }" />
          Refresh
        </button>
        <button class="btn-primary" @click="showAddKasModal = true" v-if="mainTab === 'BOOK'">
          <Plus class="w-4 h-4" />
          Tambah KAS
        </button>
      </div>
    </div>

    <!-- Main Navigation Tabs -->
    <div class="flex flex-wrap items-center gap-2 mb-6 border-b border-slate-200 pb-2">
      <button 
        class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs sm:text-sm transition-all shadow-sm cursor-pointer"
        :class="mainTab === 'BOOK' ? 'bg-indigo-600 text-white shadow-indigo-200' : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'"
        @click="mainTab = 'BOOK'"
      >
        <Landmark class="w-4 h-4" />
        Buku Kas & Transaksi
      </button>
      <button 
        class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs sm:text-sm transition-all shadow-sm cursor-pointer"
        :class="mainTab === 'INCOME_SOURCE' ? 'bg-indigo-600 text-white shadow-indigo-200' : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'"
        @click="mainTab = 'INCOME_SOURCE'"
      >
        <ArrowDownToLine class="w-4 h-4" />
        Master Sumber Penerimaan
        <span class="px-2 py-0.5 rounded-full text-xs font-bold" :class="mainTab === 'INCOME_SOURCE' ? 'bg-white/20 text-white' : 'bg-slate-100 text-slate-700'">{{ incomeSourceList.length }}</span>
      </button>
      <button 
        class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs sm:text-sm transition-all shadow-sm cursor-pointer"
        :class="mainTab === 'EXPENSE_TYPE' ? 'bg-indigo-600 text-white shadow-indigo-200' : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'"
        @click="mainTab = 'EXPENSE_TYPE'"
      >
        <Layers class="w-4 h-4" />
        Master Tipe & Akun Pengeluaran
        <span class="px-2 py-0.5 rounded-full text-xs font-bold" :class="mainTab === 'EXPENSE_TYPE' ? 'bg-white/20 text-white' : 'bg-slate-100 text-slate-700'">{{ expenseTypeList.length }}</span>
      </button>
    </div>

    <!-- TAB 1: BUKU KAS & TRANSAKSI -->
    <div v-if="mainTab === 'BOOK'">
      <!-- KAS Tabs -->
      <div class="kas-tabs">
        <button class="kas-tab" :class="activeKasTab === 'all' ? 'active-all' : ''" @click="switchTab('all')">
          Semua KAS
        </button>
      <button class="kas-tab relative" :class="activeKasTab === 'pending' ? 'active-all' : ''" @click="switchTab('pending')">
        Menunggu Validasi
        <span v-if="pendingDeposits.length > 0" class="absolute -top-2 -right-2 bg-red-500 text-white text-xs font-bold px-1.5 py-0.5 rounded-full">{{ pendingDeposits.length }}</span>
      </button>
      <button v-for="kas in kasList" :key="kas.id" class="kas-tab" 
              :class="activeKasTab === kas.id ? `active-${getKasKey(kas.kas_name)}` : ''"
              @click="switchTab(kas.id)">
        {{ kas.kas_name }}
        <span class="tab-balance">{{ formatRupiah(kas.balance) }}</span>
      </button>
    </div>

    <!-- KAS Hero Card (Active Tab) -->
    <div v-if="activeKas" class="kas-hero-card" :class="getKasColorClass(activeKas.kas_name)">
      <div class="hero-top">
        <span class="kas-code">{{ activeKas.kas_code }}</span>
        <span class="hero-badge">AKTIF</span>
      </div>
      <h2 class="hero-name">{{ activeKas.kas_name }}</h2>
      <div class="hero-balance">{{ formatRupiah(activeKas.balance) }}</div>
      <p v-if="activeKas.description" class="hero-desc">{{ activeKas.description }}</p>
      <div v-if="activeKas.bank_name" class="hero-bank">
        <Wallet class="w-4 h-4" /> {{ activeKas.bank_name }} - {{ activeKas.account_number }}
      </div>
      <div class="hero-actions">
        <button class="btn-kas-action in" @click.stop="openInModal(activeKas.id)">
          <ArrowDownToLine class="w-4 h-4" /> Catat Masuk
        </button>
        <button class="btn-kas-action out" @click.stop="openOutModal(activeKas.id)">
          <ArrowUpFromLine class="w-4 h-4" /> Catat Keluar
        </button>
        <button class="btn-kas-action transfer" @click.stop="openTransferModal()">
          <ArrowRightLeft class="w-4 h-4" /> Transfer
        </button>
        <button class="btn-kas-action edit" @click.stop="openEditKasModal(activeKas)">
          <Pencil class="w-4 h-4" /> Edit KAS
        </button>
        <button class="btn-kas-action delete" @click.stop="handleDeleteKas(activeKas.id)">
          <Trash2 class="w-4 h-4" /> Hapus
        </button>
      </div>
    </div>

    <!-- Pending Deposits Table -->
    <div v-else-if="activeKasTab === 'pending'" class="content-section">
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Tanggal</th>
              <th>Cabang Pengirim</th>
              <th>KAS Tujuan</th>
              <th class="text-right">Nominal</th>
              <th>Keterangan</th>
              <th class="text-center">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="isLoadingPending">
              <td colspan="6" class="text-center py-8">
                <RefreshCw class="w-6 h-6 spin mx-auto text-indigo-500" />
              </td>
            </tr>
            <tr v-else-if="pendingDeposits.length === 0">
              <td colspan="6" class="text-center py-8 text-gray-500">
                Tidak ada setoran yang menunggu validasi.
              </td>
            </tr>
            <tr v-else v-for="dep in pendingDeposits" :key="dep.id">
              <td>{{ formatDate(dep.mutation_date) }}</td>
              <td class="font-medium text-gray-900">{{ dep.branch?.name }}</td>
              <td>{{ dep.target_kas?.kas_name }}</td>
              <td class="text-right text-green-600 font-medium">{{ formatRupiah(dep.amount) }}</td>
              <td>
                <div class="text-xs text-gray-500">{{ dep.notes }}</div>
                <a v-if="dep.proof_url" :href="dep.proof_url" target="_blank" class="text-indigo-500 text-xs mt-1 inline-block hover:underline">Lihat Bukti</a>
              </td>
              <td class="text-center">
                <div class="flex items-center justify-center gap-2">
                  <button class="bg-green-500 hover:bg-green-600 text-white rounded text-xs py-1 px-2 font-medium" @click="handleApproveDeposit(dep.id)">Terima</button>
                  <button class="bg-red-500 hover:bg-red-600 text-white rounded text-xs py-1 px-2 font-medium" @click="handleRejectDeposit(dep.id)">Tolak</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- KAS Cards Grid (All Tab) -->
    <div v-else class="kas-grid">
      <div v-for="kas in displayedKasList" :key="kas.id" class="kas-card" :class="getKasColorClass(kas.kas_name)" @click="switchTab(kas.id)">
        <div class="kas-card-header">
          <span class="kas-name">{{ kas.kas_name }}</span>
          <span class="kas-code">{{ kas.kas_code }}</span>
        </div>
        <div class="kas-card-body">
          <div class="balance-label">Saldo Saat Ini</div>
          <div class="balance-value">{{ formatRupiah(kas.balance) }}</div>
          <div v-if="kas.bank_name" class="bank-info">
            <Wallet class="w-3 h-3" />
            {{ kas.bank_name }} - {{ kas.account_number }}
          </div>
        </div>
        <div class="kas-card-footer" @click.stop>
          <button class="btn-kas-action in" @click.stop="openInModal(kas.id)">
            <ArrowDownToLine class="w-3 h-3" /> Masuk
          </button>
          <button class="btn-kas-action out" @click.stop="openOutModal(kas.id)">
            <ArrowUpFromLine class="w-3 h-3" /> Keluar
          </button>
          <button class="btn-kas-action edit" @click.stop="openEditKasModal(kas)">
            <Pencil class="w-3 h-3" /> Edit
          </button>
        </div>
      </div>
      <div v-if="isLoadingKas" class="kas-card loading-skeleton" v-for="i in 3" :key="i"></div>
      <div v-if="!isLoadingKas && kasList.length === 0" class="empty-kas-message">
        Belum ada KAS yang ditambahkan.
      </div>
    </div>

    <!-- Actions & Filters -->
    <div v-if="activeKasTab !== 'pending'" class="content-section">
      <div class="section-toolbar">
        <div class="action-buttons">
          <button class="btn-success" @click="openInModal()">
            <ArrowDownToLine class="w-4 h-4" />
            Catat Penerimaan
          </button>
          <button class="btn-danger" @click="openOutModal()">
            <ArrowUpFromLine class="w-4 h-4" />
            Catat Pengeluaran
          </button>
          <button class="btn-indigo" @click="openTransferModal()">
            <ArrowRightLeft class="w-4 h-4" />
            Transfer Antar KAS
          </button>
          <button 
            class="px-3.5 py-2 text-xs font-bold rounded-xl border border-indigo-200 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 flex items-center gap-1.5 shadow-sm transition-all cursor-pointer"
            @click="handleRecalculateKas()"
            :disabled="isRecalculating"
            title="Hitung ulang running balance seluruh transaksi buku kas secara kronologis beruntun"
          >
            <RefreshCw class="w-3.5 h-3.5 text-indigo-600" :class="{ 'spin': isRecalculating }" />
            {{ isRecalculating ? 'Menyinkronkan...' : 'Sinkronkan & Hitung Ulang Saldo' }}
          </button>
        </div>
      </div>

      <!-- Sub-filter Tabs for Linda Kasir Pusat -->
      <div class="flex flex-wrap gap-2 mb-4 bg-slate-100 p-1.5 rounded-xl border border-slate-200">
        <button 
          class="px-4 py-2 rounded-lg text-xs font-bold transition-all"
          :class="activeSubTab === 'ALL' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500 hover:text-slate-800'"
          @click="setSubTab('ALL')"
        >
          Semua Transaksi
        </button>
        <button 
          class="px-4 py-2 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5"
          :class="activeSubTab === 'ADVANCE_OUTSTANDING' ? 'bg-purple-600 text-white shadow-sm' : 'text-purple-700 hover:bg-purple-50'"
          @click="setSubTab('ADVANCE_OUTSTANDING')"
        >
          <User class="w-3.5 h-3.5" />
          Kasbon Dinas (Audit/Karyawan)
        </button>
      </div>

      <div class="filter-bar">
        <div class="filter-group">
          <select v-model="filterKas" class="filter-input">
            <option value="">Semua KAS</option>
            <option v-for="kas in kasList" :key="kas.id" :value="kas.id">{{ kas.kas_name }}</option>
          </select>
          <select v-model="filterType" class="filter-input">
            <option value="">Semua Tipe</option>
            <option value="IN">Penerimaan (IN)</option>
            <option value="OUT">Pengeluaran (OUT)</option>
            <option value="TRANSFER_IN">Transfer Masuk</option>
            <option value="TRANSFER_OUT">Transfer Keluar</option>
          </select>
          <select v-model="filterAdvanceStatus" class="filter-input" v-if="activeSubTab === 'ALL'">
            <option value="">Status Kasbon: Semua</option>
            <option value="OUTSTANDING">Kasbon Belum Selesai (Outstanding)</option>
            <option value="SETTLED">Kasbon Selesai (Settled)</option>
          </select>
          <input type="date" v-model="filterDateFrom" class="filter-input" placeholder="Dari Tanggal" />
          <input type="date" v-model="filterDateTo" class="filter-input" placeholder="Sampai Tanggal" />
          <button class="btn-secondary" @click="fetchTransactions()">
            <Search class="w-4 h-4" /> Cari
          </button>
        </div>
      </div>

      <!-- Transaction Table -->
      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th>Tanggal</th>
              <th>KAS</th>
              <th>Tipe</th>
              <th>Keterangan & Rincian</th>
              <th class="text-right">Masuk (+)</th>
              <th class="text-right">Keluar (-)</th>
              <th class="text-right">
                <div>Saldo Akhir</div>
                <div class="text-[10px] font-normal text-slate-400 font-sans tracking-tight">Kronologis (&uarr; Baru ke Lama)</div>
              </th>
              <th class="text-center">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="isLoadingTx">
              <td colspan="8" class="text-center py-8">
                <RefreshCw class="w-6 h-6 spin mx-auto text-indigo-500" />
                <p class="mt-2 text-sm text-gray-500">Memuat transaksi...</p>
              </td>
            </tr>
            <tr v-else-if="transactions.length === 0">
              <td colspan="8" class="text-center py-8 text-gray-500">
                Tidak ada transaksi ditemukan.
              </td>
            </tr>
            <tr v-else v-for="tx in transactions" :key="tx.id">
              <td>{{ formatDate(tx.transaction_date) }}</td>
              <td class="font-medium text-gray-900">
                <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-bold"
                  :class="getKasColorBadgeClass(tx.kas_pabrik?.kas_name)">
                  {{ tx.kas_pabrik?.kas_name || '-' }}
                </span>
              </td>
              <td>
                <span class="badge" :class="getTxTypeBadgeClass(tx.transaction_type)">
                  {{ tx.transaction_type }}
                </span>
              </td>
              <td>
                <div class="text-sm font-medium text-gray-900">
                  {{ getCategoryOrSourceLabel(tx) }}
                </div>
                <div class="text-xs text-gray-500">{{ tx.description }}</div>
                
                <!-- Extended Badges: Armada Kendaraan, Piutang Cabang, Kasbon Dinas -->
                <div class="flex flex-wrap gap-1 mt-1.5">
                  <span v-if="tx.is_branch_receivable && tx.target_branch" class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-bold bg-amber-100 text-amber-900 border border-amber-200">
                    <Building class="w-3 h-3 text-amber-700" /> Piutang Cabang: {{ tx.target_branch.name }}
                  </span>
                  <span v-if="tx.vehicle" class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-bold bg-blue-100 text-blue-900 border border-blue-200">
                    <Truck class="w-3 h-3 text-blue-700" /> Armada: {{ tx.vehicle.name }} ({{ tx.vehicle.license_plate }})
                  </span>
                  <span v-if="tx.advance_status === 'OUTSTANDING'" class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-bold bg-purple-100 text-purple-900 border border-purple-200">
                    <User class="w-3 h-3 text-purple-700" /> Kasbon Dinas: {{ tx.employee ? tx.employee.first_name : 'Karyawan' }} (OUTSTANDING)
                  </span>
                  <span v-else-if="tx.advance_status === 'SETTLED'" class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-bold bg-emerald-100 text-emerald-900 border border-emerald-200">
                    <CheckCircle2 class="w-3 h-3 text-emerald-700" /> Kasbon Selesai
                    <template v-if="tx.settlement">
                      (Realisasi: {{ formatRupiah(tx.settlement.actual_expense_amount) }}
                      <span v-if="tx.settlement.refund_amount > 0">, Sisa Kembali: {{ formatRupiah(tx.settlement.refund_amount) }}</span>)
                    </template>
                  </span>
                  <span v-else-if="tx.employee" class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-medium bg-slate-100 text-slate-700">
                    <User class="w-3 h-3" /> {{ tx.employee.first_name }}
                  </span>
                </div>
              </td>
              <td class="text-right text-green-600 font-medium">
                {{ ['IN', 'TRANSFER_IN'].includes(tx.transaction_type) ? formatRupiah(tx.amount) : '-' }}
              </td>
              <td class="text-right text-red-600 font-medium">
                {{ ['OUT', 'TRANSFER_OUT'].includes(tx.transaction_type) ? formatRupiah(tx.amount) : '-' }}
              </td>
              <td class="text-right font-medium text-gray-900">
                <div class="font-bold font-mono">{{ formatRupiah(tx.balance_after) }}</div>
                <div v-if="activeKasTab === 'all' && tx.kas_pabrik" class="text-[10px] font-bold text-slate-400">
                  {{ tx.kas_pabrik.kas_name }}
                </div>
              </td>
              <td class="text-center">
                <div class="flex items-center justify-center gap-1.5 flex-wrap">
                  <button 
                    v-if="tx.advance_status === 'OUTSTANDING'" 
                    class="px-2.5 py-1 text-xs font-bold bg-purple-600 hover:bg-purple-700 text-white rounded-lg shadow-sm transition-all flex items-center gap-1"
                    @click="openSettleModal(tx)"
                    title="Selesaikan Kasbon / Settle Realisasi Nota"
                  >
                    <Coins class="w-3.5 h-3.5" />
                    Selesaikan Kasbon
                  </button>
                  <button class="btn-icon text-indigo-500 hover:bg-indigo-50" @click="openEditTxModal(tx)" title="Edit Transaksi">
                    Edit
                  </button>
                  <button class="btn-icon btn-delete" @click="handleDeleteTx(tx.id)" title="Hapus Transaksi">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
          <tfoot v-if="transactions.length > 0" class="bg-slate-100 font-bold border-t-2 border-slate-300">
            <tr>
              <td colspan="4" class="text-right py-3.5 px-4 text-slate-700">
                <div class="flex items-center justify-end gap-2">
                  <Calculator class="w-4 h-4 text-indigo-600" />
                  <span>TOTAL MUTASI TERFILTER ({{ transactions.length }} Data):</span>
                </div>
              </td>
              <td class="text-right py-3.5 px-4 text-emerald-600 text-sm font-extrabold">
                {{ formatRupiah(totalFilteredIn) }}
              </td>
              <td class="text-right py-3.5 px-4 text-rose-600 text-sm font-extrabold">
                {{ formatRupiah(totalFilteredOut) }}
              </td>
              <td class="text-right py-3.5 px-4 text-sm font-extrabold" :class="netFilteredMutation >= 0 ? 'text-indigo-700' : 'text-amber-700'">
                {{ formatRupiah(netFilteredMutation) }}
              </td>
              <td class="text-center py-3.5 px-2">
                <button @click="copyFilterSummary" class="text-xs px-2.5 py-1 bg-white border border-slate-300 rounded shadow-sm hover:bg-slate-50 text-slate-700 inline-flex items-center gap-1 font-medium" title="Salin Ringkasan Filter ke Clipboard">
                  <Copy class="w-3.5 h-3.5 text-slate-500" /> Salin
                </button>
              </td>
            </tr>
          </tfoot>
        </table>

        <!-- Sticky Auto Calculation Bar / Summary Panel for Instant Reporting -->
        <div v-if="transactions.length > 0" class="mt-4 p-4 bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 text-white rounded-xl shadow-lg border border-indigo-500/20 flex flex-col md:flex-row items-center justify-between gap-4">
          <div class="flex items-center gap-3">
            <div class="p-2.5 bg-indigo-500/20 rounded-lg text-indigo-400 border border-indigo-400/30">
              <Calculator class="w-6 h-6" />
            </div>
            <div>
              <div class="text-xs uppercase tracking-wider font-semibold text-indigo-300">Kalkulasi Otomatis Buku Kas Terfilter</div>
              <div class="text-sm font-medium text-slate-200">
                Menghitung <strong class="text-white">{{ transactions.length }}</strong> transaksi
                <span v-if="activeKas" class="text-indigo-300 font-semibold"> • {{ activeKas.kas_name }}</span>
                <span v-if="filterType" class="text-amber-300 font-semibold"> • {{ filterType === 'IN' ? 'Penerimaan Saja' : filterType === 'OUT' ? 'Pengeluaran Saja' : filterType }}</span>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 w-full md:w-auto text-right">
            <div class="bg-black/40 px-3.5 py-2 rounded-lg border border-emerald-500/30">
              <div class="text-[11px] font-semibold text-emerald-400 uppercase">Total Masuk (+)</div>
              <div class="text-base font-extrabold text-emerald-300">{{ formatRupiah(totalFilteredIn) }}</div>
            </div>
            <div class="bg-black/40 px-3.5 py-2 rounded-lg border border-rose-500/30">
              <div class="text-[11px] font-semibold text-rose-400 uppercase">Total Keluar (-)</div>
              <div class="text-base font-extrabold text-rose-300">{{ formatRupiah(totalFilteredOut) }}</div>
            </div>
            <div class="bg-black/40 px-3.5 py-2 rounded-lg border border-indigo-500/30">
              <div class="text-[11px] font-semibold text-indigo-300 uppercase">Net Mutasi</div>
              <div class="text-base font-extrabold" :class="netFilteredMutation >= 0 ? 'text-indigo-200' : 'text-amber-300'">
                {{ formatRupiah(netFilteredMutation) }}
              </div>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <button @click="copyFilterSummary" class="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-lg shadow transition-all flex items-center gap-1.5 whitespace-nowrap">
              <Copy class="w-3.5 h-3.5" />
              Salin Ringkasan Direksi
            </button>
          </div>
        </div>

        <!-- Pagination & Rows per page -->
        <div class="pagination-bar" v-if="totalTx > 0">
          <div class="pagination-info">
            <span>Menampilkan <strong>{{ startEntry }}</strong> - <strong>{{ endEntry }}</strong> dari <strong>{{ totalTx }}</strong> transaksi</span>
            <div class="limit-selector">
              <label>Tampilkan:</label>
              <select :value="txLimit" @change="(e: any) => { txLimit = Number(e.target.value); handleLimitChange() }">
                <option :value="10">10 data</option>
                <option :value="25">25 data</option>
                <option :value="50">50 data</option>
                <option :value="100">100 data</option>
                <option :value="0">Semua Data</option>
              </select>
            </div>
          </div>

          <div class="pagination-controls" v-if="txLimit > 0 && totalPages > 1">
            <button
              class="btn-page"
              :disabled="txPage <= 1"
              @click="goToPage(txPage - 1)"
              title="Halaman Sebelumnya"
            >
              <ChevronLeft class="w-4 h-4" />
              <span>Prev</span>
            </button>

            <div class="page-numbers">
              <button
                v-for="p in totalPages"
                :key="p"
                v-show="p === 1 || p === totalPages || (p >= txPage - 2 && p <= txPage + 2)"
                class="btn-page-num"
                :class="{ active: p === txPage }"
                @click="goToPage(p)"
              >
                {{ p }}
              </button>
            </div>

            <button
              class="btn-page"
              :disabled="txPage >= totalPages"
              @click="goToPage(txPage + 1)"
              title="Halaman Selanjutnya"
            >
              <span>Next</span>
              <ChevronRight class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
    </div>
    <!-- END TAB 1: BUKU KAS -->

    <!-- TAB 2: MASTER SUMBER PENERIMAAN -->
    <div v-else-if="mainTab === 'INCOME_SOURCE'" class="content-section space-y-4">
      <div class="bg-gradient-to-r from-emerald-900 via-teal-900 to-slate-900 text-white p-6 rounded-2xl shadow-lg border border-emerald-500/20 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div>
          <div class="flex items-center gap-2">
            <span class="p-2 bg-emerald-500/20 rounded-lg text-emerald-400 border border-emerald-400/30">
              <ArrowDownToLine class="w-5 h-5" />
            </span>
            <h2 class="text-xl font-black">Master Sumber Penerimaan Kas</h2>
          </div>
          <p class="text-xs text-emerald-200 mt-1 max-w-xl">
            Kelola klasifikasi sumber dana masuk kasir pabrik yang terintegrasi dengan kode akun COA Accurate.
          </p>
        </div>
        <button 
          class="px-4 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs rounded-xl shadow-lg transition-all flex items-center gap-1.5 cursor-pointer"
          @click="openAddIncomeSourceModal"
        >
          <Plus class="w-4 h-4" />
          Tambah Sumber Penerimaan
        </button>
      </div>

      <div class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th class="w-12 text-center">Urutan</th>
              <th>Kode Sumber</th>
              <th>Nama Sumber Penerimaan</th>
              <th>Akun COA / Accurate</th>
              <th>Deskripsi</th>
              <th class="text-center">Status</th>
              <th class="text-center">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="isLoadingIncomeSources">
              <td colspan="7" class="text-center py-8">
                <RefreshCw class="w-6 h-6 spin mx-auto text-emerald-500" />
                <p class="mt-2 text-xs text-gray-500">Memuat sumber penerimaan...</p>
              </td>
            </tr>
            <tr v-else-if="incomeSourceList.length === 0">
              <td colspan="7" class="text-center py-8 text-gray-500">
                Belum ada master sumber penerimaan. Klik tombol di atas untuk menambah.
              </td>
            </tr>
            <tr v-else v-for="src in incomeSourceList" :key="src.id">
              <td class="text-center font-mono text-xs text-slate-500">{{ src.sort_order }}</td>
              <td>
                <span class="font-mono text-xs font-bold px-2 py-0.5 rounded bg-slate-100 text-slate-800 border border-slate-200">
                  {{ src.code }}
                </span>
              </td>
              <td class="font-bold text-slate-900 text-sm">
                {{ src.name }}
              </td>
              <td>
                <div v-if="src.account_code || src.account_name" class="text-xs">
                  <span class="font-bold text-indigo-700 font-mono">{{ src.account_code }}</span>
                  <span v-if="src.account_name" class="text-slate-600"> - {{ src.account_name }}</span>
                </div>
                <span v-else class="text-xs text-slate-400 italic">Belum di-mapping</span>
              </td>
              <td class="text-xs text-slate-500 max-w-xs">{{ src.description || '-' }}</td>
              <td class="text-center">
                <span 
                  class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-bold"
                  :class="src.is_active ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-600'"
                >
                  {{ src.is_active ? 'Aktif' : 'Non-aktif' }}
                </span>
              </td>
              <td class="text-center">
                <div class="flex items-center justify-center gap-1.5">
                  <button 
                    class="p-1.5 text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                    @click="openEditIncomeSourceModal(src)"
                    title="Edit Sumber Penerimaan"
                  >
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button 
                    class="p-1.5 text-rose-600 hover:bg-rose-50 rounded-lg transition-colors"
                    @click="handleDeleteIncomeSource(src.id)"
                    title="Hapus Sumber Penerimaan"
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
    <!-- END TAB 2: MASTER SUMBER PENERIMAAN -->

    <!-- TAB 3: MASTER TIPE & AKUN PENGELUARAN -->
    <div v-else-if="mainTab === 'EXPENSE_TYPE'" class="content-section space-y-6">
      <div class="bg-gradient-to-r from-indigo-900 via-slate-900 to-slate-900 text-white p-6 rounded-2xl shadow-lg border border-indigo-500/20 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
        <div>
          <div class="flex items-center gap-2">
            <span class="p-2 bg-indigo-500/20 rounded-lg text-indigo-400 border border-indigo-400/30">
              <Layers class="w-5 h-5" />
            </span>
            <h2 class="text-xl font-black">Master Tipe & Akun Pengeluaran (Accurate COA)</h2>
          </div>
          <p class="text-xs text-indigo-200 mt-1 max-w-xl">
            Struktur hierarki kategori tipe pengeluaran berjenjang beserta rincian akun biaya COA yang berelasi langsung dengan sistem pembukuan.
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button 
            class="px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs rounded-xl shadow-lg transition-all flex items-center gap-1.5 cursor-pointer"
            @click="openAddExpenseTypeModal"
          >
            <Plus class="w-4 h-4" />
            Tambah Tipe Pengeluaran
          </button>
        </div>
      </div>

      <!-- Expense Types Hierarchy Cards -->
      <div v-if="isLoadingExpenseTypes" class="text-center py-12">
        <RefreshCw class="w-8 h-8 spin mx-auto text-indigo-500" />
        <p class="mt-2 text-xs text-gray-500">Memuat tipe & akun pengeluaran...</p>
      </div>

      <div v-else-if="expenseTypeList.length === 0" class="bg-white p-8 rounded-2xl border border-slate-200 text-center text-slate-500">
        Belum ada tipe pengeluaran. Klik tombol di atas untuk membuat tipe pengeluaran baru.
      </div>

      <div v-else class="space-y-6">
        <div 
          v-for="et in expenseTypeList" 
          :key="et.id" 
          class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden transition-all hover:border-indigo-300"
        >
          <!-- Tipe Header -->
          <div class="p-4 sm:p-5 bg-gradient-to-r from-slate-50 via-indigo-50/30 to-white border-b border-slate-200 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div class="flex items-start sm:items-center gap-3">
              <div class="p-2.5 rounded-xl bg-indigo-600 text-white shadow-sm shrink-0">
                <Layers class="w-5 h-5" />
              </div>
              <div>
                <div class="flex items-center gap-2 flex-wrap">
                  <h3 class="font-extrabold text-slate-900 text-base">{{ et.name }}</h3>
                  <span class="font-mono text-[11px] font-bold px-2 py-0.5 rounded bg-indigo-100 text-indigo-800 border border-indigo-200">
                    {{ et.code }}
                  </span>
                  <span 
                    class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold"
                    :class="et.is_active ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ et.is_active ? 'Aktif' : 'Non-aktif' }}
                  </span>
                </div>
                <p v-if="et.description" class="text-xs text-slate-500 mt-0.5">{{ et.description }}</p>
                <div class="text-[11px] text-slate-400 mt-1 font-medium">
                  {{ (et.accounts || []).length }} Akun Biaya Terdaftar
                </div>
              </div>
            </div>

            <!-- Actions for Type -->
            <div class="flex items-center gap-2 shrink-0 self-end sm:self-center">
              <button 
                class="px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 font-bold text-xs rounded-lg border border-indigo-200 transition-colors flex items-center gap-1 cursor-pointer"
                @click="openAddExpenseAccountModal(et.id)"
              >
                <Plus class="w-3.5 h-3.5" />
                Tambah Akun Biaya
              </button>
              <button 
                class="p-1.5 text-slate-600 hover:bg-slate-100 rounded-lg transition-colors cursor-pointer"
                @click="openEditExpenseTypeModal(et)"
                title="Edit Tipe"
              >
                <Pencil class="w-4 h-4" />
              </button>
              <button 
                class="p-1.5 text-rose-600 hover:bg-rose-50 rounded-lg transition-colors cursor-pointer"
                @click="handleDeleteExpenseType(et.id)"
                title="Hapus Tipe"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Accounts Table inside Type -->
          <div class="p-0 overflow-x-auto">
            <table class="w-full text-left border-collapse text-xs">
              <thead>
                <tr class="bg-slate-50/80 text-slate-600 border-b border-slate-200">
                  <th class="py-2.5 px-4 font-semibold w-12 text-center">Urutan</th>
                  <th class="py-2.5 px-4 font-semibold">Kode Akun Biaya</th>
                  <th class="py-2.5 px-4 font-semibold">Nama Akun Biaya</th>
                  <th class="py-2.5 px-4 font-semibold">ID / Akun Accurate COA</th>
                  <th class="py-2.5 px-4 font-semibold">Deskripsi</th>
                  <th class="py-2.5 px-4 font-semibold text-center">Status</th>
                  <th class="py-2.5 px-4 font-semibold text-center">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-if="!et.accounts || et.accounts.length === 0">
                  <td colspan="7" class="py-6 text-center text-slate-400 bg-slate-50/30">
                    Belum ada akun biaya di bawah tipe ini.
                    <button 
                      class="text-indigo-600 font-bold ml-1 hover:underline inline-flex items-center gap-0.5 cursor-pointer"
                      @click="openAddExpenseAccountModal(et.id)"
                    >
                      + Tambah sekarang
                    </button>
                  </td>
                </tr>
                <tr 
                  v-else 
                  v-for="acc in et.accounts" 
                  :key="acc.id" 
                  class="hover:bg-slate-50/60 transition-colors"
                >
                  <td class="py-2.5 px-4 text-center font-mono text-slate-400">{{ acc.sort_order }}</td>
                  <td class="py-2.5 px-4">
                    <span class="font-mono font-bold text-slate-800 bg-slate-100 px-2 py-0.5 rounded border border-slate-200">
                      {{ acc.account_code }}
                    </span>
                  </td>
                  <td class="py-2.5 px-4 font-bold text-slate-900 text-sm">
                    {{ acc.account_name }}
                  </td>
                  <td class="py-2.5 px-4">
                    <span v-if="acc.accurate_account_id" class="font-mono text-indigo-700 bg-indigo-50 px-2 py-0.5 rounded border border-indigo-100">
                      {{ acc.accurate_account_id }}
                    </span>
                    <span v-else class="text-slate-400 italic">-</span>
                  </td>
                  <td class="py-2.5 px-4 text-slate-500 max-w-xs">{{ acc.description || '-' }}</td>
                  <td class="py-2.5 px-4 text-center">
                    <span 
                      class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold"
                      :class="acc.is_active ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-600'"
                    >
                      {{ acc.is_active ? 'Aktif' : 'Non-aktif' }}
                    </span>
                  </td>
                  <td class="py-2.5 px-4 text-center">
                    <div class="flex items-center justify-center gap-1">
                      <button 
                        class="p-1 text-indigo-600 hover:bg-indigo-50 rounded transition-colors cursor-pointer"
                        @click="openEditExpenseAccountModal(acc)"
                        title="Edit Akun"
                      >
                        <Pencil class="w-3.5 h-3.5" />
                      </button>
                      <button 
                        class="p-1 text-rose-600 hover:bg-rose-50 rounded transition-colors cursor-pointer"
                        @click="handleDeleteExpenseAccount(acc.id)"
                        title="Hapus Akun"
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
    </div>
    <!-- END TAB 3: MASTER TIPE & AKUN PENGELUARAN -->

    <!-- Modal Master: Sumber Penerimaan -->
    <div v-if="showIncomeSourceModal" class="modal-overlay" @click.self="showIncomeSourceModal = false">
      <div class="modal-box">
        <div class="modal-header">
          <h2>{{ isEditingIncomeSource ? 'Edit Sumber Penerimaan' : 'Tambah Sumber Penerimaan' }}</h2>
          <button class="btn-close" @click="showIncomeSourceModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="modal-form">
          <div class="form-group">
            <label>Kode Sumber <span class="required">*</span></label>
            <input 
              type="text" 
              v-model="incomeSourceForm.code" 
              placeholder="Cth: PENJUALAN_PABRIK, SETORAN_TUNAI" 
              required 
              class="uppercase font-mono font-bold"
            />
          </div>
          <div class="form-group">
            <label>Nama Sumber Penerimaan <span class="required">*</span></label>
            <input 
              type="text" 
              v-model="incomeSourceForm.name" 
              placeholder="Cth: Penjualan Langsung Pabrik" 
              required 
            />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="form-group">
              <label>Kode Akun COA <span class="optional">(Opsional)</span></label>
              <input 
                type="text" 
                v-model="incomeSourceForm.account_code" 
                placeholder="Cth: 4-40100" 
              />
            </div>
            <div class="form-group">
              <label>Nama Akun COA <span class="optional">(Opsional)</span></label>
              <input 
                type="text" 
                v-model="incomeSourceForm.account_name" 
                placeholder="Cth: Pendapatan Penjualan" 
              />
            </div>
          </div>
          <div class="form-group">
            <label>Deskripsi</label>
            <textarea v-model="incomeSourceForm.description" rows="2" placeholder="Keterangan peruntukan sumber..."></textarea>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="form-group">
              <label>Urutan Tampilan</label>
              <input type="number" v-model="incomeSourceForm.sort_order" min="0" />
            </div>
            <div class="form-group flex flex-col justify-end">
              <label class="flex items-center gap-2 cursor-pointer pb-2">
                <input type="checkbox" v-model="incomeSourceForm.is_active" class="w-4 h-4 text-indigo-600 rounded" />
                <span class="text-xs font-bold text-slate-700">Status Aktif</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showIncomeSourceModal = false">Batal</button>
          <button class="btn-primary" @click="handleSaveIncomeSource" :disabled="!incomeSourceForm.code || !incomeSourceForm.name">
            {{ isEditingIncomeSource ? 'Simpan Perubahan' : 'Tambah Sumber' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Master: Tipe Pengeluaran (Revamped) -->
    <div v-if="showExpenseTypeModal" class="modal-overlay" @click.self="showExpenseTypeModal = false">
      <div class="modal-box overflow-hidden border border-slate-200/80 shadow-2xl rounded-2xl max-w-lg bg-white">
        <!-- Modern Header -->
        <div class="bg-gradient-to-r from-slate-900 via-indigo-950 to-indigo-900 text-white p-5 px-6 flex items-center justify-between border-b border-indigo-500/20">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-indigo-500/20 border border-indigo-400/30 flex items-center justify-center text-indigo-300 shadow-inner">
              <Layers class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-base font-black tracking-tight text-white m-0">
                  {{ isEditingExpenseType ? 'Edit Tipe Pengeluaran' : 'Tambah Tipe Pengeluaran' }}
                </h2>
                <span class="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full"
                  :class="isEditingExpenseType ? 'bg-amber-400/20 text-amber-300 border border-amber-400/30' : 'bg-emerald-400/20 text-emerald-300 border border-emerald-400/30'">
                  {{ isEditingExpenseType ? 'Ubah Master' : 'Baru' }}
                </span>
              </div>
              <p class="text-xs text-indigo-200/80 mt-0.5">Kategori induk pengelompokan akun biaya kas pabrik</p>
            </div>
          </div>
          <button class="w-8 h-8 rounded-lg bg-white/10 hover:bg-white/20 text-slate-300 hover:text-white flex items-center justify-center transition-colors cursor-pointer border border-white/10" @click="showExpenseTypeModal = false">
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Form Body -->
        <div class="p-6 space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="form-group mb-0">
              <label class="text-xs font-bold text-slate-700 flex items-center justify-between">
                <span>Kode Tipe <span class="text-rose-500">*</span></span>
              </label>
              <input 
                type="text" 
                v-model="expenseTypeForm.code" 
                placeholder="Cth: OPERASIONAL" 
                required 
                class="uppercase font-mono font-bold tracking-wider mt-1 text-sm bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
              />
              <span class="text-[11px] text-slate-400 mt-1">Identifikasi unik kode kategori</span>
            </div>

            <div class="form-group mb-0">
              <label class="text-xs font-bold text-slate-700">
                Urutan Tampilan
              </label>
              <input 
                type="number" 
                v-model="expenseTypeForm.sort_order" 
                min="0" 
                class="mt-1 text-sm font-semibold bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
              />
              <span class="text-[11px] text-slate-400 mt-1">Urutan prioritas pada daftar</span>
            </div>
          </div>

          <div class="form-group mb-0">
            <label class="text-xs font-bold text-slate-700">
              Nama Tipe Pengeluaran <span class="text-rose-500">*</span>
            </label>
            <input 
              type="text" 
              v-model="expenseTypeForm.name" 
              placeholder="Cth: Biaya Operasional Pabrik" 
              required 
              class="mt-1 text-sm font-medium bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
            />
          </div>

          <div class="form-group mb-0">
            <label class="text-xs font-bold text-slate-700">Deskripsi / Peruntukan</label>
            <textarea 
              v-model="expenseTypeForm.description" 
              rows="2" 
              placeholder="Keterangan cakupan tipe pengeluaran ini..."
              class="mt-1 text-xs bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
            ></textarea>
          </div>

          <!-- Status Toggle Card -->
          <div class="p-3.5 rounded-xl border border-slate-200 bg-slate-50/80 flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <div class="w-8 h-8 rounded-lg flex items-center justify-center font-bold text-xs"
                :class="expenseTypeForm.is_active ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-200 text-slate-600'">
                <Check v-if="expenseTypeForm.is_active" class="w-4 h-4" />
                <X v-else class="w-4 h-4" />
              </div>
              <div>
                <div class="text-xs font-bold text-slate-800">Status Master Tipe</div>
                <div class="text-[11px] text-slate-500">
                  {{ expenseTypeForm.is_active ? 'Aktif - Tampil saat input kas keluar' : 'Non-aktif - Disembunyikan dari pilihan' }}
                </div>
              </div>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" v-model="expenseTypeForm.is_active" class="sr-only peer">
              <div class="w-11 h-6 bg-slate-300 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-emerald-600"></div>
            </label>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 px-6 bg-slate-50 border-t border-slate-200 flex items-center justify-end gap-2.5">
          <button 
            class="px-4 py-2 text-xs font-bold text-slate-600 hover:text-slate-800 hover:bg-slate-200/60 rounded-xl transition-colors cursor-pointer" 
            @click="showExpenseTypeModal = false"
            :disabled="isSavingExpenseType"
          >
            Batal
          </button>
          <button 
            class="px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-xs font-extrabold rounded-xl shadow-md shadow-indigo-200 transition-all flex items-center gap-2 cursor-pointer" 
            @click="handleSaveExpenseType" 
            :disabled="isSavingExpenseType || !expenseTypeForm.code || !expenseTypeForm.name"
          >
            <RefreshCw v-if="isSavingExpenseType" class="w-3.5 h-3.5 spin" />
            <CheckCircle2 v-else class="w-3.5 h-3.5" />
            {{ isSavingExpenseType ? 'Menyimpan...' : (isEditingExpenseType ? 'Simpan Perubahan' : 'Tambah Tipe') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Master: Akun Pengeluaran (Revamped) -->
    <div v-if="showExpenseAccountModal" class="modal-overlay" @click.self="showExpenseAccountModal = false">
      <div class="modal-box overflow-hidden border border-slate-200/80 shadow-2xl rounded-2xl max-w-lg bg-white">
        <!-- Modern Header -->
        <div class="bg-gradient-to-r from-slate-900 via-indigo-950 to-indigo-900 text-white p-5 px-6 flex items-center justify-between border-b border-indigo-500/20">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-indigo-500/20 border border-indigo-400/30 flex items-center justify-center text-indigo-300 shadow-inner">
              <Tag class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-base font-black tracking-tight text-white m-0">
                  {{ isEditingExpenseAccount ? 'Edit Akun Pengeluaran' : 'Tambah Akun Pengeluaran' }}
                </h2>
                <span class="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full"
                  :class="isEditingExpenseAccount ? 'bg-amber-400/20 text-amber-300 border border-amber-400/30' : 'bg-emerald-400/20 text-emerald-300 border border-emerald-400/30'">
                  {{ isEditingExpenseAccount ? 'Ubah Akun' : 'Akun Baru' }}
                </span>
              </div>
              <p class="text-xs text-indigo-200/80 mt-0.5">Rincian pos pengeluaran & mapping Accurate COA</p>
            </div>
          </div>
          <button class="w-8 h-8 rounded-lg bg-white/10 hover:bg-white/20 text-slate-300 hover:text-white flex items-center justify-center transition-colors cursor-pointer border border-white/10" @click="showExpenseAccountModal = false">
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Form Body -->
        <div class="p-6 space-y-4">
          <!-- Parent Expense Type -->
          <div class="form-group mb-0">
            <label class="text-xs font-bold text-slate-700 flex items-center justify-between">
              <span>Tipe Pengeluaran Induk <span class="text-rose-500">*</span></span>
              <span class="text-[11px] text-indigo-600 font-semibold">Kategori Parent</span>
            </label>
            <select 
              v-model="expenseAccountForm.expense_type_id" 
              required
              class="mt-1 text-sm font-bold text-indigo-950 bg-indigo-50/40 border-indigo-200 focus:bg-white focus:border-indigo-600"
            >
              <option value="" disabled>-- Pilih Tipe Pengeluaran Induk --</option>
              <option v-for="t in expenseTypeList" :key="t.id" :value="t.id">
                {{ t.name }} ({{ t.code }})
              </option>
            </select>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="form-group mb-0">
              <label class="text-xs font-bold text-slate-700">
                Kode Akun Biaya <span class="text-rose-500">*</span>
              </label>
              <input 
                type="text" 
                v-model="expenseAccountForm.account_code" 
                placeholder="Cth: BIAYA_BBM / 6-60100" 
                required 
                class="font-mono font-bold mt-1 text-sm bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
              />
              <span class="text-[11px] text-slate-400 mt-1">Kode unik pos kas keluar</span>
            </div>

            <div class="form-group mb-0">
              <label class="text-xs font-bold text-slate-700 flex items-center justify-between">
                <span>Accurate COA</span>
                <span class="text-[10px] text-slate-400 font-normal">(Opsional)</span>
              </label>
              <input 
                type="text" 
                v-model="expenseAccountForm.accurate_account_id" 
                placeholder="Cth: ACC-6-60100" 
                class="font-mono mt-1 text-sm bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
              />
              <span class="text-[11px] text-slate-400 mt-1">No Akun Pembukuan Accurate</span>
            </div>
          </div>

          <div class="form-group mb-0">
            <label class="text-xs font-bold text-slate-700">
              Nama Akun Biaya <span class="text-rose-500">*</span>
            </label>
            <input 
              type="text" 
              v-model="expenseAccountForm.account_name" 
              placeholder="Cth: Biaya Bahan Bakar Minyak & Tol" 
              required 
              class="mt-1 text-sm font-semibold bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
            />
          </div>

          <div class="form-group mb-0">
            <label class="text-xs font-bold text-slate-700">Deskripsi / Peruntukan Akun</label>
            <textarea 
              v-model="expenseAccountForm.description" 
              rows="2" 
              placeholder="Catatan detail rincian atau peruntukan transaksi..."
              class="mt-1 text-xs bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
            ></textarea>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 items-center">
            <div class="form-group mb-0">
              <label class="text-xs font-bold text-slate-700">Urutan Tampilan</label>
              <input 
                type="number" 
                v-model="expenseAccountForm.sort_order" 
                min="0" 
                class="mt-1 text-sm font-semibold bg-slate-50 border-slate-200 focus:bg-white focus:border-indigo-500"
              />
            </div>

            <!-- Status Toggle Switch -->
            <div class="p-3 rounded-xl border border-slate-200 bg-slate-50/80 flex items-center justify-between sm:mt-5">
              <div>
                <div class="text-xs font-bold text-slate-800">Status Akun</div>
                <div class="text-[10px] text-slate-500">{{ expenseAccountForm.is_active ? 'Aktif' : 'Non-aktif' }}</div>
              </div>
              <label class="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" v-model="expenseAccountForm.is_active" class="sr-only peer">
                <div class="w-9 h-5 bg-slate-300 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-600"></div>
              </label>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 px-6 bg-slate-50 border-t border-slate-200 flex items-center justify-end gap-2.5">
          <button 
            class="px-4 py-2 text-xs font-bold text-slate-600 hover:text-slate-800 hover:bg-slate-200/60 rounded-xl transition-colors cursor-pointer" 
            @click="showExpenseAccountModal = false"
            :disabled="isSavingExpenseAccount"
          >
            Batal
          </button>
          <button 
            class="px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-xs font-extrabold rounded-xl shadow-md shadow-indigo-200 transition-all flex items-center gap-2 cursor-pointer" 
            @click="handleSaveExpenseAccount" 
            :disabled="isSavingExpenseAccount || !expenseAccountForm.expense_type_id || !expenseAccountForm.account_code || !expenseAccountForm.account_name"
          >
            <RefreshCw v-if="isSavingExpenseAccount" class="w-3.5 h-3.5 spin" />
            <CheckCircle2 v-else class="w-3.5 h-3.5" />
            {{ isSavingExpenseAccount ? 'Menyimpan...' : (isEditingExpenseAccount ? 'Simpan Perubahan' : 'Tambah Akun') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Tambah KAS -->
    <div v-if="showAddKasModal" class="modal-overlay" @click.self="showAddKasModal = false">
      <div class="modal-box">
        <div class="modal-header">
          <h2>Tambah KAS Baru</h2>
          <button class="btn-close" @click="showAddKasModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="modal-form">
          <div class="form-group">
            <label>Kode KAS <span class="required">*</span></label>
            <input type="text" v-model="kasForm.kas_code" placeholder="Cth: KAS-01" required />
          </div>
          <div class="form-group">
            <label>Nama KAS <span class="required">*</span></label>
            <input type="text" v-model="kasForm.kas_name" placeholder="Cth: KAS BESAR" required />
          </div>
          <div class="form-group">
            <label>Deskripsi</label>
            <textarea v-model="kasForm.description" rows="2" placeholder="Deskripsi peruntukan KAS"></textarea>
          </div>
          <div class="form-group">
            <label>Nama Bank <span class="optional">(Opsional)</span></label>
            <input type="text" v-model="kasForm.bank_name" placeholder="Cth: BCA" />
          </div>
          <div class="form-group">
            <label>Nomor Rekening <span class="optional">(Opsional)</span></label>
            <input type="text" v-model="kasForm.account_number" placeholder="Nomor rekening" />
          </div>

        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showAddKasModal = false">Batal</button>
          <button class="btn-primary" @click="handleAddKas" :disabled="!kasForm.kas_code || !kasForm.kas_name">Simpan KAS</button>
        </div>
      </div>
    </div>

    <!-- Modal Penerimaan (IN) -->
    <div v-if="showInModal" class="modal-overlay" @click.self="showInModal = false">
      <div class="modal-box">
        <div class="modal-header">
          <h2>Catat Penerimaan KAS (Masuk)</h2>
          <button class="btn-close" @click="showInModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="modal-form">
          <div class="form-group">
            <label>Pilih KAS <span class="required">*</span></label>
            <select v-model="inForm.kas_pabrik_id" required>
              <option value="" disabled>-- Pilih KAS --</option>
              <option v-for="kas in kasList" :key="kas.id" :value="kas.id">{{ kas.kas_name }} ({{ formatRupiah(kas.balance) }})</option>
            </select>
          </div>
          <div class="form-group">
            <label>Tanggal Transaksi <span class="required">*</span></label>
            <input type="date" v-model="inForm.transaction_date" required />
          </div>
          <div class="form-group">
            <label>Sumber Penerimaan <span class="required">*</span></label>
            <select v-model="inForm.income_source_id" @change="onIncomeSourceChange(inForm.income_source_id)" required>
              <option value="" disabled>-- Pilih Sumber Penerimaan --</option>
              <option v-for="src in incomeSourceList" :key="src.id" :value="src.id">
                {{ src.name }} ({{ src.code }})
              </option>
            </select>
          </div>
          <div class="form-group" v-if="inForm.income_source === 'BRANCH_RECEIVABLE_PAYMENT' || inForm.income_source === 'TALANGAN_CABANG'">
            <label>Cabang yang Melunasi Talangan <span class="required">*</span></label>
            <select v-model="inForm.target_branch_id" required>
              <option value="" disabled>-- Pilih Cabang --</option>
              <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>Nominal (Rp) <span class="required">*</span></label>
            <input type="number" v-model="inForm.amount" min="1" required />
          </div>
          <div class="form-group">
            <label>Keterangan</label>
            <textarea v-model="inForm.description" rows="2" placeholder="Catatan tambahan..."></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showInModal = false">Batal</button>
          <button class="btn-success" @click="handleTransaction('IN')" :disabled="!inForm.kas_pabrik_id || !inForm.amount || !inForm.income_source_id">
            Simpan Penerimaan
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Pengeluaran (OUT) -->
    <div v-if="showOutModal" class="modal-overlay" @click.self="showOutModal = false">
      <div class="modal-box max-w-lg">
        <div class="modal-header">
          <h2>Catat Pengeluaran KAS (Keluar)</h2>
          <button class="btn-close" @click="showOutModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="modal-form space-y-4">
          <div class="form-group">
            <label>Pilih KAS <span class="required">*</span></label>
            <select v-model="outForm.kas_pabrik_id" required>
              <option value="" disabled>-- Pilih KAS --</option>
              <option v-for="kas in kasList" :key="kas.id" :value="kas.id">{{ kas.kas_name }} ({{ formatRupiah(kas.balance) }})</option>
            </select>
          </div>

          <!-- Pilihan Kategori Induk / Tipe Pengeluaran (Dinamis) -->
          <div class="form-group">
            <div class="flex items-center justify-between mb-1.5">
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider">
                Kategori Induk / Tipe Pengeluaran <span class="required">*</span>
              </label>
              <span v-if="expenseTypeList.length > 0" class="text-[11px] text-slate-500">
                {{ expenseTypeList.length }} Kategori Tersedia
              </span>
            </div>
            <select 
              v-model="outForm.expense_type_id" 
              @change="onExpenseTypeChange(outForm.expense_type_id)" 
              required
              class="w-full font-medium"
            >
              <option value="" disabled>-- Pilih Kategori Induk Pengeluaran --</option>
              <option v-for="t in expenseTypeList" :key="t.id" :value="t.id">
                {{ t.name }} ({{ t.code }})
              </option>
            </select>
          </div>

          <!-- Smart Linkage Entity Sections (Armada Kendaraan / Cabang / Kasbon) -->
          <div class="space-y-3 p-3 bg-slate-50 rounded-xl border border-slate-200">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-slate-700 uppercase tracking-wider flex items-center gap-1.5">
                <Tag class="w-3.5 h-3.5 text-slate-500" />
                Penghubung Entitas Khusus (Opsional)
              </span>
              <div class="flex items-center gap-2">
                <button 
                  type="button" 
                  @click="showVehicleField = !showVehicleField"
                  class="text-[11px] px-2 py-0.5 rounded-md font-semibold transition-all border"
                  :class="showVehicleField || isVehicleSuggested ? 'bg-blue-100 text-blue-700 border-blue-300 shadow-xs' : 'bg-white text-slate-600 border-slate-300 hover:bg-slate-100'"
                >
                  <Truck class="w-3 h-3 inline-block mr-0.5" /> Armada
                </button>
                <button 
                  type="button" 
                  @click="showBranchField = !showBranchField"
                  class="text-[11px] px-2 py-0.5 rounded-md font-semibold transition-all border"
                  :class="showBranchField || isBranchSuggested ? 'bg-amber-100 text-amber-700 border-amber-300 shadow-xs' : 'bg-white text-slate-600 border-slate-300 hover:bg-slate-100'"
                >
                  <Building class="w-3 h-3 inline-block mr-0.5" /> Cabang
                </button>
                <button 
                  type="button" 
                  @click="showAdvanceField = !showAdvanceField"
                  class="text-[11px] px-2 py-0.5 rounded-md font-semibold transition-all border"
                  :class="showAdvanceField || isAdvanceSuggested ? 'bg-purple-100 text-purple-700 border-purple-300 shadow-xs' : 'bg-white text-slate-600 border-slate-300 hover:bg-slate-100'"
                >
                  <User class="w-3 h-3 inline-block mr-0.5" /> Kasbon/Dinas
                </button>
              </div>
            </div>

            <!-- Dynamic Field: Armada Kendaraan -->
            <div v-if="showVehicleField || isVehicleSuggested" class="bg-blue-50/70 p-3 rounded-lg border border-blue-200">
              <div class="flex items-center justify-between mb-1">
                <label class="block text-xs font-bold text-blue-900 flex items-center gap-1.5">
                  <Truck class="w-3.5 h-3.5 text-blue-600" />
                  Pilih Armada Kendaraan (Plat Nomor) <span class="required">*</span>
                </label>
                <button 
                  v-if="vehicles.length === 0" 
                  type="button" 
                  @click="fetchVehiclesAndEmployees" 
                  class="text-[10px] text-blue-600 hover:underline font-semibold"
                >
                  Muat Ulang Armada
                </button>
              </div>
              <select v-model="outForm.vehicle_id" :required="showVehicleField || isVehicleSuggested" class="bg-white">
                <option value="" disabled>{{ vehicles.length === 0 ? '-- Tidak ada armada ditemukan (klik Muat Ulang) --' : '-- Pilih Armada Kendaraan --' }}</option>
                <option v-for="v in vehicles" :key="v.id" :value="v.id">
                  {{ v.name }} - {{ v.license_plate }}{{ v.branch?.name ? ` (${v.branch.name})` : '' }}
                </option>
              </select>
            </div>

            <!-- Dynamic Field: Piutang Cabang -->
            <div v-if="showBranchField || isBranchSuggested" class="bg-amber-50/70 p-3 rounded-lg border border-amber-200">
              <label class="block text-xs font-bold text-amber-900 mb-1 flex items-center gap-1.5">
                <Building class="w-3.5 h-3.5 text-amber-600" />
                Pilih Cabang Debitur (Talangan Piutang Cabang) <span class="required">*</span>
              </label>
              <select v-model="outForm.target_branch_id" :required="showBranchField || isBranchSuggested" class="bg-white">
                <option value="" disabled>-- Pilih Cabang --</option>
                <option v-for="b in branches" :key="b.id" :value="b.id">
                  {{ b.name }}
                </option>
              </select>
              <p class="text-[11px] text-amber-700 mt-1">Pengeluaran ini akan otomatis dicatat sebagai piutang operasional cabang terkait.</p>
            </div>

            <!-- Dynamic Field: Kasbon Uang Muka Dinas (Audit/Karyawan) -->
            <div v-if="showAdvanceField || isAdvanceSuggested" class="bg-purple-50/70 p-3 rounded-lg border border-purple-200">
              <label class="block text-xs font-bold text-purple-900 mb-1 flex items-center gap-1.5">
                <User class="w-3.5 h-3.5 text-purple-600" />
                Pilih Karyawan Penerima Uang Muka (Kasbon Dinas) <span class="required">*</span>
              </label>
              <select v-model="outForm.employee_id" :required="showAdvanceField || isAdvanceSuggested" class="bg-white">
                <option value="" disabled>-- Pilih Karyawan --</option>
                <option v-for="e in employees" :key="e.id" :value="e.id">
                  {{ e.first_name }} {{ e.last_name || '' }} {{ e.job_position ? `(${e.job_position.title})` : '' }}
                </option>
              </select>
              <p class="text-[11px] text-purple-700 mt-1">Status uang muka akan otomatis OUTSTANDING hingga diselesaikan saat LPJ diterima.</p>
            </div>
          </div>

          <div class="form-group">
            <label>Tanggal Transaksi <span class="required">*</span></label>
            <input type="date" v-model="outForm.transaction_date" required />
          </div>

          <!-- Searchable Expense Account Dropdown -->
          <div class="form-group relative" ref="expenseAccountDropdownRef">
            <label>Akun Biaya / Pengeluaran <span class="required">*</span></label>
            <div class="searchable-select-container">
              <button
                type="button"
                @click="showExpenseAccountDropdown = !showExpenseAccountDropdown"
                class="searchable-select-trigger"
                :class="{ 'has-value': outForm.expense_account_id, 'is-open': showExpenseAccountDropdown }"
              >
                <span class="truncate">
                  {{ getSelectedExpenseAccountLabel || '-- Pilih Akun Biaya / Pengeluaran --' }}
                </span>
                <ChevronDown class="w-4 h-4 chevron-icon" :class="{ 'rotate-180': showExpenseAccountDropdown }" />
              </button>

              <div v-if="showExpenseAccountDropdown" class="searchable-dropdown-popover">
                <div class="search-input-wrapper">
                  <Search class="w-4 h-4 search-icon" />
                  <input
                    type="text"
                    v-model="expenseAccountSearch"
                    placeholder="Cari akun biaya atau kode..."
                    class="search-input"
                    autofocus
                    @click.stop
                  />
                  <button v-if="expenseAccountSearch" type="button" @click.stop="expenseAccountSearch = ''" class="clear-search-btn">
                    <X class="w-3.5 h-3.5" />
                  </button>
                </div>
                <div class="options-list custom-scrollbar max-h-60 overflow-y-auto">
                  <div v-if="filteredExpenseAccounts.length === 0" class="no-options">
                    Akun biaya tidak ditemukan
                  </div>
                  <button
                    v-for="acc in filteredExpenseAccounts"
                    :key="acc.id"
                    type="button"
                    @click="selectExpenseAccount(acc)"
                    class="option-item flex flex-col items-start p-2 text-left hover:bg-indigo-50 border-b border-slate-100 last:border-none cursor-pointer w-full"
                    :class="{ 'selected bg-indigo-50': outForm.expense_account_id === acc.id }"
                  >
                    <div class="flex items-center justify-between w-full">
                      <span class="font-bold text-slate-800 text-xs">{{ acc.account_name }}</span>
                      <span class="text-[10px] px-1.5 py-0.5 rounded bg-slate-100 text-slate-600 font-mono">{{ acc.account_code }}</span>
                    </div>
                    <div class="text-[11px] text-slate-500 mt-0.5 flex items-center gap-1">
                      <span class="text-indigo-600 font-medium">[{{ acc.expense_type?.name || 'Tipe' }}]</span>
                      <span v-if="acc.description">• {{ acc.description }}</span>
                    </div>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="form-group">
            <label>Nominal Pengeluaran (Rp) <span class="required">*</span></label>
            <input type="number" v-model="outForm.amount" min="1" required />
          </div>

          <div class="form-group">
            <label>Keterangan / Keperluan</label>
            <textarea v-model="outForm.description" rows="2" placeholder="Contoh: Beli pulpen untuk Solo / Kasbon dinas Mas Bagus ke Solo..."></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showOutModal = false">Batal</button>
          <button class="btn-danger" @click="handleTransaction('OUT')" :disabled="!outForm.kas_pabrik_id || !outForm.amount || !outForm.expense_account_id">
            Simpan Pengeluaran
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Selesaikan Kasbon Dinas (Settlement) -->
    <div v-if="showSettleModal && selectedAdvanceTx" class="modal-overlay" @click.self="showSettleModal = false">
      <div class="modal-box max-w-md">
        <div class="modal-header">
          <div class="flex items-center gap-2">
            <Coins class="w-5 h-5 text-purple-600" />
            <h2 class="text-base font-bold text-slate-900">Selesaikan Kasbon Dinas</h2>
          </div>
          <button class="btn-close" @click="showSettleModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="modal-form space-y-4">
          <!-- Info Kasbon Awal -->
          <div class="bg-purple-50/70 p-3.5 rounded-2xl border border-purple-200 space-y-1.5 text-xs text-purple-950">
            <div class="flex justify-between">
              <span class="text-purple-700">Karyawan:</span>
              <strong class="font-bold">{{ selectedAdvanceTx.employee ? selectedAdvanceTx.employee.first_name : 'Karyawan' }}</strong>
            </div>
            <div class="flex justify-between">
              <span class="text-purple-700">Uang Muka Awal:</span>
              <strong class="font-bold text-purple-900">{{ formatRupiah(selectedAdvanceTx.amount) }}</strong>
            </div>
            <div class="flex justify-between">
              <span class="text-purple-700">Tanggal Kasbon:</span>
              <span>{{ formatDate(selectedAdvanceTx.transaction_date) }}</span>
            </div>
          </div>

          <div class="form-group">
            <label class="block text-xs font-bold text-slate-700 mb-1">
              Pengeluaran Riil Sesuai Nota/Struk (Rp) <span class="required">*</span>
            </label>
            <input 
              type="number" 
              v-model.number="settleForm.actual_expense_amount" 
              min="0" 
              required 
              class="w-full font-bold text-slate-900"
            />
          </div>

          <!-- Kalkulasi Otomatis Sisa / Kurang -->
          <div 
            v-if="settleForm.actual_expense_amount < selectedAdvanceTx.amount" 
            class="bg-emerald-50 p-3 rounded-xl border border-emerald-200 text-xs text-emerald-900 flex items-center justify-between"
          >
            <div>
              <span class="font-bold text-emerald-800">Sisa Uang Kembali ke KAS:</span>
              <p class="text-[10px] text-emerald-700 mt-0.5">Saldo KAS Pabrik akan otomatis bertambah kembali realtime.</p>
            </div>
            <span class="font-black text-sm text-emerald-700">
              {{ formatRupiah(selectedAdvanceTx.amount - settleForm.actual_expense_amount) }}
            </span>
          </div>

          <div 
            v-else-if="settleForm.actual_expense_amount > selectedAdvanceTx.amount" 
            class="bg-amber-50 p-3 rounded-xl border border-amber-200 text-xs text-amber-900 flex items-center justify-between"
          >
            <div>
              <span class="font-bold text-amber-800">Kekurangan Dana (Klaim Tambahan):</span>
              <p class="text-[10px] text-amber-700 mt-0.5">Pengeluaran riil melebihi kasbon awal.</p>
            </div>
            <span class="font-black text-sm text-amber-700">
              {{ formatRupiah(settleForm.actual_expense_amount - selectedAdvanceTx.amount) }}
            </span>
          </div>

          <div 
            v-else 
            class="bg-blue-50 p-3 rounded-xl border border-blue-200 text-xs text-blue-900 flex items-center justify-between"
          >
            <span class="font-bold">Pengeluaran pas sesuai kasbon awal.</span>
            <span class="font-bold text-blue-700">Rp 0</span>
          </div>

          <!-- Bukti Struk -->
          <div class="form-group">
            <label class="block text-xs font-bold text-slate-700 mb-1">Upload Bukti Struk/Nota Realisasi</label>
            <input type="file" @change="handleSettleFileChange" accept="image/*,application/pdf" class="text-xs" />
          </div>

          <div class="form-group">
            <label class="block text-xs font-bold text-slate-700 mb-1">Catatan Realisasi</label>
            <textarea v-model="settleForm.notes" rows="2" placeholder="Contoh: Realisasi perjalanan dinas audit Solo..."></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showSettleModal = false" :disabled="isSettling">Batal</button>
          <button class="btn-indigo bg-purple-600 hover:bg-purple-700" @click="handleSettleAdvance" :disabled="isSettling || settleForm.actual_expense_amount < 0">
            {{ isSettling ? 'Memproses...' : 'Simpan & Selesaikan' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Transfer -->
    <div v-if="showTransferModal" class="modal-overlay" @click.self="showTransferModal = false">
      <div class="modal-box">
        <div class="modal-header">
          <h2>Transfer Antar KAS</h2>
          <button class="btn-close" @click="showTransferModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="modal-form">
          <div class="form-group">
            <label>Tanggal Transfer <span class="required">*</span></label>
            <input type="date" v-model="transferForm.transaction_date" required />
          </div>
          <div class="form-group">
            <label>KAS Asal (Pengirim) <span class="required">*</span></label>
            <select v-model="transferForm.from_kas_id" required>
              <option value="" disabled>-- Pilih KAS Asal --</option>
              <option v-for="kas in kasList" :key="kas.id" :value="kas.id">{{ kas.kas_name }} ({{ formatRupiah(kas.balance) }})</option>
            </select>
          </div>
          <div class="form-group">
            <label>KAS Tujuan (Penerima) <span class="required">*</span></label>
            <select v-model="transferForm.to_kas_id" required>
              <option value="" disabled>-- Pilih KAS Tujuan --</option>
              <option v-for="kas in kasList.filter(k => k.id !== transferForm.from_kas_id)" :key="kas.id" :value="kas.id">
                {{ kas.kas_name }} ({{ formatRupiah(kas.balance) }})
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>Nominal Transfer (Rp) <span class="required">*</span></label>
            <input type="number" v-model="transferForm.amount" min="1" required />
          </div>
          <div class="form-group">
            <label>Keterangan</label>
            <textarea v-model="transferForm.description" rows="2" placeholder="Catatan transfer..."></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showTransferModal = false">Batal</button>
          <button class="btn-indigo" @click="handleTransfer" :disabled="!transferForm.from_kas_id || !transferForm.to_kas_id || !transferForm.amount">
            Proses Transfer
          </button>
        </div>
      </div>
    </div>
    <!-- Modal Edit KAS -->
    <div v-if="showEditKasModal" class="modal-overlay" @click.self="showEditKasModal = false">
      <div class="modal-box">
        <div class="modal-header">
          <h2>Edit KAS</h2>
          <button class="btn-close" @click="showEditKasModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="modal-form">
          <div class="form-group">
            <label>Kode KAS <span class="required">*</span></label>
            <input type="text" v-model="editKasForm.kas_code" required />
          </div>
          <div class="form-group">
            <label>Nama KAS <span class="required">*</span></label>
            <input type="text" v-model="editKasForm.kas_name" required />
          </div>
          <div class="form-group">
            <label>Deskripsi</label>
            <textarea v-model="editKasForm.description" rows="2"></textarea>
          </div>
          <div class="form-group">
            <label>Nama Bank</label>
            <input type="text" v-model="editKasForm.bank_name" />
          </div>
          <div class="form-group">
            <label>Nomor Rekening</label>
            <input type="text" v-model="editKasForm.account_number" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showEditKasModal = false">Batal</button>
          <button class="btn-primary" @click="handleEditKas">Simpan Perubahan</button>
        </div>
      </div>
    </div>

    <!-- Modal Edit Transaksi -->
    <div v-if="showEditTxModal" class="modal-overlay" @click.self="showEditTxModal = false">
      <div class="modal-box">
        <div class="modal-header">
          <h2>Edit Transaksi</h2>
          <button class="btn-close" @click="showEditTxModal = false">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="modal-form">
          <div class="form-group">
            <label>Tanggal Transaksi <span class="required">*</span></label>
            <input type="date" v-model="editTxForm.transaction_date" required />
          </div>
          <div class="form-group" v-if="editTxForm.transaction_type === 'IN'">
            <label>Sumber Penerimaan <span class="required">*</span></label>
            <select v-model="editTxForm.income_source_id" @change="onEditIncomeSourceChange(editTxForm.income_source_id)" required>
              <option value="" disabled>-- Pilih Sumber Penerimaan --</option>
              <option v-for="src in incomeSourceList" :key="src.id" :value="src.id">
                {{ src.name }} ({{ src.code }})
              </option>
            </select>
          </div>
          <div class="form-group" v-if="editTxForm.transaction_type === 'OUT'">
            <label>Kategori Induk / Tipe Pengeluaran <span class="required">*</span></label>
            <select v-model="editTxForm.expense_type_id" @change="onEditExpenseTypeChange(editTxForm.expense_type_id)" required>
              <option value="" disabled>-- Pilih Kategori Induk --</option>
              <option v-for="t in expenseTypeList" :key="t.id" :value="t.id">
                {{ t.name }} ({{ t.code }})
              </option>
            </select>
          </div>
          <div class="form-group relative" v-if="editTxForm.transaction_type === 'OUT'" ref="editExpenseAccountDropdownRef">
            <label>Akun Pengeluaran <span class="required">*</span></label>
            <div class="searchable-select-container">
              <button
                type="button"
                @click="showEditExpenseAccountDropdown = !showEditExpenseAccountDropdown"
                class="searchable-select-trigger"
                :class="{ 'has-value': editTxForm.expense_account_id, 'is-open': showEditExpenseAccountDropdown }"
              >
                <span class="truncate">
                  {{ getSelectedEditExpenseAccountLabel || '-- Pilih Akun Pengeluaran --' }}
                </span>
                <ChevronDown class="w-4 h-4 chevron-icon" :class="{ 'rotate-180': showEditExpenseAccountDropdown }" />
              </button>

              <div v-if="showEditExpenseAccountDropdown" class="searchable-dropdown-popover">
                <div class="search-input-wrapper">
                  <Search class="w-4 h-4 search-icon" />
                  <input
                    type="text"
                    v-model="editExpenseAccountSearch"
                    placeholder="Cari akun biaya atau kode..."
                    class="search-input"
                    autofocus
                    @click.stop
                  />
                  <button v-if="editExpenseAccountSearch" type="button" @click.stop="editExpenseAccountSearch = ''" class="clear-search-btn">
                    <X class="w-3.5 h-3.5" />
                  </button>
                </div>
                <div class="options-list custom-scrollbar max-h-60 overflow-y-auto">
                  <div v-if="filteredEditExpenseAccounts.length === 0" class="no-options">
                    Akun biaya tidak ditemukan
                  </div>
                  <button
                    v-for="acc in filteredEditExpenseAccounts"
                    :key="acc.id"
                    type="button"
                    @click="selectEditExpenseAccount(acc)"
                    class="option-item flex flex-col items-start p-2 text-left hover:bg-indigo-50 border-b border-slate-100 last:border-none cursor-pointer w-full"
                    :class="{ 'selected bg-indigo-50': editTxForm.expense_account_id === acc.id }"
                  >
                    <div class="flex items-center justify-between w-full">
                      <span class="font-bold text-slate-800 text-xs">{{ acc.account_name }}</span>
                      <span class="text-[10px] px-1.5 py-0.5 rounded bg-slate-100 text-slate-600 font-mono">{{ acc.account_code }}</span>
                    </div>
                    <div class="text-[11px] text-slate-500 mt-0.5 flex items-center gap-1">
                      <span class="text-indigo-600 font-medium">[{{ acc.expense_type?.name || 'Tipe' }}]</span>
                      <span v-if="acc.description">• {{ acc.description }}</span>
                    </div>
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div class="form-group">
            <label>Nominal (Rp) <span class="required">*</span></label>
            <input type="number" v-model="editTxForm.amount" min="1" required />
          </div>
          <div class="form-group">
            <label>Keterangan</label>
            <textarea v-model="editTxForm.description" rows="2"></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showEditTxModal = false">Batal</button>
          <button class="btn-primary" @click="handleEditTx">Simpan Perubahan</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-container {
  padding: 24px 32px;
  max-width: 1400px;
  margin: 0 auto;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}
.header-icon-box {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #e0e7ff, #c7d2fe);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #4f46e5;
}
.header-icon {
  width: 24px;
  height: 24px;
}
.page-title {
  font-size: 24px;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 4px;
}
.page-subtitle {
  font-size: 14px;
  color: #64748b;
  margin: 0;
}
.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.branch-select {
  padding: 10px 16px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
  background: white;
  outline: none;
}
.branch-select:focus {
  border-color: #6366f1;
}

/* KAS Tabs */
.kas-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding: 16px 0;
  border-bottom: 2px solid #f1f5f9;
  margin-bottom: 24px;
}
.kas-tab {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 10px 16px;
  border-radius: 12px;
  border: 2px solid transparent;
  background: #f8fafc;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 100px;
  font-weight: 600;
  font-size: 13px;
  color: #64748b;
}
.kas-tab:hover { background: #f1f5f9; color: #334155; }
.kas-tab.active-all   { background: #ede9fe; border-color: #6366f1; color: #4f46e5; }
.kas-tab.active-csr   { background: #f3e8ff; border-color: #9333ea; color: #7e22ce; }
.kas-tab.active-rosok { background: #fef9c3; border-color: #ca8a04; color: #92400e; }
.kas-tab.active-besar { background: #dcfce7; border-color: #16a34a; color: #15803d; }
.kas-tab.active-bca   { background: #dbeafe; border-color: #2563eb; color: #1d4ed8; }
.kas-tab.active-petty { background: #fce7f3; border-color: #db2777; color: #be185d; }
.tab-balance {
  font-size: 11px;
  font-weight: 500;
  opacity: 0.75;
}

/* KAS Hero Card */
.kas-hero-card {
  border-radius: 20px;
  padding: 32px;
  color: white;
  margin-bottom: 24px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.1);
  position: relative;
  overflow: hidden;
}
.hero-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.hero-badge {
  font-size: 11px;
  padding: 4px 10px;
  background: rgba(255,255,255,0.25);
  border-radius: 12px;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.hero-name {
  font-size: 24px;
  font-weight: 800;
  margin: 0 0 8px;
}
.hero-balance {
  font-size: 42px;
  font-weight: 800;
  margin-bottom: 12px;
}
.hero-desc {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 16px;
  max-width: 600px;
}
.hero-bank {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  background: rgba(0,0,0,0.15);
  padding: 8px 16px;
  border-radius: 10px;
  width: max-content;
  margin-bottom: 24px;
}
.hero-actions {
  display: flex;
  gap: 12px;
}

/* KAS Grid */
.kas-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}
.kas-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.03);
  padding: 20px;
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  color: white;
}
.kas-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.08);
}
.kas-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.kas-name {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.kas-code {
  font-size: 11px;
  padding: 3px 8px;
  background: rgba(255,255,255,0.25);
  border-radius: 12px;
  font-weight: 600;
}
.kas-card-body {
  flex: 1;
}
.balance-label {
  font-size: 12px;
  opacity: 0.9;
  margin-bottom: 4px;
}
.balance-value {
  font-size: 24px;
  font-weight: 800;
  margin-bottom: 12px;
}
.bank-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  background: rgba(0,0,0,0.15);
  padding: 6px 10px;
  border-radius: 8px;
  width: max-content;
}
.kas-card-footer {
  display: flex;
  gap: 8px;
  margin-top: 20px;
}
.btn-kas-action {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  border: none;
  cursor: pointer;
  background: rgba(255,255,255,0.9);
  color: #334155;
  transition: all 0.2s;
}
.btn-kas-action.in { color: #16a34a; }
.btn-kas-action.in:hover { background: white; transform: scale(1.02); }
.btn-kas-action.out { color: #ef4444; }
.btn-kas-action.out:hover { background: white; transform: scale(1.02); }
.btn-kas-action.transfer { color: #4f46e5; }
.btn-kas-action.transfer:hover { background: white; transform: scale(1.02); }
.btn-kas-action.edit { color: #2563eb; }
.btn-kas-action.edit:hover { background: white; transform: scale(1.02); }
.btn-kas-action.delete { color: #dc2626; }
.btn-kas-action.delete:hover { background: white; transform: scale(1.02); }

/* Color classes */
.kas-csr { background: linear-gradient(135deg, #a855f7, #7e22ce); }
.kas-rosok { background: linear-gradient(135deg, #f59e0b, #d97706); }
.kas-besar { background: linear-gradient(135deg, #10b981, #059669); }
.kas-bca-linda { background: linear-gradient(135deg, #3b82f6, #1d4ed8); }
.kas-petty-cash { background: linear-gradient(135deg, #f43f5e, #be123c); }
.kas-default { background: linear-gradient(135deg, #64748b, #475569); }

/* Buttons */
.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  color: white;
  border: none;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s, transform 0.15s;
}
.btn-primary:hover { opacity: 0.92; transform: translateY(-1px); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; transform: none; }

.btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: white;
  color: #64748b;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn-secondary:hover { background: #f8fafc; }

.btn-success {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #10b981;
  color: white;
  border: none;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn-success:hover { background: #059669; }
.btn-success:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-danger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn-danger:hover { background: #dc2626; }
.btn-danger:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-indigo {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 10px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.btn-indigo:hover { background: #4338ca; }
.btn-indigo:disabled { opacity: 0.6; cursor: not-allowed; }

/* Sections */
.content-section {
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.03);
  overflow: hidden;
}
.section-toolbar {
  padding: 20px;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.action-buttons {
  display: flex;
  gap: 12px;
}

/* Filters */
.filter-bar {
  padding: 16px 20px;
  background: #f8fafc;
  border-bottom: 1px solid #f1f5f9;
}
.filter-group {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
.filter-input {
  padding: 10px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #1e293b;
  outline: none;
  background: white;
}
.filter-input:focus { border-color: #6366f1; }

/* Tables */
.table-container {
  overflow-x: auto;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
}
.data-table th, .data-table td {
  padding: 14px 20px;
  text-align: left;
  border-bottom: 1px solid #f1f5f9;
  font-size: 14px;
}
.data-table th {
  font-weight: 600;
  color: #64748b;
  background: #f8fafc;
  white-space: nowrap;
}
.data-table tbody tr:hover {
  background: #f8fafc;
}

/* Badges */
.badge {
  display: inline-flex;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}
.badge-in { background: #dcfce7; color: #16a34a; }
.badge-out { background: #fee2e2; color: #ef4444; }
.badge-tf-in { background: #e0e7ff; color: #4f46e5; }
.badge-tf-out { background: #ffedd5; color: #ea580c; }
.badge-default { background: #f1f5f9; color: #64748b; }

/* Utility */
.text-right { text-align: right !important; }
.text-center { text-align: center !important; }
.py-8 { padding-top: 32px !important; padding-bottom: 32px !important; }
.font-medium { font-weight: 500; }
.text-green-600 { color: #16a34a; }
.text-red-600 { color: #dc2626; }
.text-gray-900 { color: #111827; }
.text-gray-500 { color: #6b7280; }
.text-xs { font-size: 12px; }
.text-sm { font-size: 14px; }
.mx-auto { margin-left: auto; margin-right: auto; }
.spin { animation: spin 0.8s linear infinite; }
@keyframes spin { 100% { transform: rotate(360deg); } }

.btn-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1.5px solid #e2e8f0;
  background: white;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-delete { color: #ef4444; }
.btn-delete:hover { background: #fef2f2; border-color: #ef4444; }

/* Pagination */
.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #f8fafc;
  border-top: 1px solid #f1f5f9;
  flex-wrap: wrap;
  gap: 16px;
}
.pagination-info {
  display: flex;
  align-items: center;
  gap: 20px;
  font-size: 13px;
  color: #64748b;
  flex-wrap: wrap;
}
.limit-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #475569;
}
.limit-selector select {
  padding: 6px 10px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  font-size: 13px;
  color: #1e293b;
  font-weight: 600;
  outline: none;
  cursor: pointer;
}
.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}
.btn-page {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1.5px solid #e2e8f0;
  background: white;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-page:hover:not(:disabled) {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #0f172a;
}
.btn-page:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.page-numbers {
  display: flex;
  align-items: center;
  gap: 4px;
}
.btn-page-num {
  min-width: 32px;
  height: 32px;
  padding: 0 6px;
  border: 1.5px solid transparent;
  background: transparent;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  transition: all 0.15s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.btn-page-num:hover {
  background: #f1f5f9;
  color: #1e293b;
}
.btn-page-num.active {
  background: #4f46e5;
  color: white;
  font-weight: 700;
}

/* Modals */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
  backdrop-filter: blur(4px);
}
.modal-box {
  background: white;
  border-radius: 20px;
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
}
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}
.modal-header h2 { font-size: 17px; font-weight: 700; color: #1e293b; margin: 0; }
.btn-close {
  width: 32px;
  height: 32px;
  border: none;
  background: #f1f5f9;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #64748b;
}
.btn-close:hover { background: #e2e8f0; }

.modal-form { padding: 20px 24px 24px; }
.form-group {
  margin-bottom: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-group label {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}
.required { color: #ef4444; }
.optional { color: #94a3b8; font-weight: 400; font-size: 12px; }
.form-group select,
.form-group input,
.form-group textarea {
  width: 100%;
  padding: 12px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #1e293b;
  transition: all 0.2s;
  background: #f8fafc;
  outline: none;
  font-family: inherit;
  box-sizing: border-box;
}
.form-group select:focus,
.form-group input:focus,
.form-group textarea:focus {
  border-color: #6366f1;
  background: white;
}

.modal-footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #f1f5f9;
  padding: 0 24px 24px;
}

/* Searchable Select */
.searchable-select-container {
  position: relative;
  width: 100%;
}

.searchable-select-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border: 1.5px solid #e2e8f0;
  border-radius: 10px;
  font-size: 14px;
  color: #94a3b8;
  background: #f8fafc;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
  outline: none;
}

.searchable-select-trigger.has-value {
  color: #1e293b;
  font-weight: 500;
}

.searchable-select-trigger:focus,
.searchable-select-trigger.is-open {
  border-color: #6366f1;
  background: white;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.chevron-icon {
  flex-shrink: 0;
  transition: transform 0.2s ease;
}

.searchable-dropdown-popover {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: white;
  border: 1.5px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  z-index: 50;
  overflow: hidden;
  animation: fadeInDown 0.15s ease-out;
}

@keyframes fadeInDown {
  from {
    opacity: 0;
    transform: translateY(-6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.search-input-wrapper {
  position: relative;
  padding: 8px 10px;
  border-bottom: 1px solid #f1f5f9;
  background: #f8fafc;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 20px;
  color: #94a3b8;
  pointer-events: none;
}

.search-input {
  width: 100% !important;
  padding: 8px 30px 8px 32px !important;
  font-size: 13px !important;
  border: 1px solid #cbd5e1 !important;
  border-radius: 8px !important;
  background: white !important;
}

.search-input:focus {
  border-color: #6366f1 !important;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15) !important;
}

.clear-search-btn {
  position: absolute;
  right: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 2px;
}

.clear-search-btn:hover {
  color: #ef4444;
}

.options-list {
  max-height: 220px;
  overflow-y: auto;
  padding: 4px;
}

.no-options {
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: #94a3b8;
  font-weight: 500;
}

.option-item {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 12px;
  font-size: 13px;
  color: #334151;
  background: transparent;
  border: none;
  border-radius: 8px;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s;
}

.option-item:hover {
  background: #f1f5f9;
  color: #0f172a;
}

.option-item.selected {
  background: #eef2ff;
  color: #4338ca;
  font-weight: 600;
}

.check-icon {
  color: #4f46e5;
  flex-shrink: 0;
}
</style>
