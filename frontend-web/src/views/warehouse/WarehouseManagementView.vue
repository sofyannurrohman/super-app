<template>
  <div class="p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <div class="no-print space-y-8">
      <!-- Header -->
      <div
        class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-8 rounded-[2.5rem] shadow-sm border border-slate-100"
      >
        <div class="flex items-center gap-5">
          <div
            class="w-16 h-16 rounded-[2rem] bg-slate-900 flex items-center justify-center text-white shadow-xl shadow-slate-200"
          >
            <Warehouse v-if="activeTab === 'stock'" class="h-8 w-8" />
            <Truck v-else-if="activeTab === 'inbound'" class="h-8 w-8" />
            <ClipboardCheck v-else-if="activeTab === 'opname'" class="h-8 w-8" />
            <History v-else class="h-8 w-8" />
          </div>
          <div>
            <h2 class="text-2xl font-black tracking-tight text-slate-900">
              {{
                activeTab === 'stock'
                  ? 'Persediaan Gudang ' + currentBranchName
                  : activeTab === 'inbound'
                    ? 'Penerimaan Barang ' + currentBranchName
                    : activeTab === 'opname'
                      ? 'Stock Opname ' + currentBranchName
                      : 'Riwayat Pengiriman ' + currentBranchName
              }}
            </h2>
            <p class="text-sm text-slate-500 font-medium mt-1">
              {{
                activeTab === 'stock'
                  ? 'Pantau dan kelola stok produk fisik di cabang.'
                  : activeTab === 'inbound'
                    ? 'Konfirmasi barang masuk dari pabrik atau cabang.'
                    : activeTab === 'opname'
                      ? 'Koreksi stok fisik aktual dan pantau riwayat audit opname.'
                      : 'Lihat seluruh logistik masuk yang telah diterima.'
              }}
            </p>
          </div>
        </div>

        <div class="flex flex-wrap gap-3">
          <div class="flex bg-slate-100 p-1.5 rounded-2xl border border-slate-200/60 shadow-inner">
            <button
              @click="activeTab = 'stock'"
              :class="[
                activeTab === 'stock'
                  ? 'bg-white shadow-md text-primary'
                  : 'text-slate-500 hover:text-slate-700',
              ]"
              class="px-6 py-2.5 rounded-xl text-xs font-black uppercase tracking-wider transition-all flex items-center gap-2"
            >
              <Boxes class="h-4 w-4" /> Stok
            </button>
            <button
              @click="activeTab = 'inbound'"
              :class="[
                activeTab === 'inbound'
                  ? 'bg-white shadow-md text-primary'
                  : 'text-slate-500 hover:text-slate-700',
              ]"
              class="px-6 py-2.5 rounded-xl text-xs font-black uppercase tracking-wider transition-all flex items-center gap-2 relative"
            >
              <Truck class="h-4 w-4" /> Pengiriman Pabrik
              <span v-if="groupedPending.length > 0" class="absolute -top-1 -right-1 flex h-4 w-4">
                <span
                  class="animate-ping absolute inline-flex h-full w-full rounded-full bg-rose-400 opacity-75"
                ></span>
                <span
                  class="relative inline-flex rounded-full h-4 w-4 bg-rose-500 text-[9px] items-center justify-center text-white font-black leading-none"
                  >{{ groupedPending.length }}</span
                >
              </span>
            </button>
            <button
              @click="activeTab = 'opname'"
              :class="[
                activeTab === 'opname'
                  ? 'bg-white shadow-md text-primary'
                  : 'text-slate-500 hover:text-slate-700',
              ]"
              class="px-6 py-2.5 rounded-xl text-xs font-black uppercase tracking-wider transition-all flex items-center gap-2"
            >
              <ClipboardCheck class="h-4 w-4" /> Opname
            </button>
            <button
              @click="activeTab = 'history'"
              :class="[
                activeTab === 'history'
                  ? 'bg-white shadow-md text-primary'
                  : 'text-slate-500 hover:text-slate-700',
              ]"
              class="px-6 py-2.5 rounded-xl text-xs font-black uppercase tracking-wider transition-all flex items-center gap-2"
            >
              <History class="h-4 w-4" /> Riwayat
            </button>
          </div>

          <div class="flex flex-wrap gap-2">
            <Button
              v-if="activeTab === 'opname'"
              @click="openGlobalOpname"
              class="rounded-2xl h-12 px-6 bg-slate-900 hover:bg-slate-800 text-white font-black shadow-xl shadow-slate-200 flex gap-2"
            >
              <ClipboardCheck class="h-5 w-5 text-amber-400" /> + Input Opname
            </Button>
            <Button
              v-if="isLinkedPabrikBranch && activeTab === 'stock'"
              variant="outline"
              class="rounded-2xl h-12 px-5 font-bold border-indigo-200 bg-indigo-50/50 text-indigo-700 hover:bg-indigo-100 transition-all shadow-sm flex items-center gap-2"
              @click="handleSyncFromPabrik"
              :disabled="syncingFromPabrik"
            >
              <RefreshCw class="h-4 w-4 text-indigo-600" :class="{ 'animate-spin': syncingFromPabrik }" />
              <span>Sinkronkan dari Pabrik</span>
            </Button>
            <Button
              variant="outline"
              class="rounded-2xl h-12 px-6 font-bold border-slate-200"
              @click="refreshAll"
              :loading="warehouseStore.loading"
            >
              <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': warehouseStore.loading }" />
            </Button>
            <Button
              v-if="activeTab !== 'opname'"
              @click="showScanner = true"
              class="rounded-2xl h-12 px-8 bg-slate-900 hover:bg-slate-800 text-white font-black shadow-xl shadow-slate-200 flex gap-3"
            >
              <ScanLine class="h-5 w-5 text-primary" /> SCAN SJ
            </Button>
          </div>
        </div>
      </div>

      <!-- Main Content -->
      <div class="space-y-6">
        <!-- TAB: STOCK -->
        <div
          v-if="activeTab === 'stock'"
          class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500"
        >
          <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
            <Card
              class="border-none shadow-sm bg-white rounded-3xl p-6 relative overflow-hidden group"
            >
              <div
                class="absolute right-0 top-0 p-4 opacity-5 group-hover:scale-110 transition-transform"
              >
                <Boxes class="h-16 w-16 text-primary" />
              </div>
              <p class="text-xs font-black text-slate-400 uppercase tracking-[0.2em]">Total SKU</p>
              <h3 class="text-3xl font-black text-slate-900 mt-2">
                {{ warehouseStore.inventory.length }}
              </h3>
              <div
                class="mt-4 flex items-center text-[10px] font-bold text-emerald-600 bg-emerald-50 w-fit px-3 py-1 rounded-full"
              >
                TERDATA AKTIF
              </div>
            </Card>

            <Card
              class="border-none shadow-sm bg-white rounded-3xl p-6 relative overflow-hidden group"
            >
              <div
                class="absolute right-0 top-0 p-4 opacity-5 group-hover:scale-110 transition-transform"
              >
                <AlertCircle class="h-16 w-16 text-rose-500" />
              </div>
              <p class="text-xs font-black text-slate-400 uppercase tracking-[0.2em]">Low Stock</p>
              <h3 class="text-3xl font-black text-rose-600 mt-2">
                {{ warehouseStore.inventory.filter((i) => i.quantity < (i.min_limit || 0)).length }}
              </h3>
              <div
                class="mt-4 flex items-center text-[10px] font-bold text-rose-600 bg-rose-50 w-fit px-3 py-1 rounded-full"
              >
                PERLU RESTOCK
              </div>
            </Card>

            <div class="md:col-span-2 flex gap-4">
              <Card
                @click="showRequestModal = true"
                class="flex-1 border-none shadow-sm bg-primary text-white rounded-3xl p-6 cursor-pointer hover:scale-[1.02] transition-all group relative overflow-hidden"
              >
                <div class="relative z-10">
                  <h3 class="text-lg font-black">Request Barang</h3>
                  <p class="text-primary-foreground/70 text-xs mt-1">
                    Ajukan penambahan stok ke pabrik.
                  </p>
                  <Button
                    size="sm"
                    variant="secondary"
                    class="mt-4 rounded-xl font-bold bg-white text-primary"
                    >Buat Request</Button
                  >
                </div>
                <ArrowUpRight
                  class="absolute -right-4 -bottom-4 w-32 h-32 text-white/10 rotate-12"
                />
              </Card>
              <Card
                @click="showAdjustment = true"
                class="flex-1 border-none shadow-sm bg-white border border-slate-100 rounded-3xl p-6 cursor-pointer hover:scale-[1.02] transition-all group relative overflow-hidden"
              >
                <div class="relative z-10">
                  <h3 class="text-lg font-black text-slate-900">Stock Opname</h3>
                  <p class="text-slate-500 text-xs mt-1">Sesuaikan stok fisik yang selisih.</p>
                  <div class="mt-4 flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      class="rounded-xl font-bold border-slate-200"
                      >Adjust Stok</Button
                    >
                    <Button
                      size="sm"
                      variant="secondary"
                      @click.stop="showWarehouseLogs = true"
                      class="rounded-xl font-bold bg-slate-100 text-slate-600"
                      >Riwayat</Button
                    >
                  </div>
                </div>
                <PackageSearch
                  class="absolute -right-4 -bottom-4 w-32 h-32 text-slate-50 rotate-12"
                />
              </Card>
            </div>
          </div>

          <Card
            class="border-none shadow-xl shadow-slate-200/50 bg-white rounded-[2.5rem] overflow-hidden border border-slate-100"
          >
            <div
              class="p-8 border-b border-slate-50 flex flex-col md:flex-row justify-between gap-4"
            >
              <div class="relative w-full md:w-96">
                <Search class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
                <input
                  v-model="stockSearch"
                  placeholder="Cari produk atau batch..."
                  class="w-full pl-12 pr-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-2 focus:ring-primary/20 transition-all"
                />
              </div>
              <div class="flex items-center gap-2">
                <div class="flex bg-slate-100 p-1 rounded-xl">
                  <button
                    @click="stockTypeFilter = 'GOOD'"
                    :class="
                      stockTypeFilter === 'GOOD'
                        ? 'bg-white text-emerald-600 shadow-sm'
                        : 'text-slate-500 hover:text-slate-700'
                    "
                    class="px-4 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all"
                  >
                    Stok Bagus
                  </button>
                  <button
                    @click="stockTypeFilter = 'DAMAGED'"
                    :class="
                      stockTypeFilter === 'DAMAGED'
                        ? 'bg-white text-rose-600 shadow-sm'
                        : 'text-slate-500 hover:text-slate-700'
                    "
                    class="px-4 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all"
                  >
                    Barang Rusak
                  </button>
                  <button
                    @click="stockTypeFilter = 'ALL'"
                    :class="
                      stockTypeFilter === 'ALL'
                        ? 'bg-white text-primary shadow-sm'
                        : 'text-slate-500 hover:text-slate-700'
                    "
                    class="px-4 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all"
                  >
                    Semua
                  </button>
                </div>
                <div class="flex items-center gap-2">
                  <Badge
                    variant="outline"
                    class="hidden md:inline-flex rounded-lg px-3 py-1.5 border-slate-200 text-slate-500 font-bold"
                  >
                    {{ filteredInventory.filter((i) => i.isParent).length }} Produk
                  </Badge>
                  <Button
                    v-if="stockTypeFilter === 'DAMAGED'"
                    @click="openFactoryReturnModal"
                    size="sm"
                    class="rounded-lg h-8 px-4 font-black bg-rose-600 hover:bg-rose-700 text-white text-[10px] shadow-sm"
                  >
                    <Truck class="h-3 w-3 mr-1.5" /> Retur ke Pabrik
                  </Button>
                </div>
              </div>
            </div>
            <div class="p-0">
              <DataTable :columns="stockColumns" :data="filteredInventory">
                <template #cell-product_name="{ row }">
                  <div class="flex items-center gap-2">
                    <button
                      v-if="row.isParent && row.childCount > 1"
                      @click="toggleProductExpand(row.product_id)"
                      class="p-1 hover:bg-slate-100 rounded transition-colors flex items-center justify-center"
                    >
                      <ChevronDown v-if="expandedProductIds.has(row.product_id)" class="h-4 w-4 text-slate-500" />
                      <ChevronRight v-else class="h-4 w-4 text-slate-500" />
                    </button>
                    <div v-else-if="row.isParent" class="w-6"></div>
                    <div v-else-if="row.isChild" class="flex items-center gap-1 text-slate-400 font-mono text-xs pl-6">
                      <span>└─</span>
                    </div>
                    <div class="flex flex-col">
                      <div class="flex items-center gap-2">
                        <span :class="row.isParent ? 'font-black text-slate-900' : 'text-slate-500 font-medium text-xs'">
                          {{ row.isParent ? row.product_name : '' }}
                        </span>
                        <!-- Visual Packaging Badge for Green / White Plastic or Variants -->
                        <span
                          v-if="row.isParent && (row.product_name?.toLowerCase().includes('plastik hijau') || row.product_name?.toLowerCase().includes('hijau'))"
                          class="px-2 py-0.5 rounded-full text-[9px] font-black bg-emerald-100 text-emerald-800 border border-emerald-300 flex items-center gap-1 shadow-sm"
                        >
                          <span class="w-1.5 h-1.5 rounded-full bg-emerald-600"></span>
                          Kemasan Hijau
                        </span>
                        <span
                          v-else-if="row.isParent && (row.product_name?.toLowerCase().includes('plastik putih') || row.product_name?.toLowerCase().includes('putih'))"
                          class="px-2 py-0.5 rounded-full text-[9px] font-black bg-slate-100 text-slate-700 border border-slate-300 flex items-center gap-1 shadow-sm"
                        >
                          <span class="w-1.5 h-1.5 rounded-full bg-slate-400"></span>
                          Kemasan Putih
                        </span>
                      </div>
                      <div v-if="row.isParent && row.product?.pcs_per_unit > 1" class="text-[10px] font-bold text-slate-400 mt-0.5">
                        Konversi: 1 {{ row.product.unit || 'DUS' }} = {{ row.product.pcs_per_unit }} PCS
                      </div>
                    </div>
                  </div>
                </template>
                <template #cell-batch_no="{ row }">
                  <div v-if="row.isParent">
                    <span class="bg-blue-50 text-blue-700 px-2 py-0.5 rounded-md border border-blue-100 font-bold uppercase text-[9px]">
                      Grup Produk
                    </span>
                    <span class="text-[10px] text-slate-400 font-bold ml-1">({{ row.childCount }} Batch)</span>
                  </div>
                  <div v-else class="flex flex-col items-start pl-6">
                    <span class="font-mono text-[11px] text-slate-500">{{ row.batch_no }}</span>
                    <span
                      v-if="row.stock_type === 'DAMAGED'"
                      class="bg-rose-100 text-rose-700 px-1.5 py-0.5 rounded font-bold uppercase text-[9px] mt-1 inline-block"
                    >
                      Rusak
                    </span>
                    <span
                      v-else-if="row.stock_type === 'QUARANTINE'"
                      class="bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded font-bold uppercase text-[9px] mt-1 inline-block"
                    >
                      Karantina
                    </span>
                    <span
                      v-else
                      class="bg-emerald-100 text-emerald-700 px-1.5 py-0.5 rounded font-bold uppercase text-[9px] mt-1 inline-block"
                    >
                      Bagus
                    </span>
                  </div>
                </template>
                <template #cell-quantity="{ row }">
                  <div v-if="row.isParent" class="flex flex-col">
                    <div class="flex items-center gap-2">
                      <span class="text-lg font-black" :class="row.quantity < (row.min_limit || 0) ? 'text-rose-600' : 'text-slate-900'">
                        {{ formatStock(row.quantity, row.product) }}
                      </span>
                      <span
                        v-if="row.quantity < (row.min_limit || 0)"
                        class="bg-rose-50 text-rose-600 text-[8px] font-black px-2 py-0.5 rounded-md border border-rose-100 uppercase animate-pulse"
                      >
                        Low
                      </span>
                    </div>
                    <span class="text-[9px] font-bold text-slate-400 uppercase tracking-tighter">{{ row.quantity }} PCS</span>
                  </div>
                  <div v-else class="flex flex-col pl-6">
                    <span class="text-xs font-bold text-slate-600">
                      {{ formatStock(row.quantity, row.product) }}
                    </span>
                    <span class="text-[9px] font-bold text-slate-400 uppercase tracking-tighter">{{ row.quantity }} PCS</span>
                  </div>
                </template>
                <template #cell-min_limit="{ row }">
                  <div v-if="row.isParent" class="flex items-center gap-2">
                    <span class="font-extrabold text-slate-800">{{ row.min_limit || 0 }} PCS</span>
                    <button
                      @click="openEditThreshold(row)"
                      class="p-1.5 rounded-lg text-slate-400 hover:text-primary hover:bg-slate-50 transition-colors"
                    >
                      <svg
                        class="h-3.5 w-3.5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                        />
                      </svg>
                    </button>
                  </div>
                  <div v-else class="text-slate-300 text-xs pl-6">-</div>
                </template>
                <template #cell-actions="{ row }">
                  <div class="flex items-center gap-1.5">
                    <button
                      v-if="(!row.isParent || (row.isParent && row.childCount === 1)) && row.batch_no !== 'BACKORDER'"
                      @click="openAdjustmentFor(row)"
                      class="flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-lg text-[10px] font-black uppercase tracking-wider transition-colors whitespace-nowrap"
                    >
                      <PackageSearch class="h-3.5 w-3.5" />
                      Opname
                    </button>
                    <button
                      v-if="(!row.isParent || (row.isParent && row.childCount === 1)) && row.batch_no !== 'BACKORDER'"
                      @click="openDeleteStockModal(row)"
                      title="Hapus Data Stok"
                      class="flex items-center gap-1.5 px-2.5 py-1.5 bg-rose-50 hover:bg-rose-100 text-rose-600 hover:text-rose-700 rounded-lg text-[10px] font-black uppercase tracking-wider transition-colors whitespace-nowrap border border-rose-100"
                    >
                      <Trash2 class="h-3.5 w-3.5" />
                      Hapus
                    </button>
                  </div>
                </template>
              </DataTable>
            </div>
          </Card>
        </div>

        <!-- TAB: INBOUND -->
        <div
          v-else-if="activeTab === 'inbound'"
          class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500"
        >
          <div
            v-if="groupedPending.length === 0"
            class="bg-white rounded-[3rem] p-20 text-center border-2 border-dashed border-slate-100 animate-in fade-in zoom-in duration-700"
          >
            <div
              class="w-24 h-24 bg-slate-50 rounded-[2.5rem] flex items-center justify-center mx-auto mb-6"
            >
              <PackageSearch class="h-10 w-10 text-slate-300" />
            </div>
            <h3 class="text-xl font-black text-slate-900">Belum Ada Antrian Inbound</h3>
            <p class="text-sm text-slate-400 mt-2 max-w-xs mx-auto">
              Tidak ada pengiriman atau permintaan yang memerlukan tindakan. Pastikan Anda telah
              memilih <b>Cabang (Branch)</b> yang benar di navigasi atas.
            </p>
            <Button
              @click="refreshAll"
              variant="outline"
              class="mt-8 rounded-2xl px-12 h-14 font-black border-slate-200 hover:bg-slate-50"
              >Refresh Antrian</Button
            >
          </div>

          <Card
            v-else
            class="border-none shadow-xl shadow-slate-200/50 bg-white rounded-[2.5rem] overflow-hidden border border-slate-100"
          >
            <DataTable :columns="inboundColumns" :data="groupedPending">
              <template #cell-doNo="{ row }">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-slate-50 flex items-center justify-center">
                    <Truck v-if="row.status === 'SHIPPED'" class="h-5 w-5 text-blue-600" />
                    <MapPin v-else-if="row.status === 'ARRIVED'" class="h-5 w-5 text-emerald-600" />
                    <ClipboardList v-else class="h-5 w-5 text-orange-600" />
                  </div>
                  <span class="font-black text-slate-900">{{ row.doNo }}</span>
                </div>
              </template>

              <template #cell-from="{ row }">
                <div class="flex items-center text-xs font-bold text-slate-600">
                  <Factory class="mr-2 h-3.5 w-3.5 text-slate-300" /> {{ row.from }}
                </div>
              </template>

              <template #cell-status="{ row }">
                <Badge
                  :variant="
                    row.status === 'SHIPPED'
                      ? 'default'
                      : row.status === 'ARRIVED'
                        ? 'success'
                        : row.status === 'LOADING'
                          ? 'outline'
                          : 'secondary'
                  "
                  :class="row.status === 'LOADING' ? 'border-orange-300 text-orange-600 bg-orange-50' : ''"
                  class="rounded-lg text-[9px] font-black uppercase px-2.5 h-6"
                >
                  {{ formatStatus(row.status) }}
                </Badge>
              </template>

              <template #cell-totalQty="{ row }">
                <div class="flex flex-col gap-1.5">
                  <div
                    v-for="item in row.items"
                    :key="item.id"
                    class="flex flex-col bg-slate-50 p-2 rounded-lg border border-slate-100"
                  >
                    <span
                      class="text-xs font-black text-slate-900 truncate max-w-[200px]"
                      :title="item.product?.name"
                      >{{ item.product?.name || 'Produk Tidak Diketahui' }}</span
                    >
                    <div class="flex items-center gap-1 text-[10px] mt-0.5">
                      <span class="font-bold text-primary"
                        >{{ item.quantity }} {{ item.unit || item.product?.unit || 'UNIT' }}</span
                      >
                      <span class="text-slate-400"
                        >dengan isian
                        {{ item.pcs_per_unit || item.product?.pcs_per_unit || 1 }} PCS</span
                      >
                    </div>
                  </div>
                </div>
              </template>

              <template #cell-totalWeight="{ row }">
                <span class="font-black text-slate-900"
                  >{{ row.totalWeight.toFixed(2) }}
                  <span class="text-[10px] text-slate-400">KG</span></span
                >
              </template>

              <template #cell-targetDate="{ row }">
                <div class="space-y-1.5 text-[11px] font-bold">
                  <div
                    v-if="row.targetDate"
                    class="flex items-center gap-1.5 text-blue-600 bg-blue-50 px-2 py-1 rounded-md w-fit"
                  >
                    <Calendar class="h-3 w-3" />
                    {{
                      new Date(row.targetDate).toLocaleDateString('id-ID', {
                        day: '2-digit',
                        month: 'short',
                        year: 'numeric',
                      })
                    }}
                  </div>
                  <div
                    v-if="row.eta"
                    class="flex items-center gap-1.5 text-emerald-600 bg-emerald-50 px-2 py-1 rounded-md w-fit"
                  >
                    <Clock class="h-3 w-3" />
                    {{
                      new Date(row.eta).toLocaleDateString('id-ID', {
                        day: '2-digit',
                        month: 'short',
                      })
                    }},
                    {{
                      new Date(row.eta).toLocaleTimeString('id-ID', {
                        hour: '2-digit',
                        minute: '2-digit',
                      })
                    }}
                  </div>
                  <div
                    v-if="!row.targetDate && !row.eta && row.shippedAt"
                    class="flex items-center gap-1.5 text-blue-600 bg-blue-50 px-2 py-1 rounded-md w-fit border border-blue-100"
                  >
                    <Truck class="h-3 w-3" /> Dikirim:
                    {{
                      new Date(row.shippedAt).toLocaleDateString('id-ID', {
                        day: '2-digit',
                        month: 'short',
                      })
                    }},
                    {{
                      new Date(row.shippedAt).toLocaleTimeString('id-ID', {
                        hour: '2-digit',
                        minute: '2-digit',
                      })
                    }}
                  </div>
                  <div
                    v-else-if="!row.targetDate && !row.eta"
                    class="flex items-center gap-1.5 text-slate-500 bg-slate-50 px-2 py-1 rounded-md w-fit border border-slate-200"
                  >
                    <Calendar class="h-3 w-3" /> Dibuat:
                    {{
                      new Date(row.date).toLocaleDateString('id-ID', {
                        day: '2-digit',
                        month: 'short',
                      })
                    }}
                  </div>
                </div>
              </template>

              <template #cell-actions="{ row }">
                <div class="flex items-center gap-2">
                  <template v-if="row.status === 'SHIPPED'">
                    <Button
                      @click="handleArrive(row)"
                      :loading="processingId === row.doNo"
                      size="sm"
                      class="rounded-xl h-9 font-bold bg-blue-600 hover:bg-blue-700 text-white px-4 text-[10px]"
                    >
                      KONFIRMASI TIBA
                    </Button>
                    <Button
                      @click="handleReject(row)"
                      variant="ghost"
                      size="sm"
                      class="rounded-xl h-9 font-bold text-slate-400 hover:text-rose-600 px-3"
                      >Tolak</Button
                    >
                  </template>
                  <template v-else-if="row.status === 'ARRIVED'">
                    <Button
                      @click="handleReceive(row)"
                      :loading="processingId === row.doNo"
                      size="sm"
                      class="rounded-xl h-9 font-black bg-emerald-600 hover:bg-emerald-700 text-white px-4 text-[10px]"
                    >
                      TERIMA BARANG
                    </Button>
                  </template>
                  <template v-else-if="row.status === 'APPROVED'">
                    <!-- Delegasi: armada diurus cabang fulfillment (misal Solo), bukan cabang ini -->
                    <template v-if="row.isFulfillmentDelegation">
                      <div
                        class="flex items-center gap-1.5 text-[9px] font-bold text-blue-700 bg-blue-50 px-3 py-1.5 rounded-lg border border-blue-200 whitespace-nowrap"
                      >
                        <Truck class="h-3 w-3 shrink-0" />
                        Diproses{{ row.fulfillmentBranchName ? ' oleh ' + row.fulfillmentBranchName : ' Cabang Lain' }}
                      </div>
                    </template>
                    <!-- Transfer biasa: cabang ini yang assign armada sendiri -->
                    <template v-else>
                      <Button
                        v-if="row.vehicle || row.driver"
                        @click="handleConfirmSJ(row)"
                        :loading="processingId === row.doNo"
                        size="sm"
                        class="rounded-xl h-9 font-black bg-emerald-600 hover:bg-emerald-700 text-white px-4 text-[10px]"
                      >
                        SETUJUI SJ
                      </Button>
                      <div
                        v-else
                        class="text-[9px] font-bold text-amber-500 bg-amber-50 px-3 py-1.5 rounded-lg border border-amber-100"
                      >
                        Menunggu Armada
                      </div>
                      <Button
                        @click="handleDeletePending(row)"
                        :loading="deletingDoNo === row.doNo"
                        variant="ghost"
                        size="sm"
                        class="rounded-xl h-9 font-bold text-slate-400 hover:text-rose-600 hover:bg-rose-50 px-3 text-[10px]"
                        title="Batalkan & Hapus Data"
                      >
                        <Trash2 class="h-3.5 w-3.5 mr-1" /> Hapus
                      </Button>
                    </template>
                  </template>
                  <!-- LOADING delegasi: sedang dimuat oleh cabang lain -->
                  <template v-else-if="row.status === 'LOADING' && row.isFulfillmentDelegation">
                    <div
                      class="flex items-center gap-1.5 text-[9px] font-bold text-orange-700 bg-orange-50 px-3 py-1.5 rounded-lg border border-orange-200 whitespace-nowrap"
                    >
                      <Package class="h-3 w-3 shrink-0" />
                      Sedang Dimuat{{ row.fulfillmentBranchName ? ' di ' + row.fulfillmentBranchName : '' }}
                    </div>
                  </template>

                  <template
                    v-else-if="
                      row.status === 'REQUESTED' && row.items[0]?.initiated_by === 'FACTORY'
                    "
                  >
                    <Button
                      @click="handleApprove(row)"
                      size="sm"
                      class="rounded-xl h-9 font-bold bg-white text-orange-600 border border-orange-100 hover:bg-orange-50 px-4 text-[10px]"
                      >SETUJUI</Button
                    >
                    <Button
                      @click="handleReject(row)"
                      variant="ghost"
                      size="sm"
                      class="rounded-xl h-9 font-bold text-slate-400 hover:text-rose-600 px-3"
                      >Tolak</Button
                    >
                    <Button
                      @click="handleDeletePending(row)"
                      :loading="deletingDoNo === row.doNo"
                      variant="ghost"
                      size="sm"
                      class="rounded-xl h-9 font-bold text-slate-400 hover:text-rose-600 hover:bg-rose-50 px-3 text-[10px]"
                      title="Batalkan & Hapus Data"
                    >
                      <Trash2 class="h-3.5 w-3.5 mr-1" /> Hapus
                    </Button>
                  </template>
                  <template v-else>
                    <span
                      class="text-[10px] font-black text-slate-300 uppercase tracking-widest mr-2"
                      >Pending</span
                    >
                    <Button
                      v-if="row.status === 'REQUESTED' && row.items[0]?.initiated_by === 'BRANCH'"
                      @click="openEditRequestModal(row)"
                      variant="outline"
                      size="sm"
                      class="rounded-xl h-9 font-bold text-slate-600 hover:bg-slate-50 border-slate-200 px-3 text-[10px]"
                    >
                      <Edit2 class="h-3 w-3 mr-1" /> Edit
                    </Button>
                    <Button
                      v-if="row.status === 'REQUESTED' || row.status === 'APPROVED'"
                      @click="handleDeletePending(row)"
                      :loading="deletingDoNo === row.doNo"
                      variant="ghost"
                      size="sm"
                      class="rounded-xl h-9 font-bold text-slate-400 hover:text-rose-600 hover:bg-rose-50 px-3 text-[10px]"
                      title="Batalkan & Hapus Data"
                    >
                      <Trash2 class="h-3.5 w-3.5 mr-1" /> Hapus
                    </Button>
                  </template>

                  <Button
                    @click="printShipment(row)"
                    variant="ghost"
                    size="icon"
                    class="rounded-xl h-9 w-9 text-slate-400 hover:text-slate-600 hover:bg-slate-50 border border-slate-100"
                  >
                    <Printer class="h-4 w-4" />
                  </Button>
                </div>
              </template>
            </DataTable>
          </Card>
        </div>



        <!-- TAB: OPNAME -->
        <div
          v-else-if="activeTab === 'opname'"
          class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500"
        >
          <!-- Top KPI Cards -->
          <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
            <Card class="border-none shadow-sm bg-white rounded-3xl p-6 relative overflow-hidden group">
              <div class="absolute right-0 top-0 p-4 opacity-5 group-hover:scale-110 transition-transform">
                <ClipboardCheck class="h-16 w-16 text-primary" />
              </div>
              <p class="text-xs font-black text-slate-400 uppercase tracking-[0.2em]">Total Log Opname</p>
              <h3 class="text-3xl font-black text-slate-900 mt-2">
                {{ opnameLogs.length }}
              </h3>
              <div class="mt-4 flex items-center text-[10px] font-bold text-slate-500 bg-slate-100 w-fit px-3 py-1 rounded-full">
                AKTIVITAS AUDIT
              </div>
            </Card>

            <Card class="border-none shadow-sm bg-white rounded-3xl p-6 relative overflow-hidden group">
              <div class="absolute right-0 top-0 p-4 opacity-5 group-hover:scale-110 transition-transform">
                <ArrowUpRight class="h-16 w-16 text-emerald-500" />
              </div>
              <p class="text-xs font-black text-slate-400 uppercase tracking-[0.2em]">Koreksi Surplus (+)</p>
              <h3 class="text-3xl font-black text-emerald-600 mt-2">
                +{{ opnameSurplusTotal }} <span class="text-xs font-bold text-slate-400">PCS</span>
              </h3>
              <div class="mt-4 flex items-center text-[10px] font-bold text-emerald-600 bg-emerald-50 w-fit px-3 py-1 rounded-full">
                PENAMBAHAN FISIK
              </div>
            </Card>

            <Card class="border-none shadow-sm bg-white rounded-3xl p-6 relative overflow-hidden group">
              <div class="absolute right-0 top-0 p-4 opacity-5 group-hover:scale-110 transition-transform">
                <AlertTriangle class="h-16 w-16 text-rose-500" />
              </div>
              <p class="text-xs font-black text-slate-400 uppercase tracking-[0.2em]">Koreksi Defisit (-)</p>
              <h3 class="text-3xl font-black text-rose-600 mt-2">
                -{{ opnameDeficitTotal }} <span class="text-xs font-bold text-slate-400">PCS</span>
              </h3>
              <div class="mt-4 flex items-center text-[10px] font-bold text-rose-600 bg-rose-50 w-fit px-3 py-1 rounded-full">
                PENGURANGAN FISIK
              </div>
            </Card>

            <Card
              @click="openGlobalOpname"
              class="border-none shadow-sm bg-gradient-to-br from-amber-500 to-amber-600 text-white rounded-3xl p-6 cursor-pointer hover:scale-[1.02] transition-all group relative overflow-hidden flex flex-col justify-between"
            >
              <div class="relative z-10">
                <span class="text-[10px] font-black uppercase tracking-widest bg-white/20 px-2.5 py-1 rounded-full">
                  Quick Action
                </span>
                <h3 class="text-lg font-black mt-3">Stock Opname Baru</h3>
                <p class="text-xs text-amber-100/90 mt-1">Sesuaikan stok fisik SKU/batch sekarang.</p>
              </div>
              <div class="relative z-10 pt-4 flex items-center gap-2 font-black text-xs">
                <span>Buka Form Opname</span>
                <ArrowUpRight class="w-4 h-4" />
              </div>
              <ClipboardCheck class="absolute -right-4 -bottom-4 w-32 h-32 text-white/15 rotate-12" />
            </Card>
          </div>

          <!-- Main Opname Table Card -->
          <Card
            class="border-none shadow-xl shadow-slate-200/50 bg-white rounded-[2.5rem] overflow-hidden border border-slate-100"
          >
            <div class="p-8 border-b border-slate-50 flex flex-col md:flex-row justify-between gap-4 items-center">
              <div>
                <h3 class="text-lg font-black text-slate-900">Riwayat Penyesuaian Stock Opname</h3>
                <p class="text-xs text-slate-400 font-medium mt-0.5">
                  Daftar seluruh transaksi audit dan penyesuaian kuantitas fisik di cabang {{ currentBranchName }}.
                </p>
              </div>

              <div class="flex flex-wrap items-center gap-3 w-full md:w-auto">
                <div class="relative flex-1 md:w-64">
                  <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
                  <input
                    v-model="opnameSearchQuery"
                    placeholder="Cari produk atau batch..."
                    class="w-full pl-10 pr-4 py-2.5 bg-slate-50 border border-slate-100 rounded-2xl text-xs font-bold focus:bg-white focus:ring-2 focus:ring-primary/20 outline-none transition-all"
                  />
                </div>

                <Button
                  @click="openGlobalOpname"
                  class="rounded-xl px-5 h-10 bg-amber-600 hover:bg-amber-700 text-white font-bold text-xs shadow-md shadow-amber-500/20 flex gap-2"
                >
                  <Plus class="h-4 w-4" /> Opname Baru
                </Button>
              </div>
            </div>

            <div class="overflow-x-auto">
              <div v-if="filteredOpnameLogs.length === 0" class="p-16 text-center text-slate-400 space-y-3">
                <div class="w-16 h-16 rounded-3xl bg-slate-50 flex items-center justify-center mx-auto text-slate-300">
                  <ClipboardCheck class="w-8 h-8" />
                </div>
                <div class="font-bold text-sm text-slate-600">Belum ada riwayat stock opname</div>
                <p class="text-xs text-slate-400 max-w-sm mx-auto">
                  Gunakan tombol "+ Input Opname" untuk mencatat hasil perhitungan fisik gudang.
                </p>
              </div>
              <table v-else class="w-full text-sm text-left">
                <thead class="bg-slate-50/80 border-b border-slate-100 text-[10px] font-black text-slate-400 uppercase tracking-widest">
                  <tr>
                    <th class="px-6 py-4">Waktu</th>
                    <th class="px-6 py-4">Produk & SKU</th>
                    <th class="px-6 py-4">Alasan / Sumber</th>
                    <th class="px-6 py-4 text-center">Tipe</th>
                    <th class="px-6 py-4 text-right">Penyesuaian</th>
                    <th class="px-6 py-4">Petugas</th>
                    <th class="px-6 py-4 text-center">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr
                    v-for="log in filteredOpnameLogs"
                    :key="log.id"
                    class="hover:bg-slate-50/60 transition-colors"
                  >
                    <td class="px-6 py-4">
                      <div class="font-bold text-slate-800 text-xs">
                        {{
                          new Date(log.created_at).toLocaleDateString('id-ID', {
                            day: 'numeric',
                            month: 'short',
                            year: 'numeric',
                          })
                        }}
                      </div>
                      <div class="text-[10px] text-slate-400 font-mono">
                        {{
                          new Date(log.created_at).toLocaleTimeString('id-ID', {
                            hour: '2-digit',
                            minute: '2-digit',
                          })
                        }}
                      </div>
                    </td>
                    <td class="px-6 py-4">
                      <div class="font-black text-slate-900 text-sm">{{ log.product?.name || 'Unknown' }}</div>
                      <div class="flex items-center gap-2 mt-0.5">
                        <span class="text-[11px] font-mono font-bold text-slate-500">{{ log.product?.sku || '-' }}</span>
                        <span class="text-slate-300">•</span>
                        <span class="text-[10px] font-mono text-slate-400">Batch: {{ log.batch_no || '-' }}</span>
                      </div>
                    </td>
                    <td class="px-6 py-4">
                      <span class="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-bold bg-amber-50 text-amber-800 border border-amber-200/60">
                        {{ log.source || 'Stock Opname' }}
                      </span>
                    </td>
                    <td class="px-6 py-4 text-center">
                      <Badge
                        variant="outline"
                        :class="
                          log.type === 'IN'
                            ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                            : 'bg-rose-50 text-rose-700 border-rose-200'
                        "
                        class="text-[10px] font-black uppercase tracking-wider"
                      >
                        {{ log.type === 'IN' ? 'Koreksi +' : 'Koreksi -' }}
                      </Badge>
                    </td>
                    <td class="px-6 py-4 text-right">
                      <span
                        class="font-black text-base"
                        :class="log.type === 'IN' ? 'text-emerald-600' : 'text-rose-600'"
                      >
                        {{ log.type === 'IN' ? '+' : '-' }}{{ log.quantity }} PCS
                      </span>
                    </td>
                    <td class="px-6 py-4">
                      <div class="font-bold text-slate-700 text-xs">{{ log.processed_by || '-' }}</div>
                      <div v-if="log.approved_by && log.approved_by !== log.processed_by" class="text-[10px] text-slate-400">
                        Appr: {{ log.approved_by }}
                      </div>
                    </td>
                    <td class="px-6 py-4 text-center">
                      <Button
                        size="sm"
                        variant="outline"
                        @click="openAdjustmentFor({ product_id: log.product_id, batch_no: log.batch_no, product: log.product })"
                        class="rounded-xl h-8 text-[11px] font-bold border-slate-200 hover:bg-slate-100"
                        title="Opname item ini kembali"
                      >
                        <ClipboardCheck class="h-3.5 w-3.5 mr-1 text-primary" /> Opname
                      </Button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </Card>
        </div>

        <!-- TAB: HISTORY -->
        <div
          v-else-if="activeTab === 'history'"
          class="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500"
        >
          <Card
            class="border-none shadow-xl shadow-slate-200/50 bg-white rounded-[2.5rem] overflow-hidden border border-slate-100"
          >
            <DataTable :columns="historyColumns" :data="filteredHistory">
              <template #cell-doNo="{ row }">
                <div class="flex items-center gap-3">
                  <div
                    class="w-10 h-10 rounded-xl bg-emerald-50 flex items-center justify-center text-emerald-600 shadow-sm border border-emerald-100/50"
                  >
                    <CheckCircle2 class="h-5 w-5" />
                  </div>
                  <span class="font-black text-slate-900">{{ row.doNo }}</span>
                </div>
              </template>
              <template #cell-from="{ row }">
                <div class="flex flex-col gap-1">
                  <div class="flex items-center text-xs font-bold text-slate-600 uppercase tracking-widest">
                    <Factory class="h-3.5 w-3.5 mr-1.5 text-slate-300" /> {{ row.from }}
                  </div>
                  <div v-if="row.to" class="flex items-center text-[10px] font-bold text-slate-400 uppercase tracking-widest">
                    <MapPin class="h-3.5 w-3.5 mr-1.5 text-slate-300" /> Ke: {{ row.to }}
                  </div>
                </div>
              </template>
              <template #cell-receivedAt="{ row }">
                <div class="flex flex-col text-[11px] font-bold text-slate-500 uppercase">
                  <span class="flex items-center gap-1.5"
                    ><Calendar class="h-3 w-3" />
                    {{
                      new Date(row.receivedAt).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'short',
                        year: 'numeric',
                      })
                    }}</span
                  >
                  <span class="flex items-center gap-1.5"
                    ><Clock class="h-3 w-3" />
                    {{
                      new Date(row.receivedAt).toLocaleTimeString('id-ID', {
                        hour: '2-digit',
                        minute: '2-digit',
                      })
                    }}</span
                  >
                </div>
              </template>
              <template #cell-totalWeight="{ row }">
                <span class="text-lg font-black text-slate-900"
                  >{{ row.totalWeight.toFixed(2) }}
                  <span class="text-[10px] font-normal text-slate-400">KG</span></span
                >
              </template>
              <template #cell-items="{ row }">
                <div class="text-xs text-slate-700 font-semibold max-w-[220px] truncate" :title="row.items.map((i: any) => i.product?.name || 'Produk').join(', ')">
                  {{ row.items.map((i: any) => i.product?.name || 'Produk').join(', ') }}
                </div>
              </template>
              <template #cell-notes="{ row }">
                <span class="text-xs text-slate-500 font-bold whitespace-pre-line">{{
                  row.notes || '-'
                }}</span>
              </template>
              <template #cell-actions="{ row }">
                <div class="flex items-center gap-2">
                  <Button
                    @click="viewHistoryDetail(row)"
                    size="sm"
                    variant="outline"
                    class="rounded-xl h-9 font-bold border-slate-200 hover:bg-slate-50 text-[10px]"
                  >
                    DETAIL
                  </Button>
                  <Button
                    @click="printShipment(row)"
                    variant="ghost"
                    size="icon"
                    class="rounded-xl h-9 w-9 text-slate-400 hover:text-slate-650 hover:bg-slate-50 border border-slate-100"
                  >
                    <Printer class="h-4 w-4" />
                  </Button>
                </div>
              </template>
            </DataTable>
          </Card>
        </div>
      </div>
    </div>

    <!-- Modals (Adjustment, Request, Scanner) - Re-used from original but styled better -->

    <!-- Log Mutasi Modal -->
    <div
      v-if="showWarehouseLogs"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-4xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900"
                >Log Mutasi & Riwayat Opname</CardTitle
              >
              <CardDescription>Catatan seluruh perubahan stok fisik di gudang.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showWarehouseLogs = false"
              class="rounded-full"
            >
              <X class="h-6 w-6 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-0 bg-white max-h-[60vh] overflow-y-auto">
          <div
            v-if="warehouseStore.logs.length === 0"
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
                <th class="px-6 py-4">Diproses Oleh</th>
                <th class="px-6 py-4 text-right">Jumlah</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr
                v-for="log in warehouseStore.logs"
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
                      log.source?.startsWith('OPNAME')
                        ? 'bg-orange-50 text-orange-600 border-orange-200'
                        : 'bg-blue-50 text-blue-600 border-blue-200'
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
                  <div class="font-bold text-slate-700 text-xs">{{ log.processed_by || '-' }}</div>
                  <div v-if="log.approved_by && log.approved_by !== log.processed_by" class="text-[10px] text-slate-500 mt-0.5">
                    Approve: {{ log.approved_by }}
                  </div>
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
    <!-- Warehouse Stock Opname Modal (Modular & Scrollable UI/UX) -->
    <WarehouseOpnameModal
      v-model:open="showAdjustment"
      :items="uniqueInventoryForAdjust"
      :all-items="warehouseStore.inventory"
      :initial-item="selectedStockForAdjustment"
      :current-branch-name="currentBranchName"
      :loading="isAdjusting"
      @save="handleOpnameSave"
    />

    <!-- Request Goods Modal -->
    <div
      v-if="showRequestModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-2xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900">Request Barang</CardTitle>
              <CardDescription>Minta penambahan stok ke pabrik pusat.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showRequestModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[60vh] overflow-y-auto custom-scrollbar">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Cabang Peminta</label>
              <input
                type="text"
                :value="currentBranchName"
                disabled
                class="w-full px-5 py-3.5 bg-slate-100 border border-slate-200 rounded-2xl text-sm font-bold text-slate-500 cursor-not-allowed"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Pilih Pabrik Tujuan</label>
              <select
                v-model="requestForm.factory_id"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none appearance-none cursor-pointer"
              >
                <option value="" disabled>Pilih Pabrik...</option>
                <option
                  v-for="factory in factoryStore.factories"
                  :key="factory.id"
                  :value="factory.id"
                >
                  {{ factory.name }}
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Tanggal Permintaan</label>
              <input
                type="date"
                v-model="requestForm.target_shipment_date"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer text-slate-700"
              />
            </div>
          </div>

          <div class="space-y-4">
            <div class="flex justify-between items-center px-1">
              <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest">
                Daftar Produk yang Diminta
              </p>
              <Button
                size="sm"
                variant="ghost"
                class="text-primary font-bold hover:bg-primary/5"
                @click="addRequestItem"
              >
                <Plus class="h-4 w-4 mr-1" /> Tambah Produk
              </Button>
            </div>

            <div
              v-if="requestForm.items.length === 0"
              class="text-center py-12 bg-slate-50 rounded-3xl border border-dashed border-slate-200 text-slate-400 text-sm"
            >
              Belum ada produk dipilih.
            </div>

            <div
              v-for="(item, index) in requestForm.items"
              :key="index"
              class="flex flex-col gap-4 bg-slate-50 p-6 rounded-[2rem] border border-slate-100 relative group transition-all hover:shadow-md hover:bg-white"
            >
              <div class="grid grid-cols-1 md:grid-cols-12 gap-4">
                <div class="md:col-span-4 space-y-1">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Pilih Produk</label
                  >
                  <SearchSelect
                    v-model="item.product_id"
                    :options="finishedGoodsOptions"
                    placeholder="Pilih Produk..."
                    search-placeholder="Cari nama atau SKU produk..."
                    size="sm"
                    @change="() => onProductSelect(index)"
                  />
                </div>

                <div class="md:col-span-2 space-y-1">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Jumlah</label
                  >
                  <input
                    type="number"
                    v-model.number="item.quantity"
                    class="w-full px-4 py-3 bg-white border border-slate-100 rounded-xl text-xs font-black outline-none shadow-sm focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                <div class="md:col-span-2 space-y-1">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Satuan</label
                  >
                  <select
                    v-model="item.unit"
                    @change="handleUnitChange(item)"
                    class="w-full px-4 py-3 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none shadow-sm focus:ring-2 focus:ring-primary/20 appearance-none cursor-pointer"
                  >
                    <option v-for="u in UNIT_OPTIONS" :key="u" :value="u">{{ u }}</option>
                  </select>
                </div>

                <div class="md:col-span-2 space-y-1">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Isi/Satuan</label
                  >
                  <input
                    type="number"
                    v-model.number="item.pcs_per_unit"
                    class="w-full px-4 py-3 bg-white border border-slate-100 rounded-xl text-xs font-black outline-none shadow-sm focus:ring-2 focus:ring-primary/20"
                  />
                </div>

                <div class="md:col-span-1 flex items-end">
                  <Button
                    variant="ghost"
                    size="icon"
                    class="text-rose-500 h-12 w-full rounded-xl hover:bg-rose-50 border border-transparent hover:border-rose-100"
                    @click="removeRequestItem(index)"
                  >
                    <Trash2 class="h-5 w-5" />
                  </Button>
                </div>
              </div>

              <div
                v-if="item.product_id"
                class="flex items-center justify-between px-4 py-2 bg-white/50 rounded-xl border border-dashed border-slate-200"
              >
                <div class="flex items-center gap-2">
                  <Package class="h-3 w-3 text-primary" />
                  <span class="text-[10px] font-bold text-slate-500">Konversi Total:</span>
                  <span class="text-[10px] font-black text-slate-900"
                    >{{ item.quantity }} {{ item.unit }} × {{ item.pcs_per_unit }} PCS =
                  </span>
                  <span class="text-xs font-black text-primary"
                    >{{ item.quantity * (item.pcs_per_unit || 0) }} PCS</span
                  >
                </div>
                <div class="text-[9px] font-black text-slate-300 uppercase tracking-[0.2em]">
                  Estimasi Muatan
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Catatan Tambahan</label>
            <textarea
              v-model="requestForm.notes"
              rows="2"
              placeholder="Sebutkan alasan atau instruksi khusus..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none resize-none focus:ring-2 focus:ring-primary/20"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-8"
            @click="showRequestModal = false"
            >Batal</Button
          >
          <Button
            class="rounded-2xl px-12 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90"
            @click="handleRequest"
            :loading="isRequesting"
            >Kirim Permintaan</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- Edit Request Modal -->
    <div
      v-if="showEditRequestModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-2xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900">Edit Request Barang</CardTitle>
              <CardDescription>Ubah detail request yang masih berstatus Pending.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showEditRequestModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[60vh] overflow-y-auto custom-scrollbar">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Cabang Peminta</label>
              <div
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold text-slate-600"
              >
                {{ masterDataStore.branches.find((b) => b.id === masterDataStore.selectedWarehouseBranchId)?.name || 'Semua Cabang' }}
              </div>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Minta Dari Pabrik</label>
              <select
                v-model="editRequestForm.factory_id"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-2 focus:ring-primary/20"
              >
                <option value="" disabled>Pilih Pabrik...</option>
                <option
                  v-for="factory in factoryStore.factories"
                  :key="factory.id"
                  :value="factory.id"
                >
                  {{ factory.name }}
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Tanggal Permintaan</label>
              <input
                type="date"
                v-model="editRequestForm.target_shipment_date"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-2 focus:ring-primary/20 cursor-pointer text-slate-700"
              />
            </div>
          </div>

          <div class="space-y-4">
            <div class="flex justify-between items-center border-b pb-2">
              <label class="text-sm font-black text-slate-900">Daftar Barang (Item)</label>
              <Button
                @click="addEditRequestItem"
                variant="outline"
                size="sm"
                class="h-8 text-xs font-bold rounded-xl text-primary border-primary/20 hover:bg-primary/5"
              >
                <Plus class="h-3.5 w-3.5 mr-1" /> Tambah Baris
              </Button>
            </div>

            <div class="space-y-3">
              <div
                v-for="(item, index) in editRequestForm.items"
                :key="index"
                class="flex gap-3 items-end bg-slate-50 p-4 rounded-2xl border border-slate-100 relative group"
              >
                <div class="flex-1 space-y-2">
                  <label class="text-[10px] font-bold text-slate-400 uppercase">Produk</label>
                  <SearchSelect
                    v-model="item.product_id"
                    :options="finishedGoodsOptions"
                    placeholder="Pilih Produk..."
                    search-placeholder="Cari nama atau SKU produk..."
                    size="sm"
                    @change="() => onEditProductSelect(index)"
                  />
                </div>
                <div class="w-24 space-y-2">
                  <label class="text-[10px] font-bold text-slate-400 uppercase">Jumlah</label>
                  <input
                    type="number"
                    v-model="item.quantity"
                    min="1"
                    class="w-full px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none text-center"
                  />
                </div>
                <div class="w-24 space-y-2">
                  <label class="text-[10px] font-bold text-slate-400 uppercase">Satuan</label>
                  <select
                    v-model="item.unit"
                    class="w-full px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none"
                  >
                    <option value="PCS">PCS</option>
                    <option value="BOX">BOX</option>
                    <option value="BAL">BAL</option>
                    <option value="LUSIN">LUSIN</option>
                    <option value="KARTON">KARTON</option>
                  </select>
                </div>
                <Button
                  @click="removeEditRequestItem(index)"
                  variant="ghost"
                  size="icon"
                  class="h-9 w-9 absolute -top-2 -right-2 bg-white shadow-sm border border-slate-100 rounded-full text-slate-300 hover:text-rose-500 hover:bg-rose-50 opacity-0 group-hover:opacity-100 transition-all"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Catatan Tambahan</label>
            <textarea
              v-model="editRequestForm.notes"
              rows="2"
              placeholder="Sebutkan alasan atau instruksi khusus..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none resize-none focus:ring-2 focus:ring-primary/20"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-between items-center gap-4">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-6 text-rose-600 hover:text-rose-700 hover:bg-rose-50 flex items-center gap-2"
            @click="handleDeleteFromEditModal"
            :loading="isRequesting"
          >
            <Trash2 class="h-4 w-4" />
            Hapus Permintaan
          </Button>
          <div class="flex gap-4">
            <Button
              variant="ghost"
              class="rounded-2xl font-bold h-12 px-8"
              @click="showEditRequestModal = false"
              >Batal</Button
            >
            <Button
              class="rounded-2xl px-12 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90"
              @click="handleUpdateRequest"
              :loading="isRequesting"
              >Simpan Perubahan</Button
            >
          </div>
        </CardFooter>
      </Card>
    </div>

    <!-- Scanner Modal -->
    <div
      v-if="showScanner"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-md shadow-2xl rounded-[3rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="bg-slate-900 text-white p-8">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black flex items-center gap-3">
                <ScanLine class="h-8 w-8 text-primary" />
                Scan SJ
              </CardTitle>
              <CardDescription class="text-slate-400"
                >Verifikasi pengiriman secara instan.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showScanner = false"
              class="rounded-full text-white/50 hover:text-white"
            >
              <Plus class="h-6 w-6 rotate-45" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-8 bg-white">
          <div
            class="aspect-square bg-slate-50 rounded-[2.5rem] border-2 border-dashed border-slate-200 flex flex-col items-center justify-center text-slate-400 gap-4 relative overflow-hidden group"
          >
            <div
              class="w-48 h-48 border-2 border-primary/20 rounded-[2.5rem] flex items-center justify-center relative z-10 bg-white shadow-xl shadow-slate-100"
            >
              <div
                class="absolute inset-0 border-4 border-primary/5 rounded-[2.4rem] animate-pulse"
              ></div>
              <QrCode class="w-24 h-24 text-primary opacity-20" />
              <div
                class="absolute w-full h-[2px] bg-primary/60 top-1/2 left-0 -translate-y-1/2 animate-bounce shadow-[0_0_10px_primary]"
              ></div>
            </div>
            <p
              class="text-[10px] font-black uppercase tracking-[0.4em] text-primary/60 z-10 bg-white px-5 py-2 rounded-full border border-primary/10 shadow-sm"
            >
              Scanner Aktif
            </p>
          </div>

          <div class="space-y-3">
            <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
              >Input Manual Nomor SJ</label
            >
            <div class="relative">
              <input
                v-model="scannedDO"
                placeholder="DO-YYYYMMDD-XXXX"
                class="w-full h-16 rounded-2xl border-2 border-slate-100 focus:border-primary focus:ring-4 focus:ring-primary/5 font-mono text-center text-xl tracking-widest font-black outline-none transition-all"
                @keyup.enter="handleScanSubmit"
              />
            </div>
          </div>
        </CardContent>
        <CardFooter class="flex flex-col gap-3 p-8 bg-slate-50/50 border-t">
          <Button
            class="w-full font-black h-14 rounded-2xl text-base shadow-xl shadow-primary/20 bg-primary hover:bg-primary/90 text-white"
            @click="handleScanSubmit"
            :loading="isScanning"
          >
            KONFIRMASI PENERIMAAN
          </Button>
          <Button
            variant="ghost"
            class="w-full font-bold text-slate-400"
            @click="showScanner = false"
            >TUTUP</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- Receiving Confirmation Modal -->
    <div
      v-if="showReceiveModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-2xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900"
                >Konfirmasi Bongkar Muat</CardTitle
              >
              <CardDescription
                >Cek kondisi fisik barang dan masukkan jumlah yang diterima.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showReceiveModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[60vh] overflow-y-auto custom-scrollbar">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="bg-blue-50 border border-blue-100 p-4 rounded-2xl flex gap-4 items-center">
              <div
                class="w-10 h-10 rounded-xl bg-blue-100 flex items-center justify-center text-blue-600 shadow-sm"
              >
                <Truck class="h-5 w-5" />
              </div>
              <div>
                <p
                  class="text-[10px] font-black text-blue-400 uppercase tracking-widest leading-none mb-1"
                >
                  Surat Jalan
                </p>
                <p class="text-sm font-black text-blue-900">{{ receivingShipment?.doNo }}</p>
              </div>
            </div>
            <div
              v-if="receivingShipment?.eta"
              class="bg-emerald-50 border border-emerald-100 p-4 rounded-2xl flex gap-4 items-center"
            >
              <div
                class="w-10 h-10 rounded-xl bg-emerald-100 flex items-center justify-center text-emerald-600 shadow-sm"
              >
                <Clock class="h-5 w-5" />
              </div>
              <div>
                <p
                  class="text-[10px] font-black text-emerald-400 uppercase tracking-widest leading-none mb-1"
                >
                  Estimasi Tiba
                </p>
                <p class="text-sm font-black text-emerald-900">
                  {{
                    new Date(receivingShipment.eta).toLocaleDateString('id-ID', {
                      day: 'numeric',
                      month: 'short',
                    })
                  }},
                  {{
                    new Date(receivingShipment.eta).toLocaleTimeString('id-ID', {
                      hour: '2-digit',
                      minute: '2-digit',
                    })
                  }}
                </p>
              </div>
            </div>
          </div>
          <div class="space-y-4">
            <div
              v-for="(item, index) in receiveForm.items"
              :key="item.id"
              class="p-6 bg-slate-50 rounded-[2rem] border border-slate-100 space-y-4"
            >
              <div class="flex justify-between items-center">
                <h4 class="font-black text-slate-900">{{ item.product_name }}</h4>
                <Badge
                  variant="outline"
                  class="rounded-lg px-2 py-0.5 bg-white text-slate-400 font-bold border-slate-200"
                >
                  Sesuai SJ: {{ item.expected_quantity }} {{ item.unit }}
                </Badge>
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Kondisi Baik (GOOD)</label
                  >
                  <div class="flex gap-2">
                    <div class="relative flex-1">
                      <input
                        type="number"
                        v-model.number="item.actual_received_quantity_units"
                        class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 font-black text-lg outline-none"
                      />
                      <div
                        class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-emerald-500 uppercase"
                      >
                        {{ item.unit || 'UNIT' }}
                      </div>
                    </div>
                    <div v-if="item.pcs_per_unit > 1" class="relative flex-1">
                      <input
                        type="number"
                        v-model.number="item.actual_received_quantity_pcs"
                        class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 font-black text-lg outline-none"
                      />
                      <div
                        class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-emerald-500 uppercase"
                      >
                        PCS
                      </div>
                    </div>
                  </div>
                  <div
                    v-if="item.pcs_per_unit > 1"
                    class="px-3 pt-1 text-[9px] font-bold text-slate-400"
                  >
                    =
                    {{
                      item.actual_received_quantity_units * item.pcs_per_unit +
                      item.actual_received_quantity_pcs
                    }}
                    PCS (Total)
                  </div>
                </div>
                <div class="space-y-2">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Rusak/Pecah (DAMAGED)</label
                  >
                  <div class="flex gap-2">
                    <div class="relative flex-1">
                      <input
                        type="number"
                        v-model.number="item.damaged_quantity_units"
                        class="w-full h-12 rounded-xl border border-slate-200 focus:border-rose-500 px-4 font-black text-lg outline-none"
                      />
                      <div
                        class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-rose-500 uppercase"
                      >
                        {{ item.unit || 'UNIT' }}
                      </div>
                    </div>
                    <div v-if="item.pcs_per_unit > 1" class="relative flex-1">
                      <input
                        type="number"
                        v-model.number="item.damaged_quantity_pcs"
                        class="w-full h-12 rounded-xl border border-slate-200 focus:border-rose-500 px-4 font-black text-lg outline-none"
                      />
                      <div
                        class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-rose-500 uppercase"
                      >
                        PCS
                      </div>
                    </div>
                  </div>
                  <div
                    v-if="item.pcs_per_unit > 1"
                    class="px-3 pt-1 text-[9px] font-bold text-slate-400"
                  >
                    =
                    {{
                      item.damaged_quantity_units * item.pcs_per_unit + item.damaged_quantity_pcs
                    }}
                    PCS (Total)
                  </div>
                </div>
              </div>

              <div
                v-if="isSelisih(item)"
                class="flex items-center gap-2 text-[10px] font-bold text-orange-600 bg-orange-50 px-3 py-1.5 rounded-lg border border-orange-100"
              >
                <AlertCircle class="h-3 w-3" /> Selisih: {{ getSelisihQty(item) }} {{ item.unit }}
              </div>
            </div>
          </div>

          <div
            v-if="receiveForm.items.some((i: any) => (i.damaged_quantity_units > 0 || i.damaged_quantity_pcs > 0))"
            class="flex items-start gap-2.5 p-3.5 bg-rose-50 border border-rose-200/80 rounded-2xl text-xs text-rose-800"
          >
            <RotateCcw class="w-4 h-4 text-rose-600 mt-0.5 shrink-0" />
            <div>
              <p class="font-bold">Penyesuaian Otomatis Piutang Cabang</p>
              <p class="text-[11px] text-rose-700/90 mt-0.5">
                Barang rusak yang diinput akan otomatis masuk ke stok karantina cabang dan nilainya langsung memotong saldo Piutang Cabang (Tagihan Pabrik).
              </p>
            </div>
          </div>

          <div class="space-y-2 pt-4 border-t border-slate-100">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">
              Catatan Penerimaan / Komplain Keterlambatan
            </label>
            <textarea
              v-model="receiveForm.notes"
              rows="3"
              class="w-full rounded-2xl border border-slate-200 focus:border-primary p-4 text-sm font-semibold outline-none resize-none"
              placeholder="Masukkan catatan jika ada keterlambatan pengiriman, selisih barang, atau masalah pengiriman pabrik lainnya..."
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-8"
            @click="showReceiveModal = false"
            >Batal</Button
          >
          <Button
            class="rounded-2xl px-12 h-12 shadow-xl shadow-emerald-200 font-black text-white bg-emerald-600 hover:bg-emerald-700"
            @click="submitReceive"
            :loading="isSubmittingReceive"
            >Update Stok & Selesaikan</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- Rejection Reason Modal -->
    <div
      v-if="showRejectModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-lg shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-rose-600">Tolak Pengiriman</CardTitle>
              <CardDescription
                >Berikan alasan penolakan agar pabrik dapat melakukan penjadwalan
                ulang.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showRejectModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Alasan Penolakan</label>
            <textarea
              v-model="rejectionForm.reason"
              rows="4"
              placeholder="Contoh: Gudang penuh pada tanggal tersebut, mohon geser ke hari lain..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none resize-none focus:ring-2 focus:ring-rose-500/20 transition-all border-rose-100"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-8"
            @click="showRejectModal = false"
            >Batal</Button
          >
          <Button
            class="rounded-2xl px-12 h-12 shadow-xl shadow-rose-200 font-black text-white bg-rose-600 hover:bg-rose-700"
            @click="submitReject"
            :loading="isRejecting"
            >Konfirmasi Tolak</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- Threshold Configuration Modal -->
    <div
      v-if="showThresholdModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-md shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900">Setting Limit Stok</CardTitle>
              <CardDescription>Atur threshold stok minimum untuk produk di cabang.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showThresholdModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white">
          <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-1">
            <p
              class="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none mb-1"
            >
              Produk
            </p>
            <p class="text-sm font-black text-slate-900">
              {{ selectedThresholdItem?.product?.name }}
            </p>
            <p class="text-[10px] font-medium text-slate-500 mt-1">
              Target: {{ masterDataStore.branches.find(b => b.id === (masterDataStore.selectedWarehouseBranchId !== 'ALL_BRANCHES' ? masterDataStore.selectedWarehouseBranchId : selectedThresholdItem?.branch_id))?.name || 'Semua Cabang (Global)' }}
            </p>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Limit Threshold Baru (PCS)</label>
            <div class="relative">
              <input
                type="number"
                v-model.number="thresholdForm.limit"
                class="w-full h-14 rounded-2xl border border-slate-200 focus:border-primary px-5 font-black text-xl outline-none"
              />
              <div
                class="absolute right-4 top-1/2 -translate-y-1/2 text-[10px] font-black text-slate-400 uppercase"
              >
                PCS
              </div>
            </div>
            <p class="text-[10px] text-slate-400 ml-1 mt-1">
              Jika stok di bawah nilai ini, Pabrik akan menerima alert restock.
            </p>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-8"
            @click="showThresholdModal = false"
            >Batal</Button
          >
          <Button
            class="rounded-2xl px-12 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90"
            @click="submitThreshold"
            :loading="isSavingThreshold"
            >Simpan Limit</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- Factory Return Modal -->
    <div
      v-if="showFactoryReturnModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-2xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-rose-600">Retur Barang Rusak</CardTitle>
              <CardDescription>Kembalikan barang rusak/cacat ke pabrik.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showFactoryReturnModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[60vh] overflow-y-auto custom-scrollbar">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Pilih Pabrik Tujuan</label>
            <select
              v-model="factoryReturnForm.to_factory_id"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none appearance-none cursor-pointer"
            >
              <option value="" disabled>Pilih Pabrik...</option>
              <option
                v-for="factory in factoryStore.factories"
                :key="factory.id"
                :value="factory.id"
              >
                {{ factory.name }}
              </option>
            </select>
          </div>

          <div class="space-y-4">
            <div class="flex justify-between items-center px-1">
              <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest">
                Daftar Barang yang Diretur
              </p>
              <Button
                size="sm"
                variant="ghost"
                class="text-rose-600 font-bold hover:bg-rose-50"
                @click="addFactoryReturnItem"
              >
                <Plus class="h-4 w-4 mr-1" /> Tambah Barang
              </Button>
            </div>

            <div
              v-if="factoryReturnForm.items.length === 0"
              class="text-center py-12 bg-slate-50 rounded-3xl border border-dashed border-slate-200 text-slate-400 text-sm"
            >
              Belum ada produk yang dipilih untuk diretur ke pabrik.
            </div>

            <div
              v-for="(item, index) in factoryReturnForm.items"
              :key="index"
              class="flex flex-col gap-4 bg-slate-50 p-6 rounded-[2rem] border border-slate-200 relative group transition-all hover:shadow-md hover:bg-white"
            >
              <!-- Row 1: Tipe Retur & Subtipe Kerusakan -->
              <div class="grid grid-cols-1 md:grid-cols-12 gap-3 pb-3 border-b border-slate-200/60">
                <div class="md:col-span-4 space-y-1">
                  <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">
                    Kondisi Barang
                  </label>
                  <select
                    v-model="item.return_type"
                    @change="onReturnTypeChange(index)"
                    class="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-black outline-none shadow-sm focus:ring-2 focus:ring-slate-900/10"
                  >
                    <option value="DAMAGED">🔴 Barang Rusak (Damaged)</option>
                    <option value="GOOD">🟢 Barang Bagus (Overstock / Recall)</option>
                  </select>
                </div>

                <div v-if="item.return_type === 'DAMAGED'" class="md:col-span-4 space-y-1">
                  <label class="text-[10px] font-black text-rose-500 uppercase tracking-widest ml-1">
                    Kategori Kerusakan
                  </label>
                  <select
                    v-model="item.damage_type"
                    class="w-full px-3.5 py-2.5 bg-white border border-rose-200 rounded-xl text-xs font-bold text-rose-700 outline-none shadow-sm focus:ring-2 focus:ring-rose-500/20"
                  >
                    <option value="BOTTLE_PACKAGING">Botol / Kemasan Rusak (Bisa Repack)</option>
                    <option value="PRODUCT_CONTENT">Isi Rusak / Basi / Mengendap</option>
                    <option value="EXPIRED">Kadaluwarsa (Expired)</option>
                    <option value="TOTAL_DAMAGE">Rusak Total / Pecah Hancur</option>
                    <option value="OTHER">Lain-lain / Cacat Batch</option>
                  </select>
                </div>

                <div :class="item.return_type === 'DAMAGED' ? 'md:col-span-4' : 'md:col-span-8'" class="space-y-1">
                  <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1">
                    Keterangan Penyebab / Alasan
                  </label>
                  <input
                    type="text"
                    v-model="item.damage_reason"
                    placeholder="Contoh: Tutup botol bocor, label sobek, dll..."
                    class="w-full px-3.5 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-medium outline-none shadow-sm focus:ring-2 focus:ring-slate-900/10"
                  />
                </div>
              </div>

              <!-- Row 2: Pilih Produk, Qty, dan Delete -->
              <div class="grid grid-cols-1 md:grid-cols-12 gap-4">
                <div class="md:col-span-6 space-y-1">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Pilih Produk (Stok Gudang {{ item.return_type === 'GOOD' ? 'Bagus' : 'Rusak' }})</label
                  >
                  <select
                    v-model="item.selectedStock"
                    @change="onFactoryReturnProductSelect(index)"
                    class="w-full px-4 py-3 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none shadow-sm focus:ring-2 focus:ring-slate-900/10"
                  >
                    <option :value="null" disabled>Pilih Produk...</option>
                    <option
                      v-for="stock in getAvailableStocksForReturn(item.return_type)"
                      :key="stock.id"
                      :value="stock"
                    >
                      {{ stock.product?.name }} (Stok {{ stock.stock_type === 'GOOD' ? 'Bagus' : 'Rusak' }}:
                      {{ formatStock(stock.quantity, stock.product) }}) - Batch:
                      {{ stock.batch_no }}
                    </option>
                  </select>
                </div>

                <div class="md:col-span-5 space-y-1">
                  <div class="flex justify-between items-center ml-1">
                    <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                      Jumlah Retur ({{ item.selectedStock?.product?.unit || 'UNIT' }})
                    </label>
                    <span v-if="item.selectedStock" class="text-[10px] font-bold text-slate-500">
                      Tersedia: {{ formatStock(item.selectedStock.quantity, item.selectedStock.product) }}
                    </span>
                  </div>
                  <input
                    type="number"
                    min="1"
                    :max="item.selectedStock ? Math.max(1, Math.floor(item.selectedStock.quantity / (item.selectedStock.product?.pcs_per_unit || 1))) : undefined"
                    v-model.number="item.quantity"
                    class="w-full px-4 py-3 bg-white border border-slate-200 rounded-xl text-xs font-black outline-none shadow-sm focus:ring-2 focus:ring-slate-900/10"
                  />
                  <p v-if="item.selectedStock && (item.selectedStock.product?.pcs_per_unit || 1) > 1" class="text-[10px] font-semibold text-slate-400 ml-1">
                    = {{ (item.quantity || 0) * (item.selectedStock.product?.pcs_per_unit || 1) }} PCS (1 {{ item.selectedStock.product?.unit || 'UNIT' }} = {{ item.selectedStock.product?.pcs_per_unit }} PCS)
                  </p>
                </div>

                <div class="md:col-span-1 flex items-end">
                  <Button
                    variant="ghost"
                    size="icon"
                    class="text-rose-500 h-12 w-full rounded-xl hover:bg-rose-50 border border-transparent hover:border-rose-100"
                    @click="removeFactoryReturnItem(index)"
                  >
                    <Trash2 class="h-5 w-5" />
                  </Button>
                </div>

                <div v-if="item.selectedStock" class="md:col-span-12 -mt-2 px-1">
                  <span class="text-[10px] font-bold text-slate-700 bg-slate-100 px-3 py-1 rounded-full border border-slate-200 inline-block">
                    Tujuan Pengiriman: Gudang Pusat {{ getWarehouseNumber(item.selectedStock.product?.name, item.selectedStock.product?.category).number }} ({{ getWarehouseNumber(item.selectedStock.product?.name, item.selectedStock.product?.category).type }})
                  </span>
                </div>
              </div>
            </div>

            <!-- Delivery Instructions Summary -->
            <div v-if="factoryReturnForm.items.some(i => i.selectedStock)" class="bg-blue-50/70 p-5 rounded-2xl border border-blue-100 space-y-2">
              <h4 class="text-[11px] font-black text-blue-700 uppercase tracking-widest">Petunjuk Pengiriman Retur</h4>
              <p class="text-xs text-blue-600 font-medium">
                Sistem mendeteksi produk yang Anda pilih harus dikirim ke Gudang Pusat berikut di pabrik:
              </p>
              <div class="flex flex-wrap gap-2 pt-1">
                <span
                  v-for="wNo in Array.from(new Set(factoryReturnForm.items.filter(i => i.selectedStock).map(i => getWarehouseNumber(i.selectedStock.product?.name, i.selectedStock.product?.category).number))).sort((a,b) => a-b)"
                  :key="wNo"
                  class="text-[10px] font-black bg-blue-100 text-blue-700 px-3 py-1.5 rounded-xl border border-blue-200"
                >
                  Gudang Pusat {{ wNo }} ({{ getWarehouseNumber(factoryReturnForm.items.find(i => i.selectedStock && getWarehouseNumber(i.selectedStock.product?.name, i.selectedStock.product?.category).number === wNo).selectedStock.product?.name, factoryReturnForm.items.find(i => i.selectedStock && getWarehouseNumber(i.selectedStock.product?.name, i.selectedStock.product?.category).number === wNo).selectedStock.product?.category).type }})
                </span>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Catatan Tambahan</label>
            <textarea
              v-model="factoryReturnForm.notes"
              rows="2"
              placeholder="Detail kerusakan atau keterangan lainnya..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none resize-none focus:ring-2 focus:ring-rose-500/20"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-8"
            @click="showFactoryReturnModal = false"
            >Batal</Button
          >
          <Button
            class="rounded-2xl px-12 h-12 shadow-xl shadow-rose-200 font-black text-white bg-rose-600 hover:bg-rose-700"
            @click="submitFactoryReturn"
            :loading="isSubmittingFactoryReturn"
            >Kirim Retur</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- History Detail Modal -->
    <div
      v-if="showHistoryDetailModal && selectedHistoryShipment"
      class="fixed inset-0 bg-slate-900/70 backdrop-blur-md z-[100] flex items-center justify-center p-4"
      @click.self="showHistoryDetailModal = false"
    >
      <div
        class="w-full max-w-3xl bg-white rounded-[2rem] shadow-2xl border border-slate-100 overflow-hidden animate-in zoom-in-95 fade-in duration-300 flex flex-col max-h-[90vh]"
      >
        <!-- Modal Header with gradient -->
        <div class="relative bg-gradient-to-br from-slate-800 via-slate-900 to-slate-800 p-7 overflow-hidden flex-shrink-0">
          <!-- Background decoration -->
          <div class="absolute inset-0 opacity-10">
            <div class="absolute -top-8 -right-8 w-40 h-40 rounded-full bg-white/20 blur-2xl"></div>
            <div class="absolute -bottom-8 -left-8 w-32 h-32 rounded-full bg-emerald-400/30 blur-2xl"></div>
          </div>

          <div class="relative flex items-start justify-between gap-4">
            <div class="flex items-center gap-4">
              <div class="w-14 h-14 rounded-2xl bg-emerald-500/20 border border-emerald-400/30 flex items-center justify-center flex-shrink-0">
                <CheckCircle2 class="h-7 w-7 text-emerald-400" />
              </div>
              <div>
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-[10px] font-black text-emerald-400 uppercase tracking-[0.15em] bg-emerald-500/10 border border-emerald-400/20 px-2.5 py-1 rounded-full">
                    RECEIVED
                  </span>
                </div>
                <h2 class="text-2xl font-black text-white leading-tight">
                  {{ selectedHistoryShipment.doNo }}
                </h2>
                <p class="text-sm text-slate-400 mt-0.5">
                  Surat Jalan dari {{ selectedHistoryShipment.from }}
                </p>
              </div>
            </div>
            <button
              @click="showHistoryDetailModal = false"
              class="w-9 h-9 rounded-xl bg-white/10 hover:bg-white/20 border border-white/10 flex items-center justify-center transition-all flex-shrink-0 mt-0.5"
            >
              <X class="h-4 w-4 text-white/70" />
            </button>
          </div>

          <!-- Quick stats row -->
          <div class="relative grid grid-cols-3 gap-3 mt-6">
            <div class="bg-white/5 border border-white/10 rounded-xl p-3">
              <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mb-1">Total Berat</div>
              <div class="text-lg font-black text-white">
                {{ selectedHistoryShipment.totalWeight.toFixed(1) }}
                <span class="text-xs font-normal text-slate-400">KG</span>
              </div>
            </div>
            <div class="bg-white/5 border border-white/10 rounded-xl p-3">
              <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mb-1">Total Item</div>
              <div class="text-lg font-black text-white">
                {{ selectedHistoryShipment.items.length }}
                <span class="text-xs font-normal text-slate-400">SKU</span>
              </div>
            </div>
            <div class="bg-white/5 border border-white/10 rounded-xl p-3">
              <div class="text-[10px] font-bold text-slate-400 uppercase tracking-wider mb-1">Barang Rusak</div>
              <div class="text-lg font-black"
                :class="selectedHistoryShipment.items.reduce((s: number, i: any) => s + (i.damaged_quantity || 0), 0) > 0 ? 'text-rose-400' : 'text-emerald-400'"
              >
                {{ selectedHistoryShipment.items.reduce((s: number, i: any) => s + (i.damaged_quantity || 0), 0) }}
                <span class="text-xs font-normal text-slate-400">PCS</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Modal Body -->
        <div class="flex-1 overflow-y-auto">
          <!-- Info Cards -->
          <div class="p-6 grid grid-cols-2 gap-4">
            <!-- Asal Pengiriman -->
            <div class="bg-slate-50 border border-slate-100 rounded-2xl p-4">
              <div class="flex items-center gap-2 mb-2">
                <div class="w-7 h-7 rounded-lg bg-blue-100 flex items-center justify-center">
                  <Factory class="h-3.5 w-3.5 text-blue-600" />
                </div>
                <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider">Asal Pengiriman</span>
              </div>
              <p class="font-extrabold text-slate-800 text-sm">{{ selectedHistoryShipment.from }}</p>
            </div>

            <!-- Waktu Penerimaan -->
            <div class="bg-slate-50 border border-slate-100 rounded-2xl p-4">
              <div class="flex items-center gap-2 mb-2">
                <div class="w-7 h-7 rounded-lg bg-emerald-100 flex items-center justify-center">
                  <Calendar class="h-3.5 w-3.5 text-emerald-600" />
                </div>
                <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider">Diterima Pada</span>
              </div>
              <p class="font-extrabold text-slate-800 text-sm">
                {{
                  new Date(selectedHistoryShipment.receivedAt).toLocaleDateString('id-ID', {
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })
                }}
              </p>
              <p class="text-xs text-slate-500 font-medium mt-0.5">
                {{
                  new Date(selectedHistoryShipment.receivedAt).toLocaleTimeString('id-ID', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })
                }}
              </p>
            </div>

            <!-- Driver -->
            <div class="bg-slate-50 border border-slate-100 rounded-2xl p-4">
              <div class="flex items-center gap-2 mb-2">
                <div class="w-7 h-7 rounded-lg bg-purple-100 flex items-center justify-center">
                  <svg class="h-3.5 w-3.5 text-purple-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
                  </svg>
                </div>
                <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider">Driver</span>
              </div>
              <p class="font-extrabold text-slate-800 text-sm">
                <template v-if="typeof selectedHistoryShipment.items[0]?.driver === 'string'">
                  {{ selectedHistoryShipment.items[0]?.driver }}
                </template>
                <template v-else-if="selectedHistoryShipment.items[0]?.driver">
                  {{ selectedHistoryShipment.items[0]?.driver.first_name }} {{ selectedHistoryShipment.items[0]?.driver.last_name || '' }}
                </template>
                <template v-else>
                  Tidak ada data
                </template>
              </p>
            </div>

            <!-- Armada -->
            <div class="bg-slate-50 border border-slate-100 rounded-2xl p-4">
              <div class="flex items-center gap-2 mb-2">
                <div class="w-7 h-7 rounded-lg bg-orange-100 flex items-center justify-center">
                  <svg class="h-3.5 w-3.5 text-orange-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="1" y="3" width="15" height="13" rx="1"/><path d="M16 8h4l3 3v5h-7V8z"/><circle cx="5.5" cy="18.5" r="2.5"/><circle cx="18.5" cy="18.5" r="2.5"/>
                  </svg>
                </div>
                <span class="text-[10px] font-black text-slate-400 uppercase tracking-wider">Armada / Plat</span>
              </div>
              <p class="font-extrabold text-slate-800 text-sm">
                {{ selectedHistoryShipment.items[0]?.vehicle?.name || '-' }}
              </p>
              <p class="text-xs font-mono font-bold text-slate-500 mt-0.5">
                {{ selectedHistoryShipment.items[0]?.vehicle?.plate_number || selectedHistoryShipment.items[0]?.vehicle?.license_plate || '-' }}
              </p>
            </div>
          </div>

          <!-- Items Table -->
          <div class="px-6 pb-2">
            <h3 class="text-sm font-black text-slate-700 uppercase tracking-wider mb-3 flex items-center gap-2">
              <div class="w-1 h-4 bg-blue-500 rounded-full"></div>
              Daftar Barang
            </h3>
            <div class="border border-slate-100 rounded-2xl overflow-hidden">
              <table class="w-full text-left text-xs">
                <thead class="bg-gradient-to-r from-slate-50 to-slate-50/50 border-b border-slate-100">
                  <tr>
                    <th class="px-4 py-3 font-black text-slate-500 uppercase tracking-wider text-[10px]">Produk</th>
                    <th class="px-4 py-3 font-black text-slate-500 uppercase tracking-wider text-[10px] text-center">Batch</th>
                    <th class="px-4 py-3 font-black text-slate-500 uppercase tracking-wider text-[10px] text-right">Dikirim</th>
                    <th class="px-4 py-3 font-black text-emerald-600 uppercase tracking-wider text-[10px] text-right">Diterima</th>
                    <th class="px-4 py-3 font-black text-rose-500 uppercase tracking-wider text-[10px] text-right">Rusak</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-50">
                  <tr
                    v-for="item in selectedHistoryShipment.items"
                    :key="item.id"
                    class="hover:bg-slate-50/50 transition-colors"
                    :class="(item.damaged_quantity || 0) > 0 ? 'bg-rose-50/30' : ''"
                  >
                    <td class="px-4 py-3.5">
                      <div class="font-bold text-slate-900">{{ item.product?.name || 'Produk' }}</div>
                      <div class="text-[10px] text-slate-400 font-medium">{{ item.product?.sku || '' }}</div>
                    </td>
                    <td class="px-4 py-3.5 text-center">
                      <span class="font-mono text-[10px] font-bold bg-slate-100 text-slate-600 px-2 py-0.5 rounded-md">
                        {{ item.batch_no || 'DEFAULT' }}
                      </span>
                    </td>
                    <td class="px-4 py-3.5 text-right">
                      <span class="font-bold text-slate-500">{{ item.quantity }}</span>
                      <span class="text-[10px] text-slate-400 ml-1">{{ item.unit || item.product?.unit || 'UNIT' }}</span>
                    </td>
                    <td class="px-4 py-3.5 text-right">
                      <span class="font-black text-emerald-600">{{ item.actual_received_quantity ?? item.quantity }}</span>
                      <span class="text-[10px] text-slate-400 ml-1">{{ item.unit || item.product?.unit || 'UNIT' }}</span>
                    </td>
                    <td class="px-4 py-3.5 text-right">
                      <span
                        class="font-black"
                        :class="(item.damaged_quantity || 0) > 0 ? 'text-rose-600' : 'text-slate-300'"
                      >
                        {{ item.damaged_quantity || 0 }}
                      </span>
                      <span class="text-[10px] text-slate-400 ml-1">{{ item.unit || item.product?.unit || 'UNIT' }}</span>
                    </td>
                  </tr>
                </tbody>
                <!-- Summary row -->
                <tfoot class="bg-slate-50 border-t border-slate-200">
                  <tr>
                    <td colspan="2" class="px-4 py-3 text-[10px] font-black text-slate-400 uppercase">Total</td>
                    <td class="px-4 py-3 text-right font-black text-slate-600 text-xs">
                      {{ selectedHistoryShipment.items.reduce((s: number, i: any) => s + (i.quantity || 0), 0) }}
                    </td>
                    <td class="px-4 py-3 text-right font-black text-emerald-600 text-xs">
                      {{ selectedHistoryShipment.items.reduce((s: number, i: any) => s + (i.actual_received_quantity ?? i.quantity ?? 0), 0) }}
                    </td>
                    <td class="px-4 py-3 text-right font-black text-rose-600 text-xs">
                      {{ selectedHistoryShipment.items.reduce((s: number, i: any) => s + (i.damaged_quantity || 0), 0) }}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>

          <!-- Notes -->
          <div v-if="selectedHistoryShipment.notes" class="px-6 pb-6 mt-4">
            <div class="bg-amber-50 border border-amber-100 rounded-2xl p-4">
              <div class="flex items-center gap-2 mb-2">
                <div class="w-6 h-6 rounded-lg bg-amber-100 flex items-center justify-center">
                  <svg class="h-3.5 w-3.5 text-amber-600" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/>
                  </svg>
                </div>
                <span class="text-xs font-black text-amber-700 uppercase tracking-wider">Catatan / Komplain</span>
              </div>
              <p class="text-sm text-amber-800 font-medium whitespace-pre-line leading-relaxed">{{ selectedHistoryShipment.notes }}</p>
            </div>
          </div>

          <div v-else class="pb-4"></div>
        </div>

        <!-- Footer Actions -->
        <div class="flex items-center justify-end gap-3 p-6 border-t border-slate-100 bg-slate-50/80 flex-shrink-0">
          <button
            @click="showHistoryDetailModal = false"
            class="px-5 py-2.5 rounded-xl border border-slate-200 text-slate-600 text-sm font-bold hover:bg-slate-100 transition-colors"
          >
            Tutup
          </button>
          <button
            @click="openCorrectionModal(selectedHistoryShipment); showHistoryDetailModal = false"
            class="px-5 py-2.5 rounded-xl bg-amber-500 hover:bg-amber-600 text-white text-sm font-bold flex items-center gap-2 transition-colors shadow-sm shadow-amber-200"
          >
            <Edit2 class="w-4 h-4" />
            Edit Penerimaan
          </button>
          <button
            @click="printShipment(selectedHistoryShipment)"
            class="px-5 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-700 text-white text-sm font-bold flex items-center gap-2 transition-colors shadow-sm shadow-blue-200"
          >
            <Printer class="w-4 h-4" />
            Cetak Surat Jalan
          </button>
        </div>
      </div>
    </div>

    <!-- Correction Modal -->
    <div
      v-if="showCorrectionModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-2xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300 bg-white"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900"
                >Koreksi Penerimaan Barang</CardTitle
              >
              <CardDescription
                >Koreksi jumlah barang masuk dan catat riwayat alasan koreksi untuk audit.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showCorrectionModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[60vh] overflow-y-auto custom-scrollbar">
          <div class="bg-amber-50 border border-amber-100 p-4 rounded-2xl flex gap-4 items-center">
            <div
              class="w-10 h-10 rounded-xl bg-amber-100 flex items-center justify-center text-amber-600 shadow-sm"
            >
              <FileEdit class="h-5 w-5" />
            </div>
            <div>
              <p
                class="text-[10px] font-black text-amber-400 uppercase tracking-widest leading-none mb-1"
              >
                No. Surat Jalan
              </p>
              <p class="text-sm font-black text-amber-900">{{ correctionForm.doNo }}</p>
            </div>
          </div>
          
          <div class="space-y-4">
            <div
              v-for="(item, index) in correctionForm.items"
              :key="item.id"
              class="p-6 bg-slate-50 rounded-[2rem] border border-slate-100 space-y-4"
            >
              <div class="flex justify-between items-center">
                <h4 class="font-black text-slate-900">{{ item.product_name }}</h4>
                <Badge
                  variant="outline"
                  class="rounded-lg px-2 py-0.5 bg-white text-slate-400 font-bold border-slate-200"
                >
                  Sesuai SJ: {{ item.expected_quantity }} {{ item.unit }}
                </Badge>
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Kondisi Baik (GOOD)</label
                  >
                  <div class="flex gap-2">
                    <div class="relative flex-1">
                      <input
                        type="number"
                        v-model.number="item.actual_received_quantity_units"
                        class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 font-black text-lg outline-none"
                      />
                      <div
                        class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-emerald-500 uppercase"
                      >
                        {{ item.unit || 'UNIT' }}
                      </div>
                    </div>
                    <div v-if="item.pcs_per_unit > 1" class="relative flex-1">
                      <input
                        type="number"
                        v-model.number="item.actual_received_quantity_pcs"
                        class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 font-black text-lg outline-none"
                      />
                      <div
                        class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-emerald-500 uppercase"
                      >
                        PCS
                      </div>
                    </div>
                  </div>
                  <div
                    v-if="item.pcs_per_unit > 1"
                    class="px-3 pt-1 text-[9px] font-bold text-slate-400"
                  >
                    =
                    {{
                      item.actual_received_quantity_units * item.pcs_per_unit +
                      item.actual_received_quantity_pcs
                    }}
                    PCS (Total)
                  </div>
                </div>
                <div class="space-y-2">
                  <label
                    class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                    >Rusak/Pecah (DAMAGED)</label
                  >
                  <div class="flex gap-2">
                    <div class="relative flex-1">
                      <input
                        type="number"
                        v-model.number="item.damaged_quantity_units"
                        class="w-full h-12 rounded-xl border border-slate-200 focus:border-rose-500 px-4 font-black text-lg outline-none"
                      />
                      <div
                        class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-rose-500 uppercase"
                      >
                        {{ item.unit || 'UNIT' }}
                      </div>
                    </div>
                    <div v-if="item.pcs_per_unit > 1" class="relative flex-1">
                      <input
                        type="number"
                        v-model.number="item.damaged_quantity_pcs"
                        class="w-full h-12 rounded-xl border border-slate-200 focus:border-rose-500 px-4 font-black text-lg outline-none"
                      />
                      <div
                        class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-black text-rose-500 uppercase"
                      >
                        PCS
                      </div>
                    </div>
                  </div>
                  <div
                    v-if="item.pcs_per_unit > 1"
                    class="px-3 pt-1 text-[9px] font-bold text-slate-400"
                  >
                    =
                    {{
                      item.damaged_quantity_units * item.pcs_per_unit + item.damaged_quantity_pcs
                    }}
                    PCS (Total)
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-2 pt-4 border-t border-slate-100">
            <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">
              Alasan Koreksi / Edit Penerimaan <span class="text-rose-500">*</span>
            </label>
            <textarea
              v-model="correctionForm.reason"
              rows="3"
              class="w-full rounded-2xl border border-slate-200 focus:border-primary p-4 text-sm font-semibold outline-none resize-none"
              placeholder="Masukkan alasan pengeditan data penerimaan ini secara rinci (misal: kesalahan hitung saat bongkar truk, dsb)..."
              required
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-8"
            @click="showCorrectionModal = false"
            >Batal</Button
          >
          <Button
            class="rounded-2xl px-12 h-12 shadow-xl shadow-amber-200 font-black text-white bg-amber-600 hover:bg-amber-700"
            @click="submitCorrection"
            :loading="isSubmittingCorrection"
            >Simpan Koreksi Stok</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- MODAL KONFIRMASI HAPUS DATA STOK -->
    <div
      v-if="showDeleteStockModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4 animate-in fade-in duration-200"
    >
      <Card
        class="w-full max-w-lg shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-200 bg-white"
      >
        <CardHeader class="p-8 border-b bg-rose-50/60">
          <div class="flex justify-between items-start">
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 rounded-2xl bg-rose-100 text-rose-600 flex items-center justify-center shadow-inner">
                <Trash2 class="h-6 w-6" />
              </div>
              <div>
                <CardTitle class="text-xl font-black text-rose-950">Hapus Data Stok</CardTitle>
                <CardDescription class="text-rose-600/80 font-medium text-xs mt-0.5">
                  Tindakan permanen dan berisiko tinggi.
                </CardDescription>
              </div>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="closeDeleteStockModal"
              class="rounded-full text-slate-400 hover:text-slate-600 hover:bg-rose-100/50"
            >
              <X class="h-5 w-5" />
            </Button>
          </div>
        </CardHeader>

        <CardContent class="p-8 space-y-5 bg-white">
          <!-- Item Detail Info Box -->
          <div class="bg-slate-50 border border-slate-200/70 rounded-2xl p-4 space-y-3">
            <div class="flex justify-between items-start border-b border-slate-200/50 pb-2">
              <div>
                <p class="text-[10px] font-black uppercase tracking-wider text-slate-400">Produk</p>
                <p class="font-black text-slate-900 text-sm mt-0.5">{{ selectedStockToDelete?.product?.name || selectedStockToDelete?.product_name }}</p>
              </div>
              <Badge variant="outline" class="font-bold text-[10px]">
                Batch: {{ selectedStockToDelete?.batch_no || 'DEFAULT' }}
              </Badge>
            </div>
            <div class="bg-white p-2.5 rounded-xl border border-slate-100">
              <span class="text-[10px] font-bold text-slate-400 block uppercase">Stok Fisik</span>
              <span class="font-black text-slate-800 text-sm">
                {{ selectedStockToDelete?.quantity || 0 }} PCS
              </span>
            </div>
          </div>

          <!-- General Danger Warning If Physical Stock > 0 -->
          <div
            v-if="(selectedStockToDelete?.quantity || 0) > 0"
            class="bg-amber-50 border border-amber-200 text-amber-900 rounded-2xl p-4 flex gap-3 text-xs leading-relaxed items-start"
          >
            <AlertTriangle class="h-5 w-5 text-amber-600 shrink-0 mt-0.5" />
            <div>
              <p class="font-black text-amber-950">Perhatian: Sisa Stok Fisik Masih Ada</p>
              <p class="text-amber-800 mt-0.5">
                Stok fisik sebanyak <b>{{ selectedStockToDelete?.quantity }} PCS</b> akan dihapus dan tercatat otomatis sebagai mutasi keluar (OUT / MANUAL_DELETE) di log gudang.
              </p>
            </div>
          </div>

          <!-- Delete Form Inputs -->
          <div class="space-y-4 pt-1">
            <div class="space-y-1.5">
              <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
                Alasan Penghapusan <span class="text-rose-500">*</span>
              </label>
              <textarea
                v-model="deleteStockReason"
                rows="2"
                placeholder="Contoh: Salah input batch / pembersihan barang rusak / duplikat"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-semibold outline-none focus:ring-2 focus:ring-rose-400 focus:bg-white transition-all resize-none"
              ></textarea>
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-black text-slate-700 uppercase tracking-wider">
                Ketik <span class="text-rose-600 font-mono font-black bg-rose-50 px-1.5 py-0.5 rounded border border-rose-200">HAPUS</span> untuk konfirmasi <span class="text-rose-500">*</span>
              </label>
              <input
                type="text"
                v-model="deleteStockConfirmationText"
                placeholder="Ketik HAPUS di sini..."
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-mono font-bold outline-none focus:ring-2 focus:ring-rose-400 focus:bg-white transition-all tracking-wider"
              />
            </div>
          </div>
        </CardContent>

        <CardFooter class="p-6 bg-slate-50/80 border-t flex justify-end gap-3">
          <Button
            variant="ghost"
            @click="closeDeleteStockModal"
            class="rounded-xl font-bold text-slate-600 hover:bg-slate-200/70"
          >
            Batal
          </Button>
          <Button
            variant="destructive"
            :disabled="deleteStockConfirmationText.trim().toUpperCase() !== 'HAPUS' || !deleteStockReason.trim() || isDeletingStock"
            @click="handleConfirmDeleteStock"
            class="rounded-xl font-black bg-rose-600 hover:bg-rose-700 text-white shadow-lg shadow-rose-200 flex items-center gap-2 px-6"
          >
            <RefreshCw v-if="isDeletingStock" class="h-4 w-4 animate-spin" />
            <Trash2 v-else class="h-4 w-4" />
            Hapus Stok Permanen
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- PRINT AREA (Only visible when printing) -->
    <div class="print-only hidden print:block">
      <SuratJalanPrint
        v-if="shipmentToPrint"
        :shipment="shipmentToPrint"
        class="hidden print:block"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, watch, computed, nextTick } from 'vue'
