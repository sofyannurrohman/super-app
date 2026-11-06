<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { watchDebounced } from '@vueuse/core'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import apiClient from '@/api/axios'
import ExcelJS from 'exceljs'
import { saveAs } from 'file-saver'
import { toast } from 'vue-sonner'
import {
  TrendingUp,
  TrendingDown,
  Activity,
  Search,
  Filter,
  Package,
  ArrowRightLeft,
  RefreshCw,
  Calendar,
  ChevronDown,
  ChevronUp,
  BarChart3,
  FileSpreadsheet,
  Loader2,
  X,
  CheckCircle2,
  ShoppingBag,
  Layers,
  Download,
  Pencil,
} from 'lucide-vue-next'
import type { MarketplaceOpname } from '@/api/marketplace'
import { Card, CardContent } from '@/components/ui/card'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()

const getTodayDate = () => {
  const d = new Date()
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
  return d.toISOString().split('T')[0]
}

const searchQuery = ref('')
const selectedType = ref('ALL')
const selectedSourceType = ref<'ALL' | 'TIKTOK' | 'SHOPEE' | 'MANUAL'>('ALL')
const fromDate = ref(getTodayDate())
const toDate = ref(getTodayDate())
const expandedId = ref<string | null>(null)
const isExporting = ref(false)

const movementTypes = [
  { value: 'ALL', label: 'Semua Tipe', icon: Activity, color: 'text-slate-500', bg: 'bg-slate-100' },
  { value: 'STOCK_IN', label: 'Masuk', icon: TrendingUp, color: 'text-emerald-600', bg: 'bg-emerald-50' },
  { value: 'STOCK_OUT', label: 'Keluar', icon: TrendingDown, color: 'text-rose-600', bg: 'bg-rose-50' },
  { value: 'ADJUSTMENT', label: 'Penyesuaian', icon: ArrowRightLeft, color: 'text-violet-600', bg: 'bg-violet-50' },
]

const platformTypes = [
  { value: 'ALL', label: 'Semua Asal Platform', shortLabel: 'Semua', icon: Layers, tag: 'ALL', activeClass: 'bg-slate-900 text-white shadow-sm shadow-slate-300' },
  { value: 'TIKTOK', label: 'TikTok Shop', shortLabel: 'TikTok', icon: ShoppingBag, tag: 'TT', activeClass: 'bg-slate-950 text-white shadow-sm shadow-slate-400' },
  { value: 'SHOPEE', label: 'Shopee', shortLabel: 'Shopee', icon: ShoppingBag, tag: 'SP', activeClass: 'bg-gradient-to-r from-orange-500 to-amber-600 text-white shadow-sm shadow-orange-200' },
  { value: 'MANUAL', label: 'Manual / Non-Orderan', shortLabel: 'Manual', icon: ArrowRightLeft, tag: 'MN', activeClass: 'bg-slate-700 text-white shadow-sm shadow-slate-300' },
]

async function loadData() {
  const params: Record<string, any> = {
    page: marketplaceStore.movementsPage,
    limit: marketplaceStore.movementsLimit,
  }
  if (selectedType.value !== 'ALL') params.type = selectedType.value
  if (selectedSourceType.value !== 'ALL') params.source_type = selectedSourceType.value
  if (fromDate.value) params.from = fromDate.value
  if (toDate.value) params.to = toDate.value
  if (searchQuery.value.trim()) params.search = searchQuery.value.trim()
  await marketplaceStore.fetchStockMovements(params)
}

onMounted(() => {
  marketplaceStore.movementsPage = 1
  loadData()
  marketplaceStore.fetchStocks()
})

watch(() => masterDataStore.selectedBranchId, () => {
  marketplaceStore.movementsPage = 1
  loadData()
  marketplaceStore.fetchStocks()
})
watch([selectedType, selectedSourceType, fromDate, toDate], () => {
  marketplaceStore.movementsPage = 1
  loadData()
})
watchDebounced(searchQuery, () => {
  marketplaceStore.movementsPage = 1
  loadData()
}, { debounce: 500 })

const filteredMovements = computed(() => {
  return marketplaceStore.stockMovements || []
})

// Stats ringkasan (konsisten dan akurat sesuai seluruh filter dari backend)
const stats = computed(() => {
  const summary = marketplaceStore.movementsSummary
  if (summary && summary.total_transactions !== undefined) {
    return {
      totalIn: summary.total_in || 0,
      totalOut: summary.total_out || 0,
      totalAdj: summary.total_adj || 0,
      omzet: summary.total_omzet || 0,
      total: summary.total_transactions || marketplaceStore.movementsTotal || 0,
    }
  }
  const all = marketplaceStore.stockMovements || []
  const totalIn = all.filter(m => m.type === 'STOCK_IN').reduce((s, m) => s + m.quantity, 0)
  const totalOut = all.filter(m => m.type === 'STOCK_OUT').reduce((s, m) => s + m.quantity, 0)
  const totalAdj = all.filter(m => m.type === 'ADJUSTMENT').length
  const omzet = all.filter(m => m.type === 'STOCK_OUT').reduce((s, m) => s + (m.total_value || 0), 0)
  return { totalIn, totalOut, totalAdj, omzet, total: all.length }
})

