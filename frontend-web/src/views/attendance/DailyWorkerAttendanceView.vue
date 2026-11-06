<script setup lang="ts">
import { ref, onMounted, h, computed, watch } from 'vue'
import { toast } from 'vue-sonner'
import {
  Pencil,
  Trash2,
  Download,
  CalendarClock,
  Plus,
  Filter,
  FileSpreadsheet,
  MapPin,
  Building2,
  Clock,
  ExternalLink,
  User,
  CheckCircle2,
  LogIn,
  LogOut,
  Camera,
  Eye,
  EyeOff,
  Loader2,
} from 'lucide-vue-next'
import apiClient from '@/api/axios'
import { formatAttendanceDateTime, downloadWatermarkedAttendancePhoto } from '@/utils/watermark'
import DataTable from '@/components/DataTable.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { useMasterDataStore } from '@/stores/masterData'
import { useAuthStore } from '@/stores/auth'

const masterData = useMasterDataStore()
const authStore = useAuthStore()

const isLoading = ref(true)

const fmtTime = (timeStr: string, isEndTime = false) => {
  if (!timeStr) return ''

  // Handle HH:mm:ss format
  if (timeStr.includes(':') && !timeStr.includes('-') && !timeStr.includes('T')) {
    const parts = timeStr.split(':')
    const h = (parts[0] || '00').padStart(2, '0')
    const m = (parts[1] || '00').padStart(2, '0')
    if (isEndTime && h === '00' && m === '00') return '24.00'
    return `${h}.${m}`
  }

  const d = new Date(timeStr)
  if (isNaN(d.getTime())) return ''
  let h = d.getHours()
  let m = d.getMinutes()
  if (isEndTime && h === 0 && m === 0) return '24.00'
  return `${String(h).padStart(2, '0')}.${String(m).padStart(2, '0')}`
}
interface AttendanceRecord {
  id: string
  date: string
  name: string
  email: string
  role: string
  checkIn: string
  checkOut: string
  total: string
  status: string
  branch: string
  notes: string
  avatar: string
  clockInPhoto: string
  clockOutPhoto: string
  _raw: any
}

interface AttendanceStats {
  total: number
  onTime: number
  late: number
  permitted: number
  absent: number
  earliestToday: { name: string; timeStr: string; time: Date } | null
}

const displayData = ref<AttendanceRecord[]>([])
const employeeOptions = ref<any[]>([])
const selectedBranch = computed(() => {
  if (!masterData.selectedBranchId || masterData.selectedBranchId === 'ALL_BRANCHES') {
    return 'all'
  }
  return masterData.selectedBranchId
})

watch(selectedBranch, () => {
  fetchAttendance()
})
const filterType = ref<string>('daily')
const selectedMonth = ref<string>(new Date().toISOString().slice(0, 7)) // default current month: YYYY-MM
const selectedYear = ref<string>(new Date().getFullYear().toString())
const selectedDate = ref<string>(new Date().toISOString().split('T')[0] || '') // default today: YYYY-MM-DD
const selectedLimit = ref<string>('100')

const attendanceStats = computed<AttendanceStats>(() => {
  const data = displayData.value
  let onTime = 0,
    late = 0,
    permitted = 0,
    absent = 0
  const todayStr = new Date().toISOString().split('T')[0]
  let earliestToday: { name: string; timeStr: string; time: Date } | null = null

  for (const item of data) {
    const s = (item.status || '').toUpperCase()
    if (s === 'ON_TIME' || s === 'PRESENT') onTime++
    else if (s === 'LATE') late++
    else if (['SICK', 'LEAVE', 'PERMISSION'].includes(s)) permitted++
    else if (s === 'ABSENT') absent++

    // Find earliest today
    if (item._raw.clock_in_time) {
      const dt = new Date(item._raw.clock_in_time)
      const datePart = dt.toISOString().split('T')[0]
      if (datePart === todayStr) {
        if (!earliestToday || dt < earliestToday.time) {
          earliestToday = { name: item.name, timeStr: item.checkIn, time: dt }
        }
      }
    }
  }

  return {
    total: data.length,
    onTime,
    late,
    permitted,
    absent,
    earliestToday,
  }
})

const totalDataLabel = computed(() => {
  if (filterType.value === 'daily') return 'Total Data Harian'
  if (filterType.value === 'month') return 'Total Data Bulanan'
  if (filterType.value === 'year') return 'Total Data Tahunan'
  return 'Total Data'
})

// Card Filter State
export type DailyCardFilterType = 'ALL' | 'ON_TIME' | 'LATE' | 'ABSENT_LEAVE' | 'EARLIEST'
const activeCardFilter = ref<DailyCardFilterType>('ALL')

const toggleCardFilter = (filter: DailyCardFilterType) => {
  if (activeCardFilter.value === filter) {
    activeCardFilter.value = 'ALL'
  } else {
    activeCardFilter.value = filter
  }
}

const filteredTableData = computed(() => {
  const data = displayData.value
  if (activeCardFilter.value === 'ALL') return data
  if (activeCardFilter.value === 'ON_TIME') {
    return data.filter((item) => {
      const s = (item.status || '').toUpperCase()
      return s === 'ON_TIME' || s === 'PRESENT'
    })
  }
  if (activeCardFilter.value === 'LATE') {
    return data.filter((item) => (item.status || '').toUpperCase() === 'LATE')
  }
  if (activeCardFilter.value === 'ABSENT_LEAVE') {
    return data.filter((item) =>
      ['ABSENT', 'SICK', 'LEAVE', 'PERMISSION'].includes((item.status || '').toUpperCase()),
    )
  }
  if (activeCardFilter.value === 'EARLIEST') {
    const earliest = attendanceStats.value.earliestToday
    if (!earliest) return []
    return data.filter((item) => {
      if (!item._raw?.clock_in_time) return false
      const dt = new Date(item._raw.clock_in_time)
      const todayStr = new Date().toISOString().split('T')[0]
      return (
        dt.toISOString().split('T')[0] === todayStr &&
        dt.getTime() === earliest.time.getTime()
      )
    })
  }
  return data
})

