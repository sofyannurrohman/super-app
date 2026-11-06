<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue'
import { Check, ChevronsUpDown, Search, Wrench, MapPin, X, AlertTriangle } from 'lucide-vue-next'
import { onClickOutside } from '@vueuse/core'

export interface TechnicianItem {
  id: string
  sku: string
  name: string
  category: string
  specification?: string
  unit: string
  current_stock: number
  minimum_stock: number
  rack_location?: string
  unit_price?: number
  [key: string]: any
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    items: TechnicianItem[]
    disabled?: boolean
    placeholder?: string
    searchPlaceholder?: string
    excludeItemIds?: string[]
  }>(),
  {
    disabled: false,
    placeholder: 'Pilih barang / sparepart...',
    searchPlaceholder: 'Cari SKU, nama barang, rak, atau spesifikasi...',
    excludeItemIds: () => [],
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string, item?: TechnicianItem): void
  (e: 'select', item: TechnicianItem): void
}>()

const isOpen = ref(false)
const searchQuery = ref('')
const containerRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)

onClickOutside(containerRef, () => {
  isOpen.value = false
})

const CATEGORY_MAP: Record<string, { label: string; badgeClass: string }> = {
  MUR_BAUT: { label: '🔩 Mur & Baut', badgeClass: 'bg-amber-100 text-amber-800' },
  SPAREPART_MESIN: { label: '⚙️ Sparepart Mesin', badgeClass: 'bg-blue-100 text-blue-800' },
  KELISTRIKAN: { label: '⚡ Kelistrikan & Sensor', badgeClass: 'bg-yellow-100 text-yellow-800' },
  PELUMAS: { label: '🛢️ Pelumas & Oli', badgeClass: 'bg-purple-100 text-purple-800' },
  TOOLS: { label: '🔧 Tools & Perkakas', badgeClass: 'bg-emerald-100 text-emerald-800' },
  PNEUMATIK_HIDROLIK: { label: '💨 Pneumatik & Hidrolik', badgeClass: 'bg-cyan-100 text-cyan-800' },
  LAINNYA: { label: '📦 Lainnya', badgeClass: 'bg-slate-100 text-slate-800' },
}

const getCategoryInfo = (category: string) => {
  return CATEGORY_MAP[category] || { label: category || 'Umum', badgeClass: 'bg-slate-100 text-slate-700' }
}

const selectedItem = computed(() => {
  if (!props.modelValue) return null
  return props.items.find((it) => it.id === props.modelValue) || null
})

const filteredItems = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  let list = props.items || []

  if (!query) return list

  return list.filter((it) => {
    const skuMatch = it.sku?.toLowerCase().includes(query)
    const nameMatch = it.name?.toLowerCase().includes(query)
    const specMatch = it.specification?.toLowerCase().includes(query)
    const rackMatch = it.rack_location?.toLowerCase().includes(query)
    const catLabel = getCategoryInfo(it.category).label.toLowerCase()
    const catMatch = catLabel.includes(query) || it.category?.toLowerCase().includes(query)

    return skuMatch || nameMatch || specMatch || rackMatch || catMatch
  })
})

function selectItem(item: TechnicianItem) {
  emit('update:modelValue', item.id)
  emit('change', item.id, item)
  emit('select', item)
  isOpen.value = false
  searchQuery.value = ''
}

function selectFirstFiltered() {
  if (filteredItems.value.length > 0 && filteredItems.value[0]) {
    selectItem(filteredItems.value[0])
  }
}

function clearSelection(e: MouseEvent) {
  e.stopPropagation()
  emit('update:modelValue', '')
  emit('change', '', undefined)
}

function toggleDropdown() {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    searchQuery.value = ''
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  }
}

watch(isOpen, (val) => {
  if (val) {
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  }
})
</script>

