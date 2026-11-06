<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, reactive, watch } from 'vue'
import { useDeliveryStore } from '@/stores/delivery'
import { useSalesOrderStore } from '@/stores/salesOrder'
import { useMasterDataStore } from '@/stores/masterData'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import {
  Truck,
  CheckCircle2,
  MapPin,
  Calendar,
  Search,
  RefreshCw,
  Camera,
  X,
  User,
  Printer,
  PackagePlus,
  PackageOpen,
  Trash2,
  Pencil,
  History,
  ClipboardList,
  ShoppingBag,
  Banknote,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardFooter } from '@/components/ui/card/index'
import RefreshDataButton from '@/components/ui/RefreshDataButton.vue'
import DataTable from '@/components/DataTable.vue'
import { useDataRefresh } from '@/composables/useDataRefresh'

const deliveryStore = useDeliveryStore()
const soStore = useSalesOrderStore()
const masterStore = useMasterDataStore()

const activeTab = ref<'queue' | 'active' | 'history'>('queue')
const searchQuery = ref('')
const filterCompany = ref('ALL')
const filterSales = ref('ALL')
const filterStore = ref('ALL')
const filterDate = ref('')
const loading = ref(false)
const hasShownAccessError = ref(false)
const isSubmitting = ref(false)
const showPODModal = ref(false)
const showEditModal = ref(false)
const selectedOrder = ref<any>(null)
const editingBatch = ref<any>(null)
const isEditingBatch = ref(false)
const editingBatchId = ref<string | null>(null)

// Current date state & helpers for prominent date header
const currentTime = ref(new Date())
let timer: ReturnType<typeof setInterval> | null = null

const todayStr = computed(() => {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
})

const currentDayName = computed(() => {
  return currentTime.value.toLocaleDateString('id-ID', { weekday: 'long' })
})
const currentDayShort = computed(() => {
  return currentTime.value.toLocaleDateString('id-ID', { weekday: 'short' })
})
const currentDateNum = computed(() => {
  return currentTime.value.getDate()
})
const currentMonthYear = computed(() => {
  return currentTime.value.toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })
})
const currentFullDateStr = computed(() => {
  return currentTime.value.toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
})
const currentTimeStr = computed(() => {
  return currentTime.value.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
})

const setFilterToday = () => {
  filterDate.value = filterDate.value === todayStr.value ? '' : todayStr.value
}

const resetFilters = () => {
  filterCompany.value = 'ALL'
  filterSales.value = 'ALL'
  filterStore.value = 'ALL'
  filterDate.value = ''
  searchQuery.value = ''
}

const drivers = ref<any[]>([])
const employees = ref<any[]>([])

const filteredDrivers = computed(() => {
  const allList = [...employees.value, ...drivers.value]
  const uniqueEmps = Array.from(new Map(allList.map((e: any) => [e.id, e])).values())
  const result = uniqueEmps.filter((emp: any) => {
    const title = (emp.job_position?.title || emp.job_title || '').toLowerCase()
    return (
      title.includes('driver') ||
      title.includes('sopir') ||
      title.includes('pengemudi') ||
      title.includes('kurir') ||
      title.includes('sales') ||
      title.includes('motoris') ||
      title.includes('delivery') ||
      title.includes('logistik')
    )
  })
  return result.length > 0 ? result : uniqueEmps
})

const filteredHelpers = computed(() => {
  const allList = [...employees.value, ...drivers.value]
  const uniqueEmps = Array.from(new Map(allList.map((e: any) => [e.id, e])).values())
  const result = uniqueEmps.filter((emp: any) => {
    const title = (emp.job_position?.title || emp.job_title || '').toLowerCase()
    return (
      title.includes('helper') ||
      title.includes('kernet') ||
      title.includes('gudang') ||
      title.includes('kru') ||
      title.includes('muat') ||
      title.includes('logistik') ||
      title.includes('operator') ||
      title.includes('sales')
    )
  })
  return result.length > 0 ? result : uniqueEmps
})

const vehicles = ref<any[]>([])
const selectedOrders = ref<string[]>([]) // berisi sales_order_id

const podData = ref({ received_by: '', pod_image_url: '' })
const batchForm = reactive({ driver_id: '', daily_driver_name: '', is_daily_driver: false, helper_id: '', daily_helper_name: '', is_daily_helper: false, vehicle_id: '', delivery_date: '' })
const dailyWorkers = ref<any[]>([])

// Derived unique lists for filters
const salesEmployees = computed(() => {
  return masterStore.employees.filter((emp: any) => {
    // 0. Filter by active status
    const isActive = emp.user?.is_active !== false && emp.employment_status !== 'RESIGNED' && emp.employment_status !== 'TERMINATED'

    // 1. Filter by branch (respect global branch filter in header)
    const branchId = masterStore.selectedBranchId
    const empBranchId = emp.operating_branch_id || emp.branch_id
    const belongsToBranch = !branchId || branchId === 'ALL_BRANCHES' || empBranchId === branchId

    // 2. Filter by job title (sales / motoris / booster / marketing / task order / to / mt / kam)
    const jobTitle = (emp.job_position?.title || '').toLowerCase()
    const isSales = /\b(sales|motoris|booster|marketing|task order|take order|to|mt|kam|modern trade)\b/i.test(jobTitle)

    // 3. Filter by Sales Type
    let matchSalesType = true
    const type = masterStore.selectedSalesType
    if (type === 'sales_mt') {
      matchSalesType = /\b(mt|modern trade|kam|key account)\b/i.test(jobTitle)
    } else if (type === 'sales_to') {
      matchSalesType = /\b(to|take order)\b/i.test(jobTitle)
    } else if (type === 'sales_motoris_booster') {
      matchSalesType = /\b(motoris|booster)\b/i.test(jobTitle)
    }

    return isActive && belongsToBranch && isSales && matchSalesType
  })
})

const uniqueCompanies = computed(() => {
  const map = new Map<string, string>()
  deliveryStore.confirmedOrders.forEach((o) => {
    if (o.company?.id && o.company?.name) {
      map.set(o.company.id, o.company.name)
    } else if (o.company_id) {
      const found = masterStore.companies.find((c: any) => c.id === o.company_id)
      map.set(o.company_id, found ? found.name : o.company_id)
    }
  })
  return Array.from(map.entries()).map(([id, name]) => ({ id, name }))
})

const uniqueSales = computed(() => {
  const map = new Map<string, string>()

  // 1. Masukkan semua salesman aktif dari master data cabang
  salesEmployees.value.forEach((emp: any) => {
    if (emp && emp.id) {
      const name = `${emp.first_name || ''} ${emp.last_name || ''}`.trim()
      const title = emp.job_position?.title ? ` (${emp.job_position.title})` : ''
      map.set(emp.id, `${name}${title}`)
    }
  })

  // 2. Tambahkan jika ada salesman dari antrean pesanan yang belum terdaftar di atas
  deliveryStore.confirmedOrders.forEach((o: any) => {
    if (o.employee?.id && !map.has(o.employee.id)) {
      const name = `${o.employee.first_name || ''} ${o.employee.last_name || ''}`.trim()
      const title = o.employee?.job_position?.title ? ` (${o.employee.job_position.title})` : ''
      map.set(o.employee.id, `${name || 'Unknown'}${title}`)
    }
  })

  return Array.from(map.entries())
    .map(([id, name]) => ({ id, name }))
    .sort((a, b) => a.name.localeCompare(b.name))
})

