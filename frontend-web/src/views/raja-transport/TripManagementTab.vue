<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRajaTransportStore } from '@/stores/rajaTransport'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import { vehicleApi } from '@/api/vehicle'
import apiClient from '@/api/axios'
import {
  Truck,
  Plus,
  Search,
  Filter,
  CheckCircle2,
  Clock,
  MapPin,
  Calendar,
  DollarSign,
  AlertCircle,
  FileText,
  User,
  Check,
  X,
  Eye,
  RefreshCw,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const transportStore = useRajaTransportStore()
const masterDataStore = useMasterDataStore()
const authStore = useAuthStore()

const isAddModalOpen = ref(false)
const isDetailModalOpen = ref(false)
const isCompleteModalOpen = ref(false)
const selectedTrip = ref<any | null>(null)

// Vehicles & Drivers master data for form dropdowns
const vehiclesList = ref<any[]>([])
const driversList = ref<any[]>([])
const loadingMasters = ref(false)

// Filters
const filterStatus = ref('ALL')
const filterRegion = ref('ALL')
const searchQuery = ref('')

// Form State
const form = ref({
  vehicle_id: '',
  driver_id: '',
  helper_id: '',
  origin_city: 'Sukoharjo',
  origin_province: 'Jawa Tengah',
  destination_city: '',
  destination_province: '',
  region_type: 'IN_PROVINCE',
  departure_date: new Date().toISOString().split('T')[0],
  estimated_return_date: new Date().toISOString().split('T')[0],
  start_odometer: 0,
  cargo_description: '',
  notes: '',
})

// Complete Form State
const completeForm = ref({
  actual_return_date: new Date().toISOString().split('T')[0],
  end_odometer: 0,
  notes: '',
})

const provinces = [
  'Jawa Tengah',
  'Jawa Timur',
  'Jawa Barat',
  'DKI Jakarta',
  'DI Yogyakarta',
  'Banten',
  'Lampung',
  'Sumatera Selatan',
  'Sumatera Barat',
  'Bengkulu',
  'Jambi',
  'Riau',
  'Kepulauan Riau',
  'Sumatera Utara',
  'Aceh',
  'Bali',
  'Nusa Tenggara Barat',
  'Nusa Tenggara Timur',
  'Kalimantan Barat',
  'Kalimantan Tengah',
  'Kalimantan Selatan',
  'Kalimantan Timur',
  'Sulawesi Selatan',
  'Sulawesi Utara',
]

// Automatic Region Type & Daily Rate Calculation
const computedRegionType = computed(() => {
  if (!form.value.destination_province) return 'IN_PROVINCE'
  if (
    form.value.origin_province.trim().toLowerCase() ===
    form.value.destination_province.trim().toLowerCase()
  ) {
    return 'IN_PROVINCE'
  }
  return 'OUT_PROVINCE'
})

const computedDailyRate = computed(() => {
  return computedRegionType.value === 'OUT_PROVINCE' ? 150000 : 100000
})

const computedEstimatedDays = computed(() => {
  if (!form.value.departure_date || !form.value.estimated_return_date) return 1
  const start = new Date(form.value.departure_date).getTime()
  const end = new Date(form.value.estimated_return_date).getTime()
  const diffDays = Math.ceil((end - start) / (1000 * 60 * 60 * 24)) + 1
  return diffDays > 0 ? diffDays : 1
})

const computedTotalDriverAllowance = computed(() => {
  return computedEstimatedDays.value * computedDailyRate.value
})

async function loadMasterData() {
  loadingMasters.value = true
  try {
    const [vRes, eRes] = await Promise.all([
      vehicleApi.getVehicles({ all: 'true' }),
      apiClient.get('/employees', { params: { limit: 200 } }),
    ])
    vehiclesList.value = vRes.data || []
    driversList.value = eRes.data || []
  } catch (err) {
    console.error('Failed to load vehicles/employees', err)
  } finally {
    loadingMasters.value = false
  }
}

onMounted(async () => {
  await Promise.all([
    transportStore.fetchTrips(),
    loadMasterData(),
  ])
})

function formatCurrency(val: number) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

function formatDate(d: string | undefined) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function openAddModal() {
  form.value = {
    vehicle_id: vehiclesList.value[0]?.id || '',
    driver_id: driversList.value[0]?.id || '',
    helper_id: '',
    origin_city: 'Sukoharjo',
    origin_province: 'Jawa Tengah',
    destination_city: '',
    destination_province: 'Sumatera Selatan',
    region_type: 'OUT_PROVINCE',
    departure_date: new Date().toISOString().split('T')[0],
    estimated_return_date: new Date(Date.now() + 86400000 * 3).toISOString().split('T')[0],
    start_odometer: 0,
    cargo_description: '',
    notes: '',
  }
  isAddModalOpen.value = true
}

async function handleSubmitTrip() {
  try {
    await transportStore.createTrip({
      ...form.value,
      region_type: computedRegionType.value,
      estimated_days: computedEstimatedDays.value,
    })
    isAddModalOpen.value = false
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal membuat trip pengiriman')
  }
}

function openCompleteModal(trip: any) {
  selectedTrip.value = trip
  completeForm.value = {
    actual_return_date: new Date().toISOString().split('T')[0],
    end_odometer: (trip.start_odometer || 0) + 500,
    notes: '',
  }
  isCompleteModalOpen.value = true
}

async function handleCompleteTrip() {
  if (!selectedTrip.value) return
  try {
    await transportStore.completeTrip(selectedTrip.value.id, {
      actual_return_date: String(completeForm.value.actual_return_date || new Date().toISOString().slice(0, 10)),
      end_odometer: completeForm.value.end_odometer || 0,
      notes: completeForm.value.notes,
    })
    isCompleteModalOpen.value = false
    selectedTrip.value = null
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal menyelesaikan trip')
  }
}

const filteredTrips = computed(() => {
  return transportStore.trips.filter((t: any) => {
    if (filterStatus.value !== 'ALL' && t.status !== filterStatus.value) return false
    if (filterRegion.value !== 'ALL' && t.region_type !== filterRegion.value) return false
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase()
      const matchNumber = (t.trip_number || '').toLowerCase().includes(q)
      const matchDriver = (t.driver?.first_name || '').toLowerCase().includes(q)
      const matchPlate = (t.vehicle?.license_plate || '').toLowerCase().includes(q)
      const matchDest = (t.destination_city || '').toLowerCase().includes(q) || (t.destination_province || '').toLowerCase().includes(q)
      if (!matchNumber && !matchDriver && !matchPlate && !matchDest) return false
    }
    return true
  })
})
</script>

