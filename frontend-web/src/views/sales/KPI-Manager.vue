<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import apiClient from '@/api/axios'
import {
  Target,
  TrendingUp,
  Plus,
  Search,
  Edit3,
  BarChart3,
  Award,
  ChevronRight,
  Filter,
  Download,
  AlertCircle,
  CheckCircle2,
  X,
  Save,
  Trash2,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { toast } from 'vue-sonner'
import * as XLSX from 'xlsx'

import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'

const masterStore = useMasterDataStore()
const authStore = useAuthStore()
const canManageTargets = computed(() => {
  const role = authStore.user?.role
  return ['superadmin', 'manager', 'supervisor_sales', 'supervisor_sales_mt', 'hr_admin'].includes(role || '')
})
const isLoading = ref(true)
const showManualEntry = ref(false)
const showEditTarget = ref(false)
const searchQuery = ref('')

const now = new Date()
const currentMonth = ref(now.getMonth() + 1)
const currentYear = ref(now.getFullYear())

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

const years = Array.from({ length: 5 }, (_, i) => now.getFullYear() - 2 + i)

const kpiSummary = ref({
  totalTarget: 0,
  totalAchieved: 0,
  activeSalesmen: 0,
  averageAchievement: 0,
})

const salesmenKpi = ref<any[]>([])

const filteredSalesmenKpi = computed(() => {
  if (!searchQuery.value) return salesmenKpi.value
  const q = searchQuery.value.toLowerCase()
  return salesmenKpi.value.filter(
    (s) =>
      s.name.toLowerCase().includes(q) ||
      s.role.toLowerCase().includes(q) ||
      (s.branch && s.branch.toLowerCase().includes(q)),
  )
})

const filterMode = ref<'MONTH' | 'DATE'>('MONTH')
const selectedDate = ref(new Date().toISOString().split('T')[0])

const fetchKPIs = async () => {
  isLoading.value = true
  const month = currentMonth.value
  const year = currentYear.value
  const branchId = masterStore.selectedBranchId

  let query =
    filterMode.value === 'DATE' ? `?date=${selectedDate.value}` : `?month=${month}&year=${year}`
  if (branchId && branchId !== 'ALL' && branchId !== 'ALL_BRANCHES') {
    query += `&branch_id=${branchId}`
  }
  if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
    query += `&sales_type=${masterStore.selectedSalesType}`
  }

  try {
    // 2. Fetch Performance List
    const perfRes = await apiClient.get(`/admin/sales/reports/performance${query}`)
    let rawKpis = perfRes.data?.data || []

    // Fallback: If no data for current month, try fetching previous month's targets
    if (rawKpis.length === 0 && filterMode.value === 'MONTH') {
      let prevMonth = month - 1
      let prevYear = year
      if (prevMonth === 0) {
        prevMonth = 12
        prevYear -= 1
      }
      let prevQuery = `?month=${prevMonth}&year=${prevYear}`
      if (branchId && branchId !== 'ALL' && branchId !== 'ALL_BRANCHES') {
        prevQuery += `&branch_id=${branchId}`
      }
      if (masterStore.selectedSalesType && masterStore.selectedSalesType !== 'ALL_SALES') {
        prevQuery += `&sales_type=${masterStore.selectedSalesType}`
      }

      const prevRes = await apiClient.get(`/admin/sales/reports/performance${prevQuery}`)
      if (prevRes.data?.data && prevRes.data.data.length > 0) {
        // Reset actuals to 0 but keep the target
        rawKpis = prevRes.data.data.map((k: any) => ({
          ...k,
          ID: '', // Make it blank so it can't be deleted as it's just a fallback
          AchievedOmzetLama: 0,
          AchievedOmzetBaru: 0,
          AchievedNewStores: 0,
          TotalVisits: 0,
          EstimatedBonus: 0,
        }))
      }
    }

    if (rawKpis) {
      let kpis = rawKpis.map((kpi: any) => ({
        id: kpi.ID,
        name: kpi.Employee ? `${kpi.Employee.first_name} ${kpi.Employee.last_name}` : 'Salesman',
        role: kpi.Employee?.job_position?.title || 'Sales',
        branch: kpi.Employee?.branch?.name || 'Pusat',
        target: filterMode.value === 'DATE' ? kpi.TargetOmzet / 30 : kpi.TargetOmzet, // Approx daily target
        achievedLama: kpi.AchievedOmzetLama,
        achievedBaru: kpi.AchievedOmzetBaru,
        totalAchieved: kpi.AchievedOmzetLama + kpi.AchievedOmzetBaru,
        targetNewStores: kpi.TargetNewStores,
        achievedNewStores: kpi.AchievedNewStores,
        targetDailyVisits: kpi.target_daily_visits || 0,
        targetDailyNewStores: kpi.target_daily_new_stores || 0,
        totalVisits: kpi.TotalVisits || 0,
        employeeID: kpi.EmployeeID,
        bonus: kpi.EstimatedBonus,
      }))

      // Sort ranking omzet penjualan descending
      kpis.sort((a: any, b: any) => b.totalAchieved - a.totalAchieved)

      salesmenKpi.value = kpis

      kpiSummary.value.activeSalesmen = salesmenKpi.value.length
      const totalTarget = salesmenKpi.value.reduce((acc: number, curr: any) => acc + curr.target, 0)
      const totalAchieved = salesmenKpi.value.reduce((acc: number, curr: any) => acc + curr.totalAchieved, 0)

      kpiSummary.value.totalTarget = totalTarget
      kpiSummary.value.totalAchieved = totalAchieved
      kpiSummary.value.averageAchievement =
        totalTarget > 0 ? Math.round((totalAchieved / totalTarget) * 100) : 0
    }
  } catch (error) {
    console.error('Failed to fetch KPIs:', error)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  masterStore.fetchBranches()
  fetchKPIs()
})

