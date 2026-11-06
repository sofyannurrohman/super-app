<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import {
  FileText,
  Printer,
  RefreshCw,
  Search,
  Package,
  Store,
  Truck,
  DollarSign,
  ChevronDown,
  ChevronRight,
  Calendar,
  User,
  ClipboardList,
  CheckCircle2,
  AlertTriangle,
  RotateCcw,
  Banknote,
  CreditCard,
  Clock,
  Hash,
  Save,
  Eye,
  EyeOff,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { Badge } from '@/components/ui/badge'

const showPrintPreview = ref(false)

const masterStore = useMasterDataStore()
const authStore = useAuthStore()

// ─── State ───────────────────────────────────────────────────────────────────
const route = useRoute()
const loading = ref(false)
const selectedLPSType = ref<'PRE_DELIVERY' | 'DELIVERY' | 'MOTORIS'>('PRE_DELIVERY')
const lpsData = ref<{
  salesman?: any
  lps_type?: string
  total_cash?: number
  total_transfer?: number
  total_tempo?: number
  total_stores_visited?: number
  total_transactions?: number
  total_grand_amount?: number
  delivery_batches?: { items?: any[], [key: string]: any }[]
  store_visits?: { transactions?: any[], [key: string]: any }[]
  [key: string]: any
} | null>(null)
const selectedDate = ref(new Date().toISOString().split('T')[0])
const selectedEmployeeId = ref('')
const employees = ref<any[]>([])
const expandedStores = ref<Set<number>>(new Set())
const expandedDeliveries = ref<Set<number>>(new Set())

const isSalesMotoris = computed(() => {
  if (lpsData.value?.salesman?.sales_type === 'motoris') return true
  const emp = employees.value.find(e => e.id === selectedEmployeeId.value)
  if (emp) {
    const job = (emp.job_position?.title || emp.job_position_title || '').toLowerCase()
    return job.includes('motoris') || job.includes('booster')
  }
  return false
})

watch(selectedEmployeeId, async (newId) => {
  if (!newId) return
  const emp = employees.value.find(e => e.id === newId)
  if (emp) {
    const job = (emp.job_position?.title || emp.job_position_title || '').toLowerCase()
    if (job.includes('motoris') || job.includes('booster')) {
      selectedLPSType.value = 'MOTORIS'
    } else {
      selectedLPSType.value = 'PRE_DELIVERY'
    }
  }
  await fetchLPS()
})

watch(selectedDate, async (newDate) => {
  if (newDate && selectedEmployeeId.value) {
    await fetchLPS()
  }
})

// ─── Computed Properties for Spreadsheet & DO Summary ────────────────────────
const uniqueProducts = computed(() => {
  if (!lpsData.value) return []
  const products = new Set<string>()
  const giftKeywords = ['gelas', 'piring', 'mangkuk']

  lpsData.value.delivery_batches?.forEach((batch: any) => {
    batch.items?.forEach((item: any) => {
      if (item.product_name) {
        const nameLower = item.product_name.toLowerCase()
        if (!giftKeywords.some(g => nameLower.includes(g))) {
          products.add(item.product_name)
        }
      }
    })
  })

  lpsData.value.store_visits?.forEach((visit: any) => {
    visit.transactions?.forEach((tx: any) => {
      tx.items?.forEach((item: any) => {
        if (item.product_name) {
          const nameLower = item.product_name.toLowerCase()
          if (!giftKeywords.some(g => nameLower.includes(g))) {
            products.add(item.product_name)
          }
        }
      })
    })
  })
  return Array.from(products).sort()
})

const getProductQtyForStore = (visit: any, productName: string) => {
  let qty = 0
  visit.transactions?.forEach((tx: any) => {
    tx.items?.forEach((item: any) => {
      if (item.product_name === productName) qty += item.quantity || 0
    })
  })
  return qty
}

const getTotalProductQty = (productName: string) => {
  let total = 0
  lpsData.value?.store_visits?.forEach((visit: any) => {
    total += getProductQtyForStore(visit, productName)
  })
  return total
}

const getPaymentMethods = (visit: any) => {
  const methods = new Set<string>()
  visit.transactions?.forEach((tx: any) => {
    if (tx.payment_method) methods.add(tx.payment_method)
  })
  return Array.from(methods)
}

const getGiftQtyForStore = (visit: any, giftInitial: string) => {
  let qty = 0
  const matchStr = giftInitial === 'G' ? 'gelas' : giftInitial === 'P' ? 'piring' : giftInitial === 'M' ? 'mangkuk' : ''
  if (!matchStr) return 0
  visit.transactions?.forEach((tx: any) => {
    tx.items?.forEach((item: any) => {
      if (item.product_name && item.product_name.toLowerCase().includes(matchStr)) {
        qty += item.quantity || 0
      }
    })
  })
  return qty
}

const getTotalGiftQty = (giftInitial: string) => {
  let total = 0
  lpsData.value?.store_visits?.forEach((visit: any) => {
    total += getGiftQtyForStore(visit, giftInitial)
  })
  return total
}

const getReceiptNumbers = (visit: any) => {
  if (!visit.transactions?.length) return '-'
  const receipts = visit.transactions.map((tx: any) => tx.receipt_no || tx.so_number).filter(Boolean)
  return receipts.length ? receipts.join(', ') : '-'
}

const getVisitConfirmationStatus = (visit: any) => {
  if (!visit.transactions?.length) return { isConfirmed: true, label: 'Terkonfirmasi Admin', isConverted: false }
  const isConverted = visit.transactions.some((tx: any) => (tx.status || '').toUpperCase() === 'CONVERTED')
  const hasUnconfirmed = visit.transactions.some((tx: any) => tx.is_confirmed === false && (tx.status || '').toUpperCase() !== 'CONVERTED')
  if (hasUnconfirmed) {
    return { isConfirmed: false, label: 'Belum Dikonfirmasi Admin', isConverted: false }
  }
  if (isConverted) {
    return { isConfirmed: true, label: 'Terkonversi ke Faktur', isConverted: true }
  }
  return { isConfirmed: true, label: 'Terkonfirmasi Admin', isConverted: false }
}

const unconfirmedCount = computed(() => {
  if (!lpsData.value?.store_visits) return 0
  let count = 0
  lpsData.value.store_visits.forEach((visit: any) => {
    visit.transactions?.forEach((tx: any) => {
      const statusUpper = (tx.status || '').toUpperCase()
      if (tx.is_confirmed === false && statusUpper !== 'CONVERTED') count++
    })
  })
  return count
})

const doSummary = computed(() => {
  if (!lpsData.value) return []
  const summary: Record<string, { muat: number, pulang: number, retur: number }> = {}
  
  if (lpsData.value.delivery_batches) {
    lpsData.value.delivery_batches.forEach((batch: any) => {
      batch.items?.forEach((item: any) => {
        const name = item.product_name || 'Tanpa Nama'
        if (!summary[name]) summary[name] = { muat: 0, pulang: 0, retur: 0 }
        summary[name].muat += item.quantity_out || 0
        summary[name].pulang += item.quantity_return || 0
      })
    })
  }

  if (lpsData.value.sales_returns) {
    lpsData.value.sales_returns.forEach((sr: any) => {
      sr.items?.forEach((item: any) => {
        const name = item.product_name || 'Tanpa Nama'
        if (!summary[name]) summary[name] = { muat: 0, pulang: 0, retur: 0 }
        summary[name].retur += item.quantity || 0
      })
    })
  }

  return Object.keys(summary).sort().map(name => ({
    name,
    muat: summary[name]!.muat,
    pulang: summary[name]!.pulang,
    retur: summary[name]!.retur,
    terjual: summary[name]!.muat - summary[name]!.pulang - summary[name]!.retur
  }))
})

// ─── Fetch employees (salesman list) ─────────────────────────────────────────
async function fetchEmployees() {
  try {
    const companyId = masterStore.selectedBranchCompanyId || ''
    const params: any = { is_sales: true, limit: 'all', company_id: companyId }
    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterStore.selectedBranchId
    }
    
    const res = await apiClient.get('/employees', { params })
    employees.value = res.data?.data || res.data || []
    
    // Reset selection if current employee is no longer in the list
    if (selectedEmployeeId.value && !employees.value.find(e => e.id === selectedEmployeeId.value)) {
      selectedEmployeeId.value = ''
    }

    // Auto-select salesman if none selected and list is available (respecting route.query if present)
    if (!selectedEmployeeId.value && employees.value.length > 0) {
      if (route.query.salesman_id && employees.value.some(e => e.id === route.query.salesman_id)) {
        selectedEmployeeId.value = route.query.salesman_id as string
      } else {
        selectedEmployeeId.value = employees.value[0].id
      }
      await fetchLPS()
    }
  } catch {
    // Silently fail - use empty list
  }
}