function formatDateTime(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('id-ID', {
    day: '2-digit', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

function formatRupiah(val: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency', currency: 'IDR', maximumFractionDigits: 0,
  }).format(val)
}

function getTypeConfig(type: string) {
  switch (type) {
    case 'STOCK_IN':
      return { label: 'Barang Masuk', textColor: 'text-emerald-700', bgColor: 'bg-emerald-100', borderColor: 'border-emerald-200', dotColor: 'bg-emerald-500', icon: TrendingUp }
    case 'STOCK_OUT':
      return { label: 'Barang Keluar', textColor: 'text-rose-700', bgColor: 'bg-rose-100', borderColor: 'border-rose-200', dotColor: 'bg-rose-500', icon: TrendingDown }
    case 'ADJUSTMENT':
      return { label: 'Penyesuaian', textColor: 'text-violet-700', bgColor: 'bg-violet-100', borderColor: 'border-violet-200', dotColor: 'bg-violet-500', icon: ArrowRightLeft }
    default:
      return { label: type, textColor: 'text-slate-700', bgColor: 'bg-slate-100', borderColor: 'border-slate-200', dotColor: 'bg-slate-400', icon: Activity }
  }
}

function isShopeeOrderCode(str?: string): boolean {
  if (!str) return false
  const clean = str.replace(/^RETUR-/i, '').trim()
  return /^[0-9]{6}[0-9A-Za-z]{6,14}$/i.test(clean) && /[A-Za-z]/.test(clean)
}

function getPlatformInfo(m: MarketplaceOpname) {
  const src = (m.source_type || '').toUpperCase()
  const refStr = (m.reference || '').toUpperCase()
  const notesStr = (m.notes || '').toUpperCase()
  
  if (
    src === 'SHOPEE' || 
    refStr.includes('SHOPEE') || 
    notesStr.includes('SHOPEE') ||
    isShopeeOrderCode(m.reference) ||
    isShopeeOrderCode(m.marketplace_order_id)
  ) {
    return {
      key: 'SHOPEE',
      label: 'Shopee',
      shortLabel: 'Shopee',
      tag: 'SP',
      badgeClass: 'bg-gradient-to-r from-orange-500 to-amber-600 text-white border-orange-400',
      dotClass: 'bg-white',
      isOrder: true,
    }
  }

  if (src === 'TIKTOK' || refStr.includes('TIKTOK') || notesStr.includes('TIKTOK') || (m.marketplace_order_id && !refStr.startsWith('RETUR-'))) {
    return {
      key: 'TIKTOK',
      label: 'TikTok Shop',
      shortLabel: 'TikTok',
      tag: 'TT',
      badgeClass: 'bg-slate-900 text-white border-slate-900',
      dotClass: 'bg-cyan-400',
      isOrder: true,
    }
  }

  if (refStr.startsWith('RETUR-') || src === 'ORDER_RETURN') {
    return {
      key: 'RETURN',
      label: 'Retur Pesanan',
      shortLabel: 'Retur',
      tag: 'RT',
      badgeClass: 'bg-amber-100 text-amber-800 border-amber-300',
      dotClass: 'bg-amber-500',
      isOrder: true,
    }
  }

  return {
    key: 'MANUAL',
    label: 'Manual / Opname',
    shortLabel: 'Manual',
    tag: 'MN',
    badgeClass: 'bg-slate-100 text-slate-600 border-slate-200',
    dotClass: 'bg-slate-400',
    isOrder: false,
  }
}

function toggleExpand(id: string) {
  expandedId.value = expandedId.value === id ? null : id
}

function setDateRangePreset(preset: 'today' | '7d' | '30d' | 'all') {
  if (preset === 'today') {
    fromDate.value = getTodayDate()
    toDate.value = getTodayDate()
  } else if (preset === '7d') {
    const d = new Date()
    d.setDate(d.getDate() - 7)
    fromDate.value = d.toISOString().split('T')[0]
    toDate.value = getTodayDate()
  } else if (preset === '30d') {
    const d = new Date()
    d.setDate(d.getDate() - 30)
    fromDate.value = d.toISOString().split('T')[0]
    toDate.value = getTodayDate()
  } else if (preset === 'all') {
    fromDate.value = ''
    toDate.value = ''
  }
}

function resetFilters() {
  searchQuery.value = ''
  selectedType.value = 'ALL'
  selectedSourceType.value = 'ALL'
  fromDate.value = getTodayDate()
  toDate.value = getTodayDate()
}

function getCurrentLocalDateTime(): string {
  const now = new Date()
  now.setMinutes(now.getMinutes() - now.getTimezoneOffset())
  return now.toISOString().slice(0, 16)
}

// ── EDIT MOVEMENT STATE & LOGIC ───────────────────────────────────────────────
const showEditModal = ref(false)
const isUpdating = ref(false)
const editForm = ref({
  id: '',
  stock_id: '',
  product_name: '',
  sku: '',
  type: 'STOCK_IN' as 'STOCK_IN' | 'STOCK_OUT' | 'ADJUSTMENT',
  source_type: '',
  date: '',
  quantity: 0,
  reference: '',
  notes: '',
})

function openEditMovement(m: MarketplaceOpname) {
  let dateFormatted = ''
  if (m.created_at) {
    const d = new Date(m.created_at)
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
    dateFormatted = d.toISOString().slice(0, 16)
  }

  editForm.value = {
    id: m.id,
    stock_id: m.marketplace_stock_id,
    product_name: m.marketplace_stock?.product_name || 'Produk',
    sku: m.marketplace_stock?.sku || '',
    type: m.type,
    source_type: (m as any).source_type || 'MANUAL',
    date: dateFormatted,
    quantity: m.quantity,
    reference: m.reference || '',
    notes: m.notes || '',
  }
  showEditModal.value = true
}

async function submitEditMovement() {
  if (!editForm.value.id) return
  isUpdating.value = true
  try {
    await marketplaceStore.updateStockMovement(editForm.value.id, {
      date: editForm.value.date,
      quantity: editForm.value.quantity,
      reference: editForm.value.reference,
      notes: editForm.value.notes,
    })
    toast.success('Data pergerakan stok berhasil diperbarui')
    showEditModal.value = false
    await loadData()
  } catch (err: any) {
    toast.error('Gagal memperbarui data: ' + (err.response?.data?.error || err.message))
  } finally {
    isUpdating.value = false
  }
}

// ── EXPORT MODAL STATE & LOGIC ───────────────────────────────────────────────
const showExportModal = ref(false)
const exportScope = ref<'TIKTOK_ORDERS' | 'SHOPEE_ORDERS' | 'ALL_ORDERS' | 'ALL' | 'STOCK_IN' | 'ADJUSTMENT'>('TIKTOK_ORDERS')
const exportFromDate = ref(getTodayDate())
const exportToDate = ref(getTodayDate())
const exportDatePreset = ref<'today' | '7d' | '30d' | 'this_month' | 'all' | 'custom'>('today')
const exportSearchQuery = ref('')
const exportProgress = ref(0)
const exportProgressText = ref('Menyiapkan export...')

function openExportModal() {
  if (selectedSourceType.value === 'SHOPEE') {
    exportScope.value = 'SHOPEE_ORDERS'
  } else if (selectedSourceType.value === 'TIKTOK') {
    exportScope.value = 'TIKTOK_ORDERS'
  } else {
    exportScope.value = 'ALL'
  }
  exportFromDate.value = fromDate.value || getTodayDate()
  exportToDate.value = toDate.value || getTodayDate()
  exportSearchQuery.value = searchQuery.value || ''

  if (fromDate.value === getTodayDate() && toDate.value === getTodayDate()) {
    exportDatePreset.value = 'today'
  } else if (!fromDate.value && !toDate.value) {
    exportDatePreset.value = 'all'
  } else {
    exportDatePreset.value = 'custom'
  }

  showExportModal.value = true
}

function setExportDatePreset(preset: 'today' | '7d' | '30d' | 'this_month' | 'all') {
  exportDatePreset.value = preset
  const today = new Date()
  if (preset === 'today') {
    exportFromDate.value = getTodayDate()
    exportToDate.value = getTodayDate()
  } else if (preset === '7d') {
    const d = new Date()
    d.setDate(d.getDate() - 7)
    exportFromDate.value = d.toISOString().split('T')[0]
    exportToDate.value = getTodayDate()
  } else if (preset === '30d') {
    const d = new Date()
    d.setDate(d.getDate() - 30)
    exportFromDate.value = d.toISOString().split('T')[0]
    exportToDate.value = getTodayDate()
  } else if (preset === 'this_month') {
    const startOfMonth = new Date(today.getFullYear(), today.getMonth(), 1)
    startOfMonth.setMinutes(startOfMonth.getMinutes() - startOfMonth.getTimezoneOffset())
    exportFromDate.value = startOfMonth.toISOString().split('T')[0]
    exportToDate.value = getTodayDate()
  } else if (preset === 'all') {
    exportFromDate.value = ''
    exportToDate.value = ''
  }
}

async function executeExport() {
  if (isExporting.value) return
  isExporting.value = true

  try {
    // Ambil data bertahap agar export tidak terpotong pada batas 10.000 record.
    // Total dikembalikan oleh backend berdasarkan seluruh filter yang aktif.
    const pageSize = 1000
    const params: Record<string, any> = {
      limit: pageSize,
      page: 1,
    }
    if (masterDataStore.selectedBranchId) params.branch_id = masterDataStore.selectedBranchId
    
    // Type and source filter query for backend
    if (exportScope.value === 'STOCK_IN') {
      params.type = 'STOCK_IN'
    } else if (exportScope.value === 'ADJUSTMENT') {
      params.type = 'ADJUSTMENT'
    } else if (exportScope.value === 'TIKTOK_ORDERS') {
      params.type = 'STOCK_OUT'
      params.source_type = 'TIKTOK'
    } else if (exportScope.value === 'SHOPEE_ORDERS') {
      params.type = 'STOCK_OUT'
      params.source_type = 'SHOPEE'
    } else if (exportScope.value === 'ALL_ORDERS') {
      params.type = 'STOCK_OUT'
    }

    if (exportFromDate.value) params.from = exportFromDate.value
    if (exportToDate.value) params.to = exportToDate.value
    if (exportSearchQuery.value.trim()) params.search = exportSearchQuery.value.trim()

    const exportMovements: any[] = []
    let totalToFetch = 0
    let page = 1

    do {
      exportProgressText.value = page === 1
        ? 'Mengambil data transaksi...'
        : `Mengambil halaman ${page}...`
      const res = await apiClient.get('/marketplace/opnames', {
        params: { ...params, page },
      })
      const pageData: any[] = res.data?.data || []
      exportMovements.push(...pageData)
      totalToFetch = Number(res.data?.total || 0)
      exportProgress.value = totalToFetch > 0
        ? Math.min(100, Math.round((exportMovements.length / totalToFetch) * 100))
        : 100
      exportProgressText.value = totalToFetch > 0
        ? `Mengambil data ${exportMovements.length.toLocaleString('id-ID')} dari ${totalToFetch.toLocaleString('id-ID')}...`
        : 'Tidak ada data transaksi yang sesuai filter.'
      page += 1

      // Berhenti jika sudah semua data diambil atau backend mengembalikan halaman kosong.
      if (pageData.length === 0 || exportMovements.length >= totalToFetch) break
    } while (exportMovements.length < totalToFetch)

    exportProgress.value = 100
    exportProgressText.value = 'Menyusun file Excel...'

    const branchName =
      masterDataStore.branches.find((b: any) => b.id === masterDataStore.selectedBranchId)?.name ||
      'Semua Cabang'

    let scopeTitle = 'Laporan Mutasi Stok'
    let scopeLabel = 'Semua Data Mutasi'
    let filePrefix = 'mutasi-stok-semua'

    if (exportScope.value === 'TIKTOK_ORDERS') {
      scopeTitle = 'LAPORAN BARANG KELUAR ORDERAN TIKTOK SHOP'
      scopeLabel = 'Khusus Orderan Keluar TikTok Shop'
      filePrefix = 'mutasi-orderan-tiktok'
    } else if (exportScope.value === 'SHOPEE_ORDERS') {
      scopeTitle = 'LAPORAN BARANG KELUAR ORDERAN SHOPEE'
      scopeLabel = 'Khusus Orderan Keluar Shopee'
      filePrefix = 'mutasi-orderan-shopee'
    } else if (exportScope.value === 'ALL_ORDERS') {
      scopeTitle = 'LAPORAN BARANG KELUAR SEMUA ORDERAN MARKETPLACE'
      scopeLabel = 'Semua Orderan Keluar Marketplace (TikTok & Shopee)'
      filePrefix = 'mutasi-orderan-marketplace'
    } else if (exportScope.value === 'STOCK_IN') {
      scopeTitle = 'LAPORAN REKAPITULASI BARANG MASUK'
      scopeLabel = 'Khusus Barang Masuk'
      filePrefix = 'mutasi-barang-masuk'
    } else if (exportScope.value === 'ADJUSTMENT') {
      scopeTitle = 'LAPORAN REKAPITULASI PENYESUAIAN STOK'
      scopeLabel = 'Khusus Penyesuaian / Opname'
      filePrefix = 'mutasi-penyesuaian-stok'
    } else {
      scopeTitle = 'LAPORAN REKAPITULASI PERGERAKAN STOK LENGKAP'
      scopeLabel = 'Semua Data Mutasi'
      filePrefix = 'mutasi-stok-lengkap'
    }

    // Inisialisasi Workbook ExcelJS
    const workbook = new ExcelJS.Workbook()
    workbook.creator = 'HRIS Karsa HR'
    workbook.lastModifiedBy = 'HRIS Karsa HR'
    workbook.created = new Date()
    workbook.modified = new Date()

    // ─────────────────────────────────────────────────────────────────────────
    // 📊 TAB 1: DATA STOK & REKAPITULASI (RINGKASAN PER ITEM)
    // ─────────────────────────────────────────────────────────────────────────
    const sheet1 = workbook.addWorksheet('Data Stok & Rekapitulasi', {
      pageSetup: { fitToPage: true, orientation: 'landscape', paperSize: 9 },
      views: [{ state: 'frozen', ySplit: 5 }],
    })

    // Header Title
    sheet1.mergeCells('A1:J1')
    const title1 = sheet1.getCell('A1')
    title1.value = scopeTitle
    title1.font = { bold: true, size: 14, color: { argb: 'FFFFFFFF' } }
    title1.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E293B' } }
    title1.alignment = { horizontal: 'center', vertical: 'middle' }
    sheet1.getRow(1).height = 36

    // Subtitle Info
    sheet1.mergeCells('A2:J2')
    const sub1 = sheet1.getCell('A2')
    sub1.value = `Cabang: ${branchName}   |   Periode: ${exportFromDate.value || 'Semua'} s/d ${exportToDate.value || 'Semua'}   |   Kategori Data: ${scopeLabel}   |   Waktu Export: ${new Date().toLocaleString('id-ID')}`
    sub1.font = { italic: true, size: 9, color: { argb: 'FF475569' } }
    sub1.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF1F5F9' } }
    sub1.alignment = { horizontal: 'center', vertical: 'middle' }
    sheet1.getRow(2).height = 22

    sheet1.addRow({}) // Empty spacer
    sheet1.getRow(3).height = 8

    // Kolom Header Sheet 1
    sheet1.columns = [
      { key: 'no', width: 6 },
      { key: 'sku', width: 20 },
      { key: 'product_name', width: 36 },
      { key: 'category', width: 20 },
      { key: 'total_in', width: 16 },
      { key: 'total_out', width: 16 },
      { key: 'total_adj', width: 16 },
      { key: 'current_stock', width: 16 },
      { key: 'price', width: 18 },
      { key: 'total_omzet', width: 22 },
    ]

    const headers1 = [
      'No',
      'SKU',
      'Nama Barang / Item',
      'Brand / Kategori',
      'Total Masuk (+)',
      'Total Keluar (-)',
      'Penyesuaian (±)',
      'Stok Terkini',
      'Harga Satuan (Rp)',
      'Total Nilai Keluar (Rp)',
    ]
    const headerRow1 = sheet1.getRow(4)
    headers1.forEach((label, i) => {
      const cell = headerRow1.getCell(i + 1)
      cell.value = label
      cell.font = { bold: true, size: 10, color: { argb: 'FFFFFFFF' } }
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF0F172A' } }
      cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true }
      cell.border = {
        top: { style: 'thin', color: { argb: 'FF334155' } },
        bottom: { style: 'medium', color: { argb: 'FF0F172A' } },
        left: { style: 'thin', color: { argb: 'FF334155' } },
        right: { style: 'thin', color: { argb: 'FF334155' } },
      }
    })
    headerRow1.height = 28

    // Agregasi Data Per Barang
    const itemMap = new Map<string, {
      sku: string
      name: string
      category: string
      totalIn: number
      totalOut: number
      totalAdj: number
      currentStock: number
      price: number
      totalOmzet: number
    }>()

    exportMovements.forEach((m) => {
      const stock = m.marketplace_stock
      const key = stock?.id || stock?.sku || m.marketplace_stock_id || 'UNKNOWN'
      if (!itemMap.has(key)) {
        itemMap.set(key, {
          sku: stock?.sku || '-',
          name: stock?.product_name || 'Produk tidak diketahui',
          category: stock?.product?.brand || stock?.product?.category || 'Umum',
          totalIn: 0,
          totalOut: 0,
          totalAdj: 0,
          currentStock: stock?.quantity ?? 0,
          price: stock?.price || m.price || 0,
          totalOmzet: 0,
        })
      }

      const item = itemMap.get(key)!
      if (m.type === 'STOCK_IN') {
        item.totalIn += m.quantity
      } else if (m.type === 'STOCK_OUT') {
        item.totalOut += m.quantity
        item.totalOmzet += (m.total_value || (m.quantity * (m.price || item.price || 0)))
      } else if (m.type === 'ADJUSTMENT') {
        item.totalAdj += 1
      }
    })

    const summaryItems = Array.from(itemMap.values())

    let grandTotalIn = 0
    let grandTotalOut = 0
    let grandTotalAdj = 0
    let grandTotalCurrentStock = 0
    let grandTotalOmzet = 0

    if (summaryItems.length === 0) {
      sheet1.mergeCells('A5:J5')
      const emptyCell = sheet1.getCell('A5')
      emptyCell.value = 'Tidak ada data pergerakan stok pada filter dan periode ini.'
      emptyCell.alignment = { horizontal: 'center', vertical: 'middle' }
      emptyCell.font = { italic: true, color: { argb: 'FF94A3B8' } }
      sheet1.getRow(5).height = 28
    } else {
      summaryItems.forEach((item, idx) => {
        grandTotalIn += item.totalIn
        grandTotalOut += item.totalOut
        grandTotalAdj += item.totalAdj
        grandTotalCurrentStock += item.currentStock
        grandTotalOmzet += item.totalOmzet

        const isEven = idx % 2 === 0
        const rowBg = isEven ? 'FFFFFFFF' : 'FFF8FAFC'

        const row = sheet1.addRow({
          no: idx + 1,
          sku: item.sku,
          product_name: item.name,
          category: item.category,
          total_in: item.totalIn,
          total_out: item.totalOut,
          total_adj: item.totalAdj,
          current_stock: item.currentStock,
          price: item.price,
          total_omzet: item.totalOmzet,
        })

        row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: rowBg } }
          cell.border = {
            top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          }

          if (colNumber === 1) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
          } else if (colNumber === 2) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            cell.font = { bold: true, size: 9 }
          } else if (colNumber === 3) {
            cell.alignment = { horizontal: 'left', vertical: 'middle' }
            cell.font = { bold: true, size: 10 }
          } else if (colNumber === 4) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            cell.font = { size: 9, color: { argb: 'FF475569' } }
          } else if (colNumber === 5) {
            cell.numFmt = '#,##0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.font = { bold: true, color: { argb: 'FF059669' } }
          } else if (colNumber === 6) {
            cell.numFmt = '#,##0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.font = { bold: true, color: { argb: 'FFE11D48' } }
          } else if (colNumber === 7) {
            cell.numFmt = '#,##0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.font = { bold: true, color: { argb: 'FF7C3AED' } }
          } else if (colNumber === 8) {
            cell.numFmt = '#,##0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.font = { bold: true }
          } else if (colNumber === 9) {
            cell.numFmt = '[$Rp-421] #,##0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
          } else if (colNumber === 10) {
            cell.numFmt = '[$Rp-421] #,##0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.font = { bold: true, color: { argb: 'FF0F172A' } }
          }
        })
        row.height = 24
      })

      // Grand Total Row Sheet 1
      const totalRow1 = sheet1.addRow({
        no: '',
        sku: '',
        product_name: 'GRAND TOTAL',
        category: '',
        total_in: grandTotalIn,
        total_out: grandTotalOut,
        total_adj: grandTotalAdj,
        current_stock: grandTotalCurrentStock,
        price: '',
        total_omzet: grandTotalOmzet,
      })

      totalRow1.eachCell({ includeEmpty: true }, (cell, colNumber) => {
        cell.font = { bold: true, size: 10, color: { argb: 'FF0F172A' } }
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
        cell.border = {
          top: { style: 'medium', color: { argb: 'FF0F172A' } },
          bottom: { style: 'double', color: { argb: 'FF0F172A' } },
          left: { style: 'thin', color: { argb: 'FFCBD5E1' } },
          right: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        }

        if (colNumber === 3) {
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
        } else if (colNumber === 5 || colNumber === 6 || colNumber === 7 || colNumber === 8) {
          cell.numFmt = '#,##0'
          cell.alignment = { horizontal: 'right', vertical: 'middle' }
        } else if (colNumber === 10) {
          cell.numFmt = '[$Rp-421] #,##0'
          cell.alignment = { horizontal: 'right', vertical: 'middle' }
        }
      })
      totalRow1.height = 26
    }

    // ─────────────────────────────────────────────────────────────────────────
    // 📦 TAB 2: DETAIL TRANSAKSI MUTASI
    // ─────────────────────────────────────────────────────────────────────────
    const sheet2 = workbook.addWorksheet('Detail Transaksi Mutasi', {
      pageSetup: { fitToPage: true, orientation: 'landscape', paperSize: 9 },
      views: [{ state: 'frozen', ySplit: 5 }],
    })

    sheet2.mergeCells('A1:M1')
    const title2 = sheet2.getCell('A1')
    title2.value = `LOG DETAIL TRANSAKSI ${exportScope.value === 'TIKTOK_ORDERS' ? 'BARANG KELUAR ORDERAN TIKTOK / MARKETPLACE' : 'MUTASI BARANG'}`
    title2.font = { bold: true, size: 14, color: { argb: 'FFFFFFFF' } }
    title2.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF0F766E' } }
    title2.alignment = { horizontal: 'center', vertical: 'middle' }
    sheet2.getRow(1).height = 36

    sheet2.mergeCells('A2:M2')
    const sub2 = sheet2.getCell('A2')
    sub2.value = `Cabang: ${branchName}   |   Periode: ${exportFromDate.value || 'Semua'} s/d ${exportToDate.value || 'Semua'}   |   Total Transaksi: ${exportMovements.length} Record   |   Diekspor: ${new Date().toLocaleString('id-ID')}`
    sub2.font = { italic: true, size: 9, color: { argb: 'FF134E4A' } }
    sub2.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF0FDFA' } }
    sub2.alignment = { horizontal: 'center', vertical: 'middle' }
    sheet2.getRow(2).height = 22

    sheet2.addRow({}) // Empty spacer
    sheet2.getRow(3).height = 8

    sheet2.columns = [
      { key: 'no', width: 6 },
      { key: 'created_at', width: 20 },
      { key: 'type', width: 16 },
      { key: 'sku', width: 18 },
      { key: 'product_name', width: 34 },
      { key: 'quantity', width: 14 },
      { key: 'qty_before', width: 14 },
      { key: 'qty_after', width: 14 },
      { key: 'price', width: 18 },
      { key: 'total_value', width: 20 },
      { key: 'reference', width: 24 },
      { key: 'notes', width: 30 },
      { key: 'created_by', width: 18 },
    ]

    const headers2 = [
      'No',
      'Tanggal & Waktu',
      'Tipe Mutasi',
      'SKU',
      'Nama Barang / Produk',
      'Perubahan (Qty)',
      'Stok Sebelum',
      'Stok Sesudah',
      'Harga Satuan (Rp)',
      'Total Nilai (Rp)',
      'No. Referensi / Order ID',
      'Catatan / Keterangan',
      'Diproses Oleh',
    ]

    const headerRow2 = sheet2.getRow(4)
    headers2.forEach((label, i) => {
      const cell = headerRow2.getCell(i + 1)
      cell.value = label
      cell.font = { bold: true, size: 10, color: { argb: 'FFFFFFFF' } }
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF134E4A' } }
      cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true }
      cell.border = {
        top: { style: 'thin', color: { argb: 'FF115E59' } },
        bottom: { style: 'medium', color: { argb: 'FF042F2E' } },
        left: { style: 'thin', color: { argb: 'FF115E59' } },
        right: { style: 'thin', color: { argb: 'FF115E59' } },
      }
    })
    headerRow2.height = 28

    if (exportMovements.length === 0) {
      sheet2.mergeCells('A5:M5')
      const emptyCell = sheet2.getCell('A5')
      emptyCell.value = 'Tidak ada transaksi mutasi yang tercatat untuk filter dan periode ini.'
      emptyCell.alignment = { horizontal: 'center', vertical: 'middle' }
      emptyCell.font = { italic: true, color: { argb: 'FF94A3B8' } }
      sheet2.getRow(5).height = 28
    } else {
      let logTotalIn = 0
      let logTotalOut = 0
      let logTotalValue = 0

      exportMovements.forEach((m, idx) => {
        const typeLabel =
          m.type === 'STOCK_IN'
            ? 'Masuk (+)'
            : m.type === 'STOCK_OUT'
              ? 'Keluar (-)'
              : 'Penyesuaian (±)'

        if (m.type === 'STOCK_IN') logTotalIn += m.quantity
        if (m.type === 'STOCK_OUT') logTotalOut += m.quantity
        logTotalValue += m.total_value || 0

        const typeColor =
          m.type === 'STOCK_IN'
            ? 'FF059669'
            : m.type === 'STOCK_OUT'
              ? 'FFE11D48'
              : 'FF7C3AED'

        const typeBg =
          m.type === 'STOCK_IN'
            ? 'FFECFDF5'
            : m.type === 'STOCK_OUT'
              ? 'FFFFF1F2'
              : 'FFF5F3FF'

        const isEven = idx % 2 === 0
        const rowBg = isEven ? 'FFFFFFFF' : 'FFF8FAFC'

        const row = sheet2.addRow({
          no: idx + 1,
          created_at: formatDateTime(m.created_at),
          type: typeLabel,
          sku: m.marketplace_stock?.sku || '-',
          product_name: m.marketplace_stock?.product_name || 'Produk tidak diketahui',
          quantity: m.type === 'STOCK_OUT' ? -m.quantity : m.quantity,
          qty_before: m.quantity_before,
          qty_after: m.quantity_after,
          price: m.price || 0,
          total_value: m.total_value || 0,
          reference: m.reference || '-',
          notes: m.notes || '-',
          created_by: m.created_by || 'Admin',
        })

        row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
          cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: rowBg } }
          cell.border = {
            top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
            right: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          }

          if (colNumber === 1 || colNumber === 2) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            cell.font = { size: 9 }
          } else if (colNumber === 3) {
            cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: typeBg } }
            cell.font = { bold: true, size: 9, color: { argb: typeColor } }
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
          } else if (colNumber === 4) {
            cell.alignment = { horizontal: 'center', vertical: 'middle' }
            cell.font = { bold: true, size: 9 }
          } else if (colNumber === 5) {
            cell.alignment = { horizontal: 'left', vertical: 'middle' }
            cell.font = { bold: true, size: 10 }
          } else if (colNumber === 6) {
            cell.numFmt = '#,##0;[Red]-#,##0;0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.font = { bold: true, color: { argb: typeColor } }
          } else if (colNumber === 7 || colNumber === 8) {
            cell.numFmt = '#,##0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.font = { size: 9 }
          } else if (colNumber === 9 || colNumber === 10) {
            cell.numFmt = '[$Rp-421] #,##0'
            cell.alignment = { horizontal: 'right', vertical: 'middle' }
            cell.font = { size: 9, bold: colNumber === 10 }
          } else if (colNumber === 11 || colNumber === 12 || colNumber === 13) {
            cell.alignment = { horizontal: 'left', vertical: 'middle' }
            cell.font = { size: 9, color: { argb: 'FF475569' } }
          }
        })
        row.height = 22
      })

      // Grand Total Row Sheet 2
      const totalRow2 = sheet2.addRow({
        no: '',
        created_at: '',
        type: 'TOTAL TRANSAKSI',
        sku: `${exportMovements.length} Record`,
        product_name: `Masuk: +${logTotalIn} | Keluar: -${logTotalOut}`,
        quantity: logTotalIn - logTotalOut,
        qty_before: '',
        qty_after: '',
        price: '',
        total_value: logTotalValue,
        reference: '',
        notes: '',
        created_by: '',
      })

      totalRow2.eachCell({ includeEmpty: true }, (cell, colNumber) => {
        cell.font = { bold: true, size: 9, color: { argb: 'FF0F172A' } }
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
        cell.border = {
          top: { style: 'medium', color: { argb: 'FF0F172A' } },
          bottom: { style: 'double', color: { argb: 'FF0F172A' } },
          left: { style: 'thin', color: { argb: 'FFCBD5E1' } },
          right: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        }

        if (colNumber === 3) {
          cell.alignment = { horizontal: 'center', vertical: 'middle' }
        } else if (colNumber === 6) {
          cell.numFmt = '#,##0;[Red]-#,##0;0'
          cell.alignment = { horizontal: 'right', vertical: 'middle' }
        } else if (colNumber === 10) {
          cell.numFmt = '[$Rp-421] #,##0'
          cell.alignment = { horizontal: 'right', vertical: 'middle' }
        }
      })
      totalRow2.height = 26
    }

    // ── TRIGGER FILE DOWNLOAD ────────────────────────────────────────────────
    const buffer = await workbook.xlsx.writeBuffer()
    const blob = new Blob([buffer], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })

    const fileNameDate = `${exportFromDate.value || 'awal'}_sd_${exportToDate.value || 'akhir'}`
    const fileName = `${filePrefix}_${fileNameDate}.xlsx`

    saveAs(blob, fileName)
    showExportModal.value = false
    toast.success(`Berhasil mengekspor 2 Tab data ke ${fileName}`)
  } catch (error: any) {
    console.error('Failed to export stock movements to Excel:', error)
    toast.error('Gagal mengekspor data ke Excel: ' + (error.message || String(error)))
  } finally {
    isExporting.value = false
    exportProgress.value = 0
    exportProgressText.value = 'Menyiapkan export...'
  }
}
</script>

