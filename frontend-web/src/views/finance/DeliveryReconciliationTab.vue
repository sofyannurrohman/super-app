<script setup lang="ts">
import { ref, computed } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import {
  Search,
  CheckCircle2,
  RefreshCw,
  Truck,
  DollarSign,
  AlertTriangle
} from 'lucide-vue-next'
import { useMasterDataStore } from '@/stores/masterData'
import { onMounted, watch } from 'vue'

const props = defineProps({
  suggestedCash: { type: Number, default: 0 }
})

const emit = defineEmits(['delivery-loaded', 'delivery-cleared', 'delivery-approved'])
const masterDataStore = useMasterDataStore()

const isManualSearch = ref(false)
const todayDeliveries = ref<any[]>([])

const searchDoNo = ref('')
const loading = ref(false)
const deliveryData = ref<any>(null)
const approving = ref(false)

async function fetchTodayDeliveries() {
  try {
    const res = await apiClient.get('/finance/deliveries/today', {
      params: { branch_id: masterDataStore.selectedBranchId }
    })
    todayDeliveries.value = res.data.data || []
  } catch (err) {
    console.error('Failed to fetch today deliveries', err)
  }
}

watch(() => masterDataStore.selectedBranchId, () => {
  fetchTodayDeliveries()
  searchDoNo.value = ''
})

onMounted(() => {
  fetchTodayDeliveries()
})
const physicalCash = ref<number | null>(null)

const isCashMatched = computed(() => {
  if (!deliveryData.value || physicalCash.value === null) return false
  return physicalCash.value === deliveryData.value.total_cash
})

const cashDifference = computed(() => {
  if (!deliveryData.value || physicalCash.value === null) return 0
  return physicalCash.value - deliveryData.value.total_cash
})

watch(() => props.suggestedCash, (newVal) => {
  if (deliveryData.value && deliveryData.value.cash_settled_at === null) {
    physicalCash.value = newVal
  }
})

async function fetchDeliverySummary() {
  if (!searchDoNo.value) {
    toast.warning('Silakan masukkan Nomor Surat Jalan')
    return
  }
  
  loading.value = true
  deliveryData.value = null
  physicalCash.value = null
  try {
    const res = await apiClient.get('/finance/deliveries/summary', { 
      params: { do_no: searchDoNo.value } 
    })
    deliveryData.value = res.data.data
    emit('delivery-loaded', deliveryData.value)
  } catch (err: any) {
    emit('delivery-cleared')
    toast.error(err?.response?.data?.error || 'Gagal memuat data Surat Jalan')
  } finally {
    loading.value = false
  }
}

async function approveDelivery() {
  if (!deliveryData.value) return
  if (!isCashMatched.value && cashDifference.value !== 0) {
    if (!confirm(`Terdapat selisih kas sebesar ${formatCurrency(cashDifference.value)}. Surat jalan akan dialihkan ke status Cross-check. Lanjutkan?`)) {
      return
    }
  }

  approving.value = true
  try {
    const res = await apiClient.post('/finance/deliveries/approve', {
      batch_id: deliveryData.value.batch_id,
      physical_cash: physicalCash.value,
    })
    
    if (res.data.status === 'WAITING_CROSSCHECK') {
      toast.warning('Terdapat selisih! Surat Jalan ditahan dan dialihkan ke tahap Cross-check.')
    } else {
      toast.success('Surat Jalan berhasil divalidasi! Semua transaksi setoran telah disetujui.')
    }
    
    emit('delivery-approved')
    await fetchDeliverySummary()
  } catch (err: any) {
    toast.error('Gagal menyetujui Surat Jalan: ' + (err.response?.data?.error || err.message))
  } finally {
    approving.value = false
  }
}

const crosscheckNotes = ref('')
const resolving = ref(false)

async function resolveCrosscheck() {
  if (!confirm('Yakin ingin menyelesaikan selisih ini secara manual? Pastikan Anda sudah berkoordinasi dengan Admin Gudang.')) {
    return
  }
  resolving.value = true
  try {
    await apiClient.post('/finance/deliveries/resolve-crosscheck', {
      batch_id: deliveryData.value.batch_id,
      notes: crosscheckNotes.value
    })
    toast.success('Cross-check selesai! Surat jalan berhasil ditutup.')
    emit('delivery-approved')
    await fetchDeliverySummary()
  } catch (err: any) {
    toast.error('Gagal menyelesaikan cross-check: ' + (err.response?.data?.error || err.message))
  } finally {
    resolving.value = false
  }
}

const formatCurrency = (val: number) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val || 0)

</script>

