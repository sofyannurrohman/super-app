<template>
  <div class="p-6 lg:p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-6 lg:p-8 rounded-[2rem] shadow-sm border border-slate-100">
      <div class="flex items-center gap-5">
        <div class="w-14 h-14 rounded-[1.5rem] bg-gradient-to-br from-violet-600 to-purple-700 flex items-center justify-center text-white shadow-xl shadow-violet-200">
          <Truck class="h-7 w-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-tight text-slate-900">MT Delivery Order</h1>
          <p class="text-sm text-slate-500 font-medium mt-0.5">Buat dan kelola pengiriman ke outlet Modern Trade dari pusat</p>
        </div>
      </div>
      <button
        @click="openCreateModal"
        class="flex items-center gap-2 bg-gradient-to-r from-violet-600 to-purple-700 text-white px-6 py-3 rounded-2xl font-bold text-sm shadow-lg shadow-violet-200 hover:shadow-xl hover:-translate-y-0.5 transition-all duration-200"
      >
        <Plus class="h-4 w-4" />
        Buat DO Baru
      </button>
    </div>

    <!-- Status Filter Tabs -->
    <div class="flex flex-wrap gap-2">
      <button
        v-for="f in statusFilters"
        :key="f.value"
        @click="activeFilter = f.value"
        :class="activeFilter === f.value
          ? 'bg-violet-600 text-white shadow-lg shadow-violet-200'
          : 'bg-white text-slate-600 border border-slate-200 hover:border-violet-300'"
        class="px-4 py-2 rounded-xl font-bold text-xs transition-all"
      >
        {{ f.label }}
        <span v-if="f.value === ''" class="ml-1 bg-white/20 text-white px-1.5 py-0.5 rounded-full text-[10px]" v-show="activeFilter === f.value">{{ orders.length }}</span>
      </button>
    </div>

    <!-- Stats Row -->
    <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
      <div v-for="stat in stats" :key="stat.label" :class="`bg-white p-5 rounded-2xl border shadow-sm ${stat.border}`">
        <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">{{ stat.label }}</p>
        <p :class="`text-3xl font-black mt-1 ${stat.color}`">{{ stat.value }}</p>
      </div>
    </div>

    <!-- Orders Table -->
    <div class="bg-white rounded-[2rem] border border-slate-100 shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-slate-50 flex items-center gap-3">
        <div class="relative flex-1 max-w-sm">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchQuery"
            placeholder="Cari DO, PO, key account..."
            class="w-full pl-10 pr-4 py-2.5 bg-slate-50 border-none rounded-xl text-sm font-medium text-slate-700 placeholder-slate-400 outline-none focus:ring-2 focus:ring-violet-500/20 transition-all"
          />
        </div>
      </div>

      <div v-if="isLoading" class="flex items-center justify-center py-20">
        <div class="w-8 h-8 border-4 border-violet-500 border-t-transparent rounded-full animate-spin"></div>
      </div>
      <div v-else-if="filteredOrders.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
        <Truck class="h-16 w-16 mb-4 opacity-20" />
        <p class="font-bold text-lg">Belum ada MT Delivery Order</p>
        <p class="text-sm">Buat DO baru untuk mulai pengiriman ke outlet MT</p>
      </div>

      <table v-else class="w-full">
        <thead>
          <tr class="border-b border-slate-100 bg-slate-50/50">
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">No. DO</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden md:table-cell">Key Account</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden lg:table-cell">Cabang Sumber</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400 hidden lg:table-cell">Tujuan DC</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Item</th>
            <th class="px-6 py-4 text-left text-[10px] font-black uppercase tracking-widest text-slate-400">Status</th>
            <th class="px-6 py-4 text-right text-[10px] font-black uppercase tracking-widest text-slate-400">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          <tr
            v-for="order in filteredOrders"
            :key="order.id"
            class="hover:bg-slate-50/50 transition-colors group"
          >
            <td class="px-6 py-4">
              <p class="font-black text-sm text-slate-900">{{ order.do_number }}</p>
              <p class="text-xs text-slate-400">{{ formatDate(order.created_at) }}</p>
              <p v-if="order.po_number" class="text-xs text-violet-600 font-bold">PO: {{ order.po_number }}</p>
            </td>
            <td class="px-6 py-4 hidden md:table-cell">
              <p class="font-bold text-sm text-slate-800">{{ order.key_account?.name || '-' }}</p>
              <p class="text-xs text-slate-400">{{ order.key_account?.company_name }}</p>
            </td>
            <td class="px-6 py-4 hidden lg:table-cell">
              <span class="text-sm font-medium text-slate-600">{{ order.source_branch?.name || order.source_factory?.name || '-' }}</span>
            </td>
            <td class="px-6 py-4 hidden lg:table-cell">
              <p class="text-xs text-slate-500 max-w-[140px] truncate">{{ order.dc_destination || '-' }}</p>
            </td>
            <td class="px-6 py-4">
              <span class="bg-slate-100 text-slate-700 text-xs font-black px-3 py-1 rounded-full">{{ order.items?.length || 0 }} item</span>
            </td>
            <td class="px-6 py-4">
              <span :class="statusBadge(order.status)" class="inline-flex px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest">
                {{ statusLabel(order.status) }}
              </span>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex items-center justify-end gap-1.5">
                <button
                  @click="openDetail(order)"
                  class="p-2 text-slate-400 hover:text-violet-600 hover:bg-violet-50 rounded-xl transition-all"
                  title="Detail"
                ><Eye class="h-4 w-4" /></button>

                <button
                  v-if="canEdit(order.status)"
                  @click="openEditModal(order)"
                  class="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-xl transition-all opacity-0 group-hover:opacity-100"
                  title="Edit"
                ><Pencil class="h-4 w-4" /></button>

                <button
                  v-if="order.status === 'DRAFT'"
                  @click="openAssignModal(order)"
                  class="px-3 py-1.5 bg-violet-50 text-violet-700 hover:bg-violet-100 rounded-xl text-xs font-bold transition-all"
                >Assign Cabang</button>

                <button
                  v-if="['DRAFT','ASSIGNED'].includes(order.status)"
                  @click="handleCancel(order)"
                  class="p-2 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-xl transition-all opacity-0 group-hover:opacity-100"
                  title="Batalkan"
                ><XCircle class="h-4 w-4" /></button>

                <button
                  v-if="order.status === 'SHIPPED'"
                  @click="handleMarkDelivered(order)"
                  class="px-3 py-1.5 bg-emerald-50 text-emerald-700 hover:bg-emerald-100 rounded-xl text-xs font-bold transition-all"
                >Tandai Terkirim</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- ─── CREATE / EDIT MODAL ─── -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showFormModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showFormModal = false" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div class="sticky top-0 bg-white rounded-t-[2rem] px-8 pt-8 pb-6 border-b border-slate-100 z-10">
              <div class="flex items-center justify-between">
                <div>
                  <h2 class="text-xl font-black text-slate-900">{{ isEditing ? 'Edit DO' : 'Buat MT Delivery Order Baru' }}</h2>
                  <p class="text-sm text-slate-500 mt-1">DO akan di-assign ke cabang untuk diproses pengirimannya</p>
                </div>
                <button @click="showFormModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                  <X class="h-5 w-5" />
                </button>
              </div>
            </div>
            <div class="px-8 py-6 space-y-6">
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div class="space-y-1.5">
                  <label class="label-field">Key Account (MT) <span class="text-rose-500">*</span></label>
                  <select v-model="form.key_account_id" class="input-field">
                    <option value="">Pilih Key Account</option>
                    <option v-for="ka in keyAccounts" :key="ka.id" :value="ka.id">{{ ka.name }} — {{ ka.company_name }}</option>
                  </select>
                </div>
                <div class="space-y-1.5">
                  <label class="label-field">Nomor PO MT</label>
                  <input v-model="form.po_number" type="text" placeholder="PO-2024-001" class="input-field" />
                </div>
                <div class="space-y-1.5">
                  <label class="label-field">Sumber Barang <span class="text-rose-500">*</span></label>
                  <select v-model="form.source_type" class="input-field">
                    <option value="BRANCH">Dari Gudang Cabang</option>
                    <option value="FACTORY">Dari Pabrik</option>
                    <option value="MIXED">Campuran</option>
                  </select>
                </div>
                <div v-if="form.source_type === 'BRANCH' || form.source_type === 'MIXED'" class="space-y-1.5">
                  <label class="label-field">Cabang Sumber</label>
                  <select v-model="form.source_branch_id" class="input-field">
                    <option value="">Pilih Cabang (opsional, bisa di-assign nanti)</option>
                    <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
                  </select>
                </div>
                <div v-if="form.source_type === 'FACTORY' || form.source_type === 'MIXED'" class="space-y-1.5">
                  <label class="label-field">Gudang Pabrik Sumber</label>
                  <select v-model="form.source_factory_id" class="input-field">
                    <option value="">Pilih Gudang Pabrik</option>
                    <option v-for="f in factories" :key="f.id" :value="f.id">{{ f.name }} {{ f.warehouse_code ? `(${f.warehouse_code})` : '' }}</option>
                  </select>
                </div>
                <div class="space-y-1.5">
                  <label class="label-field">Jadwal Kirim</label>
                  <input v-model="form.scheduled_date" type="date" class="input-field" />
                </div>
                <div class="sm:col-span-2 space-y-1.5">
                  <label class="label-field">Alamat DC / Outlet Tujuan</label>
                  <input v-model="form.dc_destination" type="text" placeholder="Warehouse DC Indomaret Cikampek..." class="input-field" />
                </div>
                <div class="sm:col-span-2 space-y-1.5">
                  <label class="label-field">Catatan</label>
                  <textarea v-model="form.notes" rows="2" placeholder="Instruksi khusus, catatan DO..." class="input-field resize-none"></textarea>
                </div>
              </div>

              <!-- Section: Auto Sales Order -->
              <div class="rounded-2xl border border-violet-200 bg-violet-50 p-5 space-y-4">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="text-sm font-bold text-violet-800">Buat Sales Order Otomatis</p>
                    <p class="text-xs text-violet-500 mt-0.5">SO akan dibuat atas nama Sales MT dan menunggu approval Admin Nota</p>
                  </div>
                  <label class="relative inline-flex items-center cursor-pointer">
                    <input type="checkbox" v-model="form.create_sales_order" class="sr-only peer" />
                    <div class="w-10 h-5 bg-slate-200 peer-focus:outline-none rounded-full peer peer-checked:bg-violet-600 after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:after:translate-x-5"></div>
                  </label>
                </div>
                <div v-if="form.create_sales_order" class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-1">
                  <div class="space-y-1.5">
                    <label class="label-field">Salesman MT <span class="text-rose-500">*</span></label>
                    <select v-model="form.salesman_employee_id" class="input-field">
                      <option value="">Pilih Salesman MT</option>
                      <option v-for="emp in employees" :key="emp.id" :value="emp.id">{{ emp.name }} — {{ emp.branch?.name || emp.branch_id }}</option>
                    </select>
                  </div>
                  <div class="space-y-1.5">
                    <label class="label-field">Cabang Salesman <span class="text-rose-500">*</span></label>
                    <select v-model="form.salesman_branch_id" class="input-field">
                      <option value="">Pilih Cabang</option>
                      <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
                    </select>
                  </div>
                  <div class="sm:col-span-2 space-y-1.5">
                    <label class="label-field">Outlet / Toko MT Tujuan <span class="text-rose-500">*</span></label>
                    <select v-model="form.mt_store_id" class="input-field">
                      <option value="">Pilih Outlet MT</option>
                      <option v-for="s in mtStores" :key="s.id" :value="s.id">{{ s.name }} — {{ s.address }}</option>
                    </select>
                  </div>
                </div>
              </div>

              <!-- Items -->
              <div>
                <div class="flex items-center justify-between mb-4">
                  <h3 class="text-xs font-black uppercase tracking-widest text-slate-400">Daftar Produk</h3>
                  <button @click="addItem" class="text-xs font-bold text-violet-600 hover:text-violet-700 flex items-center gap-1">
                    <Plus class="h-3 w-3" /> Tambah Produk
                  </button>
                </div>
                <div class="space-y-3">
                  <div v-for="(item, idx) in form.items" :key="idx" class="bg-slate-50 rounded-2xl p-4 space-y-3">
                    <div class="flex items-center justify-between">
                      <span class="text-xs font-black text-slate-400 uppercase tracking-wider">Produk {{ idx + 1 }}</span>
                      <button @click="removeItem(idx)" class="text-xs text-rose-400 hover:text-rose-600 font-bold">Hapus</button>
                    </div>
                    <div class="grid grid-cols-2 gap-3">
                      <div class="col-span-2 space-y-1.5">
                        <label class="label-field">Produk <span class="text-rose-500">*</span></label>
                        <select v-model="item.product_id" class="input-field">
                          <option value="">Pilih Produk</option>
                          <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }} ({{ p.sku }})</option>
                        </select>
                      </div>
                      <div class="space-y-1.5">
                        <label class="label-field">Qty Diminta <span class="text-rose-500">*</span></label>
                        <input v-model.number="item.requested_qty" type="number" min="1" class="input-field" />
                      </div>
                      <div class="space-y-1.5">
                        <label class="label-field">Satuan</label>
                        <input v-model="item.unit" type="text" placeholder="pcs" class="input-field" />
                      </div>
                      <div class="space-y-1.5">
                        <label class="label-field">Batch No</label>
                        <input v-model="item.batch_no" type="text" placeholder="DEFAULT" class="input-field" />
                      </div>
                      <div class="space-y-1.5">
                        <label class="label-field">Catatan Item</label>
                        <input v-model="item.notes" type="text" placeholder="Opsional" class="input-field" />
                      </div>
                    </div>
                  </div>
                  <button
                    v-if="form.items.length === 0"
                    @click="addItem"
                    class="w-full py-8 border-2 border-dashed border-slate-200 rounded-2xl text-slate-400 hover:border-violet-300 hover:text-violet-500 font-bold text-sm transition-all flex items-center justify-center gap-2"
                  >
                    <Plus class="h-4 w-4" /> Tambahkan produk
                  </button>
                </div>
              </div>
            </div>
            <div class="sticky bottom-0 bg-white rounded-b-[2rem] px-8 py-6 border-t border-slate-100 flex gap-3">
              <button @click="showFormModal = false" class="flex-1 py-3 border-2 border-slate-200 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-50 transition-all">Batal</button>
              <button
                @click="handleSubmit"
                :disabled="isSubmitting || !form.key_account_id"
                class="flex-1 py-3 bg-gradient-to-r from-violet-600 to-purple-700 text-white rounded-2xl text-sm font-bold shadow-lg disabled:opacity-50 hover:shadow-xl hover:-translate-y-0.5 transition-all"
              >
                <span v-if="isSubmitting" class="flex items-center justify-center gap-2">
                  <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                </span>
                <span v-else>{{ isEditing ? 'Simpan Perubahan' : 'Buat DO' }}</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ─── ASSIGN MODAL ─── -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showAssignModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showAssignModal = false" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-md">
            <div class="px-8 pt-8 pb-6 border-b border-slate-100">
              <h2 class="text-xl font-black text-slate-900">Assign DO ke Cabang</h2>
              <p class="text-sm text-slate-500 mt-1">{{ assigningOrder?.do_number }}</p>
            </div>
            <div class="px-8 py-6 space-y-4">
              <div class="space-y-1.5">
                <label class="label-field">Pilih Cabang Sumber <span class="text-rose-500">*</span></label>
                <select v-model="assignBranchId" class="input-field">
                  <option value="">Pilih Cabang...</option>
                  <option v-for="b in branches" :key="b.id" :value="b.id">{{ b.name }}</option>
                </select>
              </div>
              <div class="bg-amber-50 border border-amber-200 rounded-2xl p-4 flex items-start gap-3">
                <AlertCircle class="h-5 w-5 text-amber-600 shrink-0 mt-0.5" />
                <p class="text-sm text-amber-700 font-medium">Cabang akan mendapatkan notifikasi dan perlu approve DO ini sebelum pengiriman.</p>
              </div>
            </div>
            <div class="px-8 py-6 border-t border-slate-100 flex gap-3">
              <button @click="showAssignModal = false" class="flex-1 py-3 border-2 border-slate-200 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-50 transition-all">Batal</button>
              <button
                @click="handleAssign"
                :disabled="isSubmitting || !assignBranchId"
                class="flex-1 py-3 bg-gradient-to-r from-violet-600 to-purple-700 text-white rounded-2xl text-sm font-bold shadow-lg disabled:opacity-50 hover:shadow-xl transition-all"
              >
                <span v-if="isSubmitting" class="flex items-center justify-center gap-2">
                  <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                </span>
                <span v-else>Assign & Kirim ke Cabang</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ─── DETAIL MODAL ─── -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showDetailModal && selectedOrder" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showDetailModal = false" />
          <div class="relative bg-white rounded-[2rem] shadow-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto">
            <div class="px-8 pt-8 pb-6 border-b border-slate-100 flex items-center justify-between">
              <div>
                <h2 class="text-xl font-black text-slate-900">{{ selectedOrder.do_number }}</h2>
                <div class="flex items-center gap-2 mt-2">
                  <span :class="statusBadge(selectedOrder.status)" class="inline-flex px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-widest">
                    {{ statusLabel(selectedOrder.status) }}
                  </span>
                  <span class="text-xs text-slate-400">Dibuat oleh {{ selectedOrder.created_by || '-' }}</span>
                </div>
              </div>
              <button @click="showDetailModal = false" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-xl transition-all">
                <X class="h-5 w-5" />
              </button>
            </div>
            <div class="px-8 py-6 space-y-6">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Key Account</p>
                  <p class="font-bold text-slate-900">{{ selectedOrder.key_account?.name || '-' }}</p>
                </div>
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">PO Number</p>
                  <p class="font-bold text-slate-900">{{ selectedOrder.po_number || '-' }}</p>
                </div>
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Sumber</p>
                  <p class="font-bold text-slate-900">{{ selectedOrder.source_branch?.name || selectedOrder.source_factory?.name || '-' }}</p>
                </div>
                <div>
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Tujuan DC</p>
                  <p class="font-bold text-slate-900">{{ selectedOrder.dc_destination || '-' }}</p>
                </div>
                <div v-if="selectedOrder.approved_by">
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Disetujui Oleh</p>
                  <p class="font-bold text-slate-900">{{ selectedOrder.approved_by }}</p>
                </div>
                <div v-if="selectedOrder.shipped_at">
                  <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Dikirim Pada</p>
                  <p class="font-bold text-slate-900">{{ formatDate(selectedOrder.shipped_at) }}</p>
                </div>
              </div>

              <div v-if="selectedOrder.notes" class="bg-slate-50 rounded-2xl p-4">
                <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-1">Catatan</p>
                <p class="text-sm text-slate-700">{{ selectedOrder.notes }}</p>
              </div>

              <div v-if="selectedOrder.rejection_reason" class="bg-rose-50 border border-rose-200 rounded-2xl p-4">
                <p class="text-xs font-black uppercase tracking-widest text-rose-400 mb-1">Alasan Penolakan / Pembatalan</p>
                <p class="text-sm text-rose-700">{{ selectedOrder.rejection_reason }}</p>
              </div>

              <div>
                <p class="text-xs font-black uppercase tracking-widest text-slate-400 mb-3">Daftar Produk</p>
                <div class="space-y-2">
                  <div
                    v-for="item in selectedOrder.items"
                    :key="item.id"
                    class="bg-slate-50 rounded-2xl p-4 flex items-center gap-4"
                  >
                    <div class="w-9 h-9 rounded-xl bg-violet-100 flex items-center justify-center shrink-0">
                      <Package class="h-4 w-4 text-violet-600" />
                    </div>
                    <div class="flex-1">
                      <p class="font-bold text-sm text-slate-900">{{ item.product?.name || 'Produk' }}</p>
                      <p class="text-xs text-slate-400">Batch: {{ item.batch_no || 'DEFAULT' }}</p>
                    </div>
                    <div class="text-right">
                      <p class="font-black text-slate-900">{{ item.shipped_qty || item.prepared_qty || item.requested_qty }}</p>
                      <p class="text-xs text-slate-400">
                        <span v-if="item.shipped_qty">dikirim</span>
                        <span v-else-if="item.prepared_qty">disiapkan</span>
                        <span v-else>diminta</span>
                        <span class="ml-1">{{ item.unit }}</span>
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Truck, Plus, Search, Pencil, X, Eye, XCircle, AlertCircle, Package
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { filterSellableProducts } from '@/utils/productUnits'

