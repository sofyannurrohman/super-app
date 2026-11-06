<script setup lang="ts">
import { ref, h, onMounted, toRaw, watch, computed } from 'vue'
import {
  CheckCircle2,
  XCircle,
  FileText,
  ExternalLink,
  MoreVertical,
  Plus,
  Pencil,
  Trash2,
  Eye,
  Building2,
  Search,
  Receipt,
  Image as ImageIcon,
  MapPin,
  ZoomIn,
  Clock,
  DollarSign,
  Layers,
  Sparkles,
  Check,
  X,
  Fuel,
  Gauge,
  Navigation,
} from 'lucide-vue-next'
import apiClient from '@/api/axios'
import DataTable from '@/components/DataTable.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { toast } from 'vue-sonner'
import { Label } from '@/components/ui/label'
import { useMasterDataStore } from '@/stores/masterData'
import { resolvePhotoUrl } from '@/utils/imageUrl'

const masterDataStore = useMasterDataStore()

// State Data
const reimbursements = ref<any[]>([])
const employees = ref<any[]>([])
const unclaimedBanners = ref<any[]>([])
const isLoading = ref(true)
const isLoadingUnclaimed = ref(false)
const isSubmitting = ref(false)
const API_BASE = (import.meta.env.VITE_API_BASE_URL || '').replace('/api', '')

// Navigation / Filter State
type TabType = 'ALL' | 'BBM' | 'GENERAL' | 'BANNER_INSTALLATION' | 'UNCLAIMED'
const activeTab = ref<TabType>('ALL')
const selectedStatus = ref<string>('ALL')
const searchQuery = ref<string>('')

// Detail Modal State (Full inspection for any status)
const isDetailModalOpen = ref(false)
const detailItem = ref<any>(null)

// Review Modal State (Approval & Rate adjustment)
const isModalOpen = ref(false)
const selectedItem = ref<any>(null)
const rejectReason = ref('')
const customUnitRate = ref(15000)
const customAmount = ref(0)

// Approval Deduction & Expense Account States
const kasPabriks = ref<any[]>([])
const expenseAccounts = ref<any[]>([])
const companyBanks = ref<any[]>([])
const branchCashBalances = ref<{ petty_cash_balance: number; manual_cash_balance: number }>({
  petty_cash_balance: 0,
  manual_cash_balance: 0,
})

const approvalForm = ref({
  fund_source: 'BRANCH_PETTY_CASH' as 'KAS_PABRIK' | 'BRANCH_PETTY_CASH' | 'BRANCH_CASH' | 'BANK_TRANSFER',
  kas_pabrik_id: '',
  bank_account_id: '',
  expense_account_id: '',
  notes: '',
})

// Lightbox Preview Image
const previewImageUrl = ref<string | null>(null)
const previewImageTitle = ref<string>('')

// CRUD Modal State
const isCRUDModalOpen = ref(false)
const isEditMode = ref(false)
const createCategory = ref<'GENERAL' | 'BANNER_INSTALLATION' | 'BBM'>('GENERAL')
const unclaimedForEmployee = ref<any[]>([])
const isLoadingEmpBanners = ref(false)
const selectedBannerIds = ref<string[]>([])
const bannerUnitRateInput = ref(15000)

const form = ref({
  id: '',
  employee_id: '',
  title: '',
  amount: 0,
  description: '',
  attachment_url: '',
  attachment_file: null as File | null,
  status: 'PENDING',
  start_km: null as number | null,
  end_km: null as number | null,
  total_km: null as number | null,
  start_km_photo_url: '',
  start_km_photo_file: null as File | null,
  end_km_photo_url: '',
  end_km_photo_file: null as File | null,
  vehicle_plate_number: '',
  fuel_liters: null as number | null,
})

// Helper to format currency
const formatRupiah = (val: number | string | null | undefined) => {
  const num = Number(val || 0)
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(num)
}

// Fetch Reimbursements
const fetchReimbursements = async () => {
  isLoading.value = true
  try {
    const branchId = masterDataStore.selectedBranchId
    const params: Record<string, string> = { limit: '200' }
    if (branchId && branchId !== 'ALL_BRANCHES') {
      params.branch_id = branchId
    }
    if (selectedStatus.value && selectedStatus.value !== 'ALL') {
      params.status = selectedStatus.value
    }
    if (activeTab.value === 'GENERAL' || activeTab.value === 'BANNER_INSTALLATION') {
      params.category = activeTab.value
    }

    const res = await apiClient.get('/reimbursements/manage', { params })
    reimbursements.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch reimbursements:', error)
    toast.error('Gagal mengambil data reimbursement')
  } finally {
    isLoading.value = false
  }
}

// Fetch Unclaimed Installed Banners
const fetchUnclaimedBanners = async () => {
  isLoadingUnclaimed.value = true
  try {
    const branchId = masterDataStore.selectedBranchId
    const params: Record<string, string> = {}
    if (branchId && branchId !== 'ALL_BRANCHES') {
      params.branch_id = branchId
    }
    const res = await apiClient.get('/branding/sales/unclaimed-installations', { params })
    unclaimedBanners.value = res.data.data || []
  } catch (error) {
    console.error('Failed to fetch unclaimed banners:', error)
  } finally {
    isLoadingUnclaimed.value = false
  }
}

// Fetch Employees list for select dropdown
const fetchEmployees = async () => {
  try {
    const res = await apiClient.get('/employees?limit=all')
    employees.value = res.data.data || []
  } catch (e) {
    console.error('Failed to fetch employees:', e)
  }
}

// Fetch Unclaimed Banners for specific employee when creating Banner claim
const fetchEmployeeUnclaimedBanners = async (employeeId: string) => {
  if (!employeeId) {
    unclaimedForEmployee.value = []
    selectedBannerIds.value = []
    return
  }
  isLoadingEmpBanners.value = true
  try {
    const res = await apiClient.get('/branding/sales/unclaimed-installations', {
      params: { employee_id: employeeId },
    })
    unclaimedForEmployee.value = res.data.data || []
    // Default select all banners
    selectedBannerIds.value = unclaimedForEmployee.value.map((b) => b.id)
    recalculateBannerClaimAmount()
  } catch (e) {
    console.error('Failed to fetch employee unclaimed banners:', e)
    unclaimedForEmployee.value = []
    selectedBannerIds.value = []
  } finally {
    isLoadingEmpBanners.value = false
  }
}

// Auto Recalculate banner claim amount
const recalculateBannerClaimAmount = () => {
  const count = selectedBannerIds.value.length
  form.value.amount = count * bannerUnitRateInput.value
  if (count > 0 && createCategory.value === 'BANNER_INSTALLATION') {
    const emp = employees.value.find((e) => e.id === form.value.employee_id)
    const empName = emp ? `${emp.first_name} ${emp.last_name || ''}`.trim() : 'Karyawan'
    const todayStr = new Date().toLocaleDateString('id-ID')
    form.value.title = `Reimbursement Pasang Spanduk (${count} Unit) - ${empName} - ${todayStr}`
  }
}

// Toggle single banner selection
const toggleBannerSelection = (id: string) => {
  const idx = selectedBannerIds.value.indexOf(id)
  if (idx > -1) {
    selectedBannerIds.value.splice(idx, 1)
  } else {
    selectedBannerIds.value.push(id)
  }
  recalculateBannerClaimAmount()
}

// Toggle all banners
const toggleSelectAllBanners = () => {
  if (selectedBannerIds.value.length === unclaimedForEmployee.value.length) {
    selectedBannerIds.value = []
  } else {
    selectedBannerIds.value = unclaimedForEmployee.value.map((b) => b.id)
  }
  recalculateBannerClaimAmount()
}

// Helper to detect if a reimbursement is a BBM claim
const isBBMClaim = (item: any) => {
  if (!item) return false
  if (item.category === 'BBM') return true
  if (item.start_km != null || item.end_km != null) return true
  const text = `${item.title || ''} ${item.description || ''}`.toLowerCase()
  return (
    text.includes('bbm') ||
    text.includes('bensin') ||
    text.includes('solar') ||
    text.includes('pertalite') ||
    text.includes('pertamax') ||
    text.includes('fuel')
  )
}

// Filtered List based on Search Query & Tab
const filteredReimbursements = computed(() => {
  let list = reimbursements.value

  if (activeTab.value === 'GENERAL') {
    list = list.filter((item) => (!item.category || item.category === 'GENERAL') && !isBBMClaim(item))
  } else if (activeTab.value === 'BANNER_INSTALLATION') {
    list = list.filter((item) => item.category === 'BANNER_INSTALLATION')
  } else if (activeTab.value === 'BBM') {
    list = list.filter((item) => isBBMClaim(item))
  }

  if (selectedStatus.value && selectedStatus.value !== 'ALL') {
    list = list.filter((item) => item.status === selectedStatus.value)
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter((item) => {
      const empName = `${item.employee?.first_name || ''} ${item.employee?.last_name || ''}`.toLowerCase()
      const title = (item.title || '').toLowerCase()
      const desc = (item.description || '').toLowerCase()
      const plate = (item.vehicle_plate_number || '').toLowerCase()
      const shopNames = (item.banner_orders || []).map((b: any) => (b.store_name || b.shop_name || '').toLowerCase()).join(' ')
      return empName.includes(q) || title.includes(q) || desc.includes(q) || plate.includes(q) || shopNames.includes(q)
    })
  }

  return list
})

// Filtered Unclaimed Banners
const filteredUnclaimedBanners = computed(() => {
  let list = unclaimedBanners.value
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter((b) => {
      const shop = (b.store_name || b.shop_name || '').toLowerCase()
      const emp = (b.employee ? `${b.employee.first_name} ${b.employee.last_name || ''}` : b.sales_name || '').toLowerCase()
      const loc = (b.location || b.shop_address || '').toLowerCase()
      return shop.includes(q) || emp.includes(q) || loc.includes(q)
    })
  }
  return list
})

// Statistics Counters
const stats = computed(() => {
  const all = reimbursements.value
  const banner = all.filter((r) => r.category === 'BANNER_INSTALLATION')
  const bbm = all.filter((r) => isBBMClaim(r))
  const general = all.filter((r) => !isBBMClaim(r) && (!r.category || r.category === 'GENERAL'))
  const pending = all.filter((r) => r.status === 'PENDING')
  const disbursed = all.filter((r) => r.disbursement_status === 'DISBURSED')

  const totalNominal = all
    .filter((r) => r.status !== 'REJECTED')
    .reduce((acc, curr) => acc + Number(curr.amount || 0), 0)

  const totalDisbursedNominal = disbursed.reduce((acc, curr) => acc + Number(curr.amount || 0), 0)

  let totalInstalledBannersCount = 0
  banner.forEach((b) => {
    totalInstalledBannersCount += b.banner_count || b.banner_orders?.length || 1
  })

  return {
    totalAll: all.length,
    totalGeneral: general.length,
    totalBanner: banner.length,
    totalBBM: bbm.length,
    totalUnclaimed: unclaimedBanners.value.length,
    pendingCount: pending.length,
    totalNominal,
    totalDisbursedNominal,
    totalInstalledBannersCount,
  }
})

// Fetch approval auxiliary data
const fetchApprovalMasterData = async () => {
  try {
    const [kasRes, bankRes] = await Promise.all([
      apiClient.get('/kas-pabrik'),
      apiClient.get('/finance/bank-accounts').catch(() => ({ data: { data: [] } })),
    ])
    kasPabriks.value = kasRes.data?.data || kasRes.data || []
    companyBanks.value = bankRes.data?.data || []
  } catch (err) {
    console.error('Failed to load approval master data:', err)
  }
}

const fetchExpenseAccounts = async (companyId?: string) => {
  try {
    const params: Record<string, string> = { type: 'EXPENSE' }
    if (companyId) params.company_id = companyId
    const res = await apiClient.get('/accounting/accounts', { params })
    const all = res.data?.data || []
    expenseAccounts.value = all.filter((a: any) => a.account_type === 'EXPENSE')
  } catch (err) {
    console.error('Failed to load expense accounts:', err)
  }
}

const fetchBranchBalances = async (branchId: string) => {
  if (!branchId) return
  try {
    const res = await apiClient.get(`/cashbank-mutations/cash-balance?branch_id=${branchId}`)
    const data = res.data?.data || {}
    branchCashBalances.value = {
      petty_cash_balance: data.petty_cash_balance || 0,
      manual_cash_balance: data.manual_cash_balance || 0,
    }
  } catch (err) {
    console.error('Failed to load branch cash balances:', err)
  }
}