const activeFilterLabel = computed(() => {
  switch (activeCardFilter.value) {
    case 'ON_TIME':
      return 'Tepat Waktu'
    case 'LATE':
      return 'Terlambat'
    case 'ABSENT_LEAVE':
      return 'Absen / Cuti'
    case 'EARLIEST':
      return 'Berangkat Terawal Hari Ini'
    default:
      return ''
  }
})

const isModalOpen = ref(false)
const isEditMode = ref(false)
const isSubmitting = ref(false)

const currentForm = ref({
  id: '',
  employee_id: '',
  date: new Date().toISOString().split('T')[0] || '',
  clock_in_time: '08.00',
  clock_out_time: '17.00',
  status: 'PRESENT',
  notes: '',
})

const canManageAttendance = computed(() => {
  const role = authStore.user?.role
  return role === 'superadmin' || role === 'hr_admin' || role === 'manager' || role === 'admin_kasir'
})

// Photo Preview State
const isPreviewOpen = ref(false)
const previewData = ref<{
  url: string
  type: 'IN' | 'OUT'
  name: string
  email: string
  role: string
  branch: string
  timeFormatted: string
  timeShort: string
  date: string
  status: string
  lat: number | null
  lng: number | null
  notes: string
}>({
  url: '',
  type: 'IN',
  name: '',
  email: '',
  role: '',
  branch: '',
  timeFormatted: '',
  timeShort: '',
  date: '',
  status: '',
  lat: null,
  lng: null,
  notes: '',
})

const openPreview = (item: any, type: 'IN' | 'OUT') => {
  const isClockIn = type === 'IN'
  const photoUrl = isClockIn ? item.clockInPhoto : item.clockOutPhoto
  if (!photoUrl) return

  const rawTime = isClockIn ? item._raw?.clock_in_time : item._raw?.clock_out_time
  const formattedTime = rawTime
    ? formatAttendanceDateTime(new Date(rawTime))
    : `${item.date} ${isClockIn ? item.checkIn : item.checkOut}`

  const lat = isClockIn ? item._raw?.clock_in_lat : item._raw?.clock_out_lat
  const lng = isClockIn ? item._raw?.clock_in_long : item._raw?.clock_out_long

  previewData.value = {
    url: `${apiClient.defaults.baseURL}${photoUrl}`,
    type,
    name: item.name || 'Pekerja Harian',
    email: item.email || '-',
    role: item.role || 'Pekerja Harian',
    branch: item.branch || item._raw?.employee?.branch?.name || 'Cabang Kantor',
    timeFormatted: formattedTime,
    timeShort: isClockIn ? item.checkIn : item.checkOut,
    date: item.date || '-',
    status: item.status || 'PRESENT',
    lat: lat ? parseFloat(lat) : null,
    lng: lng ? parseFloat(lng) : null,
    notes: item.notes || item._raw?.notes || '',
  }
  showWatermarkOverlay.value = true
  isPreviewOpen.value = true
}

const showWatermarkOverlay = ref(true)
const isDownloading = ref(false)

const handleDownloadWatermarkedPhoto = async () => {
  if (!previewData.value?.url) return
  isDownloading.value = true
  try {
    const filename = `absensi_pekerja_harian_${previewData.value.type.toLowerCase()}_${previewData.value.name.replace(/[^a-zA-Z0-9]/g, '_')}_${previewData.value.date || 'log'}.jpg`
    const locationText = previewData.value.lat && previewData.value.lng
      ? `GPS: ${previewData.value.lat}, ${previewData.value.lng}`
      : 'Lokasi Presensi Pekerja Harian'

    await downloadWatermarkedAttendancePhoto(
      previewData.value.url,
      {
        timeText: previewData.value.timeFormatted,
        branchName: previewData.value.branch,
        addressText: locationText,
        cropSquare: true,
        targetWidth: 800,
        quality: 0.85,
      },
      filename,
    )
    toast.success('Foto ber-watermark berhasil diunduh')
  } catch (err: any) {
    console.error('Download photo error:', err)
    toast.error('Gagal mengunduh foto ber-watermark')
  } finally {
    isDownloading.value = false
  }
}


