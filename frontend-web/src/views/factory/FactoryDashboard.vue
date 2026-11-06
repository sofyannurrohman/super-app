<template>
  <div class="p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-3xl font-extrabold tracking-tight text-slate-900">Dashboard Pabrik</h1>
        <p class="text-slate-500 mt-1">
          Pantau produksi, stok, dan pengiriman barang dari seluruh pabrik Karsa HR.
        </p>
      </div>
      <div class="flex gap-3">
        <Button
          variant="outline"
          class="rounded-xl bg-white border-slate-200 text-slate-600 hover:bg-slate-50"
        >
          <Download class="mr-2 h-4 w-4" /> Export Report
        </Button>
        <Button
          @click="showAddFactory = true"
          class="rounded-xl shadow-lg shadow-primary/20 transition-all hover:scale-105"
        >
          <Plus class="mr-2 h-4 w-4" /> Tambah Pabrik
        </Button>
      </div>
    </div>

    <!-- KPI Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      <Card
        @click="router.push({ name: 'factoryManagement' })"
        class="border-none shadow-sm bg-white overflow-hidden relative group cursor-pointer hover:shadow-md transition-all"
      >
        <div
          class="absolute right-0 top-0 p-3 opacity-10 group-hover:scale-110 transition-transform"
        >
          <Factory class="h-16 w-16 text-primary" />
        </div>
        <CardContent class="p-6">
          <p class="text-sm font-medium text-slate-500 uppercase tracking-wider">Total Pabrik</p>
          <h3 class="text-3xl font-bold text-slate-900 mt-2">
            {{ factoryStore.dashboardStats.total_factories }}
          </h3>
          <div class="flex items-center mt-4 text-emerald-600 text-sm font-medium">
            <span class="bg-emerald-50 px-2 py-0.5 rounded-full">Operasional Aktif</span>
          </div>
        </CardContent>
      </Card>

      <Card
        @click="router.push({ name: 'factoryProducts' })"
        class="border-none shadow-sm bg-white overflow-hidden relative group cursor-pointer hover:shadow-md transition-all"
      >
        <div
          class="absolute right-0 top-0 p-3 opacity-10 group-hover:scale-110 transition-transform"
        >
          <Package class="h-16 w-16 text-blue-600" />
        </div>
        <CardContent class="p-6">
          <p class="text-sm font-medium text-slate-500 uppercase tracking-wider">Total Produk</p>
          <h3 class="text-3xl font-bold text-slate-900 mt-2">
            {{ factoryStore.dashboardStats.total_products }}
          </h3>
          <div class="flex items-center mt-4 text-slate-500 text-sm">
            <span>Master data produk aktif</span>
          </div>
        </CardContent>
      </Card>

      <Card
        @click="router.push({ name: 'factoryTransfersGlobal' })"
        class="border-none shadow-sm bg-white overflow-hidden relative group cursor-pointer hover:shadow-md transition-all"
      >
        <div
          class="absolute right-0 top-0 p-3 opacity-10 group-hover:scale-110 transition-transform"
        >
          <Truck class="h-16 w-16 text-orange-600" />
        </div>
        <CardContent class="p-6">
          <p class="text-sm font-medium text-slate-500 uppercase tracking-wider">
            Pengiriman Pending
          </p>
          <h3 class="text-3xl font-bold text-slate-900 mt-2">
            {{ factoryStore.dashboardStats.pending_shipments }}
          </h3>
          <div class="flex items-center mt-4 text-orange-600 text-sm font-medium">
            <AlertCircle class="mr-1 h-3.5 w-3.5" />
            <span>Butuh konfirmasi gudang</span>
          </div>
        </CardContent>
      </Card>

      <Card
        @click="router.push({ name: 'factoryStockGlobal' })"
        class="border-none shadow-sm bg-white overflow-hidden relative group cursor-pointer hover:shadow-md transition-all"
      >
        <div
          class="absolute right-0 top-0 p-3 opacity-10 group-hover:scale-110 transition-transform"
        >
          <Activity class="h-16 w-16 text-indigo-600" />
        </div>
        <CardContent class="p-6">
          <p class="text-sm font-medium text-slate-500 uppercase tracking-wider">
            Produksi Hari Ini
          </p>
          <h3 class="text-3xl font-bold text-slate-900 mt-2">
            {{ factoryStore.dashboardStats.today_production }}
            <span class="text-sm font-normal text-slate-400">pcs</span>
          </h3>
          <div class="flex items-center mt-4 text-emerald-600 text-sm font-medium">
            <TrendingUp class="mr-1 h-3.5 w-3.5" />
            <span>8% vs kemarin</span>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Main Content Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Factory List (2/3 width) -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Section: Branch Transfer Requests -->
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-bold text-slate-900 flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-orange-500 animate-pulse"></span>
              Permintaan Transfer dari Cabang
            </h2>
            <Badge
              variant="outline"
              class="bg-orange-50 text-orange-600 border-orange-100 rounded-full px-3 py-1 font-bold text-[10px]"
            >
              {{ branchRequests.length }} Pending
            </Badge>
          </div>

          <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
            <CardContent class="p-0">
              <div v-if="branchRequests.length === 0" class="p-12 text-center">
                <div
                  class="w-16 h-16 bg-slate-50 rounded-full flex items-center justify-center mx-auto mb-4"
                >
                  <ClipboardList class="h-8 w-8 text-slate-400" />
                </div>
                <h3 class="text-base font-bold text-slate-900">Tidak Ada Permintaan Baru</h3>
                <p class="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
                  Semua permintaan transfer dari cabang telah diproses.
                </p>
              </div>

              <div v-else class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                  <thead>
                    <tr
                      class="bg-slate-50/75 border-b border-slate-100 text-[10px] font-bold text-slate-400 uppercase tracking-wider"
                    >
                      <th class="py-4 px-6">No. Surat Jalan</th>
                      <th class="py-4 px-6">Cabang</th>
                      <th class="py-4 px-6">Item / Produk</th>
                      <th class="py-4 px-6">Tanggal Request</th>
                      <th class="py-4 px-6 text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-50 text-sm">
                    <tr
                      v-for="req in branchRequests"
                      :key="req.delivery_order_no"
                      class="hover:bg-slate-50/50 transition-colors"
                    >
                      <td class="py-4 px-6 font-mono text-xs font-bold text-slate-700">
                        {{ req.delivery_order_no }}
                      </td>
                      <td class="py-4 px-6">
                        <span class="font-extrabold text-slate-900">{{ req.to_branch_name }}</span>
                      </td>
                      <td class="py-4 px-6">
                        <div class="flex flex-col gap-1 max-w-[280px]">
                          <span class="text-xs font-semibold text-slate-700 truncate">
                            {{
                              req.items
                                .map((i) => `${i.product_name} (${i.quantity} ${i.unit})`)
                                .join(', ')
                            }}
                          </span>
                          <span
                            class="text-[10px] text-slate-400 truncate animate-pulse"
                            v-if="req.notes"
                          >
                            Catatan: {{ req.notes }}
                          </span>
                        </div>
                      </td>
                      <td class="py-4 px-6 text-slate-500 text-xs">
                        {{ formatDateTime(req.created_at) }}
                      </td>
                      <td class="py-4 px-6 text-right">
                        <div class="flex items-center justify-end gap-2">
                          <Button
                            size="sm"
                            variant="outline"
                            class="rounded-xl border-slate-200 text-slate-600 hover:bg-slate-50 font-bold text-xs"
                            @click="openRequestDetail(req)"
                          >
                            <Eye class="h-3.5 w-3.5 mr-1" /> Detail
                          </Button>
                          <Button
                            size="sm"
                            class="rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs shadow-md shadow-emerald-600/10 transition-all hover:scale-105"
                            @click="handleApproveRequest(req)"
                          >
                            <Check class="h-3.5 w-3.5 mr-1" /> Setujui
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            class="rounded-xl font-bold text-xs shadow-md transition-all hover:scale-105"
                            @click="openRejectRequest(req)"
                          >
                            <X class="h-3.5 w-3.5 mr-1" /> Tolak
                          </Button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </div>

        <!-- Section: Branch Stock Alerts -->
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-bold text-slate-900 flex items-center gap-2">
              <span class="w-2.5 h-2.5 rounded-full bg-rose-600 animate-pulse"></span>
              Peringatan Stok Cabang Rendah
            </h2>
            <Badge
              variant="outline"
              class="bg-rose-50 text-rose-600 border-rose-100 rounded-full px-3 py-1 font-bold text-[10px]"
            >
              {{ factoryStore.lowStockAlerts.length }} Kritis
            </Badge>
          </div>

          <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
            <CardContent class="p-0">
              <div v-if="factoryStore.lowStockAlerts.length === 0" class="p-12 text-center">
                <div
                  class="w-16 h-16 bg-emerald-50 rounded-full flex items-center justify-center mx-auto mb-4"
                >
                  <CheckCircle2 class="h-8 w-8 text-emerald-600" />
                </div>
                <h3 class="text-base font-bold text-slate-900">Semua Cabang Aman</h3>
                <p class="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
                  Tidak ada produk di bawah limit threshold di seluruh cabang saat ini.
                </p>
              </div>

              <div v-else class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                  <thead>
                    <tr
                      class="bg-slate-50/75 border-b border-slate-100 text-[10px] font-bold text-slate-400 uppercase tracking-wider"
                    >
                      <th class="py-4 px-6">Cabang</th>
                      <th class="py-4 px-6">Produk</th>
                      <th class="py-4 px-6 text-center">Stok</th>
                      <th class="py-4 px-6 text-center">Threshold</th>
                      <th class="py-4 px-6 text-right">Aksi</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-50 text-sm">
                    <tr
                      v-for="alert in factoryStore.lowStockAlerts"
                      :key="alert.warehouse_stock_id"
                      class="hover:bg-slate-50/50 transition-colors"
                    >
                      <td class="py-4 px-6">
                        <span class="font-extrabold text-slate-900">{{ alert.branch_name }}</span>
                      </td>
                      <td class="py-4 px-6">
                        <div class="flex flex-col">
                          <span class="font-extrabold text-slate-900">{{
                            alert.product_name
                          }}</span>
                          <span class="text-[10px] text-slate-400 font-mono"
                            >SKU: {{ alert.product_sku }} | Batch: {{ alert.batch_no }}</span
                          >
                        </div>
                      </td>
                      <td class="py-4 px-6 text-center">
                        <Badge
                          class="bg-rose-50 hover:bg-rose-100 text-rose-600 border border-rose-100 font-extrabold rounded-lg px-2.5 py-1"
                        >
                          {{ alert.quantity }} PCS
                        </Badge>
                      </td>
                      <td class="py-4 px-6 text-center text-slate-500 font-bold">
                        {{ alert.min_limit }} PCS
                      </td>
                      <td class="py-4 px-6 text-right">
                        <Button
                          size="sm"
                          class="rounded-xl bg-primary hover:bg-primary/95 text-white font-bold text-xs shadow-md shadow-primary/10 transition-all hover:scale-105"
                          @click="openProactiveRestock(alert)"
                        >
                          <Truck class="h-3.5 w-3.5 mr-1.5" /> Kirim Restock
                        </Button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </div>

        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold text-slate-900">Lokasi Pabrik</h2>
          <Button variant="link" class="text-primary font-semibold">Lihat Semua</Button>
        </div>

        <div v-if="factoryStore.loading" class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div v-for="i in 4" :key="i" class="h-48 bg-slate-100 animate-pulse rounded-2xl"></div>
        </div>

        <div
          v-else-if="factoryStore.factories.length === 0"
          class="flex flex-col items-center justify-center py-20 bg-white rounded-3xl border-2 border-dashed border-slate-200 text-center px-6"
        >
          <div class="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center mb-4">
            <Factory class="h-10 w-10 text-slate-300" />
          </div>
          <h3 class="text-lg font-bold text-slate-900">Belum Ada Pabrik</h3>
          <p class="text-slate-500 max-w-xs mt-2">
            Daftarkan lokasi pabrik baru Anda untuk mulai mengelola inventory dan produksi.
          </p>
          <Button @click="showAddFactory = true" class="mt-6 rounded-xl px-8"
            >Tambah Sekarang</Button
          >
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div
            v-for="factory in factoryStore.factories"
            :key="factory.id"
            class="group bg-white rounded-2xl p-6 shadow-sm border border-slate-100 hover:border-primary/20 hover:shadow-xl hover:shadow-primary/5 transition-all duration-300 cursor-pointer"
            @click="goToFactory(factory.id)"
          >
            <div class="flex justify-between items-start mb-4">
              <div
                class="w-12 h-12 bg-slate-50 rounded-xl flex items-center justify-center text-primary group-hover:bg-primary group-hover:text-white transition-colors"
              >
                <Factory class="h-6 w-6" />
              </div>
              <Badge
                variant="secondary"
                class="bg-blue-50 text-blue-600 border-none rounded-full px-3 py-1"
              >
                {{ factory.branch?.name || 'No Branch' }}
              </Badge>
            </div>

            <h3 class="text-lg font-bold text-slate-900 group-hover:text-primary transition-colors">
              {{ factory.name }}
            </h3>
            <p class="text-sm text-slate-500 mt-1 flex items-center">
              <MapPin class="mr-1 h-3.5 w-3.5" /> {{ factory.location }}
            </p>

            <div class="mt-6 pt-6 border-t border-slate-50 flex items-center justify-between">
              <div class="flex -space-x-2">
                <div
                  v-for="i in 3"
                  :key="i"
                  class="w-8 h-8 rounded-full bg-slate-100 border-2 border-white flex items-center justify-center"
                >
                  <Package class="h-3 w-3 text-slate-400" />
                </div>
                <div
                  class="w-8 h-8 rounded-full bg-slate-50 border-2 border-white flex items-center justify-center text-[10px] font-bold text-slate-400"
                >
                  +5
                </div>
              </div>
              <Button
                variant="ghost"
                size="sm"
                class="text-primary font-bold group-hover:translate-x-1 transition-transform"
              >
                Kelola <ChevronRight class="ml-1 h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>

        <!-- Section: Factory Stocks -->
        <div class="space-y-4 mt-8">
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-bold text-slate-900">Total Stok Produk di Pabrik</h2>
            <Badge
              variant="outline"
              class="bg-blue-50 text-blue-600 border-blue-100 rounded-full px-3 py-1 font-bold text-[10px]"
            >
              {{ factoryStore.allInventory.length }} BATCH
            </Badge>
          </div>

          <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
            <CardContent class="p-0">
              <div v-if="factoryStore.allInventory.length === 0" class="p-12 text-center">
                <div
                  class="w-16 h-16 bg-slate-50 rounded-full flex items-center justify-center mx-auto mb-4"
                >
                  <Package class="h-8 w-8 text-slate-400" />
                </div>
                <h3 class="text-base font-bold text-slate-900">Stok Kosong</h3>
                <p class="text-xs text-slate-500 mt-1 max-w-sm mx-auto">
                  Belum ada stok barang yang tercatat di pabrik.
                </p>
              </div>

              <div v-else class="overflow-x-auto max-h-[500px]">
                <table class="w-full text-left border-collapse">
                  <thead class="sticky top-0 bg-white z-10 shadow-sm">
                    <tr
                      class="bg-slate-50/90 backdrop-blur border-b border-slate-100 text-[10px] font-bold text-slate-400 uppercase tracking-wider"
                    >
                      <th class="py-4 px-6">Produk</th>
                      <th class="py-4 px-6 text-center">Pabrik</th>
                      <th class="py-4 px-6 text-center">Batch No</th>
                      <th class="py-4 px-6 text-center">Stok</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-50 text-sm">
                    <tr
                      v-for="stock in factoryStore.allInventory"
                      :key="stock.id"
                      class="hover:bg-slate-50/50 transition-colors"
                    >
                      <td class="py-4 px-6">
                        <div class="flex flex-col">
                          <span class="font-extrabold text-slate-900">{{
                            stock.product?.name
                          }}</span>
                          <span class="text-[10px] text-slate-400 font-mono"
                            >SKU: {{ stock.product?.sku }}</span
                          >
                        </div>
                      </td>
                      <td class="py-4 px-6 text-center font-bold text-slate-600">
                        {{ getFactoryName(stock.factory_id) }}
                      </td>
                      <td class="py-4 px-6 text-center text-slate-500 font-mono text-xs">
                        {{ stock.batch_no }}
                      </td>
                      <td class="py-4 px-6 text-center">
                        <Badge
                          class="bg-blue-50 hover:bg-blue-100 text-blue-600 border border-blue-100 font-extrabold rounded-lg px-2.5 py-1"
                        >
                          {{ stock.quantity }} {{ stock.product?.unit || 'PCS' }}
                        </Badge>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <!-- Demand Analysis / Backorder Queue (1/3 width) -->
      <div class="space-y-6">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold text-slate-900">Analisis Permintaan</h2>
          <Badge
            variant="outline"
            class="bg-red-50 text-red-600 border-red-100 px-3 py-1 rounded-full text-[10px] font-bold"
          >
            Real-time
          </Badge>
        </div>

        <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
          <CardHeader class="p-6 pb-2">
            <CardTitle class="text-sm font-bold text-slate-500 uppercase tracking-wider"
              >Antrean Backorder (FIFO)</CardTitle
            >
            <CardDescription class="text-xs"
              >Produk yang habis stok & ditunggu pelanggan.</CardDescription
            >
          </CardHeader>
          <CardContent class="p-0">
            <div v-if="factoryStore.backorderDemand.length === 0" class="p-8 text-center">
              <CheckCircle2 class="h-10 w-10 text-emerald-400 mx-auto mb-3" />
              <p class="text-sm font-medium text-slate-900">Semua Terpenuhi</p>
              <p class="text-xs text-slate-500 mt-1">Tidak ada antrean pesanan menggantung.</p>
            </div>
            <div v-else class="divide-y divide-slate-50">
              <div
                v-for="item in factoryStore.backorderDemand"
                :key="item.product_id"
                class="p-5 hover:bg-slate-50/50 transition-colors group"
              >
                <div class="flex items-center gap-4">
                  <div
                    class="w-12 h-12 rounded-2xl bg-red-50 flex flex-col items-center justify-center text-red-600 group-hover:bg-red-600 group-hover:text-white transition-all duration-300"
                  >
                    <span class="text-lg font-black leading-none">{{ item.total_qty }}</span>
                    <span class="text-[8px] font-bold uppercase mt-1">PCS</span>
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex justify-between items-start">
                      <p class="text-sm font-extrabold text-slate-900 truncate">
                        {{ item.product_name }}
                      </p>
                    </div>
                    <div class="mt-2 flex items-center gap-2">
                      <Badge
                        variant="secondary"
                        class="text-[9px] font-bold bg-slate-100 text-slate-600 border-none"
                      >
                        {{ item.order_count }} Invoice
                      </Badge>
                      <div class="h-1 flex-1 bg-slate-100 rounded-full overflow-hidden mt-0.5">
                        <div class="h-full bg-red-500 w-full animate-pulse"></div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <h2 class="text-xl font-bold text-slate-900">Pengiriman Terbaru</h2>

        <Card class="border-none shadow-sm bg-white rounded-3xl overflow-hidden">
          <CardContent class="p-0">
            <div
              v-if="factoryStore.dashboardStats.recent_transfers.length === 0"
              class="p-8 text-center text-slate-400 italic text-sm"
            >
              Belum ada riwayat pengiriman terbaru.
            </div>
            <div v-else class="divide-y divide-slate-50">
              <div
                v-for="transfer in factoryStore.dashboardStats.recent_transfers"
                :key="transfer.id"
                class="p-5 hover:bg-slate-50/50 transition-colors"
              >
                <div class="flex items-center gap-4">
                  <div
                    class="w-10 h-10 rounded-xl flex items-center justify-center"
                    :class="
                      transfer.status === 'RECEIVED'
                        ? 'bg-emerald-50 text-emerald-600'
                        : 'bg-orange-50 text-orange-600'
                    "
                  >
                    <Truck class="h-5 w-5" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex justify-between items-start">
                      <p class="text-sm font-bold text-slate-900 truncate">
                        {{ transfer.delivery_order_no }}
                      </p>
                      <span class="text-[10px] text-slate-400">{{
                        formatRelativeTime(transfer.created_at)
                      }}</span>
                    </div>
                    <p class="text-xs text-slate-500 mt-0.5">
                      {{ transfer.from_factory?.name }} &rarr; {{ transfer.to_branch?.name }}
                    </p>
                    <div class="mt-2 flex items-center justify-between">
                      <span
                        class="text-[10px] font-bold px-2 py-0.5 rounded-full uppercase"
                        :class="getStatusClass(transfer.status)"
                      >
                        {{ formatStatus(transfer.status) }}
                      </span>
                      <span class="text-[10px] font-medium text-slate-400"
                        >{{ transfer.total_weight.toFixed(2) }} KG</span
                      >
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div class="p-4 bg-slate-50 text-center">
              <Button
                variant="link"
                class="text-xs text-slate-500 font-bold uppercase tracking-wider"
                >Lihat Semua Logistik</Button
              >
            </div>
          </CardContent>
        </Card>
      </div>
    </div>

    <!-- Add Factory Modal -->
    <div
      v-if="showAddFactory"
      class="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-lg shadow-2xl animate-in zoom-in-95 duration-200 rounded-3xl overflow-hidden border-none"
      >
        <CardHeader class="bg-white border-b p-6">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-xl font-bold">Tambah Pabrik Baru</CardTitle>
              <CardDescription>Daftarkan lokasi operasional pabrik Karsa HR.</CardDescription>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showAddFactory = false"
              class="rounded-full"
            >
              <Plus class="h-5 w-5 rotate-45" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-6 space-y-6 bg-white">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Nama Pabrik</label>
            <div class="relative">
              <Factory class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <input
                v-model="newFactoryForm.name"
                type="text"
                placeholder="e.g. Pabrik Utama Karsa HR"
                class="w-full pl-10 pr-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:ring-2 focus:ring-primary/20 transition-all outline-none"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Lokasi / Alamat</label>
            <div class="relative">
              <MapPin class="absolute left-3 top-3 h-4 w-4 text-slate-400" />
              <textarea
                v-model="newFactoryForm.location"
                rows="3"
                placeholder="Alamat lengkap operasional..."
                class="w-full pl-10 pr-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:ring-2 focus:ring-primary/20 outline-none resize-none"
              ></textarea>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Cabang Terkait</label>
            <div class="relative">
              <select
                v-model="newFactoryForm.branch_id"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:ring-2 focus:ring-primary/20 outline-none appearance-none"
              >
                <option value="">Pilih Cabang (Opsional)</option>
                <option
                  v-for="branch in masterDataStore.branches"
                  :key="branch.id"
                  :value="branch.id"
                >
                  {{ branch.name }}
                </option>
              </select>
              <ChevronRight
                class="absolute right-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 rotate-90"
              />
            </div>
            <p class="text-[10px] text-slate-400">
              Pabrik biasanya mensuplai stok ke satu cabang utama.
            </p>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 p-6 flex justify-end gap-3">
          <Button variant="ghost" @click="showAddFactory = false" class="rounded-xl font-bold"
            >Batal</Button
          >
          <Button
            @click="handleAddFactory"
            :loading="submitting"
            class="rounded-xl px-8 shadow-lg shadow-primary/20 font-bold"
          >
            Simpan Pabrik
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Proactive Restock Modal -->
    <div
      v-if="showRestockModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900"
                >Restock Otomatis Cabang</CardTitle
              >
              <CardDescription
                >Proaktif mengirim pasokan produk di bawah threshold ke cabang.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showRestockModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[60vh] overflow-y-auto">
          <!-- Alert Details -->
          <div class="grid grid-cols-2 gap-4">
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-1">
              <p
                class="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none"
              >
                Cabang Tujuan
              </p>
              <p class="text-sm font-black text-slate-900">{{ selectedAlert?.branch_name }}</p>
            </div>
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-1">
              <p
                class="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none"
              >
                Produk
              </p>
              <p class="text-sm font-black text-slate-900">{{ selectedAlert?.product_name }}</p>
            </div>
          </div>

          <!-- Source Factory Selection -->
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Pilih Pabrik Pengirim</label>
            <select
              v-model="restockForm.from_factory_id"
              class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 outline-none font-bold text-sm bg-white"
            >
              <option value="" disabled>Pilih Pabrik...</option>
              <option v-for="fact in factoryStore.factories" :key="fact.id" :value="fact.id">
                {{ fact.name }} ({{ fact.location }})
              </option>
            </select>
          </div>

          <!-- Quantities Info -->
          <div class="grid grid-cols-3 gap-4">
            <div class="p-4 bg-rose-50/50 rounded-2xl border border-rose-100 text-center">
              <p class="text-[9px] font-black text-rose-500 uppercase tracking-widest">
                Stok Saat Ini
              </p>
              <p class="text-lg font-black text-rose-600 mt-1">{{ selectedAlert?.quantity }} PCS</p>
            </div>
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 text-center">
              <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">
                Limit Threshold
              </p>
              <p class="text-lg font-black text-slate-800 mt-1">
                {{ selectedAlert?.min_limit }} PCS
              </p>
            </div>
            <div class="p-4 bg-emerald-50 rounded-2xl border border-emerald-100 text-center">
              <p class="text-[9px] font-black text-emerald-600 uppercase tracking-widest">
                Saran Kirim
              </p>
              <p class="text-lg font-black text-emerald-700 mt-1">
                {{
                  Math.max((selectedAlert?.min_limit || 0) - (selectedAlert?.quantity || 0), 1)
                }}
                PCS
              </p>
            </div>
          </div>

          <!-- Quantity & Logistics Form -->
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Jumlah Kirim</label>
              <input
                type="number"
                v-model.number="restockForm.quantity"
                min="1"
                class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 outline-none font-bold"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Satuan Pengiriman</label>
              <select
                v-model="restockForm.unit"
                class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 outline-none font-bold text-sm bg-white"
              >
                <option v-for="unit in UNIT_OPTIONS" :key="unit" :value="unit">{{ unit }}</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Pilih Armada (Kendaraan)</label>
              <select
                v-model="restockForm.vehicle_id"
                class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 outline-none font-bold text-sm bg-white"
              >
                <option value="">Tanpa Armada (Ekspedisi Luar)</option>
                <option v-for="veh in vehicles" :key="veh.id" :value="veh.id">
                  {{ veh.name }} ({{ veh.license_plate }}) - Cap: {{ veh.capacity }} KG
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Pilih Driver</label>
              <select
                v-model="restockForm.driver_id"
                class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 outline-none font-bold text-sm bg-white"
              >
                <option value="">Tanpa Driver</option>
                <option v-for="drv in drivers" :key="drv.id" :value="drv.id">
                  {{ drv.first_name }} {{ drv.last_name }}
                </option>
              </select>
            </div>
          </div>

          <!-- Notes -->
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Catatan Pengiriman</label>
            <input
              type="text"
              v-model="restockForm.notes"
              class="w-full h-12 rounded-xl border border-slate-200 focus:border-primary px-4 outline-none font-bold text-sm"
              placeholder="Contoh: Pengiriman otomatis di bawah threshold..."
            />
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-8"
            @click="showRestockModal = false"
            >Batal</Button
          >
          <Button
            class="rounded-2xl px-12 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90"
            @click="submitProactiveRestock"
            :loading="isSubmittingRestock"
            >Kirim Sekarang</Button
          >
        </CardFooter>
      </Card>
    </div>

    <!-- Request Detail Modal -->
    <div
      v-if="showDetailModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-2xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900"
                >Detail Permintaan Cabang</CardTitle
              >
              <CardDescription
                >Detail list barang yang diminta oleh cabang untuk transfer stok.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showDetailModal = false"
              class="rounded-full"
            >
              <X class="h-6 w-6 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[60vh] overflow-y-auto">
          <!-- Request Header Info -->
          <div class="grid grid-cols-2 gap-4">
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-1">
              <p
                class="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none"
              >
                No. Surat Jalan
              </p>
              <p class="text-sm font-black text-slate-900 font-mono">
                {{ selectedRequest?.delivery_order_no }}
              </p>
            </div>
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-1">
              <p
                class="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none"
              >
                Cabang Peminta
              </p>
              <p class="text-sm font-black text-slate-900">{{ selectedRequest?.to_branch_name }}</p>
            </div>
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-1">
              <p
                class="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none"
              >
                Tanggal Request
              </p>
              <p class="text-sm font-bold text-slate-800">
                {{ formatDateTime(selectedRequest?.created_at) }}
              </p>
            </div>
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-1">
              <p
                class="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none"
              >
                Inisiator
              </p>
              <Badge
                variant="secondary"
                class="bg-blue-50 text-blue-600 border-none rounded-full px-2.5 py-0.5 mt-0.5 text-[10px] font-bold uppercase w-fit"
              >
                {{ selectedRequest?.initiated_by }}
              </Badge>
            </div>
          </div>

          <!-- Notes -->
          <div
            v-if="selectedRequest?.notes"
            class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-1"
          >
            <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest leading-none">
              Catatan dari Cabang
            </p>
            <p class="text-sm text-slate-700 italic">"{{ selectedRequest?.notes }}"</p>
          </div>

          <!-- Items Table -->
          <div class="space-y-3">
            <h3 class="text-sm font-bold text-slate-800 ml-1">List Barang yang Diminta</h3>
            <div class="border border-slate-100 rounded-2xl overflow-hidden bg-slate-50/50">
              <table class="w-full text-left border-collapse text-sm">
                <thead>
                  <tr
                    class="bg-slate-50 text-[10px] font-bold text-slate-400 uppercase tracking-wider border-b border-slate-100"
                  >
                    <th class="p-4">Produk</th>
                    <th class="p-4 text-center">Batch No</th>
                    <th class="p-4 text-right">Jumlah</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr
                    v-for="item in selectedRequest?.items"
                    :key="item.id"
                    class="hover:bg-slate-50/75 transition-colors"
                  >
                    <td class="p-4">
                      <div class="flex flex-col">
                        <span class="font-extrabold text-slate-900">{{ item.product_name }}</span>
                        <span class="text-[10px] text-slate-400 font-mono"
                          >SKU: {{ item.product_sku }}</span
                        >
                      </div>
                    </td>
                    <td class="p-4 text-center font-mono text-xs text-slate-500">
                      {{ item.batch_no }}
                    </td>
                    <td class="p-4 text-right font-extrabold text-slate-900">
                      {{ item.quantity }} {{ item.unit }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-3">
          <Button
            variant="ghost"
            class="rounded-2xl font-bold h-12 px-6"
            @click="showDetailModal = false"
            >Tutup</Button
          >
          <Button
            variant="destructive"
            class="rounded-2xl font-bold h-12 px-6 shadow-md transition-all hover:scale-105"
            @click="openRejectRequest(selectedRequest)"
          >
            <X class="h-4 w-4 mr-1.5" /> Tolak Permintaan
          </Button>
          <Button
            class="rounded-2xl px-8 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90 transition-all hover:scale-105"
            @click="handleApproveRequest(selectedRequest)"
            :loading="isProcessingApproval"
          >
            <Check class="h-4 w-4 mr-1.5" /> Setujui Permintaan
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Reject Request Modal -->
    <div
      v-if="showRejectModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[110] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-md shadow-2xl rounded-[2rem] overflow-hidden border-none animate-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-6 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-xl font-bold text-slate-955">Tolak Permintaan</CardTitle>
              <CardDescription
                >Masukkan alasan penolakan permintaan transfer barang.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showRejectModal = false"
              class="rounded-full"
            >
              <X class="h-5 w-5 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-6 space-y-4 bg-white">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Alasan Penolakan</label>
            <textarea
              v-model="rejectionReason"
              rows="4"
              placeholder="Contoh: Stok di pabrik sedang kritis untuk produk ini..."
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 focus:border-red-500 rounded-2xl text-sm outline-none resize-none font-medium"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-6 flex justify-end gap-3">
          <Button variant="ghost" class="rounded-xl font-bold" @click="showRejectModal = false"
            >Batal</Button
          >
          <Button
            variant="destructive"
            class="rounded-xl px-6 font-bold"
            @click="handleRejectRequest"
            :loading="isProcessingRejection"
          >
            Tolak Sekarang
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useFactoryStore } from '@/stores/factory'
import { useMasterDataStore } from '@/stores/masterData'
import { useVehicleStore } from '@/stores/vehicle'
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
import { toast } from 'vue-sonner'
import {
  Plus,
  Factory,
  Package,
  Truck,
  Activity,
  AlertCircle,
  TrendingUp,
  Download,
  MapPin,
  ChevronRight,
  HelpCircle,
  CheckCircle2,
  Check,
  X,
  Eye,
  ClipboardList,
} from 'lucide-vue-next'

