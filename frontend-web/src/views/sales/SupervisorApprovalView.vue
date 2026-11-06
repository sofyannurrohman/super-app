<template>
  <div class="space-y-8 p-6">
    <!-- Header Section -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center bg-white text-slate-900 p-8 rounded-2xl shadow-sm border border-slate-200/60 gap-4"
    >
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-3">
          <ClipboardCheck class="w-8 h-8 text-indigo-600" />
          <span>Persetujuan Supervisor (SPV)</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Validasi dan setujui permintaan barang bawaan sales motoris &amp; pengiriman armada logistik (surat jalan) yang dibuat oleh Admin Nota.
        </p>
      </div>
      <div class="flex gap-2 w-full md:w-auto">
        <button
          @click="reloadData"
          class="flex items-center justify-center gap-2 px-5 py-3 bg-white border border-slate-200 rounded-2xl text-sm font-bold text-slate-600 hover:bg-slate-50 transition-colors shadow-sm w-full md:w-auto"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
          Segarkan Data
        </button>
      </div>
    </div>

    <!-- Dual Column Split Layout -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      
      <!-- LEFT COLUMN: Persetujuan Barang Bawaan Sales -->
      <div class="space-y-6 bg-white border border-slate-200/60 rounded-3xl p-6 shadow-sm">
        <div class="flex items-center justify-between border-b border-slate-100 pb-4 flex-wrap gap-2">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-indigo-50 flex items-center justify-center text-indigo-600">
              <Boxes class="w-5 h-5" />
            </div>
            <div>
              <h2 class="text-lg font-bold text-slate-800">Barang Bawaan Sales</h2>
              <p class="text-xs text-slate-400">Request stok motoris &amp; booster</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span class="bg-indigo-100 text-indigo-800 text-xs font-extrabold px-3 py-1 rounded-full">
              {{ filteredRequests.length }} Pending
            </span>
            <Button
              size="sm"
              class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl text-xs flex items-center gap-1.5 shadow-sm"
              @click="openCreateModal"
            >
              <Plus class="w-3.5 h-3.5" />
              <span>+ Request Barang</span>
            </Button>
          </div>
        </div>

        <!-- Search Left -->
        <div class="relative">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchSalesQuery"
            type="text"
            placeholder="Cari nama sales atau produk..."
            class="pl-10 h-10 w-full rounded-xl border border-slate-200 bg-slate-50/50 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500/20 transition-all placeholder:text-slate-400"
          />
        </div>

        <!-- Content Left -->
        <div v-if="loading" class="flex flex-col items-center justify-center py-20 space-y-3">
          <RefreshCw class="w-7 h-7 animate-spin text-indigo-600" />
          <p class="text-slate-400 text-xs font-semibold animate-pulse">Memuat data permintaan...</p>
        </div>

        <div v-else-if="filteredRequests.length === 0" class="flex flex-col items-center justify-center py-16 text-center bg-slate-50/50 rounded-2xl border-2 border-dashed border-slate-100">
          <div class="bg-white p-3 rounded-full shadow text-emerald-500 mb-3 ring-4 ring-emerald-50">
            <CheckCircle2 class="h-6 w-6" />
          </div>
          <h4 class="text-sm font-bold text-slate-800">Tidak ada permintaan pending</h4>
          <p class="text-xs text-slate-400 mt-1 max-w-[240px]">
            Semua permintaan barang bawaan sales telah diproses.
          </p>
        </div>

        <div v-else class="space-y-4 max-h-[65vh] overflow-y-auto pr-1">
          <!-- Request Card Left -->
          <div
            v-for="req in filteredRequests"
            :key="req.id"
            class="border border-slate-100 rounded-2xl bg-white hover:border-slate-200 transition-all p-4 space-y-3 relative shadow-sm"
          >
            <!-- Header Card -->
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="h-8 w-8 rounded-lg bg-indigo-50 flex items-center justify-center text-indigo-600 font-bold text-xs">
                  {{ getInitials(req.employee) }}
                </div>
                <div>
                  <h4 class="font-bold text-xs text-slate-800 leading-tight">
                    {{ req.employee?.first_name }} {{ req.employee?.last_name }}
                  </h4>
                  <span class="text-[9px] text-slate-400">Ref: {{ req.reference_no }}</span>
                </div>
              </div>
              <span class="text-[10px] text-slate-400 whitespace-nowrap">
                {{ formatDate(req.transfer_date) }}
              </span>
            </div>

            <!-- Product and Qty Detail -->
            <div class="bg-slate-50/80 p-3 rounded-xl border border-slate-100 space-y-2">
              <div class="font-bold text-slate-800 text-xs">{{ req.product?.name }}</div>
              <div class="grid grid-cols-3 gap-2 text-[11px] text-slate-500">
                <div>
                  <span class="block text-slate-400">Stok Gudang:</span>
                  <span class="font-bold text-slate-700">{{ req.product?.warehouse_stock || 0 }} {{ req.product?.unit }}</span>
                </div>
                <div>
                  <span class="block text-slate-400">Diminta:</span>
                  <span class="font-extrabold text-indigo-700">{{ req.quantity }} {{ req.product?.unit }}</span>
                </div>
                <div>
                  <span class="block text-slate-400">Nilai Est:</span>
                  <span class="font-bold text-slate-700">{{ formatCurrency(getRequestValue(req)) }}</span>
                </div>
              </div>
            </div>

            <!-- KPI context (always shown) -->
            <div class="bg-slate-50 p-3 rounded-xl border border-slate-100 space-y-2 mt-2">
              <div class="flex justify-between items-center text-xs">
                <span class="text-slate-500 font-semibold">Progress Omzet Bulan Ini:</span>
                <span class="font-bold" :class="getKPIPercent(req.employee_id || req.employee?.id) >= 100 && getKpiTarget(getSalesmanKPI(req.employee_id || req.employee?.id)) > 0 ? 'text-emerald-600' : 'text-indigo-600'">
                  {{ getKPIPercent(req.employee_id || req.employee?.id).toFixed(1) }}%
                </span>
              </div>
              
              <!-- Progress Bar -->
              <div class="w-full bg-slate-200 rounded-full h-1.5 overflow-hidden">
                <div 
                  class="h-1.5 rounded-full transition-all duration-500"
                  :class="getKPIPercent(req.employee_id || req.employee?.id) >= 100 && getKpiTarget(getSalesmanKPI(req.employee_id || req.employee?.id)) > 0 ? 'bg-emerald-500' : 'bg-indigo-500'"
                  :style="`width: ${Math.min(getKPIPercent(req.employee_id || req.employee?.id), 100)}%`"
                ></div>
              </div>

              <div class="flex justify-between items-center text-[10px]">
                <div>
                  <span class="text-slate-400 block">Tercapai</span>
                  <span class="font-bold text-slate-700">{{ formatCurrency(getKpiAchieved(getSalesmanKPI(req.employee_id || req.employee?.id))) }}</span>
                </div>
                <div class="text-right">
                  <span class="text-slate-400 block">Target</span>
                  <span class="font-bold text-slate-700">{{ formatCurrency(getKpiTarget(getSalesmanKPI(req.employee_id || req.employee?.id))) }}</span>
                </div>
              </div>
            </div>

            <!-- Actions block -->
            <div class="flex items-center justify-between pt-2 border-t border-slate-50 gap-4">
              <!-- Qty adjustment -->
              <div class="flex items-center gap-1.5">
                <button
                  type="button"
                  @click="adjustApprovedQty(req.id, -req.product?.pcs_per_unit || -1)"
                  class="h-7 w-7 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-lg flex items-center justify-center font-bold text-xs"
                >
                  -
                </button>
                <input
                  v-model.number="approvedQuantities[req.id]"
                  type="number"
                  class="h-7 w-14 text-center rounded-lg border border-slate-200 text-xs font-bold focus:outline-none focus:ring-1 focus:ring-indigo-500"
                  min="1"
                />
                <button
                  type="button"
                  @click="adjustApprovedQty(req.id, req.product?.pcs_per_unit || 1)"
                  class="h-7 w-7 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-lg flex items-center justify-center font-bold text-xs"
                >
                  +
                </button>
              </div>

              <!-- Button actions -->
              <div class="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  class="text-red-600 border-red-200 hover:bg-red-50 hover:text-red-700 h-8 rounded-lg text-xs"
                  @click="confirmReject(req)"
                >
                  Tolak
                </Button>
                <Button
                  size="sm"
                  class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold h-8 rounded-lg text-xs flex items-center gap-1"
                  @click="handleApprove(req)"
                >
                  Setujui
                </Button>
              </div>
            </div>
            
            <div v-if="approvedQuantities[req.id] !== req.quantity" class="text-[9px] text-right font-medium text-amber-600 italic">
              Qty disetujui akan diubah menjadi {{ approvedQuantities[req.id] }}
            </div>
          </div>
        </div>
      </div>

      <!-- RIGHT COLUMN: Persetujuan Pengiriman (SJ) -->
      <div class="space-y-6 bg-white border border-slate-200/60 rounded-3xl p-6 shadow-sm">
        <div class="flex items-center justify-between border-b border-slate-100 pb-4">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-amber-50 flex items-center justify-center text-amber-600">
              <Truck class="w-5 h-5" />
            </div>
            <div>
              <h2 class="text-lg font-bold text-slate-800">Manajemen Pengiriman</h2>
              <p class="text-xs text-slate-400">Persetujuan Batch SJ buatan Admin Nota</p>
            </div>
          </div>
          <span class="bg-amber-100 text-amber-800 text-xs font-extrabold px-3 py-1 rounded-full">
            {{ filteredBatches.length }} Pending
          </span>
        </div>

        <!-- Search Right -->
        <div class="relative">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchBatchQuery"
            type="text"
            placeholder="Cari no. SJ, armada, driver..."
            class="pl-10 h-10 w-full rounded-xl border border-slate-200 bg-slate-50/50 text-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20 transition-all placeholder:text-slate-400"
          />
        </div>

        <!-- Content Right -->
        <div v-if="loading" class="flex flex-col items-center justify-center py-20 space-y-3">
          <RefreshCw class="w-7 h-7 animate-spin text-amber-600" />
          <p class="text-slate-400 text-xs font-semibold animate-pulse">Memuat data pengiriman...</p>
        </div>

        <div v-else-if="filteredBatches.length === 0" class="flex flex-col items-center justify-center py-16 text-center bg-slate-50/50 rounded-2xl border-2 border-dashed border-slate-100">
          <div class="bg-white p-3 rounded-full shadow text-emerald-500 mb-3 ring-4 ring-emerald-50">
            <CheckCircle2 class="h-6 w-6" />
          </div>
          <h4 class="text-sm font-bold text-slate-800">Tidak ada batch pengiriman pending</h4>
          <p class="text-xs text-slate-400 mt-1 max-w-[240px]">
            Semua batch pengiriman telah disetujui supervisor.
          </p>
        </div>

        <div v-else class="space-y-4 max-h-[65vh] overflow-y-auto pr-1">
          <!-- Batch Card Right -->
          <div
            v-for="batch in filteredBatches"
            :key="batch.id"
            class="border border-slate-100 rounded-2xl bg-white hover:border-slate-200 transition-all p-4 space-y-3 shadow-sm"
          >
            <!-- Batch Header -->
            <div class="flex items-center justify-between">
              <div>
                <span class="text-xs font-bold text-amber-700 bg-amber-50 px-2 py-0.5 rounded border border-amber-100 font-mono">
                  {{ batch.delivery_order_no || 'DRAFT-BATCH' }}
                </span>
                <div class="text-[9px] text-slate-400 mt-1 font-mono">ID: {{ batch.id.substring(0, 8) }}</div>
              </div>
              <span class="text-[10px] text-slate-400">
                Dibuat: {{ formatDate(batch.created_at) }}
              </span>
            </div>

            <!-- Armada details -->
            <div class="grid grid-cols-3 gap-2 bg-slate-50 p-2.5 rounded-xl text-slate-600 text-xs border border-slate-100">
              <div>
                <span class="text-[9px] text-slate-400 block font-semibold uppercase">Driver:</span>
                <span class="font-bold text-slate-800">{{ batch.driver?.first_name || batch.daily_driver?.name || '-' }}</span>
              </div>
              <div>
                <span class="text-[9px] text-slate-400 block font-semibold uppercase">Helper:</span>
                <span class="font-bold text-slate-800">{{ batch.helper?.first_name || batch.daily_helper?.name || '-' }}</span>
              </div>
              <div>
                <span class="text-[9px] text-slate-400 block font-semibold uppercase">Kendaraan:</span>
                <span class="font-bold text-slate-800">{{ batch.vehicle ? `${batch.vehicle.license_plate} (${batch.vehicle.type})` : '-' }}</span>
              </div>
            </div>

            <!-- Orders inside batch -->
            <div class="space-y-2">
              <span class="text-[9px] text-slate-400 uppercase font-bold tracking-wider block">Nota Pengiriman:</span>
              <div class="space-y-1">
                <div
                  v-for="(item, idx) in batch.items"
                  :key="item.id"
                  class="flex items-center justify-between bg-slate-50 text-[11px] px-2.5 py-1.5 border border-slate-100 rounded-lg hover:bg-slate-100/50"
                >
                  <div class="flex items-center gap-1.5 truncate">
                    <span class="text-[9px] text-slate-400 bg-white border border-slate-200/80 px-1 rounded-sm">{{ Number(idx) + 1 }}</span>
                    <span class="font-semibold text-slate-800 truncate" :title="item.sales_order?.store?.name">
                      {{ item.sales_order?.store?.name || 'Toko Lain' }}
                    </span>
                  </div>
                  <span class="font-mono text-slate-500 text-[10px] shrink-0">
                    {{ item.sales_order?.so_number || item.sales_transaction?.receipt_no || item.sales_order_id?.substring(0, 8) || '-' }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Button to open detailed modal -->
            <button
              type="button"
              @click="openPreview(batch)"
              class="w-full py-2 bg-indigo-50 hover:bg-indigo-100 border border-indigo-100 rounded-xl text-xs font-bold text-indigo-700 transition-colors flex items-center justify-center gap-1.5 shadow-sm"
            >
              <Eye class="w-3.5 h-3.5" />
              Lihat Detail & Nominal ({{ formatCurrency(getBatchTotalAmount(batch)) }})
            </button>

            <div v-if="batch.notes" class="text-[10px] text-slate-500 bg-slate-50/50 px-2.5 py-1.5 rounded-lg italic">
              <strong>Catatan:</strong> {{ batch.notes }}
            </div>

            <!-- Actions batch -->
            <div class="flex justify-end gap-2 pt-2 border-t border-slate-50">
              <Button
                size="sm"
                variant="outline"
                class="text-red-600 border-red-200 hover:bg-red-50 hover:text-red-700 h-8 rounded-lg text-xs"
                @click="confirmRejectBatch(batch)"
              >
                Batalkan Batch
              </Button>
              <Button
                size="sm"
                class="bg-amber-600 hover:bg-amber-700 text-white font-bold h-8 rounded-lg text-xs flex items-center gap-1 shadow-sm"
                @click="handleApproveBatch(batch)"
              >
                Setujui & Terbitkan SJ
              </Button>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- Modal Preview Detail SJ -->
    <div
      v-if="selectedPreviewBatch"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm transition-opacity"
      @click.self="closePreview"
    >
      <div class="bg-white rounded-3xl border border-slate-200 shadow-2xl max-w-2xl w-full max-h-[85vh] overflow-hidden flex flex-col animate-in fade-in zoom-in-95 duration-200">
        <!-- Modal Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-slate-100 bg-slate-50/50">
          <div>
            <div class="flex items-center gap-2">
              <span class="text-xs font-bold text-amber-700 bg-amber-50 px-2 py-0.5 rounded border border-amber-100 font-mono">
                {{ selectedPreviewBatch.delivery_order_no || 'DRAFT-BATCH' }}
              </span>
              <span class="text-xs text-slate-400 font-mono">ID: {{ selectedPreviewBatch.id.substring(0, 8) }}</span>
            </div>
            <p class="text-xs text-slate-500 mt-1">Rincian produk, toko, sales, dan nominal transaksi Surat Jalan</p>
          </div>
          <button
            @click="closePreview"
            class="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Modal Content (Scrollable) -->
        <div class="p-6 overflow-y-auto space-y-6 flex-1 text-left">
          <!-- Dispatch / Armada Details -->
          <div class="grid grid-cols-3 gap-3 bg-slate-50 p-3 rounded-2xl text-slate-600 text-xs border border-slate-100">
            <div>
              <span class="text-[9px] text-slate-400 block font-bold uppercase tracking-wider">Driver:</span>
              <span class="font-extrabold text-slate-800 text-sm mt-0.5 block">{{ selectedPreviewBatch.driver ? `${selectedPreviewBatch.driver.first_name} ${selectedPreviewBatch.driver.last_name || ''}` : selectedPreviewBatch.daily_driver?.name || '-' }}</span>
            </div>
            <div>
              <span class="text-[9px] text-slate-400 block font-bold uppercase tracking-wider">Helper:</span>
              <span class="font-extrabold text-slate-800 text-sm mt-0.5 block">{{ selectedPreviewBatch.helper ? `${selectedPreviewBatch.helper.first_name} ${selectedPreviewBatch.helper.last_name || ''}` : selectedPreviewBatch.daily_helper?.name || '-' }}</span>
            </div>
            <div>
              <span class="text-[9px] text-slate-400 block font-bold uppercase tracking-wider">Kendaraan / Nopol:</span>
              <span class="font-extrabold text-slate-800 text-sm mt-0.5 block">{{ selectedPreviewBatch.vehicle ? `${selectedPreviewBatch.vehicle.license_plate} (${selectedPreviewBatch.vehicle.type})` : '-' }}</span>
            </div>
          </div>

          <!-- Orders inside batch -->
          <div class="space-y-4">
            <span class="text-[11px] text-slate-500 uppercase font-extrabold tracking-wider block">
              Daftar Toko & Pesanan:
            </span>
            <div class="space-y-4">
              <div
                v-for="(item, idx) in selectedPreviewBatch.items"
                :key="item.id"
                class="border border-slate-100 rounded-2xl bg-slate-50/30 p-4 space-y-3"
              >
                <!-- Store Header -->
                <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-1 pb-2 border-b border-slate-100">
                  <div class="flex items-center gap-2">
                    <span class="text-[10px] font-bold text-slate-500 bg-slate-200/80 px-2 py-0.5 rounded-md">
                      Toko #{{ Number(idx) + 1 }}
                    </span>
                    <span class="font-extrabold text-slate-900 text-base">
                      {{ getStore(item)?.name || 'Toko Lain' }}
                    </span>
                    <span
                      v-if="getStore(item)?.category"
                      class="text-[9px] font-extrabold px-2 py-0.5 rounded-full"
                      :class="getStore(item)?.category === 'TOKO_BARU' ? 'bg-emerald-100 text-emerald-800' : 'bg-blue-100 text-blue-800'"
                    >
                      {{ getStore(item)?.category }}
                    </span>
                  </div>
                  <span class="font-mono text-slate-600 text-xs font-bold text-right bg-slate-100 px-2 py-0.5 rounded-md border border-slate-200/50">
                    {{ item.sales_order?.so_number || item.sales_transaction?.receipt_no || item.sales_order_id?.substring(0, 8) || '-' }}
                  </span>
                </div>

                <!-- Store & Sales details -->
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs text-slate-600">
                  <div>
                    <span class="text-[10px] text-slate-400 block font-semibold uppercase tracking-wider">Alamat Toko:</span>
                    <span class="text-slate-700 font-medium block mt-0.5">{{ getStore(item)?.address || '-' }}</span>
                    <span v-if="getStore(item)?.phone_number" class="block text-slate-500 text-[10px] mt-1">
                      Telp: {{ getStore(item)?.phone_number }}
                    </span>
                  </div>
                  <div>
                    <span class="text-[10px] text-slate-400 block font-semibold uppercase tracking-wider">Sales Representative:</span>
                    <span class="font-bold text-indigo-700 bg-indigo-50 px-2 py-1 rounded-xl border border-indigo-100/50 inline-block mt-1">
                      {{ getSalesperson(item) }}
                    </span>
                  </div>
                </div>

                <!-- Products Table/List -->
                <div class="bg-white rounded-xl border border-slate-150 overflow-hidden shadow-sm">
                  <table class="w-full text-left text-xs">
                    <thead>
                      <tr class="bg-slate-50 text-slate-500 font-bold border-b border-slate-150">
                        <th class="p-2.5">Produk</th>
                        <th class="p-2.5 text-center">Qty</th>
                        <th class="p-2.5 text-right">Harga</th>
                        <th class="p-2.5 text-right">Subtotal</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="subitem in getItems(item)"
                        :key="subitem.id"
                        class="border-b border-slate-100 last:border-0 hover:bg-slate-50/30"
                      >
                        <td class="p-2.5 font-semibold text-slate-800">
                          {{ subitem.product?.name || 'Produk Tidak Diketahui' }}
                        </td>
                        <td class="p-2.5 text-center text-slate-600 font-bold">
                          {{ getItemQtyDisplay(subitem) }}
                        </td>
                        <td class="p-2.5 text-right text-slate-500">
                          {{ formatCurrency(getItemPrice(subitem)) }}
                        </td>
                        <td class="p-2.5 text-right font-extrabold text-slate-800">
                          {{ formatCurrency(subitem.subtotal || 0) }}
                        </td>
                      </tr>
                      <tr v-if="getItems(item).length === 0">
                        <td colspan="4" class="p-3 text-center text-slate-400 italic">
                          Tidak ada produk dalam order ini
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <!-- Order Total -->
                <div class="flex justify-between items-center text-xs pt-1">
                  <span class="text-slate-500 font-bold">Nominal Transaksi Toko:</span>
                  <span class="font-black text-slate-900 bg-slate-100 border border-slate-200/50 px-2.5 py-1 rounded-lg">
                    {{ formatCurrency(getOrderTotalAmount(item)) }}
                  </span>
                </div>

              </div>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="px-6 py-4 border-t border-slate-100 bg-slate-900 text-white flex flex-col sm:flex-row justify-between items-center gap-4">
          <div class="text-left w-full sm:w-auto">
            <span class="text-[10px] text-slate-400 font-bold uppercase tracking-wider block">
              Total Nominal Transaksi Surat Jalan
            </span>
            <span class="text-xs text-amber-400 font-semibold block">
              (Akumulasi semua nota di atas)
            </span>
          </div>
          <div class="flex items-center gap-6 justify-between sm:justify-end w-full sm:w-auto">
            <span class="text-2xl font-black text-amber-400 tracking-tight">
              {{ formatCurrency(getBatchTotalAmount(selectedPreviewBatch)) }}
            </span>
            <div class="flex gap-2">
              <Button
                variant="outline"
                class="text-white border-slate-700 bg-slate-800 hover:bg-slate-750 hover:text-white rounded-xl text-xs font-bold"
                @click="closePreview"
              >
                Tutup
              </Button>
              <Button
                class="bg-amber-600 hover:bg-amber-700 text-white font-bold rounded-xl text-xs flex items-center gap-1 shadow-sm"
                @click="() => { handleApproveBatch(selectedPreviewBatch); closePreview(); }"
              >
                Setujui & Terbitkan
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Buat Request Barang Bawaan Sales (Admin Nota & SPV) -->
    <div
      v-if="showCreateModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm transition-opacity"
      @click.self="closeCreateModal"
    >
      <div class="bg-white rounded-3xl border border-slate-200 shadow-2xl max-w-3xl w-full max-h-[90vh] overflow-hidden flex flex-col animate-in fade-in zoom-in-95 duration-200">
        <!-- Modal Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-slate-100 bg-slate-50/50">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-indigo-100 text-indigo-700 flex items-center justify-center">
              <PackagePlus class="w-5 h-5" />
            </div>
            <div>
              <h3 class="text-base font-bold text-slate-900">Buat Request Barang Bawaan Sales</h3>
              <p class="text-xs text-slate-500">
                Input barang bawaan sales motoris / booster oleh Admin Nota &amp; Supervisor
              </p>
            </div>
          </div>
          <button
            @click="closeCreateModal"
            class="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Modal Body (Scrollable) -->
        <div class="p-6 overflow-y-auto space-y-6 flex-1 text-left">
          <!-- Top Row: Salesman & Type -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700 flex items-center gap-1">
                <User class="w-3.5 h-3.5 text-indigo-600" />
                Pilih Salesman / Motoris <span class="text-red-500">*</span>
              </label>
              <select
                v-model="createForm.employee_id"
                class="h-10 w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 text-xs font-medium focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
              >
                <option value="">-- Pilih Salesman --</option>
                <option
                  v-for="emp in branchSalesEmployees"
                  :key="emp.id"
                  :value="emp.id"
                >
                  {{ emp.first_name }} {{ emp.last_name }} ({{ emp.job_position?.title || 'Sales' }})
                </option>
              </select>
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-bold text-slate-700 flex items-center gap-1">
                <ArrowRightLeft class="w-3.5 h-3.5 text-indigo-600" />
                Tipe Perpindahan <span class="text-red-500">*</span>
              </label>
              <select
                v-model="createForm.type"
                class="h-10 w-full rounded-xl border border-slate-200 bg-slate-50/50 px-3 text-xs font-medium focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
              >
                <option value="TRANSFER">TRANSFER (Gudang &rarr; Bawaan Sales)</option>
                <option value="RETURN">RETUR (Kembalikan ke Gudang)</option>
              </select>
            </div>
          </div>

          <!-- Items Table / List -->
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <label class="text-xs font-bold text-slate-700 flex items-center gap-1.5">
                <Boxes class="w-4 h-4 text-indigo-600" />
                Daftar Barang Bawaan <span class="text-red-500">*</span>
              </label>
              <button
                type="button"
                @click="addItemRow"
                class="text-xs font-bold text-indigo-600 hover:text-indigo-800 bg-indigo-50 hover:bg-indigo-100 px-3 py-1 rounded-lg transition-colors flex items-center gap-1"
              >
                <Plus class="w-3.5 h-3.5" />
                Tambah Barang
              </button>
            </div>

            <div class="space-y-3">
              <div
                v-for="(item, idx) in createForm.items"
                :key="idx"
                class="p-3 bg-slate-50/70 border border-slate-200/80 rounded-2xl space-y-2 relative"
              >
                <div class="flex items-start justify-between gap-2">
                  <span class="text-[10px] font-extrabold bg-slate-200 text-slate-700 px-2 py-0.5 rounded-md">
                    Item #{{ idx + 1 }}
                  </span>
                  <button
                    v-if="createForm.items.length > 1"
                    type="button"
                    @click="removeItemRow(idx)"
                    class="p-1 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-12 gap-3 items-end">
                  <!-- Product Dropdown (col 7) -->
                  <div class="md:col-span-7 space-y-1">
                    <span class="text-[10px] font-bold text-slate-500 uppercase">Pilih Produk / Bonus</span>
                    <select
                      v-model="item.item_selection"
                      class="h-9 w-full rounded-xl border border-slate-200 bg-white px-3 text-xs font-semibold focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                    >
                      <option value="">-- Pilih Barang --</option>
                      <optgroup label="Produk Penjualan">
                        <option
                          v-for="prod in factoryStore.sellableProducts"
                          :key="prod.id"
                          :value="'PRODUCT:' + prod.id"
                        >
                          {{ prod.name }} (Stok: {{ getWarehouseStock(prod.id) }} {{ prod.unit || 'PCS' }})
                        </option>
                      </optgroup>
                      <optgroup label="Barang Bonus & Hadiah">
                        <option
                          v-for="bonus in warehouseStore.bonusInventory"
                          :key="bonus.id"
                          :value="'BONUS:' + bonus.id"
                        >
                          {{ bonus.name }} [BONUS] (Stok: {{ bonus.quantity }} {{ bonus.unit || 'PCS' }})
                        </option>
                      </optgroup>
                    </select>
                  </div>

                  <!-- Quantity (col 2) -->
                  <div class="md:col-span-2 space-y-1">
                    <span class="text-[10px] font-bold text-slate-500 uppercase">Jumlah</span>
                    <input
                      v-model.number="item.quantity"
                      type="number"
                      min="1"
                      class="h-9 w-full rounded-xl border border-slate-200 bg-white px-3 text-xs font-bold text-center focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                    />
                  </div>

                  <!-- Unit & Subtotal Preview (col 3) -->
                  <div class="md:col-span-3 space-y-1">
                    <span class="text-[10px] font-bold text-slate-500 uppercase">Satuan &amp; Nilai</span>
                    <div class="h-9 flex items-center justify-between px-2.5 bg-white border border-slate-200 rounded-xl text-xs">
                      <span class="font-bold text-indigo-700">{{ getItemUnit(item.item_selection) }}</span>
                      <span class="font-mono text-[11px] text-slate-600 font-semibold">{{ formatCurrency(getItemPrice(item.item_selection) * (item.quantity || 0)) }}</span>
                    </div>
                  </div>
                </div>

                <!-- Stock warning if insufficient in TRANSFER -->
                <div
                  v-if="createForm.type === 'TRANSFER' && item.item_selection && item.quantity > getWarehouseStock(item.item_selection.replace('PRODUCT:', '').replace('BONUS:', ''))"
                  class="text-[10px] text-red-600 font-bold bg-red-50 px-2.5 py-1 rounded-lg flex items-center gap-1"
                >
                  <AlertCircle class="w-3.5 h-3.5 text-red-600" />
                  Perhatian: Jumlah yang diminta ({{ item.quantity }}) melebihi stok gudang yang tersedia ({{ getWarehouseStock(item.item_selection.replace('PRODUCT:', '').replace('BONUS:', '')) }}).
                </div>
              </div>
            </div>
          </div>

          <!-- Notes -->
          <div class="space-y-1.5">
            <label class="text-xs font-bold text-slate-700">Catatan / Keterangan Keperluan</label>
            <textarea
              v-model="createForm.notes"
              rows="2"
              placeholder="Contoh: Barang bawaan sales kanvas motoris area Barat..."
              class="w-full rounded-xl border border-slate-200 bg-slate-50/50 p-3 text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
            ></textarea>
          </div>
        </div>

        <!-- Modal Footer with Total and Dual Action Buttons -->
        <div class="px-6 py-4 border-t border-slate-100 bg-slate-900 text-white flex flex-col sm:flex-row justify-between items-center gap-4">
          <div class="text-left w-full sm:w-auto">
            <span class="text-[10px] text-slate-400 font-bold uppercase tracking-wider block">
              Total Nilai Estimasi Bawaan
            </span>
            <span class="text-xl font-black text-amber-400 tracking-tight">
              {{ formatCurrency(calculatedTotalEstimatedValue) }}
            </span>
          </div>

          <div class="flex items-center gap-2 justify-end w-full sm:w-auto flex-wrap">
            <Button
              variant="outline"
              class="text-white border-slate-700 bg-slate-800 hover:bg-slate-750 hover:text-white rounded-xl text-xs font-bold"
              @click="closeCreateModal"
              :disabled="createSubmitting"
            >
              Batal
            </Button>
            <Button
              variant="outline"
              class="border-indigo-400 text-indigo-300 bg-indigo-950/60 hover:bg-indigo-900 hover:text-white rounded-xl text-xs font-bold flex items-center gap-1.5"
              @click="submitCreateRequest('WAITING_SUPERVISOR')"
              :disabled="createSubmitting"
            >
              <Clock class="w-3.5 h-3.5 text-indigo-400" />
              <span>Ajukan ke SPV</span>
            </Button>
            <Button
              class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl text-xs flex items-center gap-1.5 shadow-sm"
              @click="submitCreateRequest('PENDING')"
              :disabled="createSubmitting"
            >
              <CheckCircle2 class="w-3.5 h-3.5" />
              <span>Simpan &amp; Langsung Setujui</span>
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Confirm Dialog -->
    <ConfirmationDialog
      v-model:open="confirmDialog.open"
      :title="confirmDialog.title"
      :description="confirmDialog.description"
      :variant="confirmDialog.variant"
      :confirm-text="confirmDialog.confirmText"
      :cancel-text="confirmDialog.cancelText"
      :is-loading="confirmDialog.isLoading"
      @confirm="confirmDialog.onConfirm"
      @cancel="confirmDialog.open = false"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed, ref, reactive, watch } from 'vue'
