<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Manajemen Gudang Pabrik</h1>
        <p class="text-muted-foreground">Kelola daftar gudang penyimpanan di setiap pabrik.</p>
      </div>
      <div class="flex items-center gap-4">
        <!-- Factory Selector -->
        <select
          v-model="selectedFactoryId"
          @change="onFactoryChange"
          class="px-4 py-2 border rounded-xl bg-white shadow-sm font-medium text-slate-700 outline-none focus:ring-2 focus:ring-primary/20 transition-all"
        >
          <option value="" disabled>Pilih Pabrik...</option>
          <option v-for="factory in factoryStore.factories" :key="factory.id" :value="factory.id">
            {{ factory.name }}
          </option>
        </select>
        
        <Button
          @click="openCreateModal"
          :disabled="!selectedFactoryId"
          class="rounded-xl shadow-lg shadow-primary/20 transition-all hover:scale-105"
        >
          <Plus class="mr-2 h-4 w-4" /> Tambah Gudang Baru
        </Button>
      </div>
    </div>

    <!-- Alert if no factory selected -->
    <div v-if="!selectedFactoryId" class="flex flex-col items-center justify-center py-20 text-center bg-slate-50/50 rounded-3xl border border-dashed">
      <div class="w-16 h-16 bg-slate-100 rounded-2xl flex items-center justify-center mb-4 text-slate-400">
        <Warehouse class="w-8 h-8" />
      </div>
      <h3 class="text-lg font-bold text-slate-800 mb-2">Pilih Pabrik Terlebih Dahulu</h3>
      <p class="text-slate-500 max-w-sm">Silakan pilih pabrik pada dropdown di atas untuk melihat dan mengelola gudang di dalamnya.</p>
    </div>

    <Card v-else class="border-none shadow-sm bg-white rounded-3xl overflow-hidden animate-in fade-in slide-in-from-bottom-4 duration-500">
      <CardContent class="p-0">
        <DataTable :columns="columns" :data="filteredWarehouses" class="border-none">
          <template #cell-name="{ row }">
            <div class="flex items-center gap-3">
              <div
                class="h-10 w-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary"
              >
                <Warehouse class="h-5 w-5" />
              </div>
              <div>
                <span class="font-bold text-slate-900 block">{{ (row as any).name }}</span>
                <span v-if="(row as any).code" class="inline-flex items-center gap-1 text-[10px] font-black uppercase tracking-widest bg-indigo-50 text-indigo-700 border border-indigo-200 px-2 py-0.5 rounded-full mt-0.5">
                  <span class="w-1.5 h-1.5 rounded-full bg-indigo-500"></span>
                  {{ (row as any).code }}
                </span>
              </div>
            </div>
          </template>
          <template #cell-type="{ row }">
            <Badge
              variant="secondary"
              class="bg-amber-100 text-amber-700 font-bold uppercase text-[10px]"
            >
              {{ (row as any).type || 'UMUM' }}
            </Badge>
          </template>
          <template #cell-description="{ row }">
            <span class="text-sm text-slate-500">{{ (row as any).description || '-' }}</span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <Button
                variant="ghost"
                size="icon"
                @click="openStockModal(row)"
                class="h-8 w-8 text-slate-400 hover:text-emerald-600 hover:bg-emerald-50"
                title="Lihat Stok"
              >
                <Package class="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                @click="openEditModal(row)"
                class="h-8 w-8 text-slate-400 hover:text-primary hover:bg-primary/5"
              >
                <Edit2 class="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                @click="handleDelete((row as any).id)"
                class="h-8 w-8 text-slate-400 hover:text-destructive hover:bg-destructive/5"
              >
                <Trash2 class="h-3.5 w-3.5" />
              </Button>
            </div>
          </template>
        </DataTable>
      </CardContent>
    </Card>

    <!-- Modal Form -->
    <div
      v-if="showModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-lg shadow-2xl rounded-[2rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900">{{
                isEditing ? 'Edit Gudang' : 'Tambah Gudang Baru'
              }}</CardTitle>
              <CardDescription>Atur informasi ruangan/gudang penyimpanan.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showModal = false"
              class="rounded-full hover:bg-slate-100"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Nama Gudang</label>
            <input
              v-model="form.name"
              type="text"
              placeholder="e.g. Gudang Bahan Baku A"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-medium focus:ring-4 focus:ring-primary/5 focus:border-primary/20 outline-none transition-all"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Kode Gudang</label>
            <input
              v-model="form.code"
              type="text"
              placeholder="e.g. GD-RAW-1"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-mono font-bold focus:ring-4 focus:ring-primary/5 focus:border-primary/20 outline-none transition-all uppercase"
              @input="(e: any) => form.code = e.target.value.toUpperCase()"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Tipe Gudang</label>
            <select
              v-model="form.type"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none appearance-none cursor-pointer"
            >
              <option value="BAHAN_BAKU">Bahan Baku (Raw Material)</option>
              <option value="BARANG_JADI">Barang Jadi (Finished Goods)</option>
              <option value="REJECT">Barang Reject/Karantina</option>
              <option value="UMUM">Umum</option>
            </select>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Deskripsi/Keterangan</label>
            <textarea
              v-model="form.description"
              rows="3"
              placeholder="Catatan tambahan mengenai fungsi gudang ini..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-medium outline-none focus:ring-4 focus:ring-slate-100 resize-none"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button variant="ghost" @click="showModal = false" class="rounded-2xl font-bold h-12 px-8"
            >Batal</Button
          >
          <Button
            @click="handleSubmit"
            :loading="loading"
            class="rounded-2xl px-12 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90"
          >
            {{ isEditing ? 'Simpan Perubahan' : 'Buat Gudang' }}
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Modal Daftar Stok Gudang -->
    <div
      v-if="showStockModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-4xl shadow-2xl rounded-[2rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900">Stok Gudang: {{ selectedWarehouseForStock?.name }}</CardTitle>
              <CardDescription>Daftar barang yang tersimpan di gudang ini.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showStockModal = false"
              class="rounded-full hover:bg-slate-100"
            >
              <X class="h-6 w-6 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-0 bg-slate-50 max-h-[60vh] overflow-y-auto">
          <DataTable
            :columns="stockColumns"
            :data="warehouseStocks"
            :isLoading="factoryStore.loading"
            class="border-none"
          >
            <template #cell-product.name="{ row }">
              <span class="font-bold text-slate-900">{{ (row as any).product?.name }}</span>
            </template>
            <template #cell-batch_info="{ row }">
              <div class="flex flex-col gap-0.5">
                <span class="text-[11px] font-black text-slate-700 font-mono tracking-tight">{{
                  (row as any).batch_no || '-'
                }}</span>
                <span class="text-[9px] font-bold text-slate-400">
                  Exp:
                  {{
                    (row as any).expiry_date
                      ? new Date((row as any).expiry_date).toLocaleDateString('id-ID', {
                          day: 'numeric',
                          month: 'short',
                          year: 'numeric',
                        })
                      : '-'
                  }}
                </span>
              </div>
            </template>
            <template #cell-quantity="{ row }">
              <div class="flex items-center gap-2">
                <span class="text-base font-black text-slate-900">{{ (row as any).quantity }}</span>
                <span class="text-[9px] font-black text-slate-400 uppercase tracking-tighter">{{
                  (row as any).product?.unit
                }}</span>
              </div>
            </template>
            <template #cell-actions="{ row }">
              <Button
                variant="ghost"
                size="sm"
                @click="openAdjustStock(row)"
                class="h-8 w-8 p-0 rounded-lg text-slate-400 hover:text-primary hover:bg-primary/5"
                title="Koreksi Stok"
              >
                <Settings2 class="h-4 w-4" />
              </Button>
            </template>
          </DataTable>
        </CardContent>
      </Card>
    </div>

    <!-- Modal Koreksi Stok (Adjust Stock) -->
    <div
      v-if="showAdjustStock"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[110] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-md shadow-2xl rounded-[2rem] overflow-hidden border-none animate-in zoom-in-95 duration-200"
      >
        <CardHeader class="p-6 border-b bg-white">
          <CardTitle class="text-xl font-bold">Koreksi Stok</CardTitle>
          <CardDescription
            >Sesuaikan jumlah stok secara manual untuk produk {{ selectedStock?.product?.name }}.</CardDescription
          >
        </CardHeader>
        <CardContent class="p-6 space-y-4 bg-white">
          <div
            class="p-4 bg-blue-50 rounded-2xl border border-blue-100 flex justify-between items-center"
          >
            <div>
              <p class="text-xs font-bold text-blue-600 uppercase tracking-widest mb-1">Produk</p>
              <p class="text-sm font-bold text-slate-900">{{ selectedStock?.product?.name }}</p>
              <p v-if="selectedStock?.batch_no" class="text-[10px] text-slate-400 mt-0.5 font-mono">
                Batch: {{ selectedStock?.batch_no }}
              </p>
            </div>
            <div class="text-right">
              <p class="text-xs font-bold text-slate-400 uppercase">Stok Sekarang</p>
              <p class="text-lg font-black text-slate-900">{{ selectedStock?.quantity }}</p>
            </div>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Jumlah Stok Baru</label>
            <input
              type="number"
              v-model="adjustStockForm.quantity"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Alasan</label>
            <textarea
              v-model="adjustStockForm.reason"
              placeholder="Audit stok, Barang rusak, dsb..."
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20 resize-none"
              rows="2"
            ></textarea>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Pindah Lokasi Gudang (Opsional)</label>
            <select
              v-model="adjustStockForm.factory_warehouse_id"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option value="">-- Tetap di Gudang Saat Ini --</option>
              <option v-for="w in factoryStore.factoryWarehouses" :key="w.id" :value="w.id">
                {{ w.name }} ({{ w.code }})
              </option>
            </select>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 p-6 flex justify-end gap-3">
          <Button variant="ghost" @click="showAdjustStock = false" class="rounded-xl font-bold"
            >Batal</Button
          >
          <Button
            @click="handleAdjustStockSubmit"
            :loading="adjustSubmitting"
            class="rounded-xl px-8 shadow-lg shadow-primary/20 font-bold bg-blue-600 hover:bg-blue-700 text-white"
          >
            Simpan Perubahan
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useFactoryStore } from '@/stores/factory'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/components/ui/card/index'
import DataTable from '@/components/DataTable.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import { Plus, Edit2, Trash2, Warehouse, Package, Settings2, X } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const factoryStore = useFactoryStore()
const selectedFactoryId = ref('')
const showModal = ref(false)
const isEditing = ref(false)
const loading = ref(false)
const currentId = ref('')