const router = useRouter()
const factoryStore = useFactoryStore()
const masterDataStore = useMasterDataStore()
const vehicleStore = useVehicleStore()

const showAddFactory = ref(false)
const showDetailModal = ref(false)
const showRejectModal = ref(false)
const selectedRequest = ref<any>(null)
const rejectionReason = ref('')
const isProcessingApproval = ref(false)
const isProcessingRejection = ref(false)

const openRequestDetail = (request: any) => {
  selectedRequest.value = request
  showDetailModal.value = true
}

const handleApproveRequest = async (request: any) => {
  isProcessingApproval.value = true
  try {
    await Promise.all(request.items.map((item: any) => factoryStore.approveTransfer(item.id)))
    toast.success(`Permintaan ${request.delivery_order_no} berhasil disetujui.`)
    showDetailModal.value = false
    factoryStore.fetchDashboardStats()
  } catch (err) {
    toast.error('Gagal menyetujui permintaan: ' + getErrorMessage(err))
  } finally {
    isProcessingApproval.value = false
  }
}

const openRejectRequest = (request: any) => {
  selectedRequest.value = request
  rejectionReason.value = ''
  showRejectModal.value = true
}

const handleRejectRequest = async () => {
  if (!rejectionReason.value.trim()) {
    toast.error('Silakan isi alasan penolakan')
    return
  }
  isProcessingRejection.value = true
  try {
    await Promise.all(
      selectedRequest.value.items.map((item: any) =>
        factoryStore.updateTransfer(item.id, {
          status: 'REJECTED',
          rejection_reason: rejectionReason.value,
        }),
      ),
    )
    toast.success(`Permintaan ${selectedRequest.value.delivery_order_no} berhasil ditolak.`)
    showRejectModal.value = false
    showDetailModal.value = false
    factoryStore.fetchDashboardStats()
  } catch (err) {
    toast.error('Gagal menolak permintaan: ' + getErrorMessage(err))
  } finally {
    isProcessingRejection.value = false
  }
}

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

