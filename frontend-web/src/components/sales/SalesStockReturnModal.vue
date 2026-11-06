<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useSalesTransferStore } from '@/stores/salesTransfer'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { toast } from 'vue-sonner'
import {
  X,
  User,
  Package,
  Layers,
  CheckCircle2,
  AlertTriangle,
  Plus,
  Trash2,
  RotateCcw,
  Sparkles,
  Info
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  open: boolean
  branchId?: string
}>()

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
  (e: 'success'): void
}>()

const transferStore = useSalesTransferStore()
const masterStore = useMasterDataStore()
const factoryStore = useFactoryStore()

const isSubmitting = ref(false)
const selectedEmployeeId = ref('')
const returnMode = ref<'SALES_STOCK' | 'FREE'>('SALES_STOCK')
const autoApprove = ref(true)
const notes = ref('')

interface ReturnItemRow {
  product_id: string
  product_name: string
  unit: string
  available_sales_qty: number
  good_quantity: number
  damaged_quantity: number
  item_notes: string
}

const items = ref<ReturnItemRow[]>([])

// Filter Salesmen
const isSalesPosition = (title?: string, deptName?: string) => {
  const t = (title || '').toLowerCase()
  const d = (deptName || '').toLowerCase()
  const isExcluded = /(operator|driver|sopir|kuli|helper|gudang|warehouse|security|cleaning|harian|validator|editor|designer|desain|purchasing|auditor|kasir|cashier|teknisi|technician)/i.test(t)
  if (isExcluded) return false
  return /(sales|motoris|booster|marketing|take order|task order|\bto\b|\bmt\b|modern trade|key account|\bkam\b|kanvas)/i.test(t) ||
         /(sales|marketing|commercial)/i.test(d)
}

const activeSalesmen = computed(() => {
  const targetBranch = props.branchId || masterStore.selectedWarehouseBranchId || masterStore.selectedBranchId
  return masterStore.employees.filter((emp: any) => {
    const isActive = emp.user?.is_active !== false &&
      emp.employment_status !== 'RESIGNED' &&
      emp.employment_status !== 'TERMINATED'
    if (!isActive) return false

    if (targetBranch && targetBranch !== 'ALL_BRANCHES') {
      if (emp.branch_id && emp.branch_id !== targetBranch) return false
    }
    const jobTitle = emp.job_position?.title || ''
    const deptName = emp.department?.name || ''
    return isSalesPosition(jobTitle, deptName)
  }).sort((a: any, b: any) => {
    const nameA = `${a.first_name} ${a.last_name || ''}`.trim()
    const nameB = `${b.first_name} ${b.last_name || ''}`.trim()
    return nameA.localeCompare(nameB)
  })
})

// Load sales stock when salesman changes
const isLoadingSalesStock = ref(false)
watch(selectedEmployeeId, async (newEmpId) => {
  items.value = []
  if (!newEmpId) return

  isLoadingSalesStock.value = true
  try {
    await transferStore.fetchSalesStock(newEmpId)
    // Pre-populate if has stock and in SALES_STOCK mode
    populateFromSalesStock()
  } catch (error) {
    console.error('Failed to load sales stock:', error)
  } finally {
    isLoadingSalesStock.value = false
  }
})

const populateFromSalesStock = () => {
  if (returnMode.value !== 'SALES_STOCK') return

  const positiveStocks = transferStore.salesStock.filter((s) => s.product_id && s.quantity > 0)
  if (positiveStocks.length === 0) {
    // If no stock, suggest free mode or empty list
    items.value = []
    return
  }

  items.value = positiveStocks.map((s) => ({
    product_id: s.product_id!,
    product_name: s.product?.name || 'Produk',
    unit: s.product?.unit || 'PCS',
    available_sales_qty: s.quantity,
    good_quantity: 0,
    damaged_quantity: 0,
    item_notes: '',
  }))
}

// Watch switch mode
watch(returnMode, (newMode) => {
  if (newMode === 'SALES_STOCK') {
    populateFromSalesStock()
  } else {
    // Free mode: keep existing or add empty row
    if (items.value.length === 0) {
      addFreeItemRow()
    }
  }
})

const addFreeItemRow = () => {
  items.value.push({
    product_id: '',
    product_name: '',
    unit: 'PCS',
    available_sales_qty: 0,
    good_quantity: 0,
    damaged_quantity: 0,
    item_notes: '',
  })
}

const removeRow = (index: number) => {
  items.value.splice(index, 1)
}

