<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useMarketplaceStore } from '@/stores/marketplace'
import { useMasterDataStore } from '@/stores/masterData'
import type { MarketplaceStock, CreateStockPayload, FactoryProduct } from '@/api/marketplace'
import {
  Package,
  Plus,
  Pencil,
  Trash2,
  ArrowDownCircle,
  ArrowUpCircle,
  SlidersHorizontal,
  RefreshCw,
  Search,
  X,
  CheckCircle2,
  AlertTriangle,
  CircleOff,
  ChevronDown,
  PackageCheck,
  FileSpreadsheet,
} from 'lucide-vue-next'
import ExcelJS from 'exceljs'
import { saveAs } from 'file-saver'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const marketplaceStore = useMarketplaceStore()
const masterDataStore = useMasterDataStore()

// State
const searchQuery = ref('')
const platformFilter = ref('ALL')
const showAddModal = ref(false)
const showEditModal = ref(false)
const showOpnameModal = ref(false)
const showPackModal = ref(false)
const showDeleteConfirm = ref(false)
const selectedStock = ref<MarketplaceStock | null>(null)
const isRefreshing = ref(false)

const productType = ref<'FACTORY' | 'CUSTOM'>('FACTORY')

const form = ref<CreateStockPayload>({
  product_id: null,
  product_name: '',
  sku: '',
  quantity: 0,
  min_stock: 10,
  price: 0,
  notes: '',
  image: null,
})

const imagePreview = ref<string | null>(null)

function handleImageChange(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (file) {
    form.value.image = file
    imagePreview.value = URL.createObjectURL(file)
  }
}

// Product picker state
const productSearch = ref('')
const showProductPicker = ref(false)
const selectedFactoryProduct = ref<FactoryProduct | null>(null)

const filteredFactoryProducts = computed(() => {
  const q = productSearch.value.toLowerCase()
  if (!q) return marketplaceStore.factoryProducts
  return marketplaceStore.factoryProducts.filter(
    (p) => (p.name || '').toLowerCase().includes(q) || (p.sku || '').toLowerCase().includes(q)
  )
})

function selectFactoryProduct(product: FactoryProduct) {
  selectedFactoryProduct.value = product
  form.value.product_id = product.id
  form.value.product_name = product.name
  form.value.sku = product.sku
  if (form.value.price === 0) {
    form.value.price = product.selling_price
  }
  showProductPicker.value = false
  productSearch.value = ''
}

function clearProductSelection() {
  selectedFactoryProduct.value = null
  form.value.product_id = null
  if (showAddModal.value) {
    form.value.product_name = ''
    form.value.sku = ''
  }
}

function resetToFactorySku() {
  if (selectedFactoryProduct.value) {
    form.value.sku = selectedFactoryProduct.value.sku
  }
}

function resetToFactoryName() {
  if (selectedFactoryProduct.value) {
    form.value.product_name = selectedFactoryProduct.value.name
  }
}

function getCurrentLocalDateTime(): string {
  const now = new Date()
  now.setMinutes(now.getMinutes() - now.getTimezoneOffset())
  return now.toISOString().slice(0, 16)
}

const opnameForm = ref({
  type: 'STOCK_IN' as 'STOCK_IN' | 'STOCK_OUT' | 'ADJUSTMENT',
  cartons: 0,
  pcs_per_carton: 10,
  pcs: 0,
  quantity: 0,
  reference: '',
  notes: '',
  date: '',
})

function updateOpnameQuantity() {
  const c = Number(opnameForm.value.cartons) || 0
  const ppc = Number(opnameForm.value.pcs_per_carton) || 0
  const p = Number(opnameForm.value.pcs) || 0
  
  opnameForm.value.quantity = (c * ppc) + p
}


let pollingInterval: ReturnType<typeof setInterval> | null = null

async function loadData() {
  await marketplaceStore.fetchStocks()
}

onMounted(async () => {
  await loadData()
  await marketplaceStore.fetchFactoryProducts()
  pollingInterval = setInterval(() => {
    loadData()
  }, 15000)
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
})

watch(() => masterDataStore.selectedBranchId, () => {
  loadData()
})

function getStockGroupPriority(stock: MarketplaceStock): number {
  const name = (stock.product_name || '').toLowerCase()
  const sku = (stock.sku || '').toLowerCase()

  // 1. Karsa HR Products
  // No brand-specific stock priority in the trial.

  // 2. Rajaku Kecap Products (excluding cuka)
  if ((name.includes('rajaku') && !name.includes('cuka')) || sku.startsWith('rh-') || sku.startsWith('rpg-')) return 2

  // 3. Jangkar Products (KM Jangkar 200, 250, 690, 700, 860, 6.2, etc.)
  if (name.includes('jangkar') || sku.includes('jangkar')) return 3

  // 4. Saus, Sambal, & Cuka
  if (
    name.includes('saus') ||
    name.includes('saos') ||
    name.includes('tomat') ||
    name.includes('cabe tani') ||
    name.includes('cuka') ||
    sku.includes('rajatomat') ||
    sku.includes('ct-') ||
    sku.includes('cuka')
  ) {
    return 4
  }

  // 5. Hadiah & Merchandise
  if (
    name.includes('solet') ||
    name.includes('pisau') ||
    name.includes('mangkok') ||
    name.includes('sendok') ||
    name.includes('hadiah') ||
    sku.includes('hadiah')
  ) {
    return 5
  }

  // 6. Kardus, Lakban, Bubble Wrap & Packing Materials
  if (
    name.includes('kardus') ||
    name.includes('bubble') ||
    name.includes('lakban') ||
    name.includes('isolasi') ||
    name.includes('fragile') ||
    sku.includes('stock - k') ||
    sku.includes('stock-k') ||
    sku.includes('stock-fragile') ||
    sku.includes('stock - bubble') ||
    sku.includes('stock - lakban')
  ) {
    return 6
  }

  // 7. Other custom items
  return 7
}

function sortStocks(items: MarketplaceStock[]): MarketplaceStock[] {
  return [...items].sort((a, b) => {
    const prioA = getStockGroupPriority(a)
    const prioB = getStockGroupPriority(b)
    if (prioA !== prioB) {
      return prioA - prioB
    }
    return (a.product_name || '').localeCompare(b.product_name || '', 'id', {
      numeric: true,
      sensitivity: 'base',
    })
  })
}

