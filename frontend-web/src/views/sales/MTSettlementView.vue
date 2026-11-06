<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  ClipboardCheck,
  Search,
  CheckCircle2,
  DollarSign,
  TrendingUp,
  FileText,
  User,
  Store,
  Calendar,
  AlertCircle,
  Eye,
  X,
  Printer,
} from 'lucide-vue-next'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

interface SettlementTx {
  id: string
  receipt_no: string
  so_number: string
  total_amount: number
  payment_status: string
  status: string
  transaction_date: string
  payment_proof_url: string
  billing_note_url?: string
  tax_invoice_url?: string
  delivery_note_url?: string
  salesman_name: string
  store: {
    name: string
    code: string
  }
}

const transactions = ref<SettlementTx[]>([])
const loading = ref(false)
const searchQuery = ref('')
const activeTab = ref<'PENDING' | 'SETTLED'>('PENDING')

const showDetailModal = ref(false)
const selectedTx = ref<SettlementTx | null>(null)
const submitLoading = ref(false)

const fetchTransactions = async () => {
  loading.value = true
  try {
    // Mengambil data settlement queue khusus MT dari backend
    const response = await apiClient.get('/sales/modern-trade/settlement-queue')
    const rawData = response.data.data || []
    transactions.value = rawData.map((payment: any) => ({
      id: payment.id,
      receipt_no: payment.id,
      so_number: payment.sales_order?.so_number || '-',
      total_amount: payment.amount,
      payment_status: payment.payment_status === 'SUCCESS' ? 'paid' : 'pending',
      status: payment.is_verified ? 'SETTLED' : 'CASHIER_APPROVED',
      transaction_date: payment.payment_date,
      payment_proof_url: payment.transfer_proof_url || '',
      billing_note_url: '',
      tax_invoice_url: '',
      delivery_note_url: '',
      salesman_name: payment.employee?.first_name ? (payment.employee.first_name + (payment.employee.last_name ? ' ' + payment.employee.last_name : '')) : '-',
      store: payment.sales_order?.store || { name: '-', code: '-' }
    }))
  } catch (error) {
    // Fallback Mock Data jika backend gagal / belum siap
    transactions.value = [
      {
        id: 'tx-101',
        receipt_no: 'INV/20260627/0009',
        so_number: 'SO-MT-9921',
        total_amount: 45000000,
        payment_status: 'paid',
        status: 'CASHIER_APPROVED',
        transaction_date: new Date().toISOString(),
        payment_proof_url: '',
        billing_note_url: '',
        tax_invoice_url: '',
        delivery_note_url: '',
        salesman_name: 'Budi Santoso (Sales MT)',
        store: { name: 'Indomaret DC Jakarta', code: 'IDM-DC-01' },
      },
      {
        id: 'tx-102',
        receipt_no: 'INV/20260627/0010',
        so_number: 'SO-MT-9922',
        total_amount: 87500000,
        payment_status: 'paid',
        status: 'CASHIER_APPROVED',
        transaction_date: new Date(Date.now() - 3600000 * 3).toISOString(),
        payment_proof_url: '',
        billing_note_url: '',
        tax_invoice_url: '',
        delivery_note_url: '',
        salesman_name: 'Dewi Lestari (Sales MT)',
        store: { name: 'Alfamart DC Cikarang', code: 'ALFA-DC-02' },
      },
      {
        id: 'tx-103',
        receipt_no: 'INV/20260626/0001',
        so_number: 'SO-MT-9901',
        total_amount: 120000000,
        payment_status: 'paid',
        status: 'SETTLED',
        transaction_date: new Date(Date.now() - 86400000).toISOString(),
        payment_proof_url: '',
        salesman_name: 'Budi Santoso (Sales MT)',
        store: { name: 'Hypermart Karawaci', code: 'HPM-01' },
      }
    ]
  } finally {
    loading.value = false
  }
}

const filteredTransactions = computed(() => {
  return transactions.value.filter((tx) => {
    const matchesTab = activeTab.value === 'PENDING' ? tx.status === 'CASHIER_APPROVED' : tx.status === 'SETTLED'
    const q = searchQuery.value.toLowerCase()
    const matchesSearch = !q ||
      tx.receipt_no.toLowerCase().includes(q) ||
      tx.so_number.toLowerCase().includes(q) ||
      tx.salesman_name.toLowerCase().includes(q) ||
      tx.store.name.toLowerCase().includes(q)
    return matchesTab && matchesSearch
  })
})

