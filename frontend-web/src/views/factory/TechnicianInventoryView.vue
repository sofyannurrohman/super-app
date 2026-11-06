<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-black tracking-tight text-slate-900 flex items-center gap-2">
          <Wrench class="w-7 h-7 text-amber-600" />
          Manajemen Stok Barang Teknisi & Sparepart
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Kelola persediaan mur & baut, suku cadang mesin, pelumas, alat kerja teknisi, pencatatan pemakaian perbaikan, dan stock opname dinamis.
        </p>
      </div>

      <!-- Actions & Factory Selector -->
      <div class="flex flex-wrap items-center gap-2.5">
        <!-- Factory Selector -->
        <select
          v-model="activeFactoryId"
          @change="onFactoryChange"
          class="h-10 rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-bold text-slate-700 shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20"
        >
          <option v-for="f in factories" :key="f.id" :value="f.id">
            🏭 {{ f.name }}
          </option>
        </select>

        <Button
          variant="outline"
          size="sm"
          @click="fetchData"
          :disabled="loading"
          class="h-10 rounded-xl border-slate-200 text-xs font-bold text-slate-700 hover:bg-slate-100 shadow-sm"
        >
          <RefreshCw class="w-4 h-4 mr-1.5" :class="{ 'animate-spin': loading }" />
          Segarkan
        </Button>

        <Button
          variant="outline"
          size="sm"
          @click="openLogsModal()"
          class="h-10 rounded-xl border-slate-200 text-xs font-bold text-slate-700 hover:bg-slate-100 shadow-sm"
        >
          <History class="w-4 h-4 mr-1.5 text-blue-600" />
          Riwayat Mutasi
        </Button>

        <Button
          size="sm"
          @click="openUsageModal"
          class="h-10 rounded-xl bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700 text-white font-bold text-xs shadow-md shadow-amber-600/20"
        >
          <Hammer class="w-4 h-4 mr-1.5" />
          Catat Pemakaian Mesin
        </Button>

        <Button
          size="sm"
          @click="openCreateModal"
          class="h-10 rounded-xl bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs shadow-md"
        >
          <Plus class="w-4 h-4 mr-1.5" />
          + Barang Baru
        </Button>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="p-5 rounded-2xl bg-white border border-slate-100 shadow-sm space-y-1">
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-xs font-bold uppercase tracking-wider">Total Item Sparepart</span>
          <div class="p-2 rounded-xl bg-amber-50 text-amber-600">
            <Boxes class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-900">{{ items.length }}</div>
        <p class="text-[11px] text-slate-500 font-medium">Item terdaftar di master teknisi</p>
      </div>

      <div class="p-5 rounded-2xl bg-white border border-slate-100 shadow-sm space-y-1">
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-xs font-bold uppercase tracking-wider">Peringatan Stok Menipis</span>
          <div class="p-2 rounded-xl bg-red-50 text-red-600">
            <AlertTriangle class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-red-600">{{ lowStockItemsCount }}</div>
        <p class="text-[11px] text-red-500 font-medium">Stok &le; batas minimum stok</p>
      </div>

      <div class="p-5 rounded-2xl bg-white border border-slate-100 shadow-sm space-y-1">
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-xs font-bold uppercase tracking-wider">Total Kategori</span>
          <div class="p-2 rounded-xl bg-blue-50 text-blue-600">
            <Layers class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-900">{{ activeCategoriesCount }}</div>
        <p class="text-[11px] text-slate-500 font-medium">Mur baut, kelistrikan, tools, dll</p>
      </div>

      <div class="p-5 rounded-2xl bg-white border border-slate-100 shadow-sm space-y-1">
        <div class="flex items-center justify-between text-slate-400">
          <span class="text-xs font-bold uppercase tracking-wider">Total Nilai Persediaan</span>
          <div class="p-2 rounded-xl bg-emerald-50 text-emerald-600">
            <Coins class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-emerald-700">Rp {{ totalInventoryValue.toLocaleString('id-ID') }}</div>
        <p class="text-[11px] text-slate-500 font-medium">Estimasi nilai stok sparepart</p>
      </div>
    </div>

    <!-- Filters & Search -->
    <div class="p-4 rounded-2xl bg-white border border-slate-100 shadow-sm space-y-3">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
        <!-- Search -->
        <div class="relative flex-1 max-w-md">
          <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <Input
            v-model="searchQuery"
            placeholder="Cari SKU, nama barang, rak, atau spesifikasi..."
            class="pl-9 h-10 rounded-xl border-slate-200 text-xs bg-slate-50/50 focus:bg-white transition-all"
          />
        </div>

        <!-- Filter Category & Low stock toggle -->
        <div class="flex flex-wrap items-center gap-2">
          <select
            v-model="selectedCategory"
            class="h-10 rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-semibold text-slate-700 focus:outline-none focus:ring-2 focus:ring-amber-500/20"
          >
            <option value="">Semua Kategori</option>
            <option v-for="cat in CATEGORIES" :key="cat.value" :value="cat.value">
              {{ cat.label }}
            </option>
          </select>

          <Button
            type="button"
            variant="outline"
            size="sm"
            @click="filterLowStockOnly = !filterLowStockOnly"
            class="h-10 rounded-xl border-slate-200 text-xs font-bold transition-colors"
            :class="filterLowStockOnly ? 'bg-red-50 text-red-700 border-red-200' : 'text-slate-600 hover:bg-slate-50'"
          >
            <AlertCircle class="w-3.5 h-3.5 mr-1.5" :class="filterLowStockOnly ? 'text-red-600' : 'text-slate-400'" />
            Stok Menipis Saja
            <span v-if="lowStockItemsCount > 0" class="ml-1.5 px-1.5 py-0.5 rounded-full text-[10px] bg-red-100 text-red-700 font-mono">
              {{ lowStockItemsCount }}
            </span>
          </Button>
        </div>
      </div>
    </div>

    <!-- Items Table -->
    <div class="rounded-2xl bg-white border border-slate-100 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-100 bg-slate-50/70 text-[11px] font-black uppercase tracking-wider text-slate-500">
              <th class="py-3 px-4">SKU / Kode</th>
              <th class="py-3 px-4">Nama Barang & Spesifikasi</th>
              <th class="py-3 px-4">Kategori</th>
              <th class="py-3 px-4">Lokasi Rak</th>
              <th class="py-3 px-4 text-center">Stok Saat Ini</th>
              <th class="py-3 px-4 text-center">Min. Stok</th>
              <th class="py-3 px-4 text-right">Est. Harga Satuan</th>
              <th class="py-3 px-4 text-center">Aksi Cepat</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs">
            <tr v-if="loading" class="text-center">
              <td colspan="8" class="py-12 text-slate-400">
                <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2 text-amber-600" />
                Memuat data stok barang teknisi...
              </td>
            </tr>

            <tr v-else-if="filteredItems.length === 0" class="text-center">
              <td colspan="8" class="py-12 text-slate-400">
                <Wrench class="w-10 h-10 mx-auto mb-2 text-slate-300 stroke-1" />
                Tidak ada data barang teknisi yang cocok dengan pencarian.
              </td>
            </tr>

            <tr
              v-for="item in filteredItems"
              :key="item.id"
              class="hover:bg-slate-50/60 transition-colors"
            >
              <!-- SKU -->
              <td class="py-3 px-4 font-mono font-bold text-slate-900">
                <span class="px-2 py-1 rounded-lg bg-slate-100 text-slate-800 border border-slate-200">
                  {{ item.sku }}
                </span>
              </td>

              <!-- Name & Spec -->
              <td class="py-3 px-4">
                <div class="font-bold text-slate-900 text-sm">{{ item.name }}</div>
                <div v-if="item.specification" class="text-[11px] text-slate-500 line-clamp-1">
                  {{ item.specification }}
                </div>
              </td>

              <!-- Category -->
              <td class="py-3 px-4">
                <span
                  class="px-2.5 py-1 rounded-full text-[10px] font-extrabold"
                  :class="getCategoryBadgeClass(item.category)"
                >
                  {{ getCategoryLabel(item.category) }}
                </span>
              </td>

              <!-- Rack Location -->
              <td class="py-3 px-4">
                <div class="flex items-center gap-1.5 text-slate-600 font-mono text-[11px]">
                  <MapPin class="w-3.5 h-3.5 text-slate-400" />
                  <span>{{ item.rack_location || '-' }}</span>
                </div>
              </td>

              <!-- Current Stock -->
              <td class="py-3 px-4 text-center">
                <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl font-mono font-black text-sm"
                  :class="getStockStatusClass(item.current_stock, item.minimum_stock)"
                >
                  <span>{{ item.current_stock }}</span>
                  <span class="text-[10px] uppercase font-bold opacity-80">{{ item.unit }}</span>
                </div>
                <div v-if="item.current_stock <= item.minimum_stock" class="text-[10px] font-bold text-red-600 mt-0.5">
                  ⚠️ Stok Menipis
                </div>
              </td>

              <!-- Min Stock -->
              <td class="py-3 px-4 text-center font-mono text-slate-500 font-medium">
                {{ item.minimum_stock }} {{ item.unit }}
              </td>

              <!-- Price -->
              <td class="py-3 px-4 text-right font-mono font-semibold text-slate-700">
                Rp {{ (item.unit_price || 0).toLocaleString('id-ID') }}
              </td>

              <!-- Actions -->
              <td class="py-3 px-4 text-center">
                <div class="flex items-center justify-center gap-1">
                  <!-- Stock Opname Action -->
                  <Button
                    size="sm"
                    variant="outline"
                    @click="openOpnameModal(item)"
                    title="Stock Opname Cepat"
                    class="h-8 px-2 rounded-lg border-amber-200 text-amber-700 bg-amber-50 hover:bg-amber-100 font-bold text-[11px]"
                  >
                    <ClipboardCheck class="w-3.5 h-3.5 mr-1" />
                    Opname
                  </Button>

                  <!-- Logs Action -->
                  <Button
                    size="sm"
                    variant="ghost"
                    @click="openLogsModal(item)"
                    title="Lihat Riwayat Mutasi"
                    class="h-8 w-8 p-0 rounded-lg text-slate-500 hover:text-blue-600 hover:bg-blue-50"
                  >
                    <History class="w-4 h-4" />
                  </Button>

                  <!-- Edit Action -->
                  <Button
                    size="sm"
                    variant="ghost"
                    @click="openEditModal(item)"
                    title="Edit Data Barang"
                    class="h-8 w-8 p-0 rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-100"
                  >
                    <Edit3 class="w-4 h-4" />
                  </Button>

                  <!-- Delete Action -->
                  <Button
                    size="sm"
                    variant="ghost"
                    @click="confirmDelete(item)"
                    title="Hapus Barang"
                    class="h-8 w-8 p-0 rounded-lg text-slate-400 hover:text-red-600 hover:bg-red-50"
                  >
                    <Trash2 class="w-4 h-4" />
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 1. DIALOG CREATE / EDIT ITEM -->
    <Dialog :open="showItemDialog" @update:open="showItemDialog = $event">
      <DialogContent class="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2 text-lg font-black text-slate-900">
            <Wrench class="w-5 h-5 text-amber-600" />
            {{ editingItem ? 'Edit Data Barang Teknisi' : 'Tambah Barang Teknisi Baru' }}
          </DialogTitle>
          <DialogDescription>
            Isi detail informasi suku cadang, mur baut, pelumas, atau alat kerja pabrik.
          </DialogDescription>
        </DialogHeader>

        <form @submit.prevent="saveItem" class="space-y-4 mt-2">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Kode SKU / Part No. *</Label>
              <Input
                v-model="itemForm.sku"
                placeholder="Misal: BAUT-M8-20, SP-BEARING-6204"
                class="rounded-xl border-slate-200 font-mono uppercase"
                required
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Kategori Barang *</Label>
              <select
                v-model="itemForm.category"
                class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-amber-500/20"
                required
              >
                <option v-for="c in CATEGORIES" :key="c.value" :value="c.value">
                  {{ c.label }}
                </option>
              </select>
            </div>

            <div class="sm:col-span-2 space-y-1">
              <Label class="text-xs font-bold text-slate-700">Nama Barang / Sparepart *</Label>
              <Input
                v-model="itemForm.name"
                placeholder="Misal: Baut Baja M8 x 20mm Grade 8.8"
                class="rounded-xl border-slate-200"
                required
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Satuan (UoM) *</Label>
              <input
                list="tech-uom-options"
                v-model="itemForm.unit"
                placeholder="PCS / SET / LTR / DUS"
                class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-bold uppercase focus:outline-none focus:ring-2 focus:ring-amber-500/20"
                required
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Lokasi Rak / Lemari</Label>
              <Input
                v-model="itemForm.rack_location"
                placeholder="Misal: RAK-A1, KOTAK-03"
                class="rounded-xl border-slate-200 font-mono"
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Batas Minimum Stok *</Label>
              <Input
                type="number"
                v-model.number="itemForm.minimum_stock"
                min="0"
                placeholder="10"
                class="rounded-xl border-slate-200 font-mono"
                required
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Estimasi Harga Satuan (Rp)</Label>
              <Input
                type="number"
                v-model.number="itemForm.unit_price"
                min="0"
                placeholder="0"
                class="rounded-xl border-slate-200 font-mono"
              />
            </div>

            <!-- Initial stock only on create -->
            <div v-if="!editingItem" class="sm:col-span-2 p-3 bg-amber-50 rounded-xl border border-amber-200 space-y-1">
              <Label class="text-xs font-bold text-amber-900">Stok Awal Fisik Saat Ini</Label>
              <Input
                type="number"
                v-model.number="itemForm.current_stock"
                min="0"
                placeholder="0"
                class="rounded-xl border-amber-300 font-mono font-bold bg-white text-amber-900"
              />
              <p class="text-[10px] text-amber-700">Stok awal akan dicatat sebagai saldo awal master teknisi.</p>
            </div>

            <div class="sm:col-span-2 space-y-1">
              <Label class="text-xs font-bold text-slate-700">Spesifikasi Teknis / Ukuran</Label>
              <textarea
                v-model="itemForm.specification"
                rows="2"
                placeholder="Misal: Material Stainless 304, Pitch 1.25mm..."
                class="flex w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-amber-500/20"
              ></textarea>
            </div>
          </div>

          <DialogFooter class="gap-2 sm:gap-0 pt-2">
            <Button type="button" variant="outline" class="rounded-xl font-bold" @click="showItemDialog = false">
              Batal
            </Button>
            <Button
              type="submit"
              :disabled="saving"
              class="rounded-xl font-bold bg-amber-600 hover:bg-amber-700 text-white shadow-md shadow-amber-600/20"
            >
              <Loader2 v-if="saving" class="w-4 h-4 mr-1.5 animate-spin" />
              Simpan Data Barang
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- 2. DIALOG QUICK STOCK OPNAME -->
    <Dialog :open="showOpnameDialog" @update:open="showOpnameDialog = $event">
      <DialogContent class="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2 text-lg font-black text-slate-900">
            <ClipboardCheck class="w-5 h-5 text-amber-600" />
            Stock Opname Cepat
          </DialogTitle>
          <DialogDescription>
            Sesuaikan jumlah stok sistem dengan hasil penghitungan fisik teknisi di lemari/rak.
          </DialogDescription>
        </DialogHeader>

        <div v-if="selectedOpnameItem" class="space-y-4 mt-2">
          <!-- Item Info Card -->
          <div class="p-3.5 rounded-2xl bg-slate-50 border border-slate-200/80 space-y-1.5">
            <div class="flex items-center justify-between">
              <span class="font-mono font-bold text-xs text-slate-500">SKU: {{ selectedOpnameItem.sku }}</span>
              <span class="px-2 py-0.5 rounded-full text-[10px] font-extrabold bg-slate-200 text-slate-700">
                {{ selectedOpnameItem.rack_location || 'Rak Teknisi' }}
              </span>
            </div>
            <div class="text-sm font-black text-slate-900">{{ selectedOpnameItem.name }}</div>
            <div class="flex items-center justify-between text-xs pt-1 border-t border-slate-200/60">
              <span class="text-slate-500">Stok Tercatat di Sistem:</span>
              <span class="font-mono font-bold text-slate-900">{{ selectedOpnameItem.current_stock }} {{ selectedOpnameItem.unit }}</span>
            </div>
          </div>

          <form @submit.prevent="submitOpname" class="space-y-3">
            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Hasil Hitung Fisik Sebenarnya *</Label>
              <Input
                type="number"
                v-model.number="opnameForm.physical_quantity"
                min="0"
                class="rounded-xl border-amber-300 text-base font-mono font-black text-amber-900 bg-amber-50/50"
                required
              />
            </div>

            <!-- Calculated Difference -->
            <div class="p-3 rounded-xl border flex items-center justify-between text-xs"
              :class="opnameDiff >= 0 ? 'bg-emerald-50 border-emerald-200 text-emerald-900' : 'bg-red-50 border-red-200 text-red-900'"
            >
              <span class="font-bold">Selisih Stok (Adjustment):</span>
              <span class="font-mono font-black text-sm">
                {{ opnameDiff > 0 ? `+${opnameDiff}` : opnameDiff }} {{ selectedOpnameItem.unit }}
              </span>
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Alasan Penyesuaian *</Label>
              <select
                v-model="opnameForm.reason"
                class="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-amber-500/20"
                required
              >
                <option value="STOCK_OPNAME_REGULER">Stock Opname Rutin / Berkala</option>
                <option value="SELISIH_HITUNG">Koreksi Selisih Hitung</option>
                <option value="BARANG_RUSAK">Barang Rusak / Afkir</option>
                <option value="BARANG_HILANG">Barang Hilang / Tidak Ditemukan</option>
                <option value="BARANG_DITEMUKAN">Barang Lebih / Ditemukan</option>
                <option value="LAINNYA">Lainnya</option>
              </select>
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Catatan / Keterangan Tambahan</Label>
              <Input
                v-model="opnameForm.notes"
                placeholder="Contoh: Dihitung ulang bersama teknisi Mas Budi..."
                class="rounded-xl border-slate-200 text-xs"
              />
            </div>

            <DialogFooter class="gap-2 sm:gap-0 pt-3">
              <Button type="button" variant="outline" class="rounded-xl font-bold" @click="showOpnameDialog = false">
                Batal
              </Button>
              <Button
                type="submit"
                :disabled="saving"
                class="rounded-xl font-bold bg-amber-600 hover:bg-amber-700 text-white shadow-md shadow-amber-600/20"
              >
                <Loader2 v-if="saving" class="w-4 h-4 mr-1.5 animate-spin" />
                Simpan Penyesuaian Stok
              </Button>
            </DialogFooter>
          </form>
        </div>
      </DialogContent>
    </Dialog>

    <!-- 3. DIALOG RECORD USAGE (PEMAKAIAN PERBAIKAN MESIN) -->
    <Dialog :open="showUsageDialog" @update:open="showUsageDialog = $event">
      <DialogContent class="sm:max-w-[800px] max-h-[92vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2 text-lg font-black text-slate-900">
            <div class="p-2 rounded-xl bg-orange-100 text-orange-700">
              <Hammer class="w-5 h-5" />
            </div>
            <span>Catat Pengeluaran / Pemakaian Barang Teknisi</span>
          </DialogTitle>
          <DialogDescription>
            Pilih suku cadang, mur & baut, atau material teknisi yang digunakan untuk perbaikan mesin pabrik.
          </DialogDescription>
        </DialogHeader>

        <form @submit.prevent="submitUsage" class="space-y-4 mt-2">
          <!-- General Info Card -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 p-4 bg-gradient-to-br from-orange-50/60 to-amber-50/40 rounded-2xl border border-orange-200/70 shadow-2xs">
            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Nama Teknisi / Petugas *</Label>
              <Input
                v-model="usageForm.technician_name"
                placeholder="Misal: Agus Santoso (Teknisi Mesin)"
                class="rounded-xl border-orange-200/80 bg-white text-xs h-9 focus:border-orange-500"
                required
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Mesin / Lokasi Tujuan Perbaikan *</Label>
              <Input
                v-model="usageForm.machine_or_purpose"
                placeholder="Misal: Mesin Extruder 02, Conveyor Jalur 1"
                class="rounded-xl border-orange-200/80 bg-white text-xs h-9 focus:border-orange-500"
                required
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Tanggal Pemakaian *</Label>
              <Input
                type="date"
                v-model="usageForm.usage_date"
                class="rounded-xl border-orange-200/80 bg-white text-xs h-9 focus:border-orange-500"
                required
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs font-bold text-slate-700">Catatan Umum Pekerjaan</Label>
              <Input
                v-model="usageForm.notes"
                placeholder="Misal: Penggantian berkala sparepart aus..."
                class="rounded-xl border-orange-200/80 bg-white text-xs h-9 focus:border-orange-500"
              />
            </div>
          </div>

          <!-- Items Used Section -->
          <div class="space-y-2.5">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Label class="text-xs font-black uppercase tracking-wider text-slate-800">
                  Rincian Barang yang Digunakan
                </Label>
                <span class="px-2 py-0.5 rounded-full text-[10px] font-black bg-orange-100 text-orange-800">
                  {{ usageForm.items.length }} Item
                </span>
              </div>

              <Button
                type="button"
                variant="outline"
                size="sm"
                @click="addUsageRow"
                class="h-8 rounded-xl border-orange-200 text-xs font-bold text-orange-700 hover:bg-orange-50 hover:border-orange-300 transition-all shadow-2xs"
              >
                <Plus class="w-3.5 h-3.5 mr-1" /> Tambah Baris
              </Button>
            </div>

            <!-- Items List Cards -->
            <div class="space-y-2.5">
              <div
                v-for="(row, idx) in usageForm.items"
                :key="idx"
                class="p-3 bg-white border rounded-2xl transition-all shadow-2xs space-y-2.5"
                :class="
                  getTechnicianItemById(row.technician_item_id) &&
                  row.quantity > (getTechnicianItemById(row.technician_item_id)?.current_stock || 0)
                    ? 'border-red-300 bg-red-50/20'
                    : 'border-slate-200/90 hover:border-slate-300'
                "
              >
                <!-- Row Header: Number & Remove button -->
                <div class="flex items-center justify-between pb-1.5 border-b border-slate-100">
                  <span class="inline-flex items-center px-2 py-0.5 rounded-md bg-slate-100 text-slate-700 text-[10px] font-black tracking-wider">
                    ITEM #{{ idx + 1 }}
                  </span>

                  <button
                    v-if="usageForm.items.length > 1"
                    type="button"
                    @click="removeUsageRow(idx)"
                    class="text-slate-400 hover:text-red-600 p-1 rounded-lg hover:bg-red-50 transition-colors"
                    title="Hapus baris ini"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>

                <!-- Fields: Item Search Select, Quantity with Unit, Notes -->
                <div class="grid grid-cols-1 sm:grid-cols-12 gap-3 items-start">
                  <!-- Searchable Item Selector -->
                  <div class="sm:col-span-6 space-y-1">
                    <Label class="text-[10px] font-bold text-slate-600">Pilih Barang Teknisi *</Label>
                    <TechnicianItemSearchSelect
                      v-model="row.technician_item_id"
                      :items="items"
                      placeholder="Ketik SKU atau nama sparepart..."
                    />
                  </div>

                  <!-- Quantity Input with Unit Label -->
                  <div class="sm:col-span-3 space-y-1">
                    <Label class="text-[10px] font-bold text-slate-600">Jumlah Keluar *</Label>
                    <div class="relative">
                      <Input
                        type="number"
                        v-model.number="row.quantity"
                        min="1"
                        placeholder="1"
                        class="h-[42px] pr-12 text-xs rounded-xl font-mono font-black"
                        :class="
                          getTechnicianItemById(row.technician_item_id) &&
                          row.quantity > (getTechnicianItemById(row.technician_item_id)?.current_stock || 0)
                            ? 'border-red-400 bg-red-50/50 text-red-900 focus:ring-red-200'
                            : 'border-slate-200 bg-white'
                        "
                        required
                      />
                      <span
                        v-if="getTechnicianItemById(row.technician_item_id)"
                        class="absolute right-2.5 top-1/2 -translate-y-1/2 text-[10px] font-black uppercase text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded"
                      >
                        {{ getTechnicianItemById(row.technician_item_id)?.unit || 'UNIT' }}
                      </span>
                    </div>

                    <!-- Stock Warning / Info -->
                    <div v-if="getTechnicianItemById(row.technician_item_id)" class="text-[10px]">
                      <span
                        v-if="row.quantity > (getTechnicianItemById(row.technician_item_id)?.current_stock || 0)"
                        class="text-red-600 font-bold flex items-center gap-1"
                      >
                        <AlertTriangle class="w-3 h-3 shrink-0" />
                        Melebihi stok (sisa: {{ getTechnicianItemById(row.technician_item_id)?.current_stock }} {{ getTechnicianItemById(row.technician_item_id)?.unit }})
                      </span>
                      <span
                        v-else-if="(getTechnicianItemById(row.technician_item_id)?.current_stock || 0) <= 0"
                        class="text-red-600 font-bold"
                      >
                        ⚠️ Stok barang saat ini 0
                      </span>
                      <span v-else class="text-slate-500 font-medium">
                        Tersedia: <strong>{{ getTechnicianItemById(row.technician_item_id)?.current_stock }}</strong> {{ getTechnicianItemById(row.technician_item_id)?.unit }}
                      </span>
                    </div>
                  </div>

                  <!-- Row Specific Notes -->
                  <div class="sm:col-span-3 space-y-1">
                    <Label class="text-[10px] font-bold text-slate-600">Keterangan Khusus</Label>
                    <Input
                      v-model="row.notes"
                      placeholder="Misal: Aus / patah"
                      class="h-[42px] text-xs rounded-xl border-slate-200 bg-white"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- Summary & Validation Info Bar -->
            <div class="p-3 bg-slate-50 border border-slate-200/80 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs">
              <div class="flex items-center gap-4 text-slate-600">
                <span>Total Item Dipilih: <strong class="text-slate-900 font-bold">{{ usageSelectedItemsCount }}</strong> jenis</span>
                <span>Total Kuantitas: <strong class="text-slate-900 font-mono font-bold">{{ usageTotalQty }}</strong> unit</span>
              </div>

              <div v-if="usageHasExceededStock" class="flex items-center gap-1.5 text-red-600 font-bold text-xs bg-red-100/70 px-2.5 py-1 rounded-xl">
                <AlertTriangle class="w-3.5 h-3.5 shrink-0" />
                <span>Ada jumlah item yang melebihi sisa stok fisik!</span>
              </div>
            </div>
          </div>

          <DialogFooter class="gap-2 sm:gap-0 pt-2 border-t border-slate-100">
            <Button type="button" variant="outline" class="rounded-xl font-bold" @click="showUsageDialog = false">
              Batal
            </Button>
            <Button
              type="submit"
              :disabled="saving || usageHasExceededStock"
              class="rounded-xl font-bold bg-orange-600 hover:bg-orange-700 text-white shadow-md shadow-orange-600/20 disabled:opacity-50"
            >
              <Loader2 v-if="saving" class="w-4 h-4 mr-1.5 animate-spin" />
              Simpan & Potong Stok
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- 4. DIALOG MUTATION LOGS (GLOBAL & PER-ITEM) -->
    <Dialog :open="showLogsDialog" @update:open="showLogsDialog = $event">
      <DialogContent class="sm:max-w-[800px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2 text-lg font-black text-slate-900">
            <History class="w-5 h-5 text-blue-600" />
            Riwayat Mutasi Stok Barang Teknisi
          </DialogTitle>
          <DialogDescription>
            {{ selectedLogItem ? `Histori keluar/masuk & opname untuk: ${selectedLogItem.name}` : 'Semua transaksi penerimaan, pemakaian teknisi, dan stock opname pabrik.' }}
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-4 mt-2">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <select
                v-model="logTypeFilter"
                @change="fetchLogs"
                class="h-9 rounded-xl border border-slate-200 bg-white px-2.5 py-1 text-xs font-semibold text-slate-700 focus:outline-none"
              >
                <option value="">Semua Tipe Mutasi</option>
                <option value="IN">Hanya Masuk (Penerimaan / Tambah)</option>
                <option value="OUT">Hanya Keluar (Pemakaian Mesin)</option>
                <option value="OPNAME">Hanya Penyesuaian Opname</option>
              </select>
            </div>

            <Button
              size="sm"
              variant="outline"
              @click="fetchLogs"
              :disabled="loadingLogs"
              class="h-9 rounded-xl text-xs font-bold"
            >
              <RefreshCw class="w-3.5 h-3.5 mr-1" :class="{ 'animate-spin': loadingLogs }" /> Refresh Log
            </Button>
          </div>

          <div class="border border-slate-200 rounded-2xl overflow-hidden bg-white">
            <div class="overflow-x-auto max-h-[400px]">
              <table class="w-full text-left border-collapse text-xs">
                <thead class="sticky top-0 bg-slate-50 border-b border-slate-200 text-[10px] font-black uppercase text-slate-500">
                  <tr>
                    <th class="py-2.5 px-3">Tanggal</th>
                    <th class="py-2.5 px-3">Barang</th>
                    <th class="py-2.5 px-3">Tipe & Sumber</th>
                    <th class="py-2.5 px-3 text-center">Qty Mutasi</th>
                    <th class="py-2.5 px-3 text-center">Stok Awal &rarr; Akhir</th>
                    <th class="py-2.5 px-3">Keterangan / Ref</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-if="loadingLogs" class="text-center">
                    <td colspan="6" class="py-8 text-slate-400">
                      <Loader2 class="w-5 h-5 animate-spin mx-auto mb-1 text-blue-600" />
                      Memuat data log mutasi...
                    </td>
                  </tr>

                  <tr v-else-if="stockLogs.length === 0" class="text-center">
                    <td colspan="6" class="py-8 text-slate-400">
                      Belum ada catatan mutasi stok untuk filter ini.
                    </td>
                  </tr>

                  <tr v-for="log in stockLogs" :key="log.id" class="hover:bg-slate-50/70">
                    <td class="py-2.5 px-3 font-mono text-[11px] text-slate-600">
                      {{ formatDateTime(log.created_at) }}
                    </td>

                    <td class="py-2.5 px-3 font-bold text-slate-900">
                      {{ log.item?.name || selectedLogItem?.name || '-' }}
                    </td>

                    <td class="py-2.5 px-3">
                      <span
                        class="px-2 py-0.5 rounded-full text-[10px] font-black"
                        :class="getLogTypeBadgeClass(log.type)"
                      >
                        {{ log.type }} - {{ log.source }}
                      </span>
                    </td>

                    <td class="py-2.5 px-3 text-center font-mono font-bold"
                      :class="log.type === 'IN' ? 'text-emerald-700' : log.type === 'OUT' ? 'text-red-700' : 'text-amber-700'"
                    >
                      {{ log.type === 'IN' ? `+${log.quantity}` : log.type === 'OUT' ? `-${log.quantity}` : `${log.quantity > 0 ? '+' : ''}${log.quantity}` }}
                    </td>

                    <td class="py-2.5 px-3 text-center font-mono text-[11px] text-slate-600">
                      {{ log.stock_before }} &rarr; <strong class="text-slate-900">{{ log.stock_after }}</strong>
                    </td>

                    <td class="py-2.5 px-3 text-slate-600 text-[11px]">
                      <div>{{ log.notes || '-' }}</div>
                      <div v-if="log.reference_no" class="font-mono text-[10px] text-slate-400">
                        Ref: {{ log.reference_no }}
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <DialogFooter class="pt-2">
          <Button type="button" variant="outline" class="rounded-xl font-bold" @click="showLogsDialog = false">
            Tutup
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Datalist for UoM options -->
    <datalist id="tech-uom-options">
      <option v-for="u in ALL_COMMON_UNITS" :key="u" :value="u">{{ u }}</option>
    </datalist>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Wrench,
  Plus,
  RefreshCw,
  Search,
  Boxes,
  AlertTriangle,
  Layers,
  Coins,
  AlertCircle,
  MapPin,
  ClipboardCheck,
  History,
  Edit3,
  Trash2,
  Loader2,
  Hammer,
} from 'lucide-vue-next'
import { useFactoryStore } from '@/stores/factory'
import { factoryApi } from '@/api/factory'
import { toast } from 'vue-sonner'
import { ALL_COMMON_UNITS } from '@/utils/productUnits'
import TechnicianItemSearchSelect from '@/components/factory/TechnicianItemSearchSelect.vue'

