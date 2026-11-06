<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Receipt,
  Search,
  Calendar,
  Filter,
  Plus,
  Trash2,
  Edit2,
  CheckCircle2,
  XCircle,
  Clock,
  Image as ImageIcon,
  Building2,
  Store,
  Users,
  ZoomIn,
  ZoomOut,
  X,
  FileText,
  DollarSign,
  Package,
  Eye,
  Check,
  Ban,
  UploadCloud,
  Download,
  Award,
  Sparkles,
  TrendingUp,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { filterSellableProducts } from '@/utils/productUnits'
import { useMasterDataStore } from '@/stores/masterData'
import * as XLSX from 'xlsx'
import { saveAs } from 'file-saver'

const masterDataStore = useMasterDataStore()

interface SpgSaleItem {
  id?: string
  product_id?: string
  product_name: string
  sku_code: string
  unit_price: number
  quantity: number
  subtotal: number
}

interface SpgReport {
  id: string
  schedule_id?: string
  spg_id: string
  key_account_id: string
  supermarket_name: string
  report_date: string
  shift: string
  target_amount: number
  total_amount: number
  total_qty: number
  base_fee: number
  incentive_bonus: number
  total_spg_payout: number
  is_target_achieved: boolean
  receipt_photo_url: string
  attendance_photo_url: string
  notes: string
  status: 'SUBMITTED' | 'APPROVED' | 'REJECTED'
  verified_by?: string
  verified_at?: string
  created_at: string
  schedule?: {
    id: string
    supermarket_name: string
    day_pattern: string
    target_sales: number
  }
  spg?: {
    id: string
    name: string
    supermarket_name: string
    phone_number: string
    photo_url: string
  }
  key_account?: {
    id: string
    name: string
    company_name: string
  }
  items?: SpgSaleItem[]
}

const reports = ref<SpgReport[]>([])
const spgs = ref<any[]>([])
const keyAccounts = ref<any[]>([])
const schedules = ref<any[]>([])
const products = ref<any[]>([])
const isLoading = ref(false)

// Filters
const searchQuery = ref('')
const selectedStatus = ref<'ALL' | 'SUBMITTED' | 'APPROVED' | 'REJECTED'>('ALL')
const selectedSpgId = ref('ALL')
const startDate = ref('')
const endDate = ref('')

// Image Preview State
const previewImageUrl = ref('')
const zoomLevel = ref(1)

const openPreview = (url: string) => {
  if (!url) return
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

// Fetch Data
const fetchReports = async () => {
  isLoading.value = true
  try {
    const params: any = {}
    if (selectedStatus.value !== 'ALL') params.status = selectedStatus.value
    if (selectedSpgId.value !== 'ALL') params.spg_id = selectedSpgId.value
    if (startDate.value) params.start_date = startDate.value
    if (endDate.value) params.end_date = endDate.value

    const res = await apiClient.get('/modern-trade/spg-reports', { params })
    reports.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil laporan penjualan SPG: ' + (err.response?.data?.error || err.message))
  } finally {
    isLoading.value = false
  }
}

const fetchSpgs = async () => {
  try {
    const res = await apiClient.get('/modern-trade/spg')
    spgs.value = res.data.data || []
  } catch (err: any) {
    console.error(err)
  }
}

const fetchKeyAccounts = async () => {
  try {
    const res = await apiClient.get('/modern-trade/key-accounts')
    keyAccounts.value = res.data.data || []
  } catch (err: any) {
    console.error(err)
  }
}

const fetchSchedules = async () => {
  try {
    const res = await apiClient.get('/modern-trade/spg-schedules')
    schedules.value = res.data.data || []
  } catch (err: any) {
    console.error(err)
  }
}

const fetchProducts = async () => {
  try {
    const res = await apiClient.get('/factory/products')
    const rawList = res.data?.data || res.data || []
    products.value = filterSellableProducts(rawList)
  } catch (err: any) {
    console.error(err)
  }
}

onMounted(() => {
  fetchReports()
  fetchSpgs()
  fetchKeyAccounts()
  fetchSchedules()
  fetchProducts()
})

// Statistics & Computations
const filteredReports = computed(() => {
  return reports.value.filter((rep) => {
    const q = searchQuery.value.toLowerCase()
    const matchesSearch =
      !q ||
      (rep.spg?.name && rep.spg.name.toLowerCase().includes(q)) ||
      (rep.key_account?.name && rep.key_account.name.toLowerCase().includes(q)) ||
      (rep.supermarket_name && rep.supermarket_name.toLowerCase().includes(q)) ||
      (rep.notes && rep.notes.toLowerCase().includes(q))

    const matchesStatus = selectedStatus.value === 'ALL' || rep.status === selectedStatus.value
    const matchesSpg = selectedSpgId.value === 'ALL' || rep.spg_id === selectedSpgId.value

    return matchesSearch && matchesStatus && matchesSpg
  })
})

const totalOmzetApproved = computed(() => {
  return reports.value
    .filter((r) => r.status === 'APPROVED')
    .reduce((sum, r) => sum + Number(r.total_amount || 0), 0)
})

const totalPayoutAll = computed(() => {
  return reports.value
    .filter((r) => r.status === 'APPROVED')
    .reduce((sum, r) => sum + Number(r.total_spg_payout || r.base_fee || 200000), 0)
})

const totalBonusGiven = computed(() => {
  return reports.value
    .filter((r) => r.status === 'APPROVED' && r.is_target_achieved)
    .reduce((sum, r) => sum + Number(r.incentive_bonus || 50000), 0)
})

const targetAchievedCount = computed(() => {
  return reports.value.filter((r) => r.is_target_achieved).length
})

const pendingReportsCount = computed(() => {
  return reports.value.filter((r) => r.status === 'SUBMITTED').length
})

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    })
  } catch {
    return dateStr
  }
}

// Modal Form State (Create / Edit)
const showFormModal = ref(false)
const isEditing = ref(false)
const formReport = ref<{
  id: string
  schedule_id?: string
  spg_id: string
  key_account_id: string
  supermarket_name: string
  report_date: string
  shift: string
  target_amount: number
  base_fee: number
  receipt_photo_url: string
  attendance_photo_url: string
  notes: string
  items: SpgSaleItem[]
}>({
  id: '',
  schedule_id: '',
  spg_id: '',
  key_account_id: '',
  supermarket_name: '',
  report_date: new Date().toISOString().substring(0, 10),
  shift: 'Pagi (10:00 - 19:00)',
  target_amount: 1000000,
  base_fee: 200000,
  receipt_photo_url: '',
  attendance_photo_url: '',
  notes: '',
  items: [],
})

