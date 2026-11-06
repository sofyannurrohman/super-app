<template>
  <div class="p-8 space-y-8 bg-[#f8fafc] min-h-screen">
    <!-- Header -->
    <div
      class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-white p-8 rounded-[2rem] shadow-sm border border-slate-100"
    >
      <div>
        <h1 class="text-3xl font-black tracking-tight text-slate-900">Manajemen Kendaraan</h1>
        <p class="text-slate-500 mt-1">
          Kelola armada operasional dan pantau riwayat penggunaan kendaraan.
        </p>
      </div>
      <div class="flex gap-2">
        <Button
          @click="$router.push('/vehicles/live-tracking')"
          variant="outline"
          class="rounded-2xl h-12 px-6 font-bold"
        >
          <Map class="mr-2 h-4 w-4" /> Live Tracking
        </Button>
        <Button
          @click="openAddModal"
          class="rounded-2xl shadow-lg shadow-primary/20 font-bold h-12 px-8"
        >
          <Plus class="mr-2 h-4 w-4" /> Tambah Kendaraan
        </Button>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
      <!-- LEFT SIDE: VEHICLE LIST (7/12) -->
      <div class="lg:col-span-7 space-y-4">
        <!-- Section Header & Filter Toolbar -->
        <div class="bg-white p-5 rounded-[2rem] border border-slate-100 shadow-sm space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="text-xl font-black text-slate-900 flex items-center gap-2">
              Daftar Armada
              <Badge variant="secondary" class="rounded-full bg-slate-100 text-slate-600 font-bold font-mono">
                {{ filteredVehicles.length }} / {{ computedVehicles.length }}
              </Badge>
            </h2>

            <Button
              v-if="isFilterActive"
              variant="ghost"
              size="sm"
              @click="resetFilters"
              class="h-8 px-3 rounded-xl text-xs font-bold text-slate-500 hover:text-red-600 hover:bg-red-50 flex items-center gap-1.5"
            >
              <RotateCcw class="h-3.5 w-3.5" /> Reset Filter
            </Button>
          </div>

          <!-- Search & Filter Controls -->
          <div class="grid grid-cols-1 md:grid-cols-12 gap-3">
            <!-- Search Bar (5/12) -->
            <div class="md:col-span-5 relative">
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Cari plat, nama, driver, atau no DO..."
                class="w-full pl-10 pr-8 py-2.5 bg-slate-50 border border-slate-200/80 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-primary focus:ring-4 focus:ring-primary/5 outline-none transition-all"
              />
              <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 pointer-events-none" />
              <button
                v-if="searchQuery"
                @click="searchQuery = ''"
                class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5 rounded-full"
              >
                <X class="h-3.5 w-3.5" />
              </button>
            </div>

            <!-- Filter Cabang (3/12) -->
            <div class="md:col-span-3">
              <select
                v-model="filterBranch"
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200/80 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-primary cursor-pointer transition-all"
              >
                <option value="ALL_BRANCHES">Semua Cabang</option>
                <option value="PUSAT">Kantor Pusat</option>
                <option v-for="b in masterStore.branches" :key="b.id" :value="b.id">
                  {{ b.name }}
                </option>
              </select>
            </div>

            <!-- Filter Tipe Mobil (2/12) -->
            <div class="md:col-span-2">
              <select
                v-model="filterType"
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200/80 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-primary cursor-pointer transition-all"
              >
                <option value="ALL_TYPES">Semua Tipe</option>
                <option value="TRUCK">Truk</option>
                <option value="VAN">Blind Van</option>
                <option value="CAR">Mobil</option>
                <option value="MOTORCYCLE">Motor</option>
              </select>
            </div>

            <!-- Filter Status (2/12) -->
            <div class="md:col-span-2">
              <select
                v-model="filterStatus"
                class="w-full px-3 py-2.5 bg-slate-50 border border-slate-200/80 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-primary cursor-pointer transition-all"
              >
                <option value="ALL_STATUS">Semua Status</option>
                <option value="AVAILABLE">Tersedia</option>
                <option value="IN_USE">Digunakan</option>
                <option value="MAINTENANCE">Servis</option>
                <option value="BROKEN">Rusak</option>
              </select>
            </div>
          </div>
        </div>

        <div
          v-if="vehicleStore.loading"
          class="flex justify-center py-12 bg-white rounded-[2rem] border border-slate-100"
        >
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>

        <div
          v-else-if="filteredVehicles.length === 0"
          class="bg-white rounded-[2rem] border border-slate-100 shadow-sm p-12 text-center space-y-3"
        >
          <div class="h-16 w-16 bg-slate-50 text-slate-300 rounded-3xl flex items-center justify-center mx-auto">
            <Truck class="h-8 w-8" />
          </div>
          <h3 class="font-black text-slate-800 text-base">Tidak ada armada yang sesuai</h3>
          <p class="text-xs text-slate-400 max-w-sm mx-auto">
            Coba sesuaikan kata kunci pencarian atau reset filter cabang/tipe mobil.
          </p>
          <Button variant="outline" size="sm" @click="resetFilters" class="rounded-xl text-xs font-bold">
            Reset Semua Filter
          </Button>
        </div>

        <div
          v-else
          class="bg-white rounded-[2rem] border border-slate-100 shadow-sm overflow-hidden"
        >
          <DataTable :columns="tableColumns" :data="filteredVehicles" class="border-none">
            <template #cell-vehicle="{ row }">
              <div class="flex items-start gap-4 cursor-pointer py-1" @click="selectVehicle(row)">
                <div
                  v-if="row.image_url"
                  class="h-12 w-12 rounded-xl overflow-hidden shadow-sm flex-shrink-0 mt-0.5"
                >
                  <img :src="baseUrl + row.image_url" class="h-full w-full object-cover" />
                </div>
                <div
                  v-else
                  class="h-12 w-12 rounded-xl flex items-center justify-center text-white shadow-sm flex-shrink-0 mt-0.5"
                  :class="statusColors[row.computedStatus]"
                >
                  <Truck class="h-6 w-6" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-black text-slate-900 font-mono">
                      {{ row.license_plate }}
                    </span>
                    <Badge variant="outline" class="text-[9px] font-bold px-1.5 py-0 rounded-md border-slate-200 text-slate-500 uppercase">
                      {{ row.type || 'TRUCK' }}
                    </Badge>
                    <Badge
                      v-if="row.gps_imei"
                      class="text-[9px] font-bold px-1.5 py-0 rounded-md bg-amber-50 text-amber-700 border border-amber-200 flex items-center gap-1 font-mono"
                      :title="'TraxCenter IMEI: ' + row.gps_imei"
                    >
                      <Radio class="h-2.5 w-2.5 text-amber-500 animate-pulse" />
                      GPS
                    </Badge>
                  </div>
                  <div class="text-xs font-bold text-slate-500">{{ row.name }}</div>
                  <div class="text-[10px] text-slate-400 mt-0.5 flex items-center gap-0.5">
                    <MapPin class="h-3 w-3" />{{ row.branch?.name || 'Kantor Pusat' }}
                  </div>

                  <!-- Active Delivery Chip when IN_USE -->
                  <div
                    v-if="row.computedStatus === 'IN_USE' && row.activeAssignment"
                    class="mt-2 p-2 rounded-xl bg-blue-50/80 border border-blue-100 text-blue-800 space-y-0.5 shadow-sm"
                  >
                    <div class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-blue-600">
                      <Truck class="h-3 w-3" />
                      <span>{{ row.activeAssignment.type_label }}</span>
                      <span class="text-blue-400">•</span>
                      <span class="font-mono">{{ row.activeAssignment.do_number }}</span>
                    </div>
                    <div class="text-[11px] font-bold text-slate-700 flex items-center gap-2">
                      <span>👤 {{ row.activeAssignment.driver_name }}</span>
                      <span class="text-slate-300">|</span>
                      <span class="truncate">{{ row.activeAssignment.origin_name }} → {{ row.activeAssignment.destination_name }}</span>
                    </div>
                  </div>

                  <div v-if="row.maintenanceStatus && row.maintenanceStatus !== 'OK'" class="mt-1 flex items-center gap-1" :class="row.maintenanceStatus === 'CRITICAL' ? 'text-red-500' : 'text-amber-500'">
                    <AlertCircle class="h-3.5 w-3.5" />
                    <span class="text-[10px] font-bold">{{ row.maintenanceStatus === 'CRITICAL' ? 'Butuh Servis' : 'Mendekati Servis' }}</span>
                  </div>
                  <div v-if="row.taxStatus && row.taxStatus !== 'OK'" class="mt-1 flex items-center gap-1" :class="row.taxStatus === 'CRITICAL' ? 'text-red-500' : 'text-amber-500'">
                    <AlertCircle class="h-3.5 w-3.5" />
                    <span class="text-[10px] font-bold">{{ row.taxStatus === 'CRITICAL' ? 'Pajak Expired' : 'Pajak Mendekati Expired' }}</span>
                  </div>
                </div>
              </div>
            </template>
            <template #cell-specs="{ row }">
              <div class="flex flex-col gap-1 cursor-pointer" @click="selectVehicle(row)">
                <div class="text-xs font-bold text-slate-700 flex items-center gap-1.5">
                  <Settings class="h-3.5 w-3.5 text-slate-400" />{{ row.model || '-' }}
                </div>
                <div class="text-[11px] font-medium text-slate-500 flex items-center gap-1.5">
                  <Scale class="h-3.5 w-3.5 text-slate-400" />{{ row.capacity }} KG
                </div>
                <div class="text-[11px] font-medium text-slate-500 flex items-center gap-1.5">
                  <Timer class="h-3.5 w-3.5 text-slate-400" />{{
                    row.mileage?.toLocaleString() || 0
                  }}
                  KM
                </div>
              </div>
            </template>
            <template #cell-status="{ row }">
              <div class="cursor-pointer space-y-1" @click="selectVehicle(row)">
                <Badge
                  :class="getStatusBadgeClass(row.computedStatus)"
                  class="rounded-full px-2 py-0.5 font-black text-[10px] uppercase border-none"
                >
                  {{ row.computedStatus }}
                </Badge>
                <div v-if="row.computedStatus === 'IN_USE' && row.activeAssignment" class="text-[10px] font-bold text-blue-600">
                  {{ row.activeAssignment.status_label }}
                </div>
              </div>
            </template>
            <template #cell-actions="{ row }">
              <div class="flex items-center gap-1 justify-end">
                <Button
                  variant="ghost"
                  size="icon"
                  @click.stop="openEditModal(row)"
                  class="h-8 w-8 rounded-xl text-primary hover:bg-primary/10"
                >
                  <Edit2 class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-8 w-8 rounded-xl text-destructive hover:bg-destructive/10"
                  @click.stop="handleDelete(row.id)"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </template>
          </DataTable>
        </div>
      </div>

      <!-- RIGHT SIDE: ACTIVITY LOG & DRIVERS (5/12) -->
      <div class="lg:col-span-5 space-y-6">
        <div class="flex items-center justify-between px-2">
          <div class="flex bg-white p-1 rounded-2xl border border-slate-100 shadow-sm">
            <button
              @click="activeTab = 'LOGS'"
              :class="[
                activeTab === 'LOGS'
                  ? 'bg-slate-900 text-white'
                  : 'text-slate-500 hover:bg-slate-50',
              ]"
              class="px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center gap-2"
            >
              <History class="h-3.5 w-3.5" /> LOG AKTIVITAS
            </button>
            <button
              @click="activeTab = 'DRIVERS'"
              :class="[
                activeTab === 'DRIVERS'
                  ? 'bg-slate-900 text-white'
                  : 'text-slate-500 hover:bg-slate-50',
              ]"
              class="px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center gap-2"
            >
              <User class="h-3.5 w-3.5" /> STATUS DRIVER
            </button>
          </div>
          <Badge
            v-if="activeTab === 'LOGS' && selectedVehicle"
            variant="outline"
            class="rounded-full border-slate-200 text-slate-500 font-mono"
            >{{ selectedVehicle.license_plate }}</Badge
          >
          <Badge
            v-else-if="activeTab === 'DRIVERS'"
            variant="secondary"
            class="rounded-full bg-primary/10 text-primary font-bold"
            >{{ driverStatus.length }}</Badge
          >
        </div>

        <div
          class="bg-white rounded-[2.5rem] border border-slate-100 shadow-sm overflow-hidden min-h-[500px] flex flex-col"
        >
          <!-- TAB: LOGS -->
          <template v-if="activeTab === 'LOGS'">
            <!-- Vehicle Detail Preview if selected -->
            <div
              v-if="selectedVehicle"
              class="p-6 border-b border-slate-50 bg-slate-50/20 space-y-4"
            >
              <div v-if="selectedVehicle.photo_front || selectedVehicle.photo_left || selectedVehicle.photo_right || selectedVehicle.photo_back || selectedVehicle.image_url" class="grid grid-cols-2 gap-2 mb-2">
                <div v-if="selectedVehicle.photo_front || selectedVehicle.image_url" class="relative w-full h-24 rounded-2xl overflow-hidden border border-slate-100 shadow-inner group">
                  <div class="absolute inset-0 bg-black/40 flex items-end p-2 opacity-0 group-hover:opacity-100 transition-opacity"><span class="text-white text-[10px] font-bold">Depan</span></div>
                  <img :src="baseUrl + (selectedVehicle.photo_front || selectedVehicle.image_url)" class="w-full h-full object-cover" />
                </div>
                <div v-if="selectedVehicle.photo_left" class="relative w-full h-24 rounded-2xl overflow-hidden border border-slate-100 shadow-inner group">
                  <div class="absolute inset-0 bg-black/40 flex items-end p-2 opacity-0 group-hover:opacity-100 transition-opacity"><span class="text-white text-[10px] font-bold">Kiri</span></div>
                  <img :src="baseUrl + selectedVehicle.photo_left" class="w-full h-full object-cover" />
                </div>
                <div v-if="selectedVehicle.photo_right" class="relative w-full h-24 rounded-2xl overflow-hidden border border-slate-100 shadow-inner group">
                  <div class="absolute inset-0 bg-black/40 flex items-end p-2 opacity-0 group-hover:opacity-100 transition-opacity"><span class="text-white text-[10px] font-bold">Kanan</span></div>
                  <img :src="baseUrl + selectedVehicle.photo_right" class="w-full h-full object-cover" />
                </div>
                <div v-if="selectedVehicle.photo_back" class="relative w-full h-24 rounded-2xl overflow-hidden border border-slate-100 shadow-inner group">
                  <div class="absolute inset-0 bg-black/40 flex items-end p-2 opacity-0 group-hover:opacity-100 transition-opacity"><span class="text-white text-[10px] font-bold">Belakang</span></div>
                  <img :src="baseUrl + selectedVehicle.photo_back" class="w-full h-full object-cover" />
                </div>
              </div>

              <!-- Tax Alert Banner -->
              <div
                v-if="selectedVehicle.taxStatus && selectedVehicle.taxStatus !== 'OK'"
                class="p-4 border rounded-3xl flex items-start gap-4 shadow-inner"
                :class="selectedVehicle.taxStatus === 'CRITICAL' ? 'bg-red-50 border-red-100' : 'bg-amber-50 border-amber-100'"
              >
                <div
                  class="h-10 w-10 rounded-2xl text-white flex items-center justify-center shadow-lg flex-shrink-0 animate-pulse"
                  :class="selectedVehicle.taxStatus === 'CRITICAL' ? 'bg-red-500 shadow-red-500/20' : 'bg-amber-500 shadow-amber-500/20'"
                >
                  <AlertCircle class="h-5 w-5" />
                </div>
                <div class="space-y-1 min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="text-[9px] font-black uppercase tracking-widest"
                      :class="selectedVehicle.taxStatus === 'CRITICAL' ? 'text-red-600' : 'text-amber-600'"
                      >{{ selectedVehicle.taxStatus === 'CRITICAL' ? 'PAJAK EXPIRED' : 'PAJAK MENDEKATI EXPIRED' }}</span
                    >
                  </div>
                  <h4 class="text-sm font-black text-slate-800 leading-tight">
                    Masa Aktif Pajak: {{ selectedVehicle.tax_expiry_date ? new Date(selectedVehicle.tax_expiry_date).toLocaleDateString('id-ID', {day: 'numeric', month: 'long', year: 'numeric'}) : '-' }}
                  </h4>
                  <p class="text-[10px] text-slate-600 font-bold mt-1">
                    Kendaraan ini {{ selectedVehicle.taxStatus === 'CRITICAL' ? 'telah habis' : 'akan habis' }} masa pajaknya dalam {{ selectedVehicle.taxExpiryDays }} hari.
                  </p>
                </div>
              </div>

              <!-- Comprehensive Active Task Card -->
              <div
                v-if="selectedVehicle.computedStatus === 'IN_USE'"
                class="p-5 bg-gradient-to-br from-blue-50/90 to-indigo-50/70 border border-blue-100 rounded-3xl shadow-sm space-y-4"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="flex items-center gap-3">
                    <div
                      class="h-10 w-10 rounded-2xl bg-blue-600 text-white flex items-center justify-center shadow-md shadow-blue-500/20 flex-shrink-0"
                    >
                      <Truck class="h-5 w-5 animate-pulse" />
                    </div>
                    <div>
                      <div class="flex items-center gap-1.5">
                        <span class="text-[9px] font-black text-blue-600 uppercase tracking-wider">
                          {{ selectedVehicle.activeAssignment?.type_label || 'Tugas Pengiriman Aktif' }}
                        </span>
                        <Badge class="bg-blue-600 text-white text-[8px] font-black px-1.5 py-0.5 rounded-md border-none uppercase">
                          {{ selectedVehicle.activeAssignment?.status_label || 'SEDANG DIGUNAKAN' }}
                        </Badge>
                      </div>
                      <div class="flex items-center gap-2 mt-0.5">
                        <h4 class="text-sm font-black text-slate-900 font-mono">
                          {{ selectedVehicle.activeAssignment?.do_number || selectedVehicle.activeTask || 'DO-PENDING' }}
                        </h4>
                        <button
                          v-if="selectedVehicle.activeAssignment?.do_number"
                          @click="copyToClipboard(selectedVehicle.activeAssignment.do_number)"
                          class="p-1 text-slate-400 hover:text-blue-600 transition-colors"
                          title="Salin No DO"
                        >
                          <Copy class="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </div>
                  </div>

                  <Button
                    v-if="selectedVehicle.activeAssignment?.redirect_path"
                    variant="outline"
                    size="sm"
                    @click="$router.push(selectedVehicle.activeAssignment.redirect_path)"
                    class="h-8 px-3 rounded-xl text-[11px] font-bold bg-white text-blue-600 border-blue-200 hover:bg-blue-50 hover:border-blue-300 shadow-sm"
                  >
                    Detail <ExternalLink class="ml-1 h-3 w-3" />
                  </Button>
                </div>

                <!-- Assignment Meta Grid -->
                <div class="grid grid-cols-2 gap-2 pt-2 border-t border-blue-100/80 text-xs">
                  <div class="bg-white/80 p-2.5 rounded-2xl border border-blue-50">
                    <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Driver / Supir</p>
                    <p class="font-bold text-slate-800 mt-0.5 truncate flex items-center gap-1">
                      <User class="h-3 w-3 text-blue-500" />
                      {{ selectedVehicle.activeAssignment?.driver_name || 'Driver' }}
                    </p>
                    <p v-if="selectedVehicle.activeAssignment?.helper_name" class="text-[10px] text-slate-500 font-medium truncate mt-0.5">
                      Helper: {{ selectedVehicle.activeAssignment.helper_name }}
                    </p>
                  </div>

                  <div class="bg-white/80 p-2.5 rounded-2xl border border-blue-50">
                    <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest">Rute Pengiriman</p>
                    <p class="font-bold text-slate-800 mt-0.5 truncate flex items-center gap-1 text-[11px]">
                      <MapPin class="h-3 w-3 text-emerald-500 flex-shrink-0" />
                      <span class="truncate">{{ selectedVehicle.activeAssignment?.origin_name || 'Pusat' }}</span>
                      <ArrowRight class="h-2.5 w-2.5 text-slate-400 flex-shrink-0" />
                      <span class="truncate">{{ selectedVehicle.activeAssignment?.destination_name || 'Tujuan' }}</span>
                    </p>
                  </div>
                </div>

                <div class="flex items-center justify-between text-[10px] font-bold text-slate-500 px-1 pt-1">
                  <span v-if="selectedVehicle.activeAssignment?.total_items">
                    Muatan: <strong class="text-slate-700">{{ selectedVehicle.activeAssignment.total_items }} unit / toko</strong>
                  </span>
                  <span v-else>Muatan: Pengiriman Logistik</span>
                  <span>
                    Mulai: {{ formatDate(selectedVehicle.activeAssignment?.started_at || selectedVehicle.activeAssignment?.created_at) }}
                  </span>
                </div>
              </div>

              <!-- Maintenance Alert Banner -->
              <div
                v-if="selectedVehicle.maintenanceStatus && selectedVehicle.maintenanceStatus !== 'OK'"
                class="p-4 border rounded-3xl flex items-start gap-4 shadow-inner"
                :class="selectedVehicle.maintenanceStatus === 'CRITICAL' ? 'bg-red-50 border-red-100' : 'bg-amber-50 border-amber-100'"
              >
                <div
                  class="h-10 w-10 rounded-2xl text-white flex items-center justify-center shadow-lg flex-shrink-0 animate-pulse"
                  :class="selectedVehicle.maintenanceStatus === 'CRITICAL' ? 'bg-red-500 shadow-red-500/20' : 'bg-amber-500 shadow-amber-500/20'"
                >
                  <AlertCircle class="h-5 w-5" />
                </div>
                <div class="space-y-2 min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="text-[9px] font-black uppercase tracking-widest"
                      :class="selectedVehicle.maintenanceStatus === 'CRITICAL' ? 'text-red-600' : 'text-amber-600'"
                      >{{ selectedVehicle.maintenanceStatus === 'CRITICAL' ? 'BUTUH SERVIS' : 'MENDEKATI BATAS SERVIS' }}</span
                    >
                  </div>
                  <h4 class="text-sm font-black text-slate-800 leading-tight">
                    Terdapat komponen yang membutuhkan servis
                  </h4>
                  <ul class="text-[11px] text-slate-600 font-bold space-y-1 mt-1">
                    <li v-for="(issue, idx) in selectedVehicle.maintenanceIssues" :key="idx" class="flex items-center justify-between">
                      <span class="flex items-center gap-1">
                         <div class="h-1.5 w-1.5 rounded-full" :class="issue.urgency === 'CRITICAL' ? 'bg-red-500' : 'bg-amber-500'"></div>
                         {{ issue.name }}
                      </span>
                      <span>{{ (issue.current_mileage || 0).toLocaleString() }} / {{ (issue.max_mileage || 0).toLocaleString() }} KM</span>
                    </li>
                  </ul>
                </div>
              </div>

              <div class="grid grid-cols-3 gap-4">
                <div class="bg-white p-3 rounded-2xl border border-slate-100 shadow-sm">
                  <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-1">
                    Kapasitas Muatan
                  </p>
                  <p class="text-lg font-black text-primary">{{ selectedVehicle.capacity }} KG</p>
                </div>
                <div class="bg-white p-3 rounded-2xl border border-slate-100 shadow-sm">
                  <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-1">
                    Volume Bak
                  </p>
                  <p class="text-lg font-black text-primary" v-if="selectedVehicle.bed_volume">{{ selectedVehicle.bed_volume }} m³</p>
                  <p class="text-xs font-bold text-slate-400 mt-1" v-else>Belum Diatur</p>
                </div>
                <div class="bg-white p-3 rounded-2xl border border-slate-100 shadow-sm">
                  <p class="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-1">
                    Tahun Kendaraan
                  </p>
                  <p class="text-lg font-black text-slate-700">{{ selectedVehicle.year }}</p>
                </div>
              </div>
            </div>

            <div
              v-if="!selectedVehicleId"
              class="flex-1 flex flex-col items-center justify-center text-center p-8"
            >
              <div
                class="bg-slate-50 h-20 w-20 rounded-[2rem] flex items-center justify-center text-slate-200 mb-4"
              >
                <History class="h-10 w-10" />
              </div>
              <h3 class="font-bold text-slate-900">Pilih Kendaraan</h3>
              <p class="text-sm text-slate-500 mt-1 max-w-xs mx-auto">
                Klik kendaraan untuk melihat riwayat aktivitasnya.
              </p>
            </div>

            <div v-else-if="loadingLogs" class="flex-1 flex items-center justify-center">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>

            <div v-else class="flex-1 p-2 overflow-y-auto custom-scrollbar min-h-0">
              <div
                v-if="vehicleStore.logs.length === 0"
                class="flex flex-col items-center justify-center h-full text-center p-8 opacity-50"
              >
                <ClipboardList class="h-12 w-12 text-slate-300 mb-3" />
                <p class="text-sm font-bold text-slate-400">Belum ada riwayat aktivitas</p>
              </div>

              <div v-else class="space-y-4 p-4">
                <div
                  v-for="log in vehicleStore.logs"
                  :key="log.id"
                  class="relative pl-8 before:absolute before:left-[11px] before:top-8 before:bottom-[-20px] before:w-[2px] before:bg-slate-100 last:before:hidden"
                >
                  <div
                    class="absolute left-0 top-1 h-6 w-6 rounded-full bg-white border-4 border-slate-50 ring-2 z-10"
                    :class="log.status !== 'COMPLETED' ? 'ring-blue-500/30' : 'ring-primary/20'"
                  ></div>
                  <div
                    class="rounded-2xl p-4 border transition-all hover:shadow-sm"
                    :class="
                      log.status !== 'COMPLETED'
                        ? 'bg-blue-50/30 border-blue-100 ring-2 ring-blue-500/5'
                        : 'bg-slate-50/50 border-slate-100 hover:bg-white'
                    "
                  >
                    <div class="flex justify-between items-start mb-2">
                      <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                        {{ formatDate(log.created_at) }}
                      </p>
                      <Badge
                        :class="[
                          log.status === 'COMPLETED'
                            ? 'bg-emerald-50 text-emerald-600'
                            : log.status === 'WAITING_APPROVAL'
                              ? 'bg-amber-50 text-amber-600'
                              : 'bg-blue-50 text-blue-600',
                        ]"
                        class="border-none text-[9px] font-black uppercase px-2 py-0.5 rounded-full"
                      >
                        {{ log.status }}
                      </Badge>
                    </div>
                    <h4
                      class="font-bold text-slate-900 text-sm mb-1 flex items-center gap-2 font-mono"
                    >
                      {{ log.delivery_order_no || 'Tugas Delivery' }}
                      <span v-if="log.status !== 'COMPLETED'" class="flex h-2 w-2 relative">
                        <span
                          class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"
                        ></span>
                        <span class="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
                      </span>
                    </h4>
                    <div class="flex items-center gap-2 text-xs text-slate-500 font-medium">
                      <User class="h-3 w-3" />
                      <span
                        >Driver:
                        {{
                          log.driver
                            ? log.driver.first_name + ' ' + (log.driver.last_name || '')
                            : 'Unknown'
                        }}</span
                      >
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>

          <!-- TAB: DRIVERS -->
          <template v-else>
            <div v-if="loadingDrivers" class="flex-1 flex items-center justify-center">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>

            <div v-else class="flex-1 p-4 overflow-y-auto custom-scrollbar min-h-0">
              <div class="space-y-3">
                <div
                  v-for="driver in driverStatus"
                  :key="driver.id"
                  class="p-5 bg-white border border-slate-100 rounded-3xl hover:border-primary/20 transition-all group shadow-sm"
                >
                  <div class="flex items-center justify-between mb-3">
                    <div class="flex items-center gap-3">
                      <div
                        class="h-10 w-10 rounded-2xl bg-slate-100 flex items-center justify-center text-slate-400 group-hover:bg-primary/10 group-hover:text-primary transition-all"
                      >
                        <User class="h-5 w-5" />
                      </div>
                      <div>
                        <h4 class="font-black text-slate-900 leading-none">
                          {{ driver.first_name }} {{ driver.last_name }}
                        </h4>
                        <p
                          class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-widest"
                        >
                          {{ driver.job_position?.title || 'DRIVER' }}
                        </p>
                      </div>
                    </div>
                    <Badge
                      :class="[
                        driver.computedStatus === 'AVAILABLE'
                          ? 'bg-emerald-50 text-emerald-600'
                          : driver.computedStatus === 'DELIVERING'
                            ? 'bg-blue-50 text-blue-600'
                            : 'bg-amber-50 text-amber-600',
                      ]"
                      class="rounded-full border-none text-[9px] font-black uppercase tracking-wider"
                    >
                      {{ driver.computedStatus }}
                    </Badge>
                  </div>

                  <div
                    v-if="driver.currentTask"
                    class="mt-4 p-3 bg-slate-50 rounded-2xl flex items-start gap-3 border border-slate-100"
                  >
                    <div class="p-2 bg-white rounded-xl text-primary shadow-sm">
                      <Truck class="h-3.5 w-3.5" />
                    </div>
                    <div>
                      <p
                        class="text-[9px] font-black text-slate-400 uppercase tracking-widest leading-none"
                      >
                        Aktifitas Saat Ini
                      </p>
                      <p class="text-xs font-black text-slate-700 mt-1.5">
                        {{ driver.currentTask }}
                      </p>
                    </div>
                  </div>
                  <div
                    v-else
                    class="mt-4 p-3 border-2 border-dashed border-slate-50 rounded-2xl flex items-center justify-center"
                  >
                    <p
                      class="text-[10px] font-bold text-slate-300 uppercase tracking-widest italic"
                    >
                      Belum ada penugasan
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- Modal Form -->
    <div
      v-if="showModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-md z-[100] flex items-center justify-center p-4"
    >
      <Card
        class="w-full max-w-xl shadow-2xl rounded-[2.5rem] overflow-hidden border-none animate-in fade-in zoom-in-95 duration-300"
      >
        <CardHeader class="p-8 border-b bg-white">
          <div class="flex justify-between items-center">
            <div>
              <CardTitle class="text-2xl font-black text-slate-900">{{
                isEditing ? 'Edit Kendaraan' : 'Tambah Kendaraan'
              }}</CardTitle>
              <CardDescription>Lengkapi spesifikasi armada operasional.</CardDescription>
            </div>
            <Button variant="ghost" size="icon" @click="showModal = false" class="rounded-full">
              <Plus class="h-6 w-6 rotate-45 text-slate-400" />
            </Button>
          </div>
        </CardHeader>
        <CardContent
          class="p-8 space-y-6 bg-white max-h-[70vh] overflow-y-auto custom-scrollbar pr-4"
        >
          <!-- Image Upload Area -->
          <div class="space-y-4">
            <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1">Foto Kendaraan</label>
            <div class="grid grid-cols-2 gap-4">
              <!-- Depan -->
              <div @click="fileInputFront?.click()" class="relative w-full h-24 rounded-2xl border-2 border-dashed border-slate-200 bg-slate-50 flex flex-col items-center justify-center cursor-pointer hover:border-primary transition-all group overflow-hidden">
                <img v-if="imagePreviewFront" :src="imagePreviewFront" class="absolute inset-0 w-full h-full object-cover" />
                <div v-else class="flex flex-col items-center text-slate-400 group-hover:text-primary"><Camera class="h-6 w-6" /><span class="text-[10px] font-bold uppercase mt-1">Depan</span></div>
                <input ref="fileInputFront" type="file" accept="image/*" class="hidden" @change="e => handleImageChange(e, 'front')" />
              </div>
              <!-- Kiri -->
              <div @click="fileInputLeft?.click()" class="relative w-full h-24 rounded-2xl border-2 border-dashed border-slate-200 bg-slate-50 flex flex-col items-center justify-center cursor-pointer hover:border-primary transition-all group overflow-hidden">
                <img v-if="imagePreviewLeft" :src="imagePreviewLeft" class="absolute inset-0 w-full h-full object-cover" />
                <div v-else class="flex flex-col items-center text-slate-400 group-hover:text-primary"><Camera class="h-6 w-6" /><span class="text-[10px] font-bold uppercase mt-1">Samping Kiri</span></div>
                <input ref="fileInputLeft" type="file" accept="image/*" class="hidden" @change="e => handleImageChange(e, 'left')" />
              </div>
              <!-- Kanan -->
              <div @click="fileInputRight?.click()" class="relative w-full h-24 rounded-2xl border-2 border-dashed border-slate-200 bg-slate-50 flex flex-col items-center justify-center cursor-pointer hover:border-primary transition-all group overflow-hidden">
                <img v-if="imagePreviewRight" :src="imagePreviewRight" class="absolute inset-0 w-full h-full object-cover" />
                <div v-else class="flex flex-col items-center text-slate-400 group-hover:text-primary"><Camera class="h-6 w-6" /><span class="text-[10px] font-bold uppercase mt-1">Samping Kanan</span></div>
                <input ref="fileInputRight" type="file" accept="image/*" class="hidden" @change="e => handleImageChange(e, 'right')" />
              </div>
              <!-- Belakang -->
              <div @click="fileInputBack?.click()" class="relative w-full h-24 rounded-2xl border-2 border-dashed border-slate-200 bg-slate-50 flex flex-col items-center justify-center cursor-pointer hover:border-primary transition-all group overflow-hidden">
                <img v-if="imagePreviewBack" :src="imagePreviewBack" class="absolute inset-0 w-full h-full object-cover" />
                <div v-else class="flex flex-col items-center text-slate-400 group-hover:text-primary"><Camera class="h-6 w-6" /><span class="text-[10px] font-bold uppercase mt-1">Belakang</span></div>
                <input ref="fileInputBack" type="file" accept="image/*" class="hidden" @change="e => handleImageChange(e, 'back')" />
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
              >Masa Aktif Pajak</label
            >
            <input
              v-model="form.tax_expiry_date"
              type="date"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-4 focus:ring-primary/5 transition-all"
            />
          </div>

          <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
                >Nama Unit</label
              >
              <input
                v-model="form.name"
                placeholder="Truk Wingbox A1"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-4 focus:ring-primary/5 transition-all"
              />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
                >Plat Nomor</label
              >
              <input
                v-model="form.license_plate"
                placeholder="B 1234 ABC"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold font-mono outline-none focus:ring-4 focus:ring-primary/5 transition-all"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
                >Model / Brand</label
              >
              <input
                v-model="form.model"
                placeholder="Hino 500 / Mitsubishi"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-4 focus:ring-primary/5 transition-all"
              />
            </div>
            <div class="space-y-2">
              <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
                >Kapasitas Muatan (KG)</label
              >
              <div class="relative">
                <input
                  type="number"
                  v-model="form.capacity"
                  placeholder="0.00"
                  class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none focus:ring-4 focus:ring-primary/5 transition-all pl-12"
                />
                <Scale class="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              </div>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
              <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
                >Tipe Kendaraan</label
              >
              <select
                v-model="form.type"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none appearance-none cursor-pointer"
              >
                <option value="TRUCK">Truk</option>
                <option value="VAN">Blind Van</option>
                <option value="CAR">Mobil Penumpang</option>
                <option value="MOTORCYCLE">Sepeda Motor</option>
              </select>
            </div>
            <div class="space-y-2">
              <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
                >Tahun Armada</label
              >
              <input
                type="number"
                v-model="form.year"
                class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1">Dimensi Bak Armada (Meter)</label>
            <div class="grid grid-cols-4 gap-4">
              <div class="relative">
                <input type="number" step="0.01" v-model="form.bed_length" placeholder="Panjang" class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none pr-8" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400">P</span>
              </div>
              <div class="relative">
                <input type="number" step="0.01" v-model="form.bed_width" placeholder="Lebar" class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none pr-8" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400">L</span>
              </div>
              <div class="relative">
                <input type="number" step="0.01" v-model="form.bed_height" placeholder="Tinggi" class="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none pr-8" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400">T</span>
              </div>
              <div class="relative">
                <input type="number" :value="form.bed_volume" disabled placeholder="Volume" class="w-full px-4 py-3 bg-slate-100 border border-slate-200 rounded-2xl text-sm font-bold text-slate-500 outline-none pr-8 cursor-not-allowed" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-bold text-slate-400">m³</span>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
              >Cabang</label
            >
            <select
              v-model="form.branch_id"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none appearance-none cursor-pointer"
            >
              <option value="">Pusat</option>
              <option v-for="branch in masterStore.branches" :key="branch.id" :value="branch.id">
                {{ branch.name }}
              </option>
            </select>
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1 flex items-center gap-1.5">
                <Radio class="h-3.5 w-3.5 text-amber-500" />
                IMEI GPS Tracker (TraxCenter)
              </label>
              <button
                v-if="form.gps_imei"
                type="button"
                @click="testGpsConnection"
                :disabled="isTestingGps"
                class="text-[11px] font-bold text-amber-600 hover:text-amber-700 bg-amber-50 hover:bg-amber-100/80 px-2.5 py-1 rounded-xl transition-all flex items-center gap-1 disabled:opacity-50"
              >
                <RefreshCw class="h-3 w-3" :class="{ 'animate-spin': isTestingGps }" />
                {{ isTestingGps ? 'Mengecek...' : 'Tes Sinyal GPS' }}
              </button>
            </div>
            <input
              v-model="form.gps_imei"
              placeholder="Contoh: 356153590836737"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold font-mono outline-none focus:ring-4 focus:ring-amber-500/10 transition-all"
            />
            <p v-if="gpsTestResult" class="text-xs px-2 py-1.5 rounded-xl font-bold flex items-center gap-1.5" :class="gpsTestResult.valid ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-rose-50 text-rose-700 border border-rose-200'">
              <span class="h-2 w-2 rounded-full" :class="gpsTestResult.valid ? 'bg-emerald-500' : 'bg-rose-500'"></span>
              {{ gpsTestResult.message }}
            </p>
            <p v-else class="text-[11px] text-slate-400 font-medium ml-1">
              Digunakan untuk tracking pengiriman live via open API TraxCenter (https://trax.center/api/?&lt;imei&gt;).
            </p>
          </div>

          <div class="space-y-2">
            <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
              >Kilometer Terakhir (KM)</label
            >
            <input
              type="number"
              v-model="form.mileage"
              class="w-full px-5 py-3.5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold outline-none"
            />
          </div>

          <div class="space-y-2">
            <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1"
              >Status Operasional</label
            >
            <div class="grid grid-cols-2 gap-3">
              <div
                v-for="status in ['AVAILABLE', 'IN_USE', 'MAINTENANCE', 'BROKEN']"
                :key="status"
                @click="form.status = status"
                class="flex items-center gap-3 p-4 rounded-2xl border cursor-pointer transition-all"
                :class="[
                  form.status === status
                    ? 'bg-primary/5 border-primary shadow-sm'
                    : 'bg-slate-50 border-slate-100 opacity-60',
                ]"
              >
                <div class="h-3 w-3 rounded-full" :class="statusColors[status]"></div>
                <span class="text-xs font-black text-slate-700">{{ status }}</span>
              </div>
            </div>
          </div>

          <!-- Bagian Komponen Rutin (Dinamis) -->
          <div class="mt-4 pt-4 border-t border-slate-100">
            <div class="flex items-center justify-between mb-3">
              <label class="text-xs font-black text-slate-400 uppercase tracking-widest ml-1">Komponen Servis Rutin (Ban, Oli, dll)</label>
              <Button type="button" variant="ghost" size="sm" @click="addComponent" class="text-primary text-xs font-bold gap-1 px-3 py-1">
                <Plus class="h-3.5 w-3.5" /> Tambah
              </Button>
            </div>
            
            <div v-if="form.components.length === 0" class="text-center py-4 bg-slate-50 border border-slate-100 border-dashed rounded-2xl text-xs text-slate-400 font-medium">
              Belum ada data komponen, klik tambah untuk melacak penggantian komponen (misal: ganti ban tiap 40,000 KM).
            </div>

            <div class="space-y-3">
              <div v-for="(comp, index) in form.components" :key="index" class="flex items-center gap-3 bg-slate-50 border border-slate-100 p-3 rounded-2xl">
                <div class="flex-1 space-y-1">
                  <input
                    v-model="comp.name"
                    placeholder="Nama (contoh: Oli Mesin)"
                    class="w-full px-3 py-2 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none focus:border-primary transition-all"
                  />
                </div>
                <div class="flex-1 space-y-1">
                  <div class="relative">
                    <input
                      type="number"
                      v-model="comp.max_mileage"
                      placeholder="Batas (KM)"
                      class="w-full px-3 py-2 pl-9 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none focus:border-primary transition-all"
                    />
                    <Gauge class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400" />
                  </div>
                </div>
                <button type="button" @click="removeComponent(index)" class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-xl transition-all">
                  <Trash2 class="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </CardContent>
        <CardFooter class="bg-slate-50/50 border-t p-8 flex justify-end gap-4">
          <Button variant="ghost" @click="showModal = false" class="rounded-2xl font-bold h-12 px-8"
            >Batal</Button
          >
          <Button
            @click="handleSubmit"
            :loading="loading"
            class="rounded-2xl px-12 h-12 shadow-xl shadow-primary/20 font-black text-white bg-primary hover:bg-primary/90"
          >
            {{ isEditing ? 'Simpan Data' : 'Tambah Armada' }}
          </Button>
        </CardFooter>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed, watch } from 'vue'