const CATEGORIES = [
  { value: 'MUR_BAUT', label: '🔩 Mur & Baut' },
  { value: 'SPAREPART_MESIN', label: '⚙️ Sparepart Mesin' },
  { value: 'KELISTRIKAN', label: '⚡ Kelistrikan & Sensor' },
  { value: 'PELUMAS', label: '🛢️ Pelumas & Oli' },
  { value: 'TOOLS', label: '🔧 Tools & Perkakas' },
  { value: 'PNEUMATIK_HIDROLIK', label: '💨 Pneumatik & Hidrolik' },
  { value: 'LAINNYA', label: '📦 Lainnya' },
]

const factoryStore = useFactoryStore()
const factories = computed(() => factoryStore.factories)
const activeFactoryId = ref<string>('')

const loading = ref(false)
const saving = ref(false)
const items = ref<any[]>([])
const searchQuery = ref('')
const selectedCategory = ref('')
const filterLowStockOnly = ref(false)

// Dialog states
const showItemDialog = ref(false)
const editingItem = ref<any | null>(null)
const itemForm = ref({
  sku: '',
  name: '',
  category: 'MUR_BAUT',
  specification: '',
  unit: 'PCS',
  current_stock: 0,
  minimum_stock: 10,
  unit_price: 0,
  rack_location: '',
  notes: '',
})

