<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[760px] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2 text-xl font-black text-slate-900">
          <PackageCheck class="w-6 h-6 text-emerald-600" />
          Penerimaan Bahan Baku dari Supplier
        </DialogTitle>
        <DialogDescription>
          Catat kedatangan fisik bahan baku dari supplier ke gudang pabrik. Stok bahan baku akan otomatis bertambah.
        </DialogDescription>
      </DialogHeader>

      <form @submit.prevent="submit" class="space-y-5 mt-2">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-slate-50 p-4 rounded-2xl border border-slate-100">
          <!-- PO Reference -->
          <div class="space-y-1.5">
            <Label for="po_ref" class="text-xs font-bold text-slate-700">Referensi PO (Opsional)</Label>
            <select
              id="po_ref"
              v-model="form.purchase_order_id"
              @change="onPOChange"
              class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
            >
              <option :value="null">-- Penerimaan Langsung (Tanpa PO) --</option>
              <option v-for="p in availablePOs" :key="p.id" :value="p.id">
                {{ p.po_number }} - {{ p.supplier?.name }} ({{ formatStatus(p.status) }})
              </option>
            </select>
          </div>

          <!-- Supplier -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between">
              <Label for="supplier" class="text-xs font-bold text-slate-700">Supplier Terdaftar (Opsional)</Label>
              <span class="text-[10px] text-slate-400">Boleh dikosongkan jika beli ecer/langsung</span>
            </div>
            <select
              id="supplier"
              v-model="form.supplier_id"
              :disabled="!!form.purchase_order_id"
              class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/20 disabled:bg-slate-100 disabled:opacity-80"
            >
              <option :value="''">-- Tanpa Supplier Terdaftar / Pembelian Langsung --</option>
              <option v-for="s in suppliers" :key="s.id" :value="s.id">
                {{ s.name }} {{ s.city ? `(${s.city})` : '' }}
              </option>
            </select>
          </div>

          <!-- Supplier Name (Manual / Free-text) -->
          <div class="space-y-1.5">
            <Label for="supplier_name" class="text-xs font-bold text-slate-700">Nama Toko / Supplier Bebas</Label>
            <Input
              id="supplier_name"
              v-model="form.supplier_name"
              :disabled="!!form.purchase_order_id && !!form.supplier_id"
              placeholder="Contoh: Toko Baut Makmur, UD Sumber Teknik, dsb"
              class="rounded-xl border-slate-200 bg-white"
            />
          </div>

          <!-- Supplier Delivery Note No (No Surat Jalan Supplier / Nota) -->
          <div class="space-y-1.5">
            <Label for="sj_no" class="text-xs font-bold text-slate-700">No. Surat Jalan / Nota Fisik *</Label>
            <Input
              id="sj_no"
              v-model="form.supplier_delivery_note_no"
              placeholder="Contoh: SJ-SUP/2026/09/001 atau NOTA-0192"
              class="rounded-xl border-slate-200 bg-white"
              required
            />
          </div>

          <div class="space-y-1.5">
            <Label for="invoice_no" class="text-xs font-bold text-slate-700">No. Faktur Supplier</Label>
            <Input id="invoice_no" v-model="form.supplier_invoice_no" placeholder="Diisi sesuai faktur supplier" class="rounded-xl border-slate-200 bg-white" />
          </div>

          <div class="space-y-1.5">
            <Label for="invoice_date" class="text-xs font-bold text-slate-700">Tanggal Faktur</Label>
            <Input id="invoice_date" type="date" v-model="form.invoice_date" class="rounded-xl border-slate-200 bg-white" />
          </div>

          <div class="space-y-1.5">
            <Label for="payment_due_date" class="text-xs font-bold text-slate-700">Jatuh Tempo Pembayaran</Label>
            <Input id="payment_due_date" type="date" v-model="form.payment_due_date" class="rounded-xl border-slate-200 bg-white" />
          </div>

          <!-- Received Date -->
          <div class="space-y-1.5">
            <Label for="received_date" class="text-xs font-bold text-slate-700">Tanggal Penerimaan Fisik *</Label>
            <Input
              id="received_date"
              type="date"
              v-model="form.received_date"
              class="rounded-xl border-slate-200 bg-white"
              required
            />
          </div>

          <!-- Notes -->
          <div class="md:col-span-2 space-y-1.5">
            <Label for="notes" class="text-xs font-bold text-slate-700">Catatan Pemeriksaan Fisik</Label>
            <Input
              id="notes"
              v-model="form.notes"
              placeholder="Kondisi kemasan rapi, barang teknisi lengkap, lolos QC pabrik..."
              class="rounded-xl border-slate-200 bg-white"
            />
          </div>
        </div>

        <!-- Unloading Cost & Kas Pabrik Section (Mbak Linda) -->
        <div class="p-4 bg-amber-50/80 rounded-2xl border border-amber-200/80 space-y-3">
          <div class="flex items-center gap-2 text-amber-900 font-bold text-xs">
            <Wallet class="w-4 h-4 text-amber-700" />
            <span>Biaya Bongkar Muat / Ongkos Kuli (Kas Pabrik Mbak Linda)</span>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div class="space-y-1">
              <Label class="text-[11px] font-semibold text-slate-700">Ongkos Bongkar (Rp)</Label>
              <Input
                type="number"
                v-model.number="form.unloading_cost"
                min="0"
                placeholder="0"
                class="h-9 text-xs rounded-xl border-amber-300 font-mono font-bold bg-white text-amber-900"
              />
            </div>

            <div class="space-y-1">
              <Label class="text-[11px] font-semibold text-slate-700">Keterangan Tenaga / Kuli</Label>
              <Input
                v-model="form.unloading_notes"
                placeholder="Contoh: Kuli harian 3 orang..."
                class="h-9 text-xs rounded-xl border-slate-200 bg-white"
              />
            </div>

            <div class="space-y-1">
              <Label class="text-[11px] font-semibold text-slate-700">Posting ke Kasir Pabrik</Label>
              <select
                v-model="form.kas_pabrik_id"
                class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-amber-500/20"
              >
                <option :value="null">-- Tidak Posting Otomatis ke Kas --</option>
                <option v-for="k in kasPabrikList" :key="k.id" :value="k.id">
                  {{ k.name }} (Saldo: Rp {{ (k.balance || 0).toLocaleString('id-ID') }})
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- Items Table -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <div>
              <Label class="text-sm font-black text-slate-900">Rincian Barang yang Diterima</Label>
              <p class="text-[11px] text-slate-500">Mendukung Bahan Baku Pabrik maupun Barang / Sparepart Teknisi</p>
            </div>
            <Button
              v-if="!form.purchase_order_id"
              type="button"
              variant="outline"
              size="sm"
              @click="addItem"
              class="rounded-xl border-slate-200 text-xs font-bold text-emerald-600 hover:bg-emerald-50"
            >
              <Plus class="w-3.5 h-3.5 mr-1" /> Tambah Baris
            </Button>
          </div>

          <div class="border border-slate-200 rounded-2xl overflow-hidden divide-y divide-slate-100 bg-white">
            <div
              v-for="(item, idx) in form.items"
              :key="idx"
              class="p-4 space-y-3 relative hover:bg-slate-50/50 transition-colors"
            >
              <div class="flex items-center justify-between text-xs font-bold text-slate-500">
                <div class="flex items-center gap-2">
                  <span>Item #{{ idx + 1 }} {{ item.product_name ? `- ${item.product_name}` : '' }}</span>
                  <span
                    v-if="item.item_type === 'TECHNICIAN_ITEM'"
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-800 flex items-center gap-1"
                  >
                    <Wrench class="w-2.5 h-2.5" /> Barang Teknisi
                  </span>
                  <span
                    v-else
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800 flex items-center gap-1"
                  >
                    <Package class="w-2.5 h-2.5" /> Bahan Baku
                  </span>
                </div>
                <button
                  v-if="!form.purchase_order_id && form.items.length > 1"
                  type="button"
                  @click="removeItem(idx)"
                  class="text-red-500 hover:text-red-700 font-semibold text-xs flex items-center gap-1"
                >
                  <Trash2 class="w-3.5 h-3.5" /> Hapus
                </button>
              </div>

              <!-- Item Type & Selector -->
              <div class="grid grid-cols-1 md:grid-cols-12 gap-3 items-end">
                <!-- Item Type Selection -->
                <div class="md:col-span-3 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Jenis Barang</Label>
                  <select
                    v-model="item.item_type"
                    :disabled="!!form.purchase_order_id"
                    @change="onItemTypeChange(item)"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs font-bold focus:outline-none focus:ring-2 focus:ring-emerald-500/20 disabled:bg-slate-100"
                  >
                    <option value="RAW_MATERIAL">📦 Bahan Baku Pabrik</option>
                    <option value="FINISHED_GOODS">✅ Barang Jadi</option>
                    <option value="TECHNICIAN_ITEM">🔧 Barang / Sparepart Teknisi</option>
                  </select>
                </div>

                <!-- Product / Technician Item Selector -->
                <div class="md:col-span-5 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">
                    {{ item.item_type === 'TECHNICIAN_ITEM' ? 'Pilih Sparepart / Mur Baut *' : (item.item_type === 'FINISHED_GOODS' ? 'Pilih Barang Jadi *' : 'Pilih Bahan Baku / Produk *') }}
                  </Label>

                  <!-- Raw Material Select -->
                  <select
                    v-if="item.item_type !== 'TECHNICIAN_ITEM'"
                    v-model="item.product_id"
                    :disabled="!!form.purchase_order_id"
                    @change="onProductSelect(item)"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-emerald-500/20 disabled:bg-slate-100 font-medium"
                    required
                  >
                    <option disabled value="">Pilih Bahan / Produk...</option>
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

                  <!-- Technician Item Select -->
                  <select
                    v-else
                    v-model="item.technician_item_id"
                    @change="onTechnicianItemSelect(item)"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-amber-500/20 font-medium"
                    required
                  >
                    <option disabled value="">Pilih Barang Teknisi...</option>
                    <option v-for="t in technicianItems" :key="t.id" :value="t.id">
                      [{{ t.sku }}] {{ t.name }} - {{ t.category }} (Stok: {{ t.current_stock }} {{ t.unit }})
                    </option>
                  </select>
                </div>

                <!-- Target Warehouse (only for raw materials) -->
                <div v-if="item.item_type !== 'TECHNICIAN_ITEM'" class="md:col-span-4 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Gudang Penyimpanan</Label>
                  <select
                    v-model="item.factory_warehouse_id"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
                  >
                    <option :value="null">-- Gudang Default --</option>
                    <option v-for="w in factoryStore.factoryWarehouses" :key="w.id" :value="w.id">
                      {{ w.name }}
                    </option>
                  </select>
                </div>

                <!-- Technician Item Detail Info -->
                <div v-else class="md:col-span-4 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Lokasi Rak / Penyimpanan</Label>
                  <div class="h-9 px-2.5 py-1.5 bg-slate-100 rounded-xl text-xs font-mono text-slate-700 flex items-center">
                    {{ item.rack_location || 'Rak Teknisi Utama' }}
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-12 gap-3 items-end">
                <!-- Batch No -->
                <div class="md:col-span-6 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Nomor Batch / Ref *</Label>
                  <Input
                    v-model="item.batch_no"
                    placeholder="BATCH-..."
                    class="h-9 text-xs rounded-xl border-slate-200 font-mono"
                    required
                  />
                </div>

                <!-- Expiry Date -->
                <div class="md:col-span-6 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Tgl Kedaluwarsa (Opsional)</Label>
                  <Input
                    type="date"
                    v-model="item.expiry_date"
                    class="h-9 text-xs rounded-xl border-slate-200"
                  />
                </div>
              </div>

              <!-- Quantity Received & Rejected -->
              <div class="grid grid-cols-1 sm:grid-cols-12 gap-3 pt-2 bg-slate-50/70 p-3 rounded-xl border border-slate-100 items-end">
                <div class="sm:col-span-4 space-y-1">
                  <div class="flex justify-between items-center text-[11px] font-semibold">
                    <span class="text-emerald-700">Jumlah Diterima (Bagus) *</span>
                    <span v-if="item.po_requested_qty" class="text-slate-400 font-mono text-[10px]">
                      Sisa PO: {{ (item.po_requested_qty - (item.po_received_qty || 0)) }} {{ item.unit }}
                    </span>
                  </div>
                  <Input
                    type="number"
                    v-model.number="item.received_qty"
                    min="0"
                    class="h-9 text-xs rounded-xl border-emerald-300 font-mono font-bold bg-white text-emerald-800"
                    required
                  />
                </div>

                <div class="sm:col-span-3 space-y-1">
                  <Label class="text-[11px] font-semibold text-red-700">Jumlah Rijek</Label>
                  <Input
                    type="number"
                    v-model.number="item.rejected_qty"
                    min="0"
                    class="h-9 text-xs rounded-xl border-red-200 font-mono bg-white text-red-700"
                  />
                </div>

                <div class="sm:col-span-2 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Satuan (UoM) *</Label>
                  <input
                    list="receipt-uom-options"
                    v-model="item.unit"
                    placeholder="KG / PCS"
                    class="flex h-9 w-full rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs font-bold uppercase focus:outline-none focus:ring-2 focus:ring-emerald-500/20"
                    required
                  />
                </div>

                <div class="sm:col-span-3 space-y-1">
                  <Label class="text-[11px] font-semibold text-slate-600">Catatan Item</Label>
                  <Input
                    v-model="item.notes"
                    placeholder="Contoh: Baut M8 x 20mm grade 8.8..."
                    class="h-9 text-xs rounded-xl border-slate-200 bg-white"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Datalist for UoM options -->
          <datalist id="receipt-uom-options">
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
            class="rounded-xl font-bold shadow-lg shadow-emerald-500/20 bg-emerald-600 hover:bg-emerald-700 text-white"
          >
            <Loader2 v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
            <PackageCheck class="mr-2 h-4 w-4" v-else />
            Simpan Penerimaan & Tambah Stok
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
import { PackageCheck, Plus, Trash2, Loader2, Wallet, Wrench, Package } from 'lucide-vue-next'
import { useFactoryStore, type RawMaterialPO } from '@/stores/factory'
import apiClient from '@/api/axios'
import { factoryApi } from '@/api/factory'
import { toast } from 'vue-sonner'
import { groupProductsByCategory, ALL_COMMON_UNITS } from '@/utils/productUnits'

