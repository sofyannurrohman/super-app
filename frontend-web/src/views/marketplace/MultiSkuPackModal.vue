<template>
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4"
    style="background: rgba(0,0,0,0.65); backdrop-filter: blur(4px)"
    @click.self="$emit('cancel')"
  >
    <!-- Modal Card -->
    <div class="bg-slate-900 border border-slate-700 rounded-2xl shadow-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto animate-in fade-in zoom-in-95 duration-200">
      
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-5 border-b border-slate-700">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-indigo-500/20 rounded-xl">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="text-indigo-400">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
              <polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/>
            </svg>
          </div>
          <div>
            <h2 class="text-white font-bold text-base">Multi-Item Packing</h2>
            <div class="flex items-center gap-2 mt-0.5">
              <p class="text-slate-400 text-xs font-mono">{{ order.order_id }}</p>
              <span v-if="order.warehouse_name" class="inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded bg-amber-500/20 text-amber-300 font-semibold border border-amber-500/30">
                <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 8.35V20a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V8.35A2 2 0 0 1 3.26 6.5l8-3.2a2 2 0 0 1 1.48 0l8 3.2A2 2 0 0 1 22 8.35Z"/><path d="M6 18h12"/><path d="M6 14h12"/><rect width="12" height="12" x="6" y="10"/></svg>
                {{ order.warehouse_name }}
              </span>
            </div>
          </div>
        </div>
        <button @click="$emit('cancel')" class="text-slate-500 hover:text-slate-300 transition-colors p-1">
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
          </svg>
        </button>
      </div>

      <div class="p-6 space-y-5">

        <!-- SECTION 1: Error Block (SKU unmapped) -->
        <div v-if="unmappedItems.length > 0" class="bg-red-950/60 border border-red-700/60 rounded-xl p-4">
          <div class="flex items-center gap-2 mb-3">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="text-red-400 shrink-0">
              <circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/>
            </svg>
            <p class="text-red-300 font-bold text-sm">Ada SKU yang belum memiliki mapping stok!</p>
          </div>
          <ul class="space-y-1 mb-4">
            <li v-for="item in unmappedItems" :key="item.sku_id" class="text-red-400 text-sm flex items-start gap-2">
              <span class="text-red-600 shrink-0">•</span>
              <span><span class="font-mono font-bold">{{ item.seller_sku }}</span> — {{ item.product_name }}</span>
            </li>
          </ul>
          <p class="text-red-500 text-xs mb-3">Setup mapping terlebih dahulu di menu <span class="font-semibold">Mapping Produk</span> sebelum melakukan packing.</p>
          <a
            href="/marketplace/mapping"
            class="inline-flex items-center gap-1.5 text-xs bg-red-800/60 hover:bg-red-700/60 text-red-300 border border-red-700/50 px-3 py-1.5 rounded-lg transition-colors font-semibold"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M5 12h14"/><path d="m12 5 7 7-7 7"/>
            </svg>
            Ke Halaman Mapping Produk
          </a>
        </div>

        <!-- SECTION 2: Item List -->
        <div>
          <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-3">
            ① Detail Item
            <span class="ml-2 bg-indigo-500/20 text-indigo-300 rounded-full px-2 py-0.5 text-xs font-bold">
              {{ order.items.length }} SKU · {{ totalItems }} item
            </span>
          </p>
          <div class="bg-slate-800/60 border border-slate-700/60 rounded-xl overflow-hidden">
            <div
              v-for="(item, i) in order.items"
              :key="item.id"
              class="flex items-center gap-3 px-4 py-3"
              :class="i < order.items.length - 1 ? 'border-b border-slate-700/60' : ''"
            >
              <!-- Mapping Status Badge -->
              <div class="shrink-0 w-5 h-5 flex items-center justify-center">
                <div v-if="getMappingStatus(item.sku_id) === 'loading'" class="w-4 h-4 border-2 border-slate-500 border-t-transparent rounded-full animate-spin"></div>
                <svg v-else-if="getMappingStatus(item.sku_id) === 'ok'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="text-emerald-400">
                  <path d="M20 6 9 17l-5-5"/>
                </svg>
                <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="text-red-400">
                  <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
                </svg>
              </div>

              <!-- Item Info -->
              <div class="flex-1 min-w-0">
                <p class="text-white font-semibold text-sm truncate">{{ item.product_name }}</p>
                <p class="text-slate-500 font-mono text-xs">{{ item.seller_sku }}</p>
              </div>

              <!-- Qty Badge -->
              <div class="shrink-0">
                <span class="bg-indigo-500/20 text-indigo-300 text-xs font-bold px-2.5 py-1 rounded-md border border-indigo-500/30">
                  x{{ item.item_count }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- SECTION 3: Pilih Kardus -->
        <div v-if="unmappedItems.length === 0">
          <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-3">
            ② Pilih Kardus
            <span class="text-slate-600 normal-case font-normal ml-1">(opsional)</span>
          </p>
          <div class="bg-slate-800/60 border border-slate-700/60 rounded-xl p-4 space-y-3">
            <div>
              <label class="text-slate-400 text-xs font-medium block mb-1.5">Kardus dari Stok Marketplace</label>
              <select
                v-model="selectedBoxStockId"
                class="w-full bg-slate-900 border border-slate-600 text-slate-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/40 focus:border-indigo-500 transition-all"
              >
                <option value="">— Tidak pakai kardus —</option>
                <option v-for="stock in stocks" :key="stock.id" :value="stock.id">
                  {{ stock.product_name }}{{ stock.sku ? ' (' + stock.sku + ')' : '' }} — stok: {{ stock.quantity }}
                </option>
              </select>
            </div>

            <div v-if="selectedBoxStockId">
              <label class="text-slate-400 text-xs font-medium block mb-1.5">Jumlah Kardus</label>
              <div class="flex items-center gap-3">
                <input
                  v-model.number="boxQty"
                  type="number"
                  min="1"
                  class="w-28 bg-slate-900 border border-slate-600 text-slate-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/40 focus:border-indigo-500 transition-all"
                />
                <span class="text-slate-500 text-xs">
                  Stok tersedia:
                  <span :class="selectedBoxStock && boxQty > selectedBoxStock.quantity ? 'text-amber-400 font-bold' : 'text-slate-400'">
                    {{ selectedBoxStock?.quantity ?? 0 }}
                  </span>
                </span>
              </div>
              <div v-if="selectedBoxStock && boxQty > selectedBoxStock.quantity" class="flex items-center gap-1.5 mt-2 text-amber-400 text-xs">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><path d="M12 9v4"/><path d="M12 17h.01"/>
                </svg>
                Stok kardus tidak cukup. Packing tetap bisa dilanjutkan.
              </div>
            </div>
          </div>
        </div>

        <!-- SECTION 3.5: Bonus / Tambahan -->
        <div v-if="unmappedItems.length === 0">
          <div class="flex items-center justify-between mb-3">
            <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider">
              ③ Barang Tambahan / Bonus
              <span class="text-slate-600 normal-case font-normal ml-1">(opsional)</span>
            </p>
            <button
              @click="addBonus"
              class="text-xs text-indigo-400 hover:text-indigo-300 flex items-center gap-1 font-medium bg-indigo-500/10 px-2.5 py-1 rounded-md transition-colors"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
              Tambah Bonus
            </button>
          </div>
          
          <div class="space-y-2">
            <div
              v-for="(bonus, idx) in bonusItems"
              :key="idx"
              class="flex flex-col sm:flex-row items-start sm:items-center gap-3 bg-slate-800/40 border border-slate-700/50 rounded-xl p-3"
            >
              <div class="flex-1 w-full">
                <select
                  v-model="bonus.stock_id"
                  class="w-full bg-slate-900 border border-slate-600 text-slate-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/40 focus:border-indigo-500 transition-all"
                >
                  <option value="">— Pilih Barang Bonus —</option>
                  <option v-for="stock in stocks" :key="stock.id" :value="stock.id">
                    {{ stock.product_name }}{{ stock.sku ? ' (' + stock.sku + ')' : '' }} — stok: {{ stock.quantity }}
                  </option>
                </select>
              </div>
              <div class="flex items-center gap-2 w-full sm:w-auto">
                <input
                  v-model.number="bonus.qty"
                  type="number"
                  min="1"
                  placeholder="Qty"
                  class="w-20 bg-slate-900 border border-slate-600 text-slate-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/40 focus:border-indigo-500 transition-all"
                />
                <button
                  @click="removeBonus(idx)"
                  class="p-2 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors ml-auto sm:ml-0"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M3 6h18"></path>
                    <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
                    <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
                  </svg>
                </button>
              </div>
            </div>
            <div v-if="bonusItems.length === 0" class="text-center py-4 bg-slate-800/20 border border-dashed border-slate-700/50 rounded-xl text-slate-500 text-xs">
              Belum ada barang tambahan.
            </div>
          </div>
        </div>

        <!-- SECTION 4: Ringkasan -->
        <div v-if="unmappedItems.length === 0">
          <p class="text-slate-400 text-xs font-semibold uppercase tracking-wider mb-3">④ Ringkasan</p>
          <div class="bg-slate-800/40 border border-slate-700/40 rounded-xl p-4 space-y-1.5">
            <div class="flex items-start gap-2 text-slate-300 text-sm">
              <span class="text-indigo-400 mt-0.5 shrink-0">•</span>
              <span>
                <span class="font-bold text-white">{{ order.items.length }} SKU</span>,
                total <span class="font-bold text-white">{{ totalItems }}</span> item akan dipack dan stok dipotong otomatis
              </span>
            </div>
            <div v-if="selectedBoxStock" class="flex items-start gap-2 text-slate-300 text-sm">
              <span class="text-indigo-400 mt-0.5 shrink-0">•</span>
              <span>
                Kardus <span class="font-bold text-white">{{ selectedBoxStock.product_name }}</span>
                (<span class="font-bold text-white">{{ boxQty }} pcs</span>) akan dikurangi dari stok
              </span>
            </div>
            <div v-else class="flex items-start gap-2 text-slate-500 text-sm">
              <span class="mt-0.5 shrink-0">•</span>
              <span>Tidak ada kardus yang dipilih</span>
            </div>
            
            <!-- Bonus summary -->
            <div v-if="validBonusItems.length > 0" class="flex items-start gap-2 text-slate-300 text-sm border-t border-slate-700/50 pt-1.5 mt-1.5">
              <span class="text-indigo-400 mt-0.5 shrink-0">•</span>
              <div>
                <span class="font-medium text-slate-300">Tambahan Bonus:</span>
                <ul class="list-disc list-inside mt-0.5 text-slate-400 ml-1">
                  <li v-for="(bonus, idx) in validBonusItems" :key="idx">
                    {{ getStockName(bonus.stock_id) }} (<span class="font-bold text-white">{{ bonus.qty }} pcs</span>)
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>

      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-700">
        <button
          @click="$emit('cancel')"
          class="px-4 py-2 rounded-xl text-slate-400 hover:text-slate-200 hover:bg-slate-800 border border-slate-700 transition-all text-sm font-semibold"
        >
          Batal
        </button>
        <button
          @click="handleConfirm"
          :disabled="!canPack || isConfirming"
          class="flex items-center gap-2 px-5 py-2 bg-gradient-to-r from-emerald-600 to-teal-500 hover:from-emerald-500 hover:to-teal-400 text-white rounded-xl font-bold text-sm shadow-lg shadow-emerald-500/20 transition-all duration-200 disabled:opacity-40 disabled:cursor-not-allowed disabled:shadow-none"
        >
          <span v-if="isConfirming" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 6 9 17l-5-5"/>
          </svg>
          {{ isConfirming ? 'Memproses...' : 'Pack Sekarang' }}
        </button>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import apiClient from '@/api/axios'
import type { MarketplaceOrderGroup, MarketplaceStock } from '@/api/marketplace'

const props = defineProps<{
  order: MarketplaceOrderGroup
  stocks: MarketplaceStock[]
}>()

const emit = defineEmits<{
  confirm: [payload: { box_stock_id?: string; box_qty?: number; additional_items?: {stock_id: string, qty: number}[] }]
  cancel: []
}>()

// Mapping status per sku_id
const mappingStatusMap = ref<Record<string, 'loading' | 'ok' | 'no_mapping'>>({})
const unmappedItems = ref<{ sku_id: string; seller_sku: string; product_name: string }[]>([])

// Box selection
const selectedBoxStockId = ref('')
const boxQty = ref(1)

// Bonus selection
const bonusItems = ref<{stock_id: string, qty: number}[]>([])
const addBonus = () => bonusItems.value.push({ stock_id: '', qty: 1 })
const removeBonus = (idx: number) => bonusItems.value.splice(idx, 1)
const validBonusItems = computed(() => bonusItems.value.filter(b => b.stock_id && b.qty > 0))

const getStockName = (id: string) => {
  const stock = props.stocks.find(s => s.id === id)
  return stock ? stock.product_name : 'Unknown'
}

const isConfirming = ref(false)

const totalItems = computed(() =>
  props.order.items.reduce((sum, i) => sum + i.item_count, 0)
)

const selectedBoxStock = computed(() =>
  props.stocks.find(s => s.id === selectedBoxStockId.value) ?? null
)

const getMappingStatus = (skuId: string): 'loading' | 'ok' | 'no_mapping' => {
  return mappingStatusMap.value[skuId] ?? 'loading'
}

const canPack = computed(() =>
  unmappedItems.value.length === 0 &&
  !isConfirming.value &&
  Object.values(mappingStatusMap.value).every(s => s !== 'loading')
)

onMounted(async () => {
  const uniqueSkuIds = [...new Set(
    props.order.items.map(i => i.sku_id).filter(Boolean)
  )]

  uniqueSkuIds.forEach(id => {
    mappingStatusMap.value[id] = 'loading'
  })

  const results = await Promise.allSettled(
    uniqueSkuIds.map(id =>
      apiClient.get('/marketplace/mappings', { params: { sku_id: id } })
    )
  )

  results.forEach((result, i) => {
    const id = uniqueSkuIds[i] as string
    const item = props.order.items.find(it => it.sku_id === id)

    if (result.status === 'fulfilled' && Array.isArray(result.value.data) && result.value.data.length > 0) {
      mappingStatusMap.value[id] = 'ok'
    } else {
      mappingStatusMap.value[id] = 'no_mapping'
      unmappedItems.value.push({
        sku_id: id,
        seller_sku: item?.seller_sku ?? id,
        product_name: item?.product_name ?? ''
      })
    }
  })
})

const handleConfirm = () => {
  if (!canPack.value) return
  isConfirming.value = true

  const payload: any = {}
  if (selectedBoxStockId.value) {
    payload.box_stock_id = selectedBoxStockId.value
    payload.box_qty = boxQty.value
  }

  if (validBonusItems.value.length > 0) {
    payload.additional_items = validBonusItems.value
  }

  emit('confirm', payload)
}
</script>
