<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import apiClient from '@/api/axios'
import { warehouseApi } from '@/api/warehouse'
import {
  X,
  Plus,
  Trash2,
  Save,
  Printer,
  ShoppingBag,
  CreditCard,
  Building2,
  Store,
  UserPlus,
  AlertTriangle,
  CheckCircle2,
  AlertCircle,
  Info,
  ArrowRight,
  Boxes,
  CalendarClock,
  Wallet
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import SearchSelect, { type SelectOption } from '@/components/SearchSelect.vue'
import ProductSearchSelect from '@/components/sales/ProductSearchSelect.vue'
import ProductSelectModal from '@/components/sales/ProductSelectModal.vue'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', sale: any, printImmediately: boolean): void
}>()

const masterDataStore = useMasterDataStore()
const factoryStore = useFactoryStore()
const authStore = useAuthStore()

const isSubmitting = ref(false)
const isLoadingData = ref(false)
const isLoadingStock = ref(false)

const branches = computed(() => masterDataStore.branches || [])
const employees = computed(() => masterDataStore.employees || [])
const stores = computed(() => masterDataStore.stores || [])
const products = computed(() => factoryStore.sellableProducts || [])
const bonusInventories = ref<any[]>([])

// Branch stock mapping: product_id -> available quantity
const branchStockMap = ref<Record<string, number>>({})

// Form state
const selectedBranchId = ref('')
const buyerType = ref<'GENERAL_WALKIN' | 'REGISTERED_STORE' | 'EMPLOYEE'>('GENERAL_WALKIN')
const buyerName = ref('')
const buyerPhone = ref('')
const selectedStoreId = ref('')
const selectedBuyerEmployeeId = ref('')
const transactionDate = ref(new Date().toISOString().split('T')[0])
const paymentMethod = ref('CASH')
const tempoTenorDays = ref<number | 'custom'>(14)
const paymentDueDate = ref<string>('')
const downPaymentAmount = ref<number>(0)
const notes = ref('')
const discountAmount = ref(0)
const employeeDiscountPercent = ref(0)

// Calculate due date based on transactionDate and tenor
const calculateDueDate = (days: number): string => {
  const dateStr: string = (transactionDate.value && transactionDate.value.length > 0)
    ? transactionDate.value
    : new Date().toISOString().split('T')[0]!
  const base = new Date(dateStr)
  base.setDate(base.getDate() + days)
  return base.toISOString().split('T')[0]!
}

const setTenor = (days: number | 'custom') => {
  tempoTenorDays.value = days
  if (days !== 'custom') {
    paymentDueDate.value = calculateDueDate(days)
  }
}

watch(transactionDate, () => {
  if (paymentMethod.value === 'TEMPO' && typeof tempoTenorDays.value === 'number') {
    paymentDueDate.value = calculateDueDate(tempoTenorDays.value)
  }
})

watch(paymentMethod, (newMethod) => {
  if (newMethod === 'TEMPO' && !paymentDueDate.value) {
    setTenor(14)
  }
})

// Confirmation Modal State
const showConfirmModal = ref(false)
const pendingPrintImmediately = ref(false)

// Inline New Store Modal State
const showNewStoreModal = ref(false)
const isSavingStore = ref(false)
const newStoreForm = ref({
  name: '',
  owner_name: '',
  phone_number: '',
  address: '',
  district: '',
  regency: '',
  province: '',
  category: 'TOKO_BARU'
})

// Product Catalog Modal State
const showProductCatalogModal = ref(false)
const catalogModalMode = ref<'single' | 'multi'>('multi')
const catalogTargetItemIndex = ref<number>(-1)

const selectedProductIdsInItems = computed(() => {
  return items.value.map(it => it.product_id).filter(Boolean)
})

interface LineItem {
  product_id: string
  quantity: number
  price: number
  discount_percent: number
  unit: string
  stock?: number
}

const items = ref<LineItem[]>([
  { product_id: '', quantity: 1, price: 0, discount_percent: 0, unit: 'Pcs' }
])

interface BonusLineItem {
  bonus_inventory_id: string
  quantity: number
  notes: string
}

const bonusItems = ref<BonusLineItem[]>([])

const formatRupiah = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const getAvailableStock = (productId: string): number => {
  if (!productId) return 0
  return branchStockMap.value[productId] ?? 0
}

const selectedBranchName = computed(() => {
  const b = branches.value.find(branch => branch.id === selectedBranchId.value)
  return b ? b.name : 'Cabang Terpilih'
})

const hasInsufficientStock = computed(() => {
  return items.value.some(it => it.product_id && (Number(it.quantity || 0) > getAvailableStock(it.product_id)))
})

const fetchBranchStock = async (branchId: string) => {
  if (!branchId) {
    branchStockMap.value = {}
    return
  }
  isLoadingStock.value = true
  try {
    const res = await warehouseApi.getInventory(branchId)
    const stockList: any[] = res.data || []
    const map: Record<string, number> = {}
    stockList.forEach((s) => {
      if (s.product_id) {
        map[s.product_id] = (map[s.product_id] || 0) + (Number(s.quantity) || 0)
      }
    })
    branchStockMap.value = map
  } catch (err) {
    console.error('Failed to fetch branch stock:', err)
    branchStockMap.value = {}
  } finally {
    isLoadingStock.value = false
  }
}

const onBranchChange = () => {
  if (selectedBranchId.value) {
    fetchBranchStock(selectedBranchId.value)
  }
}

// Load auxiliary data
const loadInitialData = async () => {
  isLoadingData.value = true
  try {
    // 1. Ensure master data
    if (!masterDataStore.branches.length) await masterDataStore.fetchBranches()
    if (!masterDataStore.employees.length) await masterDataStore.fetchEmployees()
    if (!masterDataStore.stores.length) await masterDataStore.fetchStores()

    // 2. Set default branch (current user's branch or first available)
    if (authStore.user?.branch_id) {
      selectedBranchId.value = authStore.user.branch_id
    } else if (branches.value.length > 0) {
      selectedBranchId.value = branches.value[0].id
    }

    // 3. Fetch Products from Factory Store
    await factoryStore.fetchProducts()

    // 4. Fetch Branch Stock for selected branch
    if (selectedBranchId.value) {
      await fetchBranchStock(selectedBranchId.value)
    }

    // 5. Fetch Bonus Inventories (optional)
    try {
      const bonusRes = await apiClient.get('/bonus-inventory')
      bonusInventories.value = bonusRes.data?.data || []
    } catch (_) {}
  } catch (err: any) {
    console.error('Failed loading initial data for direct sale:', err)
  } finally {
    isLoadingData.value = false
  }
}

watch(buyerType, (newType) => {
  if (newType === 'EMPLOYEE') {
    paymentMethod.value = 'CASH'
    employeeDiscountPercent.value = 10
    applyEmployeeDiscount()
  } else {
    employeeDiscountPercent.value = 0
    items.value.forEach(it => { it.discount_percent = 0 })
    if (paymentMethod.value === 'POTONG_GAJI') {
      paymentMethod.value = 'CASH'
    }
  }
})

