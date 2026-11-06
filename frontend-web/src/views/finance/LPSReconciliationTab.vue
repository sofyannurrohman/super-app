<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import {
  Calendar,
  User,
  Search,
  CheckCircle2,
  RefreshCw,
  FileText,
  DollarSign,
  AlertTriangle,
  Printer
} from 'lucide-vue-next'

const emit = defineEmits(['lps-loaded', 'lps-cleared', 'lps-approved', 'preview-lps'])

const masterStore = useMasterDataStore()
const authStore = useAuthStore()

const selectedDate = ref(new Date().toISOString().split('T')[0])
const selectedEmployeeId = ref('')
const employees = ref<any[]>([])
const loading = ref(false)
const lpsData = ref<any>(null)
const approving = ref(false)

const physicalCash = ref<number | null>(null)

const isCashMatched = computed(() => {
  if (!lpsData.value || physicalCash.value === null) return false
  return physicalCash.value === lpsData.value.total_cash
})

const cashDifference = computed(() => {
  if (!lpsData.value || physicalCash.value === null) return 0
  return physicalCash.value - lpsData.value.total_cash
})

async function fetchEmployees() {
  try {
    const companyId = masterStore.selectedBranchCompanyId || ''
    const params: any = { is_sales: true, limit: 'all', company_id: companyId }
    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterStore.selectedBranchId
    }
    const res = await apiClient.get('/employees', { params })
    employees.value = res.data?.data || res.data || []
  } catch {
    employees.value = []
  }
}

watch(() => masterStore.selectedBranchId, () => {
  fetchEmployees()
  selectedEmployeeId.value = ''
  lpsData.value = null
  emit('lps-cleared')
})

async function fetchLPS() {
  if (!selectedDate.value || !selectedEmployeeId.value) return
  
  loading.value = true
  lpsData.value = null
  physicalCash.value = null
  try {
    const params: any = { 
      date: selectedDate.value,
      employee_id: selectedEmployeeId.value
    }
    if (masterStore.selectedBranchId && masterStore.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterStore.selectedBranchId
    }
    const res = await apiClient.get('/admin/lps/daily', { params })
    lpsData.value = res.data
    emit('lps-loaded', { 
      date: selectedDate.value, 
      employee_id: selectedEmployeeId.value,
      data: lpsData.value
    })
  } catch (err: any) {
    emit('lps-cleared')
    if (err?.response?.status !== 404) {
      toast.error(err?.response?.data?.error || 'Gagal memuat data LPS')
    }
  } finally {
    loading.value = false
  }
}

async function approveLPS() {
  if (!lpsData.value || !selectedDate.value || !selectedEmployeeId.value) return
  if (!isCashMatched.value && cashDifference.value !== 0) {
    if (!confirm(`Terdapat selisih kas sebesar ${formatCurrency(cashDifference.value)}. Yakin ingin melanjutkan validasi?`)) {
      return
    }
  }

  approving.value = true
  try {
    await apiClient.post('/admin/lps/daily/approve', {
      date: selectedDate.value,
      employee_id: selectedEmployeeId.value,
    })
    toast.success('LPS berhasil divalidasi! Semua transaksi telah disetujui secara otomatis.')
    emit('lps-approved')
    await fetchLPS()
  } catch (err: any) {
    toast.error('Gagal menyetujui LPS: ' + (err.response?.data?.error || err.message))
  } finally {
    approving.value = false
  }
}

const formatCurrency = (val: number) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val || 0)