import { useWarehouseStore } from '@/stores/warehouse'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card/index'
import { Badge } from '@/components/ui/badge'
import DataTable from '@/components/DataTable.vue'
import {
  Boxes,
  Truck,
  ScanLine,
  RefreshCw,
  CheckCircle2,
  PackageSearch,
  QrCode,
  Search,
  AlertCircle,
  Plus,
  Trash2,
  ArrowUpRight,
  Warehouse,
  History,
  Calendar,
  Clock,
  MapPin,
  Factory,
  Package,
  User,
  AlertTriangle,
  Printer,
  X,
  ChevronRight,
  ChevronDown,
  Edit2,
  Send,
  ClipboardCheck,
} from 'lucide-vue-next'
import WarehouseOpnameModal from '@/components/warehouse/WarehouseOpnameModal.vue'
import SuratJalanPrint from '@/components/factory/SuratJalanPrint.vue'
import SearchSelect, { type SelectOption } from '@/components/SearchSelect.vue'
import { groupProductsByCategory } from '@/utils/productUnits'
import { toast } from 'vue-sonner'

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

const warehouseStore = useWarehouseStore()
const masterDataStore = useMasterDataStore()
const factoryStore = useFactoryStore()
const UNIT_OPTIONS = [
  'KARTON',
  'DUS',
  'BOX',
  'KRAT',
  'PCS',
  'BAL',
  'PACK',
  'KG',
  'GRAM',
  'SACHET',
  'JERIGEN',
  'BOTOL',
]