import {
  MapPin,
  Calendar,
  AlertCircle,
  Truck,
  Edit2,
  Trash2,
  Camera,
  RefreshCw,
  Scale,
  Plus,
  Gauge,
  Timer,
  Settings,
  History,
  ClipboardList,
  User,
  Map,
  Search,
  X,
  Copy,
  ExternalLink,
  ArrowRight,
  RotateCcw,
  Radio,
  Navigation,
} from 'lucide-vue-next'
import { useVehicleStore } from '@/stores/vehicle'
import { useDeliveryStore } from '@/stores/delivery'
import { useMasterDataStore } from '@/stores/masterData'
import { useFactoryStore } from '@/stores/factory'
import apiClient from '@/api/axios'
import DataTable from '@/components/DataTable.vue'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card/index'
import { Badge } from '@/components/ui/badge/index'
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

const vehicleStore = useVehicleStore()
const deliveryStore = useDeliveryStore()
const masterStore = useMasterDataStore()
const factoryStore = useFactoryStore()

const showModal = ref(false)
const isEditing = ref(false)
const loading = ref(false)
const loadingLogs = ref(false)
const loadingDrivers = ref(false)
const currentId = ref('')
const selectedVehicleId = ref('')
const activeTab = ref('LOGS') // 'LOGS' or 'DRIVERS'

