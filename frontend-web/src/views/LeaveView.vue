<script setup lang="ts">
import { ref, onMounted, h, computed, toRaw, watch } from 'vue'
import { toast } from 'vue-sonner'
import { Plus, Pencil, Trash2, CheckCircle, XCircle, FileText, Eye, ZoomIn, ZoomOut, RotateCw, RotateCcw, RefreshCw, ExternalLink } from 'lucide-vue-next'
import axios from '@/api/axios'
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

const masterData = useMasterDataStore()

watch(() => masterData.selectedBranchId, () => {
  currentPage.value = 1
  fetchLeaves()
  fetchStats()
})

const isLoading = ref(true)
const isSubmitting = ref(false)
const displayData = ref<any[]>([])
const employees = ref<any[]>([])

const currentPage = ref(1)
const pageSize = ref(10)
const totalItems = ref(0)
const searchQuery = ref('')
const stats = ref({
  ALL: 0,
  PENDING: 0,
  APPROVED: 0,
  REJECTED: 0,
})

// --- Filter ---
const activeFilter = ref('ALL')
const statusFilters = ['ALL', 'PENDING', 'APPROVED', 'REJECTED']

const filteredData = computed(() => {
  return displayData.value
})

// --- CRUD Modal ---
const isModalOpen = ref(false)
const isEditMode = ref(false)
const form = ref({
  id: '',
  employee_id: '',
  leave_type_id: '',
  start_date: '',
  end_date: '',
  reason: '',
  attachment_url: '',
  attachment_file: null as File | null,
  status: 'PENDING',
  reject_reason: '',
})

// --- Reject Modal ---
const rejectModalOpen = ref(false)
const selectedLeaveId = ref<string>('')
const rejectReason = ref<string>('')

// --- View Attachment Modal ---
const attachmentModalOpen = ref(false)
const selectedAttachmentURL = ref<string>('')
const zoomLevel = ref(1)
const rotation = ref(0)

const zoomIn = () => {
  zoomLevel.value = Math.min(Number((zoomLevel.value + 0.25).toFixed(2)), 3.5)
}
const zoomOut = () => {
  zoomLevel.value = Math.max(Number((zoomLevel.value - 0.25).toFixed(2)), 0.5)
}
const resetView = () => {
  zoomLevel.value = 1
  rotation.value = 0
}
const rotateCw = () => {
  rotation.value = (rotation.value + 90) % 360
}
const rotateCcw = () => {
  rotation.value = (rotation.value - 90 + 360) % 360
}

const isPdfAttachment = computed(() => {
  if (!selectedAttachmentURL.value) return false
  const cleanUrl = selectedAttachmentURL.value.split('?')[0]?.toLowerCase() || ''
  return cleanUrl.endsWith('.pdf') || selectedAttachmentURL.value.toLowerCase().includes('pdf')
})

const isImageAttachment = computed(() => {
  if (!selectedAttachmentURL.value) return false
  if (isPdfAttachment.value) return false
  const cleanUrl = selectedAttachmentURL.value.split('?')[0]?.toLowerCase() || ''
  return (
    /\.(jpg|jpeg|png|gif|webp|bmp|svg|jfif|heic)$/i.test(cleanUrl) ||
    selectedAttachmentURL.value.includes('/attachments/') ||
    selectedAttachmentURL.value.includes('/files/') ||
    selectedAttachmentURL.value.startsWith('data:image') ||
    !isPdfAttachment.value
  )
})

const openAddModal = () => {
  isEditMode.value = false
  form.value = {
    id: '',
    employee_id: '',
    leave_type_id: '',
    start_date: '',
    end_date: '',
    reason: '',
    attachment_url: '',
    attachment_file: null,
    status: 'PENDING',
    reject_reason: '',
  }
  isModalOpen.value = true
}

const openEditModal = (row: any) => {
  isEditMode.value = true
  form.value = {
    id: String(row.id || ''),
    employee_id: String(row.employee_id || ''),
    leave_type_id: String(row.leave_type_id || ''),
    start_date: row.start_date ? new Date(row.start_date).toISOString().slice(0, 10) : '',
    end_date: row.end_date ? new Date(row.end_date).toISOString().slice(0, 10) : '',
    reason: (row.reason || '') as string,
    attachment_url: (row.attachment_url || '') as string,
    attachment_file: null,
    status: (row.status || 'PENDING') as string,
    reject_reason: (row.reject_reason || '') as string,
  }
  isModalOpen.value = true
}

