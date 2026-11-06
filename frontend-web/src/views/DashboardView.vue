<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { toast } from 'vue-sonner'
import apiClient from '@/api/axios'
import { useAuthStore } from '@/stores/auth'
import { useMasterDataStore } from '@/stores/masterData'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
} from 'echarts/components'
import VChart from 'vue-echarts'
import {
  Users,
  CheckCircle2,
  Clock,
  AlertCircle,
  ArrowRight,
  CalendarClock,
  TrendingUp,
  Mail,
  Sparkles,
  Building2,
  ShieldAlert,
  FileText,
  PlusCircle,
  Calendar,
  UserPlus,
  RefreshCw,
  Layers,
  Timer,
  Check,
  X,
  UserCheck,
  FileCheck2,
  AlertTriangle,
  ChevronRight,
  Activity,
  Briefcase,
} from 'lucide-vue-next'

use([
  CanvasRenderer,
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
])

// --- Interfaces ---
interface DashboardSummary {
  total_employees: number
  present_today: number
  on_time_today: number
  late_today: number
  on_leave_today: number
  not_checked_in_today: number
  attendance_rate: number
  pending_leaves_count: number
  pending_overtimes_count: number
  pending_reimbursements_count: number
  expiring_contracts_count: number
}

interface AttendanceTrendPoint {
  date: string
  label: string
  present: number
  on_time: number
  late: number
  leave: number
  attendance_rate: number
}

interface DepartmentDistribution {
  department: string
  count: number
  percentage: number
}

interface EmploymentDistribution {
  status: string
  count: number
}

interface ExpiringContract {
  id: string
  name: string
  employee_id_number: string
  department: string
  job_position: string
  end_date: string
  days_remaining: number
}

interface RecentAttendance {
  id: string
  employee_id: string
  name: string
  job_position: string
  department: string
  clock_in_time: string
  status: string
  notes: string
  avatar_url: string
}

interface PendingLeave {
  id: string
  name: string
  type: string
  start_date: string
  end_date: string
  days: number
  reason: string
  created_at: string
  time_ago: string
}

interface PendingOvertime {
  id: string
  name: string
  date: string
  duration_hrs: number
  reason: string
  created_at: string
  time_ago: string
}

interface HRDashboardData {
  summary: DashboardSummary
  attendance_trend: AttendanceTrendPoint[]
  department_distribution: DepartmentDistribution[]
  employment_distribution: EmploymentDistribution[]
  expiring_contracts: ExpiringContract[]
  recent_attendance: RecentAttendance[]
  pending_leaves: PendingLeave[]
  pending_overtimes: PendingOvertime[]
}

// --- Reactive State ---
const authStore = useAuthStore()
const masterDataStore = useMasterDataStore()

const timeRange = ref<'7d' | '30d' | 'year'>('7d')
const isLoading = ref(true)
const isActionLoading = ref(false)
const activeTab = ref<'attendance' | 'leaves' | 'overtime' | 'contracts'>('attendance')
const attendanceStatusFilter = ref<'ALL' | 'ON_TIME' | 'LATE'>('ALL')

const dashboardData = ref<HRDashboardData>({
  summary: {
    total_employees: 0,
    present_today: 0,
    on_time_today: 0,
    late_today: 0,
    on_leave_today: 0,
    not_checked_in_today: 0,
    attendance_rate: 0,
    pending_leaves_count: 0,
    pending_overtimes_count: 0,
    pending_reimbursements_count: 0,
    expiring_contracts_count: 0,
  },
  attendance_trend: [],
  department_distribution: [],
  employment_distribution: [],
  expiring_contracts: [],
  recent_attendance: [],
  pending_leaves: [],
  pending_overtimes: [],
})

// --- Greetings & Date Helpers ---
const greetingMessage = computed(() => {
  const hour = new Date().getHours()
  if (hour >= 4 && hour < 11) return 'Selamat Pagi'
  if (hour >= 11 && hour < 15) return 'Selamat Siang'
  if (hour >= 15 && hour < 18) return 'Selamat Sore'
  return 'Selamat Malam'
})

