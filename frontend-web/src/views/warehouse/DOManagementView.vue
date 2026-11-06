<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useMasterDataStore } from '@/stores/masterData'
import { useWarehouseStore } from '@/stores/warehouse'
import { useVehicleStore } from '@/stores/vehicle'
import apiClient from '@/api/axios'
import { warehouseApi } from '@/api/warehouse'
import { filterSellableProducts } from '@/utils/productUnits'
import { toast } from 'vue-sonner'
import {
  Package,
  Truck,
  Search,
  RefreshCw,
  BoxSelect,
  ShieldCheck,
  Clock,
  CheckCircle2,
  AlertTriangle,
  Printer,
  Plus,
  Send,
  MapPin,
  Calendar,
  X,
  FileEdit,
} from 'lucide-vue-next'
import DataTable from '@/components/DataTable.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

const route = useRoute()

function printDO() {
  window.print();
}

const masterStore = useMasterDataStore()
const warehouseStore = useWarehouseStore()
const vehicleStore = useVehicleStore()

// ─── State ───────────────────────────────────────────────────────────────────
const batches = ref<any[]>([])
const motorisDOs = ref<any[]>([])
const loadingBatches = ref(false)
const selectedBatchId = ref<string>('')
const selectedBatch = computed(() => {
  return batches.value.find(b => b.id === selectedBatchId.value) || 
         motorisDOs.value.find(m => m.id === selectedBatchId.value) || 
         null
})

const activeTab = ref<'do' | 'spares' | 'leftovers' | 'transfers'>('do')

// DO states
const warehouseDO = ref<any | null>(null)
const loadingDO = ref(false)
const savingDO = ref(false)
const doItems = ref<any[]>([])
const doNotes = ref('')
const isFinalConfirmed = ref(false)
const isDispatched = ref(false)

// Spares (Barang Cadangan) states
const products = ref<any[]>([])
const localSpares = ref<Array<{ product_id: string; quantity: number }>>([])
const savingSpares = ref(false)
const newSpareProductId = ref('')
const newSpareQty = ref(1)

// Reactive spares representing the actual load in DO
const localSparesForDO = ref<any[]>([])

// Leftovers (Barang Susulan) states
const pendingLeftovers = ref<any[]>([])
const slippingLeftoverId = ref<string | null>(null)

// Manual Leftovers states
const showManualLeftoverModal = ref(false)
const manualLeftoverForm = ref({
  store_id: '',
  salesman_id: '',
  product_id: '',
  quantity: 1,
  notes: ''
})
const creatingManualLeftover = ref(false)

async function submitManualLeftover() {
  if (!manualLeftoverForm.value.store_id || !manualLeftoverForm.value.salesman_id || !manualLeftoverForm.value.product_id || manualLeftoverForm.value.quantity < 1) {
    toast.error('Mohon lengkapi semua data wajib.')
    return
  }

  creatingManualLeftover.value = true
  try {
    await apiClient.post('/delivery/subsequent', manualLeftoverForm.value)
    toast.success('Barang susulan berhasil ditambahkan.')
    showManualLeftoverModal.value = false
    manualLeftoverForm.value = {
      store_id: '',
      salesman_id: '',
      product_id: '',
      quantity: 1,
      notes: ''
    }
    fetchPendingLeftovers()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menambahkan barang susulan.')
  } finally {
    creatingManualLeftover.value = false
  }
}

// Sales Transfer (Mutasi Replenishment) states
const salesTransfers = ref<any[]>([])
const loadingTransfers = ref(false)
const newTransferEmployeeId = ref('')
const newTransferProductId = ref('')
const newTransferQty = ref(1)
const newTransferNotes = ref('')
const creatingTransfer = ref(false)
const slippingTransferId = ref<string | null>(null)

// Dynamic Modal States
const showDynamicInputModal = ref(false)
const dynamicModalTitle = ref('')
const dynamicModalItem = ref<any>(null)
const dynamicModalField = ref<'berangkat' | 'pulang' | 'terjual' | 'sample' | 'retur'>('berangkat')
const dynamicModalForm = ref({
  pcs_per_unit: 1,
  cartons: 0,
  pcs: 0
})

const dynamicTotalPcs = computed(() => {
  return (dynamicModalForm.value.cartons * dynamicModalForm.value.pcs_per_unit) + dynamicModalForm.value.pcs
})

function openDynamicModal(item: any, field: 'berangkat' | 'pulang' | 'terjual' | 'sample' | 'retur') {
  if (isFinalConfirmed.value) return // Cannot edit anything after finalization
  if (isDispatched.value && field === 'berangkat') {
    toast.error('DO sudah di-dispatch. Tidak dapat mengubah jumlah barang berangkat.')
    return
  }
  dynamicModalItem.value = item
  dynamicModalField.value = field
  
  const titles = {
    berangkat: 'Berangkat / Dimuat',
    pulang: 'Pulang',
    terjual: 'Terjual',
    sample: 'Sample',
    retur: 'Retur'
  }
  dynamicModalTitle.value = titles[field]
  
  const pcsPerUnit = item.custom_pcs_per_unit || 1
  dynamicModalForm.value.pcs_per_unit = pcsPerUnit
  
  if (field === 'berangkat') {
    dynamicModalForm.value.cartons = item.cartons_loaded || 0
    dynamicModalForm.value.pcs = item.pcs_loaded || 0
  } else {
    let totalPcs = 0
    if (field === 'pulang') totalPcs = item.pcs_returned_to_warehouse || 0
    if (field === 'terjual') totalPcs = item.pcs_sold || 0
    if (field === 'sample') totalPcs = item.pcs_sample || 0
    if (field === 'retur') totalPcs = item.pcs_returned_from_customer || 0
    
    dynamicModalForm.value.cartons = Math.floor(totalPcs / pcsPerUnit)
    dynamicModalForm.value.pcs = totalPcs % pcsPerUnit
  }
  
  showDynamicInputModal.value = true
}

function closeDynamicModal() {
  showDynamicInputModal.value = false
  dynamicModalItem.value = null
}

function saveDynamicModal() {
  if (!dynamicModalItem.value) return
  
  dynamicModalItem.value.custom_pcs_per_unit = dynamicModalForm.value.pcs_per_unit
  
  if (dynamicModalField.value === 'berangkat') {
    dynamicModalItem.value.cartons_loaded = dynamicModalForm.value.cartons
    dynamicModalItem.value.pcs_loaded = dynamicModalForm.value.pcs
  } else {
    const totalPcs = dynamicTotalPcs.value
    if (dynamicModalField.value === 'pulang') dynamicModalItem.value.pcs_returned_to_warehouse = totalPcs
    if (dynamicModalField.value === 'terjual') dynamicModalItem.value.pcs_sold = totalPcs
    if (dynamicModalField.value === 'sample') dynamicModalItem.value.pcs_sample = totalPcs
    if (dynamicModalField.value === 'retur') dynamicModalItem.value.pcs_returned_from_customer = totalPcs
  }
  
  closeDynamicModal()
}

const isSalesPosition = (title?: string, deptName?: string) => {
  const t = (title || '').toLowerCase()
  const d = (deptName || '').toLowerCase()
  const isExcluded = /(operator|driver|sopir|kuli|helper|gudang|warehouse|security|cleaning|harian|validator|editor|designer|desain|purchasing|auditor|kasir|cashier|teknisi|technician)/i.test(t)
  if (isExcluded) return false
  return /(sales|motoris|booster|marketing|take order|task order|\bto\b|\bmt\b|modern trade|key account|\bkam\b|kanvas)/i.test(t) ||
         /(sales|marketing|commercial)/i.test(d)
}

// Filter salesman employees (those living far/mini depot)
const salesEmployees = computed(() => {
  const branchId = masterStore.selectedWarehouseBranchId
  const list = masterStore.employees.filter((emp: any) => {
    const isActive = emp.user?.is_active !== false &&
      emp.employment_status !== 'RESIGNED' &&
      emp.employment_status !== 'TERMINATED' &&
      !String(emp.employment_status || '').toLowerCase().includes('harian')
    if (!isActive) return false

    if (branchId && branchId !== 'ALL_BRANCHES') {
      if (emp.branch_id && emp.branch_id !== branchId) return false
    }
    const jobTitle = emp.job_position?.title || ''
    const deptName = emp.department?.name || ''
    return isSalesPosition(jobTitle, deptName)
  })

  return list.sort((a: any, b: any) => {
    const nameA = `${a.first_name} ${a.last_name || ''}`.trim()
    const nameB = `${b.first_name} ${b.last_name || ''}`.trim()
    return nameA.localeCompare(nameB)
  })
})

// Search filter for SJ dropdown
const searchSJQuery = ref('')
const filteredBatchesForSelect = computed(() => {
  const q = searchSJQuery.value.trim().toLowerCase()
  let list = [...batches.value, ...motorisDOs.value]
  const wBranchId = masterStore.selectedWarehouseBranchId
  
  if (
    wBranchId &&
    wBranchId !== 'ALL_BRANCHES' &&
    wBranchId !== 'ALL'
  ) {
    list = list.filter(b => b.branch_id === wBranchId || b.is_motoris)
  }

  if (!q) return list
  return list.filter((b) => {
    const fullName = `${b.driver?.first_name || b.employee?.first_name || ''} ${b.driver?.last_name || b.employee?.last_name || ''}`.toLowerCase()
    return (
      (b.delivery_order_no || b.do_number)?.toLowerCase().includes(q) ||
      fullName.includes(q) ||
      b.vehicle?.license_plate?.toLowerCase().includes(q) ||
      (b.is_motoris && 'motoris'.includes(q))
    )
  })
})

// Calculate target SJ quantities by product from the selected batch
const batchProductQtyMap = computed(() => {
  if (!selectedBatch.value || !selectedBatch.value.items) return {}
  const map: Record<string, number> = {}
  
  selectedBatch.value.items.forEach((di: any) => {
    if (di.sales_order?.items) {
      di.sales_order.items.forEach((soi: any) => {
        if (soi.product_id) {
          map[soi.product_id] = (map[soi.product_id] || 0) + (soi.quantity || soi.ordered_quantity || 0)
        }
      })
    }
    // Include bonus items from sales order
    if (di.sales_order?.bonus_items) {
      di.sales_order.bonus_items.forEach((b: any) => {
        if (b.bonus_inventory_id) {
          map[b.bonus_inventory_id] = (map[b.bonus_inventory_id] || 0) + (b.quantity || 0)
        }
      })
    }
    // Include leftovers if linked to this batch item
    if (di.sales_subsequent_item) {
      const item = di.sales_subsequent_item
      if (item.product_id) {
        map[item.product_id] = (map[item.product_id] || 0) + (item.quantity || 0)
      }
    }
    // Include sales stock transfers linked to this batch item
    if (di.sales_transfer) {
      const item = di.sales_transfer
      if (item.product_id) {
        map[item.product_id] = (map[item.product_id] || 0) + (item.quantity || 0)
      } else if (item.bonus_inventory_id) {
        map[item.bonus_inventory_id] = (map[item.bonus_inventory_id] || 0) + (item.quantity || 0)
      }
    }
  })

  // Include Spares (Barang Cadangan) directly from batch
  if (selectedBatch.value.spares) {
    selectedBatch.value.spares.forEach((spare: any) => {
      if (spare.product_id) {
        map[spare.product_id] = (map[spare.product_id] || 0) + (spare.quantity || 0)
      }
    })
  }

  return map
})

