<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMasterDataStore } from '@/stores/masterData'
import { useWarehouseStore } from '@/stores/warehouse'
import { useVehicleStore } from '@/stores/vehicle'
import apiClient from '@/api/axios'
import { warehouseApi } from '@/api/warehouse'
import { filterSellableProducts } from '@/utils/productUnits'
import { toast } from 'vue-sonner'
import { useAuthStore } from '@/stores/auth'
import DataTable from '@/components/DataTable.vue'
import {
  Package,
  Truck,
  QrCode,
  Search,
  RefreshCw,
  CheckCircle2,
  ScanLine,
  X,
  AlertCircle,
  Clock,
  BoxSelect,
  ChevronRight,
  User,
  FileText,
  ShieldCheck,
  Receipt,
  RotateCcw,
  Info,
  Edit,
  PackageCheck,
  Eye,
  XCircle,
  Boxes,
  FileCheck2,
  Calendar,
} from 'lucide-vue-next'

const masterStore = useMasterDataStore()
const warehouseStore = useWarehouseStore()
const vehicleStore = useVehicleStore()
const authStore = useAuthStore()

const batchColumns = [
  { key: 'delivery_order_no', label: 'Nomor SJ' },
  { key: 'delivery_date', label: 'Tanggal Pengiriman' },
  { key: 'status', label: 'Status' },
  { key: 'driver', label: 'Pengemudi' },
  { key: 'vehicle', label: 'Kendaraan' },
  { key: 'stats', label: 'Item & Qty' },
  { key: 'actions', label: 'Aksi' },
]

// ─── State ───────────────────────────────────────────────────────────────────
const batches = ref<any[]>([])
const loading = ref(false)
const scanCode = ref('')
const scanLoading = ref(false)
const scanResult = ref<any | null>(null)
const scanError = ref('')
const activeTab = ref<'SUPERVISOR_APPROVED' | 'PICKING' | 'ON_DELIVERY' | 'COMPLETED'>('SUPERVISOR_APPROVED')

const showDetailModal = ref(false)
const selectedBatch = ref<any | null>(null)
const confirmingReady = ref(false)
const tableSearch = ref('')

const checkedItems = ref<Record<string, boolean>>({})

// DO states
const activeModalTab = ref<'nota' | 'do' | 'spares' | 'leftovers'>('nota')
const warehouseDO = ref<any | null>(null)
const loadingDO = ref(false)
const savingDO = ref(false)
const doItems = ref<any[]>([])
const doNotes = ref('')

// Spares (Barang Cadangan) states
const products = ref<any[]>([])
const localSpares = ref<Array<{ product_id: string; quantity: number }>>([])
const savingSpares = ref(false)
const newSpareProductId = ref('')
const newSpareQty = ref(1)

// Leftovers (Barang Susulan) states
const pendingLeftovers = ref<any[]>([])
const slippingLeftoverId = ref<string | null>(null)

// MT DO States
const showMtDetailModal = ref(false)
const mtWorkingOrder = ref<any>(null)
const mtAction = ref<'APPROVE'|'REJECT'|'LOADING'|'SHIP'|'DETAIL'|null>(null)
const mtIsSubmitting = ref(false)
const mtApproveBy = ref('')
const mtRejectReason = ref('')
const mtLoadingForm = ref<any>({ prepared_by: '', items: [] })
const mtShipForm = ref({ vehicle_id: '', driver_id: '', driver_name: '', notes: '' })
const mtVehicles = ref<any[]>([])
const mtDrivers = ref<any[]>([])

async function fetchVehicles() {
  try {
    const res = await apiClient.get('/vehicles')
    mtVehicles.value = res.data?.data || res.data || []
  } catch (e) {
    console.error('Failed to fetch vehicles', e)
  }
}

async function fetchDrivers() {
  try {
    const res = await apiClient.get('/employees?job_title=DRIVER&limit=1000')
    mtDrivers.value = res.data?.data || res.data || []
  } catch (e) {
    console.error('Failed to fetch drivers', e)
  }
}

const allItemsChecked = computed(() => {
  if (!selectedBatch.value || !selectedBatch.value.items) return false
  let totalItems = 0
  for (const di of selectedBatch.value.items) {
    if (di.sales_order?.items) {
      for (const soi of di.sales_order.items) {
        totalItems++
        if (!checkedItems.value[soi.id]) return false
      }
    }
    if (di.sales_order?.bonus_items) {
      for (const b of di.sales_order.bonus_items) {
        totalItems++
        if (!checkedItems.value[b.id]) return false
      }
    }
  }
  return totalItems > 0
})

function toggleAllInNota(di: any, event: Event) {
  const isChecked = (event.target as HTMLInputElement).checked
  if (di.sales_order?.items) {
    di.sales_order.items.forEach((soi: any) => {
      checkedItems.value[soi.id] = isChecked
    })
  }
  if (di.sales_order?.bonus_items) {
    di.sales_order.bonus_items.forEach((b: any) => {
      checkedItems.value[b.id] = isChecked
    })
  }
}

function checkAllItems() {
  if (!selectedBatch.value || !selectedBatch.value.items) return
  const currentVal = allItemsChecked.value
  selectedBatch.value.items.forEach((di: any) => {
    if (di.sales_order?.items) {
      di.sales_order.items.forEach((soi: any) => {
        checkedItems.value[soi.id] = !currentVal
      })
    }
    if (di.sales_order?.bonus_items) {
      di.sales_order.bonus_items.forEach((b: any) => {
        checkedItems.value[b.id] = !currentVal
      })
    }
  })
}

// ─── Computed ─────────────────────────────────────────────────────────────────
const filteredBatches = computed(() =>
  batches.value.filter((b) => {
    const wBranchId = masterStore.selectedWarehouseBranchId
    if (
      wBranchId &&
      wBranchId !== 'ALL_BRANCHES' &&
      b.branch_id !== wBranchId
    ) {
      return false
    }
    if (activeTab.value === 'SUPERVISOR_APPROVED') {
      return b.status === 'SUPERVISOR_APPROVED' || b.status === 'PENDING'
    }
    return b.status === activeTab.value
  }),
)

const searchedBatches = computed(() => {
  const list = filteredBatches.value
  const q = tableSearch.value.trim().toLowerCase()
  if (!q) return list
  return list.filter((b) => {
    const fullName = `${b.driver?.first_name || ''} ${b.driver?.last_name || ''}`.toLowerCase()
    return (
      b.delivery_order_no?.toLowerCase().includes(q) ||
      fullName.includes(q) ||
      b.vehicle?.license_plate?.toLowerCase().includes(q)
    )
  })
})

const supervisorApprovedCount = computed(
  () =>
    batches.value.filter(
      (b) =>
        (!masterStore.selectedWarehouseBranchId ||
          masterStore.selectedWarehouseBranchId === 'ALL_BRANCHES' ||
          b.branch_id === masterStore.selectedWarehouseBranchId) &&
        (b.status === 'SUPERVISOR_APPROVED' || b.status === 'PENDING'),
    ).length,
)
const pickingCount = computed(
  () =>
    batches.value.filter(
      (b) =>
        (!masterStore.selectedWarehouseBranchId ||
          masterStore.selectedWarehouseBranchId === 'ALL_BRANCHES' ||
          b.branch_id === masterStore.selectedWarehouseBranchId) &&
        b.status === 'PICKING',
    ).length,
)
const onDeliveryCount = computed(
  () =>
    batches.value.filter(
      (b) =>
        (!masterStore.selectedWarehouseBranchId ||
          masterStore.selectedWarehouseBranchId === 'ALL_BRANCHES' ||
          b.branch_id === masterStore.selectedWarehouseBranchId) &&
        b.status === 'ON_DELIVERY',
    ).length,
)
const completedCount = computed(
  () =>
    batches.value.filter(
      (b) =>
        (!masterStore.selectedWarehouseBranchId ||
          masterStore.selectedWarehouseBranchId === 'ALL_BRANCHES' ||
          b.branch_id === masterStore.selectedWarehouseBranchId) &&
        b.status === 'COMPLETED',
    ).length,
)

// ─── Fetch ────────────────────────────────────────────────────────────────────
async function fetchBatches() {
  if (!masterStore.selectedWarehouseBranchId) return
  loading.value = true
  try {
    const branchId = masterStore.selectedWarehouseBranchId
    const params = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    
    // Fetch normal batches
    const res = await apiClient.get(`/delivery/batches/warehouse${params}`)
    const normalBatches = (res.data.data || []).map((b: any) => ({ ...b, source: 'nota' }))
    
    // Fetch MT orders
    let mtOrders = []
    try {
      const mtRes = await apiClient.get(`/mt-delivery/branch/inbox${params}`)
      mtOrders = (mtRes.data?.data || []).map((o: any) => ({
        ...o,
        source: 'mt',
        branch_id: o.source_branch_id,
        delivery_order_no: o.do_number,
        vehicle: null,
        driver: { first_name: o.key_account?.name || 'Sales MT', last_name: 'Order' },
        delivery_date: o.scheduled_date || o.shipped_at || o.created_at,
        status: o.status === 'ASSIGNED' ? 'PENDING' : 
                o.status === 'APPROVED' ? 'SUPERVISOR_APPROVED' : 
                o.status === 'LOADING' ? 'PICKING' :
                o.status === 'SHIPPED' ? 'ON_DELIVERY' : 
                o.status === 'DELIVERED' ? 'COMPLETED' : o.status
      }))
    } catch (e) {
      console.error('Failed to fetch MT orders', e)
    }

    // Fetch Factory Delegations
    let delegationGroups: any[] = []
    try {
      await warehouseStore.fetchAllTransfers()
      const groups: Record<string, any> = {}
      warehouseStore.allTransfers
        .filter((item) => item.fulfillment_branch_id === branchId)
        .forEach((item) => {
        const doNo = item.delivery_order_no || 'REQ-' + item.id.substring(0, 8)
        if (!groups[doNo]) {
          groups[doNo] = {
            id: doNo,
            source: 'delegation',
            branch_id: item.fulfillment_branch_id,
            delivery_order_no: doNo,
            vehicle: item.vehicle,
            driver: item.driver,
            is_delegation: true,
            status: item.status === 'APPROVED' ? 'SUPERVISOR_APPROVED' : 
                    item.status === 'LOADING' ? 'PICKING' : 
                    item.status === 'SHIPPED' ? 'ON_DELIVERY' : 
                    item.status === 'ARRIVED' || item.status === 'RECEIVED' ? 'COMPLETED' : item.status,
            delivery_date: item.target_shipment_date || item.shipped_at || item.created_at,
            created_at: item.created_at,
            items: []
          }
        }
        groups[doNo].items.push(item)
      })
      delegationGroups = Object.values(groups)
    } catch (e) {
      console.error('Failed to fetch delegations', e)
    }

    batches.value = [...normalBatches, ...mtOrders, ...delegationGroups].sort((a, b) => new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime())
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal memuat Surat Jalan')
  } finally {
    loading.value = false
  }
}

