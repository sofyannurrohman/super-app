<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRajaTransportStore } from '@/stores/rajaTransport'
import { vehicleApi } from '@/api/vehicle'
import apiClient from '@/api/axios'
import {
  Wrench,
  Plus,
  Search,
  CheckCircle2,
  AlertTriangle,
  Clock,
  ShieldCheck,
  Truck,
  User,
  Eye,
  RefreshCw,
  X,
  Check,
  DollarSign,
  Building2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const transportStore = useRajaTransportStore()

const isAddModalOpen = ref(false)
const isFollowUpModalOpen = ref(false)
const selectedInspection = ref<any | null>(null)

const vehiclesList = ref<any[]>([])
const driversList = ref<any[]>([])
const tripsList = ref<any[]>([])

const filterAction = ref('ALL')
const filterStatus = ref('ALL')
const searchQuery = ref('')
const selectedFile = ref<File | null>(null)

const form = ref({
  trip_id: '',
  vehicle_id: '',
  inspector_driver_id: '',
  inspection_date: new Date().toISOString().split('T')[0],
  return_odometer: 0,
  brake_condition: 'GOOD',
  tire_condition: 'GOOD',
  engine_condition: 'GOOD',
  oil_condition: 'GOOD',
  battery_condition: 'GOOD',
  electrical_lights: 'GOOD',
  body_condition: 'GOOD',
  driver_remarks: '',
  follow_up_action: 'NO_ACTION_REQUIRED',
  parts_to_replace: '',
  workshop_name: '',
  estimated_cost: 0,
})

const followUpForm = ref({
  workshop_name: '',
  parts_replaced: '',
  actual_cost: 0,
  follow_up_status: 'COMPLETED' as 'PENDING_REPAIR' | 'IN_PROGRESS' | 'COMPLETED',
  repair_notes: '',
})

async function loadMasterData() {
  try {
    const [vRes, eRes] = await Promise.all([
      vehicleApi.getVehicles({ all: 'true' }),
      apiClient.get('/employees', { params: { limit: 200 } }),
    ])
    vehiclesList.value = vRes.data || []
    driversList.value = eRes.data || []
    tripsList.value = transportStore.trips || []
  } catch (err) {
    console.error('Failed to load masters', err)
  }
}

onMounted(async () => {
  await Promise.all([
    transportStore.fetchFleetInspections(),
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

function handleFileChange(e: any) {
  const file = e.target.files[0]
  if (file) {
    selectedFile.value = file
  }
}

function openAddModal() {
  form.value = {
    trip_id: '',
    vehicle_id: vehiclesList.value[0]?.id || '',
    inspector_driver_id: driversList.value[0]?.id || '',
    inspection_date: new Date().toISOString().split('T')[0],
    return_odometer: 0,
    brake_condition: 'GOOD',
    tire_condition: 'GOOD',
    engine_condition: 'GOOD',
    oil_condition: 'GOOD',
    battery_condition: 'GOOD',
    electrical_lights: 'GOOD',
    body_condition: 'GOOD',
    driver_remarks: '',
    follow_up_action: 'NO_ACTION_REQUIRED',
    parts_to_replace: '',
    workshop_name: '',
    estimated_cost: 0,
  }
  selectedFile.value = null
  isAddModalOpen.value = true
}

async function handleSubmitInspection() {
  try {
    const formData = new FormData()
    formData.append('vehicle_id', form.value.vehicle_id)
    formData.append('inspector_driver_id', form.value.inspector_driver_id)
    if (form.value.trip_id) formData.append('trip_id', form.value.trip_id)
    formData.append('inspection_date', String(form.value.inspection_date || new Date().toISOString().slice(0, 10)))
    formData.append('return_odometer', String(form.value.return_odometer))
    formData.append('brake_condition', form.value.brake_condition)
    formData.append('tire_condition', form.value.tire_condition)
    formData.append('engine_condition', form.value.engine_condition)
    formData.append('oil_condition', form.value.oil_condition)
    formData.append('battery_condition', form.value.battery_condition)
    formData.append('electrical_lights', form.value.electrical_lights)
    formData.append('body_condition', form.value.body_condition)
    formData.append('driver_remarks', form.value.driver_remarks)
    formData.append('follow_up_action', form.value.follow_up_action)
    formData.append('parts_to_replace', form.value.parts_to_replace)
    formData.append('workshop_name', form.value.workshop_name)
    formData.append('estimated_cost', String(form.value.estimated_cost))
    if (selectedFile.value) {
      formData.append('photo_condition', selectedFile.value)
    }

    await transportStore.createFleetInspection(formData)
    isAddModalOpen.value = false
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal menyimpan laporan inspeksi')
  }
}

function openFollowUpModal(ins: any) {
  selectedInspection.value = ins
  followUpForm.value = {
    workshop_name: ins.workshop_name || '',
    parts_replaced: ins.parts_to_replace || '',
    actual_cost: ins.actual_cost || ins.estimated_cost || 0,
    follow_up_status: ins.follow_up_status || 'COMPLETED',
    repair_notes: ins.repair_notes || '',
  }
  isFollowUpModalOpen.value = true
}

async function handleFollowUpSubmit() {
  if (!selectedInspection.value) return
  try {
    await transportStore.updateInspectionFollowUp(selectedInspection.value.id, followUpForm.value)
    isFollowUpModalOpen.value = false
    selectedInspection.value = null
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal mengupdate tindak lanjut')
  }
}

const filteredInspections = computed(() => {
  return transportStore.inspections.filter((ins: any) => {
    if (filterAction.value !== 'ALL' && ins.follow_up_action !== filterAction.value) return false
    if (filterStatus.value !== 'ALL' && ins.follow_up_status !== filterStatus.value) return false
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase()
      const matchPlate = (ins.vehicle?.license_plate || '').toLowerCase().includes(q)
      const matchDriver = (ins.driver?.first_name || '').toLowerCase().includes(q)
      const matchParts = (ins.parts_to_replace || '').toLowerCase().includes(q)
      if (!matchPlate && !matchDriver && !matchParts) return false
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
            placeholder="Cari plat armada, driver, part..."
            class="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-primary outline-none transition-all"
          />
        </div>

        <select
          v-model="filterAction"
          class="px-3 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white"
        >
          <option value="ALL">Semua Tindak Lanjut</option>
          <option value="NO_ACTION_REQUIRED">Kondisi Prima Siap Jalan</option>
          <option value="NEEDS_MINOR_REPAIR">Perlu Servis Ringan</option>
          <option value="NEEDS_PART_REPLACEMENT">Perlu Ganti Sparepart</option>
          <option value="NEEDS_OVERHAUL">Perlu Perbaikan Besar</option>
        </select>

        <select
          v-model="filterStatus"
          class="px-3 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white"
        >
          <option value="ALL">Semua Status Perbaikan</option>
          <option value="PENDING_REPAIR">Menunggu Perbaikan</option>
          <option value="IN_PROGRESS">Sedang Diperbaiki</option>
          <option value="COMPLETED">Selesai Diperbaiki</option>
        </select>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <Button variant="outline" size="sm" @click="transportStore.fetchFleetInspections()" class="rounded-2xl text-xs font-bold">
          <RefreshCw class="h-3.5 w-3.5 mr-1.5" :class="{ 'animate-spin': transportStore.loading }" />
          Refresh
        </Button>
        <Button @click="openAddModal" class="rounded-2xl text-xs font-bold shadow-md shadow-primary/20">
          <Plus class="h-4 w-4 mr-1.5" />
          Lapor Kondisi Armada
        </Button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-3xl border border-slate-200/80 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-100 bg-slate-50/50 text-[11px] font-black text-slate-400 uppercase tracking-wider">
              <th class="py-4 px-6">Tgl & Armada</th>
              <th class="py-4 px-6">Driver Pelapor & KM</th>
              <th class="py-4 px-6">Checklist Kondisi Utama</th>
              <th class="py-4 px-6">Keluhan / Catatan Driver</th>
              <th class="py-4 px-6">Tindak Lanjut & Sparepart</th>
              <th class="py-4 px-6">Status Perbaikan</th>
              <th class="py-4 px-6 text-right">Aksi Tindak Lanjut</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs font-bold">
            <tr v-if="transportStore.loading" class="text-center py-10">
              <td colspan="7" class="py-12 text-center text-slate-400 font-bold">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
                Memuat data inspeksi armada...
              </td>
            </tr>
            <tr v-else-if="filteredInspections.length === 0">
              <td colspan="7" class="py-12 text-center text-slate-400">
                <Wrench class="h-10 w-10 mx-auto mb-2 opacity-30 text-slate-400" />
                <p>Belum ada laporan kondisi armada.</p>
              </td>
            </tr>
            <tr
              v-for="ins in filteredInspections"
              :key="ins.id"
              class="hover:bg-slate-50/80 transition-colors"
            >
              <!-- Tgl & Armada -->
              <td class="py-4 px-6">
                <div class="font-mono text-slate-900 font-extrabold">{{ ins.vehicle?.license_plate || '-' }}</div>
                <div class="text-[11px] text-slate-500 font-bold">{{ ins.vehicle?.name }}</div>
                <div class="text-[10px] text-slate-400 mt-0.5">{{ formatDate(ins.inspection_date) }}</div>
              </td>

              <!-- Driver & KM -->
              <td class="py-4 px-6">
                <div class="text-slate-800 font-bold flex items-center gap-1">
                  <User class="h-3.5 w-3.5 text-primary" />
                  {{ ins.driver?.first_name }} {{ ins.driver?.last_name || '' }}
                </div>
                <div class="text-[11px] text-slate-500 font-mono mt-0.5">
                  KM: {{ ins.return_odometer?.toLocaleString() || '-' }}
                </div>
              </td>

              <!-- Checklist Kondisi -->
              <td class="py-4 px-6">
                <div class="flex flex-wrap gap-1 max-w-xs">
                  <span
                    :class="ins.brake_condition === 'GOOD' ? 'bg-slate-100 text-slate-600' : 'bg-red-100 text-red-700 font-black'"
                    class="text-[9px] px-1.5 py-0.5 rounded"
                  >Rem: {{ ins.brake_condition }}</span>
                  <span
                    :class="ins.tire_condition === 'GOOD' ? 'bg-slate-100 text-slate-600' : 'bg-amber-100 text-amber-800 font-black'"
                    class="text-[9px] px-1.5 py-0.5 rounded"
                  >Ban: {{ ins.tire_condition }}</span>
                  <span
                    :class="ins.engine_condition === 'GOOD' ? 'bg-slate-100 text-slate-600' : 'bg-red-100 text-red-700 font-black'"
                    class="text-[9px] px-1.5 py-0.5 rounded"
                  >Mesin: {{ ins.engine_condition }}</span>
                  <span
                    :class="ins.oil_condition === 'GOOD' ? 'bg-slate-100 text-slate-600' : 'bg-amber-100 text-amber-800 font-black'"
                    class="text-[9px] px-1.5 py-0.5 rounded"
                  >Oli: {{ ins.oil_condition }}</span>
                </div>
              </td>

              <!-- Keluhan Driver -->
              <td class="py-4 px-6 max-w-xs">
                <p class="text-slate-700 truncate" :title="ins.driver_remarks">
                  {{ ins.driver_remarks || 'Tidak ada keluhan.' }}
                </p>
              </td>

              <!-- Tindak Lanjut & Part -->
              <td class="py-4 px-6">
                <Badge
                  :class="{
                    'bg-emerald-100 text-emerald-800 border-emerald-200': ins.follow_up_action === 'NO_ACTION_REQUIRED',
                    'bg-blue-100 text-blue-800 border-blue-200': ins.follow_up_action === 'NEEDS_MINOR_REPAIR',
                    'bg-amber-100 text-amber-800 border-amber-200': ins.follow_up_action === 'NEEDS_PART_REPLACEMENT',
                    'bg-red-100 text-red-800 border-red-200': ins.follow_up_action === 'NEEDS_OVERHAUL',
                  }"
                  class="font-black text-[9px] px-2 py-0.5 rounded-md border uppercase mb-1"
                >
                  {{ ins.follow_up_action.replace(/_/g, ' ') }}
                </Badge>
                <div v-if="ins.parts_to_replace" class="text-[11px] text-amber-900 font-bold">
                  Part: {{ ins.parts_to_replace }}
                </div>
                <div v-if="ins.workshop_name" class="text-[10px] text-slate-500">
                  Bengkel: {{ ins.workshop_name }}
                </div>
              </td>

              <!-- Status Perbaikan -->
              <td class="py-4 px-6">
                <Badge
                  :class="{
                    'bg-amber-100 text-amber-800 border-amber-200': ins.follow_up_status === 'PENDING_REPAIR',
                    'bg-blue-100 text-blue-800 border-blue-200': ins.follow_up_status === 'IN_PROGRESS',
                    'bg-emerald-100 text-emerald-800 border-emerald-200': ins.follow_up_status === 'COMPLETED',
                  }"
                  class="rounded-full px-2.5 py-0.5 font-black text-[10px] uppercase border"
                >
                  {{ ins.follow_up_status.replace(/_/g, ' ') }}
                </Badge>
                <div v-if="ins.actual_cost > 0" class="text-[10px] text-slate-500 font-mono mt-0.5">
                  Biaya: {{ formatCurrency(ins.actual_cost) }}
                </div>
              </td>

              <!-- Aksi Tindak Lanjut -->
              <td class="py-4 px-6 text-right space-x-1">
                <Button
                  v-if="ins.follow_up_action !== 'NO_ACTION_REQUIRED' && ins.follow_up_status !== 'COMPLETED'"
                  size="sm"
                  @click="openFollowUpModal(ins)"
                  class="rounded-xl text-[11px] font-bold bg-amber-600 hover:bg-amber-700 text-white h-7 px-2.5 shadow-sm"
                >
                  <Wrench class="h-3.5 w-3.5 mr-1" /> Update Servis
                </Button>
                <span v-else class="text-[11px] text-emerald-700 font-bold">✓ Prima / Selesai</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MODAL INPUT INSPEKSI -->
    <div
      v-if="isAddModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4 overflow-y-auto"
    >
      <div class="bg-white rounded-3xl max-w-2xl w-full p-6 shadow-2xl border border-slate-100 space-y-5 my-8">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center font-black">
              <Wrench class="h-5 w-5" />
            </div>
            <div>
              <h3 class="text-base font-black text-slate-900">Laporan Kondisi Armada Pasca-Perjalanan</h3>
              <p class="text-xs text-slate-500">Inspeksi fisik dan rencana perbaikan / penggantian sparepart.</p>
            </div>
          </div>
          <button @click="isAddModalOpen = false" class="p-1.5 rounded-xl text-slate-400 hover:bg-slate-100">
            <X class="h-5 w-5" />
          </button>
        </div>

        <form @submit.prevent="handleSubmitInspection" class="space-y-4 text-xs font-bold text-slate-700">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div>
              <label class="block mb-1 text-[10px] text-slate-500 uppercase tracking-wider">Armada *</label>
              <select
                v-model="form.vehicle_id"
                required
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              >
                <option value="" disabled>Pilih Kendaraan</option>
                <option v-for="v in vehiclesList" :key="v.id" :value="v.id">
                  {{ v.license_plate }} - {{ v.name }}
                </option>
              </select>
            </div>

            <div>
              <label class="block mb-1 text-[10px] text-slate-500 uppercase tracking-wider">Driver Inspektor *</label>
              <select
                v-model="form.inspector_driver_id"
                required
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              >
                <option value="" disabled>Pilih Driver</option>
                <option v-for="d in driversList" :key="d.id" :value="d.id">
                  {{ d.first_name }} {{ d.last_name || '' }}
                </option>
              </select>
            </div>

            <div>
              <label class="block mb-1 text-[10px] text-slate-500 uppercase tracking-wider">KM Odometer Pulang *</label>
              <input
                v-model.number="form.return_odometer"
                type="number"
                required
                placeholder="0"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none font-mono"
              />
            </div>
          </div>

          <!-- Checklist Grid -->
          <div class="p-4 rounded-2xl bg-slate-50 border border-slate-100 space-y-3">
            <h4 class="text-[11px] font-black text-slate-900 flex items-center gap-1.5">
              <ShieldCheck class="h-3.5 w-3.5 text-primary" /> Checklist Kondisi Komponen Utama
            </h4>

            <div class="grid grid-cols-2 md:grid-cols-4 gap-3 text-[11px]">
              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Pengereman</label>
                <select v-model="form.brake_condition" class="w-full p-2 bg-white border border-slate-200 rounded-lg text-xs font-bold">
                  <option value="GOOD">Normal / Baik</option>
                  <option value="FAIR">Mulai Tipis</option>
                  <option value="DAMAGED">Rusak / Aus</option>
                  <option value="CRITICAL">Bahaya / Blong</option>
                </select>
              </div>

              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Kondisi Ban</label>
                <select v-model="form.tire_condition" class="w-full p-2 bg-white border border-slate-200 rounded-lg text-xs font-bold">
                  <option value="GOOD">Tebal / Baik</option>
                  <option value="FAIR">Mulai Botak</option>
                  <option value="DAMAGED">Aus / Bocor</option>
                  <option value="CRITICAL">Wajib Ganti</option>
                </select>
              </div>

              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Kondisi Mesin</label>
                <select v-model="form.engine_condition" class="w-full p-2 bg-white border border-slate-200 rounded-lg text-xs font-bold">
                  <option value="GOOD">Halus / Normal</option>
                  <option value="FAIR">Ada Bunyi Ringan</option>
                  <option value="DAMAGED">Mogok / Panas</option>
                </select>
              </div>

              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Kondisi Oli</label>
                <select v-model="form.oil_condition" class="w-full p-2 bg-white border border-slate-200 rounded-lg text-xs font-bold">
                  <option value="GOOD">Cukup & Bersih</option>
                  <option value="FAIR">Waktunya Ganti</option>
                  <option value="DAMAGED">Bocor / Kering</option>
                </select>
              </div>

              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Aki / Baterai</label>
                <select v-model="form.battery_condition" class="w-full p-2 bg-white border border-slate-200 rounded-lg text-xs font-bold">
                  <option value="GOOD">Starter Lancar</option>
                  <option value="FAIR">Mulai Lemah</option>
                  <option value="DAMAGED">Soak / Mati</option>
                </select>
              </div>

              <div>
                <label class="block mb-1 text-[10px] text-slate-500">Lampu & Kelistrikan</label>
                <select v-model="form.electrical_lights" class="w-full p-2 bg-white border border-slate-200 rounded-lg text-xs font-bold">
                  <option value="GOOD">Menyala Semua</option>
                  <option value="FAIR">Ada Redup</option>
                  <option value="DAMAGED">Putus / Mati</option>
                </select>
              </div>

              <div class="col-span-2">
                <label class="block mb-1 text-[10px] text-slate-500">Kondisi Bodi & Bak Truk</label>
                <select v-model="form.body_condition" class="w-full p-2 bg-white border border-slate-200 rounded-lg text-xs font-bold">
                  <option value="GOOD">Mulus & Rapi</option>
                  <option value="FAIR">Baret Wajar</option>
                  <option value="DAMAGED">Penyok / Rusak</option>
                </select>
              </div>
            </div>
          </div>

          <!-- Tindak Lanjut -->
          <div class="p-4 rounded-2xl bg-amber-50/70 border border-amber-200 space-y-3">
            <h4 class="text-[11px] font-black text-amber-900 flex items-center gap-1.5">
              <AlertTriangle class="h-3.5 w-3.5 text-amber-600" /> Tindak Lanjut Armada
            </h4>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label class="block mb-1 text-[10px] text-slate-600">Tindakan Selanjutnya *</label>
                <select v-model="form.follow_up_action" class="w-full p-2.5 bg-white border border-amber-200 rounded-xl text-xs font-bold">
                  <option value="NO_ACTION_REQUIRED">Kondisi Prima Siap Jalan</option>
                  <option value="NEEDS_MINOR_REPAIR">Perlu Servis Ringan / Tune Up</option>
                  <option value="NEEDS_PART_REPLACEMENT">Perlu Penggantian Sparepart</option>
                  <option value="NEEDS_OVERHAUL">Perlu Perbaikan Besar / Turun Mesin</option>
                </select>
              </div>

              <div v-if="form.follow_up_action !== 'NO_ACTION_REQUIRED'">
                <label class="block mb-1 text-[10px] text-slate-600">Part Yang Perlu Diganti</label>
                <input
                  v-model="form.parts_to_replace"
                  type="text"
                  placeholder="Contoh: Kampas rem belakang, Oli mesin 5L"
                  class="w-full p-2.5 bg-white border border-amber-200 rounded-xl text-xs font-bold"
                />
              </div>

              <div v-if="form.follow_up_action !== 'NO_ACTION_REQUIRED'">
                <label class="block mb-1 text-[10px] text-slate-600">Bengkel Rujukan</label>
                <input
                  v-model="form.workshop_name"
                  type="text"
                  placeholder="Contoh: Bengkel Armada Mandiri"
                  class="w-full p-2.5 bg-white border border-amber-200 rounded-xl text-xs font-bold"
                />
              </div>

              <div v-if="form.follow_up_action !== 'NO_ACTION_REQUIRED'">
                <label class="block mb-1 text-[10px] text-slate-600">Estimasi Biaya (Rp)</label>
                <input
                  v-model.number="form.estimated_cost"
                  type="number"
                  placeholder="0"
                  class="w-full p-2.5 bg-white border border-amber-200 rounded-xl text-xs font-bold font-mono"
                />
              </div>
            </div>
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Catatan Tambahan Keluhan Driver</label>
            <textarea
              v-model="form.driver_remarks"
              rows="2"
              placeholder="Catatan getaran, temperatur, atau kendala di jalan..."
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            ></textarea>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" @click="isAddModalOpen = false" class="rounded-xl">Batal</Button>
            <Button type="submit" class="rounded-xl bg-amber-600 hover:bg-amber-700 text-white shadow-lg shadow-amber-600/20">
              Simpan Laporan Inspeksi
            </Button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL UPDATE SERVIS -->
    <div
      v-if="isFollowUpModalOpen && selectedInspection"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4"
    >
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-slate-100 space-y-4">
        <h3 class="text-base font-black text-slate-900">
          Update Tindak Lanjut Servis Armada
        </h3>
        <p class="text-xs text-slate-500 font-mono">
          {{ selectedInspection.vehicle?.license_plate }} ({{ selectedInspection.vehicle?.name }})
        </p>

        <form @submit.prevent="handleFollowUpSubmit" class="space-y-3 text-xs font-bold text-slate-700">
          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Nama Bengkel / Teknisi</label>
            <input
              v-model="followUpForm.workshop_name"
              type="text"
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            />
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Part Yang Diganti</label>
            <input
              v-model="followUpForm.parts_replaced"
              type="text"
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            />
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Realisasi Biaya Perbaikan (Rp)</label>
            <input
              v-model.number="followUpForm.actual_cost"
              type="number"
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold font-mono outline-none"
            />
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Status Perbaikan *</label>
            <select
              v-model="followUpForm.follow_up_status"
              required
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            >
              <option value="PENDING_REPAIR">Menunggu Perbaikan</option>
              <option value="IN_PROGRESS">Sedang Dikerjakan Bengkel</option>
              <option value="COMPLETED">Selesai Diperbaiki (Siap Jalan)</option>
            </select>
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Catatan Bengkel</label>
            <textarea
              v-model="followUpForm.repair_notes"
              rows="2"
              placeholder="Catatan hasil servis..."
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            ></textarea>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" @click="isFollowUpModalOpen = false" class="rounded-xl">Batal</Button>
            <Button type="submit" class="rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold">
              Simpan & Selesaikan
            </Button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
