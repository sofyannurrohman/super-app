<template>
  <div class="p-6 lg:p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-6 lg:p-8 rounded-[2rem] shadow-sm border border-slate-100">
      <div class="flex items-center gap-5">
        <div class="w-14 h-14 rounded-[1.5rem] bg-gradient-to-br from-teal-500 to-cyan-600 flex items-center justify-center text-white shadow-xl shadow-teal-200">
          <PackageCheck class="h-7 w-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900">Penerimaan Supplier</h1>
          <p class="text-sm text-slate-500 font-medium mt-0.5">Catat kiriman barang dari supplier ke gudang cabang</p>
        </div>
      </div>
      <div class="flex gap-3">
        <button
          @click="activeTab = 'receipts'"
          :class="activeTab === 'receipts' ? 'bg-teal-600 text-white shadow-lg shadow-teal-200' : 'bg-white text-slate-600 border border-slate-200 hover:border-teal-300'"
          class="px-5 py-2.5 rounded-2xl font-bold text-sm transition-all flex items-center gap-2"
        >
          <ClipboardList class="h-4 w-4" /> Daftar Penerimaan
        </button>
        <button
          @click="activeTab = 'stock'"
          :class="activeTab === 'stock' ? 'bg-teal-600 text-white shadow-lg shadow-teal-200' : 'bg-white text-slate-600 border border-slate-200 hover:border-teal-300'"
          class="px-5 py-2.5 rounded-2xl font-bold text-sm transition-all flex items-center gap-2"
        >
          <Boxes class="h-4 w-4" /> Stok Supplier
        </button>
        <button
          v-if="activeTab === 'receipts'"
          @click="openCreateModal"
          class="flex items-center gap-2 bg-gradient-to-r from-teal-500 to-cyan-600 text-white px-6 py-2.5 rounded-2xl font-bold text-sm shadow-lg shadow-teal-200 hover:shadow-xl hover:-translate-y-0.5 transition-all duration-200"
        >
          <Plus class="h-4 w-4" />
          Catat Penerimaan
        </button>
      </div>
    </div>

    <!-- ── Tab: RECEIPTS ── -->
    <template v-if="activeTab === 'receipts'">
      <!-- Search & Filter -->
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari no. penerimaan, supplier, atau invoice..."
            class="w-full pl-11 pr-4 py-3.5 bg-white border border-slate-200 rounded-2xl text-sm font-medium text-slate-700 placeholder-slate-400 outline-none focus:ring-2 focus:ring-teal-500/20 focus:border-teal-300 transition-all"
          />
        </div>
        <select
          v-model="filterStatus"
          class="bg-white border border-slate-200 rounded-2xl px-4 py-3.5 text-sm font-bold text-slate-700 outline-none focus:ring-2 focus:ring-teal-500/20 focus:border-teal-300 transition-all"
        >
          <option value="">Semua Status</option>
          <option value="PENDING">Menunggu</option>
          <option value="RECEIVED">Diterima</option>
          <option value="PARTIAL">Sebagian</option>
          <option value="REJECTED">Ditolak</option>
        </select>
      </div>

      <!-- Receipts Table -->
      <div class="bg-white rounded-[2rem] border border-slate-100 shadow-sm overflow-hidden">
        <div v-if="isLoading" class="flex items-center justify-center py-20">
          <div class="w-8 h-8 border-4 border-teal-500 border-t-transparent rounded-full animate-spin"></div>
        </div>
        <div v-else-if="filteredReceipts.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
          <PackageCheck class="h-16 w-16 mb-4 opacity-20" />
          <p class="font-bold text-lg">Belum ada penerimaan</p>
          <p class="text-sm">Catat penerimaan barang dari supplier</p>
        </div>
        <table v-else class="w-full">
          <thead>
            <tr class="border-b border-slate-100 bg-slate-50/50">
              <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">No. Penerimaan</th>
              <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden md:table-cell">Supplier</th>
              <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden lg:table-cell">Invoice Supplier</th>
              <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden lg:table-cell">Tgl Kirim</th>
              <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Item</th>
              <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Status</th>
              <th class="px-6 py-4 text-right text-[10px] font-black uppercase tracking-widest text-slate-400">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr
              v-for="receipt in filteredReceipts"
              :key="receipt.id"
              class="hover:bg-slate-50/50 transition-colors group"
            >
              <td class="px-6 py-4">
                <p class="font-black text-sm text-slate-900">{{ receipt.receipt_no }}</p>
                <p class="text-xs text-slate-400 mt-0.5">{{ formatDate(receipt.created_at) }}</p>
              </td>
              <td class="px-6 py-4 hidden md:table-cell">
                <p class="font-bold text-sm text-slate-700">{{ receipt.supplier?.name || '-' }}</p>
              </td>
              <td class="px-6 py-4 hidden lg:table-cell">
                <span class="text-sm text-slate-500 font-medium">{{ receipt.supplier_invoice_no || '-' }}</span>
              </td>
              <td class="px-6 py-4 hidden lg:table-cell">
                <span class="text-sm text-slate-500 font-medium">{{ receipt.delivery_date ? formatDate(receipt.delivery_date) : '-' }}</span>
              </td>
              <td class="px-6 py-4">
                <span class="bg-slate-100 text-slate-700 text-xs font-black px-3 py-1 rounded-full">{{ receipt.total_items || (receipt.items?.length || 0) }} item</span>
              </td>
              <td class="px-6 py-4">
                <span :class="statusColor(receipt.status)" class="inline-flex px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest">
                  {{ statusLabel(receipt.status) }}
                </span>
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button
                    @click="openDetailModal(receipt)"
                    class="p-2 text-slate-400 hover:text-teal-600 hover:bg-teal-50 rounded-xl transition-all"
                    title="Lihat Detail"
                  >
                    <Eye class="h-4 w-4" />
                  </button>
                  <button
                    v-if="(receipt.invoice_total || 0) > (receipt.paid_amount || 0)"
                    @click="openPaymentModal(receipt)"
                    class="px-3 py-2 text-[10px] font-black text-emerald-700 bg-emerald-50 hover:bg-emerald-100 rounded-xl"
                    title="Bayar Hutang Supplier"
                  >Bayar</button>
                  <button
                    v-if="receipt.status === 'PENDING'"
                    @click="openReceiveModal(receipt)"
                    class="px-3 py-1.5 bg-teal-50 text-teal-700 hover:bg-teal-100 rounded-xl text-xs font-bold transition-all"
                  >
                    Terima
                  </button>
                  <button
                    v-if="receipt.status === 'PENDING'"
                    @click="openEditModal(receipt)"
                    class="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-xl transition-all opacity-0 group-hover:opacity-100"
                    title="Edit"
                  >
                    <Pencil class="h-4 w-4" />
                  </button>
                  <button
                    v-if="receipt.status === 'PENDING'"
                    @click="handleDelete(receipt)"
                    class="p-2 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-xl transition-all opacity-0 group-hover:opacity-100"
                    title="Hapus"
                  >
                    <Trash2 class="h-4 w-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <!-- ── Tab: STOCK ── -->
    <template v-if="activeTab === 'stock'">
      <div class="bg-white rounded-[2rem] border border-slate-100 shadow-sm overflow-hidden">
        <div v-if="isLoadingStock" class="flex items-center justify-center py-20">
          <div class="w-8 h-8 border-4 border-teal-500 border-t-transparent rounded-full animate-spin"></div>
        </div>
        <div v-else-if="supplierStocks.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
          <Boxes class="h-16 w-16 mb-4 opacity-20" />
          <p class="font-bold text-lg">Belum ada stok supplier</p>
          <p class="text-sm">Stok akan muncul setelah penerimaan dikonfirmasi</p>
        </div>
        <table v-else class="w-full">
          <thead>
            <tr class="border-b border-slate-100 bg-slate-50/50">
              <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Produk</th>
              <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Supplier</th>
              <th class="px-6 py-4 text-right text-[10px] font-black uppercase tracking-widest text-slate-400">Jumlah Stok</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            <tr v-for="stock in supplierStocks" :key="stock.id" class="hover:bg-slate-50/50 transition-colors">
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-teal-100 to-cyan-100 flex items-center justify-center shrink-0">
                    <Package class="h-4 w-4 text-teal-600" />
                  </div>
                  <div>
                    <p class="font-bold text-sm text-slate-900">{{ stock.product?.name || '-' }}</p>
                    <p class="text-xs text-slate-400">SKU: {{ stock.product?.sku || '-' }}</p>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4">
                <p class="font-medium text-sm text-slate-700">{{ stock.supplier?.name || '-' }}</p>
              </td>
              <td class="px-6 py-4 text-right">
                <span class="text-2xl font-black text-teal-700">{{ stock.quantity.toLocaleString() }}</span>
                <span class="text-xs text-slate-400 ml-1">pcs</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <!-- ─────────────────────────────────────────────────────────────────────── -->
    <!-- Modal: Create/Edit Receipt -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showFormModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="closeFormModal" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div class="sticky top-0 bg-white rounded-t-[2rem] px-8 pt-8 pb-6 border-b border-slate-100 z-10">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-xl font-black text-slate-900">{{ isEditing ? 'Edit Penerimaan' : 'Catat Penerimaan Baru' }}</h2>
                  <p class="text-sm text-slate-500 mt-1">Isi detail barang yang diterima dari supplier</p>
                </div>
                <button @click="closeFormModal" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                  <X class="h-5 w-5" />
                </button>
              </div>
            </div>
            <div class="px-8 py-6 space-y-6">
              <!-- Header info -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div class="space-y-1.5">
                  <label class="text-xs font-bold text-slate-500">Supplier <span class="text-rose-500">*</span></label>
                  <select v-model="form.supplier_id" class="input-field">
                    <option value="">Pilih Supplier</option>
                    <option v-for="s in suppliers" :key="s.id" :value="s.id">{{ s.name }}</option>
                  </select>
                </div>
                <div class="space-y-1.5">
                  <label class="text-xs font-bold text-slate-500">No. Invoice Supplier</label>
                  <input v-model="form.supplier_invoice_no" type="text" placeholder="INV-2024-001" class="input-field" />
                </div>
                <div class="space-y-1.5">
                  <label class="text-xs font-bold text-slate-500">Tanggal Faktur</label>
                  <input v-model="form.invoice_date" type="date" class="input-field" />
                </div>
                <div class="space-y-1.5">
                  <label class="text-xs font-bold text-slate-500">Jatuh Tempo Pembayaran</label>
                  <input v-model="form.payment_due_date" type="date" class="input-field" />
                </div>
                <div class="space-y-1.5">
                  <label class="text-xs font-bold text-slate-500">Tanggal Pengiriman</label>
                  <input v-model="form.delivery_date" type="date" class="input-field" />
                </div>
                <div class="space-y-1.5">
                  <label class="text-xs font-bold text-slate-500">Diterima Oleh</label>
                  <input v-model="form.received_by" type="text" placeholder="Nama penerima" class="input-field" />
                </div>
                <div class="sm:col-span-2 space-y-1.5">
                  <label class="text-xs font-bold text-slate-500">URL Foto Bukti / Faktur Supplier</label>
                  <input v-model="form.proof_image_url" type="text" placeholder="https://... atau biarkan kosong" class="input-field" />
                </div>
                <div class="sm:col-span-2 space-y-1.5">
                  <label class="text-xs font-bold text-slate-500">Catatan</label>
                  <textarea v-model="form.notes" rows="2" placeholder="Catatan penerimaan..." class="input-field resize-none"></textarea>
                </div>
              </div>

              <!-- Items -->
              <div>
                <div class="flex items-center justify-between mb-4">
                  <h3 class="text-xs font-black uppercase tracking-widest text-slate-400">Daftar Produk</h3>
                  <button @click="addItem" class="text-xs font-bold text-teal-600 hover:text-teal-700 flex items-center gap-1">
                    <Plus class="h-3 w-3" /> Tambah Produk
                  </button>
                </div>
                <div class="space-y-3">
                  <div
                    v-for="(item, idx) in form.items"
                    :key="idx"
                    class="bg-slate-50 rounded-2xl p-4 space-y-3"
                  >
                    <div class="flex items-center justify-between">
                      <span class="text-xs font-black text-slate-400 uppercase tracking-wider">Produk {{ idx + 1 }}</span>
                      <button @click="removeItem(idx)" class="text-xs text-rose-400 hover:text-rose-600 font-bold">Hapus</button>
                    </div>
                    <div class="grid grid-cols-2 gap-3">
                      <div class="col-span-2 space-y-1.5">
                        <label class="text-xs font-bold text-slate-500">Produk <span class="text-rose-500">*</span></label>
                        <select v-model="item.product_id" class="input-field">
                          <option value="">Pilih Produk</option>
                          <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }} ({{ p.sku }})</option>
                        </select>
                      </div>
                      <div class="space-y-1.5">
                        <label class="text-xs font-bold text-slate-500">Qty Diharapkan</label>
                        <input v-model.number="item.expected_qty" type="number" min="0" class="input-field" />
                      </div>
                      <div class="space-y-1.5">
                        <label class="text-xs font-bold text-slate-500">Satuan</label>
                        <input v-model="item.unit" type="text" placeholder="pcs" class="input-field" />
                      </div>
                      <div class="space-y-1.5">
                        <label class="text-xs font-bold text-slate-500">Harga Satuan (Rp)</label>
                        <input v-model.number="item.unit_price" type="number" min="0" class="input-field" />
                      </div>
                      <div class="space-y-1.5">
                        <label class="text-xs font-bold text-slate-500">Catatan Item</label>
                        <input v-model="item.notes" type="text" placeholder="Opsional" class="input-field" />
                      </div>
                    </div>
                  </div>
                  <button
                    v-if="form.items.length === 0"
                    @click="addItem"
                    class="w-full py-8 border-2 border-dashed border-slate-200 rounded-2xl text-slate-400 hover:border-teal-300 hover:text-teal-500 font-bold text-sm transition-all flex items-center justify-center gap-2"
                  >
                    <Plus class="h-4 w-4" /> Tambahkan produk yang diterima
                  </button>
                </div>
              </div>
            </div>
            <div class="sticky bottom-0 bg-white rounded-b-[2rem] px-8 py-6 border-t border-slate-100 flex gap-3">
              <button @click="closeFormModal" class="flex-1 py-3 border-2 border-slate-200 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-50 transition-all">Batal</button>
              <button
                @click="handleSubmit"
                :disabled="isSubmitting || !form.supplier_id"
                class="flex-1 py-3 bg-gradient-to-r from-teal-500 to-cyan-600 text-white rounded-2xl text-sm font-bold shadow-lg disabled:opacity-50 disabled:cursor-not-allowed hover:shadow-xl hover:-translate-y-0.5 transition-all"
              >
                <span v-if="isSubmitting" class="flex items-center justify-center gap-2">
                  <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                  Menyimpan...
                </span>
                <span v-else>{{ isEditing ? 'Simpan Perubahan' : 'Catat Penerimaan' }}</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Modal: Receive / Confirm -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showReceiveModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showReceiveModal = false" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <div class="px-8 pt-8 pb-6 border-b border-slate-100">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-xl font-black text-slate-900">Konfirmasi Penerimaan</h2>
                  <p class="text-sm text-slate-500 mt-1">{{ receivingReceipt?.receipt_no }}</p>
                </div>
                <button @click="showReceiveModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                  <X class="h-5 w-5" />
                </button>
              </div>
            </div>
            <div class="px-8 py-6 space-y-5">
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-500">Status Penerimaan</label>
                <select v-model="receiveForm.status" class="input-field">
                  <option value="RECEIVED">Semua Diterima</option>
                  <option value="PARTIAL">Sebagian Diterima</option>
                  <option value="REJECTED">Ditolak</option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-500">Diterima Oleh</label>
                <input v-model="receiveForm.received_by" type="text" placeholder="Nama penerima" class="input-field" />
              </div>

              <!-- Per-item quantities for PARTIAL -->
              <div v-if="receiveForm.status === 'PARTIAL' && receivingReceipt?.items?.length" class="space-y-3">
                <p class="text-xs font-black uppercase tracking-widest text-slate-400">Qty Diterima per Produk</p>
                <div v-for="(item, idx) in receiveForm.items" :key="idx" class="bg-slate-50 rounded-2xl p-4 flex items-center gap-4">
                  <div class="flex-1">
                    <p class="font-bold text-sm text-slate-900">{{ receivingReceipt.items[idx]?.product?.name || 'Produk' }}</p>
                    <p class="text-xs text-slate-400">Diharapkan: {{ receivingReceipt.items[idx]?.expected_qty || 0 }} pcs</p>
                  </div>
                  <div class="space-y-1">
                    <label class="text-[10px] font-bold text-slate-400 uppercase">Diterima</label>
                    <input v-model.number="item.received_qty" type="number" min="0" class="w-20 bg-white border-2 border-slate-200 rounded-xl px-3 py-2 text-sm font-bold text-center outline-none focus:border-teal-400 transition-all" />
                  </div>
                </div>
              </div>

              <div class="space-y-1.5">
                <label class="text-xs font-bold text-slate-500">Catatan</label>
                <textarea v-model="receiveForm.notes" rows="2" placeholder="Catatan penerimaan..." class="input-field resize-none"></textarea>
              </div>

              <div v-if="receiveForm.status === 'RECEIVED'" class="bg-teal-50 border border-teal-200 rounded-2xl p-4 flex items-start gap-3">
                <CheckCircle2 class="h-5 w-5 text-teal-600 shrink-0 mt-0.5" />
                <p class="text-sm text-teal-700 font-medium">Stok supplier gudang akan bertambah secara otomatis setelah konfirmasi.</p>
              </div>
            </div>
            <div class="px-8 py-6 border-t border-slate-100 flex gap-3">
              <button @click="showReceiveModal = false" class="flex-1 py-3 border-2 border-slate-200 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-50 transition-all">Batal</button>
              <button
                @click="handleReceive"
                :disabled="isSubmitting"
                class="flex-1 py-3 bg-gradient-to-r from-teal-500 to-cyan-600 text-white rounded-2xl text-sm font-bold shadow-lg disabled:opacity-50 hover:shadow-xl hover:-translate-y-0.5 transition-all"
              >
                <span v-if="isSubmitting" class="flex items-center justify-center gap-2">
                  <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                </span>
                <span v-else>Konfirmasi Penerimaan</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Modal: Detail -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showDetailModal && selectedReceipt" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showDetailModal = false" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto">
            <div class="px-8 pt-8 pb-6 border-b border-slate-100">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-xl font-black text-slate-900">{{ selectedReceipt.receipt_no }}</h2>
                  <span :class="statusColor(selectedReceipt.status)" class="inline-flex px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest mt-2">
                    {{ statusLabel(selectedReceipt.status) }}
                  </span>
                </div>
                <button @click="showDetailModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                  <X class="h-5 w-5" />
                </button>
              </div>
            </div>
            <div class="px-8 py-6 space-y-6">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Supplier</p>
                  <p class="font-bold text-slate-900">{{ selectedReceipt.supplier?.name || '-' }}</p>
                </div>
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Invoice Supplier</p>
                  <p class="font-bold text-slate-900">{{ selectedReceipt.supplier_invoice_no || '-' }}</p>
                </div>
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Tgl Kirim</p>
                  <p class="font-bold text-slate-900">{{ selectedReceipt.delivery_date ? formatDate(selectedReceipt.delivery_date) : '-' }}</p>
                </div>
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Diterima Oleh</p>
                  <p class="font-bold text-slate-900">{{ selectedReceipt.received_by || '-' }}</p>
                </div>
              </div>

              <div v-if="selectedReceipt.notes" class="bg-slate-50 rounded-2xl p-4">
                <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Catatan</p>
                <p class="text-sm text-slate-700">{{ selectedReceipt.notes }}</p>
              </div>

              <div v-if="selectedReceipt.proof_image_url" class="space-y-2">
                <p class="text-xs font-black uppercase tracking-widest text-slate-400">Foto Bukti Terima</p>
                <a :href="selectedReceipt.proof_image_url" target="_blank" class="text-sm text-teal-600 font-bold hover:underline">Lihat Foto →</a>
              </div>

              <!-- Items Table -->
              <div>
                <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-3">Daftar Produk</p>
                <div class="space-y-2">
                  <div
                    v-for="item in selectedReceipt.items"
                    :key="item.id"
                    class="bg-slate-50 rounded-2xl p-4 flex items-center gap-4"
                  >
                    <div class="flex-1">
                      <p class="font-bold text-sm text-slate-900">{{ item.product?.name || 'Produk' }}</p>
                      <p class="text-xs text-slate-400">SKU: {{ item.product?.sku || '-' }}</p>
                    </div>
                    <div class="text-right">
                      <p class="font-black text-slate-900">{{ item.received_qty }} / {{ item.expected_qty }}</p>
                      <p class="text-xs text-slate-400">diterima / diharapkan</p>
                      <p v-if="item.rejected_qty > 0" class="text-xs text-rose-500 font-bold">{{ item.rejected_qty }} ditolak</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <div v-if="showPaymentModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/40" @click="showPaymentModal = false" />
      <form class="relative bg-white rounded-2xl p-6 w-full max-w-md space-y-4 shadow-xl" @submit.prevent="handlePayment">
        <h2 class="text-xl font-black">Bayar Hutang Supplier</h2>
        <p class="text-xs text-slate-500">Sisa: Rp {{ remainingPayable.toLocaleString('id-ID') }}</p>
        <input v-model.number="paymentForm.amount" type="number" min="1" :max="remainingPayable" required class="w-full rounded-xl border p-3" placeholder="Nominal pembayaran" />
        <input v-model="paymentForm.payment_date" type="date" required class="w-full rounded-xl border p-3" />
        <input v-model="paymentForm.reference_no" class="w-full rounded-xl border p-3" placeholder="No. referensi (opsional)" />
        <textarea v-model="paymentForm.notes" class="w-full rounded-xl border p-3" placeholder="Catatan (opsional)" />
        <div class="flex justify-end gap-2"><button type="button" class="px-4 py-2 rounded-xl border" @click="showPaymentModal = false">Batal</button><button class="px-4 py-2 rounded-xl bg-emerald-600 text-white font-bold">Simpan Pembayaran</button></div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  PackageCheck, Plus, Search, Pencil, Trash2, X, Eye, Boxes, Package,
  ClipboardList, CheckCircle2
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { filterSellableProducts } from '@/utils/productUnits'

