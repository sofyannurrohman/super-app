<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  X,
  RotateCcw,
  Truck,
  Building2,
  Calendar,
  Search,
  Package,
  Plus,
  Minus,
  Trash2,
  CheckCircle2,
  AlertTriangle,
  Info,
  ShieldCheck,
  TrendingDown,
  RefreshCw,
  QrCode
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { useSalesReturnStore } from '@/stores/salesReturn'
import { useMasterDataStore } from '@/stores/masterData'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'

const props = defineProps<{
  show: boolean
  branchId: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', returnData: any): void
}>()

const returnStore = useSalesReturnStore()
const masterDataStore = useMasterDataStore()

// State
const isLoadingBatches = ref(false)
const batches = ref<any[]>([])
const batchSearchQuery = ref('')
const selectedBatch = ref<any>(null)

const selectedSalesOrder = ref<any>(null)
const returnType = ref<'POTONG_NOTA' | 'TUKAR_LANGSUNG' | 'TUKAR_PRODUK' | 'GARANSI'>('POTONG_NOTA')
const autoApproveWarehouse = ref(true) // Default true karena Admin Gudang yang memegang fisik
const notes = ref('')
const isSubmitting = ref(false)

interface ReturnItemInput {
  product_id: string
  product_name: string
  sku: string
  unit: string
  max_quantity: number
  quantity: number
  price: number
  condition: 'GOOD' | 'DAMAGED'
  reason: string
}

const items = ref<ReturnItemInput[]>([])

// Fetch active & relevant batches for this branch
async function fetchDeliveryBatches() {
  if (!props.branchId) return
  isLoadingBatches.value = true
  try {
    const params = props.branchId && props.branchId !== 'ALL_BRANCHES' && props.branchId !== 'ALL'
      ? `?branch_id=${props.branchId}`
      : ''
    const res = await apiClient.get(`/delivery/batches/warehouse${params}`)
    batches.value = res.data?.data || []
  } catch (err: any) {
    console.error('Failed to load delivery batches', err)
    toast.error('Gagal memuat daftar pengiriman / surat jalan')
  } finally {
    isLoadingBatches.value = false
  }
}

// Watch modal visibility
watch(
  () => props.show,
  (val) => {
    if (val) {
      resetForm()
      fetchDeliveryBatches()
    }
  },
  { immediate: true }
)

function resetForm() {
  selectedBatch.value = null
  selectedSalesOrder.value = null
  batchSearchQuery.value = ''
  returnType.value = 'POTONG_NOTA'
  autoApproveWarehouse.value = true
  notes.value = ''
  items.value = []
}

// Filtered batches by search (No SJ, Driver, Vehicle)
const filteredBatches = computed(() => {
  const q = batchSearchQuery.value.trim().toLowerCase()
  if (!q) return batches.value

  return batches.value.filter((b: any) => {
    const doNo = (b.delivery_order_no || '').toLowerCase()
    const driverName = (b.driver?.first_name || b.daily_driver?.name || '').toLowerCase()
    const plate = (b.vehicle?.plate_number || '').toLowerCase()
    const hasMatchingStore = b.items?.some((it: any) => {
      const storeName = (it.sales_order?.store?.name || it.sales_transaction?.store?.name || '').toLowerCase()
      const soNum = (it.sales_order?.so_number || '').toLowerCase()
      return storeName.includes(q) || soNum.includes(q)
    })
    return doNo.includes(q) || driverName.includes(q) || plate.includes(q) || hasMatchingStore
  })
})

// Stores / SOs inside the selected batch
const batchSalesOrders = computed(() => {
  if (!selectedBatch.value || !selectedBatch.value.items) return []

  const orderMap = new Map<string, any>()
  selectedBatch.value.items.forEach((it: any) => {
    if (it.sales_order) {
      orderMap.set(it.sales_order.id, it.sales_order)
    }
  })

  return Array.from(orderMap.values())
})

// When a batch is picked
function selectBatch(batch: any) {
  selectedBatch.value = batch
  selectedSalesOrder.value = null
  items.value = []

  // Auto-select if there is only 1 sales order in this batch
  if (batchSalesOrders.value.length === 1) {
    selectSalesOrder(batchSalesOrders.value[0])
  }
}