const storeOptions = computed<SelectOption[]>(() => {
  return stores.value.map((st: any) => {
    const addressParts = [st.address, st.district, st.regency, st.province].filter(Boolean)
    const fullAddress = addressParts.length > 0 ? addressParts.join(', ') : (st.address || 'Tanpa Alamat')
    const sublabelParts: string[] = []
    if (fullAddress) sublabelParts.push(fullAddress)
    if (st.owner_name) sublabelParts.push(`Pemilik: ${st.owner_name}`)
    if (st.phone_number) sublabelParts.push(`Telp: ${st.phone_number}`)

    return {
      value: st.id,
      label: st.name,
      sublabel: sublabelParts.join(' • '),
      badge: st.category === 'TOKO_BARU' ? 'Pelanggan Baru' : undefined,
      badgeColor: 'bg-emerald-50 text-emerald-700',
      extra: st,
    }
  })
})

const employeeOptions = computed<SelectOption[]>(() => {
  return employees.value.map((emp: any) => ({
    value: emp.id,
    label: `${emp.first_name} ${emp.last_name}`,
    sublabel: [emp.department?.name, emp.job_position?.title, emp.phone_number].filter(Boolean).join(' • '),
    extra: emp,
  }))
})

const onBuyerEmployeeSelect = (empId: string, option?: SelectOption) => {
  selectedBuyerEmployeeId.value = empId
  const emp = option?.extra || employees.value.find(e => e.id === empId)
  if (emp) {
    buyerName.value = `${emp.first_name} ${emp.last_name}`
    buyerPhone.value = emp.phone_number || ''
  } else {
    buyerName.value = ''
    buyerPhone.value = ''
  }
}

const onStoreSelect = (storeId: string, option?: SelectOption) => {
  selectedStoreId.value = storeId
  const st = option?.extra || stores.value.find(s => s.id === storeId)
  if (st) {
    buyerName.value = st.owner_name || st.name || ''
    buyerPhone.value = st.phone_number || ''
  } else {
    buyerName.value = ''
    buyerPhone.value = ''
  }
}

const onItemProductChange = (item: LineItem) => {
  const prod = products.value.find(p => p.id === item.product_id)
  if (prod) {
    item.price = Number(prod.selling_price || prod.factory_price || prod.cost_price || 0)
    item.unit = prod.unit || 'Pcs'
    if (buyerType.value === 'EMPLOYEE' && employeeDiscountPercent.value > 0) {
      item.discount_percent = employeeDiscountPercent.value
    }
  }
}

const applyEmployeeDiscount = () => {
  items.value.forEach(it => {
    if (it.product_id) {
      it.discount_percent = employeeDiscountPercent.value
    }
  })
}

const addItem = () => {
  items.value.push({
    product_id: '',
    quantity: 1,
    price: 0,
    discount_percent: buyerType.value === 'EMPLOYEE' ? employeeDiscountPercent.value : 0,
    unit: 'Pcs'
  })
}

const openCatalogModalForMulti = () => {
  catalogModalMode.value = 'multi'
  catalogTargetItemIndex.value = -1
  showProductCatalogModal.value = true
}

const openCatalogModalForSingle = (index: number) => {
  catalogModalMode.value = 'single'
  catalogTargetItemIndex.value = index
  showProductCatalogModal.value = true
}

const handleProductSelectedSingle = (product: any) => {
  const item = items.value[catalogTargetItemIndex.value]
  if (item && product?.id) {
    item.product_id = product.id
    onItemProductChange(item)
  }
}

const handleProductSelectedMultiple = (selected: { product: any; quantity: number }[]) => {
  if (!selected.length) return

  // If the only item is empty, replace it with the first selected item
  let startIndex = 0
  const firstExisting = items.value[0]
  const firstSelected = selected[0]
  if (items.value.length === 1 && firstExisting && !firstExisting.product_id && firstSelected) {
    firstExisting.product_id = firstSelected.product.id
    firstExisting.quantity = firstSelected.quantity || 1
    onItemProductChange(firstExisting)
    startIndex = 1
  }

  // Append or increment the rest
  for (let i = startIndex; i < selected.length; i++) {
    const itemData = selected[i]
    if (!itemData || !itemData.product) continue

    const existingItem = items.value.find(it => it.product_id === itemData.product.id)
    if (existingItem) {
      existingItem.quantity += (itemData.quantity || 1)
    } else {
      const newItem: LineItem = {
        product_id: itemData.product.id,
        quantity: itemData.quantity || 1,
        price: 0,
        discount_percent: buyerType.value === 'EMPLOYEE' ? employeeDiscountPercent.value : 0,
        unit: itemData.product.unit || 'Pcs'
      }
      onItemProductChange(newItem)
      items.value.push(newItem)
    }
  }
  toast.success(`${selected.length} produk berhasil ditambahkan ke faktur`)
}

const removeItem = (idx: number) => {
  if (items.value.length > 1) {
    items.value.splice(idx, 1)
  }
}

const addBonusItem = () => {
  bonusItems.value.push({
    bonus_inventory_id: '',
    quantity: 1,
    notes: 'Bonus Langsung'
  })
}

const removeBonusItem = (idx: number) => {
  bonusItems.value.splice(idx, 1)
}

const subtotalAmount = computed(() => {
  return items.value.reduce((sum, it) => {
    const raw = Number(it.quantity || 0) * Number(it.price || 0)
    const disc = (raw * Number(it.discount_percent || 0)) / 100
    return sum + (raw - disc)
  }, 0)
})

const finalTotalAmount = computed(() => {
  const sub = subtotalAmount.value
  const disc = Number(discountAmount.value || 0)
  return Math.max(0, sub - disc)
})

const openAddStoreModal = () => {
  newStoreForm.value = {
    name: '',
    owner_name: '',
    phone_number: '',
    address: '',
    district: '',
    regency: '',
    province: '',
    category: 'TOKO_BARU'
  }
  showNewStoreModal.value = true
}

const saveNewStore = async () => {
  if (!newStoreForm.value.name.trim()) {
    toast.error('Nama toko / pelanggan pabrik wajib diisi')
    return
  }

  isSavingStore.value = true
  try {
    const payload = {
      name: newStoreForm.value.name.trim(),
      owner_name: newStoreForm.value.owner_name.trim() || newStoreForm.value.name.trim(),
      phone_number: newStoreForm.value.phone_number.trim(),
      address: newStoreForm.value.address.trim() || 'Pembelian Langsung Pabrik',
      district: newStoreForm.value.district.trim(),
      regency: newStoreForm.value.regency.trim(),
      province: newStoreForm.value.province.trim(),
      category: 'TOKO_BARU',
      branch_id: selectedBranchId.value || undefined,
      company_id: masterDataStore.selectedBranchCompanyId || undefined
    }

    const res = await apiClient.post('/stores', payload)
    const createdStore = res.data?.data || res.data
    toast.success('Pelanggan tetap pabrik berhasil ditambahkan!')
    
    // Refresh stores list
    await masterDataStore.fetchStores()
    
    // Auto select newly created store
    if (createdStore?.id) {
      selectedStoreId.value = createdStore.id
      buyerName.value = createdStore.name
      buyerPhone.value = createdStore.phone_number || ''
    }
    
    showNewStoreModal.value = false
  } catch (err: any) {
    console.error('Failed to create new store:', err)
    toast.error(err.response?.data?.error || 'Gagal menyimpan pelanggan baru')
  } finally {
    isSavingStore.value = false
  }
}

