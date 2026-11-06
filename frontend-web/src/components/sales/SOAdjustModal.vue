<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue'
import { useSalesOrderStore, type SalesOrder } from '@/stores/salesOrder'
import { useFactoryStore } from '@/stores/factory'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'
import { X, Tag, Package, History, AlertTriangle, CheckCircle2, Loader2, Trash2, Plus, FileText, ZoomIn, ZoomOut, RotateCcw, Gift } from 'lucide-vue-next'
import api from '@/api/axios'
import ProductSearchSelect from './ProductSearchSelect.vue'
import { isAncillaryOrGiftProduct } from '@/utils/productUnits'

const props = defineProps<{ order: SalesOrder | null; show: boolean }>()
const emit = defineEmits(['close', 'saved'])
const soStore = useSalesOrderStore()
const factoryStore = useFactoryStore()
const authStore = useAuthStore()

const tab = ref<'items' | 'discount' | 'bonus' | 'debt'>('items')
const loading = ref(false)
const debtLoading = ref(false)
const debtHistory = ref<any[]>([])
const zoomLevel = ref(1)

const UNIT_OPTIONS = [
  'PCS',
  'DUS',
  'PACK',
  'KARTON',
  'RENCENG',
  'BAL',
  'BOTOL',
  'SAK',
]

// Form state - items adjust list
const adjustedItems = ref<
  {
    product_id: string
    ordered_quantity: number
    unit: string
    pieces_per_unit: number
    price: number
    discount_percent: number
  }[]
>([])
const discountAmount = ref(0)
const adjustNotes = ref('')

const bonusItems = ref<
  {
    selection_id: string
    quantity: number
    notes: string
  }[]
>([])
const bonusInventories = ref<any[]>([])

const combinedBonusOptions = computed(() => {
  const options: { id: string, name: string, quantity?: number, unit: string, type: 'inventory' | 'product' }[] = []
  
  // Bonus Inventories
  bonusInventories.value.forEach(b => {
    options.push({
      id: `inv_${b.id}`,
      name: b.name,
      quantity: b.quantity,
      unit: b.unit,
      type: 'inventory'
    })
  })

  // Barang Sampingan & Hadiah Products
  const sampinganProducts = factoryStore.products.filter(p => isAncillaryOrGiftProduct(p))
  sampinganProducts.forEach(p => {
    options.push({
      id: `prod_${p.id}`,
      name: p.name,
      unit: p.unit || 'PCS',
      type: 'product'
    })
  })

  return options
})

