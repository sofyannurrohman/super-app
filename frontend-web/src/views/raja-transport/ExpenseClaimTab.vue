<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRajaTransportStore } from '@/stores/rajaTransport'
import { vehicleApi } from '@/api/vehicle'
import apiClient from '@/api/axios'
import {
  Receipt,
  Plus,
  Search,
  CheckCircle2,
  XCircle,
  Clock,
  Fuel,
  Wrench,
  DollarSign,
  User,
  Truck,
  Eye,
  RefreshCw,
  X,
  Check,
  CreditCard,
  Banknote,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

const transportStore = useRajaTransportStore()

const isAddModalOpen = ref(false)
const isApprovalModalOpen = ref(false)
const isPreviewReceiptModalOpen = ref(false)
const previewImageURL = ref('')
const selectedClaim = ref<any | null>(null)
const approvalAction = ref<'APPROVED' | 'REJECTED'>('APPROVED')
const rejectReason = ref('')

const vehiclesList = ref<any[]>([])
const driversList = ref<any[]>([])
const tripsList = ref<any[]>([])

const filterCategory = ref('ALL')
const filterStatus = ref('ALL')
const searchQuery = ref('')
const selectedFile = ref<File | null>(null)

const form = ref({
  trip_id: '',
  driver_id: '',
  vehicle_id: '',
  category: 'BBM',
  merchant_name: '',
  fuel_liter: 0,
  current_odometer: 0,
  amount: 0,
  description: '',
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
    transportStore.fetchExpenseClaims(),
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
    category: 'BBM',
    merchant_name: '',
    fuel_liter: 0,
    current_odometer: 0,
    amount: 0,
    description: '',
  }
  selectedFile.value = null
  isAddModalOpen.value = true
}

async function handleSubmitClaim() {
  try {
    const formData = new FormData()
    formData.append('driver_id', form.value.driver_id)
    formData.append('vehicle_id', form.value.vehicle_id)
    if (form.value.trip_id) formData.append('trip_id', form.value.trip_id)
    formData.append('category', form.value.category)
    formData.append('merchant_name', form.value.merchant_name)
    if (form.value.fuel_liter > 0) formData.append('fuel_liter', String(form.value.fuel_liter))
    formData.append('current_odometer', String(form.value.current_odometer))
    formData.append('amount', String(form.value.amount))
    formData.append('description', form.value.description)
    if (selectedFile.value) {
      formData.append('receipt_image', selectedFile.value)
    }

    await transportStore.createExpenseClaim(formData)
    isAddModalOpen.value = false
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal mengupload nota klaim')
  }
}

function openApprovalModal(item: any, action: 'APPROVED' | 'REJECTED') {
  selectedClaim.value = item
  approvalAction.value = action
  rejectReason.value = ''
  isApprovalModalOpen.value = true
}

async function handleApprovalSubmit() {
  if (!selectedClaim.value) return
  try {
    await transportStore.approveExpenseClaim(selectedClaim.value.id, {
      action: approvalAction.value,
      reason: approvalAction.value === 'REJECTED' ? rejectReason.value : undefined,
    })
    isApprovalModalOpen.value = false
    selectedClaim.value = null
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal memproses approval')
  }
}

async function handleDisburse(claimId: string) {
  if (!confirm('Tandai klaim ini sebagai telah dicairkan / dibayarkan?')) return
  try {
    await transportStore.disburseExpenseClaim(claimId)
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal mencairkan klaim')
  }
}

function openPreviewReceipt(url: string) {
  previewImageURL.value = url
  isPreviewReceiptModalOpen.value = true
}

const filteredClaims = computed(() => {
  return transportStore.expenseClaims.filter((c: any) => {
    if (filterCategory.value !== 'ALL' && c.category !== filterCategory.value) return false
    if (filterStatus.value !== 'ALL' && c.status !== filterStatus.value) return false
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase()
      const matchNum = (c.claim_number || '').toLowerCase().includes(q)
      const matchMerchant = (c.merchant_name || '').toLowerCase().includes(q)
      const matchDriver = (c.driver?.first_name || '').toLowerCase().includes(q)
      const matchPlate = (c.vehicle?.license_plate || '').toLowerCase().includes(q)
      if (!matchNum && !matchMerchant && !matchDriver && !matchPlate) return false
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
            placeholder="Cari no klaim, SPBU, driver, plat..."
            class="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-primary outline-none transition-all"
          />
        </div>

        <select
          v-model="filterCategory"
          class="px-3 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white"
        >
          <option value="ALL">Semua Kategori</option>
          <option value="BBM">BBM (Bahan Bakar)</option>
          <option value="SPAREPART_ARMADA">Pergantian Sparepart</option>
          <option value="TOLL_PARKING">Tol & Parkir</option>
          <option value="EMERGENCY_REPAIR">Perbaikan Darurat</option>
        </select>

        <select
          v-model="filterStatus"
          class="px-3 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs font-bold text-slate-700 outline-none focus:bg-white"
        >
          <option value="ALL">Semua Status</option>
          <option value="PENDING">Menunggu Persetujuan</option>
          <option value="APPROVED">Disetujui</option>
          <option value="REJECTED">Ditolak</option>
        </select>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <Button variant="outline" size="sm" @click="transportStore.fetchExpenseClaims()" class="rounded-2xl text-xs font-bold">
          <RefreshCw class="h-3.5 w-3.5 mr-1.5" :class="{ 'animate-spin': transportStore.loading }" />
          Refresh
        </Button>
        <Button @click="openAddModal" class="rounded-2xl text-xs font-bold shadow-md shadow-primary/20">
          <Plus class="h-4 w-4 mr-1.5" />
          Upload Struk Nota
        </Button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-3xl border border-slate-200/80 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-100 bg-slate-50/50 text-[11px] font-black text-slate-400 uppercase tracking-wider">
              <th class="py-4 px-6">No Klaim & Tanggal</th>
              <th class="py-4 px-6">Kategori & Merchant</th>
              <th class="py-4 px-6">Driver & Armada</th>
              <th class="py-4 px-6">Liter & Odometer KM</th>
              <th class="py-4 px-6">Jumlah Nominal</th>
              <th class="py-4 px-6">Foto Struk</th>
              <th class="py-4 px-6">Status Approval</th>
              <th class="py-4 px-6 text-right">Aksi Persetujuan</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-xs font-bold">
            <tr v-if="transportStore.loading" class="text-center py-10">
              <td colspan="8" class="py-12 text-center text-slate-400 font-bold">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
                Memuat data klaim pengeluaran...
              </td>
            </tr>
            <tr v-else-if="filteredClaims.length === 0">
              <td colspan="8" class="py-12 text-center text-slate-400">
                <Receipt class="h-10 w-10 mx-auto mb-2 opacity-30 text-slate-400" />
                <p>Belum ada nota klaim yang diupload.</p>
              </td>
            </tr>
            <tr
              v-for="c in filteredClaims"
              :key="c.id"
              class="hover:bg-slate-50/80 transition-colors"
            >
              <!-- No Klaim -->
              <td class="py-4 px-6">
                <div class="font-mono text-slate-900 font-extrabold">{{ c.claim_number }}</div>
                <div class="text-[10px] text-slate-400">{{ formatDate(c.created_at) }}</div>
                <div v-if="c.trip?.trip_number" class="text-[10px] text-primary font-mono mt-0.5">
                  Trip: {{ c.trip.trip_number }}
                </div>
              </td>

              <!-- Kategori & Merchant -->
              <td class="py-4 px-6">
                <Badge
                  :class="{
                    'bg-amber-100 text-amber-800 border-amber-200': c.category === 'BBM',
                    'bg-blue-100 text-blue-800 border-blue-200': c.category === 'SPAREPART_ARMADA',
                    'bg-purple-100 text-purple-800 border-purple-200': c.category === 'TOLL_PARKING',
                    'bg-red-100 text-red-800 border-red-200': c.category === 'EMERGENCY_REPAIR',
                  }"
                  class="font-black text-[9px] px-2 py-0.5 rounded-md border uppercase mb-1"
                >
                  {{ c.category }}
                </Badge>
                <div class="text-slate-900 font-extrabold">{{ c.merchant_name || '-' }}</div>
              </td>

              <!-- Driver & Armada -->
              <td class="py-4 px-6">
                <div class="text-slate-800 font-bold flex items-center gap-1">
                  <User class="h-3.5 w-3.5 text-primary" />
                  {{ c.driver?.first_name }} {{ c.driver?.last_name || '' }}
                </div>
                <div class="text-[11px] text-slate-500 font-mono mt-0.5">
                  {{ c.vehicle?.license_plate || '-' }}
                </div>
              </td>

              <!-- Liter & KM Odometer -->
              <td class="py-4 px-6">
                <div v-if="c.fuel_liter" class="text-slate-900 font-bold">
                  {{ c.fuel_liter }} Liter
                </div>
                <div class="text-[11px] text-slate-400 font-mono">
                  KM: {{ c.current_odometer?.toLocaleString() || '-' }}
                </div>
              </td>

              <!-- Nominal -->
              <td class="py-4 px-6">
                <div class="font-mono text-slate-900 font-black text-sm">
                  {{ formatCurrency(c.amount) }}
                </div>
                <div class="text-[10px]" :class="c.disbursement_status === 'PAID' ? 'text-emerald-600 font-bold' : 'text-slate-400'">
                  {{ c.disbursement_status === 'PAID' ? '✓ Sudah Dicairkan' : 'Belum Dicairkan' }}
                </div>
              </td>

              <!-- Foto Struk -->
              <td class="py-4 px-6">
                <button
                  v-if="c.receipt_image_url"
                  @click="openPreviewReceipt(c.receipt_image_url)"
                  class="flex items-center gap-1 text-[11px] text-primary font-bold hover:underline"
                >
                  <Eye class="h-3.5 w-3.5" /> Lihat Struk
                </button>
                <span v-else class="text-slate-400 text-[11px]">-</span>
              </td>

              <!-- Status Approval -->
              <td class="py-4 px-6">
                <Badge
                  :class="{
                    'bg-amber-100 text-amber-800 border-amber-200': c.status === 'PENDING',
                    'bg-emerald-100 text-emerald-800 border-emerald-200': c.status === 'APPROVED',
                    'bg-red-100 text-red-800 border-red-200': c.status === 'REJECTED',
                  }"
                  class="rounded-full px-2.5 py-0.5 font-black text-[10px] uppercase border"
                >
                  {{ c.status === 'PENDING' ? 'MENUNGGU APPROVAL' : c.status }}
                </Badge>
                <div v-if="c.approver" class="text-[10px] text-slate-400 mt-1">
                  Oleh: {{ c.approver.first_name }}
                </div>
              </td>

              <!-- Aksi Persetujuan & Pencairan -->
              <td class="py-4 px-6 text-right space-x-1">
                <template v-if="c.status === 'PENDING'">
                  <Button
                    size="sm"
                    @click="openApprovalModal(c, 'APPROVED')"
                    class="rounded-xl text-[11px] font-bold bg-emerald-600 hover:bg-emerald-700 text-white h-7 px-2.5 shadow-sm"
                  >
                    <Check class="h-3.5 w-3.5 mr-1" /> Approve
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    @click="openApprovalModal(c, 'REJECTED')"
                    class="rounded-xl text-[11px] font-bold text-red-600 border-red-200 hover:bg-red-50 h-7 px-2.5"
                  >
                    <X class="h-3.5 w-3.5 mr-1" /> Reject
                  </Button>
                </template>
                <template v-else-if="c.status === 'APPROVED' && c.disbursement_status !== 'PAID'">
                  <Button
                    size="sm"
                    @click="handleDisburse(c.id)"
                    class="rounded-xl text-[11px] font-bold bg-blue-600 hover:bg-blue-700 text-white h-7 px-2.5"
                  >
                    <Banknote class="h-3.5 w-3.5 mr-1" /> Cairkan
                  </Button>
                </template>
                <span v-else class="text-[11px] text-slate-400 font-bold">Selesai</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- MODAL UPLOAD KLAIM -->
    <div
      v-if="isAddModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4 overflow-y-auto"
    >
      <div class="bg-white rounded-3xl max-w-xl w-full p-6 shadow-2xl border border-slate-100 space-y-5 my-8">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <div class="flex items-center gap-3">
            <div class="h-10 w-10 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center font-black">
              <Receipt class="h-5 w-5" />
            </div>
            <div>
              <h3 class="text-base font-black text-slate-900">Upload Struk Nota Pengeluaran</h3>
              <p class="text-xs text-slate-500">Klaim biaya BBM, sparepart armada, tol, dan perbaikan.</p>
            </div>
          </div>
          <button @click="isAddModalOpen = false" class="p-1.5 rounded-xl text-slate-400 hover:bg-slate-100">
            <X class="h-5 w-5" />
          </button>
        </div>

        <form @submit.prevent="handleSubmitClaim" class="space-y-4 text-xs font-bold text-slate-700">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="block mb-1 text-[10px] text-slate-500 uppercase tracking-wider">Kategori Pengeluaran *</label>
              <select
                v-model="form.category"
                required
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              >
                <option value="BBM">BBM (Bahan Bakar Minyak)</option>
                <option value="SPAREPART_ARMADA">Pergantian Part Armada</option>
                <option value="TOLL_PARKING">Tol & Parkir</option>
                <option value="EMERGENCY_REPAIR">Perbaikan Darurat / Bengkel</option>
              </select>
            </div>

            <div>
              <label class="block mb-1 text-[10px] text-slate-500 uppercase tracking-wider">Nama Toko / SPBU / Bengkel *</label>
              <input
                v-model="form.merchant_name"
                required
                type="text"
                placeholder="Contoh: SPBU Pertamina 44.xxx / Toko Onderdil Jaya"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
              />
            </div>
          </div>

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

          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div v-if="form.category === 'BBM'">
              <label class="block mb-1 text-[10px] text-slate-500">Liter BBM</label>
              <input
                v-model.number="form.fuel_liter"
                type="number"
                step="0.01"
                placeholder="0"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none font-mono"
              />
            </div>

            <div :class="form.category === 'BBM' ? '' : 'col-span-2'">
              <label class="block mb-1 text-[10px] text-slate-500">KM Odometer Saat Transaksi</label>
              <input
                v-model.number="form.current_odometer"
                type="number"
                placeholder="0"
                class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none font-mono"
              />
            </div>

            <div>
              <label class="block mb-1 text-[10px] text-slate-500 font-black text-amber-900">Nominal Total (Rp) *</label>
              <input
                v-model.number="form.amount"
                type="number"
                required
                placeholder="Contoh: 350000"
                class="w-full p-2.5 bg-amber-50/50 border border-amber-200 rounded-xl text-xs font-black text-slate-900 outline-none font-mono"
              />
            </div>
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Keterangan / Detail Part yang diganti</label>
            <textarea
              v-model="form.description"
              rows="2"
              placeholder="Contoh: Beli kampas rem belakang darurat di Palembang..."
              class="w-full p-2.5 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold outline-none"
            ></textarea>
          </div>

          <div>
            <label class="block mb-1 text-[10px] text-slate-500">Foto Struk / Nota Fisik *</label>
            <input
              type="file"
              accept="image/*"
              required
              @change="handleFileChange"
              class="w-full p-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
            />
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" @click="isAddModalOpen = false" class="rounded-xl">Batal</Button>
            <Button type="submit" class="rounded-xl bg-amber-600 hover:bg-amber-700 text-white shadow-lg shadow-amber-600/20">
              Kirim Nota untuk Persetujuan
            </Button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL APPROVAL -->
    <div
      v-if="isApprovalModalOpen && selectedClaim"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-sm p-4"
    >
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-slate-100 space-y-4">
        <h3 class="text-base font-black text-slate-900">
          Konfirmasi {{ approvalAction === 'APPROVED' ? 'Persetujuan' : 'Penolakan' }} Klaim
        </h3>
        <p class="text-xs text-slate-500">
          No Klaim: <strong class="font-mono text-slate-800">{{ selectedClaim.claim_number }}</strong>
          (Nominal: {{ formatCurrency(selectedClaim.amount) }} - {{ selectedClaim.category }})
        </p>

        <div v-if="approvalAction === 'REJECTED'">
          <label class="block mb-1 text-[10px] text-slate-500">Alasan Penolakan *</label>
          <textarea
            v-model="rejectReason"
            rows="3"
            required
            placeholder="Tuliskan alasan penolakan..."
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

    <!-- PREVIEW IMAGE MODAL -->
    <div
      v-if="isPreviewReceiptModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/80 backdrop-blur-sm p-4"
      @click="isPreviewReceiptModalOpen = false"
    >
      <div class="max-w-2xl w-full bg-white p-3 rounded-3xl shadow-2xl" @click.stop>
        <div class="flex justify-between items-center pb-2 px-2 border-b border-slate-100">
          <span class="text-xs font-bold text-slate-700">Foto Struk / Nota Fisik</span>
          <button @click="isPreviewReceiptModalOpen = false" class="p-1 rounded-lg text-slate-400 hover:bg-slate-100">
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
