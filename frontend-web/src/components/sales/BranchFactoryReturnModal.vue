<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import apiClient from '@/api/axios'
import {
  X,
  RotateCcw,
  Plus,
  Trash2,
  Save,
  Package,
  AlertTriangle,
  Building2,
  Truck,
  CheckCircle2,
  Layers,
  FileText,
  Search,
  Sparkles,
  Calendar,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import ProductSearchSelect from '@/components/sales/ProductSearchSelect.vue'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { toast } from 'vue-sonner'

const props = defineProps<{
  show: boolean
  initialTransfer?: any
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', data: any): void
}>()

const masterDataStore = useMasterDataStore()
const factoryStore = useFactoryStore()

const inputMode = ref<'from_invoice' | 'manual'>('from_invoice')
const isSubmitting = ref(false)
const isLoadingInvoices = ref(false)

const factories = computed(() => factoryStore.factories || [])
const branches = computed(() => masterDataStore.branches || [])
const vehicles = ref<any[]>([])
const employees = computed(() => masterDataStore.employees || [])
const products = computed(() => factoryStore.sellableProducts || [])

// Invoices / Transfers list
const availableInvoices = ref<any[]>([])
const selectedInvoiceNo = ref('')
const invoiceSearchQuery = ref('')

const fetchVehicles = async () => {
  try {
    const res = await apiClient.get('/vehicles')
    vehicles.value = res.data?.data || res.data || []
  } catch {
    vehicles.value = []
  }
}

const fetchInvoices = async () => {
  isLoadingInvoices.value = true
  try {
    const res = await apiClient.get('/factory/stock/transfers')
    const list: any[] = res.data || []
    
    // Group by delivery_order_no or id
    const grouped: Record<string, any> = {}
    for (const item of list) {
      if (item.status === 'REQUESTED' || item.status === 'REJECTED') continue
      const key = item.delivery_order_no || item.id
      if (!grouped[key]) {
        grouped[key] = {
          id: item.id,
          delivery_order_no: item.delivery_order_no || `SJ-${item.id?.substring(0, 8)}`,
          status: item.status,
          to_branch_id: item.to_branch_id,
          to_branch: item.to_branch,
          from_factory_id: item.from_factory_id,
          from_factory: item.from_factory,
          vehicle_id: item.vehicle_id,
          driver_id: item.driver_id,
          created_at: item.created_at,
          shipped_at: item.shipped_at,
          total_amount: 0,
          items: [],
        }
      }
      grouped[key].total_amount += (item.total_amount || 0)
      grouped[key].items.push(item)
    }

    availableInvoices.value = Object.values(grouped).sort(
      (a: any, b: any) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    )
  } catch (error) {
    console.error('Failed to load transfers for return:', error)
  } finally {
    isLoadingInvoices.value = false
  }
}

const filteredInvoices = computed(() => {
  if (!invoiceSearchQuery.value) return availableInvoices.value.slice(0, 8)
  const q = invoiceSearchQuery.value.toLowerCase()
  return availableInvoices.value.filter(
    (inv) =>
      inv.delivery_order_no?.toLowerCase().includes(q) ||
      inv.to_branch?.name?.toLowerCase().includes(q) ||
      inv.from_factory?.name?.toLowerCase().includes(q) ||
      inv.items?.some((it: any) => it.product?.name?.toLowerCase().includes(q))
  )
})

const selectedBranchId = ref('')
const selectedFactoryId = ref('')
const selectedVehicleId = ref('')
const selectedDriverId = ref('')
const returnDate = ref(new Date().toISOString().split('T')[0])
const notes = ref('')

interface ReturnItemLine {
  product_id: string
  product_name?: string
  batch_no: string
  quantity: number
  unit: string
  return_type: 'GOOD' | 'DAMAGED'
  damage_type: 'BOTTLE_PACKAGING' | 'PRODUCT_CONTENT' | 'EXPIRED' | 'TOTAL_DAMAGE' | 'OTHER'
  damage_reason: string
  price: number
  max_quantity?: number
}

const items = ref<ReturnItemLine[]>([])

const handleSelectInvoice = (inv: any) => {
  selectedInvoiceNo.value = inv.delivery_order_no
  selectedBranchId.value = inv.to_branch_id || ''
  selectedFactoryId.value = inv.from_factory_id || ''
  selectedVehicleId.value = inv.vehicle_id || ''
  selectedDriverId.value = inv.driver_id || ''
  if (inv.created_at || inv.shipped_at) {
    const rawDate = inv.shipped_at || inv.created_at
    returnDate.value = new Date(rawDate).toISOString().split('T')[0]
  }
  notes.value = `[Retur Faktur ${inv.delivery_order_no}] Pengembalian barang rusak dari cabang ${inv.to_branch?.name || ''}`

  items.value = inv.items.map((it: any) => {
    const prod = products.value.find((p) => p.id === it.product_id) || it.product
    const defaultPrice = prod?.factory_price || prod?.base_price || 0
    return {
      product_id: it.product_id,
      product_name: prod?.name || 'Produk',
      batch_no: it.batch_no || 'DEFAULT',
      quantity: 1,
      unit: it.unit || 'PCS',
      return_type: 'DAMAGED' as const,
      damage_type: 'BOTTLE_PACKAGING' as const,
      damage_reason: `Barang Rusak / Kemasan Cacat dari Faktur ${inv.delivery_order_no}`,
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
    batch_no: 'DEFAULT',
    quantity: 1,
    unit: defaultProd ? (defaultProd.unit || 'PCS') : 'PCS',
    return_type: 'DAMAGED',
    damage_type: 'BOTTLE_PACKAGING',
    damage_reason: 'Barang Rusak / Kemasan Cacat dari Cabang',
    price: defaultProd ? (defaultProd.factory_price || defaultProd.base_price || 0) : 0,
  })
}

const handleRemoveItem = (index: number) => {
  items.value.splice(index, 1)
}

const handleProductChange = (index: number, productId: string) => {
  const prod: any = products.value.find((p) => p.id === productId)
  if (prod && items.value[index]) {
    items.value[index].product_id = prod.id
    items.value[index].product_name = prod.name
    items.value[index].price = prod.factory_price || prod.base_price || 0
    items.value[index].unit = prod.unit || 'PCS'
  }
}

const formatRupiah = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const totalEstimatedReturnCredit = computed(() => {
  return items.value.reduce((sum, it) => sum + (Number(it.quantity) || 0) * (it.price || 0), 0)
})

const isFormValid = computed(() => {
  if (!selectedBranchId.value || !selectedFactoryId.value) return false
  if (items.value.length === 0) return false
  return items.value.every((it) => it.product_id && it.quantity > 0)
})

const handleSubmit = async () => {
  if (!isFormValid.value) {
    toast.error('Mohon lengkapi cabang asal, pabrik tujuan, dan daftar barang retur.')
    return
  }

  isSubmitting.value = true
  try {
    const payload = {
      from_branch_id: selectedBranchId.value,
      to_factory_id: selectedFactoryId.value,
      return_date: returnDate.value ? new Date(returnDate.value).toISOString() : undefined,
      notes: notes.value,
      vehicle_id: selectedVehicleId.value || undefined,
      driver_id: selectedDriverId.value || undefined,
      items: items.value.map((it) => ({
        product_id: it.product_id,
        batch_no: it.batch_no || 'DEFAULT',
        quantity: Number(it.quantity),
        unit: it.unit || 'PCS',
        pcs_per_unit: 1,
        return_type: it.return_type,
        damage_type: it.damage_type,
        damage_reason: it.damage_reason,
      })),
    }

    const res = await apiClient.post('/factory-returns', payload)
    toast.success('Retur Cabang ke Pabrik berhasil diajukan! Setelah diverifikasi pabrik, piutang cabang otomatis dipotong.')
    emit('created', res.data)
    emit('close')
  } catch (error: any) {
    const msg = error.response?.data?.error || error.message || 'Gagal mengajukan retur cabang'
    toast.error(msg)
  } finally {
    isSubmitting.value = false
  }
}

watch(
  () => props.show,
  async (val) => {
    if (val) {
      if (factories.value.length === 0) {
        await factoryStore.fetchFactories()
      }
      if (products.value.length === 0) {
        await factoryStore.fetchProducts()
      }
      if (branches.value.length === 0) {
        await masterDataStore.fetchBranches()
      }
      if (factories.value.length > 0 && !selectedFactoryId.value) {
        selectedFactoryId.value = factories.value[0]?.id || ''
      }
      if (branches.value.length > 0 && !selectedBranchId.value) {
        selectedBranchId.value = branches.value[0]?.id || ''
      }
      fetchVehicles()
      fetchInvoices()

      if (props.initialTransfer) {
        inputMode.value = 'from_invoice'
        handleSelectInvoice(props.initialTransfer)
      } else if (items.value.length === 0) {
        handleAddItem()
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
      <div class="px-6 py-5 bg-gradient-to-r from-slate-900 via-rose-950 to-slate-900 text-white flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-rose-500/20 border border-rose-400/30 rounded-xl">
            <RotateCcw class="w-6 h-6 text-rose-300" />
          </div>
          <div>
            <h2 class="text-xl font-black tracking-tight flex items-center gap-2">
              Input Retur Barang Cabang ke Pabrik
              <span class="text-xs bg-rose-500/30 text-rose-200 px-2 py-0.5 rounded-full font-semibold border border-rose-400/20">
                Kredit Piutang Cabang
              </span>
            </h2>
            <p class="text-xs text-slate-300 mt-0.5">
              Catat pengembalian barang rusak / BS / sisa dari Cabang ke Pabrik Utama. Persetujuan retur otomatis memotong saldo Piutang Cabang.
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

      <!-- Mode Selector -->
      <div class="px-6 pt-4 pb-2 bg-slate-50 border-b border-slate-200">
        <div class="grid grid-cols-2 gap-3 max-w-md">
          <button
            type="button"
            @click="inputMode = 'from_invoice'"
            :class="[
              'flex items-center justify-center gap-2 py-2 px-3 text-xs font-bold rounded-xl border transition-all',
              inputMode === 'from_invoice'
                ? 'bg-rose-600 text-white border-rose-600 shadow-sm'
                : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100',
            ]"
          >
            <FileText class="w-4 h-4" />
            Tarik dari Faktur / SJ Pabrik
            <span
              v-if="availableInvoices.length > 0"
              class="ml-1 px-1.5 py-0.2 bg-amber-400 text-slate-900 rounded-full text-[10px] font-black"
            >
              {{ availableInvoices.length }}
            </span>
          </button>
          <button
            type="button"
            @click="inputMode = 'manual'"
            :class="[
              'flex items-center justify-center gap-2 py-2 px-3 text-xs font-bold rounded-xl border transition-all',
              inputMode === 'manual'
                ? 'bg-rose-600 text-white border-rose-600 shadow-sm'
                : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-100',
            ]"
          >
            <Layers class="w-4 h-4" />
            Input Manual Retur
          </button>
        </div>
      </div>

      <!-- Body Form -->
      <div class="p-6 space-y-6 overflow-y-auto flex-1">
        <!-- Invoice Picker Card (when from_invoice) -->
        <div v-if="inputMode === 'from_invoice'" class="bg-rose-50/60 border border-rose-100 rounded-2xl p-4 space-y-3">
          <div class="flex items-center justify-between">
            <label class="text-xs font-black uppercase text-rose-900 tracking-wider flex items-center gap-1.5">
              <Sparkles class="w-4 h-4 text-rose-600" />
              Pilih Faktur / Surat Jalan Pengiriman Pabrik (DO)
            </label>
            <span v-if="isLoadingInvoices" class="text-xs text-rose-600 font-medium">Memuat data faktur...</span>
          </div>

          <div class="relative">
            <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              v-model="invoiceSearchQuery"
              placeholder="Cari No. Surat Jalan / Faktur (SJ-...), Nama Cabang, atau Produk..."
              class="w-full pl-9 pr-4 py-2 bg-white border border-rose-200 rounded-xl text-xs font-medium text-slate-800 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-rose-500"
            />
          </div>

          <div v-if="availableInvoices.length === 0 && !isLoadingInvoices" class="text-xs text-slate-500 py-2">
            Belum ada faktur pengiriman cabang yang tercatat. Anda dapat beralih ke mode <b>Input Manual Retur</b>.
          </div>

          <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2.5 max-h-48 overflow-y-auto pr-1">
            <div
              v-for="inv in filteredInvoices"
              :key="inv.delivery_order_no"
              @click="handleSelectInvoice(inv)"
              :class="[
                'p-2.5 rounded-xl border cursor-pointer transition-all flex flex-col justify-between text-left',
                selectedInvoiceNo === inv.delivery_order_no
                  ? 'bg-rose-600 text-white border-rose-600 shadow-md ring-2 ring-rose-400'
                  : 'bg-white hover:border-rose-300 text-slate-800 border-slate-200',
              ]"
            >
              <div>
                <div class="flex items-center justify-between text-xs font-bold">
                  <span>{{ inv.delivery_order_no }}</span>
                  <span :class="selectedInvoiceNo === inv.delivery_order_no ? 'text-rose-200' : 'text-rose-600'">
                    {{ inv.to_branch?.name || 'Cabang' }}
                  </span>
                </div>
                <div class="text-[10px] mt-1 opacity-80 flex items-center justify-between">
                  <span>{{ inv.items?.length || 0 }} item produk</span>
                  <span class="font-bold">{{ formatRupiah(inv.total_amount) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Branch & Factory Selector -->
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 gap-4">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5 flex items-center gap-1">
              <Calendar class="w-3.5 h-3.5 text-rose-600" />
              Tanggal Retur *
            </label>
            <input
              type="date"
              v-model="returnDate"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-semibold text-slate-900 focus:ring-2 focus:ring-rose-500 focus:bg-white"
            />
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Cabang Pengirim (Retur) *
            </label>
            <select
              v-model="selectedBranchId"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-semibold text-slate-900 focus:ring-2 focus:ring-rose-500 focus:bg-white"
            >
              <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Pabrik Tujuan Retur *
            </label>
            <select
              v-model="selectedFactoryId"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-semibold text-slate-900 focus:ring-2 focus:ring-rose-500 focus:bg-white"
            >
              <option v-for="f in factories" :key="f.id" :value="f.id">{{ f.name }}</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Kendaraan Angkut
            </label>
            <select
              v-model="selectedVehicleId"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-medium text-slate-900 focus:ring-2 focus:ring-rose-500 focus:bg-white"
            >
              <option value="">-- Tanpa Kendaraan Khusus --</option>
              <option v-for="v in vehicles" :key="v.id" :value="v.id">
                {{ v.name }} ({{ v.plate_number }})
              </option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Driver Pembawa
            </label>
            <select
              v-model="selectedDriverId"
              class="w-full bg-slate-50 border border-slate-300 rounded-xl px-3 py-2 text-sm font-medium text-slate-900 focus:ring-2 focus:ring-rose-500 focus:bg-white"
            >
              <option value="">-- Pilih Driver --</option>
              <option v-for="emp in employees" :key="emp.id" :value="emp.id">
                {{ emp.first_name }} {{ emp.last_name || '' }}
              </option>
            </select>
          </div>
        </div>

        <!-- Items List Cards -->
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-black text-slate-900 uppercase tracking-wider flex items-center gap-2">
              <Package class="w-4 h-4 text-rose-600" />
              Daftar Barang yang Dikembalikan (Retur)
            </h3>
            <Button
              type="button"
              variant="outline"
              size="sm"
              @click="handleAddItem"
              class="text-xs font-bold border-rose-200 text-rose-700 hover:bg-rose-50 rounded-xl"
            >
              <Plus class="w-3.5 h-3.5 mr-1" /> Tambah Baris Retur
            </Button>
          </div>

          <div class="space-y-3.5">
            <div
              v-for="(item, idx) in items"
              :key="idx"
              class="bg-slate-50/90 rounded-2xl p-4 border border-slate-200 transition-all hover:border-slate-300 relative group space-y-3"
            >
              <!-- Top Row: Item Index Badge, Product Selector & Remove Button -->
              <div class="flex items-start gap-3">
                <span class="w-7 h-7 rounded-xl bg-rose-100 text-rose-700 flex items-center justify-center text-xs font-black shrink-0 mt-0.5">
                  {{ idx + 1 }}
                </span>
                <div class="flex-1 min-w-0">
                  <label class="block text-[10px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">
                    Pilih Produk Retur *
                  </label>
                  <ProductSearchSelect
                    :model-value="item.product_id"
                    :products="products"
                    :show-catalog-button="false"
                    @update:model-value="(val) => handleProductChange(idx, val)"
                  />
                </div>
                <button
                  v-if="items.length > 1"
                  type="button"
                  @click="handleRemoveItem(idx)"
                  class="p-2 text-rose-400 hover:text-rose-600 hover:bg-rose-100/50 rounded-xl transition-colors shrink-0 mt-5 cursor-pointer"
                  title="Hapus Baris Retur"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>

              <!-- Bottom Row: Kondisi, Jenis Cacat, Qty, Harga Satuan, Subtotal Potongan -->
              <div class="grid grid-cols-2 sm:grid-cols-12 gap-3 items-end pt-2 border-t border-slate-200/80">
                <!-- Kondisi -->
                <div class="col-span-1 sm:col-span-3">
                  <label class="block text-[10px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">
                    Kondisi
                  </label>
                  <select
                    v-model="item.return_type"
                    class="w-full bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-bold text-slate-800 focus:ring-2 focus:ring-rose-500 outline-none"
                  >
                    <option value="DAMAGED">🔴 Rusak / BS</option>
                    <option value="GOOD">🟢 Stok Baik</option>
                  </select>
                </div>

                <!-- Jenis Cacat -->
                <div class="col-span-1 sm:col-span-3">
                  <label class="block text-[10px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">
                    Jenis Cacat
                  </label>
                  <select
                    v-model="item.damage_type"
                    class="w-full bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-bold text-slate-800 focus:ring-2 focus:ring-rose-500 outline-none"
                  >
                    <option value="BOTTLE_PACKAGING">Botol/Karton Rusak</option>
                    <option value="PRODUCT_CONTENT">Isi Minuman Cacat</option>
                    <option value="EXPIRED">Expired / Kadaluarsa</option>
                    <option value="TOTAL_DAMAGE">Rusak Total</option>
                    <option value="OTHER">Lainnya</option>
                  </select>
                </div>

                <!-- Kuantitas Retur -->
                <div class="col-span-1 sm:col-span-2">
                  <label class="block text-[10px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">
                    Qty Retur *
                  </label>
                  <div class="flex items-center gap-1.5">
                    <input
                      type="number"
                      min="1"
                      v-model.number="item.quantity"
                      class="w-full bg-white border border-slate-200 rounded-xl px-2.5 py-2 text-xs font-black text-center text-slate-900 focus:ring-2 focus:ring-rose-500 outline-none"
                    />
                    <span class="text-[11px] text-slate-600 font-bold px-1.5 py-1 bg-slate-200/70 rounded-lg shrink-0">
                      {{ item.unit || 'PCS' }}
                    </span>
                  </div>
                </div>

                <!-- Nilai Satuan (Rp) -->
                <div class="col-span-1 sm:col-span-2">
                  <label class="block text-[10px] font-black text-slate-500 uppercase tracking-wider mb-1 ml-0.5">
                    Nilai Satuan (Rp)
                  </label>
                  <input
                    type="number"
                    min="0"
                    v-model.number="item.price"
                    class="w-full bg-white border border-slate-200 rounded-xl px-2.5 py-2 text-xs font-black text-right text-slate-900 focus:ring-2 focus:ring-rose-500 outline-none"
                  />
                </div>

                <!-- Potongan Piutang -->
                <div class="col-span-2 sm:col-span-2 text-right">
                  <span class="block text-[9px] font-black text-rose-500 uppercase tracking-wider mb-1">
                    Potongan Piutang
                  </span>
                  <div class="text-xs font-black text-rose-600 bg-rose-50 px-2 py-2 rounded-xl border border-rose-100 truncate">
                    {{ formatRupiah((item.quantity || 0) * (item.price || 0)) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Notes & Summary -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
          <div>
            <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
              Alasan & Catatan Retur
            </label>
            <textarea
              v-model="notes"
              rows="3"
              placeholder="Jelaskan alasan pengembalian barang atau catatan fisik..."
              class="w-full bg-slate-50 border border-slate-300 rounded-xl p-3 text-xs text-slate-800 focus:ring-2 focus:ring-rose-500 focus:bg-white"
            ></textarea>
          </div>

          <div class="bg-gradient-to-br from-slate-900 to-rose-950 text-white p-5 rounded-2xl shadow-md flex flex-col justify-between">
            <div class="space-y-1.5">
              <div class="flex justify-between text-xs text-rose-200">
                <span>Total Kuantitas Retur</span>
                <span class="font-bold text-white">
                  {{ items.reduce((s, it) => s + (Number(it.quantity) || 0), 0) }} Unit
                </span>
              </div>
              <div class="flex justify-between text-xs text-rose-200">
                <span>Total Item Barang</span>
                <span class="font-bold text-white">{{ items.length }} Produk</span>
              </div>
            </div>

            <div class="pt-3 border-t border-rose-800/60 mt-3 flex justify-between items-end">
              <div>
                <span class="text-[10px] font-black uppercase text-rose-300 tracking-wider">
                  Estimasi Potongan Piutang Cabang
                </span>
                <div class="text-2xl font-black text-rose-400">
                  - {{ formatRupiah(totalEstimatedReturnCredit) }}
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
          class="bg-rose-600 hover:bg-rose-700 text-white rounded-xl px-6 py-2.5 font-bold shadow-md shadow-rose-600/20 flex items-center gap-2"
        >
          <Save class="w-4 h-4" />
          {{ isSubmitting ? 'Mengirim Pengajuan...' : 'Ajukan Retur ke Pabrik' }}
        </Button>
      </div>
    </div>
  </div>
</template>