const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

const branchRequests = computed(() => {
  const pendingTransfers = factoryStore.allTransfers.filter((t) => t.status === 'REQUESTED')
  const groups: Record<
    string,
    {
      delivery_order_no: string
      to_branch_id: string
      to_branch_name: string
      from_factory_id: string
      from_factory_name: string
      created_at: string
      initiated_by: string
      notes: string
      items: Array<{
        id: string
        product_id: string
        product_name: string
        product_sku: string
        quantity: number
        unit: string
        batch_no: string
      }>
    }
  > = {}

  for (const t of pendingTransfers) {
    const doNo = t.delivery_order_no || `REQ-${t.id.slice(0, 8)}`
    if (!groups[doNo]) {
      groups[doNo] = {
        delivery_order_no: doNo,
        to_branch_id: t.to_branch_id,
        to_branch_name: t.to_branch?.name || 'Cabang Lain',
        from_factory_id: t.from_factory_id,
        from_factory_name: t.from_factory?.name || 'Pabrik Lain',
        created_at: t.created_at,
        initiated_by: t.initiated_by || 'BRANCH',
        notes: t.notes || '',
        items: [],
      }
    }
    groups[doNo].items.push({
      id: t.id,
      product_id: t.product_id,
      product_name: t.product?.name || 'Produk Tidak Dikenal',
      product_sku: t.product?.sku || '-',
      quantity: t.quantity,
      unit: t.unit || 'PCS',
      batch_no: t.batch_no || 'DEFAULT',
    })
  }

  return Object.values(groups).sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
  )
})
const submitting = ref(false)
const showRestockModal = ref(false)
const isSubmittingRestock = ref(false)
const selectedAlert = ref<any>(null)

