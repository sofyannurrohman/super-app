<script setup lang="ts">
import { computed } from 'vue'
import {
  HoverCardRoot,
  HoverCardTrigger,
  HoverCardPortal,
  HoverCardContent,
  HoverCardArrow,
} from 'reka-ui'
import {
  Package,
  AlertTriangle,
  CheckCircle2,
  Scale,
  Layers,
  Calendar,
  AlertCircle,
} from 'lucide-vue-next'

export interface ProductBatch {
  id?: string
  batch_no: string
  available_quantity: number
  quantity?: number
  expiry_date?: string
}

const props = withDefaults(
  defineProps<{
    product?: any
    batches?: ProductBatch[]
    requiredQuantity?: number // in PCS
    stockLocation?: string
    side?: 'top' | 'right' | 'bottom' | 'left'
    align?: 'start' | 'center' | 'end'
    openDelay?: number
    closeDelay?: number
    highlightDeficient?: boolean
  }>(),
  {
    side: 'top',
    align: 'center',
    openDelay: 120,
    closeDelay: 150,
    stockLocation: 'Pabrik',
    highlightDeficient: false,
  },
)

const totalAvailableStock = computed(() => {
  if (!props.batches || props.batches.length === 0) return 0
  return props.batches.reduce((sum, b) => sum + (Number(b.available_quantity) || 0), 0)
})

const stockStatus = computed<'EMPTY' | 'DEFICIENT' | 'READY' | 'UNKNOWN'>(() => {
  if (!props.product) return 'UNKNOWN'
  if (props.batches !== undefined) {
    if (totalAvailableStock.value <= 0) return 'EMPTY'
    if (props.requiredQuantity && totalAvailableStock.value < props.requiredQuantity) {
      return 'DEFICIENT'
    }
    return 'READY'
  }
  return 'UNKNOWN'
})
</script>

