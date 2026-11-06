<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  Wallet,
  History,
  Search,
  Plus,
  Truck,
  Users,
  Calendar,
  Save,
  CheckCircle,
  Pencil,
  Trash2,
  Package,
  Eye,
  X,
  Coins,
  ArrowDownRight,
  AlertCircle,
  CheckSquare,
  Square,
  CreditCard,
  Building2,
  Clock
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

interface KasPabrik {
  id: string
  branch_id: string
  kas_code: string
  kas_name: string
  description?: string
  bank_name?: string
  account_number?: string
  balance: number
  is_active: boolean
}

const masterDataStore = useMasterDataStore()
const salaryHistory = ref<any[]>([])
const dailyWorkers = ref<any[]>([])
const kioskEmployees = ref<any[]>([])
const branches = ref<any[]>([])
const kasList = ref<KasPabrik[]>([])

const showPreviewModal = ref(false)
const previewImageUrl = ref('')
const previewModalTitle = ref('')
const previewModalSubtitle = ref('')

const selectedBranchName = computed(() => {
  return masterDataStore.selectedBranchName
})

const getFullImageUrl = (path?: string | null): string => {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:')) return path
  const apiUrl = import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || window.location.origin
  try {
    const origin = new URL(apiUrl).origin
    return origin + (path.startsWith('/') ? path : `/${path}`)
  } catch {
    return path
  }
}

const openPreviewModal = (url: string, title = 'Preview', subtitle = '') => {
  previewImageUrl.value = url
  previewModalTitle.value = title
  previewModalSubtitle.value = subtitle
  showPreviewModal.value = true
}

const closePreviewModal = () => {
  showPreviewModal.value = false
  previewImageUrl.value = ''
}

const showWorkerModal = ref(false)
const savingWorker = ref(false)
const newWorkerName = ref('')
const newWorkerBranchId = ref('')

const activeTab = ref<'UNPAID' | 'PAID'>('UNPAID')
const searchQuery = ref('')
const isLoading = ref(false)
const showModal = ref(false)
const editMode = ref(false)
const isSubmitting = ref(false)

// Pay Modal State
const showPayModal = ref(false)
const selectedPayItem = ref<any>(null)
const isPaying = ref(false)
const payForm = ref({
  kas_pabrik_id: '',
  total_paid: 0,
  payment_date: new Date().toISOString().split('T')[0],
  expense_category: 'DAILY_WORKER',
  notes: ''
})

// Batch Pay State
const selectedSalaryIds = ref<string[]>([])
const showBatchPayModal = ref(false)
const isBatchPaying = ref(false)
const batchPayForm = ref({
  kas_pabrik_id: '',
  payment_date: new Date().toISOString().split('T')[0],
  expense_category: 'DAILY_WORKER',
  notes: ''
})

const proofFile = ref<File | null>(null)
const proofFileInput = ref<HTMLInputElement | null>(null)

const onFileChangeProof = (e: any) => {
  proofFile.value = e.target.files[0] || null
}

const triggerProofFilePicker = () => proofFileInput.value?.click()

const expenseCategories = [
  { value: 'DAILY_WORKER', label: 'Gaji Harian Mingguan' },
  { value: 'KULI', label: 'Gaji Kuli Muat / Bongkar' },
  { value: 'BIAYA_BORONGAN_MUAT_KECAP', label: 'Biaya Borongan Muat Kecap' },
  { value: 'BIAYA_BORONGAN_MUAT_SAOS', label: 'Biaya Borongan Muat Saos' },
  { value: 'BIAYA_BORONGAN_MUAT_BIHUN', label: 'Biaya Borongan Muat Bihun' },
  { value: 'BIAYA_BORONGAN_MUAT_CUKA', label: 'Biaya Borongan Muat Cuka' },
  { value: 'BIAYA_BORONGAN_TRONTON', label: 'Biaya Borongan dan Muat Tronton' },
  { value: 'GAJI_BORONGAN_PACKING_KECAP', label: 'Biaya Gaji Borongan Packing Kecap' },
  { value: 'GAJI_BORONGAN_PACKING_SAOS', label: 'Biaya Gaji Borongan Packing Saos' },
  { value: 'GAJI_BORONGAN_OVEN', label: 'Biaya Gaji Borongan Oven' },
  { value: 'GAJI_PRODUKSI_KECAP', label: 'Biaya Gaji Produksi Kecap' },
  { value: 'GAJI_PRODUKSI_SAOS', label: 'Biaya Gaji Produksi Saos' },
  { value: 'PEMBERSIHAN_PRODUKSI', label: 'Biaya Pembersihan Tempat Produksi' },
  { value: 'GAJI_TUKANG_PABRIK', label: 'Biaya Gaji Tukang Pabrik' },
  { value: 'BIAYA_OPERASIONAL_CABANG', label: 'Talangan Biaya Operasional Cabang' },
  { value: 'BEBAN_LAIN_PRODUKSI', label: 'Beban Lain-Lain Produksi' },
  { value: 'BEBAN_LAIN', label: 'Beban Lain-Lain' }
]

const form = ref({
  id: '',
  daily_worker_id: '',
  type: 'OVEN BOTOL',
  date: '',
  total_paid: 0,
  notes: '',
  auto_pay: false,
  kas_pabrik_id: '',
  expense_category: 'GAJI_BORONGAN_OVEN',
  payment_date: new Date().toISOString().split('T')[0]
})

const filteredHistory = computed(() => {
  let filtered = salaryHistory.value.filter(h => h.status === activeTab.value)
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(h => 
      h.daily_worker?.name?.toLowerCase().includes(q) || 
      h.notes?.toLowerCase().includes(q) ||
      h.kas_pabrik?.kas_name?.toLowerCase().includes(q) ||
      h.kas_pabrik?.kas_code?.toLowerCase().includes(q) ||
      h.branch?.name?.toLowerCase().includes(q)
    )
  }
  return filtered
})

const activeKasList = computed(() => {
  return kasList.value.filter(k => k.is_active !== false)
})

const availableKioskEmployees = computed(() => {
  if (masterDataStore.isAllBranchesSelected) {
    return kioskEmployees.value
  }
  return kioskEmployees.value.filter((e: any) => !e.branch_id || masterDataStore.isBranchSelected(e.branch_id))
})

const selectedKasForPay = computed(() => {
  return kasList.value.find(k => k.id === payForm.value.kas_pabrik_id)
})

const selectedKasForBatchPay = computed(() => {
  return kasList.value.find(k => k.id === batchPayForm.value.kas_pabrik_id)
})

const selectedKasForCreate = computed(() => {
  return kasList.value.find(k => k.id === form.value.kas_pabrik_id)
})