const restockForm = reactive({
  from_factory_id: '',
  to_branch_id: '',
  quantity: 1,
  unit: 'KARTON',
  pcs_per_unit: 24,
  notes: '',
  vehicle_id: '',
  driver_id: '',
  target_shipment_date: '',
  estimated_arrival: '',
})

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

const openProactiveRestock = (alert: any) => {
  selectedAlert.value = alert
  restockForm.to_branch_id = alert.branch_id
  restockForm.notes = `Restock otomatis di bawah threshold (Stok: ${alert.quantity} < Limit: ${alert.min_limit})`
  restockForm.quantity = Math.max(alert.min_limit - alert.quantity, 1)

  if (factoryStore.factories && factoryStore.factories.length > 0) {
    restockForm.from_factory_id = factoryStore.factories[0]?.id || ''
  } else {
    restockForm.from_factory_id = ''
  }

  const prod = factoryStore.products.find((p) => p.id === alert.product_id)
  if (prod) {
    restockForm.unit = prod.unit || 'KARTON'
    restockForm.pcs_per_unit = prod.pcs_per_unit || 24
  } else {
    restockForm.unit = alert.product_unit || 'KARTON'
    restockForm.pcs_per_unit = alert.pcs_per_unit || 24
  }

  restockForm.vehicle_id = ''
  restockForm.driver_id = ''
  restockForm.target_shipment_date = ''
  restockForm.estimated_arrival = ''

  showRestockModal.value = true
}