// Opname Dialog states
const showOpnameDialog = ref(false)
const selectedOpnameItem = ref<any | null>(null)
const opnameForm = ref({
  physical_quantity: 0,
  reason: 'STOCK_OPNAME_REGULER',
  notes: '',
})

// Usage Dialog states
const showUsageDialog = ref(false)
const usageForm = ref({
  technician_name: '',
  machine_or_purpose: '',
  usage_date: new Date().toISOString().split('T')[0],
  notes: '',
  items: [
    {
      technician_item_id: '',
      quantity: 1,
      notes: '',
    },
  ],
})

// Logs Dialog states
const showLogsDialog = ref(false)
const selectedLogItem = ref<any | null>(null)
const loadingLogs = ref(false)
const logTypeFilter = ref('')
const stockLogs = ref<any[]>([])

// Computed Metrics
const lowStockItemsCount = computed(() => {
  return items.value.filter((it) => it.current_stock <= it.minimum_stock).length
})

const activeCategoriesCount = computed(() => {
  const cats = new Set(items.value.map((it) => it.category))
  return cats.size
})

const totalInventoryValue = computed(() => {
  return items.value.reduce((sum, it) => sum + (it.current_stock || 0) * (it.unit_price || 0), 0)
})

const filteredItems = computed(() => {
  return items.value.filter((it) => {
    if (selectedCategory.value && it.category !== selectedCategory.value) {
      return false
    }
    if (filterLowStockOnly.value && it.current_stock > it.minimum_stock) {
      return false
    }
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const matchSku = it.sku?.toLowerCase().includes(q)
      const matchName = it.name?.toLowerCase().includes(q)
      const matchSpec = it.specification?.toLowerCase().includes(q)
      const matchRack = it.rack_location?.toLowerCase().includes(q)
      return matchSku || matchName || matchSpec || matchRack
    }
    return true
  })
})

