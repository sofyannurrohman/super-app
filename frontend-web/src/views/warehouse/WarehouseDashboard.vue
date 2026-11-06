<template>
  <div class="space-y-8 pb-14 max-w-[1600px] mx-auto animate-fade-in">
    <!-- Header: Dynamic Branch Sync + Range Filter + Live Warehouse Status -->
    <div
      class="bg-white/80 backdrop-blur-md p-6 sm:p-8 rounded-3xl border border-slate-200/80 shadow-sm flex flex-col lg:flex-row lg:items-center justify-between gap-6 relative overflow-hidden"
    >
      <div class="relative z-10">
        <div class="flex items-center gap-2.5 text-xs font-extrabold uppercase tracking-wider text-primary">
          <span class="flex h-2.5 w-2.5 rounded-full bg-primary animate-ping"></span>
          <span
            class="inline-flex items-center gap-1.5 bg-primary/10 px-3 py-1 rounded-full text-primary border border-primary/20"
          >
            <Activity class="w-3.5 h-3.5" />
            Live Warehouse Center • {{ currentBranchName }}
          </span>
        </div>
        <h1 class="text-2xl sm:text-3xl font-black text-slate-900 tracking-tight mt-3 flex items-center gap-3">
          <Warehouse class="h-8 w-8 text-primary" />
          Pusat Komando Logistik & Gudang
        </h1>
        <p class="text-sm font-medium text-slate-500 mt-1 flex items-center gap-2">
          <span>{{ currentFormattedDate }}</span>
          <span class="text-slate-300">•</span>
          <span>Monitoring stok fisik, arus mutasi barang, serta distribusi logistik</span>
        </p>
      </div>

      <!-- Action & Filter Bar -->
      <div class="flex flex-wrap items-center gap-3 relative z-10">
        <!-- Range Filter -->
        <div class="flex items-center gap-1 bg-slate-100 p-1.5 rounded-2xl border border-slate-200/60 shrink-0">
          <button
            v-for="range in [
              { id: '7d', label: '7 Hari Terakhir' },
              { id: '30d', label: '30 Hari' },
              { id: 'year', label: 'Tahun Ini' },
            ]"
            :key="range.id"
            @click="timeRange = range.id as '7d' | '30d' | 'year'"
            class="px-4 py-2 text-xs font-extrabold rounded-xl transition-all duration-200"
            :class="[
              timeRange === range.id
                ? 'bg-white text-primary shadow-sm ring-1 ring-slate-200/60 font-black'
                : 'text-slate-500 hover:text-slate-800',
            ]"
          >
            {{ range.label }}
          </button>
        </div>

        <!-- Refresh Button -->
        <button
          @click="fetchDashboardData"
          :disabled="isLoading"
          class="p-2.5 bg-slate-100 hover:bg-slate-200/70 text-slate-600 rounded-2xl border border-slate-200/60 transition-all active:scale-95 disabled:opacity-50"
          title="Segarkan Data"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isLoading }" />
        </button>
      </div>

      <!-- Background Glow Element -->
      <div class="absolute -right-12 -top-12 w-64 h-64 bg-primary/5 rounded-full blur-3xl pointer-events-none"></div>
    </div>

    <!-- Quick Action Launcher Hub -->
    <div
      class="bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 rounded-3xl p-5 text-white shadow-xl shadow-slate-900/5 border border-slate-800 flex flex-wrap items-center justify-between gap-4"
    >
      <div class="flex items-center gap-3">
        <div class="p-2.5 bg-white/10 rounded-2xl backdrop-blur-md">
          <Sparkles class="w-5 h-5 text-amber-400" />
        </div>
        <div>
          <h4 class="text-sm font-extrabold tracking-wide">Pintasan Cepat Operasional Gudang</h4>
          <p class="text-xs text-slate-300">Aksi instan administrasi penerimaan, transfer sales, dan stok fisik</p>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <router-link
          to="/warehouse/stock"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all backdrop-blur-md border border-white/10 active:scale-95"
        >
          <Boxes class="w-3.5 h-3.5 text-emerald-400" />
          Terima Kiriman Pabrik
        </router-link>
        <router-link
          to="/warehouse/sales-transfers"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all backdrop-blur-md border border-white/10 active:scale-95"
        >
          <ArrowLeftRight class="w-3.5 h-3.5 text-sky-400" />
          Transfer ke Sales
        </router-link>
        <router-link
          to="/warehouse/do-management"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all backdrop-blur-md border border-white/10 active:scale-95"
        >
          <Truck class="w-3.5 h-3.5 text-amber-400" />
          Kelola DO Kirim
        </router-link>
        <router-link
          to="/warehouse/logs"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all backdrop-blur-md border border-white/10 active:scale-95"
        >
          <History class="w-3.5 h-3.5 text-teal-300" />
          Log Mutasi
        </router-link>
        <router-link
          to="/analytics"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-primary hover:bg-primary/90 text-white rounded-xl text-xs font-bold transition-all shadow-md shadow-primary/20 active:scale-95"
        >
          <Sparkles class="w-3.5 h-3.5" />
          AI Analytics
        </router-link>
      </div>
    </div>

    <!-- 5 Hero KPI Metric Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-5">
      <!-- Card 1: Total Stok Fisik -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Total Stok Fisik</span>
            <div class="p-2.5 bg-blue-50 text-blue-600 rounded-2xl group-hover:scale-110 transition-transform">
              <Package class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-20 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ (overviewData?.summary?.total_stock_physical ?? 0).toLocaleString('id-ID') }} <span class="text-xs font-bold text-slate-400">PCS</span></span>
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span class="text-slate-500 font-semibold">Tersimpan di gudang</span>
          <button @click="activeTab = 'inventory'" class="text-primary font-bold hover:underline inline-flex items-center gap-0.5">
            Katalog <ChevronRight class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      <!-- Card 2: Total SKU Produk -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Varian SKU Aktif</span>
            <div class="p-2.5 bg-indigo-50 text-indigo-600 rounded-2xl group-hover:scale-110 transition-transform">
              <Boxes class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-16 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ overviewData?.summary?.total_sku_count ?? 0 }}</span>
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span class="text-slate-500 font-semibold">Item terdaftar</span>
          <router-link to="/warehouse/stock" class="text-primary font-bold hover:underline">Kelola Limit</router-link>
        </div>
      </div>

      <!-- Card 3: Inbound Shipments -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Kiriman Pabrik (Inbound)</span>
            <div
              class="p-2.5 rounded-2xl group-hover:scale-110 transition-transform"
              :class="(overviewData?.summary?.pending_inbound_count ?? 0) > 0 ? 'bg-amber-50 text-amber-600' : 'bg-slate-50 text-slate-400'"
            >
              <Truck class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-16 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ overviewData?.summary?.pending_inbound_count ?? 0 }}</span>
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span
            :class="(overviewData?.summary?.pending_inbound_count ?? 0) > 0 ? 'text-amber-600 font-bold' : 'text-slate-400 font-semibold'"
          >
            {{ (overviewData?.summary?.pending_inbound_count ?? 0) > 0 ? 'Dalam pengiriman' : 'Tidak ada kiriman' }}
          </span>
          <button @click="activeTab = 'inbound'" class="text-primary font-bold hover:underline">Lihat</button>
        </div>
      </div>

      <!-- Card 4: Low Stock Alert -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Stok Menipis</span>
            <div
              class="p-2.5 rounded-2xl group-hover:scale-110 transition-transform"
              :class="(overviewData?.summary?.low_stock_count ?? 0) > 0 ? 'bg-rose-50 text-rose-600' : 'bg-slate-50 text-slate-400'"
            >
              <AlertTriangle class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-16 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ overviewData?.summary?.low_stock_count ?? 0 }}</span>
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span
            :class="(overviewData?.summary?.low_stock_count ?? 0) > 0 ? 'text-rose-600 font-bold' : 'text-emerald-600 font-bold'"
          >
            {{ (overviewData?.summary?.low_stock_count ?? 0) > 0 ? 'Perlu Restock!' : 'Stok Aman' }}
          </span>
          <button @click="activeTab = 'low_stock'" class="text-primary font-bold hover:underline">Prioritas</button>
        </div>
      </div>

      <!-- Card 5: Arus Mutasi Hari Ini -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Mutasi Hari Ini</span>
            <div class="p-2.5 bg-teal-50 text-teal-600 rounded-2xl group-hover:scale-110 transition-transform">
              <History class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2 flex items-baseline gap-2">
            <h3 class="text-xl font-black text-emerald-600 tracking-tight">
              +{{ overviewData?.summary?.today_in_quantity ?? 0 }}
            </h3>
            <span class="text-xs text-slate-300 font-bold">/</span>
            <h3 class="text-xl font-black text-amber-600 tracking-tight">
              -{{ overviewData?.summary?.today_out_quantity ?? 0 }}
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span class="text-slate-500 font-semibold">Barang Masuk / Keluar</span>
          <button @click="activeTab = 'logs'" class="text-primary font-bold hover:underline">Log</button>
        </div>
      </div>
    </div>

    <!-- Middle Visual Section: ECharts (Arus Mutasi & Top Stock Distribution) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Main Mutation Trend Chart (7 Cols) -->
      <div class="lg:col-span-7 bg-white rounded-3xl border border-slate-200 p-6 sm:p-7 shadow-sm flex flex-col justify-between">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <div class="flex items-center gap-2">
              <h3 class="text-lg font-black text-slate-900 tracking-tight">Arus Pergerakan Stok (Inbound vs Outbound)</h3>
              <span class="px-2.5 py-0.5 rounded-full text-[10px] font-extrabold bg-emerald-100 text-emerald-700">Data Riil</span>
            </div>
            <p class="text-xs text-slate-500 mt-1">
              Perbandingan kuantitas barang masuk (*Pabrik/Retur*) dan barang keluar (*Sales Transfer/DO*).
            </p>
          </div>
          <router-link
            to="/warehouse/logs"
            class="inline-flex items-center gap-1.5 text-xs font-extrabold text-primary hover:text-primary/80 bg-primary/5 hover:bg-primary/10 px-3.5 py-2 rounded-xl transition-all self-start sm:self-auto"
          >
            Audit Log <ArrowRight class="w-3.5 h-3.5" />
          </router-link>
        </div>

        <div class="h-[340px] w-full relative">
          <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-white/70 backdrop-blur-sm z-10">
            <RefreshCw class="w-6 h-6 animate-spin text-primary" />
          </div>
          <v-chart class="w-full h-full" :option="mutationTrendOption" autoresize />
        </div>
      </div>

      <!-- Top Stock Distribution Chart (5 Cols) -->
      <div class="lg:col-span-5 bg-white rounded-3xl border border-slate-200 p-6 sm:p-7 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-lg font-black text-slate-900 tracking-tight">Distribusi Stok Terbanyak</h3>
            <p class="text-xs text-slate-500 mt-1">Komposisi produk dengan volume fisik tertinggi di gudang</p>
          </div>
          <button @click="activeTab = 'inventory'" class="text-xs font-bold text-slate-400 hover:text-slate-600">
            Semua
          </button>
        </div>

        <div class="h-[340px] w-full relative">
          <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-white/70 backdrop-blur-sm z-10">
            <RefreshCw class="w-6 h-6 animate-spin text-primary" />
          </div>
          <v-chart class="w-full h-full" :option="topStockDistributionOption" autoresize />
        </div>
      </div>
    </div>

    <!-- Stock Health Indicator & Quick Buffer Overview -->
    <div class="bg-white rounded-3xl border border-slate-200 p-6 sm:p-7 shadow-sm">
      <div class="flex items-center justify-between mb-5">
        <div>
          <h3 class="text-lg font-black text-slate-900 tracking-tight">Status Kesehatan & Buffer Stok Fisik</h3>
          <p class="text-xs text-slate-500 mt-0.5">Klasifikasi tingkat ketersediaan stok produk terhadap batas limit minimum</p>
        </div>
        <span class="text-xs font-bold text-slate-400">Buffer Safety Stock</span>
      </div>

      <!-- 4 Sub-Boxes -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div class="bg-emerald-50/70 border border-emerald-100 p-4 rounded-2xl">
          <div class="flex items-center gap-2 text-emerald-700 text-xs font-bold">
            <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
            Stok Aman
          </div>
          <p class="text-2xl font-black text-emerald-900 mt-2">{{ overviewData?.stock_health?.safe_count ?? 0 }} SKU</p>
          <p class="text-[11px] text-emerald-600 font-semibold mt-0.5">Di atas batas minimum</p>
        </div>

        <div class="bg-amber-50/70 border border-amber-100 p-4 rounded-2xl">
          <div class="flex items-center gap-2 text-amber-700 text-xs font-bold">
            <span class="w-2 h-2 rounded-full bg-amber-500"></span>
            Waspada
          </div>
          <p class="text-2xl font-black text-amber-900 mt-2">{{ overviewData?.stock_health?.warning_count ?? 0 }} SKU</p>
          <p class="text-[11px] text-amber-600 font-semibold mt-0.5">50% - 99% dari limit</p>
        </div>

        <div class="bg-rose-50/70 border border-rose-100 p-4 rounded-2xl">
          <div class="flex items-center gap-2 text-rose-700 text-xs font-bold">
            <span class="w-2 h-2 rounded-full bg-rose-500"></span>
            Stok Kritis
          </div>
          <p class="text-2xl font-black text-rose-900 mt-2">{{ overviewData?.stock_health?.critical_count ?? 0 }} SKU</p>
          <p class="text-[11px] text-rose-600 font-semibold mt-0.5">&lt; 50% dari limit</p>
        </div>

        <div class="bg-slate-50 border border-slate-200/80 p-4 rounded-2xl">
          <div class="flex items-center gap-2 text-slate-600 text-xs font-bold">
            <span class="w-2 h-2 rounded-full bg-slate-400"></span>
            Stok Habis (0)
          </div>
          <p class="text-2xl font-black text-slate-800 mt-2">{{ overviewData?.stock_health?.out_of_stock_count ?? 0 }} SKU</p>
          <p class="text-[11px] text-slate-500 font-semibold mt-0.5">Kuantitas 0 unit</p>
        </div>
      </div>

      <!-- Segmented Health Bar -->
      <div class="mt-6 space-y-2">
        <div class="w-full h-3.5 bg-slate-100 rounded-full overflow-hidden flex gap-0.5 p-0.5">
          <div
            class="bg-emerald-500 rounded-l-full transition-all duration-700"
            :style="{ width: `${(overviewData?.summary?.total_sku_count ?? 0) > 0 ? ((overviewData?.stock_health?.safe_count ?? 0) / overviewData.summary.total_sku_count) * 100 : 0}%` }"
            title="Stok Aman"
          ></div>
          <div
            class="bg-amber-400 transition-all duration-700"
            :style="{ width: `${(overviewData?.summary?.total_sku_count ?? 0) > 0 ? ((overviewData?.stock_health?.warning_count ?? 0) / overviewData.summary.total_sku_count) * 100 : 0}%` }"
            title="Waspada"
          ></div>
          <div
            class="bg-rose-500 transition-all duration-700"
            :style="{ width: `${(overviewData?.summary?.total_sku_count ?? 0) > 0 ? ((overviewData?.stock_health?.critical_count ?? 0) / overviewData.summary.total_sku_count) * 100 : 0}%` }"
            title="Kritis"
          ></div>
          <div
            class="bg-slate-300 rounded-r-full flex-1 transition-all duration-700"
            title="Stok Habis"
          ></div>
        </div>
      </div>
    </div>

    <!-- Bottom Interactive Operational Tabs Section -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
      <!-- Tab Navigation Header -->
      <div class="p-6 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="flex flex-wrap items-center gap-2">
          <button
            @click="activeTab = 'inbound'"
            class="px-4 py-2.5 rounded-2xl text-xs font-extrabold transition-all flex items-center gap-2"
            :class="[
              activeTab === 'inbound'
                ? 'bg-slate-900 text-white shadow-md'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70',
            ]"
          >
            <Truck class="w-3.5 h-3.5" />
            Kiriman Masuk (Inbound)
            <span
              v-if="(overviewData?.summary?.pending_inbound_count ?? 0) > 0"
              class="px-2 py-0.5 rounded-full text-[10px] bg-amber-500 text-white font-black"
            >
              {{ overviewData?.summary?.pending_inbound_count ?? 0 }}
            </span>
          </button>

          <button
            @click="activeTab = 'low_stock'"
            class="px-4 py-2.5 rounded-2xl text-xs font-extrabold transition-all flex items-center gap-2"
            :class="[
              activeTab === 'low_stock'
                ? 'bg-slate-900 text-white shadow-md'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70',
            ]"
          >
            <AlertTriangle class="w-3.5 h-3.5" />
            Prioritas Restock
            <span
              v-if="(overviewData?.summary?.low_stock_count ?? 0) > 0"
              class="px-2 py-0.5 rounded-full text-[10px] bg-rose-500 text-white font-black"
            >
              {{ overviewData?.summary?.low_stock_count ?? 0 }}
            </span>
          </button>

          <button
            @click="activeTab = 'logs'"
            class="px-4 py-2.5 rounded-2xl text-xs font-extrabold transition-all flex items-center gap-2"
            :class="[
              activeTab === 'logs'
                ? 'bg-slate-900 text-white shadow-md'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70',
            ]"
          >
            <History class="w-3.5 h-3.5" />
            Log Mutasi Terkini
            <span class="px-2 py-0.5 rounded-full text-[10px] bg-slate-200 text-slate-700 font-bold">
              {{ (overviewData?.recent_logs || []).length }}
            </span>
          </button>

          <button
            @click="activeTab = 'inventory'"
            class="px-4 py-2.5 rounded-2xl text-xs font-extrabold transition-all flex items-center gap-2"
            :class="[
              activeTab === 'inventory'
                ? 'bg-slate-900 text-white shadow-md'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70',
            ]"
          >
            <Boxes class="w-3.5 h-3.5" />
            Katalog Inventaris
            <span class="px-2 py-0.5 rounded-full text-[10px] bg-slate-200 text-slate-700 font-bold">
              {{ groupedInventory.length }}
            </span>
          </button>
        </div>

        <!-- Search Filter for Inventory Tab -->
        <div v-if="activeTab === 'inventory'" class="relative w-full sm:w-64">
          <Search class="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            v-model="inventorySearchQuery"
            type="text"
            placeholder="Cari SKU atau nama produk..."
            class="w-full bg-slate-100 pl-10 pr-4 py-2 rounded-xl text-xs font-medium border border-slate-200/80 focus:bg-white focus:outline-none focus:ring-2 focus:ring-primary/20"
          />
        </div>
      </div>

      <!-- Tab Content Area -->
      <div class="p-6">
        <!-- TAB 1: Inbound Shipments -->
        <div v-if="activeTab === 'inbound'">
          <div v-if="(overviewData.pending_inbound_shipments || []).length === 0" class="py-14 text-center text-slate-400">
            <CheckCircle2 class="w-12 h-12 mx-auto mb-3 opacity-20 text-emerald-500" />
            <p class="text-sm font-bold">Tidak ada kiriman pabrik yang sedang dalam proses pengiriman.</p>
            <router-link to="/warehouse/stock" class="text-xs font-bold text-primary mt-2 inline-block hover:underline">
              Buka Penerimaan Gudang
            </router-link>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="bg-slate-50 text-[11px] font-extrabold text-slate-400 uppercase tracking-wider">
                  <th class="px-5 py-3.5 rounded-l-2xl">No. DO / Surat Jalan</th>
                  <th class="px-5 py-3.5">Asal Pabrik</th>
                  <th class="px-5 py-3.5">Produk</th>
                  <th class="px-5 py-3.5 text-center">Jumlah</th>
                  <th class="px-5 py-3.5">Status</th>
                  <th class="px-5 py-3.5 text-right rounded-r-2xl">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="shipment in (overviewData.pending_inbound_shipments || [])"
                  :key="shipment.id"
                  class="group hover:bg-slate-50/80 transition-colors"
                >
                  <td class="px-5 py-4">
                    <p class="font-mono font-bold text-slate-900 text-sm">{{ shipment.delivery_order_no }}</p>
                    <p class="text-[11px] text-slate-400 font-medium">{{ shipment.created_at }}</p>
                  </td>
                  <td class="px-5 py-4 text-xs font-bold text-slate-700">
                    {{ shipment.from_factory }}
                  </td>
                  <td class="px-5 py-4">
                    <p class="font-extrabold text-slate-900 text-sm">{{ shipment.product_name }}</p>
                  </td>
                  <td class="px-5 py-4 text-center">
                    <span class="px-3 py-1 bg-blue-50 text-blue-700 rounded-xl text-xs font-black border border-blue-100">
                      {{ (shipment.quantity || 0).toLocaleString('id-ID') }} {{ shipment.unit }}
                    </span>
                  </td>
                  <td class="px-5 py-4">
                    <span
                      class="px-3 py-1 rounded-xl text-xs font-extrabold inline-flex items-center gap-1.5"
                      :class="
                        shipment.status === 'ARRIVED'
                          ? 'bg-emerald-100 text-emerald-800'
                          : shipment.status === 'SHIPPED'
                            ? 'bg-amber-100 text-amber-800'
                            : 'bg-slate-100 text-slate-700'
                      "
                    >
                      <Truck class="w-3.5 h-3.5" />
                      {{ shipment.status }}
                    </span>
                  </td>
                  <td class="px-5 py-4 text-right">
                    <router-link
                      to="/warehouse/stock"
                      class="inline-flex items-center gap-1 bg-primary text-white px-3.5 py-2 rounded-xl text-xs font-bold shadow-sm hover:bg-primary/90 transition-all active:scale-95"
                    >
                      Terima Barang <ArrowRight class="w-3.5 h-3.5" />
                    </router-link>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- TAB 2: Low Stock Restock Priority -->
        <div v-if="activeTab === 'low_stock'">
          <div v-if="(overviewData.low_stock_items || []).length === 0" class="py-14 text-center text-slate-400">
            <CheckCircle2 class="w-12 h-12 mx-auto mb-3 opacity-20 text-emerald-500" />
            <p class="text-sm font-bold">Semua stok produk berada di atas batas aman!</p>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="bg-slate-50 text-[11px] font-extrabold text-slate-400 uppercase tracking-wider">
                  <th class="px-5 py-3.5 rounded-l-2xl">Produk</th>
                  <th class="px-5 py-3.5 text-center">Stok Saat Ini</th>
                  <th class="px-5 py-3.5 text-center">Limit Minimum</th>
                  <th class="px-5 py-3.5 text-center">Defisit</th>
                  <th class="px-5 py-3.5">Kondisi Buffer</th>
                  <th class="px-5 py-3.5 text-right rounded-r-2xl">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="item in (overviewData.low_stock_items || [])"
                  :key="item.product_id"
                  class="group hover:bg-slate-50/80 transition-colors"
                >
                  <td class="px-5 py-4">
                    <p class="font-extrabold text-slate-900 text-sm">{{ item.product_name }}</p>
                    <p class="text-xs text-slate-400 font-mono">SKU: {{ item.sku }}</p>
                  </td>
                  <td class="px-5 py-4 text-center font-black text-rose-600 text-sm">
                    {{ item.current_stock }} {{ item.unit }}
                  </td>
                  <td class="px-5 py-4 text-center font-bold text-slate-600 text-sm">
                    {{ item.min_limit }} {{ item.unit }}
                  </td>
                  <td class="px-5 py-4 text-center font-black text-rose-700 text-sm">
                    -{{ item.deficit }} {{ item.unit }}
                  </td>
                  <td class="px-5 py-4">
                    <div class="space-y-1">
                      <div class="flex items-center justify-between text-[11px] font-bold">
                        <span class="text-slate-400">Sisa Stok</span>
                        <span class="text-rose-600">{{ ((item.stock_pct ?? item.stockPct) || 0).toFixed(0) }}%</span>
                      </div>
                      <div class="w-32 bg-slate-100 h-2 rounded-full overflow-hidden">
                        <div
                          class="h-full rounded-full transition-all duration-500"
                          :class="((item.stock_pct ?? item.stockPct) || 0) < 50 ? 'bg-rose-500' : 'bg-amber-500'"
                          :style="{ width: `${Math.min((item.stock_pct ?? item.stockPct) || 0, 100)}%` }"
                        ></div>
                      </div>
                    </div>
                  </td>
                  <td class="px-5 py-4 text-right">
                    <router-link
                      to="/warehouse/stock"
                      class="inline-flex items-center gap-1 bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 px-3.5 py-1.5 rounded-xl text-xs font-black transition-all active:scale-95"
                    >
                      Restock <ArrowRight class="w-3.5 h-3.5" />
                    </router-link>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- TAB 3: Recent Mutation Logs Feed -->
        <div v-if="activeTab === 'logs'">
          <div v-if="(overviewData.recent_logs || []).length === 0" class="py-14 text-center text-slate-400">
            <History class="w-12 h-12 mx-auto mb-3 opacity-20" />
            <p class="text-sm font-bold">Belum ada mutasi tercatat pada periode ini.</p>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="bg-slate-50 text-[11px] font-extrabold text-slate-400 uppercase tracking-wider">
                  <th class="px-5 py-3.5 rounded-l-2xl">Produk</th>
                  <th class="px-5 py-3.5">Tipe & Kuantitas</th>
                  <th class="px-5 py-3.5">Sumber / Tujuan</th>
                  <th class="px-5 py-3.5">Petugas</th>
                  <th class="px-5 py-3.5 text-right rounded-r-2xl">Waktu</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="log in (overviewData.recent_logs || [])"
                  :key="log.id"
                  class="group hover:bg-slate-50/80 transition-colors"
                >
                  <td class="px-5 py-4">
                    <div class="flex items-center gap-3">
                      <div
                        class="p-2 rounded-xl shrink-0"
                        :class="log.type === 'IN' ? 'bg-emerald-50 text-emerald-600' : 'bg-amber-50 text-amber-600'"
                      >
                        <ArrowDownLeft v-if="log.type === 'IN'" class="h-4 w-4" />
                        <ArrowUpRight v-else class="h-4 w-4" />
                      </div>
                      <div>
                        <p class="font-extrabold text-slate-900 text-sm">{{ log.product_name }}</p>
                        <p class="text-xs text-slate-400 font-mono">{{ log.sku }}</p>
                      </div>
                    </div>
                  </td>
                  <td class="px-5 py-4">
                    <span
                      class="px-3 py-1 rounded-xl text-xs font-black inline-flex items-center gap-1"
                      :class="
                        log.type === 'IN'
                          ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                          : 'bg-amber-50 text-amber-700 border border-amber-200'
                      "
                    >
                      {{ log.type === 'IN' ? '+' : '-' }}{{ log.quantity.toLocaleString('id-ID') }} Unit
                    </span>
                  </td>
                  <td class="px-5 py-4 text-xs font-bold text-slate-700">
                    {{ log.source }}
                  </td>
                  <td class="px-5 py-4 text-xs text-slate-500">
                    <span class="font-bold text-slate-700">{{ log.processed_by }}</span>
                    <span v-if="log.approved_by" class="text-slate-400"> (Aprv: {{ log.approved_by }})</span>
                  </td>
                  <td class="px-5 py-4 text-right">
                    <p class="text-xs font-bold text-slate-900">{{ log.created_at }}</p>
                    <p class="text-[10px] text-slate-400 font-semibold">{{ log.time_ago }}</p>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- TAB 4: Full Inventory Catalog with Live Search -->
        <div v-if="activeTab === 'inventory'">
          <div v-if="filteredGroupedInventory.length === 0" class="py-14 text-center text-slate-400">
            <Package class="w-12 h-12 mx-auto mb-3 opacity-20" />
            <p class="text-sm font-bold">Tidak ada produk yang sesuai dengan pencarian.</p>
          </div>

          <div v-else class="overflow-x-auto max-h-[500px]">
            <table class="w-full text-left border-collapse">
              <thead class="sticky top-0 bg-white z-10 shadow-xs">
                <tr class="bg-slate-50 text-[11px] font-extrabold text-slate-400 uppercase tracking-wider">
                  <th class="px-5 py-3.5 rounded-l-2xl">Produk & SKU</th>
                  <th class="px-5 py-3.5 text-center">Batas Limit</th>
                  <th class="px-5 py-3.5 text-center">Stok Fisik</th>
                  <th class="px-5 py-3.5 text-right rounded-r-2xl">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="stock in filteredGroupedInventory"
                  :key="stock.product_id"
                  class="group hover:bg-slate-50/80 transition-colors"
                >
                  <td class="px-5 py-4">
                    <p class="font-extrabold text-slate-900 text-sm">{{ stock.product?.name }}</p>
                    <p class="text-xs text-slate-400 font-mono">SKU: {{ stock.product?.sku }}</p>
                  </td>
                  <td class="px-5 py-4 text-center font-bold text-slate-500 text-xs">
                    {{ stock.min_limit }} {{ stock.product?.unit || 'PCS' }}
                  </td>
                  <td class="px-5 py-4 text-center">
                    <span
                      class="px-3.5 py-1.5 rounded-xl text-xs font-black inline-block"
                      :class="
                        stock.quantity < stock.min_limit
                          ? 'bg-rose-50 text-rose-700 border border-rose-200'
                          : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                      "
                    >
                      {{ stock.quantity.toLocaleString('id-ID') }} {{ stock.product?.unit || 'PCS' }}
                    </span>
                  </td>
                  <td class="px-5 py-4 text-right">
                    <span
                      class="text-xs font-bold"
                      :class="stock.quantity < stock.min_limit ? 'text-rose-600 font-black' : 'text-emerald-600'"
                    >
                      {{ stock.quantity < stock.min_limit ? 'Menipis' : 'Aman' }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed, ref, watch } from 'vue'