// Define Columns for TanStack Table
const columns = computed(() => {
  const cols: any[] = [
    {
      accessorKey: 'date',
      header: 'TANGGAL',
      cell: (info: any) => h('span', { class: 'text-gray-500' }, info.getValue()),
    },
    {
      accessorFn: (row: any) => row.name || '-',
      id: 'name',
      header: 'KARYAWAN',
      cell: ({ row }: any) => {
        const user = row.original
        return h('div', { class: 'flex flex-col' }, [
          h('p', { class: 'font-bold text-gray-900 leading-tight' }, user.name),
          h('p', { class: 'text-[12px] text-gray-400 font-medium' }, user.email),
          h('p', { class: 'text-[12px] text-gray-500' }, user.role),
        ])
      },
    },
    {
      accessorKey: 'checkIn',
      header: 'JAM MASUK',
      cell: (info: any) => h('span', { class: 'font-medium text-gray-700' }, info.getValue()),
    },
    {
      accessorKey: 'checkOut',
      header: 'JAM KELUAR',
      cell: (info: any) => h('span', { class: 'font-medium text-gray-700' }, info.getValue()),
    },
    {
      accessorKey: 'total',
      header: 'TOTAL JAM',
      cell: (info: any) => h('span', { class: 'font-bold text-gray-900' }, info.getValue()),
    },
    {
      accessorKey: 'status',
      header: 'STATUS',
      cell: ({ getValue }: any) => {
        const val = (getValue() as string) || ''
        let classes = 'bg-[#fef3c7] text-[#92400e]' // warning outline
        if (val.toUpperCase() === 'ON_TIME' || val.toUpperCase() === 'PRESENT')
          classes = 'bg-[#dcfce7] text-[#166534]'
        else if (val.toUpperCase() === 'LATE') classes = 'bg-[#ffedd5] text-[#c2410c]'
        else if (['SICK', 'LEAVE', 'PERMISSION'].includes(val.toUpperCase()))
          classes = 'bg-[#e0e7ff] text-[#3730a3]'
        else if (val.toUpperCase() === 'ABSENT')
          classes = 'bg-[#fef2f2] text-[#b91c1c] ring-1 ring-[#b91c1c]/30'
        return h(
          'span',
          { class: `${classes} px-2.5 py-1 rounded-full text-[12px] font-bold` },
          val || 'UNKNOWN',
        )
      },
    },
    {
      accessorKey: 'branch',
      header: 'CABANG',
      cell: (info: any) =>
        h('span', { class: 'text-gray-500 font-medium text-[13px]' }, info.getValue() || '-'),
    },
    {
      accessorKey: 'notes',
      header: 'CATATAN',
      cell: (info: any) => h('span', { class: 'text-gray-500 text-[13px]' }, info.getValue() || '-'),
    },
    {
      id: 'photo',
      header: 'FOTO BUKTI',
      cell: ({ row }: any) => {
        const item = row.original
        return h('div', { class: 'flex items-center gap-2' }, [
          item.clockInPhoto
            ? h(
                'div',
                {
                  onClick: () => openPreview(item, 'IN'),
                  class:
                    'w-10 h-10 rounded-lg overflow-hidden border border-gray-200 cursor-pointer hover:ring-2 hover:ring-emerald-500/30 transition-all bg-gray-50 flex items-center justify-center group relative shadow-xs',
                  title: 'Lihat Foto Masuk (IN)',
                },
                [
                  h('img', {
                    src: `${apiClient.defaults.baseURL}${item.clockInPhoto}`,
                    class: 'w-full h-full object-cover group-hover:scale-110 transition-transform',
                  }),
                  h('div', {
                    class:
                      'absolute inset-0 bg-black/5 opacity-0 group-hover:opacity-100 transition-opacity',
                  }),
                  h(
                    'div',
                    {
                      class:
                        'absolute bottom-0 right-0 bg-emerald-600 text-[8px] text-white px-1 font-bold rounded-tl',
                    },
                    'IN',
                  ),
                ],
              )
            : null,
          item.clockOutPhoto
            ? h(
                'div',
                {
                  onClick: () => openPreview(item, 'OUT'),
                  class:
                    'w-10 h-10 rounded-lg overflow-hidden border border-gray-200 cursor-pointer hover:ring-2 hover:ring-indigo-500/30 transition-all bg-gray-50 flex items-center justify-center group relative shadow-xs',
                  title: 'Lihat Foto Keluar (OUT)',
                },
                [
                  h('img', {
                    src: `${apiClient.defaults.baseURL}${item.clockOutPhoto}`,
                    class: 'w-full h-full object-cover group-hover:scale-110 transition-transform',
                  }),
                  h('div', {
                    class:
                      'absolute inset-0 bg-black/5 opacity-0 group-hover:opacity-100 transition-opacity',
                  }),
                  h(
                    'div',
                    {
                      class:
                        'absolute bottom-0 right-0 bg-indigo-600 text-[8px] text-white px-1 font-bold rounded-tl',
                    },
                    'OUT',
                  ),
                ],
              )
            : null,
          !item.clockInPhoto && !item.clockOutPhoto
            ? h('span', { class: 'text-[11px] text-gray-400 font-medium italic' }, 'Tanpa Foto')
            : null,
        ])
      },
    },
  ]

  if (canManageAttendance.value) {
    cols.push({
      id: 'actions',
      header: 'AKSI',
      cell: ({ row }: any) => {
        const item = row.original
        return h('div', { class: 'flex gap-2' }, [
          h(
            Button,
            {
              variant: 'ghost',
              size: 'sm',
              onClick: () => openEditModal(item),
              class: 'h-8 px-2 text-primary hover:text-primary hover:bg-primary/5',
            },
            () => h(Pencil, { class: 'w-4 h-4' }),
          ),
          h(
            Button,
            {
              variant: 'ghost',
              size: 'sm',
              onClick: () => deleteAttendance(item.id),
              class: 'h-8 px-2 text-red-600 hover:text-red-700 hover:bg-red-50',
            },
            () => h(Trash2, { class: 'w-4 h-4' }),
          ),
        ])
      },
      enableSorting: false,
    })
  }

  return cols
})

const openAddModal = () => {
  isEditMode.value = false
  currentForm.value = {
    id: '',
    employee_id: '',
    date: new Date().toISOString().split('T')[0] || '',
    clock_in_time: '08.00',
    clock_out_time: '17.00',
    status: 'PRESENT',
    notes: '',
  }
  isModalOpen.value = true
}

const openEditModal = (item: any) => {
  isEditMode.value = true

  // parse dates specifically
  const dateStr = item._raw.clock_in_time
    ? new Date(item._raw.clock_in_time).toISOString().split('T')[0] || ''
    : ''
  const inTimeStr = item._raw.clock_in_time ? fmtTime(item._raw.clock_in_time) : ''
  const outTimeStr = item._raw.clock_out_time ? fmtTime(item._raw.clock_out_time) : ''

  currentForm.value = {
    id: item.id,
    employee_id: item._raw.employee_id,
    date: dateStr,
    clock_in_time: inTimeStr || '',
    clock_out_time: outTimeStr || '',
    status: item._raw.status || 'PRESENT',
    notes: item._raw.notes || '',
  }
  isModalOpen.value = true
}

const closeAddModal = () => {
  isModalOpen.value = false
}

const deleteAttendance = async (id: string) => {
  if (!confirm('Hapus log kehadiran ini? Operasi ini permanen.')) return
  try {
    await apiClient.delete(`/attendance/${id}`)
    toast.success('Log kehadiran berhasil dihapus!')
    fetchAttendance()
  } catch (e: any) {
    toast.error('Gagal menghapus: ' + (e.response?.data?.message || e.message))
  }
}

