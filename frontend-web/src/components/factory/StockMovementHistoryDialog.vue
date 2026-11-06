<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[80] flex items-center justify-center p-2 sm:p-4 animate-in fade-in duration-200"
    @click.self="close"
  >
    <Card
      class="w-full max-w-5xl max-h-[92vh] flex flex-col shadow-2xl rounded-[2rem] border-none bg-white overflow-hidden animate-in zoom-in-95 duration-200"
    >
      <!-- Modal Header -->
      <CardHeader class="p-6 border-b border-slate-100 bg-slate-50/50 relative shrink-0">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div class="flex items-start sm:items-center gap-4">
            <div
              class="w-12 h-12 rounded-2xl bg-indigo-50 border border-indigo-100/80 flex items-center justify-center text-indigo-600 shadow-sm shrink-0"
            >
              <Route class="h-6 w-6" />
            </div>
            <div>
              <div class="flex flex-wrap items-center gap-2 mb-1">
                <Badge
                  variant="outline"
                  class="bg-indigo-50/80 text-indigo-700 border-indigo-200 text-[10px] font-bold uppercase tracking-wider px-2 py-0.5"
                >
                  Kartu Stok & Traceability
                </Badge>
                <Badge
                  v-if="stockItem?.factory?.name"
                  variant="outline"
                  class="bg-slate-100 text-slate-700 border-slate-200 text-[10px] font-bold uppercase px-2 py-0.5"
                >
                  <Building2 class="w-3 h-3 mr-1" />
                  {{ stockItem.factory.name }}
                </Badge>
                <Badge
                  v-if="stockItem?.warehouse?.name"
                  variant="outline"
                  class="bg-slate-100 text-slate-700 border-slate-200 text-[10px] font-bold px-2 py-0.5"
                >
                  <Warehouse class="w-3 h-3 mr-1" />
                  {{ stockItem.warehouse.name }}
                </Badge>
              </div>

              <h2 class="text-xl font-black tracking-tight text-slate-900 flex items-center gap-2">
                {{ stockItem?.product?.name || 'Histori Pergerakan Produk' }}
                <span
                  v-if="stockItem?.product?.sku"
                  class="text-xs font-mono font-bold text-slate-400 font-normal"
                >
                  ({{ stockItem.product.sku }})
                </span>
              </h2>

              <div class="flex flex-wrap items-center gap-3 mt-1.5 text-xs text-slate-500">
                <span v-if="stockItem?.batch_no" class="flex items-center gap-1 font-mono font-bold text-slate-700 bg-white px-2 py-0.5 rounded-lg border border-slate-200/60">
                  <Tag class="w-3 h-3 text-indigo-500" />
                  Batch: {{ stockItem.batch_no }}
                </span>
                <span v-if="stockItem?.expiry_date" class="flex items-center gap-1 font-medium">
                  <Calendar class="w-3.5 h-3.5 text-slate-400" />
                  Exp: {{ formatDate(stockItem.expiry_date) }}
                </span>
                <span v-if="stockItem?.product?.unit" class="text-slate-400 font-medium">
                  Satuan: <strong class="text-slate-700">{{ stockItem.product.unit }}</strong>
                </span>
              </div>
            </div>
          </div>

          <!-- Close & Refresh -->
          <div class="flex items-center gap-2 self-end sm:self-center">
            <Button
              variant="outline"
              size="sm"
              @click="loadMovementData"
              class="h-9 rounded-xl border-slate-200 hover:bg-slate-100 text-xs font-bold"
              :disabled="loading"
            >
              <RefreshCw class="h-3.5 w-3.5 mr-1.5" :class="{ 'animate-spin': loading }" />
              Segarkan
            </Button>
            <Button
              variant="ghost"
              size="icon"
              @click="close"
              class="h-9 w-9 rounded-xl text-slate-400 hover:text-slate-900 hover:bg-slate-200/50"
            >
              <X class="h-5 w-5" />
            </Button>
          </div>
        </div>

        <!-- Quick Metrics -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-5">
          <div class="p-3.5 bg-white rounded-2xl border border-slate-100 shadow-sm flex flex-col">
            <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Sisa Stok Pabrik</span>
            <div class="flex items-baseline gap-1 mt-1">
              <span class="text-xl font-black text-slate-900">{{ stockItem?.quantity ?? '-' }}</span>
              <span class="text-[10px] font-bold text-slate-400 uppercase">{{ stockItem?.product?.unit || 'PCS' }}</span>
            </div>
          </div>

          <div class="p-3.5 bg-emerald-50/50 rounded-2xl border border-emerald-100/80 shadow-sm flex flex-col">
            <span class="text-[10px] font-bold text-emerald-700 uppercase tracking-wider flex items-center gap-1">
              <ArrowDownLeft class="w-3 h-3 text-emerald-600" /> Total Diproduksi
            </span>
            <div class="flex items-baseline gap-1 mt-1">
              <span class="text-xl font-black text-emerald-700">{{ totalProduced }}</span>
              <span class="text-[10px] font-bold text-emerald-600 uppercase">{{ stockItem?.product?.unit || 'PCS' }}</span>
            </div>
          </div>

          <div class="p-3.5 bg-blue-50/50 rounded-2xl border border-blue-100/80 shadow-sm flex flex-col">
            <span class="text-[10px] font-bold text-blue-700 uppercase tracking-wider flex items-center gap-1">
              <Truck class="w-3 h-3 text-blue-600" /> Terkirim ke Cabang
            </span>
            <div class="flex items-baseline gap-1 mt-1">
              <span class="text-xl font-black text-blue-700">{{ totalShippedToBranch }}</span>
              <span class="text-[10px] font-bold text-blue-600 uppercase">{{ stockItem?.product?.unit || 'PCS' }}</span>
            </div>
          </div>

          <div class="p-3.5 bg-purple-50/50 rounded-2xl border border-purple-100/80 shadow-sm flex flex-col">
            <span class="text-[10px] font-bold text-purple-700 uppercase tracking-wider flex items-center gap-1">
              <MapPin class="w-3 h-3 text-purple-600" /> Total Cabang Tujuan
            </span>
            <div class="flex items-baseline gap-1 mt-1">
              <span class="text-xl font-black text-purple-700">{{ uniqueBranchesCount }}</span>
              <span class="text-[10px] font-bold text-purple-600 uppercase">Cabang</span>
            </div>
          </div>
        </div>

        <!-- Tab Navigation -->
        <div class="flex items-center gap-2 mt-5 border-b border-slate-200/80 -mb-6">
          <button
            @click="activeTab = 'transfers'"
            class="px-4 py-2.5 text-xs font-bold transition-all relative border-b-2 flex items-center gap-2"
            :class="[
              activeTab === 'transfers'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800',
            ]"
          >
            <Truck class="w-4 h-4" />
            Alur Pengiriman ke Cabang
            <Badge
              variant="outline"
              class="ml-1 text-[10px] px-1.5 py-0 rounded-md font-bold"
              :class="activeTab === 'transfers' ? 'bg-indigo-50 border-indigo-200 text-indigo-700' : 'bg-slate-100 text-slate-600'"
            >
              {{ filteredTransfers.length }}
            </Badge>
          </button>

          <button
            @click="activeTab = 'timeline'"
            class="px-4 py-2.5 text-xs font-bold transition-all relative border-b-2 flex items-center gap-2"
            :class="[
              activeTab === 'timeline'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800',
            ]"
          >
            <History class="w-4 h-4" />
            Kronologi Kartu Stok Lengkap
            <Badge
              variant="outline"
              class="ml-1 text-[10px] px-1.5 py-0 rounded-md font-bold"
              :class="activeTab === 'timeline' ? 'bg-indigo-50 border-indigo-200 text-indigo-700' : 'bg-slate-100 text-slate-600'"
            >
              {{ filteredLogs.length }}
            </Badge>
          </button>

          <button
            @click="activeTab = 'production'"
            class="px-4 py-2.5 text-xs font-bold transition-all relative border-b-2 flex items-center gap-2"
            :class="[
              activeTab === 'production'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-slate-500 hover:text-slate-800',
            ]"
          >
            <Factory class="w-4 h-4" />
            Asal Produksi & Batch
            <Badge
              variant="outline"
              class="ml-1 text-[10px] px-1.5 py-0 rounded-md font-bold"
              :class="activeTab === 'production' ? 'bg-indigo-50 border-indigo-200 text-indigo-700' : 'bg-slate-100 text-slate-600'"
            >
              {{ matchedProductionLogs.length }}
            </Badge>
          </button>
        </div>
      </CardHeader>

      <!-- Modal Body -->
      <CardContent class="p-6 overflow-y-auto custom-scrollbar flex-1 bg-[#f8fafc]/50">
        <!-- Loading State -->
        <div v-if="loading" class="py-16 text-center text-slate-400 flex flex-col items-center justify-center">
          <RefreshCw class="w-8 h-8 animate-spin text-indigo-600 mb-3" />
          <p class="text-sm font-bold text-slate-700">Memuat riwayat pergerakan stok...</p>
          <p class="text-xs text-slate-400 mt-1">Mengambil data pengiriman, mutasi, dan histori produksi</p>
        </div>

        <!-- TAB 1: ALUR PENGIRIMAN KE CABANG (OUTFLOW TRANSFERS) -->
        <div v-else-if="activeTab === 'transfers'" class="space-y-4">
          <!-- Filter & Controls -->
          <div class="flex flex-wrap items-center justify-between gap-3 bg-white p-3.5 rounded-2xl border border-slate-200/80 shadow-sm">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-xs font-bold text-slate-500 mr-1">Filter Status:</span>
              <Button
                size="sm"
                variant="ghost"
                @click="transferStatusFilter = 'ALL'"
                class="h-7 text-xs rounded-lg px-2.5 font-bold"
                :class="transferStatusFilter === 'ALL' ? 'bg-slate-900 text-white hover:bg-slate-800' : 'text-slate-600 hover:bg-slate-100'"
              >
                Semua ({{ transfers.length }})
              </Button>
              <Button
                size="sm"
                variant="ghost"
                @click="transferStatusFilter = 'SHIPPED'"
                class="h-7 text-xs rounded-lg px-2.5 font-bold"
                :class="transferStatusFilter === 'SHIPPED' ? 'bg-amber-500 text-white hover:bg-amber-600' : 'text-amber-700 bg-amber-50 hover:bg-amber-100'"
              >
                Dalam Pengiriman
              </Button>
              <Button
                size="sm"
                variant="ghost"
                @click="transferStatusFilter = 'RECEIVED'"
                class="h-7 text-xs rounded-lg px-2.5 font-bold"
                :class="transferStatusFilter === 'RECEIVED' ? 'bg-emerald-600 text-white hover:bg-emerald-700' : 'text-emerald-700 bg-emerald-50 hover:bg-emerald-100'"
              >
                Diterima di Cabang
              </Button>
              <Button
                size="sm"
                variant="ghost"
                @click="transferStatusFilter = 'APPROVED'"
                class="h-7 text-xs rounded-lg px-2.5 font-bold"
                :class="transferStatusFilter === 'APPROVED' ? 'bg-blue-600 text-white hover:bg-blue-700' : 'text-blue-700 bg-blue-50 hover:bg-blue-100'"
              >
                Disetujui (Siap Kirim)
              </Button>
            </div>

            <!-- Batch Scope Toggle -->
            <div class="flex items-center gap-2">
              <label class="flex items-center gap-1.5 text-xs font-bold text-slate-600 cursor-pointer select-none">
                <input
                  type="checkbox"
                  v-model="filterByCurrentBatchOnly"
                  class="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
                />
                Hanya Batch Saat Ini ({{ stockItem?.batch_no || 'DEFAULT' }})
              </label>
            </div>
          </div>

          <!-- Empty State -->
          <div
            v-if="filteredTransfers.length === 0"
            class="py-16 px-4 text-center bg-white rounded-2xl border border-dashed border-slate-200"
          >
            <Truck class="w-12 h-12 text-slate-300 mx-auto mb-3" />
            <p class="text-sm font-bold text-slate-700">Belum Ada Pengiriman ke Cabang</p>
            <p class="text-xs text-slate-400 mt-1 max-w-md mx-auto">
              Belum ditemukan catatan pengiriman (Surat Jalan / DO) dari pabrik ke cabang untuk produk dan filter yang dipilih.
            </p>
          </div>

          <!-- List of Transfers -->
          <div v-else class="grid grid-cols-1 gap-3.5">
            <div
              v-for="transfer in filteredTransfers"
              :key="transfer.id"
              class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm hover:border-indigo-200 hover:shadow-md transition-all space-y-4"
            >
              <!-- Top bar: DO Number & Status Badge -->
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100">
                <div class="flex flex-wrap items-center gap-2.5">
                  <div class="flex items-center gap-1.5 bg-slate-100 px-2.5 py-1 rounded-xl text-xs font-mono font-black text-slate-800">
                    <FileText class="w-3.5 h-3.5 text-indigo-600" />
                    <span>{{ transfer.delivery_order_no || '-' }}</span>
                    <button
                      v-if="transfer.delivery_order_no"
                      @click="copyToClipboard(transfer.delivery_order_no)"
                      class="text-slate-400 hover:text-slate-700 ml-1"
                      title="Salin No DO"
                    >
                      <Copy class="w-3 h-3" />
                    </button>
                  </div>

                  <Badge
                    :class="getStatusBadgeClass(transfer.status)"
                    class="rounded-xl px-2.5 py-1 text-xs font-black uppercase tracking-wider"
                  >
                    {{ getStatusLabel(transfer.status) }}
                  </Badge>

                  <span class="text-xs text-slate-400 font-medium">
                    Dibuat: {{ formatDateTime(transfer.created_at) }}
                  </span>
                </div>

                <!-- Quantity & Unit -->
                <div class="flex items-center gap-2 self-start sm:self-auto">
                  <span class="text-xs font-bold text-slate-400 uppercase">Jumlah Terkirim:</span>
                  <div class="bg-indigo-50 border border-indigo-100 px-3 py-1 rounded-xl text-indigo-900 font-black text-sm">
                    {{ transfer.quantity }} {{ transfer.unit || 'PCS' }}
                    <span
                      v-if="transfer.pcs_per_unit && transfer.pcs_per_unit > 1"
                      class="text-xs font-bold text-indigo-600 font-normal ml-1"
                    >
                      ({{ transfer.quantity * transfer.pcs_per_unit }} PCS)
                    </span>
                  </div>
                </div>
              </div>

              <!-- Main Details Grid -->
              <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4 text-xs">
                <!-- 1. Route: Factory -> Branch -->
                <div class="space-y-1.5 p-3 rounded-xl bg-slate-50/70 border border-slate-100">
                  <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider flex items-center gap-1">
                    <MapPin class="w-3 h-3 text-indigo-500" /> Rute Pengiriman
                  </span>
                  <div class="space-y-1">
                    <div class="flex items-center gap-1.5 font-bold text-slate-700">
                      <span class="w-2 h-2 rounded-full bg-slate-400 shrink-0"></span>
                      <span class="truncate">Asal: {{ transfer.from_factory?.name || 'Pabrik' }}</span>
                    </div>
                    <div class="flex items-center gap-1.5 font-black text-slate-900">
                      <span class="w-2 h-2 rounded-full bg-emerald-500 shrink-0"></span>
                      <span class="truncate text-emerald-800">Tujuan: {{ transfer.to_branch?.name || 'Cabang' }}</span>
                    </div>
                  </div>
                </div>

                <!-- 2. Logistics: Vehicle & Driver -->
                <div class="space-y-1.5 p-3 rounded-xl bg-slate-50/70 border border-slate-100">
                  <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider flex items-center gap-1">
                    <Truck class="w-3 h-3 text-blue-500" /> Armada & Pengemudi
                  </span>
                  <div class="space-y-1">
                    <div class="flex items-center gap-1.5 font-bold text-slate-800">
                      <Truck class="w-3.5 h-3.5 text-slate-400" />
                      <span>
                        {{ transfer.vehicle?.name || 'Armada Umum' }}
                        <span v-if="transfer.vehicle?.license_plate" class="text-slate-500 font-mono">
                          ({{ transfer.vehicle.license_plate }})
                        </span>
                      </span>
                    </div>
                    <div class="flex items-center gap-1.5 text-slate-700 font-medium">
                      <User class="w-3.5 h-3.5 text-slate-400" />
                      <span>
                        Driver: <strong>{{ getDriverName(transfer) }}</strong>
                      </span>
                    </div>
                  </div>
                </div>

                <!-- 3. Timeline & Dates -->
                <div class="space-y-1.5 p-3 rounded-xl bg-slate-50/70 border border-slate-100 sm:col-span-2 md:col-span-1">
                  <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider flex items-center gap-1">
                    <Clock class="w-3 h-3 text-amber-500" /> Waktu Logistik
                  </span>
                  <div class="space-y-1 text-slate-700 font-medium">
                    <div v-if="transfer.shipped_at" class="flex items-center justify-between">
                      <span class="text-slate-500">Berangkat:</span>
                      <strong class="text-slate-800">{{ formatDateTime(transfer.shipped_at) }}</strong>
                    </div>
                    <div v-if="transfer.received_at" class="flex items-center justify-between">
                      <span class="text-emerald-600 font-bold">Tiba di Cabang:</span>
                      <strong class="text-emerald-700">{{ formatDateTime(transfer.received_at) }}</strong>
                    </div>
                    <div v-if="!transfer.shipped_at && !transfer.received_at" class="text-slate-400 italic">
                      Estimasi / Jadwal Kirim: {{ transfer.target_shipment_date ? formatDate(transfer.target_shipment_date) : 'Segera' }}
                    </div>
                  </div>
                </div>
              </div>

              <!-- Physical Reception Verification Info (if received) -->
              <div
                v-if="transfer.status === 'RECEIVED' || transfer.actual_received_quantity > 0"
                class="p-3 bg-emerald-50/70 rounded-xl border border-emerald-100 text-xs flex flex-wrap items-center justify-between gap-2"
              >
                <div class="flex items-center gap-2 text-emerald-800 font-bold">
                  <CheckCircle2 class="w-4 h-4 text-emerald-600" />
                  <span>Diterima Fisik di Gudang Cabang: {{ transfer.actual_received_quantity }} {{ transfer.unit || 'PCS' }}</span>
                </div>
                <div v-if="transfer.damaged_quantity > 0" class="text-rose-600 font-bold flex items-center gap-1">
                  <AlertCircle class="w-3.5 h-3.5" />
                  <span>Rusak / Selisih: {{ transfer.damaged_quantity }} {{ transfer.unit || 'PCS' }}</span>
                </div>
              </div>

              <!-- Additional details / Approver / Notes -->
              <div v-if="transfer.notes || transfer.approved_by_name" class="text-xs text-slate-500 flex flex-wrap items-center justify-between gap-2 pt-1">
                <span v-if="transfer.approved_by_name" class="flex items-center gap-1">
                  <ShieldCheck class="w-3.5 h-3.5 text-indigo-500" />
                  Disetujui: <strong class="text-slate-700">{{ transfer.approved_by_name }}</strong>
                  <span v-if="transfer.approved_by_position" class="text-slate-400">({{ transfer.approved_by_position }})</span>
                </span>
                <span v-if="transfer.notes" class="italic text-slate-500">
                  Catatan: "{{ transfer.notes }}"
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- TAB 2: KRONOLOGI KARTU STOK LENGKAP (MUTATION TIMELINE) -->
        <div v-else-if="activeTab === 'timeline'" class="space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-3 bg-white p-3.5 rounded-2xl border border-slate-200/80 shadow-sm">
            <div class="flex items-center gap-2">
              <span class="text-xs font-bold text-slate-500">Filter Tipe:</span>
              <Button
                size="sm"
                variant="ghost"
                @click="logTypeFilter = 'ALL'"
                class="h-7 text-xs rounded-lg px-2.5 font-bold"
                :class="logTypeFilter === 'ALL' ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-100'"
              >
                Semua Mutasi
              </Button>
              <Button
                size="sm"
                variant="ghost"
                @click="logTypeFilter = 'IN'"
                class="h-7 text-xs rounded-lg px-2.5 font-bold text-emerald-700"
                :class="logTypeFilter === 'IN' ? 'bg-emerald-600 text-white' : 'bg-emerald-50 hover:bg-emerald-100'"
              >
                Masuk (IN)
              </Button>
              <Button
                size="sm"
                variant="ghost"
                @click="logTypeFilter = 'OUT'"
                class="h-7 text-xs rounded-lg px-2.5 font-bold text-rose-700"
                :class="logTypeFilter === 'OUT' ? 'bg-rose-600 text-white' : 'bg-rose-50 hover:bg-rose-100'"
              >
                Keluar (OUT)
              </Button>
            </div>

            <div class="text-xs text-slate-400 font-bold">
              Menampilkan {{ filteredLogs.length }} catatan mutasi
            </div>
          </div>

          <!-- Empty State -->
          <div
            v-if="filteredLogs.length === 0"
            class="py-16 text-center bg-white rounded-2xl border border-dashed border-slate-200"
          >
            <History class="w-12 h-12 text-slate-300 mx-auto mb-3" />
            <p class="text-sm font-bold text-slate-700">Belum Ada Log Mutasi</p>
            <p class="text-xs text-slate-400 mt-1">Belum ada riwayat pergerakan stok untuk produk ini.</p>
          </div>

          <!-- Timeline Table -->
          <div v-else class="bg-white rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden">
            <table class="w-full text-xs text-left">
              <thead class="bg-slate-50 border-b border-slate-100 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                <tr>
                  <th class="px-5 py-3.5">Waktu</th>
                  <th class="px-5 py-3.5">Tipe</th>
                  <th class="px-5 py-3.5">Aktivitas / Sumber</th>
                  <th class="px-5 py-3.5">Batch No</th>
                  <th class="px-5 py-3.5">Lokasi Pabrik</th>
                  <th class="px-5 py-3.5 text-right">Perubahan Qty</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="log in filteredLogs"
                  :key="log.id"
                  class="hover:bg-slate-50/70 transition-colors"
                >
                  <td class="px-5 py-3.5 font-medium text-slate-700">
                    <div class="font-bold text-slate-900">{{ formatDate(log.created_at) }}</div>
                    <div class="text-[10px] text-slate-400 font-mono">{{ formatTime(log.created_at) }}</div>
                  </td>
                  <td class="px-5 py-3.5">
                    <Badge
                      :class="log.type === 'IN' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-rose-50 text-rose-700 border-rose-200'"
                      variant="outline"
                      class="font-black text-[10px] px-2 py-0.5 rounded-md"
                    >
                      {{ log.type === 'IN' ? 'MASUK (IN)' : 'KELUAR (OUT)' }}
                    </Badge>
                  </td>
                  <td class="px-5 py-3.5">
                    <div class="font-bold text-slate-800">{{ getCleanSource(log.source) }}</div>
                    <div v-if="log.source && getCleanSource(log.source) !== log.source" class="text-[10px] text-slate-400 truncate max-w-xs">{{ log.source }}</div>
                  </td>
                  <td class="px-5 py-3.5 font-mono font-bold text-slate-700">
                    {{ log.batch_no || '-' }}
                  </td>
                  <td class="px-5 py-3.5 font-bold text-slate-700 uppercase">
                    {{ log.factory?.name || 'Pabrik' }}
                  </td>
                  <td class="px-5 py-3.5 text-right">
                    <span
                      class="text-sm font-black"
                      :class="log.type === 'IN' ? 'text-emerald-600' : 'text-rose-600'"
                    >
                      {{ log.type === 'IN' ? '+' : '-' }}{{ log.quantity }}
                    </span>
                    <span class="text-[10px] text-slate-400 font-bold ml-1 uppercase">
                      {{ stockItem?.product?.unit || 'PCS' }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- TAB 3: ASAL PRODUKSI (PRODUCTION ORIGINS) -->
        <div v-else-if="activeTab === 'production'" class="space-y-4">
          <!-- Empty State -->
          <div
            v-if="matchedProductionLogs.length === 0"
            class="py-16 text-center bg-white rounded-2xl border border-dashed border-slate-200"
          >
            <Factory class="w-12 h-12 text-slate-300 mx-auto mb-3" />
            <p class="text-sm font-bold text-slate-700">Belum Ada Riwayat Log Produksi</p>
            <p class="text-xs text-slate-400 mt-1">Data produksi untuk batch ini belum tercatat.</p>
          </div>

          <!-- Production Cards -->
          <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div
              v-for="prod in matchedProductionLogs"
              :key="prod.id"
              class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm space-y-3"
            >
              <div class="flex items-center justify-between pb-3 border-b border-slate-100">
                <span class="text-xs font-bold text-slate-400 uppercase">Tanggal Produksi</span>
                <span class="text-xs font-black text-slate-900">{{ formatDateTime(prod.production_date || prod.created_at) }}</span>
              </div>

              <div class="space-y-2 text-xs">
                <div class="flex justify-between items-center">
                  <span class="text-slate-500 font-medium">Jumlah Diproduksi:</span>
                  <span class="text-base font-black text-emerald-600">+{{ prod.quantity }} PCS</span>
                </div>

                <div v-if="prod.carton_count" class="flex justify-between items-center text-slate-600">
                  <span>Packaging:</span>
                  <span class="font-bold">{{ prod.carton_count }} Karton @ {{ prod.pieces_per_carton }} isi</span>
                </div>

                <div class="flex justify-between items-center text-slate-600">
                  <span>Nomor Batch:</span>
                  <span class="font-mono font-bold text-slate-800">{{ prod.batch_no || '-' }}</span>
                </div>

                <div v-if="prod.expiry_date" class="flex justify-between items-center text-slate-600">
                  <span>Kadaluarsa:</span>
                  <span class="font-bold text-slate-800">{{ formatDate(prod.expiry_date) }}</span>
                </div>

                <div class="flex justify-between items-center text-slate-600">
                  <span>Operator:</span>
                  <span class="font-bold text-indigo-700">
                    {{ prod.employee ? `${prod.employee.first_name || ''} ${prod.employee.last_name || ''}` : 'Operator Pabrik' }}
                  </span>
                </div>

                <!-- Bahan Baku & Batch Terpakai -->
                <div v-if="canViewMaterials && prod.materials && prod.materials.length > 0" class="pt-2.5 border-t border-slate-100 space-y-1.5">
                  <div class="flex items-center gap-1.5 text-[11px] font-bold text-amber-900">
                    <Boxes class="w-3.5 h-3.5 text-amber-600" />
                    <span>Bahan Baku & Batch Terpakai:</span>
                  </div>
                  <div class="flex flex-wrap gap-1.5">
                    <div
                      v-for="(m, mIdx) in prod.materials"
                      :key="mIdx"
                      class="inline-flex items-center gap-1.5 text-[10px] bg-amber-50 text-amber-900 border border-amber-200/80 rounded-lg px-2 py-1"
                    >
                      <span class="font-bold text-slate-800">{{ m.product?.name || 'Bahan Baku' }}</span>
                      <span class="font-mono bg-amber-100 text-amber-800 font-bold px-1.5 py-0.5 rounded text-[9px]">
                        Batch: {{ m.batch_no || 'Auto' }}
                      </span>
                      <span class="font-black text-amber-950">{{ m.quantity }} {{ m.product?.unit || 'KG' }}</span>
                    </div>
                  </div>
                </div>

                <div v-if="prod.notes" class="pt-2 border-t border-slate-100 text-slate-500 italic">
                  Catatan: "{{ prod.notes }}"
                </div>
              </div>
            </div>
          </div>
        </div>
      </CardContent>

      <!-- Modal Footer -->
      <CardFooter class="p-4 sm:p-5 border-t border-slate-100 bg-white flex justify-between items-center shrink-0">
        <div class="text-xs text-slate-400 font-medium hidden sm:block">
          Sistem Pelacakan Mutasi Pabrik Utama SuperApp
        </div>
        <Button
          variant="outline"
          @click="close"
          class="rounded-xl px-6 font-bold border-slate-200 ml-auto"
        >
          Tutup
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { factoryApi } from '@/api/factory'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'

const authStore = useAuthStore()

const canViewMaterials = computed(() => {
  const role = (authStore.user?.role || '').toLowerCase()
  const email = (authStore.user?.email || '').toLowerCase().trim()

  if (['superadmin', 'admin'].includes(role)) return true
  if (role === 'supervisor_gudang') return true
  if (role === 'admin_pabrik' && email === 'rizkikusumaww@gmail.com') return true
  return false
})
import { Card, CardContent, CardHeader, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Route,
  Truck,
  History,
  Factory,
  Building2,
  Warehouse,
  Tag,
  Calendar,
  RefreshCw,
  X,
  FileText,
  Copy,
  MapPin,
  User,
  Clock,
  CheckCircle2,
  AlertCircle,
  ShieldCheck,
  ArrowDownLeft,
} from 'lucide-vue-next'

const props = defineProps<{
  isOpen: boolean
  stockItem: any
}>()

const emit = defineEmits<{
  (e: 'update:isOpen', val: boolean): void
}>()

const activeTab = ref<'transfers' | 'timeline' | 'production'>('transfers')
const loading = ref(false)
const transferStatusFilter = ref<string>('ALL')
const logTypeFilter = ref<string>('ALL')
const filterByCurrentBatchOnly = ref(false)

const transfers = ref<any[]>([])
const inventoryLogs = ref<any[]>([])
const productionLogs = ref<any[]>([])

const close = () => {
  emit('update:isOpen', false)
}

const loadMovementData = async () => {
  if (!props.stockItem?.product_id) return
  loading.value = true

  try {
    const [transfersRes, logsRes, prodRes] = await Promise.all([
      factoryApi.getAllTransfers({
        product_id: props.stockItem.product_id,
      }),
      factoryApi.getAllInventoryLogs(),
      factoryApi.getAllProductionHistory(),
    ])

    transfers.value = transfersRes.data?.data || transfersRes.data || []
    
    // Filter logs for this product
    const allLogs = logsRes.data?.data || logsRes.data || []
    inventoryLogs.value = allLogs.filter((l: any) => l.product_id === props.stockItem.product_id)

    // Filter production logs for this product
    const allProd = prodRes.data?.data || prodRes.data || []
    productionLogs.value = allProd.filter((p: any) => p.product_id === props.stockItem.product_id)
  } catch (err: any) {
    toast.error('Gagal memuat histori pergerakan stok: ' + (err.response?.data?.message || err.message))
  } finally {
    loading.value = false
  }
}

watch(
  () => props.isOpen,
  (val) => {
    if (val) {
      activeTab.value = 'transfers'
      transferStatusFilter.value = 'ALL'
      logTypeFilter.value = 'ALL'
      loadMovementData()
    }
  },
  { immediate: true }
)

// Computed stats
const filteredTransfers = computed(() => {
  return transfers.value.filter((t) => {
    // Status filter
    if (transferStatusFilter.value !== 'ALL' && t.status !== transferStatusFilter.value) {
      return false
    }
    // Batch filter
    if (filterByCurrentBatchOnly.value && props.stockItem?.batch_no) {
      if (t.batch_no !== props.stockItem.batch_no) return false
    }
    return true
  })
})

const filteredLogs = computed(() => {
  return inventoryLogs.value.filter((l) => {
    if (logTypeFilter.value !== 'ALL' && l.type !== logTypeFilter.value) {
      return false
    }
    if (filterByCurrentBatchOnly.value && props.stockItem?.batch_no) {
      if (l.batch_no !== props.stockItem.batch_no) return false
    }
    return true
  })
})

const matchedProductionLogs = computed(() => {
  return productionLogs.value.filter((p) => {
    if (filterByCurrentBatchOnly.value && props.stockItem?.batch_no) {
      return p.batch_no === props.stockItem.batch_no
    }
    return true
  })
})

const totalProduced = computed(() => {
  return matchedProductionLogs.value.reduce((acc, p) => acc + (p.quantity || 0), 0)
})

const totalShippedToBranch = computed(() => {
  return transfers.value
    .filter((t) => ['SHIPPED', 'ARRIVED', 'RECEIVED'].includes(t.status))
    .reduce((acc, t) => {
      const pcs = t.pcs_per_unit && t.pcs_per_unit > 1 ? t.quantity * t.pcs_per_unit : t.quantity
      return acc + pcs
    }, 0)
})

const uniqueBranchesCount = computed(() => {
  const branchIds = new Set(transfers.value.map((t) => t.to_branch_id || t.to_branch?.id).filter(Boolean))
  return branchIds.size
})

// Formatting helpers
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const formatTime = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const getDriverName = (transfer: any) => {
  if (transfer.driver) {
    return `${transfer.driver.first_name || ''} ${transfer.driver.last_name || ''}`.trim()
  }
  return 'Driver Internal'
}

const getCleanSource = (source: string) => {
  if (!source) return 'Mutasi Stok'
  if (source.startsWith('PRODUCTION')) return 'Produksi Pabrik'
  if (source.startsWith('TRANSFER')) return 'Pengiriman ke Cabang'
  if (source.startsWith('ADJUSTMENT')) return 'Koreksi Opname'
  if (source.startsWith('RETURN')) return 'Retur Barang'
  return source
}

const getStatusLabel = (status: string) => {
  switch (status) {
    case 'REQUESTED':
      return 'Menunggu Persetujuan'
    case 'APPROVED':
      return 'Disetujui'
    case 'SHIPPED':
      return 'Dalam Pengiriman'
    case 'ARRIVED':
      return 'Tiba di Tujuan'
    case 'RECEIVED':
      return 'Diterima di Cabang'
    case 'REJECTED':
      return 'Ditolak'
    default:
      return status || 'Draft'
  }
}

const getStatusBadgeClass = (status: string) => {
  switch (status) {
    case 'REQUESTED':
      return 'bg-slate-100 text-slate-700 border-slate-200'
    case 'APPROVED':
      return 'bg-blue-50 text-blue-700 border-blue-200'
    case 'SHIPPED':
      return 'bg-amber-50 text-amber-700 border-amber-200 animate-pulse'
    case 'ARRIVED':
      return 'bg-indigo-50 text-indigo-700 border-indigo-200'
    case 'RECEIVED':
      return 'bg-emerald-50 text-emerald-700 border-emerald-200'
    case 'REJECTED':
      return 'bg-rose-50 text-rose-700 border-rose-200'
    default:
      return 'bg-slate-50 text-slate-600 border-slate-200'
  }
}

const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
  toast.success('Nomor Surat Jalan disalin ke clipboard')
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}
</style>
