<script setup lang="ts">
import { ref, computed } from 'vue'
import { Check, ChevronsUpDown, Search, Boxes, ExternalLink, X } from 'lucide-vue-next'
import { onClickOutside } from '@vueuse/core'
import { filterSellableProducts } from '@/utils/productUnits'

const props = withDefaults(
  defineProps<{
    modelValue: string
    products: any[]
    disabled?: boolean
    excludeRawMaterials?: boolean
    getStock?: (productId: string) => number | string
    showCatalogButton?: boolean
    placeholder?: string
    fallbackLabel?: string
  }>(),
  {
    disabled: false,
    excludeRawMaterials: true,
    showCatalogButton: true,
    placeholder: 'Pilih Produk...',
    fallbackLabel: '',
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
  (e: 'open-catalog'): void
}>()

const isOpen = ref(false)
const searchQuery = ref('')
const containerRef = ref<HTMLElement | null>(null)

onClickOutside(containerRef, () => {
  isOpen.value = false
})

const formatCurrency = (v: number) =>
  new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(v || 0)

const baseProductList = computed(() => {
  if (!props.products) return []
  if (props.excludeRawMaterials) {
    return filterSellableProducts(props.products)
  }
  return props.products
})

const filteredProducts = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return baseProductList.value

  return baseProductList.value.filter((p) => {
    const nameMatch = p.name?.toLowerCase().includes(query)
    const skuMatch = p.sku?.toLowerCase().includes(query)
    const catMatch = p.category?.toLowerCase().includes(query)
    const priceStr = p.selling_price?.toString() || ''
    const priceMatch = priceStr.includes(query)

    return nameMatch || skuMatch || catMatch || priceMatch
  })
})

const selectedProduct = computed(() => {
  return props.products?.find((p) => p.id === props.modelValue)
})

function selectProduct(id: string) {
  emit('update:modelValue', id)
  emit('change', id)
  isOpen.value = false
  searchQuery.value = ''
}

function toggleDropdown() {
  if (props.disabled) return
  isOpen.value = !isOpen.value
}

function handleOpenCatalog() {
  isOpen.value = false
  emit('open-catalog')
}
</script>

