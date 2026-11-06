<template>
  <div class="p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-6 rounded-3xl shadow-sm border border-slate-100"
    >
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900">
          Manajemen Stok & Produksi Global
        </h1>
        <p class="text-sm text-slate-500 mt-0.5">
          Monitoring stok dan hasil produksi di seluruh jaringan pabrik Karsa HR.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="outline" class="rounded-xl border-slate-200" @click="refreshData">
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': factoryStore.loading }" />
          Segarkan
        </Button>
        <Button
          variant="outline"
          class="rounded-xl border-slate-200 hover:bg-slate-50"
          :disabled="syncingWarehouse"
          @click="handleSyncAllWarehouseStock"
        >
          <RefreshCw class="mr-2 h-4 w-4 text-primary" :class="{ 'animate-spin': syncingWarehouse }" />
          Sinkronkan ke Gudang
        </Button>
        <Button
          @click="openLogProduction"
          class="rounded-xl shadow-lg shadow-blue-500/10 font-bold bg-blue-600 hover:bg-blue-700 text-white"
        >
          <Plus class="mr-2 h-4 w-4" /> Log Produksi
        </Button>
        <Button
          @click="router.push('/factory/products')"
          class="rounded-xl shadow-lg shadow-primary/10 font-bold bg-primary hover:bg-primary/90"
        >
          <Package class="mr-2 h-4 w-4" /> Data Produk
        </Button>
      </div>
    </div>

    <!-- Stacked Content -->
    <div class="flex flex-col gap-10">
      <!-- TOP: GLOBAL STOCK -->
      <div class="space-y-6">
        <div class="flex items-center justify-between px-2">
          <div class="flex items-center gap-3">
            <div>
              <h2 class="text-lg font-bold text-slate-900 leading-tight">Manajemen Stok Global</h2>
              <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">
                Seluruh Lokasi Pabrik
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              @click="showInventoryLogs = true"
              class="rounded-xl h-9 mr-1"
            >
              <History class="mr-1 h-3.5 w-3.5" /> Riwayat Opname
            </Button>
            <div class="relative flex items-center">
              <Search class="h-4 w-4 text-slate-400 absolute left-3 pointer-events-none" />
              <input
                v-model="stockSearch"
                placeholder="Cari Produk, SKU, No. Batch..."
                class="text-xs bg-white border border-slate-200/80 rounded-xl pl-9 pr-8 py-2 focus:ring-2 focus:ring-primary/20 focus:border-primary w-48 sm:w-72 outline-none transition-all shadow-sm"
              />
              <button
                v-if="stockSearch"
                @click="clearStockSearch"
                class="absolute right-2.5 text-slate-400 hover:text-slate-600 p-0.5 rounded-full hover:bg-slate-100 transition-colors"
                title="Hapus pencarian"
              >
                <X class="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
        </div>

        <!-- Component Penampilan UI Summary Total Barang dari Data Batch -->
        <StockBatchSummaryCard
          :items="filteredStock"
          :searchQuery="stockSearch"
          :allStockCount="factoryStore.allInventory?.length || 0"
          @select-batch="handleSelectBatch"
          @clear-search="clearStockSearch"
        />

        <!-- Info Baris Total & Status Filter -->
        <div class="flex items-center justify-between px-2 text-xs text-slate-500 font-medium">
          <div class="flex items-center gap-2">
            <span>Menampilkan <strong class="text-slate-800 font-mono">{{ filteredStock.length }}</strong> data batch</span>
            <span class="text-slate-300">•</span>
            <span>Total Kuantitas: <strong class="text-emerald-600 font-mono font-bold">{{ totalFilteredQuantity.toLocaleString('id-ID') }}</strong> unit</span>
          </div>
          <span v-if="stockSearch" class="text-blue-600 text-[11px] font-bold">
            Difilter berdasarkan kata kunci "{{ stockSearch }}"
          </span>
        </div>

        <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
          <DataTable
            :columns="stockColumns"
            :data="filteredStock"
            :isLoading="factoryStore.loading"
            class="border-none"
          >
            <template #cell-factory_name="{ row }">
              <Badge
                variant="outline"
                class="rounded-lg text-[10px] border-slate-100 bg-slate-50 text-slate-500 py-0 px-2 font-bold uppercase"
              >
                {{ (row as any).factory?.name }}
              </Badge>
            </template>
            <template #cell-batch_info="{ row }">
              <div class="flex flex-col gap-0.5">
                <span class="text-[11px] font-black text-slate-700 font-mono tracking-tight">{{
                  (row as any).batch_no || '-'
                }}</span>
                <span class="text-[9px] font-bold text-slate-400">
                  Exp:
                  {{
                    (row as any).expiry_date
                      ? new Date((row as any).expiry_date).toLocaleDateString('id-ID', {
                          day: 'numeric',
                          month: 'short',
                          year: 'numeric',
                        })
                      : '-'
                  }}
                </span>
              </div>
            </template>
            <template #cell-warehouse_name="{ row }">
              <span class="text-[11px] font-bold text-slate-600">{{
                (row as any).warehouse?.name || 'Gudang Umum'
              }}</span>
            </template>
            <template #cell-quantity="{ row }">
              <div class="flex items-center gap-2">
                <span class="text-base font-black text-slate-900">{{ (row as any).quantity }}</span>
                <span class="text-[9px] font-black text-slate-400 uppercase tracking-tighter">{{
                  (row as any).product?.unit
                }}</span>
              </div>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-0.5">
                <Button
                  variant="ghost"
                  size="sm"
                  @click="openStockMovement(row)"
                  title="Histori pergerakan & alur distribusi stok"
                  class="h-8 w-8 p-0 rounded-lg text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 transition-colors"
                >
                  <Route class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  @click="openAdjustStock(row)"
                  title="Koreksi stok"
                  class="h-8 w-8 p-0 rounded-lg text-slate-300 hover:text-primary"
                >
                  <Settings2 class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  @click="openDeleteStock(row)"
                  title="Hapus stok"
                  class="h-8 w-8 p-0 rounded-lg text-slate-300 hover:text-rose-600"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </template>
          </DataTable>
        </Card>
      </div>

      <!-- BOTTOM: GLOBAL PRODUCTION LOGS -->
      <div class="space-y-6">
        <div class="flex items-center justify-between px-2">
          <div class="flex items-center gap-3">
            <div>
              <h2 class="text-lg font-bold text-slate-900 leading-tight">Log Produksi Global</h2>
              <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest mt-0.5">
                Aktivitas Terkini
              </p>
            </div>
          </div>
        </div>

        <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
          <DataTable
            :columns="productionColumns"
            :data="factoryStore.allProductionHistory"
            :isLoading="factoryStore.loading"
            class="border-none"
          >
            <template #cell-factory_name="{ row }">
              <span class="text-[11px] font-bold text-slate-400 uppercase">{{
                (row as any).factory?.name
              }}</span>
            </template>
            <template #cell-batch_info="{ row }">
              <div class="flex flex-col gap-0.5">
                <span class="text-[11px] font-black text-slate-700 font-mono tracking-tight">{{
                  (row as any).batch_no || '-'
                }}</span>
                <span class="text-[9px] font-bold text-slate-400">
                  Exp:
                  {{
                    (row as any).expiry_date
                      ? new Date((row as any).expiry_date).toLocaleDateString('id-ID', {
                          day: 'numeric',
                          month: 'short',
                          year: 'numeric',
                        })
                      : '-'
                  }}
                </span>
              </div>
            </template>
            <template #cell-quantity="{ row }">
              <div class="flex flex-col">
                <span class="font-black text-emerald-600">+{{ (row as any).quantity }} PCS</span>
                <span v-if="(row as any).carton_count" class="text-[9px] font-bold text-slate-400">
                  ({{ (row as any).carton_count }} Karton @
                  {{ (row as any).pieces_per_carton }} isi)
                </span>
              </div>
            </template>
            <template #cell-operators_info="{ row }">
              <div class="flex flex-col gap-1 max-w-[220px]">
                <div v-if="(row as any).operators && (row as any).operators.length > 0" class="flex flex-wrap gap-1">
                  <span
                    v-for="(op, opI) in (row as any).operators"
                    :key="opI"
                    class="text-[10px] font-bold px-2 py-0.5 rounded-lg inline-flex items-center gap-1 leading-tight"
                    :class="op.is_manual ? 'bg-amber-50 text-amber-900 border border-amber-200/70' : 'bg-slate-100 text-slate-700'"
                    :title="op.is_manual ? 'Operator Manual / Harian' : 'Karyawan Terdaftar'"
                  >
                    <span v-if="op.is_manual" class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
                    <span>{{ op.operator_name }}</span>
                  </span>
                </div>
                <span
                  v-else-if="(row as any).operator_name"
                  class="text-[11px] font-bold text-slate-700 line-clamp-2"
                  :title="(row as any).operator_name"
                >
                  {{ (row as any).operator_name }}
                </span>
                <span
                  v-else-if="(row as any).employee"
                  class="text-[11px] font-bold text-slate-700"
                >
                  {{ (row as any).employee.first_name }} {{ (row as any).employee.last_name || '' }}
                </span>
                <span v-else class="text-[10px] text-slate-400 italic">
                  -
                </span>
              </div>
            </template>
            <template #cell-materials="{ row }">
              <div v-if="(row as any).materials && (row as any).materials.length > 0">
                <Button
                  variant="outline"
                  size="sm"
                  @click="openMaterialsDetail(row)"
                  class="h-7 px-2.5 text-xs font-bold rounded-xl border-amber-200 bg-amber-50/80 text-amber-900 hover:bg-amber-100 hover:border-amber-300 shadow-2xs flex items-center gap-1.5 transition-all group"
                  title="Klik untuk melihat rincian bahan baku & batch terpakai"
                >
                  <Boxes class="h-3.5 w-3.5 text-amber-600 group-hover:scale-110 transition-transform" />
                  <span>{{ (row as any).materials.length }} Bahan Baku</span>
                </Button>
              </div>
              <span v-else class="text-[10px] text-slate-400 italic">
                -
              </span>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-0.5">
                <Button
                  variant="ghost"
                  size="icon"
                  @click="openStockMovement(row)"
                  title="Lihat Histori Pergerakan & Pengiriman Cabang"
                  class="h-7 w-7 text-slate-300 hover:text-indigo-600 hover:bg-indigo-50"
                >
                  <Route class="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  @click="openEditProduction(row)"
                  class="h-7 w-7 text-slate-300 hover:text-primary"
                >
                  <Edit2 class="h-3 w-3" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  @click="handleDeleteProduction(row)"
                  class="h-7 w-7 text-slate-300 hover:text-destructive"
                >
                  <Trash2 class="h-3 w-3" />
                </Button>
              </div>
            </template>
          </DataTable>
        </Card>
      </div>
    </div>
    <!-- Log Mutasi Modal -->
    <div
      v-if="showInventoryLogs"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-[60] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-4xl shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200"
      >
        <CardHeader class="p-6 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-xl font-bold">Log Mutasi & Riwayat Opname</CardTitle>
              <CardDescription
                >Catatan seluruh perubahan stok, termasuk hasil stok opname
                (ADJUSTMENT).</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showInventoryLogs = false"
              class="rounded-full"
            >
              <X class="h-5 w-5 text-slate-400" />
            </Button>
          </div>

          <!-- Filter categories -->
          <div class="flex flex-wrap items-center gap-2 mt-4">
            <Button
              size="sm"
              variant="ghost"
              @click="inventoryLogFilter = 'ALL'"
              class="h-7 text-xs rounded-lg px-2.5 font-bold"
              :class="inventoryLogFilter === 'ALL' ? 'bg-slate-900 text-white' : 'text-slate-600 bg-slate-100 hover:bg-slate-200'"
            >
              Semua Mutasi
            </Button>
            <Button
              size="sm"
              variant="ghost"
              @click="inventoryLogFilter = 'ADJUSTMENT'"
              class="h-7 text-xs rounded-lg px-2.5 font-bold"
              :class="inventoryLogFilter === 'ADJUSTMENT' ? 'bg-orange-500 text-white' : 'text-orange-700 bg-orange-50 hover:bg-orange-100'"
            >
              Opname & Koreksi
            </Button>
            <Button
              size="sm"
              variant="ghost"
              @click="inventoryLogFilter = 'PRODUCTION'"
              class="h-7 text-xs rounded-lg px-2.5 font-bold"
              :class="inventoryLogFilter === 'PRODUCTION' ? 'bg-blue-600 text-white' : 'text-blue-700 bg-blue-50 hover:bg-blue-100'"
            >
              Hasil Produksi
            </Button>
            <Button
              size="sm"
              variant="ghost"
              @click="inventoryLogFilter = 'TRANSFER'"
              class="h-7 text-xs rounded-lg px-2.5 font-bold"
              :class="inventoryLogFilter === 'TRANSFER' ? 'bg-indigo-600 text-white' : 'text-indigo-700 bg-indigo-50 hover:bg-indigo-100'"
            >
              Pengiriman Cabang
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-0 bg-white max-h-[60vh] overflow-y-auto">
          <div
            v-if="filteredGeneralInventoryLogs.length === 0"
            class="p-12 text-center text-slate-400 font-medium"
          >
            Belum ada log mutasi untuk filter yang dipilih.
          </div>
          <table v-else class="w-full text-sm text-left">
            <thead
              class="bg-slate-50 sticky top-0 border-b border-slate-100 text-[10px] font-black text-slate-400 uppercase tracking-widest"
            >
              <tr>
                <th class="px-6 py-4">Waktu</th>
                <th class="px-6 py-4">Lokasi Pabrik</th>
                <th class="px-6 py-4">Produk</th>
                <th class="px-6 py-4">Sumber</th>
                <th class="px-6 py-4">Tipe</th>
                <th class="px-6 py-4">Oleh / Karyawan</th>
                <th class="px-6 py-4 text-right">Jumlah</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="log in filteredGeneralInventoryLogs"
                :key="log.id"
                class="hover:bg-slate-50 transition-colors"
              >
                <td class="px-6 py-4">
                  <div class="font-bold text-slate-700">
                    {{
                      new Date(log.created_at).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'short',
                        year: 'numeric',
                      })
                    }}
                  </div>
                  <div class="text-[10px] text-slate-400">
                    {{
                      new Date(log.created_at).toLocaleTimeString('id-ID', {
                        hour: '2-digit',
                        minute: '2-digit',
                      })
                    }}
                  </div>
                </td>
                <td class="px-6 py-4">
                  <div class="font-bold text-slate-900 uppercase">
                    {{ log.factory?.name || 'Unknown' }}
                  </div>
                </td>
                <td class="px-6 py-4">
                  <div class="font-bold text-slate-900">{{ log.product?.name || 'Unknown' }}</div>
                  <div class="text-[10px] text-slate-500">Batch: {{ log.batch_no || '-' }}</div>
                </td>
                <td class="px-6 py-4">
                  <Badge
                    variant="outline"
                    :class="
                      log.source.startsWith('ADJUSTMENT')
                        ? 'bg-orange-50 text-orange-600 border-orange-200'
                        : log.source.startsWith('PRODUCTION')
                          ? 'bg-blue-50 text-blue-600 border-blue-200'
                          : ''
                    "
                  >
                    {{ log.source }}
                  </Badge>
                </td>
                <td class="px-6 py-4">
                  <span
                    :class="log.type === 'IN' ? 'text-emerald-600' : 'text-rose-600'"
                    class="font-black text-xs"
                  >
                    {{ log.type }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span class="text-[11px] font-bold text-slate-700">
                    {{ log.employee ? (log.employee.first_name + ' ' + (log.employee.last_name || '')) : (log.user ? (log.user.email || log.user.name) : 'Sistem') }}
                  </span>
                </td>
                <td class="px-6 py-4 text-right">
                  <span
                    class="font-black text-lg"
                    :class="log.type === 'IN' ? 'text-emerald-600' : 'text-rose-600'"
                  >
                    {{ log.type === 'IN' ? '+' : '-' }}{{ log.quantity }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>

    <!-- Delete Stock Confirmation Modal -->
    <div
      v-if="showDeleteStock"
      class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-[70] flex items-center justify-center p-4"
      @click.self="closeDeleteStock"
    >
      <Card class="w-full max-w-md shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200">
        <CardHeader class="p-6 border-b bg-white">
          <CardTitle class="text-xl font-bold text-rose-700">Hapus Stok?</CardTitle>
          <CardDescription>
            Tindakan ini akan menghapus saldo stok yang dipilih secara permanen.
          </CardDescription>
        </CardHeader>
        <CardContent class="p-6 bg-white">
          <div class="rounded-2xl border border-rose-100 bg-rose-50 p-4 space-y-2">
            <div class="flex justify-between gap-4">
              <span class="text-xs font-bold uppercase tracking-wide text-rose-500">Produk</span>
              <span class="text-sm font-bold text-slate-900 text-right">{{ selectedStockToDelete?.product?.name || '-' }}</span>
            </div>
            <div class="flex justify-between gap-4">
              <span class="text-xs font-bold uppercase tracking-wide text-rose-500">Pabrik</span>
              <span class="text-sm font-semibold text-slate-700 text-right">{{ selectedStockToDelete?.factory?.name || '-' }}</span>
            </div>
            <div class="flex justify-between gap-4">
              <span class="text-xs font-bold uppercase tracking-wide text-rose-500">Jumlah</span>
              <span class="text-sm font-black text-slate-900 text-right">{{ selectedStockToDelete?.quantity || 0 }} {{ selectedStockToDelete?.product?.unit || '' }}</span>
            </div>
          </div>
          <p class="mt-4 text-xs text-slate-500">Pastikan data yang dipilih sudah benar sebelum melanjutkan.</p>
        </CardContent>
        <CardFooter class="bg-slate-50/50 p-6 flex justify-end gap-3">
          <Button variant="ghost" @click="closeDeleteStock" :disabled="deletingStock" class="rounded-xl font-bold">Batal</Button>
          <Button @click="handleDeleteStock" :disabled="deletingStock" class="rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-bold">
            {{ deletingStock ? 'Menghapus...' : 'Ya, Hapus Stok' }}
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Modals (Adjust Stock) -->
    <div
      v-if="showAdjustStock"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-md shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200"
      >
        <CardHeader class="p-6 border-b bg-white">
          <CardTitle class="text-xl font-bold">Koreksi Stok</CardTitle>
          <CardDescription
            >Sesuaikan jumlah stok secara manual untuk pabrik
            {{ selectedStock?.factory?.name }}.</CardDescription
          >
        </CardHeader>
        <CardContent class="p-6 space-y-4 bg-white">
          <div
            class="p-4 bg-blue-50 rounded-2xl border border-blue-100 flex justify-between items-center"
          >
            <div>
              <p class="text-xs font-bold text-blue-600 uppercase tracking-widest mb-1">Produk</p>
              <p class="text-sm font-bold text-slate-900">{{ selectedStock?.product?.name }}</p>
              <p v-if="selectedStock?.batch_no" class="text-[10px] text-slate-400 mt-0.5 font-mono">
                Batch: {{ selectedStock?.batch_no }}
              </p>
            </div>
            <div class="text-right">
              <p class="text-xs font-bold text-slate-400 uppercase">Stok Sekarang</p>
              <p class="text-lg font-black text-slate-900">{{ selectedStock?.quantity }}</p>
            </div>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Jumlah Stok Baru</label>
            <input
              type="number"
              v-model="adjustStockForm.quantity"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Alasan</label>
            <textarea
              v-model="adjustStockForm.reason"
              placeholder="Audit stok, Barang rusak, pindah gudang, dsb..."
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20 resize-none"
              rows="2"
            ></textarea>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Pindah Lokasi Gudang (Opsional)</label>
            <select
              v-model="adjustStockForm.factory_warehouse_id"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20"
            >
              <option value="">-- Tetap di Gudang Saat Ini --</option>
              <option v-for="w in factoryStore.factoryWarehouses" :key="w.id" :value="w.id">
                {{ w.name }} ({{ w.code }})
              </option>
            </select>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 p-6 flex justify-end gap-3">
          <Button variant="ghost" @click="showAdjustStock = false" class="rounded-xl font-bold"
            >Batal</Button
          >
          <Button
            @click="handleAdjustStockSubmit"
            :loading="submitting"
            class="rounded-xl px-8 shadow-lg shadow-primary/20 font-bold bg-blue-600 hover:bg-blue-700 text-white"
          >
            Simpan Perubahan
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Log Production Modal -->
    <div
      v-if="showLogProduction"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-50 flex items-center justify-center p-2 sm:p-4"
    >
      <Card
        class="w-full max-w-2xl shadow-2xl rounded-[2rem] border-none animate-in fade-in zoom-in-95 duration-300 bg-white flex flex-col max-h-[95vh]"
      >
        <CardHeader class="p-5 sm:p-6 border-b">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div
                class="w-10 h-10 rounded-2xl bg-primary/10 flex items-center justify-center text-primary"
              >
                <Plus v-if="!isEditingProduction" class="h-5 w-5" />
                <Edit2 v-else class="h-5 w-5" />
              </div>
              <div>
                <CardTitle class="text-lg font-black text-slate-900">{{
                  isEditingProduction ? 'Edit Log Produksi' : 'Catat Produksi Global'
                }}</CardTitle>
                <CardDescription
                  class="text-[10px] font-bold text-slate-400 uppercase tracking-widest"
                  >Sinkronisasi stok otomatis</CardDescription
                >
              </div>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showLogProduction = false"
              class="rounded-xl text-slate-400 hover:text-slate-900"
            >
              <X class="h-5 w-5" />
            </Button>
          </div>
        </CardHeader>

        <CardContent class="p-5 sm:p-6 space-y-6 overflow-y-auto custom-scrollbar flex-1 pb-40">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
            <div class="space-y-1.5 sm:col-span-2">
              <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                >Pabrik Produksi</label
              >
              <div class="relative">
                <select
                  v-model="productionForm.factory_id"
                  :disabled="isEditingProduction"
                  class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none appearance-none focus:ring-4 focus:ring-primary/5 transition-all disabled:opacity-50"
                >
                  <option value="" disabled>Pilih Pabrik...</option>
                  <option v-for="f in factoryStore.factories" :key="f.id" :value="f.id">
                    {{ f.name }}
                  </option>
                </select>
                <div
                  class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-slate-400"
                >
                  <Settings2 class="h-4 w-4" />
                </div>
              </div>
            </div>
          </div>
          
          <div class="space-y-4">
            <div class="flex justify-between items-center">
              <div>
                <h3 class="text-sm font-bold text-slate-900">1. Daftar Produk Jadi yang Dihasilkan</h3>
                <p class="text-[10px] text-slate-400">Tambahkan 1 atau lebih produk yang selesai diproduksi</p>
              </div>
              <Button v-if="!isEditingProduction" variant="outline" size="sm" @click="addItem" class="h-8 rounded-lg text-xs border-primary text-primary hover:bg-primary/5">
                <Plus class="h-3 w-3 mr-1" /> Tambah Produk
              </Button>
            </div>

            <div v-for="(item, idx) in productionForm.items" :key="idx" class="p-4 bg-slate-50/50 border border-slate-100 rounded-2xl space-y-4 relative">
              <Button v-if="!isEditingProduction && productionForm.items.length > 1" variant="ghost" size="icon" class="absolute top-2 right-2 h-7 w-7 text-rose-500 hover:bg-rose-50 hover:text-rose-600 rounded-full z-10" @click="removeItem(idx)">
                <X class="h-4 w-4" />
              </Button>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
                <div class="space-y-1.5">
                  <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Pilih Produk</label
                  >
                  <SearchSelect
                    v-model="item.product_id"
                    :options="finishedGoodsOptions"
                    placeholder="Pilih Produk..."
                    search-placeholder="Ketik nama / SKU produk..."
                    empty-message="Produk tidak ditemukan"
                    @update:model-value="onProductChange(item)"
                  />
                </div>

                <div class="space-y-1.5">
                  <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Pilih Gudang Target (Opsional)</label
                  >
                  <div class="relative">
                    <select
                      v-model="item.target_warehouse_id"
                      class="w-full px-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-bold outline-none appearance-none focus:ring-4 focus:ring-primary/5 transition-all"
                    >
                      <option value="">Tidak Ditentukan (Gudang Umum)</option>
                      <option v-for="w in factoryStore.factoryWarehouses" :key="w.id" :value="w.id">
                        {{ w.name }} <span v-if="w.code">({{ w.code }})</span>
                      </option>
                    </select>
                    <div
                      class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-slate-400"
                    >
                      <Boxes class="h-4 w-4" />
                    </div>
                  </div>
                </div>
              </div>

              <!-- Batch & Expiry -->
              <div class="grid grid-cols-2 gap-5">
                <div class="space-y-1.5">
                  <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Nomor Batch</label
                  >
                  <input
                    type="text"
                    v-model="item.batch_no"
                    placeholder="Contoh: B1-2024"
                    class="w-full px-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-bold outline-none focus:ring-4 focus:ring-primary/5 transition-all"
                  />
                </div>
                <div class="space-y-1.5">
                  <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Tgl Kadaluarsa</label
                  >
                  <input
                    type="date"
                    v-model="item.expiry_date"
                    class="w-full px-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-bold outline-none focus:ring-4 focus:ring-primary/5 transition-all"
                  />
                </div>
              </div>

              <!-- Modern Carton Calculator -->
              <div
                class="p-4 bg-white rounded-[1.5rem] border border-slate-100 shadow-sm flex flex-col gap-3"
              >
                <div
                  class="flex items-center gap-2 text-primary font-black text-[10px] uppercase tracking-widest px-1"
                >
                  <Boxes class="h-3.5 w-3.5" />
                  <span>Packaging Unit (Karton / Krat)</span>
                </div>

                <div class="grid grid-cols-2 gap-4">
                  <div
                    class="p-3 bg-slate-50 rounded-xl border border-slate-100 focus-within:ring-2 focus-within:ring-primary/20 transition-all"
                  >
                    <label
                      class="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-1 block"
                      >Jumlah Karton / Krat</label
                    >
                    <input
                      type="number"
                      v-model="item.carton_count"
                      @input="syncQuantity(item)"
                      class="w-full bg-transparent border-none p-0 text-sm font-black outline-none"
                      placeholder="0"
                    />
                  </div>
                  <div
                    class="p-3 bg-slate-50 rounded-xl border border-slate-100 focus-within:ring-2 focus-within:ring-primary/20 transition-all"
                  >
                    <label
                      class="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-1 block"
                      >Isi per Karton / Krat</label
                    >
                    <input
                      type="number"
                      v-model="item.pieces_per_carton"
                      @input="syncQuantity(item)"
                      class="w-full bg-transparent border-none p-0 text-sm font-black outline-none"
                      placeholder="0"
                    />
                  </div>
                </div>

                <div
                  class="p-3 bg-primary rounded-xl shadow-md shadow-primary/10 flex items-center justify-between text-white"
                >
                  <div class="flex flex-col">
                    <span class="text-[9px] font-black text-white/60 uppercase tracking-widest"
                      >Total Produksi (PCS)</span
                    >
                    <input
                      type="number"
                      v-model="item.quantity"
                      class="text-lg font-black bg-transparent border-none p-0 outline-none w-32 placeholder:text-white/30"
                    />
                  </div>
                  <div class="w-8 h-8 rounded-lg bg-white/10 flex items-center justify-center">
                    <Activity class="h-4 w-4" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Section 2: Global Raw Material & Packaging Consumption -->
          <div v-if="canViewMaterials" class="p-5 bg-amber-50/60 rounded-[1.5rem] border border-amber-200/70 space-y-4">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <div class="flex items-center gap-2">
                  <Component class="w-4 h-4 text-amber-700" />
                  <span class="text-sm font-black text-amber-900">2. Pemakaian Bahan Baku & Kemasan {{ isEditingProduction ? '(Dapat Diedit)' : '(Global Sesi Ini)' }}</span>
                </div>
                <p class="text-[11px] text-amber-800/80 mt-0.5">
                  Inputkan seluruh bahan baku (gula, kedelai, garam, rempah) dan kemasan/wadah (botol, label, krat) yang dipakai untuk memproduksi produk di atas. Stok akan otomatis disesuaikan.
                </p>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                @click="addRawMaterial"
                class="h-8 text-xs rounded-xl border-amber-300 text-amber-900 bg-amber-100/70 hover:bg-amber-100 font-bold shrink-0 self-start sm:self-auto"
              >
                <Plus class="w-3.5 h-3.5 mr-1" /> Tambah Bahan / Kemasan
              </Button>
            </div>

            <div v-if="!productionForm.raw_materials || productionForm.raw_materials.length === 0" class="p-4 bg-white/60 rounded-xl border border-amber-200/50 text-[11px] text-amber-700/90 italic flex items-center gap-2">
              <Boxes class="w-4 h-4 text-amber-500 shrink-0" />
              <span>Belum ada bahan baku / kemasan yang ditambahkan. Stok bahan baku di pabrik tidak akan dipotong jika dikosongkan.</span>
            </div>

            <div v-else class="space-y-2.5">
              <div
                v-for="(rm, rmIdx) in productionForm.raw_materials"
                :key="rmIdx"
                class="space-y-1 bg-white p-3 rounded-xl border border-amber-100 shadow-sm"
              >
                <div class="grid grid-cols-1 sm:grid-cols-12 gap-2.5 items-center">
                  <div class="sm:col-span-4">
                    <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest block mb-1">Bahan / Wadah</label>
                    <select
                      v-model="rm.product_id"
                      @change="onRawMaterialChange(rm)"
                      class="w-full text-xs font-bold bg-slate-50 border border-slate-200 rounded-lg p-2 outline-none focus:ring-2 focus:ring-amber-500/20"
                    >
                      <option value="" disabled>-- Pilih Bahan / Wadah --</option>
                      <optgroup v-if="productionUsableMaterials.rawMaterials.length > 0" label="⭐ BAHAN BAKU UTAMA">
                        <option v-for="p in productionUsableMaterials.rawMaterials" :key="p.id" :value="p.id">
                          {{ p.sku }} - {{ p.name }} (Basis: {{ p.unit || 'KG' }})
                        </option>
                      </optgroup>
                      <optgroup v-if="productionUsableMaterials.packagingMaterials.length > 0" label="📦 KEMASAN & WADAH">
                        <option v-for="p in productionUsableMaterials.packagingMaterials" :key="p.id" :value="p.id">
                          {{ p.sku }} - {{ p.name }} (Basis: {{ p.unit || 'PCS' }})
                        </option>
                      </optgroup>
                      <optgroup v-if="productionUsableMaterials.tradingProducts.length > 0" label="🏷️ PRODUK PELENGKAP / TRADING">
                        <option v-for="p in productionUsableMaterials.tradingProducts" :key="p.id" :value="p.id">
                          {{ p.sku }} - {{ p.name }} (Basis: {{ p.unit || 'PCS' }})
                        </option>
                      </optgroup>
                      <optgroup v-if="productionUsableMaterials.giftProducts.length > 0" label="🎁 HADIAH & BARANG SAMPINGAN">
                        <option v-for="p in productionUsableMaterials.giftProducts" :key="p.id" :value="p.id">
                          {{ p.sku }} - {{ p.name }} (Basis: {{ p.unit || 'PCS' }})
                        </option>
                      </optgroup>
                    </select>
                  </div>
                  <div class="sm:col-span-3">
                    <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest block mb-1">Batch Sumber</label>
                    <select
                      v-model="rm.batch_no"
                      class="w-full text-xs font-bold bg-slate-50 border border-slate-200 rounded-lg p-2 outline-none focus:ring-2 focus:ring-amber-500/20"
                    >
                      <option value="">Auto (FEFO / Exp Terdekat)</option>
                      <option
                        v-for="b in getAvailableBatchesForRawMaterial(productionForm.factory_id, rm.product_id, rm.batch_no)"
                        :key="b.batch_no"
                        :value="b.batch_no"
                      >
                        {{ b.batch_no }} (Sisa: {{ b.quantity }} | {{ b.expiry_date ? b.expiry_date.substring(0,10) : 'No Exp' }})
                      </option>
                    </select>
                  </div>
                  <div class="sm:col-span-2">
                    <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest block mb-1">Jumlah</label>
                    <input
                      type="number"
                      v-model.number="rm.quantity"
                      min="0.001"
                      step="any"
                      placeholder="Jumlah"
                      class="w-full text-xs font-bold font-mono bg-slate-50 border border-slate-200 rounded-lg p-2 outline-none focus:ring-2 focus:ring-amber-500/20"
                    />
                  </div>
                  <div class="sm:col-span-2">
                    <label class="text-[9px] font-black text-slate-400 uppercase tracking-widest block mb-1">Satuan</label>
                    <select
                      v-model="rm.unit"
                      class="w-full text-xs font-bold uppercase bg-slate-50 border border-slate-200 rounded-lg p-2 outline-none focus:ring-2 focus:ring-amber-500/20"
                    >
                      <option v-for="u in getRawMaterialCompatibleUnits(rm.product_id)" :key="u" :value="u">{{ u }}</option>
                    </select>
                  </div>
                  <div class="sm:col-span-1 flex justify-end items-end pb-1">
                    <button
                      type="button"
                      @click="removeRawMaterial(rmIdx)"
                      class="text-rose-500 hover:text-rose-700 text-xs font-bold p-1.5 rounded-lg hover:bg-rose-50 transition-colors"
                      title="Hapus bahan"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </div>

                <!-- Live Conversion Hint -->
                <div
                  v-if="getRawMaterialConversionHint(rm)"
                  class="text-[10px] text-amber-700 font-medium px-1 pt-1 flex items-center gap-1"
                >
                  <Sparkles class="w-2.5 h-2.5 text-amber-600" />
                  <span>{{ getRawMaterialConversionHint(rm) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Multiple Operator Selector -->
          <div class="space-y-3" ref="operatorDropdownRef">
            <div class="flex items-center justify-between">
              <div>
                <label class="text-xs font-black text-slate-800 uppercase tracking-wider flex items-center gap-1.5">
                  <User class="h-3.5 w-3.5 text-primary" />
                  <span>Operator Produksi (Bisa Banyak)</span>
                  <span class="text-rose-500">*</span>
                </label>
                <p class="text-[10px] text-slate-400">
                  Pilih dari karyawan terdaftar atau tambahkan nama pekerja harian lepas secara manual.
                </p>
              </div>
              <button
                v-if="productionForm.operators.length > 0"
                type="button"
                @click="clearAllOperators"
                class="text-[10px] font-bold text-rose-500 hover:text-rose-700 transition-colors"
              >
                Hapus Semua ({{ productionForm.operators.length }})
              </button>
            </div>

            <!-- Chips / List of Selected Operators -->
            <div
              v-if="productionForm.operators.length > 0"
              class="flex flex-wrap gap-2 p-3 bg-slate-50/90 rounded-2xl border border-slate-100 min-h-[50px] items-center"
            >
              <div
                v-for="(op, opIdx) in productionForm.operators"
                :key="opIdx"
                class="inline-flex items-center gap-1.5 pl-3 pr-1.5 py-1.5 rounded-xl text-xs font-bold transition-all shadow-2xs group"
                :class="
                  op.is_manual
                    ? 'bg-amber-50 text-amber-900 border border-amber-300/80'
                    : 'bg-white text-slate-800 border border-slate-200 hover:border-primary/40'
                "
              >
                <div class="flex items-center gap-1.5">
                  <span class="text-slate-900 font-black">{{ op.operator_name }}</span>
                  <span
                    v-if="op.is_manual"
                    class="text-[8px] font-black uppercase px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-800 tracking-wider"
                  >
                    Manual / Harian
                  </span>
                  <span
                    v-else-if="op.employee_id_number"
                    class="text-[9px] font-mono font-bold text-slate-400"
                  >
                    ({{ op.employee_id_number }})
                  </span>
                </div>
                <button
                  type="button"
                  @click="removeOperator(opIdx)"
                  class="h-5 w-5 rounded-lg flex items-center justify-center text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition-colors"
                  title="Hapus operator ini"
                >
                  <X class="h-3.5 w-3.5" />
                </button>
              </div>
            </div>

            <div
              v-else
              class="p-3.5 bg-slate-50/50 rounded-2xl border border-dashed border-slate-200 text-center text-xs text-slate-400 font-medium"
            >
              Belum ada operator yang ditambahkan. Silakan pilih karyawan atau ketik nama secara manual di bawah.
            </div>

            <!-- Input and Dropdown -->
            <div class="relative">
              <div class="flex gap-2">
                <div class="relative flex-1 group">
                  <div
                    class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-primary transition-colors"
                  >
                    <Search class="h-4 w-4" />
                  </div>
                  <input
                    type="text"
                    v-model="operatorSearch"
                    @focus="showOperatorDropdown = true"
                    @keydown.enter.prevent="addManualOperator"
                    placeholder="Cari nama/NIP karyawan atau ketik nama pekerja harian..."
                    class="w-full pl-11 pr-4 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-bold outline-none focus:ring-4 focus:ring-primary/5 focus:border-primary transition-all"
                  />
                  <button
                    v-if="operatorSearch"
                    @click="operatorSearch = ''"
                    type="button"
                    class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-300 hover:text-rose-500 transition-colors"
                  >
                    <X class="h-4 w-4" />
                  </button>
                </div>
                <Button
                  type="button"
                  @click="addManualOperator"
                  :disabled="!operatorSearch.trim()"
                  variant="outline"
                  class="rounded-2xl px-4 py-3 h-auto font-bold border-slate-200 text-slate-700 hover:bg-slate-50 hover:text-primary shrink-0 flex items-center gap-1.5 shadow-2xs"
                  title="Tambah nama manual (pekerja harian / helper)"
                >
                  <Plus class="h-4 w-4 text-primary" />
                  <span>Tambah Manual</span>
                </Button>
              </div>

              <!-- Dropdown List -->
              <div
                v-if="showOperatorDropdown && (filteredEmployees.length > 0 || operatorSearch.trim())"
                class="absolute top-full left-0 right-0 mt-2 bg-white border border-slate-100 rounded-[1.5rem] shadow-2xl z-[70] overflow-hidden max-h-60 overflow-y-auto custom-scrollbar border-t-4 border-t-primary animate-in fade-in slide-in-from-top-2 duration-200"
              >
                <!-- Quick action: Add as manual -->
                <div
                  v-if="operatorSearch.trim()"
                  @click="addManualOperator"
                  class="px-5 py-3 bg-amber-50/80 hover:bg-amber-100/90 cursor-pointer flex items-center justify-between border-b border-amber-100 group/item transition-colors"
                >
                  <div class="flex items-center gap-2">
                    <Plus class="h-4 w-4 text-amber-700" />
                    <div>
                      <p class="text-xs font-black text-amber-900">
                        Tambah "<span class="underline">{{ operatorSearch.trim() }}</span>" sebagai Operator Manual
                      </p>
                      <p class="text-[10px] text-amber-700 font-medium">Pekerja harian / freelance / helper pabrik</p>
                    </div>
                  </div>
                  <Badge class="bg-amber-600 text-white text-[8px] font-black uppercase border-none px-2 py-0.5">
                    + Tambah Manual
                  </Badge>
                </div>

                <div v-if="filteredEmployees.length === 0 && !operatorSearch.trim()" class="p-6 text-center">
                  <p class="text-xs font-bold text-slate-400 uppercase">Semua karyawan sudah ditambahkan</p>
                </div>
                <div v-else>
                  <div
                    v-for="emp in filteredEmployees"
                    :key="emp.id"
                    @click="addEmployeeOperator(emp)"
                    class="px-5 py-3 hover:bg-slate-50 cursor-pointer flex items-center justify-between group/item border-b border-slate-50 last:border-none"
                  >
                    <div>
                      <p
                        class="text-[11px] font-black text-slate-900 group-hover/item:text-primary transition-colors uppercase tracking-tight"
                      >
                        {{ emp.first_name }} {{ emp.last_name }}
                      </p>
                      <div class="flex items-center gap-2 text-[9px] text-slate-400 font-bold">
                        <span v-if="emp.employee_id_number">{{ emp.employee_id_number }}</span>
                        <span v-if="emp.job_position?.name">• {{ emp.job_position.name }}</span>
                      </div>
                    </div>
                    <div class="opacity-0 group-hover/item:opacity-100 transition-opacity">
                      <Badge
                        class="bg-primary/10 text-primary text-[8px] font-black uppercase border-none h-5 px-2"
                        >+ Pilih</Badge
                      >
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-1.5">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Catatan Tambahan</label
            >
            <textarea
              v-model="productionForm.notes"
              placeholder="Tambahkan keterangan..."
              class="w-full px-5 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-4 focus:ring-primary/5 transition-all resize-none"
              rows="2"
            ></textarea>
          </div>
        </CardContent>

        <CardFooter
          class="p-5 sm:p-6 border-t bg-white rounded-b-[2rem] flex flex-col sm:flex-row gap-3"
        >
          <Button
            variant="ghost"
            @click="showLogProduction = false"
            class="rounded-2xl font-bold w-full sm:w-auto order-2 sm:order-1 h-12"
            >Batal</Button
          >
          <Button
            @click="handleProductionSubmit"
            :loading="submitting"
            :disabled="submitting"
            class="rounded-2xl px-12 shadow-xl shadow-primary/20 font-black bg-primary hover:bg-primary/90 text-white w-full sm:w-auto flex-1 order-1 sm:order-2 h-12"
          >
            {{ isEditingProduction ? 'Update Log' : 'Simpan Produksi' }}
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Stock Impact Confirmation Modal -->
    <div
      v-if="showStockImpactModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-[70] flex items-center justify-center p-4 overflow-y-auto animate-in fade-in duration-200"
    >
      <Card
        class="w-full max-w-xl bg-white shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200"
      >
        <CardHeader class="p-6 border-b border-slate-100 bg-gradient-to-r from-amber-500/10 via-orange-50/50 to-amber-500/10">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-amber-500/15 border border-amber-500/30 flex items-center justify-center text-amber-700 shadow-sm shrink-0">
              <Boxes class="h-5 w-5" />
            </div>
            <div>
              <CardTitle class="text-lg font-black text-slate-900">
                Konfirmasi Penyesuaian Stok Produksi
              </CardTitle>
              <CardDescription class="text-xs text-slate-500 mt-0.5">
                Periksa rincian pengembalian (+) dan pemotongan (-) stok sebelum menyimpan log produksi.
              </CardDescription>
            </div>
          </div>
        </CardHeader>

        <CardContent class="p-6 space-y-4 max-h-[60vh] overflow-y-auto custom-scrollbar">
          <!-- Finished Goods Impact -->
          <div v-if="stockImpactDiff.finishedGood" class="p-4 rounded-2xl border border-slate-200 bg-slate-50/50 space-y-2">
            <div class="text-[10px] font-black uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
              <Package class="w-3.5 h-3.5 text-slate-600" />
              <span>Hasil Produk Jadi (Finished Goods)</span>
            </div>
            <div class="flex items-center justify-between">
              <div>
                <div class="font-bold text-xs text-slate-800">{{ stockImpactDiff.finishedGood.name }}</div>
                <div class="text-[10px] font-mono text-slate-400">Batch: {{ stockImpactDiff.finishedGood.batch_no }}</div>
              </div>
              <div class="text-right">
                <div class="text-xs font-bold text-slate-400 line-through" v-if="stockImpactDiff.finishedGood.old_qty !== stockImpactDiff.finishedGood.new_qty">
                  {{ stockImpactDiff.finishedGood.old_qty }} PCS
                </div>
                <div class="text-sm font-black text-slate-900">
                  {{ stockImpactDiff.finishedGood.new_qty }} PCS
                </div>
                <span
                  class="text-[10px] font-bold px-2 py-0.5 rounded-full inline-block mt-0.5"
                  :class="stockImpactDiff.finishedGood.delta > 0 ? 'bg-emerald-100 text-emerald-800' : stockImpactDiff.finishedGood.delta < 0 ? 'bg-rose-100 text-rose-800' : 'bg-slate-100 text-slate-600'"
                >
                  {{ stockImpactDiff.finishedGood.delta > 0 ? `+${stockImpactDiff.finishedGood.delta} PCS (Stok Bertambah)` : stockImpactDiff.finishedGood.delta < 0 ? `${stockImpactDiff.finishedGood.delta} PCS (Stok Berkurang)` : 'Stok Tetap' }}
                </span>
              </div>
            </div>
          </div>

          <!-- Raw Materials Impact -->
          <div class="space-y-2">
            <div class="text-[10px] font-black uppercase tracking-wider text-slate-400 flex items-center justify-between">
              <span class="flex items-center gap-1.5">
                <Component class="w-3.5 h-3.5 text-amber-600" />
                <span>Penyesuaian Stok Bahan Baku & Kemasan</span>
              </span>
              <span class="text-[10px] text-slate-500 font-bold">{{ stockImpactDiff.materials.length }} Item</span>
            </div>

            <div v-if="stockImpactDiff.materials.length === 0" class="p-4 bg-slate-50 rounded-xl text-center text-xs text-slate-400 italic">
              Tidak ada bahan baku yang dicatat pada log ini.
            </div>

            <div v-else class="space-y-2">
              <div
                v-for="(item, idx) in stockImpactDiff.materials"
                :key="idx"
                class="p-3 rounded-2xl border transition-all"
                :class="item.type === 'RETURNED' ? 'bg-emerald-50/60 border-emerald-200/80' : item.type === 'DEDUCTED' ? 'bg-rose-50/60 border-rose-200/80' : 'bg-slate-50 border-slate-200'"
              >
                <div class="flex items-start justify-between gap-2">
                  <div class="space-y-0.5">
                    <div class="flex items-center gap-1.5">
                      <span
                        class="w-2 h-2 rounded-full shrink-0"
                        :class="item.type === 'RETURNED' ? 'bg-emerald-500' : item.type === 'DEDUCTED' ? 'bg-rose-500' : 'bg-slate-400'"
                      ></span>
                      <span class="font-bold text-xs text-slate-800">{{ item.name }}</span>
                    </div>
                    <div class="text-[10px] font-mono text-slate-500 pl-3.5">
                      Batch: <span class="font-bold text-slate-700">{{ item.batch_no || 'AUTO' }}</span>
                      <span v-if="item.notes" class="text-slate-400 italic ml-1.5">({{ item.notes }})</span>
                    </div>
                  </div>

                  <div class="text-right shrink-0">
                    <div class="text-xs font-black" :class="item.type === 'RETURNED' ? 'text-emerald-700' : item.type === 'DEDUCTED' ? 'text-rose-700' : 'text-slate-700'">
                      {{ item.type === 'RETURNED' ? `+${item.delta_qty}` : item.type === 'DEDUCTED' ? `-${item.delta_qty}` : '0' }} {{ item.unit }}
                    </div>
                    <Badge
                      variant="outline"
                      class="text-[9px] font-bold mt-0.5 px-1.5 py-0"
                      :class="item.type === 'RETURNED' ? 'border-emerald-300 bg-emerald-100 text-emerald-800' : item.type === 'DEDUCTED' ? 'border-rose-300 bg-rose-100 text-rose-800' : 'border-slate-200 bg-slate-100 text-slate-600'"
                    >
                      {{ item.type === 'RETURNED' ? '🟢 Dikembalikan ke Stok' : item.type === 'DEDUCTED' ? '🔴 Dipotong dari Stok' : '⚪ Tidak Berubah' }}
                    </Badge>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </CardContent>

        <CardFooter class="p-4 border-t border-slate-100 bg-slate-50 flex items-center justify-between gap-3">
          <Button
            variant="outline"
            @click="showStockImpactModal = false"
            class="rounded-xl font-bold text-xs border-slate-200 hover:bg-slate-100"
          >
            Periksa Kembali
          </Button>
          <Button
            @click="executeProductionUpdate"
            :loading="submitting"
            :disabled="submitting"
            class="rounded-xl px-6 font-black text-xs bg-primary hover:bg-primary/90 text-white shadow-lg shadow-primary/20"
          >
            Ya, Konfirmasi & Simpan
          </Button>
        </CardFooter>
      </Card>
    </div>
    
    <!-- Transfer History Modal -->
    <div
      v-if="showTransferHistory"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-[60] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-4xl shadow-2xl rounded-3xl overflow-hidden border-none animate-in zoom-in-95 duration-200"
      >
        <CardHeader class="p-6 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-xl font-bold">Histori Pengiriman Cabang</CardTitle>
              <CardDescription v-if="selectedProductionLog">
                Riwayat pengiriman untuk Batch <span class="font-black font-mono text-slate-800">{{ selectedProductionLog.batch_no }}</span>
              </CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showTransferHistory = false"
              class="rounded-full"
            >
              <X class="h-5 w-5 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-0 bg-white max-h-[60vh] overflow-y-auto">
          <div v-if="transferHistoryLoading" class="p-12 text-center text-slate-400 font-medium flex justify-center items-center">
            <RefreshCw class="w-6 h-6 animate-spin mr-2" />
            Memuat histori pengiriman...
          </div>
          <div
            v-else-if="transferHistoryData.length === 0"
            class="p-12 text-center text-slate-400 font-medium"
          >
            Belum ada pengiriman ke cabang untuk batch ini.
          </div>
          <table v-else class="w-full text-sm text-left">
            <thead
              class="bg-slate-50 sticky top-0 border-b border-slate-100 text-[10px] font-black text-slate-400 uppercase tracking-widest"
            >
              <tr>
                <th class="px-6 py-4">Tanggal Dibuat</th>
                <th class="px-6 py-4">Cabang Tujuan</th>
                <th class="px-6 py-4">Status</th>
                <th class="px-6 py-4 text-right">Jumlah (Unit)</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="log in transferHistoryData"
                :key="log.id"
                class="hover:bg-slate-50 transition-colors"
              >
                <td class="px-6 py-4">
                  <div class="font-bold text-slate-700">
                    {{
                      new Date(log.created_at).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'short',
                        year: 'numeric',
                      })
                    }}
                  </div>
                </td>
                <td class="px-6 py-4">
                  <div class="font-bold text-slate-900 uppercase">
                    {{ log.to_branch?.name || 'Unknown' }}
                  </div>
                </td>
                <td class="px-6 py-4">
                  <Badge
                    variant="outline"
                  >
                    {{ log.status }}
                  </Badge>
                </td>
                <td class="px-6 py-4 text-right">
                  <span class="font-black text-slate-900">{{ log.quantity }}</span>
                  <span class="text-[10px] text-slate-400 ml-1 font-bold">{{ log.unit }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </CardContent>
      </Card>
    </div>

    <!-- Stock Movement & Traceability History Modal -->
    <StockMovementHistoryDialog
      v-model:is-open="showStockMovement"
      :stock-item="selectedMovementStock"
    />

    <!-- Production Raw Materials & Used Batches Detail Modal -->
    <ProductionMaterialsDialog
      v-model:open="showMaterialsModal"
      :production-log="selectedLogForMaterials"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { onClickOutside } from '@vueuse/core'
import { factoryApi } from '@/api/factory'
import { useFactoryStore } from '@/stores/factory'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'
import {
  groupProductsByCategory,
  ALL_COMMON_UNITS,
  getCompatibleUnits,
  convertUnitQuantity,
  formatMultiUnit,
} from '@/utils/productUnits'
import SearchSelect, { type SelectOption } from '@/components/SearchSelect.vue'
import StockMovementHistoryDialog from '@/components/factory/StockMovementHistoryDialog.vue'
import ProductionMaterialsDialog from '@/components/factory/ProductionMaterialsDialog.vue'

const authStore = useAuthStore()

const showMaterialsModal = ref(false)
const selectedLogForMaterials = ref<any>(null)

const openMaterialsDetail = (row: any) => {
  selectedLogForMaterials.value = row
  showMaterialsModal.value = true
}

const canViewMaterials = computed(() => {
  const role = (authStore.user?.role || '').toLowerCase()
  const email = (authStore.user?.email || '').toLowerCase().trim()

  if (['superadmin', 'admin'].includes(role)) return true
  if (role === 'supervisor_gudang') return true
  if (role === 'admin_pabrik' && email === 'rizkikusumaww@gmail.com') return true
  return false
})

const getErrorMessage = (err: any): string => {
  if (!err) return 'Unknown error'
  if (typeof err === 'string') return err
  if (err.response?.data?.error) return err.response.data.error
  if (err.response?.data?.message) return err.response.data.message
  if (err.message) return err.message
  try {
    return JSON.stringify(err)
  } catch (_) {
    return String(err)
  }
}
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import DataTable from '@/components/DataTable.vue'
import StockBatchSummaryCard from '@/components/factory/StockBatchSummaryCard.vue'
import {
  Boxes,
  Activity,
  Plus,
  Search,
  RefreshCw,
  Package,
  Settings2,
  Edit2,
  Trash2,
  User,
  History,
  X,
  Truck,
  Route,
} from 'lucide-vue-next'

const router = useRouter()
const factoryStore = useFactoryStore()
const masterStore = useMasterDataStore()
const stockSearch = ref('')
const submitting = ref(false)

// Modals State
const showAdjustStock = ref(false)
const showDeleteStock = ref(false)
const deletingStock = ref(false)
const showLogProduction = ref(false)
const showInventoryLogs = ref(false)
const inventoryLogFilter = ref('ALL')
const showStockMovement = ref(false)
const selectedMovementStock = ref<any>(null)
const isEditingProduction = ref(false)
const selectedStock = ref<any>(null)
const selectedStockToDelete = ref<any>(null)
const currentProductionId = ref('')

const openStockMovement = (row: any) => {
  selectedMovementStock.value = row
  showStockMovement.value = true
}

const filteredGeneralInventoryLogs = computed(() => {
  const logs = factoryStore.allInventoryLogs || []
  if (inventoryLogFilter.value === 'ALL') return logs
  if (inventoryLogFilter.value === 'ADJUSTMENT') {
    return logs.filter((l: any) => (l.source || '').startsWith('ADJUSTMENT') || (l.source || '').startsWith('PABRIK_ADJUSTMENT'))
  }
  if (inventoryLogFilter.value === 'PRODUCTION') {
    return logs.filter((l: any) => (l.source || '').startsWith('PRODUCTION'))
  }
  if (inventoryLogFilter.value === 'TRANSFER') {
    return logs.filter((l: any) => (l.source || '').startsWith('TRANSFER'))
  }
  return logs
})

const showTransferHistory = ref(false)
const transferHistoryLoading = ref(false)
const selectedProductionLog = ref<any>(null)
const transferHistoryData = ref<any[]>([])

interface SelectedOperator {
  employee_id?: string
  operator_name: string
  is_manual: boolean
  employee_id_number?: string
  position?: string
}

// Searchable Multiple Operator State
const operatorSearch = ref('')
const showOperatorDropdown = ref(false)
const operatorDropdownRef = ref(null)

onClickOutside(operatorDropdownRef, () => {
  showOperatorDropdown.value = false
})

const filteredEmployees = computed(() => {
  const query = (operatorSearch.value || '').toLowerCase().trim()
  const employees = masterStore.employees || []
  const addedEmpIds = new Set(
    productionForm.operators.filter((o) => o.employee_id).map((o) => o.employee_id)
  )
  const available = employees.filter((emp) => !addedEmpIds.has(emp.id))
  if (!query) return available.slice(0, 25)
  return available.filter((emp) => {
    const fullName = `${emp.first_name || ''} ${emp.last_name || ''}`.toLowerCase()
    const empId = (emp.employee_id_number || '').toLowerCase()
    return fullName.includes(query) || empId.includes(query)
  })
})

const addEmployeeOperator = (emp: any) => {
  const fullName = `${emp.first_name || ''} ${emp.last_name || ''}`.trim()
  if (!productionForm.operators.some((o) => o.employee_id === emp.id)) {
    productionForm.operators.push({
      employee_id: emp.id,
      operator_name: fullName,
      is_manual: false,
      employee_id_number: emp.employee_id_number,
      position: emp.job_position?.name || emp.position,
    })
  }
  operatorSearch.value = ''
  showOperatorDropdown.value = false
}

const addManualOperator = () => {
  const rawName = operatorSearch.value.trim()
  if (!rawName) return

  if (productionForm.operators.some((o) => o.operator_name.toLowerCase() === rawName.toLowerCase())) {
    toast.error(`Operator "${rawName}" sudah ada dalam daftar.`)
    return
  }

  productionForm.operators.push({
    employee_id: undefined,
    operator_name: rawName,
    is_manual: true,
  })

  operatorSearch.value = ''
  showOperatorDropdown.value = false
}

const removeOperator = (idx: number) => {
  productionForm.operators.splice(idx, 1)
}

const clearAllOperators = () => {
  productionForm.operators = []
  operatorSearch.value = ''
}

// Forms
const adjustStockForm = reactive({ quantity: 0, reason: '', factory_warehouse_id: '' })
const productionForm = reactive({
  factory_id: '',
  employee_id: '',
  operators: [] as SelectedOperator[],
  notes: '',
  items: [
    {
      product_id: '',
      target_warehouse_id: '',
      batch_no: '',
      expiry_date: '',
      quantity: 0,
      carton_count: 0,
      pieces_per_carton: 0,
    }
  ],
  raw_materials: [] as Array<{
    product_id: string
    batch_no: string
    quantity: number
    unit: string
    notes: string
  }>
})

const generateBatchNumber = (productId: string) => {
  const product = factoryStore.products.find((p) => p.id === productId)
  if (!product) return ''

  const sku = product.sku || product.name.substring(0, 3).toUpperCase()
  const now = new Date()

  const yy = String(now.getFullYear()).slice(2)
  const mm = String(now.getMonth() + 1).padStart(2, '0')
  const dd = String(now.getDate()).padStart(2, '0')

  const hh = String(now.getHours()).padStart(2, '0')
  const min = String(now.getMinutes()).padStart(2, '0')

  return `${sku}-${yy}${mm}${dd}-${hh}${min}`
}

const productionUsableMaterials = computed(() => {
  return groupProductsByCategory(factoryStore.products)
})

const finishedGoodsOptions = computed<SelectOption[]>(() => {
  const grouped = groupProductsByCategory(factoryStore.products)
  const goods = [
    ...grouped.finishedGoods,
    ...grouped.giftProducts,
    ...grouped.tradingProducts,
  ]
  return goods.map((p) => ({
    value: p.id,
    label: p.name,
    sublabel: p.sku ? `SKU: ${p.sku}` : undefined,
  }))
})

const onProductChange = (item: any) => {
  if (!isEditingProduction.value && item.product_id) {
    item.batch_no = generateBatchNumber(item.product_id)
  }
}

const addRawMaterial = () => {
  if (!productionForm.raw_materials) productionForm.raw_materials = []
  productionForm.raw_materials.push({
    product_id: '',
    batch_no: '',
    quantity: 1,
    unit: 'KG',
    notes: '',
  })
}

const removeRawMaterial = (rmIdx: number) => {
  if (productionForm.raw_materials) {
    productionForm.raw_materials.splice(rmIdx, 1)
  }
}

const onRawMaterialChange = (rm: any) => {
  const prod = factoryStore.products.find((p) => p.id === rm.product_id)
  if (prod) {
    rm.unit = prod.unit || 'KG'
  }
  rm.batch_no = '' // Reset batch when product changes
}

const getAvailableBatchesForRawMaterial = (factoryId: string, productId: string, currentBatchNo?: string) => {
  if (!factoryId || !productId) return []
  const batches = (factoryStore.allInventory || []).filter(
    (s) =>
      s.factory_id === factoryId &&
      s.product_id === productId &&
      (s.quantity > 0 || (currentBatchNo && s.batch_no === currentBatchNo)) &&
      s.batch_no !== 'BACKORDER'
  )
  if (currentBatchNo && currentBatchNo !== '' && currentBatchNo !== 'AUTO' && !batches.some((b) => b.batch_no === currentBatchNo)) {
    batches.push({
      batch_no: currentBatchNo,
      quantity: 0,
      expiry_date: undefined,
      factory_id: factoryId,
      product_id: productId,
    } as any)
  }
  return batches
}

const getRawMaterialBaseUnit = (productId: string) => {
  const prod = factoryStore.products.find((p) => p.id === productId)
  return prod?.unit || 'KG'
}

const getRawMaterialCompatibleUnits = (productId: string) => {
  const base = getRawMaterialBaseUnit(productId)
  return getCompatibleUnits(base)
}

const getRawMaterialConversionHint = (rm: any) => {
  if (!rm.product_id || !rm.quantity) return ''
  const prod = factoryStore.products.find((p) => p.id === rm.product_id)
  if (!prod) return ''
  const baseUnit = prod.unit || 'KG'
  const currentUnit = rm.unit || baseUnit
  if (currentUnit === baseUnit) return ''
  const converted = convertUnitQuantity(Number(rm.quantity), currentUnit, baseUnit, prod.pcs_per_unit || 1)
  return `Konversi: ${rm.quantity} ${currentUnit} ≈ ${converted.toLocaleString('id-ID', { maximumFractionDigits: 3 })} ${baseUnit} (Basis Stok)`
}

watch(
  () => productionForm.factory_id,
  (newVal) => {
    if (newVal) {
      factoryStore.fetchFactoryWarehouses(newVal)
    }
  }
)

const refreshData = async () => {
  await Promise.all([
    factoryStore.fetchFactories(),
    factoryStore.fetchProducts(),
    factoryStore.fetchAllInventory(),
    factoryStore.fetchAllProductionHistory(),
    factoryStore.fetchAllInventoryLogs(),
    masterStore.fetchEmployees(),
  ])
}

const syncingWarehouse = ref(false)
const handleSyncAllWarehouseStock = async () => {
  if (!confirm('Apakah Anda yakin ingin menyinkronkan seluruh stok pabrik 1:1 ke gudang cabang (Pabrik Utama & Cabang Timur)?')) {
    return
  }
  syncingWarehouse.value = true
  try {
    await factoryStore.syncAllWarehouseStock()
    await refreshData()
    toast.success('Seluruh stok pabrik dan stok gudang berhasil disinkronkan 1:1')
  } catch (err: any) {
    toast.error('Gagal menyinkronkan stok: ' + getErrorMessage(err))
  } finally {
    syncingWarehouse.value = false
  }
}

onMounted(refreshData)

const filteredStock = computed(() => {
  const query = (stockSearch.value || '').toLowerCase().trim()
  if (!query) return factoryStore.allInventory || []

  return (factoryStore.allInventory || []).filter((item) => {
    const name = (item.product?.name || '').toLowerCase()
    const sku = (item.product?.sku || '').toLowerCase()
    const batch = (item.batch_no || '').toLowerCase()
    const category = (item.product?.category || '').toLowerCase()
    const factory = (item.factory?.name || '').toLowerCase()
    const warehouse = (item.warehouse?.name || '').toLowerCase()

    return (
      name.includes(query) ||
      sku.includes(query) ||
      batch.includes(query) ||
      category.includes(query) ||
      factory.includes(query) ||
      warehouse.includes(query)
    )
  })
})

const totalFilteredQuantity = computed(() => {
  return filteredStock.value.reduce((sum, item) => sum + (Number(item.quantity) || 0), 0)
})

const handleSelectBatch = (batchNo: string) => {
  stockSearch.value = batchNo
}

const clearStockSearch = () => {
  stockSearch.value = ''
}

const syncQuantity = (item: any) => {
  if (item.carton_count > 0 && item.pieces_per_carton > 0) {
    item.quantity = item.carton_count * item.pieces_per_carton
  }
}

const addItem = () => {
  productionForm.items.push({
    product_id: '',
    target_warehouse_id: '',
    batch_no: '',
    expiry_date: '',
    quantity: 0,
    carton_count: 0,
    pieces_per_carton: 0,
  })
}

const removeItem = (idx: number) => {
  if (productionForm.items.length > 1) {
    productionForm.items.splice(idx, 1)
  }
}

const openAdjustStock = (row: any) => {
  selectedStock.value = row
  adjustStockForm.quantity = row.quantity
  adjustStockForm.reason = ''
  adjustStockForm.factory_warehouse_id = row.factory_warehouse_id || ''
  
  if (row.factory_id) {
    factoryStore.fetchFactoryWarehouses(row.factory_id)
  }
  
  showAdjustStock.value = true
}

const openDeleteStock = (row: any) => {
  selectedStockToDelete.value = row
  showDeleteStock.value = true
}

const closeDeleteStock = () => {
  if (deletingStock.value) return
  showDeleteStock.value = false
  selectedStockToDelete.value = null
}

const handleDeleteStock = async () => {
  if (!selectedStockToDelete.value?.id) return

  deletingStock.value = true
  try {
    await factoryApi.deleteStock(selectedStockToDelete.value.id)
    await Promise.all([
      factoryStore.fetchAllInventory(),
      factoryStore.fetchAllInventoryLogs(),
    ])
    showDeleteStock.value = false
    selectedStockToDelete.value = null
    toast.success('Stok berhasil dihapus')
  } catch (err) {
    toast.error('Gagal menghapus stok: ' + getErrorMessage(err))
  } finally {
    deletingStock.value = false
  }
}

const handleAdjustStockSubmit = async () => {
  submitting.value = true
  try {
    await factoryStore.adjustStock(selectedStock.value.factory_id, {
      product_id: selectedStock.value.product_id,
      quantity: adjustStockForm.quantity,
      reason: adjustStockForm.reason,
      batch_no: selectedStock.value.batch_no || 'DEFAULT',
      factory_warehouse_id: adjustStockForm.factory_warehouse_id || undefined,
    })
    await factoryStore.fetchAllInventory()
    await factoryStore.fetchAllInventoryLogs()
    showAdjustStock.value = false
    toast.success('Stok pabrik dan stok gudang berhasil disesuaikan 1:1')
  } catch (err) {
    toast.error('Gagal menyesuaikan stok: ' + getErrorMessage(err))
  } finally {
    submitting.value = false
  }
}

const openLogProduction = () => {
  isEditingProduction.value = false
  Object.assign(productionForm, {
    factory_id: '',
    employee_id: '',
    operators: [] as SelectedOperator[],
    notes: '',
    items: [
      {
        product_id: '',
        target_warehouse_id: '',
        batch_no: '',
        expiry_date: '',
        quantity: 0,
        carton_count: 0,
        pieces_per_carton: 0,
      }
    ],
    raw_materials: [],
  })
  operatorSearch.value = ''
  showOperatorDropdown.value = false
  if (masterStore.employees.length === 0) masterStore.fetchEmployees()
  showLogProduction.value = true
}

interface MaterialImpact {
  name: string
  product_id: string
  batch_no: string
  delta_qty: number
  unit: string
  type: 'RETURNED' | 'DEDUCTED' | 'UNCHANGED'
  notes?: string
}

interface StockImpactDiff {
  finishedGood: {
    name: string
    batch_no: string
    old_qty: number
    new_qty: number
    delta: number
    unit: string
  } | null
  materials: MaterialImpact[]
}

const showStockImpactModal = ref(false)
const editingSnapshot = ref<any>(null)
const stockImpactDiff = ref<StockImpactDiff>({
  finishedGood: null,
  materials: [],
})

const openEditProduction = (item: any) => {
  isEditingProduction.value = true
  currentProductionId.value = item.id
  editingSnapshot.value = JSON.parse(JSON.stringify(item))

  const loadedOperators: SelectedOperator[] = []
  if (item.operators && item.operators.length > 0) {
    item.operators.forEach((op: any) => {
      loadedOperators.push({
        employee_id: op.employee_id || undefined,
        operator_name: op.operator_name || (op.employee ? `${op.employee.first_name || ''} ${op.employee.last_name || ''}`.trim() : 'Operator'),
        is_manual: op.is_manual || !op.employee_id,
        employee_id_number: op.employee?.employee_id_number,
        position: op.employee?.job_position?.name || op.employee?.position,
      })
    })
  } else if (item.operator_name) {
    const names = item.operator_name.split(',').map((s: string) => s.trim()).filter(Boolean)
    if (names.length > 0) {
      names.forEach((name: string) => {
        const matchingEmp = masterStore.employees.find(
          (e: any) => `${e.first_name || ''} ${e.last_name || ''}`.trim().toLowerCase() === name.toLowerCase()
        )
        loadedOperators.push({
          employee_id: matchingEmp ? matchingEmp.id : (item.employee_id || undefined),
          operator_name: name,
          is_manual: !matchingEmp && !item.employee_id,
          employee_id_number: matchingEmp?.employee_id_number || (matchingEmp ? undefined : item.employee?.employee_id_number),
          position: matchingEmp?.job_position?.name || matchingEmp?.position,
        })
      })
    }
  } else if (item.employee) {
    loadedOperators.push({
      employee_id: item.employee_id,
      operator_name: `${item.employee.first_name || ''} ${item.employee.last_name || ''}`.trim(),
      is_manual: false,
      employee_id_number: item.employee.employee_id_number,
      position: item.employee.job_position?.name || item.employee.position,
    })
  }

  Object.assign(productionForm, {
    factory_id: item.factory_id,
    employee_id: item.employee_id || '',
    operators: loadedOperators,
    notes: item.notes || '',
    items: [
      {
        product_id: item.product_id,
        target_warehouse_id: item.target_warehouse_id || '',
        batch_no: item.batch_no,
        expiry_date: item.expiry_date ? new Date(item.expiry_date).toISOString().split('T')[0] : '',
        quantity: item.quantity,
        carton_count: item.carton_count || 0,
        pieces_per_carton: item.pieces_per_carton || 0,
      }
    ],
    raw_materials: (item.materials || []).map((m: any) => ({
      product_id: m.product_id,
      batch_no: m.batch_no || '',
      quantity: m.quantity,
      unit: m.product?.unit || 'KG',
      notes: m.notes || '',
    })),
  })
  operatorSearch.value = ''
  showOperatorDropdown.value = false
  if (masterStore.employees.length === 0) masterStore.fetchEmployees()
  showLogProduction.value = true
}

const computeStockImpactDiff = () => {
  const item = productionForm.items[0]
  const prod = factoryStore.products.find((p) => p.id === item?.product_id)
  const oldQty = Number(editingSnapshot.value?.quantity) || 0
  const newQty = Number(item?.quantity) || 0
  const fgDelta = newQty - oldQty

  const fgImpact = {
    name: prod?.name || editingSnapshot.value?.product?.name || 'Produk Jadi',
    batch_no: item?.batch_no || editingSnapshot.value?.batch_no || '-',
    old_qty: oldQty,
    new_qty: newQty,
    delta: fgDelta,
    unit: prod?.unit || editingSnapshot.value?.product?.unit || 'PCS',
  }

  // Raw Materials Diff
  const oldMaterials = (editingSnapshot.value?.materials || []).map((m: any) => ({
    product_id: m.product_id,
    batch_no: m.batch_no || 'AUTO',
    quantity: Number(m.quantity) || 0,
    unit: m.product?.unit || 'KG',
    name: m.product?.name || 'Bahan Baku',
    notes: m.notes || '',
  }))

  const newMaterials = (productionForm.raw_materials || [])
    .filter((rm: any) => rm.product_id && Number(rm.quantity) > 0)
    .map((rm: any) => {
      const p = factoryStore.products.find((prod) => prod.id === rm.product_id)
      const baseUnit = p?.unit || 'KG'
      const inputUnit = rm.unit || baseUnit
      const converted = convertUnitQuantity(Number(rm.quantity), inputUnit, baseUnit, p?.pcs_per_unit || 1)
      const finalQty = Math.max(1, Math.round(converted))
      return {
        product_id: rm.product_id,
        batch_no: rm.batch_no || 'AUTO',
        quantity: finalQty,
        unit: baseUnit,
        name: p?.name || 'Bahan Baku',
        notes: rm.notes || (inputUnit !== baseUnit ? `Input: ${rm.quantity} ${inputUnit}` : ''),
      }
    })

  const materialImpacts: MaterialImpact[] = []
  const matchedNewIndices = new Set<number>()

  for (const oldMat of oldMaterials) {
    const newIdx = newMaterials.findIndex(
      (nm, idx) => !matchedNewIndices.has(idx) && nm.product_id === oldMat.product_id && nm.batch_no === oldMat.batch_no
    )

    if (newIdx !== -1) {
      matchedNewIndices.add(newIdx)
      const newMat = newMaterials[newIdx]
      if (!newMat) continue
      const diff = newMat.quantity - oldMat.quantity
      if (diff > 0) {
        materialImpacts.push({
          name: newMat.name,
          product_id: newMat.product_id,
          batch_no: newMat.batch_no,
          delta_qty: diff,
          unit: newMat.unit,
          type: 'DEDUCTED',
          notes: `Kebutuhan bertambah dari ${oldMat.quantity} ke ${newMat.quantity} ${newMat.unit}`,
        })
      } else if (diff < 0) {
        materialImpacts.push({
          name: newMat.name,
          product_id: newMat.product_id,
          batch_no: newMat.batch_no,
          delta_qty: Math.abs(diff),
          unit: newMat.unit,
          type: 'RETURNED',
          notes: `Kebutuhan berkurang dari ${oldMat.quantity} ke ${newMat.quantity} ${newMat.unit}`,
        })
      } else {
        materialImpacts.push({
          name: newMat.name,
          product_id: newMat.product_id,
          batch_no: newMat.batch_no,
          delta_qty: 0,
          unit: newMat.unit,
          type: 'UNCHANGED',
          notes: `Kuantitas tetap (${newMat.quantity} ${newMat.unit})`,
        })
      }
    } else {
      materialImpacts.push({
        name: oldMat.name,
        product_id: oldMat.product_id,
        batch_no: oldMat.batch_no,
        delta_qty: oldMat.quantity,
        unit: oldMat.unit,
        type: 'RETURNED',
        notes: `Dihapus dari pemakaian log`,
      })
    }
  }

  newMaterials.forEach((newMat, idx) => {
    if (!matchedNewIndices.has(idx)) {
      materialImpacts.push({
        name: newMat.name,
        product_id: newMat.product_id,
        batch_no: newMat.batch_no,
        delta_qty: newMat.quantity,
        unit: newMat.unit,
        type: 'DEDUCTED',
        notes: `Bahan baru ditambahkan (${newMat.quantity} ${newMat.unit})`,
      })
    }
  })

  stockImpactDiff.value = {
    finishedGood: fgImpact,
    materials: materialImpacts,
  }
}

const handleProductionSubmit = async () => {
  if (submitting.value) return

  // Auto-add text in operatorSearch if operator list is currently empty
  if (productionForm.operators.length === 0 && operatorSearch.value.trim()) {
    addManualOperator()
  }

  if (!productionForm.factory_id || productionForm.operators.length === 0 || productionForm.items.length === 0) {
    toast.error('Lengkapi pabrik, minimal 1 operator, dan produk produksi!')
    return
  }
  for (const item of productionForm.items) {
    if (!item.product_id || Number(item.quantity) <= 0) {
      toast.error('Lengkapi produk dan quantity untuk semua item!')
      return
    }
  }

  if (isEditingProduction.value) {
    computeStockImpactDiff()
    showStockImpactModal.value = true
    return
  }

  submitting.value = true
  try {
    const operatorsPayload = productionForm.operators.map((op) => ({
      employee_id: op.employee_id || undefined,
      operator_name: op.operator_name,
      is_manual: op.is_manual,
    }))

    const firstEmpId = productionForm.operators.find((op) => op.employee_id)?.employee_id || undefined
    const consolidatedNames = productionForm.operators.map((op) => op.operator_name).join(', ')

    const rawMaterialsPayload = (productionForm.raw_materials || [])
      .filter((rm: any) => rm.product_id && Number(rm.quantity) > 0)
      .map((rm: any) => {
        const prod = factoryStore.products.find((p) => p.id === rm.product_id)
        const baseUnit = prod?.unit || 'KG'
        const inputUnit = rm.unit || baseUnit
        const converted = convertUnitQuantity(Number(rm.quantity), inputUnit, baseUnit, prod?.pcs_per_unit || 1)
        const finalQty = Math.max(1, Math.round(converted))
        return {
          product_id: rm.product_id,
          batch_no: rm.batch_no || '',
          quantity: finalQty,
          unit: baseUnit,
          notes: rm.notes || (inputUnit !== baseUnit ? `Input: ${rm.quantity} ${inputUnit}` : ''),
        }
      })

    const itemsPayload = productionForm.items.map((item) => ({
      product_id: item.product_id,
      target_warehouse_id: item.target_warehouse_id || undefined,
      batch_no: item.batch_no || '',
      expiry_date: item.expiry_date || null,
      quantity: Number(item.quantity),
      carton_count: item.carton_count || 0,
      pieces_per_carton: item.pieces_per_carton || 0,
    }))

    const batchPayload = {
      employee_id: firstEmpId,
      operators: operatorsPayload,
      operator_name: consolidatedNames,
      notes: productionForm.notes,
      items: itemsPayload,
      raw_materials: rawMaterialsPayload,
    }

    await factoryStore.logBatchProduction(productionForm.factory_id, batchPayload)
    toast.success('Log produksi berhasil dicatat')
    await refreshData()
    showLogProduction.value = false
  } catch (err) {
    toast.error('Gagal menyimpan log produksi: ' + getErrorMessage(err))
  } finally {
    submitting.value = false
  }
}

const executeProductionUpdate = async () => {
  if (submitting.value) return
  submitting.value = true
  try {
    const operatorsPayload = productionForm.operators.map((op) => ({
      employee_id: op.employee_id || undefined,
      operator_name: op.operator_name,
      is_manual: op.is_manual,
    }))

    const firstEmpId = productionForm.operators.find((op) => op.employee_id)?.employee_id || undefined
    const consolidatedNames = productionForm.operators.map((op) => op.operator_name).join(', ')
    const item = productionForm.items[0]

    if (!item) {
      toast.error('Item produksi tidak ditemukan!')
      submitting.value = false
      return
    }

    const rawMaterialsPayload = (productionForm.raw_materials || [])
      .filter((rm: any) => rm.product_id && Number(rm.quantity) > 0)
      .map((rm: any) => {
        const prod = factoryStore.products.find((p) => p.id === rm.product_id)
        const baseUnit = prod?.unit || 'KG'
        const inputUnit = rm.unit || baseUnit
        const converted = convertUnitQuantity(Number(rm.quantity), inputUnit, baseUnit, prod?.pcs_per_unit || 1)
        const finalQty = Math.max(1, Math.round(converted))
        return {
          product_id: rm.product_id,
          batch_no: rm.batch_no || '',
          quantity: finalQty,
          unit: baseUnit,
          notes: rm.notes || (inputUnit !== baseUnit ? `Input: ${rm.quantity} ${inputUnit}` : ''),
        }
      })

    const payload = {
      factory_id: productionForm.factory_id,
      employee_id: firstEmpId,
      operators: operatorsPayload,
      operator_name: consolidatedNames,
      notes: productionForm.notes,
      product_id: item.product_id,
      target_warehouse_id: item.target_warehouse_id || undefined,
      batch_no: item.batch_no,
      expiry_date: item.expiry_date || null,
      quantity: Number(item.quantity),
      carton_count: item.carton_count,
      pieces_per_carton: item.pieces_per_carton,
      raw_materials: rawMaterialsPayload,
    }

    await factoryStore.updateProductionLog(
      productionForm.factory_id,
      currentProductionId.value,
      payload,
    )
    toast.success('Log produksi & penyesuaian stok berhasil diperbarui')
    showStockImpactModal.value = false
    showLogProduction.value = false
    await refreshData()
  } catch (err) {
    toast.error('Gagal memperbarui log produksi: ' + getErrorMessage(err))
  } finally {
    submitting.value = false
  }
}

const handleDeleteProduction = async (row: any) => {
  if (!confirm('Hapus log produksi global ini? Stok di pabrik terkait akan dikurangi kembali.'))
    return
  try {
    await factoryStore.deleteProductionLog(row.factory_id, row.id)
    toast.success('Log produksi berhasil dihapus')
  } catch (err: any) {
    toast.error('Gagal menghapus log produksi: ' + getErrorMessage(err))
  }
}

const openTransferHistory = async (row: any) => {
  selectedProductionLog.value = row
  showTransferHistory.value = true
  transferHistoryLoading.value = true
  transferHistoryData.value = []

  try {
    const res = await factoryApi.getAllTransfers({
      product_id: row.product_id,
      batch_no: row.batch_no,
    })
    transferHistoryData.value = res.data?.data || res.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil histori pengiriman: ' + getErrorMessage(err))
  } finally {
    transferHistoryLoading.value = false
  }
}

const stockColumns = [
  { key: 'factory_name', label: 'Lokasi Pabrik' },
  { key: 'warehouse_name', label: 'Lokasi Gudang' },
  { key: 'product.name', label: 'Produk' },
  { key: 'batch_info', label: 'Batch & Exp' },
  { key: 'quantity', label: 'Stok' },
  { key: 'actions', label: '' },
]

const productionColumns = computed(() => {
  const cols = [
    { key: 'production_date', label: 'Tanggal', type: 'datetime' },
    { key: 'factory_name', label: 'Pabrik' },
    { key: 'product.name', label: 'Produk' },
    { key: 'batch_info', label: 'Batch & Exp' },
    { key: 'quantity', label: 'Qty' },
    { key: 'operators_info', label: 'Operator Produksi' },
  ]
  if (canViewMaterials.value) {
    cols.push({ key: 'materials', label: 'Bahan Baku & Batch Terpakai' })
  }
  cols.push({ key: 'actions', label: '' })
  return cols
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 5px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 10px;
}
</style>
