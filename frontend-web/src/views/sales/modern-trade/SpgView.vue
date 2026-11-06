<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import {
  Users,
  Search,
  MapPin,
  CheckCircle2,
  AlertCircle,
  Plus,
  TrendingUp,
  Phone,
  Receipt,
  Calendar,
  DollarSign,
  Building2,
  Trash2,
  Edit2,
  Sparkles,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'

const masterDataStore = useMasterDataStore()

// Tabs: SPG Personnel, Schedules (Jumat-Minggu/Shift), Budgeting per Store
const activeTab = ref<'PERSONNEL' | 'SCHEDULES' | 'BUDGETING'>('PERSONNEL')

const searchQuery = ref('')
const selectedStatus = ref<'ALL' | 'ACTIVE' | 'INACTIVE'>('ALL')
const selectedScheduleFilter = ref<'ALL' | 'WEEKEND_FRI_SUN' | 'DAILY_MON_SAT' | 'CUSTOM'>('ALL')

const spgs = ref<any[]>([])
const keyAccounts = ref<any[]>([])
const schedules = ref<any[]>([])
const budgetSummary = ref<any>({
  store_budgets: [],
  total_overall_budget: 0,
  total_overall_target: 0,
})

const isLoading = ref(false)

const fetchSpgs = async () => {
  try {
    const res = await apiClient.get('/modern-trade/spg')
    spgs.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil data SPG: ' + (err.response?.data?.error || err.message))
  }
}

const fetchKeyAccounts = async () => {
  try {
    const res = await apiClient.get('/modern-trade/key-accounts')
    keyAccounts.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil data Key Account: ' + (err.response?.data?.error || err.message))
  }
}

const fetchSchedules = async () => {
  try {
    const res = await apiClient.get('/modern-trade/spg-schedules')
    schedules.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil jadwal SPG: ' + (err.response?.data?.error || err.message))
  }
}

const fetchBudgetSummary = async () => {
  try {
    const res = await apiClient.get('/modern-trade/spg-budget-summary')
    budgetSummary.value = res.data.data || {
      store_budgets: [],
      total_overall_budget: 0,
      total_overall_target: 0,
    }
  } catch (err: any) {
    console.error('Failed to fetch budget summary', err)
  }
}

const loadAllData = async () => {
  isLoading.value = true
  await Promise.all([fetchSpgs(), fetchKeyAccounts(), fetchSchedules(), fetchBudgetSummary()])
  isLoading.value = false
}

onMounted(() => {
  loadAllData()
})

// SPG Personnel Filters & Stats
const filteredSpgs = computed(() => {
  return spgs.value.filter((spg) => {
    const matchesSearch =
      spg.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      spg.supermarket_name.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesStatus = selectedStatus.value === 'ALL' || spg.status === selectedStatus.value
    return matchesSearch && matchesStatus
  })
})

const activeSpgCount = computed(() => spgs.value.filter(s => s.status === 'ACTIVE').length)
const totalTarget = computed(() => spgs.value.reduce((sum, s) => sum + Number(s.target_sales || 0), 0))
const totalActual = computed(() => spgs.value.reduce((sum, s) => sum + Number(s.actual_sales || 0), 0))
const salesAchievement = computed(() => totalTarget.value ? Math.round((totalActual.value / totalTarget.value) * 100) : 0)

// Schedules Filter
const filteredSchedules = computed(() => {
  return schedules.value.filter((sched) => {
    const matchesSearch =
      (sched.spg?.name || '').toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (sched.supermarket_name || '').toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (sched.key_account?.name || '').toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesPattern = selectedScheduleFilter.value === 'ALL' || sched.day_pattern === selectedScheduleFilter.value
    return matchesSearch && matchesPattern
  })
})

// Modal SPG Personnel Setup
const showModal = ref(false)
const isEditing = ref(false)
const form = ref({
  id: '',
  key_account_id: '',
  name: '',
  email: '',
  phone_number: '',
  shift_schedule: '',
  status: 'ACTIVE',
  target_sales: 0,
  actual_sales: 0,
  photo_url: '',
  supermarket_name: '',
})

const openNewModal = () => {
  isEditing.value = false
  form.value = {
    id: '',
    key_account_id: keyAccounts.value[0]?.id || '',
    name: '',
    email: '',
    phone_number: '',
    shift_schedule: 'Jumat, Sabtu, Minggu (10:00 - 19:00)',
    status: 'ACTIVE',
    target_sales: 0,
    actual_sales: 0,
    photo_url: '',
    supermarket_name: '',
  }
  showModal.value = true
}

const openEditModal = (item: any) => {
  isEditing.value = true
  form.value = {
    ...item,
    target_sales: Number(item.target_sales),
    actual_sales: Number(item.actual_sales),
  }
  showModal.value = true
}

const saveSpg = async () => {
  if (!form.value.name || !form.value.key_account_id || !form.value.email) {
    toast.error('Nama SPG, Email, dan Ritel wajib diisi!')
    return
  }

  try {
    const payload = {
      ...form.value,
      target_sales: Number(form.value.target_sales),
      actual_sales: Number(form.value.actual_sales),
    }

    if (isEditing.value) {
      await apiClient.put(`/modern-trade/spg/${form.value.id}`, payload)
      toast.success('Data SPG berhasil diperbarui!')
    } else {
      await apiClient.post('/modern-trade/spg', payload)
      toast.success('Data SPG baru berhasil ditambahkan!')
    }
    showModal.value = false
    fetchSpgs()
  } catch (err: any) {
    toast.error('Gagal menyimpan data SPG: ' + (err.response?.data?.error || err.message))
  }
}

const deleteSpg = async (id: string) => {
  if (confirm('Apakah Anda yakin ingin menghapus data SPG ini?')) {
    try {
      await apiClient.delete(`/modern-trade/spg/${id}`)
      toast.success('Data SPG berhasil dihapus!')
      fetchSpgs()
    } catch (err: any) {
      toast.error('Gagal menghapus data SPG: ' + (err.response?.data?.error || err.message))
    }
  }
}

// Modal Schedule & Budgeting Setup
const showScheduleModal = ref(false)
const isEditingSchedule = ref(false)
const scheduleForm = ref({
  id: '',
  spg_id: '',
  key_account_id: '',
  supermarket_name: '',
  start_date: new Date().toISOString().substring(0, 10),
  end_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().substring(0, 10),
  day_pattern: 'WEEKEND_FRI_SUN',
  daily_fee: 150000,
  allowance_fee: 25000,
  target_sales: 10000000,
  status: 'PLANNED',
  notes: '',
})

// Auto calculate working days in client for realtime feedback
const calculatedDaysCount = computed(() => {
  if (!scheduleForm.value.start_date || !scheduleForm.value.end_date) return 0
  const start = new Date(scheduleForm.value.start_date)
  const end = new Date(scheduleForm.value.end_date)
  if (start > end) return 0

  let days = 0
  const cur = new Date(start)
  while (cur <= end) {
    const day = cur.getDay() // 0 is Sunday, 5 is Friday, 6 is Saturday
    if (scheduleForm.value.day_pattern === 'WEEKEND_FRI_SUN') {
      if (day === 0 || day === 5 || day === 6) days++
    } else if (scheduleForm.value.day_pattern === 'DAILY_MON_SAT') {
      if (day !== 0) days++
    } else {
      days++
    }
    cur.setDate(cur.getDate() + 1)
  }
  return days > 0 ? days : 1
})

const estimatedTotalBudget = computed(() => {
  return calculatedDaysCount.value * (Number(scheduleForm.value.daily_fee || 0) + Number(scheduleForm.value.allowance_fee || 0))
})

const openNewScheduleModal = (prefilledSpg?: any) => {
  isEditingSchedule.value = false
  const now = new Date()
  const oneMonthLater = new Date(now.getFullYear(), now.getMonth() + 1, now.getDate())

  scheduleForm.value = {
    id: '',
    spg_id: prefilledSpg?.id || spgs.value[0]?.id || '',
    key_account_id: prefilledSpg?.key_account_id || keyAccounts.value[0]?.id || '',
    supermarket_name: prefilledSpg?.supermarket_name || '',
    start_date: now.toISOString().substring(0, 10),
    end_date: oneMonthLater.toISOString().substring(0, 10),
    day_pattern: 'WEEKEND_FRI_SUN',
    daily_fee: 150000,
    allowance_fee: 25000,
    target_sales: 12000000,
    status: 'PLANNED',
    notes: 'Jadwal promosi Weekend Booster (Jumat, Sabtu, Minggu)',
  }
  showScheduleModal.value = true
}

const openEditScheduleModal = (item: any) => {
  isEditingSchedule.value = true
  scheduleForm.value = {
    id: item.id,
    spg_id: item.spg_id,
    key_account_id: item.key_account_id,
    supermarket_name: item.supermarket_name,
    start_date: item.start_date ? item.start_date.substring(0, 10) : '',
    end_date: item.end_date ? item.end_date.substring(0, 10) : '',
    day_pattern: item.day_pattern || 'WEEKEND_FRI_SUN',
    daily_fee: Number(item.daily_fee),
    allowance_fee: Number(item.allowance_fee),
    target_sales: Number(item.target_sales),
    status: item.status || 'PLANNED',
    notes: item.notes || '',
  }
  showScheduleModal.value = true
}

const saveSchedule = async () => {
  if (!scheduleForm.value.spg_id || !scheduleForm.value.key_account_id || !scheduleForm.value.supermarket_name) {
    toast.error('Lengkapi personil SPG, Mitra Ritel, dan Nama Outlet Toko!')
    return
  }

  try {
    const payload = {
      ...scheduleForm.value,
      daily_fee: Number(scheduleForm.value.daily_fee),
      allowance_fee: Number(scheduleForm.value.allowance_fee),
      target_sales: Number(scheduleForm.value.target_sales),
    }

    if (isEditingSchedule.value) {
      await apiClient.put(`/modern-trade/spg-schedules/${scheduleForm.value.id}`, payload)
      toast.success('Jadwal penugasan & budget SPG berhasil diperbarui!')
    } else {
      await apiClient.post('/modern-trade/spg-schedules', payload)
      toast.success('Jadwal penugasan & alokasi budget toko berhasil dibuat!')
    }
    showScheduleModal.value = false
    fetchSchedules()
    fetchBudgetSummary()
  } catch (err: any) {
    toast.error('Gagal menyimpan jadwal: ' + (err.response?.data?.error || err.message))
  }
}

const deleteSchedule = async (id: string) => {
  if (confirm('Apakah Anda yakin ingin menghapus jadwal penugasan ini?')) {
    try {
      await apiClient.delete(`/modern-trade/spg-schedules/${id}`)
      toast.success('Jadwal penugasan berhasil dihapus!')
      fetchSchedules()
      fetchBudgetSummary()
    } catch (err: any) {
      toast.error('Gagal menghapus jadwal: ' + (err.response?.data?.error || err.message))
    }
  }
}

// Photo Upload
const isUploadingPhoto = ref(false)
const uploadPhoto = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  isUploadingPhoto.value = true
  try {
    const res = await apiClient.post('/modern-trade/merchandising/upload-photo', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    const baseUrl = import.meta.env.VITE_API_BASE_URL?.replace('/api', '') || ''
    form.value.photo_url = baseUrl + res.data.url
    toast.success('Foto SPG berhasil diunggah')
  } catch (err: any) {
    toast.error('Gagal mengunggah foto: ' + (err.response?.data?.error || err.message))
  } finally {
    isUploadingPhoto.value = false
    target.value = ''
  }
}

const formatCurrency = (val: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(val || 0)
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch (_) {
    return dateStr
  }
}

const getPatternLabel = (pat: string) => {
  switch (pat) {
    case 'WEEKEND_FRI_SUN':
      return 'Jumat, Sabtu, Minggu (Weekend Booster)'
    case 'DAILY_MON_SAT':
      return 'Senin s/d Sabtu (Full Time)'
    default:
      return 'Setiap Hari / Custom'
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="px-2.5 py-1 text-xs font-bold text-indigo-600 bg-indigo-50 border border-indigo-200/50 rounded-full">
            Modern Trade (MT)
          </span>
        </div>
        <h1 class="text-2xl md:text-3xl font-black text-slate-900 tracking-tight mt-1 flex items-center flex-wrap gap-2">
          <span>Kelola SPG, Penjadwalan & Budgeting</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-primary text-xl md:text-2xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-sm text-slate-500 mt-1">
          Atur data personel SPG, penugasan toko (Jumat, Sabtu, Minggu), alokasi budgeting, serta target sales outlet MT.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <RouterLink
          to="/sales/modern-trade/spg-reports"
          class="inline-flex items-center px-4 py-2 bg-white hover:bg-slate-50 text-slate-700 text-xs font-bold rounded-xl border border-slate-200 shadow-sm transition-all"
        >
          <Receipt class="w-4 h-4 mr-2 text-indigo-600" />
          Laporan Penjualan SPG
        </RouterLink>
        <Button
          v-if="activeTab === 'PERSONNEL'"
          @click="openNewModal"
          class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold rounded-xl shadow-md"
        >
          <Plus class="w-4 h-4 mr-2" />
          Tambah SPG Baru
        </Button>
        <Button
          v-else
          @click="openNewScheduleModal()"
          class="bg-emerald-600 hover:bg-emerald-700 text-white font-bold rounded-xl shadow-md"
        >
          <Calendar class="w-4 h-4 mr-2" />
          Assign Jadwal & Budget Toko
        </Button>
      </div>
    </div>

    <!-- Main Navigation Tabs -->
    <div class="flex border-b border-slate-200 bg-white px-3 pt-2 rounded-2xl shadow-xs gap-2">
      <button
        @click="activeTab = 'PERSONNEL'"
        class="flex items-center gap-2 px-4 py-3 text-xs font-bold border-b-2 transition-all"
        :class="activeTab === 'PERSONNEL' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-slate-500 hover:text-slate-800'"
      >
        <Users class="w-4 h-4" />
        Data Personel SPG ({{ spgs.length }})
      </button>

      <button
        @click="activeTab = 'SCHEDULES'"
        class="flex items-center gap-2 px-4 py-3 text-xs font-bold border-b-2 transition-all"
        :class="activeTab === 'SCHEDULES' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-slate-500 hover:text-slate-800'"
      >
        <Calendar class="w-4 h-4" />
        Jadwal & Penugasan Toko (Jum - Sab - Min) ({{ schedules.length }})
      </button>

      <button
        @click="activeTab = 'BUDGETING'"
        class="flex items-center gap-2 px-4 py-3 text-xs font-bold border-b-2 transition-all"
        :class="activeTab === 'BUDGETING' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-slate-500 hover:text-slate-800'"
      >
        <DollarSign class="w-4 h-4" />
        Monitoring Budgeting & Biaya Toko
      </button>
    </div>

    <!-- TAB 1: PERSONNEL -->
    <div v-if="activeTab === 'PERSONNEL'" class="space-y-6">
      <!-- Quick Stats -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
            <Users class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">SPG Aktif</p>
            <p class="text-2xl font-black text-slate-800">{{ activeSpgCount }} Personel</p>
          </div>
        </div>
        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="p-3 bg-indigo-50 text-indigo-600 rounded-xl">
            <TrendingUp class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Actual Sales</p>
            <p class="text-2xl font-black text-slate-800">{{ formatCurrency(totalActual) }}</p>
          </div>
        </div>
        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="p-3 bg-blue-50 text-blue-600 rounded-xl">
            <CheckCircle2 class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Achievement</p>
            <p class="text-2xl font-black text-slate-800">{{ salesAchievement }}% dari Target</p>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="relative w-full sm:w-[320px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari nama SPG atau outlet..."
            class="w-full pl-9 pr-4 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>

        <div class="flex items-center gap-2">
          <select
            v-model="selectedStatus"
            class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white"
          >
            <option value="ALL">Semua Status</option>
            <option value="ACTIVE">Aktif Bertugas</option>
            <option value="INACTIVE">Non-Aktif</option>
          </select>
        </div>
      </div>

      <!-- SPG Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        <div
          v-for="spg in filteredSpgs"
          :key="spg.id"
          class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden flex flex-col justify-between group"
        >
          <div class="p-5 flex gap-4 items-start">
            <img :src="spg.photo_url || 'https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?auto=format&fit=crop&q=80&w=200'" alt="Foto SPG" class="w-16 h-16 rounded-xl object-cover border border-slate-100 flex-shrink-0" />
            <div class="flex-1">
              <div class="flex justify-between items-start">
                <h3 class="font-extrabold text-slate-800 text-lg leading-tight">{{ spg.name }}</h3>
                <span
                  class="px-2 py-0.5 text-[10px] font-extrabold uppercase rounded-full tracking-wider"
                  :class="spg.status === 'ACTIVE' ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-500'"
                >
                  {{ spg.status === 'ACTIVE' ? 'Aktif' : 'Off' }}
                </span>
              </div>
              <p class="text-xs text-indigo-600 font-bold mt-0.5">{{ spg.key_account?.name || 'Ritel MT' }}</p>
              <div class="flex items-center gap-1.5 mt-2 text-xs text-slate-500">
                <MapPin class="w-3.5 h-3.5 text-slate-400" />
                <span class="truncate" :title="spg.supermarket_name">{{ spg.supermarket_name || '-' }}</span>
              </div>
              <div class="flex items-center gap-1.5 mt-1 text-xs text-slate-500">
                <Phone class="w-3.5 h-3.5 text-slate-400" />
                <span>{{ spg.phone_number || '-' }}</span>
              </div>
            </div>
          </div>

          <div class="px-5 pb-4">
            <div class="bg-slate-50 rounded-xl p-3 border border-slate-100">
              <span class="text-[10px] uppercase font-bold text-slate-400 block mb-1">Jadwal Shift / Event</span>
              <p class="text-xs font-semibold text-slate-700">{{ spg.shift_schedule || '-' }}</p>
            </div>
            
            <div class="mt-3 grid grid-cols-2 gap-3 text-xs">
              <div class="bg-indigo-50/50 p-2.5 rounded-xl border border-indigo-50 text-center">
                <span class="text-[10px] uppercase font-bold text-indigo-400 block mb-0.5">Target</span>
                <p class="font-black text-indigo-700">{{ formatCurrency(spg.target_sales) }}</p>
              </div>
              <div class="bg-emerald-50/50 p-2.5 rounded-xl border border-emerald-50 text-center">
                <span class="text-[10px] uppercase font-bold text-emerald-400 block mb-0.5">Actual</span>
                <p class="font-black text-emerald-700">{{ formatCurrency(spg.actual_sales) }}</p>
              </div>
            </div>
          </div>

          <div class="px-5 py-3 border-t border-slate-50 flex items-center justify-between gap-2 bg-slate-50/50">
            <button
              @click="openNewScheduleModal(spg)"
              class="px-2.5 py-1.5 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 rounded-lg transition-all text-xs font-bold flex items-center gap-1.5"
              title="Assign Toko & Jadwal Weekend"
            >
              <Calendar class="w-3.5 h-3.5" />
              Jadwalkan
            </button>
            <div class="flex items-center gap-1.5">
              <button
                @click="openEditModal(spg)"
                class="px-2.5 py-1.5 hover:bg-indigo-50 text-slate-500 hover:text-indigo-600 rounded-lg transition-all text-xs font-bold flex items-center gap-1"
              >
                <Edit2 class="w-3.5 h-3.5" />
                Edit
              </button>
              <button
                @click="deleteSpg(spg.id)"
                class="px-2.5 py-1.5 hover:bg-rose-50 text-slate-500 hover:text-rose-600 rounded-lg transition-all text-xs font-bold flex items-center gap-1"
              >
                <Trash2 class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 2: SCHEDULES & STORE ASSIGNMENT -->
    <div v-if="activeTab === 'SCHEDULES'" class="space-y-6">
      <!-- Filters & Pattern Filter -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="relative w-full sm:w-[320px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari SPG, Toko, atau Ritel..."
            class="w-full pl-9 pr-4 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>

        <div class="flex items-center gap-2">
          <select
            v-model="selectedScheduleFilter"
            class="px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-indigo-500 bg-white"
          >
            <option value="ALL">Semua Pola Jadwal</option>
            <option value="WEEKEND_FRI_SUN">Jumat, Sabtu, Minggu (Weekend Booster)</option>
            <option value="DAILY_MON_SAT">Senin - Sabtu (Reguler)</option>
            <option value="CUSTOM">Custom / Event Lainnya</option>
          </select>
        </div>
      </div>

      <!-- Schedule Table -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-50/80 text-slate-400 uppercase font-black tracking-wider text-[10px] border-b border-slate-100">
              <tr>
                <th class="py-3.5 px-4">Personil SPG</th>
                <th class="py-3.5 px-4">Toko / Outlet Bertugas</th>
                <th class="py-3.5 px-4">Periode Tanggal</th>
                <th class="py-3.5 px-4">Pola Hari & Hari Kerja</th>
                <th class="py-3.5 px-4">Total Budget SPG</th>
                <th class="py-3.5 px-4">Target Sales Toko</th>
                <th class="py-3.5 px-4 text-center">Status</th>
                <th class="py-3.5 px-4 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 text-slate-700">
              <tr v-if="filteredSchedules.length === 0">
                <td colspan="8" class="text-center py-8 text-slate-400 font-semibold">
                  Belum ada jadwal penugasan SPG. Klik "Assign Jadwal & Budget Toko" di atas.
                </td>
              </tr>
              <tr v-for="sched in filteredSchedules" :key="sched.id" class="hover:bg-slate-50/50 transition-colors">
                <td class="py-3.5 px-4 font-bold text-slate-900">
                  <div class="flex items-center gap-2">
                    <span class="w-2 h-2 rounded-full" :class="sched.status === 'ACTIVE' ? 'bg-emerald-500' : 'bg-amber-400'"></span>
                    <span>{{ sched.spg?.name || 'SPG Karsa HR' }}</span>
                  </div>
                </td>
                <td class="py-3.5 px-4">
                  <div class="font-bold text-slate-800">{{ sched.supermarket_name }}</div>
                  <div class="text-[11px] text-indigo-600 font-semibold">{{ sched.key_account?.name }}</div>
                </td>
                <td class="py-3.5 px-4">
                  <div class="font-semibold">{{ formatDate(sched.start_date) }} - {{ formatDate(sched.end_date) }}</div>
                </td>
                <td class="py-3.5 px-4">
                  <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-bold"
                    :class="sched.day_pattern === 'WEEKEND_FRI_SUN' ? 'bg-amber-50 text-amber-700 border border-amber-200/60' : 'bg-blue-50 text-blue-700 border border-blue-200/60'">
                    {{ sched.day_pattern === 'WEEKEND_FRI_SUN' ? 'Jumat, Sabtu, Minggu' : (sched.day_pattern === 'DAILY_MON_SAT' ? 'Senin - Sabtu' : 'Custom') }}
                  </span>
                  <div class="text-[10px] text-slate-400 font-bold mt-0.5">{{ sched.working_days_count }} Hari Kerja</div>
                </td>
                <td class="py-3.5 px-4">
                  <div class="font-extrabold text-slate-900">{{ formatCurrency(sched.total_budget) }}</div>
                  <div class="text-[10px] text-slate-400">Rate: {{ formatCurrency(sched.daily_fee + sched.allowance_fee) }}/hari</div>
                </td>
                <td class="py-3.5 px-4">
                  <div class="font-extrabold text-emerald-600">{{ formatCurrency(sched.target_sales) }}</div>
                </td>
                <td class="py-3.5 px-4 text-center">
                  <span
                    class="px-2 py-0.5 text-[10px] font-bold rounded-full uppercase"
                    :class="{
                      'bg-amber-50 text-amber-700': sched.status === 'PLANNED',
                      'bg-emerald-50 text-emerald-700': sched.status === 'ACTIVE',
                      'bg-slate-100 text-slate-600': sched.status === 'COMPLETED',
                      'bg-rose-50 text-rose-700': sched.status === 'CANCELLED',
                    }"
                  >
                    {{ sched.status }}
                  </span>
                </td>
                <td class="py-3.5 px-4 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <button
                      @click="openEditScheduleModal(sched)"
                      class="p-1.5 hover:bg-indigo-50 text-slate-400 hover:text-indigo-600 rounded-lg transition-all"
                      title="Edit Jadwal"
                    >
                      <Edit2 class="w-3.5 h-3.5" />
                    </button>
                    <button
                      @click="deleteSchedule(sched.id)"
                      class="p-1.5 hover:bg-rose-50 text-slate-400 hover:text-rose-600 rounded-lg transition-all"
                      title="Hapus Jadwal"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 3: BUDGETING MONITORING -->
    <div v-if="activeTab === 'BUDGETING'" class="space-y-6">
      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="p-3 bg-amber-50 text-amber-600 rounded-xl">
            <DollarSign class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Alokasi Budget SPG</p>
            <p class="text-2xl font-black text-slate-900">{{ formatCurrency(budgetSummary.total_overall_budget) }}</p>
          </div>
        </div>
        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
            <TrendingUp class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Target Omset Toko Terjadwal</p>
            <p class="text-2xl font-black text-emerald-600">{{ formatCurrency(budgetSummary.total_overall_target) }}</p>
          </div>
        </div>
        <div class="bg-white p-5 rounded-2xl border border-slate-100 shadow-sm flex items-center gap-4">
          <div class="p-3 bg-indigo-50 text-indigo-600 rounded-xl">
            <Building2 class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Toko Ditempati SPG</p>
            <p class="text-2xl font-black text-slate-900">{{ budgetSummary.store_budgets?.length || 0 }} Outlet</p>
          </div>
        </div>
      </div>

      <!-- Store Budget Table -->
      <div class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
        <div class="p-5 border-b border-slate-100 flex items-center justify-between">
          <h3 class="font-extrabold text-slate-800 text-base">Alokasi Biaya SPG per Outlet Toko MT</h3>
          <span class="text-xs text-slate-400 font-semibold">Estimasi ROI & Cost to Sales</span>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-50/80 text-slate-400 uppercase font-black tracking-wider text-[10px] border-b border-slate-100">
              <tr>
                <th class="py-3.5 px-4">Nama Toko Outlet</th>
                <th class="py-3.5 px-4">Key Account Ritel</th>
                <th class="py-3.5 px-4">Jumlah Penugasan</th>
                <th class="py-3.5 px-4">Total Biaya Budget SPG</th>
                <th class="py-3.5 px-4">Target Penjualan</th>
                <th class="py-3.5 px-4">Rasio Biaya (Cost / Sales)</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 text-slate-700">
              <tr v-if="!budgetSummary.store_budgets || budgetSummary.store_budgets.length === 0">
                <td colspan="6" class="text-center py-8 text-slate-400 font-semibold">
                  Belum ada data budgeting penugasan toko.
                </td>
              </tr>
              <tr v-for="sb in budgetSummary.store_budgets" :key="sb.supermarket_name" class="hover:bg-slate-50/50">
                <td class="py-3.5 px-4 font-extrabold text-slate-900">{{ sb.supermarket_name }}</td>
                <td class="py-3.5 px-4 font-bold text-indigo-600">{{ sb.key_account_name }}</td>
                <td class="py-3.5 px-4 font-semibold">{{ sb.total_schedules }} Penugasan</td>
                <td class="py-3.5 px-4 font-black text-amber-600">{{ formatCurrency(sb.total_budget) }}</td>
                <td class="py-3.5 px-4 font-black text-emerald-600">{{ formatCurrency(sb.total_target) }}</td>
                <td class="py-3.5 px-4 font-bold text-slate-700">
                  <span v-if="sb.total_target > 0" class="px-2 py-0.5 rounded-md text-[11px]" :class="((sb.total_budget / sb.total_target) * 100) > 20 ? 'bg-rose-50 text-rose-700' : 'bg-emerald-50 text-emerald-700'">
                    {{ ((sb.total_budget / sb.total_target) * 100).toFixed(1) }}%
                  </span>
                  <span v-else class="text-slate-400">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- MODAL 1: Form Data Personel SPG -->
    <div
      v-if="showModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white rounded-[24px] w-full max-w-xl shadow-2xl overflow-hidden p-6 relative my-8">
        <h3 class="text-lg font-black text-slate-900 mb-4">
          {{ isEditing ? 'Edit Data SPG' : 'Tambah SPG Baru' }}
        </h3>

        <div class="space-y-4 text-xs font-semibold text-slate-650 max-h-[65vh] overflow-y-auto pr-1">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Nama Lengkap</label>
              <input
                v-model="form.name"
                type="text"
                placeholder="Contoh: Siska Nirmala"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Email Akun (Login)</label>
              <input
                v-model="form.email"
                type="email"
                placeholder="Contoh: siska@demo.local"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Ritel MT (Key Account)</label>
              <select
                v-model="form.key_account_id"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white"
              >
                <option v-for="ka in keyAccounts" :key="ka.id" :value="ka.id">
                  {{ ka.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">No. WhatsApp / HP</label>
              <input
                v-model="form.phone_number"
                type="text"
                placeholder="Contoh: 08123456789"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Nama Outlet / Supermarket</label>
              <input
                v-model="form.supermarket_name"
                type="text"
                placeholder="Contoh: Hypermart Karawaci"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Jadwal Shift / Default</label>
              <input
                v-model="form.shift_schedule"
                type="text"
                placeholder="Contoh: Jumat, Sabtu, Minggu (10:00 - 19:00)"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Status Kepegawaian</label>
              <select
                v-model="form.status"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-indigo-500 focus:outline-none bg-white"
              >
                <option value="ACTIVE">Aktif Bertugas</option>
                <option value="INACTIVE">Non-Aktif (Off)</option>
              </select>
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Foto SPG / ID Card</label>
              <input
                type="file"
                accept="image/*"
                @change="uploadPhoto"
                class="w-full text-xs text-slate-500 file:mr-2 file:py-1 file:px-3 file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
              />
            </div>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-3">
          <Button
            @click="showModal = false"
            variant="outline"
            class="px-4 py-2 border border-slate-200 rounded-xl text-xs text-slate-500 font-bold hover:bg-slate-50"
          >
            Batal
          </Button>
          <Button
            @click="saveSpg"
            class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold"
          >
            Simpan Data SPG
          </Button>
        </div>
      </div>
    </div>

    <!-- MODAL 2: Form Assign Jadwal & Budgeting SPG -->
    <div
      v-if="showScheduleModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white rounded-[24px] w-full max-w-xl shadow-2xl overflow-hidden p-6 relative my-8">
        <div class="flex items-center gap-2 mb-4">
          <div class="p-2 bg-emerald-50 text-emerald-600 rounded-xl">
            <Calendar class="w-5 h-5" />
          </div>
          <div>
            <h3 class="text-lg font-black text-slate-900">
              {{ isEditingSchedule ? 'Edit Jadwal & Budget SPG' : 'Assign Jadwal SPG & Budget Toko' }}
            </h3>
            <p class="text-xs text-slate-500">Atur hari tugas (Jumat-Minggu/Shift) dan estimasi alokasi biaya toko.</p>
          </div>
        </div>

        <div class="space-y-4 text-xs font-semibold text-slate-650 max-h-[65vh] overflow-y-auto pr-1">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Pilih Personil SPG</label>
              <select
                v-model="scheduleForm.spg_id"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none bg-white"
              >
                <option v-for="spg in spgs" :key="spg.id" :value="spg.id">
                  {{ spg.name }} ({{ spg.supermarket_name || 'Outlet' }})
                </option>
              </select>
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Key Account Ritel</label>
              <select
                v-model="scheduleForm.key_account_id"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none bg-white"
              >
                <option v-for="ka in keyAccounts" :key="ka.id" :value="ka.id">
                  {{ ka.name }}
                </option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Nama Toko Outlet yang Ditempati</label>
            <input
              v-model="scheduleForm.supermarket_name"
              type="text"
              placeholder="Contoh: Yogya Riau Junction Bandung"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Tanggal Mulai</label>
              <input
                v-model="scheduleForm.start_date"
                type="date"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Tanggal Berakhir</label>
              <input
                v-model="scheduleForm.end_date"
                type="date"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none"
              />
            </div>
          </div>

          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Pola Hari Tugas</label>
            <select
              v-model="scheduleForm.day_pattern"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none bg-amber-50/50 font-bold text-slate-800"
            >
              <option value="WEEKEND_FRI_SUN">⭐ Jumat, Sabtu, Minggu (Weekend Booster)</option>
              <option value="DAILY_MON_SAT">Senin - Sabtu (Reguler Full Time)</option>
              <option value="CUSTOM">Setiap Hari / Custom Event</option>
            </select>
          </div>

          <!-- Live Budget Calculation Box -->
          <div class="bg-slate-50 p-4 rounded-2xl border border-slate-200 space-y-3">
            <div class="flex items-center justify-between text-[11px] font-bold text-slate-700">
              <span class="flex items-center gap-1.5 text-indigo-600">
                <Sparkles class="w-4 h-4" />
                Kalkulasi Otomatis Hari & Budget Toko:
              </span>
              <span class="bg-indigo-50 text-indigo-700 px-2 py-0.5 rounded-md font-black">
                {{ calculatedDaysCount }} Hari Tugas
              </span>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Honor SPG / Hari</label>
                <input
                  v-model.number="scheduleForm.daily_fee"
                  type="number"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none bg-white font-bold"
                />
              </div>
              <div>
                <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Uang Saku / Transport / Hari</label>
                <input
                  v-model.number="scheduleForm.allowance_fee"
                  type="number"
                  class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none bg-white font-bold"
                />
              </div>
            </div>

            <div class="pt-2 border-t border-slate-200 flex items-center justify-between">
              <span class="text-xs font-bold text-slate-500">Estimasi Total Biaya SPG:</span>
              <span class="text-base font-black text-amber-600">{{ formatCurrency(estimatedTotalBudget) }}</span>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Target Penjualan Toko (IDR)</label>
              <input
                v-model.number="scheduleForm.target_sales"
                type="number"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none font-bold text-emerald-600"
              />
            </div>
            <div>
              <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Status Penugasan</label>
              <select
                v-model="scheduleForm.status"
                class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none bg-white"
              >
                <option value="PLANNED">PLANNED (Direncanakan)</option>
                <option value="ACTIVE">ACTIVE (Sedang Berjalan)</option>
                <option value="COMPLETED">COMPLETED (Selesai)</option>
                <option value="CANCELLED">CANCELLED (Dibatalkan)</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-[10px] uppercase font-bold text-slate-400 mb-1">Catatan Tambahan / Event</label>
            <textarea
              v-model="scheduleForm.notes"
              rows="2"
              placeholder="Contoh: SPG Weekend Promo Diskon 10% + Sampling Toples"
              class="w-full px-3 py-2 border border-slate-200 rounded-xl focus:ring-1 focus:ring-emerald-500 focus:outline-none text-xs"
            ></textarea>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-3">
          <Button
            @click="showScheduleModal = false"
            variant="outline"
            class="px-4 py-2 border border-slate-200 rounded-xl text-xs text-slate-500 font-bold hover:bg-slate-50"
          >
            Batal
          </Button>
          <Button
            @click="saveSchedule"
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold"
          >
            Simpan Jadwal & Budget
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