// Filters & Search
const searchQuery = ref('')
const filterBranch = ref('ALL_BRANCHES')
const filterType = ref('ALL_TYPES')
const filterStatus = ref('ALL_STATUS')

const isFilterActive = computed(() => {
  return (
    searchQuery.value.trim() !== '' ||
    filterBranch.value !== 'ALL_BRANCHES' ||
    filterType.value !== 'ALL_TYPES' ||
    filterStatus.value !== 'ALL_STATUS'
  )
})

const resetFilters = () => {
  searchQuery.value = ''
  filterBranch.value = 'ALL_BRANCHES'
  filterType.value = 'ALL_TYPES'
  filterStatus.value = 'ALL_STATUS'
}

const copyToClipboard = (text: string) => {
  if (!text) return
  navigator.clipboard.writeText(text)
  toast.success(`No. ${text} disalin ke clipboard`)
}

const imageFileFront = ref<File | null>(null)
const imagePreviewFront = ref('')
const fileInputFront = ref<HTMLInputElement | null>(null)

const imageFileLeft = ref<File | null>(null)
const imagePreviewLeft = ref('')
const fileInputLeft = ref<HTMLInputElement | null>(null)

const imageFileRight = ref<File | null>(null)
const imagePreviewRight = ref('')
const fileInputRight = ref<HTMLInputElement | null>(null)

