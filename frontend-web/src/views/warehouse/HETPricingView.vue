<template>
  <div class="p-6 space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Kelola Harga HET Pelanggan</h1>
        <p class="text-muted-foreground">Kelola Harga Eceran Tertinggi (HET) produk untuk pelanggan retail/mobile.</p>
      </div>
    </div>

    <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
      <CardContent class="p-6">
        <div class="flex items-center gap-4 mb-6">
          <div class="relative flex-1 max-w-sm">
            <Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari produk berdasarkan nama atau SKU..."
              class="w-full pl-11 pr-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-4 focus:ring-primary/5 focus:border-primary/20 transition-all"
            />
          </div>
        </div>

        <DataTable :columns="columns" :data="filteredProducts" class="border-none">
          <template #cell-sku="{ row }">
            <span
              class="font-mono text-[11px] font-bold text-primary bg-primary/5 px-2 py-1 rounded-lg ring-1 ring-primary/10"
            >
              {{ (row as any).sku }}
            </span>
          </template>
          <template #cell-image="{ row }">
            <div
              class="h-12 w-12 rounded-xl bg-slate-100 overflow-hidden flex items-center justify-center border border-slate-100 shadow-sm"
            >
              <img
                v-if="(row as any).image_url"
                :src="`${apiBaseUrl}/v1${(row as any).image_url}`"
                class="h-full w-full object-cover"
              />
              <Package v-else class="h-6 w-6 text-slate-300" />
            </div>
          </template>
          <template #cell-product_info="{ row }">
            <div class="flex flex-col">
              <span class="font-bold text-slate-900">{{ (row as any).name }}</span>
              <div class="flex items-center gap-2 mt-0.5">
                <span class="text-xs bg-slate-100 text-slate-500 px-1.5 py-0.5 rounded font-bold uppercase">
                  {{ (row as any).category || 'No Category' }}
                </span>
                <span class="text-[10px] text-slate-400 font-medium">
                  {{ (row as any).brand || 'Tanpa Brand' }}
                </span>
              </div>
            </div>
          </template>
          <template #cell-het_price="{ row }">
            <div class="font-mono text-sm font-black text-emerald-600">
              Rp{{ (row as any).selling_price?.toLocaleString() || '0' }}
            </div>
          </template>
          <template #cell-actions="{ row }">
            <Button
              variant="outline"
              size="sm"
              @click="openEditModal(row)"
              class="rounded-xl border-emerald-100 text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700"
            >
              <Edit2 class="mr-1.5 h-3.5 w-3.5" /> Atur HET
            </Button>
          </template>
        </DataTable>
      </CardContent>
    </Card>

    <!-- Edit HET Modal -->
    <div
      v-if="showModal"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      @click.self="showModal = false"
    >
      <Card class="w-full max-w-md border-none shadow-2xl rounded-[2.5rem] bg-white overflow-hidden animate-in fade-in zoom-in duration-200">
        <CardHeader class="p-8 pb-4">
          <div class="flex items-center gap-4">
            <div class="h-12 w-12 rounded-2xl bg-emerald-50 flex items-center justify-center text-emerald-500">
              <DollarSign class="h-6 w-6" />
            </div>
            <div>
              <CardTitle class="text-xl font-black text-slate-900">Atur Harga HET Pelanggan</CardTitle>
              <CardDescription class="text-slate-400 text-xs mt-0.5 font-bold uppercase tracking-wider">Update Harga Retail</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6">
          <div class="bg-slate-50 p-4 rounded-2xl flex gap-3 items-center">
            <div class="h-10 w-10 rounded-xl bg-white overflow-hidden flex items-center justify-center border border-slate-100 shadow-sm">
              <img
                v-if="selectedProduct.image_url"
                :src="`${apiBaseUrl}/v1${selectedProduct.image_url}`"
                class="h-full w-full object-cover"
              />
              <Package v-else class="h-5 w-5 text-slate-300" />
            </div>
            <div>
              <div class="font-bold text-slate-900 text-sm">{{ selectedProduct.name }}</div>
              <div class="text-[10px] font-mono text-slate-400 font-bold uppercase mt-0.5">{{ selectedProduct.sku }}</div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Harga HET Pelanggan (Selling Price)</label>
            <div class="relative group">
              <span class="absolute left-5 top-1/2 -translate-y-1/2 text-xs font-bold text-emerald-600/50 group-focus-within:text-emerald-600 transition-colors">Rp</span>
              <input
                v-model="hetPriceInput"
                type="number"
                class="w-full pl-12 pr-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-base font-black text-emerald-700 outline-none focus:ring-4 focus:ring-emerald-500/10 focus:border-emerald-500/20"
                placeholder="0"
              />
            </div>
            <p class="text-[11px] text-slate-400 ml-1">Harga ini akan digunakan sebagai harga jual di aplikasi katalog mobile pelanggan.</p>
          </div>
        </CardContent>
        <CardFooter class="p-8 pt-0 flex gap-3">
          <Button
            variant="ghost"
            class="flex-1 rounded-2xl h-12 text-slate-500 font-bold"
            @click="showModal = false"
            :disabled="loading"
          >
            Batal
          </Button>
          <Button
            class="flex-1 rounded-2xl h-12 bg-emerald-600 hover:bg-emerald-700 text-white font-bold"
            @click="handleSave"
            :disabled="loading"
          >
            <span v-if="loading">Menyimpan...</span>
            <span v-else>Simpan Perubahan</span>
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useFactoryStore } from '@/stores/factory'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card/index'
import DataTable from '@/components/DataTable.vue'
import { Search, Edit2, Package, DollarSign } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const factoryStore = useFactoryStore()
const searchQuery = ref('')
const showModal = ref(false)
const loading = ref(false)
const selectedProduct = ref<any>({})
const hetPriceInput = ref<number>(0)
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL

const columns = [
  { key: 'image', label: '' },
  { key: 'sku', label: 'SKU' },
  {
    id: 'product_info',
    header: 'Produk & Kategori',
    accessorFn: (row: any) => `${row.name || ''} ${row.category || ''} ${row.brand || ''}`,
  },
  { key: 'het_price', label: 'Harga HET Pelanggan' },
  { key: 'actions', label: '' },
]

const filteredProducts = computed(() => {
  if (!searchQuery.value) return factoryStore.sellableProducts
  const q = searchQuery.value.toLowerCase()
  return factoryStore.sellableProducts.filter(
    (p) => p.name.toLowerCase().includes(q) || p.sku.toLowerCase().includes(q),
  )
})

onMounted(() => {
  factoryStore.fetchProducts()
})

const openEditModal = (item: any) => {
  selectedProduct.value = item
  hetPriceInput.value = item.selling_price || 0
  showModal.value = true
}

const handleSave = async () => {
  loading.value = true
  try {
    const formData = new FormData()
    formData.append('sku', selectedProduct.value.sku)
    formData.append('name', selectedProduct.value.name)
    formData.append('unit', selectedProduct.value.unit)
    formData.append('weight', (selectedProduct.value.weight || 0).toString())
    formData.append('weight_unit', selectedProduct.value.weight_unit || 'KG')
    formData.append('cost_price', (selectedProduct.value.cost_price || 0).toString())
    formData.append('factory_price', (selectedProduct.value.factory_price || 0).toString())
    formData.append('selling_price', hetPriceInput.value.toString())
    formData.append('category', selectedProduct.value.category || '')
    formData.append('brand', selectedProduct.value.brand || '')
    formData.append('specs', selectedProduct.value.specs || '')
    formData.append('description', selectedProduct.value.description || '')

    await factoryStore.updateProduct(selectedProduct.value.id, formData)
    toast.success('Harga HET berhasil diperbarui')
    showModal.value = false
    await factoryStore.fetchProducts()
  } catch (err: any) {
    toast.error('Gagal memperbarui harga HET: ' + (err.response?.data?.error || err.message))
  } finally {
    loading.value = false
  }
}
</script>