// Calculate aggregated bonus items for the selected batch
const batchBonusMap = computed(() => {
  if (!selectedBatch.value || !selectedBatch.value.items) return []
  const map = new Map<string, any>()
  
  selectedBatch.value.items.forEach((di: any) => {
    if (di.sales_order?.bonus_items) {
      di.sales_order.bonus_items.forEach((b: any) => {
        if (b.bonus_inventory_id) {
          if (map.has(b.bonus_inventory_id)) {
            const existing = map.get(b.bonus_inventory_id)
            existing.quantity += (b.quantity || 0)
          } else {
            map.set(b.bonus_inventory_id, {
              id: b.bonus_inventory_id,
              name: b.bonus_inventory?.name || 'Item Bonus',
              unit: b.bonus_inventory?.unit || 'PCS',
              quantity: b.quantity || 0,
            })
          }
        }
      })
    }
  })
  
  return Array.from(map.values())
})

// Filter pending transfers not yet in any active delivery batch
const pendingSalesTransfers = computed(() => {
  const usedIds = new Set<string>()
  batches.value.forEach(b => {
    (b.items || []).forEach((item: any) => {
      if (item.sales_transfer_id) {
        usedIds.add(item.sales_transfer_id)
      }
    })
  })
  
  return salesTransfers.value.filter((t: any) => {
    return t.status === 'PENDING' && t.type === 'TRANSFER' && !usedIds.has(t.id)
  })
})

const currentEmployeeId = computed(() => {
  if (warehouseDO.value?.employee_id) return warehouseDO.value.employee_id
  if (selectedBatch.value?.driver_id) return selectedBatch.value.driver_id
  return null
})

const pendingReturns = computed(() => {
  if (!currentEmployeeId.value) return []
  return salesTransfers.value.filter((t: any) => t.type === 'RETURN' && t.status === 'PENDING' && t.employee_id === currentEmployeeId.value)
})

function syncReturnsFromMobile() {
  if (pendingReturns.value.length === 0) {
    toast.error('Tidak ada request retur/pengembalian ke gudang dari salesman ini.')
    return
  }
  
  let syncCount = 0
  pendingReturns.value.forEach((ret: any) => {
    const doItem = doItems.value.find((i: any) => 
      (i.product_id && i.product_id === ret.product_id) || 
      (i.bonus_inventory_id && i.bonus_inventory_id === ret.bonus_inventory_id)
    )
    if (doItem) {
      doItem.pcs_returned_to_warehouse = ret.quantity
      const totalBerangkat = (doItem.cartons_loaded * doItem.custom_pcs_per_unit) + doItem.pcs_loaded
      doItem.pcs_sold = Math.max(0, totalBerangkat - doItem.pcs_returned_to_warehouse - doItem.pcs_sample - doItem.pcs_returned_from_customer)
      syncCount++
    }
  })
  
  if (syncCount > 0) {
    toast.success(`${syncCount} produk berhasil disinkronisasi dari aplikasi mobile! Jangan lupa klik Simpan Sementara atau Finalisasi.`)
  } else {
    toast.info('Request pengembalian tidak cocok dengan produk yang ada di DO ini.')
  }
}

// Helper to format quantity into carton + remaining pcs representation
function formatCartonPcs(totalPcs: number, pcsPerUnit: number) {
  const units = pcsPerUnit || 1
  const cartons = Math.floor(totalPcs / units)
  const remaining = totalPcs % units
  if (cartons > 0 && remaining > 0) {
    return `${cartons} Dus, ${remaining} Pcs`
  } else if (cartons > 0) {
    return `${cartons} Dus`
  } else {
    return `${remaining} Pcs`
  }
}

// Populate localSparesForDO when selectedBatch changes or when localSpares is fetched
function syncSparesForDO() {
  if (!selectedBatch.value) {
    localSparesForDO.value = []
    return
  }
  const sparesList = selectedBatch.value.spares || []
  localSparesForDO.value = sparesList.map((s: any) => {
    const p = s.product || products.value.find(prod => prod.id === s.product_id) || {}
    const pcsPerUnit = p.pcs_per_unit || 1
    const totalPcs = s.quantity || 0
    return {
      product_id: s.product_id,
      product_name: p.name || 'Produk Cadangan',
      sku: p.sku || '-',
      pcs_per_unit: pcsPerUnit,
      cartons: Math.floor(totalPcs / pcsPerUnit),
      pcs: totalPcs % pcsPerUnit,
      quantity: totalPcs
    }
  })
}

function updateSpareTotal(spare: any) {
  spare.quantity = (Number(spare.cartons) || 0) * spare.pcs_per_unit + (Number(spare.pcs) || 0)
}

// Watch selected warehouse branch changes to reload batches
watch(() => masterStore.selectedWarehouseBranchId, () => {
  fetchBatches()
  fetchMotorisDOs()
  fetchSalesTransfers()
})

// Watch selected batch changes to reload DO/spares
watch(selectedBatchId, (newId) => {
  if (newId) {
    const isMotoris = motorisDOs.value.some(m => m.id === newId)
    if (isMotoris) {
      fetchDOById(newId)
      localSpares.value = []
      syncSparesForDO()
    } else {
      fetchOrCreateDO(newId)
      const batch = batches.value.find(b => b.id === newId)
      if (batch) {
        localSpares.value = (batch.spares || []).map((s: any) => ({
          product_id: s.product_id,
          quantity: s.quantity
        }))
        syncSparesForDO()
      }
    }
  } else {
    warehouseDO.value = null
    doItems.value = []
    doNotes.value = ''
    localSpares.value = []
    localSparesForDO.value = []
    isFinalConfirmed.value = false
    isDispatched.value = false
  }
})

// Watch localSpares/products loading to keep localSparesForDO in sync
watch([products, batches], () => {
  syncSparesForDO()
}, { deep: true })

// ─── Fetch Data ───────────────────────────────────────────────────────────────
async function fetchBatches() {
  if (!masterStore.selectedWarehouseBranchId) return
  loadingBatches.value = true
  try {
    const branchId = masterStore.selectedWarehouseBranchId
    const params = branchId && branchId !== 'ALL_BRANCHES' && branchId !== 'ALL' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/delivery/batches/warehouse${params}`)
    const normalBatches = res.data.data || []

    // Fetch Factory Delegations
    let delegationGroups: any[] = []
    try {
      await warehouseStore.fetchPendingShipments()
      const groups: Record<string, any> = {}
      warehouseStore.pendingShipments
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
                    item.status === 'SHIPPED' ? 'ON_DELIVERY' : 
                    item.status === 'ARRIVED' || item.status === 'RECEIVED' ? 'COMPLETED' : item.status,
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

    batches.value = [...normalBatches, ...delegationGroups].sort((a, b) => new Date(b.created_at || 0).getTime() - new Date(a.created_at || 0).getTime())
    
    if (batches.value.length > 0 && !selectedBatchId.value) {
      selectedBatchId.value = batches.value[0].id
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal memuat Surat Jalan')
  } finally {
    loadingBatches.value = false
  }
}

async function fetchMotorisDOs() {
  if (!masterStore.selectedWarehouseBranchId) return
  loadingBatches.value = true
  try {
    const branchId = masterStore.selectedWarehouseBranchId
    const params = branchId && branchId !== 'ALL_BRANCHES' && branchId !== 'ALL' ? `?branch_id=${branchId}` : ''
    const res = await warehouseApi.getMotorisDOs(branchId)
    motorisDOs.value = (res.data?.data || []).map((wdo: any) => ({
      ...wdo,
      is_motoris: true,
      do_number: wdo.do_number,
      status: wdo.status,
      employee: wdo.employee
    }))
    
    // Auto select if nothing is selected and we just loaded motoris DOs
    if (motorisDOs.value.length > 0 && !selectedBatchId.value) {
      selectedBatchId.value = motorisDOs.value[0].id
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal memuat DO Motoris')
  } finally {
    loadingBatches.value = false
  }
}

async function fetchProducts() {
  try {
    const branchId = masterStore.selectedBranchId
    const params = branchId && branchId !== 'ALL_BRANCHES' && branchId !== 'ALL' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/factory/products${params}`)
    const rawList = res.data?.data || res.data || []
    products.value = filterSellableProducts(rawList)
  } catch {
    toast.error('Gagal memuat produk')
  }
}

async function fetchPendingLeftovers() {
  try {
    const branchId = masterStore.selectedBranchId
    const params = branchId && branchId !== 'ALL_BRANCHES' && branchId !== 'ALL' ? `?branch_id=${branchId}` : ''
    const res = await apiClient.get(`/delivery/subsequent/pending${params}`)
    pendingLeftovers.value = res.data?.data || []
  } catch {
    toast.error('Gagal memuat daftar pengiriman tertunda/leftover')
  }
}

async function fetchSalesTransfers() {
  const branchId = masterStore.selectedBranchId
  if (!branchId) return
  loadingTransfers.value = true
  try {
    const res = await apiClient.get(`/sales-transfers/branch/${branchId}`)
    salesTransfers.value = res.data?.data || []
  } catch {
    toast.error('Gagal memuat mutasi transfer sales')
  } finally {
    loadingTransfers.value = false
  }
}

// ─── DO Actions ──────────────────────────────────────────────────────────────
async function fetchOrCreateDO(batchId: string) {
  loadingDO.value = true
  try {
    const res = await warehouseApi.createOrGetWarehouseDO(batchId)
    const data = res.data.data
    warehouseDO.value = data
    
    if (data.status === 'FINALIZED') {
      isFinalConfirmed.value = true
    } else {
      isFinalConfirmed.value = false
    }
    
    if (data.status === 'DISPATCHED') {
      isDispatched.value = true
    } else {
      isDispatched.value = false
    }
    doNotes.value = data.notes || ''

    doItems.value = (data.items || []).map((item: any) => ({
      product_id: item.product_id,
      product: item.product,
      bonus_inventory_id: item.bonus_inventory_id,
      bonus_inventory: item.bonus_inventory,
      custom_pcs_per_unit: item.custom_pcs_per_unit || item.product?.pcs_per_unit || 1,
      cartons_loaded: item.cartons_loaded,
      pcs_loaded: item.pcs_loaded,
      pcs_returned_to_warehouse: item.pcs_returned_to_warehouse || 0,
      pcs_sold: item.pcs_sold || 0,
      pcs_sample: item.pcs_sample || 0,
      pcs_returned_from_customer: item.pcs_returned_from_customer || 0
    }))
  } catch (e: any) {
    toast.error('Gagal memuat detail lembar DO (Bukti Muat)')
  } finally {
    loadingDO.value = false
  }
}