const batchTotalAmount = computed(() => {
  const selectedItems = salaryHistory.value.filter(h => selectedSalaryIds.value.includes(h.id))
  return selectedItems.reduce((sum, item) => sum + (Number(item.total_paid) || 0), 0)
})

const isAllSelected = computed(() => {
  const unpaidItems = filteredHistory.value
  if (unpaidItems.length === 0) return false
  return unpaidItems.every(item => selectedSalaryIds.value.includes(item.id))
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedSalaryIds.value = []
  } else {
    selectedSalaryIds.value = filteredHistory.value.map(item => item.id)
  }
}

const toggleSelectItem = (id: string) => {
  const idx = selectedSalaryIds.value.indexOf(id)
  if (idx > -1) {
    selectedSalaryIds.value.splice(idx, 1)
  } else {
    selectedSalaryIds.value.push(id)
  }
}

const formatRupiah = (value: number) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value || 0)
}

const getCategoryLabel = (catValue?: string) => {
  if (!catValue) return '-'
  const found = expenseCategories.find(c => c.value === catValue)
  return found ? found.label : catValue
}

const mapTypeToExpenseCategory = (workerType: string): string => {
  const upper = (workerType || '').toUpperCase()
  if (upper.includes('KULI')) return 'KULI'
  if (upper.includes('PACKING KECAP')) return 'GAJI_BORONGAN_PACKING_KECAP'
  if (upper.includes('PACKING SAOS')) return 'GAJI_BORONGAN_PACKING_SAOS'
  if (upper.includes('OVEN')) return 'GAJI_BORONGAN_OVEN'
  if (upper.includes('PRODUKSI KECAP')) return 'GAJI_PRODUKSI_KECAP'
  if (upper.includes('PRODUKSI SAOS')) return 'GAJI_PRODUKSI_SAOS'
  if (upper.includes('CLEANING')) return 'PEMBERSIHAN_PRODUKSI'
  if (upper.includes('TUKANG') || upper.includes('TEKNISI')) return 'GAJI_TUKANG_PABRIK'
  return 'DAILY_WORKER'
}

const fetchData = async () => {
  try {
    isLoading.value = true
    const branchId = masterDataStore.selectedBranchId
    const hasBranch = branchId && branchId !== 'ALL_BRANCHES'
    const branchParam = hasBranch ? `?branch_id=${branchId}` : ''
    const empBranchParam = hasBranch ? `&branch_id=${branchId}` : ''

    const [workersRes, salariesRes, empRes, branchesRes, kasRes] = await Promise.all([
      apiClient.get(`/daily-workers${branchParam}`),
      apiClient.get(`/daily-worker-salaries${branchParam}`),
      apiClient.get(`/employees?employment_type=harian&limit=1000${empBranchParam}`),
      apiClient.get('/branches'),
      apiClient.get(`/kas-pabrik${branchParam}`).catch(() => ({ data: { data: [] } }))
    ])
    dailyWorkers.value = workersRes.data?.data || workersRes.data || []
    salaryHistory.value = salariesRes.data?.data || salariesRes.data || []
    kioskEmployees.value = empRes.data?.data || []
    branches.value = branchesRes.data?.data || []
    kasList.value = kasRes.data?.data || []

    // Set default kas if available
    if (kasList.value.length > 0 && kasList.value[0]) {
      if (!payForm.value.kas_pabrik_id || !kasList.value.find(k => k.id === payForm.value.kas_pabrik_id)) {
        payForm.value.kas_pabrik_id = kasList.value[0].id
      }
      if (!batchPayForm.value.kas_pabrik_id || !kasList.value.find(k => k.id === batchPayForm.value.kas_pabrik_id)) {
        batchPayForm.value.kas_pabrik_id = kasList.value[0].id
      }
      if (!form.value.kas_pabrik_id || !kasList.value.find(k => k.id === form.value.kas_pabrik_id)) {
        form.value.kas_pabrik_id = kasList.value[0].id
      }
    } else {
      payForm.value.kas_pabrik_id = ''
      batchPayForm.value.kas_pabrik_id = ''
      form.value.kas_pabrik_id = ''
    }
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    isLoading.value = false
  }
}

watch(() => masterDataStore.selectedBranchId, () => {
  selectedSalaryIds.value = []
  fetchData()
})

onMounted(() => {
  fetchData()
})

const openAddModal = () => {
  editMode.value = false
  form.value = {
    id: '',
    daily_worker_id: '',
    type: 'OVEN BOTOL',
    date: '',
    total_paid: 0,
    notes: '',
    auto_pay: false,
    kas_pabrik_id: activeKasList.value.length > 0 && activeKasList.value[0] ? activeKasList.value[0].id : '',
    expense_category: 'GAJI_BORONGAN_OVEN',
    payment_date: new Date().toISOString().split('T')[0]
  }
  proofFile.value = null
  showModal.value = true
}

const onFormTypeChange = () => {
  form.value.expense_category = mapTypeToExpenseCategory(form.value.type)
}

const openEditModal = (item: any) => {
  editMode.value = true
  form.value = {
    id: item.id,
    daily_worker_id: item.daily_worker_id,
    type: item.daily_worker?.type || 'OVEN BOTOL',
    date: item.date,
    total_paid: item.total_paid,
    notes: item.notes,
    auto_pay: false,
    kas_pabrik_id: item.kas_pabrik_id || (activeKasList.value.length > 0 && activeKasList.value[0] ? activeKasList.value[0].id : ''),
    expense_category: mapTypeToExpenseCategory(item.daily_worker?.type || ''),
    payment_date: new Date().toISOString().split('T')[0]
  }
  showModal.value = true
}

