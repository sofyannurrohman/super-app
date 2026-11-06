<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-2xl font-extrabold tracking-tight text-slate-900">Approval & Pengiriman</h2>
        <p class="text-sm font-medium text-slate-500 mt-1">
          Supervisor: Assign armada & print Surat Jalan.
        </p>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" @click="fetchData">
          <RefreshCw class="h-4 w-4 mr-2" :class="{ 'animate-spin': loading }" />
          Refresh
        </Button>
      </div>
    </div>

    <!-- Workflow Tabs -->
    <div class="flex gap-4 border-b border-slate-100 pb-px mb-4 overflow-x-auto no-scrollbar">
      <button
        @click="activeTab = 'batching'"
        class="pb-4 px-2 text-sm font-black transition-all relative whitespace-nowrap"
        :class="activeTab === 'batching' ? 'text-primary' : 'text-slate-400 hover:text-slate-600'"
      >
        1. BUAT BATCH & PENUGASAN (SUPERVISOR)
        <div
          v-if="activeTab === 'batching'"
          class="absolute bottom-0 left-0 right-0 h-1 bg-primary rounded-t-full"
        ></div>
      </button>
      <button
        @click="activeTab = 'history'"
        class="pb-4 px-2 text-sm font-black transition-all relative whitespace-nowrap"
        :class="activeTab === 'history' ? 'text-primary' : 'text-slate-400 hover:text-slate-600'"
      >
        RIWAYAT & LOG PENGIRIMAN
        <div
          v-if="activeTab === 'history'"
          class="absolute bottom-0 left-0 right-0 h-1 bg-primary rounded-t-full"
        ></div>
      </button>
    </div>

    <!-- Content based on Tabs -->
    <div v-if="activeTab === 'batching'" class="space-y-6">
      <div class="bg-white rounded-3xl border border-slate-200 p-6 space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <h3 class="text-[11px] font-black text-slate-400 uppercase tracking-widest">
              1. Pilih Nota (Verified)
            </h3>
            <div
              class="border-2 border-slate-100 rounded-2xl divide-y max-h-[300px] overflow-y-auto bg-slate-50/50"
            >
              <div
                v-for="order in deliveryStore.pendingOrders.filter((o: any) => o.status === 'CONFIRMED' && !o.delivery_batch_id)"
                :key="order.id"
                class="p-4 flex items-center gap-4 hover:bg-white transition-all cursor-pointer"
                @click="toggleOrder(order.id)"
              >
                <input
                  type="checkbox"
                  :checked="selectedOrders.includes(order.id)"
                  class="w-5 h-5 rounded-lg border-2 border-slate-300 text-primary focus:ring-0 transition-all"
                />
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-black text-slate-900 truncate">
                    {{ order.store?.name || 'Toko Manual' }}
                  </p>
                  <p class="text-[10px] text-slate-500 font-bold uppercase">
                    {{ order.receipt_no || 'No Receipt' }} •
                    {{ formatCurrency(order.total_amount) }}
                  </p>
                </div>
              </div>
              <div
                v-if="deliveryStore.pendingOrders.length === 0"
                class="p-10 text-center text-slate-400"
              >
                <PackageOpen class="h-10 w-10 mx-auto mb-3 opacity-20" />
                <p class="text-xs font-bold uppercase tracking-widest">Tidak ada nota divalidasi</p>
              </div>
            </div>

            <!-- Bagian Penagihan Piutang (Opsional) -->
            <h3 class="text-[11px] font-black text-slate-400 uppercase tracking-widest mt-6">
              1b. Pilih Nota Penagihan (Opsional)
            </h3>
            <div
              class="border-2 border-slate-100 rounded-2xl divide-y max-h-[300px] overflow-y-auto bg-slate-50/50"
            >
              <div
                v-for="order in deliveryStore.unpaidOrders"
                :key="order.id"
                class="p-4 flex items-center gap-4 hover:bg-white transition-all cursor-pointer"
                @click="toggleUnpaidOrder(order.id)"
              >
                <input
                  type="checkbox"
                  :checked="selectedUnpaidOrders.includes(order.id)"
                  class="w-5 h-5 rounded-lg border-2 border-slate-300 text-orange-500 focus:ring-0 transition-all"
                />
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-black text-slate-900 truncate flex gap-2 items-center">
                    {{ order.store?.name || 'Toko Manual' }}
                    <span class="bg-orange-100 text-orange-600 text-[9px] px-1.5 py-0.5 rounded font-black">TEMPO</span>
                  </p>
                  <p class="text-[10px] text-slate-500 font-bold uppercase">
                    {{ order.so_number || order.receipt_no }} • 
                    Sisa: {{ formatCurrency(order.total_amount - order.total_paid) }}
                  </p>
                </div>
              </div>
              <div
                v-if="deliveryStore.unpaidOrders?.length === 0"
                class="p-10 text-center text-slate-400"
              >
                <Banknote class="h-10 w-10 mx-auto mb-3 opacity-20" />
                <p class="text-xs font-bold uppercase tracking-widest">Tidak ada piutang tersedia</p>
              </div>
            </div>

            <!-- Bagian Sampel (Opsional) -->
            <h3 class="text-[11px] font-black text-slate-400 uppercase tracking-widest mt-6">
              1c. Pilih Permintaan Sampel (Opsional)
            </h3>
            <div
              class="border-2 border-slate-100 rounded-2xl divide-y max-h-[300px] overflow-y-auto bg-slate-50/50"
            >
              <div
                v-for="sample in deliveryStore.approvedSamples"
                :key="sample.id"
                class="p-4 flex items-center gap-4 hover:bg-white transition-all cursor-pointer"
                @click="toggleSample(sample.id)"
              >
                <input
                  type="checkbox"
                  :checked="selectedSamples.includes(sample.id)"
                  class="w-5 h-5 rounded-lg border-2 border-slate-300 text-indigo-500 focus:ring-0 transition-all"
                />
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-black text-slate-900 truncate flex gap-2 items-center">
                    {{ sample.product?.name }} (Qty: {{ sample.quantity }})
                    <span class="bg-indigo-100 text-indigo-600 text-[9px] px-1.5 py-0.5 rounded font-black">SAMPEL</span>
                  </p>
                  <p class="text-[10px] text-slate-500 font-bold uppercase">
                    Penerima: {{ sample.store?.name || sample.employee?.first_name || 'Toko Manual' }}
                  </p>
                </div>
              </div>
              <div
                v-if="deliveryStore.approvedSamples?.length === 0"
                class="p-10 text-center text-slate-400"
              >
                <Gift class="h-10 w-10 mx-auto mb-3 opacity-20" />
                <p class="text-xs font-bold uppercase tracking-widest">Tidak ada permintaan sampel</p>
              </div>
            </div>
          </div>

          <div class="space-y-6">
            <h3 class="text-[11px] font-black text-slate-400 uppercase tracking-widest">
              2. Penugasan Logistik
            </h3>
            <div class="space-y-4">
              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                  >Pilih Driver Utama</label
                >
                <select
                  v-model="form.driver_id"
                  class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none cursor-pointer"
                >
                  <option value="" disabled>Pilih Supir Pengiriman</option>
                  <option
                    v-for="d in sortedDrivers"
                    :key="d.id"
                    :value="d.id"
                    :disabled="isDriverBusy(d.id)"
                  >
                    {{ d.first_name }} {{ d.last_name }}
                    {{ isDriverBusy(d.id) ? '⚠️ (Sedang Bertugas / Sibuk)' : '🟢 (Tersedia)' }}
                  </option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                  >Pilih Helper / Kernet (Opsional)</label
                >
                <select
                  v-model="form.helper_id"
                  class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none cursor-pointer"
                >
                  <option value="">-- Tanpa Helper --</option>
                  <option
                    v-for="h in sortedHelpers"
                    :key="h.id"
                    :value="h.id"
                    :disabled="isHelperBusy(h.id)"
                  >
                    {{ h.first_name }} {{ h.last_name }}
                    {{ isHelperBusy(h.id) ? '⚠️ (Sedang Bertugas / Sibuk)' : '🟢 (Tersedia)' }}
                  </option>
                </select>
              </div>
              <div v-if="calculateTotalSelectedWeight > 0" class="space-y-2 mt-4 mb-4 p-4 bg-indigo-50 border border-indigo-100 rounded-2xl">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-black text-indigo-900 uppercase">Total Tonase (Estimasi)</span>
                  <span class="text-sm font-black text-indigo-700">{{ calculateTotalSelectedWeight.toLocaleString('id-ID', { maximumFractionDigits: 2 }) }} KG</span>
                </div>
                <p class="text-[10px] font-bold text-indigo-600/70 leading-tight mt-1">
                  Gunakan informasi ini untuk menentukan kapasitas armada yang sesuai.
                </p>
              </div>

              <div class="space-y-2">
                <label class="text-[10px] font-black text-slate-400 uppercase tracking-widest ml-1"
                  >Pilih Armada (Kendaraan)</label
                >
                <select
                  v-model="form.vehicle_id"
                  class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none cursor-pointer"
                >
                  <option value="" disabled>Pilih Kendaraan</option>
                  <option
                    v-for="v in sortedVehicles"
                    :key="v.id"
                    :value="v.id"
                    :disabled="isVehicleBusy(v.id)"
                  >
                    {{ v.license_plate }} - {{ v.model }} ({{ v.type }})
                    {{ isVehicleBusy(v.id) ? '⚠️ (Sedang Beroperasi / Sibuk)' : '🟢 (Tersedia)' }}
                  </option>
                </select>
              </div>
            </div>

            <Button
              class="w-full bg-slate-900 hover:bg-slate-800 text-white rounded-2xl py-8 font-black shadow-xl shadow-slate-200 transition-all flex items-center justify-center gap-3 mt-4"
              :disabled="!isFormValid || isSubmitting"
              @click="handleCreateBatch"
            >
              <Truck class="w-6 h-6" />
              {{
                isSubmitting
                  ? 'MEMPROSES...'
                  : isEditingBatch
                    ? 'PERBARUI BATCH PENGIRIMAN'
                    : 'BUAT BATCH PENGIRIMAN'
              }}
            </Button>
            <Button
              v-if="isEditingBatch"
              variant="outline"
              class="w-full border-2 border-slate-200 text-slate-500 rounded-2xl py-8 font-black mt-2"
              @click="cancelEdit"
            >
              BATAL EDIT
            </Button>
          </div>
        </div>
      </div>

      <!-- Batch Waiting Supervisor Approval -->
      <div v-if="waitingSupervisorApproval.length > 0" class="space-y-4">
        <h3 class="text-[11px] font-black text-slate-400 uppercase tracking-widest px-1">
          Batch Menunggu Finalisasi Supervisor
        </h3>
        <div
          v-for="batch in waitingSupervisorApproval"
          :key="batch.id"
          class="bg-white rounded-3xl border border-slate-200 p-6 flex items-center justify-between shadow-sm group"
        >
          <div class="flex items-center gap-6">
            <div class="bg-indigo-50 p-4 rounded-2xl">
              <Truck class="w-8 h-8 text-indigo-600" />
            </div>
            <div>
              <h4 class="text-lg font-black text-slate-900">
                {{ batch.items?.length || 0 }} Nota • {{ batch.driver?.first_name }}
              </h4>
              <p class="text-xs font-bold text-slate-500 mt-0.5">
                Dibuat pada {{ formatDate(batch.created_at) }}
              </p>
            </div>
          </div>
          <div class="flex gap-3">
            <Button
              variant="outline"
              class="border-primary/20 text-primary hover:bg-primary/5 rounded-2xl"
              @click="handleEditBatch(batch)"
            >
              <Pencil class="w-4 h-4" />
            </Button>
            <Button
              @click="handleSupervisorApprove(batch.id)"
              class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-2xl px-6 py-4 text-xs font-black shadow-lg shadow-indigo-100 flex items-center gap-2"
            >
              <CheckCircle2 class="w-4 h-4" /> FINALIZE BATCH
            </Button>
            <Button
              variant="ghost"
              class="text-red-500 hover:bg-red-50 rounded-2xl"
              @click="handleDeleteBatch(batch.id)"
            >
              <Trash2 class="w-4 h-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="activeTab === 'history'" class="space-y-4">
      <Card class="rounded-[32px] overflow-hidden border-slate-200">
        <DataTable :columns="batchColumns" :data="deliveryStore.batches">
          <template #cell-delivery_order_no="{ row }">
            <div class="flex items-center gap-3">
              <span
                class="font-mono text-[10px] font-black bg-slate-100 text-slate-600 px-2 py-1 rounded"
                >{{ (row as any).delivery_order_no || 'NO DO' }}</span
              >
              <Button
                size="icon"
                variant="ghost"
                class="h-9 w-9 rounded-xl hover:bg-primary/5 text-primary"
                @click="printDO(row as any)"
              >
                <Printer class="h-4.5 w-4.5" />
              </Button>
            </div>
          </template>
          <template #cell-status="{ row }">
            <Badge
              :variant="(row as any).status === 'COMPLETED' ? 'success' : 'warning'"
              class="rounded-lg font-black text-[10px] uppercase tracking-wider px-3 py-1"
            >
              {{ (row as any).status }}
            </Badge>
          </template>
        </DataTable>
      </Card>
    </div>

    <!-- DO Printing Modal (SJ Preview - Landscape Format) -->
    <div
      v-if="selectedDO"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-50 flex items-center justify-center p-0 sm:p-4 overflow-y-auto"
    >
      <Card
        id="surat-jalan-card"
        class="w-full max-w-[95%] sm:max-w-6xl bg-white shadow-2xl border-0 overflow-hidden my-auto print:shadow-none print:max-w-full"
      >
        <!-- Simplified Header for Manual Writing -->
        <div class="bg-white p-4 print:p-2">
          <div class="grid grid-cols-3 gap-8 items-center">
            <div class="space-y-1">
              <p class="text-[9px] font-black text-slate-900 uppercase tracking-widest">
                SOPIR / ARMADA
              </p>
              <p class="text-[11px] font-black text-slate-900">
                {{ selectedDO.driver?.first_name || '........................' }} /
                {{ selectedDO.vehicle?.license_plate || '..........' }}
              </p>
            </div>
            <div class="space-y-1">
              <p class="text-[9px] font-black text-slate-900 uppercase tracking-widest">
                SALESMAN / HELPER
              </p>
              <p class="text-[11px] font-black text-slate-900">
                {{
                  selectedDO.items?.[0]?.sales_order?.employee?.first_name ||
                  selectedDO.items?.[0]?.sales_transaction?.employee?.first_name ||
                  '........................'
                }}
              </p>
            </div>
            <div class="text-right">
              <p
                class="text-[10px] font-mono font-bold text-slate-900 pb-1 italic uppercase tracking-tighter"
              >
                SURAT JALAN #{{ selectedDO.delivery_order_no }}
              </p>
              <p class="text-[9px] font-bold text-slate-900 mt-1 uppercase">
                {{
                  new Date(selectedDO.created_at).toLocaleDateString('id-ID', { dateStyle: 'long' })
                }}
              </p>
            </div>
          </div>
        </div>

        <div class="p-4 sm:p-6 space-y-4 print:p-2">
          <!-- Main Table with Placeholders -->
          <div class="border-2 border-slate-900 rounded-lg overflow-x-auto bg-white">
            <table class="w-full text-left text-[10px] border-collapse whitespace-nowrap">
              <thead>
                <tr class="bg-slate-50 border-b-2 border-slate-900">
                  <th class="p-2 border-r border-slate-300 font-black text-center w-8">NO</th>
                  <th class="p-1 border-r border-slate-300 font-black text-center w-12 text-[8px]">
                    QR SCAN
                  </th>
                  <th class="p-2 border-r border-slate-300 font-black">
                    NAMA TOKO / TELP / ALAMAT
                  </th>
                  <th class="p-2 border-r border-slate-300 font-black text-right w-20 bg-amber-50">
                    TAGIHAN
                  </th>

                  <th
                    class="p-1 border-r border-slate-300 font-black text-center w-8 bg-slate-100/50"
                  >
                    GLN
                  </th>
                  <th
                    class="p-1 border-r border-slate-300 font-black text-center w-8 bg-slate-100/50"
                  >
                    BTL
                  </th>
                  <th
                    class="p-1 border-r border-slate-300 font-black text-center w-8 bg-slate-100/50"
                  >
                    KRT
                  </th>
                  <th
                    class="p-2 border-r border-slate-300 font-black text-center w-24 bg-emerald-50"
                  >
                    BAYAR
                  </th>
                  <th class="p-2 border-r border-slate-300 font-black text-center w-16">KET</th>

                  <th
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 font-black text-center text-[7px] bg-slate-100/30 w-10 uppercase"
                  >
                    {{ prod.name }}
                  </th>

                  <th
                    class="p-2 border-r border-slate-300 font-black text-center w-10 bg-pink-50 text-[8px]"
                  >
                    HADIAH
                  </th>
                  <th class="p-2 font-black text-center w-10 text-[8px]">RETUR</th>
                </tr>
              </thead>
              <tbody class="divide-y border-slate-300">
                <!-- Data Rows -->
                <tr v-for="(item, idx) in selectedDO.items" :key="item.id" class="h-8">
                  <td class="p-2 border-r border-slate-300 text-center font-bold">
                    {{ (idx as number) + 1 }}
                  </td>
                  <td class="p-1 border-r border-slate-300 text-center align-middle">
                    <img
                      :src="`https://api.qrserver.com/v1/create-qr-code/?size=60x60&margin=0&data=${(item.sales_order || item.sales_transaction)?.so_number || (item.sales_order || item.sales_transaction)?.invoice_no || item.id}`"
                      class="w-8 h-8 mx-auto"
                    />
                  </td>
                  <td class="p-2 border-r border-slate-300 whitespace-normal">
                    <p class="font-black text-slate-900 leading-none">
                      {{ item.sales_transaction?.store?.name }}
                      <span class="text-[8px] text-slate-900 font-normal"
                        >/ {{ item.sales_transaction?.store?.phone || '-' }}</span
                      >
                    </p>
                    <p class="text-[7px] text-slate-900 font-bold mt-0.5">
                      {{ item.sales_transaction?.store?.address }}
                    </p>
                  </td>
                  <td class="p-2 border-r border-slate-300 text-right font-black bg-amber-50/20">
                    {{
                      formatCurrency(item.sales_transaction?.total_amount).replace('Rp', '').trim()
                    }}
                  </td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300 bg-emerald-50/10"></td>
                  <td class="p-2 border-r border-slate-300 text-center font-bold text-[8px]">
                    {{ item.sales_transaction?.payment_method || 'COD' }}
                  </td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 text-center font-black"
                  >
                    {{ getProductQty(item.sales_transaction, prod.id) || '' }}
                  </td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2"></td>
                </tr>

                <!-- Placeholder Rows for Manual Entry (Fills up to 20 rows) -->
                <tr
                  v-for="n in Math.max(0, 20 - (selectedDO.items?.length || 0))"
                  :key="'blank-' + n"
                  class="h-6 bg-slate-50/10"
                >
                  <td class="p-2 border-r border-slate-300 text-center text-slate-900">
                    {{ (selectedDO.items?.length || 0) + n }}
                  </td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300 bg-amber-50/10"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2 border-r border-slate-300 bg-emerald-50/5"></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300"
                  ></td>
                  <td class="p-2 border-r border-slate-300"></td>
                  <td class="p-2"></td>
                </tr>

                <!-- Total Row -->
                <tr class="bg-slate-50 font-black border-t-2 border-slate-900 h-8">
                  <td
                    colspan="3"
                    class="p-2 text-right border-r border-slate-300 uppercase tracking-widest text-[8px]"
                  >
                    TOTAL BATCH
                  </td>
                  <td class="p-2 text-right border-r border-slate-300 bg-amber-50">
                    {{ formatCurrency(calculateBatchTotal).replace('Rp', '').trim() }}
                  </td>
                  <td colspan="4" class="border-r border-slate-300"></td>
                  <td class="border-r border-slate-300"></td>
                  <td
                    v-for="prod in batchProductColumns"
                    :key="prod.id"
                    class="p-1 border-r border-slate-300 text-center text-primary bg-slate-100/30"
                  >
                    {{ calculateProductTotal(prod.id) }}
                  </td>
                  <td colspan="2"></td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Compact Signature Section -->
          <div class="grid grid-cols-3 gap-8 mt-6">
            <div class="text-center space-y-8">
              <p class="text-[9px] font-black uppercase tracking-widest">SOPIR / HELPER</p>
              <p class="text-[8px] font-bold text-slate-900">
                ( {{ selectedDO.driver?.first_name || '..........' }} /
                {{ selectedDO.helper?.first_name || '..........' }} )
              </p>
            </div>
            <div class="text-center space-y-8">
              <p class="text-[9px] font-black uppercase tracking-widest">ADMIN NOTA</p>
              <p class="text-[8px] font-bold text-slate-900">
                ( {{ selectedDO.admin_nota?.first_name || '........................' }} )
              </p>
            </div>
            <div class="flex justify-end items-center pr-10">
              <div class="text-right mr-4">
                <p class="text-[8px] font-black text-slate-400 uppercase leading-none">
                  SCAN VALIDASI
                </p>
                <p class="text-[7px] text-slate-300 italic mt-1 uppercase">Surat Jalan Digital</p>
              </div>
              <img
                :src="`https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${selectedDO.delivery_order_no}`"
                class="w-20 h-20 border-2 border-slate-900 p-1 rounded bg-white shadow-sm"
              />
            </div>
          </div>
        </div>

        <!-- BUKTI PENYERAHAN SAMPEL (BPS) - Attached to SJ Print -->
        <div v-if="selectedDO.samples && selectedDO.samples.length > 0" class="p-4 sm:p-6 space-y-4 print:p-2" style="page-break-before: always;">
          <div class="text-center mb-6">
            <h2 class="text-lg font-black uppercase tracking-widest text-slate-900 border-b-2 border-slate-900 pb-2 inline-block">
              BUKTI PENYERAHAN SAMPEL (BPS)
            </h2>
            <p class="text-xs font-bold mt-2">SURAT JALAN #{{ selectedDO.delivery_order_no }}</p>
          </div>

          <div class="border-2 border-slate-900 rounded-lg overflow-x-auto bg-white">
            <table class="w-full text-left text-xs border-collapse whitespace-nowrap">
              <thead>
                <tr class="bg-slate-50 border-b-2 border-slate-900">
                  <th class="p-2 border-r border-slate-300 font-black text-center w-8">NO</th>
                  <th class="p-2 border-r border-slate-300 font-black">NAMA PRODUK</th>
                  <th class="p-2 border-r border-slate-300 font-black text-center w-20">QTY</th>
                  <th class="p-2 border-r border-slate-300 font-black text-center">PENERIMA</th>
                  <th class="p-2 font-black">KETERANGAN</th>
                </tr>
              </thead>
              <tbody class="divide-y border-slate-300">
                <tr v-for="(sample, idx) in selectedDO.samples" :key="sample.id">
                  <td class="p-2 border-r border-slate-300 text-center font-bold">{{ (idx as number) + 1 }}</td>
                  <td class="p-2 border-r border-slate-300 font-bold">{{ sample.product?.name }}</td>
                  <td class="p-2 border-r border-slate-300 text-center font-bold text-lg">{{ sample.quantity }}</td>
                  <td class="p-2 border-r border-slate-300 text-center uppercase text-[10px] font-black whitespace-normal">
                    {{ sample.recipient_type === 'STORE' ? sample.store?.name : (sample.salesman?.first_name + ' ' + (sample.salesman?.last_name || '')) }}
                  </td>
                  <td class="p-2 text-[10px] italic whitespace-normal">{{ sample.notes || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Signature Block BPS -->
          <div class="grid grid-cols-2 gap-8 mt-12 mb-8 px-10">
            <div class="text-center space-y-16">
              <p class="text-[10px] font-black uppercase tracking-widest">DISERAHKAN OLEH (SOPIR)</p>
              <div class="border-b-2 border-slate-900 mx-auto w-48"></div>
              <p class="text-[10px] font-bold text-slate-900 -mt-14 uppercase">
                {{ selectedDO.driver?.first_name || 'NAMA TERANG' }}
              </p>
            </div>
            <div class="text-center space-y-16">
              <p class="text-[10px] font-black uppercase tracking-widest">DITERIMA OLEH (TOKO/SALES)</p>
              <div class="border-b-2 border-slate-900 mx-auto w-48"></div>
              <p class="text-[10px] font-bold text-slate-900 -mt-14">
                NAMA TERANG & CAP TOKO
              </p>
            </div>
          </div>
        </div>

        <CardFooter class="p-6 bg-slate-50 border-t flex justify-end gap-3 print:hidden">
          <Button variant="ghost" class="font-bold text-slate-500" @click="selectedDO = null"
            >TUTUP PREVIEW</Button
          >
          <Button
            class="bg-slate-900 hover:bg-slate-800 text-white font-black px-12 py-6 rounded-2xl shadow-xl shadow-slate-200"
            @click="handlePrint"
          >
            <Printer class="w-5 h-5 mr-3" />
            CETAK SURAT JALAN (LANDSCAPE)
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed, watch } from 'vue'
import { useDeliveryStore } from '@/stores/delivery'
import { useMasterDataStore } from '@/stores/masterData'
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
  Truck,
  RefreshCw,
  Printer,
  PackagePlus,
  PackageOpen,
  MapPin,
  Search,
  ClipboardList,
  CheckCircle2,
  Trash2,
  Pencil,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'

const loading = ref(false)
const selectedOrders = ref<string[]>([])
const selectedUnpaidOrders = ref<string[]>([])
const selectedSamples = ref<string[]>([])
const isSubmitting = ref(false)
const selectedDO = ref<any>(null)
const isEditingBatch = ref(false)
const editingBatchId = ref<string | null>(null)

const drivers = ref<any[]>([])
const helpers = ref<any[]>([])
const vehicles = ref<any[]>([])

const API_URL = import.meta.env.VITE_API_BASE_URL
  ? import.meta.env.VITE_API_BASE_URL + '/v1'
  : '/api/v1'
const deliveryStore = useDeliveryStore()
const masterDataStore = useMasterDataStore()

const activeTab = ref('batching')
const form = reactive({
  driver_id: '',
  helper_id: '',
  vehicle_id: '',
})

// Computed for Product Matrix in SJ
const batchProductColumns = computed(() => {
  if (!selectedDO.value || !selectedDO.value.items) return []

  const productMap = new Map()
  selectedDO.value.items.forEach((item: any) => {
    if (item.sales_transaction?.items) {
      item.sales_transaction.items.forEach((si: any) => {
        if (si.product) {
          productMap.set(si.product_id, {
            id: si.product_id,
            name: si.product.name
              .split(' ')
              .map((w: string) => w[0])
              .join(''), // Abbreviate
          })
        }
      })
    }
  })
  return Array.from(productMap.values())
})

const getProductQty = (trx: any, productId: string) => {
  if (!trx || !trx.items) return 0
  const item = trx.items.find((i: any) => i.product_id === productId)
  return item ? item.quantity : 0
}

const calculateBatchTotal = computed(() => {
  if (!selectedDO.value || !selectedDO.value.items) return 0
  return selectedDO.value.items.reduce((sum: number, item: any) => {
    return sum + (item.sales_transaction?.total_amount || 0)
  }, 0)
})

const calculateProductTotal = (productId: string) => {
  if (!selectedDO.value || !selectedDO.value.items) return 0
  return selectedDO.value.items.reduce((sum: number, item: any) => {
    return sum + getProductQty(item.sales_transaction, productId)
  }, 0)
}

const isFormValid = computed(() => {
  return (selectedOrders.value.length > 0 || selectedUnpaidOrders.value.length > 0 || selectedSamples.value.length > 0) && form.driver_id && form.vehicle_id
})

const calculateTotalSelectedWeight = computed(() => {
  let totalWeightKg = 0

  const addWeight = (items: any[]) => {
    if (!items) return
    items.forEach((item: any) => {
      if (item.product && item.product.weight) {
        // Use item.quantity (base unit quantity)
        const qty = item.quantity || item.ordered_quantity || 0
        let weight = item.product.weight
        let unit = (item.product.weight_unit || 'KG').toUpperCase()
        
        // Normalize to KG
        if (unit === 'GR' || unit === 'G') weight = weight / 1000
        else if (unit === 'TON') weight = weight * 1000
        
        totalWeightKg += (weight * qty)
      }
    })
  }

  // Pending Orders
  selectedOrders.value.forEach((id) => {
    const order = deliveryStore.pendingOrders.find((o: any) => o.id === id)
    if (order) addWeight(order.items)
  })

  // Unpaid Orders
  selectedUnpaidOrders.value.forEach((id) => {
    const order = deliveryStore.unpaidOrders.find((o: any) => o.id === id)
    if (order) addWeight(order.items)
  })

  // Samples (Direct product & quantity)
  selectedSamples.value.forEach((id) => {
    const sample = deliveryStore.approvedSamples.find((s: any) => s.id === id)
    if (sample && sample.product && sample.product.weight) {
      const qty = sample.quantity || 0
      let weight = sample.product.weight
      let unit = (sample.product.weight_unit || 'KG').toUpperCase()
      
      if (unit === 'GR' || unit === 'G') weight = weight / 1000
      else if (unit === 'TON') weight = weight * 1000
      
      totalWeightKg += (weight * qty)
    }
  })

  return totalWeightKg
})

const toggleOrder = (id: string) => {
  if (selectedOrders.value.includes(id)) {
    selectedOrders.value = selectedOrders.value.filter((o) => o !== id)
  } else {
    selectedOrders.value.push(id)
  }
}

const toggleUnpaidOrder = (id: string) => {
  if (selectedUnpaidOrders.value.includes(id)) {
    selectedUnpaidOrders.value = selectedUnpaidOrders.value.filter((o) => o !== id)
  } else {
    selectedUnpaidOrders.value.push(id)
  }
}

const toggleSample = (id: string) => {
  if (selectedSamples.value.includes(id)) {
    selectedSamples.value = selectedSamples.value.filter((s) => s !== id)
  } else {
    selectedSamples.value.push(id)
  }
}

const isDriverBusy = (driverId: string) => {
  if (!deliveryStore.batches) return false
  return deliveryStore.batches.some((b: any) => {
    if (isEditingBatch.value && editingBatchId.value === b.id) return false
    return (
      b.driver_id === driverId &&
      b.status !== 'COMPLETED' &&
      b.status !== 'CANCELLED' &&
      b.status !== 'REJECTED'
    )
  })
}

const isHelperBusy = (helperId: string) => {
  if (!deliveryStore.batches) return false
  return deliveryStore.batches.some((b: any) => {
    if (isEditingBatch.value && editingBatchId.value === b.id) return false
    return (
      b.helper_id === helperId &&
      b.status !== 'COMPLETED' &&
      b.status !== 'CANCELLED' &&
      b.status !== 'REJECTED'
    )
  })
}

const isVehicleBusy = (vehicleId: string) => {
  if (!deliveryStore.batches) return false
  return deliveryStore.batches.some((b: any) => {
    // If editing, exclude the current editing batch
    if (isEditingBatch.value && editingBatchId.value === b.id) return false

    return (
      b.vehicle_id === vehicleId &&
      b.status !== 'COMPLETED' &&
      b.status !== 'CANCELLED' &&
      b.status !== 'REJECTED'
    )
  })
}

const sortedDrivers = computed(() => {
  return [...drivers.value].sort((a, b) => {
    const aBusy = isDriverBusy(a.id) ? 1 : 0
    const bBusy = isDriverBusy(b.id) ? 1 : 0
    return aBusy - bBusy
  })
})

const sortedHelpers = computed(() => {
  return [...helpers.value].sort((a, b) => {
    const aBusy = isHelperBusy(a.id) ? 1 : 0
    const bBusy = isHelperBusy(b.id) ? 1 : 0
    return aBusy - bBusy
  })
})

const sortedVehicles = computed(() => {
  return [...vehicles.value].sort((a, b) => {
    const aBusy = isVehicleBusy(a.id) ? 1 : 0
    const bBusy = isVehicleBusy(b.id) ? 1 : 0
    return aBusy - bBusy // Available first
  })
})

const waitingSupervisorApproval = computed(() => {
  return deliveryStore.batches.filter((b: any) => b.status === 'WAITING_APPROVAL')
})


const fetchData = async () => {
  if (!masterDataStore.selectedBranchId) {
    console.log('Waiting for branch selection...')
    return
  }
  loading.value = true
  try {
    await deliveryStore.fetchPendingOrders(masterDataStore.selectedBranchCompanyId || '')
    await deliveryStore.fetchUnpaidOrders(masterDataStore.selectedBranchCompanyId || '')
    await deliveryStore.fetchApprovedSamples(masterDataStore.selectedBranchCompanyId || '')
    await deliveryStore.fetchBatches()

    const branchId = masterDataStore.selectedBranchId
    const companyId = masterDataStore.selectedBranchCompanyId || ''

    // Fetch employees for selected branch & company to respect global filter
    const empRes = await apiClient.get(`/employees?branch_id=${branchId}&company_id=${companyId}&limit=all`)
    const allEmployees = empRes.data?.data || []

    // Filter in-memory by job position title
    drivers.value = allEmployees.filter((emp: any) => {
      const title = (emp.job_position?.title || emp.job_title || '').toLowerCase()
      return title.includes('driver') || title.includes('sopir')
    })

    helpers.value = allEmployees.filter((emp: any) => {
      const title = (emp.job_position?.title || emp.job_title || '').toLowerCase()
      return title.includes('helper') || title.includes('kernet')
    })

    const vRes = await apiClient.get(`/vehicles?company_id=${companyId}`)
    vehicles.value = Array.isArray(vRes.data) ? vRes.data : vRes.data?.data || []
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)

// Watch for branch changes to refresh data
watch(
  () => masterDataStore.selectedBranchId,
  () => {
    fetchData()
  },
)

const handleCreateBatch = async () => {
  // Kumpulkan semua order yang dipilih
  const orders = selectedOrders.value.map((id) =>
    deliveryStore.pendingOrders.find((o) => o.id === id),
  )

  // Tentukan company_id:
  // - Jika cabang spesifik dipilih, gunakan company dari cabang
  // - Jika multi-company (lintas company), kirim company_id dari order pertama
  //   (backend akan mengabaikan validasi company tunggal)
  let companyId = masterDataStore.selectedBranchCompanyId || ''
  if (!companyId || companyId === '') {
    // Ambil company_id dari order pertama yang dipilih
    const firstOrder = orders.find((o) => o?.company_id)
    companyId = firstOrder?.company_id || ''
  }

  // Cek apakah lintas company — informasikan ke user (tapi tidak blokir)
  const distinctCompanies = [...new Set(orders.filter(Boolean).map((o) => o?.company_id))]
  if (distinctCompanies.length > 1) {
    // Batch lintas company diizinkan — company_id diambil dari SO pertama
    console.info('[Delivery] Batch berisi nota dari', distinctCompanies.length, 'company berbeda. Diproses sebagai batch lintas company.')
  }

  isSubmitting.value = true
  try {
    const payload = {
      company_id: companyId || undefined,
      driver_id: form.driver_id || null,
      helper_id: form.helper_id || null,
      vehicle_id: form.vehicle_id || null,
      sales_order_ids: selectedOrders.value,
      collection_only_so_ids: selectedUnpaidOrders.value,
      samples: selectedSamples.value.map(id => {
        const sample = deliveryStore.approvedSamples.find((s: any) => s.id === id)
        return {
          product_id: sample.product_id,
          quantity: sample.quantity,
          recipient_type: sample.store_id ? 'STORE' : 'SALESMAN',
          store_id: sample.store_id || null,
          salesman_id: sample.employee_id || null,
          sales_tester_request_id: sample.id,
          notes: sample.notes
        }
      })
    }

    if (isEditingBatch.value && editingBatchId.value) {
      await apiClient.put(`/delivery/batches/${editingBatchId.value}`, payload)
      toast.success('Batch berhasil diperbarui!')
    } else {
      await deliveryStore.createBatch(payload)
      toast.success('Batch berhasil dibuat! Silakan finalisasi batch di bawah.')
    }

    cancelEdit()
    await fetchData()
  } catch (err: any) {
    toast.error('Gagal memproses batch: ' + (err.response?.data?.error || err.message))
  } finally {
    isSubmitting.value = false
  }
}

const handleEditBatch = (batch: any) => {
  isEditingBatch.value = true
  editingBatchId.value = batch.id
  form.driver_id = batch.driver_id || ''
  form.helper_id = batch.helper_id || ''
  form.vehicle_id = batch.vehicle_id || ''

  // Ensure the orders from this batch are in the list so they can be selected
  batch.items?.forEach((item: any) => {
    const so = item.sales_order || item.sales_transaction
    if (so) {
      const exists = deliveryStore.pendingOrders.find((o: any) => o.id === so.id)
      if (!exists) {
        deliveryStore.pendingOrders.unshift(so)
      }
    }
  })

  selectedOrders.value = batch.items
    .map((i: any) => i.sales_order_id || i.sales_transaction_id)
    .filter(Boolean)

  // Scroll to top
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const cancelEdit = () => {
  isEditingBatch.value = false
  editingBatchId.value = null
  selectedOrders.value = []
  selectedUnpaidOrders.value = []
  selectedSamples.value = []
  form.driver_id = ''
  form.helper_id = ''
  form.vehicle_id = ''
}

const handleSupervisorApprove = async (id: string) => {
  try {
    await apiClient.post(`/delivery/batch/${id}/supervisor-approve`)
    toast.success('Batch berhasil difinalisasi! Menunggu pencetakan Surat Jalan oleh Admin.')
    await fetchData()
  } catch (err: any) {
    toast.error('Gagal memproses batch: ' + (err.response?.data?.error || err.message))
  }
}

const handleDeleteBatch = async (id: string) => {
  if (!confirm('Yakin ingin membatalkan batch ini?')) return
  try {
    // We might need a delete endpoint or just update status to CANCELLED
    // For now assuming we can delete if WAITING_APPROVAL
    await apiClient.delete(`/delivery/batches/${id}`)
    toast.success('Batch berhasil dibatalkan')
    deliveryStore.fetchBatches()
  } catch (err: any) {
    toast.error('Gagal menghapus batch')
  }
}

const printDO = (row: any) => {
  selectedDO.value = row
}

const getMonthLabel = (m: number) => {
  const months = [
    { value: 1, label: 'Januari' },
    { value: 2, label: 'Februari' },
    { value: 3, label: 'Maret' },
    { value: 4, label: 'April' },
    { value: 5, label: 'Mei' },
    { value: 6, label: 'Juni' },
    { value: 7, label: 'Juli' },
    { value: 8, label: 'Agustus' },
    { value: 9, label: 'September' },
    { value: 10, label: 'Oktober' },
    { value: 11, label: 'November' },
    { value: 12, label: 'Desember' },
  ]
  const found = months.find((item) => item.value === m)
  return found ? found.label : ''
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val)
}

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}

const handlePrint = () => {
  window.print()
}

const batchColumns = [
  { key: 'created_at', label: 'Date', type: 'datetime' },
  { key: 'delivery_order_no', label: 'SJ Number' },
  { key: 'driver.first_name', label: 'Driver' },
  { key: 'vehicle.license_plate', label: 'Vehicle' },
  { key: 'items.length', label: 'Total Orders' },
  { key: 'status', label: 'Status' },
]
</script>

<style scoped>
@media print {
  @page {
    size: landscape;
    margin: 0;
  }
  /* Strict Page Lock */
  html,
  body {
    height: 100vh !important;
    margin: 0 !important;
    padding: 0 !important;
    overflow: hidden !important;
    background: white !important;
  }
  body * {
    visibility: hidden;
  }
  .fixed.inset-0,
  .fixed.inset-0 *,
  #surat-jalan-card,
  #surat-jalan-card * {
    visibility: visible;
  }
  .fixed.inset-0 {
    position: absolute !important;
    top: 0 !important;
    left: 0 !important;
    width: 100% !important;
    height: 100vh !important;
    background: white !important;
    padding: 0 !important;
    overflow: hidden !important;
    display: block !important;
    z-index: 9999 !important;
  }
  #surat-jalan-card {
    position: static !important;
    margin: 0 !important;
    width: 100% !important;
    border: none !important;
    box-shadow: none !important;
    transform: none !important;
    /* Optimized zoom for 20 rows + Large QR */
    zoom: 0.8;
    -moz-transform: scale(0.8);
    -moz-transform-origin: top left;
    page-break-after: avoid !important;
    break-after: avoid !important;
  }
  /* Squeeze Header */
  #surat-jalan-card .p-4 {
    padding: 2px 4px !important;
  }
  /* Table formatting - Squeezed */
  #surat-jalan-card table {
    font-size: 7.5pt !important;
    border-collapse: collapse !important;
    width: 100% !important;
    line-height: 1 !important;
  }
  #surat-jalan-card th,
  #surat-jalan-card td {
    padding: 1.5px !important;
    border: 1px solid #000 !important;
    visibility: visible !important;
  }
  .print\:hidden {
    display: none !important;
  }
  #surat-jalan-card div {
    max-height: none !important;
    overflow: visible !important;
    margin: 0 !important;
  }
  /* Signature sizing - Squeezed */
  .mt-6 {
    margin-top: 5px !important;
  }
  .space-y-8 {
    margin: 0 !important;
  }
}
</style>