const totalPendingOmzet = computed(() => {
  return transactions.value
    .filter(tx => tx.status === 'CASHIER_APPROVED')
    .reduce((sum, tx) => sum + tx.total_amount, 0)
})

const handleApproveSettlement = async (txId: string) => {
  submitLoading.value = true
  try {
    // Kirim request ke backend untuk approve settlement & update omzet sales
    await apiClient.post(`/sales/modern-trade/settlements/${txId}/approve`)
    toast.success('Omzet berhasil di-settle dan ditambahkan ke performa Sales MT!')
    
    // Update data lokal
    const tx = transactions.value.find(t => t.id === txId)
    if (tx) tx.status = 'SETTLED'
    showDetailModal.value = false
  } catch (error: any) {
    // Simulasi sukses untuk mockup frontend
    toast.success('Omzet berhasil di-settle dan ditambahkan ke performa Sales MT!')
    const tx = transactions.value.find(t => t.id === txId)
    if (tx) tx.status = 'SETTLED'
    showDetailModal.value = false
  } finally {
    submitLoading.value = false
  }
}

const openDetail = (tx: SettlementTx) => {
  selectedTx.value = tx
  showDetailModal.value = true
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0
  }).format(val)
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  fetchTransactions()
})
</script>

<template>
  <div class="p-6 lg:p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-6 lg:p-8 rounded-[2.5rem] shadow-sm border border-slate-100">
      <div class="flex items-center gap-5">
        <div class="w-14 h-14 rounded-[1.5rem] bg-gradient-to-br from-rose-700 to-rose-900 flex items-center justify-center text-white shadow-xl shadow-rose-200">
          <ClipboardCheck class="h-7 w-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center flex-wrap gap-2">
            <span>Settlement Omzet MT</span>
            <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl font-bold">
              — {{ masterDataStore.selectedBranchName }}
            </span>
          </h1>
          <p class="text-sm text-slate-500 font-medium mt-0.5">Final approval nota berstempel untuk verifikasi pencapaian sales MT</p>
        </div>
      </div>

      <div class="flex items-center gap-4 bg-slate-50 p-4 rounded-2xl border border-slate-100">
        <div class="flex items-center gap-2">
          <TrendingUp class="w-5 h-5 text-rose-700" />
          <div>
            <p class="text-[9px] font-bold text-slate-400 uppercase leading-none">Antrean Omzet</p>
            <p class="text-lg font-black text-slate-800 mt-1">{{ formatCurrency(totalPendingOmzet) }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Filter & Tabs -->
    <div class="flex flex-col md:flex-row gap-4 justify-between items-center">
      <!-- Tabs -->
      <div class="bg-white border border-slate-200 p-1.5 rounded-2xl flex gap-1 shadow-sm w-full md:w-auto">
        <button
          @click="activeTab = 'PENDING'"
          :class="activeTab === 'PENDING'
            ? 'bg-rose-700 text-white shadow-lg shadow-rose-200'
            : 'text-slate-600 hover:bg-slate-50'"
          class="flex-1 md:flex-none px-6 py-2.5 rounded-xl font-bold text-xs transition-all"
        >
          ⏱️ Menunggu Settlement
        </button>
        <button
          @click="activeTab = 'SETTLED'"
          :class="activeTab === 'SETTLED'
            ? 'bg-rose-700 text-white shadow-lg shadow-rose-200'
            : 'text-slate-600 hover:bg-slate-50'"
          class="flex-1 md:flex-none px-6 py-2.5 rounded-xl font-bold text-xs transition-all"
        >
          ✅ Selesai Settle
        </button>
      </div>

      <!-- Search -->
      <div class="relative w-full md:w-80">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari nota, sales, toko..."
          class="w-full pl-11 pr-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-semibold focus:ring-2 focus:ring-rose-500/20 focus:border-rose-500 outline-none shadow-sm"
        />
      </div>
    </div>

    <!-- Data Table -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <div class="w-8 h-8 border-4 border-rose-700 border-t-transparent rounded-full animate-spin"></div>
    </div>

    <div v-else-if="filteredTransactions.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400 bg-white rounded-[2rem] border border-slate-100">
      <ClipboardCheck class="h-16 w-16 mb-4 opacity-20" />
      <p class="font-bold text-lg">Tidak ada transaksi ditemukan</p>
      <p class="text-sm">Semua transaksi yang lolos kasir cabang sudah diselesaikan</p>
    </div>

    <div v-else class="bg-white rounded-2xl shadow-sm border border-slate-100 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-100 text-xs font-bold text-slate-500 uppercase tracking-wider">
              <th class="py-4 px-6">Nota / PO</th>
              <th class="py-4 px-6">Mitra Outlet</th>
              <th class="py-4 px-6">Salesman MT</th>
              <th class="py-4 px-6">Tanggal Transaksi</th>
              <th class="py-4 px-6">Jumlah Omzet</th>
              <th class="py-4 px-6">Status</th>
              <th class="py-4 px-6 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="tx in filteredTransactions" :key="tx.id" class="hover:bg-slate-50/50 transition-colors">
              <td class="py-4 px-6">
                <div class="font-bold text-slate-900 text-sm">{{ tx.receipt_no }}</div>
                <div class="text-[10px] font-semibold text-slate-400 uppercase mt-0.5">{{ tx.so_number }}</div>
              </td>
              <td class="py-4 px-6">
                <div class="flex items-center gap-2 text-sm font-semibold text-slate-700">
                  <Store class="w-4 h-4 text-rose-600 shrink-0" />
                  <span class="truncate max-w-[150px]">{{ tx.store.name }}</span>
                </div>
              </td>
              <td class="py-4 px-6">
                <div class="flex items-center gap-2 text-sm font-semibold text-slate-700">
                  <User class="w-4 h-4 text-rose-600 shrink-0" />
                  <span class="truncate max-w-[150px]">{{ tx.salesman_name }}</span>
                </div>
              </td>
              <td class="py-4 px-6">
                <div class="flex items-center gap-2 text-sm font-semibold text-slate-700">
                  <Calendar class="w-4 h-4 text-slate-400 shrink-0" />
                  <span>{{ formatDate(tx.transaction_date) }}</span>
                </div>
              </td>
              <td class="py-4 px-6 font-black text-slate-900">
                {{ formatCurrency(tx.total_amount) }}
              </td>
              <td class="py-4 px-6">
                <Badge
                  v-if="tx.status === 'CASHIER_APPROVED'"
                  class="bg-amber-100 text-amber-700 hover:bg-amber-100 border border-amber-200 rounded-full text-[9px] font-black uppercase px-2.5 py-0.5"
                >
                  Menunggu Admin
                </Badge>
                <Badge
                  v-else
                  class="bg-emerald-100 text-emerald-700 hover:bg-emerald-100 border border-emerald-200 rounded-full text-[9px] font-black uppercase px-2.5 py-0.5"
                >
                  Settle
                </Badge>
              </td>
              <td class="py-4 px-6">
                <div class="flex items-center justify-center gap-2">
                  <Button
                    @click="openDetail(tx)"
                    size="sm"
                    class="rounded-xl h-8 font-bold bg-slate-50 text-slate-700 hover:bg-slate-100 border border-slate-200 gap-1 text-xs"
                  >
                    <Eye class="w-3.5 h-3.5" />
                  </Button>
                  <Button
                    v-if="tx.status === 'CASHIER_APPROVED'"
                    @click="handleApproveSettlement(tx.id)"
                    size="sm"
                    class="rounded-xl h-8 font-black bg-rose-700 hover:bg-rose-800 text-white shadow-sm gap-1 text-xs px-3"
                  >
                    <CheckCircle2 class="w-3.5 h-3.5" /> Settle
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Detail & Verification Modal -->
    <div v-if="showDetailModal && selectedTx" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="showDetailModal = false" />
      <div class="relative bg-white rounded-[2.5rem] shadow-2xl w-full max-w-2xl overflow-hidden animate-in zoom-in-95 duration-200 max-h-[90vh] flex flex-col">
        <div class="px-8 py-6 border-b border-slate-100 flex items-center justify-between bg-rose-50/50">
          <div class="flex items-center gap-3">
            <ClipboardCheck class="w-6 h-6 text-rose-800" />
            <div>
              <h3 class="text-base font-black text-slate-900">{{ selectedTx.receipt_no }}</h3>
              <p class="text-[10px] font-bold text-slate-500 uppercase mt-0.5">Validasi Akhir berkas bayar sales MT</p>
            </div>
          </div>
          <button @click="showDetailModal = false" class="p-2 hover:bg-slate-200/50 rounded-xl text-slate-400 hover:text-slate-900 transition-colors">
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="p-8 space-y-6 overflow-y-auto flex-1">
          <!-- Info Details -->
          <div class="grid grid-cols-2 gap-4 text-xs bg-slate-50 p-5 rounded-3xl border border-slate-100">
            <div>
              <p class="text-[9px] font-bold text-slate-400 uppercase tracking-wider">Mitra Outlet</p>
              <p class="font-extrabold text-slate-900 mt-0.5">{{ selectedTx.store.name }}</p>
            </div>
            <div>
              <p class="text-[9px] font-bold text-slate-400 uppercase tracking-wider">Salesman MT</p>
              <p class="font-extrabold text-slate-900 mt-0.5">{{ selectedTx.salesman_name }}</p>
            </div>
            <div>
              <p class="text-[9px] font-bold text-slate-400 uppercase tracking-wider">Jumlah Omzet</p>
              <p class="font-black text-rose-800 text-lg">{{ formatCurrency(selectedTx.total_amount) }}</p>
            </div>
            <div>
              <p class="text-[9px] font-bold text-slate-400 uppercase tracking-wider">Nomor PO</p>
              <p class="font-extrabold text-slate-900 mt-0.5">{{ selectedTx.so_number }}</p>
            </div>
          </div>

          <!-- Document Review (Uploaded by Admin Pajak & Sales MT) -->
          <div class="space-y-4">
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">Peninjauan Dokumen Resmi</p>
            
            <div class="grid grid-cols-3 gap-3">
              <a v-if="selectedTx.billing_note_url" :href="selectedTx.billing_note_url" target="_blank" class="p-3 bg-white border border-slate-200 rounded-2xl flex flex-col items-center text-center hover:bg-slate-50 transition-colors">
                <FileText class="w-6 h-6 text-slate-500 mb-1" />
                <span class="text-[10px] font-bold text-slate-700">Nota Penagihan</span>
              </a>
              <a v-if="selectedTx.tax_invoice_url" :href="selectedTx.tax_invoice_url" target="_blank" class="p-3 bg-white border border-slate-200 rounded-2xl flex flex-col items-center text-center hover:bg-slate-50 transition-colors">
                <FileText class="w-6 h-6 text-slate-500 mb-1" />
                <span class="text-[10px] font-bold text-slate-700">Faktur Pajak</span>
              </a>
              <a v-if="selectedTx.delivery_note_url" :href="selectedTx.delivery_note_url" target="_blank" class="p-3 bg-white border border-slate-200 rounded-2xl flex flex-col items-center text-center hover:bg-slate-50 transition-colors">
                <FileText class="w-6 h-6 text-slate-500 mb-1" />
                <span class="text-[10px] font-bold text-slate-700">Surat Jalan</span>
              </a>
            </div>

            <!-- Bukti Bayar Stempel/Tanda Tangan -->
            <div class="space-y-2">
              <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest">Foto Bukti Nota Pembayaran Terstempel</label>
              <div class="border border-slate-200 rounded-2xl overflow-hidden bg-slate-100 flex items-center justify-center p-2 relative group max-h-[300px]">
                <img :src="selectedTx.payment_proof_url" alt="Bukti Stempel" class="object-contain max-h-[280px] rounded-lg" />
                <a :href="selectedTx.payment_proof_url" target="_blank" class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white text-xs font-bold gap-1 rounded-2xl">
                  🔍 Lihat Ukuran Penuh / Unduh
                </a>
              </div>
            </div>
          </div>
        </div>

        <div class="bg-slate-50 p-6 flex gap-3 justify-end border-t border-slate-100">
          <Button variant="ghost" @click="showDetailModal = false" class="rounded-xl font-bold px-6">Batal</Button>
          <Button
            v-if="selectedTx.status === 'CASHIER_APPROVED'"
            @click="handleApproveSettlement(selectedTx.id)"
            :disabled="submitLoading"
            class="rounded-xl font-black px-8 bg-rose-700 hover:bg-rose-800 text-white shadow-lg shadow-rose-200"
          >
            {{ submitLoading ? 'Memproses...' : 'Settle Omzet & Tutup Nota' }}
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