const opnameDiff = computed(() => {
  if (!selectedOpnameItem.value) return 0
  return (opnameForm.value.physical_quantity || 0) - (selectedOpnameItem.value.current_stock || 0)
})

// Methods
const onFactoryChange = () => {
  fetchData()
}

const fetchData = async () => {
  if (!activeFactoryId.value) return
  loading.value = true
  try {
    const res = await factoryApi.getTechnicianItems(activeFactoryId.value)
    items.value = (res.data?.data || []).map(normalizeTechnicianItem)
  } catch (err: any) {
    toast.error('Gagal memuat data barang teknisi: ' + (err?.response?.data?.message || err.message))
  } finally {
    loading.value = false
  }
}

// API menggunakan nama domain teknisi (code/quantity/min_limit), sedangkan
// komponen ini memakai nama yang lebih deskriptif untuk kebutuhan tampilan.
const normalizeTechnicianItem = (item: any) => ({
  ...item,
  sku: item.sku ?? item.code ?? '',
  current_stock: item.current_stock ?? item.quantity ?? 0,
  minimum_stock: item.minimum_stock ?? item.min_limit ?? 0,
  unit_price: item.unit_price ?? item.estimated_unit_price ?? 0,
  specification: item.specification ?? item.description ?? '',
})

const technicianItemPayload = (form: typeof itemForm.value, includeInitialStock: boolean) => ({
  code: form.sku.trim(),
  name: form.name.trim(),
  category: form.category,
  unit: form.unit.trim().toUpperCase(),
  ...(includeInitialStock ? { initial_quantity: Number(form.current_stock) || 0 } : {}),
  min_limit: Number(form.minimum_stock) || 0,
  rack_location: form.rack_location.trim(),
  estimated_unit_price: Number(form.unit_price) || 0,
  description: form.specification.trim() || form.notes.trim(),
})

