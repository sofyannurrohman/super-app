<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRajaTransportStore } from '@/stores/rajaTransport'
import apiClient from '@/api/axios'
import {
  Clock,
  Plus,
  Search,
  CheckCircle2,
  MapPin,
  Calendar,
  User,
  Truck,
  RefreshCw,
  X,
  Check,
  DollarSign,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const transportStore = useRajaTransportStore()

const isAddModalOpen = ref(false)
const driversList = ref<any[]>([])
const tripsList = ref<any[]>([])
const searchQuery = ref('')

const form = ref({
  trip_id: '',
  driver_id: '',
  attendance_date: new Date().toISOString().split('T')[0],
  status: 'ON_DUTY_TRIP',
  current_city: '',
  daily_allowance: 150000,
  notes: '',
})

async function loadMasterData() {
  try {
    const [eRes, tRes] = await Promise.all([
      apiClient.get('/employees', { params: { limit: 200 } }),
      transportStore.fetchTrips(),
    ])
    driversList.value = eRes.data || []
    tripsList.value = transportStore.trips || []
  } catch (err) {
    console.error('Failed to load masters', err)
  }
}

onMounted(async () => {
  await Promise.all([
    transportStore.fetchDriverAttendances(),
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

function onTripSelected() {
  const selected = tripsList.value.find((t: any) => t.id === form.value.trip_id)
  if (selected) {
    form.value.driver_id = selected.driver_id
    form.value.current_city = selected.destination_city || ''
    form.value.daily_allowance = selected.daily_allowance_rate || 100000
  }
}

function openAddModal() {
  form.value = {
    trip_id: tripsList.value[0]?.id || '',
    driver_id: tripsList.value[0]?.driver_id || driversList.value[0]?.id || '',
    attendance_date: new Date().toISOString().split('T')[0],
    status: 'ON_DUTY_TRIP',
    current_city: tripsList.value[0]?.destination_city || 'Palembang',
    daily_allowance: tripsList.value[0]?.daily_allowance_rate || 150000,
    notes: '',
  }
  isAddModalOpen.value = true
}

async function handleSubmitAttendance() {
  try {
    await transportStore.recordDriverAttendance(form.value)
    isAddModalOpen.value = false
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal mencatat absensi driver')
  }
}

const filteredAttendances = computed(() => {
  return transportStore.attendances.filter((a: any) => {
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase()
      const matchDriver = (a.driver?.first_name || '').toLowerCase().includes(q)
      const matchCity = (a.current_city || '').toLowerCase().includes(q)
      const matchTrip = (a.trip?.trip_number || '').toLowerCase().includes(q)
      if (!matchDriver && !matchCity && !matchTrip) return false
    }
    return true
  })
})
</script>

<template>
  <div class="space-y-6">
    <!-- Toolbar -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-white p-5 rounded-3xl border border-slate-200/80 shadow-sm">
      <div class="flex items-center gap-3 w-full sm:w-auto">
        <div class="relative flex-1 sm:w-72">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari driver, kota saat ini, no trip..."
            class="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-primary outline-none transition-all"
          />
        </div>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <Button variant="outline" size="sm" @click="transportStore.fetchDriverAttendances()" class="rounded-2xl text-xs font-bold">
          <RefreshCw class="h-3.5 w-3.5 mr-1.5" :class="{ 'animate-spin': transportStore.loading }" />
          Refresh
        </Button>
        <Button @click="openAddModal" class="rounded-2xl text-xs font-bold shadow-md shadow-primary/20">
          <Plus class="h-4 w-4 mr-1.5" />
          Catat Absensi Driver
        </Button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-3xl border border-slate-200/80 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-100 bg-slate-50/50 text-[11px] font-black text-slate-400 uppercase tracking-wider">
              <th class="py-4 px-6">Tanggal & Driver</th>
              <th class="py-4 px-6">No Trip & Armada</th>
              <th class="py-4 px-6">Kota / Lokasi Saat Ini</th>
              <th class="py-4 px-6">Status Kehadiran</th>
              <th class="py-4 px-6">Uang Saku Harian</th>
              <th class="py-4 px-6">Keterangan</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs font-bold">
            <tr v-if="transportStore.loading" class="text-center py-10">
              <td colspan="6" class="py-12 text-center text-slate-400 font-bold">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
                Memuat data absensi driver...
              </td>
            </tr>
            <tr v-else-if="filteredAttendances.length === 0">
              <td colspan="6" class="py-12 text-center text-slate-400">
                <Clock class="h-10 w-10 mx-auto mb-2 opacity-30 text-slate-400" />
                <p>Belum ada log absensi driver luar kota.</p>
              </td>
            </tr>
            <tr
              v-for="att in filteredAttendances"
              :key="att.id"
              class="hover:bg-slate-50/80 transition-colors"
            >
              <!-- Driver & Tgl -->
              <td class="py-4 px-6">
                <div class="text-slate-900 font-extrabold flex items-center gap-1.5">
                  <User class="h-3.5 w-3.5 text-primary" />
                  {{ att.driver?.first_name }} {{ att.driver?.last_name || '' }}
                </div>
                <div class="text-[10px] text-slate-400 mt-0.5 flex items-center gap-1">
                  <Calendar class="h-3 w-3" />
                  {{ formatDate(att.attendance_date) }}
                </div>
              </td>

              <!-- No Trip & Armada -->
              <td class="py-4 px-6">
                <div class="font-mono text-slate-900 font-bold">{{ att.trip?.trip_number || '-' }}</div>
                <div class="text-[10px] text-slate-500 font-mono mt-0.5">
                  {{ att.trip?.vehicle?.license_plate || '-' }}
                </div>
              </td>

              <!-- Kota Saat Ini -->
              <td class="py-4 px-6">
                <div class="text-slate-800 flex items-center gap-1">
                  <MapPin class="h-3.5 w-3.5 text-red-500 shrink-0" />
                  <span class="font-extrabold">{{ att.current_city || 'Dalam Perjalanan' }}</span>
                </div>
              </td>

              <!-- Status -->
              <td class="py-4 px-6">
                <Badge
                  :class="{
                    'bg-emerald-100 text-emerald-800 border-emerald-200': att.status === 'ON_DUTY_TRIP',
                    'bg-blue-100 text-blue-800 border-blue-200': att.status === 'REST',
                    'bg-purple-100 text-purple-800 border-purple-200': att.status === 'STANDBY',
                  }"
                  class="font-black text-[10px] px-2.5 py-0.5 rounded-full border uppercase"
                >
                  {{ att.status.replace(/_/g, ' ') }}
                </Badge>
              </td>

              <!-- Uang Saku Harian -->
              <td class="py-4 px-6">
                <div class="text-emerald-700 font-mono font-black">
                  {{ formatCurrency(att.daily_allowance) }}
                </div>
              </td>

              <!-- Keterangan -->
              <td class="py-4 px-6 text-slate-500 max-w-xs truncate">
                {{ att.notes || '-' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MODAL CATAT ABSENSI -->
    <div
      v-if="isAddModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4"
    >
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-slate-100 space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center font-black">
              <Clock class="h-5 w-5" />
            </div>
            <div>
              <h3 class="text-base font-black text-slate-900">Catat Kehadiran Driver Perjalanan</h3>
              <p class="text-xs text-slate-500">Monitoring tugas & uang saku harian.</p>
            </div>
          </div>
          <button @click="isAddModalOpen = false" class="p-1.5 rounded-xl text-slate-400 hover:bg-slate-100">
            <X class="h-5 w-5" />
          </button>
        </div>

        <form @submit.prevent="handleSubmitAttendance" class="space-y-3 text-xs font-bold text-slate-700">
          <div>
            <label class="block mb-1 text-[10px] text-slate-500 uppercase tracking-wider">Pilih Trip Pengiriman *</label>
            <select
              v-model="form.trip_id"
              @change="onTripSelected"
              required
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            >
              <option value="" disabled>Pilih Trip Aktif</option>
              <option v-for="t in tripsList" :key="t.id" :value="t.id">
                {{ t.trip_number }} - {{ t.driver?.first_name }} ({{ t.destination_city }})
              </option>
            </select>
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500 uppercase tracking-wider">Driver *</label>
            <select
              v-model="form.driver_id"
              required
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            >
              <option v-for="d in driversList" :key="d.id" :value="d.id">
                {{ d.first_name }} {{ d.last_name || '' }}
              </option>
            </select>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Tanggal Kehadiran *</label>
              <input
                v-model="form.attendance_date"
                type="date"
                required
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>

            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Status Tugas</label>
              <select
                v-model="form.status"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              >
                <option value="ON_DUTY_TRIP">Bertugas Kirim</option>
                <option value="REST">Istirahat / Nginap</option>
                <option value="STANDBY">Standby Bongkar Muat</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Kota Saat Ini *</label>
              <input
                v-model="form.current_city"
                type="text"
                required
                placeholder="Contoh: Palembang"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>

            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Uang Saku (Rp)</label>
              <input
                v-model.number="form.daily_allowance"
                type="number"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none font-mono"
              />
            </div>
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Catatan / Keterangan</label>
            <textarea
              v-model="form.notes"
              rows="2"
              placeholder="Catatan aktivitas hari ini..."
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            ></textarea>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" @click="isAddModalOpen = false" class="rounded-xl">Batal</Button>
            <Button type="submit" class="rounded-xl bg-amber-600 hover:bg-amber-700 text-white font-bold">
              Simpan Kehadiran
            </Button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