const isBaseUnit = (unit: string): boolean => {
  const u = (unit || '').toUpperCase()
  return (
    u === 'PCS' ||
    u === 'BOTOL' ||
    u === 'SACHET' ||
    u === 'JERIGEN' ||
    u === 'PACK' ||
    u === 'GRAM' ||
    u === 'KG'
  )
}

const handleUnitChange = (item: any) => {
  if (isBaseUnit(item.unit)) {
    item.pcs_per_unit = 1
  } else {
    const product = factoryStore.products.find((p) => p.id === item.product_id)
    if (product) {
      item.pcs_per_unit = product.pcs_per_unit || 1
    } else {
      item.pcs_per_unit = 1
    }
  }
}

// Formatters
const formatStock = (qty: number, product: any) => {
  if (qty == null) return '0'
  const pcsPerUnit = product?.pcs_per_unit || 1

  if (pcsPerUnit > 1) {
    const units = Math.floor(qty / pcsPerUnit)
    const pcs = qty % pcsPerUnit
    if (units > 0 && pcs > 0) return `${units} ${product.unit || 'UNIT'} ${pcs} PCS`
    if (units > 0) return `${units} ${product.unit || 'UNIT'}`
    return `${pcs} PCS`
  }
  return `${qty} ${product?.unit || 'PCS'}`
}