const getCategoryLabel = (cat: string) => {
  const found = CATEGORIES.find((c) => c.value === cat)
  return found ? found.label : cat
}

const getCategoryBadgeClass = (cat: string) => {
  switch (cat) {
    case 'MUR_BAUT':
      return 'bg-amber-100 text-amber-800'
    case 'SPAREPART_MESIN':
      return 'bg-blue-100 text-blue-800'
    case 'KELISTRIKAN':
      return 'bg-yellow-100 text-yellow-800'
    case 'PELUMAS':
      return 'bg-purple-100 text-purple-800'
    case 'TOOLS':
      return 'bg-emerald-100 text-emerald-800'
    case 'PNEUMATIK_HIDROLIK':
      return 'bg-cyan-100 text-cyan-800'
    default:
      return 'bg-slate-100 text-slate-800'
  }
}

const getStockStatusClass = (current: number, min: number) => {
  if (current <= 0) return 'bg-red-100 text-red-800 border border-red-200'
  if (current <= min) return 'bg-amber-100 text-amber-800 border border-amber-200'
  return 'bg-emerald-50 text-emerald-800 border border-emerald-200'
}

const getLogTypeBadgeClass = (type: string) => {
  switch (type) {
    case 'IN':
      return 'bg-emerald-100 text-emerald-800'
    case 'OUT':
      return 'bg-red-100 text-red-800'
    case 'OPNAME':
      return 'bg-amber-100 text-amber-800'
    default:
      return 'bg-slate-100 text-slate-800'
  }
}

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 1. Create / Edit Dialog
const openCreateModal = () => {
  editingItem.value = null
  itemForm.value = {
    sku: '',
    name: '',
    category: 'MUR_BAUT',
    specification: '',
    unit: 'PCS',
    current_stock: 0,
    minimum_stock: 10,
    unit_price: 0,
    rack_location: '',
    notes: '',
  }
  showItemDialog.value = true
}