const formatCurrency = (v: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(v)
const formatDate = (d: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
    : '-'

async function loadBonusInventories() {
  const branchId = props.order?.branch_id || soStore.orders[0]?.branch_id || authStore.user?.branch_id || ''
  if (!branchId) return
  try {
    const res = await api.get('/bonus-inventory/branch/' + branchId)
    bonusInventories.value = res.data || []
  } catch (e) {
    // Ignore error
  }
}

onMounted(async () => {
  factoryStore.fetchProducts()
  loadBonusInventories()
})

watch(
  () => props.order,
  (order) => {
    if (!order) return
    loadBonusInventories()
    adjustedItems.value = (order.items || []).map((i) => ({
      product_id: i.product_id,
      ordered_quantity: Number(i.ordered_quantity) || 1,
      unit: i.unit || 'PCS',
      pieces_per_unit: Number(i.pieces_per_unit) || 1,
      price: Number(i.price) || 0,
      discount_percent: Number(i.discount_percent) || 0,
    }))
    discountAmount.value = Number(order.discount_amount) || 0
    adjustNotes.value = ''
    bonusItems.value = (order.bonus_items || []).map((b: any) => ({
      selection_id: `inv_${b.bonus_inventory_id}`,
      quantity: Number(b.quantity) || 1,
      notes: b.notes || '',
    }))
    tab.value = 'items'
    debtHistory.value = []
    zoomLevel.value = 1
  },
  { immediate: true },
)

watch(tab, async (t) => {
  if (t === 'debt' && props.order && debtHistory.value.length === 0) {
    debtLoading.value = true
    try {
      debtHistory.value = await soStore.fetchStoreDebtHistory(props.order.store_id)
    } catch {
      toast.error('Gagal memuat riwayat piutang')
    } finally {
      debtLoading.value = false
    }
  }
})

// Item adjustments logic
function addItem() {
  adjustedItems.value.push({
    product_id: '',
    ordered_quantity: 1,
    unit: 'PCS',
    pieces_per_unit: 1,
    price: 0,
    discount_percent: 0,
  })
}

function addBonusItem() {
  bonusItems.value.push({
    selection_id: '',
    quantity: 1,
    notes: ''
  })
}

function removeBonusItem(index: number) {
  bonusItems.value.splice(index, 1)
}

function removeItem(index: number) {
  adjustedItems.value.splice(index, 1)
}

function onProductChange(index: number) {
  const item = adjustedItems.value[index]
  if (!item) return
  const p = factoryStore.products.find((prod) => prod.id === item.product_id)
  if (p) {
    item.price = Number(p.selling_price) || 0
  }
}

const subtotalAfterAdjust = computed(() =>
  adjustedItems.value.reduce((sum, i) => {
    const qty = (Number(i.ordered_quantity) || 0) * (Number(i.pieces_per_unit) || 1)
    const discountFactor = 1.0 - (Number(i.discount_percent) || 0) / 100.0
    return sum + qty * (Number(i.price) || 0) * discountFactor
  }, 0),
)
const totalAfterDiscount = computed({
  get: () => Math.max(0, subtotalAfterAdjust.value - (Number(discountAmount.value) || 0)),
  set: (val: number) => {
    discountAmount.value = Math.max(0, subtotalAfterAdjust.value - (Number(val) || 0))
  }
})
const originalTotal = computed(() => Number(props.order?.total_amount) || 0)

const totalDebt = computed(() =>
  debtHistory.value.reduce(
    (sum: number, t: any) => sum + ((Number(t.total_amount) || 0) - (Number(t.payment_collected_amount) || 0)),
    0,
  ),
)

async function handleSave() {
  if (!props.order) return
  if (adjustedItems.value.length === 0) {
    toast.error('Pesanan harus memiliki minimal 1 item barang')
    return
  }
  if (adjustedItems.value.some((i) => !i.product_id)) {
    toast.error('Lengkapi pilihan produk untuk semua baris item')
    return
  }
  if (adjustedItems.value.some((i) => (Number(i.ordered_quantity) || 0) <= 0)) {
    toast.error('Jumlah barang harus lebih besar dari 0')
    return
  }

  loading.value = true
  try {
    const finalBonusItems: any[] = []
    const extraAdjustedItems: any[] = []

    for (const b of bonusItems.value) {
      if (!b.selection_id) continue
      if (b.selection_id.startsWith('inv_')) {
        finalBonusItems.push({
          bonus_inventory_id: b.selection_id.replace('inv_', ''),
          quantity: Number(b.quantity) || 1,
          notes: b.notes || '',
        })
      } else if (b.selection_id.startsWith('prod_')) {
        const prodId = b.selection_id.replace('prod_', '')
        const p = factoryStore.products.find((x: any) => x.id === prodId)
        if (p) {
          extraAdjustedItems.push({
            product_id: p.id,
            ordered_quantity: Number(b.quantity) || 1,
            unit: p.unit || 'PCS',
            pieces_per_unit: Number(p.pcs_per_unit) || 1,
            price: Number(p.selling_price) || 0,
            discount_percent: 100,
          })
        }
      }
    }

    const payloadItems = [...adjustedItems.value, ...extraAdjustedItems].map((item) => ({
      product_id: item.product_id,
      ordered_quantity: Number(item.ordered_quantity) || 0,
      unit: item.unit || 'PCS',
      pieces_per_unit: Number(item.pieces_per_unit) || 1,
      price: Number(item.price) || 0,
      discount_percent: Number(item.discount_percent) || 0,
    }))

    await soStore.adjustSO(props.order.id, {
      discount_amount: Number(discountAmount.value) || 0,
      adjust_notes: adjustNotes.value || '',
      items: payloadItems,
      bonus_items: finalBonusItems,
    })
    toast.success('Pesanan berhasil disesuaikan')
    emit('saved')
    emit('close')
  } catch (e: any) {
    toast.error(e?.response?.data?.error || 'Gagal menyesuaikan pesanan')
  } finally {
    loading.value = false
  }
}

const canAdjust = computed(() => {
  const isAllowedStatus =
    props.order?.status === 'DRAFT' ||
    props.order?.status === 'CONFIRMED' ||
    props.order?.status === 'WAITING_STOCK'
    
  const isAdminNota = authStore.user?.role === 'admin_nota'
  
  return isAllowedStatus || isAdminNota
})
</script>

<template>
  <div v-if="show && order" class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="$emit('close')" />
    <div
      class="bg-white rounded-[32px] w-full relative z-10 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden"
      :class="order.is_physical_nota ? 'max-w-6xl h-[90vh]' : 'max-w-2xl'"
    >
      <!-- Header -->
      <div
        class="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-gradient-to-r from-slate-50 to-white"
      >
        <div>
          <div class="flex items-center gap-2">
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
              >ADJUST PESANAN</span
            >
            <span
              class="text-[9px] font-black px-2 py-0.5 rounded-full"
              :class="
                order.status === 'DRAFT'
                  ? 'bg-slate-100 text-slate-600'
                  : order.status === 'WAITING_STOCK'
                  ? 'bg-orange-100 text-orange-700'
                  : 'bg-green-100 text-green-700'
              "
            >
              {{ order.status }}
            </span>
          </div>
          <h3 class="text-xl font-black text-slate-900 mt-0.5">{{ order.so_number }}</h3>
          <p class="text-xs text-slate-500 font-medium">{{ order.store?.name }}</p>
        </div>
        <button
          @click="$emit('close')"
          class="p-2 hover:bg-slate-100 rounded-xl text-slate-400 hover:text-slate-900 transition-colors"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Main Layout: Split Screen or Single Column -->
      <div class="flex flex-col lg:flex-row flex-1 overflow-hidden min-h-0">
        
        <!-- Preview Nota Fisik (Left Side) -->
        <div v-if="order.is_physical_nota" class="w-full lg:w-1/2 bg-slate-200/50 border-r border-slate-200 overflow-hidden flex flex-col items-center justify-center relative p-0">
          <div class="absolute inset-0 z-0 opacity-10 bg-[radial-gradient(#94a3b8_1px,transparent_1px)] [background-size:16px_16px]"></div>
          
          <!-- Zoom Controls -->
          <div v-if="order.physical_nota_image_url" class="absolute bottom-4 left-1/2 -translate-x-1/2 z-20 flex items-center gap-1 bg-white/90 p-1.5 rounded-2xl shadow-lg backdrop-blur border border-slate-200">
             <button @click.prevent="zoomLevel = Math.max(0.5, zoomLevel - 0.25)" class="p-2 hover:bg-slate-100 rounded-xl text-slate-600 transition-colors" title="Zoom Out">
               <ZoomOut class="w-4 h-4" />
             </button>
             <span class="text-xs font-bold text-slate-700 w-12 text-center">{{ Math.round(zoomLevel * 100) }}%</span>
             <button @click.prevent="zoomLevel = Math.min(4, zoomLevel + 0.25)" class="p-2 hover:bg-slate-100 rounded-xl text-slate-600 transition-colors" title="Zoom In">
               <ZoomIn class="w-4 h-4" />
             </button>
             <div class="w-px h-4 bg-slate-300 mx-1"></div>
             <button @click.prevent="zoomLevel = 1" class="p-2 hover:bg-slate-100 rounded-xl text-slate-600 transition-colors" title="Reset Zoom">
               <RotateCcw class="w-4 h-4" />
             </button>
          </div>

          <!-- Image Container -->
          <div class="absolute inset-0 z-10 overflow-auto flex items-center justify-center">
            <div :style="{ transform: `scale(${zoomLevel})`, transformOrigin: 'center center', transition: 'transform 0.2s cubic-bezier(0.4, 0, 0.2, 1)' }" class="w-full h-full flex items-center justify-center min-w-[300px] min-h-[300px] p-8">
              <img v-if="order.physical_nota_image_url" :src="order.physical_nota_image_url" class="max-w-full max-h-full object-contain shadow-2xl rounded-2xl border-4 border-white" alt="Nota Fisik" />
              <div v-else class="flex flex-col items-center justify-center text-slate-400">
                 <FileText class="w-16 h-16 mb-2 opacity-50" />
                 <p class="font-bold text-sm">Tidak ada gambar nota</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Form Adjust (Right Side / Full) -->
        <div class="flex-1 flex flex-col min-h-0" :class="order.is_physical_nota ? 'lg:w-1/2' : ''">
          <!-- Tabs -->
      <div class="flex gap-1 px-6 pt-4 border-b border-slate-100">
        <button
          v-for="t in [
            { key: 'items', label: 'Penyesuaian Item', icon: Package },
            { key: 'discount', label: 'Diskon', icon: Tag },
            { key: 'bonus', label: 'Bonus', icon: Gift },
            { key: 'debt', label: 'Riwayat Piutang', icon: History },
          ]"
          :key="t.key"
          @click="tab = t.key as any"
          :class="[
            'flex items-center gap-1.5 px-4 py-2.5 text-xs font-black rounded-t-xl border-b-2 transition-all -mb-px',
            tab === t.key
              ? 'border-primary text-primary bg-primary/5'
              : 'border-transparent text-slate-400 hover:text-slate-600',
          ]"
        >
          <component :is="t.icon" class="w-3.5 h-3.5" />
          {{ t.label }}
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-6 space-y-4">
        <!-- Warning: non-adjustable status -->
        <div
          v-if="!canAdjust"
          class="flex items-center gap-3 bg-amber-50 border border-amber-200 rounded-2xl p-4"
        >
          <AlertTriangle class="w-5 h-5 text-amber-500 shrink-0" />
          <p class="text-sm text-amber-700 font-semibold">
            Penyesuaian hanya bisa dilakukan pada SO berstatus <strong>DRAFT</strong>,
            <strong>CONFIRMED</strong>, atau <strong>WAITING_STOCK</strong>. SO ini berstatus <strong>{{ order.status }}</strong>.
            (Kecuali bagi peran Admin Nota yang memiliki akses fleksibel)
          </p>
        </div>

        <!-- Tab: Items -->
        <div v-if="tab === 'items'" class="space-y-4">
          <div class="flex justify-between items-center">
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
              Daftar Barang & Harga
            </p>
            <button
              v-if="canAdjust"
              @click="addItem"
              class="flex items-center gap-1 px-3 py-1.5 bg-primary/10 hover:bg-primary/20 text-primary text-xs font-black rounded-xl transition-all"
            >
              <Plus class="w-3.5 h-3.5" />
              Tambah Item
            </button>
          </div>

          <div class="space-y-3 max-h-[45vh] overflow-y-auto pr-1">
            <div
              v-for="(item, idx) in adjustedItems"
              :key="idx"
              class="bg-slate-50 rounded-2xl p-4 border border-slate-200 relative group transition-all hover:border-slate-300"
            >
              <!-- Item Header (Product Select + Remove Button) -->
              <div class="flex items-center gap-3 mb-3">
                <div class="flex-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Nama Produk</label>
                  <ProductSearchSelect
                    v-model="item.product_id"
                    :products="factoryStore.sellableProducts"
                    :disabled="!canAdjust"
                    @change="onProductChange(idx)"
                  />
                </div>
                <button
                  v-if="canAdjust"
                  @click="removeItem(idx)"
                  class="p-2.5 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-xl mt-3.5 transition-colors"
                  title="Hapus item"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>

              <!-- Item Fields Grid -->
              <div class="grid grid-cols-5 gap-2">
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Jumlah</label>
                  <input
                    v-model.number="item.ordered_quantity"
                    type="number"
                    min="1"
                    :disabled="!canAdjust"
                    class="w-full bg-white border border-slate-200 rounded-xl px-3 py-2.5 text-xs font-black focus:ring-2 focus:ring-primary/20 outline-none disabled:opacity-50"
                  />
                </div>
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Satuan</label>
                  <select
                    v-model="item.unit"
                    :disabled="!canAdjust"
                    class="w-full bg-white border border-slate-200 rounded-xl px-2 py-2.5 text-xs font-bold focus:ring-2 focus:ring-primary/20 outline-none disabled:opacity-50"
                  >
                    <option v-for="u in UNIT_OPTIONS" :key="u" :value="u">{{ u }}</option>
                  </select>
                </div>
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Isi/Unit</label>
                  <input
                    v-model.number="item.pieces_per_unit"
                    type="number"
                    min="1"
                    :disabled="!canAdjust"
                    class="w-full bg-white border border-slate-200 rounded-xl px-3 py-2.5 text-xs font-black focus:ring-2 focus:ring-primary/20 outline-none disabled:opacity-50"
                  />
                </div>
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Harga Satuan</label>
                  <div class="relative">
                    <span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-[9px] font-bold text-slate-400">Rp</span>
                    <input
                      v-model.number="item.price"
                      type="number"
                      min="0"
                      :disabled="!canAdjust"
                      class="w-full bg-white border border-slate-200 rounded-xl pl-7 pr-2 py-2.5 text-xs font-black focus:ring-2 focus:ring-primary/20 outline-none disabled:opacity-50"
                    />
                  </div>
                </div>
                <div class="space-y-1">
                  <label class="text-[8px] font-black text-slate-400 uppercase ml-1">Diskon (%)</label>
                  <div class="relative">
                    <input
                      v-model.number="item.discount_percent"
                      type="number"
                      min="0"
                      max="100"
                      step="0.1"
                      :disabled="!canAdjust"
                      class="w-full bg-white border border-slate-200 rounded-xl pr-6 pl-3 py-2.5 text-xs font-black focus:ring-2 focus:ring-primary/20 outline-none disabled:opacity-50"
                    />
                    <span class="absolute right-2.5 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">%</span>
                  </div>
                </div>
              </div>

              <!-- Item Subtotal Summary -->
              <div class="mt-3 flex justify-between items-center text-[10px] font-bold text-slate-400 border-t border-slate-100 pt-2">
                <span>Total: {{ item.ordered_quantity * item.pieces_per_unit }} PCS</span>
                <span class="text-slate-600">
                  Subtotal: 
                  <span class="text-primary font-black">
                    {{ formatCurrency(item.ordered_quantity * item.pieces_per_unit * item.price * (1 - item.discount_percent/100)) }}
                  </span>
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab: Discount -->
        <div v-if="tab === 'discount'" class="space-y-5">
          <!-- Diskon per Item -->
          <div class="space-y-3">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
              >Diskon per Item (Persen %)</label
            >
            <div class="space-y-2 max-h-[30vh] overflow-y-auto pr-1">
              <div
                v-for="(item, idx) in adjustedItems"
                :key="idx"
                class="flex items-center justify-between bg-slate-50 rounded-xl p-3 border border-slate-200 hover:border-slate-300 transition-all"
              >
                <div class="flex-1 min-w-0 pr-3">
                  <p class="text-xs font-black text-slate-900 truncate">
                    {{ factoryStore.products.find((p) => p.id === item.product_id)?.name || 'Produk Tidak Dikenal' }}
                  </p>
                  <p class="text-[10px] text-slate-500 font-semibold mt-0.5">
                    {{ item.ordered_quantity }} {{ item.unit }} ({{ item.ordered_quantity * item.pieces_per_unit }} PCS) @ {{ formatCurrency(item.price) }}
                  </p>
                </div>
                <div class="flex items-center gap-4">
                  <div class="text-right">
                    <p class="text-[9px] text-slate-400 font-bold uppercase">Subtotal</p>
                    <p class="text-xs font-black text-slate-700">
                      {{ formatCurrency(item.ordered_quantity * item.pieces_per_unit * item.price * (1 - (item.discount_percent || 0)/100)) }}
                    </p>
                  </div>
                  <div class="relative w-24">
                    <input
                      v-model.number="item.discount_percent"
                      type="number"
                      min="0"
                      max="100"
                      step="0.1"
                      :disabled="!canAdjust"
                      class="w-full bg-white border border-slate-200 rounded-xl pr-6 pl-3 py-1.5 text-xs font-black text-center focus:ring-2 focus:ring-primary/20 outline-none disabled:opacity-50"
                      placeholder="0"
                    />
                    <span class="absolute right-2 top-1/2 -translate-y-1/2 text-[10px] font-bold text-slate-400">%</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
              >Diskon Global (Rupiah)</label
            >
            <div class="relative">
              <span
                class="absolute left-4 top-1/2 -translate-y-1/2 text-sm font-black text-slate-400"
                >Rp</span
              >
              <input
                v-model.number="discountAmount"
                type="number"
                min="0"
                :disabled="!canAdjust"
                class="w-full pl-10 pr-4 py-3.5 bg-slate-50 border-2 border-transparent rounded-2xl font-black text-lg text-slate-900 focus:bg-white focus:border-primary/20 focus:outline-none disabled:opacity-40"
              />
            </div>
          </div>

          <!-- Summary -->
          <div class="bg-slate-50 rounded-2xl divide-y divide-slate-100">
            <div class="flex justify-between p-4 text-sm">
              <span class="text-slate-500 font-semibold">Subtotal setelah adjust item</span>
              <span class="font-black text-slate-900">{{
                formatCurrency(subtotalAfterAdjust)
              }}</span>
            </div>
            <div class="flex justify-between p-4 text-sm text-red-500">
              <span class="font-semibold">Diskon Global</span>
              <span class="font-black">- {{ formatCurrency(discountAmount) }}</span>
            </div>
            <div class="flex justify-between p-4 items-center">
              <span class="font-black text-slate-700">TOTAL AKHIR (GRAND TOTAL)</span>
              <div class="relative w-48">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm font-black text-primary">Rp</span>
                <input
                  v-model.number="totalAfterDiscount"
                  type="number"
                  min="0"
                  :disabled="!canAdjust"
                  class="w-full pl-9 pr-3 py-2 bg-white border-2 border-primary/20 rounded-xl font-black text-lg text-primary focus:bg-white focus:border-primary/50 focus:outline-none disabled:opacity-40 text-right"
                />
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
              >Catatan Penyesuaian</label
            >
            <textarea
              v-model="adjustNotes"
              rows="3"
              :disabled="!canAdjust"
              placeholder="Alasan diskon, negosiasi harga, dll..."
              class="w-full bg-slate-50 border-2 border-transparent rounded-2xl px-4 py-3 font-medium text-sm focus:bg-white focus:border-primary/20 focus:outline-none resize-none disabled:opacity-40"
            />
          </div>
        </div>

        <!-- Tab: Bonus -->
        <div v-if="tab === 'bonus'" class="space-y-4">
          <div class="flex justify-between items-center">
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
              Daftar Barang Bonus (Hadiah)
            </p>
            <button
              v-if="canAdjust"
              @click="addBonusItem"
              class="flex items-center gap-1 px-3 py-1.5 bg-pink-50 hover:bg-pink-100 text-pink-600 text-xs font-black rounded-xl transition-all"
            >
              <Plus class="w-3.5 h-3.5" />
              Tambah Bonus
            </button>
          </div>

          <div class="space-y-3 max-h-[45vh] overflow-y-auto pr-1">
            <div v-if="bonusItems.length === 0" class="text-center py-8 bg-slate-50 rounded-2xl border border-slate-200 border-dashed">
              <Gift class="w-10 h-10 mx-auto text-slate-300 mb-2" />
              <p class="text-xs font-bold text-slate-500">Belum ada item bonus</p>
            </div>
            
            <div
              v-else
              v-for="(item, idx) in bonusItems"
              :key="idx"
              class="bg-pink-50/30 rounded-2xl p-4 border border-pink-100 relative group transition-all"
            >
              <div class="flex items-center gap-3 mb-3">
                <div class="flex-1">
                  <label class="text-[8px] font-black text-pink-400 uppercase ml-1">Pilih Inventaris</label>
                  <select
                    v-model="item.selection_id"
                    :disabled="!canAdjust"
                    class="w-full bg-white border border-pink-200 rounded-xl px-3 py-2.5 text-xs font-bold focus:ring-2 focus:ring-pink-300 outline-none disabled:opacity-50"
                  >
                    <option value="">Pilih Barang Bonus</option>
                    <option v-for="opt in combinedBonusOptions" :key="opt.id" :value="opt.id">
                      {{ opt.name }} {{ opt.type === 'inventory' ? `(Stok: ${opt.quantity} ${opt.unit})` : `(${opt.unit})` }}
                    </option>
                  </select>
                </div>
                <div class="w-24">
                  <label class="text-[8px] font-black text-pink-400 uppercase ml-1">Jumlah</label>
                  <input
                    v-model.number="item.quantity"
                    type="number"
                    min="1"
                    :disabled="!canAdjust"
                    class="w-full bg-white border border-pink-200 rounded-xl px-3 py-2.5 text-xs font-black focus:ring-2 focus:ring-pink-300 outline-none disabled:opacity-50"
                  />
                </div>
                <button
                  v-if="canAdjust"
                  @click="removeBonusItem(idx)"
                  class="p-2.5 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-xl mt-3.5 transition-colors"
                  title="Hapus bonus"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
              <div class="space-y-1">
                <label class="text-[8px] font-black text-pink-400 uppercase ml-1">Catatan</label>
                <input
                  v-model="item.notes"
                  type="text"
                  placeholder="Misal: Bonus pembelian 10 lusin"
                  :disabled="!canAdjust"
                  class="w-full bg-white border border-pink-200 rounded-xl px-3 py-2 text-xs font-medium focus:ring-2 focus:ring-pink-300 outline-none disabled:opacity-50"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Tab: Debt History -->
        <div v-if="tab === 'debt'">
          <div v-if="debtLoading" class="flex flex-col items-center py-10 text-slate-400 gap-3">
            <Loader2 class="w-8 h-8 animate-spin" />
            <p class="text-sm font-bold">Memuat riwayat piutang...</p>
          </div>
          <div
            v-else-if="debtHistory.length === 0"
            class="flex flex-col items-center py-10 text-slate-400 gap-3"
          >
            <CheckCircle2 class="w-10 h-10 text-green-400" />
            <p class="text-sm font-bold text-green-600">Tidak ada piutang outstanding</p>
            <p class="text-xs text-slate-400">
              Toko ini tidak memiliki tagihan yang belum dibayar.
            </p>
          </div>
          <div v-else class="space-y-3">
            <!-- Total debt summary -->
            <div
              class="bg-red-50 border border-red-100 rounded-2xl p-4 flex justify-between items-center"
            >
              <div>
                <p class="text-[10px] font-black text-red-400 uppercase tracking-widest">
                  Total Piutang Outstanding
                </p>
                <p class="text-xl font-black text-red-600 mt-0.5">
                  {{ formatCurrency(totalDebt) }}
                </p>
              </div>
              <AlertTriangle class="w-8 h-8 text-red-300" />
            </div>
            <!-- Each transaction -->
            <div
              v-for="tx in debtHistory"
              :key="tx.id"
              class="bg-white border border-slate-100 rounded-2xl p-4 shadow-sm"
            >
              <div class="flex justify-between items-start">
                <div>
                  <p class="font-mono text-[10px] font-bold text-slate-400">{{ tx.receipt_no }}</p>
                  <p class="font-black text-slate-900 text-sm mt-0.5">
                    {{ formatCurrency(tx.total_amount) }}
                  </p>
                  <p class="text-xs text-slate-400 mt-1">{{ formatDate(tx.transaction_date) }}</p>
                </div>
                <div class="text-right">
                  <span
                    class="text-[9px] font-black px-2 py-1 rounded-full"
                    :class="
                      tx.payment_status === 'PARTIAL'
                        ? 'bg-amber-100 text-amber-700'
                        : 'bg-red-100 text-red-600'
                    "
                  >
                    {{ tx.payment_status }}
                  </span>
                  <p class="text-xs font-black text-red-600 mt-2">
                    Sisa: {{ formatCurrency(tx.total_amount - (tx.payment_collected_amount || 0)) }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div
        class="p-5 border-t border-slate-100 bg-slate-50/50 flex gap-3 justify-between items-center"
      >
        <div class="text-xs text-slate-400 font-semibold">
          Total sebelum:
          <span class="font-black text-slate-600">{{ formatCurrency(originalTotal) }}</span>
          <span v-if="totalAfterDiscount !== originalTotal" class="ml-2 text-primary font-black">
            → {{ formatCurrency(totalAfterDiscount) }}
          </span>
        </div>
        <div class="flex gap-2">
          <button
            @click="$emit('close')"
            class="px-5 py-2.5 rounded-2xl border border-slate-200 font-bold text-sm text-slate-600 hover:bg-slate-50 transition-colors"
          >
            Batal
          </button>
          <button
            v-if="canAdjust && tab !== 'debt'"
            @click="handleSave"
            :disabled="loading"
            class="px-6 py-2.5 rounded-2xl bg-primary text-white font-bold text-sm hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            <Loader2 v-if="loading" class="w-4 h-4 animate-spin" />
            {{ loading ? 'Menyimpan...' : 'Simpan Penyesuaian' }}
          </button>
        </div>
      </div>
        </div>
      </div>
    </div>
  </div>
</template>
