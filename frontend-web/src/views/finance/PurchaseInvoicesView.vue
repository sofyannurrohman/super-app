<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6 bg-white p-6 rounded-3xl border border-slate-200/60 shadow-sm">
      <div class="flex items-center gap-4">
        <div class="p-3 bg-rose-50 text-rose-600 rounded-2xl border border-rose-100">
          <Receipt class="w-7 h-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black text-slate-900 tracking-tight flex items-center gap-2">
            Faktur Pembelian & Hutang Supplier
          </h1>
          <p class="text-sm font-medium text-slate-500 mt-1">
            Kelola tagihan faktur supplier, kontrol jatuh tempo (TOP), lampiran faktur, dan pembayaran hutang operasional.
          </p>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <!-- Branch Selector if needed -->
        <select
          v-model="selectedBranch"
          @change="fetchInvoices"
          class="bg-white border border-slate-200 px-4 py-2.5 rounded-2xl text-xs font-bold text-slate-700 shadow-sm outline-none focus:ring-2 focus:ring-rose-500/20"
        >
          <option value="ALL">Semua Cabang / Gudang</option>
          <option v-for="b in masterStore.branches" :key="b.id" :value="b.id">{{ b.name }}</option>
        </select>

        <button
          @click="fetchInvoices"
          class="p-2.5 bg-white border border-slate-200 text-slate-600 hover:text-slate-900 rounded-2xl transition-all shadow-sm active:scale-95"
          title="Refresh Data"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isLoading }" />
        </button>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white p-6 rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <div class="flex items-center justify-between mb-2">
          <p class="text-xs font-bold text-slate-500 uppercase">Total Tagihan Faktur</p>
          <div class="p-2 bg-slate-100 rounded-xl text-slate-600">
            <FileText class="w-4 h-4" />
          </div>
        </div>
        <p class="text-2xl font-black text-slate-900">{{ formatCurrency(summary.totalInvoice) }}</p>
        <p class="text-xs text-slate-400 mt-1">{{ invoices.length }} faktur tercatat</p>
      </div>

      <div class="bg-white p-6 rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <div class="flex items-center justify-between mb-2">
          <p class="text-xs font-bold text-emerald-600 uppercase">Total Terbayar</p>
          <div class="p-2 bg-emerald-50 rounded-xl text-emerald-600">
            <CheckCircle2 class="w-4 h-4" />
          </div>
        </div>
        <p class="text-2xl font-black text-emerald-600">{{ formatCurrency(summary.totalPaid) }}</p>
        <p class="text-xs text-slate-400 mt-1">{{ summary.paidCount }} faktur telah lunas</p>
      </div>

      <div class="bg-white p-6 rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <div class="flex items-center justify-between mb-2">
          <p class="text-xs font-bold text-rose-600 uppercase">Sisa Hutang Berjalan</p>
          <div class="p-2 bg-rose-50 rounded-xl text-rose-600">
            <TrendingDown class="w-4 h-4" />
          </div>
        </div>
        <p class="text-2xl font-black text-rose-700">{{ formatCurrency(summary.totalRemaining) }}</p>
        <p class="text-xs text-rose-500 font-semibold mt-1">{{ summary.unpaidCount }} faktur belum lunas</p>
      </div>

      <div class="bg-white p-6 rounded-3xl border border-slate-200/60 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <div class="flex items-center justify-between mb-2">
          <p class="text-xs font-bold text-amber-600 uppercase">Jatuh Tempo (Overdue/Due Soon)</p>
          <div class="p-2 bg-amber-50 rounded-xl text-amber-600">
            <Clock class="w-4 h-4" />
          </div>
        </div>
        <p class="text-2xl font-black text-amber-600">{{ summary.overdueCount }} Faktur</p>
        <p class="text-xs text-slate-400 mt-1">Perlu prioritas pembayaran</p>
      </div>
    </div>

    <!-- Filter & Search Bar -->
    <div class="bg-white p-4 rounded-3xl border border-slate-200/60 shadow-sm flex flex-col md:flex-row items-center justify-between gap-4">
      <div class="flex flex-wrap items-center gap-2 w-full md:w-auto">
        <button
          v-for="st in statusOptions"
          :key="st.key"
          @click="filterStatus = st.key"
          class="px-4 py-2 rounded-2xl text-xs font-bold transition-all"
          :class="filterStatus === st.key ? 'bg-slate-900 text-white shadow-sm' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'"
        >
          {{ st.label }}
        </button>
      </div>

      <div class="relative w-full md:w-80">
        <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          v-model="searchQuery"
          placeholder="Cari no. faktur, supplier, no. SJ..."
          class="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-medium text-slate-700 outline-none focus:ring-2 focus:ring-rose-500/20 focus:border-rose-300 transition-all"
        />
      </div>
    </div>

    <!-- Data Invoices Table -->
    <div class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_25px_rgba(0,0,0,0.03)] overflow-hidden">
      <div v-if="isLoading" class="flex flex-col items-center justify-center py-20">
        <Loader2 class="w-8 h-8 animate-spin text-rose-500 mb-2" />
        <p class="text-xs font-bold text-slate-400">Memuat faktur pembelian supplier...</p>
      </div>

      <div v-else-if="filteredInvoices.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
        <Receipt class="w-16 h-16 mb-3 opacity-20" />
        <p class="font-bold text-base text-slate-700">Tidak ada faktur pembelian</p>
        <p class="text-xs text-slate-400 mt-0.5">Semua faktur dari penerimaan gudang/pabrik akan muncul di sini</p>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left border-collapse min-w-[900px]">
          <thead>
            <tr class="bg-slate-900 text-white uppercase text-[10px] font-black tracking-[0.15em]">
              <th class="px-5 py-4 w-12 text-center">No</th>
              <th class="px-5 py-4">Faktur Supplier / Internal</th>
              <th class="px-5 py-4">Supplier & Cabang</th>
              <th class="px-5 py-4">Tgl Faktur & TOP</th>
              <th class="px-5 py-4 text-right">Nilai Faktur</th>
              <th class="px-5 py-4 text-right">Terbayar</th>
              <th class="px-5 py-4 text-right">Sisa Hutang</th>
              <th class="px-5 py-4 text-center">Status</th>
              <th class="px-5 py-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700 font-medium text-xs">
            <tr v-for="(inv, idx) in filteredInvoices" :key="inv.id" class="hover:bg-slate-50 transition-colors">
              <td class="px-5 py-4 text-center text-slate-400">{{ idx + 1 }}</td>
              <td class="px-5 py-4">
                <div class="flex items-center gap-2">
                  <span class="font-black text-slate-900 font-mono text-sm">{{ inv.supplier_invoice_no || inv.receipt_no }}</span>
                  <a
                    v-if="inv.proof_image_url"
                    :href="inv.proof_image_url"
                    target="_blank"
                    class="p-1 text-teal-600 hover:bg-teal-50 rounded-lg"
                    title="Lihat Foto / Dokumen Faktur"
                  >
                    <ExternalLink class="w-3.5 h-3.5" />
                  </a>
                </div>
                <p class="text-[11px] text-slate-400 font-mono mt-0.5">Ref: {{ inv.receipt_no }}</p>
              </td>
              <td class="px-5 py-4">
                <p class="font-bold text-slate-800 text-sm">{{ inv.supplier?.name || inv.supplier_name || 'Supplier Umum' }}</p>
                <p class="text-[11px] text-slate-400 mt-0.5 flex items-center gap-1">
                  <Building2 class="w-3 h-3" />
                  <span>{{ inv.branch?.name || 'Pabrik / Cabang' }}</span>
                </p>
              </td>
              <td class="px-5 py-4">
                <div class="space-y-0.5">
                  <p class="text-slate-700 font-semibold">{{ formatDate(inv.invoice_date || inv.received_date || inv.created_at) }}</p>
                  <p v-if="inv.payment_due_date" class="text-[11px] flex items-center gap-1" :class="isOverdue(inv) ? 'text-rose-600 font-bold' : 'text-slate-400'">
                    <Clock class="w-3 h-3" />
                    <span>Jatuh Tempo: {{ formatDate(inv.payment_due_date) }}</span>
                  </p>
                  <p v-else class="text-[10px] text-slate-400 font-medium">Jatuh Tempo: -</p>
                </div>
              </td>
              <td class="px-5 py-4 text-right font-black text-slate-900 text-sm">
                {{ formatCurrency(inv.invoice_total) }}
              </td>
              <td class="px-5 py-4 text-right font-bold text-emerald-600">
                {{ formatCurrency(inv.paid_amount || 0) }}
              </td>
              <td class="px-5 py-4 text-right font-black" :class="remainingDebt(inv) > 0 ? 'text-rose-600 text-sm' : 'text-slate-400'">
                {{ formatCurrency(remainingDebt(inv)) }}
              </td>
              <td class="px-5 py-4 text-center">
                <span
                  class="px-2.5 py-1 rounded-full text-[10px] font-black uppercase tracking-wider"
                  :class="paymentBadgeClass(inv)"
                >
                  {{ paymentBadgeLabel(inv) }}
                </span>
              </td>
              <td class="px-5 py-4 text-right">
                <div class="flex items-center justify-end gap-1.5">
                  <button
                    @click="openDetail(inv)"
                    class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all"
                    title="Rincian Barang & Riwayat Pembayaran"
                  >
                    <Eye class="w-4 h-4" />
                  </button>
                  <button
                    v-if="remainingDebt(inv) > 0"
                    @click="openPaymentModal(inv)"
                    class="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition-all shadow-sm flex items-center gap-1"
                  >
                    <DollarSign class="w-3.5 h-3.5" />
                    <span>Bayar</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
          <tfoot class="bg-slate-50 border-t-2 border-slate-200 text-xs font-bold text-slate-900">
            <tr>
              <td colspan="4" class="px-5 py-4 font-black">TOTAL KESELURUHAN ({{ filteredInvoices.length }} Faktur)</td>
              <td class="px-5 py-4 text-right font-black text-sm">{{ formatCurrency(filteredSummary.totalInvoice) }}</td>
              <td class="px-5 py-4 text-right font-black text-emerald-700">{{ formatCurrency(filteredSummary.totalPaid) }}</td>
              <td class="px-5 py-4 text-right font-black text-rose-700 text-sm">{{ formatCurrency(filteredSummary.totalRemaining) }}</td>
              <td colspan="2"></td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>

    <!-- Modal: Input Pembayaran Hutang Supplier -->
    <Teleport to="body">
      <div v-if="showPayModal && activeInvoice" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showPayModal = false" />
        <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-md overflow-hidden">
          <div class="p-6 border-b border-slate-100 bg-slate-50 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <div class="p-2 bg-emerald-100 text-emerald-700 rounded-xl">
                <DollarSign class="w-5 h-5" />
              </div>
              <div>
                <h3 class="font-black text-slate-900 text-base">Pembayaran Faktur Supplier</h3>
                <p class="text-xs text-slate-500 font-mono">{{ activeInvoice.supplier_invoice_no || activeInvoice.receipt_no }}</p>
              </div>
            </div>
            <button @click="showPayModal = false" class="p-1.5 text-slate-400 hover:text-slate-600 rounded-xl">
              <X class="w-5 h-5" />
            </button>
          </div>

          <form @submit.prevent="submitPayment" class="p-6 space-y-4">
            <div class="bg-slate-50 p-4 rounded-2xl space-y-1.5 border border-slate-100">
              <div class="flex justify-between text-xs text-slate-500">
                <span>Supplier:</span>
                <span class="font-bold text-slate-900">{{ activeInvoice.supplier?.name || activeInvoice.supplier_name || '-' }}</span>
              </div>
              <div class="flex justify-between text-xs text-slate-500">
                <span>Total Nilai Faktur:</span>
                <span class="font-bold text-slate-900">{{ formatCurrency(activeInvoice.invoice_total) }}</span>
              </div>
              <div class="flex justify-between text-xs text-slate-500">
                <span>Sudah Dibayar:</span>
                <span class="font-bold text-emerald-600">{{ formatCurrency(activeInvoice.paid_amount || 0) }}</span>
              </div>
              <div class="flex justify-between text-xs text-slate-700 border-t border-slate-200 pt-1.5 font-bold">
                <span>Sisa Hutang:</span>
                <span class="text-rose-600 font-black">{{ formatCurrency(remainingDebt(activeInvoice)) }}</span>
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-600">Nominal Pembayaran (Rp) <span class="text-rose-500">*</span></label>
              <input
                v-model.number="paymentForm.amount"
                type="number"
                min="1"
                :max="remainingDebt(activeInvoice)"
                required
                placeholder="Masukkan nominal"
                class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-3 text-sm font-bold text-slate-900 outline-none focus:ring-2 focus:ring-emerald-500/20"
              />
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-600">Tanggal Pembayaran <span class="text-rose-500">*</span></label>
              <input
                v-model="paymentForm.payment_date"
                type="date"
                required
                class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-2.5 text-sm font-semibold text-slate-800 outline-none focus:ring-2 focus:ring-emerald-500/20"
              />
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-600">No. Referensi / Bukti Transfer</label>
              <input
                v-model="paymentForm.reference_no"
                type="text"
                placeholder="Contoh: TRF-BCA-981293"
                class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-2.5 text-sm font-medium text-slate-800 outline-none focus:ring-2 focus:ring-emerald-500/20"
              />
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-600">Catatan Pembayaran</label>
              <textarea
                v-model="paymentForm.notes"
                rows="2"
                placeholder="Keterangan tambahan..."
                class="w-full bg-slate-50 border border-slate-200 rounded-xl px-4 py-2.5 text-sm font-medium text-slate-800 outline-none focus:ring-2 focus:ring-emerald-500/20 resize-none"
              ></textarea>
            </div>

            <div class="pt-2 flex gap-3">
              <button
                type="button"
                @click="showPayModal = false"
                class="flex-1 py-3 border border-slate-200 rounded-xl text-xs font-bold text-slate-600 hover:bg-slate-50 transition-all"
              >
                Batal
              </button>
              <button
                type="submit"
                :disabled="isSubmittingPayment || paymentForm.amount <= 0"
                class="flex-1 py-3 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition-all shadow-md disabled:opacity-50"
              >
                <span v-if="isSubmittingPayment">Menyimpan...</span>
                <span v-else>Konfirmasi Bayar</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Modal: Detail Faktur & Riwayat Pembayaran -->
    <Teleport to="body">
      <div v-if="showDetailModal && selectedInvoice" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showDetailModal = false" />
        <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <div class="sticky top-0 bg-white px-8 pt-8 pb-4 border-b border-slate-100 flex items-center justify-between z-10">
            <div>
              <h2 class="text-xl font-black text-slate-900">Rincian Faktur Pembelian</h2>
              <p class="text-xs text-slate-500 font-mono mt-0.5">{{ selectedInvoice.supplier_invoice_no || selectedInvoice.receipt_no }}</p>
            </div>
            <button @click="showDetailModal = false" class="p-2 text-slate-400 hover:text-slate-600 rounded-xl">
              <X class="w-5 h-5" />
            </button>
          </div>

          <div class="px-8 py-6 space-y-6">
            <!-- Summary Info Grid -->
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-4 bg-slate-50 p-4 rounded-2xl border border-slate-100">
              <div>
                <p class="text-[10px] font-bold text-slate-400 uppercase">Supplier</p>
                <p class="font-bold text-slate-900 text-xs mt-0.5">{{ selectedInvoice.supplier?.name || selectedInvoice.supplier_name || '-' }}</p>
              </div>
              <div>
                <p class="text-[10px] font-bold text-slate-400 uppercase">Tgl Faktur</p>
                <p class="font-bold text-slate-900 text-xs mt-0.5">{{ formatDate(selectedInvoice.invoice_date || selectedInvoice.created_at) }}</p>
              </div>
              <div>
                <p class="text-[10px] font-bold text-slate-400 uppercase">Jatuh Tempo</p>
                <p class="font-bold text-xs mt-0.5" :class="isOverdue(selectedInvoice) ? 'text-rose-600' : 'text-slate-900'">
                  {{ selectedInvoice.payment_due_date ? formatDate(selectedInvoice.payment_due_date) : '-' }}
                </p>
              </div>
              <div>
                <p class="text-[10px] font-bold text-slate-400 uppercase">Total Tagihan</p>
                <p class="font-black text-slate-900 text-sm mt-0.5">{{ formatCurrency(selectedInvoice.invoice_total) }}</p>
              </div>
              <div>
                <p class="text-[10px] font-bold text-slate-400 uppercase">Sudah Dibayar</p>
                <p class="font-black text-emerald-600 text-sm mt-0.5">{{ formatCurrency(selectedInvoice.paid_amount || 0) }}</p>
              </div>
              <div>
                <p class="text-[10px] font-bold text-slate-400 uppercase">Sisa Hutang</p>
                <p class="font-black text-rose-600 text-sm mt-0.5">{{ formatCurrency(remainingDebt(selectedInvoice)) }}</p>
              </div>
            </div>

            <!-- Items Table -->
            <div class="space-y-2">
              <h4 class="text-xs font-black uppercase tracking-wider text-slate-400">Rincian Barang yang Dibeli</h4>
              <div class="border border-slate-100 rounded-2xl overflow-hidden">
                <table class="w-full text-left text-xs">
                  <thead class="bg-slate-100/70 text-slate-600 font-bold uppercase text-[10px]">
                    <tr>
                      <th class="px-4 py-2.5">Produk / Barang</th>
                      <th class="px-4 py-2.5 text-right">Qty</th>
                      <th class="px-4 py-2.5 text-right">Harga Satuan</th>
                      <th class="px-4 py-2.5 text-right">Subtotal</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100">
                    <tr v-for="it in (selectedInvoice.items || [])" :key="it.id">
                      <td class="px-4 py-2.5 font-semibold text-slate-800">
                        {{ it.product?.name || it.product_name || 'Item' }}
                        <span v-if="it.unit" class="text-slate-400 font-normal">({{ it.unit }})</span>
                      </td>
                      <td class="px-4 py-2.5 text-right font-mono font-bold text-slate-900">{{ it.received_qty }}</td>
                      <td class="px-4 py-2.5 text-right font-mono text-slate-600">{{ formatCurrency(it.unit_price) }}</td>
                      <td class="px-4 py-2.5 text-right font-mono font-bold text-slate-900">{{ formatCurrency(it.received_qty * it.unit_price) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Payments History -->
            <div class="space-y-2">
              <h4 class="text-xs font-black uppercase tracking-wider text-slate-400">Riwayat Pembayaran ke Supplier</h4>
              <div v-if="!selectedInvoice.payments || selectedInvoice.payments.length === 0" class="p-4 bg-slate-50 rounded-2xl text-center text-xs text-slate-400">
                Belum ada catatan pembayaran untuk faktur ini.
              </div>
              <div v-else class="border border-slate-100 rounded-2xl overflow-hidden">
                <table class="w-full text-left text-xs">
                  <thead class="bg-slate-100/70 text-slate-600 font-bold uppercase text-[10px]">
                    <tr>
                      <th class="px-4 py-2.5">Tanggal</th>
                      <th class="px-4 py-2.5">No. Referensi</th>
                      <th class="px-4 py-2.5">Catatan</th>
                      <th class="px-4 py-2.5 text-right">Jumlah Dibayar</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100">
                    <tr v-for="p in selectedInvoice.payments" :key="p.id">
                      <td class="px-4 py-2.5 font-semibold text-slate-800">{{ formatDate(p.payment_date || p.created_at) }}</td>
                      <td class="px-4 py-2.5 font-mono text-slate-600">{{ p.reference_no || '-' }}</td>
                      <td class="px-4 py-2.5 text-slate-500">{{ p.notes || '-' }}</td>
                      <td class="px-4 py-2.5 text-right font-mono font-bold text-emerald-600">{{ formatCurrency(p.amount) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Proof Document Link -->
            <div v-if="selectedInvoice.proof_image_url" class="p-4 bg-teal-50 border border-teal-100 rounded-2xl flex items-center justify-between">
              <div class="flex items-center gap-2 text-teal-800 text-xs font-bold">
                <FileText class="w-4 h-4 text-teal-600" />
                <span>Dokumen / Lampiran Foto Faktur Fisik</span>
              </div>
              <a
                :href="selectedInvoice.proof_image_url"
                target="_blank"
                class="text-xs font-bold text-teal-700 underline flex items-center gap-1"
              >
                <span>Buka Dokumen</span>
                <ExternalLink class="w-3.5 h-3.5" />
              </a>
            </div>
          </div>

          <div class="px-8 py-4 border-t border-slate-100 bg-slate-50 flex justify-end">
            <button
              @click="showDetailModal = false"
              class="px-5 py-2.5 bg-slate-800 hover:bg-slate-900 text-white rounded-xl text-xs font-bold transition-all"
            >
              Tutup
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  Receipt,
  FileText,
  CheckCircle2,
  TrendingDown,
  Clock,
  Search,
  RefreshCw,
  Eye,
  DollarSign,
  Building2,
  ExternalLink,
  X,
  Loader2
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterStore = useMasterDataStore()

const isLoading = ref(false)
const isSubmittingPayment = ref(false)
const selectedBranch = ref('ALL')
const searchQuery = ref('')
const filterStatus = ref('ALL')

const invoices = ref<any[]>([])
const activeInvoice = ref<any>(null)
const showPayModal = ref(false)
const showDetailModal = ref(false)
const selectedInvoice = ref<any>(null)

const paymentForm = ref({
  amount: 0,
  payment_date: new Date().toISOString().split('T')[0],
  reference_no: '',
  notes: '',
})

const statusOptions = [
  { key: 'ALL', label: 'Semua Faktur' },
  { key: 'UNPAID', label: 'Belum Lunas' },
  { key: 'PARTIAL', label: 'Sebagian' },
  { key: 'PAID', label: 'Lunas' },
  { key: 'OVERDUE', label: 'Jatuh Tempo' },
]

const formatCurrency = (val?: number) => {
  if (!val) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

const remainingDebt = (inv: any) => {
  const total = inv.invoice_total || 0
  const paid = inv.paid_amount || 0
  return Math.max(0, total - paid)
}

const isOverdue = (inv: any) => {
  if (!inv.payment_due_date) return false
  if (remainingDebt(inv) <= 0) return false
  const due = new Date(inv.payment_due_date).getTime()
  const today = new Date().setHours(0, 0, 0, 0)
  return due < today
}

const paymentBadgeLabel = (inv: any) => {
  const rem = remainingDebt(inv)
  if (rem <= 0) return 'Lunas'
  if (isOverdue(inv)) return 'Jatuh Tempo'
  if ((inv.paid_amount || 0) > 0) return 'Sebagian'
  return 'Belum Lunas'
}

const paymentBadgeClass = (inv: any) => {
  const rem = remainingDebt(inv)
  if (rem <= 0) return 'bg-emerald-50 text-emerald-700 border border-emerald-200'
  if (isOverdue(inv)) return 'bg-rose-50 text-rose-700 border border-rose-200 animate-pulse'
  if ((inv.paid_amount || 0) > 0) return 'bg-blue-50 text-blue-700 border border-blue-200'
  return 'bg-amber-50 text-amber-700 border border-amber-200'
}

const fetchInvoices = async () => {
  isLoading.value = true
  try {
    const branchQuery = selectedBranch.value !== 'ALL' ? `?branch_id=${selectedBranch.value}` : ''
    const res = await apiClient.get(`/warehouse/supplier-receipts${branchQuery}`)
    invoices.value = res.data?.data || []
  } catch {
    toast.error('Gagal memuat data faktur pembelian')
  } finally {
    isLoading.value = false
  }
}

const summary = computed(() => {
  let totalInvoice = 0
  let totalPaid = 0
  let unpaidCount = 0
  let paidCount = 0
  let overdueCount = 0

  for (const inv of invoices.value) {
    const tot = inv.invoice_total || 0
    const pd = inv.paid_amount || 0
    totalInvoice += tot
    totalPaid += pd

    if (tot > pd) {
      unpaidCount++
      if (isOverdue(inv)) overdueCount++
    } else if (tot > 0 && tot <= pd) {
      paidCount++
    }
  }

  return {
    totalInvoice,
    totalPaid,
    totalRemaining: Math.max(0, totalInvoice - totalPaid),
    unpaidCount,
    paidCount,
    overdueCount,
  }
})

const filteredInvoices = computed(() => {
  return invoices.value.filter((inv) => {
    const rem = remainingDebt(inv)
    // Filter status
    if (filterStatus.value === 'UNPAID' && rem <= 0) return false
    if (filterStatus.value === 'PARTIAL' && (rem <= 0 || (inv.paid_amount || 0) === 0)) return false
    if (filterStatus.value === 'PAID' && rem > 0) return false
    if (filterStatus.value === 'OVERDUE' && !isOverdue(inv)) return false

    // Search query
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const matchInvNo = (inv.supplier_invoice_no || '').toLowerCase().includes(q)
      const matchReceiptNo = (inv.receipt_no || '').toLowerCase().includes(q)
      const matchSupplier = (inv.supplier?.name || inv.supplier_name || '').toLowerCase().includes(q)
      if (!matchInvNo && !matchReceiptNo && !matchSupplier) return false
    }

    return true
  })
})

const filteredSummary = computed(() => {
  let totalInvoice = 0
  let totalPaid = 0
  for (const inv of filteredInvoices.value) {
    totalInvoice += inv.invoice_total || 0
    totalPaid += inv.paid_amount || 0
  }
  return {
    totalInvoice,
    totalPaid,
    totalRemaining: Math.max(0, totalInvoice - totalPaid),
  }
})

const openPaymentModal = (inv: any) => {
  activeInvoice.value = inv
  paymentForm.value = {
    amount: remainingDebt(inv),
    payment_date: new Date().toISOString().split('T')[0],
    reference_no: '',
    notes: '',
  }
  showPayModal.value = true
}

const submitPayment = async () => {
  if (!activeInvoice.value) return
  isSubmittingPayment.value = true
  try {
    await apiClient.post(`/warehouse/supplier-receipts/${activeInvoice.value.id}/pay`, paymentForm.value)
    toast.success('Pembayaran faktur supplier berhasil dicatat')
    showPayModal.value = false
    await fetchInvoices()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal mencatat pembayaran')
  } finally {
    isSubmittingPayment.value = false
  }
}

const openDetail = async (inv: any) => {
  try {
    const res = await apiClient.get(`/warehouse/supplier-receipts/${inv.id}`)
    selectedInvoice.value = res.data?.data || inv
  } catch {
    selectedInvoice.value = inv
  }
  showDetailModal.value = true
}

onMounted(async () => {
  await masterStore.fetchBranches()
  await fetchInvoices()
})
</script>