// ─── Scan Barcode ─────────────────────────────────────────────────────────────
async function handleScan() {
  if (!scanCode.value.trim()) {
    toast.error('Masukkan kode Surat Jalan')
    return
  }
  scanLoading.value = true
  scanResult.value = null
  scanError.value = ''
  try {
    const res = await apiClient.get(
      `/delivery/batch/scan?code=${encodeURIComponent(scanCode.value.trim())}`,
    )
    scanResult.value = res.data.data
    toast.success(` Surat Jalan ditemukan: ${scanResult.value.delivery_order_no}`)
  } catch {
    scanError.value = `Surat Jalan tidak ditemukan untuk kode: "${scanCode.value}"`
  } finally {
    scanLoading.value = false
  }
}

function handleScanKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') handleScan()
}

// ─── Confirm Ready ────────────────────────────────────────────────────────────
async function fetchProducts() {
  try {
    const branchId = selectedBatch.value?.branch_id || masterStore.selectedBranchId
    const params = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/factory/products${params}`)
    const rawList = res.data?.data || res.data || []
    products.value = filterSellableProducts(rawList)
  } catch {
    toast.error('Gagal memuat produk')
  }
}

function openDetail(batch: any) {
  if (batch.source === 'mt') {
    openMtDetail(batch)
    return
  }
  selectedBatch.value = batch
  checkedItems.value = {}
  activeModalTab.value = 'nota'
  showDetailModal.value = true
  fetchOrCreateDO(batch.id)
  fetchPendingLeftovers()
  
  // Populate spares
  localSpares.value = (batch.spares || []).map((s: any) => ({
    product_id: s.product_id,
    quantity: s.quantity
  }))
  fetchProducts()
}

async function fetchPendingLeftovers() {
  try {
    const branchId = masterStore.selectedBranchId
    const params = branchId && branchId !== 'ALL_BRANCHES' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/delivery/subsequent/pending${params}`)
    pendingLeftovers.value = res.data?.data || []
  } catch {
    toast.error('Gagal memuat daftar pengiriman tertunda/leftover')
  }
}