const masterStore = useMasterDataStore()

const activeTab = ref<'receipts' | 'stock'>('receipts')
const isLoading = ref(false)
const isLoadingStock = ref(false)
const isSubmitting = ref(false)
const receipts = ref<any[]>([])
const suppliers = ref<any[]>([])
const products = ref<any[]>([])
const supplierStocks = ref<any[]>([])
const searchQuery = ref('')
const filterStatus = ref('')

const showFormModal = ref(false)
const showReceiveModal = ref(false)
const showDetailModal = ref(false)
const showPaymentModal = ref(false)
const paymentReceipt = ref<any>(null)
const paymentForm = ref({ amount: 0, payment_date: new Date().toISOString().split('T')[0], reference_no: '', notes: '' })
const remainingPayable = computed(() => Math.max(0, (paymentReceipt.value?.invoice_total || 0) - (paymentReceipt.value?.paid_amount || 0)))
const isEditing = ref(false)
const editingId = ref<string | null>(null)
const receivingReceipt = ref<any>(null)
const selectedReceipt = ref<any>(null)

const defaultFormItem = () => ({
  product_id: '',
  expected_qty: 0,
  received_qty: 0,
  rejected_qty: 0,
  unit: 'pcs',
  unit_price: 0,
  notes: '',
})

const form = ref({
  supplier_id: '',
  receipt_date: new Date().toISOString().split('T')[0],
  branch_id: masterStore.selectedWarehouseBranchId || '',
  company_id: '',
  supplier_invoice_no: '',
  invoice_date: '',
  payment_due_date: '',
  delivery_date: '',
  received_by: '',
  proof_image_url: '',
  notes: '',
  items: [defaultFormItem()],
})

