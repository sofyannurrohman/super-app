<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-[100] flex items-center justify-center p-3 sm:p-6 bg-slate-900/60 backdrop-blur-md overflow-hidden animate-in fade-in duration-200"
      @click.self="handleClose"
    >
      <div
        class="w-full max-w-xl max-h-[90vh] sm:max-h-[85vh] flex flex-col bg-white rounded-[2rem] shadow-2xl overflow-hidden border border-slate-100 animate-in zoom-in-95 duration-200"
      >
        <!-- STICKY HEADER -->
        <div
          class="flex-shrink-0 px-6 py-5 bg-gradient-to-r from-slate-900 via-slate-800 to-indigo-950 text-white border-b border-slate-800 flex items-center justify-between relative overflow-hidden"
        >
          <div class="flex items-center gap-3.5 relative z-10">
            <div
              class="w-11 h-11 rounded-2xl bg-white/10 backdrop-blur-md flex items-center justify-center border border-white/20 text-amber-400"
            >
              <ClipboardCheck class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-base sm:text-lg font-black tracking-tight text-white">Stock Opname & Koreksi Stok</h3>
                <span
                  v-if="currentBranchName"
                  class="hidden sm:inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-extrabold bg-indigo-500/30 text-indigo-200 border border-indigo-400/30"
                >
                  {{ currentBranchName }}
                </span>
              </div>
              <p class="text-xs text-slate-300 font-medium mt-0.5">
                Sesuaikan stok sistem dengan kuantitas fisik di gudang
              </p>
            </div>
          </div>

          <button
            type="button"
            @click="handleClose"
            class="relative z-10 p-2 text-slate-400 hover:text-white rounded-full hover:bg-white/10 transition-colors focus:outline-none"
            aria-label="Tutup"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- SCROLLABLE BODY -->
        <div class="flex-1 overflow-y-auto overscroll-contain p-5 sm:p-6 space-y-4 bg-slate-50/50 custom-scrollbar">
          <!-- 1. PRODUCT & BATCH SELECTOR -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700 flex items-center justify-between">
              <span>Produk & Batch *</span>
              <span v-if="selectedItem && pcsPerUnit > 1" class="text-[11px] font-medium text-slate-400">
                1 {{ unitName }} = {{ pcsPerUnit }} PCS
              </span>
            </label>

            <!-- Locked item preview (if opened directly from a row) -->
            <div
              v-if="isLockedItem && selectedItem"
              class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm flex items-center justify-between gap-3"
            >
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-slate-100 flex items-center justify-center text-slate-600">
                  <Box class="w-5 h-5" />
                </div>
                <div>
                  <div class="font-black text-slate-900 text-sm">
                    {{ selectedItem.product?.name }}
                  </div>
                  <div class="flex items-center gap-2 mt-0.5 text-xs text-slate-500 font-medium">
                    <span class="font-mono">{{ selectedItem.product?.sku }}</span>
                    <span>•</span>
                    <span class="bg-slate-100 text-slate-700 px-1.5 py-0.5 rounded text-[10px] font-mono font-bold">
                      Batch: {{ selectedItem.batch_no || 'DEFAULT' }}
                    </span>
                  </div>
                </div>
              </div>

              <div class="text-right">
                <span class="text-[10px] font-bold text-slate-400 uppercase block">Stok Sistem</span>
                <span class="text-xs font-black text-slate-800">
                  {{ currentGoodQuantity + currentDamagedQuantity }} PCS
                </span>
              </div>
            </div>

            <!-- Select product dropdown (if opened globally) -->
            <div v-else class="space-y-2">
              <div class="relative">
                <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                <input
                  v-model="productFilterQuery"
                  type="text"
                  placeholder="Ketik nama produk atau batch..."
                  class="w-full pl-10 pr-4 py-2 bg-white border border-slate-200 rounded-xl text-xs font-medium text-slate-800 outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all"
                />
              </div>

              <select
                v-model="selectedItemKey"
                @change="onProductSelectChanged"
                class="w-full px-3 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold text-slate-800 outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all cursor-pointer"
              >
                <option :value="''" disabled>-- Pilih Produk & Batch Fisik --</option>
                <option
                  v-for="item in filteredInventoryItems"
                  :key="getItemKey(item)"
                  :value="getItemKey(item)"
                >
                  {{ item.product?.name }} (Batch: {{ item.batch_no || 'DEFAULT' }}) - SKU: {{ item.product?.sku }}
                </option>
              </select>
            </div>
          </div>

          <!-- 2. STOK BAGUS (SIAP JUAL / GOOD) -->
          <div class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
                <span class="text-xs font-bold text-slate-800">
                  Fisik Stok Bagus (Siap Jual) *
                </span>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-[11px] text-slate-400 font-medium">
                  Sistem: <strong class="text-slate-700 font-bold">{{ currentGoodQuantity }} PCS</strong>
                </span>
                <span
                  v-if="goodDiff !== 0"
                  class="text-[10px] font-black px-2 py-0.5 rounded-full uppercase"
                  :class="goodDiff > 0 ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'"
                >
                  {{ goodDiff > 0 ? `+${goodDiff}` : `${goodDiff}` }} PCS
                </span>
              </div>
            </div>

            <!-- Inputs -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3" v-if="pcsPerUnit > 1">
              <div class="space-y-1">
                <label class="text-[10px] font-bold text-slate-500 uppercase">{{ unitName }}</label>
                <div class="relative">
                  <input
                    type="number"
                    min="0"
                    v-model.number="goodUnits"
                    @input="recalculateGoodTotal"
                    class="w-full px-3 py-2 bg-slate-50/70 border border-slate-200 rounded-xl text-sm font-black text-slate-900 outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 text-center"
                    placeholder="0"
                  />
                  <span class="absolute right-3 top-1/2 -translate-y-1/2 text-[11px] font-bold text-slate-400 uppercase pointer-events-none">
                    {{ unitName }}
                  </span>
                </div>
              </div>

              <div class="space-y-1">
                <label class="text-[10px] font-bold text-slate-500 uppercase">Eceran (PCS)</label>
                <div class="relative">
                  <input
                    type="number"
                    min="0"
                    v-model.number="goodPcs"
                    @input="recalculateGoodTotal"
                    class="w-full px-3 py-2 bg-slate-50/70 border border-slate-200 rounded-xl text-sm font-black text-slate-900 outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 text-center"
                    placeholder="0"
                  />
                  <span class="absolute right-3 top-1/2 -translate-y-1/2 text-[11px] font-bold text-slate-400 uppercase pointer-events-none">
                    PCS
                  </span>
                </div>
              </div>
            </div>

            <!-- Single Input for 1:1 PCS -->
            <div v-else class="relative">
              <input
                type="number"
                min="0"
                v-model.number="goodQuantity"
                class="w-full px-3 py-2 bg-slate-50/70 border border-slate-200 rounded-xl text-sm font-black text-slate-900 outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 pr-14"
                placeholder="0"
              />
              <span class="absolute right-3.5 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400 uppercase pointer-events-none">
                PCS
              </span>
            </div>
          </div>

          <!-- 3. STOK RUSAK (DAMAGED) -->
          <div class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-rose-500"></span>
                <span class="text-xs font-bold text-slate-800">
                  Fisik Stok Rusak (Damaged)
                </span>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-[11px] text-slate-400 font-medium">
                  Sistem: <strong class="text-slate-700 font-bold">{{ currentDamagedQuantity }} PCS</strong>
                </span>
                <span
                  v-if="damagedDiff !== 0"
                  class="text-[10px] font-black px-2 py-0.5 rounded-full uppercase"
                  :class="damagedDiff > 0 ? 'bg-rose-100 text-rose-800' : 'bg-emerald-100 text-emerald-800'"
                >
                  {{ damagedDiff > 0 ? `+${damagedDiff}` : `${damagedDiff}` }} PCS
                </span>
              </div>
            </div>

            <!-- Inputs -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3" v-if="pcsPerUnit > 1">
              <div class="space-y-1">
                <label class="text-[10px] font-bold text-slate-500 uppercase">{{ unitName }}</label>
                <div class="relative">
                  <input
                    type="number"
                    min="0"
                    v-model.number="damagedUnits"
                    @input="recalculateDamagedTotal"
                    class="w-full px-3 py-2 bg-slate-50/70 border border-slate-200 rounded-xl text-sm font-black text-slate-900 outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 text-center"
                    placeholder="0"
                  />
                  <span class="absolute right-3 top-1/2 -translate-y-1/2 text-[11px] font-bold text-slate-400 uppercase pointer-events-none">
                    {{ unitName }}
                  </span>
                </div>
              </div>

              <div class="space-y-1">
                <label class="text-[10px] font-bold text-slate-500 uppercase">Eceran (PCS)</label>
                <div class="relative">
                  <input
                    type="number"
                    min="0"
                    v-model.number="damagedPcs"
                    @input="recalculateDamagedTotal"
                    class="w-full px-3 py-2 bg-slate-50/70 border border-slate-200 rounded-xl text-sm font-black text-slate-900 outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 text-center"
                    placeholder="0"
                  />
                  <span class="absolute right-3 top-1/2 -translate-y-1/2 text-[11px] font-bold text-slate-400 uppercase pointer-events-none">
                    PCS
                  </span>
                </div>
              </div>
            </div>

            <!-- Single Input for 1:1 PCS -->
            <div v-else class="relative">
              <input
                type="number"
                min="0"
                v-model.number="damagedQuantity"
                class="w-full px-3 py-2 bg-slate-50/70 border border-slate-200 rounded-xl text-sm font-black text-slate-900 outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 pr-14"
                placeholder="0"
              />
              <span class="absolute right-3.5 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400 uppercase pointer-events-none">
                PCS
              </span>
            </div>
          </div>

          <!-- 4. ALASAN OPNAME & CATATAN -->
          <div class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-3">
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Alasan Penyesuaian *</label>
              <select
                v-model="reasonPreset"
                @change="onReasonPresetChange"
                class="w-full px-3 py-2.5 bg-slate-50/70 border border-slate-200 rounded-xl text-xs font-bold text-slate-800 outline-none focus:ring-2 focus:ring-slate-900/10"
              >
                <option value="Stock Opname Rutin">📋 Stock Opname Rutin / Berkala</option>
                <option value="Selisih Perhitungan Fisik">⚖️ Selisih Perhitungan Fisik / Susut</option>
                <option value="Barang Rusak / Pecah / Bocor">💥 Barang Rusak / Pecah / Bocor</option>
                <option value="Penyesuaian Retur Toko">📦 Penyesuaian Retur Toko</option>
                <option value="Koreksi Selisih Surat Jalan">🚚 Koreksi Selisih Surat Jalan</option>
                <option value="Pemusnahan Barang Kadaluwarsa">⏳ Pemusnahan Barang Kadaluwarsa</option>
                <option value="Lainnya">📝 Alasan Lainnya (Ketik Manual)</option>
              </select>
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700">Catatan Tambahan</label>
              <textarea
                v-model="customNotes"
                rows="2"
                placeholder="Keterangan opsional hasil perhitungan fisik..."
                class="w-full px-3 py-2 bg-slate-50/70 border border-slate-200 rounded-xl text-xs font-medium text-slate-800 outline-none resize-none focus:ring-2 focus:ring-slate-900/10"
              ></textarea>
            </div>
          </div>
        </div>

        <!-- STICKY FOOTER -->
        <div
          class="flex-shrink-0 px-6 py-4 bg-slate-50/90 border-t border-slate-100 flex items-center justify-between gap-3"
        >
          <div class="text-xs font-medium text-slate-500">
            <span v-if="!selectedItem" class="text-amber-600 font-bold">Pilih produk terlebih dahulu</span>
            <span v-else-if="!hasChanges" class="text-slate-400">Tidak ada perubahan kuantitas</span>
            <span v-else class="text-emerald-700 font-bold">
              ✓ Siap memperbarui {{ (goodDiff !== 0 && damagedDiff !== 0) ? '2 tipe stok' : '1 tipe stok' }}
            </span>
          </div>

          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="handleClose"
              class="px-4 py-2 rounded-xl font-bold text-xs text-slate-600 hover:bg-slate-200/60 transition-colors"
            >
              Batal
            </button>
            <button
              type="button"
              @click="handleSave"
              :disabled="loading || !canSave"
              class="px-5 py-2 rounded-xl font-black text-xs text-white bg-slate-900 hover:bg-slate-800 shadow-md shadow-slate-900/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <RefreshCw v-if="loading" class="w-3.5 h-3.5 animate-spin" />
              <Save v-else class="w-3.5 h-3.5" />
              <span>Simpan Hasil Opname</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import {
  ClipboardCheck,
  X,
  Box,
  Search,
  Save,
  RefreshCw,
  Clock,
} from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
  items?: any[]
  allItems?: any[]
  initialItem?: any | null
  currentBranchName?: string
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
  (e: 'close'): void
  (e: 'save', payload: {
    selectedItem: any
    goodQuantity: number
    currentGoodQuantity: number
    damagedQuantity: number
    currentDamagedQuantity: number
    reason: string
    autoReclassify: boolean
  }): void
}>()

