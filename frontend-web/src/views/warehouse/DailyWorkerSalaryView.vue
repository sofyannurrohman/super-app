<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Wallet,
  History,
  Search,
  Plus,
  Truck,
  Users,
  Calendar,
  Save,
  AlertCircle,
  Package
} from 'lucide-vue-next'

import { onMounted, watch } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { Eye, X } from 'lucide-vue-next'

const masterDataStore = useMasterDataStore()
const salaryHistory = ref<any[]>([])
const dailyWorkers = ref<any[]>([])
const kioskEmployees = ref<any[]>([])

const showPreviewModal = ref(false)
const previewImageUrl = ref('')
const previewModalTitle = ref('')
const previewModalSubtitle = ref('')

const getFullImageUrl = (path?: string | null): string => {
  if (!path) return ''
  if (path.startsWith('http') || path.startsWith('data:')) return path
  const apiUrl = import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || window.location.origin
  try {
    const origin = new URL(apiUrl).origin
    return origin + (path.startsWith('/') ? path : `/${path}`)
  } catch {
    return path
  }
}

const openPreviewModal = (url: string, title = 'Preview', subtitle = '') => {
  previewImageUrl.value = url
  previewModalTitle.value = title
  previewModalSubtitle.value = subtitle
  showPreviewModal.value = true
}

const closePreviewModal = () => {
  showPreviewModal.value = false
  previewImageUrl.value = ''
}

const activeTab = ref('kuli')
const searchQuery = ref('')
const isLoading = ref(false)

const kuliForm = ref({
  daily_worker_id: '',
  date: '',
  totalKg: 0,
  ratePerKg: 14000,
  geserFee: 0
})
const kuliFile = ref<File | null>(null)
const kuliFileInput = ref<HTMLInputElement | null>(null)

const onFileChangeKuli = (e: any) => {
  kuliFile.value = e.target.files[0] || null
}

const triggerKuliFilePicker = () => kuliFileInput.value?.click()

const calculatedKuliTotal = computed(() => {
  return (kuliForm.value.totalKg * (kuliForm.value.ratePerKg || 0)) + (kuliForm.value.geserFee || 0)
})

const packingForm = ref({
  daily_worker_id: '',
  date: '',
  nominal: 0,
  notes: ''
})
const packingFile = ref<File | null>(null)
const packingFileInput = ref<HTMLInputElement | null>(null)

const onFileChangePacking = (e: any) => {
  packingFile.value = e.target.files[0] || null
}

const triggerPackingFilePicker = () => packingFileInput.value?.click()

const availableWorkers = computed(() => {
  if (masterDataStore.isAllBranchesSelected) {
    return kioskEmployees.value
  }
  return kioskEmployees.value.filter((w: any) => !w.branch_id || masterDataStore.isBranchSelected(w.branch_id))
})

const resolveDailyWorkerId = async (employeeId: string, workerType: string) => {
  const emp = kioskEmployees.value.find(e => e.id === employeeId)
  if (!emp) throw new Error('Pekerja tidak ditemukan.')
  const fullName = emp.last_name ? `${emp.first_name} ${emp.last_name}`.trim() : emp.first_name.trim()

  const branchId = masterDataStore.selectedBranchId
  const targetBranchId = (branchId && branchId !== 'ALL_BRANCHES') ? branchId : emp.branch_id

  const existingDw = dailyWorkers.value.find(dw => dw.name === fullName && dw.type === workerType)
  if (existingDw) {
    return existingDw.id
  } else {
    const res = await apiClient.post('/daily-workers', {
      name: fullName,
      type: workerType,
      branch_id: targetBranchId || undefined
    })
    const newId = res.data?.data?.id || res.data?.id
    if (!newId) throw new Error('Gagal mendapatkan ID pekerja dari server. Hubungi administrator.')
    // Refresh daily workers list so future lookups work without re-creating
    const branchParam = (branchId && branchId !== 'ALL_BRANCHES') ? `?branch_id=${branchId}` : ''
    const wRes = await apiClient.get(`/daily-workers${branchParam}`)
    const wData = wRes.data?.data
    dailyWorkers.value = Array.isArray(wData) ? wData : (Array.isArray(wRes.data) ? wRes.data : [])
    return newId
  }
}

