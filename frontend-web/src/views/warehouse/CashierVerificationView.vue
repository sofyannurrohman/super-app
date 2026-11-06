<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import {
  Search,
  ScanBarcode,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Banknote,
  Calendar,
  User,
  Store,
  Receipt,
  ArrowRight,
  Filter,
  History,
  Building2,
} from 'lucide-vue-next'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import axios from 'axios'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

const searchReceipt = ref('')
const loading = ref(false)
const transaction = ref<any>(null)
const recentTransactions = ref<any[]>([])

const buildBranchParams = () => {
  const branchId = masterDataStore.selectedBranchId
  if (branchId && branchId !== 'ALL_BRANCHES') {
    return { branch_id: branchId }
  }
  return {}
}

const fetchRecent = async () => {
  try {
    const response = await axios.get('/sales/transactions/history', {
      params: buildBranchParams(),
    })
    recentTransactions.value = (response.data.data || []).slice(0, 10)
  } catch (error) {
    console.error('Failed to fetch recent transactions', error)
  }
}

const handleSearch = async () => {
  if (!searchReceipt.value) return
  loading.value = true
  transaction.value = null

  try {
    const response = await axios.get(`/sales/transactions/history`, {
      params: buildBranchParams(),
    })
    const found = (response.data.data || []).find((t: any) => t.receipt_no === searchReceipt.value)

    if (found) {
      transaction.value = found
    } else {
      alert('Transaksi tidak ditemukan!')
    }
  } catch (error) {
    console.error('Search failed', error)
  } finally {
    loading.value = false
  }
}

const handleVerify = async (id: string) => {
  if (!confirm('Konfirmasi pembayaran dan validasi nota?')) return
  try {
    // 1. Record the payment
    await apiClient.post(`/sales/transactions/${id}/payments`, {
      amount: transaction.value.total_amount - (transaction.value.paid_amount || 0),
      payment_method: 'CASH',
      notes: 'Konfirmasi Tunai oleh Kasir',
    })

    // 2. Verify the transaction status
    await apiClient.patch(`/sales/transactions/${id}/verify`, {
      receipt_no: transaction.value.receipt_no,
      total_amount: transaction.value.total_amount,
      notes: 'Verified by Cashier',
    })

    toast.success('Transaksi berhasil divalidasi dan dikonfirmasi tunai')
    transaction.value.status = 'VERIFIED'
    transaction.value.payment_status = 'paid'
    fetchRecent()
  } catch (error) {
    toast.error('Gagal memvalidasi transaksi')
    console.error(error)
  }
}

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount)
}

const formatDate = (dateStr: any, full = true) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return new Intl.DateTimeFormat('id-ID', {
    weekday: full ? 'long' : undefined,
    day: 'numeric',
    month: full ? 'long' : 'short',
    year: 'numeric',
    hour: full ? '2-digit' : undefined,
    minute: full ? '2-digit' : undefined,
  }).format(date)
}

type BadgeVariant =
  | 'default'
  | 'destructive'
  | 'outline'
  | 'secondary'
  | 'success'
  | 'warning'
  | null
  | undefined

const getStatusBadge = (status: string): { variant: BadgeVariant; class: string; text: string } => {
  if (!status) return { variant: 'outline', class: '', text: '-' }
  switch (status.toLowerCase()) {
    case 'paid':
      return { variant: 'default', class: 'bg-emerald-500 hover:bg-emerald-600', text: 'TERBAYAR' }
    case 'pending':
      return {
        variant: 'secondary',
        class: 'bg-amber-500 text-white hover:bg-amber-600',
        text: 'MENUNGGU',
      }
    case 'expired':
      return { variant: 'destructive', class: '', text: 'KADALUWARSA' }
    default:
      return { variant: 'outline', class: '', text: status.toUpperCase() }
  }
}

// Watch branch changes and re-fetch data
watch(
  () => masterDataStore.selectedBranchId,
  () => {
    transaction.value = null
    searchReceipt.value = ''
    fetchRecent()
  },
)