const openEditModal = (item: any) => {
  editingItem.value = item
  itemForm.value = {
    sku: item.sku,
    name: item.name,
    category: item.category,
    specification: item.specification || '',
    unit: item.unit || 'PCS',
    current_stock: item.current_stock || 0,
    minimum_stock: item.minimum_stock || 10,
    unit_price: item.unit_price || 0,
    rack_location: item.rack_location || '',
    notes: item.notes || '',
  }
  showItemDialog.value = true
}

const saveItem = async () => {
  if (!activeFactoryId.value) return
  saving.value = true
  try {
    if (editingItem.value) {
      await factoryApi.updateTechnicianItem(
        activeFactoryId.value,
        editingItem.value.id,
        technicianItemPayload(itemForm.value, false),
      )
      toast.success('Data barang teknisi berhasil diperbarui!')
    } else {
      await factoryApi.createTechnicianItem(
        activeFactoryId.value,
        technicianItemPayload(itemForm.value, true),
      )
      toast.success('Barang teknisi baru berhasil ditambahkan!')
    }
    showItemDialog.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menyimpan data: ' + (err?.response?.data?.message || err.message))
  } finally {
    saving.value = false
  }
}

const confirmDelete = async (item: any) => {
  if (!confirm(`Apakah Anda yakin ingin menghapus barang "${item.name}"? Data mutasi terkait juga akan terhapus.`)) {
    return
  }
  try {
    await factoryApi.deleteTechnicianItem(activeFactoryId.value, item.id)
    toast.success('Barang teknisi berhasil dihapus')
    fetchData()
  } catch (err: any) {
    toast.error('Gagal menghapus: ' + (err?.response?.data?.message || err.message))
  }
}

