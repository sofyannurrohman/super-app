<script setup lang="ts">
import { ref, onMounted, computed, reactive, watch, nextTick } from 'vue'
import { useDeliveryStore } from '@/stores/delivery'
import { useMasterDataStore } from '@/stores/masterData'
import { useLayoutStore } from '@/stores/layout'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { filterSellableProducts } from '@/utils/productUnits'
import {
  FileText,
  Search,
  Truck,
  CheckCircle2,
  Calendar,
  User,
  Car,
  ArrowRight,
  Eye,
  Clock,
  Package,
  MapPin,
  X,
  Pencil,
  Trash2,
  PlusCircle,
  Camera,
  AlertTriangle,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardFooter } from '@/components/ui/card/index'
import DataTable from '@/components/DataTable.vue'
import RefreshDataButton from '@/components/ui/RefreshDataButton.vue'
import DigitalSuratJalanPrint from '@/components/sales/DigitalSuratJalanPrint.vue'
import AppendSubsequentOrderModal from '@/components/sales/AppendSubsequentOrderModal.vue'
import { useDataRefresh } from '@/composables/useDataRefresh'
import { Printer } from 'lucide-vue-next'

const formatDate = (date: string) => {
  if (!date) return '—'
  try {
    return new Intl.DateTimeFormat('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(date))
  } catch (e) {
    return date
  }
}

const deliveryStore = useDeliveryStore()
const masterStore = useMasterDataStore()
const layoutStore = useLayoutStore()
const authStore = useAuthStore()
const searchQuery = ref('')
const selectedStatus = ref('ALL')
const selectedDate = ref('')
const selectedMonth = ref('ALL')
const selectedYear = ref('ALL')

const currentYear = new Date().getFullYear()
const years = computed(() => {
  return Array.from({ length: 5 }, (_, i) => (currentYear - i).toString())
})

const months = [
  { value: '1', label: 'Januari' },
  { value: '2', label: 'Februari' },
  { value: '3', label: 'Maret' },
  { value: '4', label: 'April' },
  { value: '5', label: 'Mei' },
  { value: '6', label: 'Juni' },
  { value: '7', label: 'Juli' },
  { value: '8', label: 'Agustus' },
  { value: '9', label: 'September' },
  { value: '10', label: 'Oktober' },
  { value: '11', label: 'November' },
  { value: '12', label: 'Desember' },
]

const showDetailModal = ref(false)
const showEditModal = ref(false)
const showAppendModal = ref(false)
const appendTargetBatch = ref<any>(null)
const selectedBatch = ref<any>(null)
const selectedDO = ref<any>(null)
const isSubmitting = ref(false)
const zoomedQR = ref<string | null>(null)
const zoomedPOD = ref<string | null>(null)

const openAppendModal = (batch: any) => {
  appendTargetBatch.value = batch
  showAppendModal.value = true
}

const onAppendSuccess = async () => {
  await deliveryStore.fetchBatches(masterStore.selectedBranchId || '')
  if (selectedBatch.value && appendTargetBatch.value && selectedBatch.value.id === appendTargetBatch.value.id) {
    const updated = deliveryStore.batches.find((b: any) => b.id === selectedBatch.value.id)
    if (updated) selectedBatch.value = updated
  }
}

const editingBatch = ref<any>(null)
const batchForm = reactive({
  driver_id: '',
  vehicle_id: '',
  status: '',
  delivery_date: '',
})

const drivers = ref<any[]>([])
const vehicles = ref<any[]>([])

const fetchDriversAndVehicles = async () => {
  const branchId = masterStore.selectedBranchId || ''
  const companyId = masterStore.selectedBranchCompanyId || ''
  try {
    const [dRes, sRes, bRes, vRes] = await Promise.allSettled([
      apiClient.get(`/employees?job_title=DRIVER&company_id=${companyId}&limit=all`),
      apiClient.get(`/employees?job_title=SALESMAN&company_id=${companyId}&limit=all`),
      apiClient.get(`/employees?branch_id=${branchId}&company_id=${companyId}&limit=all`),
      apiClient.get(`/vehicles?branch_id=${branchId}&bypass_company=true`)
    ])
    const driversData = dRes.status === 'fulfilled' ? dRes.value.data?.data || [] : []
    const salesData = sRes.status === 'fulfilled' ? sRes.value.data?.data || [] : []
    const branchData = bRes.status === 'fulfilled' ? bRes.value.data?.data || [] : []
    salesEmployees.value = salesData
    const allEmps = [...driversData, ...salesData, ...branchData]
    drivers.value = Array.from(new Map(allEmps.map((e: any) => [e.id, e])).values())
    if (vRes.status === 'fulfilled') {
      vehicles.value = vRes.value.data?.data || vRes.value.data || []
    }
  } catch (err) {
    console.error('Gagal mengambil data supir & armada', err)
  }
}

const currentBranchName = computed(() => {
  const branch = masterStore.branches.find((b) => b.id === masterStore.selectedBranchId)
  return branch ? branch.name : 'Semua Cabang'
})

const isLoading = ref(false)

const fetchData = async (force = false) => {
  isLoading.value = true
  try {
    if (masterStore.branches.length === 0 || force) {
      await masterStore.fetchBranches(force)
    }
    await Promise.allSettled([
      deliveryStore.fetchBatches(masterStore.selectedBranchId || ''),
      fetchDriversAndVehicles(),
      fetchProducts(),
    ])

    // Sinkronkan detail Surat Jalan yang sedang aktif dipilih dengan data terbaru
    if (selectedBatch.value?.id && !isReportingMode.value) {
      const updatedBatch = deliveryStore.batches.find((b: any) => b.id === selectedBatch.value.id)
      if (updatedBatch) {
        selectedBatch.value = updatedBatch
      }
    }
    if (selectedDO.value?.id) {
      const updatedDO = deliveryStore.batches.find((b: any) => b.id === selectedDO.value.id)
      if (updatedDO) {
        selectedDO.value = updatedDO
      }
    }
  } catch (err) {
    toast.error('Gagal mengambil data surat jalan')
  } finally {
    isLoading.value = false
  }
}

const {
  isRefreshing,
  lastRefreshedTimeStr,
  markRefreshed,
  refresh: triggerRefresh,
} = useDataRefresh(
  async ({ isManual }) => {
    await fetchData(isManual)
    if (showEditModal.value) {
      await Promise.allSettled([fetchPendingOrders(), fetchPendingSubsequents()])
    }
  },
  {
    successMessage: 'Data Surat Jalan berhasil diperbarui',
    canAutoRefresh: () =>
      !isReportingMode.value &&
      !showEditModal.value &&
      !showAppendModal.value &&
      !showForceInvoiceModal.value &&
      !showReviseQtyModal.value &&
      !showManualSubsequentModal.value &&
      !isSubmitting.value &&
      !isLoading.value,
  },
)

onMounted(() => {
  fetchData().then(() => markRefreshed())
})

watch(
  () => masterStore.selectedBranchId,
  () => {
    fetchData().then(() => markRefreshed())
  },
)

watch(
  () => masterStore.selectedSalesType,
  () => {
    fetchData().then(() => markRefreshed())
  },
)

const pendingOrders = ref<any[]>([])

const fetchPendingOrders = async () => {
  const branchId = masterStore.selectedBranchId
  try {
    let salesType = masterStore.selectedSalesType
    if (authStore.user?.role === 'admin_gudang' || layoutStore.currentModule === 'warehouse') {
      salesType = 'ALL_SALES'
    }
    const res = await apiClient.get(`/admin/sales-orders?branch_id=${branchId}&status=CONFIRMED&sales_type=${salesType}`)
    pendingOrders.value = res.data?.data || []
  } catch (err) {
    console.error('Gagal memuat pesanan pending', err)
  }
}

const pendingSubsequents = ref<any[]>([])

const fetchPendingSubsequents = async () => {
  const branchId = masterStore.selectedBranchId
  try {
    const res = await apiClient.get(`/delivery/subsequent/pending?branch_id=${branchId}`)
    pendingSubsequents.value = res.data?.data || []
  } catch (err) {
    console.error('Gagal memuat barang susulan pending', err)
  }
}

const products = ref<any[]>([])
const salesEmployees = ref<any[]>([])
const showManualSubsequentModal = ref(false)
const manualSubsequentForm = ref({
  store_id: '',
  salesman_id: '',
  product_id: '',
  quantity: 1,
  notes: ''
})
const isSubmittingSubsequent = ref(false)

const fetchProducts = async () => {
  try {
    const branchId = masterStore.selectedBranchId
    const params = branchId && branchId !== 'ALL_BRANCHES' && branchId !== 'ALL' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/factory/products${params}`)
    const rawData = res.data?.data || res.data || []
    products.value = filterSellableProducts(rawData)
  } catch {
    console.error('Gagal memuat produk')
  }
}

const createManualSubsequent = async () => {
  if (!manualSubsequentForm.value.store_id || !manualSubsequentForm.value.salesman_id || !manualSubsequentForm.value.product_id || manualSubsequentForm.value.quantity < 1) {
    toast.error('Harap lengkapi formulir (Toko, Salesman, Produk, Kuantitas >= 1)')
    return
  }
  isSubmittingSubsequent.value = true
  try {
    await apiClient.post('/delivery/subsequent', manualSubsequentForm.value)
    toast.success('Barang susulan manual berhasil dibuat')
    showManualSubsequentModal.value = false
    manualSubsequentForm.value = {
      store_id: '', salesman_id: '', product_id: '', quantity: 1, notes: ''
    }
    await fetchPendingSubsequents()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal membuat barang susulan manual')
  } finally {
    isSubmittingSubsequent.value = false
  }
}


const openEditModal = async (batch: any) => {
  editingBatch.value = JSON.parse(JSON.stringify(batch))
  batchForm.driver_id = batch.driver_id || ''
  batchForm.vehicle_id = batch.vehicle_id || ''
  batchForm.status = batch.status || ''
  batchForm.delivery_date = batch.delivery_date ? batch.delivery_date.split('T')[0] : ''

  await fetchDriversAndVehicles()
  await fetchPendingOrders()
  await fetchPendingSubsequents()
  await deliveryStore.fetchOutstandingInvoices(masterStore.selectedBranchId || '')
  showEditModal.value = true
}

function canEditOrDeleteBatch(status: string) {
  return (
    status !== 'PICKING' &&
    status !== 'ON_DELIVERY' &&
    status !== 'COMPLETED'
  )
}

const canEditItems = computed(() => {
  if (!editingBatch.value) return false
  return canEditOrDeleteBatch(editingBatch.value.status)
})

function removeItemFromEditingBatch(index: number | string) {
  const idx = typeof index === 'string' ? parseInt(index, 10) : index
  if (editingBatch.value && editingBatch.value.items) {
    editingBatch.value.items.splice(idx, 1)
  }
}

function isOrderSelectedForAdding(orderId: string) {
  if (!editingBatch.value || !editingBatch.value.items) return false
  return editingBatch.value.items.some(
    (i: any) => i.sales_order_id === orderId || i.sales_transaction_id === orderId
  )
}

function toggleOrderForAdding(order: any) {
  if (!editingBatch.value) return
  if (!editingBatch.value.items) {
    editingBatch.value.items = []
  }

  const idx = editingBatch.value.items.findIndex(
    (i: any) => i.sales_order_id === order.id || i.sales_transaction_id === order.id
  )
  if (idx > -1) {
    editingBatch.value.items.splice(idx, 1)
  } else {
    editingBatch.value.items.push({
      sales_order_id: order.id,
      sales_order: order,
      status: 'PENDING',
    })
  }
}

function isSubsequentSelectedForAdding(subId: string) {
  if (!editingBatch.value || !editingBatch.value.items) return false
  return editingBatch.value.items.some(
    (i: any) => i.sales_subsequent_item_id === subId
  )
}

function toggleSubsequentForAdding(sub: any) {
  if (!editingBatch.value) return
  if (!editingBatch.value.items) {
    editingBatch.value.items = []
  }

  const idx = editingBatch.value.items.findIndex(
    (i: any) => i.sales_subsequent_item_id === sub.id
  )
  if (idx > -1) {
    editingBatch.value.items.splice(idx, 1)
  } else {
    editingBatch.value.items.push({
      sales_subsequent_item_id: sub.id,
      sales_subsequent_item: sub,
      status: 'PENDING',
    })
  }
}

function isTempoInvoiceSelected(trxId: string) {
  if (!editingBatch.value || !editingBatch.value.items) return false
  return editingBatch.value.items.some(
    (i: any) => i.sales_transaction_id === trxId && i.is_collection_only === true
  )
}

function toggleTempoInvoiceForAdding(trx: any) {
  if (!editingBatch.value) return
  if (!editingBatch.value.items) {
    editingBatch.value.items = []
  }

  const idx = editingBatch.value.items.findIndex(
    (i: any) => i.sales_transaction_id === trx.id && i.is_collection_only === true
  )
  if (idx > -1) {
    editingBatch.value.items.splice(idx, 1)
  } else {
    editingBatch.value.items.push({
      sales_transaction_id: trx.id,
      sales_transaction: trx,
      is_collection_only: true,
      status: 'PENDING',
    })
  }
}


const isStartingDelivery = ref(false)

const confirmStartDelivery = async () => {
  if (!selectedBatch.value) return
  const confirmMsg = "Apakah Anda yakin ingin memulai pengiriman dan MEMOTONG STOK GUDANG? Aksi ini tidak dapat dibatalkan."
  if (!confirm(confirmMsg)) return
  
  isStartingDelivery.value = true
  try {
    await apiClient.post(`/delivery/batch/${selectedBatch.value.delivery_order_no}/start`)
    toast.success('Pengiriman berhasil dimulai, stok gudang terpotong!')
    fetchData()
    selectedBatch.value.status = 'ON_DELIVERY'
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal memulai pengiriman')
  } finally {
    isStartingDelivery.value = false
  }
}

const handleUpdateBatch = async () => {
  if (!editingBatch.value) return
  isSubmitting.value = true
  try {
    const payload = {
      company_id: editingBatch.value.company_id || masterStore.selectedBranchCompanyId,
      branch_id: editingBatch.value.branch_id || masterStore.selectedBranchId,
      driver_id: batchForm.driver_id || null,
      vehicle_id: batchForm.vehicle_id || null,
      status: batchForm.status || null,
      delivery_date: batchForm.delivery_date ? new Date(batchForm.delivery_date).toISOString() : null,
      sales_order_ids: editingBatch.value.items
        .filter((i: any) => i.sales_order_id && !i.is_collection_only)
        .map((i: any) => i.sales_order_id),
      collection_only_so_ids: editingBatch.value.items
        .filter((i: any) => i.sales_order_id && i.is_collection_only)
        .map((i: any) => i.sales_order_id),
      collection_only_trx_ids: editingBatch.value.items
        .filter((i: any) => i.sales_transaction_id && i.is_collection_only)
        .map((i: any) => i.sales_transaction_id),
      sales_transaction_ids: editingBatch.value.items
        .filter((i: any) => i.sales_transaction_id && !i.is_collection_only)
        .map((i: any) => i.sales_transaction_id),
      sales_subsequent_item_ids: editingBatch.value.items
        .map((i: any) => i.sales_subsequent_item_id)
        .filter(Boolean),
      sales_transfer_ids: editingBatch.value.items
        .map((i: any) => i.sales_transfer_id)
        .filter(Boolean),
    }

    await apiClient.put(`/delivery/batches/${editingBatch.value.id}`, payload)

    if (batchForm.status !== editingBatch.value.status) {
      if (batchForm.status === 'SUPERVISOR_APPROVED') {
        await apiClient.post(`/delivery/batch/${editingBatch.value.id}/supervisor-approve`)
      }
    }

    // Optimistic update: langsung perbarui data di store sebelum refetch selesai
    const batchId = editingBatch.value.id
    const storeIndex = deliveryStore.batches.findIndex((b: any) => b.id === batchId)
    if (storeIndex !== -1) {
      const updatedDriver = drivers.value.find((d: any) => d.id === batchForm.driver_id) || deliveryStore.batches[storeIndex].driver
      const updatedVehicle = vehicles.value.find((v: any) => v.id === batchForm.vehicle_id) || deliveryStore.batches[storeIndex].vehicle
      deliveryStore.batches[storeIndex] = {
        ...deliveryStore.batches[storeIndex],
        driver_id: batchForm.driver_id || null,
        vehicle_id: batchForm.vehicle_id || null,
        status: batchForm.status || deliveryStore.batches[storeIndex].status,
        delivery_date: batchForm.delivery_date ? new Date(batchForm.delivery_date).toISOString() : deliveryStore.batches[storeIndex].delivery_date,
        driver: updatedDriver,
        vehicle: updatedVehicle,
        items: editingBatch.value.items,
      }
      // Trigger Vue reactivity dengan replace array
      deliveryStore.batches = [...deliveryStore.batches]
    }

    toast.success('Surat Jalan berhasil diperbarui!')
    showEditModal.value = false
    // Full refetch di background untuk sinkronisasi data lengkap dari server
    fetchData()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal memperbarui Surat Jalan')
  } finally {
    isSubmitting.value = false
  }
}

const handleDeleteBatch = async (id: string) => {
  if (
    !confirm(
      'Apakah Anda yakin ingin menghapus Surat Jalan ini? Nota didalamnya akan kembali ke antrean pesanan.',
    )
  )
    return
  try {
    await apiClient.delete(`/delivery/batches/${id}`)
    toast.success('Surat Jalan berhasil dihapus!')
    await fetchData()
  } catch (err: any) {
    toast.error('Gagal menghapus Surat Jalan')
  }
}

const filteredBatches = computed(() => {
  return deliveryStore.batches
    .filter((b) => {
      let matchDateFilter = true
      let matchMonthFilter = true
      let matchYearFilter = true

      if (b.created_at) {
        const d = new Date(b.created_at)

        if (selectedDate.value) {
          const yyyy = d.getFullYear()
          const mm = String(d.getMonth() + 1).padStart(2, '0')
          const dd = String(d.getDate()).padStart(2, '0')
          matchDateFilter = `${yyyy}-${mm}-${dd}` === selectedDate.value
        }

        if (selectedMonth.value !== 'ALL') {
          matchMonthFilter = (d.getMonth() + 1).toString() === selectedMonth.value
        }

        if (selectedYear.value !== 'ALL') {
          matchYearFilter = d.getFullYear().toString() === selectedYear.value
        }
      } else {
        if (selectedDate.value || selectedMonth.value !== 'ALL' || selectedYear.value !== 'ALL') {
          matchDateFilter = false
        }
      }

      const q = searchQuery.value.toLowerCase()
      const matchesSearch =
        !q ||
        b.delivery_order_no?.toLowerCase().includes(q) ||
        b.driver?.first_name?.toLowerCase().includes(q) ||
        b.driver?.last_name?.toLowerCase().includes(q)

      const matchesStatus = selectedStatus.value === 'ALL' || b.status === selectedStatus.value

      return (
        matchesSearch && matchesStatus && matchDateFilter && matchMonthFilter && matchYearFilter
      )
    })
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})

// Computed for Product Matrix in SJ
const batchProductColumns = computed(() => {
  if (!selectedDO.value) return []

  const productMap = new Map()
  if (selectedDO.value.items) {
    selectedDO.value.items.forEach((item: any) => {
      const trx = item.sales_order || item.sales_transaction
      if (trx?.items) {
        trx.items.forEach((si: any) => {
          if (si.product) {
            productMap.set(si.product_id, {
              id: si.product_id,
              name: si.product.name
                .split(' ')
                .map((w: string) => w[0])
                .join(''), // Abbreviate
            })
          }
        })
      }
    })
  }

  if (selectedDO.value.spares) {
    selectedDO.value.spares.forEach((spare: any) => {
      if (spare.product) {
        productMap.set(spare.product_id, {
          id: spare.product_id,
          name: spare.product.name
            .split(' ')
            .map((w: string) => w[0])
            .join(''),
        })
      }
    })
  }

  return Array.from(productMap.values())
})

const getProductQty = (item: any, productId: string) => {
  const trx = item.sales_order || item.sales_transaction
  if (!trx || !trx.items) return 0
  const si = trx.items.find((i: any) => i.product_id === productId)
  return si ? si.quantity : 0
}

const getSpareQty = (productId: string) => {
  if (!selectedDO.value || !selectedDO.value.spares) return 0
  const spare = selectedDO.value.spares.find((s: any) => s.product_id === productId)
  return spare ? spare.quantity : 0
}

function hasDiscrepancy(item: any) {
  const payments = item.sales_order?.payments || item.sales_transaction?.payments || []
  return payments.some((p: any) => p.notes && p.notes.includes('[SELISIH'))
}

function getDiscrepancyNote(item: any) {
  const payments = item.sales_order?.payments || item.sales_transaction?.payments || []
  const p = payments.find((p: any) => p.notes && p.notes.includes('[SELISIH'))
  return p ? p.notes : ''
}

const calculateBatchTotal = computed(() => {
  if (!selectedDO.value || !selectedDO.value.items) return 0
  return selectedDO.value.items.reduce((sum: number, item: any) => {
    const trx = item.sales_order || item.sales_transaction
    return sum + (trx?.total_amount || 0)
  }, 0)
})

const calculateProductTotal = (productId: string) => {
  if (!selectedDO.value) return 0
  let total = 0
  if (selectedDO.value.items) {
    total += selectedDO.value.items.reduce((sum: number, item: any) => {
      return sum + getProductQty(item, productId)
    }, 0)
  }
  if (selectedDO.value.spares) {
    const spare = selectedDO.value.spares.find((s: any) => s.product_id === productId)
    if (spare) {
      total += spare.quantity || 0
    }
  }
  return total
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

const handlePrint = async () => {
  // Tunggu Vue selesai render modal sebelum membuka print dialog
  await nextTick()
  window.print()
}

const openPrintPreview = (batch: any) => {
  selectedDO.value = batch
}

const getStatusColor = (status: string) => {
  const map: Record<string, string> = {
    WAITING_APPROVAL: 'bg-yellow-100 text-yellow-700 border-yellow-200',
    SUPERVISOR_APPROVED: 'bg-blue-100 text-blue-700 border-blue-200',
    PENDING: 'bg-indigo-100 text-indigo-700 border-indigo-200',
    PICKING: 'bg-orange-100 text-orange-700 border-orange-200',
    ON_DELIVERY: 'bg-slate-800 text-white border-slate-900',
    WAITING_CROSSCHECK: 'bg-amber-100 text-amber-800 border-amber-300',
    COMPLETED: 'bg-emerald-100 text-emerald-700 border-emerald-200',
  }
  return map[status] || 'bg-slate-100 text-slate-700'
}

const getStatusLabel = (status: string) => {
  const map: Record<string, string> = {
    WAITING_APPROVAL: 'MENUNGGU PERSETUJUAN',
    SUPERVISOR_APPROVED: 'DISETUJUI SUPERVISOR',
    PENDING: 'DIPROSES ADMIN',
    PICKING: 'PENGAMBILAN BARANG',
    ON_DELIVERY: 'DALAM PENGIRIMAN',
    WAITING_CROSSCHECK: 'CROSS-CHECK KASIR (SELISIH)',
    COMPLETED: 'SELESAI',
  }
  return map[status] || status
}

const openDetail = (batch: any) => {
  selectedBatch.value = batch
}

const isReportingMode = ref(false)
const reportForm = ref<any[]>([])
const storeSearchQuery = ref('')
const forceInvoiceSearchQuery = ref('')

const filteredReportItems = computed(() => {
  if (!selectedBatch.value?.items) return []
  const items = selectedBatch.value.items.map((item: any, idx: number) => ({
    item,
    originalIndex: idx
  }))
  if (!storeSearchQuery.value.trim()) return items

  const q = storeSearchQuery.value.toLowerCase().trim()
  return items.filter(({ item }: { item: any }) => {
    const storeName = (item.sales_order?.store?.name || item.sales_transaction?.store?.name || '').toLowerCase()
    const storeAddress = (item.sales_order?.store?.address || item.sales_transaction?.store?.address || '').toLowerCase()
    const soNumber = (item.sales_order?.so_number || item.sales_transaction?.invoice_no || '').toLowerCase()
    return storeName.includes(q) || storeAddress.includes(q) || soNumber.includes(q)
  })
})

const startReporting = () => {
  if (!selectedBatch.value) return
  isReportingMode.value = true
  storeSearchQuery.value = ''
  reportForm.value = selectedBatch.value.items.map((item: any) => {
    let returnedItems = []
    const soItems = item.sales_order?.items || item.sales_transaction?.items || []
    returnedItems = soItems.map((si: any) => ({
      product_id: si.product_id,
      product: si.product,
      ordered_quantity: si.ordered_quantity || si.quantity,
      price: si.price,
      retur_tukar_qty: 0,
      retur_potong_qty: 0,
      reason: '',
      condition: 'DAMAGED', // 'GOOD' | 'DAMAGED'
      is_replace_product: false,
      replacement_product_id: '',
      replacement_quantity: 0,
      replacement_price: 0
    }))

    const defaultTotal = item.sales_order?.total_amount || 0

    return {
      item_id: item.id,
      sales_order_id: item.sales_order?.id || null,
      status: item.status === 'DELIVERED' ? 'DELIVERED' : 'DELIVERED', // Default terkirim
      notes: '',
      payment_method: item.sales_order?.payment_method || 'CASH',
      default_total: defaultTotal,
      total_amount: defaultTotal,
      paid_amount: defaultTotal,
      invoice_notes: '',
      returned_items: returnedItems
    }
  })
}

const cancelReporting = () => {
  isReportingMode.value = false
  storeSearchQuery.value = ''
  reportForm.value = []
}

const onReplacementProductChange = (prodItem: any, idx: number) => {
  if (!prodItem.replacement_product_id) {
    prodItem.replacement_price = 0
    prodItem.replacement_quantity = 0
  } else {
    const selectedProd = products.value.find((p: any) => p.id === prodItem.replacement_product_id)
    if (selectedProd) {
      prodItem.replacement_price = selectedProd.selling_price || selectedProd.price || 0
      if (!prodItem.replacement_quantity || prodItem.replacement_quantity <= 0) {
        prodItem.replacement_quantity = prodItem.retur_potong_qty || prodItem.retur_tukar_qty || 1
      }
    }
  }
  updateTotalAmount(idx)
}

const toggleReplaceProduct = (prodItem: any, idx: number) => {
  prodItem.is_replace_product = !prodItem.is_replace_product
  if (!prodItem.is_replace_product) {
    prodItem.replacement_product_id = ''
    prodItem.replacement_quantity = 0
    prodItem.replacement_price = 0
  } else {
    // Jika ganti barang atas keinginan pelanggan (bukan rusak), default condition menjadi GOOD
    prodItem.condition = 'GOOD'
  }
  updateTotalAmount(idx)
}

const updateTotalAmount = (idx: number) => {
  const rf = reportForm.value[idx]
  if (!rf) return
  const deduction = rf.returned_items.reduce((acc: number, ri: any) => acc + (Number(ri.retur_potong_qty) * (ri.price || 0)), 0)
  const addition = rf.returned_items.reduce((acc: number, ri: any) => {
    if (ri.is_replace_product && ri.replacement_product_id && ri.replacement_quantity > 0) {
      return acc + (Number(ri.replacement_quantity) * (ri.replacement_price || 0))
    }
    return acc
  }, 0)

  rf.total_amount = Math.max(0, rf.default_total - deduction + addition)
  if (rf.payment_method !== 'TEMPO') {
    rf.paid_amount = rf.total_amount
  }
}

const submitReport = async () => {
  if (!selectedBatch.value || isSubmitting.value) return
  isSubmitting.value = true
  try {
    const payload = {
      items: reportForm.value.map(rf => ({
        item_id: rf.item_id,
        status: rf.status,
        notes: rf.notes,
        payment_method: rf.payment_method || 'CASH',
        total_amount: Number(rf.total_amount || 0),
        paid_amount: Number(rf.paid_amount || 0),
        invoice_notes: rf.invoice_notes || '',
        returned_items: rf.returned_items
          .filter((ri: any) => ri.retur_tukar_qty > 0 || ri.retur_potong_qty > 0 || (ri.is_replace_product && ri.replacement_product_id))
          .map((ri: any) => ({
            product_id: ri.product_id,
            retur_tukar_qty: Number(ri.retur_tukar_qty || 0),
            retur_potong_qty: Number(ri.retur_potong_qty || 0),
            reason: ri.reason || (ri.is_replace_product ? 'Ganti produk pesanan' : ''),
            condition: ri.condition || 'DAMAGED',
            replacement_product_id: ri.is_replace_product && ri.replacement_product_id ? ri.replacement_product_id : null,
            replacement_quantity: ri.is_replace_product && ri.replacement_product_id ? Number(ri.replacement_quantity || 0) : 0,
            replacement_price: ri.is_replace_product && ri.replacement_product_id ? Number(ri.replacement_price || 0) : 0
          }))
      }))
    }
    await apiClient.post(`/delivery/batch/${selectedBatch.value.id}/admin-report`, payload)
    toast.success('Laporan pengiriman & Faktur berhasil disimpan!')

    isReportingMode.value = false
    storeSearchQuery.value = ''
    selectedBatch.value = null
    fetchData()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menyimpan laporan & faktur')
  } finally {
    isSubmitting.value = false
  }

}

watch(selectedBatch, () => {
  isReportingMode.value = false
  storeSearchQuery.value = ''
  reportForm.value = []
})

// ─── Force Invoice: Buat Faktur dari SJ COMPLETED (Daily Driver) ─────────────
const showForceInvoiceModal = ref(false)
const isCreatingInvoice = ref(false)

// Form input pembayaran per nota — diisi admin nota berdasarkan laporan kertas driver
const forceInvoiceForm = ref<{
  sales_order_id: string
  so_number: string
  store_name: string
  default_amount: number
  payment_method: string
  total_amount: number
  paid_amount: number   // Jumlah yang dibayar driver saat pengiriman (penting untuk TEMPO)
  notes: string
}[]>([])

const filteredForceInvoiceForm = computed(() => {
  const items = forceInvoiceForm.value.map((f, idx) => ({ form: f, originalIndex: idx }))
  if (!forceInvoiceSearchQuery.value.trim()) return items
  const q = forceInvoiceSearchQuery.value.toLowerCase().trim()
  return items.filter(({ form }) => {
    return (form.store_name || '').toLowerCase().includes(q) ||
      (form.so_number || '').toLowerCase().includes(q)
  })
})

const openForceInvoiceModal = () => {
  if (!selectedBatch.value) return
  forceInvoiceSearchQuery.value = ''
  // Bangun form dari items SJ
  forceInvoiceForm.value = (selectedBatch.value.items || []).flatMap((item: any) => {
    if (!item.sales_order) return []
    const amount = item.sales_order.total_amount || 0
    return [{
      sales_order_id: item.sales_order.id,
      so_number: item.sales_order.so_number || '—',
      store_name: item.sales_order.store?.name || 'Toko tidak diketahui',
      default_amount: amount,
      payment_method: item.sales_order.payment_method || 'CASH',
      total_amount: amount,
      paid_amount: amount,   // Default: bayar lunas
      notes: ''
    }]
  })
  showForceInvoiceModal.value = true
}

const submitForceInvoice = async () => {
  if (!selectedBatch.value) return
  isCreatingInvoice.value = true
  try {
    const resolvedCompanyId = masterStore.selectedBranchCompanyId || selectedBatch.value?.company_id || selectedBatch.value?.items?.[0]?.sales_order?.company_id || undefined
    const payload: any = {
      items: forceInvoiceForm.value.map(f => ({
        sales_order_id: f.sales_order_id,
        payment_method: f.payment_method,
        total_amount: Number(f.total_amount),
        paid_amount: Number(f.paid_amount),
        notes: f.notes
      }))
    }
    if (resolvedCompanyId) {
      payload.company_id = resolvedCompanyId
    }
    await apiClient.post(`/delivery/batch/${selectedBatch.value.id}/force-invoice`, payload)
    toast.success('Faktur berhasil dibuat! Silakan approve di menu Faktur Penjualan.')
    showForceInvoiceModal.value = false
    selectedBatch.value = null
    fetchData()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal membuat faktur')
  } finally {
    isCreatingInvoice.value = false
  }
}
// ─────────────────────────────────────────────────────────────────────────────

// ─── Revise Qty (Admin Nota) sebelum jalan ─────────────
const showReviseQtyModal = ref(false)
const isRevisingQty = ref(false)
const reviseForm = ref({
  sales_order_id: '',
  so_number: '',
  items: [] as any[]
})

const openReviseQtyModal = (item: any) => {
  if (!item.sales_order) return
  reviseForm.value.sales_order_id = item.sales_order.id
  reviseForm.value.so_number = item.sales_order.so_number || '—'
  reviseForm.value.items = (item.sales_order.items || []).map((si: any) => ({
    sales_order_item_id: si.id,
    product_name: si.product?.name || 'Produk',
    price: si.price,
    ordered_quantity: si.ordered_quantity || si.quantity,
    new_quantity: si.ordered_quantity || si.quantity
  }))
  showReviseQtyModal.value = true
}

const submitReviseQty = async () => {
  if (!selectedBatch.value) return
  isRevisingQty.value = true
  try {
    const payload = {
      sales_order_id: reviseForm.value.sales_order_id,
      items: reviseForm.value.items.map(ri => ({
        sales_order_item_id: ri.sales_order_item_id,
        new_quantity: Number(ri.new_quantity)
      }))
    }
    await apiClient.put(`/delivery/batch/${selectedBatch.value.id}/revise-qty`, payload)
    toast.success('Kuantitas pesanan berhasil disesuaikan!')
    showReviseQtyModal.value = false
    selectedBatch.value = null
    fetchData()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menyesuaikan kuantitas')
  } finally {
    isRevisingQty.value = false
  }
}
// ─────────────────────────────────────────────────────────────────────────────

const columns = [
  { key: 'delivery_order_no', label: 'NO SURAT JALAN' },
  { key: 'driver', label: 'DRIVER' },
  { key: 'status', label: 'STATUS' },
  { key: 'actions', label: 'AKSI' },
]
</script>

<template>
  <div class="space-y-8 animate-in fade-in duration-700 print:hidden">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row md:items-center justify-between gap-6 bg-white p-8 rounded-[40px] border border-slate-200/60 shadow-sm shadow-slate-100/50"
    >
      <div class="flex items-center gap-5">
        <div>
          <h1 class="text-3xl font-black text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
            <span>Data Surat Jalan</span>
            <span v-if="currentBranchName" class="text-primary text-2xl font-bold">
              — {{ currentBranchName }}
            </span>
          </h1>
          <p class="text-slate-500 font-bold text-sm mt-1">
            Daftar dokumen pengiriman yang telah diterbitkan
          </p>
        </div>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <RefreshDataButton
          :loading="isRefreshing || isLoading || deliveryStore.loading"
          :last-updated="lastRefreshedTimeStr"
          @click="triggerRefresh()"
        />
        <Button
          @click="$router.push('/sales/delivery')"
          class="bg-primary text-white rounded-2xl h-12 px-6 font-black shadow-lg shadow-primary/20 flex items-center gap-2"
        >
          <PlusCircle class="w-5 h-5" /> BUAT SURAT JALAN
        </Button>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
      <!-- Kiri: Filter & Tabel SJ -->
      <div class="lg:col-span-5 space-y-6">
        <!-- Filters -->
        <div class="bg-white p-6 rounded-[32px] border border-slate-200/60 shadow-sm flex flex-col gap-4">
          <div class="flex flex-col gap-4 items-center">
            <div class="relative w-full">
              <Search class="absolute left-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Cari No. SJ..."
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl pl-14 pr-6 py-3 text-sm font-bold text-slate-900 outline-none transition-all placeholder:text-slate-400 shadow-inner"
              />
            </div>
            <select
              v-model="selectedStatus"
              class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-6 py-3 text-sm font-bold text-slate-900 outline-none transition-all shadow-inner"
            >
              <option value="ALL">Semua Status</option>
              <option value="WAITING_APPROVAL">Menunggu Persetujuan</option>
              <option value="SUPERVISOR_APPROVED">Disetujui Supervisor</option>
              <option value="PENDING">Diproses Admin</option>
              <option value="ON_DELIVERY">Dalam Pengiriman</option>
              <option value="WAITING_CROSSCHECK">Perlu Cross-check Kasir (Selisih)</option>
              <option value="COMPLETED">Selesai</option>
            </select>
          </div>
        </div>

        <!-- Main Table -->
        <div class="bg-white border border-slate-200/60 rounded-[40px] overflow-hidden shadow-xl shadow-slate-200/20">
          <DataTable :columns="columns" :data="filteredBatches">
            <template #cell-delivery_order_no="{ row }">
              <div class="flex flex-col items-start gap-1">
                <span class="text-xs font-black text-slate-900">{{ row.delivery_order_no || 'DRAFT-SJ' }}</span>
                <span v-if="row.notes && row.notes.includes('Motoris')" class="bg-indigo-100 text-indigo-700 text-[9px] px-2 py-0.5 rounded-md font-black tracking-widest uppercase">Motoris</span>
              </div>
            </template>

            <template #cell-driver="{ row }">
              <div class="flex flex-col">
                <span class="text-xs font-bold text-slate-900">
                  {{ row.driver?.first_name || row.daily_driver?.name || 'Belum Assign' }}
                  <span v-if="row.helper || row.daily_helper" class="block text-[10px] text-slate-500 font-bold mt-0.5">
                    + Helper: {{ row.helper?.first_name || row.daily_helper?.name }}
                  </span>
                </span>
              </div>
            </template>

            <template #cell-status="{ row }">
              <Badge
                :class="`${getStatusColor(row.status)} rounded-full px-2 py-1 font-black text-[8px] uppercase tracking-widest border`"
                variant="outline"
              >
                {{ getStatusLabel(row.status) }}
              </Badge>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  @click="openDetail(row)"
                  :class="selectedBatch?.id === row.id ? 'bg-primary text-white hover:bg-primary/90' : 'text-slate-600 hover:bg-slate-50'"
                  class="rounded-xl font-black text-[10px] px-2.5"
                >
                  <ArrowRight v-if="selectedBatch?.id !== row.id" class="w-3.5 h-3.5" />
                  <span v-else>PILIH</span>
                </Button>
                <Button
                  v-if="!['COMPLETED', 'WAITING_CROSSCHECK'].includes(row.status)"
                  variant="ghost"
                  size="sm"
                  title="Susulkan Nota"
                  @click="openAppendModal(row)"
                  class="text-amber-600 hover:bg-amber-50 rounded-xl font-black text-[10px] px-2"
                >
                  <PlusCircle class="w-3.5 h-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  @click="openPrintPreview(row)"
                  class="text-primary hover:bg-primary/5 rounded-xl font-black text-[10px] px-2.5"
                >
                  <Printer class="w-3.5 h-3.5" />
                </Button>
              </div>
            </template>
          </DataTable>

          <div v-if="filteredBatches.length === 0" class="py-12 text-center">
            <h3 class="text-sm font-black text-slate-900">Data Tidak Ditemukan</h3>
          </div>
        </div>
      </div>

      <!-- Kanan: Detail Panel -->
      <div class="lg:col-span-7 space-y-6 lg:sticky lg:top-8">
        <div v-if="!selectedBatch" class="bg-white border border-slate-200/60 rounded-[40px] p-12 text-center shadow-xl shadow-slate-200/20 flex flex-col items-center justify-center min-h-[50vh]">
          <div class="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center mb-4">
            <Truck class="w-10 h-10 text-slate-300" />
          </div>
          <h3 class="text-xl font-black text-slate-900">Pilih Surat Jalan</h3>
          <p class="text-slate-500 font-bold mt-2">Klik tombol PILIH pada tabel di sebelah kiri untuk melihat detail Surat Jalan</p>
        </div>

        <div v-else class="bg-white border border-slate-200/60 rounded-[40px] shadow-xl shadow-slate-200/20 overflow-hidden flex flex-col">
          <div class="p-8 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
            <div class="flex items-center gap-4">
              <div
                class="w-16 h-16 rounded-2xl bg-white flex items-center justify-center border border-slate-200 shadow-sm overflow-hidden shrink-0 cursor-zoom-in hover:border-primary transition-colors"
                @click="zoomedQR = selectedBatch?.delivery_order_no"
              >
                <img
                  :src="`https://api.qrserver.com/v1/create-qr-code/?size=100x100&margin=0&data=${selectedBatch?.delivery_order_no}`"
                  class="w-12 h-12"
                />
              </div>
              <div>
                <h3 class="text-2xl font-black text-slate-900">Detail Surat Jalan</h3>
                <p class="text-xs text-slate-500 font-bold uppercase tracking-widest">
                  {{ selectedBatch?.delivery_order_no }}
                </p>
              </div>
            </div>
            <div class="flex gap-2">
              <Button
                v-if="canEditOrDeleteBatch(selectedBatch.status)"
                variant="outline"
                size="icon"
                @click="openEditModal(selectedBatch)"
                class="rounded-xl border-2 text-indigo-600 hover:bg-indigo-50 hover:border-indigo-200"
              >
                <Pencil class="w-5 h-5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                @click="selectedBatch = null"
                class="rounded-full hover:bg-slate-100"
              >
                <X class="w-6 h-6" />
              </Button>
            </div>
          </div>

          <div class="p-8 overflow-y-auto space-y-8 flex-1 bg-slate-50/20 max-h-[60vh]">
            <!-- Warning Banner Cross-check Kasir -->
            <div v-if="selectedBatch?.status === 'WAITING_CROSSCHECK'" class="bg-amber-50/90 border-2 border-amber-300 p-5 rounded-3xl flex items-start gap-4 shadow-sm animate-pulse">
              <div class="w-10 h-10 rounded-2xl bg-amber-500/20 text-amber-700 flex items-center justify-center shrink-0">
                <AlertTriangle class="w-5 h-5" />
              </div>
              <div class="space-y-1.5 flex-1">
                <div class="flex items-center justify-between">
                  <h5 class="text-xs font-black text-amber-900 uppercase tracking-wider">Perhatian: Surat Jalan Perlu Penyesuaian (Cross-check Kasir)</h5>
                  <span class="bg-amber-200 text-amber-900 text-[9px] font-black px-2.5 py-0.5 rounded-full uppercase tracking-wider">Selisih Kasir</span>
                </div>
                <p class="text-xs text-amber-800 font-medium">
                  Kasir mendeteksi adanya selisih kas fisik atau retur barang pada surat jalan ini. Silakan sesuaikan rincian tagihan/nota menggunakan tombol <strong>Buat / Lanjutkan Faktur</strong> di bawah.
                </p>
                <div v-if="selectedBatch.notes" class="text-[11px] font-bold text-amber-900 bg-amber-200/60 px-3.5 py-2 rounded-xl mt-2 border border-amber-300/60">
                  {{ selectedBatch.notes }}
                </div>
              </div>
            </div>

            <!-- Info Cards -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div class="bg-white p-4 rounded-3xl border border-slate-200 shadow-sm space-y-2">
                <div class="flex items-center gap-2 text-slate-400">
                  <User class="w-3.5 h-3.5" />
                  <span class="text-[9px] font-black uppercase tracking-widest">Supir</span>
                </div>
                <p class="text-sm font-black text-slate-900 truncate">
                  {{ selectedBatch?.driver
                      ? `${selectedBatch.driver.first_name} ${selectedBatch.driver.last_name}`
                      : (selectedBatch?.daily_driver?.name || '—') }}
                </p>
                <p v-if="selectedBatch?.daily_driver" class="text-[9px] font-bold text-amber-600 uppercase tracking-wide">Pekerja Harian</p>
              </div>
              <div v-if="selectedBatch?.helper || selectedBatch?.daily_helper" class="bg-white p-4 rounded-3xl border border-slate-200 shadow-sm space-y-2">
                <div class="flex items-center gap-2 text-slate-400">
                  <User class="w-3.5 h-3.5" />
                  <span class="text-[9px] font-black uppercase tracking-widest">Helper</span>
                </div>
                <p class="text-sm font-black text-slate-900 truncate">
                  {{ selectedBatch?.helper
                      ? `${selectedBatch.helper.first_name} ${selectedBatch.helper.last_name}`
                      : (selectedBatch?.daily_helper?.name || '—') }}
                </p>
                <p v-if="selectedBatch?.daily_helper" class="text-[9px] font-bold text-amber-600 uppercase tracking-wide">Pekerja Harian</p>
              </div>
              <div class="bg-white p-4 rounded-3xl border border-slate-200 shadow-sm space-y-2">
                <div class="flex items-center gap-2 text-slate-400">
                  <Car class="w-3.5 h-3.5" />
                  <span class="text-[9px] font-black uppercase tracking-widest">Armada</span>
                </div>
                <p class="text-sm font-black text-slate-900">
                  {{ selectedBatch?.vehicle ? selectedBatch.vehicle.license_plate : '—' }}
                </p>
              </div>
              <div class="bg-white p-4 rounded-3xl border border-slate-200 shadow-sm space-y-2">
                <div class="flex items-center gap-2 text-slate-400">
                  <Clock class="w-3.5 h-3.5" />
                  <span class="text-[9px] font-black uppercase tracking-widest">Status Terakhir</span>
                </div>
                <Badge :class="`${getStatusColor(selectedBatch?.status)} rounded-full px-3 py-1 font-black text-[8px] uppercase tracking-widest`" variant="outline">
                  {{ getStatusLabel(selectedBatch?.status) }}
                </Badge>
              </div>
            </div>

            <!-- Header Action Lapor Pengiriman -->
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-black text-slate-900 uppercase tracking-widest">
                Daftar Nota
              </h4>
              <div v-if="!['COMPLETED', 'WAITING_CROSSCHECK'].includes(selectedBatch.status) && !isReportingMode" class="flex gap-2">
                <Button
                  @click="openAppendModal(selectedBatch)"
                  class="bg-amber-500 hover:bg-amber-600 text-white rounded-xl font-bold shadow-md shadow-amber-200"
                  size="sm"
                >
                  <PlusCircle class="w-4 h-4 mr-1.5" /> Susulkan Nota
                </Button>
                <Button
                  v-if="['SUPERVISOR_APPROVED', 'PENDING', 'PICKING'].includes(selectedBatch.status)"
                  @click="confirmStartDelivery"
                  :disabled="isStartingDelivery"
                  class="bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold shadow-md shadow-emerald-200"
                  size="sm"
                >
                  <Truck class="w-4 h-4 mr-1.5" /> {{ isStartingDelivery ? 'Memproses...' : 'Mulai Kirim (Potong Stok)' }}
                </Button>
                <Button
                  @click="startReporting"
                  class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-md shadow-indigo-200"
                  size="sm"
                >
                  Lapor Pengiriman (Admin Nota)
                </Button>
              </div>
              <!-- Tombol Buat Faktur: khusus SJ COMPLETED (daily driver) -->
              <div v-else-if="['COMPLETED', 'WAITING_CROSSCHECK'].includes(selectedBatch.status) && !isReportingMode" class="flex gap-2">
                <Button
                  @click="openForceInvoiceModal"
                  class="bg-violet-600 hover:bg-violet-700 text-white rounded-xl font-bold shadow-md shadow-violet-200"
                  size="sm"
                >
                  <FileText class="w-4 h-4 mr-1.5" /> Buat / Lanjutkan Faktur
                </Button>
              </div>
              <div v-else-if="isReportingMode" class="flex gap-2">
                <Button variant="outline" size="sm" @click="cancelReporting" class="rounded-xl font-bold">Batal</Button>
                <Button size="sm" class="bg-primary hover:bg-primary/90 text-white rounded-xl font-bold" @click="submitReport" :disabled="isSubmitting">
                  {{ isSubmitting ? 'Menyimpan...' : 'Simpan & Selesai' }}
                </Button>
              </div>
            </div>

            <!-- Search Toko / Nota Filter Bar -->
            <div class="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between bg-white p-3.5 rounded-2xl border border-slate-200 shadow-sm">
              <div class="relative flex-1">
                <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <input
                  v-model="storeSearchQuery"
                  type="text"
                  placeholder="Cari nama toko, alamat, atau no. SO..."
                  class="w-full pl-9 pr-9 py-2 text-xs font-bold rounded-xl border border-slate-200 bg-slate-50/50 focus:bg-white focus:border-indigo-500 focus:ring-2 focus:ring-indigo-100 outline-none transition-all"
                />
                <button
                  v-if="storeSearchQuery"
                  @click="storeSearchQuery = ''"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5 rounded-full hover:bg-slate-100"
                  type="button"
                >
                  <X class="w-3.5 h-3.5" />
                </button>
              </div>
              <div class="text-[10px] font-black text-slate-500 shrink-0 self-center">
                <span class="bg-slate-100 text-slate-600 px-2.5 py-1 rounded-lg">
                  {{ filteredReportItems.length }} dari {{ selectedBatch?.items?.length || 0 }} Nota
                </span>
              </div>
            </div>

            <!-- Items List -->
            <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
              <div class="divide-y divide-slate-100">
                <div
                  v-for="{ item, originalIndex } in filteredReportItems"
                  :key="item.id"
                  class="p-4 md:p-6 flex flex-col gap-4 hover:bg-slate-50/50 transition-colors"
                >
                  <div class="flex items-start gap-4">
                    <div class="w-10 h-10 rounded-xl bg-slate-100 border border-slate-200 flex items-center justify-center font-black text-slate-400 shrink-0 shadow-inner text-sm">
                      {{ originalIndex + 1 }}
                    </div>
                    <div class="flex-1">
                      <div class="flex items-center gap-2">
                        <h5 class="text-sm md:text-base font-black text-slate-900">
                          {{ item.sales_order?.so_number || item.sales_transaction?.invoice_no || 'NOT-SET' }}
                        </h5>
                        <Badge
                          v-if="item.is_subsequent"
                          class="bg-amber-100 text-amber-800 border-amber-300 font-black text-[9px] px-2 py-0.5 rounded-md"
                          variant="outline"
                        >
                          SUSULAN
                        </Badge>
                      </div>
                      <div class="flex items-start gap-1.5 mt-1">
                        <MapPin class="w-3 h-3 text-slate-400 mt-0.5 shrink-0" />
                        <div class="flex flex-col">
                          <span class="text-xs font-black text-slate-700 leading-none">{{ item.sales_order?.store?.name || item.sales_transaction?.store?.name || 'Toko tidak terdaftar' }}</span>
                          <span class="text-[9px] font-bold text-slate-400 mt-1 leading-normal">{{ item.sales_order?.store?.address || item.sales_transaction?.store?.address || 'Alamat tidak diatur' }}</span>
                        </div>
                      </div>
                      <div v-if="item.notes" class="text-[10px] text-amber-700 font-bold bg-amber-50 px-2 py-0.5 rounded mt-1.5 inline-block border border-amber-200/60">
                        {{ item.notes }}
                      </div>
                    </div>
                    
                    <div v-if="!isReportingMode" class="flex flex-col items-end gap-1">
                      <Badge
                        :class="`${item.status === 'DELIVERED' ? 'bg-emerald-100 text-emerald-700' : item.status === 'FAILED' ? 'bg-red-100 text-red-700' : 'bg-slate-200 text-slate-600'} rounded-lg px-2 py-0.5 font-black text-[9px]`"
                        variant="outline"
                      >
                        {{ item.status === 'DELIVERED' ? 'DITERIMA' : item.status === 'FAILED' ? 'GAGAL' : 'PENDING' }}
                      </Badge>
                      <span class="text-xs font-black text-slate-900 mt-1">
                        Rp {{ (item.sales_order?.total_amount || item.sales_transaction?.total_amount || 0).toLocaleString('id-ID') }}
                      </span>
                      <Button
                        v-if="['PENDING', 'SUPERVISOR_APPROVED'].includes(selectedBatch.status) && item.sales_order"
                        variant="outline"
                        size="sm"
                        @click="openReviseQtyModal(item)"
                        class="text-[10px] h-7 px-2 font-bold text-slate-600 mt-1"
                      >
                        <Pencil class="w-3 h-3 mr-1" /> Sesuaikan Qty
                      </Button>
                    </div>
                  </div>

                  <div v-if="hasDiscrepancy(item)" class="mt-2 bg-amber-50 rounded-xl p-3 border border-amber-200/60 flex items-start gap-2">
                    <AlertTriangle class="w-4 h-4 text-amber-600 mt-0.5 shrink-0" />
                    <div class="flex flex-col">
                      <span class="text-[10px] font-black text-amber-800 uppercase tracking-wide">Dilaporkan Selisih oleh Kasir</span>
                      <span class="text-xs font-medium text-amber-700 mt-0.5">{{ getDiscrepancyNote(item) }}</span>
                    </div>
                  </div>

                  <div v-if="isReportingMode" class="mt-2 p-4 bg-slate-50 rounded-2xl border border-slate-200 flex flex-col gap-4">
                    <div class="flex flex-col sm:flex-row gap-3 items-end">
                      <div class="flex-1 w-full">
                        <label class="text-[10px] font-black text-slate-500 uppercase">Status Pengiriman</label>
                        <select
                          v-if="reportForm[originalIndex]"
                          v-model="reportForm[originalIndex].status"
                          class="w-full bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20 mt-1"
                        >
                          <option value="DELIVERED">Terkirim</option>
                          <option value="FAILED">Gagal Kirim (Retur Semua)</option>
                        </select>
                      </div>
                      <div class="flex-1 w-full">
                        <label class="text-[10px] font-black text-slate-500 uppercase">Catatan Tambahan</label>
                        <input
                          v-if="reportForm[originalIndex]"
                          v-model="reportForm[originalIndex].notes"
                          type="text"
                          placeholder="Kondisi pengiriman..."
                          class="w-full bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-bold outline-none focus:ring-2 focus:ring-primary/20 mt-1"
                        />
                      </div>
                    </div>

                    <!-- List Produk untuk input retur -->
                    <div v-if="reportForm[originalIndex]?.status === 'DELIVERED' && reportForm[originalIndex]?.returned_items?.length > 0" class="mt-2 flex flex-col gap-3">
                      <div class="flex items-center justify-between border-b pb-1">
                        <label class="text-[10px] font-black text-slate-500 uppercase">Detail Barang & Retur/Ganti Barang</label>
                        <span class="text-[9px] text-slate-400 font-bold">Isi jika ada retur, potong nota, atau tukar barang</span>
                      </div>

                      <div v-for="(prod, pIdx) in reportForm[originalIndex].returned_items" :key="pIdx" class="bg-white p-3.5 rounded-2xl border border-slate-200 flex flex-col gap-3">
                        <div class="grid grid-cols-1 sm:grid-cols-12 gap-3 items-center">
                          <div class="sm:col-span-4 flex flex-col">
                            <span class="text-[11px] font-black text-slate-900 leading-tight">{{ prod.product?.name || 'Produk' }}</span>
                            <span class="text-[9px] font-bold text-slate-500 mt-0.5">Pesan: {{ prod.ordered_quantity }} Pcs @ Rp {{ (prod.price || 0).toLocaleString('id-ID') }}</span>
                          </div>
                          
                          <div class="sm:col-span-2">
                            <label class="text-[9px] font-bold text-slate-500 block mb-1">Tukar Sejenis</label>
                            <input v-model.number="prod.retur_tukar_qty" type="number" min="0" :max="prod.ordered_quantity" class="w-full bg-slate-50 border border-slate-200 rounded-lg px-2 py-1.5 text-xs font-bold outline-none focus:border-primary/50 text-center" />
                          </div>

                          <div class="sm:col-span-2">
                            <label class="text-[9px] font-bold text-red-500 block mb-1">Potong / Diganti</label>
                            <input v-model.number="prod.retur_potong_qty" @input="updateTotalAmount(originalIndex)" type="number" min="0" :max="prod.ordered_quantity" class="w-full bg-red-50 border border-red-200 rounded-lg px-2 py-1.5 text-xs font-bold outline-none focus:border-red-400 text-center text-red-700" />
                          </div>

                          <div class="sm:col-span-4">
                            <label class="text-[9px] font-bold text-slate-500 block mb-1">Alasan / Catatan</label>
                            <input v-model="prod.reason" type="text" placeholder="Cth: Rusak, ingin ganti produk..." class="w-full bg-slate-50 border border-slate-200 rounded-lg px-2 py-1.5 text-xs font-bold outline-none focus:border-primary/50" />
                          </div>
                        </div>

                        <!-- Baris Tambahan: Kondisi Barang & Toggle Ganti Produk Beda Jenis (Jika ada retur/potong) -->
                        <div v-if="prod.retur_potong_qty > 0 || prod.retur_tukar_qty > 0 || prod.is_replace_product" class="pt-2 border-t border-slate-100 flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between bg-slate-50/60 p-2.5 rounded-xl">
                          <div class="flex items-center gap-3">
                            <span class="text-[9px] font-black text-slate-400 uppercase tracking-wider">Kondisi Barang Ditarik:</span>
                            <div class="flex items-center gap-2">
                              <label class="flex items-center gap-1 cursor-pointer">
                                <input type="radio" :name="`cond-${originalIndex}-${pIdx}`" value="GOOD" v-model="prod.condition" class="text-emerald-600 focus:ring-emerald-500 w-3.5 h-3.5" />
                                <span class="text-[10px] font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">🟢 Masih Bagus</span>
                              </label>
                              <label class="flex items-center gap-1 cursor-pointer">
                                <input type="radio" :name="`cond-${originalIndex}-${pIdx}`" value="DAMAGED" v-model="prod.condition" class="text-red-600 focus:ring-red-500 w-3.5 h-3.5" />
                                <span class="text-[10px] font-bold text-red-700 bg-red-50 px-2 py-0.5 rounded border border-red-200">🔴 Rusak / Cacat</span>
                              </label>
                            </div>
                          </div>

                          <!-- Toggle Ganti Produk Beda Jenis -->
                          <div class="flex items-center gap-2">
                            <button
                              type="button"
                              @click="toggleReplaceProduct(prod, originalIndex)"
                              :class="prod.is_replace_product ? 'bg-violet-600 text-white border-violet-600' : 'bg-white text-slate-700 border-slate-200 hover:bg-slate-50'"
                              class="text-[10px] font-black px-3 py-1 rounded-lg border flex items-center gap-1.5 transition-all shadow-sm"
                            >
                              <span>🔄</span>
                              <span>{{ prod.is_replace_product ? '✓ Ganti Produk Beda Jenis' : '+ Tukar Produk Lain' }}</span>
                            </button>
                          </div>
                        </div>

                        <!-- Form Detail Produk Pengganti (Jika Toggle Ganti Produk Aktif) -->
                        <div v-if="prod.is_replace_product" class="p-3 bg-violet-50/70 border border-violet-200 rounded-xl space-y-2.5">
                          <div class="flex items-center justify-between">
                            <span class="text-[10px] font-black text-violet-800 uppercase tracking-wide">Pilih Produk Pengganti (Substitusi)</span>
                            <span class="text-[9px] text-violet-600 font-bold">Otomatis menyesuaikan total tagihan faktur</span>
                          </div>
                          <div class="grid grid-cols-1 sm:grid-cols-12 gap-2.5 items-end">
                            <div class="sm:col-span-6">
                              <label class="text-[9px] font-bold text-violet-700 block mb-1">Produk Pengganti</label>
                              <select
                                v-model="prod.replacement_product_id"
                                @change="onReplacementProductChange(prod, originalIndex)"
                                class="w-full text-xs p-2 rounded-lg border border-violet-200 bg-white focus:border-violet-500 outline-none font-bold"
                              >
                                <option value="" disabled>-- Pilih Produk Pengganti --</option>
                                <option v-for="p in products" :key="p.id" :value="p.id">
                                  {{ p.name }} (Rp {{ (p.selling_price || p.price || 0).toLocaleString('id-ID') }})
                                </option>
                              </select>
                            </div>

                            <div class="sm:col-span-2">
                              <label class="text-[9px] font-bold text-violet-700 block mb-1">Qty Pengganti</label>
                              <input
                                v-model.number="prod.replacement_quantity"
                                @input="updateTotalAmount(originalIndex)"
                                type="number"
                                min="1"
                                class="w-full text-xs p-2 rounded-lg border border-violet-200 bg-white focus:border-violet-500 outline-none font-bold text-center"
                              />
                            </div>

                            <div class="sm:col-span-4">
                              <label class="text-[9px] font-bold text-violet-700 block mb-1">Harga Satuan (Rp)</label>
                              <input
                                v-model.number="prod.replacement_price"
                                @input="updateTotalAmount(originalIndex)"
                                type="number"
                                min="0"
                                class="w-full text-xs p-2 rounded-lg border border-violet-200 bg-white focus:border-violet-500 outline-none font-bold text-right"
                              />
                            </div>
                          </div>

                          <div v-if="prod.replacement_product_id && prod.replacement_quantity > 0" class="text-right text-[10px] font-black text-violet-900 pt-1">
                            Nilai Produk Pengganti: Rp {{ (Number(prod.replacement_quantity) * Number(prod.replacement_price || 0)).toLocaleString('id-ID') }}
                          </div>
                        </div>

                      </div>
                      
                      <!-- Estimasi Pemotongan / Perubahan Tagihan -->
                      <div class="bg-indigo-50 border border-indigo-100 rounded-xl p-3 flex justify-between items-center mt-1">
                        <div class="flex flex-col">
                          <span class="text-[10px] font-black text-indigo-700 uppercase tracking-wide">Estimasi Total Tagihan (Sistem)</span>
                          <span class="text-[9px] text-indigo-500 font-medium">Memperhitungkan barang ditarik (-) dan barang pengganti (+)</span>
                        </div>
                        <span class="text-sm font-black text-indigo-900">
                          Rp {{ Number(reportForm[originalIndex].total_amount || 0).toLocaleString('id-ID') }}
                        </span>
                      </div>

                      <!-- Laporan Pembayaran (Faktur) -->
                      <div v-if="reportForm[originalIndex].sales_order_id" class="mt-3 pt-3 border-t border-slate-200">
                        <label class="text-[10px] font-black text-slate-500 uppercase block mb-2">Penagihan & Pembayaran (Faktur)</label>
                        <div class="grid grid-cols-2 gap-3">
                          <div>
                            <label class="text-[9px] font-bold text-slate-400 block mb-1">Metode</label>
                            <select v-model="reportForm[originalIndex].payment_method" class="w-full text-xs p-2 rounded-lg border border-slate-200 focus:border-primary outline-none">
                              <option value="CASH">CASH</option>
                              <option value="TRANSFER">TRANSFER</option>
                              <option value="TEMPO">TEMPO (Kredit)</option>
                            </select>
                          </div>
                          <div>
                            <label class="text-[9px] font-bold text-slate-400 block mb-1">Tagihan Aktual (Rp)</label>
                            <input v-model.number="reportForm[originalIndex].total_amount" @input="reportForm[originalIndex].payment_method !== 'TEMPO' && (reportForm[originalIndex].paid_amount = Number(reportForm[originalIndex].total_amount))" type="number" min="0" class="w-full text-xs p-2 rounded-lg border border-slate-200 focus:border-primary outline-none" />
                          </div>
                          <div>
                            <label class="text-[9px] font-bold text-slate-400 block mb-1">Sudah Dibayar (Rp)</label>
                            <input v-model.number="reportForm[originalIndex].paid_amount" type="number" min="0" class="w-full text-xs p-2 rounded-lg border border-slate-200 focus:border-primary outline-none" />
                          </div>
                          <div>
                            <label class="text-[9px] font-bold text-slate-400 block mb-1">Keterangan Nota</label>
                            <input v-model="reportForm[originalIndex].invoice_notes" type="text" placeholder="Catatan tagihan..." class="w-full text-xs p-2 rounded-lg border border-slate-200 focus:border-primary outline-none" />
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                  
                </div>

                <!-- Empty State jika pencarian toko tidak ditemukan -->
                <div v-if="filteredReportItems.length === 0" class="p-8 text-center flex flex-col items-center justify-center">
                  <div class="w-12 h-12 bg-slate-100 rounded-2xl flex items-center justify-center text-slate-400 mb-3">
                    <Search class="w-6 h-6" />
                  </div>
                  <p class="text-sm font-bold text-slate-700">Toko atau nota tidak ditemukan</p>
                  <p class="text-xs text-slate-400 mt-1">Tidak ada hasil yang sesuai dengan kata kunci "{{ storeSearchQuery }}"</p>
                  <Button
                    variant="outline"
                    size="sm"
                    @click="storeSearchQuery = ''"
                    class="mt-4 rounded-xl text-xs font-bold"
                  >
                    Reset Pencarian
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal: Buat Faktur dari SJ Daily Driver (Force Invoice) -->
    <div
      v-if="showForceInvoiceModal"
      class="fixed inset-0 z-[90] flex items-center justify-center p-4 overflow-y-auto"
    >
      <div class="absolute inset-0 bg-slate-900/70 backdrop-blur-md" @click="showForceInvoiceModal = false" />
      <div
        class="relative z-10 w-full max-w-2xl bg-white rounded-[40px] shadow-2xl overflow-hidden animate-in zoom-in-95 duration-300"
      >
        <!-- Header -->
        <div class="bg-gradient-to-r from-violet-600 to-violet-700 p-7 text-white">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-white/20 rounded-2xl flex items-center justify-center shrink-0">
              <FileText class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-xl font-black tracking-tight">Buat Faktur Penjualan</h3>
              <p class="text-violet-200 text-xs font-bold mt-0.5 uppercase tracking-widest">
                {{ selectedBatch?.delivery_order_no }} · Driver Harian
              </p>
            </div>
          </div>
          <p class="mt-4 text-sm text-violet-100 font-medium leading-relaxed">
            Masukkan metode pembayaran dan total tagihan aktual per nota berdasarkan
            <span class="font-black">laporan kertas harian driver</span>. Data ini akan digunakan untuk membuat Faktur Penjualan yang siap disetujui Admin Nota.
          </p>
        </div>

        <!-- Search Filter Toko pada Modal Faktur -->
        <div v-if="forceInvoiceForm.length > 0" class="px-6 pt-4 pb-0">
          <div class="relative">
            <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              v-model="forceInvoiceSearchQuery"
              type="text"
              placeholder="Cari toko atau nomor SO..."
              class="w-full pl-9 pr-9 py-2 text-xs font-bold rounded-xl border border-slate-200 bg-slate-50 focus:bg-white focus:border-violet-500 focus:ring-2 focus:ring-violet-100 outline-none transition-all"
            />
            <button
              v-if="forceInvoiceSearchQuery"
              @click="forceInvoiceSearchQuery = ''"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5 rounded-full hover:bg-slate-100"
              type="button"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- Body: list nota -->
        <div class="p-6 max-h-[55vh] overflow-y-auto space-y-4">
          <div
            v-for="{ form, originalIndex } in filteredForceInvoiceForm"
            :key="form.sales_order_id"
            class="bg-slate-50 border border-slate-200 rounded-3xl p-5 space-y-4"
          >
            <!-- Nota header -->
            <div class="flex items-center justify-between">
              <div>
                <p class="text-xs font-black text-violet-700 uppercase tracking-widest">Nota {{ originalIndex + 1 }}</p>
                <p class="text-base font-black text-slate-900 mt-0.5">{{ form.so_number }}</p>
                <p class="text-xs font-bold text-slate-500 mt-0.5">🏪 {{ form.store_name }}</p>
              </div>
              <div class="text-right">
                <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Total SO Awal</p>
                <p class="text-sm font-black text-slate-700">Rp {{ form.default_amount.toLocaleString('id-ID') }}</p>
              </div>
            </div>

            <!-- Input fields -->
            <div class="grid grid-cols-2 gap-3">
              <!-- Payment Method -->
              <div>
                <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block mb-1.5">
                  Metode Pembayaran
                </label>
                <select
                  v-model="form.payment_method"
                  class="w-full bg-white border-2 border-slate-200 focus:border-violet-400 rounded-2xl px-4 py-2.5 text-sm font-bold text-slate-900 outline-none transition-colors"
                >
                  <option value="CASH">💵 CASH</option>
                  <option value="TRANSFER">🏦 TRANSFER</option>
                  <option value="TEMPO">📅 TEMPO (Kredit)</option>
                </select>
              </div>

              <!-- Total Tagihan Aktual -->
              <div>
                <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block mb-1.5">
                  Total Tagihan Aktual (Rp)
                </label>
                <input
                  v-model="form.total_amount"
                  @input="form.payment_method !== 'TEMPO' && (form.paid_amount = Number(form.total_amount))"
                  type="number"
                  min="0"
                  :placeholder="form.default_amount.toString()"
                  class="w-full bg-white border-2 border-slate-200 focus:border-violet-400 rounded-2xl px-4 py-2.5 text-sm font-bold text-slate-900 outline-none transition-colors"
                />
              </div>
            </div>

            <!-- Bayar Sekarang: hanya tampil untuk TEMPO -->
            <div v-if="form.payment_method === 'TEMPO'" class="bg-amber-50 border-2 border-amber-200 rounded-2xl p-4 space-y-2">
              <div class="flex items-center gap-2">
                <span class="text-lg">📅</span>
                <span class="text-xs font-black text-amber-700 uppercase tracking-widest">TEMPO — Pembayaran Pertama (Uang Muka)</span>
              </div>
              <p class="text-[11px] text-amber-600 font-medium">
                Masukkan jumlah yang sudah diterima driver dari toko saat pengiriman. Sisanya akan ditagih kemudian.
              </p>
              <input
                v-model="form.paid_amount"
                type="number"
                min="0"
                :max="Number(form.total_amount) || form.default_amount"
                :placeholder="`Maks. Rp ${(Number(form.total_amount) || form.default_amount).toLocaleString('id-ID')}`"
                class="w-full bg-white border-2 border-amber-300 focus:border-amber-500 rounded-xl px-4 py-2.5 text-sm font-bold text-slate-900 outline-none transition-colors"
              />
              <div class="flex justify-between text-xs font-bold text-amber-700 pt-1">
                <span>Sisa tagihan:</span>
                <span>Rp {{ Math.max(0, (Number(form.total_amount) || form.default_amount) - Number(form.paid_amount)).toLocaleString('id-ID') }}</span>
              </div>
            </div>

            <!-- Catatan (opsional) -->
            <div>
              <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block mb-1.5">
                Catatan (Opsional)
              </label>
              <input
                v-model="form.notes"
                type="text"
                placeholder="Contoh: ada retur 2 pcs, dibayar sebagian..."
                class="w-full bg-white border-2 border-slate-200 focus:border-violet-400 rounded-2xl px-4 py-2.5 text-xs font-bold text-slate-900 outline-none transition-colors"
              />
            </div>

            <!-- Estimasi total -->
            <div
              :class="Number(form.total_amount) !== form.default_amount && Number(form.total_amount) > 0
                ? 'bg-amber-50 border-amber-200 text-amber-800'
                : 'bg-violet-50 border-violet-100 text-violet-800'"
              class="border rounded-2xl p-3 space-y-1"
            >
              <div class="flex justify-between items-center">
                <span class="text-[10px] font-black uppercase tracking-widest">Total Faktur</span>
                <span class="text-sm font-black">
                  Rp {{ (Number(form.total_amount) || form.default_amount).toLocaleString('id-ID') }}
                  <span
                    v-if="Number(form.total_amount) !== form.default_amount && Number(form.total_amount) > 0"
                    class="text-[9px] ml-1 font-bold"
                  >
                    (berbeda dari SO)
                  </span>
                </span>
              </div>
              <div v-if="form.payment_method === 'TEMPO'" class="flex justify-between items-center border-t border-current/20 pt-1">
                <span class="text-[10px] font-bold uppercase tracking-widest opacity-75">Masuk Kasir Sekarang</span>
                <span class="text-sm font-black">Rp {{ Number(form.paid_amount).toLocaleString('id-ID') }}</span>
              </div>
            </div>

          </div>

          <!-- Empty search state -->
          <div v-if="forceInvoiceForm.length > 0 && filteredForceInvoiceForm.length === 0" class="py-8 text-center flex flex-col items-center justify-center">
            <div class="w-10 h-10 bg-slate-100 rounded-xl flex items-center justify-center text-slate-400 mb-2">
              <Search class="w-5 h-5" />
            </div>
            <p class="text-xs font-bold text-slate-600">Toko tidak ditemukan</p>
            <p class="text-[11px] text-slate-400 mt-0.5">Tidak ada nota yang cocok dengan "{{ forceInvoiceSearchQuery }}"</p>
            <button
              type="button"
              @click="forceInvoiceSearchQuery = ''"
              class="mt-3 text-xs font-bold text-violet-600 hover:text-violet-700 underline"
            >
              Hapus filter
            </button>
          </div>

          <div v-if="forceInvoiceForm.length === 0" class="py-8 text-center text-slate-400 font-bold text-sm">
            Tidak ada nota dengan Sales Order dalam SJ ini.
          </div>
        </div>

        <!-- Footer actions -->
        <div class="p-6 bg-slate-50/60 border-t border-slate-100 flex gap-3 justify-end">
          <Button
            variant="outline"
            class="rounded-2xl px-6 font-black border-2"
            @click="showForceInvoiceModal = false"
            :disabled="isCreatingInvoice"
          >
            Batal
          </Button>
          <Button
            class="bg-violet-600 hover:bg-violet-700 text-white rounded-2xl px-8 font-black shadow-lg shadow-violet-200 disabled:opacity-50"
            @click="submitForceInvoice"
            :disabled="isCreatingInvoice || forceInvoiceForm.length === 0"
          >
            <span v-if="isCreatingInvoice" class="flex items-center gap-2">
              <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Membuat Faktur...
            </span>
            <span v-else>✓ Buat Faktur Sekarang</span>
          </Button>
        </div>
      </div>
    </div>

    <!-- Modal: Revise Qty (Admin Nota) -->
    <div
      v-if="showReviseQtyModal"
      class="fixed inset-0 z-[90] flex items-center justify-center p-4 overflow-y-auto"
    >
      <div class="absolute inset-0 bg-slate-900/70 backdrop-blur-md" @click="showReviseQtyModal = false" />
      <div
        class="relative z-10 w-full max-w-2xl bg-white rounded-[40px] shadow-2xl overflow-hidden animate-in zoom-in-95 duration-300"
      >
        <!-- Header -->
        <div class="bg-gradient-to-r from-blue-600 to-indigo-700 p-7 text-white">
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 bg-white/20 rounded-2xl flex items-center justify-center shrink-0">
              <Pencil class="w-6 h-6" />
            </div>
            <div>
              <h3 class="text-xl font-black tracking-tight">Sesuaikan Qty Pesanan</h3>
              <p class="text-blue-200 text-xs font-bold mt-0.5 uppercase tracking-widest">
                Nota: {{ reviseForm.so_number }}
              </p>
            </div>
          </div>
          <p class="mt-4 text-sm text-blue-100 font-medium leading-relaxed">
            Ubah kuantitas barang sebelum Surat Jalan mulai dikirim. Perubahan akan otomatis memperbarui Sales Order dan Faktur Penjualan.
          </p>
        </div>

        <!-- Body -->
        <div class="p-6 max-h-[55vh] overflow-y-auto space-y-4">
          <div v-for="(item, idx) in reviseForm.items" :key="item.sales_order_item_id" class="flex flex-col sm:flex-row justify-between items-start sm:items-center bg-slate-50 border border-slate-200 rounded-2xl p-4 gap-4">
            <div>
              <p class="text-sm font-bold text-slate-900">{{ item.product_name }}</p>
              <p class="text-[10px] font-black text-slate-500 uppercase tracking-widest mt-1">Harga: Rp {{ item.price.toLocaleString('id-ID') }}</p>
            </div>
            <div class="flex items-center gap-3">
              <div class="text-right">
                <p class="text-[10px] font-black text-slate-400 uppercase">Qty Awal</p>
                <p class="text-sm font-bold line-through text-slate-400">{{ item.ordered_quantity }}</p>
              </div>
              <ArrowRight class="w-4 h-4 text-slate-300" />
              <div>
                <p class="text-[10px] font-black text-indigo-500 uppercase">Qty Baru</p>
                <input v-model.number="item.new_quantity" type="number" min="0" class="w-20 bg-white border-2 border-indigo-200 focus:border-indigo-500 rounded-xl px-2 py-1.5 text-sm font-bold text-center text-slate-900 outline-none" />
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-6 bg-slate-50/60 border-t border-slate-100 flex gap-3 justify-end">
          <Button
            variant="outline"
            class="rounded-2xl px-6 font-black border-2"
            @click="showReviseQtyModal = false"
            :disabled="isRevisingQty"
          >
            Batal
          </Button>
          <Button
            class="bg-blue-600 hover:bg-blue-700 text-white rounded-2xl px-8 font-black shadow-lg shadow-blue-200 disabled:opacity-50"
            @click="submitReviseQty"
            :disabled="isRevisingQty"
          >
            {{ isRevisingQty ? 'Menyimpan...' : 'Simpan Perubahan' }}
          </Button>
        </div>
      </div>
    </div>

    <!-- Zoomed QR Overlay -->
    <div v-if="zoomedQR" class="fixed inset-0 z-[100] flex items-center justify-center p-8">
      <div class="absolute inset-0 bg-slate-900/80 backdrop-blur-md" @click="zoomedQR = null" />
      <div
        class="bg-white p-12 rounded-[48px] relative z-10 shadow-2xl flex flex-col items-center gap-8 animate-in zoom-in-50 duration-300"
      >
        <div class="text-center">
          <h4 class="text-xl font-black text-slate-900 uppercase tracking-widest">SCAN VALIDASI</h4>
          <p class="text-sm font-bold text-slate-500 mt-2">{{ zoomedQR }}</p>
        </div>
        <div class="bg-white p-6 rounded-[32px] border-4 border-slate-100">
          <img
            :src="`https://api.qrserver.com/v1/create-qr-code/?size=400x400&margin=0&data=${zoomedQR}`"
            class="w-64 h-64 sm:w-80 sm:h-80"
          />
        </div>
        <Button
          @click="zoomedQR = null"
          variant="outline"
          class="rounded-2xl px-8 font-black border-2"
          >TUTUP</Button
        >
      </div>
    </div>

    <!-- Zoomed POD Overlay -->
    <div v-if="zoomedPOD" class="fixed inset-0 z-[100] flex items-center justify-center p-8">
      <div class="absolute inset-0 bg-slate-900/80 backdrop-blur-md" @click="zoomedPOD = null" />
      <div
        class="bg-white p-6 rounded-[32px] relative z-10 shadow-2xl flex flex-col items-center gap-6 animate-in zoom-in-50 duration-300 max-w-lg w-full"
      >
        <div class="text-center">
          <h4 class="text-lg font-black text-slate-900 uppercase tracking-widest">
            FOTO BUKTI PENGIRIMAN (POD)
          </h4>
        </div>
        <div
          class="bg-slate-100 rounded-2xl overflow-hidden border border-slate-200 w-full flex items-center justify-center p-1"
        >
          <img :src="zoomedPOD" class="max-h-[60vh] max-w-full object-contain rounded-xl" />
        </div>
        <Button
          @click="zoomedPOD = null"
          variant="outline"
          class="rounded-xl px-8 font-black border-2"
          >TUTUP</Button
        >
      </div>
    </div>

    <!-- DO Printing Modal (SJ Preview - Landscape Format) -->
    <div
      v-if="selectedDO"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[60] flex items-center justify-center p-0 sm:p-4 overflow-y-auto"
    >
      <Card
        id="surat-jalan-card"
        class="w-full max-w-[95%] sm:max-w-6xl bg-white shadow-2xl border-0 overflow-hidden my-auto print:shadow-none print:max-w-full"
      >
        <!-- Header -->
        <div class="bg-white p-4 print:p-2 border-b border-slate-100">
          <div class="grid grid-cols-3 gap-8 items-center">
            <div class="space-y-1">
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
                SOPIR / ARMADA
              </p>
              <p class="text-[11px] font-black text-slate-900">
                {{
                  selectedDO.driver
                    ? `${selectedDO.driver.first_name} ${selectedDO.driver.last_name}`
                    : (selectedDO.daily_driver?.name || '........................')
                }}
                / {{ selectedDO.vehicle?.license_plate || '..........' }}
              </p>
              <p v-if="selectedDO.helper || selectedDO.daily_helper" class="text-[9px] font-bold text-slate-600 mt-0.5">
                Helper: {{
                  selectedDO.helper
                    ? `${selectedDO.helper.first_name} ${selectedDO.helper.last_name}`
                    : (selectedDO.daily_helper?.name || '')
                }}
                <span v-if="selectedDO.daily_helper" class="text-amber-500 ml-1">(Harian)</span>
              </p>
            </div>
            <div class="space-y-1 text-center">
              <div
                class="inline-block bg-primary/10 px-4 py-1 rounded-full border border-primary/20 print:border-slate-900"
              >
                <p
                  class="text-[10px] font-black text-primary uppercase tracking-tighter print:text-slate-900"
                >
                  SURAT JALAN DIGITAL
                </p>
              </div>
              <p class="text-[12px] font-black text-slate-900 mt-1 uppercase tracking-tight">
                {{ selectedDO.delivery_order_no }}
              </p>
            </div>
            <div class="text-right">
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
                TANGGAL TERBIT
              </p>
              <p class="text-[10px] font-black text-slate-900 uppercase leading-tight mt-1">
                {{ formatDate(selectedDO.created_at) }}
              </p>
            </div>
          </div>
        </div>

        <div class="p-4 sm:p-6 space-y-4 print:p-2">
          <!-- Main Table -->
          <div class="border-2 border-slate-900 rounded-lg overflow-x-auto bg-white">
            <table class="w-full text-left text-[10px] border-collapse whitespace-nowrap">
              <thead>
                <tr class="bg-slate-50 border-b-2 border-slate-900">
                  <th class="p-2 border-r border-slate-300 font-black text-center w-8">NO</th>
                  <th class="p-1 border-r border-slate-300 font-black text-center w-12 text-[8px]">
                    QR SCAN
                  </th>
                  <th class="p-2 border-r border-slate-300 font-black">NAMA TOKO / ALAMAT</th>
                  <th class="p-2 border-r border-slate-300 font-black text-right w-20 bg-amber-50">
                    TAGIHAN
                  </th>
                  <th
                    class="p-2 border-r border-slate-300 font-black text-center w-24 bg-emerald-50"
                  >
                    BAYAR
                  </th>
                  <th class="p-2 border-r border-slate-300 font-black text-center w-16">KET</th>

                  <th
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 font-black text-center text-[7px] bg-slate-100/30 w-10 uppercase"
                  >
                    {{ prod.name }}
                  </th>

                  <th
                    class="p-2 border-r border-slate-300 font-black text-center w-10 bg-pink-50 text-[8px]"
                  >
                    HADIAH
                  </th>
                  <th class="p-2 font-black text-center w-10 text-[8px]">RETUR</th>
                </tr>
              </thead>
              <tbody class="divide-y border-slate-300">
                <tr v-for="(item, idx) in selectedDO.items" :key="item.id" class="h-8">
                  <td class="p-2 border-r border-slate-300 text-center font-bold">
                    {{ Number(idx) + 1 }}
                  </td>
                  <td class="p-1 border-r border-slate-300 text-center align-middle">
                    <img
                      :src="`https://api.qrserver.com/v1/create-qr-code/?size=60x60&margin=0&data=${(item.sales_order || item.sales_transaction)?.so_number || (item.sales_order || item.sales_transaction)?.invoice_no || item.id}`"
                      class="w-8 h-8 mx-auto"
                    />
                  </td>
                  <td class="p-2 border-r border-slate-300 whitespace-normal">
                    <p class="font-black text-slate-900 leading-none">
                      {{ (item.sales_order || item.sales_transaction)?.store?.name }}
                    </p>
                    <p class="text-[7px] text-slate-500 font-bold mt-0.5">
                      {{ (item.sales_order || item.sales_transaction)?.store?.address }}
                    </p>
                  </td>
                  <td class="p-2 border-r border-slate-300 text-right font-black bg-amber-50/20">
                    {{
                      formatCurrency(
                        (item.sales_order || item.sales_transaction)?.total_amount || 0,
                      )
                        .replace('Rp', '')
                        .trim()
                    }}
                  </td>
                  <td class="p-2 border-r border-slate-300 bg-emerald-50/10"></td>
                  <td class="p-2 border-r border-slate-300 text-center font-bold text-[8px]">
                    {{
                      (item.sales_order || item.sales_transaction)?.payment_method === 'VA' &&
                      (item.sales_order || item.sales_transaction)?.midtrans_bank
                        ? `VA ${(item.sales_order || item.sales_transaction).midtrans_bank.toUpperCase()}`
                        : (item.sales_order || item.sales_transaction)?.payment_method || 'COD'
                    }}
                  </td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 text-center font-black"
                  >
                    {{ getProductQty(item, prod.id) || '' }}
                  </td>
                  <td class="p-1 border-r border-slate-300 text-[6px] font-bold text-pink-700 whitespace-pre-wrap leading-tight text-center align-middle">
                    {{ (item.sales_order || item.sales_transaction)?.bonus_items?.map((b: any) => `${b.quantity} ${b.bonus_inventory?.name}`).join('\n') || '' }}
                  </td>
                  <td class="p-2"></td>
                </tr>

                <!-- ROW UNTUK SPARE / DO -->
                <tr v-if="selectedDO.spares && selectedDO.spares.length > 0" class="h-8 bg-blue-50/10">
                  <td class="p-2 border-r border-slate-300 text-center font-bold">
                    {{ (selectedDO.items?.length || 0) + 1 }}
                  </td>
                  <td class="p-1 border-r border-slate-300 text-center align-middle">
                    -
                  </td>
                  <td class="p-2 border-r border-slate-300 whitespace-normal">
                    <p class="font-black text-slate-900 leading-none">
                      [SPARE / DO]
                    </p>
                    <p class="text-[7px] text-slate-500 font-bold mt-0.5">
                      Barang Tambahan Gudang
                    </p>
                  </td>
                  <td class="p-2 border-r border-slate-300 text-right font-black bg-amber-50/20">-</td>
                  <td class="p-2 border-r border-slate-300 bg-emerald-50/10">-</td>
                  <td class="p-2 border-r border-slate-300 text-center font-bold text-[8px]">-</td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="'spare-'+prod.id"
                    class="p-1 border-r border-slate-300 text-center font-black text-blue-700"
                  >
                    {{ getSpareQty(prod.id) || '' }}
                  </td>
                  <td class="p-2 border-r border-slate-300">-</td>
                  <td class="p-2">-</td>
                </tr>

                <!-- Placeholder Rows for Manual Entry (Fills up to 15 rows) -->
                <tr
                  v-for="n in Math.max(0, 15 - (selectedDO.items?.length || 0))"
                  :key="'blank-' + n"
                  class="h-6 bg-slate-50/10"
                >
                  <td class="p-2 border-r border-slate-300 text-center text-slate-300 font-bold">
                    {{ (selectedDO.items?.length || 0) + n }}
                  </td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300 bg-amber-50/10"></td>
                  <td class="p-2 border-r border-slate-300 bg-emerald-50/5"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300"
                  ></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2"></td>
                </tr>

                <!-- Total Row -->
                <tr class="bg-slate-50 font-black border-t-2 border-slate-900 h-8">
                  <td
                    colspan="3"
                    class="p-2 text-right border-r border-slate-300 uppercase tracking-widest text-[8px]"
                  >
                    TOTAL BATCH
                  </td>
                  <td class="p-2 text-right border-r border-slate-300 bg-amber-50">
                    {{ formatCurrency(calculateBatchTotal).replace('Rp', '').trim() }}
                  </td>
                  <td colspan="2" class="border-r border-slate-300"></td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 text-center text-primary bg-slate-100/30"
                  >
                    {{ calculateProductTotal(prod.id) }}
                  </td>
                  <td colspan="2"></td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Compact Signature Section -->
          <div class="grid grid-cols-3 gap-8 mt-6">
            <div class="text-center space-y-8">
              <p class="text-[9px] font-black uppercase tracking-widest">SOPIR / HELPER</p>
              <p class="text-[8px] font-bold text-slate-900">
                ( .................................. )
              </p>
            </div>
            <div class="text-center space-y-8">
              <p class="text-[9px] font-black uppercase tracking-widest">ADMIN GUDANG</p>
              <p class="text-[8px] font-bold text-slate-900">
                ( .................................. )
              </p>
            </div>
            <div class="flex justify-end items-center pr-10">
              <div class="text-right mr-4">
                <p class="text-[8px] font-black text-slate-400 uppercase leading-none">
                  VALIDASI SCAN
                </p>
                <p class="text-[7px] text-slate-300 italic mt-1 uppercase">Sistem Digital</p>
              </div>
              <img
                :src="`https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${selectedDO.delivery_order_no}`"
                class="w-20 h-20 border-2 border-slate-900 p-1 rounded bg-white shadow-sm"
              />
            </div>
          </div>
        </div>

        <CardFooter class="p-6 bg-slate-50 border-t flex justify-end gap-3 print:hidden">
          <Button
            variant="ghost"
            class="font-bold text-slate-500 rounded-xl"
            @click="selectedDO = null"
            >TUTUP PREVIEW</Button
          >
          <Button
            class="bg-slate-900 hover:bg-slate-800 text-white font-black px-12 py-6 rounded-2xl shadow-xl shadow-slate-200"
            @click="handlePrint"
          >
            <Printer class="w-5 h-5 mr-3" />
            CETAK SURAT JALAN (LANDSCAPE)
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showEditModal = false"
      />
      <div
        class="bg-white rounded-[40px] w-full max-w-2xl relative z-10 shadow-2xl overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-300"
      >
        <div class="p-8 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
          <div>
            <h3 class="text-2xl font-black text-slate-900">Edit Surat Jalan</h3>
            <p class="text-xs text-slate-500 font-bold uppercase tracking-wider">
              {{ editingBatch?.delivery_order_no }}
            </p>
          </div>
          <Button
            variant="ghost"
            size="icon"
            @click="showEditModal = false"
            class="rounded-full hover:bg-slate-100"
          >
            <X class="w-6 h-6" />
          </Button>
        </div>

        <div class="p-8 overflow-y-auto space-y-6 flex-1">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Supir / Driver</label
              >
              <select
                v-model="batchForm.driver_id"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              >
                <option value="">— Pilih Supir —</option>
                <option v-for="d in drivers" :key="d.id" :value="d.id">
                  {{ d.first_name }} {{ d.last_name }}
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Armada / Kendaraan</label
              >
              <select
                v-model="batchForm.vehicle_id"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              >
                <option value="">— Pilih Armada —</option>
                <option v-for="v in vehicles" :key="v.id" :value="v.id">
                  {{ v.license_plate }} - {{ v.model }}
                </option>
              </select>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Status Pengiriman</label
            >
            <select
              v-model="batchForm.status"
              class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
            >
              <option value="WAITING_APPROVAL">MENUNGGU PERSETUJUAN</option>
              <option value="SUPERVISOR_APPROVED">DISETUJUI SUPERVISOR</option>
              <option value="PENDING">DIPROSES ADMIN</option>
              <option value="PICKING">PENGAMBILAN BARANG</option>
              <option value="ON_DELIVERY">DALAM PENGIRIMAN</option>
              <option value="COMPLETED">SELESAI</option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Tanggal Pengiriman</label
            >
            <input
              type="date"
              v-model="batchForm.delivery_date"
              class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
            />
          </div>

          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-black text-slate-900 uppercase tracking-wider">
                Daftar Nota dalam Batch
              </h4>
              <span class="text-[10px] font-black bg-primary/10 text-primary px-3 py-1 rounded-full"
                >{{ editingBatch?.items?.length || 0 }} NOTA</span
              >
            </div>

            <div class="space-y-2 max-h-48 overflow-y-auto pr-1">
              <div v-if="!editingBatch?.items || editingBatch.items.length === 0" class="text-xs text-slate-400 font-semibold italic text-center py-4 bg-slate-50 rounded-2xl border border-slate-100">
                Belum ada nota di dalam Surat Jalan ini.
              </div>
              <div
                v-else
                v-for="(item, idx) in editingBatch?.items"
                :key="item.id || idx"
                class="flex items-center justify-between p-4 bg-slate-50 rounded-2xl border border-slate-100"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="w-8 h-8 rounded-xl bg-white border border-slate-200 flex items-center justify-center font-black text-[10px] text-primary"
                  >
                    SO
                  </div>
                  <div>
                    <p class="text-[10px] font-black text-slate-900">
                      {{ (item.sales_order || item.sales_transaction)?.so_number || 'NO INVOICE' }}
                    </p>
                    <p class="text-[9px] text-slate-500 font-bold">
                      {{
                        (item.sales_order || item.sales_transaction)?.store?.name || 'Toko Manual'
                      }}
                    </p>
                  </div>
                </div>
                <div class="flex items-center gap-3">
                  <div class="text-right">
                    <p class="text-xs font-black text-slate-900">
                      {{
                        formatCurrency(
                          (item.sales_order || item.sales_transaction)?.total_amount || 0,
                        )
                      }}
                    </p>
                    <p class="text-[9px] font-bold text-emerald-600">{{ item.status }}</p>
                  </div>
                  <button
                    v-if="canEditItems"
                    @click="removeItemFromEditingBatch(idx)"
                    class="p-2 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-xl transition-colors"
                    title="Hapus Nota"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Add Invoice to Batch section -->
          <div v-if="canEditItems" class="space-y-4 border-t border-slate-100 pt-4">
            <h4 class="text-sm font-black text-slate-900 uppercase tracking-wider">
              Tambah Nota ke Surat Jalan
            </h4>
            <div v-if="pendingOrders.length === 0" class="text-xs text-slate-400 font-semibold italic">
              Tidak ada nota pending (CONFIRMED) yang tersedia untuk ditambahkan.
            </div>
            <div v-else class="space-y-2 max-h-40 overflow-y-auto pr-1">
              <div
                v-for="order in pendingOrders"
                :key="order.id"
                class="flex items-center justify-between p-3 bg-slate-50 hover:bg-slate-100/70 rounded-2xl border border-slate-100 transition-colors"
              >
                <div class="flex items-center gap-3">
                  <input
                    type="checkbox"
                    :id="'add-so-' + order.id"
                    :checked="isOrderSelectedForAdding(order.id)"
                    @change="toggleOrderForAdding(order)"
                    class="w-4 h-4 rounded text-primary border-slate-300 focus:ring-primary/20"
                  />
                  <label :for="'add-so-' + order.id" class="cursor-pointer">
                    <p class="text-[10px] font-black text-slate-900">{{ order.so_number }}</p>
                    <p class="text-[9px] text-slate-500 font-bold">{{ order.store?.name }}</p>
                  </label>
                </div>
                <p class="text-xs font-black text-slate-900">
                  {{ formatCurrency(order.total_amount) }}
                </p>
              </div>
            </div>
          </div>

          <!-- Add Tempo Invoice to Batch section -->
          <div v-if="canEditItems" class="space-y-4 border-t border-amber-200 pt-4 mt-4 bg-amber-50/50 rounded-xl p-3">
            <div class="flex items-center gap-2">
              <div class="w-2 h-2 rounded-full bg-amber-500 animate-pulse"></div>
              <h4 class="text-sm font-black text-amber-800 uppercase tracking-wider">
                Tagihan Tempo – Perlu Ditagih ke Toko
              </h4>
            </div>
            <p class="text-xs text-amber-600 font-medium -mt-2">Pilih faktur tempo yang belum lunas untuk ditagih driver saat pengiriman.</p>
            <div v-if="deliveryStore.outstandingInvoices.length === 0" class="text-xs text-slate-400 font-semibold italic">
              Tidak ada faktur tempo yang belum lunas.
            </div>
            <div v-else class="space-y-2 max-h-48 overflow-y-auto pr-1">
              <div
                v-for="trx in deliveryStore.outstandingInvoices"
                :key="trx.id"
                class="flex items-center justify-between p-3 rounded-2xl border transition-colors"
                :class="isTempoInvoiceSelected(trx.id) ? 'bg-amber-100 border-amber-300' : 'bg-white border-slate-100 hover:bg-amber-50'"
              >
                <div class="flex items-center gap-3">
                  <input
                    type="checkbox"
                    :id="'add-trx-' + trx.id"
                    :checked="isTempoInvoiceSelected(trx.id)"
                    @change="toggleTempoInvoiceForAdding(trx)"
                    class="w-4 h-4 rounded text-amber-500 border-slate-300 focus:ring-amber-500/20"
                  />
                  <label :for="'add-trx-' + trx.id" class="cursor-pointer">
                    <p class="text-[10px] font-black text-slate-900">{{ trx.receipt_no || trx.id.slice(0, 8).toUpperCase() }}</p>
                    <p class="text-[9px] text-slate-500 font-bold">{{ trx.store?.name }}</p>
                    <p class="text-[9px] text-amber-600 font-bold" v-if="trx.payment_due_date">
                      Jatuh Tempo: {{ formatDate(trx.payment_due_date) }}
                    </p>
                  </label>
                </div>
                <div class="text-right">
                  <p class="text-xs font-black text-slate-900">{{ formatCurrency(trx.total_amount) }}</p>
                  <p class="text-[9px] text-amber-600 font-bold">Sisa: {{ formatCurrency(trx.total_amount - (trx.paid_amount || 0)) }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Add Subsequent Items to Batch section -->
          <div v-if="canEditItems" class="space-y-4 border-t border-slate-100 pt-4 mt-4">
            <div class="flex items-center justify-between mb-2">
              <h4 class="text-sm font-black text-slate-900 uppercase tracking-wider">
                Tambah Barang Susulan
              </h4>
              <button
                @click="showManualSubsequentModal = !showManualSubsequentModal"
                class="text-xs font-bold text-primary hover:text-primary/80 transition-colors flex items-center gap-1"
                title="Buat Barang Susulan (Manual)"
              >
                <PlusCircle class="w-4 h-4" /> Custom
              </button>
            </div>

            <!-- Form Buat Susulan Manual -->
            <div v-if="showManualSubsequentModal" class="bg-indigo-50/50 rounded-xl p-4 border border-indigo-100 mt-2 space-y-3">
              <div class="flex justify-between items-center mb-1">
                <h5 class="text-xs font-bold text-indigo-800">Form Buat Susulan Manual</h5>
                <button @click="showManualSubsequentModal = false" class="text-indigo-400 hover:text-indigo-600">
                  <X class="w-4 h-4" />
                </button>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Salesman</label>
                  <select v-model="manualSubsequentForm.salesman_id" class="w-full text-xs p-2 rounded-lg border border-slate-200">
                    <option value="" disabled>Pilih Salesman</option>
                    <option v-for="emp in salesEmployees" :key="emp.id" :value="emp.id">
                      {{ emp.first_name }} {{ emp.last_name }}
                    </option>
                  </select>
                </div>
                <div>
                  <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Toko</label>
                  <select v-model="manualSubsequentForm.store_id" class="w-full text-xs p-2 rounded-lg border border-slate-200">
                    <option value="" disabled>Pilih Toko</option>
                    <option v-for="store in masterStore.stores" :key="store.id" :value="store.id">
                      {{ store.name }}
                    </option>
                  </select>
                </div>
                <div class="col-span-2">
                  <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Produk</label>
                  <select v-model="manualSubsequentForm.product_id" class="w-full text-xs p-2 rounded-lg border border-slate-200">
                    <option value="" disabled>Pilih Produk</option>
                    <option v-for="prod in products" :key="prod.id" :value="prod.id">
                      {{ prod.name }} ({{ prod.sku }})
                    </option>
                  </select>
                </div>
                <div>
                  <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Qty</label>
                  <input type="number" min="1" v-model.number="manualSubsequentForm.quantity" class="w-full text-xs p-2 rounded-lg border border-slate-200" />
                </div>
                <div class="col-span-2">
                  <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1">Catatan</label>
                  <textarea v-model="manualSubsequentForm.notes" rows="2" class="w-full text-xs p-2 rounded-lg border border-slate-200" placeholder="Cth: Retur rusak diganti baru..."></textarea>
                </div>
              </div>
              <Button @click="createManualSubsequent" :disabled="isSubmittingSubsequent" class="w-full text-xs h-8 mt-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg flex items-center justify-center">
                {{ isSubmittingSubsequent ? 'Menyimpan...' : 'Simpan & Tambah ke Antrean' }}
              </Button>
            </div>

            <div v-if="pendingSubsequents.length === 0" class="text-xs text-slate-400 font-semibold italic">
              Tidak ada barang susulan pending yang tersedia.
            </div>
            <div v-else class="space-y-2 max-h-40 overflow-y-auto pr-1">
              <div
                v-for="sub in pendingSubsequents"
                :key="sub.id"
                class="flex items-center justify-between p-3 bg-slate-50 hover:bg-slate-100/70 rounded-2xl border border-slate-100 transition-colors"
              >
                <div class="flex items-center gap-3">
                  <input
                    type="checkbox"
                    :id="'add-sub-' + sub.id"
                    :checked="isSubsequentSelectedForAdding(sub.id)"
                    @change="toggleSubsequentForAdding(sub)"
                    class="w-4 h-4 rounded text-primary border-slate-300 focus:ring-primary/20"
                  />
                  <label :for="'add-sub-' + sub.id" class="cursor-pointer">
                    <p class="text-[10px] font-black text-slate-900">{{ sub.product?.name }}</p>
                    <p class="text-[9px] text-slate-500 font-bold">Qty: {{ sub.quantity }} {{ sub.unit }}</p>
                  </label>
                </div>
                <p class="text-xs font-black text-slate-900">
                  {{ sub.store?.name || 'Unknown Store' }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div class="p-8 bg-slate-50/50 border-t border-slate-100 flex gap-4">
          <Button
            variant="outline"
            @click="showEditModal = false"
            class="flex-1 rounded-2xl py-6 font-black border-slate-200"
          >
            BATAL
          </Button>
          <Button
            @click="handleUpdateBatch"
            :disabled="isSubmitting"
            class="flex-[2] bg-primary text-white rounded-2xl py-6 font-black shadow-xl shadow-primary/20"
          >
            {{ isSubmitting ? 'MENYIMPAN...' : 'SIMPAN PERUBAHAN' }}
          </Button>
        </div>
      </div>
    </div>
  </div>

  <!-- Modal Susulkan Pesanan SO ke Surat Jalan -->
  <AppendSubsequentOrderModal
    :is-open="showAppendModal"
    :batch="appendTargetBatch"
    @close="showAppendModal = false"
    @success="onAppendSuccess"
  />

  <!-- Print-only Component (Must be outside the main print:hidden wrapper) -->
  <DigitalSuratJalanPrint v-if="selectedDO" :selectedDO="selectedDO" />
</template>


