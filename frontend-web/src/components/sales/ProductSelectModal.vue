<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import {
  Search,
  X,
  Package,
  Check,
  Plus,
  Minus,
  CheckCircle2,
  AlertCircle,
  Boxes,
  SlidersHorizontal,
  Tag
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { filterSellableProducts } from '@/utils/productUnits'

const props = withDefaults(
  defineProps<{
    show: boolean
    products: any[]
    selectedBranchName?: string
    getStock?: (productId: string) => number | string
    mode?: 'single' | 'multi'
    currentSelectedId?: string
    alreadySelectedIds?: string[]
    excludeRawMaterials?: boolean
  }>(),
  {
    mode: 'single',
    currentSelectedId: '',
    alreadySelectedIds: () => [],
    excludeRawMaterials: true,
    selectedBranchName: 'Cabang Terpilih',
  }
)

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', product: any): void
  (e: 'select-multiple', selectedItems: { product: any; quantity: number }[]): void
}>()

const searchQuery = ref('')
const selectedCategory = ref<string>('ALL')
const stockFilter = ref<'ALL' | 'IN_STOCK' | 'OUT_OF_STOCK'>('ALL')
const searchInputRef = ref<HTMLInputElement | null>(null)

// For multi-select mode: map of productId -> quantity
const multiSelectedMap = ref<Record<string, number>>({})

const formatCurrency = (v: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(v || 0)

const baseProducts = computed(() => {
  if (!props.products) return []
  if (props.excludeRawMaterials) {
    return filterSellableProducts(props.products)
  }
  return props.products
})

// Extract dynamic categories from sellable products
const availableCategories = computed(() => {
  const cats = new Set<string>()
  baseProducts.value.forEach((p) => {
    if (p.category && p.category.trim()) {
      cats.add(p.category.trim())
    }
  })
  return Array.from(cats)
})

const getProductStock = (productId: string): number => {
  if (!props.getStock) return 0
  const val = props.getStock(productId)
  return typeof val === 'number' ? val : Number(val || 0)
}

const filteredProducts = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()

  return baseProducts.value.filter((p) => {
    // 1. Search Query
    if (query) {
      const nameMatch = p.name?.toLowerCase().includes(query)
      const skuMatch = p.sku?.toLowerCase().includes(query)
      const catMatch = p.category?.toLowerCase().includes(query)
      const priceStr = p.selling_price?.toString() || ''
      const priceMatch = priceStr.includes(query)
      if (!nameMatch && !skuMatch && !catMatch && !priceMatch) {
        return false
      }
    }

    // 2. Category Filter
    if (selectedCategory.value !== 'ALL') {
      if (p.category !== selectedCategory.value) {
        return false
      }
    }

    // 3. Stock Filter
    if (props.getStock) {
      const stock = getProductStock(p.id)
      if (stockFilter.value === 'IN_STOCK' && stock <= 0) {
        return false
      }
      if (stockFilter.value === 'OUT_OF_STOCK' && stock > 0) {
        return false
      }
    }

    return true
  })
})

const totalSelectedCount = computed(() => {
  return Object.values(multiSelectedMap.value).filter((qty) => qty > 0).length
})

const isProductMultiSelected = (productId: string) => {
  return (multiSelectedMap.value[productId] || 0) > 0
}

const toggleMultiSelect = (product: any) => {
  if (isProductMultiSelected(product.id)) {
    delete multiSelectedMap.value[product.id]
  } else {
    multiSelectedMap.value[product.id] = 1
  }
}

const updateMultiQty = (productId: string, delta: number) => {
  const current = multiSelectedMap.value[productId] || 0
  const next = current + delta
  if (next <= 0) {
    delete multiSelectedMap.value[productId]
  } else {
    multiSelectedMap.value[productId] = next
  }
}

const setMultiQty = (productId: string, val: number) => {
  if (val <= 0) {
    delete multiSelectedMap.value[productId]
  } else {
    multiSelectedMap.value[productId] = val
  }
}

const onSelectSingle = (product: any) => {
  emit('select', product)
  emit('close')
}