const uniqueStores = computed(() => {
  const map = new Map<string, string>()
  deliveryStore.confirmedOrders.forEach((o) => {
    if (o.store?.id && o.store?.name) map.set(o.store.id, o.store.name)
  })
  return Array.from(map.entries()).map(([id, name]) => ({ id, name }))
})

// Antrean = SO berstatus CONFIRMED yang belum masuk batch
const filteredQueue = computed(() => {
  return deliveryStore.confirmedOrders.filter((o) => {
    if (o.status !== 'CONFIRMED' || o.delivery_batch_id) return false

    const q = searchQuery.value.toLowerCase()
    const matchSearch =
      !q ||
      o.so_number?.toLowerCase().includes(q) ||
      o.store?.name?.toLowerCase().includes(q) ||
      o.employee?.first_name?.toLowerCase().includes(q)

    const matchCompany =
      filterCompany.value === 'ALL' ||
      o.company_id === filterCompany.value ||
      o.company?.id === filterCompany.value
    const matchSales =
      filterSales.value === 'ALL' ||
      o.employee?.id === filterSales.value ||
      o.employee_id === filterSales.value
    const matchStore =
      filterStore.value === 'ALL' ||
      o.store?.id === filterStore.value ||
      o.store_id === filterStore.value

    const matchDate = !filterDate.value || (() => {
      const orderDateStr = o.order_date || o.created_at
      if (!orderDateStr) return false
      const dateObj = new Date(orderDateStr)
      const y = dateObj.getFullYear()
      const m = String(dateObj.getMonth() + 1).padStart(2, '0')
      const d = String(dateObj.getDate()).padStart(2, '0')
      return `${y}-${m}-${d}` === filterDate.value
    })()

    return matchSearch && matchCompany && matchSales && matchStore && matchDate
  })
})

const filteredActive = computed(() => {
  const branchId = masterStore.selectedBranchId
  return deliveryStore.batches.filter((b) => {
    const isActive = b.status !== 'COMPLETED' && b.status !== 'CANCELLED'
    const q = searchQuery.value.toLowerCase()
    const matchSearch =
      !q ||
      b.delivery_order_no?.toLowerCase().includes(q) ||
      b.driver?.first_name?.toLowerCase().includes(q) ||
      b.driver?.last_name?.toLowerCase().includes(q) ||
      b.daily_driver?.name?.toLowerCase().includes(q) ||
      b.vehicle?.license_plate?.toLowerCase().includes(q)

    const matchBranch =
      !branchId ||
      branchId === 'ALL_BRANCHES' ||
      b.branch_id === branchId ||
      b.items?.some((i: any) => (i.sales_order || i.sales_transaction)?.branch_id === branchId)

    const matchCompany =
      filterCompany.value === 'ALL' ||
      b.company_id === filterCompany.value ||
      b.items?.some((i: any) => (i.sales_order || i.sales_transaction)?.company_id === filterCompany.value)

    return isActive && matchSearch && matchBranch && matchCompany
  })
})

const filteredHistory = computed(() => {
  const branchId = masterStore.selectedBranchId
  return deliveryStore.batches.filter((b) => {
    const isCompleted = b.status === 'COMPLETED'
    const q = searchQuery.value.toLowerCase()
    const matchSearch =
      !q ||
      b.delivery_order_no?.toLowerCase().includes(q) ||
      b.driver?.first_name?.toLowerCase().includes(q) ||
      b.driver?.last_name?.toLowerCase().includes(q) ||
      b.daily_driver?.name?.toLowerCase().includes(q) ||
      b.vehicle?.license_plate?.toLowerCase().includes(q)

    const matchBranch =
      !branchId ||
      branchId === 'ALL_BRANCHES' ||
      b.branch_id === branchId ||
      b.items?.some((i: any) => (i.sales_order || i.sales_transaction)?.branch_id === branchId)

    const matchCompany =
      filterCompany.value === 'ALL' ||
      b.company_id === filterCompany.value ||
      b.items?.some((i: any) => (i.sales_order || i.sales_transaction)?.company_id === filterCompany.value)

    return isCompleted && matchSearch && matchBranch && matchCompany
  })
})

const pendingApprovalCount = computed(() => {
  const branchId = masterStore.selectedBranchId
  return deliveryStore.batches.filter((b) => {
    const matchBranch =
      !branchId ||
      branchId === 'ALL_BRANCHES' ||
      b.branch_id === branchId ||
      b.items?.some((i: any) => (i.sales_order || i.sales_transaction)?.branch_id === branchId)
    return b.status === 'WAITING_APPROVAL' && matchBranch
  }).length
})

const inUseDriverIds = computed(() => {
  return deliveryStore.batches
    .filter(
      (b) =>
        b.status !== 'COMPLETED' &&
        b.status !== 'CANCELLED' &&
        (!isEditingBatch.value ||
          (b.id !== editingBatchId.value && b.id !== editingBatch.value?.id)),
    )
    .map((b) => b.driver_id)
    .filter(Boolean)
})

const inUseHelperIds = computed(() => {
  return deliveryStore.batches
    .filter(
      (b) =>
        b.status !== 'COMPLETED' &&
        b.status !== 'CANCELLED' &&
        (!isEditingBatch.value ||
          (b.id !== editingBatchId.value && b.id !== editingBatch.value?.id)),
    )
    .map((b) => b.helper_id)
    .filter(Boolean)
})

const inUseVehicleIds = computed(() => {
  return deliveryStore.batches
    .filter(
      (b) =>
        b.status !== 'COMPLETED' &&
        b.status !== 'CANCELLED' &&
        (!isEditingBatch.value ||
          (b.id !== editingBatchId.value && b.id !== editingBatch.value?.id)),
    )
    .map((b) => b.vehicle_id)
    .filter(Boolean)
})

const selectedTotal = computed(() => {
  return filteredQueue.value
    .filter((o) => selectedOrders.value.includes(o.id))
    .reduce((sum, o) => sum + (o.total_amount || 0), 0)
})