const saveManualData = async () => {
  isSubmitting.value = true
  try {
    // combine Date + HH.mm strings into local Date object
    const inTimeFormatted = (currentForm.value.clock_in_time || '08.00').replace('.', ':')
    const inDateObj = new Date(`${currentForm.value.date}T${inTimeFormatted}:00`)

    let outDateObj: Date | null = null
    if (currentForm.value.clock_out_time) {
      const outTimeFormatted = currentForm.value.clock_out_time.replace('.', ':')
      outDateObj = new Date(`${currentForm.value.date}T${outTimeFormatted}:00`)
    }

    const payload = {
      employee_id: currentForm.value.employee_id,
      clock_in_time: inDateObj.toISOString(),
      clock_out_time: outDateObj ? outDateObj.toISOString() : null,
      status: currentForm.value.status,
      notes: currentForm.value.notes,
    }

    if (isEditMode.value) {
      await apiClient.put(`/attendance/${currentForm.value.id}`, payload)
      toast.success('Kehadiran berhasil diperbarui!')
    } else {
      await apiClient.post('/attendance/manual', payload)
      toast.success('Kehadiran manual berhasil ditambahkan!')
    }
    closeAddModal()
    fetchAttendance()
  } catch (error: any) {
    toast.error('Gagal menyimpan data: ' + (error.response?.data?.message || error.message))
  } finally {
    isSubmitting.value = false
  }
}



const fetchAttendance = async () => {
  try {
    let url = `/attendance/all?limit=${selectedLimit.value}&employment_type=Harian`
    if (selectedBranch.value && selectedBranch.value !== 'all') {
      url += `&branch_id=${selectedBranch.value}`
    }
    if (filterType.value === 'month' && selectedMonth.value) {
      url += `&month=${selectedMonth.value}`
    } else if (filterType.value === 'year' && selectedYear.value) {
      url += `&month=${selectedYear.value}`
    } else if (filterType.value === 'daily' && selectedDate.value) {
      url += `&month=${selectedDate.value}`
    }
    const res = await apiClient.get(url)
    if (res.data?.data && res.data.data.length > 0) {
      displayData.value = res.data.data.map((item: any) => {
        const clockIn = item.clock_in_time ? new Date(item.clock_in_time) : null
        const clockOut = item.clock_out_time ? new Date(item.clock_out_time) : null
        let totalStr = '-'
        if (clockIn && clockOut) {
          const diffMs = clockOut.getTime() - clockIn.getTime()
          const totalH = Math.floor(diffMs / 3600000)
          const totalM = Math.floor((diffMs % 3600000) / 60000)
          totalStr = `${totalH}j ${totalM}m`
        }
        return {
          id: item.id,
          date: clockIn ? clockIn.toLocaleDateString('id-ID') : 'N/A',
          name: item.employee?.first_name || 'Karyawan',
          email: item.employee?.user?.email || '-',
          role: item.employee?.job_position?.title || '-',
          checkIn: clockIn ? fmtTime(item.clock_in_time) : '-',
          checkOut: clockOut ? fmtTime(item.clock_out_time, true) : '-',
          total: totalStr,
          status: item.status || 'PRESENT',
          branch: item.employee?.branch?.name || '-',
          notes: item.notes || '-',
          avatar: `https://i.pravatar.cc/150?u=${item.employee_id}`,
          clockInPhoto: item.clock_in_photo_url,
          clockOutPhoto: item.clock_out_photo_url,
          _raw: item,
        }
      })
    } else {
      displayData.value = []
    }
  } catch (e: any) {
    console.error('Failed to fetch from API', e)
    displayData.value = []
  } finally {
    isLoading.value = false
  }
}

const fetchEmployeesForFilter = async () => {
  try {
    const res = await apiClient.get('/employees?limit=all&employment_type=harian')
    if (res.data?.data) {
      employeeOptions.value = res.data.data
    }
  } catch (error) {
    console.error('Failed fetching core employee references', error)
  }
}

