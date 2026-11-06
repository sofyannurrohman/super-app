<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  FileText,
  Search,
  Truck,
  TrendingUp,
  Upload,
  Calendar,
  AlertTriangle,
  CheckCircle2,
  Clock,
  ExternalLink,
  DollarSign,
  Plus,
  Pencil,
  Trash2,
  X,
  Camera,
  Eye,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { filterSellableProducts } from '@/utils/productUnits'
import { useMasterDataStore } from '@/stores/masterData'

// Search & filter
const searchQuery = ref('')
const selectedStatus = ref<'ALL' | 'PENDING' | 'APPROVED' | 'WAREHOUSE_RECEIVED' | 'LOADING' | 'IN_DELIVERY' | 'DELIVERED_AND_UPLOADED' | 'SETTLED'>('ALL')
const formatStatus = (status: string) => {
  const map: Record<string, string> = {
    PENDING: 'Pending',
    APPROVED: 'Approved Pajak',
    WAREHOUSE_RECEIVED: 'SJ Masuk Gudang',
    LOADING: 'Diproses Muat',
    IN_DELIVERY: 'Sedang Dikirim',
    DELIVERED_AND_UPLOADED: 'Selesai & Uploaded',
    SETTLED: 'Settlement Selesai'
  }
  return map[status] || status
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'SETTLED': return 'bg-emerald-50 text-emerald-600 border border-emerald-100'
    case 'DELIVERED_AND_UPLOADED': return 'bg-teal-50 text-teal-600 border border-teal-100'
    case 'IN_DELIVERY': return 'bg-blue-50 text-blue-600 border border-blue-100'
    case 'LOADING': return 'bg-indigo-50 text-indigo-600 border border-indigo-100'
    case 'WAREHOUSE_RECEIVED': return 'bg-purple-50 text-purple-600 border border-purple-100'
    case 'APPROVED': return 'bg-cyan-50 text-cyan-600 border border-cyan-100'
    case 'PENDING': return 'bg-amber-50 text-amber-600 border border-amber-100'
    default: return 'bg-slate-50 text-slate-600 border border-slate-100'
  }
}

const masterStore = useMasterDataStore()

const salesMTs = computed(() => {
  return masterStore.employees.filter((emp: any) => {
    const isActive = emp.user?.is_active !== false && emp.employment_status !== 'RESIGNED' && emp.employment_status !== 'TERMINATED'
    const jobTitle = (emp.job_position?.title || '').toLowerCase()
    return isActive && /\b(mt|modern trade|kam|key account)\b/i.test(jobTitle)
  })
})

const branchStock = ref<any[]>([])

const getProductStock = (productId: string) => {
  const stock = branchStock.value.find((s: any) => s.product_id === productId)
  return stock ? stock.quantity : 0
}

// Real PO Database
const purchaseOrders = ref<any[]>([])
const keyAccounts = ref<any[]>([])

// Modal form state
const showPOModal = ref(false)
const showAdvancedOptions = ref(false)
const isEditing = ref(false)
const poForm = ref({
  id: '',
  key_account_id: '',
  employee_id: '',
  po_number: '',
  order_date: '',
  delivery_deadline: '',
  total_amount: 0,
  items_ordered: 0,
  items_shipped: 0,
  unit_name: 'Kardus',
  pcs_per_unit: 1,
  factory_price: 0,
  status: 'PENDING',
  dc_destination: '',
  po_schedule: '',
  invoice_photo_url: '',
  receipt_photo_url: '',
  billing_schedule: '',
  return_schedule: '',
  event_schedule: '',
  spg_schedule: '',
  stock_update_log: '',
  promo_update_log: '',
  supermarket_name: '',
  md_name: '',
})
const poItems = ref<any[]>([])

// Auto-fill calculate total amount based on items_ordered and factory_price
watch([() => poForm.value.items_ordered, () => poForm.value.factory_price], ([qty, price]) => {
  if (qty !== undefined && price !== undefined) {
    poForm.value.total_amount = qty * price
  }
})

// Fetch data from backend
const products = ref<any[]>([])
const fetchProducts = async () => {
  try {
    const res = await apiClient.get('/factory/products')
    const rawList = Array.isArray(res.data) ? res.data : (res.data.data || [])
    products.value = filterSellableProducts(rawList)
  } catch (err: any) {
    console.error('Failed to fetch products', err)
  }
}