import { useSalesTransferStore } from '@/stores/salesTransfer'
import { useDeliveryStore } from '@/stores/delivery'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import { useWarehouseStore } from '@/stores/warehouse'
import { Button } from '@/components/ui/button'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import {
  Search,
  RefreshCw,
  Boxes,
  Truck,
  CheckCircle2,
  ClipboardCheck,
  Eye,
  X,
  Plus,
  PackagePlus,
  User,
  ArrowRightLeft,
  Trash2,
  AlertCircle,
  Clock,
} from 'lucide-vue-next'

const transferStore = useSalesTransferStore()
const deliveryStore = useDeliveryStore()
const masterDataStore = useMasterDataStore()
const factoryStore = useFactoryStore()
const warehouseStore = useWarehouseStore()

const loading = ref(false)
const searchSalesQuery = ref('')
const searchBatchQuery = ref('')
const pendingRequests = ref<any[]>([])
const performanceList = ref<any[]>([])
const approvedQuantities = reactive<Record<string, number>>({})

// Create Request Modal States (Admin Nota & SPV)
const showCreateModal = ref(false)
const createSubmitting = ref(false)
const createForm = reactive({
  employee_id: '',
  type: 'TRANSFER' as 'TRANSFER' | 'RETURN',
  notes: '',
  items: [
    {
      item_selection: '',
      quantity: 1,
      notes: '',
    },
  ],
})

