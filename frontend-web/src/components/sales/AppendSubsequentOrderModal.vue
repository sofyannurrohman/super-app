<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200"
  >
    <div
      class="bg-white rounded-3xl shadow-2xl border border-slate-200/80 w-full max-w-4xl max-h-[90vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200"
    >
      <!-- Header -->
      <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-gradient-to-r from-amber-500/10 via-amber-500/5 to-transparent">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-2xl bg-amber-500/20 text-amber-600 flex items-center justify-center font-bold">
            <PlusCircle class="w-5 h-5" />
          </div>
          <div>
            <h3 class="text-base font-black text-slate-900 tracking-tight flex items-center gap-2">
              Susulkan Pesanan / Nota ke Surat Jalan
              <span class="text-xs bg-amber-100 text-amber-800 font-bold px-2 py-0.5 rounded-full border border-amber-300">
                {{ batch?.delivery_order_no || 'SJ' }}
              </span>
            </h3>
            <p class="text-xs text-slate-500 font-medium mt-0.5">
              Pilih pesanan terkonfirmasi yang akan disisipkan ke Surat Jalan ini.
            </p>
          </div>
        </div>
        <button
          @click="close"
          class="w-8 h-8 rounded-full bg-slate-100 hover:bg-slate-200 text-slate-500 flex items-center justify-center transition-colors"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Info Bar -->
      <div class="bg-slate-50 px-6 py-3 border-b border-slate-100 grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs">
        <div>
          <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Sopir / Armada</span>
          <span class="font-black text-slate-800">
            {{ batch?.driver ? `${batch.driver.first_name} ${batch.driver.last_name || ''}` : (batch?.daily_driver?.name || '—') }}
            {{ batch?.vehicle ? `(${batch.vehicle.license_plate})` : '' }}
          </span>
        </div>
        <div>
          <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Tanggal Kirim</span>
          <span class="font-black text-slate-800">{{ formatDate(batch?.delivery_date || batch?.created_at) }}</span>
        </div>
        <div>
          <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Muatan Saat Ini</span>
          <span class="font-black text-slate-800">{{ batch?.items?.length || 0 }} Pesanan / Nota</span>
        </div>
        <div>
          <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider block">Status Surat Jalan</span>
          <span class="inline-block font-black text-[10px] uppercase px-2 py-0.5 rounded bg-blue-100 text-blue-800">
            {{ batch?.status || 'PENDING' }}
          </span>
        </div>
      </div>

      <!-- Search & Filters -->
      <div class="px-6 py-3 border-b border-slate-100 bg-white flex flex-col sm:flex-row gap-3 items-center justify-between">
        <div class="relative w-full sm:w-80">
          <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            v-model="searchQuery"
            @input="handleSearch"
            type="text"
            placeholder="Cari No SO, nama toko, salesman..."
            class="w-full pl-9 pr-4 py-2 text-xs rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 font-medium"
          />
        </div>
        <div class="text-xs text-slate-500 font-bold self-end sm:self-center">
          Ditemukan: <span class="text-slate-900 font-black">{{ eligibleOrders.length }}</span> pesanan siap susul
        </div>
      </div>

      <!-- Order List Table -->
      <div class="flex-1 overflow-y-auto p-6 space-y-3 min-h-[250px]">
        <!-- Loading State -->
        <div v-if="isLoading" class="py-16 text-center">
          <div class="inline-block animate-spin w-8 h-8 border-4 border-amber-500 border-t-transparent rounded-full mb-2"></div>
          <p class="text-xs font-bold text-slate-500">Memuat daftar pesanan yang dapat disusulkan...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="eligibleOrders.length === 0" class="py-16 text-center space-y-2">
          <div class="w-12 h-12 rounded-full bg-slate-100 text-slate-400 flex items-center justify-center mx-auto">
            <Package class="w-6 h-6" />
          </div>
          <p class="text-sm font-black text-slate-700">Tidak ada pesanan yang tersedia</p>
          <p class="text-xs text-slate-400 max-w-sm mx-auto font-medium">
            Tidak ditemukan Sales Order berstatus CONFIRMED di cabang ini yang belum masuk Surat Jalan.
          </p>
        </div>

        <!-- Orders Table -->
        <div v-else class="border border-slate-200 rounded-2xl overflow-hidden bg-white shadow-sm">
          <table class="w-full text-left text-xs border-collapse">
            <thead>
              <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-black">
                <th class="p-3 w-10 text-center">
                  <input
                    type="checkbox"
                    :checked="isAllSelected"
                    :indeterminate="isIndeterminate"
                    @change="toggleSelectAll"
                    class="rounded border-slate-300 text-amber-600 focus:ring-amber-500 w-4 h-4 cursor-pointer"
                  />
                </th>
                <th class="p-3">No. SO & Tanggal</th>
                <th class="p-3">Toko / Customer</th>
                <th class="p-3">Salesman</th>
                <th class="p-3 text-center">Total Item / Qty</th>
                <th class="p-3 text-right">Total Nilai</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="order in eligibleOrders"
                :key="order.id"
                @click="toggleSelect(order.id)"
                :class="selectedOrderIds.includes(order.id) ? 'bg-amber-50/60' : 'hover:bg-slate-50/60'"
                class="cursor-pointer transition-colors"
              >
                <td class="p-3 text-center" @click.stop>
                  <input
                    type="checkbox"
                    :value="order.id"
                    v-model="selectedOrderIds"
                    class="rounded border-slate-300 text-amber-600 focus:ring-amber-500 w-4 h-4 cursor-pointer"
                  />
                </td>
                <td class="p-3">
                  <span class="font-black text-slate-900 block">{{ order.so_number }}</span>
                  <span class="text-[10px] text-slate-400 font-medium">{{ formatDate(order.created_at) }}</span>
                </td>
                <td class="p-3">
                  <span class="font-black text-slate-800 block">{{ order.store?.name || '—' }}</span>
                  <span class="text-[10px] text-slate-400 font-medium block truncate max-w-xs">{{ order.store?.address || '—' }}</span>
                </td>
                <td class="p-3 font-bold text-slate-700">
                  {{ order.employee?.name || '—' }}
                </td>
                <td class="p-3 text-center">
                  <span class="inline-block px-2 py-0.5 rounded-full bg-slate-100 font-black text-[10px] text-slate-700">
                    {{ order.items?.length || 0 }} SKU ({{ getTotalQty(order) }} pcs)
                  </span>
                </td>
                <td class="p-3 text-right font-black text-slate-900">
                  Rp {{ (order.total_amount || 0).toLocaleString('id-ID') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Notes Input -->
        <div v-if="eligibleOrders.length > 0" class="mt-4 pt-2">
          <label class="block text-xs font-bold text-slate-700 mb-1">Catatan Susulan (Opsional)</label>
          <input
            v-model="susulanNotes"
            type="text"
            placeholder="Contoh: Susulan order toko pagi, tolong diselipkan"
            class="w-full px-3.5 py-2 text-xs rounded-xl border border-slate-200 focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 font-medium"
          />
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-4 border-t border-slate-100 bg-slate-50/80 flex flex-col sm:flex-row items-center justify-between gap-4">
        <div class="text-xs text-slate-600 font-bold">
          <span class="text-amber-600 font-black text-sm">{{ selectedOrderIds.length }}</span> pesanan dipilih
          <span v-if="selectedOrderIds.length > 0" class="text-slate-400 ml-1">
            (Total Nilai: <span class="text-slate-900 font-black">Rp {{ selectedTotalAmount.toLocaleString('id-ID') }}</span>)
          </span>
        </div>
        <div class="flex items-center gap-2 w-full sm:w-auto">
          <Button
            type="button"
            variant="outline"
            @click="close"
            :disabled="isSubmitting"
            class="flex-1 sm:flex-none rounded-xl text-xs font-bold"
          >
            Batal
          </Button>
          <Button
            type="button"
            @click="handleSubmit"
            :disabled="isSubmitting || selectedOrderIds.length === 0"
            class="flex-1 sm:flex-none bg-amber-500 hover:bg-amber-600 text-white rounded-xl text-xs font-black shadow-md shadow-amber-200 gap-1.5"
          >
            <PlusCircle class="w-4 h-4" />
            {{ isSubmitting ? 'Menyimpan...' : 'Susulkan ke Surat Jalan' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { PlusCircle, Search, X, Package } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import { useDeliveryStore } from '@/stores/delivery'

const props = defineProps<{
  isOpen: boolean
  batch: any
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success'): void
}>()

const deliveryStore = useDeliveryStore()
const isLoading = ref(false)
const isSubmitting = ref(false)
const searchQuery = ref('')
const eligibleOrders = ref<any[]>([])
const selectedOrderIds = ref<string[]>([])
const susulanNotes = ref('')

let searchTimeout: any = null

const handleSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    fetchEligibleOrders()
  }, 300)
}

const fetchEligibleOrders = async () => {
  if (!props.batch?.id) return
  isLoading.value = true
  try {
    const orders = await deliveryStore.fetchEligibleOrdersForBatch(props.batch.id, searchQuery.value)
    const existingOrderIds = new Set(
      (props.batch?.items || [])
        .map((item: any) => item.sales_order_id || item.sales_order?.id)
        .filter(Boolean)
    )
    eligibleOrders.value = (orders || []).filter((order: any) => !existingOrderIds.has(order.id))
  } catch (err: any) {
    toast.error(err?.response?.data?.error || 'Gagal memuat pesanan yang dapat disusulkan')
  } finally {
    isLoading.value = false
  }
}

watch(
  () => props.isOpen,
  (newVal) => {
    if (newVal) {
      selectedOrderIds.value = []
      susulanNotes.value = ''
      searchQuery.value = ''
      fetchEligibleOrders()
    }
  }
)

const toggleSelect = (orderId: string) => {
  const index = selectedOrderIds.value.indexOf(orderId)
  if (index > -1) {
    selectedOrderIds.value.splice(index, 1)
  } else {
    selectedOrderIds.value.push(orderId)
  }
}

const isAllSelected = computed(() => {
  return eligibleOrders.value.length > 0 && selectedOrderIds.value.length === eligibleOrders.value.length
})

const isIndeterminate = computed(() => {
  return selectedOrderIds.value.length > 0 && selectedOrderIds.value.length < eligibleOrders.value.length
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedOrderIds.value = []
  } else {
    selectedOrderIds.value = eligibleOrders.value.map((o) => o.id)
  }
}

const getTotalQty = (order: any) => {
  if (!order.items || !order.items.length) return 0
  return order.items.reduce((acc: number, item: any) => acc + (item.quantity || 0), 0)
}

const selectedTotalAmount = computed(() => {
  return eligibleOrders.value
    .filter((o) => selectedOrderIds.value.includes(o.id))
    .reduce((acc, o) => acc + (o.total_amount || 0), 0)
})

const formatDate = (date: string) => {
  if (!date) return '—'
  try {
    return new Intl.DateTimeFormat('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    }).format(new Date(date))
  } catch (e) {
    return date
  }
}

const close = () => {
  emit('close')
}

const handleSubmit = async () => {
  if (!props.batch?.id || selectedOrderIds.value.length === 0) return
  isSubmitting.value = true
  try {
    await deliveryStore.appendOrdersToBatch(props.batch.id, {
      sales_order_ids: selectedOrderIds.value,
      notes: susulanNotes.value || 'Nota Susulan',
    })
    toast.success(`${selectedOrderIds.value.length} pesanan berhasil disusulkan ke Surat Jalan`)
    emit('success')
    close()
  } catch (err: any) {
    toast.error(err?.response?.data?.error || 'Gagal menyusulkan pesanan ke Surat Jalan')
  } finally {
    isSubmitting.value = false
  }
}
</script>