watch(
  () => masterStore.selectedBranchId,
  () => {
    fetchKPIs()
  },
)

watch(
  () => masterStore.selectedSalesType,
  () => {
    fetchKPIs()
  },
)

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

const getAchievementPercent = (achieved: number, target: number) => {
  return Math.round((achieved / target) * 100)
}

const manualEntryForm = ref({
  employeeID: '',
  storeCategory: 'TOKO_LAMA',
  totalAmount: 0,
  transactionDate: new Date().toISOString().split('T')[0],
  notes: '',
})

const submitManualEntry = async () => {
  if (!manualEntryForm.value.employeeID) {
    toast.error('Pilih salesman terlebih dahulu')
    return
  }

  try {
    isLoading.value = true
    await apiClient.post('/admin/sales/manual-entry', {
      employee_id: manualEntryForm.value.employeeID,
      store_category: manualEntryForm.value.storeCategory,
      total_amount: manualEntryForm.value.totalAmount,
      transaction_date: manualEntryForm.value.transactionDate,
      store_id: '00000000-0000-0000-0000-000000000000', // Mocking store for manual entry
      company_id: 'b72883e2-8bb6-438d-92bd-391edbf9e325', // Default company
    })

    toast.success('Data berhasil disimpan dan KPI diperbarui otomatis!')
    showManualEntry.value = false
    fetchKPIs() // Refresh list
  } catch (error) {
    console.error('Failed to submit manual entry:', error)
    toast.error('Gagal menyimpan data manual')
  } finally {
    isLoading.value = false
  }
}

const editForm = ref({
  employeeID: '',
  name: '',
  targetOmzet: 0,
  targetNewStores: 0,
  targetDailyVisits: 0,
  targetDailyNewStores: 0,
  month: 0,
  year: 0,
})

const handleEditTarget = (s: any) => {
  editForm.value = {
    employeeID: s.employeeID || s.id,
    name: s.name,
    targetOmzet: s.target,
    targetNewStores: s.targetNewStores,
    targetDailyVisits: s.targetDailyVisits || 0,
    targetDailyNewStores: s.targetDailyNewStores || 0,
    month: currentMonth.value,
    year: currentYear.value,
  }
  showEditTarget.value = true
}

