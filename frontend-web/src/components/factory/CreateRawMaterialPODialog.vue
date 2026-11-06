<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2 text-xl font-black text-slate-900">
          <ShoppingCart class="w-6 h-6 text-primary" />
          {{ po ? 'Update Purchase Order Bahan Baku' : 'Buat Purchase Order (PO) Bahan Baku' }}
        </DialogTitle>
        <DialogDescription>
          Form pembuatan PO oleh Purchasing untuk pemesanan ke Supplier dan pengajuan nilai biaya ke Manajemen.
        </DialogDescription>
      </DialogHeader>

      <!-- PR Link Banner if created from PR -->
      <div v-if="form.purchase_request_id" class="p-3 bg-blue-50 border border-blue-200 rounded-2xl flex items-center justify-between text-xs">
        <div class="flex items-center gap-2 text-blue-900">
          <ClipboardList class="w-4 h-4 text-blue-600" />
          <span>Dibuat berdasarkan <strong>{{ linkedPRNumber || 'Purchase Request' }}</strong></span>
        </div>
        <button
          type="button"
          @click="clearLinkedPR"
          class="text-blue-600 hover:text-blue-800 font-semibold text-[11px] underline"
        >
          Lepaskan PR
        </button>
      </div>

      <form @submit.prevent="submit" class="space-y-5 mt-2">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-slate-50 p-4 rounded-2xl border border-slate-100">
          <!-- Supplier -->
          <div class="space-y-1.5">
            <Label for="supplier" class="text-xs font-bold text-slate-700">Supplier / Vendor *</Label>
            <select
              id="supplier"
              v-model="form.supplier_id"
              class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
              required
            >
              <option disabled value="">-- Pilih Supplier --</option>
              <option v-for="s in suppliers" :key="s.id" :value="s.id">
                {{ s.name }} {{ s.city ? `(${s.city})` : '' }}
              </option>
            </select>
          </div>

          <!-- Referensi PR (Jika belum ada) -->
          <div v-if="!form.purchase_request_id" class="space-y-1.5">
            <Label for="pr_select" class="text-xs font-bold text-slate-700">Referensi PR Gudang (Opsional)</Label>
            <select
              id="pr_select"
              v-model="selectedPRId"
              @change="onPRSelect"
              class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-primary/20 font-medium"
            >
              <option value="">-- Tanpa Referensi PR (PO Langsung) --</option>
              <option v-for="pr in pendingPRs" :key="pr.id" :value="pr.id">
                {{ pr.pr_number }} - {{ pr.created_by_name || 'Gudang' }} ({{ pr.items?.length || 0 }} item)
              </option>
            </select>
          </div>

          <!-- Order Date -->
          <div class="space-y-1.5">
            <Label for="order_date" class="text-xs font-bold text-slate-700">Tanggal Pengajuan PO *</Label>
            <Input
              id="order_date"
              type="date"
              v-model="form.order_date"
              class="rounded-xl border-slate-200 bg-white"
              required
            />
          </div>

          <!-- Estimated Arrival -->
          <div class="space-y-1.5">
            <Label for="estimated_arrival" class="text-xs font-bold text-slate-700">Estimasi Kedatangan (ETA)</Label>
            <Input
              id="estimated_arrival"
              type="date"
              v-model="form.estimated_arrival"
              class="rounded-xl border-slate-200 bg-white"
            />
          </div>

          <!-- Notes -->
          <div class="md:col-span-2 space-y-1.5">
            <Label for="notes" class="text-xs font-bold text-slate-700">Catatan Pengajuan ke Manajemen / Supplier</Label>
            <Input
              id="notes"
              v-model="form.notes"
              placeholder="Contoh: Pembelian bahan baku batch produksi, termin pembayaran 30 hari..."
              class="rounded-xl border-slate-200 bg-white"
            />
          </div>
        </div>

        <!-- Items Table -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div>
              <Label class="text-sm font-black text-slate-900">Daftar Bahan Baku & Rincian Biaya</Label>
              <p class="text-[11px] text-slate-500">Tentukan harga satuan bahan hasil negosiasi supplier.</p>
            </div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              @click="addItem"
              class="rounded-xl border-slate-200 text-xs font-bold text-primary hover:bg-primary/5"
            >
              <Plus class="w-3.5 h-3.5 mr-1" /> Tambah Bahan
            </Button>
          </div>

          <div class="border border-slate-200 rounded-2xl overflow-hidden divide-y divide-slate-100 bg-white">
            <div
              v-for="(item, idx) in form.items"
              :key="idx"
              class="p-4 space-y-3 relative hover:bg-slate-50/50 transition-colors"
            >
              <div class="flex items-center justify-between text-xs font-bold text-slate-500">
                <span>Item #{{ idx + 1 }}</span>
                <button
                  v-if="form.items.length > 1"
                  type="button"
                  @click="removeItem(idx)"
                  class="text-red-500 hover:text-red-700 font-semibold text-xs flex items-center gap-1"
                >
                  <Trash2 class="w-3.5 h-3.5" /> Hapus
                </button>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-12 gap-3 items-end">
                <!-- Product -->
                <div class="md:col-span-4 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Bahan Baku / Barang *</Label>
                  <select
                    v-model="item.product_id"
                    @change="onProductSelect(item)"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-primary/20 font-medium"
                    required
                  >
                    <option disabled value="">Pilih Barang / Bahan...</option>
                    <optgroup v-if="groupedProducts.rawMaterials.length > 0" label="⭐ BAHAN BAKU UTAMA">
                      <option v-for="p in groupedProducts.rawMaterials" :key="p.id" :value="p.id">
                        {{ p.sku }} - {{ p.name }} (Satuan: {{ p.unit || 'KG' }})
                      </option>
                    </optgroup>
                    <optgroup v-if="groupedProducts.packagingMaterials.length > 0" label="📦 KEMASAN & WADAH">
                      <option v-for="p in groupedProducts.packagingMaterials" :key="p.id" :value="p.id">
                        {{ p.sku }} - {{ p.name }} (Satuan: {{ p.unit || 'PCS' }})
                      </option>
                    </optgroup>
                    <optgroup v-if="groupedProducts.tradingProducts.length > 0" label="🏷️ PRODUK PELENGKAP / TRADING">
                      <option v-for="p in groupedProducts.tradingProducts" :key="p.id" :value="p.id">
                        {{ p.sku }} - {{ p.name }} (Satuan: {{ p.unit || 'PCS' }})
                      </option>
                    </optgroup>
                    <optgroup v-if="groupedProducts.giftProducts.length > 0" label="🎁 HADIAH & PROMOSI">
                      <option v-for="p in groupedProducts.giftProducts" :key="p.id" :value="p.id">
                        {{ p.sku }} - {{ p.name }} (Satuan: {{ p.unit || 'PCS' }})
                      </option>
                    </optgroup>
                    <optgroup v-if="groupedProducts.finishedGoods.length > 0" label="⚪ BARANG JADI">
                      <option v-for="p in groupedProducts.finishedGoods" :key="p.id" :value="p.id">
                        {{ p.sku }} - {{ p.name }} (Satuan: {{ p.unit || 'PCS' }})
                      </option>
                    </optgroup>
                  </select>
                </div>

                <!-- Target Warehouse -->
                <div class="md:col-span-2 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Gudang Tujuan</Label>
                  <select
                    v-model="item.factory_warehouse_id"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option :value="null">-- Gudang Utama --</option>
                    <option v-for="w in factoryStore.factoryWarehouses" :key="w.id" :value="w.id">
                      {{ w.name }}
                    </option>
                  </select>
                </div>

                <!-- Qty -->
                <div class="md:col-span-2 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Jumlah *</Label>
                  <Input
                    type="number"
                    v-model.number="item.requested_qty"
                    min="1"
                    class="h-9 text-xs rounded-xl border-slate-200 font-mono font-bold"
                    required
                  />
                </div>

                <!-- Satuan / UoM -->
                <div class="md:col-span-2 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Satuan (UoM) *</Label>
                  <input
                    list="po-uom-options"
                    v-model="item.unit"
                    placeholder="KG / PCS"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs font-bold uppercase focus:outline-none focus:ring-2 focus:ring-primary/20"
                    required
                  />
                </div>

                <!-- Estimated Unit Price -->
                <div class="md:col-span-2 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Harga Satuan (Rp) *</Label>
                  <Input
                    type="number"
                    v-model.number="item.estimated_unit_price"
                    min="0"
                    placeholder="0"
                    class="h-9 text-xs rounded-xl border-emerald-300 font-mono font-bold bg-emerald-50/50 text-emerald-950"
                    required
                  />
                </div>
              </div>

              <!-- Subtotal display & notes -->
              <div class="flex items-center justify-between text-xs text-slate-500 pt-1 border-t border-slate-100">
                <span class="text-[11px]">
                  Catatan: <span class="text-slate-700 italic">{{ item.notes || '-' }}</span>
                </span>
                <span class="font-medium">
                  Subtotal: <strong class="text-emerald-700 font-mono font-bold">Rp {{ ((item.requested_qty || 0) * (item.estimated_unit_price || 0)).toLocaleString('id-ID') }}</strong>
                </span>
              </div>
            </div>
          </div>

          <!-- Datalist for UoM options -->
          <datalist id="po-uom-options">
            <option v-for="u in ALL_COMMON_UNITS" :key="u" :value="u">{{ u }}</option>
          </datalist>

          <!-- Total Calculation -->
          <div class="flex justify-between items-center bg-slate-900 text-white p-4 rounded-2xl shadow-sm">
            <span class="text-sm font-bold text-slate-300">Total Nilai Pengajuan PO:</span>
            <span class="text-xl font-black text-amber-400 font-mono">
              Rp {{ totalEstimatedAmount.toLocaleString('id-ID') }}
            </span>
          </div>
        </div>

        <DialogFooter class="gap-2 sm:gap-0 pt-2">
          <Button type="button" variant="outline" class="rounded-xl font-bold" @click="$emit('update:open', false)">
            Batal
          </Button>
          <Button
            type="submit"
            :disabled="loading"
            class="rounded-xl font-bold shadow-lg shadow-primary/20 bg-primary hover:bg-primary/90 text-white"
          >
            <Loader2 v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
            <Send class="mr-2 h-4 w-4" v-else />
            {{ po ? 'Simpan Perubahan' : 'Ajukan PO ke Manajemen' }}
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
import { ShoppingCart, Plus, Trash2, Loader2, Send, ClipboardList } from 'lucide-vue-next'
import { useFactoryStore } from '@/stores/factory'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { groupProductsByCategory, ALL_COMMON_UNITS } from '@/utils/productUnits'

