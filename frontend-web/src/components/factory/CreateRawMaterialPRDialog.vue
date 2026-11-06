<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2 text-xl font-black text-slate-900">
          <ClipboardList class="w-6 h-6 text-primary" />
          {{ pr ? 'Update Purchase Request Bahan Baku' : 'Buat Purchase Request Bahan Baku' }}
        </DialogTitle>
        <DialogDescription>
          Ajukan permohonan pengadaan bahan baku ke bagian Purchasing untuk kebutuhan produksi atau persediaan gudang pabrik.
        </DialogDescription>
      </DialogHeader>

      <form @submit.prevent="submit" class="space-y-5 mt-2">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-slate-50 p-4 rounded-2xl border border-slate-100">
          <!-- Request Date -->
          <div class="space-y-1.5">
            <Label for="request_date" class="text-xs font-bold text-slate-700">Tanggal Pengajuan *</Label>
            <Input
              id="request_date"
              type="date"
              v-model="form.request_date"
              class="rounded-xl border-slate-200 bg-white"
              required
            />
          </div>

          <!-- Required Date (Tgl Dibutuhkan) -->
          <div class="space-y-1.5">
            <Label for="required_date" class="text-xs font-bold text-slate-700">Tanggal Dibutuhkan (Deadline)</Label>
            <Input
              id="required_date"
              type="date"
              v-model="form.required_date"
              class="rounded-xl border-slate-200 bg-white"
            />
          </div>

          <!-- Notes -->
          <div class="md:col-span-2 space-y-1.5">
            <Label for="notes" class="text-xs font-bold text-slate-700">Keterangan / Keperluan Produksi</Label>
            <Input
              id="notes"
              v-model="form.notes"
              placeholder="Contoh: Stok menipis, kebutuhan produksi kecap 2000 botol pekan depan..."
              class="rounded-xl border-slate-200 bg-white"
            />
          </div>
        </div>

        <!-- Items Table -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div>
              <Label class="text-sm font-black text-slate-900">Daftar Bahan Baku yang Diminta</Label>
              <p class="text-[11px] text-slate-500">Pilih bahan dan jumlah kuantitas fisik yang dibutuhkan.</p>
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
                <!-- Product Selector -->
                <div class="md:col-span-5 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Bahan Baku / Kemasan / Barang *</Label>
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
                    <optgroup v-if="groupedProducts.finishedGoods.length > 0" label="⚪ BARANG LAINNYA">
                      <option v-for="p in groupedProducts.finishedGoods" :key="p.id" :value="p.id">
                        {{ p.sku }} - {{ p.name }} (Satuan: {{ p.unit || 'PCS' }})
                      </option>
                    </optgroup>
                  </select>
                </div>

                <!-- Target Warehouse -->
                <div class="md:col-span-3 space-y-1">
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
                    list="pr-uom-options"
                    v-model="item.unit"
                    placeholder="KG / PCS"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs font-bold uppercase focus:outline-none focus:ring-2 focus:ring-primary/20"
                    required
                  />
                </div>
              </div>

              <!-- Notes for item -->
              <div class="pt-1">
                <Input
                  v-model="item.notes"
                  placeholder="Catatan tambahan item (misal: spesifikasi khusus atau prioritas)..."
                  class="h-8 text-xs rounded-xl border-slate-100 bg-slate-50 text-slate-600"
                />
              </div>
            </div>
          </div>

          <!-- Datalist for UoM options -->
          <datalist id="pr-uom-options">
            <option v-for="u in ALL_COMMON_UNITS" :key="u" :value="u">{{ u }}</option>
          </datalist>
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
            {{ pr ? 'Simpan Perubahan PR' : 'Kirim Request ke Purchasing' }}
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
import { ClipboardList, Plus, Trash2, Loader2, Send } from 'lucide-vue-next'
import { useFactoryStore } from '@/stores/factory'
import { toast } from 'vue-sonner'
import { groupProductsByCategory, ALL_COMMON_UNITS } from '@/utils/productUnits'

