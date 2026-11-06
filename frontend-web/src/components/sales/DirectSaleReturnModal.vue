<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  X,
  RotateCcw,
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  Package,
  Plus,
  Trash2,
  Receipt,
  FileText,
  Boxes,
  ArrowRight,
  Info,
  ShieldCheck,
  Building2,
  TrendingDown,
  RefreshCw
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { useSalesReturnStore } from '@/stores/salesReturn'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import { useFactoryStore } from '@/stores/factory'
import ProductSearchSelect from '@/components/sales/ProductSearchSelect.vue'
import { toast } from 'vue-sonner'

const props = defineProps<{
  show: boolean
  sale: any
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', returnData: any): void
}>()

const returnStore = useSalesReturnStore()
const masterDataStore = useMasterDataStore()
const authStore = useAuthStore()
const factoryStore = useFactoryStore()
const replacementProducts = computed(() => factoryStore.sellableProducts || [])

const isSubmitting = ref(false)
const selectedBranchId = ref('')
const returnType = ref<'POTONG_NOTA' | 'TUKAR_LANGSUNG' | 'TUKAR_PRODUK' | 'GARANSI'>('POTONG_NOTA')
const notes = ref('')
const showConfirmSummaryModal = ref(false)

interface ReturnItemForm {
  sales_item_id: string
  product_id: string
  product_name: string
  sku: string
  unit: string
  max_quantity: number
  quantity: number
  price: number
  discount_percent: number
  condition: 'GOOD' | 'DAMAGED'
  reason: string
  replacement_product_id: string
  replacement_product_name: string
  replacement_quantity: number
  replacement_price: number
}

const items = ref<ReturnItemForm[]>([])

// Populate items from sale whenever sale changes or modal opens
watch(
  () => [props.show, props.sale],
  ([show, sale]) => {
    if (show && sale) {
      notes.value = ''
      returnType.value = 'POTONG_NOTA'
      showConfirmSummaryModal.value = false
      const raw = sale.raw || sale
      const rawItems = raw.items || []

      // Resolusi Cabang dengan fallback otomatis
      if (!masterDataStore.branches || masterDataStore.branches.length === 0) {
        masterDataStore.fetchBranches()
      }
      if (replacementProducts.value.length === 0) factoryStore.fetchProducts()
      const activeFilterBranch = masterDataStore.selectedBranchId !== 'ALL_BRANCHES' && masterDataStore.selectedBranchId ? masterDataStore.selectedBranchId : ''
      const defaultBranch = masterDataStore.branches && masterDataStore.branches.length > 0 ? masterDataStore.branches[0]?.id : ''
      selectedBranchId.value = raw.branch_id || raw.branch?.id || raw.employee?.branch_id || activeFilterBranch || authStore.user?.branch_id || defaultBranch || ''

      if (rawItems.length > 0) {
        items.value = rawItems.map((item: any) => ({
          sales_item_id: item.id || '',
          product_id: item.product_id,
          product_name: item.product?.name || item.product_name || 'Produk',
          sku: item.product?.sku || '-',
          unit: item.product?.unit || item.unit || 'Pcs',
          max_quantity: item.quantity || 1,
          quantity: item.quantity || 1,
          price: item.price_at_transaction || item.price || 0,
          discount_percent: Number(item.discount_percent || 0),
          condition: 'DAMAGED',
          reason: 'Barang Rusak / Cacat',
          replacement_product_id: '',
          replacement_product_name: '',
          replacement_quantity: item.quantity || 1,
          replacement_price: 0,
        }))
      } else {
        items.value = []
      }
    }
  },
  { immediate: true },
)

const formatRupiah = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const totalReturnAmount = computed(() => {
  return items.value.reduce((sum, it) => sum + (Number(it.quantity) || 0) * (it.price || 0) *
    (returnType.value === 'TUKAR_PRODUK' ? 1 - it.discount_percent / 100 : 1), 0)
})
const totalReplacementAmount = computed(() => items.value.reduce((sum, it) => sum + (Number(it.replacement_quantity) || 0) * (Number(it.replacement_price) || 0), 0))