// When a specific store / SO is picked
function selectSalesOrder(so: any) {
  selectedSalesOrder.value = so

  // Map products inside this SalesOrder to return items
  const soItems = so.items || []
  if (soItems.length > 0) {
    items.value = soItems.map((item: any) => {
      const orderedQty = item.ordered_quantity || item.quantity || 1
      const alreadyReturned = 0 // can be populated if needed
      const maxQty = Math.max(1, orderedQty - alreadyReturned)

      return {
        product_id: item.product_id,
        product_name: item.product?.name || item.product_name || 'Produk',
        sku: item.product?.sku || '-',
        unit: item.unit || item.product?.unit || 'Pcs',
        max_quantity: maxQty,
        quantity: 0, // Admin sets the quantity to return
        price: item.price || 0,
        condition: 'DAMAGED',
        reason: 'RUSAK',
      }
    })
  } else {
    items.value = []
  }
}

const activeReturnItems = computed(() => {
  return items.value.filter((i) => i.quantity > 0)
})

const totalReturnAmount = computed(() => {
  return activeReturnItems.value.reduce((sum, it) => sum + it.quantity * it.price, 0)
})

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const formatDate = (d: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
    : '-'

// Submit Return
async function handleSubmitReturn() {
  if (!selectedSalesOrder.value) {
    toast.error('Pilih Surat Jalan dan Toko terlebih dahulu')
    return
  }

  if (activeReturnItems.value.length === 0) {
    toast.error('Tentukan minimal 1 produk dengan jumlah retur lebih dari 0')
    return
  }

  // Validate quantities
  for (const item of activeReturnItems.value) {
    if (item.quantity > item.max_quantity) {
      toast.error(`Jumlah retur untuk ${item.product_name} (${item.quantity}) melebihi kuantitas pengiriman (${item.max_quantity})`)
      return
    }
  }

  isSubmitting.value = true
  try {
    const payload = {
      so_id: selectedSalesOrder.value.id,
      branch_id: selectedSalesOrder.value.branch_id || props.branchId,
      delivery_batch_id: selectedBatch.value?.id || undefined,
      return_type: returnType.value,
      auto_approve_warehouse: autoApproveWarehouse.value,
      notes: notes.value ? notes.value.trim() : `Retur pengiriman ${selectedBatch.value?.delivery_order_no || ''}`,
      items: activeReturnItems.value.map((i) => ({
        product_id: i.product_id,
        quantity: Number(i.quantity),
        price: Number(i.price),
        condition: i.condition,
        reason: i.reason,
      })),
    }

    const created = await returnStore.createReturn(payload)
    toast.success(
      autoApproveWarehouse.value
        ? 'Retur pengiriman berhasil dibuat dan fisik barang langsung diterima ke stok gudang!'
        : 'Retur pengiriman berhasil diajukan! Menunggu ACC pemeriksaan fisik di gudang.'
    )
    emit('created', created)
    emit('close')
  } catch (err: any) {
    console.error('Submit return error:', err)
    toast.error(err.response?.data?.error || err.message || 'Gagal menyimpan retur pengiriman')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="show"
        class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm"
        @click.self="emit('close')"
      >
        <div class="bg-white rounded-[2.5rem] shadow-2xl w-full max-w-4xl max-h-[92vh] flex flex-col overflow-hidden border border-slate-100 animate-in fade-in zoom-in-95 duration-200">
          
          <!-- Header -->
          <div class="flex items-center justify-between px-8 py-6 border-b border-slate-100 bg-slate-50/70">
            <div class="flex items-center gap-3">
              <div class="w-12 h-12 rounded-2xl bg-rose-50 text-rose-600 flex items-center justify-center border border-rose-100 shadow-sm">
                <RotateCcw class="w-6 h-6" />
              </div>
              <div>
                <h2 class="text-xl font-extrabold text-slate-900 flex items-center gap-2">
                  Input Retur Pengiriman (Armada / SJ)
                </h2>
                <p class="text-xs text-slate-500 font-medium">
                  Catat pengembalian barang dari armada pengiriman toko secara fleksibel.
                </p>
              </div>
            </div>
            <button
              @click="emit('close')"
              class="p-2.5 rounded-2xl hover:bg-white text-slate-400 hover:text-slate-600 border border-transparent hover:border-slate-200 transition-all"
            >
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Body -->
          <div class="overflow-y-auto flex-1 p-8 space-y-6">

            <!-- Step 1: Pilih Surat Jalan / Batch Pengiriman -->
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <label class="text-xs font-black uppercase tracking-wider text-slate-600 flex items-center gap-1.5">
                  <Truck class="w-4 h-4 text-rose-500" />
                  1. Pilih Surat Jalan / Pengiriman Armada
                </label>
                <span v-if="selectedBatch" class="text-xs font-bold text-rose-600 bg-rose-50 px-2.5 py-1 rounded-xl border border-rose-100 flex items-center gap-1">
                  <CheckCircle2 class="w-3.5 h-3.5" /> {{ selectedBatch.delivery_order_no || 'SJ Terpilih' }}
                </span>
              </div>

              <!-- Search Batch Input -->
              <div class="relative">
                <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <input
                  v-model="batchSearchQuery"
                  type="text"
                  placeholder="Ketik No. Surat Jalan (SJ-...), Nama Driver, Plat Nomor, atau Nama Toko..."
                  class="w-full pl-10 pr-4 py-2.5 rounded-xl border border-slate-200 text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 bg-white shadow-sm"
                />
              </div>

              <!-- Batch List Pills / Mini Cards -->
              <div v-if="isLoadingBatches" class="py-6 text-center text-slate-400 flex items-center justify-center gap-2 text-xs font-bold">
                <RefreshCw class="w-4 h-4 animate-spin text-rose-500" /> Memuat data pengiriman cabang...
              </div>
              <div v-else-if="filteredBatches.length === 0" class="py-6 text-center text-slate-400 text-xs font-medium bg-slate-50 rounded-2xl border border-dashed border-slate-200">
                Tidak ada data surat jalan yang cocok dengan pencarian.
              </div>
              <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-3 max-h-48 overflow-y-auto p-1">
                <div
                  v-for="b in filteredBatches"
                  :key="b.id"
                  @click="selectBatch(b)"
                  class="p-3.5 rounded-2xl border transition-all cursor-pointer flex flex-col justify-between gap-2"
                  :class="selectedBatch?.id === b.id ? 'border-rose-500 bg-rose-50/40 shadow-sm ring-1 ring-rose-500' : 'border-slate-200 hover:border-slate-300 bg-white'"
                >
                  <div class="flex items-start justify-between">
                    <div>
                      <span class="font-extrabold text-sm text-slate-900 flex items-center gap-1.5">
                        <Truck class="w-4 h-4 text-slate-500" /> {{ b.delivery_order_no || 'SJ Tanpa Nomor' }}
                      </span>
                      <p class="text-[11px] text-slate-500 font-medium mt-0.5">
                        Supir: <span class="font-bold text-slate-700">{{ b.driver?.first_name || b.daily_driver?.name || 'Driver' }}</span> • Plat: <span class="font-bold text-slate-700">{{ b.vehicle?.plate_number || '-' }}</span>
                      </p>
                    </div>
                    <span
                      class="text-[9px] font-black uppercase px-2 py-0.5 rounded-md border"
                      :class="{
                        'bg-blue-50 text-blue-600 border-blue-200': b.status === 'ON_DELIVERY',
                        'bg-emerald-50 text-emerald-600 border-emerald-200': b.status === 'COMPLETED',
                        'bg-amber-50 text-amber-600 border-amber-200': b.status === 'SUPERVISOR_APPROVED' || b.status === 'PICKING'
                      }"
                    >
                      {{ b.status === 'ON_DELIVERY' ? 'Dalam Kirim' : b.status === 'COMPLETED' ? 'Selesai' : b.status }}
                    </span>
                  </div>
                  <div class="flex items-center justify-between text-[10px] text-slate-400 font-semibold border-t border-slate-100 pt-2">
                    <span>{{ formatDate(b.delivery_date || b.created_at) }}</span>
                    <span>{{ b.items?.length || 0 }} Rute / Nota</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 2: Pilih Toko / Nota di Surat Jalan -->
            <div v-if="selectedBatch" class="space-y-3 animate-in fade-in duration-200">
              <label class="text-xs font-black uppercase tracking-wider text-slate-600 flex items-center gap-1.5">
                <Building2 class="w-4 h-4 text-rose-500" />
                2. Pilih Toko / Nota yang Diretur
              </label>

              <div v-if="batchSalesOrders.length === 0" class="p-4 text-center text-xs text-slate-400 bg-slate-50 rounded-2xl border border-dashed border-slate-200">
                Surat Jalan ini tidak memiliki nota pesanan reguler toko.
              </div>
              <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
                <div
                  v-for="so in batchSalesOrders"
                  :key="so.id"
                  @click="selectSalesOrder(so)"
                  class="p-3.5 rounded-2xl border transition-all cursor-pointer flex flex-col justify-between"
                  :class="selectedSalesOrder?.id === so.id ? 'border-rose-500 bg-rose-50/50 shadow-sm ring-1 ring-rose-500' : 'border-slate-200 hover:border-slate-300 bg-white'"
                >
                  <div>
                    <div class="flex items-center justify-between">
                      <span class="text-xs font-black text-slate-900 truncate">{{ so.store?.name || 'Toko' }}</span>
                      <span class="text-[9px] font-bold text-slate-400 font-mono">{{ so.so_number }}</span>
                    </div>
                    <p class="text-[10px] text-slate-500 mt-1 line-clamp-1">
                      {{ so.store?.address || 'Alamat Toko' }}
                    </p>
                  </div>
                  <div class="mt-2 pt-2 border-t border-slate-100 flex items-center justify-between text-[10px]">
                    <span class="text-slate-400 font-medium">{{ so.items?.length || 0 }} Jenis Produk</span>
                    <span class="font-extrabold text-slate-700">{{ formatCurrency(so.total_amount) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 3: Tipe Perlakuan Retur & Catatan -->
            <div v-if="selectedSalesOrder" class="grid grid-cols-1 md:grid-cols-3 gap-4 p-4 rounded-2xl bg-slate-50 border border-slate-200 animate-in fade-in duration-200">
              <div>
                <label class="text-[11px] font-black uppercase text-slate-500 tracking-wider block mb-1.5">
                  Tipe Perlakuan Retur
                </label>
                <select
                  v-model="returnType"
                  class="w-full px-3 py-2 rounded-xl border border-slate-200 text-xs font-bold bg-white focus:outline-none focus:ring-2 focus:ring-rose-500"
                >
                  <option value="POTONG_NOTA">POTONG NOTA (Kredit Nota)</option>
                  <option value="TUKAR_LANGSUNG">TUKAR LANGSUNG (Armada)</option>
                  <option value="TUKAR_PRODUK">TUKAR PRODUK (Substitusi)</option>
                  <option value="GARANSI">KLAIM GARANSI</option>
                </select>
                <p class="text-[10px] text-slate-400 mt-1">
                  {{ returnType === 'POTONG_NOTA' ? 'Tagihan / piutang toko akan dipotong otomatis.' : 'Tagihan toko tetap penuh, barang diganti/diselesaikan terpisah.' }}
                </p>
              </div>

              <div class="md:col-span-2">
                <label class="text-[11px] font-black uppercase text-slate-500 tracking-wider block mb-1.5">
                  Catatan Fisik Pengembalian
                </label>
                <input
                  v-model="notes"
                  type="text"
                  placeholder="Contoh: Barang retur dibawa armada supir, pecah 2 botol saat transit..."
                  class="w-full px-3 py-2 rounded-xl border border-slate-200 text-xs focus:outline-none focus:ring-2 focus:ring-rose-500 bg-white"
                />
              </div>
            </div>

            <!-- Step 4: Rincian Produk Retur -->
            <div v-if="selectedSalesOrder" class="space-y-3 animate-in fade-in duration-200">
              <div class="flex items-center justify-between">
                <div>
                  <label class="text-xs font-black uppercase tracking-wider text-slate-600 flex items-center gap-1.5">
                    <Package class="w-4 h-4 text-rose-500" />
                    3. Rincian Barang yang Diretur
                  </label>
                  <p class="text-[11px] text-slate-400 mt-0.5">
                    Masukkan kuantitas fisik yang dikembalikan (maksimal sesuai barang yang dikirim ke toko ini).
                  </p>
                </div>
                <div class="text-right">
                  <span class="text-xs text-slate-400 font-bold block">Total Nilai Retur:</span>
                  <span class="text-base font-black text-rose-600">{{ formatCurrency(totalReturnAmount) }}</span>
                </div>
              </div>

              <!-- Product Table / Items List -->
              <div class="border border-slate-200 rounded-2xl overflow-hidden divide-y divide-slate-100 bg-white shadow-sm">
                <div
                  v-for="(item, idx) in items"
                  :key="item.product_id"
                  class="p-4 flex flex-col md:flex-row md:items-center justify-between gap-4 hover:bg-slate-50/50 transition-colors"
                  :class="{ 'bg-rose-50/30': item.quantity > 0 }"
                >
                  <!-- Info Produk -->
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-extrabold text-sm text-slate-900 truncate">{{ item.product_name }}</span>
                      <span class="text-[10px] font-bold text-slate-400 bg-slate-100 px-2 py-0.5 rounded-md">{{ item.unit }}</span>
                    </div>
                    <p class="text-[11px] text-slate-500 font-medium mt-0.5">
                      Harga Satuan: <span class="font-bold text-slate-700">{{ formatCurrency(item.price) }}</span> • Kuantitas Kirim: <span class="font-bold text-slate-700">{{ item.max_quantity }} {{ item.unit }}</span>
                    </p>
                  </div>

                  <!-- Controls: Kondisi, Alasan, Qty -->
                  <div class="flex flex-wrap items-center gap-3">
                    <!-- Kondisi -->
                    <div class="flex flex-col">
                      <label class="text-[9px] font-bold text-slate-400 uppercase tracking-wider mb-1">Kondisi</label>
                      <select
                        v-model="item.condition"
                        class="px-2.5 py-1.5 rounded-xl border border-slate-200 text-xs font-bold focus:outline-none focus:ring-2 focus:ring-rose-500 bg-white"
                      >
                        <option value="DAMAGED">🔴 RUSAK</option>
                        <option value="GOOD">🟢 BAGUS</option>
                      </select>
                    </div>

                    <!-- Alasan Retur -->
                    <div class="flex flex-col">
                      <label class="text-[9px] font-bold text-slate-400 uppercase tracking-wider mb-1">Alasan Retur</label>
                      <select
                        v-model="item.reason"
                        class="px-2.5 py-1.5 rounded-xl border border-slate-200 text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-rose-500 bg-white"
                      >
                        <option value="RUSAK">Rusak / Pecah</option>
                        <option value="KADALUARSA">Kadaluarsa</option>
                        <option value="TOLAK_CUSTOMER">Ditolak Customer</option>
                        <option value="SALAH_KIRIM">Salah Kirim</option>
                        <option value="TOKO_TUTUP">Toko Tutup</option>
                        <option value="LAINNYA">Lainnya</option>
                      </select>
                    </div>

                    <!-- Jumlah Retur -->
                    <div class="flex flex-col items-center">
                      <label class="text-[9px] font-bold text-slate-400 uppercase tracking-wider mb-1">Qty Retur</label>
                      <div class="flex items-center gap-1.5">
                        <button
                          type="button"
                          @click="item.quantity = Math.max(0, item.quantity - 1)"
                          class="w-8 h-8 flex items-center justify-center rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold transition"
                        >
                          <Minus class="w-3.5 h-3.5" />
                        </button>
                        <input
                          type="number"
                          v-model.number="item.quantity"
                          min="0"
                          :max="item.max_quantity"
                          class="w-14 text-center border border-slate-200 rounded-xl py-1 text-sm font-black focus:outline-none focus:ring-2 focus:ring-rose-500"
                        />
                        <button
                          type="button"
                          @click="item.quantity = Math.min(item.max_quantity, item.quantity + 1)"
                          class="w-8 h-8 flex items-center justify-center rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold transition"
                        >
                          <Plus class="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Option: Auto Approve Warehouse -->
              <div class="p-4 rounded-2xl bg-emerald-50/60 border border-emerald-200/80 flex items-start gap-3">
                <input
                  id="auto_receive"
                  type="checkbox"
                  v-model="autoApproveWarehouse"
                  class="mt-1 w-4 h-4 rounded text-emerald-600 focus:ring-emerald-500 border-slate-300"
                />
                <label for="auto_receive" class="text-xs cursor-pointer">
                  <span class="font-extrabold text-emerald-950 block">Langsung terima fisik ke stok gudang (Rekomendasi)</span>
                  <span class="text-emerald-700">
                    Karena retur diinput langsung oleh Admin Gudang yang memeriksa fisik barang, stok gudang (rusak/bagus) akan langsung bertambah dan status retur otomatis disahkan (APPROVED).
                  </span>
                </label>
              </div>
            </div>

          </div>

          <!-- Footer -->
          <div class="px-8 py-5 border-t border-slate-100 flex items-center justify-between bg-slate-50/80">
            <div class="text-xs text-slate-500">
              <span v-if="activeReturnItems.length > 0" class="font-bold text-slate-800">
                {{ activeReturnItems.length }} jenis produk dipilih untuk diretur
              </span>
              <span v-else>Pilih kuantitas produk yang diretur</span>
            </div>
            <div class="flex items-center gap-3">
              <Button
                variant="outline"
                class="rounded-xl px-5"
                @click="emit('close')"
              >
                Batal
              </Button>
              <Button
                @click="handleSubmitReturn"
                :disabled="isSubmitting || !selectedSalesOrder || activeReturnItems.length === 0"
                class="rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-bold px-7 shadow-md shadow-rose-600/20"
              >
                <RefreshCw v-if="isSubmitting" class="w-4 h-4 mr-2 animate-spin" />
                Simpan & Proses Retur
              </Button>
            </div>
          </div>

        </div>
      </div>
    </Transition>
  </Teleport>
</template>