const finishedGoodsOptions = computed<SelectOption[]>(() => {
  const grouped = groupProductsByCategory(factoryStore.products)
  const products = [...grouped.finishedGoods, ...grouped.tradingProducts, ...grouped.giftProducts]
  return products.map((p) => ({
    value: p.id,
    label: p.name,
    sublabel: p.sku ? `SKU: ${p.sku}` : (p.category ? `Kategori: ${p.category}` : undefined),
    badge: p.unit || undefined,
    badgeColor: 'bg-slate-100 text-slate-600',
  }))
})

// State
const activeTab = ref('stock')
const showAdjustment = ref(false)
const showRequestModal = ref(false)
const showEditRequestModal = ref(false)
const selectedEditRequest = ref<any>(null)
const showScanner = ref(false)
const showReceiveModal = ref(false)
const showWarehouseLogs = ref(false)
const showRejectModal = ref(false)
const showThresholdModal = ref(false)
const showFactoryReturnModal = ref(false)
const isSavingThreshold = ref(false)
const isSubmittingFactoryReturn = ref(false)
const isDeletingShipment = ref(false)
const deletingDoNo = ref('')

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

const selectedThresholdItem = ref<any>(null)
const thresholdForm = reactive({
  limit: 0,
})
const receivingShipment = ref<any>(null)
const processingId = ref('')
const scannedDO = ref('')
const isScanning = ref(false)
const isAdjusting = ref(false)
const isRequesting = ref(false)
const isRejecting = ref(false)
const isSubmittingReceive = ref(false)
const stockSearch = ref('')
const stockTypeFilter = ref('GOOD')
const historySearch = ref('')
const shipmentToPrint = ref<any>(null)
const selectedHistoryShipment = ref<any>(null)
const showHistoryDetailModal = ref(false)
const showCorrectionModal = ref(false)
const isSubmittingCorrection = ref(false)
const correctionForm = reactive({
  reason: '',
  items: [] as any[],
  doNo: '',
})

