<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import {
  Banknote,
  CreditCard,
  Wallet,
  Receipt,
  Calendar,
  Building2,
  RefreshCw,
  Download,
  Search,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  ZoomIn,
  ZoomOut,
  RotateCw,
  X,
  ChevronDown,
  ChevronUp,
  MapPin,
  Clock,
  User,
  Store,
  Filter,
  FileSpreadsheet
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import ExcelJS from 'exceljs'
import { saveAs } from 'file-saver'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

const isExporting = ref(false)

// State Filters
const getTodayDateStr = () => {
  const d = new Date()
  const offset = d.getTimezoneOffset() * 60000
  return new Date(d.getTime() - offset).toISOString().split('T')[0]
}

const selectedDate = ref(getTodayDateStr())
const selectedCompanyId = ref('ALL')
const selectedBranchId = ref('')
const searchFilter = ref('')
const selectedMethodFilter = ref('ALL') // 'ALL' | 'CASH' | 'TRANSFER'

// Cabang Pabrik (Pendapatan Sales Harian dikhususkan hanya untuk Pabrik Utama)
const pabrikBranches = computed(() => {
  return masterDataStore.branches.filter((b: any) =>
    (b.name || '').toLowerCase().includes('pabrik')
  )
})

// Loading & Data States
const isLoading = ref(false)
const reportData = ref<any>({
  summary: {
    date: getTodayDateStr(),
    total_cash: 0,
    total_transfer: 0,
    grand_total: 0,
    total_transactions: 0
  },
  companies: [],
  payments: []
})

// Expanded state for PT cards
const expandedCompanies = ref<Record<string, boolean>>({})

// Image Preview Modal
const showPreviewModal = ref(false)
const previewImageUrl = ref('')
const previewTitle = ref('Bukti Pembayaran / Transfer')
const previewSubtitle = ref('')
const zoomScale = ref(1.0)
const rotationDegrees = ref(0)

const openPreviewModal = (url: string, title = 'Bukti Pembayaran', subtitle = '') => {
  previewImageUrl.value = url
  previewTitle.value = title
  previewSubtitle.value = subtitle
  zoomScale.value = 1.0
  rotationDegrees.value = 0
  showPreviewModal.value = true
}

const closePreviewModal = () => {
  showPreviewModal.value = false
}

// Quick Date Helpers
const setToday = () => {
  selectedDate.value = getTodayDateStr()
  fetchDailyIncome()
}

const setYesterday = () => {
  const d = new Date()
  d.setDate(d.getDate() - 1)
  const offset = d.getTimezoneOffset() * 60000
  selectedDate.value = new Date(d.getTime() - offset).toISOString().split('T')[0]
  fetchDailyIncome()
}

// Currency Formatter
const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(val || 0)
}

// Date Formatter
const formatDateTime = (dateStr?: string | null) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(d)
}

const formatDateOnly = (dateStr?: string | null) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return dateStr
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  }).format(d)
}

// Fetch API Data
const fetchDailyIncome = async () => {
  isLoading.value = true
  try {
    // Pastikan selectedBranchId mengarah ke Pabrik Utama jika belum di-set
    if (!selectedBranchId.value) {
      const pabrik = pabrikBranches.value[0]
      if (pabrik) {
        selectedBranchId.value = pabrik.id
      }
    }

    const params: Record<string, string> = {
      date: selectedDate.value
    }
    if (selectedCompanyId.value && selectedCompanyId.value !== 'ALL') {
      params.company_id = selectedCompanyId.value
    }
    if (selectedBranchId.value) {
      params.branch_id = selectedBranchId.value
    }

    const res = await apiClient.get('/finance/sales-income/daily', { params })
    if (res.data?.data) {
      reportData.value = res.data.data

      // Auto-expand all companies by default
      if (reportData.value.companies) {
        reportData.value.companies.forEach((comp: any) => {
          if (expandedCompanies.value[comp.company_id] === undefined) {
            expandedCompanies.value[comp.company_id] = true
          }
        })
      }
    }
  } catch (err: any) {
    console.error('Failed to load daily sales income:', err)
    toast.error('Gagal memuat rekap pendapatan sales harian')
  } finally {
    isLoading.value = false
  }
}

// Toggle Company Card Expansion
const toggleCompanyExpand = (compID: string) => {
  expandedCompanies.value[compID] = !expandedCompanies.value[compID]
}

// Filtered Payments Table
const filteredPayments = computed(() => {
  const list = reportData.value?.payments || []
  return list.filter((p: any) => {
    // Payment method filter
    if (selectedMethodFilter.value === 'CASH') {
      const isCash = p.payment_method?.toUpperCase() === 'CASH' || p.payment_method?.toUpperCase() === 'TUNAI'
      if (!isCash) return false
    } else if (selectedMethodFilter.value === 'TRANSFER') {
      const isCash = p.payment_method?.toUpperCase() === 'CASH' || p.payment_method?.toUpperCase() === 'TUNAI'
      if (isCash) return false
    }

    // Search query
    if (searchFilter.value.trim()) {
      const q = searchFilter.value.toLowerCase()
      const matchStore = p.store_name?.toLowerCase().includes(q)
      const matchSales = p.salesman_name?.toLowerCase().includes(q)
      const matchRef = p.reference_no?.toLowerCase().includes(q)
      const matchBranch = p.branch_name?.toLowerCase().includes(q)
      const matchCompany = p.company_name?.toLowerCase().includes(q)
      const matchApprover = p.approver_name?.toLowerCase().includes(q)
      if (!matchStore && !matchSales && !matchRef && !matchBranch && !matchCompany && !matchApprover) {
        return false
      }
    }

    return true
  })
})