const updateReplacementProduct = (item: ReturnItemForm) => {
  const product = replacementProducts.value.find((p) => p.id === item.replacement_product_id)
  item.replacement_product_name = product?.name || ''
  item.replacement_price = Number(product?.selling_price || 0)
}

const initialInvoiceTotal = computed(() => {
  const raw = props.sale?.raw || props.sale
  return Number(raw?.total_amount || 0)
})
const paidInvoiceAmount = computed(() => Number((props.sale?.raw || props.sale)?.paid_amount || 0))

const estimatedFinalInvoiceTotal = computed(() => {
  if (returnType.value === 'POTONG_NOTA') {
    return Math.max(0, initialInvoiceTotal.value - totalReturnAmount.value)
  }
  if (returnType.value === 'TUKAR_PRODUK') {
    return initialInvoiceTotal.value - totalReturnAmount.value + totalReplacementAmount.value
  }
  return initialInvoiceTotal.value
})

const customerDisplayName = computed(() => {
  if (!props.sale) return '-'
  const raw = props.sale.raw || props.sale
  if (raw.buyer_type === 'EMPLOYEE' && raw.buyer_employee) {
    const fn = raw.buyer_employee.first_name || ''
    const ln = raw.buyer_employee.last_name || ''
    return `${fn} ${ln}`.trim() + ' (Karyawan)'
  }
  if (raw.buyer_name && typeof raw.buyer_name === 'string') {
    return raw.buyer_name
  }
  if (raw.store) {
    if (typeof raw.store === 'string') return raw.store
    if (typeof raw.store === 'object' && raw.store.name) return raw.store.name
  }
  if (props.sale.store) {
    if (typeof props.sale.store === 'string') return props.sale.store
    if (typeof props.sale.store === 'object' && props.sale.store.name) return props.sale.store.name
  }
  if (raw.sales_channel === 'DIRECT_FACTORY') {
    return 'Pelanggan Pabrik'
  }
  return '-'
})

const selectedBranchDisplayName = computed(() => {
  if (selectedBranchId.value) {
    const found = masterDataStore.branches?.find((b: any) => b.id === selectedBranchId.value)
    if (found) return found.name
  }
  const raw = props.sale?.raw || props.sale
  if (raw?.branch?.name) return raw.branch.name
  if (props.sale?.branch?.name) return props.sale.branch.name
  return 'Cabang Utama'
})

const isFormValid = computed(() => {
  if (!selectedBranchId.value) return false
  if (items.value.length === 0) return false
  if (returnType.value === 'TUKAR_PRODUK' && estimatedFinalInvoiceTotal.value < 0) return false
  return items.value.every((it) => it.product_id && it.quantity > 0 && it.quantity <= it.max_quantity &&
    (returnType.value !== 'TUKAR_PRODUK' || (it.replacement_product_id && it.replacement_product_id !== it.product_id && it.replacement_quantity > 0 && it.replacement_price > 0)))
})

const handleRemoveItem = (index: number) => {
  items.value.splice(index, 1)
}

const handleClose = () => {
  showConfirmSummaryModal.value = false
  emit('close')
}

const openConfirmSummaryModal = () => {
  if (!props.sale) return
  if (!selectedBranchId.value) {
    toast.error('Harap pilih cabang tujuan retur.')
    return
  }
  if (!isFormValid.value) {
    toast.error('Pastikan kuantitas retur valid dan tidak melebihi kuantitas pembelian.')
    return
  }
  showConfirmSummaryModal.value = true
}

