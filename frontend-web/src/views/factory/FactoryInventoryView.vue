<template>
  <div class="p-6 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header with Breadcrumbs -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-6 rounded-3xl shadow-sm border border-slate-100"
    >
      <div class="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          @click="router.back()"
          class="rounded-full hover:bg-slate-100"
        >
          <ArrowLeft class="h-5 w-5" />
        </Button>
        <div>
          <div
            class="flex items-center gap-2 text-xs font-bold text-primary uppercase tracking-widest mb-1"
          >
            <Factory class="h-3 w-3" />
            <span>Pusat Produksi & Inventory</span>
          </div>
          <h1 class="text-2xl font-extrabold tracking-tight text-slate-900">
            {{ factoryStore.currentFactory?.name || 'Memuat...' }}
          </h1>
          <p class="text-sm text-slate-500 flex items-center mt-0.5">
            <MapPin class="mr-1 h-3.5 w-3.5 text-slate-400" />
            {{ factoryStore.currentFactory?.location }}
          </p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <Button
          variant="outline"
          @click="handleSyncWarehouseStock"
          class="rounded-xl border-slate-200"
          :disabled="syncingWarehouse"
        >
          <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': syncingWarehouse }" />
          Sinkronkan ke Gudang
        </Button>
        <Button
          variant="outline"
          @click="router.push('/factory/qc-samples')"
          class="rounded-xl border-indigo-200 text-indigo-700 bg-indigo-50/50 hover:bg-indigo-100/50 font-bold"
        >
          <FlaskConical class="mr-2 h-4 w-4 text-indigo-600" /> Sampel Lab & QC
        </Button>
        <Button
          variant="outline"
          @click="showInventoryLogs = true"
          class="rounded-xl border-slate-200"
        >
          <History class="mr-2 h-4 w-4" /> Log Mutasi
        </Button>
        <Button
          @click="showSendToBranch = true"
          class="rounded-xl bg-orange-600 hover:bg-orange-700 shadow-lg shadow-orange-200 transition-all"
        >
          <Truck class="mr-2 h-4 w-4" /> Kirim ke Cabang
        </Button>
      </div>
    </div>

    <!-- Main Split Layout -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <!-- LEFT COLUMN: STOCK MANAGEMENT -->
      <div class="space-y-6">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 rounded-2xl bg-blue-50 flex items-center justify-center text-blue-600"
            >
              <Package class="h-5 w-5" />
            </div>
            <h2 class="text-xl font-bold text-slate-900">Manajemen Stok</h2>
          </div>
          <Badge variant="secondary" class="bg-blue-50 text-blue-600 border-none font-bold">
            {{ factoryStore.inventory.length }} Produk
          </Badge>
        </div>

        <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
          <DataTable :columns="stockColumns" :data="factoryStore.inventory" class="border-none">
            <template #cell-warehouse_location="{ row }">
              <span class="text-[11px] font-bold text-slate-700 bg-slate-100 px-2 py-0.5 rounded border border-slate-200/50">
                Gudang {{ getWarehouseNumber((row as any).product?.name, (row as any).product?.category).number }}
              </span>
            </template>
            <template #cell-quantity="{ row }">
              <div class="flex items-center gap-3">
                <span class="text-lg font-black text-slate-900">{{ (row as any).quantity }}</span>
                <span class="text-[10px] font-bold text-slate-400 uppercase tracking-tighter">{{
                  (row as any).product?.unit
                }}</span>
              </div>
            </template>
            <template #cell-actions="{ row }">
              <Button
                variant="ghost"
                size="sm"
                @click="openAdjustStock(row)"
                class="h-8 w-8 p-0 rounded-lg text-slate-400 hover:text-primary hover:bg-primary/5"
              >
                <Settings2 class="h-4 w-4" />
              </Button>
            </template>
          </DataTable>
        </Card>
      </div>

      <!-- RIGHT COLUMN: PRODUCTION LOG -->
      <div class="space-y-6">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 rounded-2xl bg-emerald-50 flex items-center justify-center text-emerald-600"
            >
              <Activity class="h-5 w-5" />
            </div>
            <h2 class="text-xl font-bold text-slate-900">Log Produksi</h2>
          </div>
          <Button
            size="sm"
            @click="openLogProduction"
            class="rounded-xl shadow-md shadow-primary/10"
          >
            <Plus class="mr-1 h-4 w-4" /> Log Baru
          </Button>
        </div>

        <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
          <DataTable
            :columns="productionColumns"
            :data="factoryStore.productionHistory"
            class="border-none"
          >
            <template #cell-warehouse_location="{ row }">
              <span class="text-[11px] font-bold text-slate-700 bg-slate-100 px-2 py-0.5 rounded border border-slate-200/50">
                Gudang {{ getWarehouseNumber((row as any).product?.name, (row as any).product?.category).number }}
              </span>
            </template>
            <template #cell-quantity="{ row }">
              <span class="font-bold text-emerald-600">+{{ (row as any).quantity }}</span>
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
              <div class="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  @click="openTransferHistory(row)"
                  title="Lihat Histori Pengiriman Cabang"
                  class="h-8 w-8 text-slate-400 hover:text-indigo-500"
                >
                  <Truck class="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  @click="openEditProduction(row)"
                  class="h-8 w-8 text-slate-400 hover:text-primary"
                >
                  <Edit2 class="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  @click="handleDeleteProduction((row as any).id)"
                  class="h-8 w-8 text-slate-400 hover:text-destructive"
                >
                  <Trash2 class="h-3.5 w-3.5" />
                </Button>
              </div>
            </template>
          </DataTable>
        </Card>
      </div>
    </div>

    <!-- Modals -->
    <!-- Log Production Modal -->
    <div
      v-if="showLogProduction"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-md shadow-2xl animate-in zoom-in-95 duration-200 rounded-3xl overflow-hidden border-none"
      >
        <CardHeader class="p-6 border-b bg-white">
          <CardTitle class="text-xl font-bold">{{
            isEditingProduction ? 'Edit Log Produksi' : 'Catat Hasil Produksi'
          }}</CardTitle>
          <CardDescription
            >Barang yang dicatat akan otomatis menambah stok inventory.</CardDescription
          >
        </CardHeader>
        <CardContent class="p-6 space-y-4 bg-white">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Pilih Produk</label>
            <SearchSelect
              v-model="productionForm.product_id"
              :options="finishedGoodsOptions"
              placeholder="Pilih Produk..."
              search-placeholder="Ketik nama / SKU produk..."
              empty-message="Produk tidak ditemukan"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Pilih Gudang Target (Opsional)</label>
            <select
              v-model="productionForm.target_warehouse_id"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:ring-2 focus:ring-primary/20 outline-none"
            >
              <option value="">Tidak Ditentukan (Gudang Umum)</option>
              <option v-for="w in factoryStore.factoryWarehouses" :key="w.id" :value="w.id">
                {{ w.name }} <span v-if="w.code">({{ w.code }})</span>
              </option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700">Nomor Batch</label>
              <input
                type="text"
                v-model="productionForm.batch_no"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="Contoh: B1-2024"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700">Tgl Kadaluarsa</label>
              <input
                type="date"
                v-model="productionForm.expiry_date"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20"
              />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700">Jumlah Hasil</label>
              <input
                type="number"
                v-model="productionForm.quantity"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700">ID Operator</label>
              <input
                type="text"
                v-model="productionForm.employee_id"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20"
                placeholder="UUID"
              />
            </div>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Catatan</label>
            <textarea
              v-model="productionForm.notes"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20 resize-none"
              rows="2"
            ></textarea>
          </div>

          <!-- Direct Raw Material Consumption Section -->
          <div v-if="canViewMaterials" class="p-4 bg-amber-50/60 rounded-2xl border border-amber-200/70 space-y-3">
            <div class="flex justify-between items-center">
              <div class="flex items-center gap-2">
                <Component class="w-4 h-4 text-amber-700" />
                <span class="text-xs font-black text-amber-900">Bahan Baku / Wadah Terpakai {{ isEditingProduction ? '(Dapat Diedit)' : '(Opsional)' }}</span>
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                @click="addRawMaterial"
                class="h-7 text-xs rounded-xl border-amber-300 text-amber-900 bg-amber-100/50 hover:bg-amber-100 font-bold"
              >
                <Plus class="w-3 h-3 mr-1" /> Tambah Bahan
              </Button>
            </div>

            <div v-if="productionForm.raw_materials.length === 0" class="text-[11px] text-amber-700/80 italic">
              Belum ada bahan baku yang dipilih. Stok bahan baku di pabrik tidak akan dipotong jika kosong.
            </div>

            <div v-else class="space-y-2">
              <div
                v-for="(rm, rmIdx) in productionForm.raw_materials"
                :key="rmIdx"
                class="space-y-1 bg-white p-2.5 rounded-xl border border-amber-100 shadow-sm"
              >
                <div class="grid grid-cols-1 sm:grid-cols-12 gap-2 items-center">
                  <div class="sm:col-span-4">
                    <select
                      v-model="rm.product_id"
                      @change="onRawMaterialChange(rm)"
                      class="w-full text-xs font-bold bg-slate-50 border border-slate-200 rounded-lg p-2 outline-none"
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
                    <select
                      v-model="rm.batch_no"
                      class="w-full text-xs font-bold bg-slate-50 border border-slate-200 rounded-lg p-2 outline-none"
                    >
                      <option value="">Auto (FEFO / Exp Terdekat)</option>
                      <option
                        v-for="b in getAvailableBatchesForRawMaterial(factoryId, rm.product_id, rm.batch_no)"
                        :key="b.batch_no"
                        :value="b.batch_no"
                      >
                        {{ b.batch_no }} (Sisa: {{ b.quantity }} | {{ b.expiry_date ? b.expiry_date.substring(0,10) : 'No Exp' }})
                      </option>
                    </select>
                  </div>
                  <div class="sm:col-span-2">
                    <input
                      type="number"
                      v-model.number="rm.quantity"
                      min="0.001"
                      step="any"
                      placeholder="Jumlah"
                      class="w-full text-xs font-bold font-mono bg-slate-50 border border-slate-200 rounded-lg p-2 outline-none"
                    />
                  </div>
                  <div class="sm:col-span-2">
                    <select
                      v-model="rm.unit"
                      class="w-full text-xs font-bold uppercase bg-slate-50 border border-slate-200 rounded-lg p-2 outline-none"
                    >
                      <option v-for="u in getRawMaterialCompatibleUnits(rm.product_id)" :key="u" :value="u">{{ u }}</option>
                    </select>
                  </div>
                  <div class="sm:col-span-1 flex justify-end">
                    <button
                      type="button"
                      @click="removeRawMaterial(rmIdx)"
                      class="text-red-500 hover:text-red-700 text-xs font-bold p-1 rounded-lg hover:bg-red-50"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>

                <!-- Live Conversion Hint -->
                <div
                  v-if="getRawMaterialConversionHint(rm)"
                  class="text-[10px] text-amber-700 font-medium px-1 flex items-center gap-1"
                >
                  <Sparkles class="w-2.5 h-2.5 text-amber-600" />
                  <span>{{ getRawMaterialConversionHint(rm) }}</span>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 p-6 flex justify-end gap-3">
          <Button variant="ghost" @click="showLogProduction = false" class="rounded-xl font-bold"
            >Batal</Button
          >
          <Button
            @click="handleProductionSubmit"
            :loading="submitting"
            :disabled="submitting"
            class="rounded-xl px-8 shadow-lg shadow-primary/20 font-bold"
          >
            {{ isEditingProduction ? 'Update Log' : 'Simpan & Update Stok' }}
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

    <!-- Adjust Stock Modal -->
    <div
      v-if="showAdjustStock"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card class="w-full max-w-md shadow-2xl rounded-3xl overflow-hidden border-none">
        <CardHeader class="p-6 border-b bg-white">
          <CardTitle class="text-xl font-bold">Penyesuaian Stok</CardTitle>
          <CardDescription>Koreksi jumlah stok secara manual untuk produk ini.</CardDescription>
        </CardHeader>
        <CardContent class="p-6 space-y-4 bg-white">
          <div class="p-4 bg-blue-50 rounded-2xl border border-blue-100 mb-4">
            <p class="text-xs font-bold text-blue-600 uppercase tracking-widest mb-1">Produk</p>
            <p class="text-sm font-bold text-slate-900">{{ selectedStock?.product?.name }}</p>
            <p class="text-[10px] text-slate-500 mt-0.5">
              SKU: {{ selectedStock?.product?.sku }} | Stok Saat Ini: {{ selectedStock?.quantity }}
            </p>
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
            <label class="text-sm font-bold text-slate-700">Alasan Penyesuaian</label>
            <textarea
              v-model="adjustStockForm.reason"
              placeholder="Contoh: Barang rusak, Salah input, Audit stok..."
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none focus:ring-2 focus:ring-primary/20 resize-none"
              rows="2"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 p-6 flex justify-end gap-3">
          <Button variant="ghost" @click="showAdjustStock = false" class="rounded-xl font-bold"
            >Batal</Button
          >
          <Button
            @click="handleAdjustStockSubmit"
            :loading="submitting"
            class="rounded-xl px-8 shadow-lg shadow-primary/20 font-bold bg-blue-600 hover:bg-blue-700"
          >
            Koreksi Stok
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Send to Branch Modal (Enhanced Multi-Item) -->
    <div
      v-if="showSendToBranch"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-2xl shadow-2xl rounded-3xl overflow-hidden border-none animate-in fade-in zoom-in-95 duration-200"
      >
        <CardHeader class="p-6 border-b bg-white">
          <div class="flex justify-between items-start">
            <div>
              <CardTitle class="text-xl font-bold">Kirim ke Cabang (Multi-Item)</CardTitle>
              <CardDescription
                >Pilih beberapa barang untuk dikirim dalam satu armada.</CardDescription
              >
            </div>
            <div class="text-right">
              <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">
                Total Tonase
              </p>
              <p class="text-2xl font-black text-primary">
                {{ totalShipmentWeight.toFixed(2) }} <span class="text-xs">KG</span>
              </p>
            </div>
          </div>
        </CardHeader>
        <CardContent class="p-6 space-y-6 bg-white max-h-[60vh] overflow-y-auto">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Cabang Tujuan</label>
            <select
              v-model="transferForm.to_branch_id"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none"
            >
              <option value="" disabled>Pilih Cabang...</option>
              <option v-for="b in masterDataStore.branches" :key="b.id" :value="b.id">
                {{ b.name }}
              </option>
            </select>
          </div>

          <!-- Multi Item List -->
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <label class="text-sm font-bold text-slate-700">Item Pengiriman</label>
              <Button
                variant="ghost"
                size="sm"
                @click="addTransferItem"
                class="text-primary font-bold text-xs"
              >
                <Plus class="mr-1 h-3 w-3" /> Tambah Barang
              </Button>
            </div>

            <div
              v-for="(item, index) in transferForm.items"
              :key="index"
              class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-4"
            >
              <div class="flex justify-between items-start gap-4">
                <div class="flex-1 space-y-2">
                  <label class="text-[10px] font-bold text-slate-400 uppercase"
                    >Pilih Produk & Batch</label
                  >
                  <select
                    v-model="item._identifier"
                    @change="onTransferItemChange(item)"
                    class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm outline-none"
                  >
                    <option
                      v-for="s in factoryStore.inventory"
                      :key="s.id"
                      :value="`${s.product_id}|${s.batch_no}`"
                    >
                      {{ s.product?.name }} - Batch: {{ s.batch_no }} (Stok: {{ s.quantity }})
                    </option>
                  </select>
                </div>
                <div class="w-24 space-y-2">
                  <label class="text-[10px] font-bold text-slate-400 uppercase">Jumlah</label>
                  <input
                    type="number"
                    v-model="item.quantity"
                    class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-sm outline-none"
                  />
                </div>
                <div class="w-24 space-y-2 text-right">
                  <label class="text-[10px] font-bold text-slate-400 uppercase">Tonase</label>
                  <p class="py-2 font-bold text-slate-700 text-sm">
                    {{ calculateItemWeight(item).toFixed(1) }} KG
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  @click="removeTransferItem(index)"
                  class="mt-6 h-8 w-8 text-slate-300 hover:text-destructive"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700">Metode Pembayaran</label>
              <select
                v-model="transferForm.payment_method"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none"
              >
                <option value="NET">NET (Lunas)</option>
                <option value="TEMPO">TEMPO (Kredit)</option>
              </select>
            </div>
            <div class="space-y-2" v-if="transferForm.payment_method === 'TEMPO'">
              <label class="text-sm font-bold text-slate-700">Jatuh Tempo Pembayaran</label>
              <input
                type="date"
                v-model="transferForm.payment_due_at"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Catatan Pengiriman (Opsional)</label>
            <textarea
              v-model="transferForm.notes"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none resize-none"
              rows="2"
              placeholder="Contoh: Titip armada vendor, dll"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 p-6 flex justify-end gap-3">
          <Button variant="ghost" @click="showSendToBranch = false" class="rounded-xl font-bold"
            >Batal</Button
          >
          <Button
            @click="handleTransferSubmit"
            :loading="submitting"
            class="rounded-xl px-8 shadow-lg shadow-orange-200 font-bold bg-orange-600 hover:bg-orange-700 text-white"
          >
            Konfirmasi & Buat Surat Jalan
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Log Mutasi Modal -->
    <div
      v-if="showInventoryLogs"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50 flex items-center justify-center p-4"
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
        </CardHeader>
        <CardContent class="p-0 bg-white max-h-[60vh] overflow-y-auto">
          <div
            v-if="factoryStore.inventoryLogs.length === 0"
            class="p-12 text-center text-slate-400 font-medium"
          >
            Belum ada log mutasi.
          </div>
          <table v-else class="w-full text-sm text-left">
            <thead
              class="bg-slate-50 sticky top-0 border-b border-slate-100 text-[10px] font-black text-slate-400 uppercase tracking-widest"
            >
              <tr>
                <th class="px-6 py-4">Waktu</th>
                <th class="px-6 py-4">Produk</th>
                <th class="px-6 py-4">Sumber</th>
                <th class="px-6 py-4">Tipe</th>
                <th class="px-6 py-4 text-right">Jumlah</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="log in factoryStore.inventoryLogs"
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
                  <div class="font-bold text-slate-900">{{ log.product?.name || 'Unknown' }}</div>
                  <div class="text-[10px] text-slate-500">Batch: {{ log.batch_no || '-' }}</div>
                </td>
                <td class="px-6 py-4">
                  <Badge
                    variant="outline"
                    :class="
                      log.source?.startsWith('ADJUSTMENT')
                        ? 'bg-orange-50 text-orange-600 border-orange-200'
                        : log.source?.startsWith('PRODUCTION')
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

    <!-- Production Raw Materials & Used Batches Detail Modal -->
    <ProductionMaterialsDialog
      v-model:open="showMaterialsModal"
      :production-log="selectedLogForMaterials"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
