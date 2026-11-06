<template>
  <div
    v-if="open"
    class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-50 flex items-center justify-center p-4 overflow-y-auto animate-in fade-in duration-200"
    @click.self="close"
  >
    <Card
      class="w-full max-w-2xl bg-white shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200"
    >
      <!-- HEADER -->
      <CardHeader class="p-6 border-b border-slate-100 bg-gradient-to-r from-amber-500/10 via-amber-50/50 to-orange-500/10">
        <div class="flex items-start justify-between gap-4">
          <div class="flex items-start gap-3">
            <div
              class="w-12 h-12 rounded-2xl bg-amber-500/15 border border-amber-500/30 flex items-center justify-center text-amber-700 shadow-sm shrink-0"
            >
              <Boxes class="h-6 w-6" />
            </div>
            <div>
              <div class="flex items-center gap-2 mb-1">
                <Badge variant="outline" class="bg-amber-100 text-amber-800 border-amber-300 font-bold text-[10px]">
                  Rincian Komposisi Produksi
                </Badge>
                <span class="text-xs text-slate-400 font-medium">
                  {{ formatDate(productionLog?.production_date) }}
                </span>
              </div>
              <CardTitle class="text-xl font-extrabold text-slate-900">
                {{ productionLog?.product?.name || 'Produk Jadi' }}
              </CardTitle>
              <CardDescription class="text-xs text-slate-500 flex flex-wrap items-center gap-x-3 gap-y-1 mt-1 font-medium">
                <span v-if="productionLog?.product?.sku">SKU: <strong class="text-slate-700">{{ productionLog.product.sku }}</strong></span>
                <span v-if="productionLog?.factory?.name">Pabrik: <strong class="text-slate-700">{{ productionLog.factory.name }}</strong></span>
                <span v-if="productionLog?.target_warehouse">Gudang: <strong class="text-slate-700">{{ productionLog.target_warehouse?.name || 'Utama' }}</strong></span>
              </CardDescription>
            </div>
          </div>
          <Button
            variant="ghost"
            size="icon"
            @click="close"
            class="rounded-full hover:bg-slate-200/60 text-slate-400 hover:text-slate-700 shrink-0"
          >
            <X class="h-5 w-5" />
          </Button>
        </div>

        <!-- PRODUCTION RUN SUMMARY CARD -->
        <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 mt-4 pt-3 border-t border-amber-200/50">
          <div class="p-2.5 bg-white/80 rounded-2xl border border-amber-100 shadow-2xs">
            <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Batch Produksi</div>
            <div class="flex items-center gap-1.5 mt-0.5">
              <span class="font-mono font-bold text-sm text-slate-800">{{ productionLog?.batch_no || '-' }}</span>
              <button
                v-if="productionLog?.batch_no"
                type="button"
                @click="copyText(productionLog.batch_no)"
                title="Salin Batch"
                class="text-slate-400 hover:text-amber-600 transition-colors"
              >
                <Copy class="w-3.5 h-3.5" />
              </button>
            </div>
            <div v-if="productionLog?.expiry_date" class="text-[10px] text-slate-400 font-medium">
              Exp: {{ formatDate(productionLog.expiry_date) }}
            </div>
          </div>

          <div class="p-2.5 bg-white/80 rounded-2xl border border-amber-100 shadow-2xs">
            <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Hasil Produksi</div>
            <div class="font-black text-sm text-emerald-600 mt-0.5">
              +{{ productionLog?.quantity || 0 }} {{ productionLog?.product?.unit || 'PCS' }}
            </div>
            <div v-if="productionLog?.carton_count" class="text-[10px] text-slate-400 font-medium">
              {{ productionLog.carton_count }} Karton @ {{ productionLog.pieces_per_carton }} isi
            </div>
          </div>

          <div class="col-span-2 sm:col-span-1 p-2.5 bg-white/80 rounded-2xl border border-amber-100 shadow-2xs">
            <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Bahan Terpakai</div>
            <div class="font-black text-sm text-amber-900 mt-0.5">
              {{ materials.length }} Komponen
            </div>
            <div class="text-[10px] text-slate-400 font-medium">
              Formula / Resep
            </div>
          </div>
        </div>
      </CardHeader>

      <!-- CONTENT BODY -->
      <CardContent class="p-6 bg-slate-50/50 space-y-4 max-h-[55vh] overflow-y-auto custom-scrollbar">
        <!-- Search Filter if multiple items -->
        <div v-if="materials.length > 5" class="relative">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            type="text"
            v-model="searchQuery"
            placeholder="Cari bahan baku atau nomor batch..."
            class="w-full pl-10 pr-4 py-2 bg-white border border-slate-200 rounded-xl text-xs font-medium outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-400 transition-all"
          />
        </div>

        <div v-if="filteredMaterials.length === 0" class="py-12 text-center text-slate-400">
          <Boxes class="h-10 w-10 mx-auto text-slate-300 mb-2" />
          <p class="font-bold text-sm">Tidak ada rincian bahan baku</p>
          <p class="text-xs text-slate-400 mt-0.5">
            Log produksi ini tidak mencatat pemotongan bahan baku atau tidak cocok dengan filter.
          </p>
        </div>

        <!-- Materials Table -->
        <div v-else class="border border-slate-200 rounded-2xl bg-white overflow-hidden shadow-xs">
          <table class="w-full text-xs text-left">
            <thead class="bg-slate-100/80 text-[10px] font-black text-slate-500 uppercase tracking-wider border-b border-slate-200">
              <tr>
                <th class="px-4 py-3 w-10 text-center">No</th>
                <th class="px-4 py-3">Nama Bahan Baku</th>
                <th class="px-4 py-3">Batch Bahan Terpakai</th>
                <th class="px-4 py-3 text-right">Jumlah Terpakai</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="(item, idx) in filteredMaterials"
                :key="item.id || idx"
                class="hover:bg-amber-50/30 transition-colors"
              >
                <td class="px-4 py-3 text-center font-bold text-slate-400">
                  {{ Number(idx) + 1 }}
                </td>
                <td class="px-4 py-3">
                  <div class="font-bold text-slate-900">
                    {{ item.product?.name || 'Bahan Baku' }}
                  </div>
                  <div v-if="item.product?.sku" class="text-[10px] text-slate-400 font-mono">
                    SKU: {{ item.product.sku }}
                  </div>
                  <div v-if="item.notes" class="text-[10px] text-amber-800 italic mt-0.5">
                    "{{ item.notes }}"
                  </div>
                </td>
                <td class="px-4 py-3">
                  <div class="inline-flex items-center gap-1.5">
                    <span class="font-mono text-[11px] font-bold px-2 py-0.5 rounded-lg border border-amber-200 bg-amber-50 text-amber-900">
                      {{ item.batch_no || 'Auto' }}
                    </span>
                    <button
                      v-if="item.batch_no"
                      type="button"
                      @click="copyText(item.batch_no)"
                      title="Salin Batch"
                      class="text-slate-300 hover:text-amber-700 transition-colors"
                    >
                      <Copy class="w-3 h-3" />
                    </button>
                  </div>
                </td>
                <td class="px-4 py-3 text-right">
                  <span class="font-black text-slate-900 text-sm">
                    {{ item.quantity }}
                  </span>
                  <span class="text-[10px] font-bold text-slate-500 uppercase ml-1">
                    {{ item.product?.unit || '' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>

      <!-- FOOTER -->
      <CardFooter class="p-4 bg-white border-t border-slate-100 flex items-center justify-between">
        <div class="text-xs text-slate-500 font-medium">
          Total: <strong class="text-slate-800">{{ materials.length }} Bahan Baku</strong>
        </div>
        <Button
          variant="outline"
          @click="close"
          class="rounded-xl px-6 font-bold text-xs border-slate-200 hover:bg-slate-100"
        >
          Tutup
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Boxes, X, Copy, Search } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const props = defineProps<{
  open: boolean
  productionLog: any
}>()

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
}>()

const searchQuery = ref('')

const close = () => {
  emit('update:open', false)
  searchQuery.value = ''
}

const materials = computed(() => {
  return props.productionLog?.materials || []
})

const filteredMaterials = computed(() => {
  if (!searchQuery.value.trim()) return materials.value
  const q = searchQuery.value.toLowerCase().trim()
  return materials.value.filter((m: any) => {
    const name = (m.product?.name || '').toLowerCase()
    const sku = (m.product?.sku || '').toLowerCase()
    const batch = (m.batch_no || '').toLowerCase()
    const notes = (m.notes || '').toLowerCase()
    return name.includes(q) || sku.includes(q) || batch.includes(q) || notes.includes(q)
  })
})

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const copyText = (text: string) => {
  if (!text) return
  navigator.clipboard.writeText(text)
  toast.success(`Batch ${text} disalin ke clipboard`)
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 5px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}
</style>