const exportXLSX = async () => {
  try {
    let url = '/attendance/export-xlsx?employment_type=Harian'
    if (selectedBranch.value && selectedBranch.value !== 'all') {
      url += `&branch_id=${selectedBranch.value}`
    }
    if (filterType.value === 'month' && selectedMonth.value) {
      url += `&month=${selectedMonth.value}`
    } else if (filterType.value === 'year' && selectedYear.value) {
      url += `&month=${selectedYear.value}`
    } else if (filterType.value === 'daily' && selectedDate.value) {
      url += `&month=${selectedDate.value}`
    }

    const response = await apiClient.get(url, { responseType: 'blob' })
    const blob = new Blob([response.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })
    const downloadUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.setAttribute('download', `laporan_kehadiran_${new Date().getTime()}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(downloadUrl)
  } catch (error) {
    toast.error('Gagal mengunduh laporan kehadiran.')
  }
}

onMounted(() => {
  fetchAttendance()
  fetchEmployeesForFilter()
  masterData.fetchBranches()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-gray-900">Data Kehadiran Pekerja Harian</h2>
        <p class="text-sm text-gray-500">Kelola riwayat absen manual dan otomatis pekerja harian.</p>
        <p
          class="text-[14px] font-medium text-slate-500 mt-1.5 uppercase tracking-widest flex items-center gap-2"
        >
          Monitoring Presensi & Jam Kerja
        </p>
      </div>

      <div class="flex items-center gap-3 flex-wrap">


        <!-- Type Filter -->
        <Select v-model="filterType" @update:modelValue="fetchAttendance">
          <SelectTrigger
            class="w-[110px] h-10 rounded-xl border-none bg-white shadow-lg shadow-slate-100/50 font-bold px-4 text-slate-600 focus:ring-2 focus:ring-primary/20 transition-all"
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="daily">Harian</SelectItem>
            <SelectItem value="month">Bulanan</SelectItem>
            <SelectItem value="year">Tahunan</SelectItem>
          </SelectContent>
        </Select>

        <!-- Date Filter -->
        <div class="relative">
          <input
            v-if="filterType === 'daily'"
            type="date"
            v-model="selectedDate"
            @change="fetchAttendance"
            class="h-10 border-none bg-white shadow-lg shadow-slate-100/50 text-slate-600 rounded-xl text-[13px] font-bold px-4 focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
          />
          <input
            v-if="filterType === 'month'"
            type="month"
            v-model="selectedMonth"
            @change="fetchAttendance"
            class="h-10 border-none bg-white shadow-lg shadow-slate-100/50 text-slate-600 rounded-xl text-[13px] font-bold px-4 focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
          />
          <input
            v-if="filterType === 'year'"
            type="number"
            min="2000"
            max="2099"
            step="1"
            v-model="selectedYear"
            @change="fetchAttendance"
            class="h-10 w-24 border-none bg-white shadow-lg shadow-slate-100/50 text-slate-600 rounded-xl text-[13px] font-bold px-4 focus:outline-none focus:ring-2 focus:ring-primary/20 transition-all"
          />
        </div>

        <!-- Limit Filter -->
        <Select v-model="selectedLimit" @update:modelValue="fetchAttendance">
          <SelectTrigger
            class="w-[100px] h-10 rounded-xl border-none bg-white shadow-lg shadow-slate-100/50 font-bold px-4 text-slate-600 focus:ring-2 focus:ring-primary/20 transition-all"
          >
            <SelectValue placeholder="Limit" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="50">50 Data</SelectItem>
            <SelectItem value="100">100 Data</SelectItem>
            <SelectItem value="200">200 Data</SelectItem>
            <SelectItem value="500">500 Data</SelectItem>
            <SelectItem value="0">Semua</SelectItem>
          </SelectContent>
        </Select>

        <!-- Export CSV -->
        <Button
          variant="outline"
          @click="exportXLSX"
          class="h-10 px-4 rounded-xl font-bold border-2 border-slate-50 bg-white shadow-lg shadow-slate-100/50 text-slate-600 gap-2 hover:bg-slate-50 transition-all active:scale-95"
        >
          <Download class="w-4 h-4" /> EXPORT
        </Button>

        <Button
          v-if="canManageAttendance"
          @click="openAddModal"
          class="gap-2 shadow-xl shadow-primary/20 rounded-xl h-10 px-5 font-bold bg-primary hover:bg-primary/90 transition-all hover:scale-105 active:scale-95"
        >
          <Plus class="w-4 h-4" /> INPUT MANUAL
        </Button>
      </div>
    </div>

    <!-- Stats Cards (Clickable Filters) -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
      <!-- Card 1: Total Data -->
      <div
        @click="toggleCardFilter('ALL')"
        class="group bg-white p-6 rounded-2xl border transition-all duration-300 cursor-pointer select-none hover:scale-[1.02] active:scale-[0.98] relative overflow-hidden"
        :class="
          activeCardFilter === 'ALL'
            ? 'border-primary ring-2 ring-primary/40 shadow-lg bg-primary/5'
            : 'border-slate-200 shadow-[0_4px_20px_rgba(0,0,0,0.03)] hover:shadow-md'
        "
        title="Klik untuk menampilkan semua data"
      >
        <div class="flex items-start justify-between">
          <p class="text-[11px] font-bold text-slate-500 uppercase tracking-widest leading-tight">
            {{ totalDataLabel }}
          </p>
          <div
            class="p-2 rounded-xl transition-transform group-hover:scale-110"
            :class="activeCardFilter === 'ALL' ? 'bg-primary text-white' : 'bg-primary/5 text-primary'"
          >
            <FileSpreadsheet class="w-4 h-4" />
          </div>
        </div>
        <div>
          <h3 class="text-2xl font-black text-slate-900 mt-3">{{ attendanceStats.total }}</h3>
          <div class="flex items-center justify-between mt-1">
            <p class="text-[10px] font-bold text-slate-400 uppercase">Records</p>
            <span
              v-if="activeCardFilter === 'ALL'"
              class="text-[9px] font-black text-primary bg-primary/10 px-1.5 py-0.5 rounded-md"
            >
              Semua
            </span>
          </div>
        </div>
      </div>

      <!-- Card 2: Tepat Waktu -->
      <div
        @click="toggleCardFilter('ON_TIME')"
        class="group bg-white p-6 rounded-2xl border transition-all duration-300 cursor-pointer select-none hover:scale-[1.02] active:scale-[0.98] relative overflow-hidden"
        :class="
          activeCardFilter === 'ON_TIME'
            ? 'border-emerald-500 ring-2 ring-emerald-500/40 shadow-lg bg-emerald-50/40'
            : 'border-slate-200 shadow-[0_4px_20px_rgba(0,0,0,0.03)] hover:shadow-md'
        "
        title="Klik untuk memfilter presensi tepat waktu"
      >
        <div class="flex items-start justify-between">
          <p class="text-[11px] font-bold text-slate-500 uppercase tracking-widest leading-tight">
            Tepat Waktu
          </p>
          <div
            class="p-2 rounded-xl transition-transform group-hover:scale-110"
            :class="activeCardFilter === 'ON_TIME' ? 'bg-emerald-500 text-white' : 'bg-emerald-50 text-emerald-500'"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="3"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </div>
        <div>
          <h3 class="text-2xl font-black text-slate-900 mt-3">{{ attendanceStats.onTime }}</h3>
          <div class="flex items-center justify-between mt-1">
            <p class="text-[10px] font-bold text-emerald-500 mt-1 uppercase tracking-tighter">
              On Schedule
            </p>
            <span
              v-if="activeCardFilter === 'ON_TIME'"
              class="text-[9px] font-black text-emerald-700 bg-emerald-100 px-1.5 py-0.5 rounded-md"
            >
              Aktif
            </span>
          </div>
        </div>
      </div>

      <!-- Card 3: Terlambat -->
      <div
        @click="toggleCardFilter('LATE')"
        class="group bg-white p-6 rounded-2xl border transition-all duration-300 cursor-pointer select-none hover:scale-[1.02] active:scale-[0.98] relative overflow-hidden"
        :class="
          activeCardFilter === 'LATE'
            ? 'border-orange-500 ring-2 ring-orange-500/40 shadow-lg bg-orange-50/40'
            : 'border-slate-200 shadow-[0_4px_20px_rgba(0,0,0,0.03)] hover:shadow-md'
        "
        title="Klik untuk memfilter presensi terlambat"
      >
        <div class="flex items-start justify-between">
          <p class="text-[11px] font-bold text-slate-500 uppercase tracking-widest leading-tight">
            Terlambat
          </p>
          <div
            class="p-2 rounded-xl transition-transform group-hover:scale-110"
            :class="activeCardFilter === 'LATE' ? 'bg-orange-500 text-white' : 'bg-orange-50 text-orange-500'"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="3"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3" />
            </svg>
          </div>
        </div>
        <div>
          <h3 class="text-2xl font-black text-slate-900 mt-3">{{ attendanceStats.late }}</h3>
          <div class="flex items-center justify-between mt-1">
            <p class="text-[10px] font-bold text-orange-500 mt-1 uppercase tracking-tighter">
              Delayed
            </p>
            <span
              v-if="activeCardFilter === 'LATE'"
              class="text-[9px] font-black text-orange-700 bg-orange-100 px-1.5 py-0.5 rounded-md"
            >
              Aktif
            </span>
          </div>
        </div>
      </div>

      <!-- Card 4: Absen/Cuti -->
      <div
        @click="toggleCardFilter('ABSENT_LEAVE')"
        class="group bg-white p-6 rounded-2xl border transition-all duration-300 cursor-pointer select-none hover:scale-[1.02] active:scale-[0.98] relative overflow-hidden"
        :class="
          activeCardFilter === 'ABSENT_LEAVE'
            ? 'border-red-500 ring-2 ring-red-500/40 shadow-lg bg-red-50/40'
            : 'border-slate-200 shadow-[0_4px_20px_rgba(0,0,0,0.03)] hover:shadow-md'
        "
        title="Klik untuk memfilter presensi absen, sakit, cuti, atau izin"
      >
        <div class="flex items-start justify-between">
          <p class="text-[11px] font-bold text-slate-500 uppercase tracking-widest leading-tight">
            Absen/Cuti
          </p>
          <div
            class="p-2 rounded-xl transition-transform group-hover:scale-110"
            :class="activeCardFilter === 'ABSENT_LEAVE' ? 'bg-red-500 text-white' : 'bg-red-50 text-red-500'"
          >
            <Plus class="w-4 h-4 rotate-45" />
          </div>
        </div>
        <div>
          <h3 class="text-2xl font-black text-slate-900 mt-3">
            {{ attendanceStats.absent + attendanceStats.permitted }}
          </h3>
          <div class="flex items-center justify-between mt-1">
            <p class="text-[10px] font-bold text-red-500 uppercase tracking-tighter">Off Work</p>
            <span
              v-if="activeCardFilter === 'ABSENT_LEAVE'"
              class="text-[9px] font-black text-red-700 bg-red-100 px-1.5 py-0.5 rounded-md"
            >
              Aktif
            </span>
          </div>
        </div>
      </div>

      <!-- Card 5: Berangkat Terawal -->
      <div
        @click="toggleCardFilter('EARLIEST')"
        class="group bg-white p-6 rounded-2xl border transition-all duration-300 cursor-pointer select-none hover:scale-[1.02] active:scale-[0.98] relative overflow-hidden"
        :class="
          activeCardFilter === 'EARLIEST'
            ? 'border-indigo-500 ring-2 ring-indigo-500/40 shadow-lg bg-indigo-50/40'
            : 'border-slate-200 shadow-[0_4px_20px_rgba(0,0,0,0.03)] hover:shadow-md'
        "
        title="Klik untuk menampilkan pekerja berangkat terawal hari ini"
      >
        <div class="flex items-start justify-between">
          <p class="text-[11px] font-bold text-slate-500 uppercase tracking-widest leading-tight">
            Berangkat Terawal
          </p>
          <div
            class="p-2 rounded-xl transition-transform group-hover:scale-110"
            :class="activeCardFilter === 'EARLIEST' ? 'bg-indigo-600 text-white' : 'bg-indigo-50 text-indigo-600'"
          >
            <CalendarClock class="w-4 h-4" />
          </div>
        </div>
        <div v-if="attendanceStats.earliestToday">
          <h3 class="text-[16px] font-black text-slate-900 mt-3 truncate">
            {{ attendanceStats.earliestToday.name }}
          </h3>
          <div class="flex items-center justify-between mt-1">
            <p class="text-[14px] font-black text-indigo-600 uppercase">
              {{ attendanceStats.earliestToday.timeStr }}
            </p>
            <span
              v-if="activeCardFilter === 'EARLIEST'"
              class="text-[9px] font-black text-indigo-700 bg-indigo-100 px-1.5 py-0.5 rounded-md"
            >
              Aktif
            </span>
          </div>
        </div>
        <div v-else>
          <h3 class="text-xl font-black text-slate-400 mt-3">--</h3>
          <p class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-tighter">
            Belum ada data
          </p>
        </div>
      </div>
    </div>

    <!-- DataTable Instance -->
    <div
      class="bg-white rounded-[2.5rem] border border-slate-200 shadow-2xl shadow-slate-100/50 overflow-hidden"
    >
      <DataTable :data="filteredTableData" :columns="columns" :isLoading="isLoading">
        <template #headerTitle>
          <div class="flex items-center gap-3 ml-2 flex-wrap">
            <h2 class="text-[16px] font-black text-slate-900 uppercase tracking-tighter">
              Catatan Harian Log Presensi
            </h2>

            <!-- Active Card Filter Pill -->
            <div
              v-if="activeCardFilter !== 'ALL'"
              class="flex items-center gap-2 bg-slate-100 px-3 py-1 rounded-full border border-slate-200 shadow-xs"
            >
              <span class="text-[11px] font-medium text-slate-500">Filter Card:</span>
              <span class="text-[11px] font-black text-slate-900">{{ activeFilterLabel }}</span>
              <span class="text-[11px] font-black text-primary">({{ filteredTableData.length }} Data)</span>
              <button
                type="button"
                @click="activeCardFilter = 'ALL'"
                class="ml-1 text-slate-400 hover:text-slate-700 text-xs font-bold px-1 hover:bg-slate-200 rounded-full transition-colors cursor-pointer"
                title="Reset filter"
              >
                ✕
              </button>
            </div>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Photo Preview Dialog (Mobile Watermark Inspector) -->
    <Dialog v-model:open="isPreviewOpen">
      <DialogContent class="sm:max-w-2xl p-0 bg-white rounded-3xl overflow-hidden shadow-2xl border border-slate-100 max-h-[92vh] flex flex-col">
        <!-- Header -->
        <div class="px-6 py-4 bg-slate-900 text-white flex items-center justify-between shrink-0">
          <div class="flex items-center gap-3">
            <div
              class="px-2.5 py-1 rounded-full text-[11px] font-black uppercase tracking-wider flex items-center gap-1.5"
              :class="
                previewData.type === 'IN'
                  ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                  : 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30'
              "
            >
              <LogIn v-if="previewData.type === 'IN'" class="w-3.5 h-3.5" />
              <LogOut v-else class="w-3.5 h-3.5" />
              <span>{{ previewData.type === 'IN' ? 'Foto Absen Masuk' : 'Foto Absen Pulang' }}</span>
            </div>
            <div>
              <h3 class="text-sm font-bold text-white leading-tight">
                {{ previewData.name }}
              </h3>
              <p class="text-[11px] text-slate-400 font-medium">
                {{ previewData.role }} &bull; {{ previewData.branch }}
              </p>
            </div>
          </div>

          <Button
            variant="ghost"
            size="icon"
            @click="isPreviewOpen = false"
            class="text-slate-400 hover:text-white hover:bg-slate-800 rounded-full h-8 w-8"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </Button>
        </div>

        <!-- Body: Photo & Watermark Info -->
        <div class="p-6 space-y-4 overflow-y-auto custom-scrollbar flex-1 bg-slate-50/50">
          <!-- Photo Container with Mobile-Parity Watermark Overlay -->
          <div class="relative w-full max-w-md mx-auto aspect-square rounded-2xl overflow-hidden shadow-md bg-black border border-slate-200 group">
            <img
              :src="previewData.url"
              :alt="previewData.name"
              class="w-full h-full object-cover select-none"
              crossorigin="anonymous"
            />

            <!-- Toggle Overlay Button (Floating badge on top-right) -->
            <button
              type="button"
              @click="showWatermarkOverlay = !showWatermarkOverlay"
              class="absolute top-2.5 right-2.5 z-10 px-2.5 py-1.5 rounded-xl bg-black/60 hover:bg-black/80 backdrop-blur-md text-white text-[11px] font-semibold flex items-center gap-1.5 shadow-sm transition-all border border-white/15 cursor-pointer"
              :title="showWatermarkOverlay ? 'Sembunyikan Overlay Watermark' : 'Tampilkan Overlay Watermark'"
            >
              <EyeOff v-if="showWatermarkOverlay" class="w-3.5 h-3.5 text-amber-300" />
              <Eye v-else class="w-3.5 h-3.5 text-slate-300" />
              <span class="hidden sm:inline">{{ showWatermarkOverlay ? 'Sembunyikan Watermark' : 'Tampilkan Watermark' }}</span>
            </button>

            <!-- Dynamic Watermark Bar (Mobile Parity) -->
            <transition
              enter-active-class="transition duration-200 ease-out"
              enter-from-class="opacity-0 translate-y-2"
              enter-to-class="opacity-100 translate-y-0"
              leave-active-class="transition duration-150 ease-in"
              leave-from-class="opacity-100 translate-y-0"
              leave-to-class="opacity-0 translate-y-2"
            >
              <div
                v-if="showWatermarkOverlay"
                class="absolute bottom-0 inset-x-0 bg-black/70 backdrop-blur-xs p-3 md:p-3.5 text-white space-y-1 text-left border-t border-white/10"
              >
                <!-- Time Row -->
                <div class="flex items-center gap-2 text-[11px] md:text-xs font-bold text-white tracking-wide">
                  <Clock class="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                  <span class="truncate">{{ previewData.timeFormatted }}</span>
                </div>

                <!-- Branch Row -->
                <div class="flex items-center gap-2 text-[11px] md:text-xs font-semibold text-slate-200">
                  <Building2 class="w-3.5 h-3.5 text-sky-400 shrink-0" />
                  <span class="truncate">{{ previewData.branch }}</span>
                </div>

                <!-- Location / GPS Coordinates Row -->
                <div class="flex items-start gap-2 text-[10px] md:text-[11px] text-slate-300 leading-tight">
                  <MapPin class="w-3.5 h-3.5 text-rose-400 shrink-0 mt-0.5" />
                  <span class="line-clamp-2">
                    {{ previewData.lat && previewData.lng ? `GPS: ${previewData.lat}, ${previewData.lng}` : 'Lokasi Presensi Pekerja Harian' }}
                  </span>
                </div>
              </div>
            </transition>
          </div>

          <!-- Watermark Metadata Panel -->
          <div class="bg-white rounded-2xl border border-slate-200/80 p-4 space-y-2.5 shadow-xs">
            <div class="text-[11px] font-black uppercase tracking-widest text-slate-400 flex items-center gap-1.5 pb-1 border-b border-slate-100">
              <Camera class="w-3.5 h-3.5 text-primary" />
              <span>Metadata & Watermark Log</span>
            </div>

            <!-- Time Row -->
            <div class="flex items-start gap-2.5 text-xs text-slate-700">
              <Clock class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5" />
              <div>
                <p class="font-bold text-slate-900">{{ previewData.timeFormatted }}</p>
                <p class="text-[11px] text-slate-400">Waktu tercatat pada server</p>
              </div>
            </div>

            <!-- Branch Row -->
            <div class="flex items-start gap-2.5 text-xs text-slate-700">
              <Building2 class="w-4 h-4 text-indigo-600 shrink-0 mt-0.5" />
              <div>
                <p class="font-bold text-slate-900">{{ previewData.branch }}</p>
                <p class="text-[11px] text-slate-400">Unit / Cabang Penugasan</p>
              </div>
            </div>

            <!-- Location / GPS Coordinates Row -->
            <div v-if="previewData.lat && previewData.lng" class="flex items-start gap-2.5 text-xs text-slate-700">
              <MapPin class="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
              <div class="flex-1">
                <div class="flex items-center justify-between gap-2">
                  <p class="font-bold text-slate-900">
                    Lat: {{ previewData.lat }}, Long: {{ previewData.lng }}
                  </p>
                  <a
                    :href="`https://www.google.com/maps?q=${previewData.lat},${previewData.lng}`"
                    target="_blank"
                    class="text-[11px] font-bold text-primary hover:underline flex items-center gap-1 shrink-0"
                  >
                    Buka Maps <ExternalLink class="w-3 h-3" />
                  </a>
                </div>
              </div>
            </div>

            <!-- Notes if any -->
            <div v-if="previewData.notes && previewData.notes !== '-'" class="pt-2 border-t border-slate-100 text-xs text-slate-600 flex items-start gap-2">
              <span class="font-bold text-slate-700 shrink-0">Catatan:</span>
              <span class="italic text-slate-600">{{ previewData.notes }}</span>
            </div>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="px-6 py-3.5 bg-white border-t border-slate-100 flex items-center justify-between shrink-0">
          <Button
            variant="ghost"
            size="sm"
            :disabled="isDownloading"
            @click="handleDownloadWatermarkedPhoto"
            class="text-xs font-bold text-slate-600 hover:text-slate-900 flex items-center gap-1.5 px-3 py-2 rounded-xl hover:bg-slate-100 transition-colors"
          >
            <Loader2 v-if="isDownloading" class="w-4 h-4 animate-spin text-primary" />
            <Download v-else class="w-4 h-4 text-emerald-600" />
            <span>{{ isDownloading ? 'Memproses Cap Watermark...' : 'Unduh Foto (Ber-Watermark)' }}</span>
          </Button>

          <Button
            @click="isPreviewOpen = false"
            class="bg-slate-900 hover:bg-slate-800 text-white rounded-xl text-xs font-bold px-5 h-9"
          >
            Tutup
          </Button>
        </div>
      </DialogContent>
    </Dialog>

    <!-- UI Dialog (Manual Overrides) -->
    <Dialog v-model:open="isModalOpen">
      <DialogContent
        class="sm:max-w-xl rounded-[2rem] md:rounded-[3rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <div class="bg-slate-900 p-8 md:p-10 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl md:text-2xl font-black">{{
              isEditMode ? 'Edit Log Kehadiran' : 'Input Kehadiran Manual'
            }}</DialogTitle>
            <DialogDescription
              class="text-slate-400 font-bold mt-2 uppercase text-[10px] md:text-[11px] tracking-widest border-l-4 border-primary pl-4"
            >
              Bypass Face-ID untuk memperbaiki data presensi karyawan secara administratif.
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-6 md:p-10 bg-white space-y-6 overflow-y-auto custom-scrollbar">
          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Pilih Karyawan</label>
            <Select v-model="currentForm.employee_id" :disabled="isEditMode">
              <SelectTrigger>
                <SelectValue placeholder="Pilih Karyawan..." />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="emp in employeeOptions" :key="emp.id" :value="emp.id">
                  {{ emp.first_name }} ({{ emp.user?.email }})
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Tanggal</label>
            <Input v-model="currentForm.date" type="date" />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <label class="text-[13px] font-medium text-gray-700 uppercase tracking-wider"
                >Jam Masuk (HH.mm)</label
              >
              <Input v-model="currentForm.clock_in_time" type="text" placeholder="contoh: 07.30" />
            </div>
            <div class="grid gap-2">
              <label class="text-[13px] font-medium text-gray-700 uppercase tracking-wider"
                >Jam Keluar (HH.mm)</label
              >
              <Input v-model="currentForm.clock_out_time" type="text" placeholder="contoh: 16.30" />
            </div>
          </div>

          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Status Penyelesaian</label>
            <Select v-model="currentForm.status">
              <SelectTrigger>
                <SelectValue placeholder="Status Kehadiran" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ON_TIME">Tepat Waktu (ON_TIME)</SelectItem>
                <SelectItem value="LATE">Terlambat (LATE)</SelectItem>
                <SelectItem value="ABSENT">Mangkir (ABSENT)</SelectItem>
                <SelectItem value="LEAVE">Cuti/Izin (LEAVE)</SelectItem>
                <SelectItem value="SICK">Sakit (SICK)</SelectItem>
                <SelectItem value="PERMISSION">Izin Tanpa Dibayar (PERMISSION)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="grid gap-2">
            <label class="text-[13px] font-medium text-gray-700">Catatan Khusus</label>
            <Input
              v-model="currentForm.notes"
              placeholder="e.g. Lupa absen pulang, dikoreksi manual"
            />
          </div>
        </div>

        <DialogFooter class="p-8 md:p-10 pt-0 bg-white grid grid-cols-2 gap-4 shrink-0">
          <Button
            variant="ghost"
            @click="closeAddModal"
            :disabled="isSubmitting"
            class="rounded-2xl h-12 md:h-14 font-black text-slate-400 border border-slate-100 hover:bg-slate-50"
            >BATAL</Button
          >
          <Button
            @click="saveManualData"
            :disabled="isSubmitting"
            class="bg-primary hover:bg-primary/90 text-white rounded-2xl h-12 md:h-14 font-black shadow-xl shadow-primary/20 transform active:scale-95 transition-all"
          >
            {{ isSubmitting ? 'MENYIMPAN...' : 'SIMPAN OVERRIDE' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
