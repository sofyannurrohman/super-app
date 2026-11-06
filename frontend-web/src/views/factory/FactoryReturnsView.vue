<template>
  <div class="p-8 space-y-6 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-6 bg-white p-8 rounded-[2.5rem] shadow-sm border border-slate-100"
    >
      <div class="flex items-center gap-5">
        <div class="w-14 h-14 rounded-2xl bg-rose-50 flex items-center justify-center text-rose-600 shadow-sm border border-rose-100">
          <RotateCcw class="h-7 w-7" />
        </div>
        <div>
          <h1 class="text-3xl font-black tracking-tight text-slate-900">Barang Returan Cabang</h1>
          <p class="text-sm text-slate-500 font-medium mt-1">
            Penerimaan, inspeksi QC (Barang Bagus, Rusak Botol/Repack, Rusak Isi/Scrap), dan pembaruan stok pabrik.
          </p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <Button
          variant="outline"
          class="rounded-2xl h-12 px-6 font-bold border-slate-200"
          @click="refreshAll"
          :loading="loading"
        >
          <RefreshCw class="h-4 w-4 mr-2" :class="{ 'animate-spin': loading }" /> Refresh Data
        </Button>
      </div>
    </div>

    <!-- Metrics Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
      <Card class="border-none shadow-sm rounded-[2rem] bg-white p-6 border border-slate-100">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-[11px] font-black uppercase tracking-wider text-slate-400">Total Retur Masuk</p>
            <h3 class="text-3xl font-black text-slate-900 mt-2">{{ totalReturnsCount }}</h3>
            <p class="text-xs font-semibold text-slate-500 mt-1">Semua pengiriman retur</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-slate-50 flex items-center justify-center text-slate-600">
            <Package class="h-6 w-6" />
          </div>
        </div>
      </Card>

      <Card class="border-none shadow-sm rounded-[2rem] bg-amber-500/10 p-6 border border-amber-200/50">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-[11px] font-black uppercase tracking-wider text-amber-700">Menunggu QC / Tindakan</p>
            <h3 class="text-3xl font-black text-amber-800 mt-2">{{ pendingReturns.length }}</h3>
            <p class="text-xs font-semibold text-amber-700 mt-1">Perlu diinspeksi & dikonfirmasi</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-amber-500/20 flex items-center justify-center text-amber-700">
            <AlertCircle class="h-6 w-6" />
          </div>
        </div>
      </Card>

      <Card class="border-none shadow-sm rounded-[2rem] bg-emerald-500/10 p-6 border border-emerald-200/50">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-[11px] font-black uppercase tracking-wider text-emerald-700">Selesai Diinspeksi</p>
            <h3 class="text-3xl font-black text-emerald-800 mt-2">{{ historyReturns.length }}</h3>
            <p class="text-xs font-semibold text-emerald-700 mt-1">Stok telah diperbarui</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-emerald-500/20 flex items-center justify-center text-emerald-700">
            <CheckCircle2 class="h-6 w-6" />
          </div>
        </div>
      </Card>

      <Card class="border-none shadow-sm rounded-[2rem] bg-blue-500/10 p-6 border border-blue-200/50">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-[11px] font-black uppercase tracking-wider text-blue-700">Alokasi Hasil QC</p>
            <div class="mt-2 flex items-center gap-2">
              <span class="text-xs font-black bg-emerald-100 text-emerald-800 px-2 py-0.5 rounded-md" title="Masuk Stok Jadi">
                {{ totalGoodRestocked }} Bagus
              </span>
              <span class="text-xs font-black bg-orange-100 text-orange-800 px-2 py-0.5 rounded-md" title="Masuk Repack">
                {{ totalRepackBotol }} Repack
              </span>
              <span class="text-xs font-black bg-rose-100 text-rose-800 px-2 py-0.5 rounded-md" title="Scrap / Pemusnahan">
                {{ totalScrapLimbah }} Scrap
              </span>
            </div>
            <p class="text-[11px] font-semibold text-blue-600 mt-1.5">Total unit per jenis disposisi</p>
          </div>
          <div class="w-12 h-12 rounded-2xl bg-blue-500/20 flex items-center justify-center text-blue-700">
            <SlidersHorizontal class="h-6 w-6" />
          </div>
        </div>
      </Card>
    </div>

    <!-- Navigation Tabs & Search -->
    <div class="flex flex-col sm:flex-row justify-between items-stretch sm:items-center gap-4 bg-white p-3 rounded-[2rem] shadow-sm border border-slate-100">
      <div class="flex items-center gap-2">
        <Button
          :variant="activeTab === 'pending' ? 'default' : 'ghost'"
          class="rounded-xl h-11 px-5 font-black text-xs gap-2"
          :class="activeTab === 'pending' ? 'bg-slate-900 text-white shadow-md' : 'text-slate-600'"
          @click="activeTab = 'pending'"
        >
          <Clock class="h-4 w-4" />
          Menunggu Tindakan QC
          <Badge v-if="pendingReturns.length > 0" class="bg-amber-400 text-slate-900 font-black px-1.5 py-0 text-[10px] ml-1">
            {{ pendingReturns.length }}
          </Badge>
        </Button>
        <Button
          :variant="activeTab === 'history' ? 'default' : 'ghost'"
          class="rounded-xl h-11 px-5 font-black text-xs gap-2"
          :class="activeTab === 'history' ? 'bg-slate-900 text-white shadow-md' : 'text-slate-600'"
          @click="activeTab = 'history'"
        >
          <CheckCircle2 class="h-4 w-4" />
          Riwayat & Berita Acara QC
        </Button>
        <Button
          :variant="activeTab === 'sales_returns' ? 'default' : 'ghost'"
          class="rounded-xl h-11 px-5 font-black text-xs gap-2"
          :class="activeTab === 'sales_returns' ? 'bg-slate-900 text-white shadow-md' : 'text-slate-600'"
          @click="activeTab = 'sales_returns'"
        >
          <RotateCcw class="h-4 w-4" />
          Retur Penjualan & Sampingan
          <Badge v-if="filteredSalesReturns.length > 0" class="bg-blue-600 text-white font-black px-1.5 py-0 text-[10px] ml-1">
            {{ filteredSalesReturns.length }}
          </Badge>
        </Button>
      </div>

      <div class="flex items-center gap-3">
        <div class="relative w-full sm:w-64">
          <Search class="h-4 w-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            v-model="search"
            type="text"
            placeholder="Cari No. Retur, Cabang..."
            class="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold outline-none focus:ring-2 focus:ring-slate-900/10"
          />
        </div>
      </div>
    </div>

    <!-- TAB 1: PENDING QC & PENERIMAAN -->
    <div v-if="activeTab === 'pending'" class="space-y-6">
      <div v-if="pendingReturns.length === 0" class="bg-white rounded-[2.5rem] p-16 text-center border-2 border-dashed border-slate-200">
        <div class="w-16 h-16 bg-emerald-50 rounded-2xl flex items-center justify-center mx-auto mb-4 border border-emerald-100">
          <CheckCircle2 class="h-8 w-8 text-emerald-500" />
        </div>
        <h3 class="text-lg font-black text-slate-800">Semua Retur Telah Diinspeksi!</h3>
        <p class="text-sm text-slate-400 mt-1 max-w-md mx-auto">
          Tidak ada pengiriman retur dari cabang yang sedang menunggu konfirmasi QC saat ini.
        </p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        <Card
          v-for="rtn in filteredPending"
          :key="rtn.id"
          class="border-none shadow-lg shadow-slate-200/40 bg-white rounded-[2.5rem] overflow-hidden border border-slate-100 flex flex-col justify-between transition-all hover:shadow-xl hover:-translate-y-0.5"
        >
          <div class="p-7 space-y-5">
            <!-- Header Card -->
            <div class="flex justify-between items-start gap-3">
              <div>
                <div class="text-[10px] font-black uppercase tracking-widest text-slate-400 flex items-center gap-1.5 mb-1">
                  <Store class="h-3.5 w-3.5 text-slate-400" />
                  {{ rtn.from_branch?.name || 'Cabang' }}
                </div>
                <h3 class="text-base font-black text-slate-900">
                  {{ rtn.return_no || 'RTF-' + rtn.id.substring(0, 8) }}
                </h3>
              </div>
              <Badge class="bg-orange-100 text-orange-700 border-none font-black text-[10px] px-2.5 py-1 rounded-lg uppercase tracking-wider">
                {{ rtn.status }}
              </Badge>
            </div>

            <!-- Shipment Info -->
            <div class="grid grid-cols-2 gap-2 bg-slate-50 p-3.5 rounded-2xl text-[11px] font-bold border border-slate-100">
              <div>
                <span class="text-slate-400 text-[9px] uppercase tracking-wider block">Tgl Kirim Cabang</span>
                <span class="text-slate-700">{{ formatDate(rtn.shipped_at || rtn.created_at) }}</span>
              </div>
              <div>
                <span class="text-slate-400 text-[9px] uppercase tracking-wider block">Armada / Supir</span>
                <span class="text-slate-700 truncate block">
                  {{ rtn.driver?.first_name ? `${rtn.driver?.first_name} ${rtn.driver?.last_name || ''}` : (rtn.vehicle?.plate_number || 'Ekspedisi/Lainnya') }}
                </span>
              </div>
            </div>

            <!-- Items Breakdown -->
            <div class="space-y-2.5">
              <p class="text-[10px] font-black uppercase tracking-widest text-slate-400">Rincian Barang yang Dikirim:</p>
              <div class="space-y-2 max-h-56 overflow-y-auto custom-scrollbar pr-1">
                <div
                  v-for="item in rtn.items"
                  :key="item.id"
                  class="p-3 bg-slate-50/80 rounded-2xl border border-slate-100 space-y-1.5"
                >
                  <div class="flex justify-between items-start text-xs font-bold gap-2">
                    <span class="text-slate-800 line-clamp-1">{{ item.product?.name }}</span>
                    <span class="text-slate-900 font-black whitespace-nowrap">
                      {{ item.quantity }} {{ item.unit || item.product?.unit || 'UNIT' }}
                    </span>
                  </div>

                  <div class="flex flex-wrap items-center gap-1.5">
                    <!-- Return Type Badge -->
                    <span
                      class="text-[9px] font-black px-2 py-0.5 rounded-md uppercase tracking-wider"
                      :class="getReturnTypeBadgeClass(item.return_type)"
                    >
                      {{ item.return_type === 'GOOD' ? '🟢 Barang Bagus' : '🔴 Barang Rusak' }}
                    </span>

                    <!-- Damage Subtype Badge -->
                    <span
                      v-if="item.return_type !== 'GOOD'"
                      class="text-[9px] font-bold px-2 py-0.5 rounded-md"
                      :class="getDamageTypeBadgeClass(item.damage_type)"
                    >
                      {{ getDamageTypeLabel(item.damage_type) }}
                    </span>

                    <span class="text-[9px] text-slate-400 font-semibold">
                      Batch: {{ item.batch_no || 'DEFAULT' }}
                    </span>
                  </div>

                  <div v-if="item.damage_reason" class="text-[10px] text-slate-500 italic bg-white/70 p-1.5 rounded-lg border border-slate-100">
                    "{{ item.damage_reason }}"
                  </div>
                </div>
              </div>
            </div>

            <!-- Return Notes -->
            <div v-if="rtn.notes" class="text-[11px] text-slate-500 bg-amber-50/60 p-3 rounded-2xl border border-amber-100 italic">
              "{{ rtn.notes }}"
            </div>
          </div>

          <!-- Card Footer Action -->
          <div class="p-6 pt-0 grid grid-cols-1 sm:grid-cols-2 gap-2">
            <Button
              @click="openRejectModal(rtn)"
              variant="outline"
              class="w-full rounded-2xl h-12 font-black border-rose-200 text-rose-700 hover:bg-rose-50 text-xs gap-1.5"
            >
              <XCircle class="h-4 w-4 text-rose-600" /> TOLAK RETUR
            </Button>
            <Button
              @click="openInspectModal(rtn)"
              class="w-full rounded-2xl h-12 font-black shadow-lg shadow-emerald-500/10 bg-emerald-600 hover:bg-emerald-700 text-white text-xs gap-1.5"
            >
              <ClipboardCheck class="h-4 w-4" /> SETUJUI & TERIMA
            </Button>
          </div>
        </Card>
      </div>
    </div>

    <!-- TAB 2: RIWAYAT & BERITA ACARA QC -->
    <div v-if="activeTab === 'history'" class="space-y-6">
      <Card class="border-none shadow-xl shadow-slate-200/50 bg-white rounded-[2.5rem] overflow-hidden border border-slate-100">
        <DataTable :columns="historyColumns" :data="filteredHistory">
          <template #cell-return_no="{ row }">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-emerald-50 flex items-center justify-center text-emerald-600 shadow-sm border border-emerald-100">
                <CheckCircle2 class="h-5 w-5" />
              </div>
              <div>
                <span class="font-black text-slate-900 block text-sm">{{ row.return_no || 'RTF-' + row.id.substring(0, 8) }}</span>
                <span class="text-[10px] font-bold text-slate-400 flex items-center gap-1 mt-0.5">
                  <Store class="h-3 w-3" /> {{ row.from_branch?.name || 'Cabang' }}
                </span>
              </div>
            </div>
          </template>

          <template #cell-received_at="{ row }">
            <div v-if="row.received_at" class="flex flex-col text-[11px] font-bold text-slate-600">
              <span class="flex items-center gap-1 text-slate-800">
                <Calendar class="h-3.5 w-3.5 text-slate-400" />
                {{ formatDate(row.received_at) }}
              </span>
              <span class="text-[10px] text-slate-400 mt-0.5">
                Oleh: <b class="text-slate-600">{{ row.receiver_name || row.received_by || 'Admin Pabrik' }}</b>
              </span>
            </div>
            <span v-else class="text-xs text-slate-400 font-bold">-</span>
          </template>

          <template #cell-items_breakdown="{ row }">
            <div class="space-y-1.5 py-1">
              <div
                v-for="item in row.items.slice(0, 2)"
                :key="item.id"
                class="text-[11px] bg-slate-50 px-3 py-2 rounded-xl border border-slate-100 flex flex-col gap-1"
              >
                <div class="flex justify-between items-center font-bold">
                  <span class="text-slate-800 truncate max-w-[200px]">{{ item.product?.name }}</span>
                  <span class="text-slate-900 font-black">
                    {{ item.received_quantity || item.quantity }} {{ item.unit || 'UNIT' }}
                  </span>
                </div>
                <div class="flex items-center gap-1.5 text-[9px] font-bold">
                  <span v-if="(item.received_good_qty || 0) > 0" class="text-emerald-700 bg-emerald-100 px-1.5 py-0.2 rounded">
                    +{{ item.received_good_qty }} Bagus (Restock)
                  </span>
                  <span v-if="(item.received_bottle_damaged_qty || 0) > 0" class="text-orange-700 bg-orange-100 px-1.5 py-0.2 rounded">
                    +{{ item.received_bottle_damaged_qty }} Repack
                  </span>
                  <span v-if="(item.received_content_damaged_qty || 0) > 0" class="text-rose-700 bg-rose-100 px-1.5 py-0.2 rounded">
                    +{{ item.received_content_damaged_qty }} Scrap
                  </span>
                </div>
              </div>
              <div v-if="row.items.length > 2" class="text-[10px] font-bold text-slate-400 text-center py-1 bg-slate-50 rounded-xl border border-dashed border-slate-200">
                + {{ row.items.length - 2 }} item lainnya
              </div>
            </div>
          </template>

          <template #cell-actions="{ row }">
            <Button
              variant="outline"
              size="sm"
              class="rounded-xl font-bold text-xs h-9 px-4 border-slate-200 hover:bg-slate-50 gap-1.5"
              @click="openDetailModal(row)"
            >
              <Eye class="h-3.5 w-3.5 text-slate-500" /> Lihat Berita Acara
            </Button>
          </template>
        </DataTable>
      </Card>
    </div>

    <!-- TAB 3: RETUR PENJUALAN & SAMPINGAN NOTA PABRIK -->
    <div v-if="activeTab === 'sales_returns'" class="space-y-6">
      <Card class="border-none shadow-xl shadow-slate-200/50 bg-white rounded-[2.5rem] overflow-hidden border border-slate-100">
        <DataTable :columns="salesReturnColumns" :data="filteredSalesReturns">
          <template #cell-return_no="{ row }">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-blue-50 flex items-center justify-center text-blue-600 shadow-sm border border-blue-100">
                <RotateCcw class="h-5 w-5" />
              </div>
              <div>
                <p class="font-black text-sm text-slate-800">{{ row.return_no }}</p>
                <p class="text-[11px] font-semibold text-slate-400">
                  {{ row.return_type === 'POTONG_NOTA' ? 'Potong Nota' : row.return_type }}
                </p>
              </div>
            </div>
          </template>

          <template #cell-transaction="{ row }">
            <div>
              <p class="font-black text-xs text-slate-800">{{ row.transaction?.receipt_no || '-' }}</p>
              <p class="text-[11px] font-semibold text-slate-500">
                {{ row.transaction?.store?.name || row.transaction?.buyer_name || 'Penjualan Pabrik' }}
              </p>
            </div>
          </template>

          <template #cell-created_at="{ row }">
            <span class="text-xs font-semibold text-slate-600">{{ formatDate(row.created_at) }}</span>
          </template>

          <template #cell-items="{ row }">
            <div class="space-y-1 py-1">
              <div
                v-for="it in row.items"
                :key="it.id"
                class="flex items-center gap-2 text-xs"
              >
                <span class="font-bold text-slate-800">{{ it.product?.name || 'Barang Retur' }}</span>
                <span class="px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 font-black text-[10px]">
                  {{ it.quantity }} {{ it.product?.unit || 'PCS' }}
                </span>
                <span v-if="it.price > 0" class="text-slate-400 text-[10px]">
                  @{{ formatCurrency(it.price) }}
                </span>
              </div>
            </div>
          </template>

          <template #cell-total_amount="{ row }">
            <span class="font-black text-xs text-emerald-600">{{ formatCurrency(row.total_amount) }}</span>
          </template>

          <template #cell-status="{ row }">
            <Badge class="bg-emerald-100 text-emerald-800 border-none font-black text-[10px] px-2.5 py-1 rounded-full">
              {{ row.status === 'APPROVED' ? 'DITERIMA DI GUDANG' : row.status }}
            </Badge>
          </template>
        </DataTable>
      </Card>
    </div>

    <!-- MODAL 1: INSPEKSI QC & KONFIRMASI PENERIMAAN -->
    <div
      v-if="showInspectModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-4xl max-h-[90vh] shadow-2xl rounded-[2.5rem] overflow-hidden border-none flex flex-col bg-white animate-in zoom-in-95 duration-200"
      >
        <!-- Modal Header -->
        <CardHeader class="p-7 border-b bg-white flex flex-row items-center justify-between">
          <div>
            <div class="flex items-center gap-2">
              <CardTitle class="text-2xl font-black text-slate-900">Inspeksi QC & Terima Retur</CardTitle>
              <Badge class="bg-amber-100 text-amber-800 border-none font-bold text-xs">Penerimaan Pabrik</Badge>
            </div>
            <CardDescription class="mt-1">
              Verifikasi kuantitas fisik dan tentukan alokasi hasil QC: Stok Jadi (Bagus), Repack Botol, atau Scrap Limbah.
            </CardDescription>
          </div>
          <Button
            variant="ghost"
            size="icon"
            @click="showInspectModal = false"
            class="rounded-full h-10 w-10 text-slate-400 hover:text-slate-700"
          >
            <X class="h-6 w-6" />
          </Button>
        </CardHeader>

        <!-- Modal Body -->
        <CardContent class="p-7 space-y-6 overflow-y-auto custom-scrollbar flex-1 bg-[#fcfdfd]">
          <!-- Return Meta Info -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 bg-white p-5 rounded-2xl border border-slate-100 shadow-sm">
            <div>
              <p class="text-[10px] font-black uppercase tracking-wider text-slate-400">Nomor Retur</p>
              <p class="text-sm font-black text-slate-900 mt-0.5">
                {{ selectedReturn?.return_no || 'RTF-' + selectedReturn?.id.substring(0, 8) }}
              </p>
            </div>
            <div>
              <p class="text-[10px] font-black uppercase tracking-wider text-slate-400">Cabang Pengirim</p>
              <p class="text-sm font-bold text-slate-800 mt-0.5">
                {{ selectedReturn?.from_branch?.name || 'Cabang' }}
              </p>
            </div>
            <div>
              <p class="text-[10px] font-black uppercase tracking-wider text-slate-400">Tgl Pengiriman Cabang</p>
              <p class="text-sm font-bold text-slate-800 mt-0.5">
                {{ formatDate(selectedReturn?.shipped_at || selectedReturn?.created_at) }}
              </p>
            </div>
          </div>

          <!-- Items QC Breakdown List -->
          <div class="space-y-4">
            <div class="flex justify-between items-center px-1">
              <h4 class="text-xs font-black uppercase tracking-widest text-slate-500">
                Form Pemeriksaan QC Per Item ({{ inspectForm.items.length }} Produk)
              </h4>
              <span class="text-[11px] font-bold text-slate-400">
                Kuantitas dalam satuan unit produk
              </span>
            </div>

            <div
              v-for="(item, index) in inspectForm.items"
              :key="item.item_id"
              class="p-6 bg-white rounded-3xl border border-slate-200/80 shadow-sm space-y-5"
            >
              <!-- Item Header & Branch Declaration -->
              <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 pb-4 border-b border-slate-100">
                <div>
                  <h4 class="text-base font-black text-slate-900">{{ item.product_name }}</h4>
                  <div class="flex items-center gap-2 mt-1">
                    <span class="text-[10px] font-bold text-slate-500">Batch: {{ item.batch_no || 'DEFAULT' }}</span>
                    <span class="text-[10px] text-slate-300">•</span>
                    <span
                      class="text-[9px] font-black px-2 py-0.5 rounded uppercase"
                      :class="getReturnTypeBadgeClass(item.return_type)"
                    >
                      Deklarasi Cabang: {{ item.return_type === 'GOOD' ? 'Barang Bagus' : 'Barang Rusak' }}
                    </span>
                    <span v-if="item.damage_type && item.damage_type !== 'NONE'" class="text-[9px] font-bold text-rose-600 bg-rose-50 px-2 py-0.5 rounded">
                      {{ getDamageTypeLabel(item.damage_type) }}
                    </span>
                  </div>
                </div>

                <div class="bg-slate-50 px-4 py-2 rounded-2xl border border-slate-100 text-right">
                  <span class="text-[9px] font-black uppercase tracking-wider text-slate-400 block">Qty Dikirim Cabang</span>
                  <span class="text-sm font-black text-slate-900">{{ item.sent_qty }} {{ item.unit }}</span>
                </div>
              </div>

              <!-- QC Allocation Inputs -->
              <div>
                <p class="text-[11px] font-black uppercase tracking-wider text-slate-500 mb-3">
                  Hasil Pemeriksaan Fisik & Klasifikasi QC (Unit):
                </p>
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <!-- Good Qty -->
                  <div class="bg-emerald-50/60 p-4 rounded-2xl border border-emerald-100 space-y-1.5">
                    <label class="text-[10px] font-black uppercase tracking-wider text-emerald-800 flex items-center gap-1.5">
                      <CheckCircle2 class="h-3.5 w-3.5 text-emerald-600" />
                      1. Qty Bagus (Restock)
                    </label>
                    <input
                      type="number"
                      min="0"
                      v-model.number="item.received_good_qty"
                      class="w-full px-3.5 py-2.5 bg-white border border-emerald-200 rounded-xl text-sm font-black text-emerald-900 outline-none shadow-sm focus:ring-2 focus:ring-emerald-500/20"
                    />
                    <p class="text-[9px] font-medium text-emerald-700">Masuk kembali ke stok jual produk jadi.</p>
                  </div>

                  <!-- Bottle Damaged Qty -->
                  <div class="bg-orange-50/60 p-4 rounded-2xl border border-orange-100 space-y-1.5">
                    <label class="text-[10px] font-black uppercase tracking-wider text-orange-800 flex items-center gap-1.5">
                      <Package class="h-3.5 w-3.5 text-orange-600" />
                      2. Rusak Botol (Repack)
                    </label>
                    <input
                      type="number"
                      min="0"
                      v-model.number="item.received_bottle_damaged_qty"
                      class="w-full px-3.5 py-2.5 bg-white border border-orange-200 rounded-xl text-sm font-black text-orange-900 outline-none shadow-sm focus:ring-2 focus:ring-orange-500/20"
                    />
                    <p class="text-[9px] font-medium text-orange-700">Botol/kemasan rusak; masuk area rework/repack.</p>
                  </div>

                  <!-- Content Damaged Qty -->
                  <div class="bg-rose-50/60 p-4 rounded-2xl border border-rose-100 space-y-1.5">
                    <label class="text-[10px] font-black uppercase tracking-wider text-rose-800 flex items-center gap-1.5">
                      <Trash2 class="h-3.5 w-3.5 text-rose-600" />
                      3. Rusak Isi (Scrap)
                    </label>
                    <input
                      type="number"
                      min="0"
                      v-model.number="item.received_content_damaged_qty"
                      class="w-full px-3.5 py-2.5 bg-white border border-rose-200 rounded-xl text-sm font-black text-rose-900 outline-none shadow-sm focus:ring-2 focus:ring-rose-500/20"
                    />
                    <p class="text-[9px] font-medium text-rose-700">Basi / kadaluwarsa; dicatat pemusnahan limbah.</p>
                  </div>
                </div>
              </div>

              <!-- Target Warehouse & Inspection Notes -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-1">
                <div class="space-y-1.5">
                  <label class="text-[10px] font-black uppercase tracking-wider text-slate-500">
                    Gudang Pabrik Penempatan
                  </label>
                  <select
                    v-model="item.target_factory_warehouse_id"
                    class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold text-slate-800 outline-none shadow-sm"
                  >
                    <option :value="null">Gudang Pabrik Utama (Default)</option>
                    <option
                      v-for="wh in factoryStore.factoryWarehouses"
                      :key="wh.id"
                      :value="wh.id"
                    >
                      {{ wh.name }} ({{ wh.type || 'Gudang' }})
                    </option>
                  </select>
                </div>

                <div class="space-y-1.5">
                  <label class="text-[10px] font-black uppercase tracking-wider text-slate-500">
                    Catatan Temuan QC Item
                  </label>
                  <input
                    type="text"
                    v-model="item.inspection_notes"
                    placeholder="Contoh: 2 botol pecah di kardus, isi tumpah..."
                    class="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium text-slate-800 outline-none shadow-sm"
                  />
                </div>
              </div>

              <!-- Item Subtotal Summary -->
              <div class="flex justify-between items-center bg-slate-50 px-4 py-2.5 rounded-2xl text-xs font-bold border border-slate-100">
                <span class="text-slate-500">Total Fisik Diterima Produk Ini:</span>
                <div class="flex items-center gap-2">
                  <span class="text-slate-900 font-black">
                    {{ (item.received_good_qty || 0) + (item.received_bottle_damaged_qty || 0) + (item.received_content_damaged_qty || 0) }} {{ item.unit }}
                  </span>
                  <span
                    v-if="((item.received_good_qty || 0) + (item.received_bottle_damaged_qty || 0) + (item.received_content_damaged_qty || 0)) !== item.sent_qty"
                    class="text-[10px] font-black text-amber-700 bg-amber-100 px-2 py-0.5 rounded-md"
                  >
                    Selisih: {{ ((item.received_good_qty || 0) + (item.received_bottle_damaged_qty || 0) + (item.received_content_damaged_qty || 0)) - item.sent_qty }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- General Receiver & QC Officer Information -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 bg-white p-6 rounded-3xl border border-slate-200 shadow-sm">
            <div class="space-y-1.5">
              <label class="text-[10px] font-black uppercase tracking-wider text-slate-500">
                Nama Penerima / Checker Fisik
              </label>
              <input
                type="text"
                v-model="inspectForm.receiver_name"
                placeholder="Nama staf penerimaan barang..."
                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
            <div class="space-y-1.5">
              <label class="text-[10px] font-black uppercase tracking-wider text-slate-500">
                Nama Petugas QC / Supervisor
              </label>
              <input
                type="text"
                v-model="inspectForm.inspected_by_name"
                placeholder="Nama pemeriksa QC..."
                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
            <div class="sm:col-span-2 space-y-1.5">
              <label class="text-[10px] font-black uppercase tracking-wider text-slate-500">
                Catatan Berita Acara Penerimaan Retur
              </label>
              <textarea
                v-model="inspectForm.inspection_notes"
                rows="2"
                placeholder="Catatan umum proses penerimaan atau investigasi..."
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-xs outline-none resize-none"
              ></textarea>
            </div>
          </div>
        </CardContent>

        <!-- Modal Footer -->
        <CardFooter class="bg-white border-t p-6 flex justify-end gap-3">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-6"
            @click="showInspectModal = false"
          >
            Batal
          </Button>
          <Button
            class="rounded-2xl px-8 h-12 shadow-xl shadow-emerald-500/20 font-black text-white bg-emerald-600 hover:bg-emerald-700 text-xs gap-2"
            @click="submitInspection"
            :loading="isSubmitting"
          >
            <CheckCircle2 class="h-4 w-4" /> KONFIRMASI & UPDATE STOK PABRIK
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- MODAL 2: DETAIL BERITA ACARA RETUR -->
    <div
      v-if="showDetailModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-3xl max-h-[90vh] shadow-2xl rounded-[2.5rem] overflow-hidden border-none flex flex-col bg-white animate-in zoom-in-95 duration-200"
      >
        <CardHeader class="p-7 border-b bg-white flex flex-row items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 rounded-2xl bg-emerald-50 flex items-center justify-center text-emerald-600 border border-emerald-100">
              <FileText class="h-6 w-6" />
            </div>
            <div>
              <CardTitle class="text-xl font-black text-slate-900">Berita Acara Penerimaan Retur</CardTitle>
              <CardDescription>{{ selectedDetailReturn?.return_no }}</CardDescription>
            </div>
          </div>
          <Button
            variant="ghost"
            size="icon"
            @click="showDetailModal = false"
            class="rounded-full h-10 w-10 text-slate-400 hover:text-slate-700"
          >
            <X class="h-6 w-6" />
          </Button>
        </CardHeader>

        <CardContent class="p-7 space-y-6 overflow-y-auto custom-scrollbar flex-1">
          <!-- Summary Info Grid -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 bg-slate-50 p-5 rounded-2xl border border-slate-100 text-xs">
            <div>
              <span class="text-slate-400 text-[10px] uppercase font-black tracking-wider block">Asal Cabang</span>
              <span class="font-bold text-slate-800">{{ selectedDetailReturn?.from_branch?.name || 'Cabang' }}</span>
            </div>
            <div>
              <span class="text-slate-400 text-[10px] uppercase font-black tracking-wider block">Tgl Diterima Pabrik</span>
              <span class="font-bold text-slate-800">{{ formatDate(selectedDetailReturn?.received_at) }}</span>
            </div>
            <div>
              <span class="text-slate-400 text-[10px] uppercase font-black tracking-wider block">Penerima Fisik</span>
              <span class="font-bold text-slate-800">{{ selectedDetailReturn?.receiver_name || selectedDetailReturn?.received_by || '-' }}</span>
            </div>
            <div>
              <span class="text-slate-400 text-[10px] uppercase font-black tracking-wider block">Pemeriksa QC</span>
              <span class="font-bold text-slate-800">{{ selectedDetailReturn?.inspected_by_name || '-' }}</span>
            </div>
          </div>

          <!-- Total Breakdown Summary -->
          <div class="grid grid-cols-3 gap-3">
            <div class="bg-emerald-50 p-4 rounded-2xl border border-emerald-100 text-center">
              <span class="text-[10px] font-black uppercase text-emerald-800 block">Stok Bagus (Restocked)</span>
              <span class="text-xl font-black text-emerald-900 mt-1 block">
                {{ selectedDetailReturn?.total_good_qty || 0 }} Unit
              </span>
            </div>
            <div class="bg-orange-50 p-4 rounded-2xl border border-orange-100 text-center">
              <span class="text-[10px] font-black uppercase text-orange-800 block">Rusak Botol (Repack)</span>
              <span class="text-xl font-black text-orange-900 mt-1 block">
                {{ selectedDetailReturn?.total_bottle_damaged_qty || 0 }} Unit
              </span>
            </div>
            <div class="bg-rose-50 p-4 rounded-2xl border border-rose-100 text-center">
              <span class="text-[10px] font-black uppercase text-rose-800 block">Rusak Isi (Scrap)</span>
              <span class="text-xl font-black text-rose-900 mt-1 block">
                {{ selectedDetailReturn?.total_content_damaged_qty || 0 }} Unit
              </span>
            </div>
          </div>

          <!-- Items Table -->
          <div class="space-y-3">
            <p class="text-xs font-black uppercase tracking-wider text-slate-500">Rincian Barang & Disposisi QC:</p>
            <div class="border border-slate-200 rounded-2xl overflow-hidden">
              <table class="w-full text-left text-xs">
                <thead class="bg-slate-50 text-slate-400 uppercase text-[10px] font-black border-b border-slate-200">
                  <tr>
                    <th class="p-3.5">Produk</th>
                    <th class="p-3.5">Batch</th>
                    <th class="p-3.5 text-center">Dikirim</th>
                    <th class="p-3.5 text-center">Bagus</th>
                    <th class="p-3.5 text-center">Repack</th>
                    <th class="p-3.5 text-center">Scrap</th>
                    <th class="p-3.5">Catatan QC</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-for="item in selectedDetailReturn?.items" :key="item.id" class="hover:bg-slate-50/50 font-semibold">
                    <td class="p-3.5 font-bold text-slate-800">{{ item.product?.name }}</td>
                    <td class="p-3.5 text-slate-500 text-[11px]">{{ item.batch_no || 'DEFAULT' }}</td>
                    <td class="p-3.5 text-center font-bold text-slate-700">{{ item.quantity }} {{ item.unit }}</td>
                    <td class="p-3.5 text-center font-black text-emerald-600">{{ item.received_good_qty || 0 }}</td>
                    <td class="p-3.5 text-center font-black text-orange-600">{{ item.received_bottle_damaged_qty || 0 }}</td>
                    <td class="p-3.5 text-center font-black text-rose-600">{{ item.received_content_damaged_qty || 0 }}</td>
                    <td class="p-3.5 text-slate-500 text-[11px]">{{ item.inspection_notes || '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Inspection Notes -->
          <div v-if="selectedDetailReturn?.inspection_notes" class="bg-slate-50 p-4 rounded-2xl border border-slate-100 text-xs space-y-1">
            <span class="text-[10px] font-black uppercase text-slate-400 block">Catatan Umum QC / Pemeriksa</span>
            <p class="text-slate-700 font-medium">"{{ selectedDetailReturn.inspection_notes }}"</p>
          </div>
        </CardContent>

        <CardFooter class="bg-white border-t p-6 flex justify-end gap-3">
          <Button
            variant="outline"
            class="rounded-2xl font-bold h-11 px-6 border-slate-200"
            @click="showDetailModal = false"
          >
            Tutup
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- MODAL 3: TOLAK RETUR DIALOG -->
    <div
      v-if="showRejectModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-md shadow-2xl rounded-[2.5rem] overflow-hidden border-none flex flex-col bg-white animate-in zoom-in-95 duration-200"
      >
        <CardHeader class="p-6 border-b bg-rose-50/50 flex flex-row items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-rose-100 flex items-center justify-center text-rose-600 border border-rose-200">
              <XCircle class="h-5 w-5" />
            </div>
            <div>
              <CardTitle class="text-lg font-black text-slate-900">Tolak Pengajuan Retur</CardTitle>
              <CardDescription class="text-xs">{{ selectedRejectReturn?.return_no }}</CardDescription>
            </div>
          </div>
          <Button
            variant="ghost"
            class="h-8 w-8 p-0 rounded-full text-slate-400 hover:text-slate-600"
            @click="showRejectModal = false"
          >
            <X class="h-4 w-4" />
          </Button>
        </CardHeader>

        <CardContent class="p-6 space-y-4 text-xs font-medium text-slate-700">
          <p class="text-slate-500">
            Penolakan retur ini akan membatalkan pemotongan tagihan, mengembalikan stok cabang, dan saldo piutang cabang tetap ditagihkan penuh.
          </p>

          <div class="space-y-1.5">
            <label class="font-bold text-slate-800 uppercase tracking-wider text-[10px]">
              Alasan Penolakan Retur <span class="text-rose-500">*</span>
            </label>
            <textarea
              v-model="rejectReason"
              rows="3"
              placeholder="Contoh: Bukti fisik tidak valid, segel rusak karena kelalaian cabang, klaim kadaluarsa..."
              class="w-full p-3 bg-slate-50 border border-slate-200 rounded-xl outline-none focus:ring-2 focus:ring-rose-500 resize-none font-medium"
            ></textarea>
          </div>
        </CardContent>

        <CardFooter class="bg-slate-50 p-5 border-t flex justify-end gap-2">
          <Button
            variant="ghost"
            class="rounded-xl font-bold text-xs"
            @click="showRejectModal = false"
          >
            Batal
          </Button>
          <Button
            class="rounded-xl bg-rose-600 hover:bg-rose-700 text-white font-black text-xs px-5 shadow-md shadow-rose-600/20"
            @click="submitRejectReturn"
            :loading="isSubmittingReject"
            :disabled="!rejectReason.trim()"
          >
            Tolak Retur
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, reactive } from 'vue'
import { useFactoryStore, type FactoryReturn } from '@/stores/factory'
import { useMasterDataStore } from '@/stores/masterData'
import apiClient from '@/api/axios'
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
import { toast } from 'vue-sonner'
import {
  RotateCcw,
  RefreshCw,
  Search,
  Store,
  Calendar,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Package,
  SlidersHorizontal,
  ClipboardCheck,
  Eye,
  FileText,
  X,
  Trash2,
} from 'lucide-vue-next'

const factoryStore = useFactoryStore()
const masterStore = useMasterDataStore()
const loading = ref(false)
const search = ref('')
const activeTab = ref<'pending' | 'history' | 'sales_returns'>('pending')

const showInspectModal = ref(false)
const showDetailModal = ref(false)
const showRejectModal = ref(false)
const selectedReturn = ref<FactoryReturn | null>(null)
const selectedDetailReturn = ref<FactoryReturn | null>(null)
const selectedRejectReturn = ref<any>(null)
const rejectReason = ref('')
const isSubmitting = ref(false)
const isSubmittingReject = ref(false)

// Inspection Form State
const inspectForm = reactive({
  receiver_name: '',
  inspected_by_name: '',
  inspection_notes: '',
  items: [] as any[],
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

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

const getReturnTypeBadgeClass = (returnType?: string) => {
  if (returnType === 'GOOD') {
    return 'bg-emerald-100 text-emerald-800'
  }
  return 'bg-rose-100 text-rose-800'
}

const getDamageTypeLabel = (damageType?: string) => {
  switch (damageType) {
    case 'BOTTLE_PACKAGING':
      return 'Rusak Kemasan/Botol (Bisa Repack)'
    case 'PRODUCT_CONTENT':
      return 'Isi Rusak / Basi'
    case 'EXPIRED':
      return 'Kadaluwarsa'
    case 'TOTAL_DAMAGE':
      return 'Rusak Total / Hancur'
    case 'OTHER':
      return 'Lain-lain / Cacat Batch'
    case 'NONE':
      return 'Kondisi Baik'
    default:
      return damageType || 'Rusak'
  }
}

const getDamageTypeBadgeClass = (damageType?: string) => {
  switch (damageType) {
    case 'BOTTLE_PACKAGING':
      return 'bg-orange-100 text-orange-800'
    case 'PRODUCT_CONTENT':
      return 'bg-rose-100 text-rose-800'
    case 'EXPIRED':
      return 'bg-purple-100 text-purple-800'
    case 'TOTAL_DAMAGE':
      return 'bg-slate-200 text-slate-800'
    default:
      return 'bg-slate-100 text-slate-700'
  }
}

const refreshAll = async () => {
  loading.value = true
  try {
    await Promise.all([
      factoryStore.fetchAllFactoryReturns(),
      factoryStore.fetchFactories(),
      factoryStore.fetchProducts(),
      masterStore.fetchBranches(),
      fetchSalesReturns(),
    ])
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refreshAll()
})

const totalReturnsCount = computed(() => factoryStore.factoryReturns.length)

const pendingReturns = computed(() => {
  return factoryStore.factoryReturns
    .filter(
      (r: any) => r.status === 'SHIPPED' || r.status === 'REQUESTED' || r.status === 'PENDING',
    )
    .sort((a: any, b: any) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})

const historyReturns = computed(() => {
  return factoryStore.factoryReturns
    .filter((r: any) => r.status === 'RECEIVED' || r.status === 'COMPLETED')
    .sort(
      (a: any, b: any) =>
        new Date(b.received_at || b.updated_at).getTime() -
        new Date(a.received_at || a.updated_at).getTime(),
    )
})

const filteredPending = computed(() => {
  if (!search.value) return pendingReturns.value
  const q = search.value.toLowerCase()
  return pendingReturns.value.filter(
    (r: any) =>
      r.return_no?.toLowerCase().includes(q) || r.from_branch?.name?.toLowerCase().includes(q),
  )
})

const filteredHistory = computed(() => {
  if (!search.value) return historyReturns.value
  const q = search.value.toLowerCase()
  return historyReturns.value.filter(
    (r: any) =>
      r.return_no?.toLowerCase().includes(q) || r.from_branch?.name?.toLowerCase().includes(q),
  )
})

// Metrics totals
const totalGoodRestocked = computed(() => {
  return historyReturns.value.reduce((acc, r: any) => {
    return acc + (r.total_good_qty || r.items?.reduce((s: number, i: any) => s + (i.received_good_qty || 0), 0) || 0)
  }, 0)
})

const totalRepackBotol = computed(() => {
  return historyReturns.value.reduce((acc, r: any) => {
    return acc + (r.total_bottle_damaged_qty || r.items?.reduce((s: number, i: any) => s + (i.received_bottle_damaged_qty || 0), 0) || 0)
  }, 0)
})

const totalScrapLimbah = computed(() => {
  return historyReturns.value.reduce((acc, r: any) => {
    return acc + (r.total_content_damaged_qty || r.items?.reduce((s: number, i: any) => s + (i.received_content_damaged_qty || 0), 0) || 0)
  }, 0)
})

const openInspectModal = (rtn: FactoryReturn) => {
  selectedReturn.value = rtn
  inspectForm.receiver_name = ''
  inspectForm.inspected_by_name = ''
  inspectForm.inspection_notes = ''

  // Initialize per-item inspection form
  inspectForm.items = (rtn.items || []).map((item) => {
    const isGood = item.return_type === 'GOOD'
    const isBottle = item.damage_type === 'BOTTLE_PACKAGING'
    const isContent = item.damage_type === 'PRODUCT_CONTENT' || item.damage_type === 'EXPIRED' || item.damage_type === 'TOTAL_DAMAGE'

    return {
      item_id: item.id,
      product_name: item.product?.name || 'Produk',
      batch_no: item.batch_no || 'DEFAULT',
      unit: item.unit || item.product?.unit || 'UNIT',
      sent_qty: item.quantity,
      return_type: item.return_type || 'DAMAGED',
      damage_type: item.damage_type || 'OTHER',
      damage_reason: item.damage_reason || '',
      received_good_qty: isGood ? item.quantity : 0,
      received_bottle_damaged_qty: (!isGood && !isContent) ? item.quantity : 0,
      received_content_damaged_qty: isContent ? item.quantity : 0,
      target_factory_warehouse_id: null,
      inspection_notes: '',
    }
  })

  showInspectModal.value = true
}

const submitInspection = async () => {
  if (!selectedReturn.value) return

  // Validate at least one quantity is entered
  const invalidTotal = inspectForm.items.some((i) => {
    const total = (i.received_good_qty || 0) + (i.received_bottle_damaged_qty || 0) + (i.received_content_damaged_qty || 0)
    return total < 0
  })

  if (invalidTotal) {
    toast.error('Kuantitas QC tidak boleh negatif')
    return
  }

  const payload = {
    receiver_name: inspectForm.receiver_name || 'Admin Pabrik',
    inspected_by_name: inspectForm.inspected_by_name || 'Petugas QC Pabrik',
    inspection_notes: inspectForm.inspection_notes,
    items: inspectForm.items.map((i) => ({
      item_id: i.item_id,
      received_good_qty: Number(i.received_good_qty) || 0,
      received_bottle_damaged_qty: Number(i.received_bottle_damaged_qty) || 0,
      received_content_damaged_qty: Number(i.received_content_damaged_qty) || 0,
      target_factory_warehouse_id: i.target_factory_warehouse_id || null,
      inspection_notes: i.inspection_notes || '',
    })),
  }

  isSubmitting.value = true
  try {
    await factoryStore.receiveFactoryReturn(selectedReturn.value.id, payload)
    toast.success('Retur berhasil diinspeksi & diterima. Stok pabrik telah diperbarui sesuai alokasi QC.')
    showInspectModal.value = false
    refreshAll()
  } catch (err: any) {
    toast.error('Gagal memproses penerimaan retur: ' + getErrorMessage(err))
  } finally {
    isSubmitting.value = false
  }
}

const openRejectModal = (rtn: any) => {
  selectedRejectReturn.value = rtn
  rejectReason.value = ''
  showRejectModal.value = true
}

const submitRejectReturn = async () => {
  if (!selectedRejectReturn.value) return
  if (!rejectReason.value.trim()) {
    toast.error('Mohon masukkan alasan penolakan retur.')
    return
  }

  isSubmittingReject.value = true
  try {
    await apiClient.post(`/factory-returns/${selectedRejectReturn.value.id}/reject`, {
      reason: rejectReason.value.trim(),
    })
    toast.success(`Pengajuan retur ${selectedRejectReturn.value.return_no || ''} berhasil ditolak. Saldo piutang cabang tetap utuh.`)
    showRejectModal.value = false
    refreshAll()
  } catch (err: any) {
    toast.error('Gagal menolak retur: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmittingReject.value = false
  }
}

const openDetailModal = (row: FactoryReturn) => {
  selectedDetailReturn.value = row
  showDetailModal.value = true
}

const historyColumns = [
  { key: 'return_no', label: 'No. Retur & Cabang', class: 'w-[240px]' },
  { key: 'received_at', label: 'Waktu Diterima', class: 'w-[200px]' },
  { key: 'items_breakdown', label: 'Rincian Disposisi Hasil QC' },
  { key: 'actions', label: 'Aksi', class: 'w-[180px] text-right' },
]

const salesReturnColumns = [
  { key: 'return_no', label: 'No. Retur', class: 'w-[220px]' },
  { key: 'transaction', label: 'Faktur & Pelanggan / Toko', class: 'w-[260px]' },
  { key: 'created_at', label: 'Tgl Retur', class: 'w-[160px]' },
  { key: 'items', label: 'Rincian Barang Sampingan / Wadah' },
  { key: 'total_amount', label: 'Total Nilai (Potong Nota)', class: 'w-[180px]' },
  { key: 'status', label: 'Status', class: 'w-[160px]' },
]

const factorySalesReturns = ref<any[]>([])
const isLoadingSalesReturns = ref(false)

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const fetchSalesReturns = async () => {
  isLoadingSalesReturns.value = true
  try {
    let branchId = factoryStore.selectedFactory?.branch_id || ''
    if (!branchId) {
      const pabrikBranch = masterStore.branches.find((b: any) => /pabrik/i.test(b.name))
      if (pabrikBranch) branchId = pabrikBranch.id
    }
    if (branchId) {
      const res = await apiClient.get(`/sales-returns/branch/${branchId}`)
      factorySalesReturns.value = res.data || []
    }
  } catch (err) {
    console.error('Failed to load factory sales returns:', err)
  } finally {
    isLoadingSalesReturns.value = false
  }
}

const filteredSalesReturns = computed(() => {
  if (!search.value) return factorySalesReturns.value
  const q = search.value.toLowerCase().trim()
  return factorySalesReturns.value.filter((r: any) => {
    return (
      r.return_no?.toLowerCase().includes(q) ||
      r.transaction?.receipt_no?.toLowerCase().includes(q) ||
      r.transaction?.store?.name?.toLowerCase().includes(q) ||
      r.transaction?.buyer_name?.toLowerCase().includes(q) ||
      r.items?.some((it: any) => it.product?.name?.toLowerCase().includes(q))
    )
  })
})
</script>