onMounted(() => {
  fetchEmployees()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Filters -->
    <div class="bg-white p-5 rounded-3xl border border-slate-200/60 shadow-sm flex flex-wrap gap-4 items-end">
      <!-- Date picker -->
      <div class="flex-1 min-w-[180px]">
        <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1.5">Tanggal LPS</label>
        <div class="relative">
          <Calendar class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="selectedDate"
            type="date"
            @change="fetchLPS"
            class="w-full pl-9 pr-4 py-2.5 rounded-xl border border-slate-200 focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none text-sm font-medium transition-all"
          />
        </div>
      </div>
      
      <!-- Salesman Select -->
      <div class="flex-1 min-w-[200px]">
        <label class="block text-xs font-bold text-slate-500 uppercase tracking-wider mb-1.5">Salesman</label>
        <div class="relative">
          <User class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <select
            v-model="selectedEmployeeId"
            @change="fetchLPS"
            class="w-full pl-9 pr-4 py-2.5 rounded-xl border border-slate-200 focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none text-sm font-medium transition-all appearance-none bg-white"
          >
            <option value="" disabled>Pilih Salesman</option>
            <option v-for="emp in employees" :key="emp.id" :value="emp.id">{{ emp.first_name }} {{ emp.last_name }}</option>
          </select>
        </div>
      </div>

      <button
        @click="fetchLPS"
        :disabled="loading || !selectedDate || !selectedEmployeeId"
        class="flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl bg-slate-900 hover:bg-slate-800 text-white text-sm font-bold transition-all shadow-md disabled:opacity-50"
      >
        <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
        Cari Data
      </button>
    </div>

    <!-- Empty State -->
    <div v-if="!lpsData && !loading && selectedEmployeeId" class="bg-white rounded-3xl p-12 text-center border border-slate-200 border-dashed">
      <FileText class="w-12 h-12 text-slate-300 mx-auto mb-4" />
      <h3 class="text-lg font-bold text-slate-700">Data LPS Tidak Ditemukan</h3>
      <p class="text-slate-500 mt-2 text-sm max-w-md mx-auto">Tidak ada data LPS untuk salesman dan tanggal yang dipilih, atau dokumen belum ditutup buku oleh pihak admin.</p>
    </div>

    <!-- LPS Validation Card -->
    <div v-if="lpsData" class="grid grid-cols-1 gap-6">
      
      <!-- Data dari LPS -->
      <div class="bg-white rounded-[2rem] border border-slate-200 shadow-sm overflow-hidden flex flex-col">
        <div class="bg-slate-50 p-6 border-b border-slate-200 flex items-center gap-4">
          <div class="w-12 h-12 bg-white rounded-xl shadow-sm flex items-center justify-center text-primary">
            <FileText class="w-6 h-6" />
          </div>
          <div>
            <h3 class="text-lg font-black text-slate-900">Rangkuman LPS</h3>
            <p class="text-sm font-medium text-slate-500">Status: <span class="text-indigo-600 font-bold uppercase">{{ lpsData.status }}</span></p>
          </div>
        </div>
        <div class="p-6 flex-1 space-y-4">
          <div class="flex justify-between items-center p-4 bg-slate-50 rounded-2xl border border-slate-100">
            <span class="text-slate-600 font-semibold">Total Tunai (CASH)</span>
            <span class="text-2xl font-black text-slate-900">{{ formatCurrency(lpsData.total_cash) }}</span>
          </div>
          <div class="flex justify-between items-center p-4 bg-slate-50 rounded-2xl border border-slate-100">
            <span class="text-slate-600 font-semibold">Total Transfer/QRIS</span>
            <span class="text-xl font-bold text-slate-700">{{ formatCurrency(lpsData.total_transfer) }}</span>
          </div>
          <div class="flex justify-between items-center p-4 bg-slate-50 rounded-2xl border border-slate-100">
            <span class="text-slate-600 font-semibold">Total Piutang (TEMPO)</span>
            <span class="text-xl font-bold text-slate-700">{{ formatCurrency(lpsData.total_tempo) }}</span>
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
              :disabled="lpsData.status === 'APPROVED'"
            />
          </div>

          <div v-if="typeof physicalCash === 'number'">
            <div v-if="cashDifference === 0" class="flex items-center gap-3 p-4 bg-emerald-500/20 border border-emerald-500/30 rounded-2xl mb-6">
              <CheckCircle2 class="w-6 h-6 text-emerald-400" />
              <div>
                <p class="text-emerald-400 font-bold text-sm">MATCHED! Jumlah uang pas.</p>
                <p class="text-emerald-200/70 text-xs">Setoran fisik sesuai dengan LPS.</p>
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

        <button
          v-if="lpsData.status !== 'APPROVED'"
          @click="approveLPS"
          :disabled="approving || physicalCash === null"
          class="relative z-10 w-full py-4 rounded-2xl font-black text-lg flex items-center justify-center gap-3 transition-all shadow-lg active:scale-[0.98]"
          :class="isCashMatched ? 'bg-emerald-500 hover:bg-emerald-400 text-slate-900 shadow-emerald-500/20' : 'bg-white/10 hover:bg-white/20 text-white backdrop-blur-md'"
        >
          <span v-if="approving" class="w-5 h-5 rounded-full border-2 border-current border-t-transparent animate-spin"></span>
          <template v-else>
            <CheckCircle2 class="w-6 h-6" />
            TERIMA SETORAN & TUTUP BUKU
          </template>
        </button>
        
        <div v-else class="relative z-10 w-full py-4 rounded-2xl font-black text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center gap-3">
          <CheckCircle2 class="w-6 h-6" />
          SUDAH DIVALIDASI (APPROVED)
        </div>

        <button
          @click="emit('preview-lps', { date: selectedDate, salesman_id: selectedEmployeeId })"
          class="relative z-10 w-full mt-3 py-4 rounded-2xl font-black text-sm flex items-center justify-center gap-3 transition-all active:scale-[0.98] bg-slate-800 hover:bg-slate-700 text-white shadow-lg border border-slate-700"
        >
          <Printer class="w-5 h-5" />
          LIHAT & CETAK DOKUMEN LPS
        </button>
      </div>
    </div>
  </div>
</template>