<template>
  <div class="space-y-6">
    <!-- Toolbar -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-white p-5 rounded-3xl border border-slate-200/80 shadow-sm">
      <div class="flex flex-wrap items-center gap-3 w-full sm:w-auto">
        <div class="relative flex-1 sm:w-64">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari no trip, driver, plat, tujuan..."
            class="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-primary outline-none transition-all"
          />
        </div>

        <select
          v-model="filterRegion"
          class="px-3 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white"
        >
          <option value="ALL">Semua Wilayah</option>
          <option value="IN_PROVINCE">Dalam Provinsi (Rp 100rb/hari)</option>
          <option value="OUT_PROVINCE">Luar Provinsi / Sumatera (Rp 150rb/hari)</option>
        </select>

        <select
          v-model="filterStatus"
          class="px-3 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white"
        >
          <option value="ALL">Semua Status</option>
          <option value="SCHEDULED">Terjadwal</option>
          <option value="IN_TRANSIT">Dalam Perjalanan</option>
          <option value="ARRIVED_DESTINATION">Tiba di Tujuan</option>
          <option value="COMPLETED">Selesai</option>
        </select>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <Button variant="outline" size="sm" @click="transportStore.fetchTrips()" class="rounded-2xl text-xs font-bold">
          <RefreshCw class="h-3.5 w-3.5 mr-1.5" :class="{ 'animate-spin': transportStore.loading }" />
          Refresh
        </Button>
        <Button @click="openAddModal" class="rounded-2xl text-xs font-bold shadow-md shadow-primary/20">
          <Plus class="h-4 w-4 mr-1.5" />
          Buat Penugasan Trip
        </Button>
      </div>
    </div>

    <!-- Table of Trips -->
    <div class="bg-white rounded-3xl border border-slate-200/80 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-100 bg-slate-50/50 text-[11px] font-black text-slate-400 uppercase tracking-wider">
              <th class="py-4 px-6">No Trip & Tanggal</th>
              <th class="py-4 px-6">Driver & Armada</th>
              <th class="py-4 px-6">Rute Tujuan & Wilayah</th>
              <th class="py-4 px-6">Skema Ongkos Harian</th>
              <th class="py-4 px-6">Total Uang Saku Driver</th>
              <th class="py-4 px-6">Status</th>
              <th class="py-4 px-6 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs font-bold">
            <tr v-if="transportStore.loading" class="text-center py-10">
              <td colspan="7" class="py-12 text-center text-slate-400 font-bold">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
                Memuat data trip...
              </td>
            </tr>
            <tr v-else-if="filteredTrips.length === 0">
              <td colspan="7" class="py-12 text-center text-slate-400">
                <Truck class="h-10 w-10 mx-auto mb-2 opacity-30 text-slate-400" />
                <p>Belum ada riwayat pengiriman armada.</p>
              </td>
            </tr>
            <tr
              v-for="trip in filteredTrips"
              :key="trip.id"
              class="hover:bg-slate-50/80 transition-colors"
            >
              <!-- No Trip & Tanggal -->
              <td class="py-4 px-6">
                <div class="font-mono text-slate-900 font-extrabold">{{ trip.trip_number }}</div>
                <div class="text-[11px] text-slate-400 flex items-center gap-1 mt-0.5">
                  <Calendar class="h-3 w-3" />
                  {{ formatDate(trip.departure_date) }} - {{ formatDate(trip.actual_return_date || trip.estimated_return_date) }}
                </div>
              </td>

              <!-- Driver & Armada -->
              <td class="py-4 px-6">
                <div class="text-slate-900 font-extrabold flex items-center gap-1.5">
                  <User class="h-3.5 w-3.5 text-primary" />
                  {{ trip.driver?.first_name }} {{ trip.driver?.last_name || '' }}
                </div>
                <div class="text-[11px] text-slate-500 font-mono mt-0.5 flex items-center gap-1">
                  <Truck class="h-3 w-3 text-slate-400" />
                  {{ trip.vehicle?.license_plate || 'Plat -' }} ({{ trip.vehicle?.name || '-' }})
                </div>
              </td>

              <!-- Rute Tujuan -->
              <td class="py-4 px-6">
                <div class="text-slate-800 flex items-center gap-1">
                  <MapPin class="h-3.5 w-3.5 text-red-500 shrink-0" />
                  <span class="font-black">{{ trip.destination_city || '-' }}</span>, {{ trip.destination_province }}
                </div>
                <div class="text-[10px] text-slate-400 mt-0.5">
                  Dari: {{ trip.origin_city }} ({{ trip.origin_province }})
                </div>
              </td>

              <!-- Skema Ongkos Harian -->
              <td class="py-4 px-6">
                <Badge
                  :class="trip.region_type === 'OUT_PROVINCE' ? 'bg-amber-100 text-amber-800 border-amber-200' : 'bg-blue-100 text-blue-800 border-blue-200'"
                  class="font-black text-[10px] px-2 py-0.5 rounded-full border mb-1"
                >
                  {{ trip.region_type === 'OUT_PROVINCE' ? 'Luar Provinsi / Sumatera' : 'Dalam Provinsi' }}
                </Badge>
                <div class="text-[11px] text-slate-600 font-bold">
                  {{ formatCurrency(trip.daily_allowance_rate) }} / hari ({{ trip.actual_days || trip.estimated_days }} Hari)
                </div>
              </td>

              <!-- Total Uang Saku Driver -->
              <td class="py-4 px-6">
                <div class="text-emerald-700 font-black text-sm font-mono">
                  {{ formatCurrency(trip.total_driver_fee) }}
                </div>
                <div v-if="trip.total_helper_fee > 0" class="text-[10px] text-slate-400">
                  Helper: {{ formatCurrency(trip.total_helper_fee) }}
                </div>
              </td>

              <!-- Status -->
              <td class="py-4 px-6">
                <Badge
                  :class="{
                    'bg-slate-100 text-slate-600': trip.status === 'SCHEDULED',
                    'bg-blue-100 text-blue-700 animate-pulse': trip.status === 'IN_TRANSIT',
                    'bg-purple-100 text-purple-700': trip.status === 'ARRIVED_DESTINATION',
                    'bg-emerald-100 text-emerald-800': trip.status === 'COMPLETED',
                    'bg-red-100 text-red-700': trip.status === 'CANCELLED',
                  }"
                  class="rounded-full px-2.5 py-0.5 font-black text-[10px] uppercase border-none"
                >
                  {{ trip.status }}
                </Badge>
              </td>

              <!-- Aksi -->
              <td class="py-4 px-6 text-right space-x-1">
                <Button
                  v-if="trip.status !== 'COMPLETED'"
                  variant="outline"
                  size="sm"
                  @click="openCompleteModal(trip)"
                  class="rounded-xl text-[11px] font-bold text-emerald-700 border-emerald-300 hover:bg-emerald-50 h-8"
                >
                  <CheckCircle2 class="h-3.5 w-3.5 mr-1" /> Selesaikan
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MODAL TAMBAH TRIP BARU -->
    <div
      v-if="isAddModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4 overflow-y-auto"
    >
      <div class="bg-white rounded-3xl max-w-2xl w-full p-6 shadow-2xl border border-slate-100 space-y-6 my-8">
        <div class="flex items-center justify-between border-b border-slate-100 pb-4">
          <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center font-black">
              <Truck class="h-5 w-5" />
            </div>
            <div>
              <h3 class="text-base font-black text-slate-900">Buat Penugasan Trip Baru</h3>
              <p class="text-xs text-slate-500">Kalkulasi otomatis tarif uang jalan driver (Dalam / Luar Provinsi).</p>
            </div>
          </div>
          <button @click="isAddModalOpen = false" class="p-2 rounded-xl text-slate-400 hover:bg-slate-100">
            <X class="h-5 w-5" />
          </button>
        </div>

        <form @submit.prevent="handleSubmitTrip" class="space-y-4 text-xs font-bold text-slate-700">
          <!-- Pilih Driver & Armada -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block mb-1.5 text-slate-500 uppercase tracking-wider text-[10px]">Driver Utama *</label>
              <select
                v-model="form.driver_id"
                required
                class="w-full p-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold outline-none focus:bg-white focus:border-primary"
              >
                <option value="" disabled>Pilih Driver</option>
                <option v-for="d in driversList" :key="d.id" :value="d.id">
                  {{ d.first_name }} {{ d.last_name || '' }} ({{ d.employee_id_number }})
                </option>
              </select>
            </div>

            <div>
              <label class="block mb-1.5 text-slate-500 uppercase tracking-wider text-[10px]">Armada Kendaraan *</label>
              <select
                v-model="form.vehicle_id"
                required
                class="w-full p-3 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold outline-none focus:bg-white focus:border-primary"
              >
                <option value="" disabled>Pilih Kendaraan</option>
                <option v-for="v in vehiclesList" :key="v.id" :value="v.id">
                  {{ v.license_plate }} - {{ v.name }} ({{ v.type }})
                </option>
              </select>
            </div>
          </div>

          <!-- Rute Asal & Tujuan -->
          <div class="p-4 rounded-2xl bg-slate-50 border border-slate-100 space-y-3">
            <h4 class="text-[11px] font-black text-slate-900 flex items-center gap-1.5">
              <MapPin class="h-3.5 w-3.5 text-primary" /> Detail Rute Pengiriman
            </h4>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Provinsi Asal</label>
                <select
                  v-model="form.origin_province"
                  class="w-full p-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none"
                >
                  <option v-for="p in provinces" :key="p" :value="p">{{ p }}</option>
                </select>
              </div>

              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Kota / Kab Asal</label>
                <input
                  v-model="form.origin_city"
                  type="text"
                  placeholder="Contoh: Sukoharjo"
                  class="w-full p-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none"
                />
              </div>

              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Provinsi Tujuan *</label>
                <select
                  v-model="form.destination_province"
                  required
                  class="w-full p-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none"
                >
                  <option value="" disabled>Pilih Provinsi Tujuan</option>
                  <option v-for="p in provinces" :key="p" :value="p">{{ p }}</option>
                </select>
              </div>

              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Kota / Kab Tujuan *</label>
                <input
                  v-model="form.destination_city"
                  required
                  type="text"
                  placeholder="Contoh: Palembang / Medan"
                  class="w-full p-2.5 bg-white border border-slate-200 rounded-xl text-xs font-bold outline-none"
                />
              </div>
            </div>
          </div>

          <!-- Tanggal & Durasi -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div>
              <label class="block mb-1.5 text-slate-500 uppercase tracking-wider text-[10px]">Tgl Berangkat *</label>
              <input
                v-model="form.departure_date"
                type="date"
                required
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-slate-500 uppercase tracking-wider text-[10px]">Estimasi Kembali *</label>
              <input
                v-model="form.estimated_return_date"
                type="date"
                required
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
            <div>
              <label class="block mb-1.5 text-slate-500 uppercase tracking-wider text-[10px]">KM Awal Odometer</label>
              <input
                v-model.number="form.start_odometer"
                type="number"
                placeholder="0"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none font-mono"
              />
            </div>
          </div>

          <!-- Automatic Fee Preview Card -->
          <div class="p-4 rounded-2xl bg-emerald-50 border border-emerald-200 text-emerald-900 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-[11px] font-black uppercase tracking-wider">Kalkulasi Ongkos Driver Harian</span>
              <Badge class="bg-emerald-600 text-white font-black text-[10px] px-2 py-0.5 border-none">
                {{ computedRegionType === 'OUT_PROVINCE' ? 'LUAR PROVINSI' : 'DALAM PROVINSI' }}
              </Badge>
            </div>
            <div class="grid grid-cols-3 gap-2 pt-1 border-t border-emerald-200/60 text-xs">
              <div>
                <span class="text-[10px] text-emerald-700 block">Tarif Harian:</span>
                <span class="font-extrabold font-mono">{{ formatCurrency(computedDailyRate) }} / hari</span>
              </div>
              <div>
                <span class="text-[10px] text-emerald-700 block">Estimasi Durasi:</span>
                <span class="font-extrabold font-mono">{{ computedEstimatedDays }} Hari</span>
              </div>
              <div>
                <span class="text-[10px] text-emerald-700 block">Total Uang Saku:</span>
                <span class="font-black text-emerald-800 text-sm font-mono">{{ formatCurrency(computedTotalDriverAllowance) }}</span>
              </div>
            </div>
          </div>

          <!-- Deskripsi Muatan -->
          <div>
            <label class="block mb-1.5 text-slate-500 uppercase tracking-wider text-[10px]">Deskripsi Muatan Kargo</label>
            <textarea
              v-model="form.cargo_description"
              rows="2"
              placeholder="Contoh: 500 Dus Karsa HR Clean Pengiriman Gudang Palembang"
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            ></textarea>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" @click="isAddModalOpen = false" class="rounded-xl">Batal</Button>
            <Button type="submit" class="rounded-xl bg-primary shadow-lg shadow-primary/20">Simpan & Terbitkan Trip</Button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL COMPLETE TRIP -->
    <div
      v-if="isCompleteModalOpen && selectedTrip"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4"
    >
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-slate-100 space-y-5">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <h3 class="text-base font-black text-slate-900">Selesaikan Perjalanan Trip</h3>
          <button @click="isCompleteModalOpen = false" class="p-1.5 rounded-xl text-slate-400 hover:bg-slate-100">
            <X class="h-5 w-5" />
          </button>
        </div>

        <form @submit.prevent="handleCompleteTrip" class="space-y-4 text-xs font-bold text-slate-700">
          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Tanggal Kembali Aktual *</label>
            <input
              v-model="completeForm.actual_return_date"
              type="date"
              required
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            />
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">KM Akhir Odometer *</label>
            <input
              v-model.number="completeForm.end_odometer"
              type="number"
              required
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none font-mono"
            />
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Catatan Perjalanan</label>
            <textarea
              v-model="completeForm.notes"
              rows="2"
              placeholder="Catatan kedatangan atau kendala selama di jalan..."
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            ></textarea>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" @click="isCompleteModalOpen = false" class="rounded-xl">Batal</Button>
            <Button type="submit" class="rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white shadow-lg shadow-emerald-600/20">
              Konfirmasi Selesai
            </Button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