import {
  ArrowLeft,
  Package,
  Activity,
  Plus,
  Edit2,
  Trash2,
  Settings2,
  Truck,
  History,
  MapPin,
  Factory,
  ChevronRight,
  User,
  X,
  RefreshCw,
  FlaskConical,
} from 'lucide-vue-next'

import { factoryApi } from '@/api/factory'

const route = useRoute()
const router = useRouter()
const factoryStore = useFactoryStore()
const masterDataStore = useMasterDataStore()
const factoryId = route.params.id as string

const showLogProduction = ref(false)
const showAdjustStock = ref(false)
const showSendToBranch = ref(false)
const showInventoryLogs = ref(false)
const submitting = ref(false)
const isEditingProduction = ref(false)
const currentProductionId = ref('')
const showTransferHistory = ref(false)
const transferHistoryLoading = ref(false)
const selectedProductionLog = ref<any>(null)
const transferHistoryData = ref<any[]>([])
const selectedStock = ref<any>(null)

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

const productionForm = reactive({
  product_id: '',
  target_warehouse_id: '',
  quantity: 0,
  batch_no: '',
  expiry_date: '',
  employee_id: '',
  notes: '',
  raw_materials: [] as any[],
})

const addRawMaterial = () => {
  productionForm.raw_materials.push({
    product_id: '',
    batch_no: '',
    quantity: 1,
    unit: 'KG',
    notes: '',
  })
}