const filteredStocks = computed(() => {
  let result = marketplaceStore.stocks
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(
      (s) =>
        s.product_name.toLowerCase().includes(q) || s.sku.toLowerCase().includes(q),
    )
  }
  return sortStocks(result)
})

const previewQtyAfter = computed(() => {
  if (!selectedStock.value) return 0
  const qty = opnameForm.value.quantity
  const current = selectedStock.value.quantity
  if (opnameForm.value.type === 'STOCK_IN') return current + qty
  if (opnameForm.value.type === 'STOCK_OUT') return Math.max(0, current - qty)
  if (opnameForm.value.type === 'ADJUSTMENT') return qty
  return current
})



function statusConfig(stock: MarketplaceStock) {
  const s = marketplaceStore.getStockStatus(stock)
  if (s === 'safe') return { label: 'Aman', cls: 'bg-emerald-100 text-emerald-700 border-emerald-200' }
  if (s === 'critical') return { label: 'Kritis', cls: 'bg-amber-100 text-amber-700 border-amber-200' }
  return { label: 'Habis', cls: 'bg-red-100 text-red-700 border-red-200' }
}

function openAdd() {
  productType.value = 'FACTORY'
  form.value = { product_id: null, product_name: '', sku: '', quantity: 0, min_stock: 10, price: 0, notes: '', image: null }
  imagePreview.value = null
  selectedFactoryProduct.value = null
  productSearch.value = ''
  showProductPicker.value = false
  showAddModal.value = true
}

function openEdit(stock: MarketplaceStock) {
  selectedStock.value = stock
  productType.value = stock.product_id ? 'FACTORY' : 'CUSTOM'
  form.value = {
    product_id: stock.product_id || null,
    product_name: stock.product_name,
    sku: stock.sku,
    quantity: stock.quantity,
    min_stock: stock.min_stock,
    price: stock.price,
    notes: stock.notes,
    image: null,
  }
  imagePreview.value = stock.image_url || stock.product?.image_url || null
  selectedFactoryProduct.value = stock.product || (stock.product_id ? marketplaceStore.factoryProducts.find(p => p.id === stock.product_id) || null : null)
  productSearch.value = ''
  showProductPicker.value = false
  showEditModal.value = true
}

function openOpname(stock: MarketplaceStock) {
  selectedStock.value = stock
  opnameForm.value = {
    type: 'STOCK_IN',
    cartons: 0,
    pcs_per_carton: 10,
    pcs: 0,
    quantity: 0,
    reference: '',
    notes: '',
    date: getCurrentLocalDateTime(),
  }
  showOpnameModal.value = true
}


function openDelete(stock: MarketplaceStock) {
  selectedStock.value = stock
  showDeleteConfirm.value = true
}

async function submitAdd() {
  if (productType.value === 'FACTORY' && !form.value.product_id) {
    alert('Silakan pilih produk dari pabrik terlebih dahulu')
    return
  }
  if (!form.value.product_name || !form.value.product_name.trim()) {
    alert('Nama produk wajib diisi')
    return
  }
  if (!form.value.sku || !form.value.sku.trim()) {
    alert('SKU produk wajib diisi')
    return
  }
  
  try {
    await marketplaceStore.createStock(form.value)
    showAddModal.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal menambah produk')
  }
}

async function submitEdit() {
  if (!selectedStock.value) return
  if (productType.value === 'FACTORY' && !form.value.product_id) {
    alert('Silakan pilih produk dari pabrik terlebih dahulu')
    return
  }
  if (!form.value.product_name || !form.value.product_name.trim()) {
    alert('Nama produk wajib diisi')
    return
  }
  if (!form.value.sku || !form.value.sku.trim()) {
    alert('SKU produk wajib diisi')
    return
  }
  try {
    await marketplaceStore.updateStock(selectedStock.value.id, form.value)
    showEditModal.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal memperbarui produk')
  }
}

async function submitOpname() {
  if (!selectedStock.value) return
  try {
    await marketplaceStore.doOpname(selectedStock.value.id, opnameForm.value)
    showOpnameModal.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal melakukan opname')
  }
}


async function confirmDelete() {
  if (!selectedStock.value) return
  try {
    await marketplaceStore.deleteStock(selectedStock.value.id)
    showDeleteConfirm.value = false
  } catch (e: any) {
    alert(e.response?.data?.error || 'Gagal menghapus produk')
  }
}

async function manualRefresh() {
  isRefreshing.value = true
  await marketplaceStore.fetchStocks()
  setTimeout(() => (isRefreshing.value = false), 600)
}

const lastUpdatedFormatted = computed(() => {
  if (!marketplaceStore.lastUpdated) return '-'
  return marketplaceStore.lastUpdated.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
})