const masterStore = useMasterDataStore()

const orders = ref<any[]>([])
const keyAccounts = ref<any[]>([])
const branches = ref<any[]>([])
const factories = ref<any[]>([])
const products = ref<any[]>([])
const employees = ref<any[]>([]) // Salesman MT
const mtStores = ref<any[]>([])  // Toko/outlet MT

const isLoading = ref(false)
const isSubmitting = ref(false)
const showFormModal = ref(false)
const showAssignModal = ref(false)
const showDetailModal = ref(false)
const isEditing = ref(false)
const editingId = ref<string | null>(null)
const assigningOrder = ref<any>(null)
const assignBranchId = ref('')
const selectedOrder = ref<any>(null)
const searchQuery = ref('')
const activeFilter = ref('')

const statusFilters = [
  { label: 'Semua', value: '' },
  { label: 'Draft', value: 'DRAFT' },
  { label: 'Assigned', value: 'ASSIGNED' },
  { label: 'Approved', value: 'APPROVED' },
  { label: 'Loading', value: 'LOADING' },
  { label: 'Shipped', value: 'SHIPPED' },
  { label: 'Delivered', value: 'DELIVERED' },
  { label: 'Rejected', value: 'REJECTED' },
  { label: 'Cancelled', value: 'CANCELLED' },
]

