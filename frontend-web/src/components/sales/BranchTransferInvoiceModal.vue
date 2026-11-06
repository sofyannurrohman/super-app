<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import apiClient from '@/api/axios'
import {
  X,
  Plus,
  Trash2,
  Save,
  Printer,
  Truck,
  Building2,
  Calendar,
  CreditCard,
  AlertTriangle,
  CheckCircle2,
  AlertCircle,
  Info,
  Package,
  ArrowRight,
  ClipboardList,
  Layers,
  Sparkles,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'

const props = defineProps<{
  show: boolean
  editTransfer?: any
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', transfer: any): void
  (e: 'updated', transfer: any): void
}>()

const masterDataStore = useMasterDataStore()
const factoryStore = useFactoryStore()
const authStore = useAuthStore()

const inputMode = ref<'from_request' | 'manual'>('from_request')
const isSubmitting = ref(false)
const isLoadingRequests = ref(false)
const isEditMode = computed(() => !!props.editTransfer)

const factories = computed(() => factoryStore.factories || [])
const branches = computed(() => masterDataStore.branches || [])
const employees = computed(() => masterDataStore.employees || [])
const vehicles = ref<any[]>([])
const products = computed(() => factoryStore.sellableProducts || [])

const fetchVehicles = async () => {
  try {
    const res = await apiClient.get('/vehicles')
    vehicles.value = res.data?.data || res.data || []
  } catch {
    vehicles.value = []
  }
}

// Pending requests from branches
const pendingRequests = ref<any[]>([])
const selectedRequestId = ref<string>('')

interface TransferItemLine {
  id?: string
  product_id: string
  product_name: string
  sku: string
  batch_no: string
  quantity: number
  unit: string
  pcs_per_unit: number
  price: number
  max_quantity?: number
}

// Form state
const selectedFactoryId = ref('')
const selectedBranchId = ref('')
const selectedVehicleId = ref('')
const selectedDriverId = ref('')
const shipmentDate = ref(new Date().toISOString().split('T')[0])
const estimatedArrival = ref(new Date().toISOString().split('T')[0])
const paymentMethod = ref<'NET' | 'TEMPO'>('NET')
const paymentDueAt = ref('')
const notes = ref('')

const items = ref<TransferItemLine[]>([])

const formatRupiah = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const fetchPendingRequests = async () => {
  isLoadingRequests.value = true
  try {
    const res = await apiClient.get('/factory/stock/transfers')
    const allTransfers: any[] = res.data || []
    
    // Group transfers by DeliveryOrderNo or ID where status is REQUESTED and initiated_by is BRANCH
    const reqs = allTransfers.filter(
      (t) => t.status === 'REQUESTED' && (t.initiated_by === 'BRANCH' || !t.initiated_by)
    )

    // Grouping by DeliveryOrderNo or ID
    const grouped: Record<string, any> = {}
    for (const item of reqs) {
      const key = item.delivery_order_no || item.id
      if (!grouped[key]) {
        grouped[key] = {
          id: item.id,
          delivery_order_no: item.delivery_order_no || `REQ-${item.id.substring(0, 8).toUpperCase()}`,
          to_branch_id: item.to_branch_id,
          to_branch: item.to_branch,
          from_factory_id: item.from_factory_id,
          from_factory: item.from_factory,
          created_at: item.created_at,
          target_shipment_date: item.target_shipment_date,
          notes: item.notes,
          items: [],
        }
      }
      grouped[key].items.push(item)
    }

    pendingRequests.value = Object.values(grouped)
  } catch (error: any) {
    console.error('Error fetching pending requests:', error)
  } finally {
    isLoadingRequests.value = false
  }
}

const handleSelectRequest = (requestId: string) => {
  selectedRequestId.value = requestId
  const req = pendingRequests.value.find((r) => r.id === requestId || r.delivery_order_no === requestId)
  if (!req) return

  selectedFactoryId.value = req.from_factory_id || (factories.value[0]?.id || '')
  selectedBranchId.value = req.to_branch_id || ''
  notes.value = req.notes ? `[Faktur Request Cabang] ${req.notes}` : 'Faktur Pengiriman atas Request Cabang'

  items.value = req.items.map((it: any) => {
    const prod = products.value.find((p) => p.id === it.product_id) || it.product
    const defaultPrice = prod?.factory_price || prod?.base_price || 0
    return {
      product_id: it.product_id,
      product_name: prod?.name || 'Produk',
      sku: prod?.sku || '-',
      batch_no: it.batch_no || 'DEFAULT',
      quantity: it.quantity || 1,
      unit: it.unit || 'PCS',
      pcs_per_unit: it.pcs_per_unit || 1,
      price: defaultPrice,
      max_quantity: it.quantity || 1,
    }
  })
}

const handleAddItem = () => {
  const defaultProd: any = products.value[0]
  items.value.push({
    product_id: defaultProd ? defaultProd.id : '',
    product_name: defaultProd ? defaultProd.name : '',
    sku: defaultProd ? defaultProd.sku : '',
    batch_no: 'DEFAULT',
    quantity: 1,
    unit: 'PCS',
    pcs_per_unit: 1,
    price: defaultProd ? (defaultProd.factory_price || defaultProd.base_price || 0) : 0,
  })
}

const handleProductChange = (index: number, productId: string) => {
  const prod: any = products.value.find((p) => p.id === productId)
  if (prod && items.value[index]) {
    items.value[index].product_id = prod.id
    items.value[index].product_name = prod.name
    items.value[index].sku = prod.sku || '-'
    items.value[index].price = prod.factory_price || prod.base_price || 0
    items.value[index].unit = prod.unit || 'PCS'
  }
}

const handleRemoveItem = (index: number) => {
  items.value.splice(index, 1)
}

const totalInvoiceAmount = computed(() => {
  return items.value.reduce((sum, it) => sum + (Number(it.quantity) || 0) * (Number(it.price) || 0), 0)
})

const isFormValid = computed(() => {
  if (!selectedFactoryId.value || !selectedBranchId.value) return false
  if (items.value.length === 0) return false
  return items.value.every((it) => it.product_id && it.quantity > 0)
})

const handleSubmit = async () => {
  if (!isFormValid.value) {
    toast.error('Mohon lengkapi data cabang dan daftar barang yang akan difakturkan.')
    return
  }

  isSubmitting.value = true
  try {
    const payloadItems = items.value.map((it) => ({
      id: it.id || undefined,
      product_id: it.product_id,
      batch_no: it.batch_no || 'DEFAULT',
      quantity: Number(it.quantity),
      unit: it.unit || 'PCS',
      pcs_per_unit: Number(it.pcs_per_unit) || 1,
    }))

    const payload = {
      from_factory_id: selectedFactoryId.value,
      to_branch_id: selectedBranchId.value,
      items: payloadItems,
      notes: notes.value,
      target_shipment_date: shipmentDate.value ? `${shipmentDate.value}T08:00` : undefined,
      estimated_arrival: estimatedArrival.value ? `${estimatedArrival.value}T17:00` : undefined,
      initiated_by: 'FACTORY',
      payment_method: paymentMethod.value,
      payment_due_at: paymentMethod.value === 'TEMPO' && paymentDueAt.value ? paymentDueAt.value : undefined,
      vehicle_id: selectedVehicleId.value || undefined,
      driver_id: selectedDriverId.value || undefined,
      transfer_type: 'COMMERCIAL',
      is_billable: true,
    }

    if (isEditMode.value) {
      const doNo = props.editTransfer.delivery_order_no || props.editTransfer.id
      const res = await apiClient.put(`/factory/transfer/do/${encodeURIComponent(doNo)}`, payload)
      toast.success('Faktur Pengiriman Cabang berhasil diperbarui!')
      emit('updated', res.data)
      emit('close')
    } else {
      const res = await apiClient.post(`/factory/${selectedFactoryId.value}/transfer/request`, payload)
      toast.success('Faktur Pengiriman Cabang berhasil diterbitkan dan dicatat ke Piutang Cabang!')
      emit('created', res.data)
      emit('close')
    }
  } catch (error: any) {
    const msg = error.response?.data?.error || error.message || 'Gagal menyimpan faktur cabang'
    toast.error(msg)
  } finally {
    isSubmitting.value = false
  }
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      fetchVehicles()

      if (props.editTransfer) {
        // EDIT MODE: Populate from editTransfer
        inputMode.value = 'manual'
        const et = props.editTransfer
        selectedFactoryId.value = et.from_factory_id || et.from_factory?.id || (factories.value[0]?.id || '')
        selectedBranchId.value = et.to_branch_id || et.to_branch?.id || ''
        selectedVehicleId.value = et.vehicle_id || et.vehicle?.id || ''
        selectedDriverId.value = et.driver_id || et.driver?.id || ''
        shipmentDate.value = et.target_shipment_date?.split('T')[0] || et.shipped_at?.split('T')[0] || et.created_at?.split('T')[0] || new Date().toISOString().split('T')[0]
        estimatedArrival.value = et.estimated_arrival?.split('T')[0] || ''
        paymentMethod.value = et.payment_method === 'TEMPO' ? 'TEMPO' : 'NET'
        paymentDueAt.value = et.payment_due_at?.split('T')[0] || ''
        notes.value = et.notes || ''

        if (et.items && et.items.length > 0) {
          items.value = et.items.map((it: any) => {
            const prod = products.value.find((p) => p.id === it.product_id) || it.product
            return {
              id: it.id,
              product_id: it.product_id,
              product_name: prod?.name || 'Produk',
              sku: prod?.sku || '-',
              batch_no: it.batch_no || 'DEFAULT',
              quantity: it.quantity || 1,
              unit: it.unit || 'PCS',
              pcs_per_unit: it.pcs_per_unit || 1,
              price: prod?.factory_price || prod?.base_price || 0,
            }
          })
        } else {
          const prod = products.value.find((p) => p.id === et.product_id) || et.product
          items.value = [{
            id: et.id,
            product_id: et.product_id,
            product_name: prod?.name || 'Produk',
            sku: prod?.sku || '-',
            batch_no: et.batch_no || 'DEFAULT',
            quantity: et.quantity || 1,
            unit: et.unit || 'PCS',
            pcs_per_unit: et.pcs_per_unit || 1,
            price: prod?.factory_price || prod?.base_price || 0,
          }]
        }
      } else {
        // CREATE MODE
        inputMode.value = 'from_request'
        selectedRequestId.value = ''
        selectedBranchId.value = ''
        selectedVehicleId.value = ''
        selectedDriverId.value = ''
        shipmentDate.value = new Date().toISOString().split('T')[0]
        estimatedArrival.value = new Date().toISOString().split('T')[0]
        paymentMethod.value = 'NET'
        paymentDueAt.value = ''
        notes.value = ''
        items.value = []

        if (factories.value.length > 0 && !selectedFactoryId.value) {
          selectedFactoryId.value = factories.value[0]?.id || ''
        }
        fetchPendingRequests()
        if (items.value.length === 0) {
          handleAddItem()
        }
      }
    }
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4 overflow-y-auto">
    <div class="relative w-full max-w-4xl bg-white rounded-2xl shadow-2xl border border-slate-100 overflow-hidden my-8 flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="px-6 py-5 bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 text-white flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-indigo-500/20 border border-indigo-400/30 rounded-xl">
            <Truck class="w-6 h-6 text-indigo-300" />
          </div>
          <div>
            <h2 class="text-xl font-black tracking-tight flex items-center gap-2">
              {{ isEditMode ? 'Edit Faktur Pengiriman Cabang' : 'Buat Faktur Pengiriman Cabang' }}
              <span class="text-xs bg-indigo-500/30 text-indigo-200 px-2 py-0.5 rounded-full font-semibold border border-indigo-400/20">
                {{ isEditMode ? editTransfer?.delivery_order_no : 'Pusat / Pabrik ke Cabang' }}
              </span>
            </h2>
            <p class="text-xs text-slate-300 mt-0.5">
              {{ isEditMode ? 'Perbarui data rincian pengiriman barang, armada, atau daftar produk faktur cabang.' : 'Terbitkan faktur pengiriman resmi atas request stok cabang. Nilai faktur otomatis menambah saldo Piutang Cabang (Tagihan Pabrik).' }}
            </p>
          </div>
        </div>
        <button
          @click="emit('close')"
          class="p-2 text-slate-400 hover:text-white hover:bg-white/10 rounded-xl transition-all"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Mode Selector (Only in Create Mode) -->
      <div v-if="!isEditMode" class="px-6 pt-4 pb-2 bg-slate-50 border-b border-slate-200">
        <div class="grid grid-cols-2 gap-3 max-w-md">
          <button
            type="button"
            @click="inputMode = 'from_request'"
            :class="[
              'flex items-center justify-center gap-2 py-2 px-3 text-xs font-bold rounded-xl border transition-all',
              inputMode === 'from_request'
                ? 'bg-indigo-600 text-white border-indigo-600 shadow-sm'
                : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100',
            ]"
          >
            <ClipboardList class="w-4 h-4" />
            Tarik dari Request Cabang
            <span
              v-if="pendingRequests.length > 0"
              class="ml-1 px-1.5 py-0.2 bg-amber-400 text-slate-900 rounded-full text-[10px] font-black"
            >
              {{ pendingRequests.length }}
            </span>
          </button>
          <button
            type="button"
            @click="inputMode = 'manual'"
            :class="[
              'flex items-center justify-center gap-2 py-2 px-3 text-xs font-bold rounded-xl border transition-all',
              inputMode === 'manual'
                ? 'bg-indigo-600 text-white border-indigo-600 shadow-sm'
                : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100',
            ]"
          >
            <Layers class="w-4 h-4" />
            Input Manual Pengiriman
          </button>
        </div>
      </div>

      <!-- Body Form -->
      <div class="p-6 space-y-6 overflow-y-auto flex-1">
        <!-- Tarik dari Request Section -->
        <div v-if="inputMode === 'from_request'" class="bg-indigo-50/70 border border-indigo-100 rounded-2xl p-4 space-y-3">
          <div class="flex items-center justify-between">
            <label class="text-xs font-black uppercase text-indigo-900 tracking-wider flex items-center gap-1.5">
              <Sparkles class="w-4 h-4 text-indigo-600" />
              Pilih Permintaan Barang Cabang (Request Approved)
            </label>
            <span v-if="isLoadingRequests" class="text-xs text-indigo-600 font-medium">Memuat data request...</span>
          </div>

          <div v-if="pendingRequests.length === 0 && !isLoadingRequests" class="text-xs text-slate-500 py-2">
            Tidak ada request barang cabang berstatus pending. Anda dapat beralih ke mode <b>Input Manual Pengiriman</b>.
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div
              v-for="req in pendingRequests"
              :key="req.id"
              @click="handleSelectRequest(req.id)"
              :class="[
                'p-3 rounded-xl border cursor-pointer transition-all flex flex-col justify-between text-left',
                selectedRequestId === req.id
                  ? 'bg-indigo-600 text-white border-indigo-600 shadow-md ring-2 ring-indigo-400'
                  : 'bg-white hover:border-indigo-300 text-slate-800 border-slate-200',
              ]"
            >
              <div>
                <div class="flex items-center justify-between text-xs font-bold">
                  <span>{{ req.delivery_order_no }}</span>
                  <span :class="selectedRequestId === req.id ? 'text-indigo-200' : 'text-indigo-600'">
                    {{ req.to_branch?.name || 'Cabang' }}
                  </span>
                </div>
                <div class="text-[11px] mt-1 opacity-80">
                  {{ req.items?.length || 0 }} item produk diminta
                </div>
              </div>
              <div class="text-[10px] mt-2 pt-2 border-t border-slate-100/30 opacity-70 flex justify-between">
                <span>{{ req.created_at?.split('T')[0] }}</span>
                <span class="font-bold">Pilih Request &rarr;</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Master Destination & Warehouse info -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Pabrik Asal Pengirim *
            </label>
            <select
              v-model="selectedFactoryId"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-semibold text-slate-900 focus:ring-2 focus:ring-indigo-500 focus:bg-white"
            >
              <option v-for="f in factories" :key="f.id" :value="f.id">{{ f.name }}</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Cabang Tujuan (Debitur) *
            </label>
            <select
              v-model="selectedBranchId"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-semibold text-slate-900 focus:ring-2 focus:ring-indigo-500 focus:bg-white"
            >
              <option value="" disabled>-- Pilih Cabang --</option>
              <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Metode Pembayaran Tagihan
            </label>
            <select
              v-model="paymentMethod"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-semibold text-slate-900 focus:ring-2 focus:ring-indigo-500 focus:bg-white"
            >
              <option value="NET">NET (Setoran Omzet Normal)</option>
              <option value="TEMPO">TEMPO (Jatuh Tempo Berjangka)</option>
            </select>
          </div>
        </div>

        <!-- Armada, Driver & Tanggal -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Kendaraan / Truk
            </label>
            <select
              v-model="selectedVehicleId"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-medium text-slate-900 focus:ring-2 focus:ring-indigo-500 focus:bg-white"
            >
              <option value="">-- Tanpa Kendaraan Khusus --</option>
              <option v-for="v in vehicles" :key="v.id" :value="v.id">
                {{ v.name }} ({{ v.plate_number }})
              </option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Driver / Pengemudi
            </label>
            <select
              v-model="selectedDriverId"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-medium text-slate-900 focus:ring-2 focus:ring-indigo-500 focus:bg-white"
            >
              <option value="">-- Pilih Driver --</option>
              <option v-for="emp in employees" :key="emp.id" :value="emp.id">
                {{ emp.first_name }} {{ emp.last_name || '' }}
              </option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Tgl Pengiriman
            </label>
            <input
              type="date"
              v-model="shipmentDate"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-medium text-slate-900 focus:ring-2 focus:ring-indigo-500 focus:bg-white"
            />
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Estimasi Tiba Cabang
            </label>
            <input
              type="date"
              v-model="estimatedArrival"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-medium text-slate-900 focus:ring-2 focus:ring-indigo-500 focus:bg-white"
            />
          </div>
        </div>

        <!-- Tabel Item Barang -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-black text-slate-900 uppercase tracking-wider flex items-center gap-2">
              <Package class="w-4 h-4 text-indigo-600" />
              Daftar Barang yang Dikirim & Difakturkan
            </h3>
            <Button
              type="button"
              variant="outline"
              size="sm"
              @click="handleAddItem"
              class="text-xs font-bold border-indigo-200 text-indigo-700 hover:bg-indigo-50 rounded-xl"
            >
              <Plus class="w-3.5 h-3.5 mr-1" /> Tambah Baris Produk
            </Button>
          </div>

          <div class="border border-slate-200 rounded-2xl overflow-hidden shadow-sm">
            <table class="w-full text-left text-xs border-collapse">
              <thead class="bg-slate-100 text-slate-700 uppercase font-black text-[11px] border-b border-slate-200">
                <tr>
                  <th class="p-3">Produk</th>
                  <th class="p-3 w-32">Batch No</th>
                  <th class="p-3 w-28 text-center">Kuantitas</th>
                  <th class="p-3 w-24">Satuan</th>
                  <th class="p-3 w-36 text-right">Harga Pabrik (Rp)</th>
                  <th class="p-3 w-36 text-right">Subtotal</th>
                  <th class="p-3 w-12 text-center"></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="(item, idx) in items" :key="idx" class="hover:bg-slate-50/80 transition-colors">
                  <td class="p-3">
                    <select
                      :value="item.product_id"
                      @change="(e: any) => handleProductChange(idx, e.target.value)"
                      class="w-full bg-white border border-slate-200 rounded-lg px-2.5 py-1.5 text-xs font-bold text-slate-800 focus:ring-2 focus:ring-indigo-500"
                    >
                      <option v-for="p in products" :key="p.id" :value="p.id">
                        {{ p.name }} ({{ p.sku || 'No SKU' }})
                      </option>
                    </select>
                  </td>
                  <td class="p-3">
                    <input
                      type="text"
                      v-model="item.batch_no"
                      placeholder="DEFAULT"
                      class="w-full bg-white border border-slate-200 rounded-lg px-2 py-1.5 text-xs text-slate-800"
                    />
                  </td>
                  <td class="p-3 text-center">
                    <input
                      type="number"
                      min="1"
                      v-model="item.quantity"
                      class="w-20 text-center bg-white border border-slate-200 rounded-lg px-2 py-1.5 text-xs font-bold text-slate-900"
                    />
                  </td>
                  <td class="p-3">
                    <span class="px-2 py-1 bg-slate-100 rounded text-slate-700 font-semibold">{{ item.unit }}</span>
                  </td>
                  <td class="p-3 text-right">
                    <input
                      type="number"
                      min="0"
                      v-model="item.price"
                      class="w-28 text-right bg-white border border-slate-200 rounded-lg px-2 py-1.5 text-xs font-bold text-slate-900"
                    />
                  </td>
                  <td class="p-3 text-right font-bold text-slate-900">
                    {{ formatRupiah((item.quantity || 0) * (item.price || 0)) }}
                  </td>
                  <td class="p-3 text-center">
                    <button
                      type="button"
                      @click="handleRemoveItem(idx)"
                      :disabled="items.length <= 1"
                      class="text-red-400 hover:text-red-600 disabled:opacity-30 transition-colors"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Catatan & Ringkasan Nilai Faktur -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Catatan Faktur Pengiriman
            </label>
            <textarea
              v-model="notes"
              rows="3"
              placeholder="Tambahkan nomor referensi atau catatan internal..."
              class="w-full bg-slate-50 border border-slate-300 rounded-xl p-3 text-xs text-slate-800 focus:ring-2 focus:ring-indigo-500 focus:bg-white"
            ></textarea>
          </div>

          <div class="bg-gradient-to-br from-slate-900 to-indigo-950 text-white p-5 rounded-2xl shadow-md flex flex-col justify-between">
            <div class="space-y-1.5">
              <div class="flex justify-between text-xs text-indigo-200">
                <span>Total Kuantitas Barang</span>
                <span class="font-bold text-white">
                  {{ items.reduce((s, it) => s + (Number(it.quantity) || 0), 0) }} Unit
                </span>
              </div>
              <div class="flex justify-between text-xs text-indigo-200">
                <span>Total Item Produk</span>
                <span class="font-bold text-white">{{ items.length }} Produk</span>
              </div>
            </div>

            <div class="pt-3 border-t border-indigo-800/60 mt-3 flex justify-between items-end">
              <div>
                <span class="text-[10px] font-black uppercase text-indigo-300 tracking-wider">
                  Total Nilai Faktur (Piutang Cabang)
                </span>
                <div class="text-2xl font-black text-emerald-400">
                  {{ formatRupiah(totalInvoiceAmount) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-4 bg-slate-50 border-t border-slate-200 flex items-center justify-between">
        <Button
          type="button"
          variant="outline"
          @click="emit('close')"
          class="rounded-xl border-slate-300 font-bold"
        >
          Batal
        </Button>

        <Button
          type="button"
          @click="handleSubmit"
          :disabled="isSubmitting || !isFormValid"
          class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl px-6 py-2.5 font-bold shadow-md shadow-indigo-600/20 flex items-center gap-2"
        >
          <Save class="w-4 h-4" />
          {{ isSubmitting ? (isEditMode ? 'Menyimpan...' : 'Menerbitkan...') : (isEditMode ? 'Simpan Perubahan Faktur' : 'Terbitkan Faktur Pengiriman') }}
        </Button>
      </div>
    </div>
  </div>
</template>
