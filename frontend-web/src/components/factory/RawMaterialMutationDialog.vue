<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[520px] p-0 overflow-hidden rounded-3xl border-none shadow-2xl">
      <!-- STICKY HEADER -->
      <div class="bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 p-6 text-white border-b border-slate-800">
        <DialogHeader>
          <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-2xl bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20 text-emerald-400">
              <Boxes class="h-5 w-5" />
            </div>
            <div>
              <DialogTitle class="text-lg font-black text-white">
                Mutasi Stok Bahan Baku
              </DialogTitle>
              <DialogDescription class="text-slate-400 text-xs mt-0.5">
                Penerimaan (IN), pengeluaran (OUT), atau penyesuaian (ADJUSTMENT) bahan baku.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>
      </div>

      <!-- SCROLLABLE BODY -->
      <form @submit.prevent="submit" class="flex flex-col">
        <div class="p-6 space-y-4 max-h-[72vh] sm:max-h-[75vh] overflow-y-auto custom-scrollbar bg-white">
          <!-- 1. TIPE MUTASI (Segmented Pills) -->
          <div class="space-y-1.5">
            <Label class="text-xs font-bold text-slate-700">Tipe Mutasi *</Label>
            <div class="grid grid-cols-3 gap-2 p-1 bg-slate-100/80 rounded-2xl border border-slate-200/60">
              <button
                type="button"
                @click="setMutationType('IN')"
                :class="[
                  'py-2 px-2.5 rounded-xl text-xs font-bold transition-all flex items-center justify-center gap-1.5',
                  form.type === 'IN'
                    ? 'bg-emerald-600 text-white shadow-md shadow-emerald-600/20'
                    : 'text-slate-600 hover:text-slate-900 hover:bg-white/60'
                ]"
              >
                <ArrowDownLeft class="w-3.5 h-3.5" />
                <span>Masuk (IN)</span>
              </button>

              <button
                type="button"
                @click="setMutationType('OUT')"
                :class="[
                  'py-2 px-2.5 rounded-xl text-xs font-bold transition-all flex items-center justify-center gap-1.5',
                  form.type === 'OUT'
                    ? 'bg-rose-600 text-white shadow-md shadow-rose-600/20'
                    : 'text-slate-600 hover:text-slate-900 hover:bg-white/60'
                ]"
              >
                <ArrowUpRight class="w-3.5 h-3.5" />
                <span>Keluar (OUT)</span>
              </button>

              <button
                type="button"
                @click="setMutationType('ADJUSTMENT')"
                :class="[
                  'py-2 px-2.5 rounded-xl text-xs font-bold transition-all flex items-center justify-center gap-1.5',
                  form.type === 'ADJUSTMENT'
                    ? 'bg-amber-600 text-white shadow-md shadow-amber-600/20'
                    : 'text-slate-600 hover:text-slate-900 hover:bg-white/60'
                ]"
              >
                <SlidersHorizontal class="w-3.5 h-3.5" />
                <span>Penyesuaian</span>
              </button>
            </div>
          </div>

          <!-- 2. PILIH BAHAN BAKU / WADAH -->
          <div class="space-y-1.5">
            <Label for="product" class="text-xs font-bold text-slate-700">Bahan Baku / Kemasan *</Label>
            <select
              id="product"
              v-model="form.product_id"
              @change="onProductChange"
              class="flex h-11 w-full rounded-2xl border border-slate-200 bg-slate-50/50 px-3 py-2 text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-slate-900/10"
              required
            >
              <option disabled value="">Pilih Bahan Baku / Kemasan...</option>
              <optgroup v-if="groupedProducts.rawMaterials.length > 0" label="⭐ BAHAN BAKU UTAMA">
                <option v-for="product in groupedProducts.rawMaterials" :key="product.id" :value="product.id">
                  {{ product.sku }} - {{ product.name }} (Satuan: {{ product.unit || 'KG' }})
                </option>
              </optgroup>
              <optgroup v-if="groupedProducts.packagingMaterials.length > 0" label="📦 KEMASAN & WADAH">
                <option v-for="product in groupedProducts.packagingMaterials" :key="product.id" :value="product.id">
                  {{ product.sku }} - {{ product.name }} (Satuan: {{ product.unit || 'PCS' }})
                </option>
              </optgroup>
              <optgroup v-if="groupedProducts.tradingProducts.length > 0" label="🏷️ PRODUK PELENGKAP / TRADING">
                <option v-for="product in groupedProducts.tradingProducts" :key="product.id" :value="product.id">
                  {{ product.sku }} - {{ product.name }} (Satuan: {{ product.unit || 'PCS' }})
                </option>
              </optgroup>
              <optgroup v-if="groupedProducts.giftProducts.length > 0" label="🎁 HADIAH & PROMOSI">
                <option v-for="product in groupedProducts.giftProducts" :key="product.id" :value="product.id">
                  {{ product.sku }} - {{ product.name }} (Satuan: {{ product.unit || 'PCS' }})
                </option>
              </optgroup>
            </select>
          </div>

          <!-- 3. GUDANG PENYIMPANAN -->
          <div class="space-y-1.5">
            <Label for="warehouse" class="text-xs font-bold text-slate-700">Gudang Penyimpanan</Label>
            <select
              id="warehouse"
              v-model="form.target_warehouse_id"
              class="flex h-11 w-full rounded-2xl border border-slate-200 bg-slate-50/50 px-3 py-2 text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-slate-900/10"
            >
              <option value="">Gudang Utama (Default)</option>
              <option v-for="w in factoryStore.factoryWarehouses" :key="w.id" :value="w.id">
                {{ w.name }} ({{ w.code }})
              </option>
            </select>
          </div>

          <!-- 4. NOMOR BATCH -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between">
              <Label for="batch" class="text-xs font-bold text-slate-700">
                Nomor Batch
                <span v-if="form.type === 'IN'" class="text-[10px] text-emerald-600 font-bold ml-1">(Auto Sistem)</span>
                <span v-else-if="form.type === 'OUT'" class="text-[10px] text-rose-600 font-bold ml-1">(Pilih dari Stok)</span>
              </Label>

              <!-- Regenerate Button for IN/Adjustment -->
              <button
                v-if="form.product_id && form.type !== 'OUT'"
                type="button"
                @click="regenerateBatch"
                title="Generate Ulang Nomor Batch"
                class="text-[10px] text-emerald-600 hover:text-emerald-700 font-bold flex items-center gap-1 transition-colors"
              >
                <Sparkles class="w-3 h-3" />
                <span>Generate Baru</span>
              </button>
            </div>

            <!-- If OUT type and we have existing available batches -->
            <div v-if="form.type === 'OUT'">
              <select
                v-if="availableBatches.length > 0"
                v-model="form.batch_no"
                class="flex h-11 w-full rounded-2xl border border-slate-200 bg-slate-50/50 px-3 py-2 text-xs font-mono font-bold text-slate-800 focus:outline-none focus:ring-2 focus:ring-rose-500/20"
                required
              >
                <option disabled value="">-- Pilih Batch yang Tersedia --</option>
                <option v-for="b in availableBatches" :key="b.id || b.batch_no" :value="b.batch_no">
                  [{{ b.batch_no }}] - Sisa: {{ b.quantity }} {{ selectedProductUnit }} ({{ b.factory_warehouse?.name || 'Gudang Utama' }})
                </option>
              </select>
              <div v-else class="space-y-1">
                <Input
                  id="batch"
                  v-model="form.batch_no"
                  placeholder="Misal: RM-BBSLK-260918-1530"
                  class="h-11 rounded-2xl bg-slate-50/50 border-slate-200 font-mono text-xs"
                  required
                />
                <p class="text-[10px] text-amber-600 font-medium">⚠️ Belum ada catatan stok batch aktif untuk produk ini.</p>
              </div>
            </div>

            <!-- If IN or ADJUSTMENT type -->
            <div v-else class="space-y-1">
              <div class="relative">
                <Input
                  id="batch"
                  v-model="form.batch_no"
                  placeholder="Contoh: RM-BBSLK-260918-1530"
                  class="h-11 rounded-2xl bg-slate-50/50 border-slate-200 font-mono text-xs pr-8 font-bold"
                  required
                />
                <Tag class="w-4 h-4 text-slate-400 absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none" />
              </div>
            </div>
          </div>

          <!-- 5. KUANTITAS -->
          <div class="space-y-2">
            <div class="flex justify-between items-center">
              <Label for="quantity" class="text-xs font-bold text-slate-700">Kuantitas *</Label>
              <span v-if="selectedProductUnit" class="text-[11px] font-bold text-slate-500">
                Satuan Dasar: <strong class="text-slate-800 uppercase">{{ selectedProductUnit }}</strong>
              </span>
            </div>
            
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
              <div class="sm:col-span-2 relative">
                <Input
                  id="quantity"
                  type="number"
                  v-model.number="form.quantity"
                  required
                  min="0.001"
                  step="any"
                  class="h-11 font-mono font-black text-sm rounded-2xl bg-slate-50/50 border-slate-200"
                />
              </div>
              <div class="sm:col-span-1">
                <select
                  v-model="selectedInputUnit"
                  class="w-full h-full min-h-[42px] px-3 py-1.5 bg-slate-100 border border-slate-200 rounded-2xl text-xs font-black uppercase text-slate-800 outline-none focus:ring-2 focus:ring-slate-900/10"
                >
                  <option v-for="u in compatibleUnits" :key="u" :value="u">{{ u }}</option>
                </select>
              </div>
            </div>

            <!-- Conversion Live Preview -->
            <div
              v-if="selectedInputUnit !== selectedProductUnit && form.product_id"
              class="flex items-center gap-2 p-2 bg-emerald-50/80 rounded-xl border border-emerald-200/60 text-xs text-emerald-900"
            >
              <Sparkles class="w-3.5 h-3.5 text-emerald-600 flex-shrink-0" />
              <span class="font-medium">
                Konversi ke Basis Sistem: <strong class="font-mono font-bold">{{ convertedBaseQty.toLocaleString('id-ID', { maximumFractionDigits: 3 }) }} {{ selectedProductUnit }}</strong>
              </span>
            </div>

            <p class="text-[10px] text-slate-400" v-if="form.type === 'ADJUSTMENT'">
              Masukkan jumlah saldo fisik aktual untuk penyesuaian stok.
            </p>
            <p class="text-[10px] text-slate-400" v-else>
              Masukkan jumlah kuantitas fisik yang masuk atau keluar gudang.
            </p>
          </div>

          <!-- 6. CATATAN MUTASI -->
          <div class="space-y-1.5">
            <Label for="notes" class="text-xs font-bold text-slate-700">Catatan Mutasi</Label>
            <Textarea
              id="notes"
              v-model="form.notes"
              placeholder="Contoh: Penerimaan dari supplier / Pemakaian produksi mixing..."
              rows="2"
              class="rounded-2xl bg-slate-50/50 border-slate-200 text-xs resize-none"
            />
          </div>
        </div>

        <!-- STICKY FOOTER -->
        <DialogFooter class="p-4 bg-slate-50/80 border-t border-slate-100 flex items-center justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            class="rounded-xl font-bold border-slate-200"
            @click="$emit('update:open', false)"
          >
            Batal
          </Button>
          <Button
            type="submit"
            :disabled="loading"
            :class="[
              'rounded-xl font-black text-white shadow-md transition-all',
              form.type === 'IN'
                ? 'bg-emerald-600 hover:bg-emerald-700 shadow-emerald-600/20'
                : form.type === 'OUT'
                ? 'bg-rose-600 hover:bg-rose-700 shadow-rose-600/20'
                : 'bg-amber-600 hover:bg-amber-700 shadow-amber-600/20'
            ]"
          >
            <Loader2 v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
            Simpan Mutasi Stok
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
import { Textarea } from '@/components/ui/textarea'
import {
  Loader2,
  Boxes,
  Sparkles,
  Tag,
  ArrowDownLeft,
  ArrowUpRight,
  SlidersHorizontal,
} from 'lucide-vue-next'
import { useFactoryStore } from '@/stores/factory'
import { toast } from 'vue-sonner'
import { groupProductsByCategory, getCompatibleUnits, convertUnitQuantity, formatMultiUnit } from '@/utils/productUnits'