const selectedTonnage = computed(() => {
  let totalKg = 0
  
  const getOrderById = (id: string): any => {
    return (
      deliveryStore.confirmedOrders.find((o: any) => o.id === id) ||
      editingBatch.value?.items?.find((i: any) => (i.sales_order || i.sales_transaction)?.id === id || i.sales_order_id === id || i.sales_transaction_id === id)?.sales_order ||
      editingBatch.value?.items?.find((i: any) => (i.sales_order || i.sales_transaction)?.id === id || i.sales_order_id === id || i.sales_transaction_id === id)?.sales_transaction
    )
  }
  
  for (const id of selectedOrders.value) {
    const order = getOrderById(id)
    if (order && order.items && Array.isArray(order.items)) {
      for (const item of order.items) {
        if (item.product && item.product.weight) {
          const qty = item.quantity || 0
          const weight = item.product.weight || 0
          const unit = (item.product.weight_unit || 'KG').toUpperCase()
          
          let weightInKg = weight
          if (unit === 'GR' || unit === 'GRAM') {
            weightInKg = weight / 1000
          } else if (unit === 'TON') {
            weightInKg = weight * 1000
          }
          
          totalKg += (qty * weightInKg)
        }
      }
    }
  }
  return totalKg
})

async function fetchData(force = false) {
  if (!masterStore.selectedBranchId) return
  loading.value = true
  try {
    const branchId = masterStore.selectedBranchId
    const companyId = masterStore.selectedBranchCompanyId || ''

    // Gunakan Promise.allSettled agar request yang error (misal 403) tidak menggagalkan yang lain
    const results = await Promise.allSettled([
      masterStore.fetchCompanies(force),
      masterStore.fetchEmployees(force),
      deliveryStore.fetchConfirmedOrders(branchId),
      deliveryStore.fetchBatches(branchId),
      apiClient.get(`/employees?job_title=DRIVER&company_id=${companyId}&limit=all`),
      apiClient.get(`/employees?branch_id=${branchId}&company_id=${companyId}&limit=all`),
      apiClient.get(`/vehicles?branch_id=${branchId}&bypass_company=true`),
      apiClient.get(`/employees?job_title=SALESMAN&company_id=${companyId}&limit=all`),
      apiClient.get(`/daily-workers`),
    ])

    if (results[4].status === 'fulfilled') {
      const driversData = results[4].value.data?.data || []
      let salesData = []
      if (results[7].status === 'fulfilled') {
        salesData = results[7].value.data?.data || []
      }
      drivers.value = [...driversData, ...salesData]
    }
    if (results[5].status === 'fulfilled') employees.value = results[5].value.data?.data || []
    if (results[6].status === 'fulfilled')
      vehicles.value = results[6].value.data?.data || results[6].value.data || []
    if (results[8].status === 'fulfilled')
      dailyWorkers.value = results[8].value.data?.data || results[8].value.data || []

    // Bersihkan selectedOrders dari nota yang sudah tidak ada di antrean (jika bukan sedang edit batch)
    if (!isEditingBatch.value && !showEditModal.value && selectedOrders.value.length > 0) {
      const validQueueIds = new Set(
        deliveryStore.confirmedOrders
          .filter((o: any) => o.status === 'CONFIRMED' && !o.delivery_batch_id)
          .map((o: any) => o.id),
      )
      selectedOrders.value = selectedOrders.value.filter((id) => validQueueIds.has(id))
    }

    const has403 = results.some(
      (r) => r.status === 'rejected' && r.reason?.response?.status === 403,
    )
    if (has403 && !hasShownAccessError.value) {
      toast.error('Akses ditolak (403): Beberapa data gagal dimuat karena keterbatasan hak akses.')
    }
    hasShownAccessError.value = has403
  } catch (error: any) {
    console.error('Error fetching data:', error)
  } finally {
    loading.value = false
  }
}

const toggleOrder = (id: string) => {
  selectedOrders.value = selectedOrders.value.includes(id)
    ? selectedOrders.value.filter((o) => o !== id)
    : [...selectedOrders.value, id]
}

const handleCreateBatch = async () => {
  if (selectedOrders.value.length === 0) {
    toast.error('Pilih minimal 1 nota untuk dibuatkan batch')
    return
  }

  let companyId = masterStore.selectedBranchCompanyId
  let branchId = masterStore.selectedBranchId

  // Helper to look up selected orders in both confirmedOrders and editingBatch
  const getOrderById = (id: string): any => {
    return (
      deliveryStore.confirmedOrders.find((o: any) => o.id === id) ||
      editingBatch.value?.items?.find((i: any) => (i.sales_order || i.sales_transaction)?.id === id || i.sales_order_id === id || i.sales_transaction_id === id)?.sales_order ||
      editingBatch.value?.items?.find((i: any) => (i.sales_order || i.sales_transaction)?.id === id || i.sales_order_id === id || i.sales_transaction_id === id)?.sales_transaction
    )
  }

  const orders = selectedOrders.value.map(getOrderById).filter(Boolean)

  // Validate all selected orders are from the same company and branch
  if (orders.length > 0) {
    const distinctCompanies = [...new Set(orders.map((o: any) => o.company_id))]
    const distinctBranches = [...new Set(orders.map((o: any) => o.branch_id))]

    if (!companyId || companyId === '') {
      companyId = distinctCompanies.length === 1 ? distinctCompanies[0] : null
    }
    if (!branchId || branchId === 'ALL_BRANCHES' || branchId === '') {
      branchId = distinctBranches.length === 1 ? distinctBranches[0] : null
    }
  }

  // Fallback to editingBatch info if still empty
  if ((!companyId || companyId === '') && editingBatch.value?.company_id) {
    companyId = editingBatch.value.company_id
  }
  if ((!branchId || branchId === 'ALL_BRANCHES' || branchId === '') && editingBatch.value?.branch_id) {
    branchId = editingBatch.value.branch_id
  }

  isSubmitting.value = true
  try {
    const payload = {
      company_id: companyId || null,
      branch_id: branchId && branchId !== 'ALL_BRANCHES' ? branchId : null,
      driver_id: batchForm.is_daily_driver ? null : batchForm.driver_id || null,
      daily_driver_name: batchForm.is_daily_driver ? batchForm.daily_driver_name || null : null,
      helper_id: batchForm.is_daily_helper ? null : batchForm.helper_id || null,
      daily_helper_name: batchForm.is_daily_helper ? batchForm.daily_helper_name || null : null,
      vehicle_id: batchForm.vehicle_id || null,
      delivery_date: batchForm.delivery_date ? new Date(batchForm.delivery_date).toISOString() : null,
      sales_order_ids: selectedOrders.value,
    }
    if (isEditingBatch.value) {
      await apiClient.put(`/delivery/batches/${editingBatchId.value}`, payload)
    } else {
      await deliveryStore.createBatch(payload)
    }
    toast.success(
      `Batch berhasil dibuat! Surat Jalan telah diterbitkan untuk ${selectedOrders.value.length} nota.`,
    )
    cancelEdit()
    await fetchData()
    activeTab.value = 'active'
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal membuat batch')
  } finally {
    isSubmitting.value = false
  }
}

const handleSupervisorApprove = async (id: string) => {
  try {
    await apiClient.post(`/delivery/batch/${id}/supervisor-approve`)
    toast.success('Batch disetujui Supervisor! Surat Jalan siap dicetak.')
    fetchData()
  } catch (err: any) {
    toast.error(err.response?.data?.message || err.response?.data?.error || 'Gagal menyetujui batch')
  }
}