// 2. Opname Dialog
const openOpnameModal = (item: any) => {
  selectedOpnameItem.value = item
  opnameForm.value = {
    physical_quantity: item.current_stock,
    reason: 'STOCK_OPNAME_REGULER',
    notes: '',
  }
  showOpnameDialog.value = true
}

const submitOpname = async () => {
  if (!activeFactoryId.value || !selectedOpnameItem.value) return
  saving.value = true
  try {
    await factoryApi.opnameTechnicianItem(activeFactoryId.value, selectedOpnameItem.value.id, opnameForm.value)
    toast.success('Stock opname berhasil disimpan dan mutasi tercatat!')
    showOpnameDialog.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal opname: ' + (err?.response?.data?.message || err.message))
  } finally {
    saving.value = false
  }
}

// 3. Usage Dialog
const getTechnicianItemById = (id: string) => {
  if (!id) return null
  return items.value.find((it) => it.id === id) || null
}

const openUsageModal = () => {
  usageForm.value = {
    technician_name: '',
    machine_or_purpose: '',
    usage_date: new Date().toISOString().split('T')[0],
    notes: '',
    items: [
      {
        technician_item_id: '',
        quantity: 1,
        notes: '',
      },
    ],
  }
  showUsageDialog.value = true
}

const addUsageRow = () => {
  usageForm.value.items.push({
    technician_item_id: '',
    quantity: 1,
    notes: '',
  })
}