const props = defineProps<{
  open: boolean
  factoryId: string
  pr?: any | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const factoryStore = useFactoryStore()
const loading = ref(false)

const form = ref({
  request_date: new Date().toISOString().split('T')[0],
  required_date: '',
  notes: '',
  items: [
    {
      product_id: '',
      factory_warehouse_id: null as string | null,
      requested_qty: 1,
      unit: 'KG',
      notes: '',
    },
  ],
})

const groupedProducts = computed(() => {
  return groupProductsByCategory(factoryStore.products)
})

const addItem = () => {
  form.value.items.push({
    product_id: '',
    factory_warehouse_id: null,
    requested_qty: 1,
    unit: 'KG',
    notes: '',
  })
}

const removeItem = (index: number) => {
  form.value.items.splice(index, 1)
}

const onProductSelect = (item: any) => {
  const prod = factoryStore.products.find((p) => p.id === item.product_id)
  if (prod) {
    item.unit = prod.unit || 'KG'
  }
}

const resetForm = () => {
  form.value = {
    request_date: new Date().toISOString().split('T')[0],
    required_date: '',
    notes: '',
    items: [
      {
        product_id: '',
        factory_warehouse_id: null,
        requested_qty: 1,
        unit: 'KG',
        notes: '',
      },
    ],
  }
}

const loadPR = (pr: any) => {
  form.value = {
    request_date: pr.request_date ? String(pr.request_date).slice(0, 10) : new Date().toISOString().slice(0, 10),
    required_date: pr.required_date ? String(pr.required_date).slice(0, 10) : '',
    notes: pr.notes || '',
    items: (pr.items || []).map((it: any) => ({
      product_id: it.product_id || it.product?.id || '',
      factory_warehouse_id: it.factory_warehouse_id || null,
      requested_qty: it.requested_qty || 1,
      unit: it.unit || 'KG',
      notes: it.notes || '',
    })),
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.pr) {
        loadPR(props.pr)
      } else {
        resetForm()
      }
    }
  },
  { immediate: true }
)

const submit = async () => {
  if (!props.factoryId) {
    toast.error('Pabrik wajib dipilih')
    return
  }

  if (form.value.items.length === 0) {
    toast.error('Daftar bahan baku tidak boleh kosong')
    return
  }

  for (const item of form.value.items) {
    if (!item.product_id) {
      toast.error('Pilih bahan baku untuk semua baris')
      return
    }
    if (!item.requested_qty || item.requested_qty <= 0) {
      toast.error('Jumlah kuantitas harus lebih besar dari 0')
      return
    }
    if (!item.unit) {
      toast.error('Satuan (UoM) wajib diisi untuk semua baris')
      return
    }
  }

  loading.value = true
  try {
    const payload = {
      factory_id: props.factoryId,
      request_date: form.value.request_date ? new Date(form.value.request_date) : undefined,
      required_date: form.value.required_date ? new Date(form.value.required_date) : undefined,
      notes: form.value.notes,
      items: form.value.items.map((it) => ({
        product_id: it.product_id,
        factory_warehouse_id: it.factory_warehouse_id || undefined,
        requested_qty: Number(it.requested_qty),
        unit: it.unit.toUpperCase(),
        notes: it.notes,
      })),
    }

    if (props.pr) {
      await factoryStore.updateRawMaterialPR(props.pr.id, payload)
      toast.success('Purchase Request bahan baku berhasil diperbarui')
    } else {
      await factoryStore.createRawMaterialPR(payload)
      toast.success('Purchase Request berhasil diajukan ke Purchasing')
    }

    emit('success')
    emit('update:open', false)
  } catch (err: any) {
    console.error(err)
    toast.error(typeof err === 'string' ? err : err.message || 'Gagal menyimpan Purchase Request')
  } finally {
    loading.value = false
  }
}
</script>