async function fetchDOById(id: string) {
  loadingDO.value = true
  try {
    const res = await warehouseApi.getWarehouseDOById(id)
    const data = res.data.data
    warehouseDO.value = data
    
    if (data.status === 'FINALIZED') {
      isFinalConfirmed.value = true
    } else {
      isFinalConfirmed.value = false
    }
    
    if (data.status === 'DISPATCHED') {
      isDispatched.value = true
    } else {
      isDispatched.value = false
    }
    doNotes.value = data.notes || ''

    doItems.value = (data.items || []).map((item: any) => ({
      product_id: item.product_id,
      product: item.product,
      bonus_inventory_id: item.bonus_inventory_id,
      bonus_inventory: item.bonus_inventory,
      custom_pcs_per_unit: item.custom_pcs_per_unit || item.product?.pcs_per_unit || 1,
      cartons_loaded: item.cartons_loaded,
      pcs_loaded: item.pcs_loaded,
      pcs_returned_to_warehouse: item.pcs_returned_to_warehouse || 0,
      pcs_sold: item.pcs_sold || 0,
      pcs_sample: item.pcs_sample || 0,
      pcs_returned_from_customer: item.pcs_returned_from_customer || 0
    }))
  } catch (e: any) {
    toast.error('Gagal memuat detail DO Motoris')
  } finally {
    loadingDO.value = false
  }
}

async function saveWarehouseDO() {
  if (!warehouseDO.value || !selectedBatchId.value) return
  savingDO.value = true
  try {
    if (localSparesForDO.value.length > 0) {
      await apiClient.post(`/delivery/batch/${selectedBatchId.value}/spares`, {
        spares: localSparesForDO.value.map(s => ({
          product_id: s.product_id,
          quantity: Number(s.quantity) || 0
        }))
      })
    }

    const payload = {
      id: warehouseDO.value.id,
      notes: doNotes.value.trim(),
      status: warehouseDO.value.status,
      items: doItems.value.map((item) => ({
        product_id: item.product_id,
        bonus_inventory_id: item.bonus_inventory_id,
        cartons_loaded: Number(item.cartons_loaded) || 0,
        pcs_loaded: Number(item.pcs_loaded) || 0,
        custom_pcs_per_unit: Number(item.custom_pcs_per_unit) || 1,
        pcs_returned_to_warehouse: Number(item.pcs_returned_to_warehouse) || 0,
        pcs_sold: Number(item.pcs_sold) || 0,
        pcs_sample: Number(item.pcs_sample) || 0,
        pcs_returned_from_customer: Number(item.pcs_returned_from_customer) || 0
      }))
    }
    await warehouseApi.updateWarehouseDO(payload)
    
    toast.success('Lembar DO & Barang Cadangan berhasil diperbarui!')
    const isMotoris = motorisDOs.value.some(m => m.id === selectedBatchId.value)
    if (isMotoris) {
      await fetchDOById(selectedBatchId.value)
    } else {
      await fetchOrCreateDO(selectedBatchId.value)
    }
    await fetchBatches()
    await fetchMotorisDOs()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyimpan Surat DO')
  } finally {
    savingDO.value = false
  }
}

async function dispatchWarehouseDO() {
  if (!warehouseDO.value || !selectedBatchId.value) return
  if (!confirm('Apakah Anda yakin ingin melakukan Dispatch (Kirim Barang)? Stok akan dipotong dari gudang dan masuk ke stok Sales, dan jumlah "Berangkat" tidak dapat diubah lagi.')) return
  
  savingDO.value = true
  try {
    // Implicit save before dispatch
    const payload = {
      id: warehouseDO.value.id,
      notes: doNotes.value.trim(),
      status: warehouseDO.value.status,
      items: doItems.value.map((item) => ({
        product_id: item.product_id,
        bonus_inventory_id: item.bonus_inventory_id,
        cartons_loaded: Number(item.cartons_loaded) || 0,
        pcs_loaded: Number(item.pcs_loaded) || 0,
        custom_pcs_per_unit: Number(item.custom_pcs_per_unit) || 1,
        pcs_returned_to_warehouse: Number(item.pcs_returned_to_warehouse) || 0,
        pcs_sold: Number(item.pcs_sold) || 0,
        pcs_sample: Number(item.pcs_sample) || 0,
        pcs_returned_from_customer: Number(item.pcs_returned_from_customer) || 0
      }))
    }
    await warehouseApi.updateWarehouseDO(payload)
    const isMotoris = motorisDOs.value.some(m => m.id === selectedBatchId.value)
    if (!isMotoris && localSparesForDO.value.length > 0) {
      await apiClient.post(`/delivery/batch/${selectedBatchId.value}/spares`, {
        spares: localSparesForDO.value.map(s => ({
          product_id: s.product_id,
          quantity: Number(s.quantity) || 0
        }))
      })
    }

    await warehouseApi.dispatchWarehouseDO(warehouseDO.value.id)
    toast.success('DO berhasil di-dispatch!')
    if (isMotoris) {
      await fetchDOById(selectedBatchId.value)
    } else {
      await fetchOrCreateDO(selectedBatchId.value)
    }
    await fetchBatches()
    await fetchMotorisDOs()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal melakukan dispatch DO')
  } finally {
    savingDO.value = false
  }
}

async function finalizeWarehouseDO() {
  if (!warehouseDO.value || !selectedBatchId.value) return
  if (!confirm('Apakah Anda yakin ingin melakukan Finalisasi DO? Pengembalian sisa stok akan dikalkulasi ulang ke gudang dan rekonsiliasi akan dikunci.')) return
  
  savingDO.value = true
  try {
    // Implicit save before finalize
    const payload = {
      id: warehouseDO.value.id,
      notes: doNotes.value.trim(),
      status: warehouseDO.value.status,
      items: doItems.value.map((item) => ({
        product_id: item.product_id,
        bonus_inventory_id: item.bonus_inventory_id,
        cartons_loaded: Number(item.cartons_loaded) || 0,
        pcs_loaded: Number(item.pcs_loaded) || 0,
        custom_pcs_per_unit: Number(item.custom_pcs_per_unit) || 1,
        pcs_returned_to_warehouse: Number(item.pcs_returned_to_warehouse) || 0,
        pcs_sold: Number(item.pcs_sold) || 0,
        pcs_sample: Number(item.pcs_sample) || 0,
        pcs_returned_from_customer: Number(item.pcs_returned_from_customer) || 0
      }))
    }
    await warehouseApi.updateWarehouseDO(payload)
    const isMotoris = motorisDOs.value.some(m => m.id === selectedBatchId.value)
    if (!isMotoris && localSparesForDO.value.length > 0) {
      await apiClient.post(`/delivery/batch/${selectedBatchId.value}/spares`, {
        spares: localSparesForDO.value.map(s => ({
          product_id: s.product_id,
          quantity: Number(s.quantity) || 0
        }))
      })
    }

    await warehouseApi.finalizeWarehouseDO(warehouseDO.value.id)
    toast.success('DO berhasil difinalisasi!')
    if (isMotoris) {
      await fetchDOById(selectedBatchId.value)
    } else {
      await fetchOrCreateDO(selectedBatchId.value)
    }
    await fetchBatches()
    await fetchMotorisDOs()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal melakukan finalisasi DO')
  } finally {
    savingDO.value = false
  }
}

// ─── Spares Actions ──────────────────────────────────────────────────────────
function addSpare() {
  if (!newSpareProductId.value) {
    toast.error('Pilih produk terlebih dahulu')
    return
  }
  if (newSpareQty.value <= 0) {
    toast.error('Jumlah harus lebih dari 0')
    return
  }
  
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
  
  newSpareProductId.value = ''
  newSpareQty.value = 1
}

function removeSpare(index: number) {
  localSpares.value.splice(index, 1)
}