const currentFormattedDate = computed(() => {
  return new Intl.DateTimeFormat('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date())
})

const userName = computed(() => {
  const user = authStore.user
  if (!user) return 'Admin'
  if (user.first_name) return user.first_name
  if (user.email) {
    const part = user.email.split('@')[0]
    return part.charAt(0).toUpperCase() + part.slice(1)
  }
  return 'Admin'
})

const activeBranchName = computed(() => {
  const bId = masterDataStore.selectedBranchId
  if (!bId || bId === 'ALL_BRANCHES') return 'Semua Cabang'
  const branch = masterDataStore.branches.find((b) => b.id === bId)
  return branch ? branch.name : 'Cabang Terpilih'
})

// --- Filtered Attendance ---
const filteredRecentAttendance = computed(() => {
  if (attendanceStatusFilter.value === 'ALL') {
    return dashboardData.value.recent_attendance
  }
  if (attendanceStatusFilter.value === 'ON_TIME') {
    return dashboardData.value.recent_attendance.filter((item) => item.status === 'Tepat Waktu')
  }
  return dashboardData.value.recent_attendance.filter((item) => item.status === 'Terlambat')
})

const totalPendingReviewCount = computed(() => {
  const s = dashboardData.value.summary
  return (s.pending_leaves_count || 0) + (s.pending_overtimes_count || 0)
})

// --- ECharts Options ---
const trendChartOption = computed(() => {
  const trend = dashboardData.value.attendance_trend || []
  const categories = trend.map((t) => t.label)
  const onTimeData = trend.map((t) => t.on_time)
  const lateData = trend.map((t) => t.late)
  const leaveData = trend.map((t) => t.leave)

  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      borderColor: '#334155',
      borderWidth: 1,
      borderRadius: 16,
      padding: [12, 16],
      textStyle: {
        color: '#f8fafc',
        fontFamily: 'Plus Jakarta Sans',
        fontSize: 12,
      },
      axisPointer: {
        type: 'shadow',
        shadowStyle: {
          color: 'rgba(226, 232, 240, 0.25)',
        },
      },
    },
    legend: {
      data: ['Tepat Waktu', 'Terlambat', 'Izin & Cuti'],
      bottom: 0,
      icon: 'circle',
      textStyle: {
        fontFamily: 'Plus Jakarta Sans',
        fontWeight: '600',
        color: '#64748b',
        fontSize: 12,
      },
      itemGap: 24,
    },
    grid: {
      left: '2%',
      right: '2%',
      bottom: '12%',
      top: '8%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: categories,
      axisLine: { lineStyle: { color: '#e2e8f0' } },
      axisTick: { show: false },
      axisLabel: {
        color: '#64748b',
        fontWeight: '600',
        fontSize: 11,
      },
    },
    yAxis: [
      {
        type: 'value',
        name: 'Presensi (Orang)',
        nameTextStyle: { color: '#94a3b8', fontSize: 11, padding: [0, 0, 4, 0] },
        splitLine: { lineStyle: { color: '#f1f5f9', type: 'dashed' } },
        axisLabel: { color: '#64748b', fontWeight: '600' },
      },
      {
        type: 'value',
        name: 'Cuti / Izin',
        nameTextStyle: { color: '#94a3b8', fontSize: 11, padding: [0, 0, 4, 0] },
        splitLine: { show: false },
        axisLabel: { color: '#64748b', fontWeight: '600' },
      },
    ],
    series: [
      {
        name: 'Tepat Waktu',
        type: 'bar',
        stack: 'presensi',
        data: onTimeData,
        itemStyle: {
          color: '#10b981',
          borderRadius: [0, 0, 0, 0],
        },
        barMaxWidth: 28,
      },
      {
        name: 'Terlambat',
        type: 'bar',
        stack: 'presensi',
        data: lateData,
        itemStyle: {
          color: '#f59e0b',
          borderRadius: [6, 6, 0, 0],
        },
        barMaxWidth: 28,
      },
      {
        name: 'Izin & Cuti',
        type: 'line',
        yAxisIndex: 1,
        data: leaveData,
        smooth: true,
        symbolSize: 7,
        itemStyle: { color: '#6366f1' },
        lineStyle: { width: 3, color: '#6366f1' },
      },
    ],
  }
})

const deptDistributionOption = computed(() => {
  const depts = dashboardData.value.department_distribution || []
  const total = dashboardData.value.summary.total_employees || 0

  const colors = [
    '#0ea5e9', // sky
    '#10b981', // emerald
    '#8b5cf6', // purple
    '#f59e0b', // amber
    '#ec4899', // pink
    '#06b6d4', // cyan
    '#f97316', // orange
    '#64748b', // slate
  ]

  const chartData = depts.map((d, idx) => ({
    value: d.count,
    name: d.department,
    itemStyle: { color: colors[idx % colors.length] },
  }))

  return {
    tooltip: {
      trigger: 'item',
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      borderColor: '#334155',
      borderRadius: 14,
      padding: [10, 14],
      textStyle: { color: '#f8fafc', fontFamily: 'Plus Jakarta Sans', fontSize: 12 },
      formatter: '{b}: <b style="color:#38bdf8">{c} staf</b> ({d}%)',
    },
    title: {
      text: total.toString(),
      subtext: 'Total Staf',
      left: 'center',
      top: '38%',
      textStyle: {
        fontSize: 28,
        fontWeight: 'bold',
        color: '#0f172a',
        fontFamily: 'Plus Jakarta Sans',
      },
      subtextStyle: {
        fontSize: 12,
        color: '#94a3b8',
        fontWeight: '600',
      },
    },
    legend: {
      bottom: '0',
      left: 'center',
      icon: 'circle',
      textStyle: {
        fontFamily: 'Plus Jakarta Sans',
        fontSize: 11,
        color: '#64748b',
      },
      itemGap: 12,
    },
    series: [
      {
        name: 'Departemen',
        type: 'pie',
        radius: ['52%', '76%'],
        center: ['50%', '45%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 8,
          borderColor: '#ffffff',
          borderWidth: 3,
        },
        label: { show: false },
        emphasis: {
          label: { show: false },
          scale: true,
          scaleSize: 8,
        },
        data: chartData,
      },
    ],
  }
})

// --- Helpers ---
const avatarColor = (name: string) => {
  const colors = [
    'bg-primary/10 text-primary border-primary/20',
    'bg-sky-50 text-sky-700 border-sky-200',
    'bg-emerald-50 text-emerald-700 border-emerald-200',
    'bg-amber-50 text-amber-700 border-amber-200',
    'bg-purple-50 text-purple-700 border-purple-200',
  ]
  const idx = (name.charCodeAt(0) || 0) % colors.length
  return colors[idx]
}

const getInitials = (name: string) =>
  name
    .split(' ')
    .filter(Boolean)
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2) || 'KR'

// --- Data Fetching ---
const fetchDashboardOverview = async () => {
  isLoading.value = true
  try {
    let url = `/hr/dashboard/overview?range=${timeRange.value}`
    if (masterDataStore.selectedBranchId && masterDataStore.selectedBranchId !== 'ALL_BRANCHES') {
      url += `&branch_id=${masterDataStore.selectedBranchId}`
    }
    const res = await apiClient.get(url)
    if (res.data?.data) {
      const d = res.data.data
      dashboardData.value = {
        summary: {
          total_employees: d.summary?.total_employees || 0,
          present_today: d.summary?.present_today || 0,
          on_time_today: d.summary?.on_time_today || 0,
          late_today: d.summary?.late_today || 0,
          on_leave_today: d.summary?.on_leave_today || 0,
          not_checked_in_today: d.summary?.not_checked_in_today || 0,
          attendance_rate: d.summary?.attendance_rate || 0,
          pending_leaves_count: d.summary?.pending_leaves_count || 0,
          pending_overtimes_count: d.summary?.pending_overtimes_count || 0,
          pending_reimbursements_count: d.summary?.pending_reimbursements_count || 0,
          expiring_contracts_count: d.summary?.expiring_contracts_count || 0,
        },
        attendance_trend: d.attendance_trend || [],
        department_distribution: d.department_distribution || [],
        employment_distribution: d.employment_distribution || [],
        expiring_contracts: d.expiring_contracts || [],
        recent_attendance: d.recent_attendance || [],
        pending_leaves: d.pending_leaves || [],
        pending_overtimes: d.pending_overtimes || [],
      }
    }
  } catch (e: any) {
    if (e.response?.status !== 403) {
      console.error('Failed to fetch HR dashboard overview', e)
      toast.error('Gagal memuat data ringkasan HR')
    }
  } finally {
    isLoading.value = false
  }
}

