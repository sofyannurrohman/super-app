<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  CheckCircle2,
  Clock,
  AlertTriangle,
  X,
  RefreshCw,
  Search,
  ChevronLeft,
  ChevronRight,
  Camera,
  Image as ImageIcon,
  MapPin,
  Upload,
  LayoutGrid,
  List,
  Store as StoreIcon,
  User,
  UserPlus,
  UserCheck,
  UserX,
  Users,
  ExternalLink,
  ZoomIn,
  Eye,
  Calendar,
  Filter,
  Download,
  Sparkles,
  FileSpreadsheet,
  Building2,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import ExcelJS from 'exceljs'
import { saveAs } from 'file-saver'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import { resolvePhotoUrl } from '@/utils/imageUrl'
import { formatBannerDimension } from '@/utils/bannerSize'
import StoreAddressHoverCard from '@/components/StoreAddressHoverCard.vue'

const masterStore = useMasterDataStore()
const authStore = useAuthStore()

const isAuthorizedBrandingRole = computed(() => {
  const role = authStore.user?.role
  const allowedRoles = [
    'admin_designer',
    'validator_desain',
    'validator_toko',
    'admin_purchasing',
    'operator',
    'superadmin',
    'admin',
    'hr_admin',
    'HR',
    'manager',
  ]
  return !!role && allowedRoles.includes(role)
})
const canManageInstallation = computed(() => {
  return isAuthorizedBrandingRole.value
})

const isLoading = ref(false)
const isOnline = ref(true)

// Tab state
const activeTab = ref<'WAITING' | 'HISTORY'>('WAITING')

// View mode for HISTORY tab: 'GALLERY' or 'TABLE'
const historyViewMode = ref<'GALLERY' | 'TABLE'>('GALLERY')

// Global search & branch filter
const search = ref('')
const selectedBranchFilter = computed<string>({
  get: () => masterStore.selectedBrandingBranchId || 'ALL_BRANCHES',
  set: (val: string) => {
    masterStore.setSelectedBrandingBranchId(val)
  },
})
let searchTimeout: any = null

const selectedBranchName = computed(() => {
  return masterStore.selectedBrandingBranchName
})

// History Filter state
const selectedOperatorFilter = ref('all')
const datePreset = ref<'all' | 'today' | '7d' | '30d' | 'this_month' | 'custom'>('all')
const startDate = ref('')
const endDate = ref('')

// Waiting state
const waitingTasks = ref<any[]>([])
const page = ref(1)
const limit = ref(10)
const total = ref(0)

// History state
const historyTasks = ref<any[]>([])
const historyPage = ref(1)
const historyLimit = ref(12)
const historyTotal = ref(0)

// Operators state for Assignment & Filter
const operators = ref<any[]>([])
const isLoadingOperators = ref(false)
const showAssignModal = ref(false)
const selectedTaskForAssign = ref<any>(null)
const selectedOperatorId = ref('')
const operatorSearchQuery = ref('')
const isAssigning = ref(false)

// Upload Modal state
const showUploadModal = ref(false)
const selectedTask = ref<any>(null)
const selectedFile = ref<File | null>(null)
const previewImage = ref<string | null>(null)
const latitude = ref<number | null>(null)
const longitude = ref<number | null>(null)
const installationDateTime = ref('')
const isUploading = ref(false)

const formatToDateTimeLocal = (d: Date) => {
  const pad = (n: number) => String(n).padStart(2, '0')
  const year = d.getFullYear()
  const month = pad(d.getMonth() + 1)
  const day = pad(d.getDate())
  const hours = pad(d.getHours())
  const minutes = pad(d.getMinutes())
  return `${year}-${month}-${day}T${hours}:${minutes}`
}

// Preview Modal state
const showPreviewModal = ref(false)
const previewUrl = ref('')
const previewTitle = ref('')

const fetchOperators = async () => {
  isLoadingOperators.value = true
  try {
    const res = await apiClient.get('/branding/operators')
    operators.value = res.data?.data || []
  } catch (e: any) {
    toast.error('Gagal memuat daftar operator: ' + (e.response?.data?.error || e.message))
  } finally {
    isLoadingOperators.value = false
  }
}

const filteredOperators = computed(() => {
  if (!operatorSearchQuery.value) return operators.value
  const q = operatorSearchQuery.value.toLowerCase()
  return operators.value.filter(
    (op: any) =>
      op.name?.toLowerCase().includes(q) ||
      op.email?.toLowerCase().includes(q) ||
      op.phone?.toLowerCase().includes(q) ||
      op.area_code?.toLowerCase().includes(q)
  )
})

const fetchWaitingTasks = async (silent = false) => {
  if (!silent) isLoading.value = true
  try {
    const branchId = selectedBranchFilter.value
    const filter = branchId && branchId !== 'ALL_BRANCHES' ? `&branch_id=${branchId}` : ''
    const res = await apiClient.get(
      `/branding/banners?status=installation&page=${page.value}&limit=${limit.value}&search=${search.value}${filter}`,
    )
    waitingTasks.value = (res.data?.data || []).map((b: any) => ({
      id: b.id,
      shop_name: b.shop_name || 'Toko Unknown',
      shop_address: b.shop_address || '',
      shop_phone: b.shop_phone || '',
      province: b.province || '',
      regency: b.regency || '',
      district: b.district || '',
      village: b.village || '',
      latitude: b.latitude || 0,
      longitude: b.longitude || 0,
      shop_photo: resolvePhotoUrl(b.shop_photo),
      sales_name: b.sales_name || '-',
      sales_branch: b.sales_branch || '',
      operator_id: b.operator_id || null,
      operator_name: b.operator_name || '',
      operator_phone: b.operator_phone || '',
      status: b.status,
      type: `${b.type || 'Spanduk'} ${formatBannerDimension(b.size_length, b.size_width, 'dual')}`,
      created_at: formatDate(b.validated_at || b.updated_at),
    }))
    total.value = res.data?.total || 0
    isOnline.value = true
  } catch (e: any) {
    isOnline.value = false
    toast.error('Gagal memuat data dari server.')
  } finally {
    if (!silent) isLoading.value = false
  }
}

// Today Installed state (Spotlight on top)
const todayTasks = ref<any[]>([])
const isLoadingToday = ref(false)
const isExporting = ref(false)

const currentFormattedTodayDate = computed(() => {
  return new Date().toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
})

const formatDateToISO = (d: Date) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const fetchTodayTasks = async () => {
  isLoadingToday.value = true
  try {
    const todayStr = formatDateToISO(new Date())
    const branchId = selectedBranchFilter.value
    let queryParams = `status=installed&start_date=${todayStr}&end_date=${todayStr}&limit=100&sort_by=installed_at&sort_dir=DESC`
    if (branchId && branchId !== 'ALL_BRANCHES') {
      queryParams += `&branch_id=${branchId}`
    }

    const res = await apiClient.get(`/branding/banners?${queryParams}`)
    todayTasks.value = (res.data?.data || []).map((b: any) => ({
      id: b.id,
      shop_name: b.shop_name || 'Toko Unknown',
      shop_address: b.shop_address || '',
      shop_phone: b.shop_phone || '',
      province: b.province || '',
      regency: b.regency || '',
      district: b.district || '',
      village: b.village || '',
      latitude: b.latitude || 0,
      longitude: b.longitude || 0,
      shop_photo: resolvePhotoUrl(b.shop_photo),
      sales_name: b.sales_name || '-',
      sales_branch: b.sales_branch || '',
      operator_id: b.operator_id || null,
      operator_name: b.operator_name || '',
      operator_phone: b.operator_phone || '',
      type: `${b.type || 'Spanduk'} ${formatBannerDimension(b.size_length, b.size_width, 'dual')}`,
      status: b.status,
      photo_installed: resolvePhotoUrl(b.photo_installed),
      installed_at: formatDate(b.installed_at || b.updated_at),
    }))
  } catch (e: any) {
    console.error('Gagal memuat data terpasang hari ini:', e)
  } finally {
    isLoadingToday.value = false
  }
}

const fetchHistoryTasks = async (silent = false) => {
  if (!silent) isLoading.value = true
  try {
    const branchId = selectedBranchFilter.value
    let queryParams = `status=installed&page=${historyPage.value}&limit=${historyLimit.value}&search=${encodeURIComponent(search.value)}&sort_by=installed_at&sort_dir=DESC`
    if (branchId && branchId !== 'ALL_BRANCHES') {
      queryParams += `&branch_id=${branchId}`
    }
    if (selectedOperatorFilter.value && selectedOperatorFilter.value !== 'all') {
      queryParams += `&operator_id=${encodeURIComponent(selectedOperatorFilter.value)}`
    }
    if (startDate.value) {
      queryParams += `&start_date=${startDate.value}`
    }
    if (endDate.value) {
      queryParams += `&end_date=${endDate.value}`
    }

    const res = await apiClient.get(`/branding/banners?${queryParams}`)
    historyTasks.value = (res.data?.data || []).map((b: any) => ({
      id: b.id,
      shop_name: b.shop_name || 'Toko Unknown',
      shop_address: b.shop_address || '',
      shop_phone: b.shop_phone || '',
      province: b.province || '',
      regency: b.regency || '',
      district: b.district || '',
      village: b.village || '',
      latitude: b.latitude || 0,
      longitude: b.longitude || 0,
      shop_photo: resolvePhotoUrl(b.shop_photo),
      sales_name: b.sales_name || '-',
      sales_branch: b.sales_branch || '',
      operator_id: b.operator_id || null,
      operator_name: b.operator_name || '',
      operator_phone: b.operator_phone || '',
      type: `${b.type || 'Spanduk'} ${formatBannerDimension(b.size_length, b.size_width, 'dual')}`,
      status: b.status,
      photo_installed: resolvePhotoUrl(b.photo_installed),
      installed_at: formatDate(b.installed_at || b.updated_at),
    }))
    historyTotal.value = res.data?.total || 0
    isOnline.value = true
  } catch (e: any) {
    isOnline.value = false
    toast.error('Gagal memuat riwayat dari server.')
  } finally {
    if (!silent) isLoading.value = false
  }
}