watch(() => masterStore.selectedBranchId, () => {
  fetchEmployees()
})

// ─── Fetch LPS Data ───────────────────────────────────────────────────────────
async function fetchLPS(forceLive = false) {
  if (!selectedDate.value || !selectedEmployeeId.value) return
  loading.value = true
  lpsData.value = null
  try {
    const params: any = {
      date: selectedDate.value,
      employee_id: selectedEmployeeId.value,
      lps_type: selectedLPSType.value.toLowerCase()
    }
    if (forceLive) {
      params.force_live = 'true'
    }
    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterStore.selectedBranchId
    }

    const res = await apiClient.get('/admin/lps/daily', { params })
    const rawData = res.data

    if (rawData?.lps_type && rawData.is_saved && rawData.status === 'APPROVED') {
      selectedLPSType.value = rawData.lps_type as any
    }

    // Filter out visits with no transactions and deduplicate visits/transactions
    // Group by Store AND Payment Method (as requested: "kalau metode pembayaran berbeda di pisah saja")
    if (rawData?.store_visits) {
      const uniqueVisitsMap = new Map()
      const uniquePhysicalStores = new Set()
      
      rawData.store_visits.forEach((visit: any) => {
        if (visit.transactions && visit.transactions.length > 0) {
          uniquePhysicalStores.add(visit.store_id || visit.store_name)
          visit.transactions.forEach((tx: any) => {
            const method = (tx.payment_method || 'UNKNOWN').toUpperCase()
            const key = (visit.store_id || visit.store_name) + '_' + method
            
            if (uniqueVisitsMap.has(key)) {
              const existing = uniqueVisitsMap.get(key)
              const isDuplicateTx = existing.transactions.some((etx: any) => {
                if (tx.id && etx.id && tx.id === etx.id) return true
                if (tx.receipt_no && etx.receipt_no && tx.receipt_no === etx.receipt_no) return true
                if (tx.so_number && etx.so_number && tx.so_number === etx.so_number) return true
                if (tx.receipt_no && etx.so_number && tx.receipt_no === etx.so_number) return true
                if (tx.so_number && etx.receipt_no && tx.so_number === etx.receipt_no) return true
                return false
              })
              if (!isDuplicateTx) {
                existing.transactions.push(tx)
              }
            } else {
              uniqueVisitsMap.set(key, { ...visit, transactions: [tx] })
            }
          })
        }
      })
      
      rawData.store_visits = Array.from(uniqueVisitsMap.values())

      // Kalkulasi ulang total agar otentik sesuai data yang sudah tidak duplikat
      let calcGrandAmount = 0
      let calcTransactions = 0
      let calcCash = 0
      let calcTransfer = 0
      let calcTempo = 0
      let calcStoresVisited = uniquePhysicalStores.size

      rawData.store_visits.forEach((visit: any) => {
        if (visit.transactions) {
          calcTransactions += visit.transactions.length
          let visitTotalTagihan = 0
          let visitTotalPembayaran = 0
          visit.transactions.forEach((tx: any) => {
            const tagihan = tx.total_amount || 0
            const dibayar = tx.payment_amount || tx.paid_amount || (rawData.lps_type === 'PRE_DELIVERY' ? 0 : tx.total_amount) || 0
            
            calcGrandAmount += (rawData.lps_type === 'PRE_DELIVERY' ? tagihan : dibayar)
            visitTotalTagihan += tagihan
            visitTotalPembayaran += dibayar
            
            const method = (tx.payment_method || '').toUpperCase()
            const amountForBreakdown = rawData.lps_type === 'PRE_DELIVERY' ? tagihan : dibayar
            if (method === 'CASH') calcCash += amountForBreakdown
            else if (method === 'TEMPO') calcTempo += amountForBreakdown
            else if (method.includes('QRIS') || method.includes('TRANSFER') || method.includes('VA')) calcTransfer += amountForBreakdown
            else calcTempo += amountForBreakdown
          })
          visit.total_amount = visitTotalTagihan // Original full invoice/order amount
          visit.total_pembayaran = visitTotalPembayaran // Actual received amount
        }
      })

      rawData.total_grand_amount = calcGrandAmount
      rawData.total_transactions = calcTransactions
      rawData.total_stores_visited = calcStoresVisited
      rawData.total_cash = calcCash
      rawData.total_transfer = calcTransfer
      rawData.total_tempo = calcTempo
    }

    lpsData.value = rawData

    // Auto-expand all stores for print readiness
    if (lpsData.value?.store_visits) {
      lpsData.value.store_visits.forEach((_: any, i: number) => expandedStores.value.add(i))
    }
    if (lpsData.value?.delivery_batches) {
      lpsData.value.delivery_batches.forEach((_: any, i: number) => expandedDeliveries.value.add(i))
    }
  } catch (err: any) {
    toast.error(err?.response?.data?.error || 'Gagal memuat data LPS')
  } finally {
    loading.value = false
  }
}

function toggleStore(index: number) {
  if (expandedStores.value.has(index)) expandedStores.value.delete(index)
  else expandedStores.value.add(index)
}

function toggleDelivery(index: number) {
  if (expandedDeliveries.value.has(index)) expandedDeliveries.value.delete(index)
  else expandedDeliveries.value.add(index)
}

function expandAll() {
  lpsData.value?.store_visits?.forEach((_: any, i: number) => expandedStores.value.add(i))
  lpsData.value?.delivery_batches?.forEach((_: any, i: number) => expandedDeliveries.value.add(i))
}

const saving = ref(false)