const submitProactiveRestock = async () => {
  if (!restockForm.from_factory_id) {
    toast.error('Silakan pilih Pabrik Pengirim')
    return
  }
  if (!restockForm.quantity || restockForm.quantity <= 0) {
    toast.error('Jumlah pengiriman harus lebih dari 0')
    return
  }

  isSubmittingRestock.value = true
  try {
    const payload = {
      to_branch_id: restockForm.to_branch_id,
      items: [
        {
          product_id: selectedAlert.value.product_id,
          quantity: restockForm.quantity,
          unit: restockForm.unit,
          pcs_per_unit: restockForm.pcs_per_unit,
        },
      ],
      notes: restockForm.notes,
      target_shipment_date: formatToRFC3339(restockForm.target_shipment_date),
      estimated_arrival: formatToRFC3339(restockForm.estimated_arrival),
      vehicle_id: restockForm.vehicle_id || null,
      driver_id: restockForm.driver_id || null,
      initiated_by: 'FACTORY',
    }

    await factoryStore.sendToBranch(restockForm.from_factory_id, payload)
    toast.success('Pengiriman restock berhasil dibuat.')
    showRestockModal.value = false
    factoryStore.fetchLowStockAlerts()
  } catch (err) {
    toast.error('Gagal mengirim restock: ' + getErrorMessage(err))
  } finally {
    isSubmittingRestock.value = false
  }
}