<template>
  <div class="relative w-full" ref="containerRef">
    <!-- Trigger Field -->
    <div
      @click="toggleDropdown"
      class="flex items-center justify-between w-full bg-white border border-slate-200 rounded-xl px-3 py-2 text-xs font-bold cursor-pointer hover:border-slate-300 hover:bg-slate-50/50 transition-all focus:ring-2 focus:ring-primary/20 shadow-xs"
      :class="{
        'opacity-50 cursor-not-allowed pointer-events-none': disabled,
        'border-primary/50 ring-2 ring-primary/20 bg-white': isOpen
      }"
    >
      <div class="truncate flex-1 min-w-0 pr-2">
        <template v-if="selectedProduct">
          <div class="flex items-center gap-1.5 truncate">
            <span class="truncate font-black text-slate-800">{{ selectedProduct.name }}</span>
            <span class="text-primary font-bold shrink-0">({{ formatCurrency(selectedProduct.selling_price) }})</span>
            <span
              v-if="getStock || selectedProduct.stock !== undefined"
              class="px-1.5 py-0.2 text-[10px] rounded-md font-bold shrink-0"
              :class="
                (getStock ? Number(getStock(selectedProduct.id)) : Number(selectedProduct.stock || 0)) > 0
                  ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                  : 'bg-rose-50 text-rose-700 border border-rose-200'
              "
            >
              Stok: {{ getStock ? getStock(selectedProduct.id) : selectedProduct.stock }}
            </span>
          </div>
        </template>
        <template v-else-if="fallbackLabel">
          <div class="flex items-center gap-1.5 truncate">
            <span class="truncate font-black text-slate-800">{{ fallbackLabel }}</span>
            <span class="px-1.5 py-0.2 text-[9px] rounded-md bg-amber-50 text-amber-700 border border-amber-200 font-bold shrink-0">
              Manual / Lama
            </span>
          </div>
        </template>
        <template v-else>
          <span class="text-slate-400 font-semibold flex items-center gap-1.5">
            <Boxes class="w-3.5 h-3.5 text-slate-400" />
            {{ placeholder }}
          </span>
        </template>
      </div>
      <div class="flex items-center gap-1 shrink-0 text-slate-400">
        <ChevronsUpDown class="w-3.5 h-3.5" />
      </div>
    </div>

    <!-- Dropdown Panel (Generous Width & Min Width) -->
    <div
      v-if="isOpen && !disabled"
      class="absolute z-50 left-0 mt-1.5 w-full min-w-[320px] sm:min-w-[420px] max-w-lg bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden flex flex-col animate-in fade-in zoom-in-95 duration-150"
    >
      <!-- Search & Catalog Header -->
      <div class="p-2.5 border-b border-slate-100 bg-slate-50/80 space-y-2">
        <div class="flex items-center gap-2">
          <div class="relative flex-1">
            <Search class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Cari nama, SKU, kategori, atau harga..."
              class="w-full pl-8 pr-7 py-1.5 text-xs font-semibold rounded-lg bg-white border border-slate-200 outline-none focus:border-primary focus:ring-1 focus:ring-primary/20"
              @click.stop
            />
            <button
              v-if="searchQuery"
              @click="searchQuery = ''"
              class="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5"
            >
              <X class="w-3 h-3" />
            </button>
          </div>

          <!-- Quick Open Full Modal Button -->
          <button
            v-if="showCatalogButton"
            type="button"
            @click.stop="handleOpenCatalog"
            class="flex items-center gap-1 px-2.5 py-1.5 bg-primary/10 hover:bg-primary/20 text-primary rounded-lg text-xs font-bold transition-colors shrink-0"
            title="Buka Katalog Lengkap"
          >
            <Boxes class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">Katalog</span>
            <ExternalLink class="w-3 h-3" />
          </button>
        </div>
      </div>

      <!-- Options List -->
      <div class="overflow-y-auto flex-1 p-1.5 max-h-72 divide-y divide-slate-50">
        <div v-if="filteredProducts.length === 0" class="p-6 text-center text-xs text-slate-400">
          <Boxes class="w-6 h-6 mx-auto mb-1 text-slate-300" />
          Produk tidak ditemukan
        </div>
        <div
          v-for="p in filteredProducts"
          :key="p.id"
          @click="selectProduct(p.id)"
          class="flex items-center justify-between p-2.5 cursor-pointer hover:bg-slate-50 rounded-xl transition-colors group"
          :class="{ 'bg-primary/5 text-primary border-l-2 border-primary': modelValue === p.id }"
        >
          <div class="flex-1 min-w-0 pr-3">
            <p
              class="text-xs font-black truncate group-hover:text-primary transition-colors leading-tight"
              :class="modelValue === p.id ? 'text-primary' : 'text-slate-900'"
            >
              {{ p.name }}
            </p>
            <div class="flex items-center gap-1.5 text-[10px] text-slate-500 font-semibold mt-1 flex-wrap">
              <span class="bg-slate-100 px-1.5 py-0.2 rounded font-bold">SKU: {{ p.sku || '-' }}</span>
              <span
                v-if="p.category"
                class="px-1.5 py-0.2 rounded font-bold uppercase"
                :class="
                  p.category.toUpperCase().includes('SAMPINGAN') || p.category.toUpperCase().includes('RETUR')
                    ? 'bg-rose-50 text-rose-700 border border-rose-200'
                    : 'bg-blue-50 text-blue-700'
                "
              >
                {{ p.category }}
              </span>
              <span
                v-if="getStock || p.stock !== undefined"
                :class="
                  (getStock ? Number(getStock(p.id)) : Number(p.stock || 0)) > 0
                    ? 'text-emerald-700 bg-emerald-50 px-1.5 py-0.2 rounded font-bold'
                    : 'text-rose-600 bg-rose-50 px-1.5 py-0.2 rounded font-bold'
                "
              >
                Stok: {{ getStock ? getStock(p.id) : p.stock }} {{ p.unit || 'Pcs' }}
              </span>
            </div>
          </div>
          <div class="text-right flex items-center gap-2 shrink-0">
            <div>
              <span class="text-xs font-black text-slate-900 block whitespace-nowrap">
                {{ formatCurrency(p.selling_price || 0) }}
              </span>
              <span class="text-[9px] text-slate-400 font-bold block">per {{ p.unit || 'Pcs' }}</span>
            </div>
            <div
              v-if="modelValue === p.id"
              class="w-5 h-5 rounded-full bg-primary text-white flex items-center justify-center shrink-0"
            >
              <Check class="w-3 h-3" />
            </div>
            <div v-else class="w-5 h-5 shrink-0"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
