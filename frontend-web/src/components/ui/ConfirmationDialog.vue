<script setup lang="ts">
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { AlertCircle, Trash2, CheckCircle2, AlertTriangle, Info, Loader2 } from 'lucide-vue-next'
import { computed } from 'vue'

const props = defineProps<{
  open: boolean
  title: string
  description: string
  variant?: 'danger' | 'warning' | 'success' | 'info'
  confirmText?: string
  cancelText?: string
  isLoading?: boolean
}>()

const emits = defineEmits(['update:open', 'confirm', 'cancel'])

const iconComponent = computed(() => {
  switch (props.variant) {
    case 'danger':
      return Trash2
    case 'warning':
      return AlertTriangle
    case 'success':
      return CheckCircle2
    case 'info':
    default:
      return Info
  }
})

const colorClasses = computed(() => {
  switch (props.variant) {
    case 'danger':
      return 'text-red-600 bg-red-50 border-red-100 shadow-red-100/50'
    case 'warning':
      return 'text-orange-600 bg-orange-50 border-orange-100 shadow-orange-100/50'
    case 'success':
      return 'text-emerald-600 bg-emerald-50 border-emerald-100 shadow-emerald-100/50'
    case 'info':
    default:
      return 'text-blue-600 bg-blue-50 border-blue-100 shadow-blue-100/50'
  }
})

const confirmBtnClass = computed(() => {
  switch (props.variant) {
    case 'danger':
      return 'bg-red-600 hover:bg-red-700 text-white shadow-red-200'
    case 'warning':
      return 'bg-orange-500 hover:bg-orange-600 text-white shadow-orange-200'
    case 'success':
      return 'bg-emerald-600 hover:bg-emerald-700 text-white shadow-emerald-200'
    case 'info':
    default:
      return 'bg-blue-600 hover:bg-blue-700 text-white shadow-blue-200'
  }
})
</script>

<template>
  <Dialog :open="open" @update:open="(val) => emits('update:open', val)">
    <DialogContent
      class="max-w-[420px] rounded-[2.5rem] border-none shadow-2xl p-0 overflow-hidden animate-in zoom-in-95 duration-300"
    >
      <div class="p-10 space-y-8">
        <div class="flex flex-col items-center text-center space-y-6">
          <div
            :class="[
              'p-6 rounded-full border-4 transition-all duration-500 scale-110 shadow-xl',
              colorClasses,
            ]"
          >
            <component :is="iconComponent" class="w-10 h-10" />
          </div>
          <div class="space-y-3">
            <DialogTitle class="text-2xl font-black text-slate-900 tracking-tight leading-tight">{{
              title
            }}</DialogTitle>
            <DialogDescription class="text-sm font-bold text-slate-400 leading-relaxed px-4">
              {{ description }}
            </DialogDescription>
          </div>
        </div>
      </div>

      <DialogFooter
        class="bg-slate-50/80 backdrop-blur-sm p-8 flex gap-4 border-t border-slate-100/50"
      >
        <Button
          variant="ghost"
          @click="emits('cancel')"
          class="flex-1 rounded-2xl font-bold h-14 text-slate-500 hover:bg-slate-100 transition-all"
          :disabled="isLoading"
        >
          {{ cancelText || 'Batal' }}
        </Button>
        <Button
          @click="emits('confirm')"
          :class="[
            'flex-1 rounded-2xl font-black h-14 shadow-2xl transition-all hover:scale-[1.02] active:scale-[0.98]',
            confirmBtnClass,
          ]"
          :disabled="isLoading"
        >
          <Loader2 v-if="isLoading" class="w-5 h-5 mr-2 animate-spin" />
          {{ confirmText || 'Ya, Lanjutkan' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
