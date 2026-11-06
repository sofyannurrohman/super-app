<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-200">
      <div>
        <div class="flex items-center gap-2.5">
          <div class="p-2 bg-indigo-50 text-indigo-600 rounded-xl">
            <Building2 class="w-6 h-6" />
          </div>
          <div>
            <h2 class="text-2xl font-bold tracking-tight text-slate-900">Monitoring Piutang Cabang</h2>
            <p class="text-sm text-slate-500 mt-0.5">
              Pantau total tagihan barang (Surat Jalan/DO) & talangan kas operasional (Petty Cash/Kasbon Cabang) secara terpadu.
            </p>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <button
          @click="fetchData"
          class="px-3.5 py-2 text-sm font-semibold text-slate-700 bg-white border border-slate-300 rounded-xl hover:bg-slate-50 transition-all flex items-center gap-2 shadow-sm"
          :disabled="loading"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
          <span>Refresh</span>
        </button>
        <button
          @click="handlePrint"
          class="px-4 py-2 text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-700 rounded-xl transition-all flex items-center gap-2 shadow-sm shadow-indigo-200"
        >
          <Printer class="w-4 h-4" />
          <span>Cetak Laporan</span>
        </button>
      </div>
    </div>

    <div v-if="loading && branches.length === 0" class="flex flex-col items-center justify-center py-16 bg-white rounded-2xl border border-slate-200">
      <Loader2 class="w-10 h-10 animate-spin text-indigo-600 mb-3" />
      <p class="text-sm font-medium text-slate-500">Memuat data piutang cabang & keuangan...</p>
    </div>
    
    <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 p-4 rounded-xl flex items-center gap-3">
      <AlertCircle class="w-5 h-5 text-red-500 flex-shrink-0" />
      <span class="text-sm font-medium">{{ error }}</span>
    </div>

    <template v-else>
      <!-- KPI Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden">
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-500">Total Piutang Konsolidasi</span>
            <div class="p-2 bg-indigo-50 text-indigo-600 rounded-xl">
              <Layers class="w-5 h-5" />
            </div>
          </div>
          <div class="text-2xl font-black text-slate-900 mt-2">
            {{ formatCurrency(summary.total_receivables) }}
          </div>
          <p class="text-xs text-slate-500 mt-1">Gabungan piutang barang (DO) + talangan kas</p>
        </div>
        
        <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden">
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold uppercase tracking-wider text-blue-600">Piutang Barang (DO)</span>
            <div class="p-2 bg-blue-50 text-blue-600 rounded-xl">
              <Truck class="w-5 h-5" />
            </div>
          </div>
          <div class="text-2xl font-black text-blue-700 mt-2">
            {{ formatCurrency(summary.total_good_receivables) }}
          </div>
          <p class="text-xs text-slate-500 mt-1">Tagihan Surat Jalan status belum lunas</p>
        </div>

        <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden">
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold uppercase tracking-wider text-amber-600">Piutang Keuangan (Kas)</span>
            <div class="p-2 bg-amber-50 text-amber-600 rounded-xl">
              <Wallet class="w-5 h-5" />
            </div>
          </div>
          <div class="text-2xl font-black text-amber-600 mt-2">
            {{ formatCurrency(summary.total_financial_receivables) }}
          </div>
          <p class="text-xs text-slate-500 mt-1">Talangan Petty Cash & operasional cabang</p>
        </div>

        <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm relative overflow-hidden">
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold uppercase tracking-wider text-emerald-600">Setoran Bulan Ini</span>
            <div class="p-2 bg-emerald-50 text-emerald-600 rounded-xl">
              <CheckCircle2 class="w-5 h-5" />
            </div>
          </div>
          <div class="text-2xl font-black text-emerald-600 mt-2">
            {{ formatCurrency(summary.total_payments_this_month) }}
          </div>
          <p class="text-xs text-slate-500 mt-1">{{ summary.unpaid_branches_count }} cabang belum lunas</p>
        </div>
      </div>

      <!-- Main Module Tabs -->
      <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="border-b border-slate-200 bg-slate-50/70 p-4 flex flex-col md:flex-row justify-between items-stretch md:items-center gap-4">
          <!-- Tab Navigation Buttons -->
          <div class="flex flex-wrap items-center gap-1.5 p-1 bg-slate-200/70 rounded-xl">
            <button
              @click="activeTab = 'CONSOLIDATED'"
              class="px-4 py-2 text-xs font-bold rounded-lg transition-all flex items-center gap-2"
              :class="activeTab === 'CONSOLIDATED' ? 'bg-white text-indigo-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
            >
              <Layers class="w-3.5 h-3.5" />
              1. Konsolidasi Terpadu (Barang + Kas)
            </button>
            <button
              @click="activeTab = 'GOODS'"
              class="px-4 py-2 text-xs font-bold rounded-lg transition-all flex items-center gap-2"
              :class="activeTab === 'GOODS' ? 'bg-white text-blue-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
            >
              <Truck class="w-3.5 h-3.5" />
              2. Piutang Transaksi Barang (DO)
            </button>
            <button
              @click="activeTab = 'FINANCIAL'"
              class="px-4 py-2 text-xs font-bold rounded-lg transition-all flex items-center gap-2"
              :class="activeTab === 'FINANCIAL' ? 'bg-white text-amber-700 shadow-sm' : 'text-slate-600 hover:text-slate-900'"
            >
              <Wallet class="w-3.5 h-3.5" />
              3. Modul Piutang Keuangan (Talangan Kas)
            </button>
          </div>

          <!-- Search Filter -->
          <div class="relative w-full md:w-64">
            <Search class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              v-model="searchQuery"
              placeholder="Cari nama cabang..."
              class="w-full pl-9 pr-3 py-1.5 text-xs bg-white border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>
        </div>

        <!-- TAB 1: KONSOLIDASI TERPADU -->
        <div v-if="activeTab === 'CONSOLIDATED'" class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-100/75 text-slate-600 text-xs font-bold uppercase tracking-wider border-b border-slate-200">
              <tr>
                <th class="py-3.5 px-4">Nama Gudang Cabang</th>
                <th class="py-3.5 px-4 text-right text-blue-700">Piutang Barang (DO)</th>
                <th class="py-3.5 px-4 text-right text-amber-700">Talangan Kas</th>
                <th class="py-3.5 px-4 text-right text-slate-900">Total Piutang</th>
                <th class="py-3.5 px-4 text-right text-emerald-700">Setoran Bulan Ini</th>
                <th class="py-3.5 px-4 text-center">Status</th>
                <th class="py-3.5 px-4 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="filteredBranches.length === 0">
                <td colspan="7" class="text-center py-10 text-slate-400 font-medium">
                  Tidak ada data cabang yang cocok dengan pencarian.
                </td>
              </tr>
              <tr
                v-for="branch in filteredBranches"
                :key="branch.branch_id"
                class="hover:bg-slate-50/80 transition-colors"
              >
                <td class="py-3.5 px-4 font-semibold text-slate-900">
                  <div class="flex items-center gap-2">
                    <Building class="w-4 h-4 text-slate-400" />
                    <span>{{ branch.branch_name }}</span>
                  </div>
                </td>
                <td class="py-3.5 px-4 text-right font-medium text-blue-700">
                  {{ formatCurrency(branch.good_receivable) }}
                </td>
                <td class="py-3.5 px-4 text-right font-medium text-amber-700">
                  {{ formatCurrency(branch.financial_receivable) }}
                </td>
                <td class="py-3.5 px-4 text-right font-bold text-slate-900">
                  {{ formatCurrency(branch.total_receivable) }}
                </td>
                <td class="py-3.5 px-4 text-right font-medium text-emerald-600">
                  {{ formatCurrency(branch.total_payment_this_month) }}
                </td>
                <td class="py-3.5 px-4 text-center">
                  <span
                    v-if="branch.total_receivable > 0"
                    class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold bg-amber-100 text-amber-800"
                  >
                    Belum Lunas
                  </span>
                  <span
                    v-else
                    class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold bg-emerald-100 text-emerald-800"
                  >
                    Lunas
                  </span>
                </td>
                <td class="py-3.5 px-4 text-center">
                  <button
                    @click="openBranchHistory(branch.branch_id, branch.branch_name)"
                    class="px-2.5 py-1 text-xs font-semibold bg-slate-100 hover:bg-indigo-50 text-slate-700 hover:text-indigo-600 rounded-lg border border-slate-200 transition-all flex items-center gap-1 mx-auto"
                    title="Lihat riwayat talangan kas"
                  >
                    <FileText class="w-3.5 h-3.5" />
                    <span>Detail Kas</span>
                  </button>
                </td>
              </tr>
            </tbody>
            <tfoot class="bg-slate-100 font-bold border-t-2 border-slate-300 text-slate-900">
              <tr>
                <td class="py-3 px-4">TOTAL KESELURUHAN ({{ filteredBranches.length }} Cabang):</td>
                <td class="py-3 px-4 text-right text-blue-700">{{ formatCurrency(filteredTotalGood) }}</td>
                <td class="py-3 px-4 text-right text-amber-700">{{ formatCurrency(filteredTotalFinancial) }}</td>
                <td class="py-3 px-4 text-right text-slate-900 font-extrabold">{{ formatCurrency(filteredTotalCombined) }}</td>
                <td class="py-3 px-4 text-right text-emerald-700">{{ formatCurrency(filteredTotalPayment) }}</td>
                <td colspan="2"></td>
              </tr>
            </tfoot>
          </table>
        </div>

        <!-- TAB 2: PIUTANG TRANSAKSI BARANG (DO) -->
        <div v-else-if="activeTab === 'GOODS'" class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-100/75 text-slate-600 text-xs font-bold uppercase tracking-wider border-b border-slate-200">
              <tr>
                <th class="py-3.5 px-4">Nama Gudang Cabang</th>
                <th class="py-3.5 px-4 text-right">Tagihan DO Unpaid</th>
                <th class="py-3.5 px-4 text-right text-emerald-700">Setoran DO Bulan Ini</th>
                <th class="py-3.5 px-4 text-center">Status Barang</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="filteredBranches.length === 0">
                <td colspan="4" class="text-center py-10 text-slate-400 font-medium">
                  Belum ada data piutang barang.
                </td>
              </tr>
              <tr
                v-for="branch in filteredBranches"
                :key="branch.branch_id"
                class="hover:bg-slate-50/80 transition-colors"
              >
                <td class="py-3.5 px-4 font-semibold text-slate-900">{{ branch.branch_name }}</td>
                <td class="py-3.5 px-4 text-right font-bold text-blue-700">
                  {{ formatCurrency(branch.good_receivable) }}
                </td>
                <td class="py-3.5 px-4 text-right font-medium text-emerald-600">
                  {{ formatCurrency(branch.good_payment_this_month || branch.total_payment_this_month) }}
                </td>
                <td class="py-3.5 px-4 text-center">
                  <span
                    v-if="branch.good_receivable > 0"
                    class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold bg-blue-100 text-blue-800"
                  >
                    Tagihan Aktif
                  </span>
                  <span
                    v-else
                    class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold bg-emerald-100 text-emerald-800"
                  >
                    Lunas
                  </span>
                </td>
              </tr>
            </tbody>
            <tfoot class="bg-slate-100 font-bold border-t-2 border-slate-300 text-slate-900">
              <tr>
                <td class="py-3 px-4">TOTAL PIUTANG BARANG:</td>
                <td class="py-3 px-4 text-right text-blue-700 font-extrabold">{{ formatCurrency(filteredTotalGood) }}</td>
                <td class="py-3 px-4 text-right text-emerald-700">{{ formatCurrency(filteredTotalPayment) }}</td>
                <td></td>
              </tr>
            </tfoot>
          </table>
        </div>

        <!-- TAB 3: MODUL PIUTANG KEUANGAN (TALANGAN KAS / PETTY CASH) -->
        <div v-else-if="activeTab === 'FINANCIAL'" class="p-6 space-y-6">
          <!-- Info Alert -->
          <div class="p-4 bg-amber-50 border border-amber-200 rounded-xl flex items-start gap-3">
            <Wallet class="w-5 h-5 text-amber-600 mt-0.5 flex-shrink-0" />
            <div class="text-xs text-amber-900 leading-relaxed">
              <strong class="font-bold">Otomatisasi Talangan Kas Terintegrasi:</strong> Setiap pengeluaran kas di
              <span class="font-semibold underline">Manajemen KAS Pabrik</span> yang berjenis <em>Talangan Biaya Operasional Cabang</em>
              (misal talangan Petty Cash Karsa HR Purwokerto, Cabang Solo, dsb.) secara otomatis menambah saldo piutang keuangan cabang di modul ini.
              Saat cabang menyetor balik dan dicatat di Kas Masuk (<em>Pelunasan Talangan Piutang Cabang</em>), saldo piutang langsung berkurang secara otomatis.
            </div>
          </div>

          <!-- Financial Receivables Table -->
          <div class="overflow-x-auto border border-slate-200 rounded-xl">
            <table class="w-full text-left text-sm">
              <thead class="bg-slate-100 text-slate-600 text-xs font-bold uppercase tracking-wider border-b border-slate-200">
                <tr>
                  <th class="py-3.5 px-4">Nama Cabang</th>
                  <th class="py-3.5 px-4 text-right text-rose-700">Total Talangan Kas Diberikan</th>
                  <th class="py-3.5 px-4 text-right text-emerald-700">Total Pelunasan Kas Diterima</th>
                  <th class="py-3.5 px-4 text-right text-amber-700">Saldo Piutang Berjalan</th>
                  <th class="py-3.5 px-4 text-center">Status</th>
                  <th class="py-3.5 px-4 text-center">Riwayat Transaksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-if="filteredFinancialBranches.length === 0">
                  <td colspan="6" class="text-center py-10 text-slate-400 font-medium">
                    Belum ada data talangan kas cabang.
                  </td>
                </tr>
                <tr
                  v-for="b in filteredFinancialBranches"
                  :key="b.branch_id"
                  class="hover:bg-slate-50/80 transition-colors"
                >
                  <td class="py-3.5 px-4 font-semibold text-slate-900">
                    <div class="flex items-center gap-2">
                      <Building2 class="w-4 h-4 text-slate-400" />
                      <span>{{ b.branch_name }}</span>
                    </div>
                  </td>
                  <td class="py-3.5 px-4 text-right font-medium text-rose-600">
                    {{ formatCurrency(b.total_disbursement) }}
                  </td>
                  <td class="py-3.5 px-4 text-right font-medium text-emerald-600">
                    {{ formatCurrency(b.total_repayment) }}
                  </td>
                  <td class="py-3.5 px-4 text-right font-extrabold text-amber-600">
                    {{ formatCurrency(b.outstanding_balance) }}
                  </td>
                  <td class="py-3.5 px-4 text-center">
                    <span
                      v-if="b.outstanding_balance > 0"
                      class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold bg-amber-100 text-amber-800"
                    >
                      Belum Lunas
                    </span>
                    <span
                      v-else
                      class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-bold bg-emerald-100 text-emerald-800"
                    >
                      Lunas
                    </span>
                  </td>
                  <td class="py-3.5 px-4 text-center">
                    <button
                      @click="openBranchHistory(b.branch_id, b.branch_name)"
                      class="px-3 py-1 text-xs font-semibold bg-indigo-50 hover:bg-indigo-100 text-indigo-700 rounded-lg transition-all inline-flex items-center gap-1.5"
                    >
                      <Receipt class="w-3.5 h-3.5" />
                      <span>{{ b.transactions?.length || 0 }} Transaksi</span>
                    </button>
                  </td>
                </tr>
              </tbody>
              <tfoot class="bg-slate-100 font-bold border-t-2 border-slate-300 text-slate-900">
                <tr>
                  <td class="py-3 px-4">TOTAL PIUTANG KEUANGAN:</td>
                  <td class="py-3 px-4 text-right text-rose-700">{{ formatCurrency(financialSummaryData.total_disbursed) }}</td>
                  <td class="py-3 px-4 text-right text-emerald-700">{{ formatCurrency(financialSummaryData.total_repaid) }}</td>
                  <td class="py-3 px-4 text-right text-amber-700 font-extrabold">{{ formatCurrency(financialSummaryData.total_outstanding) }}</td>
                  <td colspan="2"></td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      </div>
    </template>

    <!-- Detail Drawer / Modal Riwayat Transaksi Kas Talangan Cabang -->
    <div
      v-if="showDetailModal"
      class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4"
      @click.self="showDetailModal = false"
    >
      <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl border border-slate-200 overflow-hidden flex flex-col max-h-[90vh]">
        <!-- Header -->
        <div class="p-5 border-b border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-slate-50">
          <div>
            <h3 class="text-lg font-bold text-slate-900 flex items-center gap-2">
              <Receipt class="w-5 h-5 text-indigo-600" />
              <span>Riwayat Transaksi Talangan Kas: {{ selectedBranchName }}</span>
            </h3>
            <p class="text-xs text-slate-500 mt-0.5">
              Daftar seluruh mutasi kas keluar (talangan), pelunasan kas masuk, dan penyelesaian SPJ nota dari Kas Pabrik.
            </p>
          </div>
          <div class="flex items-center gap-2">
            <button
              @click="openPaymentModal()"
              class="px-3.5 py-2 text-xs font-bold bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl shadow-sm transition-all flex items-center gap-1.5"
            >
              <Coins class="w-4 h-4" />
              <span>+ Catat Pelunasan Talangan</span>
            </button>
            <button
              @click="showDetailModal = false"
              class="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-200 transition-colors"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>

        <div class="p-6 overflow-y-auto space-y-5">
          <!-- Metric Summary Cards in Modal -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div class="p-3.5 bg-rose-50 border border-rose-200 rounded-xl">
              <div class="text-[11px] font-bold text-rose-700 uppercase tracking-wider">Total Talangan (Keluar)</div>
              <div class="text-lg font-black text-rose-900 mt-1">{{ formatCurrency(modalSummary.disbursed) }}</div>
              <div class="text-[10px] text-rose-600 mt-0.5">Akumulasi talangan operasional cabang</div>
            </div>

            <div class="p-3.5 bg-emerald-50 border border-emerald-200 rounded-xl">
              <div class="text-[11px] font-bold text-emerald-700 uppercase tracking-wider">Total Pelunasan & SPJ (Masuk)</div>
              <div class="text-lg font-black text-emerald-900 mt-1">{{ formatCurrency(modalSummary.repaid) }}</div>
              <div class="text-[10px] text-emerald-600 mt-0.5">Setoran kembali + nota SPJ yang sah</div>
            </div>

            <div class="p-3.5 bg-amber-50 border border-amber-200 rounded-xl">
              <div class="text-[11px] font-bold text-amber-700 uppercase tracking-wider flex items-center justify-between">
                <span>Sisa Tagihan Talangan</span>
                <span
                  class="px-2 py-0.5 rounded text-[10px] font-extrabold"
                  :class="modalSummary.outstanding <= 0 ? 'bg-emerald-200 text-emerald-900' : 'bg-amber-200 text-amber-900'"
                >
                  {{ modalSummary.outstanding <= 0 ? 'LUNAS' : 'BELUM LUNAS' }}
                </span>
              </div>
              <div class="text-lg font-black text-amber-900 mt-1">{{ formatCurrency(modalSummary.outstanding) }}</div>
              <div class="text-[10px] text-amber-700 mt-0.5">Kewajiban aktif yang harus diselesaikan</div>
            </div>
          </div>

          <!-- Filter Status Tabs in Modal -->
          <div class="flex items-center gap-2 border-b border-slate-200 pb-2">
            <button
              @click="modalActiveTab = 'ALL'"
              class="px-3 py-1.5 text-xs font-bold rounded-lg transition-all"
              :class="modalActiveTab === 'ALL' ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
            >
              Semua Transaksi ({{ selectedBranchTxs.length }})
            </button>
            <button
              @click="modalActiveTab = 'UNPAID'"
              class="px-3 py-1.5 text-xs font-bold rounded-lg transition-all flex items-center gap-1.5"
              :class="modalActiveTab === 'UNPAID' ? 'bg-amber-600 text-white' : 'bg-slate-100 text-amber-700 hover:bg-slate-200'"
            >
              <Clock class="w-3.5 h-3.5" />
              <span>Belum Lunas ({{ modalUnpaidCount }})</span>
            </button>
            <button
              @click="modalActiveTab = 'SETTLED'"
              class="px-3 py-1.5 text-xs font-bold rounded-lg transition-all flex items-center gap-1.5"
              :class="modalActiveTab === 'SETTLED' ? 'bg-emerald-600 text-white' : 'bg-slate-100 text-emerald-700 hover:bg-slate-200'"
            >
              <CheckCircle2 class="w-3.5 h-3.5" />
              <span>Sudah Lunas / Dibayarkan ({{ modalSettledCount }})</span>
            </button>
          </div>

          <!-- Transaction List Table -->
          <div v-if="filteredModalTxs.length === 0" class="text-center py-12 text-slate-400">
            <Receipt class="w-12 h-12 mx-auto text-slate-300 mb-2" />
            <p class="font-medium text-xs">Tidak ada catatan transaksi untuk kategori ini.</p>
          </div>

          <div v-else class="overflow-x-auto border border-slate-200 rounded-xl">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-100 text-slate-700 font-bold uppercase tracking-wider border-b border-slate-200">
                <tr>
                  <th class="py-3 px-3">Tanggal</th>
                  <th class="py-3 px-3">Buku KAS</th>
                  <th class="py-3 px-3">Tipe</th>
                  <th class="py-3 px-3">Kategori & Keterangan</th>
                  <th class="py-3 px-3 text-right">Nominal</th>
                  <th class="py-3 px-3 text-center">Status Pelunasan</th>
                  <th class="py-3 px-3 text-center">Bukti Nota</th>
                  <th class="py-3 px-3 text-center">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="tx in filteredModalTxs" :key="tx.id" class="hover:bg-slate-50 transition-colors">
                  <td class="py-3 px-3 font-medium text-slate-900 whitespace-nowrap">{{ formatDate(tx.transaction_date) }}</td>
                  <td class="py-3 px-3 font-semibold text-indigo-700 whitespace-nowrap">{{ tx.kas_name || 'KAS Pabrik' }}</td>
                  <td class="py-3 px-3 whitespace-nowrap">
                    <span
                      class="px-2 py-0.5 rounded text-[11px] font-bold inline-flex items-center gap-1"
                      :class="tx.transaction_type === 'OUT' ? 'bg-rose-100 text-rose-800' : 'bg-emerald-100 text-emerald-800'"
                    >
                      <ArrowUpRight v-if="tx.transaction_type === 'OUT'" class="w-3 h-3 text-rose-600" />
                      <ArrowDownLeft v-else class="w-3 h-3 text-emerald-600" />
                      {{ tx.transaction_type === 'OUT' ? 'TALANGAN (OUT)' : 'PELUNASAN (IN)' }}
                    </span>
                  </td>
                  <td class="py-3 px-3 min-w-[200px]">
                    <div class="font-semibold text-slate-900">{{ tx.expense_category || tx.income_source || 'Talangan Biaya Operasional' }}</div>
                    <div class="text-slate-500 text-[11px] mt-0.5">{{ tx.description || '-' }}</div>
                  </td>
                  <td
                    class="py-3 px-3 text-right font-bold whitespace-nowrap"
                    :class="tx.transaction_type === 'OUT' ? 'text-rose-600' : 'text-emerald-600'"
                  >
                    {{ tx.transaction_type === 'OUT' ? '-' : '+' }} {{ formatCurrency(tx.amount) }}
                  </td>
                  <td class="py-3 px-3 text-center whitespace-nowrap">
                    <!-- IN Transaksi Pelunasan -->
                    <span
                      v-if="tx.transaction_type === 'IN'"
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-bold bg-emerald-100 text-emerald-800"
                    >
                      <CheckCircle2 class="w-3 h-3 text-emerald-600" /> Pelunasan Diterima
                    </span>
                    <!-- OUT Transaksi dengan Settlement SPJ -->
                    <div v-else-if="tx.settlement" class="inline-flex flex-col items-center">
                      <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-bold bg-emerald-100 text-emerald-800">
                        <CheckCircle2 class="w-3 h-3 text-emerald-600" /> Selesai SPJ / Lunas
                      </span>
                      <div class="text-[10px] text-slate-500 mt-0.5 font-medium">
                        Realisasi: {{ formatCurrency(tx.settlement.actual_expense_amount) }}
                        <span v-if="tx.settlement.refund_amount > 0">, Kembali: {{ formatCurrency(tx.settlement.refund_amount) }}</span>
                      </div>
                    </div>
                    <!-- OUT Transaksi Sudah Lunas -->
                    <span
                      v-else-if="tx.is_settled || tx.advance_status === 'SETTLED'"
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-bold bg-emerald-100 text-emerald-800"
                    >
                      <CheckCircle2 class="w-3 h-3 text-emerald-600" /> Sudah Lunas
                    </span>
                    <!-- OUT Transaksi Belum Lunas -->
                    <span
                      v-else
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-bold bg-amber-100 text-amber-800"
                    >
                      <Clock class="w-3 h-3 text-amber-600" /> Belum Lunas
                    </span>
                  </td>
                  <td class="py-3 px-3 text-center whitespace-nowrap">
                    <a
                      v-if="tx.proof_url"
                      :href="tx.proof_url"
                      target="_blank"
                      class="text-indigo-600 hover:text-indigo-800 font-semibold underline text-[11px] inline-flex items-center gap-1"
                    >
                      <FileText class="w-3.5 h-3.5" />
                      Lihat Nota
                    </a>
                    <a
                      v-else-if="tx.settlement?.proof_url"
                      :href="tx.settlement.proof_url"
                      target="_blank"
                      class="text-indigo-600 hover:text-indigo-800 font-semibold underline text-[11px] inline-flex items-center gap-1"
                    >
                      <FileText class="w-3.5 h-3.5" />
                      Bukti SPJ
                    </a>
                    <span v-else class="text-slate-400">-</span>
                  </td>
                  <td class="py-3 px-3 text-center whitespace-nowrap">
                    <div v-if="tx.transaction_type === 'OUT' && !tx.is_settled" class="flex items-center justify-center gap-1.5">
                      <button
                        @click="openPaymentModal(tx)"
                        class="px-2 py-1 text-[11px] font-bold bg-emerald-50 hover:bg-emerald-100 text-emerald-700 rounded-lg border border-emerald-200 transition-all flex items-center gap-1 shadow-xs"
                        title="Catat Pembayaran Pelunasan"
                      >
                        <Coins class="w-3 h-3" />
                        Lunasi
                      </button>
                      <button
                        @click="openSettleModal(tx)"
                        class="px-2 py-1 text-[11px] font-bold bg-indigo-50 hover:bg-indigo-100 text-indigo-700 rounded-lg border border-indigo-200 transition-all flex items-center gap-1 shadow-xs"
                        title="Tandai Lunas via SPJ Nota"
                      >
                        <Check class="w-3 h-3" />
                        SPJ
                      </button>
                    </div>
                    <span v-else class="text-slate-400">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between">
          <div class="text-xs text-slate-500">
            Total {{ filteredModalTxs.length }} data mutasi ditampilkan
          </div>
          <button
            @click="showDetailModal = false"
            class="px-4 py-2 text-sm font-semibold bg-slate-800 hover:bg-slate-900 text-white rounded-xl shadow-sm transition-all"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Form: Catat Pelunasan Talangan Kas -->
    <div
      v-if="showPaymentModal"
      class="fixed inset-0 z-[60] bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4"
      @click.self="showPaymentModal = false"
    >
      <div class="bg-white w-full max-w-lg rounded-2xl shadow-2xl border border-slate-200 overflow-hidden flex flex-col">
        <div class="p-5 border-b border-slate-200 flex items-center justify-between bg-emerald-50/50">
          <div class="flex items-center gap-2">
            <Coins class="w-5 h-5 text-emerald-600" />
            <h3 class="text-base font-bold text-slate-900">Catat Pelunasan Talangan Cabang</h3>
          </div>
          <button
            @click="showPaymentModal = false"
            class="p-1.5 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <div class="p-5 space-y-4 text-xs">
          <!-- Info Cabang & Sisa -->
          <div class="p-3 bg-slate-50 border border-slate-200 rounded-xl space-y-1">
            <div class="flex justify-between items-center">
              <span class="text-slate-500 font-medium">Cabang Penerima Talangan:</span>
              <span class="font-bold text-slate-900">{{ selectedBranchName }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-slate-500 font-medium">Sisa Piutang Berjalan:</span>
              <span class="font-extrabold text-amber-700">{{ formatCurrency(modalSummary.outstanding) }}</span>
            </div>
          </div>

          <!-- Pilihan Buku KAS Penerima -->
          <div>
            <label class="block font-bold text-slate-700 mb-1">
              Buku KAS Pabrik Penerima Setoran <span class="text-rose-500">*</span>
            </label>
            <select
              v-model="paymentForm.kas_pabrik_id"
              class="w-full p-2.5 bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:outline-none"
              required
            >
              <option value="" disabled>-- Pilih Buku KAS Pabrik --</option>
              <option v-for="kas in kasList" :key="kas.id" :value="kas.id">
                {{ kas.kas_name }} (Saldo: {{ formatCurrency(kas.balance) }})
              </option>
            </select>
          </div>

          <!-- Tanggal Pembayaran -->
          <div>
            <label class="block font-bold text-slate-700 mb-1">
              Tanggal Pelunasan / Setor <span class="text-rose-500">*</span>
            </label>
            <input
              type="date"
              v-model="paymentForm.transaction_date"
              class="w-full p-2 bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:outline-none"
              required
            />
          </div>

          <!-- Nominal Pelunasan -->
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="font-bold text-slate-700">
                Nominal Pelunasan (Rp) <span class="text-rose-500">*</span>
              </label>
              <button
                type="button"
                @click="paymentForm.amount = modalSummary.outstanding"
                class="text-[11px] font-semibold text-emerald-600 hover:text-emerald-700 underline"
              >
                Bayar Penuh Sesuai Sisa
              </button>
            </div>
            <input
              type="number"
              v-model="paymentForm.amount"
              min="1"
              placeholder="Masukkan nominal setoran..."
              class="w-full p-2.5 bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:outline-none font-bold text-sm text-slate-900"
              required
            />
          </div>

          <!-- Keterangan -->
          <div>
            <label class="block font-bold text-slate-700 mb-1">Keterangan / Catatan</label>
            <textarea
              v-model="paymentForm.description"
              rows="2"
              placeholder="Contoh: Setoran pelunasan talangan operasional cabang..."
              class="w-full p-2.5 bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:outline-none"
            ></textarea>
          </div>

          <!-- Upload Bukti Transfer -->
          <div>
            <label class="block font-bold text-slate-700 mb-1">
              Upload Bukti Transfer / Nota Setor <span class="text-slate-400 font-normal">(Opsional)</span>
            </label>
            <div class="border-2 border-dashed border-slate-200 rounded-xl p-3 text-center bg-slate-50/50 hover:bg-slate-50 transition-colors">
              <input
                type="file"
                id="paymentProofInput"
                class="hidden"
                accept="image/*,application/pdf"
                @change="onPaymentFileChange"
              />
              <label for="paymentProofInput" class="cursor-pointer flex flex-col items-center justify-center gap-1">
                <Upload class="w-5 h-5 text-slate-400" />
                <span class="text-xs font-semibold text-indigo-600">
                  {{ paymentProofFile ? paymentProofFile.name : (paymentForm.proof_url ? 'Bukti terunggah (klik untuk ganti)' : 'Pilih Berkas Bukti Transfer') }}
                </span>
                <span class="text-[10px] text-slate-400">JPG, PNG, atau PDF max 5MB</span>
              </label>
            </div>
          </div>
        </div>

        <div class="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-end gap-2">
          <button
            type="button"
            @click="showPaymentModal = false"
            class="px-4 py-2 text-xs font-semibold bg-slate-200 hover:bg-slate-300 text-slate-700 rounded-xl transition-colors"
          >
            Batal
          </button>
          <button
            type="button"
            @click="submitPayment"
            :disabled="!paymentForm.kas_pabrik_id || !paymentForm.amount || isSubmittingPayment"
            class="px-4 py-2 text-xs font-bold bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white rounded-xl shadow-sm transition-all flex items-center gap-1.5"
          >
            <Loader2 v-if="isSubmittingPayment" class="w-3.5 h-3.5 animate-spin" />
            <span>{{ isSubmittingPayment ? 'Menyimpan...' : 'Simpan Pelunasan' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Form: Tandai Lunas via SPJ / Nota Operasional -->
    <div
      v-if="showSettleModal"
      class="fixed inset-0 z-[60] bg-slate-900/60 backdrop-blur-sm flex items-center justify-center p-4"
      @click.self="showSettleModal = false"
    >
      <div class="bg-white w-full max-w-lg rounded-2xl shadow-2xl border border-slate-200 overflow-hidden flex flex-col">
        <div class="p-5 border-b border-slate-200 flex items-center justify-between bg-indigo-50/50">
          <div class="flex items-center gap-2">
            <CheckCircle2 class="w-5 h-5 text-indigo-600" />
            <h3 class="text-base font-bold text-slate-900">Tandai Lunas via SPJ / Nota Operasional</h3>
          </div>
          <button
            @click="showSettleModal = false"
            class="p-1.5 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <div class="p-5 space-y-4 text-xs">
          <!-- Info Transaksi -->
          <div class="p-3 bg-slate-50 border border-slate-200 rounded-xl space-y-1">
            <div class="flex justify-between items-center">
              <span class="text-slate-500 font-medium">Transaksi Talangan:</span>
              <span class="font-bold text-slate-900">{{ selectedSettleTx?.expense_category || 'Talangan Biaya' }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-slate-500 font-medium">Nominal Talangan Diberikan:</span>
              <span class="font-extrabold text-rose-600">{{ formatCurrency(selectedSettleTx?.amount) }}</span>
            </div>
            <div class="text-[11px] text-slate-500 mt-1">
              Catatan awal: {{ selectedSettleTx?.description || '-' }}
            </div>
          </div>

          <!-- Realisasi Biaya Aktual -->
          <div>
            <label class="block font-bold text-slate-700 mb-1">
              Realisasi Pengeluaran SPJ (Rp) <span class="text-rose-500">*</span>
            </label>
            <input
              type="number"
              v-model="settleForm.actual_expense_amount"
              min="0"
              class="w-full p-2.5 bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:outline-none font-bold text-sm text-slate-900"
              required
            />
            <p class="text-[11px] text-slate-500 mt-1">
              Jika realisasi biaya lebih kecil dari nominal talangan, sisa dana otomatis dicatat sebagai pengembalian kas (refund) ke Kas Pabrik.
            </p>
          </div>

          <!-- Catatan / Rincian SPJ -->
          <div>
            <label class="block font-bold text-slate-700 mb-1">Catatan Pertanggungjawaban SPJ <span class="text-rose-500">*</span></label>
            <textarea
              v-model="settleForm.notes"
              rows="3"
              placeholder="Contoh: Pembelian perlengkapan operasional kantor cabang sesuai nota terlampir..."
              class="w-full p-2.5 bg-white border border-slate-300 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:outline-none"
              required
            ></textarea>
          </div>

          <!-- Upload Nota Realisasi SPJ -->
          <div>
            <label class="block font-bold text-slate-700 mb-1">
              Upload Foto Struk / Nota SPJ <span class="text-slate-400 font-normal">(Opsional)</span>
            </label>
            <div class="border-2 border-dashed border-slate-200 rounded-xl p-3 text-center bg-slate-50/50 hover:bg-slate-50 transition-colors">
              <input
                type="file"
                id="settleProofInput"
                class="hidden"
                accept="image/*,application/pdf"
                @change="onSettleFileChange"
              />
              <label for="settleProofInput" class="cursor-pointer flex flex-col items-center justify-center gap-1">
                <Upload class="w-5 h-5 text-slate-400" />
                <span class="text-xs font-semibold text-indigo-600">
                  {{ settleProofFile ? settleProofFile.name : (settleForm.proof_url ? 'Bukti SPJ terunggah' : 'Pilih Berkas Nota SPJ') }}
                </span>
                <span class="text-[10px] text-slate-400">JPG, PNG, atau PDF max 5MB</span>
              </label>
            </div>
          </div>
        </div>

        <div class="p-4 border-t border-slate-200 bg-slate-50 flex items-center justify-end gap-2">
          <button
            type="button"
            @click="showSettleModal = false"
            class="px-4 py-2 text-xs font-semibold bg-slate-200 hover:bg-slate-300 text-slate-700 rounded-xl transition-colors"
          >
            Batal
          </button>
          <button
            type="button"
            @click="submitSettle"
            :disabled="!settleForm.notes || isSubmittingSettle"
            class="px-4 py-2 text-xs font-bold bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-xl shadow-sm transition-all flex items-center gap-1.5"
          >
            <Loader2 v-if="isSubmittingSettle" class="w-3.5 h-3.5 animate-spin" />
            <span>{{ isSubmittingSettle ? 'Menyimpan SPJ...' : 'Selesaikan SPJ (Lunas)' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  Building2,
  Building,
  Layers,
  Truck,
  Wallet,
  CheckCircle2,
  AlertCircle,
  Loader2,
  RefreshCw,
  Printer,
  Search,
  FileText,
  Receipt,
  X,
  Coins,
  Check,
  Clock,
  ArrowUpRight,
  ArrowDownLeft,
  Upload
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { factoryApi } from '@/api/factory'

interface Summary {
  total_receivables: number
  total_good_receivables: number
  total_financial_receivables: number
  total_payments_this_month: number
  unpaid_branches_count: number
}

interface BranchReceivable {
  branch_id: string
  branch_name: string
  good_receivable: number
  financial_receivable: number
  total_receivable: number
  good_payment_this_month?: number
  financial_payment_this_month?: number
  total_payment_this_month: number
}

interface KasAdvanceSettlementItem {
  id: string
  kas_transaction_id: string
  actual_expense_amount: number
  refund_amount: number
  shortage_amount: number
  notes?: string
  proof_url?: string
  settled_at?: string
}

interface BranchFinancialItem {
  id: string
  kas_transaction_id?: string
  branch_id: string
  branch_name: string
  kas_pabrik_id: string
  kas_name: string
  transaction_type: string
  amount: number
  expense_category?: string
  income_source?: string
  transaction_date: string
  description?: string
  proof_url?: string
  advance_status?: string
  is_settled?: boolean
  settlement?: KasAdvanceSettlementItem
  related_transaction_id?: string
  created_at?: string
}

interface BranchFinancialSummary {
  branch_id: string
  branch_name: string
  total_disbursement: number
  total_repayment: number
  outstanding_balance: number
  total_repayment_month: number
  status: string
  transactions?: BranchFinancialItem[]
}

const loading = ref(true)
const error = ref('')
const activeTab = ref<'CONSOLIDATED' | 'GOODS' | 'FINANCIAL'>('CONSOLIDATED')
const searchQuery = ref('')

const summary = ref<Summary>({
  total_receivables: 0,
  total_good_receivables: 0,
  total_financial_receivables: 0,
  total_payments_this_month: 0,
  unpaid_branches_count: 0
})

const branches = ref<BranchReceivable[]>([])
const financialBranches = ref<BranchFinancialSummary[]>([])
const financialSummaryData = ref({
  total_outstanding: 0,
  total_disbursed: 0,
  total_repaid: 0
})

// Modal Detail State
const showDetailModal = ref(false)
const selectedBranchId = ref('')
const selectedBranchName = ref('')
const selectedBranchTxs = ref<BranchFinancialItem[]>([])
const modalActiveTab = ref<'ALL' | 'UNPAID' | 'SETTLED'>('ALL')

// Form Modal: Catat Pelunasan Talangan
const showPaymentModal = ref(false)
const kasList = ref<any[]>([])
const isSubmittingPayment = ref(false)
const paymentProofFile = ref<File | null>(null)
const paymentForm = ref({
  kas_pabrik_id: '',
  amount: 0,
  transaction_date: '',
  description: '',
  proof_url: '',
  target_transaction_id: ''
})

// Form Modal: Tandai Lunas via SPJ
const showSettleModal = ref(false)
const selectedSettleTx = ref<BranchFinancialItem | null>(null)
const settleProofFile = ref<File | null>(null)
const isSubmittingSettle = ref(false)
const settleForm = ref({
  actual_expense_amount: 0,
  notes: '',
  proof_url: ''
})

const formatCurrency = (val?: number) => {
  if (!val) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(val)
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

// Filtered Branches
const filteredBranches = computed(() => {
  if (!searchQuery.value.trim()) return branches.value
  const q = searchQuery.value.toLowerCase()
  return branches.value.filter(b => b.branch_name.toLowerCase().includes(q))
})

const filteredFinancialBranches = computed(() => {
  if (!searchQuery.value.trim()) return financialBranches.value
  const q = searchQuery.value.toLowerCase()
  return financialBranches.value.filter(b => b.branch_name.toLowerCase().includes(q))
})

const filteredTotalGood = computed(() =>
  filteredBranches.value.reduce((sum, b) => sum + (Number(b.good_receivable) || 0), 0)
)
const filteredTotalFinancial = computed(() =>
  filteredBranches.value.reduce((sum, b) => sum + (Number(b.financial_receivable) || 0), 0)
)
const filteredTotalCombined = computed(() =>
  filteredBranches.value.reduce((sum, b) => sum + (Number(b.total_receivable) || 0), 0)
)
const filteredTotalPayment = computed(() =>
  filteredBranches.value.reduce((sum, b) => sum + (Number(b.total_payment_this_month) || 0), 0)
)

// Computed for Modal Detail
const selectedBranchFinancial = computed(() => {
  return financialBranches.value.find(b => b.branch_id === selectedBranchId.value)
})

const modalSummary = computed(() => {
  const b = selectedBranchFinancial.value
  return {
    disbursed: b?.total_disbursement || 0,
    repaid: b?.total_repayment || 0,
    outstanding: b?.outstanding_balance || 0,
    status: b?.status || 'LUNAS'
  }
})

const modalUnpaidCount = computed(() => {
  return selectedBranchTxs.value.filter(t => t.transaction_type === 'OUT' && !t.is_settled).length
})

const modalSettledCount = computed(() => {
  return selectedBranchTxs.value.filter(t => t.transaction_type === 'IN' || t.is_settled).length
})

const filteredModalTxs = computed(() => {
  if (modalActiveTab.value === 'UNPAID') {
    return selectedBranchTxs.value.filter(t => t.transaction_type === 'OUT' && !t.is_settled)
  }
  if (modalActiveTab.value === 'SETTLED') {
    return selectedBranchTxs.value.filter(t => t.transaction_type === 'IN' || t.is_settled)
  }
  return selectedBranchTxs.value
})

const fetchData = async () => {
  loading.value = true
  error.value = ''
  try {
    const [recRes, finRes] = await Promise.allSettled([
      factoryApi.getBranchReceivables(),
      apiClient.get('/finance/receivables/financial')
    ])

    if (recRes.status === 'fulfilled' && recRes.value.data?.data) {
      const data = recRes.value.data.data
      summary.value = {
        total_receivables: data.summary?.total_receivables || 0,
        total_good_receivables: data.summary?.total_good_receivables || 0,
        total_financial_receivables: data.summary?.total_financial_receivables || 0,
        total_payments_this_month: data.summary?.total_payments_this_month || 0,
        unpaid_branches_count: data.summary?.unpaid_branches_count || 0
      }
      branches.value = data.branches || []
    }

    if (finRes.status === 'fulfilled' && finRes.value.data?.data) {
      const finData = finRes.value.data.data
      financialBranches.value = finData.branches || []
      financialSummaryData.value = {
        total_outstanding: finData.total_outstanding || 0,
        total_disbursed: finData.total_disbursed || 0,
        total_repaid: finData.total_repaid || 0
      }

      // If detail modal is open, refresh selectedBranchTxs
      if (selectedBranchId.value) {
        const currentFinBranch = financialBranches.value.find(b => b.branch_id === selectedBranchId.value)
        selectedBranchTxs.value = currentFinBranch?.transactions || []
      }
    }
  } catch (err: any) {
    error.value = err?.response?.data?.message || 'Gagal memuat data piutang cabang'
  } finally {
    loading.value = false
  }
}

const fetchKasList = async () => {
  try {
    const res = await apiClient.get('/kas-pabrik')
    kasList.value = res.data?.data || []
    if (kasList.value.length > 0 && !paymentForm.value.kas_pabrik_id) {
      paymentForm.value.kas_pabrik_id = kasList.value[0].id
    }
  } catch (err) {
    console.error('Gagal mengambil daftar KAS Pabrik:', err)
  }
}

const openBranchHistory = (branchId: string, branchName: string) => {
  selectedBranchId.value = branchId
  selectedBranchName.value = branchName
  modalActiveTab.value = 'ALL'
  const finBranch = financialBranches.value.find(b => b.branch_id === branchId)
  selectedBranchTxs.value = finBranch?.transactions || []
  showDetailModal.value = true
}

const onPaymentFileChange = (e: any) => {
  const files = e.target.files
  if (files && files.length > 0) {
    paymentProofFile.value = files[0]
  }
}

const onSettleFileChange = (e: any) => {
  const files = e.target.files
  if (files && files.length > 0) {
    settleProofFile.value = files[0]
  }
}

const openPaymentModal = async (targetTx?: BranchFinancialItem) => {
  await fetchKasList()
  paymentProofFile.value = null
  paymentForm.value = {
    kas_pabrik_id: kasList.value[0]?.id || '',
    amount: targetTx ? targetTx.amount : (modalSummary.value.outstanding > 0 ? modalSummary.value.outstanding : 0),
    transaction_date: new Date().toISOString().split('T')[0],
    description: targetTx
      ? `Pelunasan talangan: ${targetTx.expense_category || targetTx.description || ''}`
      : `Pelunasan talangan operasional cabang ${selectedBranchName.value}`,
    proof_url: '',
    target_transaction_id: targetTx ? targetTx.id : ''
  }
  showPaymentModal.value = true
}

const submitPayment = async () => {
  if (!paymentForm.value.kas_pabrik_id) {
    toast.error('Pilih Buku KAS Pabrik penerima setoran!')
    return
  }
  if (!paymentForm.value.amount || paymentForm.value.amount <= 0) {
    toast.error('Nominal pelunasan harus lebih dari Rp 0!')
    return
  }

  isSubmittingPayment.value = true
  try {
    let uploadedProofUrl = paymentForm.value.proof_url
    if (paymentProofFile.value) {
      const formData = new FormData()
      formData.append('file', paymentProofFile.value)
      try {
        const uploadRes = await apiClient.post('/cashier-reports/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        uploadedProofUrl = uploadRes.data?.data?.url || uploadRes.data?.url || ''
      } catch (uploadErr) {
        console.warn('Upload bukti gagal:', uploadErr)
      }
    }

    const payload = {
      branch_id: selectedBranchId.value,
      kas_pabrik_id: paymentForm.value.kas_pabrik_id,
      amount: Number(paymentForm.value.amount),
      transaction_date: paymentForm.value.transaction_date,
      description: paymentForm.value.description,
      proof_url: uploadedProofUrl,
      target_transaction_id: paymentForm.value.target_transaction_id || undefined
    }

    await apiClient.post('/finance/receivables/financial/payment', payload)
    toast.success('Pelunasan talangan cabang berhasil dicatat ke Kas Pabrik!')
    showPaymentModal.value = false
    await fetchData()
  } catch (err: any) {
    toast.error(err?.response?.data?.error || err?.response?.data?.message || 'Gagal mencatat pelunasan talangan')
  } finally {
    isSubmittingPayment.value = false
  }
}

const openSettleModal = (tx: BranchFinancialItem) => {
  selectedSettleTx.value = tx
  settleProofFile.value = null
  settleForm.value = {
    actual_expense_amount: tx.amount,
    notes: `Pertanggungjawaban nota SPJ: ${tx.expense_category || tx.description || ''}`,
    proof_url: ''
  }
  showSettleModal.value = true
}

const submitSettle = async () => {
  if (!selectedSettleTx.value) return
  if (!settleForm.value.notes) {
    toast.error('Catatan pertanggungjawaban SPJ wajib diisi!')
    return
  }

  isSubmittingSettle.value = true
  try {
    let uploadedProofUrl = settleForm.value.proof_url
    if (settleProofFile.value) {
      const formData = new FormData()
      formData.append('file', settleProofFile.value)
      try {
        const uploadRes = await apiClient.post('/cashier-reports/upload', formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
        uploadedProofUrl = uploadRes.data?.data?.url || uploadRes.data?.url || ''
      } catch (uploadErr) {
        console.warn('Upload bukti SPJ gagal:', uploadErr)
      }
    }

    const payload = {
      actual_expense_amount: Number(settleForm.value.actual_expense_amount),
      notes: settleForm.value.notes,
      proof_url: uploadedProofUrl
    }

    await apiClient.post(`/finance/receivables/financial/transactions/${selectedSettleTx.value.id}/settle`, payload)
    toast.success('Transaksi talangan berhasil diselesaikan (SPJ) & status lunas!')
    showSettleModal.value = false
    await fetchData()
  } catch (err: any) {
    toast.error(err?.response?.data?.error || err?.response?.data?.message || 'Gagal menyelesaikan transaksi talangan')
  } finally {
    isSubmittingSettle.value = false
  }
}

const handlePrint = () => {
  window.print()
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
@media print {
  body {
    background: white;
  }
  button, .filter-bar {
    display: none !important;
  }
}
</style>