const isSalesPosition = (title?: string, deptName?: string) => {
  const t = (title || '').toLowerCase()
  const d = (deptName || '').toLowerCase()

  // Blacklist posisi non-sales, harian, logistik, operasional, teknis, kasir
  const isExcluded = /(operator|driver|sopir|kuli|helper|gudang|warehouse|security|cleaning|harian|validator|editor|designer|desain|purchasing|auditor|kasir|cashier|teknisi|technician)/i.test(t)
  if (isExcluded) return false

  // Whitelist sales & marketing
  const isSales = /(sales|motoris|booster|marketing|take order|task order|\bto\b|\bmt\b|modern trade|key account|\bkam\b|kanvas)/i.test(t) ||
                  /(sales|marketing|commercial)/i.test(d)

  return isSales
}

const branchSalesEmployees = computed(() => {
  const currentBranchId = masterDataStore.selectedBranchId
  const list = masterDataStore.employees.filter((emp: any) => {
    const isActive = emp.user?.is_active !== false &&
      emp.employment_status !== 'RESIGNED' &&
      emp.employment_status !== 'TERMINATED' &&
      !String(emp.employment_status || '').toLowerCase().includes('harian')
    if (!isActive) return false

    if (currentBranchId && currentBranchId !== 'ALL_BRANCHES') {
      if (emp.branch_id && emp.branch_id !== currentBranchId) return false
    }
    const jobTitle = emp.job_position?.title || ''
    const deptName = emp.department?.name || ''
    return isSalesPosition(jobTitle, deptName)
  })

  return list.sort((a: any, b: any) => {
    const nameA = `${a.first_name} ${a.last_name || ''}`.trim()
    const nameB = `${b.first_name} ${b.last_name || ''}`.trim()
    return nameA.localeCompare(nameB)
  })
})