async function saveLPS() {
  if (!lpsData.value || !selectedDate.value || !selectedEmployeeId.value) return
  saving.value = true
  try {
    await apiClient.post('/admin/lps/daily/save', {
      date: selectedDate.value,
      employee_id: selectedEmployeeId.value,
      lps_type: selectedLPSType.value,
      snapshot_data: lpsData.value
    })
    toast.success('LPS berhasil disimpan permanen')
    await fetchLPS() // Refresh status
  } catch (err: any) {
    toast.error('Gagal menyimpan LPS: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

const approving = ref(false)

async function approveLPS() {
  if (!lpsData.value || !selectedDate.value || !selectedEmployeeId.value) return
  approving.value = true
  try {
    await apiClient.post('/admin/lps/daily/approve', {
      date: selectedDate.value,
      employee_id: selectedEmployeeId.value,
      lps_type: selectedLPSType.value,
    })
    toast.success('Persetujuan berhasil! Transaksi terkait telah diperbarui.')
    await fetchLPS() // Refresh status
  } catch (err: any) {
    toast.error('Gagal menyetujui LPS: ' + (err.response?.data?.error || err.message))
  } finally {
    approving.value = false
  }
}

function handlePrint() {
  expandAll()
  // Ensure print preview data is ready, then print
  setTimeout(() => window.print(), 150)
}

function togglePrintPreview() {
  showPrintPreview.value = !showPrintPreview.value
  if (showPrintPreview.value) expandAll()
}

const formatCurrency = (val: number) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val || 0)

const formatDate = (d: string) => {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('id-ID', { weekday: 'long', day: '2-digit', month: 'long', year: 'numeric' })
}

const formatTime = (d: string) => {
  if (!d) return '-'
  return new Date(d).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}

const paymentLabel = (method: string) => {
  const map: Record<string, string> = {
    CASH: 'Tunai', QRIS: 'QRIS', TEMPO: 'Tempo', VA: 'Transfer VA',
    TRANSFER: 'Transfer', 'MIDTRANS_QRIS': 'QRIS', 'MIDTRANS_VA': 'Transfer VA'
  }
  return map[method?.toUpperCase()] || method || '-'
}

const paymentColor = (method: string) => {
  const m = method?.toUpperCase()
  if (m === 'CASH') return 'text-slate-700 bg-slate-50 border-slate-200'
  if (m === 'QRIS' || m === 'MIDTRANS_QRIS') return 'text-slate-700 bg-slate-50 border-slate-200'
  if (m === 'TEMPO') return 'text-slate-700 bg-slate-50 border-slate-200'
  return 'text-slate-700 bg-slate-50 border-slate-200'
}

const statusColor = (status: string) => {
  if (status === 'COMPLETED' || status === 'DELIVERED' || status === 'PAID') return 'text-slate-700 bg-slate-50 border-slate-200'
  if (status === 'PENDING') return 'text-slate-700 bg-slate-50 border-slate-200'
  if (status === 'FAILED') return 'text-slate-700 bg-slate-50 border-slate-200'
  return 'text-slate-600 bg-slate-50 border-slate-200'
}

// Print date label
const reportDateLabel = computed(() => formatDate(selectedDate.value || ''))

const reportTitleLabel = computed(() => {
  const dateFormatted = reportDateLabel.value.toUpperCase()
  if (selectedLPSType.value === 'PRE_DELIVERY') {
    return `DAFTAR LPS PESANAN (SEBELUM PENGIRIMAN), ${dateFormatted}`
  } else if (selectedLPSType.value === 'DELIVERY') {
    return `DAFTAR LPS PENGIRIMAN (SURAT JALAN), ${dateFormatted}`
  } else {
    return `DAFTAR LPS FAKTUR PENJUALAN (MOTORIS), ${dateFormatted}`
  }
})

const docBadgeLabel = computed(() => {
  if (selectedLPSType.value === 'PRE_DELIVERY') {
    return 'LEMBAR PENJUALAN SALES (LPS) — SEBELUM PENGIRIMAN'
  } else if (selectedLPSType.value === 'DELIVERY') {
    return 'LEMBAR PENJUALAN SALES (LPS) — HASIL PENGIRIMAN'
  } else {
    return 'LEMBAR PENJUALAN SALES (LPS) — FAKTUR MOTORIS'
  }
})

const doSummaryTitle = computed(() => {
  if (selectedLPSType.value === 'PRE_DELIVERY') {
    return 'REKAP KEBUTUHAN BARANG MUAT GUDANG (PESANAN SO TERKONFIRMASI)'
  } else if (selectedLPSType.value === 'DELIVERY') {
    return 'REKAP BARANG SURAT JALAN (MUAT/PULANG/RETUR/TERJUAL)'
  } else {
    return 'REKAP BARANG BAWAAN (TRANSFER STOK MOTORIS)'
  }
})

const lpsProducts = computed(() => {
  if (!lpsData.value) return []
  const productMap = new Map()


  lpsData.value.delivery_batches?.forEach((batch: any) => {
    batch.items?.forEach((item: any) => {
      const pName = item.product_name
      if (pName && !productMap.has(pName)) {
        productMap.set(pName, { name: pName.split(' ').map((w: string) => w[0]).join(''), fullName: pName, out: 0, return: 0 })
      }
      if (pName) {
        const p = productMap.get(pName)
        p.out += (item.quantity_out || 0)
        p.return += (item.quantity_return || 0)
      }
    })
  })

  lpsData.value.store_visits?.forEach((visit: any) => {
    visit.transactions?.forEach((tx: any) => {
      tx.items?.forEach((item: any) => {
        const pName = item.product_name
        if (pName && !productMap.has(pName)) {
          productMap.set(pName, { name: pName.split(' ').map((w: string) => w[0]).join(''), fullName: pName, out: 0, return: 0 })
        }
      })
    })
  })

  return Array.from(productMap.values())
})

const getTransactionProductQty = (tx: any, productName: string) => {
  if (!tx || !tx.items) return ''
  const item = tx.items.find((i: any) => i.product_name === productName)
  return item ? item.quantity : ''
}

const getGiftQtyForVisitTx = (tx: any, giftInitial: string) => {
  if (!tx || !tx.items) return 0
  const matchStr = giftInitial === 'G' ? 'gelas' : giftInitial === 'P' ? 'piring' : 'mangkuk'
  return tx.items.reduce((sum: number, item: any) => {
    if (item.product_name && item.product_name.toLowerCase().includes(matchStr)) {
      return sum + (item.quantity || 0)
    }
    return sum
  }, 0)
}

onMounted(async () => {
  await fetchEmployees()

  // Auto-load from query parameters if exist
  if (route.query.date) {
    selectedDate.value = route.query.date as string
  }
  if (route.query.salesman_id) {
    selectedEmployeeId.value = route.query.salesman_id as string
    fetchLPS()
  }
})
</script>

<template>
  <div class="lps-page">

    <!-- ═══════════════════════════ PRINT VIEW (visible on print OR preview mode) ══════════════════════════ -->
    <div
      id="lps-print-view"
      :class="['lps-print-area', showPrintPreview ? 'lps-preview-visible' : 'lps-print-hidden']"
      v-if="lpsData"
    >
      <!-- Title -->
      <div class="lpt-title-block">
        <div class="lpt-title">{{ reportTitleLabel }}</div>
        <div class="lpt-sales-info">
          <span>SALES : {{ lpsData.salesman?.name || '.....................' }}</span>
        </div>
      </div>

      <!-- Main LPS Table -->
      <table class="lpt-table">
        <thead>
          <tr>
            <th class="lpt-th lpt-no">No</th>
            <th class="lpt-th lpt-nota">No Nota</th>
            <th class="lpt-th lpt-toko">Nama Toko</th>
            <th class="lpt-th lpt-alamat">Alamat</th>
            <th class="lpt-th lpt-hp">No HP</th>
            <th class="lpt-th lpt-tagihan">Total Tagihan</th>
            <th class="lpt-th lpt-galon">Galon Kosong</th>
            <th class="lpt-th lpt-botol">Botol Kosong</th>
            <th class="lpt-th lpt-bayar">Metode Pembayaran</th>
            <th class="lpt-th lpt-tagihan">Total Dibayar</th>
            <!-- Dynamic products -->
            <th v-for="prod in uniqueProducts" :key="prod" class="lpt-th lpt-prod" :title="prod">{{ prod }}</th>
            <!-- Hadiah -->
            <th class="lpt-th lpt-gift" colspan="3">Hadiah</th>
            <!-- Retur -->
            <th class="lpt-th lpt-retur">Retur</th>
          </tr>
          <!-- Sub-header for hadiah -->
          <tr class="lpt-subhead">
            <th class="lpt-th" colspan="10"></th>
            <th v-for="prod in uniqueProducts" :key="prod" class="lpt-th"></th>
            <th class="lpt-th lpt-gift-sub">Gelas</th>
            <th class="lpt-th lpt-gift-sub">Piring</th>
            <th class="lpt-th lpt-gift-sub">Mangkok</th>
            <th class="lpt-th"></th>
          </tr>
        </thead>
        <tbody>
          <template v-for="(visit, vi) in lpsData.store_visits" :key="vi">
            <tr class="lpt-tr">
              <td class="lpt-td lpt-center">{{ vi + 1 }}</td>
              <td class="lpt-td lpt-small">
                <div>{{ getReceiptNumbers(visit) }}</div>
                <div v-if="!getVisitConfirmationStatus(visit).isConfirmed" style="font-size: 8px; color: #b45309; font-weight: bold;">
                  [Belum Konfirmasi]
                </div>
              </td>
              <td class="lpt-td">{{ visit.store_name }}</td>
              <td class="lpt-td">{{ visit.store_address || '-' }}</td>
              <td class="lpt-td">{{ visit.store_phone || '-' }}</td>
              <td class="lpt-td lpt-right">{{ formatCurrency(visit.total_amount || 0).replace('Rp\u00a0','').replace('Rp','').trim() }}</td>
              <td class="lpt-td lpt-center"></td>
              <td class="lpt-td lpt-center"></td>
              <td class="lpt-td lpt-center">
                <span v-for="(method, mi) in getPaymentMethods(visit)" :key="mi">
                  {{ paymentLabel(method) }}<br v-if="mi < getPaymentMethods(visit).length - 1" />
                </span>
              </td>
              <td class="lpt-td lpt-right">{{ formatCurrency(visit.total_pembayaran || 0).replace('Rp\u00a0','').replace('Rp','').trim() }}</td>
              <td v-for="prod in uniqueProducts" :key="prod" class="lpt-td lpt-center">{{ getProductQtyForStore(visit, prod) || '' }}</td>
              <td class="lpt-td lpt-center">{{ getGiftQtyForStore(visit, 'G') || '' }}</td>
              <td class="lpt-td lpt-center">{{ getGiftQtyForStore(visit, 'P') || '' }}</td>
              <td class="lpt-td lpt-center">{{ getGiftQtyForStore(visit, 'M') || '' }}</td>
              <td class="lpt-td lpt-center"></td>
            </tr>
          </template>
        </tbody>
        <tfoot>
          <tr class="lpt-total-row">
            <td class="lpt-td lpt-center" colspan="5"><strong>JUMLAH</strong></td>
            <td class="lpt-td lpt-right"><strong>{{ formatCurrency((lpsData.store_visits || []).reduce((sum: number, v: any) => sum + (v.total_amount || 0), 0)).replace('Rp\u00a0','').replace('Rp','').trim() }}</strong></td>
            <td class="lpt-td lpt-center"></td>
            <td class="lpt-td lpt-center"></td>
            <td class="lpt-td"></td>
            <td class="lpt-td lpt-right"><strong>{{ formatCurrency(lpsData.total_grand_amount || 0).replace('Rp\u00a0','').replace('Rp','').trim() }}</strong></td>
            <td v-for="prod in uniqueProducts" :key="prod" class="lpt-td lpt-center"><strong>{{ getTotalProductQty(prod) || '' }}</strong></td>
            <td class="lpt-td lpt-center"><strong>{{ getTotalGiftQty('G') || '' }}</strong></td>
            <td class="lpt-td lpt-center"><strong>{{ getTotalGiftQty('P') || '' }}</strong></td>
            <td class="lpt-td lpt-center"><strong>{{ getTotalGiftQty('M') || '' }}</strong></td>
            <td class="lpt-td"></td>
          </tr>
          <tr class="lpt-grand-row">
            <td class="lpt-td" colspan="5"><strong>TOTAL</strong></td>
            <td class="lpt-td lpt-right" colspan="10"><strong>{{ formatCurrency(lpsData.total_grand_amount || 0).replace('Rp\u00a0','').replace('Rp','').trim() }}</strong></td>
          </tr>
        </tfoot>
      </table>

      <!-- DO Summary in Print View -->
      <table class="lpt-table" style="margin-top: 2mm;">
        <thead>
          <tr>
            <th class="lpt-th" style="text-align: left;">{{ doSummaryTitle }}</th>
            <th class="lpt-th">MUAT</th>
            <th class="lpt-th">PULANG</th>
            <th class="lpt-th">RETUR</th>
            <th class="lpt-th">TERJUAL</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="doSummary.length === 0">
            <td class="lpt-td lpt-center" colspan="5">Tidak ada rekapan barang / muatan hari ini</td>
          </tr>
          <tr v-for="item in doSummary" :key="item.name">
            <td class="lpt-td"><strong>{{ item.name }}</strong></td>
            <td class="lpt-td lpt-center">{{ item.muat }}</td>
            <td class="lpt-td lpt-center">{{ item.pulang }}</td>
            <td class="lpt-td lpt-center">{{ item.retur }}</td>
            <td class="lpt-td lpt-center"><strong>{{ item.terjual }}</strong></td>
          </tr>
        </tbody>
      </table>

      <!-- Signature -->
      <div class="lpt-sign-area">
        <div class="lpt-sign-box"><div class="lpt-sign-label">ADMIN GUDANG</div><div class="lpt-sign-line"></div></div>
        <div class="lpt-sign-box"><div class="lpt-sign-label">ADMIN NOTA</div><div class="lpt-sign-line"></div></div>
        <div class="lpt-sign-box"><div class="lpt-sign-label">ADMIN KASIR</div><div class="lpt-sign-line"></div></div>
        <div class="lpt-sign-box"><div class="lpt-sign-label">SUPERVISOR SALES</div><div class="lpt-sign-line"></div></div>
        <div class="lpt-sign-box"><div class="lpt-sign-label">SALESMAN</div><div class="lpt-sign-line"></div></div>
      </div>
    </div>

    <!-- ═══════════════════════════ SCREEN VIEW (hidden on print) ════════════════════════════ -->
    <!-- SCREEN CONTROLS (hidden on print) -->
    <div class="no-print space-y-6 mb-8">
      <!-- Header -->
      <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-5 bg-white p-7 rounded-[2rem] shadow-sm border border-slate-100">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 rounded-2xl bg-slate-800 flex items-center justify-center shadow-sm">
            <FileText class="h-7 w-7 text-white" />
          </div>
          <div>
            <div class="flex items-center flex-wrap gap-2 sm:gap-3">
              <h1 class="text-2xl font-black text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
                <span>Lembar LPS</span>
                <span v-if="masterStore.selectedBranchName" class="text-primary text-xl font-bold">
                  — {{ masterStore.selectedBranchName }}
                </span>
              </h1>
              <Badge v-if="lpsData?.is_saved" variant="default" class="bg-emerald-100 text-emerald-800 border-emerald-200">
                Tersimpan ({{ lpsData.status }})
              </Badge>
            </div>
            <p class="text-sm text-slate-500 font-medium mt-0.5">Laporan Penjualan Sales Harian — Rekap DO, Kunjungan & Setoran Kasir</p>
          </div>
        </div>
        <div class="flex flex-wrap gap-3">
          <button
            @click="() => fetchLPS(false)"
            :disabled="loading"
            class="flex items-center gap-2 px-5 py-2.5 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 text-sm font-semibold transition-all"
            title="Muat data LPS"
          >
            <RefreshCw :class="['h-4 w-4', loading && 'animate-spin']" />
            Refresh
          </button>

          <button
            v-if="lpsData?.is_saved"
            @click="() => fetchLPS(true)"
            :disabled="loading"
            class="flex items-center gap-2 px-4 py-2.5 rounded-xl border border-amber-300 bg-amber-50 text-amber-800 hover:bg-amber-100 text-sm font-semibold transition-all shadow-xs"
            title="Bypass snapshot dan muat data transaksi terkini (live)"
          >
            <RefreshCw :class="['h-4 w-4', loading && 'animate-spin']" />
            Muat Data Live
          </button>

          <button
            v-if="lpsData"
            @click="togglePrintPreview"
            class="flex items-center gap-2 px-5 py-2.5 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 text-sm font-semibold transition-all"
          >
            <EyeOff v-if="showPrintPreview" class="h-4 w-4" />
            <Eye v-else class="h-4 w-4" />
            {{ showPrintPreview ? 'Sembunyikan Preview' : 'Preview Cetak' }}
          </button>
          
          <button
            v-if="lpsData && !lpsData.is_saved"
            @click="saveLPS"
            :disabled="saving"
            class="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-bold transition-all shadow-sm disabled:opacity-50"
          >
            <span v-if="saving" class="h-4 w-4 rounded-full border-2 border-white border-t-transparent animate-spin" />
            <Save v-else class="h-4 w-4" />
            Tutup Buku
          </button>

          <button
            v-if="lpsData && lpsData.is_saved && lpsData.status !== 'APPROVED' && ['CASHIER', 'SUPER_ADMIN'].includes(authStore.user?.role)"
            @click="approveLPS"
            :disabled="approving"
            class="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-emerald-500 hover:bg-emerald-600 text-white text-sm font-bold transition-all shadow-sm disabled:opacity-50"
          >
            <span v-if="approving" class="h-4 w-4 rounded-full border-2 border-white border-t-transparent animate-spin" />
            <CheckCircle2 v-else class="h-4 w-4" />
            Terima Setoran & Validasi Transaksi
          </button>

          <button
            @click="handlePrint"
            :disabled="!lpsData"
            class="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-900 text-white text-sm font-bold transition-all shadow-sm disabled:opacity-50"
          >
            <Printer class="h-4 w-4" />
            Cetak LPS
          </button>
        </div>
      </div>

      <!-- Filter Bar -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-5 space-y-4">
        <div class="flex flex-wrap gap-4 items-end">
          <!-- Date picker -->
          <div class="flex-1 min-w-[180px]">
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1.5">Tanggal</label>
            <div class="relative">
              <Calendar class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <input
                v-model="selectedDate"
                type="date"
                class="w-full pl-9 pr-4 py-2.5 rounded-xl border border-slate-200 focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100 outline-none text-sm font-medium transition-all"
              />
            </div>
          </div>

          <!-- Salesman selector -->
          <div class="flex-1 min-w-[220px]">
            <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1.5">Salesman</label>
            <div class="relative">
              <User class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <select
                v-model="selectedEmployeeId"
                class="w-full pl-9 pr-8 py-2.5 rounded-xl border border-slate-200 focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100 outline-none text-sm font-medium transition-all appearance-none bg-white"
              >
                <option value="" disabled>— Pilih Salesman —</option>
                <option v-for="emp in employees" :key="emp.id" :value="emp.id">
                  {{ emp.first_name }} {{ emp.last_name }}
                </option>
              </select>
              <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" />
            </div>
          </div>

          <button
            @click="() => fetchLPS(false)"
            :disabled="loading || !selectedEmployeeId"
            class="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-slate-900 text-white text-sm font-bold transition-all hover:bg-slate-700 disabled:opacity-50"
          >
            <Search class="h-4 w-4" />
            Tampilkan LPS
          </button>
        </div>

        <!-- LPS Mechanism Tab Switcher -->
        <div class="flex flex-wrap items-center justify-between border-t border-slate-100 pt-4 gap-3">
          <div class="flex items-center gap-2">
            <span class="text-xs font-bold text-slate-400 uppercase tracking-wider">Mekanisme LPS:</span>
            <div class="inline-flex p-1 bg-slate-100 rounded-xl border border-slate-200">
              <button
                type="button"
                @click="() => { selectedLPSType = 'PRE_DELIVERY'; fetchLPS() }"
                :class="[
                  'flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all',
                  selectedLPSType === 'PRE_DELIVERY' ? 'bg-white text-indigo-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'
                ]"
              >
                <ClipboardList class="h-3.5 w-3.5" />
                LPS Sebelum Pengiriman (Pesanan SO)
              </button>
              <button
                type="button"
                @click="() => { selectedLPSType = 'DELIVERY'; fetchLPS() }"
                :class="[
                  'flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all',
                  selectedLPSType === 'DELIVERY' ? 'bg-white text-indigo-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'
                ]"
              >
                <Truck class="h-3.5 w-3.5" />
                LPS Pengiriman (Surat Jalan)
              </button>
              <button
                type="button"
                @click="() => { selectedLPSType = 'MOTORIS'; fetchLPS() }"
                :class="[
                  'flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-bold transition-all',
                  selectedLPSType === 'MOTORIS' ? 'bg-white text-amber-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'
                ]"
              >
                <Package class="h-3.5 w-3.5" />
                LPS Faktur Penjualan (Motoris)
              </button>
            </div>
          </div>

          <div v-if="lpsData?.lps_type" class="text-xs font-medium text-slate-500">
            Tipe Terpilih: <span class="font-black text-slate-800 uppercase">{{ lpsData.lps_type }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════════════════════
         LOADING STATE
    ═══════════════════════════════════════════════════════════════════════ -->
    <div v-if="loading" class="no-print flex flex-col items-center justify-center py-24 gap-4">
      <div class="w-14 h-14 rounded-full border-4 border-slate-100 border-t-slate-800 animate-spin" />
      <p class="text-slate-500 font-medium">Memuat data LPS...</p>
    </div>

    <!-- ═══════════════════════════════════════════════════════════════════════
         EMPTY STATE
    ═══════════════════════════════════════════════════════════════════════ -->
    <div v-else-if="!lpsData && !loading" class="no-print flex flex-col items-center justify-center py-20 px-4 text-center">
      <div class="max-w-md w-full bg-white rounded-3xl border border-slate-100 shadow-sm p-8 space-y-4">
        <div class="w-16 h-16 rounded-2xl bg-amber-50 border border-amber-100 flex items-center justify-center mx-auto text-amber-600">
          <ClipboardList class="h-8 w-8" />
        </div>
        <div>
          <h3 class="text-lg font-bold text-slate-800">
            {{ selectedEmployeeId ? 'Tidak Ada Data LPS' : 'Pilih Salesman & Tanggal' }}
          </h3>
          <p class="text-slate-500 text-sm mt-1.5 leading-relaxed">
            <template v-if="!selectedEmployeeId">
              Silakan pilih Salesman pada filter di atas dan klik <strong>"Tampilkan LPS"</strong> untuk memuat lembar penjualan.
            </template>
            <template v-else>
              Tidak ditemukan data transaksi untuk Salesman dan tanggal yang dipilih pada mekanisme <strong>{{ selectedLPSType }}</strong>.
            </template>
          </p>
        </div>

        <div v-if="selectedEmployeeId" class="text-left bg-slate-50 p-4 rounded-2xl border border-slate-200/60 text-xs text-slate-600 space-y-2">
          <div class="font-bold text-slate-700 flex items-center gap-1.5">
            <AlertTriangle class="h-4 w-4 text-amber-500 shrink-0" />
            Tips Pemeriksaan:
          </div>
          <ul class="list-disc list-inside space-y-1 text-slate-500 pl-1">
            <li><strong>LPS Pre-Delivery:</strong> Hanya menampilkan Sales Order yang sudah dikonfirmasi Admin Nota (status CONFIRMED/APPROVED/IN_DELIVERY/PAID).</li>
            <li><strong>LPS Pengiriman:</strong> Hanya menampilkan data jika sudah dibuatkan Surat Jalan (Delivery Batch).</li>
            <li><strong>LPS Motoris:</strong> Hanya menampilkan transaksi kanvas langsung yang berstatus VERIFIED/APPROVED/PAID.</li>
            <li>Periksa juga apakah filter <strong>Cabang Global</strong> di header sudah sesuai.</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════════════════════
         LPS DOCUMENT (shown on screen & print)
    ═══════════════════════════════════════════════════════════════════════ -->
    <div v-if="lpsData" id="surat-jalan-card" class="w-full max-w-[95%] sm:max-w-6xl mx-auto bg-white shadow-2xl border border-slate-200 rounded-[2rem] overflow-hidden my-8 print:my-0 print:border-0 print:shadow-none print:max-w-full print:rounded-none relative z-10 lps-document">
      
      <!-- Unconfirmed Warning Banner if any -->
      <div v-if="unconfirmedCount > 0" class="no-print bg-amber-500/10 border-b border-amber-200 px-6 py-2.5 flex items-center justify-between text-xs text-amber-800">
        <div class="flex items-center gap-2 font-medium">
          <AlertTriangle class="h-4 w-4 text-amber-600 shrink-0" />
          <span>Terdapat <strong>{{ unconfirmedCount }}</strong> pesanan/transaksi yang <strong>belum dikonfirmasi admin nota</strong> (ditandai dengan badge kuning).</span>
        </div>
        <span class="text-[11px] font-bold text-amber-700 bg-amber-100 px-2 py-0.5 rounded-full">Data Fallback Live</span>
      </div>

      <!-- Header -->
      <div class="bg-white p-6 print:p-4 border-b border-slate-200">
        <div class="grid grid-cols-3 gap-8 items-center">
          <div class="space-y-1">
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
              SALESMAN
            </p>
            <p class="text-sm font-black text-slate-900 uppercase">
              {{ lpsData.salesman?.name || lpsData.salesman?.employee_code || '........................' }}
            </p>
          </div>
          <div class="space-y-1 text-center">
            <div class="inline-block bg-slate-100 px-5 py-1.5 rounded-full border border-slate-300 print:border-slate-900">
              <p class="text-xs font-black text-slate-800 uppercase tracking-tighter print:text-slate-900">
                {{ docBadgeLabel }}
              </p>
            </div>
          </div>
          <div class="text-right">
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
              TANGGAL LAPORAN
            </p>
            <p class="text-sm font-black text-slate-900 uppercase leading-tight mt-1">
              {{ reportDateLabel }}
            </p>
          </div>
        </div>
      </div>

      <!-- ══════════════════════════════════════════════════════════════════════
           SECTION 2: LIST KUNJUNGAN TOKO & TRANSAKSI (SPREADSHEET)
      ══════════════════════════════════════════════════════════════════════ -->
      <div class="lps-section">

        <div v-if="!lpsData.store_visits?.length" class="lps-empty-section py-8 flex flex-col items-center gap-2">
          <Store class="h-8 w-8 text-slate-300" />
          <span v-if="selectedLPSType === 'PRE_DELIVERY'" class="text-slate-500 font-medium text-sm">Tidak ada data pesanan Sales Order (SO) untuk salesman ini pada tanggal yang dipilih</span>
          <span v-else-if="selectedLPSType === 'DELIVERY'" class="text-slate-500 font-medium text-sm">Tidak ada data Surat Jalan / Pengiriman untuk salesman ini pada tanggal yang dipilih</span>
          <span v-else class="text-slate-500 font-medium text-sm">Tidak ada data faktur penjualan langsung untuk salesman ini pada tanggal yang dipilih</span>
        </div>

        <div v-else class="overflow-x-auto border border-slate-200 rounded-xl lps-table-wrapper">
          <table class="lps-table lps-spreadsheet-table min-w-max">
            <thead>
              <tr>
                <th class="lps-sticky-col left-0 z-10 w-10 text-center">No</th>
                <th class="lps-sticky-col left-[40px] z-10 w-44">Nomor Nota / Status</th>
                <th class="lps-sticky-col left-[216px] z-10 w-48">Toko / Nama</th>
                <th class="w-48">Alamat</th>
                <th class="w-32">No HP</th>
                <th class="text-right w-32">{{ selectedLPSType === 'PRE_DELIVERY' ? 'Total Pesanan' : 'Total Tagihan' }}</th>
                <th class="text-center w-20">Gln Ksg</th>
                <th class="text-center w-20">Btl Kmb</th>
                <th class="w-32">{{ selectedLPSType === 'PRE_DELIVERY' ? 'Metode Rencana' : 'Bayar Tagihan' }}</th>
                <th class="text-right w-32">{{ selectedLPSType === 'PRE_DELIVERY' ? 'Total Dibayar/DP' : 'Total Dibayar' }}</th>
                
                <!-- Dynamic Products -->
                <th v-for="prod in uniqueProducts" :key="prod" class="text-center whitespace-nowrap min-w-[80px]">
                  {{ prod }}
                </th>

                <!-- Hadiah -->
                <th class="text-center w-10 border-l border-slate-300" title="Gelas">G</th>
                <th class="text-center w-10" title="Piring">P</th>
                <th class="text-center w-10 border-r border-slate-300" title="Mangkuk">M</th>
                
                <!-- Retur -->
                <th class="text-center w-20">Retur</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(visit, vi) in lpsData.store_visits" :key="vi">
                <td class="lps-sticky-col left-0 z-10 bg-white text-center font-bold text-slate-500 border-r border-slate-100">{{ vi + 1 }}</td>
                <td class="lps-sticky-col left-[40px] z-10 bg-white text-xs font-mono text-slate-500 border-r border-slate-100 max-w-[11rem]">
                  <div class="flex flex-col gap-1">
                    <span class="font-bold text-slate-800 truncate" :title="getReceiptNumbers(visit)">{{ getReceiptNumbers(visit) }}</span>
                    <span
                      v-if="!getVisitConfirmationStatus(visit).isConfirmed"
                      class="inline-flex items-center gap-1 text-[9px] font-bold px-1.5 py-0.5 rounded bg-amber-100 text-amber-800 w-fit"
                      title="Data pesanan belum dikonfirmasi admin nota"
                    >
                      ⏳ Belum Dikonfirmasi Admin
                    </span>
                    <span
                      v-else-if="getVisitConfirmationStatus(visit).isConverted"
                      class="inline-flex items-center gap-1 text-[9px] font-bold px-1.5 py-0.5 rounded bg-teal-100 text-teal-800 w-fit"
                      title="Pesanan telah dikonversi menjadi faktur penjualan"
                    >
                      ✓ Terkonversi ke Faktur
                    </span>
                    <span
                      v-else-if="selectedLPSType === 'PRE_DELIVERY'"
                      class="inline-flex items-center gap-1 text-[9px] font-bold px-1.5 py-0.5 rounded bg-emerald-100 text-emerald-800 w-fit"
                    >
                      ✓ Terkonfirmasi Admin
                    </span>
                  </div>
                </td>
                <td class="lps-sticky-col left-[216px] z-10 bg-white font-bold text-slate-800 border-r border-slate-100 truncate max-w-[12rem]" :title="visit.store_name">{{ visit.store_name }}</td>
                <td class="text-xs text-slate-500 truncate max-w-[12rem]" :title="visit.store_address">{{ visit.store_address || '-' }}</td>
                <td class="text-xs text-slate-500">-</td>
                <td class="text-right font-black text-slate-800">{{ formatCurrency(visit.total_amount) }}</td>
                <td class="text-center text-slate-300">-</td>
                <td class="text-center text-slate-300">-</td>
                <td>
                  <div class="flex flex-col gap-1">
                    <span v-if="!getPaymentMethods(visit).length" class="text-slate-300">—</span>
                    <span v-for="(method, mi) in getPaymentMethods(visit)" :key="mi" :class="['lps-badge w-fit', paymentColor(method)]">
                      {{ paymentLabel(method) }}
                    </span>
                  </div>
                </td>
                <td class="text-right font-black text-emerald-600">{{ formatCurrency(visit.total_pembayaran) }}</td>

                <td v-for="prod in uniqueProducts" :key="prod" class="text-center font-bold" :class="getProductQtyForStore(visit, prod) > 0 ? 'text-indigo-600' : 'text-slate-200'">
                  {{ getProductQtyForStore(visit, prod) || '-' }}
                </td>

                <td class="text-center font-bold border-l border-slate-200" :class="getGiftQtyForStore(visit, 'G') > 0 ? 'text-indigo-600' : 'text-slate-200'">{{ getGiftQtyForStore(visit, 'G') || '-' }}</td>
                <td class="text-center font-bold" :class="getGiftQtyForStore(visit, 'P') > 0 ? 'text-indigo-600' : 'text-slate-200'">{{ getGiftQtyForStore(visit, 'P') || '-' }}</td>
                <td class="text-center font-bold border-r border-slate-200" :class="getGiftQtyForStore(visit, 'M') > 0 ? 'text-indigo-600' : 'text-slate-200'">{{ getGiftQtyForStore(visit, 'M') || '-' }}</td>

                <td class="text-center text-slate-300">-</td>
              </tr>
            </tbody>
            <tfoot>
              <tr class="lps-table-total bg-slate-50">
                <td colspan="5" class="lps-sticky-col left-0 z-10 bg-slate-50 font-black text-slate-800 text-right pr-4 border-r border-slate-200">TOTAL KESELURUHAN</td>
                <td class="text-right font-black text-slate-800 text-base">{{ formatCurrency((lpsData.store_visits || []).reduce((sum: number, v: any) => sum + (v.total_amount || 0), 0)) }}</td>
                <td class="text-center text-slate-300">-</td>
                <td class="text-center text-slate-300">-</td>
                <td></td>
                <td class="text-right font-black text-emerald-700 text-base">{{ formatCurrency(lpsData.total_grand_amount || 0) }}</td>
                <td v-for="prod in uniqueProducts" :key="prod" class="text-center font-black text-indigo-700">
                  {{ getTotalProductQty(prod) || '-' }}
                </td>

                <td class="text-center font-black text-indigo-700 border-l border-slate-200">{{ getTotalGiftQty('G') || '-' }}</td>
                <td class="text-center font-black text-indigo-700">{{ getTotalGiftQty('P') || '-' }}</td>
                <td class="text-center font-black text-indigo-700 border-r border-slate-200">{{ getTotalGiftQty('M') || '-' }}</td>

                <td class="text-center text-slate-300">-</td>
              </tr>
            </tfoot>
          </table>
          
          <!-- DO Summary Table Seamlessly Attached -->
          <table class="lps-table min-w-max border-t-2 border-slate-300">
            <thead>
              <tr class="bg-slate-100/50">
                <th colspan="9" class="text-right font-black text-slate-600 border-r border-slate-200 uppercase tracking-widest">{{ doSummaryTitle }}</th>
                <th class="text-center w-24">Muat</th>
                <th class="text-center w-24">Pulang</th>
                <th class="text-center w-24">Retur</th>
                <th class="text-center w-24 text-emerald-700">Terjual</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="doSummary.length === 0" class="bg-slate-50/50">
                <td colspan="13" class="text-center font-medium text-slate-400 py-3">Tidak ada rekapan barang (DO/Transfer) pada hari ini</td>
              </tr>
              <tr v-for="item in doSummary" :key="item.name" class="bg-slate-50/50 hover:bg-slate-100 transition-colors">
                <td colspan="9" class="text-right font-bold text-slate-800 border-r border-slate-200">{{ item.name }}</td>
                <td class="text-center font-bold text-slate-800">{{ item.muat }}</td>
                <td class="text-center font-bold text-orange-600">{{ item.pulang }}</td>
                <td class="text-center font-bold text-red-600">{{ item.retur }}</td>
                <td class="text-center font-black text-emerald-700">{{ item.terjual }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- ══════════════════════════════════════════════════════════════════════
           SECTION 3: GRAND TOTAL & SIGNATURE
      ══════════════════════════════════════════════════════════════════════ -->
      <div class="lps-section lps-section-final">

        <!-- Grand Total Table -->
        <table class="lps-table lps-grand-table">
          <thead>
            <tr>
              <th>Keterangan</th>
              <th class="text-center">Jumlah</th>
              <th class="text-right">Total (Rp)</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td class="font-semibold">Penerimaan Tunai (CASH)</td>
              <td class="text-center">—</td>
              <td class="text-right font-bold text-emerald-700">{{ formatCurrency(lpsData.total_cash || 0) }}</td>
            </tr>
            <tr>
              <td class="font-semibold">Penerimaan Transfer / QRIS</td>
              <td class="text-center">—</td>
              <td class="text-right font-bold text-blue-700">{{ formatCurrency(lpsData.total_transfer || 0) }}</td>
            </tr>
            <tr>
              <td class="font-semibold">Piutang Tempo</td>
              <td class="text-center">—</td>
              <td class="text-right font-bold text-amber-700">{{ formatCurrency(lpsData.total_tempo || 0) }}</td>
            </tr>
            <tr>
              <td class="font-semibold">Toko Dikunjungi</td>
              <td class="text-center font-bold text-slate-800">{{ lpsData.total_stores_visited }}</td>
              <td class="text-right text-slate-400">—</td>
            </tr>
            <tr>
              <td class="font-semibold">Total Transaksi</td>
              <td class="text-center font-bold text-slate-800">{{ lpsData.total_transactions }}</td>
              <td class="text-right text-slate-400">—</td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="lps-grand-total-row">
              <td class="font-black text-slate-900 text-base" colspan="2">TOTAL UANG YANG HARUS DISETORKAN</td>
              <td class="text-right font-black text-emerald-700 text-xl">{{ formatCurrency(lpsData.total_grand_amount || 0) }}</td>
            </tr>
          </tfoot>
        </table>


        <!-- Signature Area -->
        <div class="lps-signature-grid">
          <div class="lps-signature-box">
            <div class="lps-signature-title">Admin Gudang</div>
            <div class="lps-signature-space"></div>
            <div class="lps-signature-name">( ________________________ )</div>
            <div class="lps-signature-date">Tanggal: ________________</div>
          </div>
          <div class="lps-signature-box">
            <div class="lps-signature-title">Admin Nota</div>
            <div class="lps-signature-space"></div>
            <div class="lps-signature-name">( ________________________ )</div>
            <div class="lps-signature-date">Tanggal: ________________</div>
          </div>
          <div class="lps-signature-box">
            <div class="lps-signature-title">Admin Kasir</div>
            <div class="lps-signature-space"></div>
            <div class="lps-signature-name">( ________________________ )</div>
            <div class="lps-signature-date">Tanggal: ________________</div>
          </div>
          <div class="lps-signature-box">
            <div class="lps-signature-title">Supervisor Sales</div>
            <div class="lps-signature-space"></div>
            <div class="lps-signature-name">( ________________________ )</div>
            <div class="lps-signature-date">Tanggal: ________________</div>
          </div>
          <div class="lps-signature-box">
            <div class="lps-signature-title">Salesman</div>
            <div class="lps-signature-space"></div>
            <div class="lps-signature-name">( ________________________ )</div>
            <div class="lps-signature-date">Tanggal: ________________</div>
          </div>
        </div>

        <!-- Footer note -->
        <div class="lps-footer-note">
          Lembar ini dicetak otomatis dari Sistem Karsa HR Super App • Berlaku sebagai bukti setoran resmi apabila telah ditandatangani oleh ketiga pihak.
        </div>
      </div>

    </div>
    <!-- end lps-document (screen view) -->
  </div>
</template>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════
   SCREEN & PRINT TOGGLE
═══════════════════════════════════════════════════════════════════ */
/* On screen: show/hide handled by class bindings (lps-print-hidden / lps-preview-visible in global style) */
.no-print { display: unset; }

/* ═══════════════════════════════════════════════════════════════════
   LPS DOCUMENT LAYOUT (screen)
═══════════════════════════════════════════════════════════════════ */
.lps-document {
  background: white; border-radius: 1.5rem; border: 1px solid #e2e8f0;
  padding: 2.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 4px 20px rgba(0,0,0,0.04);
  max-width: 1200px; margin: 0 auto; font-family: 'Inter', -apple-system, sans-serif;
}
.lps-section { margin-bottom: 2rem; }
.lps-section-title {
  display: flex; align-items: center; gap: 0.625rem;
  font-size: 0.8rem; font-weight: 800; color: #1e293b;
  text-transform: uppercase; letter-spacing: 0.07em;
  padding: 0.75rem 1rem; background: #f8fafc;
  border-radius: 0.75rem; border: 1px solid #e2e8f0; margin-bottom: 1rem;
}
.lps-empty-section {
  display: flex; align-items: center; gap: 0.5rem;
  color: #94a3b8; font-size: 0.8rem; padding: 1rem 0.5rem; font-style: italic;
}
.lps-table-wrapper { scrollbar-width: thin; scrollbar-color: #cbd5e1 transparent; }
.lps-table-wrapper::-webkit-scrollbar { height: 8px; }
.lps-table-wrapper::-webkit-scrollbar-thumb { background-color: #cbd5e1; border-radius: 4px; }
.lps-table { width: 100%; border-collapse: collapse; font-size: 0.78rem; }
.lps-table th {
  background: #f1f5f9; color: #0f172a; padding: 0.5rem 0.75rem;
  font-weight: 800; text-transform: uppercase; font-size: 0.68rem;
  letter-spacing: 0.05em; text-align: left; border-bottom: 2px solid #cbd5e1;
}
.lps-table td { padding: 0.5rem 0.75rem; border-bottom: 1px solid #e2e8f0; color: #0f172a; }
.lps-sticky-col { position: sticky; }
.lps-table tbody tr:last-child td { border-bottom: none; }
.lps-table tfoot td { padding: 0.625rem 0.75rem; font-size: 0.8rem; }
.lps-table-total td { background: #f8fafc; border-top: 2px solid #94a3b8; }
.lps-badge {
  display: inline-flex; align-items: center; padding: 2px 7px;
  border-radius: 6px; border: 1px solid currentColor; font-size: 0.65rem;
  font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em;
}
.lps-grand-table { margin-bottom: 2rem; }
.lps-grand-total-row td { background: #fff; color: #0f172a !important; padding: 1rem 0.75rem; }
.lps-signature-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 1.5rem; margin-bottom: 1.5rem; }
.lps-signature-box { border: 1px solid #e2e8f0; border-radius: 1rem; padding: 1.25rem; text-align: center; }
.lps-signature-title { font-weight: 800; color: #374151; font-size: 0.85rem; margin-bottom: 0.5rem; }
.lps-signature-space { height: 72px; }
.lps-signature-name { font-size: 0.78rem; color: #374151; border-top: 1px solid #94a3b8; padding-top: 0.5rem; }
.lps-signature-date { font-size: 0.72rem; color: #94a3b8; margin-top: 0.5rem; }
.lps-footer-note { text-align: center; font-size: 0.7rem; color: #94a3b8; padding-top: 1rem; border-top: 1px dashed #e2e8f0; }
.lps-section-final { page-break-before: auto; }
</style>

<style>
/* ═══════════════════════════════════════════════════════════════════
   GLOBAL PRINT TOGGLE CLASSES (unscoped so print media can override)
═══════════════════════════════════════════════════════════════════ */
.lps-print-hidden { display: none; }
.lps-preview-visible {
  display: block;
  border: 2px dashed #6366f1;
  border-radius: 0.75rem;
  padding: 1rem;
  margin-bottom: 2rem;
  background: #ffffff;
  box-shadow: 0 4px 24px rgba(99,102,241,0.08);
}

/* ═══════════════════════════════════════════════════════════════════
   PRINT-ONLY STYLES (unscoped, black & white)
═══════════════════════════════════════════════════════════════════ */
@media print {
  @page {
    size: A3 landscape;
    margin: 8mm;
  }

  html, body {
    margin: 0 !important;
    padding: 0 !important;
    background: white !important;
    color: black !important;
  }

  /* Hide everything */
  body > * { display: none !important; }
  #app { display: block !important; }

  /* Hide all screen-only content */
  .no-print { display: none !important; }
  #surat-jalan-card { display: none !important; }

  /* Hide the lps-page wrapper border/bg, only show content */
  .lps-page { display: block !important; background: white !important; padding: 0 !important; }

  /* Hide screen document sections */
  .lps-document,
  .lps-section,
  .lps-section-final,
  .lps-section-title,
  .lps-empty-section,
  .lps-table-wrapper,
  .lps-grand-table,
  .lps-signature-grid,
  .lps-footer-note { display: none !important; }

  /* Show print area regardless of preview toggle state */
  .lps-print-hidden,
  .lps-preview-visible,
  #lps-print-view {
    display: block !important;
    width: 100% !important;
    font-family: Arial, Helvetica, sans-serif !important;
    font-size: 7.5pt !important;
    color: black !important;
    background: white !important;
    border: none !important;
    box-shadow: none !important;
    padding: 0 !important;
    margin: 0 !important;
  }

  /* Title block */
  .lpt-title-block {
    text-align: center;
    margin-bottom: 4mm;
  }
  .lpt-title {
    font-size: 10pt;
    font-weight: bold;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }
  .lpt-sales-info {
    font-size: 8pt;
    font-weight: bold;
    margin-top: 1mm;
  }

  /* Main table */
  .lpt-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 7pt;
    color: black;
    background: white;
    table-layout: auto;
  }
  .lpt-th {
    border: 1px solid black;
    padding: 1.5mm 1mm;
    text-align: center;
    font-weight: bold;
    font-size: 6.5pt;
    background: white;
    color: black;
    vertical-align: middle;
    white-space: nowrap;
  }
  .lpt-td {
    border: 1px solid black;
    padding: 1mm;
    font-size: 6.5pt;
    color: black;
    background: white;
    vertical-align: middle;
  }
  .lpt-center { text-align: center; }
  .lpt-right { text-align: right; }
  .lpt-small { font-size: 5.5pt; }

  /* Column widths */
  .lpt-no   { width: 6mm; }
  .lpt-nota { width: 18mm; }
  .lpt-toko { width: 25mm; }
  .lpt-alamat { width: 25mm; }
  .lpt-hp   { width: 18mm; }
  .lpt-tagihan { width: 22mm; text-align: right; }
  .lpt-galon  { width: 14mm; }
  .lpt-botol  { width: 14mm; }
  .lpt-bayar  { width: 14mm; }
  .lpt-prod   { width: 12mm; white-space: nowrap; }
  .lpt-gift   { width: 12mm; }
  .lpt-gift-sub { width: 9mm; font-size: 6pt; }
  .lpt-retur  { width: 12mm; }

  /* Sub header row (second header for hadiah sub-cols) */
  .lpt-subhead .lpt-th { font-size: 6pt; padding: 0.5mm 1mm; }

  /* Totals */
  .lpt-total-row .lpt-td { font-weight: bold; border-top: 2px solid black; }
  .lpt-grand-row .lpt-td { font-weight: bold; border-top: 1px solid black; }

  /* Signature */
  .lpt-sign-area {
    display: flex;
    justify-content: space-between;
    margin-top: 6mm;
    gap: 12mm;
  }
  .lpt-sign-box {
    text-align: center;
    min-width: 40mm;
  }
  .lpt-sign-label {
    font-size: 7pt;
    font-weight: bold;
    margin-bottom: 10mm;
  }
  .lpt-sign-line {
    border-top: 1px solid black;
    width: 40mm;
    margin: 0 auto;
  }
}
</style>