const defaultItem = () => ({
  product_id: '',
  requested_qty: 0,
  unit: 'pcs',
  batch_no: 'DEFAULT',
  notes: '',
})

const defaultForm = () => ({
  key_account_id: '',
  po_number: '',
  source_type: 'BRANCH',
  source_branch_id: '',
  source_factory_id: '',
  dc_destination: '',
  scheduled_date: '',
  notes: '',
  items: [defaultItem()],
  // Sales Order otomatis
  create_sales_order: false,
  salesman_employee_id: '' as string,
  salesman_branch_id: '' as string,
  mt_store_id: '' as string,
})

const form = ref(defaultForm())

const filteredOrders = computed(() => {
  return orders.value.filter(o => {
    const matchFilter = !activeFilter.value || o.status === activeFilter.value
    const matchSearch = !searchQuery.value ||
      o.do_number?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      o.po_number?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      o.key_account?.name?.toLowerCase().includes(searchQuery.value.toLowerCase())
    return matchFilter && matchSearch
  })
})

const stats = computed(() => [
  { label: 'Total DO', value: orders.value.length, color: 'text-slate-900', border: 'border-slate-100' },
  { label: 'Menunggu Assign', value: orders.value.filter(o => o.status === 'DRAFT').length, color: 'text-slate-600', border: 'border-slate-100' },
  { label: 'Proses Cabang', value: orders.value.filter(o => ['ASSIGNED','APPROVED','LOADING'].includes(o.status)).length, color: 'text-violet-700', border: 'border-violet-100' },
  { label: 'Dalam Pengiriman', value: orders.value.filter(o => o.status === 'SHIPPED').length, color: 'text-blue-700', border: 'border-blue-100' },
  { label: 'Terkirim', value: orders.value.filter(o => o.status === 'DELIVERED').length, color: 'text-emerald-700', border: 'border-emerald-100' },
])