const formatToRFC3339 = (dateStr: any) => {
  if (!dateStr) return null
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return null
  return date.toISOString()
}

const vehicles = computed(() => {
  return vehicleStore.vehicles.map((v) => {
    const isUsedInTransfer = factoryStore.dashboardStats.recent_transfers.some(
      (t) => t.vehicle_id === v.id && (t.status === 'SHIPPED' || t.status === 'ARRIVED'),
    )
    return {
      ...v,
      isAvailable: !isUsedInTransfer,
    }
  })
})

const drivers = computed(() => {
  return masterDataStore.employees
    .filter((emp: any) => {
      if (!emp.job_position?.title) return true
      const jobTitle = emp.job_position.title.toLowerCase()
      return (
        jobTitle.includes('driver') ||
        jobTitle.includes('sopir') ||
        jobTitle.includes('salesman') ||
        jobTitle.includes('motoris') ||
        jobTitle.includes('delivery') ||
        jobTitle.includes('kurir') ||
        jobTitle.includes('pengiriman')
      )
    })
    .map((e) => {
      const isUsedInTransfer = factoryStore.dashboardStats.recent_transfers.some(
        (t) => t.driver_id === e.id && (t.status === 'SHIPPED' || t.status === 'ARRIVED'),
      )
      return {
        ...e,
        isAvailable: !isUsedInTransfer,
      }
    })
})