import { useWarehouseStore } from '@/stores/warehouse'
import { useMasterDataStore } from '@/stores/masterData'
import apiClient from '@/api/axios'
import { toast } from 'vue-sonner'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import {
  Warehouse,
  Boxes,
  Package,
  Truck,
  AlertTriangle,
  History,
  ArrowUpRight,
  ArrowDownLeft,
  Calendar,
  ArrowLeftRight,
  ArrowRight,
  RefreshCw,
  Sparkles,
  CheckCircle2,
  ChevronRight,
  Search,
  Activity,
} from 'lucide-vue-next'

use([
  CanvasRenderer,
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
])

// --- Interfaces ---
interface WarehouseDashboardSummary {
  total_stock_physical: number
  total_sku_count: number
  low_stock_count: number
  out_of_stock_count: number
  pending_inbound_count: number
  pending_sales_transfers_count: number
  today_in_quantity: number
  today_out_quantity: number
}

interface MutationTrendPoint {
  date: string
  label: string
  in_qty: number
  out_qty: number
  net_change: number
}

interface TopStockProduct {
  id: string
  name: string
  sku: string
  unit: string
  quantity: number
  percentage: number
}

interface StockHealthSummary {
  safe_count: number
  warning_count: number
  critical_count: number
  out_of_stock_count: number
}

