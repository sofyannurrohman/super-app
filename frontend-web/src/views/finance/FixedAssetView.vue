<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import {
  Package,
  Plus,
  Edit,
  Trash2,
  Calendar,
  Check,
  X,
  RefreshCw,
  Building2
} from 'lucide-vue-next'

const isLoading = ref(false)
const assets = ref<any[]>([])
const summaryData = ref<any>(null)

const isModalOpen = ref(false)
const isSubmitting = ref(false)
const isEditMode = ref(false)

const scheduleModalOpen = ref(false)
const scheduleData = ref<any[]>([])
const scheduleAsset = ref<any>(null)
const isLoadingSchedule = ref(false)

const deprecateModalOpen = ref(false)
const deprecatePeriod = ref(new Date().toISOString().slice(0, 7)) // YYYY-MM
const isDeprecating = ref(false)
const deprecateAsset = ref<any>(null)

const form = ref({
  id: '',
  asset_code: '',
  name: '',
  category: 'MACHINERY',
  location: '',
  description: '',
  acquisition_date: new Date().toISOString().split('T')[0],
  acquisition_cost: 0,
  useful_life_years: 1,
  residual_value: 0,
  depreciation_method: 'STRAIGHT_LINE'
})

const fetchAssets = async () => {
  isLoading.value = true
  try {
    const res = await apiClient.get('/finance/fixed-assets')
    if (res.data) {
      assets.value = res.data.data || []
      summaryData.value = res.data.summary || null
    }
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal memuat data Aset Tetap')
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchAssets()
})

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const getCategoryBadgeColor = (category: string) => {
  switch (category) {
    case 'MACHINERY': return 'bg-blue-100 text-blue-700'
    case 'VEHICLE': return 'bg-green-100 text-green-700'
    case 'BUILDING': return 'bg-purple-100 text-purple-700'
    case 'IT_EQUIPMENT': return 'bg-indigo-100 text-indigo-700'
    case 'FURNITURE': return 'bg-amber-100 text-amber-700'
    default: return 'bg-gray-100 text-gray-700'
  }
}

const getStatusBadgeColor = (status: string) => {
  switch (status) {
    case 'ACTIVE': return 'bg-green-100 text-green-700'
    case 'DISPOSED': return 'bg-gray-100 text-gray-700'
    case 'SOLD': return 'bg-red-100 text-red-700'
    default: return 'bg-slate-100 text-slate-700'
  }
}

const openAddModal = () => {
  isEditMode.value = false
  form.value = {
    id: '',
    asset_code: '',
    name: '',
    category: 'MACHINERY',
    location: '',
    description: '',
    acquisition_date: new Date().toISOString().split('T')[0],
    acquisition_cost: 0,
    useful_life_years: 1,
    residual_value: 0,
    depreciation_method: 'STRAIGHT_LINE'
  }
  isModalOpen.value = true
}

const openEditModal = (item: any) => {
  isEditMode.value = true
  form.value = {
    id: item.id,
    asset_code: item.asset_code,
    name: item.name,
    category: item.category,
    location: item.location || '',
    description: item.description || '',
    acquisition_date: item.acquisition_date.split('T')[0],
    acquisition_cost: item.acquisition_cost,
    useful_life_years: item.useful_life_years,
    residual_value: item.residual_value,
    depreciation_method: item.depreciation_method
  }
  isModalOpen.value = true
}

const saveAsset = async () => {
  isSubmitting.value = true
  try {
    if (isEditMode.value) {
      await apiClient.put(`/finance/fixed-assets/${form.value.id}`, form.value)
      toast.success('Aset berhasil diperbarui')
    } else {
      await apiClient.post('/finance/fixed-assets', form.value)
      toast.success('Aset berhasil ditambahkan')
    }
    isModalOpen.value = false
    fetchAssets()
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal menyimpan aset')
  } finally {
    isSubmitting.value = false
  }
}

const deleteAsset = async (id: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus aset ini?')) return
  try {
    await apiClient.delete(`/finance/fixed-assets/${id}`)
    toast.success('Aset berhasil dihapus')
    fetchAssets()
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal menghapus aset')
  }
}

const openScheduleModal = async (item: any) => {
  scheduleAsset.value = item
  scheduleModalOpen.value = true
  isLoadingSchedule.value = true
  try {
    const res = await apiClient.get(`/finance/fixed-assets/${item.id}/schedule`)
    scheduleData.value = res.data?.data || []
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal memuat jadwal depresiasi')
  } finally {
    isLoadingSchedule.value = false
  }
}