<template>
  <div class="space-y-6">
    <!-- Filters -->
    <div class="bg-white p-5 rounded-3xl border border-slate-200/60 shadow-sm">
      <div class="flex items-center justify-between mb-3">
        <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider">Surat Jalan Hari Ini</label>
        <label class="flex items-center gap-2 text-xs text-slate-600 font-medium cursor-pointer">
          <input type="checkbox" v-model="isManualSearch" class="rounded border-slate-300 text-primary focus:ring-primary h-3.5 w-3.5">
          Cari Manual (Teks / Scan)
        </label>
      </div>

      <div class="flex flex-wrap gap-4 items-end">
        <div class="flex-1 min-w-[250px]">
          <div class="relative">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            
            <select
              v-if="!isManualSearch"
              v-model="searchDoNo"
              class="w-full pl-9 pr-4 py-2.5 rounded-xl border border-slate-200 focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none text-sm font-medium transition-all appearance-none"
            >
              <option value="">-- Pilih Surat Jalan --</option>
              <option v-for="doItem in todayDeliveries" :key="doItem.id" :value="doItem.delivery_order_no">
                {{ doItem.delivery_order_no }} ({{ doItem.driver_name }})
              </option>
            </select>

            <input
              v-else
              v-model="searchDoNo"
              type="text"
              placeholder="Ketik atau Scan Barcode Surat Jalan..."
              @keyup.enter="fetchDeliverySummary"
              class="w-full pl-9 pr-4 py-2.5 rounded-xl border border-slate-200 focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none text-sm font-medium transition-all"
            />
          </div>
        </div>

        <button
          @click="fetchDeliverySummary"
          :disabled="loading || !searchDoNo"
          class="flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white text-sm font-bold transition-all shadow-md disabled:opacity-50"
        >
          <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
          Cari Data
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!deliveryData && !loading && searchDoNo" class="bg-white rounded-3xl p-12 text-center border border-slate-200 border-dashed">
      <Truck class="w-12 h-12 text-slate-300 mx-auto mb-4" />
      <h3 class="text-lg font-bold text-slate-700">Surat Jalan Belum Dicari</h3>
      <p class="text-slate-500 mt-2 text-sm max-w-md mx-auto">Silakan ketik atau scan nomor Surat Jalan dan tekan Enter atau klik Cari Data.</p>
    </div>

    <!-- Validation Card -->
    <div v-if="deliveryData" class="grid grid-cols-1 gap-6">
      
      <!-- Data dari SJ -->
      <div class="bg-white rounded-[2rem] border border-slate-200 shadow-sm overflow-hidden flex flex-col">
        <div class="bg-slate-50 p-6 border-b border-slate-200 flex items-center gap-4">
          <div class="w-12 h-12 bg-white rounded-xl shadow-sm flex items-center justify-center text-primary">
            <Truck class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-lg font-black text-slate-900">Surat Jalan: {{ deliveryData.delivery_order_no }}</h3>
            <p class="text-sm font-medium text-slate-500">Supir: <span class="text-slate-700 font-bold">{{ deliveryData.driver_name }}</span> &bull; Status SJ: <span class="text-indigo-600 font-bold uppercase">{{ deliveryData.status }}</span></p>
          </div>
        </div>
        <div class="p-6 flex-1 space-y-4">
          <div class="flex justify-between items-center p-4 bg-slate-50 rounded-2xl border border-slate-100">
            <span class="text-slate-600 font-semibold">Total Tunai (CASH)</span>
            <span class="text-2xl font-black text-slate-900">{{ formatCurrency(deliveryData.total_cash) }}</span>
          </div>
          <div class="flex justify-between items-center p-4 bg-slate-50 rounded-2xl border border-slate-100">
            <span class="text-slate-600 font-semibold">Total Transfer/QRIS</span>
            <span class="text-xl font-bold text-slate-700">{{ formatCurrency(deliveryData.total_transfer) }}</span>
          </div>
          <div class="flex justify-between items-center p-4 bg-slate-50 rounded-2xl border border-slate-100">
            <span class="text-slate-600 font-semibold">Jumlah Tagihan Pending</span>
            <span class="text-xl font-bold text-slate-700">{{ deliveryData.pending_payment_ids?.length || 0 }} Transaksi</span>
          </div>
        </div>
      </div>

      <!-- Hitung Uang Fisik (Rekonsiliasi) -->
      <div class="bg-gradient-to-br from-indigo-900 to-slate-900 rounded-[2rem] shadow-xl p-8 relative overflow-hidden flex flex-col justify-between border border-slate-800">
        <!-- Glow -->
        <div class="absolute -right-20 -top-20 w-64 h-64 bg-indigo-500/20 blur-3xl rounded-full"></div>
        
        <div class="relative z-10">
          <div class="flex items-center gap-3 mb-6">
            <DollarSign class="w-6 h-6 text-emerald-400" />
            <h3 class="text-xl font-black text-white">Hitung Fisik Uang & Validasi</h3>
          </div>
          
          <div class="space-y-4 mb-8">
            <label class="block text-sm font-bold text-slate-300">Total Uang Tunai yang Diterima Kasir (Rp)</label>
            <input
              v-model.number="physicalCash"
              type="number"
              placeholder="0"
              class="w-full bg-white/10 text-white placeholder:text-slate-500 border border-white/20 rounded-2xl px-6 py-4 text-3xl font-black focus:outline-none focus:border-emerald-400 focus:ring-2 focus:ring-emerald-400/20 transition-all"
              :disabled="deliveryData.cash_settled_at !== null"
            />
          </div>

          <div v-if="typeof physicalCash === 'number'">
            <div v-if="cashDifference === 0" class="flex items-center gap-3 p-4 bg-emerald-500/20 border border-emerald-500/30 rounded-2xl mb-6">
              <CheckCircle2 class="w-6 h-6 text-emerald-400" />
              <div>
                <p class="text-emerald-400 font-bold text-sm">MATCHED! Jumlah uang pas.</p>
                <p class="text-emerald-200/70 text-xs">Setoran fisik sesuai dengan pengumpulan supir.</p>
              </div>
            </div>
            
            <div v-else class="flex items-start gap-3 p-4 bg-rose-500/20 border border-rose-500/30 rounded-2xl mb-6">
              <AlertTriangle class="w-6 h-6 text-rose-400 shrink-0" />
              <div>
                <p class="text-rose-400 font-bold text-sm">
                  SELISIH: {{ formatCurrency(Math.abs(cashDifference)) }} 
                  ({{ cashDifference > 0 ? 'LEBIH' : 'KURANG' }})
                </p>
                <p class="text-rose-200/70 text-xs mt-1">Uang fisik tidak sesuai dengan catatan sistem.</p>
              </div>
            </div>
          </div>
        </div>

        <!-- WAITING CROSSCHECK BLOCK -->
        <div v-if="deliveryData.status === 'WAITING_CROSSCHECK' && deliveryData.cash_settled_at === null" class="relative z-10 w-full mt-4 p-4 rounded-2xl border border-amber-500/30 bg-amber-500/10">
          <div class="flex items-start gap-3 mb-4">
            <AlertTriangle class="w-6 h-6 text-amber-400 shrink-0" />
            <div>
              <p class="text-amber-400 font-bold">STATUS: MENUNGGU CROSS-CHECK SELISIH</p>
              <p class="text-amber-200/70 text-xs mt-1">Surat jalan ini ditahan karena terdapat selisih. Silakan berkoordinasi secara offline dengan Admin Gudang. Jika sudah selesai, Anda dapat menutup surat jalan secara paksa di bawah ini.</p>
            </div>
          </div>
          
          <input
            v-model="crosscheckNotes"
            type="text"
            placeholder="Catatan Penyelesaian (Misal: Dipotong ke kasbon supir)"
            class="w-full bg-white/10 text-white placeholder:text-slate-400 border border-white/20 rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-amber-400 focus:ring-1 focus:ring-amber-400/50 transition-all mb-4"
          />
          
          <button
            @click="resolveCrosscheck"
            :disabled="resolving"
            class="w-full py-3 rounded-xl font-bold text-sm flex items-center justify-center gap-2 transition-all shadow-lg active:scale-[0.98] bg-amber-500 hover:bg-amber-400 text-slate-900 shadow-amber-500/20"
          >
            <span v-if="resolving" class="w-4 h-4 rounded-full border-2 border-current border-t-transparent animate-spin"></span>
            <template v-else>
              <CheckCircle2 class="w-5 h-5" />
              SELESAIKAN SELISIH & TUTUP BUKU
            </template>
          </button>
        </div>

        <button
          v-else-if="deliveryData.cash_settled_at === null"
          @click="approveDelivery"
          :disabled="approving || physicalCash === null"
          class="relative z-10 w-full py-4 rounded-2xl font-black text-lg flex items-center justify-center gap-3 transition-all shadow-lg active:scale-[0.98]"
          :class="isCashMatched ? 'bg-emerald-500 hover:bg-emerald-400 text-slate-900 shadow-emerald-500/20' : 'bg-white/10 hover:bg-white/20 text-white backdrop-blur-md'"
        >
          <span v-if="approving" class="w-5 h-5 rounded-full border-2 border-current border-t-transparent animate-spin"></span>
          <template v-else>
            <CheckCircle2 class="w-6 h-6" />
            TERIMA SETORAN 
          </template>
        </button>
        
        <div v-else class="relative z-10 w-full py-4 rounded-2xl font-black text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center gap-3">
          <CheckCircle2 class="w-6 h-6" />
          SUDAH DITUTUP BUKU / SELESAI
        </div>
      </div>
    </div>
  </div>
</template>