const onScheduleSelected = () => {
  if (!formReport.value.schedule_id) return
  const sched = schedules.value.find((s) => s.id === formReport.value.schedule_id)
  if (sched) {
    formReport.value.spg_id = sched.spg_id
    formReport.value.key_account_id = sched.key_account_id
    formReport.value.supermarket_name = sched.supermarket_name
    if (sched.target_sales && sched.working_days_count) {
      formReport.value.target_amount = Math.round(Number(sched.target_sales) / Number(sched.working_days_count))
    }
    if (sched.daily_fee) {
      formReport.value.base_fee = Number(sched.daily_fee)
    }
  }
}

const onSpgSelected = () => {
  const chosenSpg = spgs.value.find((s) => s.id === formReport.value.spg_id)
  if (chosenSpg) {
    formReport.value.key_account_id = chosenSpg.key_account_id || ''
    formReport.value.supermarket_name = chosenSpg.supermarket_name || ''
    if (chosenSpg.shift_schedule) {
      formReport.value.shift = chosenSpg.shift_schedule
    }
    // Try auto-find schedule for this SPG
    const activeSched = schedules.value.find((s) => s.spg_id === chosenSpg.id && s.status === 'ACTIVE')
    if (activeSched) {
      formReport.value.schedule_id = activeSched.id
      onScheduleSelected()
    }
  }
}

const openNewReportModal = () => {
  isEditing.value = false
  const defaultSpg = spgs.value[0]
  formReport.value = {
    id: '',
    schedule_id: '',
    spg_id: defaultSpg?.id || '',
    key_account_id: defaultSpg?.key_account_id || keyAccounts.value[0]?.id || '',
    supermarket_name: defaultSpg?.supermarket_name || '',
    report_date: new Date().toISOString().substring(0, 10),
    shift: defaultSpg?.shift_schedule || 'Weekend Shift (10:00 - 19:00)',
    target_amount: 1000000,
    base_fee: 200000,
    receipt_photo_url: '',
    attendance_photo_url: '',
    notes: '',
    items: [
      {
        product_name: '',
        sku_code: '',
        unit_price: 0,
        quantity: 1,
        subtotal: 0,
      },
    ],
  }
  showFormModal.value = true
}

const openEditReportModal = (rep: SpgReport) => {
  isEditing.value = true
  formReport.value = {
    id: rep.id,
    schedule_id: rep.schedule_id || '',
    spg_id: rep.spg_id,
    key_account_id: rep.key_account_id,
    supermarket_name: rep.supermarket_name,
    report_date: rep.report_date ? rep.report_date.substring(0, 10) : new Date().toISOString().substring(0, 10),
    shift: rep.shift || '',
    target_amount: Number(rep.target_amount || 0),
    base_fee: Number(rep.base_fee || 200000),
    receipt_photo_url: rep.receipt_photo_url || '',
    attendance_photo_url: rep.attendance_photo_url || '',
    notes: rep.notes || '',
    items: rep.items && rep.items.length > 0
      ? rep.items.map((it) => ({
          id: it.id,
          product_id: it.product_id,
          product_name: it.product_name,
          sku_code: it.sku_code || '',
          unit_price: Number(it.unit_price || 0),
          quantity: Number(it.quantity || 1),
          subtotal: Number(it.subtotal || 0),
        }))
      : [
          {
            product_name: '',
            sku_code: '',
            unit_price: 0,
            quantity: 1,
            subtotal: 0,
          },
        ],
  }
  showFormModal.value = true
}

const addItemRow = () => {
  formReport.value.items.push({
    product_name: '',
    sku_code: '',
    unit_price: 0,
    quantity: 1,
    subtotal: 0,
  })
}

const removeItemRow = (index: number) => {
  if (formReport.value.items.length > 1) {
    formReport.value.items.splice(index, 1)
  } else {
    toast.warning('Minimal harus ada 1 item produk')
  }
}

const onProductSelectChange = (item: SpgSaleItem, prodId: string) => {
  const p = products.value.find((prod) => prod.id === prodId)
  if (p) {
    item.product_id = p.id
    item.product_name = p.name
    item.sku_code = p.code || p.sku || ''
    item.unit_price = Number(p.price || p.selling_price || 0)
    item.subtotal = item.unit_price * item.quantity
  }
}

const updateSubtotal = (item: SpgSaleItem) => {
  item.subtotal = Number(item.unit_price || 0) * Number(item.quantity || 0)
}

const formTotalAmount = computed(() => {
  return formReport.value.items.reduce((sum, it) => sum + Number(it.subtotal || 0), 0)
})

const formTotalQty = computed(() => {
  return formReport.value.items.reduce((sum, it) => sum + Number(it.quantity || 0), 0)
})

// Realtime Incentive Calculation
const isTargetAchievedInForm = computed(() => {
  return formReport.value.target_amount > 0 && formTotalAmount.value >= formReport.value.target_amount
})

const estimatedPayoutInForm = computed(() => {
  const base = Number(formReport.value.base_fee || 200000)
  const bonus = isTargetAchievedInForm.value ? 50000 : 0
  return base + bonus
})

// Photo Upload Handler
const isUploading = ref(false)
const uploadReceiptPhoto = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  isUploading.value = true
  try {
    const res = await apiClient.post('/modern-trade/spg-reports/upload-receipt', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    const baseUrl = import.meta.env.VITE_API_BASE_URL?.replace('/api', '') || ''
    formReport.value.receipt_photo_url = baseUrl + res.data.url
    toast.success('Foto struk / bukti penjualan berhasil diunggah')
  } catch (err: any) {
    toast.error('Gagal mengunggah foto: ' + (err.response?.data?.error || err.message))
  } finally {
    isUploading.value = false
    target.value = ''
  }
}

const saveReport = async () => {
  if (!formReport.value.spg_id || !formReport.value.key_account_id || !formReport.value.supermarket_name) {
    toast.error('SPG, Ritel Key Account, dan Nama Supermarket wajib diisi!')
    return
  }

  const validItems = formReport.value.items.filter((it) => it.product_name && it.quantity > 0)
  if (validItems.length === 0) {
    toast.error('Wajib mengisi minimal 1 item produk dengan kuantiti > 0!')
    return
  }

  try {
    const payload = {
      ...formReport.value,
      target_amount: Number(formReport.value.target_amount),
      base_fee: Number(formReport.value.base_fee),
      items: validItems.map((it) => ({
        product_id: it.product_id || undefined,
        product_name: it.product_name,
        sku_code: it.sku_code,
        unit_price: Number(it.unit_price),
        quantity: Number(it.quantity),
        subtotal: Number(it.unit_price) * Number(it.quantity),
      })),
    }

    if (isEditing.value) {
      await apiClient.put(`/modern-trade/spg-reports/${formReport.value.id}`, payload)
      toast.success('Laporan penjualan SPG berhasil diperbarui!')
    } else {
      await apiClient.post('/modern-trade/spg-reports', payload)
      toast.success('Laporan penjualan SPG baru berhasil dibuat!')
    }
    showFormModal.value = false
    fetchReports()
    fetchSpgs()
  } catch (err: any) {
    toast.error('Gagal menyimpan laporan: ' + (err.response?.data?.error || err.message))
  }
}