const onFileChange = (e: any) => {
  const file = e.target.files[0]
  if (file) {
    form.value.attachment_file = file
  }
}

const closeModal = () => {
  isModalOpen.value = false
}

const saveLeave = async () => {
  if (
    !form.value.employee_id ||
    !form.value.leave_type_id ||
    !form.value.start_date ||
    !form.value.end_date
  ) {
    toast.error('Karyawan, jenis izin, dan tanggal wajib diisi.')
    return
  }
  isSubmitting.value = true
  try {
    const formData = new FormData()
    formData.append('employee_id', form.value.employee_id)
    formData.append('leave_type_id', form.value.leave_type_id)
    formData.append('start_date', form.value.start_date)
    formData.append('end_date', form.value.end_date)
    formData.append('reason', form.value.reason)
    formData.append('status', form.value.status)
    if (form.value.attachment_file) {
      formData.append('attachment', toRaw(form.value.attachment_file))
    }

    const config = {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }

    if (isEditMode.value) {
      await axios.put(`/time-off/manage/${form.value.id}`, formData, config)
      toast.success('Pengajuan berhasil diperbarui!')
    } else {
      await axios.post('/time-off/manage', formData, config)
      toast.success('Pengajuan berhasil ditambahkan!')
    }
    closeModal()
    fetchLeaves()
    fetchStats()
  } catch (e: any) {
    toast.error('Gagal menyimpan: ' + (e.response?.data?.message || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const deleteLeave = async (id: string) => {
  if (!confirm('Hapus pengajuan ini? Tindakan ini tidak dapat dibatalkan.')) return
  try {
    await axios.delete(`/time-off/manage/${id}`)
    toast.success('Pengajuan berhasil dihapus.')
    fetchLeaves()
    fetchStats()
  } catch (e: any) {
    toast.error('Gagal menghapus: ' + (e.response?.data?.message || e.message))
  }
}

const approveLeave = async (id: string) => {
  if (!confirm('Setujui pengajuan cuti/izin ini?')) return
  try {
    await axios.put(`/time-off/manage/${id}/approve`, { status: 'APPROVED' })
    toast.success('Pengajuan berhasil disetujui.')
    fetchLeaves()
    fetchStats()
  } catch (e: any) {
    toast.error('Gagal menyetujui: ' + (e.response?.data?.message || e.message))
  }
}

const openRejectModal = (id: string) => {
  selectedLeaveId.value = id
  rejectReason.value = ''
  rejectModalOpen.value = true
}

const confirmReject = async () => {
  if (!rejectReason.value.trim()) {
    toast.error('Alasan penolakan wajib diisi.')
    return
  }
  isSubmitting.value = true
  try {
    await axios.put(`/time-off/manage/${selectedLeaveId.value}/approve`, {
      status: 'REJECTED',
      reject_reason: rejectReason.value,
    })
    toast.success('Pengajuan berhasil ditolak.')
    rejectModalOpen.value = false
    fetchLeaves()
    fetchStats()
  } catch (e: any) {
    toast.error('Gagal menolak: ' + (e.response?.data?.message || e.message))
  } finally {
    isSubmitting.value = false
  }
}

const openAttachmentModal = (url: string) => {
  const baseUrl = import.meta.env.VITE_API_BASE_URL.replace('/api', '')
  selectedAttachmentURL.value = url.startsWith('http') ? url : `${baseUrl}${url}`
  zoomLevel.value = 1
  rotation.value = 0
  attachmentModalOpen.value = true
}

const fetchLeaves = async () => {
  isLoading.value = true
  try {
    const statusParam = activeFilter.value !== 'ALL' ? activeFilter.value : ''
    const params: Record<string, string> = {
      page: String(currentPage.value),
      limit: String(pageSize.value),
      status: statusParam,
      search: searchQuery.value,
    }
    if (masterData.selectedBranchId && masterData.selectedBranchId !== 'ALL_BRANCHES') {
      params.branch_id = masterData.selectedBranchId
    }
    const queryStr = new URLSearchParams(params).toString()
    const res = await axios.get(`/time-off/manage?${queryStr}`)
    if (res.data?.data) {
      displayData.value = res.data.data
      totalItems.value = res.data.meta?.total || 0
    }
  } catch (error) {
    console.error('Failed fetching leave data', error)
  } finally {
    isLoading.value = false
  }
}

const fetchStats = async () => {
  try {
    let branchParam = ''
    if (masterData.selectedBranchId && masterData.selectedBranchId !== 'ALL_BRANCHES') {
      branchParam = `&branch_id=${masterData.selectedBranchId}`
    }
    const [allRes, pendingRes, approvedRes, rejectedRes] = await Promise.all([
      axios.get(`/time-off/manage?limit=1${branchParam}`),
      axios.get(`/time-off/manage?status=PENDING&limit=1${branchParam}`),
      axios.get(`/time-off/manage?status=APPROVED&limit=1${branchParam}`),
      axios.get(`/time-off/manage?status=REJECTED&limit=1${branchParam}`),
    ])
    stats.value.ALL = allRes.data.meta?.total || 0
    stats.value.PENDING = pendingRes.data.meta?.total || 0
    stats.value.APPROVED = approvedRes.data.meta?.total || 0
    stats.value.REJECTED = rejectedRes.data.meta?.total || 0
  } catch (e) {
    console.error('Failed to fetch leave stats', e)
  }
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchLeaves()
}

const handleSearch = (query: string) => {
  searchQuery.value = query
  currentPage.value = 1
  fetchLeaves()
}

watch(activeFilter, () => {
  currentPage.value = 1
  fetchLeaves()
})

const fetchEmployees = async () => {
  try {
    const res = await axios.get('/employees?limit=all')
    employees.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

const fmt = (d: string) =>
  d
    ? new Date(d).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
    : '-'

const getInitials = (name: string) =>
  name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

const columns = [
  {
    accessorFn: (row: any) => {
      const emp = row.employee
      return emp ? `${emp.first_name || ''} ${emp.last_name || ''}`.trim() : '-'
    },
    id: 'employee',
    header: 'KARYAWAN',
    cell: (info: any) => {
      const row = info.row.original
      const emp = row.employee
      const branchName = emp?.branch?.name || '-'
      return h('div', { class: 'flex flex-col' }, [
        h('div', { class: 'font-bold text-gray-900 leading-tight text-[13px]' }, info.getValue()),
        h('div', { class: 'text-gray-500 text-[11px]' }, branchName),
      ])
    },
  },
  {
    accessorFn: (row: any) => row.leave_type?.name || '-',
    id: 'leaveType',
    header: 'JENIS CUTI/IZIN',
    cell: ({ row }: any) => {
      const name = row.original.leave_type?.name || '-'
      const isSick = name.toLowerCase().includes('sakit') || name.toLowerCase().includes('sick')
      return h('div', { class: 'flex flex-col gap-0.5' }, [
        h('span', { class: 'font-semibold text-gray-700 text-[13px]' }, name),
        isSick
          ? h(
              'span',
              { class: 'text-[11px] text-orange-600 font-medium flex items-center gap-1' },
              '★ Perlu Dok. Dokter',
            )
          : null,
      ])
    },
  },
  {
    accessorKey: 'start_date',
    id: 'period',
    header: 'PERIODE',
    cell: ({ row }: any) => {
      const r = row.original
      const start = fmt(r.start_date)
      const end = fmt(r.end_date)
      const days =
        Math.ceil(
          (new Date(r.end_date).getTime() - new Date(r.start_date).getTime()) / (1000 * 3600 * 24),
        ) + 1
      return h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'text-gray-700 text-[13px]' }, `${start} - ${end}`),
        h('span', { class: 'text-gray-400 text-[11px]' }, `${days} hari`),
      ])
    },
  },
  {
    accessorKey: 'reason',
    header: 'ALASAN',
    cell: (info: any) =>
      h(
        'span',
        {
          class: 'text-gray-500 text-[13px] max-w-[180px] truncate inline-block',
          title: info.getValue(),
        },
        info.getValue() || '-',
      ),
  },
  {
    id: 'attachment',
    header: 'LAMPIRAN (DOKTER)',
    cell: ({ row }: any) => {
      const url = row.original.attachment_url
      if (!url) return h('span', { class: 'text-gray-400 text-[12px]' }, '—')
      return h(
        'button',
        {
          class: 'flex items-center gap-1 text-primary text-[12px] font-semibold hover:underline',
          onClick: () => openAttachmentModal(url),
        },
        [h(FileText, { class: 'w-3.5 h-3.5' }), 'Lihat Lampiran'],
      )
    },
  },
  {
    accessorKey: 'created_at',
    id: 'created_at',
    header: 'DIBUAT PADA',
    cell: ({ row }: any) => {
      const d = row.original.created_at
      if (!d) return h('span', { class: 'text-gray-400 text-[13px]' }, '-')
      const date = new Date(d)
      const dateStr = date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
      const timeStr = date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
      return h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'text-gray-700 text-[13px] font-medium' }, dateStr),
        h('span', { class: 'text-gray-500 text-[11px]' }, timeStr),
      ])
    },
  },
  {
    accessorKey: 'status',
    header: 'STATUS',
    cell: ({ row, getValue }: any) => {
      const val = getValue() as string
      let cls = 'border-yellow-200 text-yellow-700 bg-yellow-50'
      if (val === 'APPROVED') cls = 'border-green-200 text-green-700 bg-green-50'
      else if (val === 'REJECTED') cls = 'border-red-200 text-red-600 bg-red-50'
      
      const approver = row.original.approver
      const approverName = approver ? `${approver.first_name || ''} ${approver.last_name || ''}`.trim() : ''
      const approverPosition = approver?.job_position?.title || ''
      const rejectReason = row.original.reject_reason || ''

      return h('div', { class: 'flex flex-col items-start gap-1 max-w-[220px]' }, [
        h('span', { class: `border ${cls} px-2.5 py-0.5 rounded-full text-[11px] font-bold inline-block` }, val),
        val === 'APPROVED' && approverName 
          ? h('div', { class: 'flex flex-col text-[10px] text-gray-500 mt-0.5 leading-tight' }, [
              h('span', { class: 'font-semibold text-gray-700' }, `✓ Disetujui: ${approverName}`),
              approverPosition ? h('span', { class: 'text-gray-400 text-[9px]' }, approverPosition) : null,
            ])
          : null,
        val === 'REJECTED' 
          ? h('div', { class: 'flex flex-col text-[10px] text-red-600 mt-0.5 leading-tight gap-0.5' }, [
              approverName ? h('span', { class: 'font-semibold text-red-700' }, `✕ Ditolak: ${approverName}`) : null,
              rejectReason 
                ? h('div', { class: 'bg-red-50 border border-red-100 rounded px-1.5 py-0.5 text-[10px] text-red-700 font-medium italic break-words w-full' }, `"${rejectReason}"`)
                : null,
            ])
          : null
      ])
    },
  },
  {
    id: 'actions',
    header: 'AKSI',
    cell: ({ row }: any) => {
      const r = row.original
      const status = r.status
      return h('div', { class: 'flex items-center gap-1' }, [
        h(
          'button',
          {
            class: 'p-1.5 rounded hover:bg-blue-50 text-blue-600 transition-colors',
            title: 'Edit',
            onClick: () => openEditModal(r),
          },
          h(Pencil, { class: 'w-3.5 h-3.5' }),
        ),

        status === 'PENDING'
          ? h(
              'button',
              {
                class: 'p-1.5 rounded hover:bg-green-50 text-green-600 transition-colors',
                title: 'Setujui',
                onClick: () => approveLeave(r.id),
              },
              h(CheckCircle, { class: 'w-3.5 h-3.5' }),
            )
          : null,

        status === 'PENDING'
          ? h(
              'button',
              {
                class: 'p-1.5 rounded hover:bg-red-50 text-red-600 transition-colors',
                title: 'Tolak',
                onClick: () => openRejectModal(r.id),
              },
              h(XCircle, { class: 'w-3.5 h-3.5' }),
            )
          : null,

        h(
          'button',
          {
            class: 'p-1.5 rounded hover:bg-rose-50 text-rose-600 transition-colors',
            title: 'Hapus',
            onClick: () => deleteLeave(r.id),
          },
          h(Trash2, { class: 'w-3.5 h-3.5' }),
        ),
      ])
    },
    enableSorting: false,
  },
]