// Check if selected item's employee belongs to Pabrik
const isSelectedItemFromPabrik = computed(() => {
  if (!selectedItem.value) return false
  const branchName = (selectedItem.value.employee?.branch?.name || '').toLowerCase()
  return branchName.includes('pabrik')
})

// Review Modal Handlers
const openReviewModal = async (item: any) => {
  selectedItem.value = item
  rejectReason.value = ''
  const count = item.banner_count || item.banner_orders?.length || 1
  customUnitRate.value = item.unit_rate > 0 ? Number(item.unit_rate) : 15000
  customAmount.value = item.amount > 0 ? Number(item.amount) : customUnitRate.value * count

  // Load auxiliary data
  await Promise.all([
    fetchApprovalMasterData(),
    fetchExpenseAccounts(item.employee?.company_id),
  ])

  if (item.employee?.branch_id) {
    await fetchBranchBalances(item.employee.branch_id)
  }

  const isPabrik = (item.employee?.branch?.name || '').toLowerCase().includes('pabrik')

  // Set default Fund Source: KAS_PABRIK for Pabrik, BRANCH_PETTY_CASH for others
  if (isPabrik) {
    approvalForm.value.fund_source = 'KAS_PABRIK'
    // Default to petty cash pabrik if exists
    const pettyKas = kasPabriks.value.find((k: any) =>
      (k.kas_name || '').toUpperCase().includes('PETTY') || (k.kas_code || '').toUpperCase().includes('PETTY')
    )
    approvalForm.value.kas_pabrik_id = pettyKas ? pettyKas.id : (kasPabriks.value[0]?.id || '')
  } else {
    approvalForm.value.fund_source = 'BRANCH_PETTY_CASH'
    approvalForm.value.kas_pabrik_id = ''
  }

  // Set default Bank if any
  approvalForm.value.bank_account_id = companyBanks.value[0]?.id || ''
  approvalForm.value.notes = ''

  // Auto recommend expense account based on category
  const cat = (item.category || '').toUpperCase()
  let matchedAcc: any = null
  if (cat === 'BBM' || (item.title && item.title.toLowerCase().includes('bbm'))) {
    matchedAcc = expenseAccounts.value.find((a: any) =>
      a.code === '5310' || (a.name || '').toLowerCase().includes('bbm') || (a.name || '').toLowerCase().includes('transport')
    )
  } else if (cat === 'BANNER_INSTALLATION') {
    matchedAcc = expenseAccounts.value.find((a: any) =>
      a.code === '5320' || (a.name || '').toLowerCase().includes('promosi') || (a.name || '').toLowerCase().includes('spanduk')
    )
  }

  if (!matchedAcc) {
    matchedAcc = expenseAccounts.value.find((a: any) =>
      a.code === '5400' || (a.name || '').toLowerCase().includes('reimburse')
    )
  }
  if (!matchedAcc) {
    matchedAcc = expenseAccounts.value.find((a: any) => a.code === '5300') || expenseAccounts.value[0]
  }
  approvalForm.value.expense_account_id = matchedAcc ? matchedAcc.id : ''

  isModalOpen.value = true
}

const openDetailModal = (item: any) => {
  detailItem.value = item
  isDetailModalOpen.value = true
}

const openLightbox = (imageUrl: string, title = 'Foto Bukti') => {
  if (!imageUrl) return
  previewImageUrl.value = resolvePhotoUrl(imageUrl)
  previewImageTitle.value = title
}

const updateStatus = async (status: 'APPROVED' | 'REJECTED') => {
  if (!selectedItem.value) return

  if (status === 'REJECTED' && !rejectReason.value) {
    toast.error('Gagal: Alasan penolakan harus diisi.')
    return
  }

  if (status === 'APPROVED') {
    if (isSelectedItemFromPabrik.value && approvalForm.value.fund_source !== 'KAS_PABRIK') {
      toast.error('Karyawan Pabrik wajib dicairkan melalui Petty Cash Pabrik!')
      return
    }
    if (approvalForm.value.fund_source === 'KAS_PABRIK' && !approvalForm.value.kas_pabrik_id) {
      toast.error('Pilih akun KAS Pabrik pemotong saldo!')
      return
    }
    if (approvalForm.value.fund_source === 'BANK_TRANSFER' && !approvalForm.value.bank_account_id) {
      toast.error('Pilih rekening bank sumber pencairan!')
      return
    }
  }

  isSubmitting.value = true
  try {
    const payload: any = {
      status: status,
      rejected_reason: status === 'REJECTED' ? rejectReason.value : null,
      fund_source: approvalForm.value.fund_source,
      kas_pabrik_id: approvalForm.value.fund_source === 'KAS_PABRIK' ? approvalForm.value.kas_pabrik_id : null,
      bank_account_id: approvalForm.value.fund_source === 'BANK_TRANSFER' ? approvalForm.value.bank_account_id : null,
      expense_account_id: approvalForm.value.expense_account_id || null,
      notes: approvalForm.value.notes,
    }

    if (status === 'APPROVED' && selectedItem.value.category === 'BANNER_INSTALLATION') {
      await apiClient.put(`/reimbursements/manage/${selectedItem.value.id}/adjust-banner-rate`, {
        unit_rate: Number(customUnitRate.value),
        amount: Number(customAmount.value),
        approve: true,
        ...payload,
      })
    } else {
      await apiClient.put(`/reimbursements/manage/${selectedItem.value.id}/approve`, payload)
    }
    toast.success(`Berhasil ${status === 'APPROVED' ? 'menyetujui & mencairkan' : 'menolak'} pengajuan.`)
    isModalOpen.value = false
    fetchReimbursements()
    fetchUnclaimedBanners()
  } catch (error: any) {
    toast.error('Gagal memperbarui status: ' + (error.response?.data?.message || error.message))
  } finally {
    isSubmitting.value = false
  }
}

// Open CRUD Add Modal
const openAddModal = (presetEmpId?: string, presetBannerId?: string) => {
  isEditMode.value = false
  createCategory.value = presetBannerId ? 'BANNER_INSTALLATION' : 'GENERAL'
  bannerUnitRateInput.value = 15000
  selectedBannerIds.value = presetBannerId ? [presetBannerId] : []
  unclaimedForEmployee.value = []

  form.value = {
    id: '',
    employee_id: presetEmpId || '',
    title: '',
    amount: 0,
    description: '',
    attachment_url: '',
    attachment_file: null,
    status: 'PENDING',
    start_km: null,
    end_km: null,
    total_km: null,
    start_km_photo_url: '',
    start_km_photo_file: null,
    end_km_photo_url: '',
    end_km_photo_file: null,
    vehicle_plate_number: '',
    fuel_liters: null,
  }

  if (presetEmpId) {
    fetchEmployeeUnclaimedBanners(presetEmpId)
  }

  isCRUDModalOpen.value = true
}

// Watch employee change in Add modal
watch(
  () => form.value.employee_id,
  (newEmpId) => {
    if (createCategory.value === 'BANNER_INSTALLATION' && newEmpId && !isEditMode.value) {
      fetchEmployeeUnclaimedBanners(newEmpId)
    }
  },
)

watch(
  () => createCategory.value,
  (newCat) => {
    if (newCat === 'BANNER_INSTALLATION' && form.value.employee_id && !isEditMode.value) {
      fetchEmployeeUnclaimedBanners(form.value.employee_id)
    }
  },
)

// Auto recalculate total KM when start_km or end_km changes
watch(
  () => [form.value.start_km, form.value.end_km],
  ([start, end]) => {
    if (start !== null && end !== null && start !== undefined && end !== undefined && start !== ('' as any) && end !== ('' as any)) {
      const s = Number(start)
      const e = Number(end)
      if (e >= s) {
        form.value.total_km = e - s
      } else {
        form.value.total_km = 0
      }
    }
  },
)

const openEditModal = (item: any) => {
  isEditMode.value = true
  createCategory.value = item.category === 'BANNER_INSTALLATION' ? 'BANNER_INSTALLATION' : (item.category === 'BBM' ? 'BBM' : 'GENERAL')
  form.value = {
    id: item.id,
    employee_id: item.employee_id,
    title: item.title,
    amount: Number(item.amount),
    description: item.description || '',
    attachment_url: item.attachment_url || '',
    attachment_file: null,
    status: item.status,
    start_km: item.start_km != null ? Number(item.start_km) : null,
    end_km: item.end_km != null ? Number(item.end_km) : null,
    total_km: item.total_km != null ? Number(item.total_km) : null,
    start_km_photo_url: item.start_km_photo_url || '',
    start_km_photo_file: null,
    end_km_photo_url: item.end_km_photo_url || '',
    end_km_photo_file: null,
    vehicle_plate_number: item.vehicle_plate_number || '',
    fuel_liters: item.fuel_liters != null ? Number(item.fuel_liters) : null,
  }
  isCRUDModalOpen.value = true
}

const onFileChange = (e: any) => {
  const file = e.target.files[0]
  if (file) {
    form.value.attachment_file = file
  }
}

const onStartKMFileChange = (e: any) => {
  const file = e.target.files[0]
  if (file) {
    form.value.start_km_photo_file = file
  }
}

const onEndKMFileChange = (e: any) => {
  const file = e.target.files[0]
  if (file) {
    form.value.end_km_photo_file = file
  }
}

const saveReimbursement = async () => {
  if (!form.value.employee_id) {
    toast.error('Harap pilih karyawan.')
    return
  }

  if (createCategory.value === 'GENERAL') {
    if (!form.value.title || !form.value.amount) {
      toast.error('Harap isi judul klaim dan nominal.')
      return
    }
  } else if (createCategory.value === 'BANNER_INSTALLATION') {
    // Banner Installation
    if (!isEditMode.value && selectedBannerIds.value.length === 0) {
      toast.error('Pilih minimal 1 spanduk terpasang untuk diajukan.')
      return
    }
    if (!form.value.title) {
      const emp = employees.value.find((e) => e.id === form.value.employee_id)
      const empName = emp ? `${emp.first_name} ${emp.last_name || ''}`.trim() : 'Karyawan'
      form.value.title = `Reimbursement Pasang Spanduk (${selectedBannerIds.value.length} Unit) - ${empName}`
    }
  } else if (createCategory.value === 'BBM') {
    if (!form.value.amount || Number(form.value.amount) <= 0) {
      toast.error('Harap masukkan nominal pembelian BBM.')
      return
    }
    if (form.value.start_km !== null && form.value.end_km !== null && Number(form.value.end_km) < Number(form.value.start_km)) {
      toast.error('KM Pulang tidak boleh lebih kecil dari KM Berangkat.')
      return
    }
    if (!form.value.title) {
      const emp = employees.value.find((e) => e.id === form.value.employee_id)
      const empName = emp ? `${emp.first_name} ${emp.last_name || ''}`.trim() : 'Karyawan'
      const kmInfo = form.value.total_km ? ` (${form.value.total_km} KM)` : ''
      form.value.title = `Klaim BBM - ${empName}${kmInfo}`
    }
  }

  isSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('employee_id', form.value.employee_id)
    formData.append('title', form.value.title)
    formData.append('amount', String(form.value.amount))
    formData.append('description', form.value.description || '')
    formData.append('status', form.value.status)
    formData.append('category', createCategory.value)

    if (createCategory.value === 'BANNER_INSTALLATION') {
      formData.append('banner_count', String(selectedBannerIds.value.length))
      formData.append('unit_rate', String(bannerUnitRateInput.value))
      formData.append('banner_ids', JSON.stringify(selectedBannerIds.value))
    }

    if (createCategory.value === 'BBM') {
      if (form.value.start_km !== null && form.value.start_km !== undefined) formData.append('start_km', String(form.value.start_km))
      if (form.value.end_km !== null && form.value.end_km !== undefined) formData.append('end_km', String(form.value.end_km))
      if (form.value.total_km !== null && form.value.total_km !== undefined) formData.append('total_km', String(form.value.total_km))
      if (form.value.vehicle_plate_number) formData.append('vehicle_plate_number', form.value.vehicle_plate_number)
      if (form.value.fuel_liters !== null && form.value.fuel_liters !== undefined) formData.append('fuel_liters', String(form.value.fuel_liters))

      if (form.value.start_km_photo_file) {
        formData.append('start_km_photo', toRaw(form.value.start_km_photo_file))
      }
      if (form.value.end_km_photo_file) {
        formData.append('end_km_photo', toRaw(form.value.end_km_photo_file))
      }
    }

    if (form.value.attachment_file) {
      formData.append('attachment', toRaw(form.value.attachment_file))
    }

    const config = {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }

    if (isEditMode.value) {
      await apiClient.put(`/reimbursements/manage/${form.value.id}`, formData, config)
      toast.success('Berhasil memperbarui reimbursement')
    } else {
      await apiClient.post('/reimbursements/manage', formData, config)
      toast.success('Berhasil menambahkan reimbursement')
    }
    isCRUDModalOpen.value = false
    fetchReimbursements()
    fetchUnclaimedBanners()
  } catch (error: any) {
    toast.error('Gagal menyimpan: ' + (error.response?.data?.message || error.message))
  } finally {
    isSubmitting.value = false
  }
}

