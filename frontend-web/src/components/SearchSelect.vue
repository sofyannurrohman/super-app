<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { Check, ChevronsUpDown, Search, Plus, X } from 'lucide-vue-next'
import { onClickOutside } from '@vueuse/core'

export interface SelectOption {
  value: string
  label: string
  sublabel?: string
  badge?: string
  badgeColor?: string
  extra?: any
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    options: SelectOption[]
    placeholder?: string
    searchPlaceholder?: string
    disabled?: boolean
    allowCustom?: boolean
    customLabelPrefix?: string
    emptyMessage?: string
    clearable?: boolean
    size?: 'sm' | 'md' | 'lg'
  }>(),
  {
    placeholder: 'Pilih...',
    searchPlaceholder: 'Cari...',
    disabled: false,
    allowCustom: false,
    customLabelPrefix: 'Gunakan: ',
    emptyMessage: 'Tidak ada pilihan ditemukan',
    clearable: false,
    size: 'md',
  },
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string, option?: SelectOption): void
  (e: 'select', option: SelectOption): void
  (e: 'custom-input', value: string): void
  (e: 'search', query: string): void
}>()

const isOpen = ref(false)
const searchQuery = ref('')
const containerRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)

watch(searchQuery, (q) => {
  emit('search', q)
})

onClickOutside(containerRef, () => {
  isOpen.value = false
})

const selectedOption = computed(() => {
  if (!props.modelValue) return null
  return props.options.find((o) => o.value === props.modelValue) || null
})

const filteredOptions = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return props.options

  return props.options.filter((opt) => {
    const valueMatch = opt.value?.toLowerCase().includes(query)
    const labelMatch = opt.label?.toLowerCase().includes(query)
    const sublabelMatch = opt.sublabel?.toLowerCase().includes(query)
    const badgeMatch = opt.badge?.toLowerCase().includes(query)
    return valueMatch || labelMatch || sublabelMatch || badgeMatch
  })
})

const canAddCustom = computed(() => {
  if (!props.allowCustom) return false
  const query = searchQuery.value.trim()
  if (!query) return false
  const exists = props.options.some((o) => o.value.toLowerCase() === query.toLowerCase() || o.label.toLowerCase() === query.toLowerCase())
  return !exists
})

function select(option: SelectOption) {
  emit('update:modelValue', option.value)
  emit('change', option.value, option)
  emit('select', option)
  isOpen.value = false
  searchQuery.value = ''
}