const showStockModal = ref(false)
const selectedWarehouseForStock = ref<any>(null)

const showAdjustStock = ref(false)
const selectedStock = ref<any>(null)
const adjustSubmitting = ref(false)
const adjustStockForm = reactive({ quantity: 0, reason: '', factory_warehouse_id: '' })

const form = reactive({
  name: '',
  code: '',
  type: 'UMUM',
  description: '',
})

const columns = [
  { key: 'name', label: 'Informasi Gudang' },
  { key: 'type', label: 'Tipe Gudang' },
  { key: 'description', label: 'Deskripsi' },
  { key: 'actions', label: '' },
]

const stockColumns = [
  { key: 'product.name', label: 'Produk' },
  { key: 'batch_info', label: 'Batch & Exp' },
  { key: 'quantity', label: 'Stok' },
  { key: 'actions', label: '' },
]

const filteredWarehouses = computed(() => {
  return factoryStore.factoryWarehouses || []
})

const warehouseStocks = computed(() => {
  if (!selectedWarehouseForStock.value) return []
  return factoryStore.allInventory.filter(
    item => item.factory_warehouse_id === selectedWarehouseForStock.value.id
  )
})

onMounted(async () => {
  await Promise.all([
    factoryStore.fetchFactories(),
    factoryStore.fetchAllInventory()
  ])
  if (factoryStore.factories.length > 0 && factoryStore.factories[0]) {
    selectedFactoryId.value = factoryStore.factories[0].id
    onFactoryChange()
  }
})

