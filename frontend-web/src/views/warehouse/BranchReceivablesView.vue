<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-2">
          <span>Tagihan Pabrik (Piutang Cabang)</span>
          <span class="text-xs font-bold bg-indigo-100 text-indigo-800 px-2.5 py-0.5 rounded-full">
            Inter-Branch Ledger
          </span>
        </h2>
        <p class="text-sm text-slate-500 mt-1">
          Pantau tagihan pengiriman stok barang dari pabrik, potongan retur barang, dan status setoran pembayaran cabang.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <Button
          variant="outline"
          class="rounded-xl border-slate-300 font-bold text-xs"
          @click="fetchData"
          :disabled="isLoading"
        >
          <RefreshCw :class="['w-4 h-4 mr-1.5', isLoading ? 'animate-spin' : '']" />
          Segarkan Data
        </Button>
      </div>
    </div>

    <!-- Summary Metric Cards -->
    <!-- Summary Metric Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
      <Card class="rounded-2xl border-slate-200/80 shadow-sm">
        <CardHeader class="pb-2">
          <CardTitle class="text-xs font-bold text-slate-400 uppercase tracking-wider">
            Total Faktur Terbit (Bruto)
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-black text-slate-900">
            {{ formatRupiah(summaryData.total_good_invoiced || summaryData.total_good_receivables) }}
          </div>
          <p class="text-[11px] text-slate-400 mt-1">Akumulasi DO / Surat Jalan aktif</p>
        </CardContent>
      </Card>

      <Card class="rounded-2xl border-slate-200/80 shadow-sm">
        <CardHeader class="pb-2">
          <CardTitle class="text-xs font-bold text-rose-500 uppercase tracking-wider flex items-center gap-1">
            <RotateCcw class="w-3.5 h-3.5" />
            Potongan Retur & Rusak
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-black text-rose-600">
            - {{ formatRupiah((summaryData.total_damaged_deductions || 0) + (summaryData.total_return_deductions || 0)) }}
          </div>
          <p class="text-[11px] text-slate-400 mt-1">Koreksi saat penerimaan & retur QC</p>
        </CardContent>
      </Card>
      
      <Card class="rounded-2xl border-slate-200/80 shadow-sm">
        <CardHeader class="pb-2">
          <CardTitle class="text-xs font-bold text-slate-400 uppercase tracking-wider">
            Total Piutang Finansial
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-black text-amber-600">
            {{ formatRupiah(summaryData.total_financial_receivables) }}
          </div>
          <p class="text-[11px] text-slate-400 mt-1">Pinjaman & biaya operasional cabang</p>
        </CardContent>
      </Card>

      <Card class="rounded-2xl border-slate-200/80 shadow-sm">
        <CardHeader class="pb-2">
          <CardTitle class="text-xs font-bold text-slate-400 uppercase tracking-wider">
            Setoran Masuk (Bulan Ini)
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-black text-emerald-600">
            {{ formatRupiah(summaryData.total_payments_this_month) }}
          </div>
          <p class="text-[11px] text-slate-400 mt-1">Pembayaran / setoran omzet bulan ini</p>
        </CardContent>
      </Card>

      <Card class="rounded-2xl border-indigo-200 bg-gradient-to-br from-slate-900 to-indigo-950 text-white shadow-md">
        <CardHeader class="pb-2">
          <CardTitle class="text-xs font-bold text-indigo-300 uppercase tracking-wider">
            Total Sisa Piutang Cabang
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-black text-emerald-400">
            {{ formatRupiah(summaryData.total_receivables) }}
          </div>
          <p class="text-[11px] text-indigo-200 mt-1">
            {{ summaryData.unpaid_branches_count }} cabang memiliki kewajiban
          </p>
        </CardContent>
      </Card>
    </div>

    <!-- Table of Receivables by Branch -->
    <Card class="rounded-2xl border-slate-200/80 shadow-sm overflow-hidden">
      <CardHeader class="bg-slate-50 border-b border-slate-200 py-4 px-6 flex flex-row items-center justify-between">
        <div>
          <CardTitle class="text-base font-black text-slate-900">
            Rincian Piutang & Setoran per Cabang
          </CardTitle>
          <p class="text-xs text-slate-500 mt-0.5">
            Daftar saldo piutang barang (dikurangi potongan retur/rusak) dan setoran untuk masing-masing cabang Karsa HR.
          </p>
        </div>
      </CardHeader>
      <CardContent class="p-0">
        <div v-if="isLoading" class="p-12 text-center text-slate-400 font-medium text-xs">
          <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-indigo-500" />
          Memuat data tagihan dan piutang cabang...
        </div>

        <div v-else-if="branches.length === 0" class="p-12 text-center text-slate-400 text-xs">
          Tidak ada data tagihan cabang.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse">
            <thead class="bg-slate-100 text-slate-700 uppercase font-black text-[11px] border-b border-slate-200">
              <tr>
                <th class="p-3.5 pl-6">Nama Cabang</th>
                <th class="p-3.5 text-right">Total Faktur Bruto</th>
                <th class="p-3.5 text-right text-rose-600">Potongan Retur & Rusak</th>
                <th class="p-3.5 text-right text-indigo-900">Piutang Barang (Net)</th>
                <th class="p-3.5 text-right">Piutang Finansial</th>
                <th class="p-3.5 text-right">Total Piutang Cabang</th>
                <th class="p-3.5 text-right">Setoran Bulan Ini</th>
                <th class="p-3.5 pr-6 text-center">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 font-medium">
              <tr v-for="b in branches" :key="b.branch_id" class="hover:bg-slate-50/80 transition-colors">
                <td class="p-3.5 pl-6">
                  <div class="font-extrabold text-slate-900 text-sm flex items-center gap-2">
                    <Building2 class="w-4 h-4 text-indigo-600" />
                    {{ b.branch_name }}
                  </div>
                </td>
                <td class="p-3.5 text-right font-bold text-slate-800">
                  {{ formatRupiah(b.total_good_invoiced || b.good_receivable) }}
                </td>
                <td class="p-3.5 text-right font-bold text-rose-600">
                  <span v-if="(b.total_damaged_deduction || 0) + (b.total_return_deduction || 0) > 0">
                    - {{ formatRupiah((b.total_damaged_deduction || 0) + (b.total_return_deduction || 0)) }}
                  </span>
                  <span v-else class="text-slate-400">Rp 0</span>
                </td>
                <td class="p-3.5 text-right font-black text-indigo-900">
                  {{ formatRupiah(b.good_receivable) }}
                </td>
                <td class="p-3.5 text-right font-bold text-slate-800">
                  {{ formatRupiah(b.financial_receivable) }}
                </td>
                <td class="p-3.5 text-right font-black text-sm" :class="b.total_receivable > 0 ? 'text-indigo-950' : 'text-slate-400'">
                  {{ formatRupiah(b.total_receivable) }}
                </td>
                <td class="p-3.5 text-right font-bold text-emerald-600">
                  {{ formatRupiah(b.total_payment_this_month) }}
                </td>
                <td class="p-3.5 pr-6 text-center">
                  <span
                    v-if="b.total_receivable <= 0"
                    class="px-2.5 py-1 bg-emerald-50 text-emerald-700 border border-emerald-200 rounded-full font-bold text-[10px]"
                  >
                    LUNAS / NIHIL
                  </span>
                  <span
                    v-else
                    class="px-2.5 py-1 bg-amber-50 text-amber-700 border border-amber-200 rounded-full font-bold text-[10px]"
                  >
                    BELUM LUNAS
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import apiClient from '@/api/axios'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { RefreshCw, Building2, Truck, CheckCircle2, RotateCcw } from 'lucide-vue-next'

const isLoading = ref(false)

const summaryData = ref({
  total_receivables: 0,
  total_good_receivables: 0,
  total_good_invoiced: 0,
  total_damaged_deductions: 0,
  total_return_deductions: 0,
  total_financial_receivables: 0,
  total_payments_this_month: 0,
  unpaid_branches_count: 0,
})

const branches = ref<any[]>([])

const formatRupiah = (val?: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const fetchData = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get('/factory/receivables')
    if (res.data) {
      summaryData.value = res.data.summary || summaryData.value
      branches.value = res.data.branches || []
    }
  } catch (error) {
    console.error('Failed to load branch receivables:', error)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>