<template>
  <div class="space-y-5 animate-in fade-in duration-500">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-3">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-slate-800 to-slate-900 shadow-lg shadow-slate-200">
            <BarChart3 class="h-6 w-6 text-white" />
          </div>
          Riwayat Pergerakan Stok
        </h1>
        <p class="text-slate-500 text-sm font-medium mt-0.5">
          Histori lengkap barang masuk, keluar, dan penyesuaian stok
        </p>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        <!-- Export Excel Button -->
        <button
          @click="openExportModal"
          :disabled="isExporting"
          class="flex items-center gap-2 px-4 py-2 text-sm font-bold bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white rounded-xl shadow-sm transition-all duration-200 cursor-pointer"
          title="Buka dialog opsi export data ke Excel"
        >
          <FileSpreadsheet class="h-4 w-4 text-emerald-100" />
          <span>Export Excel</span>
        </button>

        <!-- Refresh Button -->
        <button
          @click="loadData"
          class="flex items-center gap-2 px-4 py-2 text-sm font-bold bg-white border border-slate-200 rounded-xl hover:bg-slate-50 hover:border-slate-300 transition-all shadow-sm"
        >
          <RefreshCw class="h-4 w-4" :class="marketplaceStore.isLoading ? 'animate-spin' : ''" />
          Refresh
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <Card class="border-none shadow-sm bg-gradient-to-br from-emerald-50 to-teal-50">
        <CardContent class="p-4">
          <div class="flex items-center gap-2 mb-1">
            <TrendingUp class="h-4 w-4 text-emerald-600" />
            <span class="text-[11px] font-bold uppercase tracking-wider text-emerald-600">Total Masuk</span>
          </div>
          <p class="text-2xl font-black text-emerald-700">{{ stats.totalIn.toLocaleString('id-ID') }}</p>
          <p class="text-[11px] text-emerald-500 font-medium">unit masuk gudang</p>
        </CardContent>
      </Card>
      <Card class="border-none shadow-sm bg-gradient-to-br from-rose-50 to-red-50">
        <CardContent class="p-4">
          <div class="flex items-center gap-2 mb-1">
            <TrendingDown class="h-4 w-4 text-rose-600" />
            <span class="text-[11px] font-bold uppercase tracking-wider text-rose-600">Total Keluar</span>
          </div>
          <p class="text-2xl font-black text-rose-700">{{ stats.totalOut.toLocaleString('id-ID') }}</p>
          <p class="text-[11px] text-rose-500 font-medium">unit keluar</p>
        </CardContent>
      </Card>
      <Card class="border-none shadow-sm bg-gradient-to-br from-violet-50 to-purple-50">
        <CardContent class="p-4">
          <div class="flex items-center gap-2 mb-1">
            <ArrowRightLeft class="h-4 w-4 text-violet-600" />
            <span class="text-[11px] font-bold uppercase tracking-wider text-violet-600">Penyesuaian</span>
          </div>
          <p class="text-2xl font-black text-violet-700">{{ stats.totalAdj }}</p>
          <p class="text-[11px] text-violet-500 font-medium">kali penyesuaian</p>
        </CardContent>
      </Card>
      <Card class="border-none shadow-sm bg-gradient-to-br from-slate-50 to-slate-100">
        <CardContent class="p-4">
          <div class="flex items-center gap-2 mb-1">
            <Activity class="h-4 w-4 text-slate-700" />
            <span class="text-[11px] font-bold uppercase tracking-wider text-slate-700">Total Transaksi</span>
          </div>
          <p class="text-2xl font-black text-slate-800">{{ stats.total }}</p>
          <p class="text-[11px] text-slate-500 font-medium">total pergerakan</p>
        </CardContent>
      </Card>
    </div>

    <!-- Filter Bar -->
    <Card class="border-none shadow-sm">
      <CardContent class="p-4">
        <div class="flex flex-col gap-3.5">
          <!-- Top Row: Platform & Type Filter Tabs -->
          <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 border-b border-slate-100 pb-3">
            <!-- Platform Selector -->
            <div class="flex flex-col gap-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-slate-400">Asal Orderan / Platform:</span>
              <div class="flex flex-wrap gap-1.5">
                <button
                  v-for="p in platformTypes"
                  :key="p.value"
                  @click="selectedSourceType = p.value as any"
                  class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold border transition-all cursor-pointer"
                  :class="selectedSourceType === p.value
                    ? p.activeClass
                    : 'bg-slate-50 text-slate-600 border-slate-200 hover:border-slate-300 hover:bg-slate-100'"
                >
                  <component :is="p.icon" class="h-3.5 w-3.5" />
                  <span>{{ p.label }}</span>
                </button>
              </div>
            </div>

            <!-- Type Filter Tabs -->
            <div class="flex flex-col gap-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-slate-400">Tipe Mutasi:</span>
              <div class="flex flex-wrap gap-1.5">
                <button
                  v-for="t in movementTypes"
                  :key="t.value"
                  @click="selectedType = t.value"
                  class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold border transition-all cursor-pointer"
                  :class="selectedType === t.value
                    ? 'bg-slate-900 text-white border-slate-900 shadow-sm shadow-slate-200'
                    : 'bg-slate-50 text-slate-600 border-slate-200 hover:border-slate-300 hover:bg-slate-100'"
                >
                  <component :is="t.icon" class="h-3.5 w-3.5" />
                  {{ t.label }}
                </button>
              </div>
            </div>
          </div>

          <!-- Bottom Row: Search & Date Filter -->
          <div class="flex flex-col lg:flex-row gap-2">
            <div class="relative flex-1">
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Cari nama produk, SKU, no. referensi / Order ID, catatan..."
                class="w-full pl-9 pr-4 py-2 text-sm font-medium bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-slate-200 focus:border-slate-400 transition-all"
              />
            </div>
            <!-- Date Inputs & Presets -->
            <div class="flex flex-wrap items-center gap-2">
              <div class="flex items-center gap-1.5 bg-slate-100 p-1 rounded-xl">
                <button
                  type="button"
                  @click="setDateRangePreset('today')"
                  class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all cursor-pointer"
                  :class="fromDate === getTodayDate() && toDate === getTodayDate() ? 'bg-white shadow-sm text-slate-900' : 'text-slate-500 hover:text-slate-800'"
                >
                  Hari Ini
                </button>
                <button
                  type="button"
                  @click="setDateRangePreset('7d')"
                  class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all text-slate-500 hover:text-slate-800 cursor-pointer"
                >
                  7 Hari
                </button>
                <button
                  type="button"
                  @click="setDateRangePreset('30d')"
                  class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all text-slate-500 hover:text-slate-800 cursor-pointer"
                >
                  30 Hari
                </button>
                <button
                  type="button"
                  @click="setDateRangePreset('all')"
                  class="px-2.5 py-1 text-xs font-bold rounded-lg transition-all cursor-pointer"
                  :class="!fromDate && !toDate ? 'bg-white shadow-sm text-slate-900' : 'text-slate-500 hover:text-slate-800'"
                >
                  Semua
                </button>
              </div>
              <div class="flex items-center gap-2">
                <Calendar class="h-4 w-4 text-slate-400 shrink-0" />
                <input
                  v-model="fromDate"
                  type="date"
                  class="text-sm font-medium bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1.5 focus:outline-none focus:ring-2 focus:ring-slate-200 focus:border-slate-400 transition-all"
                />
                <span class="text-slate-400 text-sm font-medium">—</span>
                <input
                  v-model="toDate"
                  type="date"
                  class="text-sm font-medium bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1.5 focus:outline-none focus:ring-2 focus:ring-slate-200 focus:border-slate-400 transition-all"
                />
                <button
                  v-if="searchQuery || selectedType !== 'ALL' || selectedSourceType !== 'ALL' || fromDate || toDate"
                  @click="resetFilters"
                  class="text-xs font-bold text-slate-400 hover:text-slate-800 transition-colors px-2 cursor-pointer"
                >
                  Reset
                </button>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Timeline List -->
    <Card class="border-none shadow-sm overflow-hidden">
      <!-- Loading -->
      <div v-if="marketplaceStore.isLoading" class="p-16 text-center">
        <div class="animate-spin h-8 w-8 border-2 border-slate-800 border-t-transparent rounded-full mx-auto mb-3"></div>
        <p class="text-slate-400 text-sm font-medium">Memuat riwayat pergerakan stok...</p>
      </div>

      <!-- Empty -->
      <div v-else-if="filteredMovements.length === 0" class="p-16 text-center">
        <Package class="h-14 w-14 text-slate-200 mx-auto mb-3" />
        <p class="text-slate-500 font-bold text-base">Belum ada riwayat pergerakan stok</p>
        <p class="text-slate-400 text-sm font-medium mt-1">
          {{ searchQuery || selectedType !== 'ALL' || selectedSourceType !== 'ALL' || fromDate || toDate
            ? 'Tidak ada data yang sesuai filter. Coba ubah kriteria pencarian atau platform.'
            : 'Riwayat akan muncul setelah ada aktivitas barang masuk, keluar, atau penyesuaian.'
          }}
        </p>
        <div v-if="searchQuery && (fromDate || toDate)" class="mt-4">
          <button
            type="button"
            @click="setDateRangePreset('all')"
            class="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl bg-slate-900 text-white text-xs font-bold hover:bg-slate-800 transition-all shadow-sm cursor-pointer"
          >
            Cari "{{ searchQuery }}" di Semua Periode
          </button>
        </div>
      </div>

      <!-- List -->
      <div v-else>
        <!-- Table Header -->
        <div class="hidden md:grid grid-cols-[auto_1fr_auto_auto_auto_auto] gap-3 px-5 py-3 bg-slate-50 border-b border-slate-100 text-[11px] font-black uppercase tracking-wider text-slate-400">
          <div class="w-24">Tipe</div>
          <div>Produk & Platform</div>
          <div class="text-right w-20">Qty</div>
          <div class="text-right w-24">Sebelum → Sesudah</div>
          <div class="text-right w-28">Nilai</div>
          <div class="text-right w-36">Waktu</div>
        </div>

        <!-- Rows -->
        <div class="divide-y divide-slate-50">
          <div
            v-for="m in filteredMovements"
            :key="m.id"
            class="group"
          >
            <!-- Row -->
            <div
              class="grid grid-cols-1 md:grid-cols-[auto_1fr_auto_auto_auto_auto] gap-3 px-5 py-3.5 hover:bg-slate-50 cursor-pointer transition-colors items-center"
              @click="toggleExpand(m.id)"
            >
              <!-- Type Badge -->
              <div class="w-full md:w-24">
                <span
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-[11px] font-black"
                  :class="[getTypeConfig(m.type).bgColor, getTypeConfig(m.type).textColor]"
                >
                  <component :is="getTypeConfig(m.type).icon" class="h-3 w-3" />
                  {{ getTypeConfig(m.type).label }}
                </span>
              </div>

              <!-- Product Info & Platform Badge -->
              <div class="min-w-0">
                <div class="flex items-center gap-2 flex-wrap">
                  <p class="font-bold text-slate-900 text-sm truncate">
                    {{ m.marketplace_stock?.product_name || 'Produk tidak diketahui' }}
                  </p>
                  <!-- Platform Badge -->
                  <span
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[10px] font-black border"
                    :class="getPlatformInfo(m).badgeClass"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="getPlatformInfo(m).dotClass"></span>
                    {{ getPlatformInfo(m).label }}
                  </span>
                </div>
                <div class="flex items-center gap-2 mt-0.5 flex-wrap">
                  <span class="text-[11px] text-slate-400 font-medium">SKU: {{ m.marketplace_stock?.sku || '-' }}</span>
                  <span v-if="m.reference" class="text-[11px] text-slate-500 font-mono font-bold bg-slate-100 px-1.5 py-0.2 rounded border border-slate-200/60">
                    {{ m.reference }}
                  </span>
                  <span v-if="m.created_by" class="text-[11px] text-slate-400 font-medium">· {{ m.created_by }}</span>
                </div>
              </div>

              <!-- Qty Change -->
              <div class="text-right">
                <span
                  class="text-lg font-black"
                  :class="m.type === 'STOCK_IN' ? 'text-emerald-600' : m.type === 'STOCK_OUT' ? 'text-rose-600' : 'text-violet-600'"
                >
                  {{ m.type === 'STOCK_IN' ? '+' : m.type === 'STOCK_OUT' ? '-' : '±' }}{{ m.quantity }}
                </span>
                <p class="text-[11px] text-slate-400 font-medium">unit</p>
              </div>

              <!-- Before → After -->
              <div class="text-right hidden md:block">
                <div class="flex items-center gap-1 justify-end text-sm font-bold text-slate-600">
                  <span class="text-slate-400">{{ m.quantity_before }}</span>
                  <span class="text-slate-300">→</span>
                  <span :class="m.type === 'STOCK_IN' ? 'text-emerald-600' : m.type === 'STOCK_OUT' ? 'text-rose-600' : 'text-violet-600'">
                    {{ m.quantity_after }}
                  </span>
                </div>
                <p class="text-[10px] text-slate-400 font-medium">stok</p>
              </div>

              <!-- Value -->
              <div class="text-right hidden md:block">
                <p v-if="m.total_value && m.total_value > 0" class="text-sm font-black text-slate-800">
                  {{ formatRupiah(m.total_value) }}
                </p>
                <p v-else class="text-sm text-slate-300 font-medium">—</p>
              </div>

              <!-- Time + Actions -->
              <div class="text-right flex items-center gap-2 justify-end">
                <div>
                  <p class="text-xs font-bold text-slate-500">{{ formatDateTime(m.created_at) }}</p>
                </div>
                <button
                  type="button"
                  @click.stop="openEditMovement(m)"
                  class="p-1.5 text-slate-400 hover:text-primary hover:bg-primary/10 rounded-lg transition-all cursor-pointer"
                  title="Edit Data Pergerakan"
                >
                  <Pencil class="h-3.5 w-3.5" />
                </button>
                <component
                  :is="expandedId === m.id ? ChevronUp : ChevronDown"
                  class="h-4 w-4 text-slate-300 group-hover:text-slate-500 transition-colors"
                />
              </div>
            </div>

            <!-- Expanded Detail -->
            <div
              v-if="expandedId === m.id"
              class="px-5 pb-4 bg-gradient-to-b from-slate-50 to-white border-b border-slate-100"
            >
              <div class="flex items-center justify-between mt-2 mb-2">
                <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Detail Transaksi</span>
                <button
                  type="button"
                  @click.stop="openEditMovement(m)"
                  class="inline-flex items-center gap-1.5 px-3 py-1 text-xs font-bold bg-white hover:bg-slate-100 text-slate-700 border border-slate-200 rounded-lg shadow-sm transition-all cursor-pointer"
                >
                  <Pencil class="h-3.5 w-3.5 text-primary" />
                  Edit Data Pergerakan
                </button>
              </div>
              <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
                <div class="bg-white rounded-xl border border-slate-100 p-3">
                  <p class="text-[10px] uppercase font-black text-slate-400 mb-1">Referensi</p>
                  <p class="text-sm font-bold text-slate-700 font-mono">{{ m.reference || '—' }}</p>
                </div>

                <div class="bg-white rounded-xl border border-slate-100 p-3">
                  <p class="text-[10px] uppercase font-black text-slate-400 mb-1">Harga/Unit</p>
                  <p class="text-sm font-bold text-slate-700">{{ m.price ? formatRupiah(m.price) : '—' }}</p>
                </div>
                <div class="bg-white rounded-xl border border-slate-100 p-3">
                  <p class="text-[10px] uppercase font-black text-slate-400 mb-1">Total Nilai</p>
                  <p class="text-sm font-bold" :class="m.total_value > 0 ? 'text-slate-800' : 'text-slate-400'">
                    {{ m.total_value > 0 ? formatRupiah(m.total_value) : '—' }}
                  </p>
                </div>
                <div class="bg-white rounded-xl border border-slate-100 p-3">
                  <p class="text-[10px] uppercase font-black text-slate-400 mb-1">Sumber Platform</p>
                  <div class="flex items-center gap-1.5 mt-0.5">
                    <span
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-bold border"
                      :class="getPlatformInfo(m).badgeClass"
                    >
                      <span class="w-1.5 h-1.5 rounded-full" :class="getPlatformInfo(m).dotClass"></span>
                      {{ getPlatformInfo(m).label }}
                    </span>
                  </div>
                </div>
              </div>
              <div v-if="m.notes" class="mt-3 bg-white rounded-xl border border-slate-100 p-3">
                <p class="text-[10px] uppercase font-black text-slate-400 mb-1">Catatan</p>
                <p class="text-sm text-slate-600 font-medium">{{ m.notes }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer count -->
        <div class="px-5 py-3 bg-slate-50 border-t border-slate-100 flex items-center justify-between flex-wrap gap-3">
          <p class="text-xs font-bold text-slate-400">
            Menampilkan {{ filteredMovements.length }} transaksi di halaman ini (Total: {{ marketplaceStore.movementsTotal }})
          </p>
          <div class="flex items-center gap-3 text-xs font-bold">
            <span class="text-emerald-600">+{{ stats.totalIn }} masuk (hal ini)</span>
            <span class="text-rose-600">-{{ stats.totalOut }} keluar (hal ini)</span>
            <span class="text-violet-600">{{ stats.totalAdj }}x penyesuaian (hal ini)</span>
          </div>
        </div>

        <!-- Pagination Controls -->
        <div class="px-5 py-3 bg-white border-t border-slate-100 flex items-center justify-between">
          <button
            :disabled="marketplaceStore.movementsPage <= 1"
            @click="marketplaceStore.movementsPage--; loadData()"
            class="px-4 py-2 text-xs font-bold rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          >
            Sebelumnya
          </button>
          
          <span class="text-xs font-bold text-slate-500">
            Halaman {{ marketplaceStore.movementsPage }} dari {{ Math.ceil(marketplaceStore.movementsTotal / marketplaceStore.movementsLimit) || 1 }}
          </span>

          <button
            :disabled="marketplaceStore.movementsPage >= Math.ceil(marketplaceStore.movementsTotal / marketplaceStore.movementsLimit)"
            @click="marketplaceStore.movementsPage++; loadData()"
            class="px-4 py-2 text-xs font-bold rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          >
            Selanjutnya
          </button>
        </div>
      </div>
    </Card>

    <!-- ===== EXPORT EXCEL MODAL ===== -->
    <Teleport to="body">
      <div v-if="showExportModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showExportModal = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-lg overflow-hidden animate-in zoom-in-95 duration-200">
          <!-- Modal Header -->
          <div class="bg-gradient-to-r from-emerald-600 to-teal-700 p-5 text-white flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <div class="p-2 bg-white/10 rounded-xl">
                <FileSpreadsheet class="h-5 w-5 text-white" />
              </div>
              <div>
                <h2 class="text-base font-black">Export Laporan Excel</h2>
                <p class="text-emerald-100 text-xs font-medium">Pilih cakupan data dan periode mutasi yang ingin diekspor</p>
              </div>
            </div>
            <button
              type="button"
              @click="showExportModal = false"
              class="p-1.5 text-white/80 hover:text-white rounded-lg hover:bg-white/10 transition-colors"
            >
              <X class="h-5 w-5" />
            </button>
          </div>

          <!-- Modal Body -->
          <div class="p-6 space-y-5 max-h-[80vh] overflow-y-auto">
            <!-- Pilihan Cakupan Data -->
            <div>
              <label class="block text-xs font-black uppercase tracking-wider text-slate-500 mb-2">
                Pilih Kategori / Cakupan Data *
              </label>
              <div class="space-y-2">
                <!-- Option 1: TikTok Orders Only -->
                <button
                  type="button"
                  @click="exportScope = 'TIKTOK_ORDERS'"
                  class="w-full text-left p-3.5 rounded-2xl border-2 transition-all flex items-start gap-3 cursor-pointer"
                  :class="exportScope === 'TIKTOK_ORDERS'
                    ? 'bg-slate-900/5 border-slate-900 shadow-sm shadow-slate-200'
                    : 'bg-white border-slate-200 hover:border-slate-300 hover:bg-slate-50'"
                >
                  <div
                    class="p-2 rounded-xl mt-0.5"
                    :class="exportScope === 'TIKTOK_ORDERS' ? 'bg-black text-white' : 'bg-slate-100 text-slate-500'"
                  >
                    <ShoppingBag class="h-4 w-4" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-black text-sm text-slate-900">Orderan Keluar TikTok Shop Saja</span>
                      <span class="px-2 py-0.5 rounded-full text-[10px] font-black bg-slate-900 text-white">TikTok</span>
                    </div>
                    <p class="text-xs text-slate-500 mt-0.5">
                      Hanya mengekspor barang keluar dari hasil packing order pesanan TikTok Shop.
                    </p>
                  </div>
                  <CheckCircle2
                    v-if="exportScope === 'TIKTOK_ORDERS'"
                    class="h-5 w-5 text-slate-900 shrink-0 mt-1"
                  />
                </button>

                <!-- Option 2: Shopee Orders Only -->
                <button
                  type="button"
                  @click="exportScope = 'SHOPEE_ORDERS'"
                  class="w-full text-left p-3.5 rounded-2xl border-2 transition-all flex items-start gap-3 cursor-pointer"
                  :class="exportScope === 'SHOPEE_ORDERS'
                    ? 'bg-orange-50/70 border-orange-500 shadow-sm shadow-orange-100'
                    : 'bg-white border-slate-200 hover:border-slate-300 hover:bg-slate-50'"
                >
                  <div
                    class="p-2 rounded-xl mt-0.5"
                    :class="exportScope === 'SHOPEE_ORDERS' ? 'bg-orange-500 text-white' : 'bg-slate-100 text-slate-500'"
                  >
                    <ShoppingBag class="h-4 w-4" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-black text-sm text-slate-900">Orderan Keluar Shopee Saja</span>
                      <span class="px-2 py-0.5 rounded-full text-[10px] font-black bg-orange-100 text-orange-700">Shopee</span>
                    </div>
                    <p class="text-xs text-slate-500 mt-0.5">
                      Hanya mengekspor barang keluar dari hasil packing order pesanan Shopee.
                    </p>
                  </div>
                  <CheckCircle2
                    v-if="exportScope === 'SHOPEE_ORDERS'"
                    class="h-5 w-5 text-orange-600 shrink-0 mt-1"
                  />
                </button>

                <!-- Option 3: Semua Orderan Marketplace -->
                <button
                  type="button"
                  @click="exportScope = 'ALL_ORDERS'"
                  class="w-full text-left p-3.5 rounded-2xl border-2 transition-all flex items-start gap-3 cursor-pointer"
                  :class="exportScope === 'ALL_ORDERS'
                    ? 'bg-emerald-50/70 border-emerald-500 shadow-sm shadow-emerald-100'
                    : 'bg-white border-slate-200 hover:border-slate-300 hover:bg-slate-50'"
                >
                  <div
                    class="p-2 rounded-xl mt-0.5"
                    :class="exportScope === 'ALL_ORDERS' ? 'bg-emerald-600 text-white' : 'bg-slate-100 text-slate-500'"
                  >
                    <ShoppingBag class="h-4 w-4" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-black text-sm text-slate-900">Semua Orderan Keluar (TikTok & Shopee)</span>
                      <span class="px-2 py-0.5 rounded-full text-[10px] font-black bg-emerald-100 text-emerald-700">Semua Order</span>
                    </div>
                    <p class="text-xs text-slate-500 mt-0.5">
                      Mengekspor seluruh barang keluar dari orderan TikTok Shop dan Shopee.
                    </p>
                  </div>
                  <CheckCircle2
                    v-if="exportScope === 'ALL_ORDERS'"
                    class="h-5 w-5 text-emerald-600 shrink-0 mt-1"
                  />
                </button>

                <!-- Option 4: Semua Data Mutasi -->
                <button
                  type="button"
                  @click="exportScope = 'ALL'"
                  class="w-full text-left p-3.5 rounded-2xl border-2 transition-all flex items-start gap-3 cursor-pointer"
                  :class="exportScope === 'ALL'
                    ? 'bg-slate-900/5 border-slate-900 shadow-sm shadow-slate-200'
                    : 'bg-white border-slate-200 hover:border-slate-300 hover:bg-slate-50'"
                >
                  <div
                    class="p-2 rounded-xl mt-0.5"
                    :class="exportScope === 'ALL' ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-500'"
                  >
                    <Layers class="h-4 w-4" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <span class="font-black text-sm text-slate-900">Semua Data Mutasi Stok Lengkap</span>
                    <p class="text-xs text-slate-500 mt-0.5">
                      Mencakup barang masuk, keluar manual, keluar orderan, dan penyesuaian opname stok.
                    </p>
                  </div>
                  <CheckCircle2
                    v-if="exportScope === 'ALL'"
                    class="h-5 w-5 text-slate-900 shrink-0 mt-1"
                  />
                </button>

                <!-- Option 5: Barang Masuk Saja -->
                <button
                  type="button"
                  @click="exportScope = 'STOCK_IN'"
                  class="w-full text-left p-3 rounded-2xl border-2 transition-all flex items-start gap-3 cursor-pointer"
                  :class="exportScope === 'STOCK_IN'
                    ? 'bg-emerald-50/70 border-emerald-500'
                    : 'bg-white border-slate-200 hover:border-slate-300 hover:bg-slate-50'"
                >
                  <div
                    class="p-1.5 rounded-lg mt-0.5"
                    :class="exportScope === 'STOCK_IN' ? 'bg-emerald-600 text-white' : 'bg-slate-100 text-slate-500'"
                  >
                    <TrendingUp class="h-3.5 w-3.5" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <span class="font-bold text-xs text-slate-800">Khusus Barang Masuk Saja</span>
                  </div>
                  <CheckCircle2
                    v-if="exportScope === 'STOCK_IN'"
                    class="h-4 w-4 text-emerald-600 shrink-0 mt-0.5"
                  />
                </button>

                <!-- Option 6: Penyesuaian Saja -->
                <button
                  type="button"
                  @click="exportScope = 'ADJUSTMENT'"
                  class="w-full text-left p-3 rounded-2xl border-2 transition-all flex items-start gap-3 cursor-pointer"
                  :class="exportScope === 'ADJUSTMENT'
                    ? 'bg-violet-50/70 border-violet-500'
                    : 'bg-white border-slate-200 hover:border-slate-300 hover:bg-slate-50'"
                >
                  <div
                    class="p-1.5 rounded-lg mt-0.5"
                    :class="exportScope === 'ADJUSTMENT' ? 'bg-violet-600 text-white' : 'bg-slate-100 text-slate-500'"
                  >
                    <ArrowRightLeft class="h-3.5 w-3.5" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <span class="font-bold text-xs text-slate-800">Khusus Penyesuaian / Opname Saja</span>
                  </div>
                  <CheckCircle2
                    v-if="exportScope === 'ADJUSTMENT'"
                    class="h-4 w-4 text-violet-600 shrink-0 mt-0.5"
                  />
                </button>
              </div>
            </div>

            <!-- Rentang Tanggal Ekspor -->
            <div>
              <label class="block text-xs font-black uppercase tracking-wider text-slate-500 mb-2">
                Rentang Tanggal Ekspor
              </label>
              <!-- Preset Buttons -->
              <div class="flex flex-wrap gap-1.5 mb-2.5">
                <button
                  v-for="p in [
                    { id: 'today', label: 'Hari Ini' },
                    { id: '7d', label: '7 Hari Terakhir' },
                    { id: '30d', label: '30 Hari' },
                    { id: 'this_month', label: 'Bulan Ini' },
                    { id: 'all', label: 'Semua Periode' },
                  ]"
                  :key="p.id"
                  type="button"
                  @click="setExportDatePreset(p.id as any)"
                  class="px-3 py-1.5 text-xs font-bold rounded-xl border transition-all"
                  :class="exportDatePreset === p.id
                    ? 'bg-slate-900 text-white border-slate-900 shadow-sm'
                    : 'bg-slate-50 text-slate-600 border-slate-200 hover:bg-slate-100'"
                >
                  {{ p.label }}
                </button>
              </div>

              <!-- Date Pickers -->
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-[10px] font-black uppercase tracking-wider text-slate-400 mb-1">Dari Tanggal</label>
                  <input
                    v-model="exportFromDate"
                    @change="exportDatePreset = 'custom'"
                    type="date"
                    class="w-full text-xs font-bold bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
                  />
                </div>
                <div>
                  <label class="block text-[10px] font-black uppercase tracking-wider text-slate-400 mb-1">Sampai Tanggal</label>
                  <input
                    v-model="exportToDate"
                    @change="exportDatePreset = 'custom'"
                    type="date"
                    class="w-full text-xs font-bold bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
                  />
                </div>
              </div>
            </div>

            <!-- Filter Kata Kunci (Opsional) -->
            <div>
              <label class="block text-xs font-black uppercase tracking-wider text-slate-500 mb-1.5">
                Kata Kunci / No. Referensi (Opsional)
              </label>
              <div class="relative">
                <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400" />
                <input
                  v-model="exportSearchQuery"
                  type="text"
                  placeholder="Filter nama barang / SKU / Order ID tertentu..."
                  class="w-full pl-8 pr-3 py-2 text-xs font-medium bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
                />
              </div>
            </div>

            <!-- Info Summary Box -->
            <div class="p-3.5 bg-slate-50 rounded-2xl border border-slate-200 text-xs space-y-1">
              <div class="flex items-center justify-between text-slate-600">
                <span class="font-medium">Cabang:</span>
                <span class="font-bold text-slate-900">
                  {{ masterDataStore.branches.find((b: any) => b.id === masterDataStore.selectedBranchId)?.name || 'Semua Cabang' }}
                </span>
              </div>
              <div class="flex items-center justify-between text-slate-600">
                <span class="font-medium">Format File:</span>
                <span class="font-bold text-emerald-700">Excel (.xlsx) · 2 Tab (Rekapitulasi & Detail)</span>
              </div>
            </div>
            <div v-if="isExporting" class="p-3.5 bg-emerald-50 rounded-2xl border border-emerald-100 space-y-2">
              <div class="flex items-center justify-between text-xs">
                <span class="font-bold text-emerald-800">{{ exportProgressText }}</span>
                <span class="font-black text-emerald-700">{{ exportProgress }}%</span>
              </div>
              <div class="h-2 overflow-hidden rounded-full bg-emerald-100">
                <div
                  class="h-full rounded-full bg-emerald-500 transition-all duration-300"
                  :style="{ width: `${exportProgress}%` }"
                />
              </div>
            </div>
          </div>

          <!-- Modal Footer -->
          <div class="p-4 bg-slate-50 border-t border-slate-100 flex items-center gap-3">
            <button
              type="button"
              @click="showExportModal = false"
              class="flex-1 px-4 py-2.5 text-xs font-bold border border-slate-200 rounded-xl hover:bg-white text-slate-600 transition-colors"
            >
              Batal
            </button>
            <button
              type="button"
              @click="executeExport"
              :disabled="isExporting"
              class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 text-xs font-bold bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl shadow-md shadow-emerald-600/20 transition-all disabled:opacity-50"
            >
              <Loader2 v-if="isExporting" class="h-4 w-4 animate-spin text-white" />
              <Download v-else class="h-4 w-4" />
              <span>{{ isExporting ? exportProgressText : 'Download File Excel' }}</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ===== EDIT MOVEMENT MODAL ===== -->
    <Teleport to="body">
      <div v-if="showEditModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showEditModal = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-md overflow-hidden animate-in zoom-in-95 duration-200">
          <!-- Modal Header -->
          <div class="bg-gradient-to-r from-slate-800 to-slate-900 p-5 text-white flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-white/10 rounded-xl">
                <Pencil class="h-5 w-5 text-white" />
              </div>
              <div>
                <h2 class="text-base font-black">Edit Pergerakan Stok</h2>
                <p class="text-slate-300 text-xs font-medium">{{ editForm.product_name }}</p>
              </div>
            </div>
            <button @click="showEditModal = false" class="p-1.5 text-slate-400 hover:text-white rounded-lg hover:bg-white/10 transition-colors cursor-pointer">
              <X class="h-4 w-4" />
            </button>
          </div>

          <!-- Form Body -->
          <form @submit.prevent="submitEditMovement" class="p-6 space-y-4">
            <!-- Product & Type Info -->
            <div class="p-3 bg-slate-50 border border-slate-100 rounded-2xl flex items-center justify-between">
              <div>
                <p class="text-[10px] font-black uppercase tracking-wider text-slate-400">SKU</p>
                <p class="text-xs font-mono font-bold text-slate-800">{{ editForm.sku || '-' }}</p>
              </div>
              <span
                class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-black"
                :class="[getTypeConfig(editForm.type).bgColor, getTypeConfig(editForm.type).textColor]"
              >
                <component :is="getTypeConfig(editForm.type).icon" class="h-3.5 w-3.5" />
                {{ getTypeConfig(editForm.type).label }}
              </span>
            </div>

            <!-- Tanggal & Waktu -->
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500">Tanggal & Waktu Transaksi *</label>
                <button
                  type="button"
                  @click="editForm.date = getCurrentLocalDateTime()"
                  class="text-[11px] font-bold text-primary hover:underline cursor-pointer"
                >
                  Set Waktu Sekarang
                </button>
              </div>
              <input
                v-model="editForm.date"
                type="datetime-local"
                required
                class="w-full px-3 py-2 text-sm font-bold border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 text-slate-800"
              />
              <p class="text-[11px] text-slate-400 mt-1">
                Koreksi waktu transaksi agar filter laporan dan urutan riwayat sesuai.
              </p>
            </div>

            <!-- Quantity -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Jumlah Mutasi (Qty)</label>
              <input
                v-model.number="editForm.quantity"
                type="number"
                min="0"
                required
                :disabled="Boolean(editForm.source_type && editForm.source_type !== 'MANUAL')"
                class="w-full px-3 py-2 text-sm font-black border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 text-slate-800 disabled:bg-slate-100 disabled:text-slate-400"
              />
              <p v-if="editForm.source_type && editForm.source_type !== 'MANUAL'" class="text-[11px] text-amber-600 mt-1">
                * Kuantiti mutasi pesanan marketplace otomatis dikunci agar sinkron dengan order.
              </p>
              <p v-else class="text-[11px] text-slate-400 mt-1">
                Saldo stok fisik terkini akan otomatis disesuaikan dengan selisih jumlah ini.
              </p>
            </div>

            <!-- No. Referensi -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">No. Referensi / Dokumen</label>
              <input
                v-model="editForm.reference"
                type="text"
                class="w-full px-3 py-2 text-sm font-medium border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 text-slate-800"
                placeholder="No. order, surat jalan, atau batch..."
              />
            </div>

            <!-- Catatan -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Catatan / Alasan Koreksi</label>
              <textarea
                v-model="editForm.notes"
                rows="2"
                class="w-full px-3 py-2 text-sm font-medium border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 text-slate-800 resize-none"
                placeholder="Tulis alasan koreksi atau keterangan..."
              />
            </div>

            <!-- Action Buttons -->
            <div class="flex gap-3 pt-2">
              <button
                type="button"
                @click="showEditModal = false"
                class="flex-1 px-4 py-2.5 text-xs font-bold border border-slate-200 rounded-xl hover:bg-slate-50 text-slate-600 transition-colors cursor-pointer"
              >
                Batal
              </button>
              <button
                type="submit"
                :disabled="isUpdating"
                class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 text-xs font-bold bg-slate-900 hover:bg-slate-800 text-white rounded-xl shadow-md transition-all disabled:opacity-50 cursor-pointer"
              >
                <Loader2 v-if="isUpdating" class="h-4 w-4 animate-spin text-white" />
                <span>{{ isUpdating ? 'Menyimpan...' : 'Simpan Perubahan' }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