// Export to Excel (.xlsx) dengan UI & UX Profesional
const exportToExcel = async () => {
  const rows = filteredPayments.value
  if (!rows || rows.length === 0) {
    toast.error('Tidak ada data untuk diekspor')
    return
  }

  isExporting.value = true
  try {
    const workbook = new ExcelJS.Workbook()
    workbook.creator = 'Karsa HR ERP - Keuangan Pabrik'
    workbook.lastModifiedBy = 'Admin Kasir Pabrik'
    workbook.created = new Date()
    workbook.modified = new Date()

    // Ambil nama cabang aktif
    const activeBranch = masterDataStore.branches.find(b => b.id === selectedBranchId.value)
    const activeBranchName = activeBranch?.name || 'Pabrik Utama'
    const formattedReportDate = formatDateOnly(selectedDate.value)

    // ==========================================
    // SHEET 1: Rincian Setoran Sales
    // ==========================================
    const wsDetail = workbook.addWorksheet('Rincian Setoran Sales', {
      views: [{ showGridLines: true }]
    })

    // Header Laporan
    wsDetail.mergeCells('A1:L1')
    const mainTitle = wsDetail.getCell('A1')
    mainTitle.value = `REKAPITULASI PENDAPATAN SALES HARIAN - ${activeBranchName.toUpperCase()}`
    mainTitle.font = { name: 'Segoe UI', size: 14, bold: true, color: { argb: 'FFFFFFFF' } }
    mainTitle.alignment = { vertical: 'middle', horizontal: 'center' }
    mainTitle.fill = {
      type: 'pattern',
      pattern: 'solid',
      fgColor: { argb: 'FF0F766E' } // Deep Teal/Emerald
    }
    wsDetail.getRow(1).height = 30

    wsDetail.mergeCells('A2:L2')
    const subTitle = wsDetail.getCell('A2')
    subTitle.value = `Tanggal Setoran: ${formattedReportDate}   |   Waktu Unduh: ${new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date())} WIB`
    subTitle.font = { name: 'Segoe UI', size: 10, italic: true, color: { argb: 'FF475569' } }
    subTitle.alignment = { vertical: 'middle', horizontal: 'center' }
    wsDetail.getRow(2).height = 20

    // KPI Summary Section
    const totalCash = rows.reduce((acc: number, cur: any) => {
      const isCash = cur.payment_method?.toUpperCase() === 'CASH' || cur.payment_method?.toUpperCase() === 'TUNAI'
      return isCash ? acc + (Number(cur.amount) || 0) : acc
    }, 0)

    const totalTransfer = rows.reduce((acc: number, cur: any) => {
      const isCash = cur.payment_method?.toUpperCase() === 'CASH' || cur.payment_method?.toUpperCase() === 'TUNAI'
      return !isCash ? acc + (Number(cur.amount) || 0) : acc
    }, 0)

    const grandTotal = totalCash + totalTransfer

    // Baris Summary KPI
    wsDetail.mergeCells('B4:D4')
    wsDetail.getCell('B4').value = 'RINGKASAN TOTAL SETORAN'
    wsDetail.getCell('B4').font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF334155' } }

    const kpiRow1 = wsDetail.getRow(5)
    kpiRow1.getCell(2).value = 'Total Tunai (Cash):'
    kpiRow1.getCell(3).value = totalCash
    kpiRow1.getCell(3).numFmt = '"Rp"#,##0'
    kpiRow1.getCell(2).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF047857' } }
    kpiRow1.getCell(3).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF065F46' } }

    const kpiRow2 = wsDetail.getRow(6)
    kpiRow2.getCell(2).value = 'Total Transfer (TF):'
    kpiRow2.getCell(3).value = totalTransfer
    kpiRow2.getCell(3).numFmt = '"Rp"#,##0'
    kpiRow2.getCell(2).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF4338CA' } }
    kpiRow2.getCell(3).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF3730A3' } }

    const kpiRow3 = wsDetail.getRow(7)
    kpiRow3.getCell(2).value = 'GRAND TOTAL PENDAPATAN:'
    kpiRow3.getCell(3).value = grandTotal
    kpiRow3.getCell(3).numFmt = '"Rp"#,##0'
    kpiRow3.getCell(2).font = { name: 'Segoe UI', size: 11, bold: true, color: { argb: 'FF0F172A' } }
    kpiRow3.getCell(3).font = { name: 'Segoe UI', size: 11, bold: true, color: { argb: 'FF047857' } }

    kpiRow3.getCell(2).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD1FAE5' } }
    kpiRow3.getCell(3).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD1FAE5' } }

    kpiRow1.getCell(5).value = 'Total Transaksi Nota:'
    kpiRow1.getCell(6).value = `${rows.length} Transaksi`
    kpiRow1.getCell(5).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF64748B' } }
    kpiRow1.getCell(6).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF1E293B' } }

    // Table Column Headers
    const headers = [
      'NO',
      'TANGGAL BAYAR',
      'WAKTU ACC KASIR',
      'NO. REFERENSI (SO / FAKTUR)',
      'PERUSAHAAN (PT)',
      'CABANG',
      'NAMA TOKO / PELANGGAN',
      'SALESMAN',
      'METODE',
      'NOMINAL (RP)',
      'KASIR VERIFIKATOR',
      'CATATAN'
    ]

    const headerRow = wsDetail.getRow(9)
    headers.forEach((h, idx) => {
      const cell = headerRow.getCell(idx + 1)
      cell.value = h
      cell.fill = {
        type: 'pattern',
        pattern: 'solid',
        fgColor: { argb: 'FF1E293B' } // Slate 800
      }
      cell.font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FFFFFFFF' } }
      cell.alignment = {
        vertical: 'middle',
        horizontal: idx === 0 || idx === 8 ? 'center' : (idx === 9 ? 'right' : 'left'),
        wrapText: true
      }
      cell.border = {
        top: { style: 'thin', color: { argb: 'FF94A3B8' } },
        bottom: { style: 'medium', color: { argb: 'FF0F766E' } },
        left: { style: 'thin', color: { argb: 'FF334155' } },
        right: { style: 'thin', color: { argb: 'FF334155' } }
      }
    })
    headerRow.height = 26

    // Populate Data Rows
    let currentRowIdx = 10
    rows.forEach((p: any, idx: number) => {
      const row = wsDetail.getRow(currentRowIdx)
      const isEven = idx % 2 === 0
      const bgColor = isEven ? 'FFFFFFFF' : 'FFF8FAFC'

      const isCash = p.payment_method?.toUpperCase() === 'CASH' || p.payment_method?.toUpperCase() === 'TUNAI'
      const paymentMethodLabel = isCash ? 'TUNAI' : (p.payment_method || 'TRANSFER')

      row.getCell(1).value = idx + 1
      row.getCell(2).value = p.payment_date ? formatDateOnly(p.payment_date) : '-'
      row.getCell(3).value = p.approved_at ? formatDateTime(p.approved_at) : (p.payment_date ? formatDateTime(p.payment_date) : '-')
      row.getCell(4).value = p.reference_no || '-'
      row.getCell(5).value = p.company_name || '-'
      row.getCell(6).value = p.branch_name || '-'
      row.getCell(7).value = (p.store_name || '-').toUpperCase()
      row.getCell(8).value = p.salesman_name || '-'
      row.getCell(9).value = paymentMethodLabel
      row.getCell(10).value = Number(p.amount) || 0
      row.getCell(11).value = p.approver_name || 'Kasir'
      row.getCell(12).value = p.notes || '-'

      // Cell Styling
      for (let c = 1; c <= 12; c++) {
        const cell = row.getCell(c)
        cell.font = { name: 'Segoe UI', size: 9.5 }
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: bgColor } }
        cell.border = {
          top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          right: { style: 'thin', color: { argb: 'FFE2E8F0' } }
        }
        cell.alignment = { vertical: 'middle', horizontal: 'left' }
      }

      // Alignment specifics
      row.getCell(1).alignment = { vertical: 'middle', horizontal: 'center' }
      row.getCell(2).alignment = { vertical: 'middle', horizontal: 'center' }
      row.getCell(3).alignment = { vertical: 'middle', horizontal: 'center' }
      row.getCell(4).alignment = { vertical: 'middle', horizontal: 'center' }
      row.getCell(4).font = { name: 'Segoe UI', size: 9.5, bold: true, color: { argb: 'FF1E293B' } }
      row.getCell(7).font = { name: 'Segoe UI', size: 9.5, bold: true, color: { argb: 'FF0F172A' } }
      
      // Method badge coloring
      row.getCell(9).alignment = { vertical: 'middle', horizontal: 'center' }
      if (isCash) {
        row.getCell(9).font = { name: 'Segoe UI', size: 9, bold: true, color: { argb: 'FF047857' } }
        row.getCell(9).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFECFDF5' } }
      } else {
        row.getCell(9).font = { name: 'Segoe UI', size: 9, bold: true, color: { argb: 'FF4338CA' } }
        row.getCell(9).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFEEF2FF' } }
      }

      // Amount column formatting
      row.getCell(10).numFmt = '"Rp"#,##0'
      row.getCell(10).alignment = { vertical: 'middle', horizontal: 'right' }
      row.getCell(10).font = { name: 'Segoe UI', size: 9.5, bold: true, color: { argb: 'FF0F172A' } }

      row.height = 22
      currentRowIdx++
    })

    // Table Footer Total Row
    const footerRow = wsDetail.getRow(currentRowIdx)
    wsDetail.mergeCells(`A${currentRowIdx}:I${currentRowIdx}`)
    const totalLabelCell = footerRow.getCell(1)
    totalLabelCell.value = 'TOTAL KESELURUHAN PENDAPATAN SALES:'
    totalLabelCell.font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF1E293B' } }
    totalLabelCell.alignment = { vertical: 'middle', horizontal: 'right' }
    totalLabelCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    const totalValCell = footerRow.getCell(10)
    totalValCell.value = { formula: `SUM(J10:J${currentRowIdx - 1})`, result: grandTotal }
    totalValCell.numFmt = '"Rp"#,##0'
    totalValCell.font = { name: 'Segoe UI', size: 11, bold: true, color: { argb: 'FF047857' } }
    totalValCell.alignment = { vertical: 'middle', horizontal: 'right' }
    totalValCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD1FAE5' } }

    // Borders for footer
    for (let c = 1; c <= 12; c++) {
      const cell = footerRow.getCell(c)
      cell.border = {
        top: { style: 'medium', color: { argb: 'FF0F766E' } },
        bottom: { style: 'medium', color: { argb: 'FF0F766E' } },
        left: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        right: { style: 'thin', color: { argb: 'FFCBD5E1' } }
      }
    }
    footerRow.height = 25

    // Set Column Widths (Auto-fitted with safety margin)
    wsDetail.columns = [
      { width: 6 },   // No
      { width: 18 },  // Tgl Bayar
      { width: 22 },  // Waktu ACC
      { width: 24 },  // No. Ref
      { width: 24 },  // PT
      { width: 20 },  // Cabang
      { width: 32 },  // Nama Toko
      { width: 24 },  // Salesman
      { width: 14 },  // Metode
      { width: 22 },  // Nominal
      { width: 22 },  // Kasir
      { width: 30 }   // Catatan
    ]

    // ==========================================
    // SHEET 2: Ringkasan per PT (Executive Summary)
    // ==========================================
    const wsSummary = workbook.addWorksheet('Ringkasan per PT', {
      views: [{ showGridLines: true }]
    })

    wsSummary.mergeCells('A1:G1')
    const summaryHeader = wsSummary.getCell('A1')
    summaryHeader.value = `RINGKASAN PENDAPATAN SALES PER PERUSAHAAN (PT) - ${activeBranchName.toUpperCase()}`
    summaryHeader.font = { name: 'Segoe UI', size: 13, bold: true, color: { argb: 'FFFFFFFF' } }
    summaryHeader.alignment = { vertical: 'middle', horizontal: 'center' }
    summaryHeader.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF0F766E' } }
    wsSummary.getRow(1).height = 28

    wsSummary.mergeCells('A2:G2')
    const summarySub = wsSummary.getCell('A2')
    summarySub.value = `Periode: ${formattedReportDate}`
    summarySub.font = { name: 'Segoe UI', size: 10, italic: true, color: { argb: 'FF64748B' } }
    summarySub.alignment = { vertical: 'middle', horizontal: 'center' }
    wsSummary.getRow(2).height = 18

    const summaryColHeaders = [
      'NO',
      'NAMA PERUSAHAAN (PT)',
      'KODE PT',
      'TOTAL TUNAI (RP)',
      'TOTAL TRANSFER (RP)',
      'TOTAL PENDAPATAN (RP)',
      'JUMLAH NOTA'
    ]

    const sumHeadRow = wsSummary.getRow(4)
    summaryColHeaders.forEach((sh, idx) => {
      const cell = sumHeadRow.getCell(idx + 1)
      cell.value = sh
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E293B' } }
      cell.font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FFFFFFFF' } }
      cell.alignment = {
        vertical: 'middle',
        horizontal: idx === 0 || idx === 2 || idx === 6 ? 'center' : (idx >= 3 ? 'right' : 'left')
      }
      cell.border = {
        top: { style: 'thin', color: { argb: 'FF94A3B8' } },
        bottom: { style: 'medium', color: { argb: 'FF0F766E' } }
      }
    })
    sumHeadRow.height = 24

    const companies = reportData.value?.companies || []
    let sumRowIdx = 5

    companies.forEach((comp: any, idx: number) => {
      const row = wsSummary.getRow(sumRowIdx)
      const isEven = idx % 2 === 0
      const bgColor = isEven ? 'FFFFFFFF' : 'FFF8FAFC'

      row.getCell(1).value = idx + 1
      row.getCell(2).value = comp.company_name || 'Tanpa PT'
      row.getCell(3).value = comp.company_code || '-'
      row.getCell(4).value = Number(comp.total_cash) || 0
      row.getCell(5).value = Number(comp.total_transfer) || 0
      row.getCell(6).value = Number(comp.total_amount) || 0
      row.getCell(7).value = comp.transaction_count || 0

      for (let c = 1; c <= 7; c++) {
        const cell = row.getCell(c)
        cell.font = { name: 'Segoe UI', size: 10 }
        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: bgColor } }
        cell.border = {
          top: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          left: { style: 'thin', color: { argb: 'FFE2E8F0' } },
          right: { style: 'thin', color: { argb: 'FFE2E8F0' } }
        }
      }

      row.getCell(1).alignment = { vertical: 'middle', horizontal: 'center' }
      row.getCell(2).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF0F172A' } }
      row.getCell(3).alignment = { vertical: 'middle', horizontal: 'center' }

      row.getCell(4).numFmt = '"Rp"#,##0'
      row.getCell(4).alignment = { vertical: 'middle', horizontal: 'right' }
      row.getCell(4).font = { name: 'Segoe UI', size: 10, color: { argb: 'FF047857' } }

      row.getCell(5).numFmt = '"Rp"#,##0'
      row.getCell(5).alignment = { vertical: 'middle', horizontal: 'right' }
      row.getCell(5).font = { name: 'Segoe UI', size: 10, color: { argb: 'FF4338CA' } }

      row.getCell(6).numFmt = '"Rp"#,##0'
      row.getCell(6).alignment = { vertical: 'middle', horizontal: 'right' }
      row.getCell(6).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF0F172A' } }

      row.getCell(7).alignment = { vertical: 'middle', horizontal: 'center' }

      row.height = 22
      sumRowIdx++
    })

    // Total Row in Summary
    const sumTotalRow = wsSummary.getRow(sumRowIdx)
    wsSummary.mergeCells(`A${sumRowIdx}:C${sumRowIdx}`)
    sumTotalRow.getCell(1).value = 'TOTAL KESELURUHAN:'
    sumTotalRow.getCell(1).font = { name: 'Segoe UI', size: 10, bold: true }
    sumTotalRow.getCell(1).alignment = { vertical: 'middle', horizontal: 'right' }
    sumTotalRow.getCell(1).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    sumTotalRow.getCell(4).value = { formula: `SUM(D5:D${sumRowIdx - 1})`, result: totalCash }
    sumTotalRow.getCell(4).numFmt = '"Rp"#,##0'
    sumTotalRow.getCell(4).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF047857' } }
    sumTotalRow.getCell(4).alignment = { vertical: 'middle', horizontal: 'right' }
    sumTotalRow.getCell(4).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    sumTotalRow.getCell(5).value = { formula: `SUM(E5:E${sumRowIdx - 1})`, result: totalTransfer }
    sumTotalRow.getCell(5).numFmt = '"Rp"#,##0'
    sumTotalRow.getCell(5).font = { name: 'Segoe UI', size: 10, bold: true, color: { argb: 'FF4338CA' } }
    sumTotalRow.getCell(5).alignment = { vertical: 'middle', horizontal: 'right' }
    sumTotalRow.getCell(5).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    sumTotalRow.getCell(6).value = { formula: `SUM(F5:F${sumRowIdx - 1})`, result: grandTotal }
    sumTotalRow.getCell(6).numFmt = '"Rp"#,##0'
    sumTotalRow.getCell(6).font = { name: 'Segoe UI', size: 11, bold: true, color: { argb: 'FF047857' } }
    sumTotalRow.getCell(6).alignment = { vertical: 'middle', horizontal: 'right' }
    sumTotalRow.getCell(6).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD1FAE5' } }

    sumTotalRow.getCell(7).value = { formula: `SUM(G5:G${sumRowIdx - 1})`, result: rows.length }
    sumTotalRow.getCell(7).font = { name: 'Segoe UI', size: 10, bold: true }
    sumTotalRow.getCell(7).alignment = { vertical: 'middle', horizontal: 'center' }
    sumTotalRow.getCell(7).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }

    for (let c = 1; c <= 7; c++) {
      sumTotalRow.getCell(c).border = {
        top: { style: 'medium', color: { argb: 'FF0F766E' } },
        bottom: { style: 'medium', color: { argb: 'FF0F766E' } },
        left: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        right: { style: 'thin', color: { argb: 'FFCBD5E1' } }
      }
    }
    sumTotalRow.height = 25

    wsSummary.columns = [
      { width: 6 },   // No
      { width: 34 },  // Nama Perusahaan
      { width: 14 },  // Kode PT
      { width: 22 },  // Total Tunai
      { width: 22 },  // Total Transfer
      { width: 24 },  // Total Pendapatan
      { width: 16 }   // Jumlah Nota
    ]

    // ==========================================
    // Save as .XLSX
    // ==========================================
    const buffer = await workbook.xlsx.writeBuffer()
    const fileName = `Rekap_Pendapatan_Sales_Pabrik_${selectedDate.value}.xlsx`
    saveAs(new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }), fileName)
    toast.success('File Excel (.xlsx) berhasil diunduh')
  } catch (error) {
    console.error('Gagal mengekspor data ke Excel:', error)
    toast.error('Gagal mengekspor data ke format Excel')
  } finally {
    isExporting.value = false
  }
}