const getWarehouseStock = (productIdOrBonusId: string) => {
  if (!productIdOrBonusId) return 0
  const goodStocks = warehouseStore.inventory.filter(
    (i: any) => i.product_id === productIdOrBonusId && (!i.stock_type || i.stock_type === 'GOOD') && i.batch_no !== 'BACKORDER',
  )
  if (goodStocks.length > 0) {
    return goodStocks.reduce((sum: number, i: any) => sum + (Number(i.quantity) || 0), 0)
  }
  const bonus = warehouseStore.bonusInventory.find((b: any) => b.id === productIdOrBonusId)
  if (bonus) return bonus.quantity
  return 0
}

const getItemUnit = (itemSelection: string) => {
  if (!itemSelection) return '-'
  if (itemSelection.startsWith('PRODUCT:')) {
    const prodId = itemSelection.replace('PRODUCT:', '')
    const prod = factoryStore.sellableProducts.find((p: any) => p.id === prodId)
    return prod?.unit || 'PCS'
  }
  if (itemSelection.startsWith('BONUS:')) {
    const bonusId = itemSelection.replace('BONUS:', '')
    const bonus = warehouseStore.bonusInventory.find((b: any) => b.id === bonusId)
    return bonus?.unit || 'PCS'
  }
  return '-'
}