onMounted(() => {
  fetchRecent()
})
</script>

<template>
  <div class="p-8 space-y-8 max-w-7xl mx-auto">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-black text-slate-900 tracking-tight">Verifikasi Kasir</h1>
        <p class="text-slate-500 font-medium">
          Validasi nota digital dan rekonsiliasi pembayaran sales.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <div class="bg-white p-1 rounded-2xl border border-slate-200 shadow-sm flex items-center">
          <Badge variant="outline" class="border-none text-[10px] font-bold px-3">HARI INI</Badge>
          <div class="h-4 w-[1px] bg-slate-200 mx-1"></div>
          <span class="text-xs font-bold text-slate-700 px-3">{{
            formatDate(new Date().toISOString())
          }}</span>
        </div>
        <!-- Branch Filter Indicator -->
        <div
          v-if="masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES'"
          class="flex items-center gap-2 bg-primary/5 border border-primary/20 px-4 py-2 rounded-2xl"
        >
          <Building2 class="w-4 h-4 text-primary shrink-0" />
          <span class="text-xs font-bold text-primary">
            {{
              masterDataStore.branches.find((b) => b.id === masterDataStore.selectedBranchId)
                ?.name ?? 'Cabang Terpilih'
            }}
          </span>
        </div>
        <div v-else class="flex items-center gap-2 bg-slate-50 border border-slate-200 px-4 py-2 rounded-2xl">
          <Building2 class="w-4 h-4 text-slate-400 shrink-0" />
          <span class="text-xs font-bold text-slate-500">Semua Cabang</span>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left: Search & Verification -->
      <div class="lg:col-span-2 space-y-8">
        <Card
          class="border-none shadow-xl shadow-slate-200/60 overflow-hidden ring-1 ring-slate-200/60"
        >
          <CardHeader class="bg-slate-50/50 pb-8">
            <div class="flex items-center gap-3 mb-2">
              <div class="p-2 bg-primary/10 rounded-xl">
                <ScanBarcode class="w-5 h-5 text-primary" />
              </div>
              <CardTitle class="text-xl font-bold">Cari Nota Digital</CardTitle>
            </div>
            <CardDescription
              >Masukkan atau scan nomor nota (INV/...) untuk memverifikasi
              transaksi.</CardDescription
            >

            <div class="mt-6 flex gap-3">
              <div class="relative flex-1">
                <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                <Input
                  v-model="searchReceipt"
                  @keyup.enter="handleSearch"
                  placeholder="Scan barcode nota di sini..."
                  class="pl-12 py-7 rounded-2xl border-slate-200 focus:ring-primary/20 text-lg font-bold tracking-wider"
                />
              </div>
              <Button
                @click="handleSearch"
                class="h-auto px-8 rounded-2xl font-bold gap-2"
                size="lg"
              >
                <ArrowRight class="w-5 h-5" />
                Cek Nota
              </Button>
            </div>
          </CardHeader>

          <!-- Result Area -->
          <CardContent
            v-if="transaction"
            class="p-8 animate-in fade-in slide-in-from-bottom-4 duration-500"
          >
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
              <div class="space-y-6">
                <div class="p-6 bg-slate-50 rounded-3xl space-y-4">
                  <div class="flex items-center justify-between">
                    <span class="text-[10px] font-black uppercase tracking-widest text-slate-400"
                      >Status Pembayaran</span
                    >
                    <Badge
                      :variant="getStatusBadge(transaction.payment_status).variant"
                      :class="`px-4 py-1.5 rounded-full font-black text-[10px] ${getStatusBadge(transaction.payment_status).class}`"
                    >
                      {{ getStatusBadge(transaction.payment_status).text }}
                    </Badge>
                  </div>
                  <div class="flex items-center gap-4">
                    <div class="p-3 bg-white rounded-2xl shadow-sm border border-slate-200">
                      <Banknote class="w-8 h-8 text-emerald-500" />
                    </div>
                    <div>
                      <p class="text-xs font-bold text-slate-400 uppercase tracking-tight">
                        Total Tagihan
                      </p>
                      <h2 class="text-3xl font-black text-slate-900">
                        {{ formatCurrency(transaction.total_amount) }}
                      </h2>
                    </div>
                  </div>
                </div>

                <div class="space-y-4 px-2">
                  <div class="flex items-center gap-3">
                    <Store class="w-5 h-5 text-primary" />
                    <div class="flex-1">
                      <p class="text-[10px] font-bold text-slate-400 uppercase">Toko / Customer</p>
                      <p class="font-bold text-slate-900">{{ transaction.store?.name }}</p>
                    </div>
                  </div>
                  <div class="flex items-center gap-3">
                    <User class="w-5 h-5 text-indigo-500" />
                    <div class="flex-1">
                      <p class="text-[10px] font-bold text-slate-400 uppercase">Salesman</p>
                      <p class="font-bold text-slate-900">{{ transaction.salesman_name }}</p>
                    </div>
                  </div>
                  <div class="flex items-center gap-3">
                    <Calendar class="w-5 h-5 text-amber-500" />
                    <div class="flex-1">
                      <p
                        class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1"
                      >
                        Waktu Transaksi
                      </p>
                      <p class="font-bold text-slate-900">
                        {{
                          formatDate(
                            transaction.visit?.check_in_time ||
                              transaction.created_at ||
                              transaction.transaction_date,
                          )
                        }}
                      </p>
                    </div>
                  </div>
                </div>

                <!-- Bukti Pembayaran Resmi Modern Trade (MT) -->
                <div
                  v-if="transaction.payment_proof_url || transaction.is_mt"
                  class="p-6 bg-violet-50/50 border border-violet-100 rounded-3xl space-y-3"
                >
                  <div class="flex items-center justify-between">
                    <span class="text-[10px] font-black uppercase tracking-widest text-violet-600"
                      >Bukti Pembayaran Terstempel (MT)</span
                    >
                    <Badge variant="outline" class="border-violet-200 text-violet-700 bg-violet-100/50 text-[9px] font-black uppercase tracking-wider"
                      >Wajib Diperiksa</Badge
                    >
                  </div>
                  <div v-if="transaction.payment_proof_url" class="relative group rounded-2xl overflow-hidden border border-slate-200 bg-white max-h-[220px] flex items-center justify-center">
                    <img
                      :src="transaction.payment_proof_url"
                      alt="Bukti Nota Pembayaran Stempel"
                      class="object-contain max-h-[200px] w-full p-2 group-hover:scale-105 transition-transform duration-300"
                    />
                    <a
                      :href="transaction.payment_proof_url"
                      target="_blank"
                      class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white text-xs font-bold gap-1 rounded-2xl"
                    >
                      🔍 Perbesar Gambar
                    </a>
                  </div>
                  <div v-else class="p-4 bg-amber-50 border border-amber-200 rounded-2xl text-amber-800 text-[11px] font-semibold flex items-center gap-2">
                    <AlertCircle class="w-4 h-4 text-amber-500 shrink-0" />
                    <span>Belum ada unggahan bukti pembayaran berstempel dari Sales MT.</span>
                  </div>
                </div>
              </div>

              <div class="space-y-6">
                <div class="border border-slate-100 rounded-3xl overflow-hidden shadow-sm">
                  <div
                    class="bg-slate-50 px-6 py-4 border-b border-slate-100 flex items-center justify-between"
                  >
                    <span class="text-xs font-black text-slate-500 uppercase tracking-wider"
                      >Item Pesanan</span
                    >
                    <Badge variant="outline" class="font-bold text-[10px]"
                      >{{ (transaction.items || []).length }} PRODUK</Badge
                    >
                  </div>
                  <div class="max-h-[300px] overflow-y-auto p-4 space-y-3">
                    <div
                      v-for="item in transaction.items"
                      :key="item.id"
                      class="flex items-center gap-3 p-3 bg-white rounded-xl border border-slate-100 shadow-sm"
                    >
                      <div
                        class="w-10 h-10 bg-slate-50 rounded-lg flex items-center justify-center font-bold text-xs text-primary"
                      >
                        {{ item.quantity }}x
                      </div>
                      <div class="flex-1 min-w-0">
                        <p class="font-bold text-sm text-slate-900 truncate">
                          {{ item.product_name }}
                        </p>
                        <p class="text-[10px] font-medium text-slate-500">
                          {{ formatCurrency(item.price) }} / item
                        </p>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="flex gap-3">
                  <Button
                    @click="handleVerify(transaction.id)"
                    v-if="
                      transaction.payment_method === 'CASH' &&
                      transaction.payment_status === 'pending'
                    "
                    class="flex-1 h-14 rounded-2xl font-bold bg-emerald-500 hover:bg-emerald-600 gap-2 shadow-lg shadow-emerald-500/20"
                  >
                    <CheckCircle2 class="w-5 h-5" />
                    Konfirmasi Tunai
                  </Button>
                  <Button
                    variant="outline"
                    class="flex-1 h-14 rounded-2xl font-bold border-slate-200 gap-2"
                  >
                    <Receipt class="w-5 h-5" />
                    Cetak Ulang Nota
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>

          <div v-else-if="!loading" class="p-20 text-center space-y-4">
            <div
              class="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center mx-auto border border-slate-100"
            >
              <Receipt class="w-10 h-10 text-slate-300" />
            </div>
            <div class="max-w-xs mx-auto">
              <p class="font-bold text-slate-900">Belum ada nota yang dimuat</p>
              <p class="text-sm text-slate-500 font-medium">
                Scan barcode pada nota digital pelanggan atau sales untuk melihat rincian.
              </p>
            </div>
          </div>

          <div v-if="loading" class="p-20 text-center">
            <div
              class="w-12 h-12 border-4 border-primary/30 border-t-primary rounded-full animate-spin mx-auto"
            ></div>
            <p class="mt-4 text-sm font-bold text-slate-500">Mencari data transaksi...</p>
          </div>
        </Card>
      </div>

      <!-- Right: Recent Activity -->
      <div class="space-y-8">
        <Card
          class="border-none shadow-xl shadow-slate-200/60 ring-1 ring-slate-200/60 overflow-hidden h-full"
        >
          <CardHeader class="pb-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <History class="w-4 h-4 text-slate-400" />
                <CardTitle class="text-sm font-black uppercase tracking-widest text-slate-500"
                  >Aktifitas Terbaru</CardTitle
                >
              </div>
              <Button
                variant="ghost"
                size="sm"
                class="h-7 text-[10px] font-black uppercase tracking-wider text-primary"
                >Lihat Semua</Button
              >
            </div>
          </CardHeader>
          <CardContent class="p-0">
            <div class="divide-y divide-slate-100">
              <div
                v-for="item in recentTransactions"
                :key="item.id"
                class="p-5 hover:bg-slate-50 transition-colors cursor-pointer group"
              >
                <div class="flex justify-between items-start mb-2">
                  <div class="space-y-0.5">
                    <h4 class="font-black text-slate-900">{{ item.receipt_no }}</h4>
                    <p class="text-[10px] font-bold text-slate-400">
                      {{
                        formatDate(
                          item.visit?.check_in_time || item.created_at || item.transaction_date,
                          false,
                        )
                      }}
                      • {{ item.store?.name }}
                    </p>
                  </div>
                  <Badge
                    :variant="getStatusBadge(item.payment_status).variant"
                    :class="`text-[8px] font-black px-2 py-0.5 ${getStatusBadge(item.payment_status).class}`"
                  >
                    {{ getStatusBadge(item.payment_status).text }}
                  </Badge>
                </div>
                <p class="text-sm font-black text-slate-900">
                  {{ formatCurrency(item.total_amount) }}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-in {
  animation-duration: 0.5s;
}
</style>
