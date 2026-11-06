<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRajaTransportStore } from '@/stores/rajaTransport'
import { vehicleApi } from '@/api/vehicle'
import apiClient from '@/api/axios'
import {
  RotateCcw,
  Plus,
  Search,
  CheckCircle2,
  XCircle,
  Clock,
  MapPin,
  Calendar,
  DollarSign,
  FileText,
  User,
  Truck,
  Eye,
  RefreshCw,
  X,
  Check,
  Percent,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const transportStore = useRajaTransportStore()

const isAddModalOpen = ref(false)
const isApprovalModalOpen = ref(false)
const isPreviewProofModalOpen = ref(false)
const previewImageURL = ref('')
const selectedBackhaul = ref<any | null>(null)
const approvalAction = ref<'APPROVED' | 'REJECTED'>('APPROVED')
const approvalNote = ref('')

const vehiclesList = ref<any[]>([])
const driversList = ref<any[]>([])
const tripsList = ref<any[]>([])

const filterStatus = ref('ALL')
const searchQuery = ref('')
const selectedFile = ref<File | null>(null)

const form = ref({
  trip_id: '',
  driver_id: '',
  vehicle_id: '',
  origin_location: '',
  destination_location: '',
  customer_name: '',
  customer_phone: '',
  cargo_name: '',
  cargo_weight_kg: 0,
  freight_cost: 0,
  fee_percentage: 10,
  notes: '',
})

// Auto calculate fee 10%
const computedFeeAmount = computed(() => {
  return (form.value.fee_percentage / 100) * (form.value.freight_cost || 0)
})

const computedNetAmount = computed(() => {
  return (form.value.freight_cost || 0) - computedFeeAmount.value
})

async function loadMasterData() {
  try {
    const [vRes, eRes, tRes] = await Promise.all([
      vehicleApi.getVehicles({ all: 'true' }),
      apiClient.get('/employees', { params: { limit: 200 } }),
      transportStore.fetchTrips(),
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
    transportStore.fetchBackhauls(),
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
    driver_id: driversList.value[0]?.id || '',
    vehicle_id: vehiclesList.value[0]?.id || '',
    origin_location: 'Palembang',
    destination_location: 'Sukoharjo',
    customer_name: '',
    customer_phone: '',
    cargo_name: '',
    cargo_weight_kg: 0,
    freight_cost: 0,
    fee_percentage: 10,
    notes: '',
  }
  selectedFile.value = null
  isAddModalOpen.value = true
}

async function handleSubmitBackhaul() {
  try {
    const formData = new FormData()
    formData.append('driver_id', form.value.driver_id)
    formData.append('vehicle_id', form.value.vehicle_id)
    if (form.value.trip_id) formData.append('trip_id', form.value.trip_id)
    formData.append('origin_location', form.value.origin_location)
    formData.append('destination_location', form.value.destination_location)
    formData.append('customer_name', form.value.customer_name)
    formData.append('customer_phone', form.value.customer_phone)
    formData.append('cargo_name', form.value.cargo_name)
    formData.append('cargo_weight_kg', String(form.value.cargo_weight_kg))
    formData.append('freight_cost', String(form.value.freight_cost))
    formData.append('fee_percentage', String(form.value.fee_percentage))
    formData.append('notes', form.value.notes)
    if (selectedFile.value) {
      formData.append('attachment_proof', selectedFile.value)
    }

    await transportStore.createBackhaul(formData)
    isAddModalOpen.value = false
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal menyimpan muatan balik')
  }
}

function openApprovalModal(item: any, action: 'APPROVED' | 'REJECTED') {
  selectedBackhaul.value = item
  approvalAction.value = action
  approvalNote.value = ''
  isApprovalModalOpen.value = true
}

async function handleApprovalSubmit() {
  if (!selectedBackhaul.value) return
  try {
    await transportStore.approveBackhaul(selectedBackhaul.value.id, {
      action: approvalAction.value,
      notes: approvalAction.value === 'APPROVED' ? approvalNote.value : undefined,
      reason: approvalAction.value === 'REJECTED' ? approvalNote.value : undefined,
    })
    isApprovalModalOpen.value = false
    selectedBackhaul.value = null
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal memproses persetujuan')
  }
}

function openPreviewProof(url: string) {
  previewImageURL.value = url
  isPreviewProofModalOpen.value = true
}

const filteredBackhauls = computed(() => {
  return transportStore.backhauls.filter((b: any) => {
    if (filterStatus.value !== 'ALL' && b.status !== filterStatus.value) return false
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase()
      const matchNum = (b.backhaul_number || '').toLowerCase().includes(q)
      const matchCust = (b.customer_name || '').toLowerCase().includes(q)
      const matchCargo = (b.cargo_name || '').toLowerCase().includes(q)
      const matchDriver = (b.driver?.first_name || '').toLowerCase().includes(q)
      if (!matchNum && !matchCust && !matchCargo && !matchDriver) return false
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
            placeholder="Cari no balikan, customer, kargo..."
            class="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-primary outline-none transition-all"
          />
        </div>

        <select
          v-model="filterStatus"
          class="px-3 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white"
        >
          <option value="ALL">Semua Status Approval</option>
          <option value="PENDING">Menunggu Persetujuan (Pending)</option>
          <option value="APPROVED">Disetujui (Approved)</option>
          <option value="REJECTED">Ditolak (Rejected)</option>
        </select>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <Button variant="outline" size="sm" @click="transportStore.fetchBackhauls()" class="rounded-2xl text-xs font-bold">
          <RefreshCw class="h-3.5 w-3.5 mr-1.5" :class="{ 'animate-spin': transportStore.loading }" />
          Refresh
        </Button>
        <Button @click="openAddModal" class="rounded-2xl text-xs font-bold shadow-md shadow-primary/20">
          <Plus class="h-4 w-4 mr-1.5" />
          Input Muatan Balik
        </Button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-3xl border border-slate-200/80 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-100 bg-slate-50/50 text-[11px] font-black text-slate-400 uppercase tracking-wider">
              <th class="py-4 px-6">No Balikan & Tgl</th>
              <th class="py-4 px-6">Rute & Customer</th>
              <th class="py-4 px-6">Driver & Armada</th>
              <th class="py-4 px-6">Muatan & Bobot</th>
              <th class="py-4 px-6">Total Ongkos & Fee 10%</th>
              <th class="py-4 px-6">Bukti Nota/SJ</th>
              <th class="py-4 px-6">Status Approval</th>
              <th class="py-4 px-6 text-right">Aksi Persetujuan</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs font-bold">
            <tr v-if="transportStore.loading" class="text-center py-10">
              <td colspan="8" class="py-12 text-center text-slate-400 font-bold">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
                Memuat data pengiriman kembali...
              </td>
            </tr>
            <tr v-else-if="filteredBackhauls.length === 0">
              <td colspan="8" class="py-12 text-center text-slate-400">
                <RotateCcw class="h-10 w-10 mx-auto mb-2 opacity-30 text-slate-400" />
                <p>Belum ada data pengiriman muatan kembali.</p>
              </td>
            </tr>
            <tr
              v-for="b in filteredBackhauls"
              :key="b.id"
              class="hover:bg-slate-50/80 transition-colors"
            >
              <!-- No Balikan -->
              <td class="py-4 px-6">
                <div class="font-mono text-slate-900 font-extrabold">{{ b.backhaul_number }}</div>
                <div class="text-[10px] text-slate-400">{{ formatDate(b.created_at) }}</div>
                <div v-if="b.trip?.trip_number" class="text-[10px] text-primary font-mono mt-0.5">
                  Trip: {{ b.trip.trip_number }}
                </div>
              </td>

              <!-- Rute & Customer -->
              <td class="py-4 px-6">
                <div class="text-slate-900 font-extrabold">{{ b.customer_name }}</div>
                <div class="text-[11px] text-slate-500 flex items-center gap-1 mt-0.5">
                  <MapPin class="h-3 w-3 text-red-500 shrink-0" />
                  <span>{{ b.origin_location }}</span> → <span>{{ b.destination_location }}</span>
                </div>
                <div v-if="b.customer_phone" class="text-[10px] text-slate-400">Tel: {{ b.customer_phone }}</div>
              </td>

              <!-- Driver & Armada -->
              <td class="py-4 px-6">
                <div class="text-slate-800 font-bold flex items-center gap-1">
                  <User class="h-3.5 w-3.5 text-primary" />
                  {{ b.driver?.first_name }} {{ b.driver?.last_name || '' }}
                </div>
                <div class="text-[11px] text-slate-500 font-mono mt-0.5">
                  {{ b.vehicle?.license_plate || '-' }}
                </div>
              </td>

              <!-- Muatan & Bobot -->
              <td class="py-4 px-6">
                <div class="text-slate-900 font-black">{{ b.cargo_name }}</div>
                <div v-if="b.cargo_weight_kg > 0" class="text-[11px] text-slate-500">
                  {{ b.cargo_weight_kg.toLocaleString() }} Kg
                </div>
              </td>

              <!-- Total Ongkos & Fee 10% -->
              <td class="py-4 px-6">
                <div class="font-extrabold text-slate-900 font-mono">
                  {{ formatCurrency(b.freight_cost) }}
                </div>
                <div class="text-[11px] text-amber-700 font-black flex items-center gap-1 font-mono">
                  <span>Fee 10%:</span>
                  <span>{{ formatCurrency(b.fee_amount) }}</span>
                </div>
                <div class="text-[10px] text-slate-400 font-mono">
                  Bersih: {{ formatCurrency(b.net_amount) }}
                </div>
              </td>

              <!-- Bukti Nota/SJ -->
              <td class="py-4 px-6">
                <button
                  v-if="b.attachment_proof_url"
                  @click="openPreviewProof(b.attachment_proof_url)"
                  class="flex items-center gap-1 text-[11px] text-primary font-bold hover:underline"
                >
                  <Eye class="h-3.5 w-3.5" /> Lihat Bukti
                </button>
                <span v-else class="text-slate-400 text-[11px]">-</span>
              </td>

              <!-- Status Approval -->
              <td class="py-4 px-6">
                <Badge
                  :class="{
                    'bg-amber-100 text-amber-800 border-amber-200': b.status === 'PENDING',
                    'bg-emerald-100 text-emerald-800 border-emerald-200': b.status === 'APPROVED',
                    'bg-red-100 text-red-800 border-red-200': b.status === 'REJECTED',
                  }"
                  class="rounded-full px-2.5 py-0.5 font-black text-[10px] uppercase border"
                >
                  {{ b.status === 'PENDING' ? 'MENUNGGU APPROVAL' : b.status }}
                </Badge>
                <div v-if="b.approver" class="text-[10px] text-slate-400 mt-1">
                  Oleh: {{ b.approver.first_name }}
                </div>
              </td>

              <!-- Aksi Persetujuan -->
              <td class="py-4 px-6 text-right space-x-1">
                <template v-if="b.status === 'PENDING'">
                  <Button
                    size="sm"
                    @click="openApprovalModal(b, 'APPROVED')"
                    class="rounded-xl text-[11px] font-bold bg-emerald-600 hover:bg-emerald-700 text-white h-7 px-2.5 shadow-sm"
                  >
                    <Check class="h-3.5 w-3.5 mr-1" /> Approve
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    @click="openApprovalModal(b, 'REJECTED')"
                    class="rounded-xl text-[11px] font-bold text-red-600 border-red-200 hover:bg-red-50 h-7 px-2.5"
                  >
                    <X class="h-3.5 w-3.5 mr-1" /> Reject
                  </Button>
                </template>
                <span v-else class="text-[11px] text-slate-400 font-bold">Telah Diproses</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MODAL INPUT MUATAN BALIK -->
    <div
      v-if="isAddModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4 overflow-y-auto"
    >
      <div class="bg-white rounded-3xl max-w-xl w-full p-6 shadow-2xl border border-slate-100 space-y-5 my-8">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center font-black">
              <RotateCcw class="h-5 w-5" />
            </div>
            <div>
              <h3 class="text-base font-black text-slate-900">Input Pengiriman Muatan Balik (Backhaul)</h3>
              <p class="text-xs text-slate-500">Hitungan otomatis fee 10% dan alur persetujuan.</p>
            </div>
          </div>
          <button @click="isAddModalOpen = false" class="p-1.5 rounded-xl text-slate-400 hover:bg-slate-100">
            <X class="h-5 w-5" />
          </button>
        </div>

        <form @submit.prevent="handleSubmitBackhaul" class="space-y-4 text-xs font-bold text-slate-700">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="block mb-1 text-[10px] text-slate-500 uppercase tracking-wider">Driver *</label>
              <select
                v-model="form.driver_id"
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
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Lokasi Asal Balikan (Titik A) *</label>
              <input
                v-model="form.origin_location"
                required
                type="text"
                placeholder="Contoh: Palembang"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Lokasi Tujuan (Titik B) *</label>
              <input
                v-model="form.destination_location"
                required
                type="text"
                placeholder="Contoh: Sukoharjo / Pabrik"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Nama Customer / Pengirim *</label>
              <input
                v-model="form.customer_name"
                required
                type="text"
                placeholder="Nama Pengirim Barang"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
            <div>
              <label class="block mb-1 text-[10px] text-slate-500">No HP / WhatsApp Customer</label>
              <input
                v-model="form.customer_phone"
                type="text"
                placeholder="08xxxxxxxx"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Nama Muatan Kargo *</label>
              <input
                v-model="form.cargo_name"
                required
                type="text"
                placeholder="Contoh: 100 Karung Pupuk / Baut"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
            <div>
              <label class="block mb-1 text-[10px] text-slate-500">Estimasi Bobot (Kg)</label>
              <input
                v-model.number="form.cargo_weight_kg"
                type="number"
                placeholder="0"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
          </div>

          <!-- Total Ongkos & Fee 10% Live Calculation -->
          <div class="p-4 rounded-2xl bg-amber-50/80 border border-amber-200/80 space-y-3">
            <div>
              <label class="block mb-1 text-[10px] text-amber-900 uppercase tracking-wider font-extrabold">
                Biaya Total Ongkosan Balik (Rp) *
              </label>
              <input
                v-model.number="form.freight_cost"
                type="number"
                required
                placeholder="Contoh: 5000000"
                class="w-full p-3 bg-white border border-amber-200 rounded-xl text-sm font-black text-slate-900 outline-none font-mono"
              />
            </div>

            <div class="grid grid-cols-2 gap-2 pt-2 border-t border-amber-200/60 text-xs">
              <div class="bg-white p-2.5 rounded-xl border border-amber-100">
                <span class="text-[10px] text-slate-500 block">Hitungan Fee (10%):</span>
                <span class="text-amber-800 font-black font-mono text-sm">{{ formatCurrency(computedFeeAmount) }}</span>
              </div>
              <div class="bg-white p-2.5 rounded-xl border border-amber-100">
                <span class="text-[10px] text-slate-500 block">Nilai Bersih:</span>
                <span class="text-emerald-700 font-black font-mono text-sm">{{ formatCurrency(computedNetAmount) }}</span>
              </div>
            </div>
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Upload Foto Bukti Surat Jalan / Muatan</label>
            <input
              type="file"
              accept="image/*"
              @change="handleFileChange"
              class="w-full p-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
            />
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" @click="isAddModalOpen = false" class="rounded-xl">Batal</Button>
            <Button type="submit" class="rounded-xl bg-amber-600 hover:bg-amber-700 text-white shadow-lg shadow-amber-600/20">
              Kirim untuk Approval
            </Button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL APPROVAL -->
    <div
      v-if="isApprovalModalOpen && selectedBackhaul"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4"
    >
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-slate-100 space-y-4">
        <h3 class="text-base font-black text-slate-900">
          Konfirmasi {{ approvalAction === 'APPROVED' ? 'Persetujuan' : 'Penolakan' }} Muatan Balik
        </h3>
        <p class="text-xs text-slate-500">
          No Balikan: <strong class="font-mono text-slate-800">{{ selectedBackhaul.backhaul_number }}</strong>
          (Total Ongkos: {{ formatCurrency(selectedBackhaul.freight_cost) }} | Fee 10%: {{ formatCurrency(selectedBackhaul.fee_amount) }})
        </p>

        <div>
          <label class="block mb-1 text-[10px] text-slate-500">
            {{ approvalAction === 'APPROVED' ? 'Catatan Approval (Opsional)' : 'Alasan Penolakan *' }}
          </label>
          <textarea
            v-model="approvalNote"
            rows="3"
            :required="approvalAction === 'REJECTED'"
            placeholder="Tuliskan catatan persetujuan atau alasan penolakan..."
            class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
          ></textarea>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <Button variant="outline" @click="isApprovalModalOpen = false" class="rounded-xl">Batal</Button>
          <Button
            @click="handleApprovalSubmit"
            :class="approvalAction === 'APPROVED' ? 'bg-emerald-600 hover:bg-emerald-700' : 'bg-red-600 hover:bg-red-700'"
            class="rounded-xl text-white font-bold"
          >
            {{ approvalAction === 'APPROVED' ? 'Setujui (Approve)' : 'Tolak (Reject)' }}
          </Button>
        </div>
      </div>
    </div>

    <!-- MODAL PREVIEW PROOF IMAGE -->
    <div
      v-if="isPreviewProofModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/80 backdrop-blur-sm p-4"
      @click="isPreviewProofModalOpen = false"
    >
      <div class="max-w-2xl w-full bg-white p-3 rounded-3xl shadow-2xl" @click.stop>
        <div class="flex justify-between items-center pb-2 px-2 border-b border-slate-100">
          <span class="text-xs font-bold text-slate-700">Bukti Surat Jalan / Muatan</span>
          <button @click="isPreviewProofModalOpen = false" class="p-1 rounded-lg text-slate-400 hover:bg-slate-100">
            <X class="h-4 w-4" />
          </button>
        </div>
        <div class="p-2 max-h-[75vh] overflow-auto flex justify-center">
          <img :src="previewImageURL" class="max-w-full rounded-2xl object-contain shadow-inner" />
        </div>
      </div>
    </div>
  </div>
</template>