const getItemPrice = (itemOrSelection: any) => {
  if (!itemOrSelection) return 0
  if (typeof itemOrSelection === 'string') {
    if (itemOrSelection.startsWith('PRODUCT:')) {
      const prodId = itemOrSelection.replace('PRODUCT:', '')
      const prod = factoryStore.sellableProducts.find((p: any) => p.id === prodId)
      return prod?.selling_price || 0
    }
    return 0
  }
  return itemOrSelection.price !== undefined ? itemOrSelection.price : (itemOrSelection.price_at_transaction || 0)
}

const calculatedTotalEstimatedValue = computed(() => {
  return createForm.items.reduce((sum, it) => {
    if (!it.item_selection || !it.quantity) return sum
    const price = getItemPrice(it.item_selection)
    return sum + price * it.quantity
  }, 0)
})

const addItemRow = () => {
  createForm.items.push({
    item_selection: '',
    quantity: 1,
    notes: '',
  })
}

const removeItemRow = (idx: number) => {
  if (createForm.items.length > 1) {
    createForm.items.splice(idx, 1)
  }
}

const resetCreateForm = () => {
  createForm.employee_id = ''
  createForm.type = 'TRANSFER'
  createForm.notes = ''
  createForm.items = [
    {
      item_selection: '',
      quantity: 1,
      notes: '',
    },
  ]
}