const removeUsageRow = (idx: number) => {
  usageForm.value.items.splice(idx, 1)
}

const usageTotalQty = computed(() => {
  return usageForm.value.items.reduce((sum, it) => sum + (Number(it.quantity) || 0), 0)
})

const usageSelectedItemsCount = computed(() => {
  return usageForm.value.items.filter((it) => it.technician_item_id).length
})

const usageHasExceededStock = computed(() => {
  return usageForm.value.items.some((row) => {
    if (!row.technician_item_id) return false
    const item = getTechnicianItemById(row.technician_item_id)
    if (!item) return false
    return Number(row.quantity) > (item.current_stock || 0)
  })
})

const submitUsage = async () => {
  if (!activeFactoryId.value) return
  const validItems = usageForm.value.items.filter((it) => it.technician_item_id && it.quantity > 0)
  if (validItems.length === 0) {
    toast.error('Pilih minimal satu barang teknisi yang dipakai dengan jumlah > 0')
    return
  }

  // Cek validasi stok sebelum submit ke backend
  for (const row of validItems) {
    const item = getTechnicianItemById(row.technician_item_id)
    if (item && row.quantity > item.current_stock) {
      toast.error(`Stok "${item.name}" tidak mencukupi! Tersedia: ${item.current_stock} ${item.unit}, diminta: ${row.quantity} ${item.unit}`)
      return
    }
  }

  saving.value = true
  try {
    await factoryApi.recordTechnicianUsage(activeFactoryId.value, {
      ...usageForm.value,
      items: validItems,
    })
    toast.success('Pemakaian barang teknisi berhasil dicatat dan stok dipotong!')
    showUsageDialog.value = false
    fetchData()
  } catch (err: any) {
    toast.error('Gagal mencatat pemakaian: ' + (err?.response?.data?.message || err.message))
  } finally {
    saving.value = false
  }
}

// 4. Logs Dialog
const openLogsModal = (item?: any) => {
  selectedLogItem.value = item || null
  logTypeFilter.value = ''
  showLogsDialog.value = true
  fetchLogs()
}

const fetchLogs = async () => {
  if (!activeFactoryId.value) return
  loadingLogs.value = true
  try {
    const params: any = {}
    if (selectedLogItem.value) {
      params.item_id = selectedLogItem.value.id
    }
    if (logTypeFilter.value) {
      params.type = logTypeFilter.value
    }
    const res = await factoryApi.getTechnicianStockLogs(activeFactoryId.value, params)
    stockLogs.value = res.data?.data || []
  } catch (err: any) {
    toast.error('Gagal memuat log mutasi: ' + (err?.response?.data?.message || err.message))
  } finally {
    loadingLogs.value = false
  }
}

onMounted(async () => {
  await factoryStore.fetchFactories()
  if (factories.value.length > 0 && factories.value[0]) {
    activeFactoryId.value = factories.value[0].id
    fetchData()
  }
})
</script>