const removeRawMaterial = (rmIdx: number) => {
  productionForm.raw_materials.splice(rmIdx, 1)
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
  const batches = (factoryStore.inventory || []).filter(
    (s) =>
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

const adjustStockForm = reactive({
  quantity: 0,
  reason: '',
})

const transferForm = reactive({
  to_branch_id: '',
  items: [{ _identifier: '', product_id: '', batch_no: '', quantity: 0 }] as any[],
  notes: '',
  payment_method: 'NET',
  payment_due_at: '',
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

watch(
  () => productionForm.product_id,
  (newVal: string, oldVal?: string) => {
    if (newVal && (!isEditingProduction.value || (oldVal && newVal !== oldVal))) {
      productionForm.batch_no = generateBatchNumber(newVal)
    }
  },
)

const addTransferItem = () => {
  transferForm.items.push({ _identifier: '', product_id: '', batch_no: '', quantity: 0 })
}

const onTransferItemChange = (item: any) => {
  if (item._identifier) {
    const parts = item._identifier.split('|')
    item.product_id = parts[0]
    item.batch_no = parts[1]
  }
}

const removeTransferItem = (index: number) => {
  if (transferForm.items.length > 1) {
    transferForm.items.splice(index, 1)
  }
}

const calculateItemWeight = (item: any) => {
  const product = factoryStore.products.find((p) => p.id === item.product_id)
  return product ? product.weight * item.quantity : 0
}

const totalShipmentWeight = computed(() => {
  return transferForm.items.reduce((total, item) => total + calculateItemWeight(item), 0)
})

onMounted(async () => {
  await factoryStore.fetchFactoryDetail(factoryId)
  factoryStore.fetchProducts()
  factoryStore.fetchInventory(factoryId)
  factoryStore.fetchProductionHistory(factoryId)
  factoryStore.fetchInventoryLogs(factoryId)
  factoryStore.fetchFactoryWarehouses(factoryId)
  masterDataStore.fetchBranches()
})

const openLogProduction = () => {
  isEditingProduction.value = false
  Object.assign(productionForm, {
    product_id: '',
    target_warehouse_id: '',
    quantity: 0,
    batch_no: '',
    expiry_date: '',
    employee_id: '',
    notes: '',
    raw_materials: [],
  })
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
  Object.assign(productionForm, {
    product_id: item.product_id,
    target_warehouse_id: item.target_warehouse_id || '',
    quantity: item.quantity,
    batch_no: item.batch_no,
    expiry_date: item.expiry_date ? new Date(item.expiry_date).toISOString().split('T')[0] : '',
    employee_id: item.employee_id,
    notes: item.notes,
    raw_materials: (item.materials || []).map((m: any) => ({
      product_id: m.product_id,
      batch_no: m.batch_no || '',
      quantity: m.quantity,
      unit: m.product?.unit || 'KG',
      notes: m.notes || '',
    })),
  })
  showLogProduction.value = true
}

const computeStockImpactDiff = () => {
  const prod = factoryStore.products.find((p) => p.id === productionForm.product_id)
  const oldQty = Number(editingSnapshot.value?.quantity) || 0
  const newQty = Number(productionForm.quantity) || 0
  const fgDelta = newQty - oldQty

  const fgImpact = {
    name: prod?.name || editingSnapshot.value?.product?.name || 'Produk Jadi',
    batch_no: productionForm.batch_no || editingSnapshot.value?.batch_no || '-',
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
  if (!productionForm.product_id || Number(productionForm.quantity) <= 0) {
    toast.error('Lengkapi produk dan jumlah produksi!')
    return
  }

  if (isEditingProduction.value) {
    computeStockImpactDiff()
    showStockImpactModal.value = true
    return
  }

  submitting.value = true
  try {
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
      ...productionForm,
      quantity: Number(productionForm.quantity),
      expiry_date: productionForm.expiry_date || null,
      target_warehouse_id: productionForm.target_warehouse_id || undefined,
      raw_materials: rawMaterialsPayload,
    }

    await factoryStore.logProduction(factoryId, payload)
    toast.success('Log produksi berhasil dicatat')
    showLogProduction.value = false
    await factoryStore.fetchInventory(factoryId)
    await factoryStore.fetchProductionHistory(factoryId)
    await factoryStore.fetchInventoryLogs(factoryId)
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
      ...productionForm,
      quantity: Number(productionForm.quantity),
      expiry_date: productionForm.expiry_date || null,
      target_warehouse_id: productionForm.target_warehouse_id || undefined,
      raw_materials: rawMaterialsPayload,
    }

    await factoryStore.updateProductionLog(factoryId, currentProductionId.value, payload)
    toast.success('Log produksi & penyesuaian stok berhasil diperbarui')
    showStockImpactModal.value = false
    showLogProduction.value = false
    await factoryStore.fetchInventory(factoryId)
    await factoryStore.fetchProductionHistory(factoryId)
    await factoryStore.fetchInventoryLogs(factoryId)
  } catch (err) {
    toast.error('Gagal memperbarui log produksi: ' + getErrorMessage(err))
  } finally {
    submitting.value = false
  }
}

const handleDeleteProduction = async (id: string) => {
  if (!confirm('Hapus log produksi ini? Stok barang akan otomatis dikurangi kembali.')) return
  try {
    await factoryStore.deleteProductionLog(factoryId, id)
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

const openAdjustStock = (row: any) => {
  selectedStock.value = row
  adjustStockForm.quantity = row.quantity
  adjustStockForm.reason = ''
  showAdjustStock.value = true
}

const syncingWarehouse = ref(false)
const handleSyncWarehouseStock = async () => {
  syncingWarehouse.value = true
  try {
    await factoryStore.syncWarehouseStock(factoryId)
    toast.success('Stok pabrik dan stok gudang berhasil disinkronkan 1:1')
  } catch (err) {
    toast.error('Gagal sinkronisasi stok: ' + getErrorMessage(err))
  } finally {
    syncingWarehouse.value = false
  }
}

const handleAdjustStockSubmit = async () => {
  submitting.value = true
  try {
    await factoryStore.adjustStock(factoryId, {
      product_id: selectedStock.value.product_id,
      batch_no: selectedStock.value.batch_no,
      quantity: adjustStockForm.quantity,
      reason: adjustStockForm.reason,
    })
    showAdjustStock.value = false
    toast.success('Stok pabrik dan stok gudang berhasil disesuaikan 1:1')
  } catch (err) {
    toast.error('Gagal menyesuaikan stok: ' + getErrorMessage(err))
  } finally {
    submitting.value = false
  }
}

const handleTransferSubmit = async () => {
  if (!transferForm.to_branch_id || transferForm.items.length === 0) return

  // Validate all items have product selected
  for (const item of transferForm.items) {
    if (!item.product_id) {
      toast.error('Mohon pilih produk untuk semua baris barang.')
      return
    }
  }

  submitting.value = true
  try {
    await factoryStore.sendToBranch(factoryId, transferForm)
    showSendToBranch.value = false
    toast.success('Permintaan pengiriman berhasil dibuat')
    // Reset form
    Object.assign(transferForm, {
      to_branch_id: '',
      items: [{ _identifier: '', product_id: '', batch_no: '', quantity: 0 }],
      notes: '',
    })
  } catch (err) {
    toast.error('Gagal membuat pengiriman: ' + getErrorMessage(err))
  } finally {
    submitting.value = false
  }
}

const getWarehouseNumber = (productName: string, category: string) => {
  const name = (productName || '').toLowerCase()
  const cat = (category || '').toLowerCase()
  
  if (name.includes('kecap') || cat.includes('kecap')) {
    return { number: 1, type: 'Kecap' }
  }
  if (name.includes('saus') || cat.includes('saus')) {
    return { number: 2, type: 'Saus' }
  }
  if (name.includes('minyak') || cat.includes('minyak')) {
    return { number: 3, type: 'Minyak Goreng' }
  }
  if (name.includes('sambal') || cat.includes('sambal')) {
    return { number: 4, type: 'Sambal' }
  }
  if (name.includes('bumbu') || cat.includes('bumbu')) {
    return { number: 5, type: 'Bumbu Dapur' }
  }
  if (name.includes('mie') || name.includes('noodle') || cat.includes('mie') || cat.includes('noodle')) {
    return { number: 6, type: 'Mie & Noodle' }
  }
  if (name.includes('tepung') || cat.includes('tepung')) {
    return { number: 7, type: 'Tepung' }
  }
  if (name.includes('gula') || cat.includes('gula')) {
    return { number: 8, type: 'Gula' }
  }
  if (name.includes('beras') || cat.includes('beras')) {
    return { number: 9, type: 'Beras' }
  }
  if (name.includes('garam') || cat.includes('garam')) {
    return { number: 10, type: 'Garam' }
  }
  if (name.includes('penyedap') || cat.includes('penyedap')) {
    return { number: 11, type: 'Penyedap Rasa' }
  }
  return { number: 12, type: 'Umum / Lain-lain' }
}

const stockColumns = [
  { key: 'product.sku', label: 'SKU' },
  { key: 'product.name', label: 'Produk' },
  { key: 'batch_no', label: 'Batch' },
  { key: 'warehouse_location', label: 'Lokasi Gudang' },
  { key: 'expiry_date', label: 'Kadaluarsa', type: 'date' },
  { key: 'quantity', label: 'Stok Gudang' },
  { key: 'actions', label: '' },
]

const productionColumns = computed(() => {
  const cols = [
    { key: 'production_date', label: 'Tanggal', type: 'datetime' },
    { key: 'product.name', label: 'Produk' },
    { key: 'batch_no', label: 'Batch' },
    { key: 'warehouse_location', label: 'Lokasi Gudang' },
    { key: 'quantity', label: 'Qty' },
  ]
  if (canViewMaterials.value) {
    cols.push({ key: 'materials', label: 'Bahan Baku & Batch Terpakai' })
  }
  cols.push({ key: 'actions', label: '' })
  return cols
})
</script>
