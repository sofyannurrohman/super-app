<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import * as XLSX from 'xlsx'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import MarketplaceUnsettledPanel from './MarketplaceUnsettledPanel.vue'
import {
  Wallet,
  Clock,
  CheckCircle2,
  XCircle,
  RotateCcw,
  RefreshCw,
  Search,
  Filter,
  Calendar,
  Package,
  Truck,
  ArrowDownRight,
  TrendingUp,
  Percent,
  FileSpreadsheet,
  Upload,
  AlertCircle,
  AlertTriangle,
  HelpCircle,
  ArrowUpRight,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  Info,
  Layers,
  Sparkles,
  ShieldAlert,
  Box,
  Lock,
  Coins,
  TrendingDown,
  Ban,
} from 'lucide-vue-next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()

const isSyncingFinance = ref(false)
const financeSyncProgress = ref<{ status?: string; platform?: string; phase?: string; synced_count: number; processed_count: number; statements_count: number; paid_statements_count?: number; settled_statements_count?: number; latest_payment_at?: string } | null>(null)
const isSyncingOrderFinance = ref<Record<string, boolean>>({})

// Active Tab: order reconciliation plus independent unpaid TikTok snapshot.
const activeTab = ref<'ALL' | 'SETTLED' | 'HOLD' | 'REFUND_RETURN' | 'TO_BE_PAID'>('ALL')

// Filters
const searchKeyword = ref('')
const selectedMarketplace = ref('ALL')
const selectedReturnStatus = ref('ALL')
const selectedPackingStatus = ref('ALL')
const dateType = ref<'settlement_date' | 'created_at'>('created_at')

const getTodayDate = (): string => {
  const d = new Date()
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
  return d.toISOString().split('T')[0] ?? ''
}

const getSevenDaysAgoDate = (): string => {
  const d = new Date()
  d.setDate(d.getDate() - 7)
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
  return d.toISOString().split('T')[0] ?? ''
}

const getThirtyDaysAgoDate = (): string => {
  const d = new Date()
  d.setDate(d.getDate() - 30)
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset())
  return d.toISOString().split('T')[0] ?? ''
}

const getStartOfMonthDate = (): string => {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
}

const selectedPeriodPreset = ref<'all' | 'today' | '7days' | '30days' | 'this_month' | 'custom'>('this_month')
const dateFrom = ref(getStartOfMonthDate())
const dateTo = ref(getTodayDate())
const currentPage = ref(1)
const pageSize = ref(20)
const isAutoClassifying = ref(false)
const showMoreActionsDropdown = ref(false)
const showAuditDetails = ref(false)

// Modal states
const showDetailModal = ref(false)
const selectedOrder = ref<any | null>(null)

const showRestockModal = ref(false)
const restockTargetOrder = ref<any | null>(null)
const restockNotes = ref('')
const isRestocking = ref(false)

const showImportModal = ref(false)
const showFinanceSyncModal = ref(false)
const selectedSyncDays = ref(30)
const selectedFinanceMarketplace = ref<'tiktok' | 'shopee'>('tiktok')
const importMarketplace = ref('tiktok')
const isImporting = ref(false)
const importProgressText = ref('')
const importRecords = ref<any[]>([])
const importFile = ref<File | null>(null)
const importError = ref<string | null>(null)
const detectedReportType = ref<'TIKTOK_INCOME' | 'TIKTOK_ON_HOLD' | 'SHOPEE' | 'UNKNOWN'>('UNKNOWN')
const importOnlyPackedOrders = ref(true)
const importSummaryStats = ref({
  totalRows: 0,
  totalAmount: 0,
  settledCount: 0,
  holdCount: 0,
  refundCount: 0,
})

const tabs = [
  { id: 'ALL', label: 'Semua Transaksi', icon: Layers },
  { id: 'SETTLED', label: 'Dana Sudah Cair (Settled)', icon: CheckCircle2 },
  { id: 'HOLD', label: 'Tertahan / Belum Verifikasi', icon: Clock },
  { id: 'REFUND_RETURN', label: 'Batal & Retur (Refund & Restock)', icon: RotateCcw },
  { id: 'TO_BE_PAID', label: 'Untuk Dibayar TikTok', icon: Wallet },
]

function formatCurrency(amount: number | undefined, currency = 'IDR') {
  if (amount === undefined || isNaN(amount)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatDate(dateStr: string | null | undefined) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatDateOnly(dateStr: string | null | undefined) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function getSettlementBadge(status: string | undefined) {
  const map: Record<string, { label: string; class: string; desc: string }> = {
    SETTLED: {
      label: 'Dana Cair',
      class: 'bg-emerald-100 text-emerald-800 border border-emerald-300 font-bold',
      desc: 'Dana sudah masuk ke saldo seller / rekening',
    },
    PENDING_SETTLEMENT: {
      label: 'Dana Tertahan (Hold)',
      class: 'bg-amber-100 text-amber-800 border border-amber-300 font-semibold',
      desc: 'Pesanan sedang dikirim / menunggu garansi selesai',
    },
    HOLD: {
      label: 'Dana Tertahan (Hold)',
      class: 'bg-amber-100 text-amber-800 border border-amber-300 font-semibold',
      desc: 'Pesanan sedang dikirim / menunggu garansi selesai',
    },
    ESCROW: {
      label: 'Escrow Platform',
      class: 'bg-yellow-100 text-yellow-800 border border-yellow-300 font-medium',
      desc: 'Dana di-hold di escrow TikTok/Shopee',
    },
    REFUNDED: {
      label: 'Dana Dikembalikan',
      class: 'bg-rose-100 text-rose-800 border border-rose-300 font-semibold',
      desc: 'Pesanan dibatalkan atau dana direfund ke pembeli',
    },
    CANCELLED: {
      label: 'Batal / Refund',
      class: 'bg-rose-100 text-rose-800 border border-rose-300 font-medium',
      desc: 'Pesanan dibatalkan',
    },
    FAILED_DELIVERY: {
      label: 'Gagal Kirim (RTS)',
      class: 'bg-purple-100 text-purple-800 border border-purple-300 font-bold',
      desc: 'Paket gagal terkirim dan kembali ke seller (RTS)',
    },
  }
  return (
    map[status || 'PENDING_SETTLEMENT'] || {
      label: status || 'Pending',
      class: 'bg-slate-100 text-slate-700 border border-slate-200 font-medium',
      desc: '',
    }
  )
}

function isUnverifiedSettlement(order: any): boolean {
  return order.settlement_status === 'SETTLED'
    && !order.finance_synced_at
    && !(Number(order.fee_and_tax_amount) > 0)
    && !(Number(order.platform_fee) > 0)
}

function settlementDisplayAmount(order: any): number {
  const officialAmount = Number(order.settlement_amount)
  if (order.marketplace === 'tiktok' && order.settlement_status === 'SETTLED') {
    return Number.isFinite(officialAmount) ? officialAmount : 0
  }
  return officialAmount || Number(order.total_amount || 0) - Number(order.fee_and_tax_amount || order.platform_fee || 0)
}

function getReturnStatusBadge(status: string | undefined) {
  const map: Record<string, { label: string; class: string }> = {
    NO_RETURN: { label: 'Tidak Ada Retur', class: 'bg-slate-100 text-slate-500 border border-slate-200' },
    PENDING_INSPECTION: {
      label: 'Menunggu Restock Gudang',
      class: 'bg-amber-100 text-amber-800 border border-amber-300 font-semibold',
    },
    RESTOCKED: {
      label: 'Sudah Direstock ke Gudang',
      class: 'bg-teal-100 text-teal-800 border border-teal-300 font-bold',
    },
    DAMAGED_LOSS: {
      label: 'Barang Rusak / Hilang (Loss)',
      class: 'bg-rose-100 text-rose-800 border border-rose-300 font-semibold',
    },
  }
  return map[status || 'NO_RETURN'] || { label: status || '-', class: 'bg-slate-100 text-slate-600 border border-slate-200' }
}

function getPhysicalStatusBadge(order: any) {
  const isPacked =
    order.is_packed ||
    !!order.packed_by ||
    ['PACKED', 'VERIFIED', 'COMPLETED', 'DELIVERED', 'IN_TRANSIT', 'AWAITING_COLLECTION'].includes(order.order_status)
  if (isPacked) {
    return {
      isPacked: true,
      label: 'Sudah Dipack',
      class: 'bg-blue-100 text-blue-800 border border-blue-200 font-bold',
      actor: order.packed_by || (['COMPLETED', 'DELIVERED'].includes(order.order_status) ? 'Selesai / Terkirim' : 'Staff Packing'),
    }
  }
  return {
    isPacked: false,
    label: 'Belum Dipack',
    class: 'bg-slate-100 text-slate-500 border border-slate-200 font-medium',
    actor: null,
  }
}

function getReconciliationBadge(order: any) {
  const status = order.reconciliation_status
  if (status === 'PACKED_RETURN_DISPATCHED' || status === 'RETURN_DISPATCHED') {
    return {
      type: 'DISPATCHED_RETURN',
      label: 'Dipack & Batal (Fisik Keluar)',
      fullLabel: '🚨 Batal: Barang Sudah Keluar Gudang!',
      class: 'bg-rose-600 text-white border border-rose-700 font-bold shadow-sm shadow-rose-200',
      desc: 'Barang fisik sudah keluar gudang namun order dibatalkan. Pantau paket retur & lakukan restock saat fisik diterima kembali.',
      isAlert: true,
    }
  }
  if (status === 'RESTOCKED' || order.restocked_at) {
    return {
      type: 'RESTOCKED',
      label: 'Sudah Direstock',
      fullLabel: '📦 Fisik Sudah Kembali ke Gudang',
      class: 'bg-teal-100 text-teal-800 border border-teal-300 font-bold',
      desc: 'Barang retur fisik telah diverifikasi dan stok gudang sudah ditambah kembali.',
      isAlert: false,
    }
  }
  if ((order.order_status === 'FAILED_DELIVERY' || order.settlement_status === 'FAILED_DELIVERY') && !order.restocked_at) {
    return {
      type: 'FAILED_DELIVERY',
      label: 'Gagal Kirim (RTS)',
      fullLabel: '🚚 Pengantaran Gagal (Kurir RTS)',
      class: 'bg-purple-600 text-white border border-purple-700 font-bold shadow-sm shadow-purple-200',
      desc: 'Paket gagal diantar kurir dan dalam pengembalian (RTS). Restock ke gudang saat barang diterima.',
      isAlert: true,
    }
  }
  if (status === 'RETURN_UNPACKED') {
    return {
      type: 'RETURN_UNPACKED',
      label: 'Batal (Belum Dipack)',
      fullLabel: '⚪ Batal Sebelum Dipack (Aman)',
      class: 'bg-slate-100 text-slate-600 border border-slate-200 font-medium',
      desc: 'Pesanan dibatalkan sebelum barang keluar dari gudang. Tidak ada selisih stok.',
      isAlert: false,
    }
  }
  if (status === 'PACKED_MATCH_SETTLED' || status === 'MATCH_SETTLED') {
    return {
      type: 'MATCH_SETTLED',
      label: 'Dipack & Dana Cair',
      fullLabel: '🟢 Barang Terkirim & Dana Cair',
      class: 'bg-emerald-100 text-emerald-800 border border-emerald-300 font-bold',
      desc: 'Pesanan telah dipack/keluar dan dana pencairan telah cair masuk ke saldo seller.',
      isAlert: false,
    }
  }
  if (status === 'PACKED_IN_TRANSIT_HOLD' || status === 'IN_TRANSIT_HOLD') {
    return {
      type: 'IN_TRANSIT_HOLD',
      label: 'Dipack (Dana Hold)',
      fullLabel: '🟡 Barang Dipack & Dana On-Hold',
      class: 'bg-amber-100 text-amber-800 border border-amber-300 font-medium',
      desc: 'Barang sudah dikirim kurir, dana masih tertahan di escrow menunggu konfirmasi pesanan pembeli.',
      isAlert: false,
    }
  }
  if (status === 'PACKED_UNRECONCILED') {
    return {
      type: 'PACKED_UNRECONCILED',
      label: 'Dipack (Belum Terdata)',
      fullLabel: '⚪ Sudah Dipack (Belum Masuk Laporan TikTok)',
      class: 'bg-slate-100 text-slate-600 border border-slate-200 font-medium',
      desc: 'Pesanan sudah dipack di gudang, namun belum tercatat di laporan Income maupun On-Hold.',
      isAlert: false,
    }
  }
  if (status === 'ANOMALY_SETTLED_UNPACKED') {
    return {
      type: 'ANOMALY',
      label: 'Cair (Belum Scan Pack)',
      fullLabel: '🔵 Dana Cair (Belum Scan Pack di App)',
      class: 'bg-blue-100 text-blue-800 border border-blue-300 font-medium',
      desc: 'Dana sudah cair di marketplace namun belum tercatat scan packing di Super App.',
      isAlert: false,
    }
  }
  return {
    type: 'PENDING',
    label: 'Menunggu Proses',
    fullLabel: 'Menunggu Proses',
    class: 'bg-slate-100 text-slate-600 border-slate-200',
    desc: '',
    isAlert: false,
  }
}

async function loadData(includeSummary = true) {
  await marketplaceStore.fetchSettlementOrders({
    tab: activeTab.value,
    marketplace: selectedMarketplace.value === 'ALL' ? undefined : selectedMarketplace.value,
    return_status: selectedReturnStatus.value === 'ALL' ? undefined : selectedReturnStatus.value,
    packing_status: selectedPackingStatus.value === 'ALL' ? undefined : selectedPackingStatus.value,
    keyword: searchKeyword.value,
    date_type: dateType.value,
    from: dateFrom.value,
    to: dateTo.value,
    page: currentPage.value,
    pageSize: pageSize.value,
  })
  if (includeSummary) {
    await marketplaceStore.fetchSettlementSummary({
      from: dateFrom.value,
      to: dateTo.value,
      marketplace: selectedMarketplace.value === 'ALL' ? undefined : selectedMarketplace.value,
      date_type: dateType.value,
    })
  }
}

function switchTab(tabId: 'ALL' | 'SETTLED' | 'HOLD' | 'REFUND_RETURN' | 'TO_BE_PAID') {
  activeTab.value = tabId
  if (tabId === 'TO_BE_PAID') return
  if (tabId === 'SETTLED') {
    dateType.value = 'settlement_date'
  }
  currentPage.value = 1
  loadData()
}

function handleSearch() {
  currentPage.value = 1
  loadData(false)
}

function applyPeriodPreset(preset: 'all' | 'today' | '7days' | '30days' | 'this_month' | 'custom') {
  selectedPeriodPreset.value = preset
  if (preset === 'all') {
    dateFrom.value = ''
    dateTo.value = ''
  } else if (preset === 'today') {
    dateFrom.value = getTodayDate()
    dateTo.value = getTodayDate()
  } else if (preset === '7days') {
    dateFrom.value = getSevenDaysAgoDate()
    dateTo.value = getTodayDate()
  } else if (preset === '30days') {
    dateFrom.value = getThirtyDaysAgoDate()
    dateTo.value = getTodayDate()
  } else if (preset === 'this_month') {
    dateFrom.value = getStartOfMonthDate()
    dateTo.value = getTodayDate()
  }
  currentPage.value = 1
  loadData()
}

function handleCustomDateChange() {
  selectedPeriodPreset.value = 'custom'
  currentPage.value = 1
  loadData()
}

function resetFilters() {
  searchKeyword.value = ''
  selectedMarketplace.value = 'ALL'
  selectedReturnStatus.value = 'ALL'
  selectedPackingStatus.value = 'ALL'
  dateType.value = 'created_at'
  selectedPeriodPreset.value = 'this_month'
  dateFrom.value = getStartOfMonthDate()
  dateTo.value = getTodayDate()
  currentPage.value = 1
  loadData()
}

const totalPages = computed(() =>
  Math.max(1, Math.ceil(marketplaceStore.settlementOrdersTotal / pageSize.value))
)

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  loadData(false)
}

async function handleAutoClassify() {
  isAutoClassifying.value = true
  try {
    const res = await marketplaceStore.autoClassifySettlements()
    toast.success(res?.message || 'Auto-rekonsiliasi status selesai!')
    await loadData()
  } catch (error: any) {
    toast.error('Gagal menjalankan auto-rekonsiliasi: ' + (error.response?.data?.error || error.message || 'Error tidak diketahui'))
  } finally {
    isAutoClassifying.value = false
  }
}

let financePollingTimer: ReturnType<typeof setTimeout> | null = null

function clearFinancePolling() {
  if (financePollingTimer) clearTimeout(financePollingTimer)
  financePollingTimer = null
}

function pollFinanceStatus(syncId: string, failedAttempts = 0) {
  financePollingTimer = setTimeout(async () => {
    try {
      const result = await marketplaceStore.getFinanceSyncStatus(syncId)
      financeSyncProgress.value = result
      if (result.status === 'processing') {
        pollFinanceStatus(syncId)
        return
      }
      clearFinancePolling()
      isSyncingFinance.value = false
      if (result.status === 'failed') {
        toast.error(`Sinkronisasi finansial gagal: ${result.error || 'Terjadi kesalahan'}`)
      } else if (result.platform === 'tiktok' && result.synced_count > 0 && !result.paid_statements_count && !result.settled_statements_count) {
        toast.warning(`${result.synced_count} rincian pesanan cocok, tetapi tidak ada statement PAID atau SETTLED. Dana cair belum terverifikasi.`)
      } else if (result.synced_count > 0) {
        toast.success(`Sinkronisasi selesai: ${result.synced_count} pesanan unik cocok dari ${result.statements_count} statement${result.platform === 'tiktok' ? ` (${result.settled_statements_count || 0} masuk saldo, ${result.paid_statements_count || 0} transfer PAID)` : ''}.`)
      } else {
        toast.warning(`Tidak ada transaksi yang cocok. Periksa akun toko dan periode pencairan ${selectedFinanceMarketplace.value}.`)
      }
      loadData()
    } catch (error: any) {
      if (failedAttempts < 3) {
        pollFinanceStatus(syncId, failedAttempts + 1)
      } else {
        clearFinancePolling()
        isSyncingFinance.value = false
        toast.error('Status sinkronisasi tidak dapat dibaca. Periksa koneksi lalu muat ulang halaman.')
      }
    }
  }, 3000)
}

async function handleSyncFinance(days?: number | MouseEvent) {
  const syncDays = typeof days === 'number' ? days : selectedSyncDays.value
  isSyncingFinance.value = true
  financeSyncProgress.value = null
  showFinanceSyncModal.value = false
  clearFinancePolling()
  let inBackground = false
  try {
    const res = await marketplaceStore.syncFinanceTransactions(syncDays, selectedFinanceMarketplace.value)
    if (res?.in_background && res.sync_id) {
      inBackground = true
      toast.info(res?.message || 'Sinkronisasi finansial sedang berjalan di latar belakang...')
      pollFinanceStatus(res.sync_id)
    } else if (res?.in_background) {
      throw new Error('Server tidak mengirim ID sinkronisasi')
    } else {
      toast.success(res?.message || 'Sinkronisasi finansial selesai!')
      await loadData()
    }
  } catch (error: any) {
    if (error?.code === 'ECONNABORTED' || (error?.message && error.message.includes('timeout'))) {
      toast.error('Waktu penarikan transaksi finansial habis (timeout). Silakan coba ulangi kembali.')
    } else {
      toast.error('Gagal menarik transaksi finansial: ' + (error.response?.data?.error || error.message || 'Error tidak diketahui'))
    }
  } finally {
    if (!inBackground) isSyncingFinance.value = false
  }
}

async function handleSyncOrderFinance(orderId: string, marketplace: string = 'tiktok') {
  isSyncingOrderFinance.value[orderId] = true
  try {
    const res = await marketplaceStore.syncOrderFinance(orderId, marketplace)
    if (marketplace === 'tiktok' && (res.settlement_amount_provided === false || res.fee_and_tax_provided === false)) {
      toast.warning(res.message)
    } else if (marketplace === 'tiktok' && res.fee_and_tax_provided && res.fee_and_tax_amount === 0) {
      toast.info(res.message)
    } else {
      toast.success(res?.message || `Berhasil menarik finansial pesanan ${orderId}`)
    }
    await loadData()
    if (selectedOrder.value && selectedOrder.value.order_id === orderId) {
      if (res.settlement_amount_provided !== false) selectedOrder.value.settlement_amount = res.settlement_amount
      if (res.fee_and_tax_provided !== false) selectedOrder.value.fee_and_tax_amount = res.fee_and_tax_amount
      selectedOrder.value.shipping_cost_amount = res.shipping_cost_amount
      selectedOrder.value.settlement_status = res.settlement_status
      selectedOrder.value.fee_tax_breakdown = res.fee_tax_breakdown
      selectedOrder.value.finance_synced_at = new Date().toISOString()
    }
  } catch (error: any) {
    toast.error('Gagal menarik finansial pesanan: ' + (error.response?.data?.error || error.message || 'Error tidak diketahui'))
  } finally {
    isSyncingOrderFinance.value[orderId] = false
  }
}

// ───────────────────────────────────────────────
// Restock Action
// ───────────────────────────────────────────────
function openRestockModal(order: any) {
  restockTargetOrder.value = order
  restockNotes.value = `Pengembalian fisik barang pesanan ${order.order_id}`
  showRestockModal.value = true
}

async function executeRestock() {
  if (!restockTargetOrder.value) return
  isRestocking.value = true
  try {
    const res = await marketplaceStore.restockOrder(
      restockTargetOrder.value.order_id,
      restockNotes.value,
    )
    toast.success(res.message || 'Barang berhasil di-restock ke inventori gudang!')
    showRestockModal.value = false
    restockTargetOrder.value = null
    await loadData()
  } catch (err: any) {
    toast.error(err.response?.data?.error || err.message || 'Gagal memproses restock')
  } finally {
    isRestocking.value = false
  }
}

// ───────────────────────────────────────────────
// Detail Read-Only Modal
// ───────────────────────────────────────────────
function openOrderDetail(order: any) {
  selectedOrder.value = order
  showDetailModal.value = true
}

// ───────────────────────────────────────────────
// Settlement File Import Parser (CSV / TSV / Spreadsheet)
// ───────────────────────────────────────────────
function openImportModal() {
  importFile.value = null
  importRecords.value = []
  importError.value = null
  detectedReportType.value = 'UNKNOWN'
  importOnlyPackedOrders.value = true
  importSummaryStats.value = {
    totalRows: 0,
    totalAmount: 0,
    settledCount: 0,
    holdCount: 0,
    refundCount: 0,
  }
  showImportModal.value = true
}

function handleFileUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  importFile.value = file
  importError.value = null
  importRecords.value = []
  detectedReportType.value = 'UNKNOWN'

  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const buffer = e.target?.result as ArrayBuffer
      parseSettlementBuffer(buffer)
    } catch (err: any) {
      importError.value = 'Gagal membaca berkas: ' + err.message
    }
  }
  reader.readAsArrayBuffer(file)
}

