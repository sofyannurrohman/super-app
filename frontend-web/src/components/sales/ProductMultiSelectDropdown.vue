<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  Package,
  ChevronDown,
  Check,
  CheckCheck,
  RotateCcw,
  Search,
  Sparkles,
  Layers,
  X,
} from 'lucide-vue-next'
import {
  buildProductFamilies,
  type ProductFamilyGroup,
  type ProductVariantItem,
} from '@/utils/productFamily'

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    products: any[]
    placeholder?: string
    allowSelectAll?: boolean
    compact?: boolean
  }>(),
  {
    modelValue: () => [],
    products: () => [],
    placeholder: 'Pilih Produk / Induk Produk...',
    allowSelectAll: true,
    compact: false,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
  (e: 'change', value: string[]): void
}>()

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const searchQuery = ref('')

const families = computed(() => {
  return buildProductFamilies(props.products)
})

const filteredFamilies = computed(() => {
  if (!searchQuery.value.trim()) return families.value
  const q = searchQuery.value.toLowerCase().trim()

  return families.value
    .map((fam) => {
      const matchFam = fam.name.toLowerCase().includes(q) || fam.category.toLowerCase().includes(q)
      const matchingVariants = fam.variants.filter(
        (v) =>
          v.name.toLowerCase().includes(q) ||
          (v.sku && v.sku.toLowerCase().includes(q)) ||
          (v.unit && v.unit.toLowerCase().includes(q))
      )

      if (matchFam) {
        return fam
      } else if (matchingVariants.length > 0) {
        return {
          ...fam,
          variants: matchingVariants,
        }
      }
      return null
    })
    .filter((f): f is ProductFamilyGroup => f !== null)
})

const selectedIdsSet = computed(() => new Set(props.modelValue))

const isFamilyFullySelected = (fam: ProductFamilyGroup) => {
  if (!fam.productIds.length) return false
  return fam.productIds.every((id) => selectedIdsSet.value.has(id))
}

const isFamilyPartiallySelected = (fam: ProductFamilyGroup) => {
  const count = fam.productIds.filter((id) => selectedIdsSet.value.has(id)).length
  return count > 0 && count < fam.productIds.length
}

const getSelectedCountInFamily = (fam: ProductFamilyGroup) => {
  return fam.productIds.filter((id) => selectedIdsSet.value.has(id)).length
}

const toggleVariant = (variantId: string) => {
  const next = new Set(props.modelValue)
  if (next.has(variantId)) {
    next.delete(variantId)
  } else {
    next.add(variantId)
  }
  const result = Array.from(next)
  emit('update:modelValue', result)
  emit('change', result)
}

const toggleFamily = (fam: ProductFamilyGroup) => {
  const next = new Set(props.modelValue)
  const isAll = isFamilyFullySelected(fam)

  if (isAll) {
    // Unselect all variants in this family
    fam.productIds.forEach((id) => next.delete(id))
  } else {
    // Select all variants in this family
    fam.productIds.forEach((id) => next.add(id))
  }
  const result = Array.from(next)
  emit('update:modelValue', result)
  emit('change', result)
}

const selectAllProducts = () => {
  const allIds = props.products.map((p) => p.id).filter(Boolean)
  emit('update:modelValue', allIds)
  emit('change', allIds)
}

const clearSelection = () => {
  emit('update:modelValue', [])
  emit('change', [])
}

const displayButtonLabel = computed(() => {
  if (props.modelValue.length === 0) {
    return props.placeholder
  }
  if (props.modelValue.length === props.products.length && props.products.length > 0) {
    return `Semua Produk (${props.products.length} Varian)`
  }

  // Check if exactly one or more whole families are selected
  const fullySelectedFams = families.value.filter((fam) => isFamilyFullySelected(fam))
  const totalInSelectedFams = fullySelectedFams.reduce((sum, f) => sum + f.productIds.length, 0)

  if (fullySelectedFams.length === 1 && totalInSelectedFams === props.modelValue.length) {
    const firstFam = fullySelectedFams[0]
    if (firstFam) {
      return `${firstFam.name} (Semua ${firstFam.productIds.length} Ukuran)`
    }
  } else if (fullySelectedFams.length > 1 && totalInSelectedFams === props.modelValue.length) {
    return `${fullySelectedFams.length} Induk Produk (${props.modelValue.length} Varian)`
  }

  if (props.modelValue.length === 1) {
    const p = props.products.find((item) => item.id === props.modelValue[0])
    return p ? p.name : '1 Produk Terpilih'
  }

  return `${props.modelValue.length} Varian Terpilih`
})

