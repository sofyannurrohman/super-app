<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[520px] p-0 overflow-hidden rounded-3xl border-none shadow-2xl">
      <!-- STICKY HEADER -->
      <div class="bg-gradient-to-r from-amber-600 to-amber-700 p-6 text-white border-b border-amber-700">
        <DialogHeader>
          <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-2xl bg-white/10 flex items-center justify-center border border-white/20">
              <ClipboardCheck class="h-6 w-6 text-white" />
            </div>
            <div>
              <DialogTitle class="text-lg font-black text-white">
                {{ isRowOpname ? 'Koreksi / Opname Stok' : 'Stock Opname & Saldo Awal' }}
              </DialogTitle>
              <DialogDescription class="text-amber-100/80 text-xs mt-0.5">
                Sesuaikan stok sistem dengan saldo fisik aktual di gudang pabrik.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>
      </div>

      <!-- SCROLLABLE BODY -->
      <form @submit.prevent="submit" class="flex flex-col">
        <div class="p-6 space-y-4 max-h-[72vh] sm:max-h-[75vh] overflow-y-auto custom-scrollbar bg-white">
          <!-- 1. ITEM SELECTOR OR LOCKED ITEM -->
          <div v-if="isRowOpname" class="p-4 bg-slate-50 rounded-2xl border border-slate-200 space-y-1">
            <div class="text-[10px] font-bold uppercase text-slate-500 tracking-wider">Bahan Baku Terpilih</div>
            <div class="flex items-center justify-between">
              <div>
                <div class="font-black text-slate-900 text-sm">{{ stockItem?.product?.name }}</div>
                <div class="text-xs text-slate-500 font-mono mt-0.5">{{ stockItem?.product?.sku }} • {{ stockItem?.product?.category || 'BAHAN BAKU' }}</div>
              </div>
              <Badge variant="secondary" class="bg-slate-200 text-slate-700 font-bold border-none text-xs">
                {{ stockItem?.product?.unit || 'KG' }}
              </Badge>
            </div>
          </div>

          <div v-else class="space-y-1.5">
            <Label for="product" class="text-xs font-bold text-slate-700">Pilih Bahan Baku / Kemasan *</Label>
            <select
              id="product"
              v-model="form.product_id"
              class="flex h-11 w-full rounded-2xl border border-slate-200 bg-slate-50/50 px-3 py-2 text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-amber-500/20"
              required
              @change="handleProductChange"
            >
              <option disabled value="">Pilih Bahan Baku / Kemasan...</option>
              <optgroup v-if="groupedProducts.rawMaterials.length > 0" label="⭐ BAHAN BAKU UTAMA">
                <option v-for="product in groupedProducts.rawMaterials" :key="product.id" :value="product.id">
                  {{ product.name }} ({{ product.sku }}) - {{ product.unit }}
                </option>
              </optgroup>
              <optgroup v-if="groupedProducts.packagingMaterials.length > 0" label="📦 KEMASAN & WADAH">
                <option v-for="product in groupedProducts.packagingMaterials" :key="product.id" :value="product.id">
                  {{ product.name }} ({{ product.sku }}) - {{ product.unit }}
                </option>
              </optgroup>
              <optgroup v-if="groupedProducts.tradingProducts.length > 0" label="🏷️ PRODUK PELENGKAP / TRADING">
                <option v-for="product in groupedProducts.tradingProducts" :key="product.id" :value="product.id">
                  {{ product.name }} ({{ product.sku }}) - {{ product.unit }}
                </option>
              </optgroup>
              <optgroup v-if="groupedProducts.giftProducts.length > 0" label="🎁 HADIAH & PROMOSI">
                <option v-for="product in groupedProducts.giftProducts" :key="product.id" :value="product.id">
                  {{ product.name }} ({{ product.sku }}) - {{ product.unit }}
                </option>
              </optgroup>
            </select>
          </div>

          <!-- 2. WAREHOUSE & BATCH -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div class="space-y-1.5">
              <Label for="warehouse" class="text-xs font-bold text-slate-700">Gudang Penyimpanan</Label>
              <select
                id="warehouse"
                v-model="form.target_warehouse_id"
                class="flex h-11 w-full rounded-2xl border border-slate-200 bg-slate-50/50 px-3 py-2 text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-amber-500/20"
                :disabled="isRowOpname && !!stockItem?.factory_warehouse_id"
              >
                <option value="">Gudang Utama (Default)</option>
                <option v-for="w in factoryStore.factoryWarehouses" :key="w.id" :value="w.id">
                  {{ w.name }} ({{ w.code }})
                </option>
              </select>
            </div>

            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <Label for="batch" class="text-xs font-bold text-slate-700">Nomor Batch</Label>
                <button
                  v-if="!isRowOpname && form.product_id"
                  type="button"
                  @click="regenerateOpnameBatch"
                  class="text-[10px] text-amber-600 hover:text-amber-700 font-bold flex items-center gap-1"
                >
                  <Sparkles class="w-3 h-3" /> Auto Batch
                </button>
              </div>
              <Input
                id="batch"
                v-model="form.batch_no"
                list="opname-batches-list"
                placeholder="Misal: RM-BB01-260918-1530"
                class="h-11 rounded-2xl bg-slate-50/50 border-slate-200 font-mono text-xs"
                :disabled="isRowOpname"
              />
              <datalist id="opname-batches-list">
                <option v-for="b in availableProductBatches" :key="b" :value="b">{{ b }}</option>
              </datalist>
            </div>
          </div>

          <!-- 3. STOK FISIK AKTUAL -->
          <div class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-3">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-1">
              <Label for="actual_qty" class="text-xs font-bold text-slate-800">
                Stok Fisik Aktual *
              </Label>
              <div class="flex items-center gap-2">
                <span class="text-[11px] text-slate-400 font-medium">
                  Stok Sistem: <strong class="text-slate-700 font-mono">{{ formatMultiUnit(systemStock, currentUnit, selectedProduct?.pcs_per_unit).summary }}</strong>
                </span>
                <span
                  v-if="variance !== 0"
                  class="text-[10px] font-black px-2 py-0.5 rounded-full uppercase font-mono"
                  :class="variance > 0 ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'"
                >
                  {{ varianceSign }}{{ Math.abs(variance).toLocaleString('id-ID') }} {{ selectedInputUnit }}
                </span>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
              <div class="sm:col-span-2 relative">
                <input
                  id="actual_qty"
                  type="number"
                  v-model.number="form.actual_quantity"
                  step="any"
                  min="0"
                  class="w-full px-3 py-2 bg-slate-50/70 border border-slate-200 rounded-xl text-sm font-black text-slate-900 outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500"
                  placeholder="0"
                  required
                />
              </div>
              <div class="sm:col-span-1">
                <select
                  v-model="selectedInputUnit"
                  class="w-full h-full min-h-[38px] px-3 py-1.5 bg-slate-100 border border-slate-200 rounded-xl text-xs font-black uppercase text-slate-800 outline-none focus:ring-2 focus:ring-amber-500/20"
                >
                  <option v-for="u in compatibleUnits" :key="u" :value="u">{{ u }}</option>
                </select>
              </div>
            </div>

            <!-- Conversion Live Preview -->
            <div
              v-if="selectedInputUnit !== currentUnit"
              class="flex items-center gap-2 p-2 bg-amber-50/80 rounded-xl border border-amber-200/60 text-xs text-amber-900"
            >
              <Sparkles class="w-3.5 h-3.5 text-amber-600 flex-shrink-0" />
              <span class="font-medium">
                Konversi ke Basis Sistem: <strong class="font-mono font-bold">{{ convertedBaseQty.toLocaleString('id-ID', { maximumFractionDigits: 3 }) }} {{ currentUnit }}</strong>
              </span>
            </div>
          </div>

          <!-- 4. ALASAN OPNAME & CATATAN -->
          <div class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-3">
            <div class="space-y-1.5">
              <Label for="reason_preset" class="text-xs font-bold text-slate-700">Alasan Penyesuaian *</Label>
              <select
                id="reason_preset"
                v-model="form.reason_preset"
                class="flex h-11 w-full rounded-2xl border border-slate-200 bg-slate-50/50 px-3 py-2 text-xs font-bold focus:outline-none focus:ring-2 focus:ring-amber-500/20"
                required
              >
                <option value="Stock Opname Rutin">📋 Stock Opname Rutin / Berkala</option>
                <option value="Saldo Awal Sistem">🚀 Setup Saldo Awal Gudang</option>
                <option value="Selisih Timbangan / Susut">⚖️ Selisih Timbangan / Penyusutan Fisik</option>
                <option value="Barang Rusak / Bocor / Kadaluarsa">⚠️ Barang Rusak / Bocor / Rusak Fisik</option>
                <option value="Koreksi Salah Catat Mutasi">✏️ Koreksi Kesalahan Input Sebelumnya</option>
                <option value="Lainnya">📝 Alasan Lainnya (Ketik di Catatan)</option>
              </select>
            </div>

            <div class="space-y-1.5">
              <Label for="notes" class="text-xs font-bold text-slate-700">Keterangan Tambahan</Label>
              <Textarea
                id="notes"
                v-model="form.notes"
                placeholder="Contoh: Dihitung oleh Tim Gudang Pabrik shift pagi..."
                rows="2"
                class="rounded-2xl bg-slate-50/50 border-slate-200 text-xs resize-none"
              />
            </div>
          </div>
        </div>

        <!-- STICKY FOOTER -->
        <DialogFooter class="p-4 bg-slate-50/80 border-t border-slate-100 flex items-center justify-end gap-2">
          <Button type="button" variant="outline" @click="$emit('update:open', false)" class="rounded-xl font-bold border-slate-200">
            Batal
          </Button>
          <Button type="submit" :disabled="loading" class="rounded-xl font-black bg-amber-600 hover:bg-amber-700 text-white shadow-md shadow-amber-600/20">
            <Loader2 v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
            Simpan Hasil Opname
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Loader2, ClipboardCheck, Sparkles } from 'lucide-vue-next'
import { useFactoryStore } from '@/stores/factory'
import { groupProductsByCategory, getCompatibleUnits, convertUnitQuantity, formatMultiUnit } from '@/utils/productUnits'
import { toast } from 'vue-sonner'