function parseCurrencyNumber(val: any): number {
  if (val === undefined || val === null) return 0
  if (typeof val === 'number') return isNaN(val) ? 0 : val
  let s = String(val).trim().replace(/^Rp\s?/i, '').replace(/["']/g, '')
  if (!s || s === '/' || s === '-') return 0
  // Check if negative in parentheses e.g. (15.000)
  const isNegative = s.startsWith('-') || (s.startsWith('(') && s.endsWith(')'))
  s = s.replace(/[()]/g, '')

  if (s.includes('.') && s.includes(',')) {
    // Format 1.234.567,89 (ID/EU format)
    if (s.lastIndexOf(',') > s.lastIndexOf('.')) {
      s = s.replace(/\./g, '').replace(',', '.')
    } else {
      // Format 1,234,567.89 (US format)
      s = s.replace(/,/g, '')
    }
  } else if (s.includes(',')) {
    const parts = s.split(',')
    if (parts.length === 2 && (parts[1]?.length ?? 0) <= 2) {
      s = s.replace(',', '.')
    } else {
      s = s.replace(/,/g, '')
    }
  } else if (s.includes('.')) {
    const parts = s.split('.')
    if (parts.length > 2 || (parts.length === 2 && (parts[1]?.length ?? 0) === 3)) {
      s = s.replace(/\./g, '')
    }
  }

  const num = parseFloat(s.replace(/[^0-9.-]/g, '')) || 0
  return isNegative ? -Math.abs(num) : num
}

function parseDateValue(rawVal: any): string {
  if (!rawVal || rawVal === '/' || rawVal === '-') {
    return ''
  }
  if (rawVal instanceof Date && !isNaN(rawVal.getTime())) {
    return rawVal.toISOString()
  }
  const s = String(rawVal).trim()
  if (!s || s === '/' || s === '-') return ''

  // Replace slash or dot with dash for YYYY/MM/DD
  const normalized = s.replace(/\//g, '-')
  const parsed = new Date(normalized)
  if (!isNaN(parsed.getTime())) {
    return parsed.toISOString()
  }
  return ''
}

function fixWorksheetRange(ws: any) {
  if (!ws) return
  let minR = Infinity, maxR = -1
  let minC = Infinity, maxC = -1
  for (const cell in ws) {
    if (cell[0] === '!') continue
    const decoded = XLSX.utils.decode_cell(cell)
    if (decoded.r < minR) minR = decoded.r
    if (decoded.r > maxR) maxR = decoded.r
    if (decoded.c < minC) minC = decoded.c
    if (decoded.c > maxC) maxC = decoded.c
  }
  if (maxR >= 0 && maxC >= 0) {
    ws['!ref'] = XLSX.utils.encode_range({
      s: { r: minR, c: minC },
      e: { r: maxR, c: maxC },
    })
  }
}

function parseSettlementBuffer(buffer: ArrayBuffer) {
  const workbook = XLSX.read(buffer, { type: 'array', cellDates: true })
  if (!workbook.SheetNames || workbook.SheetNames.length === 0) {
    importError.value = 'Berkas tidak memiliki lembar data (sheet).'
    return
  }

  // Pilih sheet yang tepat: prioritaskan sheet dengan nama 'detail pesanan', 'pesanan', 'order', atau 'transaksi'
  let targetSheetName = workbook.SheetNames[0]
  for (const name of workbook.SheetNames) {
    const lower = name.toLowerCase()
    if (
      lower.includes('detail pesanan') ||
      lower.includes('pesanan') ||
      lower.includes('order') ||
      lower.includes('transaksi')
    ) {
      targetSheetName = name
      break
    }
  }

  if (!targetSheetName) {
    importError.value = 'Lembar data tidak ditemukan.'
    return
  }
  const sheet = workbook.Sheets[targetSheetName]
  if (!sheet) {
    importError.value = 'Lembar data kosong.'
    return
  }

  // Kalibrasi ulang jangkauan sheet agar tidak terpotong oleh metadata dimension XML TikTok yang keliru
  fixWorksheetRange(sheet)

  const rawRows: any[][] = XLSX.utils.sheet_to_json(sheet, { header: 1, defval: '' })

  if (rawRows.length < 2) {
    importError.value = 'Berkas kosong atau tidak memiliki baris data.'
    return
  }

  // Detect format & locate Header row
  let headerRowIdx = -1
  let isTikTokOnHold = false
  let isTikTokIncome = false

  for (let r = 0; r < Math.min(10, rawRows.length); r++) {
    const row = rawRows[r]
    if (!row) continue
    const rowStr = row.map((c) => String(c || '').toLowerCase()).join(' ')

    // Check for On-Hold header
    if (
      rowStr.includes('estimasi jumlah penyelesaian') ||
      rowStr.includes('alasan penunggakan') ||
      rowStr.includes('estimasi waktu penyelesaian') ||
      rowStr.includes('total transaksi')
    ) {
      if (rowStr.includes('id pesanan') || rowStr.includes('jenis transaksi') || rowStr.includes('estimasi jumlah')) {
        headerRowIdx = r
        isTikTokOnHold = true
        break
      }
    }

    // Check for Income header
    if (
      rowStr.includes('jumlah penyelesaian pembayaran') ||
      rowStr.includes('total pendapatan') ||
      rowStr.includes('subtotal setelah diskon') ||
      rowStr.includes('biaya komisi platform')
    ) {
      if (rowStr.includes('id pesanan') || rowStr.includes('penyesuaian') || rowStr.includes('jenis transaksi')) {
        headerRowIdx = r
        isTikTokIncome = true
        break
      }
    }

    // Generic fallback header detection
    if (
      rowStr.includes('order id') ||
      rowStr.includes('no. pesanan') ||
      rowStr.includes('settlement') ||
      rowStr.includes('payout')
    ) {
      headerRowIdx = r
      break
    }
  }

  if (headerRowIdx === -1) {
    headerRowIdx = 0
  }

  if (isTikTokOnHold) {
    detectedReportType.value = 'TIKTOK_ON_HOLD'
  } else if (isTikTokIncome) {
    detectedReportType.value = 'TIKTOK_INCOME'
  } else if (importMarketplace.value === 'shopee') {
    detectedReportType.value = 'SHOPEE'
  } else {
    detectedReportType.value = 'TIKTOK_INCOME' // Default fallback
  }

  const rawHeaderRow = rawRows[headerRowIdx] || []
  const headers = rawHeaderRow.map((h: any) =>
    String(h || '')
      .trim()
      .replace(/^["']|["']$/g, '')
      .toLowerCase()
  )

  // 1. Order ID Column
  let orderIdIdx = headers.findIndex(
    (h) =>
      h.includes('id pesanan/penyesuaian') ||
      h.includes('id pesanan terkait') ||
      h.includes('id pesanan') ||
      h.includes('order id') ||
      h.includes('no. pesanan') ||
      h.includes('nomor pesanan') ||
      h.includes('order sn') ||
      h.includes('order_id') ||
      h === 'order' ||
      h === 'pesanan'
  )

  // 2. On-Hold specific columns
  let estAmountIdx = headers.findIndex((h) => h.includes('estimasi jumlah penyelesaian'))
  let estNotesIdx = headers.findIndex((h) => h.includes('estimasi waktu penyelesaian'))
  let holdReasonIdx = headers.findIndex((h) => h.includes('alasan penunggakan'))
  let shippedDateIdx = headers.findIndex((h) => h.includes('waktu pengiriman pesanan') || h.includes('shipped time'))
  let deliveredDateIdx = headers.findIndex((h) => h.includes('tanggal pengantaran pesanan') || h.includes('delivered time'))

  // 3. Income specific columns
  let grossIdx = headers.findIndex(
    (h) =>
      h.includes('subtotal setelah diskon penjual') ||
      h.includes('subtotal produk') ||
      h.includes('total pendapatan') ||
      h.includes('harga pesanan') ||
      h.includes('subtotal sebelum diskon') ||
      h.includes('total pesanan')
  )

  let payoutIdx = headers.findIndex(
    (h) =>
      h.includes('jumlah penyelesaian pembayaran') ||
      h.includes('total pendapatan penyelesaian') ||
      h.includes('penyelesaian pembayaran') ||
      h.includes('settlement amount') ||
      h.includes('net settlement') ||
      h.includes('penghasilan bersih') ||
      h.includes('total yang dilepaskan') ||
      h.includes('net payout') ||
      h.includes('payout')
  )

  let refundIdx = headers.findIndex(
    (h) =>
      h.includes('total pengembalian dana setelah diskon') ||
      h.includes('total pengembalian dana') ||
      h.includes('pengembalian dana pembeli') ||
      h.includes('refund amount')
  )

  let totalBiayaIdx = headers.findIndex((h) => h === 'total biaya' || h === 'total fee' || h === 'total_biaya' || h === 'biaya platform')
  let feeIndices: number[] = []
  if (totalBiayaIdx !== -1) {
    feeIndices = [totalBiayaIdx]
  } else {
    headers.forEach((h, idx) => {
      if (
        (h.includes('biaya komisi platform') ||
          h.includes('biaya pembayaran') ||
          h.includes('biaya pemrosesan') ||
          h.includes('biaya platform') ||
          h.includes('biaya layanan')) &&
        !h.includes('afiliasi') &&
        !h.includes('kreator') &&
        !h.includes('ongkir') &&
        !h.includes('shipping')
      ) {
        feeIndices.push(idx)
      }
    })
  }

  let commIdx = headers.findIndex(
    (h) =>
      h.includes('komisi afiliasi') ||
      h.includes('komisi mitra afiliasi') ||
      h.includes('komisi kreator') ||
      h.includes('affiliate commission')
  )

  let dateIdx = headers.findIndex(
    (h) =>
      h.includes('waktu pembayaran pesanan') ||
      h.includes('waktu penyelesaian') ||
      h.includes('tanggal penyelesaian') ||
      h.includes('settlement date') ||
      h.includes('waktu pembuatan transaksi') ||
      h.includes('waktu pemesanan') ||
      h.includes('tanggal pencairan') ||
      h.includes('created time')
  )

  if (orderIdIdx === -1) orderIdIdx = 0

  const orderMap = new Map<string, any>()
  let statSettled = 0
  let statHold = 0
  let statRefund = 0
  let statTotalAmount = 0

  for (let i = headerRowIdx + 1; i < rawRows.length; i++) {
    const row = rawRows[i]
    if (!row || row.length === 0) continue

    const rawOrderId = String(row[orderIdIdx] || '').trim()
    if (!rawOrderId) continue

    const orderId = rawOrderId.replace(/[^a-zA-Z0-9_-]/g, '')
    if (!orderId || orderId.length < 5) continue

    if (isTikTokOnHold) {
      // ──────────────────────────────────────────
      // Mode 1: TikTok On-Hold Report
      // ──────────────────────────────────────────
      const estAmount = estAmountIdx !== -1 ? Math.abs(parseCurrencyNumber(row[estAmountIdx])) : 0
      const estNotes = estNotesIdx !== -1 ? String(row[estNotesIdx] || '').trim() : ''
      const holdReason = holdReasonIdx !== -1 ? String(row[holdReasonIdx] || '').trim() : ''
      const shippedAtStr = shippedDateIdx !== -1 ? parseDateValue(row[shippedDateIdx]) : ''
      const deliveredAtStr = deliveredDateIdx !== -1 ? parseDateValue(row[deliveredDateIdx]) : ''
      const dateStr = dateIdx !== -1 ? parseDateValue(row[dateIdx]) : new Date().toISOString()

      if (orderMap.has(orderId)) {
        const existing = orderMap.get(orderId)
        if (estAmount > 0) existing.settlement_amount = estAmount
        if (holdReason) existing.hold_reason = holdReason
        if (estNotes) existing.estimated_settlement_notes = estNotes
      } else {
        orderMap.set(orderId, {
          order_id: orderId,
          settlement_amount: estAmount,
          platform_fee: 0,
          commission_fee: 0,
          gross_amount: estAmount,
          refund_amount: 0,
          settlement_date: dateStr || new Date().toISOString(),
          status: 'HOLD',
          hold_reason: holdReason || 'Menunggu penyelesaian pesanan',
          estimated_settlement_notes: estNotes || 'Dikirim + 3 hari',
          shipped_at: shippedAtStr || undefined,
          delivered_at: deliveredAtStr || undefined,
          only_packed_orders: importOnlyPackedOrders.value,
        })
        statHold++
        statTotalAmount += estAmount
      }
    } else {
      // ──────────────────────────────────────────
      // Mode 2: TikTok Income / General Report
      // ──────────────────────────────────────────
      let payout = payoutIdx !== -1 ? Math.abs(parseCurrencyNumber(row[payoutIdx])) : 0
      let refundVal = refundIdx !== -1 ? parseCurrencyNumber(row[refundIdx]) : 0
      let absRefund = Math.abs(refundVal)

      let fee = 0
      if (feeIndices.length > 0) {
        for (const fIdx of feeIndices) {
          if (row[fIdx] !== undefined) {
            fee += Math.abs(parseCurrencyNumber(row[fIdx]))
          }
        }
      }

      let comm = commIdx !== -1 ? Math.abs(parseCurrencyNumber(row[commIdx])) : 0
      let gross = grossIdx !== -1 ? Math.abs(parseCurrencyNumber(row[grossIdx])) : 0
      let sDate = dateIdx !== -1 ? parseDateValue(row[dateIdx]) : new Date().toISOString()
      if (!sDate) sDate = new Date().toISOString()

      let status = 'SETTLED'
      if (absRefund > 0 && payout === 0) {
        status = 'REFUNDED'
      }

      if (orderMap.has(orderId)) {
        const existing = orderMap.get(orderId)
        existing.settlement_amount += payout
        existing.platform_fee += fee
        existing.commission_fee += comm
        existing.gross_amount += gross
        if (absRefund > 0) existing.refund_amount += absRefund
      } else {
        orderMap.set(orderId, {
          order_id: orderId,
          gross_amount: gross,
          settlement_amount: payout,
          platform_fee: fee,
          commission_fee: comm,
          refund_amount: absRefund,
          settlement_date: sDate,
          status: status,
          only_packed_orders: importOnlyPackedOrders.value,
        })
        if (status === 'REFUNDED') {
          statRefund++
        } else {
          statSettled++
          statTotalAmount += payout
        }
      }
    }
  }

  const parsed = Array.from(orderMap.values()).map((rec) => {
    // If payout is 0 but gross & fee exist, compute payout = gross - fee - comm
    if (rec.status === 'SETTLED' && rec.settlement_amount === 0 && rec.gross_amount > 0 && rec.platform_fee > 0) {
      rec.settlement_amount = Math.max(0, rec.gross_amount - rec.platform_fee - rec.commission_fee)
    }
    // Set only_packed_orders flag on each record
    rec.only_packed_orders = importOnlyPackedOrders.value
    return rec
  })

  importRecords.value = parsed
  importSummaryStats.value = {
    totalRows: parsed.length,
    totalAmount: statTotalAmount,
    settledCount: statSettled,
    holdCount: statHold,
    refundCount: statRefund,
  }

  if (parsed.length === 0) {
    importError.value =
      'Tidak ada data pesanan valid yang ditemukan. Pastikan berkas memiliki kolom ID Pesanan dan Jumlah Penyelesaian / Estimasi Pembayaran.'
  }
}

async function submitImport() {
  if (importRecords.value.length === 0) return
  isImporting.value = true
  importProgressText.value = 'Mempersiapkan data...'
  try {
    // Ensure all records reflect current only_packed_orders toggle
    const payloadRecords = importRecords.value.map((r) => ({
      ...r,
      only_packed_orders: importOnlyPackedOrders.value,
    }))

    const chunkSize = 500
    let totalUpdated = 0
    const totalCount = payloadRecords.length

    for (let i = 0; i < totalCount; i += chunkSize) {
      const chunk = payloadRecords.slice(i, i + chunkSize)
      const currentBatchEnd = Math.min(i + chunkSize, totalCount)
      importProgressText.value = `Merekonsiliasi ${currentBatchEnd} dari ${totalCount} data...`

      const res = await marketplaceStore.importSettlements({
        marketplace: importMarketplace.value,
        records: chunk,
      })
      totalUpdated += res?.updated_count || 0
    }

    toast.success(`Berhasil merekonsiliasi ${totalUpdated} pesanan!`)
    showImportModal.value = false
    await loadData()
  } catch (err: any) {
    toast.error(err.response?.data?.error || err.message || 'Gagal mengimpor settlement')
  } finally {
    isImporting.value = false
    importProgressText.value = ''
  }
}

function exportToCsv() {
  if (marketplaceStore.settlementOrders.length === 0) {
    toast.error('Tidak ada data yang dapat diekspor')
    return
  }

  const rows: string[][] = [
    [
      'Order ID',
      'Marketplace',
      'Status Order Marketplace',
      'Status Fisik Packing',
      'Petugas Packing',
      'Status Settlement',
      'Status Retur Barang',
      'Tanggal Order',
      'Tanggal Cair / Est. Cair',
      'Alasan Hold / Catatan',
      'Alasan Batal',
      'Pihak Pembatal',
      'Alasan Gagal Kirim (RTS)',
      'Total Nilai Order Asli (Rp)',
      'Biaya Platform/Admin (Rp)',
      'Ongkir Gagal Kirim (Rp)',
      'Biaya Packing (Rp)',
      'Modal HPP (Rp)',
      'Komisi Afiliasi (Rp)',
      'Jumlah Penyelesaian Pembayaran (Rp)',
      'Dana Refund (Rp)',
      'Nama Pembeli',
      'No. Resi',
    ],
  ]

  for (const o of marketplaceStore.settlementOrders) {
    rows.push([
      `"${o.order_id}"`,
      `"${o.marketplace || 'tiktok'}"`,
      `"${o.order_status || '-'}"`,
      `"${o.packed_by ? 'SUDAH DIPACK' : 'BELUM DIPACK'}"`,
      `"${o.packed_by || '-'}"`,
      `"${o.settlement_status || 'PENDING_SETTLEMENT'}"`,
      `"${o.return_status || 'NO_RETURN'}"`,
      `"${o.created_at ? new Date(o.created_at).toISOString().replace('T', ' ').substring(0, 19) : '-'}"`,
      `"${o.settlement_date ? new Date(o.settlement_date).toISOString().replace('T', ' ').substring(0, 19) : (o.estimated_settlement_notes || '-')}"`,
      `"${(o.hold_reason || '').replace(/"/g, '""')}"`,
      `"${(o.cancel_reason || '').replace(/"/g, '""')}"`,
      `"${(o.cancel_user || '').replace(/"/g, '""')}"`,
      `"${(o.fail_delivery_reason || '').replace(/"/g, '""')}"`,
      `${o.total_amount || 0}`,
      `${o.platform_fee || 0}`,
      `${o.failed_delivery_shipping_fee || 0}`,
      `${o.packaging_cost || 0}`,
      `${o.cogs_amount || 0}`,
      `${o.commission_fee || 0}`,
      `${o.settlement_amount || 0}`,
      `${o.refund_amount || 0}`,
      `"${(o.buyer_username || '').replace(/"/g, '""')}"`,
      `"${o.tracking_id || '-'}"`,
    ])
  }

  const csvContent = 'data:text/csv;charset=utf-8,\uFEFF' + rows.map((e) => e.join(',')).join('\n')
  const encodedUri = encodeURI(csvContent)
  const link = document.createElement('a')
  const dateSuffix = dateFrom.value && dateTo.value ? `${dateFrom.value}_sd_${dateTo.value}` : getTodayDate()
  const filterTypeSuffix = dateType.value === 'settlement_date' ? 'tgl_pencairan' : 'tgl_pesanan'
  link.setAttribute('href', encodedUri)
  link.setAttribute('download', `laporan_pencairan_${filterTypeSuffix}_${dateSuffix}.csv`)
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

onMounted(() => {
  loadData()
})

onUnmounted(() => {
  clearFinancePolling()
})

watch(
  () => masterDataStore.selectedBranchId,
  () => {
    currentPage.value = 1
    loadData()
  },
)
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-500">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2.5 rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 text-white shadow-lg shadow-emerald-200">
            <Wallet class="h-6 w-6" />
          </div>
          Pencairan &amp; Keuangan Marketplace
        </h1>
        <p class="text-slate-500 text-xs sm:text-sm font-medium mt-1">
          Tracking dana cair, escrow hold, fee platform, serta restock &amp; refund barang batal/retur
        </p>
      </div>

      <!-- Action Buttons Group (Primary CTA + Secondary Dropdown) -->
      <div class="flex items-center gap-2.5 relative">
        <button
          @click="showFinanceSyncModal = true"
          :disabled="isSyncingFinance"
          class="flex items-center gap-2 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-700 hover:to-teal-700 text-white px-4 py-2.5 rounded-xl font-bold text-xs sm:text-sm shadow-lg shadow-emerald-200 hover:shadow-xl hover:shadow-emerald-300 transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
          title="Tarik data finansial melalui API TikTok atau Shopee"
        >
          <Coins class="h-4 w-4" :class="{ 'animate-spin': isSyncingFinance }" />
          {{ isSyncingFinance ? 'Menyinkronkan...' : 'Tarik Finansial API' }}
        </button>

        <!-- Dropdown Aksi Lainnya -->
        <div class="relative">
          <button
            @click="showMoreActionsDropdown = !showMoreActionsDropdown"
            class="flex items-center gap-1.5 bg-white border border-slate-200 hover:border-slate-300 text-slate-700 px-3.5 py-2.5 rounded-xl font-bold text-xs sm:text-sm shadow-sm hover:bg-slate-50 transition-all"
          >
            <span>Opsi &amp; Aksi</span>
            <ChevronDown class="h-4 w-4 text-slate-400 transition-transform" :class="{ 'rotate-180': showMoreActionsDropdown }" />
          </button>

          <!-- Dropdown Menu -->
          <div
            v-if="showMoreActionsDropdown"
            class="absolute right-0 top-full mt-1.5 w-60 bg-white border border-slate-200 rounded-2xl shadow-xl py-1.5 z-30 animate-in fade-in zoom-in-95 duration-150"
          >
            <button
              @click="openImportModal(); showMoreActionsDropdown = false"
              class="w-full px-4 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-emerald-50 hover:text-emerald-700 flex items-center gap-2.5 transition-colors"
            >
              <Upload class="h-4 w-4 text-emerald-600" />
              Upload Laporan Settlement
            </button>
            <button
              @click="exportToCsv(); showMoreActionsDropdown = false"
              class="w-full px-4 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-blue-50 hover:text-blue-700 flex items-center gap-2.5 transition-colors"
            >
              <FileSpreadsheet class="h-4 w-4 text-blue-600" />
              Ekspor Data (CSV)
            </button>
            <div class="border-t border-slate-100 my-1"></div>
            <button
              @click="handleAutoClassify(); showMoreActionsDropdown = false"
              :disabled="isAutoClassifying"
              class="w-full px-4 py-2 text-left text-xs font-semibold text-slate-700 hover:bg-indigo-50 hover:text-indigo-700 flex items-center gap-2.5 disabled:opacity-50 transition-colors"
            >
              <Sparkles class="h-4 w-4 text-indigo-600" :class="{ 'animate-spin': isAutoClassifying }" />
              {{ isAutoClassifying ? 'Merekonsiliasi...' : 'Auto-Rekonsiliasi Status' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="financeSyncProgress" class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-medium text-emerald-900" role="status">
      <p class="font-bold">{{ financeSyncProgress.phase || 'Sinkronisasi finansial sedang berjalan' }}</p>
      <p class="mt-1">{{ financeSyncProgress.processed_count }} record diperiksa · {{ financeSyncProgress.synced_count }} pesanan unik cocok · {{ financeSyncProgress.statements_count }} statement/pencairan ditemukan<span v-if="financeSyncProgress.platform === 'tiktok'"> · {{ financeSyncProgress.settled_statements_count || 0 }} masuk saldo · {{ financeSyncProgress.paid_statements_count || 0 }} transfer PAID</span></p>
      <p v-if="financeSyncProgress.latest_payment_at" class="mt-1">Pembayaran TikTok terbaru dalam hasil API: {{ formatDateOnly(financeSyncProgress.latest_payment_at) }}</p>
    </div>

    <!-- 4 Executive Summary Metrics Cards (Vibrant Colorful Palette) -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- 1. Dana Cair Bersih (Net Settled) -->
      <Card class="border border-emerald-200/80 bg-gradient-to-br from-emerald-50 via-teal-50/60 to-emerald-100/70 shadow-sm rounded-2xl">
        <CardContent class="p-4 flex items-center gap-3.5">
          <div class="p-3 bg-gradient-to-br from-emerald-500 to-teal-600 text-white rounded-2xl flex-shrink-0 shadow-md shadow-emerald-200">
            <CheckCircle2 class="h-5 w-5" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between">
              <p class="text-[11px] text-emerald-700 font-bold uppercase tracking-wider truncate">Dana Cair Bersih</p>
              <span class="text-[10px] font-bold text-emerald-800 bg-emerald-100/90 px-2 py-0.5 rounded-full border border-emerald-200">Masuk Saldo</span>
            </div>
            <p class="text-xl font-black text-emerald-950 truncate mt-0.5" :title="formatCurrency(marketplaceStore.settlementSummary.total_packed_settled_amount || marketplaceStore.settlementSummary.total_settled_amount)">
              {{ formatCurrency(marketplaceStore.settlementSummary.total_packed_settled_amount || marketplaceStore.settlementSummary.total_settled_amount) }}
            </p>
            <p class="text-[11px] text-emerald-700/80 font-semibold mt-0.5 truncate">
              {{ marketplaceStore.settlementSummary.total_packed_settled_orders || marketplaceStore.settlementSummary.total_settled_orders }} Pesanan | Fee: {{ formatCurrency(marketplaceStore.settlementSummary.total_platform_fees || 0) }}
            </p>
          </div>
        </CardContent>
      </Card>

      <!-- 2. Dana On-Hold (Escrow Hold) -->
      <Card class="border border-amber-200/80 bg-gradient-to-br from-amber-50 via-yellow-50/60 to-amber-100/70 shadow-sm rounded-2xl">
        <CardContent class="p-4 flex items-center gap-3.5">
          <div class="p-3 bg-gradient-to-br from-amber-500 to-yellow-600 text-white rounded-2xl flex-shrink-0 shadow-md shadow-amber-200">
            <Clock class="h-5 w-5" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between">
              <p class="text-[11px] text-amber-700 font-bold uppercase tracking-wider truncate">Tertahan / Belum Verifikasi</p>
              <span class="text-[10px] font-bold text-amber-800 bg-amber-100/90 px-2 py-0.5 rounded-full border border-amber-200">In-Transit</span>
            </div>
            <p class="text-xl font-black text-amber-950 truncate mt-0.5" :title="formatCurrency(marketplaceStore.settlementSummary.total_packed_hold_amount || marketplaceStore.settlementSummary.total_pending_amount)">
              {{ formatCurrency(marketplaceStore.settlementSummary.total_packed_hold_amount || marketplaceStore.settlementSummary.total_pending_amount) }}
            </p>
            <p class="text-[11px] text-amber-700/80 font-semibold mt-0.5 truncate">
              {{ marketplaceStore.settlementSummary.total_packed_hold_orders || marketplaceStore.settlementSummary.total_pending_orders }} Pesanan Sedang Dikirim
            </p>
          </div>
        </CardContent>
      </Card>

      <!-- 3. Batal, RTS & Kerugian -->
      <Card class="border border-rose-200/80 bg-gradient-to-br from-rose-50 via-red-50/60 to-rose-100/70 shadow-sm rounded-2xl">
        <CardContent class="p-4 flex items-center gap-3.5">
          <div class="p-3 bg-gradient-to-br from-rose-500 to-red-600 text-white rounded-2xl flex-shrink-0 shadow-md shadow-rose-200">
            <ShieldAlert class="h-5 w-5" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between">
              <p class="text-[11px] text-rose-700 font-bold uppercase tracking-wider truncate">Batal &amp; RTS</p>
              <span class="text-[10px] font-bold text-rose-800 bg-rose-100/90 px-2 py-0.5 rounded-full border border-rose-200">Refund Buyer</span>
            </div>
            <p class="text-xl font-black text-rose-950 truncate mt-0.5" :title="formatCurrency(marketplaceStore.settlementSummary.total_lost_revenue || 0)">
              -{{ formatCurrency(marketplaceStore.settlementSummary.total_lost_revenue || 0) }}
            </p>
            <p class="text-[11px] text-rose-700/80 font-semibold mt-0.5 truncate">
              {{ (marketplaceStore.settlementSummary.total_cancelled_orders || 0) + (marketplaceStore.settlementSummary.total_failed_delivery_orders || 0) }} Kasus | Beban: -{{ formatCurrency((marketplaceStore.settlementSummary.total_platform_loss || 0) + (marketplaceStore.settlementSummary.total_packaging_loss || 0)) }}
            </p>
          </div>
        </CardContent>
      </Card>

      <!-- 4. Rekonsiliasi Gudang -->
      <Card class="border border-blue-200/80 bg-gradient-to-br from-blue-50 via-indigo-50/60 to-blue-100/70 shadow-sm rounded-2xl">
        <CardContent class="p-4 flex items-center gap-3.5">
          <div class="p-3 bg-gradient-to-br from-blue-500 to-indigo-600 text-white rounded-2xl flex-shrink-0 shadow-md shadow-blue-200">
            <Package class="h-5 w-5" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between">
              <p class="text-[11px] text-blue-700 font-bold uppercase tracking-wider truncate">Fisik Gudang</p>
              <span v-if="(marketplaceStore.settlementSummary.total_dispatched_return_orders || 0) > 0" class="text-[10px] font-bold text-amber-800 bg-amber-100 px-2 py-0.5 rounded-full border border-amber-300 animate-pulse">
                {{ marketplaceStore.settlementSummary.total_dispatched_return_orders }} Butuh Restock
              </span>
              <span v-else class="text-[10px] font-bold text-blue-800 bg-blue-100/90 px-2 py-0.5 rounded-full border border-blue-200">
                Sinkron
              </span>
            </div>
            <p class="text-xl font-black text-blue-950 truncate mt-0.5">
              {{ marketplaceStore.settlementSummary.total_packed_orders || 0 }} Dipack
            </p>
            <p class="text-[11px] text-blue-700/80 font-semibold mt-0.5 truncate">
              Barang Fisik Keluar &amp; Terverifikasi
            </p>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Alert Banner for Dispatched Cancelled/Return Orders -->
    <div
      v-if="(marketplaceStore.settlementSummary.total_dispatched_return_orders || 0) > 0"
      class="p-3.5 bg-gradient-to-r from-amber-50 via-orange-50 to-amber-100/80 border border-amber-300/80 rounded-2xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 shadow-md shadow-amber-100"
    >
      <div class="flex items-center gap-2.5">
        <div class="p-2 bg-gradient-to-br from-amber-500 to-orange-600 text-white rounded-xl shrink-0 shadow-sm shadow-orange-200">
          <AlertTriangle class="h-5 w-5" />
        </div>
        <div>
          <p class="font-bold text-amber-950 text-xs sm:text-sm">
            Perhatian: {{ marketplaceStore.settlementSummary.total_dispatched_return_orders }} pesanan batal sudah pernah keluar gudang (Nilai: {{ formatCurrency(marketplaceStore.settlementSummary.total_dispatched_return_value) }}).
          </p>
          <p class="text-amber-800 text-[11px] sm:text-xs mt-0.5">
            Pastikan fisik retur diterima kembali lalu klik <strong>"Restock ke Gudang"</strong> agar stok kembali sinkron.
          </p>
        </div>
      </div>
      <button
        @click="selectedPackingStatus = 'NEEDS_RESTOCK'; activeTab = 'REFUND_RETURN'; handleSearch()"
        class="px-3.5 py-2 bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700 text-white text-xs font-bold rounded-xl shadow-md shadow-orange-200 transition-all whitespace-nowrap shrink-0 flex items-center gap-1.5"
      >
        <RotateCcw class="h-3.5 w-3.5" />
        Lihat Order Butuh Restock
      </button>
    </div>

    <!-- Tab Navigation (Clean & Prominent) -->
    <div class="bg-slate-100 border border-slate-200/80 rounded-2xl p-1 shadow-inner flex flex-wrap gap-1">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="switchTab(tab.id as any)"
        class="flex-1 min-w-[170px] flex items-center justify-center gap-2 py-2.5 px-3 rounded-xl text-xs transition-all duration-200"
        :class="
          activeTab === tab.id
            ? 'bg-white text-slate-900 shadow-md border border-slate-200 font-black'
            : 'text-slate-600 hover:text-slate-900 font-semibold hover:bg-white/50'
        "
      >
        <component
          :is="tab.icon"
          class="h-4 w-4"
          :class="{
            'text-blue-600': activeTab === tab.id && tab.id === 'ALL',
            'text-emerald-600': activeTab === tab.id && tab.id === 'SETTLED',
            'text-amber-600': activeTab === tab.id && tab.id === 'HOLD',
            'text-rose-600': activeTab === tab.id && tab.id === 'REFUND_RETURN',
            'text-orange-600': activeTab === tab.id && tab.id === 'TO_BE_PAID',
          }"
        />
        {{ tab.label }}
      </button>
    </div>

    <MarketplaceUnsettledPanel v-if="activeTab === 'TO_BE_PAID'" />
    <template v-if="activeTab !== 'TO_BE_PAID'">

    <!-- Collapsible Financial Loss Audit (Khusus Tab Batal & Retur) -->
    <div
      v-if="activeTab === 'REFUND_RETURN'"
      class="bg-white border border-slate-200 rounded-2xl p-3.5 text-slate-900 shadow-xs transition-all"
    >
      <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
        <div class="flex items-center gap-2.5">
          <div class="p-1.5 bg-slate-100 text-slate-700 rounded-lg">
            <ShieldAlert class="h-4 w-4" />
          </div>
          <div>
            <h4 class="text-xs font-black text-slate-900 flex items-center gap-2">
              Audit Kerugian Finansial &amp; Risiko Stok Fisik
            </h4>
            <p class="text-[11px] text-slate-600 mt-0.5">
              Omzet Hilang: <strong class="text-slate-900">-{{ formatCurrency(marketplaceStore.settlementSummary.total_lost_revenue || 0) }}</strong> • Beban Riil RTS/Packing: <strong class="text-slate-900">-{{ formatCurrency((marketplaceStore.settlementSummary.total_platform_loss || 0) + (marketplaceStore.settlementSummary.total_packaging_loss || 0)) }}</strong> • Risiko HPP Belum Restock: <strong class="text-slate-900">{{ formatCurrency(marketplaceStore.settlementSummary.total_dispatched_cogs_loss || 0) }}</strong>
            </p>
          </div>
        </div>
        <button
          @click="showAuditDetails = !showAuditDetails"
          class="text-[11px] font-bold text-slate-700 hover:text-slate-900 bg-slate-100 hover:bg-slate-200/70 border border-slate-200 px-2.5 py-1 rounded-lg transition-colors flex items-center gap-1 shrink-0"
        >
          <span>{{ showAuditDetails ? 'Tutup Rincian' : 'Lihat 3 Lapisan Beban' }}</span>
          <ChevronDown class="h-3 w-3 transition-transform" :class="{ 'rotate-180': showAuditDetails }" />
        </button>
      </div>

      <div v-if="showAuditDetails" class="grid grid-cols-1 md:grid-cols-3 gap-3 mt-3 pt-3 border-t border-slate-100 text-xs animate-in fade-in duration-200">
        <div class="bg-slate-50 rounded-xl p-2.5 border border-slate-200">
          <p class="text-[10px] text-slate-500 font-bold uppercase">1. Kerugian Beban Platform (Cash Outflow)</p>
          <p class="text-lg font-black text-slate-900 mt-0.5">
            -{{ formatCurrency((marketplaceStore.settlementSummary.total_platform_loss || 0) + (marketplaceStore.settlementSummary.total_packaging_loss || 0)) }}
          </p>
          <p class="text-[10px] text-slate-500 mt-0.5">• Ongkir RTS: {{ formatCurrency(marketplaceStore.settlementSummary.total_failed_delivery_fee_loss || 0) }}</p>
          <p class="text-[10px] text-slate-500">• Bahan Kemasan: {{ formatCurrency(marketplaceStore.settlementSummary.total_packaging_loss || 0) }}</p>
        </div>

        <div class="bg-slate-50 rounded-xl p-2.5 border border-slate-200">
          <p class="text-[10px] text-slate-500 font-bold uppercase">2. Risiko Modal HPP Barang Fisik</p>
          <p class="text-lg font-black text-slate-900 mt-0.5">
            {{ formatCurrency(marketplaceStore.settlementSummary.total_dispatched_cogs_loss || 0) }}
          </p>
          <p class="text-[10px] text-slate-500 mt-0.5">• {{ marketplaceStore.settlementSummary.total_dispatched_return_orders || 0 }} Paket barang keluar belum kembali</p>
        </div>

        <div class="bg-slate-50 rounded-xl p-2.5 border border-slate-200">
          <p class="text-[10px] text-slate-500 font-bold uppercase">3. Kehilangan Omzet (Lost GMV)</p>
          <p class="text-lg font-black text-slate-900 mt-0.5">
            -{{ formatCurrency(marketplaceStore.settlementSummary.total_lost_revenue || 0) }}
          </p>
          <p class="text-[10px] text-slate-500 mt-0.5">• {{ marketplaceStore.settlementSummary.total_cancelled_orders || 0 }} Batal + {{ marketplaceStore.settlementSummary.total_failed_delivery_orders || 0 }} RTS</p>
        </div>
      </div>
    </div>

    <!-- Compact Smart Filter Bar -->
    <Card class="border-none shadow-sm">
      <CardContent class="p-3.5">
        <div class="flex flex-col lg:flex-row items-stretch lg:items-center justify-between gap-3">
          <!-- Search & Selects -->
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5 flex-1">
            <!-- Search Keyword -->
            <div class="relative">
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <input
                v-model="searchKeyword"
                @keyup.enter="handleSearch"
                type="text"
                placeholder="Cari Order ID, SKU, Resi..."
                class="w-full pl-9 pr-3 py-2 border border-slate-200 rounded-xl text-xs focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 bg-slate-50 font-medium"
              />
            </div>

            <!-- Marketplace Platform -->
            <div class="relative">
              <select
                v-model="selectedMarketplace"
                @change="handleSearch"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl text-xs focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 bg-slate-50 font-medium text-slate-800"
              >
                <option value="ALL">Semua Marketplace</option>
                <option value="tiktok">TikTok Shop</option>
                <option value="shopee">Shopee</option>
              </select>
            </div>

            <!-- Physical Packing / Reconciliation Status Filter -->
            <div class="relative">
              <select
                v-model="selectedPackingStatus"
                @change="handleSearch"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl text-xs focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 bg-slate-50 font-medium text-slate-800"
              >
                <template v-if="activeTab === 'REFUND_RETURN'">
                  <option value="ALL">Semua Kasus Batal &amp; Retur</option>
                  <option value="FAILED_DELIVERY">🚚 Pengantaran Gagal (RTS Kurir)</option>
                  <option value="CANCELLED_ONLY">🚫 Pesanan Dibatalkan Saja</option>
                  <option value="NEEDS_RESTOCK">🚨 Dipack &amp; Butuh Restock Fisik</option>
                  <option value="RESTOCKED">✓ Sudah Direstock ke Gudang</option>
                </template>
                <template v-else>
                  <option value="ALL">Semua Status Fisik &amp; Rekon</option>
                  <option value="PACKED_SETTLED">🟢 Dipack &amp; Dana Cair</option>
                  <option value="PACKED_HOLD">🟡 Dipack &amp; Dana On-Hold</option>
                  <option value="NEEDS_RESTOCK">🚨 Dipack &amp; Batal (Butuh Restock)</option>
                  <option value="PACKED">📦 Semua Sudah Dipack</option>
                  <option value="RESTOCKED">✓ Sudah Direstock</option>
                </template>
              </select>
            </div>

            <!-- Periode Preset Dropdown -->
            <div class="relative">
              <select
                v-model="selectedPeriodPreset"
                @change="applyPeriodPreset(selectedPeriodPreset as any)"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl text-xs focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 bg-slate-50 font-medium text-slate-800"
              >
                <option value="all">Periode: Semua Data</option>
                <option value="this_month">Periode: Bulan Ini</option>
                <option value="30days">Periode: 30 Hari Terakhir</option>
                <option value="7days">Periode: 7 Hari Terakhir</option>
                <option value="today">Periode: Hari Ini</option>
                <option value="custom">Periode: Tanggal Kustom</option>
              </select>
            </div>
          </div>

          <!-- Right Controls: Custom Dates & Toggle Date Basis -->
          <div class="flex items-center gap-2 flex-wrap shrink-0">
            <template v-if="selectedPeriodPreset === 'custom'">
              <input
                v-model="dateFrom"
                @change="handleCustomDateChange"
                type="date"
                class="px-2.5 py-1.5 border border-slate-200 rounded-xl text-xs bg-slate-50 font-medium"
              />
              <span class="text-xs text-slate-400">-</span>
              <input
                v-model="dateTo"
                @change="handleCustomDateChange"
                type="date"
                class="px-2.5 py-1.5 border border-slate-200 rounded-xl text-xs bg-slate-50 font-medium"
              />
            </template>

            <!-- Toggle Date Basis -->
            <div class="flex items-center bg-slate-100 p-0.5 rounded-xl border border-slate-200">
              <button
                @click="dateType = 'created_at'; handleSearch()"
                class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold transition-all flex items-center gap-1"
                :class="dateType === 'created_at' ? 'bg-white text-slate-900 shadow-xs' : 'text-slate-500 hover:text-slate-800'"
                title="Filter berdasarkan tanggal order dibuat"
              >
                <Calendar class="h-3 w-3" />
                Tgl Order
              </button>
              <button
                @click="dateType = 'settlement_date'; handleSearch()"
                class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold transition-all flex items-center gap-1"
                :class="dateType === 'settlement_date' ? 'bg-white text-emerald-700 shadow-xs font-black' : 'text-slate-500 hover:text-slate-800'"
                title="Filter berdasarkan tanggal dana cair"
              >
                <CheckCircle2 class="h-3 w-3" />
                Tgl Cair
              </button>
            </div>

            <button
              @click="resetFilters"
              class="p-2 rounded-xl border border-slate-200 text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors"
              title="Reset semua filter"
            >
              <RotateCcw class="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Main Table Card -->
    <Card class="border border-slate-200 bg-white shadow-xs rounded-2xl overflow-hidden">
      <CardHeader class="pb-3 border-b border-slate-100 flex flex-row items-center justify-between">
        <div>
          <!-- Header khusus Tab Batal & Retur -->
          <template v-if="activeTab === 'REFUND_RETURN'">
            <CardTitle class="text-base font-black text-slate-900 flex items-center gap-2">
              <ShieldAlert class="h-4 w-4 text-slate-700" />
              Daftar Pesanan Batal, Gagal Antar (RTS) &amp; Audit Kerugian
            </CardTitle>
            <p class="text-xs text-slate-400 mt-0.5">
              Menampilkan {{ marketplaceStore.settlementOrders.length }} dari {{ marketplaceStore.settlementOrdersTotal }} transaksi batal &amp; gagal kirim
            </p>
          </template>
          <!-- Header standar Tab Reguler -->
          <template v-else>
            <CardTitle class="text-base font-black text-slate-900 flex items-center gap-2">
              <TrendingUp class="h-4 w-4 text-slate-700" />
              Daftar Transaksi, Pencairan &amp; Rekonsiliasi Fisik
            </CardTitle>
            <p class="text-xs text-slate-400 mt-0.5">
              Menampilkan {{ marketplaceStore.settlementOrders.length }} dari {{ marketplaceStore.settlementOrdersTotal }} transaksi
            </p>
          </template>
        </div>

        <!-- Summary Stats di Kanan Header -->
        <div v-if="activeTab === 'REFUND_RETURN'" class="flex items-center gap-4 text-xs font-bold">
          <div class="text-right">
            <span class="text-slate-400">Total Nilai Batal: </span>
            <span class="text-slate-900 font-black">-{{ formatCurrency(marketplaceStore.settlementSummary.total_lost_revenue || 0) }}</span>
          </div>
          <div class="text-right">
            <span class="text-slate-400">Beban Biaya RTS/Packing: </span>
            <span class="text-slate-900 font-black">-{{ formatCurrency((marketplaceStore.settlementSummary.total_platform_loss || 0) + (marketplaceStore.settlementSummary.total_packaging_loss || 0)) }}</span>
          </div>
        </div>
        <div v-else class="flex items-center gap-4 text-xs font-bold">
          <div class="text-right">
            <span class="text-slate-400">Total Cair: </span>
            <span class="text-slate-900 font-black">{{ formatCurrency(marketplaceStore.settlementTotalAmount) }}</span>
          </div>
          <div class="text-right">
            <span class="text-slate-400">Total Biaya Platform: </span>
            <span class="text-slate-900 font-black">{{ formatCurrency(marketplaceStore.settlementTotalFee) }}</span>
          </div>
        </div>
      </CardHeader>

      <CardContent class="p-0">
        <!-- Loading State -->
        <div v-if="marketplaceStore.isSettlementLoading" class="flex flex-col items-center justify-center py-20 gap-3">
          <RefreshCw class="h-8 w-8 text-slate-600 animate-spin" />
          <p class="text-sm font-semibold text-slate-500">Memuat rincian settlement &amp; rekonsiliasi...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="marketplaceStore.settlementOrders.length === 0" class="flex flex-col items-center justify-center py-20 gap-3">
          <div class="p-4 bg-slate-100 rounded-2xl">
            <Wallet class="h-10 w-10 text-slate-300" />
          </div>
          <p class="text-slate-700 font-bold">Tidak ada data transaksi yang sesuai filter</p>
          <p class="text-slate-400 text-xs">Coba sesuaikan filter status fisik / tanggal pencairan atau upload laporan settlement.</p>
        </div>

        <!-- ─────────────────────────────────────────────── -->
        <!-- TABEL KHUSUS TAB BATAL, RETUR & RTS -->
        <!-- ─────────────────────────────────────────────── -->
        <div v-else-if="activeTab === 'REFUND_RETURN'" class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead class="bg-slate-50 border-b border-slate-200">
              <tr>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Order ID &amp; Platform</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Rincian Produk &amp; SKU</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Status Fisik Gudang</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Jenis Batal &amp; Alasan</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Rekonsiliasi &amp; Restock</th>
                <th class="text-right py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Nilai Order (Refund)</th>
                <th class="text-right py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Audit Beban Kerugian</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <template v-for="order in marketplaceStore.settlementOrders" :key="order.order_id">
                <tr
                  class="hover:bg-slate-50/80 transition-colors cursor-pointer"
                  :class="getReconciliationBadge(order).isAlert ? 'bg-slate-50/50 border-l-4 border-l-slate-400' : ''"
                  @click="openOrderDetail(order)"
                  title="Klik untuk melihat rincian pesanan"
                >
                  <!-- 1. Order ID & Marketplace -->
                  <td class="py-3 px-4 align-top">
                    <div class="font-bold text-slate-900 flex items-center justify-between">
                      <span class="hover:text-slate-600 transition-colors">{{ order.order_id }}</span>
                      <button
                        @click.stop="handleSyncOrderFinance(order.order_id, order.marketplace)"
                        :disabled="isSyncingOrderFinance[order.order_id]"
                        class="p-1 rounded-md text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors"
                        :title="`Tarik rincian finansial ${order.marketplace || 'TikTok'}`"
                      >
                        <RefreshCw class="h-3 w-3" :class="{ 'animate-spin': isSyncingOrderFinance[order.order_id] }" />
                      </button>
                    </div>
                    <div class="flex items-center gap-1 mt-1">
                      <span
                        class="px-2 py-0.5 rounded-md text-[10px] font-black uppercase tracking-wider"
                        :class="order.marketplace === 'tiktok' ? 'bg-slate-900 text-white' : 'bg-slate-700 text-white'"
                      >
                        {{ order.marketplace || 'tiktok' }}
                      </span>
                      <span class="text-slate-400 text-xs">|</span>
                      <span class="text-xs text-slate-500 truncate max-w-[110px]" :title="order.buyer_username">
                        {{ order.buyer_username || 'Customer' }}
                      </span>
                    </div>
                    <p class="text-[11px] text-slate-400 mt-1 flex items-center gap-1">
                      <Calendar class="h-3 w-3" />
                      {{ formatDateOnly(order.created_at) }}
                    </p>
                  </td>

                  <!-- 2. Items list -->
                  <td class="py-3 px-4 align-top max-w-[260px]">
                    <div v-for="item in order.items" :key="item.id" class="mb-1.5 last:mb-0">
                      <p class="text-xs font-bold text-slate-800 line-clamp-2 leading-tight" :title="item.product_name">
                        {{ item.product_name || 'Produk' }}
                      </p>
                      <div class="flex items-center gap-2 mt-0.5">
                        <span class="text-[11px] font-mono text-slate-500 bg-slate-100 px-1.5 py-0.2 rounded border border-slate-200/60">
                          SKU: {{ item.seller_sku || '-' }}
                        </span>
                        <span class="text-xs font-bold text-slate-700">x{{ item.item_count }}</span>
                      </div>
                    </div>
                  </td>

                  <!-- 3. Status Fisik Gudang -->
                  <td class="py-3 px-4 align-top">
                    <div class="space-y-1">
                      <span
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs border"
                        :class="getPhysicalStatusBadge(order).class"
                      >
                        <Box v-if="getPhysicalStatusBadge(order).isPacked" class="h-3.5 w-3.5 text-slate-700" />
                        <Clock v-else class="h-3.5 w-3.5 text-slate-400" />
                        {{ getPhysicalStatusBadge(order).label }}
                      </span>
                      <p v-if="getPhysicalStatusBadge(order).actor" class="text-[11px] text-slate-500 font-medium">
                        Oleh: <strong class="text-slate-700">{{ getPhysicalStatusBadge(order).actor }}</strong>
                      </p>
                      <p v-if="order.tracking_id" class="text-[11px] font-mono text-slate-600 truncate max-w-[130px]" :title="order.tracking_id">
                        Resi: {{ order.tracking_id }}
                      </p>
                      <p v-if="order.shipping_provider && order.shipping_provider !== '-'" class="text-[10px] text-slate-400 font-medium">
                        {{ order.shipping_provider }}
                      </p>
                    </div>
                  </td>

                  <!-- 4. Jenis Batal & Alasan -->
                  <td class="py-3 px-4 align-top max-w-[240px]">
                    <div class="space-y-1">
                      <span
                        v-if="order.order_status === 'FAILED_DELIVERY' || order.settlement_status === 'FAILED_DELIVERY'"
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs border bg-slate-100 text-slate-800 border-slate-300 font-bold"
                      >
                        <Truck class="h-3.5 w-3.5 text-slate-600" />
                        Gagal Kirim (RTS)
                      </span>
                      <span
                        v-else
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs border bg-slate-100 text-slate-800 border-slate-300 font-bold"
                      >
                        <XCircle class="h-3.5 w-3.5 text-slate-600" />
                        Pesanan Dibatalkan
                      </span>

                      <!-- Alasan Rinci -->
                      <div v-if="order.fail_delivery_reason" class="text-[11px] text-slate-800 bg-slate-50 p-1.5 rounded-lg border border-slate-200" :title="order.fail_delivery_reason">
                        <span class="font-bold flex items-center gap-1 text-slate-900">
                          <Truck class="h-3 w-3 text-slate-600 shrink-0" />
                          RTS:
                        </span>
                        <span class="line-clamp-2 leading-tight mt-0.5">{{ order.fail_delivery_reason }}</span>
                      </div>
                      <div v-else-if="order.cancel_reason" class="text-[11px] text-slate-800 bg-slate-50 p-1.5 rounded-lg border border-slate-200" :title="order.cancel_reason">
                        <span class="font-bold flex items-center gap-1 text-slate-900">
                          <XCircle class="h-3 w-3 text-slate-600 shrink-0" />
                          Batal:
                        </span>
                        <span class="line-clamp-2 leading-tight mt-0.5">{{ order.cancel_reason }}</span>
                        <span v-if="order.cancel_user" class="text-[10px] text-slate-500 block mt-0.5">Oleh: {{ order.cancel_user }}</span>
                      </div>
                      <div v-else-if="order.refund_reason" class="text-[11px] text-slate-600 italic line-clamp-1" :title="order.refund_reason">
                        "{{ order.refund_reason }}"
                      </div>
                    </div>
                  </td>

                  <!-- 5. Rekonsiliasi Fisik & Restock -->
                  <td class="py-3 px-4 align-top">
                    <div class="space-y-1">
                      <!-- Status Fisik Barang -->
                      <div v-if="order.restocked_at">
                        <span class="inline-block px-2.5 py-1 rounded-lg text-[11px] bg-slate-100 text-slate-800 border border-slate-300 font-bold">
                          ✓ Sudah Direstock ke Gudang
                        </span>
                        <p class="text-[10px] text-slate-500 font-medium mt-0.5">
                          Pada: {{ formatDateOnly(order.restocked_at) }}
                        </p>
                      </div>
                      <div v-else-if="order.is_packed || order.packed_by">
                        <span class="inline-block px-2.5 py-1 rounded-lg text-[11px] bg-slate-900 text-white border border-slate-950 font-bold shadow-xs">
                          Butuh Restock (Barang Keluar)
                        </span>
                        <p class="text-[10px] text-slate-500 font-medium mt-0.5">
                          Barang fisik sudah keluar gudang
                        </p>
                      </div>
                      <div v-else>
                        <span class="inline-block px-2.5 py-1 rounded-lg text-[11px] bg-slate-100 text-slate-600 border border-slate-200 font-medium">
                          Batal Sebelum Dipack (Aman)
                        </span>
                      </div>

                      <!-- Tombol Restock -->
                      <button
                        v-if="!order.restocked_at && order.return_status !== 'RESTOCKED' && (order.is_packed || order.packed_by)"
                        @click.stop="openRestockModal(order)"
                        class="mt-1.5 inline-flex items-center gap-1 px-2.5 py-1 bg-slate-900 hover:bg-slate-800 text-white rounded-lg text-[11px] font-bold shadow-xs transition-all"
                      >
                        <RotateCcw class="h-3 w-3" />
                        Restock ke Gudang
                      </button>
                    </div>
                  </td>

                  <!-- 6. Nilai Pesanan (Refund Buyer - Minus) -->
                  <td class="py-3 px-4 align-top text-right font-bold">
                    <span class="font-black text-sm text-slate-900">
                      -{{ formatCurrency(order.total_amount) }}
                    </span>
                    <p class="text-[10px] text-slate-500 font-semibold mt-0.5">
                      Refund: -{{ formatCurrency(order.refund_amount || order.total_amount) }}
                    </p>
                    <span class="inline-block px-1.5 py-0.5 bg-slate-100 text-slate-700 border border-slate-200 rounded text-[9px] font-bold mt-1">
                      Dana Dikembalikan
                    </span>
                  </td>

                  <!-- 7. Audit Beban Kerugian -->
                  <td class="py-3 px-4 align-top text-right">
                    <div class="space-y-0.5 text-right font-medium text-xs">
                      <p v-if="order.failed_delivery_shipping_fee && order.failed_delivery_shipping_fee > 0" class="text-slate-800 font-bold text-[11px]" title="Ongkir Gagal Kirim yang dibebankan ke seller">
                        Ongkir RTS: -{{ formatCurrency(order.failed_delivery_shipping_fee) }}
                      </p>
                      <p v-if="order.packaging_cost && order.packaging_cost > 0" class="text-slate-600 text-[11px]" title="Estimasi bahan kemasan (kardus/bubble wrap)">
                        Biaya Kardus: -{{ formatCurrency(order.packaging_cost) }}
                      </p>
                      <p v-if="order.cogs_amount && order.cogs_amount > 0" class="text-slate-500 font-mono text-[10px]" title="Modal Pokok Produk (HPP)">
                        Risiko HPP: {{ formatCurrency(order.cogs_amount) }}
                      </p>
                      <p v-if="((order.failed_delivery_shipping_fee || 0) + (order.packaging_cost || 0)) > 0" class="text-slate-900 font-black text-xs border-t border-slate-100 pt-0.5 mt-0.5">
                        Beban Riil: -{{ formatCurrency((order.failed_delivery_shipping_fee || 0) + (order.packaging_cost || 0)) }}
                      </p>
                      <p v-else class="text-slate-400 text-[11px]">
                        Tidak Ada Potongan Ekstra
                      </p>
                      <span
                        class="inline-block px-1.5 py-0.5 rounded text-[9px] font-bold mt-1"
                        :class="order.settlement_status === 'SETTLED' ? 'bg-slate-100 text-slate-800 border border-slate-300' : 'bg-slate-100 text-slate-500'"
                      >
                        {{ order.settlement_status === 'SETTLED' ? 'Sempat Cair (Masuk Saldo)' : 'Belum Cair / Refund' }}
                      </span>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <!-- ─────────────────────────────────────────────── -->
        <!-- TABEL STANDAR SETTLEMENT (TAB ALL, SETTLED, HOLD) -->
        <!-- ─────────────────────────────────────────────── -->
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead class="bg-slate-50 border-b border-slate-200">
              <tr>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Order ID &amp; Platform</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Rincian Produk &amp; SKU</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Status Fisik Gudang</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Status Finansial</th>
                <th class="text-left py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Rekonsiliasi &amp; Retur</th>
                <th class="text-right py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Nilai Order Asli</th>
                <th class="text-right py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider">Potongan Fee</th>
                <th class="text-right py-3 px-4 text-xs font-black text-slate-600 uppercase tracking-wider" title="Jumlah Penyelesaian Pembayaran (Dana Bersih Masuk ke Saldo Seller)">Jumlah Penyelesaian (Dana Cair)</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <template v-for="order in marketplaceStore.settlementOrders" :key="order.order_id">
                <tr
                  class="hover:bg-slate-50/80 transition-colors cursor-pointer"
                  :class="getReconciliationBadge(order).isAlert ? 'bg-slate-50/50 border-l-4 border-l-slate-400' : ''"
                  @click="openOrderDetail(order)"
                  title="Klik untuk melihat rincian pesanan"
                >
                  <!-- Order ID & Marketplace -->
                  <td class="py-3 px-4 align-top">
                    <div class="font-bold text-slate-900 flex items-center justify-between">
                      <span class="hover:text-slate-600 transition-colors">{{ order.order_id }}</span>
                      <button
                        @click.stop="handleSyncOrderFinance(order.order_id, order.marketplace)"
                        :disabled="isSyncingOrderFinance[order.order_id]"
                        class="p-1 rounded-md text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition-colors"
                        :title="`Tarik rincian finansial ${order.marketplace || 'TikTok'}`"
                      >
                        <RefreshCw class="h-3 w-3" :class="{ 'animate-spin': isSyncingOrderFinance[order.order_id] }" />
                      </button>
                    </div>
                    <div class="flex items-center gap-1 mt-1">
                      <span
                        class="px-2 py-0.5 rounded-md text-[10px] font-black uppercase tracking-wider"
                        :class="order.marketplace === 'tiktok' ? 'bg-slate-900 text-white' : 'bg-slate-700 text-white'"
                      >
                        {{ order.marketplace || 'tiktok' }}
                      </span>
                      <span class="text-slate-400 text-xs">|</span>
                      <span class="text-xs text-slate-500 truncate max-w-[110px]" :title="order.buyer_username">
                        {{ order.buyer_username || 'Customer' }}
                      </span>
                    </div>
                    <p class="text-[11px] text-slate-400 mt-1 flex items-center gap-1">
                      <Calendar class="h-3 w-3" />
                      {{ formatDateOnly(order.created_at) }}
                    </p>
                  </td>

                  <!-- Items list -->
                  <td class="py-3 px-4 align-top max-w-[260px]">
                    <div v-for="item in order.items" :key="item.id" class="mb-1.5 last:mb-0">
                      <p class="text-xs font-bold text-slate-800 line-clamp-2 leading-tight" :title="item.product_name">
                        {{ item.product_name || 'Produk' }}
                      </p>
                      <div class="flex items-center gap-2 mt-0.5">
                        <span class="text-[11px] font-mono text-slate-500 bg-slate-100 px-1.5 py-0.2 rounded border border-slate-200/60">
                          SKU: {{ item.seller_sku || '-' }}
                        </span>
                        <span class="text-xs font-bold text-slate-700">x{{ item.item_count }}</span>
                      </div>
                    </div>
                  </td>

                  <!-- Status Fisik Gudang (Super App) -->
                  <td class="py-3 px-4 align-top">
                    <div class="space-y-1">
                      <span
                        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs border"
                        :class="getPhysicalStatusBadge(order).class"
                      >
                        <Box v-if="getPhysicalStatusBadge(order).isPacked" class="h-3.5 w-3.5 text-slate-700" />
                        <Clock v-else class="h-3.5 w-3.5 text-slate-400" />
                        {{ getPhysicalStatusBadge(order).label }}
                      </span>
                      <p v-if="getPhysicalStatusBadge(order).actor" class="text-[11px] text-slate-500 font-medium">
                        Oleh: <strong class="text-slate-700">{{ getPhysicalStatusBadge(order).actor }}</strong>
                      </p>
                      <p v-if="order.tracking_id" class="text-[11px] font-mono text-slate-400 truncate max-w-[130px]" :title="order.tracking_id">
                        Resi: {{ order.tracking_id }}
                      </p>
                    </div>
                  </td>

                  <!-- Status Finansial Marketplace -->
                  <td class="py-3 px-4 align-top">
                    <span
                      class="inline-block px-2.5 py-1 rounded-lg text-xs border"
                      :class="isUnverifiedSettlement(order) ? 'bg-amber-100 text-amber-800 border border-amber-300 font-semibold' : getSettlementBadge(order.settlement_status).class"
                    >
                      {{ isUnverifiedSettlement(order) ? 'Pencairan Belum Terverifikasi' : getSettlementBadge(order.settlement_status).label }}
                    </span>

                    <!-- Tanggal Cair / On-Hold Reason -->
                    <div v-if="order.settlement_status === 'SETTLED' && order.settlement_date && !isUnverifiedSettlement(order)" class="text-[11px] text-slate-700 font-bold mt-1">
                      Cair: {{ formatDateOnly(order.settlement_date) }}
                    </div>
                    <div v-else-if="order.settlement_status === 'HOLD' || order.settlement_status === 'PENDING_SETTLEMENT'" class="mt-1 space-y-0.5">
                      <p v-if="order.hold_reason" class="text-[11px] text-slate-700 font-medium line-clamp-1" :title="order.hold_reason">
                        ⏳ {{ order.hold_reason }}
                      </p>
                      <p v-if="order.estimated_settlement_notes" class="text-[10px] text-slate-600 font-bold">
                        {{ order.estimated_settlement_notes }}
                      </p>
                      <p v-else-if="order.estimated_settlement_date" class="text-[10px] text-slate-500 font-medium">
                        Est: {{ formatDateOnly(order.estimated_settlement_date) }}
                      </p>
                    </div>
                    <div v-else class="text-[11px] text-slate-400 mt-1">
                      Belum cair
                    </div>
                  </td>

                  <!-- Rekonsiliasi & Status Retur -->
                  <td class="py-3 px-4 align-top">
                    <div class="space-y-1">
                      <span
                        class="inline-block px-2.5 py-1 rounded-lg text-[11px] border"
                        :class="getReconciliationBadge(order).class"
                        :title="getReconciliationBadge(order).desc"
                      >
                        {{ getReconciliationBadge(order).label }}
                      </span>

                      <!-- Detail info jika restocked / refund / perlu restock -->
                      <div v-if="order.restocked_at" class="text-[11px] text-slate-700 font-semibold">
                        ✓ Direstock: {{ formatDateOnly(order.restocked_at) }}
                      </div>
                      <div v-else-if="order.fail_delivery_reason" class="text-[11px] text-slate-800 bg-slate-50 p-1.5 rounded-lg border border-slate-200" :title="order.fail_delivery_reason">
                        <span class="font-bold flex items-center gap-1 text-slate-900">
                          <Truck class="h-3 w-3 text-slate-600 shrink-0" />
                          RTS:
                        </span>
                        <span class="line-clamp-2 leading-tight mt-0.5">{{ order.fail_delivery_reason }}</span>
                      </div>
                      <div v-else-if="order.cancel_reason" class="text-[11px] text-slate-700 bg-slate-50 p-1.5 rounded-lg border border-slate-200" :title="order.cancel_reason">
                        <span class="font-bold flex items-center gap-1 text-slate-900">
                          <XCircle class="h-3 w-3 text-slate-600 shrink-0" />
                          Batal:
                        </span>
                        <span class="line-clamp-2 leading-tight mt-0.5">{{ order.cancel_reason }}</span>
                        <span v-if="order.cancel_user" class="text-[10px] text-slate-500 block mt-0.5">Oleh: {{ order.cancel_user }}</span>
                      </div>
                      <div v-else-if="order.refund_reason" class="text-[11px] text-slate-600 italic line-clamp-1" :title="order.refund_reason">
                        "{{ order.refund_reason }}"
                      </div>
                      <button
                        v-if="(getReconciliationBadge(order).type === 'DISPATCHED_RETURN' || getReconciliationBadge(order).type === 'FAILED_DELIVERY' || order.order_status === 'CANCELED' || order.order_status === 'CANCELLED' || order.order_status === 'FAILED_DELIVERY' || order.order_status === 'RETURN_SUCCESS') && !order.restocked_at && order.return_status !== 'RESTOCKED'"
                        @click.stop="openRestockModal(order)"
                        class="mt-1 inline-flex items-center gap-1 px-2 py-1 bg-slate-900 hover:bg-slate-800 text-white rounded-lg text-[11px] font-bold shadow-xs transition-all"
                      >
                        <RotateCcw class="h-3 w-3" />
                        Restock ke Gudang
                      </button>
                    </div>
                  </td>

                  <!-- Original Order Value (Unmodified) -->
                  <td class="py-3 px-4 align-top text-right font-bold text-slate-900">
                    {{ formatCurrency(order.total_amount) }}
                  </td>

                  <!-- Platform Fee & Taxes -->
                  <td class="py-3 px-4 align-top text-right text-xs font-bold" :class="((order.fee_and_tax_amount || order.platform_fee) || 0) > 0 ? 'text-slate-900' : 'text-slate-400'">
                    {{ ((order.fee_and_tax_amount || order.platform_fee) || 0) > 0 ? '-' + formatCurrency(order.fee_and_tax_amount || order.platform_fee) : 'Rp 0' }}
                    <span
                      v-if="order.settlement_status === 'SETTLED' && !order.finance_synced_at && !((order.fee_and_tax_amount || order.platform_fee) || 0)"
                      class="block text-[9px] text-amber-600 font-medium mt-0.5"
                        :title="`Fee ${order.marketplace || 'TikTok'} belum tersinkron untuk pesanan ini`"
                    >
                      (Belum Sinkron Fee)
                    </span>
                    <span
                      v-else-if="order.marketplace === 'tiktok' && order.settlement_status === 'SETTLED' && !((order.fee_and_tax_amount || order.platform_fee) || 0)"
                      class="block text-[9px] text-amber-600 font-medium mt-0.5"
                      title="Fee Rp0 pada data tersimpan; periksa rincian finansial order di TikTok"
                    >
                      (Fee Rp0, perlu verifikasi)
                    </span>
                    <p v-if="order.shipping_cost_amount && order.shipping_cost_amount > 0" class="text-[10px] text-slate-600 font-semibold mt-0.5">
                      Ongkir: -{{ formatCurrency(order.shipping_cost_amount) }}
                    </p>
                    <p v-if="order.failed_delivery_shipping_fee && order.failed_delivery_shipping_fee > 0" class="text-[10px] text-slate-700 font-bold mt-0.5" title="Ongkir Gagal Kirim yang dibebankan ke seller">
                      Ongkir RTS: -{{ formatCurrency(order.failed_delivery_shipping_fee) }}
                    </p>
                    <p v-if="order.cogs_amount && order.cogs_amount > 0" class="text-[10px] text-slate-500 font-mono mt-0.5" title="Modal Pokok Produk (HPP)">
                      HPP: {{ formatCurrency(order.cogs_amount) }}
                    </p>
                    <p v-if="order.adjustment_amount && order.adjustment_amount !== 0" class="text-[10px] text-slate-600 font-semibold mt-0.5">
                      Adj: {{ formatCurrency(order.adjustment_amount) }}
                    </p>
                  </td>

                  <!-- Net Settlement Amount / Est. Amount -->
                  <td class="py-3 px-4 align-top text-right">
                    <span
                      class="font-black text-sm"
                      :class="order.settlement_status === 'SETTLED' && !isUnverifiedSettlement(order) ? 'text-emerald-700 font-extrabold' : 'text-slate-900'"
                    >
                      {{ formatCurrency(settlementDisplayAmount(order)) }}
                    </span>
                    <p v-if="order.settlement_status === 'HOLD' || order.settlement_status === 'PENDING_SETTLEMENT'" class="text-[10px] text-amber-700 font-semibold mt-0.5">
                      (Estimasi Pencairan)
                    </p>
                    <p v-else-if="isUnverifiedSettlement(order)" class="text-[10px] text-amber-700 font-semibold mt-0.5">
                      Angka sementara; belum ada bukti fee/pencairan resmi
                    </p>
                    <p v-else-if="order.marketplace === 'tiktok' && order.settlement_status === 'SETTLED' && !(Number(order.settlement_amount) > 0)" class="text-[10px] text-amber-700 font-semibold mt-0.5">
                      Nominal cair Rp0; periksa statement TikTok
                    </p>
                    <p v-else-if="order.settlement_status === 'SETTLED'" class="text-[10px] text-emerald-600 font-bold mt-0.5">
                      ✓ Dana Bersih Cair
                    </p>
                    <p v-if="(order.refund_amount || 0) > 0" class="text-[10px] text-rose-600 font-semibold mt-0.5">
                      Refund: -{{ formatCurrency(order.refund_amount) }}
                    </p>
                    <p v-if="(order.failed_delivery_shipping_fee || 0) + (order.packaging_cost || 0) > 0" class="text-[10px] text-slate-800 font-bold mt-0.5">
                      Beban: -{{ formatCurrency((order.failed_delivery_shipping_fee || 0) + (order.packaging_cost || 0)) }}
                    </p>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <!-- Pagination Controls -->
        <div v-if="marketplaceStore.settlementOrders.length > 0" class="flex items-center justify-between px-4 py-3 border-t border-slate-100 bg-slate-50/50">
          <span class="text-xs text-slate-500 font-medium">
            Halaman {{ currentPage }} dari {{ totalPages }} (Total {{ marketplaceStore.settlementOrdersTotal }} transaksi)
          </span>

          <div class="flex items-center gap-2">
            <button
              @click="goToPage(currentPage - 1)"
              :disabled="currentPage <= 1"
              class="p-1.5 rounded-lg border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              <ChevronLeft class="h-4 w-4" />
            </button>
            <span class="text-xs font-bold text-slate-700 px-2">{{ currentPage }}</span>
            <button
              @click="goToPage(currentPage + 1)"
              :disabled="currentPage >= totalPages"
              class="p-1.5 rounded-lg border border-slate-200 bg-white text-slate-600 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              <ChevronRight class="h-4 w-4" />
            </button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- ─────────────────────────────────────────────── -->
    </template>

    <!-- MODAL 1: RINCIAN RESMI PESANAN (READ-ONLY AUDIT) -->
    <!-- ─────────────────────────────────────────────── -->
    <div
      v-if="showDetailModal && selectedOrder"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200"
    >
      <div class="bg-white rounded-3xl shadow-2xl max-w-3xl w-full max-h-[90vh] overflow-y-auto border border-slate-100 flex flex-col">
        <!-- Header -->
        <div class="p-6 border-b border-slate-100 flex items-center justify-between sticky top-0 bg-white z-10">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-2xl bg-slate-100 text-slate-700">
              <Wallet class="h-6 w-6" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-lg font-black text-slate-900">Rincian Rekonsiliasi &amp; Finansial</h3>
                <span class="inline-flex items-center gap-1 text-[10px] font-black px-2 py-0.5 rounded-full bg-slate-100 text-slate-700 border border-slate-200">
                  <Lock class="h-3 w-3 text-slate-500" /> Read-Only
                </span>
              </div>
              <p class="text-xs text-slate-500 font-mono mt-0.5">
                No. Pesanan: <span class="font-bold text-slate-800">{{ selectedOrder.order_id }}</span>
                <span v-if="selectedOrder.marketplace" class="ml-2 uppercase text-[10px] px-2 py-0.5 rounded-md bg-slate-100 text-slate-600 font-bold font-sans border border-slate-200/60">
                  {{ selectedOrder.marketplace }}
                </span>
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button
              @click="handleSyncOrderFinance(selectedOrder.order_id, selectedOrder.marketplace)"
              :disabled="isSyncingOrderFinance[selectedOrder.order_id]"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-100 border border-slate-300 text-slate-800 font-bold text-xs hover:bg-slate-200 transition-colors disabled:opacity-50"
              :title="`Tarik rincian finansial pesanan ini dari ${selectedOrder.marketplace || 'TikTok'}`"
            >
              <RefreshCw class="h-3.5 w-3.5" :class="{ 'animate-spin': isSyncingOrderFinance[selectedOrder.order_id] }" />
              {{ isSyncingOrderFinance[selectedOrder.order_id] ? 'Menarik...' : 'Tarik Finansial (Drill-Down)' }}
            </button>
            <button
              @click="showDetailModal = false"
              class="p-2 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
            >
              <XCircle class="h-5 w-5" />
            </button>
          </div>
        </div>

        <!-- Body -->
        <div class="p-6 space-y-6 flex-1">
          <!-- Data Integrity Banner -->
          <div class="p-3.5 bg-slate-50 border border-slate-200 rounded-2xl flex items-start gap-3">
            <ShieldAlert class="h-5 w-5 text-slate-700 shrink-0 mt-0.5" />
            <div class="text-xs text-slate-900">
              <p class="font-bold mb-0.5">🔒 Data Terproteksi &amp; Terpisah dengan Aman</p>
              <p class="text-slate-600 leading-relaxed">
                Nilai pesanan fisik dan status packing di Super App terjaga tanpa ditimpa. Rincian finansial diambil dari API marketplace yang sesuai.
              </p>
            </div>
          </div>

          <!-- Financial Breakdown Cards -->
          <div>
            <h4 class="text-xs font-black uppercase text-slate-400 tracking-wider mb-3">Ringkasan Finansial Marketplace</h4>
            <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
              <div class="bg-slate-50 p-3.5 rounded-2xl border border-slate-200">
                <p class="text-[11px] text-slate-500 font-bold uppercase">Nilai Order Asli (GMV)</p>
                <p class="text-base font-black text-slate-900 mt-1">{{ formatCurrency(selectedOrder.total_amount) }}</p>
                <p class="text-[10px] text-slate-400 mt-0.5">Total belanja pembeli</p>
              </div>
              <div class="bg-slate-50 p-3.5 rounded-2xl border border-slate-200">
                <p class="text-[11px] text-slate-500 font-bold uppercase">Fee &amp; Pajak</p>
                <p class="text-base font-black text-slate-900 mt-1">- {{ formatCurrency(selectedOrder.fee_and_tax_amount || selectedOrder.platform_fee || 0) }}</p>
                <p class="text-[10px] text-slate-400 mt-0.5">
                  {{ (selectedOrder.fee_and_tax_amount || selectedOrder.platform_fee || 0) > 0 ? 'Admin & layanan platform' : selectedOrder.marketplace === 'tiktok' && selectedOrder.settlement_status === 'SETTLED' ? 'Fee Rp0 pada data; cek rincian TikTok' : 'Belum ada potongan' }}
                </p>
              </div>
              <div class="bg-slate-50 p-3.5 rounded-2xl border border-slate-200">
                <p class="text-[11px] text-slate-500 font-bold uppercase">Ongkir Platform</p>
                <p class="text-base font-black text-slate-900 mt-1">- {{ formatCurrency(selectedOrder.shipping_cost_amount || selectedOrder.shipping_fee_deduction || 0) }}</p>
                <p class="text-[10px] text-slate-400 mt-0.5">
                  {{ (selectedOrder.shipping_cost_amount || 0) > 0 ? 'Potongan logistik TikTok' : 'Ditanggung platform/pembeli' }}
                </p>
              </div>
              <div class="bg-slate-50 p-3.5 rounded-2xl border border-slate-200">
                <p class="text-[11px] text-slate-500 font-bold uppercase">Penyesuaian (Adj)</p>
                <p class="text-base font-black text-slate-900 mt-1">
                  {{ formatCurrency(selectedOrder.adjustment_amount || 0) }}
                </p>
                <p class="text-[10px] text-slate-400 mt-0.5">
                  {{ (selectedOrder.adjustment_amount || 0) !== 0 ? 'Penyesuaian kebijakan TikTok' : 'Tidak ada adj' }}
                </p>
              </div>
              <div class="bg-slate-900 text-white p-3.5 rounded-2xl shadow-xs col-span-2 sm:col-span-1">
                <p class="text-[11px] text-slate-300 font-bold uppercase">
                  {{ isUnverifiedSettlement(selectedOrder) ? 'Nominal Belum Terverifikasi' : selectedOrder.settlement_status === 'HOLD' ? 'Estimasi Dana Cair' : 'Dana Cair (Settlement)' }}
                </p>
                <p class="text-lg font-black text-white mt-1">
                  {{ formatCurrency(selectedOrder.settlement_amount || 0) }}
                </p>
                <p class="text-[10px] text-slate-300 mt-0.5">
                  {{ isUnverifiedSettlement(selectedOrder) ? 'Belum ada bukti pencairan/fee resmi' : selectedOrder.settlement_status === 'HOLD' ? 'Tertahan di escrow' : 'Masuk ke saldo seller' }}
                </p>
              </div>
            </div>

            <!-- Breakdown Detail Box if available -->
            <div
              v-if="selectedOrder.fee_tax_breakdown"
              class="mt-3 bg-slate-50 border border-slate-200 p-3.5 rounded-2xl text-xs space-y-1.5"
            >
              <div class="flex items-center justify-between font-bold text-slate-700">
                <span>Rincian Komponen Biaya &amp; Pajak (Breakdown v202501):</span>
                <span class="text-[10px] font-mono text-slate-700 bg-slate-100 px-2 py-0.5 rounded border border-slate-200">Official API</span>
              </div>
              <div class="font-mono text-[11px] text-slate-600 bg-white p-2.5 rounded-xl border border-slate-100 overflow-x-auto max-h-36 overflow-y-auto whitespace-pre-wrap">
                {{ selectedOrder.fee_tax_breakdown }}
              </div>
            </div>

            <!-- On-Hold / Refund Reason Box -->
            <div
              v-if="selectedOrder.settlement_status === 'HOLD' || selectedOrder.hold_reason || selectedOrder.estimated_settlement_notes"
              class="mt-3 bg-slate-50 border border-slate-200 p-3.5 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-2"
            >
              <div>
                <span class="text-xs font-bold text-slate-900">Alasan Dana Tertahan (On-Hold): </span>
                <span class="text-xs font-bold text-slate-700 ml-1">
                  {{ selectedOrder.hold_reason || 'Menunggu paket diterima / estimasi waktu penyelesaian' }}
                </span>
                <p v-if="selectedOrder.estimated_settlement_notes" class="text-[11px] text-slate-500 mt-0.5">
                  Estimasi waktu penyelesaian: <strong>{{ selectedOrder.estimated_settlement_notes }}</strong>
                </p>
              </div>
              <span class="text-[11px] font-bold px-2.5 py-1 rounded-full bg-slate-100 text-slate-800 border border-slate-300 shrink-0 self-start sm:self-center">
                Escrow On-Hold
              </span>
            </div>

            <div
              v-if="(selectedOrder.refund_amount && selectedOrder.refund_amount > 0) || ['CANCELED', 'CANCELLED', 'FAILED_DELIVERY'].includes(selectedOrder.order_status)"
              class="mt-3 bg-slate-50 border border-slate-200 p-3.5 rounded-2xl flex items-center justify-between"
            >
              <div>
                <span class="text-xs font-bold text-slate-900">Dana Dikembalikan ke Pembeli (Refund): </span>
                <span class="text-sm font-black text-slate-900 ml-1">
                  -{{ formatCurrency(selectedOrder.refund_amount || selectedOrder.total_amount) }}
                </span>
                <p v-if="selectedOrder.refund_reason" class="text-[11px] text-slate-500 mt-0.5">
                  Alasan: {{ selectedOrder.refund_reason }}
                </p>
              </div>
              <span class="text-[11px] font-bold px-2.5 py-1 rounded-full bg-slate-100 text-slate-800 border border-slate-300">
                Refunded
              </span>
            </div>

            <!-- Cancellation & RTS Financial Loss Breakdown Box -->
            <div
              v-if="['CANCELED', 'CANCELLED', 'FAILED_DELIVERY'].includes(selectedOrder.order_status) || ['CANCELLED', 'FAILED_DELIVERY', 'REFUNDED'].includes(selectedOrder.settlement_status) || selectedOrder.cancel_reason || selectedOrder.fail_delivery_reason || (selectedOrder.failed_delivery_shipping_fee && selectedOrder.failed_delivery_shipping_fee > 0)"
              class="mt-4 bg-slate-50 border border-slate-200 p-4 rounded-2xl space-y-3"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <div class="p-1.5 bg-slate-800 text-white rounded-lg shadow-xs">
                    <ShieldAlert class="h-4 w-4" />
                  </div>
                  <div>
                    <h5 class="text-xs font-black uppercase text-slate-900">Analisis Kerugian Pembatalan &amp; Gagal Antar (RTS)</h5>
                    <p class="text-[11px] text-slate-500">Rincian beban biaya riil dan risiko modal untuk pesanan ini</p>
                  </div>
                </div>
                <span
                  class="text-[11px] font-bold px-2.5 py-1 rounded-full border bg-slate-100 text-slate-800 border-slate-300"
                >
                  {{ selectedOrder.order_status === 'FAILED_DELIVERY' || selectedOrder.settlement_status === 'FAILED_DELIVERY' ? 'Pengantaran Gagal (RTS)' : 'Pesanan Dibatalkan' }}
                </span>
              </div>

              <!-- Reason boxes -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                <div v-if="selectedOrder.cancel_reason" class="p-2.5 bg-white rounded-xl border border-slate-200">
                  <p class="text-[11px] font-bold text-slate-500">Alasan Pembatalan:</p>
                  <p class="font-bold text-slate-900 mt-0.5">{{ selectedOrder.cancel_reason }}</p>
                  <p v-if="selectedOrder.cancel_user" class="text-[10px] text-slate-400 mt-0.5">Dibatalkan oleh: <span class="font-semibold text-slate-600">{{ selectedOrder.cancel_user }}</span></p>
                </div>
                <div v-if="selectedOrder.fail_delivery_reason" class="p-2.5 bg-white rounded-xl border border-slate-200">
                  <p class="text-[11px] font-bold text-slate-500">Alasan Gagal Kirim (Kurir RTS):</p>
                  <p class="font-bold text-slate-900 mt-0.5">{{ selectedOrder.fail_delivery_reason }}</p>
                </div>
              </div>

              <!-- 3 Metrics in Modal -->
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5 pt-1">
                <div class="p-2.5 bg-white rounded-xl border border-slate-200">
                  <p class="text-[10px] font-bold text-slate-500 uppercase">Ongkir Gagal Kirim (Beban Platform)</p>
                  <p class="text-sm font-black text-slate-900 mt-0.5">
                    {{ formatCurrency(selectedOrder.failed_delivery_shipping_fee || 0) }}
                  </p>
                  <p class="text-[10px] text-slate-400 mt-0.5">Potongan saldo TikTok Seller</p>
                </div>
                <div class="p-2.5 bg-white rounded-xl border border-slate-200">
                  <p class="text-[10px] font-bold text-slate-500 uppercase">Modal Barang Fisik (HPP)</p>
                  <p class="text-sm font-black text-slate-900 mt-0.5">
                    {{ formatCurrency(selectedOrder.cogs_amount || 0) }}
                  </p>
                  <p class="text-[10px] text-slate-400 mt-0.5">
                    {{ selectedOrder.restocked_at ? 'Sudah direstock ke gudang' : 'Belum restock (Risiko HPP)' }}
                  </p>
                </div>
                <div class="p-2.5 bg-white rounded-xl border border-slate-200">
                  <p class="text-[10px] font-bold text-slate-500 uppercase">Potensi Omzet Hilang (GMV)</p>
                  <p class="text-sm font-black text-slate-900 mt-0.5">
                    -{{ formatCurrency(selectedOrder.total_amount || 0) }}
                  </p>
                  <p class="text-[10px] text-slate-400 mt-0.5">Nilai transaksi pesanan batal</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Transaction & Settlement Timeline Info -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="bg-slate-50 p-4 rounded-2xl border border-slate-200/80 space-y-2.5">
              <h4 class="text-xs font-black uppercase text-slate-400 tracking-wider flex items-center gap-1.5">
                <Calendar class="h-3.5 w-3.5 text-slate-500" /> Informasi Waktu &amp; Tanggal
              </h4>
              <div class="flex items-center justify-between text-xs py-1 border-b border-slate-200/60">
                <span class="text-slate-500">Tanggal Checkout:</span>
                <span class="font-bold text-slate-800">{{ formatDate(selectedOrder.created_at) }}</span>
              </div>
              <div v-if="selectedOrder.shipped_at" class="flex items-center justify-between text-xs py-1 border-b border-slate-200/60">
                <span class="text-slate-500">Waktu Pengiriman:</span>
                <span class="font-bold text-slate-800">{{ formatDate(selectedOrder.shipped_at) }}</span>
              </div>
              <div v-if="selectedOrder.delivered_at" class="flex items-center justify-between text-xs py-1 border-b border-slate-200/60">
                <span class="text-slate-500">Tanggal Pengantaran:</span>
                <span class="font-bold text-slate-800">{{ formatDate(selectedOrder.delivered_at) }}</span>
              </div>
              <div class="flex items-center justify-between text-xs py-1">
                <span class="text-slate-500">Tanggal Pencairan Dana:</span>
                <span class="font-black text-slate-900">
                  {{
                    isUnverifiedSettlement(selectedOrder)
                      ? 'Belum terverifikasi'
                      : selectedOrder.settlement_date
                      ? formatDate(selectedOrder.settlement_date)
                      : (selectedOrder.settlement_status === 'SETTLED'
                          ? formatDate(selectedOrder.settled_at || selectedOrder.updated_at)
                          : (selectedOrder.estimated_settlement_notes ? selectedOrder.estimated_settlement_notes : 'Belum Cair (Escrow/Hold)'))
                  }}
                </span>
              </div>
            </div>

            <div class="bg-slate-50 p-4 rounded-2xl border border-slate-200/80 space-y-2.5">
              <h4 class="text-xs font-black uppercase text-slate-400 tracking-wider flex items-center gap-1.5">
                <Truck class="h-3.5 w-3.5 text-slate-500" /> Logistik &amp; Fisik Barang
              </h4>
              <div class="flex items-center justify-between text-xs py-1 border-b border-slate-200/60">
                <span class="text-slate-500">Kurir / Ekspedisi:</span>
                <span class="font-bold text-slate-800">{{ selectedOrder.shipping_provider || '-' }}</span>
              </div>
              <div class="flex items-center justify-between text-xs py-1 border-b border-slate-200/60">
                <span class="text-slate-500">No. Resi Pengiriman:</span>
                <span class="font-mono font-bold text-slate-800">{{ selectedOrder.tracking_number || selectedOrder.tracking_id || '-' }}</span>
              </div>
              <div class="flex items-center justify-between text-xs py-1">
                <span class="text-slate-500">Fisik Packing Gudang:</span>
                <span
                  class="text-[10px] font-bold px-2 py-0.5 rounded-md border"
                  :class="getPhysicalStatusBadge(selectedOrder).class"
                >
                  {{ getPhysicalStatusBadge(selectedOrder).label }}
                </span>
              </div>
            </div>
          </div>

          <!-- Items / Produk Details -->
          <div>
            <h4 class="text-xs font-black uppercase text-slate-400 tracking-wider mb-2 flex items-center gap-1.5">
              <Package class="h-3.5 w-3.5 text-slate-500" /> Rincian Produk Pesanan
            </h4>
            <div class="border border-slate-200 rounded-2xl overflow-hidden">
              <div v-if="selectedOrder.items && selectedOrder.items.length > 0" class="divide-y divide-slate-100">
                <div
                  v-for="(item, idx) in selectedOrder.items"
                  :key="idx"
                  class="p-3.5 flex items-center justify-between text-xs hover:bg-slate-50/50"
                >
                  <div class="space-y-0.5">
                    <p class="font-bold text-slate-900">{{ item.product_name || item.item_name || 'Produk Marketplace' }}</p>
                    <p class="text-[11px] text-slate-500 font-mono">
                      SKU: <span class="font-semibold text-slate-700">{{ item.seller_sku || item.sku || '-' }}</span> | Qty: <span class="font-black text-slate-900">{{ item.item_count || item.quantity || 1 }}x</span>
                    </p>
                  </div>
                  <div class="text-right">
                    <p class="font-black text-slate-900">
                      {{ formatCurrency(item.total_amount || item.total_price || (selectedOrder.total_amount && selectedOrder.items.length === 1 ? selectedOrder.total_amount : 0)) }}
                    </p>
                    <p class="text-[10px] text-slate-400">
                      @ {{ formatCurrency((item.total_amount || item.total_price) ? ((item.total_amount || item.total_price) / (item.item_count || item.quantity || 1)) : ((selectedOrder.total_amount && selectedOrder.items.length === 1) ? (selectedOrder.total_amount / (item.item_count || 1)) : 0)) }}
                    </p>
                  </div>
                </div>
              </div>
              <div v-else class="p-4 text-center text-xs text-slate-400">
                Rincian produk tidak tersedia untuk pesanan ini.
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-4 sm:p-6 border-t border-slate-100 flex items-center justify-between bg-slate-50 rounded-b-3xl">
          <div class="flex items-center gap-2 text-xs text-slate-400">
            <Lock class="h-3.5 w-3.5 text-slate-400" />
            <span>ID Pesanan: <span class="font-mono font-medium text-slate-600">{{ selectedOrder.order_id }}</span></span>
          </div>

          <div class="flex items-center gap-2">
            <button
              v-if="(['CANCELED', 'CANCELLED', 'FAILED_DELIVERY'].includes(selectedOrder.order_status) || ['CANCELLED', 'FAILED_DELIVERY'].includes(selectedOrder.settlement_status)) && !selectedOrder.restocked_at && selectedOrder.return_status !== 'RESTOCKED'"
              @click="showDetailModal = false; openRestockModal(selectedOrder)"
              class="px-4 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs flex items-center gap-1.5 shadow-xs"
            >
              <RotateCcw class="h-3.5 w-3.5" />
              Restock Fisik Barang
            </button>
            <button
              @click="showDetailModal = false"
              class="px-5 py-2.5 rounded-xl bg-slate-200 hover:bg-slate-300 text-slate-800 font-bold text-xs transition-colors"
            >
              Tutup
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ─────────────────────────────────────────────── -->
    <!-- MODAL 2: KONFIRMASI RESTOCK BARANG KE GUDANG   -->
    <!-- ─────────────────────────────────────────────── -->
    <div
      v-if="showRestockModal && restockTargetOrder"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200"
    >
      <div class="bg-white rounded-3xl shadow-2xl max-w-md w-full border border-slate-100 overflow-hidden">
        <div class="p-6 text-center space-y-3">
          <div class="mx-auto w-14 h-14 rounded-2xl bg-slate-100 text-slate-700 flex items-center justify-center">
            <RotateCcw class="h-7 w-7" />
          </div>
          <h3 class="text-lg font-black text-slate-900">Konfirmasi Restock Fisik Barang</h3>
          <p class="text-xs text-slate-500 leading-relaxed">
            Barang dari pesanan <span class="font-bold text-slate-800 font-mono">{{ restockTargetOrder.order_id }}</span> akan dikembalikan ke stok aktif gudang marketplace dan otomatis mencatat mutasi Opname <b>Stock In</b>.
          </p>

          <!-- List Items To Restock -->
          <div class="text-left bg-slate-50 p-3 rounded-2xl border border-slate-100 space-y-2 mt-2">
            <div v-for="it in restockTargetOrder.items" :key="it.id" class="flex justify-between items-center text-xs">
              <span class="font-bold text-slate-700 truncate max-w-[220px]">{{ it.product_name }}</span>
              <span class="font-mono bg-slate-200 text-slate-800 px-2 py-0.5 rounded-md font-bold border border-slate-300">+{{ it.item_count }} Unit</span>
            </div>
          </div>

          <div class="text-left mt-3">
            <label class="block text-xs font-bold text-slate-700 mb-1">Catatan Restock (Opsional)</label>
            <input
              v-model="restockNotes"
              type="text"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl text-xs focus:outline-none focus:ring-2 focus:ring-slate-400"
              placeholder="Misal: Paket RTS diterima dalam kondisi utuh dan segel"
            />
          </div>
        </div>

        <div class="p-4 border-t border-slate-100 flex items-center justify-end gap-2.5 bg-slate-50">
          <button
            @click="showRestockModal = false"
            class="px-4 py-2.5 rounded-xl border border-slate-200 text-slate-600 font-bold text-xs hover:bg-slate-100"
          >
            Batal
          </button>
          <button
            @click="executeRestock"
            :disabled="isRestocking"
            class="px-5 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs shadow-xs disabled:opacity-50"
          >
            {{ isRestocking ? 'Memproses Restock...' : 'Ya, Terima & Restock ke Gudang' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ─────────────────────────────────────────────── -->
    <!-- MODAL 3: IMPORT LAPORAN SETTLEMENT TIKTOK/SHOPEE-->
    <!-- ─────────────────────────────────────────────── -->
    <div
      v-if="showImportModal"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200"
    >
      <div class="bg-white rounded-3xl shadow-2xl max-w-2xl w-full border border-slate-100 overflow-hidden flex flex-col max-h-[90vh]">
        <div class="p-6 border-b border-slate-100 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-2xl bg-slate-100 text-slate-700">
              <Upload class="h-6 w-6" />
            </div>
            <div>
              <h3 class="text-lg font-black text-slate-900">Upload Laporan Settlement &amp; On-Hold</h3>
              <p class="text-xs text-slate-500">Mendukung file Transaksi Penghasilan (Income) maupun Dana Tertahan (On-Hold)</p>
            </div>
          </div>
          <button
            @click="showImportModal = false"
            class="p-2 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100"
          >
            <XCircle class="h-5 w-5" />
          </button>
        </div>

        <div class="p-6 space-y-4 overflow-y-auto flex-1">
          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1">Pilih Platform</label>
            <select
              v-model="importMarketplace"
              class="w-full px-3 py-2.5 border border-slate-200 rounded-xl text-sm font-semibold focus:outline-none focus:ring-2 focus:ring-slate-400"
            >
              <option value="tiktok">TikTok Shop (Laporan Transaksi Penghasilan / On-Hold)</option>
              <option value="shopee">Shopee (Laporan Penghasilan / Saldo Penjual)</option>
            </select>
          </div>

          <!-- File Upload Box -->
          <div
            class="border-2 border-dashed border-slate-200 rounded-2xl p-6 text-center hover:border-slate-400 transition-colors bg-slate-50 cursor-pointer"
            @click="($refs.fileInput as HTMLInputElement)?.click()"
          >
            <input
              ref="fileInput"
              type="file"
              accept=".xlsx,.xls,.csv,.txt"
              @change="handleFileUpload"
              class="hidden"
            />
            <div class="mx-auto w-12 h-12 rounded-2xl bg-slate-100 text-slate-700 flex items-center justify-center mb-2">
              <FileSpreadsheet class="h-6 w-6" />
            </div>
            <p class="text-xs font-bold text-slate-800">
              {{ importFile ? importFile.name : 'Klik untuk memilih berkas Excel (.xlsx / .xls) atau CSV Laporan TikTok' }}
            </p>
            <p class="text-[11px] text-slate-400 mt-1">Otomatis mendeteksi format Tab Income (Dana Cair) maupun Tab On-Hold (Dana Tertahan)</p>
          </div>

          <!-- Selective Packed Matching Setting (Crucial Requirement) -->
          <div class="p-3.5 bg-slate-50 border border-slate-200 rounded-2xl space-y-2">
            <label class="flex items-start gap-3 cursor-pointer">
              <input
                type="checkbox"
                v-model="importOnlyPackedOrders"
                class="mt-0.5 h-4 w-4 rounded border-slate-300 text-slate-900 focus:ring-slate-500"
              />
              <div class="text-xs">
                <span class="font-black text-slate-900">
                  Hanya Cocokkan dengan Pesanan yang SUDAH SELESAI PACKING di Gudang (Direkomendasikan)
                </span>
                <p class="text-slate-500 text-[11px] mt-0.5 leading-relaxed">
                  Menjaga integritas data Order Marketplace di sistem. Hanya pesanan yang sudah dipack di gudang yang akan di-update status pencairan &amp; nominal cairnya tanpa mengubah data pesanan lainnya.
                </p>
              </div>
            </label>
          </div>

          <div v-if="importError" class="p-3 bg-slate-100 border border-slate-300 text-slate-800 rounded-xl text-xs font-semibold">
            {{ importError }}
          </div>

          <!-- Live Preview & Format Detection Table -->
          <div v-if="importRecords.length > 0" class="space-y-3">
            <div class="bg-slate-50 border border-slate-200 rounded-2xl p-3.5 flex items-center justify-between">
              <div>
                <div class="flex items-center gap-2">
                  <p class="text-xs font-black text-slate-900">✓ Data Terdeteksi &amp; Siap Direkonsiliasi</p>
                  <span
                    v-if="detectedReportType === 'TIKTOK_ON_HOLD'"
                    class="px-2 py-0.5 text-[10px] font-black rounded-md bg-slate-100 text-slate-800 border border-slate-300"
                  >
                    Format: TikTok On-Hold (Dana Tertahan)
                  </span>
                  <span
                    v-else-if="detectedReportType === 'TIKTOK_INCOME'"
                    class="px-2 py-0.5 text-[10px] font-black rounded-md bg-slate-100 text-slate-800 border border-slate-300"
                  >
                    Format: TikTok Income (Penghasilan Cair)
                  </span>
                </div>
                <p class="text-[11px] text-slate-600 mt-1">
                  {{ importSummaryStats.totalRows }} pesanan ditemukan:
                  <span class="font-bold">{{ importSummaryStats.settledCount }} Cair</span>,
                  <span class="font-bold">{{ importSummaryStats.holdCount }} On-Hold</span>,
                  <span class="font-bold">{{ importSummaryStats.refundCount }} Refund</span>.
                  Total estimasi dana: <span class="font-black">{{ formatCurrency(importSummaryStats.totalAmount) }}</span>.
                </p>
              </div>
              <span class="bg-slate-900 text-white font-black text-xs px-2.5 py-1 rounded-lg shrink-0">
                {{ importRecords.length }} Pesanan
              </span>
            </div>

            <!-- Table Preview 5 sample records -->
            <div>
              <p class="text-xs font-bold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <Sparkles class="h-3.5 w-3.5 text-slate-600" /> Pratinjau 5 Data Pertama dari Berkas:
              </p>
              <div class="border border-slate-200 rounded-xl overflow-hidden text-xs">
                <table class="w-full text-left">
                  <thead class="bg-slate-100 text-slate-700 font-bold border-b border-slate-200">
                    <tr>
                      <th class="p-2">No. Pesanan</th>
                      <th class="p-2">Status</th>
                      <th class="p-2 text-right">Estimasi / Dana Cair</th>
                      <th class="p-2">Alasan / Catatan Waktu</th>
                      <th class="p-2">Tgl Transaksi</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100">
                    <tr v-for="(rec, idx) in importRecords.slice(0, 5)" :key="idx" class="hover:bg-slate-50">
                      <td class="p-2 font-mono font-bold text-slate-800">{{ rec.order_id }}</td>
                      <td class="p-2">
                        <span
                          class="px-2 py-0.5 rounded text-[10px] font-bold bg-slate-100 text-slate-800 border border-slate-300"
                        >
                          {{ rec.status === 'HOLD' ? 'ON-HOLD' : (rec.status === 'REFUNDED' ? 'REFUND' : 'CAIR') }}
                        </span>
                      </td>
                      <td class="p-2 text-right font-black text-slate-900">
                        {{ formatCurrency(rec.settlement_amount) }}
                      </td>
                      <td class="p-2 text-slate-600 text-[11px]">
                        {{ rec.hold_reason || rec.estimated_settlement_notes || '-' }}
                      </td>
                      <td class="p-2 text-slate-500 font-medium">{{ formatDateOnly(rec.settlement_date) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        <div class="p-6 border-t border-slate-100 flex items-center justify-end gap-3 bg-slate-50">
          <button
            @click="showImportModal = false"
            class="px-4 py-2.5 rounded-xl border border-slate-200 text-slate-600 font-bold text-xs hover:bg-slate-100"
          >
            Batal
          </button>
          <button
            @click="submitImport"
            :disabled="importRecords.length === 0 || isImporting"
            class="px-5 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs shadow-xs disabled:opacity-50 flex items-center gap-1.5"
          >
            <Upload class="h-4 w-4" />
            {{ isImporting ? (importProgressText || 'Memproses...') : `Impor & Rekonsiliasi (${importRecords.length} Data)` }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Dialog: Tarik Finansial Marketplace -->
    <div
      v-if="showFinanceSyncModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
    >
      <div class="bg-white rounded-3xl shadow-2xl max-w-lg w-full overflow-hidden border border-slate-100 animate-in zoom-in-95 duration-200">
        <div class="p-6 border-b border-slate-100 flex items-center justify-between bg-white">
          <div class="flex items-center gap-3">
            <div class="p-2.5 bg-slate-900 text-white rounded-2xl shadow-xs">
              <Coins class="h-5 w-5" />
            </div>
            <div>
              <h3 class="text-base font-black text-slate-900">Tarik Finansial Marketplace</h3>
              <p class="text-xs text-slate-500 font-medium">TikTok Finance API atau Shopee Payment API.</p>
            </div>
          </div>
          <button
            @click="showFinanceSyncModal = false"
            class="p-2 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
          >
            <XCircle class="h-5 w-5" />
          </button>
        </div>

        <div class="p-6 space-y-4">
          <div class="flex gap-2">
            <button type="button" @click="selectedFinanceMarketplace = 'tiktok'" class="px-4 py-2 rounded-xl text-xs font-bold border" :class="selectedFinanceMarketplace === 'tiktok' ? 'bg-slate-900 text-white' : 'bg-white text-slate-700'">TikTok</button>
            <button type="button" @click="selectedFinanceMarketplace = 'shopee'" class="px-4 py-2 rounded-xl text-xs font-bold border" :class="selectedFinanceMarketplace === 'shopee' ? 'bg-slate-900 text-white' : 'bg-white text-slate-700'">Shopee</button>
          </div>
          <p class="text-xs font-bold text-slate-700 uppercase tracking-wider">Pilih rentang data finansial:</p>
          <p v-if="selectedFinanceMarketplace === 'tiktok'" class="text-xs text-slate-500">TikTok memilih berdasarkan tanggal statement. Tanggal pesanan dan tanggal dana dibayar bisa berbeda. Filter tanggal pada tabel diatur terpisah.</p>

          <div class="grid grid-cols-1 gap-2.5">
            <button
              type="button"
              @click="selectedSyncDays = 0"
              class="flex items-center justify-between p-3.5 rounded-2xl border-2 text-left transition-all"
              :class="selectedSyncDays === 0 ? 'border-slate-900 bg-slate-50/80 shadow-xs' : 'border-slate-200 hover:border-slate-300 bg-white'"
            >
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-xl" :class="selectedSyncDays === 0 ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-600'">
                  <Layers class="h-4 w-4" />
                </div>
                <div>
                  <p class="text-sm font-black text-slate-900">Semua Riwayat (Full Sync)</p>
                  <p class="text-xs text-slate-500">Menarik seluruh riwayat pencairan dan rincian fee pesanan yang tersedia melalui API</p>
                </div>
              </div>
            </button>

            <button
              type="button"
              @click="selectedSyncDays = 180"
              class="flex items-center justify-between p-3.5 rounded-2xl border-2 text-left transition-all"
              :class="selectedSyncDays === 180 ? 'border-slate-900 bg-slate-50/80 shadow-xs' : 'border-slate-200 hover:border-slate-300 bg-white'"
            >
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-xl" :class="selectedSyncDays === 180 ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-600'">
                  <Calendar class="h-4 w-4" />
                </div>
                <div>
                  <p class="text-sm font-black text-slate-900">180 Hari Terakhir (6 Bulan)</p>
                  <p class="text-xs text-slate-500">Membaca statement dalam 180 hari terakhir</p>
                </div>
              </div>
            </button>

            <button
              type="button"
              @click="selectedSyncDays = 90"
              class="flex items-center justify-between p-3.5 rounded-2xl border-2 text-left transition-all"
              :class="selectedSyncDays === 90 ? 'border-slate-900 bg-slate-50/80 shadow-xs' : 'border-slate-200 hover:border-slate-300 bg-white'"
            >
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-xl" :class="selectedSyncDays === 90 ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-600'">
                  <Calendar class="h-4 w-4" />
                </div>
                <div>
                  <p class="text-sm font-black text-slate-900">90 Hari Terakhir (3 Bulan)</p>
                  <p class="text-xs text-slate-500">Penarikan transaksi statement dari kuartal terakhir</p>
                </div>
              </div>
            </button>

            <button
              type="button"
              @click="selectedSyncDays = 30"
              class="flex items-center justify-between p-3.5 rounded-2xl border-2 text-left transition-all"
              :class="selectedSyncDays === 30 ? 'border-slate-900 bg-slate-50/80 shadow-xs' : 'border-slate-200 hover:border-slate-300 bg-white'"
            >
              <div class="flex items-center gap-3">
                <div class="p-2 rounded-xl" :class="selectedSyncDays === 30 ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-600'">
                  <RefreshCw class="h-4 w-4" />
                </div>
                <div>
                  <p class="text-sm font-black text-slate-900">30 Hari Terakhir (Cepat)</p>
                  <p class="text-xs text-slate-500">Membaca statement dalam 30 hari terakhir</p>
                </div>
              </div>
              <span v-if="selectedSyncDays === 30" class="text-xs font-black text-slate-900 bg-slate-200 px-2 py-0.5 rounded-lg border border-slate-300">Pilihan Utama</span>
            </button>
          </div>
        </div>

        <div class="p-6 border-t border-slate-100 flex items-center justify-end gap-3 bg-slate-50">
          <button
            @click="showFinanceSyncModal = false"
            class="px-4 py-2.5 rounded-xl border border-slate-200 text-slate-600 font-bold text-xs hover:bg-slate-100"
          >
            Batal
          </button>
          <button
            @click="handleSyncFinance(selectedSyncDays)"
            :disabled="isSyncingFinance"
            class="px-5 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs shadow-xs disabled:opacity-50 flex items-center gap-2"
          >
            <Coins class="h-4 w-4" :class="{ 'animate-spin': isSyncingFinance }" />
            {{ isSyncingFinance ? 'Menyinkronkan...' : 'Mulai Tarik Finansial' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Scoped transitions */
</style>