const fetchPurchaseOrders = async () => {
  try {
    const res = await apiClient.get('/modern-trade/purchase-orders')
    purchaseOrders.value = (res.data.data || []).map((po: any) => ({
      ...po,
      key_account: po.key_account?.name || 'Ritel Modern',
      order_date: po.order_date ? po.order_date.split('T')[0] : '',
      delivery_deadline: po.delivery_deadline ? po.delivery_deadline.split('T')[0] : '',
    }))
  } catch (err: any) {
    toast.error('Gagal mengambil data Purchase Order: ' + (err.response?.data?.error || err.message))
  }
}

const fetchKeyAccounts = async () => {
  try {
    const res = await apiClient.get('/modern-trade/key-accounts')
    keyAccounts.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil data Key Account: ' + (err.response?.data?.error || err.message))
  }
}

onMounted(() => {
  fetchPurchaseOrders()
  fetchKeyAccounts()
  fetchProducts()
  masterStore.fetchEmployees()
  masterStore.fetchBranches()
})

const addPOItem = () => {
  poItems.value.push({ product_id: '', quantity: 1, unit_price: 0 })
}

const removePOItem = (index: number) => {
  poItems.value.splice(index, 1)
}

const onProductSelect = (item: any) => {
  const product = products.value.find(p => p.id === item.product_id)
  if (product) {
    item.unit_price = product.price || 0
  }
}

watch(poItems, (newItems) => {
  let totalQty = 0
  let totalAmt = 0
  newItems.forEach(item => {
    totalQty += item.quantity || 0
    totalAmt += (item.quantity || 0) * (item.unit_price || 0)
  })
  poForm.value.items_ordered = totalQty
  poForm.value.total_amount = totalAmt
}, { deep: true })

watch(() => poForm.value.employee_id, async (newEmpId) => {
  if (!newEmpId) {
    branchStock.value = []
    return
  }
  const emp = masterStore.employees.find((e: any) => e.id === newEmpId)
  if (emp && emp.branch_id) {
    try {
      const res = await apiClient.get('/warehouse/stock', { params: { branch_id: emp.branch_id } })
      branchStock.value = res.data || []
    } catch (e) {
      console.error('Failed to fetch branch stock', e)
      branchStock.value = []
    }
  } else {
    branchStock.value = []
  }
})

const openNewPOModal = () => {
  isEditing.value = false
  const today = new Date().toISOString().split('T')[0] || ''
  const nextWeek = new Date(new Date().setDate(new Date().getDate() + 5)).toISOString().split('T')[0] || ''
  poForm.value = {
    id: '',
    key_account_id: keyAccounts.value[0]?.id || '',
    employee_id: '',
    po_number: '',
    order_date: today,
    delivery_deadline: nextWeek,
    total_amount: 0,
    items_ordered: 0,
    items_shipped: 0,
    unit_name: 'Kardus',
    pcs_per_unit: 1,
    factory_price: 0,
    status: 'PENDING',
    dc_destination: '',
    po_schedule: today,
    invoice_photo_url: '',
    receipt_photo_url: '',
    billing_schedule: '',
    return_schedule: '',
    event_schedule: '',
    spg_schedule: '',
    stock_update_log: '',
    promo_update_log: '',
    supermarket_name: '',
    md_name: '',
  }
  poItems.value = []
  if (products.value.length > 0) {
    addPOItem()
  }
  showAdvancedOptions.value = false
  showPOModal.value = true
}

const openEditPOModal = (po: any) => {
  isEditing.value = true
  showAdvancedOptions.value = false
  poForm.value = {
    ...po,
    po_schedule: po.po_schedule ? po.po_schedule.split('T')[0] : '',
    billing_schedule: po.billing_schedule ? po.billing_schedule.split('T')[0] : '',
    return_schedule: po.return_schedule ? po.return_schedule.split('T')[0] : '',
    event_schedule: po.event_schedule ? po.event_schedule.split('T')[0] : '',
    spg_schedule: po.spg_schedule ? po.spg_schedule.split('T')[0] : '',
  }
  // Try to load items if exists from backend, otherwise fallback to empty
  poItems.value = po.items || []
  showPOModal.value = true
}