const openPODModal = (batch: any) => {
  selectedOrder.value = batch
  podData.value = { received_by: '', pod_image_url: '' }
  showPODModal.value = true
}

const handleConfirmPOD = async () => {
  loading.value = true
  try {
    await apiClient.post(`/delivery/batches/${selectedOrder.value.id}/receive`, podData.value)
    toast.success('POD Berhasil!')
    showPODModal.value = false
    fetchData()
  } catch (e: any) {
    toast.error('Gagal konfirmasi POD')
  } finally {
    loading.value = false
  }
}

const cancelEdit = () => {
  isEditingBatch.value = false
  editingBatchId.value = null
  selectedOrders.value = []
  batchForm.driver_id = ''
  batchForm.helper_id = ''
  batchForm.vehicle_id = ''
  batchForm.delivery_date = ''
}

const handleEditBatch = (batch: any) => {
  editingBatch.value = JSON.parse(JSON.stringify(batch)) // Deep copy
  batchForm.driver_id = batch.driver_id || ''
  batchForm.helper_id = batch.helper_id || ''
  batchForm.vehicle_id = batch.vehicle_id || ''
  batchForm.delivery_date = batch.delivery_date ? batch.delivery_date.split('T')[0] : ''

  // Ensure the orders from this batch are in confirmedOrders so they can be resolved/displayed properly
  batch.items?.forEach((item: any) => {
    const so = item.sales_order || item.sales_transaction
    if (so) {
      const exists = deliveryStore.confirmedOrders.find((o: any) => o.id === so.id)
      if (!exists) {
        deliveryStore.confirmedOrders.unshift(so)
      }
    }
  })

  selectedOrders.value = batch.items.map((i: any) => i.sales_order_id).filter(Boolean)
  showEditModal.value = true
}

const handleUpdateBatch = async () => {
  if (selectedOrders.value.length === 0) {
    toast.error('Batch tidak boleh kosong')
    return
  }

  let companyId: string | null = masterStore.selectedBranchCompanyId
  let branchId: string | null = masterStore.selectedBranchId

  // Helper to look up selected orders
  const getOrderById = (id: string): any => {
    return (
      deliveryStore.confirmedOrders.find((o: any) => o.id === id) ||
      editingBatch.value?.items?.find((i: any) => (i.sales_order || i.sales_transaction)?.id === id || i.sales_order_id === id || i.sales_transaction_id === id)?.sales_order ||
      editingBatch.value?.items?.find((i: any) => (i.sales_order || i.sales_transaction)?.id === id || i.sales_order_id === id || i.sales_transaction_id === id)?.sales_transaction
    )
  }

  const orders = selectedOrders.value.map(getOrderById).filter(Boolean)

  const distinctCompanies = [...new Set(orders.map((o: any) => o.company_id))]
  const distinctBranches = [...new Set(orders.map((o: any) => o.branch_id))]

  if (!companyId || companyId === '') {
    if (orders.length > 0 && distinctCompanies.length === 1) {
      companyId = distinctCompanies[0]
    } else if (editingBatch.value?.company_id) {
      companyId = editingBatch.value.company_id
    } else {
      companyId = null
    }
  }

  if (!branchId || branchId === 'ALL_BRANCHES' || branchId === '') {
    if (orders.length > 0 && distinctBranches.length === 1) {
      branchId = distinctBranches[0]
    } else if (editingBatch.value?.branch_id) {
      branchId = editingBatch.value.branch_id
    } else {
      branchId = null
    }
  }

  isSubmitting.value = true
  try {
    const payload = {
      company_id: companyId || null,
      branch_id: branchId && branchId !== 'ALL_BRANCHES' ? branchId : null,
      driver_id: batchForm.driver_id || null,
      helper_id: batchForm.helper_id || null,
      vehicle_id: batchForm.vehicle_id || null,
      delivery_date: batchForm.delivery_date ? new Date(batchForm.delivery_date).toISOString() : null,
      sales_order_ids: selectedOrders.value,
    }
    await apiClient.put(`/delivery/batches/${editingBatch.value.id}`, payload)
    toast.success('Batch berhasil diperbarui!')
    showEditModal.value = false
    fetchData()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal update batch')
  } finally {
    isSubmitting.value = false
  }
}

const handleDeleteBatch = async (id: string) => {
  if (
    !confirm(
      'Apakah Anda yakin ingin menghapus batch ini? Nota di dalamnya akan kembali ke antrean.',
    )
  )
    return
  try {
    await apiClient.delete(`/delivery/batches/${id}`)
    toast.success('Batch berhasil dihapus')
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menghapus batch')
  }
}

const getBatchStatusColor = (status: string) => {
  const map: Record<string, string> = {
    WAITING_APPROVAL: 'bg-yellow-100 text-yellow-700',
    SUPERVISOR_APPROVED: 'bg-blue-100 text-blue-700',
    PICKING: 'bg-orange-100 text-orange-700',
    ON_DELIVERY: 'bg-indigo-100 text-indigo-700',
    COMPLETED: 'bg-green-100 text-green-700',
  }
  return map[status] || 'bg-slate-100 text-slate-600'
}

const getBatchStatusLabel = (status: string) => {
  const map: Record<string, string> = {
    WAITING_APPROVAL: 'Menunggu Persetujuan',
    SUPERVISOR_APPROVED: 'Surat Jalan Terbit',
    PICKING: 'Driver di Gudang',
    ON_DELIVERY: 'Sedang Dikirim',
    COMPLETED: 'Selesai',
  }
  return map[status] || status
}