const deleteReimbursement = async (id: string) => {
  if (!confirm('Hapus data reimbursement ini?')) return
  try {
    await apiClient.delete(`/reimbursements/manage/${id}`)
    toast.success('Berhasil menghapus reimbursement')
    fetchReimbursements()
    fetchUnclaimedBanners()
  } catch (error: any) {
    toast.error('Gagal menghapus: ' + (error.response?.data?.message || error.message))
  }
}

onMounted(() => {
  fetchReimbursements()
  fetchUnclaimedBanners()
  fetchEmployees()
})

// Auto-refresh when branch selection changes
watch(
  () => masterDataStore.selectedBranchId,
  () => {
    fetchReimbursements()
    fetchUnclaimedBanners()
  },
)

// Active branch label
const activeBranchLabel = computed(() => {
  const id = masterDataStore.selectedBranchId
  if (!id || id === 'ALL_BRANCHES') return 'Semua Cabang'
  return masterDataStore.branches.find((b) => b.id === id)?.name ?? 'Cabang Terpilih'
})

// Table Columns Definition
const columns = [
  {
    accessorKey: 'created_at',
    id: 'created_at',
    header: 'TANGGAL',
    cell: (info: any) =>
      h(
        'span',
        { class: 'text-slate-500 font-medium text-xs whitespace-nowrap' },
        info.getValue() ? new Date(info.getValue()).toLocaleDateString('id-ID', {
          day: '2-digit',
          month: 'short',
          year: 'numeric',
        }) : '-',
      ),
  },
  {
    accessorFn: (row: any) => {
      const emp = row.employee
      return emp ? `${emp.first_name} ${emp.last_name || ''}`.trim() : '-'
    },
    id: 'employee',
    header: 'KARYAWAN',
    cell: ({ row, getValue }: any) => {
      const emp = row.original.employee
      const branchName = emp?.branch?.name || ''
      const pos = emp?.job_position?.title || ''
      return h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'font-bold text-slate-900 text-xs leading-tight' }, getValue()),
        branchName || pos
          ? h('span', { class: 'text-[11px] text-slate-400 font-medium mt-0.5' }, [
              branchName,
              pos ? ` • ${pos}` : '',
            ])
          : null,
      ])
    },
  },
  {
    accessorKey: 'title',
    header: 'KATEGORI & JUDUL KLAIM',
    cell: ({ row, getValue }: any) => {
      const isBanner = row.original.category === 'BANNER_INSTALLATION'
      const isBBM = isBBMClaim(row.original)
      const count = row.original.banner_count || row.original.banner_orders?.length || 0
      const unitRate = Number(row.original.unit_rate || 0)
      const startKm = row.original.start_km
      const endKm = row.original.end_km
      const totalKm = row.original.total_km || (startKm != null && endKm != null ? endKm - startKm : null)
      const plate = row.original.vehicle_plate_number

      return h('div', { class: 'flex flex-col items-start gap-1.5 max-w-sm' }, [
        h('div', { class: 'flex items-center gap-1.5 flex-wrap' }, [
          isBanner
            ? h(
                'span',
                {
                  class:
                    'px-2 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 text-[10px] font-black rounded-md inline-flex items-center gap-1',
                },
                `🏷️ Pasang Spanduk (${count} Unit)`,
              )
            : isBBM
              ? h(
                'span',
                {
                  class:
                    'px-2 py-0.5 bg-amber-50 text-amber-800 border border-amber-200 text-[10px] font-black rounded-md inline-flex items-center gap-1',
                },
                '⛽ Klaim BBM',
              )
              : h(
                  'span',
                  {
                    class:
                      'px-2 py-0.5 bg-blue-50 text-blue-700 border border-blue-200 text-[10px] font-black rounded-md inline-flex items-center gap-1',
                  },
                  '📄 Operasional',
                ),
          isBanner && unitRate > 0
            ? h(
                'span',
                { class: 'text-[10px] text-emerald-600 font-bold bg-emerald-50/50 px-1.5 py-0.5 rounded' },
                `@ Rp ${unitRate.toLocaleString('id-ID')}`,
              )
            : null,
          isBBM && (startKm != null || endKm != null || totalKm != null)
            ? h(
                'span',
                { class: 'text-[10px] text-amber-800 font-bold bg-amber-100/60 border border-amber-200 px-1.5 py-0.5 rounded' },
                `KM: ${startKm != null ? Number(startKm).toLocaleString('id-ID') : '-'} → ${endKm != null ? Number(endKm).toLocaleString('id-ID') : '-'} (${totalKm != null ? Number(totalKm).toLocaleString('id-ID') : '-'} KM)`,
              )
            : null,
          isBBM && plate
            ? h(
                'span',
                { class: 'text-[10px] text-slate-600 font-bold bg-slate-100 px-1.5 py-0.5 rounded' },
                `🚗 ${plate}`,
              )
            : null,
        ]),
        h('span', { class: 'font-semibold text-slate-800 text-xs line-clamp-2' }, getValue() || '-'),
      ])
    },
  },
  {
    accessorKey: 'amount',
    header: 'NOMINAL',
    cell: ({ row, getValue }: any) => {
      const amt = Number(getValue())
      const isBanner = row.original.category === 'BANNER_INSTALLATION'
      if (isBanner && amt === 0 && row.original.status === 'PENDING') {
        return h('div', { class: 'flex flex-col' }, [
          h('span', { class: 'text-amber-600 font-bold text-xs italic' }, 'Menunggu Tarif Kasir'),
        ])
      }
      return h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'font-black text-slate-900 text-sm whitespace-nowrap' }, `Rp ${amt.toLocaleString('id-ID')}`),
      ])
    },
  },
  {
    accessorKey: 'status',
    header: 'STATUS REVIEW',
    cell: ({ row, getValue }: any) => {
      const val = (getValue() as string) || 'PENDING'
      let badgeCls = 'bg-amber-100 text-amber-800 border-amber-200'
      let label = 'PENDING'
      if (val === 'APPROVED') {
        badgeCls = 'bg-emerald-100 text-emerald-800 border-emerald-200'
        label = 'DISETUJUI'
      } else if (val === 'REJECTED') {
        badgeCls = 'bg-rose-100 text-rose-800 border-rose-200'
        label = 'DITOLAK'
      }

      const approver = row.original.approver
      const approverName = approver ? `${approver.first_name || ''} ${approver.last_name || ''}`.trim() : ''

      return h('div', { class: 'flex flex-col items-start gap-1' }, [
        h(
          'span',
          { class: `${badgeCls} px-2.5 py-0.5 text-[10px] font-black rounded-full border inline-block` },
          label,
        ),
        val === 'APPROVED' && approverName
          ? h('span', { class: 'text-[10px] text-slate-500 font-medium leading-none' }, `oleh ${approverName}`)
          : null,
      ])
    },
  },
  {
    accessorKey: 'disbursement_status',
    header: 'PENCAIRAN',
    cell: ({ row }: any) => {
      const status = row.original.status
      const disbStatus = row.original.disbursement_status || 'UNPAID'
      const disbVia = row.original.disbursed_via || ''

      if (disbStatus === 'DISBURSED') {
        let viaLabel = 'Cair'
        const kasPabrik = row.original.kas_pabrik?.kas_name
        const expenseAcc = row.original.expense_account
        const accLabel = expenseAcc ? `${expenseAcc.code} - ${expenseAcc.name}` : ''

        if (disbVia === 'PETTY_CASH') {
          viaLabel = kasPabrik ? `Petty Cash Pabrik (${kasPabrik})` : 'Kas Kecil'
        } else if (disbVia === 'CASH') {
          viaLabel = 'Kas Tunai'
        } else if (disbVia === 'BANK_TRANSFER') {
          viaLabel = 'Transfer Bank'
        }

        return h('div', { class: 'flex flex-col items-start gap-1 max-w-[180px]' }, [
          h(
            'span',
            {
              class:
                'px-2 py-0.5 bg-emerald-50 text-emerald-800 border border-emerald-200 text-[10px] font-black rounded-md inline-flex items-center gap-1 leading-tight',
            },
            `✓ ${viaLabel}`,
          ),
          accLabel
            ? h('span', { class: 'text-[9px] text-slate-500 font-semibold truncate max-w-full leading-none' }, `Pos: ${accLabel}`)
            : null,
        ])
      }

      if (status === 'APPROVED') {
        return h(
          'span',
          {
            class:
              'px-2 py-0.5 bg-amber-50 text-amber-700 border border-amber-200 text-[10px] font-black rounded-md inline-flex items-center gap-1',
          },
          'Siap Dicairkan',
        )
      }

      return h('span', { class: 'text-[11px] text-slate-400 font-medium' }, 'Belum Cair')
    },
  },
  {
    id: 'attachment',
    header: 'BUKTI / FOTO',
    cell: ({ row }: any) => {
      const item = row.original
      const isBanner = item.category === 'BANNER_INSTALLATION'
      const isBBM = isBBMClaim(item)
      const firstBannerPhoto = item.banner_orders?.[0]?.photo_installed
      const url = isBanner
        ? (firstBannerPhoto || item.attachment_url)
        : (item.attachment_url || item.start_km_photo_url || item.end_km_photo_url)

      if (!url) return h('span', { class: 'text-slate-400 italic text-[11px]' }, 'Tidak ada')

      return h(
        'button',
        {
          onClick: () =>
            openLightbox(
              url,
              isBanner
                ? 'Foto Pemasangan Spanduk'
                : isBBM
                  ? (item.attachment_url ? 'Struk Pembelian BBM' : item.start_km_photo_url ? 'Foto Odometer KM Awal' : 'Foto Odometer KM Akhir')
                  : 'Nota / Bukti Pembayaran',
            ),
          class:
            'inline-flex items-center gap-1.5 px-2.5 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-bold transition-all',
        },
        [
          h(ImageIcon, { class: 'w-3.5 h-3.5 text-primary' }),
          isBanner ? 'Foto Pasang' : isBBM ? 'Struk BBM' : 'Lihat Nota',
        ],
      )
    },
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const item = row.original
      return h('div', { class: 'flex items-center gap-1' }, [
        // Detail Action (for all)
        h(
          'button',
          {
            class: 'p-1.5 rounded-lg hover:bg-slate-100 text-slate-600 transition-colors',
            title: 'Lihat Detail Lengkap',
            onClick: () => openDetailModal(item),
          },
          h(Eye, { class: 'w-4 h-4' }),
        ),

        // Review Action (Only for PENDING)
        item.status === 'PENDING'
          ? h(
              'button',
              {
                class: 'p-1.5 rounded-lg bg-amber-50 hover:bg-amber-100 text-amber-700 transition-colors font-bold text-xs flex items-center gap-1',
                title: 'Review & Persetujuan',
                onClick: () => openReviewModal(item),
              },
              [h(CheckCircle2, { class: 'w-3.5 h-3.5 text-amber-600' }), 'Review'],
            )
          : null,

        // Edit Action (For General only or if not disbursed)
        item.disbursement_status !== 'DISBURSED' && item.status === 'PENDING'
          ? h(
              'button',
              {
                class: 'p-1.5 rounded-lg hover:bg-blue-50 text-blue-600 transition-colors',
                title: 'Edit Pengajuan',
                onClick: () => openEditModal(item),
              },
              h(Pencil, { class: 'w-3.5 h-3.5' }),
            )
          : null,

        // Delete Action
        item.disbursement_status !== 'DISBURSED'
          ? h(
              'button',
              {
                class: 'p-1.5 rounded-lg hover:bg-rose-50 text-rose-600 transition-colors',
                title: 'Hapus',
                onClick: () => deleteReimbursement(item.id),
              },
              h(Trash2, { class: 'w-3.5 h-3.5' }),
            )
          : null,
      ])
    },
    enableSorting: false,
  },
]
</script>