const openCorrectionModal = (group: any) => {
  correctionForm.doNo = group.doNo
  correctionForm.items = group.items.map((item: any) => {
    const pcsPerUnit = item.pcs_per_unit || 1
    const actualReceivedUnits = Math.floor(item.actual_received_quantity / pcsPerUnit)
    const actualReceivedPcs = item.actual_received_quantity % pcsPerUnit
    const damagedUnits = Math.floor(item.damaged_quantity / pcsPerUnit)
    const damagedPcs = item.damaged_quantity % pcsPerUnit

    return {
      id: item.id,
      product_name: item.product?.name,
      unit: item.unit || item.product?.unit || 'UNIT',
      expected_quantity: item.quantity,
      actual_received_quantity_units: actualReceivedUnits,
      actual_received_quantity_pcs: actualReceivedPcs,
      damaged_quantity_units: damagedUnits,
      damaged_quantity_pcs: damagedPcs,
      pcs_per_unit: pcsPerUnit,
    }
  })
  correctionForm.reason = ''
  showCorrectionModal.value = true
  showHistoryDetailModal.value = false
}

const submitCorrection = async () => {
  if (!correctionForm.reason) {
    toast.error('Alasan koreksi wajib diisi')
    return
  }

  isSubmittingCorrection.value = true
  try {
    for (const item of correctionForm.items) {
      const totalActualPcs =
        item.actual_received_quantity_units * item.pcs_per_unit + item.actual_received_quantity_pcs
      const totalDamagedPcs =
        item.damaged_quantity_units * item.pcs_per_unit + item.damaged_quantity_pcs

      await warehouseStore.correctShipmentReceipt(item.id, {
        actual_received_quantity: totalActualPcs,
        damaged_quantity: totalDamagedPcs,
        reason: correctionForm.reason,
      })
    }
    showCorrectionModal.value = false
    toast.success('Penerimaan barang berhasil dikoreksi dan stok diperbarui.')
    refreshAll()
  } catch (err) {
    toast.error('Gagal melakukan koreksi: ' + getErrorMessage(err))
  } finally {
    isSubmittingCorrection.value = false
  }
}