const onConfirmMultiSelect = () => {
  const results: { product: any; quantity: number }[] = []
  Object.entries(multiSelectedMap.value).forEach(([prodId, qty]) => {
    if (qty > 0) {
      const prod = baseProducts.value.find((p) => p.id === prodId)
      if (prod) {
        results.push({ product: prod, quantity: qty })
      }
    }
  })
  emit('select-multiple', results)
  emit('close')
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      searchQuery.value = ''
      selectedCategory.value = 'ALL'
      stockFilter.value = 'ALL'
      multiSelectedMap.value = {}
      nextTick(() => {
        searchInputRef.value?.focus()
      })
    }
  }
)
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 overflow-y-auto bg-black/60 backdrop-blur-xs flex items-center justify-center p-2 sm:p-4 animate-in fade-in duration-200"
  >
    <div
      class="relative bg-white rounded-2xl sm:rounded-3xl shadow-2xl max-w-5xl w-full flex flex-col h-[94vh] sm:h-[88vh] overflow-hidden border border-slate-200"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-5 sm:px-6 py-4 border-b border-slate-100 bg-slate-50/80 shrink-0">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-primary/10 rounded-2xl text-primary font-black shadow-xs">
            <Boxes class="w-6 h-6" />
          </div>
          <div>
            <h3 class="font-extrabold text-slate-900 text-base sm:text-lg leading-tight flex items-center gap-2">
              Katalog & Pemilih Produk
              <span
                v-if="mode === 'multi'"
                class="text-[11px] font-bold bg-primary/10 text-primary px-2.5 py-0.5 rounded-full"
              >
                Mode Pilih Banyak (Multi-Select)
              </span>
            </h3>
            <p class="text-xs text-slate-500 font-medium">
              Stok Gudang: <strong class="text-slate-800 underline">{{ selectedBranchName }}</strong> • {{ baseProducts.length }} Produk Tersedia
            </p>
          </div>
        </div>
        <Button
          variant="outline"
          size="icon"
          @click="emit('close')"
          class="rounded-xl hover:bg-slate-200/70 h-9 w-9 border-slate-200 transition-colors"
        >
          <X class="w-4 h-4 text-slate-600" />
        </Button>
      </div>

      <!-- Search & Filters Toolbar -->
      <div class="p-4 sm:px-6 border-b border-slate-100 bg-white space-y-3 shrink-0">
        <div class="flex flex-col sm:flex-row gap-2.5">
          <!-- Search Input -->
          <div class="relative flex-1">
            <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              placeholder="Cari berdasarkan nama barang, SKU/kode, atau harga..."
              class="w-full pl-10 pr-9 py-2.5 bg-slate-50 hover:bg-slate-100/60 focus:bg-white border border-slate-200 rounded-xl text-xs sm:text-sm font-semibold text-slate-800 placeholder:text-slate-400 focus:ring-2 focus:ring-primary/20 focus:border-primary outline-none transition-all"
            />
            <button
              v-if="searchQuery"
              @click="searchQuery = ''"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5 rounded-full"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>

          <!-- Stock Filter Toggle -->
          <div v-if="getStock" class="flex items-center rounded-xl bg-slate-100 p-1 border border-slate-200 shrink-0">
            <button
              type="button"
              @click="stockFilter = 'ALL'"
              :class="[
                'px-3 py-1.5 text-xs font-bold rounded-lg transition-all whitespace-nowrap',
                stockFilter === 'ALL' ? 'bg-white text-slate-900 shadow-xs' : 'text-slate-500 hover:text-slate-800'
              ]"
            >
              Semua
            </button>
            <button
              type="button"
              @click="stockFilter = 'IN_STOCK'"
              :class="[
                'px-3 py-1.5 text-xs font-bold rounded-lg transition-all whitespace-nowrap flex items-center gap-1',
                stockFilter === 'IN_STOCK' ? 'bg-emerald-500 text-white shadow-xs' : 'text-slate-500 hover:text-slate-800'
              ]"
            >
              <span class="w-2 h-2 rounded-full bg-emerald-300"></span> Ada Stok
            </button>
            <button
              type="button"
              @click="stockFilter = 'OUT_OF_STOCK'"
              :class="[
                'px-3 py-1.5 text-xs font-bold rounded-lg transition-all whitespace-nowrap',
                stockFilter === 'OUT_OF_STOCK' ? 'bg-rose-500 text-white shadow-xs' : 'text-slate-500 hover:text-slate-800'
              ]"
            >
              Stok Kosong
            </button>
          </div>
        </div>

        <!-- Category Filter Pills (Horizontal Scroll) -->
        <div v-if="availableCategories.length > 0" class="flex items-center gap-1.5 overflow-x-auto no-scrollbar pb-0.5">
          <span class="text-[10px] font-black uppercase text-slate-400 tracking-wider mr-1 shrink-0 flex items-center gap-1">
            <Tag class="w-3 h-3" /> Kategori:
          </span>
          <button
            type="button"
            @click="selectedCategory = 'ALL'"
            :class="[
              'px-3 py-1 text-xs font-bold rounded-full transition-all whitespace-nowrap shrink-0 border',
              selectedCategory === 'ALL'
                ? 'bg-slate-900 text-white border-slate-900 shadow-xs'
                : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-50'
            ]"
          >
            Semua ({{ baseProducts.length }})
          </button>
          <button
            v-for="cat in availableCategories"
            :key="cat"
            type="button"
            @click="selectedCategory = cat"
            :class="[
              'px-3 py-1 text-xs font-bold rounded-full transition-all whitespace-nowrap shrink-0 border',
              selectedCategory === cat
                ? 'bg-primary text-white border-primary shadow-xs'
                : 'bg-white text-slate-600 border-slate-200 hover:bg-slate-50'
            ]"
          >
            {{ cat }}
          </button>
        </div>
      </div>

      <!-- Products Grid List -->
      <div class="p-4 sm:p-6 overflow-y-auto flex-1 min-h-0 bg-slate-50/50">
        <!-- Empty State -->
        <div
          v-if="filteredProducts.length === 0"
          class="flex flex-col items-center justify-center h-64 text-center p-6 bg-white rounded-2xl border border-dashed border-slate-200"
        >
          <div class="w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center text-slate-400 mb-3">
            <Package class="w-6 h-6" />
          </div>
          <h4 class="font-bold text-slate-800 text-sm mb-1">Produk Tidak Ditemukan</h4>
          <p class="text-xs text-slate-500 max-w-xs">
            Coba sesuaikan kata kunci pencarian atau ubah filter kategori/stok.
          </p>
          <Button
            v-if="searchQuery || selectedCategory !== 'ALL' || stockFilter !== 'ALL'"
            variant="outline"
            size="sm"
            @click="searchQuery = ''; selectedCategory = 'ALL'; stockFilter = 'ALL'"
            class="mt-3 text-xs font-bold"
          >
            Reset Semua Filter
          </Button>
        </div>

        <!-- Product Cards Grid -->
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
          <div
            v-for="p in filteredProducts"
            :key="p.id"
            @click="mode === 'single' ? onSelectSingle(p) : toggleMultiSelect(p)"
            :class="[
              'group relative bg-white rounded-2xl p-4 border transition-all duration-150 flex flex-col justify-between cursor-pointer select-none',
              mode === 'single' && currentSelectedId === p.id
                ? 'border-primary ring-2 ring-primary/20 bg-primary/5 shadow-sm'
                : mode === 'multi' && isProductMultiSelected(p.id)
                ? 'border-primary ring-2 ring-primary/20 bg-primary/5 shadow-sm'
                : alreadySelectedIds.includes(p.id)
                ? 'border-slate-200 bg-slate-50/70 hover:border-slate-300'
                : 'border-slate-200 hover:border-primary/50 hover:shadow-md'
            ]"
          >
            <!-- Card Header: Title, SKU & Category -->
            <div>
              <div class="flex items-start justify-between gap-2 mb-2">
                <div class="flex-1 min-w-0">
                  <h4 class="font-black text-slate-900 text-sm group-hover:text-primary transition-colors leading-snug">
                    {{ p.name }}
                  </h4>
                  <div class="flex items-center gap-2 mt-1 flex-wrap">
                    <span class="text-[10px] font-extrabold px-2 py-0.5 rounded-md bg-slate-100 text-slate-600">
                      SKU: {{ p.sku || '-' }}
                    </span>
                    <span v-if="p.category" class="text-[10px] font-bold px-2 py-0.5 rounded-md bg-blue-50 text-blue-700">
                      {{ p.category }}
                    </span>
                    <span v-if="alreadySelectedIds.includes(p.id)" class="text-[10px] font-black px-1.5 py-0.5 rounded bg-amber-100 text-amber-800">
                      Sudah di Tabel
                    </span>
                  </div>
                </div>

                <!-- Single Select Check Badge -->
                <div v-if="mode === 'single'">
                  <div
                    v-if="currentSelectedId === p.id"
                    class="w-6 h-6 rounded-full bg-primary text-white flex items-center justify-center shrink-0 shadow-xs"
                  >
                    <Check class="w-3.5 h-3.5" />
                  </div>
                </div>

                <!-- Multi-select checkbox -->
                <div v-else-if="mode === 'multi'">
                  <div
                    :class="[
                      'w-6 h-6 rounded-lg border flex items-center justify-center transition-all shrink-0',
                      isProductMultiSelected(p.id)
                        ? 'bg-primary border-primary text-white shadow-xs'
                        : 'border-slate-300 bg-white hover:border-slate-400'
                    ]"
                  >
                    <Check v-if="isProductMultiSelected(p.id)" class="w-4 h-4 stroke-[3]" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Card Footer: Stock & Price Info -->
            <div class="pt-3 mt-3 border-t border-slate-100 flex items-end justify-between gap-2">
              <div>
                <span class="text-[10px] font-bold uppercase text-slate-400 block tracking-wider">Harga Jual</span>
                <span class="text-sm sm:text-base font-black text-slate-900">
                  {{ formatCurrency(p.selling_price || p.factory_price || p.cost_price || 0) }}
                </span>
                <span class="text-[11px] text-slate-500 font-medium"> / {{ p.unit || 'Pcs' }}</span>
              </div>

              <!-- Real-time Stock Badge -->
              <div v-if="getStock" class="text-right">
                <span class="text-[10px] font-bold uppercase text-slate-400 block tracking-wider">Stok Gudang</span>
                <span
                  :class="[
                    'inline-flex items-center gap-1 text-xs font-black px-2.5 py-0.5 rounded-lg border mt-0.5',
                    getProductStock(p.id) > 0
                      ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                      : 'bg-rose-50 text-rose-700 border-rose-200'
                  ]"
                >
                  <span
                    :class="[
                      'w-1.5 h-1.5 rounded-full',
                      getProductStock(p.id) > 0 ? 'bg-emerald-500' : 'bg-rose-500'
                    ]"
                  ></span>
                  {{ getProductStock(p.id) }} {{ p.unit || 'Pcs' }}
                </span>
              </div>
            </div>

            <!-- Multi-select Quantity Stepper (Inside Card when selected) -->
            <div
              v-if="mode === 'multi' && isProductMultiSelected(p.id)"
              @click.stop
              class="mt-3 pt-2.5 border-t border-primary/20 flex items-center justify-between bg-primary/10 -mx-4 -mb-4 px-4 py-2.5 rounded-b-2xl animate-in fade-in duration-150"
            >
              <span class="text-xs font-black text-primary">Jumlah Masuk (Qty):</span>
              <div class="flex items-center gap-1.5">
                <button
                  type="button"
                  @click="updateMultiQty(p.id, -1)"
                  class="w-7 h-7 rounded-lg bg-white border border-slate-200 hover:bg-slate-100 text-slate-700 flex items-center justify-center font-bold text-xs transition-colors shadow-xs"
                >
                  <Minus class="w-3.5 h-3.5" />
                </button>
                <input
                  type="number"
                  min="1"
                  :value="multiSelectedMap[p.id] || 1"
                  @input="(e: any) => setMultiQty(p.id, Number(e.target.value))"
                  class="w-12 bg-white border border-slate-300 rounded-lg py-1 text-center text-xs font-black text-slate-900 outline-none focus:ring-2 focus:ring-primary/30"
                />
                <button
                  type="button"
                  @click="updateMultiQty(p.id, 1)"
                  class="w-7 h-7 rounded-lg bg-white border border-slate-200 hover:bg-slate-100 text-slate-700 flex items-center justify-center font-bold text-xs transition-colors shadow-xs"
                >
                  <Plus class="w-3.5 h-3.5" />
                </button>
                <span class="text-xs font-bold text-slate-600 ml-1">{{ p.unit || 'Pcs' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Bottom Action Bar (in Multi-select mode) -->
      <div
        v-if="mode === 'multi'"
        class="px-5 sm:px-6 py-3.5 bg-slate-50 border-t border-slate-200 flex items-center justify-between gap-3 shrink-0"
      >
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-xl bg-primary text-white font-black flex items-center justify-center text-xs shadow-xs">
            {{ totalSelectedCount }}
          </div>
          <span class="text-xs sm:text-sm font-bold text-slate-700">
            Produk terpilih untuk ditambahkan
          </span>
        </div>

        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" @click="emit('close')" class="text-xs font-bold">
            Batal
          </Button>
          <Button
            size="sm"
            @click="onConfirmMultiSelect"
            :disabled="totalSelectedCount === 0"
            class="text-xs font-bold gap-1.5 shadow-md"
          >
            <CheckCircle2 class="w-4 h-4" /> Masukkan ke Faktur ({{ totalSelectedCount }})
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