<template>
  <div class="space-y-6">
    <!-- Header Page -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <h1 class="text-2xl font-black text-slate-900 tracking-tight">
            Pusat Persetujuan Reimbursement
          </h1>
          <span class="px-2.5 py-0.5 bg-primary/10 text-primary text-xs font-black rounded-full">
            MANAGEMENT
          </span>
        </div>
        <p class="text-xs md:text-sm font-medium text-slate-500 mt-1">
          Kelola dan setujui klaim operasional serta reimbursement pemasangan spanduk lapangan secara real-time.
        </p>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-2.5">
        <Button
          @click="openAddModal()"
          class="gap-2 h-11 px-5 rounded-2xl text-xs md:text-sm font-bold shadow-lg shadow-primary/20"
        >
          <Plus class="w-4 h-4" />
          Tambah Pengajuan
        </Button>
      </div>
    </div>

    <!-- KPI Statistic Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-4">
      <!-- Total Nominal -->
      <div class="p-4 md:p-5 rounded-3xl bg-white border border-slate-200/80 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between">
          <span class="text-[10px] md:text-xs font-black text-slate-400 uppercase tracking-wider">
            Total Nilai Klaim
          </span>
          <div class="w-8 h-8 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center">
            <DollarSign class="w-4 h-4" />
          </div>
        </div>
        <div class="mt-3">
          <p class="text-lg md:text-xl font-black text-slate-900 truncate">
            {{ formatRupiah(stats.totalNominal) }}
          </p>
          <p class="text-[10px] md:text-xs text-slate-500 mt-0.5">
            {{ stats.totalAll }} Pengajuan Diajukan
          </p>
        </div>
      </div>

      <!-- Menunggu Review -->
      <div class="p-4 md:p-5 rounded-3xl bg-amber-50/60 border border-amber-200/80 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between">
          <span class="text-[10px] md:text-xs font-black text-amber-700 uppercase tracking-wider">
            Menunggu Review
          </span>
          <div class="w-8 h-8 rounded-xl bg-amber-100 text-amber-700 flex items-center justify-center">
            <Clock class="w-4 h-4" />
          </div>
        </div>
        <div class="mt-3">
          <p class="text-lg md:text-xl font-black text-amber-900">
            {{ stats.pendingCount }} Pengajuan
          </p>
          <p class="text-[10px] md:text-xs text-amber-700 font-medium mt-0.5">
            Perlu persetujuan admin/kasir
          </p>
        </div>
      </div>

      <!-- Pasang Spanduk -->
      <div class="p-4 md:p-5 rounded-3xl bg-emerald-50/60 border border-emerald-200/80 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between">
          <span class="text-[10px] md:text-xs font-black text-emerald-800 uppercase tracking-wider">
            Spanduk Terpasang
          </span>
          <div class="w-8 h-8 rounded-xl bg-emerald-100 text-emerald-700 flex items-center justify-center">
            <Layers class="w-4 h-4" />
          </div>
        </div>
        <div class="mt-3">
          <p class="text-lg md:text-xl font-black text-emerald-950">
            {{ stats.totalInstalledBannersCount }} Spanduk
          </p>
          <p class="text-[10px] md:text-xs text-emerald-700 font-medium mt-0.5">
            Dari {{ stats.totalBanner }} klaim spanduk
          </p>
        </div>
      </div>

      <!-- Spanduk Belum Diklaim -->
      <div
        @click="activeTab = 'UNCLAIMED'"
        class="p-4 md:p-5 rounded-3xl bg-slate-900 text-white shadow-sm flex flex-col justify-between cursor-pointer hover:bg-slate-800 transition-all group"
      >
        <div class="flex items-center justify-between">
          <span class="text-[10px] md:text-xs font-black text-slate-400 uppercase tracking-wider">
            Spanduk Siap Klaim
          </span>
          <div class="w-8 h-8 rounded-xl bg-white/10 text-emerald-400 flex items-center justify-center group-hover:scale-110 transition-transform">
            <Sparkles class="w-4 h-4" />
          </div>
        </div>
        <div class="mt-3">
          <p class="text-lg md:text-xl font-black text-white">
            {{ stats.totalUnclaimed }} Unit Belum Klaim
          </p>
          <p class="text-[10px] md:text-xs text-emerald-400 font-semibold mt-0.5 flex items-center gap-1">
            Lihat daftar spanduk &rarr;
          </p>
        </div>
      </div>
    </div>

    <!-- Main Navigation Tabs & Filters Bar -->
    <div class="bg-white p-4 rounded-3xl border border-slate-200/80 shadow-sm space-y-4">
      <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <!-- Segmented Tab Controls -->
        <div class="flex items-center p-1 bg-slate-100 rounded-2xl overflow-x-auto custom-scrollbar">
          <button
            @click="activeTab = 'ALL'"
            :class="[
              'px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center gap-2 shrink-0',
              activeTab === 'ALL'
                ? 'bg-white text-slate-900 shadow-sm'
                : 'text-slate-500 hover:text-slate-900',
            ]"
          >
            Semua Klaim
            <span class="px-1.5 py-0.5 text-[10px] rounded-full bg-slate-200 text-slate-700">
              {{ stats.totalAll }}
            </span>
          </button>

          <button
            @click="activeTab = 'GENERAL'"
            :class="[
              'px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center gap-2 shrink-0',
              activeTab === 'GENERAL'
                ? 'bg-white text-blue-600 shadow-sm'
                : 'text-slate-500 hover:text-slate-900',
            ]"
          >
            📄 Operasional
            <span class="px-1.5 py-0.5 text-[10px] rounded-full bg-blue-100 text-blue-700">
              {{ stats.totalGeneral }}
            </span>
          </button>

          <button
            @click="activeTab = 'BBM'"
            :class="[
              'px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center gap-2 shrink-0',
              activeTab === 'BBM'
                ? 'bg-white text-amber-800 shadow-sm'
                : 'text-slate-500 hover:text-slate-900',
            ]"
          >
            ⛽ Klaim BBM
            <span class="px-1.5 py-0.5 text-[10px] rounded-full bg-amber-100 text-amber-800">
              {{ stats.totalBBM }}
            </span>
          </button>

          <button
            @click="activeTab = 'BANNER_INSTALLATION'"
            :class="[
              'px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center gap-2 shrink-0',
              activeTab === 'BANNER_INSTALLATION'
                ? 'bg-white text-emerald-600 shadow-sm'
                : 'text-slate-500 hover:text-slate-900',
            ]"
          >
            🏷️ Pasang Spanduk
            <span class="px-1.5 py-0.5 text-[10px] rounded-full bg-emerald-100 text-emerald-700">
              {{ stats.totalBanner }}
            </span>
          </button>

          <button
            @click="activeTab = 'UNCLAIMED'"
            :class="[
              'px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center gap-2 shrink-0',
              activeTab === 'UNCLAIMED'
                ? 'bg-slate-900 text-white shadow-sm'
                : 'text-slate-500 hover:text-slate-900',
            ]"
          >
            ✨ Spanduk Belum Diklaim
            <span class="px-1.5 py-0.5 text-[10px] rounded-full bg-amber-400 text-slate-900 font-bold">
              {{ stats.totalUnclaimed }}
            </span>
          </button>
        </div>

        <!-- Branch Indicator & Status Filter -->
        <div class="flex items-center gap-2.5 flex-wrap">
          <div class="flex items-center gap-2 px-3 py-1.5 rounded-xl border bg-slate-50 text-slate-700 text-xs font-bold">
            <Building2 class="w-3.5 h-3.5 text-primary" />
            <span>{{ activeBranchLabel }}</span>
          </div>

          <!-- Status Select (when on Reimbursement tabs) -->
          <div v-if="activeTab !== 'UNCLAIMED'" class="w-36">
            <Select v-model="selectedStatus" @update:model-value="fetchReimbursements">
              <SelectTrigger class="h-9 rounded-xl text-xs font-bold border-slate-200">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ALL">Semua Status</SelectItem>
                <SelectItem value="PENDING">PENDING</SelectItem>
                <SelectItem value="APPROVED">DISETUJUI</SelectItem>
                <SelectItem value="REJECTED">DITOLAK</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      <!-- Search Input Bar -->
      <div class="relative">
        <Search class="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
        <Input
          v-model="searchQuery"
          placeholder="Cari berdasarkan nama karyawan, judul klaim, atau nama toko spanduk..."
          class="pl-10 h-10 rounded-2xl border-slate-200 text-xs md:text-sm bg-slate-50/60 focus:bg-white"
        />
        <button
          v-if="searchQuery"
          @click="searchQuery = ''"
          class="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
        >
          <X class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- TAB CONTENT: Reimbursements Table (ALL, GENERAL, BANNER) -->
    <div v-if="activeTab !== 'UNCLAIMED'">
      <DataTable
        :data="filteredReimbursements"
        :columns="columns"
        :isLoading="isLoading"
      />
    </div>

    <!-- TAB CONTENT: Unclaimed Banners View -->
    <div v-else class="space-y-4">
      <div class="p-4 bg-emerald-50/60 border border-emerald-200 rounded-3xl flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div>
          <h3 class="text-sm font-black text-emerald-950 flex items-center gap-1.5">
            <Sparkles class="w-4 h-4 text-emerald-600" />
            Spanduk Selesai Terpasang & Belum Diajukan Klaim
          </h3>
          <p class="text-xs text-emerald-800 font-medium mt-0.5">
            Daftar ini adalah spanduk yang telah diunggah bukti pasangnya oleh tim lapangan dan siap dibuatkan pengajuan reimbursement.
          </p>
        </div>
      </div>

      <div v-if="isLoadingUnclaimed" class="p-12 text-center text-slate-400">
        <div class="animate-spin w-8 h-8 border-4 border-primary border-t-transparent rounded-full mx-auto mb-2"></div>
        <p class="text-xs font-bold">Memuat data spanduk...</p>
      </div>

      <div
        v-else-if="filteredUnclaimedBanners.length === 0"
        class="p-12 text-center bg-white rounded-3xl border border-slate-200"
      >
        <p class="text-sm font-bold text-slate-600">Tidak ada spanduk terpasang yang belum diklaim.</p>
        <p class="text-xs text-slate-400 mt-1">Semua spanduk terpasang sudah memiliki klaim reimbursement.</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="banner in filteredUnclaimedBanners"
          :key="banner.id"
          class="bg-white rounded-3xl p-4 border border-slate-200 shadow-sm flex flex-col justify-between gap-3 hover:shadow-md transition-all"
        >
          <div class="flex gap-3">
            <div class="relative group cursor-pointer shrink-0">
              <img
                v-if="banner.photo_installed"
                :src="banner.photo_installed.startsWith('http') ? banner.photo_installed : `${API_BASE}${banner.photo_installed}`"
                alt="Foto Spanduk"
                class="w-20 h-20 object-cover rounded-2xl border border-slate-200 group-hover:opacity-90"
                @click="openLightbox(banner.photo_installed, banner.store_name || banner.shop_name)"
              />
              <div
                v-else
                class="w-20 h-20 rounded-2xl bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-400"
              >
                <ImageIcon class="w-6 h-6" />
              </div>
            </div>

            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-1.5 flex-wrap">
                <span class="px-2 py-0.5 bg-emerald-100 text-emerald-800 text-[10px] font-black rounded-md">
                  {{ banner.banner_type || 'Standard' }}
                </span>
                <span class="text-[10px] font-bold text-slate-500">
                  {{ banner.size_length || 3 }}x{{ banner.size_width || 1 }}m
                </span>
              </div>
              <h4 class="font-bold text-slate-900 text-sm truncate mt-1">
                {{ banner.store_name || banner.shop_name || 'Toko Unknown' }}
              </h4>
              <p class="text-[11px] text-slate-500 line-clamp-1 mt-0.5">
                📍 {{ banner.location || banner.shop_address || 'Lokasi tidak terdata' }}
              </p>
              <p class="text-[10px] text-slate-400 font-semibold mt-1">
                Sales: {{ banner.employee ? `${banner.employee.first_name} ${banner.employee.last_name || ''}` : banner.sales_name || '-' }}
              </p>
            </div>
          </div>

          <div class="flex items-center justify-between pt-2 border-t border-slate-100">
            <span class="text-[10px] text-slate-400 font-medium">
              Pasang: {{ banner.installed_at ? new Date(banner.installed_at).toLocaleDateString('id-ID') : '-' }}
            </span>

            <Button
              size="sm"
              @click="openAddModal(banner.employee_id || banner.installer_id || banner.operator_id, banner.id)"
              class="h-8 px-3 rounded-xl text-xs font-bold gap-1 bg-emerald-600 hover:bg-emerald-700 text-white"
            >
              <Plus class="w-3.5 h-3.5" />
              Buat Klaim
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- DETAIL MODAL (Full Inspection for Any Status) -->
    <Dialog v-model:open="isDetailModalOpen">
      <DialogContent
        class="sm:max-w-2xl rounded-[2rem] md:rounded-[3rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <DialogHeader class="p-6 md:p-8 bg-slate-900 text-white relative shrink-0">
          <div class="flex items-center justify-between">
            <div>
              <div class="flex items-center gap-2">
                <span
                  :class="[
                    'px-2.5 py-0.5 text-[10px] font-black rounded-md border uppercase',
                    detailItem?.category === 'BANNER_INSTALLATION'
                      ? 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30'
                      : isBBMClaim(detailItem)
                        ? 'bg-amber-500/20 text-amber-400 border-amber-500/30'
                        : 'bg-blue-500/20 text-blue-400 border-blue-500/30',
                  ]"
                >
                  {{ detailItem?.category === 'BANNER_INSTALLATION' ? '🏷️ Pasang Spanduk' : isBBMClaim(detailItem) ? '⛽ Klaim BBM' : '📄 Operasional' }}
                </span>
                <span
                  :class="[
                    'px-2.5 py-0.5 text-[10px] font-black rounded-full uppercase',
                    detailItem?.status === 'APPROVED' ? 'bg-emerald-600 text-white' : detailItem?.status === 'REJECTED' ? 'bg-red-600 text-white' : 'bg-amber-500 text-slate-900',
                  ]"
                >
                  {{ detailItem?.status }}
                </span>
              </div>
              <DialogTitle class="text-lg md:text-xl font-black tracking-tight mt-2 text-white">
                {{ detailItem?.title }}
              </DialogTitle>
            </div>
          </div>
        </DialogHeader>

        <div v-if="detailItem" class="p-6 md:p-8 space-y-6 overflow-y-auto custom-scrollbar flex-1 bg-slate-50">
          <!-- Employee & Finance Summary Card -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div class="p-4 bg-white rounded-2xl border border-slate-200">
              <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider block">Karyawan</span>
              <p class="font-bold text-slate-900 text-sm truncate mt-1">
                {{ detailItem.employee?.first_name }} {{ detailItem.employee?.last_name || '' }}
              </p>
              <span class="text-[11px] text-slate-500">{{ detailItem.employee?.branch?.name || '-' }}</span>
            </div>

            <div class="p-4 bg-white rounded-2xl border border-slate-200">
              <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider block">Nominal Klaim</span>
              <p class="font-black text-primary text-base md:text-lg mt-1">
                Rp {{ Number(detailItem.amount).toLocaleString('id-ID') }}
              </p>
              <span v-if="detailItem.unit_rate > 0" class="text-[10px] text-slate-500">
                @ Rp {{ Number(detailItem.unit_rate).toLocaleString('id-ID') }}/spanduk
              </span>
            </div>

            <div class="p-4 bg-white rounded-2xl border border-slate-200">
              <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider block">Pencairan</span>
              <p class="font-bold text-slate-900 text-xs mt-1">
                {{ detailItem.disbursement_status === 'DISBURSED' ? '✓ Sudah Dicairkan' : 'Belum Dicairkan' }}
              </p>
              <div v-if="detailItem.disbursement_status === 'DISBURSED'" class="flex flex-col gap-0.5 mt-1">
                <span class="text-[10px] text-emerald-700 font-bold">
                  {{
                    detailItem.kas_pabrik
                      ? `Petty Cash Pabrik (${detailItem.kas_pabrik.kas_name})`
                      : detailItem.bank_account
                        ? `Bank ${detailItem.bank_account.bank_name}`
                        : detailItem.disbursed_via === 'CASH'
                          ? 'Kas Tunai Cabang'
                          : 'Kas Kecil Cabang'
                  }}
                </span>
                <span v-if="detailItem.expense_account" class="text-[9px] text-slate-500 font-medium">
                  Pos: {{ detailItem.expense_account.code }} - {{ detailItem.expense_account.name }}
                </span>
              </div>
            </div>
          </div>

          <!-- BBM ODOMETER VALIDATION SECTION -->
          <div
            v-if="isBBMClaim(detailItem)"
            class="p-5 bg-amber-50/70 border border-amber-200/90 rounded-3xl space-y-4"
          >
            <div class="flex items-center justify-between">
              <h4 class="text-xs font-black text-amber-950 uppercase tracking-wider flex items-center gap-1.5">
                <span>⛽</span> Data Odometer & Validasi Perjalanan BBM
              </h4>
              <span v-if="detailItem.vehicle_plate_number" class="px-2.5 py-1 bg-amber-200/80 text-amber-900 text-[10px] font-black rounded-lg">
                🚗 Plat: {{ detailItem.vehicle_plate_number }}
              </span>
            </div>

            <!-- KM Metrics Grid: KM Awal, KM Akhir, Total Jarak -->
            <div class="grid grid-cols-3 gap-3">
              <div class="p-3 bg-white rounded-2xl border border-amber-200/80 text-center">
                <span class="text-[10px] font-bold text-slate-400 uppercase block">KM Awal</span>
                <p class="text-base font-black text-slate-800 mt-0.5">
                  {{ detailItem.start_km != null ? Number(detailItem.start_km).toLocaleString('id-ID') : '-' }} <span class="text-[10px] font-normal text-slate-400">KM</span>
                </p>
              </div>

              <div class="p-3 bg-white rounded-2xl border border-amber-200/80 text-center">
                <span class="text-[10px] font-bold text-slate-400 uppercase block">KM Akhir</span>
                <p class="text-base font-black text-slate-800 mt-0.5">
                  {{ detailItem.end_km != null ? Number(detailItem.end_km).toLocaleString('id-ID') : '-' }} <span class="text-[10px] font-normal text-slate-400">KM</span>
                </p>
              </div>

              <div class="p-3 bg-amber-100/60 rounded-2xl border border-amber-300 text-center">
                <span class="text-[10px] font-black text-amber-800 uppercase block">Total Jarak</span>
                <p class="text-base font-black text-amber-950 mt-0.5">
                  {{ detailItem.total_km != null ? Number(detailItem.total_km).toLocaleString('id-ID') : (detailItem.start_km != null && detailItem.end_km != null ? (Number(detailItem.end_km) - Number(detailItem.start_km)).toLocaleString('id-ID') : '-') }} <span class="text-[10px] font-bold text-amber-700">KM</span>
                </p>
              </div>
            </div>

            <div v-if="detailItem.start_km == null && detailItem.end_km == null" class="p-2.5 bg-amber-100/60 rounded-xl border border-amber-300 text-amber-900 text-xs font-semibold flex items-center gap-2">
              <span>ℹ️</span> Data KM Awal & KM Akhir tidak dicantumkan oleh pemohon.
            </div>

            <div v-if="detailItem.fuel_liters" class="text-xs text-amber-900 font-medium">
              ⛽ Estimasi / Pengisian BBM: <strong>{{ detailItem.fuel_liters }} Liter</strong>
            </div>

            <!-- Photos: Start KM, End KM, Struk Nota -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-1">
              <!-- Start KM Photo -->
              <div class="p-3 bg-white rounded-2xl border border-amber-200/80 flex flex-col items-center text-center">
                <span class="text-[10px] font-black text-slate-500 uppercase tracking-wider mb-2">Foto Odometer KM Awal</span>
                <div
                  v-if="detailItem.start_km_photo_url"
                  class="cursor-pointer group relative w-full h-28 rounded-xl overflow-hidden border border-slate-200"
                  @click="openLightbox(detailItem.start_km_photo_url, 'Foto Odometer KM Awal')"
                >
                  <img
                    :src="resolvePhotoUrl(detailItem.start_km_photo_url)"
                    alt="KM Awal"
                    class="w-full h-full object-cover group-hover:scale-105 transition-transform"
                  />
                  <div class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white text-xs font-bold transition-opacity">
                    Zoom
                  </div>
                </div>
                <div v-else class="w-full h-28 rounded-xl bg-slate-100 flex items-center justify-center text-slate-400 text-xs italic">
                  Tidak ada foto
                </div>
              </div>

              <!-- End KM Photo -->
              <div class="p-3 bg-white rounded-2xl border border-amber-200/80 flex flex-col items-center text-center">
                <span class="text-[10px] font-black text-slate-500 uppercase tracking-wider mb-2">Foto Odometer KM Akhir</span>
                <div
                  v-if="detailItem.end_km_photo_url"
                  class="cursor-pointer group relative w-full h-28 rounded-xl overflow-hidden border border-slate-200"
                  @click="openLightbox(detailItem.end_km_photo_url, 'Foto Odometer KM Akhir')"
                >
                  <img
                    :src="resolvePhotoUrl(detailItem.end_km_photo_url)"
                    alt="KM Akhir"
                    class="w-full h-full object-cover group-hover:scale-105 transition-transform"
                  />
                  <div class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white text-xs font-bold transition-opacity">
                    Zoom
                  </div>
                </div>
                <div v-else class="w-full h-28 rounded-xl bg-slate-100 flex items-center justify-center text-slate-400 text-xs italic">
                  Tidak ada foto
                </div>
              </div>

              <!-- Receipt Photo -->
              <div class="p-3 bg-white rounded-2xl border border-amber-200/80 flex flex-col items-center text-center">
                <span class="text-[10px] font-black text-slate-500 uppercase tracking-wider mb-2">Struk / Nota SPBU</span>
                <div
                  v-if="detailItem.attachment_url"
                  class="cursor-pointer group relative w-full h-28 rounded-xl overflow-hidden border border-slate-200"
                  @click="openLightbox(detailItem.attachment_url, 'Struk Pembelian BBM')"
                >
                  <img
                    :src="resolvePhotoUrl(detailItem.attachment_url)"
                    alt="Struk BBM"
                    class="w-full h-full object-cover group-hover:scale-105 transition-transform"
                  />
                  <div class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 flex items-center justify-center text-white text-xs font-bold transition-opacity">
                    Zoom
                  </div>
                </div>
                <div v-else class="w-full h-28 rounded-xl bg-slate-100 flex items-center justify-center text-slate-400 text-xs italic">
                  Tidak ada nota
                </div>
              </div>
            </div>
          </div>

          <!-- Description -->
          <div class="p-4 bg-white rounded-2xl border border-slate-200 space-y-1">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider">Catatan / Deskripsi</span>
            <p class="text-xs md:text-sm text-slate-700 leading-relaxed">
              {{ detailItem.description || 'Tidak ada catatan tambahan.' }}
            </p>
          </div>

          <!-- Banner Orders List if Banner Category -->
          <div v-if="detailItem.category === 'BANNER_INSTALLATION'" class="space-y-3">
            <div class="flex items-center justify-between">
              <h4 class="text-xs font-black text-slate-800 uppercase tracking-wider">
                Daftar Spanduk Terpasang ({{ detailItem.banner_orders?.length || detailItem.banner_count || 0 }} Unit)
              </h4>
            </div>

            <div v-if="detailItem.banner_orders && detailItem.banner_orders.length > 0" class="space-y-3">
              <div
                v-for="(b, bIdx) in detailItem.banner_orders"
                :key="b.id || bIdx"
                class="p-4 bg-white border border-slate-200 rounded-2xl flex flex-col sm:flex-row gap-4 items-start sm:items-center"
              >
                <div class="relative cursor-pointer shrink-0" @click="openLightbox(b.photo_installed, b.store_name || b.shop_name)">
                  <img
                    v-if="b.photo_installed"
                    :src="b.photo_installed.startsWith('http') ? b.photo_installed : `${API_BASE}${b.photo_installed}`"
                    alt="Bukti Pasang"
                    class="w-16 h-16 object-cover rounded-xl border border-slate-200 hover:opacity-90 transition-opacity"
                  />
                  <div v-else class="w-16 h-16 rounded-xl bg-slate-100 flex items-center justify-center text-slate-400">
                    <ImageIcon class="w-5 h-5" />
                  </div>
                </div>

                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <span class="font-bold text-slate-900 text-sm truncate">
                      {{ b.store_name || b.shop_name || 'Toko' }}
                    </span>
                    <span class="px-2 py-0.5 bg-slate-100 text-slate-600 text-[10px] font-bold rounded">
                      {{ b.size_length || 3 }}x{{ b.size_width || 1 }}m
                    </span>
                  </div>
                  <p class="text-xs text-slate-500 line-clamp-1 mt-0.5">
                    📍 {{ b.location || b.shop_address || '-' }}
                  </p>
                  <p v-if="b.installed_at" class="text-[10px] text-slate-400 mt-1">
                    Dipasang pada: {{ new Date(b.installed_at).toLocaleString('id-ID') }}
                  </p>
                </div>

                <div class="flex sm:flex-col items-center gap-2 shrink-0">
                  <Button
                    v-if="b.photo_installed"
                    size="sm"
                    variant="outline"
                    @click="openLightbox(b.photo_installed, b.store_name || b.shop_name)"
                    class="h-8 rounded-xl text-xs font-bold gap-1"
                  >
                    <ZoomIn class="w-3.5 h-3.5" />
                    Foto Penuh
                  </Button>
                </div>
              </div>
            </div>

            <div v-else class="p-6 bg-white border border-slate-200 rounded-2xl text-center text-slate-400 text-xs">
              Tidak ada rincian data spanduk terhubung.
            </div>
          </div>

          <!-- General Attachment if exists (for non-BBM or if BBM already rendered above) -->
          <div v-else-if="!isBBMClaim(detailItem) && detailItem.attachment_url" class="space-y-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider">Foto Bukti Nota / Lampiran</span>
            <div
              @click="openLightbox(detailItem.attachment_url, 'Lampiran Nota')"
              class="p-4 bg-white border border-slate-200 rounded-2xl cursor-pointer hover:bg-slate-100/50 transition-all text-center"
            >
              <img
                :src="resolvePhotoUrl(detailItem.attachment_url)"
                alt="Nota Bukti"
                class="max-h-56 mx-auto object-contain rounded-xl"
              />
              <span class="text-xs font-bold text-primary mt-2 inline-block">Klik untuk melihat ukuran penuh</span>
            </div>
          </div>

          <!-- Approver / Rejection Details -->
          <div v-if="detailItem.rejected_reason" class="p-4 bg-rose-50 border border-rose-200 rounded-2xl space-y-1">
            <span class="text-[10px] font-black text-rose-700 uppercase tracking-wider">Alasan Penolakan</span>
            <p class="text-xs text-rose-900 font-medium leading-relaxed">{{ detailItem.rejected_reason }}</p>
          </div>
        </div>

        <DialogFooter class="p-4 md:p-6 bg-white border-t border-slate-200 flex justify-end">
          <Button variant="ghost" @click="isDetailModalOpen = false" class="rounded-xl font-bold text-xs h-10 px-6">
            TUTUP
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- REVIEW MODAL (Approval & Banner Rate / BBM KM Inspection) -->
    <Dialog v-model:open="isModalOpen">
      <DialogContent
        class="sm:max-w-xl rounded-[2rem] md:rounded-[3rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <DialogHeader class="p-6 md:p-8 bg-slate-900 text-white relative shrink-0">
          <DialogTitle class="text-xl font-black tracking-tight">Review Pengajuan Klaim</DialogTitle>
          <DialogDescription class="text-slate-400 text-xs mt-1">
            Tinjau nominal, rincian odometer BBM, dan bukti spanduk sebelum memberikan keputusan persetujuan.
          </DialogDescription>
        </DialogHeader>

        <div v-if="selectedItem" class="p-6 md:p-8 space-y-6 overflow-y-auto custom-scrollbar flex-1 bg-white">
          <!-- Summary Info -->
          <div class="grid grid-cols-2 gap-3">
            <div class="p-3.5 bg-slate-50 rounded-2xl border border-slate-100">
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Karyawan</p>
              <p class="font-bold text-slate-900 text-xs truncate">
                {{ selectedItem.employee?.first_name }} {{ selectedItem.employee?.last_name || '' }}
              </p>
            </div>
            <div class="p-3.5 bg-slate-50 rounded-2xl border border-slate-100">
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Kategori</p>
              <p class="font-bold text-slate-900 text-xs truncate">
                {{ selectedItem.category === 'BANNER_INSTALLATION' ? '🏷️ Pasang Spanduk' : isBBMClaim(selectedItem) ? '⛽ Klaim BBM' : '📄 Operasional' }}
              </p>
            </div>
          </div>

          <!-- BBM KM & Odometer Validation for Cashier Review -->
          <div
            v-if="isBBMClaim(selectedItem)"
            class="p-5 bg-amber-50/80 border border-amber-200 rounded-3xl space-y-4"
          >
            <div class="flex items-center justify-between">
              <div>
                <h4 class="text-xs font-black text-amber-950 uppercase tracking-wider flex items-center gap-1.5">
                  <span>⛽</span> Validasi Odometer & Perjalanan BBM (Kasir)
                </h4>
                <p class="text-[11px] text-amber-800 font-medium mt-0.5">
                  Verifikasi kecocokan KM sebelum memberikan persetujuan.
                </p>
              </div>
              <span v-if="selectedItem.vehicle_plate_number" class="px-2.5 py-1 bg-amber-200 text-amber-950 text-[10px] font-black rounded-lg">
                🚗 {{ selectedItem.vehicle_plate_number }}
              </span>
            </div>

            <!-- KM Odometer Summary: KM Awal, KM Akhir, Total Jarak -->
            <div class="grid grid-cols-3 gap-2 text-center">
              <div class="p-2.5 bg-white rounded-xl border border-amber-200">
                <span class="text-[10px] font-bold text-slate-400 block">KM Awal</span>
                <span class="text-sm font-black text-slate-800">{{ selectedItem.start_km != null ? Number(selectedItem.start_km).toLocaleString('id-ID') : '-' }}</span>
              </div>
              <div class="p-2.5 bg-white rounded-xl border border-amber-200">
                <span class="text-[10px] font-bold text-slate-400 block">KM Akhir</span>
                <span class="text-sm font-black text-slate-800">{{ selectedItem.end_km != null ? Number(selectedItem.end_km).toLocaleString('id-ID') : '-' }}</span>
              </div>
              <div class="p-2.5 bg-amber-100/80 rounded-xl border border-amber-300">
                <span class="text-[10px] font-black text-amber-800 block">Total Jarak</span>
                <span class="text-sm font-black text-amber-950">
                  {{ selectedItem.total_km != null ? Number(selectedItem.total_km).toLocaleString('id-ID') : (selectedItem.start_km != null && selectedItem.end_km != null ? (Number(selectedItem.end_km) - Number(selectedItem.start_km)).toLocaleString('id-ID') : '-') }} KM
                </span>
              </div>
            </div>

            <div v-if="selectedItem.start_km == null && selectedItem.end_km == null" class="p-2.5 bg-amber-100/60 rounded-xl border border-amber-300 text-amber-900 text-xs font-semibold flex items-center gap-2">
              <span>ℹ️</span> Data KM Awal & KM Akhir tidak dicantumkan oleh pemohon.
            </div>

            <div v-if="selectedItem.fuel_liters" class="text-xs text-amber-900 font-medium">
              ⛽ Estimasi / Pengisian BBM: <strong>{{ selectedItem.fuel_liters }} Liter</strong>
            </div>

            <!-- Photo Thumbnails for Quick Inspection -->
            <div class="grid grid-cols-3 gap-2 pt-1">
              <div class="flex flex-col items-center">
                <span class="text-[9px] font-bold text-slate-500 mb-1">Foto KM Awal</span>
                <div
                  v-if="selectedItem.start_km_photo_url"
                  class="cursor-pointer group relative w-full h-20 rounded-xl overflow-hidden border border-slate-200 bg-white"
                  @click="openLightbox(selectedItem.start_km_photo_url, 'Foto Odometer Awal')"
                >
                  <img
                    :src="resolvePhotoUrl(selectedItem.start_km_photo_url)"
                    class="w-full h-full object-cover group-hover:scale-105 transition-transform"
                  />
                </div>
                <div v-else class="w-full h-20 rounded-xl bg-slate-100 flex items-center justify-center text-[10px] text-slate-400 italic">
                  Tanpa foto
                </div>
              </div>

              <div class="flex flex-col items-center">
                <span class="text-[9px] font-bold text-slate-500 mb-1">Foto KM Akhir</span>
                <div
                  v-if="selectedItem.end_km_photo_url"
                  class="cursor-pointer group relative w-full h-20 rounded-xl overflow-hidden border border-slate-200 bg-white"
                  @click="openLightbox(selectedItem.end_km_photo_url, 'Foto Odometer Akhir')"
                >
                  <img
                    :src="resolvePhotoUrl(selectedItem.end_km_photo_url)"
                    class="w-full h-full object-cover group-hover:scale-105 transition-transform"
                  />
                </div>
                <div v-else class="w-full h-20 rounded-xl bg-slate-100 flex items-center justify-center text-[10px] text-slate-400 italic">
                  Tanpa foto
                </div>
              </div>

              <div class="flex flex-col items-center">
                <span class="text-[9px] font-bold text-slate-500 mb-1">Struk SPBU</span>
                <div
                  v-if="selectedItem.attachment_url"
                  class="cursor-pointer group relative w-full h-20 rounded-xl overflow-hidden border border-slate-200 bg-white"
                  @click="openLightbox(selectedItem.attachment_url, 'Struk BBM SPBU')"
                >
                  <img
                    :src="resolvePhotoUrl(selectedItem.attachment_url)"
                    class="w-full h-full object-cover group-hover:scale-105 transition-transform"
                  />
                </div>
                <div v-else class="w-full h-20 rounded-xl bg-slate-100 flex items-center justify-center text-[10px] text-slate-400 italic">
                  Tanpa struk
                </div>
              </div>
            </div>
          </div>

          <!-- Banner Installation Rate Adjustment for Cashier -->
          <div
            v-if="selectedItem.category === 'BANNER_INSTALLATION'"
            class="p-5 bg-emerald-50 border border-emerald-200 rounded-3xl space-y-4"
          >
            <div class="flex items-center justify-between">
              <div>
                <h4 class="text-xs font-black text-emerald-950 uppercase tracking-wider">
                  Penyesuaian Tarif Pasang Spanduk
                </h4>
                <p class="text-[11px] text-emerald-800 font-medium mt-0.5">
                  Total: <strong>{{ selectedItem.banner_count || selectedItem.banner_orders?.length || 1 }} Spanduk</strong> terpasang
                </p>
              </div>
              <span class="px-2.5 py-1 bg-emerald-600 text-white text-[10px] font-black rounded-lg">
                KASIR
              </span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-emerald-900 uppercase tracking-wider">
                  Tarif per Spanduk (Rp)
                </label>
                <input
                  type="number"
                  v-model.number="customUnitRate"
                  @input="customAmount = customUnitRate * (selectedItem.banner_count || selectedItem.banner_orders?.length || 1)"
                  class="w-full bg-white border border-emerald-300 rounded-xl px-3 py-2 text-sm font-bold text-slate-900 focus:ring-2 focus:ring-emerald-500/20 focus:outline-none"
                  placeholder="15000"
                />
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-black text-emerald-900 uppercase tracking-wider">
                  Total Nominal Pencairan (Rp)
                </label>
                <input
                  type="number"
                  v-model.number="customAmount"
                  class="w-full bg-white border border-emerald-300 rounded-xl px-3 py-2 text-sm font-black text-emerald-700 focus:ring-2 focus:ring-emerald-500/20 focus:outline-none"
                  placeholder="90000"
                />
              </div>
            </div>
          </div>

          <!-- Standard Amount Display for Non-banner -->
          <div v-else class="p-4 bg-slate-50 rounded-2xl border border-slate-100 flex items-center justify-between">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
              Nominal Klaim
            </span>
            <span class="font-black text-primary text-xl">
              Rp {{ Number(selectedItem.amount).toLocaleString('id-ID') }}
            </span>
          </div>

          <!-- SALDO PEMOTONG & AKUN PENGELUARAN (PILIHAN DETAIL KASIR/APPROVER) -->
          <div class="p-5 bg-blue-50/60 border border-blue-200/80 rounded-3xl space-y-4">
            <div class="flex items-center justify-between">
              <div>
                <h4 class="text-xs font-black text-blue-950 uppercase tracking-wider flex items-center gap-1.5">
                  <span>💳</span> Sumber Saldo & Akun Pengeluaran
                </h4>
                <p class="text-[11px] text-blue-800 font-medium mt-0.5">
                  Tentukan kas pemotong dan klasifikasi akun beban saat disetujui.
                </p>
              </div>
              <span
                class="px-2.5 py-1 text-[10px] font-black rounded-lg uppercase"
                :class="isSelectedItemFromPabrik ? 'bg-amber-600 text-white' : 'bg-blue-600 text-white'"
              >
                {{ isSelectedItemFromPabrik ? 'Unit: Pabrik' : 'Unit: Cabang' }}
              </span>
            </div>

            <!-- 1. Pilihan Sumber Saldo -->
            <div class="space-y-2">
              <label class="text-[10px] font-black text-blue-900 uppercase tracking-wider block">
                Sumber Saldo Pemotong
              </label>

              <!-- Khusus Karyawan Pabrik -->
              <div v-if="isSelectedItemFromPabrik" class="space-y-2">
                <div class="p-3 bg-amber-100/50 border border-amber-300/80 rounded-2xl text-[11px] text-amber-900 font-semibold flex items-center gap-2">
                  <span>ℹ️</span> Petty Cash hanya berlaku untuk cabang Pabrik Utama.
                </div>
                <div class="space-y-1">
                  <span class="text-[10px] font-bold text-slate-500 uppercase">Pilih Akun Kas Pabrik:</span>
                  <select
                    v-model="approvalForm.kas_pabrik_id"
                    class="w-full px-3 py-2 bg-white border border-blue-300 rounded-xl text-xs font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-400"
                  >
                    <option v-for="k in kasPabriks" :key="k.id" :value="k.id">
                      {{ k.kas_name }} (Saldo: {{ formatRupiah(k.balance) }})
                    </option>
                  </select>
                </div>
              </div>

              <!-- Karyawan Cabang Non-Pabrik -->
              <div v-else class="space-y-2.5">
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                  <label
                    class="p-2.5 rounded-xl border flex flex-col items-start cursor-pointer transition-all"
                    :class="approvalForm.fund_source === 'BRANCH_PETTY_CASH' ? 'bg-emerald-50 border-emerald-500 text-emerald-950 font-bold shadow-sm' : 'bg-white border-slate-200 text-slate-600'"
                  >
                    <input type="radio" value="BRANCH_PETTY_CASH" v-model="approvalForm.fund_source" class="hidden" />
                    <span class="text-xs font-bold">Kas Kecil Cabang</span>
                    <span class="text-[10px] text-emerald-700 mt-0.5">Saldo: {{ formatRupiah(branchCashBalances.petty_cash_balance) }}</span>
                  </label>

                  <label
                    class="p-2.5 rounded-xl border flex flex-col items-start cursor-pointer transition-all"
                    :class="approvalForm.fund_source === 'BRANCH_CASH' ? 'bg-amber-50 border-amber-500 text-amber-950 font-bold shadow-sm' : 'bg-white border-slate-200 text-slate-600'"
                  >
                    <input type="radio" value="BRANCH_CASH" v-model="approvalForm.fund_source" class="hidden" />
                    <span class="text-xs font-bold">Kas Tunai Cabang</span>
                    <span class="text-[10px] text-amber-700 mt-0.5">Saldo: {{ formatRupiah(branchCashBalances.manual_cash_balance) }}</span>
                  </label>

                  <label
                    class="p-2.5 rounded-xl border flex flex-col items-start cursor-pointer transition-all"
                    :class="approvalForm.fund_source === 'BANK_TRANSFER' ? 'bg-blue-50 border-blue-500 text-blue-950 font-bold shadow-sm' : 'bg-white border-slate-200 text-slate-600'"
                  >
                    <input type="radio" value="BANK_TRANSFER" v-model="approvalForm.fund_source" class="hidden" />
                    <span class="text-xs font-bold">Transfer Bank</span>
                    <span class="text-[10px] text-blue-700 mt-0.5">{{ companyBanks.length }} Rekening Aktif</span>
                  </label>
                </div>

                <!-- Dropdown Rekening Bank jika Transfer Bank dipilih -->
                <div v-if="approvalForm.fund_source === 'BANK_TRANSFER'" class="p-3 bg-white rounded-xl border border-blue-200 space-y-1">
                  <span class="text-[10px] font-bold text-slate-500 uppercase">Pilih Rekening Bank Perusahaan:</span>
                  <select
                    v-model="approvalForm.bank_account_id"
                    class="w-full px-3 py-2 border border-slate-200 rounded-lg text-xs font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-400 bg-white"
                  >
                    <option v-for="b in companyBanks" :key="b.id" :value="b.id">
                      {{ b.bank_name }} - {{ b.account_number }} (a/n {{ b.account_holder }}) - Saldo: {{ formatRupiah(b.manual_balance) }}
                    </option>
                  </select>
                </div>
              </div>
            </div>

            <!-- 2. Pilihan Akun Pengeluaran (Beban) -->
            <div class="space-y-1.5 pt-1">
              <label class="text-[10px] font-black text-blue-900 uppercase tracking-wider block">
                Akun Pengeluaran / Beban (CoA PSAK)
              </label>
              <select
                v-model="approvalForm.expense_account_id"
                class="w-full px-3 py-2.5 bg-white border border-blue-300 rounded-xl text-xs font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-400"
              >
                <option value="">-- Tanpa Akun Khusus (Default 5400 Beban Reimbursement) --</option>
                <option v-for="acc in expenseAccounts" :key="acc.id" :value="acc.id">
                  {{ acc.code }} - {{ acc.name }}
                </option>
              </select>
              <p class="text-[10px] text-blue-700/80 font-medium">
                Pilih pos beban agar pembukuan otomatis terintegrasi dan akurat sesuai jenis pengeluaran.
              </p>
            </div>

            <!-- 3. Catatan Pencairan / Approval -->
            <div class="space-y-1 pt-1">
              <label class="text-[10px] font-black text-blue-900 uppercase tracking-wider block">
                Catatan Kasir / Pembukuan (Opsional)
              </label>
              <input
                type="text"
                v-model="approvalForm.notes"
                placeholder="Contoh: Diserahkan tunai ke karyawan..."
                class="w-full px-3 py-2 bg-white border border-blue-200 rounded-xl text-xs text-slate-800 focus:outline-none focus:ring-2 focus:ring-blue-400"
              />
            </div>
          </div>

          <!-- Rejection Reason Input -->
          <div class="space-y-2">
            <p class="text-[11px] font-black text-slate-400 uppercase tracking-wider">
              Alasan Penolakan (Wajib jika ditolak)
            </p>
            <Textarea
              v-model="rejectReason"
              placeholder="Tuliskan alasan penolakan pengajuan jika ditolak..."
              class="min-h-[80px] rounded-2xl border-slate-200 text-xs"
            />
          </div>
        </div>

        <DialogFooter class="p-6 bg-slate-50 flex flex-col sm:flex-row gap-3">
          <Button
            variant="ghost"
            @click="isModalOpen = false"
            class="rounded-2xl font-bold text-slate-400 text-xs h-12 flex-1 sm:flex-none"
          >
            BATAL
          </Button>

          <div class="flex gap-3 flex-1">
            <Button
              variant="outline"
              @click="updateStatus('REJECTED')"
              :disabled="isSubmitting"
              class="rounded-2xl font-bold border-rose-200 text-rose-600 hover:bg-rose-50 h-12 flex-1"
            >
              <XCircle class="w-4 h-4" />
              TOLAK
            </Button>

            <Button
              @click="updateStatus('APPROVED')"
              :disabled="isSubmitting"
              class="rounded-2xl font-black bg-emerald-600 hover:bg-emerald-700 text-white h-12 flex-1 shadow-lg shadow-emerald-200"
            >
              <CheckCircle2 class="w-4 h-4" />
              SETUJUI
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- CRUD MODAL (Add / Edit Reimbursement) -->
    <Dialog v-model:open="isCRUDModalOpen">
      <DialogContent
        class="sm:max-w-2xl rounded-[2rem] md:rounded-[3rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <DialogHeader class="p-6 md:p-8 bg-primary text-white relative shrink-0">
          <DialogTitle class="text-xl font-black tracking-tight">
            {{ isEditMode ? 'Edit Pengajuan Reimbursement' : 'Tambah Pengajuan Reimbursement' }}
          </DialogTitle>
          <DialogDescription class="text-white/80 text-xs mt-1">
            Pilih kategori klaim dan lengkapi rincian pengeluaran berikut.
          </DialogDescription>
        </DialogHeader>

        <div class="p-6 md:p-8 space-y-5 overflow-y-auto custom-scrollbar flex-1 bg-white">
          <!-- Kategori Selector (Only in Create Mode) -->
          <div v-if="!isEditMode" class="space-y-2">
            <Label class="text-[10px] font-black text-slate-400 uppercase tracking-wider pl-1">
              Kategori Pengajuan
            </Label>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <button
                type="button"
                @click="createCategory = 'GENERAL'"
                :class="[
                  'p-3 rounded-2xl border text-left transition-all flex flex-col gap-1.5',
                  createCategory === 'GENERAL'
                    ? 'border-primary bg-primary/5 text-primary shadow-sm font-black'
                    : 'border-slate-200 hover:border-slate-300 text-slate-600 font-bold',
                ]"
              >
                <div class="flex items-center gap-2">
                  <FileText class="w-4 h-4 shrink-0" />
                  <span class="text-xs">Operasional Umum</span>
                </div>
                <div class="text-[10px] font-medium text-slate-400">ATK, makan, logistik</div>
              </button>

              <button
                type="button"
                @click="createCategory = 'BBM'"
                :class="[
                  'p-3 rounded-2xl border text-left transition-all flex flex-col gap-1.5',
                  createCategory === 'BBM'
                    ? 'border-amber-600 bg-amber-50 text-amber-900 shadow-sm font-black'
                    : 'border-slate-200 hover:border-slate-300 text-slate-600 font-bold',
                ]"
              >
                <div class="flex items-center gap-2">
                  <span class="text-sm">⛽</span>
                  <span class="text-xs">Klaim BBM</span>
                </div>
                <div class="text-[10px] font-medium text-slate-400">Bensin, Solar, KM</div>
              </button>

              <button
                type="button"
                @click="createCategory = 'BANNER_INSTALLATION'"
                :class="[
                  'p-3 rounded-2xl border text-left transition-all flex flex-col gap-1.5',
                  createCategory === 'BANNER_INSTALLATION'
                    ? 'border-emerald-600 bg-emerald-50 text-emerald-800 shadow-sm font-black'
                    : 'border-slate-200 hover:border-slate-300 text-slate-600 font-bold',
                ]"
              >
                <div class="flex items-center gap-2">
                  <Layers class="w-4 h-4 shrink-0 text-emerald-600" />
                  <span class="text-xs">Pasang Spanduk</span>
                </div>
                <div class="text-[10px] font-medium text-slate-400">Spanduk lapangan</div>
              </button>
            </div>
          </div>

          <!-- Karyawan Select -->
          <div class="space-y-1.5">
            <Label class="text-[10px] font-black text-slate-400 uppercase tracking-wider pl-1">
              Karyawan Pemohon
            </Label>
            <Select v-model="form.employee_id" :disabled="isEditMode">
              <SelectTrigger class="h-11 rounded-2xl border-slate-200 text-xs font-bold">
                <SelectValue placeholder="Pilih Karyawan..." />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="e in employees" :key="e.id" :value="e.id">
                  {{ e.first_name }} {{ e.last_name || '' }} ({{ e.branch?.name || '-' }})
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- BBM SPECIFIC FIELDS: KM, Plate, Photos -->
          <div
            v-if="createCategory === 'BBM'"
            class="space-y-4 p-4 bg-amber-50/70 border border-amber-200 rounded-3xl"
          >
            <div class="flex items-center justify-between">
              <h4 class="text-xs font-black text-amber-950 uppercase tracking-wider flex items-center gap-1.5">
                <span>⛽</span> Data Odometer & Perjalanan BBM
              </h4>
              <span class="text-[10px] text-amber-700 font-bold bg-amber-100 px-2 py-0.5 rounded-md">
                Validasi Kasir
              </span>
            </div>

            <!-- Plat Nomor Kendaraan & Liter BBM -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div class="space-y-1">
                <label class="text-[10px] font-black text-slate-600 uppercase tracking-wider">
                  Plat Nomor Kendaraan
                </label>
                <Input
                  v-model="form.vehicle_plate_number"
                  placeholder="Contoh: B 1234 XYZ / AD 5678 CD"
                  class="h-10 rounded-xl border-amber-300 bg-white text-xs font-bold"
                />
              </div>

              <div class="space-y-1">
                <label class="text-[10px] font-black text-slate-600 uppercase tracking-wider">
                  Volume BBM (Liter) - Opsional
                </label>
                <Input
                  type="number"
                  step="0.1"
                  :model-value="form.fuel_liters ?? undefined"
                  @update:model-value="form.fuel_liters = $event === '' || $event == null ? null : Number($event)"
                  placeholder="Contoh: 15.5"
                  class="h-10 rounded-xl border-amber-300 bg-white text-xs font-bold"
                />
              </div>
            </div>

            <!-- KM Berangkat, KM Pulang, and Total KM Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div class="space-y-1">
                <label class="text-[10px] font-black text-slate-600 uppercase tracking-wider">
                  KM Berangkat (Awal)
                </label>
                <Input
                  type="number"
                  :model-value="form.start_km ?? undefined"
                  @update:model-value="form.start_km = $event === '' || $event == null ? null : Number($event)"
                  placeholder="Contoh: 12500"
                  class="h-10 rounded-xl border-amber-300 bg-white text-xs font-bold"
                />
              </div>

              <div class="space-y-1">
                <label class="text-[10px] font-black text-slate-600 uppercase tracking-wider">
                  KM Pulang (Akhir)
                </label>
                <Input
                  type="number"
                  :model-value="form.end_km ?? undefined"
                  @update:model-value="form.end_km = $event === '' || $event == null ? null : Number($event)"
                  placeholder="Contoh: 12620"
                  class="h-10 rounded-xl border-amber-300 bg-white text-xs font-bold"
                />
              </div>

              <div class="space-y-1">
                <label class="text-[10px] font-black text-amber-900 uppercase tracking-wider">
                  Total Jarak (KM)
                </label>
                <div class="h-10 px-3 rounded-xl bg-amber-100 border border-amber-300 flex items-center font-black text-amber-950 text-sm">
                  {{ form.total_km != null ? form.total_km : 0 }} KM
                </div>
              </div>
            </div>

            <!-- Upload Photos: Odometer Start, Odometer End, Nota SPBU -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 border-t border-amber-200">
              <div class="space-y-1">
                <label class="text-[10px] font-black text-slate-600 uppercase tracking-wider block truncate">
                  Foto KM Berangkat
                </label>
                <Input
                  type="file"
                  @change="onStartKMFileChange"
                  accept="image/*"
                  class="h-10 rounded-xl border-amber-300 bg-white text-[11px] pt-1.5"
                />
                <span v-if="form.start_km_photo_url" class="text-[10px] text-emerald-600 font-bold block truncate">
                  ✓ Foto sudah tersimpan
                </span>
              </div>

              <div class="space-y-1">
                <label class="text-[10px] font-black text-slate-600 uppercase tracking-wider block truncate">
                  Foto KM Pulang
                </label>
                <Input
                  type="file"
                  @change="onEndKMFileChange"
                  accept="image/*"
                  class="h-10 rounded-xl border-amber-300 bg-white text-[11px] pt-1.5"
                />
                <span v-if="form.end_km_photo_url" class="text-[10px] text-emerald-600 font-bold block truncate">
                  ✓ Foto sudah tersimpan
                </span>
              </div>

              <div class="space-y-1">
                <label class="text-[10px] font-black text-slate-600 uppercase tracking-wider block truncate">
                  Struk / Nota SPBU
                </label>
                <Input
                  type="file"
                  @change="onFileChange"
                  accept="image/*,.pdf"
                  class="h-10 rounded-xl border-amber-300 bg-white text-[11px] pt-1.5"
                />
                <span v-if="form.attachment_url" class="text-[10px] text-emerald-600 font-bold block truncate">
                  ✓ Nota sudah tersimpan
                </span>
              </div>
            </div>
          </div>

          <!-- BANNER INSTALLATION SPECIFIC: Checklist Spanduk Terpasang -->
          <div
            v-if="createCategory === 'BANNER_INSTALLATION' && !isEditMode"
            class="space-y-3 p-4 bg-emerald-50/70 border border-emerald-200 rounded-3xl"
          >
            <div class="flex items-center justify-between">
              <div>
                <h4 class="text-xs font-black text-emerald-950 uppercase tracking-wider">
                  Pilih Spanduk Yang Diajukan
                </h4>
                <p class="text-[11px] text-emerald-800 font-medium">
                  {{ selectedBannerIds.length }} dari {{ unclaimedForEmployee.length }} spanduk dipilih
                </p>
              </div>

              <button
                type="button"
                v-if="unclaimedForEmployee.length > 0"
                @click="toggleSelectAllBanners"
                class="text-[10px] font-black text-emerald-700 bg-white border border-emerald-300 px-2.5 py-1 rounded-lg hover:bg-emerald-100/50 transition-colors"
              >
                {{ selectedBannerIds.length === unclaimedForEmployee.length ? 'Batal Pilih Semua' : 'Pilih Semua' }}
              </button>
            </div>

            <!-- Loading indicator -->
            <div v-if="isLoadingEmpBanners" class="p-6 text-center text-emerald-800 text-xs font-bold">
              Memuat daftar spanduk terpasang karyawan...
            </div>

            <!-- Empty state -->
            <div
              v-else-if="unclaimedForEmployee.length === 0"
              class="p-6 bg-white/80 rounded-2xl border border-emerald-200 text-center text-xs text-slate-500 font-medium"
            >
              {{ form.employee_id ? 'Tidak ada spanduk terpasang baru untuk karyawan ini.' : 'Silakan pilih karyawan terlebih dahulu.' }}
            </div>

            <!-- Banners Checklist -->
            <div v-else class="space-y-2 max-h-56 overflow-y-auto custom-scrollbar pr-1">
              <div
                v-for="b in unclaimedForEmployee"
                :key="b.id"
                @click="toggleBannerSelection(b.id)"
                :class="[
                  'p-3 rounded-2xl border transition-all flex items-center gap-3 cursor-pointer select-none',
                  selectedBannerIds.includes(b.id)
                    ? 'bg-white border-emerald-500 shadow-sm'
                    : 'bg-white/60 border-slate-200 opacity-70',
                ]"
              >
                <div
                  :class="[
                    'w-5 h-5 rounded-lg border flex items-center justify-center shrink-0 transition-colors',
                    selectedBannerIds.includes(b.id)
                      ? 'bg-emerald-600 border-emerald-600 text-white'
                      : 'border-slate-300 bg-white',
                  ]"
                >
                  <Check v-if="selectedBannerIds.includes(b.id)" class="w-3.5 h-3.5" />
                </div>

                <img
                  v-if="b.photo_installed"
                  :src="b.photo_installed.startsWith('http') ? b.photo_installed : `${API_BASE}${b.photo_installed}`"
                  alt="Spanduk"
                  class="w-10 h-10 object-cover rounded-xl border border-slate-200 shrink-0"
                />

                <div class="flex-1 min-w-0">
                  <p class="font-bold text-xs text-slate-900 truncate">
                    {{ b.store_name || b.shop_name || 'Toko' }}
                  </p>
                  <p class="text-[10px] text-slate-500 truncate">
                    {{ b.banner_type || 'Standard' }} • {{ b.size_length || 3 }}x{{ b.size_width || 1 }}m
                  </p>
                </div>

                <span class="text-[10px] text-slate-400 font-medium shrink-0">
                  {{ b.installed_at ? new Date(b.installed_at).toLocaleDateString('id-ID') : '-' }}
                </span>
              </div>
            </div>

            <!-- Unit Rate Settings -->
            <div class="grid grid-cols-2 gap-3 pt-2 border-t border-emerald-200">
              <div class="space-y-1">
                <label class="text-[10px] font-black text-emerald-900 uppercase tracking-wider">
                  Tarif per Spanduk (Rp)
                </label>
                <Input
                  type="number"
                  v-model.number="bannerUnitRateInput"
                  @input="recalculateBannerClaimAmount"
                  class="h-10 rounded-xl border-emerald-300 bg-white text-xs font-bold"
                />
              </div>

              <div class="space-y-1">
                <label class="text-[10px] font-black text-emerald-900 uppercase tracking-wider">
                  Total Nominal Klaim (Rp)
                </label>
                <div class="h-10 px-3 rounded-xl bg-white border border-emerald-300 flex items-center font-black text-emerald-700 text-sm">
                  Rp {{ form.amount.toLocaleString('id-ID') }}
                </div>
              </div>
            </div>
          </div>

          <!-- Judul Klaim -->
          <div class="space-y-1.5">
            <Label class="text-[10px] font-black text-slate-400 uppercase tracking-wider pl-1">
              Judul Pengajuan
            </Label>
            <Input
              v-model="form.title"
              placeholder="Contoh: Reimbursement BBM Luar Kota / Pasang Spanduk"
              class="h-11 rounded-2xl border-slate-200 text-xs font-medium"
            />
          </div>

          <!-- Nominal Klaim (Standard input for General & BBM) -->
          <div v-if="createCategory !== 'BANNER_INSTALLATION'" class="space-y-1.5">
            <Label class="text-[10px] font-black text-slate-400 uppercase tracking-wider pl-1">
              {{ createCategory === 'BBM' ? 'Total Nominal Pembelian BBM (Rp)' : 'Nominal Pengajuan (Rp)' }}
            </Label>
            <Input
              type="number"
              v-model="form.amount"
              placeholder="0"
              class="h-11 rounded-2xl border-slate-200 text-xs font-bold text-slate-900"
            />
          </div>

          <!-- Deskripsi -->
          <div class="space-y-1.5">
            <Label class="text-[10px] font-black text-slate-400 uppercase tracking-wider pl-1">
              Deskripsi / Catatan Tambahan
            </Label>
            <Textarea
              v-model="form.description"
              placeholder="Tuliskan keterangan detail pengeluaran..."
              class="min-h-[70px] rounded-2xl border-slate-200 text-xs"
            />
          </div>

          <!-- Upload Nota (For General only, as BBM has its own upload grid) -->
          <div v-if="createCategory === 'GENERAL'" class="space-y-1.5">
            <Label class="text-[10px] font-black text-slate-400 uppercase tracking-wider pl-1">
              Unggah Nota / Bukti Bayar
            </Label>
            <Input
              type="file"
              @change="onFileChange"
              accept="image/*,.pdf"
              class="h-11 rounded-2xl border-slate-200 text-xs pt-2"
            />
          </div>
        </div>

        <DialogFooter class="p-6 bg-slate-50 border-t border-slate-100 flex gap-3">
          <Button
            variant="ghost"
            @click="isCRUDModalOpen = false"
            class="rounded-2xl font-bold text-slate-400 text-xs h-12 flex-1 sm:flex-none"
          >
            BATAL
          </Button>

          <Button
            @click="saveReimbursement"
            :disabled="isSubmitting"
            class="bg-primary hover:bg-primary/90 text-white rounded-2xl h-12 px-8 font-black text-xs shadow-lg shadow-primary/20 flex-1"
          >
            {{ isSubmitting ? 'MENYIMPAN...' : isEditMode ? 'SIMPAN PERUBAHAN' : 'TAMBAH PENGAJUAN' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- LIGHTBOX PREVIEW MODAL -->
    <Dialog :open="!!previewImageUrl" @update:open="previewImageUrl = null">
      <DialogContent class="sm:max-w-3xl p-0 overflow-hidden bg-slate-950 border-none rounded-3xl shadow-2xl">
        <div class="relative p-4 flex flex-col items-center justify-center min-h-[300px]">
          <button
            @click="previewImageUrl = null"
            class="absolute top-4 right-4 z-10 w-9 h-9 bg-white/20 hover:bg-white/30 text-white rounded-full flex items-center justify-center backdrop-blur-md transition-all"
          >
            <X class="w-5 h-5" />
          </button>
          <img
            v-if="previewImageUrl"
            :src="previewImageUrl"
            :alt="previewImageTitle"
            class="max-h-[80vh] w-auto object-contain rounded-xl shadow-lg"
          />
          <p class="text-xs text-white/80 font-bold mt-3">{{ previewImageTitle }}</p>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