const savePO = async () => {
  if (!poForm.value.po_number || !poForm.value.key_account_id) {
    toast.error('Nomor PO dan Ritel wajib diisi!')
    return
  }

  try {
    const payload = {
      ...poForm.value,
      items: poItems.value
    }

    if (isEditing.value) {
      await apiClient.put(`/modern-trade/purchase-orders/${poForm.value.id}`, payload)
      toast.success('Purchase Order berhasil diperbarui!')
    } else {
      await apiClient.post('/modern-trade/purchase-orders', payload)
      toast.success('Purchase Order baru didaftarkan dan draf SO dibuat!')
    }
    showPOModal.value = false
    fetchPurchaseOrders()
  } catch (err: any) {
    toast.error('Gagal menyimpan PO: ' + (err.response?.data?.error || err.message))
  }
}

const deletePO = async (id: string) => {
  if (confirm('Apakah Anda yakin ingin menghapus PO ini?')) {
    try {
      await apiClient.delete(`/modern-trade/purchase-orders/${id}`)
      toast.success('PO berhasil dihapus!')
      fetchPurchaseOrders()
    } catch (err: any) {
      toast.error('Gagal menghapus PO: ' + (err.response?.data?.error || err.message))
    }
  }
}

// Tax Approval State
const showTaxModal = ref(false)
const showPreviewModal = ref(false)
const previewDocs = ref<any>(null)

const openPreviewModal = (po: any) => {
  previewDocs.value = {
    invoice: po.invoice_photo_url,
    receipt: po.receipt_photo_url,
    delivery: po.delivery_note_url
  }
  showPreviewModal.value = true
}
const taxForm = ref({
  id: '',
  billing_note_file: null as File | null,
  tax_invoice_file: null as File | null,
  delivery_note_file: null as File | null,
  destination_branch_id: ''
})

const openTaxModal = (po: any) => {
  taxForm.value = {
    id: po.id,
    billing_note_file: null,
    tax_invoice_file: null,
    delivery_note_file: null,
    destination_branch_id: ''
  }
  showTaxModal.value = true
}