// Internal Form State
const selectedItem = ref<any | null>(null)
const selectedItemKey = ref<string>('')
const productFilterQuery = ref('')

const currentGoodQuantity = ref(0)
const currentDamagedQuantity = ref(0)

const goodUnits = ref(0)
const goodPcs = ref(0)
const goodQuantity = ref(0)

const damagedUnits = ref(0)
const damagedPcs = ref(0)
const damagedQuantity = ref(0)

const reasonPreset = ref('Stock Opname Rutin')
const customNotes = ref('')

const isLockedItem = computed(() => !!props.initialItem)

const pcsPerUnit = computed(() => {
  return selectedItem.value?.product?.pcs_per_unit || 1
})

const unitName = computed(() => {
  return selectedItem.value?.product?.unit || 'DUS'
})

const goodDiff = computed(() => {
  return goodQuantity.value - currentGoodQuantity.value
})

const damagedDiff = computed(() => {
  return damagedQuantity.value - currentDamagedQuantity.value
})

const hasChanges = computed(() => {
  return goodDiff.value !== 0 || damagedDiff.value !== 0
})

const finalReason = computed(() => {
  if (customNotes.value.trim()) {
    return `[${reasonPreset.value}] ${customNotes.value.trim()}`
  }
  return reasonPreset.value
})