onMounted(() => {
  fetchLeaves()
  fetchStats()
  fetchEmployees()
  masterData.fetchLeaveTypes()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-[22px] font-bold text-gray-900 leading-tight">Pengajuan Cuti & Izin</h1>
        <p class="text-[14px] text-gray-500 mt-1">
          Kelola, validasi, dan proses permohonan cuti dan izin karyawan.
        </p>
      </div>
      <Button @click="openAddModal" class="flex items-center gap-2 shrink-0">
        <Plus class="w-4 h-4" />
        Tambah Pengajuan
      </Button>
    </div>

    <!-- Status Filter Tabs -->
    <div class="flex gap-2 flex-wrap">
      <button
        v-for="f in statusFilters"
        :key="f"
        @click="activeFilter = f"
        :class="[
          'px-4 py-1.5 rounded-full text-[13px] font-semibold border transition-colors',
          activeFilter === f
            ? 'bg-primary text-white border-primary'
            : 'bg-white text-gray-600 border-gray-300 hover:border-primary/50',
        ]"
      >
        {{ f }}
      </button>
    </div>

    <!-- Stats Summary -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div
        v-for="stat in [
          { label: 'Total Pengajuan', color: 'text-gray-700', key: 'ALL' },
          { label: 'Menunggu', color: 'text-yellow-600', key: 'PENDING' },
          { label: 'Disetujui', color: 'text-green-600', key: 'APPROVED' },
          { label: 'Ditolak', color: 'text-red-500', key: 'REJECTED' },
        ]"
        :key="stat.label"
        class="bg-white rounded-xl border border-gray-200 px-4 py-3 shadow-sm cursor-pointer hover:shadow-md transition-shadow"
        @click="activeFilter = stat.key"
      >
        <p class="text-xs text-gray-500 font-medium">{{ stat.label }}</p>
        <p :class="['text-2xl font-bold mt-1', stat.color]">
          {{ stats[stat.key as keyof typeof stats] }}
        </p>
      </div>
    </div>

    <!-- Table -->
    <DataTable
      :data="filteredData"
      :columns="columns"
      :isLoading="isLoading"
      :serverSide="true"
      :totalItems="totalItems"
      :currentPage="currentPage"
      :pageSize="pageSize"
      @update:page="handlePageChange"
      @search="handleSearch"
    />

    <!-- ─────────── ADD/EDIT MODAL ─────────── -->
    <Dialog v-model:open="isModalOpen">
      <DialogContent
        class="sm:max-w-lg rounded-[2rem] md:rounded-[3rem] p-0 overflow-hidden border-none shadow-2xl max-h-[95vh] flex flex-col"
      >
        <div class="bg-slate-900 p-8 md:p-10 text-white relative shrink-0">
          <DialogHeader>
            <DialogTitle class="text-xl md:text-2xl font-black">{{
              isEditMode ? 'Edit Pengajuan' : 'Tambah Pengajuan Cuti/Izin'
            }}</DialogTitle>
            <DialogDescription
              class="text-slate-400 font-bold mt-2 uppercase text-[10px] md:text-[11px] tracking-widest border-l-4 border-primary pl-4"
            >
              Isi data pengajuan cuti atau izin karyawan. Lampirkan surat dokter untuk izin sakit.
            </DialogDescription>
          </DialogHeader>
        </div>

        <div class="p-6 md:p-10 bg-white space-y-6 overflow-y-auto custom-scrollbar flex-1">
          <!-- Karyawan -->
          <div class="grid gap-2">
            <label class="text-sm font-medium">Karyawan <span class="text-red-500">*</span></label>
            <Select v-model="form.employee_id">
              <SelectTrigger><SelectValue placeholder="Pilih Karyawan" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="e in employees" :key="e.id" :value="e.id">
                  {{ e.first_name }} {{ e.last_name || '' }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Jenis Cuti -->
          <div class="grid gap-2">
            <label class="text-sm font-medium"
              >Jenis Cuti / Izin <span class="text-red-500">*</span></label
            >
            <Select v-model="form.leave_type_id">
              <SelectTrigger><SelectValue placeholder="Pilih Jenis Izin" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="lt in masterData.leaveTypes" :key="lt.id" :value="lt.id">{{
                  lt.name
                }}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Tanggal -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="grid gap-2">
              <label class="text-sm font-medium"
                >Tanggal Mulai <span class="text-red-500">*</span></label
              >
              <Input type="date" v-model="form.start_date" />
            </div>
            <div class="grid gap-2">
              <label class="text-sm font-medium"
                >Tanggal Selesai <span class="text-red-500">*</span></label
              >
              <Input type="date" v-model="form.end_date" />
            </div>
          </div>

          <!-- Alasan -->
          <div class="grid gap-2">
            <label class="text-sm font-medium">Alasan / Keterangan</label>
            <textarea
              v-model="form.reason"
              class="min-h-[72px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              placeholder="Alasan cuti atau izin..."
            ></textarea>
          </div>

          <!-- Lampiran surat dokter -->
          <div class="grid gap-2">
            <label class="text-sm font-medium flex items-center gap-1.5">
              <FileText class="w-4 h-4 text-orange-500" />
              Lampiran Surat Dokter
            </label>
            <Input type="file" @change="onFileChange" accept=".jpg,.jpeg,.png,.pdf" />
            <div v-if="form.attachment_url" class="flex items-center gap-2 mt-1">
              <span class="text-[11px] text-gray-500">Berkas saat ini:</span>
              <button
                type="button"
                @click="openAttachmentModal(form.attachment_url)"
                class="text-[11px] text-primary font-semibold hover:underline flex items-center gap-1"
              >
                <Eye class="w-3 h-3" /> Lihat
              </button>
            </div>
            <p class="text-xs text-gray-500 italic">
              Wajib diisi untuk izin sakit agar dapat divalidasi oleh manajerial.
            </p>
          </div>

          <!-- Status (Only in Edit Mode) -->
          <div v-if="isEditMode" class="grid gap-2 border-t pt-4">
            <label class="text-sm font-medium"
              >Status Pengajuan <span class="text-red-500">*</span></label
            >
            <Select v-model="form.status">
              <SelectTrigger><SelectValue placeholder="Update Status" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="PENDING">PENDING</SelectItem>
                <SelectItem value="APPROVED">APPROVED</SelectItem>
                <SelectItem value="REJECTED">REJECTED</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Rejection Reason (If Rejected) -->
          <div v-if="isEditMode && (form.status === 'REJECTED' || form.reject_reason)" class="grid gap-2 bg-red-50/70 border border-red-200 rounded-xl p-3.5">
            <label class="text-xs font-bold text-red-700 uppercase tracking-wider flex items-center gap-1.5">
              <span class="w-2 h-2 rounded-full bg-red-500"></span>
              Alasan Penolakan HR
            </label>
            <p class="text-sm text-red-900 font-medium leading-relaxed">
              {{ form.reject_reason || 'Belum ada catatan alasan penolakan.' }}
            </p>
          </div>
        </div>

        <DialogFooter class="p-8 md:p-10 pt-0 bg-white grid grid-cols-2 gap-4 shrink-0">
          <Button
            variant="ghost"
            @click="closeModal"
            :disabled="isSubmitting"
            class="rounded-2xl h-12 md:h-14 font-black text-slate-400 border border-slate-100 hover:bg-slate-50 transition-all"
            >BATAL</Button
          >
          <Button
            @click="saveLeave"
            :disabled="isSubmitting"
            class="bg-primary hover:bg-primary/90 text-white rounded-2xl h-12 md:h-14 font-black shadow-xl shadow-primary/20 transform active:scale-95 transition-all"
          >
            {{ isSubmitting ? 'MENYIMPAN...' : 'SIMPAN PENGAJUAN' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- ─────────── REJECT MODAL ─────────── -->
    <Dialog v-model:open="rejectModalOpen">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle class="text-xl text-red-600">Tolak Pengajuan</DialogTitle>
          <DialogDescription>
            Cuti yang ditolak akan otomatis membatalkan pemotongan saldo cuti karyawan.
          </DialogDescription>
        </DialogHeader>
        <div class="grid gap-4 py-4">
          <div class="grid gap-2">
            <label class="text-sm font-medium"
              >Alasan Penolakan <span class="text-red-500">*</span></label
            >
            <textarea
              v-model="rejectReason"
              class="min-h-[72px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              placeholder="Masukkan alasan penolakan..."
            ></textarea>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="rejectModalOpen = false" :disabled="isSubmitting"
            >Batal</Button
          >
          <Button
            variant="destructive"
            @click="confirmReject"
            :disabled="isSubmitting || !rejectReason.trim()"
          >
            {{ isSubmitting ? 'Memproses...' : 'Tolak Pengajuan' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- ─────────── ATTACHMENT VIEWER MODAL ─────────── -->
    <Dialog v-model:open="attachmentModalOpen">
      <DialogContent class="sm:max-w-4xl max-h-[92vh] flex flex-col p-6 rounded-[2rem] overflow-hidden bg-white shadow-2xl border-slate-100">
        <DialogHeader class="pb-3 border-b border-slate-100">
          <div class="flex items-center justify-between">
            <div class="space-y-1">
              <DialogTitle class="flex items-center gap-2 text-lg font-black text-slate-800">
                <FileText class="w-5 h-5 text-indigo-600" />
                Dokumen Surat Dokter / Bukti Lampiran
              </DialogTitle>
              <DialogDescription class="text-xs font-medium text-slate-400">
                Validasi keabsahan dokumen dan bukti yang dilampirkan karyawan.
              </DialogDescription>
            </div>
            <a
              v-if="selectedAttachmentURL"
              :href="selectedAttachmentURL"
              target="_blank"
              class="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl text-xs font-bold transition-all"
              title="Buka file asli di tab baru"
            >
              <ExternalLink class="w-3.5 h-3.5" />
              Tab Baru
            </a>
          </div>
        </DialogHeader>

        <div class="flex-1 my-2 min-h-[360px] h-[58vh] max-h-[600px] relative bg-slate-950 rounded-2xl overflow-auto flex items-center justify-center border border-slate-800/80 p-3 select-none">
          <!-- PDF Viewer -->
          <iframe
            v-if="isPdfAttachment"
            :src="selectedAttachmentURL"
            class="w-full h-full rounded-xl bg-white border-0"
            frameborder="0"
          ></iframe>

          <!-- Image Viewer -->
          <div
            v-else-if="isImageAttachment"
            class="w-full h-full flex items-center justify-center overflow-hidden"
          >
            <img
              :src="selectedAttachmentURL"
              :style="{ transform: `scale(${zoomLevel}) rotate(${rotation}deg)` }"
              class="max-w-full max-h-full object-contain transition-transform duration-200 rounded-lg shadow-lg"
              alt="Lampiran Cuti/Izin"
            />
          </div>

          <!-- Fallback Format Unknown -->
          <div v-else class="text-center p-8">
            <FileText class="w-16 h-16 text-slate-500 mx-auto mb-4" />
            <p class="text-slate-300 font-semibold text-sm">Format tidak dapat dipratinjau langsung di dalam browser.</p>
            <a
              :href="selectedAttachmentURL"
              target="_blank"
              class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold transition-all shadow-md"
            >
              <ExternalLink class="w-4 h-4" />
              Buka File di Tab Baru
            </a>
          </div>

          <!-- Floating Image Controls Overlay -->
          <div
            v-if="isImageAttachment"
            class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-1.5 bg-slate-900/85 backdrop-blur-md px-3 py-1.5 rounded-2xl border border-white/10 shadow-2xl z-20"
          >
            <button
              @click="zoomOut"
              class="p-2 text-slate-300 hover:text-white hover:bg-white/15 rounded-xl transition-colors"
              title="Zoom Out"
            >
              <ZoomOut class="w-4 h-4" />
            </button>
            <span class="text-white text-xs font-black w-12 text-center select-none font-mono">
              {{ Math.round(zoomLevel * 100) }}%
            </span>
            <button
              @click="zoomIn"
              class="p-2 text-slate-300 hover:text-white hover:bg-white/15 rounded-xl transition-colors"
              title="Zoom In"
            >
              <ZoomIn class="w-4 h-4" />
            </button>

            <div class="w-px h-5 bg-white/20 mx-1"></div>

            <button
              @click="rotateCcw"
              class="p-2 text-slate-300 hover:text-white hover:bg-white/15 rounded-xl transition-colors"
              title="Putar Kiri 90°"
            >
              <RotateCcw class="w-4 h-4" />
            </button>
            <button
              @click="rotateCw"
              class="p-2 text-slate-300 hover:text-white hover:bg-white/15 rounded-xl transition-colors"
              title="Putar Kanan 90°"
            >
              <RotateCw class="w-4 h-4" />
            </button>

            <div class="w-px h-5 bg-white/20 mx-1"></div>

            <button
              @click="resetView"
              class="p-2 text-slate-300 hover:text-white hover:bg-white/15 rounded-xl transition-colors"
              title="Reset Tampilan (100% & 0°)"
            >
              <RefreshCw class="w-4 h-4" />
            </button>
          </div>
        </div>

        <DialogFooter class="pt-3 border-t border-slate-100 flex items-center justify-between sm:justify-between w-full">
          <p class="text-[11px] font-mono text-slate-400 truncate max-w-[65%]">
            {{ selectedAttachmentURL }}
          </p>
          <Button variant="outline" class="rounded-xl font-bold text-xs h-9 px-4" @click="attachmentModalOpen = false">
            Tutup
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