function selectCustom() {
  const query = searchQuery.value.trim()
  if (!query) return
  const customOpt: SelectOption = {
    value: query,
    label: query,
  }
  emit('update:modelValue', query)
  emit('change', query, customOpt)
  emit('custom-input', query)
  isOpen.value = false
  searchQuery.value = ''
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
      class="flex items-center justify-between w-full bg-slate-50 border border-slate-200/80 font-bold text-slate-700 cursor-pointer hover:border-slate-300 hover:bg-white transition-all shadow-xs"
      :class="[
        size === 'sm' ? 'px-3 py-2 text-xs rounded-xl' : size === 'lg' ? 'px-5 py-3.5 text-base rounded-2xl' : 'px-4 py-3 text-sm rounded-2xl',
        disabled ? 'opacity-50 cursor-not-allowed pointer-events-none' : '',
        isOpen ? 'ring-2 ring-primary/20 border-primary/50 bg-white' : '',
      ]"
    >
      <div class="truncate flex-1 min-w-0 pr-2">
        <template v-if="selectedOption">
          <div class="flex items-center gap-2 truncate">
            <span class="truncate text-slate-800 font-extrabold">{{ selectedOption.label }}</span>
            <span
              v-if="selectedOption.badge"
              class="px-2 py-0.5 text-[10px] rounded-full font-black uppercase tracking-wider shrink-0"
              :class="selectedOption.badgeColor || 'bg-slate-200 text-slate-700'"
            >
              {{ selectedOption.badge }}
            </span>
          </div>
        </template>
        <template v-else-if="modelValue">
          <!-- Fallback when modelValue exists but not matched in options (e.g. custom string) -->
          <span class="text-slate-800 font-extrabold truncate">{{ modelValue }}</span>
        </template>
        <template v-else>
          <span class="text-slate-400 font-medium">{{ placeholder }}</span>
        </template>
      </div>

      <div class="flex items-center gap-1.5 shrink-0">
        <button
          v-if="clearable && modelValue && !disabled"
          @click="clearSelection"
          type="button"
          class="p-1 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-200 transition-colors"
        >
          <X class="w-3.5 h-3.5" />
        </button>
        <ChevronsUpDown class="w-4 h-4 text-slate-400" />
      </div>
    </div>

    <!-- Dropdown Menu -->
    <div
      v-if="isOpen && !disabled"
      class="absolute z-50 left-0 right-0 mt-1.5 bg-white border border-slate-200 rounded-2xl shadow-xl overflow-hidden flex flex-col animate-in fade-in zoom-in-95 duration-150 max-h-72 min-w-[280px]"
    >
      <!-- Search Input Header -->
      <div class="p-2.5 border-b border-slate-100 flex items-center gap-2 bg-slate-50/50">
        <Search class="w-4 h-4 text-slate-400 shrink-0 ml-1" />
        <input
          ref="searchInputRef"
          v-model="searchQuery"
          type="text"
          :placeholder="searchPlaceholder"
          class="w-full text-xs font-semibold outline-none bg-transparent text-slate-800 placeholder:text-slate-400"
          @keydown.enter.prevent="canAddCustom ? selectCustom() : filteredOptions[0] && select(filteredOptions[0])"
          @click.stop
        />
        <button
          v-if="searchQuery"
          @click="searchQuery = ''"
          class="p-1 text-slate-400 hover:text-slate-600 rounded-md"
        >
          <X class="w-3 h-3" />
        </button>
      </div>

      <!-- Options List -->
      <div class="overflow-y-auto flex-1 p-1.5 divide-y divide-slate-50">
        <!-- Custom Input Option -->
        <div
          v-if="canAddCustom"
          @click="selectCustom"
          class="p-2.5 flex items-center justify-between text-xs font-bold text-primary bg-primary/5 hover:bg-primary/10 rounded-xl cursor-pointer transition-colors mb-1"
        >
          <div class="flex items-center gap-2 truncate">
            <Plus class="w-4 h-4 shrink-0" />
            <span class="truncate">{{ customLabelPrefix }}"{{ searchQuery.trim() }}"</span>
          </div>
          <span class="text-[10px] font-black uppercase tracking-wider px-2 py-0.5 bg-primary text-white rounded-lg shrink-0">
            Gunakan
          </span>
        </div>

        <!-- Empty State -->
        <div
          v-if="filteredOptions.length === 0 && !canAddCustom"
          class="py-6 px-4 text-center text-xs font-semibold text-slate-400"
        >
          {{ emptyMessage }}
        </div>

        <!-- Option Items -->
        <div
          v-for="opt in filteredOptions"
          :key="opt.value"
          @click="select(opt)"
          class="flex items-center justify-between p-2.5 cursor-pointer hover:bg-slate-50 rounded-xl transition-colors group"
          :class="{
            'bg-primary/5 text-primary font-black': modelValue === opt.value,
          }"
        >
          <div class="flex-1 min-w-0 pr-3">
            <div class="flex items-center gap-2">
              <span
                class="text-xs font-bold truncate group-hover:text-primary transition-colors"
                :class="modelValue === opt.value ? 'text-primary font-black' : 'text-slate-800'"
              >
                {{ opt.label }}
              </span>
              <span
                v-if="opt.badge"
                class="px-2 py-0.5 text-[9px] rounded-md font-black uppercase tracking-wider shrink-0"
                :class="opt.badgeColor || 'bg-slate-100 text-slate-600'"
              >
                {{ opt.badge }}
              </span>
            </div>
            <p v-if="opt.sublabel" class="text-[10px] font-medium text-slate-400 truncate mt-0.5">
              {{ opt.sublabel }}
            </p>
          </div>

          <Check
            v-if="modelValue === opt.value"
            class="w-4 h-4 text-primary shrink-0"
          />
        </div>
      </div>
    </div>
  </div>
</template>