const executeSubmit = async () => {
  if (!props.sale) return
  const raw = props.sale.raw || props.sale

  isSubmitting.value = true
  try {
    const payload = {
      transaction_id: raw.id,
      branch_id: selectedBranchId.value || raw.branch_id,
      return_type: returnType.value,
      notes: notes.value,
      items: items.value.map((it) => ({
        ...(it.sales_item_id ? { sales_item_id: it.sales_item_id } : {}),
        product_id: it.product_id,
        quantity: Number(it.quantity),
        price: it.price,
        condition: it.condition,
        reason: it.reason,
        ...(returnType.value === 'TUKAR_PRODUK' ? {
          replacement_product_id: it.replacement_product_id,
          replacement_quantity: Number(it.replacement_quantity),
          replacement_price: it.replacement_price,
        } : {}),
      })),
    }

    const created = await returnStore.createReturn(payload)
    toast.success(returnType.value === 'TUKAR_PRODUK'
      ? 'Tukar produk diajukan dan faktur langsung diperbarui. Stok menunggu persetujuan gudang.'
      : 'Pengajuan retur penjualan berhasil dibuat dan menunggu verifikasi gudang.')
    showConfirmSummaryModal.value = false
    emit('created', created)
    handleClose()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal mengajukan retur penjualan')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4 overflow-y-auto"
  >
    <div
      class="bg-white rounded-[2rem] shadow-2xl border border-slate-100 w-full max-w-3xl overflow-hidden animate-in fade-in zoom-in-95 duration-200"
    >
      <!-- Modal Header -->
      <div class="px-8 py-6 bg-slate-900 text-white flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-red-500/20 text-red-400 rounded-xl">
            <RotateCcw class="w-6 h-6" />
          </div>
          <div>
            <h2 class="text-xl font-black">Pelaporan Retur Faktur Penjualan</h2>
            <p class="text-xs text-slate-300 mt-0.5">
              Faktur No: <span class="font-bold text-white">{{ sale?.receipt_no || sale?.raw?.receipt_no || '-' }}</span> • 
              Pelanggan: <span class="font-bold text-white">{{ customerDisplayName }}</span>
            </p>
          </div>
        </div>
        <button
          @click="handleClose"
          class="p-2 text-slate-400 hover:text-white rounded-full hover:bg-slate-800 transition-colors"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-8 space-y-6 max-h-[75vh] overflow-y-auto">
        <!-- Return Type & Notes -->
        <div class="space-y-3">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                Cabang Tujuan Retur *
              </label>
              <select
                v-model="selectedBranchId"
                class="w-full px-4 py-2.5 rounded-xl border border-slate-200 bg-slate-50 font-bold text-sm text-slate-800 focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500 transition-all cursor-pointer"
              >
                <option value="" disabled>-- Pilih Cabang --</option>
                <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">
                  {{ b.name }}
                </option>
              </select>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                Jenis Retur *
              </label>
              <select
                v-model="returnType"
                class="w-full px-4 py-2.5 rounded-xl border border-slate-200 bg-slate-50 font-bold text-sm text-slate-800 focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500 transition-all cursor-pointer"
              >
                <option value="POTONG_NOTA">1. Potong Tagihan / Kredit Nota (POTONG_NOTA)</option>
                <option value="TUKAR_LANGSUNG">2. Tukar Barang Sejenis Langsung (TUKAR_LANGSUNG)</option>
                <option value="TUKAR_PRODUK">3. Ganti dengan Barang Lain (TUKAR_PRODUK)</option>
                <option value="GARANSI">4. Klaim Garansi / Retur Cacat (GARANSI)</option>
              </select>
            </div>

            <div>
              <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                Catatan / Keterangan
              </label>
              <input
                v-model="notes"
                type="text"
                placeholder="Contoh: Barang retur karena bocor kemasan..."
                class="w-full px-4 py-2.5 rounded-xl border border-slate-200 bg-slate-50 text-sm text-slate-800 focus:outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500 transition-all"
              />
            </div>
          </div>

          <!-- Dynamic Contextual Banner explaining Return Mechanism -->
          <div
            v-if="returnType === 'POTONG_NOTA'"
            class="bg-blue-50 border border-blue-200 p-3 rounded-xl flex items-start gap-2.5 text-xs text-blue-900"
          >
            <Info class="w-4 h-4 text-blue-600 shrink-0 mt-0.5" />
            <p class="leading-relaxed">
              <strong>Mekanisme Potong Nota:</strong> Kuantitas barang pada faktur dan total nilai tagihan akan <strong>berkurang</strong> sesuai barang yang dikembalikan. Fisik barang yang diretur akan masuk kembali ke stok gudang cabang.
            </p>
          </div>
          <div
            v-else-if="returnType === 'TUKAR_LANGSUNG'"
            class="bg-emerald-50 border border-emerald-200 p-3 rounded-xl flex items-start gap-2.5 text-xs text-emerald-900"
          >
            <RefreshCw class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5" />
            <p class="leading-relaxed">
              <strong>Mekanisme Tukar Barang:</strong> Kuantitas dan total tagihan faktur <strong>TETAP SAMA</strong> (tidak berkurang). Barang rusak masuk ke gudang (<span class="font-bold">IN</span>), dan barang pengganti bagus keluar dari gudang (<span class="font-bold">OUT</span>).
            </p>
          </div>
          <div
            v-else-if="returnType === 'TUKAR_PRODUK'"
            class="bg-amber-50 border border-amber-200 p-3 rounded-xl flex items-start gap-2.5 text-xs text-amber-900"
          >
            <Info class="w-4 h-4 shrink-0 mt-0.5" />
            <p>Barang pada faktur diganti produk lain dari katalog dan total faktur langsung menyesuaikan saat pengajuan dibuat. Stok barang asal dan pengganti berubah setelah Admin Gudang menyetujui pertukaran.</p>
          </div>
          <div
            v-else-if="returnType === 'GARANSI'"
            class="bg-purple-50 border border-purple-200 p-3 rounded-xl flex items-start gap-2.5 text-xs text-purple-900"
          >
            <ShieldCheck class="w-4 h-4 text-purple-600 shrink-0 mt-0.5" />
            <p class="leading-relaxed">
              <strong>Mekanisme Klaim Garansi:</strong> Kuantitas dan tagihan faktur <strong>TETAP SAMA</strong>. Unit cacat klaim masuk ke gudang (<span class="font-bold">IN</span>), dan unit pengganti baru dikeluarkan dari gudang (<span class="font-bold">OUT</span>).
            </p>
          </div>
        </div>

        <!-- Return Items Table -->
        <div>
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-sm font-black text-slate-900 flex items-center gap-2">
              <Package class="w-4 h-4 text-red-600" />
              Item Barang yang Diretur
            </h3>
            <span class="text-xs text-slate-500 font-medium">
              Sesuaikan kuantitas & kondisi barang yang dikembalikan
            </span>
          </div>

          <div v-if="items.length === 0" class="text-center py-8 bg-slate-50 rounded-2xl border border-dashed border-slate-200">
            <AlertCircle class="w-8 h-8 text-slate-300 mx-auto mb-2" />
            <p class="text-xs text-slate-500 font-bold">Tidak ada item produk pada faktur ini.</p>
          </div>

          <div v-else class="space-y-3">
            <div
              v-for="(item, idx) in items"
              :key="item.product_id"
              class="p-4 rounded-2xl border border-slate-200 bg-slate-50/50 hover:bg-slate-50 transition-all space-y-3"
            >
              <div class="flex items-start justify-between">
                <div>
                  <h4 class="font-bold text-slate-900 text-sm">{{ item.product_name }}</h4>
                  <p class="text-xs text-slate-500">
                    SKU: <span class="font-mono text-slate-600">{{ item.sku }}</span> •
                    Harga Beli: <span class="font-semibold text-slate-700">{{ formatRupiah(item.price) }}</span> / {{ item.unit }} • 
                    Maksimal Retur: <span class="font-bold text-slate-800">{{ item.max_quantity }} {{ item.unit }}</span>
                  </p>
                </div>
                <button
                  @click="handleRemoveItem(idx)"
                  class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  title="Hapus item dari retur"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 border-t border-slate-200/60">
                <!-- Quantity -->
                <div>
                  <label class="block text-[11px] font-bold text-slate-600 mb-1">
                    Jumlah Retur ({{ item.unit }}) *
                  </label>
                  <input
                    v-model.number="item.quantity"
                    type="number"
                    min="1"
                    :max="item.max_quantity"
                    class="w-full px-3 py-1.5 rounded-lg border border-slate-300 bg-white font-black text-sm text-slate-900 focus:outline-none focus:border-red-500"
                  />
                </div>

                <!-- Condition -->
                <div>
                  <label class="block text-[11px] font-bold text-slate-600 mb-1">
                    Kondisi Fisik Barang
                  </label>
                  <select
                    v-model="item.condition"
                    class="w-full px-3 py-1.5 rounded-lg border border-slate-300 bg-white font-bold text-xs text-slate-800 focus:outline-none focus:border-red-500"
                  >
                    <option value="DAMAGED">RUSAK / CACAT (Masuk Karantina)</option>
                    <option value="GOOD">BAGUS / BAIK (Bisa Dijual Lagi)</option>
                  </select>
                </div>

                <!-- Reason -->
                <div>
                  <label class="block text-[11px] font-bold text-slate-600 mb-1">
                    Alasan Retur
                  </label>
                  <input
                    v-model="item.reason"
                    type="text"
                    placeholder="Alasan retur..."
                    class="w-full px-3 py-1.5 rounded-lg border border-slate-300 bg-white text-xs text-slate-800 focus:outline-none focus:border-red-500"
                  />
                </div>
              </div>
              <div v-if="returnType === 'TUKAR_PRODUK'" class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-3 border-t border-amber-200">
                <div>
                  <label class="block text-[11px] font-bold text-amber-800 mb-1">Barang Pengganti *</label>
                  <ProductSearchSelect
                    v-model="item.replacement_product_id"
                    :products="replacementProducts"
                    :show-catalog-button="false"
                    placeholder="Pilih barang lain dari katalog..."
                    @change="() => updateReplacementProduct(item)"
                  />
                </div>
                <div>
                  <label class="block text-[11px] font-bold text-amber-800 mb-1">Jumlah Pengganti *</label>
                  <input v-model.number="item.replacement_quantity" type="number" min="1" class="w-full px-3 py-2 rounded-lg border border-amber-300 bg-white text-sm font-bold" />
                </div>
                <p class="sm:col-span-2 text-xs text-amber-800">Harga barang pengganti: {{ formatRupiah(item.replacement_price) }} / unit · Total: {{ formatRupiah(item.replacement_quantity * item.replacement_price) }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Summary Banner -->
        <div class="bg-gradient-to-r from-red-50 to-orange-50 border border-red-100 p-5 rounded-2xl flex items-center justify-between">
          <div>
            <p class="text-xs font-bold text-slate-500 uppercase tracking-wider">Total Nilai Barang Retur</p>
            <h3 class="text-2xl font-black text-red-600">{{ formatRupiah(totalReturnAmount) }}</h3>
            <p v-if="returnType === 'TUKAR_PRODUK'" class="text-xs font-bold text-amber-800 mt-1">Nilai pengganti {{ formatRupiah(totalReplacementAmount) }} · Selisih {{ formatRupiah(totalReplacementAmount - totalReturnAmount) }}</p>
          </div>
          <div class="text-right text-xs text-slate-500">
            <span class="font-bold text-slate-700">{{ items.length }} Item Terpilih</span>
            <p class="mt-0.5 text-[11px]">Klik lanjut untuk melihat ringkasan mutasi stok gudang & dampak faktur.</p>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-8 py-5 bg-slate-50 border-t border-slate-100 flex items-center justify-end gap-3">
        <Button
          @click="handleClose"
          type="button"
          variant="outline"
          class="rounded-xl px-5 font-bold text-slate-700"
        >
          Batal
        </Button>
        <Button
          @click="openConfirmSummaryModal"
          :disabled="isSubmitting || !isFormValid"
          class="rounded-xl px-6 font-black bg-red-600 hover:bg-red-700 text-white shadow-lg shadow-red-500/20"
        >
          <RotateCcw class="w-4 h-4 mr-2" />
          Lanjut ke Ringkasan Retur
        </Button>
      </div>
    </div>

    <!-- POP-UP KONFIRMASI & RINGKASAN PERGERAKAN STOK GUDANG -->
    <div
      v-if="showConfirmSummaryModal"
      class="fixed inset-0 z-70 flex items-center justify-center bg-black/70 backdrop-blur-xs p-4 overflow-y-auto"
    >
      <div
        class="bg-white rounded-2xl shadow-2xl border border-slate-200 w-full max-w-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150 flex flex-col max-h-[90vh]"
      >
        <!-- Modal Summary Header -->
        <div class="px-6 py-4 bg-amber-50 border-b border-amber-200 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="p-2.5 bg-amber-500/10 text-amber-600 rounded-xl font-bold">
              <Boxes class="w-6 h-6" />
            </div>
            <div>
              <h3 class="font-bold text-slate-900 text-base">Konfirmasi & Ringkasan Pergerakan Stok Retur</h3>
              <p class="text-xs text-slate-500">Periksa detail barang yang bergerak masuk/keluar gudang & dampak faktur</p>
            </div>
          </div>
          <button
            @click="showConfirmSummaryModal = false"
            :disabled="isSubmitting"
            class="text-slate-400 hover:text-slate-600 p-1 rounded-lg"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Modal Summary Body -->
        <div class="p-6 overflow-y-auto space-y-4 text-xs text-slate-700">
          <!-- Metadata Card -->
          <div class="grid grid-cols-2 gap-3 bg-slate-50 p-3.5 rounded-xl border border-slate-200">
            <div>
              <span class="text-slate-500 text-[10px] uppercase font-bold tracking-wider block">No. Faktur Penjualan</span>
              <span class="font-black text-slate-900 text-sm block mt-0.5">{{ sale?.receipt_no || sale?.raw?.receipt_no || '-' }}</span>
              <span class="text-[10px] text-slate-500">Cabang: {{ selectedBranchDisplayName }}</span>
            </div>
            <div>
              <span class="text-slate-500 text-[10px] uppercase font-bold tracking-wider block">Jenis Retur Diajukan</span>
              <span class="inline-flex items-center font-black text-xs px-2 py-0.5 rounded-md mt-0.5" :class="{
                'bg-blue-100 text-blue-800': returnType === 'POTONG_NOTA',
                'bg-emerald-100 text-emerald-800': returnType === 'TUKAR_LANGSUNG',
                'bg-amber-100 text-amber-800': returnType === 'TUKAR_PRODUK',
                'bg-purple-100 text-purple-800': returnType === 'GARANSI',
              }">
                {{ returnType }}
              </span>
              <span class="text-[10px] text-slate-500 block mt-0.5">Pembeli: {{ customerDisplayName }}</span>
            </div>
          </div>

          <!-- Explanation Banner based on return type -->
          <div
            v-if="returnType === 'POTONG_NOTA'"
            class="bg-blue-50 border border-blue-200 p-3 rounded-xl flex items-start gap-2.5 text-blue-900"
          >
            <Info class="w-4 h-4 text-blue-600 shrink-0 mt-0.5" />
            <div class="space-y-0.5 text-[11px]">
              <p class="font-bold">Dampak Faktur & Gudang (POTONG NOTA):</p>
              <p>• Kuantitas barang pada faktur akan <strong>berkurang</strong> sesuai item retur.</p>
              <p>• Total tagihan faktur akan <strong>berkurang Rp {{ formatRupiah(totalReturnAmount) }}</strong>.</p>
              <p>• Fisik barang yang diretur <strong>MASUK (IN)</strong> ke gudang cabang.</p>
            </div>
          </div>
          <div
            v-else-if="returnType === 'TUKAR_LANGSUNG'"
            class="bg-emerald-50 border border-emerald-200 p-3 rounded-xl flex items-start gap-2.5 text-emerald-900"
          >
            <RefreshCw class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5" />
            <div class="space-y-0.5 text-[11px]">
              <p class="font-bold">Dampak Faktur & Gudang (TUKAR BARANG SEJENIS):</p>
              <p>• Kuantitas barang & nilai tagihan faktur <strong>TETAP SAMA (Tidak Berkurang)</strong>.</p>
              <p>• Barang rusak dikembalikan pembeli <strong>MASUK (IN)</strong> ke gudang karantina.</p>
              <p>• Unit baru pengganti yang bagus <strong>KELUAR (OUT)</strong> dari gudang cabang.</p>
            </div>
          </div>
          <div
            v-else-if="returnType === 'TUKAR_PRODUK'"
            class="bg-amber-50 border border-amber-200 p-3 rounded-xl flex items-start gap-2.5 text-amber-900"
          >
            <Info class="w-4 h-4 shrink-0 mt-0.5" />
            <p>Faktur langsung berubah sebesar {{ formatRupiah(totalReplacementAmount - totalReturnAmount) }} setelah pengajuan. Barang retur masuk dan pengganti keluar gudang setelah persetujuan Admin Gudang.</p>
          </div>
          <div
            v-else-if="returnType === 'GARANSI'"
            class="bg-purple-50 border border-purple-200 p-3 rounded-xl flex items-start gap-2.5 text-purple-900"
          >
            <ShieldCheck class="w-4 h-4 text-purple-600 shrink-0 mt-0.5" />
            <div class="space-y-0.5 text-[11px]">
              <p class="font-bold">Dampak Faktur & Gudang (KLAIM GARANSI):</p>
              <p>• Kuantitas barang & nilai tagihan faktur <strong>TETAP SAMA (Tidak Berkurang)</strong>.</p>
              <p>• Unit rusak klaim <strong>MASUK (IN)</strong> ke gudang.</p>
              <p>• Unit baru pengganti garansi <strong>KELUAR (OUT)</strong> dari gudang cabang.</p>
            </div>
          </div>

          <!-- Stock Movement Table -->
          <div class="space-y-1.5">
            <span class="font-bold text-slate-800 flex items-center gap-1.5">
              <Boxes class="w-4 h-4 text-primary" /> Detail Barang & Mutasi Fisik Gudang:
            </span>

            <div class="border border-slate-200 rounded-xl overflow-hidden shadow-2xs">
              <table class="w-full text-left text-xs border-collapse">
                <thead>
                  <tr class="bg-slate-100 border-b border-slate-200 text-slate-700 font-bold">
                    <th class="p-2.5 w-8 text-center">No</th>
                    <th class="p-2.5">Produk & SKU</th>
                    <th class="p-2.5 w-24 text-center">Stok Masuk (IN)</th>
                    <th class="p-2.5 w-28 text-center">Stok Keluar (OUT)</th>
                    <th class="p-2.5 w-28 text-right">Dampak Nota</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 bg-white">
                  <tr v-for="(item, idx) in items" :key="idx" class="hover:bg-slate-50/50">
                    <td class="p-2.5 text-center text-slate-400 font-medium">{{ idx + 1 }}</td>
                    <td class="p-2.5 font-medium text-slate-800">
                      <div class="font-bold text-slate-900">{{ item.product_name }}</div>
                      <div class="text-[10px] text-slate-400 font-mono">SKU: {{ item.sku }}</div>
                    </td>
                    <td class="p-2.5 text-center font-bold text-emerald-700 bg-emerald-50/30">
                      +{{ item.quantity }} {{ item.unit }}
                      <span class="block text-[9px] font-normal" :class="item.condition === 'DAMAGED' ? 'text-rose-600' : 'text-emerald-600'">
                        ({{ item.condition === 'DAMAGED' ? 'RUSAK' : 'BAGUS' }})
                      </span>
                    </td>
                    <td class="p-2.5 text-center font-bold">
                      <template v-if="returnType === 'POTONG_NOTA'">
                        <span class="text-slate-400 font-medium">-</span>
                      </template>
                      <template v-else>
                        <span class="text-rose-600 bg-rose-50/40 px-1.5 py-0.5 rounded">
                          -{{ returnType === 'TUKAR_PRODUK' ? item.replacement_quantity : item.quantity }} {{ item.unit }}
                        </span>
                        <span class="block text-[9px] text-emerald-600 font-normal">
                          {{ returnType === 'TUKAR_PRODUK' ? item.replacement_product_name : 'UNIT BAGUS' }}
                        </span>
                      </template>
                    </td>
                    <td class="p-2.5 text-right font-bold">
                      <span v-if="returnType === 'POTONG_NOTA'" class="text-rose-600">
                        -{{ formatRupiah(item.quantity * item.price) }}
                      </span>
                      <span v-else-if="returnType === 'TUKAR_PRODUK'" class="text-amber-700">
                        {{ formatRupiah(item.replacement_quantity * item.replacement_price - item.quantity * item.price * (1 - item.discount_percent / 100)) }}
                      </span>
                      <span v-else class="text-slate-500 font-medium text-[11px]">
                        Rp 0 (Ditukar)
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Financial Summary Comparison -->
          <div class="bg-slate-50 p-3.5 rounded-xl border border-slate-200 space-y-1.5 text-xs">
            <div class="flex justify-between items-center text-slate-600">
              <span>Nilai Tagihan Faktur Awal:</span>
              <span class="font-bold text-slate-800">{{ formatRupiah(initialInvoiceTotal) }}</span>
            </div>
            <div v-if="returnType === 'POTONG_NOTA'" class="flex justify-between items-center text-rose-600">
              <span>Potongan Nilai Retur:</span>
              <span class="font-bold">- {{ formatRupiah(totalReturnAmount) }}</span>
            </div>
            <div v-if="returnType === 'TUKAR_PRODUK'" class="flex justify-between items-center text-amber-700">
              <span>Selisih Harga Barang Pengganti:</span>
              <span class="font-bold">{{ formatRupiah(totalReplacementAmount - totalReturnAmount) }}</span>
            </div>
            <div class="flex justify-between items-center pt-1.5 border-t border-slate-200 font-black text-sm">
              <span class="text-slate-900 uppercase">Estimasi Tagihan Akhir Faktur:</span>
              <span class="text-primary text-base">{{ formatRupiah(estimatedFinalInvoiceTotal) }}</span>
            </div>
            <template v-if="returnType === 'TUKAR_PRODUK' && paidInvoiceAmount > 0">
              <div class="flex justify-between items-center text-slate-600">
                <span>Pembayaran yang sudah tercatat:</span>
                <span class="font-bold">{{ formatRupiah(paidInvoiceAmount) }}</span>
              </div>
              <div class="flex justify-between items-center font-bold" :class="paidInvoiceAmount > estimatedFinalInvoiceTotal ? 'text-amber-700' : 'text-slate-700'">
                <span>{{ paidInvoiceAmount > estimatedFinalInvoiceTotal ? 'Kelebihan bayar, perlu diselesaikan kasir:' : 'Sisa tagihan setelah penukaran:' }}</span>
                <span>{{ formatRupiah(Math.abs(estimatedFinalInvoiceTotal - paidInvoiceAmount)) }}</span>
              </div>
            </template>
          </div>
        </div>

        <!-- Modal Summary Footer -->
        <div class="px-6 py-4 border-t border-slate-200 bg-slate-50 flex items-center justify-end gap-3">
          <Button
            type="button"
            variant="outline"
            :disabled="isSubmitting"
            @click="showConfirmSummaryModal = false"
            class="text-xs font-semibold px-4 py-2"
          >
            Batal / Periksa Kembali
          </Button>
          <Button
            type="button"
            :disabled="isSubmitting"
            @click="executeSubmit"
            class="text-xs font-bold px-5 py-2 bg-red-600 hover:bg-red-700 text-white gap-2 shadow-sm"
          >
            <span v-if="isSubmitting" class="animate-spin">⏳</span>
            <span v-else>Ya, Konfirmasi & Ajukan Retur</span>
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