const setDatePreset = (preset: 'all' | 'today' | '7d' | '30d' | 'this_month' | 'custom') => {
  datePreset.value = preset
  const now = new Date()

  if (preset === 'all') {
    startDate.value = ''
    endDate.value = ''
  } else if (preset === 'today') {
    const todayStr = formatDateToISO(now)
    startDate.value = todayStr
    endDate.value = todayStr
  } else if (preset === '7d') {
    const past7 = new Date()
    past7.setDate(now.getDate() - 6)
    startDate.value = formatDateToISO(past7)
    endDate.value = formatDateToISO(now)
  } else if (preset === '30d') {
    const past30 = new Date()
    past30.setDate(now.getDate() - 29)
    startDate.value = formatDateToISO(past30)
    endDate.value = formatDateToISO(now)
  } else if (preset === 'this_month') {
    const firstDay = new Date(now.getFullYear(), now.getMonth(), 1)
    startDate.value = formatDateToISO(firstDay)
    endDate.value = formatDateToISO(now)
  }
  historyPage.value = 1
  fetchHistoryTasks()
}

const onCustomDateChange = () => {
  datePreset.value = 'custom'
  historyPage.value = 1
  fetchHistoryTasks()
}

const onOperatorFilterChange = () => {
  historyPage.value = 1
  fetchHistoryTasks()
}

const resetHistoryFilters = () => {
  selectedOperatorFilter.value = 'all'
  search.value = ''
  setDatePreset('all')
}

const hasActiveHistoryFilters = computed(() => {
  return (
    selectedOperatorFilter.value !== 'all' ||
    Boolean(startDate.value) ||
    Boolean(endDate.value) ||
    Boolean(search.value)
  )
})

const fetchTasks = (silent = false) => {
  if (activeTab.value === 'WAITING') {
    fetchWaitingTasks(silent)
  } else {
    fetchTodayTasks()
    fetchHistoryTasks(silent)
  }
}

watch(() => masterStore.selectedBrandingBranchId, () => {
  page.value = 1
  historyPage.value = 1
  fetchTasks()
})

// ── Export to Excel ────────────────────────────────────────────────────────
const exportHistoryToExcel = async () => {
  if (isExporting.value) return
  isExporting.value = true
  toast.loading('Menyiapkan file Excel riwayat terpasang...', { id: 'export-history-excel' })
  try {
    const branchId = selectedBranchFilter.value
    let queryParams = `status=installed&limit=5000&sort_by=installed_at&sort_dir=DESC`
    if (branchId && branchId !== 'ALL_BRANCHES') {
      queryParams += `&branch_id=${branchId}`
    }
    if (search.value) {
      queryParams += `&search=${encodeURIComponent(search.value)}`
    }
    if (selectedOperatorFilter.value && selectedOperatorFilter.value !== 'all') {
      queryParams += `&operator_id=${encodeURIComponent(selectedOperatorFilter.value)}`
    }
    if (startDate.value) {
      queryParams += `&start_date=${startDate.value}`
    }
    if (endDate.value) {
      queryParams += `&end_date=${endDate.value}`
    }

    const res = await apiClient.get(`/branding/banners?${queryParams}`)
    const rawData = res.data?.data || []

    if (rawData.length === 0) {
      toast.error('Tidak ada data terpasang untuk diexport.', { id: 'export-history-excel' })
      return
    }

    const branchName = selectedBranchName.value || 'Semua Cabang'
    const workbook = new ExcelJS.Workbook()
    workbook.creator = 'HRIS Karsa HR SuperApp'
    workbook.created = new Date()

    const worksheet = workbook.addWorksheet('Bukti Tayang Terpasang', {
      views: [{ state: 'frozen', ySplit: 7 }],
      pageSetup: { orientation: 'landscape', fitToPage: true, fitToWidth: 1, fitToHeight: 0 },
    })

    // 1. Header Title (Row 1-2)
    worksheet.mergeCells('A1:R1')
    const titleCell = worksheet.getCell('A1')
    titleCell.value = 'LAPORAN BUKTI TAYANG & PEMASANGAN SPANDUK'
    titleCell.font = { bold: true, size: 14, color: { argb: 'FFFFFFFF' } }
    titleCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF065F46' } } // Emerald 800
    titleCell.alignment = { horizontal: 'center', vertical: 'middle' }
    worksheet.getRow(1).height = 34

    worksheet.mergeCells('A2:R2')
    const subTitleCell = worksheet.getCell('A2')
    subTitleCell.value = 'Perusahaan Demo — DIVISI BRANDING & MARKETING'
    subTitleCell.font = { bold: true, size: 9, color: { argb: 'FFD1FAE5' } } // Emerald 100
    subTitleCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF047857' } } // Emerald 700
    subTitleCell.alignment = { horizontal: 'center', vertical: 'middle' }
    worksheet.getRow(2).height = 20

    // 2. Metadata / Filter Bar (Row 3)
    let periodeLabel = 'Semua Periode'
    if (startDate.value && endDate.value) {
      periodeLabel = `${startDate.value} s/d ${endDate.value}`
    } else if (startDate.value) {
      periodeLabel = `Sejak ${startDate.value}`
    } else if (endDate.value) {
      periodeLabel = `Sampai ${endDate.value}`
    } else if (datePreset.value === 'today') {
      periodeLabel = `Hari Ini (${formatDateToISO(new Date())})`
    } else if (datePreset.value === '7d') {
      periodeLabel = '7 Hari Terakhir'
    } else if (datePreset.value === '30d') {
      periodeLabel = '30 Hari Terakhir'
    } else if (datePreset.value === 'this_month') {
      periodeLabel = 'Bulan Ini'
    }

    const opFilterObj = operators.value.find((o) => String(o.id) === String(selectedOperatorFilter.value))
    const operatorFilterLabel = selectedOperatorFilter.value === 'all' ? 'Semua Operator' : (opFilterObj?.name || selectedOperatorFilter.value)
    const searchFilterLabel = search.value ? `"${search.value}"` : 'Semua'

    worksheet.mergeCells('A3:R3')
    const metaCell = worksheet.getCell('A3')
    metaCell.value = `Cabang: ${branchName}   |   Periode: ${periodeLabel}   |   Operator: ${operatorFilterLabel}   |   Search: ${searchFilterLabel}   |   Total Terpasang: ${rawData.length} Item   |   Diexport: ${new Date().toLocaleString('id-ID')}`
    metaCell.font = { italic: true, size: 9, color: { argb: 'FF334155' } }
    metaCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF1F5F9' } }
    metaCell.alignment = { horizontal: 'center', vertical: 'middle' }
    worksheet.getRow(3).height = 22

    worksheet.addRow({}) // Row 4 spacer
    worksheet.getRow(4).height = 8

    // 3. Summary KPI Block (Row 5)
    const totalLuasM2 = rawData.reduce((acc: number, b: any) => {
      const l = Number(b.size_length || 0)
      const w = Number(b.size_width || 0)
      const qty = Number(b.qty || 1)
      const lMeter = l > 20 ? l / 100 : l
      const wMeter = w > 20 ? w / 100 : w
      return acc + (lMeter * wMeter * qty)
    }, 0)

    const uniqueStores = new Set(rawData.map((b: any) => b.store_name || b.shop_name).filter(Boolean)).size
    const uniqueOperators = new Set(rawData.map((b: any) => b.operator_name).filter(Boolean)).size

    worksheet.mergeCells('A5:C5')
    const kpi1 = worksheet.getCell('A5')
    kpi1.value = `📦 TOTAL TERPASANG: ${rawData.length} Spanduk`
    kpi1.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
    kpi1.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
    kpi1.alignment = { horizontal: 'center', vertical: 'middle' }

    worksheet.mergeCells('D5:G5')
    const kpi2 = worksheet.getCell('D5')
    kpi2.value = `📐 TOTAL LUAS: ${totalLuasM2.toFixed(2)} m²`
    kpi2.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
    kpi2.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
    kpi2.alignment = { horizontal: 'center', vertical: 'middle' }

    worksheet.mergeCells('H5:L5')
    const kpi3 = worksheet.getCell('H5')
    kpi3.value = `🏪 TOKO UNIK: ${uniqueStores} Toko`
    kpi3.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
    kpi3.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
    kpi3.alignment = { horizontal: 'center', vertical: 'middle' }

    worksheet.mergeCells('M5:R5')
    const kpi4 = worksheet.getCell('M5')
    kpi4.value = `👷 OPERATOR TERLIBAT: ${uniqueOperators} Orang`
    kpi4.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
    kpi4.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
    kpi4.alignment = { horizontal: 'center', vertical: 'middle' }
    worksheet.getRow(5).height = 24

    worksheet.addRow({}) // Row 6 spacer
    worksheet.getRow(6).height = 8

    // 4. Columns Setup
    worksheet.columns = [
      { key: 'no', width: 6 },
      { key: 'installed_at', width: 20 },
      { key: 'shop_name', width: 28 },
      { key: 'shop_address', width: 34 },
      { key: 'village', width: 18 },
      { key: 'district', width: 18 },
      { key: 'regency', width: 20 },
      { key: 'province', width: 18 },
      { key: 'shop_phone', width: 16 },
      { key: 'gmaps_link', width: 22 },
      { key: 'sales_name', width: 22 },
      { key: 'sales_branch', width: 16 },
      { key: 'operator_name', width: 22 },
      { key: 'operator_phone', width: 16 },
      { key: 'dimension', width: 22 },
      { key: 'area_m2', width: 14 },
      { key: 'qty', width: 10 },
      { key: 'photo_installed_link', width: 24 },
    ]

    // 5. Header Row (Row 7)
    const headerRow = worksheet.getRow(7)
    const columnTitles = [
      'No',
      'Waktu Pemasangan',
      'Nama Toko',
      'Alamat Toko',
      'Kelurahan/Desa',
      'Kecamatan',
      'Kota/Kabupaten',
      'Provinsi',
      'No Telp Toko',
      'Lokasi GPS (Maps)',
      'Sales Pemohon',
      'Cabang Sales',
      'Operator Pasang',
      'No Telp Operator',
      'Tipe & Ukuran',
      'Luas (m²)',
      'Qty',
      'Bukti Pemasangan',
    ]

    columnTitles.forEach((title, idx) => {
      const cell = headerRow.getCell(idx + 1)
      cell.value = title
      cell.font = { bold: true, size: 10, color: { argb: 'FFFFFFFF' } }
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF047857' } } // Emerald 700
      cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true }
      cell.border = {
        top: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        bottom: { style: 'medium', color: { argb: 'FF065F46' } },
        left: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        right: { style: 'thin', color: { argb: 'FFCBD5E1' } },
      }
    })
    headerRow.height = 28

    // 6. Data Rows
    rawData.forEach((item: any, index: number) => {
      const rowIdx = index + 8
      const row = worksheet.getRow(rowIdx)

      const lengthVal = Number(item.size_length || 0)
      const widthVal = Number(item.size_width || 0)
      const qtyVal = Number(item.qty || 1)
      const lMeter = lengthVal > 20 ? lengthVal / 100 : lengthVal
      const wMeter = widthVal > 20 ? widthVal / 100 : widthVal
      const areaM2 = +(lMeter * wMeter * qtyVal).toFixed(2)
      const dimensionStr = `${item.type || 'Spanduk'} ${formatBannerDimension(lengthVal, widthVal, 'dual')}`

      const lat = item.latitude || 0
      const lng = item.longitude || 0
      const hasCoords = Boolean(lat && lng && (lat !== 0 || lng !== 0))

      const photoUrl = resolvePhotoUrl(item.photo_installed)

      const rowValues = [
        index + 1,
        formatDate(item.installed_at || item.updated_at),
        item.store_name || item.shop_name || '-',
        item.shop_address || item.address || '-',
        item.village || '-',
        item.district || '-',
        item.regency || '-',
        item.province || '-',
        item.shop_phone || '-',
        hasCoords ? `https://www.google.com/maps?q=${lat},${lng}` : '-',
        item.sales_name || item.mysql_sales_name || '-',
        item.sales_branch || '-',
        item.operator_name || '-',
        item.operator_phone || '-',
        dimensionStr,
        areaM2,
        qtyVal,
        photoUrl || '-',
      ]

      const isEven = index % 2 === 1
      const bgArgb = isEven ? 'FFF8FAFC' : 'FFFFFFFF'

      rowValues.forEach((val, cIdx) => {
        const cell = row.getCell(cIdx + 1)
        
        // Handle Google Maps Hyperlink
        if (cIdx === 9 && hasCoords) {
          cell.value = {
            text: `📍 Maps (${Number(lat).toFixed(4)}, ${Number(lng).toFixed(4)})`,
            hyperlink: `https://www.google.com/maps?q=${lat},${lng}`,
          }
          cell.font = { size: 9, color: { argb: 'FF2563EB' }, underline: true }
        }
        // Handle Photo Installed Hyperlink
        else if (cIdx === 17 && photoUrl) {
          cell.value = {
            text: '📷 Buka Foto Bukti',
            hyperlink: photoUrl,
          }
          cell.font = { size: 9, color: { argb: 'FF059669' }, underline: true }
        }
        else {
          cell.value = val
          cell.font = { size: 9.5, color: { argb: 'FF1E293B' } }
        }

        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: bgArgb } }
        cell.border = {
          top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
        }

        // Alignments
        if (cIdx === 0) {
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
        } else if (cIdx === 1 || cIdx === 8 || cIdx === 9 || cIdx === 11 || cIdx === 13 || cIdx === 17) {
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
        } else if (cIdx === 15 || cIdx === 16) {
          cell.alignment = { horizontal: 'right', vertical: 'middle' }
        } else {
          cell.alignment = { horizontal: 'left', vertical: 'middle' }
        }
      })
      row.height = 24
    })

    // 7. Total Row at bottom
    const totalRowIdx = rawData.length + 8
    const totalRow = worksheet.getRow(totalRowIdx)
    worksheet.mergeCells(`A${totalRowIdx}:O${totalRowIdx}`)
    const totalLabelCell = worksheet.getCell(`A${totalRowIdx}`)
    totalLabelCell.value = 'TOTAL KESELURUHAN'
    totalLabelCell.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
    totalLabelCell.alignment = { horizontal: 'right', vertical: 'middle' }
    totalLabelCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    const totalAreaCell = worksheet.getCell(`P${totalRowIdx}`)
    totalAreaCell.value = { formula: `SUM(P8:P${totalRowIdx - 1})` }
    totalAreaCell.numFmt = '#,##0.00'
    totalAreaCell.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
    totalAreaCell.alignment = { horizontal: 'right', vertical: 'middle' }
    totalAreaCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    const totalQtyCell = worksheet.getCell(`Q${totalRowIdx}`)
    totalQtyCell.value = { formula: `SUM(Q8:Q${totalRowIdx - 1})` }
    totalQtyCell.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
    totalQtyCell.alignment = { horizontal: 'right', vertical: 'middle' }
    totalQtyCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    const lastCell = worksheet.getCell(`R${totalRowIdx}`)
    lastCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    for (let c = 1; c <= 18; c++) {
      const cell = totalRow.getCell(c)
      cell.border = {
        top: { style: 'thin', color: { argb: 'FF94A3B8' } },
        bottom: { style: 'double', color: { argb: 'FF0F172A' } },
        left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
        right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
      }
    }
    totalRow.height = 26

    // 8. Generate and download buffer
    const buffer = await workbook.xlsx.writeBuffer()
    const blob = new Blob([buffer], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })
    const sanitizedBranch = branchName.replace(/[^a-zA-Z0-9_-]/g, '_')
    const timestampStr = new Date().toISOString().slice(0, 10)
    saveAs(blob, `Bukti_Tayang_Terpasang_${sanitizedBranch}_${timestampStr}.xlsx`)

    toast.success(`Berhasil export ${rawData.length} data bukti tayang ke Excel!`, { id: 'export-history-excel' })
  } catch (err: any) {
    console.error('Export Excel error:', err)
    toast.error('Gagal mengekspor file Excel: ' + (err.message || 'Terjadi kesalahan.'), { id: 'export-history-excel' })
  } finally {
    isExporting.value = false
  }
}