const props = defineProps<{
  open: boolean
  factoryId: string
  po?: any | null
  prefilledPr?: any | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const factoryStore = useFactoryStore()
const loading = ref(false)
const suppliers = ref<any[]>([])
const selectedPRId = ref('')
const linkedPRNumber = ref('')

const form = ref({
  supplier_id: '',
  purchase_request_id: null as string | null,
  order_date: new Date().toISOString().split('T')[0],
  estimated_arrival: '',
  notes: '',
  items: [
    {
      product_id: '',
      factory_warehouse_id: null as string | null,
      requested_qty: 1,
      unit: 'KG',
      estimated_unit_price: 0,
      notes: '',
    },
  ],
})

const pendingPRs = computed(() => {
  return factoryStore.rawMaterialPRs.filter((p) => p.status === 'PENDING' && p.factory_id && p.items?.every((it) => it.product_id))
})

const groupedProducts = computed(() => {
  return groupProductsByCategory(factoryStore.products)
})

const totalEstimatedAmount = computed(() => {
  return form.value.items.reduce((sum, item) => {
    return sum + (item.requested_qty || 0) * (item.estimated_unit_price || 0)
  }, 0)
})

const fetchSuppliers = async () => {
  try {
    const res = await apiClient.get('/warehouse/suppliers')
    suppliers.value = res.data?.data || []
  } catch (err) {
    console.error('Failed to load suppliers:', err)
  }
}

const addItem = () => {
  form.value.items.push({
    product_id: '',
    factory_warehouse_id: null,
    requested_qty: 1,
    unit: 'PCS',
    estimated_unit_price: 0,
    notes: '',
  })
}

const removeItem = (index: number) => {
  form.value.items.splice(index, 1)
}

const onProductSelect = (item: any) => {
  const prod = factoryStore.products.find((p) => p.id === item.product_id)
  if (prod) {
    item.unit = prod.unit || 'PCS'
    if (prod.cost_price && !item.estimated_unit_price) {
      item.estimated_unit_price = prod.cost_price
    }
  }
}

const onPRSelect = () => {
  if (!selectedPRId.value) return
  const pr = factoryStore.rawMaterialPRs.find((p) => p.id === selectedPRId.value)
  if (pr) {
    loadFromPR(pr)
  }
}

const clearLinkedPR = () => {
  form.value.purchase_request_id = null
  linkedPRNumber.value = ''
  selectedPRId.value = ''
}

const loadFromPR = (pr: any) => {
  form.value.purchase_request_id = pr.id
  linkedPRNumber.value = pr.pr_number
  if (pr.notes && !form.value.notes) {
    form.value.notes = `Kebutuhan dari PR ${pr.pr_number}: ${pr.notes}`
  }
  if (pr.required_date && !form.value.estimated_arrival) {
    form.value.estimated_arrival = String(pr.required_date).slice(0, 10)
  }
  if (pr.items && pr.items.length > 0) {
    form.value.items = pr.items.map((it: any) => {
      const prod = factoryStore.products.find((p) => p.id === (it.product_id || it.product?.id))
      return {
        product_id: it.product_id || it.product?.id,
        factory_warehouse_id: it.factory_warehouse_id || null,
        requested_qty: it.requested_qty || 1,
        unit: it.unit || prod?.unit || 'KG',
        estimated_unit_price: prod?.cost_price || 0,
        notes: it.notes || '',
      }
    })
  }
}

const resetForm = () => {
  form.value = {
    supplier_id: '',
    purchase_request_id: null,
    order_date: new Date().toISOString().split('T')[0],
    estimated_arrival: '',
    notes: '',
    items: [
      {
        product_id: '',
        factory_warehouse_id: null,
        requested_qty: 1,
        unit: 'KG',
        estimated_unit_price: 0,
        notes: '',
      },
    ],
  }
  linkedPRNumber.value = ''
  selectedPRId.value = ''
}

const loadPO = (po: any) => {
  form.value = {
    supplier_id: po.supplier_id || po.supplier?.id || '',
    purchase_request_id: po.purchase_request_id || null,
    order_date: po.order_date ? String(po.order_date).slice(0, 10) : new Date().toISOString().slice(0, 10),
    estimated_arrival: po.estimated_arrival ? String(po.estimated_arrival).slice(0, 10) : '',
    notes: po.notes || '',
    items: (po.items || []).map((it: any) => ({
      product_id: it.product_id,
      factory_warehouse_id: it.factory_warehouse_id || null,
      requested_qty: it.requested_qty,
      unit: it.unit || 'PCS',
      estimated_unit_price: it.estimated_unit_price || 0,
      notes: it.notes || '',
    })),
  }
  linkedPRNumber.value = po.purchase_request?.pr_number || ''
}

watch(
  () => props.open,
  (val) => {
    if (val) {
      if (props.po) {
        loadPO(props.po)
      } else if (props.prefilledPr) {
        resetForm()
        loadFromPR(props.prefilledPr)
      } else {
        resetForm()
      }
      fetchSuppliers()
      if (props.factoryId) {
        factoryStore.fetchFactoryWarehouses(props.factoryId)
        factoryStore.fetchProducts()
        factoryStore.fetchRawMaterialPRs({ factory_id: props.factoryId, status: 'PENDING' })
      }
    }
  }
)

const submit = async () => {
  if (!props.factoryId) {
    toast.error('Pilih pabrik terlebih dahulu')
    return
  }

  if (!form.value.supplier_id) {
    toast.error('Pilih supplier terlebih dahulu')
    return
  }

  if (form.value.items.some((it) => !it.product_id || it.requested_qty <= 0)) {
    toast.error('Pastikan semua bahan baku dipilih dan kuantitas lebih dari 0')
    return
  }

  loading.value = true
  try {
    const payload = {
      factory_id: props.factoryId,
      supplier_id: form.value.supplier_id,
      purchase_request_id: form.value.purchase_request_id || undefined,
      order_date: form.value.order_date ? new Date(form.value.order_date) : undefined,
      estimated_arrival: form.value.estimated_arrival ? new Date(form.value.estimated_arrival) : undefined,
      notes: form.value.notes,
      items: form.value.items.map((it) => ({
        product_id: it.product_id,
        factory_warehouse_id: it.factory_warehouse_id || undefined,
        requested_qty: it.requested_qty,
        unit: it.unit,
        estimated_unit_price: it.estimated_unit_price || 0,
        notes: it.notes,
      })),
    }

    if (props.po) {
      await factoryStore.updateRawMaterialPO(props.po.id, payload)
      toast.success('Purchase Order Bahan Baku berhasil diperbarui!')
    } else {
      await factoryStore.createRawMaterialPO(payload)
      toast.success('Purchase Order Bahan Baku berhasil diajukan ke Manajemen!')
    }
    emit('success')
    emit('update:open', false)
  } catch (err: any) {
    toast.error(err?.toString() || 'Gagal membuat Purchase Order')
  } finally {
    loading.value = false
  }
}
</script>