const viewHistoryDetail = (row: any) => {
  selectedHistoryShipment.value = row
  showHistoryDetailModal.value = true
}

const expandedProductIds = ref(new Set<string>())
const toggleProductExpand = (productId: string) => {
  if (expandedProductIds.value.has(productId)) {
    expandedProductIds.value.delete(productId)
  } else {
    expandedProductIds.value.add(productId)
  }
  expandedProductIds.value = new Set(expandedProductIds.value)
}

const currentBranchName = computed(() => {
  const branch = masterDataStore.branches.find((b) => b.id === masterDataStore.selectedWarehouseBranchId)
  return branch ? branch.name : 'Semua Cabang'
})

const isLinkedPabrikBranch = computed(() => {
  const branch = masterDataStore.branches.find((b) => b.id === masterDataStore.selectedWarehouseBranchId)
  if (!branch) return false
  const nameLower = (branch.name || '').toLowerCase()
  return nameLower.includes('pabrik')
})

const syncingFromPabrik = ref(false)
const handleSyncFromPabrik = async () => {
  if (!confirm(`Apakah Anda yakin ingin menyinkronkan stok gudang ${currentBranchName.value} 1:1 dari data stok pabrik?`)) {
    return
  }
  syncingFromPabrik.value = true
  const tid = toast.loading('Menyinkronkan stok dari pabrik...')
  try {
    await factoryStore.syncAllWarehouseStock()
    await warehouseStore.fetchInventory()
    toast.success(`Stok gudang ${currentBranchName.value} berhasil disinkronkan 1:1 dari data pabrik.`, { id: tid })
  } catch (err: any) {
    toast.error('Gagal menyinkronkan stok: ' + getErrorMessage(err), { id: tid })
  } finally {
    syncingFromPabrik.value = false
  }
}

const printShipment = (group: any) => {
  shipmentToPrint.value = group
  setTimeout(() => {
    window.print()
  }, 300)
}

const receiveForm = reactive({
  items: [] as any[],
  notes: '',
})

const rejectionForm = reactive({
  reason: '',
  group: null as any,
})

const selectedStockForAdjustment = ref<any | null>(null)
const opnameSearchQuery = ref('')

const openGlobalOpname = () => {
  selectedStockForAdjustment.value = null
  showAdjustment.value = true
}