async function slipLeftover(subId: string) {
  if (!selectedBatch.value) return
  slippingLeftoverId.value = subId
  try {
    await apiClient.post(`/delivery/batch/${selectedBatch.value.id}/subsequent/${subId}/slip`)
    toast.success('Barang susulan/leftover berhasil diselipkan ke dalam pengiriman ini!')
    await fetchOrCreateDO(selectedBatch.value.id)
    await fetchPendingLeftovers()
    await fetchBatches()
    const updatedBatch = batches.value.find(b => b.id === selectedBatch.value.id)
    if (updatedBatch) {
      selectedBatch.value = updatedBatch
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyelipkan barang susulan')
  } finally {
    slippingLeftoverId.value = null
  }
}


function addSpare() {
  if (!newSpareProductId.value) {
    toast.error('Pilih produk terlebih dahulu')
    return
  }
  if (newSpareQty.value <= 0) {
    toast.error('Jumlah harus lebih dari 0')
    return
  }
  
  // Check if already exists
  const existingIndex = localSpares.value.findIndex(s => s.product_id === newSpareProductId.value)
  if (existingIndex > -1) {
    const item = localSpares.value[existingIndex]
    if (item) {
      item.quantity += newSpareQty.value
    }
  } else {
    localSpares.value.push({
      product_id: newSpareProductId.value,
      quantity: newSpareQty.value
    })
  }
  
  // Reset form
  newSpareProductId.value = ''
  newSpareQty.value = 1
}

function removeSpare(index: number) {
  localSpares.value.splice(index, 1)
}

async function saveSpares() {
  if (!selectedBatch.value) return
  savingSpares.value = true
  try {
    await apiClient.post(`/delivery/batch/${selectedBatch.value.id}/spares`, {
      spares: localSpares.value.map(s => ({
        product_id: s.product_id,
        quantity: Number(s.quantity) || 0
      }))
    })
    toast.success('Barang cadangan berhasil disimpan!')
    
    // Update the local list/batch data
    selectedBatch.value.spares = localSpares.value.map(s => {
      const p = products.value.find(prod => prod.id === s.product_id)
      return {
        product_id: s.product_id,
        quantity: s.quantity,
        product: p
      }
    })
    
    await fetchBatches()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyimpan barang cadangan')
  } finally {
    savingSpares.value = false
  }
}

async function fetchOrCreateDO(batchId: string) {
  loadingDO.value = true
  try {
    const res = await warehouseApi.createOrGetWarehouseDO(batchId)
    warehouseDO.value = res.data.data
    doNotes.value = res.data.data.notes || ''
    doItems.value = (res.data.data.items || []).map((item: any) => ({
      product_id: item.product_id,
      product: item.product,
      bonus_inventory_id: item.bonus_inventory_id,
      bonus_inventory: item.bonus_inventory,
      cartons_loaded: item.cartons_loaded,
      pcs_loaded: item.pcs_loaded
    }))
  } catch (e: any) {
    toast.error('Gagal memuat detail lembar DO (Bukti Muat)')
  } finally {
    loadingDO.value = false
  }
}

async function saveWarehouseDO() {
  if (!warehouseDO.value) return
  savingDO.value = true
  try {
    const payload = {
      id: warehouseDO.value.id,
      notes: doNotes.value,
      items: doItems.value.map((item) => ({
        product_id: item.product_id || undefined,
        bonus_inventory_id: item.bonus_inventory_id || undefined,
        cartons_loaded: Number(item.cartons_loaded) || 0,
        pcs_loaded: Number(item.pcs_loaded) || 0
      }))
    }
    await warehouseApi.updateWarehouseDO(payload)
    toast.success('Surat DO (Bukti Muat) berhasil disimpan!')
    // Update local copy
    if (selectedBatch.value) {
      await fetchOrCreateDO(selectedBatch.value.id)
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyimpan Surat DO')
  } finally {
    savingDO.value = false
  }
}

async function confirmWarehouseReady(batchId: string) {
  confirmingReady.value = true
  try {
    await apiClient.post(`/delivery/batch/${batchId}/warehouse-ready`)
    toast.success('Barang dikonfirmasi siap! Driver bisa mengambil.')
    showDetailModal.value = false
    await fetchBatches()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal mengkonfirmasi')
  } finally {
    confirmingReady.value = false
  }
}

const startingDelivery = ref(false)

async function startDelivery(doNo: string) {
  if (!confirm('Apakah Anda yakin ingin memulai pengiriman? Stok gudang akan otomatis dipotong.')) return
  
  startingDelivery.value = true
  try {
    await apiClient.post(`/delivery/batch/${doNo}/start`)
    toast.success('Pengiriman dimulai! Stok gudang berhasil dipotong.')
    showDetailModal.value = false
    await fetchBatches()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal memulai pengiriman')
  } finally {
    startingDelivery.value = false
  }
}

function handleQuickAction(row: any) {
  if (row.source === 'mt') {
    openMtDetail(row)
  } else {
    confirmWarehouseReady(row.id)
  }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
const formatDate = (d: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : '-'

const formatDeliveryDate = (d?: string) => {
  if (!d) return '-'
  const date = new Date(d)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const totalItemsInBatch = (batch: any) => {
  if (batch.source === 'mt') return batch.items?.length || 0;
  return batch.items?.reduce((sum: number, di: any) => sum + (di.sales_order?.items?.length || 0), 0) || 0;
}

const totalQtyInBatch = (batch: any) => {
  if (batch.source === 'mt') return batch.items?.reduce((sum: number, item: any) => sum + (item.requested_qty || 0), 0) || 0;
  let qty = 0
  batch.items?.forEach((di: any) => {
    di.sales_order?.items?.forEach((soi: any) => {
      qty += soi.quantity || 0
    })
  })
  return qty
}

// ─── MT DO Handlers ─────────────────────────────────────────────────────────────
const openMtDetail = async (order: any) => {
  mtWorkingOrder.value = order
  
  if (order.status === 'PENDING') mtAction.value = 'APPROVE'
  else if (order.status === 'SUPERVISOR_APPROVED') mtAction.value = 'LOADING'
  else if (order.status === 'PICKING') mtAction.value = 'SHIP'
  else mtAction.value = 'DETAIL'

  if (mtAction.value === 'APPROVE') {
    mtApproveBy.value = authStore.user?.name || authStore.user?.first_name || authStore.user?.email || 'Admin Gudang'
  }
  if (mtAction.value === 'LOADING') {
    mtLoadingForm.value = {
      prepared_by: '',
      items: (order.items || []).map((i: any) => ({
        product_id: i.product_id,
        prepared_qty: i.requested_qty || 0,
        shipped_qty: i.requested_qty || 0,
        batch_no: i.batch_no || 'DEFAULT',
        notes: '',
      })),
    }
  }
  if (mtAction.value === 'SHIP') {
    mtShipForm.value = { vehicle_id: '', driver_id: '', driver_name: '', notes: '' }
    if (mtVehicles.value.length === 0) await fetchVehicles()
    if (mtDrivers.value.length === 0) await fetchDrivers()
  }

  showMtDetailModal.value = true
}

const handleMtSubmit = async () => {
  if (!mtWorkingOrder.value) return
  mtIsSubmitting.value = true
  try {
    if (mtAction.value === 'APPROVE') {
      await apiClient.post(`/mt-delivery/${mtWorkingOrder.value.id}/approve`, { approved_by: mtApproveBy.value || 'Gudang' })
      toast.success('MT DO disetujui! Siapkan barang.')
    } else if (mtAction.value === 'LOADING') {
      await apiClient.post(`/mt-delivery/${mtWorkingOrder.value.id}/loading`, mtLoadingForm.value)
      toast.success('Status pemuatan barang diperbarui!')
    } else if (mtAction.value === 'SHIP') {
      await apiClient.post(`/mt-delivery/${mtWorkingOrder.value.id}/ship`, {
        vehicle_id: mtShipForm.value.vehicle_id || null,
        driver_id: mtShipForm.value.driver_id || null,
        notes: (!mtShipForm.value.driver_id && mtShipForm.value.driver_name)
          ? `Supir Luar: ${mtShipForm.value.driver_name}. ${mtShipForm.value.notes}`
          : mtShipForm.value.notes,
      })
      toast.success('MT DO berhasil dikirim! 🚛')
    }
    showMtDetailModal.value = false
    await fetchBatches()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal memproses MT DO')
  } finally {
    mtIsSubmitting.value = false
  }
}

// ─── DELEGATION SELECTION LOGIC ──────────────────────────────────────────────
const showProcessModal = ref(false)
const delegationInventory = ref<any[]>([])
const processForm = ref<{
  id: string
  doNo: string
  vehicle_id: string
  driver_id: string
  branch_id?: string
  items: Array<{
    id: string
    product_id: string
    product_name: string
    quantity: number
    unit: string
    pcs_per_unit: number
    batch_no: string
  }>
}>({
  id: '',
  doNo: '',
  vehicle_id: '',
  driver_id: '',
  items: [],
})

const availableVehicles = computed(() => {
  return vehicleStore.vehicles.filter((v: any) => v.status === 'AVAILABLE')
})

const availableDrivers = computed(() => {
  return masterStore.employees.filter((emp: any) => {
    const jobTitle = (emp.job_position?.title || '').toLowerCase()
    return jobTitle.includes('driver') || jobTitle.includes('sopir') || jobTitle.includes('supir')
  })
})

const getBranchBatchesForProduct = (productId: string) => {
  return delegationInventory.value
    .filter(
      (s: any) =>
        s.product_id === productId &&
        (!s.stock_type || s.stock_type === 'GOOD') &&
        s.quantity > 0,
    )
    .map((s: any) => ({
      batch_no: (s.batch_no || 'DEFAULT').trim() || 'DEFAULT',
      quantity: s.quantity,
    }))
}

const openProcessModal = async (row: any) => {
  const branchId = row.branch_id || masterStore.selectedWarehouseBranchId
  try {
    const res = await warehouseApi.getInventory(branchId)
    delegationInventory.value = res.data || []
  } catch (e) {
    console.error('Failed to load branch inventory', e)
  }

  processForm.value = {
    id: row.id,
    doNo: row.delivery_order_no,
    vehicle_id: row.vehicle?.id || '',
    driver_id: row.driver?.id || '',
    branch_id: branchId,
    items: (row.items || []).map((i: any) => {
      const availBatches = getBranchBatchesForProduct(i.product_id)
      const selectedBatch = i.batch_no || (availBatches[0]?.batch_no || '')
      return {
        id: i.id,
        product_id: i.product_id,
        product_name: i.product?.name || 'Produk',
        quantity: i.quantity,
        unit: i.unit || 'PCS',
        pcs_per_unit: i.pcs_per_unit || 1,
        batch_no: selectedBatch,
      }
    }),
  }
  showProcessModal.value = true
}

const submitProcessDelegation = async () => {
  if (!processForm.value.vehicle_id || !processForm.value.driver_id) {
    toast.error('Harap pilih armada dan sopir pengiriman.')
    return
  }

  for (const item of processForm.value.items || []) {
    if (!item.batch_no) {
      toast.error(`Harap pilih batch fisik untuk produk '${item.product_name}'.`)
      return
    }
  }

  const group = batches.value.find((b) => b.id === processForm.value.id)
  if (!group || !group.is_delegation) return

  try {
    for (const item of processForm.value.items) {
      const originalItem = group.items.find((i: any) => i.id === item.id)
      await apiClient.put(`/factory/transfer/${item.id}`, {
        vehicle_id: processForm.value.vehicle_id,
        driver_id: processForm.value.driver_id,
        batch_no: item.batch_no,
        status: 'LOADING',
        target_shipment_date: originalItem?.target_shipment_date || null,
        estimated_arrival: originalItem?.estimated_arrival || null,
        notes: originalItem?.notes || '',
      })
    }

    toast.success('Tugas pengiriman & batch berhasil dialokasikan, silakan muat barang.')
    showProcessModal.value = false
    await fetchBatches()
  } catch (err: any) {
    let msg = err
    if (typeof err === 'string') {
      msg = err
    } else if (err.response?.data?.error) {
      msg = err.response.data.error
    } else if (err.message) {
      msg = err.message
    } else {
      try {
        msg = JSON.stringify(err)
      } catch (e) {}
    }
    toast.error('Gagal menugaskan delegasi: ' + msg)
  }
}

const dispatchDelegation = async (row: any) => {
  try {
    for (const item of row.items) {
      await warehouseStore.confirmShipment(item.id)
    }
    toast.success('Stok gudang berhasil dipotong dan pengiriman delegasi dimulai.')
    await fetchBatches()
  } catch (err: any) {
    let msg = err
    if (typeof err === 'string') {
      msg = err
    } else if (err.response?.data?.error) {
      msg = err.response.data.error
    } else if (err.message) {
      msg = err.message
    } else {
      try { msg = JSON.stringify(err) } catch(e) {}
    }
    toast.error('Gagal memulai pengiriman delegasi: ' + msg)
  }
}

const currentBranchName = computed(() => {
  const id = masterStore.selectedWarehouseBranchId
  if (!id || id === 'ALL_BRANCHES') return 'Semua Cabang'
  const branch = masterStore.branches.find((b) => b.id === id)
  return branch ? branch.name : 'Semua Cabang'
})

watch(
  () => masterStore.selectedWarehouseBranchId,
  (newId) => {
    if (newId) fetchBatches()
  },
)

onMounted(() => {
  fetchBatches()
  masterStore.fetchEmployees()
  vehicleStore.fetchVehicles()
  masterStore.fetchBranches()
  if (masterStore.selectedWarehouseBranchId) fetchBatches()
})
</script><template>
  <div class="space-y-4 p-4 md:p-6">
    <!-- ─── Header ─────────────────────────────────────────────────── -->
    <div
      class="flex flex-col sm:flex-row justify-between items-start sm:items-center bg-white text-slate-900 p-4 md:p-6 rounded-2xl shadow-sm border border-slate-100 gap-4"
    >
      <div>
        <h1 class="text-xl md:text-2xl font-black text-slate-900 tracking-tight flex items-center gap-2">
          Surat Jalan Masuk Gudang - <span class="text-primary font-extrabold">{{ currentBranchName }}</span>
        </h1>
        <p class="text-xs text-slate-500 mt-0.5">
          Terima &amp; siapkan barang berdasarkan Surat Jalan yang diterbitkan supervisor.
        </p>
      </div>
      <button
        @click="fetchBatches"
        class="flex items-center gap-1.5 px-4 py-2 bg-white border border-slate-200 rounded-xl text-xs font-bold text-slate-650 hover:bg-slate-50 transition-colors shadow-sm"
      >
        <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" />
        Segarkan
      </button>
    </div>

    <!-- ─── Top Dashboard: Scan & Stats ────────────────────────────── -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <!-- Scan Barcode Panel (Left) -->
      <div
        class="md:col-span-2 bg-gradient-to-br from-primary to-red-800 rounded-2xl p-4 shadow-sm shadow-primary/5 flex flex-col justify-center relative overflow-hidden"
      >
        <div
          class="absolute top-0 right-0 w-32 h-32 bg-white/5 rounded-full blur-2xl -translate-y-1/2 translate-x-1/3"
        ></div>

        <div class="flex items-center gap-2.5 mb-2 relative z-10">
          <div
            class="w-8 h-8 bg-white/20 backdrop-blur-md rounded-lg flex items-center justify-center shrink-0 border border-white/10"
          >
            <ScanLine class="w-4 h-4 text-white" />
          </div>
          <div>
            <p class="text-white font-black text-sm leading-none">Scan Surat Jalan</p>
            <p class="text-red-200 text-[10px] font-medium mt-1">
              Gunakan scanner atau input nomor resi secara manual
            </p>
          </div>
        </div>

        <div class="flex gap-2 relative z-10">
          <div class="flex-1 relative group">
            <QrCode
              class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-red-300 transition-colors"
            />
            <input
              v-model="scanCode"
              @keydown="handleScanKeydown"
              type="text"
              placeholder="Contoh: SJ-20250519-abc123"
              class="w-full pl-9 pr-3 py-2 bg-white/15 backdrop-blur-md border border-white/20 rounded-xl text-white placeholder-red-200/70 font-semibold text-xs focus:outline-none focus:bg-white/25 focus:border-white/40 transition-all"
            />
          </div>
          <button
            @click="handleScan"
            :disabled="scanLoading"
            class="px-4 py-2 bg-white text-red-750 font-black text-xs rounded-xl hover:bg-red-50 disabled:opacity-50 transition-all shadow-sm flex items-center justify-center gap-1.5 shrink-0"
          >
            <Search class="w-3.5 h-3.5" />
            <span>{{ scanLoading ? '...' : 'Cari' }}</span>
          </button>
        </div>

        <!-- Scan Result -->
        <div
          v-if="scanResult"
          class="mt-2.5 bg-white/20 backdrop-blur-md rounded-xl p-2.5 border border-white/20 flex items-center justify-between gap-3 relative z-10 shadow-sm"
        >
          <div>
            <p class="text-white font-black text-sm leading-tight">
              {{ scanResult.delivery_order_no }}
            </p>
            <div class="flex items-center gap-1.5 mt-0.5 text-red-100 text-[10px] font-semibold">
              <span class="flex items-center gap-1"
                ><User class="w-2.5 h-2.5" /> {{ scanResult.driver?.first_name }}</span
              >
              <span>•</span>
              <span>{{ scanResult.items?.length || 0 }} nota</span>
            </div>
          </div>
          <div class="flex items-center gap-1.5">
            <button
              @click="openDetail(scanResult)"
              class="flex items-center justify-center gap-1 px-3 py-1.5 bg-white text-red-700 font-bold text-[10px] rounded-lg hover:bg-red-50 transition-colors shadow-xs"
            >
              Detail <ChevronRight class="w-3 h-3" />
            </button>
          </div>
        </div>
        <div
          v-if="scanError"
          class="mt-2 bg-red-900/40 backdrop-blur-md border border-red-400/20 rounded-xl p-2 flex items-center gap-2 relative z-10"
        >
          <AlertCircle class="w-3.5 h-3.5 text-red-300 shrink-0" />
          <p class="text-red-100 text-[11px] font-bold">{{ scanError }}</p>
        </div>
      </div>

      <!-- Stats Cards (Right) -->
      <div class="flex flex-col sm:flex-row md:flex-col gap-3">
        <div
          class="bg-white border border-slate-100 hover:border-amber-200 rounded-2xl p-4 shadow-sm hover:shadow flex-1 flex flex-col justify-center relative overflow-hidden group transition-all"
        >
          <div class="flex items-center justify-between relative z-10">
            <div>
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-wider mb-0.5">
                Perlu Disiapkan
              </p>
              <p class="text-2xl font-black text-slate-800 tracking-tight">
                {{ supervisorApprovedCount }}<span class="text-xs text-slate-400 font-bold ml-1">SJ</span>
              </p>
            </div>
            <div
              class="w-9 h-9 bg-amber-50 rounded-lg flex items-center justify-center shrink-0 border border-amber-100"
            >
              <Clock class="w-5 h-5 text-amber-600" />
            </div>
          </div>
        </div>

        <div
          class="bg-white border border-slate-100 hover:border-red-200 rounded-2xl p-4 shadow-sm hover:shadow flex-1 flex flex-col justify-center relative overflow-hidden group transition-all"
        >
          <div class="flex items-center justify-between relative z-10">
            <div>
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-wider mb-0.5">
                Sedang Disiapkan
              </p>
              <p class="text-2xl font-black text-slate-800 tracking-tight">
                {{ pickingCount }}<span class="text-xs text-slate-400 font-bold ml-1">SJ</span>
              </p>
            </div>
            <div
              class="w-9 h-9 bg-red-50 rounded-lg flex items-center justify-center shrink-0 border border-red-100"
            >
              <BoxSelect class="w-5 h-5 text-primary" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ─── Tabs & Search Bar ──────────────────────────────────────── -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white p-3 border border-slate-100 rounded-2xl shadow-xs">
      <div class="flex flex-wrap gap-2">
        <button
          @click="activeTab = 'SUPERVISOR_APPROVED'"
          :class="
            activeTab === 'SUPERVISOR_APPROVED'
              ? 'bg-amber-500 text-white shadow-sm'
              : 'bg-slate-50 border border-slate-200 text-slate-500 hover:bg-slate-100'
          "
          class="flex items-center gap-1.5 px-4 py-2.5 rounded-xl text-xs font-black transition-all"
        >
          <Clock class="w-3.5 h-3.5" />
          Perlu Disiapkan
          <span
            v-if="supervisorApprovedCount > 0"
            class="bg-white/20 text-[10px] font-black px-1.5 py-0.5 rounded-full ml-1"
          >
            {{ supervisorApprovedCount }}
          </span>
        </button>
        <button
          @click="activeTab = 'PICKING'"
          :class="
            activeTab === 'PICKING'
              ? 'bg-primary text-white shadow-sm'
              : 'bg-slate-50 border border-slate-200 text-slate-500 hover:bg-slate-100'
          "
          class="flex items-center gap-1.5 px-4 py-2.5 rounded-xl text-xs font-black transition-all"
        >
          <BoxSelect class="w-3.5 h-3.5" />
          Sedang Disiapkan
          <span
            v-if="pickingCount > 0"
            class="bg-white/20 text-[10px] font-black px-1.5 py-0.5 rounded-full ml-1"
          >
            {{ pickingCount }}
          </span>
        </button>
        <button
          @click="activeTab = 'ON_DELIVERY'"
          :class="
            activeTab === 'ON_DELIVERY'
              ? 'bg-blue-600 text-white shadow-sm'
              : 'bg-slate-50 border border-slate-200 text-slate-500 hover:bg-slate-100'
          "
          class="flex items-center gap-1.5 px-4 py-2.5 rounded-xl text-xs font-black transition-all"
        >
          <Truck class="w-3.5 h-3.5" />
          Sedang Dikirim (Lacak)
          <span
            v-if="onDeliveryCount > 0"
            class="bg-white/20 text-[10px] font-black px-1.5 py-0.5 rounded-full ml-1"
          >
            {{ onDeliveryCount }}
          </span>
        </button>
        <button
          @click="activeTab = 'COMPLETED'"
          :class="
            activeTab === 'COMPLETED'
              ? 'bg-green-600 text-white shadow-sm'
              : 'bg-slate-50 border border-slate-200 text-slate-500 hover:bg-slate-100'
          "
          class="flex items-center gap-1.5 px-4 py-2.5 rounded-xl text-xs font-black transition-all"
        >
          <CheckCircle2 class="w-3.5 h-3.5" />
          Selesai
          <span
            v-if="completedCount > 0"
            class="bg-white/20 text-[10px] font-black px-1.5 py-0.5 rounded-full ml-1"
          >
            {{ completedCount }}
          </span>
        </button>
      </div>

      <!-- Search Input -->
      <div class="relative w-full sm:w-[260px]">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input
          v-model="tableSearch"
          type="text"
          placeholder="Cari No. SJ, Driver, Plat..."
          class="w-full pl-9 pr-3 py-2 border border-slate-200 rounded-xl text-xs font-semibold focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary"
        />
      </div>
    </div>

    <!-- ─── Loading ────────────────────────────────────────────────── -->
    <div v-if="loading" class="flex flex-col items-center py-20 text-slate-400 gap-2">
      <RefreshCw class="w-6 h-6 animate-spin text-slate-400" />
      <p class="text-xs font-bold">Memuat Surat Jalan...</p>
    </div>

    <!-- ─── Empty ──────────────────────────────────────────────────── -->
    <div
      v-else-if="searchedBatches.length === 0"
      class="text-center py-16 bg-slate-50 rounded-[30px] border-2 border-dashed border-slate-100"
    >
      <div
        class="w-16 h-16 bg-white rounded-2xl shadow-xs flex items-center justify-center mx-auto mb-4"
      >
        <CheckCircle2 class="w-8 h-8 text-green-300" />
      </div>
      <h3 class="text-base font-black text-slate-800">
        {{
          activeTab === 'SUPERVISOR_APPROVED'
            ? 'Tidak Ada SJ Yang Perlu Disiapkan'
            : activeTab === 'PICKING'
              ? 'Tidak Ada Yang Sedang Disiapkan'
              : activeTab === 'ON_DELIVERY'
                ? 'Tidak Ada Pengiriman Sedang Berlangsung'
                : 'Belum Ada Pengiriman Selesai'
        }}
      </h3>
      <p class="text-slate-400 text-xs mt-1">
        {{
          activeTab === 'SUPERVISOR_APPROVED'
            ? 'Semua Surat Jalan sudah selesai diproses.'
            : activeTab === 'PICKING'
              ? 'Belum ada barang yang sedang dalam proses picking.'
              : activeTab === 'ON_DELIVERY'
                ? 'Semua armada pengiriman sedang berada di gudang.'
                : 'Belum ada riwayat pengiriman selesai hari ini.'
        }}
      </p>
    </div>

    <!-- ─── Batch Table ────────────────────────────────────────────── -->
    <div
      v-else
      class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden p-1.5"
    >
      <DataTable :columns="batchColumns" :data="searchedBatches">
        <template #cell-delivery_order_no="{ row }">
          <div class="flex flex-col">
            <span
              class="font-mono text-sm font-black text-slate-900 cursor-pointer hover:text-primary transition-colors flex items-center gap-1.5"
              @click="openDetail(row)"
            >
              {{ row.delivery_order_no }}
              <span v-if="row.source === 'mt'" class="bg-green-100 text-green-700 text-[9px] px-1.5 py-0.5 rounded uppercase tracking-widest font-bold">MT</span>
              <span v-else-if="row.source === 'delegation'" class="bg-orange-100 text-orange-700 border border-orange-200 text-[9px] px-1.5 py-0.5 rounded uppercase tracking-widest font-black">DELEGASI PABRIK</span>
              <span v-else class="bg-blue-100 text-blue-700 text-[9px] px-1.5 py-0.5 rounded uppercase tracking-widest font-bold">Nota</span>
            </span>
            <span v-if="row.source === 'mt'" class="text-[10px] text-slate-400 font-bold mt-0.5 flex items-center gap-1"><FileCheck2 class="w-3 h-3 text-slate-400" /> Dokumen DO Tersedia</span>
            <span v-else-if="row.source === 'delegation'" class="text-[10px] text-orange-500/80 font-bold mt-0.5">Pengiriman cabang ke cabang</span>
            <span v-else class="text-[10px] text-slate-400 font-bold mt-0.5"
              >{{ row.items?.length || 0 }} nota</span
            >
          </div>
        </template>
        <template #cell-delivery_date="{ row }">
          <div class="flex items-center gap-1.5 text-xs font-semibold text-slate-700 whitespace-nowrap">
            <Calendar class="w-3.5 h-3.5 text-slate-400 shrink-0" />
            <span>{{ formatDeliveryDate(row.delivery_date || row.created_at) }}</span>
          </div>
        </template>
        <template #cell-status="{ row }">
          <span
            class="text-[10px] font-black px-3 py-1.5 rounded-xl uppercase tracking-widest whitespace-nowrap"
            :class="
              row.status === 'SUPERVISOR_APPROVED' || row.status === 'PENDING'
                ? 'bg-amber-100 text-amber-700'
                : row.status === 'PICKING'
                  ? 'bg-red-100 text-red-700'
                  : row.status === 'ON_DELIVERY'
                    ? 'bg-blue-100 text-blue-700'
                    : 'bg-green-100 text-green-700'
            "
          >
            {{
              row.status === 'SUPERVISOR_APPROVED' || row.status === 'PENDING'
                ? '⏳ Perlu Disiapkan'
                : row.status === 'PICKING'
                  ? '📦 Picking'
                  : row.status === 'ON_DELIVERY'
                    ? '🚚 Pengiriman'
                    : '✅ Selesai'
            }}
          </span>
        </template>
        <template #cell-driver="{ row }">
          <div class="flex items-center gap-2">
            <div
              class="w-8 h-8 bg-slate-100 rounded-full flex items-center justify-center shrink-0"
            >
              <User class="w-4 h-4 text-slate-400" />
            </div>
            <div class="flex flex-col">
              <span class="font-bold text-slate-900 text-sm whitespace-nowrap leading-tight">
                {{ row.driver?.first_name || row.daily_driver?.name || 'Belum Assign' }} {{ row.driver?.last_name || '' }}
              </span>
              <span v-if="row.source === 'mt'" class="text-[10px] font-bold text-slate-500 mt-0.5">Sales MT PIC</span>
              <span v-if="row.helper || row.daily_helper" class="text-[10px] font-bold text-slate-500 mt-0.5">
                + Helper: {{ row.helper?.first_name || row.daily_helper?.name }}
              </span>
            </div>
          </div>
        </template>
        <template #cell-vehicle="{ row }">
          <span
            class="text-xs font-bold text-slate-600 flex items-center gap-1.5 whitespace-nowrap"
          >
            <Truck class="w-4 h-4 text-slate-400" />
            {{ row.vehicle?.license_plate || 'Belum assign' }}
          </span>
        </template>
        <template #cell-stats="{ row }">
          <div class="flex flex-col gap-0.5" v-if="row.source === 'delegation'">
            <span class="text-xs font-bold text-slate-700">{{ row.items?.length || 0 }} Item</span>
            <span class="text-[10px] font-medium text-slate-400"
              >{{ row.items?.reduce((sum: number, i: any) => sum + (i.quantity || 0), 0) || 0 }} Qty Total</span
            >
          </div>
          <div class="flex flex-col gap-0.5" v-else>
            <span class="text-xs font-bold text-slate-700">{{ totalItemsInBatch(row) }} SKU</span>
            <span class="text-[10px] font-medium text-slate-400"
              >{{ totalQtyInBatch(row) }} Qty Total</span
            >
          </div>
        </template>
        <template #cell-actions="{ row }">
          <div class="flex items-center gap-2">
            <div v-if="row.source === 'delegation'" class="flex items-center gap-2">
              <button
                v-if="row.status === 'SUPERVISOR_APPROVED'"
                @click="openProcessModal(row)"
                class="bg-orange-500 hover:bg-orange-600 text-white rounded-xl px-4 py-2 font-black text-xs flex items-center gap-2 transition-all shadow-md shadow-orange-200 whitespace-nowrap"
              >
                <Truck class="w-4 h-4" /> Tugaskan & Muat
              </button>
              <button
                v-else-if="row.status === 'PICKING'"
                @click="dispatchDelegation(row)"
                class="bg-blue-600 hover:bg-blue-700 text-white rounded-xl px-4 py-2 font-black text-xs flex items-center gap-2 transition-all shadow-md shadow-blue-200 whitespace-nowrap"
              >
                <Truck class="w-4 h-4" /> Mulai Pengiriman
              </button>
              <div
                v-else-if="row.status === 'ON_DELIVERY'"
                class="bg-blue-50 border border-blue-200 rounded-xl px-4 py-2 flex items-center gap-2 whitespace-nowrap w-max"
              >
                <Truck class="w-4 h-4 text-blue-600" />
                <span class="text-blue-600 font-black text-xs">Sedang Dikirim</span>
              </div>
              <div
                v-else-if="row.status === 'COMPLETED'"
                class="bg-green-50 border border-green-200 rounded-xl px-4 py-2 flex items-center gap-2 whitespace-nowrap w-max"
              >
                <CheckCircle2 class="w-4 h-4 text-green-600" />
                <span class="text-green-600 font-black text-xs">Selesai & Diterima</span>
              </div>
            </div>
            <template v-else>
              <button
                @click="openDetail(row)"
                class="px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-650 font-bold text-[10px] uppercase tracking-wider rounded-xl transition-colors"
              >
                Detail &amp; DO
              </button>
              <button
                v-if="row.status === 'SUPERVISOR_APPROVED' || row.status === 'PENDING'"
                @click.stop="handleQuickAction(row)"
                :disabled="confirmingReady"
                class="bg-amber-500 hover:bg-amber-600 text-white rounded-xl px-4 py-2 font-black text-xs flex items-center gap-2 transition-all shadow-md shadow-amber-200 disabled:opacity-50 whitespace-nowrap"
              >
                <CheckCircle2 class="w-4 h-4" />
                {{ row.source === 'mt' ? (row.status === 'PENDING' ? 'Approve DO' : 'Muat Barang') : 'Konfirmasi Siap' }}
              </button>
              <button
                v-else-if="row.status === 'PICKING' && row.source !== 'mt' && !row.is_delegation"
                @click.stop="startDelivery(row.delivery_order_no)"
                :disabled="startingDelivery"
                class="bg-blue-600 hover:bg-blue-700 text-white rounded-xl px-4 py-2 font-black text-xs flex items-center gap-2 transition-all shadow-md shadow-blue-200 disabled:opacity-50 whitespace-nowrap"
              >
                <Truck class="w-4 h-4" />
                Mulai Kirim (Potong Stok)
              </button>
              <button
                v-else-if="row.status === 'PICKING' && row.source === 'mt'"
                @click.stop="handleQuickAction(row)"
                class="bg-blue-600 hover:bg-blue-700 text-white rounded-xl px-4 py-2 font-black text-xs flex items-center gap-2 transition-all shadow-md shadow-blue-200 whitespace-nowrap"
              >
                <Truck class="w-4 h-4" />
                Mulai Kirim MT
              </button>
              <div
                v-else-if="row.status === 'PICKING' && row.is_delegation"
                class="bg-red-50 border border-red-200 rounded-xl px-4 py-2 flex items-center gap-2 whitespace-nowrap w-max"
              >
                <BoxSelect class="w-4 h-4 text-primary" />
                <span class="text-primary font-black text-xs">Sedang Disiapkan</span>
              </div>
              <div
                v-else-if="row.status === 'ON_DELIVERY'"
                class="bg-blue-50 border border-blue-200 rounded-xl px-4 py-2 flex items-center gap-2 whitespace-nowrap w-max"
              >
                <Truck class="w-4 h-4 text-blue-600" />
                <span class="text-blue-600 font-black text-xs">Dalam Perjalanan</span>
              </div>
              <div
                v-else-if="row.status === 'COMPLETED'"
                class="bg-green-50 border border-green-200 rounded-xl px-4 py-2 flex items-center gap-2 whitespace-nowrap w-max"
              >
                <CheckCircle2 class="w-4 h-4 text-green-600" />
                <span class="text-green-600 font-black text-xs">Selesai Diterima</span>
              </div>
            </template>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- ─── Detail Modal ───────────────────────────────────────────── -->
    <div
      v-if="showDetailModal && selectedBatch"
      class="fixed inset-0 z-50 flex items-center justify-center p-4"
    >
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showDetailModal = false"
      />
      <div
        class="bg-white rounded-[32px] w-full max-w-2xl relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden"
      >
        <!-- Header -->
        <div
          class="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-gradient-to-r from-red-50 to-slate-50"
        >
          <div>
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
              SURAT JALAN
            </p>
            <h3 class="text-xl font-black text-slate-900">{{ selectedBatch.delivery_order_no }}</h3>
            <div class="flex items-center gap-2 mt-1">
              <span
                class="text-[9px] font-black px-2 py-1 rounded-lg uppercase"
                :class="
                  selectedBatch.status === 'SUPERVISOR_APPROVED' || selectedBatch.status === 'PENDING'
                    ? 'bg-amber-100 text-amber-700'
                    : selectedBatch.status === 'PICKING'
                      ? 'bg-red-100 text-red-700'
                      : selectedBatch.status === 'ON_DELIVERY'
                        ? 'bg-blue-100 text-blue-700'
                        : 'bg-green-100 text-green-700'
                "
              >
                {{
                  selectedBatch.status === 'SUPERVISOR_APPROVED' || selectedBatch.status === 'PENDING'
                    ? '⏳ Perlu Disiapkan'
                    : selectedBatch.status === 'PICKING'
                      ? '📦 Picking'
                      : selectedBatch.status === 'ON_DELIVERY'
                        ? '🚚 Pengiriman'
                        : '✅ Selesai'
                }}
              </span>
              <span class="text-xs text-slate-400 font-medium">
                {{ selectedBatch.driver?.first_name || selectedBatch.daily_driver?.name || 'Belum Assign' }} {{ selectedBatch.driver?.last_name || '' }} •
                {{ selectedBatch.vehicle?.license_plate || 'Armada blm ditentukan' }} •
                Tgl Kirim: {{ formatDeliveryDate(selectedBatch.delivery_date || selectedBatch.created_at) }}
              </span>
            </div>
          </div>
          <button
            @click="showDetailModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Navigation Tabs -->
        <div class="flex border-b border-slate-100 px-6 bg-slate-50/30">
          <button
            @click="activeModalTab = 'nota'"
            :class="
              activeModalTab === 'nota'
                ? 'border-primary text-primary'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            "
            class="py-3 px-4 text-xs font-black border-b-2 -mb-px transition-colors flex items-center gap-1.5 focus:outline-none"
          >
            <FileText class="w-3.5 h-3.5" />
            Rincian Nota &amp; Checklist
          </button>
          <button
            @click="activeModalTab = 'do'"
            :class="
              activeModalTab === 'do'
                ? 'border-primary text-primary'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            "
            class="py-3 px-4 text-xs font-black border-b-2 -mb-px transition-colors flex items-center gap-1.5 focus:outline-none"
          >
            <ShieldCheck class="w-3.5 h-3.5" />
            Surat Lembar DO (Bukti Muat)
            <span class="bg-amber-100 text-amber-800 text-[9px] px-1.5 py-0.5 rounded-full font-bold">Baru</span>
          </button>
          <button
            @click="activeModalTab = 'spares'"
            :class="
              activeModalTab === 'spares'
                ? 'border-primary text-primary'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            "
            class="py-3 px-4 text-xs font-black border-b-2 -mb-px transition-colors flex items-center gap-1.5 focus:outline-none"
          >
            <BoxSelect class="w-3.5 h-3.5" />
            Barang Cadangan (Spares)
          </button>
          <button
            @click="activeModalTab = 'leftovers'"
            :class="
              activeModalTab === 'leftovers'
                ? 'border-primary text-primary'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            "
            class="py-3 px-4 text-xs font-black border-b-2 -mb-px transition-colors flex items-center gap-1.5 focus:outline-none"
          >
            <Clock class="w-3.5 h-3.5" />
            Leftover / Susulan
            <span
              v-if="pendingLeftovers.length > 0"
              class="bg-rose-100 text-rose-800 text-[9px] px-1.5 py-0.5 rounded-full font-bold ml-1"
            >
              {{ pendingLeftovers.length }}
            </span>
          </button>
        </div>

        <!-- Body -->
        <div class="flex-1 overflow-y-auto p-6">
          <!-- TAB 1: Rincian Nota & Checklist -->
          <div v-if="activeModalTab === 'nota'" class="space-y-4">
            <div class="flex items-center justify-between">
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                Daftar Item Per Nota
              </p>
              <Button
                @click="checkAllItems"
                size="sm"
                variant="outline"
                class="h-7 rounded-lg text-[10px] font-bold border-slate-200 hover:bg-slate-50 text-slate-600 px-2"
              >
                {{ allItemsChecked ? 'Hapus Semua Centang' : 'Centang Semua Item' }}
              </Button>
            </div>

            <div
              v-for="(di, seqIdx) in selectedBatch.items"
              :key="di.id"
              class="bg-slate-50 rounded-2xl p-4 border border-slate-100"
            >
              <!-- Nota header -->
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-2">
                  <span
                    class="w-6 h-6 bg-primary text-white rounded-full text-[10px] font-black flex items-center justify-center"
                  >
                    {{ Number(seqIdx) + 1 }}
                  </span>
                  <div>
                    <p class="font-black text-slate-900 text-sm leading-tight">
                      {{ di.sales_order?.store?.name || 'Unknown Store' }}
                    </p>
                    <p class="font-mono text-[10px] text-slate-400">
                      {{ di.sales_order?.so_number }}
                    </p>
                  </div>
                </div>
                <span
                  class="text-[9px] font-black px-2 py-1 rounded-lg"
                  :class="
                    di.status === 'DELIVERED'
                      ? 'bg-green-100 text-green-700'
                      : 'bg-slate-200 text-slate-500'
                  "
                >
                  {{ di.status }}
                </span>
              </div>

              <!-- Product list Table -->
              <div class="mt-3 overflow-x-auto">
                <table class="w-full text-left border-collapse">
                  <thead>
                    <tr
                      class="border-b border-slate-200 text-[10px] uppercase tracking-widest text-slate-400"
                    >
                      <th class="py-2 pl-2 w-10">
                        <input
                          type="checkbox"
                          @change="toggleAllInNota(di, $event)"
                          class="rounded border-slate-300 text-primary focus:ring-primary w-4 h-4 cursor-pointer"
                        />
                      </th>
                      <th class="py-2">SKU</th>
                      <th class="py-2">Produk</th>
                      <th class="py-2 text-right pr-2">Qty</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="soi in di.sales_order?.items"
                      :key="soi.id"
                      class="border-b border-slate-100 last:border-0 hover:bg-slate-100/50 transition-colors"
                    >
                      <td class="py-2 pl-2">
                        <input
                          type="checkbox"
                          v-model="checkedItems[soi.id]"
                          class="rounded border-slate-300 text-primary focus:ring-primary w-4 h-4 cursor-pointer"
                        />
                      </td>
                      <td class="py-2 font-mono text-[10px] text-slate-500">
                        {{ soi.product?.sku }}
                      </td>
                      <td class="py-2 text-xs font-semibold text-slate-700">
                        {{ soi.product?.name }}
                      </td>
                      <td class="py-2 text-right pr-2">
                        <span
                          class="font-black text-slate-900 bg-white px-2 py-0.5 rounded-lg border border-slate-200 text-xs"
                        >
                          {{ soi.quantity }} {{ soi.product?.unit || 'pcs' }}
                        </span>
                      </td>
                    </tr>
                    <tr
                      v-for="b in di.sales_order?.bonus_items"
                      :key="b.id"
                      class="border-b border-slate-100 last:border-0 hover:bg-emerald-50/50 transition-colors"
                    >
                      <td class="py-2 pl-2">
                        <input
                          type="checkbox"
                          v-model="checkedItems[b.id]"
                          class="rounded border-slate-300 text-emerald-600 focus:ring-emerald-600 w-4 h-4 cursor-pointer"
                        />
                      </td>
                      <td class="py-2 font-mono text-[10px] text-emerald-500 font-bold">
                        BONUS
                      </td>
                      <td class="py-2 text-xs font-semibold text-emerald-700">
                        {{ b.bonus_inventory?.name || 'Item Bonus' }}
                        <p class="text-[9px] text-emerald-600/70" v-if="b.notes">{{ b.notes }}</p>
                      </td>
                      <td class="py-2 text-right pr-2">
                        <span
                          class="font-black text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-lg border border-emerald-100 text-xs"
                        >
                          {{ b.quantity }} {{ b.bonus_inventory?.unit || 'pcs' }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <!-- TAB 2: Surat Lembar DO (Bukti Muat) -->
          <div v-else-if="activeModalTab === 'do'" class="space-y-4">
            <div v-if="loadingDO" class="flex flex-col items-center py-10 text-slate-400 gap-2">
              <RefreshCw class="w-6 h-6 animate-spin text-slate-400" />
              <p class="text-xs font-bold">Memuat data lembar DO...</p>
            </div>

            <div v-else class="space-y-4">
              <!-- DO info banner -->
              <div class="bg-amber-50 border border-amber-200 rounded-2xl p-4 flex flex-col md:flex-row md:items-center justify-between gap-3 shadow-xs">
                <div>
                  <p class="text-[10px] font-black text-amber-700 uppercase tracking-widest leading-none">Nomor DO Terbit (Sistem)</p>
                  <p class="text-base font-black text-slate-900 font-mono mt-1">{{ warehouseDO?.do_number || 'Sedang memuat...' }}</p>
                  <p class="text-[10px] text-slate-500 font-medium mt-0.5">Sinkronisasi aktivitas muatan fisik gudang dengan persediaan.</p>
                </div>
                <div class="flex items-center gap-2">
                  <router-link
                    v-if="selectedBatch"
                    :to="`/warehouse/do-management?batch_id=${selectedBatch.id}`"
                    target="_blank"
                    class="px-3.5 py-2 bg-slate-900 hover:bg-slate-800 text-white text-xs font-black rounded-xl transition-all shadow-md flex items-center justify-center gap-1.5 whitespace-nowrap"
                    title="Buka form A4 dan cetak lembar DO di tab baru"
                  >
                    <span>🖨️ Cetak Lembar DO</span>
                  </router-link>
                  <button
                    @click="saveWarehouseDO"
                    :disabled="savingDO"
                    class="px-4 py-2 bg-amber-500 hover:bg-amber-600 disabled:opacity-50 text-white text-xs font-black rounded-xl transition-all shadow-md shadow-amber-250 flex items-center justify-center gap-1.5 whitespace-nowrap"
                  >
                    <RefreshCw v-if="savingDO" class="w-3.5 h-3.5 animate-spin" />
                    <span v-else>💾 Simpan DO</span>
                  </button>
                </div>
              </div>

              <!-- Product list table to adjust loaded carton/pcs -->
              <div class="bg-white border border-slate-100 rounded-2xl p-4 shadow-sm space-y-3">
                <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5 border-b border-slate-100 pb-2">
                  <Package class="w-4 h-4 text-slate-500" />
                  Rincian Aktual Barang Dimuat
                </h4>
                
                <div class="overflow-x-auto">
                  <table class="w-full text-left border-collapse">
                    <thead>
                      <tr class="border-b border-slate-200 text-[10px] uppercase tracking-widest text-slate-400">
                        <th class="py-2">SKU</th>
                        <th class="py-2">Produk</th>
                        <th class="py-2 text-center w-28">Karton Dimuat</th>
                        <th class="py-2 text-center w-24">Pcs Dimuat</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="item in doItems" :key="item.product_id || item.bonus_inventory_id" class="border-b border-slate-100 last:border-0 hover:bg-slate-50/50 transition-colors">
                        <td class="py-3 font-mono text-[10px]">
                          <span v-if="item.product_id" class="text-slate-500">{{ item.product?.sku }}</span>
                          <span v-else class="px-1.5 py-0.5 bg-emerald-100 text-emerald-800 rounded font-black text-[9px]">BONUS</span>
                        </td>
                        <td class="py-3 text-xs font-bold text-slate-800 leading-tight">
                          <template v-if="item.product_id">
                            {{ item.product?.name }}
                            <p class="text-[9px] text-slate-400 font-medium mt-0.5">Konversi: 1 Karton = {{ item.product?.pcs_per_unit || 1 }} Pcs</p>
                          </template>
                          <template v-else>
                            <span class="text-emerald-700 font-black">{{ item.bonus_inventory?.name || 'Item Bonus' }}</span>
                            <span class="ml-1 px-1.5 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 rounded text-[9px] font-bold">Barang Bonus</span>
                            <p class="text-[9px] text-emerald-600/70 font-medium mt-0.5">Satuan: {{ item.bonus_inventory?.unit || 'pcs' }}</p>
                          </template>
                        </td>
                        <td class="py-3 text-center">
                          <input
                            v-if="item.product_id"
                            v-model.number="item.cartons_loaded"
                            type="number"
                            min="0"
                            class="w-20 px-2.5 py-1.5 text-center border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-amber-500 focus:border-amber-500 text-xs font-bold"
                          />
                          <span v-else class="text-xs text-slate-300 font-bold">-</span>
                        </td>
                        <td class="py-3 text-center">
                          <input
                            v-model.number="item.pcs_loaded"
                            type="number"
                            min="0"
                            class="w-16 px-2 py-1.5 text-center border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-amber-500 focus:border-amber-500 text-xs font-bold"
                          />
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- DO Notes -->
              <div class="space-y-1">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Catatan Tambahan DO / Muatan</label>
                <textarea
                  v-model="doNotes"
                  placeholder="Ketik catatan khusus muat, misalnya: 'Supir memverifikasi muatan lengkap. 1 karton segel plastik tambahan.'"
                  rows="2"
                  class="w-full p-3 border border-slate-200 rounded-2xl text-xs focus:outline-none focus:ring-1 focus:ring-primary"
                ></textarea>
              </div>

              <!-- DO Info Roles accordion/cards -->
              <div class="bg-gradient-to-br from-slate-50 to-slate-100 rounded-2xl p-4 border border-slate-200/60 space-y-3">
                <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5">
                  <Info class="w-4 h-4 text-primary" />
                  Alur &amp; Peran Krusial Lembar DO dalam Distribusi
                </h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-[11px]">
                  <!-- Peran 1 -->
                  <div class="bg-white p-3 rounded-xl border border-slate-100/80 flex gap-2.5">
                    <div class="shrink-0 w-6 h-6 rounded-lg bg-red-50 flex items-center justify-center text-primary">
                      <Package class="w-3.5 h-3.5" />
                    </div>
                    <div>
                      <p class="font-bold text-slate-800">1. Warehouse Release (Potong Stok)</p>
                      <p class="text-slate-500 text-[10px] mt-0.5">Menjadi trigger sistem untuk memotong stok inventori gudang secara sinkron saat barang dimuat.</p>
                    </div>
                  </div>
                  <!-- Peran 2 -->
                  <div class="bg-white p-3 rounded-xl border border-slate-100/80 flex gap-2.5">
                    <div class="shrink-0 w-6 h-6 rounded-lg bg-blue-50 flex items-center justify-center text-blue-600">
                      <Truck class="w-3.5 h-3.5" />
                    </div>
                    <div>
                      <p class="font-bold text-slate-800">2. Dokumen Perjalanan (Transit Doc)</p>
                      <p class="text-slate-500 text-[10px] mt-0.5">Surat legal supir/kurir di jalan raya untuk membuktikan keabsahan barang bawaan.</p>
                    </div>
                  </div>
                  <!-- Peran 3 -->
                  <div class="bg-white p-3 rounded-xl border border-slate-100/80 flex gap-2.5">
                    <div class="shrink-0 w-6 h-6 rounded-lg bg-green-50 flex items-center justify-center text-green-600">
                      <ShieldCheck class="w-3.5 h-3.5" />
                    </div>
                    <div>
                      <p class="font-bold text-slate-800">3. Bukti Serah Terima Valid (POD)</p>
                      <p class="text-slate-500 text-[10px] mt-0.5">Ditandatangani &amp; dicap penerima sebagai bukti sah tanggung jawab barang berpindah ke pelanggan.</p>
                    </div>
                  </div>
                  <!-- Peran 4 -->
                  <div class="bg-white p-3 rounded-xl border border-slate-100/80 flex gap-2.5">
                    <div class="shrink-0 w-6 h-6 rounded-lg bg-amber-50 flex items-center justify-center text-amber-600">
                      <Receipt class="w-3.5 h-3.5" />
                    </div>
                    <div>
                      <p class="font-bold text-slate-800">4. Dasar Invoicing (Tagihan)</p>
                      <p class="text-slate-500 text-[10px] mt-0.5">Finance hanya akan menerbitkan invoice/faktur setelah lembar DO kembali lengkap bertanda tangan.</p>
                    </div>
                  </div>
                  <!-- Peran 5 -->
                  <div class="bg-white p-3 rounded-xl border border-slate-100/80 flex gap-2.5 md:col-span-2">
                    <div class="shrink-0 w-6 h-6 rounded-lg bg-purple-50 flex items-center justify-center text-purple-650">
                      <RotateCcw class="w-3.5 h-3.5" />
                    </div>
                    <div>
                      <p class="font-bold text-slate-800">5. Rekonsiliasi Retur &amp; Koreksi Lapangan</p>
                      <p class="text-slate-500 text-[10px] mt-0.5">Setiap barang rusak atau ditolak di lokasi dicatat langsung pada lembar DO untuk disinkronkan kembali saat supir kembali ke gudang.</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- TAB 3: Barang Cadangan (Spares) -->
          <div v-else-if="activeModalTab === 'spares'" class="space-y-4">
            <div class="bg-indigo-50 border border-indigo-100 rounded-2xl p-4 flex flex-col md:flex-row md:items-center justify-between gap-3 shadow-xs">
              <div>
                <p class="text-[10px] font-black text-indigo-700 uppercase tracking-widest leading-none">Manajemen Barang Cadangan (Spares)</p>
                <p class="text-xs text-slate-600 font-medium mt-1.5">
                  Tambahkan barang cadangan agar tim pengiriman memiliki fleksibilitas untuk menukar produk atau mengatasi barang rusak di lapangan.
                </p>
              </div>
              <button
                @click="saveSpares"
                :disabled="savingSpares"
                class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-xs font-black rounded-xl transition-all shadow-md flex items-center justify-center gap-1.5 whitespace-nowrap"
              >
                <RefreshCw v-if="savingSpares" class="w-3.5 h-3.5 animate-spin" />
                <span v-else>💾 Simpan Cadangan</span>
              </button>
            </div>

            <!-- Form Add Spare -->
            <div class="bg-white border border-slate-200 rounded-2xl p-4 shadow-xs space-y-3">
              <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5 pb-2 border-b border-slate-100">
                <BoxSelect class="w-4 h-4 text-indigo-600" />
                Tambah Barang Cadangan
              </h4>
              <div class="flex flex-col sm:flex-row gap-3 items-end">
                <div class="flex-1 space-y-1 w-full">
                  <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Pilih Produk</label>
                  <select
                    v-model="newSpareProductId"
                    class="w-full p-2.5 border border-slate-200 rounded-xl text-xs font-bold focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white"
                  >
                    <option value="" disabled>-- Pilih Produk --</option>
                    <option v-for="p in products" :key="p.id" :value="p.id">
                      {{ p.name }} ({{ p.sku }}) - Stok: {{ p.warehouse_stock ?? 0 }} pcs
                    </option>
                  </select>
                </div>
                <div class="w-24 space-y-1">
                  <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Qty (Pcs)</label>
                  <input
                    v-model.number="newSpareQty"
                    type="number"
                    min="1"
                    class="w-full p-2.5 border border-slate-200 rounded-xl text-xs font-bold focus:outline-none focus:ring-1 focus:ring-indigo-500 text-center"
                  />
                </div>
                <button
                  @click="addSpare"
                  type="button"
                  class="px-4 py-2.5 bg-slate-900 hover:bg-slate-800 text-white text-xs font-black rounded-xl transition-all shadow-xs h-[38px] flex items-center gap-1 w-full sm:w-auto justify-center"
                >
                  Tambah
                </button>
              </div>
            </div>

            <!-- List Spares -->
            <div class="bg-white border border-slate-200 rounded-2xl p-4 shadow-xs space-y-3">
              <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5 pb-2 border-b border-slate-100">
                <Package class="w-4 h-4 text-slate-500" />
                Daftar Muatan Cadangan Saat Ini
              </h4>

              <div v-if="localSpares.length === 0" class="text-center py-8 text-slate-400 text-xs font-medium">
                Belum ada barang cadangan yang ditambahkan untuk pengiriman ini.
              </div>

              <div v-else class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                  <thead>
                    <tr class="border-b border-slate-200 text-[10px] uppercase tracking-widest text-slate-400">
                      <th class="py-2">SKU</th>
                      <th class="py-2">Produk</th>
                      <th class="py-2 text-center w-24">Jumlah (Pcs)</th>
                      <th class="py-2 text-center w-20">Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(spare, idx) in localSpares" :key="spare.product_id" class="border-b border-slate-100 last:border-0 hover:bg-slate-50/50 transition-colors">
                      <td class="py-3 font-mono text-[10px] text-slate-500">
                        {{ products.find(p => p.id === spare.product_id)?.sku || '-' }}
                      </td>
                      <td class="py-3 text-xs font-bold text-slate-800 leading-tight">
                        {{ products.find(p => p.id === spare.product_id)?.name || 'Produk tidak ditemukan' }}
                      </td>
                      <td class="py-3 text-center">
                        <input
                          v-model.number="spare.quantity"
                          type="number"
                          min="1"
                          class="w-16 px-2 py-1 text-center border border-slate-200 rounded-lg text-xs font-bold focus:outline-none focus:ring-1 focus:ring-indigo-500"
                        />
                      </td>
                      <td class="py-3 text-center">
                        <button
                          @click="removeSpare(idx)"
                          class="p-1.5 bg-red-50 hover:bg-red-100 text-red-600 rounded-lg transition-colors font-bold text-xs"
                          title="Hapus"
                        >
                          Hapus
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <!-- TAB 4: Leftover / Susulan -->
          <div v-else-if="activeModalTab === 'leftovers'" class="space-y-4">
            <div class="bg-indigo-50 border border-indigo-100 rounded-2xl p-4 shadow-xs">
              <p class="text-[10px] font-black text-indigo-700 uppercase tracking-widest leading-none">Leftover &amp; Pengiriman Tertunda (Cabang Ini)</p>
              <p class="text-xs text-slate-600 font-medium mt-1.5">
                Berikut adalah daftar barang retur / klaim garansi yang disetujui untuk dikirim susulan. Anda dapat menyelipkan barang-barang ini ke dalam batch pengiriman supir saat ini jika muatan masih mencukupi.
              </p>
            </div>

            <div class="bg-white border border-slate-200 rounded-2xl p-4 shadow-xs space-y-3">
              <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5 pb-2 border-b border-slate-100">
                <Clock class="w-4 h-4 text-slate-500" />
                Daftar Barang Susulan Tersedia
              </h4>

              <div v-if="pendingLeftovers.length === 0" class="text-center py-8 text-slate-400 text-xs font-medium">
                Tidak ada barang susulan / leftover yang pending untuk cabang ini.
              </div>

              <div v-else class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                  <thead>
                    <tr class="border-b border-slate-200 text-[10px] uppercase tracking-widest text-slate-400">
                      <th class="py-2">Toko Asal</th>
                      <th class="py-2">Produk</th>
                      <th class="py-2 text-center w-24">Jumlah</th>
                      <th class="py-2 text-center w-24">Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in pendingLeftovers" :key="item.id" class="border-b border-slate-100 last:border-0 hover:bg-slate-50/50 transition-colors">
                      <td class="py-3 text-xs">
                        <div class="font-bold text-slate-800">
                          {{ item.sales_order?.store?.name || item.product_exchange?.store?.name || 'Toko' }}
                        </div>
                        <div class="text-[10px] text-slate-400">
                          {{ item.sales_order?.so_number || item.product_exchange?.exchange_no || '-' }}
                        </div>
                      </td>
                      <td class="py-3 text-xs">
                        <div class="font-bold text-indigo-700 leading-tight">
                          {{ item.product?.name || 'Produk' }}
                        </div>
                        <div class="text-[10px] font-mono text-slate-400">
                          SKU: {{ item.product?.sku || '-' }}
                        </div>
                      </td>
                      <td class="py-3 text-center text-xs font-black text-slate-700">
                        {{ item.quantity }} Pcs
                      </td>
                      <td class="py-3 text-center">
                        <button
                          @click="slipLeftover(item.id)"
                          :disabled="slippingLeftoverId === item.id"
                          class="px-2.5 py-1.5 bg-indigo-50 hover:bg-indigo-100 disabled:opacity-50 text-indigo-600 rounded-lg transition-colors font-black text-[10px]"
                        >
                          <span v-if="slippingLeftoverId === item.id">Menyelipkan...</span>
                          <span v-else>🚚 Selipkan</span>
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-5 border-t border-slate-100 bg-slate-50/50 flex gap-3">
          <button
            @click="showDetailModal = false"
            class="flex-1 px-5 py-3 rounded-2xl border border-slate-200 font-bold text-sm text-slate-600 hover:bg-slate-50 transition-colors"
          >
            Tutup
          </button>
          <button
            v-if="selectedBatch.status === 'SUPERVISOR_APPROVED' || selectedBatch.status === 'PENDING'"
            @click="confirmWarehouseReady(selectedBatch.id)"
            :disabled="confirmingReady || !allItemsChecked"
            class="flex-[2] bg-amber-500 hover:bg-amber-600 text-white rounded-2xl py-3 font-black text-sm flex items-center justify-center gap-2 transition-all shadow-lg shadow-amber-200/50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <CheckCircle2 class="w-4 h-4" />
            {{
              confirmingReady
                ? 'Memproses...'
                : allItemsChecked
                  ? '✅ Konfirmasi Barang Siap Diambil Driver'
                  : 'Centang Semua Barang Untuk Konfirmasi'
            }}
          </button>
          <div
            v-else
            class="flex-[2] bg-red-50 border border-red-200 rounded-2xl py-3 flex items-center justify-center gap-2"
          >
            <BoxSelect class="w-4 h-4 text-primary" />
            <span class="text-primary font-black text-sm"
              >Sudah dikonfirmasi — Driver sedang picking</span
            >
          </div>
        </div>
      </div>
    </div>

    <!-- ─── MT DO MODAL ─── -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showMtDetailModal && mtWorkingOrder" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showMtDetailModal = false" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <div class="sticky top-0 bg-white rounded-t-[2rem] px-8 pt-8 pb-6 border-b border-slate-100 z-10 flex justify-between items-center">
              <div>
                <h2 class="text-xl font-black text-slate-900">{{ mtWorkingOrder.do_number }}</h2>
                <p class="text-sm text-slate-500 mt-1">{{ mtWorkingOrder.key_account?.name }}</p>
              </div>
              <button @click="showMtDetailModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                <X class="h-5 w-5" />
              </button>
            </div>
            
            <div class="px-8 py-6 space-y-5">
              <!-- Dokumen -->
              <div class="grid grid-cols-2 gap-4">
                <div v-if="mtWorkingOrder.sales_order?.delivery_note_url || mtWorkingOrder.delivery_note_url">
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Surat Jalan (Lampiran)</p>
                  <a
                    :href="mtWorkingOrder.sales_order?.delivery_note_url || mtWorkingOrder.delivery_note_url"
                    target="_blank"
                    class="text-xs font-bold text-violet-700 hover:underline flex items-center gap-1 bg-violet-50 px-3 py-2 rounded-xl inline-block"
                  >
                    📥 Buka Dokumen DO
                  </a>
                </div>
              </div>

              <!-- Approve form -->
              <div v-if="mtAction === 'APPROVE'" class="space-y-4">
                <div class="bg-emerald-50 border border-emerald-200 rounded-2xl p-4">
                  <p class="text-sm text-emerald-700 font-medium">Setujui DO ini untuk mulai menyiapkan barang.</p>
                </div>
                <div class="space-y-1.5">
                  <label class="text-[10px] font-black text-slate-400 uppercase">Disetujui Oleh</label>
                  <input v-model="mtApproveBy" type="text" disabled class="w-full bg-slate-100 border-2 border-transparent rounded-xl px-4 py-3 text-sm font-bold text-slate-500 cursor-not-allowed outline-none" />
                </div>
              </div>

              <!-- Loading form -->
              <div v-if="mtAction === 'LOADING'" class="space-y-4">
                 <div class="space-y-3">
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400">Qty Dimuat</p>
                  <div v-for="(item, idx) in mtLoadingForm.items" :key="idx" class="bg-slate-50 rounded-2xl p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
                    <div>
                      <p class="font-bold text-sm text-slate-900">{{ mtWorkingOrder.items?.[idx]?.product?.name }}</p>
                      <p class="text-xs text-slate-400">Diminta: {{ mtWorkingOrder.items?.[idx]?.requested_qty }}</p>
                    </div>
                    <div class="flex items-center gap-2">
                      <label class="text-[10px] font-black text-slate-400 uppercase">Dimuat:</label>
                      <input v-model.number="item.shipped_qty" type="number" min="0" class="w-20 bg-white border-2 border-slate-200 rounded-xl px-3 py-2 text-sm font-bold text-center outline-none focus:border-indigo-400" />
                    </div>
                  </div>
                </div>
              </div>

              <!-- Ship form -->
              <div v-if="mtAction === 'SHIP'" class="space-y-4">
                <div class="bg-indigo-50 border-l-4 border-indigo-500 p-4 rounded-r-xl mb-4">
                  <p class="text-xs font-bold text-indigo-800 uppercase tracking-wider mb-1">Sales MT (Penagihan)</p>
                  <p class="text-sm font-black text-indigo-950">{{ mtWorkingOrder.sales_order?.key_account?.name || mtWorkingOrder.key_account?.name || 'Sales MT (Tidak ada data)' }}</p>
                </div>
                <div class="space-y-1.5">
                  <label class="text-[10px] font-black text-slate-400 uppercase">Kendaraan (Plat Nomor)</label>
                  <select v-model="mtShipForm.vehicle_id" class="w-full bg-slate-50 border-2 border-transparent rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-purple-300">
                    <option value="">Pilih kendaraan...</option>
                    <option v-for="v in mtVehicles" :key="v.id" :value="v.id">{{ v.license_plate }}</option>
                  </select>
                </div>
                <div class="space-y-1.5">
                  <label class="text-[10px] font-black text-slate-400 uppercase">Supir Internal (Dari Sistem)</label>
                  <select v-model="mtShipForm.driver_id" class="w-full bg-slate-50 border-2 border-transparent rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-purple-300">
                    <option value="">-- Bukan supir internal --</option>
                    <option v-for="d in mtDrivers" :key="d.id" :value="d.id">{{ d.first_name }} {{ d.last_name }}</option>
                  </select>
                </div>
                <div class="space-y-1.5" v-if="!mtShipForm.driver_id">
                  <label class="text-[10px] font-black text-slate-400 uppercase">Supir Eksternal (Manual)</label>
                  <input v-model="mtShipForm.driver_name" type="text" placeholder="Ketik nama supir luar..." class="w-full bg-slate-50 border-2 border-transparent rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-purple-300" />
                </div>
                <div class="space-y-1.5">
                  <label class="text-[10px] font-black text-slate-400 uppercase">Catatan / Keterangan</label>
                  <input v-model="mtShipForm.notes" type="text" placeholder="Catatan opsional..." class="w-full bg-slate-50 border-2 border-transparent rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-purple-300" />
                </div>
              </div>

              <!-- Info Detail only -->
              <div v-if="mtAction === 'DETAIL'" class="space-y-4">
                <div v-for="item in mtWorkingOrder.items" :key="item.id" class="bg-slate-50 rounded-2xl p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
                  <div class="flex-1 min-w-0">
                    <p class="font-bold text-sm text-slate-900 truncate">{{ item.product?.name || 'Produk' }}</p>
                  </div>
                  <div class="text-right">
                    <p class="text-xs text-slate-400">Diminta / Dimuat</p>
                    <p class="font-black text-slate-900">{{ item.requested_qty }} / {{ item.shipped_qty || item.prepared_qty || '—' }}</p>
                  </div>
                </div>
              </div>

            </div>
            
            <div v-if="mtAction !== 'DETAIL'" class="sticky bottom-0 bg-white rounded-b-[2rem] px-8 py-6 border-t border-slate-100 flex gap-3">
              <button @click="showMtDetailModal = false" class="flex-1 py-3 border-2 border-slate-200 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-50 transition-all">Batal</button>
              <button
                @click="handleMtSubmit"
                :disabled="mtIsSubmitting"
                class="flex-1 py-3 bg-gradient-to-r from-indigo-500 to-blue-600 text-white rounded-2xl text-sm font-bold shadow-lg disabled:opacity-50 hover:shadow-xl transition-all"
              >
                <span v-if="mtIsSubmitting">...</span>
                <span v-else>{{ mtAction === 'APPROVE' ? 'Approve DO' : mtAction === 'LOADING' ? 'Muat Barang' : 'Kirim Sekarang' }}</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Modal Proses Delegasi (Pemilihan Sopir, Kendaraan, & Alokasi Batch) -->
    <Teleport to="body">
      <div v-if="showProcessModal" class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[200] flex items-center justify-center p-4">
        <div class="bg-white w-full max-w-2xl rounded-[2.5rem] shadow-2xl overflow-hidden animate-in zoom-in-95 duration-300 max-h-[90vh] flex flex-col">
          <div class="p-6 md:p-8 border-b border-slate-100 flex items-start justify-between shrink-0 bg-white">
            <div>
              <div class="flex items-center gap-2">
                <span class="px-2.5 py-1 rounded-lg bg-orange-100 text-orange-700 text-[10px] font-black tracking-widest uppercase">Delegasi Pabrik</span>
                <span class="font-mono text-xs font-black text-slate-500">{{ processForm.doNo }}</span>
              </div>
              <h3 class="text-xl md:text-2xl font-black text-slate-900 mt-1">Alokasi Muatan & Batch Gudang</h3>
              <p class="text-xs text-slate-500 font-medium mt-0.5">Tentukan armada, sopir, dan batch fisik yang dimuat dari gudang cabang ini.</p>
            </div>
            <button @click="showProcessModal = false" class="p-2 hover:bg-slate-100 rounded-xl transition-colors">
              <X class="w-5 h-5 text-slate-400" />
            </button>
          </div>
          
          <div class="p-6 md:p-8 space-y-6 bg-slate-50/50 overflow-y-auto custom-scrollbar flex-1">
            <!-- Armada & Sopir -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-xs font-bold text-slate-700 ml-1">Armada (Kendaraan) <span class="text-rose-500">*</span></label>
                <select
                  v-model="processForm.vehicle_id"
                  class="w-full px-4 py-3 bg-white border border-slate-200 rounded-2xl text-xs font-bold outline-none focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20"
                >
                  <option value="">Pilih Armada...</option>
                  <option
                    v-for="v in availableVehicles"
                    :key="v.id"
                    :value="v.id"
                  >
                    {{ v.name }} ({{ v.license_plate }}) - Cap: {{ v.capacity }} KG
                  </option>
                </select>
              </div>

              <div class="space-y-2">
                <label class="text-xs font-bold text-slate-700 ml-1">Sopir (Driver) <span class="text-rose-500">*</span></label>
                <select
                  v-model="processForm.driver_id"
                  class="w-full px-4 py-3 bg-white border border-slate-200 rounded-2xl text-xs font-bold outline-none focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20"
                >
                  <option value="">Pilih Sopir...</option>
                  <option
                    v-for="e in availableDrivers"
                    :key="e.id"
                    :value="e.id"
                  >
                    {{ e.first_name }} {{ e.last_name }}
                  </option>
                </select>
              </div>
            </div>

            <!-- Daftar Item & Pilihan Batch Fisik Gudang -->
            <div class="space-y-3 pt-4 border-t border-slate-200/60">
              <div class="flex items-center justify-between">
                <label class="text-xs font-black text-slate-800 uppercase tracking-wider flex items-center gap-1.5">
                  <Package class="w-4 h-4 text-orange-500" />
                  Alokasi Batch Fisik Barang
                </label>
                <span class="text-[10px] font-bold text-slate-400">Pilih batch dari stok gudang Anda</span>
              </div>

              <div class="space-y-3">
                <div
                  v-for="(item, idx) in processForm.items"
                  :key="item.id || idx"
                  class="p-4 bg-white border border-slate-200/80 rounded-2xl shadow-xs space-y-3"
                >
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-xs font-black text-slate-900">{{ item.product_name }}</p>
                      <p class="text-[10px] text-slate-400 font-bold mt-0.5">
                        Permintaan: <span class="text-slate-800 font-black">{{ item.quantity }} {{ item.unit }}</span>
                        <span v-if="(item.pcs_per_unit || 1) > 1" class="text-slate-400"> ({{ item.quantity * item.pcs_per_unit }} PCS)</span>
                      </p>
                    </div>
                  </div>

                  <div class="space-y-1">
                    <label class="text-[10px] font-bold text-slate-500 uppercase">Batch Fisik Gudang <span class="text-rose-500">*</span></label>
                    <select
                      v-model="item.batch_no"
                      class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none focus:border-orange-500"
                    >
                      <option value="" disabled>Pilih Batch Gudang...</option>
                      <option
                        v-for="b in getBranchBatchesForProduct(item.product_id)"
                        :key="b.batch_no"
                        :value="b.batch_no"
                      >
                        Batch {{ b.batch_no }} (Tersedia: {{ b.quantity }} PCS)
                      </option>
                    </select>
                    <p v-if="getBranchBatchesForProduct(item.product_id).length === 0" class="text-[10px] font-bold text-rose-500 flex items-center gap-1 mt-1">
                      <AlertCircle class="w-3 h-3" /> Tidak ada batch stok GOOD tersedia di gudang cabang ini!
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="p-6 md:p-8 bg-white border-t border-slate-100 flex justify-end gap-3 shrink-0">
            <button class="px-6 py-3 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold rounded-xl transition-colors text-xs" @click="showProcessModal = false">Batal</button>
            <button class="px-8 py-3 bg-orange-500 hover:bg-orange-600 text-white font-black rounded-xl shadow-lg shadow-orange-500/30 transition-all text-xs flex items-center gap-2" @click="submitProcessDelegation">
              <Truck class="w-4 h-4" /> Alokasikan & Siap Muat
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

/* Make table rows more compact to minimize scrolling */
:deep(.py-4) {
  padding-top: 0.5rem !important;
  padding-bottom: 0.5rem !important;
}

:deep(th.py-4) {
  padding-top: 0.75rem !important;
  padding-bottom: 0.75rem !important;
}

:deep(tr) {
  height: auto !important;
}
</style>