interface LowStockItem {
  product_id: string
  product_name: string
  sku: string
  unit: string
  current_stock: number
  min_limit: number
  deficit: number
  stock_pct?: number
  stockPct?: number
}

interface PendingInboundShipment {
  id: string
  delivery_order_no: string
  from_factory: string
  to_branch: string
  product_name: string
  quantity: number
  unit: string
  status: string
  vehicle_plate: string
  driver_name: string
  created_at: string
  target_date: string
}

interface RecentLogItem {
  id: string
  product_name: string
  sku: string
  type: string
  quantity: number
  source: string
  processed_by: string
  approved_by: string
  created_at: string
  time_ago: string
}

interface WarehouseOverviewResponse {
  summary: WarehouseDashboardSummary
  mutation_trend: MutationTrendPoint[]
  top_products: TopStockProduct[]
  stock_health: StockHealthSummary
  low_stock_items: LowStockItem[]
  pending_inbound_shipments: PendingInboundShipment[]
  recent_logs: RecentLogItem[]
}

// Stores
const warehouseStore = useWarehouseStore()
const masterDataStore = useMasterDataStore()

// State
const timeRange = ref<'7d' | '30d' | 'year'>('7d')
const isLoading = ref(true)
const activeTab = ref<'inbound' | 'low_stock' | 'logs' | 'inventory'>('inbound')
const inventorySearchQuery = ref('')