const imageFileBack = ref<File | null>(null)
const imagePreviewBack = ref('')
const fileInputBack = ref<HTMLInputElement | null>(null)

const drivers = ref<any[]>([])

const baseUrl = ''

const tableColumns = [
  { key: 'vehicle', label: 'KENDARAAN' },
  { key: 'specs', label: 'SPESIFIKASI' },
  { key: 'status', label: 'STATUS' },
  { key: 'actions', label: '' },
]

const form = reactive({
  name: '',
  license_plate: '',
  model: '',
  type: 'TRUCK',
  status: 'AVAILABLE',
  year: new Date().getFullYear(),
  mileage: 0,
  capacity: 0,
  bed_length: 0,
  bed_width: 0,
  bed_height: 0,
  bed_volume: 0,
  branch_id: '',
  tax_expiry_date: '',
  gps_imei: '',
  components: [] as Array<{ id?: string; name: string; max_mileage: number }>,
})

watch(() => [form.bed_length, form.bed_width, form.bed_height], ([l, w, h]) => {
  form.bed_volume = Number((Number(l) * Number(w) * Number(h)).toFixed(2))
})

const addComponent = () => {
  form.components.push({ name: '', max_mileage: 10000 })
}

const removeComponent = (index: number) => {
  form.components.splice(index, 1)
}