const openDeprecateModal = (item: any) => {
  deprecateAsset.value = item
  deprecatePeriod.value = new Date().toISOString().slice(0, 7)
  deprecateModalOpen.value = true
}

const runDepreciation = async () => {
  if (!deprecateAsset.value) return
  isDeprecating.value = true
  try {
    await apiClient.post(`/finance/fixed-assets/${deprecateAsset.value.id}/depreciate?period=${deprecatePeriod.value}`)
    toast.success('Depresiasi berhasil dijalankan')
    deprecateModalOpen.value = false
    fetchAssets()
  } catch (e: any) {
    console.error(e)
    toast.error('Gagal menjalankan depresiasi')
  } finally {
    isDeprecating.value = false
  }
}

const calculatedMonthlyDepreciation = computed(() => {
  if (form.value.depreciation_method === 'STRAIGHT_LINE') {
    const cost = form.value.acquisition_cost || 0
    const residual = form.value.residual_value || 0
    const years = form.value.useful_life_years || 1
    const months = years * 12
    return months > 0 ? (cost - residual) / months : 0
  }
  return 0 // For DECLINING_BALANCE it's more complex, simplified for preview
})
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center gap-2">
          Manajemen Aset Tetap
        </h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <Package class="w-4 h-4 text-primary" />
          Pencatatan aset perusahaan dan depresiasi otomatis.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <button
          @click="fetchAssets"
          class="p-2.5 bg-white border border-slate-200 text-slate-600 hover:text-slate-900 rounded-2xl transition-all shadow-sm active:scale-95"
        >
          <RefreshCw class="w-5 h-5" :class="{ 'animate-spin': isLoading }" />
        </button>

        <button
          @click="openAddModal"
          class="flex items-center gap-2 bg-slate-900 text-white hover:bg-slate-800 px-5 py-2.5 rounded-2xl text-sm font-bold transition-all shadow-md active:scale-95"
        >
          <Plus class="w-4 h-4" />
          <span>Tambah Aset</span>
        </button>
      </div>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <div class="bg-blue-50/50 p-6 rounded-3xl border border-blue-100 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-blue-500 mb-2 uppercase">Total Aset</p>
        <p class="text-xl font-black text-blue-700">{{ summaryData?.total_assets || 0 }}</p>
      </div>
      <div class="bg-green-50/50 p-6 rounded-3xl border border-green-100 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-green-500 mb-2 uppercase">Nilai Perolehan</p>
        <p class="text-xl font-black text-green-700">{{ formatCurrency(summaryData?.total_acquisition_cost || 0) }}</p>
      </div>
      <div class="bg-amber-50/50 p-6 rounded-3xl border border-amber-100 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-amber-500 mb-2 uppercase">Akumulasi Depresiasi</p>
        <p class="text-xl font-black text-amber-700">{{ formatCurrency(summaryData?.total_accumulated_depreciation || 0) }}</p>
      </div>
      <div class="bg-purple-50/50 p-6 rounded-3xl border border-purple-100 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
        <p class="text-xs font-bold text-purple-500 mb-2 uppercase">Nilai Buku</p>
        <p class="text-xl font-black text-purple-700">{{ formatCurrency(summaryData?.total_book_value || 0) }}</p>
      </div>
    </div>

    <!-- Data Table -->
    <div v-if="isLoading" class="space-y-4">
      <div v-for="i in 3" :key="i" class="h-16 bg-slate-100 animate-pulse rounded-2xl w-full"></div>
    </div>

    <div v-else-if="assets.length === 0" class="bg-white rounded-3xl border border-slate-200/60 p-16 text-center text-slate-400 shadow-[0_4px_20px_rgba(0,0,0,0.02)]">
      <Package class="w-16 h-16 mx-auto mb-4 text-slate-300 opacity-50" />
      <h3 class="text-lg font-bold text-slate-700">Belum ada aset</h3>
      <p class="text-sm text-slate-500 mt-1">
        Silakan tambah aset baru.
      </p>
    </div>

    <div v-else class="bg-white rounded-3xl border border-slate-200/60 shadow-[0_4px_25px_rgba(0,0,0,0.03)] overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse min-w-[1000px]">
          <thead>
            <tr class="bg-slate-900 text-white uppercase text-[10px] font-black tracking-[0.15em]">
              <th class="px-6 py-5.5 border-r border-slate-800">Kode & Nama Aset</th>
              <th class="px-6 py-5.5 border-r border-slate-800">Kategori</th>
              <th class="px-6 py-5.5 border-r border-slate-800">Tgl Perolehan</th>
              <th class="px-6 py-5.5 border-r border-slate-800 text-right">Nilai Perolehan</th>
              <th class="px-6 py-5.5 border-r border-slate-800 text-right">Nilai Buku</th>
              <th class="px-6 py-5.5 border-r border-slate-800 text-center">Status</th>
              <th class="px-6 py-5.5 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-slate-700 font-medium">
            <tr v-for="item in assets" :key="item.id" class="hover:bg-slate-50">
              <td class="px-6 py-4.5 border-r border-slate-100">
                <div class="font-bold text-slate-900">{{ item.asset_code }}</div>
                <div class="text-sm text-slate-500">{{ item.name }}</div>
              </td>
              <td class="px-6 py-4.5 border-r border-slate-100">
                <span :class="['px-2.5 py-1 rounded-lg text-xs font-bold', getCategoryBadgeColor(item.category)]">
                  {{ item.category.replace('_', ' ') }}
                </span>
              </td>
              <td class="px-6 py-4.5 border-r border-slate-100 text-sm">{{ item.acquisition_date.split('T')[0] }}</td>
              <td class="px-6 py-4.5 text-right font-bold text-slate-800 border-r border-slate-100">{{ formatCurrency(item.acquisition_cost) }}</td>
              <td class="px-6 py-4.5 text-right font-bold text-purple-600 border-r border-slate-100">{{ formatCurrency(item.book_value) }}</td>
              <td class="px-6 py-4.5 text-center border-r border-slate-100">
                <span :class="['px-2.5 py-1 rounded-lg text-xs font-bold', getStatusBadgeColor(item.status)]">
                  {{ item.status }}
                </span>
              </td>
              <td class="px-6 py-4.5">
                <div class="flex items-center justify-center gap-2">
                  <button @click="openEditModal(item)" class="p-2 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors" title="Edit">
                    <Edit class="w-4 h-4" />
                  </button>
                  <button @click="openScheduleModal(item)" class="p-2 text-slate-400 hover:text-purple-600 hover:bg-purple-50 rounded-lg transition-colors" title="Jadwal Depresiasi">
                    <Calendar class="w-4 h-4" />
                  </button>
                  <button @click="openDeprecateModal(item)" class="p-2 text-slate-400 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors" title="Jalankan Depresiasi">
                    <Check class="w-4 h-4" />
                  </button>
                  <button @click="deleteAsset(item.id)" class="p-2 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors" title="Hapus">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <!-- Add/Edit Modal -->
  <div v-if="isModalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm">
    <div class="bg-white rounded-3xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
      <div class="flex items-center justify-between p-6 border-b border-slate-100">
        <h2 class="text-xl font-bold text-slate-900">{{ isEditMode ? 'Edit Aset' : 'Tambah Aset Baru' }}</h2>
        <button @click="isModalOpen = false" class="p-2 text-slate-400 hover:bg-slate-100 rounded-full transition-colors">
          <X class="w-5 h-5" />
        </button>
      </div>
      
      <form @submit.prevent="saveAsset" class="p-6 space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Kode Aset *</label>
            <input v-model="form.asset_code" required class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20" />
          </div>
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Nama Aset *</label>
            <input v-model="form.name" required class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20" />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Kategori *</label>
            <select v-model="form.category" required class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20">
              <option value="MACHINERY">Mesin</option>
              <option value="VEHICLE">Kendaraan</option>
              <option value="BUILDING">Bangunan</option>
              <option value="IT_EQUIPMENT">Peralatan IT</option>
              <option value="FURNITURE">Furnitur</option>
              <option value="OTHER">Lainnya</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Lokasi</label>
            <input v-model="form.location" class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20" />
          </div>
        </div>

        <div>
          <label class="block text-sm font-bold text-slate-700 mb-1">Deskripsi</label>
          <textarea v-model="form.description" rows="2" class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"></textarea>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Tanggal Perolehan *</label>
            <input type="date" v-model="form.acquisition_date" required class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20" />
          </div>
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Nilai Perolehan (IDR) *</label>
            <input type="number" v-model="form.acquisition_cost" required min="0" class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20" />
          </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Umur Ekonomis (Tahun) *</label>
            <input type="number" v-model="form.useful_life_years" required min="1" class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20" />
          </div>
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Nilai Sisa/Residu (IDR)</label>
            <input type="number" v-model="form.residual_value" min="0" class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20" />
          </div>
          <div>
            <label class="block text-sm font-bold text-slate-700 mb-1">Metode *</label>
            <select v-model="form.depreciation_method" required class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20">
              <option value="STRAIGHT_LINE">Garis Lurus</option>
              <option value="DECLINING_BALANCE">Saldo Menurun</option>
            </select>
          </div>
        </div>

        <div class="p-4 bg-blue-50 border border-blue-100 rounded-xl mt-4">
          <p class="text-xs font-bold text-blue-600 mb-1">Estimasi Depresiasi Bulanan</p>
          <p class="text-lg font-black text-blue-800">{{ formatCurrency(calculatedMonthlyDepreciation) }}</p>
        </div>

        <div class="pt-4 flex justify-end gap-3">
          <button type="button" @click="isModalOpen = false" class="px-5 py-2.5 bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 rounded-xl font-bold transition-colors">Batal</button>
          <button type="submit" :disabled="isSubmitting" class="px-5 py-2.5 bg-slate-900 text-white hover:bg-slate-800 rounded-xl font-bold transition-colors disabled:opacity-50">
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Aset' }}
          </button>
        </div>
      </form>
    </div>
  </div>

  <!-- Schedule Modal -->
  <div v-if="scheduleModalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm">
    <div class="bg-white rounded-3xl shadow-2xl w-full max-w-4xl max-h-[90vh] flex flex-col">
      <div class="flex items-center justify-between p-6 border-b border-slate-100">
        <div>
          <h2 class="text-xl font-bold text-slate-900">Jadwal Depresiasi</h2>
          <p class="text-sm text-slate-500">{{ scheduleAsset?.asset_code }} - {{ scheduleAsset?.name }}</p>
        </div>
        <button @click="scheduleModalOpen = false" class="p-2 text-slate-400 hover:bg-slate-100 rounded-full transition-colors">
          <X class="w-5 h-5" />
        </button>
      </div>
      
      <div class="p-6 overflow-y-auto flex-1">
        <div v-if="isLoadingSchedule" class="flex justify-center py-10">
          <RefreshCw class="w-8 h-8 text-primary animate-spin" />
        </div>
        <div v-else-if="scheduleData.length === 0" class="text-center py-10 text-slate-500">
          Belum ada jadwal depresiasi
        </div>
        <table v-else class="w-full text-left border-collapse text-sm">
          <thead>
            <tr class="bg-slate-50 text-slate-600 font-bold border-b border-slate-200">
              <th class="px-4 py-3">Periode</th>
              <th class="px-4 py-3 text-right">Nilai Buku Awal</th>
              <th class="px-4 py-3 text-right">Depresiasi</th>
              <th class="px-4 py-3 text-right">Akumulasi</th>
              <th class="px-4 py-3 text-right">Nilai Buku Akhir</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="(row, i) in scheduleData" :key="i" class="hover:bg-slate-50">
              <td class="px-4 py-3 font-medium text-slate-700">{{ row.period }}</td>
              <td class="px-4 py-3 text-right text-slate-600">{{ formatCurrency(row.beginning_book_value) }}</td>
              <td class="px-4 py-3 text-right text-red-500 font-bold">{{ formatCurrency(row.depreciation_amount) }}</td>
              <td class="px-4 py-3 text-right text-slate-600">{{ formatCurrency(row.accumulated_depreciation) }}</td>
              <td class="px-4 py-3 text-right text-purple-600 font-bold">{{ formatCurrency(row.ending_book_value) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <!-- Run Depreciation Confirm Modal -->
  <div v-if="deprecateModalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/50 backdrop-blur-sm">
    <div class="bg-white rounded-3xl shadow-2xl w-full max-w-md">
      <div class="p-6 border-b border-slate-100">
        <h2 class="text-xl font-bold text-slate-900">Jalankan Depresiasi</h2>
      </div>
      
      <div class="p-6 space-y-4">
        <p class="text-slate-600 text-sm">Anda akan menjalankan depresiasi untuk aset <strong>{{ deprecateAsset?.name }}</strong>.</p>
        
        <div>
          <label class="block text-sm font-bold text-slate-700 mb-1">Periode (Bulan/Tahun)</label>
          <input type="month" v-model="deprecatePeriod" required class="w-full bg-slate-50 border border-slate-200 px-4 py-2.5 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20" />
        </div>
      </div>
      
      <div class="p-6 pt-0 flex justify-end gap-3">
        <button @click="deprecateModalOpen = false" class="px-5 py-2.5 bg-white border border-slate-200 text-slate-700 hover:bg-slate-50 rounded-xl font-bold transition-colors">Batal</button>
        <button @click="runDepreciation" :disabled="isDeprecating" class="px-5 py-2.5 bg-green-600 text-white hover:bg-green-700 rounded-xl font-bold transition-colors disabled:opacity-50">
          {{ isDeprecating ? 'Proses...' : 'Jalankan' }}
        </button>
      </div>
    </div>
  </div>
</template>