const newFactoryForm = reactive({
  name: '',
  location: '',
  branch_id: '',
})

onMounted(() => {
  factoryStore.fetchDashboardStats()
  factoryStore.fetchFactories()
  factoryStore.fetchProducts()
  factoryStore.fetchLowStockAlerts()
  factoryStore.fetchBackorderDemand(masterDataStore.selectedBranchCompanyId || '')
  factoryStore.fetchAllInventory()
  factoryStore.fetchAllTransfers()
  masterDataStore.fetchBranches()
  masterDataStore.fetchEmployees()
  vehicleStore.fetchVehicles()
})

const getFactoryName = (id: string) => {
  const f = factoryStore.factories.find((f: any) => f.id === id)
  return f ? f.name : '-'
}

const formatRelativeTime = (dateStr: string) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diffInMs = now.getTime() - date.getTime()
  const diffInMins = Math.floor(diffInMs / (1000 * 60))
  const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60))
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24))

  if (diffInMins < 1) return 'Baru saja'
  if (diffInMins < 60) return `${diffInMins}m yang lalu`
  if (diffInHours < 24) return `${diffInHours}j yang lalu`
  return `${diffInDays}h yang lalu`
}

const formatStatus = (status: string) => {
  switch (status) {
    case 'REQUESTED':
      return 'Menunggu Approval'
    case 'APPROVED':
      return 'DI ACC GUDANG'
    case 'SHIPPED':
      return 'Dalam Pengiriman'
    case 'ARRIVED':
      return 'Sampai di Tujuan'
    case 'RECEIVED':
      return 'Telah Diterima'
    default:
      return status
  }
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'RECEIVED':
      return 'bg-emerald-50 text-emerald-600 border border-emerald-100 shadow-sm'
    case 'SHIPPED':
      return 'bg-blue-50 text-blue-600 border border-blue-100'
    case 'ARRIVED':
      return 'bg-cyan-50 text-cyan-600 border border-cyan-100'
    case 'REQUESTED':
      return 'bg-orange-50 text-orange-600 border border-orange-100'
    case 'APPROVED':
      return 'bg-emerald-50 text-emerald-700 border border-emerald-200 font-black'
    default:
      return 'bg-slate-50 text-slate-600 border border-slate-100'
  }
}

const goToFactory = (id: string) => {
  router.push({ name: 'factoryInventory', params: { id } })
}

const handleAddFactory = async () => {
  if (!newFactoryForm.name) return
  submitting.value = true
  try {
    await factoryStore.createFactory(newFactoryForm)
    showAddFactory.value = false
    // Reset form
    newFactoryForm.name = ''
    newFactoryForm.location = ''
    newFactoryForm.branch_id = ''
  } catch (err) {
    alert(err)
  } finally {
    submitting.value = false
  }
}
</script>