// Watchers
watch([selectedDate, selectedCompanyId, selectedBranchId], () => {
  if (selectedBranchId.value) {
    fetchDailyIncome()
  }
})

watch(() => masterDataStore.branches, (branches) => {
  if (branches.length > 0 && !selectedBranchId.value) {
    const pabrik = branches.find((b: any) => (b.name || '').toLowerCase().includes('pabrik'))
    if (pabrik) {
      selectedBranchId.value = pabrik.id
    }
  }
}, { immediate: true })

onMounted(async () => {
  await Promise.all([
    masterDataStore.fetchCompanies(),
    masterDataStore.fetchBranches()
  ])
  if (!selectedBranchId.value) {
    const pabrik = masterDataStore.branches.find((b: any) => (b.name || '').toLowerCase().includes('pabrik'))
    if (pabrik) {
      selectedBranchId.value = pabrik.id
    }
  }
  fetchDailyIncome()
})
</script>

<template>
  <div class="p-4 sm:p-6 lg:p-8 space-y-6 max-w-7xl mx-auto">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm">
      <div class="space-y-1">
        <div class="flex items-center gap-2.5">
          <div class="p-2.5 bg-emerald-50 text-emerald-600 rounded-xl border border-emerald-100">
            <Banknote class="w-6 h-6" />
          </div>
          <div>
            <h1 class="text-xl sm:text-2xl font-black text-slate-800 tracking-tight">
              Pendapatan Sales Harian
            </h1>
            <p class="text-xs sm:text-sm text-slate-500 font-medium">
              Rekapitulasi setoran tunai & transfer sales terverifikasi kasir per PT dan cabang
            </p>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex flex-wrap items-center gap-2.5">
        <button
          @click="fetchDailyIncome"
          :disabled="isLoading"
          class="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold text-slate-700 bg-slate-50 hover:bg-slate-100 border border-slate-200 rounded-xl transition-all disabled:opacity-50"
        >
          <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': isLoading }" />
          <span>Segarkan</span>
        </button>

        <button
          @click="exportToExcel"
          :disabled="isLoading || isExporting || filteredPayments.length === 0"
          class="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold text-white bg-emerald-600 hover:bg-emerald-700 rounded-xl shadow-sm transition-all disabled:opacity-50"
        >
          <RefreshCw v-if="isExporting" class="w-3.5 h-3.5 animate-spin" />
          <FileSpreadsheet v-else class="w-3.5 h-3.5" />
          <span>{{ isExporting ? 'Menyiapkan Excel...' : 'Ekspor Excel (.xlsx)' }}</span>
        </button>
      </div>
    </div>

    <!-- Filter Control Toolbar -->
    <div class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm space-y-4">
      <div class="flex items-center gap-2 text-xs font-bold text-slate-400 uppercase tracking-wider">
        <Filter class="w-4 h-4 text-emerald-600" />
        <span>Filter Rekapitulasi</span>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3.5">
        <!-- Tanggal Filter -->
        <div class="space-y-1.5">
          <label class="text-xs font-semibold text-slate-600 flex items-center justify-between">
            <span>Tanggal Setoran</span>
            <span class="flex gap-1 text-[11px] font-bold text-emerald-600">
              <button @click="setToday" type="button" class="hover:underline">Hari Ini</button>
              <span>•</span>
              <button @click="setYesterday" type="button" class="hover:underline">Kemarin</button>
            </span>
          </label>
          <div class="relative">
            <input
              type="date"
              v-model="selectedDate"
              class="w-full pl-9 pr-3 py-2 text-xs font-medium text-slate-800 bg-slate-50/50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:border-transparent outline-none transition-all"
            />
            <Calendar class="w-4 h-4 text-slate-400 absolute left-3 top-2.5 pointer-events-none" />
          </div>
        </div>

        <!-- Perusahaan / PT Filter -->
        <div class="space-y-1.5">
          <label class="text-xs font-semibold text-slate-600">Perusahaan (PT)</label>
          <div class="relative">
            <select
              v-model="selectedCompanyId"
              class="w-full pl-9 pr-8 py-2 text-xs font-medium text-slate-800 bg-slate-50/50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:border-transparent outline-none transition-all appearance-none"
            >
              <option value="ALL">Semua PT / Perusahaan</option>
              <option
                v-for="comp in masterDataStore.companies"
                :key="comp.id"
                :value="comp.id"
              >
                {{ comp.name }}
              </option>
            </select>
            <Building2 class="w-4 h-4 text-slate-400 absolute left-3 top-2.5 pointer-events-none" />
            <ChevronDown class="w-3.5 h-3.5 text-slate-400 absolute right-3 top-3 pointer-events-none" />
          </div>
        </div>

        <!-- Cabang Filter (Khusus Pabrik Utama) -->
        <div class="space-y-1.5">
          <label class="text-xs font-semibold text-slate-600 flex items-center justify-between">
            <span>Cabang</span>
            <span class="text-[10px] font-bold text-amber-600 bg-amber-50 px-1.5 py-0.5 rounded border border-amber-200">Khusus Pabrik</span>
          </label>
          <div class="relative">
            <select
              v-model="selectedBranchId"
              class="w-full pl-9 pr-8 py-2 text-xs font-medium text-slate-800 bg-slate-50/50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:border-transparent outline-none transition-all appearance-none"
            >
              <option
                v-for="b in (pabrikBranches.length > 0 ? pabrikBranches : masterDataStore.branches.filter(br => (br.name || '').toLowerCase().includes('pabrik')))"
                :key="b.id"
                :value="b.id"
              >
                {{ b.name }}
              </option>
            </select>
            <MapPin class="w-4 h-4 text-slate-400 absolute left-3 top-2.5 pointer-events-none" />
            <ChevronDown class="w-3.5 h-3.5 text-slate-400 absolute right-3 top-3 pointer-events-none" />
          </div>
        </div>

        <!-- Mode Pencarian Cepat -->
        <div class="space-y-1.5">
          <label class="text-xs font-semibold text-slate-600">Cari Cepat</label>
          <div class="relative">
            <input
              type="text"
              v-model="searchFilter"
              placeholder="Cari toko, salesman, ref..."
              class="w-full pl-9 pr-3 py-2 text-xs font-medium text-slate-800 bg-slate-50/50 border border-slate-200 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:border-transparent outline-none transition-all placeholder:text-slate-400"
            />
            <Search class="w-4 h-4 text-slate-400 absolute left-3 top-2.5 pointer-events-none" />
          </div>
        </div>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Card 1: Total Setoran Tunai -->
      <div class="bg-gradient-to-br from-emerald-500/10 via-white to-white p-5 rounded-2xl border border-emerald-200/60 shadow-sm relative overflow-hidden group hover:shadow-md transition-all">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-bold text-emerald-700 tracking-wider uppercase">Setoran Tunai (Cash)</span>
          <div class="p-2 bg-emerald-500/10 text-emerald-600 rounded-xl group-hover:scale-110 transition-transform">
            <Banknote class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-800 tracking-tight">
          {{ formatCurrency(reportData.summary.total_cash) }}
        </div>
        <p class="text-[11px] text-slate-400 mt-1 flex items-center gap-1 font-medium">
          <CheckCircle2 class="w-3 h-3 text-emerald-500" />
          Uang fisik terverifikasi kasir
        </p>
      </div>

      <!-- Card 2: Total Setoran Transfer -->
      <div class="bg-gradient-to-br from-indigo-500/10 via-white to-white p-5 rounded-2xl border border-indigo-200/60 shadow-sm relative overflow-hidden group hover:shadow-md transition-all">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-bold text-indigo-700 tracking-wider uppercase">Setoran Transfer (TF)</span>
          <div class="p-2 bg-indigo-500/10 text-indigo-600 rounded-xl group-hover:scale-110 transition-transform">
            <CreditCard class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-800 tracking-tight">
          {{ formatCurrency(reportData.summary.total_transfer) }}
        </div>
        <p class="text-[11px] text-slate-400 mt-1 flex items-center gap-1 font-medium">
          <CheckCircle2 class="w-3 h-3 text-indigo-500" />
          Bank transfer / QRIS terverifikasi
        </p>
      </div>

      <!-- Card 3: Grand Total Pendapatan -->
      <div class="bg-gradient-to-br from-amber-500/10 via-white to-white p-5 rounded-2xl border border-amber-200/60 shadow-sm relative overflow-hidden group hover:shadow-md transition-all">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-bold text-amber-700 tracking-wider uppercase">Grand Total Pendapatan</span>
          <div class="p-2 bg-amber-500/10 text-amber-600 rounded-xl group-hover:scale-110 transition-transform">
            <Wallet class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-800 tracking-tight text-amber-600">
          {{ formatCurrency(reportData.summary.grand_total) }}
        </div>
        <p class="text-[11px] text-slate-400 mt-1 flex items-center gap-1 font-medium">
          <span>Akumulasi Tunai + Transfer</span>
        </p>
      </div>

      <!-- Card 4: Total Transaksi -->
      <div class="bg-gradient-to-br from-slate-500/10 via-white to-white p-5 rounded-2xl border border-slate-200/80 shadow-sm relative overflow-hidden group hover:shadow-md transition-all">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-bold text-slate-600 tracking-wider uppercase">Setoran Disetujui</span>
          <div class="p-2 bg-slate-500/10 text-slate-600 rounded-xl group-hover:scale-110 transition-transform">
            <Receipt class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-800 tracking-tight">
          {{ reportData.summary.total_transactions }}
          <span class="text-xs font-normal text-slate-400">transaksi</span>
        </div>
        <p class="text-[11px] text-slate-400 mt-1 flex items-center gap-1 font-medium">
          <span>{{ formatDateOnly(selectedDate) }}</span>
        </p>
      </div>
    </div>

    <!-- Rekapitulasi Per PT (Company Grouping Cards) -->
    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-base sm:text-lg font-black text-slate-800 tracking-tight flex items-center gap-2">
            <Building2 class="w-5 h-5 text-emerald-600" />
            <span>Rekapitulasi Pendapatan Per PT</span>
          </h2>
          <p class="text-xs text-slate-400 font-medium">
            Rincian akumulasi setoran tunai & transfer pada setiap entitas perusahaan
          </p>
        </div>
      </div>

      <div v-if="isLoading" class="bg-white p-8 rounded-2xl border border-slate-200 text-center space-y-2">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-emerald-600"></div>
        <p class="text-xs text-slate-500 font-medium">Sedang memuat data per PT...</p>
      </div>

      <div v-else-if="!reportData.companies || reportData.companies.length === 0" class="bg-white p-8 rounded-2xl border border-slate-200 text-center">
        <AlertCircle class="w-8 h-8 text-slate-300 mx-auto mb-2" />
        <p class="text-sm font-semibold text-slate-600">Tidak ada data perusahaan untuk filter ini</p>
      </div>

      <!-- List of Company Cards -->
      <div v-else class="space-y-3.5">
        <div
          v-for="comp in reportData.companies"
          :key="comp.company_id"
          class="bg-white rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden transition-all"
        >
          <!-- Company Card Header -->
          <div
            @click="toggleCompanyExpand(comp.company_id)"
            class="p-5 flex flex-col md:flex-row md:items-center justify-between gap-4 cursor-pointer hover:bg-slate-50/70 transition-colors"
          >
            <div class="flex items-center gap-3">
              <div class="p-2.5 bg-slate-100 text-slate-700 rounded-xl border border-slate-200 shrink-0">
                <Building2 class="w-5 h-5" />
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="text-base font-black text-slate-800 tracking-tight">
                    {{ comp.company_name }}
                  </h3>
                  <span
                    v-if="comp.company_code"
                    class="px-2 py-0.5 text-[10px] font-extrabold bg-emerald-50 text-emerald-700 border border-emerald-200 rounded-md"
                  >
                    {{ comp.company_code }}
                  </span>
                </div>
                <p class="text-xs text-slate-400 font-medium mt-0.5">
                  {{ comp.transaction_count || 0 }} transaksi terverifikasi • {{ comp.branches?.length || 0 }} unit cabang
                </p>
              </div>
            </div>

            <!-- Financial Pill Breakdown -->
            <div class="flex flex-wrap items-center gap-2.5 md:gap-4 text-xs">
              <div class="px-3 py-1.5 bg-emerald-50/80 border border-emerald-100 rounded-xl flex items-center gap-1.5">
                <span class="text-emerald-700 font-semibold">Tunai:</span>
                <span class="font-extrabold text-emerald-800">{{ formatCurrency(comp.total_cash) }}</span>
              </div>

              <div class="px-3 py-1.5 bg-indigo-50/80 border border-indigo-100 rounded-xl flex items-center gap-1.5">
                <span class="text-indigo-700 font-semibold">Transfer:</span>
                <span class="font-extrabold text-indigo-800">{{ formatCurrency(comp.total_transfer) }}</span>
              </div>

              <div class="px-3.5 py-1.5 bg-slate-900 text-white rounded-xl flex items-center gap-2 font-bold shadow-sm">
                <span>Total PT:</span>
                <span class="text-emerald-400 font-black">{{ formatCurrency(comp.total_amount) }}</span>
              </div>

              <button
                type="button"
                class="p-1.5 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 transition-colors"
              >
                <ChevronUp v-if="expandedCompanies[comp.company_id]" class="w-4 h-4" />
                <ChevronDown v-else class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Sub-Branch Table (Collapsible) -->
          <div
            v-if="expandedCompanies[comp.company_id]"
            class="border-t border-slate-100 bg-slate-50/40 p-4 sm:p-5"
          >
            <div v-if="!comp.branches || comp.branches.length === 0" class="text-center py-4 text-xs text-slate-400">
              Belum ada setoran tercatat untuk cabang di bawah PT ini pada tanggal terpilih.
            </div>

            <div v-else class="overflow-x-auto rounded-xl border border-slate-200/80 bg-white">
              <table class="w-full text-left text-xs">
                <thead class="bg-slate-50/80 border-b border-slate-200/80 text-[11px] font-bold text-slate-500 uppercase tracking-wider">
                  <tr>
                    <th scope="col" class="py-3 px-4">Cabang</th>
                    <th scope="col" class="py-3 px-4 text-right">Setoran Tunai (Cash)</th>
                    <th scope="col" class="py-3 px-4 text-right">Setoran Transfer (TF)</th>
                    <th scope="col" class="py-3 px-4 text-right">Total Pendapatan Cabang</th>
                    <th scope="col" class="py-3 px-4 text-center">Jumlah Nota</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 font-medium text-slate-700">
                  <tr
                    v-for="branch in comp.branches"
                    :key="branch.branch_id"
                    class="hover:bg-slate-50/80 transition-colors"
                  >
                    <td class="py-3 px-4 font-bold text-slate-800 flex items-center gap-2">
                      <MapPin class="w-3.5 h-3.5 text-slate-400" />
                      <span>{{ branch.branch_name }}</span>
                    </td>
                    <td class="py-3 px-4 text-right font-semibold text-emerald-700">
                      {{ formatCurrency(branch.cash_amount) }}
                    </td>
                    <td class="py-3 px-4 text-right font-semibold text-indigo-700">
                      {{ formatCurrency(branch.transfer_amount) }}
                    </td>
                    <td class="py-3 px-4 text-right font-black text-slate-900">
                      {{ formatCurrency(branch.total_amount) }}
                    </td>
                    <td class="py-3 px-4 text-center">
                      <span class="inline-flex px-2 py-0.5 text-[10px] font-bold rounded-full bg-slate-100 text-slate-700">
                        {{ branch.transaction_count }} nota
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Detailed Transactions Table -->
    <div class="bg-white rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden space-y-4 p-5">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-base sm:text-lg font-black text-slate-800 tracking-tight flex items-center gap-2">
            <Receipt class="w-5 h-5 text-emerald-600" />
            <span>Rincian Setoran Sales Terverifikasi</span>
          </h2>
          <p class="text-xs text-slate-400 font-medium">
            Daftar seluruh transaksi setoran yang telah disetujui kasir pada tanggal terpilih
          </p>
        </div>

        <!-- Filter Method Badge Selector -->
        <div class="flex items-center gap-1.5 p-1 bg-slate-100 rounded-xl text-xs font-semibold">
          <button
            @click="selectedMethodFilter = 'ALL'"
            :class="[selectedMethodFilter === 'ALL' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500 hover:text-slate-700']"
            class="px-2.5 py-1 rounded-lg transition-all"
          >
            Semua ({{ reportData.payments?.length || 0 }})
          </button>
          <button
            @click="selectedMethodFilter = 'CASH'"
            :class="[selectedMethodFilter === 'CASH' ? 'bg-white text-emerald-700 shadow-sm' : 'text-slate-500 hover:text-slate-700']"
            class="px-2.5 py-1 rounded-lg transition-all"
          >
            Tunai
          </button>
          <button
            @click="selectedMethodFilter = 'TRANSFER'"
            :class="[selectedMethodFilter === 'TRANSFER' ? 'bg-white text-indigo-700 shadow-sm' : 'text-slate-500 hover:text-slate-700']"
            class="px-2.5 py-1 rounded-lg transition-all"
          >
            Transfer
          </button>
        </div>
      </div>

      <!-- Table View -->
      <div class="overflow-x-auto rounded-xl border border-slate-200/80">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-50 border-b border-slate-200 text-[11px] font-bold text-slate-500 uppercase tracking-wider">
            <tr>
              <th scope="col" class="py-3 px-3.5 text-center w-12">No</th>
              <th scope="col" class="py-3 px-3.5">Waktu ACC & Tgl</th>
              <th scope="col" class="py-3 px-3.5">No. Referensi</th>
              <th scope="col" class="py-3 px-3.5">Toko / Pelanggan</th>
              <th scope="col" class="py-3 px-3.5">Salesman</th>
              <th scope="col" class="py-3 px-3.5">PT & Cabang</th>
              <th scope="col" class="py-3 px-3.5 text-center">Metode</th>
              <th scope="col" class="py-3 px-3.5 text-right">Nominal</th>
              <th scope="col" class="py-3 px-3.5">Kasir Verifikator</th>
              <th scope="col" class="py-3 px-3.5 text-center">Bukti</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 font-medium text-slate-700">
            <tr v-if="isLoading">
              <td colspan="10" class="py-8 text-center text-slate-400">
                <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-emerald-600 mb-1"></div>
                <div>Memuat rincian setoran...</div>
              </td>
            </tr>

            <tr v-else-if="filteredPayments.length === 0">
              <td colspan="10" class="py-10 text-center text-slate-400 space-y-1">
                <Receipt class="w-8 h-8 text-slate-300 mx-auto" />
                <p class="font-semibold text-slate-600">Tidak ada data setoran yang cocok</p>
                <p class="text-[11px]">Silakan pilih tanggal lain atau periksa filter pencarian Anda</p>
              </td>
            </tr>

            <tr
              v-else
              v-for="(item, idx) in filteredPayments"
              :key="item.id"
              class="hover:bg-slate-50/80 transition-colors"
            >
              <td class="py-3 px-3.5 text-center text-slate-400 font-semibold">
                {{ idx + 1 }}
              </td>

              <td class="py-3 px-3.5 space-y-0.5 whitespace-nowrap">
                <div class="font-bold text-slate-800 flex items-center gap-1.5">
                  <Clock class="w-3.5 h-3.5 text-slate-400" />
                  <span>{{ formatDateTime(item.approved_at || item.payment_date) }}</span>
                </div>
                <div class="text-[10px] text-slate-400">
                  Bayar: {{ formatDateOnly(item.payment_date) }}
                </div>
              </td>

              <td class="py-3 px-3.5 whitespace-nowrap">
                <span class="font-bold text-slate-900 bg-slate-100 px-2 py-0.5 rounded border border-slate-200">
                  {{ item.reference_no || '-' }}
                </span>
              </td>

              <td class="py-3 px-3.5 font-bold text-slate-800">
                <div class="flex items-center gap-1.5">
                  <Store class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                  <span class="truncate max-w-[180px]">{{ item.store_name || '-' }}</span>
                </div>
              </td>

              <td class="py-3 px-3.5 text-slate-700">
                <div class="flex items-center gap-1.5">
                  <User class="w-3.5 h-3.5 text-slate-400 shrink-0" />
                  <span class="truncate max-w-[140px]">{{ item.salesman_name || '-' }}</span>
                </div>
              </td>

              <td class="py-3 px-3.5 space-y-0.5">
                <div class="font-semibold text-slate-800 truncate max-w-[150px]">
                  {{ item.company_name || '-' }}
                </div>
                <div class="text-[10px] text-slate-400 flex items-center gap-1">
                  <MapPin class="w-3 h-3 text-slate-400" />
                  <span>{{ item.branch_name || '-' }}</span>
                </div>
              </td>

              <td class="py-3 px-3.5 text-center whitespace-nowrap">
                <span
                  v-if="item.payment_method?.toUpperCase() === 'CASH' || item.payment_method?.toUpperCase() === 'TUNAI'"
                  class="inline-flex items-center gap-1 px-2.5 py-1 text-[10px] font-bold rounded-lg bg-emerald-50 text-emerald-700 border border-emerald-200"
                >
                  <Banknote class="w-3 h-3" />
                  TUNAI
                </span>
                <span
                  v-else
                  class="inline-flex items-center gap-1 px-2.5 py-1 text-[10px] font-bold rounded-lg bg-indigo-50 text-indigo-700 border border-indigo-200"
                >
                  <CreditCard class="w-3 h-3" />
                  {{ item.payment_method || 'TRANSFER' }}
                </span>
              </td>

              <td class="py-3 px-3.5 text-right font-black text-slate-900 whitespace-nowrap">
                {{ formatCurrency(item.amount) }}
              </td>

              <td class="py-3 px-3.5 whitespace-nowrap text-slate-700">
                <span class="text-[11px] font-medium flex items-center gap-1">
                  <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" />
                  {{ item.approver_name || 'Kasir' }}
                </span>
              </td>

              <td class="py-3 px-3.5 text-center whitespace-nowrap">
                <button
                  v-if="item.transfer_proof_url"
                  @click="openPreviewModal(item.transfer_proof_url, `Bukti: ${item.reference_no}`, item.store_name)"
                  type="button"
                  class="inline-flex items-center gap-1 px-2 py-1 text-[10px] font-bold text-blue-700 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 transition-colors"
                >
                  <ExternalLink class="w-3 h-3" />
                  Lihat
                </button>
                <span v-else class="text-[10px] text-slate-300">-</span>
              </td>
            </tr>
          </tbody>

          <!-- Table Footer Summary -->
          <tfoot v-if="filteredPayments.length > 0" class="bg-slate-50 border-t-2 border-slate-200 text-xs font-black text-slate-900">
            <tr>
              <td colspan="7" class="py-3 px-3.5 text-right uppercase tracking-wider text-slate-600">
                Total Setoran Ditampilkan:
              </td>
              <td class="py-3 px-3.5 text-right text-emerald-700 font-black text-sm whitespace-nowrap">
                {{ formatCurrency(filteredPayments.reduce((acc: number, cur: any) => acc + (cur.amount || 0), 0)) }}
              </td>
              <td colspan="2" class="py-3 px-3.5 text-slate-400 font-semibold text-[11px]">
                {{ filteredPayments.length }} transaksi
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>

    <!-- Image Zoom Preview Modal -->
    <div
      v-if="showPreviewModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/80 backdrop-blur-sm"
      @click.self="closePreviewModal"
    >
      <div class="bg-white rounded-2xl max-w-2xl w-full shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <!-- Modal Header -->
        <div class="px-5 py-4 border-b border-slate-100 flex items-center justify-between">
          <div>
            <h3 class="text-sm font-black text-slate-800">{{ previewTitle }}</h3>
            <p v-if="previewSubtitle" class="text-xs text-slate-400 font-medium">{{ previewSubtitle }}</p>
          </div>

          <div class="flex items-center gap-1.5">
            <button
              @click="zoomScale = Math.min(zoomScale + 0.25, 3.0)"
              class="p-1.5 text-slate-500 hover:text-slate-800 rounded-lg hover:bg-slate-100"
              title="Perbesar"
            >
              <ZoomIn class="w-4 h-4" />
            </button>
            <button
              @click="zoomScale = Math.max(zoomScale - 0.25, 0.5)"
              class="p-1.5 text-slate-500 hover:text-slate-800 rounded-lg hover:bg-slate-100"
              title="Perkecil"
            >
              <ZoomOut class="w-4 h-4" />
            </button>
            <button
              @click="rotationDegrees = (rotationDegrees + 90) % 360"
              class="p-1.5 text-slate-500 hover:text-slate-800 rounded-lg hover:bg-slate-100"
              title="Putar"
            >
              <RotateCw class="w-4 h-4" />
            </button>
            <button
              @click="closePreviewModal"
              class="p-1.5 text-slate-400 hover:text-slate-700 rounded-lg hover:bg-slate-100"
              title="Tutup"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>

        <!-- Modal Body Image View -->
        <div class="p-4 bg-slate-100/70 overflow-auto flex items-center justify-center min-h-[350px]">
          <img
            :src="previewImageUrl"
            alt="Preview Bukti"
            :style="{
              transform: `scale(${zoomScale}) rotate(${rotationDegrees}deg)`,
              transition: 'transform 0.2s ease-in-out'
            }"
            class="max-h-[60vh] max-w-full object-contain rounded-lg shadow-md cursor-grab active:cursor-grabbing"
          />
        </div>
      </div>
    </div>
  </div>
</template>