// --- Quick Actions: Leaves & Overtime Approvals ---
const approveLeave = async (id: string) => {
  if (!confirm('Yakin ingin menyetujui pengajuan cuti ini?')) return
  isActionLoading.value = true
  try {
    await apiClient.put(`/time-off/manage/${id}/approve`, { status: 'APPROVED' })
    toast.success('Pengajuan cuti berhasil disetujui!')
    await fetchDashboardOverview()
  } catch (e: any) {
    toast.error('Gagal menyetujui cuti: ' + (e.response?.data?.message || e.message))
  } finally {
    isActionLoading.value = false
  }
}

const rejectLeave = async (id: string) => {
  const reason = prompt('Alasan penolakan cuti:')
  if (!reason) return
  isActionLoading.value = true
  try {
    await apiClient.put(`/time-off/manage/${id}/approve`, {
      status: 'REJECTED',
      reject_reason: reason,
    })
    toast.success('Pengajuan cuti ditolak.')
    await fetchDashboardOverview()
  } catch (e: any) {
    toast.error('Gagal menolak cuti: ' + (e.response?.data?.message || e.message))
  } finally {
    isActionLoading.value = false
  }
}

const approveOvertime = async (id: string) => {
  if (!confirm('Yakin ingin menyetujui pengajuan lembur ini?')) return
  isActionLoading.value = true
  try {
    await apiClient.put(`/overtimes/${id}/status`, { status: 'APPROVED' })
    toast.success('Pengajuan lembur disetujui!')
    await fetchDashboardOverview()
  } catch (e: any) {
    toast.error('Gagal memproses lembur: ' + (e.response?.data?.message || e.message))
  } finally {
    isActionLoading.value = false
  }
}

const rejectOvertime = async (id: string) => {
  const reason = prompt('Alasan penolakan lembur:')
  if (!reason) return
  isActionLoading.value = true
  try {
    await apiClient.put(`/overtimes/${id}/status`, { status: 'REJECTED', reject_reason: reason })
    toast.success('Pengajuan lembur ditolak.')
    await fetchDashboardOverview()
  } catch (e: any) {
    toast.error('Gagal memproses lembur: ' + (e.response?.data?.message || e.message))
  } finally {
    isActionLoading.value = false
  }
}

// Watchers
watch(
  () => masterDataStore.selectedBranchId,
  () => {
    fetchDashboardOverview()
  },
)

watch(timeRange, () => {
  fetchDashboardOverview()
})

onMounted(() => {
  fetchDashboardOverview()
})
</script>