const statusBadge = (s: string) => {
  const m: Record<string, string> = {
    DRAFT: 'bg-slate-100 text-slate-600 border border-slate-200',
    ASSIGNED: 'bg-amber-50 text-amber-700 border border-amber-200',
    APPROVED: 'bg-blue-50 text-blue-700 border border-blue-200',
    LOADING: 'bg-indigo-50 text-indigo-700 border border-indigo-200',
    SHIPPED: 'bg-purple-50 text-purple-700 border border-purple-200',
    DELIVERED: 'bg-emerald-50 text-emerald-700 border border-emerald-200',
    REJECTED: 'bg-rose-50 text-rose-700 border border-rose-200',
    CANCELLED: 'bg-slate-100 text-slate-500 border border-slate-200',
  }
  return m[s] || 'bg-slate-100 text-slate-600'
}

const statusLabel = (s: string) => {
  const m: Record<string, string> = {
    DRAFT: 'Draft', ASSIGNED: 'Assigned', APPROVED: 'Approved',
    LOADING: 'Loading', SHIPPED: 'Dikirim', DELIVERED: 'Terkirim',
    REJECTED: 'Ditolak', CANCELLED: 'Dibatalkan',
  }
  return m[s] || s
}

const canEdit = (status: string) => ['DRAFT', 'ASSIGNED'].includes(status)
const formatDate = (d: string) => d ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '-'

