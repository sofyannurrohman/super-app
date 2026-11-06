<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

withDefaults(
  defineProps<{
    loading?: boolean
    lastUpdated?: string
    label?: string
    size?: 'sm' | 'md' | 'lg'
  }>(),
  {
    loading: false,
    lastUpdated: '',
    label: 'Segarkan Data',
    size: 'md',
  },
)

const emit = defineEmits<{
  (e: 'click'): void
}>()
</script>

<template>
  <Button
    type="button"
    variant="outline"
    :disabled="loading"
    @click="emit('click')"
    :title="
      lastUpdated
        ? `Ambil data terkini (Terakhir diperbarui: ${lastUpdated})`
        : 'Ambil data terkini dari server'
    "
    :class="[
      'bg-white hover:bg-slate-50 text-slate-700 border-slate-200 shadow-xs transition-all active:scale-95 flex items-center gap-2.5 disabled:opacity-60',
      size === 'lg'
        ? 'rounded-2xl h-14 px-5 font-black text-xs sm:text-sm'
        : size === 'sm'
          ? 'rounded-xl h-10 px-3.5 font-bold text-xs'
          : 'rounded-2xl px-4 py-3 sm:px-5 sm:py-6 font-bold text-xs sm:text-sm',
    ]"
  >
    <RefreshCw
      class="w-4 h-4 text-primary shrink-0 transition-transform"
      :class="{ 'animate-spin': loading }"
    />
    <div class="flex flex-col items-start leading-tight text-left">
      <span class="font-extrabold tracking-tight">
        {{ loading ? 'Memperbarui...' : label }}
      </span>
      <span
        v-if="lastUpdated"
        class="text-[10px] font-medium text-slate-400 hidden sm:inline-block"
      >
        Update: {{ lastUpdated }}
      </span>
    </div>
  </Button>
</template>