<template>
  <div class="relative w-full" ref="containerRef">
    <!-- Trigger Button -->
    <div
      @click="toggleDropdown"
      tabindex="0"
      @keydown.down.prevent="isOpen = true"
      @keydown.enter.prevent="toggleDropdown"
      class="group flex items-center justify-between w-full min-h-[42px] bg-white border border-slate-200 rounded-xl px-3 py-1.5 text-xs cursor-pointer hover:border-orange-300 hover:bg-orange-50/20 transition-all shadow-xs select-none"
      :class="{
        'opacity-50 cursor-not-allowed pointer-events-none': disabled,
        'border-orange-500 ring-2 ring-orange-500/20 bg-white shadow-sm': isOpen,
      }"
    >
      <div class="truncate flex-1 min-w-0 pr-2">
        <template v-if="selectedItem">
          <div class="flex items-center gap-2 flex-wrap sm:flex-nowrap">
            <!-- SKU Badge -->
            <span class="px-2 py-0.5 rounded-lg bg-slate-100 text-slate-800 border border-slate-200 font-mono font-black text-[11px] shrink-0">
              {{ selectedItem.sku }}
            </span>

            <!-- Item Name -->
            <span class="font-black text-slate-900 truncate text-xs">
              {{ selectedItem.name }}
            </span>

            <!-- Category Pill -->
            <span
              class="px-2 py-0.5 rounded-full text-[10px] font-black shrink-0 hidden md:inline-flex"
              :class="getCategoryInfo(selectedItem.category).badgeClass"
            >
              {{ getCategoryInfo(selectedItem.category).label }}
            </span>

            <!-- Stock Pill -->
            <span
              class="px-2 py-0.5 rounded-lg font-mono font-black text-[10px] shrink-0"
              :class="
                selectedItem.current_stock <= 0
                  ? 'bg-red-100 text-red-800 border border-red-200'
                  : selectedItem.current_stock <= selectedItem.minimum_stock
                  ? 'bg-amber-100 text-amber-800 border border-amber-200'
                  : 'bg-emerald-50 text-emerald-800 border border-emerald-200'
              "
            >
              Stok: {{ selectedItem.current_stock }} {{ selectedItem.unit }}
            </span>
          </div>
        </template>
        <template v-else>
          <span class="text-slate-400 font-medium flex items-center gap-1.5">
            <Wrench class="w-3.5 h-3.5 text-slate-400 group-hover:text-orange-500 transition-colors" />
            {{ placeholder }}
          </span>
        </template>
      </div>

      <div class="flex items-center gap-1 shrink-0 text-slate-400">
        <button
          v-if="modelValue && !disabled"
          @click="clearSelection"
          type="button"
          class="p-1 hover:text-slate-700 rounded-lg hover:bg-slate-100 transition-colors mr-0.5"
          title="Hapus Pilihan"
        >
          <X class="w-3.5 h-3.5" />
        </button>
        <ChevronsUpDown class="w-4 h-4 text-slate-400 group-hover:text-slate-600 transition-colors" />
      </div>
    </div>

    <!-- Dropdown Menu -->
    <div
      v-if="isOpen && !disabled"
      class="absolute z-50 left-0 mt-1.5 w-full min-w-[320px] sm:min-w-[480px] max-w-xl bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden flex flex-col animate-in fade-in zoom-in-95 duration-150"
    >
      <!-- Search Input Header -->
      <div class="p-2.5 border-b border-slate-100 bg-slate-50/90 space-y-1.5">
        <div class="relative">
          <Search class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
          <input
            ref="searchInputRef"
            v-model="searchQuery"
            type="text"
            :placeholder="searchPlaceholder"
            class="w-full pl-9 pr-8 py-2 text-xs font-semibold rounded-xl bg-white border border-slate-200 outline-none focus:border-orange-500 focus:ring-2 focus:ring-orange-500/20 text-slate-800 placeholder:text-slate-400 shadow-2xs"
            @click.stop
            @keydown.esc.stop="isOpen = false"
            @keydown.enter.prevent="selectFirstFiltered"
          />
          <button
            v-if="searchQuery"
            @click="searchQuery = ''"
            type="button"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>
        <div class="flex items-center justify-between text-[10px] text-slate-500 px-1 font-medium">
          <span>Ditemukan: <strong class="text-slate-800">{{ filteredItems.length }}</strong> barang</span>
          <span class="text-slate-400">Tekan Enter untuk memilih item pertama</span>
        </div>
      </div>

      <!-- Items List -->
      <div class="overflow-y-auto max-h-72 p-1.5 divide-y divide-slate-100/80">
        <div v-if="filteredItems.length === 0" class="py-8 px-4 text-center text-xs text-slate-400 space-y-1">
          <Wrench class="w-8 h-8 mx-auto text-slate-300 stroke-1" />
          <p class="font-bold text-slate-600">Barang tidak ditemukan</p>
          <p class="text-[11px]">Coba cari dengan kata kunci SKU, nama part, atau lokasi rak lain.</p>
        </div>

        <div
          v-for="item in filteredItems"
          :key="item.id"
          @click="selectItem(item)"
          class="p-2.5 flex items-start justify-between cursor-pointer hover:bg-orange-50/50 rounded-xl transition-all group"
          :class="{
            'bg-orange-50/80 border-l-4 border-orange-500 pl-3': modelValue === item.id,
            'opacity-60': item.current_stock <= 0 && modelValue !== item.id,
          }"
        >
          <!-- Left info: SKU, Name, Category, Spec, Rack -->
          <div class="flex-1 min-w-0 pr-3 space-y-1">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="px-1.5 py-0.5 rounded bg-slate-100 text-slate-800 font-mono font-bold text-[10px] border border-slate-200">
                {{ item.sku }}
              </span>
              <span
                class="text-xs font-black group-hover:text-orange-700 transition-colors leading-tight"
                :class="modelValue === item.id ? 'text-orange-950' : 'text-slate-900'"
              >
                {{ item.name }}
              </span>
              <span
                class="px-2 py-0.2 rounded-full text-[9px] font-extrabold"
                :class="getCategoryInfo(item.category).badgeClass"
              >
                {{ getCategoryInfo(item.category).label }}
              </span>
            </div>

            <!-- Spec & Location -->
            <div class="flex items-center gap-3 text-[11px] text-slate-500 flex-wrap">
              <span v-if="item.specification" class="line-clamp-1 italic text-slate-600">
                {{ item.specification }}
              </span>
              <span v-if="item.rack_location" class="flex items-center gap-1 font-mono text-[10px] text-slate-600 font-medium">
                <MapPin class="w-3 h-3 text-slate-400" /> {{ item.rack_location }}
              </span>
            </div>
          </div>

          <!-- Right info: Stock & Check -->
          <div class="flex items-center gap-2.5 shrink-0 pt-0.5">
            <div class="text-right">
              <div
                class="inline-flex items-center gap-1 px-2 py-0.5 rounded-lg font-mono font-black text-xs"
                :class="
                  item.current_stock <= 0
                    ? 'bg-red-50 text-red-700 border border-red-200'
                    : item.current_stock <= item.minimum_stock
                    ? 'bg-amber-50 text-amber-700 border border-amber-200'
                    : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                "
              >
                <span>{{ item.current_stock }}</span>
                <span class="text-[9px] uppercase">{{ item.unit }}</span>
              </div>
              <div v-if="item.current_stock <= 0" class="text-[9px] font-bold text-red-600 flex items-center justify-end gap-0.5 mt-0.5">
                <AlertTriangle class="w-2.5 h-2.5" /> Stok Habis
              </div>
              <div v-else-if="item.current_stock <= item.minimum_stock" class="text-[9px] font-bold text-amber-600 mt-0.5">
                Stok Menipis
              </div>
            </div>

            <div
              v-if="modelValue === item.id"
              class="w-5 h-5 rounded-full bg-orange-600 text-white flex items-center justify-center shadow-xs"
            >
              <Check class="w-3.5 h-3.5 stroke-[3]" />
            </div>
            <div v-else class="w-5 h-5"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