const fetchAll = async () => {
  isLoading.value = true
  try {
    const [ordersRes, kaRes, branchRes, factoryRes, productRes, empRes, storeRes] = await Promise.all([
      apiClient.get('/mt-delivery'),
      apiClient.get('/modern-trade/key-accounts'),
      apiClient.get('/branches'),
      apiClient.get('/factory'),
      apiClient.get('/factory/products'),
      apiClient.get('/employees').catch(() => ({ data: [] })),
      apiClient.get('/stores').catch(() => ({ data: [] })),
    ])
    orders.value = ordersRes.data?.data || []
    keyAccounts.value = kaRes.data?.data || kaRes.data || []
    branches.value = branchRes.data?.data || branchRes.data || []
    factories.value = factoryRes.data || []
    products.value = filterSellableProducts(productRes.data?.data || productRes.data || [])
    employees.value = empRes.data?.data || empRes.data || []
    mtStores.value = storeRes.data?.data || storeRes.data || []
  } catch {
    toast.error('Gagal memuat data')
  } finally {
    isLoading.value = false
  }
}

const openCreateModal = () => {
  isEditing.value = false
  editingId.value = null
  form.value = defaultForm()
  showFormModal.value = true
}

const openEditModal = (order: any) => {
  isEditing.value = true
  editingId.value = order.id
  form.value = {
    key_account_id: order.key_account_id || '',
    po_number: order.po_number || '',
    source_type: order.source_type || 'BRANCH',
    source_branch_id: order.source_branch_id || '',
    source_factory_id: order.source_factory_id || '',
    dc_destination: order.dc_destination || '',
    scheduled_date: order.scheduled_date ? order.scheduled_date.split('T')[0] : '',
    notes: order.notes || '',
    items: (order.items || []).map((i: any) => ({
      product_id: i.product_id || '',
      requested_qty: i.requested_qty || 0,
      unit: i.unit || 'pcs',
      batch_no: i.batch_no || 'DEFAULT',
      notes: i.notes || '',
    })),
    create_sales_order: order.create_sales_order || false,
    salesman_employee_id: order.salesman_employee_id || '',
    salesman_branch_id: order.salesman_branch_id || '',
    mt_store_id: order.mt_store_id || '',
  }
  if (form.value.items.length === 0) form.value.items.push(defaultItem())
  showFormModal.value = true
}

