<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useSalesOrderStore, type SalesOrder, type SOStatus } from '@/stores/salesOrder'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { getDistance } from '@/utils/geolocation'
import {
  ClipboardList,
  CheckCircle2,
  XCircle,
  FileText,
  Plus,
  X,
  Trash2,
  ChevronDown,
  Search,
  RefreshCw,
  ArrowRight,
  SlidersHorizontal,
  MapPin,
  MapPinOff,
  AlertTriangle,
  RotateCcw,
  Navigation,
  Clock,
  Banknote,
  Layers,
  Sparkles,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card/index'
import SOAdjustModal from '@/components/sales/SOAdjustModal.vue'
import SOWaitingStockModal from '@/components/sales/SOWaitingStockModal.vue'
import ProductSearchSelect from '@/components/sales/ProductSearchSelect.vue'
import OrderLocationMapModal from '@/components/sales/OrderLocationMapModal.vue'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'
import RefreshDataButton from '@/components/ui/RefreshDataButton.vue'
import SearchSelect, { type SelectOption } from '@/components/SearchSelect.vue'
import { useDataRefresh } from '@/composables/useDataRefresh'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const soStore = useSalesOrderStore()
const masterStore = useMasterDataStore()
const factoryStore = useFactoryStore()
const authStore = useAuthStore()

const selectedStatus = ref<SOStatus | 'ALL' | 'DELETED'>('ALL')
const selectedCompanyId = ref('ALL')
const selectedEmployeeId = ref('ALL')
const selectedDate = ref('')
const selectedMonth = ref('ALL')
const selectedYear = ref('ALL')
const searchQuery = ref('')

const currentPage = ref(1)
const itemsPerPage = ref(10) // 0 for all

const salesEmployees = computed(() => {
  return masterStore.employees.filter((emp: any) => {
    // 0. Filter by active status
    const isActive = emp.user?.is_active !== false && emp.employment_status !== 'RESIGNED' && emp.employment_status !== 'TERMINATED'

    // 1. Filter by branch (respect global branch filter in header)
    const branchId = masterStore.selectedBranchId
    const empBranchId = emp.operating_branch_id || emp.branch_id
    const belongsToBranch = !branchId || branchId === 'ALL_BRANCHES' || empBranchId === branchId

    // 2. Filter by job title (sales / motoris / booster / marketing / task order / to)
    const jobTitle = (emp.job_position?.title || '').toLowerCase()
    const isSales = /\b(sales|motoris|booster|marketing|task order|take order|to|mt|kam|modern trade)\b/i.test(jobTitle)

    // 3. Filter by Sales Type
    let matchSalesType = true
    const type = masterStore.selectedSalesType
    if (type === 'sales_mt') {
      matchSalesType = /\b(mt|modern trade|kam|key account)\b/i.test(jobTitle)
    } else if (type === 'sales_to') {
      matchSalesType = /\b(to|take order)\b/i.test(jobTitle)
    } else if (type === 'sales_motoris_booster') {
      matchSalesType = /\b(motoris|booster)\b/i.test(jobTitle)
    }

    return isActive && belongsToBranch && isSales && matchSalesType
  })
})

const years = computed(() => {
  const currentYear = new Date().getFullYear()
  return Array.from({ length: 5 }, (_, i) => (currentYear - i).toString())
})

const months = [
  { value: '1', label: 'Januari' },
  { value: '2', label: 'Februari' },
  { value: '3', label: 'Maret' },
  { value: '4', label: 'April' },
  { value: '5', label: 'Mei' },
  { value: '6', label: 'Juni' },
  { value: '7', label: 'Juli' },
  { value: '8', label: 'Agustus' },
  { value: '9', label: 'September' },
  { value: '10', label: 'Oktober' },
  { value: '11', label: 'November' },
  { value: '12', label: 'Desember' },
]
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const showAdjustModal = ref(false)
const showWaitingStockModal = ref(false)
const selectedOrder = ref<SalesOrder | null>(null)
const selectedOrderForWaitingStock = ref<SalesOrder | null>(null)
const loading = ref(false)

function openWaitingStockModal(order: SalesOrder) {
  selectedOrderForWaitingStock.value = order
  showWaitingStockModal.value = true
}

const showCreateStoreModal = ref(false)
const isCreatingStore = ref(false)
const storeForm = ref({
  name: '',
  owner_name: '',
  phone_number: '',
  address: '',
  assigned_employee_id: '',
  category: 'TOKO_BARU',
})

async function handleCreateStore() {
  if (!storeForm.value.name || !storeForm.value.address) {
    toast.error('Nama Toko dan Alamat wajib diisi')
    return
  }

  isCreatingStore.value = true
  try {
    const formData = new FormData()
    formData.append('name', storeForm.value.name)
    formData.append('owner_name', storeForm.value.owner_name)
    formData.append('phone_number', storeForm.value.phone_number)
    formData.append('address', storeForm.value.address)
    // Send 0 for lat/long if not provided
    formData.append('latitude', '0')
    formData.append('longitude', '0')
    formData.append('is_active', 'true')
    formData.append('category', storeForm.value.category)
    formData.append('company_id', masterStore.selectedBranchCompanyId)

    if (storeForm.value.assigned_employee_id) {
      formData.append('assigned_employee_id', storeForm.value.assigned_employee_id)
    }

    const res = await apiClient.post('/stores', formData)
    
    // Refresh stores data
    await masterStore.fetchStores(masterStore.selectedBranchCompanyId)
    if (form.value.employee_id) {
      await fetchStoresForSalesman(form.value.employee_id)
    }
    
    toast.success('Toko berhasil ditambahkan')
    
    // Auto-select the newly created store if returned by API
    if (res.data?.data?.id) {
      form.value.store_id = res.data.data.id
    }
    
    showCreateStoreModal.value = false
    
    // Reset form
    storeForm.value = {
      name: '',
      owner_name: '',
      phone_number: '',
      address: '',
      assigned_employee_id: '',
      category: 'TOKO_BARU',
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menambahkan toko')
  } finally {
    isCreatingStore.value = false
  }
}

function openCreateStoreModal() {
  storeForm.value.assigned_employee_id = form.value.employee_id || ''
  showCreateStoreModal.value = true
}

const confirmDialog = ref({
  open: false,
  title: '',
  description: '',
  variant: 'info' as 'danger' | 'warning' | 'success' | 'info',
  confirmText: '',
  cancelText: '',
  isLoading: false,
  onConfirm: () => {},
})

const expandedItemRowIds = ref<string[]>([])
const toggleExpand = (id: string) => {
  if (expandedItemRowIds.value.includes(id)) {
    expandedItemRowIds.value = expandedItemRowIds.value.filter(i => i !== id)
  } else {
    expandedItemRowIds.value.push(id)
  }
}

const rejectModal = ref({
  open: false,
  id: '',
  notes: '',
  isAdmin: false,
  isLoading: false,
})

// --- Export e-Faktur State ---
const showExportModal = ref(false)
const exportForm = ref({
  start_date: '',
  end_date: '',
  company_id: ''
})

async function handleExportEFaktur() {
  if (!exportForm.value.start_date || !exportForm.value.end_date) {
    toast.error('Pilih tanggal awal dan akhir')
    return
  }
  
  try {
    const res = await apiClient.get('/tax/efaktur/export', {
      params: {
        start_date: exportForm.value.start_date,
        end_date: exportForm.value.end_date,
        company_id: exportForm.value.company_id || masterStore.selectedBranchCompanyId
      },
      responseType: 'blob'
    })
    
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `efaktur_${exportForm.value.start_date}_${exportForm.value.end_date}.csv`)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    toast.success('Export berhasil')
    showExportModal.value = false
  } catch(e) {
    toast.error('Gagal export e-Faktur')
  }
}

// --- Set Faktur Pajak State ---
const showSetFakturModal = ref(false)
const fakturForm = ref({
  sales_order_id: '',
  tax_invoice_number: '',
  tax_invoice_date: '',
  tax_rate: 11,
})
const fakturOrderAmount = ref(0)
const fakturLoading = ref(false)

function openSetFaktur(order: SalesOrder) {
  fakturForm.value = {
    sales_order_id: order.id,
    tax_invoice_number: order.tax_invoice_number || '',
    tax_invoice_date: order.tax_invoice_date || '',
    tax_rate: 11
  }
  fakturOrderAmount.value = order.total_amount || 0
  showSetFakturModal.value = true
}

const fakturDpp = computed(() => {
  return fakturOrderAmount.value / (1 + (fakturForm.value.tax_rate / 100))
})
const fakturPpn = computed(() => {
  return fakturOrderAmount.value - fakturDpp.value
})

async function submitSetFaktur() {
  if (!fakturForm.value.tax_invoice_number || !fakturForm.value.tax_invoice_date) {
    toast.error('Lengkapi NSFP dan Tanggal Faktur')
    return
  }
  fakturLoading.value = true
  try {
    await apiClient.post('/tax/invoice', fakturForm.value)
    toast.success('Faktur Pajak berhasil disimpan')
    showSetFakturModal.value = false
    fetchOrders()
  } catch (error: any) {
    toast.error(error?.response?.data?.error || 'Gagal menyimpan faktur pajak')
  } finally {
    fakturLoading.value = false
  }
}

// --- Form ---
const getTodayDateString = () => {
  const d = new Date()
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const form = ref({
  flow_type: 'TO', 
  employee_id: '',
  store_id: '',
  store_category: 'TOKO_LAMA',
  order_date: getTodayDateString(),
  payment_method: 'CASH',
  paid_amount: 0,
  tenor: 1,
  notes: '',
  items: [
    { product_id: '', ordered_quantity: 1, unit: 'PCS', pieces_per_unit: 1, price: 0 },
  ] as any[],
})

// Dynamic stores for the selected salesman
const salesmanStores = ref<any[]>([])
const isLoadingSalesmanStores = ref(false)

const fetchStoresForSalesman = async (employeeId?: string) => {
  if (!employeeId) {
    salesmanStores.value = []
    return
  }
  isLoadingSalesmanStores.value = true
  try {
    const res = await apiClient.get('/stores', {
      params: {
        assigned_employee_id: employeeId,
        limit: 1000,
        status: 'ACTIVE',
        company_id: masterStore.selectedBranchCompanyId || '',
      }
    })
    salesmanStores.value = res.data?.data || []
  } catch (err) {
    console.error('Failed to fetch stores for salesman:', err)
  } finally {
    isLoadingSalesmanStores.value = false
  }
}

watch(
  () => form.value.employee_id,
  async (newEmpId) => {
    if (newEmpId) {
      await fetchStoresForSalesman(newEmpId)
    } else {
      salesmanStores.value = []
    }
  },
  { immediate: true }
)

const filteredStores = computed(() => {
  if (!form.value.employee_id) {
    return masterStore.stores
  }

  // Combine stores from dynamic salesman fetch and masterStore stores
  const storeMap = new Map<string, any>()

  // 1. Add stores fetched specifically for this salesman (ensures full coverage beyond limit 500)
  salesmanStores.value.forEach((s: any) => {
    if (s && s.id) storeMap.set(s.id, s)
  })

  // 2. Add stores from masterStore that match assigned_employee_id
  masterStore.stores.forEach((s: any) => {
    if (s && s.id && s.assigned_employee_id === form.value.employee_id) {
      if (!storeMap.has(s.id)) {
        storeMap.set(s.id, s)
      }
    }
  })

  return Array.from(storeMap.values())
})

const storeOptions = computed<SelectOption[]>(() => {
  return filteredStores.value.map((s: any) => {
    const addressParts = [s.address, s.kelurahan, s.kecamatan, s.city].filter(Boolean)
    const fullAddress = addressParts.length > 0 ? addressParts.join(', ') : (s.address || 'Alamat belum diatur')

    return {
      value: s.id,
      label: s.name,
      sublabel: fullAddress,
      badge: s.category === 'TOKO_BARU' ? 'Toko Baru' : undefined,
      badgeColor: 'bg-emerald-50 text-emerald-700',
      extra: s,
    }
  })
})

const handleStoreChange = (storeId: string, option?: SelectOption) => {
  form.value.store_id = storeId
  if (option?.extra?.category) {
    form.value.store_category = option.extra.category
  }
}

const UNIT_OPTIONS = [
  'KARTON',
  'DUS',
  'BOX',
  'KRAT',
  'PCS',
  'BAL',
  'PACK',
  'KG',
  'GRAM',
  'SACHET',
  'JERIGEN',
  'BOTOL',
]

const statusOptions = [
  'ALL',
  'DRAFT',
  'WAITING_TAX_DOCS',
  'CONFIRMED',
  'IN_DELIVERY',
  'DELIVERED',
  'PAID',
  'WAITING_WAREHOUSE',
  'WAITING_STOCK',
  'SHIPPED',
  'CONVERTED',
  'CANCELLED',
  'REJECTED',
  'DELETED',
]

const statusColor: Record<string, string> = {
  DRAFT: 'bg-slate-100 text-slate-600',
  WAITING_TAX_DOCS: 'bg-violet-100 text-violet-700 border border-violet-200',
  CONFIRMED: 'bg-green-100 text-green-700',
  IN_DELIVERY: 'bg-blue-100 text-blue-700',
  DELIVERED: 'bg-teal-100 text-teal-700',
  PAID: 'bg-emerald-100 text-emerald-700',
  WAITING_WAREHOUSE: 'bg-amber-100 text-amber-700',
  WAITING_STOCK: 'bg-orange-100 text-orange-700',
  PROCESSING: 'bg-blue-100 text-blue-700',
  SHIPPED: 'bg-indigo-100 text-indigo-700',
  CONVERTED: 'bg-emerald-100 text-emerald-700',
  CANCELLED: 'bg-red-100 text-red-600',
  REJECTED: 'bg-red-50 text-red-400',
  DELETED: 'bg-rose-100 text-rose-700 border border-rose-200',
}

// Tax Approval Modal State
const showTaxModal = ref(false)
const taxForm = ref({
  id: '',
  billing_note_url: '',
  tax_invoice_url: '',
  delivery_note_url: '',
  destination_branch_id: '',
})
const taxLoading = ref(false)

const filteredOrders = computed(() => {
  return soStore.orders
})

// Pesanan yang disetujui Admin Nota
const approvedOrders = computed(() => {
  return filteredOrders.value.filter(
    (o) =>
      !['CANCELLED', 'REJECTED', 'DELETED'].includes(o.status) &&
      (Boolean(o.admin_nota_confirmed_at) ||
        ['CONFIRMED', 'APPROVED', 'IN_DELIVERY', 'DELIVERED', 'PAID', 'SHIPPED', 'CONVERTED'].includes(o.status)),
  )
})

// Pesanan yang menunggu konfirmasi Admin Nota
const pendingApprovalOrders = computed(() => {
  return filteredOrders.value.filter(
    (o) =>
      !o.admin_nota_confirmed_at &&
      ['DRAFT', 'WAITING_TAX_DOCS', 'WAITING_WAREHOUSE', 'WAITING_STOCK', 'PROCESSING'].includes(o.status),
  )
})

// Pesanan yang ditolak atau dibatalkan
const rejectedOrders = computed(() => {
  return filteredOrders.value.filter((o) => ['REJECTED', 'CANCELLED'].includes(o.status))
})

const waitingStockOrders = computed(() => {
  return filteredOrders.value.filter((o) => o.status === 'WAITING_STOCK')
})

// Akumulasi nominal (Total Amount) & Total Counts dari Seluruh Data Sesuai Filter
const totalApprovedAmount = computed(() => {
  if (soStore.summary?.total_approved_amount !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_approved_amount
  }
  return approvedOrders.value.reduce((sum, o) => sum + (o.total_amount || 0), 0)
})

const totalApprovedCount = computed(() => {
  if (soStore.summary?.total_approved_count !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_approved_count
  }
  return approvedOrders.value.length
})

const totalPendingAmount = computed(() => {
  if (soStore.summary?.total_pending_amount !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_pending_amount
  }
  return pendingApprovalOrders.value.reduce((sum, o) => sum + (o.total_amount || 0), 0)
})

const totalPendingCount = computed(() => {
  if (soStore.summary?.total_pending_count !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_pending_count
  }
  return pendingApprovalOrders.value.length
})

const totalWaitingStockAmount = computed(() => {
  if (soStore.summary?.total_waiting_stock_amount !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_waiting_stock_amount
  }
  return waitingStockOrders.value.reduce((sum, o) => sum + (o.total_amount || 0), 0)
})

const totalWaitingStockCount = computed(() => {
  if (soStore.summary?.total_waiting_stock_count !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_waiting_stock_count
  }
  return waitingStockOrders.value.length
})

const totalRejectedAmount = computed(() => {
  if (soStore.summary?.total_rejected_amount !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_rejected_amount
  }
  return rejectedOrders.value.reduce((sum, o) => sum + (o.total_amount || 0), 0)
})

const totalRejectedCount = computed(() => {
  if (soStore.summary?.total_rejected_count !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_rejected_count
  }
  return rejectedOrders.value.length
})

const totalFilteredAmount = computed(() => {
  if (soStore.summary?.total_filtered_amount !== undefined && soStore.totalOrders > 0) {
    return soStore.summary.total_filtered_amount
  }
  return filteredOrders.value.reduce((sum, o) => sum + (o.total_amount || 0), 0)
})

const totalFilteredCount = computed(() => {
  if (soStore.totalOrders !== undefined && soStore.totalOrders > 0) {
    return soStore.totalOrders
  }
  return filteredOrders.value.length
})

const getItemStockBreakdown = (item: any) => {
  const totalPieces = (Number(item.ordered_quantity) || 1) * (Number(item.pieces_per_unit) || 1)
  const batches = item.batches || []

  const backorderBatch = batches.find((b: any) => b.batch_no === 'BACKORDER')
  const backorderPieces = backorderBatch ? Number(backorderBatch.quantity || 0) : 0

  const realBatches = batches.filter((b: any) => b.batch_no !== 'BACKORDER' && Number(b.quantity || 0) > 0)
  const readyPieces = realBatches.reduce((sum: number, b: any) => sum + Number(b.quantity || 0), 0)

  const hasBackorder = backorderPieces > 0
  const isFullyReady = backorderPieces === 0 && (readyPieces >= totalPieces || batches.length > 0)

  const piecesPerUnit = Number(item.pieces_per_unit) || 1
  const backorderUnits = Math.ceil(backorderPieces / piecesPerUnit)
  const readyUnits = Math.floor(readyPieces / piecesPerUnit)

  const readyPercentage = totalPieces > 0 ? Math.min(100, Math.round((readyPieces / totalPieces) * 100)) : 100

  return {
    totalPieces,
    readyPieces,
    backorderPieces,
    backorderUnits,
    readyUnits,
    hasBackorder,
    isFullyReady,
    readyPercentage,
    realBatches,
  }
}

const getSOBackorderSummary = (order: SalesOrder) => {
  if (!order || !order.items || order.items.length === 0) return null

  let totalBackorderItems = 0
  let totalShortagePieces = 0

  for (const item of order.items) {
    const breakdown = getItemStockBreakdown(item)
    if (breakdown.hasBackorder) {
      totalBackorderItems++
      totalShortagePieces += breakdown.backorderPieces
    }
  }

  return {
    hasBackorder: totalBackorderItems > 0 || order.status === 'WAITING_STOCK',
    totalBackorderItems,
    totalShortagePieces,
  }
}

const isRecheckingStock = ref(false)
async function handleRecheckStock(order: SalesOrder) {
  isRecheckingStock.value = true
  try {
    const res = await soStore.recheckStock(order.id)
    toast.success(res.message || 'Pengecekan stok berhasil')
    if (selectedOrder.value && selectedOrder.value.id === order.id) {
      selectedOrder.value = res.data
    }
    await fetchOrders()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal memeriksa stok pesanan')
  } finally {
    isRecheckingStock.value = false
  }
}

const stockSummary = computed(() => {
  if (!selectedOrder.value) return []
  return selectedOrder.value.items.map((item) => ({
    ...item,
    productName: item.product?.name || item.product_id,
    breakdown: getItemStockBreakdown(item),
  }))
})

const parsedMtInfo = computed(() => {
  if (!selectedOrder.value || !selectedOrder.value.notes) return null
  try {
    const data = JSON.parse(selectedOrder.value.notes)
    if (data && (data.po_number || data.fulfillment_type)) {
      return data
    }
  } catch (e) {
    // Not JSON
  }
  return null
})

const itemTotal = computed(() =>
  form.value.items.reduce((sum, i) => sum + i.ordered_quantity * i.pieces_per_unit * i.price, 0),
)

function addItem() {
  form.value.items.push({
    product_id: '',
    ordered_quantity: 1,
    unit: 'PCS',
    pieces_per_unit: 1,
    price: 0,
  })
}
function removeItem(idx: number) {
  form.value.items.splice(idx, 1)
}
function onProductChange(idx: number) {
  const p = factoryStore.products.find((p) => p.id === form.value.items[idx].product_id)
  if (p) form.value.items[idx].price = p.selling_price || 0
}

const branchInventory = ref<any[]>([])

async function fetchBranchInventory() {
  if (!masterStore.selectedBranchId || masterStore.selectedBranchId === 'ALL_BRANCHES') {
    branchInventory.value = []
    return
  }
  try {
    const res = await apiClient.get('/warehouse/stock', { params: { branch_id: masterStore.selectedBranchId } })
    branchInventory.value = res.data
  } catch(e) {
    console.error('Failed to fetch branch inventory', e)
    branchInventory.value = []
  }
}

const getProductStock = (productId: string) => {
  const stock = branchInventory.value.find((s: any) => s.product_id === productId)
  return stock ? stock.quantity : 0
}

let searchDebounceTimeout: ReturnType<typeof setTimeout> | null = null

async function fetchOrders() {
  if (!masterStore.selectedBranchId) return
  let formattedMonth = selectedMonth.value
  if (selectedMonth.value !== 'ALL' && selectedYear.value !== 'ALL') {
    formattedMonth = `${selectedYear.value}-${selectedMonth.value.padStart(2, '0')}`
  } else if (selectedMonth.value !== 'ALL') {
    const y = new Date().getFullYear()
    formattedMonth = `${y}-${selectedMonth.value.padStart(2, '0')}`
  }

  await soStore.fetchByBranch(
    masterStore.selectedBranchId,
    selectedStatus.value,
    currentPage.value,
    itemsPerPage.value,
    {
      employee_id: selectedEmployeeId.value,
      company_id: selectedCompanyId.value,
      date: selectedDate.value,
      month: formattedMonth,
      search: searchQuery.value,
    },
  )
}

async function handleCreate() {
  if (form.value.flow_type !== 'OFFICE' && !form.value.employee_id) {
    toast.error('Pilih salesman terlebih dahulu')
    return
  }
  if (!form.value.store_id) {
    toast.error('Pilih toko terlebih dahulu')
    return
  }
  if (form.value.items.some((i) => !i.product_id)) {
    toast.error('Lengkapi semua produk')
    return
  }

  let finalEmployeeId = form.value.employee_id
  if (form.value.flow_type === 'OFFICE') {
    const currentAdmin = masterStore.employees.find((e: any) => e.user_id === authStore.user?.id)
    if (!currentAdmin) {
      toast.error('Gagal menemukan data karyawan Anda. Pastikan akun Anda terhubung ke data Employee.')
      return
    }
    finalEmployeeId = currentAdmin.id
  }

  const selectedEmp = masterStore.employees.find((e: any) => e.id === finalEmployeeId)
  const branchId = selectedEmp?.operating_branch_id || selectedEmp?.branch_id || masterStore.selectedBranchId
  const companyId = selectedEmp?.company_id || masterStore.selectedBranchCompanyId

  if (!branchId || branchId === 'ALL_BRANCHES') {
    toast.error('Cabang tidak valid. Silakan pastikan Anda memilih cabang yang benar.')
    return
  }

  loading.value = true
  try {
    const orderDateISO = form.value.order_date
      ? new Date(form.value.order_date).toISOString()
      : new Date().toISOString()

    if (form.value.flow_type === 'TO' || form.value.flow_type === 'OFFICE') {
      await soStore.createSO({
        branch_id: branchId,
        company_id: companyId,
        employee_id: finalEmployeeId,
        store_id: form.value.store_id,
        store_category: form.value.store_category,
        order_date: orderDateISO,
        notes: form.value.notes,
        items: form.value.items.map((i) => ({
          product_id: i.product_id,
          ordered_quantity: i.ordered_quantity,
          unit: i.unit,
          pieces_per_unit: i.pieces_per_unit,
          price: i.price,
        })),
      })
      await fetchOrders()
    } else {
      // Direct Sales (Motoris / Booster) - creates SalesTransaction directly
      const totalAmount = itemTotal.value
      await apiClient.post('/sales/transactions', {
        company_id: companyId,
        store_id: form.value.store_id,
        employee_id: form.value.employee_id,
        total_amount: totalAmount,
        paid_amount: form.value.payment_method === 'CASH' ? totalAmount : form.value.paid_amount,
        store_category: form.value.store_category,
        transaction_date: orderDateISO,
        payment_method: form.value.payment_method,
        tenor: form.value.payment_method === 'TEMPO' ? form.value.tenor : 1,
        notes: form.value.notes,
        items: form.value.items.map((i) => ({
          product_id: i.product_id,
          quantity: i.ordered_quantity * i.pieces_per_unit,
          ordered_quantity: i.ordered_quantity,
          unit: i.unit,
          pieces_per_unit: i.pieces_per_unit,
          price: i.price,
        })),
      })
    }
    toast.success('Pesanan berhasil dibuat')
    showCreateModal.value = false
    form.value = {
      flow_type: 'TO',
      employee_id: '',
      store_id: '',
      store_category: 'TOKO_LAMA',
      order_date: getTodayDateString(),
      payment_method: 'CASH',
      paid_amount: 0,
      tenor: 1,
      notes: '',
      items: [{ product_id: '', ordered_quantity: 1, unit: 'PCS', pieces_per_unit: 1, price: 0 }],
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal membuat pesanan')
  } finally {
    loading.value = false
  }
}

function handleConfirm(id: string) {
  confirmDialog.value = {
    open: true,
    title: 'Konfirmasi Pesanan',
    description: 'Konfirmasi SO? Stok gudang akan di-reserve.',
    variant: 'success',
    confirmText: 'Ya, Konfirmasi',
    cancelText: 'Kembali',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await soStore.confirmSO(id)
        toast.success('SO dikonfirmasi, stok berhasil di-reserve')
        confirmDialog.value.open = false
      } catch (e: any) {
        toast.error(e?.response?.data?.error || 'Gagal konfirmasi')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

/** Admin Nota: Verifikasi nota salesman → status CONFIRMED, nota siap di-batch */
function handleAdminConfirm(id: string) {
  confirmDialog.value = {
    open: true,
    title: 'Verifikasi Nota',
    description:
      'Verifikasi nota ini? Status akan berubah menjadi CONFIRMED dan nota bisa masuk Batch Pengiriman.',
    variant: 'success',
    confirmText: 'Verifikasi',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await soStore.adminConfirmSO(id)
        toast.success('Nota berhasil diverifikasi. Supervisor bisa membuat batch pengiriman.')
        confirmDialog.value.open = false
      } catch (e: any) {
        toast.error(e?.response?.data?.error || 'Gagal verifikasi nota')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

/** Super Admin: Cek apakah pesanan bisa diselesaikan instan (1-klik) */
function canSuperAdminComplete(order: SalesOrder): boolean {
  if (authStore.user?.role !== 'superadmin') return false
  if (order.deleted_at) return false
  return order.status === 'CONFIRMED'
}

/** Super Admin: 1-Klik Selesaikan Pesanan (Pengiriman, Faktur, Kasir, Omzet) */
function handleSuperAdminComplete(order: SalesOrder) {
  confirmDialog.value = {
    open: true,
    title: `⚡ Selesaikan Pesanan ${order.so_number}`,
    description: `Jalankan penyelesaian 1-klik untuk pesanan ${order.so_number} (${order.store?.name || ''})?\n\nOperasi ini akan langsung:\n1. Menyetujui Nota (Admin Nota)\n2. Menyelesaikan Pengiriman & POD\n3. Menerbitkan & Memverifikasi Faktur (Verified)\n4. Menyetujui Setoran Kasir (Status LUNAS / Paid)\n5. Mengesahkan Penerimaan Omzet & Target KPI Salesman`,
    variant: 'success',
    confirmText: 'Ya, Selesaikan Sekarang',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await soStore.superAdminCompleteSO(order.id)
        toast.success(`Pesanan ${order.so_number} berhasil diselesaikan secara penuh!`)
        confirmDialog.value.open = false
        if (showDetailModal.value && selectedOrder.value?.id === order.id) {
          const updated = await soStore.fetchByID(order.id)
          selectedOrder.value = updated
        }
      } catch (e: any) {
        toast.error(e?.response?.data?.error || 'Gagal menyelesaikan pesanan')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

const isBulkCompleting = ref(false)

const currentBranchScopeLabel = computed(() => {
  const branchId = masterStore.selectedBranchId
  if (!branchId || branchId === 'ALL_BRANCHES') {
    return 'Seluruh Cabang'
  }
  return masterStore.selectedBranchName || 'Cabang Terpilih'
})

// Gunakan real-time count dari database untuk pesanan berstatus CONFIRMED yang mencakup semua halaman & riwayat bulan lalu
const realConfirmedOrdersCount = computed(() => {
  return soStore.pendingStats?.confirmed_count || 0
})

async function refreshPendingStats() {
  if (authStore.user?.role === 'superadmin') {
    await soStore.fetchPendingStats(
      masterStore.selectedBranchId,
      selectedCompanyId.value,
    )
  }
}

/** Super Admin: 1-Klik Selesaikan SEMUA Pesanan SO Confirmed (Mendukung Seluruh Cabang maupun Cabang Spesifik) */
async function handleBulkSuperAdminComplete() {
  const branchId = masterStore.selectedBranchId
  const isAllBranches = !branchId || branchId === 'ALL_BRANCHES'
  const scopeName = isAllBranches ? 'SELURUH CABANG' : (masterStore.selectedBranchName || 'Cabang Terpilih')
  
  // Refresh stats first to get absolute real-time numbers from database
  await refreshPendingStats()
  const stats = soStore.pendingStats
  const count = stats.confirmed_count || 0

  if (count === 0) {
    toast.info(`Tidak ada pesanan SO berstatus CONFIRMED pada ${scopeName}`)
    return
  }

  confirmDialog.value = {
    open: true,
    title: `⚡ Selesaikan SEMUA SO Confirmed (${scopeName})`,
    description: `Ditemukan ${count} pesanan SO berstatus CONFIRMED dari seluruh periode pada ${scopeName}.\n\nApakah Anda yakin ingin menyelesaikan SEMUA ${count} pesanan SO Confirmed tersebut dalam 1 kali klik?\n\nOperasi massal ini akan langsung:\n1. Menyetujui semua nota (Admin Nota)\n2. Menyelesaikan Pengiriman & POD\n3. Menerbitkan & Memverifikasi Faktur (Verified)\n4. Menyetujui Setoran Kasir (Status Lunas)\n5. Mengesahkan Penerimaan Omzet & Target KPI Salesman\n\n*Catatan: Pesanan Draft dan Waiting Stock TIDAK akan diproses.`,
    variant: 'success',
    confirmText: `Ya, Selesaikan (${count} SO Confirmed)`,
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      isBulkCompleting.value = true
      try {
        const res = await soStore.superAdminBulkCompleteSO({
          branch_id: branchId as string,
          company_id: selectedCompanyId.value,
        })
        const data = res.data
        if (data && data.failed_count > 0) {
          if (data.success_count > 0) {
            toast.warning(`Selesai sebagian: ${data.success_count} pesanan CONFIRMED selesai, ${data.failed_count} gagal:\n${data.errors?.slice(0, 3).join('\n')}`)
          } else {
            toast.error(`Gagal memproses ${data.failed_count} pesanan:\n${data.errors?.slice(0, 3).join('\n')}`)
          }
        } else if (data) {
          toast.success(`Berhasil memproses massal: ${data.success_count} pesanan CONFIRMED selesai`)
        } else {
          toast.success(res.message || 'Penyelesaian massal pesanan CONFIRMED berhasil diproses!')
        }
        confirmDialog.value.open = false
        await fetchOrders()
        await refreshPendingStats()
      } catch (e: any) {
        toast.error(e?.response?.data?.error || 'Gagal memproses penyelesaian massal pesanan SO')
      } finally {
        confirmDialog.value.isLoading = false
        isBulkCompleting.value = false
      }
    },
  }
}

function openReject(id: string, isAdmin: boolean = false) {
  rejectModal.value = {
    open: true,
    id,
    notes: '',
    isAdmin,
    isLoading: false,
  }
}

async function submitReject() {
  if (!rejectModal.value.notes) {
    toast.error('Alasan penolakan harus diisi')
    return
  }
  rejectModal.value.isLoading = true
  try {
    if (rejectModal.value.isAdmin) {
      await soStore.adminRejectSO(rejectModal.value.id, rejectModal.value.notes)
    } else {
      await soStore.rejectSO(rejectModal.value.id, rejectModal.value.notes)
    }
    toast.success('Pesanan berhasil ditolak')
    rejectModal.value.open = false
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menolak pesanan')
  } finally {
    rejectModal.value.isLoading = false
  }
}

function openTaxApproval(order: SalesOrder) {
  taxForm.value = {
    id: order.id,
    billing_note_url: order.billing_note_url || '',
    tax_invoice_url: order.tax_invoice_url || '',
    delivery_note_url: order.delivery_note_url || '',
    destination_branch_id: order.destination_branch_id || '',
  }
  showTaxModal.value = true
}

async function submitTaxApproval() {
  if (!taxForm.value.billing_note_url || !taxForm.value.tax_invoice_url || !taxForm.value.delivery_note_url || !taxForm.value.destination_branch_id) {
    toast.error('Semua dokumen (Nota Penagihan, Faktur Pajak, Surat Jalan) dan Cabang Gudang Pengirim wajib diisi!')
    return
  }
  taxLoading.value = true
  try {
    await soStore.approveTaxDocuments(taxForm.value.id, {
      billing_note_url: taxForm.value.billing_note_url,
      tax_invoice_url: taxForm.value.tax_invoice_url,
      delivery_note_url: taxForm.value.delivery_note_url,
      destination_branch_id: taxForm.value.destination_branch_id,
    })
    toast.success('Dokumen pajak berhasil disetujui, pesanan sekarang CONFIRMED')
    showTaxModal.value = false
    await fetchOrders()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyetujui dokumen pajak')
  } finally {
    taxLoading.value = false
  }
}

function handleCancel(id: string) {
  confirmDialog.value = {
    open: true,
    title: 'Batalkan Pesanan',
    description: 'Batalkan SO ini? Reserve stok akan dilepas.',
    variant: 'warning',
    confirmText: 'Ya, Batalkan',
    cancelText: 'Kembali',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await soStore.cancelSO(id)
        toast.success('SO dibatalkan, stok dilepas')
        confirmDialog.value.open = false
      } catch (e: any) {
        toast.error(e?.response?.data?.error || 'Gagal membatalkan')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

function handleConvert(id: string) {
  confirmDialog.value = {
    open: true,
    title: 'Terbitkan Faktur',
    description: 'Terbitkan Faktur Penjualan? Stok akan DIPOTONG PERMANEN dari gudang.',
    variant: 'warning',
    confirmText: 'Terbitkan',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        const invoice = await soStore.convertToInvoice(id, masterStore.selectedBranchCompanyId)
        toast.success(`Faktur ${invoice.receipt_no} berhasil diterbitkan!`)
        confirmDialog.value.open = false
      } catch (e: any) {
        toast.error(e?.response?.data?.error || 'Gagal menerbitkan faktur')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

function canConvertDirect(order: SalesOrder) {
  return ['DRAFT', 'CONFIRMED', 'APPROVED'].includes(order.status) && !order.invoice_id && !order.delivery_batch_id && !order.is_mt
}

function handleConvertDirect(order: SalesOrder) {
  confirmDialog.value = {
    open: true,
    title: 'Langsung ke Faktur Penjualan',
    description: `Terbitkan faktur untuk SO ${order.so_number} tanpa pengiriman? Stok bronjong sales yang tersedia akan dipakai lebih dulu, lalu sisanya dari gudang cabang. Pesanan DRAFT akan sekaligus disetujui. Pastikan barang sudah diserahkan kepada pelanggan.`,
    variant: 'warning',
    confirmText: 'Terbitkan Faktur',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        const invoice = await soStore.convertDirectToInvoice(order.id)
        toast.success(`Faktur ${invoice.receipt_no} berhasil diterbitkan tanpa pengiriman`)
        confirmDialog.value.open = false
        showDetailModal.value = false
        await fetchOrders()
      } catch (e: any) {
        toast.error(e?.response?.data?.error || 'Gagal menerbitkan faktur langsung')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

const showDeleteSafetyModal = ref(false)
const deleteTargetSO = ref<SalesOrder | null>(null)
const deleteReason = ref('')
const deleteConfirmText = ref('')
const isDeletingSO = ref(false)

function handleDelete(order: SalesOrder) {
  deleteTargetSO.value = order
  deleteReason.value = ''
  deleteConfirmText.value = ''
  showDeleteSafetyModal.value = true
}

async function confirmDeleteSO() {
  if (!deleteTargetSO.value) return
  if (!deleteReason.value.trim()) {
    toast.error('Alasan penghapusan wajib diisi')
    return
  }
  if (deleteConfirmText.value.trim() !== deleteTargetSO.value.so_number) {
    toast.error('Nomor SO konfirmasi tidak cocok')
    return
  }
  isDeletingSO.value = true
  try {
    await soStore.deleteSO(deleteTargetSO.value.id, masterStore.selectedBranchId as string, deleteReason.value)
    toast.success('Pesanan berhasil dihapus (Soft Delete)')
    showDeleteSafetyModal.value = false
    deleteTargetSO.value = null
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menghapus pesanan')
  } finally {
    isDeletingSO.value = false
  }
}

function handleRestore(order: SalesOrder) {
  confirmDialog.value = {
    open: true,
    title: 'Pulihkan Pesanan SO',
    description: `Pulihkan pesanan ${order.so_number} (${order.store?.name || ''}) kembali ke daftar aktif?`,
    variant: 'info',
    confirmText: 'Ya, Pulihkan',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await soStore.restoreSO(order.id, masterStore.selectedBranchId as string)
        toast.success('Pesanan berhasil dipulihkan')
        confirmDialog.value.open = false
      } catch (e: any) {
        toast.error(e?.response?.data?.error || 'Gagal memulihkan pesanan')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

function openDetail(order: SalesOrder) {
  selectedOrder.value = order
  showDetailModal.value = true
}

function openAdjust(order: SalesOrder) {
  selectedOrder.value = order
  showAdjustModal.value = true
}
const isMapModalOpen = ref(false)
const selectedOrderForMap = ref<any>(null)

const openMapModal = (order: any) => {
  if (getOrderLocationStatus(order).status === 'UNKNOWN') return
  selectedOrderForMap.value = order
  isMapModalOpen.value = true
}
const getOrderLocationStatus = (order: any) => {
  const lat = order?.latitude !== undefined && order?.latitude !== null ? Number(order.latitude) : null
  const lng = order?.longitude !== undefined && order?.longitude !== null ? Number(order.longitude) : null
  const stLat = order?.store?.latitude !== undefined && order?.store?.latitude !== null ? Number(order.store.latitude) : null
  const stLng = order?.store?.longitude !== undefined && order?.store?.longitude !== null ? Number(order.store.longitude) : null

  if (
    lat === null ||
    lng === null ||
    stLat === null ||
    stLng === null ||
    isNaN(lat) ||
    isNaN(lng) ||
    isNaN(stLat) ||
    isNaN(stLng) ||
    lat === 0 ||
    lng === 0 ||
    stLat === 0 ||
    stLng === 0
  ) {
    return { status: 'UNKNOWN', text: 'Tidak ada data', color: 'text-slate-400', bg: 'bg-slate-100' }
  }
  const dist = getDistance(lat, lng, stLat, stLng)
  if (dist > 500) {
    return { status: 'OUTSIDE', text: `Luar Radius (${Math.round(dist)}m)`, color: 'text-red-600', bg: 'bg-red-50', icon: MapPinOff }
  }
  return { status: 'INSIDE', text: 'Sesuai Radius', color: 'text-emerald-600', bg: 'bg-emerald-50', icon: MapPin }
}

const formatCurrency = (v: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(v)
const formatDate = (d: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : '-'
const getCompanyName = (companyId: string) => {
  const company = masterStore.companies.find((c) => c.id === companyId)
  return company ? company.name : 'Perusahaan'
}

const {
  isRefreshing,
  lastRefreshedTimeStr,
  markRefreshed,
  refresh: triggerRefresh,
} = useDataRefresh(
  async ({ isManual }) => {
    const tasks: Promise<unknown>[] = [
      fetchOrders(),
      fetchBranchInventory(),
      refreshPendingStats(),
    ]
    if (isManual) {
      tasks.push(
        factoryStore.fetchProducts(true),
        masterStore.fetchBranches(true),
        masterStore.fetchCompanies(true),
        masterStore.fetchEmployees(true),
        masterStore.fetchStores(masterStore.selectedBranchCompanyId, true),
      )
      if (form.value.employee_id) {
        tasks.push(fetchStoresForSalesman(form.value.employee_id))
      }
    }
    await Promise.allSettled(tasks)

    // Sinkronkan detail pesanan yang sedang terbuka dengan data terkini
    if (selectedOrder.value?.id && (showDetailModal.value || showAdjustModal.value)) {
      const updated = soStore.orders.find((o) => o.id === selectedOrder.value?.id)
      if (updated) {
        selectedOrder.value = updated
      } else {
        try {
          const freshOrder = await soStore.fetchByID(selectedOrder.value.id)
          if (freshOrder) selectedOrder.value = freshOrder
        } catch {}
      }
    }
    if (selectedOrderForWaitingStock.value?.id && showWaitingStockModal.value) {
      const updated = soStore.orders.find((o) => o.id === selectedOrderForWaitingStock.value?.id)
      if (updated) selectedOrderForWaitingStock.value = updated
    }
  },
  {
    successMessage: 'Data Pesanan SO berhasil diperbarui',
    canAutoRefresh: () =>
      !showCreateModal.value &&
      !showAdjustModal.value &&
      !showTaxModal.value &&
      !showSetFakturModal.value &&
      !showDeleteSafetyModal.value &&
      !loading.value &&
      !isBulkCompleting.value,
  },
)

onMounted(() => {
  // Fire all background fetches without blocking render
  factoryStore.fetchProducts()
  masterStore.fetchBranches()
  masterStore.fetchCompanies()
  masterStore.fetchEmployees()
  masterStore.fetchStores(masterStore.selectedBranchCompanyId)
  // Fetch orders (main data) immediately
  Promise.allSettled([
    fetchOrders(),
    fetchBranchInventory(),
    refreshPendingStats(),
  ]).then(() => markRefreshed())
})
watch(
  () => masterStore.selectedBranchId,
  () => {
    currentPage.value = 1
    fetchOrders()
    fetchBranchInventory()
    masterStore.fetchStores(masterStore.selectedBranchCompanyId)
    selectedEmployeeId.value = 'ALL'
    form.value.employee_id = ''
    refreshPendingStats()
    markRefreshed()
  },
)
watch(
  () => masterStore.selectedSalesType,
  () => {
    currentPage.value = 1
    fetchOrders()
    markRefreshed()
  },
)

watch(
  [
    selectedStatus,
    selectedCompanyId,
    selectedEmployeeId,
    selectedDate,
    selectedMonth,
    selectedYear,
  ],
  () => {
    currentPage.value = 1
    fetchOrders()
    refreshPendingStats()
    markRefreshed()
  },
)

watch(searchQuery, () => {
  if (searchDebounceTimeout) clearTimeout(searchDebounceTimeout)
  searchDebounceTimeout = setTimeout(() => {
    currentPage.value = 1
    fetchOrders()
    markRefreshed()
  }, 400)
})

watch([currentPage, itemsPerPage], () => {
  fetchOrders()
  markRefreshed()
})
</script>

<template>
  <div class="space-y-8 p-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-3">
          <ClipboardList class="w-8 h-8 text-primary" />
          <span>Pesanan Order (SO)</span>
          <span v-if="masterStore.selectedBranchName" class="text-primary text-2xl font-bold">
            — {{ masterStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Kelola pesanan sebelum diterbitkan sebagai Faktur Penjualan.
        </p>
      </div>
      <div class="flex items-center flex-wrap gap-2">
        <RefreshDataButton
          :loading="isRefreshing || soStore.loading"
          :last-updated="lastRefreshedTimeStr"
          @click="triggerRefresh()"
        />

        <!-- Super Admin: Bulk Auto Complete Button (Seluruh Cabang / Cabang Tertentu) -->
        <Button
          v-if="authStore.user?.role === 'superadmin'"
          @click="handleBulkSuperAdminComplete"
          :disabled="isBulkCompleting || realConfirmedOrdersCount === 0"
          class="rounded-2xl px-5 py-6 font-black flex items-center gap-2 shadow-lg bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-white border border-amber-400/40 transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
          :title="`Selesaikan ${realConfirmedOrdersCount} pesanan SO Confirmed pada ${currentBranchScopeLabel}`"
        >
          <Sparkles class="w-5 h-5 text-amber-100" :class="{ 'animate-spin': isBulkCompleting }" />
          <span>Selesaikan Semua SO Confirmed - {{ currentBranchScopeLabel }}</span>
          <span class="ml-1 px-2 py-0.5 rounded-full bg-white/25 text-xs font-black">
            {{ realConfirmedOrdersCount }}
          </span>
        </Button>

        <Button
          @click="showExportModal = true"
          class="rounded-2xl px-6 py-6 font-bold flex items-center gap-2 shadow-lg bg-green-600 hover:bg-green-700 text-white"
        >
          <FileText class="w-5 h-5" /> Export e-Faktur DJP
        </Button>
        <Button
          @click="showCreateModal = true"
          class="rounded-2xl px-6 py-6 font-bold flex items-center gap-2 shadow-lg shadow-primary/20"
        >
          <Plus class="w-5 h-5" /> Buat Pesanan Baru
        </Button>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-col gap-2">
      <div class="flex flex-col lg:flex-row gap-2">
        <div class="relative flex-1">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari nomor SO, toko, atau salesman..."
            class="w-full pl-11 pr-4 py-3.5 bg-white border border-slate-200 rounded-2xl text-sm font-medium focus:ring-2 focus:ring-primary/20 focus:border-primary/30 outline-none shadow-sm"
          />
        </div>
        <div class="flex flex-wrap gap-2">
          <input
            v-model="selectedDate"
            type="date"
            class="bg-white border border-slate-200 rounded-2xl px-4 py-3.5 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20 shadow-sm min-w-[140px]"
            title="Filter Tanggal"
          />
          <select
            v-model="selectedMonth"
            class="bg-white border border-slate-200 rounded-2xl px-4 py-3.5 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20 shadow-sm min-w-[120px]"
          >
            <option value="ALL">Semua Bulan</option>
            <option v-for="m in months" :key="m.value" :value="m.value">{{ m.label }}</option>
          </select>
          <select
            v-model="selectedYear"
            class="bg-white border border-slate-200 rounded-2xl px-4 py-3.5 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20 shadow-sm min-w-[120px]"
          >
            <option value="ALL">Semua Tahun</option>
            <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
          </select>
        </div>
      </div>
      <div class="flex flex-wrap gap-2">
        <select
          v-model="selectedCompanyId"
          class="bg-white border border-slate-200 rounded-2xl px-4 py-3.5 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20 shadow-sm min-w-[160px]"
        >
          <option value="ALL">Semua Perusahaan</option>
          <option v-for="c in masterStore.companies" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <select
          v-model="selectedEmployeeId"
          class="bg-white border border-slate-200 rounded-2xl px-4 py-3.5 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20 shadow-sm min-w-[160px]"
        >
          <option value="ALL">Semua Salesman</option>
          <option v-for="e in salesEmployees" :key="e.id" :value="e.id">
            {{ e.first_name }} {{ e.last_name }} ({{ e.job_position?.title || 'Sales' }})
          </option>
        </select>
        <select
          v-model="selectedStatus"
          class="bg-white border border-slate-200 rounded-2xl px-4 py-3.5 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20 shadow-sm min-w-[160px]"
        >
          <option value="ALL">Semua Status</option>
          <option v-for="s in statusOptions.filter((x) => x !== 'ALL')" :key="s" :value="s">
            {{ s }}
          </option>
        </select>
      </div>
    </div>

    <!-- Summary Metrics / Counters Dinamis -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- 1. Disetujui Admin Nota (Main Highlight) -->
      <div class="bg-gradient-to-br from-emerald-600 via-emerald-600 to-teal-700 text-white rounded-3xl p-5 shadow-lg shadow-emerald-600/20 relative overflow-hidden flex flex-col justify-between">
        <div class="absolute -right-4 -bottom-4 w-24 h-24 bg-white/10 rounded-full blur-xl pointer-events-none"></div>
        <div class="flex items-center justify-between">
          <span class="text-[10px] font-black uppercase tracking-wider text-emerald-100">Disetujui Admin Nota</span>
          <div class="w-8 h-8 rounded-xl bg-white/20 flex items-center justify-center">
            <CheckCircle2 class="w-4 h-4 text-white" />
          </div>
        </div>
        <div class="mt-3">
          <p class="text-xl sm:text-2xl font-black tracking-tight">
            {{ formatCurrency(totalApprovedAmount) }}
          </p>
          <p class="text-xs font-bold text-emerald-100 mt-0.5 flex items-center gap-1.5">
            <span class="inline-block w-2 h-2 rounded-full bg-emerald-300 animate-pulse"></span>
            {{ totalApprovedCount }} Pesanan Siap Kirim / Selesai
          </p>
        </div>
      </div>

      <!-- 2. Menunggu Konfirmasi (Pending) -->
      <div class="bg-white border border-amber-200/80 rounded-3xl p-5 shadow-sm flex flex-col justify-between hover:border-amber-300 transition-colors">
        <div class="flex items-center justify-between">
          <span class="text-[10px] font-black uppercase tracking-wider text-amber-700">Menunggu Konfirmasi</span>
          <div class="w-8 h-8 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center border border-amber-200">
            <Clock class="w-4 h-4" />
          </div>
        </div>
        <div class="mt-3">
          <p class="text-xl sm:text-2xl font-black text-slate-900 tracking-tight">
            {{ formatCurrency(totalPendingAmount) }}
          </p>
          <p class="text-xs font-bold text-amber-600 mt-0.5">
            {{ totalPendingCount }} Pesanan Baru / DRAFT
          </p>
        </div>
      </div>

      <!-- 3. Ditolak / Dibatalkan -->
      <div class="bg-white border border-slate-200 rounded-3xl p-5 shadow-sm flex flex-col justify-between hover:border-red-200 transition-colors">
        <div class="flex items-center justify-between">
          <span class="text-[10px] font-black uppercase tracking-wider text-slate-500">Ditolak / Batal</span>
          <div class="w-8 h-8 rounded-xl bg-red-50 text-red-600 flex items-center justify-center border border-red-100">
            <XCircle class="w-4 h-4" />
          </div>
        </div>
        <div class="mt-3">
          <p class="text-xl sm:text-2xl font-black text-slate-900 tracking-tight">
            {{ formatCurrency(totalRejectedAmount) }}
          </p>
          <p class="text-xs font-bold text-red-500 mt-0.5">
            {{ totalRejectedCount }} Pesanan Ditolak / Batal
          </p>
        </div>
      </div>

      <!-- 4. Grand Total Tampil Sesuai Filter -->
      <div class="bg-slate-900 text-white rounded-3xl p-5 shadow-lg shadow-slate-900/10 flex flex-col justify-between">
        <div class="flex items-center justify-between">
          <span class="text-[10px] font-black uppercase tracking-wider text-slate-400">Total Nilai Tampil</span>
          <div class="w-8 h-8 rounded-xl bg-slate-800 text-slate-300 flex items-center justify-center">
            <Banknote class="w-4 h-4" />
          </div>
        </div>
        <div class="mt-3">
          <p class="text-xl sm:text-2xl font-black tracking-tight">
            {{ formatCurrency(totalFilteredAmount) }}
          </p>
          <p class="text-xs font-bold text-slate-400 mt-0.5">
            {{ totalFilteredCount }} Total Pesanan Sesuai Filter
          </p>
        </div>
      </div>
    </div>

    <!-- Backorder / Waiting Stock Alert Card if any -->
    <div
      v-if="waitingStockOrders.length > 0 && selectedStatus !== 'WAITING_STOCK'"
      class="bg-orange-50/90 border border-orange-200 rounded-3xl p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4 shadow-sm"
    >
      <div class="flex items-center gap-3.5">
        <div class="w-10 h-10 rounded-2xl bg-orange-600 text-white flex items-center justify-center shrink-0 shadow-md shadow-orange-600/20">
          <AlertTriangle class="w-5 h-5" />
        </div>
        <div>
          <div class="flex items-center gap-2 flex-wrap">
            <span class="text-sm font-black text-orange-950">
              Terdapat {{ waitingStockOrders.length }} Pesanan Menunggu Ketersediaan Stok (Backorder)
            </span>
            <Badge class="bg-orange-600 text-white border-0 text-[10px] font-black tracking-wider uppercase">
              {{ formatCurrency(totalWaitingStockAmount) }}
            </Badge>
          </div>
          <p class="text-xs text-orange-800/80 mt-0.5 font-medium">
            Pesanan tertahan karena sebagian produk belum ready di gudang cabang. Klik untuk memfilter dan memantau pemenuhan stok.
          </p>
        </div>
      </div>
      <Button
        @click="selectedStatus = 'WAITING_STOCK'"
        size="sm"
        class="rounded-xl font-black text-xs bg-orange-600 hover:bg-orange-700 text-white shadow-sm shrink-0 px-4 h-9"
      >
        Lihat Pesanan Menunggu Stok
      </Button>
    </div>

    <!-- Orders List -->
    <div
      v-if="soStore.loading"
      class="text-center py-20 text-slate-400 flex flex-col items-center gap-3"
    >
      <RefreshCw class="w-8 h-8 animate-spin" />
      <p class="text-sm font-bold">Memuat pesanan...</p>
    </div>

    <div
      v-else-if="filteredOrders.length === 0"
      class="text-center py-20 bg-slate-50 rounded-3xl border-2 border-dashed border-slate-200"
    >
      <ClipboardList class="w-12 h-12 text-slate-300 mx-auto mb-3" />
      <p class="font-black text-slate-500">Tidak ada pesanan ditemukan</p>
      <p class="text-sm text-slate-400 mt-1">Buat pesanan baru atau ubah filter status</p>
    </div>

    <div v-else class="bg-white border border-slate-200 rounded-[2.5rem] shadow-sm overflow-hidden">
      <Table class="w-full">
        <TableHeader class="bg-slate-50/70 border-b border-slate-100">
          <TableRow>
            <TableHead class="font-black text-[10px] uppercase tracking-widest pl-6 w-[18%]">
              No. SO & Perusahaan
            </TableHead>
            <TableHead class="font-black text-[10px] uppercase tracking-widest w-[20%]">
              Toko & Lokasi GPS
            </TableHead>
            <TableHead class="font-black text-[10px] uppercase tracking-widest w-[13%]">
              Salesman
            </TableHead>
            <TableHead class="font-black text-[10px] uppercase tracking-widest w-[19%]">
              Item & Promosi
            </TableHead>
            <TableHead class="font-black text-[10px] uppercase tracking-widest text-right w-[11%]">
              Total Tagihan
            </TableHead>
            <TableHead class="font-black text-[10px] uppercase tracking-widest text-center w-[9%]">
              Status
            </TableHead>
            <TableHead class="font-black text-[10px] uppercase tracking-widest text-right pr-6 w-[10%]">
              Aksi
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody class="divide-y divide-slate-100">
          <TableRow
            v-for="order in filteredOrders"
            :key="order.id"
            class="group hover:bg-slate-50/60 transition-colors"
          >
            <!-- 1. NO SO & PERUSAHAAN -->
            <TableCell class="pl-6 py-4 align-top">
              <div class="flex flex-col space-y-1">
                <div class="flex flex-wrap items-center gap-1.5">
                  <span class="font-mono text-xs font-black text-slate-800 tracking-tight">
                    {{ order.so_number }}
                  </span>
                  <Badge
                    variant="outline"
                    class="text-[9px] font-black uppercase tracking-wider px-1.5 py-0.5 border-slate-200 bg-slate-100 text-slate-700 rounded shadow-none"
                  >
                    {{ getCompanyName(order.company_id) }}
                  </Badge>
                  <span v-if="order.is_mt" class="bg-violet-100 text-violet-700 font-extrabold text-[8px] px-1.5 py-0.5 rounded tracking-wide uppercase">MT</span>
                  <span v-if="order.is_physical_nota" class="bg-amber-100 text-amber-700 font-extrabold text-[8px] px-1.5 py-0.5 rounded tracking-wide uppercase flex items-center gap-0.5"><FileText class="w-3 h-3"/> FISIK</span>
                  <span v-if="order.tax_invoice_number" class="bg-blue-100 text-blue-700 font-extrabold text-[8px] px-1.5 py-0.5 rounded tracking-wide flex items-center gap-0.5" title="Faktur Pajak">FP: {{ order.tax_invoice_number }}</span>
                </div>
                <div class="flex items-center text-[11px] font-bold text-slate-500 gap-1.5">
                  <Clock class="w-3 h-3 text-slate-400" />
                  <span>{{ formatDate(order.order_date) }}</span>
                </div>
              </div>
            </TableCell>

            <!-- 2. TOKO & LOKASI GPS -->
            <TableCell class="py-4 align-top">
              <div class="flex flex-col space-y-1.5">
                <div class="flex items-center gap-2">
                  <span class="font-black text-slate-900 text-sm leading-tight">
                    {{ order.store?.name || 'Unknown Store' }}
                  </span>
                </div>

                <div class="flex flex-wrap items-center gap-1.5">
                  <span
                    class="text-[9px] font-extrabold px-2 py-0.5 rounded-full uppercase tracking-wider"
                    :class="order.store_category === 'TOKO_BARU' ? 'bg-indigo-50 text-indigo-700 border border-indigo-100' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ order.store_category ? order.store_category.replace('_', ' ') : 'TOKO' }}
                  </span>

                  <!-- GPS Location Radius Badge -->
                  <Badge
                    v-if="getOrderLocationStatus(order).status !== 'UNKNOWN'"
                    :class="`${getOrderLocationStatus(order).bg} ${getOrderLocationStatus(order).color} text-[9px] font-bold py-0.5 px-2 border-0 hover:bg-opacity-80 cursor-pointer shadow-none flex items-center gap-1`"
                    @click="openMapModal(order)"
                    title="Klik untuk memverifikasi lokasi di Peta"
                  >
                    <component :is="getOrderLocationStatus(order).icon" class="w-3 h-3" />
                    <span>{{ getOrderLocationStatus(order).text }}</span>
                  </Badge>
                  <span v-else class="text-[10px] text-slate-400 font-medium italic">
                    Lokasi: Tidak ada data
                  </span>
                </div>
              </div>
            </TableCell>

            <!-- 3. SALESMAN -->
            <TableCell class="py-4 align-top">
              <div class="flex flex-col">
                <span class="font-black text-slate-800 text-xs leading-tight">
                  {{ order.employee?.first_name }} {{ order.employee?.last_name }}
                </span>
                <span v-if="order.employee?.job_position?.title" class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mt-0.5">
                  {{ order.employee.job_position.title }}
                </span>
              </div>
            </TableCell>

            <!-- 4. ITEM & PROMOSI (Item + Bonus + Diskon) -->
            <TableCell class="py-4 align-top">
              <div class="flex flex-col space-y-1.5">
                <!-- Products list -->
                <div class="flex flex-wrap gap-1">
                  <Badge
                    variant="outline"
                    v-for="item in (expandedItemRowIds.includes(order.id) ? order.items : order.items.slice(0, 2))"
                    :key="item.id"
                    :class="`${getItemStockBreakdown(item).hasBackorder && (order.status === 'WAITING_STOCK' || order.status === 'DRAFT') ? 'border-amber-300 bg-amber-50/90 text-amber-900' : 'border-slate-200 bg-white text-slate-800'} text-[9px] font-bold px-2 py-0.5 h-auto shadow-none flex items-center gap-1`"
                    :title="getItemStockBreakdown(item).hasBackorder ? `Kekurangan: ${getItemStockBreakdown(item).backorderPieces} PCS (Menunggu Pasokan)` : 'Stok Tersedia'"
                  >
                    <span class="truncate max-w-[110px]">{{ item.product?.name?.split(' ')[0] }}...</span>
                    <span>x{{ item.ordered_quantity }}</span>
                    <span
                      v-if="getItemStockBreakdown(item).hasBackorder && (order.status === 'WAITING_STOCK' || order.status === 'DRAFT')"
                      class="text-[8px] font-black text-amber-800 bg-amber-200/80 px-1 py-0.2 rounded"
                    >
                      ⏳ -{{ getItemStockBreakdown(item).backorderPieces }}
                    </span>
                  </Badge>
                  <button
                    v-if="order.items.length > 2"
                    class="text-[9px] font-extrabold text-primary ml-1 hover:underline transition-colors cursor-pointer outline-none"
                    @click.stop="toggleExpand(order.id)"
                  >
                    {{ expandedItemRowIds.includes(order.id) ? 'Sembunyikan' : `+${order.items.length - 2} more` }}
                  </button>
                </div>

                <!-- Backorder Shortage Summary Tag if WAITING_STOCK -->
                <div v-if="order.status === 'WAITING_STOCK'" class="pt-0.5">
                  <button
                    @click.stop="openWaitingStockModal(order)"
                    class="text-[9px] font-extrabold text-orange-800 bg-orange-100 hover:bg-orange-200 border border-orange-200 px-2 py-0.5 rounded-md inline-flex items-center gap-1 shadow-none transition-colors cursor-pointer text-left"
                    title="Klik untuk membuka Analisis & Rincian Stok Backorder"
                  >
                    <AlertTriangle class="w-3 h-3 text-orange-600 shrink-0" />
                    <span>Menunggu Stok: {{ getSOBackorderSummary(order)?.totalBackorderItems || 0 }} item defisit</span>
                  </button>
                </div>

                <!-- Bonus & Discount Tags -->
                <div class="flex flex-wrap items-center gap-1.5 pt-0.5" v-if="(order.bonus_items && order.bonus_items.length > 0) || (order.discount_amount || 0) > 0">
                  <!-- Bonus Badge -->
                  <template v-if="order.bonus_items && order.bonus_items.length > 0">
                    <Badge
                      v-for="b in (expandedItemRowIds.includes(order.id) ? order.bonus_items : order.bonus_items.slice(0, 1))"
                      :key="b.id"
                      variant="secondary"
                      class="text-[9px] font-extrabold px-1.5 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200/80 shadow-none"
                    >
                      🎁 {{ b.bonus_inventory?.name?.split(' ')[0] || 'Bonus' }} x{{ b.quantity }}
                    </Badge>
                  </template>

                  <!-- Diskon Tag -->
                  <span
                    v-if="(order.discount_amount || 0) > 0"
                    class="font-black text-[9px] text-rose-600 bg-rose-50 border border-rose-100 px-1.5 py-0.5 rounded"
                  >
                    🏷️ -{{ formatCurrency(order.discount_amount || 0) }}
                  </span>
                </div>
              </div>
            </TableCell>

            <!-- 5. TOTAL TAGIHAN -->
            <TableCell class="text-right py-4 align-top">
              <div class="flex flex-col items-end">
                <span class="font-black text-slate-900 text-sm tracking-tight">
                  {{ formatCurrency(order.total_amount) }}
                </span>
                <span v-if="(order.discount_amount || 0) > 0" class="text-[9px] font-bold text-rose-500">
                  Hemat {{ formatCurrency(order.discount_amount || 0) }}
                </span>
              </div>
            </TableCell>

            <!-- 6. STATUS -->
            <TableCell class="py-4 align-top text-center">
              <div class="flex flex-col items-center">
                <span
                  @click="order.status === 'WAITING_STOCK' ? openWaitingStockModal(order) : undefined"
                  :class="[
                    order.deleted_at ? statusColor['DELETED'] : statusColor[order.status],
                    order.status === 'WAITING_STOCK' ? 'cursor-pointer hover:ring-2 hover:ring-orange-300' : '',
                    'text-[9px] font-black px-2.5 py-1 rounded-full uppercase tracking-wider shadow-sm inline-block transition-all',
                  ]"
                  :title="order.status === 'WAITING_STOCK' ? 'Klik untuk analisis detail stok & backorder' : ''"
                >
                  {{ order.deleted_at ? 'TERHAPUS' : order.status }}
                </span>
                <div
                  v-if="order.deleted_at"
                  class="mt-0.5 text-[8px] text-rose-600 font-semibold max-w-[100px] truncate"
                  :title="`Dihapus: ${order.deleted_by?.first_name || 'Admin'}\nAlasan: ${order.deletion_reason || '-'}`"
                >
                  {{ order.deletion_reason || 'Dihapus' }}
                </div>
              </div>
            </TableCell>

            <!-- 7. AKSI -->
            <TableCell class="text-right pr-6 py-4 align-top">
              <div class="flex items-center justify-end gap-1.5">
                <!-- Restore button if deleted -->
                <Button
                  v-if="order.deleted_at || selectedStatus === 'DELETED'"
                  size="sm"
                  @click="handleRestore(order)"
                  class="h-7 px-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-[9px] font-black shadow-sm flex items-center gap-1"
                  title="Pulihkan Pesanan Ini"
                >
                  <RotateCcw class="w-3 h-3" /> PULIHKAN
                </Button>

                <template v-else>
                  <!-- Super Admin Fast-Track Complete Button -->
                  <Button
                    v-if="canSuperAdminComplete(order)"
                    size="sm"
                    @click="handleSuperAdminComplete(order)"
                    class="h-7 px-2 rounded-lg bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-white text-[9px] font-black shadow-sm flex items-center gap-1 border border-amber-400/30 shrink-0"
                    title="⚡ Super Admin: Selesaikan Pesanan Ini Secara Instan (Pengiriman + Faktur + Kasir + Omzet)"
                  >
                    <Sparkles class="w-3 h-3 text-amber-100" /> AUTO COMPLETE
                  </Button>

                  <Button
                    variant="ghost"
                    size="sm"
                    @click="openDetail(order)"
                    class="h-7 w-7 p-0 rounded-lg hover:bg-primary/10 hover:text-primary transition-all text-slate-500"
                    title="Lihat Detail Pesanan"
                  >
                    <FileText class="w-3.5 h-3.5" />
                  </Button>

                  <!-- Waiting Stock Diagnostic Modal Button -->
                  <Button
                    v-if="order.status === 'WAITING_STOCK'"
                    variant="ghost"
                    size="sm"
                    @click="openWaitingStockModal(order)"
                    class="h-7 w-7 p-0 rounded-lg hover:bg-orange-50 hover:text-orange-600 transition-all text-orange-500"
                    title="Buka Analisis & Rincian Stok Backorder"
                  >
                    <AlertTriangle class="w-3.5 h-3.5" />
                  </Button>

                  <!-- Adjust button for DRAFT, CONFIRMED, and WAITING_STOCK -->
                  <Button
                    v-if="order.status === 'DRAFT' || order.status === 'CONFIRMED' || order.status === 'WAITING_STOCK'"
                    variant="ghost"
                    size="sm"
                    @click="openAdjust(order)"
                    class="h-7 w-7 p-0 rounded-lg hover:bg-violet-50 hover:text-violet-600 transition-all text-slate-500"
                    title="Adjust pesanan (diskon, qty, piutang)"
                  >
                    <SlidersHorizontal class="w-3.5 h-3.5" />
                  </Button>

                  <!-- Quick Recheck Stock for WAITING_STOCK -->
                  <Button
                    v-if="order.status === 'WAITING_STOCK'"
                    variant="ghost"
                    size="sm"
                    @click="handleRecheckStock(order)"
                    :disabled="isRecheckingStock"
                    class="h-7 w-7 p-0 rounded-lg hover:bg-orange-50 hover:text-orange-600 transition-all text-orange-500"
                    title="Cek ulang ketersediaan stok gudang (Sinkronisasi Backorder)"
                  >
                    <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': isRecheckingStock }" />
                  </Button>

                  <template v-if="order.status === 'DRAFT' || order.status === 'WAITING_TAX_DOCS'">
                    <!-- Tombol khusus Admin Pajak untuk Modern Trade (MT) -->
                    <Button
                      v-if="order.is_mt"
                      size="sm"
                      @click="openTaxApproval(order)"
                      class="h-7 px-2.5 rounded-lg bg-violet-600 hover:bg-violet-700 text-white text-[9px] font-black shadow-sm flex items-center gap-1"
                    >
                      <FileText class="w-3 h-3" /> APPROVE PAJAK
                    </Button>

                    <Button
                      v-else
                      size="sm"
                      @click="handleAdminConfirm(order.id)"
                      class="h-7 px-2.5 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-[9px] font-black shadow-sm"
                    >
                      APPROVE
                    </Button>

                    <Button
                      variant="ghost"
                      size="sm"
                      @click="openReject(order.id, true)"
                      class="h-7 w-7 p-0 rounded-lg text-red-400 hover:bg-red-50 hover:text-red-600"
                      title="Tolak Pesanan"
                    >
                      <XCircle class="w-3.5 h-3.5" />
                    </Button>
                  </template>

                  <Button
                    v-if="order.status !== 'DRAFT' && order.status !== 'CANCELLED' && order.status !== 'REJECTED' && order.status !== 'WAITING_TAX_DOCS'"
                    variant="ghost"
                    size="sm"
                    @click="openSetFaktur(order)"
                    class="h-7 w-7 p-0 rounded-lg text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700"
                    title="Set Faktur Pajak"
                  >
                    <FileText class="w-3.5 h-3.5" />
                  </Button>

                  <template v-if="order.status === 'SHIPPED'">
                    <Button
                      size="sm"
                      @click="handleConvert(order.id)"
                      class="h-7 px-2.5 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white text-[9px] font-black shadow-sm flex items-center gap-1"
                    >
                      <FileText class="w-3 h-3" /> INVOICE
                    </Button>
                  </template>

                  <Button
                    v-if="canConvertDirect(order)"
                    size="sm"
                    @click="handleConvertDirect(order)"
                    class="h-7 px-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-[9px] font-black shadow-sm flex items-center gap-1"
                    title="Terbitkan faktur langsung tanpa pengiriman untuk SO motoris"
                  >
                    <FileText class="w-3 h-3" /> LANGSUNG KE FAKTUR
                  </Button>

                  <template
                    v-if="
                      order.status === 'CONFIRMED' ||
                      order.status === 'WAITING_WAREHOUSE' ||
                      order.status === 'WAITING_STOCK'
                    "
                  >
                    <Button
                      variant="ghost"
                      size="sm"
                      @click="handleCancel(order.id)"
                      class="h-7 w-7 p-0 rounded-lg text-slate-400 hover:bg-red-50 hover:text-red-600"
                      title="Batalkan Pesanan"
                    >
                      <XCircle class="w-3.5 h-3.5" />
                    </Button>
                  </template>

                  <Button
                    v-if="order.status === 'DRAFT' || order.status === 'REJECTED'"
                    variant="ghost"
                    size="sm"
                    @click="handleDelete(order)"
                    class="h-7 w-7 p-0 rounded-lg text-slate-300 hover:bg-red-50 hover:text-red-600"
                    title="Hapus (Soft Delete)"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </Button>
                </template>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- Pagination -->
    <div class="flex items-center justify-between mt-4">
      <div class="flex items-center gap-2 text-sm text-slate-500">
        <span>Tampilkan</span>
        <select
          v-model="itemsPerPage"
          class="border border-slate-200 rounded-md px-2 py-1 outline-none focus:border-primary focus:ring-1 focus:ring-primary"
          @change="currentPage = 1"
        >
          <option :value="10">10</option>
          <option :value="20">20</option>
          <option :value="50">50</option>
          <option :value="0">Semua</option>
        </select>
        <span>data per halaman. Total: {{ soStore.totalOrders }} data.</span>
      </div>

      <div class="flex items-center gap-2" v-if="itemsPerPage > 0">
        <Button
          variant="outline"
          size="sm"
          :disabled="currentPage === 1"
          @click="currentPage--"
        >
          Sebelumnya
        </Button>
        <span class="text-sm font-medium mx-2">
          Halaman {{ currentPage }} dari {{ Math.ceil(soStore.totalOrders / itemsPerPage) || 1 }}
        </span>
        <Button
          variant="outline"
          size="sm"
          :disabled="currentPage >= Math.ceil(soStore.totalOrders / itemsPerPage)"
          @click="currentPage++"
        >
          Selanjutnya
        </Button>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showCreateModal = false"
      />
      <div
        class="bg-white rounded-[32px] w-full max-w-xl relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden"
      >
        <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between">
          <div>
            <h3 class="text-xl font-black text-slate-900">Buat Pesanan Baru</h3>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-0.5">
              {{ form.flow_type === 'TO' ? 'Stok belum dipotong hingga faktur diterbitkan' : (form.flow_type === 'DIRECT' ? 'Stok langsung memotong kendaraan salesman' : 'Pesanan langsung dari kantor cabang (Omzet Office)') }}
            </p>
          </div>
          <button
            @click="showCreateModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-6 overflow-y-auto space-y-5">
          <!-- Tipe Alur Transaksi -->
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Tipe Alur Transaksi</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                @click="form.flow_type = 'TO'"
                :class="[
                  'px-3 py-3 text-[10px] font-black rounded-2xl border transition-all text-center uppercase tracking-wider',
                  form.flow_type === 'TO'
                    ? 'bg-primary border-primary text-white shadow-md shadow-primary/20'
                    : 'bg-slate-50 border-slate-100 text-slate-500 hover:bg-slate-100'
                ]"
              >
                Sales TO
              </button>
              <button
                type="button"
                @click="form.flow_type = 'DIRECT'"
                :class="[
                  'px-3 py-3 text-[10px] font-black rounded-2xl border transition-all text-center uppercase tracking-wider',
                  form.flow_type === 'DIRECT'
                    ? 'bg-primary border-primary text-white shadow-md shadow-primary/20'
                    : 'bg-slate-50 border-slate-100 text-slate-500 hover:bg-slate-100'
                ]"
              >
                Motoris (Direct)
              </button>
              <button
                type="button"
                @click="form.flow_type = 'OFFICE'"
                :class="[
                  'px-3 py-3 text-[10px] font-black rounded-2xl border transition-all text-center uppercase tracking-wider',
                  form.flow_type === 'OFFICE'
                    ? 'bg-primary border-primary text-white shadow-md shadow-primary/20'
                    : 'bg-slate-50 border-slate-100 text-slate-500 hover:bg-slate-100'
                ]"
              >
                Omzet Office
              </button>
            </div>
          </div>

          <!-- Tanggal Pesanan -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Tanggal Pesanan</label>
              <span class="text-[9px] font-medium text-slate-400">Dapat diisi tanggal hari ini atau tanggal lampau</span>
            </div>
            <input
              v-model="form.order_date"
              type="date"
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-3.5 font-bold text-slate-900 focus:bg-white focus:border-primary/20 outline-none"
            />
          </div>

          <!-- Salesman -->
          <div v-if="form.flow_type !== 'OFFICE'" class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Salesman</label>
            <select
              v-model="form.employee_id"
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-3.5 font-bold text-slate-900 focus:bg-white focus:border-primary/20 outline-none appearance-none"
            >
              <option value="">Pilih Salesman</option>
              <option v-for="e in salesEmployees" :key="e.id" :value="e.id">
                {{ e.first_name }} {{ e.last_name }} ({{ e.job_position?.title || 'Sales' }})
              </option>
            </select>
          </div>

          <!-- Toko -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Toko</label>
              <button
                type="button"
                @click="openCreateStoreModal"
                class="text-[10px] font-black text-primary flex items-center gap-1 hover:underline"
              >
                <Plus class="w-3 h-3" /> Tambah Toko Baru
              </button>
            </div>
            <SearchSelect
              v-model="form.store_id"
              :options="storeOptions"
              placeholder="Pilih Toko"
              search-placeholder="Cari nama toko atau alamat..."
              :clearable="true"
              size="lg"
              @change="handleStoreChange"
            />
          </div>

          <!-- Kategori -->
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Kategori Toko</label>
            <select
              v-model="form.store_category"
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-3.5 font-bold text-slate-900 focus:bg-white focus:border-primary/20 outline-none appearance-none"
            >
              <option value="TOKO_LAMA">Toko Lama</option>
              <option value="TOKO_BARU">Toko Baru</option>
            </select>
          </div>

          <!-- Metode Pembayaran & Detail (Only for DIRECT Sales Motoris / Booster) -->
          <div v-if="form.flow_type === 'DIRECT'" class="space-y-4 bg-slate-50 p-5 rounded-[2rem] border border-slate-100">
            <h4 class="text-xs font-black text-slate-600 uppercase tracking-wider mb-2">Pengaturan Pembayaran Direct</h4>
            <div class="grid grid-cols-2 gap-3">
              <div class="space-y-1">
                <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Metode</label>
                <select
                  v-model="form.payment_method"
                  class="w-full bg-white border-0 rounded-xl px-3 py-2 text-xs font-bold focus:ring-2 focus:ring-primary/20 outline-none"
                >
                  <option value="CASH">CASH / TUNAI</option>
                  <option value="TEMPO">TEMPO / KREDIT</option>
                </select>
              </div>
              <div v-if="form.payment_method === 'TEMPO'" class="space-y-1">
                <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Tenor (Minggu)</label>
                <input
                  v-model.number="form.tenor"
                  type="number"
                  min="1"
                  class="w-full bg-white border-0 rounded-xl px-3 py-2 text-xs font-bold focus:ring-2 focus:ring-primary/20 outline-none"
                />
              </div>
            </div>
            <div v-if="form.payment_method === 'TEMPO'" class="space-y-1 mt-2">
              <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Uang Muka / DP (Rp)</label>
              <input
                v-model.number="form.paid_amount"
                type="number"
                min="0"
                class="w-full bg-white border-0 rounded-xl px-3 py-2 text-xs font-bold focus:ring-2 focus:ring-primary/20 outline-none"
              />
            </div>
          </div>

          <!-- Items -->
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
                >Item Pesanan</label
              >
              <button
                @click="addItem"
                class="text-[10px] font-black text-primary flex items-center gap-1 hover:underline"
              >
                <Plus class="w-3 h-3" /> Tambah
              </button>
            </div>
            <div
              v-for="(item, idx) in form.items"
              :key="idx"
              class="space-y-2 bg-slate-50 p-4 rounded-3xl border border-slate-100"
            >
              <div class="flex items-center gap-2">
                <ProductSearchSelect
                  v-model="item.product_id"
                  :products="factoryStore.sellableProducts"
                  :get-stock="getProductStock"
                  @change="() => onProductChange(idx)"
                  class="flex-1"
                />
                <button
                  @click="removeItem(idx)"
                  class="p-2.5 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-xl transition-colors"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>

              <div class="grid grid-cols-4 gap-2">
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Jumlah</label>
                  <input
                    v-model.number="item.ordered_quantity"
                    type="number"
                    min="1"
                    class="w-full bg-white border-0 rounded-xl px-3 py-2 text-xs font-bold focus:ring-2 focus:ring-primary/20 outline-none"
                  />
                </div>
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Satuan</label>
                  <select
                    v-model="item.unit"
                    class="w-full bg-white border-0 rounded-xl px-2 py-2 text-[10px] font-bold focus:ring-2 focus:ring-primary/20 outline-none"
                  >
                    <option v-for="u in UNIT_OPTIONS" :key="u" :value="u">{{ u }}</option>
                  </select>
                </div>
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1"
                    >Isi/Unit</label
                  >
                  <input
                    v-model.number="item.pieces_per_unit"
                    type="number"
                    min="1"
                    class="w-full bg-white border-0 rounded-xl px-3 py-2 text-xs font-bold focus:ring-2 focus:ring-primary/20 outline-none"
                  />
                </div>
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1"
                    >Harga Satuan Dasar</label
                  >
                  <div class="relative">
                    <span
                      class="absolute left-2 top-1/2 -translate-y-1/2 text-[9px] text-slate-400 font-bold"
                      >Rp</span
                    >
                    <input
                      v-model.number="item.price"
                      type="number"
                      class="w-full bg-white border-0 rounded-xl pl-6 pr-2 py-2 text-[10px] font-bold focus:ring-2 focus:ring-primary/20 outline-none"
                    />
                  </div>
                </div>
              </div>
              <div
                v-if="item.ordered_quantity > 1 || item.pieces_per_unit > 1"
                class="px-2 pt-1 flex justify-between items-center text-[9px] font-bold text-slate-400"
              >
                <span>Total: {{ item.ordered_quantity * item.pieces_per_unit }} PCS</span>
                <span
                  >Subtotal:
                  {{
                    formatCurrency(item.ordered_quantity * item.pieces_per_unit * item.price)
                  }}</span
                >
              </div>
            </div>
          </div>

          <div class="bg-slate-50 rounded-2xl px-5 py-3 flex justify-between items-center">
            <span class="text-xs font-black text-slate-500 uppercase tracking-wider"
              >Total Nilai Pesanan</span
            >
            <span class="text-lg font-black text-primary">{{ formatCurrency(itemTotal) }}</span>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
              >Catatan</label
            >
            <textarea
              v-model="form.notes"
              rows="2"
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-5 py-3 font-medium text-sm focus:bg-white focus:border-primary/20 outline-none resize-none"
            />
          </div>
        </div>

        <div class="p-5 border-t border-slate-100 bg-slate-50/50 flex gap-3 justify-end">
          <Button
            variant="outline"
            @click="showCreateModal = false"
            class="rounded-2xl font-bold px-6"
            >Batal</Button
          >
          <Button
            @click="handleCreate"
            :disabled="loading"
            class="rounded-2xl font-bold px-8 shadow-lg shadow-primary/20"
          >
            {{ loading ? 'Menyimpan...' : 'Buat Pesanan' }}
          </Button>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <div
      v-if="showDetailModal && selectedOrder"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
    >
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showDetailModal = false"
      />
      <div
        class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden"
      >
        <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between">
          <div>
            <h3 class="text-xl font-black text-slate-900">{{ selectedOrder.so_number }}</h3>
            <span
              :class="[
                statusColor[selectedOrder.status],
                'text-[10px] font-black px-3 py-1 rounded-lg uppercase tracking-wider mt-1 inline-block',
              ]"
              >{{ selectedOrder.status }}</span
            >
          </div>
          <button
            @click="showDetailModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 overflow-y-auto space-y-4">
          <!-- Location Radius Verification inside SO Detail Modal -->
          <div 
            v-if="getOrderLocationStatus(selectedOrder).status !== 'UNKNOWN'"
            class="p-3.5 rounded-2xl flex items-center justify-between gap-3 border shadow-sm"
            :class="getOrderLocationStatus(selectedOrder).status === 'OUTSIDE' ? 'bg-red-50/90 border-red-200' : 'bg-emerald-50/90 border-emerald-200'"
          >
            <div class="flex items-center gap-2.5">
              <div 
                :class="[
                  'p-2 rounded-xl text-white shrink-0 shadow-sm',
                  getOrderLocationStatus(selectedOrder).status === 'OUTSIDE' ? 'bg-red-600' : 'bg-emerald-600'
                ]"
              >
                <component :is="getOrderLocationStatus(selectedOrder).icon" class="w-4 h-4" />
              </div>
              <div>
                <div class="flex items-center gap-1.5">
                  <span class="text-xs font-black text-slate-900">Radius Lokasi</span>
                  <Badge 
                    :class="`${getOrderLocationStatus(selectedOrder).bg} ${getOrderLocationStatus(selectedOrder).color} border-0 text-[10px] font-black uppercase tracking-wider`"
                  >
                    {{ getOrderLocationStatus(selectedOrder).text }}
                  </Badge>
                </div>
                <p class="text-[11px] text-slate-500 mt-0.5">
                  {{ getOrderLocationStatus(selectedOrder).status === 'OUTSIDE' ? 'Pesanan dibuat di luar radius 500m dari toko.' : 'Pesanan sesuai radius toko.' }}
                </p>
              </div>
            </div>
            <Button 
              @click="openMapModal(selectedOrder)"
              size="sm"
              variant="outline"
              class="h-8 px-3 rounded-xl font-bold text-xs bg-white hover:bg-slate-100 border-slate-300 text-slate-700 shadow-sm shrink-0 flex items-center gap-1"
            >
              <Navigation class="w-3.5 h-3.5 text-primary" />
              Lihat Peta
            </Button>
          </div>

          <!-- Waiting Stock Status Alert Banner in Detail Modal -->
          <div 
            v-if="selectedOrder.status === 'WAITING_STOCK'"
            class="p-4 rounded-2xl bg-orange-50 border border-orange-200 flex flex-col gap-3 shadow-sm"
          >
            <div class="flex items-start gap-3">
              <div class="p-2 rounded-xl bg-orange-600 text-white shrink-0 shadow-sm mt-0.5">
                <AlertTriangle class="w-4 h-4" />
              </div>
              <div class="flex-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="text-xs font-black text-orange-950">Status: Menunggu Ketersediaan Stok</span>
                  <Badge class="bg-orange-600 text-white border-0 text-[10px] font-black uppercase tracking-wider">
                    Backorder
                  </Badge>
                </div>
                <p class="text-[11px] text-orange-800 mt-1 leading-relaxed">
                  Pesanan tertahan karena terdapat produk yang kuantitasnya melebihi stok gudang cabang. Sistem otomatis mengalokasikan stok baru saat barang masuk (FIFO), atau Anda dapat klik <strong>Cek Ulang Stok</strong> sekarang.
                </p>
              </div>
            </div>
            <div class="flex items-center justify-end gap-2 pt-1 border-t border-orange-200/60 flex-wrap">
              <Button
                @click="openWaitingStockModal(selectedOrder)"
                size="sm"
                variant="outline"
                class="h-8 px-3 rounded-xl font-bold text-xs bg-white hover:bg-orange-50 text-orange-700 border-orange-300 shadow-sm flex items-center gap-1.5"
              >
                <Layers class="w-3.5 h-3.5" />
                <span>Analisis Lengkap Stok</span>
              </Button>
              <Button
                @click="handleRecheckStock(selectedOrder)"
                :disabled="isRecheckingStock"
                size="sm"
                class="h-8 px-3 rounded-xl font-bold text-xs bg-orange-600 hover:bg-orange-700 text-white shadow-sm flex items-center gap-1.5"
              >
                <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': isRecheckingStock }" />
                <span>{{ isRecheckingStock ? 'Memeriksa...' : 'Cek Ulang Stok' }}</span>
              </Button>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Toko</p>
              <p class="font-bold text-slate-900 mt-1">{{ selectedOrder.store?.name }}</p>
            </div>
            <div>
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Salesman
              </p>
              <p class="font-bold text-slate-900 mt-1">
                {{ selectedOrder.employee?.first_name }} {{ selectedOrder.employee?.last_name }}
              </p>
            </div>
            <div>
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Tanggal Pesanan
              </p>
              <p class="font-bold text-slate-900 mt-1">
                {{ formatDate(selectedOrder.order_date) }}
              </p>
            </div>
            <div v-if="selectedOrder.confirmed_at">
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Dikonfirmasi Oleh
              </p>
              <p class="font-bold text-slate-900 mt-1">
                {{ selectedOrder.admin_nota_confirmed_by?.first_name || selectedOrder.confirmed_by?.first_name || 'Admin' }}
                {{ selectedOrder.admin_nota_confirmed_by?.last_name || selectedOrder.confirmed_by?.last_name || '' }}
              </p>
              <p class="text-[10px] text-slate-500 font-medium" v-if="selectedOrder.admin_nota_confirmed_by?.job_position?.title || selectedOrder.confirmed_by?.job_position?.title">
                ({{ selectedOrder.admin_nota_confirmed_by?.job_position?.title || selectedOrder.confirmed_by?.job_position?.title }})
              </p>
              <p class="text-[10px] text-slate-400 mt-0.5">
                {{ formatDate(selectedOrder.confirmed_at) }}
              </p>
            </div>
          </div>

          <div class="border-t border-slate-100 pt-4">
            <div class="flex items-center justify-between mb-3">
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Detail Item Pesanan
              </p>
              <span v-if="selectedOrder.status === 'WAITING_STOCK'" class="text-[10px] font-bold text-orange-600 flex items-center gap-1">
                <AlertTriangle class="w-3 h-3" />
                Rincian Ketersediaan Stok
              </span>
            </div>
            
            <div class="space-y-3">
              <div
                v-for="item in stockSummary"
                :key="item.id"
                class="bg-slate-50 p-3.5 rounded-2xl border border-slate-100/80 space-y-2.5"
              >
                <div class="flex items-start justify-between gap-2">
                  <div>
                    <p class="font-bold text-sm text-slate-900">{{ item.productName }}</p>
                    <p class="text-[10px] text-slate-400 font-mono">{{ item.product?.sku }}</p>
                  </div>
                  <div class="text-right shrink-0">
                    <p class="font-black text-primary">{{ item.ordered_quantity }} {{ item.unit }}</p>
                    <p class="text-[9px] text-slate-400 font-bold" v-if="item.pieces_per_unit > 1">
                      ({{ item.quantity }} PCS)
                    </p>
                    <p class="text-[10px] text-slate-400">{{ formatCurrency(item.subtotal) }}</p>
                  </div>
                </div>

                <!-- Stock Allocation & Backorder Details -->
                <div 
                  v-if="item.breakdown && (selectedOrder.status === 'WAITING_STOCK' || (item.batches && item.batches.length > 0))"
                  class="bg-white p-2.5 rounded-xl border border-slate-100 space-y-2 text-xs shadow-none"
                >
                  <div class="flex items-center justify-between text-[11px] flex-wrap gap-1">
                    <span class="font-bold text-slate-600">Kesiapan Stok:</span>
                    <div class="flex items-center gap-1.5">
                      <Badge 
                        v-if="item.breakdown.hasBackorder"
                        class="bg-orange-100 text-orange-700 border-orange-200 text-[10px] font-bold shadow-none"
                      >
                        ⏳ Kurang {{ item.breakdown.backorderPieces }} PCS ({{ item.breakdown.backorderUnits }} {{ item.unit }})
                      </Badge>
                      <Badge 
                        v-else
                        class="bg-emerald-100 text-emerald-700 border-emerald-200 text-[10px] font-bold shadow-none"
                      >
                        ✅ Siap Kirim ({{ item.breakdown.readyPieces }} PCS)
                      </Badge>
                    </div>
                  </div>

                  <!-- Progress Bar -->
                  <div class="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden flex">
                    <div 
                      class="bg-emerald-500 h-full transition-all duration-300"
                      :style="{ width: `${item.breakdown.readyPercentage}%` }"
                      title="Stok Siap"
                    ></div>
                    <div 
                      v-if="item.breakdown.hasBackorder"
                      class="bg-orange-400 h-full transition-all duration-300"
                      :style="{ width: `${100 - item.breakdown.readyPercentage}%` }"
                      title="Menunggu Stok"
                    ></div>
                  </div>

                  <!-- Batches Breakdown List -->
                  <div class="space-y-1 pt-1" v-if="item.batches && item.batches.length > 0">
                    <p class="text-[9px] font-black text-slate-400 uppercase tracking-wider">Alokasi Batch Gudang:</p>
                    <div class="flex flex-wrap gap-1.5">
                      <span 
                        v-for="b in item.breakdown.realBatches" 
                        :key="b.id || b.batch_no"
                        class="text-[10px] font-semibold bg-emerald-50 text-emerald-800 border border-emerald-200 px-2 py-0.5 rounded-md flex items-center gap-1"
                      >
                        <CheckCircle2 class="w-2.5 h-2.5 text-emerald-600 shrink-0" />
                        Batch {{ b.batch_no }}: {{ b.quantity }} PCS
                      </span>
                      <span 
                        v-if="item.breakdown.hasBackorder"
                        class="text-[10px] font-bold bg-orange-50 text-orange-800 border border-orange-200 px-2 py-0.5 rounded-md flex items-center gap-1"
                      >
                        <Clock class="w-2.5 h-2.5 text-orange-600 shrink-0" />
                        Menunggu Pasokan: {{ item.breakdown.backorderPieces }} PCS
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="border-t border-slate-100 pt-4" v-if="selectedOrder.bonus_items && selectedOrder.bonus_items.length > 0">
            <p class="text-[10px] font-black text-emerald-500 uppercase tracking-widest mb-3">
              Item Bonus (Gratis)
            </p>
            <div class="space-y-2">
              <div
                v-for="bonus in selectedOrder.bonus_items"
                :key="bonus.id"
                class="flex items-center justify-between bg-emerald-50 p-3 rounded-2xl border border-emerald-100"
              >
                <div>
                  <p class="font-bold text-sm text-emerald-900">{{ bonus.bonus_inventory?.name || 'Item Bonus' }}</p>
                  <p class="text-[10px] text-emerald-600/70" v-if="bonus.notes">{{ bonus.notes }}</p>
                </div>
                <div class="text-right">
                  <p class="font-black text-emerald-600">{{ bonus.quantity }} {{ bonus.bonus_inventory?.unit || 'PCS' }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- MT Info or regular notes -->
          <div v-if="parsedMtInfo" class="bg-violet-50/50 border border-violet-100 rounded-[24px] p-5 space-y-3 mt-4">
            <p class="text-[10px] font-black text-violet-600 uppercase tracking-widest">
              Informasi Modern Trade (MT)
            </p>
            <div class="grid grid-cols-2 gap-4 text-xs">
              <div>
                <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Nomor PO</p>
                <p class="font-extrabold text-violet-700 mt-1 text-sm bg-violet-100/50 px-2.5 py-1 rounded-lg inline-block">{{ parsedMtInfo.po_number || '-' }}</p>
              </div>
              <div>
                <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Fulfillment / Stok</p>
                <p class="font-bold text-slate-800 mt-1">
                  {{ parsedMtInfo.fulfillment_type === 'BRANCH' ? 'Cabang' : 'Pusat / Pabrik' }}
                  <span v-if="parsedMtInfo.fulfillment_branch_name" class="text-[10px] text-slate-500 block font-normal">
                    ({{ parsedMtInfo.fulfillment_branch_name }})
                  </span>
                </p>
              </div>
            </div>
            <div v-if="parsedMtInfo.notes" class="border-t border-violet-100/50 pt-2.5">
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Catatan Kunjungan</p>
              <p class="text-xs text-slate-700 font-bold mt-1">{{ parsedMtInfo.notes }}</p>
            </div>
          </div>
          <div v-else-if="selectedOrder.notes" class="bg-slate-50 rounded-[24px] p-5 mt-4">
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1.5">
              Catatan Pesanan
            </p>
            <p class="text-xs text-slate-700 font-bold leading-relaxed">{{ selectedOrder.notes }}</p>
          </div>

          <div class="border border-slate-100 rounded-2xl overflow-hidden mt-4">
            <div
              v-if="(selectedOrder.discount_amount || 0) > 0"
              class="bg-red-50 p-4 flex justify-between items-center border-b border-slate-100"
            >
              <span class="font-bold text-red-600">Diskon Penyesuaian</span>
              <span class="text-lg font-black text-red-600"
                >-{{ formatCurrency(selectedOrder.discount_amount || 0) }}</span
              >
            </div>
            <div class="bg-primary/5 p-4 flex justify-between items-center">
              <span class="font-black text-slate-700">Total Tagihan Bersih</span>
              <span class="text-xl font-black text-primary">{{
                formatCurrency(selectedOrder.total_amount)
              }}</span>
            </div>
          </div>

          <div
            v-if="selectedOrder.reject_notes"
            class="bg-red-50 border border-red-100 rounded-2xl p-4"
          >
            <p class="text-[10px] font-black text-red-500 uppercase tracking-widest mb-1">
              Alasan Penolakan
            </p>
            <p class="text-sm text-red-700 font-medium">{{ selectedOrder.reject_notes }}</p>
          </div>
        </div>

        <!-- Detail Modal Footer -->
        <div class="bg-slate-50 px-6 py-4 border-t border-slate-100 flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <Button
              v-if="canConvertDirect(selectedOrder)"
              size="sm"
              @click="handleConvertDirect(selectedOrder)"
              class="rounded-xl font-black text-xs gap-1.5 bg-indigo-600 hover:bg-indigo-700 text-white"
            >
              <FileText class="w-4 h-4" /> Langsung ke Faktur
            </Button>
            <Button
              v-if="selectedOrder.status === 'DRAFT' || selectedOrder.status === 'CONFIRMED' || selectedOrder.status === 'WAITING_STOCK'"
              variant="outline"
              size="sm"
              @click="showDetailModal = false; openAdjust(selectedOrder)"
              class="rounded-xl font-black text-xs gap-1.5 text-violet-700 bg-violet-50 hover:bg-violet-100 border-violet-200"
            >
              <SlidersHorizontal class="w-4 h-4" />
              Adjust Pesanan
            </Button>

            <!-- Super Admin Fast-Track Complete Button in Detail Modal -->
            <Button
              v-if="canSuperAdminComplete(selectedOrder)"
              size="sm"
              @click="handleSuperAdminComplete(selectedOrder)"
              class="rounded-xl font-black text-xs gap-1.5 bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-600 hover:to-amber-700 text-white shadow-sm border border-amber-400/30"
            >
              <Sparkles class="w-4 h-4 text-amber-100" />
              Auto Complete (Super Admin)
            </Button>
          </div>
          <Button
            variant="ghost"
            size="sm"
            @click="showDetailModal = false"
            class="rounded-xl font-bold text-xs"
          >
            Tutup
          </Button>
        </div>
      </div>
    </div>

    <!-- Adjust Modal -->
    <SOAdjustModal
      :order="selectedOrder"
      :show="showAdjustModal"
      @close="showAdjustModal = false"
      @saved="fetchOrders"
    />

    <!-- Waiting Stock Analysis Modal -->
    <SOWaitingStockModal
      :order="selectedOrderForWaitingStock"
      :show="showWaitingStockModal"
      @close="showWaitingStockModal = false"
      @refresh="fetchOrders"
      @open-adjust="openAdjust"
    />

    <!-- Reject Modal -->
    <div v-if="rejectModal.open" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="rejectModal.open = false"
      />
      <div
        class="bg-white rounded-[32px] w-full max-w-md relative z-10 shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200"
      >
        <div class="p-8 space-y-6">
          <div class="flex flex-col items-center text-center space-y-4">
            <div class="p-4 rounded-full bg-red-50 text-red-500 mb-2">
              <XCircle class="w-10 h-10" />
            </div>
            <h3 class="text-2xl font-black text-slate-900">Tolak Pesanan</h3>
            <p class="text-sm text-slate-500">
              Masukkan alasan penolakan untuk pesanan ini agar sales mengetahui penyebabnya.
            </p>
          </div>

          <textarea
            v-model="rejectModal.notes"
            rows="3"
            placeholder="Ketik alasan penolakan di sini..."
            class="w-full bg-slate-50 border border-slate-200 rounded-2xl px-5 py-4 text-sm focus:ring-2 focus:ring-red-500/20 focus:border-red-500 outline-none resize-none"
          />
        </div>

        <div class="bg-slate-50 p-6 flex gap-3 justify-end border-t border-slate-100">
          <Button
            variant="ghost"
            @click="rejectModal.open = false"
            class="rounded-xl font-bold px-6"
            >Batal</Button
          >
          <Button
            @click="submitReject"
            :disabled="rejectModal.isLoading"
            class="rounded-xl font-black px-6 bg-red-600 hover:bg-red-700 text-white"
          >
            {{ rejectModal.isLoading ? 'Menyimpan...' : 'Tolak Pesanan' }}
          </Button>
        </div>
      </div>
    </div>

    <!-- Tax Approval & Documents Upload Modal (Modern Trade) -->
    <div v-if="showTaxModal" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showTaxModal = false"
      />
      <div
        class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200"
      >
        <div class="px-8 py-6 border-b border-slate-100 flex items-center justify-between bg-violet-50/50">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-violet-100 text-violet-700 rounded-xl">
              <FileText class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-lg font-black text-slate-900">Persetujuan & Dokumen Pajak MT</h3>
              <p class="text-[10px] font-bold text-slate-500 uppercase tracking-widest mt-0.5">Modern Trade B2G / B2B Workflow</p>
            </div>
          </div>
          <button
            @click="showTaxModal = false"
            class="p-2 hover:bg-slate-200/50 rounded-xl text-slate-400 hover:text-slate-900 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-8 space-y-6 max-h-[60vh] overflow-y-auto">
          <div class="p-4 bg-amber-50 border border-amber-200 rounded-2xl flex gap-3 text-amber-800 text-xs leading-relaxed">
            <span class="font-bold">⚠️ Informasi:</span>
            <span>Pesanan institusi memerlukan unggahan 3 dokumen resmi dari portal instansi pemerintah sebelum dikonfirmasi ke gudang.</span>
          </div>

          <!-- Nota Penagihan -->
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest flex justify-between">
              <span>1. Nota Penagihan (Billing Note)</span>
              <span class="text-red-500">*Wajib</span>
            </label>
            <div class="flex gap-2">
              <input
                v-model="taxForm.billing_note_url"
                type="text"
                placeholder="Unggah berkas atau masukkan URL PDF Nota Penagihan..."
                class="flex-1 bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-xs font-semibold focus:ring-2 focus:ring-violet-500/20 focus:border-violet-500 outline-none"
              />
              <Button
                type="button"
                variant="outline"
                @click="taxForm.billing_note_url = ''"
                class="rounded-xl px-4 text-xs font-bold text-violet-700 border-violet-200 bg-violet-50/50 hover:bg-violet-100/50"
              >
                Auto Mock
              </Button>
            </div>
          </div>

          <!-- Faktur Pajak -->
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest flex justify-between">
              <span>2. Faktur Pajak (Tax Invoice)</span>
              <span class="text-red-500">*Wajib</span>
            </label>
            <div class="flex gap-2">
              <input
                v-model="taxForm.tax_invoice_url"
                type="text"
                placeholder="Unggah berkas atau masukkan URL PDF Faktur Pajak..."
                class="flex-1 bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-xs font-semibold focus:ring-2 focus:ring-violet-500/20 focus:border-violet-500 outline-none"
              />
              <Button
                type="button"
                variant="outline"
                @click="taxForm.tax_invoice_url = ''"
                class="rounded-xl px-4 text-xs font-bold text-violet-700 border-violet-200 bg-violet-50/50 hover:bg-violet-100/50"
              >
                Auto Mock
              </Button>
            </div>
          </div>

          <!-- Surat Jalan -->
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest flex justify-between">
              <span>3. Surat Jalan Resmi (Delivery Note)</span>
              <span class="text-red-500">*Wajib</span>
            </label>
            <div class="flex gap-2">
              <input
                v-model="taxForm.delivery_note_url"
                type="text"
                placeholder="Unggah berkas atau masukkan URL PDF Surat Jalan..."
                class="flex-1 bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-xs font-semibold focus:ring-2 focus:ring-violet-500/20 focus:border-violet-500 outline-none"
              />
              <Button
                type="button"
                variant="outline"
                @click="taxForm.delivery_note_url = ''"
                class="rounded-xl px-4 text-xs font-bold text-violet-700 border-violet-200 bg-violet-50/50 hover:bg-violet-100/50"
              >
                Auto Mock
              </Button>
            </div>
          </div>

          <!-- Cabang Gudang Pengirim -->
          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest flex justify-between">
              <span>4. Cabang Gudang Pengirim</span>
              <span class="text-red-500">*Wajib</span>
            </label>
            <select
              v-model="taxForm.destination_branch_id"
              class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-xs font-semibold focus:ring-2 focus:ring-violet-500/20 focus:border-violet-500 outline-none"
            >
              <option value="" disabled selected>Pilih Cabang Gudang Pengirim...</option>
              <option
                v-for="branch in masterStore.branches"
                :key="branch.id"
                :value="branch.id"
              >
                {{ branch.name }}
              </option>
            </select>
          </div>
        </div>

        <div class="bg-slate-50 p-6 flex gap-3 justify-end border-t border-slate-100">
          <Button
            variant="ghost"
            @click="showTaxModal = false"
            class="rounded-xl font-bold px-6"
            >Batal</Button
          >
          <Button
            @click="submitTaxApproval"
            :disabled="taxLoading"
            class="rounded-xl font-black px-8 bg-violet-600 hover:bg-violet-700 text-white shadow-lg shadow-violet-200"
          >
            {{ taxLoading ? 'Menyimpan...' : 'Approve & Terbitkan PO' }}
          </Button>
        </div>
      </div>
    </div>

    <!-- Confirmation Dialog -->
    <ConfirmationDialog
      v-model:open="confirmDialog.open"
      :title="confirmDialog.title"
      :description="confirmDialog.description"
      :variant="confirmDialog.variant"
      :confirm-text="confirmDialog.confirmText"
      :cancel-text="confirmDialog.cancelText"
      :is-loading="confirmDialog.isLoading"
      @confirm="confirmDialog.onConfirm"
      @cancel="confirmDialog.open = false"
    />

    <!-- Create Store Modal -->
    <div v-if="showCreateStoreModal" class="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showCreateStoreModal = false"
      />
      <div class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl flex flex-col overflow-hidden">
        <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-xl font-black text-slate-900">Tambah Toko Baru</h3>
          <button @click="showCreateStoreModal = false" class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Nama Toko <span class="text-red-500">*</span></label>
            <input v-model="storeForm.name" type="text" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none" placeholder="Masukkan nama toko" />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Alamat Toko <span class="text-red-500">*</span></label>
            <textarea v-model="storeForm.address" rows="2" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none resize-none" placeholder="Masukkan alamat lengkap toko"></textarea>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-1">
              <label class="text-xs font-bold text-slate-700">Nama Pemilik</label>
              <input v-model="storeForm.owner_name" type="text" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none" placeholder="Opsional" />
            </div>
            <div class="space-y-1">
              <label class="text-xs font-bold text-slate-700">No. Telepon</label>
              <input v-model="storeForm.phone_number" type="text" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none" placeholder="Opsional" />
            </div>
          </div>
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Terikat dengan Salesman?</label>
            <select v-model="storeForm.assigned_employee_id" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm font-semibold focus:ring-2 focus:ring-primary/20 outline-none">
              <option value="">Tidak Terikat (Toko Bebas)</option>
              <option v-for="e in salesEmployees" :key="e.id" :value="e.id">
                {{ e.first_name }} {{ e.last_name }}
              </option>
            </select>
            <p class="text-[10px] text-slate-500 mt-1">Jika terikat, toko ini hanya akan muncul untuk salesman tersebut.</p>
          </div>
        </div>
        <div class="bg-slate-50 p-5 flex justify-end gap-3 border-t border-slate-100">
          <Button variant="ghost" @click="showCreateStoreModal = false" class="rounded-xl font-bold">Batal</Button>
          <Button @click="handleCreateStore" :disabled="isCreatingStore" class="rounded-xl font-bold px-6 bg-primary text-white hover:bg-primary/90">
            {{ isCreatingStore ? 'Menyimpan...' : 'Simpan Toko' }}
          </Button>
        </div>
      </div>
    </div>

    <!-- Export e-Faktur Modal -->
    <div v-if="showExportModal" class="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showExportModal = false" />
      <div class="bg-white rounded-[32px] w-full max-w-sm relative z-10 shadow-2xl flex flex-col overflow-hidden">
        <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-xl font-black text-slate-900">Export e-Faktur DJP</h3>
          <button @click="showExportModal = false" class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Tanggal Awal <span class="text-red-500">*</span></label>
            <input v-model="exportForm.start_date" type="date" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none" />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Tanggal Akhir <span class="text-red-500">*</span></label>
            <input v-model="exportForm.end_date" type="date" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none" />
          </div>
        </div>
        <div class="bg-slate-50 p-5 flex justify-end gap-3 border-t border-slate-100">
          <Button variant="ghost" @click="showExportModal = false" class="rounded-xl font-bold">Batal</Button>
          <Button @click="handleExportEFaktur" class="rounded-xl font-bold px-6 bg-green-600 hover:bg-green-700 text-white">
            Export CSV
          </Button>
        </div>
      </div>
    </div>

    <!-- Set Faktur Pajak Modal -->
    <div v-if="showSetFakturModal" class="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showSetFakturModal = false" />
      <div class="bg-white rounded-[32px] w-full max-w-md relative z-10 shadow-2xl flex flex-col overflow-hidden">
        <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-xl font-black text-slate-900">Set Faktur Pajak</h3>
          <button @click="showSetFakturModal = false" class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Nomor Seri Faktur Pajak (NSFP) <span class="text-red-500">*</span></label>
            <input v-model="fakturForm.tax_invoice_number" type="text" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none" placeholder="000.000-00.00000000" />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Tanggal Faktur <span class="text-red-500">*</span></label>
            <input v-model="fakturForm.tax_invoice_date" type="date" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none" />
          </div>
          <div class="space-y-1">
            <label class="text-xs font-bold text-slate-700">Tarif PPN (%) <span class="text-red-500">*</span></label>
            <input v-model.number="fakturForm.tax_rate" type="number" class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-primary/20 outline-none" />
          </div>
          <div class="bg-slate-50 p-4 rounded-xl border border-slate-200 space-y-2 mt-2">
            <div class="flex justify-between text-sm">
              <span class="text-slate-500 font-bold">Total Order</span>
              <span class="font-black text-slate-900">{{ formatCurrency(fakturOrderAmount) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-slate-500 font-bold">DPP</span>
              <span class="font-black text-slate-900">{{ formatCurrency(fakturDpp) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-slate-500 font-bold">PPN ({{ fakturForm.tax_rate }}%)</span>
              <span class="font-black text-slate-900">{{ formatCurrency(fakturPpn) }}</span>
            </div>
          </div>
        </div>
        <div class="bg-slate-50 p-5 flex justify-end gap-3 border-t border-slate-100">
          <Button variant="ghost" @click="showSetFakturModal = false" class="rounded-xl font-bold">Batal</Button>
          <Button @click="submitSetFaktur" :disabled="fakturLoading" class="rounded-xl font-bold px-6 bg-primary text-white hover:bg-primary/90">
            {{ fakturLoading ? 'Menyimpan...' : 'Simpan NSFP' }}
          </Button>
        </div>
      </div>
    </div>
    
    <!-- Safety Soft Delete Modal -->
    <div v-if="showDeleteSafetyModal && deleteTargetSO" class="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showDeleteSafetyModal = false" />
      <Card class="relative w-full max-w-lg bg-white rounded-3xl shadow-2xl border-none overflow-hidden z-10 animate-in fade-in zoom-in-95 duration-200">
        <div class="p-6 space-y-5">
          <div class="flex items-center gap-3 text-rose-600">
            <div class="p-3 bg-rose-50 rounded-2xl">
              <AlertTriangle class="w-6 h-6 text-rose-600" />
            </div>
            <div>
              <h3 class="font-extrabold text-lg text-slate-900 leading-tight">Hapus Pesanan SO</h3>
              <p class="text-xs text-rose-600 font-semibold mt-0.5">Penghapusan Aman (Soft Delete - Data dapat dipulihkan)</p>
            </div>
          </div>

          <div class="bg-slate-50 p-4 rounded-2xl border border-slate-200 space-y-2 text-xs">
            <div class="flex justify-between">
              <span class="text-slate-500 font-medium">No. Pesanan (SO):</span>
              <span class="font-mono font-bold text-slate-900">{{ deleteTargetSO.so_number }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500 font-medium">Toko:</span>
              <span class="font-bold text-slate-900">{{ deleteTargetSO.store?.name || '-' }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500 font-medium">Salesman:</span>
              <span class="font-bold text-slate-900">{{ deleteTargetSO.employee?.first_name }} {{ deleteTargetSO.employee?.last_name }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500 font-medium">Total Nominal:</span>
              <span class="font-black text-slate-900">{{ formatCurrency(deleteTargetSO.total_amount) }}</span>
            </div>
          </div>

          <div class="space-y-3">
            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                Alasan Penghapusan <span class="text-rose-500">*</span>
              </label>
              <textarea
                v-model="deleteReason"
                placeholder="Contoh: Salah input toko, pesanan dibatalkan customer, double pesanan..."
                rows="2"
                class="w-full bg-slate-50 border border-slate-200 rounded-xl p-3 text-xs font-medium text-slate-900 placeholder:text-slate-400 focus:bg-white focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 outline-none transition-all"
              ></textarea>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                Ketik <span class="font-mono text-rose-600 bg-rose-50 px-1.5 py-0.5 rounded">{{ deleteTargetSO.so_number }}</span> untuk konfirmasi <span class="text-rose-500">*</span>
              </label>
              <input
                v-model="deleteConfirmText"
                type="text"
                :placeholder="deleteTargetSO.so_number"
                class="w-full bg-slate-50 border border-slate-200 rounded-xl p-3 text-xs font-mono font-bold text-slate-900 placeholder:text-slate-300 focus:bg-white focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 outline-none transition-all"
              />
            </div>
          </div>

          <div class="flex gap-3 pt-2">
            <Button
              variant="outline"
              class="flex-1 rounded-xl py-5 font-bold text-xs"
              @click="showDeleteSafetyModal = false"
            >
              Batal
            </Button>
            <Button
              class="flex-1 bg-rose-600 hover:bg-rose-700 text-white font-bold rounded-xl py-5 text-xs shadow-lg shadow-rose-600/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="isDeletingSO || !deleteReason.trim() || deleteConfirmText.trim() !== deleteTargetSO.so_number"
              @click="confirmDeleteSO"
            >
              <RefreshCw v-if="isDeletingSO" class="w-4 h-4 animate-spin mr-1.5" />
              <span>{{ isDeletingSO ? 'Menghapus...' : 'Ya, Hapus Pesanan' }}</span>
            </Button>
          </div>
        </div>
      </Card>
    </div>
    
    <OrderLocationMapModal 
      :open="isMapModalOpen" 
      @update:open="isMapModalOpen = $event" 
      :order-data="selectedOrderForMap" 
    />
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
</style>