const openCreateModal = async () => {
  showCreateModal.value = true
  try {
    await Promise.all([
      masterDataStore.fetchEmployees(true),
      warehouseStore.fetchInventory(),
      warehouseStore.fetchBonusInventory(),
      factoryStore.fetchProducts(),
    ])
  } catch (err) {
    console.error('Failed to load master data for request modal:', err)
  }
}

const closeCreateModal = () => {
  showCreateModal.value = false
  resetCreateForm()
}

const submitCreateRequest = async (targetStatus: 'WAITING_SUPERVISOR' | 'PENDING') => {
  if (!createForm.employee_id) {
    toast.error('Silakan pilih salesman terlebih dahulu.')
    return
  }
  if (!createForm.items || createForm.items.length === 0) {
    toast.error('Daftar barang tidak boleh kosong.')
    return
  }
  for (const [i, it] of createForm.items.entries()) {
    if (!it || !it.item_selection) {
      toast.error(`Pilih produk/barang pada item #${i + 1}`)
      return
    }
    if (!it.quantity || it.quantity <= 0) {
      toast.error(`Jumlah pada item #${i + 1} harus lebih dari 0`)
      return
    }
  }

  createSubmitting.value = true
  try {
    const payload = {
      employee_id: createForm.employee_id,
      type: createForm.type,
      status: targetStatus,
      notes: createForm.notes,
      items: createForm.items.map((it) => {
        const isProduct = it.item_selection.startsWith('PRODUCT:')
        return {
          product_id: isProduct ? it.item_selection.replace('PRODUCT:', '') : undefined,
          bonus_inventory_id: !isProduct ? it.item_selection.replace('BONUS:', '') : undefined,
          quantity: it.quantity,
          unit: getItemUnit(it.item_selection),
          notes: it.notes || createForm.notes,
        }
      }),
    }

    await transferStore.createBatchTransfer(payload, masterDataStore.selectedBranchId)

    if (targetStatus === 'WAITING_SUPERVISOR') {
      toast.success('Permintaan barang bawaan berhasil diajukan dan menunggu persetujuan SPV.')
    } else {
      toast.success('Permintaan barang bawaan berhasil disimpan dan langsung disetujui.')
    }

    closeCreateModal()
    await reloadData()
  } catch (err: any) {
    toast.error(err || 'Gagal membuat request barang bawaan')
  } finally {
    createSubmitting.value = false
  }
}