const overviewData = ref<WarehouseOverviewResponse>({
  summary: {
    total_stock_physical: 0,
    total_sku_count: 0,
    low_stock_count: 0,
    out_of_stock_count: 0,
    pending_inbound_count: 0,
    pending_sales_transfers_count: 0,
    today_in_quantity: 0,
    today_out_quantity: 0,
  },
  mutation_trend: [],
  top_products: [],
  stock_health: {
    safe_count: 0,
    warning_count: 0,
    critical_count: 0,
    out_of_stock_count: 0,
  },
  low_stock_items: [],
  pending_inbound_shipments: [],
  recent_logs: [],
})

// Current Branch & Date
const currentBranchName = computed(() => {
  const id = masterDataStore.selectedWarehouseBranchId
  if (!id || id === 'ALL_BRANCHES') return 'Semua Cabang'
  const branch = masterDataStore.branches.find((b) => b.id === id)
  return branch ? branch.name : 'Semua Cabang'
})

const currentFormattedDate = computed(() => {
  return new Intl.DateTimeFormat('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date())
})

// ECharts Options
const mutationTrendOption = computed(() => {
  const trend = overviewData.value.mutation_trend || []
  const categories = trend.map((t) => t.label)
  const inData = trend.map((t) => t.in_qty)
  const outData = trend.map((t) => t.out_qty)

  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      borderColor: '#334155',
      borderWidth: 1,
      borderRadius: 16,
      padding: [12, 16],
      textStyle: { color: '#f8fafc', fontFamily: 'Plus Jakarta Sans', fontSize: 12 },
    },
    legend: {
      data: ['Barang Masuk (IN)', 'Barang Keluar (OUT)'],
      bottom: 0,
      icon: 'circle',
      textStyle: { fontFamily: 'Plus Jakarta Sans', fontWeight: '600', color: '#64748b', fontSize: 12 },
      itemGap: 24,
    },
    grid: { left: '2%', right: '2%', bottom: '12%', top: '8%', containLabel: true },
    xAxis: {
      type: 'category',
      data: categories,
      axisLine: { lineStyle: { color: '#e2e8f0' } },
      axisTick: { show: false },
      axisLabel: { color: '#64748b', fontWeight: '600', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      name: 'Jumlah Unit (PCS)',
      nameTextStyle: { color: '#94a3b8', fontSize: 11, padding: [0, 0, 4, 0] },
      splitLine: { lineStyle: { color: '#f1f5f9', type: 'dashed' } },
      axisLabel: { color: '#64748b', fontWeight: '600' },
    },
    series: [
      {
        name: 'Barang Masuk (IN)',
        type: 'line',
        smooth: true,
        data: inData,
        itemStyle: { color: '#10b981' },
        lineStyle: { width: 3, color: '#10b981' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(16, 185, 129, 0.35)' },
              { offset: 1, color: 'rgba(16, 185, 129, 0.02)' },
            ],
          },
        },
      },
      {
        name: 'Barang Keluar (OUT)',
        type: 'line',
        smooth: true,
        data: outData,
        itemStyle: { color: '#f59e0b' },
        lineStyle: { width: 3, color: '#f59e0b' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(245, 158, 11, 0.35)' },
              { offset: 1, color: 'rgba(245, 158, 11, 0.02)' },
            ],
          },
        },
      },
    ],
  }
})