const submitEditTarget = async () => {
  try {
    isLoading.value = true
    await apiClient.post('/admin/sales/targets', {
      employee_id: editForm.value.employeeID,
      month: editForm.value.month,
      year: editForm.value.year,
      target_omzet: editForm.value.targetOmzet,
      target_new_stores: editForm.value.targetNewStores,
      target_daily_visits: editForm.value.targetDailyVisits,
      target_daily_new_stores: editForm.value.targetDailyNewStores,
    })
    toast.success('Target berhasil diperbarui!')
    showEditTarget.value = false
    fetchKPIs()
  } catch (error) {
    toast.error('Gagal memperbarui target')
  } finally {
    isLoading.value = false
  }
}
const handleDeleteKPI = async (id: string) => {
  if (!id) {
    toast.error('Data target tidak ditemukan')
    return
  }
  if (!confirm('Hapus data target/KPI untuk salesman ini pada bulan terpilih?')) return
  try {
    isLoading.value = true
    await apiClient.delete(`/admin/sales/targets/${id}`)
    toast.success('Target berhasil dihapus!')
    fetchKPIs()
  } catch (error) {
    toast.error('Gagal menghapus target')
  } finally {
    isLoading.value = false
  }
}

const isExporting = ref(false)
const exportToExcel = () => {
  if (isExporting.value) return
  isExporting.value = true
  
  try {
    if (salesmenKpi.value.length === 0) {
      toast.error('Tidak ada data KPI untuk di-export')
      return
    }

    const dataToExport = salesmenKpi.value.map(s => {
      return {
        'Nama Salesman': s.name,
        'Jabatan': s.role,
        'Target Omzet': s.target,
        'Omzet Lama': s.achievedLama,
        'Omzet Baru': s.achievedBaru,
        'Total Realisasi': s.totalAchieved,
        'Pencapaian Omzet (%)': s.target > 0 ? Math.round((s.totalAchieved / s.target) * 100) : 0,
        'Target Toko Baru (Titik)': s.targetNewStores,
        'Realisasi Toko Baru': s.achievedNewStores,
        'Target Harian Kunjungan': s.targetDailyVisits,
        'Target Harian Toko Baru': s.targetDailyNewStores,
        'Total Kunjungan': s.totalVisits,
        'Estimasi Bonus': s.bonus
      }
    })

    const ws = XLSX.utils.json_to_sheet(dataToExport)
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, "KPI Sales")

    const filterName = filterMode.value === 'DATE' ? selectedDate.value : `${months.find(m => m.value === currentMonth.value)?.label}_${currentYear.value}`
    XLSX.writeFile(wb, `Laporan_KPI_Sales_${filterName}.xlsx`)
    toast.success('Data berhasil di-export ke Excel')
  } catch (error) {
    console.error('Export failed:', error)
    toast.error('Gagal meng-export data')
  } finally {
    isExporting.value = false
  }
}
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-6">
      <div>
        <h1 class="text-3xl font-extrabold text-slate-900 tracking-tight flex items-center flex-wrap gap-3">
          <Target class="w-8 h-8 text-primary" />
          <span>KPI Manager</span>
          <span v-if="masterStore.selectedBranchName" class="text-primary text-2xl font-bold">
            — {{ masterStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-[15px] font-medium text-slate-500 mt-1.5 flex items-center gap-2">
          <Award class="w-4 h-4 text-amber-500" />
          Atur target dan pantau performa insentif tim sales.
        </p>
      </div>
      <div v-if="canManageTargets" class="flex items-center gap-3">
        <Button
          @click="showManualEntry = true"
          class="bg-primary hover:bg-primary/90 text-white px-6 py-6 rounded-2xl shadow-lg shadow-primary/20 flex items-center gap-2 font-bold transition-all active:scale-95"
        >
          <Plus class="w-5 h-5" /> Input Nota Manual
        </Button>
      </div>
    </div>

    <!-- Summary Stats -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div
        class="bg-white p-6 rounded-3xl border border-slate-200 shadow-sm relative overflow-hidden group"
      >
        <div class="relative z-10">
          <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest">
            Total Target Divisi
          </p>
          <h3 class="text-2xl font-black text-slate-900 mt-2">
            {{ formatCurrency(kpiSummary.totalTarget) }}
          </h3>
          <div class="mt-4 h-1.5 w-full bg-slate-100 rounded-full overflow-hidden">
            <div
              class="h-full bg-primary rounded-full transition-all duration-1000"
              :style="`width: ${getAchievementPercent(kpiSummary.totalAchieved, kpiSummary.totalTarget)}%`"
            ></div>
          </div>
        </div>
        <TrendingUp
          class="absolute -right-4 -bottom-4 w-24 h-24 text-slate-50 group-hover:text-primary/5 transition-colors duration-500"
        />
      </div>

      <div class="bg-white p-6 rounded-3xl border border-slate-200 shadow-sm">
        <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest">
          Realisasi Saat Ini
        </p>
        <h3 class="text-2xl font-black text-emerald-600 mt-2">
          {{ formatCurrency(kpiSummary.totalAchieved) }}
        </h3>
        <p class="text-[11px] font-bold text-slate-400 mt-4 flex items-center gap-1">
          <CheckCircle2 class="w-3 h-3 text-emerald-500" />
          {{ getAchievementPercent(kpiSummary.totalAchieved, kpiSummary.totalTarget) }}% dari target
          bulan ini
        </p>
      </div>

      <div class="bg-white p-6 rounded-3xl border border-slate-200 shadow-sm">
        <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest">
          Salesman Aktif
        </p>
        <h3 class="text-2xl font-black text-slate-900 mt-2">
          {{ kpiSummary.activeSalesmen }} Orang
        </h3>
        <p class="text-[11px] font-bold text-slate-400 mt-4">Semua dalam pantauan live</p>
      </div>

      <div class="bg-white p-6 rounded-3xl border border-slate-200 shadow-sm">
        <p class="text-[11px] font-black text-slate-400 uppercase tracking-widest">
          Rata-rata Capaian
        </p>
        <h3 class="text-2xl font-black text-primary mt-2">{{ kpiSummary.averageAchievement }}%</h3>
        <div class="mt-4 flex gap-1">
          <div
            v-for="i in 5"
            :key="i"
            class="h-1.5 flex-1 rounded-full"
            :class="i <= 4 ? 'bg-primary' : 'bg-slate-100'"
          ></div>
        </div>
      </div>
    </div>

    <!-- Main List Section -->
    <div
      class="bg-white rounded-[32px] border border-slate-200 shadow-sm overflow-hidden flex flex-col"
    >
      <!-- Toolbar -->
      <div
        class="p-6 border-b border-slate-100 flex flex-col md:flex-row gap-4 items-center justify-between bg-slate-50/30"
      >
        <div class="relative w-full md:w-96">
          <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari salesman..."
            class="w-full pl-10 pr-4 py-2.5 bg-white border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary/10 transition-all font-bold text-sm"
          />
        </div>
        <div class="flex items-center gap-3 w-full md:w-auto">
          <select
            v-model="filterMode"
            @change="fetchKPIs"
            class="flex-1 md:w-32 px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-black text-slate-600 focus:ring-2 focus:ring-primary/10 transition-all appearance-none"
          >
            <option value="MONTH">Bulan</option>
            <option value="DATE">Harian</option>
          </select>

          <template v-if="filterMode === 'MONTH'">
            <select
              v-model="currentMonth"
              @change="fetchKPIs"
              class="flex-1 md:w-32 px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-black text-slate-600 focus:ring-2 focus:ring-primary/10 transition-all appearance-none"
            >
              <option v-for="m in months" :key="m.value" :value="m.value">{{ m.label }}</option>
            </select>
            <select
              v-model="currentYear"
              @change="fetchKPIs"
              class="flex-1 md:w-24 px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-black text-slate-600 focus:ring-2 focus:ring-primary/10 transition-all appearance-none"
            >
              <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
            </select>
          </template>

          <template v-else>
            <input
              v-model="selectedDate"
              @change="fetchKPIs"
              type="date"
              class="flex-1 md:w-40 px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-black text-slate-600 focus:ring-2 focus:ring-primary/10 transition-all appearance-none"
            />
          </template>

          <button
            @click="exportToExcel"
            :disabled="isExporting"
            class="flex-1 md:flex-none flex items-center justify-center gap-2 px-4 py-2.5 bg-white border border-slate-200 rounded-xl text-xs font-black text-slate-600 hover:bg-slate-50 transition-all disabled:opacity-50"
          >
            <Download class="w-3.5 h-3.5" /> {{ isExporting ? 'Exporting...' : 'Export' }}
          </button>
        </div>
      </div>

      <!-- Table Content -->
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr
              class="bg-slate-50/80 uppercase text-[10px] font-black text-slate-400 tracking-[0.15em]"
            >
              <th class="px-8 py-5">Nama / Jabatan</th>
              <th class="px-8 py-5">Target Omzet</th>
              <th class="px-8 py-5">Omzet Lama</th>
              <th class="px-8 py-5">Omzet Baru</th>
              <th class="px-8 py-5">Total Realisasi</th>
              <th class="px-8 py-5">Toko Baru (Titik)</th>
              <th class="px-8 py-5">Kunjungan Sales</th>
              <th class="px-8 py-5">Estimasi Bonus</th>
              <th class="px-8 py-5 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="isLoading" v-for="i in 4" :key="i" class="animate-pulse">
              <td v-for="j in 9" :key="j" class="px-8 py-6">
                <div class="h-4 bg-slate-100 rounded w-full"></div>
              </td>
            </tr>

            <tr v-else-if="filteredSalesmenKpi.length === 0">
              <td colspan="9" class="py-16 text-center text-slate-400">
                <Target class="w-10 h-10 mx-auto mb-2 text-slate-300" />
                <p class="font-bold text-sm">Tidak ada data target/performa sales pada periode ini.</p>
              </td>
            </tr>

            <tr
              v-else
              v-for="s in filteredSalesmenKpi"
              :key="s.employeeID || s.id"
              class="group hover:bg-slate-50/50 transition-all"
            >
              <td class="px-8 py-6">
                <div class="flex items-center gap-4">
                  <div
                    class="w-10 h-10 rounded-xl bg-slate-100 flex items-center justify-center font-black text-slate-500 group-hover:bg-primary/10 group-hover:text-primary transition-colors"
                  >
                    {{ s.name.charAt(0) }}
                  </div>
                  <div>
                    <p class="font-extrabold text-slate-900 leading-none">{{ s.name }}</p>
                    <div class="flex items-center gap-2 mt-1.5 flex-wrap">
                      <span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">
                        {{ s.role }}
                      </span>
                      <span class="text-slate-300">•</span>
                      <span class="text-[10px] font-bold text-primary bg-primary/5 px-2 py-0.5 rounded-md">
                        {{ s.branch }}
                      </span>
                    </div>
                  </div>
                </div>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-black text-slate-900">{{ formatCurrency(s.target) }}</p>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-bold text-slate-600">{{ formatCurrency(s.achievedLama) }}</p>
              </td>
              <td class="px-8 py-6">
                <p class="text-sm font-bold text-blue-600">{{ formatCurrency(s.achievedBaru) }}</p>
              </td>
              <td class="px-8 py-6">
                <div class="space-y-1.5">
                  <div class="flex justify-between text-[11px] font-bold">
                    <span class="text-slate-900">{{
                      formatCurrency(s.achievedLama + s.achievedBaru)
                    }}</span>
                    <span class="text-slate-400"
                      >{{
                        s.target > 0
                          ? Math.round(((s.achievedLama + s.achievedBaru) / s.target) * 100)
                          : 0
                      }}%</span
                    >
                  </div>
                  <div class="h-2 w-40 bg-slate-100 rounded-full overflow-hidden flex">
                    <div
                      class="h-full bg-slate-400"
                      :style="`width: ${s.target > 0 ? (s.achievedLama / s.target) * 100 : 0}%`"
                    ></div>
                    <div
                      class="h-full bg-emerald-500"
                      :style="`width: ${s.target > 0 ? (s.achievedBaru / s.target) * 100 : 0}%`"
                    ></div>
                  </div>
                </div>
              </td>
              <td class="px-8 py-6">
                <div class="flex flex-col gap-1">
                  <p
                    class="text-sm font-black"
                    :class="
                      s.achievedNewStores >= s.targetNewStores
                        ? 'text-emerald-600'
                        : 'text-slate-900'
                    "
                  >
                    {{ s.achievedNewStores }} / {{ s.targetNewStores }}
                  </p>
                  <div
                    v-if="s.achievedNewStores >= s.targetNewStores"
                    class="flex items-center gap-1"
                  >
                    <CheckCircle2 class="w-3 h-3 text-emerald-500" />
                    <span class="text-[9px] font-black text-emerald-600 uppercase tracking-tighter"
                      >Goal Reached</span
                    >
                  </div>
                </div>
              </td>
              <td class="px-8 py-6">
                <div class="flex flex-col gap-1">
                  <p class="text-sm font-black text-slate-900">
                    {{ s.totalVisits }}
                    <span class="text-[10px] font-bold text-slate-400">Total</span>
                  </p>
                  <p class="text-[11px] font-bold text-slate-500">
                    Target Harian: <span class="text-slate-900">{{ s.targetDailyVisits }}</span>
                  </p>
                </div>
              </td>
              <td class="px-8 py-6">
                <div
                  class="bg-amber-50 px-3 py-1.5 rounded-xl border border-amber-100 inline-block"
                >
                  <p class="text-xs font-black text-amber-700">{{ formatCurrency(s.bonus) }}</p>
                </div>
              </td>
              <td class="px-8 py-6 text-center">
                <div v-if="canManageTargets" class="flex items-center justify-center gap-2">
                  <Button
                    @click="handleEditTarget(s)"
                    variant="outline"
                    size="sm"
                    class="h-8 px-3 rounded-lg border-slate-200 text-slate-600 hover:bg-primary hover:text-white transition-all font-bold gap-1.5 cursor-pointer"
                  >
                    <Edit3 class="w-3.5 h-3.5" /> Edit
                  </Button>
                  <Button
                    @click="handleDeleteKPI(s.id)"
                    :disabled="!s.id"
                    variant="outline"
                    size="sm"
                    class="h-8 px-3 rounded-lg border-red-100 text-red-500 hover:bg-red-500 hover:text-white transition-all font-bold gap-1.5 disabled:opacity-30 cursor-pointer"
                  >
                    <Trash2 class="w-3.5 h-3.5" /> Hapus
                  </Button>
                </div>
                <span v-else class="text-xs font-semibold text-slate-400">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Manual Entry Modal -->
    <Transition name="fade">
      <div
        v-if="showManualEntry"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm"
      >
        <div
          class="bg-white w-full max-w-md rounded-[32px] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-300"
        >
          <div
            class="p-8 border-b border-slate-100 flex items-center justify-between bg-slate-50/50"
          >
            <h3 class="text-xl font-extrabold text-slate-900">Input Nota Manual</h3>
            <button
              @click="showManualEntry = false"
              class="p-2 hover:bg-slate-200 rounded-xl transition-all cursor-pointer"
            >
              <X class="w-5 h-5 text-slate-500" />
            </button>
          </div>

          <div class="p-8 space-y-6">
            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                >Pilih Salesman</label
              >
              <select
                v-model="manualEntryForm.employeeID"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none cursor-pointer"
              >
                <option value="" disabled>Pilih Salesman</option>
                <option v-for="s in salesmenKpi" :key="s.employeeID" :value="s.employeeID">{{ s.name }} ({{ s.branch }})</option>
              </select>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                  >Kategori Toko</label
                >
                <select
                  v-model="manualEntryForm.storeCategory"
                  class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all appearance-none"
                >
                  <option value="TOKO_LAMA">Toko Lama</option>
                  <option value="TOKO_BARU">Toko Baru</option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                  >Total Omzet</label
                >
                <input
                  v-model.number="manualEntryForm.totalAmount"
                  type="number"
                  placeholder="Rp 0"
                  class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
                />
              </div>
            </div>

            <div
              class="p-4 bg-primary/5 rounded-2xl border border-primary/10 flex items-start gap-3"
            >
              <AlertCircle class="w-5 h-5 text-primary shrink-0 mt-0.5" />
              <p class="text-[11px] font-bold text-primary leading-relaxed">
                Input manual akan langsung memperbarui KPI salesman dan perhitungan bonus bulan
                berjalan. Pastikan data nota fisik sudah valid.
              </p>
            </div>
          </div>

          <div class="p-8 bg-slate-50 flex gap-3">
            <button
              @click="showManualEntry = false"
              class="flex-1 py-4 text-sm font-black text-slate-500 hover:text-slate-900 transition-all"
            >
              BATAL
            </button>
            <Button
              @click="submitManualEntry"
              class="flex-[2] py-6 rounded-2xl font-black shadow-lg shadow-primary/20"
              >SIMPAN DATA</Button
            >
          </div>
        </div>
      </div>
    </Transition>

    <!-- Edit Target Modal -->
    <Transition name="fade">
      <div
        v-if="showEditTarget"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm"
      >
        <div
          class="bg-white w-full max-w-md rounded-[32px] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-300"
        >
          <div
            class="p-8 border-b border-slate-100 flex items-center justify-between bg-slate-50/50"
          >
            <div>
              <h3 class="text-xl font-extrabold text-slate-900">Edit Target Sales</h3>
              <p class="text-[10px] font-black text-slate-400 uppercase tracking-widest mt-0.5">
                {{ editForm.name }}
              </p>
            </div>
            <button
              @click="showEditTarget = false"
              class="p-2 hover:bg-slate-200 rounded-xl transition-all"
            >
              <X class="w-5 h-5 text-slate-500" />
            </button>
          </div>

          <div class="p-8 space-y-6">
            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                >Target Omzet (Rp)</label
              >
              <input
                v-model.number="editForm.targetOmzet"
                type="number"
                min="0"
                placeholder="Rp 0"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>

            <div class="space-y-2">
              <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                >Target Toko Baru (Titik)</label
              >
              <input
                v-model.number="editForm.targetNewStores"
                type="number"
                min="0"
                placeholder="0"
                class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                  >Target Harian Kunjungan</label
                >
                <input
                  v-model.number="editForm.targetDailyVisits"
                  type="number"
                  min="0"
                  placeholder="0"
                  class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
                />
              </div>
              <div class="space-y-2">
                <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                  >Target Harian Toko Baru</label
                >
                <input
                  v-model.number="editForm.targetDailyNewStores"
                  type="number"
                  min="0"
                  placeholder="0"
                  class="w-full bg-slate-50 border-none rounded-2xl px-5 py-4 font-bold text-slate-900 focus:ring-2 focus:ring-primary/20 transition-all"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                  >Bulan</label
                >
                <div
                  class="w-full bg-slate-100 rounded-2xl px-5 py-4 font-bold text-slate-500 text-sm border border-slate-200"
                >
                  {{ months.find((m) => m.value === editForm.month)?.label }}
                </div>
              </div>
              <div class="space-y-2">
                <label class="text-[11px] font-black text-slate-400 uppercase tracking-widest"
                  >Tahun</label
                >
                <div
                  class="w-full bg-slate-100 rounded-2xl px-5 py-4 font-bold text-slate-500 text-sm border border-slate-200"
                >
                  {{ editForm.year }}
                </div>
              </div>
            </div>
          </div>

          <div class="p-8 bg-slate-50 flex gap-3">
            <button
              @click="showEditTarget = false"
              class="flex-1 py-4 text-sm font-black text-slate-500 hover:text-slate-900 transition-all"
            >
              BATAL
            </button>
            <Button
              @click="submitEditTarget"
              class="flex-[2] py-6 rounded-2xl font-black shadow-lg shadow-primary/20 flex items-center justify-center gap-2"
            >
              <Save class="w-5 h-5" /> SIMPAN PERUBAHAN
            </Button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