// Modal preview states
const selectedPreviewBatch = ref<any>(null)
const openPreview = (batch: any) => {
  selectedPreviewBatch.value = batch
}
const closePreview = () => {
  selectedPreviewBatch.value = null
}

// Helper methods for detailed delivery batch view
const getStore = (item: any) => {
  return item.sales_order?.store || item.sales_transaction?.store
}

const getSalesperson = (item: any) => {
  const emp = item.sales_order?.employee || item.sales_transaction?.employee
  if (!emp) return '-'
  return `${emp.first_name || ''} ${emp.last_name || ''}`.trim() || '-'
}

const getItems = (item: any) => {
  return item.sales_order?.items || item.sales_transaction?.items || []
}

const getItemQtyDisplay = (orderItem: any) => {
  if (orderItem.unit && orderItem.ordered_quantity !== undefined) {
    return `${orderItem.ordered_quantity} ${orderItem.unit}`
  }
  const unit = orderItem.product?.unit || 'PCS'
  return `${orderItem.quantity} ${unit}`
}

const getOrderTotalAmount = (item: any) => {
  return item.sales_order?.total_amount || item.sales_transaction?.total_amount || 0
}

const getBatchTotalAmount = (batch: any) => {
  if (!batch.items) return 0
  return batch.items.reduce((sum: number, item: any) => {
    return sum + getOrderTotalAmount(item)
  }, 0)
}

const confirmDialog = ref({
  open: false,
  title: '',
  description: '',
  variant: 'info' as 'danger' | 'warning' | 'success' | 'info',
  confirmText: '',
  cancelText: '',
  isLoading: false,
  onConfirm: () => {},
})

// Load pending requests, delivery batches, and performance list
const reloadData = async () => {
  loading.value = true
  try {
    // 1. Fetch pending supervisor transfers (Barang Bawaan Sales)
    const transfers = await transferStore.fetchSupervisorPending(
      masterDataStore.selectedBranchId
    )
    pendingRequests.value = transfers || []

    // Initialize the approved quantities to the original requested quantities
    pendingRequests.value.forEach((req) => {
      approvedQuantities[req.id] = req.quantity
    })

    // 2. Fetch pending delivery batches
    await deliveryStore.fetchBatches(masterDataStore.selectedBranchId)

    // 3. Fetch sales KPIs for alignment context
    const month = new Date().getMonth() + 1
    const year = new Date().getFullYear()
    
    // Calculate previous month for carry-over targets
    let prevMonth = month - 1
    let prevYear = year
    if (prevMonth === 0) {
      prevMonth = 12
      prevYear = year - 1
    }

    // Build branch query parameter
    let branchFilter = ''
    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      branchFilter = `&branch_id=${masterDataStore.selectedBranchId}`
    }

    const [perfRes, prevPerfRes] = await Promise.all([
      apiClient.get(`/admin/sales/reports/performance?month=${month}&year=${year}${branchFilter}`),
      apiClient.get(`/admin/sales/reports/performance?month=${prevMonth}&year=${prevYear}${branchFilter}`)
    ])
    
    const currentKPIs = perfRes.data.data || []
    const prevKPIs = prevPerfRes.data.data || []
    
    // Merge KPIs: if current KPI doesn't exist or has 0 target, try to use previous month's target
    performanceList.value = currentKPIs.map((kpi: any) => {
      if (!kpi.TargetOmzet || kpi.TargetOmzet === 0) {
        const prevKpi = prevKPIs.find((p: any) => p.EmployeeID === kpi.EmployeeID)
        if (prevKpi) {
          kpi.TargetOmzet = prevKpi.TargetOmzet
        }
      }
      return kpi
    })
    
    // Add missing KPIs from previous month (so they show up with 0 achieved but correct target)
    prevKPIs.forEach((prevKpi: any) => {
      const exists = performanceList.value.find((p: any) => p.EmployeeID === prevKpi.EmployeeID)
      if (!exists) {
        performanceList.value.push({
          ...prevKpi,
          AchievedOmzet: 0,
          AchievedOmzetLama: 0,
          AchievedOmzetBaru: 0,
          ID: undefined // Clear ID so it's treated as carry-over
        })
      }
    })
  } catch (err: any) {
    toast.error(err.message || 'Gagal memuat data persetujuan')
  } finally {
    loading.value = false
  }
}

// Search filter left: Sales request
const filteredRequests = computed(() => {
  if (!searchSalesQuery.value.trim()) return pendingRequests.value

  const query = searchSalesQuery.value.toLowerCase().trim()
  return pendingRequests.value.filter((req) => {
    const salesmanName = `${req.employee?.first_name} ${req.employee?.last_name}`.toLowerCase()
    const productName = (req.product?.name || '').toLowerCase()
    const sku = (req.product?.sku || '').toLowerCase()
    return salesmanName.includes(query) || productName.includes(query) || sku.includes(query)
  })
})