const resetForm = () => {
  buyerType.value = 'GENERAL_WALKIN'
  buyerName.value = ''
  buyerPhone.value = ''
  selectedStoreId.value = ''
  selectedBuyerEmployeeId.value = ''
  notes.value = ''
  discountAmount.value = 0
  paymentMethod.value = 'CASH'
  tempoTenorDays.value = 14
  paymentDueDate.value = ''
  downPaymentAmount.value = 0
  items.value = [{ product_id: '', quantity: 1, price: 0, discount_percent: 0, unit: 'Pcs' }]
  bonusItems.value = []
  showConfirmModal.value = false
}

const openConfirmModal = (printImmediately: boolean) => {
  if (!selectedBranchId.value) {
    toast.error('Silakan pilih Cabang / Pabrik asal')
    return
  }
  if (items.value.some(it => !it.product_id || Number(it.quantity) <= 0)) {
    toast.error('Pastikan semua item barang telah dipilih dan kuantitas minimal 1')
    return
  }
  if (finalTotalAmount.value <= 0) {
    toast.error('Total transaksi harus lebih dari 0')
    return
  }
  if (paymentMethod.value === 'TEMPO') {
    if (!paymentDueDate.value) {
      toast.error('Silakan tentukan tanggal jatuh tempo untuk pembayaran tempo')
      return
    }
    const tDate = transactionDate.value ? new Date(transactionDate.value) : new Date()
    if (new Date(paymentDueDate.value) < tDate) {
      toast.error('Tanggal jatuh tempo tidak boleh sebelum tanggal transaksi')
      return
    }
  }

  pendingPrintImmediately.value = printImmediately
  showConfirmModal.value = true
}

const executeSubmitDirectSale = async () => {
  let finalBuyerName = buyerName.value.trim()
  if (buyerType.value === 'GENERAL_WALKIN' && !finalBuyerName) {
    finalBuyerName = 'Pelanggan Walk-in (Umum)'
  }

  let finalPaymentStatus = 'PAID'
  let finalPaidAmount = finalTotalAmount.value

  if (paymentMethod.value === 'POTONG_GAJI') {
    finalPaymentStatus = 'UNPAID'
    finalPaidAmount = 0
  } else if (paymentMethod.value === 'TEMPO') {
    const dp = Number(downPaymentAmount.value || 0)
    if (dp > 0 && dp < finalTotalAmount.value) {
      finalPaymentStatus = 'PARTIAL'
      finalPaidAmount = dp
    } else if (dp >= finalTotalAmount.value && finalTotalAmount.value > 0) {
      finalPaymentStatus = 'PAID'
      finalPaidAmount = finalTotalAmount.value
    } else {
      finalPaymentStatus = 'UNPAID'
      finalPaidAmount = 0
    }
  }

  if (authStore.user?.role === 'admin_nota') {
    finalPaymentStatus = 'UNPAID'
    finalPaidAmount = 0
  }

  const payload = {
    branch_id: selectedBranchId.value,
    sales_channel: 'DIRECT_FACTORY',
    buyer_type: buyerType.value,
    buyer_name: finalBuyerName,
    buyer_phone: buyerPhone.value.trim(),
    buyer_employee_id: buyerType.value === 'EMPLOYEE' ? selectedBuyerEmployeeId.value || undefined : undefined,
    store_id: buyerType.value === 'REGISTERED_STORE' ? selectedStoreId.value || undefined : undefined,
    transaction_date: transactionDate.value,
    payment_due_date: paymentMethod.value === 'TEMPO' && paymentDueDate.value ? new Date(paymentDueDate.value).toISOString() : undefined,
    payment_method: paymentMethod.value,
    payment_status: finalPaymentStatus,
    paid_amount: finalPaidAmount,
    total_amount: finalTotalAmount.value,
    discount_amount: Number(discountAmount.value || 0),
    notes: notes.value,
    items: items.value.map(it => ({
      product_id: it.product_id,
      quantity: Number(it.quantity),
      price: Number(it.price),
      discount_percent: Number(it.discount_percent || 0),
    })),
    bonus_items: bonusItems.value
      .filter(b => b.bonus_inventory_id)
      .map(b => ({
        bonus_inventory_id: b.bonus_inventory_id,
        quantity: Number(b.quantity),
        notes: b.notes
      }))
  }

  isSubmitting.value = true
  try {
    const res = await apiClient.post('/admin/sales/direct-sales', payload)
    toast.success('Faktur penjualan langsung berhasil dibuat! Stok gudang cabang telah dipotong.')
    const createdSale = res.data?.data
    const shouldPrint = pendingPrintImmediately.value
    showConfirmModal.value = false
    resetForm()
    emit('created', createdSale, shouldPrint)
  } catch (err: any) {
    console.error('Error creating direct sale:', err)
    toast.error(err.response?.data?.error || 'Gagal membuat faktur penjualan langsung')
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  if (props.show) {
    loadInitialData()
  }
})