const onProductSelect = (index: number, prodId: string) => {
  const prod = factoryStore.sellableProducts.find((p) => p.id === prodId)
  if (prod) {
    items.value[index].product_id = prod.id
    items.value[index].product_name = prod.name
    items.value[index].unit = prod.unit || 'PCS'
    
    // Check if sales actually has stock for this product
    const sStock = transferStore.salesStock.find((s) => s.product_id === prod.id)
    items.value[index].available_sales_qty = sStock?.quantity || 0
  }
}

// Total calculations
const totalGoodQty = computed(() => items.value.reduce((sum, item) => sum + (Number(item.good_quantity) || 0), 0))
const totalDamagedQty = computed(() => items.value.reduce((sum, item) => sum + (Number(item.damaged_quantity) || 0), 0))
const grandTotalQty = computed(() => totalGoodQty.value + totalDamagedQty.value)

// Close modal
const closeModal = () => {
  emit('update:open', false)
  selectedEmployeeId.value = ''
  items.value = []
  notes.value = ''
  returnMode.value = 'SALES_STOCK'
  autoApprove.value = true
}

// Submit
const handleSubmit = async () => {
  if (!selectedEmployeeId.value) {
    toast.error('Silakan pilih Salesman terlebih dahulu')
    return
  }

  // Filter items that have quantity > 0
  const validItems = items.value.filter((it) => {
    const totalItem = (Number(it.good_quantity) || 0) + (Number(it.damaged_quantity) || 0)
    return it.product_id && totalItem > 0
  })

  if (validItems.length === 0) {
    toast.error('Masukkan minimal 1 produk dengan kuantitas retur (Bagus atau Rusak) lebih dari 0')
    return
  }

  // If in SALES_STOCK mode, check limit
  if (returnMode.value === 'SALES_STOCK') {
    for (const it of validItems) {
      const totalItem = (Number(it.good_quantity) || 0) + (Number(it.damaged_quantity) || 0)
      if (totalItem > it.available_sales_qty) {
        toast.error(`Kuantitas retur ${it.product_name} (${totalItem}) melebihi saldo bawaan sales (${it.available_sales_qty}). Ganti ke Mode Bebas jika ingin mengabaikan batas.`)
        return
      }
    }
  }

  isSubmitting.value = true
  try {
    const targetBranch = props.branchId || masterStore.selectedWarehouseBranchId || masterStore.selectedBranchId
    
    const payloadItems = validItems.map((it) => {
      const good = Number(it.good_quantity) || 0
      const damaged = Number(it.damaged_quantity) || 0
      const total = good + damaged
      return {
        product_id: it.product_id,
        quantity: total,
        actual_quantity: good,
        damaged_quantity: damaged,
        unit: it.unit || 'pcs',
        notes: it.item_notes || notes.value,
      }
    })

    const payload = {
      branch_id: targetBranch,
      employee_id: selectedEmployeeId.value,
      type: 'RETURN',
      notes: notes.value ? `[RETUR_SALES] ${notes.value}` : '[RETUR_SALES] Pengembalian barang bawaan sales',
      ignore_stock_validation: returnMode.value === 'FREE',
      auto_approve: autoApprove.value,
      items: payloadItems,
    }

    await transferStore.createBatchTransfer(payload, targetBranch)
    toast.success(
      autoApprove.value
        ? 'Retur barang bawaan sales berhasil dibuat & stok gudang langsung bertambah!'
        : 'Pengajuan retur barang bawaan sales berhasil dibuat (Menunggu persetujuan gudang).'
    )
    emit('success')
    closeModal()
  } catch (error: any) {
    toast.error(error?.message || error || 'Gagal memproses retur barang bawaan sales')
  } finally {
    isSubmitting.value = false
  }
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    if (factoryStore.sellableProducts.length === 0) {
      factoryStore.fetchSellableProducts()
    }
    if (masterStore.employees.length === 0) {
      masterStore.fetchEmployees()
    }
  }
})
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
  >
    <div
      class="bg-white w-full max-w-4xl rounded-[28px] shadow-2xl border border-slate-100 flex flex-col max-h-[92vh] overflow-hidden"
    >
      <!-- Modal Header -->
      <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 text-white">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-2xl bg-white/10 flex items-center justify-center backdrop-blur-md border border-white/10 text-emerald-400">
            <RotateCcw class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-base font-black tracking-tight text-white flex items-center gap-2">
              Retur Barang Bawaan Sales
              <span class="text-[10px] bg-emerald-500/20 text-emerald-300 border border-emerald-400/30 px-2 py-0.5 rounded-full font-bold">
                Fleksibel
              </span>
            </h2>
            <p class="text-xs text-slate-300 mt-0.5">
              Tarik kembali sisa bawaan atau barang kanvas dari sales ke stok gudang (Bagus & Rusak)
            </p>
          </div>
        </div>
        <button
          @click="closeModal"
          class="w-8 h-8 rounded-full bg-white/10 hover:bg-white/20 text-slate-300 hover:text-white flex items-center justify-center transition-all"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 overflow-y-auto space-y-5 flex-1">
        <!-- Salesman Selection & Mode Selector -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-slate-50 p-4 rounded-2xl border border-slate-200/80">
          <!-- Salesman Dropdown -->
          <div class="space-y-1.5">
            <label class="text-xs font-black text-slate-700 flex items-center gap-1.5 uppercase tracking-wider">
              <User class="w-3.5 h-3.5 text-blue-600" /> Pilih Salesman / Motoris
            </label>
            <select
              v-model="selectedEmployeeId"
              class="w-full text-xs font-semibold px-3 py-2.5 rounded-xl border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500"
            >
              <option value="">-- Pilih Salesman Cabang --</option>
              <option v-for="emp in activeSalesmen" :key="emp.id" :value="emp.id">
                {{ emp.first_name }} {{ emp.last_name || '' }} ({{ emp.job_position?.title || 'Sales' }})
              </option>
            </select>
          </div>

          <!-- Mode Toggle: Mengacu Bawaan vs Bebas -->
          <div class="space-y-1.5">
            <label class="text-xs font-black text-slate-700 flex items-center gap-1.5 uppercase tracking-wider">
              <Layers class="w-3.5 h-3.5 text-purple-600" /> Mode Validasi Stok Bawaan
            </label>
            <div class="flex items-center gap-2 p-1 bg-white rounded-xl border border-slate-200">
              <button
                type="button"
                @click="returnMode = 'SALES_STOCK'"
                class="flex-1 py-1.5 px-3 rounded-lg text-xs font-black transition-all flex items-center justify-center gap-1.5"
                :class="returnMode === 'SALES_STOCK' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-500 hover:text-slate-800'"
              >
                <Sparkles class="w-3 h-3" /> Mengacu Bawaan Sales
              </button>
              <button
                type="button"
                @click="returnMode = 'FREE'"
                class="flex-1 py-1.5 px-3 rounded-lg text-xs font-black transition-all flex items-center justify-center gap-1.5"
                :class="returnMode === 'FREE' ? 'bg-purple-600 text-white shadow-sm' : 'text-slate-500 hover:text-slate-800'"
              >
                <RotateCcw class="w-3 h-3" /> Mode Bebas / Tanpa Acuan
              </button>
            </div>
          </div>
        </div>

        <!-- Mode Explanation Notice -->
        <div
          v-if="returnMode === 'SALES_STOCK'"
          class="flex items-start gap-2.5 bg-emerald-50 border border-emerald-200/80 rounded-2xl p-3 text-xs text-emerald-900"
        >
          <Info class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5" />
          <div>
            <span class="font-bold">Mode Mengacu Bawaan Sales:</span>
            Menampilkan produk & saldo yang tercatat sedang dibawa oleh sales. Kuantitas retur dibatasi maksimal saldo saat ini.
          </div>
        </div>
        <div
          v-else
          class="flex items-start gap-2.5 bg-purple-50 border border-purple-200/80 rounded-2xl p-3 text-xs text-purple-900"
        >
          <Sparkles class="w-4 h-4 text-purple-600 shrink-0 mt-0.5" />
          <div>
            <span class="font-bold">Mode Bebas / Tanpa Acuan (Fleksibel):</span>
            Admin gudang bebas memilih produk apapun dari katalog dan kuantitas berapapun. Cocok untuk barang titipan toko, sisa sampel fisik, atau yang belum sempat dicatat di sistem pindah barang.
          </div>
        </div>

        <!-- Loading State for Sales Stock -->
        <div v-if="isLoadingSalesStock" class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-6 w-6 border-2 border-slate-300 border-t-emerald-600"></div>
          <p class="text-xs font-bold text-slate-500 mt-2">Memuat saldo barang bawaan sales...</p>
        </div>

        <!-- Items Table Container -->
        <div v-else class="space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="text-xs font-black uppercase tracking-wider text-slate-700 flex items-center gap-1.5">
              <Package class="w-4 h-4 text-slate-500" />
              Daftar Barang yang Ditarik / Diretur
              <span class="text-slate-400 font-bold">({{ items.length }} Produk)</span>
            </h3>

            <!-- Button Add Item in Free Mode -->
            <Button
              v-if="returnMode === 'FREE'"
              type="button"
              @click="addFreeItemRow"
              size="sm"
              class="h-8 px-3 text-xs font-bold bg-purple-50 hover:bg-purple-100 text-purple-700 border border-purple-200 rounded-xl flex items-center gap-1"
            >
              <Plus class="w-3.5 h-3.5" /> Tambah Baris Produk
            </Button>
          </div>

          <!-- If empty items in SALES_STOCK mode -->
          <div
            v-if="items.length === 0 && selectedEmployeeId"
            class="text-center py-10 bg-slate-50 rounded-2xl border border-dashed border-slate-200"
          >
            <Package class="w-10 h-10 text-slate-300 mx-auto mb-2" />
            <p class="text-xs font-bold text-slate-600">Salesman ini tidak memiliki saldo barang bawaan aktif di sistem.</p>
            <p class="text-[11px] text-slate-400 mt-0.5 mb-3">
              Gunakan <strong>Mode Bebas</strong> jika Anda ingin menginput penerimaan barang secara fleksibel.
            </p>
            <Button
              type="button"
              @click="returnMode = 'FREE'"
              size="sm"
              class="text-xs font-bold bg-purple-600 hover:bg-purple-700 text-white rounded-xl"
            >
              Beralih ke Mode Bebas
            </Button>
          </div>

          <!-- Items Table -->
          <div v-else-if="items.length > 0" class="border border-slate-200 rounded-2xl overflow-hidden shadow-sm">
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs">
                <thead class="bg-slate-100/80 text-slate-700 font-black uppercase text-[10px] tracking-wider border-b border-slate-200">
                  <tr>
                    <th class="py-3 px-3 w-8 text-center">#</th>
                    <th class="py-3 px-3 min-w-[200px]">Nama Produk</th>
                    <th class="py-3 px-3 w-28 text-center">Sisa Sales</th>
                    <th class="py-3 px-3 w-32 text-center text-emerald-700 bg-emerald-50/50">
                      🟢 Qty Bagus
                    </th>
                    <th class="py-3 px-3 w-32 text-center text-rose-700 bg-rose-50/50">
                      🔴 Qty Rusak
                    </th>
                    <th class="py-3 px-3 min-w-[140px]">Keterangan Item</th>
                    <th v-if="returnMode === 'FREE'" class="py-3 px-2 w-10 text-center"></th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 font-medium text-slate-700">
                  <tr
                    v-for="(item, idx) in items"
                    :key="idx"
                    class="hover:bg-slate-50/70 transition-colors"
                  >
                    <td class="py-3 px-3 text-center text-slate-400 font-bold">{{ idx + 1 }}</td>
                    
                    <!-- Product Selection or Display -->
                    <td class="py-3 px-3">
                      <div v-if="returnMode === 'SALES_STOCK'">
                        <div class="font-black text-slate-900 text-xs">{{ item.product_name }}</div>
                        <div class="text-[10px] text-slate-400">Unit: {{ item.unit }}</div>
                      </div>
                      <div v-else>
                        <select
                          :value="item.product_id"
                          @change="onProductSelect(idx, ($event.target as HTMLSelectElement).value)"
                          class="w-full text-xs font-semibold px-2.5 py-1.5 rounded-lg border border-slate-200 bg-white focus:outline-none focus:ring-1 focus:ring-purple-500"
                        >
                          <option value="">-- Pilih Produk Gudang --</option>
                          <option v-for="prod in factoryStore.sellableProducts" :key="prod.id" :value="prod.id">
                            {{ prod.name }} ({{ prod.unit || 'pcs' }})
                          </option>
                        </select>
                      </div>
                    </td>

                    <!-- Sisa Bawaan Sales -->
                    <td class="py-3 px-3 text-center">
                      <span class="inline-block px-2 py-0.5 rounded-md text-[11px] font-black" :class="item.available_sales_qty > 0 ? 'bg-blue-50 text-blue-700' : 'bg-slate-100 text-slate-500'">
                        {{ item.available_sales_qty }} {{ item.unit }}
                      </span>
                    </td>

                    <!-- Qty Bagus -->
                    <td class="py-3 px-3 bg-emerald-50/30">
                      <div class="flex items-center justify-center gap-1">
                        <input
                          v-model.number="item.good_quantity"
                          type="number"
                          min="0"
                          placeholder="0"
                          class="w-20 text-center text-xs font-black py-1.5 rounded-lg border border-emerald-300 bg-white focus:outline-none focus:ring-2 focus:ring-emerald-500/30 text-emerald-800"
                        />
                        <span class="text-[10px] text-emerald-600 font-bold">{{ item.unit }}</span>
                      </div>
                    </td>

                    <!-- Qty Rusak -->
                    <td class="py-3 px-3 bg-rose-50/30">
                      <div class="flex items-center justify-center gap-1">
                        <input
                          v-model.number="item.damaged_quantity"
                          type="number"
                          min="0"
                          placeholder="0"
                          class="w-20 text-center text-xs font-black py-1.5 rounded-lg border border-rose-300 bg-white focus:outline-none focus:ring-2 focus:ring-rose-500/30 text-rose-800"
                        />
                        <span class="text-[10px] text-rose-600 font-bold">{{ item.unit }}</span>
                      </div>
                    </td>

                    <!-- Item Notes -->
                    <td class="py-3 px-3">
                      <input
                        v-model="item.item_notes"
                        type="text"
                        placeholder="Alasan / no batch..."
                        class="w-full text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 bg-white focus:outline-none focus:ring-1 focus:ring-slate-400"
                      />
                    </td>

                    <!-- Action Remove (Free mode) -->
                    <td v-if="returnMode === 'FREE'" class="py-3 px-2 text-center">
                      <button
                        type="button"
                        @click="removeRow(idx)"
                        class="p-1 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                        title="Hapus baris"
                      >
                        <Trash2 class="w-3.5 h-3.5" />
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Table Summary Footer -->
            <div class="p-3 bg-slate-50 border-t border-slate-200 flex flex-wrap items-center justify-between gap-3 text-xs">
              <span class="font-bold text-slate-500">Ringkasan Total Fisik:</span>
              <div class="flex items-center gap-4">
                <span class="font-bold text-emerald-700 bg-emerald-100/60 px-2.5 py-1 rounded-lg border border-emerald-200">
                  🟢 Total Bagus: <strong>{{ totalGoodQty }}</strong>
                </span>
                <span class="font-bold text-rose-700 bg-rose-100/60 px-2.5 py-1 rounded-lg border border-rose-200">
                  🔴 Total Rusak: <strong>{{ totalDamagedQty }}</strong>
                </span>
                <span class="font-black text-slate-900 bg-white px-3 py-1 rounded-lg border border-slate-200 shadow-sm">
                  📦 Grand Total: <strong>{{ grandTotalQty }} Unit</strong>
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Global Notes & Auto Approve Checkbox -->
        <div class="space-y-3 pt-2 border-t border-slate-100">
          <div class="space-y-1.5">
            <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
              Catatan Pengembalian (Opsional)
            </label>
            <input
              v-model="notes"
              type="text"
              placeholder="Contoh: Sisa kanvas sore rute pasar gede, 2 botol remuk di jalan..."
              class="w-full text-xs font-medium px-3.5 py-2.5 rounded-xl border border-slate-200 bg-white focus:outline-none focus:ring-2 focus:ring-slate-400/20 focus:border-slate-400"
            />
          </div>

          <!-- Auto Approve Option -->
          <label class="flex items-start gap-2.5 p-3 rounded-2xl bg-amber-50/70 border border-amber-200/80 cursor-pointer select-none">
            <input
              type="checkbox"
              v-model="autoApprove"
              class="mt-0.5 rounded text-emerald-600 focus:ring-emerald-500 w-4 h-4 cursor-pointer"
            />
            <div class="text-xs text-amber-900">
              <span class="font-black text-slate-900 flex items-center gap-1.5">
                <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" />
                Langsung terima fisik ke stok gudang (Auto-Approve)
              </span>
              <p class="text-[11px] text-amber-700 mt-0.5">
                Jika dicentang, stok fisik gudang (stok bagus & stok karantina rusak) akan langsung bertambah seketika tanpa perlu approval penerimaan tambahan.
              </p>
            </div>
          </label>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-4 border-t border-slate-100 bg-slate-50 flex items-center justify-between">
        <Button
          type="button"
          variant="outline"
          @click="closeModal"
          :disabled="isSubmitting"
          class="rounded-xl text-xs font-bold px-4"
        >
          Batal
        </Button>
        <div class="flex items-center gap-2">
          <Button
            type="button"
            @click="handleSubmit"
            :disabled="isSubmitting || grandTotalQty === 0 || !selectedEmployeeId"
            class="rounded-xl text-xs font-bold px-6 bg-emerald-600 hover:bg-emerald-700 text-white shadow-md shadow-emerald-600/20 flex items-center gap-2"
          >
            <CheckCircle2 v-if="!isSubmitting" class="w-4 h-4" />
            <div v-else class="animate-spin rounded-full h-3.5 w-3.5 border-2 border-white border-t-transparent"></div>
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan & Proses Retur' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