async function exportToExcel() {
  const workbook = new ExcelJS.Workbook()
  workbook.creator = 'HRIS Karsa HR'
  workbook.created = new Date()

  const sheet = workbook.addWorksheet('Stok Marketplace', {
    pageSetup: { fitToPage: true, orientation: 'landscape' },
    views: [{ state: 'frozen', ySplit: 3 }],
  })

  // ── TITLE ROW ────────────────────────────────────────────────────────
  sheet.mergeCells('A1:H1')
  const titleCell = sheet.getCell('A1')
  titleCell.value = `LAPORAN STOK MARKETPLACE — ${new Date().toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })}`
  titleCell.font = { bold: true, size: 14, color: { argb: 'FFFFFFFF' } }
  titleCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E3A5F' } }
  titleCell.alignment = { horizontal: 'center', vertical: 'middle' }
  sheet.getRow(1).height = 36

  // ── SUBTITLE ROW ─────────────────────────────────────────────────────
  sheet.mergeCells('A2:H2')
  const subCell = sheet.getCell('A2')
  subCell.value = `Total Produk: ${filteredStocks.value.length} SKU  •  Diekspor: ${new Date().toLocaleString('id-ID')}`
  subCell.font = { italic: true, size: 10, color: { argb: 'FF64748B' } }
  subCell.alignment = { horizontal: 'center', vertical: 'middle' }
  subCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF8FAFC' } }
  sheet.getRow(2).height = 22

  // ── COLUMN DEFINITIONS ───────────────────────────────────────────────
  sheet.columns = [
    { key: 'no',           width: 6  },
    { key: 'product_name', width: 34 },
    { key: 'sku',          width: 20 },
    { key: 'quantity',     width: 14 },
    { key: 'min_stock',    width: 12 },
    { key: 'price',        width: 18 },
    { key: 'status',       width: 12 },
    { key: 'notes',        width: 30 },
  ]

  // ── HEADER ROW ───────────────────────────────────────────────────────
  const headerLabels = ['No', 'Nama Produk', 'SKU', 'Stok Saat Ini', 'Min Stok', 'Harga (Rp)', 'Status', 'Catatan']
  const headerRow = sheet.getRow(3)
  headerLabels.forEach((label, i) => {
    const cell = headerRow.getCell(i + 1)
    cell.value = label
    cell.font = { bold: true, size: 10, color: { argb: 'FFFFFFFF' } }
    cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF334155' } }
    cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true }
    cell.border = {
      top:    { style: 'thin',   color: { argb: 'FF475569' } },
      bottom: { style: 'medium', color: { argb: 'FF0F172A' } },
      left:   { style: 'thin',   color: { argb: 'FF475569' } },
      right:  { style: 'thin',   color: { argb: 'FF475569' } },
    }
  })
  headerRow.height = 28

  // ── DATA ROWS ────────────────────────────────────────────────────────
  filteredStocks.value.forEach((stock, index) => {
    const status = marketplaceStore.getStockStatus(stock)
    const statusLabel = status === 'safe' ? 'Aman' : status === 'critical' ? 'Kritis' : 'Habis'

    const rowBgColor =
      status === 'safe'
        ? index % 2 === 0 ? 'FFF0FDF4' : 'FFD1FAE5'
        : status === 'critical'
          ? index % 2 === 0 ? 'FFFEFCE8' : 'FFFEF08A'
          : index % 2 === 0 ? 'FFFEF2F2' : 'FFFECACA'

    const statusTextColor =
      status === 'safe' ? 'FF065F46' : status === 'critical' ? 'FF713F12' : 'FF7F1D1D'

    const row = sheet.addRow({
      no:           index + 1,
      product_name: stock.product_name,
      sku:          stock.sku,
      quantity:     stock.quantity,
      min_stock:    stock.min_stock,
      price:        stock.price,
      status:       statusLabel,
      notes:        stock.notes || '',
    })

    row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
      cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: rowBgColor } }
      cell.border = {
        top:    { style: 'thin', color: { argb: 'FFE2E8F0' } },
        bottom: { style: 'thin', color: { argb: 'FFE2E8F0' } },
        left:   { style: 'thin', color: { argb: 'FFE2E8F0' } },
        right:  { style: 'thin', color: { argb: 'FFE2E8F0' } },
      }
      cell.alignment = { vertical: 'middle', wrapText: false }

      if (colNumber === 1) {
        cell.alignment = { horizontal: 'center', vertical: 'middle' }
        cell.font = { size: 9, color: { argb: 'FF64748B' } }
      } else if (colNumber === 2) {
        cell.font = { bold: true, size: 10 }
        cell.alignment = { vertical: 'middle', wrapText: true }
      } else if (colNumber === 3) {
        cell.font = { size: 9, color: { argb: 'FF475569' } }
      } else if (colNumber === 4) {
        cell.alignment = { horizontal: 'right', vertical: 'middle' }
        cell.font = { bold: true, size: 12 }
      } else if (colNumber === 5) {
        cell.alignment = { horizontal: 'right', vertical: 'middle' }
        cell.font = { size: 10, color: { argb: 'FF64748B' } }
      } else if (colNumber === 6) {
        cell.numFmt = '#,##0'
        cell.alignment = { horizontal: 'right', vertical: 'middle' }
        cell.font = { size: 10 }
      } else if (colNumber === 7) {
        cell.alignment = { horizontal: 'center', vertical: 'middle' }
        cell.font = { bold: true, color: { argb: statusTextColor } }
      } else if (colNumber === 8) {
        cell.font = { italic: true, size: 9, color: { argb: 'FF64748B' } }
        cell.alignment = { vertical: 'middle', wrapText: true }
      }
    })

    row.height = 22
  })

  // ── EMPTY STATE ──────────────────────────────────────────────────────
  if (filteredStocks.value.length === 0) {
    sheet.mergeCells('A4:H4')
    const emptyCell = sheet.getCell('A4')
    emptyCell.value = 'Tidak ada data stok ditemukan.'
    emptyCell.alignment = { horizontal: 'center', vertical: 'middle' }
    emptyCell.font = { italic: true, color: { argb: 'FF94A3B8' } }
    sheet.getRow(4).height = 28
  }

  // ── FOOTER SUMMARY ───────────────────────────────────────────────────
  const safeCount     = filteredStocks.value.filter(s => marketplaceStore.getStockStatus(s) === 'safe').length
  const criticalCount = filteredStocks.value.filter(s => marketplaceStore.getStockStatus(s) === 'critical').length
  const emptyCount    = filteredStocks.value.filter(s => marketplaceStore.getStockStatus(s) === 'empty').length

  const sepRow = sheet.addRow({})
  sheet.mergeCells(`A${sepRow.number}:H${sepRow.number}`)
  sepRow.getCell(1).fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2E8F0' } }
  sepRow.height = 4

  const footerRow = sheet.addRow({})
  sheet.mergeCells(`A${footerRow.number}:H${footerRow.number}`)
  const footerCell = footerRow.getCell(1)
  footerCell.value = `Ringkasan:   ✅ Aman: ${safeCount}   ⚠️ Kritis: ${criticalCount}   ❌ Habis: ${emptyCount}   |   Total: ${filteredStocks.value.length} SKU`
  footerCell.font = { bold: true, size: 10, color: { argb: 'FF334155' } }
  footerCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF1F5F9' } }
  footerCell.alignment = { horizontal: 'center', vertical: 'middle' }
  footerRow.height = 26

  // ── DOWNLOAD ─────────────────────────────────────────────────────────
  const buffer = await workbook.xlsx.writeBuffer()
  const blob = new Blob([buffer], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  })
  saveAs(blob, `stok-marketplace-${new Date().toISOString().slice(0, 10)}.xlsx`)
}
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-700">
    <!-- Page Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-3">
          <div class="p-2 rounded-xl bg-gradient-to-br from-primary to-primary/80 shadow-md shadow-primary/20">
            <Package class="h-6 w-6 text-white" />
          </div>
          Stok Marketplace
        </h1>
        <p class="text-slate-500 text-sm font-medium mt-1">
          Kelola stok produk di semua platform marketplace secara realtime
        </p>
      </div>
      <div class="flex items-center gap-3 flex-wrap">
        <!-- LIVE indicator -->
        <div class="flex items-center gap-2 bg-emerald-50 border border-emerald-200 px-3 py-1.5 rounded-xl text-xs font-bold text-emerald-700">
          <span class="relative flex h-2 w-2">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
          </span>
          LIVE · Diperbarui {{ lastUpdatedFormatted }}
        </div>
        <button
          @click="manualRefresh"
          class="p-2 rounded-xl border border-slate-200 text-slate-500 hover:text-primary hover:border-primary/30 transition-all shadow-sm"
        >
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': isRefreshing }" />
        </button>
        <!-- Export Excel -->
        <button
          @click="exportToExcel"
          class="flex items-center gap-2 px-4 py-2 rounded-xl border border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-100 hover:border-emerald-300 transition-all text-sm font-bold shadow-sm"
          title="Export data stok ke file Excel"
        >
          <FileSpreadsheet class="h-4 w-4" />
          Export Excel
        </button>
        <Button
          @click="openAdd"
          class="bg-gradient-to-r from-primary to-primary/80 text-white hover:from-primary/90 hover:to-primary shadow-md shadow-primary/20 rounded-xl font-bold gap-2"
        >
          <Plus class="h-4 w-4" />
          Tambah Produk
        </Button>
      </div>

    </div>

    <!-- Filters -->
    <Card class="border-none shadow-sm">
      <CardContent class="p-4 flex flex-col md:flex-row gap-3 items-center">
        <div class="relative flex-1 w-full md:max-w-xs">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari nama produk atau SKU..."
            class="w-full pl-9 pr-4 py-2.5 text-sm font-medium bg-slate-50 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/50 transition-all"
          />
        </div>
        <div class="flex items-center gap-2 flex-wrap">
        </div>
        <div class="md:ml-auto text-xs font-bold text-slate-400">
          {{ filteredStocks.length }} produk
        </div>
      </CardContent>
    </Card>

    <!-- Stock Table -->
    <Card class="border-none shadow-sm overflow-hidden">
      <div v-if="marketplaceStore.isLoading && filteredStocks.length === 0" class="p-16 text-center">
        <div class="animate-spin h-8 w-8 border-2 border-primary border-t-transparent rounded-full mx-auto mb-3"></div>
        <p class="text-slate-400 text-sm font-medium">Memuat data stok...</p>
      </div>
      <div v-else-if="filteredStocks.length === 0" class="p-16 text-center">
        <Package class="h-12 w-12 text-slate-200 mx-auto mb-3" />
        <p class="text-slate-400 font-bold">Belum ada produk</p>
        <p class="text-slate-400 text-xs mt-1">Klik "Tambah Produk" untuk mulai</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-100">
              <th class="text-left px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Produk</th>
              <th class="text-right px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Stok</th>
              <th class="text-right px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Min Stok</th>
              <th class="text-center px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Status</th>
              <th class="text-center px-4 py-3.5 text-[11px] font-black uppercase tracking-widest text-slate-400">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr
              v-for="stock in filteredStocks"
              :key="stock.id"
              class="hover:bg-primary/5/30 transition-colors group"
            >
              <!-- Product -->
              <td class="px-4 py-3.5">
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-lg overflow-hidden bg-slate-100 flex-shrink-0 border border-slate-200 flex items-center justify-center">
                    <img v-if="stock.image_url || stock.product?.image_url" :src="stock.image_url || stock.product?.image_url" class="h-full w-full object-cover" :alt="stock.product_name" />
                    <Package v-else class="h-5 w-5 text-slate-400" />
                  </div>
                  <div>
                    <p class="font-bold text-slate-900 text-sm">{{ stock.product_name }}</p>
                    <p class="text-xs text-slate-400 font-medium mt-0.5">SKU: {{ stock.sku }}</p>
                  </div>
                </div>
              </td>
              <!-- Quantity -->
              <td class="px-4 py-3.5">
                <div class="flex flex-col gap-1 w-44 ml-auto">
                  <div class="flex items-center justify-between font-black text-xl px-1"
                    :class="marketplaceStore.getStockStatus(stock) === 'safe' ? 'text-slate-900' : marketplaceStore.getStockStatus(stock) === 'critical' ? 'text-amber-600' : 'text-red-600'">
                    <span>{{ stock.quantity.toLocaleString('id-ID') }}</span>
                    <span class="text-[10px] text-slate-400 font-bold">TOTAL</span>
                  </div>
                </div>
              </td>
              <!-- Min Stock -->
              <td class="px-4 py-3.5 text-right">
                <span class="text-sm font-bold text-slate-500">{{ stock.min_stock.toLocaleString('id-ID') }}</span>
              </td>
              <!-- Status Badge -->
              <td class="px-4 py-3.5 text-center">
                <span
                  class="text-xs font-black px-2.5 py-1 rounded-full border"
                  :class="statusConfig(stock).cls"
                >
                  <CheckCircle2 v-if="marketplaceStore.getStockStatus(stock) === 'safe'" class="inline h-3 w-3 mr-1" />
                  <AlertTriangle v-else-if="marketplaceStore.getStockStatus(stock) === 'critical'" class="inline h-3 w-3 mr-1" />
                  <CircleOff v-else class="inline h-3 w-3 mr-1" />
                  {{ statusConfig(stock).label }}
                </span>
              </td>
              <!-- Actions -->
              <td class="px-4 py-3.5">
                <div class="flex items-center justify-center gap-1.5">
                  <button
                    @click="openOpname(stock)"
                    class="p-2 rounded-lg bg-primary/5 text-primary hover:bg-primary/10 transition-colors"
                    title="Opname Stok"
                  >
                    <SlidersHorizontal class="h-3.5 w-3.5" />
                  </button>
                  <button
                    @click="openEdit(stock)"
                    class="p-2 rounded-lg bg-blue-50 text-blue-600 hover:bg-blue-100 transition-colors"
                    title="Edit Produk"
                  >
                    <Pencil class="h-3.5 w-3.5" />
                  </button>
                  <button
                    @click="openDelete(stock)"
                    class="p-2 rounded-lg bg-red-50 text-red-500 hover:bg-red-100 transition-colors"
                    title="Hapus Produk"
                  >
                    <Trash2 class="h-3.5 w-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </Card>

    <!-- ===== ADD MODAL ===== -->
    <Teleport to="body">
      <div v-if="showAddModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showAddModal = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-lg overflow-hidden animate-in zoom-in-95 duration-200">
          <div class="bg-gradient-to-r from-primary to-primary/80 p-5 text-white">
            <h2 class="text-lg font-black flex items-center gap-2"><Plus class="h-5 w-5" /> Tambah Produk Marketplace</h2>
          </div>
          <form @submit.prevent="submitAdd" class="p-6 space-y-4">
            <!-- Product Type Toggle -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Jenis Produk</label>
              <div class="grid grid-cols-2 gap-2 bg-slate-100 p-1 rounded-xl">
                <button 
                  type="button" 
                  @click="productType = 'FACTORY'; clearProductSelection()"
                  class="py-2 text-xs font-bold rounded-lg transition-all"
                  :class="productType === 'FACTORY' ? 'bg-white shadow-sm text-primary' : 'text-slate-500 hover:text-slate-700'"
                >
                  Produk Pabrik
                </button>
                <button 
                  type="button" 
                  @click="productType = 'CUSTOM'; clearProductSelection()"
                  class="py-2 text-xs font-bold rounded-lg transition-all"
                  :class="productType === 'CUSTOM' ? 'bg-white shadow-sm text-primary' : 'text-slate-500 hover:text-slate-700'"
                >
                  Produk Custom / Bundle
                </button>
              </div>
            </div>

            <!-- Factory Product Picker -->
            <div v-if="productType === 'FACTORY'" class="space-y-3">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Produk Pabrik *</label>
                
                <!-- Selected product display -->
                <div v-if="selectedFactoryProduct" class="flex items-center justify-between gap-3 p-3 bg-primary/5 border-2 border-primary/30 rounded-xl">
                  <div>
                    <p class="font-black text-sm text-primary">{{ selectedFactoryProduct.name }}</p>
                    <p class="text-xs text-primary font-bold">SKU Asli: {{ selectedFactoryProduct.sku }} · Harga: Rp {{ selectedFactoryProduct.selling_price.toLocaleString('id-ID') }}</p>
                  </div>
                  <button type="button" @click="clearProductSelection" class="p-1.5 text-primary/70 hover:text-primary rounded-lg hover:bg-primary/10" title="Ganti Produk Pabrik">
                    <X class="h-4 w-4" />
                  </button>
                </div>

                <!-- Picker toggle -->
                <button
                  v-else
                  type="button"
                  @click="showProductPicker = !showProductPicker"
                  class="w-full flex items-center justify-between px-4 py-3 text-sm font-bold border-2 border-dashed border-slate-300 rounded-xl hover:border-primary/50 hover:bg-primary/5 transition-all text-slate-500 hover:text-primary"
                >
                  <span class="flex items-center gap-2"><Package class="h-4 w-4" /> Pilih produk dari pabrik...</span>
                  <ChevronDown class="h-4 w-4" :class="{ 'rotate-180': showProductPicker }" />
                </button>

                <!-- Searchable dropdown -->
                <div v-if="showProductPicker && !selectedFactoryProduct" class="mt-2 border border-slate-200 rounded-2xl shadow-lg overflow-hidden">
                  <div class="p-3 border-b border-slate-100">
                    <div class="relative">
                      <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
                      <input v-model="productSearch" type="text" placeholder="Cari nama produk / SKU..."
                        class="w-full pl-9 pr-4 py-2 text-sm font-medium bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/20" />
                    </div>
                  </div>
                  <div class="max-h-48 overflow-y-auto">
                    <div v-if="filteredFactoryProducts.length === 0" class="p-6 text-center text-slate-400 text-sm font-bold">
                      Produk tidak ditemukan
                    </div>
                    <button
                      v-for="product in filteredFactoryProducts"
                      :key="product.id"
                      type="button"
                      @click="selectFactoryProduct(product)"
                      class="w-full flex items-center justify-between px-4 py-3 hover:bg-primary/5 transition-colors border-b border-slate-50 last:border-0"
                    >
                      <div class="text-left">
                        <p class="font-bold text-sm text-slate-900">{{ product.name }}</p>
                        <p class="text-xs text-slate-400">SKU: {{ product.sku }} · {{ product.category }}</p>
                      </div>
                      <span class="text-xs font-black text-emerald-600">Rp {{ product.selling_price.toLocaleString('id-ID') }}</span>
                    </button>
                  </div>
                </div>
              </div>

              <!-- Editable SKU & Name for selected factory product in Add modal -->
              <div v-if="selectedFactoryProduct" class="space-y-3 pt-1">
                <div>
                  <div class="flex items-center justify-between mb-1.5">
                    <label class="block text-xs font-black uppercase tracking-widest text-slate-500">Nama Produk (Marketplace) *</label>
                    <button 
                      v-if="form.product_name !== selectedFactoryProduct.name" 
                      type="button" 
                      @click="resetToFactoryName" 
                      class="text-[11px] font-bold text-primary hover:underline"
                    >
                      Reset ke Nama Pabrik
                    </button>
                  </div>
                  <input v-model="form.product_name" required type="text" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" />
                </div>
                <div>
                  <div class="flex items-center justify-between mb-1.5">
                    <label class="block text-xs font-black uppercase tracking-widest text-slate-500">SKU (Marketplace) *</label>
                    <button 
                      v-if="form.sku !== selectedFactoryProduct.sku" 
                      type="button" 
                      @click="resetToFactorySku" 
                      class="text-[11px] font-bold text-primary hover:underline"
                    >
                      Gunakan SKU Pabrik ({{ selectedFactoryProduct.sku }})
                    </button>
                  </div>
                  <input v-model="form.sku" required type="text" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-mono font-bold" />
                </div>
              </div>
            </div>

            <!-- Custom Product Inputs -->
            <div v-if="productType === 'CUSTOM'" class="space-y-4">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Nama Produk Bundle / Custom *</label>
                <input v-model="form.product_name" required type="text" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" placeholder="Contoh: Paket Bundling Sabun" />
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">SKU *</label>
                <input v-model="form.sku" required type="text" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-mono font-bold" placeholder="Contoh: BNDL-SBN-01" />
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Foto Produk (Opsional)</label>
                <div class="flex items-center gap-3">
                  <div v-if="imagePreview" class="h-12 w-12 rounded-lg overflow-hidden border border-slate-200 shrink-0">
                    <img :src="imagePreview" class="h-full w-full object-cover" />
                  </div>
                  <input type="file" accept="image/*" @change="handleImageChange" class="w-full text-sm text-slate-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-bold file:bg-primary/5 file:text-primary hover:file:bg-primary/10 transition-colors" />
                </div>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Stok Saat Ini</label>
                <input v-model.number="form.quantity" type="number" min="0" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" />
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Min Stok Alert</label>
                <input v-model.number="form.min_stock" type="number" min="0" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" />
              </div>
              <div class="col-span-2">
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Harga</label>
                <input v-model.number="form.price" type="number" min="0" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" />
              </div>
              <div class="col-span-2">
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Catatan</label>
                <textarea v-model="form.notes" rows="2" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium resize-none" placeholder="Opsional..." />
              </div>
            </div>
            <div class="flex gap-3 pt-2">
              <button type="button" @click="showAddModal = false" class="flex-1 px-4 py-2.5 text-sm font-bold border border-slate-200 rounded-xl hover:bg-slate-50 transition-colors">Batal</button>
              <button type="submit" class="flex-1 px-4 py-2.5 text-sm font-bold bg-gradient-to-r from-primary to-primary/80 text-white rounded-xl hover:from-primary/90 hover:to-primary transition-all shadow-md">Simpan Produk</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- ===== EDIT MODAL ===== -->
    <Teleport to="body">
      <div v-if="showEditModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showEditModal = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-lg overflow-hidden animate-in zoom-in-95 duration-200">
          <div class="bg-gradient-to-r from-primary to-primary/80 p-5 text-white">
            <h2 class="text-lg font-black flex items-center gap-2"><Pencil class="h-5 w-5" /> Edit Produk</h2>
            <p class="text-blue-200 text-sm font-medium mt-0.5">{{ selectedStock?.product_name }}</p>
          </div>
          <form @submit.prevent="submitEdit" class="p-6 space-y-4">
            <!-- Product Type Toggle -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Jenis Produk</label>
              <div class="grid grid-cols-2 gap-2 bg-slate-100 p-1 rounded-xl">
                <button 
                  type="button" 
                  @click="productType = 'FACTORY'; clearProductSelection()"
                  class="py-2 text-xs font-bold rounded-lg transition-all"
                  :class="productType === 'FACTORY' ? 'bg-white shadow-sm text-primary' : 'text-slate-500 hover:text-slate-700'"
                >
                  Produk Pabrik
                </button>
                <button 
                  type="button" 
                  @click="productType = 'CUSTOM'; clearProductSelection()"
                  class="py-2 text-xs font-bold rounded-lg transition-all"
                  :class="productType === 'CUSTOM' ? 'bg-white shadow-sm text-primary' : 'text-slate-500 hover:text-slate-700'"
                >
                  Produk Custom / Bundle
                </button>
              </div>
            </div>

            <!-- Factory Product Picker -->
            <div v-if="productType === 'FACTORY'" class="space-y-3">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Hubungan Produk Pabrik</label>
                
                <!-- Selected product display -->
                <div v-if="selectedFactoryProduct" class="flex items-center justify-between gap-3 p-3 bg-primary/5 border-2 border-primary/30 rounded-xl">
                  <div>
                    <p class="font-black text-sm text-primary">{{ selectedFactoryProduct.name }}</p>
                    <p class="text-xs text-primary font-bold">SKU Pabrik: {{ selectedFactoryProduct.sku }} · Harga: Rp {{ selectedFactoryProduct.selling_price.toLocaleString('id-ID') }}</p>
                  </div>
                  <button type="button" @click="clearProductSelection" class="p-1.5 text-primary/70 hover:text-primary rounded-lg hover:bg-primary/10" title="Lepas Hubungan Produk Pabrik">
                    <X class="h-4 w-4" />
                  </button>
                </div>

                <!-- Picker toggle -->
                <button
                  v-else
                  type="button"
                  @click="showProductPicker = !showProductPicker"
                  class="w-full flex items-center justify-between px-4 py-3 text-sm font-bold border-2 border-dashed border-slate-300 rounded-xl hover:border-primary/50 hover:bg-primary/5 transition-all text-slate-500 hover:text-primary"
                >
                  <span class="flex items-center gap-2"><Package class="h-4 w-4" /> Hubungkan ke produk dari pabrik...</span>
                  <ChevronDown class="h-4 w-4" :class="{ 'rotate-180': showProductPicker }" />
                </button>

                <!-- Searchable dropdown -->
                <div v-if="showProductPicker && !selectedFactoryProduct" class="mt-2 border border-slate-200 rounded-2xl shadow-lg overflow-hidden">
                  <div class="p-3 border-b border-slate-100">
                    <div class="relative">
                      <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
                      <input v-model="productSearch" type="text" placeholder="Cari nama produk / SKU..."
                        class="w-full pl-9 pr-4 py-2 text-sm font-medium bg-slate-50 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/20" />
                    </div>
                  </div>
                  <div class="max-h-48 overflow-y-auto">
                    <div v-if="filteredFactoryProducts.length === 0" class="p-6 text-center text-slate-400 text-sm font-bold">
                      Produk tidak ditemukan
                    </div>
                    <button
                      v-for="product in filteredFactoryProducts"
                      :key="product.id"
                      type="button"
                      @click="selectFactoryProduct(product)"
                      class="w-full flex items-center justify-between px-4 py-3 hover:bg-primary/5 transition-colors border-b border-slate-50 last:border-0"
                    >
                      <div class="text-left">
                        <p class="font-bold text-sm text-slate-900">{{ product.name }}</p>
                        <p class="text-xs text-slate-400">SKU: {{ product.sku }} · {{ product.category }}</p>
                      </div>
                      <span class="text-xs font-black text-emerald-600">Rp {{ product.selling_price.toLocaleString('id-ID') }}</span>
                    </button>
                  </div>
                </div>
              </div>

              <!-- Editable SKU & Name for Factory Product in Edit Modal -->
              <div class="space-y-3 pt-1">
                <div>
                  <div class="flex items-center justify-between mb-1.5">
                    <label class="block text-xs font-black uppercase tracking-widest text-slate-500">Nama Produk (Marketplace) *</label>
                    <button 
                      v-if="selectedFactoryProduct && form.product_name !== selectedFactoryProduct.name" 
                      type="button" 
                      @click="resetToFactoryName" 
                      class="text-[11px] font-bold text-primary hover:underline"
                    >
                      Reset ke Nama Pabrik
                    </button>
                  </div>
                  <input v-model="form.product_name" required type="text" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" />
                </div>
                <div>
                  <div class="flex items-center justify-between mb-1.5">
                    <label class="block text-xs font-black uppercase tracking-widest text-slate-500">SKU (Marketplace) *</label>
                    <button 
                      v-if="selectedFactoryProduct && form.sku !== selectedFactoryProduct.sku" 
                      type="button" 
                      @click="resetToFactorySku" 
                      class="text-[11px] font-bold text-primary hover:underline"
                    >
                      Gunakan SKU Pabrik ({{ selectedFactoryProduct.sku }})
                    </button>
                  </div>
                  <input v-model="form.sku" required type="text" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-mono font-bold" />
                </div>
              </div>
            </div>

            <!-- Custom Product Inputs -->
            <div v-if="productType === 'CUSTOM'" class="space-y-4">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Nama Produk Bundle / Custom *</label>
                <input v-model="form.product_name" required type="text" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" placeholder="Contoh: Paket Bundling Sabun" />
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">SKU *</label>
                <input v-model="form.sku" required type="text" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-mono font-bold" placeholder="Contoh: BNDL-SBN-01" />
              </div>
            </div>

            <!-- Foto Produk -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Foto Produk (Opsional)</label>
              <div class="flex items-center gap-3">
                <div v-if="imagePreview" class="relative h-12 w-12 rounded-lg overflow-hidden border border-slate-200 shrink-0">
                  <img :src="imagePreview" class="h-full w-full object-cover" />
                  <button type="button" @click="form.image = null; imagePreview = null" class="absolute -top-1 -right-1 bg-red-500 text-white rounded-full p-0.5 shadow-sm hover:bg-red-600">
                    <X class="w-3 h-3" />
                  </button>
                </div>
                <input type="file" accept="image/*" @change="handleImageChange" class="w-full text-sm text-slate-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-bold file:bg-primary/5 file:text-primary hover:file:bg-primary/10 transition-colors" />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Min Stok Alert</label>
                <input v-model.number="form.min_stock" type="number" min="0" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" />
              </div>
              <div>
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Harga</label>
                <input v-model.number="form.price" type="number" min="0" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium" />
              </div>
              <div class="col-span-2">
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Catatan</label>
                <textarea v-model="form.notes" rows="2" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 font-medium resize-none" placeholder="Opsional..." />
              </div>
            </div>
            <div class="flex gap-3 pt-2">
              <button type="button" @click="showEditModal = false" class="flex-1 px-4 py-2.5 text-sm font-bold border border-slate-200 rounded-xl hover:bg-slate-50 transition-colors">Batal</button>
              <button type="submit" class="flex-1 px-4 py-2.5 text-sm font-bold bg-gradient-to-r from-primary to-primary/80 text-white rounded-xl hover:from-primary/90 hover:to-primary transition-all shadow-md">Simpan Perubahan</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- ===== OPNAME MODAL ===== -->
    <Teleport to="body">
      <div v-if="showOpnameModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showOpnameModal = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-md overflow-hidden animate-in zoom-in-95 duration-200">
          <div class="bg-gradient-to-r from-red-600 to-red-700 p-5 text-white">
            <h2 class="text-lg font-black flex items-center gap-2">
              <SlidersHorizontal class="h-5 w-5" /> Opname Stok
            </h2>
            <p class="text-red-200 text-sm font-medium mt-0.5">{{ selectedStock?.product_name }}</p>
          </div>
          <form @submit.prevent="submitOpname" class="p-6 space-y-5">
            <!-- Type selector -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-2.5">Jenis Opname</label>
              <div class="grid grid-cols-3 gap-2">
                <button
                  v-for="t in [{ value: 'STOCK_IN', label: 'Masuk', color: 'emerald' }, { value: 'STOCK_OUT', label: 'Keluar', color: 'red' }, { value: 'ADJUSTMENT', label: 'Sesuaikan', color: 'blue' }]"
                  :key="t.value"
                  type="button"
                  @click="opnameForm.type = t.value as any"
                  class="py-2.5 rounded-xl text-xs font-black border-2 transition-all"
                  :class="opnameForm.type === t.value
                    ? t.color === 'emerald' ? 'bg-emerald-500 text-white border-emerald-500 shadow-md'
                      : t.color === 'red' ? 'bg-red-500 text-white border-red-500 shadow-md'
                      : 'bg-blue-500 text-white border-blue-500 shadow-md'
                    : 'bg-slate-50 text-slate-600 border-slate-200 hover:border-slate-300'"
                >
                  {{ t.label }}
                </button>
              </div>
            </div>

            <!-- Tanggal & Waktu Opname -->
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <label class="block text-xs font-black uppercase tracking-widest text-slate-500">Tanggal & Waktu Opname *</label>
                <button
                  type="button"
                  @click="opnameForm.date = getCurrentLocalDateTime()"
                  class="text-[11px] font-bold text-red-600 hover:underline"
                >
                  Set Waktu Sekarang
                </button>
              </div>
              <input
                v-model="opnameForm.date"
                type="datetime-local"
                required
                class="w-full px-3 py-2 text-sm font-bold border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-red-300 text-slate-800"
              />
              <p class="text-[11px] text-slate-400 mt-1">
                Pilih tanggal & waktu opname fisik. Riwayat pergerakan stok akan tercatat sesuai tanggal ini.
              </p>
            </div>

            <!-- Dynamic Quantity -->
            <div class="space-y-3">
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500">
                {{ opnameForm.type === 'ADJUSTMENT' ? 'Stok Baru' : 'Jumlah' }}
              </label>
              <div class="grid grid-cols-3 gap-3">
                <div>
                  <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1">Jml Dus</label>
                  <input
                    v-model.number="opnameForm.cartons"
                    @input="updateOpnameQuantity"
                    type="number"
                    min="0"
                    class="w-full px-2 py-2 text-sm font-black border-2 border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-red-300 focus:border-red-400 text-center"
                  />
                </div>
                <div>
                  <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1">Isi per Dus</label>
                  <input
                    v-model.number="opnameForm.pcs_per_carton"
                    @input="updateOpnameQuantity"
                    type="number"
                    min="1"
                    class="w-full px-2 py-2 text-sm font-black border-2 border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-red-300 focus:border-red-400 text-center"
                  />
                </div>
                <div>
                  <label class="block text-[10px] font-black uppercase tracking-widest text-slate-400 mb-1">Sisa Pcs</label>
                  <input
                    v-model.number="opnameForm.pcs"
                    @input="updateOpnameQuantity"
                    type="number"
                    min="0"
                    class="w-full px-2 py-2 text-sm font-black border-2 border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-red-300 focus:border-red-400 text-center"
                  />
                </div>
              </div>
              <div class="flex items-center justify-between p-3 bg-red-50 rounded-xl border border-red-100 mt-2">
                <span class="text-xs font-black text-red-700 uppercase tracking-widest">Total Aktual:</span>
                <span class="text-xl font-black text-red-900">{{ opnameForm.quantity }} Pcs</span>
              </div>
            </div>

            <!-- Preview -->
            <div class="bg-slate-50 rounded-2xl p-4 flex items-center justify-between">
              <div class="text-center">
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">Sebelum</p>
                <p class="text-2xl font-black text-slate-700 mt-1">{{ selectedStock?.quantity }}</p>
              </div>
              <div class="flex-1 flex items-center justify-center">
                <div class="w-12 h-0.5 bg-slate-300"></div>
                <span class="mx-2 text-lg"
                  :class="opnameForm.type === 'STOCK_IN' ? '→' : opnameForm.type === 'STOCK_OUT' ? '→' : '≈'">
                  {{ opnameForm.type === 'STOCK_IN' ? '➕' : opnameForm.type === 'STOCK_OUT' ? '➖' : '✏️' }}
                </span>
                <div class="w-12 h-0.5 bg-slate-300"></div>
              </div>
              <div class="text-center">
                <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">Sesudah</p>
                <p
                  class="text-2xl font-black mt-1"
                  :class="previewQtyAfter < (selectedStock?.quantity ?? 0) ? 'text-red-600' : previewQtyAfter > (selectedStock?.quantity ?? 0) ? 'text-emerald-600' : 'text-blue-600'"
                >{{ previewQtyAfter }}</p>
              </div>
            </div>

            <!-- Reference -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">No. Referensi</label>
              <input v-model="opnameForm.reference" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-red-300 font-medium" placeholder="No. order / no. retur / dll." />
            </div>

            <!-- Notes -->
            <div>
              <label class="block text-xs font-black uppercase tracking-widest text-slate-500 mb-1.5">Catatan</label>
              <textarea v-model="opnameForm.notes" rows="2" class="w-full px-3 py-2.5 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-red-300 font-medium resize-none" placeholder="Keterangan opsional..." />
            </div>

            <div class="flex gap-3">
              <button type="button" @click="showOpnameModal = false" class="flex-1 px-4 py-2.5 text-sm font-bold border border-slate-200 rounded-xl hover:bg-slate-50 transition-colors">Batal</button>
              <button
                type="submit"
                :disabled="marketplaceStore.isOpnameLoading"
                class="flex-1 px-4 py-2.5 text-sm font-bold bg-gradient-to-r from-red-600 to-red-700 text-white rounded-xl hover:from-red-700 hover:to-red-800 transition-all shadow-md disabled:opacity-60"
              >
                {{ marketplaceStore.isOpnameLoading ? 'Menyimpan...' : 'Konfirmasi Opname' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- ===== DELETE CONFIRM ===== -->
    <Teleport to="body">
      <div v-if="showDeleteConfirm" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm" @click="showDeleteConfirm = false" />
        <div class="relative bg-white rounded-3xl shadow-2xl w-full max-w-sm overflow-hidden animate-in zoom-in-95 duration-200 p-6 text-center">
          <div class="p-4 bg-red-100 rounded-full w-16 h-16 flex items-center justify-center mx-auto mb-4">
            <Trash2 class="h-8 w-8 text-red-500" />
          </div>
          <h3 class="text-lg font-black text-slate-900">Hapus Produk?</h3>
          <p class="text-sm text-slate-500 font-medium mt-2 mb-6">
            Produk <strong>"{{ selectedStock?.product_name }}"</strong> akan dinonaktifkan dan tidak akan tampil di daftar stok.
          </p>
          <div class="flex gap-3">
            <button @click="showDeleteConfirm = false" class="flex-1 px-4 py-2.5 text-sm font-bold border border-slate-200 rounded-xl hover:bg-slate-50 transition-colors">Batal</button>
            <button @click="confirmDelete" class="flex-1 px-4 py-2.5 text-sm font-bold bg-red-500 text-white rounded-xl hover:bg-red-600 transition-colors shadow-md">Ya, Hapus</button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