const deleteReport = async (id: string) => {
  if (confirm('Apakah Anda yakin ingin menghapus laporan penjualan ini?')) {
    try {
      await apiClient.delete(`/modern-trade/spg-reports/${id}`)
      toast.success('Laporan penjualan berhasil dihapus!')
      fetchReports()
      fetchSpgs()
    } catch (err: any) {
      toast.error('Gagal menghapus laporan: ' + (err.response?.data?.error || err.message))
    }
  }
}

// Verification Modal State
const showVerifyModal = ref(false)
const selectedReportForVerify = ref<SpgReport | null>(null)
const verifyStatus = ref<'APPROVED' | 'REJECTED'>('APPROVED')
const verifyNotes = ref('')

const openVerifyDialog = (rep: SpgReport, status: 'APPROVED' | 'REJECTED') => {
  selectedReportForVerify.value = rep
  verifyStatus.value = status
  verifyNotes.value = rep.notes || ''
  showVerifyModal.value = true
}

const submitVerification = async () => {
  if (!selectedReportForVerify.value) return
  try {
    await apiClient.put(`/modern-trade/spg-reports/${selectedReportForVerify.value.id}/verify`, {
      status: verifyStatus.value,
      notes: verifyNotes.value,
    })
    toast.success(
      verifyStatus.value === 'APPROVED'
        ? 'Laporan penjualan berhasil disetujui! Omzet & honor SPG diperbarui.'
        : 'Laporan penjualan telah ditolak.',
    )
    showVerifyModal.value = false
    fetchReports()
    fetchSpgs()
  } catch (err: any) {
    toast.error('Gagal memverifikasi laporan: ' + (err.response?.data?.error || err.message))
  }
}

// Excel Import & Template Download
const showImportModal = ref(false)
const isImporting = ref(false)
const parsedExcelData = ref<any[]>([])

const downloadTemplate = () => {
  const wsData = [
    [
      'Nama SPG',
      'Ritel MT / Key Account',
      'Nama Supermarket Outlet',
      'Tanggal (YYYY-MM-DD)',
      'Shift',
      'Target Penjualan (IDR)',
      'Gaji Dasar (IDR)',
      'Nama Produk',
      'SKU',
      'Harga Satuan (IDR)',
      'Jumlah Qty',
      'Catatan',
    ],
    [
      'Siska Nirmala',
      'Yogya Group',
      'Yogya Riau Junction Bandung',
      new Date().toISOString().substring(0, 10),
      'Weekend (10:00 - 19:00)',
      1000000,
      200000,
      'Karsa HR Choco Crunch 250gr',
      'SKU-WCC-250',
      25000,
      40,
      'Promo Weekend Beli 2 Gratis Toples',
    ],
    [
      'Siska Nirmala',
      'Yogya Group',
      'Yogya Riau Junction Bandung',
      new Date().toISOString().substring(0, 10),
      'Weekend (10:00 - 19:00)',
      1000000,
      200000,
      'Karsa HR Wafer Vanilla 150gr',
      'SKU-WWV-150',
      15000,
      20,
      '',
    ],
  ]

  const ws = XLSX.utils.aoa_to_sheet(wsData)
  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, 'Template Penjualan SPG')
  const wbout = XLSX.write(wb, { bookType: 'xlsx', type: 'array' })
  saveAs(new Blob([wbout], { type: 'application/octet-stream' }), 'Template_Laporan_Penjualan_SPG.xlsx')
  toast.success('Template Excel berhasil diunduh!')
}

const handleExcelUpload = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (evt) => {
    try {
      const data = new Uint8Array(evt.target?.result as ArrayBuffer)
      const workbook = XLSX.read(data, { type: 'array' })
      const sheetName = workbook.SheetNames[0]
      if (!sheetName) {
        toast.error('Sheet Excel tidak ditemukan!')
        return
      }
      const firstSheet = workbook.Sheets[sheetName]
      if (!firstSheet) {
        toast.error('Sheet Excel tidak dapat dibaca!')
        return
      }
      const rows: any[] = XLSX.utils.sheet_to_json(firstSheet, { header: 1 })

      if (rows.length < 2) {
        toast.error('File Excel kosong atau format tidak sesuai!')
        return
      }

      // Group rows by SPG + Store + Date
      const groups: { [key: string]: any } = {}
      for (let i = 1; i < rows.length; i++) {
        const row = rows[i]
        if (!row || row.length === 0 || !row[0]) continue

        const spgName = String(row[0] || '').trim()
        const kaName = String(row[1] || '').trim()
        const storeName = String(row[2] || '').trim()
        const dateStr = String(row[3] || '').trim()
        const shift = String(row[4] || 'Regular').trim()
        const targetAmount = Number(row[5] || 1000000)
        const baseFee = Number(row[6] || 200000)
        const prodName = String(row[7] || '').trim()
        const sku = String(row[8] || '').trim()
        const unitPrice = Number(row[9] || 0)
        const qty = Number(row[10] || 1)
        const notes = String(row[11] || '').trim()

        const groupKey = `${spgName}_${storeName}_${dateStr}`
        if (!groups[groupKey]) {
          // Resolve SPG & Key Account
          const matchedSpg = spgs.value.find((s) => s.name.toLowerCase() === spgName.toLowerCase()) || spgs.value[0]
          const matchedKa = keyAccounts.value.find((k) => k.name.toLowerCase() === kaName.toLowerCase()) || keyAccounts.value[0]

          groups[groupKey] = {
            spg_id: matchedSpg?.id || '',
            spg_name: spgName,
            key_account_id: matchedKa?.id || '',
            key_account_name: kaName,
            supermarket_name: storeName,
            report_date: dateStr || new Date().toISOString().substring(0, 10),
            shift: shift,
            target_amount: targetAmount,
            base_fee: baseFee,
            notes: notes,
            items: [],
          }
        }

        if (prodName) {
          groups[groupKey].items.push({
            product_name: prodName,
            sku_code: sku,
            unit_price: unitPrice,
            quantity: qty,
            subtotal: unitPrice * qty,
          })
        }
      }

      parsedExcelData.value = Object.values(groups)
      showImportModal.value = true
      toast.success(`Berhasil mem-parsing ${parsedExcelData.value.length} sesi laporan penjualan dari Excel`)
    } catch (err: any) {
      toast.error('Gagal membaca file Excel: ' + err.message)
    } finally {
      target.value = ''
    }
  }
  reader.readAsArrayBuffer(file)
}