const canSave = computed(() => {
  return !!selectedItem.value && hasChanges.value && !!finalReason.value.trim()
})

const getItemKey = (item: any) => {
  if (!item) return ''
  return `${item.product_id}_${item.batch_no || 'DEFAULT'}`
}

const filteredInventoryItems = computed(() => {
  const items = (props.items || []).filter((item: any) => item.batch_no !== 'BACKORDER')
  if (!productFilterQuery.value.trim()) return items
  const q = productFilterQuery.value.toLowerCase()
  return items.filter((item: any) => {
    const name = item.product?.name?.toLowerCase() || ''
    const sku = item.product?.sku?.toLowerCase() || ''
    const batch = item.batch_no?.toLowerCase() || ''
    return name.includes(q) || sku.includes(q) || batch.includes(q)
  })
})

const syncSubUnitsFromTotals = () => {
  const ppu = pcsPerUnit.value
  if (ppu > 1) {
    goodUnits.value = Math.floor(goodQuantity.value / ppu)
    goodPcs.value = goodQuantity.value % ppu
    damagedUnits.value = Math.floor(damagedQuantity.value / ppu)
    damagedPcs.value = damagedQuantity.value % ppu
  } else {
    goodUnits.value = 0
    goodPcs.value = goodQuantity.value
    damagedUnits.value = 0
    damagedPcs.value = damagedQuantity.value
  }
}