const openDetail = async (order: any) => {
  try {
    const res = await apiClient.get(`/mt-delivery/${order.id}`)
    selectedOrder.value = res.data?.data || order
  } catch {
    selectedOrder.value = order
  }
  showDetailModal.value = true
}

const openAssignModal = (order: any) => {
  assigningOrder.value = order
  assignBranchId.value = order.source_branch_id || ''
  showAssignModal.value = true
}

const addItem = () => form.value.items.push(defaultItem())
const removeItem = (idx: number) => form.value.items.splice(idx, 1)

const handleSubmit = async () => {
  if (!form.value.key_account_id) return
  // Validasi field salesman jika toggle SO aktif
  if (form.value.create_sales_order) {
    if (!form.value.salesman_employee_id) { toast.error('Pilih Salesman MT'); return }
    if (!form.value.salesman_branch_id) { toast.error('Pilih Cabang Salesman'); return }
    if (!form.value.mt_store_id) { toast.error('Pilih Outlet MT Tujuan'); return }
  }
  isSubmitting.value = true
  try {
    const payload: Record<string, any> = {
      key_account_id: form.value.key_account_id,
      po_number: form.value.po_number,
      source_type: form.value.source_type,
      source_branch_id: form.value.source_branch_id || null,
      source_factory_id: form.value.source_factory_id || null,
      dc_destination: form.value.dc_destination,
      scheduled_date: form.value.scheduled_date ? new Date(form.value.scheduled_date).toISOString() : null,
      notes: form.value.notes,
      items: form.value.items,
      // Field salesman — null jika toggle SO tidak aktif, ID jika aktif
      salesman_employee_id: form.value.create_sales_order ? (form.value.salesman_employee_id || null) : null,
      salesman_branch_id: form.value.create_sales_order ? (form.value.salesman_branch_id || null) : null,
      mt_store_id: form.value.create_sales_order ? (form.value.mt_store_id || null) : null,
    }
    if (isEditing.value && editingId.value) {
      await apiClient.put(`/mt-delivery/${editingId.value}`, payload)
      toast.success('DO berhasil diperbarui!')
    } else {
      await apiClient.post('/mt-delivery', payload)
      toast.success(form.value.create_sales_order ? 'MT DO & Sales Order berhasil dibuat!' : 'MT Delivery Order berhasil dibuat!')
    }
    showFormModal.value = false
    await fetchAll()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal menyimpan DO')
  } finally {
    isSubmitting.value = false
  }
}

