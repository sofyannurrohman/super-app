<template>
  <div class="p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Main UI (Hidden when printing) -->
    <div class="no-print space-y-8">
      <!-- Header -->
      <div
        class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-8 rounded-[2rem] shadow-sm border border-slate-100"
      >
        <div>
          <div
            class="flex items-center gap-2 text-xs font-bold text-orange-600 uppercase tracking-[0.2em] mb-2"
          >
            <Truck class="h-4 w-4" />
            <span>Logistics & Distribution</span>
          </div>
          <h1 class="text-3xl font-black tracking-tight text-slate-900">Manajemen Pengiriman</h1>
          <p class="text-slate-500 mt-1">
            Kelola surat jalan dan proses permintaan barang dari cabang.
          </p>
        </div>
        <div class="flex gap-3">
          <Button
            variant="outline"
            class="rounded-2xl border-slate-200 bg-white hover:bg-slate-50 text-slate-600 font-bold"
            @click="factoryStore.fetchAllTransfers()"
          >
            <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': factoryStore.loading }" />
            Segarkan
          </Button>
          <Button
            @click="openManualShipmentModal"
            class="rounded-2xl shadow-lg shadow-primary/20 font-bold"
          >
            <Plus class="mr-2 h-4 w-4" /> Buat Surat Jalan Manual
          </Button>
        </div>
      </div>

      <!-- Main Content Layout -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
        <!-- LEFT SIDE: ACTIVE & COMPLETED SHIPMENTS (8/12) -->
        <div class="lg:col-span-8 space-y-6">
          <DataTable
            :columns="shipmentColumns"
            :data="processedShipments"
            :isLoading="factoryStore.loading"
            class="shadow-xl shadow-slate-200/50 rounded-[2rem] overflow-hidden"
          >
            <template #headerTitle>
              <h2 class="text-xl font-black text-slate-900 flex items-center gap-2">
                Daftar Surat Jalan
                <Badge variant="secondary" class="rounded-full bg-slate-100 text-slate-500">{{
                  processedShipments.length
                }}</Badge>
              </h2>
            </template>

            <template #cell-doNo="{ row }">
              <div class="flex items-center gap-3">
                <div
                  class="bg-slate-50 p-2 rounded-xl border border-slate-100 group-hover:bg-white group-hover:shadow-sm transition-all"
                >
                  <Truck class="h-4 w-4 text-primary" />
                </div>
                <div>
                  <p class="text-[11px] font-black text-slate-900 font-mono tracking-tight flex items-center gap-2">
                    {{ row.doNo }}
                    <Badge v-if="row.items[0]?.fulfillment_branch_id" variant="secondary" class="text-[8px] bg-purple-50 text-purple-700 border-purple-200">🏪 Cabang Delegasi</Badge>
                    <Badge v-else variant="secondary" class="text-[8px] bg-slate-100 text-slate-600 border-slate-200">🏭 Pabrik Pusat</Badge>
                  </p>
                  <div
                    v-if="row.items[0]?.target_shipment_date"
                    class="flex items-center gap-1 text-[9px] font-bold text-slate-400 mt-0.5"
                  >
                    <Calendar class="h-2.5 w-2.5" />
                    {{
                      new Date(row.items[0].target_shipment_date).toLocaleDateString('id-ID', {
                        day: 'numeric',
                        month: 'short',
                      })
                    }}
                  </div>
                </div>
              </div>
            </template>

            <template #cell-to="{ row }">
              <div class="space-y-1">
                <p class="text-xs font-black text-slate-700 truncate max-w-[120px]">{{ row.to }}</p>
                <div
                  v-if="row.items[0]?.driver"
                  class="flex items-center gap-1 text-[9px] font-bold text-slate-400"
                >
                  <User class="h-2.5 w-2.5" /> {{ row.items[0].driver.first_name }}
                </div>
              </div>
            </template>

            <template #cell-items="{ row }">
              <div class="space-y-2">
                <div class="flex flex-col gap-1.5">
                  <div
                    v-for="(item, idx) in row.items.slice(0, 3)"
                    :key="item.id"
                    class="text-[10px] bg-slate-50/80 px-2.5 py-2 rounded-lg border border-slate-100/60 shadow-sm"
                  >
                    <ProductHoverCard
                      :product="item.product || getProductById(item.product_id)"
                      :batches="getBatchesForProduct(item.product_id)"
                      stock-location="Pabrik"
                    >
                      <span class="font-bold text-slate-700 block truncate cursor-pointer hover:text-primary transition-colors">{{
                        item.product?.name || 'Produk'
                      }}</span>
                    </ProductHoverCard>
                    <div class="flex items-center justify-between mt-1">
                      <span class="text-slate-500 font-medium"
                        >{{ item.quantity }} {{ item.unit || item.product?.unit || 'UNIT' }}
                        <span
                          v-if="
                            (item.unit || item.product?.unit) !== 'PCS' &&
                            (item.pcs_per_unit || item.product?.pcs_per_unit || 1) > 1
                          "
                          class="text-slate-400"
                          >(Isi {{ item.pcs_per_unit || item.product?.pcs_per_unit || 1 }})</span
                        ></span
                      >
                      <span class="font-black text-primary"
                        >{{
                          item.quantity * (item.pcs_per_unit || item.product?.pcs_per_unit || 1)
                        }}
                        PCS</span
                      >
                    </div>
                  </div>
                  <div
                    v-if="row.items.length > 3"
                    class="text-[9px] font-bold text-slate-400 text-center py-0.5 bg-slate-50 rounded-md border border-slate-100 border-dashed"
                  >
                    + {{ row.items.length - 3 }} item lainnya
                  </div>
                </div>
                <div class="flex items-center justify-between pt-1.5 border-t border-slate-100">
                  <div class="flex items-center gap-1.5">
                    <Badge
                      variant="secondary"
                      class="text-[9px] font-black bg-blue-50 text-blue-600 px-2 py-0.5 rounded-md border border-blue-100/50"
                    >
                      Total:
                      {{
                        row.items.reduce(
                          (acc: number, curr: any) =>
                            acc + curr.quantity * (curr.pcs_per_unit || 1),
                          0,
                        )
                      }}
                      PCS
                    </Badge>
                    <Badge
                      variant="outline"
                      class="text-[9px] font-black bg-slate-50 text-slate-500 px-2 py-0.5 rounded-md border-slate-200"
                    >
                      {{ row.items.length }} SKU
                    </Badge>
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="text-[10px] font-black text-slate-700">{{ row.totalWeight.toFixed(2) }} KG</span>
                    <span v-if="row.totalVolume > 0" class="text-[10px] font-bold text-blue-600 bg-blue-50 px-1.5 py-0.5 rounded">📦 {{ row.totalVolume.toFixed(3) }} m³</span>
                  </div>
                </div>
              </div>
            </template>

            <template #cell-billing="{ row }">
              <div class="space-y-1.5">
                <div v-if="row.transferType === 'INTERNAL_TRANSFER' || row.paymentMethod === 'INTERNAL' || !row.isBillable">
                  <Badge variant="outline" class="text-[9px] font-black bg-emerald-50 text-emerald-700 border-emerald-200 px-2 py-0.5 rounded-md">
                    🔄 Mutasi Bebas Hutang
                  </Badge>
                  <p class="text-[9px] font-bold text-slate-400 mt-1">Non-Tagihan Cabang</p>
                </div>
                <div v-else>
                  <p class="text-xs font-black text-slate-800">
                    Rp {{ row.totalAmount?.toLocaleString('id-ID') }}
                  </p>
                  <div class="flex flex-col gap-1">
                    <span class="text-[9px] font-bold text-slate-500 flex items-center gap-1">
                      Metode: <Badge variant="outline" class="text-[8px] px-1 py-0">{{ row.paymentMethod }}</Badge>
                    </span>
                    <span class="text-[9px] font-bold text-slate-500 flex items-center gap-1">
                      Status: 
                      <Badge :variant="row.paymentStatus === 'PAID' ? 'success' : 'secondary'" class="text-[8px] px-1 py-0">
                        {{ row.paymentStatus === 'PAID' ? 'LUNAS' : 'BELUM BAYAR' }}
                      </Badge>
                    </span>
                  </div>
                </div>
              </div>
            </template>

            <template #cell-status="{ row }">
              <div class="flex flex-col gap-1.5 items-start">
                <Badge
                  :class="getStatusClass(row.status)"
                  class="rounded-lg px-2.5 py-1 text-[9px] font-black uppercase tracking-wider border-none shadow-sm"
                >
                  {{ row.status }}
                </Badge>
                <div
                  v-if="(row.status === 'APPROVED' || row.status === 'SHIPPED' || row.status === 'RECEIVED') && (row.items?.[0]?.approved_by_name || row.items?.[0]?.approved_by)"
                  class="flex flex-col text-[10px] text-slate-500 leading-tight"
                >
                  <span class="font-bold text-slate-700">
                    ✓ {{ row.items?.[0]?.approved_by_name || `${row.items?.[0]?.approved_by?.first_name || ''} ${row.items?.[0]?.approved_by?.last_name || ''}`.trim() }}
                  </span>
                  <span class="text-[9px] text-slate-400">
                    {{ row.items?.[0]?.approved_by_position || row.items?.[0]?.approved_by?.job_position?.title || 'Admin Pabrik' }}
                  </span>
                </div>
                <div
                  v-if="row.status === 'REJECTED' && row.rejectionReason"
                  class="text-[9px] font-bold text-rose-500 bg-rose-50 px-2 py-1 rounded-lg border border-rose-100 max-w-[150px] leading-snug shadow-sm"
                >
                  <span class="text-[7px] uppercase font-black block mb-0.5 opacity-60"
                    >Alasan Penolakan:</span
                  >
                  {{ row.rejectionReason }}
                </div>
              </div>
            </template>

            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1">
                <template
                  v-if="
                    row.status === 'REQUESTED' ||
                    row.status === 'APPROVED' ||
                    row.status === 'REJECTED'
                  "
                >
                  <Button
                    v-if="row.status === 'APPROVED'"
                    @click="processShipment(row)"
                    size="sm"
                    class="bg-primary hover:bg-primary/90 text-white font-black h-8 px-3 rounded-lg text-[9px] shadow-sm tracking-tighter"
                  >
                    KIRIM
                  </Button>
                  <Button
                    @click="openEditModal(row)"
                    variant="ghost"
                    size="sm"
                    class="h-8 w-8 text-slate-400 hover:text-primary hover:bg-slate-50 rounded-lg shrink-0"
                  >
                    <RefreshCw class="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    @click="handleDeleteShipment(row)"
                    variant="ghost"
                    size="sm"
                    class="h-8 w-8 text-slate-400 hover:text-destructive hover:bg-destructive/5 rounded-lg shrink-0"
                  >
                    <Trash2 class="h-3.5 w-3.5" />
                  </Button>
                </template>
                <Button
                  @click="printShipment(row)"
                  variant="ghost"
                  size="sm"
                  class="h-8 w-8 text-slate-400 hover:text-slate-600 hover:bg-slate-50 rounded-lg shrink-0"
                >
                  <Printer class="h-3.5 w-3.5" />
                </Button>
              </div>
            </template>
          </DataTable>
        </div>

        <!-- RIGHT SIDE: INCOMING REQUESTS (4/12) -->
        <div class="lg:col-span-4 space-y-6">
          <div class="flex items-center justify-between px-2">
            <h2 class="text-xl font-black text-slate-900 flex items-center gap-2 text-orange-600">
              Permintaan Cabang
              <Badge
                variant="secondary"
                class="rounded-full bg-orange-100 text-orange-600 border-none"
                >{{ incomingRequests.length }}</Badge
              >
            </h2>
          </div>

          <div
            v-if="incomingRequests.length === 0"
            class="bg-white rounded-[2rem] p-12 text-center border border-dashed border-slate-200"
          >
            <p class="text-sm text-slate-400">Tidak ada permintaan barang baru.</p>
          </div>

          <div
            v-for="request in incomingRequests"
            :key="request.doNo"
            class="bg-white rounded-3xl border border-orange-100 shadow-sm overflow-hidden"
          >
            <div
              class="p-5 bg-orange-50/50 border-b border-orange-100 flex justify-between items-center"
            >
              <div>
                <p class="text-[10px] font-bold text-orange-400 uppercase tracking-widest">
                  ID Permintaan
                </p>
                <h4 class="text-sm font-black text-orange-900 font-mono">{{ request.doNo }}</h4>
              </div>
              <div class="flex flex-col items-end gap-1">
                <Badge
                  :class="getStatusClass(request.status)"
                  class="border-none text-[9px] font-bold uppercase"
                  >{{ request.status }}</Badge
                >
                <div
                  v-if="request.status === 'REJECTED' && request.rejectionReason"
                  class="text-[9px] font-bold text-rose-500 bg-rose-50/50 px-2 py-1 rounded-lg border border-rose-100 max-w-[150px] leading-snug mt-1 shadow-sm text-left"
                >
                  <span class="text-[7px] uppercase font-black block mb-0.5 opacity-60"
                    >Alasan Penolakan:</span
                  >
                  {{ request.rejectionReason }}
                </div>
              </div>
            </div>
            <div class="p-5 space-y-4">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p class="text-[10px] font-bold text-slate-400 uppercase mb-1">Tujuan Pengiriman</p>
                  <div class="flex items-center gap-2">
                    <MapPin class="h-4 w-4 text-primary" />
                    <span class="text-sm font-black text-slate-700">{{ request.to }}</span>
                  </div>
                </div>
                <div v-if="request.items[0]?.target_shipment_date">
                  <p class="text-[10px] font-bold text-slate-400 uppercase mb-1">Target Pengiriman</p>
                  <div class="flex items-center gap-1.5 text-blue-600 bg-blue-50 px-2.5 py-1 rounded-xl w-fit">
                    <Calendar class="h-3.5 w-3.5" />
                    <span class="text-xs font-black">
                      {{
                        new Date(request.items[0].target_shipment_date).toLocaleDateString('id-ID', {
                          day: '2-digit',
                          month: 'short',
                          year: 'numeric',
                        })
                      }}
                    </span>
                  </div>
                </div>
              </div>

              <div class="space-y-2">
                <div
                  v-for="item in request.items"
                  :key="item.id"
                  class="bg-slate-50 p-4 rounded-2xl border border-slate-100 space-y-2"
                >
                  <div class="flex justify-between items-start gap-2">
                    <ProductHoverCard
                      :product="item.product || getProductById(item.product_id)"
                      :batches="getBatchesForProduct(item.product_id, item.id)"
                      :required-quantity="item.quantity * (item.pcs_per_unit || item.product?.pcs_per_unit || 1)"
                      stock-location="Pabrik"
                      highlight-deficient
                      class="flex-1"
                    >
                      <div class="cursor-pointer group flex items-start gap-1">
                        <span class="text-xs font-black text-slate-700 group-hover:text-primary transition-colors leading-tight">
                          {{ item.product?.name }}
                        </span>
                        <Info class="w-3 h-3 text-slate-400 group-hover:text-primary shrink-0 mt-0.5" />
                      </div>
                    </ProductHoverCard>
                    <span class="text-xs font-black text-primary shrink-0">
                      {{ item.quantity }} {{ item.unit || item.product?.unit }}
                      <span
                        v-if="(item.pcs_per_unit || item.product?.pcs_per_unit || 1) > 1"
                        class="text-slate-400 font-normal"
                      >
                        ({{ item.quantity }} x
                        {{ item.pcs_per_unit || item.product?.pcs_per_unit }} PCS)
                      </span>
                    </span>
                  </div>

                  <!-- Quick Stock Warning Badge in Request Card -->
                  <div class="flex items-center justify-between gap-2 pt-1.5 border-t border-slate-200/60 text-[10px]">
                    <div class="flex items-center gap-1.5">
                      <span class="text-slate-400 font-bold uppercase tracking-wider text-[9px]">Stok Pabrik:</span>
                      <template v-if="getBatchesForProduct(item.product_id, item.id).length === 0">
                        <span class="px-1.5 py-0.5 rounded bg-rose-100 text-rose-700 font-bold text-[9px] flex items-center gap-1">
                          <AlertTriangle class="w-2.5 h-2.5" /> Kosong
                        </span>
                      </template>
                      <template v-else-if="(() => {
                        const totalAvail = getBatchesForProduct(item.product_id, item.id).reduce((sum: number, b: any) => sum + (b.available_quantity || 0), 0);
                        return totalAvail < (item.quantity * (item.pcs_per_unit || item.product?.pcs_per_unit || 1));
                      })()">
                        <span class="px-1.5 py-0.5 rounded bg-amber-100 text-amber-800 font-bold text-[9px] flex items-center gap-1">
                          <AlertTriangle class="w-2.5 h-2.5" /> Kurang ({{ getBatchesForProduct(item.product_id, item.id).reduce((sum: number, b: any) => sum + (b.available_quantity || 0), 0) }} PCS)
                        </span>
                      </template>
                      <template v-else>
                        <span class="text-emerald-700 font-bold text-[9px]">
                          {{ getBatchesForProduct(item.product_id, item.id).reduce((sum: number, b: any) => sum + (b.available_quantity || 0), 0) }} PCS
                        </span>
                      </template>
                    </div>
                    <span class="text-[9px] font-bold text-slate-400 uppercase">
                      Total: {{ item.quantity * (item.pcs_per_unit || item.product?.pcs_per_unit || 1) }} PCS
                    </span>
                  </div>
                </div>
              </div>

              <div v-if="request.status === 'REQUESTED'">
                <template v-if="request.items[0]?.initiated_by === 'BRANCH'">
                  <Button
                    @click="handleApproveRequest(request)"
                    class="w-full rounded-2xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold h-12 shadow-lg shadow-emerald-200 mt-2"
                  >
                    Setujui Permintaan
                  </Button>
                </template>
                <template v-else>
                  <div
                    class="w-full py-4 text-center text-[10px] font-black text-orange-500 uppercase tracking-widest bg-orange-50 rounded-2xl border border-orange-100 mt-2"
                  >
                    Menunggu Approval Cabang
                  </div>
                </template>
              </div>

              <div v-else-if="request.status === 'APPROVED'">
                <Button
                  @click="processShipment(request)"
                  class="w-full rounded-2xl bg-primary hover:bg-primary/90 text-white font-bold h-12 shadow-lg shadow-primary/20 mt-2"
                >
                  Proses & Kirim (Buat SJ)
                </Button>
              </div>

              <!-- CRUD Actions for Factory Initiated Shipments -->
              <div
                v-if="
                  request.status === 'REQUESTED' ||
                  request.status === 'APPROVED' ||
                  request.status === 'REJECTED'
                "
                class="grid grid-cols-2 gap-2 mt-4 pt-4 border-t border-slate-100"
              >
                <Button
                  variant="outline"
                  size="sm"
                  @click="openEditModal(request)"
                  class="rounded-xl font-bold text-slate-600 hover:bg-slate-50 border-slate-200 h-10"
                >
                  <RefreshCw class="h-3.5 w-3.5 mr-2" /> Edit
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  @click="handleDeleteShipment(request)"
                  class="rounded-xl font-bold text-destructive hover:bg-destructive/5 border-destructive/20 h-10"
                >
                  <Trash2 class="h-3.5 w-3.5 mr-2" /> Hapus
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- PRINT AREA (Only visible when printing) -->
    <div class="print-only hidden print:block">
      <SuratJalanPrint
        v-if="shipmentToPrint"
        :shipment="shipmentToPrint"
        class="hidden print:block"
      />
    </div>

    <ConfirmationDialog
      v-model:open="confirmDialog.open"
      :title="confirmDialog.title"
      :description="confirmDialog.description"
      :variant="confirmDialog.variant"
      :confirm-text="confirmDialog.confirmText"
      :is-loading="processing"
      @confirm="confirmDialog.onConfirm"
      @cancel="confirmDialog.open = false"
    />

    <div
      v-if="showProcessModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4 md:p-6"
    >
      <Card
        class="w-full max-w-4xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-300 flex flex-col max-h-[90vh]"
      >
        <CardHeader class="p-8 border-b bg-white shrink-0">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900"
                >Proses & Terbitkan Surat Jalan</CardTitle
              >
              <CardDescription
                >Sesuaikan item, tentukan jadwal, dan assign armada pengiriman.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showProcessModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-8 bg-white overflow-y-auto custom-scrollbar">
          <!-- Summary Info -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="bg-blue-50 p-6 rounded-3xl border border-blue-100 flex items-center gap-4">
              <div
                class="bg-white h-12 w-12 rounded-2xl flex items-center justify-center shadow-sm"
              >
                <Building2 class="h-6 w-6 text-blue-600" />
              </div>
              <div>
                <p class="text-[10px] font-bold text-blue-400 uppercase tracking-widest">
                  Tujuan Cabang
                </p>
                <h4 class="text-lg font-black text-blue-900">{{ selectedRequest?.to }}</h4>
              </div>
            </div>
            <div
              class="bg-slate-50 p-6 rounded-3xl border border-slate-100 flex items-center gap-4"
            >
              <div
                class="bg-white h-12 w-12 rounded-2xl flex items-center justify-center shadow-sm"
              >
                <FileText class="h-6 w-6 text-slate-600" />
              </div>
              <div>
                <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest">
                  ID Permintaan
                </p>
                <h4 class="text-lg font-black text-slate-900 font-mono">
                  {{ selectedRequest?.doNo }}
                </h4>
              </div>
            </div>
          </div>

          <!-- Pilihan Jenis Pengiriman -->
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Jenis Pengiriman</label>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <label
                class="flex items-start gap-3 p-4 rounded-2xl border cursor-pointer transition-all"
                :class="processForm.transfer_type === 'COMMERCIAL' ? 'border-primary bg-primary/5 text-primary' : 'border-slate-100 bg-slate-50/50 text-slate-600 hover:border-slate-200'"
              >
                <input
                  type="radio"
                  value="COMMERCIAL"
                  v-model="processForm.transfer_type"
                  class="mt-1"
                />
                <div>
                  <div class="font-black text-xs">📦 Beli Cabang (Tagihan Pabrik)</div>
                  <div class="text-[10px] text-slate-500 mt-0.5">Dikenakan hutang/piutang ke cabang</div>
                </div>
              </label>

              <label
                class="flex items-start gap-3 p-4 rounded-2xl border cursor-pointer transition-all"
                :class="processForm.transfer_type === 'INTERNAL_TRANSFER' ? 'border-emerald-500 bg-emerald-50/50 text-emerald-700' : 'border-slate-100 bg-slate-50/50 text-slate-600 hover:border-slate-200'"
              >
                <input
                  type="radio"
                  value="INTERNAL_TRANSFER"
                  v-model="processForm.transfer_type"
                  class="mt-1"
                />
                <div>
                  <div class="font-black text-xs">🔄 Titip Stok / Mutasi Internal</div>
                  <div class="text-[10px] text-slate-500 mt-0.5">Pindah stok tanpa hutang (Bebas Tagihan)</div>
                </div>
              </label>
            </div>
          </div>

          <!-- Logistics Assignment -->
          <div class="space-y-6">
            <p
              class="text-[11px] font-black text-primary uppercase tracking-[0.2em] flex items-center gap-2"
            >
              <Truck class="h-4 w-4" /> Penugasan Logistik
            </p>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="space-y-2">
                <label class="text-sm font-bold text-slate-700 ml-1">Asal Gudang Pengiriman</label>
                <select
                  v-model="processForm.fulfillment_branch_id"
                  class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
                >
                  <option value="">🏢 Pusat (Pabrik)</option>
                  <option
                    v-for="branch in masterDataStore.branches"
                    :key="branch.id"
                    :value="branch.id"
                  >
                    🏪 Cabang ({{ branch.name }})
                  </option>
                </select>
              </div>
              <div class="space-y-2" v-if="!processForm.fulfillment_branch_id">
                <label class="text-sm font-bold text-slate-700 ml-1">Armada (Kendaraan)</label>
                <select
                  v-model="processForm.vehicle_id"
                  class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
                >
                  <option value="">Pilih Armada...</option>
                  <option
                    v-for="v in availableVehicles"
                    :key="v.id"
                    :value="v.id"
                    :disabled="!v.isAvailable"
                  >
                    {{ v.name }} ({{ v.license_plate }}) - Cap: {{ v.capacity }} KG
                    {{ !v.isAvailable ? '(Sedang Digunakan)' : '' }}
                  </option>
                </select>
              </div>
              <div class="space-y-2" v-if="!processForm.fulfillment_branch_id">
                <label class="text-sm font-bold text-slate-700 ml-1">Sopir (Driver)</label>
                <div class="flex gap-2">
                  <select
                    v-model="processForm.driver_id"
                    class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
                  >
                    <option value="">Pilih Sopir...</option>
                    <option
                      v-for="e in availableDrivers"
                      :key="e.id"
                      :value="e.id"
                      :disabled="!e.isAvailable"
                    >
                      {{ e.first_name }} {{ e.last_name }}
                      {{ !e.isAvailable ? '(Sedang Mengirim)' : '' }}
                    </option>
                  </select>
                  <Button
                    variant="outline"
                    class="rounded-2xl shrink-0 h-[50px] font-bold border-dashed border-2 hover:bg-slate-50 text-slate-600"
                    @click="showAddDriverModal = true"
                  >
                    <Plus class="h-4 w-4 mr-1" /> Sopir
                  </Button>
                </div>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <label class="text-sm font-bold text-slate-700 ml-1">Tgl Kirim</label>
                  <input
                    type="datetime-local"
                    v-model="processForm.target_shipment_date"
                    class="w-full px-4 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-xs font-bold outline-none"
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-sm font-bold text-slate-700 ml-1">Est. Sampai</label>
                  <input
                    type="datetime-local"
                    v-model="processForm.estimated_arrival"
                    class="w-full px-4 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-xs font-bold outline-none"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Item Adjustments -->
          <div class="space-y-4 pt-4 border-t border-slate-50">
            <div class="flex flex-wrap justify-between items-center gap-2">
              <p
                class="text-[11px] font-black text-orange-600 uppercase tracking-[0.2em] flex items-center gap-2"
              >
                <Package class="h-4 w-4" /> Daftar Item & Penyesuaian
              </p>
              <div class="flex items-center gap-2">
                <Button
                  v-if="!processForm.fulfillment_branch_id"
                  size="sm"
                  type="button"
                  variant="outline"
                  @click="autoAllocateAllProcessItemsFEFO"
                  class="rounded-xl font-bold h-9 border-amber-300 text-amber-800 bg-amber-50/80 hover:bg-amber-100 shadow-sm"
                  title="Otomatis bagi batch berdasarkan tanggal kadaluarsa / stok terlama (FEFO)"
                >
                  <Zap class="h-4 w-4 mr-1 text-amber-600 fill-amber-500" /> Auto-Alokasi Semua (FEFO)
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  @click="addProcessItem"
                  class="rounded-xl font-bold h-9"
                >
                  <Plus class="h-4 w-4 mr-1" /> Tambah Produk Lain
                </Button>
              </div>
            </div>

            <div class="space-y-3">
              <div
                v-for="(item, index) in processForm.items"
                :key="index"
                class="bg-slate-50/50 p-5 rounded-[2rem] border border-slate-100 group transition-all hover:bg-white hover:shadow-sm"
              >
                <div class="grid grid-cols-1 md:grid-cols-12 gap-4 items-end">
                  <div class="md:col-span-3 space-y-1">
                    <div class="flex items-center justify-between ml-1">
                      <label
                        class="text-[10px] font-bold text-slate-400 uppercase tracking-widest"
                        >Produk</label
                      >
                      <ProductHoverCard
                        v-if="item.product_id"
                        :product="getProductById(item.product_id)"
                        :batches="getBatchesForProduct(item.product_id, item.id, processForm.fulfillment_branch_id)"
                        :required-quantity="item.quantity * item.pcs_per_unit"
                        :stock-location="processForm.fulfillment_branch_id ? 'Cabang Delegasi' : 'Pabrik'"
                        highlight-deficient
                      >
                        <button type="button" class="text-[9px] font-bold text-primary hover:underline flex items-center gap-0.5">
                          <Info class="w-3 h-3" /> Detail
                        </button>
                      </ProductHoverCard>
                    </div>
                    <select
                      v-model="item.product_id"
                      :disabled="item.original"
                      class="w-full px-4 py-3 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none disabled:opacity-60"
                    >
                      <option v-for="p in factoryStore.sellableProducts" :key="p.id" :value="p.id">
                        {{ p.name }}
                      </option>
                    </select>
                  </div>
                  <div class="md:col-span-3 space-y-1">
                    <div class="flex items-center justify-between ml-1">
                      <label
                        class="text-[10px] font-bold text-slate-400 uppercase tracking-widest"
                        >Batch</label
                      >
                      <button
                        v-if="!processForm.fulfillment_branch_id && item.product_id"
                        type="button"
                        @click="openMultiBatchModalForProcess(index)"
                        class="text-[9px] font-bold text-blue-600 hover:text-blue-800 flex items-center gap-0.5 transition-colors"
                        title="Pilih dan bagi alokasi dari beberapa batch"
                      >
                        <Layers class="w-3 h-3" /> Multi-Batch
                      </button>
                    </div>
                    <div
                      v-if="processForm.fulfillment_branch_id"
                      class="w-full px-4 py-3 bg-amber-50/80 border border-amber-200/70 rounded-xl text-xs font-bold text-amber-900 flex items-center justify-between"
                    >
                      <span class="truncate">🏷️ Diatur Gudang Cabang</span>
                      <span class="text-[10px] font-black text-amber-700 bg-amber-100/80 px-2 py-0.5 rounded-md shrink-0">
                        {{ getTotalBranchStock(item.product_id, processForm.fulfillment_branch_id) }} PCS
                      </span>
                    </div>
                    <select
                      v-else
                      v-model="item.batch_no"
                      class="w-full px-4 py-3 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                    >
                      <option value="" disabled>Pilih Batch...</option>
                      <option
                        v-for="s in getBatchesForProduct(item.product_id, item.id)"
                        :key="s.batch_no"
                        :value="s.batch_no"
                      >
                        Batch {{ s.batch_no }} (Sisa: {{ s.available_quantity }})
                      </option>
                    </select>
                  </div>
                  <div class="md:col-span-2 space-y-1">
                    <label
                      class="text-[10px] font-bold text-slate-400 uppercase tracking-widest ml-1"
                      >Jumlah</label
                    >
                    <input
                      type="number"
                      v-model.number="item.quantity"
                      class="w-full px-4 py-3 bg-white border border-slate-100 rounded-xl text-xs font-black outline-none"
                    />
                  </div>
                  <div class="md:col-span-2 space-y-1">
                    <label
                      class="text-[10px] font-bold text-slate-400 uppercase tracking-widest ml-1"
                      >Satuan</label
                    >
                    <select
                      v-model="item.unit"
                      @change="handleUnitChange(item)"
                      class="w-full px-4 py-3 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                    >
                      <option v-for="u in UNIT_OPTIONS" :key="u" :value="u">{{ u }}</option>
                    </select>
                  </div>
                  <div class="md:col-span-2 space-y-1">
                    <label
                      class="text-[10px] font-bold text-slate-400 uppercase tracking-widest ml-1"
                      >PCS / {{ item.unit }}</label
                    >
                    <input
                      type="number"
                      :disabled="isBaseUnit(item.unit)"
                      v-model.number="item.pcs_per_unit"
                      class="w-full px-4 py-3 bg-white border border-slate-100 rounded-xl text-xs font-black outline-none disabled:bg-slate-50 disabled:text-slate-400"
                    />
                  </div>
                  <div class="md:col-span-1 flex items-end">
                    <Button
                      variant="ghost"
                      size="icon"
                      @click="removeProcessItem(index)"
                      class="h-11 w-full text-rose-400 hover:text-rose-600 hover:bg-rose-50 rounded-xl"
                    >
                      <Trash2 class="h-5 w-5" />
                    </Button>
                  </div>
                </div>
                <div
                  v-if="item.product_id"
                  class="mt-3 flex items-center justify-between px-3 py-2 bg-white/50 rounded-xl border border-dashed border-slate-200"
                >
                  <div class="text-[10px] font-bold text-slate-500 flex items-center flex-wrap gap-1.5">
                    <span>Total: <strong class="text-primary font-black">{{ item.quantity * item.pcs_per_unit }} PCS</strong></span>
                    <span class="text-slate-300">|</span>
                    <span>Est. Berat: <strong class="text-slate-900 font-black">{{ calculateItemWeight(item) }} KG</strong></span>
                    <span class="text-slate-300">|</span>
                    <span>Est. Kubikasi: <strong class="text-blue-600 font-black">{{ calculateItemVolume(item) }} m³</strong></span>
                  </div>
                  <div
                    v-if="item.original && item.quantity !== item.original_quantity"
                    class="text-[10px] font-black text-orange-500 italic"
                  >
                    * Diubah dari permintaan asli ({{ item.original_quantity }})
                  </div>
                </div>

                <!-- Warning Alerts for Stock -->
                <div v-if="item.product_id" class="space-y-2">
                  <!-- Check if branch total stock is deficient when delegated -->
                  <div
                    v-if="processForm.fulfillment_branch_id && getTotalBranchStock(item.product_id, processForm.fulfillment_branch_id) < (item.quantity * item.pcs_per_unit)"
                    class="mt-2 text-rose-500 font-bold text-xs flex items-center gap-1.5 bg-rose-50 border border-rose-100 rounded-xl px-4 py-2.5 animate-pulse"
                  >
                    <AlertTriangle class="h-4 w-4 shrink-0 text-rose-500" />
                    <span>
                      Total stok untuk <strong>{{ getProductById(item.product_id)?.name || 'produk ini' }}</strong> di cabang tidak mencukupi!
                      (Tersedia: {{ getTotalBranchStock(item.product_id, processForm.fulfillment_branch_id) }} PCS, Butuh: {{ item.quantity * item.pcs_per_unit }} PCS)
                    </span>
                  </div>
                  <!-- Check if factory has no stock -->
                  <div
                    v-else-if="!processForm.fulfillment_branch_id && getBatchesForProduct(item.product_id, item.id).length === 0"
                    class="mt-2 text-rose-500 font-bold text-xs flex items-center gap-1.5 bg-rose-50 border border-rose-100 rounded-xl px-4 py-2.5 animate-pulse"
                  >
                    <AlertTriangle class="h-4 w-4 shrink-0 text-rose-500" />
                    <span>
                      Stok untuk
                      <ProductHoverCard
                        :product="getProductById(item.product_id)"
                        :batches="getBatchesForProduct(item.product_id, item.id)"
                        :required-quantity="item.quantity * item.pcs_per_unit"
                        stock-location="Pabrik"
                        highlight-deficient
                      >
                        <span class="underline decoration-dotted font-black cursor-pointer text-rose-700 hover:text-rose-900">
                          {{ getProductById(item.product_id)?.name || 'produk ini' }}
                        </span>
                      </ProductHoverCard>
                      belum memiliki stok tersedia di pabrik! (Stok Tersedia: 0)
                    </span>
                  </div>
                  <!-- Check if selected batch has insufficient stock in factory -->
                  <div
                    v-else-if="!processForm.fulfillment_branch_id && item.batch_no && (() => {
                      const selectedBatch = getBatchesForProduct(item.product_id, item.id).find((b) => (b.batch_no || 'DEFAULT').trim() === (item.batch_no || 'DEFAULT').trim());
                      return selectedBatch && selectedBatch.available_quantity < (item.quantity * item.pcs_per_unit);
                    })()"
                    class="mt-2 text-amber-800 font-medium text-xs flex flex-col sm:flex-row sm:items-center justify-between gap-2 bg-amber-50 border border-amber-200 rounded-xl p-3"
                  >
                    <div class="flex items-center gap-2 flex-1 min-w-0">
                      <AlertTriangle class="h-4 w-4 shrink-0 text-amber-600" />
                      <span>
                        Stok batch <strong>{{ item.batch_no }}</strong> tidak mencukupi
                        (Tersedia: {{ getBatchesForProduct(item.product_id, item.id).find((b) => (b.batch_no || 'DEFAULT').trim() === (item.batch_no || 'DEFAULT').trim())?.available_quantity }} PCS, 
                        Butuh: {{ item.quantity * item.pcs_per_unit }} PCS)
                      </span>
                    </div>
                    <div class="flex items-center gap-1.5 shrink-0">
                      <Button
                        type="button"
                        size="sm"
                        @click="autoAllocateItemFEFO(index)"
                        class="h-7 px-2.5 text-[11px] font-bold bg-amber-600 hover:bg-amber-700 text-white rounded-lg shadow-sm"
                        title="Bagi batch secara otomatis sesuai tanggal kadaluarsa terlama"
                      >
                        <Zap class="w-3 h-3 mr-1 fill-white" /> Bagi Batch FEFO
                      </Button>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        @click="openMultiBatchModalForProcess(index)"
                        class="h-7 px-2.5 text-[11px] font-bold bg-white text-slate-700 border-amber-300 hover:bg-amber-100/50 rounded-lg"
                        title="Pilih proporsi batch secara manual"
                      >
                        <Layers class="w-3 h-3 mr-1 text-blue-600" /> Multi-Batch
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Dual Metric: Tonase & Kubikasi Summary Card -->
            <div class="space-y-3">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Kolom 1: Tonase -->
                <div
                  class="p-4 rounded-2xl border transition-all"
                  :class="isProcessOverWeight ? 'bg-red-50/70 border-red-200' : 'bg-slate-50 border-slate-200'"
                >
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2.5">
                      <div class="p-2 rounded-xl bg-white shadow-sm">
                        <Scale class="h-4 w-4" :class="isProcessOverWeight ? 'text-red-500' : 'text-emerald-600'" />
                      </div>
                      <div>
                        <p class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Tonase</p>
                        <p class="text-base font-black" :class="isProcessOverWeight ? 'text-red-600' : 'text-slate-900'">
                          {{ totalProcessWeight.toFixed(2) }} KG
                        </p>
                      </div>
                    </div>
                    <div v-if="selectedProcessVehicle" class="text-right">
                      <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Kapasitas Truk</p>
                      <p class="text-sm font-black text-slate-700">{{ selectedProcessVehicle.capacity || 0 }} KG</p>
                    </div>
                  </div>
                  <div v-if="selectedProcessVehicle && selectedProcessVehicle.capacity > 0" class="space-y-1">
                    <div class="w-full bg-slate-200/80 rounded-full h-2 overflow-hidden">
                      <div
                        class="h-full transition-all duration-300 rounded-full"
                        :class="processWeightPercentage > 100 ? 'bg-red-500' : processWeightPercentage > 85 ? 'bg-amber-500' : 'bg-emerald-500'"
                        :style="{ width: `${Math.min(processWeightPercentage, 100)}%` }"
                      ></div>
                    </div>
                    <div class="flex justify-between items-center text-[10px] font-bold">
                      <span :class="isProcessOverWeight ? 'text-red-600' : 'text-slate-500'">
                        Keterisian: {{ processWeightPercentage }}%
                      </span>
                      <span v-if="isProcessOverWeight" class="text-red-600 font-black">
                        (+{{ (totalProcessWeight - selectedProcessVehicle.capacity).toFixed(1) }} KG Over)
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Kolom 2: Kubikasi -->
                <div
                  class="p-4 rounded-2xl border transition-all"
                  :class="isProcessOverVolume ? 'bg-red-50/70 border-red-200' : 'bg-slate-50 border-slate-200'"
                >
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2.5">
                      <div class="p-2 rounded-xl bg-white shadow-sm">
                        <Package class="h-4 w-4" :class="isProcessOverVolume ? 'text-red-500' : 'text-blue-600'" />
                      </div>
                      <div>
                        <p class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Kubikasi</p>
                        <p class="text-base font-black" :class="isProcessOverVolume ? 'text-red-600' : 'text-slate-900'">
                          {{ totalProcessVolume.toFixed(4) }} m³
                        </p>
                      </div>
                    </div>
                    <div v-if="selectedProcessVehicle" class="text-right">
                      <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Volume Bak</p>
                      <p class="text-sm font-black text-slate-700">
                        {{ selectedProcessVehicle.bed_volume ? `${selectedProcessVehicle.bed_volume} m³` : 'Belum diisi' }}
                      </p>
                    </div>
                  </div>
                  <div v-if="selectedProcessVehicle && (selectedProcessVehicle.bed_volume || 0) > 0" class="space-y-1">
                    <div class="w-full bg-slate-200/80 rounded-full h-2 overflow-hidden">
                      <div
                        class="h-full transition-all duration-300 rounded-full"
                        :class="processVolumePercentage > 100 ? 'bg-red-500' : processVolumePercentage > 85 ? 'bg-amber-500' : 'bg-blue-500'"
                        :style="{ width: `${Math.min(processVolumePercentage, 100)}%` }"
                      ></div>
                    </div>
                    <div class="flex justify-between items-center text-[10px] font-bold">
                      <span :class="isProcessOverVolume ? 'text-red-600' : 'text-slate-500'">
                        Keterisian: {{ processVolumePercentage }}%
                      </span>
                      <span v-if="isProcessOverVolume" class="text-red-600 font-black">
                        (+{{ (totalProcessVolume - (selectedProcessVehicle.bed_volume || 0)).toFixed(3) }} m³ Over)
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Warning Banner Non-Blocking -->
              <div
                v-if="!processForm.fulfillment_branch_id && selectedProcessVehicle && (isProcessOverWeight || isProcessOverVolume)"
                class="flex items-start gap-2.5 px-4 py-3 bg-amber-50 border border-amber-200 text-amber-900 rounded-2xl text-xs font-semibold"
              >
                <AlertTriangle class="h-4 w-4 shrink-0 text-amber-600 mt-0.5" />
                <div class="space-y-0.5">
                  <p class="font-black text-amber-800 uppercase tracking-wider text-[11px]">
                    Peringatan Kapasitas Armada (Operasional Tetap Dapat Lanjut)
                  </p>
                  <p class="text-[11px] text-amber-900/90 leading-relaxed">
                    <span v-if="isProcessOverWeight && isProcessOverVolume">
                      Muatan melebihi <strong>Tonase ({{ processWeightPercentage }}%)</strong> dan <strong>Volume Bak ({{ processVolumePercentage }}%)</strong>.
                    </span>
                    <span v-else-if="isProcessOverWeight">
                      Muatan melebihi <strong>Tonase ({{ processWeightPercentage }}%)</strong>, namun ruang bak masih mencukupi ({{ processVolumePercentage }}%).
                    </span>
                    <span v-else-if="isProcessOverVolume">
                      Muatan melebihi <strong>Volume Bak ({{ processVolumePercentage }}%)</strong>, namun tonase berat masih aman ({{ processWeightPercentage }}%).
                    </span>
                    Surat Jalan tetap dapat diterbitkan jika armada memungkinkan untuk dimuat.
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Catatan Tambahan untuk SJ</label>
            <textarea
              v-model="processForm.notes"
              rows="2"
              placeholder="Informasi tambahan pengiriman..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none resize-none"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter
          class="bg-slate-50/50 border-t p-8 flex flex-col md:flex-row justify-between items-center gap-4 shrink-0"
        >
          <div class="text-[10px] font-bold text-slate-400 max-w-xs text-center md:text-left">
            <span v-if="processForm.fulfillment_branch_id">
              Pengiriman akan <span class="text-orange-600 font-black">DIDELEGASIKAN</span> ke Cabang. Admin Gudang Cabang akan menentukan batch fisik & armada saat memuat barang.
            </span>
            <span v-else>
              Dengan menerbitkan SJ, stok pabrik akan terpotong dan status pengiriman akan berubah
              menjadi <span class="text-blue-600 font-black">SHIPPED</span>.
            </span>
          </div>
          <div class="flex gap-4 w-full md:w-auto">
            <Button
              variant="ghost"
              @click="showProcessModal = false"
              class="rounded-2xl font-bold h-12 px-8"
              >Batal</Button
            >
            <Button
              @click="handleConfirmProcess"
              :loading="processing"
              class="rounded-2xl px-12 h-12 shadow-xl font-black text-white transition-all"
              :class="processForm.fulfillment_branch_id ? 'bg-orange-500 hover:bg-orange-600 shadow-orange-500/20' : 'bg-primary hover:bg-primary/90 shadow-primary/20'"
            >
              {{ processForm.fulfillment_branch_id ? 'Delegasikan ke Cabang' : 'Terbitkan SJ & Kirim' }}
            </Button>
          </div>
        </CardFooter>
      </Card>
    </div>

    <!-- Manual Shipment Modal -->
    <div
      v-if="showManualModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-3xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900"
                >Buat Surat Jalan Manual</CardTitle
              >
              <CardDescription
                >Kirim barang ke cabang tanpa melalui permintaan sistem.</CardDescription
              >
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="showManualModal = false"
              class="rounded-full"
            >
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[60vh] overflow-y-auto">
          <!-- Pilihan Jenis Pengiriman -->
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Jenis Pengiriman</label>
            <div class="grid grid-cols-2 gap-3">
              <label
                class="flex items-start gap-3 p-3.5 rounded-2xl border-2 cursor-pointer transition-all"
                :class="manualShipmentForm.transfer_type === 'COMMERCIAL' ? 'border-primary bg-primary/5 text-primary' : 'border-slate-100 bg-slate-50/50 text-slate-600 hover:border-slate-200'"
              >
                <input
                  type="radio"
                  value="COMMERCIAL"
                  v-model="manualShipmentForm.transfer_type"
                  class="mt-1"
                />
                <div>
                  <div class="font-black text-xs">📦 Beli Cabang (Tagihan Pabrik)</div>
                  <div class="text-[11px] text-slate-400 mt-0.5">Dikenakan hutang/piutang ke cabang</div>
                </div>
              </label>

              <label
                class="flex items-start gap-3 p-3.5 rounded-2xl border-2 cursor-pointer transition-all"
                :class="manualShipmentForm.transfer_type === 'INTERNAL_TRANSFER' ? 'border-emerald-500 bg-emerald-50/50 text-emerald-700' : 'border-slate-100 bg-slate-50/50 text-slate-600 hover:border-slate-200'"
              >
                <input
                  type="radio"
                  value="INTERNAL_TRANSFER"
                  v-model="manualShipmentForm.transfer_type"
                  class="mt-1 text-emerald-600"
                />
                <div>
                  <div class="font-black text-xs">🔄 Titip Stok / Mutasi Internal</div>
                  <div class="text-[11px] text-slate-400 mt-0.5">Pindah stok tanpa hutang (Bebas Tagihan)</div>
                </div>
              </label>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Asal Gudang Pengiriman</label>
              <select
                v-model="manualShipmentForm.fulfillment_branch_id"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              >
                <option value="">🏢 Pusat (Pabrik)</option>
                <option
                  v-for="branch in masterDataStore.branches"
                  :key="branch.id"
                  :value="branch.id"
                >
                  🏪 Cabang ({{ branch.name }})
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Ke Cabang / Gudang</label>
              <select
                v-model="manualShipmentForm.to_branch_id"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              >
                <option value="" disabled>Pilih Cabang...</option>
                <option
                  v-for="branch in masterDataStore.branches"
                  :key="branch.id"
                  :value="branch.id"
                >
                  {{ branch.name }}
                </option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2" v-if="!manualShipmentForm.fulfillment_branch_id">
              <label class="text-sm font-bold text-slate-700 ml-1 flex items-center gap-2">
                <Truck class="h-4 w-4 text-primary" /> Pilih Armada (Kendaraan)
              </label>
              <select
                v-model="manualShipmentForm.vehicle_id"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              >
                <option value="">Pilih Armada...</option>
                <option
                  v-for="v in availableVehicles"
                  :key="v.id"
                  :value="v.id"
                  :disabled="!v.isAvailable"
                >
                  {{ v.name }} ({{ v.license_plate }}) - Cap: {{ v.capacity }} KG
                  {{ !v.isAvailable ? '(Sedang Digunakan)' : '' }}
                </option>
              </select>
            </div>
            <div class="space-y-2" v-if="!manualShipmentForm.fulfillment_branch_id">
              <label class="text-sm font-bold text-slate-700 ml-1 flex items-center gap-2">
                <User class="h-4 w-4 text-orange-500" /> Pilih Sopir (Driver)
              </label>
              <div class="flex gap-2">
                <select
                  v-model="manualShipmentForm.driver_id"
                  class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
                >
                  <option value="">Pilih Sopir...</option>
                  <option
                    v-for="e in availableDrivers"
                    :key="e.id"
                    :value="e.id"
                    :disabled="!e.isAvailable"
                  >
                    {{ e.first_name }} {{ e.last_name }}
                    {{ !e.isAvailable ? '(Sedang Mengirim)' : '' }}
                  </option>
                </select>
                <Button
                  variant="outline"
                  class="rounded-2xl shrink-0 h-[50px] font-bold border-dashed border-2 hover:bg-slate-50 text-slate-600"
                  @click="showAddDriverModal = true"
                >
                  <Plus class="h-4 w-4 mr-1" /> Sopir
                </Button>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1 flex items-center gap-2">
                <Calendar class="h-4 w-4 text-primary" /> Tanggal Pengiriman
              </label>
              <input
                type="datetime-local"
                v-model="manualShipmentForm.target_shipment_date"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1 flex items-center gap-2">
                <Clock class="h-4 w-4 text-orange-500" /> Estimasi Sampai
              </label>
              <input
                type="datetime-local"
                v-model="manualShipmentForm.estimated_arrival"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              />
            </div>
          </div>

          <div class="space-y-4">
            <div class="flex flex-wrap justify-between items-center gap-2">
              <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-1">
                Daftar Item
              </p>
              <div class="flex items-center gap-2">
                <Button
                  v-if="!manualShipmentForm.fulfillment_branch_id"
                  size="sm"
                  type="button"
                  variant="outline"
                  @click="autoAllocateAllManualItemsFEFO"
                  class="rounded-xl font-bold h-8 border-amber-300 text-amber-800 bg-amber-50/80 hover:bg-amber-100 shadow-sm text-xs"
                  title="Otomatis bagi batch berdasarkan tanggal kadaluarsa / stok terlama (FEFO)"
                >
                  <Zap class="h-3.5 w-3.5 mr-1 text-amber-600 fill-amber-500" /> Auto-Alokasi Semua (FEFO)
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  @click="addManualItem"
                  class="rounded-xl font-bold"
                >
                  <Plus class="h-4 w-4 mr-1" /> Tambah Item
                </Button>
              </div>
            </div>

            <div class="space-y-3">
              <div
                v-for="(item, index) in manualShipmentForm.items"
                :key="index"
                class="bg-slate-50/50 p-4 rounded-2xl border border-slate-100 space-y-3 animate-in fade-in duration-200"
              >
                <div class="flex gap-3 items-end">
                  <div class="flex-1 space-y-1">
                    <div class="flex items-center justify-between">
                      <label class="text-[10px] font-bold text-slate-400 uppercase">Produk</label>
                      <ProductHoverCard
                        v-if="item.product_id"
                        :product="getProductById(item.product_id)"
                        :batches="getBatchesForProduct(item.product_id, undefined, manualShipmentForm.fulfillment_branch_id)"
                        :required-quantity="item.quantity * item.pcs_per_unit"
                        :stock-location="manualShipmentForm.fulfillment_branch_id ? 'Cabang Delegasi' : 'Pabrik'"
                        highlight-deficient
                      >
                        <button type="button" class="text-[9px] font-bold text-primary hover:underline flex items-center gap-0.5">
                          <Info class="w-3 h-3" /> Detail
                        </button>
                      </ProductHoverCard>
                    </div>
                    <select
                      v-model="item.product_id"
                      class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                    >
                      <option v-for="p in factoryStore.sellableProducts" :key="p.id" :value="p.id">
                        {{ p.name }}
                      </option>
                    </select>
                    <div class="text-[9px] font-bold text-slate-400 mt-1 flex items-center gap-1.5 flex-wrap" v-if="item.product_id">
                      <span>Est. Berat: <strong class="text-slate-700 font-bold">{{ calculateItemWeight(item) }} KG</strong></span>
                      <span class="text-slate-300">•</span>
                      <span>Est. Kubikasi: <strong class="text-blue-600 font-bold">{{ calculateItemVolume(item) }} m³</strong></span>
                    </div>
                  </div>
                  <div class="w-36 space-y-1">
                    <div class="flex items-center justify-between">
                      <label class="text-[10px] font-bold text-slate-400 uppercase">Batch</label>
                      <button
                        v-if="!manualShipmentForm.fulfillment_branch_id && item.product_id"
                        type="button"
                        @click="openMultiBatchModalForManual(index)"
                        class="text-[9px] font-bold text-blue-600 hover:text-blue-800 flex items-center gap-0.5 transition-colors"
                        title="Pilih dan bagi alokasi dari beberapa batch"
                      >
                        <Layers class="w-3 h-3" /> Multi-Batch
                      </button>
                    </div>
                    <div
                      v-if="manualShipmentForm.fulfillment_branch_id"
                      class="w-full px-3 py-2.5 bg-amber-50/80 border border-amber-200/70 rounded-xl text-xs font-bold text-amber-900 flex items-center justify-between"
                    >
                      <span class="truncate">🏷️ Gudang Cabang</span>
                      <span class="text-[10px] font-black text-amber-700 bg-amber-100/80 px-1.5 py-0.5 rounded shrink-0">
                        {{ getTotalBranchStock(item.product_id, manualShipmentForm.fulfillment_branch_id) }}
                      </span>
                    </div>
                    <select
                      v-else
                      v-model="item.batch_no"
                      class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                    >
                      <option value="" disabled>Pilih Batch...</option>
                      <option
                        v-for="s in getBatchesForProduct(item.product_id, undefined)"
                        :key="s.batch_no"
                        :value="s.batch_no"
                      >
                        Batch {{ s.batch_no }} (Sisa: {{ s.available_quantity }})
                      </option>
                    </select>
                  </div>
                  <div class="w-24 space-y-1">
                    <label class="text-[10px] font-bold text-slate-400 uppercase">Satuan</label>
                    <select
                      v-model="item.unit"
                      @change="handleUnitChange(item)"
                      class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                    >
                      <option v-for="u in UNIT_OPTIONS" :key="u" :value="u">{{ u }}</option>
                    </select>
                  </div>
                  <div class="w-28 space-y-1" v-if="!isBaseUnit(item.unit)">
                    <label class="text-[10px] font-bold text-slate-400 uppercase">PCS / {{ item.unit }}</label>
                    <input
                      type="number"
                      v-model.number="item.pcs_per_unit"
                      class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-black outline-none"
                    />
                  </div>
                  <div class="w-24 space-y-1">
                    <label class="text-[10px] font-bold text-slate-400 uppercase">Jumlah</label>
                    <input
                      type="number"
                      v-model.number="item.quantity"
                      class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-black outline-none"
                    />
                  </div>
                  <div class="space-y-1">
                    <label class="text-[10px] font-bold text-slate-400 uppercase opacity-0"
                      >Hapus</label
                    >
                    <Button
                      variant="ghost"
                      size="icon"
                      @click="removeManualItem(index)"
                      :disabled="manualShipmentForm.items.length === 1"
                      class="text-rose-500 hover:text-rose-700 hover:bg-rose-50 rounded-xl"
                    >
                      <Trash2 class="w-4 h-4" />
                    </Button>
                  </div>
                </div>

                <!-- Warning Alerts for Stock in Manual Shipment -->
                <div v-if="item.product_id" class="mt-2 space-y-2">
                  <!-- Check if branch total stock is deficient when delegated -->
                  <div
                    v-if="manualShipmentForm.fulfillment_branch_id && getTotalBranchStock(item.product_id, manualShipmentForm.fulfillment_branch_id) < (item.quantity * item.pcs_per_unit)"
                    class="text-rose-500 font-bold text-xs flex items-center gap-1.5 bg-rose-50 border border-rose-100 rounded-xl px-4 py-2 animate-pulse"
                  >
                    <AlertTriangle class="h-4 w-4 shrink-0 text-rose-500" />
                    <span>
                      Total stok untuk <strong>{{ getProductById(item.product_id)?.name || 'produk ini' }}</strong> di cabang tidak mencukupi!
                      (Tersedia: {{ getTotalBranchStock(item.product_id, manualShipmentForm.fulfillment_branch_id) }} PCS, Butuh: {{ item.quantity * item.pcs_per_unit }} PCS)
                    </span>
                  </div>
                  <!-- Check if factory has no stock -->
                  <div
                    v-else-if="!manualShipmentForm.fulfillment_branch_id && getBatchesForProduct(item.product_id, undefined).length === 0"
                    class="text-rose-500 font-bold text-xs flex items-center gap-1.5 bg-rose-50 border border-rose-100 rounded-xl px-4 py-2 animate-pulse"
                  >
                    <AlertTriangle class="h-4 w-4 shrink-0 text-rose-500" />
                    <span>
                      Stok untuk
                      <ProductHoverCard
                        :product="getProductById(item.product_id)"
                        :batches="getBatchesForProduct(item.product_id, undefined)"
                        :required-quantity="item.quantity * item.pcs_per_unit"
                        stock-location="Pabrik"
                        highlight-deficient
                      >
                        <span class="underline decoration-dotted font-black cursor-pointer text-rose-700 hover:text-rose-900">
                          {{ getProductById(item.product_id)?.name || 'produk ini' }}
                        </span>
                      </ProductHoverCard>
                      belum memiliki stok tersedia di pabrik! (Stok Tersedia: 0)
                    </span>
                  </div>
                  <!-- Check if selected batch has insufficient stock in factory -->
                  <div
                    v-else-if="!manualShipmentForm.fulfillment_branch_id && item.batch_no && (() => {
                      const selectedBatch = getBatchesForProduct(item.product_id, undefined).find((b) => (b.batch_no || 'DEFAULT').trim() === (item.batch_no || 'DEFAULT').trim());
                      return selectedBatch && selectedBatch.available_quantity < (item.quantity * item.pcs_per_unit);
                    })()"
                    class="text-amber-800 font-medium text-xs flex flex-col sm:flex-row sm:items-center justify-between gap-2 bg-amber-50 border border-amber-200 rounded-xl p-3"
                  >
                    <div class="flex items-center gap-2 flex-1 min-w-0">
                      <AlertTriangle class="h-4 w-4 shrink-0 text-amber-600" />
                      <span>
                        Stok batch <strong>{{ item.batch_no }}</strong> tidak mencukupi
                        (Tersedia: {{ getBatchesForProduct(item.product_id, undefined).find((b) => (b.batch_no || 'DEFAULT').trim() === (item.batch_no || 'DEFAULT').trim())?.available_quantity }} PCS, 
                        Butuh: {{ item.quantity * item.pcs_per_unit }} PCS)
                      </span>
                    </div>
                    <div class="flex items-center gap-1.5 shrink-0">
                      <Button
                        type="button"
                        size="sm"
                        @click="autoAllocateManualItemFEFO(index)"
                        class="h-7 px-2.5 text-[11px] font-bold bg-amber-600 hover:bg-amber-700 text-white rounded-lg shadow-sm"
                        title="Bagi batch secara otomatis sesuai tanggal kadaluarsa terlama"
                      >
                        <Zap class="w-3 h-3 mr-1 fill-white" /> Bagi Batch FEFO
                      </Button>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        @click="openMultiBatchModalForManual(index)"
                        class="h-7 px-2.5 text-[11px] font-bold bg-white text-slate-700 border-amber-300 hover:bg-amber-100/50 rounded-lg"
                        title="Pilih proporsi batch secara manual"
                      >
                        <Layers class="w-3 h-3 mr-1 text-blue-600" /> Multi-Batch
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Dual Metric: Tonase & Kubikasi Summary Card -->
            <div class="space-y-3">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Kolom 1: Tonase -->
                <div
                  class="p-4 rounded-2xl border transition-all"
                  :class="isOverWeight ? 'bg-red-50/70 border-red-200' : 'bg-slate-50 border-slate-200'"
                >
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2.5">
                      <div class="p-2 rounded-xl bg-white shadow-sm">
                        <Scale class="h-4 w-4" :class="isOverWeight ? 'text-red-500' : 'text-emerald-600'" />
                      </div>
                      <div>
                        <p class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Tonase</p>
                        <p class="text-base font-black" :class="isOverWeight ? 'text-red-600' : 'text-slate-900'">
                          {{ totalManualWeight.toFixed(2) }} KG
                        </p>
                      </div>
                    </div>
                    <div v-if="selectedVehicle" class="text-right">
                      <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Kapasitas Truk</p>
                      <p class="text-sm font-black text-slate-700">{{ selectedVehicle.capacity || 0 }} KG</p>
                    </div>
                  </div>
                  <div v-if="selectedVehicle && selectedVehicle.capacity > 0" class="space-y-1">
                    <div class="w-full bg-slate-200/80 rounded-full h-2 overflow-hidden">
                      <div
                        class="h-full transition-all duration-300 rounded-full"
                        :class="manualWeightPercentage > 100 ? 'bg-red-500' : manualWeightPercentage > 85 ? 'bg-amber-500' : 'bg-emerald-500'"
                        :style="{ width: `${Math.min(manualWeightPercentage, 100)}%` }"
                      ></div>
                    </div>
                    <div class="flex justify-between items-center text-[10px] font-bold">
                      <span :class="isOverWeight ? 'text-red-600' : 'text-slate-500'">
                        Keterisian: {{ manualWeightPercentage }}%
                      </span>
                      <span v-if="isOverWeight" class="text-red-600 font-black">
                        (+{{ (totalManualWeight - selectedVehicle.capacity).toFixed(1) }} KG Over)
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Kolom 2: Kubikasi -->
                <div
                  class="p-4 rounded-2xl border transition-all"
                  :class="isOverVolume ? 'bg-red-50/70 border-red-200' : 'bg-slate-50 border-slate-200'"
                >
                  <div class="flex items-center justify-between mb-2">
                    <div class="flex items-center gap-2.5">
                      <div class="p-2 rounded-xl bg-white shadow-sm">
                        <Package class="h-4 w-4" :class="isOverVolume ? 'text-red-500' : 'text-blue-600'" />
                      </div>
                      <div>
                        <p class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Kubikasi</p>
                        <p class="text-base font-black" :class="isOverVolume ? 'text-red-600' : 'text-slate-900'">
                          {{ totalManualVolume.toFixed(4) }} m³
                        </p>
                      </div>
                    </div>
                    <div v-if="selectedVehicle" class="text-right">
                      <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Volume Bak</p>
                      <p class="text-sm font-black text-slate-700">
                        {{ selectedVehicle.bed_volume ? `${selectedVehicle.bed_volume} m³` : 'Belum diisi' }}
                      </p>
                    </div>
                  </div>
                  <div v-if="selectedVehicle && (selectedVehicle.bed_volume || 0) > 0" class="space-y-1">
                    <div class="w-full bg-slate-200/80 rounded-full h-2 overflow-hidden">
                      <div
                        class="h-full transition-all duration-300 rounded-full"
                        :class="manualVolumePercentage > 100 ? 'bg-red-500' : manualVolumePercentage > 85 ? 'bg-amber-500' : 'bg-blue-500'"
                        :style="{ width: `${Math.min(manualVolumePercentage, 100)}%` }"
                      ></div>
                    </div>
                    <div class="flex justify-between items-center text-[10px] font-bold">
                      <span :class="isOverVolume ? 'text-red-600' : 'text-slate-500'">
                        Keterisian: {{ manualVolumePercentage }}%
                      </span>
                      <span v-if="isOverVolume" class="text-red-600 font-black">
                        (+{{ (totalManualVolume - (selectedVehicle.bed_volume || 0)).toFixed(3) }} m³ Over)
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Warning Banner Non-Blocking -->
              <div
                v-if="!manualShipmentForm.fulfillment_branch_id && selectedVehicle && (isOverWeight || isOverVolume)"
                class="flex items-start gap-2.5 px-4 py-3 bg-amber-50 border border-amber-200 text-amber-900 rounded-2xl text-xs font-semibold"
              >
                <AlertTriangle class="h-4 w-4 shrink-0 text-amber-600 mt-0.5" />
                <div class="space-y-0.5">
                  <p class="font-black text-amber-800 uppercase tracking-wider text-[11px]">
                    Peringatan Kapasitas Armada (Operasional Tetap Dapat Lanjut)
                  </p>
                  <p class="text-[11px] text-amber-900/90 leading-relaxed">
                    <span v-if="isOverWeight && isOverVolume">
                      Muatan melebihi <strong>Tonase ({{ manualWeightPercentage }}%)</strong> dan <strong>Volume Bak ({{ manualVolumePercentage }}%)</strong>.
                    </span>
                    <span v-else-if="isOverWeight">
                      Muatan melebihi <strong>Tonase ({{ manualWeightPercentage }}%)</strong>, namun ruang bak masih mencukupi ({{ manualVolumePercentage }}%).
                    </span>
                    <span v-else-if="isOverVolume">
                      Muatan melebihi <strong>Volume Bak ({{ manualVolumePercentage }}%)</strong>, namun tonase berat masih aman ({{ manualWeightPercentage }}%).
                    </span>
                    Surat Jalan tetap dapat diterbitkan jika armada memungkinkan untuk dimuat.
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Catatan</label>
            <textarea
              v-model="manualShipmentForm.notes"
              rows="2"
              placeholder="Catatan untuk sopir atau gudang..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none resize-none"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            @click="showManualModal = false"
            class="rounded-2xl font-bold h-12 px-8"
            >Batal</Button
          >
          <Button
            @click="handleManualShipment"
            :loading="processing"
            class="rounded-2xl px-12 h-12 shadow-xl font-black text-white transition-all"
            :class="manualShipmentForm.fulfillment_branch_id ? 'bg-orange-500 hover:bg-orange-600 shadow-orange-500/20' : 'bg-primary hover:bg-primary/90 shadow-primary/20'"
          >
            {{ manualShipmentForm.fulfillment_branch_id ? 'Delegasikan Pengiriman' : 'Terbitkan Surat Jalan' }}
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Edit Shipment Modal -->
    <div
      v-if="showEditModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-3xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900">Edit Pengiriman</CardTitle>
              <CardDescription
                >Ubah detail pengiriman ({{ selectedRequest?.doNo }}).</CardDescription
              >
            </div>
            <Button variant="ghost" size="icon" @click="showEditModal = false" class="rounded-full">
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 space-y-6 bg-white max-h-[70vh] overflow-y-auto">
          <!-- Pilihan Jenis Pengiriman Edit -->
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Jenis Pengiriman</label>
            <div class="grid grid-cols-2 gap-3">
              <label
                class="flex items-start gap-3 p-3.5 rounded-2xl border-2 cursor-pointer transition-all"
                :class="editForm.transfer_type === 'COMMERCIAL' ? 'border-primary bg-primary/5 text-primary' : 'border-slate-100 bg-slate-50/50 text-slate-600 hover:border-slate-200'"
              >
                <input
                  type="radio"
                  value="COMMERCIAL"
                  v-model="editForm.transfer_type"
                  class="mt-1"
                />
                <div>
                  <div class="font-black text-xs">📦 Beli Cabang (Tagihan Pabrik)</div>
                  <div class="text-[11px] text-slate-400 mt-0.5">Dikenakan hutang/piutang ke cabang</div>
                </div>
              </label>

              <label
                class="flex items-start gap-3 p-3.5 rounded-2xl border-2 cursor-pointer transition-all"
                :class="editForm.transfer_type === 'INTERNAL_TRANSFER' ? 'border-emerald-500 bg-emerald-50/50 text-emerald-700' : 'border-slate-100 bg-slate-50/50 text-slate-600 hover:border-slate-200'"
              >
                <input
                  type="radio"
                  value="INTERNAL_TRANSFER"
                  v-model="editForm.transfer_type"
                  class="mt-1 text-emerald-600"
                />
                <div>
                  <div class="font-black text-xs">🔄 Titip Stok / Mutasi Internal</div>
                  <div class="text-[11px] text-slate-400 mt-0.5">Pindah stok tanpa hutang (Bebas Tagihan)</div>
                </div>
              </label>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1">Dari Pabrik</label>
              <select
                v-model="editForm.from_factory_id"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              >
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
              <label class="text-sm font-bold text-slate-700 ml-1">Ke Cabang / Gudang</label>
              <select
                v-model="editForm.to_branch_id"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              >
                <option value="" disabled>Pilih Cabang...</option>
                <option
                  v-for="branch in masterDataStore.branches"
                  :key="branch.id"
                  :value="branch.id"
                >
                  {{ branch.name }}
                </option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1 flex items-center gap-2">
                <Truck class="h-4 w-4 text-primary" /> Pilih Armada (Kendaraan)
              </label>
              <select
                v-model="editForm.vehicle_id"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              >
                <option value="">Pilih Armada...</option>
                <option
                  v-for="v in availableVehicles"
                  :key="v.id"
                  :value="v.id"
                  :disabled="!v.isAvailable"
                >
                  {{ v.name }} ({{ v.license_plate }}) - Cap: {{ v.capacity }} KG
                  {{ !v.isAvailable ? '(Sedang Digunakan)' : '' }}
                </option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1 flex items-center gap-2">
                <User class="h-4 w-4 text-orange-500" /> Pilih Sopir (Driver)
              </label>
              <div class="flex gap-2">
                <select
                  v-model="editForm.driver_id"
                  class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
                >
                  <option value="">Pilih Sopir...</option>
                  <option
                    v-for="e in availableDrivers"
                    :key="e.id"
                    :value="e.id"
                    :disabled="!e.isAvailable"
                  >
                    {{ e.first_name }} {{ e.last_name }}
                    {{ !e.isAvailable ? '(Sedang Mengirim)' : '' }}
                  </option>
                </select>
                <Button
                  variant="outline"
                  class="rounded-2xl shrink-0 h-[50px] font-bold border-dashed border-2 hover:bg-slate-50 text-slate-600"
                  @click="showAddDriverModal = true"
                >
                  <Plus class="h-4 w-4 mr-1" /> Sopir
                </Button>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1 flex items-center gap-2">
                <Calendar class="h-4 w-4 text-primary" /> Tanggal Pengiriman
              </label>
              <input
                type="datetime-local"
                v-model="editForm.target_shipment_date"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              />
            </div>
            <div class="space-y-2">
              <label class="text-sm font-bold text-slate-700 ml-1 flex items-center gap-2">
                <Clock class="h-4 w-4 text-orange-500" /> Estimasi Sampai
              </label>
              <input
                type="datetime-local"
                v-model="editForm.estimated_arrival"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              />
            </div>
          </div>

          <div class="space-y-4">
            <div class="flex justify-between items-center">
              <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-1">
                Edit Item & Jumlah
              </p>
              <Button size="sm" variant="outline" @click="addEditItem" class="rounded-xl font-bold">
                <Plus class="h-4 w-4 mr-1" /> Tambah Item
              </Button>
            </div>
            <div class="space-y-3">
              <div
                v-for="(item, index) in editForm.items"
                :key="index"
                class="flex items-end gap-3 bg-slate-50 p-4 rounded-2xl border border-slate-100"
              >
                <div class="flex-1 space-y-1">
                  <div class="flex items-center justify-between">
                    <label class="text-[10px] font-bold text-slate-400 uppercase">Produk</label>
                    <ProductHoverCard
                      v-if="item.product_id"
                      :product="getProductById(item.product_id)"
                      :batches="getBatchesForProduct(item.product_id, item.id)"
                      :required-quantity="item.quantity * item.pcs_per_unit"
                      stock-location="Pabrik"
                      highlight-deficient
                    >
                      <button type="button" class="text-[9px] font-bold text-primary hover:underline flex items-center gap-0.5">
                        <Info class="w-3 h-3" /> Detail
                      </button>
                    </ProductHoverCard>
                  </div>
                  <select
                    v-model="item.product_id"
                    class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                  >
                    <option v-for="p in factoryStore.sellableProducts" :key="p.id" :value="p.id">
                      {{ p.name }}
                    </option>
                  </select>
                </div>
                <div class="w-32 space-y-1">
                  <label class="text-[10px] font-bold text-slate-400 uppercase">Batch</label>
                  <select
                    v-model="item.batch_no"
                    class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                  >
                    <option value="" disabled>Pilih...</option>
                    <option
                      v-for="s in getBatchesForProduct(item.product_id, item.id)"
                      :key="s.batch_no"
                      :value="s.batch_no"
                    >
                      Batch {{ s.batch_no }} ({{ s.available_quantity }})
                    </option>
                  </select>
                </div>
                <div class="w-24 space-y-1">
                  <label class="text-[10px] font-bold text-slate-400 uppercase">Unit</label>
                  <select
                    v-model="item.unit"
                    @change="handleUnitChange(item)"
                    class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                  >
                    <option v-for="u in UNIT_OPTIONS" :key="u" :value="u">{{ u }}</option>
                  </select>
                </div>
                <div v-if="!isBaseUnit(item.unit)" class="w-24 space-y-1">
                  <label class="text-[10px] font-bold text-slate-400 uppercase"
                    >PCS / {{ item.unit }}</label
                  >
                  <input
                    type="number"
                    v-model.number="item.pcs_per_unit"
                    class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                  />
                </div>
                <div class="w-20 space-y-1">
                  <label class="text-[10px] font-bold text-slate-400 uppercase">Jumlah</label>
                  <input
                    type="number"
                    v-model.number="item.quantity"
                    class="w-full px-4 py-2.5 bg-white border border-slate-100 rounded-xl text-xs font-bold outline-none"
                  />
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  class="text-destructive h-10 w-10 hover:bg-destructive/5"
                  @click="removeEditItem(index)"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">Catatan</label>
            <textarea
              v-model="editForm.notes"
              rows="2"
              placeholder="Catatan untuk sopir atau gudang..."
              class="w-full px-5 py-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm outline-none resize-none"
            ></textarea>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button
            variant="ghost"
            @click="showEditModal = false"
            class="rounded-2xl font-bold h-12 px-8"
            >Batal</Button
          >
          <Button
            @click="handleUpdateShipment"
            :loading="processing"
            class="rounded-2xl px-12 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90"
          >
            Simpan Perubahan
          </Button>
        </CardFooter>
      </Card>
    </div>
    <!-- Modal Tambah Sopir -->
    <!-- Modal Tambah Sopir -->
    <div
      v-if="showAddDriverModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[110] flex items-center justify-center p-4"
    >
      <Card class="w-full max-w-md shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-300">
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-xl font-black text-slate-900">Tambah Sopir Pocokan</CardTitle>
              <CardDescription>Tambahkan data sopir sementara untuk pengiriman ini.</CardDescription>
            </div>
            <Button variant="ghost" size="icon" @click="showAddDriverModal = false" class="rounded-full">
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="p-8 bg-white space-y-4">
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">Nama Lengkap</label>
            <input
              v-model="newDriver.firstName"
              placeholder="Contoh: Budi Santoso"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-xl text-sm font-bold outline-none focus:border-primary"
            />
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700">No. HP / WhatsApp</label>
            <input
              v-model="newDriver.phoneNumber"
              placeholder="Contoh: 08123456789"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-xl text-sm font-bold outline-none focus:border-primary"
            />
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-6 flex justify-end gap-3">
          <Button variant="ghost" @click="showAddDriverModal = false" class="rounded-xl font-bold">
            Batal
          </Button>
          <Button
            @click="submitAddDriver"
            :loading="isSubmittingDriver"
            class="rounded-xl font-bold bg-primary hover:bg-primary/90 text-white"
          >
            Simpan Sopir
          </Button>
        </CardFooter>
      </Card>
    </div>

    <!-- Modal Alokasi Multi-Batch Interaktif -->
    <div
      v-if="showMultiBatchModal && multiBatchContext"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[120] flex items-center justify-center p-4"
    >
      <Card class="w-full max-w-xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-200 bg-white">
        <CardHeader class="p-6 bg-slate-900 text-white">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="p-2.5 bg-blue-500/20 text-blue-400 rounded-2xl border border-blue-400/20">
                <Layers class="w-5 h-5" />
              </div>
              <div>
                <CardTitle class="text-lg font-black text-white">Distribusi Multi-Batch</CardTitle>
                <CardDescription class="text-slate-400 text-xs mt-0.5">
                  Bagi pemenuhan pesanan dari beberapa batch stok pabrik.
                </CardDescription>
              </div>
            </div>
            <Button
              variant="ghost"
              size="icon"
              @click="closeMultiBatchModal"
              class="text-slate-400 hover:text-white hover:bg-white/10 rounded-full h-8 w-8"
            >
              <X class="w-4 h-4" />
            </Button>
          </div>
        </CardHeader>

        <CardContent class="p-6 space-y-4 max-h-[65vh] overflow-y-auto">
          <!-- Info Box: Target & Produk -->
          <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 flex items-center justify-between">
            <div>
              <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Produk</p>
              <p class="text-sm font-black text-slate-800">{{ getProductById(multiBatchContext.productId)?.name || 'Produk' }}</p>
            </div>
            <div class="text-right">
              <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Target Kebutuhan</p>
              <p class="text-base font-black text-primary">{{ multiBatchContext.targetPieces }} PCS</p>
            </div>
          </div>

          <!-- Quick Action: Rekomendasi FEFO -->
          <div class="flex items-center justify-between bg-amber-50/70 p-3.5 rounded-2xl border border-amber-200/70">
            <div class="flex items-center gap-2.5">
              <Zap class="w-4 h-4 text-amber-600 fill-amber-500 shrink-0" />
              <div>
                <p class="text-xs font-black text-amber-900">Rekomendasi Otomatis FEFO</p>
                <p class="text-[10px] text-amber-700">Utamakan batch terdekat expired & produksi terlama.</p>
              </div>
            </div>
            <Button
              size="sm"
              type="button"
              @click="applyFEFOToMultiBatchContext"
              class="rounded-xl font-bold bg-amber-600 hover:bg-amber-700 text-white text-xs h-8 shadow-sm shrink-0"
            >
              Terapkan FEFO
            </Button>
          </div>

          <!-- Daftar Batch -->
          <div class="space-y-2.5">
            <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest ml-1">
              Daftar Batch di Pabrik
            </p>
            <div
              v-if="multiBatchAllocations.length === 0"
              class="p-8 text-center bg-slate-50 rounded-2xl border border-dashed text-slate-400 text-xs font-bold"
            >
              Tidak ada batch dengan stok tersedia untuk produk ini di pabrik.
            </div>
            <div
              v-for="(row, rIdx) in multiBatchAllocations"
              :key="rIdx"
              class="p-3.5 rounded-2xl border transition-all"
              :class="row.allocatedPieces > 0 ? 'bg-blue-50/50 border-blue-200 shadow-sm' : 'bg-slate-50/40 border-slate-100 hover:border-slate-200'"
            >
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div class="space-y-0.5">
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-black text-slate-900 font-mono">Batch {{ row.batch_no }}</span>
                    <Badge v-if="row.allocatedPieces > 0" class="bg-blue-600 text-[8px] font-bold px-1.5 py-0 h-4">Terpilih</Badge>
                  </div>
                  <div class="flex items-center gap-2 text-[10px] text-slate-500 font-medium">
                    <span>Sisa: <strong class="text-slate-800 font-bold">{{ row.available_quantity }} PCS</strong></span>
                    <span v-if="row.expiry_date">• Exp: {{ new Date(row.expiry_date).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' }) }}</span>
                  </div>
                </div>

                <div class="flex items-center gap-2">
                  <div class="flex items-center gap-1.5">
                    <input
                      type="number"
                      min="0"
                      :max="row.available_quantity"
                      v-model.number="row.allocatedPieces"
                      class="w-24 px-3 py-1.5 bg-white border border-slate-200 rounded-xl text-xs font-black text-right outline-none focus:border-blue-500"
                      placeholder="0"
                    />
                    <span class="text-xs font-bold text-slate-400">PCS</span>
                  </div>
                  <Button
                    size="sm"
                    type="button"
                    variant="outline"
                    @click="row.allocatedPieces = Math.min(row.available_quantity, Math.max(0, multiBatchContext.targetPieces - totalMultiBatchAllocated + row.allocatedPieces))"
                    class="h-7 px-2 text-[10px] font-bold rounded-lg border-slate-200 text-slate-600 hover:bg-slate-100"
                    title="Ambil sisa kebutuhan atau maksimal stok"
                  >
                    Max
                  </Button>
                  <Button
                    v-if="row.allocatedPieces > 0"
                    size="icon"
                    type="button"
                    variant="ghost"
                    @click="row.allocatedPieces = 0"
                    class="h-7 w-7 text-rose-400 hover:text-rose-600 rounded-lg"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
            </div>
          </div>

          <!-- Total Allocation Summary -->
          <div
            class="p-4 rounded-2xl border flex items-center justify-between"
            :class="totalMultiBatchAllocated === multiBatchContext.targetPieces ? 'bg-emerald-50 border-emerald-200 text-emerald-900' : totalMultiBatchAllocated > multiBatchContext.targetPieces ? 'bg-red-50 border-red-200 text-red-900' : 'bg-slate-50 border-slate-200 text-slate-700'"
          >
            <div>
              <p class="text-[10px] font-bold uppercase tracking-wider opacity-70">Total Dialokasikan</p>
              <p class="text-base font-black">
                {{ totalMultiBatchAllocated }} / {{ multiBatchContext.targetPieces }} PCS
              </p>
            </div>
            <div class="text-right">
              <Badge
                :variant="totalMultiBatchAllocated === multiBatchContext.targetPieces ? 'default' : 'secondary'"
                class="font-black text-xs"
                :class="totalMultiBatchAllocated === multiBatchContext.targetPieces ? 'bg-emerald-600 text-white' : totalMultiBatchAllocated > multiBatchContext.targetPieces ? 'bg-red-100 text-red-700' : 'bg-amber-100 text-amber-700'"
              >
                {{
                  totalMultiBatchAllocated === multiBatchContext.targetPieces
                    ? '✓ Pas Sesuai Kebutuhan'
                    : totalMultiBatchAllocated > multiBatchContext.targetPieces
                    ? `Kelebihan ${totalMultiBatchAllocated - multiBatchContext.targetPieces} PCS`
                    : `Kurang ${multiBatchContext.targetPieces - totalMultiBatchAllocated} PCS`
                }}
              </Badge>
            </div>
          </div>
        </CardContent>

        <CardFooter class="bg-slate-50 border-t p-5 flex justify-end gap-3">
          <Button variant="ghost" @click="closeMultiBatchModal" class="rounded-xl font-bold">
            Batal
          </Button>
          <Button
            @click="confirmApplyMultiBatch"
            :disabled="totalMultiBatchAllocated <= 0"
            class="rounded-xl font-bold bg-blue-600 hover:bg-blue-700 text-white shadow-lg shadow-blue-500/20"
          >
            <CheckCircle2 class="w-4 h-4 mr-1.5" /> Terapkan Alokasi
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed, reactive, watch } from 'vue'
import { warehouseApi } from '@/api/warehouse'
import apiClient from '@/api/axios'
import { useFactoryStore } from '@/stores/factory'
import { useMasterDataStore } from '@/stores/masterData'
import { useVehicleStore } from '@/stores/vehicle'
import SuratJalanPrint from '@/components/factory/SuratJalanPrint.vue'
import DataTable from '@/components/DataTable.vue'
import { useDeliveryStore } from '@/stores/delivery'
import { Button } from '@/components/ui/button'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from '@/components/ui/dialog'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '@/components/ui/card/index'
import { Badge } from '@/components/ui/badge'
import {
  Truck,
  RefreshCw,
  Plus,
  Printer,
  MapPin,
  Building2,
  Package,
  Trash2,
  ArrowRight,
  Calendar,
  Clock,
  User,
  Scale,
  AlertTriangle,
  Info,
  Zap,
  Layers,
  Sparkles,
  CheckCircle2,
  X,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import ProductHoverCard from '@/components/ProductHoverCard.vue'

const getProductById = (productId: string) => {
  return factoryStore.products.find((p) => p.id === productId)
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

const factoryStore = useFactoryStore()
const masterDataStore = useMasterDataStore()
const deliveryStore = useDeliveryStore()

const shipmentColumns = [
  { key: 'doNo', label: 'No. Surat Jalan' },
  { key: 'to', label: 'Tujuan' },
  { key: 'items', label: 'Item & Muatan' },
  { key: 'billing', label: 'Tagihan' },
  { key: 'status', label: 'Status' },
  { key: 'actions', label: 'Aksi' },
]
const vehicleStore = useVehicleStore()
const showProcessModal = ref(false)
const showManualModal = ref(false)
const processing = ref(false)
const confirmDialog = reactive({
  open: false,
  title: '',
  description: '',
  variant: 'info' as 'danger' | 'warning' | 'success' | 'info',
  confirmText: '',
  onConfirm: () => {},
})
const selectedRequest = ref<any>(null)
const shipmentToPrint = ref<any>(null)

const globalBranchInventory = ref<any[]>([])

const fetchBranchInventory = async () => {
  try {
    const res = await warehouseApi.getInventory('ALL_BRANCHES')
    globalBranchInventory.value = res.data || []
  } catch (err) {
    console.error('Failed to fetch global branch inventory', err)
  }
}

const getTotalBranchStock = (productId: string, fulfillmentBranchId: string) => {
  if (!fulfillmentBranchId) return 0
  const items = globalBranchInventory.value.filter(
    (s: any) =>
      s.branch_id === fulfillmentBranchId &&
      s.product_id === productId &&
      (!s.stock_type || s.stock_type === 'GOOD'),
  )
  return items.reduce((sum: number, item: any) => sum + (item.quantity || 0), 0)
}

const getBatchesForProduct = (productId: string, excludeTransferId?: string, fulfillmentBranchId?: string) => {
  let sourceInventory = factoryStore.inventory
  let isBranch = false
  if (fulfillmentBranchId) {
    sourceInventory = globalBranchInventory.value.filter((s: any) => s.branch_id === fulfillmentBranchId)
    isBranch = true
  }

  // Filter GOOD stock if from branch
  const items = sourceInventory.filter(
    (i: any) => i.product_id === productId && (!isBranch || !i.stock_type || i.stock_type === 'GOOD'),
  )

  // Group by normalized batch_no
  const batchMap = new Map<string, any>()
  for (const item of items) {
    const bNo = (item.batch_no || 'DEFAULT').trim() || 'DEFAULT'
    if (!batchMap.has(bNo)) {
      batchMap.set(bNo, {
        ...item,
        batch_no: bNo,
        quantity: 0,
      })
    }
    batchMap.get(bNo).quantity += item.quantity || 0
  }

  return Array.from(batchMap.values())
    .map((batch: any) => {
      let reservedPieces = 0
      if (!isBranch) {
        // Cari transfer yang masih REQUESTED / APPROVED untuk batch ini di pabrik
        const pendingTransfers = factoryStore.allTransfers.filter(
          (t: any) =>
            t.product_id === productId &&
            ((t.batch_no || 'DEFAULT').trim() === batch.batch_no) &&
            (t.status === 'REQUESTED' || t.status === 'APPROVED') &&
            t.id !== excludeTransferId,
        )

        // Hitung total quantity yang sudah ter-reserve di transfer yang belum dikirim
        reservedPieces = pendingTransfers.reduce(
          (total: number, t: any) => total + t.quantity * (t.pcs_per_unit || 1),
          0,
        )
      }

      return {
        ...batch,
        available_quantity: Math.max(0, batch.quantity - reservedPieces),
      }
    })
    .filter((b: any) => b.available_quantity > 0)
}

const printShipment = (shipment: any) => {
  shipmentToPrint.value = shipment
  // Wait for barcode to render
  setTimeout(() => {
    window.print()
  }, 300)
}

const showAddDriverModal = ref(false)
const isSubmittingDriver = ref(false)
const newDriver = reactive({
  firstName: '',
  phoneNumber: '',
})

const submitAddDriver = async () => {
  if (!newDriver.firstName) {
    toast.error('Nama lengkap harus diisi')
    return
  }
  isSubmittingDriver.value = true
  try {
    let jobPositionId = masterDataStore.jobPositions.find((jp: any) => 
      jp.title.toLowerCase().includes('sopir pabrik') || jp.title.toLowerCase().includes('driver pusat')
    )?.id

    if (!jobPositionId) {
      jobPositionId = masterDataStore.jobPositions.find((jp: any) => 
        jp.title.toLowerCase().includes('driver') || jp.title.toLowerCase().includes('sopir') || jp.title.toLowerCase().includes('kurir')
      )?.id
    }

    if (!jobPositionId) {
      toast.error('Jabatan "Sopir" atau "Driver" tidak ditemukan di master data')
      isSubmittingDriver.value = false
      return
    }

    const uniqueSuffix = Math.floor(Math.random() * 100000)
    const payload = {
      name: newDriver.firstName,
      email: `driver_${uniqueSuffix}@demo.local`,
      phoneNumber: newDriver.phoneNumber,
      jobPositionId: jobPositionId,
      employmentStatus: 'Harian',
      joinDate: new Date().toISOString().split('T')[0],
      employeeIDNumber: `DRV-${uniqueSuffix}`,
      password: 'Password123!',
    }

    await apiClient.post('/employees', payload)
    toast.success('Sopir baru berhasil ditambahkan')
    
    // Refresh employees
    await masterDataStore.fetchEmployees(true)
    
    // Clear form
    newDriver.firstName = ''
    newDriver.phoneNumber = ''
    showAddDriverModal.value = false
  } catch (error: any) {
    toast.error('Gagal menambahkan sopir: ' + getErrorMessage(error))
  } finally {
    isSubmittingDriver.value = false
  }
}

const processForm = reactive({
  factory_id: '',
  fulfillment_branch_id: '',
  transfer_type: 'COMMERCIAL',
  notes: '',
  vehicle_id: '',
  driver_id: '',
  target_shipment_date: '',
  estimated_arrival: '',
  items: [] as any[],
})

const manualShipmentForm = reactive({
  from_factory_id: '',
  to_branch_id: '',
  fulfillment_branch_id: '',
  transfer_type: 'COMMERCIAL',
  notes: '',
  target_shipment_date: '',
  estimated_arrival: '',
  vehicle_id: '',
  driver_id: '',
  items: [] as any[],
})

const addProcessItem = () => {
  processForm.items.push({
    product_id: '',
    batch_no: '',
    quantity: 1,
    unit: 'PCS',
    pcs_per_unit: 1,
    original: false,
  })
}

const removeProcessItem = (index: number) => {
  processForm.items.splice(index, 1)
}

interface BatchAllocation {
  batch_no: string
  quantity: number
  available_quantity: number
  expiry_date?: string
}

const allocateBatchesFEFO = (
  productId: string,
  totalNeededPieces: number,
  excludeTransferId?: string,
  fulfillmentBranchId?: string,
): BatchAllocation[] => {
  const availableBatches = getBatchesForProduct(productId, excludeTransferId, fulfillmentBranchId)
    .sort((a: any, b: any) => {
      // 1. Expiry date ascending (earliest expiring first)
      if (a.expiry_date && b.expiry_date) {
        return new Date(a.expiry_date).getTime() - new Date(b.expiry_date).getTime()
      }
      if (a.expiry_date) return -1
      if (b.expiry_date) return 1
      // 2. Created at ascending (oldest first)
      if (a.created_at && b.created_at) {
        return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      }
      // 3. Batch no alphabetical
      return (a.batch_no || '').localeCompare(b.batch_no || '')
    })

  let remaining = totalNeededPieces
  const allocations: BatchAllocation[] = []

  for (const b of availableBatches) {
    if (remaining <= 0) break
    const take = Math.min(b.available_quantity, remaining)
    if (take > 0) {
      allocations.push({
        batch_no: b.batch_no,
        quantity: take,
        available_quantity: b.available_quantity,
        expiry_date: b.expiry_date,
      })
      remaining -= take
    }
  }

  return allocations
}

const autoAllocateItemFEFO = (index: number) => {
  const item = processForm.items[index]
  if (!item || !item.product_id) {
    toast.error('Pilih produk terlebih dahulu')
    return
  }

  const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
  const neededPieces = item.quantity * pcsPerUnit
  const allocations = allocateBatchesFEFO(item.product_id, neededPieces, item.id, processForm.fulfillment_branch_id)

  if (allocations.length === 0) {
    toast.error(`Tidak ada stok tersedia di pabrik untuk produk '${getProductById(item.product_id)?.name || item.product_id}'`)
    return
  }

  const totalAllocated = allocations.reduce((sum, a) => sum + a.quantity, 0)
  if (totalAllocated < neededPieces) {
    toast.warning(
      `Stok di pabrik hanya mencukupi ${totalAllocated} dari ${neededPieces} PCS yang dibutuhkan. Semua stok batch yang ada telah dialokasikan.`
    )
  }

  const first = allocations[0]
  if (!first) return

  const isOriginalUnitMultiple = item.unit !== 'PCS' && !isBaseUnit(item.unit) && (first.quantity % (item.pcs_per_unit || 1) === 0)

  item.batch_no = first.batch_no
  if (isOriginalUnitMultiple) {
    item.quantity = first.quantity / (item.pcs_per_unit || 1)
  } else {
    item.unit = 'PCS'
    item.pcs_per_unit = 1
    item.quantity = first.quantity
  }

  const newItemsToInsert = allocations.slice(1).map((alloc) => {
    const isMultiple = item.unit !== 'PCS' && !isBaseUnit(item.unit) && (alloc.quantity % (item.pcs_per_unit || 1) === 0)
    return {
      product_id: item.product_id,
      batch_no: alloc.batch_no,
      quantity: isMultiple ? alloc.quantity / (item.pcs_per_unit || 1) : alloc.quantity,
      unit: isMultiple ? item.unit : 'PCS',
      pcs_per_unit: isMultiple ? item.pcs_per_unit || 1 : 1,
      original: false,
    }
  })

  if (newItemsToInsert.length > 0) {
    processForm.items.splice(index + 1, 0, ...newItemsToInsert)
  }

  toast.success(
    `Berhasil mengalokasikan ${allocations.length} batch (${totalAllocated} PCS) secara FEFO untuk ${getProductById(item.product_id)?.name || 'produk'}`
  )
}

const autoAllocateAllProcessItemsFEFO = () => {
  if (processForm.items.length === 0) return
  let allocatedCount = 0

  for (let i = processForm.items.length - 1; i >= 0; i--) {
    const item = processForm.items[i]
    if (!item.product_id) continue

    const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
    const neededPieces = item.quantity * pcsPerUnit
    const availableBatches = getBatchesForProduct(item.product_id, item.id, processForm.fulfillment_branch_id)
    const selectedBatch = availableBatches.find((b: any) => (b.batch_no || 'DEFAULT').trim() === (item.batch_no || 'DEFAULT').trim())

    if (!item.batch_no || !selectedBatch || selectedBatch.available_quantity < neededPieces) {
      autoAllocateItemFEFO(i)
      allocatedCount++
    }
  }

  if (allocatedCount === 0) {
    toast.info('Semua item sudah memiliki batch dengan stok yang mencukupi.')
  }
}

const autoAllocateManualItemFEFO = (index: number) => {
  const item = manualShipmentForm.items[index]
  if (!item || !item.product_id) {
    toast.error('Pilih produk terlebih dahulu')
    return
  }

  const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
  const neededPieces = item.quantity * pcsPerUnit
  const allocations = allocateBatchesFEFO(item.product_id, neededPieces, undefined, manualShipmentForm.fulfillment_branch_id)

  if (allocations.length === 0) {
    toast.error(`Tidak ada stok tersedia di pabrik untuk produk '${getProductById(item.product_id)?.name || item.product_id}'`)
    return
  }

  const totalAllocated = allocations.reduce((sum, a) => sum + a.quantity, 0)
  if (totalAllocated < neededPieces) {
    toast.warning(
      `Stok di pabrik hanya mencukupi ${totalAllocated} dari ${neededPieces} PCS yang dibutuhkan. Semua stok batch yang ada telah dialokasikan.`
    )
  }

  const first = allocations[0]
  if (!first) return

  const isOriginalUnitMultiple = item.unit !== 'PCS' && !isBaseUnit(item.unit) && (first.quantity % (item.pcs_per_unit || 1) === 0)

  item.batch_no = first.batch_no
  if (isOriginalUnitMultiple) {
    item.quantity = first.quantity / (item.pcs_per_unit || 1)
  } else {
    item.unit = 'PCS'
    item.pcs_per_unit = 1
    item.quantity = first.quantity
  }

  const newItemsToInsert = allocations.slice(1).map((alloc) => {
    const isMultiple = item.unit !== 'PCS' && !isBaseUnit(item.unit) && (alloc.quantity % (item.pcs_per_unit || 1) === 0)
    return {
      product_id: item.product_id,
      batch_no: alloc.batch_no,
      quantity: isMultiple ? alloc.quantity / (item.pcs_per_unit || 1) : alloc.quantity,
      unit: isMultiple ? item.unit : 'PCS',
      pcs_per_unit: isMultiple ? item.pcs_per_unit || 1 : 1,
    }
  })

  if (newItemsToInsert.length > 0) {
    manualShipmentForm.items.splice(index + 1, 0, ...newItemsToInsert)
  }

  toast.success(
    `Berhasil mengalokasikan ${allocations.length} batch (${totalAllocated} PCS) secara FEFO untuk ${getProductById(item.product_id)?.name || 'produk'}`
  )
}

const autoAllocateAllManualItemsFEFO = () => {
  if (manualShipmentForm.items.length === 0) return
  let allocatedCount = 0

  for (let i = manualShipmentForm.items.length - 1; i >= 0; i--) {
    const item = manualShipmentForm.items[i]
    if (!item.product_id) continue

    const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
    const neededPieces = item.quantity * pcsPerUnit
    const availableBatches = getBatchesForProduct(item.product_id, undefined, manualShipmentForm.fulfillment_branch_id)
    const selectedBatch = availableBatches.find((b: any) => (b.batch_no || 'DEFAULT').trim() === (item.batch_no || 'DEFAULT').trim())

    if (!item.batch_no || !selectedBatch || selectedBatch.available_quantity < neededPieces) {
      autoAllocateManualItemFEFO(i)
      allocatedCount++
    }
  }

  if (allocatedCount === 0) {
    toast.info('Semua item sudah memiliki batch dengan stok yang mencukupi.')
  }
}

// Multi-Batch Modal States
const showMultiBatchModal = ref(false)
const multiBatchContext = ref<{
  isManual: boolean
  itemIndex: number
  productId: string
  targetPieces: number
  originalUnit: string
  originalPcsPerUnit: number
} | null>(null)

interface MultiBatchRow {
  batch_no: string
  available_quantity: number
  expiry_date?: string
  allocatedPieces: number
}
const multiBatchAllocations = ref<MultiBatchRow[]>([])

const totalMultiBatchAllocated = computed(() => {
  return multiBatchAllocations.value.reduce((sum, row) => sum + (Number(row.allocatedPieces) || 0), 0)
})

const openMultiBatchModalForProcess = (index: number) => {
  const item = processForm.items[index]
  if (!item || !item.product_id) {
    toast.error('Pilih produk terlebih dahulu')
    return
  }

  const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
  const targetPieces = item.quantity * pcsPerUnit

  multiBatchContext.value = {
    isManual: false,
    itemIndex: index,
    productId: item.product_id,
    targetPieces,
    originalUnit: item.unit,
    originalPcsPerUnit: pcsPerUnit,
  }

  const availableBatches = getBatchesForProduct(item.product_id, item.id, processForm.fulfillment_branch_id)
  multiBatchAllocations.value = availableBatches.map((b: any) => {
    let initialAlloc = 0
    if ((b.batch_no || 'DEFAULT').trim() === (item.batch_no || 'DEFAULT').trim()) {
      initialAlloc = Math.min(b.available_quantity, targetPieces)
    }
    return {
      batch_no: b.batch_no,
      available_quantity: b.available_quantity,
      expiry_date: b.expiry_date,
      allocatedPieces: initialAlloc,
    }
  })

  showMultiBatchModal.value = true
}

const openMultiBatchModalForManual = (index: number) => {
  const item = manualShipmentForm.items[index]
  if (!item || !item.product_id) {
    toast.error('Pilih produk terlebih dahulu')
    return
  }

  const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
  const targetPieces = item.quantity * pcsPerUnit

  multiBatchContext.value = {
    isManual: true,
    itemIndex: index,
    productId: item.product_id,
    targetPieces,
    originalUnit: item.unit,
    originalPcsPerUnit: pcsPerUnit,
  }

  const availableBatches = getBatchesForProduct(item.product_id, undefined, manualShipmentForm.fulfillment_branch_id)
  multiBatchAllocations.value = availableBatches.map((b: any) => {
    let initialAlloc = 0
    if ((b.batch_no || 'DEFAULT').trim() === (item.batch_no || 'DEFAULT').trim()) {
      initialAlloc = Math.min(b.available_quantity, targetPieces)
    }
    return {
      batch_no: b.batch_no,
      available_quantity: b.available_quantity,
      expiry_date: b.expiry_date,
      allocatedPieces: initialAlloc,
    }
  })

  showMultiBatchModal.value = true
}

const closeMultiBatchModal = () => {
  showMultiBatchModal.value = false
  multiBatchContext.value = null
  multiBatchAllocations.value = []
}

const applyFEFOToMultiBatchContext = () => {
  if (!multiBatchContext.value) return
  const fefoList = allocateBatchesFEFO(
    multiBatchContext.value.productId,
    multiBatchContext.value.targetPieces,
    multiBatchContext.value.isManual ? undefined : processForm.items[multiBatchContext.value.itemIndex]?.id,
    multiBatchContext.value.isManual ? manualShipmentForm.fulfillment_branch_id : processForm.fulfillment_branch_id
  )

  const fefoMap = new Map(fefoList.map(f => [f.batch_no, f.quantity]))
  for (const row of multiBatchAllocations.value) {
    row.allocatedPieces = fefoMap.get(row.batch_no) || 0
  }
}

const confirmApplyMultiBatch = () => {
  if (!multiBatchContext.value) return
  const validAllocations = multiBatchAllocations.value.filter((r) => r.allocatedPieces > 0)
  if (validAllocations.length === 0) {
    toast.error('Mohon alokasikan minimal 1 batch dengan kuantitas > 0')
    return
  }

  const { isManual, itemIndex, productId, originalUnit, originalPcsPerUnit } = multiBatchContext.value
  const targetFormItems = isManual ? manualShipmentForm.items : processForm.items
  const currentItem = targetFormItems[itemIndex]
  if (!currentItem) return

  const first = validAllocations[0]
  if (!first) return

  const isOriginalUnitMultiple = originalUnit !== 'PCS' && !isBaseUnit(originalUnit) && (first.allocatedPieces % originalPcsPerUnit === 0)

  currentItem.product_id = productId
  currentItem.batch_no = first.batch_no
  if (isOriginalUnitMultiple) {
    currentItem.quantity = first.allocatedPieces / originalPcsPerUnit
    currentItem.unit = originalUnit
    currentItem.pcs_per_unit = originalPcsPerUnit
  } else {
    currentItem.quantity = first.allocatedPieces
    currentItem.unit = 'PCS'
    currentItem.pcs_per_unit = 1
  }

  const newItemsToInsert = validAllocations.slice(1).map((alloc) => {
    const isMultiple = originalUnit !== 'PCS' && !isBaseUnit(originalUnit) && (alloc.allocatedPieces % originalPcsPerUnit === 0)
    return {
      product_id: productId,
      batch_no: alloc.batch_no,
      quantity: isMultiple ? alloc.allocatedPieces / originalPcsPerUnit : alloc.allocatedPieces,
      unit: isMultiple ? originalUnit : 'PCS',
      pcs_per_unit: isMultiple ? originalPcsPerUnit : 1,
      original: false,
    }
  })

  if (newItemsToInsert.length > 0) {
    targetFormItems.splice(itemIndex + 1, 0, ...newItemsToInsert)
  }

  toast.success(`Berhasil menerapkan alokasi untuk ${validAllocations.length} batch (${totalMultiBatchAllocated.value} PCS)`)
  closeMultiBatchModal()
}

const calculateItemWeight = (item: any) => {
  const product = factoryStore.products.find((p) => p.id === item.product_id)
  if (!product) return '0.00'
  let weight = product.weight * item.quantity * item.pcs_per_unit
  if (product.weight_unit?.toUpperCase() === 'GR' || product.weight_unit?.toUpperCase() === 'ML')
    weight /= 1000
  else if (product.weight_unit?.toUpperCase() === 'TON')
    weight *= 1000
  return weight.toFixed(2)
}

const calculateItemVolume = (item: any) => {
  const product = factoryStore.products.find((p) => p.id === item.product_id)
  if (!product) return '0.0000'
  let vol = Number(product.volume || 0)
  if (vol === 0 && product.length && product.width && product.height) {
    vol = (Number(product.length) * Number(product.width) * Number(product.height)) / 1000000
  }
  const totalVol = vol * (Number(item.quantity) || 0)
  return totalVol.toFixed(4)
}

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

const addManualItem = () => {
  manualShipmentForm.items.push({
    product_id: '',
    batch_no: '',
    quantity: 1,
    unit: 'PCS',
    pcs_per_unit: 1,
  })
}

const removeManualItem = (index: number) => {
  manualShipmentForm.items.splice(index, 1)
}

const openManualShipmentModal = () => {
  manualShipmentForm.transfer_type = 'COMMERCIAL'
  manualShipmentForm.items = [
    { product_id: '', batch_no: '', quantity: 1, unit: 'PCS', pcs_per_unit: 1 },
  ]
  manualShipmentForm.notes = ''
  manualShipmentForm.vehicle_id = ''
  manualShipmentForm.driver_id = ''
  manualShipmentForm.fulfillment_branch_id = ''

  if (factoryStore.factories && factoryStore.factories.length > 0) {
    manualShipmentForm.from_factory_id = factoryStore.factories[0]?.id || ''
  }

  fetchBranchInventory()
  showManualModal.value = true
}

const formatToRFC3339 = (dateStr: any) => {
  if (!dateStr) return null
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return null
  return date.toISOString()
}

const handleManualShipment = async () => {
  if (
    !manualShipmentForm.from_factory_id ||
    !manualShipmentForm.to_branch_id ||
    manualShipmentForm.items.length === 0
  ) {
    toast.error('Mohon lengkapi data pengiriman.')
    return
  }

  // Verify all items have a product selected
  for (const item of manualShipmentForm.items) {
    if (!item.product_id) {
      toast.error('Mohon pilih produk untuk semua baris barang.')
      return
    }
  }

  // Stock pre-validation
  if (manualShipmentForm.fulfillment_branch_id) {
    const requiredProductQty: Record<string, number> = {}
    for (const item of manualShipmentForm.items) {
      const product = factoryStore.products.find((p: any) => p.id === item.product_id)
      if (!product) continue
      const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
      const totalPieces = item.quantity * pcsPerUnit
      requiredProductQty[item.product_id] = (requiredProductQty[item.product_id] || 0) + totalPieces
    }

    const branchName =
      masterDataStore.branches.find((b: any) => b.id === manualShipmentForm.fulfillment_branch_id)?.name ||
      'Cabang Delegasi'

    for (const [productId, needed] of Object.entries(requiredProductQty)) {
      const product = factoryStore.products.find((p: any) => p.id === productId)
      const availableTotal = getTotalBranchStock(productId, manualShipmentForm.fulfillment_branch_id)
      if (availableTotal < needed) {
        toast.error(
          `Gagal mendelegasikan: Total stok '${product?.name || productId}' di ${branchName} tidak mencukupi. Butuh: ${needed} PCS, Total Tersedia: ${availableTotal} PCS`,
        )
        return
      }
    }
  } else {
    // Factory stock check
    const requiredPieces: Record<string, { productId: string; batchNo: string; needed: number }> = {}
    for (const item of manualShipmentForm.items) {
      const product = factoryStore.products.find((p: any) => p.id === item.product_id)
      if (!product) continue
      const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
      const totalPieces = item.quantity * pcsPerUnit
      const batchNo = (item.batch_no || 'DEFAULT').trim() || 'DEFAULT'
      const key = `${item.product_id}:::${batchNo}`
      if (!requiredPieces[key]) {
        requiredPieces[key] = { productId: item.product_id, batchNo, needed: 0 }
      }
      requiredPieces[key].needed += totalPieces
    }

    for (const { productId, batchNo, needed } of Object.values(requiredPieces)) {
      const product = factoryStore.products.find((p: any) => p.id === productId)
      const availableBatches = getBatchesForProduct(productId, undefined)
      const matchedBatch = availableBatches.find(
        (b: any) => (b.batch_no || 'DEFAULT').trim() === batchNo,
      )
      const availableStock = matchedBatch ? matchedBatch.available_quantity : 0
      if (availableStock < needed) {
        toast.error(
          `Gagal membuat pengiriman: Stok '${product?.name || productId}' (Batch ${batchNo}) di Pabrik tidak mencukupi. Butuh: ${needed}, Tersedia: ${availableStock}`,
        )
        return
      }
    }
  }

  processing.value = true
  try {
    const isInternal = manualShipmentForm.transfer_type === 'INTERNAL_TRANSFER'
    const itemsToSubmit = manualShipmentForm.items.map((i: any) => ({
      ...i,
      batch_no: manualShipmentForm.fulfillment_branch_id ? '' : (i.batch_no || 'DEFAULT').trim() || 'DEFAULT',
    }))

    await factoryStore.sendToBranch(manualShipmentForm.from_factory_id, {
      to_branch_id: manualShipmentForm.to_branch_id,
      items: itemsToSubmit,
      notes: manualShipmentForm.notes,
      target_shipment_date: formatToRFC3339(manualShipmentForm.target_shipment_date),
      estimated_arrival: formatToRFC3339(manualShipmentForm.estimated_arrival),
      vehicle_id: manualShipmentForm.fulfillment_branch_id ? null : (manualShipmentForm.vehicle_id || null),
      driver_id: manualShipmentForm.fulfillment_branch_id ? null : (manualShipmentForm.driver_id || null),
      fulfillment_branch_id: manualShipmentForm.fulfillment_branch_id || null,
      transfer_type: manualShipmentForm.transfer_type,
      is_billable: !isInternal,
      payment_method: isInternal ? 'INTERNAL' : 'NET',
      initiated_by: 'FACTORY',
    })

    const branchName =
      masterDataStore.branches.find((b: any) => b.id === manualShipmentForm.fulfillment_branch_id)?.name ||
      'Cabang Delegasi'

    toast.success(
      manualShipmentForm.fulfillment_branch_id
        ? `Pengiriman manual berhasil didelegasikan ke Cabang ${branchName}.`
        : 'Pengiriman manual berhasil dibuat.',
    )
    showManualModal.value = false
    await factoryStore.fetchAllTransfers()
  } catch (err) {
    toast.error('Gagal membuat pengiriman: ' + getErrorMessage(err))
  } finally {
    processing.value = false
  }
}

const handleReceiveReturn = async (returnId: string) => {
  confirmDialog.title = 'Konfirmasi Terima Retur'
  confirmDialog.description =
    'Apakah Anda yakin barang rusak ini sudah diterima secara fisik di pabrik?'
  confirmDialog.variant = 'info'
  confirmDialog.confirmText = 'Ya, Terima Barang'
  confirmDialog.onConfirm = async () => {
    processing.value = true
    try {
      await factoryStore.receiveFactoryReturn(returnId)
      toast.success('Retur berhasil diterima. Stok pabrik telah bertambah.')
    } catch (err: any) {
      toast.error('Gagal menerima retur: ' + getErrorMessage(err))
    } finally {
      processing.value = false
      confirmDialog.open = false
    }
  }
  confirmDialog.open = true
}

onMounted(async () => {
  factoryStore.fetchAllTransfers()
  factoryStore.fetchFactories()
  factoryStore.fetchProducts()
  factoryStore.fetchAllInventory()
  factoryStore.fetchAllFactoryReturns()
  masterDataStore.fetchBranches()
  masterDataStore.fetchEmployees()
  masterDataStore.fetchJobPositions()
  vehicleStore.fetchVehicles()
  deliveryStore.fetchBatches()
  fetchBranchInventory()
})

const availableVehicles = computed(() => {
  return vehicleStore.vehicles.map((v) => {
    const isUsedInTransfer = factoryStore.allTransfers.some(
      (t) => t.vehicle_id === v.id && (t.status === 'SHIPPED' || t.status === 'ARRIVED'),
    )
    const isUsedInBatch = deliveryStore.batches.some(
      (b) => b.vehicle_id === v.id && b.status !== 'COMPLETED',
    )
    return {
      ...v,
      isAvailable: !isUsedInTransfer && !isUsedInBatch,
    }
  })
})

const availableDrivers = computed(() => {
  return masterDataStore.employees
    .filter((emp: any) => {
      let jobTitle = (emp.job_position?.title || '').toLowerCase()
      if (!jobTitle && emp.job_position_id) {
        const jp = masterDataStore.jobPositions.find((j: any) => j.id === emp.job_position_id)
        if (jp) jobTitle = jp.title.toLowerCase()
      }

      let branchName = (emp.branch?.name || '').toLowerCase()
      if (!branchName && emp.branch_id) {
        const b = masterDataStore.branches.find((br: any) => br.id === emp.branch_id)
        if (b) branchName = b.name.toLowerCase()
      }

      const isSopirPabrik = jobTitle.includes('driver') || 
                            jobTitle.includes('sopir') || 
                            jobTitle.includes('kurir') || 
                            jobTitle.includes('pengiriman')

      const isCabangPusat = branchName.includes('pusat') || 
                            branchName.includes('pabrik') || 
                            !emp.branch_id

      return isSopirPabrik && isCabangPusat
    })
    .map((e) => {
      const isUsedInTransfer = factoryStore.allTransfers.some(
        (t) => t.driver_id === e.id && (t.status === 'SHIPPED' || t.status === 'ARRIVED'),
      )
      const isUsedInBatch = deliveryStore.batches.some(
        (b) => b.driver_id === e.id && b.status !== 'COMPLETED',
      )
      return {
        ...e,
        isAvailable: !isUsedInTransfer && !isUsedInBatch,
      }
    })
})

const totalManualWeight = computed(() => {
  return manualShipmentForm.items.reduce((total: number, item: any) => {
    return total + parseFloat(calculateItemWeight(item))
  }, 0)
})

const totalManualVolume = computed(() => {
  return manualShipmentForm.items.reduce((total: number, item: any) => {
    return total + parseFloat(calculateItemVolume(item))
  }, 0)
})

const selectedVehicle = computed(() => {
  return vehicleStore.vehicles.find((v) => v.id === manualShipmentForm.vehicle_id)
})

const isOverWeight = computed(() => {
  if (!selectedVehicle.value || !selectedVehicle.value.capacity) return false
  return totalManualWeight.value > selectedVehicle.value.capacity
})

const isOverVolume = computed(() => {
  if (!selectedVehicle.value || !selectedVehicle.value.bed_volume) return false
  return totalManualVolume.value > selectedVehicle.value.bed_volume
})

const manualWeightPercentage = computed(() => {
  if (!selectedVehicle.value || !selectedVehicle.value.capacity) return 0
  return Math.round((totalManualWeight.value / selectedVehicle.value.capacity) * 100)
})

const manualVolumePercentage = computed(() => {
  if (!selectedVehicle.value || !selectedVehicle.value.bed_volume) return 0
  return Math.round((totalManualVolume.value / selectedVehicle.value.bed_volume) * 100)
})

const totalProcessWeight = computed(() => {
  return processForm.items.reduce((total: number, item: any) => {
    return total + parseFloat(calculateItemWeight(item))
  }, 0)
})

const totalProcessVolume = computed(() => {
  return processForm.items.reduce((total: number, item: any) => {
    return total + parseFloat(calculateItemVolume(item))
  }, 0)
})

const selectedProcessVehicle = computed(() => {
  return vehicleStore.vehicles.find((v) => v.id === processForm.vehicle_id)
})

const isProcessOverWeight = computed(() => {
  if (!selectedProcessVehicle.value || !selectedProcessVehicle.value.capacity) return false
  return totalProcessWeight.value > selectedProcessVehicle.value.capacity
})

const isProcessOverVolume = computed(() => {
  if (!selectedProcessVehicle.value || !selectedProcessVehicle.value.bed_volume) return false
  return totalProcessVolume.value > selectedProcessVehicle.value.bed_volume
})

const processWeightPercentage = computed(() => {
  if (!selectedProcessVehicle.value || !selectedProcessVehicle.value.capacity) return 0
  return Math.round((totalProcessWeight.value / selectedProcessVehicle.value.capacity) * 100)
})

const processVolumePercentage = computed(() => {
  if (!selectedProcessVehicle.value || !selectedProcessVehicle.value.bed_volume) return 0
  return Math.round((totalProcessVolume.value / selectedProcessVehicle.value.bed_volume) * 100)
})

const groupedTransfers = computed(() => {
  const groups: Record<string, any> = {}

  factoryStore.allTransfers.forEach((item: any) => {
    const doNo = item.delivery_order_no || 'DRAFT-' + item.id.substring(0, 8)
    // Status disimpan per produk, bukan per SJ. Pisahkan produk yang belum
    // diproses dari produk yang sudah SHIPPED/ARRIVED dalam SJ yang sama.
    const groupKey = `${doNo}::${item.status}`
    if (!groups[groupKey]) {
      groups[groupKey] = {
        doNo,
        from: item.from_factory?.name,
        to: item.to_branch?.name,
        totalWeight: 0,
        totalVolume: 0,
        status: item.status,
        createdAt: item.created_at,
        rejectionReason: item.rejection_reason,
        paymentMethod: item.payment_method,
        paymentStatus: item.payment_status,
        paymentDueAt: item.payment_due_at,
        transferType: item.transfer_type || (item.payment_method === 'INTERNAL' ? 'INTERNAL_TRANSFER' : 'COMMERCIAL'),
        isBillable: item.is_billable !== false && item.payment_method !== 'INTERNAL' && item.transfer_type !== 'INTERNAL_TRANSFER',
        totalAmount: 0,
        items: [],
      }
    }
    groups[groupKey].items.push(item)
    groups[groupKey].totalWeight += Number(item.total_weight || 0)
    groups[groupKey].totalVolume += Number(item.total_volume || parseFloat(calculateItemVolume(item)))
    groups[groupKey].totalAmount += Number(item.total_amount || 0)

    // Keep the earliest created_at for the group
    if (new Date(item.created_at) < new Date(groups[groupKey].createdAt)) {
      groups[groupKey].createdAt = item.created_at
    }
  })

  return Object.values(groups).sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  )
})

const incomingRequests = computed(() => {
  // Requests from Branch that need Factory approval OR have been approved but not yet shipped
  return groupedTransfers.value.filter(
    (g) =>
      (g.status === 'REQUESTED' || g.status === 'APPROVED' || g.status === 'REJECTED') &&
      g.items[0]?.initiated_by === 'BRANCH',
  )
})

const processedShipments = computed(() => {
  // Shipments that this factory initiated or have been shipped/received
  return groupedTransfers.value.filter((g) => {
    // Show if factory initiated (regardless of status)
    const isFactoryInitiated = g.items[0]?.initiated_by === 'FACTORY'
    // Show if it's already in transit or completed
    const isInTransitOrDone =
      g.status === 'SHIPPED' || g.status === 'ARRIVED' || g.status === 'RECEIVED'
    // Show if branch request is approved by factory
    const isBranchApproved = g.status === 'APPROVED' && g.items[0]?.initiated_by === 'BRANCH'

    return isFactoryInitiated || isInTransitOrDone || isBranchApproved
  })
})

const pendingFactoryReturns = computed(() => {
  return factoryStore.factoryReturns.filter((ret) => ret.status === 'SHIPPED')
})

const showEditModal = ref(false)
const editForm = reactive({
  notes: '',
  target_shipment_date: '',
  estimated_arrival: '',
  vehicle_id: '',
  driver_id: '',
  from_factory_id: '',
  to_branch_id: '',
  transfer_type: 'COMMERCIAL',
  items: [] as any[],
})

watch(
  () => processForm.factory_id,
  (newVal) => {
    if (newVal) factoryStore.fetchInventory(newVal)
  },
)

watch(
  () => processForm.fulfillment_branch_id,
  (newVal) => {
    if (newVal) {
      processForm.vehicle_id = ''
      processForm.driver_id = ''
      fetchBranchInventory()
    }
  },
)

watch(
  () => manualShipmentForm.from_factory_id,
  (newVal) => {
    if (newVal) factoryStore.fetchInventory(newVal)
  },
)

watch(
  () => manualShipmentForm.fulfillment_branch_id,
  (newVal) => {
    if (newVal) {
      manualShipmentForm.vehicle_id = ''
      manualShipmentForm.driver_id = ''
      fetchBranchInventory()
    }
  },
)

watch(
  () => editForm.from_factory_id,
  (newVal) => {
    if (newVal) factoryStore.fetchInventory(newVal)
  },
)

const addEditItem = () => {
  editForm.items.push({
    product_id: '',
    quantity: 1,
    unit: 'PCS',
    pcs_per_unit: 1,
  })
}

const removeEditItem = (index: number) => {
  editForm.items.splice(index, 1)
}

const openEditModal = (group: any) => {
  selectedRequest.value = group
  const firstItem = group.items[0]
  editForm.notes = firstItem.notes || ''
  editForm.target_shipment_date = firstItem.target_shipment_date
    ? new Date(firstItem.target_shipment_date).toISOString().slice(0, 16)
    : ''
  editForm.estimated_arrival = firstItem.estimated_arrival
    ? new Date(firstItem.estimated_arrival).toISOString().slice(0, 16)
    : ''
  editForm.vehicle_id = firstItem.vehicle_id || ''
  editForm.driver_id = firstItem.driver_id || ''
  editForm.from_factory_id = firstItem.from_factory_id || firstItem.from_factory?.id || ''
  editForm.to_branch_id = firstItem.to_branch_id || firstItem.to_branch?.id || ''
  editForm.transfer_type = firstItem.transfer_type || (firstItem.payment_method === 'INTERNAL' ? 'INTERNAL_TRANSFER' : 'COMMERCIAL')

  editForm.items = group.items.map((item: any) => ({
    id: item.id,
    product_id: item.product_id,
    batch_no: item.batch_no || '',
    quantity: item.quantity,
    unit: item.unit || 'PCS',
    pcs_per_unit: isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1,
  }))

  showEditModal.value = true
}

const handleUpdateShipment = async () => {
  // Verify all items have a product selected
  for (const item of editForm.items) {
    if (!item.product_id) {
      toast.error('Mohon pilih produk untuk semua baris barang.')
      return
    }
  }

  processing.value = true
  try {
    // 1. Identify removed items
    const currentItemIds = editForm.items.filter((i) => i.id).map((i) => i.id)
    const removedItems = selectedRequest.value.items.filter(
      (i: any) => !currentItemIds.includes(i.id),
    )

    for (const item of removedItems) {
      await factoryStore.deleteTransfer(item.id)
    }

    const isInternal = editForm.transfer_type === 'INTERNAL_TRANSFER'

    // 2. Update existing or create new items
    for (const item of editForm.items) {
      const originalItem = selectedRequest.value?.items?.find((i: any) => i.id === item.id)
      const payload: any = {
        product_id: item.product_id,
        quantity: item.quantity,
        unit: item.unit,
        pcs_per_unit: item.pcs_per_unit,
        notes: editForm.notes,
        target_shipment_date: formatToRFC3339(editForm.target_shipment_date),
        estimated_arrival: formatToRFC3339(editForm.estimated_arrival),
        vehicle_id: editForm.vehicle_id || null,
        driver_id: editForm.driver_id || null,
        from_factory_id: editForm.from_factory_id || null,
        to_branch_id: editForm.to_branch_id || null,
        status: originalItem?.status || selectedRequest.value?.status || 'REQUESTED',
        rejection_reason: originalItem?.rejection_reason || '',
        initiated_by: originalItem?.initiated_by || selectedRequest.value?.items?.[0]?.initiated_by || 'FACTORY',
        delivery_order_no: selectedRequest.value.doNo,
        transfer_type: editForm.transfer_type,
        is_billable: !isInternal,
        payment_method: isInternal ? 'INTERNAL' : (item.payment_method === 'INTERNAL' ? 'NET' : item.payment_method || 'NET'),
        payment_status: isInternal ? 'EXEMPT' : (item.payment_status === 'EXEMPT' ? 'UNPAID' : item.payment_status || 'UNPAID'),
      }

      if (item.id) {
        await factoryStore.updateTransfer(item.id, payload)
      } else {
        await factoryStore.createTransfer(payload)
      }
    }
    
    // Clear draft so new changes reflect immediately if SJ modal is opened
    if (selectedRequest.value?.doNo) {
      localStorage.removeItem('sj_draft_' + selectedRequest.value.doNo)
    }
    
    toast.success('Pengiriman berhasil diperbarui.')
    showEditModal.value = false
    await factoryStore.fetchAllTransfers()
  } catch (err) {
    toast.error('Gagal memperbarui pengiriman: ' + getErrorMessage(err))
  } finally {
    processing.value = false
  }
}

const handleDeleteShipment = (group: any) => {
  confirmDialog.title = 'Hapus Pengiriman?'
  confirmDialog.description = `Apakah Anda yakin ingin menghapus pengiriman ${group.doNo}? Tindakan ini tidak dapat dibatalkan dan stok akan dikembalikan.`
  confirmDialog.variant = 'danger'
  confirmDialog.confirmText = 'Ya, Hapus'
  confirmDialog.open = true
  confirmDialog.onConfirm = async () => {
    processing.value = true
    try {
      for (const item of group.items) {
        await factoryStore.deleteTransfer(item.id)
      }
      toast.success('Pengiriman berhasil dihapus.')
      await factoryStore.fetchAllTransfers()
      confirmDialog.open = false
    } catch (err) {
      toast.error('Gagal menghapus pengiriman: ' + getErrorMessage(err))
    } finally {
      processing.value = false
    }
  }
}

const processShipment = (request: any) => {
  selectedRequest.value = request

  // Load draft if exists
  const draftKey = 'sj_draft_' + (request.doNo || request.id || 'new')
  const draftStr = localStorage.getItem(draftKey)
  if (draftStr) {
    try {
      const draft = JSON.parse(draftStr)
      if (draft && draft.factory_id) {
        Object.assign(processForm, draft)
        if (processForm.factory_id) {
          factoryStore.fetchInventory(processForm.factory_id)
        }
        showProcessModal.value = true
        return // Early return, use draft instead of fresh init
      }
    } catch (e) {
      console.error('Error loading draft', e)
    }
  }

  processForm.notes = request.items[0]?.notes || ''
  processForm.vehicle_id = request.items[0]?.vehicle_id || ''
  processForm.driver_id = request.items[0]?.driver_id || ''
  processForm.target_shipment_date = request.items[0]?.target_shipment_date
    ? new Date(request.items[0].target_shipment_date).toISOString().slice(0, 16)
    : ''
  processForm.estimated_arrival = request.items[0]?.estimated_arrival
    ? new Date(request.items[0].estimated_arrival).toISOString().slice(0, 16)
    : ''
  processForm.transfer_type =
    request.items[0]?.transfer_type ||
    (request.items[0]?.payment_method === 'INTERNAL' ? 'INTERNAL_TRANSFER' : 'COMMERCIAL')

  // Clone items for adjustment
  processForm.items = request.items.map((item: any) => ({
    id: item.id,
    product_id: item.product_id,
    batch_no: item.batch_no || '',
    quantity: item.quantity,
    unit: item.unit || 'PCS',
    pcs_per_unit: isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1,
    original_quantity: item.quantity,
    original: true,
  }))

  let targetFactoryId = factoryStore.factories[0]?.id || ''
  if (request.items[0]?.from_factory_id) {
    targetFactoryId = request.items[0].from_factory_id
  }
  processForm.factory_id = targetFactoryId
  processForm.fulfillment_branch_id = request.items[0]?.fulfillment_branch_id || ''

  if (targetFactoryId) {
    factoryStore.fetchInventory(targetFactoryId)
  }
  fetchBranchInventory()
  showProcessModal.value = true
}

const handleConfirmProcess = async () => {
  if (!processForm.factory_id) {
    toast.error('Pilih pabrik pengirim')
    return
  }

  // Verify all items have a product selected
  for (const item of processForm.items) {
    if (!item.product_id) {
      toast.error('Mohon pilih produk untuk semua baris barang.')
      return
    }
  }

  // PRE-VALIDATION: Check if sufficient stock is available for all items combined
  if (processForm.fulfillment_branch_id) {
    const requiredProductQty: Record<string, number> = {}
    for (const item of processForm.items) {
      const product = factoryStore.products.find((p: any) => p.id === item.product_id)
      if (!product) continue
      const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
      const totalPieces = item.quantity * pcsPerUnit
      requiredProductQty[item.product_id] = (requiredProductQty[item.product_id] || 0) + totalPieces
    }

    const branchName =
      masterDataStore.branches.find((b: any) => b.id === processForm.fulfillment_branch_id)?.name ||
      'Cabang Delegasi'

    for (const [productId, needed] of Object.entries(requiredProductQty)) {
      const product = factoryStore.products.find((p: any) => p.id === productId)
      const availableTotal = getTotalBranchStock(productId, processForm.fulfillment_branch_id)
      if (availableTotal < needed) {
        toast.error(
          `Gagal mendelegasikan: Total stok '${product?.name || productId}' di ${branchName} tidak mencukupi. Butuh: ${needed} PCS, Total Tersedia: ${availableTotal} PCS`,
        )
        return // Abort early
      }
    }
  } else {
    // Normal Factory Transfer: Check specific batch stock in Factory
    const requiredPieces: Record<string, { productId: string; batchNo: string; needed: number }> = {}

    for (const item of processForm.items) {
      const product = factoryStore.products.find((p: any) => p.id === item.product_id)
      if (!product) continue

      const pcsPerUnit = isBaseUnit(item.unit) ? 1 : item.pcs_per_unit || 1
      const totalPieces = item.quantity * pcsPerUnit
      const batchNo = (item.batch_no || 'DEFAULT').trim() || 'DEFAULT'

      const key = `${item.product_id}:::${batchNo}`
      if (!requiredPieces[key]) {
        requiredPieces[key] = { productId: item.product_id, batchNo, needed: 0 }
      }
      requiredPieces[key].needed += totalPieces
    }

    for (const { productId, batchNo, needed } of Object.values(requiredPieces)) {
      const product = factoryStore.products.find((p: any) => p.id === productId)
      const availableBatches = getBatchesForProduct(productId, undefined)
      const matchedBatch = availableBatches.find(
        (b: any) => (b.batch_no || 'DEFAULT').trim() === batchNo,
      )
      const availableStock = matchedBatch ? matchedBatch.available_quantity : 0

      if (availableStock < needed) {
        toast.error(
          `Gagal memproses: Stok produk '${
            product?.name || productId
          }' (Batch ${batchNo}) di Pabrik tidak mencukupi. Butuh: ${needed}, Tersedia: ${availableStock}`,
        )
        return // Abort early to prevent partial shipments
      }
    }
  }

  processing.value = true
  try {
    // 1. Update each existing item with potential adjustments and Penugasan Logistik
    // 2. Add new items if any

    const isInternal = processForm.transfer_type === 'INTERNAL_TRANSFER'

    // First, handle updates to original items
    const originalItems = processForm.items.filter((i) => i.original)
    for (const item of originalItems) {
      const currentTransfer = factoryStore.allTransfers.find((t: any) => t.id === item.id)

      // Skip if already processed in a previous attempt (partial success)
      if (
        currentTransfer &&
        (currentTransfer.status === 'SHIPPED' ||
          currentTransfer.status === 'ARRIVED' ||
          currentTransfer.status === 'RECEIVED')
      ) {
        continue
      }

      const originalTransfer = selectedRequest.value.items.find((t: any) => t.id === item.id)

      // Update details first (keep original status)
      await factoryStore.updateTransfer(item.id, {
        quantity: item.quantity,
        unit: item.unit,
        pcs_per_unit: item.pcs_per_unit,
        batch_no: processForm.fulfillment_branch_id ? '' : (item.batch_no || 'DEFAULT').trim() || 'DEFAULT',
        notes: processForm.notes,
        target_shipment_date: formatToRFC3339(processForm.target_shipment_date),
        estimated_arrival: formatToRFC3339(processForm.estimated_arrival),
        vehicle_id: processForm.fulfillment_branch_id ? null : (processForm.vehicle_id || null),
        driver_id: processForm.fulfillment_branch_id ? null : (processForm.driver_id || null),
        fulfillment_branch_id: processForm.fulfillment_branch_id || null,
        from_factory_id: processForm.factory_id,
        transfer_type: processForm.transfer_type,
        is_billable: !isInternal,
        payment_method: isInternal ? 'INTERNAL' : (item.payment_method === 'INTERNAL' ? 'NET' : item.payment_method || 'NET'),
        payment_status: isInternal ? 'EXEMPT' : (item.payment_status === 'EXEMPT' ? 'UNPAID' : item.payment_status || 'UNPAID'),
      })

      // Approve if it was REQUESTED
      if (originalTransfer?.status !== 'APPROVED' && currentTransfer?.status !== 'APPROVED') {
        await factoryStore.approveTransfer(item.id)
      }

      // Ship it (Deducts stock in backend)
      if (!processForm.fulfillment_branch_id) {
        await factoryStore.shipTransfer(item.id)
      }
    }

    // Handle new items added by factory
    const newItems = processForm.items.filter((i) => !i.original)
    if (newItems.length > 0) {
      for (const item of newItems) {
        const payload = {
          from_factory_id: processForm.factory_id,
          to_branch_id: selectedRequest.value.items[0].to_branch_id,
          product_id: item.product_id,
          quantity: item.quantity,
          unit: item.unit,
          pcs_per_unit: item.pcs_per_unit,
          batch_no: processForm.fulfillment_branch_id ? '' : (item.batch_no || 'DEFAULT').trim() || 'DEFAULT',
          notes: processForm.notes,
          target_shipment_date: formatToRFC3339(processForm.target_shipment_date),
          estimated_arrival: formatToRFC3339(processForm.estimated_arrival),
          vehicle_id: processForm.fulfillment_branch_id ? null : (processForm.vehicle_id || null),
          driver_id: processForm.fulfillment_branch_id ? null : (processForm.driver_id || null),
          fulfillment_branch_id: processForm.fulfillment_branch_id || null,
          delivery_order_no: selectedRequest.value.doNo,
          initiated_by: 'FACTORY',
          status: 'APPROVED', // Create as APPROVED so it can be shipped
          transfer_type: processForm.transfer_type,
          is_billable: !isInternal,
          payment_method: isInternal ? 'INTERNAL' : 'NET',
          payment_status: isInternal ? 'EXEMPT' : 'UNPAID',
        }
        const created = await factoryStore.createTransfer(payload)
        item.id = created.id
        item.original = true // Mark as original so we don't recreate it on retry
        if (!processForm.fulfillment_branch_id) {
          await factoryStore.shipTransfer(created.id)
        }
      }
    }

    // Fetch latest inventory to reflect deducted stocks
    await factoryStore.fetchInventory(processForm.factory_id)

    const branchName =
      masterDataStore.branches.find((b: any) => b.id === processForm.fulfillment_branch_id)?.name ||
      'Cabang Delegasi'

    toast.success(
      processForm.fulfillment_branch_id
        ? `Tugas pengiriman berhasil didelegasikan ke Cabang ${branchName} (Status: APPROVED).`
        : 'Surat Jalan diterbitkan. Pengiriman kini dalam perjalanan (SHIPPED).'
    )
    showProcessModal.value = false
    
    // Clear draft
    if (selectedRequest.value) {
      const draftKey = 'sj_draft_' + (selectedRequest.value.doNo || selectedRequest.value.id || 'new')
      localStorage.removeItem(draftKey)
    }

    await factoryStore.fetchAllTransfers()
  } catch (err) {
    toast.error('Gagal memproses pengiriman: ' + getErrorMessage(err))
  } finally {
    processing.value = false
  }
}

const handleApproveRequest = (request: any) => {
  confirmDialog.title = 'Setujui Permintaan?'
  confirmDialog.description =
    'Apakah Anda yakin ingin menyetujui permintaan barang dari cabang ini? Pastikan stok mencukupi.'
  confirmDialog.variant = 'success'
  confirmDialog.confirmText = 'Ya, Setujui'
  confirmDialog.open = true
  confirmDialog.onConfirm = async () => {
    processing.value = true
    try {
      for (const item of request.items) {
        await factoryStore.approveTransfer(item.id)
      }
      toast.success('Permintaan telah disetujui.')
      await factoryStore.fetchAllTransfers()
      confirmDialog.open = false
    } catch (err) {
      toast.error('Gagal menyetujui permintaan: ' + getErrorMessage(err))
    } finally {
      processing.value = false
    }
  }
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'RECEIVED':
      return 'bg-emerald-50 text-emerald-600 border-emerald-100'
    case 'ARRIVED':
      return 'bg-cyan-50 text-cyan-600 border-cyan-100'
    case 'SHIPPED':
      return 'bg-blue-50 text-blue-600 border-blue-100'
    case 'REQUESTED':
      return 'bg-orange-50 text-orange-600 border-orange-100'
    case 'APPROVED':
      return 'bg-emerald-50 text-emerald-700 border-emerald-200 shadow-sm'
    case 'REJECTED':
      return 'bg-rose-50 text-rose-600 border-rose-100 shadow-sm'
    default:
      return 'bg-slate-50 text-slate-500 border-slate-100'
  }
}

watch(
  () => processForm,
  (newVal) => {
    if (showProcessModal.value && selectedRequest.value) {
      const draftKey = 'sj_draft_' + (selectedRequest.value.doNo || selectedRequest.value.id || 'new')
      localStorage.setItem(draftKey, JSON.stringify(newVal))
    }
  },
  { deep: true }
)
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

@media print {
  .no-print {
    display: none !important;
  }
  .print-only {
    display: block !important;
  }
  body {
    background: white !important;
    padding: 0 !important;
    margin: 0 !important;
  }
}
</style>