// Close on outside click
const handleClickOutside = (e: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('mousedown', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', handleClickOutside)
})
</script>

<template>
  <div ref="dropdownRef" class="relative inline-block text-left select-none">
    <!-- Trigger Button -->
    <button
      type="button"
      @click="isOpen = !isOpen"
      :class="[
        'flex items-center gap-2 bg-white border border-slate-200 rounded-xl transition-all shadow-2xs hover:border-slate-300 focus:outline-none focus:ring-2 focus:ring-primary/20',
        compact ? 'px-2.5 py-1.5 text-xs' : 'px-3.5 py-2 text-xs',
        isOpen ? 'ring-2 ring-primary/20 border-primary' : '',
      ]"
    >
      <div class="p-1 rounded-lg bg-emerald-50 text-emerald-600 shrink-0">
        <Package class="w-3.5 h-3.5" />
      </div>
      <span
        :class="[
          'font-extrabold truncate text-left max-w-[220px] sm:max-w-[280px]',
          modelValue.length > 0 ? 'text-slate-800' : 'text-slate-500',
        ]"
      >
        {{ displayButtonLabel }}
      </span>

      <!-- Count Badge -->
      <span
        v-if="modelValue.length > 0"
        class="ml-auto bg-emerald-600 text-white font-black text-[10px] px-1.5 py-0.5 rounded-full shrink-0 leading-none shadow-2xs"
      >
        {{ modelValue.length }}
      </span>

      <ChevronDown
        :class="[
          'w-4 h-4 text-slate-400 transition-transform duration-200 shrink-0',
          isOpen ? 'rotate-180 text-primary' : '',
        ]"
      />
    </button>

    <!-- Dropdown Panel -->
    <div
      v-if="isOpen"
      class="absolute right-0 sm:left-0 mt-2 w-[340px] sm:w-[420px] bg-white rounded-2xl border border-slate-200 shadow-2xl z-[9999] overflow-hidden flex flex-col font-sans animate-in fade-in zoom-in-95 duration-150"
    >
      <!-- Header / Search -->
      <div class="p-3 border-b border-slate-100 bg-slate-50/80 space-y-2.5">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-1.5 text-slate-800">
            <Sparkles class="w-3.5 h-3.5 text-amber-500" />
            <span class="text-xs font-black tracking-tight">Pilih Induk & Varian Produk</span>
          </div>
          <span class="text-[10px] font-bold text-slate-500">
            {{ modelValue.length }} dari {{ products.length }} varian dipilih
          </span>
        </div>

        <div class="relative">
          <Search class="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari induk produk, ukuran, SKU..."
            class="w-full bg-white pl-8 pr-7 py-1.5 text-xs font-bold text-slate-800 placeholder-slate-400 rounded-xl border border-slate-200 outline-none focus:ring-2 focus:ring-primary/20"
          />
          <button
            v-if="searchQuery"
            @click="searchQuery = ''"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>

        <!-- Action Buttons -->
        <div class="flex items-center justify-between gap-2 pt-1">
          <button
            type="button"
            @click="selectAllProducts"
            class="text-[11px] font-extrabold text-primary hover:text-primary-700 flex items-center gap-1 hover:underline cursor-pointer"
          >
            <CheckCheck class="w-3.5 h-3.5" />
            <span>Pilih Semua Produk</span>
          </button>
          <button
            type="button"
            @click="clearSelection"
            class="text-[11px] font-extrabold text-rose-500 hover:text-rose-700 flex items-center gap-1 hover:underline cursor-pointer"
          >
            <RotateCcw class="w-3 h-3" />
            <span>Reset Pilihan</span>
          </button>
        </div>
      </div>

      <!-- Families & Variants List -->
      <div class="max-h-[360px] overflow-y-auto divide-y divide-slate-100 p-2 space-y-2">
        <div
          v-for="fam in filteredFamilies"
          :key="fam.id"
          class="rounded-xl border border-slate-100 bg-white hover:border-slate-200 transition-all p-2"
        >
          <!-- Induk Header Banner -->
          <div
            class="flex items-center justify-between p-2 rounded-lg bg-slate-50 hover:bg-slate-100/80 transition-colors cursor-pointer"
            @click="toggleFamily(fam)"
          >
            <div class="flex items-center gap-2.5 min-w-0">
              <!-- Checkbox Family -->
              <div
                :class="[
                  'w-4.5 h-4.5 rounded-md border flex items-center justify-center transition-all shrink-0',
                  isFamilyFullySelected(fam)
                    ? 'bg-emerald-600 border-emerald-600 text-white'
                    : isFamilyPartiallySelected(fam)
                    ? 'bg-emerald-50 border-emerald-500 text-emerald-600'
                    : 'bg-white border-slate-300 text-transparent',
                ]"
              >
                <Check v-if="isFamilyFullySelected(fam)" class="w-3 h-3 stroke-[3]" />
                <span v-else-if="isFamilyPartiallySelected(fam)" class="w-2 h-0.5 bg-emerald-600 rounded-full"></span>
              </div>

              <div class="min-w-0">
                <div class="flex items-center gap-1.5">
                  <h4 class="text-xs font-black text-slate-900 truncate">
                    {{ fam.name }}
                  </h4>
                  <span class="text-[9px] font-extrabold uppercase px-1.5 py-0.5 rounded-md bg-white border border-slate-200 text-slate-500">
                    {{ fam.category }}
                  </span>
                </div>
                <p class="text-[10px] font-bold text-slate-400 mt-0.5">
                  {{ fam.variants.length }} Varian Ukuran Kemasan
                </p>
              </div>
            </div>

            <!-- Select All in Family Badge Button -->
            <button
              type="button"
              class="shrink-0 text-[10px] font-black px-2 py-1 rounded-lg border transition-all"
              :class="[
                isFamilyFullySelected(fam)
                  ? 'bg-emerald-600 text-white border-emerald-600 shadow-2xs'
                  : 'bg-white text-slate-600 border-slate-200 hover:border-emerald-500 hover:text-emerald-700',
              ]"
            >
              {{ isFamilyFullySelected(fam) ? 'Semua Aktif' : 'Pilih Semua' }}
            </button>
          </div>

          <!-- Variants Children Sub-List -->
          <div class="mt-1.5 pl-6 pr-1 space-y-1">
            <label
              v-for="v in fam.variants"
              :key="v.id"
              class="flex items-center justify-between p-1.5 rounded-lg hover:bg-emerald-50/60 transition-all cursor-pointer group"
              @click.stop="toggleVariant(v.id)"
            >
              <div class="flex items-center gap-2 min-w-0">
                <div
                  :class="[
                    'w-3.5 h-3.5 rounded border flex items-center justify-center transition-all shrink-0',
                    selectedIdsSet.has(v.id)
                      ? 'bg-emerald-600 border-emerald-600 text-white'
                      : 'bg-white border-slate-300 group-hover:border-slate-400 text-transparent',
                  ]"
                >
                  <Check v-if="selectedIdsSet.has(v.id)" class="w-2.5 h-2.5 stroke-[3]" />
                </div>
                <span
                  :class="[
                    'text-xs truncate transition-colors',
                    selectedIdsSet.has(v.id) ? 'font-black text-slate-900' : 'font-medium text-slate-600',
                  ]"
                >
                  {{ v.name }}
                </span>
              </div>

              <span
                v-if="v.unit"
                class="text-[9px] font-bold text-slate-400 uppercase shrink-0 px-1 py-0.5 rounded bg-slate-50"
              >
                {{ v.unit }}
              </span>
            </label>
          </div>
        </div>

        <div v-if="filteredFamilies.length === 0" class="py-8 text-center text-slate-400">
          <Package class="w-8 h-8 mx-auto mb-2 opacity-40" />
          <p class="text-xs font-bold">Tidak ada produk yang cocok dengan pencarian.</p>
        </div>
      </div>

      <!-- Footer Done -->
      <div class="p-2.5 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
        <span class="text-[11px] font-bold text-slate-500">
          {{ modelValue.length }} item dipilih
        </span>
        <button
          type="button"
          @click="isOpen = false"
          class="px-4 py-1.5 bg-primary hover:bg-primary-700 text-white text-xs font-black rounded-xl shadow-sm transition-all cursor-pointer"
        >
          Terapkan
        </button>
      </div>
    </div>
  </div>
</template>