const props = defineProps<{
  open: boolean
  factoryId: string
  stockItem?: any
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const factoryStore = useFactoryStore()
const loading = ref(false)

const isRowOpname = computed(() => !!props.stockItem)

const groupedProducts = computed(() => {
  return groupProductsByCategory(factoryStore.products)
})

const generateRawMaterialBatch = (productId: string) => {
  const product = factoryStore.products.find((p) => p.id === productId)
  if (!product) return 'DEFAULT'

  const rawSku = (product.sku || product.name.substring(0, 4))
    .replace(/[^a-zA-Z0-9]/g, '')
    .toUpperCase()

  const now = new Date()
  const yy = String(now.getFullYear()).slice(2)
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')
  const hh = String(now.getHours()).padStart(2, '0')
  const min = String(now.getMinutes()).padStart(2, '0')

  return `RM-${rawSku}-${yy}${mm}${dd}-${hh}${min}`
}

const availableProductBatches = computed(() => {
  if (!form.value.product_id) return []
  const batches = factoryStore.rawMaterialStock
    .filter((s: any) => s.product_id === form.value.product_id && s.batch_no)
    .map((s: any) => s.batch_no)
  return Array.from(new Set(batches))
})

const form = ref({
  product_id: '',
  target_warehouse_id: '',
  batch_no: 'DEFAULT',
  actual_quantity: 0,
  reason_preset: 'Stock Opname Rutin',
  notes: '',
})

const selectedInputUnit = ref('KG')

const selectedProduct = computed(() => {
  if (isRowOpname.value) return props.stockItem?.product
  return factoryStore.products.find((p) => p.id === form.value.product_id)
})

const currentUnit = computed(() => {
  return selectedProduct.value?.unit || 'KG'
})

const compatibleUnits = computed(() => {
  return getCompatibleUnits(currentUnit.value)
})

const convertedBaseQty = computed(() => {
  const raw = form.value.actual_quantity || 0
  if (selectedInputUnit.value === currentUnit.value) return raw
  return convertUnitQuantity(raw, selectedInputUnit.value, currentUnit.value, selectedProduct.value?.pcs_per_unit || 1)
})

const systemStock = computed(() => {
  if (isRowOpname.value) {
    return props.stockItem?.quantity || 0
  }
  if (!form.value.product_id) return 0
  const matched = factoryStore.rawMaterialStock.find(
    (s: any) =>
      s.product_id === form.value.product_id &&
      (s.batch_no || 'DEFAULT') === (form.value.batch_no || 'DEFAULT') &&
      ((s.factory_warehouse_id || '') === (form.value.target_warehouse_id || ''))
  )
  return matched ? matched.quantity : 0
})

const variance = computed(() => {
  return Math.round(convertedBaseQty.value) - systemStock.value
})

const varianceSign = computed(() => {
  if (variance.value > 0) return '+'
  if (variance.value < 0) return '-'
  return ''
})

function handleProductChange() {
  if (!isRowOpname.value && form.value.product_id) {
    if (!form.value.batch_no || form.value.batch_no === 'DEFAULT') {
      form.value.batch_no = generateRawMaterialBatch(form.value.product_id)
    }
  }
  selectedInputUnit.value = currentUnit.value
  form.value.actual_quantity = systemStock.value
}

function regenerateOpnameBatch() {
  if (form.value.product_id) {
    form.value.batch_no = generateRawMaterialBatch(form.value.product_id)
    toast.info('Nomor batch baru berhasil digenerate')
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.factoryId) {
        factoryStore.fetchFactoryWarehouses(props.factoryId)
        if (factoryStore.products.length === 0) {
          factoryStore.fetchProducts()
        }
      }

      if (props.stockItem) {
        selectedInputUnit.value = props.stockItem.product?.unit || 'KG'
        form.value = {
          product_id: props.stockItem.product_id,
          target_warehouse_id: props.stockItem.factory_warehouse_id || '',
          batch_no: props.stockItem.batch_no || 'DEFAULT',
          actual_quantity: props.stockItem.quantity || 0,
          reason_preset: 'Stock Opname Rutin',
          notes: '',
        }
      } else {
        selectedInputUnit.value = 'KG'
        form.value = {
          product_id: '',
          target_warehouse_id: '',
          batch_no: 'DEFAULT',
          actual_quantity: 0,
          reason_preset: 'Stock Opname Rutin',
          notes: '',
        }
      }
    }
  }
)

async function submit() {
  if (!form.value.product_id) {
    toast.error('Pilih bahan baku terlebih dahulu')
    return
  }
  if (form.value.actual_quantity < 0) {
    toast.error('Stok fisik aktual tidak boleh negatif')
    return
  }

  loading.value = true
  const reasonText = form.value.notes
    ? `[${form.value.reason_preset}] ${form.value.notes}`
    : `[${form.value.reason_preset}] Penyesuaian stok opname`

  try {
    const finalQuantity = Math.round(convertedBaseQty.value)
    await factoryStore.mutateRawMaterial(props.factoryId, {
      product_id: form.value.product_id,
      target_warehouse_id: form.value.target_warehouse_id || undefined,
      batch_no: form.value.batch_no || 'DEFAULT',
      quantity: finalQuantity,
      type: 'ADJUSTMENT',
      notes: reasonText,
    })

    toast.success(
      `Hasil Stock Opname berhasil disimpan! Saldo stok diperbarui menjadi ${finalQuantity} ${currentUnit.value}`
    )
    emit('success')
    emit('update:open', false)
  } catch (err: any) {
    toast.error('Gagal menyimpan hasil stock opname: ' + (err.toString() || 'Terjadi kesalahan'))
  } finally {
    loading.value = false
  }
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