<template>
  <HoverCardRoot :open-delay="openDelay" :close-delay="closeDelay">
    <HoverCardTrigger as-child>
      <slot>
        <span
          class="cursor-pointer font-bold text-slate-700 hover:text-primary transition-colors underline decoration-dotted underline-offset-4"
          :class="{
            'text-rose-600 decoration-rose-300 hover:text-rose-700 font-black':
              highlightDeficient && stockStatus === 'EMPTY',
            'text-amber-600 decoration-amber-300 hover:text-amber-700 font-black':
              highlightDeficient && stockStatus === 'DEFICIENT',
          }"
        >
          {{ product?.name || 'Detail Produk' }}
        </span>
      </slot>
    </HoverCardTrigger>

    <HoverCardPortal>
      <HoverCardContent
        :side="side"
        :align="align"
        :side-offset="8"
        class="z-[99999] w-80 max-w-[90vw] rounded-2xl bg-white p-4 shadow-2xl border border-slate-200/80 outline-none animate-in fade-in zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 duration-150 text-slate-800"
      >
        <div v-if="product" class="space-y-3">
          <!-- Header: SKU & Category -->
          <div class="flex items-center justify-between gap-2 flex-wrap text-[10px]">
            <span
              v-if="product.sku"
              class="font-mono font-bold bg-slate-100 text-slate-700 px-2 py-0.5 rounded-md border border-slate-200 tracking-tight"
            >
              SKU: {{ product.sku }}
            </span>
            <div class="flex items-center gap-1.5 ml-auto flex-wrap">
              <span
                v-if="product.category"
                class="font-bold bg-blue-50 text-blue-700 px-2 py-0.5 rounded-md border border-blue-100/60"
              >
                {{ product.category }}
              </span>
              <span
                v-if="product.brand"
                class="font-bold bg-purple-50 text-purple-700 px-2 py-0.5 rounded-md border border-purple-100/60"
              >
                {{ product.brand }}
              </span>
            </div>
          </div>

          <!-- Product Name -->
          <div>
            <h4 class="text-xs font-black text-slate-900 leading-snug tracking-tight">
              {{ product.name }}
            </h4>
            <p v-if="product.description" class="text-[10px] text-slate-500 mt-1 line-clamp-2">
              {{ product.description }}
            </p>
          </div>

          <!-- Specs & Kemasan -->
          <div class="grid grid-cols-2 gap-2 text-[10px] bg-slate-50/80 p-2.5 rounded-xl border border-slate-100">
            <div class="flex items-center gap-1.5 text-slate-600">
              <Package class="w-3.5 h-3.5 text-primary shrink-0" />
              <span>
                1 {{ product.unit || 'Unit' }} =
                <strong class="font-bold text-slate-900">{{ product.pcs_per_unit || 1 }}</strong> PCS
              </span>
            </div>
            <div class="flex items-center gap-1.5 text-slate-600 justify-end">
              <Scale class="w-3.5 h-3.5 text-amber-500 shrink-0" />
              <span>
                <strong class="font-bold text-slate-900">{{ product.weight || 0 }}</strong>
                {{ product.weight_unit || 'GR' }}
              </span>
            </div>
          </div>

          <!-- Stock Availability Section (If batches provided) -->
          <div
            v-if="batches !== undefined"
            class="p-2.5 rounded-xl border text-[11px] space-y-2 transition-colors"
            :class="{
              'bg-rose-50/70 border-rose-200 text-rose-900': stockStatus === 'EMPTY',
              'bg-amber-50/70 border-amber-200 text-amber-900': stockStatus === 'DEFICIENT',
              'bg-emerald-50/70 border-emerald-200 text-emerald-900': stockStatus === 'READY',
              'bg-slate-50 border-slate-200 text-slate-700': stockStatus === 'UNKNOWN',
            }"
          >
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-1.5 font-black text-[10px] uppercase tracking-wider">
                <AlertTriangle v-if="stockStatus === 'EMPTY' || stockStatus === 'DEFICIENT'" class="w-3.5 h-3.5 shrink-0" />
                <CheckCircle2 v-else-if="stockStatus === 'READY'" class="w-3.5 h-3.5 text-emerald-600 shrink-0" />
                <Layers v-else class="w-3.5 h-3.5 text-slate-500 shrink-0" />
                <span>Stok {{ stockLocation }}</span>
              </div>
              <span
                class="text-[10px] font-black px-2 py-0.5 rounded-md"
                :class="{
                  'bg-rose-100 text-rose-700': stockStatus === 'EMPTY',
                  'bg-amber-100 text-amber-700': stockStatus === 'DEFICIENT',
                  'bg-emerald-100 text-emerald-700': stockStatus === 'READY',
                  'bg-slate-200 text-slate-700': stockStatus === 'UNKNOWN',
                }"
              >
                <template v-if="stockStatus === 'EMPTY'">HABIS (0 PCS)</template>
                <template v-else-if="stockStatus === 'DEFICIENT'">
                  KURANG ({{ totalAvailableStock }} / {{ requiredQuantity }} PCS)
                </template>
                <template v-else-if="stockStatus === 'READY'">
                  TERSEDIA ({{ totalAvailableStock }} PCS)
                </template>
                <template v-else>{{ totalAvailableStock }} PCS</template>
              </span>
            </div>

            <!-- Batch Details Breakdown -->
            <div v-if="batches.length > 0" class="pt-1.5 border-t border-current/10 space-y-1">
              <p class="text-[9px] font-bold text-slate-500 uppercase tracking-widest">
                Daftar Batch Siap Kirim:
              </p>
              <div class="space-y-1 max-h-28 overflow-y-auto pr-1">
                <div
                  v-for="(b, idx) in batches.slice(0, 5)"
                  :key="b.id || idx"
                  class="flex items-center justify-between text-[10px] bg-white/70 px-2 py-1 rounded-md border border-black/5"
                >
                  <span class="font-mono font-bold text-slate-700">
                    Batch {{ b.batch_no }}
                  </span>
                  <div class="flex items-center gap-2">
                    <span
                      v-if="b.expiry_date"
                      class="text-[9px] text-slate-400 flex items-center gap-0.5"
                    >
                      <Calendar class="w-2.5 h-2.5" />
                      {{
                        new Date(b.expiry_date).toLocaleDateString('id-ID', {
                          month: 'short',
                          year: '2-digit',
                        })
                      }}
                    </span>
                    <span class="font-black text-slate-900">
                      {{ b.available_quantity }} PCS
                    </span>
                  </div>
                </div>
                <p
                  v-if="batches.length > 5"
                  class="text-[9px] font-bold text-center text-slate-400 pt-0.5 italic"
                >
                  + {{ batches.length - 5 }} batch lainnya
                </p>
              </div>
            </div>
            <div v-else class="text-[10px] font-semibold text-rose-600 italic">
              Tidak ada batch dengan stok tersedia di {{ stockLocation.toLowerCase() }}.
            </div>
          </div>
        </div>

        <div v-else class="text-center py-2 text-xs text-slate-400 flex items-center justify-center gap-1.5">
          <AlertCircle class="w-4 h-4 text-slate-400" />
          <span>Informasi produk tidak ditemukan</span>
        </div>

        <HoverCardArrow class="fill-white drop-shadow-sm" />
      </HoverCardContent>
    </HoverCardPortal>
  </HoverCardRoot>
</template>