watch(() => props.show, (newVal) => {
  if (newVal) {
    loadInitialData()
  }
})
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-xs flex items-center justify-center p-0 sm:p-4">
    <div class="relative bg-white rounded-none sm:rounded-2xl shadow-2xl max-w-4xl lg:max-w-5xl w-full flex flex-col h-[100dvh] sm:h-auto sm:max-h-[92vh] overflow-hidden">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 sm:px-6 py-3.5 sm:py-4 border-b border-slate-200 bg-slate-50 shrink-0">
        <div class="flex items-center gap-2.5 sm:gap-3">
          <div class="p-2 sm:p-2.5 bg-primary/10 rounded-xl text-primary font-bold">
            <ShoppingBag class="w-5 h-5 sm:w-6 sm:h-6" />
          </div>
          <div>
            <h3 class="font-bold text-slate-900 text-base sm:text-lg leading-tight">Buat Faktur Penjualan Langsung</h3>
            <p class="text-[11px] sm:text-xs text-slate-500">Penjualan langsung pabrik / gudang cabang</p>
          </div>
        </div>
        <Button variant="outline" size="icon" @click="emit('close')" class="rounded-xl hover:bg-slate-100 h-9 w-9">
          <X class="w-4 h-4 text-slate-500" />
        </Button>
      </div>

      <!-- Form Body -->
      <div class="p-4 sm:p-6 overflow-y-auto space-y-5 sm:space-y-6 text-slate-800 flex-1 min-h-0 custom-scrollbar">
        <!-- 1. Branch & Date -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3.5 sm:gap-4 bg-slate-50 p-3.5 sm:p-4 rounded-xl border border-slate-200">
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="block text-xs font-bold text-slate-700">Pabrik / Cabang Asal (Pemotong Stok) *</label>
              <span v-if="isLoadingStock" class="text-[10px] text-amber-600 font-medium animate-pulse">Memuat stok...</span>
            </div>
            <select
              v-model="selectedBranchId"
              @change="onBranchChange"
              class="w-full bg-white border border-slate-300 rounded-lg px-3 py-2 text-xs font-medium focus:ring-2 focus:ring-primary/20 outline-none"
            >
              <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
            </select>
            <p class="text-[10px] sm:text-[11px] text-slate-500 mt-1">
              * Stok barang akan otomatis dipotong dari gudang cabang ini.
            </p>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1">Tanggal Transaksi *</label>
            <input type="date" v-model="transactionDate" class="w-full bg-white border border-slate-300 rounded-lg px-3 py-2 text-xs font-medium focus:ring-2 focus:ring-primary/20 outline-none" />
          </div>
        </div>

        <!-- 2. Buyer Profile & Type -->
        <div class="space-y-3">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <span class="text-xs font-bold text-slate-700 uppercase tracking-wider">Tipe Pembeli</span>
            <div class="flex overflow-x-auto no-scrollbar rounded-xl bg-slate-100 p-1 border border-slate-200 gap-1">
              <button
                type="button"
                @click="buyerType = 'GENERAL_WALKIN'"
                :class="[
                  'px-3 py-1.5 text-xs font-bold rounded-lg transition-all whitespace-nowrap shrink-0',
                  buyerType === 'GENERAL_WALKIN' ? 'bg-white text-primary shadow-xs' : 'text-slate-600 hover:text-slate-900'
                ]"
              >
                1. Pelanggan Luar / Walk-in
              </button>
              <button
                type="button"
                @click="buyerType = 'REGISTERED_STORE'"
                :class="[
                  'px-3 py-1.5 text-xs font-bold rounded-lg transition-all whitespace-nowrap shrink-0',
                  buyerType === 'REGISTERED_STORE' ? 'bg-white text-primary shadow-xs' : 'text-slate-600 hover:text-slate-900'
                ]"
              >
                2. Pelanggan Tetap Pabrik
              </button>
              <button
                type="button"
                @click="buyerType = 'EMPLOYEE'"
                :class="[
                  'px-3 py-1.5 text-xs font-bold rounded-lg transition-all whitespace-nowrap shrink-0',
                  buyerType === 'EMPLOYEE' ? 'bg-white text-primary shadow-xs' : 'text-slate-600 hover:text-slate-900'
                ]"
              >
                3. Karyawan Internal
              </button>
            </div>
          </div>

          <!-- Dynamic Buyer Form Fields -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3.5 sm:gap-4 bg-slate-50 p-3.5 sm:p-4 rounded-xl border border-slate-200">
            <!-- Walk-in Case -->
            <template v-if="buyerType === 'GENERAL_WALKIN'">
              <div>
                <label class="block text-xs font-bold text-slate-700 mb-1">Nama Pembeli (Walk-in) *</label>
                <input
                  type="text"
                  v-model="buyerName"
                  placeholder="Contoh: Bpk. H. Sukirno / Bu Siti"
                  class="w-full bg-white border border-slate-300 rounded-lg px-3 py-2 text-xs font-medium focus:ring-2 focus:ring-primary/20 outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-bold text-slate-700 mb-1">No. WhatsApp / HP</label>
                <input
                  type="text"
                  v-model="buyerPhone"
                  placeholder="08xxxxxxxxxx"
                  class="w-full bg-white border border-slate-300 rounded-lg px-3 py-2 text-xs font-medium focus:ring-2 focus:ring-primary/20 outline-none"
                />
              </div>
            </template>

            <!-- Store Case -->
            <template v-else-if="buyerType === 'REGISTERED_STORE'">
              <div class="md:col-span-2 flex flex-col sm:flex-row gap-2 items-stretch sm:items-end">
                <div class="flex-1 w-full min-w-0">
                  <div class="flex items-center justify-between mb-1">
                    <label class="text-xs font-bold text-slate-700">Pilih Toko / Pelanggan Pabrik *</label>
                  </div>
                  <SearchSelect
                    v-model="selectedStoreId"
                    :options="storeOptions"
                    placeholder="Pilih Pelanggan Tetap Pabrik..."
                    search-placeholder="Ketik untuk mencari nama toko, pemilik, alamat, no hp..."
                    :clearable="true"
                    size="sm"
                    @change="onStoreSelect"
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  @click="openAddStoreModal"
                  class="h-9 px-3 gap-1 text-xs font-bold text-primary border-primary/30 hover:bg-primary/5 shrink-0 justify-center"
                >
                  <UserPlus class="w-3.5 h-3.5" /> + Pelanggan Baru
                </Button>
              </div>
              <div>
                <label class="block text-xs font-bold text-slate-700 mb-1">Nama Pemilik / PIC</label>
                <input
                  type="text"
                  v-model="buyerName"
                  class="w-full bg-white border border-slate-300 rounded-lg px-3 py-2 text-xs font-medium focus:ring-2 focus:ring-primary/20 outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-bold text-slate-700 mb-1">No. WhatsApp / HP</label>
                <input
                  type="text"
                  v-model="buyerPhone"
                  placeholder="08xxxxxxxxxx"
                  class="w-full bg-white border border-slate-300 rounded-lg px-3 py-2 text-xs font-medium focus:ring-2 focus:ring-primary/20 outline-none"
                />
              </div>
            </template>

            <!-- Employee Case -->
            <template v-else>
              <div>
                <label class="block text-xs font-bold text-slate-700 mb-1">Pilih Karyawan Pembeli *</label>
                <SearchSelect
                  v-model="selectedBuyerEmployeeId"
                  :options="employeeOptions"
                  placeholder="Pilih Karyawan..."
                  search-placeholder="Ketik untuk mencari nama karyawan, jabatan, divisi..."
                  :clearable="true"
                  size="sm"
                  @change="onBuyerEmployeeSelect"
                />
              </div>
              <div>
                <div class="flex items-center justify-between mb-1">
                  <label class="text-xs font-bold text-slate-700">Diskon Karyawan (%)</label>
                  <span class="text-[11px] text-primary font-bold">Otomatis Terpasang</span>
                </div>
                <div class="flex items-center gap-2">
                  <input
                    type="number"
                    min="0"
                    max="100"
                    v-model.number="employeeDiscountPercent"
                    @input="applyEmployeeDiscount"
                    class="w-24 bg-white border border-slate-300 rounded-lg px-3 py-2 text-xs font-bold text-slate-800 focus:ring-2 focus:ring-primary/20 outline-none"
                  />
                  <span class="text-xs text-slate-500 font-medium">% per item barang</span>
                </div>
              </div>
            </template>
          </div>
        </div>

        <!-- 3. Items Selection Cards (Spacious, Anti-Squeeze Layout) -->
        <div class="space-y-3">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div>
              <span class="text-xs font-bold text-slate-700 uppercase tracking-wider block">Item Produk Penjualan</span>
              <span class="text-[11px] text-slate-500">Pilih produk dan tentukan kuantitas yang dijual</span>
            </div>
            <div class="flex items-center gap-2">
              <Button
                size="sm"
                type="button"
                @click="openCatalogModalForMulti"
                class="gap-1.5 text-xs bg-primary hover:bg-primary/90 text-white font-bold shadow-xs"
              >
                <Boxes class="w-3.5 h-3.5" /> + Pilih dari Katalog
              </Button>
              <Button
                size="sm"
                variant="outline"
                type="button"
                @click="addItem"
                class="gap-1.5 text-xs text-slate-700 font-bold border-slate-300 hover:bg-slate-50"
              >
                <Plus class="w-3.5 h-3.5" /> Tambah Baris
              </Button>
            </div>
          </div>

          <!-- Unified Responsive Item Cards -->
          <div class="space-y-3">
            <div
              v-for="(item, idx) in items"
              :key="'item-' + idx"
              class="bg-slate-50/90 hover:bg-slate-50 p-3.5 sm:p-4 rounded-2xl border border-slate-200/90 hover:border-slate-300 transition-all space-y-3 shadow-2xs"
            >
              <!-- Top Row: Product Selector, Catalog Button & Delete -->
              <div class="flex items-start gap-2 sm:gap-3">
                <span class="w-6 h-6 rounded-lg bg-primary/10 text-primary flex items-center justify-center text-[11px] font-black shrink-0 mt-1">
                  {{ idx + 1 }}
                </span>
                <div class="flex-1 min-w-0">
                  <label class="block text-[9px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">Pilih Produk & Stok Cabang *</label>
                  <div class="flex items-center gap-1.5">
                    <div class="flex-1 min-w-0">
                      <ProductSearchSelect
                        v-model="item.product_id"
                        :products="products"
                        :get-stock="getAvailableStock"
                        @open-catalog="openCatalogModalForSingle(idx)"
                        @change="() => onItemProductChange(item)"
                      />
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      @click="openCatalogModalForSingle(idx)"
                      class="h-9 px-2.5 text-primary border-primary/30 hover:bg-primary/5 shrink-0"
                      title="Buka Katalog Produk"
                    >
                      <Boxes class="w-4 h-4" />
                    </Button>
                  </div>
                  <!-- Stock Indicator Badges -->
                  <div v-if="item.product_id" class="mt-1.5 flex flex-wrap items-center gap-1.5 text-[11px]">
                    <span :class="getAvailableStock(item.product_id) > 0 ? 'text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-md border border-emerald-200 font-medium' : 'text-rose-700 bg-rose-50 px-2 py-0.5 rounded-md border border-rose-200 font-bold'">
                      Stok di {{ selectedBranchName }}: {{ getAvailableStock(item.product_id) }} {{ item.unit || 'Pcs' }}
                    </span>
                    <span v-if="Number(item.quantity) > getAvailableStock(item.product_id)" class="text-rose-600 bg-rose-100/80 px-2 py-0.5 rounded-md text-[10px] font-bold">
                      ⚠️ Melebihi stok gudang
                    </span>
                  </div>
                </div>
                <button
                  v-if="items.length > 1"
                  type="button"
                  @click="removeItem(idx)"
                  class="p-2 text-rose-400 hover:text-rose-600 hover:bg-rose-50 rounded-xl transition-colors shrink-0 mt-4 cursor-pointer"
                  title="Hapus Produk"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>

              <!-- Bottom Row: Input Fields & Subtotal Grid -->
              <div class="grid grid-cols-1 sm:grid-cols-12 gap-2.5 sm:gap-3 items-end pt-2.5 border-t border-slate-200/70">
                <div class="sm:col-span-3">
                  <label class="block text-[9px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">Jumlah (Qty) *</label>
                  <div class="flex items-center gap-1.5">
                    <input
                      type="number"
                      min="1"
                      v-model.number="item.quantity"
                      placeholder="1"
                      class="w-full bg-white border border-slate-300 rounded-xl px-3 py-2 text-xs font-black text-center text-slate-900 focus:ring-2 focus:ring-primary/20 focus:border-primary/50 outline-none transition-all shadow-2xs"
                    />
                    <span class="text-xs text-slate-600 font-bold px-1.5 py-1 bg-slate-100 rounded-lg border border-slate-200 shrink-0">{{ item.unit || 'Pcs' }}</span>
                  </div>
                </div>
                <div class="sm:col-span-4">
                  <label class="block text-[9px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">Harga Satuan (Rp)</label>
                  <div class="relative">
                    <span class="absolute left-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">Rp</span>
                    <input
                      type="number"
                      min="0"
                      v-model.number="item.price"
                      placeholder="0"
                      class="w-full bg-white border border-slate-300 rounded-xl pl-8 pr-3 py-2 text-xs font-bold text-right text-slate-900 focus:ring-2 focus:ring-primary/20 focus:border-primary/50 outline-none transition-all shadow-2xs"
                    />
                  </div>
                </div>
                <div class="sm:col-span-2">
                  <label class="block text-[9px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">Diskon (%)</label>
                  <div class="relative">
                    <input
                      type="number"
                      min="0"
                      max="100"
                      v-model.number="item.discount_percent"
                      placeholder="0"
                      class="w-full bg-white border border-slate-300 rounded-xl pl-3 pr-6 py-2 text-xs font-bold text-center text-slate-900 focus:ring-2 focus:ring-primary/20 focus:border-primary/50 outline-none transition-all shadow-2xs"
                    />
                    <span class="absolute right-2.5 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">%</span>
                  </div>
                </div>
                <div class="sm:col-span-3">
                  <label class="block text-[9px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5 text-right sm:text-right">Subtotal</label>
                  <div class="bg-white border border-slate-200 rounded-xl px-3 py-2 flex items-center justify-end h-[38px] shadow-2xs">
                    <span class="text-xs font-black text-primary truncate">
                      {{ formatRupiah((item.quantity * item.price) * (1 - (item.discount_percent || 0)/100)) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 4. Payment & Settlement (Direct Omset Cabang, Approval via Kasir) -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6 bg-slate-50 p-4 sm:p-5 rounded-2xl border border-slate-200">
          <div class="space-y-4">
            <div>
              <label class="block text-xs font-bold text-slate-700 mb-1">Metode Pembayaran *</label>
              <div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
                <button
                  type="button"
                  @click="paymentMethod = 'CASH'"
                  :class="[
                    'px-3 py-2.5 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 transition-all',
                    paymentMethod === 'CASH'
                      ? 'bg-emerald-600 text-white border-emerald-600 shadow-sm'
                      : 'bg-white text-slate-700 border-slate-300 hover:bg-slate-50'
                  ]"
                >
                  <CreditCard class="w-3.5 h-3.5" /> Tunai (Kasir)
                </button>
                <button
                  type="button"
                  @click="paymentMethod = 'TRANSFER'"
                  :class="[
                    'px-3 py-2.5 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 transition-all',
                    paymentMethod === 'TRANSFER'
                      ? 'bg-emerald-600 text-white border-emerald-600 shadow-sm'
                      : 'bg-white text-slate-700 border-slate-300 hover:bg-slate-50'
                  ]"
                >
                  <Wallet class="w-3.5 h-3.5" /> Transfer Bank
                </button>
                <button
                  type="button"
                  @click="paymentMethod = 'TEMPO'"
                  :class="[
                    'px-3 py-2.5 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 transition-all',
                    paymentMethod === 'TEMPO'
                      ? 'bg-amber-600 text-white border-amber-600 shadow-sm'
                      : 'bg-white text-slate-700 border-slate-300 hover:bg-slate-50'
                  ]"
                >
                  <CalendarClock class="w-3.5 h-3.5" /> Kredit / Tempo
                </button>
                <button
                  v-if="buyerType === 'EMPLOYEE'"
                  type="button"
                  @click="paymentMethod = 'POTONG_GAJI'"
                  :class="[
                    'px-3 py-2.5 text-xs font-bold rounded-xl border flex items-center justify-center gap-1.5 transition-all col-span-2 sm:col-span-3',
                    paymentMethod === 'POTONG_GAJI'
                      ? 'bg-indigo-600 text-white border-indigo-600 shadow-sm'
                      : 'bg-white text-slate-700 border-slate-300 hover:bg-slate-50'
                  ]"
                >
                  Potong Gaji (Payroll)
                </button>
              </div>

              <!-- Panel Konfigurasi Kredit / Tempo -->
              <div v-if="paymentMethod === 'TEMPO'" class="mt-3 p-3.5 bg-amber-50/80 border border-amber-200 rounded-xl space-y-3 animate-in fade-in duration-150">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-black text-amber-900 uppercase tracking-wider flex items-center gap-1.5">
                    <CalendarClock class="w-3.5 h-3.5 text-amber-600" /> Pengaturan Jatuh Tempo & Uang Muka
                  </span>
                  <span class="text-[10px] text-amber-700 font-bold bg-amber-100 px-2 py-0.5 rounded-md">
                    Piutang Penjualan
                  </span>
                </div>

                <!-- Tenor Presets -->
                <div>
                  <label class="block text-[10px] font-bold text-amber-900 uppercase tracking-wider mb-1.5">Pilihan Tenor Hari</label>
                  <div class="grid grid-cols-3 sm:grid-cols-6 gap-1.5">
                    <button
                      type="button"
                      @click="setTenor(7)"
                      :class="[
                        'py-1.5 text-xs font-bold rounded-lg border transition-all',
                        tempoTenorDays === 7 ? 'bg-amber-600 text-white border-amber-600 shadow-2xs' : 'bg-white text-slate-700 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      7 Hari
                    </button>
                    <button
                      type="button"
                      @click="setTenor(14)"
                      :class="[
                        'py-1.5 text-xs font-bold rounded-lg border transition-all',
                        tempoTenorDays === 14 ? 'bg-amber-600 text-white border-amber-600 shadow-2xs' : 'bg-white text-slate-700 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      14 Hari
                    </button>
                    <button
                      type="button"
                      @click="setTenor(21)"
                      :class="[
                        'py-1.5 text-xs font-bold rounded-lg border transition-all',
                        tempoTenorDays === 21 ? 'bg-amber-600 text-white border-amber-600 shadow-2xs' : 'bg-white text-slate-700 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      21 Hari
                    </button>
                    <button
                      type="button"
                      @click="setTenor(30)"
                      :class="[
                        'py-1.5 text-xs font-bold rounded-lg border transition-all',
                        tempoTenorDays === 30 ? 'bg-amber-600 text-white border-amber-600 shadow-2xs' : 'bg-white text-slate-700 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      30 Hari
                    </button>
                    <button
                      type="button"
                      @click="setTenor(45)"
                      :class="[
                        'py-1.5 text-xs font-bold rounded-lg border transition-all',
                        tempoTenorDays === 45 ? 'bg-amber-600 text-white border-amber-600 shadow-2xs' : 'bg-white text-slate-700 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      45 Hari
                    </button>
                    <button
                      type="button"
                      @click="setTenor('custom')"
                      :class="[
                        'py-1.5 text-xs font-bold rounded-lg border transition-all',
                        tempoTenorDays === 'custom' ? 'bg-amber-600 text-white border-amber-600 shadow-2xs' : 'bg-white text-slate-700 border-amber-200 hover:bg-amber-100/50'
                      ]"
                    >
                      Kustom
                    </button>
                  </div>
                </div>

                <!-- Due Date & DP Grid -->
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                  <div>
                    <label class="block text-[10px] font-bold text-amber-900 uppercase tracking-wider mb-1">Tanggal Jatuh Tempo *</label>
                    <input
                      type="date"
                      v-model="paymentDueDate"
                      class="w-full bg-white border border-amber-300 rounded-lg px-3 py-1.5 text-xs font-bold text-slate-800 focus:ring-2 focus:ring-amber-400 outline-none shadow-2xs"
                    />
                  </div>
                  <div>
                    <label class="block text-[10px] font-bold text-amber-900 uppercase tracking-wider mb-1">Uang Muka / DP (Opsional)</label>
                    <div class="relative">
                      <span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">Rp</span>
                      <input
                        type="number"
                        min="0"
                        :max="finalTotalAmount"
                        v-model.number="downPaymentAmount"
                        placeholder="0"
                        class="w-full bg-white border border-amber-300 rounded-lg pl-7 pr-3 py-1.5 text-xs font-bold text-right text-slate-800 focus:ring-2 focus:ring-amber-400 outline-none shadow-2xs"
                      />
                    </div>
                  </div>
                </div>

                <!-- Info Sisa Tagihan Tempo -->
                <div class="flex items-center justify-between text-xs pt-1 border-t border-amber-200 text-amber-900">
                  <span class="font-medium">Sisa Piutang / Tagihan:</span>
                  <span class="font-black text-amber-950">
                    {{ formatRupiah(Math.max(0, finalTotalAmount - Number(downPaymentAmount || 0))) }}
                  </span>
                </div>
              </div>

              <p v-if="authStore.user?.role === 'admin_nota'" class="text-[10px] sm:text-[11px] text-slate-500 mt-2">
                Penerimaan diverifikasi melalui Persetujuan Kasir. Admin Kasir mencatat setoran faktur, memilih PT dan rekening bank tujuan, lalu Kasir Pusat memverifikasi melalui Laporan Kasir & Bank → Mutasi Kas & Bank.
              </p>
              <p v-else-if="paymentMethod !== 'TEMPO'" class="text-[10px] sm:text-[11px] text-slate-500 mt-2 italic">
                * Setoran tunai/transfer akan masuk ke antrean <strong>Persetujuan Kasir</strong> untuk diverifikasi kasir.
              </p>
              <p v-else class="text-[10px] sm:text-[11px] text-amber-700 mt-2 italic font-medium">
                * Faktur Tempo akan dicatat sebagai piutang (UNPAID/PARTIAL) dan dapat dilunasi kapan saja melalui penagihan kasir.
              </p>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 mb-1">Catatan Transaksi</label>
              <textarea
                v-model="notes"
                rows="2"
                placeholder="Catatan tambahan..."
                class="w-full bg-white border border-slate-300 rounded-lg px-3 py-2 text-xs font-medium focus:ring-2 focus:ring-primary/20 outline-none"
              ></textarea>
            </div>
          </div>

          <!-- Total Calculation Card -->
          <div class="bg-white p-4 rounded-xl border border-slate-200 flex flex-col justify-between space-y-4">
            <div class="space-y-2 text-xs">
              <div class="flex justify-between py-1 border-b border-slate-100">
                <span class="text-slate-600">Subtotal Penjualan:</span>
                <span class="font-bold text-slate-800">{{ formatRupiah(subtotalAmount) }}</span>
              </div>
              <div class="flex justify-between items-center py-1 border-b border-slate-100">
                <span class="text-slate-600">Potongan Diskon Tambahan:</span>
                <input
                  type="number"
                  min="0"
                  v-model.number="discountAmount"
                  class="w-28 sm:w-32 bg-slate-50 border border-slate-300 rounded px-2 py-1 text-xs text-right font-bold"
                />
              </div>
              <div class="flex justify-between items-center py-2 border-b-2 border-slate-800 text-sm">
                <span class="font-black text-slate-900 uppercase">TOTAL DIBAYAR:</span>
                <span class="font-black text-primary text-base sm:text-lg">{{ formatRupiah(finalTotalAmount) }}</span>
              </div>
            </div>

            <div class="pt-2 flex flex-col gap-2">
              <Button
                type="button"
                :disabled="isSubmitting || finalTotalAmount <= 0"
                @click="openConfirmModal(true)"
                class="w-full h-11 gap-2 bg-primary hover:bg-primary/90 text-white font-bold py-2.5 rounded-xl shadow-sm cursor-pointer justify-center"
              >
                <Printer class="w-4 h-4" /> Simpan & Cetak Faktur
              </Button>
              <Button
                type="button"
                variant="outline"
                :disabled="isSubmitting || finalTotalAmount <= 0"
                @click="openConfirmModal(false)"
                class="w-full h-10 gap-2 font-medium py-2 rounded-xl text-slate-700 hover:bg-slate-100 cursor-pointer justify-center"
              >
                <Save class="w-4 h-4" /> Simpan Saja
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Konfirmasi Penerbitan Faktur & Pemotongan Stok -->
    <div v-if="showConfirmModal" class="fixed inset-0 z-70 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div class="bg-white rounded-2xl shadow-2xl max-w-2xl w-full flex flex-col max-h-[90vh] overflow-hidden border border-slate-200 animate-in fade-in zoom-in-95 duration-150">
        <!-- Header Modal Konfirmasi -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 bg-amber-50/50">
          <div class="flex items-center gap-3">
            <div class="p-2.5 bg-amber-500/10 rounded-xl text-amber-600 font-bold">
              <AlertTriangle class="w-6 h-6" />
            </div>
            <div>
              <h3 class="font-bold text-slate-900 text-base">Konfirmasi Penerbitan Faktur Penjualan</h3>
              <p class="text-xs text-slate-500">Periksa detail pemotongan stok cabang dan nilai transaksi</p>
            </div>
          </div>
          <button @click="showConfirmModal = false" :disabled="isSubmitting" class="text-slate-400 hover:text-slate-600 p-1 rounded-lg">
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Body Modal Konfirmasi -->
        <div class="p-6 overflow-y-auto space-y-4 text-xs text-slate-700">
          <!-- Banner Peringatan Stok -->
          <div class="bg-amber-50 border border-amber-200 p-3.5 rounded-xl flex items-start gap-3">
            <Info class="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
            <div class="space-y-1">
              <p class="font-bold text-amber-900">Perhatian: Stok Gudang Akan Langsung Terpotong</p>
              <p class="text-amber-800 leading-relaxed text-[11px]">
                Dengan menerbitkan faktur ini, sistem akan <strong>memotong stok fisik barang jadi</strong> dari gudang cabang <strong class="text-amber-950 underline">{{ selectedBranchName }}</strong> dan mencatat riwayat mutasi keluar (OUT).
              </p>
            </div>
          </div>

          <!-- Ringkasan Info Transaksi -->
          <div class="grid grid-cols-2 gap-3 bg-slate-50 p-3.5 rounded-xl border border-slate-200">
            <div>
              <span class="text-slate-500 text-[10px] uppercase font-bold tracking-wider block">Cabang Pemotong Stok</span>
              <span class="font-black text-slate-900 text-sm flex items-center gap-1.5 mt-0.5">
                <Building2 class="w-4 h-4 text-primary" /> {{ selectedBranchName }}
              </span>
            </div>
            <div>
              <span class="text-slate-500 text-[10px] uppercase font-bold tracking-wider block">Pelanggan / Pembeli</span>
              <span class="font-bold text-slate-800 text-xs block mt-0.5">
                {{ buyerName.trim() || (buyerType === 'GENERAL_WALKIN' ? 'Pelanggan Walk-in (Umum)' : '-') }}
              </span>
              <span class="text-[10px] text-slate-500">
                {{ buyerType }} • 
                <span :class="paymentMethod === 'TEMPO' ? 'text-amber-700 font-bold' : ''">
                  {{ paymentMethod === 'TEMPO' ? `TEMPO (Jatuh Tempo: ${paymentDueDate || '-'})` : paymentMethod }}
                </span>
              </span>
            </div>
          </div>

          <!-- Tabel Rincian Pemotongan Stok -->
          <div class="space-y-2">
            <div class="flex items-center justify-between font-bold text-slate-800">
              <span class="flex items-center gap-1.5"><Boxes class="w-4 h-4 text-primary" /> Rincian Barang & Perubahan Stok Gudang:</span>
              <span class="text-[11px] text-slate-500 font-normal">Total Item: {{ items.length }}</span>
            </div>

            <div class="border border-slate-200 rounded-xl overflow-hidden shadow-2xs">
              <table class="w-full text-left text-xs border-collapse">
                <thead>
                  <tr class="bg-slate-100 border-b border-slate-200 text-slate-700 font-bold">
                    <th class="p-2.5 w-8 text-center">No</th>
                    <th class="p-2.5">Produk & SKU</th>
                    <th class="p-2.5 w-20 text-center">Qty Potong</th>
                    <th class="p-2.5 w-24 text-center">Stok Awal</th>
                    <th class="p-2.5 w-24 text-center">Sisa Stok</th>
                    <th class="p-2.5 w-28 text-right">Subtotal</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 bg-white">
                  <tr v-for="(item, idx) in items" :key="idx" class="hover:bg-slate-50/50">
                    <td class="p-2.5 text-center text-slate-400 font-medium">{{ idx + 1 }}</td>
                    <td class="p-2.5 font-medium text-slate-800">
                      <div class="font-bold text-slate-900">
                        {{ products.find(p => p.id === item.product_id)?.name || 'Produk belum dipilih' }}
                      </div>
                      <div class="text-[10px] text-slate-400 font-mono">
                        SKU: {{ products.find(p => p.id === item.product_id)?.sku || '-' }}
                      </div>
                    </td>
                    <td class="p-2.5 text-center font-bold text-rose-600 bg-rose-50/50">
                      -{{ item.quantity }} {{ item.unit }}
                    </td>
                    <td class="p-2.5 text-center font-medium">
                      <span :class="getAvailableStock(item.product_id) >= Number(item.quantity) ? 'text-emerald-700 font-semibold' : 'text-amber-600 font-bold'">
                        {{ getAvailableStock(item.product_id) }} {{ item.unit }}
                      </span>
                    </td>
                    <td class="p-2.5 text-center font-bold">
                      <span :class="(getAvailableStock(item.product_id) - Number(item.quantity)) >= 0 ? 'text-emerald-700 font-bold' : 'text-rose-600 bg-rose-50 px-1.5 py-0.5 rounded'">
                        {{ getAvailableStock(item.product_id) - Number(item.quantity) }} {{ item.unit }}
                      </span>
                    </td>
                    <td class="p-2.5 text-right font-bold text-slate-900">
                      {{ formatRupiah((item.quantity * item.price) * (1 - (item.discount_percent || 0)/100)) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Bonus Items (Jika Ada) -->
          <div v-if="bonusItems.filter(b => b.bonus_inventory_id).length > 0" class="space-y-2">
            <div class="flex items-center justify-between font-bold text-slate-800">
              <span class="flex items-center gap-1.5 text-indigo-700"><Boxes class="w-4 h-4" /> Barang Bonus Transaksi (Potong Stok Bonus):</span>
              <span class="text-[11px] text-slate-500 font-normal">{{ bonusItems.filter(b => b.bonus_inventory_id).length }} Item Bonus</span>
            </div>

            <div class="border border-indigo-100 rounded-xl overflow-hidden bg-indigo-50/20">
              <table class="w-full text-left text-xs border-collapse">
                <thead>
                  <tr class="bg-indigo-50 text-indigo-900 font-bold border-b border-indigo-100">
                    <th class="p-2.5 w-8 text-center">No</th>
                    <th class="p-2.5">Item Bonus</th>
                    <th class="p-2.5 w-24 text-center">Qty Potong</th>
                    <th class="p-2.5">Catatan</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-indigo-50 bg-white">
                  <tr v-for="(bItem, bIdx) in bonusItems.filter(b => b.bonus_inventory_id)" :key="bIdx">
                    <td class="p-2.5 text-center text-slate-400 font-medium">{{ bIdx + 1 }}</td>
                    <td class="p-2.5 font-bold text-slate-800">
                      {{ bonusInventories.find(b => b.id === bItem.bonus_inventory_id)?.name || 'Bonus Item' }}
                    </td>
                    <td class="p-2.5 text-center font-bold text-rose-600 bg-rose-50/40">
                      -{{ bItem.quantity }} Pcs
                    </td>
                    <td class="p-2.5 text-slate-500 italic">
                      {{ bItem.notes || '-' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Alert jika stok tidak mencukupi -->
          <div v-if="hasInsufficientStock" class="bg-rose-50 border border-rose-200 p-3 rounded-xl flex items-start gap-2.5 text-rose-800">
            <AlertCircle class="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
            <p class="text-[11px] leading-relaxed">
              <strong>Peringatan Stok Kurang:</strong> Ada produk yang kuantitas fakturnya melebihi stok yang tercatat di gudang cabang ini. Transaksi tetap dapat dilanjutkan namun stok gudang akan bernilai minus.
            </p>
          </div>

          <!-- Summary Nilai Transaksi -->
          <div class="bg-slate-50 p-3 rounded-xl border border-slate-200 flex justify-between items-center text-xs">
            <span class="font-bold text-slate-700">Total Transaksi Faktur:</span>
            <span class="font-black text-primary text-base">{{ formatRupiah(finalTotalAmount) }}</span>
          </div>
        </div>

        <!-- Footer Modal Konfirmasi -->
        <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 bg-slate-50">
          <Button
            type="button"
            variant="outline"
            :disabled="isSubmitting"
            @click="showConfirmModal = false"
            class="text-xs font-semibold px-4 py-2"
          >
            Batal / Periksa Kembali
          </Button>
          <Button
            type="button"
            :disabled="isSubmitting"
            @click="executeSubmitDirectSale"
            class="text-xs font-bold px-5 py-2 bg-primary hover:bg-primary/90 text-white gap-2 shadow-sm"
          >
            <span v-if="isSubmitting" class="animate-spin">⏳</span>
            <span v-else>Ya, Terbitkan & Potong Stok</span>
          </Button>
        </div>
      </div>
    </div>

    <!-- Inline Modal Tambah Pelanggan Baru -->
    <div v-if="showNewStoreModal" class="fixed inset-0 z-60 bg-black/50 flex items-center justify-center p-4">
      <div class="bg-white rounded-2xl shadow-xl max-w-md w-full p-5 space-y-4">
        <div class="flex items-center justify-between border-b pb-3">
          <h4 class="font-bold text-slate-900 text-sm flex items-center gap-2">
            <Store class="w-4 h-4 text-primary" /> Tambah Pelanggan Tetap Pabrik
          </h4>
          <button @click="showNewStoreModal = false" class="text-slate-400 hover:text-slate-600">
            <X class="w-4 h-4" />
          </button>
        </div>

        <div class="space-y-3 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Nama Toko / Pelanggan *</label>
            <input
              type="text"
              v-model="newStoreForm.name"
              placeholder="Contoh: Toko Berkah Abadi / Bpk. Haji Suparman"
              class="w-full bg-slate-50 border border-slate-300 rounded-lg px-3 py-2 font-medium focus:bg-white focus:ring-2 focus:ring-primary/20 outline-none"
            />
          </div>
          <div>
            <label class="block font-bold text-slate-700 mb-1">Nama Pemilik / PIC</label>
            <input
              type="text"
              v-model="newStoreForm.owner_name"
              placeholder="Contoh: Suparman"
              class="w-full bg-slate-50 border border-slate-300 rounded-lg px-3 py-2 font-medium focus:bg-white focus:ring-2 focus:ring-primary/20 outline-none"
            />
          </div>
          <div>
            <label class="block font-bold text-slate-700 mb-1">No. WhatsApp / Telepon</label>
            <input
              type="text"
              v-model="newStoreForm.phone_number"
              placeholder="08xxxxxxxxxx"
              class="w-full bg-slate-50 border border-slate-300 rounded-lg px-3 py-2 font-medium focus:bg-white focus:ring-2 focus:ring-primary/20 outline-none"
            />
          </div>
          <div>
            <label class="block font-bold text-slate-700 mb-1">Alamat Pelanggan</label>
            <textarea
              v-model="newStoreForm.address"
              rows="2"
              placeholder="Alamat lengkap..."
              class="w-full bg-slate-50 border border-slate-300 rounded-lg px-3 py-2 font-medium focus:bg-white focus:ring-2 focus:ring-primary/20 outline-none"
            ></textarea>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t">
          <Button variant="outline" size="sm" @click="showNewStoreModal = false" class="text-xs">
            Batal
          </Button>
          <Button size="sm" :disabled="isSavingStore" @click="saveNewStore" class="text-xs font-bold">
            Simpan Pelanggan
          </Button>
        </div>
      </div>
    </div>

    <!-- Modal Katalog & Pemilih Produk Lengkap -->
    <ProductSelectModal
      :show="showProductCatalogModal"
      :products="products"
      :selected-branch-name="selectedBranchName"
      :get-stock="getAvailableStock"
      :mode="catalogModalMode"
      :current-selected-id="catalogTargetItemIndex >= 0 ? items[catalogTargetItemIndex]?.product_id : ''"
      :already-selected-ids="selectedProductIdsInItems"
      @close="showProductCatalogModal = false"
      @select="handleProductSelectedSingle"
      @select-multiple="handleProductSelectedMultiple"
    />
  </div>
</template>