const onSearchInput = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    if (activeTab.value === 'WAITING') {
      page.value = 1
    } else {
      historyPage.value = 1
    }
    fetchTasks()
  }, 500)
}

watch(activeTab, () => {
  search.value = ''
  if (activeTab.value === 'WAITING') {
    page.value = 1
  } else {
    historyPage.value = 1
  }
  fetchTasks()
})

onMounted(async () => {
  if (!masterStore.branches || masterStore.branches.length === 0) {
    await masterStore.fetchBranches()
  }
  fetchOperators()
  fetchTasks()
})

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

const openAssignModal = (task: any) => {
  selectedTaskForAssign.value = task
  selectedOperatorId.value = task.operator_id || ''
  operatorSearchQuery.value = ''
  showAssignModal.value = true
  if (operators.value.length === 0) {
    fetchOperators()
  }
}

const assignOperator = async () => {
  if (!selectedTaskForAssign.value) {
    return
  }

  isAssigning.value = true
  const isClearing = !selectedOperatorId.value
  toast.loading(isClearing ? 'Mengosongkan penugasan operator...' : 'Menugaskan operator spanduk...', { id: 'assign-op' })
  try {
    const res = await apiClient.put(`/branding/banners/${selectedTaskForAssign.value.id}/assign-operator`, {
      operator_id: selectedOperatorId.value || '',
    })
    toast.success(res.data?.message || (isClearing ? 'Penugasan operator berhasil dikosongkan.' : 'Operator spanduk berhasil ditugaskan.'), { id: 'assign-op' })
    showAssignModal.value = false
    fetchWaitingTasks(true)
  } catch (e: any) {
    toast.error('Gagal memperbarui penugasan: ' + (e.response?.data?.error || e.message), { id: 'assign-op' })
  } finally {
    isAssigning.value = false
  }
}

const openUploadModal = (task: any) => {
  selectedTask.value = task
  selectedFile.value = null
  previewImage.value = null
  latitude.value = null
  longitude.value = null
  installationDateTime.value = formatToDateTimeLocal(new Date())
  showUploadModal.value = true
}

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    selectedFile.value = file
    previewImage.value = URL.createObjectURL(file)
  }
}

const getGeolocation = () => {
  if (!navigator.geolocation) {
    toast.error('Geolocation tidak didukung oleh browser ini.')
    return
  }
  
  toast.loading('Mendapatkan lokasi...', { id: 'geo' })
  navigator.geolocation.getCurrentPosition(
    (position) => {
      latitude.value = position.coords.latitude
      longitude.value = position.coords.longitude
      toast.success('Lokasi berhasil didapatkan.', { id: 'geo' })
    },
    (error) => {
      console.error(error)
      toast.error('Gagal mendapatkan lokasi. Pastikan izin lokasi aktif.', { id: 'geo' })
    },
    { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 }
  )
}