const submitForm = async () => {
  if (!form.value.daily_worker_id || form.value.total_paid <= 0) {
    toast.error('Mohon lengkapi pekerja dan nominal tagihan dengan benar.')
    return
  }

  if (form.value.auto_pay && !form.value.kas_pabrik_id) {
    toast.error('Mohon pilih Kas untuk pembayaran langsung.')
    return
  }
  
  isSubmitting.value = true
  try {
    let finalDailyWorkerId = form.value.daily_worker_id
    let empBranchId = ''
    
    if (!editMode.value) {
      // Find selected kiosk employee
      const emp = kioskEmployees.value.find(e => e.id === form.value.daily_worker_id)
      if (!emp) {
        toast.error('Pekerja tidak ditemukan.')
        return
      }
      empBranchId = emp.branch_id || ''
      const fullName = emp.last_name ? `${emp.first_name} ${emp.last_name}`.trim() : emp.first_name.trim()

      const branchId = masterDataStore.selectedBranchId
      const targetBranchId = (branchId && branchId !== 'ALL_BRANCHES') ? branchId : empBranchId

      // Check if this worker + type combo exists in dailyWorkers
      const existingDw = dailyWorkers.value.find(dw => dw.name === fullName && dw.type === form.value.type)
      if (existingDw) {
        finalDailyWorkerId = existingDw.id
      } else {
        // Create new in /daily-workers
        const res = await apiClient.post('/daily-workers', {
          name: fullName,
          type: form.value.type,
          branch_id: targetBranchId || undefined
        })
        finalDailyWorkerId = res.data?.data?.id || res.data?.id
      }
    }

    if (editMode.value) {
      await apiClient.put(`/daily-worker-salaries/${form.value.id}`, {
        total_paid: form.value.total_paid,
        notes: form.value.notes
      })
      toast.success('Data gaji berhasil diperbarui!')
    } else {
      const branchId = masterDataStore.selectedBranchId
      const targetBranchId = (branchId && branchId !== 'ALL_BRANCHES') ? branchId : empBranchId

      const formData = new FormData()
      formData.append('daily_worker_id', String(finalDailyWorkerId))
      if (targetBranchId) {
        formData.append('branch_id', String(targetBranchId))
      }
      formData.append('total_paid', String(form.value.total_paid))
      formData.append('notes', form.value.notes || '')
      if (form.value.auto_pay) {
        formData.append('auto_pay', 'true')
        formData.append('kas_pabrik_id', form.value.kas_pabrik_id || '')
        formData.append('expense_category', form.value.expense_category || '')
        formData.append('payment_date', form.value.payment_date || '')
      }
      if (proofFile.value) formData.append('proof_image', proofFile.value)

      await apiClient.post('/daily-worker-salaries', formData)
      toast.success(form.value.auto_pay ? 'Tagihan gaji berhasil dibuat dan dibayar dari Kas!' : 'Data tagihan gaji berhasil ditambahkan!')
    }
    showModal.value = false
    fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || error.message || 'Gagal menyimpan data gaji')
  } finally {
    isSubmitting.value = false
  }
}

// Single Pay Modal Handlers
const openPayModal = (item: any) => {
  selectedPayItem.value = item
  payForm.value = {
    kas_pabrik_id: activeKasList.value.length > 0 && activeKasList.value[0] ? activeKasList.value[0].id : '',
    total_paid: item.total_paid || 0,
    payment_date: new Date().toISOString().split('T')[0],
    expense_category: mapTypeToExpenseCategory(item.daily_worker?.type || ''),
    notes: item.notes || `Gaji ${item.daily_worker?.name || 'Pekerja Harian'} (${item.daily_worker?.type || '-'})`
  }
  showPayModal.value = true
}

const confirmPaySalary = async () => {
  if (!selectedPayItem.value) return
  if (!payForm.value.kas_pabrik_id) {
    toast.error('Pilih sumber KAS terlebih dahulu.')
    return
  }

  isPaying.value = true
  try {
    await apiClient.put(`/daily-worker-salaries/${selectedPayItem.value.id}/pay`, {
      kas_pabrik_id: payForm.value.kas_pabrik_id,
      total_paid: Number(payForm.value.total_paid),
      payment_date: payForm.value.payment_date,
      expense_category: payForm.value.expense_category,
      notes: payForm.value.notes
    })
    toast.success('Gaji berhasil dibayar & transaksi Kas berhasil dicatat!')
    showPayModal.value = false
    // Remove from selected batch if was checked
    const idx = selectedSalaryIds.value.indexOf(selectedPayItem.value.id)
    if (idx > -1) selectedSalaryIds.value.splice(idx, 1)
    fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal membayar gaji')
  } finally {
    isPaying.value = false
  }
}

// Batch Pay Handlers
const openBatchPayModal = () => {
  if (selectedSalaryIds.value.length === 0) {
    toast.error('Pilih minimal satu tagihan gaji untuk dibayar.')
    return
  }
  batchPayForm.value = {
    kas_pabrik_id: activeKasList.value.length > 0 && activeKasList.value[0] ? activeKasList.value[0].id : '',
    payment_date: new Date().toISOString().split('T')[0],
    expense_category: 'DAILY_WORKER',
    notes: `Pembayaran ${selectedSalaryIds.value.length} Tagihan Gaji Pekerja Harian`
  }
  showBatchPayModal.value = true
}

const confirmBatchPay = async () => {
  if (selectedSalaryIds.value.length === 0) return
  if (!batchPayForm.value.kas_pabrik_id) {
    toast.error('Pilih sumber KAS terlebih dahulu.')
    return
  }

  isBatchPaying.value = true
  try {
    await apiClient.post('/daily-worker-salaries/batch-pay', {
      salary_ids: selectedSalaryIds.value,
      kas_pabrik_id: batchPayForm.value.kas_pabrik_id,
      payment_date: batchPayForm.value.payment_date,
      expense_category: batchPayForm.value.expense_category,
      notes: batchPayForm.value.notes
    })
    toast.success(`${selectedSalaryIds.value.length} tagihan gaji berhasil dibayar dari Kas!`)
    showBatchPayModal.value = false
    selectedSalaryIds.value = []
    fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal memproses pembayaran massal')
  } finally {
    isBatchPaying.value = false
  }
}

const deleteSalary = async (item: any) => {
  const isPaid = item.status === 'PAID'
  const msg = isPaid
    ? 'Tagihan ini sudah berstatus LUNAS. Menghapus data ini akan OTOMATIS mengembalikan saldo Kas dan menghapus transaksi Kas terkait. Lanjutkan?'
    : 'Apakah Anda yakin ingin menghapus tagihan gaji ini?'
  
  if (!confirm(msg)) return
  
  try {
    await apiClient.delete(`/daily-worker-salaries/${item.id}`)
    toast.success(isPaid ? 'Data gaji berhasil dihapus & saldo Kas telah dikembalikan.' : 'Data gaji berhasil dihapus!')
    const idx = selectedSalaryIds.value.indexOf(item.id)
    if (idx > -1) selectedSalaryIds.value.splice(idx, 1)
    fetchData()
  } catch (error: any) {
    toast.error(error.response?.data?.error || 'Gagal menghapus data')
  }
}

const openWorkerModal = () => {
  newWorkerName.value = ''
  const currentBranchId = masterDataStore.selectedBranchId
  if (currentBranchId && currentBranchId !== 'ALL_BRANCHES') {
    newWorkerBranchId.value = currentBranchId
  } else {
    newWorkerBranchId.value = branches.value.length > 0 ? branches.value[0].id : ''
  }
  showWorkerModal.value = true
}