const formatCurrency = (val: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
const formatDate = (date: string) =>
  date
    ? new Date(date).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    : '-'

const {
  isRefreshing,
  lastRefreshedTimeStr,
  markRefreshed,
  refresh: triggerRefresh,
} = useDataRefresh(
  async ({ isManual }) => {
    await fetchData(isManual)
  },
  {
    successMessage: 'Data pengiriman berhasil diperbarui',
    canAutoRefresh: () =>
      !showPODModal.value &&
      !showEditModal.value &&
      !isEditingBatch.value &&
      !isSubmitting.value &&
      !loading.value,
  },
)

onMounted(() => {
  timer = setInterval(() => {
    currentTime.value = new Date()
  }, 1000)
  if (masterStore.selectedBranchId) {
    fetchData().then(() => markRefreshed())
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
watch(() => masterStore.selectedBranchId, () => {
  filterSales.value = 'ALL'
  fetchData().then(() => markRefreshed())
})
watch(() => masterStore.selectedSalesType, () => {
  filterSales.value = 'ALL'
  fetchData().then(() => markRefreshed())
})
watch(activeTab, (tab) => {
  if (masterStore.selectedBranchId) {
    if (tab === 'queue') {
      deliveryStore.fetchConfirmedOrders(masterStore.selectedBranchId).then(() => markRefreshed())
    } else if (tab === 'active' || tab === 'history') {
      deliveryStore.fetchBatches(masterStore.selectedBranchId).then(() => markRefreshed())
    }
  }
})
</script>

<template>
  <div class="space-y-8 p-6">
    <!-- Header with Prominent Current Date Widget -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6 bg-white p-6 md:p-8 rounded-[36px] border border-slate-200/80 shadow-xs">
      <div class="space-y-1.5">
        <div class="flex items-center gap-3.5">
          <div class="p-3.5 bg-primary/10 text-primary rounded-2xl shadow-xs">
            <Truck class="w-8 h-8" />
          </div>
          <div>
            <h1 class="text-2xl md:text-3xl font-black text-slate-900 tracking-tight flex items-center flex-wrap gap-2">
              <span>Manajemen Pengiriman</span>
              <span v-if="masterStore.selectedBranchName" class="text-primary text-xl md:text-2xl font-bold">
                — {{ masterStore.selectedBranchName }}
              </span>
            </h1>
            <p class="text-xs md:text-sm text-slate-500 font-bold mt-0.5">
              Buat Batch Pengiriman dari Nota → Pilih Driver & Helper & Armada → Surat Jalan Terbit.
            </p>
          </div>
        </div>
      </div>

      <!-- Widget Tanggal Diperbesar (Enlarged Current Date Card) -->
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex items-center gap-3.5 bg-gradient-to-r from-slate-50 via-primary/5 to-slate-50 border-2 border-primary/20 px-5 py-3 rounded-2xl shadow-xs">
          <div class="w-14 h-14 bg-primary text-white rounded-2xl flex flex-col items-center justify-center font-black shadow-md shadow-primary/20 shrink-0">
            <span class="text-[9px] uppercase tracking-widest opacity-90 leading-none">{{ currentDayShort }}</span>
            <span class="text-2xl leading-none mt-1 font-black">{{ currentDateNum }}</span>
          </div>
          <div>
            <div class="flex items-center gap-2">
              <span class="text-[9px] font-black uppercase tracking-wider bg-primary/20 text-primary px-2 py-0.5 rounded-full">
                HARI INI
              </span>
              <span class="text-[10px] font-mono font-black text-slate-500">
                {{ currentTimeStr }} WIB
              </span>
            </div>
            <h2 class="text-base md:text-lg font-black text-slate-900 mt-0.5 tracking-tight">
              {{ currentFullDateStr }}
            </h2>
          </div>
        </div>

        <RefreshDataButton
          size="lg"
          :loading="isRefreshing || loading || deliveryStore.loading"
          :last-updated="lastRefreshedTimeStr"
          @click="triggerRefresh()"
        />
      </div>
    </div>

    <div class="flex gap-2 bg-slate-100 p-1.5 rounded-[24px] w-fit shadow-inner">
      <button
        v-for="tab in [
          { id: 'queue', label: '1. Buat Batch', icon: PackagePlus },
          { id: 'active', label: '2. Dalam Pengiriman', icon: Truck },
          { id: 'history', label: '3. Riwayat', icon: History },
        ]"
        :key="tab.id"
        @click="activeTab = tab.id as any"
        class="relative flex items-center gap-2 px-6 py-3 rounded-[18px] text-xs font-black transition-all"
        :class="
          activeTab === tab.id
            ? 'bg-white text-primary shadow-md'
            : 'text-slate-400 hover:text-slate-600'
        "
      >
        <component :is="tab.icon" class="w-4 h-4" /> {{ tab.label }}
        <span
          v-if="tab.id === 'active' && pendingApprovalCount > 0"
          class="absolute -top-1.5 -right-1.5 bg-red-500 text-white text-[9px] font-black w-5 h-5 flex items-center justify-center rounded-full shadow-md animate-bounce"
        >
          {{ pendingApprovalCount }}
        </span>
      </button>
    </div>

    <!-- Tab 1: Antrian Nota Terverifikasi & Buat Batch -->
    <div
      v-if="activeTab === 'queue'"
      class="grid grid-cols-1 lg:grid-cols-12 gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500"
    >
      <!-- Daftar SO yang siap dibatch -->
      <div class="lg:col-span-7 space-y-4">
        <div class="flex flex-col gap-4 px-2">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">
                Nota Terverifikasi Admin
              </h3>
              <p class="text-[9px] text-slate-300 mt-0.5">
                Centang nota yang ingin digabungkan dalam satu armada pengiriman
              </p>
            </div>
          </div>

          <!-- Filters -->
          <div
            class="flex flex-wrap items-center gap-2 bg-slate-50 p-2.5 rounded-2xl border border-slate-100"
          >
            <div class="relative min-w-[180px] flex-1">
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400" />
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Cari No. SO / Toko / Sales..."
                class="w-full pl-9 pr-4 py-2 bg-white border border-slate-200 rounded-xl text-[10px] font-bold outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>

            <!-- Filter Tanggal Nota -->
            <div class="flex items-center gap-1.5">
              <div class="relative flex items-center">
                <Calendar class="absolute left-2.5 w-3.5 h-3.5 text-slate-400 pointer-events-none" />
                <input
                  v-model="filterDate"
                  type="date"
                  title="Filter Tanggal Nota Pesanan"
                  class="bg-white border border-slate-200 rounded-xl pl-8 pr-2.5 py-1.5 text-[10px] font-bold text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer"
                />
              </div>
              <button
                type="button"
                @click="setFilterToday"
                class="px-2.5 py-2 text-[10px] font-black rounded-xl border transition-all"
                :class="
                  filterDate === todayStr
                    ? 'bg-primary border-primary text-white shadow-xs'
                    : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-100'
                "
              >
                Hari Ini
              </button>
            </div>

            <select
              v-model="filterCompany"
              class="bg-white border border-slate-200 rounded-xl px-3 py-2 text-[10px] font-bold text-slate-600 outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer"
            >
              <option value="ALL">Semua Perusahaan</option>
              <option v-for="c in uniqueCompanies" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
            <select
              v-model="filterSales"
              class="bg-white border border-slate-200 rounded-xl px-3 py-2 text-[10px] font-bold text-slate-600 outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer"
            >
              <option value="ALL">Semua Salesman</option>
              <option v-for="s in uniqueSales" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
            <select
              v-model="filterStore"
              class="bg-white border border-slate-200 rounded-xl px-3 py-2 text-[10px] font-bold text-slate-600 outline-none focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer max-w-[160px] truncate"
            >
              <option value="ALL">Semua Tempat / Toko</option>
              <option v-for="s in uniqueStores" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>

            <button
              @click="resetFilters"
              class="text-[9px] font-black text-slate-400 hover:text-red-500 px-2 flex items-center gap-1 transition-colors ml-auto"
            >
              <X class="w-3 h-3" /> RESET
            </button>
          </div>
        </div>
        <div
          class="bg-white border border-slate-200 rounded-[32px] overflow-hidden shadow-sm divide-y divide-slate-50 max-h-[600px] overflow-y-auto"
        >
          <div
            v-for="order in filteredQueue"
            :key="order.id"
            @click="toggleOrder(order.id)"
            class="p-5 flex items-center gap-4 hover:bg-slate-50 transition-all cursor-pointer group"
            :class="
              selectedOrders.includes(order.id) ? 'bg-primary/5 border-l-4 border-l-primary' : ''
            "
          >
            <div
              class="w-6 h-6 rounded-lg border-2 flex items-center justify-center transition-all shrink-0"
              :class="
                selectedOrders.includes(order.id)
                  ? 'bg-primary border-primary'
                  : 'border-slate-200 group-hover:border-primary/50'
              "
            >
              <CheckCircle2 v-if="selectedOrders.includes(order.id)" class="w-4 h-4 text-white" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-0.5">
                <p class="text-[10px] font-black text-primary font-mono">{{ order.so_number }}</p>
                <span
                  class="text-[8px] font-bold bg-green-100 text-green-700 px-1.5 py-0.5 rounded-full uppercase"
                  >CONFIRMED</span
                >
              </div>
              <h4 class="text-sm font-black text-slate-900 truncate">{{ order.store?.name }}</h4>
              <p class="text-[9px] text-slate-400 font-bold flex items-center gap-1 mt-0.5">
                <User class="w-3 h-3" /> {{ order.employee?.first_name }}
                {{ order.employee?.last_name }}
                <span class="text-slate-200">•</span>
                <MapPin class="w-3 h-3" /> {{ order.store?.address }}
              </p>
            </div>
            <div class="text-right shrink-0">
              <p class="text-sm font-black text-slate-900">
                {{ formatCurrency(order.total_amount) }}
              </p>
              <p class="text-[9px] text-slate-400">{{ formatDate(order.order_date) }}</p>
            </div>
          </div>
          <div v-if="filteredQueue.length === 0" class="p-16 text-center text-slate-300">
            <PackageOpen class="w-12 h-12 mx-auto mb-4 opacity-10" />
            <p class="text-[10px] font-black uppercase tracking-[0.2em]">
              Belum ada nota yang diverifikasi
            </p>
            <p class="text-[9px] mt-1 text-slate-200">
              Minta Admin Nota untuk memverifikasi pesanan salesman
            </p>
          </div>
        </div>
      </div>

      <!-- Panel Buat Batch -->
      <div class="lg:col-span-5 space-y-4">
        <div class="bg-white border border-slate-200 rounded-[40px] p-8 shadow-sm">
          <h3 class="text-xl font-black text-slate-900 mb-1">Buat Surat Jalan</h3>
          <p class="text-[10px] text-slate-500 mb-6 font-bold">
            Surat Jalan terbit otomatis saat batch dibuat
          </p>

          <!-- Summary Pilihan -->
          <div
            v-if="selectedOrders.length > 0"
            class="bg-slate-50 border border-slate-100 rounded-2xl p-4 mb-6 space-y-1"
          >
            <div class="flex justify-between text-xs">
              <span class="text-slate-500">Nota dipilih</span>
              <span class="font-black text-slate-900">{{ selectedOrders.length }} nota</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-slate-500">Total tagihan</span>
              <span class="font-black text-primary">{{ formatCurrency(selectedTotal) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-slate-500">Estimasi Berat</span>
              <span class="font-black text-amber-600">{{ selectedTonnage > 1000 ? (selectedTonnage / 1000).toFixed(2) + ' Ton' : selectedTonnage.toFixed(2) + ' Kg' }}</span>
            </div>
          </div>

          <div class="space-y-4">
            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Pilih Supir (Opsional)</label
              >
              <select
                v-if="!batchForm.is_daily_driver"
                v-model="batchForm.driver_id"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              >
                <option value="">Belum ada supir yang dipilih</option>
                <option
                  v-for="d in filteredDrivers"
                  :key="d.id"
                  :value="d.id"
                  :disabled="inUseDriverIds.includes(d.id)"
                  :class="{
                    'text-slate-400 bg-slate-100': inUseDriverIds.includes(d.id),
                    'text-slate-900': !inUseDriverIds.includes(d.id),
                  }"
                >
                  {{ d.first_name }} {{ d.last_name }}
                  {{ inUseDriverIds.includes(d.id) ? '(Sedang Bertugas)' : '' }}
                </option>
              </select>

              <div v-if="batchForm.is_daily_driver" class="space-y-4">
                <input
                  list="daily-driver-list"
                  v-model="batchForm.daily_driver_name"
                  placeholder="Ketik atau pilih nama Driver..."
                  class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
                />
                <datalist id="daily-driver-list">
                  <option v-for="w in dailyWorkers.filter(w => w.type === 'Driver')" :key="w.id" :value="w.name"></option>
                </datalist>
              </div>
            </div>
            
            <div class="space-y-2 mt-4">
              <label class="flex items-center gap-2 cursor-pointer text-sm font-bold text-slate-800">
                <input type="checkbox" v-model="batchForm.is_daily_driver" class="w-4 h-4 text-primary rounded" />
                Gunakan Driver Harian (Eksternal)
              </label>
            </div>

            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Pilih Helper / Kernet (Opsional)</label
              >
              <select
                v-if="!batchForm.is_daily_helper"
                v-model="batchForm.helper_id"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              >
                <option value="">Belum ada helper yang dipilih</option>
                <option
                  v-for="e in filteredHelpers"
                  :key="e.id"
                  :value="e.id"
                  :disabled="inUseHelperIds.includes(e.id) || batchForm.driver_id === e.id"
                  :class="{
                    'text-slate-400 bg-slate-100':
                      inUseHelperIds.includes(e.id) || batchForm.driver_id === e.id,
                    'text-slate-900': !(
                      inUseHelperIds.includes(e.id) || batchForm.driver_id === e.id
                    ),
                  }"
                >
                  {{ e.first_name }} {{ e.last_name }} ({{ e.job_position?.title || e.job_title || 'Helper' }})
                  {{ inUseHelperIds.includes(e.id) ? '(Sedang Bertugas)' : '' }}
                </option>
              </select>

              <div v-if="batchForm.is_daily_helper" class="space-y-4">
                <input
                  list="daily-helper-list"
                  v-model="batchForm.daily_helper_name"
                  placeholder="Ketik atau pilih nama Helper..."
                  class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
                />
                <datalist id="daily-helper-list">
                  <option v-for="w in dailyWorkers.filter(w => w.type === 'Helper')" :key="w.id" :value="w.name"></option>
                </datalist>
              </div>
            </div>
            
            <div class="space-y-2 mt-2">
              <label class="flex items-center gap-2 cursor-pointer text-sm font-bold text-slate-800">
                <input type="checkbox" v-model="batchForm.is_daily_helper" class="w-4 h-4 text-primary rounded" />
                Gunakan Helper Harian (Eksternal)
              </label>
            </div>
            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Pilih Armada (Opsional)</label
              >
              <select
                v-model="batchForm.vehicle_id"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              >
                <option value="">Belum ada armada yang dipilih</option>
                <option
                  v-for="v in vehicles"
                  :key="v.id"
                  :value="v.id"
                  :disabled="inUseVehicleIds.includes(v.id)"
                  :class="{
                    'text-slate-400 bg-slate-100': inUseVehicleIds.includes(v.id),
                    'text-slate-900': !inUseVehicleIds.includes(v.id),
                  }"
                >
                  {{ v.license_plate }} {{ v.capacity ? `(Kapasitas: ${v.capacity} Kg)` : '' }}
                  {{ inUseVehicleIds.includes(v.id) ? '(Sedang Digunakan)' : '' }}
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Tanggal Pengiriman (Opsional)</label
              >
              <input
                type="date"
                v-model="batchForm.delivery_date"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              />
            </div>
          </div>

          <Button
            class="w-full mt-6 bg-primary hover:bg-primary/90 text-white rounded-2xl py-8 font-black shadow-xl shadow-primary/20"
            :disabled="selectedOrders.length === 0 || isSubmitting"
            @click="handleCreateBatch"
          >
            <Truck class="w-5 h-5 mr-2" />
            {{
              isSubmitting ? 'MEMBUAT...' : `TERBITKAN SURAT JALAN (${selectedOrders.length} NOTA)`
            }}
          </Button>
        </div>
      </div>
    </div>

    <!-- Tab 2: Batch Aktif -->
    <div
      v-if="activeTab === 'active'"
      class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500"
    >
      <div v-if="filteredActive.length === 0" class="text-center py-20 text-slate-300">
        <Truck class="w-16 h-16 mx-auto mb-4 opacity-10" />
        <p class="font-black text-[10px] uppercase tracking-[0.2em]">Tidak ada pengiriman aktif</p>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div
          v-for="batch in filteredActive"
          :key="batch.id"
          class="bg-white border-2 rounded-[32px] p-6 shadow-sm hover:shadow-xl transition-all relative overflow-hidden group"
          :class="
            batch.status === 'WAITING_APPROVAL'
              ? 'border-yellow-400 ring-4 ring-yellow-400/10'
              : 'border-slate-200'
          "
        >
          <div class="flex items-center justify-between mb-4">
            <span
              class="text-[9px] font-black px-3 py-1 rounded-full uppercase tracking-widest"
              :class="getBatchStatusColor(batch.status)"
            >
              {{ getBatchStatusLabel(batch.status) }}
            </span>
            <div class="text-right">
              <span class="text-[10px] font-black text-slate-300 font-mono block">{{
                batch.delivery_order_no
              }}</span>
              <span v-if="batch.delivery_date" class="text-[9px] font-black text-primary block mt-0.5">
                📅 {{ formatDate(batch.delivery_date) }}
              </span>
            </div>
          </div>
          <div class="flex items-center gap-3 mb-5">
            <div
              class="w-12 h-12 rounded-2xl bg-slate-50 border border-slate-100 flex items-center justify-center shrink-0"
            >
              <User class="w-6 h-6 text-slate-300" />
            </div>
            <div>
              <h4 class="text-base font-black text-slate-900 leading-tight">
                {{ batch.driver?.first_name || batch.daily_driver?.name || '— Belum Ditugaskan —' }}
              </h4>
              <p class="text-[9px] font-black text-slate-400 uppercase">
                {{ batch.vehicle?.license_plate || 'Armada belum ditentukan' }} •
                {{ batch.items?.length || 0 }} Nota
                <span v-if="batch.helper || batch.daily_helper"> • (H: {{ batch.helper?.first_name || batch.daily_helper?.name }})</span>
              </p>
            </div>
          </div>
          <!-- Daftar toko dalam batch -->
          <div class="bg-slate-50 rounded-2xl p-3 mb-5 max-h-32 overflow-y-auto">
            <div v-for="item in batch.items" :key="item.id" class="flex items-center gap-2 py-1">
              <ShoppingBag class="w-3 h-3 text-slate-300 shrink-0" />
              <span class="text-[10px] font-bold text-slate-600 truncate">{{
                item.sales_order?.store?.name || item.sales_transaction?.store?.name || '—'
              }}</span>
              <span
                class="ml-auto text-[8px] font-black px-1.5 py-0.5 rounded-full shrink-0"
                :class="
                  item.status === 'DELIVERED'
                    ? 'bg-green-100 text-green-700'
                    : item.status === 'FAILED'
                      ? 'bg-red-100 text-red-700'
                      : 'bg-slate-100 text-slate-500'
                "
              >
                {{ item.status }}
              </span>
            </div>
          </div>
          <div class="flex gap-2">
            <!-- Edit & Delete Buttons -->
            <template
              v-if="batch.status === 'WAITING_APPROVAL' || batch.status === 'SUPERVISOR_APPROVED'"
            >
              <Button
                variant="outline"
                @click="handleEditBatch(batch)"
                class="bg-white border-slate-200 text-slate-600 hover:bg-slate-50 rounded-2xl py-5 px-4 font-black"
              >
                <Pencil class="w-4 h-4" />
              </Button>
              <Button
                variant="outline"
                @click="handleDeleteBatch(batch.id)"
                class="bg-white border-red-100 text-red-500 hover:bg-red-50 rounded-2xl py-5 px-4 font-black"
              >
                <Trash2 class="w-4 h-4" />
              </Button>
            </template>

            <Button
              v-if="batch.status === 'WAITING_APPROVAL'"
              @click="handleSupervisorApprove(batch.id)"
              class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white rounded-2xl py-5 font-black text-[10px]"
            >
              ✓ Setujui & Terbitkan SJ
            </Button>

            <Button
              v-if="batch.status === 'PICKING' || batch.status === 'ON_DELIVERY'"
              @click="openPODModal(batch)"
              class="flex-1 bg-slate-900 text-white rounded-2xl py-5 font-black text-[10px]"
            >
              <CheckCircle2 class="w-3.5 h-3.5 mr-1" /> Input POD
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab 3: Riwayat -->
    <div
      v-if="activeTab === 'history'"
      class="bg-white border border-slate-200 rounded-[32px] overflow-hidden shadow-sm animate-in fade-in slide-in-from-bottom-4 duration-500"
    >
      <DataTable
        :columns="[
          { key: 'created_at', label: 'Tanggal', type: 'datetime' },
          { key: 'delivery_order_no', label: 'No SJ' },
          { key: 'driver.first_name', label: 'Supir' },
          { key: 'items.length', label: 'Nota' },
          { key: 'status', label: 'Status' },
        ]"
        :data="filteredHistory"
      >
        <template #cell-status>
          <Badge
            class="rounded-lg font-black text-[10px] uppercase px-3 py-1 bg-green-100 text-green-700"
            >SELESAI</Badge
          >
        </template>
      </DataTable>
    </div>

    <!-- POD Modal -->
    <div v-if="showPODModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showPODModal = false"
      />
      <div
        class="bg-white rounded-[40px] w-full max-w-md relative z-10 shadow-2xl overflow-hidden p-8 space-y-6"
      >
        <h3 class="text-2xl font-black text-slate-900">Konfirmasi POD</h3>
        <p class="text-xs text-slate-500 -mt-4">Batch: {{ selectedOrder?.delivery_order_no }}</p>
        <input
          v-model="podData.received_by"
          type="text"
          placeholder="Nama Penerima di Toko?"
          class="w-full bg-slate-50 border-2 border-transparent focus:border-primary rounded-2xl px-5 py-4 font-black outline-none transition-all"
        />
        <Button
          @click="handleConfirmPOD"
          :disabled="loading"
          class="w-full bg-primary text-white rounded-2xl py-7 font-black shadow-xl shadow-primary/20"
        >
          {{ loading ? 'PROSES...' : 'KONFIRMASI TERIMA' }}
        </Button>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm"
        @click="showEditModal = false"
      />
      <div
        class="bg-white rounded-[40px] w-full max-w-2xl relative z-10 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
      >
        <div class="p-8 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
          <div>
            <h3 class="text-2xl font-black text-slate-900">Edit Surat Jalan</h3>
            <p class="text-xs text-slate-500 font-bold uppercase tracking-wider">
              {{ editingBatch?.delivery_order_no }}
            </p>
          </div>
          <Button variant="ghost" size="icon" @click="showEditModal = false" class="rounded-full">
            <X class="w-6 h-6" />
          </Button>
        </div>

        <div class="p-8 overflow-y-auto space-y-6 flex-1">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Supir</label
              >
              <select
                v-model="batchForm.driver_id"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              >
                <option value="">— Pilih Supir —</option>
                 <option
                  v-for="d in filteredDrivers"
                  :key="d.id"
                  :value="d.id"
                  :disabled="inUseDriverIds.includes(d.id)"
                  :class="{
                    'text-slate-400 bg-slate-100': inUseDriverIds.includes(d.id),
                    'text-slate-900': !inUseDriverIds.includes(d.id),
                  }"
                >
                  {{ d.first_name }} {{ d.last_name }}
                  {{ inUseDriverIds.includes(d.id) ? '(Sedang Bertugas)' : '' }}
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Helper</label
              >
              <select
                v-model="batchForm.helper_id"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              >
                <option value="">— Pilih Helper —</option>
                <option
                  v-for="e in filteredHelpers"
                  :key="e.id"
                  :value="e.id"
                  :disabled="inUseHelperIds.includes(e.id) || batchForm.driver_id === e.id"
                  :class="{
                    'text-slate-400 bg-slate-100':
                      inUseHelperIds.includes(e.id) || batchForm.driver_id === e.id,
                    'text-slate-900': !(
                      inUseHelperIds.includes(e.id) || batchForm.driver_id === e.id
                    ),
                  }"
                >
                  {{ e.first_name }} {{ e.last_name }} ({{ e.job_position?.title || e.job_title || 'Helper' }})
                  {{ inUseHelperIds.includes(e.id) ? '(Sedang Bertugas)' : '' }}
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Armada</label
              >
              <select
                v-model="batchForm.vehicle_id"
                class="w-full bg-slate-50 border-2 border-transparent focus:border-primary/20 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all"
              >
                <option value="">— Pilih Armada —</option>
                <option
                  v-for="v in vehicles"
                  :key="v.id"
                  :value="v.id"
                  :disabled="inUseVehicleIds.includes(v.id)"
                  :class="{
                    'text-slate-400 bg-slate-100': inUseVehicleIds.includes(v.id),
                    'text-slate-900': !inUseVehicleIds.includes(v.id),
                  }"
                >
                  {{ v.license_plate }} {{ v.capacity ? `(Kapasitas: ${v.capacity} Kg)` : '' }}
                  {{ inUseVehicleIds.includes(v.id) ? '(Sedang Digunakan)' : '' }}
                </option>
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
          </div>

          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-black text-slate-900 uppercase tracking-wider">
                Daftar Nota dalam Batch
              </h4>
              <span class="text-[10px] font-black bg-primary/10 text-primary px-3 py-1 rounded-full"
                >{{ selectedOrders.length }} NOTA</span
              >
            </div>

            <div class="space-y-2">
              <div
                v-for="orderId in selectedOrders"
                :key="orderId"
                class="flex items-center justify-between p-4 bg-slate-50 rounded-2xl border border-slate-100 group"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="w-8 h-8 rounded-xl bg-white border border-slate-200 flex items-center justify-center font-black text-[10px] text-primary"
                  >
                    {{ selectedOrders.indexOf(orderId) + 1 }}
                  </div>
                  <div>
                    <p class="text-[10px] font-black text-slate-900">
                      {{
                        editingBatch?.items?.find((i: any) => i.sales_order_id === orderId)
                          ?.sales_order?.so_number ||
                        deliveryStore.confirmedOrders.find((o: any) => o.id === orderId)
                          ?.so_number ||
                        'NOTA BARU'
                      }}
                    </p>
                    <p class="text-[9px] text-slate-500 font-bold">
                      {{
                        editingBatch?.items?.find((i: any) => i.sales_order_id === orderId)
                          ?.sales_order?.store?.name ||
                        deliveryStore.confirmedOrders.find((o: any) => o.id === orderId)?.store
                          ?.name ||
                        '—'
                      }}
                    </p>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  @click="selectedOrders = selectedOrders.filter((id) => id !== orderId)"
                  class="text-red-400 hover:text-red-600 hover:bg-red-50 rounded-xl"
                >
                  <Trash2 class="w-4 h-4" />
                </Button>
              </div>
            </div>

            <div class="pt-4 border-t border-slate-100">
              <label
                class="text-[9px] font-black text-slate-400 uppercase tracking-widest ml-1 mb-2 block"
                >Tambah Nota dari Antrean</label
              >
              <div class="max-h-40 overflow-y-auto space-y-2 pr-2">
                <div
                  v-for="q in deliveryStore.confirmedOrders.filter(
                    (o: any) => !selectedOrders.includes(o.id),
                  )"
                  :key="q.id"
                  @click="selectedOrders.push(q.id)"
                  class="p-3 bg-white border border-slate-200 rounded-xl flex items-center justify-between cursor-pointer hover:border-primary/50 transition-all"
                >
                  <div>
                    <p class="text-[9px] font-black text-primary">{{ q.so_number }}</p>
                    <p class="text-[10px] font-black text-slate-900">{{ q.store?.name }}</p>
                  </div>
                  <PackagePlus class="w-4 h-4 text-slate-300" />
                </div>
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
            :disabled="isSubmitting || selectedOrders.length === 0"
            class="flex-[2] bg-primary text-white rounded-2xl py-6 font-black shadow-xl shadow-primary/20"
          >
            {{ isSubmitting ? 'MENYIMPAN...' : 'SIMPAN PERUBAHAN' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
