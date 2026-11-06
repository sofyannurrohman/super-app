<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRecipeStore, type Recipe } from '@/stores/recipe'
import { useFactoryStore } from '@/stores/factory'
import { toast } from 'vue-sonner'
import {
  ClipboardList,
  Plus,
  Trash2,
  Save,
  Search,
  RefreshCw,
  X,
  Package,
  ArrowRight,
  Sparkles,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import {
  formatMultiUnit,
  getCompatibleUnits,
  convertUnitQuantity,
} from '@/utils/productUnits'

const recipeStore = useRecipeStore()
const factoryStore = useFactoryStore()

const searchQuery = ref('')
const showModal = ref(false)
const loading = ref(false)

const newRecipe = ref<Recipe & { items: any[] }>({
  finished_product_id: '',
  description: '',
  items: [],
})

const filteredRecipes = computed(() => {
  return recipeStore.recipes.filter((r) => {
    const q = searchQuery.value.toLowerCase()
    return !q || r.finished_product?.name?.toLowerCase().includes(q)
  })
})

// availableProducts: only finished goods / sellable products (barang jadi yang bisa diproduksi)
const availableProducts = computed(() => factoryStore.sellableProducts)
// rawMaterials: only raw materials and packaging (bahan baku & kemasan untuk resep)
const rawMaterials = computed(() => factoryStore.rawMaterialProducts)

async function fetchInitialData() {
  await recipeStore.fetchRecipes()
  await factoryStore.fetchProducts()
}

function addIngredient() {
  newRecipe.value.items.push({
    raw_product_id: '',
    quantity: 1,
    unit: 'KG',
  })
}

function removeIngredient(index: number) {
  newRecipe.value.items.splice(index, 1)
}

function onIngredientProductChange(item: any) {
  const prod = factoryStore.products.find((p) => p.id === item.raw_product_id)
  if (prod) {
    item.unit = prod.unit || 'KG'
  }
}

function getIngredientCompatibleUnits(productId: string) {
  const prod = factoryStore.products.find((p) => p.id === productId)
  return getCompatibleUnits(prod?.unit || 'KG')
}

function getIngredientConversionHint(item: any) {
  if (!item.raw_product_id || !item.quantity) return ''
  const prod = factoryStore.products.find((p) => p.id === item.raw_product_id)
  if (!prod) return ''
  const baseUnit = prod.unit || 'KG'
  const currentUnit = item.unit || baseUnit
  if (currentUnit === baseUnit) return ''
  const converted = convertUnitQuantity(Number(item.quantity), currentUnit, baseUnit, prod.pcs_per_unit || 1)
  return `Konversi: ${item.quantity} ${currentUnit} ≈ ${converted.toLocaleString('id-ID', { maximumFractionDigits: 4 })} ${baseUnit} (Basis Stok)`
}

async function handleSave() {
  if (!newRecipe.value.finished_product_id || newRecipe.value.items.length === 0) {
    toast.error('Lengkapi data resep')
    return
  }

  loading.value = true
  try {
    const payload = {
      finished_product_id: newRecipe.value.finished_product_id,
      description: newRecipe.value.description,
      items: newRecipe.value.items.map((it: any) => {
        const prod = factoryStore.products.find((p) => p.id === it.raw_product_id)
        const baseUnit = prod?.unit || 'KG'
        const inputUnit = it.unit || baseUnit
        const converted = convertUnitQuantity(Number(it.quantity), inputUnit, baseUnit, prod?.pcs_per_unit || 1)
        return {
          raw_product_id: it.raw_product_id,
          quantity: converted,
        }
      }),
    }
    await recipeStore.createRecipe(payload as any)
    toast.success('Resep berhasil disimpan')
    showModal.value = false
    newRecipe.value = { finished_product_id: '', description: '', items: [] }
  } catch (e: any) {
    toast.error('Gagal menyimpan resep')
  } finally {
    loading.value = false
  }
}

async function handleDelete(id: string) {
  if (!confirm('Hapus resep ini?')) return
  await recipeStore.deleteRecipe(id)
  toast.success('Resep dihapus')
}

onMounted(fetchInitialData)
</script>

<template>
  <div class="space-y-8 p-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center gap-3">
          <ClipboardList class="w-8 h-8 text-indigo-600" />
          Resep Produksi (BOM)
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Definisikan bahan baku untuk setiap barang jadi untuk pemotongan stok otomatis.
        </p>
      </div>
      <Button
        @click="showModal = true"
        class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-2xl shadow-lg shadow-indigo-500/20 px-6 py-6 font-black text-xs"
      >
        <Plus class="w-4 h-4 mr-2" /> Buat Resep Baru
      </Button>
    </div>

    <!-- Search -->
    <div class="relative max-w-md">
      <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Cari barang jadi..."
        class="w-full pl-11 pr-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-medium outline-none focus:ring-2 focus:ring-indigo-500/20"
      />
    </div>

    <!-- List -->
    <div v-if="recipeStore.loading" class="flex justify-center py-24">
      <RefreshCw class="w-8 h-8 animate-spin text-indigo-600" />
    </div>

    <div
      v-else-if="filteredRecipes.length === 0"
      class="text-center py-24 bg-slate-50 rounded-[40px] border-2 border-dashed border-slate-200"
    >
      <Package class="w-12 h-12 text-slate-300 mx-auto mb-4" />
      <h3 class="text-lg font-black text-slate-900">Belum Ada Resep</h3>
      <p class="text-slate-500 text-sm">
        Klik tombol di atas untuk mulai mendefinisikan resep produksi.
      </p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="recipe in filteredRecipes"
        :key="recipe.id"
        class="bg-white border border-slate-200 rounded-[32px] p-8 shadow-sm hover:shadow-xl transition-all relative group"
      >
        <button
          @click="handleDelete(recipe.id!)"
          class="absolute top-6 right-6 p-2 text-slate-300 hover:text-red-500 transition-colors opacity-0 group-hover:opacity-100"
        >
          <Trash2 class="w-5 h-5" />
        </button>

        <div class="bg-indigo-50 w-12 h-12 rounded-2xl flex items-center justify-center mb-6">
          <Package class="w-6 h-6 text-indigo-600" />
        </div>

        <h4 class="text-xl font-black text-slate-900 mb-1">{{ recipe.finished_product?.name }}</h4>
        <p class="text-xs text-slate-500 font-medium mb-6 uppercase tracking-widest">
          {{ recipe.finished_product?.sku }}
        </p>

          <div class="space-y-4 pt-6 border-t border-slate-100">
          <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
            Kebutuhan Bahan per Unit:
          </p>
          <div
            v-for="item in recipe.items"
            :key="item.id"
            class="flex items-center justify-between text-xs font-bold text-slate-700 bg-slate-50 px-4 py-3 rounded-xl"
          >
            <span>{{ item.raw_product?.name }}</span>
            <span class="text-indigo-600 font-black">
              {{ formatMultiUnit(item.quantity, item.raw_product?.unit, item.raw_product?.pcs_per_unit).summary }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showModal = false" />
      <div
        class="bg-white rounded-[40px] w-full max-w-2xl relative z-10 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]"
      >
        <div class="p-8 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-2xl font-black text-slate-900">Definisikan Resep</h3>
          <button
            @click="showModal = false"
            class="p-2 hover:bg-slate-100 rounded-xl text-slate-400"
          >
            <X class="w-6 h-6" />
          </button>
        </div>

        <div class="p-8 overflow-y-auto space-y-8 flex-1">
          <!-- Step 1: Finished Product -->
          <div class="space-y-4">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Barang Jadi yang Dihasilkan</label
            >
            <select
              v-model="newRecipe.finished_product_id"
              class="w-full bg-slate-50 border-2 border-transparent focus:border-indigo-500 focus:bg-white rounded-2xl px-5 py-4 font-bold text-slate-900 outline-none transition-all appearance-none"
            >
              <option value="" disabled>Pilih Produk Jadi...</option>
              <option v-for="p in availableProducts" :key="p.id" :value="p.id">
                {{ p.name }} ({{ p.sku }})
              </option>
            </select>
          </div>

          <!-- Step 2: Ingredients -->
          <div class="space-y-4">
            <div class="flex items-center justify-between ml-1">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest"
                >Bahan & Kemasan (Input)</label
              >
              <button
                @click="addIngredient"
                class="text-indigo-600 font-black text-xs hover:underline flex items-center gap-1"
              >
                <Plus class="w-4 h-4" /> Tambah Bahan
              </button>
            </div>

            <div class="space-y-3">
              <div
                v-for="(item, index) in newRecipe.items"
                :key="index"
                class="p-3 bg-slate-50/80 rounded-2xl border border-slate-100 space-y-2 animate-in fade-in slide-in-from-top-2 duration-300"
              >
                <div class="flex gap-3 items-center">
                  <div class="flex-1">
                    <select
                      v-model="item.raw_product_id"
                      @change="onIngredientProductChange(item)"
                      class="w-full bg-white border border-slate-200 focus:border-indigo-500 rounded-xl px-3 py-2.5 font-bold text-slate-900 outline-none transition-all text-xs"
                    >
                      <option value="" disabled>Pilih Bahan Baku...</option>
                      <option v-for="p in rawMaterials" :key="p.id" :value="p.id">
                        {{ p.sku }} - {{ p.name }} (Basis: {{ p.unit || 'KG' }})
                      </option>
                    </select>
                  </div>
                  <div class="w-24">
                    <input
                      v-model="item.quantity"
                      type="number"
                      step="any"
                      min="0.0001"
                      placeholder="Qty"
                      class="w-full bg-white border border-slate-200 focus:border-indigo-500 rounded-xl px-3 py-2.5 font-mono font-bold text-slate-900 outline-none transition-all text-xs"
                    />
                  </div>
                  <div class="w-24">
                    <select
                      v-model="item.unit"
                      class="w-full bg-white border border-slate-200 focus:border-indigo-500 rounded-xl px-2 py-2.5 font-bold text-slate-900 outline-none transition-all text-xs uppercase"
                    >
                      <option v-for="u in getIngredientCompatibleUnits(item.raw_product_id)" :key="u" :value="u">{{ u }}</option>
                    </select>
                  </div>
                  <button
                    @click="removeIngredient(index)"
                    class="p-2.5 text-red-400 hover:text-red-600 transition-colors bg-red-50 rounded-xl"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>

                <!-- Conversion Live Preview -->
                <div
                  v-if="getIngredientConversionHint(item)"
                  class="text-[10px] text-amber-700 font-medium px-1 flex items-center gap-1"
                >
                  <Sparkles class="w-3 h-3 text-amber-600" />
                  <span>{{ getIngredientConversionHint(item) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="p-8 border-t border-slate-100 bg-slate-50/50">
          <Button
            @click="handleSave"
            :disabled="loading"
            class="w-full bg-slate-900 hover:bg-indigo-600 text-white rounded-2xl py-7 font-black shadow-xl transition-all"
          >
            {{ loading ? 'Menyimpan...' : 'Simpan Resep Produksi' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