const handleAssign = async () => {
  if (!assignBranchId.value || !assigningOrder.value) return
  isSubmitting.value = true
  try {
    await apiClient.post(`/mt-delivery/${assigningOrder.value.id}/assign`, { branch_id: assignBranchId.value })
    toast.success('DO berhasil di-assign ke cabang!')
    showAssignModal.value = false
    await fetchAll()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal assign DO')
  } finally {
    isSubmitting.value = false
  }
}

const handleCancel = async (order: any) => {
  const reason = prompt('Alasan pembatalan (opsional):') ?? ''
  if (reason === null) return
  try {
    await apiClient.post(`/mt-delivery/${order.id}/cancel`, { reason })
    toast.success('DO berhasil dibatalkan')
    await fetchAll()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal membatalkan DO')
  }
}

const handleMarkDelivered = async (order: any) => {
  if (!confirm(`Tandai DO ${order.do_number} sudah diterima MT?`)) return
  try {
    await apiClient.post(`/mt-delivery/${order.id}/delivered`)
    toast.success('DO ditandai sudah terkirim')
    await fetchAll()
  } catch (err: any) {
    toast.error(err.response?.data?.error || 'Gagal update status')
  }
}

onMounted(fetchAll)
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
.input-field:focus { border-color: rgba(139, 92, 246, 0.3); background: white; }
.label-field { display: block; font-size: 0.75rem; font-weight: 700; color: #64748b; margin-bottom: 0.375rem; }
.modal-enter-active, .modal-leave-active { transition: all 0.25s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
