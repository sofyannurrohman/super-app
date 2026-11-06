<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import apiClient from '@/api/axios'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'
import {
  TrendingUp,
  Plus,
  Search,
  Filter,
  FileDown,
  RefreshCw,
  Clock,
  CheckCircle2,
  AlertCircle,
  FileText,
  Calendar,
  Building2,
  User,
  Eye,
  Edit3,
  Trash2,
  Award,
  Layers,
  ChevronRight,
  ShieldCheck,
  Target,
  Sparkles,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import DailyKpiReportFormDialog from './DailyKpiReportFormDialog.vue'
import DailyKpiReportReviewDialog from './DailyKpiReportReviewDialog.vue'

const masterDataStore = useMasterDataStore()
const authStore = useAuthStore()

const isManagerOrAdmin = computed(() => {
  const role = authStore.user?.role || ''
  return ['superadmin', 'manager', 'supervisor_sales', 'supervisor_sales_mt', 'hr_admin'].includes(role)
})

// Tab & Filters
const activeTab = ref<'MY_REPORTS' | 'PENDING_REVIEW' | 'ANALYTICS'>('MY_REPORTS')
const filterDate = ref(new Date().toISOString().split('T')[0])
const filterStatus = ref('ALL')
const filterRole = ref('ALL')
const searchQuery = ref('')

const reports = ref<any[]>([])
const isLoading = ref(false)

// Dialog states
const isFormDialogOpen = ref(false)
const selectedReportIdForEdit = ref<string | null>(null)
const isReviewDialogOpen = ref(false)
const selectedReportIdForReview = ref<string | null>(null)

const formatCurrency = (val: number) => {
  if (isNaN(val)) return '0'
  return new Intl.NumberFormat('id-ID').format(val)
}

const getAchievementBadgeColor = (percent: number) => {
  if (percent >= 100) return 'bg-emerald-100 text-emerald-800 border-emerald-300'
  if (percent >= 80) return 'bg-amber-100 text-amber-800 border-amber-300'
  return 'bg-rose-100 text-rose-800 border-rose-300'
}

const fetchReports = async () => {
  isLoading.value = true
  try {
    let url = `/sales/daily-kpi-reports?date=${filterDate.value}&limit=50`
    const bId = masterDataStore.selectedBranchId
    if (bId && bId !== 'ALL' && bId !== 'ALL_BRANCHES') {
      url += `&branch_id=${bId}`
    }
    if (filterStatus.value !== 'ALL') {
      url += `&status=${filterStatus.value}`
    }
    if (filterRole.value !== 'ALL') {
      url += `&reporter_role=${filterRole.value}`
    }

    const res = await apiClient.get(url)
    reports.value = res.data.data || []
  } catch (err: any) {
    toast.error('Gagal mengambil daftar laporan: ' + (err.response?.data?.message || err.message))
  } finally {
    isLoading.value = false
  }
}

// Filtered lists per tab
const myReports = computed(() => {
  const currentEmpId = authStore.user?.employee?.id
  const isSuper = authStore.user?.role === 'superadmin'

  return reports.value.filter((r) => {
    const matchesSearch =
      !searchQuery.value ||
      (r.reporter_employee &&
        `${r.reporter_employee.first_name} ${r.reporter_employee.last_name || ''}`
          .toLowerCase()
          .includes(searchQuery.value.toLowerCase())) ||
      (r.area_name && r.area_name.toLowerCase().includes(searchQuery.value.toLowerCase()))

    if (!matchesSearch) return false

    if (isSuper) return true
    if (currentEmpId && r.reporter_employee_id === currentEmpId) return true
    return true // Fallback: display reports of branch for SPVs
  })
})

const pendingReviewReports = computed(() => {
  return reports.value.filter((r) => r.status === 'SUBMITTED')
})

// Summary Metrics for Tab 3
const analytics = computed(() => {
  let totalCalls = 0
  let targetCalls = 0
  let totalEC = 0
  let targetEC = 0
  let totalSales = 0
  let targetSales = 0

  const allObstacles: any[] = []

  for (const r of reports.value) {
    totalCalls += Number(r.summary_realization_call) || 0
    targetCalls += Number(r.summary_target_call) || 0
    totalEC += Number(r.summary_realization_ec) || 0
    targetEC += Number(r.summary_target_ec) || 0
    totalSales += Number(r.summary_realization_sales) || 0
    targetSales += Number(r.summary_target_sales) || 0

    if (r.action_plans && Array.isArray(r.action_plans)) {
      for (const ap of r.action_plans) {
        allObstacles.push({
          ...ap,
          reporter_name: r.reporter_employee ? `${r.reporter_employee.first_name} ${r.reporter_employee.last_name || ''}`.trim() : '-',
          area_name: r.area_name || (r.branch ? r.branch.name : '-'),
        })
      }
    }
  }

  const capaiCall = targetCalls > 0 ? (totalCalls / targetCalls) * 100 : 0
  const capaiEC = targetEC > 0 ? (totalEC / targetEC) * 100 : 0
  const capaiSales = targetSales > 0 ? (totalSales / targetSales) * 100 : 0

  return {
    totalCalls,
    targetCalls,
    capaiCall,
    totalEC,
    targetEC,
    capaiEC,
    totalSales,
    targetSales,
    capaiSales,
    obstacles: allObstacles,
  }
})

// Actions
const openCreateDialog = () => {
  selectedReportIdForEdit.value = null
  isFormDialogOpen.value = true
}

const openEditDialog = (reportId: string) => {
  selectedReportIdForEdit.value = reportId
  isFormDialogOpen.value = true
}

const openReviewDialog = (reportId: string) => {
  selectedReportIdForReview.value = reportId
  isReviewDialogOpen.value = true
}

const downloadExcel = (reportId: string) => {
  window.open(`${apiClient.defaults.baseURL}/sales/daily-kpi-reports/${reportId}/export?token=${localStorage.getItem('token') || ''}`, '_blank')
}

const deleteReport = async (reportId: string) => {
  if (!confirm('Apakah Anda yakin ingin menghapus draft laporan ini?')) return
  try {
    await apiClient.delete(`/sales/daily-kpi-reports/${reportId}`)
    toast.success('Laporan berhasil dihapus')
    fetchReports()
  } catch (err: any) {
    toast.error('Gagal menghapus laporan: ' + (err.response?.data?.message || err.message))
  }
}

watch([filterDate, filterStatus, filterRole, () => masterDataStore.selectedBranchId], () => {
  fetchReports()
})

onMounted(() => {
  fetchReports()
})
</script>

<template>
  <div class="space-y-6 p-4 sm:p-6 max-w-7xl mx-auto">
    
    <!-- PAGE HEADER -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center bg-white p-6 sm:p-8 rounded-3xl shadow-xs border border-slate-200/80 gap-4">
      <div class="space-y-1">
        <h1 class="text-2xl sm:text-3xl font-black text-slate-800 tracking-tight flex items-center flex-wrap gap-3">
          <div class="w-10 h-10 rounded-2xl bg-indigo-600 flex items-center justify-center text-white shadow-md shadow-indigo-200">
            <Award class="w-6 h-6" />
          </div>
          <span>Pelaporan Kinerja Harian Sales</span>
          <span v-if="masterDataStore.selectedBranchName" class="text-indigo-600 text-xl font-bold">
            — {{ masterDataStore.selectedBranchName }}
          </span>
        </h1>
        <p class="text-xs sm:text-sm text-slate-500 font-medium">
          Format baku pelaporan KPI harian berjenjang (SPV &rarr; ASM &rarr; NSM &rarr; GM) lengkap dengan rincian Call, EC, Penjualan, &amp; Tindak Lanjut Kendala.
        </p>
      </div>

      <div class="flex items-center gap-3 w-full md:w-auto">
        <Button
          variant="outline"
          class="rounded-2xl font-bold gap-2 text-slate-700 border-slate-200 hover:bg-slate-50"
          :disabled="isLoading"
          @click="fetchReports"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isLoading }" />
          <span>Segarkan Data</span>
        </Button>

        <Button
          class="rounded-2xl font-extrabold gap-2 bg-indigo-600 hover:bg-indigo-700 text-white shadow-md shadow-indigo-200"
          @click="openCreateDialog"
        >
          <Plus class="w-4 h-4" />
          <span>+ Buat Laporan Harian</span>
        </Button>
      </div>
    </div>

    <!-- TABS NAVIGATION -->
    <div class="flex items-center gap-2 border-b border-slate-200 overflow-x-auto pb-px">
      <button
        @click="activeTab = 'MY_REPORTS'"
        :class="[
          'px-5 py-3 text-sm font-extrabold rounded-t-2xl transition-all flex items-center gap-2',
          activeTab === 'MY_REPORTS'
            ? 'bg-white text-indigo-600 border-t-2 border-x border-indigo-600 shadow-2xs'
            : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
        ]"
      >
        <FileText class="w-4 h-4" />
        <span>Daftar Laporan Harian</span>
        <span class="text-xs px-2 py-0.5 rounded-full bg-slate-100 text-slate-700 font-bold ml-1">
          {{ myReports.length }}
        </span>
      </button>

      <button
        @click="activeTab = 'PENDING_REVIEW'"
        :class="[
          'px-5 py-3 text-sm font-extrabold rounded-t-2xl transition-all flex items-center gap-2',
          activeTab === 'PENDING_REVIEW'
            ? 'bg-white text-indigo-600 border-t-2 border-x border-indigo-600 shadow-2xs'
            : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
        ]"
      >
        <ShieldCheck class="w-4 h-4" />
        <span>Review Laporan Masuk</span>
        <span
          v-if="pendingReviewReports.length > 0"
          class="text-xs px-2 py-0.5 rounded-full bg-amber-500 text-white font-black ml-1 animate-pulse"
        >
          {{ pendingReviewReports.length }}
        </span>
      </button>

      <button
        @click="activeTab = 'ANALYTICS'"
        :class="[
          'px-5 py-3 text-sm font-extrabold rounded-t-2xl transition-all flex items-center gap-2',
          activeTab === 'ANALYTICS'
            ? 'bg-white text-indigo-600 border-t-2 border-x border-indigo-600 shadow-2xs'
            : 'text-slate-500 hover:text-slate-800 hover:bg-slate-50',
        ]"
      >
        <TrendingUp class="w-4 h-4" />
        <span>Ringkasan &amp; Kendala Lapangan</span>
      </button>
    </div>

    <!-- FILTER BAR -->
    <div class="bg-white p-4 rounded-2xl border border-slate-200/80 shadow-xs flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center flex-wrap gap-3">
        <!-- Date Filter -->
        <div class="flex items-center gap-2 bg-slate-50 px-3 py-2 rounded-xl border border-slate-200 text-xs font-semibold text-slate-700">
          <Calendar class="w-4 h-4 text-indigo-500" />
          <span>Tanggal:</span>
          <input
            type="date"
            v-model="filterDate"
            class="bg-transparent font-bold text-slate-800 outline-none"
          />
        </div>

        <!-- Status Filter -->
        <div class="flex items-center gap-2 bg-slate-50 px-3 py-2 rounded-xl border border-slate-200 text-xs font-semibold text-slate-700">
          <span>Status:</span>
          <select v-model="filterStatus" class="bg-transparent font-bold text-slate-800 outline-none cursor-pointer">
            <option value="ALL">Semua Status</option>
            <option value="DRAFT">Draft</option>
            <option value="SUBMITTED">Menunggu Review</option>
            <option value="APPROVED">Disetujui (Approved)</option>
            <option value="REVISED">Perlu Revisi</option>
          </select>
        </div>

        <!-- Role Filter -->
        <div class="flex items-center gap-2 bg-slate-50 px-3 py-2 rounded-xl border border-slate-200 text-xs font-semibold text-slate-700">
          <span>Jenjang:</span>
          <select v-model="filterRole" class="bg-transparent font-bold text-slate-800 outline-none cursor-pointer">
            <option value="ALL">Semua Jenjang</option>
            <option value="SPV">SPV &rarr; ASM</option>
            <option value="ASM">ASM &rarr; NSM</option>
            <option value="NSM">NSM &rarr; GM</option>
            <option value="GM">GM / Direksi</option>
          </select>
        </div>
      </div>

      <!-- Search Input -->
      <div class="relative w-full sm:w-64">
        <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          type="text"
          v-model="searchQuery"
          placeholder="Cari SPV atau area..."
          class="w-full text-xs font-medium pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>
    </div>

    <!-- TAB 1: DAFTAR LAPORAN HARIAN -->
    <div v-if="activeTab === 'MY_REPORTS'" class="space-y-4">
      <div class="overflow-x-auto bg-white rounded-3xl border border-slate-200 shadow-xs">
        <table class="w-full text-xs text-left">
          <thead class="bg-slate-900 text-white font-extrabold uppercase tracking-wider text-[11px]">
            <tr>
              <th class="py-3.5 px-4">Tanggal</th>
              <th class="py-3.5 px-4">Pelapor (SPV / ASM)</th>
              <th class="py-3.5 px-4">Cabang / Area</th>
              <th class="py-3.5 px-3 text-center">Pencapaian Call</th>
              <th class="py-3.5 px-3 text-center">Pencapaian EC</th>
              <th class="py-3.5 px-3 text-center">Pencapaian Penjualan</th>
              <th class="py-3.5 px-3 text-center">Status</th>
              <th class="py-3.5 px-4 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="isLoading">
              <td colspan="8" class="py-12 text-center text-slate-400 font-medium">
                <RefreshCw class="w-6 h-6 animate-spin mx-auto mb-2 text-indigo-500" />
                Memuat data laporan...
              </td>
            </tr>
            <tr v-else-if="myReports.length === 0">
              <td colspan="8" class="py-12 text-center text-slate-400 font-medium">
                <div class="max-w-sm mx-auto space-y-3">
                  <div class="w-12 h-12 rounded-2xl bg-indigo-50 flex items-center justify-center text-indigo-600 mx-auto">
                    <FileText class="w-6 h-6" />
                  </div>
                  <p class="text-sm font-bold text-slate-700">Belum ada laporan harian pada tanggal ini</p>
                  <p class="text-xs text-slate-400">Klik tombol <strong>"+ Buat Laporan Harian"</strong> di atas untuk membuat laporan tim.</p>
                </div>
              </td>
            </tr>
            <tr v-for="r in myReports" :key="r.id" class="hover:bg-slate-50/70 transition-colors">
              <!-- Tanggal -->
              <td class="p-4 font-bold text-slate-800 whitespace-nowrap">
                {{ r.report_date ? new Date(r.report_date).toLocaleDateString('id-ID', { dateStyle: 'medium' }) : '-' }}
              </td>

              <!-- Pelapor -->
              <td class="p-4">
                <div class="font-extrabold text-slate-900 flex items-center gap-1.5">
                  <span>{{ r.reporter_employee ? `${r.reporter_employee.first_name} ${r.reporter_employee.last_name || ''}`.trim() : '-' }}</span>
                  <span class="text-[10px] font-black px-2 py-0.5 rounded-md bg-indigo-50 text-indigo-700">
                    {{ r.reporter_role || 'SPV' }}
                  </span>
                </div>
              </td>

              <!-- Area -->
              <td class="p-4 font-semibold text-slate-700">
                {{ r.area_name || (r.branch ? r.branch.name : '-') }}
              </td>

              <!-- Capai Call -->
              <td class="p-4 text-center">
                <div class="flex flex-col items-center">
                  <span :class="['px-2.5 py-0.5 rounded-full font-black text-xs border shadow-2xs', getAchievementBadgeColor(r.summary_achievement_call_percent)]">
                    {{ Number(r.summary_achievement_call_percent || 0).toFixed(0) }}%
                  </span>
                  <span class="text-[10px] text-slate-400 mt-1 font-semibold">
                    {{ r.summary_realization_call }} / {{ r.summary_target_call }}
                  </span>
                </div>
              </td>

              <!-- Capai EC -->
              <td class="p-4 text-center">
                <div class="flex flex-col items-center">
                  <span :class="['px-2.5 py-0.5 rounded-full font-black text-xs border shadow-2xs', getAchievementBadgeColor(r.summary_achievement_ec_percent)]">
                    {{ Number(r.summary_achievement_ec_percent || 0).toFixed(0) }}%
                  </span>
                  <span class="text-[10px] text-slate-400 mt-1 font-semibold">
                    {{ r.summary_realization_ec }} / {{ r.summary_target_ec }}
                  </span>
                </div>
              </td>

              <!-- Capai Penjualan -->
              <td class="p-4 text-center">
                <div class="flex flex-col items-center">
                  <span :class="['px-2.5 py-0.5 rounded-full font-black text-xs border shadow-2xs', getAchievementBadgeColor(r.summary_achievement_sales_percent)]">
                    {{ Number(r.summary_achievement_sales_percent || 0).toFixed(0) }}%
                  </span>
                  <span class="text-[10px] text-emerald-700 font-bold mt-1">
                    Rp {{ formatCurrency(r.summary_realization_sales) }}
                  </span>
                </div>
              </td>

              <!-- Status -->
              <td class="p-4 text-center">
                <span class="px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-wider inline-block"
                  :class="{
                    'bg-emerald-100 text-emerald-800': r.status === 'APPROVED',
                    'bg-amber-100 text-amber-800': r.status === 'SUBMITTED',
                    'bg-rose-100 text-rose-800': r.status === 'REVISED',
                    'bg-slate-200 text-slate-700': r.status === 'DRAFT',
                  }"
                >
                  {{ r.status === 'SUBMITTED' ? 'MENUNGGU REVIEW' : r.status }}
                </span>
              </td>

              <!-- Aksi -->
              <td class="p-4 text-center">
                <div class="flex items-center justify-center gap-1.5">
                  <!-- Review / Detail -->
                  <Button
                    size="sm"
                    variant="outline"
                    class="h-8 px-2.5 rounded-xl font-bold gap-1 text-slate-700"
                    @click="openReviewDialog(r.id)"
                  >
                    <Eye class="w-3.5 h-3.5" /> Detail
                  </Button>

                  <!-- Edit (Draft or Revised) -->
                  <Button
                    v-if="r.status === 'DRAFT' || r.status === 'REVISED' || authStore.user?.role === 'superadmin'"
                    size="sm"
                    variant="outline"
                    class="h-8 px-2.5 rounded-xl font-bold gap-1 text-indigo-600 border-indigo-200 hover:bg-indigo-50"
                    @click="openEditDialog(r.id)"
                  >
                    <Edit3 class="w-3.5 h-3.5" /> Edit
                  </Button>

                  <!-- Export Excel -->
                  <button
                    type="button"
                    title="Export Excel Template"
                    @click="downloadExcel(r.id)"
                    class="p-1.5 rounded-xl text-slate-400 hover:text-emerald-600 hover:bg-emerald-50 transition-colors"
                  >
                    <FileDown class="w-4 h-4" />
                  </button>

                  <!-- Delete Draft -->
                  <button
                    v-if="r.status === 'DRAFT' || authStore.user?.role === 'superadmin'"
                    type="button"
                    title="Hapus Draft"
                    @click="deleteReport(r.id)"
                    class="p-1.5 rounded-xl text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition-colors"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- TAB 2: REVIEW LAPORAN MASUK -->
    <div v-if="activeTab === 'PENDING_REVIEW'" class="space-y-4">
      <div v-if="pendingReviewReports.length === 0" class="bg-white p-12 rounded-3xl border border-slate-200 text-center">
        <div class="w-12 h-12 rounded-2xl bg-emerald-50 flex items-center justify-center text-emerald-600 mx-auto mb-3">
          <CheckCircle2 class="w-6 h-6" />
        </div>
        <h3 class="text-base font-extrabold text-slate-800">Semua Laporan Telah Direview</h3>
        <p class="text-xs text-slate-400 mt-1">Tidak ada antrian laporan harian yang menunggu validasi Anda.</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
        <div
          v-for="r in pendingReviewReports"
          :key="r.id"
          class="bg-white p-6 rounded-3xl border border-slate-200 shadow-xs hover:shadow-md transition-all space-y-4 flex flex-col justify-between"
        >
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-slate-400 flex items-center gap-1">
                <Calendar class="w-3.5 h-3.5" />
                {{ r.report_date ? new Date(r.report_date).toLocaleDateString('id-ID', { dateStyle: 'medium' }) : '-' }}
              </span>
              <span class="text-[10px] font-black px-2.5 py-0.5 rounded-full bg-amber-100 text-amber-800">
                MENUNGGU REVIEW
              </span>
            </div>

            <div>
              <h4 class="text-base font-black text-slate-900">
                {{ r.reporter_employee ? `${r.reporter_employee.first_name} ${r.reporter_employee.last_name || ''}`.trim() : '-' }}
              </h4>
              <p class="text-xs text-slate-500 font-medium mt-0.5">
                {{ r.reporter_role || 'SPV' }} &bull; {{ r.area_name || (r.branch ? r.branch.name : '-') }}
              </p>
            </div>

            <!-- KPI Pill Grid -->
            <div class="grid grid-cols-3 gap-2 p-3 bg-slate-50 rounded-2xl border border-slate-100 text-center">
              <div>
                <span class="text-[10px] font-bold text-slate-400 uppercase block">Call</span>
                <span class="text-xs font-black text-slate-800 mt-0.5 block">
                  {{ Number(r.summary_achievement_call_percent || 0).toFixed(0) }}%
                </span>
              </div>
              <div>
                <span class="text-[10px] font-bold text-slate-400 uppercase block">EC</span>
                <span class="text-xs font-black text-slate-800 mt-0.5 block">
                  {{ Number(r.summary_achievement_ec_percent || 0).toFixed(0) }}%
                </span>
              </div>
              <div>
                <span class="text-[10px] font-bold text-slate-400 uppercase block">Omzet</span>
                <span class="text-xs font-black text-emerald-700 mt-0.5 block">
                  {{ Number(r.summary_achievement_sales_percent || 0).toFixed(0) }}%
                </span>
              </div>
            </div>

            <div v-if="r.tomorrow_priority" class="text-xs text-slate-600 bg-amber-50/50 p-2.5 rounded-xl border border-amber-100 line-clamp-2">
              <strong class="text-amber-900">Prioritas Besok:</strong> {{ r.tomorrow_priority }}
            </div>
          </div>

          <div class="flex items-center gap-2 pt-2 border-t border-slate-100">
            <Button
              class="w-full rounded-xl font-extrabold gap-1 bg-indigo-600 hover:bg-indigo-700 text-white"
              size="sm"
              @click="openReviewDialog(r.id)"
            >
              <Eye class="w-3.5 h-3.5" /> Review Sekarang
            </Button>
            <button
              type="button"
              @click="downloadExcel(r.id)"
              class="p-2 rounded-xl text-slate-400 hover:text-emerald-600 hover:bg-emerald-50 transition-colors"
            >
              <FileDown class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 3: RINGKASAN & ANALISIS KENDALA -->
    <div v-if="activeTab === 'ANALYTICS'" class="space-y-6">
      <!-- 3 Metrics Card -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
        <div class="bg-white p-6 rounded-3xl border border-slate-200 shadow-xs">
          <span class="text-xs font-bold text-slate-400 uppercase tracking-wider block">Total Realisasi Call Tim</span>
          <div class="mt-2 flex items-baseline gap-2">
            <span class="text-3xl font-black text-slate-900">{{ analytics.totalCalls }}</span>
            <span class="text-xs font-semibold text-slate-400">/ Target {{ analytics.targetCalls }}</span>
          </div>
          <div class="mt-3 flex items-center justify-between text-xs font-bold">
            <span class="text-slate-500">Pencapaian:</span>
            <span :class="['px-2.5 py-0.5 rounded-full border', getAchievementBadgeColor(analytics.capaiCall)]">
              {{ analytics.capaiCall.toFixed(1) }}%
            </span>
          </div>
        </div>

        <div class="bg-white p-6 rounded-3xl border border-slate-200 shadow-xs">
          <span class="text-xs font-bold text-slate-400 uppercase tracking-wider block">Total Realisasi EC Tim</span>
          <div class="mt-2 flex items-baseline gap-2">
            <span class="text-3xl font-black text-slate-900">{{ analytics.totalEC }}</span>
            <span class="text-xs font-semibold text-slate-400">/ Target {{ analytics.targetEC }}</span>
          </div>
          <div class="mt-3 flex items-center justify-between text-xs font-bold">
            <span class="text-slate-500">Pencapaian:</span>
            <span :class="['px-2.5 py-0.5 rounded-full border', getAchievementBadgeColor(analytics.capaiEC)]">
              {{ analytics.capaiEC.toFixed(1) }}%
            </span>
          </div>
        </div>

        <div class="bg-white p-6 rounded-3xl border border-slate-200 shadow-xs">
          <span class="text-xs font-bold text-slate-400 uppercase tracking-wider block">Total Omzet Penjualan Bersih</span>
          <div class="mt-2 flex items-baseline gap-2">
            <span class="text-2xl font-black text-emerald-600">Rp {{ formatCurrency(analytics.totalSales) }}</span>
          </div>
          <div class="mt-3 flex items-center justify-between text-xs font-bold">
            <span class="text-slate-500">Target: Rp {{ formatCurrency(analytics.targetSales) }}</span>
            <span :class="['px-2.5 py-0.5 rounded-full border', getAchievementBadgeColor(analytics.capaiSales)]">
              {{ analytics.capaiSales.toFixed(1) }}%
            </span>
          </div>
        </div>
      </div>

      <!-- Obstacles Table -->
      <div class="bg-white rounded-3xl border border-slate-200 shadow-xs overflow-hidden">
        <div class="p-6 border-b border-slate-200">
          <h3 class="text-base font-extrabold text-slate-900 flex items-center gap-2">
            <AlertCircle class="w-5 h-5 text-amber-500" />
            <span>Matriks Kendala Lapangan &amp; Tindak Lanjut Tim</span>
          </h3>
          <p class="text-xs text-slate-500 mt-0.5">Daftar kendala yang dilaporkan oleh para SPV pada tanggal ini</p>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-xs text-left">
            <thead class="bg-slate-900 text-white font-extrabold uppercase tracking-wider text-[11px]">
              <tr>
                <th class="py-3.5 px-4">Salesman</th>
                <th class="py-3.5 px-4">Pelapor &bull; Area</th>
                <th class="py-3.5 px-4">Alasan Belum Capai Target</th>
                <th class="py-3.5 px-4">Kendala Lapangan</th>
                <th class="py-3.5 px-4">Tindak Lanjut SPV</th>
                <th class="py-3.5 px-3">PIC</th>
                <th class="py-3.5 px-3">Batas Waktu</th>
                <th class="py-3.5 px-4">Dukungan Diperlukan</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="analytics.obstacles.length === 0">
                <td colspan="8" class="py-8 text-center text-slate-400 font-medium">
                  Tidak ada kendala lapangan atau evaluasi yang tercatat pada tanggal ini.
                </td>
              </tr>
              <tr v-for="(ob, idx) in analytics.obstacles" :key="idx" class="hover:bg-slate-50/70">
                <td class="p-4 font-bold text-slate-800">{{ ob.employee_name }}</td>
                <td class="p-4 text-slate-600 font-medium">
                  {{ ob.reporter_name }} &bull; {{ ob.area_name }}
                </td>
                <td class="p-4 text-slate-700">{{ ob.reason_unachieved || '-' }}</td>
                <td class="p-4 text-slate-700">{{ ob.field_obstacle || '-' }}</td>
                <td class="p-4 font-semibold text-slate-900">{{ ob.action_follow_up || '-' }}</td>
                <td class="p-4 text-slate-600">{{ ob.pic_name || '-' }}</td>
                <td class="p-4 text-slate-600">{{ ob.due_date ? new Date(ob.due_date).toLocaleDateString('id-ID') : '-' }}</td>
                <td class="p-4 text-amber-900 font-medium bg-amber-50/40">{{ ob.support_needed || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- DIALOG FORM (CREATE / EDIT) -->
    <DailyKpiReportFormDialog
      v-model:open="isFormDialogOpen"
      :report-id="selectedReportIdForEdit"
      :initial-date="filterDate"
      :initial-branch-id="masterDataStore.selectedBranchId"
      @saved="fetchReports"
    />

    <!-- DIALOG REVIEW (APPROVE / REVISE) -->
    <DailyKpiReportReviewDialog
      v-model:open="isReviewDialogOpen"
      :report-id="selectedReportIdForReview"
      @reviewed="fetchReports"
    />

  </div>
</template>