// Search filter right: Delivery Batches
const filteredBatches = computed(() => {
  const allWaiting = deliveryStore.batches.filter((b: any) => b.status === 'WAITING_APPROVAL')
  if (!searchBatchQuery.value.trim()) return allWaiting

  const query = searchBatchQuery.value.toLowerCase().trim()
  return allWaiting.filter((b: any) => {
    const doNo = (b.delivery_order_no || '').toLowerCase()
    const driverName = b.driver ? `${b.driver.first_name} ${b.driver.last_name || ''}`.toLowerCase() : ''
    const helperName = b.helper ? `${b.helper.first_name} ${b.helper.last_name || ''}`.toLowerCase() : ''
    const plate = b.vehicle ? (b.vehicle.license_plate || '').toLowerCase() : ''
    const hasOrderMatch = b.items?.some((item: any) => {
      const storeName = (item.sales_order?.store?.name || '').toLowerCase()
      const soNum = (item.sales_order?.so_number || '').toLowerCase()
      return storeName.includes(query) || soNum.includes(query)
    })
    return doNo.includes(query) || driverName.includes(query) || helperName.includes(query) || plate.includes(query) || hasOrderMatch
  })
})

// Calculations
const getRequestValue = (req: any) => {
  const price = req.product?.selling_price || 0
  return req.quantity * price
}

const getSalesmanKPI = (employeeId: string) => {
  if (!employeeId) return null;
  return performanceList.value.find((p) => {
    return p.EmployeeID === employeeId || 
           p.employee_id === employeeId || 
           p.employeeId === employeeId ||
           p.Employee?.ID === employeeId ||
           p.Employee?.id === employeeId ||
           p.employee?.id === employeeId ||
           p.employee?.ID === employeeId ||
           p.employee?.employee_id === employeeId;
  });
}

const getKpiAchieved = (kpi: any) => {
  if (!kpi) return 0;
  return kpi.AchievedOmzet ?? kpi.achieved_omzet ?? kpi.achievedOmzet ?? 0;
}

const getKpiTarget = (kpi: any) => {
  if (!kpi) return 0;
  return kpi.TargetOmzet ?? kpi.target_omzet ?? kpi.targetOmzet ?? 0;
}

const getKPIPercent = (employeeId: string) => {
  const kpi = getSalesmanKPI(employeeId)
  if (!kpi) return 0
  const target = getKpiTarget(kpi);
  const achieved = getKpiAchieved(kpi);
  if (!target) return 0
  return (achieved / target) * 100
}

const adjustApprovedQty = (id: string, amount: number) => {
  const current = approvedQuantities[id] || 0
  const adjusted = current + amount
  if (adjusted >= 1) {
    approvedQuantities[id] = adjusted
  }
}

// Actions Left: Sales Request Approve/Reject
const handleApprove = async (req: any) => {
  const approvedQty = approvedQuantities[req.id] || req.quantity
  loading.value = true
  try {
    await transferStore.approveSupervisorTransfer(req.id, approvedQty)
    toast.success(`Permintaan barang oleh ${req.employee?.first_name} berhasil disetujui!`)
    await reloadData()
  } catch (err: any) {
    toast.error(err || 'Gagal menyetujui permintaan')
  } finally {
    loading.value = false
  }
}

const confirmReject = (req: any) => {
  confirmDialog.value = {
    open: true,
    title: 'Tolak Permintaan Barang',
    description: `Apakah Anda yakin ingin menolak permintaan ${req.quantity} ${req.product?.unit} ${req.product?.name} oleh ${req.employee?.first_name}?`,
    variant: 'danger',
    confirmText: 'Ya, Tolak',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await transferStore.rejectSupervisorTransfer(req.id)
        toast.success(`Permintaan barang berhasil ditolak.`)
        confirmDialog.value.open = false
        await reloadData()
      } catch (err: any) {
        toast.error(err || 'Gagal menolak permintaan')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

// Actions Right: Delivery Batch Approve/Delete (Cancel)
const handleApproveBatch = async (batch: any) => {
  loading.value = true
  try {
    await apiClient.post(`/delivery/batch/${batch.id}/supervisor-approve`)
    toast.success(`Batch pengiriman ${batch.delivery_order_no || batch.id} berhasil disetujui! Surat Jalan terbit.`)
    await reloadData()
  } catch (err: any) {
    const errMsg = err.response?.data?.error || err.message || 'Gagal menyetujui pengiriman'
    toast.error(errMsg)
  } finally {
    loading.value = false
  }
}

const confirmRejectBatch = (batch: any) => {
  confirmDialog.value = {
    open: true,
    title: 'Batalkan Batch Pengiriman',
    description: `Apakah Anda yakin ingin membatalkan batch pengiriman ${batch.delivery_order_no || 'DRAFT-BATCH'}? Tindakan ini akan menghapus rancangan batch dan mengembalikan order ke antrean.`,
    variant: 'danger',
    confirmText: 'Ya, Batalkan',
    cancelText: 'Batal',
    isLoading: false,
    onConfirm: async () => {
      confirmDialog.value.isLoading = true
      try {
        await apiClient.delete(`/delivery/batches/${batch.id}`)
        toast.success(`Batch pengiriman berhasil dibatalkan.`)
        confirmDialog.value.open = false
        await reloadData()
      } catch (err: any) {
        toast.error(err.message || 'Gagal membatalkan batch')
      } finally {
        confirmDialog.value.isLoading = false
      }
    },
  }
}

// Helpers
const getInitials = (employee: any) => {
  if (!employee) return 'S'
  const first = employee.first_name?.[0] || ''
  const last = employee.last_name?.[0] || ''
  return (first + last).toUpperCase() || 'S'
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(val)
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

watch(
  () => masterDataStore.selectedBranchId,
  () => {
    reloadData()
  },
)

onMounted(async () => {
  await reloadData()
})
</script>

<style scoped>
/* Custom styled scrollbar for listing height limits */
.max-h-\[65vh\]::-webkit-scrollbar {
  width: 5px;
}
.max-h-\[65vh\]::-webkit-scrollbar-track {
  background: transparent;
}
.max-h-\[65vh\]::-webkit-scrollbar-thumb {
  background: #e2e8f0;
  border-radius: 9999px;
}
.max-h-\[65vh\]::-webkit-scrollbar-thumb:hover {
  background: #cbd5e1;
}
</style>