const submitUpload = async () => {
  if (!selectedTask.value || !selectedFile.value) {
    toast.error('Pilih foto bukti pemasangan terlebih dahulu.')
    return
  }
  if (!latitude.value || !longitude.value) {
    toast.error('Lokasi GPS belum didapatkan. Klik "Ambil Lokasi Saat Ini".')
    return
  }

  isUploading.value = true
  toast.loading('Mengunggah bukti pemasangan...', { id: 'upload' })
  
  const formData = new FormData()
  formData.append('photo_installed', selectedFile.value)
  formData.append('latitude', latitude.value.toString())
  formData.append('longitude', longitude.value.toString())
  if (installationDateTime.value) {
    formData.append('installed_at', installationDateTime.value)
  }

  try {
    await apiClient.post(`/branding/banners/${selectedTask.value.id}/install`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    
    toast.success('Bukti pemasangan berhasil diunggah.', { id: 'upload' })
    showUploadModal.value = false
    fetchWaitingTasks(true)
  } catch (e: any) {
    toast.error('Gagal mengunggah: ' + (e.response?.data?.error || e.message), { id: 'upload' })
  } finally {
    isUploading.value = false
  }
}

const takeBanner = async (task: any) => {
  if (!confirm(`Ambil spanduk ini untuk pemasangan di ${task.shop_name}?`)) return
  try {
    toast.loading('Memproses...', { id: 'take' })
    await apiClient.put(`/branding/banners/${task.id}/operator-take`)
    toast.success('Spanduk berhasil diambil.', { id: 'take' })
    fetchWaitingTasks(true)
  } catch (e: any) {
    toast.error('Gagal mengambil spanduk: ' + (e.response?.data?.error || e.message), { id: 'take' })
  }
}

const openPreview = (url: string, title = 'Preview Foto') => {
  if (!url) return
  const finalUrl = resolvePhotoUrl(url)
  previewUrl.value = finalUrl
  previewTitle.value = title
  showPreviewModal.value = true
}

const closePreview = () => {
  showPreviewModal.value = false
  previewUrl.value = ''
  previewTitle.value = ''
}

const nextPage = () => {
  if (activeTab.value === 'WAITING') {
    if (page.value * limit.value < total.value) {
      page.value++
      fetchWaitingTasks()
    }
  } else {
    if (historyPage.value * historyLimit.value < historyTotal.value) {
      historyPage.value++
      fetchHistoryTasks()
    }
  }
}

const prevPage = () => {
  if (activeTab.value === 'WAITING') {
    if (page.value > 1) {
      page.value--
      fetchWaitingTasks()
    }
  } else {
    if (historyPage.value > 1) {
      historyPage.value--
      fetchHistoryTasks()
    }
  }
}

const totalPages = () => {
  if (activeTab.value === 'WAITING') {
    return Math.ceil(total.value / limit.value) || 1
  }
  return Math.ceil(historyTotal.value / historyLimit.value) || 1
}

const currentPage = () => {
  return activeTab.value === 'WAITING' ? page.value : historyPage.value
}

const canNextPage = () => {
  return currentPage() < totalPages()
}

const canPrevPage = () => {
  return currentPage() > 1
}

const isDriveUrl = (url: string) => {
  return url && url.includes('drive.google.com')
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <div class="flex items-center flex-wrap gap-2 sm:gap-3">
          <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <span>Bukti Tayang / Pemasangan</span>
            <span class="text-primary text-2xl font-bold">
              — {{ selectedBranchName }}
            </span>
          </h1>
          <span
            v-if="!canManageInstallation"
            class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-black bg-slate-100 text-slate-600 border border-slate-200 shadow-sm"
          >
            <Eye class="w-3.5 h-3.5 text-slate-400" /> Mode Pantau
          </span>
        </div>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <Camera class="w-4 h-4 text-primary" />
          Kelola dan dokumentasi foto bukti pemasangan spanduk di toko.
        </p>
      </div>
      <button
        @click="fetchTasks()"
        class="flex items-center gap-2 bg-white px-4 py-2 rounded-xl border border-slate-200 shadow-sm font-bold text-slate-600 hover:bg-slate-50 transition"
      >
        <RefreshCw :class="['w-4 h-4', { 'animate-spin': isLoading }]" /> Refresh
      </button>
    </div>

    <div
      v-if="!isOnline"
      class="bg-red-50 text-red-600 p-4 rounded-xl border border-red-200 flex items-center gap-3"
    >
      <AlertTriangle class="w-5 h-5" />
      <span class="font-medium"
        >Sistem sedang offline atau tidak dapat menghubungi server Branding.</span
      >
    </div>

    <!-- Tabs -->
    <div class="flex gap-4 border-b border-slate-200 overflow-x-auto pb-1">
      <button
        @click="activeTab = 'WAITING'"
        :class="[
          'whitespace-nowrap pb-3 text-sm font-bold transition-all border-b-2 flex items-center gap-2',
          activeTab === 'WAITING'
            ? 'border-primary text-primary'
            : 'border-transparent text-slate-400 hover:text-slate-600',
        ]"
      >
        <span>Menunggu Pemasangan</span>
        <span
          v-if="total > 0 && activeTab === 'WAITING'"
          class="px-2 py-0.5 text-xs rounded-full bg-primary/10 text-primary font-black"
        >
          {{ total }}
        </span>
      </button>
      <button
        @click="activeTab = 'HISTORY'"
        :class="[
          'whitespace-nowrap pb-3 text-sm font-bold transition-all border-b-2 flex items-center gap-2',
          activeTab === 'HISTORY'
            ? 'border-emerald-500 text-emerald-500'
            : 'border-transparent text-slate-400 hover:text-slate-600',
        ]"
      >
        <span>Terpasang (Galeri Riwayat)</span>
        <span
          v-if="historyTotal > 0 && activeTab === 'HISTORY'"
          class="px-2 py-0.5 text-xs rounded-full bg-emerald-50 text-emerald-600 font-black border border-emerald-100"
        >
          {{ historyTotal }}
        </span>
      </button>
    </div>

    <!-- Main Container -->
    <div
      class="bg-white border border-slate-200 rounded-3xl shadow-sm overflow-hidden flex flex-col"
    >
      <!-- SPOTLIGHT: TERPASANG HARI INI (Hanya tampil di Tab HISTORY) -->
      <div
        v-if="activeTab === 'HISTORY'"
        class="bg-gradient-to-br from-emerald-50/90 via-emerald-50/40 to-teal-50/20 border-b border-emerald-100/80 p-5 sm:p-6"
      >
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-2xl bg-emerald-500 text-white flex items-center justify-center shadow-sm shrink-0">
              <Sparkles class="w-4 h-4" />
            </div>
            <div>
              <div class="flex items-center gap-2.5 flex-wrap">
                <h2 class="text-base font-black text-slate-900 tracking-tight">
                  Terpasang Hari Ini
                </h2>
                <span
                  class="px-2.5 py-0.5 rounded-full text-xs font-black bg-emerald-100 text-emerald-800 border border-emerald-200/80 shadow-2xs"
                >
                  {{ todayTasks.length }} Pemasangan
                </span>
              </div>
              <p class="text-xs text-slate-500 font-medium mt-0.5">
                Ringkasan spanduk yang berhasil dipasang oleh operator pada {{ currentFormattedTodayDate }}.
              </p>
            </div>
          </div>
        </div>

        <!-- Loading State for Today Tasks -->
        <div v-if="isLoadingToday" class="py-6 text-center">
          <RefreshCw class="w-6 h-6 text-emerald-500 animate-spin mx-auto mb-2" />
          <span class="text-xs text-slate-500 font-medium">Memuat data terpasang hari ini...</span>
        </div>

        <!-- Empty State for Today Tasks -->
        <div
          v-else-if="todayTasks.length === 0"
          class="bg-white/80 border border-dashed border-emerald-200 rounded-2xl p-4 sm:p-5 text-center"
        >
          <Clock class="w-6 h-6 text-emerald-400 mx-auto mb-1.5" />
          <p class="text-xs font-bold text-slate-700">Belum ada spanduk terpasang hari ini</p>
          <p class="text-[11px] text-slate-400 mt-0.5">
            Pemasangan yang diunggah oleh operator hari ini akan langsung otomatis muncul di sorotan atas ini.
          </p>
        </div>

        <!-- Cards List for Today Tasks -->
        <div
          v-else
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3.5"
        >
          <div
            v-for="task in todayTasks"
            :key="'today-' + task.id"
            class="bg-white rounded-2xl border border-emerald-100 p-3 shadow-2xs hover:shadow-md hover:border-emerald-300 transition-all flex flex-col justify-between group relative overflow-hidden"
          >
            <div class="absolute top-0 right-0 bg-emerald-500 text-white text-[9px] font-black px-2 py-0.5 rounded-bl-lg uppercase tracking-wider">
              Hari Ini
            </div>

            <div class="flex items-start gap-3">
              <!-- Foto Bukti Thumbnail -->
              <div
                class="relative w-16 h-16 rounded-xl overflow-hidden bg-slate-100 border border-slate-200 shrink-0 cursor-pointer group/img"
                @click="task.photo_installed && openPreview(task.photo_installed, `Bukti Pemasangan - ${task.shop_name}`)"
              >
                <img
                  v-if="task.photo_installed && !isDriveUrl(task.photo_installed)"
                  :src="task.photo_installed"
                  alt="Bukti Pasang"
                  class="w-full h-full object-cover group-hover/img:scale-110 transition-transform duration-300"
                  loading="lazy"
                />
                <div
                  v-else
                  class="flex items-center justify-center w-full h-full text-slate-400"
                >
                  <ImageIcon class="w-6 h-6 text-slate-300" />
                </div>
                <div
                  v-if="task.photo_installed"
                  class="absolute inset-0 bg-black/30 opacity-0 group-hover/img:opacity-100 transition-opacity flex items-center justify-center text-white"
                >
                  <ZoomIn class="w-4 h-4" />
                </div>
              </div>

              <!-- Info Singkat -->
              <div class="min-w-0 flex-1 pr-10">
                <h4 class="text-xs font-black text-slate-900 truncate" :title="task.shop_name">
                  {{ task.shop_name }}
                </h4>
                <p class="text-[11px] text-slate-500 font-medium truncate mt-0.5" :title="task.shop_address">
                  {{ task.shop_address || 'Alamat tidak tercatat' }}
                </p>
                <div class="flex items-center gap-1.5 mt-1.5">
                  <span class="inline-flex items-center px-1.5 py-0.5 rounded bg-emerald-50 text-emerald-700 text-[10px] font-bold border border-emerald-100">
                    {{ task.type }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Footer Kartu -->
            <div class="mt-2.5 pt-2 border-t border-slate-100 flex items-center justify-between text-[10px] text-slate-500">
              <div class="flex items-center gap-1 truncate">
                <UserCheck class="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                <span class="truncate font-semibold text-slate-700">{{ task.operator_name || 'Operator' }}</span>
              </div>
              <span class="text-slate-400 font-medium shrink-0 flex items-center gap-0.5">
                <Clock class="w-3 h-3 text-slate-400" />
                {{ task.installed_at }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Toolbar -->
      <div
        class="p-5 border-b border-slate-200 flex flex-col gap-4 bg-slate-50/50"
      >
        <div class="flex flex-col lg:flex-row gap-3 justify-between items-stretch lg:items-center">
          <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 flex-1">
            <!-- Search Bar -->
            <div class="relative w-full sm:w-72">
              <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
              <input
                type="text"
                :placeholder="
                  activeTab === 'WAITING'
                    ? 'Cari toko menunggu pemasangan...'
                    : 'Cari riwayat toko terpasang...'
                "
                v-model="search"
                @input="onSearchInput"
                class="w-full pl-10 pr-4 py-2 bg-white border border-slate-200 rounded-xl text-xs sm:text-sm font-medium focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all placeholder:text-slate-400 shadow-xs"
              />
            </div>

            <!-- Filter Cabang -->
            <div class="relative w-full sm:w-48">
              <Building2 class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
              <select
                v-model="selectedBranchFilter"
                @change="() => { page = 1; historyPage = 1; fetchTasks() }"
                class="w-full pl-9 pr-8 py-2 bg-white border border-slate-200 rounded-xl text-xs font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition shadow-xs cursor-pointer"
              >
                <option value="ALL_BRANCHES">Semua Cabang</option>
                <option disabled>──────────</option>
                <option v-for="branch in masterStore.branches" :key="branch.id" :value="String(branch.id)">
                  {{ branch.name }}
                </option>
              </select>
            </div>

            <!-- Filter Operator (untuk Tab HISTORY) -->
            <div v-if="activeTab === 'HISTORY'" class="relative w-full sm:w-56">
              <Users class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
              <select
                v-model="selectedOperatorFilter"
                @change="onOperatorFilterChange"
                class="w-full pl-9 pr-8 py-2 bg-white border border-slate-200 rounded-xl text-xs font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 transition shadow-xs appearance-none cursor-pointer"
              >
                <option value="all">Semua Operator</option>
                <option v-for="op in operators" :key="op.id" :value="op.id">
                  {{ op.name }}
                </option>
              </select>
              <div class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-slate-400 text-[10px]">
                ▼
              </div>
            </div>
          </div>

          <!-- View Mode Switch, Export Excel & Reset (for HISTORY tab) -->
          <div v-if="activeTab === 'HISTORY'" class="flex flex-wrap items-center gap-2 justify-end">
            <!-- Export Excel Button -->
            <button
              @click="exportHistoryToExcel"
              :disabled="isExporting"
              class="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-bold text-emerald-700 bg-emerald-50 hover:bg-emerald-100 rounded-xl border border-emerald-200 transition shadow-xs cursor-pointer disabled:opacity-50"
              title="Export data terpasang ke format Excel (.xlsx)"
            >
              <Download v-if="!isExporting" class="w-3.5 h-3.5" />
              <RefreshCw v-else class="w-3.5 h-3.5 animate-spin" />
              <span>{{ isExporting ? 'Mengexport...' : 'Export Excel' }}</span>
            </button>

            <!-- Reset filter button -->
            <button
              v-if="hasActiveHistoryFilters"
              @click="resetHistoryFilters"
              class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-bold text-red-600 bg-red-50 hover:bg-red-100 rounded-xl border border-red-200 transition shadow-xs cursor-pointer"
              title="Reset semua filter"
            >
              <X class="w-3.5 h-3.5" /> Reset
            </button>

            <div class="flex items-center gap-1 bg-slate-200/60 p-1 rounded-xl">
              <button
                @click="historyViewMode = 'GALLERY'"
                :class="[
                  'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold transition-all cursor-pointer',
                  historyViewMode === 'GALLERY'
                    ? 'bg-white text-emerald-600 shadow-xs'
                    : 'text-slate-500 hover:text-slate-700',
                ]"
              >
                <LayoutGrid class="w-3.5 h-3.5" /> Galeri
              </button>
              <button
                @click="historyViewMode = 'TABLE'"
                :class="[
                  'flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold transition-all cursor-pointer',
                  historyViewMode === 'TABLE'
                    ? 'bg-white text-emerald-600 shadow-xs'
                    : 'text-slate-500 hover:text-slate-700',
                ]"
              >
                <List class="w-3.5 h-3.5" /> Tabel
              </button>
            </div>
          </div>
        </div>

        <!-- Custom Date Range & Preset Filter Bar (untuk Tab HISTORY) -->
        <div
          v-if="activeTab === 'HISTORY'"
          class="flex flex-wrap items-center gap-2 pt-2 border-t border-slate-200/60"
        >
          <span class="text-xs font-bold text-slate-500 flex items-center gap-1">
            <Filter class="w-3.5 h-3.5 text-slate-400" /> Periode:
          </span>

          <div class="flex flex-wrap items-center gap-2 bg-white border border-slate-200 p-1 rounded-xl shadow-xs">
            <div class="flex items-center gap-0.5 bg-slate-100 p-0.5 rounded-lg">
              <button
                type="button"
                @click="setDatePreset('all')"
                :class="[
                  'px-2 py-1 text-[11px] font-bold rounded-md transition-all cursor-pointer',
                  datePreset === 'all' && !startDate && !endDate ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                Semua
              </button>
              <button
                type="button"
                @click="setDatePreset('today')"
                :class="[
                  'px-2 py-1 text-[11px] font-bold rounded-md transition-all cursor-pointer',
                  datePreset === 'today' ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                Hari Ini
              </button>
              <button
                type="button"
                @click="setDatePreset('7d')"
                :class="[
                  'px-2 py-1 text-[11px] font-bold rounded-md transition-all cursor-pointer',
                  datePreset === '7d' ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                7 Hari
              </button>
              <button
                type="button"
                @click="setDatePreset('30d')"
                :class="[
                  'px-2 py-1 text-[11px] font-bold rounded-md transition-all cursor-pointer',
                  datePreset === '30d' ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                30 Hari
              </button>
              <button
                type="button"
                @click="setDatePreset('this_month')"
                :class="[
                  'px-2 py-1 text-[11px] font-bold rounded-md transition-all cursor-pointer',
                  datePreset === 'this_month' ? 'bg-white shadow-xs text-slate-900' : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                Bulan Ini
              </button>
            </div>

            <div class="flex items-center gap-1.5 text-slate-500 text-xs font-semibold px-1">
              <Calendar class="w-3.5 h-3.5 text-slate-400 shrink-0" />
              <input
                v-model="startDate"
                type="date"
                @change="onCustomDateChange"
                class="bg-slate-50 border border-slate-200 rounded-lg px-2 py-1 text-[11px] font-medium text-slate-700 focus:outline-none focus:ring-1 focus:ring-emerald-500/30"
                title="Tanggal Awal"
              />
              <span class="text-slate-400">—</span>
              <input
                v-model="endDate"
                type="date"
                @change="onCustomDateChange"
                class="bg-slate-50 border border-slate-200 rounded-lg px-2 py-1 text-[11px] font-medium text-slate-700 focus:outline-none focus:ring-1 focus:ring-emerald-500/30"
                title="Tanggal Akhir"
              />
              <button
                v-if="startDate || endDate"
                type="button"
                @click="setDatePreset('all')"
                class="text-slate-400 hover:text-red-500 p-0.5 cursor-pointer"
                title="Reset rentang tanggal"
              >
                <X class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB 1: MENUNGGU PEMASANGAN -->
      <div v-if="activeTab === 'WAITING'" class="overflow-x-auto w-full">
        <table class="w-full text-left border-collapse min-w-[800px]">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200">
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Tanggal Validasi</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Nama Toko & Sales</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Tipe & Ukuran</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Status Pengerjaan</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="isLoading && waitingTasks.length === 0">
              <td colspan="5" class="py-16 text-center">
                <RefreshCw class="w-8 h-8 text-primary animate-spin mx-auto mb-3" />
                <p class="text-sm font-medium text-slate-500">Memuat data spanduk...</p>
              </td>
            </tr>
            <tr v-else-if="waitingTasks.length === 0">
              <td colspan="5" class="py-16 text-center">
                <CheckCircle2 class="w-12 h-12 text-emerald-400 mx-auto mb-3" />
                <p class="text-base font-bold text-slate-700">Tidak ada spanduk menunggu pemasangan!</p>
                <p class="text-xs text-slate-400 mt-1">Semua spanduk sudah terpasang atau belum ada data di cabang ini.</p>
              </td>
            </tr>
            <template v-else>
              <tr
                v-for="task in waitingTasks"
                :key="task.id"
                class="hover:bg-slate-50/50 transition-colors group"
              >
                <td class="py-4 px-6 align-top w-44">
                  <div class="flex items-center gap-2 text-sm font-bold text-slate-600 whitespace-nowrap">
                    <Clock class="w-4 h-4 text-slate-400" />
                    {{ task.created_at }}
                  </div>
                </td>
                <td class="py-4 px-6 align-top max-w-[280px]">
                  <div class="text-sm font-black text-slate-900 truncate">{{ task.shop_name }}</div>
                  <!-- Hover Card Alamat Toko -->
                  <StoreAddressHoverCard
                    :address="task.shop_address"
                    :shop-name="task.shop_name"
                    :province="task.province"
                    :regency="task.regency"
                    :district="task.district"
                    :village="task.village"
                    :phone="task.shop_phone"
                    :latitude="task.latitude"
                    :longitude="task.longitude"
                    :sales-name="task.sales_name"
                    :sales-branch="task.sales_branch"
                    class="mt-1 text-[11px]"
                  />
                  <div class="text-[11px] font-medium text-slate-500 mt-2 flex items-center gap-2">
                    <span>Sales: <strong class="text-slate-700">{{ task.sales_name }}</strong></span>
                    <span v-if="task.sales_branch" class="px-1.5 py-0.5 rounded bg-slate-100 text-[10px] text-slate-600 font-semibold">
                      {{ task.sales_branch }}
                    </span>
                  </div>
                </td>
                <td class="py-4 px-6 align-top">
                  <div class="inline-flex items-center px-2.5 py-1 rounded-lg bg-slate-100 text-xs font-bold text-slate-700 whitespace-nowrap">
                    {{ task.type }}
                  </div>
                </td>
                <td class="py-4 px-6 align-top w-56">
                  <div v-if="task.operator_id || task.status === 'diterima_operator'" class="space-y-1">
                    <span
                      class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-black border bg-blue-50 text-blue-700 border-blue-200"
                    >
                      <UserCheck class="w-4 h-4 text-blue-600" />
                      {{ task.operator_name ? `Dibawa: ${task.operator_name}` : 'Dibawa Operator' }}
                    </span>
                    <div v-if="task.operator_phone" class="text-[10px] text-slate-400 font-medium pl-1">
                      Telp: {{ task.operator_phone }}
                    </div>
                  </div>
                  <span
                    v-else
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-black border bg-amber-50 text-amber-700 border-amber-200"
                  >
                    <AlertTriangle class="w-4 h-4 text-amber-600" />
                    Siap Dipasang
                  </span>
                </td>
                <td class="py-4 px-6 align-top w-52 text-center">
                  <div v-if="canManageInstallation" class="flex flex-col gap-2">
                    <button
                      v-if="!task.operator_id || task.status === 'ready_for_operator'"
                      @click="openAssignModal(task)"
                      class="flex items-center justify-center gap-2 w-full bg-indigo-600 hover:bg-indigo-700 text-white px-3 py-2 rounded-xl text-xs font-black shadow-sm transition-all"
                    >
                      <UserPlus class="w-4 h-4" /> Tugaskan Operator
                    </button>
                    <template v-else>
                      <button
                        @click="openUploadModal(task)"
                        class="flex items-center justify-center gap-2 w-full bg-emerald-500 hover:bg-emerald-600 text-white px-3 py-2 rounded-xl text-xs font-black shadow-sm transition-all"
                      >
                        <Upload class="w-4 h-4" /> Upload Bukti
                      </button>
                      <button
                        @click="openAssignModal(task)"
                        class="flex items-center justify-center gap-1.5 w-full bg-white hover:bg-slate-50 text-slate-700 border border-slate-200 px-3 py-1.5 rounded-xl text-[11px] font-bold shadow-sm transition-all"
                      >
                        <Users class="w-3.5 h-3.5 text-slate-500" /> Ganti Operator
                      </button>
                    </template>
                  </div>
                  <span
                    v-else
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-100 text-slate-600 text-xs font-bold border border-slate-200"
                  >
                    <Clock class="w-3.5 h-3.5 text-slate-400" /> Menunggu Pemasangan
                  </span>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <!-- TAB 2: TERPASANG (GALLERY VIEW) -->
      <div v-else-if="activeTab === 'HISTORY' && historyViewMode === 'GALLERY'" class="p-6">
        <div v-if="isLoading && historyTasks.length === 0" class="py-16 text-center">
          <RefreshCw class="w-8 h-8 text-emerald-500 animate-spin mx-auto mb-3" />
          <p class="text-sm font-medium text-slate-500">Memuat galeri terpasang...</p>
        </div>

        <div v-else-if="historyTasks.length === 0" class="py-16 text-center">
          <StoreIcon class="w-12 h-12 text-slate-300 mx-auto mb-3" />
          <p class="text-base font-bold text-slate-700">Belum ada riwayat pemasangan.</p>
          <p class="text-xs text-slate-400 mt-1">Spanduk yang sudah diupload buktinya akan muncul di galeri ini.</p>
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
          <div
            v-for="task in historyTasks"
            :key="task.id"
            class="bg-white border border-slate-200/80 rounded-2xl overflow-hidden shadow-sm hover:shadow-md transition-all flex flex-col group"
          >
            <!-- Card Header -->
            <div class="p-4 border-b border-slate-100 flex items-start justify-between gap-3 bg-slate-50/40">
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
                  <h3 class="text-sm font-black text-slate-900 truncate" :title="task.shop_name">
                    {{ task.shop_name }}
                  </h3>
                  <span class="inline-flex items-center px-2 py-0.5 rounded-md bg-emerald-50 text-emerald-600 text-[10px] font-black border border-emerald-100 whitespace-nowrap">
                    Terpasang
                  </span>
                </div>
                <!-- Hover Card Alamat Toko -->
                <StoreAddressHoverCard
                  :address="task.shop_address"
                  :shop-name="task.shop_name"
                  :province="task.province"
                  :regency="task.regency"
                  :district="task.district"
                  :village="task.village"
                  :phone="task.shop_phone"
                  :latitude="task.latitude"
                  :longitude="task.longitude"
                  :sales-name="task.sales_name"
                  :sales-branch="task.sales_branch"
                  class="mt-0.5 text-[11px]"
                />
              </div>
            </div>

            <!-- Photos Showcase (Dual Gallery: Foto Toko & Bukti Pasang) -->
            <div class="grid grid-cols-2 gap-1.5 p-3 bg-slate-100/50">
              <!-- Foto Toko -->
              <div class="flex flex-col gap-1.5">
                <span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider px-1">
                  Foto Toko
                </span>
                <div
                  class="relative aspect-[4/3] rounded-xl overflow-hidden bg-slate-200 border border-slate-200/80 flex items-center justify-center group/img cursor-pointer"
                  @click="task.shop_photo && openPreview(task.shop_photo, `Foto Toko - ${task.shop_name}`)"
                >
                  <img
                    v-if="task.shop_photo"
                    :src="task.shop_photo"
                    alt="Foto Toko"
                    class="w-full h-full object-cover group-hover/img:scale-105 transition-transform duration-300"
                    loading="lazy"
                  />
                  <div
                    v-else
                    class="flex flex-col items-center justify-center text-slate-400 text-[11px] font-medium p-2 text-center"
                  >
                    <StoreIcon class="w-6 h-6 mb-1 text-slate-300" />
                    <span>Tanpa Foto Toko</span>
                  </div>

                  <!-- Hover overlay if photo exists -->
                  <div
                    v-if="task.shop_photo"
                    class="absolute inset-0 bg-black/30 opacity-0 group-hover/img:opacity-100 transition-opacity flex items-center justify-center text-white"
                  >
                    <ZoomIn class="w-5 h-5" />
                  </div>
                </div>
              </div>

              <!-- Foto Bukti Pasang -->
              <div class="flex flex-col gap-1.5">
                <span class="text-[10px] font-bold text-slate-500 uppercase tracking-wider px-1">
                  Bukti Pemasangan
                </span>
                <div
                  class="relative aspect-[4/3] rounded-xl overflow-hidden bg-slate-200 border border-slate-200/80 flex items-center justify-center group/img cursor-pointer"
                  @click="task.photo_installed && openPreview(task.photo_installed, `Bukti Pemasangan - ${task.shop_name}`)"
                >
                  <img
                    v-if="task.photo_installed && !isDriveUrl(task.photo_installed)"
                    :src="task.photo_installed"
                    alt="Bukti Pasang"
                    class="w-full h-full object-cover group-hover/img:scale-105 transition-transform duration-300"
                    loading="lazy"
                  />
                  <div
                    v-else-if="task.photo_installed && isDriveUrl(task.photo_installed)"
                    class="flex flex-col items-center justify-center text-indigo-600 bg-indigo-50/50 w-full h-full p-2 text-center"
                  >
                    <ImageIcon class="w-6 h-6 mb-1 text-indigo-500" />
                    <span class="text-[10px] font-bold">Google Drive</span>
                    <span class="text-[9px] text-indigo-400">Klik utk lihat</span>
                  </div>
                  <div
                    v-else
                    class="flex flex-col items-center justify-center text-slate-400 text-[11px] font-medium p-2 text-center"
                  >
                    <ImageIcon class="w-6 h-6 mb-1 text-slate-300" />
                    <span>Tanpa Bukti</span>
                  </div>

                  <!-- Hover overlay if photo exists -->
                  <div
                    v-if="task.photo_installed"
                    class="absolute inset-0 bg-black/30 opacity-0 group-hover/img:opacity-100 transition-opacity flex items-center justify-center text-white"
                  >
                    <ZoomIn class="w-5 h-5" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Card Info Body -->
            <div class="p-4 flex-1 flex flex-col justify-between space-y-3">
              <div class="flex items-center justify-between text-xs font-semibold text-slate-600 pt-1">
                <span class="px-2.5 py-1 rounded-lg bg-slate-100 text-slate-700 font-bold">
                  {{ task.type }}
                </span>
                <span class="text-[11px] text-slate-400 flex items-center gap-1 font-medium">
                  <Clock class="w-3.5 h-3.5" /> {{ task.installed_at }}
                </span>
              </div>

              <div class="pt-2 border-t border-slate-100 space-y-1.5">
                <div class="flex items-center justify-between text-xs text-slate-500">
                  <div class="flex items-center gap-1.5 truncate">
                    <User class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    <span class="truncate">Sales: <strong class="text-slate-700">{{ task.sales_name }}</strong></span>
                  </div>
                  <span v-if="task.sales_branch" class="px-2 py-0.5 rounded bg-slate-100 text-[10px] font-bold text-slate-600 shrink-0">
                    {{ task.sales_branch }}
                  </span>
                </div>
                <div class="flex items-center justify-between text-xs text-slate-500">
                  <div class="flex items-center gap-1.5 truncate">
                    <UserCheck class="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                    <span class="truncate">Operator: <strong class="text-slate-800">{{ task.operator_name || 'Tidak tercatat' }}</strong></span>
                  </div>
                  <span v-if="task.operator_phone" class="text-[10px] text-slate-400 font-medium shrink-0">
                    {{ task.operator_phone }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB 2: TERPASANG (TABLE VIEW OPTION) -->
      <div v-else-if="activeTab === 'HISTORY' && historyViewMode === 'TABLE'" class="overflow-x-auto w-full">
        <table class="w-full text-left border-collapse min-w-[800px]">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200">
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Tanggal Pasang</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Nama Toko & Alamat</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Sales & Operator</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Tipe & Ukuran</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap">Foto Toko & Bukti</th>
              <th class="py-4 px-6 text-xs font-black text-slate-500 uppercase tracking-wider whitespace-nowrap text-center">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="isLoading && historyTasks.length === 0">
              <td colspan="6" class="py-16 text-center">
                <RefreshCw class="w-8 h-8 text-emerald-500 animate-spin mx-auto mb-3" />
                <p class="text-sm font-medium text-slate-500">Memuat data riwayat...</p>
              </td>
            </tr>
            <tr v-else-if="historyTasks.length === 0">
              <td colspan="6" class="py-16 text-center">
                <p class="text-base font-bold text-slate-700">Belum ada riwayat pemasangan.</p>
              </td>
            </tr>
            <template v-else>
              <tr
                v-for="task in historyTasks"
                :key="task.id"
                class="hover:bg-slate-50/50 transition-colors group"
              >
                <td class="py-4 px-6 align-top w-44">
                  <div class="flex items-center gap-2 text-sm font-bold text-slate-600 whitespace-nowrap">
                    <Clock class="w-4 h-4 text-slate-400" />
                    {{ task.installed_at }}
                  </div>
                </td>
                <td class="py-4 px-6 align-top max-w-[260px]">
                  <div class="text-sm font-black text-slate-900 truncate">{{ task.shop_name }}</div>
                  <!-- Hover Card Alamat Toko -->
                  <StoreAddressHoverCard
                    :address="task.shop_address"
                    :shop-name="task.shop_name"
                    :province="task.province"
                    :regency="task.regency"
                    :district="task.district"
                    :village="task.village"
                    :latitude="task.latitude"
                    :longitude="task.longitude"
                    :sales-name="task.sales_name"
                    :sales-branch="task.sales_branch"
                    class="mt-1 text-[11px]"
                  />
                </td>
                <td class="py-4 px-6 align-top max-w-[240px]">
                  <div class="space-y-1.5">
                    <div class="flex items-center gap-1.5 text-xs text-slate-600">
                      <User class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                      <span class="truncate">Sales: <strong class="text-slate-800">{{ task.sales_name }}</strong></span>
                      <span v-if="task.sales_branch" class="px-1.5 py-0.5 rounded bg-slate-100 text-[10px] text-slate-600 font-semibold shrink-0">
                        {{ task.sales_branch }}
                      </span>
                    </div>
                    <div class="flex items-center gap-1.5 text-xs text-slate-600">
                      <UserCheck class="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                      <span class="truncate">Operator: <strong class="text-slate-800">{{ task.operator_name || 'Tidak tercatat' }}</strong></span>
                    </div>
                  </div>
                </td>
                <td class="py-4 px-6 align-top">
                  <div class="inline-flex items-center px-2.5 py-1 rounded-lg bg-slate-100 text-xs font-bold text-slate-700 whitespace-nowrap">
                    {{ task.type }}
                  </div>
                </td>
                <td class="py-4 px-6 align-top w-64">
                  <div class="flex items-center gap-2">
                    <button
                      v-if="task.shop_photo"
                      @click="openPreview(task.shop_photo, `Foto Toko - ${task.shop_name}`)"
                      class="inline-flex items-center gap-1 text-[11px] font-bold text-slate-700 bg-slate-100 hover:bg-slate-200 px-2.5 py-1.5 rounded-lg transition cursor-pointer"
                    >
                      <StoreIcon class="w-3.5 h-3.5" /> Toko
                    </button>
                    <button
                      v-if="task.photo_installed"
                      @click="openPreview(task.photo_installed, `Bukti Pemasangan - ${task.shop_name}`)"
                      class="inline-flex items-center gap-1 text-[11px] font-bold text-white bg-indigo-500 hover:bg-indigo-600 px-2.5 py-1.5 rounded-lg transition cursor-pointer"
                    >
                      <ImageIcon class="w-3.5 h-3.5" /> Bukti
                    </button>
                    <span v-if="!task.shop_photo && !task.photo_installed" class="text-xs text-slate-400 italic">
                      Tidak ada foto
                    </span>
                  </div>
                </td>
                <td class="py-4 px-6 align-top w-36 text-center">
                  <span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-emerald-50 text-emerald-600 text-xs font-black border border-emerald-100">
                    <CheckCircle2 class="w-4 h-4" /> Selesai
                  </span>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="p-4 border-t border-slate-200 flex flex-col sm:flex-row items-center justify-between gap-4 bg-slate-50/50">
        <div class="text-sm font-medium text-slate-500">
          Menampilkan <span class="font-bold text-slate-700">{{ activeTab === 'WAITING' ? waitingTasks.length : historyTasks.length }}</span> dari
          <span class="font-bold text-slate-700">{{
            activeTab === 'WAITING' ? total : historyTotal
          }}</span>
          total data
        </div>
        <div class="flex items-center gap-3">
          <button
            @click="prevPage"
            :disabled="!canPrevPage()"
            class="flex items-center justify-center w-8 h-8 rounded-lg border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 hover:text-slate-900 disabled:opacity-50 disabled:cursor-not-allowed transition-colors shadow-sm"
          >
            <ChevronLeft class="w-4 h-4" />
          </button>
          <span class="text-sm font-bold text-slate-700 min-w-[3rem] text-center">
            {{ currentPage() }} / {{ totalPages() }}
          </span>
          <button
            @click="nextPage"
            :disabled="!canNextPage()"
            class="flex items-center justify-center w-8 h-8 rounded-lg border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 hover:text-slate-900 disabled:opacity-50 disabled:cursor-not-allowed transition-colors shadow-sm"
          >
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- Assign Operator Modal -->
    <div v-if="showAssignModal" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="!isAssigning && (showAssignModal = false)"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-lg relative z-10 shadow-2xl p-6 sm:p-8 animate-in fade-in zoom-in duration-300 flex flex-col max-h-[90vh]"
      >
        <!-- Header -->
        <div class="flex justify-between items-center pb-4 border-b border-slate-100">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-indigo-50 border border-indigo-100 flex items-center justify-center text-indigo-600 shadow-sm">
              <Users class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-lg font-black text-slate-900">Tugaskan Operator Spanduk</h3>
              <p class="text-xs text-slate-500 font-medium">Pilih operator lapangan untuk memasang spanduk ini</p>
            </div>
          </div>
          <button
            v-if="!isAssigning"
            @click="showAssignModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-600 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Banner Info Summary -->
        <div class="my-4 p-3.5 bg-slate-50 border border-slate-200/70 rounded-2xl space-y-1.5">
          <div class="flex items-center justify-between text-xs">
            <span class="text-slate-500 font-semibold">Toko Target:</span>
            <span class="font-black text-slate-900 truncate max-w-[240px]">{{ selectedTaskForAssign?.shop_name }}</span>
          </div>
          <div class="flex items-center justify-between text-xs" v-if="selectedTaskForAssign?.type">
            <span class="text-slate-500 font-semibold">Tipe & Ukuran:</span>
            <span class="font-bold text-slate-800">{{ selectedTaskForAssign?.type }}</span>
          </div>
          <div class="flex items-center justify-between text-xs" v-if="selectedTaskForAssign?.sales_name">
            <span class="text-slate-500 font-semibold">Sales Pemohon:</span>
            <span class="font-bold text-slate-800">{{ selectedTaskForAssign?.sales_name }}</span>
          </div>
          <div class="flex items-center justify-between text-xs" v-if="selectedTaskForAssign?.shop_address">
            <span class="text-slate-500 font-semibold">Alamat:</span>
            <span class="font-medium text-slate-600 truncate max-w-[240px]" :title="selectedTaskForAssign?.shop_address">{{ selectedTaskForAssign?.shop_address }}</span>
          </div>
        </div>

        <!-- Search Operator Input -->
        <div class="mb-3">
          <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
            Pilih Operator Lapangan
          </label>
          <div class="relative">
            <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              v-model="operatorSearchQuery"
              type="text"
              placeholder="Cari nama atau telepon operator..."
              class="w-full pl-9 pr-4 py-2.5 bg-slate-50/70 border border-slate-200 rounded-xl text-xs font-medium focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all"
            />
          </div>
        </div>

        <!-- Operator List (Scrollable) -->
        <div class="flex-1 overflow-y-auto space-y-2 pr-1 min-h-[160px] max-h-[260px]">
          <!-- Pilihan Kosongkan Penugasan (Unassign) -->
          <label
            :class="[
              'flex items-center justify-between p-3 rounded-2xl border cursor-pointer transition-all',
              selectedOperatorId === ''
                ? 'bg-amber-50/80 border-amber-500 shadow-sm ring-1 ring-amber-500'
                : 'bg-amber-50/30 border-dashed border-amber-200 hover:bg-amber-50/60'
            ]"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div
                :class="[
                  'w-9 h-9 rounded-xl font-black text-xs flex items-center justify-center shrink-0',
                  selectedOperatorId === '' ? 'bg-amber-600 text-white' : 'bg-amber-100 text-amber-700'
                ]"
              >
                <UserX class="w-4 h-4" />
              </div>
              <div class="min-w-0">
                <div class="text-xs font-black text-slate-900 truncate">Kosongkan Penugasan Operator</div>
                <div class="text-[11px] text-amber-700/80 font-medium mt-0.5">
                  Kembalikan status ke Siap Dipasang (Belum Ditugaskan)
                </div>
              </div>
            </div>
            <div class="flex items-center gap-2 pl-2 shrink-0">
              <input
                type="radio"
                :name="'operator_select'"
                :value="''"
                v-model="selectedOperatorId"
                class="w-4 h-4 text-amber-600 focus:ring-amber-500 border-amber-300"
              />
            </div>
          </label>

          <div v-if="isLoadingOperators" class="py-10 text-center">
            <RefreshCw class="w-6 h-6 text-indigo-500 animate-spin mx-auto mb-2" />
            <p class="text-xs font-medium text-slate-500">Memuat data operator...</p>
          </div>
          <div v-else-if="filteredOperators.length === 0" class="py-8 text-center bg-slate-50 rounded-2xl border border-dashed border-slate-200">
            <UserPlus class="w-8 h-8 text-slate-300 mx-auto mb-1.5" />
            <p class="text-xs font-bold text-slate-600">Tidak ada operator ditemukan</p>
            <p class="text-[11px] text-slate-400">Pastikan akun operator sudah terdaftar di cabang ini.</p>
          </div>
          <label
            v-else
            v-for="op in filteredOperators"
            :key="op.id"
            :class="[
              'flex items-center justify-between p-3 rounded-2xl border cursor-pointer transition-all',
              selectedOperatorId === op.id
                ? 'bg-indigo-50/70 border-indigo-500 shadow-sm ring-1 ring-indigo-500'
                : 'bg-white border-slate-200/80 hover:bg-slate-50/70'
            ]"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div
                :class="[
                  'w-9 h-9 rounded-xl font-black text-xs flex items-center justify-center shrink-0',
                  selectedOperatorId === op.id ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-700'
                ]"
              >
                {{ op.name?.charAt(0)?.toUpperCase() || 'O' }}
              </div>
              <div class="min-w-0">
                <div class="text-xs font-black text-slate-900 truncate">{{ op.name }}</div>
                <div class="flex items-center gap-2 text-[11px] text-slate-500 font-medium mt-0.5">
                  <span v-if="op.phone">{{ op.phone }}</span>
                  <span v-if="op.area_code" class="px-1.5 py-0.2 rounded bg-slate-100 text-[10px] font-semibold text-slate-600">
                    {{ op.area_code }}
                  </span>
                </div>
              </div>
            </div>
            <div class="flex items-center gap-2 pl-2 shrink-0">
              <input
                type="radio"
                :name="'operator_select'"
                :value="op.id"
                v-model="selectedOperatorId"
                class="w-4 h-4 text-indigo-600 focus:ring-indigo-500 border-slate-300"
              />
            </div>
          </label>
        </div>

        <!-- Modal Footer Actions -->
        <div class="mt-5 pt-4 border-t border-slate-100 flex items-center justify-end gap-3">
          <button
            type="button"
            @click="showAssignModal = false"
            :disabled="isAssigning"
            class="px-4 py-2.5 rounded-xl border border-slate-200 text-xs font-bold text-slate-600 hover:bg-slate-50 transition-colors disabled:opacity-50"
          >
            Batal
          </button>
          <button
            type="button"
            @click="assignOperator"
            :disabled="isAssigning"
            :class="[
              'flex items-center gap-2 px-5 py-2.5 rounded-xl text-white text-xs font-black shadow-lg disabled:opacity-50 disabled:shadow-none transition-all',
              selectedOperatorId === ''
                ? 'bg-amber-600 hover:bg-amber-700 shadow-amber-600/20'
                : 'bg-indigo-600 hover:bg-indigo-700 shadow-indigo-500/20'
            ]"
          >
            <RefreshCw v-if="isAssigning" class="w-4 h-4 animate-spin" />
            <UserX v-else-if="selectedOperatorId === ''" class="w-4 h-4" />
            <UserCheck v-else class="w-4 h-4" />
            {{ isAssigning ? 'Menyimpan...' : (selectedOperatorId === '' ? 'Kosongkan Penugasan' : 'Konfirmasi Penugasan') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Upload Modal -->
    <div v-if="showUploadModal" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="!isUploading && (showUploadModal = false)"
      ></div>
      <div
        class="bg-white rounded-[32px] w-full max-w-md relative z-10 shadow-2xl p-8 animate-in fade-in zoom-in duration-300"
      >
        <div class="flex justify-between items-center mb-6">
          <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
            <Camera class="w-6 h-6 text-primary" /> Bukti Pemasangan
          </h3>
          <button
            v-if="!isUploading"
            @click="showUploadModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="space-y-6">
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-2">Toko Target</label>
            <div class="p-3 bg-slate-50 border border-slate-100 rounded-xl font-medium text-slate-600 text-sm">
              {{ selectedTask?.shop_name }}
            </div>
          </div>

          <div>
            <label class="block text-sm font-bold text-slate-700 mb-2 flex items-center gap-1.5">
              <Calendar class="w-4 h-4 text-emerald-600" /> Tanggal & Waktu Pemasangan
            </label>
            <input
              v-model="installationDateTime"
              type="datetime-local"
              class="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-xl text-xs sm:text-sm font-semibold text-slate-700 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all shadow-xs"
            />
            <p class="text-[11px] text-slate-400 mt-1 font-medium">
              Otomatis terisi waktu saat ini, sesuaikan jika pemasangan fisik dilakukan sebelumnya.
            </p>
          </div>

          <div>
            <label class="block text-sm font-bold text-slate-700 mb-2">Lokasi Pemasangan (GPS)</label>
            <div class="flex gap-3 items-center">
              <button 
                @click="getGeolocation"
                type="button"
                class="flex items-center gap-2 px-4 py-2 bg-blue-50 text-blue-600 border border-blue-100 hover:bg-blue-100 rounded-xl text-sm font-bold transition-colors"
              >
                <MapPin class="w-4 h-4" /> Ambil Lokasi Saat Ini
              </button>
              <div v-if="latitude && longitude" class="text-xs font-bold text-emerald-600 flex items-center gap-1">
                <CheckCircle2 class="w-4 h-4" /> Didapatkan
              </div>
            </div>
            <p v-if="!latitude || !longitude" class="text-xs text-rose-500 mt-2 font-medium flex items-center gap-1">
              <AlertTriangle class="w-3.5 h-3.5" /> Lokasi wajib didapatkan sebelum submit.
            </p>
          </div>

          <div>
            <label class="block text-sm font-bold text-slate-700 mb-2">Foto / Gambar (Wajib)</label>
            <input
              type="file"
              accept="image/*"
              capture="environment"
              @change="handleFileChange"
              class="w-full text-sm text-slate-500 file:mr-4 file:py-2.5 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-bold file:bg-primary/10 file:text-primary hover:file:bg-primary/20"
            />
            
            <div v-if="previewImage" class="mt-4 border-2 border-slate-200 border-dashed rounded-2xl overflow-hidden bg-slate-50">
              <img :src="previewImage" alt="Preview" class="w-full h-48 object-cover" />
            </div>
          </div>
          
          <button
            @click="submitUpload"
            :disabled="!selectedFile || !latitude || !longitude || isUploading"
            class="w-full bg-primary hover:bg-primary-focus disabled:bg-slate-200 disabled:text-slate-400 text-white py-4 rounded-2xl font-black transition-all shadow-lg flex items-center justify-center gap-2"
          >
            <RefreshCw v-if="isUploading" class="w-5 h-5 animate-spin" />
            {{ isUploading ? 'Mengirim...' : 'Upload & Selesaikan Pemasangan' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Preview Modal -->
    <div v-if="showPreviewModal" class="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/80 backdrop-blur-sm"
        @click="closePreview"
      ></div>
      <div class="bg-white rounded-[32px] w-full max-w-5xl h-[85vh] relative z-10 shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in duration-300">
        <div class="flex justify-between items-center p-6 border-b border-slate-200">
          <h3 class="text-lg font-black text-slate-900 flex items-center gap-2">
            <ImageIcon class="w-5 h-5 text-indigo-500" /> {{ previewTitle || 'Preview Bukti Pemasangan' }}
          </h3>
          <button
            @click="closePreview"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="flex-1 bg-slate-100 p-2 sm:p-4 flex items-center justify-center overflow-auto">
          <iframe
            v-if="isDriveUrl(previewUrl)"
            :src="previewUrl"
            class="w-full h-full rounded-2xl border-0 shadow-inner bg-white"
            allow="autoplay"
            allowfullscreen
          ></iframe>
          <img
            v-else
            :src="previewUrl"
            alt="Preview Image"
            class="max-w-full max-h-full object-contain rounded-xl shadow-md"
          />
        </div>
      </div>
    </div>

  </div>
</template>