const props = defineProps<{
  open: boolean
  factoryId: string
  initialPO?: RawMaterialPO | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const factoryStore = useFactoryStore()
const loading = ref(false)
const suppliers = ref<any[]>([])
const kasPabrikList = ref<any[]>([])
const technicianItems = ref<any[]>([])

const generateDefaultBatch = () => {
  const now = new Date()
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  const d = String(now.getDate()).padStart(2, '0')
  return `BATCH-${y}${m}${d}`
}

const form = ref({
  purchase_order_id: null as string | null,
  supplier_id: '',
  supplier_name: '',
  supplier_delivery_note_no: '',
  supplier_invoice_no: '',
  invoice_date: '',
  payment_due_date: '',
  received_date: new Date().toISOString().split('T')[0],
  notes: '',
  unloading_cost: 0,
  unloading_notes: '',
  kas_pabrik_id: null as string | null,
  items: [
    {
      item_type: 'RAW_MATERIAL',
      po_item_id: null as string | null,
      product_id: '',
      technician_item_id: '',
      product_name: '',
      rack_location: '',
      factory_warehouse_id: null as string | null,
      batch_no: generateDefaultBatch(),
      expiry_date: '',
      received_qty: 0,
      rejected_qty: 0,
      unit: 'PCS',
      unit_price: 0,
      notes: '',
      po_requested_qty: 0,
      po_received_qty: 0,
    },
  ],
})

const groupedProducts = computed(() => {
  return groupProductsByCategory(factoryStore.products)
})

const availablePOs = computed(() => {
  return factoryStore.rawMaterialPOs.filter(
    (p) => p.status === 'APPROVED' || p.status === 'PARTIAL_RECEIVED'
  )
})

const formatStatus = (status?: string) => {
  if (status === 'APPROVED') return 'Disetujui'
  if (status === 'PARTIAL_RECEIVED') return 'Diterima Sebagian'
  return status || ''
}

const fetchSuppliers = async () => {
  try {
    const res = await apiClient.get('/warehouse/suppliers')
    suppliers.value = res.data?.data || []
  } catch (err) {
    console.error('Failed to fetch suppliers:', err)
  }
}

const fetchKasPabrik = async () => {
  try {
    const res = await apiClient.get('/kas-pabrik')
    kasPabrikList.value = res.data?.data || []
  } catch (err) {
    console.error('Failed to fetch kas pabrik:', err)
  }
}

const fetchTechnicianItems = async () => {
  if (!props.factoryId) return
  try {
    const res = await factoryApi.getTechnicianItems(props.factoryId)
    technicianItems.value = res.data?.data || []
  } catch (err) {
    console.error('Failed to fetch technician items:', err)
  }
}

const addItem = () => {
  form.value.items.push({
    item_type: 'RAW_MATERIAL',
    po_item_id: null,
    product_id: '',
    technician_item_id: '',
    product_name: '',
    rack_location: '',
    factory_warehouse_id: null,
    batch_no: generateDefaultBatch(),
    expiry_date: '',
    received_qty: 0,
    rejected_qty: 0,
    unit: 'PCS',
    unit_price: 0,
    notes: '',
    po_requested_qty: 0,
    po_received_qty: 0,
  })
}

const removeItem = (idx: number) => {
  form.value.items.splice(idx, 1)
}

const onItemTypeChange = (item: any) => {
  item.product_id = ''
  item.technician_item_id = ''
  item.product_name = ''
  item.rack_location = ''
}

const onProductSelect = (item: any) => {
  const prod = factoryStore.products.find((p) => p.id === item.product_id)
  if (prod) {
    item.product_name = prod.name
    item.unit = prod.unit || 'PCS'
  }
}

const onTechnicianItemSelect = (item: any) => {
  const tech = technicianItems.value.find((t) => t.id === item.technician_item_id)
  if (tech) {
    item.product_name = tech.name
    item.unit = tech.unit || 'PCS'
    item.rack_location = tech.rack_location || ''
    item.unit_price = tech.unit_price || 0
  }
}

const onPOChange = () => {
  if (!form.value.purchase_order_id) {
    return
  }

  const selectedPO = factoryStore.rawMaterialPOs.find((p) => p.id === form.value.purchase_order_id)
  if (selectedPO) {
    applyPOToForm(selectedPO)
  }
}

const applyPOToForm = (po: RawMaterialPO) => {
  form.value.purchase_order_id = po.id
  form.value.supplier_id = po.supplier_id
  form.value.items = (po.items || []).map((it) => {
    const remaining = Math.max(0, it.requested_qty - (it.received_qty || 0))
    return {
      item_type: 'RAW_MATERIAL',
      po_item_id: it.id || null,
      product_id: it.product_id,
      technician_item_id: '',
      product_name: it.product?.name || '',
      rack_location: '',
      factory_warehouse_id: it.factory_warehouse_id || null,
      batch_no: generateDefaultBatch(),
      expiry_date: '',
      received_qty: remaining,
      rejected_qty: 0,
      unit: it.unit || it.product?.unit || 'PCS',
      unit_price: it.estimated_unit_price || 0,
      notes: '',
      po_requested_qty: it.requested_qty,
      po_received_qty: it.received_qty || 0,
    }
  })
}

const resetForm = () => {
  form.value = {
    purchase_order_id: null,
    supplier_id: '',
    supplier_name: '',
    supplier_delivery_note_no: '',
    supplier_invoice_no: '',
    invoice_date: '',
    payment_due_date: '',
    received_date: new Date().toISOString().split('T')[0],
    notes: '',
    unloading_cost: 0,
    unloading_notes: '',
    kas_pabrik_id: null,
    items: [
      {
        item_type: 'RAW_MATERIAL',
        po_item_id: null,
        product_id: '',
        technician_item_id: '',
        product_name: '',
        rack_location: '',
        factory_warehouse_id: null,
        batch_no: generateDefaultBatch(),
        expiry_date: '',
        received_qty: 0,
        rejected_qty: 0,
        unit: 'KG',
        unit_price: 0,
        notes: '',
        po_requested_qty: 0,
        po_received_qty: 0,
      },
    ],
  }
}

watch(
  () => props.open,
  (val) => {
    if (val) {
      resetForm()
      fetchSuppliers()
      fetchKasPabrik()
      fetchTechnicianItems()
      if (props.factoryId) {
        factoryStore.fetchRawMaterialPOs({ factory_id: props.factoryId })
        factoryStore.fetchFactoryWarehouses(props.factoryId)
        factoryStore.fetchProducts()
      }

      if (props.initialPO) {
        applyPOToForm(props.initialPO)
      }
    }
  }
)

const submit = async () => {
  if (!props.factoryId) {
    toast.error('Pilih pabrik terlebih dahulu')
    return
  }

  const validItems = form.value.items.filter((it) => {
    const hasTarget = it.item_type === 'TECHNICIAN_ITEM' ? !!it.technician_item_id : !!it.product_id
    return hasTarget && (it.received_qty > 0 || it.rejected_qty > 0)
  })

  if (validItems.length === 0) {
    toast.error('Pilih barang dan masukkan jumlah yang diterima minimal 1 item')
    return
  }

  loading.value = true
  try {
    const payload = {
      factory_id: props.factoryId,
      purchase_order_id: form.value.purchase_order_id || undefined,
      supplier_id: form.value.supplier_id || undefined,
      supplier_name: form.value.supplier_name || undefined,
      supplier_delivery_note_no: form.value.supplier_delivery_note_no,
      supplier_invoice_no: form.value.supplier_invoice_no,
      invoice_date: form.value.invoice_date ? new Date(form.value.invoice_date) : undefined,
      payment_due_date: form.value.payment_due_date ? new Date(form.value.payment_due_date) : undefined,
      received_date: form.value.received_date ? new Date(form.value.received_date) : undefined,
      notes: form.value.notes,
      unloading_cost: form.value.unloading_cost || 0,
      unloading_notes: form.value.unloading_notes || '',
      kas_pabrik_id: form.value.kas_pabrik_id || undefined,
      items: validItems.map((it) => ({
        item_type: it.item_type || 'RAW_MATERIAL',
        po_item_id: it.po_item_id || undefined,
        product_id: it.item_type === 'TECHNICIAN_ITEM' ? undefined : it.product_id,
        technician_item_id: it.item_type === 'TECHNICIAN_ITEM' ? it.technician_item_id : undefined,
        factory_warehouse_id: it.item_type === 'TECHNICIAN_ITEM' ? undefined : (it.factory_warehouse_id || undefined),
        batch_no: it.batch_no || generateDefaultBatch(),
        expiry_date: it.expiry_date ? new Date(it.expiry_date) : undefined,
        received_qty: it.received_qty,
        rejected_qty: it.rejected_qty,
        unit: it.unit,
        unit_price: it.unit_price,
        notes: it.notes,
      })),
    }

    await factoryStore.createRawMaterialReceipt(payload)
    toast.success('Penerimaan berhasil disimpan dan stok gudang / teknisi diperbarui!')
    emit('success')
    emit('update:open', false)
  } catch (err: any) {
    toast.error(err?.toString() || 'Gagal menyimpan penerimaan')
  } finally {
    loading.value = false
  }
}
</script>