<template>
  <div class="space-y-8 pb-14 max-w-[1600px] mx-auto animate-fade-in">
    <!-- Header Page: Dynamic Greeting + Command Filter Bar -->
    <div
      class="bg-white/80 backdrop-blur-md p-6 sm:p-8 rounded-3xl border border-slate-200/80 shadow-sm flex flex-col lg:flex-row lg:items-center justify-between gap-6 relative overflow-hidden"
    >
      <div class="relative z-10">
        <div class="flex items-center gap-2.5 text-xs font-extrabold uppercase tracking-wider text-primary">
          <span class="flex h-2.5 w-2.5 rounded-full bg-primary animate-ping"></span>
          <span class="inline-flex items-center gap-1.5 bg-primary/10 px-3 py-1 rounded-full text-primary border border-primary/20">
            <Activity class="w-3.5 h-3.5" />
            Live HR Intelligence • {{ activeBranchName }}
          </span>
        </div>
        <h1 class="text-2xl sm:text-3xl font-black text-slate-900 tracking-tight mt-3">
          {{ greetingMessage }}, {{ userName }}! 👋
        </h1>
        <p class="text-sm font-medium text-slate-500 mt-1 flex items-center gap-2">
          <span>{{ currentFormattedDate }}</span>
          <span class="text-slate-300">•</span>
          <span>Pantauan real-time operasional presensi & manajemen talenta</span>
        </p>
      </div>

      <!-- Action & Filter Bar -->
      <div class="flex flex-wrap items-center gap-3 relative z-10">
        <!-- Range Filter -->
        <div class="flex items-center gap-1 bg-slate-100 p-1.5 rounded-2xl border border-slate-200/60 shrink-0">
          <button
            v-for="range in [
              { id: '7d', label: '7 Hari Terakhir' },
              { id: '30d', label: '30 Hari' },
              { id: 'year', label: 'Tahun Ini' },
            ]"
            :key="range.id"
            @click="timeRange = range.id as '7d' | '30d' | 'year'"
            class="px-4 py-2 text-xs font-extrabold rounded-xl transition-all duration-200"
            :class="[
              timeRange === range.id
                ? 'bg-white text-primary shadow-sm ring-1 ring-slate-200/60 font-black'
                : 'text-slate-500 hover:text-slate-800',
            ]"
          >
            {{ range.label }}
          </button>
        </div>

        <!-- Refresh Button -->
        <button
          @click="fetchDashboardOverview"
          :disabled="isLoading"
          class="p-2.5 bg-slate-100 hover:bg-slate-200/70 text-slate-600 rounded-2xl border border-slate-200/60 transition-all active:scale-95 disabled:opacity-50"
          title="Segarkan Data"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': isLoading }" />
        </button>
      </div>

      <!-- Background Glow Element -->
      <div class="absolute -right-12 -top-12 w-64 h-64 bg-primary/5 rounded-full blur-3xl pointer-events-none"></div>
    </div>

    <!-- Quick Action Launcher Hub -->
    <div class="bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 rounded-3xl p-5 text-white shadow-xl shadow-slate-900/5 border border-slate-800 flex flex-wrap items-center justify-between gap-4">
      <div class="flex items-center gap-3">
        <div class="p-2.5 bg-white/10 rounded-2xl backdrop-blur-md">
          <Sparkles class="w-5 h-5 text-amber-400" />
        </div>
        <div>
          <h4 class="text-sm font-extrabold tracking-wide">Aksi Cepat Manajemen HR</h4>
          <p class="text-xs text-slate-300">Pintasan instan untuk administrasi harian karyawan & presensi</p>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <router-link
          to="/employees"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all backdrop-blur-md border border-white/10 active:scale-95"
        >
          <UserPlus class="w-3.5 h-3.5 text-primary-300" />
          Karyawan Baru
        </router-link>
        <router-link
          to="/attendance"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all backdrop-blur-md border border-white/10 active:scale-95"
        >
          <Clock class="w-3.5 h-3.5 text-emerald-400" />
          Presensi Manual
        </router-link>
        <router-link
          to="/announcements"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all backdrop-blur-md border border-white/10 active:scale-95"
        >
          <Mail class="w-3.5 h-3.5 text-sky-400" />
          Buat Pengumuman
        </router-link>
        <router-link
          to="/pkwt"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl text-xs font-bold transition-all backdrop-blur-md border border-white/10 active:scale-95"
        >
          <FileCheck2 class="w-3.5 h-3.5 text-amber-400" />
          Surat PKWT
        </router-link>
        <router-link
          to="/analytics"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 bg-primary hover:bg-primary/90 text-white rounded-xl text-xs font-bold transition-all shadow-md shadow-primary/20 active:scale-95"
        >
          <Sparkles class="w-3.5 h-3.5" />
          AI Analytics
        </router-link>
      </div>
    </div>

    <!-- 5 Hero KPI Metric Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-5">
      <!-- Card 1: Total Karyawan -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Total Karyawan</span>
            <div class="p-2.5 bg-blue-50 text-blue-600 rounded-2xl group-hover:scale-110 transition-transform">
              <Users class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-16 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ dashboardData.summary.total_employees }}</span>
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span class="text-slate-500 font-semibold">Staf Aktif</span>
          <router-link to="/employees" class="text-primary font-bold hover:underline inline-flex items-center gap-0.5">
            Daftar <ChevronRight class="w-3.5 h-3.5" />
          </router-link>
        </div>
      </div>

      <!-- Card 2: Hadir Hari Ini -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Hadir Hari Ini</span>
            <div class="p-2.5 bg-emerald-50 text-emerald-600 rounded-2xl group-hover:scale-110 transition-transform">
              <CheckCircle2 class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2 flex items-baseline gap-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-16 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ dashboardData.summary.present_today }}</span>
            </h3>
            <span class="text-xs font-bold text-slate-400">/ {{ dashboardData.summary.total_employees }}</span>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 space-y-1.5">
          <div class="flex items-center justify-between text-xs font-bold">
            <span class="text-slate-500">Tingkat Kehadiran</span>
            <span class="text-emerald-600">{{ ((dashboardData.summary?.attendance_rate) || 0).toFixed(1) }}%</span>
          </div>
          <div class="w-full bg-slate-100 h-1.5 rounded-full overflow-hidden">
            <div
              class="bg-emerald-500 h-full rounded-full transition-all duration-700"
              :style="{ width: `${dashboardData.summary?.attendance_rate || 0}%` }"
            ></div>
          </div>
        </div>
      </div>

      <!-- Card 3: Ketepatan Waktu -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Tepat Waktu</span>
            <div class="p-2.5 bg-teal-50 text-teal-600 rounded-2xl group-hover:scale-110 transition-transform">
              <Clock class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-16 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ dashboardData.summary.on_time_today }}</span>
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span class="text-amber-600 font-bold flex items-center gap-1">
            <AlertCircle class="w-3.5 h-3.5" />
            {{ dashboardData.summary.late_today }} Terlambat
          </span>
          <router-link to="/attendance" class="text-primary font-bold hover:underline">Log</router-link>
        </div>
      </div>

      <!-- Card 4: Pending Approvals -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">Review Pending</span>
            <div
              class="p-2.5 rounded-2xl group-hover:scale-110 transition-transform"
              :class="totalPendingReviewCount > 0 ? 'bg-amber-50 text-amber-600' : 'bg-slate-50 text-slate-400'"
            >
              <CalendarClock class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-16 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ totalPendingReviewCount }}</span>
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span class="text-slate-500 font-semibold">
            {{ dashboardData.summary.pending_leaves_count }} Cuti • {{ dashboardData.summary.pending_overtimes_count }} Lembur
          </span>
          <button @click="activeTab = 'leaves'" class="text-primary font-bold hover:underline">Proses</button>
        </div>
      </div>

      <!-- Card 5: Kontrak Berakhir (PKWT) -->
      <div
        class="group bg-white p-5 rounded-3xl border border-slate-200 shadow-sm hover:shadow-md transition-all duration-300 relative overflow-hidden flex flex-col justify-between"
      >
        <div>
          <div class="flex items-start justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-400">PKWT &lt; 30 Hari</span>
            <div
              class="p-2.5 rounded-2xl group-hover:scale-110 transition-transform"
              :class="dashboardData.summary.expiring_contracts_count > 0 ? 'bg-rose-50 text-rose-600' : 'bg-slate-50 text-slate-400'"
            >
              <ShieldAlert class="w-5 h-5" />
            </div>
          </div>
          <div class="mt-2">
            <h3 class="text-3xl font-black text-slate-900 tracking-tight">
              <span v-if="isLoading" class="inline-block w-16 h-8 bg-slate-100 rounded-lg animate-pulse"></span>
              <span v-else>{{ dashboardData.summary.expiring_contracts_count }}</span>
            </h3>
          </div>
        </div>
        <div class="mt-4 pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
          <span
            :class="dashboardData.summary.expiring_contracts_count > 0 ? 'text-rose-600 font-bold' : 'text-slate-400 font-semibold'"
          >
            {{ dashboardData.summary.expiring_contracts_count > 0 ? 'Perlu evaluasi' : 'Semua aman' }}
          </span>
          <button @click="activeTab = 'contracts'" class="text-primary font-bold hover:underline">Detail</button>
        </div>
      </div>
    </div>

    <!-- Middle Section: Interactive ECharts (Trend & Department Distribution) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Main Trend Chart (7 Cols) -->
      <div class="lg:col-span-7 bg-white rounded-3xl border border-slate-200 p-6 sm:p-7 shadow-sm flex flex-col justify-between">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <div class="flex items-center gap-2">
              <h3 class="text-lg font-black text-slate-900 tracking-tight">Tren Presensi & Disiplin</h3>
              <span class="px-2.5 py-0.5 rounded-full text-[10px] font-extrabold bg-emerald-100 text-emerald-700">Data Riil</span>
            </div>
            <p class="text-xs text-slate-500 mt-1">
              Komparasi jumlah kehadiran tepat waktu, terlambat, serta pengajuan cuti & izin.
            </p>
          </div>
          <router-link
            to="/attendance"
            class="inline-flex items-center gap-1.5 text-xs font-extrabold text-primary hover:text-primary/80 bg-primary/5 hover:bg-primary/10 px-3.5 py-2 rounded-xl transition-all self-start sm:self-auto"
          >
            Rekap Lengkap <ArrowRight class="w-3.5 h-3.5" />
          </router-link>
        </div>

        <div class="h-[340px] w-full relative">
          <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-white/70 backdrop-blur-sm z-10">
            <RefreshCw class="w-6 h-6 animate-spin text-primary" />
          </div>
          <v-chart class="w-full h-full" :option="trendChartOption" autoresize />
        </div>
      </div>

      <!-- Department Distribution Chart (5 Cols) -->
      <div class="lg:col-span-5 bg-white rounded-3xl border border-slate-200 p-6 sm:p-7 shadow-sm flex flex-col justify-between">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-lg font-black text-slate-900 tracking-tight">Distribusi Departemen</h3>
            <p class="text-xs text-slate-500 mt-1">Sebaran talenta aktif berdasarkan divisi kerja</p>
          </div>
          <router-link to="/organization" class="text-xs font-bold text-slate-400 hover:text-slate-600">
            Kelola
          </router-link>
        </div>

        <div class="h-[340px] w-full relative">
          <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-white/70 backdrop-blur-sm z-10">
            <RefreshCw class="w-6 h-6 animate-spin text-primary" />
          </div>
          <v-chart class="w-full h-full" :option="deptDistributionOption" autoresize />
        </div>
      </div>
    </div>

    <!-- Status Kehadiran Hari Ini Breakdown & Employment Types Bar -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Breakdown Kehadiran Hari Ini (8 Cols) -->
      <div class="lg:col-span-8 bg-white rounded-3xl border border-slate-200 p-6 sm:p-7 shadow-sm">
        <div class="flex items-center justify-between mb-5">
          <div>
            <h3 class="text-lg font-black text-slate-900 tracking-tight">Status Presensi Hari Ini</h3>
            <p class="text-xs text-slate-500 mt-0.5">Ringkasan kedisiplinan dan status staf per hari ini</p>
          </div>
          <span class="text-xs font-bold text-slate-400">Pembaruan otomatis</span>
        </div>

        <!-- 4 Sub-Boxes -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div class="bg-emerald-50/70 border border-emerald-100 p-4 rounded-2xl">
            <div class="flex items-center gap-2 text-emerald-700 text-xs font-bold">
              <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
              Tepat Waktu
            </div>
            <p class="text-2xl font-black text-emerald-900 mt-2">{{ dashboardData.summary.on_time_today }}</p>
            <p class="text-[11px] text-emerald-600 font-semibold mt-0.5">
              {{ dashboardData.summary.total_employees > 0 ? ((dashboardData.summary.on_time_today / dashboardData.summary.total_employees) * 100).toFixed(1) : 0 }}% dari total
            </p>
          </div>

          <div class="bg-amber-50/70 border border-amber-100 p-4 rounded-2xl">
            <div class="flex items-center gap-2 text-amber-700 text-xs font-bold">
              <span class="w-2 h-2 rounded-full bg-amber-500"></span>
              Terlambat
            </div>
            <p class="text-2xl font-black text-amber-900 mt-2">{{ dashboardData.summary.late_today }}</p>
            <p class="text-[11px] text-amber-600 font-semibold mt-0.5">
              {{ dashboardData.summary.total_employees > 0 ? ((dashboardData.summary.late_today / dashboardData.summary.total_employees) * 100).toFixed(1) : 0 }}% dari total
            </p>
          </div>

          <div class="bg-sky-50/70 border border-sky-100 p-4 rounded-2xl">
            <div class="flex items-center gap-2 text-sky-700 text-xs font-bold">
              <span class="w-2 h-2 rounded-full bg-sky-500"></span>
              Cuti & Izin
            </div>
            <p class="text-2xl font-black text-sky-900 mt-2">{{ dashboardData.summary.on_leave_today }}</p>
            <p class="text-[11px] text-sky-600 font-semibold mt-0.5">Telah disetujui</p>
          </div>

          <div class="bg-slate-50 border border-slate-200/80 p-4 rounded-2xl">
            <div class="flex items-center gap-2 text-slate-600 text-xs font-bold">
              <span class="w-2 h-2 rounded-full bg-slate-400"></span>
              Belum Hadir
            </div>
            <p class="text-2xl font-black text-slate-800 mt-2">{{ dashboardData.summary.not_checked_in_today }}</p>
            <p class="text-[11px] text-slate-500 font-semibold mt-0.5">Belum tap clock-in</p>
          </div>
        </div>

        <!-- Segmented Proportion Bar -->
        <div class="mt-6 space-y-2">
          <div class="w-full h-3.5 bg-slate-100 rounded-full overflow-hidden flex gap-0.5 p-0.5">
            <div
              class="bg-emerald-500 rounded-l-full transition-all duration-700"
              :style="{ width: `${dashboardData.summary.total_employees > 0 ? (dashboardData.summary.on_time_today / dashboardData.summary.total_employees) * 100 : 0}%` }"
              title="Tepat Waktu"
            ></div>
            <div
              class="bg-amber-400 transition-all duration-700"
              :style="{ width: `${dashboardData.summary.total_employees > 0 ? (dashboardData.summary.late_today / dashboardData.summary.total_employees) * 100 : 0}%` }"
              title="Terlambat"
            ></div>
            <div
              class="bg-sky-400 transition-all duration-700"
              :style="{ width: `${dashboardData.summary.total_employees > 0 ? (dashboardData.summary.on_leave_today / dashboardData.summary.total_employees) * 100 : 0}%` }"
              title="Cuti/Izin"
            ></div>
            <div
              class="bg-slate-300 rounded-r-full flex-1 transition-all duration-700"
              title="Belum Hadir"
            ></div>
          </div>
        </div>
      </div>

      <!-- Tipe Kepegawaian (4 Cols) -->
      <div class="lg:col-span-4 bg-white rounded-3xl border border-slate-200 p-6 sm:p-7 shadow-sm flex flex-col justify-between">
        <div>
          <h3 class="text-lg font-black text-slate-900 tracking-tight">Status Kepegawaian</h3>
          <p class="text-xs text-slate-500 mt-0.5">Komposisi kontrak dan jenis status staf</p>

          <div class="mt-5 space-y-3.5">
            <div
              v-for="st in dashboardData.employment_distribution"
              :key="st.status"
              class="flex items-center justify-between p-3 rounded-2xl bg-slate-50 border border-slate-100"
            >
              <div class="flex items-center gap-2.5">
                <div class="p-2 bg-white rounded-xl shadow-xs text-primary border border-slate-100">
                  <Briefcase class="w-4 h-4" />
                </div>
                <span class="text-xs font-bold text-slate-700">{{ st.status }}</span>
              </div>
              <span class="text-xs font-black text-slate-900 bg-white px-3 py-1 rounded-xl shadow-xs border border-slate-100">
                {{ st.count }} orang
              </span>
            </div>
          </div>
        </div>

        <router-link
          to="/employees"
          class="mt-6 w-full text-center py-2.5 text-xs font-bold text-primary hover:bg-primary/5 rounded-2xl transition-all border border-dashed border-primary/20"
        >
          Lihat Manajemen Karyawan
        </router-link>
      </div>
    </div>

    <!-- Bottom Interactive Tabs Section: Live Stream, Approvals, & PKWT Warnings -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
      <!-- Tab Header Navigation -->
      <div class="p-6 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="flex flex-wrap items-center gap-2">
          <button
            @click="activeTab = 'attendance'"
            class="px-4 py-2.5 rounded-2xl text-xs font-extrabold transition-all flex items-center gap-2"
            :class="[
              activeTab === 'attendance'
                ? 'bg-slate-900 text-white shadow-md'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70',
            ]"
          >
            <Clock class="w-3.5 h-3.5" />
            Live Presensi Hari Ini
            <span class="px-2 py-0.5 rounded-full text-[10px] bg-white/20 text-white font-black">
              {{ (dashboardData.recent_attendance || []).length }}
            </span>
          </button>

          <button
            @click="activeTab = 'leaves'"
            class="px-4 py-2.5 rounded-2xl text-xs font-extrabold transition-all flex items-center gap-2"
            :class="[
              activeTab === 'leaves'
                ? 'bg-slate-900 text-white shadow-md'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70',
            ]"
          >
            <CalendarClock class="w-3.5 h-3.5" />
            Pengajuan Cuti & Izin
            <span
              v-if="(dashboardData.summary?.pending_leaves_count || 0) > 0"
              class="px-2 py-0.5 rounded-full text-[10px] bg-amber-500 text-white font-black"
            >
              {{ dashboardData.summary?.pending_leaves_count || 0 }}
            </span>
          </button>

          <button
            @click="activeTab = 'overtime'"
            class="px-4 py-2.5 rounded-2xl text-xs font-extrabold transition-all flex items-center gap-2"
            :class="[
              activeTab === 'overtime'
                ? 'bg-slate-900 text-white shadow-md'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70',
            ]"
          >
            <Timer class="w-3.5 h-3.5" />
            Pengajuan Lembur
            <span
              v-if="(dashboardData.summary?.pending_overtimes_count || 0) > 0"
              class="px-2 py-0.5 rounded-full text-[10px] bg-sky-500 text-white font-black"
            >
              {{ dashboardData.summary?.pending_overtimes_count || 0 }}
            </span>
          </button>

          <button
            @click="activeTab = 'contracts'"
            class="px-4 py-2.5 rounded-2xl text-xs font-extrabold transition-all flex items-center gap-2"
            :class="[
              activeTab === 'contracts'
                ? 'bg-slate-900 text-white shadow-md'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70',
            ]"
          >
            <ShieldAlert class="w-3.5 h-3.5" />
            Monitoring PKWT
            <span
              v-if="(dashboardData.summary?.expiring_contracts_count || 0) > 0"
              class="px-2 py-0.5 rounded-full text-[10px] bg-rose-500 text-white font-black"
            >
              {{ dashboardData.summary?.expiring_contracts_count || 0 }}
            </span>
          </button>
        </div>

        <!-- Filter Sub-Buttons for Attendance Tab -->
        <div v-if="activeTab === 'attendance'" class="flex items-center gap-1.5 bg-slate-100 p-1 rounded-xl">
          <button
            v-for="flt in [
              { id: 'ALL', label: 'Semua' },
              { id: 'ON_TIME', label: 'Tepat Waktu' },
              { id: 'LATE', label: 'Terlambat' },
            ]"
            :key="flt.id"
            @click="attendanceStatusFilter = flt.id as any"
            class="px-3 py-1 text-[11px] font-bold rounded-lg transition-all"
            :class="attendanceStatusFilter === flt.id ? 'bg-white text-slate-900 shadow-xs' : 'text-slate-500'"
          >
            {{ flt.label }}
          </button>
        </div>
      </div>

      <!-- Tab Content Area -->
      <div class="p-6">
        <!-- TAB 1: Live Attendance -->
        <div v-if="activeTab === 'attendance'">
          <div v-if="isLoading" class="space-y-4">
            <div v-for="i in 4" :key="i" class="flex items-center gap-4 animate-pulse">
              <div class="w-10 h-10 rounded-2xl bg-slate-100"></div>
              <div class="flex-1 space-y-2">
                <div class="h-4 bg-slate-100 rounded w-1/4"></div>
                <div class="h-3 bg-slate-100 rounded w-1/3"></div>
              </div>
            </div>
          </div>

          <div v-else-if="filteredRecentAttendance.length === 0" class="py-14 text-center text-slate-400">
            <UserCheck class="w-12 h-12 mx-auto mb-3 opacity-20" />
            <p class="text-sm font-bold">Belum ada catatan log masuk pada filter ini.</p>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="bg-slate-50 text-[11px] font-extrabold text-slate-400 uppercase tracking-wider">
                  <th class="px-5 py-3.5 rounded-l-2xl">Karyawan</th>
                  <th class="px-5 py-3.5">Departemen</th>
                  <th class="px-5 py-3.5">Jam Masuk</th>
                  <th class="px-5 py-3.5">Catatan</th>
                  <th class="px-5 py-3.5 text-right rounded-r-2xl">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="att in filteredRecentAttendance"
                  :key="att.id"
                  class="group hover:bg-slate-50/80 transition-colors"
                >
                  <td class="px-5 py-4">
                    <div class="flex items-center gap-3">
                      <div
                        class="w-10 h-10 rounded-2xl flex items-center justify-center text-xs font-black border"
                        :class="avatarColor(att.name)"
                      >
                        {{ getInitials(att.name) }}
                      </div>
                      <div>
                        <p class="font-extrabold text-slate-900 text-sm group-hover:text-primary transition-colors">
                          {{ att.name }}
                        </p>
                        <p class="text-xs text-slate-400 font-medium">{{ att.job_position }}</p>
                      </div>
                    </div>
                  </td>
                  <td class="px-5 py-4 text-xs font-bold text-slate-600">
                    <span class="inline-flex items-center gap-1 bg-slate-100 px-2.5 py-1 rounded-lg">
                      <Building2 class="w-3 h-3 text-slate-400" />
                      {{ att.department }}
                    </span>
                  </td>
                  <td class="px-5 py-4 text-sm font-black text-slate-900">
                    {{ att.clock_in_time }}
                  </td>
                  <td class="px-5 py-4 text-xs text-slate-500 max-w-xs truncate">
                    {{ att.notes || '-' }}
                  </td>
                  <td class="px-5 py-4 text-right">
                    <span
                      class="px-3 py-1 rounded-xl text-xs font-extrabold inline-flex items-center gap-1"
                      :class="
                        att.status === 'Tepat Waktu'
                          ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                          : 'bg-amber-50 text-amber-700 border border-amber-200'
                      "
                    >
                      <span
                        class="w-1.5 h-1.5 rounded-full"
                        :class="att.status === 'Tepat Waktu' ? 'bg-emerald-500' : 'bg-amber-500'"
                      ></span>
                      {{ att.status }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- TAB 2: Pending Leaves -->
        <div v-if="activeTab === 'leaves'">
          <div v-if="(dashboardData.pending_leaves || []).length === 0" class="py-14 text-center text-slate-400">
            <CheckCircle2 class="w-12 h-12 mx-auto mb-3 opacity-20 text-emerald-500" />
            <p class="text-sm font-bold">Semua pengajuan cuti telah diproses!</p>
            <router-link to="/leaves" class="text-xs font-bold text-primary mt-2 inline-block hover:underline">
              Buka Riwayat Cuti
            </router-link>
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            <div
              v-for="leave in (dashboardData.pending_leaves || [])"
              :key="leave.id"
              class="bg-slate-50/70 border border-slate-200/80 rounded-3xl p-5 hover:shadow-md transition-all flex flex-col justify-between"
            >
              <div>
                <div class="flex items-start justify-between">
                  <div>
                    <h4 class="font-black text-slate-900 text-sm">{{ leave.name }}</h4>
                    <span class="text-xs font-bold text-primary bg-primary/10 px-2.5 py-0.5 rounded-md inline-block mt-1">
                      {{ leave.type }} • {{ leave.days }} Hari
                    </span>
                  </div>
                  <span class="text-[11px] font-bold text-slate-400 bg-white px-2 py-0.5 rounded-md border border-slate-100">
                    {{ leave.time_ago }}
                  </span>
                </div>

                <div class="mt-4 p-3 bg-white rounded-2xl border border-slate-100 text-xs space-y-1">
                  <p class="text-slate-500"><b class="text-slate-700">Periode:</b> {{ leave.start_date }} s/d {{ leave.end_date }}</p>
                  <p class="text-slate-500"><b class="text-slate-700">Alasan:</b> {{ leave.reason || 'Tidak ada keterangan' }}</p>
                </div>
              </div>

              <div class="flex gap-2 mt-5">
                <button
                  @click="approveLeave(leave.id)"
                  :disabled="isActionLoading"
                  class="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl py-2 text-xs font-black transition-all shadow-sm active:scale-95 disabled:opacity-50 flex items-center justify-center gap-1"
                >
                  <Check class="w-3.5 h-3.5" />
                  Setujui
                </button>
                <button
                  @click="rejectLeave(leave.id)"
                  :disabled="isActionLoading"
                  class="flex-1 bg-white hover:bg-rose-50 text-rose-600 border border-rose-200 rounded-xl py-2 text-xs font-black transition-all active:scale-95 disabled:opacity-50 flex items-center justify-center gap-1"
                >
                  <X class="w-3.5 h-3.5" />
                  Tolak
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- TAB 3: Pending Overtime -->
        <div v-if="activeTab === 'overtime'">
          <div v-if="(dashboardData.pending_overtimes || []).length === 0" class="py-14 text-center text-slate-400">
            <CheckCircle2 class="w-12 h-12 mx-auto mb-3 opacity-20 text-emerald-500" />
            <p class="text-sm font-bold">Semua pengajuan lembur telah diproses!</p>
            <router-link to="/overtimes" class="text-xs font-bold text-primary mt-2 inline-block hover:underline">
              Buka Riwayat Lembur
            </router-link>
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            <div
              v-for="ot in (dashboardData.pending_overtimes || [])"
              :key="ot.id"
              class="bg-slate-50/70 border border-slate-200/80 rounded-3xl p-5 hover:shadow-md transition-all flex flex-col justify-between"
            >
              <div>
                <div class="flex items-start justify-between">
                  <div>
                    <h4 class="font-black text-slate-900 text-sm">{{ ot.name }}</h4>
                    <span class="text-xs font-bold text-sky-600 bg-sky-100 px-2.5 py-0.5 rounded-md inline-block mt-1">
                      {{ ot.duration_hrs }} Jam Lembur
                    </span>
                  </div>
                  <span class="text-[11px] font-bold text-slate-400 bg-white px-2 py-0.5 rounded-md border border-slate-100">
                    {{ ot.time_ago }}
                  </span>
                </div>

                <div class="mt-4 p-3 bg-white rounded-2xl border border-slate-100 text-xs space-y-1">
                  <p class="text-slate-500"><b class="text-slate-700">Tanggal:</b> {{ ot.date }}</p>
                  <p class="text-slate-500"><b class="text-slate-700">Tugas:</b> {{ ot.reason || 'Tidak ada keterangan' }}</p>
                </div>
              </div>

              <div class="flex gap-2 mt-5">
                <button
                  @click="approveOvertime(ot.id)"
                  :disabled="isActionLoading"
                  class="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl py-2 text-xs font-black transition-all shadow-sm active:scale-95 disabled:opacity-50 flex items-center justify-center gap-1"
                >
                  <Check class="w-3.5 h-3.5" />
                  Setujui
                </button>
                <button
                  @click="rejectOvertime(ot.id)"
                  :disabled="isActionLoading"
                  class="flex-1 bg-white hover:bg-rose-50 text-rose-600 border border-rose-200 rounded-xl py-2 text-xs font-black transition-all active:scale-95 disabled:opacity-50 flex items-center justify-center gap-1"
                >
                  <X class="w-3.5 h-3.5" />
                  Tolak
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- TAB 4: Expiring PKWT Contracts -->
        <div v-if="activeTab === 'contracts'">
          <div v-if="(dashboardData.expiring_contracts || []).length === 0" class="py-14 text-center text-slate-400">
            <ShieldAlert class="w-12 h-12 mx-auto mb-3 opacity-20 text-emerald-500" />
            <p class="text-sm font-bold">Tidak ada kontrak karyawan yang akan berakhir dalam 60 hari ke depan.</p>
            <router-link to="/pkwt" class="text-xs font-bold text-primary mt-2 inline-block hover:underline">
              Kelola Arsip PKWT
            </router-link>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="bg-slate-50 text-[11px] font-extrabold text-slate-400 uppercase tracking-wider">
                  <th class="px-5 py-3.5 rounded-l-2xl">Karyawan</th>
                  <th class="px-5 py-3.5">Departemen & Jabatan</th>
                  <th class="px-5 py-3.5">Akhir Kontrak</th>
                  <th class="px-5 py-3.5">Sisa Waktu</th>
                  <th class="px-5 py-3.5 text-right rounded-r-2xl">Tindakan</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr
                  v-for="contract in (dashboardData.expiring_contracts || [])"
                  :key="contract.id"
                  class="hover:bg-slate-50/80 transition-colors"
                >
                  <td class="px-5 py-4">
                    <p class="font-extrabold text-slate-900 text-sm">{{ contract.name }}</p>
                    <p class="text-xs text-slate-400 font-mono">{{ contract.employee_id_number }}</p>
                  </td>
                  <td class="px-5 py-4 text-xs font-bold text-slate-700">
                    <p>{{ contract.job_position }}</p>
                    <p class="text-slate-400 font-normal">{{ contract.department }}</p>
                  </td>
                  <td class="px-5 py-4 text-sm font-bold text-slate-900">
                    {{ contract.end_date }}
                  </td>
                  <td class="px-5 py-4">
                    <span
                      class="px-3 py-1 rounded-xl text-xs font-black inline-flex items-center gap-1.5"
                      :class="
                        contract.days_remaining <= 15
                          ? 'bg-rose-100 text-rose-700 border border-rose-200'
                          : 'bg-amber-100 text-amber-800 border border-amber-200'
                      "
                    >
                      <AlertTriangle class="w-3.5 h-3.5" />
                      {{ contract.days_remaining }} Hari Lagi
                    </span>
                  </td>
                  <td class="px-5 py-4 text-right">
                    <router-link
                      to="/pkwt"
                      class="inline-flex items-center gap-1 bg-primary text-white px-3 py-1.5 rounded-xl text-xs font-bold shadow-sm hover:bg-primary/90 transition-all active:scale-95"
                    >
                      Perpanjang PKWT <ArrowRight class="w-3 h-3" />
                    </router-link>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.v-chart {
  font-family: 'Plus Jakarta Sans', sans-serif !important;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fadeIn 0.35s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