const topStockDistributionOption = computed(() => {
  const top = overviewData.value.top_products || []
  const total = overviewData.value.summary.total_stock_physical || 0

  const colors = [
    '#3b82f6', // blue
    '#10b981', // emerald
    '#8b5cf6', // purple
    '#f59e0b', // amber
    '#06b6d4', // cyan
    '#ec4899', // pink
    '#f97316', // orange
    '#64748b', // slate
  ]

  const chartData = top.map((p, idx) => ({
    value: p.quantity,
    name: p.name,
    itemStyle: { color: colors[idx % colors.length] },
  }))

  return {
    tooltip: {
      trigger: 'item',
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      borderColor: '#334155',
      borderRadius: 14,
      padding: [10, 14],
      textStyle: { color: '#f8fafc', fontFamily: 'Plus Jakarta Sans', fontSize: 12 },
      formatter: '{b}: <b style="color:#38bdf8">{c} unit</b> ({d}%)',
    },
    title: {
      text: total.toLocaleString('id-ID'),
      subtext: 'Total Unit',
      left: 'center',
      top: '38%',
      textStyle: { fontSize: 26, fontWeight: 'bold', color: '#0f172a', fontFamily: 'Plus Jakarta Sans' },
      subtextStyle: { fontSize: 12, color: '#94a3b8', fontWeight: '600' },
    },
    legend: {
      bottom: '0',
      left: 'center',
      icon: 'circle',
      textStyle: { fontFamily: 'Plus Jakarta Sans', fontSize: 11, color: '#64748b' },
      itemGap: 12,
    },
    series: [
      {
        name: 'Top Produk',
        type: 'pie',
        radius: ['52%', '76%'],
        center: ['50%', '45%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 8, borderColor: '#ffffff', borderWidth: 3 },
        label: { show: false },
        emphasis: { label: { show: false }, scale: true, scaleSize: 8 },
        data: chartData,
      },
    ],
  }
})

// Grouped Inventory & Search
const groupedInventory = computed(() => {
  const map = new Map<string, any>()
  warehouseStore.inventory.forEach((item) => {
    if (map.has(item.product_id)) {
      const existing = map.get(item.product_id)
      existing.quantity += item.quantity
    } else {
      map.set(item.product_id, { ...item })
    }
  })

  return Array.from(map.values()).sort((a, b) => {
    if (a.product?.name && b.product?.name) {
      return a.product.name.localeCompare(b.product.name)
    }
    return 0
  })
})

const filteredGroupedInventory = computed(() => {
  if (!inventorySearchQuery.value.trim()) return groupedInventory.value
  const q = inventorySearchQuery.value.toLowerCase()
  return groupedInventory.value.filter(
    (item) =>
      item.product?.name?.toLowerCase().includes(q) ||
      item.product?.sku?.toLowerCase().includes(q),
  )
})

// Data Fetching
const fetchDashboardData = async () => {
  isLoading.value = true
  try {
    let url = `/warehouse/dashboard/overview?range=${timeRange.value}`
    if (masterDataStore.selectedWarehouseBranchId && masterDataStore.selectedWarehouseBranchId !== 'ALL_BRANCHES') {
      url += `&branch_id=${masterDataStore.selectedWarehouseBranchId}`
    }
    const res = await apiClient.get(url)
    if (res.data?.data) {
      const d = res.data.data
      overviewData.value = {
        summary: {
          total_stock_physical: d.summary?.total_stock_physical || 0,
          total_sku_count: d.summary?.total_sku_count || 0,
          low_stock_count: d.summary?.low_stock_count || 0,
          out_of_stock_count: d.summary?.out_of_stock_count || 0,
          pending_inbound_count: d.summary?.pending_inbound_count || 0,
          pending_sales_transfers_count: d.summary?.pending_sales_transfers_count || 0,
          today_in_quantity: d.summary?.today_in_quantity || 0,
          today_out_quantity: d.summary?.today_out_quantity || 0,
        },
        mutation_trend: d.mutation_trend || [],
        top_products: d.top_products || [],
        stock_health: {
          safe_count: d.stock_health?.safe_count || 0,
          warning_count: d.stock_health?.warning_count || 0,
          critical_count: d.stock_health?.critical_count || 0,
          out_of_stock_count: d.stock_health?.out_of_stock_count || 0,
        },
        low_stock_items: (d.low_stock_items || []).map((item: any) => ({
          ...item,
          stockPct: item.stock_pct ?? item.stockPct ?? 0,
          stock_pct: item.stock_pct ?? item.stockPct ?? 0,
        })),
        pending_inbound_shipments: d.pending_inbound_shipments || [],
        recent_logs: d.recent_logs || [],
      }
    }
    // Also sync inventory store for catalog tab
    await warehouseStore.fetchInventory()
  } catch (e: any) {
    console.error('Failed to fetch warehouse dashboard data', e)
    toast.error('Gagal memuat data dashboard gudang')
  } finally {
    isLoading.value = false
  }
}

// Watchers
watch(
  () => masterDataStore.selectedWarehouseBranchId,
  () => {
    fetchDashboardData()
  },
)

watch(timeRange, () => {
  fetchDashboardData()
})

onMounted(async () => {
  await masterDataStore.fetchBranches()
  fetchDashboardData()
})
</script>

<style scoped>
.v-chart {
  font-family: 'Plus Jakarta Sans', sans-serif !important;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fadeIn 0.35s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