const onFactoryChange = async () => {
  if (!selectedFactoryId.value) return
  await factoryStore.fetchFactoryWarehouses(selectedFactoryId.value)
}

const openCreateModal = () => {
  isEditing.value = false
  Object.assign(form, {
    name: '',
    code: '',
    type: 'UMUM',
    description: '',
  })
  showModal.value = true
}

const openEditModal = (item: any) => {
  isEditing.value = true
  currentId.value = item.id
  Object.assign(form, {
    name: item.name,
    code: item.code || '',
    type: item.type || 'UMUM',
    description: item.description || '',
  })
  showModal.value = true
}

const handleSubmit = async () => {
  if (!form.name) {
    toast.error('Nama Gudang harus diisi')
    return
  }
  if (!selectedFactoryId.value) {
    toast.error('Pabrik tidak terpilih')
    return
  }

  loading.value = true
  try {
    if (isEditing.value) {
      await factoryStore.updateFactoryWarehouse(selectedFactoryId.value, currentId.value, form)
      toast.success('Gudang berhasil diperbarui')
    } else {
      await factoryStore.createFactoryWarehouse(selectedFactoryId.value, form)
      toast.success('Gudang baru berhasil ditambahkan')
    }
    showModal.value = false
  } catch (err: any) {
    toast.error(err)
  } finally {
    loading.value = false
  }
}

const handleDelete = async (id: string) => {
  if (!confirm('Hapus gudang ini?')) return
  try {
    await factoryStore.deleteFactoryWarehouse(selectedFactoryId.value, id)
    toast.success('Gudang berhasil dihapus')
  } catch (err: any) {
    toast.error(err)
  }
}

const openStockModal = (warehouse: any) => {
  selectedWarehouseForStock.value = warehouse
  showStockModal.value = true
}

const openAdjustStock = (row: any) => {
  selectedStock.value = row
  adjustStockForm.quantity = row.quantity
  adjustStockForm.reason = ''
  adjustStockForm.factory_warehouse_id = row.factory_warehouse_id || ''
  
  if (row.factory_id) {
    factoryStore.fetchFactoryWarehouses(row.factory_id)
  }
  
  showAdjustStock.value = true
}

const handleAdjustStockSubmit = async () => {
  adjustSubmitting.value = true
  try {
    await factoryStore.adjustStock(selectedStock.value.factory_id, {
      product_id: selectedStock.value.product_id,
      quantity: adjustStockForm.quantity,
      reason: adjustStockForm.reason,
      batch_no: selectedStock.value.batch_no || 'DEFAULT',
      factory_warehouse_id: adjustStockForm.factory_warehouse_id || undefined,
    })
    await factoryStore.fetchAllInventory()
    showAdjustStock.value = false
    toast.success('Stok berhasil disesuaikan')
  } catch (err: any) {
    toast.error('Gagal menyesuaikan stok: ' + (err.response?.data?.error || err.message || err))
  } finally {
    adjustSubmitting.value = false
  }
}
</script>