const receiveForm = ref<any>({ status: 'RECEIVED', received_by: '', notes: '', items: [] })

const filteredReceipts = computed(() =>
  receipts.value.filter(r => {
    const matchSearch = !searchQuery.value ||
      r.receipt_no?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      r.supplier?.name?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      r.supplier_invoice_no?.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchStatus = !filterStatus.value || r.status === filterStatus.value
    return matchSearch && matchStatus
  })
)

const statusColor = (s: string) => {
  const m: Record<string, string> = {
    PENDING: 'bg-amber-50 text-amber-700 border border-amber-200',
    RECEIVED: 'bg-emerald-50 text-emerald-700 border border-emerald-200',
    PARTIAL: 'bg-blue-50 text-blue-700 border border-blue-200',
    REJECTED: 'bg-rose-50 text-rose-700 border border-rose-200',
  }
  return m[s] || 'bg-slate-100 text-slate-600'
}

const statusLabel = (s: string) => {
  const m: Record<string, string> = {
    PENDING: 'Menunggu',
    RECEIVED: 'Diterima',
    PARTIAL: 'Sebagian',
    REJECTED: 'Ditolak',
  }
  return m[s] || s
}

const formatDate = (d: string) => {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

const fetchReceipts = async () => {
  isLoading.value = true
  try {
    const branchId = masterStore.selectedWarehouseBranchId
    const url = branchId && branchId !== 'ALL_BRANCHES'
      ? `/warehouse/supplier-receipts?branch_id=${branchId}`
      : '/warehouse/supplier-receipts'
    const res = await apiClient.get(url)
    receipts.value = res.data?.data || []
  } catch {
    toast.error('Gagal memuat data penerimaan')
  } finally {
    isLoading.value = false
  }
}

const fetchSupplierStock = async () => {
  isLoadingStock.value = true
  try {
    const branchId = masterStore.selectedWarehouseBranchId
    const url = branchId && branchId !== 'ALL_BRANCHES'
      ? `/warehouse/supplier-stock?branch_id=${branchId}`
      : '/warehouse/supplier-stock'
    const res = await apiClient.get(url)
    supplierStocks.value = res.data?.data || []
  } catch {
    toast.error('Gagal memuat stok supplier')
  } finally {
    isLoadingStock.value = false
  }
}

const fetchSuppliers = async () => {
  try {
    const res = await apiClient.get('/warehouse/suppliers')
    suppliers.value = res.data?.data || []
  } catch {}
}

const fetchProducts = async () => {
  try {
    const res = await apiClient.get('/factory/products')
    const rawList = res.data?.data || res.data || []
    products.value = filterSellableProducts(rawList)
  } catch {}
}

watch(activeTab, (tab) => {
  if (tab === 'stock') fetchSupplierStock()
})

const addItem = () => form.value.items.push(defaultFormItem())
const removeItem = (idx: number) => form.value.items.splice(idx, 1)

const openCreateModal = () => {
  isEditing.value = false
  editingId.value = null
  form.value = {
    supplier_id: '',
    receipt_date: new Date().toISOString().split('T')[0],
    branch_id: masterStore.selectedWarehouseBranchId || '',
    company_id: '',
    supplier_invoice_no: '',
    invoice_date: '',
    payment_due_date: '',
    delivery_date: '',
    received_by: '',
    proof_image_url: '',
    notes: '',
    items: [defaultFormItem()],
  }
  showFormModal.value = true
}

const openEditModal = (receipt: any) => {
  isEditing.value = true
  editingId.value = receipt.id
  form.value = {
    supplier_id: receipt.supplier_id || '',
    receipt_date: receipt.receipt_date ? new Date(receipt.receipt_date).toISOString().split('T')[0] : new Date().toISOString().split('T')[0],
    branch_id: receipt.branch_id || masterStore.selectedWarehouseBranchId || '',
    company_id: receipt.company_id || '',
    supplier_invoice_no: receipt.supplier_invoice_no || '',
    invoice_date: receipt.invoice_date ? receipt.invoice_date.split('T')[0] : '',
    payment_due_date: receipt.payment_due_date ? receipt.payment_due_date.split('T')[0] : '',
    delivery_date: receipt.delivery_date ? receipt.delivery_date.split('T')[0] : '',
    received_by: receipt.received_by || '',
    proof_image_url: receipt.proof_image_url || '',
    notes: receipt.notes || '',
    items: (receipt.items || []).map((i: any) => ({
      product_id: i.product_id || '',
      expected_qty: i.expected_qty || 0,
      received_qty: i.received_qty || 0,
      rejected_qty: i.rejected_qty || 0,
      unit: i.unit || 'pcs',
      unit_price: i.unit_price || 0,
      notes: i.notes || '',
    })),
  }
  if (form.value.items.length === 0) form.value.items.push(defaultFormItem())
  showFormModal.value = true
}

const openReceiveModal = (receipt: any) => {
  receivingReceipt.value = receipt
  receiveForm.value = {
    status: 'RECEIVED',
    received_by: '',
    notes: '',
    items: (receipt.items || []).map((i: any) => ({
      product_id: i.product_id,
      received_qty: i.expected_qty || 0,
      rejected_qty: 0,
      notes: '',
    })),
  }
  showReceiveModal.value = true
}

const openDetailModal = async (receipt: any) => {
  try {
    const res = await apiClient.get(`/warehouse/supplier-receipts/${receipt.id}`)
    selectedReceipt.value = res.data?.data || receipt
  } catch {
    selectedReceipt.value = receipt
  }
  showDetailModal.value = true
}

const openPaymentModal = (receipt: any) => {
  paymentReceipt.value = receipt
  paymentForm.value = { amount: remainingPayable.value, payment_date: new Date().toISOString().split('T')[0], reference_no: '', notes: '' }
  showPaymentModal.value = true
}

const handlePayment = async () => {
  if (!paymentReceipt.value) return
  try {
    await apiClient.post(`/warehouse/supplier-receipts/${paymentReceipt.value.id}/pay`, paymentForm.value)
    toast.success('Pembayaran supplier berhasil dicatat')
    showPaymentModal.value = false
    await fetchReceipts()
  } catch (err: any) { toast.error(err.response?.data?.error || 'Gagal mencatat pembayaran') }
}

const closeFormModal = () => { showFormModal.value = false }

const handleSubmit = async () => {
  if (!form.value.supplier_id) return
  isSubmitting.value = true
  try {
    const payload = {
      ...form.value,
      invoice_date: form.value.invoice_date ? new Date(form.value.invoice_date).toISOString() : null,
      payment_due_date: form.value.payment_due_date ? new Date(form.value.payment_due_date).toISOString() : null,
      delivery_date: form.value.delivery_date ? new Date(form.value.delivery_date).toISOString() : null,
      branch_id: masterStore.selectedWarehouseBranchId || form.value.branch_id,
    }
    if (isEditing.value && editingId.value) {
      await apiClient.put(`/warehouse/supplier-receipts/${editingId.value}`, payload)
      toast.success('Penerimaan berhasil diperbarui!')
    } else {
      await apiClient.post('/warehouse/supplier-receipts', payload)
      toast.success('Penerimaan berhasil dicatat!')
    }
    closeFormModal()
    await fetchReceipts()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menyimpan penerimaan')
  } finally {
    isSubmitting.value = false
  }
}

const handleReceive = async () => {
  if (!receivingReceipt.value) return
  isSubmitting.value = true
  try {
    await apiClient.post(`/warehouse/supplier-receipts/${receivingReceipt.value.id}/receive`, receiveForm.value)
    toast.success('Penerimaan dikonfirmasi! Stok supplier telah diperbarui.')
    showReceiveModal.value = false
    await fetchReceipts()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal mengkonfirmasi penerimaan')
  } finally {
    isSubmitting.value = false
  }
}

const handleDelete = async (receipt: any) => {
  if (!confirm(`Hapus penerimaan "${receipt.receipt_no}"?`)) return
  try {
    await apiClient.delete(`/warehouse/supplier-receipts/${receipt.id}`)
    toast.success('Penerimaan berhasil dihapus')
    await fetchReceipts()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menghapus penerimaan')
  }
}

onMounted(async () => {
  await masterStore.fetchBranches()
  await Promise.all([fetchReceipts(), fetchSuppliers(), fetchProducts()])
})
</script>

<style scoped>
.input-field {
  width: 100%;
  background: #f8fafc;
  border: 2px solid transparent;
  border-radius: 1rem;
  padding: 0.875rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: #0f172a;
  outline: none;
  transition: all 0.2s;
}
.input-field:focus {
  border-color: rgba(20, 184, 166, 0.3);
  background: white;
}
.modal-enter-active,
.modal-leave-active {
  transition: all 0.25s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