const saveNewWorker = async () => {
  if (!newWorkerName.value || !newWorkerBranchId.value) {
    toast.error('Nama Pekerja dan Cabang wajib diisi')
    return
  }

  savingWorker.value = true
  try {
    const uniqueId = Math.random().toString(36).substring(2, 8)
    const payload = {
      name: newWorkerName.value,
      email: `harian_${uniqueId}@demo.local`,
      branchId: newWorkerBranchId.value,
      employmentStatus: 'Harian'
    }
    
    await apiClient.post('/employees', payload)
    toast.success('Pekerja baru ditambahkan. Anda bisa memilihnya sekarang.')
    showWorkerModal.value = false
    await fetchData()
    
    const createdWorker = kioskEmployees.value.find(e => e.first_name === newWorkerName.value || `${e.first_name} ${e.last_name || ''}`.trim() === newWorkerName.value)
    if (createdWorker) {
      form.value.daily_worker_id = createdWorker.id
    }
  } catch (error) {
    toast.error('Gagal menambahkan pekerja')
  } finally {
    savingWorker.value = false
  }
}
</script>

<template>
  <div class="p-6 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 flex items-center gap-2">
          <Wallet class="w-6 h-6 text-primary" />
          Penggajian Pekerja Harian
          <span v-if="masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES'" class="text-xs font-semibold px-2.5 py-1 rounded-full bg-primary/10 text-primary border border-primary/20 flex items-center gap-1">
            <Building2 class="w-3.5 h-3.5" /> {{ selectedBranchName }}
          </span>
        </h1>
        <p class="text-slate-500 mt-1">Validasi, bayar, dan hubungkan pengeluaran gaji Kuli, Driver, dan Packing langsung ke KAS Cabang.</p>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        <button 
          @click="openWorkerModal"
          class="px-4 py-2 bg-slate-100 text-slate-700 rounded-xl hover:bg-slate-200 transition-colors flex items-center gap-2 font-medium"
        >
          <Users class="w-4 h-4" /> Tambah Pekerja Harian
        </button>
        <button 
          @click="openAddModal"
          class="px-4 py-2 bg-primary text-white rounded-xl hover:bg-primary/90 transition-colors flex items-center gap-2 font-medium shadow-sm"
        >
          <Plus class="w-4 h-4" /> Input Tagihan Gaji Baru
        </button>
      </div>
    </div>

    <!-- Active Kas Summary Cards -->
    <div v-if="activeKasList.length > 0" class="mb-6 grid grid-cols-1 md:grid-cols-3 gap-3">
      <div 
        v-for="kas in activeKasList.slice(0, 3)" 
        :key="kas.id"
        class="bg-white border border-slate-200/80 rounded-2xl p-4 shadow-sm flex items-center justify-between"
      >
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
            <Coins class="w-5 h-5" />
          </div>
          <div>
            <div class="text-xs font-semibold text-slate-400 uppercase tracking-wider">{{ kas.kas_code }}</div>
            <div class="text-sm font-bold text-slate-800 truncate max-w-[160px]">{{ kas.kas_name }}</div>
          </div>
        </div>
        <div class="text-right">
          <div class="text-xs text-slate-400">Saldo Kas</div>
          <div class="text-sm font-extrabold text-emerald-600">{{ formatRupiah(kas.balance) }}</div>
        </div>
      </div>
    </div>

    <!-- Top Action Bar: Tabs & Batch Pay -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-4">
      <!-- Type Toggle -->
      <div class="flex p-1 bg-slate-100 rounded-xl max-w-sm">
        <button 
          @click="activeTab = 'UNPAID'; selectedSalaryIds = []"
          class="flex-1 py-2 px-3 text-sm font-semibold rounded-lg transition-all flex items-center justify-center gap-2"
          :class="activeTab === 'UNPAID' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
        >
          <History class="w-4 h-4" /> Belum Dibayar
          <span 
            v-if="salaryHistory.filter(s => s.status === 'UNPAID').length > 0"
            class="px-1.5 py-0.5 text-xs rounded-full bg-amber-100 text-amber-800 font-bold"
          >
            {{ salaryHistory.filter(s => s.status === 'UNPAID').length }}
          </span>
        </button>
        <button 
          @click="activeTab = 'PAID'; selectedSalaryIds = []"
          class="flex-1 py-2 px-3 text-sm font-semibold rounded-lg transition-all flex items-center justify-center gap-2"
          :class="activeTab === 'PAID' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
        >
          <CheckCircle class="w-4 h-4" /> Riwayat Lunas
        </button>
      </div>

      <!-- Batch Pay Button (When UNPAID & checked) -->
      <div v-if="activeTab === 'UNPAID' && selectedSalaryIds.length > 0" class="flex items-center gap-3">
        <div class="text-sm font-medium text-slate-600">
          <span class="font-bold text-primary">{{ selectedSalaryIds.length }}</span> tagihan dipilih (Total: <span class="font-bold text-slate-900">{{ formatRupiah(batchTotalAmount) }}</span>)
        </div>
        <button 
          @click="openBatchPayModal"
          class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-medium shadow-sm transition-colors flex items-center gap-2 text-sm"
        >
          <Coins class="w-4 h-4" /> Bayar Sekaligus dari Kas
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="p-4 border-b border-slate-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 bg-slate-50/50">
        <div class="flex items-center gap-2">
          <h2 class="font-semibold text-slate-800">
            {{ activeTab === 'UNPAID' ? 'Daftar Tagihan Belum Dibayar' : 'Riwayat Penggajian Lunas' }}
          </h2>
          <span class="text-xs px-2 py-0.5 bg-slate-200/80 text-slate-700 rounded-full font-medium">
            {{ filteredHistory.length }} data
          </span>
        </div>
        <div class="relative w-full sm:w-72">
          <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input 
            v-model="searchQuery"
            type="text" 
            placeholder="Cari pekerja, catatan, kas..."
            class="w-full pl-9 pr-4 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
          >
        </div>
      </div>
      
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-slate-200">
          <thead class="bg-slate-50 text-xs font-semibold text-slate-500 uppercase tracking-wider">
            <tr>
              <th v-if="activeTab === 'UNPAID'" scope="col" class="w-10 px-4 py-3 text-center">
                <button 
                  @click="toggleSelectAll" 
                  type="button" 
                  class="text-slate-400 hover:text-slate-700 flex items-center justify-center mx-auto"
                >
                  <CheckSquare v-if="isAllSelected" class="w-4 h-4 text-primary" />
                  <Square v-else class="w-4 h-4" />
                </button>
              </th>
              <th scope="col" class="px-6 py-3 text-left">Tanggal & Pekerja</th>
              <th scope="col" class="px-6 py-3 text-left">Tipe Pekerjaan</th>
              <th scope="col" class="px-6 py-3 text-left">Nominal</th>
              <th v-if="activeTab === 'PAID'" scope="col" class="px-6 py-3 text-left">Sumber KAS & Pengeluaran</th>
              <th scope="col" class="px-6 py-3 text-left">Catatan & Bukti</th>
              <th scope="col" class="px-6 py-3 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-slate-200 text-sm">
            <tr 
              v-for="item in filteredHistory" 
              :key="item.id" 
              class="hover:bg-slate-50/80 transition-colors"
              :class="{ 'bg-primary/5': selectedSalaryIds.includes(item.id) }"
            >
              <!-- Checkbox -->
              <td v-if="activeTab === 'UNPAID'" class="px-4 py-4 text-center">
                <button 
                  @click="toggleSelectItem(item.id)" 
                  type="button" 
                  class="text-slate-400 hover:text-slate-700 flex items-center justify-center mx-auto"
                >
                  <CheckSquare v-if="selectedSalaryIds.includes(item.id)" class="w-4 h-4 text-primary" />
                  <Square v-else class="w-4 h-4" />
                </button>
              </td>

              <!-- Tanggal & Pekerja -->
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="font-bold text-slate-900">{{ item.daily_worker?.name || 'Pekerja Harian' }}</div>
                <div class="text-xs text-slate-500 flex items-center gap-2 mt-0.5 flex-wrap">
                  <span class="flex items-center gap-1">
                    <Calendar class="w-3 h-3" />
                    {{ new Date(item.date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) }}
                  </span>
                  <span v-if="item.branch?.name" class="px-2 py-0.5 rounded-md bg-slate-100 text-slate-600 font-medium text-[11px] flex items-center gap-1 border border-slate-200">
                    <Building2 class="w-3 h-3 text-slate-400" /> {{ item.branch.name }}
                  </span>
                </div>
              </td>

              <!-- Tipe Pekerjaan -->
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-200">
                  <Users v-if="item.daily_worker?.type?.toLowerCase().includes('kuli')" class="w-3.5 h-3.5 text-amber-600" />
                  <Truck v-else-if="item.daily_worker?.type?.toLowerCase().includes('driver')" class="w-3.5 h-3.5 text-blue-600" />
                  <Package v-else class="w-3.5 h-3.5 text-indigo-600" />
                  {{ item.daily_worker?.type || 'Harian' }}
                </span>
              </td>

              <!-- Nominal -->
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="font-extrabold text-slate-900 text-base">{{ formatRupiah(item.total_paid) }}</div>
              </td>

              <!-- Sumber KAS & Kategori (Khusus Tab Lunas) -->
              <td v-if="activeTab === 'PAID'" class="px-6 py-4 whitespace-nowrap">
                <div v-if="item.kas_pabrik" class="flex flex-col gap-1">
                  <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-bold bg-amber-50 text-amber-800 border border-amber-200">
                    <Coins class="w-3.5 h-3.5 text-amber-600 shrink-0" />
                    {{ item.kas_pabrik.kas_name }} ({{ item.kas_pabrik.kas_code }})
                  </span>
                  <div class="text-[11px] text-slate-500 flex items-center gap-1">
                    <span>Kat: {{ getCategoryLabel(item.kas_transaction?.expense_category) }}</span>
                  </div>
                </div>
                <div v-else class="text-xs text-slate-400 italic">
                  Tanpa KAS Terhubung
                </div>
              </td>

              <!-- Catatan & Bukti -->
              <td class="px-6 py-4 text-xs text-slate-600 max-w-xs">
                <div class="truncate">{{ item.notes || '-' }}</div>
                <div v-if="item.proof_url" class="mt-1.5">
                  <button 
                    @click="openPreviewModal(getFullImageUrl(item.proof_url), 'Bukti Hitung', item.daily_worker?.name)" 
                    class="font-medium px-2 py-0.5 rounded-full bg-indigo-50 text-indigo-700 hover:bg-indigo-100 border border-indigo-200 inline-flex items-center gap-1 transition-colors text-[11px]"
                  >
                    <Eye class="w-3 h-3" /> Lihat Bukti
                  </button>
                </div>
              </td>

              <!-- Aksi -->
              <td class="px-6 py-4 whitespace-nowrap text-right font-medium">
                <div v-if="item.status === 'UNPAID'" class="flex items-center justify-end gap-1.5">
                  <button 
                    @click="openPayModal(item)" 
                    class="px-3 py-1.5 bg-emerald-50 text-emerald-700 hover:bg-emerald-100 rounded-xl flex items-center gap-1.5 transition-colors border border-emerald-200 font-semibold text-xs shadow-sm"
                  >
                    <Coins class="w-3.5 h-3.5" /> Bayar dari Kas
                  </button>
                  <button 
                    @click="openEditModal(item)" 
                    title="Ubah Tagihan"
                    class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                  >
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button 
                    @click="deleteSalary(item)" 
                    title="Hapus Tagihan"
                    class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
                <div v-else class="flex items-center justify-end gap-2">
                  <span class="text-emerald-700 bg-emerald-50 border border-emerald-200 px-2.5 py-1 rounded-full text-xs font-bold flex items-center gap-1">
                    <CheckCircle class="w-3.5 h-3.5" /> Lunas
                  </span>
                  <button 
                    @click="deleteSalary(item)" 
                    title="Batalkan & Kembalikan Saldo Kas"
                    class="p-1 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>

            <!-- Empty State -->
            <tr v-if="filteredHistory.length === 0">
              <td :colspan="activeTab === 'PAID' ? 6 : 6" class="px-6 py-12 text-center">
                <div class="flex flex-col items-center justify-center text-slate-500">
                  <History class="w-10 h-10 mb-3 text-slate-300" />
                  <p class="font-medium text-slate-900">Tidak ada data</p>
                  <p class="text-sm mt-1">Belum ada data gaji pekerja harian dengan status ini.</p>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal Pembayaran Satuan ("Bayar Gaji dari Kas") -->
    <div v-if="showPayModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm transition-opacity" @click="showPayModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-lg overflow-hidden animate-in fade-in zoom-in-95 duration-200 border border-slate-100">
        <!-- Header -->
        <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-gradient-to-r from-amber-500/10 via-emerald-500/5 to-transparent">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-emerald-600 text-white flex items-center justify-center shadow-md shadow-emerald-200">
              <Coins class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-lg font-bold text-slate-900">Pembayaran Gaji via KAS</h3>
              <p class="text-xs text-slate-500">Otomatis mencatat pengeluaran KAS dan memotong saldo.</p>
            </div>
          </div>
          <button @click="showPayModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-6 space-y-5">
          <!-- Summary Banner -->
          <div class="bg-slate-50 border border-slate-200/80 rounded-2xl p-4 flex items-center justify-between">
            <div>
              <div class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Penerima Gaji</div>
              <div class="text-base font-bold text-slate-900 mt-0.5">{{ selectedPayItem?.daily_worker?.name }}</div>
              <div class="text-xs text-slate-500">{{ selectedPayItem?.daily_worker?.type }}</div>
            </div>
            <div class="text-right">
              <div class="text-xs font-semibold text-slate-400 uppercase tracking-wider">Total Gaji</div>
              <div class="text-xl font-black text-emerald-600 mt-0.5">{{ formatRupiah(payForm.total_paid) }}</div>
            </div>
          </div>

          <!-- Pilihan KAS (Dropdown Utama) -->
          <div class="space-y-1.5">
            <label class="text-sm font-bold text-slate-800 flex items-center justify-between">
              <span class="flex items-center gap-1.5">
                <Coins class="w-4 h-4 text-amber-600" />
                Pilih Sumber KAS Pabrik <span class="text-red-500">*</span>
              </span>
              <span v-if="selectedKasForPay" class="text-xs font-semibold text-emerald-600">
                Saldo: {{ formatRupiah(selectedKasForPay.balance) }}
              </span>
            </label>
            <select 
              v-model="payForm.kas_pabrik_id"
              class="w-full px-3.5 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white text-slate-800 font-medium text-sm"
            >
              <option value="" disabled>-- Pilih Akun Kas --</option>
              <option v-for="kas in activeKasList" :key="kas.id" :value="kas.id">
                [{{ kas.kas_code }}] {{ kas.kas_name }} — Saldo: {{ formatRupiah(kas.balance) }}
              </option>
            </select>
          </div>

          <!-- Kalkulasi Saldo Kas Preview -->
          <div v-if="selectedKasForPay" class="bg-amber-50/50 border border-amber-200/60 rounded-xl p-3.5 text-xs text-slate-700 space-y-1.5">
            <div class="flex justify-between font-medium">
              <span class="text-slate-500">Saldo Awal Kas:</span>
              <span class="font-bold text-slate-800">{{ formatRupiah(selectedKasForPay.balance) }}</span>
            </div>
            <div class="flex justify-between font-medium text-red-600">
              <span>Pengeluaran Gaji:</span>
              <span class="font-bold">- {{ formatRupiah(payForm.total_paid) }}</span>
            </div>
            <div class="border-t border-amber-200/60 pt-1 flex justify-between font-bold text-sm">
              <span class="text-slate-800">Sisa Saldo Kas:</span>
              <span :class="selectedKasForPay.balance - payForm.total_paid < 0 ? 'text-red-600' : 'text-emerald-700'">
                {{ formatRupiah(selectedKasForPay.balance - payForm.total_paid) }}
              </span>
            </div>
            <p v-if="selectedKasForPay.balance - payForm.total_paid < 0" class="text-[11px] text-red-600 font-semibold flex items-center gap-1 mt-1">
              <AlertCircle class="w-3.5 h-3.5 shrink-0" /> Saldo Kas tidak mencukupi untuk pembayaran ini!
            </p>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <!-- Tanggal Pembayaran -->
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Tanggal Pembayaran</label>
              <input 
                v-model="payForm.payment_date" 
                type="date" 
                class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700 bg-white"
              >
            </div>

            <!-- Kategori Pengeluaran -->
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Kategori Kas</label>
              <select 
                v-model="payForm.expense_category" 
                class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white text-slate-700"
              >
                <option v-for="cat in expenseCategories" :key="cat.value" :value="cat.value">
                  {{ cat.label }}
                </option>
              </select>
            </div>
          </div>

          <!-- Catatan / Deskripsi -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Catatan Pengeluaran Kas <span class="text-slate-400 font-normal">(Opsional)</span></label>
            <input 
              v-model="payForm.notes" 
              type="text" 
              placeholder="Contoh: Gaji kuli bongkar kontainer #12..." 
              class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700"
            >
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button 
            @click="showPayModal = false" 
            class="px-4 py-2 text-sm font-semibold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors"
          >
            Batal
          </button>
          <button 
            @click="confirmPaySalary" 
            :disabled="isPaying || !payForm.kas_pabrik_id"
            class="px-5 py-2 bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-bold rounded-xl transition-colors flex items-center gap-2 shadow-sm disabled:opacity-50"
          >
            <Coins class="w-4 h-4" v-if="!isPaying" />
            <span v-if="isPaying">Memproses...</span>
            <span v-else>Konfirmasi & Bayar</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Pembayaran Massal ("Batch Pay") -->
    <div v-if="showBatchPayModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm transition-opacity" @click="showBatchPayModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-lg overflow-hidden animate-in fade-in zoom-in-95 duration-200 border border-slate-100">
        <!-- Header -->
        <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-gradient-to-r from-emerald-500/10 via-amber-500/5 to-transparent">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-emerald-600 text-white flex items-center justify-center shadow-md shadow-emerald-200">
              <Coins class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-lg font-bold text-slate-900">Bayar Sekaligus via KAS</h3>
              <p class="text-xs text-slate-500">Membayar {{ selectedSalaryIds.length }} tagihan gaji pekerja harian sekaligus.</p>
            </div>
          </div>
          <button @click="showBatchPayModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Body -->
        <div class="p-6 space-y-5">
          <!-- Summary Banner -->
          <div class="bg-emerald-50/50 border border-emerald-200/60 rounded-2xl p-4 flex items-center justify-between">
            <div>
              <div class="text-xs font-semibold text-emerald-800 uppercase tracking-wider">Jumlah Tagihan</div>
              <div class="text-xl font-black text-emerald-900 mt-0.5">{{ selectedSalaryIds.length }} Orang</div>
            </div>
            <div class="text-right">
              <div class="text-xs font-semibold text-emerald-800 uppercase tracking-wider">Total Pembayaran</div>
              <div class="text-2xl font-black text-emerald-600 mt-0.5">{{ formatRupiah(batchTotalAmount) }}</div>
            </div>
          </div>

          <!-- Pilihan KAS (Dropdown Utama) -->
          <div class="space-y-1.5">
            <label class="text-sm font-bold text-slate-800 flex items-center justify-between">
              <span class="flex items-center gap-1.5">
                <Coins class="w-4 h-4 text-amber-600" />
                Pilih Sumber KAS Pabrik <span class="text-red-500">*</span>
              </span>
              <span v-if="selectedKasForBatchPay" class="text-xs font-semibold text-emerald-600">
                Saldo: {{ formatRupiah(selectedKasForBatchPay.balance) }}
              </span>
            </label>
            <select 
              v-model="batchPayForm.kas_pabrik_id"
              class="w-full px-3.5 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white text-slate-800 font-medium text-sm"
            >
              <option value="" disabled>-- Pilih Akun Kas --</option>
              <option v-for="kas in activeKasList" :key="kas.id" :value="kas.id">
                [{{ kas.kas_code }}] {{ kas.kas_name }} — Saldo: {{ formatRupiah(kas.balance) }}
              </option>
            </select>
          </div>

          <!-- Kalkulasi Saldo Kas Preview -->
          <div v-if="selectedKasForBatchPay" class="bg-amber-50/50 border border-amber-200/60 rounded-xl p-3.5 text-xs text-slate-700 space-y-1.5">
            <div class="flex justify-between font-medium">
              <span class="text-slate-500">Saldo Awal Kas:</span>
              <span class="font-bold text-slate-800">{{ formatRupiah(selectedKasForBatchPay.balance) }}</span>
            </div>
            <div class="flex justify-between font-medium text-red-600">
              <span>Total Pengeluaran:</span>
              <span class="font-bold">- {{ formatRupiah(batchTotalAmount) }}</span>
            </div>
            <div class="border-t border-amber-200/60 pt-1 flex justify-between font-bold text-sm">
              <span class="text-slate-800">Sisa Saldo Kas:</span>
              <span :class="selectedKasForBatchPay.balance - batchTotalAmount < 0 ? 'text-red-600' : 'text-emerald-700'">
                {{ formatRupiah(selectedKasForBatchPay.balance - batchTotalAmount) }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <!-- Tanggal Pembayaran -->
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Tanggal Pembayaran</label>
              <input 
                v-model="batchPayForm.payment_date" 
                type="date" 
                class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700 bg-white"
              >
            </div>

            <!-- Kategori Pengeluaran -->
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Kategori Kas</label>
              <select 
                v-model="batchPayForm.expense_category" 
                class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white text-slate-700"
              >
                <option v-for="cat in expenseCategories" :key="cat.value" :value="cat.value">
                  {{ cat.label }}
                </option>
              </select>
            </div>
          </div>

          <!-- Catatan -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Catatan Batch <span class="text-slate-400 font-normal">(Opsional)</span></label>
            <input 
              v-model="batchPayForm.notes" 
              type="text" 
              placeholder="Catatan untuk pembayaran batch ini..." 
              class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700"
            >
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button 
            @click="showBatchPayModal = false" 
            class="px-4 py-2 text-sm font-semibold text-slate-600 rounded-xl hover:bg-slate-200 transition-colors"
          >
            Batal
          </button>
          <button 
            @click="confirmBatchPay" 
            :disabled="isBatchPaying || !batchPayForm.kas_pabrik_id"
            class="px-5 py-2 bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-bold rounded-xl transition-colors flex items-center gap-2 shadow-sm disabled:opacity-50"
          >
            <Coins class="w-4 h-4" v-if="!isBatchPaying" />
            <span v-if="isBatchPaying">Memproses Pembayaran...</span>
            <span v-else>Bayar Semua ({{ formatRupiah(batchTotalAmount) }})</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Form (Input Tagihan Gaji Baru / Ubah) -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-200 border border-slate-100">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-lg font-bold text-slate-900">{{ editMode ? 'Ubah Tagihan Gaji' : 'Input Tagihan Gaji Baru' }}</h3>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">
            ✕
          </button>
        </div>
        
        <div class="p-6 space-y-4 max-h-[75vh] overflow-y-auto">
          <!-- Pilih Pekerja -->
          <div class="space-y-1.5" v-if="!editMode">
            <label class="text-sm font-medium text-slate-700">Nama Pekerja (Dari Data Kiosk) <span class="text-red-500">*</span></label>
            <div class="flex gap-2">
              <select v-model="form.daily_worker_id" class="flex-1 px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white text-sm">
                <option value="">Pilih Pekerja...</option>
                <option v-for="w in availableKioskEmployees" :key="w.id" :value="w.id">
                  {{ w.first_name }} {{ w.last_name || '' }} ({{ w.pin ? 'PIN: '+w.pin : 'Tanpa PIN' }})
                </option>
              </select>
            </div>
          </div>

          <!-- Tipe Pekerjaan -->
          <div class="space-y-1.5" v-if="!editMode">
            <label class="text-sm font-medium text-slate-700">Tipe Pekerjaan <span class="text-red-500">*</span></label>
            <select 
              v-model="form.type" 
              @change="onFormTypeChange"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white text-slate-700 text-sm"
            >
              <option value="Kuli">Kuli Bongkar</option>
              <option value="Driver">Driver</option>
              <option value="Packing Kecap">Packing Kecap</option>
              <option value="Packing Saos">Packing Saos</option>
              <option value="OVEN BOTOL">OVEN BOTOL</option>
              <option value="PRODUKSI KECAP">PRODUKSI KECAP</option>
              <option value="PRODUKSI SAOS">PRODUKSI SAOS</option>
              <option value="DROP GULA">DROP GULA</option>
              <option value="TIM MUAT & BONGKAR TRONTON">TIM MUAT & BONGKAR TRONTON</option>
              <option value="TIM MUAT NON TRONTON">TIM MUAT NON TRONTON</option>
              <option value="TIM BONGKAR BAHAN BAKU">TIM BONGKAR BAHAN BAKU</option>
              <option value="TIM CLEANING KECAP">TIM CLEANING KECAP</option>
              <option value="TIM CLEANING SAOS">TIM CLEANING SAOS</option>
              <option value="TEKNISI PABRIK">TEKNISI PABRIK</option>
              <option value="TUKANG PABRIK">TUKANG PABRIK</option>
              <option value="HARIAN PABRIK">HARIAN PABRIK</option>
            </select>
          </div>

          <!-- Nominal -->
          <div class="space-y-1.5">
            <label class="text-sm font-medium text-slate-700 flex justify-between">
              <span>Nominal Pembayaran (Rp) <span class="text-red-500">*</span></span>
            </label>
            <input v-model="form.total_paid" type="number" min="0" placeholder="0" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all font-semibold text-slate-900 text-sm">
          </div>

          <!-- Catatan -->
          <div class="space-y-1.5">
            <label class="text-sm font-medium text-slate-700">Catatan <span class="text-xs text-slate-400 font-normal">(Opsional)</span></label>
            <textarea v-model="form.notes" rows="2" placeholder="Detail pekerjaan..." class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700 text-sm"></textarea>
          </div>

          <!-- Foto Bukti -->
          <div class="space-y-1.5" v-if="!editMode">
            <label class="text-sm font-medium text-slate-700 flex items-center gap-1">
              Upload Foto Bukti Hitung <span class="text-xs text-slate-400 font-normal">(Opsional)</span>
            </label>
            <input ref="proofFileInput" type="file" accept="image/*" @change="onFileChangeProof" class="hidden" />
            <button type="button" @click="triggerProofFilePicker"
              class="w-full flex items-center gap-3 px-4 py-2 border border-slate-200 rounded-xl text-sm text-slate-600 hover:bg-slate-50 transition-colors text-left">
              <span class="shrink-0 text-xs font-semibold bg-indigo-50 text-indigo-600 px-2.5 py-1 rounded-lg">Pilih File</span>
              <span class="truncate">{{ proofFile ? proofFile.name : 'Belum ada file dipilih...' }}</span>
            </button>
          </div>

          <!-- Opsi Langsung Bayar dari KAS (Hanya saat create baru) -->
          <div v-if="!editMode" class="border-t border-slate-100 pt-4 space-y-3">
            <label class="flex items-center gap-2 cursor-pointer select-none">
              <input 
                type="checkbox" 
                v-model="form.auto_pay"
                class="rounded text-primary focus:ring-primary w-4 h-4"
              >
              <span class="text-sm font-bold text-slate-800 flex items-center gap-1.5">
                <Coins class="w-4 h-4 text-amber-600" />
                Langsung Bayar dari KAS Pabrik
              </span>
            </label>

            <!-- Dropdown Kas jika auto_pay aktif -->
            <div v-if="form.auto_pay" class="bg-amber-50/60 border border-amber-200/80 rounded-2xl p-3.5 space-y-3 animate-in fade-in duration-150">
              <div class="space-y-1">
                <label class="text-xs font-bold text-slate-700">Pilih Kas Pabrik <span class="text-red-500">*</span></label>
                <select 
                  v-model="form.kas_pabrik_id"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl bg-white text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-primary/20"
                >
                  <option v-for="kas in activeKasList" :key="kas.id" :value="kas.id">
                    [{{ kas.kas_code }}] {{ kas.kas_name }} (Saldo: {{ formatRupiah(kas.balance) }})
                  </option>
                </select>
              </div>

              <div class="grid grid-cols-2 gap-2">
                <div class="space-y-1">
                  <label class="text-xs font-bold text-slate-700">Tanggal Bayar</label>
                  <input v-model="form.payment_date" type="date" class="w-full px-2.5 py-1.5 border border-slate-200 rounded-xl text-xs bg-white text-slate-700">
                </div>
                <div class="space-y-1">
                  <label class="text-xs font-bold text-slate-700">Kategori Kas</label>
                  <select v-model="form.expense_category" class="w-full px-2.5 py-1.5 border border-slate-200 rounded-xl text-xs bg-white text-slate-700">
                    <option v-for="cat in expenseCategories" :key="cat.value" :value="cat.value">{{ cat.label }}</option>
                  </select>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button @click="showModal = false" class="px-4 py-2 font-medium text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">
            Batal
          </button>
          <button 
            @click="submitForm" 
            :disabled="isSubmitting"
            class="px-5 py-2 bg-primary text-white font-bold rounded-xl hover:bg-primary/90 transition-colors flex items-center gap-2 shadow-sm text-sm disabled:opacity-50"
          >
            <Save class="w-4 h-4" v-if="!isSubmitting" />
            <span v-if="isSubmitting">Menyimpan...</span>
            <span v-else>{{ form.auto_pay ? 'Simpan & Bayar Kas' : 'Simpan Tagihan' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Worker Modal -->
    <div v-if="showWorkerModal" class="fixed inset-0 z-[60] overflow-y-auto flex items-center justify-center p-4">
      <div class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm transition-opacity" @click="showWorkerModal = false"></div>
      <div class="relative bg-white rounded-3xl shadow-xl w-full max-w-sm overflow-hidden animate-in fade-in zoom-in-95 duration-200 border border-slate-100">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-lg font-bold text-slate-900">Tambah Pekerja Harian</h3>
          <button @click="showWorkerModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition-colors">
            ✕
          </button>
        </div>
        
        <div class="p-6 space-y-4">
          <div class="space-y-1.5">
            <label class="text-sm font-medium text-slate-700">Nama Lengkap</label>
            <input v-model="newWorkerName" type="text" placeholder="Masukkan nama..." class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700 text-sm">
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-medium text-slate-700">Cabang</label>
            <select v-model="newWorkerBranchId" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white text-slate-700 text-sm">
              <option v-for="b in branches" :key="b.id" :value="b.id">
                {{ b.name }}
              </option>
            </select>
            <p class="text-xs text-slate-500 mt-1">Pekerja akan langsung mendapatkan PIN Kiosk setelah disimpan.</p>
          </div>
        </div>

        <div class="p-4 border-t border-slate-100 bg-slate-50 flex justify-end gap-3">
          <button @click="showWorkerModal = false" class="px-4 py-2 font-medium text-slate-600 rounded-xl hover:bg-slate-200 transition-colors text-sm">
            Batal
          </button>
          <button @click="saveNewWorker" :disabled="savingWorker" class="px-4 py-2 bg-primary text-white font-bold rounded-xl hover:bg-primary/90 transition-colors flex items-center gap-2 shadow-sm disabled:opacity-50 text-sm">
            <Save class="w-4 h-4" v-if="!savingWorker" /> 
            <span v-if="savingWorker">Menyimpan...</span>
            <span v-else>Tambah</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Image Preview Modal -->
    <div v-if="showPreviewModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/80 backdrop-blur-sm transition-opacity">
      <div class="bg-white rounded-3xl shadow-2xl w-full max-w-3xl overflow-hidden flex flex-col max-h-[90vh]">
        <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div>
            <h3 class="font-bold text-slate-900">{{ previewModalTitle }}</h3>
            <p v-if="previewModalSubtitle" class="text-sm text-slate-500">{{ previewModalSubtitle }}</p>
          </div>
          <button @click="closePreviewModal" class="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 overflow-auto flex-1 flex items-center justify-center bg-slate-100/50">
          <img :src="previewImageUrl" alt="Preview Bukti" class="max-w-full max-h-full rounded-2xl shadow-sm border border-slate-200 object-contain" />
        </div>
      </div>
    </div>
  </div>
</template>