async function saveSpares() {
  if (!selectedBatchId.value) return
  savingSpares.value = true
  try {
    await apiClient.post(`/delivery/batch/${selectedBatchId.value}/spares`, {
      spares: localSpares.value.map(s => ({
        product_id: s.product_id,
        quantity: Number(s.quantity) || 0
      }))
    })
    toast.success('Barang cadangan berhasil disimpan!')
    
    await fetchBatches()
    const updated = batches.value.find(b => b.id === selectedBatchId.value)
    if (updated) {
      localSpares.value = (updated.spares || []).map((s: any) => ({
        product_id: s.product_id,
        quantity: s.quantity
      }))
      syncSparesForDO()
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyimpan barang cadangan')
  } finally {
    savingSpares.value = false
  }
}

// ─── Leftovers Actions ────────────────────────────────────────────────────────
async function slipLeftover(subId: string) {
  if (!selectedBatchId.value) {
    toast.error('Pilih Surat Jalan terlebih dahulu')
    return
  }
  slippingLeftoverId.value = subId
  try {
    await apiClient.post(`/delivery/batch/${selectedBatchId.value}/subsequent/${subId}/slip`)
    toast.success('Barang susulan/leftover berhasil diselipkan ke dalam pengiriman ini!')
    await fetchOrCreateDO(selectedBatchId.value)
    await fetchPendingLeftovers()
    await fetchBatches()
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyelipkan barang susulan')
  } finally {
    slippingLeftoverId.value = null
  }
}

// ─── Sales Transfer Actions ──────────────────────────────────────────────────
async function createSalesTransfer() {
  if (!newTransferEmployeeId.value) {
    toast.error('Pilih Salesman terlebih dahulu')
    return
  }
  if (!newTransferProductId.value) {
    toast.error('Pilih produk terlebih dahulu')
    return
  }
  if (newTransferQty.value <= 0) {
    toast.error('Jumlah harus lebih besar dari 0')
    return
  }
  creatingTransfer.value = true
  try {
    const branchId = masterStore.selectedWarehouseBranchId
    const prod = products.value.find(p => p.id === newTransferProductId.value)
    
    await apiClient.post(`/sales-transfers/branch/${branchId}`, {
      employee_id: newTransferEmployeeId.value,
      product_id: newTransferProductId.value,
      quantity: Number(newTransferQty.value) || 1,
      unit: prod?.unit || 'PCS',
      type: 'TRANSFER',
      notes: newTransferNotes.value
    })
    toast.success('Permintaan transfer stok sales berhasil dibuat!')
    newTransferProductId.value = ''
    newTransferQty.value = 1
    newTransferNotes.value = ''
    await fetchSalesTransfers()
  } catch (err: any) {
    toast.error(err?.response?.data?.error || 'Gagal membuat transfer')
  } finally {
    creatingTransfer.value = false
  }
}

async function slipSalesTransfer(trfId: string) {
  if (!selectedBatchId.value || !selectedBatch.value) {
    toast.error('Pilih Surat Jalan terlebih dahulu')
    return
  }
  slippingTransferId.value = trfId
  try {
    const currentOrders = (selectedBatch.value.items || [])
      .filter((i: any) => i.sales_order_id)
      .map((i: any) => i.sales_order_id)
      
    const currentTrxs = (selectedBatch.value.items || [])
      .filter((i: any) => i.sales_transaction_id)
      .map((i: any) => i.sales_transaction_id)
      
    const currentSubsequents = (selectedBatch.value.items || [])
      .filter((i: any) => i.sales_subsequent_item_id)
      .map((i: any) => i.sales_subsequent_item_id)
      
    const currentTransfers = (selectedBatch.value.items || [])
      .filter((i: any) => i.sales_transfer_id)
      .map((i: any) => i.sales_transfer_id)
      
    if (!currentTransfers.includes(trfId)) {
      currentTransfers.push(trfId)
    }
    
    const payload = {
      driver_id: selectedBatch.value.driver_id,
      vehicle_id: selectedBatch.value.vehicle_id,
      status: selectedBatch.value.status,
      delivery_date: selectedBatch.value.delivery_date,
      sales_order_ids: currentOrders,
      sales_transaction_ids: currentTrxs,
      sales_subsequent_item_ids: currentSubsequents,
      sales_transfer_ids: currentTransfers,
      spares: (selectedBatch.value.spares || []).map((s: any) => ({
        product_id: s.product_id,
        quantity: s.quantity
      }))
    }
    
    await apiClient.put(`/delivery/batches/${selectedBatchId.value}`, payload)
    toast.success('Mutasi stok sales berhasil diselipkan ke dalam Surat Jalan ini!')
    await fetchBatches()
    await fetchSalesTransfers()
  } catch (err: any) {
    toast.error(err?.response?.data?.error || 'Gagal menyelipkan transfer stok')
  } finally {
    slippingTransferId.value = null
  }
}

onMounted(async () => {
  await Promise.all([
    fetchBatches(),
    fetchMotorisDOs(),
    fetchProducts(),
    fetchPendingLeftovers(),
    fetchSalesTransfers(),
    masterStore.fetchEmployees(),
    masterStore.fetchStores(masterStore.selectedBranchCompanyId),
    vehicleStore.fetchVehicles()
  ])
  if (route.query.batch_id) {
    selectedBatchId.value = String(route.query.batch_id)
  }
})

const getErrorMessage = (err: any): string => {
  if (!err) return 'Unknown error'
  if (typeof err === 'string') return err
  if (err.response?.data?.error) return err.response.data.error
  if (err.response?.data?.message) return err.response.data.message
  if (err.message) return err.message
  try {
    return JSON.stringify(err)
  } catch (_) {
    return String(err)
  }
}

function getDoItem(id: string) {
  return doItems.value.find((i: any) => i.product_id === id || i.bonus_inventory_id === id)
}

const printableProducts = computed(() => {
  // 1. Regular products that have target / DO / spares
  const regularItems = products.value.filter((p: any) => {
    const hasTarget = (batchProductQtyMap.value[p.id] || 0) > 0;
    const doItem = getDoItem(p.id);
    const hasDO = doItem && ((doItem.cartons_loaded || 0) > 0 || (doItem.pcs_loaded || 0) > 0);
    const hasSpare = localSparesForDO.value.some((s: any) => s.product_id === p.id && s.total_pcs > 0);
    return hasTarget || hasDO || hasSpare;
  }).map(p => ({
    id: p.id,
    name: p.name,
    sku: p.sku,
    is_bonus: false,
    pcs_per_unit: p.pcs_per_unit || 1
  }));

  // 2. Bonus items from DO or target
  const bonusItems: any[] = [];
  doItems.value.forEach((item: any) => {
    if (item.bonus_inventory_id) {
      const bId = item.bonus_inventory_id;
      const bName = item.bonus_inventory?.name || 'Item Bonus';
      const bUnit = item.bonus_inventory?.unit || 'Pcs';
      bonusItems.push({
        id: bId,
        name: `[BONUS] ${bName} (${bUnit})`,
        sku: 'BONUS',
        is_bonus: true,
        pcs_per_unit: 1
      });
    }
  });

  // Also check if any bonus in batchProductQtyMap is not in doItems
  if (selectedBatch.value?.items) {
    selectedBatch.value.items.forEach((di: any) => {
      if (di.sales_order?.bonus_items) {
        di.sales_order.bonus_items.forEach((b: any) => {
          if (b.bonus_inventory_id && !bonusItems.some(bi => bi.id === b.bonus_inventory_id)) {
            const bName = b.bonus_inventory?.name || 'Item Bonus';
            const bUnit = b.bonus_inventory?.unit || 'Pcs';
            bonusItems.push({
              id: b.bonus_inventory_id,
              name: `[BONUS] ${bName} (${bUnit})`,
              sku: 'BONUS',
              is_bonus: true,
              pcs_per_unit: 1
            });
          }
        });
      }
    });
  }

  return [...regularItems, ...bonusItems];
});

function getPrintableActual(productId: string, type: 'carton' | 'pcs') {
  const doItem = getDoItem(productId);
  const spareItem = localSparesForDO.value.find((s: any) => s.product_id === productId);
  
  let totalCarton = 0;
  let totalPcs = 0;
  
  if (doItem) {
    totalCarton += (doItem.cartons_loaded || 0);
    totalPcs += (doItem.pcs_loaded || 0);
  }
  
  if (spareItem) {
    totalCarton += (spareItem.cartons || 0);
    totalPcs += (spareItem.pcs || 0);
  }
  
  if (totalCarton === 0 && totalPcs === 0) return '';
  
  return type === 'carton' ? totalCarton : totalPcs;
}

function getPrintableDynamicValue(productId: string, field: 'berangkat' | 'pulang' | 'terjual' | 'sample' | 'retur') {
  const doItem = getDoItem(productId);
  const spareItem = localSparesForDO.value.find((s: any) => s.product_id === productId);
  
  if (!doItem && !spareItem) return '';
  
  let totalPcs = 0;
  const pcsPerUnit = doItem?.custom_pcs_per_unit || spareItem?.pcs_per_unit || 1;

  if (field === 'berangkat') {
    if (doItem) {
      totalPcs += (doItem.cartons_loaded || 0) * pcsPerUnit + (doItem.pcs_loaded || 0);
    }
    if (spareItem) {
      totalPcs += (spareItem.cartons || 0) * pcsPerUnit + (spareItem.pcs || 0);
    }
  } else if (doItem) {
    if (field === 'pulang') totalPcs = doItem.pcs_returned_to_warehouse || 0;
    if (field === 'terjual') totalPcs = doItem.pcs_sold || 0;
    if (field === 'sample') totalPcs = doItem.pcs_sample || 0;
    if (field === 'retur') totalPcs = doItem.pcs_returned_from_customer || 0;
  }
  
  if (totalPcs === 0) return '';

  // If pcsPerUnit is 1 or it's a bonus item, format directly as Pcs
  if (pcsPerUnit === 1 || doItem?.bonus_inventory_id) {
    return `${totalPcs} P`;
  }
  
  // Format into compact short string for print (e.g. "2 D, 3 P")
  const cartons = Math.floor(totalPcs / pcsPerUnit);
  const remaining = totalPcs % pcsPerUnit;
  
  if (cartons > 0 && remaining > 0) {
    return `${cartons} D, ${remaining} P`;
  } else if (cartons > 0) {
    return `${cartons} D`;
  } else {
    return `${remaining} P`;
  }
}
</script>

<template>
  <div class="space-y-6 p-6 print:hidden">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-black text-slate-900 tracking-tight">Manajemen Lembar DO &amp;</h1>
        <p class="text-xs text-slate-500 font-semibold mt-1">
          Kelola bukti muat DO, cadangan, dan susulan langsung tanpa harus membuka rincian Surat Jalan.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <button
          v-if="selectedBatchId"
          @click="printDO"
          class="p-2.5 bg-indigo-600 border border-indigo-700 rounded-xl hover:bg-indigo-700 text-white transition-colors shadow-xs flex items-center gap-2"
          title="Cetak DO"
        >
          <Printer class="w-4 h-4" />
          <span class="text-xs font-bold hidden sm:inline">Cetak DO</span>
        </button>
        <button
          @click="() => { fetchBatches(); fetchPendingLeftovers(); fetchSalesTransfers(); }"
          class="p-2.5 bg-white border border-slate-200 rounded-xl hover:bg-slate-50 text-slate-600 transition-colors shadow-xs"
          title="Refresh Data"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loadingBatches }" />
        </button>
      </div>
    </div>

    <!-- Main Workspace Layout: Customer DO -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      
      <!-- LEFT SIDE: SJ Selector list -->
      <div class="lg:col-span-1 bg-white border border-slate-200/60 rounded-3xl p-5 shadow-xs space-y-4 flex flex-col h-[calc(100vh-14rem)] min-h-[480px]">
        <div>
          <h2 class="text-sm font-black text-slate-800 flex items-center gap-2">
            <Truck class="w-4 h-4 text-primary" />
            Pilih Surat Jalan (Active Batch)
          </h2>
          <p class="text-[11px] text-slate-400 mt-0.5">Pilih pengiriman aktif untuk mulai mengelola muatan.</p>
        </div>

        <!-- Search Bar -->
        <div class="relative">
          <Search class="absolute left-3 top-2.5 w-4 h-4 text-slate-400" />
          <input
            v-model="searchSJQuery"
            type="text"
            placeholder="Cari No SJ, supir, plat nomor..."
            class="w-full pl-9 pr-4 py-2 border border-slate-200 rounded-xl text-xs focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary font-medium"
          />
        </div>

        <!-- SJ List Scroll Container -->
        <div class="flex-1 overflow-y-auto space-y-2 pr-1">
          <div v-if="loadingBatches" class="flex flex-col items-center justify-center py-10 text-slate-400 gap-2">
            <RefreshCw class="w-5 h-5 animate-spin" />
            <p class="text-xs font-bold">Memuat daftar SJ...</p>
          </div>
          
          <div v-else-if="filteredBatchesForSelect.length === 0" class="text-center py-10 text-slate-400 text-xs font-semibold">
            Tidak ada Surat Jalan ditemukan
          </div>

          <div
            v-else
            v-for="b in filteredBatchesForSelect"
            :key="b.id"
            @click="selectedBatchId = b.id"
            :class="[
              selectedBatchId === b.id
                ? 'border-primary bg-primary/5 text-primary'
                : 'border-slate-100 hover:border-slate-300 hover:bg-slate-50/50'
            ]"
            class="p-3 border rounded-2xl cursor-pointer transition-all space-y-2"
          >
            <div class="flex justify-between items-start">
              <span class="font-mono text-xs font-bold leading-none select-all" :class="selectedBatchId === b.id ? 'text-primary' : 'text-slate-900'">
                {{ b.delivery_order_no || b.do_number || 'DRAFT-SJ' }}
                <span v-if="b.is_delegation" class="ml-1 bg-orange-100 text-orange-700 text-[8px] px-1.5 py-0.5 rounded font-black tracking-widest uppercase">Delegasi</span>
                <span v-else-if="b.is_motoris || (b.notes && b.notes.includes('Motoris'))" class="ml-1 bg-indigo-100 text-indigo-700 text-[8px] px-1.5 py-0.5 rounded font-black tracking-widest uppercase">Motoris</span>
              </span>
              <span
                class="text-[9px] font-black px-1.5 py-0.5 rounded-lg"
                :class="[
                  b.status === 'SUPERVISOR_APPROVED' || b.status === 'PENDING' || b.status === 'DRAFT' ? 'bg-amber-100 text-amber-800' :
                  b.status === 'PICKING' ? 'bg-indigo-100 text-indigo-800' :
                  b.status === 'ON_DELIVERY' || b.status === 'DISPATCHED' ? 'bg-blue-100 text-blue-800' :
                  'bg-green-100 text-green-800'
                ]"
              >
                {{ b.status }}
              </span>
            </div>
            
            <div class="grid grid-cols-2 gap-2 text-[10px] font-semibold text-slate-500">
              <div class="flex items-center gap-1.5">
                <span class="font-bold text-slate-800 truncate">{{ b.is_motoris && b.employee ? `${b.employee.first_name} ${b.employee.last_name}` : (b.driver ? `${b.driver.first_name} ${b.driver.last_name}` : (b.daily_driver ? b.daily_driver.name : '-')) }}</span>
              </div>
              <div class="text-right">
                <span class="font-mono text-slate-700">{{ b.vehicle?.license_plate || (b.is_motoris ? 'MOTORIS' : '-') }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- RIGHT SIDE: Tab Panel details for currently selected SJ -->
      <div class="lg:col-span-2 space-y-6">
        
        <!-- Batch details summary -->
        <div v-if="selectedBatch" class="bg-slate-900 text-white rounded-3xl p-5 shadow-sm flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
          <div>
            <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest leading-none">Terpilih Untuk Manajemen</p>
            <h3 class="text-lg font-black mt-1 font-mono tracking-tight">{{ selectedBatch.delivery_order_no || selectedBatch.do_number || 'NO_SJ' }}</h3>
            <div class="flex flex-wrap gap-x-4 gap-y-1 mt-2 text-xs text-slate-300 font-semibold">
              <span class="flex items-center gap-1">👤 Supir/Sales: {{ selectedBatch.is_motoris && selectedBatch.employee ? `${selectedBatch.employee.first_name} ${selectedBatch.employee.last_name}` : (selectedBatch.driver ? `${selectedBatch.driver.first_name} ${selectedBatch.driver.last_name}` : (selectedBatch.daily_driver ? selectedBatch.daily_driver.name : '-')) }}</span>
              <span class="flex items-center gap-1">🚛 Plat: {{ selectedBatch.vehicle?.license_plate || (selectedBatch.is_motoris ? 'MOTORIS' : '-') }}</span>
            </div>
          </div>
          <div class="flex gap-2">
            <span v-if="!selectedBatch.is_motoris" class="px-3 py-1 bg-slate-800 text-slate-200 text-xs font-black rounded-xl">
              {{ selectedBatch.items?.length || 0 }} Nota
            </span>
            <span v-if="!selectedBatch.is_motoris" class="px-3 py-1 bg-slate-800 text-slate-200 text-xs font-black rounded-xl">
              {{ selectedBatch.spares?.length || 0 }} Spares
            </span>
            <span v-if="selectedBatch.is_motoris" class="px-3 py-1 bg-indigo-900/50 text-indigo-200 text-xs font-black rounded-xl border border-indigo-700/50">
              DO MOTORIS
            </span>
          </div>
        </div>

        <div v-else class="bg-amber-50 border border-amber-200 rounded-3xl p-6 text-center text-amber-800 font-semibold text-xs">
          Silakan pilih salah satu Surat Jalan di kolom kiri untuk mulai mengelola.
        </div>

        <!-- Right Workspace: Delegation Read-only View -->
        <div v-if="selectedBatch?.is_delegation" class="bg-white border border-slate-200/60 rounded-3xl p-6 shadow-xs space-y-6">
          <div class="border-b border-slate-100 pb-4">
            <h3 class="text-xl font-black text-slate-900 flex items-center gap-2">
              <Send class="w-5 h-5 text-orange-500" /> Detail Surat Jalan Delegasi
            </h3>
            <p class="text-xs text-slate-500 font-medium mt-1">Data armada dan sopir untuk {{ selectedBatch.delivery_order_no }} dari Pabrik.</p>
          </div>
          
          <div class="grid grid-cols-2 gap-4">
            <div class="bg-slate-50 rounded-2xl p-4 border border-slate-100">
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Kendaraan</p>
              <p class="text-sm font-bold text-slate-800">
                {{ selectedBatch.vehicle?.name || '-' }}
                <span v-if="selectedBatch.vehicle?.license_plate" class="text-slate-500">({{ selectedBatch.vehicle.license_plate }})</span>
              </p>
            </div>
            <div class="bg-slate-50 rounded-2xl p-4 border border-slate-100">
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">Sopir</p>
              <p class="text-sm font-bold text-slate-800">
                {{ selectedBatch.driver?.first_name || '-' }} {{ selectedBatch.driver?.last_name || '' }}
              </p>
            </div>
          </div>

          <!-- New section for displaying items -->
          <div class="space-y-3 mt-6">
            <h4 class="text-sm font-black text-slate-800 flex items-center gap-2">
              <Package class="w-4 h-4 text-primary" /> Daftar Barang
              <span class="bg-slate-100 text-slate-600 text-[10px] px-2 py-0.5 rounded-full">{{ selectedBatch.items?.length || 0 }} SKU</span>
            </h4>
            <div v-if="selectedBatch.items?.length > 0" class="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-[300px] overflow-y-auto custom-scrollbar pr-2">
              <div v-for="item in selectedBatch.items" :key="item.id" class="bg-slate-50 border border-slate-100 rounded-2xl p-4 flex justify-between items-center hover:bg-slate-100/50 transition-colors">
                <div class="min-w-0 pr-2">
                  <p class="text-xs font-black text-slate-800 truncate">{{ item.product?.name || 'Produk Tidak Diketahui' }}</p>
                  <p class="text-[10px] text-slate-500 font-bold mt-0.5">Batch: <span class="font-mono">{{ item.batch_no || '-' }}</span></p>
                </div>
                <div class="text-right shrink-0">
                  <p class="text-sm font-black text-primary">{{ item.quantity }} <span class="text-[10px] text-slate-500">{{ item.unit || 'PCS' }}</span></p>
                  <p v-if="(item.pcs_per_unit || 1) > 1" class="text-[9px] font-bold text-slate-400 mt-0.5">Total: {{ item.quantity * item.pcs_per_unit }} PCS</p>
                </div>
              </div>
            </div>
            <div v-else class="text-center py-6 bg-slate-50 rounded-2xl border border-dashed border-slate-200">
              <p class="text-xs font-semibold text-slate-400">Tidak ada barang dalam daftar pengiriman ini.</p>
            </div>
          </div>

          <div class="pt-4 flex justify-end">
            <Button 
              class="rounded-xl h-11 px-8 font-black bg-slate-100 text-slate-700 hover:bg-slate-200 shadow-sm" 
              @click="() => {}"
            >
              <Printer class="w-4 h-4 mr-2" /> Cetak SJ Delegasi
            </Button>
          </div>
        </div>

        <!-- Interactive Tabs for DO, Spares, and Leftovers -->
        <div v-else-if="selectedBatch" class="bg-white border border-slate-200/60 rounded-3xl p-6 shadow-xs space-y-6">
          
          <!-- Tab Navigation -->
          <div class="flex border-b border-slate-200/60 -mx-6 px-6 pb-0 overflow-x-auto">
            <button
              @click="activeTab = 'do'"
              :class="activeTab === 'do' ? 'border-primary text-primary' : 'border-transparent text-slate-500 hover:text-slate-700'"
              class="py-3 px-4 text-xs font-black border-b-2 -mb-px transition-colors flex items-center gap-1.5 focus:outline-none whitespace-nowrap"
            >
              <ShieldCheck class="w-4 h-4" />
              Lembar DO (Bukti Muat)
            </button>
            <button
              @click="activeTab = 'spares'"
              :class="activeTab === 'spares' ? 'border-primary text-primary' : 'border-transparent text-slate-500 hover:text-slate-700'"
              class="py-3 px-4 text-xs font-black border-b-2 -mb-px transition-colors flex items-center gap-1.5 focus:outline-none whitespace-nowrap"
            >
              <BoxSelect class="w-4 h-4" />
              Tambah Barang Cadangan
            </button>
            <button
              @click="activeTab = 'leftovers'"
              :class="activeTab === 'leftovers' ? 'border-primary text-primary' : 'border-transparent text-slate-500 hover:text-slate-700'"
              class="py-3 px-4 text-xs font-black border-b-2 -mb-px transition-colors flex items-center gap-1.5 focus:outline-none whitespace-nowrap"
            >
              <Clock class="w-4 h-4" />
              Tambah Barang Susulan
              <span v-if="pendingLeftovers.length > 0" class="bg-rose-100 text-rose-800 text-[9px] px-1.5 py-0.5 rounded-full font-bold ml-1">
                {{ pendingLeftovers.length }}
              </span>
            </button>
            <!-- Kirim Stok Sales (Mutasi) tab option -->
            <button
              @click="activeTab = 'transfers'"
              :class="activeTab === 'transfers' ? 'border-primary text-primary' : 'border-transparent text-slate-500 hover:text-slate-700'"
              class="py-3 px-4 text-xs font-black border-b-2 -mb-px transition-colors flex items-center gap-1.5 focus:outline-none whitespace-nowrap"
            >
              <BoxSelect class="w-4 h-4 text-emerald-600" />
              Kirim Stok Sales (Mutasi)
              <span v-if="pendingSalesTransfers.length > 0" class="bg-emerald-100 text-emerald-800 text-[9px] px-1.5 py-0.5 rounded-full font-bold ml-1">
                {{ pendingSalesTransfers.length }}
              </span>
            </button>
          </div>

          <!-- TAB CONTENT: DO (Bukti Muat) -->
          <div v-if="activeTab === 'do'" class="space-y-4">
            <div v-if="loadingDO" class="flex flex-col items-center py-10 text-slate-400 gap-2">
              <RefreshCw class="w-6 h-6 animate-spin text-slate-400" />
              <p class="text-xs font-bold">Memuat data lembar DO...</p>
            </div>

            <div v-else class="space-y-4">
              <!-- DO header/status card -->
              <div class="bg-amber-50 border border-amber-200 rounded-2xl p-4 flex flex-col md:flex-row md:items-center justify-between gap-3 shadow-xs">
                <div>
                  <div class="flex items-center gap-2">
                    <p class="text-[10px] font-black text-amber-700 uppercase tracking-widest leading-none">Nomor DO Terbit</p>
                    <span v-if="isFinalConfirmed" class="bg-emerald-100 text-emerald-800 text-[9px] px-1.5 py-0.5 rounded-md font-black">
                      ✓ FINAL SORE
                    </span>
                  </div>
                  <p class="text-base font-black text-slate-900 font-mono mt-1">{{ warehouseDO?.do_number || 'Memproses DO...' }}</p>
                  <p class="text-[10px] text-slate-500 font-semibold mt-0.5">Input jumlah barang muatan sore/akhir hari. Gunakan format Dus dan Pcs.</p>
                </div>
                
                <div class="flex items-center gap-2">
                  <button
                    v-if="!isFinalConfirmed && isDispatched && pendingReturns.length > 0"
                    @click="syncReturnsFromMobile"
                    class="px-3 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-black rounded-xl transition-all shadow-md flex items-center justify-center gap-1.5 whitespace-nowrap print:hidden mr-2"
                  >
                    <span>🔄 Tarik Request Mobile ({{ pendingReturns.length }})</span>
                  </button>

                  <button
                    v-if="!isFinalConfirmed && isDispatched"
                    @click="finalizeWarehouseDO"
                    :disabled="savingDO"
                    class="px-3 py-2 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white text-xs font-black rounded-xl transition-all shadow-md flex items-center justify-center gap-1.5 whitespace-nowrap print:hidden"
                  >
                    <RefreshCw v-if="savingDO" class="w-3.5 h-3.5 animate-spin" />
                    <span v-else>✅ Finalisasi DO (Tutup DO)</span>
                  </button>

                  <button
                    v-if="!isFinalConfirmed && !isDispatched"
                    @click="dispatchWarehouseDO"
                    :disabled="savingDO"
                    class="px-3 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-xs font-black rounded-xl transition-all shadow-md flex items-center justify-center gap-1.5 whitespace-nowrap print:hidden"
                  >
                    <RefreshCw v-if="savingDO" class="w-3.5 h-3.5 animate-spin" />
                    <span v-else>Muat DO Berangkat</span>
                  </button>

                  <button
                    v-if="!isFinalConfirmed"
                    @click="saveWarehouseDO"
                    :disabled="savingDO"
                    class="px-4 py-2 bg-amber-500 hover:bg-amber-600 disabled:opacity-50 text-white text-xs font-black rounded-xl transition-all shadow-md flex items-center justify-center gap-1.5 whitespace-nowrap print:hidden"
                  >
                    <RefreshCw v-if="savingDO" class="w-3.5 h-3.5 animate-spin" />
                    <span v-else>Simpan DO Sementara</span>
                  </button>
                </div>
              </div>

              <!-- DO Main Product Table -->
              <div class="bg-white border border-slate-250 rounded-2xl p-4 shadow-sm space-y-3">
                <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5 border-b border-slate-100 pb-2">
                  <Package class="w-4 h-4 text-slate-500" />
                  Rincian Barang Orderan yang Dimuat
                </h4>
                
                <div class="overflow-x-auto">
                  <table class="w-full text-left border-collapse">
                    <thead>
                      <tr class="border-b border-slate-200 text-[10px] uppercase tracking-widest text-slate-400">
                        <th class="py-2">Nama Produk</th>
                        <th class="py-2 text-center w-40">Target Pesanan SJ</th>
                        <th class="py-2 text-center w-28 text-blue-600">Berangkat (Dimuat)</th>
                        <th class="py-2 text-center w-20 text-emerald-600 border-l border-slate-200 pl-2">Pulang (Pcs)</th>
                        <th class="py-2 text-center w-20 text-indigo-600">Terjual (Pcs)</th>
                        <th class="py-2 text-center w-20 text-amber-600">Sampel (Pcs)</th>
                        <th class="py-2 text-center w-20 text-rose-600">Retur (Pcs)</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="item in doItems" :key="item.product_id || item.bonus_inventory_id" class="border-b border-slate-100 last:border-0 hover:bg-slate-50/50 transition-colors">
                        <td class="py-3 text-xs font-bold text-slate-800 leading-tight">
                          <span v-if="item.product_id">
                            {{ item.product?.name || 'Produk Tanpa Nama' }}
                            <p class="text-[9px] text-slate-400 font-mono mt-0.5">SKU: {{ item.product?.sku }}</p>
                          </span>
                          <span v-else-if="item.bonus_inventory_id" class="text-amber-600">
                            {{ item.bonus_inventory?.name || 'Bonus Inventory' }} <span class="px-1.5 py-0.5 bg-amber-100 text-amber-700 rounded text-[10px] font-bold">BONUS</span>
                          </span>
                        </td>
                        
                        <td class="py-3 text-center text-xs font-bold text-slate-650">
                          {{ batchProductQtyMap[item.product_id || item.bonus_inventory_id] || 0 }} Pcs
                          <p class="text-[9px] text-indigo-650 font-medium">
                            ({{ formatCartonPcs(batchProductQtyMap[item.product_id || item.bonus_inventory_id] || 0, item.custom_pcs_per_unit || 1) }})
                          </p>
                        </td>

                        <td class="py-3 text-center">
                          <button @click="openDynamicModal(item, 'berangkat')" class="w-full flex flex-col items-center justify-center p-1.5 rounded-lg border border-dashed border-slate-300 hover:border-blue-400 hover:bg-blue-50 transition-colors">
                            <span class="text-xs font-black text-blue-700">{{ (item.cartons_loaded * (item.custom_pcs_per_unit || 1)) + item.pcs_loaded }} Pcs</span>
                            <span class="text-[9px] text-slate-500 font-semibold">{{ item.cartons_loaded }} Dus, {{ item.pcs_loaded }} Pcs</span>
                          </button>
                        </td>

                        <td class="py-3 text-center border-l border-slate-100 pl-2">
                          <button @click="openDynamicModal(item, 'pulang')" :disabled="isFinalConfirmed" class="w-full flex flex-col items-center justify-center p-1.5 rounded-lg border border-dashed border-slate-300 hover:border-emerald-400 hover:bg-emerald-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="text-xs font-black text-emerald-700">{{ item.pcs_returned_to_warehouse || 0 }} Pcs</span>
                            <span class="text-[9px] text-slate-500 font-semibold">{{ formatCartonPcs(item.pcs_returned_to_warehouse || 0, item.custom_pcs_per_unit || 1) }}</span>
                          </button>
                        </td>

                        <td class="py-3 text-center">
                          <button @click="openDynamicModal(item, 'terjual')" :disabled="isFinalConfirmed" class="w-full flex flex-col items-center justify-center p-1.5 rounded-lg border border-dashed border-slate-300 hover:border-indigo-400 hover:bg-indigo-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="text-xs font-black text-indigo-700">{{ item.pcs_sold || 0 }} Pcs</span>
                            <span class="text-[9px] text-slate-500 font-semibold">{{ formatCartonPcs(item.pcs_sold || 0, item.custom_pcs_per_unit || 1) }}</span>
                          </button>
                        </td>

                        <td class="py-3 text-center">
                          <button @click="openDynamicModal(item, 'sample')" :disabled="isFinalConfirmed" class="w-full flex flex-col items-center justify-center p-1.5 rounded-lg border border-dashed border-slate-300 hover:border-amber-400 hover:bg-amber-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="text-xs font-black text-amber-700">{{ item.pcs_sample || 0 }} Pcs</span>
                            <span class="text-[9px] text-slate-500 font-semibold">{{ formatCartonPcs(item.pcs_sample || 0, item.custom_pcs_per_unit || 1) }}</span>
                          </button>
                        </td>

                        <td class="py-3 text-center">
                          <button @click="openDynamicModal(item, 'retur')" :disabled="isFinalConfirmed" class="w-full flex flex-col items-center justify-center p-1.5 rounded-lg border border-dashed border-slate-300 hover:border-rose-400 hover:bg-rose-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                            <span class="text-xs font-black text-rose-700">{{ item.pcs_returned_from_customer || 0 }} Pcs</span>
                            <span class="text-[9px] text-slate-500 font-semibold">{{ formatCartonPcs(item.pcs_returned_from_customer || 0, item.custom_pcs_per_unit || 1) }}</span>
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- DO Spares Table: Separate and informative spares tracking within fleet load -->
              <div class="bg-indigo-50/40 border border-indigo-150 rounded-2xl p-4 shadow-sm space-y-3">
                <h4 class="text-xs font-black text-indigo-900 flex items-center gap-1.5 border-b border-indigo-100 pb-2">
                  <BoxSelect class="w-4 h-4 text-indigo-600" />
                  Rincian Barang Cadangan yang Dimuat 
                </h4>
                
                <div v-if="localSparesForDO.length === 0" class="text-center py-6 text-slate-400 text-xs font-semibold">
                  Belum ada barang cadangan yang ditambahkan. (Gunakan tab "Tambah Barang Cadangan" untuk menambahkan).
                </div>
                
                <div v-else class="overflow-x-auto">
                  <table class="w-full text-left border-collapse">
                    <thead>
                      <tr class="border-b border-indigo-200 text-[10px] uppercase tracking-widest text-indigo-800">
                        <th class="py-2">Nama Produk (Barang Cadangan)</th>
                        <th class="py-2 text-center w-28">Kardus/Krat</th>
                        <th class="py-2 text-center w-24">Sisa Pcs</th>
                        <th class="py-2 text-center w-28">Total Muatan Cadangan</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="spare in localSparesForDO" :key="spare.product_id" class="border-b border-indigo-50 last:border-0 hover:bg-white/60 transition-colors">
                        <td class="py-3 text-xs font-bold text-slate-800 leading-tight">
                          {{ spare.product_name }}
                          <p class="text-[9px] text-slate-400 font-mono mt-0.5">SKU: {{ spare.sku }} | 1 Dus = {{ spare.pcs_per_unit }} Pcs</p>
                        </td>
                        
                        <td class="py-3 text-center">
                          <div class="flex items-center justify-center gap-1">
                            <input
                              v-model.number="spare.cartons"
                              type="number"
                              min="0"
                              class="w-20 px-2.5 py-1.5 text-center border border-indigo-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 text-xs font-black"
                              @input="updateSpareTotal(spare)"
                            />
                            <span class="text-[10px] text-slate-400 font-semibold">dus</span>
                          </div>
                        </td>
                        
                        <td class="py-3 text-center">
                          <div class="flex items-center justify-center gap-1">
                            <input
                              v-model.number="spare.pcs"
                              type="number"
                              min="0"
                              class="w-16 px-2 py-1.5 text-center border border-indigo-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 text-xs font-black"
                              @input="updateSpareTotal(spare)"
                            />
                            <span class="text-[10px] text-slate-400 font-semibold">pcs</span>
                          </div>
                        </td>

                        <td class="py-3 text-center text-xs font-black text-indigo-900">
                          {{ spare.quantity }} Pcs
                          <p class="text-[9px] text-slate-400 font-semibold">
                            ({{ formatCartonPcs(spare.quantity, spare.pcs_per_unit) }})
                          </p>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <!-- DO Bonus Items Summary Table (Informational only) -->
              <div v-if="batchBonusMap.length > 0" class="bg-emerald-50/40 border border-emerald-150 rounded-2xl p-4 shadow-sm space-y-3">
                <h4 class="text-xs font-black text-emerald-900 flex items-center gap-1.5 border-b border-emerald-100 pb-2">
                  <Package class="w-4 h-4 text-emerald-600" />
                  Rincian Item Bonus (Harus Dimuat ke Armada)
                </h4>
                
                <div class="overflow-x-auto">
                  <table class="w-full text-left border-collapse">
                    <thead>
                      <tr class="border-b border-emerald-200 text-[10px] uppercase tracking-widest text-emerald-800">
                        <th class="py-2">Nama Bonus Item</th>
                        <th class="py-2 text-center">Satuan</th>
                        <th class="py-2 text-center w-32">Total Harus Dimuat</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="bonus in batchBonusMap" :key="bonus.id" class="border-b border-emerald-50 last:border-0 hover:bg-white/60 transition-colors">
                        <td class="py-3 text-xs font-bold text-slate-800 leading-tight">
                          <span class="text-[9px] font-black text-emerald-700 bg-emerald-100 px-1.5 py-0.5 rounded mr-1">BONUS</span>
                          {{ bonus.name }}
                        </td>
                        <td class="py-3 text-center text-xs font-semibold text-slate-500 uppercase">
                          {{ bonus.unit }}
                        </td>
                        <td class="py-3 text-center text-xs font-black text-emerald-900">
                          {{ bonus.quantity }} {{ bonus.unit }}
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
                  class="w-full p-3 border border-slate-200 rounded-2xl text-xs focus:outline-none focus:ring-1 focus:ring-primary font-medium"
                ></textarea>
              </div>
            </div>
          </div>

          <!-- TAB CONTENT: Spares Management -->
          <div v-else-if="activeTab === 'spares'" class="space-y-6">
            <div class="bg-indigo-50 border border-indigo-100 rounded-2xl p-4 flex flex-col md:flex-row md:items-center justify-between gap-3 shadow-xs">
              <div>
                <p class="text-[10px] font-black text-indigo-700 uppercase tracking-widest leading-none">Manajemen Barang Cadangan (Spares)</p>
                <p class="text-xs text-slate-600 font-semibold mt-1">
                  Tambahkan barang cadangan tambahan untuk tim pengiriman jika ada retur instan di toko.
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
                <BoxSelect class="w-4 h-4 text-indigo-650" />
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

              <div v-if="localSpares.length === 0" class="text-center py-8 text-slate-400 text-xs font-semibold">
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
                          class="p-1.5 bg-red-50 hover:bg-red-100 text-red-650 rounded-lg transition-colors font-black text-[10px]"
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

          <!-- TAB CONTENT: Leftovers -->
          <div v-else-if="activeTab === 'leftovers'" class="space-y-4">
            <div class="bg-indigo-50 border border-indigo-100 rounded-2xl p-4 shadow-xs">
              <p class="text-[10px] font-black text-indigo-700 uppercase tracking-widest leading-none">Leftover &amp; Pengiriman Tertunda (Cabang Ini)</p>
              <p class="text-xs text-slate-600 font-semibold mt-1.5">
                Berikut adalah daftar barang retur / klaim garansi yang disetujui untuk dikirim susulan. Anda dapat menyelipkan barang-barang ini ke dalam batch pengiriman supir saat ini jika muatan masih mencukupi.
              </p>
            </div>

            <div class="bg-white border border-slate-200 rounded-2xl p-4 shadow-xs space-y-3">
              <div class="flex items-center justify-between pb-2 border-b border-slate-100">
                <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5">
                  <Clock class="w-4 h-4 text-slate-500" />
                  Daftar Barang Susulan Tersedia
                </h4>
                <button
                  @click="showManualLeftoverModal = true"
                  class="flex items-center gap-1.5 px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 text-xs font-black rounded-xl transition-colors shadow-sm"
                >
                  <Plus class="w-3.5 h-3.5" />
                  Susulkan Barang
                </button>
              </div>

              <div v-if="pendingLeftovers.length === 0" class="text-center py-8 text-slate-400 text-xs font-semibold">
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
                          {{ item.sales_order?.store?.name || item.product_exchange?.store?.name || item.store?.name || 'Toko' }}
                        </div>
                        <div class="text-[10px] text-slate-400 font-mono">
                          {{ item.sales_order?.so_number || item.product_exchange?.exchange_no || 'Manual' }}
                        </div>
                        <div v-if="item.notes" class="text-[9px] text-rose-500 font-semibold mt-0.5 line-clamp-1" :title="item.notes">
                          Catatan: {{ item.notes }}
                        </div>
                      </td>
                      <td class="py-3 text-xs font-semibold">
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

          <!-- TAB CONTENT: Kirim Stok Sales (Mutasi) -->
          <div v-else-if="activeTab === 'transfers'" class="space-y-6">
            <div class="bg-emerald-50 border border-emerald-100 rounded-2xl p-4 shadow-xs">
              <p class="text-[10px] font-black text-emerald-800 uppercase tracking-widest leading-none">Mutasi Pengiriman Stok Ke Rumah Sales (Mini Depo)</p>
              <p class="text-xs text-slate-650 font-semibold mt-1">
                Kirim stok tambahan untuk sales jarak jauh agar dapat disimpan di rumah dan dijual langsung.
              </p>
            </div>

            <!-- Form Create Transfer -->
            <div class="bg-white border border-slate-200 rounded-2xl p-5 shadow-xs space-y-4">
              <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5 pb-2 border-b border-slate-100">
                <Package class="w-4 h-4 text-emerald-600" />
                Buat Permintaan Pengiriman Stok Baru
              </h4>
              
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div class="space-y-1">
                  <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Pilih Salesman (Penerima)</label>
                  <select
                    v-model="newTransferEmployeeId"
                    class="w-full p-2.5 border border-slate-200 rounded-xl text-xs font-bold focus:outline-none focus:ring-1 focus:ring-emerald-500 bg-white"
                  >
                    <option value="" disabled>-- Pilih Salesman --</option>
                    <option v-for="emp in salesEmployees" :key="emp.id" :value="emp.id">
                      {{ emp.first_name }} {{ emp.last_name }} ({{ emp.job_position?.title || 'Sales' }})
                    </option>
                  </select>
                </div>

                <div class="space-y-1">
                  <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Pilih Produk</label>
                  <select
                    v-model="newTransferProductId"
                    class="w-full p-2.5 border border-slate-200 rounded-xl text-xs font-bold focus:outline-none focus:ring-1 focus:ring-emerald-500 bg-white"
                  >
                    <option value="" disabled>-- Pilih Produk --</option>
                    <option v-for="p in products" :key="p.id" :value="p.id">
                      {{ p.name }} ({{ p.sku }}) - Stok: {{ p.warehouse_stock ?? 0 }} pcs
                    </option>
                  </select>
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 items-end">
                <div class="sm:col-span-1 space-y-1">
                  <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Jumlah (Pcs)</label>
                  <input
                    v-model.number="newTransferQty"
                    type="number"
                    min="1"
                    class="w-full p-2.5 border border-slate-200 rounded-xl text-xs font-bold focus:outline-none focus:ring-1 focus:ring-emerald-500 text-center font-mono"
                  />
                </div>

                <div class="sm:col-span-2 space-y-1">
                  <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Catatan Mutasi</label>
                  <input
                    v-model="newTransferNotes"
                    type="text"
                    placeholder="Contoh: Tambah stok mingguan sales motoris..."
                    class="w-full p-2.5 border border-slate-200 rounded-xl text-xs font-bold focus:outline-none focus:ring-1 focus:ring-emerald-500"
                  />
                </div>
              </div>

              <div class="flex justify-end">
                <button
                  @click="createSalesTransfer"
                  :disabled="creatingTransfer"
                  class="px-5 py-2.5 bg-emerald-650 hover:bg-emerald-700 disabled:opacity-50 text-white text-xs font-black rounded-xl transition-all shadow-sm flex items-center gap-1.5"
                >
                  <RefreshCw v-if="creatingTransfer" class="w-3.5 h-3.5 animate-spin" />
                  <span>Kirim Permintaan Stok</span>
                </button>
              </div>
            </div>

            <!-- List Pending Transfers -->
            <div class="bg-white border border-slate-200 rounded-2xl p-5 shadow-xs space-y-3">
              <h4 class="text-xs font-black text-slate-800 flex items-center gap-1.5 pb-2 border-b border-slate-100">
                <Clock class="w-4 h-4 text-slate-500" />
                Permintaan Kirim Stok Tertunda (Pending)
              </h4>

              <div v-if="pendingSalesTransfers.length === 0" class="text-center py-8 text-slate-400 text-xs font-semibold">
                Tidak ada permintaan kirim stok sales yang tertunda saat ini.
              </div>

              <div v-else class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                  <thead>
                    <tr class="border-b border-slate-200 text-[10px] uppercase tracking-widest text-slate-400">
                      <th class="py-2">Penerima (Sales)</th>
                      <th class="py-2">Produk</th>
                      <th class="py-2 text-center w-24">Jumlah</th>
                      <th class="py-2">Catatan</th>
                      <th class="py-2 text-center w-24">Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="t in pendingSalesTransfers" :key="t.id" class="border-b border-slate-100 last:border-0 hover:bg-slate-50/50 transition-colors">
                      <td class="py-3 text-xs font-bold text-slate-800">
                        {{ t.employee?.first_name }} {{ t.employee?.last_name }}
                        <p class="text-[9px] text-slate-450 font-medium">Reff: {{ t.reference_no || '-' }}</p>
                      </td>
                      <td class="py-3 text-xs font-semibold">
                        <div class="font-bold text-indigo-750 leading-tight">
                          {{ t.product?.name || 'Produk' }}
                        </div>
                        <div class="text-[9px] font-mono text-slate-450">
                          SKU: {{ t.product?.sku || '-' }}
                        </div>
                      </td>
                      <td class="py-3 text-center text-xs font-black text-slate-700">
                        {{ t.quantity }} Pcs
                      </td>
                      <td class="py-3 text-xs text-slate-500 italic max-w-xs truncate">
                        {{ t.notes || '-' }}
                      </td>
                      <td class="py-3 text-center">
                        <button
                          @click="slipSalesTransfer(t.id)"
                          :disabled="slippingTransferId === t.id"
                          class="px-2.5 py-1.5 bg-emerald-50 hover:bg-emerald-100 disabled:opacity-50 text-emerald-600 rounded-lg transition-colors font-black text-[10px] whitespace-nowrap"
                        >
                          <span v-if="slippingTransferId === t.id">Menyelipkan...</span>
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
      </div>
      
    </div>
  </div>

  <!-- Dedicated Print Only Layout (A4 format checklist of all products) -->
  <div class="hidden print:block bg-white text-black font-sans w-full">
    <div class="border-b-2 border-black pb-4 mb-4">
      <h2 class="text-2xl font-black uppercase text-center mb-2 tracking-wide">LEMBAR BUKTI MUAT KENDARAAN (DO)</h2>
      <div class="flex justify-between text-sm mt-4">
         <div>
            <p><strong>NO DO:</strong> {{ warehouseDO?.do_number || '-' }}</p>
            <p><strong>NO SURAT JALAN:</strong> {{ selectedBatch?.delivery_order_no || '-' }}</p>
         </div>
         <div class="text-right">
            <p><strong>Tanggal:</strong> {{ new Date().toLocaleDateString('id-ID') }}</p>
            <p><strong>Supir / Plat:</strong> {{ selectedBatch?.driver ? `${selectedBatch?.driver?.first_name} ${selectedBatch?.driver?.last_name}` : (selectedBatch?.daily_driver ? selectedBatch?.daily_driver?.name : '-') }} / {{ selectedBatch?.vehicle?.license_plate || '-' }}</p>
         </div>
      </div>
    </div>
    
    <table class="w-full text-left border-collapse border border-black text-[10px]">
      <thead>
        <tr>
           <th class="border border-black p-1.5 w-8 text-center">NO</th>
           <th class="border border-black p-1.5">BARANG</th>
           <th class="border border-black p-1.5 w-20 text-center">BERANGKAT</th>
           <th class="border border-black p-1.5 w-16 text-center">PULANG</th>
           <th class="border border-black p-1.5 w-16 text-center">TERJUAL</th>
           <th class="border border-black p-1.5 w-16 text-center">SAMPLE</th>
           <th class="border border-black p-1.5 w-16 text-center">RETUR</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="i in Math.max(37, printableProducts.length)" :key="i">
           <td class="border border-black p-1 text-center font-semibold">{{ i }}</td>
           <td class="border border-black p-1">
              <div v-if="printableProducts[i - 1]">
                <span class="font-bold text-[11px] leading-tight">{{ printableProducts[i - 1].name }}</span>
              </div>
           </td>
           <td class="border border-black p-1 text-center font-bold text-xs">
              <span v-if="printableProducts[i - 1]">{{ getPrintableDynamicValue(printableProducts[i - 1].id, 'berangkat') }}</span>
           </td>
           <td class="border border-black p-1 text-center font-bold text-xs">
              <span v-if="printableProducts[i - 1]">{{ getPrintableDynamicValue(printableProducts[i - 1].id, 'pulang') }}</span>
           </td>
           <td class="border border-black p-1 text-center font-bold text-xs">
              <span v-if="printableProducts[i - 1]">{{ getPrintableDynamicValue(printableProducts[i - 1].id, 'terjual') }}</span>
           </td>
           <td class="border border-black p-1 text-center font-bold text-xs">
              <span v-if="printableProducts[i - 1]">{{ getPrintableDynamicValue(printableProducts[i - 1].id, 'sample') }}</span>
           </td>
           <td class="border border-black p-1 text-center font-bold text-xs">
              <span v-if="printableProducts[i - 1]">{{ getPrintableDynamicValue(printableProducts[i - 1].id, 'retur') }}</span>
           </td>
        </tr>
      </tbody>
    </table>
    
    <div class="mt-8 flex justify-between px-10 text-sm">
       <div class="text-center">
          <p>Disiapkan Oleh (Gudang)</p>
          <br><br><br>
          <p>(.........................)</p>
       </div>
       <div class="text-center">
          <p>Diperiksa Oleh (Checker)</p>
          <br><br><br>
          <p>(.........................)</p>
       </div>
       <div class="text-center">
          <p>Diterima Oleh (Supir)</p>
          <br><br><br>
          <p>(.........................)</p>
       </div>
    </div>
  </div>

  <!-- Manual Leftover Modal -->
  <div v-if="showManualLeftoverModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 print:hidden">
    <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm" @click="showManualLeftoverModal = false"></div>
    <div class="bg-white rounded-2xl shadow-xl w-full max-w-md relative z-10 overflow-hidden flex flex-col max-h-[90vh]">
      <div class="p-4 border-b border-slate-100 bg-slate-50/50 flex justify-between items-center shrink-0">
        <h3 class="text-sm font-black text-slate-800 flex items-center gap-2">
          <Plus class="w-4 h-4 text-indigo-500" />
          Buat Susulan Barang (Manual)
        </h3>
        <button @click="showManualLeftoverModal = false" class="text-slate-400 hover:text-slate-600 transition-colors">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      
      <div class="p-5 overflow-y-auto space-y-4">
        <div>
          <label class="block text-[11px] font-bold text-slate-600 uppercase tracking-widest mb-1.5">Salesman</label>
          <select
            v-model="manualLeftoverForm.salesman_id"
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-sm font-bold text-slate-800 outline-none focus:border-indigo-500 transition-colors"
          >
            <option value="" disabled>Pilih Salesman</option>
            <option v-for="emp in salesEmployees" :key="emp.id" :value="emp.id">
              {{ emp.first_name }} {{ emp.last_name }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-[11px] font-bold text-slate-600 uppercase tracking-widest mb-1.5">Toko (Store)</label>
          <select
            v-model="manualLeftoverForm.store_id"
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-sm font-bold text-slate-800 outline-none focus:border-indigo-500 transition-colors"
          >
            <option value="" disabled>Pilih Toko</option>
            <option v-for="store in masterStore.stores" :key="store.id" :value="store.id">
              {{ store.name }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-[11px] font-bold text-slate-600 uppercase tracking-widest mb-1.5">Barang (Produk)</label>
          <select
            v-model="manualLeftoverForm.product_id"
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-sm font-bold text-slate-800 outline-none focus:border-indigo-500 transition-colors"
          >
            <option value="" disabled>Pilih Produk</option>
            <option v-for="prod in products" :key="prod.id" :value="prod.id">
              {{ prod.name }} (SKU: {{ prod.sku }})
            </option>
          </select>
        </div>

        <div>
          <label class="block text-[11px] font-bold text-slate-600 uppercase tracking-widest mb-1.5">Kuantitas (Pcs)</label>
          <input
            v-model.number="manualLeftoverForm.quantity"
            type="number"
            min="1"
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-sm font-bold text-slate-800 outline-none focus:border-indigo-500 transition-colors"
          />
        </div>

        <div>
          <label class="block text-[11px] font-bold text-slate-600 uppercase tracking-widest mb-1.5">Catatan Susulan</label>
          <textarea
            v-model="manualLeftoverForm.notes"
            rows="3"
            placeholder="Tambahkan catatan kenapa barang ini menjadi susulan agar tim pengiriman tahu..."
            class="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-sm font-semibold text-slate-800 outline-none focus:border-indigo-500 transition-colors resize-none"
          ></textarea>
        </div>
      </div>
      
      <div class="p-4 border-t border-slate-100 flex justify-end gap-2 bg-slate-50/50 shrink-0">
        <button
          @click="showManualLeftoverModal = false"
          class="px-4 py-2 text-sm font-bold text-slate-600 hover:text-slate-800 transition-colors"
        >
          Batal
        </button>
        <button
          @click="submitManualLeftover"
          :disabled="creatingManualLeftover"
          class="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-sm font-black rounded-xl transition-all shadow-sm flex items-center gap-2"
        >
          <RefreshCw v-if="creatingManualLeftover" class="w-4 h-4 animate-spin" />
          <span>Simpan Susulan</span>
        </button>
      </div>
    </div>
  </div>

  <!-- Modal Input Dinamis Pcs (di luar semua modal lain, sibling dari wrapper utama) -->
  <Teleport to="body">
    <div v-if="showDynamicInputModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm z-[200] flex items-center justify-center p-4" @click.self="closeDynamicModal">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md overflow-hidden">
        <div class="p-4 border-b border-slate-100 flex justify-between items-center bg-slate-50">
          <h3 class="font-bold text-slate-800 flex items-center gap-2">
            <Package class="w-4 h-4 text-indigo-600" />
            Input {{ dynamicModalTitle }}
          </h3>
          <button @click="closeDynamicModal" class="text-slate-400 hover:text-slate-600 transition-colors p-1 rounded-lg hover:bg-slate-100">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div class="bg-indigo-50 p-3 rounded-xl border border-indigo-100">
            <p class="text-xs font-bold text-indigo-900">{{ dynamicModalItem?.product?.name }}</p>
            <p class="text-[10px] text-indigo-500 font-mono mt-0.5">SKU: {{ dynamicModalItem?.product?.sku }}</p>
          </div>

          <div class="space-y-4">
            <div>
              <label class="text-xs font-bold text-slate-700 block mb-1.5">Isi per Dus/Krat <span class="text-slate-400 font-normal">(pcs)</span></label>
              <input type="number" v-model.number="dynamicModalForm.pcs_per_unit" min="1" class="w-full px-3 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-300 text-sm font-bold transition-shadow" />
            </div>
            
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="text-xs font-bold text-slate-700 block mb-1.5">Jumlah Dus/Krat</label>
                <div class="relative">
                  <input type="number" v-model.number="dynamicModalForm.cartons" min="0" class="w-full px-3 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-300 text-sm font-bold pr-12 transition-shadow" />
                  <span class="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400 uppercase">Dus</span>
                </div>
              </div>
              <div>
                <label class="text-xs font-bold text-slate-700 block mb-1.5">Sisa Pcs</label>
                <div class="relative">
                  <input type="number" v-model.number="dynamicModalForm.pcs" min="0" class="w-full px-3 py-2.5 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-300 text-sm font-bold pr-12 transition-shadow" />
                  <span class="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400 uppercase">Pcs</span>
                </div>
              </div>
            </div>

            <div class="pt-4 border-t border-slate-200 flex justify-between items-center">
              <span class="text-xs font-semibold text-slate-500">Total Keseluruhan:</span>
              <span class="text-2xl font-black text-indigo-600">{{ dynamicTotalPcs }} <span class="text-sm">Pcs</span></span>
            </div>
          </div>
        </div>
        <div class="p-4 bg-slate-50 border-t border-slate-100 flex justify-end gap-2">
          <button @click="closeDynamicModal" class="px-4 py-2 text-xs font-bold text-slate-600 hover:bg-slate-100 rounded-lg transition-colors">Batal</button>
          <button @click="saveDynamicModal" class="px-5 py-2 text-xs font-bold bg-indigo-600 text-white hover:bg-indigo-700 rounded-xl transition-colors shadow-md">✓ Simpan</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
@media print {
  @page {
    size: A4;
    margin: 10mm;
  }
  :deep(header), :deep(nav), :deep(aside) {
    display: none !important;
  }
  body, html {
    background-color: white !important;
    margin: 0 !important;
    padding: 0 !important;
  }
}
</style>