const uniqueInventoryForAdjust = computed(() => {
  const unique = new Map()

  // 1. Masukkan batch fisik nyata dari stok gudang (eksklusikan antrean BACKORDER)
  for (const item of warehouseStore.inventory) {
    if (item.batch_no === 'BACKORDER') continue
    const key = `${item.product_id}_${item.batch_no || 'DEFAULT'}`
    if (!unique.has(key)) {
      unique.set(key, item)
    }
  }

  // 2. Tambahkan produk master yang belum memiliki batch fisik di cabang ini (misal produk baru / stok 0 yang hanya ada backorder)
  const activeBranchId =
    (masterDataStore.selectedWarehouseBranchId !== 'ALL_BRANCHES' && masterDataStore.selectedWarehouseBranchId) ||
    warehouseStore.inventory[0]?.branch_id ||
    ''

  const existingProductIds = new Set(Array.from(unique.values()).map((u: any) => u.product_id))

  for (const p of factoryStore.products || []) {
    if (!existingProductIds.has(p.id)) {
      const defaultKey = `${p.id}_DEFAULT`
      unique.set(defaultKey, {
        product_id: p.id,
        product: p,
        batch_no: 'DEFAULT',
        branch_id: activeBranchId,
        quantity: 0,
        reserved_quantity: 0,
        stock_type: 'GOOD',
      })
    }
  }

  return Array.from(unique.values())
})

const opnameLogs = computed(() => {
  return warehouseStore.logs.filter((log: any) => {
    const src = (log.source || '').toUpperCase()
    return (
      src.includes('OPNAME') ||
      src.includes('ADJUST') ||
      src.includes('KOREKSI') ||
      src.includes('PENYESUAIAN')
    )
  })
})

const opnameSurplusTotal = computed(() => {
  return opnameLogs.value
    .filter((l: any) => l.type === 'IN')
    .reduce((acc: number, cur: any) => acc + (cur.quantity || 0), 0)
})

const opnameDeficitTotal = computed(() => {
  return opnameLogs.value
    .filter((l: any) => l.type === 'OUT')
    .reduce((acc: number, cur: any) => acc + (cur.quantity || 0), 0)
})

const filteredOpnameLogs = computed(() => {
  let logs = opnameLogs.value
  if (!opnameSearchQuery.value.trim()) return logs
  const q = opnameSearchQuery.value.toLowerCase()
  return logs.filter((l: any) => {
    const name = l.product?.name?.toLowerCase() || ''
    const sku = l.product?.sku?.toLowerCase() || ''
    const batch = l.batch_no?.toLowerCase() || ''
    const source = l.source?.toLowerCase() || ''
    const processor = l.processed_by?.toLowerCase() || ''
    return (
      name.includes(q) ||
      sku.includes(q) ||
      batch.includes(q) ||
      source.includes(q) ||
      processor.includes(q)
    )
  })
})

const handleOpnameSave = async (payload: {
  selectedItem: any
  goodQuantity: number
  currentGoodQuantity: number
  damagedQuantity: number
  currentDamagedQuantity: number
  reason: string
  autoReclassify: boolean
}) => {
  if (!payload.selectedItem) {
    toast.error('Silakan pilih produk terlebih dahulu.')
    return
  }
  if (!payload.reason) {
    toast.error('Alasan penyesuaian wajib diisi.')
    return
  }

  const rawBatch = payload.selectedItem.batch_no
  const safeBatchNo = (rawBatch && rawBatch !== 'BACKORDER') ? rawBatch : 'DEFAULT'

  isAdjusting.value = true
  try {
    const promises = []
    const branchId =
      payload.selectedItem.branch_id || masterDataStore.selectedWarehouseBranchId

    // Adjust GOOD if changed
    if (payload.goodQuantity !== payload.currentGoodQuantity) {
      promises.push(
        warehouseStore.adjustStock({
          branch_id: branchId,
          product_id: payload.selectedItem.product_id,
          batch_no: safeBatchNo,
          quantity: payload.goodQuantity,
          reason: payload.reason,
          stock_type: 'GOOD',
        })
      )
    }

    // Adjust DAMAGED if changed
    if (payload.damagedQuantity !== payload.currentDamagedQuantity) {
      promises.push(
        warehouseStore.adjustStock({
          branch_id: branchId,
          product_id: payload.selectedItem.product_id,
          batch_no: safeBatchNo,
          quantity: payload.damagedQuantity,
          reason: payload.reason,
          stock_type: 'DAMAGED',
        })
      )
    }

    if (promises.length === 0) {
      toast.info('Tidak ada perubahan kuantitas fisik.')
      isAdjusting.value = false
      return
    }

    await Promise.all(promises)
    showAdjustment.value = false
    selectedStockForAdjustment.value = null
    await refreshAll()
    toast.success('Hasil Stock Opname berhasil disimpan.')
  } catch (err) {
    toast.error('Gagal: ' + getErrorMessage(err))
  } finally {
    isAdjusting.value = false
  }
}

const getTodayDateString = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

watch(showRequestModal, (isOpen) => {
  if (isOpen) {
    if (factoryStore.factories && factoryStore.factories.length > 0 && !requestForm.factory_id) {
      requestForm.factory_id = factoryStore.factories[0]?.id || ''
    }
    if (!requestForm.target_shipment_date) {
      requestForm.target_shipment_date = getTodayDateString()
    }
  }
})

watch(showFactoryReturnModal, (isOpen) => {
  if (isOpen && factoryStore.factories && factoryStore.factories.length > 0 && !factoryReturnForm.to_factory_id) {
    factoryReturnForm.to_factory_id = factoryStore.factories[0]?.id || ''
  }
})

const requestForm = reactive({
  factory_id: '',
  notes: '',
  target_shipment_date: getTodayDateString(),
  items: [] as any[],
})

const editRequestForm = reactive({
  id: '',
  factory_id: '',
  notes: '',
  target_shipment_date: getTodayDateString(),
  items: [] as any[],
})

const factoryReturnForm = reactive({
  to_factory_id: '',
  notes: '',
  items: [] as any[],
})

// Methods
const addRequestItem = () => {
  requestForm.items.push({
    product_id: '',
    quantity: 1,
    unit: 'Karton',
    pcs_per_unit: 1,
  })
}

const onProductSelect = (index: number) => {
  const item = requestForm.items[index]
  const product = factoryStore.products.find((p) => p.id === item.product_id)
  if (product) {
    item.unit = product.unit || 'Karton'
    if (isBaseUnit(item.unit)) {
      item.pcs_per_unit = 1
    } else {
      item.pcs_per_unit = product.pcs_per_unit || 1
    }
  }
}

const onEditProductSelect = (index: number) => {
  const item = editRequestForm.items[index]
  const product = factoryStore.products.find((p) => p.id === item.product_id)
  if (product) {
    item.unit = product.unit || 'Karton'
    if (isBaseUnit(item.unit)) {
      item.pcs_per_unit = 1
    } else {
      item.pcs_per_unit = product.pcs_per_unit || 1
    }
  }
}

const removeRequestItem = (index: number) => {
  requestForm.items.splice(index, 1)
}

const addEditRequestItem = () => {
  editRequestForm.items.push({
    id: '',
    product_id: '',
    quantity: 1,
    unit: 'Karton',
    pcs_per_unit: 1,
  })
}

const removeEditRequestItem = (index: number) => {
  editRequestForm.items.splice(index, 1)
}

const openEditRequestModal = (row: any) => {
  selectedEditRequest.value = row
  const firstItem = row.items[0] || {}
  
  editRequestForm.id = firstItem.id || ''
  editRequestForm.factory_id = firstItem.from_factory_id || ''
  editRequestForm.notes = firstItem.notes || ''
  if (firstItem.target_shipment_date) {
    try {
      const d = new Date(firstItem.target_shipment_date)
      const year = d.getFullYear()
      const month = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      editRequestForm.target_shipment_date = `${year}-${month}-${day}`
    } catch {
      editRequestForm.target_shipment_date = getTodayDateString()
    }
  } else {
    editRequestForm.target_shipment_date = getTodayDateString()
  }
  editRequestForm.items = row.items.map((i: any) => ({
    id: i.id,
    product_id: i.product_id,
    quantity: i.quantity,
    unit: i.unit || 'PCS',
    pcs_per_unit: i.pcs_per_unit || 1,
  }))
  
  showEditRequestModal.value = true
}

const getAvailableStocksForReturn = (returnType: string) => {
  if (returnType === 'GOOD') {
    return warehouseStore.inventory.filter((i) => i.stock_type === 'GOOD' && i.quantity > 0)
  }
  // Untuk retur DAMAGED, tampilkan seluruh stok cabang (DAMAGED maupun GOOD) yang tersedia
  return warehouseStore.inventory.filter((i) => i.quantity > 0)
}

const openFactoryReturnModal = () => {
  factoryReturnForm.to_factory_id = ''
  factoryReturnForm.notes = ''
  factoryReturnForm.items = []
  addFactoryReturnItem()
  showFactoryReturnModal.value = true
  if (factoryStore.factories.length === 0) {
    factoryStore.fetchFactories()
  }
}

const addFactoryReturnItem = () => {
  factoryReturnForm.items.push({
    return_type: 'DAMAGED',
    damage_type: 'BOTTLE_PACKAGING',
    damage_reason: '',
    selectedStock: null,
    quantity: 1,
  })
}

const onReturnTypeChange = (index: number) => {
  const item = factoryReturnForm.items[index]
  item.selectedStock = null
  item.quantity = 1
  if (item.return_type === 'GOOD') {
    item.damage_type = 'NONE'
  } else if (!item.damage_type || item.damage_type === 'NONE') {
    item.damage_type = 'BOTTLE_PACKAGING'
  }
}

const onFactoryReturnProductSelect = (index: number) => {
  const item = factoryReturnForm.items[index]
  if (!item.selectedStock) return
  item.quantity = 1
}

const removeFactoryReturnItem = (index: number) => {
  factoryReturnForm.items.splice(index, 1)
}

const submitFactoryReturn = async () => {
  if (!factoryReturnForm.to_factory_id) {
    toast.error('Pabrik tujuan wajib dipilih')
    return
  }
  if (factoryReturnForm.items.length === 0) {
    toast.error('Minimal pilih 1 barang untuk diretur')
    return
  }

  const unselected = factoryReturnForm.items.some((i) => !i.selectedStock)
  if (unselected) {
    toast.error('Pastikan semua barang yang diretur telah dipilih produknya')
    return
  }

  const invalidQty = factoryReturnForm.items.some((i) => !i.quantity || i.quantity <= 0)
  if (invalidQty) {
    toast.error('Jumlah retur harus lebih dari 0')
    return
  }

  const firstBranchId = factoryReturnForm.items[0]?.selectedStock?.branch_id
  const hasMultipleBranches = factoryReturnForm.items.some(
    (i) => i.selectedStock?.branch_id !== firstBranchId,
  )
  if (hasMultipleBranches) {
    toast.error('Semua produk retur harus berasal dari cabang yang sama')
    return
  }

  const payload = {
    from_branch_id: firstBranchId || masterDataStore.selectedWarehouseBranchId,
    to_factory_id: factoryReturnForm.to_factory_id,
    notes: factoryReturnForm.notes,
    items: factoryReturnForm.items.map((i) => ({
      product_id: i.selectedStock.product_id,
      quantity: i.quantity,
      batch_no: i.selectedStock.batch_no,
      return_type: i.return_type || 'DAMAGED',
      damage_type: i.return_type === 'GOOD' ? 'NONE' : (i.damage_type || 'OTHER'),
      damage_reason: i.damage_reason || '',
    })),
  }

  isSubmittingFactoryReturn.value = true
  try {
    await warehouseStore.createFactoryReturn(payload)
    toast.success('Retur barang berhasil dibuat dan dikirim ke pabrik.')
    showFactoryReturnModal.value = false
  } catch (err: any) {
    toast.error('Gagal mengirim retur: ' + getErrorMessage(err))
  } finally {
    isSubmittingFactoryReturn.value = false
  }
}

const refreshAll = async () => {
  if (!masterDataStore.selectedWarehouseBranchId) {
    toast.error('Pilih cabang terlebih dahulu')
    return
  }

  const tid = toast.loading('Memperbarui data gudang...')

  try {
    await Promise.all([
      warehouseStore.fetchInventory(),
      warehouseStore.fetchPendingShipments(),
      warehouseStore.fetchAllTransfers(),
      warehouseStore.fetchFactoryReturns(),
      warehouseStore.fetchLogs(),
    ])

    const count = warehouseStore.pendingShipments.length
    toast.success('Data gudang berhasil diperbarui.', { id: tid })
  } catch (error: any) {
    toast.error(`Gagal mengambil data: ${error.message || 'Unknown Error'}`, { id: tid })
  }
}

const formatStatus = (status: string) => {
  switch (status) {
    case 'SHIPPED':
      return 'Dalam Perjalanan'
    case 'ARRIVED':
      return 'Siap Unload'
    case 'LOADING':
      return 'Sedang Dimuat'
    case 'REQUESTED':
      return 'Menunggu Approval'
    case 'APPROVED':
      return 'Disetujui / Siap Proses'
    default:
      return status
  }
}

// Computed for Filters & Groups
const filteredInventory = computed(() => {
  let list = warehouseStore.inventory.filter((i) => i.quantity > 0 || (i.min_limit || 0) > 0)
  if (stockTypeFilter.value !== 'ALL') {
    list = list.filter((i) => (i.stock_type || 'GOOD') === stockTypeFilter.value)
  }
  if (stockSearch.value) {
    const q = stockSearch.value.toLowerCase()
    list = list.filter(
      (i) => i.product?.name?.toLowerCase().includes(q) || i.batch_no?.toLowerCase().includes(q),
    )
  }

  // Group by product_id
  const groups: Record<string, any[]> = {}
  list.forEach((item) => {
    const pid = item.product_id || item.product?.id || 'unknown'
    if (!groups[pid]) {
      groups[pid] = []
    }
    groups[pid].push(item)
  })

  const flatList: any[] = []
  Object.entries(groups).forEach(([pid, children]) => {
    const defaultBatch = children.find((c) => c.batch_no === 'DEFAULT') || children[0]
    const totalQty = children.reduce((sum, c) => sum + (c.quantity || 0), 0)

    const latestUpdate = children.reduce((max, c) => {
      if (!c.updated_at) return max
      if (!max) return c.updated_at
      return new Date(c.updated_at) > new Date(max) ? c.updated_at : max
    }, '')

    const maxMinLimit = children.reduce((max, c) => Math.max(max, c.min_limit || 0), 0)
    const effectiveBranchId =
      (masterDataStore.selectedWarehouseBranchId !== 'ALL_BRANCHES' && masterDataStore.selectedWarehouseBranchId) ||
      defaultBatch.branch_id ||
      ''

    const parentRow = {
      isParent: true,
      isChild: false,
      product_id: pid,
      product_name: defaultBatch.product?.name || 'Produk Tidak Dikenal',
      product: defaultBatch.product,
      quantity: totalQty,
      min_limit: maxMinLimit || defaultBatch.min_limit || 0,
      batch_no: defaultBatch.batch_no,
      branch_id: effectiveBranchId,
      updated_at: latestUpdate,
      childCount: children.length,
      stock_id: children.length === 1 ? defaultBatch.id : undefined,
      id: `parent-${pid}`,
    }

    flatList.push(parentRow)

    const isExpanded = expandedProductIds.value.has(pid) || !!stockSearch.value
    if (isExpanded) {
      children.forEach((c, index) => {
        flatList.push({
          ...c,
          stock_id: c.id,
          isParent: false,
          isChild: true,
          product_name: c.product?.name,
          id: `child-${c.id || index}-${pid}`,
        })
      })
    }
  })

  return flatList
})


const groupedPending = computed(() => {
  const groups: Record<string, any> = {}
  warehouseStore.pendingShipments
    .filter((item) => item.to_branch_id === masterDataStore.selectedWarehouseBranchId)
    .forEach((item) => {
    const doNo = item.delivery_order_no || 'REQ-' + item.id.substring(0, 8)
    // Status berlaku per produk. Satu SJ bisa berisi produk yang sudah dikirim
    // dan produk yang masih APPROVED, sehingga aksinya harus dipisahkan.
    const groupKey = `${doNo}::${item.status}`
    if (!groups[groupKey]) {
      groups[groupKey] = {
        doNo,
        from: item.fulfillment_branch?.name || item.from_factory?.name || 'Cabang Lain',
        to: item.to_branch?.name || currentBranchName.value,
        date: item.created_at,
        status: item.status,
        totalQty: 0,
        totalWeight: 0,
        items: [],
        vehicle: null,
        driver: null,
        targetDate: item.target_shipment_date,
        eta: item.estimated_arrival,
        shippedAt: item.shipped_at,
        // Flag delegasi: jika fulfillment_branch_id ada, maka ini diproses oleh cabang lain
        isFulfillmentDelegation: !!item.fulfillment_branch_id,
        fulfillmentBranchName: item.fulfillment_branch?.name || '',
      }
    }
    groups[groupKey].items.push(item)
    groups[groupKey].totalQty += item.quantity
    groups[groupKey].totalWeight += item.total_weight
  })

  for (const group of Object.values(groups)) {
    group.vehicle = group.items.find((i: any) => i.vehicle)?.vehicle || null
    group.driver = group.items.find((i: any) => i.driver)?.driver || null
    if (!group.shippedAt) {
      group.shippedAt = group.items.find((i: any) => i.shipped_at)?.shipped_at
    }
  }

  return Object.values(groups).sort(
    (a, b) => new Date(b.date).getTime() - new Date(a.date).getTime(),
  )
})



const groupedHistory = computed(() => {
  const groups: Record<string, any> = {}
  warehouseStore.allTransfers
    .filter((s) => s.status === 'RECEIVED')
    .forEach((item) => {
      const doNo = item.delivery_order_no || 'DO-' + item.id.substring(0, 8)
      if (!groups[doNo]) {
        groups[doNo] = {
          doNo,
          from: item.fulfillment_branch?.name || item.from_factory?.name || 'Cabang Lain',
          to: item.to_branch?.name || 'Cabang Lain',
          receivedAt: item.received_at || item.created_at,
          totalWeight: 0,
          items: [],
          notes: '',
        }
      }
      groups[doNo].items.push(item)
      groups[doNo].totalWeight += item.total_weight
      if (item.notes && !groups[doNo].notes.includes(item.notes)) {
        if (groups[doNo].notes) {
          groups[doNo].notes += ' | ' + item.notes
        } else {
          groups[doNo].notes = item.notes
        }
      }
    })
  return Object.values(groups).sort(
    (a, b) => new Date(b.receivedAt).getTime() - new Date(a.receivedAt).getTime(),
  )
})