const recalculateGoodTotal = () => {
  const ppu = pcsPerUnit.value
  if (ppu > 1) {
    goodQuantity.value = Math.max(0, (goodUnits.value || 0) * ppu + (goodPcs.value || 0))
  } else {
    goodQuantity.value = Math.max(0, goodPcs.value || 0)
  }
}

const recalculateDamagedTotal = () => {
  const ppu = pcsPerUnit.value
  if (ppu > 1) {
    damagedQuantity.value = Math.max(0, (damagedUnits.value || 0) * ppu + (damagedPcs.value || 0))
  } else {
    damagedQuantity.value = Math.max(0, damagedPcs.value || 0)
  }
}

const onReasonPresetChange = () => {
  // no-op, reactively handled
}

const onProductSelectChanged = () => {
  const found = filteredInventoryItems.value.find((i: any) => getItemKey(i) === selectedItemKey.value)
  if (found) {
    loadItemData(found)
  }
}

const loadItemData = (item: any) => {
  selectedItem.value = item
  selectedItemKey.value = getItemKey(item)

  const items = (props.allItems || props.items || []).filter((i: any) => i.batch_no !== 'BACKORDER')
  const itemBatch = item.batch_no || 'DEFAULT'

  const goodItem = items.find(
    (i: any) =>
      i.product_id === item.product_id &&
      (i.batch_no || 'DEFAULT') === itemBatch &&
      (i.stock_type === 'GOOD' || !i.stock_type)
  )
  const damagedItem = items.find(
    (i: any) =>
      i.product_id === item.product_id &&
      (i.batch_no || 'DEFAULT') === itemBatch &&
      i.stock_type === 'DAMAGED'
  )

  currentGoodQuantity.value = goodItem ? (goodItem.quantity || 0) : (item.stock_type === 'GOOD' || !item.stock_type ? (item.quantity || 0) : 0)
  currentDamagedQuantity.value = damagedItem ? (damagedItem.quantity || 0) : (item.stock_type === 'DAMAGED' ? (item.quantity || 0) : 0)

  goodQuantity.value = currentGoodQuantity.value
  damagedQuantity.value = currentDamagedQuantity.value

  syncSubUnitsFromTotals()
}