const submitTaxForm = async () => {
  if (!taxForm.value.billing_note_file || !taxForm.value.tax_invoice_file || !taxForm.value.delivery_note_file || !taxForm.value.destination_branch_id) {
    toast.error('Semua dokumen pajak dan cabang pengirim wajib diisi (berkas harus diunggah)!')
    return
  }
  try {
    const formData = new FormData()
    formData.append('billing_note_file', taxForm.value.billing_note_file)
    formData.append('tax_invoice_file', taxForm.value.tax_invoice_file)
    formData.append('delivery_note_file', taxForm.value.delivery_note_file)
    formData.append('destination_branch_id', taxForm.value.destination_branch_id)
    
    await apiClient.patch(`/admin/sales-orders/mt-approve-tax/${taxForm.value.id}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    toast.success('Persetujuan Pajak PO MT berhasil!')
    showTaxModal.value = false
    fetchPurchaseOrders()
  } catch (err: any) {
    toast.error('Gagal approve pajak: ' + (err.response?.data?.error || err.message))
  }
}

const approvePODelivery = async (po: any) => {
  if (!confirm(`Konfirmasi penerimaan barang untuk PO ${po.po_number}? Saldo stok toko MT akan otomatis bertambah.`)) {
    return
  }
  try {
    const res = await apiClient.post(`/modern-trade/purchase-orders/${po.id}/approve-delivery`)
    toast.success(res.data?.message || 'Penerimaan pengiriman berhasil di-approve!')
    fetchPurchaseOrders()
  } catch (err: any) {
    toast.error('Gagal approve pengiriman: ' + (err.response?.data?.error || err.message))
  }
}

// SLA Overall Removed

const totalPOValue = computed(() =>
  purchaseOrders.value.reduce((sum, item) => sum + Number(item.total_amount || 0), 0)
)

// Filtering
const filteredPOs = computed(() => {
  return purchaseOrders.value.filter((po) => {
    const matchesSearch =
      po.po_number.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      po.key_account.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (po.dc_destination || '').toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesStatus =
      selectedStatus.value === 'ALL' || po.status === selectedStatus.value
    return matchesSearch && matchesStatus
  })
})

// File upload handler
const isUploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const triggerUpload = () => {
  if (keyAccounts.value.length === 0) {
    toast.error('Belum ada data Key Account ritel modern di sistem! Silakan daftarkan Key Account terlebih dahulu di tab KAM.')
    return
  }
  fileInput.value?.click()
}

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  isUploading.value = true
  try {
    const res = await apiClient.post('/modern-trade/purchase-orders/upload-edi', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    toast.success(`File PO ${res.data.filename || file.name} berhasil diunggah dan diproses secara dinamis!`)
    fetchPurchaseOrders()
  } catch (err: any) {
    toast.error('Gagal mengunggah file PO: ' + (err.response?.data?.error || err.message))
  } finally {
    isUploading.value = false
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  }
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2.5 py-1 text-xs font-bold text-red-600 bg-red-50 border border-red-200/50 rounded-full">
            Modern Trade (MT)
          </span>
        </div>
        <h1 class="text-2xl md:text-3xl font-black text-slate-900 tracking-tight mt-1 flex items-center flex-wrap gap-2">
          <span>PO & Pengiriman MT</span>
          <span v-if="masterStore.selectedBranchName" class="text-primary text-xl md:text-2xl font-bold">
            — {{ masterStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Pantau pesanan masuk ritel modern, status pemenuhan logistik, dan pencapaian target service level (SLA).
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <input
          type="file"
          ref="fileInput"
          @change="handleFileUpload"
          accept=".edi,.xml,.pdf,.xlsx,.csv,.txt"
          class="hidden"
        />
        <Button
          @click="openNewPOModal"
          class="bg-red-600 hover:bg-red-700 text-white font-bold rounded-xl shadow-md border border-red-200"
        >
          <Plus class="w-4 h-4 mr-2" />
          Tambah PO Manual
        </Button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-red-50 text-red-600 rounded-xl">
          <FileText class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total PO Diterima</p>
          <p class="text-2xl font-black text-slate-800">{{ purchaseOrders.length }} PO</p>
        </div>
      </div>

      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-amber-50 text-amber-600 rounded-xl">
          <Clock class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Menunggu Pengiriman</p>
          <p class="text-2xl font-black text-slate-800">
            {{ purchaseOrders.filter(po => po.status === 'PENDING').length }} PO
          </p>
        </div>
      </div>
      <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-3 bg-blue-50 text-blue-600 rounded-xl">
          <DollarSign class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Nilai Order</p>
          <p class="text-xl font-black text-slate-800 truncate" :title="formatCurrency(totalPOValue)">
            {{ formatCurrency(totalPOValue) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div class="relative w-full sm:w-[320px]">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari No. PO atau Ritel..."
          class="w-full pl-9 pr-4 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-red-500"
        />
      </div>

      <div class="flex items-center gap-2">
        <select
          v-model="selectedStatus"
          class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-red-500 bg-white"
        >
          <option value="ALL">Semua Status</option>
          <option value="PENDING">Pending (Baru Dibuat)</option>
          <option value="APPROVED">Approved (Pajak)</option>
          <option value="WAREHOUSE_RECEIVED">SJ Masuk Gudang</option>
          <option value="LOADING">Diproses Muat</option>
          <option value="IN_DELIVERY">Sedang Dikirim</option>
          <option value="DELIVERED_AND_UPLOADED">Selesai Kirim & Bukti TF</option>
          <option value="SETTLED">Settlement Selesai</option>
        </select>
      </div>
    </div>

    <!-- Data Table -->
    <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full border-collapse text-left">
          <thead>
            <tr class="bg-slate-50/70 border-b border-slate-100 text-xs font-black text-slate-400 uppercase tracking-wider">
              <th class="py-4 px-6">No. PO & Ritel</th>
              <th class="py-4 px-6">Tgl Order</th>
              <th class="py-4 px-6">Batas Pengiriman</th>
              <th class="py-4 px-6">Dikirim Oleh</th>
              <th class="py-4 px-6 text-right">Kuantitas</th>
              <th class="py-4 px-6 text-right">Nilai Tagihan</th>
              <th class="py-4 px-6 text-center">Status</th>
              <th class="py-4 px-6 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50 text-slate-700 text-sm">
            <tr v-for="po in filteredPOs" :key="po.id" class="hover:bg-slate-50/30 transition-all">
              <td class="py-4 px-6">
                <div>
                  <span class="font-extrabold text-slate-900 block">{{ po.po_number }}</span>
                  <span class="text-xs text-red-600 font-bold block">{{ po.key_account }}</span>
                  <span class="text-[10px] text-slate-400 font-medium" v-if="po.supermarket_name || po.md_name">
                    {{ po.supermarket_name }} • MD: {{ po.md_name }}
                  </span>
                </div>
              </td>
              <td class="py-4 px-6 text-xs font-medium text-slate-500">
                <div class="flex flex-col gap-1">
                  <div class="flex items-center gap-1.5">
                    <Calendar class="w-3.5 h-3.5" />
                    <span>{{ po.order_date }}</span>
                  </div>
                  <div class="flex items-center gap-1.5 text-red-600 font-semibold" v-if="po.po_schedule">
                    <Calendar class="w-3.5 h-3.5 text-red-400" />
                    <span>Jadwal PO: {{ po.po_schedule.split('T')[0] }}</span>
                  </div>
                </div>
              </td>
              <td class="py-4 px-6 text-xs font-medium text-slate-500">
                <div class="flex flex-col gap-1">
                  <div class="flex items-center gap-1.5" :class="po.status === 'PENDING' ? 'text-amber-600 font-bold' : ''">
                    <Clock class="w-3.5 h-3.5" />
                    <span> {{ po.delivery_deadline }}</span>
                  </div>
                  <div class="flex items-center gap-1.5 text-slate-600" v-if="po.billing_schedule">
                    <Calendar class="w-3.5 h-3.5 text-slate-400" />
                    <span>Jadwal Tagihan: {{ po.billing_schedule.split('T')[0] }}</span>
                  </div>
                </div>
              </td>
              <td class="py-4 px-6 text-xs font-medium text-slate-500">
                <div class="flex flex-col gap-1">
                  <span class="font-semibold text-slate-650">{{ po.dc_destination }}</span>
                  <div class="flex flex-wrap gap-1 text-[10px]">
                    <span class="bg-rose-50 text-rose-600 px-1 rounded font-bold" v-if="po.return_schedule">Retur: {{ po.return_schedule.split('T')[0] }}</span>
                    <span class="bg-blue-50 text-blue-600 px-1 rounded font-bold" v-if="po.event_schedule">Event: {{ po.event_schedule.split('T')[0] }}</span>
                    <span class="bg-amber-50 text-amber-600 px-1 rounded font-bold" v-if="po.spg_schedule">SPG: {{ po.spg_schedule.split('T')[0] }}</span>
                  </div>
                </div>
              </td>
              <td class="py-4 px-6 text-right font-medium">
                <div class="flex flex-col items-end gap-1">
                  <div v-if="po.sales_order && po.sales_order.items && po.sales_order.items.length > 0" class="flex flex-col items-end w-full">
                    <div v-for="item in po.sales_order.items" :key="item.id" class="text-[10px] text-slate-600 truncate max-w-[150px]">
                      {{ item.product?.name || 'Item' }} <span class="font-bold text-slate-800 ml-1">x{{ item.quantity }}</span>
                    </div>
                  </div>
                  <span v-else>{{ po.items_ordered }} Pcs</span>
                </div>
              </td>

              <td class="py-4 px-6 text-right font-black text-slate-800">
                <div class="flex flex-col items-end gap-1">
                  <span>{{ formatCurrency(po.total_amount) }}</span>
                  <span class="text-[9px] text-emerald-600 max-w-[120px] truncate" :title="po.promo_update_log" v-if="po.promo_update_log">Promo: {{ po.promo_update_log }}</span>
                </div>
              </td>
              <td class="py-4 px-6 text-center">
                <div class="flex flex-col items-center gap-1">
                  <span
                    class="px-2.5 py-1 rounded-lg font-bold text-[10px] tracking-wide"
                    :class="getStatusClass(po.status)"
                  >
                    {{ formatStatus(po.status) }}
                  </span>
                  <span class="text-[9px] text-slate-400 truncate max-w-[120px] mt-1" :title="po.stock_update_log" v-if="po.stock_update_log">Stock: {{ po.stock_update_log }}</span>
                </div>
              </td>
              <td class="py-4 px-6 text-center">
                <div class="flex items-center justify-center gap-2">
                  <button
                    v-if="po.invoice_photo_url || po.receipt_photo_url || po.delivery_note_url"
                    @click="openPreviewModal(po)"
                    class="p-1.5 hover:bg-blue-50 text-slate-400 hover:text-blue-600 rounded-lg transition-colors"
                    title="Preview Dokumen"
                  >
                    <Eye class="w-4 h-4" />
                  </button>
                  <button
                    v-if="po.status === 'PENDING'"
                    @click="openTaxModal(po)"
                    class="p-1.5 hover:bg-emerald-50 text-slate-400 hover:text-emerald-600 rounded-lg transition-colors"
                    title="Approve Pajak"
                  >
                    <CheckCircle2 class="w-4 h-4" />
                  </button>
                  <button
                    v-if="po.status !== 'DELIVERED_AND_UPLOADED' && po.status !== 'SETTLED'"
                    @click="approvePODelivery(po)"
                    class="p-1.5 hover:bg-teal-50 text-slate-400 hover:text-teal-600 rounded-lg transition-colors"
                    title="Approve Terima Barang (Tambah Stok Toko MT)"
                  >
                    <Truck class="w-4 h-4" />
                  </button>
                  <button
                    @click="openEditPOModal(po)"
                    class="p-1.5 hover:bg-red-50 text-slate-400 hover:text-red-600 rounded-lg transition-colors"
                    title="Ubah PO"
                  >
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button
                    @click="deletePO(po.id)"
                    class="p-1.5 hover:bg-rose-50 text-slate-400 hover:text-rose-600 rounded-lg transition-colors"
                    title="Hapus PO"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredPOs.length === 0">
              <td colspan="9" class="py-8 text-center text-slate-400">
                Tidak ada data Purchase Order ditemukan.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- PO Modal Form -->
    <div
      v-if="showPOModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white rounded-[24px] w-full max-w-2xl shadow-2xl overflow-hidden p-6 relative my-8">
        <h3 class="text-lg font-black text-slate-900 mb-4">
          {{ isEditing ? 'Edit Purchase Order' : 'Tambah Purchase Order' }}
        </h3>

        <div class="space-y-4 text-xs font-semibold text-slate-650 max-h-[60vh] overflow-y-auto pr-1">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Sales MT</label>
              <select
                v-model="poForm.employee_id"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none bg-white"
              >
                <option value="">Pilih Sales MT</option>
                <option v-for="emp in salesMTs" :key="emp.id" :value="emp.id">
                  {{ emp.first_name }} {{ emp.last_name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Ritel Modern (Key Account)</label>
              <select
                v-model="poForm.key_account_id"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none bg-white"
              >
                <option v-for="ka in keyAccounts" :key="ka.id" :value="ka.id">
                  {{ ka.name }} ({{ ka.company_name }})
                </option>
              </select>
            </div>
          </div>
          
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Nomor Purchase Order (PO)</label>
              <input
                v-model="poForm.po_number"
                type="text"
                placeholder="Contoh: PO-2026-0001"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="mt-4 mb-2">
            <button @click="showAdvancedOptions = !showAdvancedOptions" class="text-[11px] flex items-center gap-1 text-indigo-600 font-bold hover:text-indigo-800 transition-colors">
              <span v-if="!showAdvancedOptions">Tampilkan Pengaturan Lanjutan (Detail & Checklist)</span>
              <span v-else>Sembunyikan Pengaturan Lanjutan</span>
            </button>
          </div>

          <template v-if="showAdvancedOptions">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Supermarket Name</label>
                <input
                  v-model="poForm.supermarket_name"
                  type="text"
                  placeholder="Contoh: Indomaret Point Sudirman"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">MD Penanggung Jawab</label>
              <input
                v-model="poForm.md_name"
                type="text"
                placeholder="Contoh: Budi Santoso"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Tanggal Order</label>
              <input
                v-model="poForm.order_date"
                type="date"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Batas Pengiriman</label>
              <input
                v-model="poForm.delivery_deadline"
                type="date"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Satuan Dinamis</label>
              <input
                v-model="poForm.unit_name"
                type="text"
                placeholder="Contoh: Kardus"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Isi per Satuan (Pcs)</label>
              <input
                v-model.number="poForm.pcs_per_unit"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Ordered Qty</label>
              <input
                v-model.number="poForm.items_ordered"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Harga Pabrik (per Satuan)</label>
              <input
                v-model.number="poForm.factory_price"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Total Harga (Fleksibel)</label>
              <input
                v-model.number="poForm.total_amount"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Shipped Qty</label>
                <input
                  v-model.number="poForm.items_shipped"
                  type="number"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
            </div>
          </template>
          
          <div class="grid grid-cols-2 gap-4 mt-3">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Tujuan DC (Distribution Center)</label>
              <input
                v-model="poForm.dc_destination"
                type="text"
                placeholder="Contoh: DC Indomaret Tangerang"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Status PO</label>
              <select
                v-model="poForm.status"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none bg-white"
              >
                <option value="PENDING">Pending (Baru Dibuat)</option>
                <option value="APPROVED">Approved (Pajak)</option>
                <option value="WAREHOUSE_RECEIVED">SJ Masuk Gudang</option>
                <option value="LOADING">Diproses Muat</option>
                <option value="IN_DELIVERY">Sedang Dikirim</option>
                <option value="DELIVERED_AND_UPLOADED">Selesai Kirim & Bukti TF</option>
                <option value="SETTLED">Settlement Selesai</option>
              </select>
            </div>
          </div>

          <!-- Product Items Selection -->
          <div class="border-t border-slate-100 pt-3">
            <div class="flex justify-between items-center mb-3">
              <h4 class="text-xs font-black text-slate-800">Daftar Produk Pesanan</h4>
              <Button @click="addPOItem" size="sm" variant="outline" class="h-7 text-[10px] px-2 py-0 border-dashed border-red-200 text-red-600 hover:bg-red-50">
                <Plus class="w-3 h-3 mr-1" /> Tambah Produk
              </Button>
            </div>
            <div v-for="(item, idx) in poItems" :key="idx" class="flex gap-2 items-center mb-2 bg-slate-50 p-2 rounded-lg border border-slate-100">
              <div class="flex-1">
                <select
                  v-model="item.product_id"
                  @change="onProductSelect(item)"
                  class="w-full px-2 py-1.5 text-xs border border-slate-200 rounded-md focus:ring-1 focus:ring-red-500 focus:outline-none bg-white"
                >
                  <option value="" disabled>Pilih Produk</option>
                  <option v-for="p in products" :key="p.id" :value="p.id">
                    {{ p.sku || p.id }} - {{ p.name }} (Stok: {{ getProductStock(p.id) }})
                  </option>
                </select>
              </div>
              <div class="w-20">
                <input
                  v-model.number="item.quantity"
                  type="number"
                  min="1"
                  placeholder="Qty"
                  class="w-full px-2 py-1.5 text-xs border border-slate-200 rounded-md focus:ring-1 focus:ring-red-500 focus:outline-none text-center"
                />
              </div>
              <div class="w-28 text-right">
                <span class="text-xs font-semibold text-slate-700">{{ formatCurrency(item.unit_price * item.quantity) }}</span>
              </div>
              <button @click="removePOItem(idx)" class="text-slate-400 hover:text-rose-500 p-1">
                <Trash2 class="w-3.5 h-3.5" />
              </button>
            </div>
            <div v-if="poItems.length === 0" class="text-center py-4 text-xs text-slate-400 border border-dashed border-slate-200 rounded-lg">
              Belum ada produk yang dipilih
            </div>
          </div>

          <div class="border-t border-slate-100 pt-3" v-if="isEditing && showAdvancedOptions">
            <h4 class="text-xs font-black text-black mb-3">Mekanisme Checklist Supervisor & Operasional</h4>
            
            <div class="grid grid-cols-2 gap-4 mb-3">
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Jadwal PO</label>
                <input
                  v-model="poForm.po_schedule"
                  type="date"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Jadwal Penagihan</label>
                <input
                  v-model="poForm.billing_schedule"
                  type="date"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4 mb-3">
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Jadwal Retur</label>
                <input
                  v-model="poForm.return_schedule"
                  type="date"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Jadwal Event</label>
                <input
                  v-model="poForm.event_schedule"
                  type="date"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4 mb-3">
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Jadwal SPG</label>
                <input
                  v-model="poForm.spg_schedule"
                  type="date"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Update Stok Log</label>
                <input
                  v-model="poForm.stock_update_log"
                  type="text"
                  placeholder="Contoh: Stok ditambah 100 pcs untuk promo"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4 mb-3">
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Update Promo Log</label>
                <input
                  v-model="poForm.promo_update_log"
                  type="text"
                  placeholder="Contoh: Beli 2 gratis 1 mulai 25 Juni"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Foto Nota / Invoice URL</label>
                <input
                  v-model="poForm.invoice_photo_url"
                  type="text"
                  placeholder="https://..."
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
                />
              </div>
            </div>

            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Foto Tanda Terima MT URL</label>
              <input
                v-model="poForm.receipt_photo_url"
                type="text"
                placeholder="https://..."
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none"
              />
            </div>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-3">
          <Button
            @click="showPOModal = false"
            variant="outline"
            class="px-4 py-2 border border-slate-200 rounded-xl text-xs text-slate-500 font-bold hover:bg-slate-50"
          >
            Batal
          </Button>
          <Button
            @click="savePO"
            class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-xl text-xs font-bold"
          >
            Simpan PO
          </Button>
        </div>
      </div>
    </div>

    <!-- Tax Approval Modal -->
    <div
      v-if="showTaxModal"
      class="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white rounded-[24px] w-full max-w-lg shadow-2xl overflow-hidden p-6 relative my-8">
        <h3 class="text-lg font-black text-slate-900 mb-4">
          Approve Pajak PO MT
        </h3>
        <p class="text-xs text-slate-500 mb-4">
          Unggah (masukkan URL) dokumen faktur pajak, faktur penjualan, surat jalan, dan tentukan cabang pengirim.
        </p>

        <div class="space-y-4 text-xs font-semibold text-slate-650">
          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Faktur Penjualan (Billing Note)</label>
            <input
              type="file"
              accept=".pdf,.png,.jpg,.jpeg"
              @change="(e) => taxForm.billing_note_file = (e.target as HTMLInputElement).files?.[0] || null"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none file:mr-4 file:py-1 file:px-4 file:rounded-full file:border-0 file:text-[10px] file:font-semibold file:bg-red-50 file:text-red-700 hover:file:bg-red-100"
            />
          </div>
          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Faktur Pajak (Tax Invoice)</label>
            <input
              type="file"
              accept=".pdf,.png,.jpg,.jpeg"
              @change="(e) => taxForm.tax_invoice_file = (e.target as HTMLInputElement).files?.[0] || null"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none file:mr-4 file:py-1 file:px-4 file:rounded-full file:border-0 file:text-[10px] file:font-semibold file:bg-red-50 file:text-red-700 hover:file:bg-red-100"
            />
          </div>
          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Surat Jalan (Delivery Note)</label>
            <input
              type="file"
              accept=".pdf,.png,.jpg,.jpeg"
              @change="(e) => taxForm.delivery_note_file = (e.target as HTMLInputElement).files?.[0] || null"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none file:mr-4 file:py-1 file:px-4 file:rounded-full file:border-0 file:text-[10px] file:font-semibold file:bg-red-50 file:text-red-700 hover:file:bg-red-100"
            />
          </div>
          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Cabang Pengirim</label>
            <select
              v-model="taxForm.destination_branch_id"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-red-500 focus:outline-none bg-white"
            >
              <option value="">Pilih Cabang</option>
              <option v-for="branch in masterStore.branches" :key="branch.id" :value="branch.id">
                {{ branch.name }}
              </option>
            </select>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-3">
          <Button
            @click="showTaxModal = false"
            variant="outline"
            class="px-4 py-2 border border-slate-200 rounded-xl text-xs text-slate-500 font-bold hover:bg-slate-50"
          >
            Batal
          </Button>
          <Button
            @click="submitTaxForm"
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold"
          >
            Submit Approve
          </Button>
        </div>
      </div>
    </div>

  </div>

    <!-- Preview Modal -->
    <div
      v-if="showPreviewModal"
      class="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white rounded-[24px] w-full max-w-lg shadow-2xl overflow-hidden p-6 relative my-8">
        <h3 class="text-lg font-black text-slate-900 mb-4 flex items-center justify-between">
          <span>Preview Dokumen Pajak</span>
          <button @click="showPreviewModal = false" class="text-slate-400 hover:text-slate-600">
            <X class="w-5 h-5" />
          </button>
        </h3>
        
        <div class="space-y-4 text-xs font-semibold text-slate-650" v-if="previewDocs">
          <div v-if="previewDocs.receipt" class="p-3 bg-slate-50 rounded-xl flex justify-between items-center">
            <span>Faktur Penjualan (Billing Note)</span>
            <a :href="previewDocs.receipt" target="_blank" class="text-red-600 hover:underline">Lihat Dokumen</a>
          </div>
          <div v-if="previewDocs.invoice" class="p-3 bg-slate-50 rounded-xl flex justify-between items-center">
            <span>Faktur Pajak (Tax Invoice)</span>
            <a :href="previewDocs.invoice" target="_blank" class="text-red-600 hover:underline">Lihat Dokumen</a>
          </div>
          <div v-if="previewDocs.delivery" class="p-3 bg-slate-50 rounded-xl flex justify-between items-center">
            <span>Surat Jalan (Delivery Note)</span>
            <a :href="previewDocs.delivery" target="_blank" class="text-red-600 hover:underline">Lihat Dokumen</a>
          </div>
          
          <div v-if="!previewDocs.receipt && !previewDocs.invoice && !previewDocs.delivery" class="text-center text-slate-500 py-4">
            Belum ada dokumen yang diunggah.
          </div>
        </div>
      </div>
    </div>
</template>