const props = defineProps<{
  open: boolean
  factoryId: string
  defaultType?: string
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const factoryStore = useFactoryStore()
const loading = ref(false)

const groupedProducts = computed(() => {
  return groupProductsByCategory(factoryStore.products)
})

const selectedProduct = computed(() => {
  return factoryStore.products.find((prod) => prod.id === form.value.product_id)
})

const selectedProductUnit = computed(() => {
  return selectedProduct.value?.unit || ''
})

const selectedInputUnit = ref('KG')

const compatibleUnits = computed(() => {
  return getCompatibleUnits(selectedProductUnit.value || 'KG')
})

const convertedBaseQty = computed(() => {
  const raw = form.value.quantity || 0
  const baseUnit = selectedProductUnit.value || 'KG'
  if (selectedInputUnit.value === baseUnit) return raw
  return convertUnitQuantity(raw, selectedInputUnit.value, baseUnit, selectedProduct.value?.pcs_per_unit || 1)
})

const availableBatches = computed(() => {
  if (!form.value.product_id) return []
  return factoryStore.rawMaterialStock.filter(
    (s: any) => s.product_id === form.value.product_id && (s.quantity > 0 || form.value.type === 'ADJUSTMENT')
  )
})

const form = ref({
  product_id: '',
  target_warehouse_id: '',
  type: 'IN',
  quantity: 1,
  batch_no: '',
  notes: '',
})

const setMutationType = (type: 'IN' | 'OUT' | 'ADJUSTMENT') => {
  form.value.type = type
  onTypeChange()
}

const generateRawMaterialBatch = (productId: string) => {
  const product = factoryStore.products.find((p) => p.id === productId)
  if (!product) return ''

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

const onProductChange = () => {
  if (!form.value.product_id) return
  selectedInputUnit.value = selectedProductUnit.value || 'KG'
  if (form.value.type === 'IN') {
    form.value.batch_no = generateRawMaterialBatch(form.value.product_id)
  } else if (form.value.type === 'OUT') {
    if (availableBatches.value.length > 0 && availableBatches.value[0]) {
      form.value.batch_no = availableBatches.value[0].batch_no || ''
    } else {
      form.value.batch_no = ''
    }
  } else {
    form.value.batch_no = generateRawMaterialBatch(form.value.product_id)
  }
}

const onTypeChange = () => {
  if (!form.value.product_id) return
  if (form.value.type === 'IN') {
    if (!form.value.batch_no || form.value.batch_no === 'DEFAULT') {
      form.value.batch_no = generateRawMaterialBatch(form.value.product_id)
    }
  } else if (form.value.type === 'OUT') {
    if (availableBatches.value.length > 0 && availableBatches.value[0]) {
      form.value.batch_no = availableBatches.value[0].batch_no || ''
    }
  }
}

const regenerateBatch = () => {
  if (form.value.product_id) {
    form.value.batch_no = generateRawMaterialBatch(form.value.product_id)
    toast.info('Nomor batch baru berhasil digenerate')
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      selectedInputUnit.value = 'KG'
      form.value = {
        product_id: '',
        target_warehouse_id: '',
        type: props.defaultType || 'IN',
        quantity: 1,
        batch_no: '',
        notes: '',
      }
      if (props.factoryId) {
        factoryStore.fetchFactoryWarehouses(props.factoryId)
        factoryStore.fetchProducts()
        factoryStore.fetchRawMaterialStock(props.factoryId)
      }
    }
  }
)

async function submit() {
  if (!form.value.product_id) {
    toast.error('Pilih bahan baku terlebih dahulu')
    return
  }

  // Ensure batch_no is populated
  if (!form.value.batch_no.trim()) {
    form.value.batch_no = generateRawMaterialBatch(form.value.product_id)
  }

  const finalQuantity = Math.round(convertedBaseQty.value)
  if (finalQuantity <= 0) {
    toast.error('Kuantitas mutasi harus lebih dari 0')
    return
  }

  loading.value = true
  try {
    const payload = {
      ...form.value,
      quantity: finalQuantity,
    }
    await factoryStore.mutateRawMaterial(props.factoryId, payload)
    toast.success('Mutasi bahan baku berhasil dicatat dengan batch ' + form.value.batch_no)
    emit('success')
    emit('update:open', false)
    form.value = {
      product_id: '',
      target_warehouse_id: '',
      type: 'IN',
      quantity: 1,
      batch_no: '',
      notes: '',
    }
  } catch (err: any) {
    toast.error(err.toString())
  } finally {
    loading.value = false
  }
}
</script>