// Initialize when modal is opened
watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.initialItem && props.initialItem.batch_no !== 'BACKORDER') {
        loadItemData(props.initialItem)
      } else if (filteredInventoryItems.value && filteredInventoryItems.value.length > 0) {
        loadItemData(filteredInventoryItems.value[0])
      } else {
        selectedItem.value = null
        selectedItemKey.value = ''
      }
      reasonPreset.value = 'Stock Opname Rutin'
      customNotes.value = ''
      productFilterQuery.value = ''
    }
  },
  { immediate: true }
)

const handleClose = () => {
  emit('update:open', false)
  emit('close')
}

const handleSave = () => {
  if (!canSave.value || !selectedItem.value) return
  const safeItem = {
    ...selectedItem.value,
    batch_no: (selectedItem.value.batch_no && selectedItem.value.batch_no !== 'BACKORDER') ? selectedItem.value.batch_no : 'DEFAULT',
  }
  emit('save', {
    selectedItem: safeItem,
    goodQuantity: goodQuantity.value,
    currentGoodQuantity: currentGoodQuantity.value,
    damagedQuantity: damagedQuantity.value,
    currentDamagedQuantity: currentDamagedQuantity.value,
    reason: finalReason.value,
    autoReclassify: false,
  })
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 9999px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 9999px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>