const filteredHistory = computed(() => {
  if (!searchQuery.value) return salaryHistory.value
  return salaryHistory.value.filter(h => 
    h.daily_worker?.name?.toLowerCase().includes(searchQuery.value.toLowerCase()) || 
    h.notes?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    h.branch?.name?.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const formatRupiah = (value: number) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value)
}

const fetchData = async () => {
  try {
    isLoading.value = true
    const branchId = masterDataStore.selectedBranchId
    const hasBranch = branchId && branchId !== 'ALL_BRANCHES'
    const branchParam = hasBranch ? `?branch_id=${branchId}` : ''
    const empBranchParam = hasBranch ? `&branch_id=${branchId}` : ''

    const [workersRes, salariesRes, empRes] = await Promise.all([
      apiClient.get(`/daily-workers${branchParam}`),
      apiClient.get(`/daily-worker-salaries${branchParam}`),
      apiClient.get(`/employees?employment_type=harian&limit=1000${empBranchParam}`)
    ])
    const wData = workersRes.data?.data;
    dailyWorkers.value = Array.isArray(wData) ? wData : (Array.isArray(workersRes.data) ? workersRes.data : [])
    const sData = salariesRes.data?.data;
    salaryHistory.value = Array.isArray(sData) ? sData : (Array.isArray(salariesRes.data) ? salariesRes.data : [])
    kioskEmployees.value = empRes.data?.data || empRes.data || []
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    isLoading.value = false
  }
}

watch(() => masterDataStore.selectedBranchId, () => {
  fetchData()
})

onMounted(() => {
  fetchData()
})

const submitKuliSalary = async () => {
  if (!kuliForm.value.daily_worker_id || kuliForm.value.totalKg <= 0 || !kuliForm.value.date) {
    toast.error('Mohon lengkapi data dengan benar')
    return
  }

  try {
    const finalDailyWorkerId = await resolveDailyWorkerId(kuliForm.value.daily_worker_id, 'Kuli')
    const emp = kioskEmployees.value.find(e => e.id === kuliForm.value.daily_worker_id)
    const branchId = masterDataStore.selectedBranchId
    const targetBranchId = (branchId && branchId !== 'ALL_BRANCHES') ? branchId : emp?.branch_id

    const formData = new FormData()
    formData.append('daily_worker_id', String(finalDailyWorkerId))
    if (targetBranchId) {
      formData.append('branch_id', String(targetBranchId))
    }
    formData.append('date', kuliForm.value.date)
    formData.append('total_paid', String(calculatedKuliTotal.value))
    formData.append('notes', `Gaji Kuli: ${kuliForm.value.totalKg} Kg @ Rp${kuliForm.value.ratePerKg}. Geser: ${kuliForm.value.geserFee}`)
    if (kuliFile.value) formData.append('proof_image', kuliFile.value)

    await apiClient.post('/daily-worker-salaries', formData)
    toast.success('Gaji Kuli berhasil disimpan!')
    kuliForm.value = { daily_worker_id: '', date: '', totalKg: 0, ratePerKg: 14000, geserFee: 0 } as any
    kuliFile.value = null
    fetchData()
  } catch (error: any) {
    toast.error(error?.message || 'Gagal menyimpan gaji')
  }
}

const submitPackingSalary = async () => {
  if (!packingForm.value.daily_worker_id || packingForm.value.nominal <= 0) {
    toast.error('Mohon lengkapi data packing.')
    return
  }
  
  try {
    const finalDailyWorkerId = await resolveDailyWorkerId(packingForm.value.daily_worker_id, 'Packing')
    const emp = kioskEmployees.value.find(e => e.id === packingForm.value.daily_worker_id)
    const branchId = masterDataStore.selectedBranchId
    const targetBranchId = (branchId && branchId !== 'ALL_BRANCHES') ? branchId : emp?.branch_id

    const formData = new FormData()
    formData.append('daily_worker_id', String(finalDailyWorkerId))
    if (targetBranchId) {
      formData.append('branch_id', String(targetBranchId))
    }
    formData.append('total_paid', String(packingForm.value.nominal))
    formData.append('notes', packingForm.value.notes)
    if (packingFile.value) formData.append('proof_image', packingFile.value)

    await apiClient.post('/daily-worker-salaries', formData)
    toast.success('Gaji Packing berhasil disimpan!')
    packingForm.value = { daily_worker_id: '', date: '', nominal: 0, notes: '' }
    packingFile.value = null
    fetchData()
  } catch (error: any) {
    toast.error(error?.message || 'Gagal menyimpan gaji')
  }
}
</script>

<template>
  <div class="p-6 max-w-7xl mx-auto">
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-slate-900 flex items-center gap-2">
        <Wallet class="w-6 h-6 text-primary" />
        Manajemen Gaji Pekerja Harian
      </h1>
      <p class="text-slate-500 mt-1">Kelola dan catat pembayaran gaji untuk Kuli dan Driver harian.</p>
    </div>

    <!-- Two-column Layout -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      
      <!-- Left Column: History -->
      <div class="lg:col-span-2 space-y-4">
        <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden flex flex-col h-full">
          <div class="p-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
            <h2 class="font-semibold text-slate-800 flex items-center gap-2">
              <History class="w-5 h-5 text-slate-500" />
              Riwayat Pembayaran
            </h2>
            <div class="relative w-64">
              <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
              <input 
                v-model="searchQuery"
                type="text" 
                placeholder="Cari pekerja..."
                class="w-full pl-9 pr-4 py-2 bg-white border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
              >
            </div>
          </div>
          
          <div class="p-4 flex-1 overflow-auto">
            <div class="space-y-3">
              <div v-for="item in filteredHistory" :key="item.id" class="p-4 rounded-xl border border-slate-100 hover:border-slate-200 hover:shadow-sm transition-all bg-white flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div class="flex items-start gap-4">
                  <div class="p-2.5 rounded-lg shrink-0" :class="item.daily_worker?.type === 'Kuli' ? 'bg-orange-50 text-orange-600' : 'bg-blue-50 text-blue-600'">
                    <Users v-if="item.daily_worker?.type === 'Kuli'" class="w-5 h-5" />
                    <Truck v-else class="w-5 h-5" />
                  </div>
                  <div>
                    <h3 class="font-bold text-slate-900">{{ item.daily_worker?.name }}</h3>
                    <div class="flex flex-wrap items-center gap-3 mt-1 text-sm text-slate-500">
                      <span class="flex items-center gap-1"><Calendar class="w-3.5 h-3.5" /> {{ new Date(item.date).toLocaleDateString('id-ID') }}</span>
                      <span class="flex items-center gap-1 text-slate-600 border-l border-slate-200 pl-3">
                        {{ item.notes }}
                      </span>
                    </div>
                  </div>
                </div>
                <div class="text-right flex flex-row sm:flex-col items-center sm:items-end justify-between sm:justify-center w-full sm:w-auto mt-2 sm:mt-0 pt-3 sm:pt-0 border-t sm:border-0 border-slate-100">
                  <span class="font-bold text-slate-900 text-lg">{{ formatRupiah(item.total_paid) }}</span>
                  <div class="flex items-center gap-2 mt-1">
                    <button v-if="item.proof_url" @click="openPreviewModal(getFullImageUrl(item.proof_url), 'Bukti Hitung', item.daily_worker?.name)" class="text-xs font-medium px-2 py-1 rounded-full bg-indigo-50 text-indigo-700 hover:bg-indigo-100 border border-indigo-200 flex items-center gap-1 transition-colors">
                      <Eye class="w-3.5 h-3.5" /> Lihat Bukti
                    </button>
                    <span v-if="item.status === 'PAID'" class="text-xs font-medium px-2 py-1 rounded-full bg-green-50 text-green-700 border border-green-200/50">Lunas</span>
                    <span v-else class="text-xs font-medium px-2 py-1 rounded-full bg-yellow-50 text-yellow-700 border border-yellow-200/50">Menunggu Kasir</span>
                  </div>
                </div>
              </div>
              
              <div v-if="filteredHistory.length === 0" class="py-12 text-center flex flex-col items-center">
                <div class="w-16 h-16 bg-slate-50 rounded-full flex items-center justify-center mb-3">
                  <History class="w-8 h-8 text-slate-300" />
                </div>
                <h3 class="text-slate-900 font-medium">Belum ada riwayat</h3>
                <p class="text-slate-500 text-sm mt-1">Pembayaran yang diinput akan muncul di sini.</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column: Input Form -->
      <div class="space-y-4">
        <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden sticky top-6">
          <div class="p-4 border-b border-slate-100 bg-slate-50/50">
            <h2 class="font-semibold text-slate-800 flex items-center gap-2">
              <Plus class="w-5 h-5 text-slate-500" />
              Input Gaji Pekerja
            </h2>
          </div>
          
          <div class="p-4">
            <!-- Type Toggle -->
            <div class="flex p-1 bg-slate-100 rounded-xl mb-6">
              <button 
                @click="activeTab = 'kuli'"
                class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all flex items-center justify-center gap-2"
                :class="activeTab === 'kuli' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
              >
                <Users class="w-4 h-4 hidden sm:block" /> Kuli
              </button>
              <button 
                @click="activeTab = 'packing'"
                class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all flex items-center justify-center gap-2"
                :class="activeTab === 'packing' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
              >
                <Package class="w-4 h-4 hidden sm:block" /> Packing
              </button>
              <button 
                @click="activeTab = 'driver'"
                class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all flex items-center justify-center gap-2"
                :class="activeTab === 'driver' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-700'"
              >
                <Truck class="w-4 h-4 hidden sm:block" /> Driver
              </button>
            </div>

            <!-- Kuli Form -->
            <div v-if="activeTab === 'kuli'" class="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700">Nama Kuli</label>
                <select v-model="kuliForm.daily_worker_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white">
                  <option value="">Pilih Kuli...</option>
                  <option v-for="w in availableWorkers" :key="w.id" :value="w.id">
                    {{ w.first_name }} {{ w.last_name || '' }} ({{ w.pin ? 'PIN: '+w.pin : 'Tanpa PIN' }})
                  </option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700">Tanggal</label>
                <input v-model="kuliForm.date" type="date" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700">
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-slate-700 flex justify-between">
                    <span>Total Muatan (KG)</span>
                  </label>
                  <input v-model="kuliForm.totalKg" type="number" min="0" placeholder="0" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all">
                </div>
                <div class="space-y-1.5">
                  <label class="text-sm font-medium text-slate-700 flex justify-between">
                    <span>Tarif per KG (Rp)</span>
                  </label>
                  <input v-model="kuliForm.ratePerKg" type="number" min="0" placeholder="14000" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all">
                </div>
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700 flex items-center gap-1">
                  Biaya Geser Barang <span class="text-xs text-slate-400 font-normal">(Opsional)</span>
                </label>
                <div class="relative">
                  <span class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 font-medium">Rp</span>
                  <input v-model="kuliForm.geserFee" type="number" min="0" placeholder="0" class="w-full pl-10 pr-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all">
                </div>
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700 flex items-center gap-1">
                  Upload Foto Bukti Hitung <span class="text-xs text-slate-400 font-normal">(Wajib)</span>
                </label>
                <input ref="kuliFileInput" type="file" accept="image/*" @change="onFileChangeKuli" class="hidden" />
                <button type="button" @click="triggerKuliFilePicker"
                  class="w-full flex items-center gap-3 px-4 py-2.5 border border-slate-200 rounded-xl text-sm text-slate-600 hover:bg-slate-50 transition-colors text-left">
                  <span class="shrink-0 text-xs font-semibold bg-indigo-50 text-indigo-600 px-2.5 py-1 rounded-lg">Pilih File</span>
                  <span class="truncate">{{ kuliFile ? kuliFile.name : 'Belum ada file dipilih...' }}</span>
                </button>
              </div>
              
              <div class="mt-6 pt-4 border-t border-slate-100">
                <div class="flex justify-between items-center mb-4">
                  <span class="text-sm font-medium text-slate-500">Estimasi Pembayaran</span>
                  <span class="text-xl font-bold text-primary">{{ formatRupiah(calculatedKuliTotal) }}</span>
                </div>
                <button @click="submitKuliSalary" class="w-full py-2.5 bg-primary hover:bg-primary/90 text-white font-semibold rounded-xl flex items-center justify-center gap-2 transition-all shadow-sm">
                  <Save class="w-4 h-4" /> Simpan Pembayaran Kuli
                </button>
              </div>
            </div>

            <!-- Driver Form -->
            <div v-if="activeTab === 'driver'" class="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
              <div class="py-12 text-center flex flex-col items-center">
                <div class="w-16 h-16 bg-blue-50 rounded-full flex items-center justify-center mb-3">
                  <Truck class="w-8 h-8 text-blue-300" />
                </div>
                <h3 class="text-slate-900 font-medium px-4">Tagihan Gaji Driver Harian (Eksternal) otomatis dicatat saat Surat Jalan diselesaikan oleh admin.</h3>
                <p class="text-slate-500 text-sm mt-2 px-4">Jika ada penyesuaian, hubungi pihak admin logistik atau keuangan.</p>
              </div>
            </div>

            <!-- Packing Form -->
            <div v-if="activeTab === 'packing'" class="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700">Nama Pekerja Packing</label>
                <select v-model="packingForm.daily_worker_id" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all bg-white">
                  <option value="">Pilih Pekerja...</option>
                  <option v-for="w in availableWorkers" :key="w.id" :value="w.id">
                    {{ w.first_name }} {{ w.last_name || '' }} ({{ w.pin ? 'PIN: '+w.pin : 'Tanpa PIN' }})
                  </option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700">Tanggal</label>
                <input v-model="packingForm.date" type="date" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700">
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700 flex justify-between">
                  <span>Nominal Pembayaran (Rp)</span>
                </label>
                <input v-model="packingForm.nominal" type="number" min="0" placeholder="0" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all">
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700">Catatan <span class="text-xs text-slate-400 font-normal">(Rincian borongan pcs/karton)</span></label>
                <textarea v-model="packingForm.notes" rows="2" placeholder="Contoh: 1000 pcs Kecap Manis" class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all text-slate-700"></textarea>
              </div>
              <div class="space-y-1.5">
                <label class="text-sm font-medium text-slate-700 flex items-center gap-1">
                  Upload Foto Bukti Hitung <span class="text-xs text-slate-400 font-normal">(Wajib)</span>
                </label>
                <input ref="packingFileInput" type="file" accept="image/*" @change="onFileChangePacking" class="hidden" />
                <button type="button" @click="triggerPackingFilePicker"
                  class="w-full flex items-center gap-3 px-4 py-2.5 border border-slate-200 rounded-xl text-sm text-slate-600 hover:bg-slate-50 transition-colors text-left">
                  <span class="shrink-0 text-xs font-semibold bg-indigo-50 text-indigo-600 px-2.5 py-1 rounded-lg">Pilih File</span>
                  <span class="truncate">{{ packingFile ? packingFile.name : 'Belum ada file dipilih...' }}</span>
                </button>
              </div>
              
              <div class="mt-6 pt-4 border-t border-slate-100">
                <div class="flex justify-between items-center mb-4">
                  <span class="text-sm font-medium text-slate-500">Estimasi Pembayaran</span>
                  <span class="text-xl font-bold text-primary">{{ formatRupiah(packingForm.nominal) }}</span>
                </div>
                <button @click="submitPackingSalary" class="w-full py-2.5 bg-primary hover:bg-primary/90 text-white font-semibold rounded-xl flex items-center justify-center gap-2 transition-all shadow-sm">
                  <Save class="w-4 h-4" /> Simpan Gaji Packing
                </button>
              </div>
            </div>

          </div>
        </div>
      </div>

    </div>

    <!-- Image Preview Modal -->
    <div v-if="showPreviewModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/80 backdrop-blur-sm transition-opacity">
      <div class="bg-white rounded-2xl shadow-2xl w-full max-w-3xl overflow-hidden flex flex-col max-h-[90vh]">
        <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
          <div>
            <h3 class="font-bold text-slate-900">{{ previewModalTitle }}</h3>
            <p v-if="previewModalSubtitle" class="text-sm text-slate-500">{{ previewModalSubtitle }}</p>
          </div>
          <button @click="closePreviewModal" class="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition-colors">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 overflow-auto flex-1 flex items-center justify-center bg-slate-100/50">
          <img :src="previewImageUrl" alt="Preview Bukti" class="max-w-full max-h-full rounded-xl shadow-sm border border-slate-200 object-contain" />
        </div>
      </div>
    </div>
  </div>
</template>