const filteredHistory = computed(() => {
  if (!historySearch.value) return groupedHistory.value
  const q = historySearch.value.toLowerCase()
  return groupedHistory.value.filter(
    (s) => s.doNo.toLowerCase().includes(q) || s.from.toLowerCase().includes(q),
  )
})

// Handlers
const handleArrive = async (group: any) => {
  processingId.value = group.doNo
  try {
    for (const item of group.items) {
      await warehouseStore.arriveShipment(item.id)
    }
    toast.success('Kedatangan dikonfirmasi. Silahkan lakukan bongkar muat.')
    refreshAll()
  } catch (err) {
    toast.error('Gagal: ' + getErrorMessage(err))
  } finally {
    processingId.value = ''
  }
}

const handleReceive = (group: any) => {
  receivingShipment.value = group
  receiveForm.items = group.items.map((item: any) => ({
    id: item.id,
    product_name: item.product?.name,
    unit: item.unit || item.product?.unit || 'UNIT',
    expected_quantity: item.quantity,
    actual_received_quantity_units: item.quantity,
    actual_received_quantity_pcs: 0,
    damaged_quantity_units: 0,
    damaged_quantity_pcs: 0,
    pcs_per_unit: item.pcs_per_unit || 1,
  }))
  receiveForm.notes = ''
  showReceiveModal.value = true
}

const isSelisih = (item: any) => {
  const totalReceivedPcs =
    item.actual_received_quantity_units * item.pcs_per_unit + item.actual_received_quantity_pcs
  const totalDamagedPcs =
    item.damaged_quantity_units * item.pcs_per_unit + item.damaged_quantity_pcs
  const expectedPcs = item.expected_quantity * item.pcs_per_unit
  return totalReceivedPcs + totalDamagedPcs !== expectedPcs
}

const getSelisihQty = (item: any) => {
  const totalReceivedPcs =
    item.actual_received_quantity_units * item.pcs_per_unit + item.actual_received_quantity_pcs
  const totalDamagedPcs =
    item.damaged_quantity_units * item.pcs_per_unit + item.damaged_quantity_pcs
  const expectedPcs = item.expected_quantity * item.pcs_per_unit
  return (totalReceivedPcs + totalDamagedPcs - expectedPcs) / item.pcs_per_unit
}

const submitReceive = async () => {
  isSubmittingReceive.value = true
  try {
    for (const item of receiveForm.items) {
      const totalActualPcs =
        item.actual_received_quantity_units * item.pcs_per_unit + item.actual_received_quantity_pcs
      const totalDamagedPcs =
        item.damaged_quantity_units * item.pcs_per_unit + item.damaged_quantity_pcs

      await warehouseStore.receiveShipment(item.id, {
        actual_received_quantity: totalActualPcs,
        damaged_quantity: totalDamagedPcs,
        notes: receiveForm.notes,
      })
    }
    showReceiveModal.value = false
    toast.success('Penerimaan berhasil dikonfirmasi dengan penyesuaian kondisi barang.')
    refreshAll()
  } catch (err) {
    toast.error('Gagal: ' + getErrorMessage(err))
  } finally {
    isSubmittingReceive.value = false
  }
}

const handleApprove = async (group: any) => {
  try {
    for (const item of group.items) await warehouseStore.approveShipment(item.id)
    toast.success('Disetujui.')
    refreshAll()
  } catch (err) {
    toast.error('Gagal: ' + getErrorMessage(err))
  }
}

const handleReject = (group: any) => {
  rejectionForm.group = group
  rejectionForm.reason = ''
  showRejectModal.value = true
}

const submitReject = async () => {
  if (!rejectionForm.reason) {
    toast.error('Alasan penolakan wajib diisi')
    return
  }

  isRejecting.value = true
  try {
    for (const item of rejectionForm.group.items) {
      await warehouseStore.rejectShipment(item.id, rejectionForm.reason)
    }
    toast.success('Pengiriman ditolak.')
    showRejectModal.value = false
    refreshAll()
  } catch (err) {
    toast.error('Gagal menolak: ' + getErrorMessage(err))
  } finally {
    isRejecting.value = false
  }
}

const openEditThreshold = (row: any) => {
  selectedThresholdItem.value = row
  thresholdForm.limit = row.min_limit || 0
  showThresholdModal.value = true
}

const submitThreshold = async () => {
  if (!selectedThresholdItem.value) return
  isSavingThreshold.value = true
  try {
    const newLimit = Number(thresholdForm.limit) || 0
    const targetBranchId =
      (masterDataStore.selectedWarehouseBranchId !== 'ALL_BRANCHES' && masterDataStore.selectedWarehouseBranchId) ||
      selectedThresholdItem.value.branch_id ||
      ''

    await warehouseStore.setStockLimit({
      branch_id: targetBranchId,
      product_id: selectedThresholdItem.value.product_id,
      batch_no: selectedThresholdItem.value.batch_no || '',
      limit: newLimit,
    })
    selectedThresholdItem.value.min_limit = newLimit
    showThresholdModal.value = false
    toast.success('Threshold limit berhasil diperbarui.')
    await refreshAll()
  } catch (err) {
    toast.error('Gagal memperbarui threshold: ' + getErrorMessage(err))
  } finally {
    isSavingThreshold.value = false
  }
}

const handleConfirmSJ = async (group: any) => {
  processingId.value = group.doNo
  try {
    for (const item of group.items) {
      await warehouseStore.confirmShipment(item.id)
    }
    toast.success('Surat Jalan telah disetujui. Barang dalam perjalanan.')
    refreshAll()
  } catch (err) {
    toast.error('Gagal menyetujui SJ: ' + getErrorMessage(err))
  } finally {
    processingId.value = ''
  }
}

const handleScanSubmit = async () => {
  if (!scannedDO.value) return
  isScanning.value = true
  try {
    // For Scan SJ, we fetch the DO first to allow partial receiving,
    // but for simplicity we'll assume full receive or redirect to a modal if needed.
    // For now, let's just do full receive but we could improve this later.
    await warehouseStore.receiveByDO(scannedDO.value, {
      actual_received_quantity: 0, // This needs to be handled better in backend if 0
      damaged_quantity: 0,
    })
    showScanner.value = false
    scannedDO.value = ''
    toast.success('Barang diterima via Scan SJ.')
    refreshAll()
  } catch (err) {
    toast.error('Gagal: ' + getErrorMessage(err))
  } finally {
    isScanning.value = false
  }
}

const handleRequest = async () => {
  if (!requestForm.factory_id) return

  const validItems = requestForm.items.filter((i) => i.product_id && i.quantity > 0)
  if (validItems.length === 0) {
    toast.error('Pilih setidaknya satu produk dengan jumlah yang benar')
    return
  }

  function openTransferRequestModal() {
    if (!masterDataStore.selectedWarehouseBranchId || masterDataStore.selectedWarehouseBranchId === 'ALL_BRANCHES') {
      toast.error('Pilih cabang spesifik terlebih dahulu untuk meminta barang')
      return
    }
  }

  isRequesting.value = true
  try {
    await warehouseStore.requestGoods(requestForm.factory_id, {
      to_branch_id: masterDataStore.selectedWarehouseBranchId,
      items: validItems,
      notes: requestForm.notes,
      target_shipment_date: requestForm.target_shipment_date || undefined,
      initiated_by: 'BRANCH',
    })
    showRequestModal.value = false
    // Reset form
    requestForm.factory_id = ''
    requestForm.notes = ''
    requestForm.target_shipment_date = getTodayDateString()
    requestForm.items = []

    refreshAll()
    toast.success('Permintaan stok terkirim.')
  } catch (err) {
    toast.error('Gagal: ' + getErrorMessage(err))
  } finally {
    isRequesting.value = false
  }
}

const handleUpdateRequest = async () => {
  if (!editRequestForm.factory_id) {
    toast.error('Pilih pabrik tujuan.')
    return
  }

  const validItems = editRequestForm.items.filter((i) => i.product_id && i.quantity > 0)
  if (validItems.length === 0) {
    toast.error('Pilih setidaknya satu produk dengan jumlah yang benar.')
    return
  }

  isRequesting.value = true
  try {
    const currentItemIds = validItems.filter((i) => i.id).map((i) => i.id)
    const removedItems = selectedEditRequest.value.items.filter(
      (i: any) => !currentItemIds.includes(i.id),
    )

    // Delete removed items
    for (const item of removedItems) {
      await factoryStore.deleteTransfer(item.id)
    }

    // Update existing or create new items
    for (const item of validItems) {
      const payload = {
        product_id: item.product_id,
        quantity: item.quantity,
        unit: item.unit,
        pcs_per_unit: item.pcs_per_unit || 1,
        notes: editRequestForm.notes,
        target_shipment_date: editRequestForm.target_shipment_date || undefined,
        from_factory_id: editRequestForm.factory_id,
        to_branch_id: masterDataStore.selectedWarehouseBranchId,
        status: 'REQUESTED',
        initiated_by: 'BRANCH',
        delivery_order_no: selectedEditRequest.value.doNo,
      }

      if (item.id) {
        await factoryStore.updateTransfer(item.id, payload)
      } else {
        await factoryStore.createTransfer(payload)
      }
    }

    showEditRequestModal.value = false
    refreshAll()
    toast.success('Request berhasil diperbarui.')
  } catch (err) {
    toast.error('Gagal memperbarui request: ' + getErrorMessage(err))
  } finally {
    isRequesting.value = false
  }
}

const handleDeletePending = async (group: any) => {
  if (!group || !group.items || group.items.length === 0) return

  if (group.status !== 'REQUESTED' && group.status !== 'APPROVED') {
    toast.error('Hanya pengiriman/permintaan berstatus Pending/Belum Dikirim yang dapat dibatalkan atau dihapus.')
    return
  }

  const confirmMessage = `Apakah Anda yakin ingin membatalkan dan menghapus data pengiriman/permintaan ${group.doNo}? Tindakan ini permanen dan tidak dapat dibatalkan.`
  if (!confirm(confirmMessage)) return

  deletingDoNo.value = group.doNo
  isDeletingShipment.value = true
  try {
    for (const item of group.items) {
      if (item.id) {
        await factoryStore.deleteTransfer(item.id)
      }
    }
    toast.success(`Data ${group.doNo} berhasil dibatalkan dan dihapus.`)
    refreshAll()
  } catch (err) {
    toast.error('Gagal menghapus data: ' + getErrorMessage(err))
  } finally {
    deletingDoNo.value = ''
    isDeletingShipment.value = false
  }
}

const handleDeleteFromEditModal = async () => {
  if (!selectedEditRequest.value || !selectedEditRequest.value.items) return

  const confirmMessage = `Apakah Anda yakin ingin membatalkan dan menghapus seluruh permintaan ${selectedEditRequest.value.doNo}? Tindakan ini permanen.`
  if (!confirm(confirmMessage)) return

  isRequesting.value = true
  try {
    for (const item of selectedEditRequest.value.items) {
      if (item.id) {
        await factoryStore.deleteTransfer(item.id)
      }
    }
    showEditRequestModal.value = false
    toast.success('Permintaan berhasil dibatalkan dan dihapus.')
    refreshAll()
  } catch (err) {
    toast.error('Gagal menghapus permintaan: ' + getErrorMessage(err))
  } finally {
    isRequesting.value = false
  }
}

const openAdjustmentFor = (row: any) => {
  const goodItem = warehouseStore.inventory.find(
    (i) =>
      i.product_id === row.product_id &&
      i.batch_no === row.batch_no &&
      (i.stock_type === 'GOOD' || !i.stock_type)
  )
  const damagedItem = warehouseStore.inventory.find(
    (i) =>
      i.product_id === row.product_id &&
      i.batch_no === row.batch_no &&
      i.stock_type === 'DAMAGED'
  )
  selectedStockForAdjustment.value = goodItem || damagedItem || row
  showAdjustment.value = true
}

// Delete Stock Handlers
const showDeleteStockModal = ref(false)
const selectedStockToDelete = ref<any>(null)
const deleteStockReason = ref('')
const deleteStockConfirmationText = ref('')
const isDeletingStock = ref(false)

const openDeleteStockModal = (row: any) => {
  const stockId = row.stock_id || (row.id && !String(row.id).startsWith('parent-') && !String(row.id).startsWith('child-') ? row.id : null)
  let foundStock = null
  if (stockId) {
    foundStock = warehouseStore.inventory.find((i) => i.id === stockId)
  }
  if (!foundStock) {
    foundStock = warehouseStore.inventory.find(
      (i) => i.product_id === row.product_id && i.batch_no === row.batch_no && (i.stock_type === row.stock_type || !row.stock_type || !i.stock_type)
    )
  }
  selectedStockToDelete.value = foundStock || row
  deleteStockReason.value = ''
  deleteStockConfirmationText.value = ''
  showDeleteStockModal.value = true
}

const closeDeleteStockModal = () => {
  showDeleteStockModal.value = false
  selectedStockToDelete.value = null
  deleteStockReason.value = ''
  deleteStockConfirmationText.value = ''
}

const handleConfirmDeleteStock = async () => {
  if (!selectedStockToDelete.value?.id) {
    toast.error('ID data stok tidak ditemukan.')
    return
  }
  if (deleteStockConfirmationText.value.trim().toUpperCase() !== 'HAPUS') {
    toast.error('Ketik kata "HAPUS" untuk mengonfirmasi.')
    return
  }
  if (!deleteStockReason.value.trim()) {
    toast.error('Alasan penghapusan wajib diisi.')
    return
  }

  isDeletingStock.value = true
  try {
    await warehouseStore.deleteStock(
      selectedStockToDelete.value.id,
      deleteStockReason.value.trim()
    )
    toast.success('Data stok berhasil dihapus permanen.')
    closeDeleteStockModal()
    refreshAll()
  } catch (err: any) {
    toast.error(getErrorMessage(err))
  } finally {
    isDeletingStock.value = false
  }
}

// Table Columns
const stockColumns = [
  { key: 'product_name', label: 'Produk', class: 'font-black text-slate-700' },
  {
    key: 'batch_no',
    label: 'Batch No & Tipe',
    render: (val: string, item: any) => {
      let badge = ''
      if (item.stock_type === 'DAMAGED') {
        badge =
          '<span class="bg-rose-100 text-rose-700 px-1.5 py-0.5 rounded font-bold uppercase text-[9px] mt-1 inline-block">Rusak</span>'
      } else if (item.stock_type === 'QUARANTINE') {
        badge =
          '<span class="bg-amber-100 text-amber-700 px-1.5 py-0.5 rounded font-bold uppercase text-[9px] mt-1 inline-block">Karantina</span>'
      } else {
        badge =
          '<span class="bg-emerald-100 text-emerald-700 px-1.5 py-0.5 rounded font-bold uppercase text-[9px] mt-1 inline-block">Bagus</span>'
      }
      return `<div class="flex flex-col items-start"><span class="font-mono text-[11px] text-slate-500">${val}</span>${badge}</div>`
    },
  },
  {
    key: 'quantity',
    label: 'Stok Gudang (Fisik)',
    render: (val: number, item: any) => {
      if (item.batch_no === 'BACKORDER') {
        return `<span class="text-[10px] font-black text-rose-500 bg-rose-50 px-2 py-1 rounded-md border border-rose-100 uppercase tracking-widest">Antrean Kosong</span>`
      }
      const isLow = val < (item?.min_limit || 0)
      return `<div class="flex flex-col">
        <div class="flex items-center gap-2">
          <span class="text-lg font-black ${isLow ? 'text-rose-600' : 'text-slate-900'}">${formatStock(val, item.product)}</span>
          ${isLow ? '<span class="bg-rose-50 text-rose-600 text-[8px] font-black px-2 py-0.5 rounded-md border border-rose-100 uppercase animate-pulse">Low</span>' : ''}
        </div>
        <span class="text-[9px] font-bold text-slate-400 uppercase tracking-tighter">${val} PCS</span>
      </div>`
    },
  },
  {
    key: 'min_limit',
    label: 'Batas Minimum Stok',
  },
  {
    key: 'updated_at',
    label: 'Update Terakhir',
    render: (v: string) => {
      if (!v) return '-'
      const date = new Date(v)
      return `<span class="text-[10px] font-bold text-slate-400 uppercase">${date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })}</span>`
    },
  },
  {
    key: 'actions',
    label: '',
  },
]

const inboundColumns = [
  { key: 'doNo', label: 'Surat Jalan' },
  { key: 'from', label: 'Asal Pengiriman' },
  { key: 'status', label: 'Status' },
  { key: 'totalQty', label: 'Item' },
  { key: 'totalWeight', label: 'Berat' },
  { key: 'targetDate', label: 'Jadwal / ETA' },
  { key: 'actions', label: 'Aksi' },
]

const historyColumns = [
  { key: 'doNo', label: 'Surat Jalan' },
  { key: 'from', label: 'Asal' },
  { key: 'receivedAt', label: 'Diterima Pada' },
  { key: 'totalWeight', label: 'Total Muatan' },
  { key: 'items', label: 'Produk' },
  { key: 'notes', label: 'Catatan / Komplain' },
  { key: 'actions', label: 'Aksi' },
]

// Refresh data when branch changes
watch(
  () => masterDataStore.selectedWarehouseBranchId,
  () => {
    refreshAll()
  },
)

onMounted(() => {
  masterDataStore.fetchBranches()
  masterDataStore.fetchCompanies()
  masterDataStore.fetchEmployees()
  refreshAll()
  factoryStore.fetchFactories()
  factoryStore.fetchProducts()
})
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
</style>