const submitParsedExcel = async () => {
  if (parsedExcelData.value.length === 0) return
  isImporting.value = true
  let successCount = 0
  try {
    for (const reportData of parsedExcelData.value) {
      try {
        await apiClient.post('/modern-trade/spg-reports', reportData)
        successCount++
      } catch (err) {
        console.error('Failed to import row', reportData, err)
      }
    }
    toast.success(`Berhasil mengimpor ${successCount} laporan penjualan SPG!`)
    showImportModal.value = false
    parsedExcelData.value = []
    fetchReports()
    fetchSpgs()
  } catch (err: any) {
    toast.error('Terjadi kendala saat impor: ' + err.message)
  } finally {
    isImporting.value = false
  }
}

// Detail View Modal State
const showDetailModal = ref(false)
const detailedReport = ref<SpgReport | null>(null)

const openDetailDialog = (rep: SpgReport) => {
  detailedReport.value = rep
  showDetailModal.value = true
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2.5 py-1 text-xs font-bold text-indigo-600 bg-indigo-50 border border-indigo-200/50 rounded-full">
            Modern Trade (MT)
          </span>
        </div>
        <h1 class="text-2xl md:text-3xl font-black text-slate-900 tracking-tight mt-1 flex items-center flex-wrap gap-2">
          <span>Laporan Penjualan & Insentif SPG</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl md:text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Pencatatan manual/Excel dan kalkulasi gaji dinamis (Gaji Rp 200.000 + Bonus Target Rp 50.000 / shift).
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-2.5">
        <!-- Download Template -->
        <button
          @click="downloadTemplate"
          class="inline-flex items-center px-3.5 py-2 bg-white hover:bg-slate-50 text-slate-700 text-xs font-bold rounded-xl border border-slate-200 shadow-xs transition-all"
          title="Unduh Template Excel"
        >
          <Download class="w-4 h-4 mr-1.5 text-indigo-600" />
          Template Excel
        </button>

        <!-- Import Excel -->
        <label class="inline-flex items-center px-3.5 py-2 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 text-xs font-bold rounded-xl border border-emerald-200 shadow-xs transition-all cursor-pointer">
          <UploadCloud class="w-4 h-4 mr-1.5 text-emerald-600" />
          Import Excel
          <input type="file" accept=".xlsx, .xls" @change="handleExcelUpload" class="hidden" />
        </label>

        <!-- Buat Manual -->
        <Button
          @click="openNewReportModal"
          class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl shadow-md border border-indigo-200"
        >
          <Plus class="w-4 h-4 mr-1.5" />
          Input Manual
        </Button>
      </div>
    </div>

    <!-- Quick Stats Cards with Incentive Highlights -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
          <DollarSign class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Omzet Disetujui</p>
          <p class="text-xl font-black text-emerald-700">{{ formatCurrency(totalOmzetApproved) }}</p>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-amber-50 text-amber-600 rounded-xl">
          <Award class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Bonus Capai Target</p>
          <p class="text-xl font-black text-amber-600">{{ formatCurrency(totalBonusGiven) }}</p>
          <span class="text-[10px] text-slate-400 font-bold">{{ targetAchievedCount }} shift capai target</span>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-indigo-50 text-indigo-600 rounded-xl">
          <TrendingUp class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Pembayaran SPG</p>
          <p class="text-xl font-black text-indigo-700">{{ formatCurrency(totalPayoutAll) }}</p>
          <span class="text-[10px] text-slate-400 font-bold">Gaji pokok + bonus</span>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-blue-50 text-blue-600 rounded-xl">
          <Clock class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Menunggu Verifikasi</p>
          <p class="text-2xl font-black text-slate-800">{{ pendingReportsCount }} Laporan</p>
        </div>
      </div>
    </div>

    <!-- Filter & Control Bar -->
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-4 flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex flex-wrap items-center gap-3 flex-1">
        <div class="relative w-full sm:w-[260px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari SPG, ritel, atau toko..."
            class="w-full pl-9 pr-4 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>

        <select
          v-model="selectedSpgId"
          class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white"
        >
          <option value="ALL">Semua SPG</option>
          <option v-for="spg in spgs" :key="spg.id" :value="spg.id">
            {{ spg.name }} ({{ spg.supermarket_name || 'MT' }})
          </option>
        </select>

        <select
          v-model="selectedStatus"
          class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white"
        >
          <option value="ALL">Semua Status</option>
          <option value="SUBMITTED">Menunggu Verifikasi</option>
          <option value="APPROVED">Disetujui (Approved)</option>
          <option value="REJECTED">Ditolak (Rejected)</option>
        </select>

        <div class="flex items-center gap-1.5 text-xs text-slate-500">
          <input
            v-model="startDate"
            type="date"
            @change="fetchReports"
            class="px-2.5 py-1.5 border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white text-xs"
          />
          <span>s/d</span>
          <input
            v-model="endDate"
            type="date"
            @change="fetchReports"
            class="px-2.5 py-1.5 border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white text-xs"
          />
        </div>
      </div>
    </div>

    <!-- Data Table -->
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full border-collapse text-left">
          <thead>
            <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
              <th class="py-4 px-5">Tanggal & Shift</th>
              <th class="py-4 px-5">SPG & Outlet Bertugas</th>
              <th class="py-4 px-5">Produk Terjual</th>
              <th class="py-4 px-5 text-right">Omzet & Target</th>
              <th class="py-4 px-5 text-center">Gaji & Bonus SPG</th>
              <th class="py-4 px-5 text-center">Bukti Struk</th>
              <th class="py-4 px-5 text-center">Status</th>
              <th class="py-4 px-5 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50 text-slate-700 text-sm">
            <tr v-for="rep in filteredReports" :key="rep.id" class="hover:bg-slate-50/30 transition-all">
              <!-- Tanggal & Shift -->
              <td class="py-4 px-5 text-xs">
                <div class="font-extrabold text-slate-900 flex items-center gap-1.5">
                  <Calendar class="w-3.5 h-3.5 text-slate-400" />
                  {{ formatDate(rep.report_date) }}
                </div>
                <div class="text-[11px] text-slate-500 mt-0.5 flex items-center gap-1">
                  <Clock class="w-3 h-3 text-slate-400" />
                  {{ rep.shift || 'Regular' }}
                </div>
              </td>

              <!-- SPG & Penugasan -->
              <td class="py-4 px-5">
                <div class="flex items-center gap-3">
                  <img
                    :src="rep.spg?.photo_url || 'https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?auto=format&fit=crop&q=80&w=100'"
                    class="w-9 h-9 rounded-xl object-cover border border-slate-100 flex-shrink-0"
                    alt="SPG"
                  />
                  <div>
                    <span class="font-extrabold text-slate-900 block text-xs">{{ rep.spg?.name || 'SPG' }}</span>
                    <span class="text-[11px] text-indigo-600 font-bold block">{{ rep.key_account?.name || 'Key Account' }}</span>
                    <span class="text-[11px] text-slate-400 block truncate max-w-[160px]">{{ rep.supermarket_name }}</span>
                  </div>
                </div>
              </td>

              <!-- Rincian Produk -->
              <td class="py-4 px-5 text-xs">
                <div v-if="rep.items && rep.items.length > 0" class="space-y-1 max-w-[200px]">
                  <div
                    v-for="(item, idx) in rep.items.slice(0, 2)"
                    :key="idx"
                    class="flex items-center justify-between text-[11px] bg-slate-50 px-2 py-1 rounded-lg border border-slate-100"
                  >
                    <span class="font-bold text-slate-800 truncate max-w-[120px]">{{ item.product_name }}</span>
                    <span class="text-indigo-600 font-black">x{{ item.quantity }}</span>
                  </div>
                  <button
                    v-if="rep.items.length > 2"
                    @click="openDetailDialog(rep)"
                    class="text-[10px] text-indigo-600 font-bold hover:underline block"
                  >
                    +{{ rep.items.length - 2 }} produk lainnya
                  </button>
                </div>
                <span v-else class="text-slate-400 text-xs italic">-</span>
              </td>

              <!-- Total Omzet & Target -->
              <td class="py-4 px-5 text-right">
                <span class="font-black text-slate-900 text-sm block">{{ formatCurrency(rep.total_amount) }}</span>
                <span class="text-[10px] font-bold text-slate-400 block">
                  Target: {{ formatCurrency(rep.target_amount || 0) }}
                </span>
                <span
                  v-if="rep.target_amount > 0"
                  class="inline-block mt-0.5 px-1.5 py-0.2 text-[9px] font-extrabold rounded"
                  :class="rep.is_target_achieved ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'"
                >
                  {{ rep.is_target_achieved ? '✓ Capai Target' : 'Belum Capai' }}
                </span>
              </td>

              <!-- Gaji & Bonus SPG -->
              <td class="py-4 px-5 text-center text-xs">
                <div class="font-black text-slate-900">
                  {{ formatCurrency(rep.total_spg_payout || rep.base_fee || 200000) }}
                </div>
                <div v-if="rep.is_target_achieved" class="text-[10px] font-extrabold text-amber-600 flex items-center justify-center gap-0.5 mt-0.5">
                  <Sparkles class="w-3 h-3" />
                  +{{ formatCurrency(rep.incentive_bonus || 50000) }} Bonus
                </div>
                <div v-else class="text-[10px] text-slate-400 font-medium">
                  Gaji dasar
                </div>
              </td>

              <!-- Bukti Struk -->
              <td class="py-4 px-5 text-center">
                <button
                  v-if="rep.receipt_photo_url"
                  @click="openPreview(rep.receipt_photo_url)"
                  class="inline-flex items-center gap-1 px-2.5 py-1 text-[11px] font-bold text-indigo-600 bg-indigo-50 hover:bg-indigo-100 border border-indigo-200/50 rounded-xl transition-all"
                >
                  <ImageIcon class="w-3.5 h-3.5" />
                  Foto
                </button>
                <span v-else class="text-slate-400 text-[11px] italic">-</span>
              </td>

              <!-- Status -->
              <td class="py-4 px-5 text-center">
                <span
                  class="px-2.5 py-1 text-[10px] font-extrabold uppercase rounded-full tracking-wider inline-flex items-center gap-1"
                  :class="{
                    'bg-amber-50 text-amber-600 border border-amber-200/60': rep.status === 'SUBMITTED',
                    'bg-emerald-50 text-emerald-600 border border-emerald-200/60': rep.status === 'APPROVED',
                    'bg-rose-50 text-rose-600 border border-rose-200/60': rep.status === 'REJECTED',
                  }"
                >
                  {{ rep.status === 'SUBMITTED' ? 'Menunggu' : rep.status === 'APPROVED' ? 'Disetujui' : 'Ditolak' }}
                </span>
              </td>

              <!-- Aksi -->
              <td class="py-4 px-5 text-center">
                <div class="flex items-center justify-center gap-1.5">
                  <button
                    @click="openDetailDialog(rep)"
                    class="p-1.5 hover:bg-slate-100 text-slate-500 rounded-lg transition-all"
                    title="Detail Laporan"
                  >
                    <Eye class="w-4 h-4" />
                  </button>

                  <button
                    v-if="rep.status === 'SUBMITTED'"
                    @click="openVerifyDialog(rep, 'APPROVED')"
                    class="p-1.5 hover:bg-emerald-50 text-emerald-600 rounded-lg transition-all"
                    title="Setujui Laporan"
                  >
                    <Check class="w-4 h-4" />
                  </button>

                  <button
                    v-if="rep.status === 'SUBMITTED'"
                    @click="openVerifyDialog(rep, 'REJECTED')"
                    class="p-1.5 hover:bg-rose-50 text-rose-600 rounded-lg transition-all"
                    title="Tolak Laporan"
                  >
                    <Ban class="w-4 h-4" />
                  </button>

                  <button
                    @click="openEditReportModal(rep)"
                    class="p-1.5 hover:bg-indigo-50 text-indigo-600 rounded-lg transition-all"
                    title="Edit Laporan"
                  >
                    <Edit2 class="w-4 h-4" />
                  </button>

                  <button
                    @click="deleteReport(rep.id)"
                    class="p-1.5 hover:bg-rose-50 text-rose-600 rounded-lg transition-all"
                    title="Hapus Laporan"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>

            <tr v-if="filteredReports.length === 0">
              <td colspan="8" class="py-10 text-center text-slate-400">
                <Receipt class="w-8 h-8 mx-auto mb-2 text-slate-300" />
                <p class="font-bold text-slate-500">Belum ada laporan penjualan SPG ditemukan.</p>
                <p class="text-xs text-slate-400 mt-0.5">Gunakan tombol "Input Manual" atau "Import Excel".</p>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MODAL 1: Form Input / Edit Laporan Penjualan SPG -->
    <div
      v-if="showFormModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white rounded-[24px] w-full max-w-3xl shadow-2xl overflow-hidden p-6 relative my-8 max-h-[90vh] flex flex-col">
        <div class="flex items-center justify-between pb-4 border-b border-slate-100">
          <div>
            <h3 class="text-lg font-black text-slate-900">
              {{ isEditing ? 'Edit Laporan Penjualan SPG' : 'Input Laporan Penjualan SPG' }}
            </h3>
            <p class="text-xs text-slate-400 mt-0.5">Catat hasil penjualan produk dan sistem otomatis menghitung bonus target</p>
          </div>
          <button @click="showFormModal = false" class="p-1 text-slate-400 hover:text-slate-600 rounded-lg">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="space-y-4 text-xs font-semibold text-slate-650 overflow-y-auto pr-1 py-4 flex-1">
          <!-- Opsi Pilih Jadwal SPG -->
          <div v-if="schedules.length > 0" class="bg-indigo-50/60 p-3 rounded-2xl border border-indigo-100">
            <label class="block text-[10px] uppercase font-bold text-indigo-700 mb-1">Pilih dari Jadwal Penugasan SPG (Opsional)</label>
            <select
              v-model="formReport.schedule_id"
              @change="onScheduleSelected"
              class="w-full px-3 py-2 border border-indigo-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white text-xs font-bold text-slate-800"
            >
              <option value="">-- Tanpa Jadwal Tertentu / Pilih Manual --</option>
              <option v-for="sc in schedules" :key="sc.id" :value="sc.id">
                {{ sc.spg?.name }} @ {{ sc.supermarket_name }} ({{ sc.day_pattern === 'WEEKEND_FRI_SUN' ? 'Jum-Sab-Min' : 'Reguler' }}) - Target {{ formatCurrency(sc.target_sales / (sc.working_days_count || 1)) }}/hari
              </option>
            </select>
          </div>

          <!-- Info SPG & Outlet -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3 bg-slate-50/70 p-4 rounded-2xl border border-slate-100">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">SPG Bertugas</label>
              <select
                v-model="formReport.spg_id"
                @change="onSpgSelected"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white text-xs"
              >
                <option v-for="spg in spgs" :key="spg.id" :value="spg.id">
                  {{ spg.name }} ({{ spg.supermarket_name || 'MT' }})
                </option>
              </select>
            </div>

            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Ritel MT (Key Account)</label>
              <select
                v-model="formReport.key_account_id"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white text-xs"
              >
                <option v-for="ka in keyAccounts" :key="ka.id" :value="ka.id">
                  {{ ka.name }}
                </option>
              </select>
            </div>

            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Nama Outlet / Supermarket</label>
              <input
                v-model="formReport.supermarket_name"
                type="text"
                placeholder="Contoh: Hypermart Karawaci"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white text-xs"
              />
            </div>
          </div>

          <!-- Target & Gaji Dasar -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Tanggal Penjualan</label>
              <input
                v-model="formReport.report_date"
                type="date"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white text-xs"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Target Sales Shift Ini (IDR)</label>
              <input
                v-model.number="formReport.target_amount"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white text-xs font-bold text-emerald-600"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Gaji Pokok / Shift (IDR)</label>
              <input
                v-model.number="formReport.base_fee"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white text-xs font-bold"
              />
            </div>
          </div>

          <!-- Product Item Builder -->
          <div class="border border-slate-100 rounded-2xl p-4 bg-white space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-black text-slate-800 uppercase tracking-wider">Rincian Produk Terjual</span>
              <Button
                type="button"
                @click="addItemRow"
                size="sm"
                variant="outline"
                class="text-[11px] font-bold border-indigo-200 text-indigo-600 hover:bg-indigo-50 h-7"
              >
                <Plus class="w-3.5 h-3.5 mr-1" />
                Tambah Baris Produk
              </Button>
            </div>

            <div class="space-y-2">
              <div
                v-for="(item, idx) in formReport.items"
                :key="idx"
                class="grid grid-cols-12 gap-2 items-center bg-slate-50/70 p-2.5 rounded-xl border border-slate-100"
              >
                <div class="col-span-5">
                  <label class="block text-[9px] uppercase font-bold text-slate-400 mb-0.5">Produk</label>
                  <div class="flex flex-col gap-1">
                    <select
                      @change="(e: any) => onProductSelectChange(item, e.target.value)"
                      class="w-full px-2 py-1.5 border border-slate-200 rounded-lg text-xs bg-white focus:outline-none"
                    >
                      <option value="">-- Pilih dari Master Produk --</option>
                      <option v-for="p in products" :key="p.id" :value="p.id">
                        {{ p.name }} ({{ formatCurrency(p.price || p.selling_price || 0) }})
                      </option>
                    </select>
                    <input
                      v-model="item.product_name"
                      type="text"
                      placeholder="Atau ketik nama produk..."
                      class="w-full px-2 py-1 border border-slate-200 rounded-lg text-[11px] bg-white focus:outline-none"
                    />
                  </div>
                </div>

                <div class="col-span-3">
                  <label class="block text-[9px] uppercase font-bold text-slate-400 mb-0.5">Harga Satuan</label>
                  <input
                    v-model.number="item.unit_price"
                    type="number"
                    @input="updateSubtotal(item)"
                    class="w-full px-2 py-1.5 border border-slate-200 rounded-lg text-xs bg-white focus:outline-none"
                  />
                </div>

                <div class="col-span-2">
                  <label class="block text-[9px] uppercase font-bold text-slate-400 mb-0.5">Qty</label>
                  <input
                    v-model.number="item.quantity"
                    type="number"
                    min="1"
                    @input="updateSubtotal(item)"
                    class="w-full px-2 py-1.5 border border-slate-200 rounded-lg text-xs bg-white focus:outline-none text-center font-bold"
                  />
                </div>

                <div class="col-span-2 flex items-center justify-between pl-1">
                  <div>
                    <label class="block text-[9px] uppercase font-bold text-slate-400 mb-0.5">Subtotal</label>
                    <span class="text-[11px] font-black text-slate-800">{{ formatCurrency(item.subtotal) }}</span>
                  </div>
                  <button
                    type="button"
                    @click="removeItemRow(idx)"
                    class="p-1 hover:bg-rose-50 text-slate-400 hover:text-rose-600 rounded-lg transition-all"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Grand Total & Dynamic Incentive Preview -->
            <div class="pt-3 border-t border-slate-100 flex flex-col md:flex-row md:items-center justify-between gap-3 bg-slate-50/50 p-3 rounded-xl">
              <div>
                <span class="text-xs font-bold text-slate-500">Total Omzet: </span>
                <span class="text-base font-black text-slate-900">{{ formatCurrency(formTotalAmount) }}</span>
                <span class="text-xs text-slate-400 ml-2">({{ formTotalQty }} Pcs)</span>
              </div>

              <!-- Incentive Badge -->
              <div class="flex items-center gap-2">
                <span
                  class="px-2.5 py-1 rounded-xl text-xs font-black flex items-center gap-1.5"
                  :class="isTargetAchievedInForm ? 'bg-emerald-100 text-emerald-800 border border-emerald-200' : 'bg-amber-50 text-amber-800 border border-amber-200'"
                >
                  <Sparkles class="w-4 h-4 text-amber-500" />
                  {{ isTargetAchievedInForm ? 'Target Tercapai (+Rp 50.000 Bonus)' : 'Kurang dari Target (Gaji Standar)' }}
                </span>
                <span class="text-xs font-bold text-slate-700">
                  Estimasi Payout: <b class="text-indigo-600 font-black">{{ formatCurrency(estimatedPayoutInForm) }}</b>
                </span>
              </div>
            </div>
          </div>

          <!-- Struk & Bukti Penjualan -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Foto Struk Kasir / Bukti Penjualan</label>
              <input
                type="file"
                accept="image/*"
                @change="uploadReceiptPhoto"
                class="w-full text-xs text-slate-500 file:mr-2 file:py-1 file:px-3 file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
              />
              <div v-if="formReport.receipt_photo_url" class="mt-2 flex items-center gap-2">
                <img :src="formReport.receipt_photo_url" class="w-12 h-12 object-cover rounded-lg border border-slate-200" />
                <span class="text-[10px] text-emerald-600 font-bold">Struk terunggah</span>
              </div>
            </div>

            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Catatan Tambahan</label>
              <textarea
                v-model="formReport.notes"
                rows="2"
                placeholder="Catatan push sales, promo demo toples..."
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none text-xs"
              ></textarea>
            </div>
          </div>
        </div>

        <div class="pt-4 border-t border-slate-100 flex items-center justify-end gap-3">
          <Button
            @click="showFormModal = false"
            variant="outline"
            class="px-4 py-2 border border-slate-200 rounded-xl text-xs text-slate-500 font-bold"
          >
            Batal
          </Button>
          <Button
            @click="saveReport"
            class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold"
          >
            Simpan Laporan Penjualan
          </Button>
        </div>
      </div>
    </div>

    <!-- MODAL 2: Excel Import Preview Modal -->
    <div
      v-if="showImportModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white rounded-[24px] w-full max-w-4xl shadow-2xl overflow-hidden p-6 relative my-8 max-h-[90vh] flex flex-col">
        <div class="flex items-center justify-between pb-4 border-b border-slate-100">
          <div>
            <h3 class="text-lg font-black text-slate-900">Konfirmasi Import Excel Penjualan SPG</h3>
            <p class="text-xs text-slate-400">{{ parsedExcelData.length }} sesi laporan siap disimpan</p>
          </div>
          <button @click="showImportModal = false" class="p-1 text-slate-400 hover:text-slate-600 rounded-lg">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="overflow-y-auto py-4 flex-1 space-y-3">
          <div v-for="(item, idx) in parsedExcelData" :key="idx" class="p-3 bg-slate-50 rounded-2xl border border-slate-200 text-xs">
            <div class="flex justify-between items-start mb-2">
              <div>
                <span class="font-extrabold text-slate-900 text-sm">{{ item.spg_name }}</span>
                <span class="text-slate-400 ml-2 font-bold">{{ item.supermarket_name }} ({{ formatDate(item.report_date) }})</span>
              </div>
              <div class="text-right">
                <span class="text-xs font-black text-emerald-600">Target: {{ formatCurrency(item.target_amount) }}</span>
              </div>
            </div>

            <div class="space-y-1">
              <div v-for="(p, pIdx) in item.items" :key="pIdx" class="flex justify-between text-[11px] text-slate-600">
                <span>• {{ p.product_name }} (x{{ p.quantity }})</span>
                <span class="font-bold">{{ formatCurrency(p.subtotal) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="pt-4 border-t border-slate-100 flex items-center justify-end gap-3">
          <Button
            @click="showImportModal = false"
            variant="outline"
            class="px-4 py-2 border border-slate-200 rounded-xl text-xs text-slate-500 font-bold"
          >
            Batal
          </Button>
          <Button
            @click="submitParsedExcel"
            :disabled="isImporting"
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold"
          >
            {{ isImporting ? 'Menyimpan...' : 'Konfirmasi Simpan Semua' }}
          </Button>
        </div>
      </div>
    </div>

    <!-- MODAL 3: Verification Dialog Modal -->
    <div
      v-if="showVerifyModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
    >
      <div class="bg-white rounded-[24px] w-full max-w-md shadow-2xl p-6 relative">
        <h3 class="text-lg font-black text-slate-900 mb-2">
          {{ verifyStatus === 'APPROVED' ? 'Setujui Laporan Penjualan?' : 'Tolak Laporan Penjualan?' }}
        </h3>
        <p class="text-xs text-slate-500 mb-4">
          {{
            verifyStatus === 'APPROVED'
              ? 'Setelah disetujui, omzet dan honor SPG (termasuk bonus jika capai target) akan tercatat di sistem penggajian.'
              : 'Laporan akan ditandai ditolak.'
          }}
        </p>

        <div class="space-y-3 mb-5">
          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Catatan Supervisor</label>
            <textarea
              v-model="verifyNotes"
              rows="3"
              placeholder="Tambahkan catatan verifikasi..."
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none text-xs"
            ></textarea>
          </div>
        </div>

        <div class="flex items-center justify-end gap-2">
          <Button
            @click="showVerifyModal = false"
            variant="outline"
            class="px-4 py-2 text-xs font-bold text-slate-500 rounded-xl"
          >
            Batal
          </Button>
          <Button
            @click="submitVerification"
            :class="
              verifyStatus === 'APPROVED'
                ? 'bg-emerald-600 hover:bg-emerald-700 text-white'
                : 'bg-rose-600 hover:bg-rose-700 text-white'
            "
            class="px-4 py-2 text-xs font-bold rounded-xl"
          >
            {{ verifyStatus === 'APPROVED' ? 'Konfirmasi Setujui' : 'Konfirmasi Tolak' }}
          </Button>
        </div>
      </div>
    </div>

    <!-- MODAL 4: Detail View Modal -->
    <div
      v-if="showDetailModal && detailedReport"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
    >
      <div class="bg-white rounded-[24px] w-full max-w-2xl shadow-2xl p-6 relative max-h-[90vh] flex flex-col">
        <div class="flex items-center justify-between pb-4 border-b border-slate-100">
          <div>
            <h3 class="text-lg font-black text-slate-900">Detail Laporan Penjualan SPG</h3>
            <p class="text-xs text-slate-400">{{ formatDate(detailedReport.report_date) }} • {{ detailedReport.shift || 'Shift Regular' }}</p>
          </div>
          <button @click="showDetailModal = false" class="p-1 text-slate-400 hover:text-slate-600 rounded-lg">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="space-y-4 text-xs overflow-y-auto py-4 flex-1">
          <!-- SPG Info Header with Payout Breakdown -->
          <div class="flex items-center justify-between bg-slate-50 p-4 rounded-2xl border border-slate-100">
            <div class="flex items-center gap-3">
              <img
                :src="detailedReport.spg?.photo_url || 'https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?auto=format&fit=crop&q=80&w=100'"
                class="w-12 h-12 rounded-xl object-cover border border-slate-200"
              />
              <div>
                <p class="font-black text-slate-900 text-sm">{{ detailedReport.spg?.name }}</p>
                <p class="text-indigo-600 font-bold text-xs">{{ detailedReport.key_account?.name }} - {{ detailedReport.supermarket_name }}</p>
                <p class="text-slate-400 text-[11px]">{{ detailedReport.spg?.phone_number || '-' }}</p>
              </div>
            </div>

            <div class="text-right">
              <span class="text-[10px] uppercase font-bold text-slate-400 block">Total Omzet Penjualan</span>
              <span class="font-black text-slate-900 text-lg text-emerald-600">{{ formatCurrency(detailedReport.total_amount) }}</span>
              <span class="text-[10px] text-slate-500 font-bold block">Target: {{ formatCurrency(detailedReport.target_amount || 0) }}</span>
            </div>
          </div>

          <!-- Payout Breakdown Box -->
          <div class="p-3 bg-indigo-50/50 rounded-2xl border border-indigo-100 flex items-center justify-between">
            <div>
              <span class="text-xs font-bold text-slate-700 block">Rincian Pembayaran SPG (Honor Shift)</span>
              <span class="text-[11px] text-slate-500">Gaji Dasar: {{ formatCurrency(detailedReport.base_fee || 200000) }}</span>
              <span v-if="detailedReport.is_target_achieved" class="text-[11px] text-emerald-600 font-black block">
                + Bonus Capai Target: {{ formatCurrency(detailedReport.incentive_bonus || 50000) }}
              </span>
            </div>
            <div class="text-right">
              <span class="text-[10px] uppercase font-bold text-slate-400 block">Total Honor</span>
              <span class="text-base font-black text-indigo-700">{{ formatCurrency(detailedReport.total_spg_payout || 200000) }}</span>
            </div>
          </div>

          <!-- Product Table -->
          <div>
            <h4 class="font-bold text-slate-800 text-xs mb-2 uppercase tracking-wider">Item Produk Terjual</h4>
            <div class="border border-slate-100 rounded-2xl overflow-hidden">
              <table class="w-full text-left text-xs">
                <thead class="bg-slate-50 text-[10px] font-black text-slate-400 uppercase">
                  <tr>
                    <th class="py-2.5 px-4">Nama Produk</th>
                    <th class="py-2.5 px-4 text-center">SKU</th>
                    <th class="py-2.5 px-4 text-right">Harga Satuan</th>
                    <th class="py-2.5 px-4 text-center">Qty</th>
                    <th class="py-2.5 px-4 text-right">Subtotal</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-50 text-slate-700">
                  <tr v-for="item in detailedReport.items" :key="item.id">
                    <td class="py-2.5 px-4 font-bold text-slate-800">{{ item.product_name }}</td>
                    <td class="py-2.5 px-4 text-center text-slate-400">{{ item.sku_code || '-' }}</td>
                    <td class="py-2.5 px-4 text-right">{{ formatCurrency(item.unit_price) }}</td>
                    <td class="py-2.5 px-4 text-center font-black text-indigo-600">{{ item.quantity }}</td>
                    <td class="py-2.5 px-4 text-right font-black">{{ formatCurrency(item.subtotal) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Photo & Notes -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div v-if="detailedReport.receipt_photo_url">
              <h4 class="font-bold text-slate-800 text-xs mb-2 uppercase tracking-wider">Foto Struk Kasir</h4>
              <img
                :src="detailedReport.receipt_photo_url"
                @click="openPreview(detailedReport.receipt_photo_url)"
                class="w-full max-h-48 object-cover rounded-xl border border-slate-200 cursor-pointer hover:opacity-90"
              />
            </div>
            <div>
              <h4 class="font-bold text-slate-800 text-xs mb-2 uppercase tracking-wider">Catatan</h4>
              <div class="p-3 bg-slate-50 rounded-xl border border-slate-100 text-slate-600 min-h-[80px]">
                {{ detailedReport.notes || 'Tidak ada catatan khusus.' }}
              </div>
            </div>
          </div>
        </div>

        <div class="pt-4 border-t border-slate-100 flex items-center justify-between">
          <div>
            <span
              class="px-2.5 py-1 text-[10px] font-extrabold uppercase rounded-full tracking-wider"
              :class="{
                'bg-amber-50 text-amber-600': detailedReport.status === 'SUBMITTED',
                'bg-emerald-50 text-emerald-600': detailedReport.status === 'APPROVED',
                'bg-rose-50 text-rose-600': detailedReport.status === 'REJECTED',
              }"
            >
              Status: {{ detailedReport.status }}
            </span>
          </div>
          <Button @click="showDetailModal = false" class="px-4 py-2 bg-slate-800 hover:bg-slate-900 text-white rounded-xl text-xs font-bold">
            Tutup
          </Button>
        </div>
      </div>
    </div>

    <!-- MODAL 5: Image Preview Modal -->
    <div
      v-if="previewImageUrl"
      class="fixed inset-0 z-[100] flex items-center justify-center bg-slate-900/90 backdrop-blur-sm"
      @click.self="closePreview"
    >
      <div class="relative w-full h-full flex flex-col items-center justify-center p-4" @click.self="closePreview">
        <div class="absolute top-6 right-6 flex items-center gap-4 bg-white/10 backdrop-blur-md px-4 py-2 rounded-full border border-white/20 shadow-2xl z-10">
          <button @click="zoomOut" class="text-white hover:text-indigo-300 transition-colors" title="Zoom Out">
            <ZoomOut class="w-6 h-6" />
          </button>
          <span class="text-white text-xs font-bold w-12 text-center">{{ Math.round(zoomLevel * 100) }}%</span>
          <button @click="zoomIn" class="text-white hover:text-indigo-300 transition-colors" title="Zoom In">
            <ZoomIn class="w-6 h-6" />
          </button>
          <div class="w-px h-4 bg-white/30 mx-2"></div>
          <button @click="closePreview" class="text-white hover:text-rose-400 transition-colors" title="Tutup">
            <X class="w-6 h-6" />
          </button>
        </div>

        <div 
          class="overflow-auto max-w-full max-h-full transition-transform duration-200 ease-out flex items-center justify-center pointer-events-none"
          :style="{ transform: `scale(${zoomLevel})` }"
        >
          <img 
            :src="previewImageUrl" 
            class="max-w-[90vw] max-h-[90vh] object-contain rounded-lg shadow-2xl ring-1 ring-white/20 pointer-events-auto"
            alt="Preview"
          />
        </div>
      </div>
    </div>
  </div>
</template>
