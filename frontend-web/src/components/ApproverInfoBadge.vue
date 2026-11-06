<template>
  <div v-if="name || position" :class="containerClass">
    <div class="flex items-center gap-1.5 min-w-0">
      <div
        v-if="showAvatar"
        class="w-5 h-5 rounded-full bg-indigo-100 text-indigo-700 font-bold text-[9px] flex items-center justify-center shrink-0 border border-indigo-200"
      >
        {{ initials }}
      </div>
      <div class="flex flex-col min-w-0 leading-tight">
        <span class="font-semibold text-slate-800 text-[12px] truncate" :title="fullName">
          {{ fullName }}
        </span>
        <span
          v-if="position"
          class="text-[10px] font-medium text-slate-500 truncate"
          :title="position"
        >
          {{ position }}
        </span>
      </div>
    </div>
    <span
      v-if="formattedDate"
      class="text-[9px] text-slate-400 font-normal mt-0.5 whitespace-nowrap"
    >
      {{ formattedDate }}
    </span>
  </div>
  <div v-else class="text-slate-400 text-xs italic">
    —
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    name?: string
    position?: string
    date?: string | Date | null
    status?: string
    showAvatar?: boolean
    variant?: 'badge' | 'plain' | 'compact'
  }>(),
  {
    name: '',
    position: '',
    date: null,
    status: '',
    showAvatar: true,
    variant: 'plain',
  }
)

const fullName = computed(() => {
  return props.name?.trim() || 'Admin / Approver'
})

const initials = computed(() => {
  const parts = fullName.value.split(' ').filter(Boolean)
  const firstWord = parts[0] || ''
  const secondWord = parts[1] || ''
  if (firstWord && secondWord) {
    return (firstWord.charAt(0) + secondWord.charAt(0)).toUpperCase()
  }
  return (firstWord.substring(0, 2) || 'AD').toUpperCase()
})

const formattedDate = computed(() => {
  if (!props.date) return ''
  try {
    const d = new Date(props.date)
    if (isNaN(d.getTime())) return ''
    return d.toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return ''
  }
})

const containerClass = computed(() => {
  if (props.variant === 'badge') {
    return 'inline-flex flex-col bg-slate-50 border border-slate-200/80 rounded-lg px-2 py-1 max-w-full'
  }
  if (props.variant === 'compact') {
    return 'flex flex-col text-[11px] leading-tight max-w-full'
  }
  return 'flex flex-col max-w-full'
})
</script>