const fetchDrivers = async () => {
  loadingDrivers.value = true
  try {
    await masterStore.fetchEmployees()
    await deliveryStore.fetchBatches()
  } finally {
    loadingDrivers.value = false
  }
}

const driverStatus = computed(() => {
  const filteredDrivers = masterStore.employees.filter((emp: any) => {
    const activeBatch = deliveryStore.batches.find(
      (b) => b.driver_id === emp.id && b.status !== 'COMPLETED',
    )
    const activeTransfer = factoryStore.allTransfers.find(
      (t: any) => t.driver_id === emp.id && (t.status === 'SHIPPED' || t.status === 'ARRIVED'),
    )

    // Always include if they are actively delivering
    if (activeBatch || activeTransfer) return true

    // If job title is missing, include them just in case they are assigned
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

  const statusArray = filteredDrivers.map((driver) => {
    const activeBatch = deliveryStore.batches.find(
      (b) => b.driver_id === driver.id && b.status !== 'COMPLETED',
    )
    const activeTransfer = factoryStore.allTransfers.find(
      (t: any) => t.driver_id === driver.id && (t.status === 'SHIPPED' || t.status === 'ARRIVED'),
    )

    let status = 'AVAILABLE'
    let currentTask = ''

    if (activeTransfer) {
      status = 'DELIVERING'
      const from = activeTransfer.from_factory?.name || 'Pabrik'
      const to = activeTransfer.to_branch?.name || 'Cabang'
      currentTask = `${activeTransfer.delivery_order_no || 'Pindah Gudang'} (${from} → ${to})`
    } else if (activeBatch) {
      if (
        activeBatch.status === 'ON_DELIVERY' ||
        activeBatch.status === 'PICKING' ||
        activeBatch.status === 'SUPERVISOR_APPROVED'
      ) {
        status = 'DELIVERING'
        currentTask = `${activeBatch.delivery_order_no || 'Tugas Pengiriman'} (Cabang → Pelanggan)`
      } else {
        status = 'WORKING'
        currentTask = 'Persiapan / Standby'
      }
    }

    return {
      ...driver,
      computedStatus: status,
      currentTask,
    }
  })

  return statusArray.sort((a, b) => {
    const rank = (s: string) => (s === 'DELIVERING' ? 0 : s === 'WORKING' ? 1 : 2)
    return rank(a.computedStatus) - rank(b.computedStatus)
  })
})

const computedVehicles = computed(() => {
  return vehicleStore.vehicles.map((vehicle) => {
    const activeTransfer = factoryStore.allTransfers.find(
      (t: any) => t.vehicle_id === vehicle.id && (t.status === 'SHIPPED' || t.status === 'ARRIVED' || t.status === 'APPROVED' || t.status === 'PREPARING'),
    )

    const activeBatch = deliveryStore.batches.find(
      (b: any) => b.vehicle_id === vehicle.id && b.status !== 'COMPLETED',
    )

    let currentStatus = vehicle.status
    if (vehicle.active_assignment) {
      currentStatus = 'IN_USE'
    } else if (currentStatus === 'AVAILABLE') {
      if (
        activeTransfer ||
        (activeBatch &&
          (activeBatch.status === 'ON_DELIVERY' ||
            activeBatch.status === 'PICKING' ||
            activeBatch.status === 'SUPERVISOR_APPROVED'))
      ) {
        currentStatus = 'IN_USE'
      }
    }

    // Resolve active assignment details
    let activeAssignment = vehicle.active_assignment
    if (!activeAssignment) {
      if (activeTransfer) {
        activeAssignment = {
          delivery_type: 'FACTORY_TRANSFER',
          type_label: 'Pindah Gudang Pabrik',
          reference_id: activeTransfer.id,
          do_number: activeTransfer.delivery_order_no || 'DO-TRANSFER',
          status: activeTransfer.status,
          status_label: activeTransfer.status === 'SHIPPED' ? 'Dalam Perjalanan' : (activeTransfer.status === 'ARRIVED' ? 'Tiba di Tujuan' : activeTransfer.status),
          driver_name: activeTransfer.driver ? `${activeTransfer.driver.first_name} ${activeTransfer.driver.last_name || ''}` : 'Driver Pabrik',
          origin_name: activeTransfer.from_factory?.name || 'Pabrik Utama',
          destination_name: activeTransfer.to_branch?.name || 'Cabang Tujuan',
          total_items: activeTransfer.quantity,
          created_at: activeTransfer.created_at,
          redirect_path: '/factory/transfers',
        }
      } else if (activeBatch) {
        activeAssignment = {
          delivery_type: 'SALES_DELIVERY',
          type_label: 'Pengiriman Sales Cabang',
          reference_id: activeBatch.id,
          do_number: activeBatch.delivery_order_no || 'SJ-BATCH',
          status: activeBatch.status,
          status_label: activeBatch.status === 'ON_DELIVERY' ? 'Sedang Dikirim' : (activeBatch.status === 'PICKING' ? 'Muat Barang Gudang' : 'Disetujui SPV'),
          driver_name: activeBatch.driver ? `${activeBatch.driver.first_name} ${activeBatch.driver.last_name || ''}` : (activeBatch.daily_driver?.name ? `${activeBatch.daily_driver.name} (Harian)` : 'Driver Cabang'),
          helper_name: activeBatch.helper ? `${activeBatch.helper.first_name} ${activeBatch.helper.last_name || ''}` : activeBatch.daily_helper?.name,
          origin_name: 'Gudang Cabang',
          destination_name: 'Toko Pelanggan',
          total_items: activeBatch.items?.length || 0,
          created_at: activeBatch.created_at,
          redirect_path: '/sales/surat-jalan',
        }
      }
    }

    let maintenanceStatus = 'OK'
    const maintenanceIssues: any[] = []

    if ((vehicle as any).components && Array.isArray((vehicle as any).components)) {
      (vehicle as any).components.forEach((comp: any) => {
        if (comp.is_active && comp.max_mileage > 0) {
          const ratio = (comp.current_mileage || 0) / comp.max_mileage
          if (ratio >= 1.0) {
            maintenanceIssues.push({ ...comp, urgency: 'CRITICAL' })
            maintenanceStatus = 'CRITICAL'
          } else if (ratio >= 0.9) {
            maintenanceIssues.push({ ...comp, urgency: 'WARNING' })
            if (maintenanceStatus === 'OK') maintenanceStatus = 'WARNING'
          }
        }
      })
    }

    let taxStatus = 'OK'
    let taxExpiryDays = 999
    if ((vehicle as any).tax_expiry_date) {
      const expiry = new Date((vehicle as any).tax_expiry_date)
      const now = new Date()
      const diffTime = expiry.getTime() - now.getTime()
      taxExpiryDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
      if (taxExpiryDays <= 0) {
        taxStatus = 'CRITICAL' // expired
      } else if (taxExpiryDays <= 30) {
        taxStatus = 'WARNING'
      }
    }

    return {
      ...vehicle,
      computedStatus: currentStatus,
      activeAssignment,
      activeTask: activeAssignment?.do_number || activeTransfer?.delivery_order_no || activeBatch?.delivery_order_no || '',
      maintenanceStatus,
      maintenanceIssues,
      taxStatus,
      taxExpiryDays,
    }
  })
})

const filteredVehicles = computed(() => {
  return computedVehicles.value.filter((vehicle) => {
    // Search query filter
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const matchPlate = vehicle.license_plate?.toLowerCase().includes(q)
      const matchName = vehicle.name?.toLowerCase().includes(q)
      const matchModel = vehicle.model?.toLowerCase().includes(q)
      const matchDO = vehicle.activeAssignment?.do_number?.toLowerCase().includes(q)
      const matchDriver = vehicle.activeAssignment?.driver_name?.toLowerCase().includes(q)
      if (!matchPlate && !matchName && !matchModel && !matchDO && !matchDriver) {
        return false
      }
    }

    // Branch filter
    if (filterBranch.value !== 'ALL_BRANCHES') {
      if (filterBranch.value === 'PUSAT') {
        if (vehicle.branch_id) return false
      } else {
        if (vehicle.branch_id !== filterBranch.value) return false
      }
    }

    // Vehicle Type filter
    if (filterType.value !== 'ALL_TYPES') {
      if (vehicle.type?.toUpperCase() !== filterType.value.toUpperCase()) {
        return false
      }
    }

    // Status filter
    if (filterStatus.value !== 'ALL_STATUS') {
      if (vehicle.computedStatus !== filterStatus.value) {
        return false
      }
    }

    return true
  })
})

onMounted(() => {
  masterStore.fetchBranches()
  vehicleStore.fetchVehicles()
  fetchDrivers()
  factoryStore.fetchAllTransfers()
})

watch(
  () => masterStore.selectedBranchId,
  () => {
    vehicleStore.fetchVehicles()
    fetchDrivers()
    factoryStore.fetchAllTransfers()
  },
)

const selectedVehicle = computed(() => {
  return computedVehicles.value.find((v) => v.id === selectedVehicleId.value)
})

const selectVehicle = async (vehicle: any) => {
  selectedVehicleId.value = vehicle.id
  loadingLogs.value = true
  try {
    await vehicleStore.fetchVehicleLogs(vehicle.id)
    if (vehicle.computedStatus === 'IN_USE') {
      const activeTask = await vehicleStore.fetchVehicleActiveTask(vehicle.id)
      if (activeTask) {
        vehicle.active_assignment = activeTask
      }
    }
  } finally {
    loadingLogs.value = false
  }
}

const handleImageChange = (e: any, type: string) => {
  const file = e.target.files[0]
  if (file) {
    const url = URL.createObjectURL(file)
    if (type === 'front') {
      imageFileFront.value = file
      imagePreviewFront.value = url
    } else if (type === 'left') {
      imageFileLeft.value = file
      imagePreviewLeft.value = url
    } else if (type === 'right') {
      imageFileRight.value = file
      imagePreviewRight.value = url
    } else if (type === 'back') {
      imageFileBack.value = file
      imagePreviewBack.value = url
    }
  }
}

const isTestingGps = ref(false)
const gpsTestResult = ref<{ valid: boolean; message: string } | null>(null)

const testGpsConnection = async () => {
  const imei = form.gps_imei?.trim()
  if (!imei) {
    toast.error('Masukkan IMEI GPS Tracker terlebih dahulu')
    return
  }

  isTestingGps.value = true
  gpsTestResult.value = null
  try {
    const res = await apiClient.get(`/vehicles/traxcenter/${imei}`)
    const data = res.data
    if (data && data.lat !== false && data.lat !== 'false' && data.lat !== undefined && data.lat !== null) {
      gpsTestResult.value = {
        valid: true,
        message: `Sinyal Aktif! Lat: ${data.lat}, Lon: ${data.lon}, Speed: ${data.speed || 0} km/h, Heading: ${data.heading || 0}°`
      }
      toast.success('Koneksi GPS TraxCenter berhasil terhubung')
    } else {
      gpsTestResult.value = {
        valid: false,
        message: 'Perangkat GPS belum mengirimkan sinyal koordinat aktif (lat/lon bernilai false).'
      }
      toast.warning('GPS terdaftar namun belum ada sinyal koordinat aktif')
    }
  } catch (err: any) {
    const msg = err.response?.data?.error || err.message || 'Gagal menghubungi server TraxCenter'
    gpsTestResult.value = {
      valid: false,
      message: `Gagal cek GPS: ${msg}`
    }
    toast.error('Gagal memeriksa GPS TraxCenter')
  } finally {
    isTestingGps.value = false
  }
}

const openAddModal = () => {
  isEditing.value = false
  imageFileFront.value = null; imagePreviewFront.value = ''
  imageFileLeft.value = null; imagePreviewLeft.value = ''
  imageFileRight.value = null; imagePreviewRight.value = ''
  imageFileBack.value = null; imagePreviewBack.value = ''
  gpsTestResult.value = null
  
  Object.assign(form, {
    name: '',
    license_plate: '',
    model: '',
    type: 'TRUCK',
    status: 'AVAILABLE',
    year: new Date().getFullYear(),
    mileage: 0,
    capacity: 0,
    bed_length: 0,
    bed_width: 0,
    bed_height: 0,
    bed_volume: 0,
    branch_id: '',
    tax_expiry_date: '',
    gps_imei: '',
    components: [],
  })
  showModal.value = true
}

const openEditModal = (vehicle: any) => {
  isEditing.value = true
  currentId.value = vehicle.id
  gpsTestResult.value = null
  
  imageFileFront.value = null
  imagePreviewFront.value = (vehicle.photo_front || vehicle.image_url) ? baseUrl + (vehicle.photo_front || vehicle.image_url) : ''
  imageFileLeft.value = null
  imagePreviewLeft.value = vehicle.photo_left ? baseUrl + vehicle.photo_left : ''
  imageFileRight.value = null
  imagePreviewRight.value = vehicle.photo_right ? baseUrl + vehicle.photo_right : ''
  imageFileBack.value = null
  imagePreviewBack.value = vehicle.photo_back ? baseUrl + vehicle.photo_back : ''
  
  let formattedTaxDate = ''
  if (vehicle.tax_expiry_date) {
    const d = new Date(vehicle.tax_expiry_date)
    const yyyy = d.getFullYear()
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    formattedTaxDate = `${yyyy}-${mm}-${dd}`
  }

  Object.assign(form, {
    name: vehicle.name,
    license_plate: vehicle.license_plate,
    model: vehicle.model,
    type: vehicle.type,
    status: vehicle.status,
    year: vehicle.year,
    mileage: vehicle.mileage,
    capacity: vehicle.capacity || 0,
    bed_length: vehicle.bed_length || 0,
    bed_width: vehicle.bed_width || 0,
    bed_height: vehicle.bed_height || 0,
    bed_volume: vehicle.bed_volume || 0,
    branch_id: vehicle.branch_id || '',
    tax_expiry_date: formattedTaxDate,
    gps_imei: vehicle.gps_imei || '',
    components: vehicle.components ? JSON.parse(JSON.stringify(vehicle.components)) : [],
  })
  showModal.value = true
}

const handleSubmit = async () => {
  if (!form.name || !form.license_plate) {
    toast.error('Nama dan Plat Nomor harus diisi')
    return
  }

  const formData = new FormData()
  Object.keys(form).forEach((key) => {
    if (key === 'components') return // handled below
    const val = (form as any)[key]
    if (val !== null && val !== undefined) {
      formData.append(key, val.toString())
    }
  })

  // Append components
  if (form.components && form.components.length > 0) {
    formData.append('components_json', JSON.stringify(form.components))
  }

  if (imageFileFront.value) {
    formData.append('photo_front', imageFileFront.value)
  }
  if (imageFileLeft.value) {
    formData.append('photo_left', imageFileLeft.value)
  }
  if (imageFileRight.value) {
    formData.append('photo_right', imageFileRight.value)
  }
  if (imageFileBack.value) {
    formData.append('photo_back', imageFileBack.value)
  }

  loading.value = true
  try {
    if (isEditing.value) {
      await vehicleStore.updateVehicle(currentId.value, formData)
      toast.success('Data kendaraan diperbarui')
    } else {
      await vehicleStore.addVehicle(formData)
      toast.success('Kendaraan baru ditambahkan')
    }
    showModal.value = false
  } catch (err) {
    toast.error('Gagal menyimpan: ' + getErrorMessage(err))
  } finally {
    loading.value = false
  }
}

const handleDelete = async (id: string) => {
  if (!confirm('Hapus kendaraan ini dari armada?')) return
  try {
    await vehicleStore.removeVehicle(id)
    toast.success('Kendaraan dihapus')
    if (selectedVehicleId.value === id) {
      selectedVehicleId.value = ''
    }
  } catch (err) {
    toast.error('Gagal menghapus: ' + getErrorMessage(err))
  }
}

const formatDate = (dateStr?: string | null) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const statusColors: Record<string, string> = {
  AVAILABLE: 'bg-emerald-500',
  IN_USE: 'bg-blue-500',
  MAINTENANCE: 'bg-orange-500',
  BROKEN: 'bg-rose-500',
}

const getStatusBadgeClass = (status: string) => {
  switch (status) {
    case 'AVAILABLE':
      return 'bg-emerald-50 text-emerald-600'
    case 'IN_USE':
      return 'bg-blue-50 text-blue-600'
    case 'MAINTENANCE':
      return 'bg-orange-50 text-orange-600'
    case 'BROKEN':
      return 'bg-rose-50 text-rose-600'
    default:
      return 'bg-slate-50 text-slate-500'
  }
}
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
.mask-gradient {
  mask-image: linear-gradient(to left, rgba(0, 0, 0, 1), rgba(0, 0, 0, 0));
}
</style>
